# Deployment Verification — Phase X.Y.Z
<!-- Copy this file to docs/deployment/YYYY-MM-DD-phase-X.Y.Z/verification.md -->
<!-- Fill in all [ ] fields before marking the deployment complete -->

**Phase:** X.Y.Z  
**Date:** YYYY-MM-DD  
**Deployed By:** _(name or script)_  
**Verified By:** _(name or agent)_  
**Target:** Intel NUC (`ssh nuc`, `~/apps/gkcircle`)  
**Public URL:** https://gkcircle.com

---

## Overall Status

| Section | Status |
|---------|--------|
| Deploy Gates (G1–G7) | [ ] ✅ PASS / ❌ FAIL |
| Verify Gates (V1–V7) | [ ] ✅ PASS / ❌ FAIL |
| Smoke Tests (S1–S7) | [ ] ✅ PASS / ❌ FAIL |
| Release Manifest | [ ] Complete / Pending |
| **Overall** | [ ] **✅ DEPLOYMENT COMPLETE** / **❌ INCOMPLETE** |

---

## Deploy Gates — `deploy-nuc-remote.ps1` Output

<!--
Paste the gate summary from deploy-nuc-remote.ps1 here.
Example:

  G1  ✅  Git SHA: bd27a338 (clean)
  G2  ✅  Clean tree: clean
  G3  ✅  Docker build: backend d53715da | frontend 7d8c995e
  G4  ✅  DB backup: backup/20260628_142301.sql.gz (12 MB)
  G5  ✅  Migrate deploy: 1 new migration applied
  G6  ✅  Migrate status: all applied no pending
  G7  ✅  Compose up: containers recreated
-->

```
[paste deploy gate summary here]
```

| Gate | Name | Result | Detail |
|------|------|--------|--------|
| G1 | Git SHA | [ ] ✅ / ❌ | SHA: ________ |
| G2 | Clean tree | [ ] ✅ / ❌ | |
| G3 | Docker build | [ ] ✅ / ❌ | backend: ________ frontend: ________ |
| G4 | DB backup | [ ] ✅ / ❌ | file: ________ size: ________ |
| G5 | Migrate deploy | [ ] ✅ / ❌ | ________ |
| G6 | Migrate status | [ ] ✅ / ❌ | all applied / pending: ________ |
| G7 | Compose up | [ ] ✅ / ❌ | |

---

## Infrastructure Verification — `verify-nuc-deployment.ps1 v1.0`

<!--
Paste the V7 summary block from verify-nuc-deployment.ps1 here.
Example:

════════════════════════════════════════════════════════════
 GK Circle NUC — VERIFICATION SUMMARY
════════════════════════════════════════════════════════════
  V1  ✅  Container times:    backend 14s | frontend 12s | nginx 8s
  V2  ✅  Image IDs:          backend d53715da ✓ | frontend 7d8c995e ✓
  V3  ✅  Health:             {"status":"ok"}
  V4  ✅  Docker logs:        clean (no ERROR/FATAL in last 50 lines)
  V5  ✅  Backend API:        GET /api/v1/auth/session → HTTP 401 (reachable)
  V6  ✅  Frontend:           HTTP 200 OK
  V7  ✅  Overall result
════════════════════════════════════════════════════════════
-->

```
[paste V7 summary block here]
```

**Failure owner for V1–V4:** Ops / Docker / Nginx

| Gate | Stage | Result | Detail |
|------|-------|--------|--------|
| V1 | Infrastructure | [ ] ✅ / ❌ | backend: ___s frontend: ___s |
| V2 | Infrastructure | [ ] ✅ / ❌ | backend match: __ frontend match: __ |
| V3 | Infrastructure | [ ] ✅ / ❌ | `{"status":"ok"}` |
| V4 | Infrastructure | [ ] ✅ / ❌ | clean / __ ERROR lines |

**Failure owner for V5–V6:** Backend / Frontend / Auth

| Gate | Stage | Result | Detail |
|------|-------|--------|--------|
| V5 | Application | [ ] ✅ / ❌ | HTTP ________ |
| V6 | Application | [ ] ✅ / ❌ | HTTP 200 |
| V7 | Summary | [ ] ✅ PASS / ❌ FAIL | |

---

## Smoke Tests — `deployment-smoke.spec.ts`

**Run command:**
```powershell
cd frontend
$env:PLAYWRIGHT_EXTERNAL_SERVER = "1"
$env:QA_STUDENT_PASSWORD = "<from secure storage>"
$env:QA_CREATOR_PASSWORD = "<from secure storage>"
$env:QA_ADMIN_PASSWORD = "<from secure storage>"
npx playwright test playwright/deployment-smoke.spec.ts --project=chromium-desktop
```

**Playwright output:**
```
[paste playwright output here]
```

| Test | Actor | Result | Screenshot |
|------|--------|--------|------------|
| S1: Health API | Anonymous | [ ] ✅ / ❌ | — |
| S2: Auth reachable | Anonymous | [ ] ✅ / ❌ | — |
| S3: QA student login | `qa.student` | [ ] ✅ / ❌ | — |
| S4: Student dashboard | `qa.student` | [ ] ✅ / ❌ | screenshots/s4-student-dashboard.png |
| S5: Test library | `qa.student` | [ ] ✅ / ❌ | screenshots/s5-test-library.png |
| S6: Creator dashboard | `qa.creator` | [ ] ✅ / ❌ | screenshots/s6-creator-dashboard.png |
| S7: Admin dashboard | `qa.admin` | [ ] ✅ / ❌ | screenshots/s7-admin-dashboard.png |

---

## Evidence Artifacts

| Artifact | Status |
|----------|--------|
| `deployment-report.md` | [ ] Present |
| `verification.md` (this file) | [ ] Complete |
| `RELEASE-MANIFEST.md` | [ ] Complete |
| `migration-report.md` | [ ] Present |
| `known-issues.md` | [ ] Present |
| `screenshots/` (S4–S7) | [ ] Captured |
| DB backup path | backup/____________ |

---

## Known Issues

_(List any issues found during verification. Use `known-issues.md` for full detail.)_

- None / see known-issues.md

---

## Sign-Off

> Deployment is ✅ **COMPLETE** — all gates pass, all evidence captured, Release Manifest complete.

**Signed off by:** ________________  
**Date:** YYYY-MM-DD
