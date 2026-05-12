"""Connector Agent stub.

Stage 0 skeleton. Week 4 wires real OAuth to GitHub, Slack, and Google
Drive per docs/plans/STAGE_0_PLAN.md §Week 4. Mirage-mounted vs.
discrete-MCP behavior depends on Week 2 probe outcome.
"""

from __future__ import annotations


def name() -> str:
    """Return the agent identifier used by the runner."""
    return "galileo-onboarding-connector"
