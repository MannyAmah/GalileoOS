# SOLUTION: Coverage drill-down — four buckets, four treatments

Produced during plan-PR #10 (Mirage probe apparatus). The apparatus's test coverage came in at 89.6% — 0.4% under the >90% bar a previous turn had set for "the indivisibility claim is structurally honest." A uniform response ("everything under 90% gets a test" or "everything under 90% gets removed") would have been wrong: some of the uncovered lines were structurally untestable without changing the API, others were defensive guards worth a one-line boundary test, and at least one was a missing-failure-mode diagnostic that needed a new mock.

The discipline below is the generalization. Apply it to any Galileo PR whose coverage report comes in below the 90% bar.

## The four-bucket pattern

When package coverage falls below the bar, categorize uncovered lines into structural buckets *before* deciding treatment. Each bucket has a specific correct answer.

### Bucket 1 — error propagation from dependency methods

Pattern: `err := dep.Method(...); if err != nil { return wrap(err) }`. The branch is unreachable because the mocks don't return errors.

Treatment: **test**. Add a mock that returns a sentinel error from each dependency method, plus one test per probe asserting `err != nil` and that the returned error wraps with a method-naming prefix. Cost is uniform (~10 lines per probe). This is the most consequential uncovered category: the error-propagation contract is part of the apparatus's public surface, and the maker/checker discipline depends on it being verified independently of the dependency it measures. Without these tests, a real-backend failure could be attributed to the backend when the truth was apparatus mishandling.

### Bucket 2 — defensive input validation

Pattern: `if len(arg) == 0 { return errors.New("...") }` at the top of an exported function, or a short-content guard inside a parser.

Treatment: **test**. Two lines per guard: call the function with a deliberately-malformed input, assert the specific error. Cost is trivial; precedent matters. An apparatus that documents "we test every branch" is more trustworthy than one that documents "we test the branches that matter."

### Bucket 3 — structurally untestable without API change

Pattern: a fallback whose precondition cannot occur on supported runtimes (e.g., `crypto/rand.Read` failure), or a path that requires dependency injection to exercise.

Treatment: **document, not test**. Three options exist — (a) inject the dependency as a parameter, (b) document the gap with the structural reason, (c) remove the fallback. Option (a) changes API surface for ~1% coverage gain; usually wrong. Option (c) loses correct defensive behavior; usually wrong. Option (b) is the right call: add a doc comment on the function explaining why the branch is untested, what it would take to test it, and why the cost-benefit doesn't justify the test. The honest discipline is that the apparatus doesn't claim 100% coverage; it claims every uncovered line has a structural justification documented.

### Bucket 4 — missing failure-mode diagnostic

Pattern: a branch that exists to report a specific named failure mode, but no mock currently triggers that failure mode. Example: the "file missing post-restore" diagnostic in `RunSnapshotProbe` — distinct from byte-drift because the failure shape is "file dropped" rather than "file mutated."

Treatment: **add the mock and test**. Not testing it means the apparatus is structurally blind to one specific real-backend failure mode it claims to produce diagnostic output for. Cost is ~20 lines per missing mode (one mock + one test). This is the bucket most likely to be missed during initial implementation; it surfaces during coverage drill-down because the diagnostic branch is never executed.

## How to use this in practice

1. Run `go test -coverprofile=/tmp/cover.out ./...` then `go tool cover -func=/tmp/cover.out`.
2. For each function below the bar, run `awk '$NF == "0" {print}' /tmp/cover.out` (or open `go tool cover -html=/tmp/cover.out`) to identify the specific uncovered blocks.
3. For each uncovered block, categorize into one of the four buckets.
4. Apply the bucket-specific treatment.
5. Re-run coverage. Confirm everything not in Bucket 3 is now covered; Bucket 3 items must each have a doc comment naming the structural reason.
6. The PR description names the residual coverage gap explicitly and points readers to the doc comments on Bucket 3 functions.

## Structural lesson

The coverage drill-down is more important than the coverage number. "We're at 89.6%" is not actionable information; "we're at 89.6% and here are the four buckets the uncovered lines fall into" is. Uniform rules across all uncovered code (either "test everything" or "skip everything below threshold") are usually wrong because the four buckets have four different correct treatments. A discipline that produces uniform answers across heterogeneous categories is degraded discipline.

The pattern is reusable for any Go package — the buckets are language-agnostic in shape, though the specific patterns (Workspace error propagation, `crypto/rand` fallback) are Go-flavored.
