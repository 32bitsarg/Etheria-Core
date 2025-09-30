package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"server-backend/models"
	"server-backend/repository"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type AdminHandler struct {
	playerRepo      *repository.PlayerRepository
	worldRepo       *repository.WorldRepository
	villageRepo     *repository.VillageRepository
	battleRepo      *repository.BattleRepository
	economyRepo     *repository.EconomyRepository
	allianceRepo    *repository.AllianceRepository
	achievementRepo *repository.AchievementRepository
	eventRepo       *repository.EventRepository
	currencyRepo    *repository.CurrencyRepository
	logger          *zap.Logger
}

func NewAdminHandler(
	playerRepo *repository.PlayerRepository,
	worldRepo *repository.WorldRepository,
	villageRepo *repository.VillageRepository,
	battleRepo *repository.BattleRepository,
	economyRepo *repository.EconomyRepository,
	allianceRepo *repository.AllianceRepository,
	achievementRepo *repository.AchievementRepository,
	eventRepo *repository.EventRepository,
	currencyRepo *repository.CurrencyRepository,
	logger *zap.Logger,
) *AdminHandler {
	return &AdminHandler{
		playerRepo:      playerRepo,
		worldRepo:       worldRepo,
		villageRepo:     villageRepo,
		battleRepo:      battleRepo,
		economyRepo:     economyRepo,
		allianceRepo:    allianceRepo,
		achievementRepo: achievementRepo,
		eventRepo:       eventRepo,
		currencyRepo:    currencyRepo,
		logger:          logger,
	}
}

// ServerStats representa las estadísticas del servidor
type ServerStats struct {
	TotalPlayers        int     `json:"totalPlayers"`
	ActivePlayers       int     `json:"activePlayers"`
	TotalWorlds         int     `json:"totalWorlds"`
	ActiveWorlds        int     `json:"activeWorlds"`
	TotalBattles        int     `json:"totalBattles"`
	BattlesToday        int     `json:"battlesToday"`
	TotalTrades         int     `json:"totalTrades"`
	TradesToday         int     `json:"tradesToday"`
	ServerUptime        int64   `json:"serverUptime"`
	CPUUsage            float64 `json:"cpuUsage"`
	MemoryUsage         float64 `json:"memoryUsage"`
	DatabaseConnections int     `json:"databaseConnections"`
}

// GetServerStats obtiene las estadísticas del servidor
func (h *AdminHandler) GetServerStats(w http.ResponseWriter, r *http.Request) {
	// Obtener estadísticas básicas
	totalPlayers, err := h.playerRepo.GetTotalPlayers()
	if err != nil {
		h.logger.Error("Error obteniendo total de jugadores", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	activePlayers, err := h.playerRepo.GetActivePlayers()
	if err != nil {
		h.logger.Error("Error obteniendo jugadores activos", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	totalWorlds, err := h.worldRepo.GetTotalWorlds()
	if err != nil {
		h.logger.Error("Error obteniendo total de mundos", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	activeWorlds, err := h.worldRepo.GetActiveWorlds()
	if err != nil {
		h.logger.Error("Error obteniendo mundos activos", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	totalBattles, err := h.battleRepo.GetTotalBattles()
	if err != nil {
		h.logger.Error("Error obteniendo total de batallas", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	battlesToday, err := h.battleRepo.GetBattlesToday()
	if err != nil {
		h.logger.Error("Error obteniendo batallas de hoy", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	totalTrades, err := h.economyRepo.GetTotalTrades()
	if err != nil {
		h.logger.Error("Error obteniendo total de transacciones", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	tradesToday, err := h.economyRepo.GetTradesToday()
	if err != nil {
		h.logger.Error("Error obteniendo transacciones de hoy", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	// Crear respuesta con estadísticas
	stats := ServerStats{
		TotalPlayers:        totalPlayers,
		ActivePlayers:       activePlayers,
		TotalWorlds:         totalWorlds,
		ActiveWorlds:        activeWorlds,
		TotalBattles:        totalBattles,
		BattlesToday:        battlesToday,
		TotalTrades:         totalTrades,
		TradesToday:         tradesToday,
		ServerUptime:        time.Now().Unix(), // Placeholder - implementar lógica real
		CPUUsage:            25.5,              // Placeholder - implementar monitoreo real
		MemoryUsage:         45.2,              // Placeholder - implementar monitoreo real
		DatabaseConnections: 10,                // Placeholder - implementar monitoreo real
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// GetPlayers obtiene la lista de jugadores para el dashboard
func (h *AdminHandler) GetPlayers(w http.ResponseWriter, r *http.Request) {
	// Obtener jugadores con datos completos
	players, err := h.playerRepo.GetAllPlayersForAdmin()
	if err != nil {
		h.logger.Error("Error obteniendo jugadores", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	// Obtener datos relacionados para cada jugador
	for _, player := range players {
		// Obtener aldeas del jugador
		villages, err := h.playerRepo.GetPlayerVillages(player.ID)
		if err != nil {
			h.logger.Error("Error obteniendo aldeas del jugador",
				zap.Error(err),
				zap.String("player_id", player.ID.String()))
			continue
		}
		// Convertir []Village a []VillageWithDetails vacío (solo con el campo Village)
		var villagesWithDetails []models.VillageWithDetails
		for _, v := range villages {
			villagesWithDetails = append(villagesWithDetails, models.VillageWithDetails{Village: v})
		}
		player.Villages = villagesWithDetails

		// Obtener logros del jugador
		achievements, err := h.playerRepo.GetPlayerAchievements(player.ID)
		if err != nil {
			h.logger.Error("Error obteniendo logros del jugador",
				zap.Error(err),
				zap.String("player_id", player.ID.String()))
			continue
		}
		player.Achievements = achievements

		// Obtener títulos del jugador
		titles, err := h.playerRepo.GetPlayerTitles(player.ID)
		if err != nil {
			h.logger.Error("Error obteniendo títulos del jugador",
				zap.Error(err),
				zap.String("player_id", player.ID.String()))
			continue
		}
		player.Titles = titles
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(players)
}

// GetWorlds obtiene la lista de mundos para el dashboard
func (h *AdminHandler) GetWorlds(w http.ResponseWriter, r *http.Request) {
	worlds, err := h.worldRepo.GetAllWorlds()
	if err != nil {
		h.logger.Error("Error obteniendo mundos", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(worlds)
}

// GetVillages obtiene la lista de aldeas para el dashboard
func (h *AdminHandler) GetVillages(w http.ResponseWriter, r *http.Request) {
	villages, err := h.villageRepo.GetAllVillages()
	if err != nil {
		h.logger.Error("Error obteniendo aldeas", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(villages)
}

// GetBattles obtiene la lista de batallas para el dashboard
func (h *AdminHandler) GetBattles(w http.ResponseWriter, r *http.Request) {
	battles, err := h.battleRepo.GetAllBattles()
	if err != nil {
		h.logger.Error("Error obteniendo batallas", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(battles)
}

// GetAlliances obtiene la lista de alianzas para el dashboard
func (h *AdminHandler) GetAlliances(w http.ResponseWriter, r *http.Request) {
	alliances, err := h.allianceRepo.GetAlliances()
	if err != nil {
		h.logger.Error("Error obteniendo alianzas", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(alliances)
}

// GetEvents obtiene la lista de eventos para el dashboard
func (h *AdminHandler) GetEvents(w http.ResponseWriter, r *http.Request) {
	events, err := h.eventRepo.GetAllEvents()
	if err != nil {
		h.logger.Error("Error obteniendo eventos", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}

// CreateWorld crea un nuevo mundo
func (h *AdminHandler) CreateWorld(w http.ResponseWriter, r *http.Request) {
	var req models.CreateWorldRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Error decodificando solicitud de creación de mundo", zap.Error(err))
		http.Error(w, "Error decodificando la solicitud", http.StatusBadRequest)
		return
	}

	// Validar datos de entrada
	if req.Name == "" {
		http.Error(w, "El nombre del mundo es requerido", http.StatusBadRequest)
		return
	}
	if req.MaxPlayers <= 0 {
		http.Error(w, "El número máximo de jugadores debe ser mayor a 0", http.StatusBadRequest)
		return
	}
	if req.WorldType == "" {
		req.WorldType = "normal" // Valor por defecto
	}

	// Crear el mundo
	world, err := h.worldRepo.CreateWorld(req.Name, req.Description, req.WorldType, req.MaxPlayers)
	if err != nil {
		h.logger.Error("Error creando mundo", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(world)
}

// UpdateWorld actualiza un mundo existente
func (h *AdminHandler) UpdateWorld(w http.ResponseWriter, r *http.Request) {
	// Obtener el ID del mundo de la URL
	worldIDStr := chi.URLParam(r, "id")
	worldID, err := uuid.Parse(worldIDStr)
	if err != nil {
		http.Error(w, "ID de mundo inválido", http.StatusBadRequest)
		return
	}

	var req models.UpdateWorldRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Error decodificando solicitud de actualización de mundo", zap.Error(err))
		http.Error(w, "Error decodificando la solicitud", http.StatusBadRequest)
		return
	}

	// Validar datos de entrada
	if req.Name == "" {
		http.Error(w, "El nombre del mundo es requerido", http.StatusBadRequest)
		return
	}
	if req.MaxPlayers <= 0 {
		http.Error(w, "El número máximo de jugadores debe ser mayor a 0", http.StatusBadRequest)
		return
	}
	if req.WorldType == "" {
		req.WorldType = "normal" // Valor por defecto
	}

	// Actualizar el mundo
	world, err := h.worldRepo.UpdateWorld(worldID, req.Name, req.Description, req.WorldType, req.MaxPlayers)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Mundo no encontrado", http.StatusNotFound)
			return
		}
		h.logger.Error("Error actualizando mundo", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(world)
}

// DeleteWorld elimina un mundo
func (h *AdminHandler) DeleteWorld(w http.ResponseWriter, r *http.Request) {
	// Obtener el ID del mundo de la URL
	worldIDStr := chi.URLParam(r, "id")
	worldID, err := uuid.Parse(worldIDStr)
	if err != nil {
		http.Error(w, "ID de mundo inválido", http.StatusBadRequest)
		return
	}

	// Eliminar el mundo
	err = h.worldRepo.DeleteWorld(worldID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Mundo no encontrado", http.StatusNotFound)
			return
		}
		if err.Error() == "no se puede eliminar un mundo con jugadores activos" {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		h.logger.Error("Error eliminando mundo", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	response := models.WorldActionResponse{
		Success: true,
		Message: "Mundo eliminado exitosamente",
		WorldID: worldID.String(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// StartWorld activa un mundo
func (h *AdminHandler) StartWorld(w http.ResponseWriter, r *http.Request) {
	// Obtener el ID del mundo de la URL
	worldIDStr := chi.URLParam(r, "id")
	worldID, err := uuid.Parse(worldIDStr)
	if err != nil {
		http.Error(w, "ID de mundo inválido", http.StatusBadRequest)
		return
	}

	// Activar el mundo
	err = h.worldRepo.StartWorld(worldID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Mundo no encontrado", http.StatusNotFound)
			return
		}
		if err.Error() == "el mundo ya está activo" {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		h.logger.Error("Error activando mundo", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	response := models.WorldActionResponse{
		Success: true,
		Message: "Mundo activado exitosamente",
		WorldID: worldID.String(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// StopWorld desactiva un mundo
func (h *AdminHandler) StopWorld(w http.ResponseWriter, r *http.Request) {
	// Obtener el ID del mundo de la URL
	worldIDStr := chi.URLParam(r, "id")
	worldID, err := uuid.Parse(worldIDStr)
	if err != nil {
		http.Error(w, "ID de mundo inválido", http.StatusBadRequest)
		return
	}

	// Desactivar el mundo
	err = h.worldRepo.StopWorld(worldID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Mundo no encontrado", http.StatusNotFound)
			return
		}
		if err.Error() == "el mundo ya está inactivo" {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		h.logger.Error("Error desactivando mundo", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	response := models.WorldActionResponse{
		Success: true,
		Message: "Mundo desactivado exitosamente",
		WorldID: worldID.String(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetWorldStatus obtiene el estado completo de un mundo
func (h *AdminHandler) GetWorldStatus(w http.ResponseWriter, r *http.Request) {
	// Obtener el ID del mundo de la URL
	worldIDStr := chi.URLParam(r, "id")
	worldID, err := uuid.Parse(worldIDStr)
	if err != nil {
		http.Error(w, "ID de mundo inválido", http.StatusBadRequest)
		return
	}

	// Obtener el estado del mundo
	status, err := h.worldRepo.GetWorldStatus(worldID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Mundo no encontrado", http.StatusNotFound)
			return
		}
		h.logger.Error("Error obteniendo estado del mundo", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// AddPlayerToWorld agrega un jugador a un mundo específico
func (h *AdminHandler) AddPlayerToWorld(w http.ResponseWriter, r *http.Request) {
	// Obtener el ID del mundo de la URL
	worldIDStr := chi.URLParam(r, "id")
	worldID, err := uuid.Parse(worldIDStr)
	if err != nil {
		http.Error(w, "ID de mundo inválido", http.StatusBadRequest)
		return
	}

	// Decodificar el body
	var req struct {
		PlayerID string `json:"playerId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Error decodificando solicitud", zap.Error(err))
		http.Error(w, "Error decodificando la solicitud", http.StatusBadRequest)
		return
	}

	// Validar playerID
	if req.PlayerID == "" {
		http.Error(w, "PlayerID es requerido", http.StatusBadRequest)
		return
	}

	playerID, err := uuid.Parse(req.PlayerID)
	if err != nil {
		http.Error(w, "PlayerID inválido", http.StatusBadRequest)
		return
	}

	// Agregar jugador al mundo
	err = h.worldRepo.AddPlayerToWorld(playerID, worldID)
	if err != nil {
		if err.Error() == "mundo no encontrado" {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if err.Error() == "el mundo no está online" {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		if err.Error() == "el mundo está lleno" {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		h.logger.Error("Error agregando jugador al mundo", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	response := models.WorldActionResponse{
		Success: true,
		Message: "Jugador agregado al mundo exitosamente",
		WorldID: worldID.String(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// RemovePlayerFromWorld remueve un jugador de un mundo
func (h *AdminHandler) RemovePlayerFromWorld(w http.ResponseWriter, r *http.Request) {
	// Obtener el ID del mundo de la URL
	worldIDStr := chi.URLParam(r, "id")
	worldID, err := uuid.Parse(worldIDStr)
	if err != nil {
		http.Error(w, "ID de mundo inválido", http.StatusBadRequest)
		return
	}

	// Obtener el ID del jugador de la URL
	playerIDStr := chi.URLParam(r, "playerId")
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		http.Error(w, "ID de jugador inválido", http.StatusBadRequest)
		return
	}

	// Remover jugador del mundo
	err = h.worldRepo.RemovePlayerFromWorld(playerID, worldID)
	if err != nil {
		h.logger.Error("Error removiendo jugador del mundo", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	response := models.WorldActionResponse{
		Success: true,
		Message: "Jugador removido del mundo exitosamente",
		WorldID: worldID.String(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetWorldPlayers obtiene la lista de jugadores en un mundo
func (h *AdminHandler) GetWorldPlayers(w http.ResponseWriter, r *http.Request) {
	// Obtener el ID del mundo de la URL
	worldIDStr := chi.URLParam(r, "id")
	worldID, err := uuid.Parse(worldIDStr)
	if err != nil {
		http.Error(w, "ID de mundo inválido", http.StatusBadRequest)
		return
	}

	// Obtener jugadores del mundo
	players, err := h.worldRepo.GetWorldPlayers(worldID)
	if err != nil {
		h.logger.Error("Error obteniendo jugadores del mundo", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(players)
}

// GetCurrencyConfig obtiene la configuración de monedas
func (h *AdminHandler) GetCurrencyConfig(w http.ResponseWriter, r *http.Request) {
	config, err := h.currencyRepo.GetCurrencyConfig()
	if err != nil {
		h.logger.Error("Error obteniendo configuración de monedas", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
}

// UpdateCurrencyConfig actualiza la configuración de monedas
func (h *AdminHandler) UpdateCurrencyConfig(w http.ResponseWriter, r *http.Request) {
	var req models.UpdateCurrencyConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Error decodificando solicitud de actualización de monedas", zap.Error(err))
		http.Error(w, "Error decodificando la solicitud", http.StatusBadRequest)
		return
	}

	// Validar datos de entrada
	if req.GlobalCoin == "" {
		http.Error(w, "El nombre de la moneda global es requerido", http.StatusBadRequest)
		return
	}
	if req.WorldCoin == "" {
		http.Error(w, "El nombre de la moneda de mundo es requerido", http.StatusBadRequest)
		return
	}

	// Actualizar configuración
	config, err := h.currencyRepo.UpdateCurrencyConfig(req.GlobalCoin, req.WorldCoin)
	if err != nil {
		h.logger.Error("Error actualizando configuración de monedas", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
}

// GetPlayerCurrencyBalance obtiene el balance de monedas de un jugador
func (h *AdminHandler) GetPlayerCurrencyBalance(w http.ResponseWriter, r *http.Request) {
	// Obtener el ID del jugador de la URL
	playerIDStr := chi.URLParam(r, "playerId")
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		http.Error(w, "ID de jugador inválido", http.StatusBadRequest)
		return
	}

	// Obtener el ID del mundo de query params (opcional)
	worldIDStr := r.URL.Query().Get("worldId")
	var worldID *uuid.UUID
	if worldIDStr != "" {
		if id, err := uuid.Parse(worldIDStr); err == nil {
			worldID = &id
		}
	}

	// Obtener balance
	balance, err := h.currencyRepo.GetPlayerCurrencyBalance(playerID, worldID)
	if err != nil {
		h.logger.Error("Error obteniendo balance de monedas", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(balance)
}

// AddCurrency agrega monedas a un jugador
func (h *AdminHandler) AddCurrency(w http.ResponseWriter, r *http.Request) {
	var req models.AddCurrencyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Error decodificando solicitud de agregar monedas", zap.Error(err))
		http.Error(w, "Error decodificando la solicitud", http.StatusBadRequest)
		return
	}

	// Validar datos de entrada
	if req.PlayerID == "" {
		http.Error(w, "PlayerID es requerido", http.StatusBadRequest)
		return
	}
	if req.Amount <= 0 {
		http.Error(w, "La cantidad debe ser mayor a 0", http.StatusBadRequest)
		return
	}
	if req.CurrencyType == "" {
		http.Error(w, "El tipo de moneda es requerido", http.StatusBadRequest)
		return
	}

	playerID, err := uuid.Parse(req.PlayerID)
	if err != nil {
		http.Error(w, "PlayerID inválido", http.StatusBadRequest)
		return
	}

	// Agregar monedas según el tipo
	if req.CurrencyType == "global" {
		err = h.currencyRepo.AddGlobalCurrency(playerID, req.Amount, req.Description)
	} else if req.CurrencyType == "world" {
		if req.WorldID == "" {
			http.Error(w, "WorldID es requerido para moneda de mundo", http.StatusBadRequest)
			return
		}
		worldID, err := uuid.Parse(req.WorldID)
		if err != nil {
			http.Error(w, "WorldID inválido", http.StatusBadRequest)
			return
		}
		err = h.currencyRepo.AddWorldCurrency(playerID, worldID, req.Amount, req.Description)
	} else {
		http.Error(w, "Tipo de moneda inválido", http.StatusBadRequest)
		return
	}

	if err != nil {
		h.logger.Error("Error agregando monedas", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	response := models.WorldActionResponse{
		Success: true,
		Message: "Monedas agregadas exitosamente",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// TransferCurrency transfiere monedas entre jugadores
func (h *AdminHandler) TransferCurrency(w http.ResponseWriter, r *http.Request) {
	var req models.TransferCurrencyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Error decodificando solicitud de transferencia", zap.Error(err))
		http.Error(w, "Error decodificando la solicitud", http.StatusBadRequest)
		return
	}

	// Validar datos de entrada
	if req.FromPlayerID == "" || req.ToPlayerID == "" {
		http.Error(w, "FromPlayerID y ToPlayerID son requeridos", http.StatusBadRequest)
		return
	}
	if req.Amount <= 0 {
		http.Error(w, "La cantidad debe ser mayor a 0", http.StatusBadRequest)
		return
	}
	if req.CurrencyType == "" {
		http.Error(w, "El tipo de moneda es requerido", http.StatusBadRequest)
		return
	}

	fromPlayerID, err := uuid.Parse(req.FromPlayerID)
	if err != nil {
		http.Error(w, "FromPlayerID inválido", http.StatusBadRequest)
		return
	}

	toPlayerID, err := uuid.Parse(req.ToPlayerID)
	if err != nil {
		http.Error(w, "ToPlayerID inválido", http.StatusBadRequest)
		return
	}

	// Transferir monedas según el tipo
	if req.CurrencyType == "global" {
		err = h.currencyRepo.TransferGlobalCurrency(fromPlayerID, toPlayerID, req.Amount, req.Description)
	} else if req.CurrencyType == "world" {
		if req.WorldID == "" {
			http.Error(w, "WorldID es requerido para moneda de mundo", http.StatusBadRequest)
			return
		}
		worldID, err := uuid.Parse(req.WorldID)
		if err != nil {
			http.Error(w, "WorldID inválido", http.StatusBadRequest)
			return
		}
		err = h.currencyRepo.TransferWorldCurrency(fromPlayerID, toPlayerID, worldID, req.Amount, req.Description)
	} else {
		http.Error(w, "Tipo de moneda inválido", http.StatusBadRequest)
		return
	}

	if err != nil {
		if err.Error() == "fondos insuficientes" {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		h.logger.Error("Error transfiriendo monedas", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	response := models.WorldActionResponse{
		Success: true,
		Message: "Transferencia realizada exitosamente",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetCurrencyTransactions obtiene el historial de transacciones de un jugador
func (h *AdminHandler) GetCurrencyTransactions(w http.ResponseWriter, r *http.Request) {
	// Obtener el ID del jugador de la URL
	playerIDStr := chi.URLParam(r, "playerId")
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		http.Error(w, "ID de jugador inválido", http.StatusBadRequest)
		return
	}

	// Obtener límite de query params (opcional, por defecto 50)
	limit := 50
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 1000 {
			limit = l
		}
	}

	// Obtener transacciones
	transactions, err := h.currencyRepo.GetCurrencyTransactions(playerID, limit)
	if err != nil {
		h.logger.Error("Error obteniendo transacciones", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(transactions)
}

// GetCurrencyStats obtiene estadísticas de monedas
func (h *AdminHandler) GetCurrencyStats(w http.ResponseWriter, r *http.Request) {
	// Obtener el ID del mundo de query params (opcional)
	worldIDStr := r.URL.Query().Get("worldId")

	var stats map[string]interface{}
	var err error

	if worldIDStr != "" {
		// Estadísticas de mundo específico
		worldID, err := uuid.Parse(worldIDStr)
		if err != nil {
			http.Error(w, "WorldID inválido", http.StatusBadRequest)
			return
		}
		stats, err = h.currencyRepo.GetWorldCurrencyStats(worldID)
	} else {
		// Estadísticas globales
		stats, err = h.currencyRepo.GetGlobalCurrencyStats()
	}

	if err != nil {
		h.logger.Error("Error obteniendo estadísticas de monedas", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
