@echo off
echo ========================================
echo Instalador de Redis para Windows
echo ========================================

echo.
echo Para Windows, se recomienda usar WSL2 (Windows Subsystem for Linux)
echo o instalar Redis manualmente.
echo.

echo Opciones de instalación:
echo 1. Usar WSL2 (Recomendado)
echo 2. Instalación manual con Chocolatey
echo 3. Instalación manual desde archivos
echo 4. Usar Docker
echo.

set /p choice="Selecciona una opción (1-4): "

if "%choice%"=="1" goto wsl2
if "%choice%"=="2" goto chocolatey
if "%choice%"=="3" goto manual
if "%choice%"=="4" goto docker
goto invalid

:wsl2
echo.
echo ========================================
echo Instalando Redis con WSL2
echo ========================================
echo.

echo Verificando si WSL2 está instalado...
wsl --list --verbose

if %errorlevel% neq 0 (
    echo WSL2 no está instalado. Instalando...
    echo.
    echo Ejecutando: wsl --install
    wsl --install
    echo.
    echo Por favor, reinicia tu computadora y ejecuta este script nuevamente.
    pause
    exit /b 1
)

echo.
echo WSL2 está instalado. Instalando Redis en Ubuntu...
wsl sudo apt update
wsl sudo apt install -y redis-server
wsl sudo systemctl enable redis-server
wsl sudo systemctl start redis-server

echo.
echo Verificando instalación...
wsl redis-cli ping

if %errorlevel% equ 0 (
    echo ✅ Redis instalado correctamente en WSL2
    echo.
    echo Para usar Redis desde Windows:
    echo wsl redis-cli ping
    echo wsl redis-cli
) else (
    echo ❌ Error instalando Redis en WSL2
)

goto end

:chocolatey
echo.
echo ========================================
echo Instalando Redis con Chocolatey
echo ========================================
echo.

echo Verificando si Chocolatey está instalado...
choco --version

if %errorlevel% neq 0 (
    echo Chocolatey no está instalado. Instalando...
    echo.
    echo Ejecutando PowerShell como administrador...
    powershell -Command "Set-ExecutionPolicy Bypass -Scope Process -Force; [System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072; iex ((New-Object System.Net.WebClient).DownloadString('https://community.chocolatey.org/install.ps1'))"
    echo.
    echo Por favor, cierra y abre una nueva ventana de comandos como administrador.
    pause
    exit /b 1
)

echo.
echo Instalando Redis...
choco install redis-64 -y

echo.
echo Verificando instalación...
redis-cli ping

if %errorlevel% equ 0 (
    echo ✅ Redis instalado correctamente con Chocolatey
) else (
    echo ❌ Error instalando Redis con Chocolatey
)

goto end

:manual
echo.
echo ========================================
echo Instalación Manual de Redis
echo ========================================
echo.

echo Para instalar Redis manualmente en Windows:
echo.
echo 1. Descarga Redis para Windows desde:
echo    https://github.com/microsoftarchive/redis/releases
echo.
echo 2. Ejecuta el instalador como administrador
echo.
echo 3. Configura Redis como servicio:
echo    redis-server --service-install redis.windows.conf
echo.
echo 4. Inicia el servicio:
echo    redis-server --service-start
echo.
echo 5. Verifica la instalación:
echo    redis-cli ping
echo.

echo ¿Deseas abrir la página de descarga?
set /p open="S/N: "
if /i "%open%"=="S" (
    start https://github.com/microsoftarchive/redis/releases
)

goto end

:docker
echo.
echo ========================================
echo Instalando Redis con Docker
echo ========================================
echo.

echo Verificando si Docker está instalado...
docker --version

if %errorlevel% neq 0 (
    echo Docker no está instalado. Instalando Docker Desktop...
    echo.
    echo Descargando Docker Desktop...
    start https://www.docker.com/products/docker-desktop
    echo.
    echo Por favor, instala Docker Desktop y ejecuta este script nuevamente.
    pause
    exit /b 1
)

echo.
echo Ejecutando Redis con Docker...
docker run --name redis-mmo -p 6379:6379 -d redis:latest

echo.
echo Verificando instalación...
docker exec redis-mmo redis-cli ping

if %errorlevel% equ 0 (
    echo ✅ Redis ejecutándose en Docker
    echo.
    echo Comandos útiles:
    echo - Conectar: docker exec -it redis-mmo redis-cli
    echo - Detener: docker stop redis-mmo
    echo - Iniciar: docker start redis-mmo
    echo - Eliminar: docker rm redis-mmo
) else (
    echo ❌ Error ejecutando Redis en Docker
)

goto end

:invalid
echo Opción inválida. Por favor, selecciona 1, 2, 3 o 4.
pause
exit /b 1

:end
echo.
echo ========================================
echo Configuración del Servidor MMO
echo ========================================
echo.
echo Una vez que Redis esté instalado, asegúrate de que tu archivo
echo config.yaml tenga la configuración correcta:
echo.
echo redis:
echo   host: localhost
echo   port: 6379
echo   password: ""
echo   db: 0
echo.
echo Para probar la conexión desde tu servidor MMO:
echo curl http://localhost:8080/api/health
echo.
echo Para monitorear Redis:
echo redis-cli monitor
echo.
echo ¡Instalación completada!
pause 