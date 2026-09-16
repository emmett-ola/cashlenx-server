$pinFile = Join-Path $PSScriptRoot "..\..\docker\dependencies\images.env"

function Assert-DependencyImageReference {
    param(
        [string]$Image,
        [ValidateSet("mongo", "mysql")]
        [string]$Repository
    )
    if ($Image -notmatch "^$Repository`:[A-Za-z0-9_.-]+@sha256:[0-9a-f]{64}$") {
        throw "$Repository test image must be digest-pinned."
    }
}

function Get-DependencyImagePin {
    param(
        [ValidateSet("MONGO_IMAGE", "MYSQL_IMAGE")]
        [string]$Key
    )
    $line = Get-Content -LiteralPath $pinFile |
        Where-Object { $_ -match "^$([regex]::Escape($Key))=" } |
        Select-Object -Last 1
    if (-not $line) { throw "$Key is missing from docker/dependencies/images.env." }
    $image = ($line -split "=", 2)[1].Trim()
    $repository = if ($Key -eq "MONGO_IMAGE") { "mongo" } else { "mysql" }
    Assert-DependencyImageReference -Image $image -Repository $repository
    return $image
}

$PinnedMongoImage = Get-DependencyImagePin -Key MONGO_IMAGE
$PinnedMySqlImage = Get-DependencyImagePin -Key MYSQL_IMAGE
