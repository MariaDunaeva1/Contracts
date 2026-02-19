# Production Deployment Guide

Complete guide for deploying Finetune Studio to production with monitoring, security, and reliability.

## 📋 Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Prerequisites](#prerequisites)
- [Quick Start](#quick-start)
- [Deployment Options](#deployment-options)
- [Configuration](#configuration)
- [Monitoring](#monitoring)
- [Maintenance](#maintenance)
- [Troubleshooting](#troubleshooting)

## Overview

This guide covers production deployment with:

- ✅ Multi-stage Docker builds (800MB → 150MB)
- ✅ Structured logging with JSON format
- ✅ Prometheus metrics
- ✅ Rate limiting & security hardening
- ✅ Graceful shutdown
- ✅ Health checks & monitoring
- ✅ Automated backups
- ✅ SSL/TLS support

## Architecture

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

## Prerequisites

### System Requirements

**Minimum**:
- 2 CPU cores
- 4GB RAM
- 50GB storage
- Ubuntu 20.04+ or similar Linux

**Recommended**:
- 4 CPU cores
- 8GB RAM
- 100GB storage
- Ubuntu 22.04 LTS

### Software Requirements

- Docker 20.10+
- Docker Compose 2.0+
- Git
- curl

### Accounts & Credentials

- Kaggle API credentials ([get here](https://www.kaggle.com/settings))
- Domain name (optional)
- SSL certificate (optional, can use Let's Encrypt)

## Quick Start

### 1. Clone Repository

```bash
git clone https://github.com/yourusername/finetune-studio.git
cd finetune-studio
```

### 2. Configure Environment

```bash
# Copy environment template
cp .env.example .env

# Edit configuration
nano .env
```

**Critical settings to change**:

```bash
# Change these passwords!
DB_PASSWORD=your_secure_password_here
MINIO_PASSWORD=your_secure_minio_password

# Add your Kaggle credentials
KAGGLE_USERNAME=your_kaggle_username
KAGGLE_KEY=your_kaggle_api_key

# Set your domain
ALLOWED_ORIGINS=https://yourdomain.com

# Production settings
APP_ENV=production
LOG_FORMAT=json
LOG_LEVEL=info
```

### 3. Deploy

```bash
# Build and start all services
docker compose -f docker-compose.prod.yml up -d --build

# Check status
docker compose -f docker-compose.prod.yml ps

# View logs
docker compose -f docker-compose.prod.yml logs -f
```

### 4. Verify

```bash
# Health check
curl http://localhost:8080/api/v1/health

# Access frontend
open http://localhost
```

## Deployment Options

### Option 1: Oracle Cloud Free Tier (Recommended for Free)

**Pros**:
- ✅ Free forever
- ✅ 2-4 VMs with 12-24GB RAM
- ✅ 200GB storage
- ✅ Full control

**Cons**:
- ❌ Manual setup required
- ❌ You manage updates

**Guide**: [docs/DEPLOY_ORACLE_CLOUD.md](docs/DEPLOY_ORACLE_CLOUD.md)

**Time**: 30-60 minutes

### Option 2: Railway.app (Recommended for Ease)

**Pros**:
- ✅ 10-minute setup
- ✅ Automatic SSL
- ✅ GitHub integration
- ✅ Zero maintenance

**Cons**:
- ❌ $10-20/month cost
- ❌ Less control

**Guide**: [docs/DEPLOY_RAILWAY.md](docs/DEPLOY_RAILWAY.md)

**Time**: 10-15 minutes

### Option 3: Your Own Server

**Requirements**:
- VPS or dedicated server
- Root access
- Public IP address

**Steps**:
1. Install Docker & Docker Compose
2. Clone repository
3. Configure `.env`
4. Run `docker compose -f docker-compose.prod.yml up -d`
5. Setup SSL with Let's Encrypt

## Configuration

### Environment Variables

See [.env.example](.env.example) for all available options.

**Key configurations**:

#### Application
```bash
APP_ENV=production          # Environment
APP_VERSION=1.0.0          # Version
PORT=8080                  # API port
```

#### Database
```bash
DATABASE_URL=postgres://user:pass@host:5432/db
DB_MAX_CONNECTIONS=25
DB_MAX_IDLE_CONNECTIONS=5
DB_CONNECTION_MAX_LIFETIME=5m
```

#### Storage
```bash
MINIO_ENDPOINT=minio:9000
MINIO_USER=minioadmin
MINIO_PASSWORD=changeme
MINIO_USE_SSL=false        # Set true in production
MINIO_BUCKET=finetune-models
```

#### Security
```bash
ALLOWED_ORIGINS=https://yourdomain.com
RATE_LIMIT_REQUESTS_PER_MINUTE=100
RATE_LIMIT_EXPENSIVE_ENDPOINTS=10
MAX_REQUEST_SIZE_MB=10
```

#### Workers
```bash
WORKER_POOL_SIZE=5         # Adjust based on RAM
WORKER_TIMEOUT=24h
```

#### Logging
```bash
LOG_LEVEL=info             # debug, info, warn, error
LOG_FORMAT=json            # json or console
METRICS_ENABLED=true
```

### Resource Limits

Edit `docker-compose.prod.yml` to set limits:

```yaml
backend:
  deploy:
    resources:
      limits:
        cpus: '2.0'
        memory: 2G
      reservations:
        cpus: '1.0'
        memory: 512M
```

## Monitoring

### Health Checks

```bash
# Basic health check
curl http://localhost:8080/api/v1/health

# Expected response:
{
  "status": "healthy",
  "version": "1.0.0",
  "uptime": "2h30m",
  "services": {
    "database": {"status": "up", "response_time": "2ms"},
    "storage": {"status": "up", "response_time": "5ms"},
    "workers": {"status": "up"}
  }
}
```

### Metrics

```bash
# Prometheus metrics
curl http://localhost:8080/api/v1/metrics

# Key metrics:
# - http_requests_total
# - http_request_duration_seconds
# - worker_pool_active_jobs
# - worker_pool_queued_jobs
# - db_connections_active
# - db_connections_idle
```

### Logs

```bash
# View all logs
docker compose -f docker-compose.prod.yml logs -f

# View specific service
docker compose -f docker-compose.prod.yml logs -f backend

# Last 100 lines
docker compose -f docker-compose.prod.yml logs --tail=100 backend

# Follow with grep
docker compose -f docker-compose.prod.yml logs -f backend | grep ERROR
```

### Resource Usage

```bash
# Docker stats
docker stats

# System resources
htop

# Disk usage
df -h
du -sh /var/lib/docker
```

## Maintenance

### Backups

#### Automated Daily Backups

```bash
# Make backup script executable
chmod +x scripts/backup.sh

# Test backup
./scripts/backup.sh

# Setup cron job
crontab -e

# Add daily backup at 2 AM
0 2 * * * /path/to/finetune-studio/scripts/backup.sh >> /var/log/backup.log 2>&1
```

#### Manual Backup

```bash
# Backup database
docker exec finetune-postgres pg_dump -U finetune finetune_db > backup.sql

# Backup MinIO data
docker run --rm -v minio_data:/data -v $(pwd):/backup alpine tar czf /backup/minio_backup.tar.gz /data
```

#### Restore from Backup

```bash
# Make restore script executable
chmod +x scripts/restore.sh

# Restore database
./scripts/restore.sh backups/db_backup_20260219_120000.sql.gz
```

### Updates

```bash
# Pull latest changes
git pull

# Rebuild and restart
docker compose -f docker-compose.prod.yml up -d --build

# Clean old images
docker image prune -f
```

### SSL Certificate Renewal

```bash
# Renew Let's Encrypt certificate
sudo certbot renew

# Restart frontend
docker compose -f docker-compose.prod.yml restart frontend
```

### Log Rotation

Add to `docker-compose.prod.yml`:

```yaml
logging:
  driver: "json-file"
  options:
    max-size: "10m"
    max-file: "3"
```

## Troubleshooting

### Quick Diagnostics

```bash
# Run production test suite
./scripts/test_production.sh

# Check service status
docker compose -f docker-compose.prod.yml ps

# View recent errors
docker compose -f docker-compose.prod.yml logs --tail=50 | grep ERROR

# Check resource usage
docker stats --no-stream
```

### Common Issues

See [docs/TROUBLESHOOTING_PRODUCTION.md](docs/TROUBLESHOOTING_PRODUCTION.md) for detailed solutions.

**Quick fixes**:

```bash
# Service won't start
docker compose -f docker-compose.prod.yml restart

# Database connection issues
docker compose -f docker-compose.prod.yml restart postgres backend

# Out of memory
# Edit .env: WORKER_POOL_SIZE=2
docker compose -f docker-compose.prod.yml restart backend

# Disk space issues
docker system prune -a
```

## Security Checklist

Before going live:

- [ ] Change all default passwords
- [ ] Set `ALLOWED_ORIGINS` to your domain
- [ ] Enable SSL/TLS (`MINIO_USE_SSL=true`)
- [ ] Configure firewall (ports 80, 443, 22 only)
- [ ] Setup automated backups
- [ ] Enable automatic security updates
- [ ] Review logs for suspicious activity
- [ ] Test disaster recovery procedure

## Performance Tuning

### For 4GB RAM

```bash
WORKER_POOL_SIZE=2
DB_MAX_CONNECTIONS=15
```

### For 8GB RAM

```bash
WORKER_POOL_SIZE=5
DB_MAX_CONNECTIONS=25
```

### For 16GB+ RAM

```bash
WORKER_POOL_SIZE=10
DB_MAX_CONNECTIONS=50
```

## Documentation

- **Deployment Guides**:
  - [Oracle Cloud](docs/DEPLOY_ORACLE_CLOUD.md)
  - [Railway.app](docs/DEPLOY_RAILWAY.md)
  
- **Operations**:
  - [Production Checklist](docs/PRODUCTION_CHECKLIST.md)
  - [Troubleshooting](docs/TROUBLESHOOTING_PRODUCTION.md)
  
- **Development**:
  - [API Examples](API_EXAMPLES.md)
  - [Quick Start](QUICK_START.md)

## Support

- **Documentation**: See `/docs` folder
- **Issues**: GitHub Issues
- **Health Check**: `http://your-server/api/v1/health`
- **Metrics**: `http://your-server/api/v1/metrics`

## Next Steps

1. ✅ Deploy to production
2. ✅ Setup monitoring
3. ✅ Configure backups
4. ✅ Setup SSL
5. ✅ Test disaster recovery
6. ✅ Train your first model!

---

**Production-ready ML training platform** 🚀

Built with Go, PostgreSQL, MinIO, and Docker.
