"""Crawler activities — per-source enumeration and manifest emission.

The Crawler walks each authenticated source through the same client
shape the Connector verified and emits one manifest row per source
into ``tenant_manifests``. Stage 0 caps (per plan §3.3 step 5): 50K
documents per source, 6 hours of wall-clock per workflow, $50 of LLM
spend per workflow. The LLM cap is enforced by the gateway's budget
middleware against ``cost_events`` — the crawler emits no LLM calls
itself, so the cap is reached only if downstream agents (Stage 1) call
LLMs during ingestion.

Manifest schema (rows are JSON arrays inside ``manifest_json``):
- ``path`` — source-relative path (repo path / channel ID / drive file ID)
- ``size_bytes`` — int
- ``content_type`` — MIME or source-specific (``application/vnd.github.repo`` etc.)
- ``last_modified_unix`` — int, source-reported

The crawler is read-only by construction. Every per-source dispatch
uses the read-only OAuth scope set named in ADR-0005; no write API on
any client is invoked.
"""

from __future__ import annotations

import asyncio
import json
import logging
import time
from typing import Any, Final

import psycopg
from googleapiclient.discovery import build as google_build
from google.oauth2 import service_account
from mcp import ClientSession, StdioServerParameters
from mcp.client.stdio import stdio_client
from slack_sdk import WebClient
from temporalio import activity

from .connector import _load_credential
from .credentials import CredentialStore
from .sources import SourceKind

logger = logging.getLogger(__name__)

_GITHUB_MCP_IMAGE: Final[str] = "ghcr.io/github/github-mcp-server:v1.0.4"
_GDRIVE_SCOPES: Final[list[str]] = ["https://www.googleapis.com/auth/drive.readonly"]
_MAX_DOCS_PER_SOURCE: Final[int] = 50_000
_MAX_WALL_CLOCK_SECONDS: Final[int] = 6 * 60 * 60


async def _crawl_github(pat: str) -> list[dict[str, Any]]:
    """Enumerate repositories the PAT can read. The github MCP server
    exposes a ``list_repositories`` tool against the authenticated user;
    we call it once and convert each repo into a manifest row.
    """

    params = StdioServerParameters(
        command="docker",
        args=[
            "run",
            "-i",
            "--rm",
            "--init",
            "-e",
            "GITHUB_PERSONAL_ACCESS_TOKEN",
            _GITHUB_MCP_IMAGE,
        ],
        env={"GITHUB_PERSONAL_ACCESS_TOKEN": pat},
    )
    rows: list[dict[str, Any]] = []
    async with stdio_client(params) as (read, write):
        async with ClientSession(read, write) as session:
            await session.initialize()
            tools = await session.list_tools()
            tool_names = {t.name for t in tools.tools}
            # github-mcp-server names this ``search_repositories`` in current
            # releases; ``list_repositories`` is the older alias. Try the
            # current name first; fall back if the server is older.
            target = None
            for candidate in ("search_repositories", "list_repositories"):
                if candidate in tool_names:
                    target = candidate
                    break
            if target is None:
                raise RuntimeError(
                    f"github-mcp-server has neither search_repositories nor list_repositories; "
                    f"available tools: {sorted(tool_names)}"
                )
            result = await session.call_tool(target, arguments={"query": "user:@me"})
            for block in result.content:
                # Narrow the union by attribute existence; mcp's content
                # blocks are TextContent | ImageContent | ... and only
                # TextContent carries a .text field.
                text = getattr(block, "text", None)
                if not isinstance(text, str):
                    continue
                try:
                    payload = json.loads(text)
                except json.JSONDecodeError:
                    continue
                for repo in payload.get(
                    "items", payload if isinstance(payload, list) else []
                ):
                    rows.append(
                        {
                            "path": repo.get("full_name", ""),
                            "size_bytes": int(repo.get("size", 0)) * 1024,
                            "content_type": "application/vnd.github.repo",
                            "last_modified_unix": _parse_iso_unix(
                                repo.get("updated_at")
                            ),
                        }
                    )
                    if len(rows) >= _MAX_DOCS_PER_SOURCE:
                        return rows
    return rows


def _crawl_slack(bot_token: str) -> list[dict[str, Any]]:
    client = WebClient(token=bot_token)
    rows: list[dict[str, Any]] = []
    cursor: str | None = None
    while True:
        response = client.conversations_list(
            cursor=cursor, limit=200, types="public_channel,private_channel"
        )
        channels: list[dict[str, Any]] = response.get("channels", []) or []
        for channel in channels:
            rows.append(
                {
                    "path": channel.get("id", ""),
                    "size_bytes": 0,  # Slack doesn't report channel size; ingestion will compute
                    "content_type": "application/vnd.slack.channel",
                    "last_modified_unix": int(channel.get("updated", 0)) // 1000,
                }
            )
            if len(rows) >= _MAX_DOCS_PER_SOURCE:
                return rows
        metadata: dict[str, Any] = response.get("response_metadata") or {}
        cursor = metadata.get("next_cursor") or None
        if not cursor:
            break
    return rows


def _crawl_gdrive(service_account_json: str) -> list[dict[str, Any]]:
    info = json.loads(service_account_json)
    # See connector.py:_verify_gdrive for the rationale on this silence.
    creds = service_account.Credentials.from_service_account_info(  # type: ignore[no-untyped-call]
        info, scopes=_GDRIVE_SCOPES
    )
    service = google_build("drive", "v3", credentials=creds, cache_discovery=False)
    rows: list[dict[str, Any]] = []
    page_token: str | None = None
    while True:
        response = (
            service.files()
            .list(
                pageSize=200,
                pageToken=page_token,
                fields="nextPageToken, files(id, name, size, mimeType, modifiedTime)",
                q="trashed = false",
            )
            .execute()
        )
        for f in response.get("files", []):
            rows.append(
                {
                    "path": f["id"],
                    "size_bytes": int(f.get("size", 0)),
                    "content_type": f.get("mimeType", "application/octet-stream"),
                    "last_modified_unix": _parse_iso_unix(f.get("modifiedTime")),
                }
            )
            if len(rows) >= _MAX_DOCS_PER_SOURCE:
                return rows
        page_token = response.get("nextPageToken")
        if not page_token:
            break
    return rows


def _parse_iso_unix(iso: str | None) -> int:
    if not iso:
        return 0
    # Both GitHub and Google return RFC 3339 with a trailing Z. fromisoformat
    # in 3.12+ handles the Z suffix; earlier versions need replace.
    from datetime import datetime

    return int(datetime.fromisoformat(iso.replace("Z", "+00:00")).timestamp())


@activity.defn
async def crawl_source(tenant_id: str, source_kind: str, database_url: str) -> None:
    """Activity: enumerate one source and emit its manifest row.

    Updates ``tenant_manifests.crawl_status`` to ``in_progress`` at
    start, ``completed`` on success, ``failed`` on exception. Enforces
    the 50K-docs cap and the 6-hour wall-clock cap; the LLM cap is
    enforced upstream by the gateway against ``cost_events``.
    """

    kind = SourceKind(source_kind)
    store = CredentialStore()
    started = time.monotonic()

    with psycopg.connect(database_url) as conn:
        plaintext = _load_credential(conn, store, tenant_id, kind)
        _upsert_manifest_status(
            conn, tenant_id, kind, "in_progress", manifest_json=None
        )
        conn.commit()

    activity.logger.info("crawl_source start tenant=%s kind=%s", tenant_id, kind.value)

    try:
        if kind is SourceKind.GITHUB:
            rows = await _crawl_github(plaintext.decode())
        elif kind is SourceKind.SLACK:
            rows = await asyncio.to_thread(_crawl_slack, plaintext.decode())
        elif kind is SourceKind.GDRIVE:
            rows = await asyncio.to_thread(_crawl_gdrive, plaintext.decode())
        else:
            from typing import assert_never

            assert_never(kind)
        if time.monotonic() - started > _MAX_WALL_CLOCK_SECONDS:
            raise RuntimeError(
                f"crawl exceeded wall-clock cap ({_MAX_WALL_CLOCK_SECONDS}s)"
            )
        with psycopg.connect(database_url) as conn:
            _upsert_manifest_status(
                conn, tenant_id, kind, "completed", manifest_json=rows
            )
            conn.commit()
        activity.logger.info(
            "crawl_source ok tenant=%s kind=%s docs=%d",
            tenant_id,
            kind.value,
            len(rows),
        )
    except Exception:
        with psycopg.connect(database_url) as conn:
            _upsert_manifest_status(conn, tenant_id, kind, "failed", manifest_json=None)
            conn.commit()
        raise


def _upsert_manifest_status(
    conn: psycopg.Connection,
    tenant_id: str,
    kind: SourceKind,
    status: str,
    *,
    manifest_json: list[dict[str, Any]] | None,
) -> None:
    if manifest_json is None:
        conn.execute(
            """
            INSERT INTO public.tenant_manifests (tenant_id, source_kind, crawl_status, crawl_started_at)
            VALUES (%s, %s, %s, NOW())
            ON CONFLICT (tenant_id, source_kind) DO UPDATE
            SET crawl_status = EXCLUDED.crawl_status,
                crawl_started_at = COALESCE(public.tenant_manifests.crawl_started_at, NOW())
            """,
            (tenant_id, kind.value, status),
        )
    else:
        conn.execute(
            """
            INSERT INTO public.tenant_manifests (
                tenant_id, source_kind, crawl_status,
                crawl_completed_at, document_count, total_bytes, manifest_json
            ) VALUES (%s, %s, %s, NOW(), %s, %s, %s::jsonb)
            ON CONFLICT (tenant_id, source_kind) DO UPDATE
            SET crawl_status = EXCLUDED.crawl_status,
                crawl_completed_at = EXCLUDED.crawl_completed_at,
                document_count = EXCLUDED.document_count,
                total_bytes = EXCLUDED.total_bytes,
                manifest_json = EXCLUDED.manifest_json
            """,
            (
                tenant_id,
                kind.value,
                status,
                len(manifest_json),
                sum(int(r.get("size_bytes", 0)) for r in manifest_json),
                json.dumps(manifest_json),
            ),
        )
