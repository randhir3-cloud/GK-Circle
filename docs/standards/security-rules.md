# GK Circle Security Rules

Version: 1.1

Status: Mandatory

**v1.1 (2026-07-12):** QA Credential Rule (`backend/.env` single source). Supplements v1.0; AGENTS.md wins on conflict.

---

# Purpose

These rules govern all security-related decisions, implementations, integrations, infrastructure, APIs, authentication systems, payments, AI systems, databases, and user data across GK Circle.

Security is mandatory.

Security is not optional.

Security is not a later phase.

Security begins during architecture.

---

# Security First Rule

Every feature must be designed with security in mind.

Security is a core requirement.

Not an enhancement.

Not a future task.

Not a post-launch task.

---

# OWASP Rule

All implementations must comply with:

OWASP Top 10

OWASP API Security Top 10

OWASP Authentication Guidelines

OWASP Session Management Guidelines

---

# Secure By Default Rule

The default behavior of any system must be secure.

Examples:

Private before public.

Denied before allowed.

Validated before processed.

Authenticated before trusted.

---

# Authentication Rule

Authentication is not complete until:

User Registered

↓

Email Verified

↓

Login Successful

↓

Session Created

↓

Protected Routes Enforced

↓

Logout Verified

All steps must work.

---

# Authorization Rule

Authentication and authorization are separate.

Always verify:

Who is the user?

↓

What is the user allowed to do?

Role-based access control is required where appropriate.

---

# Role-Based Access Control Rule

Roles may include:

Super Admin

Admin

Moderator

Mentor

Creator

Student

Guest

Permissions must be enforced server-side.

Never trust frontend roles.

---

# Session Security Rule

Sessions must:

Expire correctly

Be validated

Be revocable

Be protected from theft

Do not store sensitive authentication state in localStorage.

Prefer secure cookies.

---

# JWT Security Rule

JWTs must be:

Signed

Validated

Expired

Rotated when necessary

Never trust claims without validation.

---

# Secret Management Rule

Secrets must never be hardcoded.

Use:

.env

Secret Manager

Vault

Environment Variables

Never commit secrets to source control.

---

# QA Credential Rule

QA account passwords are secrets. They must never be hardcoded or duplicated across files.

**Single source of truth:** `backend/.env`

Required variables:

- `QA_STUDENT_PASSWORD`
- `QA_CREATOR_PASSWORD`
- `QA_ADMIN_PASSWORD`
- `QA_SUPERADMIN_PASSWORD`

Rules:

- Frontend applications and `frontend/.env.local` must **not** store QA passwords
- Playwright and QA scripts load credentials via `backend/scripts/qa/load-qa-env.ts` or `frontend/playwright/helpers/qa-governance.ts` → `loadQaEnvFromBackend()`
- Tests must use `getQaUser()` / `loginQaViaApi()` — never embed password strings in specs

Full QA identity rules: **`docs/testing/qa-account-governance.md`**

# Environment Variable Rule

Never expose:

Database Secrets

Service Role Keys

SMTP Credentials

Payment Secrets

AI Provider Secrets

Private Tokens

to frontend applications.

---

# Database Security Rule

Database is the source of truth.

Protect:

User Data

Payments

Enrollments

AI Data

Community Data

Analytics

All access must be controlled.

---

# Row Level Security Rule

When using client-accessible databases:

Enable RLS.

Create policies.

Verify policies.

RLS without policies is incomplete.

---

# Service Role Rule

Service role credentials must never be exposed to clients.

Use service roles only on trusted backend systems.

---

# Input Validation Rule

Never trust user input.

Validate:

Type

Format

Length

Permissions

Ownership

Server-side validation is mandatory.

---

# Output Encoding Rule

User-generated content must be safely rendered.

Prevent:

XSS

HTML Injection

Script Injection

Unsafe rendering

---

# API Security Rule

Every API must include:

Validation

Authentication

Authorization

Error Handling

Rate Limiting

Logging

Testing

---

# Rate Limiting Rule

Protect:

Login

Registration

Password Reset

AI Endpoints

Payments

Email Endpoints

Expensive Operations

from abuse.

---

# CORS Rule

Restrict API access to approved origins.

Avoid:

Access-Control-Allow-Origin: *

on sensitive endpoints.

---

# File Upload Security Rule

Validate:

File Type

MIME Type

File Size

Permissions

Storage Policies

Never trust file extensions.

---

# Malware Prevention Rule

Uploaded files must never be executable.

Store uploads outside executable paths.

Use safe storage policies.

---

# Email Security Rule

Protect:

Verification Emails

Password Reset Emails

Invitations

Notifications

Never expose sensitive tokens unnecessarily.

Tokens must expire.

---

# Password Reset Security Rule

Reset tokens must be:

Unique

Secure

Time Limited

Single Use

---

# Payment Security Rule

Never grant access based solely on frontend success.

Verify:

Payment Provider

↓

Webhook

↓

Order Update

↓

Access Grant

Server-side verification required.

---

# Webhook Security Rule

All webhooks must verify signatures.

Never trust incoming webhook requests.

Validate origin and signature.

---

# AI Security Rule

AI systems must:

Validate Inputs

Rate Limit Requests

Monitor Usage

Track Costs

Prevent Prompt Abuse

Prevent Data Leakage

Monitor Failures

---

# RAG Security Rule

RAG systems must:

Track Sources

Validate Sources

Prevent Source Fabrication

Prevent Unauthorized Data Retrieval

Never expose restricted content.

---

# Dependency Security Rule

All dependencies must:

Be maintained

Be reviewed

Be actively supported

Avoid abandoned packages.

---

# Package Audit Rule

Regularly run:

npm audit

pnpm audit

dependency scanning

before releases.

---

# Logging Rule

Logs must not expose:

Passwords

Tokens

Secrets

Private Keys

Sensitive User Data

---

# Monitoring Rule

Monitor:

Authentication Failures

Payment Failures

Email Failures

AI Failures

Database Failures

Queue Failures

Security Incidents

---

# Sentry Rule

Capture:

Exceptions

API Failures

Queue Failures

Background Job Failures

Third-party Failures

Security Events

---

# Data Privacy Rule

Collect only necessary data.

Minimize retention.

Protect user privacy.

Follow applicable privacy regulations.

---

# Backup Rule

Critical data requires backups.

Backups must be:

Automated

Verified

Restorable

Test recovery procedures.

---

# Disaster Recovery Rule

Prepare for:

Database Failure

Server Failure

Storage Failure

Provider Failure

Recovery plans must exist.

---

# Security Audit Rule

Before production releases:

Run security audit.

Review:

Authentication

Authorization

Database

Payments

Emails

AI

Dependencies

Infrastructure

---

# Security Checklist Rule

Verify:

✓ Secrets Protected

✓ Validation Exists

✓ Authentication Works

✓ Authorization Works

✓ RLS Configured

✓ Rate Limiting Enabled

✓ File Uploads Secured

✓ Payments Verified

✓ Dependencies Audited

✓ Monitoring Enabled

before completion.

---

# Security Incident Rule

If a vulnerability is discovered:

Stop

↓

Assess

↓

Contain

↓

Fix

↓

Verify

↓

Document

Never ignore security issues.

---

# Production Readiness Rule

A feature cannot be production-ready unless security verification is complete.

Security is part of Definition of Done.

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

## Code-Convention Addenda (Concrete Patterns GK Circle Must Use)

The v1.0 rules above are correct but abstract. This addendum pins them to the actual codebase.

### JWT Lifetime Convention

| Token | Lifetime | Storage | Rotation |
|---|---|---|---|
| Access | 15 minutes | Frontend memory (Zustand) | Re-issue on every refresh |
| Refresh | 7 days | httpOnly, Secure, SameSite=Strict cookie | Rotate on every refresh; family-revoke on reuse |

`backend/src/auth/auth.controller.ts:66-71` and `:99-104` already implement this. Verified compliant.

### Account Lockout Convention (corrected 2026-06-05)

After 10 failed login attempts within a rolling 15-minute window:

- Increment `User.failedLoginAttempts`
- Set `User.lockoutUntil = now + 15 minutes`
- Return `401 Unauthorized` with a generic "Invalid credentials" message (no enumeration)
- `auth.service.ts:login` must check `lockoutUntil` before `bcrypt.compare`

**Correction history**: The original v0.5 addendum (also appended 2026-06-05) stated "5" attempts. The code at `backend/src/auth/auth.service.ts:70` implements 10 (`if (newFailedAttempts >= 10)`). Per the decision recorded in [../active-context/decisions.md](../active-context/decisions.md) on 2026-06-05, **the code is the source of truth and documentation reflects reality**. The threshold is 10, not 5.

The `20260603123012_add_lockout_fields` migration adds the columns. Runtime enforcement is in `auth.service.ts:login` (lockout check at line 56–63, threshold at line 70).

### Rate Limiter Convention (Redis-Backed)

`backend/src/auth/guards/rate-limit.guard.ts` currently uses an in-process `Map`. This is a multi-instance anti-pattern.

Replace with `@nestjs/throttler` configured with a Redis store:

```typescript
ThrottlerModule.forRoot({
  throttlers: [
    { name: 'login', limit: 5, ttl: 60_000 },
    { name: 'register', limit: 3, ttl: 3_600_000 },
  ],
  storage: new ThrottlerStorageRedisService(redisClient),
}),
```

The current `if (process.env.NODE_ENV !== 'production') return true;` bypass must be removed. Rate limits must run in development so Playwright can verify the limit behavior.

### EnrollmentGuard Convention

A canonical `EnrollmentGuard` must be created (location: `backend/src/Courses/guards/enrollment.guard.ts` or `backend/src/auth/guards/enrollment.guard.ts` per architectural decision). It must:

1. Resolve the resource (lesson, file, or test) to a Course
2. Verify the requesting user has an `Enrollment` with `status: 'ACTIVE'` for that Course
3. Throw `ForbiddenException` (403) if not enrolled
4. Be wrapped in `withSentryApi` for observability
5. Return the same response (404) for missing resources as for forbidden resources in some scenarios — but never both 404 and 403 for the same request

It must be applied to:

- `GET /api/v1/lessons/:id`
- `GET /api/v1/uploads/videos/:filename`
- `GET /api/v1/uploads/pdfs/:filename`
- `GET /api/v1/tests/:id/attempts`

### CORS Startup Validation Convention

`backend/src/main.ts:27-30` accepts `origin: process.env.FRONTEND_URL || 'http://localhost:3000'`. The fallback is a production-day-1 outage risk.

`env.validation.ts` must require `FRONTEND_URL` in production. If absent, the app must fail-fast at startup with a clear error.

### Sentry Sensitive-Data Scrubbing Convention

`main.ts:13-18` initializes Sentry without a `beforeSend` scrubber. Add:

```typescript
Sentry.init({
  dsn: process.env.SENTRY_DSN,
  beforeSend(event) {
    if (event.request) {
      delete event.request.cookies;
      if (event.request.headers) {
        delete event.request.headers.authorization;
        delete event.request.headers.cookie;
      }
    }
    if (event.extra) {
      delete event.extra.password;
      delete event.extra.refreshToken;
      delete event.extra.accessToken;
    }
    return event;
  },
  // ... other config
});
```

### Security Regression Test Convention

Create `frontend/playwright/security-regression.spec.ts` (in addition to the existing auth specs) that verifies:

1. Logout invalidates session (refresh after logout returns 401)
2. Account lockout after 5 failed logins
3. Cross-user resource access returns 403
4. Unauthenticated request to protected endpoint returns 401
5. File download requires enrollment
6. Admin endpoint requires admin role
7. `getCourseBySlug` does not leak `videoUrl`/`pdfUrl`/`articleBody`/`content` to unenrolled users
8. `getCourseBySlug` does not leak other users' `LessonProgress` to the requesting user

## Files Currently Out of Security Compliance (as of 2026-06-05)

| File | Issue | Required Action |
|---|---|---|
| `backend/src/auth/guards/rate-limit.guard.ts` | In-memory rate limiter; dev bypass | Migrate to `@nestjs/throttler` + Redis; remove bypass |
| `backend/src/uploads/uploads.controller.ts:60-80` | Only `JwtAuthGuard`; no enrollment check for `videos`/`pdfs` | Add `EnrollmentGuard` for `videos` and `pdfs` directories |
| `backend/src/Courses/Courses.service.ts:65-85` | `getCourseBySlug` leaks all users' `progress` rows | Filter `progress` by `req.user.id` or omit for non-enrolled |
| `backend/src/Courses/Courses.service.ts:101-109` | `createEnrollment` has no duplicate guard | Add `findFirst` check before create |
| `backend/src/main.ts:13-18` | Sentry init lacks `beforeSend` scrubbing | Add scrubber per convention above |
| `backend/src/main.ts:27-30` | CORS fallback to localhost in production | Fail-fast in production if `FRONTEND_URL` missing |
| (missing) `backend/src/Courses/guards/enrollment.guard.ts` | `EnrollmentGuard` not implemented | Create per convention above |

These items are the basis for the Sprint 1 security fix set in the audit report.

---

## Sprint 1 Phase A: Verified Security Boundaries (2026-06-06, SPG-5)

The following boundaries were verified by backend e2e regression (102 tests). Evidence: `docs/active-context/spg-5-evidence.md`, audit: `docs/active-context/idor-audit-sprint-1-phase-a.md`.

| Boundary | Verified | Test suite |
|----------|----------|------------|
| Lesson access requires enrollment | 2026-06-06 | `lessons-access`, `cross-account-access` |
| Course slug does not leak other users' progress | 2026-06-06 | `Courses-slug-privacy`, `cross-account-access` |
| File download requires enrollment/ownership | 2026-06-06 | `uploads-authorization`, `cross-account-access` |
| Cross-account isolation (User B ≠ User A resources) | 2026-06-06 | `cross-account-access` |
| Progress ownership (`req.user.id` only) | 2026-06-06 | `progress-security` |
| Enrollment scoped per Course | 2026-06-06 | `enrollment-ownership-scope`, `modules-authorization` |

**Deferred**: Frontend Playwright mirror (`security-phase-a.spec.ts`) — Sprint 2 Phase D.

---

# Final Directive

Security is a feature.

Security is a requirement.

Security is a release blocker.

If security is not verified:

The feature is not complete.
