#!/bin/bash

echo "========================================"
echo "Instalador de Redis para el Servidor MMO"
echo "========================================"

# Detectar sistema operativo
if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    # Linux
    echo "Detectado: Linux"
    
    # Detectar distribución
    if [ -f /etc/debian_version ]; then
        echo "Distribución: Debian/Ubuntu"
        echo "Instalando Redis..."
        sudo apt update
        sudo apt install -y redis-server
        
        # Configurar Redis para iniciar automáticamente
        sudo systemctl enable redis-server
        sudo systemctl start redis-server
        
    elif [ -f /etc/redhat-release ]; then
        echo "Distribución: Red Hat/CentOS/Fedora"
        echo "Instalando Redis..."
        sudo yum install -y redis
        # O para versiones más nuevas:
        # sudo dnf install -y redis
        
        # Configurar Redis para iniciar automáticamente
        sudo systemctl enable redis
        sudo systemctl start redis
        
    else
        echo "Distribución no reconocida. Instalando desde fuente..."
        # Instalación desde fuente
        cd /tmp
        wget http://download.redis.io/redis-stable.tar.gz
        tar xvzf redis-stable.tar.gz
        cd redis-stable
        make
        sudo make install
        
        # Crear directorio de configuración
        sudo mkdir -p /etc/redis
        sudo cp redis.conf /etc/redis/
        
        # Crear servicio systemd
        sudo tee /etc/systemd/system/redis.service > /dev/null <<EOF
[Unit]
Description=Redis In-Memory Data Store
After=network.target

[Service]
Type=forking
ExecStart=/usr/local/bin/redis-server /etc/redis/redis.conf
ExecStop=/usr/local/bin/redis-cli shutdown
Restart=always

[Install]
WantedBy=multi-user.target
EOF
        
        sudo systemctl daemon-reload
        sudo systemctl enable redis
        sudo systemctl start redis
    fi
    
elif [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS
    echo "Detectado: macOS"
    echo "Instalando Redis con Homebrew..."
    
    # Verificar si Homebrew está instalado
    if ! command -v brew &> /dev/null; then
        echo "Homebrew no está instalado. Instalando..."
        /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
    fi
    
    brew install redis
    
    # Iniciar Redis
    brew services start redis
    
elif [[ "$OSTYPE" == "msys" ]] || [[ "$OSTYPE" == "cygwin" ]]; then
    # Windows con Git Bash o similar
    echo "Detectado: Windows (Git Bash/Cygwin)"
    echo "Para Windows, se recomienda usar WSL2 o instalar Redis manualmente."
    echo "Visita: https://redis.io/docs/getting-started/installation/install-redis-on-windows/"
    
else
    echo "Sistema operativo no reconocido: $OSTYPE"
    echo "Por favor, instala Redis manualmente desde: https://redis.io/download"
    exit 1
fi

# Verificar instalación
echo ""
echo "Verificando instalación de Redis..."

if command -v redis-server &> /dev/null; then
    echo "✅ Redis instalado correctamente"
    
    # Verificar si Redis está ejecutándose
    if redis-cli ping &> /dev/null; then
        echo "✅ Redis está ejecutándose"
        echo "✅ Conexión exitosa: $(redis-cli ping)"
    else
        echo "⚠️  Redis no está ejecutándose. Iniciando..."
        if [[ "$OSTYPE" == "darwin"* ]]; then
            brew services start redis
        else
            sudo systemctl start redis-server 2>/dev/null || sudo systemctl start redis 2>/dev/null
        fi
        
        # Esperar un momento y verificar nuevamente
        sleep 2
        if redis-cli ping &> /dev/null; then
            echo "✅ Redis iniciado correctamente"
        else
            echo "❌ Error iniciando Redis"
        fi
    fi
    
    # Mostrar información de Redis
    echo ""
    echo "Información de Redis:"
    echo "Versión: $(redis-server --version)"
    echo "Puerto: 6379"
    echo "Comando de conexión: redis-cli"
    
    # Configuración recomendada para el servidor MMO
    echo ""
    echo "Configuración recomendada para el servidor MMO:"
    echo "1. Asegúrate de que Redis esté configurado para iniciar automáticamente"
    echo "2. Configura un límite de memoria apropiado en redis.conf"
    echo "3. Habilita la persistencia si es necesario"
    echo ""
    echo "Para configurar Redis:"
    echo "- Archivo de configuración: /etc/redis/redis.conf (Linux) o /usr/local/etc/redis.conf (macOS)"
    echo "- Comando para editar: sudo nano /etc/redis/redis.conf"
    
else
    echo "❌ Error: Redis no se instaló correctamente"
    exit 1
fi

echo ""
echo "========================================"
echo "¡Instalación completada!"
echo "========================================"
echo ""
echo "Próximos pasos:"
echo "1. Asegúrate de que Redis esté ejecutándose"
echo "2. Configura tu archivo config.yaml con los parámetros de Redis"
echo "3. Ejecuta tu servidor MMO"
echo ""
echo "Para probar la conexión:"
echo "redis-cli ping"
echo ""
echo "Para monitorear Redis:"
echo "redis-cli monitor" 