# 🚀 Inicio Rápido - LexAnalyzer

## Paso 1: Configurar API Key de Groq

1. Ve a https://console.groq.com/keys
2. Regístrate (gratis)
3. Crea una API key
4. Copia la key (empieza con `gsk_...`)

## Paso 2: Editar .env

Abre el archivo `.env` en la raíz del proyecto y verifica que tenga:

```env
LLM_PROVIDER=groq
GROQ_API_KEY=gsk_tu_key_aqui_pegala

KAGGLE_USERNAME=maradg
KAGGLE_KEY=f05dd2a1f3521ec27edb85f8d0e62a0d
```

## Paso 3: Iniciar Docker

### Opción A: PowerShell (Recomendado)

```powershell
.\start-lexanalyzer.bat
```

### Opción B: CMD

```cmd
start-lexanalyzer.bat
```

### Opción C: Manual

```bash
docker-compose up --build
```

## Paso 4: Esperar

Primera vez: 3-5 minutos (descarga imágenes y construye)
Siguientes veces: 30 segundos

Verás:
```
✓ Docker is installed
✓ Docker services started
✓ Services ready
```

## Paso 5: Abrir la Aplicación

Abre en tu navegador:
```
http://localhost:3000/contract-analysis.html
```

## Paso 6: Analizar un Contrato

1. Marca el checkbox "Use Fine-Tuned Model" (recomendado)
2. Pega el texto de un contrato
3. Click "Analyze Contract"
4. Espera 10-30 segundos
5. ¡Listo! Verás el análisis completo

## Comandos Útiles

### Ver logs
```powershell
.\logs.bat
# o
docker-compose logs -f
```

### Detener todo
```powershell
.\stop-lexanalyzer.bat
# o
docker-compose down
```

### Reiniciar un servicio
```bash
docker-compose restart backend
docker-compose restart rag-service
```

## Problemas Comunes

### "El término 'start-lexanalyzer.bat' no se reconoce"

**Solución:** Estás en PowerShell, usa:
```powershell
.\start-lexanalyzer.bat
```

### "Port already in use"

**Solución:**
```bash
docker-compose down
docker-compose up --build
```

### "GROQ_API_KEY not set"

**Solución:** Verifica que el archivo `.env` tiene tu API key de Groq

### "Backend build failed"

**Solución:**
```bash
# Limpiar y reconstruir
docker-compose down -v
docker-compose build --no-cache
docker-compose up
```

### Ver logs detallados
```bash
docker-compose logs backend
docker-compose logs rag-service
```

## Servicios y Puertos

Una vez iniciado, tendrás acceso a:

- **Frontend**: http://localhost:3000
- **Análisis de Contratos**: http://localhost:3000/contract-analysis.html
- **Backend API**: http://localhost:8080
- **RAG Service**: http://localhost:8001
- **MinIO Console**: http://localhost:9001 (minioadmin/minioadmin)
- **PostgreSQL**: localhost:5432

## Verificar que Todo Funciona

### 1. Backend
```bash
curl http://localhost:8080/api/v1/health
```

Debe responder:
```json
{"status":"ok","services":{"db":"up","storage":"up"}}
```

### 2. RAG Service
```bash
curl http://localhost:8001/health
```

Debe responder:
```json
{"status":"healthy","llm_service":"available",...}
```

### 3. Frontend
Abre http://localhost:3000 - debe cargar el dashboard

## Siguiente Paso

Lee [CONFIGURACION_GROQ.md](CONFIGURACION_GROQ.md) para entender cómo funciona el sistema.

## Ayuda

Si algo no funciona:
1. Lee [DOCKER_TROUBLESHOOTING.md](DOCKER_TROUBLESHOOTING.md)
2. Ejecuta `docker-compose logs` y busca errores
3. Abre un issue en GitHub con los logs

---

**¡Listo!** Ya puedes analizar contratos con IA 🎉
