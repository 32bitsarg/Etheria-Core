package routes

import (
	"server-backend/auth"
	"server-backend/handlers"
	"server-backend/middleware"
	"server-backend/repository"
	"server-backend/services"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// SetupAllRoutes configura todas las rutas de la aplicación
func SetupAllRoutes(r *gin.Engine, handlers *Handlers, repos *Repositories, services *Services, authMiddleware *middleware.AuthMiddleware, logger *zap.Logger) {
	// Configurar middleware global
	setupGlobalMiddleware(r, logger)

	// Configurar rutas públicas
	SetupAuthRoutes(r, handlers.Auth, logger)

	// Configurar rutas protegidas con middleware de autenticación Gin nativo
	protected := r.Group("/")
	protected.Use(authMiddleware.RequireAuthGin())

	// Configurar todas las rutas protegidas en el grupo centralizado
	SetupVillageRoutes(protected, handlers.Village, services.Resource, repos.Village, logger)
	SetupChatRoutes(protected, handlers.Chat, authMiddleware, logger)
	SetupPlayerRoutes(protected, repos.Player, repos.Village, logger)
	SetupAllianceRoutes(protected, handlers.Alliance, logger)
	SetupUnitRoutes(protected, handlers.Unit, logger)
	SetupBuildingRoutes(protected, repos.Village, logger)

	// Configurar rutas protegidas de autenticación
	protected.GET("/auth/profile", handlers.Auth.GetProfile)
	protected.PUT("/auth/profile", handlers.Auth.UpdateProfile)
	protected.POST("/auth/logout", handlers.Auth.Logout)

	logger.Info("✅ Todas las rutas configuradas exitosamente")
}

// setupGlobalMiddleware configura el middleware global
func setupGlobalMiddleware(r *gin.Engine, logger *zap.Logger) {
	// Middleware de logging personalizado y legible
	r.Use(middleware.CustomLoggingMiddleware(middleware.LoggingConfig{
		Logger:       logger,
		SkipPaths:    []string{"/health", "/api/health"},
		EnableColors: true,
	}))

	// CORS
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token")
		c.Header("Access-Control-Expose-Headers", "Link")
		c.Header("Access-Control-Allow-Credentials", "false")
		c.Header("Access-Control-Max-Age", "300")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})
}

// Handlers contiene todos los handlers
type Handlers struct {
	Auth     *handlers.AuthHandler
	Village  *handlers.VillageHandler
	Chat     *handlers.ChatHandler
	Alliance *handlers.AllianceHandler
	Unit     *handlers.UnitHandler
}

// Repositories contiene todos los repositorios
type Repositories struct {
	Player   *repository.PlayerRepository
	Village  *repository.VillageRepository
	Alliance *repository.AllianceRepository
	Unit     *repository.UnitRepository
}

// Services contiene todos los servicios
type Services struct {
	Resource *services.ResourceService
	JWT      *auth.JWTManager
	Redis    *services.RedisService
	Chat     *services.ChatService
}
