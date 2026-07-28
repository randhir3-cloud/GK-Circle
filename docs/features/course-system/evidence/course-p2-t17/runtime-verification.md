# T17 Runtime Verification

## Environment

The run targeted only localhost in Compose project `gkcirclev2`. API, web,
Kratos, PostgreSQL, Redis, and Mailpit were running locally. PostgreSQL had no
published host port. No production or NUC endpoint was configured or accessed.

## Correlated workflow

1. Kratos authenticated the approved local Course-admin QA identity.
2. The admin API returned an existing Course and direct-child chain.
3. The UI selected four CourseNode levels one server-provided level at a time.
4. Admin UI POST created a uniquely titled ARTICLE in DRAFT state.
5. Reload and re-selection returned the same API-assigned row.
6. The learner API/UI omitted that DRAFT ID.
7. Admin UI PATCH changed only publication state to PUBLISHED.
8. Learner API returned the item; DOM IDs exactly matched API array order.
9. Learner detail returned the same title and default empty block envelope,
   rendered as `No content available.`
10. Previous/next links, when present, matched only the API-provided IDs.
11. Desktop 1280×900 and mobile 360×800 layouts rendered successfully.
12. A fresh unauthenticated context received HTTP 401 and displayed an alert.
13. Authenticated DELETE returned 200; subsequent admin GET returned 404.

Sanitized PostgreSQL correlation after the run:

```text
deep_node_ancestor_count=4
temporary_t17_rows=0
deep_node_learning_items=2
```

The two final deep-node rows are the pre-existing verification seeds. T17
created no Course hierarchy and left no LearningItem or enrollment residue.
