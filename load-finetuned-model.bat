@echo off
echo ========================================
echo Cargando modelo fine-tuneado en Ollama
echo ========================================
echo.

echo Verificando que Ollama este corriendo...
curl -s http://localhost:11434/api/tags >nul 2>&1
if errorlevel 1 (
    echo ERROR: Ollama no esta corriendo
    echo Por favor inicia Ollama primero
    pause
    exit /b 1
)

echo Ollama esta corriendo correctamente
echo.

echo Creando Modelfile para el modelo fine-tuneado...
(
echo FROM llama3.2:3b
echo.
echo # Cargar adaptadores LoRA
echo ADAPTER ./models/lora_model/adapter_model.safetensors
echo.
echo # Configuracion del modelo
echo PARAMETER temperature 0.7
echo PARAMETER top_p 0.9
echo PARAMETER top_k 40
echo.
echo # System prompt para analisis legal
echo SYSTEM You are a legal contract analysis expert. Analyze contracts carefully and provide detailed, accurate information about clauses, risks, and obligations.
) > Modelfile.legal

echo.
echo Creando modelo en Ollama...
ollama create legal-contract-analyzer -f Modelfile.legal

if errorlevel 1 (
    echo.
    echo ERROR: No se pudo crear el modelo
    echo Verifica que los archivos del modelo esten en models/lora_model/
    pause
    exit /b 1
)

echo.
echo ========================================
echo Modelo cargado exitosamente!
echo ========================================
echo.
echo Nombre del modelo: legal-contract-analyzer
echo.
echo Para probarlo:
echo   ollama run legal-contract-analyzer "Analiza esta clausula..."
echo.
echo Limpiando archivo temporal...
del Modelfile.legal

echo.
echo Modelos disponibles en Ollama:
ollama list

pause
