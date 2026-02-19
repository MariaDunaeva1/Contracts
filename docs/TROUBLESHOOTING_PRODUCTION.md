# Production Troubleshooting Guide

Common issues and solutions for production deployments.

## Table of Contents

1. [Service Won't Start](#service-wont-start)
2. [Database Connection Issues](#database-connection-issues)
3. [Storage Connection Issues](#storage-connection-issues)
4. [High Memory Usage](#high-memory-usage)
5. [Slow Response Times](#slow-response-times)
6. [Worker Pool Issues](#worker-pool-issues)
7. [SSL/HTTPS Issues](#sslhttps-issues)
8. [CORS Errors](#cors-errors)
9. [Rate Limiting Issues](#rate-limiting-issues)
10. [Disk Space Issues](#disk-space-issues)

---

## Service Won't Start

### Symptoms
- Container exits immediately
- Health check fails
- Service shows as "unhealthy"

### Diagnosis

```bash
# Check container status
docker compose -f docker-compose.prod.yml ps

# View logs
docker compose -f docker-compose.prod.yml logs backend

# Check specific container
docker logs finetune-backend
```

### Common Causes & Solutions

#### 1. Missing Environment Variables

**Error**: `required env var DATABASE_URL not set`

**Solution**:
```bash
# Check .env file exists
ls -la .env

# Verify all required variables are set
grep -E "DATABASE_URL|MINIO_|KAGGLE_" .env

# Restart services
docker compose -f docker-compose.prod.yml restart
```

#### 2. Port Already in Use

**Error**: `bind: address already in use`

**Solution**:
```bash
# Find process using port 8080
sudo lsof -i :8080

# Kill the process
sudo kill -9 <PID>

# Or change port in .env
echo "PORT=8081" >> .env
```

#### 3. Out of Memory During Build

**Error**: `signal: killed` during build

**Solution**:
```bash
# Build with less parallelism
docker compose -f docker-compose.prod.yml build --no-cache --parallel 1

# Or increase swap space
sudo fallocate -l 4G /swapfile
sudo chmod 600 /swapfile
sudo mkswap /swapfile
sudo swapon /swapfile
```

---

## Database Connection Issues

### Symptoms
- `connection refused`
- `database "finetune_db" does not exist`
- `password authentication failed`

### Diagnosis

```bash
# Check database container
docker compose -f docker-compose.prod.yml ps postgres

# Check database logs
docker compose -f docker-compose.prod.yml logs postgres

# Test connection manually
docker exec -it finetune-postgres psql -U finetune -d finetune_db
```

### Solutions

#### 1. Database Not Ready

**Error**: `connection refused`

**Solution**:
```bash
# Wait for database to be ready
docker compose -f docker-compose.prod.yml up -d postgres
sleep 10

# Check health
docker compose -f docker-compose.prod.yml ps postgres

# Restart backend
docker compose -f docker-compose.prod.yml restart backend
```

#### 2. Wrong Credentials

**Error**: `password authentication failed`

**Solution**:
```bash
# Verify credentials in .env
grep DB_ .env

# Reset database with new credentials
docker compose -f docker-compose.prod.yml down -v
docker compose -f docker-compose.prod.yml up -d
```

#### 3. Database Doesn't Exist

**Error**: `database "finetune_db" does not exist`

**Solution**:
```bash
# Create database manually
docker exec -it finetune-postgres createdb -U finetune finetune_db

# Or recreate with volumes
docker compose -f docker-compose.prod.yml down -v
docker compose -f docker-compose.prod.yml up -d
```

---

## Storage Connection Issues

### Symptoms
- `connection refused` to MinIO
- `Access Denied` errors
- Files not uploading

### Diagnosis

```bash
# Check MinIO container
docker compose -f docker-compose.prod.yml ps minio

# Check MinIO logs
docker compose -f docker-compose.prod.yml logs minio

# Test MinIO access
curl http://localhost:9000/minio/health/live
```

### Solutions

#### 1. MinIO Not Ready

**Solution**:
```bash
# Restart MinIO
docker compose -f docker-compose.prod.yml restart minio

# Wait for health check
sleep 30

# Restart backend
docker compose -f docker-compose.prod.yml restart backend
```

#### 2. Wrong Credentials

**Solution**:
```bash
# Verify credentials
grep MINIO_ .env

# Reset MinIO
docker compose -f docker-compose.prod.yml down
docker volume rm contracts_minio_data
docker compose -f docker-compose.prod.yml up -d
```

#### 3. Bucket Doesn't Exist

**Solution**:
```bash
# Create bucket manually
docker exec -it finetune-minio mc alias set local http://localhost:9000 minioadmin minioadmin
docker exec -it finetune-minio mc mb local/finetune-models
```

---

## High Memory Usage

### Symptoms
- OOM (Out of Memory) errors
- Container killed
- System becomes unresponsive

### Diagnosis

```bash
# Check memory usage
docker stats

# Check system memory
free -h

# Check container limits
docker inspect finetune-backend | grep -i memory
```

### Solutions

#### 1. Reduce Worker Pool Size

```bash
# Edit .env
nano .env

# Set lower worker count
WORKER_POOL_SIZE=2

# Restart
docker compose -f docker-compose.prod.yml restart backend
```

#### 2. Set Memory Limits

Add to `docker-compose.prod.yml`:

```yaml
backend:
  deploy:
    resources:
      limits:
        memory: 2G
      reservations:
        memory: 512M
```

#### 3. Increase Swap Space

```bash
# Create 4GB swap
sudo fallocate -l 4G /swapfile
sudo chmod 600 /swapfile
sudo mkswap /swapfile
sudo swapon /swapfile

# Make permanent
echo '/swapfile none swap sw 0 0' | sudo tee -a /etc/fstab
```

---

## Slow Response Times

### Symptoms
- API requests take > 1 second
- Frontend feels sluggish
- Timeouts

### Diagnosis

```bash
# Test response time
time curl http://localhost:8080/api/v1/health

# Check metrics
curl http://localhost:8080/api/v1/metrics | grep http_request_duration

# Check database connections
docker exec -it finetune-postgres psql -U finetune -d finetune_db -c "SELECT count(*) FROM pg_stat_activity;"
```

### Solutions

#### 1. Database Connection Pool Exhausted

```bash
# Increase pool size in .env
DB_MAX_CONNECTIONS=50
DB_MAX_IDLE_CONNECTIONS=10

# Restart
docker compose -f docker-compose.prod.yml restart backend
```

#### 2. High CPU Usage

```bash
# Check CPU usage
docker stats

# Reduce worker pool
WORKER_POOL_SIZE=2

# Add CPU limits
# In docker-compose.prod.yml:
cpus: '2.0'
```

#### 3. Network Latency

```bash
# Check network
ping -c 5 localhost

# Check DNS
nslookup postgres

# Restart networking
docker compose -f docker-compose.prod.yml restart
```

---

## Worker Pool Issues

### Symptoms
- Jobs stuck in "pending"
- Jobs never start
- Workers not processing

### Diagnosis

```bash
# Check backend logs
docker compose -f docker-compose.prod.yml logs backend | grep -i worker

# Check job status
curl http://localhost:8080/api/v1/jobs

# Check metrics
curl http://localhost:8080/api/v1/metrics | grep worker_pool
```

### Solutions

#### 1. Workers Not Started

```bash
# Restart backend
docker compose -f docker-compose.prod.yml restart backend

# Check logs for worker initialization
docker compose -f docker-compose.prod.yml logs backend | grep "Worker pool"
```

#### 2. All Workers Busy

```bash
# Increase worker pool size
echo "WORKER_POOL_SIZE=5" >> .env

# Restart
docker compose -f docker-compose.prod.yml restart backend
```

#### 3. Kaggle API Issues

```bash
# Verify Kaggle credentials
docker exec -it finetune-backend kaggle datasets list

# If fails, update credentials in .env
KAGGLE_USERNAME=your_username
KAGGLE_KEY=your_key
```

---

## SSL/HTTPS Issues

### Symptoms
- Certificate errors
- "Not secure" warning
- HTTPS not working

### Diagnosis

```bash
# Check certificate
openssl s_client -connect yourdomain.com:443 -servername yourdomain.com

# Check certificate expiry
echo | openssl s_client -connect yourdomain.com:443 2>/dev/null | openssl x509 -noout -dates

# Check Nginx config
docker exec -it finetune-frontend nginx -t
```

### Solutions

#### 1. Certificate Expired

```bash
# Renew with Certbot
sudo certbot renew

# Restart frontend
docker compose -f docker-compose.prod.yml restart frontend
```

#### 2. Certificate Not Found

```bash
# Get new certificate
sudo certbot certonly --standalone -d yourdomain.com

# Mount in docker-compose.prod.yml:
volumes:
  - /etc/letsencrypt:/etc/letsencrypt:ro
```

#### 3. Mixed Content Warnings

Update frontend to use HTTPS for API calls:
```javascript
const API_BASE_URL = 'https://api.yourdomain.com/api/v1';
```

---

## CORS Errors

### Symptoms
- "CORS policy" errors in browser console
- API calls fail from frontend
- OPTIONS requests fail

### Diagnosis

```bash
# Test CORS headers
curl -I -X OPTIONS http://localhost:8080/api/v1/health

# Check allowed origins
docker compose -f docker-compose.prod.yml exec backend env | grep ALLOWED_ORIGINS
```

### Solutions

#### 1. Wrong Origin

```bash
# Update .env with correct origin
ALLOWED_ORIGINS=https://yourdomain.com,https://www.yourdomain.com

# Restart
docker compose -f docker-compose.prod.yml restart backend
```

#### 2. Missing Credentials

Frontend needs:
```javascript
fetch(url, {
  credentials: 'include'  // Add this
})
```

#### 3. Preflight Fails

Check Nginx config allows OPTIONS:
```nginx
if ($request_method = 'OPTIONS') {
    add_header 'Access-Control-Allow-Origin' '*';
    add_header 'Access-Control-Allow-Methods' 'GET, POST, OPTIONS';
    return 204;
}
```

---

## Rate Limiting Issues

### Symptoms
- 429 Too Many Requests
- Legitimate requests blocked
- Rate limit too strict/loose

### Diagnosis

```bash
# Check rate limit settings
grep RATE_LIMIT .env

# Test rate limiting
for i in {1..150}; do curl -s -o /dev/null -w "%{http_code}\n" http://localhost:8080/api/v1/health; done
```

### Solutions

#### 1. Adjust Rate Limits

```bash
# Edit .env
RATE_LIMIT_REQUESTS_PER_MINUTE=200
RATE_LIMIT_EXPENSIVE_ENDPOINTS=20

# Restart
docker compose -f docker-compose.prod.yml restart backend
```

#### 2. Whitelist IPs

Add to backend code:
```go
if c.ClientIP() == "trusted.ip.address" {
    c.Next()
    return
}
```

#### 3. Use Redis for Distributed Rate Limiting

For multiple backend instances, use Redis-based rate limiting.

---

## Disk Space Issues

### Symptoms
- "No space left on device"
- Containers won't start
- Logs stop writing

### Diagnosis

```bash
# Check disk space
df -h

# Check Docker disk usage
docker system df

# Check largest directories
du -sh /* | sort -h
```

### Solutions

#### 1. Clean Docker Resources

```bash
# Remove unused images
docker image prune -a

# Remove unused volumes
docker volume prune

# Remove unused containers
docker container prune

# Clean everything
docker system prune -a --volumes
```

#### 2. Clean Logs

```bash
# Truncate Docker logs
sudo sh -c "truncate -s 0 /var/lib/docker/containers/*/*-json.log"

# Configure log rotation in docker-compose.prod.yml:
logging:
  driver: "json-file"
  options:
    max-size: "10m"
    max-file: "3"
```

#### 3. Clean Old Backups

```bash
# Remove backups older than 7 days
find ./backups -name "*.sql" -mtime +7 -delete
```

---

## Getting Help

If you can't resolve the issue:

1. **Collect Information**:
   ```bash
   # System info
   uname -a
   docker --version
   docker compose version
   
   # Service status
   docker compose -f docker-compose.prod.yml ps
   
   # Recent logs
   docker compose -f docker-compose.prod.yml logs --tail=100
   
   # Resource usage
   docker stats --no-stream
   ```

2. **Check Documentation**:
   - README.md
   - DEPLOYMENT.md
   - API_EXAMPLES.md

3. **Search Issues**:
   - GitHub Issues
   - Stack Overflow

4. **Create Issue**:
   - Include system info
   - Include error logs
   - Include steps to reproduce

---

## Prevention

### Regular Maintenance

```bash
# Weekly
- Review logs for errors
- Check disk space
- Update dependencies

# Monthly
- Rotate secrets
- Test backups
- Review metrics

# Quarterly
- Security audit
- Performance review
- Update documentation
```

### Monitoring

Setup alerts for:
- Service down
- High error rate
- High memory usage
- Disk space low
- Certificate expiring

### Backups

Automate daily backups:
```bash
# Add to crontab
0 2 * * * /path/to/backup.sh
```

---

**Last Updated**: 2026-02-19
