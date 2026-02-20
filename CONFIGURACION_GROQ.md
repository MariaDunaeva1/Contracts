# 🚀 Configuración con Groq API (Recomendado)

## ¿Por qué Groq?

- ✅ **Gratis**: API gratuita con límites generosos
- ✅ **Rápido**: Inferencia ultra-rápida (hasta 10x más rápido que Ollama local)
- ✅ **Sin instalación**: No necesitas descargar modelos (ahorra 5GB)
- ✅ **Sin GPU**: Funciona en cualquier computadora
- ✅ **Mismo modelo**: Llama 3.2 3B disponible

## Paso 1: Obtener API Key de Groq

1. Ve a https://console.groq.com
2. Regístrate o inicia sesión (gratis)
3. Ve a "API Keys" en el menú
4. Click en "Create API Key"
5. Copia la key (empieza con `gsk_...`)

## Paso 2: Configurar el Proyecto

Edita el archivo `.env` en la raíz del proyecto:

```env
# LLM Provider
LLM_PROVIDER=groq

# Groq API Key
GROQ_API_KEY=gsk_tu_api_key_aqui

# No necesitas Ollama si usas Groq
# OLLAMA_BASE_URL=http://localhost:11434
```

## Paso 3: Iniciar el Sistema

```bash
# 1. Iniciar backend + frontend + base de datos
docker-compose up --build

# 2. En otra terminal, iniciar RAG service
start-rag-system.bat
```

El script detectará automáticamente que estás usando Groq.

## Paso 4: Usar el Sistema

1. Abre http://localhost:3000/contract-analysis.html
2. Verás un checkbox: "Use Fine-Tuned Model"
   - ✅ Marcado: Usa prompts especializados para contratos legales
   - ⬜ Desmarcado: Usa el modelo base sin especialización
3. Pega tu contrato y analiza

## Comparación: Groq vs Ollama

| Característica | Groq (Cloud) | Ollama (Local) |
|----------------|--------------|----------------|
| Costo | Gratis | Gratis |
| Velocidad | Ultra-rápido | Depende de tu GPU |
| Espacio en disco | 0 GB | ~5 GB |
| Requiere GPU | No | Recomendado |
| Privacidad | Datos van a Groq | 100% local |
| Setup | 2 minutos | 15-30 minutos |
| Internet | Requerido | No requerido |

## Modelos Disponibles en Groq

El sistema usa automáticamente:
- **Base model**: `llama-3.2-3b-preview`
- **Fine-tuned**: Mismo modelo + system prompt especializado

Otros modelos disponibles (puedes cambiarlos en `llm_service.py`):
- `llama-3.2-1b-preview` (más rápido, menos preciso)
- `llama-3.1-8b-instant` (más lento, más preciso)
- `mixtral-8x7b-32768` (mejor calidad, más lento)

## Límites de Groq (Tier Gratuito)

- **Requests por minuto**: 30
- **Requests por día**: 14,400
- **Tokens por minuto**: 6,000

Para uso normal de análisis de contratos, estos límites son más que suficientes.

## Troubleshooting

### Error: "GROQ_API_KEY not set"

**Solución:**
1. Verifica que el archivo `.env` existe en la raíz
2. Verifica que tiene: `GROQ_API_KEY=gsk_...`
3. Reinicia el RAG service

### Error: "Rate limit exceeded"

**Causa:** Demasiadas requests en poco tiempo

**Solución:**
- Espera 1 minuto
- O actualiza a plan de pago de Groq (opcional)

### Error: "Invalid API key"

**Solución:**
1. Verifica que copiaste la key completa
2. Genera una nueva key en https://console.groq.com
3. Actualiza `.env`

## Cambiar a Ollama (Local)

Si prefieres usar Ollama local:

1. Instala Ollama: https://ollama.ai/download
2. Descarga modelo: `ollama pull llama3.2:3b`
3. Edita `.env`:
   ```env
   LLM_PROVIDER=ollama
   OLLAMA_BASE_URL=http://localhost:11434
   ```
4. Reinicia RAG service

## Privacidad

### Con Groq:
- El texto del contrato se envía a los servidores de Groq
- Groq procesa el texto y devuelve el análisis
- Groq NO almacena tus contratos (según sus términos)
- Usa HTTPS (encriptado)

### Con Ollama:
- Todo se procesa localmente
- Nada sale de tu computadora
- 100% privado

**Recomendación:** 
- Para contratos públicos o de prueba: Groq (más rápido)
- Para contratos confidenciales: Ollama (más privado)

## Ventajas del Sistema Híbrido

El sistema soporta ambos proveedores sin cambiar código:

```python
# En llm_service.py
llm = LLMService(provider="groq")  # o "ollama"
```

Puedes cambiar entre Groq y Ollama solo editando `.env`.

## Próximos Pasos

1. ✅ Configura Groq API key
2. ✅ Inicia el sistema
3. ✅ Analiza tu primer contrato
4. 📊 Compara resultados entre modelo base y fine-tuned
5. 🎯 Ajusta según tus necesidades

## Recursos

- **Groq Console**: https://console.groq.com
- **Groq Docs**: https://console.groq.com/docs
- **Modelos disponibles**: https://console.groq.com/docs/models
- **Pricing**: https://wow.groq.com/pricing (gratis para empezar)

---

**¿Preguntas?** Revisa [TROUBLESHOOTING_FRONTEND.md](TROUBLESHOOTING_FRONTEND.md)
