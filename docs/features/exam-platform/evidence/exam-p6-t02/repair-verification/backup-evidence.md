# Backup Evidence

Date: 2026-07-28  
Task: EXAM-P6-T02 Migration Reconciliation  

## Backup Details

| Item | Value |
|---|---|
| Backup file | `before-migration-reconciliation-binary.dump` |
| Format | PostgreSQL custom format (`pg_dump --format=custom`) |
| Created at | 2026-07-28T02:25:11+05:30 (local) |
| File size | 122,574 bytes |
| SHA256 | `DB05F4EFA16B6E5B8611A37402B7309F03C8B1B46E1E083DAD0745CC0B4D225C` |
| Database | `gk_circle` |
| DB user | `gk_circle` |
| Options | `--no-owner --no-privileges` |
| Contains | All public schema tables, data, constraints, indexes |

## Database Contents at Backup Time

| Table | Row count |
|---|---|
| questions | 38 |
| assessment_attempts | 41 |
| attempt_answers | 43 |
| course_enrollments | 1 |

The database contained **valuable data** (38 questions, 41 attempts, 43 answers created during development). This data was preserved through the repair.

## Data Preservation Verification (After Repair)

Post-repair query results:

| Table | Pre-repair | Post-repair | Change |
|---|---|---|---|
| questions | 38 | 38 | None |
| assessment_attempts | 41 | 41 | None |
| attempt_answers | 43 | 43 | None |

**All data preserved.**

## Backup Command

```powershell
docker compose exec -T db sh -c "pg_dump --format=custom --no-owner --no-privileges -U gk_circle gk_circle" > before-migration-reconciliation-binary.dump
```

## Restore Command (if needed)

```bash
docker compose exec -T db pg_restore \
  --format=custom \
  --no-owner \
  --no-privileges \
  -U gk_circle \
  -d gk_circle \
  before-migration-reconciliation-binary.dump
```

> [!CAUTION]
> The backup file is stored in the repository root. Do not commit it — it contains database data including potentially sensitive rows. It is in `.gitignore` as `*.dump` (or should be added).
