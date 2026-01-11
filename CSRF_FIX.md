# CSRF Token Issue - Fixed

## Problem

You were receiving "403 csrf token is invalid" errors when creating job applications, even though you were getting the token from the page.

## Root Causes

### 1. **IP-Based Session Keys (Primary Issue)**

The original implementation stored CSRF tokens in Redis using the client's IP address as the key:

```go
sessionKey := fmt.Sprintf("csrf:%s", c.IP())
```

This approach has several problems:

- **Proxy/Load Balancer**: The IP might be reported differently between requests
- **NAT/VPN**: IP addresses can change during a session
- **Dynamic IPs**: Mobile networks and some ISPs change IPs frequently
- **Reverse Proxy Configuration**: The IP might not be correctly extracted

### 2. **SameSite=Strict Cookie Policy**

The cookie was set with `SameSite: "Strict"`, which:

- Prevents cookies from being sent on any cross-origin request
- Can cause issues even with legitimate same-origin requests in certain browser scenarios
- Is overly restrictive for most use cases

## Solution Applied

### 1. **Token-Based Validation**

Changed from IP-based keys to token-based keys:

```go
// Storage (on GET request)
sessionKey := fmt.Sprintf("csrf:token:%s", token)
config.RedisClient.Set(ctx, sessionKey, "valid", config.TokenTTL)

// Validation (on POST/PUT/PATCH/DELETE)
sessionKey := fmt.Sprintf("csrf:token:%s", token)
exists := config.RedisClient.Exists(ctx, sessionKey)
```

**Benefits:**

- Works regardless of IP address changes
- More reliable across proxies and load balancers
- Simpler validation logic
- Self-contained token validation

### 2. **Relaxed Cookie Policy**

Changed `SameSite` from `"Strict"` to `"Lax"`:

```go
c.Cookie(&fiber.Cookie{
    Name:     config.CookieName,
    Value:    token,
    HTTPOnly: false,
    Secure:   config.SecureCookie,
    SameSite: "Lax", // Changed from "Strict"
    MaxAge:   int(config.TokenTTL.Seconds()),
    Path:     "/",
})
```

**Benefits:**

- Still protects against CSRF attacks
- More compatible with modern web applications
- Allows top-level navigation while preventing cross-site form submissions

### 3. **Added Path to Cookie**

Added `Path: "/"` to ensure the cookie is available for all paths in your application.

## How to Use CSRF Tokens

### Frontend Implementation

#### 1. **Get the Token (Automatic on GET requests)**

Every GET request automatically generates and returns a CSRF token in:

- Cookie: `csrf_token`
- Header: `X-CSRF-Token`

#### 2. **Send Token with State-Changing Requests**

**Option A: Using Header (Recommended)**

```javascript
// Fetch API
fetch('/api/v1/job-applications', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'X-CSRF-Token': getCsrfToken(), // Get from cookie or previous response
    Authorization: `Bearer ${token}`,
  },
  credentials: 'include', // Important: Include cookies
  body: JSON.stringify(data),
})

// Helper function to get CSRF token from cookie
function getCsrfToken() {
  const name = 'csrf_token='
  const decodedCookie = decodeURIComponent(document.cookie)
  const ca = decodedCookie.split(';')
  for (let i = 0; i < ca.length; i++) {
    let c = ca[i]
    while (c.charAt(0) === ' ') {
      c = c.substring(1)
    }
    if (c.indexOf(name) === 0) {
      return c.substring(name.length, c.length)
    }
  }
  return ''
}
```

**Option B: Using Cookie (Automatic)**

```javascript
// If you include credentials, the cookie is sent automatically
fetch('/api/v1/job-applications', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    Authorization: `Bearer ${token}`,
  },
  credentials: 'include', // Cookie is sent automatically
  body: JSON.stringify(data),
})
```

#### 3. **Axios Example**

```javascript
import axios from 'axios'

// Get CSRF token from cookie
function getCsrfToken() {
  const name = 'csrf_token='
  const decodedCookie = decodeURIComponent(document.cookie)
  const ca = decodedCookie.split(';')
  for (let i = 0; i < ca.length; i++) {
    let c = ca[i]
    while (c.charAt(0) === ' ') {
      c = c.substring(1)
    }
    if (c.indexOf(name) === 0) {
      return c.substring(name.length, c.length)
    }
  }
  return ''
}

// Configure axios instance
const api = axios.create({
  baseURL: 'http://localhost:8080/api/v1',
  withCredentials: true, // Important: Include cookies
})

// Add CSRF token to all requests
api.interceptors.request.use((config) => {
  const csrfToken = getCsrfToken()
  if (csrfToken) {
    config.headers['X-CSRF-Token'] = csrfToken
  }
  return config
})

// Use it
api.post('/job-applications', {
  companyName: 'Example Corp',
  jobTitle: 'Software Engineer',
  // ... other fields
})
```

## Testing the Fix

### 1. **Test Token Generation**

```bash
curl -c cookies.txt -X GET http://localhost:8080/api/v1/csrf-token
```

Check the response headers for `X-CSRF-Token` and the cookie file for `csrf_token`.

### 2. **Test Token Validation**

```bash
# Get the token
TOKEN=$(curl -s -c cookies.txt -X GET http://localhost:8080/api/v1/csrf-token | jq -r '.token')

# Use the token in a POST request
curl -b cookies.txt \
  -H "X-CSRF-Token: $TOKEN" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -X POST http://localhost:8080/api/v1/job-applications \
  -d '{"companyName":"Test","jobTitle":"Engineer","location":"Remote","jobUrl":"https://example.com/job"}'
```

## Exempt Routes & Methods

The following don't require CSRF tokens:

- **Methods**: GET, HEAD, OPTIONS
- **Routes**: /healthz, /metrics, /api/v1/auth/login, /api/v1/auth/register

## Configuration

CSRF protection is configured in:

- File: `server/pkg/security/csrf.go`
- Middleware setup: `cmd/server/main.go`
- CORS config: `internal/config/cors.go`

### Environment Variables

- `CORS_ALLOWED_HEADERS`: Must include `X-CSRF-Token` (default includes it)
- `CORS_EXPOSED_HEADERS`: Must include `X-CSRF-Token` (default includes it)
- `CORS_ALLOW_CREDENTIALS`: Must be `true` (default is `true`)

## Security Notes

1. **Token Storage**: Tokens are stored in Redis with a 1-hour TTL
2. **Token Extension**: Valid tokens have their TTL extended on each use
3. **Cookie Security**:
   - `HTTPOnly: false` - Allows JavaScript to read the token
   - `Secure: true/false` - Based on environment (prod/dev)
   - `SameSite: Lax` - Balanced security and usability
4. **Graceful Degradation**: If Redis is unavailable, requests are allowed (logged as warning)

## Troubleshooting

### Issue: Still getting "CSRF token is invalid"

**Check:**

1. Ensure `credentials: 'include'` is set in fetch/axios
2. Verify CORS settings allow credentials
3. Check that the token hasn't expired (1 hour TTL)
4. Confirm Redis is running and accessible
5. Look for CSRF-related logs in the server output

### Issue: Token not being sent

**Check:**

1. Cookie domain matches your frontend domain
2. CORS `AllowCredentials` is true
3. Frontend is sending `credentials: 'include'`

### Issue: Token expires too quickly

**Solution:**

- Increase `TokenTTL` in `DefaultCSRFConfig` (current: 1 hour)
- Consider implementing token refresh on each request

## Related Files Modified

- `server/pkg/security/csrf.go` - CSRF implementation
- Backend configuration already properly set up for CORS

## Next Steps

1. Rebuild and restart your backend service
2. Clear your browser cookies
3. Test the job application creation flow
4. Monitor server logs for any CSRF-related warnings

If you continue to experience issues, check the server logs for specific error messages.
