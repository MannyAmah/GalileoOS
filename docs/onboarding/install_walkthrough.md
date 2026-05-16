# Install walkthrough — Stage 0 Onboarding Crew (30 minutes)

This walkthrough takes a fresh Ubuntu 24.04 VM (or equivalent macOS/Linux dev machine) from `git clone` to a passing `manifest-check`. The senior-engineer install-walkthrough session of the Stage 0 gate follows this doc step-by-step; if you can't complete it in 30 minutes, that's a gate finding worth reporting in `docs/closeouts/STAGE_0_GATE.md`.

> **Video reference:** `[video link pending recording — to be added as a post-merge follow-up to PR-D after the doc and code stabilize]`

## Prerequisites

Three runtime prerequisites. All three are likely already installed on a dev machine; the walkthrough lists installation hints for the bare-VM case.

1. **Docker** ≥24 with the `docker` CLI on PATH. Stage 0's `make up` brings up Postgres + Temporal + LiteLLM via `docker compose`; the Onboarding Crew's github connector invokes `docker run` for the MCP server.
   - Ubuntu: `curl -fsSL https://get.docker.com | sh && sudo usermod -aG docker $USER` (then log out and back in).
   - macOS: install Docker Desktop from docker.com.
2. **Python ≥3.12.** `pyproject.toml` floors at 3.12.
   - Ubuntu 24.04 ships 3.12 as `python3`. Verify with `python3 --version`.
   - macOS: `brew install python@3.12`.
3. **Go ≥1.26** for the `manifest-check` binary. Not required at runtime for the worker or CLI — only when you run the gate check.
   - `https://go.dev/dl/` for the binary tarball; matches CI's Go pin.

No Node.js prerequisite. The original plan implied npm-distributed MCP servers; PR-D's plan-deviation per ADR-0005 substituted Docker subprocess for github (and direct SDKs for slack/gdrive), so Node.js is no longer needed.

## Step 1 — Clone and build (5 min)

```bash
git clone https://github.com/MannyAmah/GalileoOS.git
cd GalileoOS
make stage0-jwt-setup    # writes kernel/auth/dev-keys/{private.pem,public.pem}; gitignored
make up                  # docker compose up -d postgres temporal litellm jaeger otel-collector
docker compose -f deploy/compose/docker-compose.yml ps   # confirm 5/5 healthy
```

If any service is not healthy, re-run `docker compose ps` after 30 seconds — Temporal's start-period is 60 seconds.

## Step 2 — Install the Python worker (5 min)

```bash
cd agents
pip install -e ".[dev]"   # installs the 10 runtime deps + pytest + types-PyYAML
```

The `[dev]` extras pull `pytest` and `types-PyYAML` for the test pass. Skip `[dev]` if you only need to run the worker, not the tests.

Run the unit tests to confirm the install:

```bash
pytest -v onboarding/tests
```

Six tests should pass: four for the credentials store roundtrip + AAD binding + nonce uniqueness, two for `SourceKind` dispatch + dataclass frozenness.

## Step 3 — Prepare credentials (5 min)

The Onboarding Crew connects to three sources. Each credential goes into its own file in `~/.galileo/` (the path is referenced from `sources.yaml` in the next step). Files in `~/.galileo/` should be `chmod 600`.

### GitHub fine-grained PAT

Create a [fine-grained PAT](https://github.com/settings/personal-access-tokens/new) with **only** these scopes:

- `Repository access` → All repositories (or a specific repo for the walkthrough)
- `Repository permissions` → **Contents: Read**, **Metadata: Read** (read-only)
- No organization permissions, no other repository permissions

Save the token:

```bash
mkdir -p ~/.galileo && chmod 700 ~/.galileo
echo -n "github_pat_xxx..." > ~/.galileo/github.pat
chmod 600 ~/.galileo/github.pat
```

### Slack bot token

Create a [Slack app](https://api.slack.com/apps) in your test workspace. Under **OAuth & Permissions** → **Bot Token Scopes**, add exactly these three scopes:

- `channels:read`
- `groups:read`
- `users:read`

Do **not** add `im:read`, `mpim:read`, or any write scope. Install the app to your workspace; copy the **Bot User OAuth Token** (begins with `xoxb-`).

```bash
echo -n "xoxb-xxx..." > ~/.galileo/slack.bot.token
chmod 600 ~/.galileo/slack.bot.token
```

### Google Drive service account

Create a [GCP service account](https://console.cloud.google.com/iam-admin/serviceaccounts) and download its JSON key. Grant the service account `Viewer` access to the test Drive folder.

```bash
mv ~/Downloads/your-project-xxx.json ~/.galileo/gdrive-service-account.json
chmod 600 ~/.galileo/gdrive-service-account.json
```

## Step 4 — Write the sources.yaml (2 min)

```bash
cat > ~/.galileo/sources.yaml <<'EOF'
tenant_id: "01JCDA8GCRT9DW8M2EJBHBWNC2"   # any UUIDv7; the gateway accepts it as-is in Stage 0
sources:
  - kind: github
    credential_path: ~/.galileo/github.pat
  - kind: slack
    credential_path: ~/.galileo/slack.bot.token
  - kind: gdrive
    credential_path: ~/.galileo/gdrive-service-account.json
EOF
```

## Step 5 — Run the worker and the CLI (10 min)

Open two terminals. The worker subscribes to the `galileo-onboarding-crew` Temporal task queue; the CLI submits the workflow.

**Terminal A — worker:**

```bash
cd agents
export GALILEO_GATEWAY_DATABASE_URL='postgres://galileo:galileo@localhost:5432/galileo?sslmode=disable'
galileo-onboarding-worker
```

**Terminal B — CLI:**

```bash
cd agents
export GALILEO_GATEWAY_DATABASE_URL='postgres://galileo:galileo@localhost:5432/galileo?sslmode=disable'
galileo-onboarding --config ~/.galileo/sources.yaml
```

The CLI encrypts each credential under HKDF-derived AES-256-GCM, persists to `tenant_credentials`, then triggers `ConnectorWorkflow` followed by `CrawlerWorkflow`. Expected runtime against the test workspace: ~2-5 minutes for github (first Docker pull may take longer), ~10s each for slack and gdrive.

If any source fails authentication, the workflow surfaces a `crawl_status = 'failed'` row in `tenant_manifests`. Re-running the CLI with a corrected credential file re-encrypts and re-runs.

## Step 6 — Validate the manifest (3 min)

```bash
cd kernel
go run ./cmd/manifest-check -tenant 01JCDA8GCRT9DW8M2EJBHBWNC2
```

Expected output (exact tenant UUID will differ):

```
[gate] wall-clock: OK (all sources within 6h0m0s cap)
[gate] LLM cost: OK spent_cents=0 cap_cents=5000
[gate] org-snapshot: OK completed=3/3 (100%)
[gate] Skill recommendation precision: N/A (Skill-Selector Agent deferred to Stage 1 per ADR-0003)
[gate] destructive actions: OK (0 — Stage 0 read-only by construction)
[gate] manifest-check: all dimensions passed
```

Exit code 0 means the Stage 0 gate passes for this run.

## Troubleshooting

- **`docker run` fails with "permission denied" on the Docker socket** — log out and back in after `usermod -aG docker $USER`, or prefix with `sudo` for the walkthrough.
- **`stage0 dev key not found` from the CLI** — re-run `make stage0-jwt-setup` from the repo root.
- **`github MCP subprocess failed: ... auth`** — the PAT is wrong, expired, or missing one of the two scopes. Re-create with `contents:read` and `metadata:read`.
- **`slack auth.test failed: invalid_auth`** — the bot token is wrong. Copy the **Bot User OAuth Token** (starts with `xoxb-`), not the User OAuth Token.
- **`gdrive auth probe failed: ... 403`** — the service account hasn't been granted access to the test folder, or the Drive API isn't enabled on the GCP project.
- **`org-snapshot: FAIL completed=2/3 (66%)`** — one source failed. Check `tenant_manifests.crawl_status` for which source-kind needs attention; the worker logs name the failure reason.
