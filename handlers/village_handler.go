package handlers

import (
	"fmt"
	"net/http"
	"server-backend/repository"
	"server-backend/services"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type VillageHandler struct {
	villageRepo         *repository.VillageRepository
	constructionService *services.ConstructionService
	logger              *zap.Logger
}

func NewVillageHandler(villageRepo *repository.VillageRepository, constructionService *services.ConstructionService, logger *zap.Logger) *VillageHandler {
	return &VillageHandler{
		villageRepo:         villageRepo,
		constructionService: constructionService,
		logger:              logger,
	}
}

func (h *VillageHandler) GetVillage(c *gin.Context) {
	// Obtener el ID del jugador del contexto
	playerIDStr := c.GetString("player_id")
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		h.logger.Error("Error parseando ID de jugador", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno del servidor"})
		return
	}

	// Obtener las aldeas del jugador
	villages, err := h.villageRepo.GetVillagesByPlayerID(playerID)
	if err != nil {
		h.logger.Error("Error obteniendo aldeas", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno del servidor"})
		return
	}

	// Si no hay aldeas, devolver error
	if len(villages) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No tienes aldeas"})
		return
	}

	// Por ahora, devolver la primera aldea (se puede mejorar para manejar múltiples aldeas)
	village := villages[0]

	c.JSON(http.StatusOK, village)
}

func (h *VillageHandler) GetPlayerVillages(c *gin.Context) {
	// Obtener el ID del jugador del contexto
	playerIDStr := c.GetString("player_id")
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		h.logger.Error("Error parseando ID de jugador", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno del servidor"})
		return
	}

	// Obtener las aldeas del jugador
	villages, err := h.villageRepo.GetVillagesByPlayerID(playerID)
	if err != nil {
		h.logger.Error("Error obteniendo aldeas", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno del servidor"})
		return
	}

	c.JSON(http.StatusOK, villages)
}

type UpgradeBuildingRequest struct {
	BuildingType string `json:"building_type"`
}

func (h *VillageHandler) UpgradeBuilding(c *gin.Context) {
	villageIDStr := c.Param("villageID")
	villageID, err := uuid.Parse(villageIDStr)
	if err != nil {
		h.logger.Error("Error parseando villageID",
			zap.String("villageID", villageIDStr),
			zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de aldea inválido"})
		return
	}

	// Obtener la aldea para verificar permisos
	village, err := h.villageRepo.GetVillageByID(villageID)
	if err != nil {
		h.logger.Error("Error obteniendo aldea", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno del servidor"})
		return
	}
	if village == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Aldea no encontrada"})
		return
	}

	// Verificar que el jugador tiene acceso a la aldea
	playerIDStr := c.GetString("player_id")
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		h.logger.Error("Error parseando ID de jugador", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno del servidor"})
		return
	}
	if village.Village.PlayerID != playerID {
		c.JSON(http.StatusForbidden, gin.H{"error": "No tienes permiso para actualizar esta aldea"})
		return
	}

	// Decodificar la solicitud
	var req UpgradeBuildingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error decodificando la solicitud"})
		return
	}

	// Validar el tipo de edificio
	if strings.TrimSpace(req.BuildingType) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "El tipo de edificio es requerido"})
		return
	}

	// Usar el servicio de construcción para manejar la mejora
	result, err := h.constructionService.UpgradeBuilding(villageID, req.BuildingType)
	if err != nil {
		switch err {
		case services.ErrInsufficientResources:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Recursos insuficientes para mejorar el edificio"})
		case services.ErrBuildingMaxLevel:
			c.JSON(http.StatusBadRequest, gin.H{"error": "El edificio ya está en su nivel máximo"})
		case services.ErrBuildingUpgrading:
			c.JSON(http.StatusBadRequest, gin.H{"error": "El edificio ya está siendo mejorado"})
		case services.ErrInvalidBuildingType:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Tipo de edificio inválido"})
		case services.ErrTownHallRequired:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Se requiere un ayuntamiento de nivel superior"})
		default:
			h.logger.Error("Error mejorando edificio", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno del servidor"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Edificio mejorado exitosamente",
		"result":  result,
	})
}

// GetBuildingUpgradeInfo obtiene información sobre la mejora de un edificio
func (h *VillageHandler) GetBuildingUpgradeInfo(c *gin.Context) {
	villageIDStr := c.Param("villageID")
	buildingType := c.Param("buildingType")

	villageID, err := uuid.Parse(villageIDStr)
	if err != nil {
		h.logger.Error("Error parseando villageID",
			zap.String("villageID", villageIDStr),
			zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de aldea inválido"})
		return
	}

	if buildingType == "" {
		h.logger.Error("buildingType vacío")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tipo de edificio requerido"})
		return
	}

	// Verificar permisos
	village, err := h.villageRepo.GetVillageByID(villageID)
	if err != nil {
		h.logger.Error("Error obteniendo aldea", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno del servidor"})
		return
	}
	if village == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Aldea no encontrada"})
		return
	}

	playerIDStr := c.GetString("player_id")
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		h.logger.Error("Error parseando ID de jugador", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno del servidor"})
		return
	}
	if village.Village.PlayerID != playerID {
		c.JSON(http.StatusForbidden, gin.H{"error": "No tienes permiso para ver esta aldea"})
		return
	}

	// Obtener información de mejora usando el servicio avanzado
	upgradeInfo, err := h.constructionService.GetUpgradeInfo(villageID, buildingType)
	if err != nil {
		h.logger.Error("Error obteniendo información de mejora", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno del servidor"})
		return
	}

	c.JSON(http.StatusOK, upgradeInfo)
}

// CheckBuildingRequirements verifica los requisitos para construir un edificio
func (h *VillageHandler) CheckBuildingRequirements(c *gin.Context) {
	// Obtener el ID de la aldea de la URL
	villageIDStr := c.Param("villageID")
	villageID, err := uuid.Parse(villageIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de aldea inválido"})
		return
	}

	// Obtener el tipo de edificio y nivel objetivo de los query parameters
	buildingType := c.Query("buildingType")
	if buildingType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tipo de edificio requerido"})
		return
	}

	targetLevelStr := c.Query("targetLevel")
	targetLevel := 1 // Por defecto
	if targetLevelStr != "" {
		if level, err := strconv.Atoi(targetLevelStr); err == nil {
			targetLevel = level
		}
	}

	// Verificar permisos
	village, err := h.villageRepo.GetVillageByID(villageID)
	if err != nil {
		h.logger.Error("Error obteniendo aldea", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno del servidor"})
		return
	}
	if village == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Aldea no encontrada"})
		return
	}

	playerIDStr := c.GetString("player_id")
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		h.logger.Error("Error parseando ID de jugador", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno del servidor"})
		return
	}
	if village.Village.PlayerID != playerID {
		c.JSON(http.StatusForbidden, gin.H{"error": "No tienes permiso para ver esta aldea"})
		return
	}

	// Verificar requisitos usando el servicio avanzado
	requirements, err := h.constructionService.CheckBuildingRequirements(villageID, buildingType, targetLevel)
	if err != nil {
		h.logger.Error("Error verificando requisitos", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno del servidor"})
		return
	}

	c.JSON(http.StatusOK, requirements)
}

// ProcessConstructionQueue procesa la cola de construcción de una aldea
func (h *VillageHandler) ProcessConstructionQueue(c *gin.Context) {
	// Obtener el ID de la aldea de la URL
	villageIDStr := c.Param("villageID")
	villageID, err := uuid.Parse(villageIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de aldea inválido"})
		return
	}

	// Verificar permisos
	village, err := h.villageRepo.GetVillageByID(villageID)
	if err != nil {
		h.logger.Error("Error obteniendo aldea", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno del servidor"})
		return
	}
	if village == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Aldea no encontrada"})
		return
	}

	playerIDStr := c.GetString("player_id")
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		h.logger.Error("Error parseando ID de jugador", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno del servidor"})
		return
	}
	if village.Village.PlayerID != playerID {
		c.JSON(http.StatusForbidden, gin.H{"error": "No tienes permiso para procesar esta aldea"})
		return
	}

	// Procesar cola de construcción usando el servicio avanzado
	results, err := h.constructionService.ProcessConstructionQueue(villageID)
	if err != nil {
		h.logger.Error("Error procesando cola de construcción", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno del servidor"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Cola de construcción procesada",
		"results": results,
	})
}

// GetConstructionQueue obtiene la cola de construcción de una aldea
func (h *VillageHandler) GetConstructionQueue(c *gin.Context) {
	// Obtener el ID de la aldea de la URL
	villageIDStr := c.Param("villageID")
	villageID, err := uuid.Parse(villageIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de aldea inválido"})
		return
	}

	// Verificar permisos
	village, err := h.villageRepo.GetVillageByID(villageID)
	if err != nil {
		h.logger.Error("Error obteniendo aldea", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno del servidor"})
		return
	}
	if village == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Aldea no encontrada"})
		return
	}

	playerIDStr := c.GetString("player_id")
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		h.logger.Error("Error parseando ID de jugador", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno del servidor"})
		return
	}
	if village.Village.PlayerID != playerID {
		c.JSON(http.StatusForbidden, gin.H{"error": "No tienes permiso para ver esta aldea"})
		return
	}

	// Obtener cola de construcción
	queue, err := h.constructionService.GetConstructionQueue(villageID)
	if err != nil {
		h.logger.Error("Error obteniendo cola de construcción", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno del servidor"})
		return
	}

	c.JSON(http.StatusOK, queue)
}

// CancelUpgrade cancela la mejora de un edificio
func (h *VillageHandler) CancelUpgrade(c *gin.Context) {
	// Obtener el ID de la aldea de la URL
	villageIDStr := c.Param("villageID")
	villageID, err := uuid.Parse(villageIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de aldea inválido"})
		return
	}

	// Obtener el tipo de edificio de la URL
	buildingType := c.Param("buildingType")
	if buildingType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tipo de edificio requerido"})
		return
	}

	// Verificar permisos
	village, err := h.villageRepo.GetVillageByID(villageID)
	if err != nil {
		h.logger.Error("Error obteniendo aldea", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno del servidor"})
		return
	}
	if village == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Aldea no encontrada"})
		return
	}

	playerIDStr := c.GetString("player_id")
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		h.logger.Error("Error parseando ID de jugador", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno del servidor"})
		return
	}
	if village.Village.PlayerID != playerID {
		c.JSON(http.StatusForbidden, gin.H{"error": "No tienes permiso para cancelar en esta aldea"})
		return
	}

	// Cancelar la mejora con reembolso
	result, err := h.constructionService.CancelUpgradeWithRefund(villageID, buildingType)
	if err != nil {
		switch err.Error() {
		case "el edificio no está siendo mejorado":
			c.JSON(http.StatusBadRequest, gin.H{"error": "El edificio no está siendo mejorado"})
		case "tipo de edificio inválido":
			c.JSON(http.StatusBadRequest, gin.H{"error": "Tipo de edificio inválido"})
		case "aldea no encontrada":
			c.JSON(http.StatusNotFound, gin.H{"error": "Aldea no encontrada"})
		default:
			h.logger.Error("Error cancelando mejora", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno del servidor"})
		}
		return
	}

	// Respuesta exitosa
	c.JSON(http.StatusOK, gin.H{
		"message":           "Mejora cancelada exitosamente",
		"building_type":     result.BuildingType,
		"refund_amount":     result.RefundAmount,
		"refund_percentage": result.RefundPercentage,
		"original_cost":     result.OriginalCost,
		"cancelled_at":      result.CancelledAt,
		"time_remaining":    result.TimeRemaining.String(),
		"refund_reason":     result.RefundReason,
	})
}

// CompleteBuildingUpgrade completa la mejora de un edificio
func (h *VillageHandler) CompleteBuildingUpgrade(c *gin.Context) {
	villageIDStr := c.Param("villageID")
	villageID, err := uuid.Parse(villageIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de aldea inválido"})
		return
	}

	buildingType := c.Param("buildingType")
	if buildingType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tipo de edificio requerido"})
		return
	}

	// Verificar permisos
	village, err := h.villageRepo.GetVillageByID(villageID)
	if err != nil {
		h.logger.Error("Error obteniendo aldea", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno del servidor"})
		return
	}
	if village == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Aldea no encontrada"})
		return
	}

	playerIDStr := c.GetString("player_id")
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		h.logger.Error("Error parseando ID de jugador", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno del servidor"})
		return
	}
	if village.Village.PlayerID != playerID {
		c.JSON(http.StatusForbidden, gin.H{"error": "No tienes permiso para acceder a esta aldea"})
		return
	}

	// Completar la mejora
	err = h.constructionService.CompleteUpgrade(villageID, buildingType)
	if err != nil {
		switch err.Error() {
		case "el edificio no está siendo mejorado":
			c.JSON(http.StatusBadRequest, gin.H{"error": "El edificio no está siendo mejorado"})
		case "la mejora aún no ha terminado":
			c.JSON(http.StatusBadRequest, gin.H{"error": "La mejora aún no ha terminado"})
		case services.ErrInvalidBuildingType.Error():
			c.JSON(http.StatusBadRequest, gin.H{"error": "Tipo de edificio inválido"})
		default:
			h.logger.Error("Error completando mejora", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno del servidor"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Mejora completada exitosamente",
	})
}

// CancelBuildingUpgrade cancela la mejora de un edificio con reembolso del 50%
func (h *VillageHandler) CancelBuildingUpgrade(c *gin.Context) {
	villageIDStr := c.Param("villageID")
	villageID, err := uuid.Parse(villageIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de aldea inválido"})
		return
	}

	buildingType := c.Param("buildingType")
	if buildingType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tipo de edificio requerido"})
		return
	}

	// Verificar permisos
	village, err := h.villageRepo.GetVillageByID(villageID)
	if err != nil {
		h.logger.Error("Error obteniendo aldea", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno del servidor"})
		return
	}
	if village == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Aldea no encontrada"})
		return
	}

	playerIDStr := c.GetString("player_id")
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		h.logger.Error("Error parseando ID de jugador", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno del servidor"})
		return
	}
	if village.Village.PlayerID != playerID {
		c.JSON(http.StatusForbidden, gin.H{"error": "No tienes permiso para cancelar mejoras en esta aldea"})
		return
	}

	// Cancelar la mejora con reembolso
	result, err := h.constructionService.CancelUpgradeWithRefund(villageID, buildingType)
	if err != nil {
		switch err.Error() {
		case "el edificio no está siendo mejorado":
			c.JSON(http.StatusBadRequest, gin.H{"error": "El edificio no está siendo mejorado"})
		case "tipo de edificio inválido":
			c.JSON(http.StatusBadRequest, gin.H{"error": "Tipo de edificio inválido"})
		case "aldea no encontrada":
			c.JSON(http.StatusNotFound, gin.H{"error": "Aldea no encontrada"})
		default:
			h.logger.Error("Error cancelando mejora", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno del servidor"})
		}
		return
	}

	// Respuesta exitosa
	c.JSON(http.StatusOK, gin.H{
		"message":           "Mejora cancelada exitosamente",
		"building_type":     result.BuildingType,
		"refund_amount":     result.RefundAmount,
		"refund_percentage": result.RefundPercentage,
		"original_cost":     result.OriginalCost,
		"cancelled_at":      result.CancelledAt,
		"time_remaining":    result.TimeRemaining.String(),
		"refund_reason":     result.RefundReason,
	})
}

// GetBuildingUpgradeTimeRemaining obtiene el tiempo restante de mejora en tiempo real
func (h *VillageHandler) GetBuildingUpgradeTimeRemaining(c *gin.Context) {
	// Obtener el ID de la aldea de la URL
	villageIDStr := c.Param("villageID")
	villageID, err := uuid.Parse(villageIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de aldea inválido"})
		return
	}

	// Obtener el tipo de edificio de la URL
	buildingType := c.Param("buildingType")
	if buildingType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tipo de edificio requerido"})
		return
	}

	// Verificar permisos
	village, err := h.villageRepo.GetVillageByID(villageID)
	if err != nil {
		h.logger.Error("Error obteniendo aldea", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno del servidor"})
		return
	}
	if village == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Aldea no encontrada"})
		return
	}

	playerIDStr := c.GetString("player_id")
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		h.logger.Error("Error parseando ID de jugador", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno del servidor"})
		return
	}
	if village.Village.PlayerID != playerID {
		c.JSON(http.StatusForbidden, gin.H{"error": "No tienes permiso para acceder a esta aldea"})
		return
	}

	// Obtener información del edificio
	building, exists := village.Buildings[buildingType]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Edificio no encontrado"})
		return
	}

	// Verificar si está mejorándose
	if !building.IsUpgrading || building.UpgradeCompletionTime == nil {
		c.JSON(http.StatusOK, gin.H{
			"building_type":   buildingType,
			"is_upgrading":    false,
			"time_remaining":  0,
			"completion_time": nil,
			"formatted_time":  "00:00",
		})
		return
	}

	// Calcular tiempo restante
	now := time.Now()
	timeRemaining := building.UpgradeCompletionTime.Sub(now)

	// Si ya terminó, devolver 0
	if timeRemaining <= 0 {
		c.JSON(http.StatusOK, gin.H{
			"building_type":   buildingType,
			"is_upgrading":    false,
			"time_remaining":  0,
			"completion_time": building.UpgradeCompletionTime,
			"formatted_time":  "00:00",
			"can_complete":    true,
		})
		return
	}

	// Formatear tiempo en formato MM:SS
	minutes := int(timeRemaining.Minutes())
	seconds := int(timeRemaining.Seconds()) % 60
	formattedTime := fmt.Sprintf("%02d:%02d", minutes, seconds)

	c.JSON(http.StatusOK, gin.H{
		"building_type":   buildingType,
		"is_upgrading":    true,
		"time_remaining":  int64(timeRemaining.Seconds()),
		"completion_time": building.UpgradeCompletionTime,
		"formatted_time":  formattedTime,
		"can_complete":    false,
	})
}

// GetConstructionQueueStatus obtiene el estado de la cola de construcción
func (h *VillageHandler) GetConstructionQueueStatus(c *gin.Context) {
	villageIDStr := c.Param("villageID")

	villageID, err := uuid.Parse(villageIDStr)
	if err != nil {
		h.logger.Error("Error parseando villageID",
			zap.String("villageID", villageIDStr),
			zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de aldea inválido"})
		return
	}

	// Verificar que la aldea existe y pertenece al jugador
	village, err := h.villageRepo.GetVillageByID(villageID)
	if err != nil {
		h.logger.Error("Error obteniendo aldea", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno del servidor"})
		return
	}
	if village == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Aldea no encontrada"})
		return
	}

	playerIDStr := c.GetString("player_id")
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		h.logger.Error("Error parseando ID de jugador", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno del servidor"})
		return
	}
	if village.Village.PlayerID != playerID {
		c.JSON(http.StatusForbidden, gin.H{"error": "No tienes permiso para ver esta aldea"})
		return
	}

	// Obtener estado de la cola de construcción
	queueStatus, err := h.constructionService.GetConstructionQueueStatus(villageID)
	if err != nil {
		h.logger.Error("Error obteniendo estado de cola", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno del servidor"})
		return
	}

	c.JSON(http.StatusOK, queueStatus)
}
