# Production Deployment Checklist

Use this checklist to ensure a successful production deployment of Finetune Studio.

## Pre-Deployment

### Configuration

- [ ] Copy `.env.example` to `.env`
- [ ] Set `APP_ENV=production`
- [ ] Set `LOG_FORMAT=json`
- [ ] Set `LOG_LEVEL=info` (or `warn` for less verbose)
- [ ] Configure `PORT` (default: 8080)

### Database

- [ ] Set strong `DB_PASSWORD` (min 16 characters, mixed case, numbers, symbols)
- [ ] Configure `DATABASE_URL` with correct credentials
- [ ] Set `DB_MAX_CONNECTIONS` based on expected load (default: 25)
- [ ] Set `DB_MAX_IDLE_CONNECTIONS` (default: 5)
- [ ] Set `DB_CONNECTION_MAX_LIFETIME` (default: 5m)
- [ ] Enable SSL for database in production (`sslmode=require`)

### Storage (MinIO/S3)

- [ ] Set strong `MINIO_PASSWORD` (min 16 characters)
- [ ] Configure `MINIO_ENDPOINT` (without http://)
- [ ] Set `MINIO_USE_SSL=true` for production
- [ ] Configure `MINIO_BUCKET` name
- [ ] Verify bucket exists or will be auto-created
- [ ] Test S3/MinIO connectivity

### Kaggle API

- [ ] Obtain Kaggle API credentials from https://www.kaggle.com/settings
- [ ] Set `KAGGLE_USERNAME`
- [ ] Set `KAGGLE_KEY`
- [ ] Verify credentials work: `kaggle datasets list`

### Security

- [ ] Configure `ALLOWED_ORIGINS` with production domain(s)
- [ ] Remove `*` from CORS origins
- [ ] Set `RATE_LIMIT_REQUESTS_PER_MINUTE` (default: 100)
- [ ] Set `RATE_LIMIT_EXPENSIVE_ENDPOINTS` (default: 10)
- [ ] Set `MAX_REQUEST_SIZE_MB` (default: 10)
- [ ] Review all passwords are changed from defaults
- [ ] Ensure no secrets in git repository
- [ ] Add `.env` to `.gitignore`

### Workers

- [ ] Set `WORKER_POOL_SIZE` based on available resources
  - 2 workers for 4GB RAM
  - 3-5 workers for 8GB+ RAM
- [ ] Set `WORKER_TIMEOUT` (default: 24h)

### Monitoring

- [ ] Set `METRICS_ENABLED=true`
- [ ] Verify Prometheus metrics endpoint works
- [ ] Configure log aggregation (optional)

## Infrastructure Setup

### Server/VM

- [ ] Provision server with adequate resources:
  - Minimum: 2 CPU cores, 4GB RAM, 50GB storage
  - Recommended: 4 CPU cores, 8GB RAM, 100GB storage
- [ ] Install Docker (version 20.10+)
- [ ] Install Docker Compose (version 2.0+)
- [ ] Configure firewall rules:
  - Allow port 80 (HTTP)
  - Allow port 443 (HTTPS)
  - Allow port 22 (SSH)
  - Block all other ports
- [ ] Setup SSH key authentication
- [ ] Disable password authentication for SSH
- [ ] Configure automatic security updates

### Domain & SSL

- [ ] Register domain name (if using custom domain)
- [ ] Point DNS A record to server IP
- [ ] Wait for DNS propagation (can take 24-48 hours)
- [ ] Install Certbot for Let's Encrypt
- [ ] Obtain SSL certificate
- [ ] Configure automatic certificate renewal
- [ ] Test HTTPS access

### Networking

- [ ] Configure firewall (ufw, iptables, or cloud firewall)
- [ ] Setup reverse proxy (Nginx) if needed
- [ ] Configure rate limiting at network level
- [ ] Setup DDoS protection (Cloudflare, etc.)
- [ ] Test connectivity from external network

## Deployment

### Build & Start

- [ ] Clone repository to server
- [ ] Checkout production branch/tag
- [ ] Copy and configure `.env` file
- [ ] Build Docker images:
  ```bash
  docker compose -f docker-compose.prod.yml build
  ```
- [ ] Start services:
  ```bash
  docker compose -f docker-compose.prod.yml up -d
  ```
- [ ] Wait for all services to be healthy (2-3 minutes)

### Verify Services

- [ ] Check all containers are running:
  ```bash
  docker compose -f docker-compose.prod.yml ps
  ```
- [ ] Verify PostgreSQL is healthy
- [ ] Verify MinIO is healthy
- [ ] Verify backend is healthy
- [ ] Verify frontend is accessible
- [ ] Check logs for errors:
  ```bash
  docker compose -f docker-compose.prod.yml logs
  ```

## Post-Deployment Validation

### Health Checks

- [ ] Test health endpoint:
  ```bash
  curl http://localhost:8080/api/v1/health
  ```
- [ ] Verify all services show "up" status
- [ ] Check response time < 100ms
- [ ] Test from external network

### API Endpoints

- [ ] Test dataset upload:
  ```bash
  curl -X POST http://localhost:8080/api/v1/datasets \
    -F "file=@test.json" \
    -F "name=test-dataset"
  ```
- [ ] Test dataset list:
  ```bash
  curl http://localhost:8080/api/v1/datasets
  ```
- [ ] Test job creation
- [ ] Test job listing
- [ ] Test logs streaming (SSE)
- [ ] Test model listing
- [ ] Test model download

### Frontend

- [ ] Access frontend in browser
- [ ] Verify dashboard loads
- [ ] Test dataset upload page
- [ ] Test training wizard
- [ ] Test training view page
- [ ] Test evaluation view page
- [ ] Check browser console for errors
- [ ] Test on mobile device

### Performance

- [ ] Run load test (see `scripts/load_test.sh`)
- [ ] Verify response times under load
- [ ] Check rate limiting works
- [ ] Monitor resource usage (CPU, memory, disk)
- [ ] Verify no memory leaks over time

### Monitoring

- [ ] Access metrics endpoint:
  ```bash
  curl http://localhost:8080/api/v1/metrics
  ```
- [ ] Verify metrics are being collected
- [ ] Setup monitoring dashboard (Grafana, etc.)
- [ ] Configure alerts for:
  - Service down
  - High error rate (>5%)
  - High response time (p95 >1s)
  - High memory usage (>80%)
  - High disk usage (>80%)

### Logging

- [ ] Verify logs are in JSON format
- [ ] Check log level is appropriate
- [ ] Verify no sensitive data in logs
- [ ] Setup log aggregation (optional)
- [ ] Configure log rotation
- [ ] Test log search and filtering

## Backup & Recovery

### Database Backups

- [ ] Create backup script (see `scripts/backup.sh`)
- [ ] Test backup script manually
- [ ] Setup automated daily backups (cron)
- [ ] Verify backups are created successfully
- [ ] Test database restore from backup
- [ ] Configure backup retention (keep 7 days)
- [ ] Store backups off-server (S3, etc.)

### Storage Backups

- [ ] Configure MinIO/S3 versioning
- [ ] Setup periodic snapshots
- [ ] Test file restore
- [ ] Document backup locations

### Disaster Recovery

- [ ] Document recovery procedure
- [ ] Test full system restore
- [ ] Verify RTO (Recovery Time Objective) < 1 hour
- [ ] Verify RPO (Recovery Point Objective) < 24 hours
- [ ] Create runbook for common failures

## Security Hardening

### Application Security

- [ ] All secrets in environment variables (not code)
- [ ] CORS restricted to production domains
- [ ] Rate limiting enabled and tested
- [ ] Request size limits configured
- [ ] Input validation on all endpoints
- [ ] SQL injection prevention (using ORM)
- [ ] XSS prevention (proper escaping)
- [ ] CSRF protection (if using cookies)

### Infrastructure Security

- [ ] Firewall configured and enabled
- [ ] SSH key authentication only
- [ ] Fail2ban installed and configured
- [ ] Automatic security updates enabled
- [ ] Unnecessary services disabled
- [ ] Root login disabled
- [ ] Strong passwords on all accounts
- [ ] 2FA enabled where possible

### Network Security

- [ ] SSL/TLS enabled (HTTPS only)
- [ ] SSL certificate valid and trusted
- [ ] HTTP redirects to HTTPS
- [ ] Security headers configured:
  - X-Frame-Options
  - X-Content-Type-Options
  - X-XSS-Protection
  - Referrer-Policy
- [ ] DDoS protection enabled

### Compliance

- [ ] Privacy policy in place (if collecting user data)
- [ ] Terms of service in place
- [ ] GDPR compliance (if serving EU users)
- [ ] Data retention policy documented
- [ ] User data deletion process

## Maintenance

### Regular Tasks

- [ ] Weekly: Review logs for errors
- [ ] Weekly: Check disk space usage
- [ ] Weekly: Review security alerts
- [ ] Monthly: Update dependencies
- [ ] Monthly: Review and rotate secrets
- [ ] Monthly: Test backup restore
- [ ] Quarterly: Security audit
- [ ] Quarterly: Performance review

### Update Procedure

- [ ] Document update procedure
- [ ] Test updates in staging first
- [ ] Schedule maintenance window
- [ ] Notify users of downtime
- [ ] Create backup before update
- [ ] Update and test
- [ ] Rollback plan ready

### Monitoring & Alerts

- [ ] Setup uptime monitoring (UptimeRobot, etc.)
- [ ] Configure email/SMS alerts
- [ ] Monitor SSL certificate expiration
- [ ] Monitor disk space
- [ ] Monitor memory usage
- [ ] Monitor error rates
- [ ] Review metrics weekly

## Documentation

### Internal Documentation

- [ ] Document deployment procedure
- [ ] Document update procedure
- [ ] Document rollback procedure
- [ ] Document backup/restore procedure
- [ ] Document troubleshooting steps
- [ ] Document architecture diagram
- [ ] Document API endpoints
- [ ] Document environment variables

### User Documentation

- [ ] Create user guide
- [ ] Document API usage
- [ ] Create video tutorials (optional)
- [ ] FAQ page
- [ ] Troubleshooting guide
- [ ] Contact/support information

## Final Checks

### Pre-Launch

- [ ] All checklist items completed
- [ ] Staging environment tested
- [ ] Load testing completed
- [ ] Security scan completed
- [ ] Backup tested
- [ ] Monitoring configured
- [ ] Documentation complete
- [ ] Team trained on operations

### Launch

- [ ] Deploy to production
- [ ] Verify all services healthy
- [ ] Test critical user flows
- [ ] Monitor for first 24 hours
- [ ] Be ready for quick rollback

### Post-Launch

- [ ] Monitor error rates
- [ ] Monitor performance metrics
- [ ] Collect user feedback
- [ ] Address any issues quickly
- [ ] Document lessons learned

## Rollback Procedure

If something goes wrong:

1. [ ] Stop services:
   ```bash
   docker compose -f docker-compose.prod.yml down
   ```

2. [ ] Restore database backup:
   ```bash
   psql $DATABASE_URL < backups/backup_YYYYMMDD.sql
   ```

3. [ ] Checkout previous version:
   ```bash
   git checkout <previous-tag>
   ```

4. [ ] Rebuild and restart:
   ```bash
   docker compose -f docker-compose.prod.yml up -d --build
   ```

5. [ ] Verify health:
   ```bash
   curl http://localhost:8080/api/v1/health
   ```

6. [ ] Notify users of rollback

## Success Criteria

Deployment is successful when:

- ✅ All services are running and healthy
- ✅ Health check returns 200 OK
- ✅ All API endpoints respond correctly
- ✅ Frontend loads without errors
- ✅ Can create and run training jobs
- ✅ Logs are properly formatted
- ✅ Metrics are being collected
- ✅ Backups are working
- ✅ Monitoring is active
- ✅ SSL certificate is valid
- ✅ No security vulnerabilities
- ✅ Performance meets requirements
- ✅ Documentation is complete

## Support Contacts

- **Technical Lead**: [name@email.com]
- **DevOps**: [devops@email.com]
- **On-Call**: [oncall@email.com]
- **Emergency**: [phone number]

## Notes

Add any deployment-specific notes here:

---

**Deployment Date**: _______________
**Deployed By**: _______________
**Version**: _______________
**Sign-off**: _______________
