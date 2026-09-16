param(
    [ValidateSet("all", "mongodb", "mysql")]
    [string]$Database = "all"
)

$ErrorActionPreference = "Stop"

$repoPath = (Resolve-Path (Join-Path $PSScriptRoot "..\..")).Path
. (Join-Path $PSScriptRoot "dependency-image-pins.ps1")
$backupBase = Join-Path $repoPath "backups"
$runId = ([Guid]::NewGuid().ToString("N").Substring(0, 12))
$runRoot = Join-Path $backupBase "smoke-$runId"
$envFileName = ".env.data-protection-smoke-$runId"
$envFilePath = Join-Path $repoPath $envFileName
$keyFileName = ".env.data-protection-key-$runId"
$keyFilePath = Join-Path $repoPath $keyFileName
$bashPath = "C:\Program Files\Git\bin\bash.exe"
$sourceContainer = $null

if (-not (Test-Path -LiteralPath $bashPath -PathType Leaf)) {
    throw "Git Bash is required at $bashPath"
}
if (-not (Get-Command docker -ErrorAction SilentlyContinue)) {
    throw "Docker is required."
}

function Assert-CommandSucceeded {
    param([string]$Operation)
    if ($LASTEXITCODE -ne 0) {
        throw "$Operation failed with exit code $LASTEXITCODE."
    }
}

function Wait-MongoDB {
    param([string]$Container, [string]$Password)
    for ($attempt = 0; $attempt -lt 60; $attempt++) {
        & docker exec $Container mongosh --quiet --username root --password $Password --authenticationDatabase admin --eval "quit(db.runCommand({ ping: 1 }).ok ? 0 : 1)" "127.0.0.1:27017/cashlenx" 2>$null
        if ($LASTEXITCODE -eq 0) { return }
        Start-Sleep -Seconds 2
    }
    throw "Disposable MongoDB did not become ready."
}

function Wait-MySQL {
    param([string]$Container, [string]$Password)
    for ($attempt = 0; $attempt -lt 120; $attempt++) {
        & docker exec --env "MYSQL_PWD=$Password" $Container mysqladmin ping --host=127.0.0.1 --user=root --silent 2>$null
        if ($LASTEXITCODE -eq 0) { return }
        Start-Sleep -Seconds 2
    }
    throw "Disposable MySQL did not become ready."
}

function Invoke-ProfileSmoke {
    param([ValidateSet("mongodb", "mysql")][string]$Engine)

    $password = "smoke-$runId-$Engine"
    $script:sourceContainer = "cashlenx-backup-source-$Engine-$runId"
    $image = if ($Engine -eq "mongodb") { $PinnedMongoImage } else { $PinnedMySqlImage }

    & docker image inspect $image *> $null
    Assert-CommandSucceeded "Inspect $image"

    if ($Engine -eq "mongodb") {
        & docker run --detach --rm --network none --name $script:sourceContainer `
            --env MONGO_INITDB_ROOT_USERNAME=root `
            --env "MONGO_INITDB_ROOT_PASSWORD=$password" `
            --env MONGO_INITDB_DATABASE=cashlenx `
            $image *> $null
        Assert-CommandSucceeded "Start disposable MongoDB"
        Wait-MongoDB -Container $script:sourceContainer -Password $password
        & docker exec $script:sourceContainer mongosh --quiet --username root --password $password --authenticationDatabase admin `
            --eval "const d=db.getSiblingDB('cashlenx'); d.schema_migrations.insertOne({version:1,name:'smoke',checksum:'smoke',dirty:false}); d.smoke.insertOne({value:'retained'});" `
            "127.0.0.1:27017/cashlenx" *> $null
        Assert-CommandSucceeded "Seed disposable MongoDB"
    }
    else {
        & docker run --detach --rm --network none --name $script:sourceContainer `
            --tmpfs /var/lib/mysql:rw,noexec,nosuid,size=1g `
            --env "MYSQL_ROOT_PASSWORD=$password" `
            --env MYSQL_DATABASE=cashlenx `
            $image *> $null
        Assert-CommandSucceeded "Start disposable MySQL"
        Wait-MySQL -Container $script:sourceContainer -Password $password
        & docker exec --env "MYSQL_PWD=$password" $script:sourceContainer mysql --host=127.0.0.1 --user=root cashlenx `
            --execute="CREATE TABLE schema_migrations (version BIGINT PRIMARY KEY, name VARCHAR(255), checksum VARCHAR(64), dirty BOOLEAN); INSERT INTO schema_migrations VALUES (1, 'smoke', 'smoke', false); CREATE TABLE smoke (id INT PRIMARY KEY, value VARCHAR(32)); INSERT INTO smoke VALUES (1, 'retained');" *> $null
        Assert-CommandSucceeded "Seed disposable MySQL"
    }

    $settings = @(
        "DB_TYPE=$Engine"
        "DB_NAME=cashlenx"
        "MONGO_CONTAINER_NAME=$script:sourceContainer"
        "MYSQL_CONTAINER_NAME=$script:sourceContainer"
        "BACKUP_ROOT=./backups/smoke-$runId/$Engine"
        "BACKUP_ENCRYPTION_KEY_FILE=./$keyFileName"
        "BACKUP_MIN_FREE_MIB=1"
        "BACKUP_DAILY_RETENTION=1"
        "BACKUP_WEEKLY_RETENTION=1"
        "BACKUP_MONTHLY_RETENTION=1"
        "BACKUP_MAX_AGE_HOURS=26"
        "RESTORE_DRILL_EVIDENCE_DIR="
    )
    Set-Content -LiteralPath $envFilePath -Value $settings -Encoding utf8NoBOM

    Push-Location $repoPath
    try {
        $previousEnvFile = $env:ENV_FILE
        $env:ENV_FILE = $envFileName
        & $bashPath scripts/data-protection/backup.sh daily
        Assert-CommandSucceeded "$Engine encrypted backup"

        $tierPath = Join-Path $runRoot "$Engine\daily"
        $artifacts = @(Get-ChildItem -LiteralPath $tierPath -Filter "*.tar.gz.enc")
        if ($artifacts.Count -ne 1) {
            throw "$Engine backup produced $($artifacts.Count) artifacts; expected 1."
        }
        $artifact = $artifacts[0]
        & $bashPath scripts/data-protection/restore-drill.sh $artifact.FullName
        Assert-CommandSucceeded "$Engine disposable restore drill"

        $evidencePath = Join-Path $runRoot "$Engine\restore-drills"
        if (-not (Get-ChildItem -LiteralPath $evidencePath -Filter "*.json" -ErrorAction SilentlyContinue)) {
            throw "$Engine restore drill did not create evidence."
        }

        $corruptPath = Join-Path $runRoot "$Engine\corrupt"
        New-Item -ItemType Directory -Path $corruptPath -Force | Out-Null
        $corruptArtifact = Join-Path $corruptPath $artifact.Name
        Copy-Item -LiteralPath $artifact.FullName -Destination $corruptArtifact
        Copy-Item -LiteralPath "$($artifact.FullName).sha256" -Destination "$corruptArtifact.sha256"
        $bytes = [System.IO.File]::ReadAllBytes($corruptArtifact)
        $bytes[$bytes.Length - 1] = $bytes[$bytes.Length - 1] -bxor 0x01
        [System.IO.File]::WriteAllBytes($corruptArtifact, $bytes)
        & $bashPath scripts/data-protection/restore-drill.sh $corruptArtifact *> $null
        if ($LASTEXITCODE -eq 0) {
            throw "$Engine corrupt backup unexpectedly passed restore validation."
        }

        Start-Sleep -Seconds 1
        & $bashPath scripts/data-protection/backup.sh daily
        Assert-CommandSucceeded "$Engine retention backup"
        $retained = @(Get-ChildItem -LiteralPath $tierPath -Filter "*.tar.gz.enc")
        if ($retained.Count -ne 1) {
            throw "$Engine retention kept $($retained.Count) artifacts; expected 1."
        }
    }
    finally {
        $env:ENV_FILE = $previousEnvFile
        Pop-Location
        & docker rm -f $script:sourceContainer *> $null
        $script:sourceContainer = $null
    }
}

try {
    Set-Content -LiteralPath $keyFilePath -Value "smoke-key-$runId" -Encoding utf8NoBOM
    $targets = if ($Database -eq "all") { @("mongodb", "mysql") } else { @($Database) }
    foreach ($target in $targets) {
        Invoke-ProfileSmoke -Engine $target
    }
    Write-Host "Data-protection smoke passed: $($targets -join ', ')"
}
finally {
    if ($sourceContainer) {
        & docker rm -f $sourceContainer *> $null
    }
    Remove-Item -LiteralPath $envFilePath -Force -ErrorAction SilentlyContinue
    Remove-Item -LiteralPath $keyFilePath -Force -ErrorAction SilentlyContinue

    $resolvedRepo = [System.IO.Path]::GetFullPath($repoPath).TrimEnd([System.IO.Path]::DirectorySeparatorChar) + [System.IO.Path]::DirectorySeparatorChar
    $resolvedRunRoot = [System.IO.Path]::GetFullPath($runRoot)
    if ($resolvedRunRoot.StartsWith($resolvedRepo, [System.StringComparison]::OrdinalIgnoreCase) -and (Test-Path -LiteralPath $resolvedRunRoot)) {
        Remove-Item -LiteralPath $resolvedRunRoot -Recurse -Force
    }
}
