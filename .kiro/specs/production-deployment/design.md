# Production Deployment - Design Document

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                     Production Stack                         │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  ┌──────────────┐      ┌──────────────┐                     │
│  │   Frontend   │─────▶│    Nginx     │                     │
│  │  (Static)    │      │  (Reverse    │                     │
│  │              │      │   Proxy)     │                     │
│  └──────────────┘      └──────┬───────┘                     │
│                               │                              │
│                               ▼                              │
│                        ┌──────────────┐                     │
│                        │   Backend    │                     │
│                        │   (Go API)   │                     │
│                        │              │                     │
│                        │ - Metrics    │                     │
│                        │ - Health     │                     │
│                        │ - Logging    │                     │
│                        │ - Workers    │                     │
│                        └──────┬───────┘                     │
│                               │                              │
│                    ┌──────────┴──────────┐                 │
│                    ▼                     ▼                  │
│             ┌──────────────┐      ┌──────────────┐         │
│             │  PostgreSQL  │      │    MinIO     │         │
│             │  (Persistent)│      │  (Persistent)│         │
│             └──────────────┘      └──────────────┘         │
│                                                               │
└─────────────────────────────────────────────────────────────┘
```

## Component Designs

### 1. Multi-Stage Docker Builds

#### Backend Dockerfile
```dockerfile
# Stage 1: Builder
FROM golang:1.23-alpine AS builder
WORKDIR /build
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o server cmd/server/main.go

# Stage 2: Runtime
FROM alpine:3.19
RUN apk add --no-cache ca-certificates python3 py3-pip && \
    pip install kaggle --break-system-packages && \
    addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser
WORKDIR /app
COPY --from=builder /build/server .
COPY --from=builder /build/templates ./templates
USER appuser
EXPOSE 8080
CMD ["./server"]
```

**Benefits:**
- Image size reduction: ~800MB → ~150MB
- No build tools in production image
- Non-root user for security
- Optimized binary with stripped symbols

#### Frontend Dockerfile
```dockerfile
# Stage 1: Build (if needed for asset processing)
FROM nginx:alpine
COPY nginx.conf /etc/nginx/nginx.conf
COPY . /usr/share/nginx/html
RUN chown -R nginx:nginx /usr/share/nginx/html && \
    chmod -R 755 /usr/share/nginx/html
USER nginx
EXPOSE 80
```

### 2. Monitoring & Observability

#### Enhanced Health Check Endpoint
```go
type HealthResponse struct {
    Status   string            `json:"status"`
    Version  string            `json:"version"`
    Uptime   string            `json:"uptime"`
    Services map[string]Health `json:"services"`
}

type Health struct {
    Status      string  `json:"status"`
    ResponseTime string  `json:"response_time"`
    Details     string  `json:"details,omitempty"`
}

// GET /api/v1/health
// Returns:
// - 200 if all services healthy
// - 503 if any critical service down
```

#### Metrics Endpoint
```go
// GET /api/v1/metrics
// Prometheus format:
// - http_requests_total{method, path, status}
// - http_request_duration_seconds{method, path}
// - worker_pool_active_jobs
// - worker_pool_queued_jobs
// - db_connections_active
// - db_connections_idle
```

#### Structured Logging
```go
// Use uber-go/zap for structured logging
logger.Info("request completed",
    zap.String("method", method),
    zap.String("path", path),
    zap.Int("status", status),
    zap.Duration("duration", duration),
    zap.String("user_agent", ua),
)
```

### 3. Security Hardening

#### Rate Limiting
```go
// Use gin-contrib/rate
import "github.com/gin-contrib/rate"

// Global rate limit: 100 req/min per IP
limiter := rate.NewLimiter(rate.Limit(100), 100)

// Expensive endpoints: 10 req/min
expensiveLimiter := rate.NewLimiter(rate.Limit(10), 10)
```

#### CORS Configuration
```go
// Production CORS
config := cors.Config{
    AllowOrigins:     []string{os.Getenv("ALLOWED_ORIGINS")},
    AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
    AllowHeaders:     []string{"Content-Type", "Authorization"},
    ExposeHeaders:    []string{"Content-Length"},
    AllowCredentials: true,
    MaxAge:           12 * time.Hour,
}
```

#### Request Size Limits
```go
// Limit request body size
r.Use(func(c *gin.Context) {
    c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 10<<20) // 10MB
    c.Next()
})
```

### 4. Configuration Management

#### Environment Variables Structure
```bash
# Application
APP_ENV=production
APP_VERSION=1.0.0
PORT=8080

# Database
DATABASE_URL=postgres://user:pass@host:5432/db?sslmode=require
DB_MAX_CONNECTIONS=25
DB_MAX_IDLE_CONNECTIONS=5
DB_CONNECTION_MAX_LIFETIME=5m

# MinIO/S3
MINIO_ENDPOINT=minio:9000
MINIO_USER=minioadmin
MINIO_PASSWORD=changeme
MINIO_USE_SSL=true
MINIO_BUCKET=finetune-models

# Kaggle
KAGGLE_USERNAME=your_username
KAGGLE_KEY=your_api_key

# Security
ALLOWED_ORIGINS=https://yourdomain.com
RATE_LIMIT_REQUESTS_PER_MINUTE=100
MAX_REQUEST_SIZE_MB=10

# Workers
WORKER_POOL_SIZE=5
WORKER_TIMEOUT=24h

# Monitoring
LOG_LEVEL=info
LOG_FORMAT=json
METRICS_ENABLED=true
```

#### Config Validation
```go
func LoadConfig() (*Config, error) {
    cfg := &Config{}
    
    // Required fields
    required := []string{
        "DATABASE_URL",
        "MINIO_ENDPOINT",
        "MINIO_USER",
        "MINIO_PASSWORD",
        "KAGGLE_USERNAME",
        "KAGGLE_KEY",
    }
    
    for _, key := range required {
        if os.Getenv(key) == "" {
            return nil, fmt.Errorf("required env var %s not set", key)
        }
    }
    
    // Load with defaults
    cfg.Port = getEnvOrDefault("PORT", "8080")
    cfg.LogLevel = getEnvOrDefault("LOG_LEVEL", "info")
    // ...
    
    return cfg, nil
}
```

### 5. Production Docker Compose

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:15-alpine
    restart: unless-stopped
    environment:
      POSTGRES_USER: ${DB_USER}
      POSTGRES_PASSWORD: ${DB_PASSWORD}
      POSTGRES_DB: ${DB_NAME}
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./backups:/backups
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${DB_USER}"]
      interval: 10s
      timeout: 5s
      retries: 5
    networks:
      - backend

  minio:
    image: minio/minio:latest
    restart: unless-stopped
    command: server /data --console-address ":9001"
    environment:
      MINIO_ROOT_USER: ${MINIO_USER}
      MINIO_ROOT_PASSWORD: ${MINIO_PASSWORD}
    volumes:
      - minio_data:/data
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:9000/minio/health/live"]
      interval: 30s
      timeout: 20s
      retries: 3
    networks:
      - backend

  backend:
    build:
      context: ./backend
      dockerfile: Dockerfile
    restart: unless-stopped
    environment:
      - APP_ENV=production
      - DATABASE_URL=postgres://${DB_USER}:${DB_PASSWORD}@postgres:5432/${DB_NAME}?sslmode=disable
      - MINIO_ENDPOINT=minio:9000
      - MINIO_USER=${MINIO_USER}
      - MINIO_PASSWORD=${MINIO_PASSWORD}
      - MINIO_USE_SSL=false
      - KAGGLE_USERNAME=${KAGGLE_USERNAME}
      - KAGGLE_KEY=${KAGGLE_KEY}
      - ALLOWED_ORIGINS=${ALLOWED_ORIGINS}
      - LOG_LEVEL=info
      - LOG_FORMAT=json
    depends_on:
      postgres:
        condition: service_healthy
      minio:
        condition: service_healthy
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8080/api/v1/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
    networks:
      - backend
      - frontend

  frontend:
    build:
      context: ./frontend
      dockerfile: Dockerfile
    restart: unless-stopped
    ports:
      - "${FRONTEND_PORT:-80}:80"
    depends_on:
      - backend
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:80/"]
      interval: 30s
      timeout: 10s
      retries: 3
    networks:
      - frontend

networks:
  backend:
    driver: bridge
  frontend:
    driver: bridge

volumes:
  postgres_data:
    driver: local
  minio_data:
    driver: local
```

### 6. Graceful Shutdown

```go
func main() {
    // ... setup ...
    
    srv := &http.Server{
        Addr:    ":8080",
        Handler: r,
    }
    
    // Start server in goroutine
    go func() {
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("Server error: %v", err)
        }
    }()
    
    // Wait for interrupt signal
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    
    log.Println("Shutting down server...")
    
    // Stop accepting new jobs
    worker.Pool.Stop()
    
    // Graceful shutdown with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    if err := srv.Shutdown(ctx); err != nil {
        log.Fatal("Server forced to shutdown:", err)
    }
    
    log.Println("Server exited")
}
```

## Deployment Strategies

### Oracle Cloud Free Tier
- 2x AMD VM (1 OCPU, 1GB RAM each) OR 4x ARM VM (1 OCPU, 6GB RAM each)
- 200GB block storage
- 10TB outbound transfer/month

**Setup:**
1. Create VM instance
2. Install Docker & Docker Compose
3. Configure firewall (ports 80, 443, 8080)
4. Clone repository
5. Configure `.env` file
6. Run `docker-compose -f docker-compose.prod.yml up -d`
7. Setup SSL with Let's Encrypt

### Railway.app
- $5/month free credit
- Automatic SSL
- GitHub integration

**Setup:**
1. Connect GitHub repository
2. Configure environment variables in dashboard
3. Deploy from main branch
4. Railway handles SSL and domain

## Performance Optimizations

### Database Connection Pooling
```go
sqlDB, _ := db.DB()
sqlDB.SetMaxOpenConns(25)
sqlDB.SetMaxIdleConns(5)
sqlDB.SetConnMaxLifetime(5 * time.Minute)
```

### Response Compression
```go
import "github.com/gin-contrib/gzip"
r.Use(gzip.Gzip(gzip.DefaultCompression))
```

### Static Asset Caching
```nginx
location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg)$ {
    expires 1y;
    add_header Cache-Control "public, immutable";
}
```

## Monitoring Dashboard

### Key Metrics to Track
1. Request rate (req/s)
2. Error rate (%)
3. Response time (p50, p95, p99)
4. Active workers
5. Queue depth
6. Database connections
7. Storage usage

### Alerting Thresholds
- Error rate > 5%
- Response time p95 > 1s
- Queue depth > 100
- Database connections > 80% of max
- Storage usage > 80%

## Rollback Procedure

1. Stop services: `docker-compose -f docker-compose.prod.yml down`
2. Restore database backup: `psql < backups/backup_YYYYMMDD.sql`
3. Checkout previous version: `git checkout <previous-tag>`
4. Rebuild: `docker-compose -f docker-compose.prod.yml build`
5. Start: `docker-compose -f docker-compose.prod.yml up -d`
6. Verify health: `curl http://localhost:8080/api/v1/health`

## Security Checklist

- [ ] All secrets in environment variables
- [ ] Database uses SSL in production
- [ ] MinIO uses SSL in production
- [ ] CORS restricted to production domain
- [ ] Rate limiting enabled
- [ ] Request size limits configured
- [ ] Non-root Docker users
- [ ] Firewall configured (only 80, 443 open)
- [ ] Regular security updates scheduled
- [ ] Backup strategy implemented

## Testing Strategy

### Pre-Deployment Tests
1. Build all images successfully
2. All services start and pass health checks
3. Database migrations run successfully
4. API endpoints respond correctly
5. Frontend loads and connects to backend
6. Worker pool processes jobs
7. Logs are properly formatted
8. Metrics endpoint returns data

### Post-Deployment Validation
1. Health check returns 200
2. Can create dataset
3. Can create job
4. Can view logs
5. Can download model
6. Frontend accessible
7. SSL certificate valid
8. Monitoring data flowing

## Implementation Order

1. **Phase 1: Core Infrastructure** (2-3 hours)
   - Multi-stage Dockerfiles
   - Production docker-compose.yml
   - Environment configuration

2. **Phase 2: Monitoring** (2-3 hours)
   - Enhanced health checks
   - Metrics endpoint
   - Structured logging

3. **Phase 3: Security** (2 hours)
   - Rate limiting
   - CORS hardening
   - Request validation

4. **Phase 4: Reliability** (1-2 hours)
   - Graceful shutdown
   - Connection pooling
   - Health checks

5. **Phase 5: Documentation** (2 hours)
   - Deployment guides
   - Environment templates
   - Troubleshooting guide

**Total Estimated Time: 9-12 hours**
