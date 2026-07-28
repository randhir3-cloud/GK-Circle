# COURSE-P2-T13 Technical Verification

| Check | Result |
|---|---|
| ALL retained; AUTHENTICATED kept when authenticated / omitted when not | PASSED |
| HIDDEN / INSTRUCTOR removed | PASSED |
| PREMIUM omitted when unauthorized; retained when `PremiumAuthorized` | PASSED |
| Mixed document order preserved; empty / all-hidden → non-nil `[]` with version | PASSED |
| Deep-copy / non-mutation of original `json.RawMessage` | PASSED |
| List projection atomic failure (no partial list) | PASSED |
| Unknown mode omitted; malformed metadata → wrapped error | PASSED |
| Admin retrieval unfiltered; SQL draft publish filter still 404 | PASSED |
| Production defaults `Authenticated=true`, `PremiumAuthorized=false` | PASSED |
| Learner controller: hidden/instructor/premium absent; auth present; order preserved | PASSED |
| Admin contrast: all blocks present | PASSED |
| Controllers have no visibility switches / block iteration | PASSED |
| Local `go test ./models/ -run LearningItemVisibilityRuntime` | PASSED |
| Local `go test ./controllers/api/v1/ -run "LearningItemVisibility\|LearnerLearningItem"` | PASSED |
| Local `go test ./...` | PASSED |
| Docker `api-verify go vet ./...` | PASSED |
| Docker `api-verify go test ./...` | PASSED |
| `docker compose config --quiet` | PASSED |
| Ledger sync×2 + check | PASSED |
| Frontend | NOT_APPLICABLE |
| Database migration | NONE |

## Notes

- Projection is repository-owned after SQL publish filtering.
- No Course premium entitlement exists; production learner access keeps `PremiumAuthorized=false`.
- COURSE-P2-T12 enrollment remains BLOCKED and was not started.
