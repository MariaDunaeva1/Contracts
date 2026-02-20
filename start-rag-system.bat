@echo off
echo ========================================
echo   LexAnalyzer - RAG System
echo ========================================
echo.

echo [1/4] Checking LLM Provider...
if "%LLM_PROVIDER%"=="ollama" (
    echo Using Ollama (local)...
    curl -s http://localhost:11434/api/tags >nul 2>&1
    if %errorlevel% neq 0 (
        echo ERROR: Ollama is not running!
        echo Please start Ollama first or switch to Groq in .env
        pause
        exit /b 1
    )
    echo ✓ Ollama is running
) else (
    echo Using Groq API (cloud)...
    if "%GROQ_API_KEY%"=="" (
        echo ERROR: GROQ_API_KEY not set in .env file!
        echo.
        echo To use Groq:
        echo 1. Get API key from https://console.groq.com/keys
        echo 2. Add to .env: GROQ_API_KEY=your_key_here
        echo.
        echo Or switch to Ollama: LLM_PROVIDER=ollama
        pause
        exit /b 1
    )
    echo ✓ Groq API key configured
)

echo.
echo [2/4] Installing Python dependencies...
cd backend\services
pip install -q -r requirements.txt
if %errorlevel% neq 0 (
    echo ERROR: Failed to install dependencies
    pause
    exit /b 1
)
echo ✓ Dependencies installed

echo.
echo [3/4] Starting RAG Service...
start "RAG Service" cmd /k "python rag_service.py"
timeout /t 5 /nobreak >nul

echo.
echo [4/4] Checking RAG Service health...
timeout /t 3 /nobreak >nul
curl -s http://localhost:8001/health
if %errorlevel% neq 0 (
    echo WARNING: RAG service may not be ready yet
    echo Wait a few more seconds and check http://localhost:8001/health
)

echo.
echo ========================================
echo   LexAnalyzer Started!
echo ========================================
echo.
echo Services:
echo   - RAG Service: http://localhost:8001
echo   - Health Check: http://localhost:8001/health
echo   - API Docs: http://localhost:8001/docs
echo.
echo LLM Provider: %LLM_PROVIDER%
echo.
echo Next steps:
echo   1. Start Go backend: cd backend ^&^& go run cmd/server/main.go
echo   2. Start frontend: cd frontend ^&^& python serve.py
echo   3. Open: http://localhost:3000/contract-analysis.html
echo.
echo Press any key to exit...
pause >nul
