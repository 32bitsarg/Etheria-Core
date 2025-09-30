@echo off
echo ========================================
echo CONFIGURACION DE BASE DE DATOS MMO
echo ========================================
echo.

REM Verificar si PostgreSQL está instalado
where psql >nul 2>nul
if %errorlevel% neq 0 (
    echo ERROR: PostgreSQL no está instalado o no está en el PATH
    echo Por favor instala PostgreSQL desde: https://www.postgresql.org/download/
    pause
    exit /b 1
)

echo PostgreSQL encontrado. Configurando base de datos...
echo.

REM Solicitar información de conexión
set /p DB_HOST=Host de la base de datos (localhost): 
if "%DB_HOST%"=="" set DB_HOST=localhost

set /p DB_PORT=Puerto (5432): 
if "%DB_PORT%"=="" set DB_PORT=5432

set /p DB_USER=Usuario (postgres): 
if "%DB_USER%"=="" set DB_USER=postgres

set /p DB_PASSWORD=Contraseña: 
if "%DB_PASSWORD%"=="" (
    echo ERROR: La contraseña es obligatoria
    pause
    exit /b 1
)

set /p DB_NAME=Nombre de la base de datos (mmo_db): 
if "%DB_NAME%"=="" set DB_NAME=mmo_db

echo.
echo Configurando base de datos: %DB_NAME%
echo Host: %DB_HOST%:%DB_PORT%
echo Usuario: %DB_USER%
echo.

REM Crear base de datos si no existe
echo Creando base de datos...
psql -h %DB_HOST% -p %DB_PORT% -U %DB_USER% -c "CREATE DATABASE %DB_NAME%;" 2>nul
if %errorlevel% neq 0 (
    echo La base de datos ya existe o hubo un error. Continuando...
)

REM Ejecutar script consolidado
echo Ejecutando script de configuración...
psql -h %DB_HOST% -p %DB_PORT% -U %DB_USER% -d %DB_NAME% -f "..\database\createdb.sql"

if %errorlevel% equ 0 (
    echo.
    echo ========================================
    echo BASE DE DATOS CONFIGURADA EXITOSAMENTE!
    echo ========================================
    echo.
    echo La base de datos MMO está lista para usar.
    echo Puedes iniciar el servidor con: go run main.go
    echo.
) else (
    echo.
    echo ========================================
    echo ERROR AL CONFIGURAR LA BASE DE DATOS
    echo ========================================
    echo.
    echo Revisa los errores arriba y verifica:
    echo - Que PostgreSQL esté ejecutándose
    echo - Que las credenciales sean correctas
    echo - Que tengas permisos para crear bases de datos
    echo.
)

pause
