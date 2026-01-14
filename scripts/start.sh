#!/bin/bash

# CashLenX Docker Bootstrap Script
# This script helps you easily start the CashLenX application with Docker

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Print colored output
print_info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

print_header() {
    echo -e "\n${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}  $1${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}\n"
}

# Check if Docker is installed
check_docker() {
    if ! command -v docker &> /dev/null; then
        print_error "Docker is not installed. Please install Docker first."
        exit 1
    fi

    if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
        print_error "Docker Compose is not installed. Please install Docker Compose first."
        exit 1
    fi

    print_success "Docker and Docker Compose are installed"
}

# Check if .env file exists
check_env_file() {
    if [ ! -f .env ]; then
        print_warning ".env file not found"
        print_info "Creating .env from .env.sample..."

        if [ -f .env.sample ]; then
            cp .env.sample .env
            print_success ".env file created from .env.sample"
            print_warning "Please review and update .env with your configuration"
        else
            print_error ".env.sample file not found"
            exit 1
        fi
    else
        print_success ".env file exists"
    fi
}

# Load environment variables from .env
load_env() {
    if [ -f .env ]; then
        export $(cat .env | grep -v '^#' | grep -v '^[[:space:]]*$' | xargs)
    fi
}

# Get services to start
get_services() {
    local services=""

    # Check if services are specified in .env
    if [ -n "$ENABLE_SERVICES" ]; then
        services=$ENABLE_SERVICES
        print_info "Using services from .env: $services" >&2
    else
        # Interactive mode
        print_header "Service Selection"
        echo "Which services do you want to start?"
        echo ""
        echo "1) MongoDB + Backend (default)"
        echo "2) MySQL + Backend"
        echo "3) MongoDB only"
        echo "4) MySQL only"
        echo "5) Backend only (requires external database)"
        echo "6) Custom selection"
        echo ""
        read -p "Enter your choice [1-6] (default: 1): " choice
        choice=${choice:-1}

        case $choice in
            1)
                services="mongodb,backend"
                ;;
            2)
                services="mysql,backend"
                ;;
            3)
                services="mongodb"
                ;;
            4)
                services="mysql"
                ;;
            5)
                services="backend"
                print_warning "Backend only mode requires an external database!" >&2
                print_info "Make sure DB_TYPE and corresponding DB_URI are configured in .env" >&2
                ;;
            6)
                echo ""
                read -p "Enter services (comma-separated: mongodb,mysql,backend): " services
                ;;
            *)
                print_error "Invalid choice"
                exit 1
                ;;
        esac
    fi

    echo "$services"
}

# Convert services to docker-compose profiles
build_profiles() {
    local services=$1
    local profiles=""

    IFS=',' read -ra SERVICE_ARRAY <<< "$services"
    for service in "${SERVICE_ARRAY[@]}"; do
        service=$(echo "$service" | xargs) # trim whitespace
        if [ -n "$profiles" ]; then
            profiles="$profiles,$service"
        else
            profiles="$service"
        fi
    done

    echo "$profiles"
}

# Start services
start_services() {
    local profiles=$1

    print_header "Starting CashLenX Services"
    print_info "Profiles: $profiles"

    # Check if using docker-compose or docker compose
    if command -v docker-compose &> /dev/null; then
        COMPOSE_CMD="docker-compose"
    else
        COMPOSE_CMD="docker compose"
    fi

    # Build profile flags
    local profile_flags=""
    IFS=',' read -ra PROFILE_ARRAY <<< "$profiles"
    for profile in "${PROFILE_ARRAY[@]}"; do
        profile=$(echo "$profile" | xargs) # trim whitespace
        profile_flags="$profile_flags --profile $profile"
    done

    # Stop existing services first
    print_info "Stopping existing services..."
    $COMPOSE_CMD $profile_flags down

    # Start services with fresh build
    print_info "Starting services with Docker Compose..."
    $COMPOSE_CMD $profile_flags up -d --build

    print_success "Services started successfully!"
}

# Show service status
show_status() {
    print_header "Service Status"

    if command -v docker-compose &> /dev/null; then
        COMPOSE_CMD="docker-compose"
    else
        COMPOSE_CMD="docker compose"
    fi

    $COMPOSE_CMD ps
}

# Show access information
show_info() {
    print_header "Access Information"

    if [[ "$ENABLE_SERVICES" == *"backend"* ]] || [[ "$1" == *"backend"* ]]; then
        local port=${SERVER_PORT:-8080}
        echo -e "Backend API: ${GREEN}http://localhost:$port${NC}"
    fi

    if [[ "$ENABLE_SERVICES" == *"mongodb"* ]] || [[ "$1" == *"mongodb"* ]]; then
        local port=${MONGO_PORT:-27017}
        local user=${MONGO_ROOT_USERNAME:-cashlenx}
        local pass=${MONGO_ROOT_PASSWORD:-cashlenx123}
        echo -e "MongoDB: ${GREEN}mongodb://$user:$pass@localhost:$port${NC}"
    fi

    if [[ "$ENABLE_SERVICES" == *"mysql"* ]] || [[ "$1" == *"mysql"* ]]; then
        local port=${MYSQL_PORT:-3306}
        local user=${MYSQL_USER:-cashlenx}
        local pass=${MYSQL_PASSWORD:-cashlenx123}
        echo -e "MySQL: ${GREEN}mysql://$user:$pass@localhost:$port${NC}"
    fi

    echo ""
    print_info "View logs: docker-compose logs -f"
    print_info "Stop services: docker-compose down"
    print_info "Stop and remove volumes: docker-compose down -v"
}

# Main function
main() {
    print_header "CashLenX Docker Bootstrap"

    # Check prerequisites
    check_docker
    check_env_file

    # Load environment variables
    load_env

    # Get services to start
    services=$(get_services)
    profiles=$(build_profiles "$services")

    # Start services
    start_services "$profiles"

    # Show status
    show_status

    # Show access information
    show_info "$services"

    print_header "Done!"
    print_success "CashLenX is ready to use!"
}

# Run main function
main
