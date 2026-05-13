# SOLUTION — CI guard verification

**Defect chain.** The `buf breaking` shell guard introduced in PR #5 had a latent bug in its `--against` ref syntax (`'.git#branch=main'`, which resolves to a local `refs/heads/main` that `actions/checkout` never creates on PR runners). Because the guard exited 0 on the introducing PR (no `origin/main` schema baseline existed yet), the broken code path inside the `then` branch never executed and CI passed. PR #6 was the first PR after merge to fall through the guard; the bug surfaced immediately. Fix: PR #7.

**First generalization (guard verification).** Any guard pattern that conditionally skips a code path on the introducing PR must include a sentinel test that exercises the *non-skip* branch with synthetic inputs, so the guard's protected path is verified at least once before the guard is relied on in production.

**Second generalization (workflow blast radius).** CI workflow files execute from the PR branch, not from main. A defect in `.github/workflows/*.yml` therefore propagates to every concurrently-open PR until each one is rebased onto post-fix main. The blast radius of a bad workflow change is every open PR, not just the one that introduced the defect. The operational implication: CI workflow changes are queue-flushing — open them at the front of the queue and merge before any other PR opens behind them.

**Concretely.** When adding a new guarded CI check, also add a one-shot sentinel job (or a unit test against the guard's body) that forces the else branch to execute during the PR that introduces the guard. For `buf breaking`, that would have meant either pre-staging a sentinel `.proto` file on `main` before opening PR #5, or running a one-off CI job that simulates a non-empty `origin/main` schema state. Neither was done; the result was a working-from-luck guard that broke on the very next PR — and then broke a second concurrent PR (this SOLUTION doc's own PR #8) because that PR was branched from pre-fix main and ran the same buggy workflow.

**Origin.** Stage 0 PR #5 → surfaced on PR #6 → propagated to PR #8 → fixed in PR #7. 2026-05-12 → 2026-05-13.
