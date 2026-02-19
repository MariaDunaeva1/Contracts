@echo off
echo ========================================
echo Testing Production Setup Locally
echo ========================================
echo.

echo [1/6] Checking if .env exists...
if exist .env (
    echo ✓ .env file found
) else (
    echo ✗ .env file not found
    echo Creating from template...
    copy .env.example .env
    echo.
    echo ⚠️  IMPORTANT: Edit .env file with your credentials before continuing!
    echo Press any key when ready...
    pause >nul
)
echo.

echo [2/6] Stopping any running containers...
docker-compose down
echo.

echo [3/6] Building production images...
docker-compose -f docker-compose.prod.yml build
if %errorlevel% neq 0 (
    echo ✗ Build failed
    exit /b 1
)
echo ✓ Build successful
echo.

echo [4/6] Starting services...
docker-compose -f docker-compose.prod.yml up -d
if %errorlevel% neq 0 (
    echo ✗ Failed to start services
    exit /b 1
)
echo ✓ Services started
echo.

echo [5/6] Waiting for services to be healthy (60s)...
timeout /t 60 /nobreak >nul
echo.

echo [6/6] Running health checks...
echo.

echo Testing backend health...
curl -s http://localhost:8080/api/v1/health
echo.
echo.

echo Testing frontend...
curl -s -o nul -w "Frontend Status: %%{http_code}\n" http://localhost:3000
echo.

echo Testing metrics endpoint...
curl -s -o nul -w "Metrics Status: %%{http_code}\n" http://localhost:8080/api/v1/metrics
echo.

echo ========================================
echo Production Test Complete!
echo ========================================
echo.
echo Services running:
docker-compose -f docker-compose.prod.yml ps
echo.
echo Access the application:
echo   Frontend: http://localhost:3000
echo   Backend:  http://localhost:8080
echo   Health:   http://localhost:8080/api/v1/health
echo   Metrics:  http://localhost:8080/api/v1/metrics
echo   MinIO:    http://localhost:9001
echo.
echo View logs:
echo   docker-compose -f docker-compose.prod.yml logs -f
echo.
echo Stop services:
echo   docker-compose -f docker-compose.prod.yml down
echo.
