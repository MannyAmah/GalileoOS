"""Source-kind dataclasses for the Onboarding Crew.

The Connector and Crawler dispatch on ``SourceKind`` to one of three
integration paths (github via MCP subprocess; slack via slack_sdk;
gdrive via google-api-python-client). The credential payload shape
differs per kind; the dataclasses in this module make the shape
explicit at the type level so mypy --strict catches drift between the
Connector's persistence call and the Crawler's read.
"""

from __future__ import annotations

from dataclasses import dataclass
from enum import StrEnum


class SourceKind(StrEnum):
    """Stage 0 source-kinds. Add new kinds here AND in the dispatch
    tables in connector.py and crawler.py — a SourceKind without a
    matching dispatch entry trips a mypy exhaustiveness error."""

    GITHUB = "github"
    SLACK = "slack"
    GDRIVE = "gdrive"


@dataclass(frozen=True, slots=True)
class GitHubCredential:
    pat: str
    """Fine-grained PAT; scopes locked to contents:read, metadata:read."""


@dataclass(frozen=True, slots=True)
class SlackCredential:
    bot_token: str
    """xoxb-... token; scopes locked to channels:read, groups:read, users:read."""


@dataclass(frozen=True, slots=True)
class GoogleDriveCredential:
    service_account_json: str
    """Full service-account JSON document; scope locked to drive.readonly."""


@dataclass(frozen=True, slots=True)
class SourceConfig:
    """One entry from the operator's sources.yaml — what source-kind to
    connect and where to find its credentials. The credentials field
    holds the loaded credential value, not a path; the CLI is what
    reads paths into values before invoking the workflow."""

    kind: SourceKind
    credential: GitHubCredential | SlackCredential | GoogleDriveCredential
