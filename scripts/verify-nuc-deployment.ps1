# Verify the isolated GK Circle v2 NUC candidate.

param(
    [string]$NucHost = "nuc",
    [string]$NucProject = "/home/randhir/apps/gk-circle-v2",
    [int]$CandidatePort = 3200
)

$ErrorActionPreference = "Stop"

$remote = @'
set -eu
cd "__PROJECT__"
export GK_CIRCLE_HTTP_PORT="__PORT__"
compose='docker compose --env-file .env -f docker-compose.nuc.yml'
$compose config --quiet
$compose ps
test -z "$($compose ps --status unhealthy -q)"
$compose exec -T db sh -c 'pg_isready -U "$POSTGRES_USER" -d "$POSTGRES_DB"' < /dev/null
$compose exec -T redis sh -c 'redis-cli -a "$REDIS_PASSWORD" ping' < /dev/null | grep PONG
curl -fsS "http://127.0.0.1:__PORT__/healthz" >/dev/null
curl -fsS "http://127.0.0.1:__PORT__/" | grep -F "GK Circle" >/dev/null
$compose exec -T kratos wget -qO- http://localhost:4434/health/ready < /dev/null >/dev/null
echo "VERIFY_OK"
exit 0 # Keep PowerShell's pipeline CR outside shell syntax.
'@.Replace('__PROJECT__', $NucProject).Replace('__PORT__', $CandidatePort.ToString())

$output = $remote.Replace("`r", "") | ssh $NucHost "bash -s"
if ($LASTEXITCODE -ne 0 -or $output -notcontains "VERIFY_OK") {
    throw "NUC verification failed.`n$($output -join "`n")"
}

$output
Write-Host "Automated candidate verification passed. Complete registration/login/test smoke verification before cutover."
