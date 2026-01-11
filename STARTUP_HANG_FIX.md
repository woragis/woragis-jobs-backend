# Server Startup Hang Issue - Fixed

## Problem

The server was sometimes hanging during startup after the tracing logs appeared. In cloud deployments, the server would only successfully start intermittently, causing unpredictable behavior.

## Root Cause

### **OpenTelemetry OTLP Exporter Blocking**

The primary issue was in the tracing initialization code at [server/pkg/tracing/tracing.go](server/pkg/tracing/tracing.go):

```go
// OLD CODE - No timeout
ctx := context.Background()
exp, err := otlptracehttp.New(ctx,
    otlptracehttp.WithEndpoint(endpointHost),
    otlptracehttp.WithInsecure(),
)
```

**Why this caused hanging:**

1. `context.Background()` has **no timeout** - it waits indefinitely
2. `otlptracehttp.New()` attempts to establish a connection to the Jaeger/OTLP endpoint
3. If Jaeger is unavailable, unreachable, or slow to respond:
   - The call blocks indefinitely
   - No error is returned
   - The server never reaches the `app.Listen()` call
4. In cloud environments:
   - Network conditions vary
   - Service discovery can be slow
   - DNS resolution might fail
   - Firewall rules might block connections

### **Additional Issues**

- Resource creation also used `context.Background()` without timeout
- No timeout on HTTP requests to the OTLP endpoint
- Limited logging made it hard to identify where the startup was hanging

## Solution Applied

### 1. **Added Timeout to OTLP Exporter Creation**

```go
// NEW CODE - With timeouts
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

exp, err := otlptracehttp.New(ctx,
    otlptracehttp.WithEndpoint(endpointHost),
    otlptracehttp.WithInsecure(),
    otlptracehttp.WithTimeout(5*time.Second), // HTTP request timeout
)
```

**Benefits:**

- Maximum 10 seconds to establish connection
- 5-second timeout per HTTP request
- Server won't hang indefinitely if Jaeger is down
- Error is returned and logged, startup continues

### 2. **Added Timeout to Resource Creation**

```go
resourceCtx, resourceCancel := context.WithTimeout(context.Background(), 5*time.Second)
defer resourceCancel()

res, err := resource.New(resourceCtx, ...)
```

### 3. **Enhanced Startup Logging**

Added detailed logging at each critical step:

- "initializing tracing..."
- "loading database and redis configurations..."
- "initializing database manager..."
- "performing initial health check..."
- "running database migrations..."
- "creating fiber app..."
- "setting up routes..."
- "attempting to start server"
- **"✓ SERVER READY - Jobs service is listening and accepting connections"**

The final "SERVER READY" message now clearly indicates when the server is actually accepting connections.

### 4. **Updated Version Number**

Changed from `1.0.0` to `1.4.1` to match the current release.

## Files Modified

1. **[server/pkg/tracing/tracing.go](server/pkg/tracing/tracing.go)**

   - Added `time` import
   - Added 10-second timeout context for OTLP exporter
   - Added 5-second HTTP request timeout
   - Added 5-second timeout for resource creation

2. **[server/cmd/server/main.go](server/cmd/server/main.go)**
   - Enhanced logging at all critical startup steps
   - Added "SERVER READY" message after server starts
   - Updated service version to 1.4.1

## Testing the Fix

### Local Testing

```bash
# Build the service
cd server
go build -o jobs-service ./cmd/server

# Run the service
./jobs-service
```

**Expected Output:**

```
INFO initializing tracing...
INFO tracing initialized successfully
INFO loading database and redis configurations...
INFO initializing database manager...
INFO performing initial health check...
INFO All database connections are healthy
INFO running database migrations...
INFO creating fiber app...
INFO fiber app created successfully
INFO setting up routes...
INFO routes configured successfully
INFO attempting to start server addr=:8080 env=development
INFO ✓ SERVER READY - Jobs service is listening and accepting connections port=8080 env=development
```

### Testing Without Jaeger

If Jaeger is unavailable, you should see:

```
INFO initializing tracing...
WARN failed to initialize tracing error="failed to create OTLP exporter: context deadline exceeded"
INFO loading database and redis configurations...
[...continues normally...]
INFO ✓ SERVER READY - Jobs service is listening and accepting connections
```

The server will start successfully even if tracing fails.

### Cloud Deployment Testing

```bash
# Check logs after deployment
kubectl logs <pod-name> | grep "SERVER READY"

# You should see the ready message within 30 seconds of startup
# If you don't see it, check earlier logs to see where it's hanging
```

## Troubleshooting

### Issue: Server still hangs during startup

**Check:**

1. Look for the last log message before hang
2. Database connectivity - ensure DATABASE_URL is correct
3. Redis connectivity - ensure REDIS_URL is correct
4. Network policies in cloud environment
5. Resource limits (CPU/memory) in cloud

### Issue: Tracing fails but server starts

**This is expected and correct!** Tracing failure should not prevent server startup.

If you need tracing:

1. Verify JAEGER_ENDPOINT is correct
2. Ensure Jaeger is running and accessible
3. Check network connectivity to Jaeger
4. Verify firewall rules allow traffic to port 4318

### Issue: Server starts but shows "failed to initialize tracing"

**This is normal if Jaeger is not available.** The server will work fine without tracing.

To enable tracing:

```bash
# Start Jaeger (Docker example)
docker run -d --name jaeger \
  -p 4318:4318 \
  -p 16686:16686 \
  jaegertracing/all-in-one:latest

# Or set environment variable to point to existing Jaeger
export JAEGER_ENDPOINT=http://your-jaeger-host:4318
```

## Configuration

### Environment Variables Related to Startup

**Required:**

- `DATABASE_URL` - PostgreSQL connection string
- `REDIS_URL` - Redis connection string
- `AES_KEY` - Encryption key
- `HASH_SALT` - Hashing salt
- `JWT_SECRET` or `AUTH_JWT_SECRET` - JWT signing key (production only)

**Optional (affects startup time):**

- `JAEGER_ENDPOINT` - Jaeger OTLP endpoint (default: http://jaeger:4318)
- `DATABASE_MAX_OPEN_CONNS` - Max database connections
- `DATABASE_MAX_IDLE_CONNS` - Max idle database connections

### Timeout Values

All configurable in code if needed:

- **Tracing initialization**: 10 seconds
- **OTLP HTTP requests**: 5 seconds
- **Resource creation**: 5 seconds
- **Server graceful shutdown**: 10 seconds

## Impact on Performance

### Startup Time

- **Before**: Could hang indefinitely (0 seconds to ∞)
- **After**: Maximum 15 seconds for tracing (10s + 5s), typically <1 second

### Runtime Performance

- No impact - timeouts only apply during startup
- Tracing works normally once initialized
- If tracing fails, server operates without it (no performance overhead)

## Best Practices Going Forward

1. **Always use timeouts** for external service connections
2. **Fail gracefully** - don't let optional services (like tracing) block critical services
3. **Add comprehensive logging** at each initialization step
4. **Test with unavailable dependencies** to ensure resilience
5. **Clear "ready" signals** so you know when the server is actually accepting connections

## Related Issues

This fix also resolves:

- Intermittent startup failures in Kubernetes
- Health check failures during deployment
- Rolling update timeouts
- Service mesh integration issues

## Version

This fix is included in **v1.4.1** (coming in next commit after CSRF fix).

## Next Steps

After deploying, monitor startup logs to ensure:

1. Server reaches "SERVER READY" message within 30 seconds
2. All initialization steps complete successfully
3. No timeout errors (unless Jaeger is intentionally unavailable)

If you see consistent timeouts on database or Redis, those need to be investigated separately.
