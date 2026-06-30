param([string]$MySqlImage = "mysql:8.0")

$ErrorActionPreference = "Continue"
$repoRoot = Split-Path -Parent $PSScriptRoot
$container = "cashlenx-mysql-migrations-$(Get-Date -Format yyyyMMddHHmmss)-$PID"

try {
    docker run -d --name $container `
        -e MYSQL_ROOT_PASSWORD=cashlenx123 $MySqlImage | Out-Null
    if ($LASTEXITCODE -ne 0) { throw "MySQL container failed to start" }

    $ready = $false
    for ($attempt = 0; $attempt -lt 90; $attempt++) {
        docker exec $container mysql -uroot -pcashlenx123 `
            -e 'SELECT 1' 2>$null | Out-Null
        if ($LASTEXITCODE -eq 0) { $ready = $true; break }
        Start-Sleep -Seconds 1
    }
    if (-not $ready) { throw "MySQL authenticated readiness timed out" }

    $migrationsPath = Resolve-Path (Join-Path $repoRoot "migrations")
    docker cp $migrationsPath "${container}:/tmp/migrations" | Out-Null
    if ($LASTEXITCODE -ne 0) { throw "Failed to copy migrations" }

    Get-ChildItem $migrationsPath -Filter *.sql | Sort-Object Name | ForEach-Object {
        Write-Host "Applying $($_.Name)..."
        docker exec $container sh -c `
            "mysql -uroot -pcashlenx123 < /tmp/migrations/$($_.Name)" 2>$null | Out-Null
        if ($LASTEXITCODE -ne 0) { throw "Migration failed: $($_.Name)" }
    }

    $tables = docker exec $container mysql -uroot -pcashlenx123 -N `
        -e "SELECT table_name FROM information_schema.tables WHERE table_schema='cashlenx' ORDER BY table_name" 2>$null
    $expected = @("cash_flows", "categories", "operation_confirm_codes", "refresh_tokens", "user_configurations", "users")
    foreach ($table in $expected) {
        if ($tables -notcontains $table) { throw "Missing migrated table: $table" }
    }
    Write-Host "MySQL migration smoke completed successfully."
} finally {
    docker rm -f $container 2>$null | Out-Null
}
