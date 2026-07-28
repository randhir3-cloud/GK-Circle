# Baseline Verification

- Run ID: `2026-07-25_course-p1-t01-baseline`
- Branch: `chore/ci-verification`
- Commit: `eeac599f05eaf936c7f61db4a3deeac3c9063f59`
- Environment: local Windows, local Node, local Docker Compose
- Production/NUC accessed: No

| Command | Result | Exit | Duration |
|---|---:|---:|---:|
| `git diff --check` | PASSED | 0 | 88 ms |
| `docker compose config --quiet` | PASSED | 0 | 384 ms |
| Local Docker `api-verify` vet/test/race command | PASSED | 0 | 15,444 ms |
| `npm run lint` in `app` | PASSED | 0 | 9,518 ms |
| `npm test -- --run` in `app` | FAILED_BASELINE | 1 | 21,310 ms |
| `npm run build` in `app` | PASSED | 0 | 406,851 ms |
| Local Playwright run with isolated output | PASSED | 0 | 26,051 ms |
| Course migration apply/rollback | NOT_AVAILABLE | — | — |

Vitest reports 17 failing files and 35 failing tests, with 7 files and 42
tests passing. This is classified `FAILED_BASELINE` because the ledger patch
changes no tracked file under `app/`; every reported failure resolves to
existing application or test paths. It does not block the ledger assessment,
but it remains a repository-health issue for later work.

The first Docker API and Playwright attempts were blocked by sandbox access to
the Docker named pipe and Chromium process launch respectively. Approved local
runs succeeded. Exact timestamps, environments, and notes are recorded in
`baseline-verification.json`.

No Course migration command was run because no Course migration exists and the
plan expressly forbids creating one.
