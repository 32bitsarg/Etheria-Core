package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"server-backend/models"
	"server-backend/repository"
	"server-backend/services"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type WorldClientHandler struct {
	worldRepo    *repository.WorldRepository
	playerRepo   *repository.PlayerRepository
	villageRepo  *repository.VillageRepository
	allianceRepo *repository.AllianceRepository
	battleRepo   *repository.BattleRepository
	economyRepo  *repository.EconomyRepository
	worldService *services.WorldService
	logger       *zap.Logger
}

func NewWorldClientHandler(
	worldRepo *repository.WorldRepository,
	playerRepo *repository.PlayerRepository,
	villageRepo *repository.VillageRepository,
	allianceRepo *repository.AllianceRepository,
	battleRepo *repository.BattleRepository,
	economyRepo *repository.EconomyRepository,
	worldService *services.WorldService,
	logger *zap.Logger,
) *WorldClientHandler {
	return &WorldClientHandler{
		worldRepo:    worldRepo,
		playerRepo:   playerRepo,
		villageRepo:  villageRepo,
		allianceRepo: allianceRepo,
		battleRepo:   battleRepo,
		economyRepo:  economyRepo,
		worldService: worldService,
		logger:       logger,
	}
}

// GetWorlds obtiene la lista de mundos disponibles para el cliente
func (h *WorldClientHandler) GetWorlds(w http.ResponseWriter, r *http.Request) {
	worlds, err := h.worldRepo.GetAllWorlds()
	if err != nil {
		h.logger.Error("Error obteniendo mundos", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	var clientWorlds []models.WorldClientResponse
	for _, world := range worlds {
		// Solo mostrar mundos activos
		if !world.IsActive {
			continue
		}

		// Calcular si está lleno y si se puede unir
		isFull := world.CurrentPlayers >= world.MaxPlayers
		canJoin := world.IsOnline && !isFull

		// Calcular uptime si está online
		var uptime string
		if world.IsOnline && world.LastStartedAt != nil {
			uptime = time.Since(*world.LastStartedAt).String()
		}

		// Determinar características según el tipo de mundo
		features := models.WorldFeatures{
			PvPEnabled:       world.WorldType != "peaceful",
			AlliancesEnabled: true,
			TradingEnabled:   true,
			EventsEnabled:    true,
		}

		clientWorld := models.WorldClientResponse{
			ID:             world.ID,
			Name:           world.Name,
			Description:    world.Description,
			MaxPlayers:     world.MaxPlayers,
			CurrentPlayers: world.CurrentPlayers,
			IsOnline:       world.IsOnline,
			WorldType:      world.WorldType,
			Status:         world.Status,
			PlayerCount:    world.CurrentPlayers,
			IsFull:         isFull,
			CanJoin:        canJoin,
			LastStartedAt:  world.LastStartedAt,
			Uptime:         uptime,
			Features:       features,
		}

		clientWorlds = append(clientWorlds, clientWorld)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(clientWorlds)
}

// GetWorld obtiene los detalles de un mundo específico
func (h *WorldClientHandler) GetWorld(w http.ResponseWriter, r *http.Request) {
	worldIDStr := chi.URLParam(r, "id")
	worldID, err := uuid.Parse(worldIDStr)
	if err != nil {
		http.Error(w, "ID de mundo inválido", http.StatusBadRequest)
		return
	}

	world, err := h.worldRepo.GetWorldByID(worldID)
	if err != nil {
		h.logger.Error("Error obteniendo mundo", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	if world == nil {
		http.Error(w, "Mundo no encontrado", http.StatusNotFound)
		return
	}

	// Calcular si está lleno y si se puede unir
	isFull := world.CurrentPlayers >= world.MaxPlayers
	canJoin := world.IsOnline && !isFull

	// Calcular uptime si está online
	var uptime string
	if world.IsOnline && world.LastStartedAt != nil {
		uptime = time.Since(*world.LastStartedAt).String()
	}

	// Determinar características según el tipo de mundo
	features := models.WorldFeatures{
		PvPEnabled:       world.WorldType != "peaceful",
		AlliancesEnabled: true,
		TradingEnabled:   true,
		EventsEnabled:    true,
	}

	clientWorld := models.WorldClientResponse{
		ID:             world.ID,
		Name:           world.Name,
		Description:    world.Description,
		MaxPlayers:     world.MaxPlayers,
		CurrentPlayers: world.CurrentPlayers,
		IsOnline:       world.IsOnline,
		WorldType:      world.WorldType,
		Status:         world.Status,
		PlayerCount:    world.CurrentPlayers,
		IsFull:         isFull,
		CanJoin:        canJoin,
		LastStartedAt:  world.LastStartedAt,
		Uptime:         uptime,
		Features:       features,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(clientWorld)
}

// JoinWorld permite a un jugador unirse a un mundo
func (h *WorldClientHandler) JoinWorld(w http.ResponseWriter, r *http.Request) {
	// Obtener el ID del jugador del contexto
	playerIDStr := r.Context().Value("player_id").(string)
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		h.logger.Error("Error parseando ID de jugador", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	worldIDStr := chi.URLParam(r, "id")
	worldID, err := uuid.Parse(worldIDStr)
	if err != nil {
		http.Error(w, "ID de mundo inválido", http.StatusBadRequest)
		return
	}

	var req models.WorldJoinRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Error decodificando solicitud", zap.Error(err))
		http.Error(w, "Error decodificando la solicitud", http.StatusBadRequest)
		return
	}

	// Validar datos de entrada
	if req.VillageName == "" {
		req.VillageName = "Mi Aldea"
	}
	if req.StartingLocation == "" {
		req.StartingLocation = "random"
	}

	// Verificar que el mundo existe y está disponible
	world, err := h.worldRepo.GetWorldByID(worldID)
	if err != nil {
		h.logger.Error("Error obteniendo mundo", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	if world == nil {
		http.Error(w, "Mundo no encontrado", http.StatusNotFound)
		return
	}

	if !world.IsOnline {
		http.Error(w, "El mundo no está disponible", http.StatusServiceUnavailable)
		return
	}

	if world.CurrentPlayers >= world.MaxPlayers {
		http.Error(w, "El mundo está lleno", http.StatusConflict)
		return
	}

	// Verificar que el jugador no esté ya en este mundo
	currentWorld, err := h.worldRepo.GetPlayerCurrentWorld(playerID)
	if err != nil {
		h.logger.Error("Error verificando mundo actual del jugador", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	if currentWorld != nil && currentWorld.ID == worldID {
		http.Error(w, "Ya estás en este mundo", http.StatusBadRequest)
		return
	}

	// Unir al jugador al mundo
	err = h.worldRepo.AddPlayerToWorld(playerID, worldID)
	if err != nil {
		h.logger.Error("Error uniendo jugador al mundo", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	// Crear aldea inicial para el jugador con coordenadas aleatorias
	x, y, err := h.villageRepo.GenerateRandomCoordinates(worldID)
	if err != nil {
		h.logger.Error("Error generando coordenadas aleatorias", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}
	
	village, err := h.villageRepo.CreateVillage(playerID, worldID, req.VillageName, x, y)
	if err != nil {
		h.logger.Error("Error creando aldea inicial", zap.Error(err))
		// No fallar la operación, solo loggear el error
	}

	// Recursos iniciales
	startingResources := models.ResourceSet{
		Gold:  1000,
		Wood:  500,
		Stone: 300,
		Food:  200,
	}

	villageID := ""
	if village != nil {
		villageID = village.Village.ID.String()
	}

	response := models.WorldJoinResponse{
		Success:           true,
		Message:           "Te has unido al mundo exitosamente",
		WorldID:           worldID.String(),
		VillageID:         villageID,
		StartingResources: startingResources,
		RedirectUrl:       "/game/world/" + worldID.String(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// LeaveWorld permite a un jugador salir de un mundo
func (h *WorldClientHandler) LeaveWorld(w http.ResponseWriter, r *http.Request) {
	// Obtener el ID del jugador del contexto
	playerIDStr := r.Context().Value("player_id").(string)
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		h.logger.Error("Error parseando ID de jugador", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	worldIDStr := chi.URLParam(r, "id")
	worldID, err := uuid.Parse(worldIDStr)
	if err != nil {
		http.Error(w, "ID de mundo inválido", http.StatusBadRequest)
		return
	}

	// Eliminar todas las aldeas del jugador en este mundo
	err = h.villageRepo.DeleteVillagesByPlayerAndWorld(playerID, worldID)
	if err != nil {
		h.logger.Error("Error eliminando aldeas del jugador", zap.Error(err))
		// No fallar, continuar con el proceso
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
		Message: "Has salido del mundo",
		WorldID: worldID.String(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetCurrentWorld obtiene el mundo actual del jugador
func (h *WorldClientHandler) GetCurrentWorld(w http.ResponseWriter, r *http.Request) {
	// Obtener el ID del jugador del contexto
	playerIDStr := r.Context().Value("player_id").(string)
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		h.logger.Error("Error parseando ID de jugador", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	// Obtener el mundo actual del jugador
	world, err := h.worldRepo.GetPlayerCurrentWorld(playerID)
	if err != nil {
		h.logger.Error("Error obteniendo mundo actual del jugador", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	if world == nil {
		http.Error(w, "No estás en ningún mundo", http.StatusNotFound)
		return
	}

	// Obtener información adicional del jugador
	player, err := h.playerRepo.GetPlayerByID(playerID)
	if err != nil {
		h.logger.Error("Error obteniendo información del jugador", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	// Obtener número de aldeas del jugador
	villages, err := h.villageRepo.GetVillagesByPlayerID(playerID)
	if err != nil {
		h.logger.Error("Error obteniendo aldeas del jugador", zap.Error(err))
		// No fallar, usar 0 como valor por defecto
	}

	villageCount := 0
	if villages != nil {
		villageCount = len(villages)
	}

	response := models.PlayerWorldInfo{
		WorldID:      world.ID.String(),
		WorldName:    world.Name,
		JoinedAt:     player.LastLogin, // Usar last login como aproximación
		LastSeen:     time.Now(),
		VillageCount: villageCount,
		IsActive:     true,
		CanLeave:     true,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetWorldStats obtiene las estadísticas de un mundo
func (h *WorldClientHandler) GetWorldStats(w http.ResponseWriter, r *http.Request) {
	worldIDStr := chi.URLParam(r, "id")
	worldID, err := uuid.Parse(worldIDStr)
	if err != nil {
		http.Error(w, "ID de mundo inválido", http.StatusBadRequest)
		return
	}

	world, err := h.worldRepo.GetWorldByID(worldID)
	if err != nil {
		h.logger.Error("Error obteniendo mundo", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	if world == nil {
		http.Error(w, "Mundo no encontrado", http.StatusNotFound)
		return
	}

	// Obtener estadísticas básicas
	playerCount := world.CurrentPlayers
	maxPlayers := world.MaxPlayers

	// Obtener estadísticas adicionales (simplificadas por ahora)
	allianceCount := 0 // TODO: Implementar
	villageCount := 0  // TODO: Implementar
	battleCount := 0   // TODO: Implementar
	tradeCount := 0    // TODO: Implementar

	// Top players (simplificado)
	topPlayers := []models.TopPlayer{
		{
			Username:     "Player1",
			Level:        25,
			VillageCount: 5,
		},
	}

	// Actividad reciente (simplificado)
	recentActivity := []models.ActivityItem{
		{
			Type:        "battle",
			Description: "Player1 atacó a Player2",
			Timestamp:   time.Now().Add(-time.Hour),
		},
	}

	stats := models.WorldStats{
		WorldID:        worldID.String(),
		Name:           world.Name,
		PlayerCount:    playerCount,
		MaxPlayers:     maxPlayers,
		AllianceCount:  allianceCount,
		VillageCount:   villageCount,
		BattleCount:    battleCount,
		TradeCount:     tradeCount,
		TopPlayers:     topPlayers,
		RecentActivity: recentActivity,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// GetWorldPlayers obtiene la lista de jugadores en un mundo
func (h *WorldClientHandler) GetWorldPlayers(w http.ResponseWriter, r *http.Request) {
	worldIDStr := chi.URLParam(r, "id")
	worldID, err := uuid.Parse(worldIDStr)
	if err != nil {
		http.Error(w, "ID de mundo inválido", http.StatusBadRequest)
		return
	}

	// Obtener parámetros de query
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	_ = r.URL.Query().Get("sort") // TODO: Implementar ordenamiento

	limit := 50 // Default
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	offset := 0
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Obtener jugadores del mundo
	players, err := h.worldRepo.GetWorldPlayers(worldID)
	if err != nil {
		h.logger.Error("Error obteniendo jugadores del mundo", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	// Convertir a formato de respuesta
	var clientPlayers []models.WorldPlayerListItem
	for _, player := range players {
		clientPlayer := models.WorldPlayerListItem{
			PlayerID:     player.PlayerID.String(),
			Username:     player.Username,
			Level:        player.Level,
			VillageCount: player.VillageCount,
			AllianceName: nil, // TODO: Obtener nombre de alianza
			LastSeen:     player.LastSeen,
			IsOnline:     player.IsActive,
		}
		clientPlayers = append(clientPlayers, clientPlayer)
	}

	// Aplicar paginación
	start := offset
	end := start + limit
	if start >= len(clientPlayers) {
		start = len(clientPlayers)
	}
	if end > len(clientPlayers) {
		end = len(clientPlayers)
	}

	result := clientPlayers[start:end]

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// GetWorldStatus obtiene el estado de un mundo
func (h *WorldClientHandler) GetWorldStatus(w http.ResponseWriter, r *http.Request) {
	worldIDStr := chi.URLParam(r, "id")
	worldID, err := uuid.Parse(worldIDStr)
	if err != nil {
		http.Error(w, "ID de mundo inválido", http.StatusBadRequest)
		return
	}

	world, err := h.worldRepo.GetWorldByID(worldID)
	if err != nil {
		h.logger.Error("Error obteniendo mundo", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	if world == nil {
		http.Error(w, "Mundo no encontrado", http.StatusNotFound)
		return
	}

	isFull := world.CurrentPlayers >= world.MaxPlayers
	canJoin := world.IsOnline && !isFull

	status := models.WorldStatusResponse{
		WorldID:           worldID.String(),
		IsOnline:          world.IsOnline,
		IsFull:            isFull,
		CanJoin:           canJoin,
		MaintenanceMode:   world.Status == "maintenance",
		EstimatedWaitTime: 0, // TODO: Implementar cálculo de tiempo de espera
		ServerLoad:        float64(world.CurrentPlayers) / float64(world.MaxPlayers),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}
