# Deploy the isolated GK Circle v2 candidate to the NUC.

param(
    [int]$CandidatePort = 3200
)

$ErrorActionPreference = "Stop"
& (Join-Path $PSScriptRoot "deploy-nuc-remote.ps1") -CandidatePort $CandidatePort
