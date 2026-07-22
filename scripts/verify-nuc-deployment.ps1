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
$compose exec -T db sh -c 'pg_isready -U "$POSTGRES_USER" -d "$POSTGRES_DB"'
$compose exec -T redis sh -c 'redis-cli -a "$REDIS_PASSWORD" ping' | grep PONG
curl -fsS "http://127.0.0.1:__PORT__/healthz" >/dev/null
curl -fsS "http://127.0.0.1:__PORT__/" >/dev/null
curl -fsS "http://127.0.0.1:__PORT__/kratos/health/ready" >/dev/null
echo "VERIFY_OK"
'@.Replace('__PROJECT__', $NucProject).Replace('__PORT__', $CandidatePort.ToString())

$output = $remote.Replace("`r", "") | ssh $NucHost "bash -s"
if ($LASTEXITCODE -ne 0 -or $output -notcontains "VERIFY_OK") {
    throw "NUC verification failed.`n$($output -join "`n")"
}

$output
Write-Host "Automated candidate verification passed. Complete registration/login/test smoke verification before cutover."
