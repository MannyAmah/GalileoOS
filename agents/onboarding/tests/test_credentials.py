"""Smoke tests for the credentials store. Deterministic and offline —
generates an ephemeral Ed25519 keypair, exercises the HKDF + AES-GCM
roundtrip, and verifies the associated-data binding prevents cross-row
decrypt."""

from __future__ import annotations

import tempfile
from pathlib import Path

import pytest
from cryptography.hazmat.primitives import serialization
from cryptography.hazmat.primitives.asymmetric.ed25519 import Ed25519PrivateKey

from onboarding.credentials import CredentialStore, CredentialStoreError


def _write_ephemeral_keypair(dirpath: Path) -> Path:
    key = Ed25519PrivateKey.generate()
    pem = key.private_bytes(
        encoding=serialization.Encoding.PEM,
        format=serialization.PrivateFormat.PKCS8,
        encryption_algorithm=serialization.NoEncryption(),
    )
    path = dirpath / "private.pem"
    path.write_bytes(pem)
    return path


def test_roundtrip_basic() -> None:
    with tempfile.TemporaryDirectory() as td:
        key_path = _write_ephemeral_keypair(Path(td))
        store = CredentialStore(dev_key_path=str(key_path))
        plaintext = b"ghp_test_token_123"
        ct = store.encrypt(plaintext, associated_data=b"tenant-a:github")
        assert store.decrypt(ct, associated_data=b"tenant-a:github") == plaintext


def test_associated_data_binding_prevents_cross_row_decrypt() -> None:
    with tempfile.TemporaryDirectory() as td:
        key_path = _write_ephemeral_keypair(Path(td))
        store = CredentialStore(dev_key_path=str(key_path))
        ct = store.encrypt(b"secret", associated_data=b"tenant-a:github")
        with pytest.raises(Exception):  # cryptography raises InvalidTag
            store.decrypt(ct, associated_data=b"tenant-b:github")


def test_missing_dev_key_raises() -> None:
    with pytest.raises(CredentialStoreError, match="not found"):
        CredentialStore(dev_key_path="/nonexistent/path/private.pem")


def test_nonces_differ_across_encrypts() -> None:
    """GCM nonce reuse with the same key is catastrophic; the
    encrypt helper must use a fresh random nonce every call. Two
    encrypts of the same plaintext should produce distinct ciphertexts
    even before the tag (the first 12 bytes are the prepended nonce)."""

    with tempfile.TemporaryDirectory() as td:
        key_path = _write_ephemeral_keypair(Path(td))
        store = CredentialStore(dev_key_path=str(key_path))
        a = store.encrypt(b"same", associated_data=b"x")
        b = store.encrypt(b"same", associated_data=b"x")
        assert a[:12] != b[:12], "nonce reuse detected"
