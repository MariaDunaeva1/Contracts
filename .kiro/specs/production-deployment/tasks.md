# Production Deployment - Implementation Tasks

## Phase 1: Core Infrastructure (2-3 hours)

### Task 1.1: Create Multi-Stage Backend Dockerfile
**File:** `backend/Dockerfile.prod`

**Steps:**
1. Create builder stage with Go 1.23
2. Copy and download dependencies
3. Build optimized binary with `-ldflags="-w -s"`
4. Create runtime stage with Alpine
5. Add non-root user (appuser)
6. Copy binary and templates
7. Set proper permissions
8. Expose port 8080

**Acceptance:**
- Image size < 200MB
- Binary runs as non-root user
- Build completes in < 3 minutes

### Task 1.2: Create Production Docker Compose
**File:** `docker-compose.prod.yml`

**Steps:**
1. Add health checks to all services
2. Configure restart policies (unless-stopped)
3. Add network isolation (backend/frontend networks)
4. Configure volume persistence
5. Add depends_on with conditions
6. Remove development volumes
7. Add resource limits

**Acceptance:**
- All services have health checks
- Services restart on failure
- Networks properly isolated
- Volumes persist data

### Task 1.3: Create Environment Template
**File:** `.env.example`

**Steps:**
1. Document all required variables
2. Add descriptions for each variable
3. Provide example values
4. Group by category
5. Mark required vs optional

**Acceptance:**
- All environment variables documented
- Clear descriptions provided
- Example values are safe (no real secrets)

## Phase 2: Monitoring & Observability (2-3 hours)

### Task 2.1: Implement Structured Logging
**File:** `backend/internal/logger/logger.go`

**Steps:**
1. Add uber-go/zap dependency
2. Create logger initialization function
3. Configure JSON format for production
4. Add log levels (debug, info, warn, error)
5. Create helper functions for common log patterns

**Acceptance:**
- Logs output in JSON format
- Log level configurable via env var
- No sensitive data in logs

### Task 2.2: Enhance Health Check Endpoint
**File:** `backend/cmd/server/main.go`

**Steps:**
1. Create HealthResponse struct
2. Check database connectivity with timeout
3. Check MinIO connectivity
4. Check worker pool status
5. Return 503 if any critical service down
6. Add version and uptime info

**Acceptance:**
- Returns 200 when all services healthy
- Returns 503 when any service down
- Response time < 100ms
- Includes all service statuses

### Task 2.3: Add Metrics Endpoint
**File:** `backend/internal/metrics/metrics.go`

**Steps:**
1. Add prometheus/client_golang dependency
2. Create metrics collectors
3. Track HTTP requests (count, duration)
4. Track worker pool stats
5. Track database connection pool
6. Expose /api/v1/metrics endpoint

**Acceptance:**
- Metrics in Prometheus format
- HTTP metrics include method, path, status
- Worker metrics show active/queued jobs
- Database metrics show connection pool usage

### Task 2.4: Add Request Logging Middleware
**File:** `backend/internal/middleware/logging.go`

**Steps:**
1. Create logging middleware
2. Log request method, path, status
3. Log request duration
4. Log user agent
5. Log error details on failures
6. Skip health check endpoint

**Acceptance:**
- All requests logged with duration
- Errors include stack traces
- Health checks not logged (too noisy)

## Phase 3: Security Hardening (2 hours)

### Task 3.1: Implement Rate Limiting
**File:** `backend/internal/middleware/ratelimit.go`

**Steps:**
1. Add golang.org/x/time/rate dependency
2. Create rate limiter per IP
3. Global limit: 100 req/min
4. Expensive endpoints: 10 req/min
5. Return 429 when limit exceeded
6. Add Retry-After header

**Acceptance:**
- Rate limiting works per IP
- Different limits for different endpoints
- Returns proper 429 status
- Includes Retry-After header

### Task 3.2: Harden CORS Configuration
**File:** `backend/cmd/server/main.go`

**Steps:**
1. Add gin-contrib/cors dependency
2. Load allowed origins from env var
3. Restrict methods to needed ones
4. Set proper headers
5. Configure credentials handling
6. Set max age for preflight cache

**Acceptance:**
- CORS only allows configured origins
- Credentials handled properly
- Preflight requests cached

### Task 3.3: Add Request Size Limits
**File:** `backend/internal/middleware/sizelimit.go`

**Steps:**
1. Create middleware for request size limiting
2. Set default limit to 10MB
3. Make limit configurable via env var
4. Return 413 when exceeded
5. Apply to all routes

**Acceptance:**
- Requests > 10MB rejected
- Returns 413 status
- Limit configurable

### Task 3.4: Add Input Validation
**File:** `backend/internal/validator/validator.go`

**Steps:**
1. Add go-playground/validator dependency
2. Create validation helper functions
3. Validate all request bodies
4. Return 400 with clear error messages
5. Sanitize file paths

**Acceptance:**
- Invalid requests return 400
- Error messages are clear
- No path traversal vulnerabilities

## Phase 4: Reliability (1-2 hours)

### Task 4.1: Implement Graceful Shutdown
**File:** `backend/cmd/server/main.go`

**Steps:**
1. Create signal handler for SIGINT/SIGTERM
2. Stop accepting new jobs
3. Wait for in-flight requests (30s timeout)
4. Close database connections
5. Close MinIO connections
6. Log shutdown completion

**Acceptance:**
- Server shuts down gracefully
- In-flight requests complete
- No data loss on shutdown
- Timeout prevents hanging

### Task 4.2: Configure Database Connection Pooling
**File:** `backend/internal/database/db.go`

**Steps:**
1. Set MaxOpenConns to 25
2. Set MaxIdleConns to 5
3. Set ConnMaxLifetime to 5 minutes
4. Make values configurable via env vars
5. Add connection pool metrics

**Acceptance:**
- Connection pool properly configured
- Connections recycled after lifetime
- Pool size configurable

### Task 4.3: Add Response Compression
**File:** `backend/cmd/server/main.go`

**Steps:**
1. Add gin-contrib/gzip dependency
2. Enable gzip middleware
3. Configure compression level
4. Skip compression for SSE endpoints
5. Add Vary header

**Acceptance:**
- Responses compressed with gzip
- SSE endpoints not compressed
- Proper headers set

### Task 4.4: Optimize Frontend Nginx Config
**File:** `frontend/nginx.prod.conf`

**Steps:**
1. Add gzip compression
2. Configure static asset caching
3. Add security headers
4. Configure proxy timeouts
5. Add rate limiting

**Acceptance:**
- Static assets cached for 1 year
- Gzip enabled for text files
- Security headers present
- Proxy handles long requests

## Phase 5: Documentation (2 hours)

### Task 5.1: Create Deployment Guide for Oracle Cloud
**File:** `docs/DEPLOY_ORACLE_CLOUD.md`

**Steps:**
1. Document VM creation
2. Document Docker installation
3. Document firewall configuration
4. Document SSL setup with Let's Encrypt
5. Document deployment steps
6. Add troubleshooting section

**Acceptance:**
- Complete step-by-step guide
- Includes all commands
- Covers common issues

### Task 5.2: Create Deployment Guide for Railway
**File:** `docs/DEPLOY_RAILWAY.md`

**Steps:**
1. Document GitHub connection
2. Document environment variable setup
3. Document deployment process
4. Document domain configuration
5. Add cost estimation

**Acceptance:**
- Complete step-by-step guide
- Includes screenshots
- Covers pricing

### Task 5.3: Create Production Checklist
**File:** `docs/PRODUCTION_CHECKLIST.md`

**Steps:**
1. Pre-deployment checklist
2. Deployment steps
3. Post-deployment validation
4. Monitoring setup
5. Backup configuration
6. Security verification

**Acceptance:**
- Comprehensive checklist
- Can be used as runbook
- Covers all critical items

### Task 5.4: Create Troubleshooting Guide
**File:** `docs/TROUBLESHOOTING_PRODUCTION.md`

**Steps:**
1. Common deployment issues
2. Service health check failures
3. Database connection issues
4. MinIO connection issues
5. Worker pool issues
6. Performance problems

**Acceptance:**
- Covers common issues
- Includes diagnostic commands
- Provides solutions

### Task 5.5: Update Main README
**File:** `README.md`

**Steps:**
1. Add production deployment section
2. Link to deployment guides
3. Add architecture diagram
4. Update requirements
5. Add monitoring section

**Acceptance:**
- README reflects production setup
- Links to all guides
- Clear and concise

## Testing Tasks

### Task T.1: Create Production Test Script
**File:** `scripts/test_production.sh`

**Steps:**
1. Test all services start
2. Test health checks pass
3. Test API endpoints
4. Test frontend loads
5. Test worker processes job
6. Test metrics endpoint
7. Test logs format

**Acceptance:**
- Script tests all critical paths
- Returns non-zero on failure
- Can run in CI/CD

### Task T.2: Create Load Test Script
**File:** `scripts/load_test.sh`

**Steps:**
1. Use Apache Bench or similar
2. Test API endpoints under load
3. Verify rate limiting works
4. Check response times
5. Monitor resource usage

**Acceptance:**
- Tests realistic load
- Verifies performance targets
- Identifies bottlenecks

## Implementation Order

Execute tasks in this order for optimal workflow:

**Day 1 (4-5 hours):**
- Task 1.1: Multi-stage Dockerfile
- Task 1.2: Production docker-compose
- Task 1.3: Environment template
- Task 2.1: Structured logging
- Task 2.2: Enhanced health check

**Day 2 (4-5 hours):**
- Task 2.3: Metrics endpoint
- Task 2.4: Request logging
- Task 3.1: Rate limiting
- Task 3.2: CORS hardening
- Task 3.3: Request size limits

**Day 3 (3-4 hours):**
- Task 4.1: Graceful shutdown
- Task 4.2: Connection pooling
- Task 4.3: Response compression
- Task 4.4: Nginx optimization
- Task T.1: Production test script

**Day 4 (2-3 hours):**
- Task 5.1: Oracle Cloud guide
- Task 5.2: Railway guide
- Task 5.3: Production checklist
- Task 5.4: Troubleshooting guide
- Task 5.5: Update README

**Total: 13-17 hours across 4 days**

## Dependencies Between Tasks

```
1.1 (Dockerfile) ──┐
1.2 (Compose)     ├──> T.1 (Test)
1.3 (Env)        ──┘

2.1 (Logging) ──┐
2.2 (Health)    ├──> 2.4 (Request Logging)
2.3 (Metrics) ──┘

3.1 (Rate Limit) ──┐
3.2 (CORS)         ├──> T.1 (Test)
3.3 (Size Limit) ──┘

4.1 (Shutdown) ──┐
4.2 (Pool)       ├──> T.1 (Test)
4.3 (Compress) ──┘

All Tasks ──> 5.x (Documentation)
```

## Success Criteria

- [ ] All Docker images build successfully
- [ ] All services start and pass health checks
- [ ] All tests pass
- [ ] Documentation is complete
- [ ] Security checklist verified
- [ ] Performance targets met
- [ ] Can deploy to Oracle Cloud
- [ ] Can deploy to Railway
- [ ] Monitoring is functional
- [ ] Rollback procedure tested
