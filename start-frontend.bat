@echo off
echo ========================================
echo Starting Finetune Studio Frontend
echo ========================================
echo.

REM Check if Python is available
python --version >nul 2>&1
if %errorlevel% neq 0 (
    echo ERROR: Python is not installed or not in PATH
    echo Please install Python 3.x
    pause
    exit /b 1
)

echo Starting frontend server on http://localhost:3000
echo Backend should be running on http://localhost:8080
echo.
echo Press Ctrl+C to stop
echo.

cd frontend
python serve.py

pause
