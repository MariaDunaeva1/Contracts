# 📚 API Examples

Complete examples for all endpoints in the Logs Streaming & Model Download system.

## Base URL

```
http://localhost:8080/api/v1
```

---

## 📊 Logs Endpoints

### 1. Stream Logs (SSE)

**Endpoint:** `GET /api/v1/jobs/:id/logs`

**Description:** Real-time log streaming using Server-Sent Events

**cURL:**
```bash
curl -N http://localhost:8080/api/v1/jobs/1/logs
```

**JavaScript:**
```javascript
const eventSource = new EventSource('http://localhost:8080/api/v1/jobs/1/logs');

eventSource.onmessage = (event) => {
  console.log('New logs:', event.data);
  // Format: [HH:MM:SS] [LEVEL] Message
};

eventSource.onerror = (error) => {
  console.error('SSE error:', error);
  eventSource.close();
};

// Close connection
eventSource.close();
```

**Response (Stream):**
```
data: [10:30:45] [INFO] Starting training job
data: [10:30:46] [INFO] Loading dataset...
data: [10:30:50] [INFO] Epoch 1/5 - Loss: 0.45
data: [10:31:00] [WARN] GPU memory usage high
data: [10:31:15] [INFO] Epoch 2/5 - Loss: 0.32
```

### 2. Get Logs (JSON)

**Endpoint:** `GET /api/v1/jobs/:id/logs?limit=100`

**Description:** Get logs in JSON format with pagination

**cURL:**
```bash
curl http://localhost:8080/api/v1/jobs/1/logs?limit=50
```

**Response:**
```json
{
  "job_id": 1,
  "logs": [
    {
      "ID": 1,
      "job_id": 1,
      "level": "info",
      "message": "Starting training job",
      "source": "worker",
      "timestamp": "2024-02-19T10:30:45Z"
    },
    {
      "ID": 2,
      "job_id": 1,
      "level": "info",
      "message": "Loading dataset...",
      "source": "kernel",
      "timestamp": "2024-02-19T10:30:46Z"
    }
  ],
  "count": 2
}
```

### 3. Create Log Entry

**Endpoint:** `POST /api/v1/jobs/:id/logs`

**Description:** Create a new log entry (for kernels to push logs)

**cURL:**
```bash
curl -X POST http://localhost:8080/api/v1/jobs/1/logs \
  -H "Content-Type: application/json" \
  -d '{
    "level": "info",
    "message": "Training epoch 1 completed",
    "source": "kernel"
  }'
```

**Request Body:**
```json
{
  "level": "info",
  "message": "Training epoch 1 completed",
  "source": "kernel"
}
```

**Response:**
```json
{
  "ID": 3,
  "job_id": 1,
  "level": "info",
  "message": "Training epoch 1 completed",
  "source": "kernel",
  "timestamp": "2024-02-19T10:35:00Z"
}
```

---

## 🤖 Model Endpoints

### 1. List Models

**Endpoint:** `GET /api/v1/models`

**Query Parameters:**
- `page` (default: 1)
- `limit` (default: 10)
- `base_model` (filter by base model)
- `status` (filter by status: uploading, ready, error)
- `date_from` (filter by creation date)
- `date_to` (filter by creation date)

**cURL:**
```bash
# List all models
curl http://localhost:8080/api/v1/models

# Filter by base model
curl "http://localhost:8080/api/v1/models?base_model=llama-3.2-3b"

# Filter by status
curl "http://localhost:8080/api/v1/models?status=ready"

# Pagination
curl "http://localhost:8080/api/v1/models?page=2&limit=20"

# Multiple filters
curl "http://localhost:8080/api/v1/models?base_model=llama-3.2-3b&status=ready&page=1&limit=10"
```

**Response:**
```json
{
  "data": [
    {
      "ID": 1,
      "name": "Model from Job 1",
      "description": "Fine-tuned model from dataset: Sentiment Analysis",
      "base_model": "llama-3.2-3b",
      "type": "lora",
      "job_id": 1,
      "storage_path": "1",
      "lora_adapters_path": "1/lora_adapters",
      "gguf_path": "1/gguf",
      "training_metrics": {
        "loss": 0.25,
        "accuracy": 0.89,
        "epochs": 5
      },
      "status": "ready",
      "total_size": 1073741824,
      "created_at": "2024-02-19T10:00:00Z"
    }
  ],
  "total": 1,
  "page": 1,
  "limit": 10
}
```

### 2. Get Model Details

**Endpoint:** `GET /api/v1/models/:id`

**Description:** Get complete model details with presigned download URLs

**cURL:**
```bash
curl http://localhost:8080/api/v1/models/1
```

**Response:**
```json
{
  "model": {
    "ID": 1,
    "name": "Model from Job 1",
    "description": "Fine-tuned model from dataset: Sentiment Analysis",
    "base_model": "llama-3.2-3b",
    "type": "lora",
    "job_id": 1,
    "job": {
      "ID": 1,
      "dataset_id": 1,
      "status": "completed",
      "configuration": {
        "epochs": 5,
        "learning_rate": 0.0001
      }
    },
    "storage_path": "1",
    "lora_adapters_path": "1/lora_adapters",
    "gguf_path": "1/gguf",
    "files": {
      "adapter_model.bin": 524288000,
      "adapter_config.json": 1024,
      "model-q4_k_m.gguf": 549453824
    },
    "training_metrics": {
      "loss": 0.25,
      "accuracy": 0.89,
      "epochs": 5,
      "training_time": "2h 15m"
    },
    "eval_results": {
      "accuracy": 0.91,
      "f1_score": 0.89
    },
    "status": "ready",
    "total_size": 1073741824,
    "created_at": "2024-02-19T10:00:00Z"
  },
  "download_links": {
    "lora_adapters": "https://minio:9000/models/1/lora_adapters?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=...",
    "gguf": "https://minio:9000/models/1/gguf?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=..."
  },
  "files": {
    "adapter_model.bin": 524288000,
    "adapter_config.json": 1024,
    "model-q4_k_m.gguf": 549453824
  }
}
```

### 3. Download Model (ZIP)

**Endpoint:** `GET /api/v1/models/:id/download`

**Description:** Download complete model as ZIP archive (streaming)

**cURL:**
```bash
# Download to file
curl -O http://localhost:8080/api/v1/models/1/download

# Download with custom name
curl -o my-model.zip http://localhost:8080/api/v1/models/1/download

# Check headers only
curl -I http://localhost:8080/api/v1/models/1/download
```

**Response Headers:**
```
Content-Type: application/zip
Content-Disposition: attachment; filename=model-1-20240219.zip
Transfer-Encoding: chunked
```

**ZIP Contents:**
```
model-1-20240219.zip
├── 1/
│   ├── lora_adapters/
│   │   ├── adapter_model.bin
│   │   └── adapter_config.json
│   ├── gguf/
│   │   └── model-q4_k_m.gguf
│   ├── metrics.json
│   └── README.md
```

### 4. Create Model

**Endpoint:** `POST /api/v1/models`

**Description:** Create a new model record

**cURL:**
```bash
curl -X POST http://localhost:8080/api/v1/models \
  -H "Content-Type: application/json" \
  -d '{
    "name": "My Custom Model",
    "description": "Fine-tuned for sentiment analysis",
    "base_model": "llama-3.2-3b",
    "type": "lora",
    "job_id": 1,
    "storage_path": "1",
    "lora_adapters_path": "1/lora_adapters",
    "gguf_path": "1/gguf",
    "status": "ready"
  }'
```

**Request Body:**
```json
{
  "name": "My Custom Model",
  "description": "Fine-tuned for sentiment analysis",
  "base_model": "llama-3.2-3b",
  "type": "lora",
  "job_id": 1,
  "storage_path": "1",
  "lora_adapters_path": "1/lora_adapters",
  "gguf_path": "1/gguf",
  "status": "ready"
}
```

**Response:**
```json
{
  "ID": 2,
  "name": "My Custom Model",
  "description": "Fine-tuned for sentiment analysis",
  "base_model": "llama-3.2-3b",
  "type": "lora",
  "job_id": 1,
  "storage_path": "1",
  "status": "ready",
  "total_size": 1073741824,
  "created_at": "2024-02-19T11:00:00Z"
}
```

### 5. Update Model

**Endpoint:** `PUT /api/v1/models/:id`

**Description:** Update model metadata

**cURL:**
```bash
curl -X PUT http://localhost:8080/api/v1/models/1 \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Updated Model Name",
    "description": "Updated description",
    "status": "ready"
  }'
```

**Request Body:**
```json
{
  "name": "Updated Model Name",
  "description": "Updated description",
  "status": "ready",
  "training_metrics": {
    "loss": 0.20,
    "accuracy": 0.92
  }
}
```

**Response:**
```json
{
  "ID": 1,
  "name": "Updated Model Name",
  "description": "Updated description",
  "status": "ready",
  "training_metrics": {
    "loss": 0.20,
    "accuracy": 0.92
  },
  "updated_at": "2024-02-19T11:30:00Z"
}
```

### 6. Delete Model

**Endpoint:** `DELETE /api/v1/models/:id`

**Description:** Delete a model record

**cURL:**
```bash
curl -X DELETE http://localhost:8080/api/v1/models/1
```

**Response:**
```json
{
  "message": "Model deleted successfully"
}
```

---

## 🎯 Evaluation Endpoints

### 1. Create Evaluation

**Endpoint:** `POST /api/v1/models/:id/evaluate`

**Description:** Create a new evaluation job

**cURL:**
```bash
curl -X POST http://localhost:8080/api/v1/models/1/evaluate \
  -H "Content-Type: application/json" \
  -d '{
    "test_set_path": "datasets/test_split.json",
    "base_model_name": "llama3.2:3b"
  }'
```

**Request Body:**
```json
{
  "test_set_path": "datasets/test_split.json",
  "base_model_name": "llama3.2:3b"
}
```

**Response:**
```json
{
  "evaluation_id": 1,
  "status": "pending",
  "message": "Evaluation job created"
}
```

### 2. Get Evaluation

**Endpoint:** `GET /api/v1/evaluations/:id`

**Description:** Get evaluation results

**cURL:**
```bash
curl http://localhost:8080/api/v1/evaluations/1
```

**Response:**
```json
{
  "id": 1,
  "model_id": 1,
  "status": "completed",
  "test_set_path": "datasets/test_split.json",
  "base_model_name": "llama3.2:3b",
  "fine_tuned_name": "Model from Job 1",
  "results": {
    "base_model": {
      "accuracy": 0.72,
      "f1_score": 0.70,
      "precision": 0.71,
      "recall": 0.69,
      "avg_response_time_ms": 450
    },
    "fine_tuned": {
      "accuracy": 0.89,
      "f1_score": 0.87,
      "precision": 0.88,
      "recall": 0.86,
      "avg_response_time_ms": 460
    },
    "improvement": {
      "accuracy_delta": "+23.6%",
      "f1_score_delta": "+24.3%",
      "precision_delta": "+23.9%",
      "recall_delta": "+24.6%"
    }
  },
  "examples": [
    {
      "input": "I love this product!",
      "expected": "positive",
      "base_model_output": "neutral",
      "base_model_correct": false,
      "fine_tuned_output": "positive",
      "fine_tuned_correct": true,
      "winner": "fine_tuned"
    },
    {
      "input": "This is terrible",
      "expected": "negative",
      "base_model_output": "negative",
      "base_model_correct": true,
      "fine_tuned_output": "negative",
      "fine_tuned_correct": true,
      "winner": "tie"
    }
  ],
  "started_at": "2024-02-19T12:00:00Z",
  "completed_at": "2024-02-19T12:05:00Z"
}
```

### 3. List Evaluations

**Endpoint:** `GET /api/v1/evaluations`

**Query Parameters:**
- `page` (default: 1)
- `limit` (default: 10)
- `model_id` (filter by model)
- `status` (filter by status)

**cURL:**
```bash
# List all evaluations
curl http://localhost:8080/api/v1/evaluations

# Filter by model
curl "http://localhost:8080/api/v1/evaluations?model_id=1"

# Filter by status
curl "http://localhost:8080/api/v1/evaluations?status=completed"
```

**Response:**
```json
{
  "data": [
    {
      "ID": 1,
      "model_id": 1,
      "status": "completed",
      "base_model_name": "llama3.2:3b",
      "fine_tuned_name": "Model from Job 1",
      "started_at": "2024-02-19T12:00:00Z",
      "completed_at": "2024-02-19T12:05:00Z"
    }
  ],
  "total": 1,
  "page": 1,
  "limit": 10
}
```

### 4. Update Evaluation

**Endpoint:** `PUT /api/v1/evaluations/:id`

**Description:** Update evaluation results (typically called by evaluation worker)

**cURL:**
```bash
curl -X PUT http://localhost:8080/api/v1/evaluations/1 \
  -H "Content-Type: application/json" \
  -d '{
    "status": "completed",
    "results": {
      "base_model": {
        "accuracy": 0.72,
        "f1_score": 0.70
      },
      "fine_tuned": {
        "accuracy": 0.89,
        "f1_score": 0.87
      }
    }
  }'
```

**Request Body:**
```json
{
  "status": "completed",
  "results": {
    "base_model": {
      "accuracy": 0.72,
      "f1_score": 0.70
    },
    "fine_tuned": {
      "accuracy": 0.89,
      "f1_score": 0.87
    }
  },
  "examples": [
    {
      "input": "Test input",
      "expected": "positive",
      "base_model_output": "neutral",
      "fine_tuned_output": "positive"
    }
  ]
}
```

**Response:**
```json
{
  "ID": 1,
  "model_id": 1,
  "status": "completed",
  "results": {...},
  "examples": [...],
  "completed_at": "2024-02-19T12:05:00Z"
}
```

---

## 🔄 Complete Workflow Example

### Step 1: Create Job
```bash
JOB_ID=$(curl -s -X POST http://localhost:8080/api/v1/jobs \
  -H "Content-Type: application/json" \
  -d '{"dataset_id": 1, "configuration": {"epochs": 3}}' \
  | jq -r '.ID')

echo "Created Job ID: $JOB_ID"
```

### Step 2: Monitor Logs
```bash
# Stream logs in real-time
curl -N http://localhost:8080/api/v1/jobs/$JOB_ID/logs
```

### Step 3: Check Job Status
```bash
STATUS=$(curl -s http://localhost:8080/api/v1/jobs/$JOB_ID | jq -r '.status')
echo "Job Status: $STATUS"
```

### Step 4: Get Model
```bash
MODEL_ID=$(curl -s "http://localhost:8080/api/v1/models?job_id=$JOB_ID" \
  | jq -r '.data[0].ID')

echo "Model ID: $MODEL_ID"
```

### Step 5: Download Model
```bash
curl -O http://localhost:8080/api/v1/models/$MODEL_ID/download
unzip -l model-*.zip
```

### Step 6: Create Evaluation
```bash
EVAL_ID=$(curl -s -X POST http://localhost:8080/api/v1/models/$MODEL_ID/evaluate \
  -H "Content-Type: application/json" \
  -d '{"test_set_path": "datasets/test_split.json"}' \
  | jq -r '.evaluation_id')

echo "Evaluation ID: $EVAL_ID"
```

### Step 7: Get Evaluation Results
```bash
curl -s http://localhost:8080/api/v1/evaluations/$EVAL_ID | jq '.'
```

---

## 🐍 Python Examples

### Stream Logs
```python
import requests
import json

def stream_logs(job_id):
    url = f'http://localhost:8080/api/v1/jobs/{job_id}/logs'
    
    with requests.get(url, stream=True) as response:
        for line in response.iter_lines():
            if line:
                decoded = line.decode('utf-8')
                if decoded.startswith('data: '):
                    log_data = decoded[6:]  # Remove 'data: ' prefix
                    print(log_data)

stream_logs(1)
```

### Download Model
```python
import requests

def download_model(model_id, output_file='model.zip'):
    url = f'http://localhost:8080/api/v1/models/{model_id}/download'
    
    with requests.get(url, stream=True) as response:
        response.raise_for_status()
        with open(output_file, 'wb') as f:
            for chunk in response.iter_content(chunk_size=8192):
                f.write(chunk)
    
    print(f'Downloaded to {output_file}')

download_model(1)
```

### Create and Monitor Evaluation
```python
import requests
import time

def create_evaluation(model_id):
    url = f'http://localhost:8080/api/v1/models/{model_id}/evaluate'
    data = {
        'test_set_path': 'datasets/test_split.json',
        'base_model_name': 'llama3.2:3b'
    }
    
    response = requests.post(url, json=data)
    result = response.json()
    return result['evaluation_id']

def get_evaluation(eval_id):
    url = f'http://localhost:8080/api/v1/evaluations/{eval_id}'
    response = requests.get(url)
    return response.json()

# Create evaluation
eval_id = create_evaluation(1)
print(f'Created evaluation: {eval_id}')

# Poll until complete
while True:
    eval_data = get_evaluation(eval_id)
    status = eval_data['status']
    print(f'Status: {status}')
    
    if status in ['completed', 'failed']:
        print('Results:', eval_data['results'])
        break
    
    time.sleep(5)
```

---

## 📝 Notes

- All timestamps are in ISO 8601 format (UTC)
- Presigned URLs expire after 24 hours
- SSE connections timeout after 5 minutes of inactivity
- Maximum file upload size: 500MB
- ZIP downloads are streamed (no server-side storage)
- Evaluation jobs run asynchronously

## 🔐 Authentication

Currently, the API does not require authentication. In production, add:

```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
  http://localhost:8080/api/v1/models
```
