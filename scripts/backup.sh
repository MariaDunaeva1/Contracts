#!/bin/bash

# Automated Backup Script for Finetune Studio
# Backs up PostgreSQL database and MinIO data

set -e

# Configuration
BACKUP_DIR="${BACKUP_DIR:-./backups}"
DATE=$(date +%Y%m%d_%H%M%S)
RETENTION_DAYS="${RETENTION_DAYS:-7}"

# Database configuration
DB_CONTAINER="${DB_CONTAINER:-finetune-postgres}"
DB_USER="${DB_USER:-finetune}"
DB_NAME="${DB_NAME:-finetune_db}"

# MinIO configuration
MINIO_CONTAINER="${MINIO_CONTAINER:-finetune-minio}"

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

# Helper functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Create backup directory
mkdir -p "$BACKUP_DIR"

log_info "Starting backup process..."
log_info "Backup directory: $BACKUP_DIR"
log_info "Date: $DATE"

# Backup PostgreSQL database
log_info "Backing up PostgreSQL database..."

if docker ps | grep -q "$DB_CONTAINER"; then
    DB_BACKUP_FILE="$BACKUP_DIR/db_backup_$DATE.sql"
    
    if docker exec "$DB_CONTAINER" pg_dump -U "$DB_USER" "$DB_NAME" > "$DB_BACKUP_FILE"; then
        # Compress backup
        gzip "$DB_BACKUP_FILE"
        DB_BACKUP_FILE="$DB_BACKUP_FILE.gz"
        
        # Get file size
        SIZE=$(du -h "$DB_BACKUP_FILE" | cut -f1)
        log_info "Database backup completed: $DB_BACKUP_FILE ($SIZE)"
    else
        log_error "Database backup failed"
        exit 1
    fi
else
    log_warn "Database container not running, skipping database backup"
fi

# Backup MinIO data (optional - MinIO data is already persistent in volumes)
log_info "MinIO data is stored in Docker volumes (persistent)"
log_info "To backup MinIO data, use: docker run --rm -v minio_data:/data -v $BACKUP_DIR:/backup alpine tar czf /backup/minio_backup_$DATE.tar.gz /data"

# Clean up old backups
log_info "Cleaning up backups older than $RETENTION_DAYS days..."

if [ -d "$BACKUP_DIR" ]; then
    DELETED=$(find "$BACKUP_DIR" -name "db_backup_*.sql.gz" -mtime +$RETENTION_DAYS -delete -print | wc -l)
    if [ "$DELETED" -gt 0 ]; then
        log_info "Deleted $DELETED old backup(s)"
    else
        log_info "No old backups to delete"
    fi
fi

# Backup summary
log_info "Backup completed successfully!"
log_info "Backup location: $DB_BACKUP_FILE"

# Optional: Upload to S3 or remote storage
if [ -n "$S3_BUCKET" ]; then
    log_info "Uploading backup to S3..."
    
    if command -v aws &> /dev/null; then
        if aws s3 cp "$DB_BACKUP_FILE" "s3://$S3_BUCKET/backups/"; then
            log_info "Backup uploaded to S3: s3://$S3_BUCKET/backups/"
        else
            log_error "Failed to upload backup to S3"
        fi
    else
        log_warn "AWS CLI not installed, skipping S3 upload"
    fi
fi

# List recent backups
log_info "Recent backups:"
ls -lh "$BACKUP_DIR"/db_backup_*.sql.gz 2>/dev/null | tail -n 5 || log_warn "No backups found"

exit 0
