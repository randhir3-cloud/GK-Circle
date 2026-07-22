# Build and start GK Circle v2 on an isolated NUC port.
# This script does not stop or modify the previous production stack.

param(
    [string]$NucHost = "nuc",
    [string]$NucProject = "/home/randhir/apps/gk-circle-v2",
    [int]$CandidatePort = 3200
)

$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot
Set-Location $root

if ($CandidatePort -eq 3100) {
    throw "Port 3100 is reserved for the current public gateway. Verify on 3200 before cutover."
}
if (git status --porcelain) {
    throw "Working tree must be clean before deployment."
}

$branch = (git branch --show-current).Trim()
$localSha = (git rev-parse HEAD).Trim()
$remoteShaLine = git ls-remote origin "refs/heads/$branch"
if (-not $remoteShaLine) { throw "Branch '$branch' is not published to origin." }
$remoteSha = ($remoteShaLine -split "`t")[0]
if ($localSha -ne $remoteSha) { throw "Local HEAD is not the published branch HEAD." }

$remote = @'
set -eu
project="__PROJECT__"
branch="__BRANCH__"
expected_sha="__SHA__"
candidate_port="__PORT__"

test -d "$project/.git" || { echo "ERROR: run scripts/update-project.ps1 first"; exit 1; }
cd "$project"
test -f .env || { echo "ERROR: missing server-only .env"; exit 1; }
if grep -Eq '(^|=)change[_-]?me' .env; then
  echo "ERROR: .env still contains change_me placeholders"
  exit 1
fi
test -z "$(git status --porcelain)" || { echo "ERROR: NUC checkout is dirty"; exit 1; }

git fetch origin "$branch"
git switch "$branch"
git pull --ff-only origin "$branch"
actual_sha=$(git rev-parse HEAD)
test "$actual_sha" = "$expected_sha" || { echo "ERROR: SHA mismatch"; exit 1; }

export GK_CIRCLE_HTTP_PORT="$candidate_port"
compose='docker compose --env-file .env -f docker-compose.nuc.yml'
$compose config --quiet
$compose up -d --build
$compose ps

for attempt in $(seq 1 30); do
  if curl -fsS "http://127.0.0.1:${candidate_port}/healthz" >/dev/null; then
    echo "CANDIDATE_HEALTHY:http://127.0.0.1:${candidate_port}"
    exit 0
  fi
  sleep 5
done

$compose logs --tail 100 gateway api web kratos
echo "ERROR: candidate health check timed out"
exit 1 # Keep PowerShell's pipeline CR outside shell syntax.
'@.Replace('__PROJECT__', $NucProject).Replace('__BRANCH__', $branch).Replace('__SHA__', $localSha).Replace('__PORT__', $CandidatePort.ToString())

$remote.Replace("`r", "") | ssh $NucHost "bash -s"
if ($LASTEXITCODE -ne 0) { throw "NUC candidate deployment failed." }

Write-Host "GK Circle v2 candidate is running on NUC port $CandidatePort."
Write-Host "Run scripts/verify-nuc-deployment.ps1 before any public cutover."
