# 🚀 Quick Start Guide

## Prerequisites

1. **Go 1.21+** installed
2. **PostgreSQL** running
3. **MinIO** running
4. **Docker** (optional, for Ollama)

## Step 1: Setup Environment

Create `.env` file in project root:

```bash
DATABASE_URL=postgresql://postgres:postgres@localhost:5432/finetune_studio
MINIO_ENDPOINT=localhost:9000
MINIO_USER=minioadmin
MINIO_PASSWORD=minioadmin
MINIO_USE_SSL=false
```

## Step 2: Start Services

### Option A: Using Docker Compose (Recommended)

```bash
docker-compose up -d
```

This starts:
- PostgreSQL (port 5432)
- MinIO (port 9000, console 9001)

### Option B: Manual Setup

**PostgreSQL:**
```bash
docker run -d \
  --name postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=finetune_studio \
  -p 5432:5432 \
  postgres:15
```

**MinIO:**
```bash
docker run -d \
  --name minio \
  -p 9000:9000 \
  -p 9001:9001 \
  -e MINIO_ROOT_USER=minioadmin \
  -e MINIO_ROOT_PASSWORD=minioadmin \
  minio/minio server /data --console-address ":9001"
```

## Step 3: Start Backend Server

```bash
cd backend
go run ./cmd/server
```

You should see:
```
✅ MinIO client initialized
📂 Bucket exists: datasets
📂 Bucket exists: models
📂 Bucket exists: logs
🚀 Worker Pool started with 5 workers
🚀 Server running on :8080
```

## Step 4: Test the System

### Test 1: Health Check

```bash
curl http://localhost:8080/api/v1/health
```

Expected response:
```json
{
  "status": "ok",
  "services": {
    "db": "up",
    "storage": "up"
  }
}
```

### Test 2: Create a Job

```bash
curl -X POST http://localhost:8080/api/v1/jobs \
  -H "Content-Type: application/json" \
  -d '{
    "dataset_id": 1,
    "configuration": {
      "epochs": 3,
      "base_model": "llama-3.2-3b"
    }
  }'
```

### Test 3: Stream Logs (Real-time)

Open `frontend/logs_viewer.html` in your browser, or:

```bash
curl -N http://localhost:8080/api/v1/jobs/1/logs
```

### Test 4: List Models

```bash
curl http://localhost:8080/api/v1/models | jq '.'
```

### Test 5: View Evaluation

Open `frontend/evaluation_viewer.html` in your browser.

## Step 5: Setup Ollama (For Evaluation)

```bash
# Start Ollama
docker run -d -p 11434:11434 --name ollama ollama/ollama

# Pull base model
docker exec ollama ollama pull llama3.2:3b

# Test Ollama
curl http://localhost:11434/api/tags
```

## Step 6: Run Evaluation

```bash
python scripts/evaluate_model.py \
  --test-set data/contracts/ledgar_finetune_test.json \
  --base-model llama3.2:3b \
  --finetuned-model my-model \
  --labels "contract" "non-contract" \
  --output evaluation_results.json
```

## Common Issues

### Issue: "Database connection failed"

**Solution:**
```bash
# Check PostgreSQL is running
docker ps | grep postgres

# Check connection
psql -h localhost -U postgres -d finetune_studio
```

### Issue: "MinIO connection failed"

**Solution:**
```bash
# Check MinIO is running
docker ps | grep minio

# Access MinIO console
open http://localhost:9001
# Login: minioadmin / minioadmin
```

### Issue: "Port already in use"

**Solution:**
```bash
# Find process using port 8080
netstat -ano | findstr :8080

# Kill process (Windows)
taskkill /PID <PID> /F

# Or change port in main.go:
# r.Run(":8081")
```

### Issue: "Ollama not responding"

**Solution:**
```bash
# Check Ollama is running
docker ps | grep ollama

# Restart Ollama
docker restart ollama

# Check logs
docker logs ollama
```

## Testing Workflow

### Complete Test Flow:

1. **Create Dataset** (if not exists)
```bash
curl -X POST http://localhost:8080/api/v1/datasets \
  -F "file=@data/contracts/ledgar_finetune_train.json" \
  -F "name=Legal Contracts" \
  -F "description=Contract classification dataset"
```

2. **Create Job**
```bash
curl -X POST http://localhost:8080/api/v1/jobs \
  -H "Content-Type: application/json" \
  -d '{
    "dataset_id": 1,
    "configuration": {"epochs": 3}
  }'
```

3. **Monitor Logs** (in browser)
```
Open: frontend/logs_viewer.html
Enter Job ID: 1
Click: Connect
```

4. **Wait for Completion**
```bash
# Check job status
curl http://localhost:8080/api/v1/jobs/1 | jq '.status'
```

5. **View Model**
```bash
# List models
curl http://localhost:8080/api/v1/models

# Get model details
curl http://localhost:8080/api/v1/models/1 | jq '.'
```

6. **Download Model**
```bash
curl -O http://localhost:8080/api/v1/models/1/download
unzip -l model-*.zip
```

7. **Create Evaluation**
```bash
curl -X POST http://localhost:8080/api/v1/models/1/evaluate \
  -H "Content-Type: application/json" \
  -d '{
    "test_set_path": "datasets/test_split.json",
    "base_model_name": "llama3.2:3b"
  }'
```

8. **View Evaluation** (in browser)
```
Open: frontend/evaluation_viewer.html
Enter Evaluation ID: 1
Click: Load Evaluation
```

## Automated Testing

Run the complete test suite:

```bash
bash scripts/test_logs_and_models.sh
```

## Development Tips

### Hot Reload (Optional)

Install Air for hot reload:
```bash
go install github.com/cosmtrek/air@latest
cd backend
air
```

### Database Migrations

GORM auto-migrates on startup. To reset:
```bash
# Drop all tables
psql -h localhost -U postgres -d finetune_studio -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"

# Restart server (will recreate tables)
go run ./cmd/server
```

### View MinIO Files

1. Open http://localhost:9001
2. Login: minioadmin / minioadmin
3. Browse buckets: datasets, models, logs

### Debug Logs

Add debug logging:
```go
log.SetLevel(log.DebugLevel)
```

## Production Deployment

### Build for Production

```bash
cd backend
go build -o server ./cmd/server
./server
```

### Environment Variables

Set in production:
```bash
export DATABASE_URL="postgresql://user:pass@prod-db:5432/finetune_studio"
export MINIO_ENDPOINT="minio.prod.com:9000"
export MINIO_USE_SSL=true
```

### Docker Deployment

```bash
docker build -t finetune-studio-backend ./backend
docker run -d \
  -p 8080:8080 \
  -e DATABASE_URL="..." \
  -e MINIO_ENDPOINT="..." \
  finetune-studio-backend
```

## Next Steps

1. ✅ System is running
2. ✅ Create your first job
3. ✅ Monitor logs in real-time
4. ✅ Download trained model
5. ✅ Run evaluation
6. ✅ Compare results

## Support

- Documentation: `backend/LOGS_AND_MODELS_README.md`
- Implementation: `IMPLEMENTATION_SUMMARY.md`
- Issues: Check logs in `backend/` directory

## Useful Commands

```bash
# Check all services
docker ps

# View logs
docker logs postgres
docker logs minio
docker logs ollama

# Stop all services
docker-compose down

# Clean up
docker-compose down -v  # Remove volumes too

# Rebuild
cd backend
go clean
go build -o server.exe ./cmd/server
```

Happy fine-tuning! 🎉
