# T17 Cleanup Proof

Cleanup runs from `finally` through the authenticated browser context.

For every created item:

1. Admin DELETE must return HTTP 200.
2. Admin GET for the exact Course/node/item URL must return HTTP 404.
3. Final PostgreSQL count for `T17 Learning Item %` must be zero.

If the test created an enrollment, DELETE must return 200 and GET must report
`enrolled: false`. The approved QA identity was already enrolled during the
final run, so T17 did not create or remove enrollment state.

The initial timeout left two precisely identified temporary items because the
overall Playwright timeout cancelled teardown. A one-time local cleanup first
read each exact ID, validated its `T17 Learning Item` title, deleted it, and
verified HTTP 404. The temporary cleanup script was removed immediately.

Final result:

```text
temporary_t17_rows=0
```
