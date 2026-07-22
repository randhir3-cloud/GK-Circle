# GK Circle — Production Deployment Checklist

**Version:** 1.1 (updated 2026-06-28 — split infra/application verification, script versioning, Phase 3.2 roadmap)
**Status:** Permanent — applies to every NUC production deployment
**Owner:** Engineering

---

## Purpose

This document defines the mandatory gate criteria for every GK Circle production release. A deployment is not complete until **all gates pass and all evidence is captured**.

---

## Definition of Done

> A deployment is ✅ **COMPLETE** only when:
>
> 1. All 7 **Deploy Gates** (G1–G7) pass → reported by `deploy-nuc-remote.ps1 v2.0`
> 2. All 4 **Infrastructure Verify Gates** (V1–V4) pass → reported by `verify-nuc-deployment.ps1 v1.0`
> 3. All 3 **Application Verify Gates** (V5–V7) pass → reported by `verify-nuc-deployment.ps1 v1.0`
> 4. All 7 **Smoke Tests** (S1–S7) pass → reported by `deployment-smoke.spec.ts`
> 5. **Release Manifest** is complete in the deployment folder
> 6. **Evidence artifacts** are present (screenshots, logs, backup path)

---

## Deploy Gates (run by `deploy-nuc-remote.ps1`)

| Gate | Name | Pass Condition |
|------|------|---------------|
| G1 | Git SHA | `git rev-parse HEAD` matches expected; `git pull` succeeded or was skipped with warning |
| G2 | Clean working tree | `git status --short` is empty |
| G3 | Docker build | `docker compose build` exits 0; backend and frontend image IDs captured |
| G4 | Database backup | `pg_dump → backup/YYYYMMDD_HHMMSS.sql.gz` exits 0; file size recorded |
| G5 | Prisma migrate deploy | `prisma migrate deploy` exits 0 |
| G6 | Prisma migrate status | `prisma migrate status` shows 0 pending, 0 failed |
| G7 | Compose up | `docker compose up -d` exits 0 |

---

## Verify Gates (run by `verify-nuc-deployment.ps1 v1.0`)

Verification is split into two explicit stages so failures have an immediate, unambiguous owner:

### Stage 1 — Infrastructure Verification

_Failures here mean a platform/ops problem. The application has not been reached yet._

| Gate | Name | Pass Condition |
|------|------|---------------|
| V1 | Container recreation | Backend and frontend containers `StartedAt` within 300s of verification run |
| V2 | Image ID match | Running container image IDs = IDs from G3 build |
| V3 | Health endpoint | `GET /health` → `{"status":"ok"}` |
| V4 | Docker logs clean | No `ERROR`/`FATAL` in last 50 backend log lines |

### Stage 2 — Application Verification

_Failures here mean an application/API/routing problem. Infrastructure is confirmed healthy._

| Gate | Name | Pass Condition |
|------|------|---------------|
| V5 | Backend API reachable | `GET /api/v1/auth/session` → HTTP 401 (backend alive, DB connected) |
| V6 | Frontend availability | `GET /` → HTTP 200 |
| V7 | Summary | All V1–V6 pass; summary block pasted into `verification.md` |

---

## Playwright Smoke Tests (run by `deployment-smoke.spec.ts`)

_The smoke suite is the Application Verification layer that requires real user sessions.
Failures here mean an authenticated route, authorization, or rendering problem._

QA accounts: `qa.student@gkcircle.com`, `qa.creator@gkcircle.com`, `qa.admin@gkcircle.com`
Credentials: stored in secure storage — never in source control.

| Test | Actor | Pass Condition |
|------|--------|---------------|
| S1 | Anonymous | `GET /health` → `{"status":"ok"}` |
| S2 | Anonymous | `GET /api/v1/auth/session` → HTTP 401 or 200 (not 502/503) |
| S3 | `qa.student` | Login → JWT issued |
| S4 | `qa.student` | `/dashboard` renders, no error page |
| S5 | `qa.student` | `/tests` renders, no error page |
| S6 | `qa.creator` | `/creator` renders, no error page, not redirected to login |
| S7 | `qa.admin` | `/admin` renders, no error page, not redirected to login |

**Command:**
```powershell
cd frontend
$env:PLAYWRIGHT_EXTERNAL_SERVER = "1"
$env:QA_STUDENT_PASSWORD = "<from secure storage>"
$env:QA_CREATOR_PASSWORD = "<from secure storage>"
$env:QA_ADMIN_PASSWORD = "<from secure storage>"
npx playwright test playwright/deployment-smoke.spec.ts --project=chromium-desktop
```

---

## Required Evidence Artifacts

Every deployment folder (`docs/deployment/YYYY-MM-DD-phase-X.Y.Z/`) must contain:

| Artifact | Description |
|----------|-------------|
| `deployment-report.md` | Overall deployment status and gate summary |
| `verification.md` | Detailed gate-by-gate results (V7 output pasted here) |
| `RELEASE-MANIFEST.md` | Version, SHA, images, migrations, features, rollback SHA |
| `migration-report.md` | Prisma migration details (G5/G6 output) |
| `known-issues.md` | Any observed issues and their resolution status |
| `screenshots/` | Playwright smoke screenshots (S4–S7) |

---

## Deployment Report Folder Naming

```
docs/deployment/YYYY-MM-DD-phase-X.Y.Z/
```

Example:
```
docs/deployment/2026-06-28-phase-3.1.7/
    deployment-report.md
    verification.md
    RELEASE-MANIFEST.md
    migration-report.md
    known-issues.md
    screenshots/
    logs/
```

---

## Database Backup Policy

- Every deployment automatically creates a backup at `~/apps/gkcircle/backup/YYYYMMDD_HHMMSS.sql.gz`
- **Retention:** Keep the last 10 backups. Review and prune after each deployment.
- Automated retention is not currently implemented — prune manually:
  ```bash
  ssh nuc 'ls -t ~/apps/gkcircle/backup/*.sql.gz | tail -n +11 | xargs rm -f'
  ```

---

## QA Account Policy

| Account | Role | Purpose |
|---------|------|---------|
| `qa.student@gkcircle.com` | student | Student dashboard, test library, practice mode verification |
| `qa.creator@gkcircle.com` | creator | Creator dashboard, product management verification |
| `qa.admin@gkcircle.com` | admin | Admin panel verification |

**Rules:**
- Never used by real users
- Never appear in XP rankings, leaderboards, analytics, or email campaigns
- Safe to reset credentials after each release
- All QA identity checks go through `QAIdentityService` — never hardcode email prefix checks

**Re-provisioning:**
```powershell
.\scripts\provision-qa-accounts.ps1
```

---

## Deployment Workflow

```
┌──────────────────────────────────────────────────────────┐
│  DEPLOY                                                  │
│  .\scripts\deploy-nuc-remote.ps1  (v2.0)                │
│  G1 Git SHA                                              │
│  G2 Clean tree                                           │
│  G3 Docker build                                         │
│  G4 DB backup                                            │
│  G5 Prisma migrate deploy                                │
│  G6 Prisma migrate status                                │
│  G7 docker compose up                                    │
│  → Exits on first failure with rollback guidance         │
└──────────────────┬───────────────────────────────────────┘
                   │
                   ▼
┌──────────────────────────────────────────────────────────┐
│  INFRASTRUCTURE VERIFICATION                             │
│  .\scripts\verify-nuc-deployment.ps1  (v1.0)            │
│  V1 Container recreation (StartedAt recency)             │
│  V2 Image ID match (running = freshly built)             │
│  V3 Health endpoint reachable                            │
│  V4 Docker logs clean                                    │
│  → Failure owner: Ops / Docker / Nginx                   │
└──────────────────┬───────────────────────────────────────┘
                   │
                   ▼
┌──────────────────────────────────────────────────────────┐
│  APPLICATION VERIFICATION                                │
│  .\scripts\verify-nuc-deployment.ps1  (continued)       │
│  V5 Backend API (HTTP 401 → DB connected)                │
│  V6 Frontend (HTTP 200)                                  │
│  V7 Summary → paste into verification.md                 │
│  → Failure owner: Backend / Frontend / Auth              │
│                                                          │
│  npx playwright test deployment-smoke.spec.ts            │
│  S1 Health API                                           │
│  S2 Auth endpoint reachable                              │
│  S3 QA student login → JWT issued                        │
│  S4 Student dashboard renders                            │
│  S5 Test library renders                                 │
│  S6 Creator dashboard renders                            │
│  S7 Admin dashboard renders                              │
│  → Failure owner: Application routes / Auth flows        │
└──────────────────┬───────────────────────────────────────┘
                   │
                   ▼
┌──────────────────────────────────────────────────────────┐
│  DOCUMENT                                                │
│  Fill in verification.md  (DEPLOYMENT-CHECKLIST-TEMPLATE)│
│  Fill in RELEASE-MANIFEST.md  (RELEASE-MANIFEST-TEMPLATE)│
│  Commit evidence to docs/deployment/YYYY-MM-DD-phase-*   │
└──────────────────┬───────────────────────────────────────┘
                   │
                   ▼
              ✅  DEPLOYMENT COMPLETE
```

---

## When to Run Verify Independently

`verify-nuc-deployment.ps1` can and should be run any time:
- After a manual `docker compose restart` on the NUC
- After a GitHub Actions deployment
- After rotating credentials or certificates
- To audit the current production state on-demand

---

## Script Versioning

Deployment scripts are versioned. The version is declared in the script header and recorded
in the Release Manifest so you know exactly which deployment workflow produced each release.

| Script | Current Version | Purpose |
|--------|-----------------|---------|
| `deploy-nuc-remote.ps1` | **v2.0** | Deploy (G1–G7) |
| `verify-nuc-deployment.ps1` | **v1.0** | Verify infrastructure + application (V1–V7) |
| `deployment-smoke.spec.ts` | **v1.0** | Playwright smoke suite (S1–S7) |

When a script is updated, bump its version and record the change in `CHANGELOG.md`.

---

## Phase 3.2 Roadmap — Dedicated Deployment Operations

> **Not for this phase.** Track as a future improvement.

As GK Circle grows and deployment operations multiply, consolidate into dedicated scripts:

```
scripts/
    deploy/
        deploy.ps1      # what deploy-nuc-remote.ps1 is today
        verify.ps1      # what verify-nuc-deployment.ps1 is today
        rollback.ps1    # restore from backup, roll back Git SHA
        backup.ps1      # on-demand DB backup (not tied to deploy cycle)
        restore.ps1     # restore specific backup by filename
```

This keeps the `scripts/` root clean and makes each operation independently callable.
Implement when you have 3+ deployment targets or the rollback flow is executed more than once per quarter.
