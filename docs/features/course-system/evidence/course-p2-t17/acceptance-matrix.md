# T17 Acceptance Matrix

| Frozen acceptance | Implemented | Verified | Evidence | Result |
|---|---:|---:|---:|---|
| `npx playwright test` covers admin create item on deep node → publish → learner view. | Yes | Yes: 11/11 full Playwright tests, T17 included | Test source, runtime record, screenshots, command log | PASS |
| Evidence under `docs/features/course-system/evidence/course-p2-t17/`. | Yes | Yes | This canonical bundle and SHA-256 manifest | PASS |

The acceptance wording, points, ID, and dependencies were not changed.

Additional repository gates also pass: local-environment proof, focused
Playwright, lint, focused LearningItem Vitest, production build, focused/full
Go tests, Compose configuration and web build, service health, PostgreSQL
correlation, signed-out denial, cleanup verification, screenshots, production
audit, double status synchronization, consistency check, and scoped diff check.

The full Vitest command remains the inherited baseline: 35 failed and 105
passed. No Course or LearningItem test fails.
