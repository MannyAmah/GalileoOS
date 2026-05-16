"""Temporal workflow definitions for the Onboarding Crew.

Stage 0 ships two workflows:

- ``ConnectorWorkflow`` — verify auth for each source-kind. No
  ingestion; failure surfaces as ``crawl_status = 'failed'``.
- ``CrawlerWorkflow`` — enumerate each authenticated source and emit a
  manifest row.

Both workflows are deterministic — they call activities for every I/O
operation and never import psycopg, slack_sdk, or other clients
directly. Activity-side code (``connector.py``, ``crawler.py``) is what
opens connections and decrypts credentials.

The workflows take primitive types (str tenant_id, list[str]
source_kinds) rather than dataclasses so Temporal's default data
converter handles them without custom payload codecs.
"""

from __future__ import annotations

from datetime import timedelta

from temporalio import workflow

with workflow.unsafe.imports_passed_through():
    from .connector import verify_source_auth
    from .crawler import crawl_source


_ACTIVITY_TIMEOUT = timedelta(minutes=10)
_HEARTBEAT_TIMEOUT = timedelta(seconds=60)


@workflow.defn
class ConnectorWorkflow:
    """Verify credentials for each named source-kind. Fails the
    workflow if any verification fails — Stage 0 treats partial-auth
    as a failure mode the operator should fix before crawling, not as
    a soft warning.
    """

    @workflow.run
    async def run(self, tenant_id: str, source_kinds: list[str], database_url: str) -> None:
        for kind in source_kinds:
            await workflow.execute_activity(
                verify_source_auth,
                args=(tenant_id, kind, database_url),
                start_to_close_timeout=_ACTIVITY_TIMEOUT,
                heartbeat_timeout=_HEARTBEAT_TIMEOUT,
            )


@workflow.defn
class CrawlerWorkflow:
    """Enumerate each source-kind sequentially. Sequential rather than
    parallel for Stage 0 — three sources at modest size fit easily in
    one wall-clock pass; parallelism is a Stage 1 question once the
    Ingestion Agent adds real LLM cost.
    """

    @workflow.run
    async def run(self, tenant_id: str, source_kinds: list[str], database_url: str) -> None:
        for kind in source_kinds:
            await workflow.execute_activity(
                crawl_source,
                args=(tenant_id, kind, database_url),
                start_to_close_timeout=_ACTIVITY_TIMEOUT,
                heartbeat_timeout=_HEARTBEAT_TIMEOUT,
            )
