# Checker Integrity Verification

The checker rejected every deliberately corrupted isolated fixture:

- cross-document next-task mismatch;
- multiple `IN_PROGRESS` tasks without `parallel=true`;
- phase weights not totaling 100%;
- `VERIFIED` task without evidence;
- `VERIFIED` task with unchecked acceptance criteria;
- `VERIFIED` task with an unresolved risk marker.

During the initial real sync/check pair, it also detected that Phase 2 declared
85 points while its task rows totaled 83. The declaration was corrected to 83
before the first successful sync/check.

The read-only check and sync idempotency results are recorded in
`isolated-clean-copy.md` and the hash manifests under `hashes/`.
