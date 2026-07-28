# COURSE-P2-T09 Closure Audit

Audited: 2026-07-27 (CLOSE-02 → completion under Project Completion Rule)

| Frozen acceptance item | Implemented | Automated verification | Runtime evidence | Result |
|---|---|---|---|---|
| Admin UI lists, creates, edits, and deletes node-local items at any depth | Yes | Focused Vitest 12/12 PASS | Desktop deep-node CRUD PASS | PASS |
| Server item order; no `child_node_ids[]` hierarchy authority | Yes | Focused Vitest PASS | UI order A then B matched server positions 0,1 | PASS |
| Simple lazy Course/CourseNode picker independent of Phase 1 editor | Yes | Focused Vitest PASS | Lazy root→child levels 1–3 to depth-4 node | PASS |
| Evidence bundle; migration NONE | Yes | Bundle + hashes | Screenshots desktop+mobile | PASS |

## Root-cause fixes applied (in scope)

| Blocker | Fix |
|---|---|
| No allowlisted admin credentials | Seeded `local.course.admin@gk-circle.local` + `PUBLIC_QUIZ_ADMIN_EMAILS` |
| API binary without Course routes | `docker compose build api` + recreate |
| `relation "courses" does not exist` | Restored missing historical migration files; `migrate up` for Course tables |

## Automated gates

| Gate | Result |
|---|---|
| Focused Vitest Course suite | PASS 12/12 |
| Full suite | INHERITED BASELINE 35 failed / 94 passed (T09 tests pass) |
| Lint / build / compose web | PASS (prior CLOSE-02 + unchanged T09 UI) |
| Signed-out denial | PASS |
| Authenticated allowlisted runtime | PASS |
| Desktop / mobile screenshots | PASS |

Status: `VERIFIED`

Database Migration: NONE (for T09 task itself)

Breaking Changes: NO
