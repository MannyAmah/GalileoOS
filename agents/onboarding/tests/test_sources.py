"""Smoke tests for the SourceKind enum and dispatch dataclasses."""

from __future__ import annotations

import pytest

from onboarding.sources import (
    GitHubCredential,
    GoogleDriveCredential,
    SlackCredential,
    SourceConfig,
    SourceKind,
)


def test_source_kind_str_round_trip() -> None:
    assert SourceKind("github") is SourceKind.GITHUB
    assert SourceKind("slack") is SourceKind.SLACK
    assert SourceKind("gdrive") is SourceKind.GDRIVE
    assert SourceKind.GITHUB.value == "github"


def test_unknown_source_kind_rejected() -> None:
    with pytest.raises(ValueError):
        SourceKind("notion")


def test_source_config_dataclass_is_frozen() -> None:
    cfg = SourceConfig(
        kind=SourceKind.GITHUB, credential=GitHubCredential(pat="ghp_xxx")
    )
    with pytest.raises(Exception):
        cfg.kind = SourceKind.SLACK  # type: ignore[misc]


def test_each_credential_carries_expected_field() -> None:
    g = GitHubCredential(pat="ghp_xxx")
    s = SlackCredential(bot_token="xoxb-xxx")
    d = GoogleDriveCredential(service_account_json='{"type":"service_account"}')
    assert g.pat.startswith("ghp_")
    assert s.bot_token.startswith("xoxb-")
    assert d.service_account_json.startswith("{")
