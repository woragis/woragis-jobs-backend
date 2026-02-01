#!/bin/bash

# Resume Download & Retrieval Test Suite
# Tests resume download, retrieval, and generation workflow

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration
API_BASE_URL="${API_BASE_URL:-http://localhost:3000}"
AUTH_SERVICE_URL="${AUTH_SERVICE_URL:-http://localhost:3001}"
TIMEOUT=10
OUTPUT_DIR="./test-downloads"

# Test user credentials (configure for your environment)
TEST_USER_EMAIL="${TEST_USER_EMAIL:-test@example.com}"
TEST_USER_PASSWORD="${TEST_USER_PASSWORD:-testpassword123}"

# Create output directory for downloaded files
mkdir -p "$OUTPUT_DIR"

function print_header() {
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}"
}

function print_subheader() {
    echo -e "${CYAN}--- $1 ---${NC}"
}

function print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

function print_info() {
    echo -e "${CYAN}ℹ $1${NC}"
}

function print_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

function print_error() {
    echo -e "${RED}✗ $1${NC}"
}

# Check if curl is available
function check_curl() {
    if ! command -v curl &> /dev/null; then
        print_error "curl not found. Please install curl."
        exit 1
    fi
    print_success "curl found"
}

# Check if services are running
function check_services() {
    print_header "Checking Services"
    
    print_info "Checking Jobs Service at $API_BASE_URL..."
    if curl -s -m $TIMEOUT "$API_BASE_URL/healthz" > /dev/null 2>&1; then
        print_success "Jobs Service is running"
    else
        print_error "Jobs Service is not responding at $API_BASE_URL"
        print_info "Make sure the jobs backend is running with: docker-compose up -d"
        exit 1
    fi
}

# Get CSRF token
function get_csrf_token() {
    print_subheader "Getting CSRF Token"
    
    # First, get CSRF token from the API (GET request doesn't need CSRF to get CSRF token)
    local response=$(curl -s -X GET "$API_BASE_URL/api/v1/csrf-token" \
        -c "$OUTPUT_DIR/cookies.txt" \
        -D "$OUTPUT_DIR/csrf-headers.txt" \
        -m $TIMEOUT)
    
    # Extract CSRF token from X-CSRF-Token header
    if [ -f "$OUTPUT_DIR/csrf-headers.txt" ]; then
        CSRF_TOKEN=$(grep -i "X-CSRF-Token:" "$OUTPUT_DIR/csrf-headers.txt" | awk '{print $2}' | tr -d '\r\n')
        if [ -n "$CSRF_TOKEN" ]; then
            print_success "CSRF token obtained: ${CSRF_TOKEN:0:20}..."
            return 0
        fi
    fi
    
    print_warning "Could not obtain CSRF token"
    return 1
}

# Get auth token
function get_auth_token() {
    print_subheader "Getting Auth Token"
    
    # First get CSRF token (required for login)
    get_csrf_token
    
    # Try to get token from auth service using the CSRF token
    local login_payload="{\"email\": \"$TEST_USER_EMAIL\", \"password\": \"$TEST_USER_PASSWORD\"}"
    
    local response=$(curl -s -X POST "$AUTH_SERVICE_URL/api/v1/auth/login" \
        -H "Content-Type: application/json" \
        -H "X-CSRF-Token: $CSRF_TOKEN" \
        -b "$OUTPUT_DIR/cookies.txt" \
        -d "$login_payload" \
        -m $TIMEOUT)
    
    # Try different JSON extraction methods for the token
    if echo "$response" | grep -q "token"; then
        AUTH_TOKEN=$(echo "$response" | grep -o '"token":"[^"]*' | head -1 | cut -d'"' -f4)
        if [ -z "$AUTH_TOKEN" ]; then
            AUTH_TOKEN=$(echo "$response" | grep -o '"access_token":"[^"]*' | head -1 | cut -d'"' -f4)
        fi
        if [ -n "$AUTH_TOKEN" ]; then
            print_success "Auth token obtained: ${AUTH_TOKEN:0:20}..."
            return 0
        fi
    fi
    
    print_warning "Could not obtain auth token from auth service"
    print_info "Response: $response"
    print_info "Using mock token. Set TEST_USER_EMAIL and TEST_USER_PASSWORD for real auth."
    AUTH_TOKEN="mock-jwt-token-for-testing"
    return 1
}

# Test: List all resumes
function test_list_resumes() {
    print_subheader "Test 1: List All Resumes"
    
    local response=$(curl -s -X GET "$API_BASE_URL/api/v1/resumes" \
        -H "Authorization: Bearer $AUTH_TOKEN" \
        -H "Content-Type: application/json" \
        -H "X-CSRF-Token: $CSRF_TOKEN" \
        -b "$OUTPUT_DIR/cookies.txt" \
        -m $TIMEOUT \
        -w "\n%{http_code}")
    
    local http_code=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | sed '$d')
    
    if [[ $http_code == 200 ]]; then
        print_success "Listed resumes (HTTP $http_code)"
        echo "$body" | jq . 2>/dev/null || echo "$body"
        
        # Extract first resume ID for later tests
        FIRST_RESUME_ID=$(echo "$body" | jq -r '.[0].id // empty' 2>/dev/null)
        if [ -n "$FIRST_RESUME_ID" ]; then
            print_success "Found resume ID: $FIRST_RESUME_ID"
        else
            print_warning "No resumes found in list"
        fi
    else
        print_error "Failed to list resumes (HTTP $http_code)"
        echo "$body"
    fi
    
    echo ""
}

# Test: Get single resume by ID
function test_get_resume() {
    print_subheader "Test 2: Get Single Resume by ID"
    
    if [ -z "$FIRST_RESUME_ID" ]; then
        print_warning "Skipping: No resume ID available"
        return
    fi
    
    local response=$(curl -s -X GET "$API_BASE_URL/api/v1/resumes/$FIRST_RESUME_ID" \
        -H "Authorization: Bearer $AUTH_TOKEN" \
        -H "Content-Type: application/json" \
        -H "X-CSRF-Token: $CSRF_TOKEN" \
        -b "$OUTPUT_DIR/cookies.txt" \
        -m $TIMEOUT \
        -w "\n%{http_code}")
    
    local http_code=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | sed '$d')
    
    if [[ $http_code == 200 ]]; then
        print_success "Retrieved resume metadata (HTTP $http_code)"
        echo "$body" | jq . 2>/dev/null || echo "$body"
    else
        print_error "Failed to get resume (HTTP $http_code)"
        echo "$body"
    fi
    
    echo ""
}

# Test: Download resume by ID
function test_download_resume_by_id() {
    print_subheader "Test 3: Download Resume by ID"
    
    if [ -z "$FIRST_RESUME_ID" ]; then
        print_warning "Skipping: No resume ID available"
        return
    fi
    
    local output_file="$OUTPUT_DIR/resume-by-id-$FIRST_RESUME_ID.pdf"
    
    local response=$(curl -s -X GET "$API_BASE_URL/api/v1/resumes/$FIRST_RESUME_ID/download" \
        -H "Authorization: Bearer $AUTH_TOKEN" \
        -H "X-CSRF-Token: $CSRF_TOKEN" \
        -b "$OUTPUT_DIR/cookies.txt" \
        -m $TIMEOUT \
        -w "\n%{http_code}" \
        -D "$OUTPUT_DIR/headers-by-id.txt" \
        -o "$output_file")
    
    local http_code=$(tail -n1 <<< "$response")
    
    if [[ $http_code == 200 ]]; then
        local file_size=$(stat -f%z "$output_file" 2>/dev/null || stat -c%s "$output_file" 2>/dev/null || echo "unknown")
        print_success "Downloaded resume by ID (HTTP $http_code, Size: $file_size bytes)"
        print_info "Saved to: $output_file"
        
        # Check response headers
        if grep -q "Content-Disposition" "$OUTPUT_DIR/headers-by-id.txt"; then
            print_success "Content-Disposition header present"
            grep "Content-Disposition" "$OUTPUT_DIR/headers-by-id.txt"
        fi
    else
        print_error "Failed to download resume by ID (HTTP $http_code)"
        [ -f "$output_file" ] && cat "$output_file"
    fi
    
    echo ""
}

# Test: Download resume by user ID (public endpoint)
function test_download_resume_by_user() {
    print_subheader "Test 4: Download Resume by User ID (Public Endpoint)"
    
    if [ -z "$FIRST_RESUME_ID" ]; then
        print_warning "Skipping: No resume ID available"
        return
    fi
    
    # First get a resume to find the user ID
    local resume_info=$(curl -s -X GET "$API_BASE_URL/api/v1/resumes/$FIRST_RESUME_ID" \
        -H "Authorization: Bearer $AUTH_TOKEN" \
        -H "X-CSRF-Token: $CSRF_TOKEN" \
        -b "$OUTPUT_DIR/cookies.txt" \
        -m $TIMEOUT)
    
    local user_id=$(echo "$resume_info" | jq -r '.user_id // empty' 2>/dev/null)
    
    if [ -z "$user_id" ]; then
        print_warning "Could not extract user_id from resume"
        return
    fi
    
    local output_file="$OUTPUT_DIR/resume-by-user-$user_id.pdf"
    
    local response=$(curl -s -X GET "$API_BASE_URL/resume/download?userId=$user_id" \
        -m $TIMEOUT \
        -w "\n%{http_code}" \
        -D "$OUTPUT_DIR/headers-by-user.txt" \
        -o "$output_file")
    
    local http_code=$(tail -n1 <<< "$response")
    
    if [[ $http_code == 200 ]]; then
        local file_size=$(stat -f%z "$output_file" 2>/dev/null || stat -c%s "$output_file" 2>/dev/null || echo "unknown")
        print_success "Downloaded resume by user ID (HTTP $http_code, Size: $file_size bytes)"
        print_info "Saved to: $output_file"
    else
        print_warning "Failed to download resume by user ID (HTTP $http_code)"
        [ -f "$output_file" ] && head -c 200 "$output_file"
    fi
    
    echo ""
}

# Test: Create a resume
function test_create_resume() {
    print_subheader "Test 5: Create Resume"
    
    local response=$(curl -s -X POST "$API_BASE_URL/api/v1/resumes" \
        -H "Authorization: Bearer $AUTH_TOKEN" \
        -H "Content-Type: application/json" \
        -H "X-CSRF-Token: $CSRF_TOKEN" \
        -b "$OUTPUT_DIR/cookies.txt" \
        -d '{
            "title": "Test Resume",
            "tags": ["test", "curl"],
            "is_main": false
        }' \
        -m $TIMEOUT \
        -w "\n%{http_code}")
    
    local http_code=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | sed '$d')
    
    if [[ $http_code == 201 ]]; then
        print_success "Created resume (HTTP $http_code)"
        echo "$body" | jq . 2>/dev/null || echo "$body"
        
        # Extract created resume ID
        CREATED_RESUME_ID=$(echo "$body" | jq -r '.id // empty' 2>/dev/null)
        if [ -n "$CREATED_RESUME_ID" ]; then
            print_success "Created resume ID: $CREATED_RESUME_ID"
        fi
    else
        print_error "Failed to create resume (HTTP $http_code)"
        echo "$body"
    fi
    
    echo ""
}

# Test: List resumes with tag filter
function test_list_resumes_with_filter() {
    print_subheader "Test 6: List Resumes with Tag Filter"
    
    local response=$(curl -s -X GET "$API_BASE_URL/api/v1/resumes?tags=test,curl" \
        -H "Authorization: Bearer $AUTH_TOKEN" \
        -H "Content-Type: application/json" \
        -H "X-CSRF-Token: $CSRF_TOKEN" \
        -b "$OUTPUT_DIR/cookies.txt" \
        -m $TIMEOUT \
        -w "\n%{http_code}")
    
    local http_code=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | sed '$d')
    
    if [[ $http_code == 200 ]]; then
        print_success "Listed filtered resumes (HTTP $http_code)"
        echo "$body" | jq . 2>/dev/null || echo "$body"
    else
        print_warning "Failed to list filtered resumes (HTTP $http_code)"
    fi
    
    echo ""
}

# Test: Generate resume
function test_generate_resume() {
    print_subheader "Test 7: Generate Resume"
    
    if [ -z "$CREATED_RESUME_ID" ]; then
        print_warning "Skipping: No created resume ID available"
        return
    fi
    
    local response=$(curl -s -X POST "$API_BASE_URL/api/v1/resumes/generate" \
        -H "Authorization: Bearer $AUTH_TOKEN" \
        -H "Content-Type: application/json" \
        -H "X-CSRF-Token: $CSRF_TOKEN" \
        -b "$OUTPUT_DIR/cookies.txt" \
        -d "{
            \"resume_id\": \"$CREATED_RESUME_ID\",
            \"template\": \"modern\",
            \"format\": \"pdf\"
        }" \
        -m $TIMEOUT \
        -w "\n%{http_code}")
    
    local http_code=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | sed '$d')
    
    if [[ $http_code == 202 ]] || [[ $http_code == 200 ]]; then
        print_success "Resume generation triggered (HTTP $http_code)"
        echo "$body" | jq . 2>/dev/null || echo "$body"
        
        # Extract job ID for status polling
        JOB_ID=$(echo "$body" | jq -r '.job_id // .id // empty' 2>/dev/null)
        if [ -n "$JOB_ID" ]; then
            print_success "Job ID: $JOB_ID"
        fi
    else
        print_error "Failed to generate resume (HTTP $http_code)"
        echo "$body"
    fi
    
    echo ""
}

# Test: Get job status
function test_get_job_status() {
    print_subheader "Test 8: Get Resume Generation Job Status"
    
    if [ -z "$JOB_ID" ]; then
        print_warning "Skipping: No job ID available"
        return
    fi
    
    # Poll status a few times
    for i in {1..3}; do
        print_info "Polling status (attempt $i/3)..."
        
        local response=$(curl -s -X GET "$API_BASE_URL/api/v1/resumes/jobs/$JOB_ID" \
            -H "Authorization: Bearer $AUTH_TOKEN" \
            -H "X-CSRF-Token: $CSRF_TOKEN" \
            -b "$OUTPUT_DIR/cookies.txt" \
            -m $TIMEOUT \
            -w "\n%{http_code}")
        
        local http_code=$(echo "$response" | tail -n1)
        local body=$(echo "$response" | sed '$d')
        
        if [[ $http_code == 200 ]]; then
            echo "$body" | jq . 2>/dev/null || echo "$body"
            
            local status=$(echo "$body" | jq -r '.status // empty' 2>/dev/null)
            if [[ "$status" == "completed" ]]; then
                print_success "Resume generation completed!"
                break
            elif [[ "$status" == "failed" ]]; then
                print_error "Resume generation failed"
                break
            fi
        else
            print_warning "Failed to get job status (HTTP $http_code)"
        fi
        
        if [ $i -lt 3 ]; then
            sleep 2
        fi
    done
    
    echo ""
}

# Test: Authentication error (no token)
function test_auth_error() {
    print_subheader "Test 9: Authentication Error (No Token)"
    
    local response=$(curl -s -X GET "$API_BASE_URL/api/v1/resumes" \
        -H "Content-Type: application/json" \
        -H "X-CSRF-Token: $CSRF_TOKEN" \
        -b "$OUTPUT_DIR/cookies.txt" \
        -m $TIMEOUT \
        -w "\n%{http_code}")
    
    local http_code=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | sed '$d')
    
    if [[ $http_code == 401 ]] || [[ $http_code == 403 ]]; then
        print_success "Authentication properly enforced (HTTP $http_code)"
    else
        print_warning "Expected 401/403, got HTTP $http_code"
    fi
    
    echo ""
}

# Test: 404 error (non-existent resume)
function test_not_found_error() {
    print_subheader "Test 10: Not Found Error (Non-existent Resume)"
    
    local response=$(curl -s -X GET "$API_BASE_URL/api/v1/resumes/nonexistent-id" \
        -H "Authorization: Bearer $AUTH_TOKEN" \
        -H "X-CSRF-Token: $CSRF_TOKEN" \
        -b "$OUTPUT_DIR/cookies.txt" \
        -m $TIMEOUT \
        -w "\n%{http_code}")
    
    local http_code=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | sed '$d')
    
    if [[ $http_code == 404 ]]; then
        print_success "404 error properly returned (HTTP $http_code)"
    else
        print_warning "Expected 404, got HTTP $http_code"
    fi
    
    echo ""
}

# Test: Resume preview (public endpoint)
function test_preview_resume() {
    print_subheader "Test 11: Resume Preview (Public Endpoint)"
    
    if [ -z "$FIRST_RESUME_ID" ]; then
        print_warning "Skipping: No resume ID available"
        return
    fi
    
    local response=$(curl -s -X GET "$API_BASE_URL/resume/preview?resumeId=$FIRST_RESUME_ID" \
        -H "X-CSRF-Token: $CSRF_TOKEN" \
        -b "$OUTPUT_DIR/cookies.txt" \
        -m $TIMEOUT \
        -w "\n%{http_code}")
    
    local http_code=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | sed '$d')
    
    if [[ $http_code == 200 ]]; then
        print_success "Retrieved resume preview (HTTP $http_code)"
    else
        print_warning "Failed to get preview (HTTP $http_code)"
    fi
    
    echo ""
}

# Main execution
function main() {
    print_header "Resume Download & Retrieval Test Suite"
    print_info "API Base URL: $API_BASE_URL"
    print_info "Auth Service URL: $AUTH_SERVICE_URL"
    echo ""
    
    check_curl
    check_services
    
    if ! get_auth_token; then
        print_warning "Using mock auth token - some tests may fail"
    fi
    
    echo ""
    test_list_resumes
    test_get_resume
    test_download_resume_by_id
    test_download_resume_by_user
    test_create_resume
    test_list_resumes_with_filter
    test_generate_resume
    test_get_job_status
    test_auth_error
    test_not_found_error
    test_preview_resume
    
    print_header "Test Summary"
    print_info "Downloaded files saved to: $OUTPUT_DIR"
    print_success "Test suite completed"
}

# Run main
main
