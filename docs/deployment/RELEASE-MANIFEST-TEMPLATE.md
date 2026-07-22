# Release Manifest — Phase X.Y.Z
<!-- Copy this file to docs/deployment/YYYY-MM-DD-phase-X.Y.Z/RELEASE-MANIFEST.md -->
<!-- Complete ALL fields. This is the authoritative record of what was released. -->

---

## Identity

| Field | Value |
|-------|-------|
| Version | Phase X.Y.Z |
| Release Name | _(e.g. "Student Assessment Experience Integration")_ |
| Git SHA | ________ (short) / ________ (full) |
| Safety Tag | `pre-nuc-phase-X.Y.Z` |
| Deployment Time | YYYY-MM-DD HH:MM UTC |
| Deployed By | _(operator name or `deploy-nuc-remote.ps1`)_ |
| Verified By | _(name or agent)_ |
| Target | Intel NUC · `~/apps/gkcircle` · https://gkcircle.com |
| Deploy Script | `deploy-nuc-remote.ps1 v2.0` |
| Verify Script | `verify-nuc-deployment.ps1 v1.0` |

---

## Runtime State at Release

| Component | Image ID | Container Started At |
|-----------|----------|---------------------|
| Backend (`gk-circle-backend`) | ________ | YYYY-MM-DD HH:MM:SS UTC |
| Frontend (`gk-circle-frontend`) | ________ | YYYY-MM-DD HH:MM:SS UTC |
| Nginx (`gk-circle-nginx`) | nginx:alpine | _(unchanged)_ |
| Postgres | _(external)_ | _(unchanged)_ |

---

## Environment Snapshot

_Capture these at deployment time. They become invaluable during future upgrades and incident investigations._

Capture with:
```bash
ssh nuc 'docker compose version --short; docker version --format "{{.Server.Version}}"; node --version; docker exec gk-circle-postgres psql --version; docker compose exec backend npx prisma --version 2>/dev/null | head -1; docker compose exec frontend node -e "const p=require(\"/app/package.json\"); console.log(\"next:\",p.dependencies.next)" 2>/dev/null; docker compose exec backend node -e "const p=require(\"/app/package.json\"); console.log(\"nestjs:\",p.dependencies[\"@nestjs/core\"])" 2>/dev/null'
```

| Component | Version |
|-----------|---------|
| Docker Compose | v________ |
| Docker Engine | v________ |
| Node.js (NUC host) | v________ |
| PostgreSQL | ________ |
| Prisma | ________ |
| Next.js | ________ |
| NestJS | ________ |

---

## Database

| Field | Value |
|-------|-------|
| Migration IDs applied | _(list migration names)_ |
| Total migrations | _(N applied, 0 pending)_ |
| DB backup | `backup/YYYYMMDD_HHMMSS.sql.gz` (__ MB) |
| Rollback SQL | _(the above backup restores pre-release state)_ |

---

## Rollback

| Field | Value |
|-------|-------|
| Rollback SHA | ________ (previous release commit) |
| Rollback tag | `pre-nuc-phase-X.Y.Z` |
| Rollback procedure | `git checkout <rollback-sha>` → `docker compose build` → `docker compose up -d` |
| DB rollback | Restore from backup listed above (migrations may need manual rollback) |

---

## Features Shipped

<!-- List each feature/capability delivered by this release -->

- _(Feature 1)_
- _(Feature 2)_

---

## Breaking Changes

<!-- List any breaking changes to APIs, schemas, or user-facing behavior -->

- None
<!-- OR -->
- _(Breaking change description + migration path)_

---

## Verification

| Verification Step | Result |
|-------------------|--------|
| **Infrastructure** | |
| V1 Container recreation | ✅ PASS / ❌ FAIL |
| V2 Image ID match | ✅ PASS / ❌ FAIL |
| V3 Health endpoint | ✅ PASS / ❌ FAIL |
| V4 Docker logs clean | ✅ PASS / ❌ FAIL |
| **Application** | |
| V5 Backend API reachable | ✅ PASS / ❌ FAIL |
| V6 Frontend availability | ✅ PASS / ❌ FAIL |
| V7 Overall summary | ✅ PASS / ❌ FAIL |
| **Smoke Tests** | |
| S1–S7 Playwright smoke | ✅ PASS / ❌ FAIL |
| **Deploy Gates** | |
| G1–G7 Deploy pipeline | ✅ PASS / ❌ FAIL |
| Full verification report | See `verification.md` |

---

## Known Issues

<!-- Issues observed at release time — link to known-issues.md for detail -->

- None
<!-- OR -->
- _(Issue description — see known-issues.md #KI-X.Y.Z-NN)_

---

## Audit Trail

| Document | Location |
|----------|----------|
| Deployment report | `deployment-report.md` |
| Verification log | `verification.md` |
| Migration report | `migration-report.md` |
| Known issues | `known-issues.md` |
| Screenshots | `screenshots/` |
| DB backup | `~/apps/gkcircle/backup/YYYYMMDD_HHMMSS.sql.gz` |
