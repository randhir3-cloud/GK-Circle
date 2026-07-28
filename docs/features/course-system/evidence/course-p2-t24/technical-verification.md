# COURSE-P2-T24 Technical Verification

| Check | Result |
|---|---|
| Learner published list excludes drafts; empty all-draft/empty node | PASSED |
| Learner published get: published OK; draft/missing/wrong-scope → not-found | PASSED |
| Admin list/get still return drafts | PASSED |
| Learner HTTP list has no draft ID/title/`publish_state=DRAFT` | PASSED |
| Learner HTTP get draft≡missing public 404 fields | PASSED |
| Admin HTTP list/get drafts; non-admin forbidden | PASSED |
| Local `go test ./models/ -run LearningItemDraftFiltering` | PASSED |
| Local `go test ./controllers/api/v1/ -run LearningItemDraftFiltering` | PASSED |
| Local `go test ./...` | PASSED |
| Docker `api-verify go vet ./...` | PASSED |
| Docker `api-verify go test ./...` | PASSED |
| Ledger sync/check + no-op second sync | PASSED |
| Frontend | NOT_APPLICABLE |
| Database migration | NONE |

## Notes

- Publish filtering remains repository-owned (`ListPublished*` / `GetPublished*`).
- No production model/controller changes were required; existing T07/T08 contract held.
