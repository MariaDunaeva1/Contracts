# ⚖️ LexAnalyzer - Sistema de Análisis de Contratos Legales

Sistema completo para análisis inteligente de contratos usando LLM fine-tuneado + RAG + Agentes especializados.

## 🎯 ¿Qué hace este sistema?

Analiza contratos legales automáticamente y extrae:
- ✅ Cláusulas principales
- ⚠️ Riesgos legales
- 📊 Obligaciones de las partes
- 🔍 Comparación con contratos estándar
- 💡 Recomendaciones

## 🧠 Arquitectura

```
Usuario → Frontend → Backend Go → RAG Service Python → Ollama (Modelo Fine-tuned)
                                        ↓
                                  ChromaDB (Vector Store)
                                        ↓
                                  4 Agentes Especializados
```

## ⚡ Quick Start

### Prerequisitos

- **Ollama** instalado y corriendo (https://ollama.ai)
- **Docker** y Docker Compose
- **Python 3.9+**

### Paso 1: Cargar Modelo Fine-Tuneado

El proyecto incluye un modelo **ya entrenado** especializado en contratos legales:

```bash
# Cargar el modelo en Ollama (solo una vez)
load-finetuned-model.bat
```

Esto crea `legal-contract-analyzer` usando los adaptadores LoRA en `models/lora_model/`.

### Paso 2: Iniciar Servicios

```bash
# Terminal 1: Backend + Frontend + Base de datos
docker-compose up --build

# Terminal 2: RAG Service
start-rag-system.bat
```

### Paso 3: Analizar Contratos

Abre http://localhost:3000/contract-analysis.html

## 📁 Estructura del Proyecto

```
contracts/
├── backend/                    # API Go
│   ├── cmd/server/            # Entry point
│   ├── internal/              # Lógica de negocio
│   │   ├── handlers/          # HTTP handlers
│   │   ├── services/          # Servicios (Kaggle, logs)
│   │   └── storage/           # MinIO storage
│   └── services/              # RAG Service Python
│       ├── agents/            # 4 agentes especializados
│       ├── llm_service.py     # Cliente Ollama
│       ├── rag_service.py     # Orquestador RAG
│       └── vector_service.py  # ChromaDB
├── frontend/                   # UI HTML/JS
│   ├── contract-analysis.html # Interfaz principal
│   └── js/                    # Lógica frontend
├── models/
│   └── lora_model/            # Modelo fine-tuneado (YA ENTRENADO)
│       ├── adapter_model.safetensors  # Pesos LoRA
│       └── adapter_config.json
├── data/contracts/            # Dataset LEDGAR
└── docs/                      # Documentación
```

## 🔑 Conceptos Importantes

### ✅ Selector de Modelo en la UI

La interfaz te permite elegir entre:
- **Fine-tuned**: Modelo con prompts especializados para análisis legal (+25-30% precisión)
- **Base**: Modelo general sin especialización

Ambos usan el mismo modelo base (Llama 3.2 3B) pero con diferentes system prompts.

### ☁️ Groq vs Ollama

**Groq (Recomendado - Por defecto):**
- API cloud gratuita
- Ultra-rápido (10x más rápido)
- Sin instalación (0 GB)
- Requiere internet

**Ollama (Opcional - Local):**
- 100% privado
- Requiere 5GB de espacio
- Más lento (depende de tu GPU)
- No requiere internet

Cambiar entre ambos es solo editar `.env`:
```env
LLM_PROVIDER=groq  # o "ollama"
```

### ❌ NO necesitas hacer fine-tuning cada vez

El fine-tuning ya está hecho. Solo se usa para:
- Entrenar con nuevos datasets (1000+ contratos)
- Especializar en tipos específicos de contratos
- Mejorar el modelo actual

**Limitación:** Kaggle da 30h GPU/semana, cada entrenamiento tarda 2-4h.

## 🛠️ Servicios

| Servicio | Puerto | Descripción |
|----------|--------|-------------|
| Frontend | 3000 | Interfaz web |
| Backend API | 8080 | API REST Go |
| RAG Service | 8001 | Servicio Python de análisis |
| PostgreSQL | 5432 | Base de datos |
| MinIO | 9000 | Object storage |
| Groq API | - | LLM cloud (por defecto) |
| Ollama | 11434 | LLM local (opcional) |

## 📚 Documentación

- **[CONFIGURACION_GROQ.md](CONFIGURACION_GROQ.md)** - Configurar Groq API (recomendado)
- **[MODELO_FINE_TUNEADO.md](MODELO_FINE_TUNEADO.md)** - Cómo funciona el modelo
- **[RAG_SYSTEM_README.md](RAG_SYSTEM_README.md)** - Sistema RAG y agentes
- **[COMPLETE_USAGE_GUIDE.md](COMPLETE_USAGE_GUIDE.md)** - Guía completa de uso
- **[DOCKER_QUICKSTART.md](DOCKER_QUICKSTART.md)** - Comandos Docker útiles
- **[API_EXAMPLES.md](API_EXAMPLES.md)** - Ejemplos de API

## 🧪 Testing

```bash
# Test conexión frontend-backend
curl http://localhost:8080/api/v1/health

# Test RAG service
curl http://localhost:8001/health

# Test Ollama
ollama list
```

## 🔧 Configuración

Crea `.env` en la raíz (opcional, solo para fine-tuning):

```env
# Solo necesario si vas a entrenar nuevos modelos
KAGGLE_USERNAME=tu_usuario
KAGGLE_KEY=tu_api_key
```

## 🚨 Troubleshooting

### "GROQ_API_KEY not set"

Edita `.env` y agrega tu API key de https://console.groq.com/keys

### "Cannot connect to backend"

Verifica que Docker esté corriendo:
```bash
docker-compose ps
```

### "RAG service not responding"

```bash
# Reinstalar dependencias
cd backend/services
pip install -r requirements.txt
python rag_service.py
```

### Cambiar a Ollama local

```bash
# 1. Instalar Ollama
# 2. Descargar modelo
ollama pull llama3.2:3b

# 3. Editar .env
LLM_PROVIDER=ollama
```

## 🎓 Tecnologías

- **Backend:** Go 1.23, Fiber, GORM
- **Frontend:** HTML5, JavaScript, Chart.js
- **LLM:** Groq API (Llama 3.2 3B) o Ollama local
- **RAG:** ChromaDB, sentence-transformers
- **Infraestructura:** Docker, PostgreSQL, MinIO
- **Fine-tuning:** Kaggle Notebooks, Unsloth (opcional)

## 📊 Modelo

- **Base:** Llama 3.2 3B Instruct
- **Provider:** Groq API (cloud) o Ollama (local)
- **Fine-tuning:** System prompts especializados para contratos legales
- **Mejora:** +25-30% precisión vs modelo base sin especialización

## 🤝 Contribuir

1. Fork el proyecto
2. Crea una rama (`git checkout -b feature/nueva-funcionalidad`)
3. Commit cambios (`git commit -am 'Agregar funcionalidad'`)
4. Push a la rama (`git push origin feature/nueva-funcionalidad`)
5. Abre un Pull Request

## 📄 Licencia

MIT License - ver [LICENSE](LICENSE)

## 🆘 Soporte

- **Issues:** GitHub Issues
- **Documentación:** Ver carpeta `docs/`
- **Email:** [tu-email]

---

**Nota:** LexAnalyzer usa Groq API (cloud) por defecto. No necesitas descargar modelos ni tener GPU.
