package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"server-backend/models"
	"server-backend/repository"
	"server-backend/services"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type BattleHandler struct {
	battleRepo    *repository.BattleRepository
	villageRepo   *repository.VillageRepository
	unitRepo      *repository.UnitRepository
	battleService *services.BattleService
	logger        *zap.Logger
}

func NewBattleHandler(
	battleRepo *repository.BattleRepository,
	villageRepo *repository.VillageRepository,
	unitRepo *repository.UnitRepository,
	logger *zap.Logger,
	redisService *services.RedisService,
) *BattleHandler {
	battleService := services.NewBattleService(battleRepo, villageRepo, unitRepo, logger, redisService)
	return &BattleHandler{
		battleRepo:    battleRepo,
		villageRepo:   villageRepo,
		unitRepo:      unitRepo,
		battleService: battleService,
		logger:        logger,
	}
}

// AttackVillage inicia un ataque a una aldea
func (h *BattleHandler) AttackVillage(w http.ResponseWriter, r *http.Request) {
	var request models.BattleRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.logger.Error("Error decodificando request", zap.Error(err))
		http.Error(w, "Error decodificando request", http.StatusBadRequest)
		return
	}

	// Obtener playerID del contexto (asumiendo que está autenticado)
	playerIDStr := r.Context().Value("player_id").(string)
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		http.Error(w, "ID de jugador inválido", http.StatusBadRequest)
		return
	}

	request.AttackerID = playerID

	// Crear la batalla usando el servicio
	battle, err := h.battleService.CreateBattle(&request)
	if err != nil {
		h.logger.Error("Error creando batalla", zap.Error(err))
		http.Error(w, "Error creando batalla: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Batalla creada exitosamente",
		"data":    battle,
	})
}

// GetBattle obtiene una batalla específica
func (h *BattleHandler) GetBattle(w http.ResponseWriter, r *http.Request) {
	battleIDStr := chi.URLParam(r, "battleID")
	battleID, err := uuid.Parse(battleIDStr)
	if err != nil {
		http.Error(w, "ID de batalla inválido", http.StatusBadRequest)
		return
	}

	battle, err := h.battleRepo.GetBattle(battleID)
	if err != nil {
		h.logger.Error("Error obteniendo batalla", zap.Error(err))
		http.Error(w, "Batalla no encontrada", http.StatusNotFound)
		return
	}

	// Verificar autorización
	playerIDStr := r.Context().Value("player_id").(string)
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		http.Error(w, "ID de jugador inválido", http.StatusBadRequest)
		return
	}

	if battle.AttackerID != playerID && battle.DefenderID != playerID {
		http.Error(w, "No autorizado", http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    battle,
	})
}

// GetBattleWithDetails obtiene una batalla con todos sus detalles
func (h *BattleHandler) GetBattleWithDetails(w http.ResponseWriter, r *http.Request) {
	battleIDStr := chi.URLParam(r, "battleID")
	battleID, err := uuid.Parse(battleIDStr)
	if err != nil {
		http.Error(w, "ID de batalla inválido", http.StatusBadRequest)
		return
	}

	// Verificar autorización
	playerIDStr := r.Context().Value("player_id").(string)
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		http.Error(w, "ID de jugador inválido", http.StatusBadRequest)
		return
	}

	battle, err := h.battleRepo.GetBattle(battleID)
	if err != nil {
		h.logger.Error("Error obteniendo batalla", zap.Error(err))
		http.Error(w, "Batalla no encontrada", http.StatusNotFound)
		return
	}

	if battle.AttackerID != playerID && battle.DefenderID != playerID {
		http.Error(w, "No autorizado", http.StatusForbidden)
		return
	}

	details, err := h.battleService.GetBattleWithDetails(battleID)
	if err != nil {
		h.logger.Error("Error obteniendo detalles de batalla", zap.Error(err))
		http.Error(w, "Error obteniendo detalles de batalla", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    details,
	})
}

// GetBattleWaves obtiene las oleadas de una batalla
func (h *BattleHandler) GetBattleWaves(w http.ResponseWriter, r *http.Request) {
	battleIDStr := chi.URLParam(r, "battleID")
	battleID, err := uuid.Parse(battleIDStr)
	if err != nil {
		http.Error(w, "ID de batalla inválido", http.StatusBadRequest)
		return
	}

	// Verificar autorización
	playerIDStr := r.Context().Value("player_id").(string)
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		http.Error(w, "ID de jugador inválido", http.StatusBadRequest)
		return
	}

	battle, err := h.battleRepo.GetBattle(battleID)
	if err != nil {
		h.logger.Error("Error obteniendo batalla", zap.Error(err))
		http.Error(w, "Batalla no encontrada", http.StatusNotFound)
		return
	}

	if battle.AttackerID != playerID && battle.DefenderID != playerID {
		http.Error(w, "No autorizado", http.StatusForbidden)
		return
	}

	waves, err := h.battleRepo.GetBattleWaves(battleID)
	if err != nil {
		h.logger.Error("Error obteniendo oleadas", zap.Error(err))
		http.Error(w, "Error obteniendo oleadas", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    waves,
	})
}

// GetBattleRankings obtiene los rankings de batalla
func (h *BattleHandler) GetBattleRankings(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	rankings, err := h.battleRepo.GetBattleRankings(limit)
	if err != nil {
		h.logger.Error("Error obteniendo rankings", zap.Error(err))
		http.Error(w, "Error obteniendo rankings", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    rankings,
	})
}

// GetPlayerUnits obtiene las unidades del jugador
func (h *BattleHandler) GetPlayerUnits(w http.ResponseWriter, r *http.Request) {
	playerIDStr := r.Context().Value("player_id").(string)
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		http.Error(w, "ID de jugador inválido", http.StatusBadRequest)
		return
	}

	units, err := h.battleRepo.GetPlayerUnits(playerID)
	if err != nil {
		h.logger.Error("Error obteniendo unidades", zap.Error(err))
		http.Error(w, "Error obteniendo unidades", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    units,
	})
}

// TrainUnits entrena unidades
func (h *BattleHandler) TrainUnits(w http.ResponseWriter, r *http.Request) {
	playerIDStr := r.Context().Value("player_id").(string)
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		http.Error(w, "ID de jugador inválido", http.StatusBadRequest)
		return
	}

	var request struct {
		UnitID   string `json:"unit_id"`
		Quantity int    `json:"quantity"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.logger.Error("Error decodificando request", zap.Error(err))
		http.Error(w, "Error decodificando request", http.StatusBadRequest)
		return
	}

	unitID, err := uuid.Parse(request.UnitID)
	if err != nil {
		http.Error(w, "ID de unidad inválido", http.StatusBadRequest)
		return
	}

	if request.Quantity <= 0 {
		http.Error(w, "Cantidad debe ser mayor a 0", http.StatusBadRequest)
		return
	}

	if err := h.battleRepo.TrainUnits(playerID, unitID, request.Quantity); err != nil {
		h.logger.Error("Error entrenando unidades", zap.Error(err))
		http.Error(w, "Error entrenando unidades: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Unidades en entrenamiento",
	})
}

// GetMilitaryUnits obtiene las unidades militares disponibles
func (h *BattleHandler) GetMilitaryUnits(w http.ResponseWriter, r *http.Request) {
	unitType := r.URL.Query().Get("type")
	category := r.URL.Query().Get("category")

	units, err := h.battleRepo.GetMilitaryUnits(unitType, category)
	if err != nil {
		h.logger.Error("Error obteniendo unidades militares", zap.Error(err))
		http.Error(w, "Error obteniendo unidades militares", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    units,
	})
}

// GetPlayerBattles obtiene las batallas de un jugador
func (h *BattleHandler) GetPlayerBattles(w http.ResponseWriter, r *http.Request) {
	playerIDStr := r.Context().Value("player_id").(string)
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		http.Error(w, "ID de jugador inválido", http.StatusBadRequest)
		return
	}

	limit := 50 // Por defecto
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	battles, err := h.battleRepo.GetBattlesByPlayer(playerID, limit)
	if err != nil {
		h.logger.Error("Error obteniendo batallas del jugador", zap.Error(err))
		http.Error(w, "Error obteniendo batallas", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    battles,
	})
}

// GetActiveBattles obtiene las batallas activas
func (h *BattleHandler) GetActiveBattles(w http.ResponseWriter, r *http.Request) {
	limit := 20
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	battles, err := h.battleService.GetActiveBattles(limit)
	if err != nil {
		h.logger.Error("Error obteniendo batallas activas", zap.Error(err))
		http.Error(w, "Error obteniendo batallas activas", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    battles,
	})
}

// GetIncomingAttacks obtiene los ataques entrantes para un jugador
func (h *BattleHandler) GetIncomingAttacks(w http.ResponseWriter, r *http.Request) {
	playerIDStr := r.Context().Value("player_id").(string)
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		http.Error(w, "ID de jugador inválido", http.StatusBadRequest)
		return
	}

	limit := 20
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	// Obtener batallas donde el jugador es defensor
	battles, err := h.battleRepo.GetBattlesByPlayer(playerID, limit)
	if err != nil {
		h.logger.Error("Error obteniendo ataques entrantes", zap.Error(err))
		http.Error(w, "Error obteniendo ataques entrantes", http.StatusInternalServerError)
		return
	}

	// Filtrar solo las batallas donde es defensor
	incomingAttacks := []models.Battle{}
	for _, battle := range battles {
		if battle.DefenderID == playerID && (battle.Status == "pending" || battle.Status == "active") {
			incomingAttacks = append(incomingAttacks, battle)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    incomingAttacks,
	})
}

// GetOutgoingAttacks obtiene los ataques salientes de un jugador
func (h *BattleHandler) GetOutgoingAttacks(w http.ResponseWriter, r *http.Request) {
	playerIDStr := r.Context().Value("player_id").(string)
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		http.Error(w, "ID de jugador inválido", http.StatusBadRequest)
		return
	}

	limit := 20
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	// Obtener batallas donde el jugador es atacante
	battles, err := h.battleRepo.GetBattlesByPlayer(playerID, limit)
	if err != nil {
		h.logger.Error("Error obteniendo ataques salientes", zap.Error(err))
		http.Error(w, "Error obteniendo ataques salientes", http.StatusInternalServerError)
		return
	}

	// Filtrar solo las batallas donde es atacante
	outgoingAttacks := []models.Battle{}
	for _, battle := range battles {
		if battle.AttackerID == playerID && (battle.Status == "pending" || battle.Status == "active") {
			outgoingAttacks = append(outgoingAttacks, battle)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    outgoingAttacks,
	})
}

// GetBattleStatistics obtiene las estadísticas de batalla de un jugador
func (h *BattleHandler) GetBattleStatistics(w http.ResponseWriter, r *http.Request) {
	playerIDStr := r.Context().Value("player_id").(string)
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		http.Error(w, "ID de jugador inválido", http.StatusBadRequest)
		return
	}

	stats, err := h.battleService.GetBattleStatistics(playerID)
	if err != nil {
		h.logger.Error("Error obteniendo estadísticas", zap.Error(err))
		http.Error(w, "Error obteniendo estadísticas", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    stats,
	})
}

// GetBattleReport obtiene el reporte detallado de una batalla
func (h *BattleHandler) GetBattleReport(w http.ResponseWriter, r *http.Request) {
	battleIDStr := chi.URLParam(r, "battleID")
	battleID, err := uuid.Parse(battleIDStr)
	if err != nil {
		http.Error(w, "ID de batalla inválido", http.StatusBadRequest)
		return
	}

	// Verificar autorización
	playerIDStr := r.Context().Value("player_id").(string)
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		http.Error(w, "ID de jugador inválido", http.StatusBadRequest)
		return
	}

	battle, err := h.battleRepo.GetBattle(battleID)
	if err != nil {
		h.logger.Error("Error obteniendo batalla", zap.Error(err))
		http.Error(w, "Batalla no encontrada", http.StatusNotFound)
		return
	}

	if battle.AttackerID != playerID && battle.DefenderID != playerID {
		http.Error(w, "No autorizado", http.StatusForbidden)
		return
	}

	// Obtener reporte completo
	report, err := h.battleService.GetBattleWithDetails(battleID)
	if err != nil {
		h.logger.Error("Error obteniendo reporte", zap.Error(err))
		http.Error(w, "Error obteniendo reporte", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    report,
	})
}

// GetBattleLog obtiene el log detallado de una batalla
func (h *BattleHandler) GetBattleLog(w http.ResponseWriter, r *http.Request) {
	battleIDStr := chi.URLParam(r, "battleID")
	battleID, err := uuid.Parse(battleIDStr)
	if err != nil {
		http.Error(w, "ID de batalla inválido", http.StatusBadRequest)
		return
	}

	// Verificar autorización
	playerIDStr := r.Context().Value("player_id").(string)
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		http.Error(w, "ID de jugador inválido", http.StatusBadRequest)
		return
	}

	battle, err := h.battleRepo.GetBattle(battleID)
	if err != nil {
		h.logger.Error("Error obteniendo batalla", zap.Error(err))
		http.Error(w, "Batalla no encontrada", http.StatusNotFound)
		return
	}

	if battle.AttackerID != playerID && battle.DefenderID != playerID {
		http.Error(w, "No autorizado", http.StatusForbidden)
		return
	}

	// Obtener oleadas (que contienen el log)
	waves, err := h.battleRepo.GetBattleWaves(battleID)
	if err != nil {
		h.logger.Error("Error obteniendo log de batalla", zap.Error(err))
		http.Error(w, "Error obteniendo log de batalla", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"battle_id": battleID,
			"waves":     waves,
		},
	})
}

// CancelBattle cancela una batalla pendiente
func (h *BattleHandler) CancelBattle(w http.ResponseWriter, r *http.Request) {
	battleIDStr := chi.URLParam(r, "battleID")
	battleID, err := uuid.Parse(battleIDStr)
	if err != nil {
		http.Error(w, "ID de batalla inválido", http.StatusBadRequest)
		return
	}

	playerIDStr := r.Context().Value("player_id").(string)
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		http.Error(w, "ID de jugador inválido", http.StatusBadRequest)
		return
	}

	if err := h.battleService.CancelBattle(battleID, playerID); err != nil {
		h.logger.Error("Error cancelando batalla", zap.Error(err))
		http.Error(w, "Error cancelando batalla: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Batalla cancelada exitosamente",
	})
}

// ProcessBattle procesa una batalla (solo para administradores o sistema)
func (h *BattleHandler) ProcessBattle(w http.ResponseWriter, r *http.Request) {
	battleIDStr := chi.URLParam(r, "battleID")
	battleID, err := uuid.Parse(battleIDStr)
	if err != nil {
		http.Error(w, "ID de batalla inválido", http.StatusBadRequest)
		return
	}

	// TODO: Verificar permisos de administrador
	// Por ahora, permitir procesamiento

	if err := h.battleService.ProcessBattle(battleID); err != nil {
		h.logger.Error("Error procesando batalla", zap.Error(err))
		http.Error(w, "Error procesando batalla: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Batalla procesada exitosamente",
	})
}

// GetVillageDefenses obtiene las defensas de una aldea
func (h *BattleHandler) GetVillageDefenses(w http.ResponseWriter, r *http.Request) {
	villageIDStr := r.URL.Query().Get("village_id")
	if villageIDStr == "" {
		http.Error(w, "ID de aldea requerido", http.StatusBadRequest)
		return
	}

	villageID, err := uuid.Parse(villageIDStr)
	if err != nil {
		http.Error(w, "ID de aldea inválido", http.StatusBadRequest)
		return
	}

	// Obtener unidades de la aldea (defensas)
	// TODO: Implementar cuando se conecte con el sistema de aldeas
	defenses := map[string]interface{}{
		"village_id": villageID,
		"units":      []interface{}{},
		"buildings":  []interface{}{},
		"traps":      []interface{}{},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    defenses,
	})
}
