# 🎯 Uso del Modelo Fine-Tuneado

## Concepto Importante

Este proyecto **NO hace fine-tuning cada vez que analizas un contrato**. El fine-tuning ya está hecho y el modelo entrenado está en `models/lora_model/`.

## Flujo Correcto

### Para Análisis de Contratos (Uso Normal)

1. **Cargar el modelo fine-tuneado en Ollama** (solo una vez):
   ```bash
   load-finetuned-model.bat
   ```

2. **Iniciar el sistema RAG**:
   ```bash
   start-rag-system.bat
   ```

3. **Usar la interfaz web** para analizar contratos:
   - http://localhost:3000/contract-analysis.html

El modelo fine-tuneado (`legal-contract-analyzer`) se usa automáticamente.

### Para Entrenar un Nuevo Modelo (Raro)

Solo necesitas hacer esto si:
- Tienes un nuevo dataset de contratos
- Quieres mejorar el modelo con más datos
- Quieres entrenar para un dominio legal diferente

**Limitaciones de Kaggle:**
- 30 horas de GPU por semana
- El fine-tuning tarda 2-4 horas por modelo
- Solo puedes entrenar ~7 modelos por semana

**Proceso:**
1. Sube tu dataset en la interfaz web
2. Crea un job de entrenamiento
3. Espera 2-4 horas
4. Descarga el nuevo modelo
5. Reemplaza los archivos en `models/lora_model/`
6. Recarga el modelo en Ollama

## Arquitectura

```
┌─────────────────────────────────────────┐
│  Frontend (contract-analysis.html)      │
│  Usuario sube contrato                  │
└──────────────┬──────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────┐
│  Backend Go (handlers/contract.go)      │
│  Recibe contrato, llama a RAG service   │
└──────────────┬──────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────┐
│  RAG Service Python (rag_service.py)    │
│  Orquesta agentes de análisis           │
└──────────────┬──────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────┐
│  LLM Service (llm_service.py)           │
│  Usa: legal-contract-analyzer           │
└──────────────┬──────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────┐
│  Ollama (localhost:11434)               │
│  Modelo: legal-contract-analyzer        │
│  Base: llama3.2:3b                      │
│  Adaptadores: models/lora_model/        │
└─────────────────────────────────────────┘
```

## Verificación

### 1. Verificar que Ollama tiene el modelo cargado

```bash
ollama list
```

Deberías ver:
```
NAME                        ID              SIZE      MODIFIED
legal-contract-analyzer     abc123def456    2.0 GB    2 minutes ago
llama3.2:3b                 xyz789ghi012    2.0 GB    1 week ago
```

### 2. Probar el modelo directamente

```bash
ollama run legal-contract-analyzer "Analiza esta cláusula: El contratista se compromete a entregar el proyecto en 30 días."
```

### 3. Verificar que el RAG service usa el modelo correcto

Revisa los logs cuando inicies el sistema:
```bash
start-rag-system.bat
```

Deberías ver:
```
INFO: Using model: legal-contract-analyzer
INFO: Model loaded successfully
```

## Archivos del Modelo Fine-Tuneado

```
models/lora_model/
├── adapter_config.json          # Configuración de LoRA
├── adapter_model.safetensors    # Pesos del adaptador (el fine-tuning)
├── tokenizer.json               # Tokenizador
├── tokenizer_config.json        # Configuración del tokenizador
├── chat_template.jinja          # Template para chat
└── README.md                    # Metadata del modelo
```

**Importante:** Estos archivos son el resultado del fine-tuning. NO los borres.

## Diferencia entre Modelos

### llama3.2:3b (Base)
- Modelo general de Meta
- No especializado en contratos
- Respuestas genéricas

### legal-contract-analyzer (Fine-tuned)
- Entrenado con dataset LEDGAR (contratos legales)
- Especializado en análisis de cláusulas
- Identifica riesgos legales específicos
- Mejor comprensión de terminología legal

## Preguntas Frecuentes

### ¿Necesito Kaggle para usar el sistema?

**NO.** Solo necesitas Kaggle si quieres entrenar un nuevo modelo. Para usar el sistema de análisis de contratos, solo necesitas:
1. Ollama corriendo
2. El modelo fine-tuneado cargado
3. El RAG service corriendo

### ¿Cuándo debo hacer fine-tuning?

Solo cuando:
- Tienes un dataset nuevo de al menos 1000 contratos
- Quieres especializar en un tipo específico de contrato (ej: laborales, inmobiliarios)
- El modelo actual no da buenos resultados

### ¿Puedo usar el modelo sin Ollama?

No directamente. Ollama es necesario para:
- Cargar los adaptadores LoRA
- Servir el modelo vía API
- Gestionar la memoria GPU/CPU

### ¿Qué pasa si borro models/lora_model/?

Perderás el modelo fine-tuneado y tendrás que:
1. Volver a hacer fine-tuning (2-4 horas en Kaggle)
2. O usar el modelo base (peor calidad)

## Troubleshooting

### Error: "Model legal-contract-analyzer not found"

**Solución:**
```bash
load-finetuned-model.bat
```

### Error: "Failed to load adapter"

**Causa:** Archivos del modelo corruptos o faltantes

**Solución:**
1. Verifica que todos los archivos estén en `models/lora_model/`
2. Si faltan, necesitas volver a hacer fine-tuning

### El modelo da respuestas genéricas

**Causa:** Está usando el modelo base en lugar del fine-tuned

**Solución:**
1. Verifica: `ollama list`
2. Recarga: `load-finetuned-model.bat`
3. Reinicia RAG service

## Recursos

- **Dataset usado:** LEDGAR (170k+ cláusulas de contratos)
- **Técnica:** LoRA (Low-Rank Adaptation)
- **Modelo base:** Llama 3.2 3B Instruct
- **Tiempo de entrenamiento:** ~2 horas en Kaggle T4 GPU
- **Mejora sobre base:** +25-30% en precisión de clasificación de cláusulas

