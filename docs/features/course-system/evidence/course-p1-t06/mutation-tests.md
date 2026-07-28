# COURSE-P1-T06 Mutation Tests

Focused tests cover:

- parameter validation without database access (nil course/parent/node IDs, duplicates);
- sibling-set matching (missing entries, extra IDs, foreign replacements);
- empty sibling scope success and mismatch rejection, including commit-failure wrapping;
- single-sibling and already-canonical no-ops with no UPDATE statements;
- unique-constraint-safe two-node swap and three-node rotation via temporary staging;
- int32 overflow rejection for temporary position ranges;
- update-set verification conflict when RETURNING yields no row;
- deterministic staging and verification for a 100-sibling reverse reorder;
- transaction begin/lock/update/commit failure paths with rollback;
- `verifyReorderUpdatedIDs` duplicate-hiding and exact-set invariants;
- T03/T04/T05 regression coverage remaining green.

Isolated PostgreSQL verification rotated three root siblings in one transaction
and ran two concurrent reorders that serialized on the Course-row lock without
violating sibling-position uniqueness.
