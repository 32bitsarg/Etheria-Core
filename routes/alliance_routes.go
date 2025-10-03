package routes

import (
	"server-backend/handlers"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// SetupAllianceRoutes configura todas las rutas relacionadas con alianzas
func SetupAllianceRoutes(r *gin.RouterGroup, allianceHandler *handlers.AllianceHandler, logger *zap.Logger) {
	// Grupo de rutas de alianzas (ya protegido por el grupo padre)
	allianceGroup := r.Group("/api/alliances")

	allianceGroup.GET("/", allianceHandler.GetAlliances)
	allianceGroup.POST("/", allianceHandler.CreateAlliance)
	allianceGroup.GET("/:id", allianceHandler.GetAlliance)
	allianceGroup.POST("/:id/join", allianceHandler.JoinAlliance)
	allianceGroup.POST("/:id/leave", allianceHandler.LeaveAlliance)

	logger.Info("✅ Rutas de alianzas configuradas exitosamente")
}
