# GK Circle Architecture Rules

Version: 1.1

Status: Mandatory

**v1.1 (2026-07-12):** Repository Reuse Rule, Parallel System Prohibition, Mandatory Repository Search Rule. Supplements v1.0; AGENTS.md wins on conflict.

---

# Purpose

These rules govern all architecture decisions across GK Circle.

All implementations must comply with these standards.

Architecture consistency is more important than implementation convenience.

---

# CRITICAL ARCHITECTURE CHANGE RULE

Core architecture decisions require approval.

Never replace:

* Frontend Framework
* Backend Framework
* Database
* Authentication System
* Search Engine
* Payment Provider
* AI Infrastructure
* Queue System
* Storage System

without explicit approval.

Prefer extending architecture over replacing architecture.

---

# CRITICAL SCOPE CONTROL RULE

Implement only what is requested.

Do not:

* Rewrite unrelated systems
* Refactor stable modules
* Replace working components
* Change architecture during bug fixes

If improvements are identified:

1. Document them.
2. Create implementation plan.
3. Obtain approval.

Stay within approved scope.

---

# Documentation First Rule

Before implementation create:

* Architecture
* Specification
* Data Model
* API Design
* Testing Plan

Code comes after architecture.

---

# Single Source Of Truth Rule

Never create duplicate systems.

Examples:

Bad:

Auth System A

Auth System B

---

Bad:

Course Service A

Course Service B

---

Good:

One System

Extended Through Features

---

# Course Architecture Rule

Courses are the root educational entity.

All learning content must belong to a Course.

Structure:

Course

↓

Module

↓

Lesson

↓

Assessment

↓

Analytics

---

Supported Course Types

* Course
* Test Series
* Current Affairs
* Mentorship
* Live Batch
* Interview Program
* Crash Course

---

# Test Engine Rule

Maintain a single test engine.

Do NOT create:

* Course Tests
* Mock Tests Engine
* Quiz Engine
* Practice Engine

as separate systems.

Use:

Test

↓

Question

↓

Submission

↓

Results

with configuration.

---

# Community Integration Rule

Courses integrate with Communities.

Structure:

Course

↓

Enrollment

↓

Community Access

↓

Discussion

↓

Mentorship

↓

Live Sessions

Never create isolated communities disconnected from Courses.

---

# Domain Ownership Rule

Every domain has a clear owner.

Examples:

Auth Domain

User Domain

Course Domain

Test Domain

Community Domain

AI Domain

Payment Domain

Notification Domain

Analytics Domain

Do not mix responsibilities.

---

# API-First Rule

All business logic must be exposed through APIs.

Frontend must never contain business logic.

Business logic belongs in backend services.

---

# Reusable Architecture Rule

Build reusable systems.

Avoid:

* Feature-specific code
* Hardcoded workflows
* One-off implementations

Prefer:

* Shared services
* Shared components
* Shared modules

---

# Database Evolution Rule

Database evolves.

Database is not rebuilt.

Use:

Additive Migrations

Never destructive migrations.

---

# CRITICAL MIGRATION RULE

Never execute:

DROP TABLE

TRUNCATE

DELETE ALL

prisma migrate reset

without approval.

Always prefer:

prisma migrate dev

with additive migrations.

---

# Backward Compatibility Rule

Existing functionality must continue working.

When changing:

* APIs
* Database Schemas
* Services

maintain compatibility whenever possible.

---

# Event-Driven Readiness Rule

Design domains so future event-driven architecture is possible.

Examples:

Enrollment Created

Payment Completed

Course Published

User Registered

Test Submitted

Events should be identifiable even if message queues are added later.

---

# Feature Flag Rule

Experimental functionality must be behind feature flags.

Never expose incomplete functionality to production users.

---

# Scalability Rule

Build systems assuming:

* Millions of users
* Millions of questions
* Millions of tests
* Millions of enrollments

Avoid architecture that only works at small scale.

---

# Monolith First Rule

Current architecture:

Modular Monolith

Do not prematurely create microservices.

Microservices are introduced only when justified.

---

# Microservices Roadmap

Current:

Modular Monolith

Future:

Auth Service

Course Service

Test Service

Community Service

AI Service

Notification Service

Analytics Service

Payment Service

Architecture should remain compatible with future extraction.

---

# Active Context Rule

Before implementation read:

docs/active-context/latest-session.md

docs/active-context/active-plan.md

docs/active-context/active-tasks.md

Maintain continuity.

Do not restart solved work.

---

# No Duplicate Directory Rule

Before creating:

* Components
* Services
* Hooks
* Utilities
* Modules

search existing project first.

Reuse before creating new files.

---

# Repository Reuse Rule

Before creating any of the following, **search the repository** and reuse existing implementations whenever possible. Parallel implementations are prohibited without explicit approval.

| Artifact | Search locations (examples) |
|----------|----------------------------|
| Component | `frontend/src/components/`, feature folders |
| DTO | `backend/src/**/dto/` |
| Hook | `frontend/src/**/hooks/` |
| Store | `frontend/src/**/stores/` |
| Service | `backend/src/**/*.service.ts` |
| Controller | `backend/src/**/*.controller.ts` |
| Utility | `backend/src/common/`, `frontend/src/lib/` |
| API route / module | Existing Nest modules, Next.js app routes |
| Migration | `backend/prisma/migrations/` — extend schema, do not fork |

If similar functionality exists, extend it. Document why extension is insufficient before proposing a new parallel system.

---

# Mandatory Repository Search Rule

**Mandatory for all agents:** Before creating any of the following, search the repository and confirm no suitable implementation exists:

- DTO
- Component
- Service
- Hook
- Store
- Migration / schema change
- Utility
- API module or route

Use `Glob`, `Grep`, or codebase search across `backend/src/`, `frontend/src/`, and `backend/prisma/`. Document search performed in the implementation plan. Duplicate parallel implementations are **prohibited** without explicit approval and an ADR.

This rule extends §No Duplicate Directory Rule and §Repository Reuse Rule — all three apply together.

---

# Parallel System Prohibition Rule

Never create a second implementation of a core platform capability without explicit approval.

Forbidden without approval:

- Second auth system
- Second uploader / file pipeline
- Second notification system
- Second question engine
- Second test engine
- Second QA account framework

Extend the existing system. See also §Single Source Of Truth Rule and §Test Engine Rule.

---

# Dependency Rule

Do not introduce new dependencies unless:

1. Existing solution unavailable.
2. Dependency actively maintained.
3. Security reviewed.
4. Approval received.

Prefer fewer dependencies.

---

# Architecture Review Rule

Major architecture changes require:

Problem

↓

Options

↓

Tradeoffs

↓

Recommendation

↓

Approval

↓

Implementation

---

# v0.5 Addendum (Appended 2026-06-05)

This addendum supplements the v1.0 rules above. It does not replace them. Where this addendum and the v1.0 rules conflict, the more specific rule wins. Where this addendum and AGENTS.md conflict, AGENTS.md wins.

## Dependency Rule (Mandatory)

This document does not replace AGENTS.md.

It supplements AGENTS.md.

If this document conflicts with AGENTS.md, AGENTS.md wins.

If this document conflicts with another standards file, the more specific standard wins.

If ambiguity exists, document the ambiguity in an ADR before implementation.

## Verification Requirement (Mandatory)

A rule is not considered satisfied because:

- Code exists
- Configuration exists
- Tests exist
- Documentation exists

A rule is satisfied only when evidence exists.

Evidence may include:

- Automated tests
- Database verification
- API verification
- Security verification
- Playwright verification
- Production telemetry

## Module Boundary Code-Convention

The v1.0 "Domain Ownership Rule" and "Modular Monolith Rule" require module isolation. This addendum pins the rule to the file system.

### Module Path Convention

Every backend module must live at `backend/src/<module-name>/` and have the following minimum files:

```
backend/src/<module-name>/
├── <module-name>.module.ts
├── <module-name>.controller.ts
├── <module-name>.service.ts
└── dto/
```

Sub-modules (e.g., `Courses/lessons/`) follow the same pattern. The `lessons/` sub-module must have its own `lessons.module.ts` and be imported by `Courses.module.ts`.

### Module Communication

| Direction | Allowed Mechanism |
|---|---|
| A → B (request/response) | Inject B's exported service |
| A → B (fire-and-forget) | Emit domain event; B subscribes |
| A → B (database) | FORBIDDEN. A must call B's service. |

### Current State of the Codebase

| Module | Path | State |
|---|---|---|
| auth | `backend/src/auth/` | ✅ Complete |
| users | `backend/src/users/` | ✅ Complete |
| circles | `backend/src/circles/` | ✅ Complete (note: name is `circles`, domain is `community`) |
| testing | `backend/src/testing/` | ✅ Has controller + service; live-test gateway planned but not implemented |
| admin | `backend/src/admin/` | ✅ Has questions sub-module |
| Courses | `backend/src/Courses/` | ⚠️ Has sub-modules but `lessons/` and `analytics/` lack `module.ts` |
| uploads | `backend/src/uploads/` | ⚠️ Has controller; lacks `service.ts` (controller directly uses `StorageService`) |
| common | `backend/src/common/` | ✅ Shared utilities |
| mail | `backend/src/mail/` | ⚠️ Has provider; lacks `service.ts` |
| **ai** | (missing) | ❌ Module does not exist |
| **analytics** | (mixed into Courses) | ❌ Not its own top-level module |
| **creator** | (missing) | ❌ Module does not exist |
| **notifications** | (missing) | ❌ Module does not exist |
| **payments** | (missing) | ❌ Module does not exist |
| **search** | (missing) | ❌ Module does not exist |
| **learning** | (mixed into Courses) | ❌ Should be top-level, not `Courses/lessons/` |

### Mandatory Top-Level Modules (per architecture.md § "Backend Modules")

The following top-level modules must exist in `backend/src/`. Items marked **MUST CREATE** do not currently exist:

- `auth/` ✅
- `users/` ✅
- `learning/` **MUST CREATE** (extract from `Courses/lessons/`)
- `testing/` ✅
- `community/` **MUST CREATE or document** (current is `circles/`, which is fine; the architecture.md §"Backend Modules" calls it `community`)
- `ai/` **MUST CREATE**
- `analytics/` **MUST CREATE** (extract from `Courses/analytics/`)
- `creator/` **MUST CREATE**
- `notifications/` **MUST CREATE**
- `payments/` **MUST CREATE**
- `search/` **MUST CREATE**
- `admin/` ✅
- `common/` ✅
- `uploads/` ✅
- `Courses/` ✅ (becomes a coordination module for the catalog)

### Module Extraction Order (Sprint 1 → Sprint 3)

1. **Sprint 1**: Create `lessons.module.ts` and `analytics.module.ts` (sub-module boundaries within `Courses/`)
2. **Sprint 1**: Create `EnrollmentGuard` (cross-cutting, lives in `common/guards/`)
3. **Sprint 2**: Extract `learning/` from `Courses/lessons/`
4. **Sprint 2**: Extract `analytics/` from `Courses/analytics/`
5. **Sprint 2**: Create `ai/` module (RAG pipeline + AI tutor endpoints)
6. **Sprint 2**: Create `notifications/` module
7. **Sprint 3**: Create `payments/`, `creator/`, `search/`

### Frontend Folder Structure Compliance

The v1.0 "Feature Folder Rule" is correctly implemented in the frontend at `frontend/src/features/`. The audit found no current violations.

However, the v1.0 "Page Composition Rule" is violated in `frontend/src/app/(student)/learn/[slug]/page.tsx:16` — a page component directly calls `getCourseBySlug` from inside a `useEffect`. This must be refactored to use a TanStack Query hook from `features/learning/hooks/`.

## Compatibility With Other Standards

This document defers to:

- [backend-rules.md](docs/standards/backend-rules.md) for NestJS service and module patterns
- [security-rules.md](docs/standards/security-rules.md) for security rules
- [Course-rules.md](docs/standards/Course-rules.md) for Course catalog rules
- [live-exam-rules.md](docs/standards/live-exam-rules.md) for live test architecture
- [rag-rules.md](docs/standards/rag-rules.md) for RAG pipeline architecture

---

# ADR Requirement Rule (Mandatory)

This rule prevents architecture drift by requiring a formal Architecture Decision Record (ADR) before any of the following changes are implemented.

## Changes That Require an ADR

Any of the following MUST be documented as an ADR in `docs/adrs/NNNN-title.md` **before** implementation begins:

1. **New top-level module** in `backend/src/` or `frontend/src/features/`
2. **New database root entity** (a new model that is not a join table, sub-table, or extension of an existing model)
3. **New external service integration** (payment provider, email provider, search engine, vector DB, etc.)
4. **New AI provider** (OpenAI, Anthropic, Ollama, OpenRouter, self-hosted)
5. **New payment provider** (Razorpay, Stripe, PayPal, etc.)
6. **New storage provider** (S3, R2, MinIO, local filesystem abstraction)
7. **New authentication provider** (Google, Apple, GitHub, SSO/SAML, magic link)
8. **New notification channel** (push, SMS, in-app)
9. **New search engine or search architecture change**
10. **Breaking API changes** (removing fields, changing types, renaming endpoints, changing auth)
11. **Cross-domain architecture changes** (a change that touches ≥ 2 top-level modules)
12. **Database topology changes** (read replicas, sharding, partitioning, server splits)
13. **Deployment topology changes** (new regions, new CDN, new load balancer)
14. **Architectural pattern introduction** (event sourcing, CQRS, saga, outbox)
15. **Frontend framework, state management, or routing change**
16. **Removing a dependency, framework, or major library**

## Changes That Do NOT Require an ADR

The following are routine and do NOT require an ADR:

- Adding a new endpoint within an existing module
- Adding new fields to an existing model (additive migration)
- Adding new DTOs
- Adding new tests
- Adding new UI components within an existing feature
- Bug fixes that do not change behavior
- Performance optimizations that preserve external behavior
- Documentation updates

When in doubt, write the ADR. A 10-minute ADR is cheaper than a 10-day revert.

## ADR Format (Mandatory)

ADRs live in `docs/adrs/` and follow this format (from [documentation-rules.md](docs/standards/documentation-rules.md)):

```markdown
# NNNN. Title

**Status:** Proposed | Accepted | Superseded by NNNN
**Date:** YYYY-MM-DD
**Author:** <name or agent>

## Context

What is the situation? What forces are at play?

## Options Considered

What alternatives were evaluated?

## Decision

What was decided?

## Consequences

What becomes easier? What becomes harder? What new risks are introduced?

## Compliance

How will this decision be enforced? What tests, linters, or processes verify it?
```

ADRs are immutable once accepted. To change a decision, write a new ADR that supersedes the old one (do not edit the old ADR).

## ADR Authority

| Decision Type | Author | Approver |
|---|---|---|
| New top-level module | Module owner | Tech lead |
| New external service | Module owner | Architecture team |
| New AI/payment/storage/auth provider | Module owner | Tech lead + security review |
| Breaking API change | Affected module owners | Tech lead |
| Database topology | Backend lead | Architecture team |
| Frontend framework change | Frontend lead | Architecture team |

---

# Final Directive

Architecture is a long-term asset.

Short-term convenience must never damage long-term maintainability.

Protect architecture integrity at all times.
