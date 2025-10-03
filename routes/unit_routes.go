package routes

import (
	"server-backend/handlers"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// SetupUnitRoutes configura todas las rutas relacionadas con unidades
func SetupUnitRoutes(r *gin.RouterGroup, unitHandler *handlers.UnitHandler, logger *zap.Logger) {
	// Grupo de rutas de unidades (ya protegido por el grupo padre)
	unitGroup := r.Group("/api/units")

	unitGroup.GET("/", unitHandler.GetUnits)
	unitGroup.POST("/train", unitHandler.TrainUnits)

	logger.Info("✅ Rutas de unidades configuradas exitosamente")
}
