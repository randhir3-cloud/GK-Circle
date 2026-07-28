# T16 Technical Verification

| Check | Result | Objective evidence |
|---|---|---|
| Focused frontend tests | PASS | 9 files, 63 tests |
| Frontend lint | PASS | Exit 0 after formatting only T16-owned tests |
| Full frontend suite | INHERITED BASELINE | 35 failed, 105 passed; same 17 unrelated legacy files and failure count; all T16 tests green |
| Nuxt production build | PASS | Exit 0 |
| Full Go suite | PASS | `go test ./...` |
| Compose config | PASS | `docker compose config --quiet` |
| Compose web build | PASS | `gkcirclev2-web:latest` built |
| Local service health | PASS | API, web, Kratos, PostgreSQL, Redis healthy |
| Authenticated runtime | PASS | Documented local Kratos QA identity |
| Admin persistence | PASS | Create, refresh read-back, update, cancel delete, confirmed delete |
| Learner contract | PASS | Published item visible; draft sibling excluded; detail renderer empty state |
| PostgreSQL correlation | PASS | Final deep-node rows exactly positions 0/1; temporary position-2 row removed |
| Responsive/accessibility | PASS | Desktop and 360px viewport; semantic headings, labels, table, links, buttons |
| Browser console | PASS | No warning or error entries |
| Signed-out denial | PASS | Protected detail redirected away after logout |
| Production change audit | PASS | Production source modified: NO |

## Full-suite classification

The pre-T16 baseline was 35 failed / 96 passed. T16 adds two test files and nine
passing tests, producing 35 failed / 105 passed. The failure count and the 17
legacy quiz/report test files are unchanged and outside the Course/LearningItem
surface. The frozen composition and renderer suites pass completely.

## Renderer fixture note

The retained deep-node learner item has a valid empty block array and therefore
renders `No content available.` in the current smoke. Supported block dispatch,
safe links, malformed/unsupported fallbacks, immutability, and responsive
contracts pass in the focused renderer/URL suites. The authenticated
representative-block runtime remains independently sealed in T14 evidence.
