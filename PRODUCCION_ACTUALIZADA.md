# ✅ Actualización de Producción - LexAnalyzer

## Cambios Realizados

### 1. Docker Compose de Producción

**Archivo:** `docker-compose.prod.yml`

✅ Agregado servicio `rag-service` con:
- Variables de entorno para Groq API
- Volumen persistente para ChromaDB
- Health checks
- Límites de recursos

✅ Renombrado contenedores:
- `finetune-*` → `lexanalyzer-*`

✅ Agregado volumen `chroma_data` para persistencia

### 2. Variables de Entorno

**Archivo:** `.env.example`

Nuevas variables requeridas:
```env
# LLM Provider
LLM_PROVIDER=groq
GROQ_API_KEY=your_key_here
OLLAMA_BASE_URL=http://localhost:11434

# Kaggle (para training)
KAGGLE_USERNAME=your_username
KAGGLE_KEY=your_key
```

### 3. Scripts de Producción

**Nuevo archivo:** `start-production.bat`

Script automatizado que:
- Verifica Docker
- Construye imágenes
- Inicia servicios
- Verifica salud

### 4. Dockerfiles

**backend/dockerfile** (desarrollo):
- ✅ Arreglado: Copia `go.mod` y `go.sum`
- ✅ Eliminado: `go mod tidy` problemático
- ✅ Agregado: Script de entrypoint para Kaggle

**backend/Dockerfile.prod** (producción):
- ✅ Ya estaba correcto (multi-stage build)
- ✅ Usa `go.mod` y `go.sum` correctamente

## Arquitectura de Producción

```
Internet
    ↓
[Frontend:80] ← nginx
    ↓
[Backend:8080] ← Go API
    ↓
[RAG Service:8001] ← Python/FastAPI
    ↓
[Groq API] ← LLM Cloud
    ↓
[PostgreSQL:5432] ← Database
[MinIO:9000] ← Object Storage
[ChromaDB] ← Vector Store
```

## Servicios en Producción

| Servicio | Puerto | Acceso | Descripción |
|----------|--------|--------|-------------|
| Frontend | 80 | Público | Nginx sirviendo HTML/JS |
| Backend | 8080 | Interno | API Go |
| RAG Service | 8001 | Interno | API Python |
| PostgreSQL | 5432 | Interno | Base de datos |
| MinIO | 9000 | Interno | Object storage |
| MinIO Console | 9001 | Público | Admin UI |

## Deployment

### Opción 1: Script Automático

```bash
.\start-production.bat
```

### Opción 2: Manual

```bash
# 1. Configurar .env
cp .env.example .env
# Editar .env con tus valores

# 2. Construir e iniciar
docker-compose -f docker-compose.prod.yml up --build -d

# 3. Verificar salud
docker-compose -f docker-compose.prod.yml ps
```

### Opción 3: Con CI/CD

```yaml
# .github/workflows/deploy.yml
name: Deploy to Production

on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      
      - name: Deploy to server
        env:
          GROQ_API_KEY: ${{ secrets.GROQ_API_KEY }}
          KAGGLE_USERNAME: ${{ secrets.KAGGLE_USERNAME }}
          KAGGLE_KEY: ${{ secrets.KAGGLE_KEY }}
        run: |
          docker-compose -f docker-compose.prod.yml up --build -d
```

## Configuración Requerida

### 1. Variables de Entorno

Crea `.env` con:

```env
# App
APP_ENV=production
APP_VERSION=1.0.0
PORT=8080

# Database
DB_USER=lexanalyzer
DB_PASSWORD=STRONG_PASSWORD_HERE
DB_NAME=lexanalyzer_db

# MinIO
MINIO_USER=admin
MINIO_PASSWORD=STRONG_PASSWORD_HERE
MINIO_BUCKET=lexanalyzer-models

# LLM
LLM_PROVIDER=groq
GROQ_API_KEY=gsk_your_key_here

# Kaggle (opcional)
KAGGLE_USERNAME=your_username
KAGGLE_KEY=your_key

# Security
ALLOWED_ORIGINS=https://yourdomain.com
RATE_LIMIT_REQUESTS_PER_MINUTE=100

# Monitoring
METRICS_ENABLED=true
LOG_LEVEL=info
LOG_FORMAT=json
```

### 2. Recursos Recomendados

**Mínimo:**
- CPU: 2 cores
- RAM: 4GB
- Disk: 20GB

**Recomendado:**
- CPU: 4 cores
- RAM: 8GB
- Disk: 50GB

**Óptimo:**
- CPU: 8 cores
- RAM: 16GB
- Disk: 100GB

### 3. Puertos a Abrir

En firewall/security groups:
- `80` - Frontend (HTTP)
- `443` - Frontend (HTTPS) - si usas SSL
- `9001` - MinIO Console (opcional, solo para admin)

## Health Checks

Todos los servicios tienen health checks automáticos:

```bash
# Ver estado
docker-compose -f docker-compose.prod.yml ps

# Debe mostrar "healthy" en todos
```

### Endpoints de Health

```bash
# Backend
curl http://localhost:8080/api/v1/health

# RAG Service
curl http://localhost:8001/health

# Frontend
curl http://localhost:80
```

## Monitoreo

### Ver Logs

```bash
# Todos los servicios
docker-compose -f docker-compose.prod.yml logs -f

# Servicio específico
docker-compose -f docker-compose.prod.yml logs -f backend
docker-compose -f docker-compose.prod.yml logs -f rag-service
```

### Métricas

El backend expone métricas en formato JSON:
```bash
curl http://localhost:8080/api/v1/metrics
```

### Recursos

```bash
# Ver uso de recursos
docker stats

# Ver por servicio
docker stats lexanalyzer-backend
docker stats lexanalyzer-rag
```

## Backup

### Base de Datos

```bash
# Backup
docker-compose -f docker-compose.prod.yml exec postgres \
  pg_dump -U lexanalyzer lexanalyzer_db > backup.sql

# Restore
docker-compose -f docker-compose.prod.yml exec -T postgres \
  psql -U lexanalyzer lexanalyzer_db < backup.sql
```

### MinIO (Modelos)

```bash
# Backup volumen
docker run --rm \
  -v lexanalyzer_minio_data:/data \
  -v $(pwd):/backup \
  alpine tar czf /backup/minio-backup.tar.gz /data

# Restore
docker run --rm \
  -v lexanalyzer_minio_data:/data \
  -v $(pwd):/backup \
  alpine tar xzf /backup/minio-backup.tar.gz -C /
```

### ChromaDB (Vectores)

```bash
# Backup
docker run --rm \
  -v lexanalyzer_chroma_data:/data \
  -v $(pwd):/backup \
  alpine tar czf /backup/chroma-backup.tar.gz /data
```

## Actualización

### Rolling Update

```bash
# 1. Pull nuevas imágenes
docker-compose -f docker-compose.prod.yml pull

# 2. Reconstruir servicios
docker-compose -f docker-compose.prod.yml build

# 3. Actualizar (sin downtime)
docker-compose -f docker-compose.prod.yml up -d --no-deps --build backend
docker-compose -f docker-compose.prod.yml up -d --no-deps --build rag-service
docker-compose -f docker-compose.prod.yml up -d --no-deps --build frontend
```

### Zero-Downtime Update

```bash
# 1. Escalar servicios
docker-compose -f docker-compose.prod.yml up -d --scale backend=2

# 2. Actualizar uno por uno
docker-compose -f docker-compose.prod.yml up -d --no-deps --build backend

# 3. Volver a escala normal
docker-compose -f docker-compose.prod.yml up -d --scale backend=1
```

## Troubleshooting

### Servicio no inicia

```bash
# Ver logs
docker-compose -f docker-compose.prod.yml logs servicio

# Reconstruir
docker-compose -f docker-compose.prod.yml build --no-cache servicio
docker-compose -f docker-compose.prod.yml up -d servicio
```

### Health check falla

```bash
# Verificar manualmente
docker-compose -f docker-compose.prod.yml exec backend wget -O- http://localhost:8080/api/v1/health

# Ver logs
docker-compose -f docker-compose.prod.yml logs backend
```

### Groq API no funciona

```bash
# Verificar variable
docker-compose -f docker-compose.prod.yml exec rag-service env | grep GROQ

# Debe mostrar: GROQ_API_KEY=gsk_...
```

## Seguridad

### 1. Cambiar Passwords

En `.env`:
```env
DB_PASSWORD=STRONG_RANDOM_PASSWORD
MINIO_PASSWORD=STRONG_RANDOM_PASSWORD
```

### 2. Configurar HTTPS

Usa un reverse proxy (nginx/traefik) con Let's Encrypt:

```yaml
# docker-compose.prod.yml
services:
  nginx-proxy:
    image: nginxproxy/nginx-proxy
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - /var/run/docker.sock:/tmp/docker.sock:ro
      - ./certs:/etc/nginx/certs
```

### 3. Rate Limiting

Ya configurado en `.env`:
```env
RATE_LIMIT_REQUESTS_PER_MINUTE=100
RATE_LIMIT_EXPENSIVE_ENDPOINTS=10
```

### 4. Firewall

```bash
# Permitir solo puertos necesarios
ufw allow 80/tcp
ufw allow 443/tcp
ufw enable
```

## Costos

### Groq API (LLM)
- Tier gratuito: 14,400 requests/día
- Suficiente para ~500 análisis/día
- Costo adicional: $0.10 por 1M tokens

### Infraestructura
- VPS básico: $5-10/mes (DigitalOcean, Linode)
- VPS recomendado: $20-40/mes
- Cloud (AWS/GCP): $50-100/mes

### Total Estimado
- Desarrollo/Testing: $0-10/mes
- Producción pequeña: $20-50/mes
- Producción mediana: $50-150/mes

## Soporte

Para problemas en producción:
1. Revisa logs: `docker-compose -f docker-compose.prod.yml logs`
2. Verifica health: `docker-compose -f docker-compose.prod.yml ps`
3. Consulta [DOCKER_TROUBLESHOOTING.md](DOCKER_TROUBLESHOOTING.md)
4. Abre issue en GitHub con logs

---

**Listo para producción** ✅
