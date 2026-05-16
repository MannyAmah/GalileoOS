"""Connector activities — per-source authentication verification.

The Connector's job is narrow: confirm that the credentials persisted
by the CLI actually authenticate against each named source. It does
**no** ingestion; it does **no** enumeration; it does **no** write
operations against any source. Verification only.

Per-source dispatch (see ADR-0005):
- ``github`` → Docker subprocess of ``ghcr.io/github/github-mcp-server`` via
  the mcp Python SDK's ``stdio_client``. Tool call: ``list_tools`` returns
  the server's tool catalog if and only if the PAT authenticates.
- ``slack`` → ``slack_sdk.WebClient.auth_test()`` — Slack's documented
  zero-scope auth probe.
- ``gdrive`` → ``service.about().get(fields="user")`` — Google's documented
  one-RTT auth probe scoped to ``drive.readonly``.

Each verifier raises on failure; the workflow catches and surfaces the
failure into the crawl_status update. None of this code path opens a
write surface — the read-only scopes are enforced upstream at the OAuth
authorization step.
"""

from __future__ import annotations

import asyncio
import json
import logging
from typing import Final

import psycopg
from googleapiclient.discovery import build as google_build
from google.oauth2 import service_account
from mcp import ClientSession, StdioServerParameters
from mcp.client.stdio import stdio_client
from slack_sdk import WebClient
from slack_sdk.errors import SlackApiError
from temporalio import activity

from .credentials import CredentialStore
from .sources import SourceKind

logger = logging.getLogger(__name__)

_GITHUB_MCP_IMAGE: Final[str] = "ghcr.io/github/github-mcp-server:v1.0.4"
_GDRIVE_SCOPES: Final[list[str]] = ["https://www.googleapis.com/auth/drive.readonly"]


class AuthVerificationError(Exception):
    """Raised when a credential fails to authenticate against its source.
    Surfaced into ``tenant_manifests.crawl_status = 'failed'`` by the
    workflow so the operator-facing diagnostic is the source-named
    failure, not a Python traceback."""


def _aad(tenant_id: str, kind: SourceKind) -> bytes:
    """Associated-data binding for the GCM ciphertext. Prevents a row
    lifted to a different (tenant, source) from decrypting."""

    return f"{tenant_id}:{kind.value}".encode()


def _load_credential(
    conn: psycopg.Connection,
    store: CredentialStore,
    tenant_id: str,
    kind: SourceKind,
) -> bytes:
    row = conn.execute(
        "SELECT encrypted_payload FROM public.tenant_credentials WHERE tenant_id = %s AND source_kind = %s",
        (tenant_id, kind.value),
    ).fetchone()
    if row is None:
        raise AuthVerificationError(
            f"no credential persisted for tenant={tenant_id} source={kind.value}; "
            "run `galileo-onboarding --config sources.yaml --tenant <UUID>` first"
        )
    return store.decrypt(row[0], associated_data=_aad(tenant_id, kind))


async def _verify_github(pat: str) -> None:
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
    try:
        async with stdio_client(params) as (read, write):
            async with ClientSession(read, write) as session:
                await session.initialize()
                tools = await session.list_tools()
                if not tools.tools:
                    raise AuthVerificationError(
                        "github-mcp-server returned empty tool list — PAT scopes likely insufficient"
                    )
    except AuthVerificationError:
        raise
    except Exception as exc:  # mcp subprocess errors land here
        raise AuthVerificationError(f"github MCP subprocess failed: {exc}") from exc


def _verify_slack(bot_token: str) -> None:
    client = WebClient(token=bot_token)
    try:
        response = client.auth_test()
    except SlackApiError as exc:
        raise AuthVerificationError(
            f"slack auth.test failed: {exc.response.get('error')}"
        ) from exc
    if not response.get("ok"):
        raise AuthVerificationError(
            f"slack auth.test returned not-ok: {response.data!r}"
        )


def _verify_gdrive(service_account_json: str) -> None:
    try:
        info = json.loads(service_account_json)
        # google.oauth2 declares the type but the function's return is Any;
        # ``no-untyped-call`` fires at the call site in our strictly-typed
        # code rather than inside google.oauth2. Scope the silence to the
        # exact line — the override in pyproject.toml relaxes the third-
        # party boundary, but mypy's strict caller-side check still fires.
        creds = service_account.Credentials.from_service_account_info(  # type: ignore[no-untyped-call]
            info, scopes=_GDRIVE_SCOPES
        )
        service = google_build("drive", "v3", credentials=creds, cache_discovery=False)
        service.about().get(fields="user").execute()
    except Exception as exc:
        raise AuthVerificationError(f"gdrive auth probe failed: {exc}") from exc


@activity.defn
async def verify_source_auth(
    tenant_id: str, source_kind: str, database_url: str
) -> None:
    """Activity: load and verify a single source's credentials.

    Reads the encrypted credential from Postgres, decrypts it locally
    (the AES key lives in this process, derived from the local Ed25519
    dev key), and runs the per-source auth probe. Raises
    ``AuthVerificationError`` on any failure — the workflow surfaces it
    into the manifest's ``crawl_status`` so operators see "github auth
    failed" instead of "Python exception."
    """

    kind = SourceKind(source_kind)
    store = CredentialStore()
    with psycopg.connect(database_url) as conn:
        plaintext = _load_credential(conn, store, tenant_id, kind)

    activity.logger.info(
        "verify_source_auth start tenant=%s kind=%s", tenant_id, kind.value
    )
    if kind is SourceKind.GITHUB:
        await _verify_github(plaintext.decode())
    elif kind is SourceKind.SLACK:
        await asyncio.to_thread(_verify_slack, plaintext.decode())
    elif kind is SourceKind.GDRIVE:
        await asyncio.to_thread(_verify_gdrive, plaintext.decode())
    else:
        # mypy exhaustiveness — adding a new SourceKind without dispatch is a build error.
        from typing import assert_never

        assert_never(kind)
    activity.logger.info(
        "verify_source_auth ok tenant=%s kind=%s", tenant_id, kind.value
    )
