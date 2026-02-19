# Deployment Guide: Railway.app

Deploy Finetune Studio to Railway.app with automatic SSL, GitHub integration, and minimal configuration.

## What You Get

- **$5/month free credit** (enough for light usage)
- **Automatic SSL certificates**
- **GitHub integration** (auto-deploy on push)
- **Built-in monitoring**
- **Custom domains**
- **No server management**

## Cost Estimate

Railway charges based on usage:
- **Compute**: $0.000463/GB-hour
- **Memory**: $0.000231/GB-hour  
- **Estimated**: $10-20/month for moderate usage
- **Free tier**: $5/month credit

## Prerequisites

- GitHub account
- Railway account (sign up at https://railway.app)
- Kaggle API credentials
- Credit card (for usage beyond free tier)

## Step 1: Prepare Repository

### 1.1 Push to GitHub

```bash
# Initialize git (if not already)
git init
git add .
git commit -m "Initial commit"

# Create GitHub repo and push
git remote add origin https://github.com/yourusername/finetune-studio.git
git push -u origin main
```

### 1.2 Add Railway Configuration

Railway automatically detects Docker Compose, but we need to specify the production file.

Create `railway.json` in project root:

```json
{
  "build": {
    "builder": "DOCKERFILE",
    "dockerfilePath": "backend/Dockerfile.prod"
  },
  "deploy": {
    "startCommand": "./server",
    "healthcheckPath": "/api/v1/health",
    "healthcheckTimeout": 100,
    "restartPolicyType": "ON_FAILURE",
    "restartPolicyMaxRetries": 10
  }
}
```

Commit and push:
```bash
git add railway.json
git commit -m "Add Railway configuration"
git push
```

## Step 2: Create Railway Project

### 2.1 Login to Railway

1. Go to https://railway.app
2. Click **Login** → **Login with GitHub**
3. Authorize Railway

### 2.2 Create New Project

1. Click **New Project**
2. Select **Deploy from GitHub repo**
3. Choose your `finetune-studio` repository
4. Railway will detect the Dockerfile

## Step 3: Add Services

Railway needs separate services for each component.

### 3.1 Add PostgreSQL

1. Click **New** → **Database** → **Add PostgreSQL**
2. Railway automatically provisions the database
3. Note: Connection string is auto-generated

### 3.2 Add Redis (Optional, for caching)

1. Click **New** → **Database** → **Add Redis**
2. Useful for rate limiting and caching

## Step 4: Configure Environment Variables

### 4.1 Backend Service Variables

Click on your backend service → **Variables** tab:

```bash
# Application
APP_ENV=production
APP_VERSION=1.0.0
PORT=8080

# Database (auto-filled by Railway)
DATABASE_URL=${{Postgres.DATABASE_URL}}

# Database pool
DB_MAX_CONNECTIONS=25
DB_MAX_IDLE_CONNECTIONS=5
DB_CONNECTION_MAX_LIFETIME=5m

# MinIO/S3 - Use Railway's built-in storage or external S3
MINIO_ENDPOINT=s3.amazonaws.com
MINIO_USER=your_aws_access_key
MINIO_PASSWORD=your_aws_secret_key
MINIO_USE_SSL=true
MINIO_BUCKET=finetune-models

# Kaggle API
KAGGLE_USERNAME=your_kaggle_username
KAGGLE_KEY=your_kaggle_api_key

# Security
ALLOWED_ORIGINS=https://your-app.railway.app
RATE_LIMIT_REQUESTS_PER_MINUTE=100
RATE_LIMIT_EXPENSIVE_ENDPOINTS=10
MAX_REQUEST_SIZE_MB=10

# Workers
WORKER_POOL_SIZE=3
WORKER_TIMEOUT=24h

# Logging
LOG_LEVEL=info
LOG_FORMAT=json
METRICS_ENABLED=true
```

### 4.2 Frontend Service Variables

Create a separate service for frontend:

1. Click **New** → **GitHub Repo** → Select same repo
2. Set **Root Directory**: `frontend`
3. Set **Dockerfile Path**: `frontend/Dockerfile.prod`

Add variables:
```bash
API_URL=https://your-backend.railway.app
```

## Step 5: Configure Networking

### 5.1 Generate Domain

1. Click backend service → **Settings** → **Networking**
2. Click **Generate Domain**
3. Copy the URL (e.g., `your-app.railway.app`)

### 5.2 Update CORS

Update backend's `ALLOWED_ORIGINS`:
```bash
ALLOWED_ORIGINS=https://your-frontend.railway.app,https://your-custom-domain.com
```

### 5.3 Custom Domain (Optional)

1. Go to **Settings** → **Networking** → **Custom Domain**
2. Add your domain (e.g., `app.yourdomain.com`)
3. Add CNAME record to your DNS:
   ```
   CNAME app.yourdomain.com -> your-app.railway.app
   ```
4. Railway automatically provisions SSL

## Step 6: Deploy

### 6.1 Trigger Deployment

Railway automatically deploys on push to main branch.

Manual deployment:
1. Go to **Deployments** tab
2. Click **Deploy**

### 6.2 Monitor Deployment

Watch the build logs:
1. Click on the deployment
2. View **Build Logs** and **Deploy Logs**

Deployment takes 3-5 minutes.

## Step 7: Verify Deployment

### 7.1 Health Check

```bash
curl https://your-app.railway.app/api/v1/health
```

Expected response:
```json
{
  "status": "healthy",
  "version": "1.0.0",
  "services": {
    "database": {"status": "up"},
    "storage": {"status": "up"},
    "workers": {"status": "up"}
  }
}
```

### 7.2 Access Application

Open: `https://your-frontend.railway.app`

## Step 8: Setup Storage (MinIO Alternative)

Railway doesn't have built-in object storage. Options:

### Option A: AWS S3 (Recommended)

1. Create S3 bucket
2. Create IAM user with S3 access
3. Update environment variables:
```bash
MINIO_ENDPOINT=s3.amazonaws.com
MINIO_USER=your_aws_access_key_id
MINIO_PASSWORD=your_aws_secret_access_key
MINIO_USE_SSL=true
MINIO_BUCKET=your-bucket-name
```

### Option B: Cloudflare R2 (S3-compatible, cheaper)

1. Create R2 bucket at https://dash.cloudflare.com
2. Create API token
3. Update environment variables:
```bash
MINIO_ENDPOINT=your-account-id.r2.cloudflarestorage.com
MINIO_USER=your_r2_access_key
MINIO_PASSWORD=your_r2_secret_key
MINIO_USE_SSL=true
MINIO_BUCKET=your-bucket-name
```

### Option C: Railway Volume (Limited)

Railway provides persistent volumes:
1. Service → **Settings** → **Volumes**
2. Add volume: `/data` → 10GB
3. Note: More expensive than S3

## Step 9: Setup Monitoring

### 9.1 Railway Metrics

Railway provides built-in metrics:
- **CPU usage**
- **Memory usage**
- **Network traffic**
- **Request count**

Access: Service → **Metrics** tab

### 9.2 Application Metrics

View Prometheus metrics:
```bash
curl https://your-app.railway.app/api/v1/metrics
```

### 9.3 Logs

View logs in real-time:
1. Service → **Logs** tab
2. Filter by level (info, error, etc.)

## Step 10: Setup Automatic Backups

### 10.1 Database Backups

Railway doesn't auto-backup. Setup manual backups:

Create GitHub Action (`.github/workflows/backup.yml`):

```yaml
name: Database Backup

on:
  schedule:
    - cron: '0 2 * * *'  # Daily at 2 AM
  workflow_dispatch:

jobs:
  backup:
    runs-on: ubuntu-latest
    steps:
      - name: Backup Database
        env:
          DATABASE_URL: ${{ secrets.RAILWAY_DATABASE_URL }}
        run: |
          pg_dump $DATABASE_URL > backup_$(date +%Y%m%d).sql
          
      - name: Upload to S3
        uses: aws-actions/configure-aws-credentials@v1
        with:
          aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
          aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
          aws-region: us-east-1
          
      - name: Copy to S3
        run: |
          aws s3 cp backup_$(date +%Y%m%d).sql s3://your-backup-bucket/
```

## Continuous Deployment

### Auto-Deploy on Push

Railway automatically deploys when you push to main:

```bash
git add .
git commit -m "Update feature"
git push origin main
```

Railway will:
1. Detect changes
2. Build new Docker image
3. Run health checks
4. Deploy with zero downtime

### Rollback

If deployment fails:
1. Go to **Deployments** tab
2. Find previous successful deployment
3. Click **Redeploy**

## Cost Optimization

### Reduce Costs

1. **Reduce worker pool size**:
   ```bash
   WORKER_POOL_SIZE=2
   ```

2. **Use S3 instead of Railway volumes**
   - S3: $0.023/GB/month
   - Railway volume: $0.25/GB/month

3. **Optimize memory usage**:
   - Set resource limits in Dockerfile
   - Monitor memory usage

4. **Use Railway's sleep feature**:
   - Services sleep after 30 min inactivity (free tier)
   - Wake up on first request

### Monitor Costs

1. Go to **Account** → **Usage**
2. View current month's usage
3. Set up billing alerts

## Troubleshooting

### Build Fails

Check build logs:
1. **Deployments** → Click failed deployment
2. View **Build Logs**
3. Common issues:
   - Missing dependencies
   - Dockerfile errors
   - Out of memory during build

### Service Won't Start

Check deploy logs:
1. **Deployments** → **Deploy Logs**
2. Common issues:
   - Missing environment variables
   - Database connection failed
   - Port binding issues

### Database Connection Issues

1. Verify `DATABASE_URL` is set correctly
2. Check database service is running
3. Test connection:
   ```bash
   railway run psql $DATABASE_URL
   ```

### Out of Memory

1. Check **Metrics** tab
2. Reduce `WORKER_POOL_SIZE`
3. Upgrade plan if needed

## Scaling

### Vertical Scaling

Railway automatically scales resources based on usage.

To set limits:
1. Service → **Settings** → **Resources**
2. Set memory limit (e.g., 2GB)
3. Set CPU limit

### Horizontal Scaling

Railway supports multiple replicas:
1. Service → **Settings** → **Replicas**
2. Set number of instances
3. Load balancing is automatic

Note: Costs multiply by number of replicas.

## Security

Railway provides:
- ✅ Automatic SSL certificates
- ✅ Private networking between services
- ✅ Environment variable encryption
- ✅ DDoS protection

Additional steps:
1. **Rotate secrets regularly**
2. **Use Railway's secret management**
3. **Enable 2FA on Railway account**
4. **Restrict GitHub access**

## Support

- **Railway Docs**: https://docs.railway.app
- **Discord**: https://discord.gg/railway
- **Status**: https://status.railway.app

## Comparison: Railway vs Oracle Cloud

| Feature | Railway | Oracle Cloud |
|---------|---------|--------------|
| **Cost** | $10-20/month | Free forever |
| **Setup Time** | 10 minutes | 30-60 minutes |
| **SSL** | Automatic | Manual (Let's Encrypt) |
| **Scaling** | Automatic | Manual |
| **Monitoring** | Built-in | DIY |
| **Maintenance** | Zero | Regular updates |
| **Best For** | Quick deployment, low maintenance | Cost-sensitive, full control |

## Next Steps

- Setup custom domain
- Configure S3 storage
- Setup database backups
- Add monitoring alerts
- Configure CI/CD pipeline

Your ML training platform is now live! 🚀
