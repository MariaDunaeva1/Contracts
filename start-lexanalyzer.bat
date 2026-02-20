@echo off
echo ========================================
echo   LexAnalyzer - Complete Setup
echo ========================================
echo.

REM Check if .env exists
if not exist .env (
    echo ERROR: .env file not found!
    echo.
    echo Please create .env file with:
    echo   LLM_PROVIDER=groq
    echo   GROQ_API_KEY=your_key_here
    echo   KAGGLE_USERNAME=your_username
    echo   KAGGLE_KEY=your_key
    echo.
    pause
    exit /b 1
)

echo [1/3] Checking Docker...
docker --version >nul 2>&1
if %errorlevel% neq 0 (
    echo ERROR: Docker is not installed or not running!
    echo Please install Docker Desktop from https://www.docker.com/products/docker-desktop
    pause
    exit /b 1
)
echo ✓ Docker is installed

echo.
echo [2/3] Starting Docker services...
echo This may take a few minutes on first run...
docker-compose up --build -d

if %errorlevel% neq 0 (
    echo ERROR: Failed to start Docker services
    echo.
    echo Try:
    echo   docker-compose down
    echo   docker-compose up --build
    pause
    exit /b 1
)

echo ✓ Docker services started

echo.
echo [3/3] Waiting for services to be ready...
timeout /t 10 /nobreak >nul

echo.
echo ========================================
echo   LexAnalyzer is Ready!
echo ========================================
echo.
echo Services:
echo   Frontend:  http://localhost:3000
echo   Backend:   http://localhost:8080
echo   RAG API:   http://localhost:8001
echo   MinIO:     http://localhost:9001
echo.
echo Main App:
echo   http://localhost:3000/contract-analysis.html
echo.
echo To view logs:
echo   docker-compose logs -f
echo.
echo To stop:
echo   docker-compose down
echo.
pause
