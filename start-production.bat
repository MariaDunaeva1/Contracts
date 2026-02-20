@echo off
echo ========================================
echo   LexAnalyzer - Production Deployment
echo ========================================
echo.

REM Check if .env exists
if not exist .env (
    echo ERROR: .env file not found!
    echo.
    echo Please create .env file from .env.example
    echo and configure all required variables.
    pause
    exit /b 1
)

echo [1/4] Checking Docker...
docker --version >nul 2>&1
if %errorlevel% neq 0 (
    echo ERROR: Docker is not installed or not running!
    pause
    exit /b 1
)
echo ✓ Docker is installed

echo.
echo [2/4] Pulling latest images...
docker-compose -f docker-compose.prod.yml pull

echo.
echo [3/4] Building and starting services...
echo This may take several minutes...
docker-compose -f docker-compose.prod.yml up --build -d

if %errorlevel% neq 0 (
    echo ERROR: Failed to start services
    echo.
    echo Check logs with:
    echo   docker-compose -f docker-compose.prod.yml logs
    pause
    exit /b 1
)

echo ✓ Services started

echo.
echo [4/4] Waiting for services to be healthy...
timeout /t 30 /nobreak >nul

echo.
echo ========================================
echo   LexAnalyzer Production is Running!
echo ========================================
echo.
echo Services:
echo   Frontend:  http://localhost (port 80)
echo   Backend:   Internal only
echo   RAG API:   Internal only
echo   MinIO:     http://localhost:9001
echo.
echo Main App:
echo   http://localhost/contract-analysis.html
echo.
echo To view logs:
echo   docker-compose -f docker-compose.prod.yml logs -f
echo.
echo To stop:
echo   docker-compose -f docker-compose.prod.yml down
echo.
echo Health checks:
echo   docker-compose -f docker-compose.prod.yml ps
echo.
pause
