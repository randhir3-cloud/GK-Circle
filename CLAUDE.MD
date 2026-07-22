# GK Circle Project Context

Version: 4.0

Read `AGENTS.md` first. This file provides concise product and implementation context for AI-assisted maintenance.

## Mission

GK Circle helps State PCS aspirants practise, assess, revise, and compete through reliable examination workflows. The platform is evolving from the jovVix real-time quiz foundation; maintainers should adapt its proven quiz engine instead of inventing parallel test systems.

## Product model

The existing quiz entity is the technical foundation for:

- practice sets;
- subject and topic tests;
- current-affairs tests;
- previous-year questions;
- full mock examinations;
- scheduled live competitions.

Product labels may say “practice set”, “mock exam”, or “subject test”, but they should share the existing question, session, answer, scoring, result, and analytics infrastructure.

## Required taxonomy

PCS content must be classifiable by:

- commission and state;
- examination and stage (prelims, mains objective where supported, interview preparation later);
- subject and topic;
- year and source;
- language;
- difficulty;
- test mode and timing.

Do not hardcode one state commission into platform architecture. State-specific content belongs in data and configuration.

## Technology

- `app/`: Nuxt 3, Vue 3, Pinia, Tailwind, Vitest.
- `api/`: Go 1.23, Fiber, SQL migrations, WebSockets.
- PostgreSQL: application and Kratos schemas.
- Ory Kratos: registration, login, recovery, verification, sessions.
- Redis-compatible service: cache and live coordination.
- MinIO: object storage.
- Docker Compose: local and NUC deployment.

NestJS, Next.js, and Prisma are not part of this repository.

## Identity

- Product name: GK Circle
- Public URL: `https://gkcircle.com`
- Repository: `https://github.com/randhir3-cloud/GK-Circle-v2`
- Upstream: `https://github.com/Improwised/jovVix`
- Licence: GNU AGPL-3.0; preserve upstream attribution.

## Current migration priorities

1. Remove legacy jovVix product branding while retaining legal attribution.
2. Align documentation and scripts with the real Go/Nuxt architecture.
3. Make the inherited quiz flows usable as PCS examination workflows.
4. Deploy GK Circle v2 safely to the NUC with an isolated fresh database.
5. Preserve the paused previous production stack as rollback until verification is complete.

## Engineering expectations

- Search before creating files or services.
- Keep authentication and authorization enforced server-side.
- Keep secrets outside Git.
- Use additive SQL migrations with down scripts where feasible.
- Verify builds, tests, Compose configuration, health endpoints, and real workflows.
- Do not hide inherited failures; fix relevant root causes or record them explicitly.
- Do not delete the old NUC stack or its data during the v2 cutover.

## Definition of done

A change is complete only when its acceptance criteria are met and applicable checks pass. Deployment is complete only after the public site, API health, authentication, WebSocket path, database persistence, and rollback route are verified.
