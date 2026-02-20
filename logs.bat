@echo off
echo ========================================
echo   LexAnalyzer - Service Logs
echo ========================================
echo.
echo Press Ctrl+C to stop viewing logs
echo.

docker-compose logs -f --tail=50
