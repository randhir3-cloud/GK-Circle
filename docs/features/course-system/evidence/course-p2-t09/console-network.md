# COURSE-P2-T09 Console and Network Record (CLOSE-02)

## Scope

Recorded around automated registration, capability-gated redirect, and
signed-out denial. Allowlisted Course Content CRUD was not reached.

## Observations

- Registration and login completed without user-visible JS failure in the
  exercised admin shell.
- Direct Learning Items navigation while authenticated without
  `canCreatePublicQuiz` performed an in-app redirect to `/admin/quiz/list-quiz`
  (application capability gate; not an unexpected auth loss).
- Signed-out Learning Items navigation redirected to `/account/login` without
  exposing Course Content UI.
- No allowlisted Learning Item create/edit/delete mutations were issued.

## Closure impact

Console/network for required allowlisted CRUD flows: **BLOCKED** (same auth
blocker). Signed-out denial path: **PASS**.
