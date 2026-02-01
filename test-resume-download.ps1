# Resume Download & Retrieval Test Suite (PowerShell)
# Tests resume download, retrieval, and generation workflow

# Configuration
$ApiBaseUrl = $env:API_BASE_URL -or "http://localhost:3000"
$AuthServiceUrl = $env:AUTH_SERVICE_URL -or "http://localhost:3001"
$Timeout = 10
$OutputDir = "./test-downloads"
$TestUserEmail = $env:TEST_USER_EMAIL -or "test@example.com"
$TestUserPassword = $env:TEST_USER_PASSWORD -or "testpassword123"

# Create output directory
if (-not (Test-Path $OutputDir)) {
    New-Item -ItemType Directory -Path $OutputDir | Out-Null
}

# Color functions
function Write-Header {
    param([string]$Message)
    Write-Host "========================================" -ForegroundColor Blue
    Write-Host $Message -ForegroundColor Blue
    Write-Host "========================================" -ForegroundColor Blue
}

function Write-SubHeader {
    param([string]$Message)
    Write-Host "--- $Message ---" -ForegroundColor Cyan
}

function Write-Success {
    param([string]$Message)
    Write-Host "✓ $Message" -ForegroundColor Green
}

function Write-Info {
    param([string]$Message)
    Write-Host "ℹ $Message" -ForegroundColor Cyan
}

function Write-Warning {
    param([string]$Message)
    Write-Host "⚠ $Message" -ForegroundColor Yellow
}

function Write-Error {
    param([string]$Message)
    Write-Host "✗ $Message" -ForegroundColor Red
}

# Check if curl is available
function Check-Curl {
    try {
        $null = curl.exe --version
        Write-Success "curl found"
    } catch {
        Write-Error "curl not found. Please install curl or use curl.exe from Git Bash."
        exit 1
    }
}

# Check if services are running
function Check-Services {
    Write-Header "Checking Services"
    
    Write-Info "Checking Jobs Service at $ApiBaseUrl..."
    try {
        $response = curl.exe -s -m $Timeout "$ApiBaseUrl/healthz" 2>$null
        Write-Success "Jobs Service is running"
    } catch {
        Write-Error "Jobs Service is not responding at $ApiBaseUrl"
        Write-Info "Make sure the jobs backend is running with: docker-compose up -d"
        exit 1
    }
}

# Get auth token
function Get-AuthToken {
    Write-SubHeader "Getting Auth Token"
    
    try {
        $authPayload = @{
            email = $TestUserEmail
            password = $TestUserPassword
        } | ConvertTo-Json
        
        $response = curl.exe -s -X POST "$AuthServiceUrl/api/v1/auth/login" `
            -H "Content-Type: application/json" `
            -d $authPayload `
            -m $Timeout 2>$null
        
        if ($response -match '"token":"[^"]*') {
            $script:AuthToken = ($response | Select-String -Pattern '"token":"([^"]*)"' | ForEach-Object { $_.Matches[0].Groups[1].Value })
            Write-Success "Auth token obtained: $($script:AuthToken.Substring(0, 20))..."
            return $true
        } else {
            Write-Warning "Could not obtain auth token from auth service"
            Write-Info "Using mock token. Set TEST_USER_EMAIL and TEST_USER_PASSWORD for real auth."
            $script:AuthToken = "mock-jwt-token-for-testing"
            return $false
        }
    } catch {
        Write-Warning "Error getting auth token: $_"
        $script:AuthToken = "mock-jwt-token-for-testing"
        return $false
    }
}

# Test: List all resumes
function Test-ListResumes {
    Write-SubHeader "Test 1: List All Resumes"
    
    try {
        $response = curl.exe -s -X GET "$ApiBaseUrl/api/v1/resumes" `
            -H "Authorization: Bearer $script:AuthToken" `
            -H "Content-Type: application/json" `
            -m $Timeout 2>$null
        
        if ($response) {
            $data = $response | ConvertFrom-Json -ErrorAction SilentlyContinue
            Write-Success "Listed resumes"
            Write-Host ($data | ConvertTo-Json) -ForegroundColor Gray
            
            # Extract first resume ID
            if ($data -is [array] -and $data.Count -gt 0) {
                $script:FirstResumeId = $data[0].id
                Write-Success "Found resume ID: $script:FirstResumeId"
            } elseif ($data.id) {
                $script:FirstResumeId = $data.id
                Write-Success "Found resume ID: $script:FirstResumeId"
            } else {
                Write-Warning "No resumes found in list"
            }
        } else {
            Write-Error "Failed to list resumes"
        }
    } catch {
        Write-Error "Error listing resumes: $_"
    }
    
    Write-Host ""
}

# Test: Get single resume by ID
function Test-GetResume {
    Write-SubHeader "Test 2: Get Single Resume by ID"
    
    if (-not $script:FirstResumeId) {
        Write-Warning "Skipping: No resume ID available"
        return
    }
    
    try {
        $response = curl.exe -s -X GET "$ApiBaseUrl/api/v1/resumes/$script:FirstResumeId" `
            -H "Authorization: Bearer $script:AuthToken" `
            -H "Content-Type: application/json" `
            -m $Timeout 2>$null
        
        if ($response) {
            $data = $response | ConvertFrom-Json -ErrorAction SilentlyContinue
            Write-Success "Retrieved resume metadata"
            Write-Host ($data | ConvertTo-Json) -ForegroundColor Gray
        } else {
            Write-Error "Failed to get resume"
        }
    } catch {
        Write-Error "Error getting resume: $_"
    }
    
    Write-Host ""
}

# Test: Download resume by ID
function Test-DownloadResumeById {
    Write-SubHeader "Test 3: Download Resume by ID"
    
    if (-not $script:FirstResumeId) {
        Write-Warning "Skipping: No resume ID available"
        return
    }
    
    try {
        $outputFile = Join-Path $OutputDir "resume-by-id-$script:FirstResumeId.pdf"
        $headerFile = Join-Path $OutputDir "headers-by-id.txt"
        
        $response = curl.exe -s -X GET "$ApiBaseUrl/api/v1/resumes/$script:FirstResumeId/download" `
            -H "Authorization: Bearer $script:AuthToken" `
            -D $headerFile `
            -o $outputFile `
            -m $Timeout `
            -w "%{http_code}" 2>$null
        
        if ($response -eq "200") {
            $fileSize = (Get-Item $outputFile -ErrorAction SilentlyContinue).Length
            Write-Success "Downloaded resume by ID (HTTP $response, Size: $fileSize bytes)"
            Write-Info "Saved to: $outputFile"
            
            # Check response headers
            if (Test-Path $headerFile) {
                $headers = Get-Content $headerFile
                if ($headers -match "Content-Disposition") {
                    Write-Success "Content-Disposition header present"
                    $headers | Select-String "Content-Disposition"
                }
            }
        } else {
            Write-Warning "Failed to download resume by ID (HTTP $response)"
        }
    } catch {
        Write-Error "Error downloading resume: $_"
    }
    
    Write-Host ""
}

# Test: Download resume by user ID (public endpoint)
function Test-DownloadResumeByUser {
    Write-SubHeader "Test 4: Download Resume by User ID (Public Endpoint)"
    
    if (-not $script:FirstResumeId) {
        Write-Warning "Skipping: No resume ID available"
        return
    }
    
    try {
        # First get a resume to find the user ID
        $resumeInfo = curl.exe -s -X GET "$ApiBaseUrl/api/v1/resumes/$script:FirstResumeId" `
            -H "Authorization: Bearer $script:AuthToken" `
            -m $Timeout 2>$null
        
        $data = $resumeInfo | ConvertFrom-Json -ErrorAction SilentlyContinue
        $userId = $data.user_id -or $data.userId
        
        if (-not $userId) {
            Write-Warning "Could not extract user_id from resume"
            return
        }
        
        $outputFile = Join-Path $OutputDir "resume-by-user-$userId.pdf"
        
        $response = curl.exe -s -X GET "$ApiBaseUrl/resume/download?userId=$userId" `
            -o $outputFile `
            -m $Timeout `
            -w "%{http_code}" 2>$null
        
        if ($response -eq "200") {
            $fileSize = (Get-Item $outputFile -ErrorAction SilentlyContinue).Length
            Write-Success "Downloaded resume by user ID (HTTP $response, Size: $fileSize bytes)"
            Write-Info "Saved to: $outputFile"
        } else {
            Write-Warning "Failed to download resume by user ID (HTTP $response)"
        }
    } catch {
        Write-Error "Error downloading resume by user: $_"
    }
    
    Write-Host ""
}

# Test: Create a resume
function Test-CreateResume {
    Write-SubHeader "Test 5: Create Resume"
    
    try {
        $payload = @{
            title = "Test Resume"
            tags = @("test", "curl")
            is_main = $false
        } | ConvertTo-Json
        
        $response = curl.exe -s -X POST "$ApiBaseUrl/api/v1/resumes" `
            -H "Authorization: Bearer $script:AuthToken" `
            -H "Content-Type: application/json" `
            -d $payload `
            -m $Timeout 2>$null
        
        if ($response) {
            $data = $response | ConvertFrom-Json -ErrorAction SilentlyContinue
            Write-Success "Created resume"
            Write-Host ($data | ConvertTo-Json) -ForegroundColor Gray
            
            # Extract created resume ID
            $script:CreatedResumeId = $data.id
            if ($script:CreatedResumeId) {
                Write-Success "Created resume ID: $script:CreatedResumeId"
            }
        } else {
            Write-Error "Failed to create resume"
        }
    } catch {
        Write-Error "Error creating resume: $_"
    }
    
    Write-Host ""
}

# Test: List resumes with tag filter
function Test-ListResumesWithFilter {
    Write-SubHeader "Test 6: List Resumes with Tag Filter"
    
    try {
        $response = curl.exe -s -X GET "$ApiBaseUrl/api/v1/resumes?tags=test,curl" `
            -H "Authorization: Bearer $script:AuthToken" `
            -H "Content-Type: application/json" `
            -m $Timeout 2>$null
        
        if ($response) {
            $data = $response | ConvertFrom-Json -ErrorAction SilentlyContinue
            Write-Success "Listed filtered resumes"
            Write-Host ($data | ConvertTo-Json) -ForegroundColor Gray
        } else {
            Write-Warning "Failed to list filtered resumes"
        }
    } catch {
        Write-Error "Error listing filtered resumes: $_"
    }
    
    Write-Host ""
}

# Test: Generate resume
function Test-GenerateResume {
    Write-SubHeader "Test 7: Generate Resume"
    
    if (-not $script:CreatedResumeId) {
        Write-Warning "Skipping: No created resume ID available"
        return
    }
    
    try {
        $payload = @{
            resume_id = $script:CreatedResumeId
            template = "modern"
            format = "pdf"
        } | ConvertTo-Json
        
        $response = curl.exe -s -X POST "$ApiBaseUrl/api/v1/resumes/generate" `
            -H "Authorization: Bearer $script:AuthToken" `
            -H "Content-Type: application/json" `
            -d $payload `
            -m $Timeout 2>$null
        
        if ($response) {
            $data = $response | ConvertFrom-Json -ErrorAction SilentlyContinue
            Write-Success "Resume generation triggered"
            Write-Host ($data | ConvertTo-Json) -ForegroundColor Gray
            
            # Extract job ID
            $script:JobId = $data.job_id -or $data.id
            if ($script:JobId) {
                Write-Success "Job ID: $script:JobId"
            }
        } else {
            Write-Error "Failed to generate resume"
        }
    } catch {
        Write-Error "Error generating resume: $_"
    }
    
    Write-Host ""
}

# Test: Get job status
function Test-GetJobStatus {
    Write-SubHeader "Test 8: Get Resume Generation Job Status"
    
    if (-not $script:JobId) {
        Write-Warning "Skipping: No job ID available"
        return
    }
    
    try {
        for ($i = 1; $i -le 3; $i++) {
            Write-Info "Polling status (attempt $i/3)..."
            
            $response = curl.exe -s -X GET "$ApiBaseUrl/api/v1/resumes/jobs/$script:JobId" `
                -H "Authorization: Bearer $script:AuthToken" `
                -m $Timeout 2>$null
            
            if ($response) {
                $data = $response | ConvertFrom-Json -ErrorAction SilentlyContinue
                Write-Host ($data | ConvertTo-Json) -ForegroundColor Gray
                
                $status = $data.status
                if ($status -eq "completed") {
                    Write-Success "Resume generation completed!"
                    break
                } elseif ($status -eq "failed") {
                    Write-Error "Resume generation failed"
                    break
                }
            }
            
            if ($i -lt 3) {
                Start-Sleep -Seconds 2
            }
        }
    } catch {
        Write-Error "Error getting job status: $_"
    }
    
    Write-Host ""
}

# Test: Authentication error
function Test-AuthError {
    Write-SubHeader "Test 9: Authentication Error (No Token)"
    
    try {
        $response = curl.exe -s -X GET "$ApiBaseUrl/api/v1/resumes" `
            -H "Content-Type: application/json" `
            -m $Timeout `
            -w "%{http_code}" 2>$null | Select-Object -Last 1
        
        if ($response -eq "401" -or $response -eq "403") {
            Write-Success "Authentication properly enforced (HTTP $response)"
        } else {
            Write-Warning "Expected 401/403, got HTTP $response"
        }
    } catch {
        Write-Error "Error testing auth: $_"
    }
    
    Write-Host ""
}

# Test: 404 error
function Test-NotFoundError {
    Write-SubHeader "Test 10: Not Found Error (Non-existent Resume)"
    
    try {
        $response = curl.exe -s -X GET "$ApiBaseUrl/api/v1/resumes/nonexistent-id" `
            -H "Authorization: Bearer $script:AuthToken" `
            -m $Timeout `
            -w "%{http_code}" 2>$null | Select-Object -Last 1
        
        if ($response -eq "404") {
            Write-Success "404 error properly returned (HTTP $response)"
        } else {
            Write-Warning "Expected 404, got HTTP $response"
        }
    } catch {
        Write-Error "Error testing 404: $_"
    }
    
    Write-Host ""
}

# Test: Resume preview
function Test-PreviewResume {
    Write-SubHeader "Test 11: Resume Preview (Public Endpoint)"
    
    if (-not $script:FirstResumeId) {
        Write-Warning "Skipping: No resume ID available"
        return
    }
    
    try {
        $response = curl.exe -s -X GET "$ApiBaseUrl/resume/preview?resumeId=$script:FirstResumeId" `
            -m $Timeout 2>$null
        
        if ($response) {
            Write-Success "Retrieved resume preview"
        } else {
            Write-Warning "Failed to get preview"
        }
    } catch {
        Write-Error "Error getting preview: $_"
    }
    
    Write-Host ""
}

# Main execution
function Main {
    Write-Header "Resume Download & Retrieval Test Suite (PowerShell)"
    Write-Info "API Base URL: $ApiBaseUrl"
    Write-Info "Auth Service URL: $AuthServiceUrl"
    Write-Host ""
    
    Check-Curl
    Check-Services
    
    if (-not (Get-AuthToken)) {
        Write-Warning "Using mock auth token - some tests may fail"
    }
    
    Write-Host ""
    Test-ListResumes
    Test-GetResume
    Test-DownloadResumeById
    Test-DownloadResumeByUser
    Test-CreateResume
    Test-ListResumesWithFilter
    Test-GenerateResume
    Test-GetJobStatus
    Test-AuthError
    Test-NotFoundError
    Test-PreviewResume
    
    Write-Header "Test Summary"
    Write-Info "Downloaded files saved to: $OutputDir"
    Write-Success "Test suite completed"
}

# Run main
Main
