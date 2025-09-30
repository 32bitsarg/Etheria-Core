# 🏰 Etheria MMO Server - Backend Completo

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go)](https://golang.org/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-14+-316192?style=for-the-badge&logo=postgresql)](https://postgresql.org/)
[![Redis](https://img.shields.io/badge/Redis-7+-DC382D?style=for-the-badge&logo=redis)](https://redis.io/)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=for-the-badge&logo=docker)](https://docker.com/)
[![License](https://img.shields.io/badge/License-MIT-green?style=for-the-badge)](LICENSE)

> **Servidor backend completo para un MMO de estrategia medieval tipo Clash of Clans**  
> Desarrollado por [@32bitsarg](https://github.com/32bitsarg)

Un servidor backend robusto y escalable desarrollado en Go para crear juegos MMO de estrategia medieval. Incluye todas las funcionalidades necesarias para un juego completo: autenticación, gestión de mundos, aldeas, recursos, edificios, unidades militares, chat en tiempo real, comercio, alianzas, combate, héroes, misiones, logros y mucho más.

## 🎮 Características Principales

### 🛡️ Sistema de Autenticación
- **JWT con bcrypt** - Autenticación segura y escalable
- **Gestión de sesiones** - Sesiones persistentes con Redis
- **Middleware de seguridad** - Validación automática de tokens
- **Perfiles de usuario** - Gestión completa de perfiles

### 🌍 Sistema de Mundos
- **Múltiples mundos** - Soporte para varios servidores/mundos
- **Configuración dinámica** - Cada mundo con sus propias reglas
- **Sistema de coordenadas** - Mapa hexagonal para aldeas
- **Exploración** - Sistema de exploración del mundo

### 🏘️ Sistema de Aldeas
- **Construcción avanzada** - 8 tipos de edificios mejorables
- **Recursos dinámicos** - Madera, piedra, comida y oro con generación automática
- **Producción en tiempo real** - Cálculo automático de recursos
- **Defensas** - Sistema de defensas y murallas

### ⚔️ Sistema de Combate
- **Batallas PvP** - Combate entre jugadores
- **6 tipos de unidades** - Cada una con estadísticas únicas
- **Sistema de héroes** - Unidades especiales con habilidades
- **Reportes detallados** - Análisis completo de batallas

### 💬 Chat en Tiempo Real
- **WebSockets** - Comunicación instantánea
- **Canales múltiples** - Global, alianza, privado
- **Sistema de moderación** - Herramientas de administración
- **Historial persistente** - Mensajes guardados en base de datos

### 🤝 Sistema de Alianzas
- **Gestión completa** - Crear, unirse, administrar alianzas
- **Roles y permisos** - Líder, oficial, miembro
- **Contribuciones** - Sistema de contribuciones y beneficios
- **Rankings** - Clasificaciones de alianzas

### 🎯 Sistema de Misiones
- **Misiones diarias/semanales** - Objetivos regulares
- **Sistema de progreso** - Seguimiento automático
- **Recompensas** - Recursos y experiencia
- **Categorías** - Historia, batalla, construcción, etc.

### 🏆 Sistema de Logros
- **Logros automáticos** - Desbloqueo por acciones
- **Categorías múltiples** - Construcción, batalla, social, etc.
- **Recompensas** - Recursos, títulos, experiencia
- **Progreso visual** - Seguimiento de avance

### 🔬 Sistema de Investigación
- **Tecnologías mejorables** - Investigación de mejoras
- **Efectos en juego** - Mejoras reales en aldeas
- **Árbol tecnológico** - Dependencias entre tecnologías
- **Costos escalables** - Sistema de costos progresivos

### 💰 Sistema Económico
- **Comercio entre jugadores** - Mercado de recursos
- **Ofertas y demandas** - Sistema de ofertas
- **Transacciones seguras** - Validación automática
- **Historial de comercio** - Registro de todas las transacciones

### 📊 Rankings y Estadísticas
- **Rankings en tiempo real** - Jugadores, alianzas, aldeas
- **Múltiples categorías** - Poder, riqueza, batallas, etc.
- **Estadísticas detalladas** - Análisis completo de progreso
- **Historial de rankings** - Evolución temporal

## 🚀 Inicio Rápido

### Prerrequisitos
- **Go 1.21+**
- **PostgreSQL 14+**
- **Redis** (opcional, para sesiones)
- **Docker** (opcional)

### Instalación

1. **Clonar el repositorio**
```bash
git clone https://github.com/32bitsarg/ethernia-mmo-server.git
cd ethernia-mmo-server
```

2. **Configurar la base de datos**
```bash
# Crear base de datos
createdb etheria_db

# Ejecutar migraciones
psql -d etheria_db -f database/createdb.sql
```

3. **Configurar variables de entorno**
```bash
# Copiar configuración
cp config/config.yaml.example config/config.yaml

# Editar configuración
nano config/config.yaml
```

4. **Instalar dependencias**
```bash
go mod download
```

5. **Ejecutar el servidor**
```bash
go run .
```

El servidor estará disponible en `http://localhost:8080`

## 🐳 Docker (Recomendado)

### Desarrollo con Docker Compose
```bash
# Iniciar todos los servicios
docker-compose -f deployments/docker-compose.yml up -d

# Ver logs
docker-compose -f deployments/docker-compose.yml logs -f

# Detener servicios
docker-compose -f deployments/docker-compose.yml down
```

### Servicios incluidos
- **PostgreSQL** - Base de datos principal
- **Redis** - Cache y sesiones
- **API Server** - Servidor backend

## 📡 API Endpoints

### 🔐 Autenticación
```http
POST /auth/register     # Registro de usuario
POST /auth/login        # Inicio de sesión
GET  /auth/profile      # Perfil del usuario
PUT  /auth/profile      # Actualizar perfil
```

### 🌍 Mundos
```http
GET  /api/worlds                    # Listar mundos disponibles
GET  /api/worlds/{id}               # Información del mundo
POST /api/worlds/{id}/join          # Unirse a un mundo
GET  /api/worlds/{id}/villages      # Aldeas del mundo
```

### 🏘️ Aldeas
```http
GET  /api/villages                  # Aldeas del jugador
GET  /api/villages/{id}             # Información de aldea
POST /api/villages/{id}/buildings/upgrade  # Mejorar edificio
GET  /api/villages/{id}/resources   # Recursos de aldea
GET  /api/villages/{id}/production  # Producción de aldea
```

### ⚔️ Batallas
```http
GET  /api/battles                   # Historial de batallas
POST /api/battles/attack            # Atacar aldea
GET  /api/battles/{id}              # Detalles de batalla
GET  /api/battles/reports           # Reportes de batalla
```

### 💬 Chat
```http
GET  /api/chat/channels             # Canales disponibles
POST /api/chat/channels             # Crear canal
GET  /api/chat/channels/{id}/messages  # Mensajes del canal
POST /api/chat/channels/{id}/messages  # Enviar mensaje
WS   /ws                            # WebSocket para chat
```

### 🤝 Alianzas
```http
GET  /api/alliances                 # Listar alianzas
POST /api/alliances                 # Crear alianza
GET  /api/alliances/my              # Mi alianza
GET  /api/alliances/{id}            # Información de alianza
POST /api/alliances/{id}/join       # Unirse a alianza
POST /api/alliances/{id}/leave      # Salir de alianza
```

### 🎯 Misiones
```http
GET  /api/quests                    # Misiones disponibles
GET  /api/quests/active             # Misiones activas
POST /api/quests/{id}/accept        # Aceptar misión
POST /api/quests/{id}/complete      # Completar misión
```

### 🦸 Héroes
```http
GET  /api/player/heroes            # Héroes del jugador
GET  /api/player/heroes/active     # Héroes activos
POST /api/player/heroes/{id}/upgrade  # Mejorar héroe
GET  /api/player/heroes/{id}       # Información de héroe
```

### 🔬 Investigación
```http
GET  /api/research                 # Tecnologías del jugador
POST /api/research/{id}/research    # Investigar tecnología
GET  /api/research/available       # Tecnologías disponibles
```

### 🏆 Logros
```http
GET  /api/achievements             # Logros del jugador
GET  /api/achievements/available   # Logros disponibles
POST /api/achievements/{id}/claim   # Reclamar logro
```

### 📊 Rankings
```http
GET  /api/rankings/players         # Ranking de jugadores
GET  /api/rankings/alliances       # Ranking de alianzas
GET  /api/rankings/villages        # Ranking de aldeas
```

### 💰 Economía
```http
GET  /api/economy/market           # Mercado de recursos
POST /api/economy/offers           # Crear oferta
GET  /api/economy/offers           # Ofertas disponibles
POST /api/economy/offers/{id}/buy  # Comprar oferta
```

### 🎪 Eventos
```http
GET  /api/events                   # Eventos activos
GET  /api/events/{id}              # Información de evento
POST /api/events/{id}/participate  # Participar en evento
```

### 🏅 Títulos
```http
GET  /api/titles/player/titles     # Títulos del jugador
POST /api/titles/equip             # Equipar título
GET  /api/titles/available         # Títulos disponibles
```

### 🔔 Notificaciones
```http
GET  /api/notifications            # Notificaciones del jugador
PUT  /api/notifications/{id}/read  # Marcar como leída
DELETE /api/notifications/{id}      # Eliminar notificación
```

### 🛠️ Administración
```http
GET  /api/admin/stats              # Estadísticas del servidor
GET  /api/admin/players            # Gestión de jugadores
POST /api/admin/ban                # Banear jugador
POST /api/admin/unban              # Desbanear jugador
```

### 🏥 Health Check
```http
GET  /health                       # Estado del servidor
GET  /api/server/info              # Información del servidor
```

## 🏗️ Arquitectura

### Estructura del Proyecto
```
ethernia-mmo-server/
├── main.go                    # Punto de entrada
├── server/                    # Configuración del servidor
├── handlers/                  # Manejadores HTTP
│   ├── auth_handler.go       # Autenticación
│   ├── village_handler.go     # Aldeas
│   ├── battle_handler.go      # Batallas
│   ├── chat_handler.go        # Chat
│   ├── alliance_handler.go    # Alianzas
│   └── ...                    # Otros handlers
├── services/                  # Lógica de negocio
│   ├── auth_service.go        # Servicios de autenticación
│   ├── village_service.go     # Servicios de aldeas
│   ├── battle_service.go      # Servicios de batallas
│   └── ...                    # Otros servicios
├── repository/                # Acceso a datos
│   ├── player_repository.go    # Repositorio de jugadores
│   ├── village_repository.go   # Repositorio de aldeas
│   └── ...                    # Otros repositorios
├── models/                    # Modelos de datos
│   ├── player.go              # Modelo de jugador
│   ├── village.go             # Modelo de aldea
│   ├── battle.go              # Modelo de batalla
│   └── ...                    # Otros modelos
├── middleware/                # Middleware HTTP
│   ├── auth.go                # Middleware de autenticación
│   └── validation.go          # Middleware de validación
├── auth/                      # Sistema de autenticación
│   └── jwt.go                 # Manejo de JWT
├── websocket/                 # WebSockets
│   └── manager.go             # Gestor de conexiones
├── config/                    # Configuración
│   ├── config.go              # Configuración principal
│   ├── config.yaml            # Archivo de configuración
│   └── database.go            # Configuración de BD
├── database/                  # Base de datos
│   ├── createdb.sql           # Script de creación
│   └── migrations/            # Migraciones
├── scripts/                   # Scripts de utilidad
│   ├── setup-database.sh     # Configurar BD
│   ├── start-dev.sh           # Iniciar desarrollo
│   └── ...                    # Otros scripts
└── deployments/               # Despliegue
    └── docker-compose.yml     # Docker Compose
```

### Tecnologías Utilizadas
- **Go 1.21+** - Lenguaje principal
- **Chi Router** - Enrutamiento HTTP
- **PostgreSQL** - Base de datos principal
- **Redis** - Cache y sesiones
- **JWT** - Autenticación
- **WebSockets** - Comunicación en tiempo real
- **Zap** - Logging estructurado
- **Viper** - Configuración
- **Docker** - Containerización

## 🧪 Testing

### Tests Unitarios
```bash
# Ejecutar todos los tests
go test ./...

# Tests con coverage
go test -cover ./...

# Tests específicos
go test ./handlers
go test ./services
```

### Tests de Integración
```bash
# Configurar base de datos de test
export DB_NAME=ethernia_test_db

# Ejecutar tests de integración
go test -tags=integration ./...
```

## 📈 Monitoreo

### Métricas Disponibles
- **Uptime del servidor**
- **Número de conexiones activas**
- **Tiempo de respuesta de endpoints**
- **Uso de memoria y CPU**
- **Errores y excepciones**
- **Actividad de usuarios**

### Logs
Los logs se generan en formato estructurado usando Zap:
- **INFO**: Operaciones normales
- **WARN**: Advertencias del sistema
- **ERROR**: Errores que requieren atención
- **DEBUG**: Información detallada para desarrollo

## 🚀 Despliegue

### Desarrollo Local
```bash
# Instalar dependencias
go mod download

# Configurar base de datos
./scripts/setup-database.sh

# Iniciar servidor
go run .
```

### Docker
```bash
# Construir imagen
docker build -t ethernia-mmo-server .

# Ejecutar contenedor
docker run -p 8080:8080 ethernia-mmo-server
```

### Producción
```bash
# Compilar para producción
GOOS=linux GOARCH=amd64 go build -o server main.go

# Configurar variables de entorno
export DB_HOST=production-db-host
export DB_PASSWORD=secure-password

# Ejecutar migraciones
./scripts/setup-database.sh

# Iniciar servidor
./server
```

## 🔄 Roadmap

### ✅ Completado (Implementado y Funcional)
- [x] **Sistema de Autenticación JWT** - Registro, login, perfiles, sesiones con Redis
- [x] **Gestión de Aldeas** - Creación, recursos, edificios con sistema de construcción avanzado
- [x] **Sistema de Construcción** - Mejora de edificios, cola de construcción, requisitos
- [x] **Sistema de Recursos** - Madera, piedra, comida, oro con generación automática
- [x] **Sistema de Batallas** - PvP, simulación de combate, reportes detallados, estadísticas
- [x] **Sistema de Unidades Militares** - Entrenamiento, estadísticas de combate, niveles
- [x] **Chat en Tiempo Real** - WebSockets, canales globales/alianza, moderación básica
- [x] **Sistema de Alianzas** - Creación, gestión de miembros, roles, rankings
- [x] **Sistema de Misiones** - Quests diarias/semanales, progreso, recompensas, categorías
- [x] **Sistema Económico** - Mercado, intercambio de monedas, estadísticas
- [x] **Sistema de Rankings** - Clasificaciones de jugadores, alianzas, aldeas
- [x] **Sistema de Notificaciones** - WebSocket, Redis pub/sub
- [x] **Base de Datos Completa** - PostgreSQL con todas las tablas y relaciones
- [x] **API REST Completa** - Todos los endpoints documentados y funcionales
- [x] **Middleware de Seguridad** - Autenticación, validación, rate limiting básico
- [x] **Logging Estructurado** - Zap logger con diferentes niveles
- [x] **Configuración Dinámica** - YAML config con variables de entorno

### 🚧 Parcialmente Implementado (Funcional pero con TODOs)
- [x] **Sistema de Héroes** - Modelos completos, servicios básicos (falta integración completa)
- [x] **Sistema de Investigación** - Modelos completos, servicios básicos (falta integración completa)
- [x] **Sistema de Logros** - Modelos completos, servicios básicos (falta integración completa)
- [x] **Sistema de Títulos** - Modelos completos, servicios básicos (falta integración completa)
- [x] **Sistema de Eventos** - Modelos completos, servicios básicos (falta integración completa)
- [x] **Herramientas de Administración** - Estructura básica (falta implementación completa)

### 📋 Próximas Funcionalidades (No Implementadas)
- [ ] **Sistema de Comercio Avanzado** - Subastas, contratos, economía dinámica
- [ ] **Sistema de Combate Mejorado** - Formaciones, tácticas, terreno, clima
- [ ] **Sistema de Clanes** - Estructura jerárquica avanzada
- [ ] **Sistema de Torneos** - Competencias organizadas
- [ ] **Sistema de Eventos Especiales** - Eventos temporales del servidor
- [ ] **Sistema de Recompensas Diarias** - Login rewards, daily bonuses
- [ ] **Sistema de Moderación Avanzada** - Herramientas de admin completas
- [ ] **Sistema de Analytics** - Métricas detalladas del juego
- [ ] **API GraphQL** - Alternativa a REST
- [ ] **Microservicios** - Arquitectura distribuida
- [ ] **Kubernetes Deployment** - Orquestación de contenedores

### 🛠️ Mejoras Técnicas Pendientes
- [ ] **Cache con Redis Optimizado** - Cache inteligente y estrategias de invalidación
- [ ] **Rate Limiting Avanzado** - Límites por usuario, endpoint, IP
- [ ] **API Versioning** - Control de versiones de API
- [ ] **OpenAPI/Swagger Docs** - Documentación interactiva
- [ ] **Tests de Carga** - Pruebas de rendimiento
- [ ] **CI/CD Pipeline** - Automatización de despliegue
- [ ] **Monitoring Dashboard** - Panel de monitoreo en tiempo real
- [ ] **Métricas con Prometheus** - Métricas de aplicación
- [ ] **Tracing con Jaeger** - Trazabilidad de requests
- [ ] **Tests Unitarios Completos** - Cobertura de tests al 100%
- [ ] **Tests de Integración** - Pruebas end-to-end

## 🤝 Contribución

### Cómo Contribuir
1. Fork el repositorio
2. Crea una rama para tu feature (`git checkout -b feature/nueva-funcionalidad`)
3. Commit tus cambios (`git commit -am 'Agregar nueva funcionalidad'`)
4. Push a la rama (`git push origin feature/nueva-funcionalidad`)
5. Crea un Pull Request

### Estándares de Código
- Usar `gofmt` para formateo
- Seguir las convenciones de Go
- Escribir tests para nuevas funcionalidades
- Documentar funciones públicas
- Usar nombres descriptivos para variables y funciones

### Reportar Bugs
- Usar el sistema de issues de GitHub
- Incluir pasos para reproducir el bug
- Adjuntar logs relevantes
- Especificar versión del servidor y sistema operativo

## 📄 Licencia

Este proyecto está bajo la Licencia MIT. Ver el archivo [LICENSE](LICENSE) para más detalles.

## 🆘 Soporte

### Canales de Soporte
- **Issues de GitHub**: Para bugs y feature requests
- **Discussions**: Para preguntas generales
- **Wiki**: Documentación adicional

### Recursos Útiles
- [Documentación de la API](API_DOCUMENTATION.md)
- [Guía de Instalación](#inicio-rápido)
- [Ejemplos de Uso](API_DOCUMENTATION.md#ejemplos-de-uso)

### Comunidad
- Únete a nuestro Discord
- Participa en las discusiones de GitHub
- Comparte tus proyectos creados con este servidor

## 🌟 Características Destacadas

### 🎮 Para Desarrolladores de Juegos
- **API REST completa** - Todos los endpoints necesarios
- **WebSockets** - Comunicación en tiempo real
- **Documentación detallada** - Ejemplos en múltiples lenguajes
- **Arquitectura escalable** - Preparado para miles de jugadores

### 🏗️ Para Desarrolladores Backend
- **Código limpio** - Arquitectura bien estructurada
- **Tests incluidos** - Cobertura de tests unitarios
- **Docker ready** - Fácil despliegue
- **Monitoreo integrado** - Logs y métricas

### 🎯 Para Desarrolladores Frontend
- **Endpoints bien definidos** - Respuestas consistentes
- **Autenticación JWT** - Fácil integración
- **WebSockets** - Chat y notificaciones en tiempo real
- **CORS configurado** - Listo para desarrollo web

---

## 🎮 Clientes Compatibles

Este servidor backend es compatible con cualquier tecnología frontend:

- **Flutter** (Android, iOS, Web, Desktop)
- **React** (Web, React Native)
- **Vue.js** (Web)
- **Angular** (Web)
- **Unity** (PC, Mobile, WebGL)
- **Unreal Engine** (PC, Mobile)
- **JavaScript/TypeScript** (Web, Node.js)
- **Python** (Web, Desktop)
- **C#** (Web, Desktop, Mobile)
- **Java** (Web, Desktop, Mobile)

**¡Solo necesitas hacer requests HTTP y conectar WebSockets!** 🚀

---

**Desarrollado con ❤️ por [@32bitsarg](https://github.com/32bitsarg)**

*¿Tienes preguntas? ¡Abre un issue o únete a la discusión!*