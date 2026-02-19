# Deployment Guide: Oracle Cloud Free Tier

This guide walks you through deploying Finetune Studio on Oracle Cloud's Always Free tier.

## What You Get (Free Forever)

- **2x AMD VMs**: 1 OCPU, 1GB RAM each
- **OR 4x ARM VMs**: 1 OCPU, 6GB RAM each (Recommended)
- **200GB Block Storage**
- **10TB Outbound Transfer/month**
- **No credit card required after trial**

## Prerequisites

- Oracle Cloud account (sign up at https://cloud.oracle.com/free)
- Domain name (optional, can use IP address)
- Kaggle API credentials

## Step 1: Create VM Instance

### 1.1 Launch Instance

1. Log into Oracle Cloud Console
2. Navigate to **Compute** → **Instances**
3. Click **Create Instance**

### 1.2 Configure Instance

**Name**: `finetune-studio`

**Image**: 
- Select **Ubuntu 22.04** (recommended)
- Or **Oracle Linux 8**

**Shape**:
- Click **Change Shape**
- Select **Ampere** (ARM-based)
- Choose **VM.Standard.A1.Flex**
- Set **OCPUs**: 2
- Set **Memory**: 12 GB
- This uses your full free tier allocation on one VM

**Networking**:
- Create new VCN or use existing
- Assign public IP: **Yes**
- Note down the public IP address

**SSH Keys**:
- Generate new key pair or upload existing
- **Download private key** (you'll need this!)

### 1.3 Create Instance

Click **Create** and wait 2-3 minutes for provisioning.

## Step 2: Configure Firewall

### 2.1 Security List Rules

1. Go to **Networking** → **Virtual Cloud Networks**
2. Click your VCN → **Security Lists** → **Default Security List**
3. Click **Add Ingress Rules**

Add these rules:

| Source CIDR | Protocol | Port Range | Description |
|-------------|----------|------------|-------------|
| 0.0.0.0/0   | TCP      | 80         | HTTP        |
| 0.0.0.0/0   | TCP      | 443        | HTTPS       |
| 0.0.0.0/0   | TCP      | 22         | SSH         |

### 2.2 OS Firewall (Ubuntu)

SSH into your instance:

```bash
ssh -i /path/to/private-key ubuntu@<PUBLIC_IP>
```

Configure firewall:

```bash
# Allow HTTP, HTTPS, SSH
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw allow 22/tcp

# Enable firewall
sudo ufw enable
```

## Step 3: Install Docker

```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Add user to docker group
sudo usermod -aG docker $USER

# Install Docker Compose
sudo apt install docker-compose-plugin -y

# Verify installation
docker --version
docker compose version

# Log out and back in for group changes to take effect
exit
```

SSH back in:
```bash
ssh -i /path/to/private-key ubuntu@<PUBLIC_IP>
```

## Step 4: Clone and Configure Application

```bash
# Clone repository
git clone https://github.com/yourusername/finetune-studio.git
cd finetune-studio

# Create .env file from template
cp .env.example .env

# Edit configuration
nano .env
```

### 4.1 Configure Environment Variables

Update these critical values in `.env`:

```bash
# Database - Change password!
DB_PASSWORD=your_secure_password_here

# MinIO - Change password!
MINIO_PASSWORD=your_secure_minio_password

# Kaggle API
KAGGLE_USERNAME=your_kaggle_username
KAGGLE_KEY=your_kaggle_api_key

# CORS - Add your domain
ALLOWED_ORIGINS=http://<YOUR_PUBLIC_IP>,https://yourdomain.com

# Production settings
APP_ENV=production
LOG_FORMAT=json
LOG_LEVEL=info
```

Save and exit (Ctrl+X, Y, Enter).

## Step 5: Deploy Application

```bash
# Build and start services
docker compose -f docker-compose.prod.yml up -d --build

# Check status
docker compose -f docker-compose.prod.yml ps

# View logs
docker compose -f docker-compose.prod.yml logs -f
```

Wait 2-3 minutes for all services to start.

## Step 6: Verify Deployment

### 6.1 Health Check

```bash
curl http://localhost:8080/api/v1/health
```

Expected response:
```json
{
  "status": "healthy",
  "version": "1.0.0",
  "uptime": "2m30s",
  "services": {
    "database": {
      "status": "up",
      "response_time": "2ms"
    },
    "storage": {
      "status": "up",
      "response_time": "5ms"
    },
    "workers": {
      "status": "up"
    }
  }
}
```

### 6.2 Access Application

Open browser: `http://<YOUR_PUBLIC_IP>`

You should see the Finetune Studio dashboard.

## Step 7: Setup SSL with Let's Encrypt (Optional but Recommended)

### 7.1 Install Certbot

```bash
sudo apt install certbot python3-certbot-nginx -y
```

### 7.2 Get SSL Certificate

```bash
# Stop frontend temporarily
docker compose -f docker-compose.prod.yml stop frontend

# Get certificate (replace with your domain)
sudo certbot certonly --standalone -d yourdomain.com

# Restart frontend
docker compose -f docker-compose.prod.yml start frontend
```

### 7.3 Configure Nginx for SSL

Create `frontend/nginx.ssl.conf`:

```nginx
server {
    listen 80;
    server_name yourdomain.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name yourdomain.com;

    ssl_certificate /etc/letsencrypt/live/yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/yourdomain.com/privkey.pem;

    # ... rest of nginx config ...
}
```

Update `docker-compose.prod.yml` to mount certificates:

```yaml
frontend:
  volumes:
    - /etc/letsencrypt:/etc/letsencrypt:ro
  ports:
    - "80:80"
    - "443:443"
```

Restart:
```bash
docker compose -f docker-compose.prod.yml restart frontend
```

## Step 8: Setup Automatic Backups

### 8.1 Database Backup Script

Create `scripts/backup.sh`:

```bash
#!/bin/bash
BACKUP_DIR="/home/ubuntu/finetune-studio/backups"
DATE=$(date +%Y%m%d_%H%M%S)

# Create backup directory
mkdir -p $BACKUP_DIR

# Backup database
docker exec finetune-postgres pg_dump -U finetune finetune_db > \
  $BACKUP_DIR/db_backup_$DATE.sql

# Keep only last 7 days
find $BACKUP_DIR -name "db_backup_*.sql" -mtime +7 -delete

echo "Backup completed: db_backup_$DATE.sql"
```

Make executable:
```bash
chmod +x scripts/backup.sh
```

### 8.2 Setup Cron Job

```bash
# Edit crontab
crontab -e

# Add daily backup at 2 AM
0 2 * * * /home/ubuntu/finetune-studio/scripts/backup.sh >> /home/ubuntu/backup.log 2>&1
```

## Step 9: Monitoring

### 9.1 View Logs

```bash
# All services
docker compose -f docker-compose.prod.yml logs -f

# Specific service
docker compose -f docker-compose.prod.yml logs -f backend

# Last 100 lines
docker compose -f docker-compose.prod.yml logs --tail=100 backend
```

### 9.2 Check Metrics

```bash
curl http://localhost:8080/api/v1/metrics
```

### 9.3 Resource Usage

```bash
# Docker stats
docker stats

# System resources
htop
```

## Maintenance

### Update Application

```bash
cd finetune-studio

# Pull latest changes
git pull

# Rebuild and restart
docker compose -f docker-compose.prod.yml up -d --build

# Clean old images
docker image prune -f
```

### Restart Services

```bash
# Restart all
docker compose -f docker-compose.prod.yml restart

# Restart specific service
docker compose -f docker-compose.prod.yml restart backend
```

### View Service Status

```bash
docker compose -f docker-compose.prod.yml ps
```

## Troubleshooting

### Services Won't Start

```bash
# Check logs
docker compose -f docker-compose.prod.yml logs

# Check disk space
df -h

# Check memory
free -h
```

### Database Connection Issues

```bash
# Check database is running
docker compose -f docker-compose.prod.yml ps postgres

# Check database logs
docker compose -f docker-compose.prod.yml logs postgres

# Restart database
docker compose -f docker-compose.prod.yml restart postgres
```

### Out of Memory

ARM instances have 12GB RAM total. If you run out:

```bash
# Check memory usage
docker stats

# Reduce worker pool size in .env
WORKER_POOL_SIZE=2

# Restart
docker compose -f docker-compose.prod.yml restart backend
```

### Port Already in Use

```bash
# Check what's using port 80
sudo lsof -i :80

# Kill process if needed
sudo kill <PID>
```

## Cost Optimization

The free tier is sufficient for:
- **Light usage**: 5-10 training jobs/day
- **Small models**: Up to 3B parameters
- **Moderate storage**: ~100GB of datasets/models

To optimize:
- Set `WORKER_POOL_SIZE=2` for lower memory usage
- Clean up old models regularly
- Use smaller batch sizes in training

## Security Best Practices

1. **Change all default passwords** in `.env`
2. **Enable firewall** (ufw)
3. **Keep system updated**: `sudo apt update && sudo apt upgrade`
4. **Use SSH keys** (disable password auth)
5. **Setup SSL** with Let's Encrypt
6. **Regular backups** (automated with cron)
7. **Monitor logs** for suspicious activity

## Support

- **Documentation**: See `/docs` folder
- **Issues**: GitHub Issues
- **Logs**: `docker compose logs`
- **Health**: `http://<IP>/api/v1/health`

## Next Steps

- Setup domain name and SSL
- Configure automated backups
- Setup monitoring alerts
- Customize training templates
- Add more datasets

Enjoy your free, production-ready ML training platform! 🚀
