#!/bin/bash

# Database Restore Script for Finetune Studio
# Restores PostgreSQL database from backup

set -e

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

# Configuration
DB_CONTAINER="${DB_CONTAINER:-finetune-postgres}"
DB_USER="${DB_USER:-finetune}"
DB_NAME="${DB_NAME:-finetune_db}"
BACKUP_FILE="$1"

# Check if backup file provided
if [ -z "$BACKUP_FILE" ]; then
    log_error "Usage: $0 <backup_file>"
    log_info "Example: $0 backups/db_backup_20260219_120000.sql.gz"
    exit 1
fi

# Check if backup file exists
if [ ! -f "$BACKUP_FILE" ]; then
    log_error "Backup file not found: $BACKUP_FILE"
    exit 1
fi

# Check if database container is running
if ! docker ps | grep -q "$DB_CONTAINER"; then
    log_error "Database container is not running"
    log_info "Start it with: docker compose -f docker-compose.prod.yml up -d postgres"
    exit 1
fi

# Warning
log_warn "⚠️  WARNING: This will REPLACE the current database!"
log_warn "Database: $DB_NAME"
log_warn "Backup file: $BACKUP_FILE"
echo ""
read -p "Are you sure you want to continue? (yes/no): " CONFIRM

if [ "$CONFIRM" != "yes" ]; then
    log_info "Restore cancelled"
    exit 0
fi

# Create backup of current database before restore
log_info "Creating backup of current database..."
SAFETY_BACKUP="backups/pre_restore_backup_$(date +%Y%m%d_%H%M%S).sql"
mkdir -p backups
docker exec "$DB_CONTAINER" pg_dump -U "$DB_USER" "$DB_NAME" > "$SAFETY_BACKUP"
log_info "Safety backup created: $SAFETY_BACKUP"

# Stop backend to prevent connections
log_info "Stopping backend service..."
docker compose -f docker-compose.prod.yml stop backend 2>/dev/null || true

# Decompress if needed
TEMP_FILE="$BACKUP_FILE"
if [[ "$BACKUP_FILE" == *.gz ]]; then
    log_info "Decompressing backup file..."
    TEMP_FILE="/tmp/restore_$(date +%s).sql"
    gunzip -c "$BACKUP_FILE" > "$TEMP_FILE"
fi

# Drop and recreate database
log_info "Dropping existing database..."
docker exec "$DB_CONTAINER" psql -U "$DB_USER" -c "DROP DATABASE IF EXISTS $DB_NAME;"

log_info "Creating new database..."
docker exec "$DB_CONTAINER" psql -U "$DB_USER" -c "CREATE DATABASE $DB_NAME;"

# Restore database
log_info "Restoring database from backup..."
if cat "$TEMP_FILE" | docker exec -i "$DB_CONTAINER" psql -U "$DB_USER" -d "$DB_NAME"; then
    log_info "Database restored successfully!"
else
    log_error "Database restore failed!"
    log_warn "Attempting to restore from safety backup..."
    
    # Restore from safety backup
    docker exec "$DB_CONTAINER" psql -U "$DB_USER" -c "DROP DATABASE IF EXISTS $DB_NAME;"
    docker exec "$DB_CONTAINER" psql -U "$DB_USER" -c "CREATE DATABASE $DB_NAME;"
    cat "$SAFETY_BACKUP" | docker exec -i "$DB_CONTAINER" psql -U "$DB_USER" -d "$DB_NAME"
    
    log_info "Restored from safety backup"
    exit 1
fi

# Clean up temp file
if [ "$TEMP_FILE" != "$BACKUP_FILE" ]; then
    rm -f "$TEMP_FILE"
fi

# Restart backend
log_info "Starting backend service..."
docker compose -f docker-compose.prod.yml start backend

# Wait for backend to be healthy
log_info "Waiting for backend to be healthy..."
sleep 5

# Verify restore
log_info "Verifying restore..."
if curl -s http://localhost:8080/api/v1/health | grep -q "healthy\|ok"; then
    log_info "✓ Backend is healthy"
    log_info "✓ Database restore completed successfully!"
else
    log_warn "Backend health check failed, please verify manually"
fi

log_info "Safety backup kept at: $SAFETY_BACKUP"

exit 0
