#!/bin/bash

# CSRF Token Test Script
# This script tests the CSRF token functionality

echo "======================================"
echo "CSRF Token Functionality Test"
echo "======================================"
echo ""

# Configuration
BASE_URL="${BASE_URL:-http://localhost:8080}"
API_URL="${BASE_URL}/api/v1"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test 1: Get CSRF Token
echo "Test 1: Getting CSRF Token"
echo "----------------------------"

RESPONSE=$(curl -s -c /tmp/csrf_cookies.txt -D /tmp/csrf_headers.txt "${API_URL}/csrf-token")
echo "Response: $RESPONSE"

# Extract token from header
CSRF_TOKEN=$(grep -i "X-CSRF-Token:" /tmp/csrf_headers.txt | awk '{print $2}' | tr -d '\r\n')

if [ -n "$CSRF_TOKEN" ]; then
    echo -e "${GREEN}✓ CSRF token received in header: ${CSRF_TOKEN:0:20}...${NC}"
else
    echo -e "${RED}✗ No CSRF token in header${NC}"
fi

# Check cookie
if grep -q "csrf_token" /tmp/csrf_cookies.txt; then
    COOKIE_TOKEN=$(grep "csrf_token" /tmp/csrf_cookies.txt | awk '{print $7}')
    echo -e "${GREEN}✓ CSRF token set in cookie: ${COOKIE_TOKEN:0:20}...${NC}"
else
    echo -e "${RED}✗ No CSRF token in cookie${NC}"
fi

echo ""

# Test 2: Verify Cookie Properties
echo "Test 2: Cookie Properties"
echo "----------------------------"
if grep -q "csrf_token" /tmp/csrf_cookies.txt; then
    cat /tmp/csrf_cookies.txt | grep "csrf_token"
    
    # Check for SameSite attribute in headers
    COOKIE_HEADER=$(grep -i "Set-Cookie: csrf_token" /tmp/csrf_headers.txt)
    echo "Cookie header: $COOKIE_HEADER"
    
    if echo "$COOKIE_HEADER" | grep -qi "SameSite=Lax"; then
        echo -e "${GREEN}✓ SameSite=Lax is set${NC}"
    else
        echo -e "${YELLOW}⚠ SameSite attribute not Lax${NC}"
    fi
fi

echo ""

# Test 3: Test with missing token (should fail)
echo "Test 3: POST without CSRF Token (should fail)"
echo "-----------------------------------------------"

RESPONSE=$(curl -s -w "\nHTTP_CODE:%{http_code}" -X POST "${API_URL}/job-applications" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer fake-token" \
    -d '{"companyName":"Test","jobTitle":"Test"}')

HTTP_CODE=$(echo "$RESPONSE" | grep "HTTP_CODE:" | cut -d: -f2)

if [ "$HTTP_CODE" = "403" ]; then
    echo -e "${GREEN}✓ Request correctly rejected (403)${NC}"
    echo "$RESPONSE" | grep -v "HTTP_CODE:"
else
    echo -e "${RED}✗ Unexpected response code: $HTTP_CODE${NC}"
fi

echo ""

# Test 4: Test with token in header (should work if authenticated)
echo "Test 4: POST with CSRF Token in Header"
echo "----------------------------------------"

if [ -n "$CSRF_TOKEN" ]; then
    echo "Using CSRF token: ${CSRF_TOKEN:0:20}..."
    
    RESPONSE=$(curl -s -w "\nHTTP_CODE:%{http_code}" -b /tmp/csrf_cookies.txt \
        -X POST "${API_URL}/job-applications" \
        -H "Content-Type: application/json" \
        -H "X-CSRF-Token: $CSRF_TOKEN" \
        -H "Authorization: Bearer YOUR_ACTUAL_JWT_TOKEN" \
        -d '{"companyName":"Test Corp","jobTitle":"Software Engineer","location":"Remote","jobUrl":"https://example.com/job","website":"linkedin"}')
    
    HTTP_CODE=$(echo "$RESPONSE" | grep "HTTP_CODE:" | cut -d: -f2)
    
    if [ "$HTTP_CODE" = "401" ]; then
        echo -e "${YELLOW}⚠ Authentication required (401) - This is expected if you don't have a valid JWT${NC}"
        echo "To test fully, replace YOUR_ACTUAL_JWT_TOKEN with a real token"
    elif [ "$HTTP_CODE" = "403" ]; then
        echo -e "${RED}✗ CSRF validation failed (403)${NC}"
        echo "$RESPONSE" | grep -v "HTTP_CODE:"
    elif [ "$HTTP_CODE" = "201" ] || [ "$HTTP_CODE" = "200" ]; then
        echo -e "${GREEN}✓ Request successful ($HTTP_CODE)${NC}"
        echo "$RESPONSE" | grep -v "HTTP_CODE:" | jq '.' 2>/dev/null || echo "$RESPONSE" | grep -v "HTTP_CODE:"
    else
        echo -e "${YELLOW}⚠ Unexpected response code: $HTTP_CODE${NC}"
        echo "$RESPONSE" | grep -v "HTTP_CODE:"
    fi
else
    echo -e "${RED}✗ No CSRF token available for testing${NC}"
fi

echo ""

# Test 5: Test with token in cookie only
echo "Test 5: POST with CSRF Token in Cookie"
echo "----------------------------------------"

if [ -n "$CSRF_TOKEN" ]; then
    RESPONSE=$(curl -s -w "\nHTTP_CODE:%{http_code}" -b /tmp/csrf_cookies.txt \
        -X POST "${API_URL}/job-applications" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer YOUR_ACTUAL_JWT_TOKEN" \
        -d '{"companyName":"Test Corp","jobTitle":"Software Engineer","location":"Remote","jobUrl":"https://example.com/job","website":"linkedin"}')
    
    HTTP_CODE=$(echo "$RESPONSE" | grep "HTTP_CODE:" | cut -d: -f2)
    
    if [ "$HTTP_CODE" = "401" ]; then
        echo -e "${YELLOW}⚠ Authentication required (401) - This is expected if you don't have a valid JWT${NC}"
    elif [ "$HTTP_CODE" = "403" ]; then
        echo -e "${RED}✗ CSRF validation failed (403)${NC}"
        echo "$RESPONSE" | grep -v "HTTP_CODE:"
    elif [ "$HTTP_CODE" = "201" ] || [ "$HTTP_CODE" = "200" ]; then
        echo -e "${GREEN}✓ Request successful ($HTTP_CODE)${NC}"
        echo "$RESPONSE" | grep -v "HTTP_CODE:" | jq '.' 2>/dev/null || echo "$RESPONSE" | grep -v "HTTP_CODE:"
    else
        echo -e "${YELLOW}⚠ Unexpected response code: $HTTP_CODE${NC}"
        echo "$RESPONSE" | grep -v "HTTP_CODE:"
    fi
fi

echo ""
echo "======================================"
echo "Test Summary"
echo "======================================"
echo ""
echo "If you see CSRF validation failures (403), make sure:"
echo "1. The backend service is running"
echo "2. Redis is running and accessible"
echo "3. CORS is properly configured"
echo "4. You're using a valid JWT token for authenticated endpoints"
echo ""
echo "For full testing with authentication, obtain a JWT token and"
echo "replace YOUR_ACTUAL_JWT_TOKEN in the script."

# Cleanup
rm -f /tmp/csrf_cookies.txt /tmp/csrf_headers.txt
