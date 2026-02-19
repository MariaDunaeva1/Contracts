# Logs Streaming & Model Download System

Sistema completo de logs en tiempo real con SSE y descarga de modelos entrenados.

## Características Implementadas

### ✅ 1. Logs Streaming con SSE

#### Endpoints:

**GET /api/v1/jobs/:id/logs** (SSE Stream)
- Streaming en tiempo real de logs
- Actualización cada 2 segundos
- Heartbeat para mantener conexión
- Auto-agregación desde MinIO

**GET /api/v1/jobs/:id/logs** (JSON)
- Obtener logs en formato JSON
- Paginación con parámetro `limit`
- Útil para debugging

**POST /api/v1/jobs/:id/logs**
- Endpoint para que kernels suban logs
- Almacenamiento en DB

#### Características:
- Logs agregados desde MinIO cada 30s
- Broadcast vía SSE a clientes conectados
- Latencia < 3 segundos
- Formato: `[HH:MM:SS] [LEVEL] Message`

### ✅ 2. Model Management

#### Endpoints:

**GET /api/v1/models**
- Lista todos los modelos entrenados
- Filtros: `base_model`, `status`, `date_from`, `date_to`
- Paginación: `page`, `limit`
- Incluye métricas de training

**GET /api/v1/models/:id**
- Detalles completos del modelo
- Metadata (nombre, descripción, base_model)
- File sizes y paths
- Presigned URLs (24h expiry)
- Training metrics
- Evaluation results

**GET /api/v1/models/:id/download**
- Descarga ZIP con todos los archivos del modelo
- Streaming directo (no guarda en disco)
- Incluye:
  - `lora_adapters/` (adapter_model.bin, config.json)
  - `gguf/` (model-q4_k_m.gguf)
  - `README.md`
  - `metrics.json`

**POST /api/v1/models**
- Crear registro de modelo manualmente

**PUT /api/v1/models/:id**
- Actualizar metadata del modelo

**DELETE /api/v1/models/:id**
- Eliminar modelo

### ✅ 3. Evaluation System

#### Endpoints:

**POST /api/v1/models/:id/evaluate**
- Crear job de evaluación
- Compara base model vs fine-tuned
- Request body:
```json
{
  "test_set_path": "datasets/test_split.json",
  "base_model_name": "llama3.2:3b"
}
```

**GET /api/v1/evaluations/:id**
- Obtener resultados de evaluación
- Incluye:
  - Métricas (accuracy, F1, precision, recall)
  - Comparación side-by-side
  - 20 ejemplos con predicciones
  - Improvement deltas

**GET /api/v1/evaluations**
- Lista todas las evaluaciones
- Filtros: `model_id`, `status`

**PUT /api/v1/evaluations/:id**
- Actualizar resultados de evaluación

### ✅ 4. Worker Integration

El worker automáticamente crea un registro de modelo cuando un job se completa:

```go
func (w *WorkerPool) handleKernelComplete(job *models.Job) {
    // 1. Update job status
    updateJobStatus(job, "completed")
    
    // 2. Fetch metrics from MinIO
    metrics := fetchMetricsFromMinIO(job.ID)
    
    // 3. Create model record
    model := Model{
        JobID:              job.ID,
        LoRAAdaptersPath:   fmt.Sprintf("models/%d/lora_adapters", job.ID),
        GGUFPath:          fmt.Sprintf("models/%d/gguf", job.ID),
        BaseModel:         job.BaseModel,
        TrainingMetrics:   metrics,
        Status:            "ready",
    }
    db.Create(&model)
}
```

## Testing

### 1. Test SSE Logs Streaming

```bash
# Terminal 1: Start server
cd backend
go run ./cmd/server

# Terminal 2: Stream logs
curl -N http://localhost:8080/api/v1/jobs/1/logs
```

### 2. Test Model Download

```bash
# Download model
curl -O http://localhost:8080/api/v1/models/1/download

# Verify ZIP contents
unzip -l model-*.zip
```

### 3. Test Evaluation

```bash
# Create evaluation
curl -X POST http://localhost:8080/api/v1/models/1/evaluate \
  -H "Content-Type: application/json" \
  -d '{
    "test_set_path": "datasets/test_split.json",
    "base_model_name": "llama3.2:3b"
  }'

# Get results
curl http://localhost:8080/api/v1/evaluations/1
```

### 4. Automated Test Script

```bash
# Run complete test suite
bash scripts/test_logs_and_models.sh
```

### 5. Python Evaluation Script

```bash
# Install dependencies
pip install requests scikit-learn

# Run evaluation
python scripts/evaluate_model.py \
  --test-set data/contracts/ledgar_finetune_test.json \
  --base-model llama3.2:3b \
  --finetuned-model my-finetuned-model \
  --labels "positive" "negative" "neutral" \
  --output evaluation_results.json
```

## Database Models

### Model
```go
type Model struct {
    ID               uint
    Name             string
    Description      string
    BaseModel        string
    Type             string  // lora, full, gguf
    JobID            *uint
    StoragePath      string
    LoRAAdaptersPath string
    GGUFPath         string
    Files            JSON
    TrainingMetrics  JSON
    EvalResults      JSON
    Status           string  // uploading, ready, error
    TotalSize        int64
}
```

### Evaluation
```go
type Evaluation struct {
    ID            uint
    ModelID       uint
    JobID         *uint
    Status        string  // pending, running, completed, failed
    TestSetPath   string
    BaseModelName string
    FineTunedName string
    Results       JSON
    Examples      JSON
    StartedAt     *time.Time
    CompletedAt   *time.Time
    ErrorMessage  string
}
```

### LogEntry
```go
type LogEntry struct {
    ID        uint
    JobID     uint
    Level     string     // info, warn, error
    Message   string
    Source    string     // worker, kernel, system
    Timestamp time.Time
}
```

## MinIO Structure

```
buckets/
├── datasets/
│   └── {dataset_id}/
│       ├── data.json
│       └── test_split.json
├── models/
│   └── {job_id}/
│       ├── lora_adapters/
│       │   ├── adapter_model.bin
│       │   └── adapter_config.json
│       ├── gguf/
│       │   └── model-q4_k_m.gguf
│       ├── metrics.json
│       └── README.md
└── logs/
    └── {job_id}/
        ├── log_20240219_100000.json
        └── log_20240219_100030.json
```

## Performance Metrics

- ✅ SSE latency: < 3 segundos
- ✅ ZIP generation: < 10 segundos
- ✅ Download speed: Full bandwidth
- ✅ Log aggregation: Cada 30 segundos
- ✅ Presigned URL expiry: 24 horas

## Example Responses

### GET /api/v1/models/1
```json
{
  "model": {
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
    "total_size": 1073741824
  },
  "download_links": {
    "lora_adapters": "https://minio:9000/models/1/lora_adapters?X-Amz-...",
    "gguf": "https://minio:9000/models/1/gguf?X-Amz-..."
  }
}
```

### GET /api/v1/evaluations/1
```json
{
  "id": 1,
  "model_id": 1,
  "status": "completed",
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
      "f1_score_delta": "+24.3%"
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
    }
  ]
}
```

## Next Steps

1. Implementar UI para visualización de logs en tiempo real
2. Agregar gráficos de métricas de training
3. Implementar comparación side-by-side en UI
4. Agregar soporte para múltiples formatos de modelo
5. Implementar sistema de notificaciones cuando evaluación completa

## Troubleshooting

### Logs no aparecen
- Verificar que MinIO esté corriendo
- Verificar bucket "logs" existe
- Verificar permisos de lectura

### Download falla
- Verificar que modelo tenga status "ready"
- Verificar archivos existen en MinIO
- Verificar presigned URLs no expiraron

### Evaluation no funciona
- Verificar Ollama está corriendo (puerto 11434)
- Verificar modelos están cargados en Ollama
- Verificar test set existe y tiene formato correcto
