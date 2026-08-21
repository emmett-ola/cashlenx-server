param(
    [ValidateSet("mongodb", "mysql")]
    [string]$Database = "mongodb",
    [int]$ServerPort = 18082,
    [string]$MongoImage = "mongo:7.0",
    [string]$MySqlImage = "mysql:8.0"
)

$ErrorActionPreference = "Stop"
$serverRoot = Resolve-Path (Join-Path $PSScriptRoot "..")
$runId = "$(Get-Date -Format yyyyMMddHHmmss)-$PID"
$container = "cashlenx-budget-smoke-$Database-$runId"
$dbName = if ($Database -eq "mongodb") { "cashlenx_budget_smoke_$($runId -replace '-', '_')" } else { "cashlenx" }
$tempDir = Join-Path $env:TEMP "cashlenx-budget-smoke-$runId"
$serverExe = Join-Path $tempDir "cashlenx-server.exe"
$stdoutLog = Join-Path $tempDir "server.out.log"
$stderrLog = Join-Path $tempDir "server.err.log"
$server = $null

function Invoke-Api {
    param(
        [string]$Method,
        [string]$Path,
        [object]$Body,
        [string]$Token
    )
    $headers = @{}
    if ($Token) { $headers.Authorization = "Bearer $Token" }
    $parameters = @{
        Method = $Method
        Uri = "http://127.0.0.1:$ServerPort/api/v0$Path"
        Headers = $headers
    }
    if ($null -ne $Body) {
        $parameters.ContentType = "application/json"
        $parameters.Body = $Body | ConvertTo-Json -Depth 8
    }
    return Invoke-RestMethod @parameters
}

New-Item -ItemType Directory -Path $tempDir | Out-Null

try {
    Push-Location $serverRoot
    try {
        go build -o $serverExe .
        if ($LASTEXITCODE -ne 0) { throw "server build failed" }
    } finally {
        Pop-Location
    }

    if ($Database -eq "mongodb") {
        docker run -d --name $container -p 127.0.0.1::27017 `
            -e MONGO_INITDB_ROOT_USERNAME=cashlenx `
            -e MONGO_INITDB_ROOT_PASSWORD=cashlenx123 `
            -e MONGO_INITDB_DATABASE=$dbName $MongoImage | Out-Null
        if ($LASTEXITCODE -ne 0) { throw "MongoDB container failed to start" }
        $dbPort = ((docker port $container 27017/tcp) -split ':')[-1].Trim()
        $ErrorActionPreference = "Continue"
        for ($attempt = 0; $attempt -lt 60; $attempt++) {
            docker exec $container mongosh `
                "mongodb://cashlenx:cashlenx123@localhost:27017/admin?authSource=admin" `
                --quiet --eval 'db.adminCommand({ ping: 1 }).ok' 2>$null | Out-Null
            if ($LASTEXITCODE -eq 0) { break }
            if ($attempt -eq 59) { throw "MongoDB readiness timed out" }
            Start-Sleep -Seconds 1
        }
        $ErrorActionPreference = "Stop"
    } else {
        docker run -d --name $container -p 127.0.0.1::3306 `
            -e MYSQL_ROOT_PASSWORD=cashlenx123 -e MYSQL_DATABASE=$dbName `
            -e MYSQL_USER=cashlenx -e MYSQL_PASSWORD=cashlenx123 $MySqlImage | Out-Null
        if ($LASTEXITCODE -ne 0) { throw "MySQL container failed to start" }
        $dbPort = ((docker port $container 3306/tcp) -split ':')[-1].Trim()
        $ErrorActionPreference = "Continue"
        for ($attempt = 0; $attempt -lt 90; $attempt++) {
            docker exec $container mysql -uroot -pcashlenx123 -e 'SELECT 1' 2>$null | Out-Null
            if ($LASTEXITCODE -eq 0) { break }
            if ($attempt -eq 89) { throw "MySQL readiness timed out" }
            Start-Sleep -Seconds 1
        }
        $ErrorActionPreference = "Stop"
        $schemaPath = Resolve-Path (Join-Path $serverRoot "docker\mysql\init-mysql.sql")
        docker cp $schemaPath "${container}:/tmp/init-mysql.sql" | Out-Null
        $ErrorActionPreference = "Continue"
        docker exec $container sh -c 'mysql -uroot -pcashlenx123 cashlenx < /tmp/init-mysql.sql' 2>$null | Out-Null
        $schemaExitCode = $LASTEXITCODE
        $ErrorActionPreference = "Stop"
        if ($schemaExitCode -ne 0) { throw "MySQL schema initialization failed" }
    }

    $env:ENV = "test"
    $env:SERVER_HOST = "127.0.0.1"
    $env:API_VERSION = "v0"
    $env:SCHEMA_VALIDATION = "true"
    $env:JWT_SECRET = "budget-smoke-secret"
    $env:ADMIN_USERNAME = "admin"
    $env:ADMIN_PASSWORD = "admin"
    $env:DB_TYPE = $Database
    $env:DB_NAME = $dbName
    if ($Database -eq "mongodb") {
        $env:MONGO_DB_URI = "mongodb://cashlenx:cashlenx123@localhost:$dbPort/${dbName}?authSource=admin&retryWrites=false"
    } else {
        $env:MYSQL_DB_URI = "cashlenx:cashlenx123@tcp(localhost:$dbPort)"
    }

    $server = Start-Process -FilePath $serverExe `
        -ArgumentList @("open", "start", "-p", "$ServerPort") `
        -WorkingDirectory $serverRoot -WindowStyle Hidden -PassThru `
        -RedirectStandardOutput $stdoutLog -RedirectStandardError $stderrLog

    for ($attempt = 0; $attempt -lt 90; $attempt++) {
        if ($server.HasExited) { throw "API server exited: $([IO.File]::ReadAllText($stderrLog))" }
        try {
            Invoke-Api -Method GET -Path "/open/health" | Out-Null
            break
        } catch {
            if ($attempt -eq 89) { throw "API readiness timed out: $([IO.File]::ReadAllText($stderrLog))" }
            Start-Sleep -Seconds 1
        }
    }

    try {
        Invoke-Api -Method GET -Path "/budget?period=2026-08" | Out-Null
        throw "unauthenticated budget list unexpectedly succeeded"
    } catch {
        if ($_.Exception.Response.StatusCode.value__ -ne 401) { throw }
    }

    $login = Invoke-Api -Method POST -Path "/open/auth/login" -Body @{ username = "admin"; password = "admin" }
    $token = $login.data.access_token
    if (-not $token) { throw "admin login did not return an access token" }

    $profile = Invoke-Api -Method PUT -Path "/user/profile" -Token $token -Body @{
        nickname = "Budget Admin"; gender = "others"; phone_number = "+65 6123 4567"
        location = "Singapore"; birth_date = "1992-08-21"
    }
    if ($profile.data.phone_number -ne "+65 6123 4567" -or $profile.data.location -ne "Singapore" -or $profile.data.birth_date -ne "1992-08-21") {
        throw "extended profile update did not persist: $($profile | ConvertTo-Json -Depth 8 -Compress)"
    }
    $profileRead = Invoke-Api -Method GET -Path "/user/profile" -Token $token
    if ($profileRead.data.phone_number -ne "+65 6123 4567" -or $profileRead.data.location -ne "Singapore" -or $profileRead.data.birth_date -ne "1992-08-21") {
        throw "extended profile read did not match update: $($profileRead | ConvertTo-Json -Depth 8 -Compress)"
    }

    $configuration = Invoke-Api -Method PUT -Path "/user/configuration" -Token $token -Body @{
        display_language = "zh-Hans"; currency_code = "SGD"; active_theme_color = "#008080"
    }
    if ($configuration.data.display_language -ne "zh-Hans" -or $configuration.data.currency_code -ne "SGD" -or $configuration.data.active_theme_color -ne "#008080") {
        throw "configuration update did not persist: $($configuration | ConvertTo-Json -Depth 8 -Compress)"
    }
    $configurationRead = Invoke-Api -Method GET -Path "/user/configuration" -Token $token
    if ($configurationRead.data.display_language -ne "zh-Hans" -or $configurationRead.data.currency_code -ne "SGD" -or $configurationRead.data.active_theme_color -ne "#008080") {
        throw "configuration read did not match update: $($configurationRead | ConvertTo-Json -Depth 8 -Compress)"
    }

    $category = Invoke-Api -Method POST -Path "/category" -Token $token -Body @{
        name = "Budget Smoke"; type = "expense"; remark = "disposable budget smoke"
    }
    $categoryId = if ($category.data.id) { $category.data.id } else { $category.data.Id }
    if (-not $categoryId) { throw "category create did not return an ID" }

    $budget = Invoke-Api -Method POST -Path "/budget" -Token $token -Body @{
        category_id = $categoryId; period = "2026-08"; limit_amount = 500
    }
    $budgetId = $budget.data.id
    if (-not $budgetId) { throw "budget create did not return an ID" }

    Invoke-Api -Method POST -Path "/cash/expense" -Token $token -Body @{
        belongs_date = "20260821"; category_name = "Budget Smoke"; amount = 125.5; description = "Budget smoke expense"
    } | Out-Null

    $listed = Invoke-Api -Method GET -Path "/budget?period=2026-08" -Token $token
    $listData = if ($listed.data -is [System.Array]) { $listed.data } elseif ($null -ne $listed.data.data) { $listed.data.data } else { $listed.data }
    $item = @($listData) | Where-Object { $_.id -eq $budgetId }
    if (@($item).Count -ne 1) { throw "created budget $budgetId was not listed: $($listed | ConvertTo-Json -Depth 8 -Compress)" }
    if ([math]::Abs([double]$item.spent_amount - 125.5) -gt 0.001) { throw "spent amount was not derived from cash flows: $($item | ConvertTo-Json -Depth 8 -Compress)" }

    $updated = Invoke-Api -Method PUT -Path "/budget/$budgetId" -Token $token -Body @{
        category_id = $categoryId; period = "2026-08"; limit_amount = 750
    }
    if ([double]$updated.data.limit_amount -ne 750) { throw "budget update did not persist" }

    Invoke-Api -Method DELETE -Path "/budget/$budgetId" -Token $token | Out-Null
    $afterDelete = Invoke-Api -Method GET -Path "/budget?period=2026-08" -Token $token
    $afterDeleteData = if ($afterDelete.data -is [System.Array]) { $afterDelete.data } elseif ($null -ne $afterDelete.data.data) { $afterDelete.data.data } else { $afterDelete.data }
    if (@($afterDeleteData | Where-Object { $_.id -eq $budgetId }).Count -ne 0) { throw "deleted budget remained visible" }

    Write-Output "Budget smoke passed for $Database"
} finally {
    if ($server -and -not $server.HasExited) {
        Stop-Process -Id $server.Id -Force -ErrorAction SilentlyContinue
    }
    docker rm -f $container 2>$null | Out-Null
    if (Test-Path $tempDir) {
        Remove-Item -LiteralPath $tempDir -Recurse -Force -ErrorAction SilentlyContinue
    }
}
