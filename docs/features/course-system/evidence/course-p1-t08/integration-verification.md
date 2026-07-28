# COURSE-P1-T08 Integration Verification

Environment: Fiber `app.Test` with sqlmock CourseModel and route middleware chain
(real `KratosAuthenticated` for 401; injected authenticated identities for
403/admin success).

| Case | Result |
|---|---|
| Unauthenticated create | PASSED (401) |
| Non-admin list | PASSED (403) |
| Admin create → list → get → patch | PASSED |
| Owner derived from auth context | PASSED |
| Empty list non-nil array | PASSED |
| Missing course 404 | PASSED |
| Malformed UUID 400 | PASSED |
| Explicit null PATCH | PASSED |
| CourseNode route absent | PASSED |
| CourseNode model regression | PASSED |

No production or NUC access.
