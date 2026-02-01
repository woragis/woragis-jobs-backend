# Resume Download & Retrieval Test Suite

Comprehensive curl-based testing suite for resume API endpoints including download, retrieval, and generation workflows.

## Overview

This test suite validates all resume API endpoints:

- Resume retrieval (list, get single)
- Resume download (by ID and by user ID)
- Resume generation workflow
- Generated resume retrieval
- Job status tracking
- Authentication and error handling

## Test Files

### Linux/Mac

```bash
test-resume-download.sh
```

### Windows

```powershell
test-resume-download.ps1
```

## Setup

### Prerequisites

1. **Services Running** - Ensure these services are running:
   - Jobs Backend: `docker-compose up -d` (port 3000)
   - Auth Service: Running on port 3001 (or configured URL)
   - PostgreSQL: Database backing the jobs service
   - Redis: Cache/session storage (optional)

2. **curl/PowerShell** - Must be available in your PATH
   - Linux/Mac: `curl --version`
   - Windows: `curl.exe --version` (part of Windows 10+)

3. **Environment Variables** (Optional):
   ```bash
   export API_BASE_URL=http://localhost:3000
   export AUTH_SERVICE_URL=http://localhost:3001
   export TEST_USER_EMAIL=test@example.com
   export TEST_USER_PASSWORD=testpassword123
   ```

### Quick Start

#### Linux/Mac

```bash
cd backend/jobs
chmod +x test-resume-download.sh
./test-resume-download.sh
```

#### Windows PowerShell

```powershell
cd backend/jobs
.\test-resume-download.ps1
```

Or set execution policy if needed:

```powershell
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

## Test Scenarios

### Test 1: List All Resumes

**Endpoint:** `GET /api/v1/resumes`

Lists all resumes for the authenticated user with pagination support.

**What it tests:**

- Resume list retrieval
- Authentication requirement
- Response format

**Example Response:**

```json
[
  {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "title": "Senior Developer Resume",
    "user_id": "user-123",
    "file_name": "resume.pdf",
    "tags": ["python", "go", "devops"],
    "is_main": true,
    "status": "published",
    "created_at": "2024-01-15T10:30:00Z"
  }
]
```

### Test 2: Get Single Resume

**Endpoint:** `GET /api/v1/resumes/:id`

Retrieves metadata for a specific resume by ID.

**What it tests:**

- Resume metadata retrieval
- Resume ID resolution
- 404 handling for non-existent resumes

**Example Response:**

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "title": "Senior Developer Resume",
  "user_id": "user-123",
  "file_name": "resume.pdf",
  "file_size": 245000,
  "tags": ["python", "go", "devops"],
  "is_main": true,
  "is_featured": false,
  "status": "published",
  "applications_used": 5,
  "interview_rate": 0.4,
  "offer_rate": 0.2,
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-20T14:22:00Z"
}
```

### Test 3: Download Resume by ID

**Endpoint:** `GET /api/v1/resumes/:id/download`

Downloads the resume file for a specific resume ID.

**What it tests:**

- File download functionality
- Content-Disposition headers
- File integrity
- Binary response handling

**HTTP Headers:**

```
Content-Type: application/pdf
Content-Disposition: attachment; filename="resume.pdf"
Content-Length: 245000
```

**Output:**

- File saved to `test-downloads/resume-by-id-{id}.pdf`
- Headers saved to `test-downloads/headers-by-id.txt`

### Test 4: Download Resume by User ID (Public)

**Endpoint:** `GET /resume/download?userId={id}`

Downloads the main resume for a specific user (public endpoint).

**What it tests:**

- Public endpoint access (no auth)
- User ID parameter resolution
- Main resume selection

**Output:**

- File saved to `test-downloads/resume-by-user-{userId}.pdf`
- Headers saved to `test-downloads/headers-by-user.txt`

### Test 5: Create Resume

**Endpoint:** `POST /api/v1/resumes`

Creates a new resume entry.

**What it tests:**

- Resume creation
- Field validation
- Response contains ID

**Request Payload:**

```json
{
  "title": "Test Resume",
  "tags": ["test", "curl"],
  "is_main": false
}
```

### Test 6: List Resumes with Filter

**Endpoint:** `GET /api/v1/resumes?tags=test,curl`

Lists resumes filtered by tags.

**What it tests:**

- Query parameter filtering
- Multiple tag filtering

### Test 7: Generate Resume

**Endpoint:** `POST /api/v1/resumes/generate`

Triggers resume generation from the resume-generator service.

**What it tests:**

- Asynchronous job creation
- Job ID tracking
- Integration with resume-generator

**Request Payload:**

```json
{
  "resume_id": "550e8400-e29b-41d4-a716-446655440000",
  "template": "modern",
  "format": "pdf"
}
```

**Example Response:**

```json
{
  "job_id": "job-123",
  "resume_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "pending",
  "created_at": "2024-01-20T15:00:00Z"
}
```

### Test 8: Get Job Status

**Endpoint:** `GET /api/v1/resumes/jobs/:id`

Polls the status of a resume generation job.

**What it tests:**

- Job status tracking
- Polling mechanism
- Job completion detection

**Example Response:**

```json
{
  "job_id": "job-123",
  "resume_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "completed",
  "generated_file_path": "/resume-data/generated/resume-123.pdf",
  "output_format": "pdf",
  "file_size": 256000,
  "created_at": "2024-01-20T15:00:00Z",
  "completed_at": "2024-01-20T15:02:30Z"
}
```

### Test 9: Authentication Error

**Endpoint:** `GET /api/v1/resumes` (without token)

Tests that endpoints properly reject requests without authentication.

**What it tests:**

- 401/403 error handling
- Authentication requirement enforcement

**Expected:** HTTP 401 or 403

### Test 10: Not Found Error

**Endpoint:** `GET /api/v1/resumes/nonexistent-id`

Tests 404 error for non-existent resources.

**What it tests:**

- 404 error handling
- Proper error messages

**Expected:** HTTP 404

### Test 11: Resume Preview (Public)

**Endpoint:** `GET /resume/preview?resumeId={id}`

Retrieves preview/metadata for a public resume.

**What it tests:**

- Public preview endpoint
- Resume metadata exposure

## Output and Logs

### Downloaded Files

All downloaded files are saved to `test-downloads/` directory:

- `resume-by-id-{id}.pdf` - Resume downloaded by ID
- `resume-by-user-{userId}.pdf` - Resume downloaded by user
- `headers-by-id.txt` - Response headers from ID download
- `headers-by-user.txt` - Response headers from user download

### Console Output

Tests produce color-coded output:

- 🟢 Green: Success
- 🟡 Yellow: Warning/Skipped
- 🔴 Red: Error
- 🔵 Blue: Headers/sections

Example:

```
✓ curl found
✓ Jobs Service is running
✓ Auth token obtained: mock-jwt-token-for...
--- Test 1: List All Resumes ---
✓ Listed resumes
✓ Found resume ID: 550e8400-e29b-41d4-a716-446655440000
```

## Troubleshooting

### Services Not Running

**Error:** "Jobs Service is not responding"

**Solution:**

```bash
cd backend/jobs
docker-compose up -d
# Wait for services to start
docker-compose logs -f
```

### Authentication Failures

**Error:** HTTP 401 or 403 on protected endpoints

**Solution 1:** Provide valid credentials

```bash
export TEST_USER_EMAIL=your-real-email@example.com
export TEST_USER_PASSWORD=your-real-password
./test-resume-download.sh
```

**Solution 2:** Check auth service is running

```bash
curl http://localhost:3001/api/v1/health
```

### Download Failures

**Error:** HTTP 404 or 500 on download

**Check:**

1. Resume exists: `curl http://localhost:3000/api/v1/resumes -H "Authorization: Bearer $TOKEN"`
2. File exists in storage: `ls -la ~/resume-data/`
3. Permissions: `stat ~/resume-data/resume-*.pdf`

### curl Not Found (Windows)

**Error:** "curl not found"

**Solutions:**

1. Use Windows 10+ (curl.exe pre-installed)
2. Install Git Bash or WSL
3. Install curl via Chocolatey:
   ```powershell
   choco install curl
   ```

## API Endpoint Summary

| Method | Endpoint                       | Protected | Purpose             |
| ------ | ------------------------------ | --------- | ------------------- |
| GET    | `/api/v1/resumes`              | Yes       | List resumes        |
| POST   | `/api/v1/resumes`              | Yes       | Create resume       |
| GET    | `/api/v1/resumes/:id`          | Yes       | Get resume          |
| PUT    | `/api/v1/resumes/:id`          | Yes       | Update resume       |
| DELETE | `/api/v1/resumes/:id`          | Yes       | Delete resume       |
| GET    | `/api/v1/resumes/:id/download` | Yes       | Download by ID      |
| POST   | `/api/v1/resumes/generate`     | Yes       | Generate resume     |
| GET    | `/api/v1/resumes/jobs/:id`     | Yes       | Get job status      |
| GET    | `/resume/download`             | No        | Download by user ID |
| GET    | `/resume/preview`              | No        | Preview resume      |

## Configuration

### Environment Variables

| Variable             | Default                 | Description        |
| -------------------- | ----------------------- | ------------------ |
| `API_BASE_URL`       | `http://localhost:3000` | Jobs service URL   |
| `AUTH_SERVICE_URL`   | `http://localhost:3001` | Auth service URL   |
| `TEST_USER_EMAIL`    | `test@example.com`      | Test user email    |
| `TEST_USER_PASSWORD` | `testpassword123`       | Test user password |

### Custom Configuration

**Linux/Mac:**

```bash
API_BASE_URL=http://api.example.com AUTH_SERVICE_URL=http://auth.example.com ./test-resume-download.sh
```

**Windows:**

```powershell
$env:API_BASE_URL = "http://api.example.com"
$env:AUTH_SERVICE_URL = "http://auth.example.com"
.\test-resume-download.ps1
```

## Integration with CI/CD

### GitHub Actions

```yaml
- name: Run Resume Tests
  run: |
    cd backend/jobs
    bash test-resume-download.sh
  env:
    API_BASE_URL: http://localhost:3000
    TEST_USER_EMAIL: ci-test@example.com
    TEST_USER_PASSWORD: ${{ secrets.TEST_PASSWORD }}
```

### GitLab CI

```yaml
test_resume_api:
  stage: test
  script:
    - cd backend/jobs
    - bash test-resume-download.sh
  services:
    - docker:dind
  variables:
    API_BASE_URL: http://localhost:3000
```

## Performance Metrics

After running tests, check downloaded file sizes and generation times:

```bash
ls -lh test-downloads/
# Example output:
# -rw-r--r-- 245K Jan 20 15:02 resume-by-id-550e8400.pdf
# -rw-r--r-- 245K Jan 20 15:03 resume-by-user-user-123.pdf
```

## Next Steps

After validating API endpoints with these tests:

1. **Implement Advanced Features:**
   - Resume batch download/ZIP export
   - Resume comparison
   - Enhanced metrics tracking

2. **Optimize Performance:**
   - Caching strategies
   - File storage optimization
   - Generation speed improvements

3. **Enhance Generated Resumes:**
   - Multiple template support
   - Advanced formatting
   - Custom styling

4. **Integration Testing:**
   - Test with frontend workflows
   - E2E testing with Playwright
   - Load testing

## Support

For issues or questions:

1. Check the troubleshooting section
2. Review API logs: `docker-compose logs -f jobs-backend`
3. Check PostgreSQL: `docker-compose logs -f postgres`
4. Monitor HTTP publisher integration
