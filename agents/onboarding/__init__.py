"""Galileo Onboarding Crew — six-agent Temporal workflow per plan §3.

Stage 0 scope (Week 4 / PR-D): Connector + Crawler agents wired against
the internal test workspace. Full crew (Ingestion, Org-Mapper,
Skill-Selector, QA) lands in Stage 1.

Per-source dispatch lives in ``connector`` and ``crawler``; see
``docs/decisions/0005-mcp-per-source-vs-mixed.md`` for the structural
reasoning behind MCP-for-github / direct-SDK-for-slack-and-gdrive.
"""

__all__ = [
    "connector",
    "crawler",
    "credentials",
    "sources",
    "workflows",
]
