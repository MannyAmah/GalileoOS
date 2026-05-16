"""Stage 0 credentials store. Lives in Python because the Onboarding Crew
(only PR-D consumer) is Python. If a Go service needs to read credentials,
migrate this to ``kernel/auth/credentials.go`` behind a gRPC contract rather
than calling Python from Go. See ADR-0005 reversal triggers.

Encryption: AES-256-GCM. Key derivation: HKDF-SHA256 over the Stage 0
Ed25519 dev private key bytes as IKM, fixed salt, info string
``galileo-stage0-credentials-encryption``, output 32 bytes. Standard
primitives from ``cryptography.hazmat.primitives`` — no novel
cryptography. The Ed25519 private key already exists for JWT signing
(``make stage0-jwt-setup``); reusing its bytes as IKM means the
credentials store has the same operational handle as the JWT-signing
key, and operators rotate one thing rather than two.
"""

from __future__ import annotations

import os
import secrets
from pathlib import Path
from typing import Final

from cryptography.hazmat.primitives import hashes, serialization
from cryptography.hazmat.primitives.ciphers.aead import AESGCM
from cryptography.hazmat.primitives.kdf.hkdf import HKDF

_HKDF_SALT: Final[bytes] = b"galileo-stage0-credentials-salt-v1"
_HKDF_INFO: Final[bytes] = b"galileo-stage0-credentials-encryption"
_KEY_BYTES: Final[int] = 32  # AES-256
_NONCE_BYTES: Final[int] = 12  # GCM standard
_DEFAULT_DEV_KEY_PATH: Final[str] = "kernel/auth/dev-keys/private.pem"


class CredentialStoreError(Exception):
    """Raised when the credentials store can't load its key material.
    Distinct from ``cryptography``'s ``InvalidTag`` (decrypt failures),
    which the caller catches directly."""


def _load_ikm(dev_key_path: str) -> bytes:
    """Load the Ed25519 private key bytes for use as HKDF input keying
    material. Fail loud if the key file is missing — Stage 0 expects
    ``make stage0-jwt-setup`` to have run."""

    path = Path(dev_key_path)
    if not path.is_file():
        raise CredentialStoreError(
            f"Stage 0 dev key not found at {dev_key_path}. "
            "Run `make stage0-jwt-setup` to generate the Ed25519 keypair."
        )
    pem = path.read_bytes()
    key = serialization.load_pem_private_key(pem, password=None)
    return key.private_bytes(
        encoding=serialization.Encoding.Raw,
        format=serialization.PrivateFormat.Raw,
        encryption_algorithm=serialization.NoEncryption(),
    )


def _derive_aes_key(ikm: bytes) -> bytes:
    hkdf = HKDF(algorithm=hashes.SHA256(), length=_KEY_BYTES, salt=_HKDF_SALT, info=_HKDF_INFO)
    return hkdf.derive(ikm)


class CredentialStore:
    """AES-256-GCM wrapper around an HKDF-derived key. Construct once
    per worker process; reuse across encrypt/decrypt calls.

    On-wire format for ``encrypt`` output: ``nonce (12 bytes) || ciphertext+tag``.
    The 12-byte nonce is prepended so ``decrypt`` is self-contained
    against the persisted bytes; no separate nonce column in Postgres.
    """

    def __init__(self, *, dev_key_path: str | None = None) -> None:
        path = dev_key_path or os.environ.get(
            "GALILEO_ONBOARDING_DEV_KEY_PATH", _DEFAULT_DEV_KEY_PATH
        )
        ikm = _load_ikm(path)
        self._aead = AESGCM(_derive_aes_key(ikm))

    def encrypt(self, plaintext: bytes, *, associated_data: bytes = b"") -> bytes:
        """Encrypt ``plaintext``. ``associated_data`` is bound into the
        GCM tag (e.g., ``f"{tenant_id}:{source_kind}".encode()`` so a
        ciphertext lifted to a different (tenant, source) row fails to
        decrypt)."""

        nonce = secrets.token_bytes(_NONCE_BYTES)
        ct = self._aead.encrypt(nonce, plaintext, associated_data)
        return nonce + ct

    def decrypt(self, ciphertext: bytes, *, associated_data: bytes = b"") -> bytes:
        if len(ciphertext) < _NONCE_BYTES:
            raise CredentialStoreError("ciphertext too short to contain a nonce")
        nonce, body = ciphertext[:_NONCE_BYTES], ciphertext[_NONCE_BYTES:]
        return self._aead.decrypt(nonce, body, associated_data)
