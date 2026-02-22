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
