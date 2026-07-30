# GK Circle Maintainer Guide

Version: 3.2

Status: Mandatory

GK Circle is a production-oriented State PCS examination preparation platform built from the open-source jovVix foundation. Maintainers must preserve the working architecture, the AGPL-3.0 licence, and upstream attribution while evolving the product for GK Circle.

## Implementation gate

Before changing the repository:

1. Read this file and `CLAUDE.MD`.
2. Read `docs/standards/index.md` and follow its mandatory reading order.
3. Before implementation or verification, report the standards loaded, the task-specific standards, and any conflict or stop condition.
4. Inspect the current implementation and search before creating parallel code.
5. Write a scoped plan. Obtain approval for architectural, database, deployment, security, or destructive work.
6. Verify before claiming completion.

## Actual architecture

| Area | Technology | Location |
|---|---|---|
| Web application | Nuxt 3, Vue 3, TypeScript/JavaScript, Pinia | `app/` |
| API | Go 1.25, Fiber v2 | `api/` |
| Database | PostgreSQL 15, SQL migrations | `api/database/migrations/` |
| Authentication | Ory Kratos | `api/pkg/kratos/` |
| Cache and coordination | Redis-compatible service | Compose service `redis` |
| Live interaction | WebSockets | Go API and Nuxt client |
| Object storage | MinIO | Compose service `minio` |
| Local email | Mailpit | Compose service `mailpit` |
| Runtime | Docker Compose | repository root |

Do not introduce NestJS, Next.js, Prisma, or a second authentication/test engine unless an approved ADR explicitly replaces the current architecture.

## Product scope

GK Circle adapts the existing quiz engine into reusable examination modes:

- PCS practice sets
- Full-length mock examinations
- Subject and topic tests
- Previous-year-question practice
- Current-affairs tests
- Timed live competitions
- Rankings, answer review, and performance analytics

The initial platform is State-PCS capable rather than hardcoded to one commission. Commission, state, exam stage, subject, topic, language, year, and difficulty belong in configurable taxonomy or metadata.

## Repository rules

- Extend the existing quiz, question, session, answer, scoring, authentication, and reporting systems.
- Do not duplicate core systems.
- Keep API business rules in Go; the Nuxt client must not become the source of truth for permissions or scoring.
- Use additive SQL migrations. Never edit an applied migration.
- Never run destructive migrations, delete production volumes, or replace production data without an approved backup and rollback plan.
- Do not commit `.env`, credentials, tokens, private keys, database dumps, generated build output, or local agent tooling.
- Example configuration must use obvious placeholders.
- Preserve the existing `upstream` remote for `Improwised/jovVix`.
- Preserve `LICENSE.txt` and original attribution required by AGPL-3.0.

## Naming

- Product: `GK Circle`
- Repository: `randhir3-cloud/GK-Circle`
- Go module: `github.com/randhir3-cloud/GK-Circle-v2/api`
- Binary: `gk-circle`
- Docker resources: `gk-circle-*`
- Database identifier: `gk_circle`
- Public site: `https://gkcircle.com`

Legacy `jovVix` references are permitted only in historical attribution, upstream links, migration notes, and licence notices.

## Testing and evidence

Run the checks relevant to the change:

```text
API:       go test ./...
Frontend:  npm run lint
Frontend:  npm test -- --run
Build:     npm run build
Compose:   docker compose config --quiet
Runtime:   health checks and workflow smoke tests
```

For UI or end-to-end workflow changes, verify authentication, practice/test creation, joining, answering, scoring, results, and reports in a real running stack. Do not use dummy success responses or disable checks.

Feature completion evidence must exercise the real Nuxt client, Go API, Ory Kratos authentication, and PostgreSQL persistence where those layers are in scope. Mocks and isolated fixtures remain valid for unit tests, but they are not substitutes for real-system completion evidence. Development seed data is permitted only through documented, local-only seed or QA workflows that persist through the normal application data path.

If a required tool is unavailable or an inherited test is failing, report it precisely. Do not claim the task is complete.

## NUC deployment safety

- Project directory: `/home/randhir/apps/gk-circle-v2`
- Deploy with `docker-compose.nuc.yml` and a server-only `.env`.
- Keep the previous GK Circle stack and its volumes intact until cutover is verified.
- Back up the target database before every deployment after it contains data.
- Use `git pull --ff-only`; never force-reset the server checkout.
- Run migrations through the Go `gk-circle migrate up` command.
- Verify API, web, Kratos, database, Redis, and public HTTPS before declaring success.
- Production traffic changes require explicit approval and a tested rollback route.

## Stop conditions

Stop and ask for direction when work requires:

- destructive database or volume operations;
- migration of data from a different schema;
- replacement of the current frameworks or authentication provider;
- removal of upstream attribution or licence material;
- production traffic cutover without a working rollback;
- ambiguous examination requirements that would hardcode the wrong commission or syllabus.

## Completion report

Every completion report must include:

- changes made;
- checks run and their results;
- known failures or unverified areas;
- deployment or rollback notes when applicable;
- `Breaking Changes: YES/NO`;
- database migration status.
