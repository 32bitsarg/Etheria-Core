package routes

import (
	"net/http"
	"server-backend/handlers"
	"server-backend/repository"
	"server-backend/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// SetupVillageRoutes configura todas las rutas relacionadas con aldeas
func SetupVillageRoutes(r *gin.RouterGroup, villageHandler *handlers.VillageHandler, resourceService *services.ResourceService, villageRepo *repository.VillageRepository, logger *zap.Logger) {
	// Grupo de rutas de aldeas (ya protegido por el grupo padre)
	villageGroup := r.Group("/api/villages")

	// Rutas básicas de aldeas
	villageGroup.GET("/", villageHandler.GetPlayerVillages)
	villageGroup.GET("/:villageID", villageHandler.GetVillage)

	// Rutas de construcción de edificios
	villageGroup.GET("/:villageID/buildings/:buildingType/upgrade-info", villageHandler.GetBuildingUpgradeInfo)
	villageGroup.POST("/:villageID/buildings/:buildingType/upgrade", villageHandler.UpgradeBuilding)
	villageGroup.POST("/:villageID/buildings/:buildingType/complete", villageHandler.CompleteBuildingUpgrade)
	villageGroup.DELETE("/:villageID/buildings/:buildingType/upgrade", villageHandler.CancelBuildingUpgrade)
	villageGroup.GET("/:villageID/buildings/:buildingType/time-remaining", villageHandler.GetBuildingUpgradeTimeRemaining)

	// Rutas de cola de construcción
	villageGroup.GET("/:villageID/buildings/:buildingType/requirements", villageHandler.CheckBuildingRequirements)
	villageGroup.POST("/:villageID/construction-queue/process", villageHandler.ProcessConstructionQueue)
	villageGroup.GET("/:villageID/construction-queue", villageHandler.GetConstructionQueue)
	villageGroup.GET("/:villageID/construction-queue/status", villageHandler.GetConstructionQueueStatus)

	// Rutas de recursos
	resourceGroup := r.Group("/resources")
	resourceGroup.Use(func(c *gin.Context) {
		// Middleware de autenticación se aplicará aquí
		c.Next()
	})

	resourceGroup.GET("/village/:villageID/production", func(c *gin.Context) {
		villageIDStr := c.Param("villageID")
		villageID, err := uuid.Parse(villageIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ID de aldea inválido"})
			return
		}
		production, err := resourceService.GetVillageProduction(villageID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, production)
	})

	resourceGroup.GET("/village/:villageID/storage", func(c *gin.Context) {
		villageIDStr := c.Param("villageID")
		villageID, err := uuid.Parse(villageIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ID de aldea inválido"})
			return
		}
		storage, err := resourceService.GetVillageStorage(villageID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, storage)
	})

	resourceGroup.GET("/village/:villageID/current", func(c *gin.Context) {
		villageIDStr := c.Param("villageID")
		villageID, err := uuid.Parse(villageIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ID de aldea inválido"})
			return
		}
		resources, err := resourceService.GetResourceInfo(villageID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, resources)
	})

	// Rutas de clientes
	clientGroup := r.Group("/client/villages")
	clientGroup.Use(func(c *gin.Context) {
		// Middleware de autenticación se aplicará aquí
		c.Next()
	})

	clientGroup.GET("/:villageID", villageHandler.GetVillage)
	clientGroup.GET("/:villageID/resources", func(c *gin.Context) {
		villageIDStr := c.Param("villageID")
		villageID, err := uuid.Parse(villageIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ID de aldea inválido"})
			return
		}
		resources, err := resourceService.GetResourceInfo(villageID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, resources)
	})
	clientGroup.GET("/:villageID/production", func(c *gin.Context) {
		villageIDStr := c.Param("villageID")
		villageID, err := uuid.Parse(villageIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ID de aldea inválido"})
			return
		}
		production, err := resourceService.GetVillageProduction(villageID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, production)
	})
	clientGroup.GET("/:villageID/storage", func(c *gin.Context) {
		villageIDStr := c.Param("villageID")
		villageID, err := uuid.Parse(villageIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ID de aldea inválido"})
			return
		}
		storage, err := resourceService.GetVillageStorage(villageID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, storage)
	})

	logger.Info("✅ Rutas de aldeas configuradas exitosamente")
}
