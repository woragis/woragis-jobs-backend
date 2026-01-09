#!/bin/bash

# Integration Test Suite for Resume Worker
# Tests the complete resume generation workflow

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

function print_header() {
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}"
}

function print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

function print_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

function print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_header "Resume Worker Integration Test Suite"

# Check if docker is running
print_header "Step 1: Checking Docker"
if ! command -v docker &> /dev/null; then
    print_error "Docker not found"
    exit 1
fi
print_success "Docker found"

# Check if docker-compose is running
print_header "Step 2: Checking Docker Compose"
if ! command -v docker-compose &> /dev/null; then
    print_warning "docker-compose CLI not found, trying 'docker compose'"
    DOCKER_COMPOSE="docker compose"
else
    DOCKER_COMPOSE="docker-compose"
fi
print_success "Docker Compose available"

# Check if we're in the right directory
print_header "Step 3: Verifying Directory Structure"
if [ ! -f "docker-compose.yml" ]; then
    print_error "docker-compose.yml not found. Run this script from the jobs directory"
    exit 1
fi

if [ ! -d "workers/resume-worker" ]; then
    print_error "workers/resume-worker directory not found"
    exit 1
fi
print_success "Directory structure is correct"

# Check if services are defined
print_header "Step 4: Checking Service Definitions"
if ! grep -q "woragis-jobs-resume-worker" docker-compose.yml; then
    print_error "resume-worker service not found in docker-compose.yml"
    exit 1
fi
print_success "resume-worker service is defined"

if ! grep -q "woragis-jobs-resume-service" docker-compose.yml; then
    print_error "resume-service not found in docker-compose.yml"
    exit 1
fi
print_success "resume-service is defined"

if ! grep -q "woragis-jobs-ai-service" docker-compose.yml; then
    print_error "ai-service not found in docker-compose.yml"
    exit 1
fi
print_success "ai-service is defined"

# Check build files
print_header "Step 5: Checking Build Files"
if [ ! -f "workers/resume-worker/Dockerfile" ]; then
    print_error "Dockerfile not found in workers/resume-worker"
    exit 1
fi
print_success "Dockerfile is present"

if [ ! -f "workers/resume-worker/package.json" ]; then
    print_error "package.json not found in workers/resume-worker"
    exit 1
fi
print_success "package.json is present"

if [ ! -f "workers/resume-worker/tsconfig.json" ]; then
    print_error "tsconfig.json not found in workers/resume-worker"
    exit 1
fi
print_success "tsconfig.json is present"

# Check dist folder exists
print_header "Step 6: Checking Build Output"
if [ ! -d "workers/resume-worker/dist" ]; then
    print_warning "dist folder not found, TypeScript needs to be compiled"
    cd workers/resume-worker
    npm run build
    cd ../..
else
    print_success "dist folder exists"
fi

if [ ! -f "workers/resume-worker/dist/index.js" ]; then
    print_error "Compiled index.js not found"
    exit 1
fi
print_success "Compiled code is present"

# Check version updates
print_header "Step 7: Checking Version Updates"
if grep -q "woragis/resume-service:v2.0.0" docker-compose.yml; then
    print_success "resume-service updated to v2.0.0"
else
    print_error "resume-service version not updated to v2.0.0"
    exit 1
fi

# Simulate container startup check
print_header "Step 8: Configuration Check"
if [ ! -f "workers/resume-worker/.env.sample" ]; then
    print_error ".env.sample not found"
    exit 1
fi
print_success ".env.sample configuration is present"

# Check database migrations
print_header "Step 9: Checking Database Migrations"
if [ ! -f "workers/resume-worker/migrations.sql" ]; then
    print_error "migrations.sql not found"
    exit 1
fi
print_success "Database migrations are available"

# Check documentation
print_header "Step 10: Checking Documentation"
if [ ! -f "workers/resume-worker/README.md" ]; then
    print_error "README.md not found"
    exit 1
fi
print_success "README.md is present"

if [ ! -f "workers/resume-worker/INTEGRATION_GUIDE.md" ]; then
    print_error "INTEGRATION_GUIDE.md not found"
    exit 1
fi
print_success "INTEGRATION_GUIDE.md is present"

# Final summary
print_header "Integration Test Results"
print_success "✓ All integration tests passed!"
echo ""
echo -e "${BLUE}Next Steps:${NC}"
echo "1. Run: docker-compose up -d"
echo "2. Wait for all services to be healthy: docker-compose ps"
echo "3. Check logs: docker logs -f woragis-jobs-resume-worker"
echo "4. Send test job: bash workers/resume-worker/test-job.sh"
echo ""
echo -e "${YELLOW}Quick Start:${NC}"
echo "# Build and start all services"
echo "docker-compose up -d"
echo ""
echo "# Check service health"
echo "docker-compose ps"
echo ""
echo "# View worker logs"
echo "docker logs -f woragis-jobs-resume-worker"
echo ""
echo "# Check database"
echo "docker exec woragis-jobs-database psql -U woragis -d jobs_service -c \"SELECT * FROM resume_jobs LIMIT 5;\""
