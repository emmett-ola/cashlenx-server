#!/usr/bin/env pwsh

<#
.SYNOPSIS
CashLenX Docker Bootstrap Script
This script helps you easily start the CashLenX application with Docker

.DESCRIPTION
This PowerShell script provides a bootstrap experience for starting CashLenX services using Docker Compose.
It includes service selection, environment file management, and service status display.

.EXAMPLE
.start.ps1
Starts the script in interactive mode

.EXAMPLE
$env:ENABLE_SERVICES="mongodb,backend" ; .start.ps1
Starts services specified in the environment variable
#>

# Define colors for output
$RED = "`x1B[0;31m"
$GREEN = "`x1B[0;32m"
$YELLOW = "`x1B[1;33m"
$BLUE = "`x1B[0;34m"
$NC = "`x1B[0m" # No Color

# Print colored output
function Print-Info {
    param([string]$Message)
    Write-Output "${BLUE}ℹ${NC} $Message"
}

function Print-Success {
    param([string]$Message)
    Write-Output "${GREEN}✓${NC} $Message"
}

function Print-Warning {
    param([string]$Message)
    Write-Output "${YELLOW}⚠${NC} $Message"
}

function Print-Error {
    param([string]$Message)
    Write-Output "${RED}✗${NC} $Message"
}

function Print-Header {
    param([string]$Title)
    Write-Output "`n${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    Write-Output "${BLUE}  $Title${NC}"
    Write-Output "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}`n"
}

# Check if Docker is installed
function Check-Docker {
    if (-not (Get-Command docker -ErrorAction SilentlyContinue)) {
        Print-Error "Docker is not installed. Please install Docker first."
        exit 1
    }

    $composeCmd = Get-Command docker-compose -ErrorAction SilentlyContinue
    $composeSubcommand = docker compose version 2>&1
    if (-not $composeCmd -and $composeSubcommand -match "docker: 'compose' is not a docker command") {
        Print-Error "Docker Compose is not installed. Please install Docker Compose first."
        exit 1
    }

    Print-Success "Docker and Docker Compose are installed"
}

# Check if .env file exists
function Check-EnvFile {
    if (-not (Test-Path -Path .env -PathType Leaf)) {
        Print-Warning ".env file not found"
        Print-Info "Creating .env from .env.sample..."

        if (Test-Path -Path .env.sample -PathType Leaf) {
            Copy-Item -Path .env.sample -Destination .env
            Print-Success ".env file created from .env.sample"
            Print-Warning "Please review and update .env with your configuration"
        } else {
            Print-Error ".env.sample file not found"
            exit 1
        }
    } else {
        Print-Success ".env file exists"
    }
}

# Load environment variables from .env
function Load-Env {
    if (Test-Path -Path .env -PathType Leaf) {
        $envContent = Get-Content -Path .env | Where-Object { $_ -notmatch '^#' -and $_ -notmatch '^\s*$' }
        foreach ($line in $envContent) {
            $parts = $line -split '=', 2
            if ($parts.Length -eq 2) {
                $key = $parts[0].Trim()
                $value = $parts[1].Trim()
                [Environment]::SetEnvironmentVariable($key, $value, "Process")
            }
        }
    }
}

# Get services to start
function Get-Services {
    $services = ""

    # Check if services are specified in .env
    if (-not [string]::IsNullOrEmpty($env:ENABLE_SERVICES)) {
        $services = $env:ENABLE_SERVICES
        Print-Info "Using services from .env: $services" >&2
    } else {
        # Interactive mode
        Print-Header "Service Selection"
        Write-Output "Which services do you want to start?"
        Write-Output ""
        Write-Output "1) MongoDB + Backend (default)"
        Write-Output "2) MySQL + Backend"
        Write-Output "3) MongoDB only"
        Write-Output "4) MySQL only"
        Write-Output "5) Backend only (requires external database)"
        Write-Output "6) Custom selection"
        Write-Output ""
        $choice = Read-Host "Enter your choice [1-6] (default: 1)"
        if ([string]::IsNullOrEmpty($choice)) {
            $choice = "1"
        }

        switch ($choice) {
            "1" {
                $services = "mongodb,backend"
            }
            "2" {
                $services = "mysql,backend"
            }
            "3" {
                $services = "mongodb"
            }
            "4" {
                $services = "mysql"
            }
            "5" {
                $services = "backend"
                Print-Warning "Backend only mode requires an external database!" >&2
                Print-Info "Make sure DB_TYPE and corresponding DB_URI are configured in .env" >&2
            }
            "6" {
                Write-Output ""
                $services = Read-Host "Enter services (comma-separated: mongodb,mysql,backend)"
            }
            default {
                Print-Error "Invalid choice"
                exit 1
            }
        }
    }

    return $services
}

# Convert services to docker-compose profiles
function Build-Profiles {
    param([string]$services)
    $profiles = ""

    $serviceArray = $services -split ',' | ForEach-Object { $_.Trim() }
    $profiles = $serviceArray -join ','

    return $profiles
}

# Start services
function Start-Services {
    param([string]$profiles)

    Print-Header "Starting CashLenX Services"
    Print-Info "Profiles: $profiles"

    # Check if using docker-compose or docker compose
    $composeCmd = if (Get-Command docker-compose -ErrorAction SilentlyContinue) {
        "docker-compose"
    } else {
        "docker compose"
    }

    # Build profile flags
    $profileFlags = ""
    $profileArray = $profiles -split ',' | ForEach-Object { $_.Trim() }
    foreach ($profile in $profileArray) {
        $profileFlags += " --profile $profile"
    }

    # Stop existing services first
    Print-Info "Stopping existing services..."
    Invoke-Expression "$composeCmd$profileFlags down"

    # Start services with fresh build
    Print-Info "Starting services with Docker Compose..."
    Invoke-Expression "$composeCmd$profileFlags up -d --build"

    Print-Success "Services started successfully!"
}

# Show service status
function Show-Status {
    Print-Header "Service Status"

    $composeCmd = if (Get-Command docker-compose -ErrorAction SilentlyContinue) {
        "docker-compose"
    } else {
        "docker compose"
    }

    Invoke-Expression "$composeCmd ps"
}

# Show access information
function Show-Info {
    param([string]$services)
    Print-Header "Access Information"

    if ($env:ENABLE_SERVICES -like "*backend*" -or $services -like "*backend*") {
        $port = if (-not [string]::IsNullOrEmpty($env:SERVER_PORT)) { $env:SERVER_PORT } else { "8080" }
        Write-Output "Backend API: ${GREEN}http://localhost:$port${NC}"
    }

    if ($env:ENABLE_SERVICES -like "*mongodb*" -or $services -like "*mongodb*") {
        $port = if (-not [string]::IsNullOrEmpty($env:MONGO_PORT)) { $env:MONGO_PORT } else { "27017" }
        $user = if (-not [string]::IsNullOrEmpty($env:MONGO_ROOT_USERNAME)) { $env:MONGO_ROOT_USERNAME } else { "cashlenx" }
        $pass = if (-not [string]::IsNullOrEmpty($env:MONGO_ROOT_PASSWORD)) { $env:MONGO_ROOT_PASSWORD } else { "cashlenx123" }
        Write-Output "MongoDB: ${GREEN}mongodb://$user:$pass@localhost:$port${NC}"
    }

    if ($env:ENABLE_SERVICES -like "*mysql*" -or $services -like "*mysql*") {
        $port = if (-not [string]::IsNullOrEmpty($env:MYSQL_PORT)) { $env:MYSQL_PORT } else { "3306" }
        $user = if (-not [string]::IsNullOrEmpty($env:MYSQL_USER)) { $env:MYSQL_USER } else { "cashlenx" }
        $pass = if (-not [string]::IsNullOrEmpty($env:MYSQL_PASSWORD)) { $env:MYSQL_PASSWORD } else { "cashlenx123" }
        Write-Output "MySQL: ${GREEN}mysql://$user:$pass@localhost:$port${NC}"
    }

    Write-Output ""
    Print-Info "View logs: docker-compose logs -f"
    Print-Info "Stop services: docker-compose down"
    Print-Info "Stop and remove volumes: docker-compose down -v"
}

# Main function
function Main {
    Print-Header "CashLenX Docker Bootstrap"

    # Check prerequisites
    Check-Docker
    Check-EnvFile

    # Load environment variables
    Load-Env

    # Get services to start
    $services = Get-Services
    $profiles = Build-Profiles -services $services

    # Start services
    Start-Services -profiles $profiles

    # Show status
    Show-Status

    # Show access information
    Show-Info -services $services

    Print-Header "Done!"
    Print-Success "CashLenX is ready to use!"
}

# Run main function
Main