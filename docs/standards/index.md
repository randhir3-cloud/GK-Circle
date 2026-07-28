# GK Circle Standards Index

Version: 3.2

These standards govern the current Go/Fiber and Nuxt/Vue repository.

## Required reading

Read in this order:

1. `AGENTS.md`
2. `CLAUDE.MD`
3. `docs/standards/index.md`
4. `docs/standards/architecture-rules.md`
5. `docs/standards/security-rules.md`
6. `docs/standards/ai-rules.md`
7. `docs/standards/testing-rules.md`
8. The task-specific standards below

Before changing or verifying the repository, report:

- the mandatory standards loaded;
- the standards specific to the task;
- any conflict, ambiguity, or stop condition.

No implementation, verification decision, blocker classification, or completion claim may occur before this declaration.

## Authority order

When instructions conflict, use this order:

1. frozen acceptance criteria and approved ADRs;
2. `AGENTS.md`;
3. `CLAUDE.MD`;
4. maintained standards;
5. the current task prompt;
6. future or legacy planning material.

Do not rewrite frozen history to resolve a conflict. Stop when the higher authority does not provide enough direction.

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

## Verification ownership

Authorized agents may perform task-level runtime verification, including browser automation, against the real local system. The manual-operator requirement in `module-certification-standard.md` remains mandatory only for formal module certification. Task-level automation does not itself grant formal module certification or replace any explicitly required human sign-off.
