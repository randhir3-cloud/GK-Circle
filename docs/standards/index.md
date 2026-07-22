# GK Circle Standards Index

Version: 3.0

These standards govern the current Go/Fiber and Nuxt/Vue repository.

## Required reading

Read in this order:

1. `AGENTS.md`
2. `CLAUDE.MD`
3. `docs/standards/architecture-rules.md`
4. `docs/standards/security-rules.md`
5. The task-specific standards below

## Maintained standards

| Task | Standard |
|---|---|
| Go API, SQL, migrations, WebSockets | `backend-rules.md` |
| Nuxt/Vue UI | `frontend-rules.md` |
| Tests and evidence | `testing-rules.md` |
| Documentation | `documentation-rules.md` and `documentation-governance.md` |
| Docker, NUC, backup, release | `operations-rules.md` and `devops-rules.md` |
| PCS product behavior | `course-rules.md` and `live-exam-rules.md` |

## Legacy planning documents

Several files in this directory were imported from an earlier, different GK Circle architecture. Any statement that assumes NestJS, Next.js, Prisma, a `backend/` directory, or a `frontend/` directory is legacy planning material and is not applicable to this repository. It must not override `AGENTS.md`, `CLAUDE.MD`, or the maintained standards above.

Creator economy, RAG, mobile, and advanced admin documents describe possible future capabilities. They do not authorize implementation or architecture changes.

## Required workflow

1. Inspect existing code and search for reusable implementation.
2. Define acceptance criteria and risks.
3. Obtain approval for architecture, database, security, deployment, or destructive work.
4. Implement the smallest compatible change.
5. Run applicable checks and collect evidence.
6. Update affected documentation.
7. Report breaking changes and migration status.

Never claim success from code existence alone.
