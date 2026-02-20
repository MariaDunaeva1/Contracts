# Contract Analysis with RAG + Agents

Sistema de análisis de contratos usando Retrieval-Augmented Generation (RAG) con ChromaDB y agentes de IA.

## 🎯 Características

### 1. Búsqueda Semántica
- **ChromaDB** para almacenamiento vectorial
- **Sentence Transformers** para embeddings (all-MiniLM-L6-v2)
- Búsqueda por similitud coseno
- Filtros por metadata (tipo de cláusula, nivel de riesgo)

### 2. Sistema de Agentes
- **Extractor Agent**: Extrae cláusulas del contrato
- **Comparator Agent**: Compara con contratos históricos
- **Risk Agent**: Evalúa riesgo general del contrato
- **Orchestrator**: Coordina el workflow completo

### 3. Análisis Inteligente
- Extracción automática de cláusulas
- Comparación con contratos históricos
- Evaluación de favorabilidad (-1 a 1)
- Assessment de riesgo (high/medium/low)
- Executive summary automático

## 🏗️ Arquitectura

```
Frontend (HTML/JS)
    ↓
Go Backend (API Gateway)
    ↓
Python FastAPI (RAG Service)
    ↓
┌─────────────┬──────────────┐
│  ChromaDB   │  Ollama LLM  │
│  (Vectors)  │  (Analysis)  │
└─────────────┴──────────────┘
```

## 📦 Componentes

### Backend Services (Python)

**1. vector_service.py**
- Gestión de ChromaDB
- Indexación de contratos
- Búsqueda semántica
- CRUD de cláusulas

**2. llm_service.py**
- Cliente de Ollama
- Generación de texto
- Extracción de JSON
- Health checks

**3. agents/**
- `base_agent.py`: Clase base abstracta
- `extractor_agent.py`: Extracción de cláusulas
- `comparator_agent.py`: Comparación con históricos
- `risk_agent.py`: Evaluación de riesgo
- `orchestrator.py`: Coordinación de agentes

**4. rag_service.py**
- FastAPI server
- Endpoints REST
- Integración de servicios

### Go Backend

**contract.go**
- Handler para endpoints de contratos
- Proxy a servicio Python
- Gestión de errores

### Frontend

**contract-analysis.html**
- UI para upload de contratos
- Visualización de resultados
- Comparación side-by-side

**contract-analysis.js**
- Llamadas a API
- Renderizado de resultados
- Manejo de estados

## 🚀 Instalación

### 1. Instalar Ollama

```bash
# macOS/Linux
curl https://ollama.ai/install.sh | sh

# Windows
# Descargar desde https://ollama.ai/download

# Pull modelo
ollama pull llama3.2:3b
```

### 2. Instalar Dependencias Python

```bash
cd backend/services
pip install -r requirements.txt
```

### 3. Iniciar Servicios

**Opción A: Docker Compose (Recomendado)**

```bash
docker-compose up --build
```

**Opción B: Manual**

```bash
# Terminal 1: Ollama (si no está corriendo)
ollama serve

# Terminal 2: RAG Service
cd backend/services
python rag_service.py

# Terminal 3: Go Backend
cd backend
go run cmd/server/main.go

# Terminal 4: Frontend
cd frontend
python serve.py
```

## 📡 API Endpoints

### Análisis de Contratos

**POST /api/v1/contracts/analyze**
```json
{
  "contract_text": "Full contract text...",
  "contract_name": "Supplier Agreement 2024"
}
```

Response:
```json
{
  "contract_id": "contract_abc123",
  "status": "completed",
  "clauses": [...],
  "comparisons": [...],
  "risk_assessment": {
    "overall_risk": "medium",
    "risk_score": 0.6,
    "top_risks": [...],
    "executive_summary": "..."
  },
  "summary": {
    "total_clauses": 8,
    "high_risk_count": 2,
    "unfavorable_count": 1
  }
}
```

### Búsqueda Semántica

**POST /api/v1/clauses/search**
```json
{
  "query": "indemnification clause",
  "top_k": 5,
  "filters": {
    "clause_type": "indemnification"
  }
}
```

### Health Check

**GET /api/v1/rag/health**
```json
{
  "status": "healthy",
  "llm_service": "available",
  "vector_service": "healthy",
  "total_clauses_indexed": 150
}
```

### Estadísticas

**GET /api/v1/rag/stats**
```json
{
  "vector_database": {
    "total_clauses": 150,
    "collection_name": "contracts"
  },
  "llm_models": ["llama3.2:3b"],
  "llm_available": true
}
```

## 🎨 Uso del Frontend

1. Navega a `http://localhost:3000/contract-analysis.html`
2. Pega el texto del contrato
3. (Opcional) Añade un nombre al contrato
4. Click en "Analyze Contract"
5. Espera 30-60 segundos
6. Revisa los resultados:
   - Executive Summary con nivel de riesgo
   - Cláusulas extraídas con análisis
   - Comparación con contratos históricos
   - Recomendaciones

## 🔧 Configuración

### Variables de Entorno

```bash
# Ollama
OLLAMA_BASE_URL=http://localhost:11434

# ChromaDB
CHROMA_DB_PATH=./chroma_db

# RAG Service
RAG_SERVICE_PORT=8001
```

### Modelos LLM

Por defecto usa `llama3.2:3b`. Para cambiar:

```python
# En llm_service.py
self.model = "llama3.2:3b"  # Cambiar aquí
```

Modelos recomendados:
- `llama3.2:3b` - Rápido, buena calidad (default)
- `llama3.2:7b` - Mejor calidad, más lento
- `mistral:7b` - Alternativa rápida

## 📊 Ejemplo de Análisis

### Input
```
INDEMNIFICATION CLAUSE

Company shall indemnify and hold harmless Client from any and all 
claims, damages, losses, and expenses arising from Company's 
performance under this Agreement, without limitation.
```

### Output
```json
{
  "clause": {
    "type": "indemnification",
    "text": "Company shall indemnify...",
    "risk_level": "high",
    "reasoning": "Unlimited liability without caps"
  },
  "comparison": {
    "favorability_score": -0.7,
    "comparison": "This clause is significantly less favorable than 
                   historical contracts which typically include 
                   liability caps of $500K-$1M",
    "risks": [
      "Unlimited liability exposure",
      "No carve-outs for third-party claims"
    ],
    "recommendation": "Negotiate liability cap and exclusions"
  },
  "similar_clauses": [
    {
      "contract_name": "Supplier Agreement 2023",
      "similarity": 0.89,
      "text": "Company shall indemnify Client up to $500,000..."
    }
  ]
}
```

## 🐛 Troubleshooting

### Error: "Cannot connect to RAG service"
```bash
# Verificar que el servicio está corriendo
curl http://localhost:8001/health

# Si no responde, iniciar manualmente
cd backend/services
python rag_service.py
```

### Error: "LLM service unavailable"
```bash
# Verificar Ollama
ollama list

# Iniciar Ollama si no está corriendo
ollama serve

# Pull modelo si no existe
ollama pull llama3.2:3b
```

### Error: "ChromaDB initialization failed"
```bash
# Limpiar base de datos
rm -rf chroma_db/

# Reiniciar servicio
python rag_service.py
```

### Análisis muy lento
- Usar modelo más pequeño: `llama3.2:3b` en vez de `7b`
- Reducir `top_k` en búsquedas (default: 5)
- Limitar texto del contrato a 4000 caracteres

## 📈 Performance

### Benchmarks (llama3.2:3b)

- **Indexación**: ~2-3 segundos por contrato
- **Búsqueda semántica**: ~200-300ms
- **Extracción de cláusulas**: ~10-15 segundos
- **Análisis completo**: ~30-45 segundos

### Optimizaciones

1. **Batch Processing**: Indexar múltiples contratos en paralelo
2. **Caching**: Cachear embeddings de cláusulas comunes
3. **Model Quantization**: Usar modelos cuantizados (Q4_K_M)
4. **GPU**: Usar GPU para Ollama (10x más rápido)

## 🔐 Seguridad

- ✅ No se almacenan contratos completos (solo cláusulas)
- ✅ ChromaDB local (no cloud)
- ✅ Ollama local (no API externa)
- ✅ Sin logging de datos sensibles
- ⚠️ Añadir autenticación en producción
- ⚠️ Encriptar ChromaDB en producción

## 📚 Referencias

- [ChromaDB Docs](https://docs.trychroma.com/)
- [Sentence Transformers](https://www.sbert.net/)
- [Ollama](https://ollama.ai/)
- [FastAPI](https://fastapi.tiangolo.com/)

## 🤝 Contribuir

1. Fork el repositorio
2. Crea una rama: `git checkout -b feature/nueva-funcionalidad`
3. Commit: `git commit -am 'Add nueva funcionalidad'`
4. Push: `git push origin feature/nueva-funcionalidad`
5. Pull Request

## 📝 TODO

- [ ] Soporte multi-idioma
- [ ] OCR para contratos escaneados
- [ ] Export de reportes en PDF
- [ ] Integración con sistemas legales
- [ ] Fine-tuning del modelo en contratos específicos
- [ ] Dashboard de analytics
- [ ] API de webhooks para notificaciones
- [ ] Comparación de múltiples contratos simultáneos

## 📄 Licencia

MIT License - Ver LICENSE file
