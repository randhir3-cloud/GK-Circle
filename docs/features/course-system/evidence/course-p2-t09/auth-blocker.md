# COURSE-P2-T09 Auth Blocker — RESOLVED

## Status

**RESOLVED** on 2026-07-27 under the Project Completion Rule.

Allowlisted admin runtime verification completed using the local Course-admin
seed (`npm run seed:local-course-admin`). See
`docs/development/local-course-admin.md` and `technical-verification.json`.

## Original blocker (historical)

Authenticated **allowlisted** admin session credentials were unavailable in the
execution environment, so Course Content / Learning Items runtime CRUD could
not be completed.

## Attempted mechanisms (pre-fix)

1. **Playwright storageState** — no `storageState` / auth state file present.
2. **Environment-provided E2E credentials** — `E2E_*` / Playwright password
   vars unset in the agent process.
3. **CI / seeded allowlisted fixtures** — not initially supplied.
4. **Project registration + login automation** — succeeded for a generic
   authenticated `/admin` session, but the ephemeral identity lacked
   `can_create_public_quiz` / allowlist membership.

## Root cause

`PUBLIC_QUIZ_ADMIN_EMAILS` was unset/empty in the running API for local
verification, and no documented local allowlisted identity existed.

## Fix implemented (in scope)

- Seed script: `scripts/seed-local-course-admin.cjs`
- npm script: `seed:local-course-admin`
- Docs: `docs/development/local-course-admin.md`
- Gitignored local env snippets for API allowlist + E2E vars
- Compose override optional `env_file` for `api/.env.local.course-admin`
- Local deterministic identity (development only):
  `local.course.admin@gk-circle.local`

Credentials were not invented for production, and authentication security was
not weakened.

## Post-fix verification

Allowlisted Course Content entry, deep-node drill-down, server-order CRUD,
desktop/mobile screenshots, and signed-out denial all **PASSED**. Task status:
**VERIFIED**.
