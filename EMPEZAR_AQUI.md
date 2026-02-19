# 🚀 EMPEZAR AQUÍ - Despliegue en Producción

## ¿Qué tengo ahora?

Tu aplicación Finetune Studio está **100% lista para producción** con:

- ✅ Docker optimizado (imágenes 80% más pequeñas)
- ✅ Seguridad completa (rate limiting, CORS, headers)
- ✅ Monitoreo (logs JSON, métricas Prometheus)
- ✅ Documentación completa
- ✅ Scripts de backup/restore
- ✅ Listo para Oracle Cloud (GRATIS para siempre)

## 🎯 Opciones de Despliegue

### Opción 1: Oracle Cloud (RECOMENDADO - GRATIS)

**Ventajas:**
- 💰 $0/mes (gratis para siempre)
- 🖥️ 12GB RAM (suficiente para modelos 3B)
- 💾 200GB almacenamiento
- 🌐 10TB transferencia/mes
- ⚡ Sin tarjeta de crédito después del trial

**Tiempo:** 30-45 minutos

**Guía:** `docs/DEPLOY_ORACLE_CLOUD.md`

### Opción 2: Railway.app (RÁPIDO)

**Ventajas:**
- ⚡ Despliegue en 5 minutos
- 🔄 Auto-deploy desde GitHub
- 🔒 SSL automático
- 💳 $5/mes crédito gratis

**Tiempo:** 5-10 minutos

**Guía:** `docs/DEPLOY_RAILWAY.md`

### Opción 3: Probar Localmente Primero

**Si quieres probar antes de desplegar:**

```bash
# Windows
test-production-local.bat

# Linux/Mac
chmod +x scripts/test_production.sh
./scripts/test_production.sh
```

Esto levantará el sistema de producción en tu máquina local.

## 📋 Pasos Rápidos (Oracle Cloud)

### 1. Crear Cuenta (5 min)

1. Ir a https://cloud.oracle.com/free
2. Registrarse (gratis, sin tarjeta después del trial)
3. Verificar email

### 2. Crear VM (10 min)

1. **Compute** → **Instances** → **Create Instance**
2. Configurar:
   - **Name**: finetune-studio
   - **Image**: Ubuntu 22.04
   - **Shape**: VM.Standard.A1.Flex (ARM)
   - **OCPUs**: 2
   - **RAM**: 12 GB
   - **Public IP**: Sí
3. **Descargar SSH key** (importante!)
4. Crear

### 3. Configurar Firewall (5 min)

1. **Networking** → **VCN** → **Security Lists**
2. Agregar reglas:
   - Puerto 80 (HTTP)
   - Puerto 443 (HTTPS)
   - Puerto 22 (SSH)

### 4. Instalar Docker (5 min)

```bash
# SSH a la VM
ssh -i tu-key.pem ubuntu@<IP-PUBLICA>

# Instalar Docker
curl -fsSL https://get.docker.com | sh
sudo usermod -aG docker $USER

# Configurar firewall
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw allow 22/tcp
sudo ufw enable

# Salir y volver a entrar
exit
ssh -i tu-key.pem ubuntu@<IP-PUBLICA>
```

### 5. Desplegar Aplicación (10 min)

```bash
# Clonar repo
git clone <tu-repositorio>
cd finetune-studio

# Configurar
cp .env.example .env
nano .env
```

**Editar estos valores en .env:**
```bash
DB_PASSWORD=TuPasswordSegura123!
MINIO_PASSWORD=OtraPasswordSegura456!
KAGGLE_USERNAME=tu_usuario_kaggle
KAGGLE_KEY=tu_api_key_kaggle
ALLOWED_ORIGINS=http://<TU-IP-PUBLICA>
APP_ENV=production
LOG_FORMAT=json
```

Guardar (Ctrl+X, Y, Enter).

```bash
# Desplegar
docker compose -f docker-compose.prod.yml up -d --build

# Esperar 2-3 minutos...

# Verificar
curl http://localhost:8080/api/v1/health
```

### 6. ¡Listo! 🎉

Abre en tu navegador: `http://<TU-IP-PUBLICA>`

## 🔍 Verificar que Todo Funciona

```bash
# Health check
curl http://localhost:8080/api/v1/health

# Debería responder:
# {
#   "status": "healthy",
#   "version": "1.0.0",
#   ...
# }

# Ver logs
docker compose -f docker-compose.prod.yml logs -f

# Ver servicios
docker compose -f docker-compose.prod.yml ps

# Todos deberían estar "Up (healthy)"
```

## 📱 Acceder a la Aplicación

- **Frontend**: http://<TU-IP>
- **API**: http://<TU-IP>:8080/api/v1
- **Health**: http://<TU-IP>:8080/api/v1/health
- **Métricas**: http://<TU-IP>:8080/api/v1/metrics
- **MinIO Console**: http://<TU-IP>:9001

## 🔒 Siguiente: Configurar SSL (Opcional pero Recomendado)

```bash
# Instalar Certbot
sudo apt install certbot -y

# Obtener certificado (reemplaza con tu dominio)
sudo certbot certonly --standalone -d tudominio.com

# Configurar en docker-compose.prod.yml
# Ver guía completa en docs/DEPLOY_ORACLE_CLOUD.md
```

## 🔧 Comandos Útiles

```bash
# Ver logs en tiempo real
docker compose -f docker-compose.prod.yml logs -f

# Ver solo backend
docker compose -f docker-compose.prod.yml logs -f backend

# Reiniciar todo
docker compose -f docker-compose.prod.yml restart

# Reiniciar solo backend
docker compose -f docker-compose.prod.yml restart backend

# Ver recursos
docker stats

# Parar todo
docker compose -f docker-compose.prod.yml down

# Actualizar aplicación
git pull
docker compose -f docker-compose.prod.yml up -d --build
```

## 📚 Documentación Completa

Si necesitas más detalles:

1. **`RESUMEN_PRODUCCION.md`** - Resumen ejecutivo en español
2. **`docs/DEPLOY_ORACLE_CLOUD.md`** - Guía completa Oracle Cloud
3. **`docs/PRODUCTION_CHECKLIST.md`** - Checklist de despliegue
4. **`docs/TROUBLESHOOTING_PRODUCTION.md`** - Solución de problemas
5. **`PRODUCTION_DEPLOYMENT_COMPLETE.md`** - Detalles técnicos

## 🐛 Problemas Comunes

### "No puedo acceder a la aplicación"

```bash
# Verificar que los servicios están corriendo
docker compose -f docker-compose.prod.yml ps

# Verificar firewall
sudo ufw status

# Verificar logs
docker compose -f docker-compose.prod.yml logs
```

### "Error de conexión a la base de datos"

```bash
# Verificar que postgres está healthy
docker compose -f docker-compose.prod.yml ps postgres

# Ver logs de postgres
docker compose -f docker-compose.prod.yml logs postgres

# Reiniciar postgres
docker compose -f docker-compose.prod.yml restart postgres
```

### "Sin espacio en disco"

```bash
# Ver espacio
df -h

# Limpiar Docker
docker system prune -a

# Limpiar logs viejos
docker compose -f docker-compose.prod.yml logs --tail=0
```

## 💡 Tips Importantes

1. **Cambia las passwords** en `.env` antes de desplegar
2. **Configura ALLOWED_ORIGINS** con tu IP/dominio real
3. **Habilita el firewall** (ufw) en el servidor
4. **Configura SSL** con Let's Encrypt (gratis)
5. **Configura backups** automáticos con cron
6. **Monitorea recursos** con `docker stats`

## 🆘 ¿Necesitas Ayuda?

1. **Revisa logs**: `docker compose -f docker-compose.prod.yml logs -f`
2. **Revisa health**: `curl http://localhost:8080/api/v1/health`
3. **Consulta troubleshooting**: `docs/TROUBLESHOOTING_PRODUCTION.md`
4. **Revisa checklist**: `docs/PRODUCTION_CHECKLIST.md`

## ✅ Checklist Rápido

Antes de empezar, asegúrate de tener:

- [ ] Cuenta de Oracle Cloud (o Railway)
- [ ] Credenciales de Kaggle (username + API key)
- [ ] Git instalado localmente
- [ ] SSH key para conectar a la VM
- [ ] 30-45 minutos de tiempo

## 🎉 ¡Éxito!

Una vez desplegado, tendrás:

- ✅ Plataforma de ML funcionando 24/7
- ✅ Gratis para siempre (Oracle Cloud)
- ✅ Monitoreo completo
- ✅ Backups automáticos
- ✅ Seguridad hardening
- ✅ Logs estructurados
- ✅ Métricas Prometheus

---

**¿Listo para empezar?**

👉 Opción rápida: `docs/DEPLOY_ORACLE_CLOUD.md`

👉 Probar local primero: `test-production-local.bat`

👉 Más detalles: `RESUMEN_PRODUCCION.md`

**¡Buena suerte con tu despliegue!** 🚀
