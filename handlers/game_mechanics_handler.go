package handlers

import (
	"encoding/json"
	"net/http"
	"server-backend/services"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// GameMechanicsHandler maneja las funciones avanzadas del juego
type GameMechanicsHandler struct {
	gameMechanicsService *services.GameMechanicsService
	logger               *zap.Logger
}

// NewGameMechanicsHandler crea un nuevo handler para las funciones avanzadas del juego
func NewGameMechanicsHandler(gameMechanicsService *services.GameMechanicsService, logger *zap.Logger) *GameMechanicsHandler {
	return &GameMechanicsHandler{
		gameMechanicsService: gameMechanicsService,
		logger:               logger,
	}
}

// CalculateResourceProduction calcula la producción de recursos de una aldea
func (h *GameMechanicsHandler) CalculateResourceProduction(w http.ResponseWriter, r *http.Request) {
	// Obtener el ID de la aldea de la URL
	villageIDStr := chi.URLParam(r, "villageID")
	villageID, err := uuid.Parse(villageIDStr)
	if err != nil {
		http.Error(w, "ID de aldea inválido", http.StatusBadRequest)
		return
	}

	// Verificar permisos (esto debería venir del middleware de autenticación)
	playerIDStr := r.Context().Value("player_id").(string)
	if playerIDStr == "" {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	// Calcular producción de recursos
	production, err := h.gameMechanicsService.CalculateResourceProduction(villageID)
	if err != nil {
		h.logger.Error("Error calculando producción de recursos", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(production)
}

// UpdateVillageResources actualiza los recursos de una aldea
func (h *GameMechanicsHandler) UpdateVillageResources(w http.ResponseWriter, r *http.Request) {
	// Obtener el ID de la aldea de la URL
	villageIDStr := chi.URLParam(r, "villageID")
	villageID, err := uuid.Parse(villageIDStr)
	if err != nil {
		http.Error(w, "ID de aldea inválido", http.StatusBadRequest)
		return
	}

	// Verificar permisos
	playerIDStr := r.Context().Value("player_id").(string)
	if playerIDStr == "" {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	// Actualizar recursos de la aldea
	err = h.gameMechanicsService.UpdateVillageResources(villageID)
	if err != nil {
		h.logger.Error("Error actualizando recursos de la aldea", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Recursos de la aldea actualizados exitosamente",
	})
}

// CalculateBattleOutcome calcula el resultado de una batalla
func (h *GameMechanicsHandler) CalculateBattleOutcome(w http.ResponseWriter, r *http.Request) {
	// Verificar permisos
	playerIDStr := r.Context().Value("player_id").(string)
	if playerIDStr == "" {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	// Parsear el cuerpo de la petición
	var battleRequest struct {
		AttackerUnits        json.RawMessage `json:"attacker_units"`
		DefenderUnits        json.RawMessage `json:"defender_units"`
		AttackerHeroes       json.RawMessage `json:"attacker_heroes"`
		DefenderHeroes       json.RawMessage `json:"defender_heroes"`
		Terrain              string          `json:"terrain"`
		Weather              string          `json:"weather"`
		AttackerTechnologies json.RawMessage `json:"attacker_technologies"`
		DefenderTechnologies json.RawMessage `json:"defender_technologies"`
	}

	if err := json.NewDecoder(r.Body).Decode(&battleRequest); err != nil {
		http.Error(w, "Cuerpo de petición inválido", http.StatusBadRequest)
		return
	}

	// Calcular resultado de la batalla
	result, err := h.gameMechanicsService.CalculateBattleOutcome(
		battleRequest.AttackerUnits,
		battleRequest.DefenderUnits,
		battleRequest.AttackerHeroes,
		battleRequest.DefenderHeroes,
		battleRequest.Terrain,
		battleRequest.Weather,
		battleRequest.AttackerTechnologies,
		battleRequest.DefenderTechnologies,
	)
	if err != nil {
		h.logger.Error("Error calculando resultado de batalla", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// CalculateTradeRates calcula las tasas de intercambio
func (h *GameMechanicsHandler) CalculateTradeRates(w http.ResponseWriter, r *http.Request) {
	// Verificar permisos
	playerIDStr := r.Context().Value("player_id").(string)
	if playerIDStr == "" {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	// Obtener parámetros de la query
	resourceType := r.URL.Query().Get("resourceType")
	if resourceType == "" {
		http.Error(w, "Tipo de recurso requerido", http.StatusBadRequest)
		return
	}

	worldIDStr := r.URL.Query().Get("worldID")
	var worldID *uuid.UUID
	if worldIDStr != "" {
		if id, err := uuid.Parse(worldIDStr); err == nil {
			worldID = &id
		}
	}

	// Calcular tasas de intercambio
	result, err := h.gameMechanicsService.CalculateTradeRates(resourceType, worldID)
	if err != nil {
		h.logger.Error("Error calculando tasas de intercambio", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// ProcessAllianceBenefits procesa los beneficios de una alianza
func (h *GameMechanicsHandler) ProcessAllianceBenefits(w http.ResponseWriter, r *http.Request) {
	// Verificar permisos
	playerIDStr := r.Context().Value("player_id").(string)
	if playerIDStr == "" {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	// Obtener el ID de la alianza de la URL
	allianceIDStr := chi.URLParam(r, "allianceID")
	allianceID, err := uuid.Parse(allianceIDStr)
	if err != nil {
		http.Error(w, "ID de alianza inválido", http.StatusBadRequest)
		return
	}

	// Procesar beneficios de la alianza
	results, err := h.gameMechanicsService.ProcessAllianceBenefits(allianceID)
	if err != nil {
		h.logger.Error("Error procesando beneficios de alianza", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// CalculatePlayerScore calcula la puntuación del jugador
func (h *GameMechanicsHandler) CalculatePlayerScore(w http.ResponseWriter, r *http.Request) {
	// Verificar permisos
	playerIDStr := r.Context().Value("player_id").(string)
	if playerIDStr == "" {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	// Obtener el ID del jugador de la URL o usar el del contexto
	playerID := playerIDStr
	if playerIDFromURL := chi.URLParam(r, "playerID"); playerIDFromURL != "" {
		playerID = playerIDFromURL
	}

	playerUUID, err := uuid.Parse(playerID)
	if err != nil {
		http.Error(w, "ID de jugador inválido", http.StatusBadRequest)
		return
	}

	// Calcular puntuación del jugador
	result, err := h.gameMechanicsService.CalculatePlayerScore(playerUUID)
	if err != nil {
		h.logger.Error("Error calculando puntuación del jugador", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// GenerateDailyRewards genera recompensas diarias
func (h *GameMechanicsHandler) GenerateDailyRewards(w http.ResponseWriter, r *http.Request) {
	// Verificar permisos
	playerIDStr := r.Context().Value("player_id").(string)
	if playerIDStr == "" {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		http.Error(w, "ID de jugador inválido", http.StatusBadRequest)
		return
	}

	// Generar recompensas diarias
	result, err := h.gameMechanicsService.GenerateDailyRewards(playerID)
	if err != nil {
		h.logger.Error("Error generando recompensas diarias", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// CleanupInactiveData limpia datos inactivos (solo para administradores)
func (h *GameMechanicsHandler) CleanupInactiveData(w http.ResponseWriter, r *http.Request) {
	// Verificar permisos de administrador
	playerIDStr := r.Context().Value("player_id").(string)
	if playerIDStr == "" {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	// Verificar si es administrador (esto debería venir del middleware)
	isAdmin := r.Context().Value("is_admin").(bool)
	if !isAdmin {
		http.Error(w, "Acceso denegado", http.StatusForbidden)
		return
	}

	// Obtener días de antigüedad de los query parameters
	daysOldStr := r.URL.Query().Get("daysOld")
	daysOld := 30 // Por defecto
	if daysOldStr != "" {
		if days, err := strconv.Atoi(daysOldStr); err == nil && days > 0 {
			daysOld = days
		}
	}

	// Limpiar datos inactivos
	results, err := h.gameMechanicsService.CleanupInactiveData(daysOld)
	if err != nil {
		h.logger.Error("Error limpiando datos inactivos", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":  "Limpieza de datos completada",
		"days_old": daysOld,
		"results":  results,
	})
}
