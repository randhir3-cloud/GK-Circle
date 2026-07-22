# GK Circle Operations Rules

Version: 1.1

Status: Mandatory

**v1.1 (2026-07-12):** No Silent Breaking Changes, Commit Rule, Repository Health Rule. Supplements v1.0; AGENTS.md wins on conflict.

---

# Purpose

These rules govern deployment, infrastructure, DevOps, CI/CD, monitoring, release management, disaster recovery, backups, production readiness, and operational excellence across GK Circle.

Operational failures are platform failures.

Production readiness is mandatory.

---

# Production First Rule

Every system must be built assuming it will eventually run in production.

Avoid:

* Temporary production shortcuts
* Untracked infrastructure changes
* Manual production fixes
* Hidden dependencies

Build production-ready systems.

---

# Infrastructure As Code Rule

Infrastructure must be reproducible.

Prefer:

* Docker
* Docker Compose
* Terraform (future)
* Kubernetes manifests (future)

Avoid undocumented manual server configuration.

---

# Environment Rule

Maintain separate environments:

Development

↓

Staging

↓

Production

Do not mix environments.

Never test experimental features directly in production.

---

# Configuration Rule

Configuration must be environment-based.

Use:

.env

Secret Managers

Environment Variables

Do not hardcode environment-specific values.

---

# Secrets Rule

Secrets must never be:

Committed

Logged

Exposed

Hardcoded

Examples:

JWT Secrets

SMTP Credentials

Database Passwords

API Keys

Payment Secrets

AI Provider Keys

---

# CI/CD Rule

All deployments must be automated.

Minimum pipeline:

Lint

↓

Type Check

↓

Unit Tests

↓

Integration Tests

↓

Build

↓

Security Checks

↓

Deploy

Manual deployments should be minimized.

---

# Build Verification Rule

A successful build does not prove a feature works.

Build verification must be followed by:

Testing

↓

Playwright Verification

↓

Evidence Collection

---

# Deployment Rule

Every deployment must have:

Rollback Strategy

Health Checks

Monitoring

Deployment Verification

Do not deploy blindly.

---

# Release Management Rule

Releases should include:

Features

Bug Fixes

Database Changes

Infrastructure Changes

Security Updates

Document all releases.

---

# Feature Flag Rule

Experimental functionality must use feature flags.

Do not expose unfinished functionality.

Feature flags must support:

Enable

Disable

Rollback

without redeployment.

---

# Monitoring Rule

Monitor:

Frontend

Backend

Database

Queues

Email

Payments

AI Systems

OCR

Storage

Monitoring is mandatory.

---

# Observability Rule

Critical workflows must provide:

Logs

Metrics

Tracing

Alerts

Sentry Events

Systems that cannot be observed cannot be trusted.

---

# Alerting Rule

Critical failures must trigger alerts.

Examples:

Authentication Failures

Payment Failures

Database Failures

Queue Failures

AI Failures

Email Failures

Infrastructure Failures

---

# Logging Rule

Logs must be:

Structured

Searchable

Useful

Avoid:

console.log

for production diagnostics.

Logs should answer:

Who

What

When

Why

---

# Health Check Rule

Every service should expose health endpoints.

Verify:

Application

Database

Redis

Queues

External Providers

before reporting healthy.

---

# Backup Rule

Critical systems require backups.

Backup:

Database

Storage Metadata

Configuration

Important Operational Data

Backups must be automated.

---

# Backup Verification Rule

A backup is not valid because it exists.

Verify:

Backup Created

↓

Backup Restored

↓

Data Valid

Only then is backup verified.

---

# Disaster Recovery Rule

Prepare for:

Database Failure

Storage Failure

Server Failure

Region Failure

Provider Failure

Recovery plans must exist.

---

# Recovery Time Objective Rule

Define:

RTO (Recovery Time Objective)

RPO (Recovery Point Objective)

for critical systems.

Document targets.

---

# Database Operations Rule

Database changes require:

Migration Plan

Rollback Plan

Testing

Verification

Avoid direct production modifications.

---

# Production Data Rule

Never use production data:

* For local development
* For testing
* For demos

without proper protection and approval.

---

# Scaling Rule

Design systems assuming future scale.

Plan for:

Millions of Users

Millions of Tests

Millions of Questions

Millions of Enrollments

Avoid architecture that only works at small scale.

---

# Queue Operations Rule

Background jobs must be:

Observable

Retryable

Monitored

Recoverable

Silent failures are prohibited.

---

# Email Operations Rule

Email systems must track:

Sent

Delivered

Failed

Bounced

Rejected

Do not assume delivery.

---

# Payment Operations Rule

Monitor:

Checkout

Payment Provider

Webhooks

Order Updates

Access Grants

Payment systems require end-to-end observability.

---

# AI Operations Rule

Track:

Requests

Costs

Latency

Failures

Provider Errors

Rate Limits

AI systems must be measurable.

---

# Security Operations Rule

Regularly perform:

Dependency Audits

Security Reviews

Access Reviews

Secret Rotation

Permission Reviews

Security is continuous.

---

# Incident Response Rule

When incidents occur:

Detect

↓

Contain

↓

Investigate

↓

Fix

↓

Verify

↓

Document

Never hide incidents.

---

# Postmortem Rule

Major incidents require postmortems.

Document:

What Happened

Root Cause

Impact

Resolution

Prevention Plan

---

# Operational Documentation Rule

Maintain:

Runbooks

Deployment Guides

Recovery Procedures

Infrastructure Documentation

Incident Procedures

Documentation is part of operations.

---

# Production Readiness Checklist

Before production release:

✓ Tests Pass

✓ Playwright Passes

✓ Security Review Complete

✓ Monitoring Enabled

✓ Alerts Configured

✓ Backups Verified

✓ Rollback Plan Exists

✓ Documentation Updated

✓ Evidence Collected

---

# Final Directive

Reliable systems are built through discipline.

Operational excellence is not optional.

Production readiness is a requirement.

If a system cannot be deployed, monitored, recovered, and maintained safely:

It is not complete.

---

# v1.0 Addendum — Architecture-Intent Standard

This addendum extends the v0.5 baseline above with architecture-intent depth. The v0.5 rules above remain in force. This addendum provides the schema, the matrix, and the gate criteria that the v0.5 rules reference.

## Dependency Rule (AGENTS.md Wins)

If a rule in this file conflicts with [AGENTS.md](../../AGENTS.md), AGENTS.md wins. If a rule in this file conflicts with another standards file, the more-specific file wins (live-exam-specific observability lives in [live-exam-rules.md](live-exam-rules.md); security audit lives in [admin-panel-rules.md](admin-panel-rules.md); financial observability lives in [creator-economy-rules.md](creator-economy-rules.md)). If both files are at equal specificity, the more recently updated file wins. Record the decision in [../active-context/decisions.md](../active-context/decisions.md).

## Verification Requirement (Existence ≠ Evidence)

A monitoring check, alert, runbook, or backup is not "done" because the file exists. It is done when:

- The check has been fired by a synthetic incident or a chaos test, and the alert reached the right human within the SLA
- The runbook has been executed end-to-end by someone other than the author, with screenshots/timestamps captured
- The backup has been restored to a scratch environment, schema and row count verified, and the restore drill logged

Existence is the start. Evidence is the finish.

---

# Observability — Five Pillars (Mandatory)

GK Circle uses the five-pillar observability model. Each pillar has a single, named tool, a single, named schema, and a single, named owner. The pillars are **not interchangeable** — they answer different questions:

| # | Pillar | Question it answers | Tool | Owner |
|---|---|---|---|---|
| 1 | **Logs** | What happened? | OpenSearch + Pino (Node) / stdlib logging (workers) | Platform |
| 2 | **Metrics** | How much / how often? | Prometheus + Grafana | Platform |
| 3 | **Traces** | Where did the time go? | OpenTelemetry (OTLP) + Jaeger | Platform |
| 4 | **Errors** | What broke and where? | Sentry | App Eng |
| 5 | **Audit Events** | Who did what, when, and what changed? | Append-only `AuditLog` table + replicated cold storage | Security |
| 6 | **Business Metrics** | Is the Course working? | ClickHouse + Metabase | Growth |

**The pillars are not mixed.** A Sentry error does not contain log lines. A log line does not contain a metric. A metric does not contain a trace. Cross-pillar linking is done by `correlationId`, `userId`, `liveExamSessionId`, etc., not by stuffing one pillar's data into another.

Every service MUST emit to **all five operational pillars (1–5)**. Business metrics (6) are emitted by user-facing event streams, not by every service. A service that emits only logs (a common anti-pattern) is a Phase 0.3 integrity violation.

## 1. Logs (Mandatory)

- Structured JSON, emitted via shared `Logger` wrapper
- Standard fields: `ts` (ISO 8601 UTC ms), `level`, `service`, `version`, `env`, `region`, `traceId`, `spanId`, `correlationId`, `userId`, `message`
- Levels: `debug` (1% sampled in prod), `info`, `warn`, `error`, `fatal`
- **PII redaction at the wrapper** — not at the call site. Fields automatically stripped: `password`, `token`, `apiKey`, `secret`, `authorization`, `cookie`, `setCookie`, `aadhaar`, `pan`, `cardNumber`, `cvv`, `requestBody`, `responseBody`
- Volume budget: ≤ 1 MB/minute at steady state per service. > 2 MB/minute for > 5 minutes → `SEV-3` alert
- Raw request/response bodies are forbidden in logs (PII and IP risk)

## 2. Metrics (Mandatory)

- Every service exposes a Prometheus `/metrics` endpoint (scraped every 15s)
- Naming: `gkc_<subsystem>_<name>_<unit>` (e.g., `gkc_live_exam_session_started_total`, `gkc_rag_query_latency_seconds`)
- Required per service: `gkc_<svc>_request_total{route,method,status}`, `gkc_<svc>_request_latency_seconds{...}`, `gkc_<svc>_error_total{...}`, `gkc_<svc>_active_connections{kind}`, `gkc_<svc>_up`
- **High-risk domain metrics** (must emit in addition to the standard set):
  - Live Exam: `gkc_live_exam_session_state_transition_total{from,to}`, `gkc_live_exam_timer_drift_seconds`, `gkc_live_exam_anti_cheat_triggered_total{trigger}`
  - RAG: `gkc_rag_query_citation_coverage_ratio`, `gkc_rag_query_faithfulness_score`, `gkc_rag_query_token_cost_inr_total`
  - Creator Economy: `gkc_creator_ledger_invariant_violation_total`, `gkc_creator_payout_processing_seconds`
  - Payment: `gkc_payment_gateway_latency_seconds{gateway}`, `gkc_payment_refund_pending_total`
  - Auth: `gkc_auth_login_total{outcome,method}`, `gkc_auth_mfa_challenge_total{outcome}`

## 3. Traces (Mandatory)

- OpenTelemetry SDK, OTLP → collector → Jaeger
- Propagation: HTTP (`traceparent`), gRPC, Postgres (via `pg`), Redis (via `ioredis`), Socket.IO (custom propagator), BullMQ (custom span)
- Must trace: every HTTP request (root), every DB query (child), every external API call (Razorpay, OpenAI, Qdrant, Sentry SDK), every BullMQ job, every WebSocket event, every RAG pipeline stage (retrieval, re-rank, generation, citation verification)
- Must NOT trace: logging the span, health probe spam, high-frequency heartbeats
- Sampling: prod 1% head + 100% tail (errors, p99 latency, anti-cheat events)

## 4. Sentry (Mandatory)

- Wrapped in a `SentryWrapper` (per Global Sentry Instrumentation in AGENTS.md)
- `sendDefaultPii: false` (always); custom PII scrubber; `beforeSend` strips PII regex matches
- `release` = deployed git SHA; `environment` = dev/staging/prod; `user.id` set after auth (no PII)
- Tags: `correlationId`, `liveExamSessionId`, `testId` where relevant
- Sentry is **not** used for logs (use OpenSearch) or for business metrics (use ClickHouse)
- Required alarms: new error type → page in 5 min; error rate > 1% for 5 min → page; release regression → page release captain

## 5. Audit Events (Mandatory — separate from logs)

- Append-only `AuditLog` table (per [admin-panel-rules.md](admin-panel-rules.md) §"Append-Only Audit Log")
- **Not** a Sentry feature. **Not** a log line. **Not** a metric.
- 7-year retention (regulatory). Tamper-evident via content-hash chain. Replicated to immutable cold storage.
- Generated for: every action in [admin-panel-rules.md](admin-panel-rules.md) §"Action Catalog"; every payout state change; every grade change to a Live Test answer; every AI moderation action

## 6. Business Metrics (Mandatory — separate from operations metrics)

- ClickHouse + Metabase
- Required: DAL, WAL, Questions Solved, Tests Completed, Study Time, Community Engagement, Learning Retention (D7/D30), AI Usage, Creator Revenue, Conversion, Churn
- Refreshed hourly (or daily for cohort metrics)
- **Never** contains PII (aggregated counts only) or secrets or internal IP addresses

---

# Alerting Tiers (Mandatory)

| Tier | Mechanism | Reaches | When |
|---|---|---|---|
| **Page** | PagerDuty → phone | On-call | SEV-1 and SEV-2 |
| **Page (after 5 min)** | PagerDuty → phone | On-call | SEV-3 if not acknowledged in 5 min |
| **Ticket** | Jira | Service owner | SEV-3 (acked) and SEV-4 |
| **Slack** | #gkc-alerts | Channel | All severities |
| **Email** | ops@gkcircle.in | Distribution list | Daily digest of SEV-3 and SEV-4 |

**Alert hygiene rules** (every alert must satisfy all):
- Actionable (tells the responder what to do first; "investigate" is not an action)
- Has a runbook link
- Has a deduplication key (5-min window)
- Has a primary + secondary owner
- Has an SLA (max unacknowledged time)
- Auto-resolves (alerts that resolve within 10 min page no one)
- Tested in the chaos suite (no alert without a test)

---

# Production Incident Severity Matrix (Mandatory)

The contract between engineering, support, and leadership. Reviewed quarterly.

## SEV-1 — Platform Unusable

**Definition**: Platform unavailable to a significant fraction of users OR a critical integrity guarantee is violated. Material revenue loss. Material trust damage.

**Examples**:
- Site 100% down (5xx > 50% for > 2 minutes)
- Database unreachable from all services
- Confirmed security breach (data exfiltration, unauthorized admin access)
- All admin actions failing (moderation impossible)
- All live exams failing to start (tens of thousands affected)
- Creator payouts miscalculated (ledger invariant violated) for > 1 hour
- AI Tutor generating harmful content at scale (faithfulness score < 50%)

**Response**:
- Acknowledge: ≤ 5 minutes
- First status update: ≤ 15 minutes
- Status page update: within 15 minutes
- Incident commander: on-call lead
- Comms lead: VP Engineering or delegate
- Customer comms: status page + email to affected users within 1 hour
- Resolution target: ≤ 4 hours
- Postmortem: required, published within 5 business days, **public**

**Authority of incident commander** (unilateral):
- Roll back any deploy
- Failover any database
- Disable any feature flag
- Take the platform into maintenance mode
- Engage the leadership war room

## SEV-2 — Major Feature Unavailable

**Definition**: Major feature broken for many users. Platform otherwise up. Workaround exists but is painful.

**Examples**:
- Login failing for 10%+ of users
- Live exam timer drifting (server vs client > 5 seconds)
- RAG returning answers without citations
- Creator payout processing delayed > 24 hours
- One payment gateway down (Razorpay down, Cashfree up)
- AI MCQ explanation broken
- Mobile app crashes on launch for > 5% of users on a specific OS

**Response**:
- Acknowledge: ≤ 15 minutes
- First status update: ≤ 30 minutes
- Status page update: within 30 minutes
- Incident commander: on-call lead
- Resolution target: ≤ 24 hours
- Postmortem: required, published within 10 business days

**Authority**: Can roll back a deploy, disable a feature flag, but NOT failover a database without VP Eng sign-off.

## SEV-3 — Partial Degradation

**Definition**: Feature degraded. Most users unaffected. Workaround clear and not painful.

**Examples**:
- A single notification type (e.g., push) failing
- Search returns results in 3s instead of 300ms
- Image uploads failing for 1% of users
- Admin moderation action slow (> 30s)
- A specific test series has a broken question
- Admin report shows stale data (> 1 hour old)

**Response**:
- Acknowledge: ≤ 1 hour
- First status update: ≤ 4 hours
- Status page update: only if issue persists > 4 hours
- Incident commander: on-call engineer (no commander, just an owner)
- Resolution target: ≤ 5 business days
- Postmortem: optional, encouraged if it recurs

## SEV-4 — Minor Issue

**Definition**: Cosmetic, edge-case, or single-user. Not user-visible at scale.

**Examples**:
- Typo in an admin label
- Rare crash on a specific device model
- Badge that doesn't unlock
- Minor data inconsistency that resolves itself
- New feature flag not behaving exactly as expected, no user impact

**Response**:
- Acknowledge: next business day
- First status update: next status meeting
- Status page update: never
- Incident commander: ticket owner
- Resolution target: next sprint
- Postmortem: not required

## Severity Decision Authority

- On-call engineer assigns initial severity within 15 minutes
- Incident commander can change severity as the situation evolves
- SEV-1 and SEV-2 must be confirmed by VP Engineering or delegate within 1 hour
- SEV-3 and SEV-4 are at the on-call engineer's discretion
- Disagreement → escalate to VP Engineering

---

# Backups — Required Schedule (Mandatory)

| Data | Method | Frequency | Hot Retention | Cold Retention |
|---|---|---|---|---|
| PostgreSQL primary | `pg_dump` + WAL archiving | daily + continuous WAL | 35 days | 1 year |
| PostgreSQL replicas | WAL streaming | continuous | 35 days | 1 year |
| Redis (sessions, cache) | RDB snapshot + AOF | hourly + continuous AOF | 7 days | — |
| Object storage (S3) | Versioned + cross-region replication | on write | indefinite | indefinite |
| ClickHouse | Native backup to S3 | daily | 90 days | — |
| Qdrant vectors | Snapshot to S3 | daily | 30 days | — |
| AuditLog | Replicated to immutable cold storage | on write | 1 year | 6 years (7 total) |
| Configuration (Terraform, K8s manifests) | Git + OCI registry for images | on commit | indefinite | — |
| Secrets | **NOT backed up** — stored in secret manager only | — | — | — |

All backups encrypted AES-256 at rest.

## Backup Verification (Mandatory)

- **Weekly**: automated restore of a random backup to a scratch environment, schema + row count verified
- **Monthly**: full restore drill — entire database restored to sandbox, end-to-end smoke test
- **Quarterly**: disaster recovery drill — entire region failover exercised
- **Annually**: third-party audit of backup completeness and recoverability

A backup is not verified because it exists. It is verified because it has been **restored**.

---

# Recovery — RTO / RPO Matrix (Mandatory)

| Data Tier | RTO (Recovery Time Objective) | RPO (Recovery Point Objective) |
|---|---|---|
| **Critical** (Postgres, Redis, Object storage) | 1 hour | 5 minutes (WAL streaming) |
| **Important** (ClickHouse, Qdrant) | 4 hours | 24 hours |
| **Standard** (analytics exports) | 24 hours | 24 hours |
| **Cold** (archived reports) | 72 hours | 7 days |

## Disaster Recovery Tiers

| Tier | Failure mode | Recovery procedure | Tested |
|---|---|---|---|
| Single instance | One VM / pod crashes | Auto-restart, no action | Continuously |
| Single AZ | One availability zone down | Auto-failover to other AZs | Quarterly |
| Single region | One region down | Manual failover to standby region, 1-hour RTO | Quarterly |
| Data corruption | Database corruption detected | Restore from backup + replay WAL, 4-hour RTO | Quarterly |
| Ransomware / malicious admin | Encryption or deletion by attacker | Restore from immutable cold storage, 24-hour RTO | Annually |
| Account compromise | AWS / GCP / GitHub account compromised | Rotate all credentials, restore from backup, 24-hour RTO | Annually |

## Runbook Requirements (Mandatory)

Every runbook has: title; severity tier; pre-conditions; numbered step-by-step procedure (copy-pasteable); verification steps; rollback steps (what to do if the recovery itself fails); owner. A runbook is not "done" until it has been executed end-to-end at least once.

---

# Feature Flag Lifecycle (Mandatory)

| State | Description | Who can flip |
|---|---|---|
| `OFF` (default) | Code path not taken, no user impact | Engineering |
| `INTERNAL` | Only GK Circle employees and test users | Engineering |
| `BETA` | Opt-in users or % rollout | Engineering + Course |
| `ON` (default for all) | Feature is live for all | Engineering + Course |
| `KILLED` | Code removed, flag removed | Engineering |

**Requirements** (every flag must satisfy):
- Has an owner
- Has an expiry date (default 90 days after `ON`)
- Has a kill switch (operator can flip to `OFF` in < 30 seconds)
- Flags changing paid-feature behavior require Course + Eng lead sign-off
- Flags changing security-sensitive behavior (auth, RBAC, payment) require a security review
- Stale flags (> 90 days at `ON`) appear in the weekly cleanup report

---

# SLA / SLO — Service Tiers (Mandatory)

| Tier | Services | Uptime SLO | p95 Latency | Error Rate |
|---|---|---|---|---|
| **Tier 1 (Critical)** | auth, live-exam, payment, billing | 99.95% (≤ 4.4 hr/yr) | 500ms | 0.1% |
| **Tier 2 (Important)** | community, rag, mcq, test, learning | 99.9% (≤ 8.8 hr/yr) | 1s | 0.5% |
| **Tier 3 (Best effort)** | analytics, admin panel, marketing site, search | 99.5% (≤ 43.8 hr/yr) | 2s | 1% |

## Error Budget (Mandatory)

For each Tier, `error_budget = 100% - SLO`. If a service consumes > 50% of its monthly error budget, new feature deploys are **paused** until the budget is restored or an exception is granted by VP Engineering. Tracked in Grafana, reviewed in the weekly SRE sync.

## User-Facing SLAs (Customer Contract)

- GK Circle platform: 99.9% monthly uptime (free tier)
- GK Circle Pro: 99.95% monthly uptime (paid tier)
- Live Tests: 99.99% per-exam reliability. If a Live Test is canceled by a platform issue, the test is rescheduled and all enrolled users receive 2× credit.

---

# Data Retention (Mandatory)

| Data Class | Hot Storage | Cold Storage | Total Retention | Deletion Method |
|---|---|---|---|---|
| **Logs** (OpenSearch) | 30 days | 60 days (compressed) | 90 days | TTL + scheduled delete |
| **Metrics** (Prometheus) | 15 days | 365 days (downsampled) | 1 year | TTL + downsampling |
| **Traces** (Jaeger) | 7 days | — | 7 days | TTL |
| **Sentry errors** | 90 days | — | 90 days | Sentry TTL |
| **Audit events** | 1 year (Postgres) | 6 years (cold S3) | **7 years** | Append-only; no deletion except by legal hold release |
| **Analytics events** (ClickHouse) | 90 days | 2 years (aggregated) | 2 years | TTL + aggregation |
| **Backups** (Postgres) | 35 days | 1 year | 1 year | TTL |
| **Backups** (ClickHouse) | 30 days | 90 days | 4 months | TTL |
| **Backups** (Qdrant) | 30 days | — | 30 days | TTL |
| **User-generated content** (posts, comments, MCQ attempts) | indefinite | indefinite | indefinite | User-initiated delete + GDPR right-to-erasure |
| **User PII** | indefinite | indefinite | until account deletion + 30 days | GDPR right-to-erasure + soft delete with 30-day recovery |
| **Financial records** (orders, refunds, payouts) | 7 years (Postgres) | 7 years (cold S3) | **7 years** | Append-only, regulatory |

## Enforcement (Mandatory)

- TTLs at the storage level (S3 lifecycle, OpenSearch ISM, ClickHouse TTL, Postgres partitioning)
- Weekly job reports any data exceeding its retention and verifies deletion
- Quarterly audit verifies retention policy is being applied (drift detection)
- Owned by the DPO role; reviewed annually

## Right to Erasure (GDPR / DPDP Act)

- User-initiated account deletion triggers 30-day soft delete (recoverable in that window)
- After 30 days, PII anonymized (user row preserved for referential integrity; PII fields nulled)
- Audit log entries **not** deleted (regulatory) but PII within them redacted to a hash
- Financial records **not** deleted (regulatory)

---

# Compatibility With Other Standards

| Standard | Relationship |
|---|---|
| [architecture-rules.md](architecture-rules.md) | This file is the operational implementation of the architecture's service tiers |
| [security-rules.md](security-rules.md) | Audit events here are governed by the audit log rules there |
| [backend-rules.md](backend-rules.md) | Logger / metrics / traces wrappers are built per the backend rules |
| [devops-rules.md](devops-rules.md) | Environment matrix and CI/CD live in devops-rules.md; runbooks live here |
| [admin-panel-rules.md](admin-panel-rules.md) | Audit events here are emitted by the actions defined there |
| [creator-economy-rules.md](creator-economy-rules.md) | Creator metrics here are computed from the ledger defined there |
| [live-exam-rules.md](live-exam-rules.md) | Live exam metrics here are emitted by the engine defined there |
| [rag-rules.md](rag-rules.md) | RAG metrics here measure the system defined there |

---

# Sprint 1 Compliance Checklist

- [ ] All 5 observability pillars emitting from live-exam, auth, payment, creator-economy services
- [ ] Sentry configured with `sendDefaultPii: false` and the PII scrubber
- [ ] Prometheus `/metrics` endpoints exposed by every backend service
- [ ] OpenTelemetry instrumentation covers HTTP, DB, Redis, Socket.IO, BullMQ
- [ ] Tail-based sampling configured in the OTel collector
- [ ] Audit events emitted for every action in [admin-panel-rules.md](admin-panel-rules.md) §"Action Catalog"
- [ ] All four SEV tiers have a runbook with at least one executed end-to-end test
- [ ] Backup verification job has run at least weekly for 4 weeks
- [ ] RTO/RPO targets documented in the runbook and verified by a drill
- [ ] Feature flags for all new features are in place, with owners and expiry dates
- [ ] Tier 1 service SLOs (99.95% uptime, 500ms p95, 0.1% error) are in Grafana
- [ ] Error budget is tracked in Grafana
- [ ] Data retention TTLs are applied at the storage level
- [ ] GDPR right-to-erasure flow is implemented and tested

---

---

# No Silent Breaking Changes Rule

Every change that may affect consumers must disclose:

**Breaking Changes: YES / NO**

If **YES**, document before merge or release:

- Affected APIs (endpoints, payloads, status codes)
- Affected routes (frontend paths, redirects)
- Affected database (schema, data migration, rollback SQL)
- Migration required (YES/NO + path to migration/script)
- Rollback procedure (steps to revert safely)

Agents must state this in completion reports. CI/release owners must reject releases with undocumented breaking changes.

---

# Commit Rule

Every commit must be:

- **One logical change** — single purpose
- **Reversible** — safe to revert without orphan state
- **Documented** — message explains why, not only what
- **Tested** — related tests pass for the change scope

Never mix in one commit:

- Bug fix + feature
- Feature + refactor
- Refactor + cleanup
- Unrelated file changes

Squashing unrelated work to "save time" violates this rule.

---

# Repository Health Rule

Every implementation should leave the repository in a **better** state than before:

- Fewer warnings (lint, typecheck, build)
- Fewer unresolved TODOs in touched files (resolve or ticket)
- No new dead files or unused exports in touched areas
- Better docs for changed behavior
- Better tests for changed behavior

Do **not** increase technical debt unless explicitly approved. Opportunistic cleanup in touched files is encouraged when scope-safe.

---

# Final Directive (v1.0)

A service without observability is a service that cannot be operated.
A service without a runbook is a service that cannot be recovered.
A service without an SLO is a service that cannot be improved.
An alert without a runbook is noise.
A backup that has never been restored is a hope, not a backup.
A SEV-1 without a public postmortem is a missed lesson.

Every service ships with: all 5 operational pillars + a runbook + an SLO + a feature flag (for new behavior) + an owner.

Every alert is actionable, has a runbook link, and is tested.

Every SEV-1 is followed by a public postmortem.

Operations is not a separate concern. Operations is part of the platform.
