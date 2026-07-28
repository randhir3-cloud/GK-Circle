# COURSE-P1-T07 PostgreSQL Verification

Environment: temporary local PostgreSQL 15 database `course_p1_t07_verify`
(container `course-p1-t07-pg`, port `55433`). Deleted after verification.

| Case | Result |
|---|---|
| Delete leaf | PASSED |
| Delete branch; unrelated root survives | PASSED |
| Delete large deep subtree; unrelated branch unchanged | PASSED |
| Concurrent delete vs reorder | PASSED |
| Concurrent delete vs move | PASSED |
| Rollback on missing node | PASSED |

Observations:

- Single-statement boundary-safe subtree `DELETE ... RETURNING id` removed the
  locked root and descendants without violating the self-referential
  `ON DELETE RESTRICT` parent FK when all matching rows were removed together.
- Concurrent hierarchy mutations serialized on the Course-row lock.
- No sibling renumbering or path rewriting occurred for surviving nodes.

Subtree deletion implemented: Yes  
Archive implemented: No  
Restore implemented: No  
Database Migration: NONE  
Production touched: No  
NUC touched: No
