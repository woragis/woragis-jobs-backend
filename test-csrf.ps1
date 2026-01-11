# CSRF Token Test Script for Windows PowerShell
# This script tests the CSRF token functionality

Write-Host "======================================" -ForegroundColor Cyan
Write-Host "CSRF Token Functionality Test" -ForegroundColor Cyan
Write-Host "======================================" -ForegroundColor Cyan
Write-Host ""

# Configuration
$baseUrl = if ($env:BASE_URL) { $env:BASE_URL } else { "http://localhost:8080" }
$apiUrl = "$baseUrl/api/v1"

# Test 1: Get CSRF Token
Write-Host "Test 1: Getting CSRF Token" -ForegroundColor Yellow
Write-Host "----------------------------" -ForegroundColor Yellow

try {
    $response = Invoke-WebRequest -Uri "$apiUrl/csrf-token" -Method Get -SessionVariable session
    $csrfToken = $response.Headers["X-CSRF-Token"]
    
    Write-Host "Response Status:" $response.StatusCode -ForegroundColor White
    
    if ($csrfToken) {
        $tokenPreview = $csrfToken.Substring(0, [Math]::Min(20, $csrfToken.Length))
        Write-Host "✓ CSRF token received in header: $tokenPreview..." -ForegroundColor Green
    } else {
        Write-Host "✗ No CSRF token in header" -ForegroundColor Red
    }
    
    # Check cookie
    $csrfCookie = $session.Cookies.GetCookies($apiUrl) | Where-Object { $_.Name -eq "csrf_token" }
    if ($csrfCookie) {
        $cookiePreview = $csrfCookie.Value.Substring(0, [Math]::Min(20, $csrfCookie.Value.Length))
        Write-Host "✓ CSRF token set in cookie: $cookiePreview..." -ForegroundColor Green
    } else {
        Write-Host "✗ No CSRF token in cookie" -ForegroundColor Red
    }
    
    Write-Host ""
    
    # Test 2: Test with missing token (should fail)
    Write-Host "Test 2: POST without CSRF Token (should fail)" -ForegroundColor Yellow
    Write-Host "-----------------------------------------------" -ForegroundColor Yellow
    
    $body = @{
        companyName = "Test"
        jobTitle = "Test"
    } | ConvertTo-Json
    
    try {
        $response2 = Invoke-WebRequest -Uri "$apiUrl/job-applications" `
            -Method Post `
            -ContentType "application/json" `
            -Body $body `
            -Headers @{ "Authorization" = "Bearer fake-token" }
        
        Write-Host "✗ Request should have failed but got:" $response2.StatusCode -ForegroundColor Red
    }
    catch {
        $statusCode = $_.Exception.Response.StatusCode.value__
        if ($statusCode -eq 403) {
            Write-Host "✓ Request correctly rejected (403)" -ForegroundColor Green
        } else {
            Write-Host "⚠ Unexpected status code: $statusCode" -ForegroundColor Yellow
        }
    }
    
    Write-Host ""
    
    # Test 3: Test with token in header
    Write-Host "Test 3: POST with CSRF Token in Header" -ForegroundColor Yellow
    Write-Host "----------------------------------------" -ForegroundColor Yellow
    
    if ($csrfToken) {
        Write-Host "Using CSRF token: $tokenPreview..." -ForegroundColor White
        
        $body3 = @{
            companyName = "Test Corp"
            jobTitle = "Software Engineer"
            location = "Remote"
            jobUrl = "https://example.com/job"
            website = "linkedin"
        } | ConvertTo-Json
        
        try {
            $response3 = Invoke-WebRequest -Uri "$apiUrl/job-applications" `
                -Method Post `
                -ContentType "application/json" `
                -Body $body3 `
                -Headers @{ 
                    "Authorization" = "Bearer YOUR_ACTUAL_JWT_TOKEN"
                    "X-CSRF-Token" = $csrfToken
                } `
                -WebSession $session
            
            Write-Host "✓ Request successful ($($response3.StatusCode))" -ForegroundColor Green
            Write-Host ($response3.Content | ConvertFrom-Json | ConvertTo-Json -Depth 10) -ForegroundColor White
        }
        catch {
            $statusCode = $_.Exception.Response.StatusCode.value__
            if ($statusCode -eq 401) {
                Write-Host "⚠ Authentication required (401) - This is expected if you don't have a valid JWT" -ForegroundColor Yellow
                Write-Host "To test fully, replace YOUR_ACTUAL_JWT_TOKEN with a real token" -ForegroundColor Yellow
            }
            elseif ($statusCode -eq 403) {
                Write-Host "✗ CSRF validation failed (403)" -ForegroundColor Red
                Write-Host $_.Exception.Message -ForegroundColor Red
            }
            else {
                Write-Host "⚠ Unexpected status code: $statusCode" -ForegroundColor Yellow
                Write-Host $_.Exception.Message -ForegroundColor Yellow
            }
        }
    } else {
        Write-Host "✗ No CSRF token available for testing" -ForegroundColor Red
    }
    
    Write-Host ""
    
    # Test 4: Test with token in cookie only
    Write-Host "Test 4: POST with CSRF Token in Cookie" -ForegroundColor Yellow
    Write-Host "----------------------------------------" -ForegroundColor Yellow
    
    if ($csrfToken) {
        try {
            $response4 = Invoke-WebRequest -Uri "$apiUrl/job-applications" `
                -Method Post `
                -ContentType "application/json" `
                -Body $body3 `
                -Headers @{ 
                    "Authorization" = "Bearer YOUR_ACTUAL_JWT_TOKEN"
                } `
                -WebSession $session
            
            Write-Host "✓ Request successful ($($response4.StatusCode))" -ForegroundColor Green
        }
        catch {
            $statusCode = $_.Exception.Response.StatusCode.value__
            if ($statusCode -eq 401) {
                Write-Host "⚠ Authentication required (401) - This is expected if you don't have a valid JWT" -ForegroundColor Yellow
            }
            elseif ($statusCode -eq 403) {
                Write-Host "✗ CSRF validation failed (403)" -ForegroundColor Red
            }
            else {
                Write-Host "⚠ Unexpected status code: $statusCode" -ForegroundColor Yellow
            }
        }
    }
    
}
catch {
    Write-Host "✗ Error during test:" -ForegroundColor Red
    Write-Host $_.Exception.Message -ForegroundColor Red
}

Write-Host ""
Write-Host "======================================" -ForegroundColor Cyan
Write-Host "Test Summary" -ForegroundColor Cyan
Write-Host "======================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "If you see CSRF validation failures (403), make sure:" -ForegroundColor White
Write-Host "1. The backend service is running" -ForegroundColor White
Write-Host "2. Redis is running and accessible" -ForegroundColor White
Write-Host "3. CORS is properly configured" -ForegroundColor White
Write-Host "4. You're using a valid JWT token for authenticated endpoints" -ForegroundColor White
Write-Host ""
Write-Host "For full testing with authentication, obtain a JWT token and" -ForegroundColor White
Write-Host "replace YOUR_ACTUAL_JWT_TOKEN in the script." -ForegroundColor White
