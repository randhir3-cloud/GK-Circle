# Migration Root Cause Analysis

Date: 2026-07-28  
Task: EXAM-P6-T02 Migration Reconciliation  
Status: **ROOT CAUSE IDENTIFIED AND REPAIRED**

---

## Environment Capture

| Item | Value |
|---|---|
| Repository root | `E:/GK Circle v2` |
| Branch | `chore/ci-verification` |
| Commit SHA | `eeac599f05eaf936c7f61db4a3deeac3c9063f59` |
| Go version | `go1.26.5 windows/amd64` |
| Node version | `v24.16.0` |
| npm version | `11.16.0` |
| Docker Compose version | `v5.1.3` |
| API image ID | `fe0b7f54248e` (created `2026-07-28T01:36:03Z`) |
| API image tag | `gkcirclev2-api:latest` |
| Migration image (before repair) | `0f569e8aea07` (created `2026-07-27T00:26:10Z`) |
| Migration image tag | `gkcirclev2-migration:latest` |
| DB container name | `gk-circle-db` |
| DB volume name | `gk-circle-pgdata` |
| DB name | `gk_circle` |
| DB user | `gk_circle` |
| Migration tool | `github.com/rubenv/sql-migrate` (FileMigrationSource) |
| Migration metadata table | `gorp_migrations` |
| Migration directory (in container) | `/gk-circle/database/migrations` |
| Last applied migration (before repair) | `20260727100000_create_course_enrollments_table.up.sql` (applied `2026-07-27 01:10:35+00`) |

---

## Observed Errors

### Error 1: Missing column
```
pq: column "answer_review_status" of relation "questions" does not exist
```

### Error 2: Unknown migration
```
migrate: Unable to create migration plan because of 
20260727100000_create_course_enrollments_table.down.sql: unknown migration in database
```

---

## Root Cause: Stale Migration Image (Scenario G + B)

The root cause is a **stale Docker migration image** combined with a **partially-applied migration chain**.

### Timeline reconstruction

| Time | Event |
|---|---|
| `2026-07-27T00:26:10Z` | Migration image built with migrations up through `20260726094500`. This image does NOT include `20260727100000` or later migrations. |
| `2026-07-27 01:10:34+00` | `docker compose up` run. Migration container (from old image) successfully applied `20260727100000_create_course_enrollments_table` — wait, this is contradictory. |
| `2026-07-27 01:10:35+00` | DB ledger shows `20260727100000` applied successfully at this time. |
| `2026-07-28T01:36:03Z` | API image rebuilt with ALL current migrations (`20260728120000` and earlier). |
| `2026-07-28 01:35:10` | `docker compose up` run. Migration container uses OLD image (`gkcirclev2-migration` from `2026-07-27T00:26:10`). The old image does NOT know about `20260727100000` (which IS in the DB ledger as applied). This causes "unknown migration in database" error. |
| `2026-07-28 02:31` | Migration image **rebuilt** from current source. Image now contains all migrations including `20260727120000_question_revisions_and_answer_authority`. Migration container runs successfully, applying 14 pending migrations. |

### Why `answer_review_status` was missing

Migration `20260727120000_question_revisions_and_answer_authority.up.sql` adds `answer_review_status` to `questions`. This migration was never applied to the database before the repair because:
1. The first migration run (2026-07-27 01:10) used an old image that stopped at `20260727100000`.
2. The second migration run (2026-07-28 01:35) failed with the "unknown migration" error before reaching `20260727120000`.

### Why the "unknown migration" error appeared

The `gorp_migrations` table contained `20260727100000_create_course_enrollments_table` as applied. The migration container at `2026-07-28 01:35` used the `gkcirclev2-migration` image built on `2026-07-27T00:26:10Z`. This image was built BEFORE the `20260727100000` migration files existed in the repository build context. Therefore, `sql-migrate` encountered a ledger entry it could not find in its embedded `FileMigrationSource` and reported "unknown migration in database."

### Evidence

- Migration image creation timestamp: `2026-07-27T00:26:10Z`
- DB ledger entry for `20260727100000`: applied `2026-07-27 01:10:35+00`
- Migration container log: `Error: Unable to create migration plan because of 20260727100000_create_course_enrollments_table.down.sql: unknown migration in database`
- API image (which WAS rebuilt) contains all 105 migration files through `20260728120000`

---

## Affected Environments

- Local Compose development stack (`gk-circle-pgdata` volume)

## Production Risk

None at this stage — this is a local development database with development data only.

## Correct Repair Path

**Case B (database data must be preserved):** Rebuild the `gkcirclev2-migration` image from the current source, then run `migrate up`. The current source contains all migration files. The migration tool will:
1. Recognize `20260727100000` as a known (applied) migration
2. Apply all subsequent pending migrations in order
3. Result: all 105 migration file pairs recorded and applied

---

## Repair Executed

```bash
docker compose build migration  # Rebuilds migration image from current source
docker compose run --rm migration  # Runs migrate up — applies 14 pending migrations
```

Result: All migrations applied successfully. No data lost.
