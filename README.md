# ⚖️ LexAnalyzer - Sistema de Análisis de Contratos Legales

Sistema completo para análisis inteligente de contratos legales usando grandes modelos de lenguaje (LLMs), Generación Aumentada por Recuperación (RAG) y Agentes especializados.

## 🎯 ¿Qué hace este sistema?

Analiza contratos legales automáticamente y extrae de forma estructurada:
- ✅ Cláusulas principales
- ⚠️ Riesgos legales y penalizaciones
- 📊 Obligaciones de las partes
- 🔍 Comparación con contratos estándar
- 💡 Recomendaciones de negociación

## 🧠 Arquitectura

```text
Usuario → Frontend (Web) → Backend (Go) → RAG Service (Python/FastAPI) → Groq / Ollama (LLM)
                                                   ↓
                                         ChromaDB (Vector Store)
                                                   ↓
                                        4 Agentes Especializados
```

El sistema ahora incluye una **Base de Conocimiento (Knowledge Base)** que soporta la ingesta nativa de múltiples formatos de documentos legales para enriquecer el contexto del análisis: `.pdf`, `.docx`, `.txt`, `.csv`, `.md` y `.json/.jsonl`.

## ⚡ Quick Start

Asegúrate de tener Docker y Docker Compose instalados.

1. **Configurar API (Importante)**
   Crea un archivo `.env` en la raíz del proyecto y añade tu API Key de Groq (el proveedor recomendado por su extrema velocidad y gratuidad):
   ```env
   GROQ_API_KEY=tu_api_key_aqui
   LLM_PROVIDER=groq
   ```

2. **Lanzar todo el ecosistema**
   Todos los microservicios están orquestados; basta con un solo comando:
   ```bash
   docker-compose up --build
   ```
   *(Nota: La primera vez tardará varios minutos en descargar los modelos de embeddings de Python).*

3. **Acceder a la Interfaz**
   Abre en tu navegador: http://localhost:3000

## 🔑 Conceptos Importantes

### ✅ Selector de Modelos Dinámico
La interfaz de *Contract Analysis* incluye un menú desplegable que lista dinámicamente los modelos disponibles conectados al sistema, divididos en:
- **Modelos Base:** Modelos fundacionales listos para uso rápido (ej. Llama 3).
- **Modelos Fine-tuned:** Modelos reentrenados y especializados en el módulo de *Training* que han finalizado su aprendizaje con éxito.

### ☁️ Groq vs Ollama
**Groq (Recomendado - Por defecto):**
- API en la nube gratuita. Inferencia ultra-rápida (500+ tokens/segundo).
- No requiere hardware local sofisticado.

**Ollama (Opcional - Local):**
- 100% privado y offline.
- Requiere tener el modelo descargado localmente (`ollama pull llama3.2`) y cambiar en `.env`: `LLM_PROVIDER=ollama`.

### 🗄️ Almacenamiento y Archivos (MinIO)
LexAnalyzer utiliza **MinIO** (servidor de almacenamiento de objetos compatible con Amazon S3) dentro de Docker para resguardar todos los documentos originales subidos al sistema (`.pdf`, `.docx`, etc.).
- **Consola Web (UI):** Puedes explorar los archivos en crudo accediendo a `http://localhost:9001` (Usuario: `minioadmin` / Contraseña: `minioadmin`).
- **Funcionamiento:** El Backend de Go crea automáticamente el bucket necesario y sube los archivos de los usuarios. El microservicio RAG (Python) posteriormente descarga temporalmente fragmentos de estos archivos desde MinIO cuando necesita analizarlos para buscar cláusulas.

### 🎯 Fine-Tuning Integrado (Kaggle)
El sistema incluye un **Pipeline de Entrenamiento (Fine-Tuning)** completo gestionado desde la interfaz web, sin necesidad de tocar código:
1. **Sube tus propios Datasets:** Formato `.json` o `.jsonl` en la sección *Knowledge Base*.
2. **Lanza un Trabajo (Job):** Selecciona un modelo base, tu dataset, y haz clic en *Start Fine-Tuning*.
3. **Automatización en Kaggle:** El Backend de Go de LexAnalyzer se conecta automáticamente con la API de Kaggle, levanta un cuaderno jupyter temporal con aceleración GPU (T4x2 gratuitas) y comienza a entrenar tu modelo usando técnicas de parametrización eficiente (LoRA / Unsloth).
4. **Despliegue Inmediato:** Cuando Kaggle termina, el modelo entrenado se registra en el sistema y aparece automáticamente en el **Selector de Modelos** de la interfaz para poder usarlo en tus próximos análisis de contratos.

## 🛠️ Servicios Activos

Al levantar el sistema, se despliegan automáticamente los siguientes microservicios internos:

| Servicio | Puerto | Descripción |
|----------|--------|-------------|
| Frontend | 3000 | Interfaz web HTML/JS servida por Nginx |
| Backend API | 8080 | API REST ultra-rápida desarrollada en Go |
| RAG Service | 8001 | Motor de IA en Python (FastAPI, Langchain, ChromaDB) |
| PostgreSQL | 5432 | Base de datos relacional para metadatos |
| MinIO | 9000 | Object Storage (clon S3) para almacenar los PDFs y documentos |

## 🧪 Estructura Básica del Proyecto

```
contracts/
├── backend/                    # API Go (Gestión de base de datos y archivos)
│   └── services/               # Motor de IA en Python (RAG, Agentes, embeddings)
├── frontend/                   # Interfaz de Usuario (HTML, Vanilla JS, CSS)
├── chroma_db/                  # Base de datos vectorial persistente
├── data/                       # Datasets de ejemplo
└── docker-compose.yml          # Orquestador
```

## 🚨 Troubleshooting

- **Error "Failed to load" o "Cannot connection to backend":** Verifica que los servicios de Docker se hayan levantado completamente sin errores de memoria (sobre todo el RAG service). Asegúrate de acceder a través del puerto `3000`.
- **Análisis muy lento:** Si estás usando Ollama y no tienes GPU dedicada, el análisis de grandes contratos puede llevar varios minutos. Pásate a Groq configurando `.env`.
- **Despliegue en la nube:** El proyecto cuenta con un `docker-compose.prod.yml` optimizado para servidores como Oracle Cloud (compatible con ARM Ampere A1 de 24GB).

## 📄 Licencia

MIT License
