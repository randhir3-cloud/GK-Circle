# COURSE-P1-T07 Delete Tests

Focused tests cover:

- validation without database access (nil Course, nil node);
- missing Course, missing node, and cross-Course node (course-scoped not found);
- leaf deletion;
- branch and deep subtree deletion;
- unrelated sibling/branch survival (delete scope limited to boundary-safe path);
- exact deleted-ID verification, including missing, unexpected, and duplicate RETURNING IDs;
- boundary-unsafe path rejection and empty subtree lock conflict;
- begin/query/delete/commit failures with rollback;
- `verifyDeleteIDs` set invariants;
- T03–T06 regression remaining green.

Subtree deletion implemented: Yes  
Archive implemented: No  
Restore implemented: No  
Database Migration: NONE
