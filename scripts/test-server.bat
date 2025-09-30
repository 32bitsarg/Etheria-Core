@echo off
echo ========================================
echo    PRUEBA DEL SERVIDOR MMO BACKEND
echo ========================================
echo.

echo [1/5] Verificando que el servidor compile...
go build -o server.exe main.go
if %errorlevel% neq 0 (
    echo ERROR: El servidor no compila correctamente
    pause
    exit /b 1
)
echo ✓ Servidor compilado correctamente
echo.

echo [2/5] Verificando dependencias...
go mod tidy
echo ✓ Dependencias verificadas
echo.

echo [3/5] Verificando archivo de configuración...
if not exist "config\config.yaml" (
    echo ADVERTENCIA: No se encontró config.yaml
    echo Creando archivo de configuración por defecto...
    copy "config\config.yaml.example" "config\config.yaml" >nul 2>&1
    if %errorlevel% neq 0 (
        echo ERROR: No se pudo crear el archivo de configuración
        pause
        exit /b 1
    )
    echo ✓ Archivo de configuración creado
) else (
    echo ✓ Archivo de configuración encontrado
)
echo.

echo [4/5] Verificando estructura de directorios...
if not exist "database" mkdir database
if not exist "database\migrations" mkdir database\migrations
if not exist "scripts" mkdir scripts
echo ✓ Estructura de directorios verificada
echo.

echo [5/5] Iniciando servidor de prueba...
echo.
echo El servidor se iniciará en http://localhost:8080
echo Presiona Ctrl+C para detener el servidor
echo.
echo ========================================
echo    SERVIDOR INICIADO
echo ========================================
echo.

timeout /t 3 /nobreak >nul

start "Servidor MMO" cmd /k "go run main.go"

echo.
echo Servidor iniciado en segundo plano
echo Para detener el servidor, cierra la ventana del servidor
echo.
echo Prueba los siguientes endpoints:
echo - Health check: http://localhost:8080/health
echo - Documentación: Revisa el README.md
echo.
pause 