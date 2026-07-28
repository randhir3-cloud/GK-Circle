# Course System Architecture — Recursive Course Tree

Canonical module architecture for the Universal Chained Course System.

Hierarchy persistence authority:
[ADR-023](../../../../architecture/ADR/ADR-023-canonical-course-hierarchy-model.md).

Module clarifications:
[DECISIONS.md](../DECISIONS.md) (**D-004**, **D-005**).

Agent entry:
[README.md](../README.md).

---

## 1. Aggregate overview

```text
Course
└── CourseNode
    ├── parent_id → CourseNode | NULL     # AUTHORITATIVE structural edge
    ├── children  → derived WHERE parent_id = id   # PROJECTION only
    └── LearningItem*                     # ordered content on any node
```

- **Course** is the aggregate, ownership, and authorization root.
- **CourseNode** is the sole recursive structural hierarchy persistence model.
- **LearningItem** is ordered learning content attached to a CourseNode at any depth.
- CourseNode and LearningItem MUST remain separate persistence models unless a superseding ADR requires otherwise.
- Top-level overview modules (My Classes, Tests, PDFs, Mentorship, …) are root CourseNodes, templates, or configuration — not fixed Course columns.

**Non-normative example only** (not enforced schema levels):

`Subject → Module → Chapter → Lecture`

Builders may nest arbitrarily; persisted `node_type` today remains `SECTION` | `SUBJECT` | `TOPIC` until an ADR-approved expansion.

---

## 2. CourseNode persisted relationship

Each CourseNode persists (among other ADR-023 fields):

| Field | Role |
|---|---|
| `id` | Stable UUID |
| `course_id` | Owning Course (FK `ON DELETE RESTRICT`) |
| `parent_id` | **Sole authoritative structural edge** |
| `title` | Display title |
| `node_type` | Typed node (`SECTION` \| `SUBJECT` \| `TOPIC`) |
| `position` | Sibling order |
| `path` | Backend-owned UUID materialized path |
| `status` | Lifecycle (`DRAFT` \| `PUBLISHED` \| `ARCHIVED`) |
| timestamps | `created_at`, `updated_at` |

**Forbidden persistence:** authoritative `child_node_ids[]`, children JSONB, or any duplicated child-reference array.

Lifecycle naming: CourseNode uses **`status`**. LearningItem uses **`publish_state`**. Do not conflate.

---

## 3. Root-node contract

A root CourseNode has:

```text
parent_id = NULL
```

- Unlimited ordered roots per Course via Course-scoped `position`.
- Unique `(course_id, position)` among roots (`parent_id IS NULL`).

---

## 4. Nested-node contract

A nested CourseNode has:

```text
parent_id = direct parent CourseNode ID
```

Children are derived using:

```text
child.course_id = parent.course_id
AND child.parent_id = parent.id
```

- Parent and child MUST belong to the same Course.
- A node MUST NOT parent itself (`parent_id <> id`).

---

## 5. Computed child projections

API and domain models MAY expose computed projections:

- nested `children[]` on tree / subtree responses
- `child_count`, `has_children` on flat / builder list responses
- `learning_item_count` (later builder convenience)

These are **computed projections**, never authoritative persistence.

Canonical child retrieval:

```sql
SELECT *
FROM course_nodes
WHERE course_id = $1
  AND parent_id = $2
ORDER BY position ASC, id ASC;
```

---

## 6. Relationship invariants

Normative invariants (**D-005**), documented verbatim:

1. Every CourseNode belongs to exactly one Course.
2. Every root has parent_id = NULL.
3. Every non-root has exactly one parent_id.
4. Parent must belong to the same Course.
5. A node cannot parent itself.
6. A node cannot move under one of its descendants.
7. Children are derived from parent_id references and are not stored as child arrays.
8. A node may have zero, one, or unlimited children.
9. Sibling order is determined by position; child positions are parent-scoped and root positions are Course-scoped.
10. Moving a node moves its complete descendant subtree.
11. Node IDs remain unchanged during movement.
12. Parent, paths, sibling positions, and any future denormalized hierarchy metadata are updated transactionally.
13. Cross-Course movement is rejected.
14. Cross-Course copying is a separate validated workflow.
15. Full-tree and subtree APIs may expose computed children[].
16. Large trees may use lazy or paginated child APIs without changing the relationship contract.
17. No product MAX_*_DEPTH rule is permitted.

Additional delivery rules:

* Empty Course trees return successful non-null empty collections.
* Learner APIs must not leak draft or unauthorized structure/content.
* `course_nodes.course_id` FK is `ON DELETE RESTRICT` (migration `20260725134707`).

---

## 7. Ordering

- Sibling order = `position ASC`, then `id ASC` as tie-breaker in reads.
- Child positions are unique per `(course_id, parent_id)`.
- Root positions are unique per `(course_id)` where `parent_id IS NULL`.
- Reorder commits canonical `0..n-1` positions for the sibling set.

---

## 8. Materialized path

- Materialized path is an **identifier-based** ordered UUID sequence (ADR-023 / **D-001**).
- Slug paths (e.g. `/history/modern/`) are **non-normative illustrations only**, not the persistence codec.
- Path is backend-owned. Clients MUST NOT supply or rely on path in normal create/update/move DTOs.
- Path is **not** exposed on normal client JSON models unless a separately approved API task requires it.

---

## 9. Depth handling

- Nesting depth is data-driven; the schema does not change with depth (**D-004**).
- There is **no** persisted `depth` column today (ADR-023 / T03).
- Depth, if needed, is **computable** from hierarchy walk or path segments.
- A future denormalized `depth` column requires an approved ADR amendment and additive migration, and MUST NEVER enforce a maximum level.
- Product constants such as `MAX_COURSE_DEPTH`, `MAX_NODE_DEPTH`, or `MAX_HIERARCHY_LEVELS` are forbidden as product rules.
- Operational limits (payload size, rate limiting, page size, abuse protection) do **not** redefine the hierarchy as fixed-depth.

---

## 10. Full-tree retrieval

**Existing:** `GET /api/v1/admin/courses/:course_id/nodes/tree`

- Returns the Course hierarchy with computed nested `children[]`.
- Empty trees return successful non-null empty collections.
- Reads are Course-scoped.

---

## 11. Lazy child retrieval

**Existing:** `GET /api/v1/admin/courses/:course_id/nodes/:node_id/children` (direct children).

**Planned:** paginated / large-branch lazy retrieval with `has_children` / `child_count` builder projections without changing the parent-child contract.

---

## 12. Subtree retrieval target

**Planned:** dedicated subtree read returning a target node plus all descendants as computed nested `children[]`.

Do not claim this dedicated route exists until verified.

---

## 13. Ancestor and breadcrumb target

**Planned:** Course-scoped ancestor / breadcrumb reads (ordered root → current) without exposing path on normal DTOs.

Do not claim these routes exist until verified.

---

## 14. Transactional movement

**Existing:** `PATCH /api/v1/admin/courses/:course_id/nodes/:node_id/move`

In one transaction:

- validate Course scope, destination parent, exact position, and cycle boundaries;
- update `parent_id` and sibling `position`s;
- rewrite moved-node and descendant `path`s;
- preserve node IDs and Course membership;
- reject self-parent, descendant-parent, and cross-Course moves.

Any future denormalized depth metadata must update in the same transaction.

---

## 15. Transactional reordering

**Existing:** `POST /api/v1/admin/courses/:course_id/nodes/reorder`

- Validates ordered sibling ID sets.
- Locks Course (and parent when applicable) and siblings.
- Writes canonical positions; already-canonical orders are no-ops.

---

## 16. Transactional deletion

**Existing:** `DELETE /api/v1/admin/courses/:course_id/nodes/:node_id`

- Deletes the target node and its complete descendant subtree.
- Unrelated branches remain untouched.
- No automatic archive/soft-delete in the current verified contract.

---

## 17. Cross-Course movement versus copying

- **Move across Courses:** rejected.
- **Copy across Courses:** a separate validated workflow (**Deferred**; not implemented).

---

## 18. CourseNode and LearningItem coexistence

The same CourseNode may simultaneously hold:

- zero or more child CourseNodes;
- zero or more ordered LearningItems.

LearningItems attach at any depth; they never store structural child-node arrays.

**Non-normative example:** a container CourseNode labeled “Lecture 1” may hold Video/PDF/Quiz LearningItems while also parenting further CourseNodes.

---

## 19. Structural tree versus learning sequence versus prerequisites

| Graph | Edge meaning | Status |
|---|---|---|
| Structural CourseNode tree | `parent_id` hierarchy | Existing |
| Learning sequence | Deterministic preorder / prev-next / continue-learning | Phase 2 (node-local) + Phase 5 (course-wide); course-wide planned |
| Prerequisite dependency DAG | Unlocking edges; distinct from structural cycles | Planned (Phase 5+) |

Do not collapse these graphs into one edge type.

---

## 20. Admin builder implications

Builder UX (**Deferred** / Phase 1 UI gated) should:

- expand/collapse branches;
- show `child_count` / `has_children` / later `learning_item_count` as projections;
- support breadcrumb navigation once ancestor APIs exist;
- provide keyboard / non-drag move that calls server move/reorder APIs;
- lazy-load large branches;
- never treat a local `child_node_ids[]` cache as source of truth after refresh.

---

## 21. Learner navigation implications

- Learner structure/content reads must filter by publication and authorization server-side.
- **Existing:** authenticated learner LearningItem GET list/get (published only; Course enrollment required — D-006 / T12).
- **Planned:** publish-safe learner outline / tree.
- Do not claim learner tree, subtree, or breadcrumb APIs already exist.

---

## 22. Large-tree and deep-tree safety

- Prefer CTE / iterative traversal over unsafe deep application recursion.
- Operational page/payload/rate limits are abuse controls, not product depth.
- Deep-tree stress suites (e.g. ≥25 levels) are roadmap verification, not claimed by historical Phase 1 VERIFIED tasks unless their evidence says so.

---

## 23. Authorization

- Admin Course / CourseNode / LearningItem APIs: Kratos authentication + current quiz-admin allowlist (known temporary gate; not a new role model).
- Learner LearningItem APIs: `KratosAuthenticated` + Course enrollment (`course_enrollments`; D-006 / COURSE-P2-T12). Unenrolled callers receive documented denial without LearningItem payloads.
- Unauthorized mutations must leave no node/path/position/descendant row changes.

---

## 24. Publication filtering

| Entity | Field | Learner rule |
|---|---|---|
| CourseNode | `status` | Draft/archived structure must not leak beyond documented omission/not-found policy |
| LearningItem | `publish_state` | Only `PUBLISHED` via repository-owned published reads |

Controllers must not invent client-trusted publish filters.

---

## 25. Known implementation gaps

Labeled capability status:

| Capability | Label |
|---|---|
| Course root persistence + admin Course APIs | Existing |
| Recursive CourseNode persistence | Existing |
| Course-scoped parent relationships | Existing |
| Root and child ordering | Existing |
| Hierarchy reads (roots, children, full tree) | Existing |
| Transactional move / reorder / subtree delete | Existing |
| Admin CourseNode create/read/mutation APIs | Existing |
| LearningItem admin CRUD + learner published reads | Existing |
| Dedicated subtree read | Planned |
| Ancestor / breadcrumb read | Planned |
| Learner published tree / outline | Planned |
| Lazy/paginated large-branch retrieval + builder counts | Planned |
| Course admin UI / recursive builder | Deferred |
| Cross-Course copy subtree | Deferred |
| Enrollment enforcement | Deferred |
| Runtime visibility evaluation | Deferred |
| Archive/restore workflows | Deferred |
| Authorization role redesign | Deferred |

---

## 26. ADR and decision references

| Reference | Role |
|---|---|
| [ADR-023](../../../../architecture/ADR/ADR-023-canonical-course-hierarchy-model.md) | Canonical hierarchy persistence decision |
| [D-001](../DECISIONS.md#d-001--use-identifier-based-materialised-paths) | UUID materialized paths |
| [D-002](../DECISIONS.md#d-002--canonical-typed-coursenode-hierarchy) | Typed CourseNode backbone |
| [D-003](../DECISIONS.md#d-003--parallel-phase-2-learningitem-persistence-before-phase-1-ui-completion) | Parallel Phase 2 authorization |
| [D-004](../DECISIONS.md#d-004--unlimited-logical-course-tree-depth) | Unlimited logical depth |
| [D-005](../DECISIONS.md#d-005--coursenode-parent-child-relationship-contract) | Parent-child Relationship Contract |
| [SYSTEM_ARCHITECTURE.md](../../../SYSTEM_ARCHITECTURE.md) | Platform + Course hierarchy snapshot |
| [README.md](../README.md) | Agent protocol + contract summary |

### Domain response shapes (documentation target)

Separate:

- **Persisted CourseNode** — id, course_id, parent_id, title, node_type, position, path (not in normal client DTOs), lifecycle `status`, timestamps
- **CourseTreeNode** — persisted fields + computed `children[]`
- **CourseNodeListItem** — persisted fields + `has_children`, `child_count` (and later learning-item counts)

JSON examples in docs may use camelCase for readability; Go/API wire format remains repository conventions (`course_id`, snake_case) unless an approved API change task says otherwise.

### Admin API surface (Existing / Planned / Deferred)

| Intent | Route / capability | Label |
|---|---|---|
| Add root/child | `POST /api/v1/admin/courses/:course_id/nodes` (`parent_id` null/set) | Existing |
| List roots | `GET /api/v1/admin/courses/:course_id/nodes` | Existing |
| Get node | `GET /api/v1/admin/courses/:course_id/nodes/:node_id` | Existing |
| Direct children | `GET .../nodes/:node_id/children` | Existing |
| Full tree | `GET .../nodes/tree` | Existing |
| Move | `PATCH .../nodes/:node_id/move` | Existing |
| Reorder siblings | `POST .../nodes/reorder` | Existing |
| Delete subtree | `DELETE .../nodes/:node_id` | Existing |
| Dedicated subtree read | — | Planned |
| Ancestors / breadcrumbs | — | Planned |
| Learner published tree | — | Planned |
| Lazy/paginated large-branch + builder counts | — | Planned |
| Cross-Course copy | — | Deferred |
| Course / tree builder UI | — | Deferred |

---

## Implementation snapshot (evidence-backed)

### Detected platform

- ✓ Nuxt 3 and Vue 3 frontend.
- ✓ Go 1.25 and Fiber v2 API.
- ✓ PostgreSQL 15, goqu, and sql-migrate.
- ✓ Ory Kratos, Redis-compatible coordination, MinIO, Vitest, Playwright, and Compose.

### Implemented by Course System

- ✓ Deterministic ledger sync/check; ADR-023 accepted; decisions D-001–D-005.
- ✓ `courses` and `course_nodes` tables; LearningItem tables with `publish_state`.
- ✓ No authoritative `child_node_ids` column; no persisted `depth` column.
- ✓ CourseNode create/lookup/root/child/hierarchy reads; nested hierarchy assembly.
- ✓ Transactional move, sibling reorder, and subtree deletion.
- ✓ Admin Course and CourseNode HTTP APIs under `/api/v1/admin/courses`.
- ✓ Admin LearningItem HTTP CRUD; learner published LearningItem GET.
- ✗ Course UI builder, dedicated subtree/ancestor APIs, learner tree, copy-subtree: Planned / Deferred as labeled above.
- ✓ Learner LearningItem enrollment gate (COURSE-P2-T12 / D-006).
- ✓ Runtime visibility projection (COURSE-P2-T13).

### Known limitations

- Admin Course APIs reuse the quiz-admin allowlist as the current administrative gate.
- Learner LearningItem APIs require Kratos authentication and Course enrollment (`course_enrollments`).
- Phase 1 UI remains separately gated under D-003.
- Historical VERIFIED tasks prove adjacency + path + mutations; they do not claim 25-level stress suites, ancestor APIs, or builder UI unless their evidence says so.
