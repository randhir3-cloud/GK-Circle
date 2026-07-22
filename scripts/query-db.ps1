# Execute a read-only SQL query against the GK Circle v2 NUC database.

param(
    [Parameter(Mandatory = $true)][string]$Query,
    [string]$NucHost = "nuc"
)

$ErrorActionPreference = "Stop"
if ($Query -notmatch '^\s*(SELECT|WITH|SHOW|EXPLAIN)\b') {
    throw "Only SELECT, WITH, SHOW, or EXPLAIN statements are allowed."
}
if ($Query -match '(?i)\b(INSERT|UPDATE|DELETE|DROP|ALTER|TRUNCATE|CREATE|GRANT|REVOKE|COPY)\b') {
    throw "Mutating SQL is prohibited by this read-only helper."
}

$encoded = [Convert]::ToBase64String([Text.Encoding]::UTF8.GetBytes($Query))
ssh $NucHost "cd /home/randhir/apps/gk-circle-v2 && echo $encoded | base64 -d | docker compose --env-file .env -f docker-compose.nuc.yml exec -T db sh -c 'psql -v ON_ERROR_STOP=1 -U \"`$POSTGRES_USER\" -d \"`$POSTGRES_DB\"'"
if ($LASTEXITCODE -ne 0) { throw "Query failed." }
