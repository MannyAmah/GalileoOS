"""Temporal worker for the Onboarding Crew.

Subscribes to the ``galileo-onboarding-crew`` task queue (separate from
the Go agent-runner's ``galileo-agent-runner`` queue — Python and Go
workers do not share a queue). Registers both Stage 0 workflows and
both Stage 0 activities.

Run via the ``galileo-onboarding-worker`` console script
(``pyproject.toml [project.scripts]``):

    galileo-onboarding-worker

Environment variables:
- ``GALILEO_TEMPORAL_HOSTPORT`` — default ``localhost:7233``
- ``GALILEO_TEMPORAL_NAMESPACE`` — default ``default``
- ``GALILEO_ONBOARDING_DEV_KEY_PATH`` — default ``kernel/auth/dev-keys/private.pem``
- ``GALILEO_GATEWAY_DATABASE_URL`` — Postgres URL used by activities

The worker fails fast on missing required env vars rather than running
with silent defaults that would only fail at activity execution time.
"""

from __future__ import annotations

import asyncio
import logging
import os
import signal
import sys

from temporalio.client import Client
from temporalio.worker import Worker

from .connector import verify_source_auth
from .crawler import crawl_source
from .workflows import ConnectorWorkflow, CrawlerWorkflow

TASK_QUEUE = "galileo-onboarding-crew"

logger = logging.getLogger(__name__)


def _require_env(name: str) -> str:
    value = os.environ.get(name)
    if not value:
        sys.stderr.write(f"galileo-onboarding-worker: missing required env var {name}\n")
        sys.exit(2)
    return value


async def _run() -> None:
    logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(name)s: %(message)s")

    hostport = os.environ.get("GALILEO_TEMPORAL_HOSTPORT", "localhost:7233")
    namespace = os.environ.get("GALILEO_TEMPORAL_NAMESPACE", "default")
    _require_env("GALILEO_GATEWAY_DATABASE_URL")  # validate at startup, used inside activities

    client = await Client.connect(hostport, namespace=namespace)
    worker = Worker(
        client,
        task_queue=TASK_QUEUE,
        workflows=[ConnectorWorkflow, CrawlerWorkflow],
        activities=[verify_source_auth, crawl_source],
    )

    loop = asyncio.get_running_loop()
    stop = asyncio.Event()
    for sig in (signal.SIGINT, signal.SIGTERM):
        loop.add_signal_handler(sig, stop.set)

    logger.info("galileo-onboarding-worker ready hostport=%s namespace=%s", hostport, namespace)

    async with worker:
        await stop.wait()
    logger.info("galileo-onboarding-worker stopped")


def main() -> None:
    asyncio.run(_run())


if __name__ == "__main__":
    main()
