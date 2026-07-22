# Read-only database summary for the GK Circle v2 NUC stack.

param([string]$NucHost = "nuc")

$ErrorActionPreference = "Stop"
$sql = @'
SELECT current_database() AS database, current_user AS role;
SELECT schemaname, tablename FROM pg_tables WHERE schemaname NOT IN ('pg_catalog', 'information_schema') ORDER BY 1, 2;
'@

$encoded = [Convert]::ToBase64String([Text.Encoding]::UTF8.GetBytes($sql))
ssh $NucHost "cd /home/randhir/apps/gk-circle-v2 && echo $encoded | base64 -d | docker compose --env-file .env -f docker-compose.nuc.yml exec -T db sh -c 'psql -U \"`$POSTGRES_USER\" -d \"`$POSTGRES_DB\"'"
if ($LASTEXITCODE -ne 0) { throw "Database audit failed." }
