# COURSE-P1-T10 Integration Verification

Environment: Fiber `app.Test` with sqlmock CourseNodeModel and route middleware
chain (real `KratosAuthenticated` for 401; injected authenticated identities for
403/admin success).

| Case | Result |
|---|---|
| Unauthenticated move | PASSED (401) |
| Non-admin reorder/delete | PASSED (403) |
| Move position presence rules | PASSED (400) |
| Move missing parent | PASSED (404) |
| Move root noop → GET tree | PASSED |
| Reorder structural validation | PASSED (400) |
| Reorder mismatch | PASSED (400) |
| Reorder children → GET children | PASSED |
| Delete missing node/course | PASSED (404) |
| Delete leaf → GET tree | PASSED |
| Still-absent routes | PASSED (404/405) |
| T05–T09 regression | PASSED |

Docker Compose `api-verify` Go 1.23 vet/test/race/build: PASSED  
`docker compose config --quiet`: PASSED  
Frontend: NOT_APPLICABLE  
Database Migration: NONE  
Production / NUC: NOT_TOUCHED
