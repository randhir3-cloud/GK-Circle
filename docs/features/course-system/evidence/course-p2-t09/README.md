# COURSE-P2-T09 Canonical Evidence

Task: `COURSE-P2-T09 — Admin Item Composition UI`

Status: VERIFIED

- Implemented on: 2026-07-26
- Verified on: 2026-07-27 (CLOSE-02 + Project Completion Rule unblock)
- Branch: `chore/ci-verification`
- Commit at implementation: `eeac599f05eaf936c7f61db4a3deeac3c9063f59` plus working-tree changes
- Database Migration: NONE (T09 UI-only; local stack applied existing Course System migrations)
- Backend/API changes for T09 feature: NONE
- Local verification infrastructure added: seeded course-admin identity + allowlist wiring
- Production/NUC touched: No
- Breaking Changes: NO
- Production Behaviour Changes: NONE

## Unblock path

Authenticated allowlisted runtime was blocked by missing local test identity /
empty `PUBLIC_QUIZ_ADMIN_EMAILS` and an API image/migrations gap. Under the
Project Completion Rule these in-scope blockers were fixed:

1. `npm run seed:local-course-admin` — deterministic local allowlisted admin
2. Compose override loads `api/.env.local.course-admin`
3. API image rebuilt with Course routes
4. Existing Course System migrations applied locally
5. Runtime CRUD + screenshots completed with the seeded account

Docs: `docs/development/local-course-admin.md`

Evidence:

- `closure-audit.md` / `acceptance-closure-matrix.md`
- `technical-verification.md` / `technical-verification.json`
- `runtime-audit.md` / `auth-blocker.md` (historical blocker + resolution)
- `screenshots.md` + `screenshots/`
- `commands/` + `hashes/sha256.txt`
