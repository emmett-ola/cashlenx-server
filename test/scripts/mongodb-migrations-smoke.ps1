param([string]$MongoImage = "")

$ErrorActionPreference = "Stop"
$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..\..")).Path
. (Join-Path $PSScriptRoot "dependency-image-pins.ps1")
if (-not $MongoImage) { $MongoImage = $PinnedMongoImage }
Assert-DependencyImageReference -Image $MongoImage -Repository mongo
$container = "cashlenx-mongodb-migrations-$(Get-Date -Format yyyyMMddHHmmss)-$PID"

try {
    docker run -d --name $container -p 127.0.0.1::27017 `
        -e MONGO_INITDB_ROOT_USERNAME=cashlenx `
        -e MONGO_INITDB_ROOT_PASSWORD=cashlenx123 $MongoImage | Out-Null
    if ($LASTEXITCODE -ne 0) { throw "MongoDB container failed to start" }

    $ready = $false
    $ErrorActionPreference = "Continue"
    for ($attempt = 0; $attempt -lt 240; $attempt++) {
        docker exec $container mongosh `
            "mongodb://cashlenx:cashlenx123@localhost:27017/admin?authSource=admin" `
            --quiet --eval 'db.adminCommand({ ping: 1 }).ok' 2>$null | Out-Null
        if ($LASTEXITCODE -eq 0) { $ready = $true; break }
        Start-Sleep -Seconds 1
    }
    $ErrorActionPreference = "Stop"
    if (-not $ready) { throw "MongoDB readiness timed out" }

    $mappedPort = ((docker port $container 27017/tcp) -split ':')[-1].Trim()
    $env:MONGO_TEST_URI = "mongodb://cashlenx:cashlenx123@127.0.0.1:$mappedPort/admin?authSource=admin&retryWrites=false"
    Push-Location $repoRoot
    try {
        go test -tags=integration -run 'TestMongo.*Integration' -v ./migrations
        if ($LASTEXITCODE -ne 0) { throw "MongoDB migration integration tests failed" }
    } finally {
        Pop-Location
    }
    Write-Output "MongoDB migration smoke completed successfully."
} finally {
    Remove-Item Env:MONGO_TEST_URI -ErrorAction SilentlyContinue
    $previousErrorActionPreference = $ErrorActionPreference
    $ErrorActionPreference = "SilentlyContinue"
    docker rm -f $container 2>$null | Out-Null
    $ErrorActionPreference = $previousErrorActionPreference
}
