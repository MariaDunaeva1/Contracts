# 🐳 Docker Troubleshooting - LexAnalyzer

## Problemas Comunes y Soluciones

### 1. Error: "Cannot connect to Docker daemon"

**Causa:** Docker Desktop no está corriendo

**Solución:**
```bash
# Windows: Abre Docker Desktop desde el menú inicio
# Espera a que el icono de Docker en la bandeja del sistema esté verde
```

### 2. Error: "Port already in use"

**Causa:** Otro servicio está usando los puertos 3000, 8080, 8001, 5432, 9000 o 9001

**Solución:**
```bash
# Ver qué está usando el puerto
netstat -ano | findstr :3000
netstat -ano | findstr :8080

# Detener servicios existentes
docker-compose down

# O cambiar puertos en docker-compose.yml
```

### 3. Error: "GROQ_API_KEY not set"

**Causa:** Variables de entorno no se están pasando al contenedor

**Solución:**
```bash
# 1. Verifica que .env existe en la raíz del proyecto
# 2. Verifica que tiene:
LLM_PROVIDER=groq
GROQ_API_KEY=gsk_tu_key_aqui

# 3. Reinicia Docker
docker-compose down
docker-compose up --build
```

### 4. Error: "Backend build failed"

**Causa:** Problema con el Dockerfile del backend

**Solución:**
```bash
# Ver logs completos
docker-compose build backend

# Si falla por kaggle.json, verifica backend/dockerfile
# Debe tener el script de entrypoint
```

### 5. Error: "RAG service not responding"

**Causa:** El servicio RAG no puede conectarse a Groq o falta alguna dependencia

**Solución:**
```bash
# Ver logs del servicio
docker-compose logs rag-service

# Reconstruir el servicio
docker-compose build rag-service
docker-compose up -d rag-service

# Verificar que GROQ_API_KEY está configurado
docker-compose exec rag-service env | grep GROQ
```

### 6. Error: "Database connection failed"

**Causa:** PostgreSQL no está listo cuando el backend intenta conectarse

**Solución:**
```bash
# Reiniciar servicios en orden
docker-compose down
docker-compose up -d postgres
timeout /t 10
docker-compose up -d backend frontend rag-service
```

### 7. Error: "MinIO not accessible"

**Causa:** MinIO no está corriendo o no está listo

**Solución:**
```bash
# Ver logs de MinIO
docker-compose logs minio

# Reiniciar MinIO
docker-compose restart minio

# Acceder a consola: http://localhost:9001
# Usuario: minioadmin
# Password: minioadmin
```

### 8. Frontend muestra "Failed to load"

**Causa:** Backend no está respondiendo o CORS no está configurado

**Solución:**
```bash
# 1. Verificar que backend está corriendo
curl http://localhost:8080/api/v1/health

# 2. Ver logs del backend
docker-compose logs backend

# 3. Verificar que frontend puede alcanzar backend
docker-compose exec frontend ping backend
```

### 9. Error: "Volume mount failed"

**Causa:** Permisos o ruta incorrecta en volúmenes

**Solución:**
```bash
# Eliminar volúmenes y recrear
docker-compose down -v
docker-compose up --build

# Verificar que las rutas existen
# - ./backend (debe existir)
# - ./chroma_db (se crea automáticamente)
```

### 10. Servicios muy lentos

**Causa:** Docker Desktop con pocos recursos

**Solución:**
```
1. Abre Docker Desktop
2. Settings → Resources
3. Aumenta:
   - CPUs: 4+
   - Memory: 8GB+
   - Swap: 2GB+
4. Apply & Restart
```

## Comandos Útiles

### Ver estado de servicios
```bash
docker-compose ps
```

### Ver logs de todos los servicios
```bash
docker-compose logs -f
```

### Ver logs de un servicio específico
```bash
docker-compose logs -f backend
docker-compose logs -f rag-service
docker-compose logs -f frontend
```

### Reiniciar un servicio
```bash
docker-compose restart backend
docker-compose restart rag-service
```

### Reconstruir un servicio
```bash
docker-compose build backend
docker-compose up -d backend
```

### Entrar a un contenedor
```bash
docker-compose exec backend sh
docker-compose exec rag-service bash
```

### Limpiar todo
```bash
# Detener y eliminar contenedores
docker-compose down

# Detener y eliminar contenedores + volúmenes
docker-compose down -v

# Eliminar imágenes también
docker-compose down -v --rmi all
```

### Ver uso de recursos
```bash
docker stats
```

## Verificación Paso a Paso

### 1. Verificar Docker
```bash
docker --version
docker-compose --version
```

### 2. Verificar .env
```bash
type .env
```

Debe contener:
```
LLM_PROVIDER=groq
GROQ_API_KEY=gsk_...
KAGGLE_USERNAME=...
KAGGLE_KEY=...
```

### 3. Construir servicios
```bash
docker-compose build
```

### 4. Iniciar servicios
```bash
docker-compose up -d
```

### 5. Verificar que están corriendo
```bash
docker-compose ps
```

Todos deben estar "Up":
- postgres
- minio
- backend
- frontend
- rag-service

### 6. Verificar salud de servicios
```bash
# Backend
curl http://localhost:8080/api/v1/health

# RAG Service
curl http://localhost:8001/health

# Frontend
curl http://localhost:3000
```

### 7. Abrir aplicación
```
http://localhost:3000/contract-analysis.html
```

## Logs de Ejemplo (Correcto)

### Backend
```
[INFO] Starting server on :8080
[INFO] Connected to database
[INFO] Connected to MinIO
[INFO] Server ready
```

### RAG Service
```
[LexAnalyzer] Initializing services...
[LexAnalyzer] Services initialized successfully (Provider: groq)
[LexAnalyzer] Starting server on http://0.0.0.0:8001
INFO:     Started server process
INFO:     Uvicorn running on http://0.0.0.0:8001
```

### Frontend
```
/docker-entrypoint.sh: Configuration complete; ready for start up
```

## Problemas Persistentes

Si nada funciona:

1. **Limpieza completa:**
```bash
docker-compose down -v --rmi all
docker system prune -a
```

2. **Reiniciar Docker Desktop**

3. **Reconstruir desde cero:**
```bash
docker-compose build --no-cache
docker-compose up
```

4. **Verificar logs detallados:**
```bash
docker-compose logs --tail=100
```

5. **Reportar issue con:**
   - Salida de `docker-compose ps`
   - Salida de `docker-compose logs`
   - Contenido de `.env` (sin API keys)
   - Sistema operativo y versión de Docker

## Contacto

Si el problema persiste, abre un issue en GitHub con:
- Descripción del problema
- Logs relevantes
- Pasos para reproducir
- Sistema operativo
- Versión de Docker

---

**Tip:** Usa `start-lexanalyzer.bat` para inicio automático con verificaciones.
