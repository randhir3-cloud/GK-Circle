# Local Course Admin Seed (development only)

Purpose: unblock authenticated Course System admin verification (e.g. COURSE-P2-T09)
with a deterministic local allowlisted identity. Not for production or NUC.

## Identity

| Field | Value |
|---|---|
| Email | `local.course.admin@gk-circle.local` |
| Password | `LocalDev!CourseAdmin1` (override with `LOCAL_COURSE_ADMIN_PASSWORD`) |
| Capability | Listed in `PUBLIC_QUIZ_ADMIN_EMAILS` → `canCreatePublicQuiz` / Course Content nav |

## Commands

```bash
node scripts/seed-local-course-admin.cjs
docker compose up -d api
```

Generated gitignored files:

- `api/.env.local.course-admin` — sets `PUBLIC_QUIZ_ADMIN_EMAILS`
- `.env.e2e.local` — Playwright / agent E2E variables (`E2E_CREATOR_EMAIL`, `E2E_TEST_PASSWORD`, …)

Compose loads the API snippet via `docker-compose.override.yml` (`required: false`).

## Agent / Playwright login

1. Run the seed + recreate/restart `api`.
2. Load `.env.e2e.local` into the shell (or read `E2E_CREATOR_EMAIL` / `E2E_TEST_PASSWORD` from it for automation).
3. Sign in at `/account/login` with that identity.
4. Confirm sidebar shows **Course Content** → `/admin/courses/learning-items`.

Do not invent alternate production credentials. Do not commit the generated env snippets.
