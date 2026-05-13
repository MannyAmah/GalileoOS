# CLOSEOUT: Layer 3 Mirage placement reconsidered

| Field | Value |
| --- | --- |
| Decision date | 2026-05-13 |
| Decider | Emmanuel |
| Maker | Emmanuel + Claude Opus 4.7 (architects of the original plan that placed Mirage at Layer 3 with a Stage 0 probe gate) |
| Checker | Doc-reading round during PR #13 inline planning (2026-05-13), which surfaced Mirage's language surface and embedding-only deployment model |
| Supersedes | `docs/galileo_os_infrastructure_plan.md` Layer 3 substrate language ("Mirage VFS pending probe"); STAGE_0_PLAN.md §Week 2 probe (live measurement no longer the right experiment) |
| Status | Accepted |
| Type | Plan-deviation per [`CLAUDE.md` §Escalation](../../CLAUDE.md). **First plan-deviation in the project's history.** The shape used here — closeout + plan edits + ADR + reversal triggers — is the template for future deviations. |

## Structural finding

Mirage ships **Python and TypeScript SDKs only**. From Mirage's own documentation: *"Python and TypeScript SDKs give your AI agents a virtual filesystem directly inside FastAPI, Express, browser apps, or any async runtime, no separate process required."* No Go SDK exists. Mirage is designed for **in-process embedding in agent code**, not for use as a standalone substrate that a different-language host calls into.

Galileo's kernel is Go (locked in `docs/galileo_os_infrastructure_plan.md` §5 — "one service = one language"). Placing Mirage at Layer 3 (Brain substrate) as the original plan did would require either:

- A permanent **Python sidecar process** that runs `mirage-ai`, exposes its Workspace API over HTTP/gRPC, and that the Go kernel talks to as a client; or
- Reimplementing Mirage's filesystem semantics in Go from scratch, which defeats the adoption argument.

Neither was named in the original plan. The sidecar cost — every Galileo deployment running a Python service for the lifetime of the product to bridge a v0.0.x dependency — was not visible at adoption-time.

This is not a Mirage failure. It is a **placement failure**. Mirage fits cleanly where it was designed to fit: inside Python or TypeScript agent code. Galileo has Python agents (`agents/` per repo structure). The right placement is Layer 5 (Integrations), as a per-agent library choice — not Layer 3 (Brain substrate), as a kernel-side substrate.

## Three readings considered

**Reading 1 — Layer 3 with Python sidecar.** Keep Mirage at Layer 3; introduce a Python sidecar in `mcp-servers/mirage-sidecar/` that exposes Mirage over HTTP; Go kernel talks to the sidecar. **Rejected.** A permanent sidecar is a heavy architectural commitment for a v0.0.1 dependency. It also forfeits the language-purity discipline ("one service = one language") that exists specifically to make Galileo's operational surface small. The sidecar adds a process, a Dockerfile, a Python pin, an HTTP API contract to maintain against an upstream v0.0.x — for every Galileo tenant, forever, until Mirage publishes Go bindings or a native server (which it may never do).

**Reading 2 — Layer 5 as agent-side library (chosen).** Mirage stays available to Galileo, but at the integrations layer, not the substrate layer. Python agents that want Mirage's unified-filesystem abstraction import `mirage-ai` directly and use it in-process. Go kernel knows nothing about Mirage. Per-agent choice between Mirage's bash vocabulary and direct connectors; no kernel-level dependency. This fits Mirage's own deployment model exactly and preserves Galileo's kernel-language purity.

**Reading 3 — Drop Mirage entirely.** No Layer 5 Mirage either. Agents use only direct connectors. **Rejected.** Mirage's value proposition — unified bash vocabulary across heterogeneous backends (`grep -i 'refund' /slack/support/*.json | head -20` then `cat /github/api/refund.py`) — is real and useful at the agent layer for ingestion-heavy Onboarding Crew agents. Dropping entirely forfeits that benefit. Layer 5 keeps the option alive without committing the kernel.

## Why Reading 2 over Reading 1

Four reasons, ordered by structural weight:

1. **Language-purity discipline matches Mirage's deployment model.** Mirage is designed to live inside the host process. Reading 2 puts it there. Reading 1 fights the design.
2. **No permanent operational cost.** No sidecar to maintain, version-pin, secure, monitor, snapshot, scale, or recover. A v0.0.x dependency should not earn a permanent operational footprint.
3. **Destructive-action defenses shift cleanly.** The original plan listed three defenses for the destructive-action lockdown: read-only OAuth scopes, Temporal-signal approval, and Mirage snapshot-before-write. The third assumed Mirage was in the kernel. Under Reading 2: the **kernel doesn't snapshot data; agents do**, if they're using Mirage. The Temporal workflow gates destructive operations on the agent recording a pre-write snapshot; Postgres PITR covers Brain state separately. Defense responsibility shifts from kernel to agent code — a cleaner separation, not a weaker one.
4. **Probe replaced by reasoning.** Reading 1 would have required a multi-day live probe run (Mirage + sidecar against three thresholds) to make a decision the docs already make for us. Reading 2's argument is structural and verifiable in an afternoon. The pre-registered probe gate was the right discipline; it just turns out the structural mismatch was discoverable before any measurement, which means the probe never gets to run.

## What the PR #10 apparatus becomes

Not wasted. The `Workspace` interface, the five probe functions, the orthogonal failure-mode mocks, the 98.1% coverage discipline, the size-exception precedent, the coverage drill-down — all still real, still valuable. The apparatus is now a **general verification harness for any kernel-side data connector** (S3, Postgres, discrete MCP wrappers, future Layer 3 substrate candidates), not a Mirage-specific probe.

A follow-up PR will rename `kernel/probe/mirage/` to a name that reflects this generalization (e.g., `kernel/probe/connector/`). Out of scope for this PR — kept directory-stable here to keep PR #13 doc-only.

## Maker-checker provenance

This closeout exists because the **maker-checker discipline applied to plans, not just code.** The original plan named Mirage as Layer 3 substrate based on its marketing description (unified filesystem, multi-backend mounts, snapshot/rollback) without verifying its language surface against Galileo's kernel-language choice. That was a maker-side miss: the architects (Emmanuel + Claude) wrote the plan without reading Mirage's deployment-model documentation closely enough.

The checker pass happened during PR #13's inline-planning round when Mirage's docs were read directly. The language mismatch surfaced immediately ("no Go SDK; embed inside FastAPI/Express/browser"); the structural consequence (permanent sidecar or wrong placement) followed in one step.

**The discipline pattern this round produced for future use:** *"Read the dependency's deployment-model documentation before encoding it in the plan."* Marketing copy describes what a tool does. Deployment documentation describes how it fits into a host architecture. The latter is what determines whether a tool can occupy a given role in a host system. The two are different reads; the second is the one that prevents this class of plan-deviation.

What survived this rethink: the eight-layer architecture itself, the kernel/agent language partition (Go kernel, Python agents), the calibration-before-implementation discipline, the apparatus harness (generalized rather than replaced), the v7 build rules, and the destructive-action defense *categories* (read-only scopes, Temporal gating, pre-write snapshots) even though the third defense's implementation shifts from kernel to agent code. The plan's structural skeleton held. One specific substrate placement was wrong, and the correction was local.

## Downstream consequences (summary; full edits in this same PR)

- **Plan §Layer 3 (Brain substrate):** Mirage VFS removed. Replace with pgvector + AGE for Brain state, with discrete kernel-side connectors (S3, Postgres) wired through the Workspace interface from PR #10's apparatus.
- **Plan §Layer 5 (Integrations):** Mirage added as a documented agent-side option. Per-agent choice between Mirage's unified-filesystem abstraction and direct connectors. Cost and benefit named.
- **Plan §3.3 step 4 (Onboarding Crew Mount step):** reworded — Crawler Agent is a Python agent that *may* use Mirage if it wants. The crawl mechanism is agent-side, not kernel-side.
- **Plan §Destructive-action lockdown:** Defense #3 reworded — kernel Temporal-signal gates on the agent's pre-write snapshot artifact; Mirage is one tool an agent may use to produce that snapshot, not the kernel-level mechanism. Postgres PITR covers Brain state independently.
- **STAGE_0_PLAN.md §Week 2:** the live probe is no longer the right experiment. Week 2 instead delivers this closeout + plan edits + ADR; Week 2 exit criterion changes from "PROBE_MIRAGE_STAGE0.md committed" to "CLOSEOUT_LAYER3_MIRAGE_RECONSIDERED.md committed + plan + ADR reflect Reading 2."
- **`docs/decisions/0003-mirage-layer-relocation.md`:** new ADR captures the decision with reversal triggers.

## Reversal triggers (full list in the ADR; preview here)

This decision reverses if:

(a) Mirage publishes an **official Go SDK**, with stability commitments matching Galileo's kernel needs.
(b) Mirage publishes a **native server mode** exposing its API over a documented network protocol (HTTP, gRPC), maintained upstream — not a sidecar Galileo writes itself.
(c) Galileo's kernel acquires a need to crawl heterogeneous backends **itself** (i.e., not delegated to Python agents). This would change Galileo's architecture more than it would change Mirage's role.

If any trigger fires, the placement is re-evaluated against then-current Mirage state. The current closeout becomes historical context.

## v7 rule 3 framing

Closeouts exist for **every** architectural decision, including ones that reverse earlier plan assumptions. This closeout names the structural finding ("Mirage's deployment model is in-process-embedding; Layer 3 placement required a sidecar that was not named in the original plan"), not a softened or evasive version ("Mirage didn't fit Galileo" or "Mirage adoption postponed pending further evaluation"). Both alternatives would describe the same artifact (PR #13's closeout doc) while obscuring the actual structural mismatch and what was learned from it. The precision is what makes the closeout useful to future contributors who will read it when they wonder why the canonical plan was edited mid-Stage-0 and what the principle was.

## Template for future plan-deviations

Plan-deviations in this project follow the four-part pattern this PR establishes: (1) a closeout doc naming the structural finding and the reasoning chain that led to the deviation; (2) edits to the canonical plan reflecting the new structure; (3) an ADR with metadata, supersession info, and named reversal triggers; (4) a follow-up PR for any code changes the deviation implies (here, the eventual rename of `kernel/probe/mirage/` to a name reflecting its generalized role). Future deviations cite this PR as precedent.
