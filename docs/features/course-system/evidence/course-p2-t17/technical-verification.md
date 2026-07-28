# T17 Technical Verification

| Check | Result | Objective outcome |
|---|---|---|
| Environment identity | PASS | Web/API are localhost; Compose project `gkcirclev2`; DB has no host port |
| Focused T17 Playwright | PASS | 1/1, fresh run, no retry |
| Frozen full Playwright | PASS | 11/11, fresh run, no retries |
| Focused Course/LearningItem Vitest | PASS | 9 files, 63 tests |
| Full Vitest | INHERITED BASELINE | 35 failed, 105 passed; no LearningItem failure |
| Frontend lint | PASS | Exit 0 |
| Nuxt production build | PASS | Build complete |
| Focused learner backend | PASS | Controller and model packages |
| Full Go suite | PASS | `go test ./...` |
| Compose configuration | PASS | `docker compose config --quiet` |
| Compose web build | PASS | Image `gkcirclev2-web:latest` built |
| Local service health | PASS | API, web, Kratos, DB, Redis, Mailpit healthy |
| Persistence correlation | PASS | Four-node ancestor chain; real API/UI values |
| Cleanup | PASS | DELETE 200, GET 404, final DB count 0 |
| Desktop/mobile screenshots | PASS | Six stable-name PNGs |
| Production audit | PASS | Production source modified: NO |

## Failure investigation

The first executable focused attempt timed out because the signed-out learner
page displays its API 401 inline rather than redirecting to login. The retry
reproduced the timeout. Cleanup was then performed against the two exact
test-created IDs after validating their `T17 Learning Item` titles; both
returned 404 afterward.

The next attempt reached the final denial assertion and showed that the
established middleware message does not contain the word `unauthenticated`.
The test was corrected to verify the authoritative HTTP 401 and a visible,
nonempty alert without freezing middleware prose.

After those deterministic test corrections, a fresh focused run and two fresh
full runs passed without retry-only recovery. The preserved timeout trace is
under `traces/`.

## Decision

All frozen acceptance and repository completion gates applicable to T17 pass.
COURSE-P2-T17 is VERIFIED.
