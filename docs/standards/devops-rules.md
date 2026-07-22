# GK Circle DevOps Rules

Version: 1.0

Status: Mandatory

---

# Infrastructure

Docker First

CI/CD First

Infrastructure As Code

---

# Deployment Pipeline

Lint

↓

Tests

↓

Security Scan

↓

Build

↓

Deploy

↓

Verify

---

# Monitoring

Sentry

Prometheus

Grafana

OpenTelemetry

---

# Backup Rule

Daily Backups

Weekly Recovery Test

Monthly DR Drill

---

# Production Rule

No direct production modifications.

Everything goes through CI/CD.

---

# Cost Monitoring Rule

Monitor:

AI Costs

Hosting Costs

Storage Costs

Bandwidth Costs

Email Costs

---

# v1.0 Addendum — Architecture-Intent Standard

This addendum extends the v0.5 baseline above with architecture-intent depth. The v0.5 rules above remain in force. This addendum provides the environment matrix, migration policy, secret management forbid list, and cost controls that the v0.5 rules reference.

## Dependency Rule (AGENTS.md Wins)

If a rule in this file conflicts with [AGENTS.md](../../AGENTS.md), AGENTS.md wins. If a rule in this file conflicts with another standards file, the more-specific file wins (security secrets live in [security-rules.md](security-rules.md); backup procedures live in [operations-rules.md](operations-rules.md); cost alerts live in [operations-rules.md](operations-rules.md) §"Alerting"). If both files are at equal specificity, the more recently updated file wins. Record the decision in [../active-context/decisions.md](../active-context/decisions.md).

## Verification Requirement (Existence ≠ Evidence)

A pipeline is not "done" because the YAML exists. It is done when:
- The pipeline has been run end-to-end on a real PR and the artifacts (build, image, deploy) have been verified
- The rollback procedure has been tested in staging (the rollback ran, the previous version came back, smoke tests passed)
- The migration has been applied to a copy of production and the data integrity check passed
- The secret has been rotated at least once and the rotation procedure has been documented

Existence is the start. Evidence is the finish.

---

# Environment Matrix (Mandatory)

GK Circle operates four environments. They are isolated, gated, and serve distinct purposes. A deploy to a higher environment does not happen until the lower environment has been verified.

| Environment | Purpose | Data | Who can deploy | Access |
|---|---|---|---|---|
| **local** | Developer machine. Single developer. | Synthetic seed data only. No production data. | The developer | Developer only |
| **dev** (a.k.a. development) | Shared engineering environment. Auto-deploy from `main` branch. | Synthetic seed data + anonymized subset of staging. NO real user data, NO production PII. | Engineering team (auto via CI) | Engineering team |
| **staging** | Pre-production. Mirrors production architecture (same managed services, same scaling, same config except sizes). | Anonymized production snapshot (last 7 days), refreshed weekly. NO PII. | Release captain (manual gate) | Engineering + QA + Course |
| **production** (a.k.a. prod) | Live system. Real users. Real money. | The real data. | Release captain (manual gate + change advisory board for high-risk) | On-call rotation + SRE |

## What Is Allowed In Each Environment

| Action | local | dev | staging | production |
|---|---|---|---|---|
| Run uncommitted code | ✅ | ⚠️ via feature branch only | ❌ | ❌ |
| Seed synthetic data | ✅ | ✅ | ❌ | ❌ |
| Run destructive DB migration | ✅ (on a copy) | ✅ (on a copy) | ⚠️ with explicit approval + backup before + rollback plan | ❌ without ADR + backup + rollback + change advisory board |
| Use real user data | ❌ | ❌ | ❌ (anonymized snapshot only) | ✅ (the only place it lives) |
| Connect to production DB | ❌ | ❌ | ⚠️ read-only replicas only, for debugging | ✅ |
| Bypass CI | ❌ | ❌ | ❌ | ❌ |
| Skip security scan | ⚠️ (with `// SECURITY:` comment) | ❌ | ❌ | ❌ |
| Deploy on a Friday | ✅ | ✅ | ❌ (no staging deploys Fri 16:00+ IST) | ❌ (no prod deploys Fri 18:00+ IST, weekends, or Indian public holidays without VP Eng sign-off) |
| Use production secrets | ❌ | ❌ | ⚠️ staging-mirrored secrets only (separate KMS namespace) | ✅ |
| Generate real payments | ❌ | ❌ | ⚠️ via Razorpay test mode only | ✅ |
| Run load test | ✅ (against local infra) | ✅ (against dev infra, < 100 RPS) | ✅ (against staging, < 5,000 RPS, scheduled) | ⚠️ against production only with VP Eng + SRE lead sign-off, never during business hours |

## Environment Promotion Gate

A change moves local → dev → staging → production **only** when:

1. **local → dev**: the developer's PR has CI green (lint, type, unit, integration, security scan)
2. **dev → staging**: the release captain has reviewed the change set, the staging deploy succeeded, and the smoke tests pass
3. **staging → production**: the change has been in staging for ≥ 24 hours with no regressions, the rollback procedure has been tested in staging, the change advisory board (for high-risk) has signed off, and the deploy window is not a Friday evening, weekend, or holiday

No environment skipping. A change tested only in dev does not go to production.

---

# Migration Policy (Mandatory)

Database migrations are the most dangerous production change. They are governed by this policy.

## Additive First (Mandatory)

The default migration is **additive only**: add a column with a default, add a table, add an index, add a new nullable column. Additive migrations are safe to deploy without coordination and can be rolled back trivially.

## Destructive Migration Requires an ADR (Mandatory)

A destructive migration is:
- `DROP COLUMN`
- `DROP TABLE`
- `ALTER COLUMN ... TYPE` (type change that requires a rewrite)
- `RENAME COLUMN` or `RENAME TABLE`
- Any migration that locks the table for > 1 second on a table with > 1M rows
- Any migration that backfills > 1M rows in a single transaction

A destructive migration **MUST** have an ADR that documents:
- The reason the change is necessary (with the customer or customer need)
- The alternatives considered (e.g., a feature flag, a parallel column, a view)
- The estimated lock time and backfill time
- The rollback plan (what happens if the migration fails mid-flight, or fails in production days later)
- The blast radius (which users are affected, what is the user-visible impact if it goes wrong)
- The communication plan (when users are notified, what they are told)

The ADR is reviewed and approved by the VP Engineering or delegate **before** the migration is merged.

## Backup Before Destructive Migration (Mandatory)

Before any destructive migration runs in **any** environment above `local`:
1. A backup is taken (full `pg_dump` + WAL position marker)
2. The backup is verified to be restorable (a smoke test is run on the backup)
3. The backup retention is extended (the backup is preserved for at least 90 days post-migration, regardless of normal retention)
4. The backup is in a separate, isolated location (not the same volume as the source DB)

If the migration takes longer than 30 minutes, an incremental backup is taken every 30 minutes.

## Rollback Plan Mandatory (Mandatory)

Every migration has a written rollback plan that includes:
- The exact commands to reverse the migration (e.g., `ALTER TABLE ... ADD COLUMN ...`)
- The data preservation strategy (if the migration renames a column and code has not yet caught up, how do we ensure no data is lost in the window?)
- The maximum acceptable rollback time (a migration that takes 10 minutes to apply and 4 hours to roll back is a problem; it should be split)
- The trigger for executing the rollback (a query latency spike? an error rate? a manual call?)
- The person authorized to call the rollback (the on-call lead, the release captain, the VP Eng — specified in the ADR)

A migration without a tested rollback plan is not approved.

## Zero-Downtime Migrations (Mandatory for Tier 1)

Tier 1 services (auth, live-exam, payment, billing) require zero-downtime migrations. The standard pattern is the **expand-migrate-contract** trilogy:

1. **Expand**: add the new column (nullable, with default), deploy code that writes to both old and new
2. **Migrate**: backfill the new column from the old in batches (no lock)
3. **Contract**: once backfill is complete, deploy code that reads only the new column, then drop the old column in a follow-up migration (with its own ADR)

Skipping expand-migrate-contract and going directly to "rename the column" in a single deploy is a Phase 0.3 integrity violation.

## Migration Checklist (Mandatory)

Before any migration merges:

- [ ] Migration has been applied to a copy of production (size and data shape match)
- [ ] Rollback plan has been executed at least once in staging
- [ ] ADR exists (for destructive migrations)
- [ ] Backup has been taken and verified
- [ ] Migration has been applied to staging and the staging smoke tests pass
- [ ] Migration is idempotent (running it twice produces the same schema)
- [ ] Migration is reversible (or ADR documents why irreversibility is acceptable)
- [ ] Migration has been reviewed by a second engineer (not the author)
- [ ] Estimated lock time is documented
- [ ] For Tier 1: expand-migrate-contract pattern is used

---

# Secret Management — Forbid List (Mandatory)

Secrets are governed by a strict forbid list. The list is enforced by pre-commit hooks (gitleaks, trufflehog), CI scanners, and a quarterly manual audit. A violation of the forbid list is a security incident.

## Forbidden in Source Control

Hard forbid. A pre-commit hook blocks the commit. A CI scanner blocks the PR. A retrospective finding pages the security team.

- **API keys** (any provider): `apiKey`, `api_key`, `apikey`, `API_KEY=...`
- **Private keys** (any crypto): `BEGIN PRIVATE KEY`, `BEGIN RSA PRIVATE KEY`, `BEGIN OPENSSH PRIVATE KEY`, `BEGIN EC PRIVATE KEY`
- **Database connection strings** with credentials: `postgres://user:pass@...`, `mysql://user:pass@...`, `mongodb://user:pass@...`, `redis://user:pass@...`
- **OAuth client secrets**: `clientSecret`, `client_secret`
- **JWT secrets**: `JWT_SECRET`, `jwtSecret`, `jwt_secret`
- **Session secrets**: `SESSION_SECRET`, `sessionSecret`
- **Encryption keys**: `ENCRYPTION_KEY`, `encryptionKey`, `kmsKeyId`
- **Webhook secrets**: `webhookSecret`, `webhook_secret`
- **Payment gateway secrets**: Razorpay key secret, Cashfree secret, Stripe secret key
- **Cloud provider credentials**: AWS access key + secret, GCP service account JSON, Azure connection string
- **Email service credentials**: SMTP password, SendGrid API key, Postmark token
- **AI provider keys**: OpenAI API key, Anthropic API key, Google AI API key, Pinecone API key, Qdrant API key
- **OAuth tokens**: `access_token`, `refresh_token`, `id_token`
- **Personal access tokens** (GitHub, GitLab, npm, PyPI): `ghp_...`, `glpat-...`, `npm_...`

## Forbidden in Source Code (Even With Source Control Ignored)

A secret in `*.local`, `*.example`, `*.template` files is still a secret if it is real. Use the placeholder pattern:

```bash
# ✅ Allowed — placeholder
RAZORPAY_KEY_ID=rzp_test_REPLACE_ME
JWT_SECRET=changeme-generate-a-real-one
DATABASE_URL=postgres://user:password@localhost:5432/dbname

# ❌ Forbidden — real value
RAZORPAY_KEY_ID=rzp_live_AbCdEf123456
JWT_SECRET=super-secret-jwt-signing-key-do-not-use
DATABASE_URL=postgres://admin:s3cret@prod-db.ap-south-1.rds.amazonaws.com:5432/gkcircle
```

The placeholder is replaced at deploy time from the secret manager.

## Forbidden in Screenshots

A screenshot in a Jira ticket, a Slack message, a Notion page, a Playwright artifact, a Sentry breadcrumb, a log line, a CI artifact, a test fixture, or a documentation page must NOT contain:
- Browser DevTools showing a real `Authorization` header
- Terminal output showing a real env var
- Database client (pgAdmin, DataGrip) showing real data
- Cloud console (AWS, GCP, Razorpay dashboard) showing real credentials
- API response showing a real user record with full PII

The rule: if a screenshot is going outside the engineer's local machine, every value must be redacted. Use `***REDACTED***` overlays or fake values.

## Forbidden in Documentation

Documentation pages (including this one) and code examples must use placeholders, not real values. A `code` block showing `RAZORPAY_KEY_SECRET=your_key_secret_here` is acceptable; a `code` block showing `RAZORPAY_KEY_SECRET=Hg7...xyz` is not, even if the value is from a test gateway.

## Required Storage

| Secret class | Storage |
|---|---|
| Production secrets (DB, payment, AI, cloud) | AWS Secrets Manager (or HashiCorp Vault) — accessed at runtime via IAM role, never written to disk |
| Staging secrets | AWS Secrets Manager (separate KMS namespace) — accessed at runtime via IAM role |
| CI/CD secrets | GitHub Actions Secrets — accessed in workflow via `secrets.*` context, masked in logs |
| Local dev secrets | `.env.local` (gitignored) + `direnv` for auto-loading |
| Test fixtures | Faker-generated values, never real secrets |

## Secret Rotation (Mandatory)

| Secret class | Rotation cadence | Owner |
|---|---|---|
| Database password | 90 days | Platform |
| JWT signing key | 30 days | Security |
| API keys (third-party) | 180 days OR on any personnel change | Security |
| Cloud IAM keys | 90 days, prefer IAM roles | Platform |
| Webhook secrets | 365 days OR on any personnel change | Security |
| Encryption keys (KMS) | 365 days (key rotation) | Security |

Every rotation is logged in the AuditLog (per [admin-panel-rules.md](admin-panel-rules.md)).

---

# Cost Controls (Mandatory)

GK Circle's cost surface is non-trivial: AI tokens (the largest single line item), RAG vector storage, live exam infrastructure, video/media storage, and bandwidth. Without active cost controls, the platform can become unprofitable at scale. This section is the contract that prevents that.

## Cost Monitoring Domains

| Domain | What is measured | Tool | Alert threshold |
|---|---|---|---|
| **AI tokens** | Input/output tokens per model per day, in ₹ | OpenAI/Anthropic usage API → ClickHouse → Grafana | > 110% of 7-day rolling average for 3 consecutive days |
| **AI provider cost** | Per-query cost in ₹, p50/p95/p99 | OpenAI/Anthropic billing → Grafana | Daily cost > ₹50,000 OR p99 cost > ₹5/query |
| **RAG vector storage** | Qdrant storage in GB, embeddings count | Qdrant metrics → Grafana | > 5 TB OR growth > 10% week-over-week |
| **RAG query cost** | Per-query cost in ₹ (embedding + LLM) | Custom instrumentation → ClickHouse | > ₹2/query p50 |
| **Live exam infrastructure** | Concurrent sessions, compute, bandwidth | Prometheus → Grafana | > 80% of provisioned capacity for > 5 minutes |
| **Video / media storage** | S3 storage in GB, transfer out GB | AWS billing → Grafana | > 5 TB OR transfer > ₹50,000/month |
| **Object storage (general)** | S3 total storage, request count, transfer | AWS billing → Grafana | > 10 TB OR > 100M requests/month |
| **Database** | RDS instance hours, storage, IOPS, backup storage | AWS billing → Grafana | > 80% of provisioned IOPS OR > 5 TB |
| **Bandwidth (CDN)** | CloudFront transfer out, request count | AWS billing → Grafana | > ₹100,000/month |
| **Email** | SES send count, bounce rate, complaint rate | SES metrics → Grafana | Bounce > 5% OR complaint > 0.1% (sender reputation risk) |
| **SMS / OTP** | Provider send count, delivery rate | Provider API → Grafana | Delivery < 95% |
| **Third-party services** | Per-vendor monthly spend | Vendor billing → ClickHouse | > ₹25,000/month per vendor OR new vendor > ₹10,000/month |

## Budget Alerts (Mandatory)

| Threshold | Action |
|---|---|
| **80% of monthly budget consumed** | Slack alert to #gkc-eng and #gkc-finance. Weekly review of the cost dashboard. |
| **100% of monthly budget consumed** | Page the on-call + finance lead. AI services throttle to free tier (graceful degradation per [live-exam-rules.md](live-exam-rules.md) §"Graceful Degradation" pattern). |
| **120% of monthly budget consumed** | SEV-2 incident. Engineering Director and CFO sign-off required to continue spending. |
| **150% of monthly budget consumed** | SEV-1 incident. CEO + CFO + CTO in a war room. Hard cut on non-essential features. |

Monthly budgets are set quarterly by Engineering + Finance and signed off by the CFO. The budget is broken down by:
- AI (largest)
- Infrastructure (compute + storage + bandwidth)
- Third-party SaaS (auth, monitoring, email, SMS, payment, etc.)
- Cloud overhead (egress, snapshots, etc.)

## Usage Quotas (Mandatory)

To prevent a single user or a single creator from consuming the entire AI budget:

| Tier | Daily AI tokens | Daily RAG queries | Daily video uploads | Daily storage writes |
|---|---|---|---|---|
| Free user | 50,000 | 100 | 5 | 100 MB |
| Pro user | 500,000 | 1,000 | 50 | 1 GB |
| Creator | 5,000,000 | 10,000 | 500 | 10 GB |
| Enterprise / Test partner | Custom (set per contract) | Custom | Custom | Custom |

A user exceeding the daily quota gets a 429 with a `Retry-After` header. A user exceeding the daily quota 7 days in a row is flagged for review by the growth team. Quotas are not silent — the user is told they hit the limit and is offered an upgrade path.

## AI Token Monitoring (Mandatory)

Because AI is the largest cost line, it gets its own monitoring depth:

| Metric | What it tells us |
|---|---|
| `gkc_ai_tokens_input_total{model}` | Input token volume by model |
| `gkc_ai_tokens_output_total{model}` | Output token volume by model |
| `gkc_ai_cost_inr_total{model,user_tier}` | Actual cost in ₹ |
| `gkc_ai_cost_inr_per_query{model,user_tier}` | Per-query cost distribution |
| `gkc_ai_cache_hit_ratio` | Prompt cache hit rate (target: > 30% for RAG) |
| `gkc_ai_prompt_version_total{version}` | Cost broken down by prompt version (so we can see if a new prompt is more expensive) |
| `gkc_ai_embedding_version_total{version}` | Cost broken down by embedding version |
| `gkc_ai_faithfulness_score{model,prompt_version}` | Quality vs cost tradeoff (per [rag-rules.md](rag-rules.md) §"Faithfulness Verifier") |

Every prompt change is A/B tested for cost before full rollout. A prompt that costs 2× more is adopted only if it improves faithfulness by ≥ 5 percentage points (or has a documented quality reason).

## Storage Growth Monitoring (Mandatory)

| Metric | What it tells us | Alert |
|---|---|---|
| `gkc_storage_bytes_total{kind}` | Storage by kind (video, image, qdrant, postgres, clickhouse) | > 80% of provisioned OR growth > 10% week-over-week |
| `gkc_storage_bytes_per_user_average` | Per-user average storage | > 500 MB/user for free tier |
| `gkc_storage_bytes_per_creator_average` | Per-creator average storage | > 5 GB/creator |
| `gkc_storage_lifecycle_transition_total{kind,from,to}` | Lifecycle transitions (S3 → IA → Glacier → expired) | < 90% of objects transitioned as planned |
| `gkc_storage_orphaned_bytes_total` | Objects not referenced by any DB row | > 1% of total storage |

A weekly report lists the top 20 users by storage and the top 20 creators by storage. The growth team reviews for any outlier that needs outreach.

---

# CI/CD Pipeline (Mandatory)

The deploy pipeline is identical across services. Stages:

1. **Lint** — ESLint, Prettier, Ruff, etc. (project-specific)
2. **Type check** — TypeScript, mypy
3. **Unit tests** — must pass with ≥ 80% coverage on changed files
4. **Integration tests** — must pass
5. **Security scan** — `npm audit`, `pip-audit`, `gitleaks`, `trufflehog`, Snyk
6. **Build** — Docker image, signed with cosign, pushed to OCI registry
7. **Deploy to dev** — auto on merge to `main`
8. **Smoke test in dev** — auto
9. **Deploy to staging** — manual gate by release captain
10. **Smoke test in staging** — auto, must pass for 24 hours
11. **Deploy to production** — manual gate by release captain
12. **Smoke test in production** — auto, canary 5% → 25% → 100% over 30 minutes
13. **Post-deploy verification** — error rate, latency, business metrics, SLO budget consumption

If any stage fails, the pipeline halts. A failed prod deploy auto-rolls back via the pipeline.

## Rollback Strategy (Mandatory)

Every deploy has a one-command rollback. The rollback is tested in staging before the deploy. The rollback is a separate, versioned artifact (not "deploy the previous version", which is fragile). The rollback is exercised in every quarterly DR drill.

## Image Signing (Mandatory)

Every container image is signed with `cosign` and verified at deploy time. An unsigned image cannot be deployed. The signing key is stored in a hardware security module (HSM), not in a developer laptop.

---

# Infrastructure as Code (Mandatory)

All infrastructure is defined in code (Terraform / Pulumi / AWS CDK). No manual console changes. A manual console change is a violation and must be reverted and replaced with a code change. Drift detection runs daily and pages on any deviation.

The `main` branch of the `infrastructure` repo is the source of truth. A PR to `main` triggers a plan → applies on merge to `main`. A PR to a feature branch can apply to `dev` for testing.

---

# Compatibility With Other Standards

| Standard | Relationship |
|---|---|
| [architecture-rules.md](architecture-rules.md) | This file is the devops implementation of the architecture's environment and migration strategy |
| [security-rules.md](security-rules.md) | The secret management forbid list here is the operational form of the secret rules there |
| [backend-rules.md](backend-rules.md) | CI/CD stages here run the tests defined there |
| [operations-rules.md](operations-rules.md) | Backup procedures here are referenced by the runbooks there; cost alerts are tier-1 monitoring there |
| [creator-economy-rules.md](creator-economy-rules.md) | Storage quotas here are the operational form of the creator tier limits there |
| [rag-rules.md](rag-rules.md) | AI token monitoring here measures the system defined there |

---

# Sprint 1 Compliance Checklist

- [ ] All 4 environments exist and are isolated (local, dev, staging, production)
- [ ] Environment promotion gate is enforced (local → dev → staging → production)
- [ ] Migration policy is documented in the operations runbook
- [ ] Expand-migrate-contract pattern is used for all Tier 1 destructive migrations
- [ ] Every destructive migration has an ADR before merge
- [ ] Every migration has a tested rollback plan
- [ ] Pre-commit hook (gitleaks) blocks secret commits
- [ ] CI scanner (trufflehog) blocks secret-containing PRs
- [ ] Quarterly secret rotation is on the security calendar
- [ ] All secrets are in AWS Secrets Manager / HashiCorp Vault (no .env in production)
- [ ] Monthly cost dashboard exists in Grafana with the 5-axis breakdown (AI, infra, third-party, cloud overhead, bandwidth)
- [ ] 80% / 100% / 120% / 150% budget alerts are configured
- [ ] Per-user and per-creator usage quotas are enforced at the API layer
- [ ] AI token monitoring has the 8-metric breakdown
- [ ] Storage growth monitoring has the 5-metric breakdown
- [ ] Quarterly DR drill has been executed at least once

---

# Final Directive (v1.0)

A migration without a backup is gambling.
A migration without a rollback plan is gambling with borrowed money.
A secret in source control is a breach waiting to happen.
A secret in a screenshot is a breach waiting to happen twice.
A service without usage quotas is a denial-of-wallet attack waiting to happen.
A cost dashboard that is not reviewed is a cost surprise waiting to happen.

Production discipline is not bureaucracy. Production discipline is what lets us ship fast without breaking things.

Every migration is additive first, destructive only with an ADR, backed up before, and rollback-tested.
Every secret lives in the secret manager. Nowhere else.
Every environment is isolated and gated.
Every cost has a budget and an alert.

DevOps is the difference between a prototype and a platform.
