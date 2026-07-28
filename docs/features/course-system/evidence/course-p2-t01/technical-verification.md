# COURSE-P2-T01 Technical Verification

- Date: 2026-07-26
- Branch: `chore/ci-verification`
- Commit: `eeac599f05eaf936c7f61db4a3deeac3c9063f59`
- Database Migration: ONE (`learning_items`)
- HTTP APIs: No
- Production/NUC: No
- Breaking Changes: NO

## Checks

| Check | Result |
|---|---|
| Isolated PostgreSQL apply / inspect / rollback / reapply | PASSED |
| Course/CourseNode data unchanged across migration cycle | PASSED |
| `docker compose config --quiet` | PASSED |
| Docker `api-verify` vet / test / race / build | PASSED (shared with T02) |
| Ledger sync/check after evidence promotion | PASSED |
