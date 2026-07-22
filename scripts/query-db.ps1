param (
    [string]$Query
)
ssh nuc "docker exec gk-circle-postgres psql -U gk_user -d gk_circle -c `"$Query`""
