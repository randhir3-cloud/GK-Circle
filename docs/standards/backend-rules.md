# GK Circle Backend Rules

Version: 1.0

Status: Mandatory

---

# Purpose

These rules govern all backend development in GK Circle.

All APIs, services, modules, database interactions, authentication systems, AI integrations, payments, notifications, and background jobs must follow these standards.

Backend consistency is more important than implementation convenience.

---

# Approved Backend Stack

Core Stack:

* NestJS
* TypeScript
* PostgreSQL
* Prisma ORM
* Redis
* BullMQ
* JWT Authentication
* Sentry
* Docker

Do not introduce alternative backend frameworks without approval.

---

# Modular Monolith Rule

Current Architecture:

Modular Monolith

Structure:

Auth

Users

Courses

Tests

Communities

Payments

Notifications

Analytics

AI

Each domain must remain isolated.

Do not mix domain responsibilities.

---

# Service Layer Rule

Controllers must remain thin.

Controllers:

* Validate requests
* Authorize requests
* Call services
* Return responses

Business logic belongs in services.

Never place business logic in controllers.

---

# Repository Rule

Database access must be centralized.

Do not scatter Prisma queries throughout the codebase.

Prefer:

Controller

↓

Service

↓

Repository / Prisma Layer

↓

Database

---

# API First Rule

All business functionality must be accessible through APIs.

Frontend must never contain business logic.

Business rules belong in backend services.

---

# API Standards Rule

Every API must provide:

* Validation
* Authentication
* Authorization
* Error Handling
* Logging
* Documentation
* Testing

Incomplete APIs are prohibited.

---

# Validation Rule

Never trust client input.

All inputs must be validated server-side.

Use:

Zod

or

class-validator

for every request.

Frontend validation is UX.

Backend validation is security.

---

# Authentication Rule

Authentication is complete only when:

* Registration works
* Email verification works
* Login works
* Logout works
* Password reset works
* Session validation works
* Middleware protection works
* Role permissions work

All must be verified.

---

# Authorization Rule

Authentication and Authorization are separate.

Always verify:

Who is the user?

↓

What is the user allowed to do?

Use RBAC wherever appropriate.

---

# JWT Rule

Never trust JWT contents blindly.

Validate:

* Signature
* Expiration
* Claims
* User existence

before authorizing actions.

---

# Email Rule

Email functionality is not complete because SMTP is configured.

Email functionality is complete only when:

Email Sent

↓

Email Delivered

↓

Email Received

↓

Link Works

↓

Workflow Completes

Verification required.

---

# Password Reset Rule

Must verify:

Token Created

↓

Email Delivered

↓

Link Opened

↓

Password Updated

↓

Login Works

---

# Database Rule

PostgreSQL is the source of truth.

Never bypass database validation.

Never trust frontend state over database state.

---

# Prisma Rule

Use:

prisma migrate dev

for schema evolution.

Avoid:

prisma db push

for long-term production workflows.

Commit migration files.

Maintain migration history.

---

# Migration Rule

Forbidden without approval:

DROP TABLE

TRUNCATE

DELETE ALL

prisma migrate reset

Destructive operations require:

Backup

↓

Rollback Plan

↓

Approval

---

# Additive Migration Rule

Prefer:

Add Columns

Add Tables

Add Relationships

Add Indexes

Avoid destructive changes.

---

# Query Rule

Avoid:

N+1 Queries

Repeated Queries

Inefficient Pagination

Over-fetching

Optimize before scaling issues appear.

---

# Index Rule

Add indexes for:

Frequently filtered fields

Frequently searched fields

Foreign keys

Large tables

Performance must be intentional.

---

# Pagination Rule

Large datasets must use pagination.

Never return unlimited records.

---

# Soft Delete Rule

Prefer:

deletedAt

over physical deletion.

Preserve historical data whenever possible.

---

# Event Readiness Rule

Design systems for future event-driven architecture.

Examples:

UserRegistered

EnrollmentCreated

PaymentCompleted

TestSubmitted

CoursePublished

Events should be identifiable even before queues are introduced.

---

# Queue Rule

Long-running tasks must use queues.

Examples:

Email Sending

Notifications

AI Processing

OCR

Report Generation

Background Imports

Do not block request threads.

---

# Redis Rule

Use Redis for:

Caching

Rate Limiting

Queues

Sessions (if applicable)

Avoid using Redis as the primary data store.

---

# Caching Rule

Cache expensive operations.

Do not cache sensitive data without careful consideration.

Always define cache invalidation strategy.

---

# Observability Rule

Every critical workflow must be observable.

Authentication

Payments

Enrollment

AI

Email

OCR

must provide:

* Structured Logging
* Metrics
* Tracing
* Error Monitoring

---

# Logging Rule

Use structured logs.

Avoid:

console.log

for production workflows.

Logs should provide:

Who

What

When

Why

---

# Sentry Rule

All critical failures must be reported.

Capture:

Unhandled Exceptions

API Failures

Queue Failures

Third-party Failures

Background Job Failures

Do not silently swallow errors.

---

# Error Handling Rule

Never expose:

Stack Traces

Database Details

Secrets

Internal Paths

to end users.

Return safe errors.

Log detailed errors internally.

---

# AI Backend Rule

AI requests must support:

Timeouts

Retries

Rate Limits

Monitoring

Fallback Handling

Source Tracking

Do not trust model output blindly.

---

# OCR Rule

OCR workflows must:

Validate files

Limit file size

Track processing status

Handle failures gracefully

Support retries

---

# Payment Rule

Payment success is not proof.

Verify:

Payment Created

↓

Gateway Confirms

↓

Webhook Received

↓

Order Updated

↓

Access Granted

Only then is payment complete.

---

# Notification Rule

Notifications must support:

Retry Logic

Delivery Tracking

Failure Monitoring

Audit Logs

Do not assume delivery.

---

# File Upload Rule

Validate:

Type

Size

Permissions

Storage Policies

Never trust file extensions alone.

---

# Security Rule

Security is mandatory.

Every backend feature must follow:

Authentication

Authorization

Validation

Rate Limiting

Secret Management

Dependency Security

Security Auditing

before completion.

---

# Performance Rule

Avoid:

Blocking Operations

Inefficient Queries

Heavy Computation In Requests

Memory Leaks

Large Payloads

Performance is a feature.

---

# Testing Rule

Every backend feature requires:

Unit Tests

Integration Tests

Workflow Verification

Failure Testing

Edge Case Testing

---

# Completion Rule

Backend work is complete only when:

✓ APIs Work

✓ Validation Exists

✓ Security Verified

✓ Tests Pass

✓ Observability Added

✓ Documentation Updated

✓ Playwright Verification Completed

✓ Evidence Captured

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

## Code-Convention Addenda

The v1.0 rules above are correct but abstract. This addendum pins them to the actual NestJS code in `backend/src/`.

### Module File Structure (Mandatory)

Every feature module must have its own `module.ts`. A module that does not is in violation.

```
feature-name/
├── feature-name.module.ts           # REQUIRED
├── feature-name.controller.ts
├── feature-name.service.ts
├── dto/
└── tests/  (or *.spec.ts co-located)
```

This applies equally to sub-modules. A sub-module under another module (e.g., `Courses/lessons/`) is itself a feature module and must have its own `module.ts`.

**Violations in the current codebase:**

| File | Issue | Required Action |
|---|---|---|
| `backend/src/Courses/lessons/` | No `lessons.module.ts` | Create `lessons.module.ts`, import into `Courses.module.ts` |
| `backend/src/Courses/analytics/` | No `analytics.service.ts`; controller injects `PrismaService` directly | Create `analytics.service.ts`, move DB calls out of controller, create `analytics.module.ts` |
| `backend/src/courses/courses.module.ts:6-11` | Imports `LessonsController`/`LessonsService`/`AnalyticsController` directly into `CoursesModule` providers/controllers | Refactor: each sub-module is registered via its own `module.ts` |

### Controller Anti-Pattern: Direct Prisma Injection

`backend/src/Courses/analytics/analytics.controller.ts:7` injects `PrismaService` directly:

```typescript
@Controller('Courses/:courseId/analytics')
export class AnalyticsController {
  constructor(private readonly prisma: PrismaService) {}  // VIOLATION
}
```

Controllers must not access the database. Move all Prisma calls to `AnalyticsService` and inject the service. This is non-negotiable per the v1.0 "Service Layer Rule" above.

### Sentry Instrumentation Convention (Mandatory)

Per AGENTS.md "Global Sentry Instrumentation" and the v1.0 "Sentry Rule" above, every service method that performs business logic or database access must be wrapped in `Sentry.startSpan` or use one of the canonical wrappers in `backend/src/common/sentry-wrappers.ts`:

- `withSentryApi(handler, { route, method })`
- `withSentryDb(operation, { operation: 'findUnique', table: 'User' })`
- `withSentryAI(task, { provider, model, feature })`
- `withSentryJob(job, { name, queue })`

A "service method" includes controllers' inline service calls. The pattern is:

```typescript
@Injectable()
export class FeatureService {
  async findById(id: string) {
    return Sentry.startSpan(
      { op: 'feature.service', name: 'findById' },
      async () => this.prisma.feature.findUnique({ where: { id } }),
    );
  }
}
```

**Files currently without Sentry instrumentation (violations):**

| File | Action |
|---|---|
| `backend/src/Courses/Courses.service.ts` | Wrap every method in `Sentry.startSpan` |
| `backend/src/Courses/lessons/lessons.service.ts` | Wrap every method in `Sentry.startSpan` |
| `backend/src/uploads/uploads.controller.ts` | Wrap each `@Get`/`@Post` handler in `Sentry.startSpan` |
| `backend/src/auth/mail.service.ts` | Verify Sentry usage; add if missing |

`backend/src/testing/services/test-gateway.service.ts` is the model implementation. `backend/src/circles/circles.service.ts` is also compliant.

### DTO Type Safety (Mandatory)

`backend/src/Courses/lessons/lessons.controller.ts:11` accepts `createLessonDto: any`. This is a violation per the v1.0 "Validation Rule" and TypeScript strict mode.

Every controller method must accept a typed DTO class:

```typescript
export class CreateLessonDto {
  @IsString() @IsNotEmpty() @MaxLength(200) title: string;
  @IsString() @IsNotEmpty() moduleId: string;
  @IsEnum(LessonType) contentType: LessonType;
  // ...
}
```

### Sub-Module Extraction (Sprint 1 Backlog)

The Sprint 1 plan includes the following backend refactors, which this standard now requires:

1. **Create `backend/src/Courses/lessons/lessons.module.ts`** with `controllers: [LessonsController]`, `providers: [LessonsService]`, `exports: [LessonsService]`.
2. **Create `backend/src/Courses/analytics/analytics.service.ts`** with all current Prisma logic from the controller. Create `analytics.module.ts`. Inject the service in the controller.
3. **Update `Courses.module.ts`** to import `LessonsModule` and `AnalyticsModule` instead of declaring their controllers/providers directly.
4. **Add Sentry spans** to all methods in the two extracted services.

### Anti-Patterns Found (Recap)

- `lessons.service.ts:46` sets `status: 'published'` as a string literal; should be the `LessonVersionStatus` enum value.
- `analytics.controller.ts` injects `PrismaService` directly.
- `Courses.service.ts:101-109` lacks a duplicate-enrollment check.
- `Courses.service.ts:65-85` returns `progress` for all users, not just the requesting user.

## Compatibility With Other Standards

This document defers to:

- [security-rules.md](docs/standards/security-rules.md) for all security rules (auth, authz, rate limiting, file security)
- [architecture-rules.md](docs/standards/architecture-rules.md) for module boundaries and service extraction rules
- [Course-rules.md](docs/standards/Course-rules.md) for Course-domain rules (Course, Module, Lesson schemas, EnrollmentGuard location)
- [live-exam-rules.md](docs/standards/live-exam-rules.md) for live test engine rules (WebSocket gateway, Redis state)
- [rag-rules.md](docs/standards/rag-rules.md) for AI/RAG service rules

---

# Final Directive

Backend systems must be:

Reliable

Observable

Secure

Scalable

Maintainable

Production Ready

Build for millions of users, not dozens.
