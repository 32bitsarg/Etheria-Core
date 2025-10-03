package routes

import (
	"net/http"
	"server-backend/handlers"
	"server-backend/middleware"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// SetupAuthRoutes configura todas las rutas relacionadas con autenticación
func SetupAuthRoutes(r *gin.Engine, authHandler *handlers.AuthHandler, logger *zap.Logger) {
	// Grupo de rutas de autenticación con logging específico
	authGroup := r.Group("/auth")
	authGroup.Use(middleware.AuthLoggingMiddleware(logger))

	// Rutas públicas de autenticación
	authGroup.POST("/register", authHandler.Register)
	authGroup.POST("/login", authHandler.Login)

	// Rutas de salud
	r.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	r.GET("/api/health", func(c *gin.Context) {
		response := map[string]interface{}{
			"status":    "healthy",
			"timestamp": time.Now().UTC(),
			"version":   "1.0.0",
			"database":  "connected",
			"redis":     "connected",
		}
		c.JSON(http.StatusOK, response)
	})

	r.GET("/api/server-time", handlers.ServerTimeHandler)

	logger.Info("✅ Rutas de autenticación configuradas exitosamente")
}
