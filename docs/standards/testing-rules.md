# GK Circle Testing Rules

Version: 1.1

Status: Mandatory

**v1.1 (2026-07-12):** QA account governance, Playwright preconditions, repository cleanliness, Production Verification Rule, strengthened Definition of Done. Supplements v1.0; AGENTS.md wins on conflict.

---

# Purpose

These rules govern all testing, verification, validation, QA, Playwright execution, evidence collection, and feature completion requirements across GK Circle.

Code quality is mandatory.

Testing is not optional.

Verification is not optional.

Evidence is required.

---

# Core Principle

Implementation is not proof.

Configuration is not verification.

Success messages are not confirmation.

A feature is complete only when evidence proves it works.

---

# Testing Pyramid Rule

Every feature should be tested using:

Unit Tests

↓

Integration Tests

↓

End-to-End Tests

↓

User Workflow Verification

Avoid relying solely on one testing layer.

---

# Required Testing Levels

Every feature must be evaluated against:

1. Unit Tests
2. Integration Tests
3. API Tests
4. End-to-End Tests
5. Workflow Validation
6. Security Validation
7. Regression Testing

---

# Unit Testing Rule

Unit tests must verify:

* Business Logic
* Validation Logic
* Utility Functions
* Edge Cases
* Error Conditions

Unit tests do NOT prove the feature works end-to-end.

---

# Integration Testing Rule

Integration tests must verify:

* Service Interactions
* Database Interactions
* API Flows
* Event Flows
* External Service Integrations

Integration tests verify systems working together.

---

# End-to-End Rule

Critical workflows require E2E testing.

Examples:

Authentication

Enrollment

Payments

Course Purchase

Email Verification

Password Reset

AI Tutor

Live Tests

Current Affairs

Community Participation

---

# Playwright Rule

Playwright is the primary E2E verification framework.

Every completed feature must be validated using real Playwright execution whenever possible.

Playwright must use:

Real APIs

Real Database

Real Authentication

Real Workflows

Mocked workflows are not valid completion evidence.

---

# QA Account Governance Rule

GK Circle maintains exactly **four** immutable QA identities. Full specification: **`docs/testing/qa-account-governance.md`**.

Agents and tests MUST:

- Use only `qa.student@gkcircle.com`, `qa.creator@gkcircle.com`, `qa.admin@gkcircle.com`, `qa.superadmin@gkcircle.com`
- Obtain credentials via `getQaUser()`, `loginAs()`, or `loginQaViaApi()` only
- Run `verify-qa-accounts.ts` to sync accounts before Playwright (local/RCF global setup does this automatically)
- Run `cleanup-ephemeral-users.ts` to remove non-canonical test users before or after suites

Agents MUST NEVER:

- Create new QA accounts, timestamp users, random emails, or duplicate QA identities
- Hardcode QA passwords in specs, helpers, or scripts
- Register users in Playwright except in **`frontend/playwright/tests/auth-*.spec.ts`** (registration-flow specs only)

Additional QA users require explicit approval before creation.

---

# Playwright Preconditions Rule

Before running Playwright against local or RCF targets (unless `PLAYWRIGHT_SKIP_QA_GOVERNANCE=1`):

1. Sync QA accounts — `backend/scripts/qa/verify-qa-accounts.ts`
2. Cleanup ephemeral users — `backend/scripts/qa/cleanup-ephemeral-users.ts`
3. Reset QA activity (optional per suite) — `backend/scripts/qa/reset-qa-data.ts`

Password source: **`backend/.env`** only (see `security-rules.md` §QA Credential Rule).

Verification spec: `frontend/playwright/qa-login-verification.spec.ts` (all four roles).

---

# CRITICAL PLAYWRIGHT RULE

Failing Playwright tests must be:

Investigated

↓

Fixed

↓

Retested

Do not:

* Skip tests
* Disable tests
* Ignore tests
* Mark failing tests as acceptable

---

# Screenshot Evidence Rule

Playwright verification must capture evidence.

Required screenshots:

Feature Entry

↓

Feature Interaction

↓

Feature Success State

↓

Critical Workflow Completion

Store screenshots as project evidence.

---

# Video Recording Rule

For critical workflows enable:

Playwright Video Recording

Examples:

Authentication

Payments

Enrollments

Course Creation

Admin Actions

Email Verification

---

# Authentication Verification Rule

Authentication is complete only when:

User Registers

↓

Database User Created

↓

Verification Email Sent

↓

Verification Link Works

↓

User Verified

↓

Login Works

↓

Protected Route Accessible

↓

Logout Works

All steps must be verified.

---

# Email Verification Rule

Email functionality must verify:

Email Generated

↓

Email Delivered

↓

Inbox Receives Email

↓

Link Extracted

↓

Link Opened

↓

Workflow Completed

SMTP configuration alone is not evidence.

---

# Password Reset Verification Rule

Password Reset must verify:

Reset Requested

↓

Token Generated

↓

Email Delivered

↓

Reset Link Opened

↓

Password Changed

↓

Login Works

---

# Course Verification Rule

Course Lifecycle:

Create Course

↓

Database Record Created

↓

Course Appears In UI

↓

Course Updated

↓

Changes Persist

↓

Course Retrieved

↓

Course Archived

All stages verified.

---

# Enrollment Verification Rule

Enrollment Workflow:

Enroll User

↓

Enrollment Created

↓

Access Granted

↓

Community Access Granted

↓

Analytics Updated

Verification required.

---

# Payment Verification Rule

Payment Workflow:

Checkout Started

↓

Payment Created

↓

Gateway Approved

↓

Webhook Received

↓

Order Updated

↓

Access Granted

↓

Receipt Generated

Verification required.

---

# AI Feature Verification Rule

AI Features require verification of:

Prompt Handling

Response Generation

Error Handling

Timeout Handling

Source Attribution

Rate Limiting

User Workflow

Do not assume AI works because a response appeared.

---

# Database Verification Rule

Database changes require:

Migration Executed

↓

Schema Verified

↓

Data Integrity Verified

↓

Queries Tested

↓

Rollback Strategy Verified

---

# Regression Testing Rule

Every change must be evaluated for regression risk.

Verify:

Existing Features

Authentication

Courses

Tests

Communities

Payments

AI

were not broken.

---

# Cross-Browser Rule

Critical workflows should be verified in:

Chromium

Firefox

WebKit

when practical.

---

# Mobile Testing Rule

Verify:

Mobile

Tablet

Desktop

for user-facing functionality.

Mobile-first verification is required.

---

# Accessibility Testing Rule

Verify:

Keyboard Navigation

Focus States

Screen Reader Support

Semantic Structure

Color Contrast

Accessibility failures are bugs.

---

# Performance Testing Rule

Verify:

Page Load Performance

API Response Times

Bundle Impact

Database Performance

Avoid introducing regressions.

---

# Evidence Collection Rule

Acceptable evidence:

* Test Results
* Playwright Reports
* Screenshots
* Videos
* Database Records
* API Responses
* Logs
* Monitoring Data

Evidence must be objective.

---

# Completion Language Rule

Forbidden:

"It should work now."

"It probably works."

"It appears fixed."

"It is likely resolved."

Required:

"Verified."

"Confirmed."

"Observed."

"Evidence attached."

"Playwright passed."

---

# Definition Of Done

A feature is DONE only when:

✓ Requirements Implemented

✓ Typecheck Passes

✓ Build Passes

✓ Unit Tests Pass

✓ Integration Tests Pass

✓ Playwright Passes (when UI/workflow affected)

✓ Manual Verification Complete (when applicable)

✓ Screenshots Captured (when required)

✓ Security Verified

✓ Documentation Updated (affected docs only)

✓ Regression Checked

✓ Evidence Exists

A task cannot be marked **COMPLETE** until every applicable item above is satisfied.

---

# Repository Cleanliness Rule

The repository must not accumulate technical debt from testing or abandoned work.

Remove or avoid:

- Dead code and unused exports
- Duplicate files and parallel implementations
- Unused scripts and legacy seed scripts
- Obsolete QA accounts and temporary database users
- Stale documentation that contradicts current governance

When legacy test identities or scripts are replaced, delete them in the same change set. Run `cleanup-ephemeral-users.ts` after migrations that retire timestamp-based test users.

---

# CRITICAL COMPLETION RULE

Code Written ≠ Complete

Build Success ≠ Complete

Tests Written ≠ Complete

Configuration Exists ≠ Complete

Feature Complete = Verified End-To-End + Evidence

---

# Deterministic Synchronization Rule

Fixed delays are a last resort.

Avoid:

- waitForTimeout()
- sleep()
- arbitrary delays
- retry loops without root cause analysis

Before using fixed delays, investigate:

- hydration completion
- application state readiness
- network completion
- database state readiness
- queue completion
- observable lifecycle events

Prefer deterministic synchronization:

- waitForFunction()
- event-based readiness
- application readiness indicators
- explicit hydration signals
- observable state transitions

A fixed delay may only be used when:

1. Root cause has been identified.
2. Deterministic synchronization is not practical.
3. Alternative solutions were evaluated.
4. Stability testing proves reliability.
5. Evidence is documented.

Time-based synchronization is not considered a root-cause fix by default.

# Synchronization Verification Rule

When introducing waits in automated tests:

Prefer:

- readiness signals
- lifecycle events
- application state
- observable conditions

over:

- fixed delays
- arbitrary timeouts
- sleep statements

Fixed delays require evidence.

The AI must demonstrate:

1. Root cause identified.
2. Alternative synchronization methods evaluated.
3. Deterministic methods are unavailable or impractical.
4. Reliability testing performed.
5. Evidence collected.

A passing test with a timeout is not proof that the timeout is the correct solution.

---

# Production Verification Rule

Before claiming **Ready**, **Done**, **Released**, or **Certified**, the following must exist (as applicable to the change):

| Gate | Required |
|------|----------|
| Typecheck | ✅ Pass |
| Build | ✅ Pass |
| Unit / integration tests | ✅ Pass |
| Playwright | ✅ Pass (when UI/workflow affected) |
| Screenshots | ✅ Captured (when required) |
| Logs / traces | ✅ Available for failures |
| Documentation | ✅ Updated (affected docs) |
| Rollback plan | ✅ Documented (when deploy/migration) |

"Tests pass" alone is insufficient for production-facing claims. See also `AGENTS.md` §Breaking Changes Disclosure and `operations-rules.md` §No Silent Breaking Changes Rule.

---

# Final Directive

Never replace testing with confidence.

Never replace verification with assumptions.

Never replace implementation with proof.

If evidence does not exist:

The feature is not proven.

If proof does not exist:

The task is not complete.


TEST EVIDENCE RULE

A Playwright test is NOT considered executed unless all of the following are provided:

1. Exact command executed
2. Raw Playwright summary
3. Total tests run
4. Passed count
5. Failed count
6. Exit code
7. results.json location
8. Screenshot evidence (if required)

Timers, waits, sleeps, countdowns, retries, or "timer expired" messages are NOT accepted as proof of execution.