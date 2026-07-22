# Publish the current committed branch and update the isolated NUC checkout.

param(
    [string]$NucHost = "nuc",
    [string]$NucProject = "/home/randhir/apps/gk-circle-v2"
)

$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot
Set-Location $root

$expectedOrigin = "https://github.com/randhir3-cloud/GK-Circle-v2.git"
$origin = (git remote get-url origin).Trim()
if ($origin -ne $expectedOrigin) {
    throw "Unexpected origin '$origin'. Expected '$expectedOrigin'."
}

if (git status --porcelain) {
    throw "Working tree is not clean. Review and commit changes before publishing."
}

$branch = (git branch --show-current).Trim()
if (-not $branch) { throw "Detached HEAD is not deployable." }

git push -u origin $branch
if ($LASTEXITCODE -ne 0) { throw "Git push failed." }

$remote = @'
set -eu
project="__PROJECT__"
branch="__BRANCH__"
repo="https://github.com/randhir3-cloud/GK-Circle-v2.git"

if [ ! -d "$project/.git" ]; then
  mkdir -p "$(dirname "$project")"
  git clone --branch "$branch" --single-branch "$repo" "$project"
else
  cd "$project"
  test -z "$(git status --porcelain)" || { echo "ERROR: NUC checkout is dirty"; exit 1; }
  git fetch origin "$branch"
  git switch "$branch"
  git pull --ff-only origin "$branch"
fi

cd "$project"
git rev-parse --short HEAD
if [ ! -f .env ]; then
  echo "ACTION_REQUIRED: create $project/.env from .env.example and replace every change_me value"
fi
exit 0 # Keep PowerShell's pipeline CR outside shell syntax.
'@.Replace('__PROJECT__', $NucProject).Replace('__BRANCH__', $branch)

$remote.Replace("`r", "") | ssh $NucHost "bash -s"
if ($LASTEXITCODE -ne 0) { throw "NUC checkout update failed." }

Write-Host "Published origin/$branch and updated $NucHost`:$NucProject"
