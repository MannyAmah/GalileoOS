"""CLI entry for the Onboarding Crew.

Reads an operator-supplied ``sources.yaml``, encrypts each credential
under the per-tenant + per-source associated-data binding, persists to
``tenant_credentials``, then triggers ``ConnectorWorkflow`` followed
by ``CrawlerWorkflow`` on the ``galileo-onboarding-crew`` task queue.

The CLI is the **only** invocation surface in Stage 0. Operator-review
UI is deferred to Stage 1 per spec §6.

Example ``sources.yaml``:

    tenant_id: "01HNG7A0K0V8K8K0V8K8K0V8K8"  # UUIDv7
    sources:
      - kind: github
        credential_path: ~/.galileo/github.pat
      - kind: slack
        credential_path: ~/.galileo/slack.bot.token
      - kind: gdrive
        credential_path: ~/.galileo/gdrive-service-account.json

Run via:

    galileo-onboarding --config sources.yaml
"""

from __future__ import annotations

import argparse
import asyncio
import logging
import os
import sys
import uuid
from pathlib import Path
from typing import Any

import psycopg
import yaml
from temporalio.client import Client

from .credentials import CredentialStore
from .sources import SourceKind
from .workflows import ConnectorWorkflow, CrawlerWorkflow
from .worker import TASK_QUEUE

logger = logging.getLogger(__name__)


def _load_config(path: Path) -> tuple[str, list[tuple[SourceKind, Path]]]:
    raw = yaml.safe_load(path.read_text())
    if not isinstance(raw, dict):
        raise ValueError(f"{path}: expected a mapping at top level")
    tenant_id = str(raw["tenant_id"])
    uuid.UUID(tenant_id)  # validate UUID shape; raises ValueError on bad input
    entries: list[tuple[SourceKind, Path]] = []
    for entry in raw.get("sources", []):
        kind = SourceKind(entry["kind"])
        cred_path = Path(entry["credential_path"]).expanduser()
        if not cred_path.is_file():
            raise FileNotFoundError(
                f"{path}: credential_path {cred_path} does not exist"
            )
        entries.append((kind, cred_path))
    if not entries:
        raise ValueError(f"{path}: sources list is empty")
    return tenant_id, entries


def _persist_credentials(
    database_url: str,
    tenant_id: str,
    entries: list[tuple[SourceKind, Path]],
) -> list[str]:
    """Encrypt and upsert each credential. Returns the list of
    source-kinds that were persisted (in input order) so the caller
    can pass them to the workflow."""

    from .connector import _aad  # private helper; CLI is the only other caller

    store = CredentialStore()
    persisted: list[str] = []
    with psycopg.connect(database_url) as conn:
        for kind, path in entries:
            plaintext = path.read_bytes()
            ciphertext = store.encrypt(plaintext, associated_data=_aad(tenant_id, kind))
            conn.execute(
                """
                INSERT INTO tenant_credentials (tenant_id, source_kind, encrypted_payload)
                VALUES (%s, %s, %s)
                ON CONFLICT (tenant_id, source_kind) DO UPDATE
                SET encrypted_payload = EXCLUDED.encrypted_payload, updated_at = NOW()
                """,
                (tenant_id, kind.value, ciphertext),
            )
            persisted.append(kind.value)
        conn.commit()
    return persisted


async def _trigger_workflows(
    tenant_id: str, source_kinds: list[str], database_url: str
) -> None:
    hostport = os.environ.get("GALILEO_TEMPORAL_HOSTPORT", "localhost:7233")
    namespace = os.environ.get("GALILEO_TEMPORAL_NAMESPACE", "default")
    client = await Client.connect(hostport, namespace=namespace)

    connector_id = f"onboarding-connector-{tenant_id}"
    crawler_id = f"onboarding-crawler-{tenant_id}"

    logger.info("starting ConnectorWorkflow id=%s", connector_id)
    await client.execute_workflow(
        ConnectorWorkflow.run,
        args=(tenant_id, source_kinds, database_url),
        id=connector_id,
        task_queue=TASK_QUEUE,
    )

    logger.info("starting CrawlerWorkflow id=%s", crawler_id)
    await client.execute_workflow(
        CrawlerWorkflow.run,
        args=(tenant_id, source_kinds, database_url),
        id=crawler_id,
        task_queue=TASK_QUEUE,
    )


def _parse_args(argv: list[str]) -> Any:
    parser = argparse.ArgumentParser(prog="galileo-onboarding")
    parser.add_argument("--config", required=True, type=Path, help="sources.yaml path")
    return parser.parse_args(argv)


def main(argv: list[str] | None = None) -> None:
    logging.basicConfig(
        level=logging.INFO, format="%(asctime)s %(levelname)s %(name)s: %(message)s"
    )
    args = _parse_args(argv if argv is not None else sys.argv[1:])

    database_url = os.environ.get("GALILEO_GATEWAY_DATABASE_URL")
    if not database_url:
        sys.stderr.write(
            "galileo-onboarding: GALILEO_GATEWAY_DATABASE_URL is required\n"
        )
        sys.exit(2)

    tenant_id, entries = _load_config(args.config)
    logger.info("persisting %d credentials for tenant=%s", len(entries), tenant_id)
    source_kinds = _persist_credentials(database_url, tenant_id, entries)

    asyncio.run(_trigger_workflows(tenant_id, source_kinds, database_url))
    logger.info("onboarding workflows complete for tenant=%s", tenant_id)


if __name__ == "__main__":
    main()
