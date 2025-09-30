#!/bin/bash

echo "========================================"
echo "CONFIGURACION DE BASE DE DATOS MMO"
echo "========================================"
echo

# Verificar si PostgreSQL está instalado
if ! command -v psql &> /dev/null; then
    echo "ERROR: PostgreSQL no está instalado"
    echo "Por favor instala PostgreSQL desde: https://www.postgresql.org/download/"
    exit 1
fi

echo "PostgreSQL encontrado. Configurando base de datos..."
echo

# Solicitar información de conexión
read -p "Host de la base de datos (localhost): " DB_HOST
DB_HOST=${DB_HOST:-localhost}

read -p "Puerto (5432): " DB_PORT
DB_PORT=${DB_PORT:-5432}

read -p "Usuario (postgres): " DB_USER
DB_USER=${DB_USER:-postgres}

read -s -p "Contraseña: " DB_PASSWORD
echo
if [ -z "$DB_PASSWORD" ]; then
    echo "ERROR: La contraseña es obligatoria"
    exit 1
fi

read -p "Nombre de la base de datos (mmo_db): " DB_NAME
DB_HOST=${DB_NAME:-mmo_db}

echo
echo "Configurando base de datos: $DB_NAME"
echo "Host: $DB_HOST:$DB_PORT"
echo "Usuario: $DB_USER"
echo

# Crear base de datos si no existe
echo "Creando base de datos..."
PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -c "CREATE DATABASE $DB_NAME;" 2>/dev/null
if [ $? -ne 0 ]; then
    echo "La base de datos ya existe o hubo un error. Continuando..."
fi

# Ejecutar script consolidado
echo "Ejecutando script de configuración..."
PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f "../database/createdb.sql"

if [ $? -eq 0 ]; then
    echo
    echo "========================================"
    echo "BASE DE DATOS CONFIGURADA EXITOSAMENTE!"
    echo "========================================"
    echo
    echo "La base de datos MMO está lista para usar."
    echo "Puedes iniciar el servidor con: go run main.go"
    echo
else
    echo
    echo "========================================"
    echo "ERROR AL CONFIGURAR LA BASE DE DATOS"
    echo "========================================"
    echo
    echo "Revisa los errores arriba y verifica:"
    echo "- Que PostgreSQL esté ejecutándose"
    echo "- Que las credenciales sean correctas"
    echo "- Que tengas permisos para crear bases de datos"
    echo
fi
