# SOLUTION: Dependency upgrades are chains, not single-version changes

Produced during PR #11 (Go pin bump 1.23 → Path W: runtime 1.26, `go.mod` 1.25, golangci-lint v2.12.2, action @v7). What was scoped as a one-line `go-version` bump took **three CI iterations** to land — each iteration revealed a downstream coupling the prior bump had hidden.

The discipline below is the generalization. Apply it to any version-pin bump going forward.

## The coupling chain from PR #11 (worked example)

Three links, each surfaced only after the prior link's bump landed and CI ran.

### Link 1 — runtime → language baseline

**Trigger:** PR #10's apparatus introduced the first Go caller transitively reaching 6 stdlib CVEs unfixed in Go 1.23.x: `GO-2025-4007` (crypto/x509), `GO-2025-4009` (encoding/pem), `GO-2025-4010` (net/url), `GO-2025-4011` (encoding/asn1), `GO-2026-4601` (net/url), `GO-2026-4602` (os).

**Proposed change:** bump `actions/setup-go@v5 go-version: 1.23 → 1.26`.

**Hidden coupling:** `kernel/go.mod`'s `go` directive controls language semantics independently of the installed runtime. Leaving it at 1.23 was technically viable (runtime/language split), but committed the project to forfeiting Go 1.24/1.25 language features. **Latest-1 posture** chosen: runtime tracks current stable, `go.mod` tracks one major behind, so the project gets stdlib CVE fixes at upstream cadence and language features at ecosystem-tooling cadence.

### Link 2 — language baseline → golangci-lint binary build-version

**Trigger:** Bumping `go.mod`'s `go` directive past golangci-lint v1.62.0's build-Go version (1.23).

**Failure symptom in CI:** `can't load config: the Go language version (go1.23) used to build golangci-lint is lower than the targeted Go version (1.25.0)`.

**Hidden coupling:** golangci-lint's documented "latest-1" build policy means its binary build-Go must be ≥ the target's `go.mod` directive. Bumped golangci-lint v1.62.0 → v2.12.2 (built with Go 1.25.0). v2 is a major version bump from v1 — verified locally that no `.golangci.yml` config migration was needed and that the broader v2 default linter set fired only one warning against existing code (handled separately).

### Link 3 — golangci-lint binary version → GitHub Action wrapper compatibility

**Trigger:** Bumping the binary from v1 to v2.

**Failure symptom in CI:** `invalid version string 'v2.12.2', golangci-lint v2 is not supported by golangci-lint-action v6, you must update to golangci-lint-action v7`.

**Hidden coupling:** The GitHub Action wrapper (`golangci/golangci-lint-action`) is versioned independently of the binary. Major Action versions track major binary versions. Bumped `@v6 → @v7`.

## The structural insight

**Dependency upgrades are best modeled as a chain, not a single change.** The chain may be one link long; often it's two or three; in adversarial cases it's longer. The cost of treating a bump as one link when it's actually three is one CI run per missed link — informative but wasteful.

Each link in the chain has the same shape: "upgrading X requires Y, which requires Z." Each link's symptom is a specific error string visible only at CI time, not at planning time, because the coupling lives in the tooling chain's documentation rather than in the version specifier itself.

## The 5-minute upfront check

Before scoping a version bump as a single-line PR, ask:

1. **What does the thing I'm upgrading transitively depend on?** (Runtime → language semantics → tooling that consumes language semantics → wrappers around that tooling.)
2. **What depends on the thing I'm upgrading?** (Reverse direction — anything pinned to a specific major version of what I'm bumping.)
3. **Are any of those dependencies pinned by major version in the repo?** (Pin tables, action `@vN` references, build-time policies documented in upstream READMEs.)
4. **Has the upstream of any of those dependencies released a new major to keep pace?** (golangci-lint v1 → v2 happened because Go versions kept moving; the Action wrapper follows the binary.)

A 5-minute scan of upstream READMEs and the repo's pin table catches most of the chain. The remainder surfaces at CI time, and the discipline is to treat each surfaced link as a coupling discovery rather than a one-off fix — name it, decide on the treatment deliberately, document it.

## When this lesson applies vs doesn't

Applies to runtime/build-tool/Action-wrapper version bumps and major-version bumps of anything the build chain reads. Doesn't apply to bug-fix patches of leaf dependencies (no downstream policy attached), application code, or doc-only changes.

## Cross-references

- PR #11: the original incident, with full iteration history in its description.
- CLAUDE.md `### Latest-1 language posture`: the policy adopted as a consequence of Link 1's coupling.
- `docs/solutions/SOLUTION_COVERAGE_DRILL_DOWN.md`: a parallel "categorize before treating" discipline produced during PR #10.
