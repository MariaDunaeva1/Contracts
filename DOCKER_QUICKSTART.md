# 🐳 Docker Quick Start

## Inicio Rápido (1 comando)

```bash
docker-compose up --build
```

Espera a que todos los servicios inicien (1-2 minutos) y luego abre:

**Frontend:** http://localhost:3000  
**Backend API:** http://localhost:8080  
**MinIO Console:** http://localhost:9001 (minioadmin/minioadmin)

## Servicios

| Servicio | Puerto | URL |
|----------|--------|-----|
| Frontend | 3000 | http://localhost:3000 |
| Backend | 8080 | http://localhost:8080 |
| PostgreSQL | 5432 | localhost:5432 |
| MinIO | 9000 | http://localhost:9000 |
| MinIO Console | 9001 | http://localhost:9001 |

## Comandos Útiles

```bash
# Iniciar servicios
docker-compose up

# Iniciar en background
docker-compose up -d

# Rebuild y reiniciar
docker-compose up --build

# Ver logs
docker-compose logs -f

# Ver logs de un servicio específico
docker-compose logs -f backend
docker-compose logs -f frontend

# Detener servicios
docker-compose down

# Detener y eliminar volúmenes (limpieza completa)
docker-compose down -v

# Reiniciar un servicio
docker-compose restart backend
docker-compose restart frontend
```

## Verificación

### 1. Verificar que todos los servicios están corriendo

```bash
docker-compose ps
```

Deberías ver:
```
NAME                COMMAND                  SERVICE    STATUS
postgres            "docker-entrypoint..."   postgres   Up
minio               "/usr/bin/docker-ent..."  minio      Up
backend             "./server"               backend    Up
frontend            "nginx -g 'daemon of..."  frontend   Up
```

### 2. Verificar Backend

```bash
curl http://localhost:8080/api/v1/health
```

Respuesta esperada:
```json
{"status":"ok","services":{"db":"up","storage":"up"}}
```

### 3. Verificar Frontend

Abre en navegador: http://localhost:3000

Deberías ver el dashboard sin errores.

## Troubleshooting

### Problema: "Port already in use"

```bash
# Ver qué está usando el puerto
netstat -ano | findstr :3000
netstat -ano | findstr :8080

# Detener servicios existentes
docker-compose down

# O cambiar puertos en docker-compose.yml
```

### Problema: "Cannot connect to backend"

```bash
# Ver logs del backend
docker-compose logs backend

# Reiniciar backend
docker-compose restart backend

# Verificar que backend está corriendo
docker-compose ps backend
```

### Problema: "Database connection failed"

```bash
# Ver logs de postgres
docker-compose logs postgres

# Reiniciar postgres
docker-compose restart postgres

# Verificar variables de entorno en docker-compose.yml
```

### Problema: Frontend muestra "Failed to load dashboard"

1. Verifica que backend está corriendo:
```bash
curl http://localhost:8080/api/v1/health
```

2. Abre DevTools (F12) en el navegador y revisa Console

3. Verifica que Nginx está proxy-ing correctamente:
```bash
docker-compose logs frontend
```

## Desarrollo

### Editar código sin rebuild

Los volúmenes están montados, así que puedes editar:

**Backend:**
```bash
# Los cambios requieren rebuild
docker-compose up --build backend
```

**Frontend:**
```bash
# Los cambios se reflejan automáticamente
# Solo refresca el navegador (Ctrl+R)
```

### Acceder a contenedores

```bash
# Backend
docker-compose exec backend sh

# Frontend
docker-compose exec frontend sh

# PostgreSQL
docker-compose exec postgres psql -U finetune -d finetune_db
```

## Limpieza

### Eliminar todo

```bash
# Detener y eliminar contenedores, redes, volúmenes
docker-compose down -v

# Eliminar imágenes también
docker-compose down -v --rmi all
```

### Limpiar Docker completamente

```bash
# Eliminar contenedores detenidos
docker container prune

# Eliminar imágenes sin usar
docker image prune -a

# Eliminar volúmenes sin usar
docker volume prune

# Eliminar todo (¡cuidado!)
docker system prune -a --volumes
```

## Variables de Entorno

Crea un archivo `.env` en la raíz del proyecto:

```env
# Kaggle (opcional)
KAGGLE_USERNAME=your_username
KAGGLE_KEY=your_api_key

# Database
POSTGRES_USER=finetune
POSTGRES_PASSWORD=finetune_pass
POSTGRES_DB=finetune_db

# MinIO
MINIO_ROOT_USER=minioadmin
MINIO_ROOT_PASSWORD=minioadmin
```

## Producción

Para producción, modifica `docker-compose.yml`:

```yaml
services:
  backend:
    restart: always
    environment:
      - DATABASE_URL=postgres://user:pass@prod-db:5432/db
      - MINIO_ENDPOINT=prod-minio:9000
      - MINIO_USE_SSL=true
  
  frontend:
    restart: always
    # Agregar SSL/HTTPS
```

## Monitoreo

```bash
# Ver uso de recursos
docker stats

# Ver logs en tiempo real
docker-compose logs -f --tail=100

# Ver solo errores
docker-compose logs | grep -i error
```

## Backup

### Base de datos

```bash
# Backup
docker-compose exec postgres pg_dump -U finetune finetune_db > backup.sql

# Restore
docker-compose exec -T postgres psql -U finetune finetune_db < backup.sql
```

### MinIO (modelos y datasets)

```bash
# Backup volumen
docker run --rm -v contracts_minio_data:/data -v $(pwd):/backup alpine tar czf /backup/minio-backup.tar.gz /data

# Restore
docker run --rm -v contracts_minio_data:/data -v $(pwd):/backup alpine tar xzf /backup/minio-backup.tar.gz -C /
```

## ✅ Todo Listo

Una vez que `docker-compose up` esté corriendo sin errores:

1. Abre http://localhost:3000
2. Deberías ver el dashboard
3. Sube un dataset
4. Crea un training job
5. Monitorea el progreso
6. Descarga el modelo
7. Evalúa los resultados

¡Disfruta! 🚀
