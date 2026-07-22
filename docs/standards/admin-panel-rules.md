# GK Circle Admin Panel Rules

Version: 1.0

Status: Mandatory

---

# Purpose

Govern Super Admin and Administrative Systems.

---

# Super Admin Rule

Super Admin controls:

Users

Courses

Communities

Payments

Creators

AI

Analytics

Moderation

---

# Audit Rule

Every admin action must create:

Audit Log

↓

Actor

↓

Action

↓

Target

↓

Timestamp

---

# Moderation Rule

Support:

Content Review

User Reports

Community Moderation

Creator Verification

---

# AI Monitoring Rule

Admins must monitor:

AI Usage

Costs

Errors

Provider Health

---

# Analytics Rule

Provide platform-wide metrics:

Users

Revenue

Retention

Engagement

Learning Outcomes

---

# Security Rule

Admin access requires:

RBAC

MFA

Session Monitoring

Activity Logging

---

# v0.5 Addendum — Architecture-Intent Standard

The v1.0 section above defines the principles (Super Admin scope, Audit, Moderation, AI Monitoring, Analytics, Security with RBAC/MFA). This addendum makes those principles enforceable as code, configuration, and process. The structure mirrors the other v0.5 standards.

This is **architecture-intent**: the admin panel is partially implemented. The rules below are the contract it MUST satisfy. The next revision (v1.0 of this file) will merge these rules into the body.

The admin panel is the highest-trust surface in GK Circle. A compromised admin account is a full platform compromise. The standards below are tuned for that risk profile.

---

# Dependency Rule (Mandatory)

This document does not replace AGENTS.md.

It supplements AGENTS.md.

If this document conflicts with AGENTS.md, AGENTS.md wins.

If this document conflicts with another standards file, the more specific standard wins.

If ambiguity exists, document the ambiguity in an ADR before implementation.

# Verification Requirement (Mandatory)

A rule is not considered satisfied because code, configuration, tests, or documentation exist. A rule is satisfied only when evidence exists. For admin panel, evidence must include:

- A penetration test confirming that an admin without the required permission cannot perform the action (every protected endpoint is tested)
- A log review confirming that every admin action produces an audit log entry
- An MFA enforcement test confirming that an admin without MFA cannot access the panel
- An impersonation audit confirming that impersonation actions are logged and the impersonated user is notified

---

# Audit Logs (Mandatory)

Every admin action MUST create an `AuditLog` entry. This is the foundation of admin accountability and the legal record of platform operations.

## Audit Log Fields (Mandatory)

```typescript
interface AuditLog {
  id: string;                      // UUID
  actorId: string;                 // Admin user who performed the action
  actorRole: 'SUPER_ADMIN' | 'ADMIN' | 'MODERATOR' | 'SUPPORT' | 'SYSTEM';
  action: string;                  // See Action Catalog below
  targetType: 'USER' | 'Course' | 'COMMUNITY' | 'POST' | 'COMMENT' | 'TEST' | 'LESSON' | 'PAYMENT' | 'ENROLLMENT' | 'AI_CONFIG' | 'PROMPT' | 'EMBEDDING_MODEL' | 'ROLE' | 'PERMISSION' | 'SUBSCRIPTION' | 'COUPON' | 'REFUND' | 'OTHER';
  targetId: string;                // ID of the target entity
  metadata: Record<string, any>;   // Action-specific context
  ipAddress: string;               // Source IP
  userAgent: string;               // Browser/app
  sessionId: string;               // Admin session
  impersonatedUserId?: string;     // If impersonating, the target user
  reason?: string;                 // Why the action was taken
  createdAt: string;               // ISO 8601
}
```

The `AuditLog` table is **append-only**. No update, no delete, no soft delete. Even Super Admin cannot modify an existing audit log entry. A `corrections` mechanism (a new audit log entry referencing the original) is used if a correction is needed.

## Action Catalog (Mandatory)

The `action` field is a string in the format `<domain>.<verb>`. The following is the canonical catalog:

### User Management

| Action | Description | Required Permission |
|---|---|---|
| `user.create` | Create a new user (admin-created) | `user:create` |
| `user.update` | Update user profile fields | `user:update` |
| `user.suspend` | Suspend a user account | `user:suspend` |
| `user.unsuspend` | Lift a suspension | `user:suspend` |
| `user.delete` | Soft-delete a user | `user:delete` |
| `user.restore` | Restore a soft-deleted user | `user:delete` |
| `user.verify` | Mark a user as verified (e.g., KYC) | `user:verify` |
| `user.unverify` | Remove verification | `user:verify` |
| `user.role.assign` | Assign a role to a user | `role:assign` |
| `user.role.revoke` | Revoke a role | `role:assign` |
| `user.impersonate.start` | Begin impersonating a user | `user:impersonate` |
| `user.impersonate.end` | End impersonation | `user:impersonate` |
| `user.password.reset` | Force a password reset | `user:update` |
| `user.mfa.reset` | Reset MFA for a user | `user:update` |

### Course Management

| Action | Description | Required Permission |
|---|---|---|
| `course.create` | Create a Course | `course:create` |
| `course.update` | Update Course fields | `course:update` |
| `course.publish` | Publish a Course | `course:publish` |
| `course.unpublish` | Unpublish a Course | `course:publish` |
| `course.archive` | Archive a Course | `course:update` |
| `course.delete` | Soft-delete a Course | `course:delete` |
| `course.pricing.update` | Change pricing | `course:update` |
| `course.visibility.update` | Change visibility | `course:update` |
| `course.feature` | Feature a Course (homepage) | `course:feature` |
| `course.unfeature` | Remove from featured | `course:feature` |

### Community Moderation

| Action | Description | Required Permission |
|---|---|---|
| `post.lock` | Lock a post (no more comments) | `community:moderate` |
| `post.unlock` | Unlock a post | `community:moderate` |
| `post.pin` | Pin a post | `community:moderate` |
| `post.unpin` | Unpin a post | `community:moderate` |
| `post.remove` | Remove a post | `community:moderate` |
| `post.restore` | Restore a removed post | `community:moderate` |
| `post.shadow` | Shadow-delete a post (only author sees) | `community:moderate` |
| `comment.remove` | Remove a comment | `community:moderate` |
| `comment.restore` | Restore a removed comment | `community:moderate` |
| `community.lock` | Lock a community | `community:moderate` |
| `community.unlock` | Unlock a community | `community:moderate` |
| `community.ban.user` | Ban a user from a community | `community:moderate` |
| `community.unban.user` | Unban a user | `community:moderate` |

### Financial

| Action | Description | Required Permission |
|---|---|---|
| `payment.refund` | Refund a payment | `payment:refund` |
| `payment.refund.partial` | Partial refund | `payment:refund` |
| `enrollment.grant` | Grant a free enrollment | `enrollment:grant` |
| `enrollment.revoke` | Revoke an enrollment | `enrollment:grant` |
| `coupon.create` | Create a coupon | `coupon:create` |
| `coupon.update` | Update a coupon | `coupon:update` |
| `coupon.delete` | Delete a coupon | `coupon:delete` |
| `payout.approve` | Approve a creator payout | `payout:approve` |
| `payout.reject` | Reject a payout | `payout:approve` |
| `payout.hold` | Hold a payout for review | `payout:approve` |

### AI

| Action | Description | Required Permission |
|---|---|---|
| `ai.prompt.deploy` | Deploy a new prompt version | `ai:prompt` |
| `ai.prompt.rollback` | Rollback to a previous prompt version | `ai:prompt` |
| `ai.model.enable` | Enable a new AI model | `ai:model` |
| `ai.model.disable` | Disable an AI model | `ai:model` |
| `ai.embedding.reindex.start` | Start a full re-indexing job | `ai:reindex` |
| `ai.embedding.reindex.cancel` | Cancel a re-indexing job | `ai:reindex` |

### Platform

| Action | Description | Required Permission |
|---|---|---|
| `role.create` | Create a new role | `role:create` |
| `role.update` | Update role permissions | `role:update` |
| `role.delete` | Delete a role | `role:create` |
| `permission.grant` | Grant a permission to a role | `role:update` |
| `permission.revoke` | Revoke a permission | `role:update` |
| `feature_flag.enable` | Enable a feature flag | `platform:flag` |
| `feature_flag.disable` | Disable a feature flag | `platform:flag` |
| `config.update` | Update a platform configuration value | `platform:config` |

This catalog is the source of truth. New actions are added by extending the catalog and the corresponding permission string.

## Audit Log Retention (Mandatory)

- All audit logs are retained for **7 years** (regulatory requirement for financial actions; standard for user actions)
- Logs are stored in a write-only table (no UPDATE, no DELETE grants)
- Logs are replicated to a cold-storage archive after 90 days for disaster recovery
- Logs are searchable via the admin panel with full-text and structured filters
- Logs are exportable as CSV and JSON for compliance audits

## Audit Log Search (Mandatory)

Admins with `audit:read` permission can search the audit log with:

- Date range
- Actor (user or role)
- Action (with `*` wildcard)
- Target type and ID
- IP address
- Metadata fields (with JSON path queries)

The search results are paginated (50 per page) and can be exported.

## Log Integrity (Mandatory)

To detect tampering, each `AuditLog` row's `contentHash` is computed from the entire row's JSON. The hashes are:

- Stored alongside the log entry
- Periodically verified against the stored hash
- A mismatch is a Sentry alert at `fatal` level and pages the on-call

Hash chain: each log entry's hash includes the previous entry's hash, forming a Merkle-like chain. A break in the chain indicates tampering.

---

# Moderation (Mandatory)

The moderation subsystem handles user reports, content review, and platform safety.

## User Reports (Mandatory)

A user MAY report another user, post, comment, message, or Course for review. A `Report` record is created:

```typescript
interface Report {
  id: string;
  reporterId: string;             // User who submitted the report
  reportedType: 'USER' | 'POST' | 'COMMENT' | 'MESSAGE' | 'Course' | 'COMMUNITY' | 'TEST' | 'LESSON';
  reportedId: string;
  reason: 'SPAM' | 'HARASSMENT' | 'HATE_SPEECH' | 'SEXUAL_CONTENT' | 'VIOLENCE' | 'MISINFORMATION' | 'COPYRIGHT' | 'IMPERSONATION' | 'CHEATING' | 'OTHER';
  description?: string;           // Free-text explanation
  evidence?: string[];            // URLs, screenshots
  status: 'PENDING' | 'IN_REVIEW' | 'RESOLVED' | 'DISMISSED';
  assignedTo?: string;            // Moderator userId
  resolution?: string;            // Moderator's resolution
  resolvedBy?: string;
  resolvedAt?: string;
  createdAt: string;
}
```

## Report Queue (Mandatory)

Reports are routed to a moderation queue. The queue is sorted by:

1. Severity (HARASSMENT, HATE_SPEECH, VIOLENCE > SPAM, COPYRIGHT)
2. Age (older first)
3. Reporter trust score (higher trust first)

A moderator with `moderation:review` permission can claim a report (assigning it to themselves). The claim is exclusive; another moderator cannot claim it.

## Moderation Actions (Mandatory)

A moderator's resolution is one of:

| Action | Description | When |
|---|---|---|
| `NO_ACTION` | Report is invalid | Report was a mistake or not actionable |
| `WARNING` | Issue a warning to the offender | First offense, minor |
| `CONTENT_REMOVAL` | Remove the offending content | Content violates policy |
| `SHADOW_REMOVAL` | Hide the content (only the author sees) | Borderline content |
| `TEMPORARY_BAN` | Ban the user for 1/7/30 days | Repeat offense |
| `PERMANENT_BAN` | Permanently ban the user | Severe or repeated offense |
| `ESCALATE` | Escalate to a senior moderator or admin | Outside moderator's authority |

Each action produces an `AuditLog` entry and a notification to the affected user (with an appeal link, where applicable).

## Moderation SLA (Mandatory)

| Severity | First Response | Resolution Target |
|---|---|---|
| `CRITICAL` (violence, CSAM, imminent harm) | 1 hour | 4 hours |
| `HIGH` (harassment, hate speech) | 4 hours | 24 hours |
| `MEDIUM` (spam, copyright) | 24 hours | 72 hours |
| `LOW` (other) | 72 hours | 7 days |

A breach of the SLA is logged to Sentry and reviewed weekly.

## AI-Assisted Moderation (Optional)

For high-volume categories (e.g., spam in communities), an AI model can pre-classify reports. The AI's classification is shown to the moderator, but the moderator makes the final decision. The AI's accuracy is measured weekly against moderator decisions; an accuracy drop below 80% triggers a re-training or replacement.

## Auto-Moderation (Mandatory for Specific Categories)

For specific violation categories, the system MAY auto-act:

- **Spam**: if a user posts > 5 identical or near-identical messages in 1 hour, the posts are auto-removed and the user is warned
- **Profanity**: if a post contains a known profanity word and the user has been warned ≥ 3 times, the post is auto-removed
- **Link spam**: if a post contains > 3 links and the user has < 7 days tenure, the post is auto-flagged for review

Auto-moderation actions are logged in `AuditLog` with `actorId = 'SYSTEM'` and the `actorRole = 'SYSTEM'`.

---

# No Hard Delete (Mandatory)

Hard delete is forbidden. All entities that can be referenced (e.g., a post that has comments) are soft-deleted.

## Soft Delete Pattern (Mandatory)

Every soft-deletable model has:

- `deletedAt: DateTime?` (null if not deleted)
- `deletedBy: String?` (userId of the deleter, or `'SYSTEM'`)
- `deletedReason: String?` (optional, e.g., "User account deleted")

The Prisma middleware automatically adds a `where: { deletedAt: null }` filter on every query unless the caller explicitly opts out (e.g., an admin viewing the trash).

## Soft Delete Behavior (Mandatory)

When an entity is soft-deleted:

1. The row is updated with `deletedAt = now()`, `deletedBy = <actorId>`, `deletedReason = <reason>`
2. An `AuditLog` entry is created
3. The entity disappears from public queries
4. The entity remains in the database for the retention period (default: 90 days for non-financial, 7 years for financial)
5. After the retention period, a cron job permanently deletes the row

## Cascade Soft Delete (Mandatory)

If a parent entity is soft-deleted, all children are also soft-deleted (in the same transaction). For example, deleting a `Community` also deletes its `Posts`, which deletes their `Comments`. The deletion cascade is logged in the parent's audit log entry metadata.

## Restoration (Mandatory)

A soft-deleted entity can be restored within the retention period by an admin with the corresponding `restore` permission. Restoration:

- Sets `deletedAt = null`, `deletedBy = null`, `deletedReason = null`
- Creates an `AuditLog` entry (`*.restore`)
- Re-publishes the entity (e.g., a restored post is visible again)

Restoration of deeply nested entities is supported (e.g., restoring a deleted community also restores its posts).

## Hard Delete Exception (Mandatory)

Hard delete is allowed ONLY for:

1. **GDPR / privacy erasure**: a user requests data deletion under GDPR right-to-be-forgotten. Hard delete is performed by a separate, audited process.
2. **Test data**: data marked as `isTestData: true` at creation. Hard delete is allowed via a script, not the admin panel.
3. **PII accidental capture**: if PII is accidentally stored (e.g., a user submitted their password in a comment), the specific field is hard-deleted (NOT the entire row). The action is audited.

Hard delete is NEVER performed by a regular admin action. Even Super Admin cannot perform a hard delete through the admin panel.

---

# Role-Based Access Control (Mandatory)

RBAC is the foundation of admin security. Every admin endpoint checks the actor's role and permissions.

## Roles (Mandatory)

| Role | Description | Default Permissions |
|---|---|---|
| `SUPER_ADMIN` | Full platform access | All permissions |
| `ADMIN` | Most platform operations | Excludes ownership transfer, AI model enable/disable, role management |
| `MODERATOR` | Community and content moderation | `community:moderate`, `moderation:review`, `audit:read` |
| `SUPPORT` | Read user info, basic support actions | `user:read`, `audit:read` (limited) |
| `FINANCE` | Payment, refund, payout operations | `payment:refund`, `payout:approve`, `audit:read` (limited) |
| `AI_ENGINEER` | AI model and prompt management | `ai:prompt`, `ai:model`, `ai:reindex`, `audit:read` (limited) |
| `CONTENT_REVIEWER` | Read-only access to content | `content:read`, `moderation:review` |

A user has exactly one primary role. They MAY have additional roles (e.g., a user is both `ADMIN` and `FINANCE`).

## Permission String Format (Mandatory)

Permissions are strings in the format `<domain>:<action>`. The `*` wildcard matches all actions in a domain. The `*:*` wildcard matches everything (Super Admin only).

Examples:

- `user:read` — read user data
- `user:update` — update user data
- `community:*` — all community actions
- `*:*` — everything (Super Admin only)

## Permission Checks (Mandatory)

Every protected endpoint MUST have a permission check. The check is via a `@RequirePermission('user:update')` decorator or a `requirePermission('user:update')` guard. A check MUST happen at the route level (not inside the service) to ensure it is not bypassed.

## Permission Inheritance (Forbidden)

Permissions do not inherit. A user with `user:read` does NOT automatically have `user:update`. This is a deliberate design choice to make permissions explicit and auditable.

## Role Templates (Mandatory)

A "role template" is a named set of permissions (e.g., the `MODERATOR` template). When a role is created, it is based on a template. The template cannot be modified after creation (a new template must be created instead). This ensures that the permissions of all `MODERATOR`s are consistent.

## Permission Audit (Mandatory)

Every role assignment and permission change is logged. The full permission set of a user is reconstructable from the audit log.

## Permission Denied Response (Mandatory)

A denied permission returns:

```json
{
  "error": "FORBIDDEN",
  "code": "INSUFFICIENT_PERMISSIONS",
  "message": "You do not have the required permission: user:delete",
  "requiredPermission": "user:delete"
}
```

The response does NOT leak which permissions the user DOES have. This prevents attackers from enumerating the permission system.

---

# MFA and Session Security (Mandatory)

Admin access requires Multi-Factor Authentication (MFA). The first admin to log in to a fresh deployment is forced to set up MFA.

## MFA Setup (Mandatory)

The admin sets up MFA via:

1. TOTP (Google Authenticator, Authy, 1Password, etc.) — the default
2. WebAuthn (hardware key, e.g., YubiKey) — recommended for SUPER_ADMIN
3. SMS (deprecated; only as a fallback)

MFA setup requires verifying the chosen method before it is enabled.

## MFA Enforcement (Mandatory)

| Role | MFA Required |
|---|---|
| `SUPER_ADMIN` | Yes, always. WebAuthn strongly recommended. |
| `ADMIN` | Yes, always. |
| `MODERATOR` | Yes, always. |
| `SUPPORT` | Yes, always. |
| `FINANCE` | Yes, always. |
| `AI_ENGINEER` | Yes, always. |
| `CONTENT_REVIEWER` | Yes (can use TOTP) |

A user without MFA cannot access the admin panel. The admin panel route returns `403 MFA_REQUIRED` and redirects to the MFA setup flow.

## Session Management (Mandatory)

- Admin sessions expire after **15 minutes of inactivity** (down from the standard 24 hours for users)
- Admin sessions are bound to the IP and User-Agent
- An IP or User-Agent change requires re-authentication (MFA prompt)
- A user has at most **3 active admin sessions** at a time (newest login invalidates older)

## Session Activity Log (Mandatory)

Every admin session action (page load, API call) is logged with:

- `sessionId`
- `actorId`
- `action` (page route or API endpoint)
- `ipAddress`
- `userAgent`
- `timestamp`

The session activity log is searchable by the user themselves (to review their own activity) and by Super Admin (to review any session).

## Forced Logout (Mandatory)

Super Admin can force-logout any admin session. The forced logout is logged and the affected admin is notified by email.

## Concurrent Session Detection

If the same admin logs in from ≥ 2 distinct IPs within 5 minutes, both sessions are flagged. The admin receives a "New login from <location>" alert. The existing session continues, but the alert is shown in the admin panel for verification.

---

# Impersonation (Mandatory)

An admin MAY impersonate another user for support purposes. Impersonation is heavily restricted and audited.

## Impersonation Rules (Mandatory)

| Rule | Description |
|---|---|
| Who can impersonate | Users with `user:impersonate` permission. Default: `SUPER_ADMIN`, `ADMIN`, `SUPPORT`. |
| Who can be impersonated | Any non-admin user. Admins cannot be impersonated. |
| Duration | Maximum 1 hour per impersonation session. Re-authentication (MFA) is required to extend. |
| Action restrictions | The impersonator cannot: change the user's password, change the user's email, change the user's MFA, perform a payment, change the user's role, delete the user's account. Other actions are logged and allowed. |
| Visual indicator | The impersonated session shows a persistent "You are being impersonated by <admin>" banner. The admin's UI shows a persistent "Impersonating <user>" banner. |
| Notification | The impersonated user is notified by email within 5 minutes of the impersonation starting. |
| Audit | Every action taken during impersonation is logged in `AuditLog` with `impersonatedUserId` set. |

## Impersonation Start (Mandatory)

To start impersonation, the admin must:

1. Have `user:impersonate` permission
2. Provide a written reason (≥ 20 characters) explaining why
3. Re-authenticate (MFA prompt)
4. Confirm the action in a confirmation dialog

The action is logged. The user is notified. The impersonation session is created with a 1-hour TTL.

## Impersonation End (Mandatory)

Impersonation ends:

1. When the admin clicks "End impersonation"
2. When the 1-hour TTL expires
3. When the admin's session expires (15-minute inactivity)
4. When the admin navigates to an impersonation-restricted action (e.g., password change)
5. When an emergency "Stop all impersonations" button is pressed (Super Admin only)

The end is logged. The user is notified that impersonation has ended.

## Impersonation Audit (Mandatory)

A dedicated `ImpersonationLog` table stores:

- `id`
- `adminId`
- `adminEmail`
- `userId` (impersonated)
- `userEmail`
- `reason`
- `startedAt`
- `endedAt?`
- `actionsTaken: AuditLog[]` (the actions performed during impersonation)
- `ipAddress`
- `userAgent`

A user can view their own `ImpersonationLog` history at any time.

## Impersonation Anti-Patterns (Forbidden)

- Impersonating an admin
- Impersonating without a written reason
- Impersonating without MFA re-authentication
- Impersonating for > 1 hour without explicit re-authorization
- Performing restricted actions during impersonation
- Bypassing the user notification

---

# Analytics and Dashboards (Mandatory)

The admin panel exposes platform-wide metrics. The metrics are surfaced via Grafana dashboards and via the admin panel's analytics page.

## Required Dashboards (Mandatory)

| Dashboard | Audience | Cadence |
|---|---|---|
| Platform Overview | All admins | Real-time |
| User Growth & Retention | All admins | Daily |
| Revenue & Payments | FINANCE, SUPER_ADMIN | Real-time |
| Creator Economy | FINANCE, SUPER_ADMIN | Daily |
| AI Costs & Usage | AI_ENGINEER, SUPER_ADMIN | Real-time |
| Live Exam Health | Live exam owners, SUPER_ADMIN | Real-time |
| Moderation Queue | MODERATOR, SUPER_ADMIN | Real-time |
| Audit Log Search | All admins with `audit:read` | Real-time |

## Metric Definitions (Mandatory)

All metrics are computed from first principles. The metric definition is in code; the dashboard reads the code. There is no "magic" or undocumented metric.

| Metric | Definition |
|---|---|
| DAU (Daily Active Users) | Distinct users with ≥ 1 session in the last 24 hours, ending at the current time |
| WAU (Weekly Active Users) | Distinct users with ≥ 1 session in the last 7 days |
| MAU (Monthly Active Users) | Distinct users with ≥ 1 session in the last 30 days |
| Retention (D1) | % of new users who return on day 2 |
| Retention (D7) | % of new users who return on day 8 |
| Retention (D30) | % of new users who return on day 31 |
| LTV (Lifetime Value) | Sum of revenue from a user cohort, divided by cohort size |
| CAC (Customer Acquisition Cost) | Marketing spend / new users acquired in the period |
| ARPU (Average Revenue Per User) | Total revenue / MAU |
| Churn (Monthly) | % of users active in month N-1 who are not active in month N |

A metric MUST be defined in code with a docstring explaining its computation. Dashboard readers link to the docstring.

## PII in Analytics (Mandatory)

Analytics dashboards do NOT display PII (email, phone, name). If a metric is "top 10 spenders," the user is shown as "User 12345" with a deep link to the user profile (which requires permission to view).

## Metric Backfilling (Mandatory)

When a metric's definition changes, the historical data is backfilled. A backfill is a separate, audited process. The old metric values are preserved; the new metric values are stored in a separate table or with a `version` field.

---

# AI Monitoring (Mandatory)

Admins with `ai:read` permission can monitor:

- AI usage by user, by use case, by model
- AI costs (by provider, by model, by use case)
- AI errors (rate, types, recovery)
- Provider health (latency, error rate, quota)

## AI Cost Dashboard (Mandatory)

A real-time dashboard shows:

- Spend by provider (last 1 hour, 24 hours, 7 days, 30 days)
- Spend by model
- Spend by use case (chat, MCQ explanation, doubt solver, etc.)
- Spend per user (top 100)
- Forecasted spend for the current month (based on rate of use)

A spend spike (≥ 50% above the 7-day moving average) is alerted via Sentry and email to AI_ENGINEER and SUPER_ADMIN.

## AI Quality Monitoring (Mandatory)

For each RAG use case, the dashboard shows:

- Faithfulness score (last 24h, 7d, 30d)
- Hallucination rate
- Average retrieval recall
- User feedback (thumbs up/down ratio)
- User-flagged hallucinations

A regression (≥ 5% drop in faithfulness or ≥ 2% increase in hallucination rate) blocks AI model changes and triggers an incident.

## Provider Quota Monitoring (Mandatory)

Each provider has a quota (spending limit, rate limit, token limit). The dashboard shows:

- Current usage vs quota
- Quota reset time
- Forecasted quota exhaustion time

A quota at 80% triggers an alert. A quota at 95% blocks new requests (with a fallback).

## Prompt and Model Change Audit (Mandatory)

Every `ai.prompt.deploy`, `ai.model.enable`, and `ai.embedding.reindex.start` action is logged in `AuditLog`. The audit log includes the change's diff (before/after prompt text, before/after model config).

---

# Compatibility With Other Standards

This document defers to:

- [security-rules.md](docs/standards/security-rules.md) for auth, MFA, and rate limiting
- [architecture-rules.md](docs/standards/architecture-rules.md) for module boundaries and the ADR requirement
- [backend-rules.md](docs/standards/backend-rules.md) for service patterns and Sentry instrumentation
- [Course-rules.md](docs/standards/Course-rules.md) for Course moderation
- [creator-economy-rules.md](docs/standards/creator-economy-rules.md) for creator-related moderation (suspension, payouts, etc.)

---

# Sprint 1 Compliance Checklist

Since the admin panel is partially implemented, this section will be expanded in v1.0. For Sprint 1:

- [ ] ADR written for: RBAC library (CASL vs custom), MFA library (Auth0 vs custom), audit log storage (PostgreSQL vs append-only log service)
- [ ] `AuditLog` table created with append-only enforcement at the DB level
- [ ] `Role`, `Permission`, and `RolePermission` tables created
- [ ] `@RequirePermission` decorator implemented
- [ ] MFA setup flow with TOTP and WebAuthn
- [ ] Admin session management with 15-minute inactivity timeout
- [ ] Impersonation flow with all restrictions and audit
- [ ] Moderation queue with claim/exclusivity
- [ ] Auto-moderation rules for spam and profanity
- [ ] Grafana dashboard for platform overview
- [ ] AI cost dashboard with quota monitoring

---

# Final Directive

The admin panel is the highest-trust surface in GK Circle.

A compromised admin is a compromised platform.

A missing audit log is a missing legal record.

A hard delete is a lost truth.

Build admin that is fully audited, MFA-protected, RBAC-enforced, and impersonation-restricted.

Verify all four.
