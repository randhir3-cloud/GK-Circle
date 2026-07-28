# GK Circle Repository Architecture Assessment

- Task: `COURSE-P1-T01`
- Assessment date: 2026-07-25
- Repository: `randhir3-cloud/GK-Circle-v2`
- Branch at assessment: `chore/ci-verification`
- Commit at assessment: `eeac599f05eaf936c7f61db4a3deeac3c9063f59`
- Scope: read-only repository architecture assessment

This is the canonical repository assessment. It records detected repository
behavior separately from work implemented by the Course System initiative.

## Backend and data

| Conclusion | Direct repository evidence |
|---|---|
| The API is Go 1.23 with Fiber v2. | `api/go.mod:3` declares Go 1.23; `api/go.mod:13` declares Fiber v2.52.12; `api/cli/api.go:31` constructs the Fiber application. |
| PostgreSQL is the configured primary application database, using `lib/pq`. | `api/go.mod:19` declares `lib/pq`; `api/database/main.go:75` opens the PostgreSQL connection. |
| goqu is the query builder. | `api/go.mod:7` declares goqu v9.18.0; `api/database/main.go:79-81` wraps the PostgreSQL connection with goqu. |
| Migrations use `rubenv/sql-migrate`, with both up and down execution paths. | `api/go.mod:24` declares sql-migrate; `api/cli/migration.go:152` runs PostgreSQL up migrations and `api/cli/migration.go:157` runs down migrations; SQL files are stored in `api/database/migrations/`. |
| The local migration entrypoint is the Compose `migration` service. | `docker-compose.yaml:21-32` defines the service and runs the `gk-circle migrate up` command. |
| Ory Kratos supplies authentication/session identity. | `api/pkg/kratos/kratos.yml` is the Kratos configuration; `api/middlewares/authenticated.go:72-81` validates the request through Kratos `/whoami`; Compose declares Kratos services at `docker-compose.yaml:74-101`. |
| Authorization remains application-specific and server-side. | `api/routes/main.go:245-370` applies authentication and permission middleware to protected groups; `api/middlewares/quiz_permission.go:41` checks the configured public-quiz administrator policy; `api/config/quiz.go:10-17` defines that policy. |
| API JSON responses use shared JSend-style helpers. | `api/utils/json_response.go:10-23` defines success, fail, and error helpers; `api/controllers/api/v1/quiz_category_controller.go:65-96` demonstrates their controller use. |
| Existing assessment foundations are Quiz, Question, and QuizCategory models. | `api/models/quiz.go:18` defines `QuizModel`; `api/models/questions.go:31` and `api/models/questions.go:62` define `Question` and `QuestionModel`; `api/models/quiz_category.go:15` defines `QuizCategoryModel`. |

No ORM or Prisma runtime was detected. The active data-access convention is
SQL migrations plus goqu-backed Go models.

## Frontend

| Conclusion | Direct repository evidence |
|---|---|
| The web application is Nuxt 3 with Vue 3 conventions. | `app/package.json:43` declares Nuxt `^3.14.1592`; `app/nuxt.config.ts` is the application configuration; route components are present under `app/pages/`. |
| Pinia is the state-management convention. | `app/package.json:52` declares the Nuxt Pinia integration and `app/package.json:63` declares Pinia; `app/nuxt.config.ts:91-99` registers the module. |
| Tailwind and shadcn-nuxt are established UI tooling. | `app/nuxt.config.ts:2` imports the Tailwind Vite plugin, `app/nuxt.config.ts:99` registers shadcn-nuxt, and `app/nuxt.config.ts:170` activates Tailwind. |
| Frontend scripts are npm-based within `app/`. | `app/package.json:6-14` defines build, lint, and Vitest commands. |

## Testing, Compose, and CI

| Conclusion | Direct repository evidence |
|---|---|
| Backend tests use Go testing with Testify available. | `api/go.mod:28` declares Testify; Go test files are colocated under `api/`. |
| Frontend unit/component tests use Vitest and Nuxt test utilities. | `app/package.json:14` maps `npm test` to Vitest; `app/package.json:48` declares Vitest; `app/vitest.config.ts:1-4` defines the Nuxt Vitest configuration. |
| Browser E2E tests use Playwright. | `app/package.json:27` declares `@playwright/test`; `app/playwright.config.ts:1-30` defines the Playwright projects. |
| The local stack contains PostgreSQL, migration, Redis, API, web, Kratos, MinIO, Mailpit, and API verification services. | Service declarations are in `docker-compose.yaml:2-146`; PostgreSQL uses the 15.2 Alpine image at line 3 and Kratos uses v1.0.0 at lines 75 and 86. |
| CI runs for pull requests and pushes, and includes API verification, frontend lint/build, and Playwright. | `.github/workflows/verify.yaml:3-5` defines triggers; lines 83-84 run the API verification service; lines 109-113 run frontend lint/build; lines 115-201 define Playwright execution and artifacts. |

## Detected architecture versus Course System work

### Detected

- Nuxt 3 web application.
- Go/Fiber API.
- PostgreSQL with goqu and sql-migrate.
- Ory Kratos authentication.
- Redis-compatible cache and live coordination.
- MinIO object storage.
- Vitest and Playwright frontend testing.
- Docker Compose verification profile and GitHub Actions CI.

### Implemented by Course System in this task

- Persistent development ledger and weighted task tracking.
- Read-only check and deterministic sync commands.
- Canonical assessment, baseline, and checker-integrity evidence.
- Handoff/resume protocol.

### Explicitly not implemented

- Course or CourseNode Go models.
- `courses` or `course_nodes` tables.
- Course database migrations.
- Course APIs, authorization policies, pages, stores, or components.
- Phase 2 functionality.

A repository search of `api/` and `app/` for `CourseNode`, `course_nodes`, and
Course table definitions returned no active implementation at assessment time.

## Standards findings requiring later resolution

These findings do not authorize an architecture change:

1. `docs/standards/course-rules.md:41-52` defines the canonical hierarchy as
   `Course -> CourseSubject -> CourseTopic -> TopicContent`, while the current
   Phase 1 ledger proposes a generic `CourseNode`. Before `COURSE-P1-T03`, an
   approved ADR or ledger correction is required if those structures are not
   demonstrably compatible.
2. `docs/standards/course-rules.md:170` requires an ADR for future hierarchy
   layers, reinforcing the need to resolve the CourseNode boundary before its
   implementation.
3. `docs/standards/course-rules.md:531` refers to Prisma even though
   `AGENTS.md:33`, `CLAUDE.MD:48`, and the repository implementation establish
   Go, goqu, and sql-migrate. Per `docs/standards/index.md:30`, the actual
   architecture and higher-priority maintainer guidance control.
4. `docs/standards/course-rules.md:514` says Course terminology is reserved for
   a future marketplace domain, conflicting with the same standard's active
   Course-domain rules. This wording must be clarified separately.

## Assessment boundary

`COURSE-P1-T02` remains `NOT_STARTED`. Its database design and migration require
explicit approval. No production or NUC environment was accessed during this
assessment.
