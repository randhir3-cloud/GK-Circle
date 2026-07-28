# COURSE-P2-T05 Changed Files

## Authorized

- `api/models/learning_item_visibility.go` (new)
- `api/models/learning_item_visibility_test.go` (new)
- `api/models/learning_item_metadata.go` (block visibility field + normalize wiring)
- `api/models/learning_item.go` (visibility semantic errors)
- `api/models/learning_item_test.go` (create/update visibility cases + canonical metadata expectations)
- Course System ledger / status surfaces for T05 verification
- Evidence under `docs/features/course-system/evidence/course-p2-t05/`

## Explicitly excluded

- Runtime visibility evaluation / authorization
- Controllers, routes, DTOs, frontend, rendering
- Database migrations
- Course / CourseNode hierarchy behaviour changes
- Production / NUC deployment
- Phase 1 UI (`COURSE-P1-T11`)
