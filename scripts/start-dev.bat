@echo off
REM Script para iniciar el entorno de desarrollo del servidor MMO en Windows

echo 🚀 Iniciando entorno de desarrollo MMO...

REM Verificar si Docker está instalado
docker --version >nul 2>&1
if %errorlevel% neq 0 (
    echo ❌ Docker no está instalado. Por favor instala Docker Desktop primero.
    pause
    exit /b 1
)

REM Verificar si docker-compose está instalado
docker-compose --version >nul 2>&1
if %errorlevel% neq 0 (
    echo ❌ docker-compose no está instalado. Por favor instala docker-compose primero.
    pause
    exit /b 1
)

REM Detener contenedores existentes
echo 🛑 Deteniendo contenedores existentes...
docker-compose -f deployments/docker-compose.yml down

REM Iniciar PostgreSQL y Redis
echo 🐘 Iniciando PostgreSQL y Redis...
docker-compose -f deployments/docker-compose.yml up -d postgres redis

REM Esperar a que PostgreSQL esté listo
echo ⏳ Esperando a que PostgreSQL esté listo...
timeout /t 10 /nobreak >nul

REM Verificar que PostgreSQL esté funcionando
echo 🔍 Verificando conexión a PostgreSQL...
docker-compose -f deployments/docker-compose.yml exec postgres pg_isready -U mmo_user -d mmo_db

if %errorlevel% equ 0 (
    echo ✅ PostgreSQL está funcionando correctamente
) else (
    echo ❌ Error: PostgreSQL no está funcionando
    pause
    exit /b 1
)

echo.
echo 🎮 Entorno de desarrollo listo!
echo 📊 PostgreSQL: localhost:5432
echo 🔴 Redis: localhost:6379
echo.
echo Para iniciar el servidor, ejecuta:
echo go run .
echo.
echo Para ver los logs de PostgreSQL:
echo docker-compose -f deployments/docker-compose.yml logs postgres
echo.
pause 