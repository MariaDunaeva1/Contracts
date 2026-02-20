@echo off
echo ========================================
echo   Stopping LexAnalyzer
echo ========================================
echo.

docker-compose down

echo.
echo ✓ All services stopped
echo.
pause
