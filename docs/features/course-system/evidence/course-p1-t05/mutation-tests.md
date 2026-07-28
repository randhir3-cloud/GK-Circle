# COURSE-P1-T05 Mutation Tests

Focused tests cover:

- input validation without database access;
- root-to-root same-position no-op and same-parent reorder deferral;
- root-to-parent subtree move with exact destination position;
- leaf/subtree path handling, private-path JSON protection, and Course scope;
- self-parent and descendant-parent cycle rejection;
- missing Course/node and cross-Course destination parent;
- exact sibling-position conflicts without automatic renumbering;
- rollback on persistence failures, commit failures, and subtree rewrite-count mismatch;
- boundary-safe path selection: `abc/child` is included under `abc`, while `abcd` is excluded.

The local PostgreSQL transaction locked two subtree rows, rewrote exactly two
paths, updated the moved root parent/position, and preserved a boundary-control
path ending in an extra character.
