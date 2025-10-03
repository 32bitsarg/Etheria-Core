package routes

import (
	"net/http"
	"server-backend/repository"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// SetupBuildingRoutes configura todas las rutas relacionadas con edificios
func SetupBuildingRoutes(r *gin.RouterGroup, villageRepo *repository.VillageRepository, logger *zap.Logger) {
	// Grupo de rutas de edificios (ya protegido por el grupo padre)
	buildingGroup := r.Group("/buildings")

	buildingGroup.GET("/", func(c *gin.Context) {
		buildings, err := villageRepo.GetBuildingTypes()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, buildings)
	})

	buildingGroup.GET("/:buildingID", func(c *gin.Context) {
		buildingID := c.Param("buildingID")
		building, err := villageRepo.GetBuildingType(buildingID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, building)
	})

	logger.Info("✅ Rutas de edificios configuradas exitosamente")
}
