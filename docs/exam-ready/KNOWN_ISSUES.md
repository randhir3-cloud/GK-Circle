# Known Issues — Phase 3.1.8

> Log discovered bugs, constraints, and edge cases here. Assign priority and owner.

---

## Planning Phase Issues

### ISSUE-001: TestSeries currently requires a productId

**Discovered:** 2026-07-04  
**Affects:** TICKET-036 (Test Collections)  
**Priority:** Medium  
**Details:** The existing `TestSeries` model has `productId String` (non-nullable). When implementing Test Collections, standalone series not tied to a Product will need either (a) a nullable migration for `productId` or (b) a system/placeholder Product created per creator. This needs a decision before TICKET-036 starts.  
**Recommendation:** Make `productId` nullable as part of the `m2_collections` migration. This is additive (nullable) and safe.  
**Status:** Open — decision needed before TICKET-036

---

### ISSUE-002: Profile.streak is not currently incremented

**Discovered:** 2026-07-04  
**Affects:** TICKET-023 (Daily Streak)  
**Priority:** Low  
**Details:** `Profile.streak` field exists in schema but no service currently writes to it. TICKET-023 must implement the increment logic (check `lastActivityDate`, increment or reset) as part of the answer submission flow.  
**Status:** Open — addressed in TICKET-023

---

### ISSUE-003: Note model has no questionId field

**Discovered:** 2026-07-04  
**Affects:** TICKET-031 (Personal Notes on Questions)  
**Priority:** Low  
**Details:** The `Note` model has `subjectId` and `topicId` but no `questionId`. This is resolved by the m1_questions migration (adds `questionId String?` to Note). TICKET-031 must run AFTER TICKET-001.  
**Status:** Open — resolved by TICKET-001 migration

---

### ISSUE-004: Import rollback does not clean up uploaded images

**Discovered:** 2026-07-04  
**Affects:** TICKET-005 (Import History + Rollback)  
**Priority:** Medium  
**Details:** If a CSV import includes image URLs that were uploaded separately, rolling back the import (soft-deleting the questions) does not delete the uploaded images from S3/MinIO. This is a storage leak.  
**Recommendation:** For MVP, accept this limitation and document it. Phase 4 cleanup job can reclaim orphaned uploads.  
**Status:** Accepted for MVP — documented as known limitation

---

### ISSUE-005: TestSeries.productId nullable migration needed for Collections

**Discovered:** 2026-07-04  
**Affects:** TICKET-001 (Schema Migrations), TICKET-036 (Collections)  
**Priority:** High  
**Details:** Same as ISSUE-001 above. The m2_collections migration must include making `TestSeries.productId` nullable to support creator-owned series not attached to a Product.  
**Status:** Open — must be resolved in TICKET-001

---

## Bug Bash — UI Audit (post RC-1)

### BUG-UI-001: Dark Mode Audit

**Discovered:** 2026-07-10  
**Priority:** Medium  
**Details:** Product dashboard and content-management flows not fully verified in dark mode during RC-1 product gate. Audit while creating Punjab PCS content.  
**Status:** Open

### BUG-UI-002: Mobile Audit

**Discovered:** 2026-07-10  
**Priority:** Medium  
**Details:** Product dashboard tabs, stat cards, and delete dependency dialog not fully verified at mobile breakpoints (360–390px) during RC-1 product gate. Audit during real study sessions.  
**Status:** Open

