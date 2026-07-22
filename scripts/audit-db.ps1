$ErrorActionPreference = "Stop"

function Invoke-NucBash([string]$Script) {
    $lf = $Script.Replace("`r", "")
    $tmp = [System.IO.Path]::GetTempFileName()
    try {
        [System.IO.File]::WriteAllText($tmp, $lf, (New-Object System.Text.UTF8Encoding $false))
        $out = cmd /c "type `"$tmp`" | ssh nuc bash -s" 2>&1
    } finally {
        Remove-Item $tmp -ErrorAction SilentlyContinue
    }
    if ($LASTEXITCODE -ne 0) {
        throw "Remote bash failed (exit $LASTEXITCODE): $out"
    }
    return ($out | Out-String).Trim()
}

$query = @'
docker exec gk-circle-postgres psql -U gk_user -d gk_circle -c '
SELECT u.email, ca.status 
FROM users u 
LEFT JOIN creator_applications ca ON u.id = ca."userId" 
WHERE u.email LIKE '\''%mq7a6bf0%'\'' 
ORDER BY u.email;
'
'@

$out = Invoke-NucBash $query
Write-Host $out
