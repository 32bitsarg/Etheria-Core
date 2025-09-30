package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"server-backend/repository"
	"server-backend/services"
	"strings"
	"time"

	"strconv"

	"github.com/go-chi/chi/v5"
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

func (h *VillageHandler) GetVillage(w http.ResponseWriter, r *http.Request) {
	// Obtener el ID del jugador del contexto
	playerIDStr := r.Context().Value("player_id").(string)
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		h.logger.Error("Error parseando ID de jugador", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	// Obtener las aldeas del jugador
	villages, err := h.villageRepo.GetVillagesByPlayerID(playerID)
	if err != nil {
		h.logger.Error("Error obteniendo aldeas", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	// Si no hay aldeas, devolver error
	if len(villages) == 0 {
		http.Error(w, "No tienes aldeas", http.StatusNotFound)
		return
	}

	// Por ahora, devolver la primera aldea (se puede mejorar para manejar múltiples aldeas)
	village := villages[0]

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(village)
}

func (h *VillageHandler) GetPlayerVillages(w http.ResponseWriter, r *http.Request) {
	// Obtener el ID del jugador del contexto
	playerIDStr := r.Context().Value("player_id").(string)
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		h.logger.Error("Error parseando ID de jugador", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	// Obtener las aldeas del jugador
	villages, err := h.villageRepo.GetVillagesByPlayerID(playerID)
	if err != nil {
		h.logger.Error("Error obteniendo aldeas", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(villages)
}

type UpgradeBuildingRequest struct {
	BuildingType string `json:"building_type"`
}

func (h *VillageHandler) UpgradeBuilding(w http.ResponseWriter, r *http.Request) {
	// Obtener el ID de la aldea de la URL
	villageIDStr := chi.URLParam(r, "id")
	villageID, err := uuid.Parse(villageIDStr)
	if err != nil {
		http.Error(w, "ID de aldea inválido", http.StatusBadRequest)
		return
	}

	// Obtener la aldea para verificar permisos
	village, err := h.villageRepo.GetVillageByID(villageID)
	if err != nil {
		h.logger.Error("Error obteniendo aldea", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}
	if village == nil {
		http.Error(w, "Aldea no encontrada", http.StatusNotFound)
		return
	}

	// Verificar que el jugador tiene acceso a la aldea
	playerIDStr := r.Context().Value("player_id").(string)
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		h.logger.Error("Error parseando ID de jugador", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}
	if village.Village.PlayerID != playerID {
		http.Error(w, "No tienes permiso para actualizar esta aldea", http.StatusForbidden)
		return
	}

	// Decodificar la solicitud
	var req UpgradeBuildingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Error decodificando la solicitud", http.StatusBadRequest)
		return
	}

	// Validar el tipo de edificio
	if strings.TrimSpace(req.BuildingType) == "" {
		http.Error(w, "El tipo de edificio es requerido", http.StatusBadRequest)
		return
	}

	// Usar el servicio de construcción para manejar la mejora
	result, err := h.constructionService.UpgradeBuilding(villageID, req.BuildingType)
	if err != nil {
		switch err {
		case services.ErrInsufficientResources:
			http.Error(w, "Recursos insuficientes para mejorar el edificio", http.StatusBadRequest)
		case services.ErrBuildingMaxLevel:
			http.Error(w, "El edificio ya está en su nivel máximo", http.StatusBadRequest)
		case services.ErrBuildingUpgrading:
			http.Error(w, "El edificio ya está siendo mejorado", http.StatusBadRequest)
		case services.ErrInvalidBuildingType:
			http.Error(w, "Tipo de edificio inválido", http.StatusBadRequest)
		case services.ErrTownHallRequired:
			http.Error(w, "Se requiere un ayuntamiento de nivel superior", http.StatusBadRequest)
		default:
			h.logger.Error("Error mejorando edificio", zap.Error(err))
			http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Edificio mejorado exitosamente",
		"result":  result,
	})
}

// GetBuildingUpgradeInfo obtiene información sobre la mejora de un edificio
func (h *VillageHandler) GetBuildingUpgradeInfo(w http.ResponseWriter, r *http.Request) {
	// Obtener el ID de la aldea de la URL
	villageIDStr := chi.URLParam(r, "villageID")
	villageID, err := uuid.Parse(villageIDStr)
	if err != nil {
		http.Error(w, "ID de aldea inválido", http.StatusBadRequest)
		return
	}

	// Obtener el tipo de edificio de la URL
	buildingType := chi.URLParam(r, "buildingType")
	if buildingType == "" {
		http.Error(w, "Tipo de edificio requerido", http.StatusBadRequest)
		return
	}

	// Verificar permisos
	village, err := h.villageRepo.GetVillageByID(villageID)
	if err != nil {
		h.logger.Error("Error obteniendo aldea", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}
	if village == nil {
		http.Error(w, "Aldea no encontrada", http.StatusNotFound)
		return
	}

	playerIDStr := r.Context().Value("player_id").(string)
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		h.logger.Error("Error parseando ID de jugador", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}
	if village.Village.PlayerID != playerID {
		http.Error(w, "No tienes permiso para ver esta aldea", http.StatusForbidden)
		return
	}

	// Obtener información de mejora usando el servicio avanzado
	upgradeInfo, err := h.constructionService.GetUpgradeInfo(villageID, buildingType)
	if err != nil {
		h.logger.Error("Error obteniendo información de mejora", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(upgradeInfo)
}

// CheckBuildingRequirements verifica los requisitos para construir un edificio
func (h *VillageHandler) CheckBuildingRequirements(w http.ResponseWriter, r *http.Request) {
	// Obtener el ID de la aldea de la URL
	villageIDStr := chi.URLParam(r, "villageID")
	villageID, err := uuid.Parse(villageIDStr)
	if err != nil {
		http.Error(w, "ID de aldea inválido", http.StatusBadRequest)
		return
	}

	// Obtener el tipo de edificio y nivel objetivo de los query parameters
	buildingType := r.URL.Query().Get("buildingType")
	if buildingType == "" {
		http.Error(w, "Tipo de edificio requerido", http.StatusBadRequest)
		return
	}

	targetLevelStr := r.URL.Query().Get("targetLevel")
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
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}
	if village == nil {
		http.Error(w, "Aldea no encontrada", http.StatusNotFound)
		return
	}

	playerIDStr := r.Context().Value("player_id").(string)
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		h.logger.Error("Error parseando ID de jugador", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}
	if village.Village.PlayerID != playerID {
		http.Error(w, "No tienes permiso para ver esta aldea", http.StatusForbidden)
		return
	}

	// Verificar requisitos usando el servicio avanzado
	requirements, err := h.constructionService.CheckBuildingRequirements(villageID, buildingType, targetLevel)
	if err != nil {
		h.logger.Error("Error verificando requisitos", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(requirements)
}

// ProcessConstructionQueue procesa la cola de construcción de una aldea
func (h *VillageHandler) ProcessConstructionQueue(w http.ResponseWriter, r *http.Request) {
	// Obtener el ID de la aldea de la URL
	villageIDStr := chi.URLParam(r, "villageID")
	villageID, err := uuid.Parse(villageIDStr)
	if err != nil {
		http.Error(w, "ID de aldea inválido", http.StatusBadRequest)
		return
	}

	// Verificar permisos
	village, err := h.villageRepo.GetVillageByID(villageID)
	if err != nil {
		h.logger.Error("Error obteniendo aldea", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}
	if village == nil {
		http.Error(w, "Aldea no encontrada", http.StatusNotFound)
		return
	}

	playerIDStr := r.Context().Value("player_id").(string)
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		h.logger.Error("Error parseando ID de jugador", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}
	if village.Village.PlayerID != playerID {
		http.Error(w, "No tienes permiso para procesar esta aldea", http.StatusForbidden)
		return
	}

	// Procesar cola de construcción usando el servicio avanzado
	results, err := h.constructionService.ProcessConstructionQueue(villageID)
	if err != nil {
		h.logger.Error("Error procesando cola de construcción", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Cola de construcción procesada",
		"results": results,
	})
}

// GetConstructionQueue obtiene la cola de construcción de una aldea
func (h *VillageHandler) GetConstructionQueue(w http.ResponseWriter, r *http.Request) {
	// Obtener el ID de la aldea de la URL
	villageIDStr := chi.URLParam(r, "villageID")
	villageID, err := uuid.Parse(villageIDStr)
	if err != nil {
		http.Error(w, "ID de aldea inválido", http.StatusBadRequest)
		return
	}

	// Verificar permisos
	village, err := h.villageRepo.GetVillageByID(villageID)
	if err != nil {
		h.logger.Error("Error obteniendo aldea", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}
	if village == nil {
		http.Error(w, "Aldea no encontrada", http.StatusNotFound)
		return
	}

	playerIDStr := r.Context().Value("player_id").(string)
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		h.logger.Error("Error parseando ID de jugador", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}
	if village.Village.PlayerID != playerID {
		http.Error(w, "No tienes permiso para ver esta aldea", http.StatusForbidden)
		return
	}

	// Obtener cola de construcción
	queue, err := h.constructionService.GetConstructionQueue(villageID)
	if err != nil {
		h.logger.Error("Error obteniendo cola de construcción", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(queue)
}

// CancelUpgrade cancela la mejora de un edificio
func (h *VillageHandler) CancelUpgrade(w http.ResponseWriter, r *http.Request) {
	// Obtener el ID de la aldea de la URL
	villageIDStr := chi.URLParam(r, "villageID")
	villageID, err := uuid.Parse(villageIDStr)
	if err != nil {
		http.Error(w, "ID de aldea inválido", http.StatusBadRequest)
		return
	}

	// Obtener el tipo de edificio de la URL
	buildingType := chi.URLParam(r, "buildingType")
	if buildingType == "" {
		http.Error(w, "Tipo de edificio requerido", http.StatusBadRequest)
		return
	}

	// Verificar permisos
	village, err := h.villageRepo.GetVillageByID(villageID)
	if err != nil {
		h.logger.Error("Error obteniendo aldea", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}
	if village == nil {
		http.Error(w, "Aldea no encontrada", http.StatusNotFound)
		return
	}

	playerIDStr := r.Context().Value("player_id").(string)
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		h.logger.Error("Error parseando ID de jugador", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}
	if village.Village.PlayerID != playerID {
		http.Error(w, "No tienes permiso para cancelar en esta aldea", http.StatusForbidden)
		return
	}

	// Cancelar mejora usando el servicio avanzado
	err = h.constructionService.CancelUpgrade(villageID, buildingType)
	if err != nil {
		h.logger.Error("Error cancelando mejora", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Mejora cancelada exitosamente",
	})
}

// CompleteBuildingUpgrade completa la mejora de un edificio
func (h *VillageHandler) CompleteBuildingUpgrade(w http.ResponseWriter, r *http.Request) {
	// Obtener el ID de la aldea de la URL
	villageIDStr := chi.URLParam(r, "villageID")
	villageID, err := uuid.Parse(villageIDStr)
	if err != nil {
		http.Error(w, "ID de aldea inválido", http.StatusBadRequest)
		return
	}

	// Obtener el tipo de edificio de la URL
	buildingType := chi.URLParam(r, "buildingType")
	if buildingType == "" {
		http.Error(w, "Tipo de edificio requerido", http.StatusBadRequest)
		return
	}

	// Verificar permisos
	village, err := h.villageRepo.GetVillageByID(villageID)
	if err != nil {
		h.logger.Error("Error obteniendo aldea", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}
	if village == nil {
		http.Error(w, "Aldea no encontrada", http.StatusNotFound)
		return
	}

	playerIDStr := r.Context().Value("player_id").(string)
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		h.logger.Error("Error parseando ID de jugador", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}
	if village.Village.PlayerID != playerID {
		http.Error(w, "No tienes permiso para acceder a esta aldea", http.StatusForbidden)
		return
	}

	// Completar la mejora
	err = h.constructionService.CompleteUpgrade(villageID, buildingType)
	if err != nil {
		switch err.Error() {
		case "el edificio no está siendo mejorado":
			http.Error(w, "El edificio no está siendo mejorado", http.StatusBadRequest)
		case "la mejora aún no ha terminado":
			http.Error(w, "La mejora aún no ha terminado", http.StatusBadRequest)
		case services.ErrInvalidBuildingType.Error():
			http.Error(w, "Tipo de edificio inválido", http.StatusBadRequest)
		default:
			h.logger.Error("Error completando mejora", zap.Error(err))
			http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Mejora completada exitosamente",
	})
}

// GetBuildingUpgradeTimeRemaining obtiene el tiempo restante de mejora en tiempo real
func (h *VillageHandler) GetBuildingUpgradeTimeRemaining(w http.ResponseWriter, r *http.Request) {
	// Obtener el ID de la aldea de la URL
	villageIDStr := chi.URLParam(r, "villageID")
	villageID, err := uuid.Parse(villageIDStr)
	if err != nil {
		http.Error(w, "ID de aldea inválido", http.StatusBadRequest)
		return
	}

	// Obtener el tipo de edificio de la URL
	buildingType := chi.URLParam(r, "buildingType")
	if buildingType == "" {
		http.Error(w, "Tipo de edificio requerido", http.StatusBadRequest)
		return
	}

	// Verificar permisos
	village, err := h.villageRepo.GetVillageByID(villageID)
	if err != nil {
		h.logger.Error("Error obteniendo aldea", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}
	if village == nil {
		http.Error(w, "Aldea no encontrada", http.StatusNotFound)
		return
	}

	playerIDStr := r.Context().Value("player_id").(string)
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		h.logger.Error("Error parseando ID de jugador", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}
	if village.Village.PlayerID != playerID {
		http.Error(w, "No tienes permiso para acceder a esta aldea", http.StatusForbidden)
		return
	}

	// Obtener información del edificio
	building, exists := village.Buildings[buildingType]
	if !exists {
		http.Error(w, "Edificio no encontrado", http.StatusNotFound)
		return
	}

	// Verificar si está mejorándose
	if !building.IsUpgrading || building.UpgradeCompletionTime == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
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
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"building_type":   buildingType,
		"is_upgrading":    true,
		"time_remaining":  int64(timeRemaining.Seconds()),
		"completion_time": building.UpgradeCompletionTime,
		"formatted_time":  formattedTime,
		"can_complete":    false,
	})
}
