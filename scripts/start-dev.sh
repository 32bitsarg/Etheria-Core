#!/bin/bash

# Script para iniciar el entorno de desarrollo del servidor MMO

echo "🚀 Iniciando entorno de desarrollo MMO..."

# Verificar si Docker está instalado
if ! command -v docker &> /dev/null; then
    echo "❌ Docker no está instalado. Por favor instala Docker primero."
    exit 1
fi

# Verificar si docker-compose está instalado
if ! command -v docker-compose &> /dev/null; then
    echo "❌ docker-compose no está instalado. Por favor instala docker-compose primero."
    exit 1
fi

# Detener contenedores existentes
echo "🛑 Deteniendo contenedores existentes..."
docker-compose -f deployments/docker-compose.yml down

# Iniciar PostgreSQL y Redis
echo "🐘 Iniciando PostgreSQL y Redis..."
docker-compose -f deployments/docker-compose.yml up -d postgres redis

# Esperar a que PostgreSQL esté listo
echo "⏳ Esperando a que PostgreSQL esté listo..."
sleep 10

# Verificar que PostgreSQL esté funcionando
echo "🔍 Verificando conexión a PostgreSQL..."
docker-compose -f deployments/docker-compose.yml exec postgres pg_isready -U mmo_user -d mmo_db

if [ $? -eq 0 ]; then
    echo "✅ PostgreSQL está funcionando correctamente"
else
    echo "❌ Error: PostgreSQL no está funcionando"
    exit 1
fi

echo "🎮 Entorno de desarrollo listo!"
echo "📊 PostgreSQL: localhost:5432"
echo "🔴 Redis: localhost:6379"
echo ""
echo "Para iniciar el servidor, ejecuta:"
echo "go run ."
echo ""
echo "Para ver los logs de PostgreSQL:"
echo "docker-compose -f deployments/docker-compose.yml logs postgres" 