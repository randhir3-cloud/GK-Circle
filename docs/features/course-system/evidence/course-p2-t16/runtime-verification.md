# Authenticated Runtime Verification

Date: 2026-07-27  
Environment: local Docker Compose only  
Authentication: documented local Kratos QA identity; no credential material
recorded

## Admin flow

1. Signed in normally and confirmed capability-gated `Course Content`.
2. Selected `T09 Verification Course`.
3. Drilled one direct-child request at a time to `Depth 4 Node`.
4. Observed server-returned rows at positions 0 and 1.
5. Created a temporary DRAFT item at position 2.
6. Reloaded the browser, repeated the explicit drill-down, and read the item
   back.
7. Updated its title, cancelled deletion once, then confirmed deletion.

## Learner and persistence flow

- Learner list showed the published position-0 item and excluded the DRAFT
  position-1 sibling.
- Detail rendered the valid empty-block state: `No content available.`
- PostgreSQL final read-back contained exactly the two original rows; no
  temporary T16 row remained.
- Browser console contained no warning or error entries.
- Desktop and 360px viewport layouts retained usable headings, controls,
  selectors, cards, and links.
- Logout was the final browser action; revisiting the protected detail was
  denied by redirect.

Representative non-empty block and safe-link runtime behavior is already
authenticated and sealed under the verified T14 evidence. T16 independently
reran all renderer and URL regression tests.
