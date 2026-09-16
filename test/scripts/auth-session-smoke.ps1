param(
    [ValidateSet("mongodb", "mysql")]
    [string]$Database = "mongodb",
    [int]$ServerPort = 18084,
    [string]$ServerImage = "cashlenx-server:auth-session-smoke",
    [string]$MongoImage = "mongo:7.0",
    [string]$MySqlImage = "mysql:8.0",
    [switch]$SkipBuild
)

$ErrorActionPreference = "Stop"
$serverRoot = (Resolve-Path (Join-Path $PSScriptRoot "..\..")).Path
$runId = "$(Get-Date -Format yyyyMMddHHmmss)-$PID"
$network = "cashlenx-auth-smoke-$runId"
$databaseContainer = "cashlenx-auth-db-$runId"
$serverContainer = "cashlenx-auth-api-$runId"
$databaseName = "cashlenx_auth_smoke_$($runId -replace '-', '_')"
$baseUrl = "http://127.0.0.1:$ServerPort/api/v1"
$adminPassword = "AuthSmokeAdmin456!"
$newPassword = "AuthSmoke456!"

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
        Uri = "$baseUrl$Path"
        Headers = $headers
        UserAgent = "CashLenX-Auth-Smoke"
    }
    if ($null -ne $Body) {
        $parameters.ContentType = "application/json"
        $parameters.Body = $Body | ConvertTo-Json -Depth 8
    }
    return Invoke-RestMethod @parameters
}

function Assert-Unauthorized {
    param([scriptblock]$Action, [string]$FailureMessage)

    try {
        & $Action | Out-Null
        throw $FailureMessage
    } catch {
        if ($null -eq $_.Exception.Response -or $_.Exception.Response.StatusCode.value__ -ne 401) {
            throw
        }
    }
}

try {
    if (-not $SkipBuild) {
        Push-Location $serverRoot
        try {
            $buildTime = (Get-Date).ToUniversalTime().ToString("o")
            docker build --file docker/Dockerfile `
                --build-arg PRODUCT_VERSION=0.0.0-smoke `
                --build-arg GIT_COMMIT=working-tree `
                --build-arg "BUILD_TIME=$buildTime" `
                --tag $ServerImage . | Out-Null
            if ($LASTEXITCODE -ne 0) { throw "server image build failed" }
        } finally {
            Pop-Location
        }
    }

    docker network create $network | Out-Null
    if ($LASTEXITCODE -ne 0) { throw "smoke network creation failed" }

    if ($Database -eq "mongodb") {
        docker run -d --name $databaseContainer --network $network --network-alias database `
            -e MONGO_INITDB_ROOT_USERNAME=cashlenx `
            -e MONGO_INITDB_ROOT_PASSWORD=cashlenx123 `
            -e MONGO_INITDB_DATABASE=$databaseName $MongoImage | Out-Null
        if ($LASTEXITCODE -ne 0) { throw "MongoDB container failed to start" }
        $databaseReady = $false
        $ErrorActionPreference = "Continue"
        for ($attempt = 0; $attempt -lt 240; $attempt++) {
            docker exec $databaseContainer mongosh `
                "mongodb://cashlenx:cashlenx123@localhost:27017/admin?authSource=admin" `
                --quiet --eval 'db.adminCommand({ ping: 1 }).ok' 2>$null | Out-Null
            if ($LASTEXITCODE -eq 0) { $databaseReady = $true; break }
            Start-Sleep -Seconds 1
        }
        $ErrorActionPreference = "Stop"
        if (-not $databaseReady) { throw "MongoDB readiness timed out" }
        $databaseUri = "mongodb://cashlenx:cashlenx123@database:27017/${databaseName}?authSource=admin&retryWrites=false"
    } else {
        $schemaPath = (Resolve-Path (Join-Path $serverRoot "docker\dependencies\mysql\init-mysql.sql")).Path
        docker run -d --name $databaseContainer --network $network --network-alias database `
            -e MYSQL_ROOT_PASSWORD=cashlenx123 -e MYSQL_DATABASE=$databaseName `
            -e MYSQL_USER=cashlenx -e MYSQL_PASSWORD=cashlenx123 `
            -v "${schemaPath}:/docker-entrypoint-initdb.d/001-schema.sql:ro" $MySqlImage | Out-Null
        if ($LASTEXITCODE -ne 0) { throw "MySQL container failed to start" }
        $databaseReady = $false
        $ErrorActionPreference = "Continue"
        for ($attempt = 0; $attempt -lt 300; $attempt++) {
            docker exec $databaseContainer mysqladmin ping -h 127.0.0.1 -uroot -pcashlenx123 --silent 2>$null | Out-Null
            if ($LASTEXITCODE -eq 0) { $databaseReady = $true; break }
            Start-Sleep -Seconds 1
        }
        $ErrorActionPreference = "Stop"
        if (-not $databaseReady) { throw "MySQL readiness timed out" }
        $databaseUri = "cashlenx:cashlenx123@tcp(database:3306)"
    }

    $databaseEnvironment = if ($Database -eq "mongodb") {
        @("-e", "MONGO_DB_URI=$databaseUri")
    } else {
        @("-e", "MYSQL_DB_URI=$databaseUri")
    }
    $runArguments = @(
        "run", "-d", "--name", $serverContainer, "--network", $network,
        "-p", "127.0.0.1:${ServerPort}:10063",
        "-e", "ENV=test", "-e", "SERVER_HOST=0.0.0.0", "-e", "SERVER_PORT=10063",
        "-e", "API_VERSION=v1", "-e", "SCHEMA_VALIDATION=false",
        "-e", "JWT_SECRET=auth-smoke-secret", "-e", "ADMIN_USERNAME=admin",
        "-e", "ADMIN_PASSWORD=$adminPassword", "-e", "DB_TYPE=$Database", "-e", "DB_NAME=$databaseName"
    ) + $databaseEnvironment + @($ServerImage, "-p", "10063")
    docker @runArguments | Out-Null
    if ($LASTEXITCODE -ne 0) { throw "API container failed to start" }

    for ($attempt = 0; $attempt -lt 180; $attempt++) {
        if ((docker inspect --format '{{.State.Running}}' $serverContainer) -ne "true") {
            $logs = docker logs $serverContainer 2>&1
            throw "API container exited before readiness: $logs"
        }
        try {
            Invoke-Api -Method GET -Path "/open/health" | Out-Null
            break
        } catch {
            if ($attempt -eq 179) {
                $logs = docker logs $serverContainer 2>&1
                throw "API readiness timed out: $logs"
            }
            Start-Sleep -Seconds 1
        }
    }

    $login = Invoke-Api -Method POST -Path "/open/auth/login" -Body @{
        username = "admin"; password = $adminPassword; device_id = "auth-smoke"; device_name = "Auth Smoke"
    }
    $accessToken = $login.data.access_token
    $refreshToken = $login.data.refresh_token
    if (-not $accessToken -or -not $refreshToken) { throw "login did not return a complete session" }

    $sessions = Invoke-Api -Method GET -Path "/auth/tokens" -Token $accessToken
    if (@($sessions.data | Where-Object { -not [string]::IsNullOrEmpty($_.token) }).Count -ne 0) {
        throw "session inventory exposed a refresh credential or digest"
    }

    $rotated = Invoke-Api -Method POST -Path "/open/auth/login" -Body @{
        refresh_token = $refreshToken; device_id = "auth-smoke"; device_name = "Auth Smoke"
    }
    $rotatedAccessToken = $rotated.data.access_token
    $rotatedRefreshToken = $rotated.data.refresh_token
    if (-not $rotatedAccessToken -or -not $rotatedRefreshToken -or $rotatedRefreshToken -eq $refreshToken) {
        throw "refresh rotation did not return a new complete session"
    }
    Assert-Unauthorized -FailureMessage "replayed refresh token unexpectedly succeeded" -Action {
        Invoke-Api -Method POST -Path "/open/auth/login" -Body @{
            refresh_token = $refreshToken; device_id = "auth-smoke"; device_name = "Auth Smoke"
        }
    }

    Invoke-Api -Method PUT -Path "/user/password" -Token $rotatedAccessToken -Body @{
        old_password = $adminPassword; new_password = $newPassword
    } | Out-Null
    Assert-Unauthorized -FailureMessage "password change left a refresh session active" -Action {
        Invoke-Api -Method POST -Path "/open/auth/login" -Body @{
            refresh_token = $rotatedRefreshToken; device_id = "auth-smoke"; device_name = "Auth Smoke"
        }
    }

    $afterPasswordChange = Invoke-Api -Method POST -Path "/open/auth/login" -Body @{
        username = "admin"; password = $newPassword; device_id = "auth-smoke"; device_name = "Auth Smoke"
    }
    $logoutRefreshToken = $afterPasswordChange.data.refresh_token
    Invoke-Api -Method POST -Path "/open/auth/logout" -Body @{ refresh_token = $logoutRefreshToken } | Out-Null
    Assert-Unauthorized -FailureMessage "logout left the refresh session active" -Action {
        Invoke-Api -Method POST -Path "/open/auth/login" -Body @{
            refresh_token = $logoutRefreshToken; device_id = "auth-smoke"; device_name = "Auth Smoke"
        }
    }

    if ($Database -eq "mongodb") {
        $invalidStoredTokens = docker exec $databaseContainer mongosh `
            "mongodb://cashlenx:cashlenx123@localhost:27017/$databaseName`?authSource=admin" `
            --quiet --eval 'db.refresh_tokens.countDocuments({token: {$not: /^sha256:[0-9a-f]{64}$/}})'
    } else {
        $invalidStoredTokens = docker exec -e MYSQL_PWD=cashlenx123 $databaseContainer mysql -N -ucashlenx $databaseName `
            -e "SELECT COUNT(*) FROM refresh_tokens WHERE token NOT REGEXP '^sha256:[0-9a-f]{64}$';"
    }
    if ([int]("$invalidStoredTokens".Trim()) -ne 0) { throw "database contains a non-digest refresh credential" }

    Write-Output "Auth session lifecycle smoke passed for $Database"
} finally {
    $previousErrorActionPreference = $ErrorActionPreference
    $ErrorActionPreference = "SilentlyContinue"
    docker rm -f $serverContainer 2>$null | Out-Null
    docker rm -f $databaseContainer 2>$null | Out-Null
    docker network rm $network 2>$null | Out-Null
    $ErrorActionPreference = $previousErrorActionPreference
}
