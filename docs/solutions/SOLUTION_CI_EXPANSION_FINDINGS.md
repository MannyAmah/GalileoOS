# SOLUTION — CI expansion findings (audit scoping + native-config preference)

Two findings surfaced by plan-PR #9's CI expansion on first run. Both were real signals, not gotchas; the CI surface doing its job is the point. The generalized lessons below outlast the specific defects.

## Finding 1 — Security audits scope to project dependencies, not runtime environment

**Concrete defect.** plan-PR #9's first run added `pip-audit --strict` (default: audit the entire installed environment via `pip freeze`) to the python job. The runner's installed env at that point included `black 24.10.0` (real CVE: GHSA-3936-cmfr-pm3m, fix in 26.3.1) plus `pip 26.0.1` itself (GHSA-58qw-9mgm-455v, GHSA-jp4c-xjxw-mgf9, fix in 26.1). The black CVE is real and actionable from this repo. The pip CVEs are in the GitHub Actions runner's own pip — no contributor can act on those from `MannyAmah/GalileoOS`.

**Fix applied.** (a) Bumped `black` pin 24.* → 26.* in `.github/workflows/ci.yml` and `.devcontainer/post-create.sh` per the [CI ↔ devcontainer co-change policy](../../CLAUDE.md#tool-version-pins-ci--devcontainer-co-change). (b) Re-scoped `pip-audit --strict` to the positional `project_path` argument (`pip-audit --strict .` from `agents/`) so it reads `agents/pyproject.toml`'s declared dependencies, not the runner's full env. Stage 0 declares zero runtime deps → no-op pass; Stage 1 adds LangGraph / CrewAI / Agno → activates meaningfully.

**Generalization.** When introducing audit checks (`pip-audit`, `npm audit`, `govulncheck`, future equivalents), **scope them to declared project dependencies**, not the runtime environment. Audits that include the runner's own tooling produce alerts the project cannot act on. Unactionable alerts become noise; noise gets ignored; the audit stops doing its job. The discipline is to keep the signal-to-noise ratio high enough that the first real CVE in a project dep is unambiguously the first thing the human sees.

**Concretely.** `pip-audit` takes a positional project path; pass `.` from the project's working directory. `npm audit --omit=dev` (already used in this repo) scopes to production deps. `govulncheck ./...` from a Go module audits only the module's transitive imports. None of these audit the runner.

## Finding 2 — Prefer native config formats over compatibility bridges

**Concrete defect.** plan-PR #9's first run used `eslint.config.mjs` that bridged the legacy `eslint-config-next` preset into ESLint 9's flat-config format via `@eslint/eslintrc`'s `FlatCompat`. ESLint exited with `TypeError: Converting circular structure to JSON ... property 'react' closes the circle` — `eslint-config-next`'s plugin self-references break `JSON.stringify` in the validator's error formatter. The bridge fails silently in the happy path and loudly when a plugin trips a corner case.

**Fix applied.** Replaced the FlatCompat bridge with a native ESLint 9 flat config — `@eslint/js` recommended + `typescript-eslint` strict + stylistic, no `eslint-config-next`. The two Stage 0 stub `.tsx` files (`layout.tsx`, `page.tsx`) have no Next.js-specific patterns to lint. When Stage 1 ships real Next.js code, `eslint-config-next` will likely have native flat-config support and is re-added as a single-line extension.

Side effect: ESLint flagged `next.config.js`'s `module.exports = nextConfig` (CommonJS) as `'module' is not defined`. Converted to `next.config.mjs` with `export default nextConfig`. Next.js 16 supports both; ESM keeps the lint config simple by not needing a CommonJS-globals special case.

**Generalization.** When introducing a linting or build toolchain, **prefer the tool's native config format over a compatibility bridge** to a legacy preset. Bridges (FlatCompat, polyfills, transpilers-as-config-loaders) work most of the time but fail in subtle ways when plugin authors use self-references, dynamic imports, or other patterns that don't survive serialization. Native config is more code today and less debt tomorrow. The cost of writing 20 lines of native config now is much smaller than the cost of debugging a bridge's edge case under deadline pressure later.

**Concretely.** For ESLint 9: skip `FlatCompat`. Use `@eslint/js` + framework-specific TypeScript-ESLint configs directly. When a preset has native flat-config support (Next.js will; many ecosystem plugins already do), re-add as a single-line extension. For Vite + TypeScript: prefer `vite.config.ts` over `vite.config.js` so the type-check is part of the build. For Python toolchains migrating from `setup.py` → `pyproject.toml`: prefer the native PEP 621 metadata over `setuptools.config.expand_section_loaders` patterns.

## Meta-finding — Expanded CI surfacing findings on first run is the surface working

A CI expansion PR that *didn't* surface findings on first run would be more suspicious than one that did. `pip-audit` caught a real CVE; ESLint caught a real toolchain incompatibility. The checks are checking. Both findings + their fixes shipped in the same PR (plan-PR #9) per the v7-rule-7 compounding-the-lesson discipline.

## Origin

plan-PR #9 → first CI run surfaced both findings → fixes applied in-place (iteration within stated scope, not bundling) → this solutions doc captures the lessons. 2026-05-13.
