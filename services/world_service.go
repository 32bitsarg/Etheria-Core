package services

import (
	"errors"
	"server-backend/models"
	"server-backend/repository"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

var (
	ErrNoWorldsAvailable    = errors.New("no hay mundos disponibles")
	ErrWorldFull            = errors.New("el mundo está lleno")
	ErrPlayerAlreadyInWorld = errors.New("el jugador ya está en un mundo")
)

type WorldService struct {
	worldRepo    *repository.WorldRepository
	playerRepo   *repository.PlayerRepository
	villageRepo  *repository.VillageRepository
	allianceRepo *repository.AllianceRepository
	battleRepo   *repository.BattleRepository
	economyRepo  *repository.EconomyRepository
	logger       *zap.Logger
}

func NewWorldService(
	worldRepo *repository.WorldRepository,
	playerRepo *repository.PlayerRepository,
	villageRepo *repository.VillageRepository,
	allianceRepo *repository.AllianceRepository,
	battleRepo *repository.BattleRepository,
	economyRepo *repository.EconomyRepository,
	logger *zap.Logger,
) *WorldService {
	return &WorldService{
		worldRepo:    worldRepo,
		playerRepo:   playerRepo,
		villageRepo:  villageRepo,
		allianceRepo: allianceRepo,
		battleRepo:   battleRepo,
		economyRepo:  economyRepo,
		logger:       logger,
	}
}

// GetAvailableWorlds obtiene mundos disponibles para el cliente
func (s *WorldService) GetAvailableWorlds() ([]models.WorldClientResponse, error) {
	worlds, err := s.worldRepo.GetAllWorlds()
	if err != nil {
		return nil, err
	}

	var clientWorlds []models.WorldClientResponse
	for _, world := range worlds {
		// Solo incluir mundos activos
		if !world.IsActive {
			continue
		}

		clientWorld := s.convertToClientResponse(world)
		clientWorlds = append(clientWorlds, clientWorld)
	}

	return clientWorlds, nil
}

// GetWorldDetails obtiene detalles completos de un mundo
func (s *WorldService) GetWorldDetails(worldID uuid.UUID) (*models.WorldClientResponse, error) {
	world, err := s.worldRepo.GetWorldByID(worldID)
	if err != nil {
		return nil, err
	}

	if world == nil {
		return nil, errors.New("mundo no encontrado")
	}

	clientWorld := s.convertToClientResponse(world)
	return &clientWorld, nil
}

// JoinWorld maneja la lógica de unirse a un mundo
func (s *WorldService) JoinWorld(playerID, worldID uuid.UUID, villageName, startingLocation string) (*models.WorldJoinResponse, error) {
	// Verificar que el mundo existe y está disponible
	world, err := s.worldRepo.GetWorldByID(worldID)
	if err != nil {
		return nil, err
	}

	if world == nil {
		return nil, errors.New("mundo no encontrado")
	}

	if !world.IsOnline {
		return nil, errors.New("el mundo no está disponible")
	}

	if world.CurrentPlayers >= world.MaxPlayers {
		return nil, errors.New("el mundo está lleno")
	}

	// Verificar que el jugador no esté ya en este mundo
	currentWorld, err := s.worldRepo.GetPlayerCurrentWorld(playerID)
	if err != nil {
		return nil, err
	}

	if currentWorld != nil && currentWorld.ID == worldID {
		return nil, errors.New("ya estás en este mundo")
	}

	// Unir al jugador al mundo
	err = s.worldRepo.AddPlayerToWorld(playerID, worldID)
	if err != nil {
		return nil, err
	}

	// Crear aldea inicial con coordenadas aleatorias
	x, y, err := s.villageRepo.GenerateRandomCoordinates(worldID)
	if err != nil {
		s.logger.Error("Error generando coordenadas aleatorias", zap.Error(err))
		return nil, err
	}
	
	village, err := s.villageRepo.CreateVillage(playerID, worldID, villageName, x, y)
	villageID := ""
	if err != nil {
		s.logger.Error("Error creando aldea inicial", zap.Error(err))
		// No fallar la operación completa
	} else if village != nil {
		villageID = village.Village.ID.String()
	}

	// Recursos iniciales
	startingResources := models.ResourceSet{
		Gold:  1000,
		Wood:  500,
		Stone: 300,
		Food:  200,
	}

	response := &models.WorldJoinResponse{
		Success:           true,
		Message:           "Te has unido al mundo exitosamente",
		WorldID:           worldID.String(),
		VillageID:         villageID,
		StartingResources: startingResources,
		RedirectUrl:       "/game/world/" + worldID.String(),
	}

	return response, nil
}

// LeaveWorld maneja la lógica de salir de un mundo
func (s *WorldService) LeaveWorld(playerID, worldID uuid.UUID) error {
	// Verificar que el jugador esté en este mundo
	currentWorld, err := s.worldRepo.GetPlayerCurrentWorld(playerID)
	if err != nil {
		return err
	}

	if currentWorld == nil || currentWorld.ID != worldID {
		return errors.New("no estás en este mundo")
	}

	// Remover jugador del mundo
	err = s.worldRepo.RemovePlayerFromWorld(playerID, worldID)
	if err != nil {
		return err
	}

	return nil
}

// GetPlayerCurrentWorld obtiene el mundo actual del jugador
func (s *WorldService) GetPlayerCurrentWorld(playerID uuid.UUID) (*models.PlayerWorldInfo, error) {
	world, err := s.worldRepo.GetPlayerCurrentWorld(playerID)
	if err != nil {
		return nil, err
	}

	if world == nil {
		return nil, errors.New("no estás en ningún mundo")
	}

	// Obtener información adicional del jugador
	player, err := s.playerRepo.GetPlayerByID(playerID)
	if err != nil {
		return nil, err
	}

	// Obtener número de aldeas
	villages, err := s.villageRepo.GetVillagesByPlayerID(playerID)
	if err != nil {
		s.logger.Error("Error obteniendo aldeas del jugador", zap.Error(err))
		// No fallar, usar 0 como valor por defecto
	}

	villageCount := 0
	if villages != nil {
		villageCount = len(villages)
	}

	response := &models.PlayerWorldInfo{
		WorldID:      world.ID.String(),
		WorldName:    world.Name,
		JoinedAt:     player.LastLogin,
		LastSeen:     time.Now(),
		VillageCount: villageCount,
		IsActive:     true,
		CanLeave:     true,
	}

	return response, nil
}

// GetWorldStats obtiene estadísticas detalladas de un mundo
func (s *WorldService) GetWorldStats(worldID uuid.UUID) (*models.WorldStats, error) {
	world, err := s.worldRepo.GetWorldByID(worldID)
	if err != nil {
		return nil, err
	}

	if world == nil {
		return nil, errors.New("mundo no encontrado")
	}

	// Obtener estadísticas básicas
	playerCount := world.CurrentPlayers
	maxPlayers := world.MaxPlayers

	// TODO: Implementar estadísticas reales
	allianceCount := 0
	villageCount := 0
	battleCount := 0
	tradeCount := 0

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

	stats := &models.WorldStats{
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

	return stats, nil
}

// GetWorldStatus obtiene el estado actual de un mundo
func (s *WorldService) GetWorldStatus(worldID uuid.UUID) (*models.WorldStatusResponse, error) {
	world, err := s.worldRepo.GetWorldByID(worldID)
	if err != nil {
		return nil, err
	}

	if world == nil {
		return nil, errors.New("mundo no encontrado")
	}

	isFull := world.CurrentPlayers >= world.MaxPlayers
	canJoin := world.IsOnline && !isFull

	status := &models.WorldStatusResponse{
		WorldID:           worldID.String(),
		IsOnline:          world.IsOnline,
		IsFull:            isFull,
		CanJoin:           canJoin,
		MaintenanceMode:   world.Status == "maintenance",
		EstimatedWaitTime: 0, // TODO: Implementar cálculo
		ServerLoad:        float64(world.CurrentPlayers) / float64(world.MaxPlayers),
	}

	return status, nil
}

// convertToClientResponse convierte un World a WorldClientResponse
func (s *WorldService) convertToClientResponse(world *models.World) models.WorldClientResponse {
	isFull := world.CurrentPlayers >= world.MaxPlayers
	canJoin := world.IsOnline && !isFull

	var uptime string
	if world.IsOnline && world.LastStartedAt != nil {
		uptime = time.Since(*world.LastStartedAt).String()
	}

	features := models.WorldFeatures{
		PvPEnabled:       world.WorldType != "peaceful",
		AlliancesEnabled: true,
		TradingEnabled:   true,
		EventsEnabled:    true,
	}

	return models.WorldClientResponse{
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
}
