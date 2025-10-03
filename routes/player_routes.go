package routes

import (
	"net/http"
	"server-backend/repository"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// SetupPlayerRoutes configura todas las rutas relacionadas con jugadores
func SetupPlayerRoutes(r *gin.RouterGroup, playerRepo *repository.PlayerRepository, villageRepo *repository.VillageRepository, logger *zap.Logger) {
	// Grupo de rutas de jugadores (ya protegido por el grupo padre)
	playerGroup := r.Group("/players")

	playerGroup.GET("/", func(c *gin.Context) {
		players, err := playerRepo.GetAllPlayers()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, players)
	})

	playerGroup.GET("/:playerID", func(c *gin.Context) {
		playerIDStr := c.Param("playerID")
		playerID, err := uuid.Parse(playerIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ID de jugador inválido"})
			return
		}
		player, err := playerRepo.GetPlayerByID(playerID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, player)
	})

	playerGroup.GET("/:playerID/villages", func(c *gin.Context) {
		playerIDStr := c.Param("playerID")
		playerID, err := uuid.Parse(playerIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ID de jugador inválido"})
			return
		}
		villages, err := villageRepo.GetVillagesByPlayerID(playerID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, villages)
	})

	logger.Info("✅ Rutas de jugadores configuradas exitosamente")
}
