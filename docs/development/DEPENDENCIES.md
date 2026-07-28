# GK Circle Feature Dependency Graph

Strict feature dependencies of the Universal Chained Course System after DOC-CS-T02-R1.
Canonical ownership and task-level edges: `modules/course-system/ROADMAP.md`.

## Dependency Graph

```mermaid
graph TD
  P1[P1 Structure] --> P2[P2 Item composition]
  P2 --> P3[P3 Resources]
  P2 --> P4[P4 Assessments]
  P1 --> P5[P5 Course-wide nav and access]
  P2 --> P5
  P3 --> P5
  P4 --> P5
  P2 --> P6[P6 Progress]
  P5 --> P6
  P1 --> P7[P7 Templates and studio]
  P2 --> P7
  P3 --> P7
  P4 --> P7
  P6 --> P7
  P7 --> P8[P8 Hardening]
```

## Architectural Ordering Rules

1. Design-gate tasks precede related migrations.
2. A phase cannot be marked `VERIFIED` until checklist tasks are verified with evidence.
3. Later phases must not introduce product `MAX_*_DEPTH` or authoritative `child_node_ids[]`.
4. D-003 allows Phase 2 LearningItem work while Phase 1 UI remains open.
5. Phase 5 consumes Phase 1 breadcrumb/ancestor APIs; it does not redefine them.
