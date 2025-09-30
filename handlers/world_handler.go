package handlers

import (
	"encoding/json"
	"net/http"
	"server-backend/repository"
	"server-backend/services"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type WorldHandler struct {
	worldService *services.WorldService
	worldRepo    *repository.WorldRepository
	villageRepo  *repository.VillageRepository
	logger       *zap.Logger
}

func NewWorldHandler(worldRepo *repository.WorldRepository, playerRepo *repository.PlayerRepository, villageRepo *repository.VillageRepository, allianceRepo *repository.AllianceRepository, battleRepo *repository.BattleRepository, economyRepo *repository.EconomyRepository, logger *zap.Logger) *WorldHandler {
	worldService := services.NewWorldService(worldRepo, playerRepo, villageRepo, allianceRepo, battleRepo, economyRepo, logger)
	return &WorldHandler{
		worldService: worldService,
		worldRepo:    worldRepo,
		villageRepo:  villageRepo,
		logger:       logger,
	}
}

// GetWorlds obtiene todos los mundos disponibles
func (h *WorldHandler) GetWorlds(w http.ResponseWriter, r *http.Request) {
	worlds, err := h.worldService.GetAvailableWorlds()
	if err != nil {
		h.logger.Error("Error obteniendo mundos", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(worlds)
}

// GetWorld obtiene información de un mundo específico
func (h *WorldHandler) GetWorld(w http.ResponseWriter, r *http.Request) {
	worldIDStr := chi.URLParam(r, "worldID")
	worldID, err := uuid.Parse(worldIDStr)
	if err != nil {
		http.Error(w, "ID de mundo inválido", http.StatusBadRequest)
		return
	}

	world, err := h.worldService.GetWorldDetails(worldID)
	if err != nil {
		h.logger.Error("Error obteniendo mundo", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}
	if world == nil {
		http.Error(w, "Mundo no encontrado", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(world)
}

// AssignToWorld asigna automáticamente al jugador al mundo menos poblado
func (h *WorldHandler) AssignToWorld(w http.ResponseWriter, r *http.Request) {
	// TODO: Implementar asignación automática
	http.Error(w, "Asignación automática no implementada", http.StatusNotImplemented)
}

// JoinWorld permite al jugador unirse a un mundo específico
func (h *WorldHandler) JoinWorld(w http.ResponseWriter, r *http.Request) {
	// Obtener el ID del mundo de la URL
	worldIDStr := chi.URLParam(r, "worldID")
	worldID, err := uuid.Parse(worldIDStr)
	if err != nil {
		http.Error(w, "ID de mundo inválido", http.StatusBadRequest)
		return
	}

	// Obtener el ID del jugador del contexto
	playerIDStr := r.Context().Value("player_id").(string)
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		h.logger.Error("Error parseando ID de jugador", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	// Unir al jugador al mundo específico
	response, err := h.worldService.JoinWorld(playerID, worldID, "Mi Aldea", "random")
	if err != nil {
		switch err {
		case services.ErrNoWorldsAvailable:
			http.Error(w, "No hay mundos disponibles", http.StatusServiceUnavailable)
		case services.ErrWorldFull:
			http.Error(w, "El mundo está lleno", http.StatusServiceUnavailable)
		case services.ErrPlayerAlreadyInWorld:
			http.Error(w, "Ya tienes una aldea en un mundo", http.StatusConflict)
		default:
			h.logger.Error("Error uniendo jugador a mundo", zap.Error(err))
			http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// GetPlayerVillages obtiene las aldeas del jugador
func (h *WorldHandler) GetPlayerVillages(w http.ResponseWriter, r *http.Request) {
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
