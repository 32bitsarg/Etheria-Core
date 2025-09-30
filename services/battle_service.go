package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"server-backend/models"
	"server-backend/repository"
	"server-backend/websocket"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type BattleService struct {
	battleRepo   *repository.BattleRepository
	villageRepo  *repository.VillageRepository
	unitRepo     *repository.UnitRepository
	logger       *zap.Logger
	wsManager    *websocket.Manager
	redisService *RedisService
}

type BattleData struct {
	ID           int64                  `json:"id"`
	AttackerID   int64                  `json:"attacker_id"`
	DefenderID   int64                  `json:"defender_id"`
	AttackerName string                 `json:"attacker_name"`
	DefenderName string                 `json:"defender_name"`
	Status       string                 `json:"status"` // "pending", "in_progress", "completed"
	StartTime    time.Time              `json:"start_time"`
	EndTime      time.Time              `json:"end_time"`
	Result       string                 `json:"result"` // "attacker_win", "defender_win", "draw"
	Units        map[string]interface{} `json:"units"`
	Rewards      map[string]interface{} `json:"rewards"`
}

type MatchmakingRequest struct {
	PlayerID    int64     `json:"player_id"`
	Username    string    `json:"username"`
	Level       int       `json:"level"`
	RequestedAt time.Time `json:"requested_at"`
}

func NewBattleService(battleRepo *repository.BattleRepository, villageRepo *repository.VillageRepository, unitRepo *repository.UnitRepository, logger *zap.Logger, redisService *RedisService) *BattleService {
	return &BattleService{
		battleRepo:   battleRepo,
		villageRepo:  villageRepo,
		unitRepo:     unitRepo,
		logger:       logger,
		wsManager:    nil, // Se establecerá después con SetWebSocketManager
		redisService: redisService,
	}
}

// CreateBattle crea una nueva batalla con Redis
func (s *BattleService) CreateBattle(request *models.BattleRequest) (*models.Battle, error) {
	// Rate limiting: verificar que el jugador no esté atacando demasiado rápido
	rateLimitKey := fmt.Sprintf("battle_rate_limit:%s", request.AttackerID.String())
	attackCount, err := s.redisService.GetCounter(rateLimitKey)
	if err == nil && attackCount >= 5 { // Máximo 5 ataques por hora
		return nil, fmt.Errorf("límite de ataques excedido. Intenta de nuevo en 1 hora")
	}

	// Validar la solicitud
	if err := s.validateBattleRequest(request); err != nil {
		return nil, fmt.Errorf("solicitud de batalla inválida: %w", err)
	}

	// Verificar que el atacante tiene las unidades necesarias
	if err := s.validateAttackerUnits(request.AttackerID, request.Units); err != nil {
		return nil, fmt.Errorf("unidades insuficientes: %w", err)
	}

	// Guardar la batalla usando el repositorio
	config := map[string]interface{}{
		"units":           request.Units,
		"mode":            request.Mode,
		"formation":       request.Formation,
		"tactics":         request.Tactics,
		"terrain":         request.Terrain,
		"weather":         request.Weather,
		"max_waves":       request.MaxWaves,
		"max_duration":    request.MaxDuration,
		"advanced_config": request.AdvancedConfig,
	}

	createdBattle, err := s.battleRepo.CreateBattle(request.AttackerID, request.DefenderVillageID, request.BattleType, request.Mode, config)
	if err != nil {
		return nil, fmt.Errorf("error creando batalla: %w", err)
	}

	// Incrementar contador de rate limiting
	s.redisService.IncrementCounter(rateLimitKey)
	s.redisService.SetCounter(rateLimitKey, attackCount+1, time.Hour) // Expira en 1 hora

	// Cache de batalla activa
	battleKey := fmt.Sprintf("battle:%s", createdBattle.ID.String())
	s.redisService.SetCache(battleKey, createdBattle, time.Hour) // Cache por 1 hora

	// Agregar a lista de batallas activas usando cola
	ctx := context.Background()
	s.redisService.AddToQueue(ctx, "active_battles", createdBattle.ID.String())

	// Notificar a los jugadores sobre la nueva batalla
	s.notifyBattleCreated(createdBattle)

	// Actualizar cache de rankings
	s.updateBattleRankingsCache()

	return createdBattle, nil
}

// validateBattleRequest valida una solicitud de batalla
func (s *BattleService) validateBattleRequest(request *models.BattleRequest) error {
	if request.AttackerVillageID == uuid.Nil {
		return fmt.Errorf("aldea atacante requerida")
	}
	if request.DefenderVillageID == uuid.Nil {
		return fmt.Errorf("aldea defensora requerida")
	}
	if request.BattleType == "" {
		return fmt.Errorf("tipo de batalla requerido")
	}
	if request.Mode == "" {
		return fmt.Errorf("modo de batalla requerido")
	}
	if len(request.Units) == 0 {
		return fmt.Errorf("se requieren unidades para la batalla")
	}

	// Validar tipos de batalla permitidos
	allowedTypes := map[string]bool{
		"pvp":   true,
		"pve":   true,
		"siege": true,
		"raid":  true,
	}
	if !allowedTypes[request.BattleType] {
		return fmt.Errorf("tipo de batalla no válido")
	}

	// Validar modos permitidos
	allowedModes := map[string]bool{
		"basic":    true,
		"advanced": true,
	}
	if !allowedModes[request.Mode] {
		return fmt.Errorf("modo de batalla no válido")
	}

	return nil
}

// validateAttackerUnits verifica que el atacante tiene las unidades necesarias
func (s *BattleService) validateAttackerUnits(playerID uuid.UUID, units map[string]int) error {
	playerUnits, err := s.battleRepo.GetPlayerUnits(playerID)
	if err != nil {
		return fmt.Errorf("error obteniendo unidades del jugador: %w", err)
	}

	// Crear mapa de unidades disponibles
	availableUnits := make(map[string]int)
	for _, unit := range playerUnits {
		availableUnits[unit.UnitID.String()] = unit.Quantity
	}

	// Verificar que tiene suficientes unidades
	for unitType, requiredQuantity := range units {
		unitID, err := uuid.Parse(unitType)
		if err != nil {
			return fmt.Errorf("ID de unidad inválido: %s", unitType)
		}

		available, exists := availableUnits[unitID.String()]
		if !exists || available < requiredQuantity {
			return fmt.Errorf("unidades insuficientes para %s", unitType)
		}
	}

	return nil
}

// GetPlayerBattleReports obtiene las batallas de un jugador (reportes)
func (s *BattleService) GetPlayerBattleReports(playerID uuid.UUID, limit int) ([]models.Battle, error) {
	return s.battleRepo.GetBattlesByPlayer(playerID, limit)
}

// GetBattleReport obtiene el detalle de una batalla
func (s *BattleService) GetBattleReport(battleID uuid.UUID) (*models.Battle, error) {
	return s.battleRepo.GetBattle(battleID)
}

// GetBattleWithDetails obtiene una batalla con todos sus detalles (con cache)
func (s *BattleService) GetBattleWithDetails(battleID uuid.UUID) (*models.BattleWithDetails, error) {
	// Intentar obtener del cache primero
	cacheKey := fmt.Sprintf("battle_details:%s", battleID.String())
	var cachedDetails models.BattleWithDetails
	if err := s.redisService.GetCache(cacheKey, &cachedDetails); err == nil {
		return &cachedDetails, nil
	}

	battle, err := s.battleRepo.GetBattle(battleID)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo batalla: %w", err)
	}

	waves, err := s.battleRepo.GetBattleWaves(battleID)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo oleadas: %w", err)
	}

	// Obtener unidades del atacante y defensor
	attackerUnits, err := s.battleRepo.GetPlayerUnits(battle.AttackerID)
	if err != nil {
		s.logger.Warn("Error obteniendo unidades del atacante", zap.Error(err))
		attackerUnits = []models.PlayerUnit{}
	}

	defenderUnits, err := s.battleRepo.GetPlayerUnits(battle.DefenderID)
	if err != nil {
		s.logger.Warn("Error obteniendo unidades del defensor", zap.Error(err))
		defenderUnits = []models.PlayerUnit{}
	}

	details := &models.BattleWithDetails{
		Battle:        battle,
		Waves:         waves,
		AttackerUnits: attackerUnits,
		DefenderUnits: defenderUnits,
	}

	// Cache por 30 minutos
	s.redisService.SetCache(cacheKey, details, 30*time.Minute)

	return details, nil
}

// ProcessBattle procesa una batalla con Redis
func (s *BattleService) ProcessBattle(battleID uuid.UUID) error {
	// Obtener batalla del cache o base de datos
	battleKey := fmt.Sprintf("battle:%s", battleID.String())
	var battle models.Battle
	if err := s.redisService.GetCache(battleKey, &battle); err != nil {
		// Si no está en cache, obtener de la base de datos
		battlePtr, err := s.battleRepo.GetBattle(battleID)
		if err != nil {
			return fmt.Errorf("error obteniendo batalla: %w", err)
		}
		battle = *battlePtr
	}

	if battle.Status != "pending" {
		return fmt.Errorf("la batalla no está pendiente")
	}

	// Simular la batalla
	result, err := s.simulateBattle(&battle)
	if err != nil {
		return fmt.Errorf("error simulando batalla: %w", err)
	}

	// Actualizar estado de la batalla
	battle.Status = "completed"
	now := time.Now()
	battle.EndTime = &now
	battle.Winner = result.Winner

	// Guardar en base de datos
	if err := s.battleRepo.UpdateBattle(&battle); err != nil {
		return fmt.Errorf("error actualizando batalla: %w", err)
	}

	// Actualizar cache
	s.redisService.SetCache(battleKey, battle, time.Hour)

	// Remover de batallas activas usando cola
	ctx := context.Background()
	s.redisService.RemoveFromQueue(ctx, "active_battles", int64(battleID.ID()))

	// Actualizar estadísticas de jugadores
	s.updatePlayerBattleStatistics(&battle, result)

	// Notificar a los jugadores
	s.notifyBattleCompleted(&battle, result)

	// Actualizar cache de rankings
	s.updateBattleRankingsCache()

	return nil
}

// simulateBattle simula una batalla y retorna el resultado
func (s *BattleService) simulateBattle(battle *models.Battle) (*BattleResult, error) {
	// Obtener unidades de ambos bandos
	attackerUnits, err := s.battleRepo.GetPlayerUnits(battle.AttackerID)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo unidades del atacante: %w", err)
	}

	defenderUnits, err := s.battleRepo.GetPlayerUnits(battle.DefenderID)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo unidades del defensor: %w", err)
	}

	// Calcular poder total de cada bando
	attackerPower := s.calculateTotalPower(attackerUnits)
	defenderPower := s.calculateTotalPower(defenderUnits)

	// Simular resultado basado en poder y aleatoriedad
	result := &BattleResult{
		AttackerLosses: "{}",
		DefenderLosses: "{}",
	}

	// Fórmula simple: poder + aleatoriedad
	attackerRoll := rand.Float64() * 0.2 // ±10% aleatoriedad
	defenderRoll := rand.Float64() * 0.2

	attackerFinalPower := attackerPower * (1 + attackerRoll)
	defenderFinalPower := defenderPower * (1 + defenderRoll)

	if attackerFinalPower > defenderFinalPower {
		result.Winner = "attacker"
		// Calcular pérdidas proporcionales
		ratio := defenderFinalPower / attackerFinalPower
		result.AttackerLosses = s.calculateLosses(attackerUnits, 1-ratio)
		result.DefenderLosses = s.calculateLosses(defenderUnits, 0.8) // 80% de pérdidas para el perdedor
	} else if defenderFinalPower > attackerFinalPower {
		result.Winner = "defender"
		ratio := attackerFinalPower / defenderFinalPower
		result.AttackerLosses = s.calculateLosses(attackerUnits, 0.8)
		result.DefenderLosses = s.calculateLosses(defenderUnits, 1-ratio)
	} else {
		result.Winner = "draw"
		result.AttackerLosses = s.calculateLosses(attackerUnits, 0.5)
		result.DefenderLosses = s.calculateLosses(defenderUnits, 0.5)
	}

	return result, nil
}

// calculateTotalPower calcula el poder total de un conjunto de unidades
func (s *BattleService) calculateTotalPower(units []models.PlayerUnit) float64 {
	totalPower := 0.0
	for _, unit := range units {
		// Poder = (ataque + defensa) * cantidad * nivel
		unitPower := float64(unit.CurrentAttack+unit.CurrentDefense) * float64(unit.Quantity) * float64(unit.Level)
		totalPower += unitPower
	}
	return totalPower
}

// calculateLosses calcula las pérdidas de unidades
func (s *BattleService) calculateLosses(units []models.PlayerUnit, lossRatio float64) string {
	losses := make(map[string]int)
	for _, unit := range units {
		losses[unit.UnitID.String()] = int(float64(unit.Quantity) * lossRatio)
	}

	lossesJSON, _ := json.Marshal(losses)
	return string(lossesJSON)
}

// updatePlayerBattleStatistics actualiza las estadísticas de batalla de los jugadores
func (s *BattleService) updatePlayerBattleStatistics(battle *models.Battle, result *BattleResult) {
	// Actualizar estadísticas del atacante
	attackerStats, err := s.battleRepo.GetBattleStatistics(battle.AttackerID)
	if err != nil {
		s.logger.Error("Error obteniendo estadísticas del atacante", zap.Error(err))
		return
	}

	attackerStats.TotalBattles++
	if result.Winner == "attacker" {
		attackerStats.BattlesWon++
	} else if result.Winner == "defender" {
		attackerStats.BattlesLost++
	}

	if attackerStats.TotalBattles > 0 {
		attackerStats.WinRate = float64(attackerStats.BattlesWon) / float64(attackerStats.TotalBattles) * 100
	}

	attackerStats.LastBattleDate = time.Now()
	if battle.Duration > attackerStats.LongestBattleTime {
		attackerStats.LongestBattleTime = battle.Duration
	}
	if attackerStats.ShortestBattleTime == 0 || battle.Duration < attackerStats.ShortestBattleTime {
		attackerStats.ShortestBattleTime = battle.Duration
	}

	if err := s.battleRepo.UpdateBattleStatistics(attackerStats); err != nil {
		s.logger.Error("Error actualizando estadísticas del atacante", zap.Error(err))
	}

	// Actualizar estadísticas del defensor
	defenderStats, err := s.battleRepo.GetBattleStatistics(battle.DefenderID)
	if err != nil {
		s.logger.Error("Error obteniendo estadísticas del defensor", zap.Error(err))
		return
	}

	defenderStats.TotalBattles++
	if result.Winner == "defender" {
		defenderStats.BattlesWon++
	} else if result.Winner == "attacker" {
		defenderStats.BattlesLost++
	}

	if defenderStats.TotalBattles > 0 {
		defenderStats.WinRate = float64(defenderStats.BattlesWon) / float64(defenderStats.TotalBattles) * 100
	}

	defenderStats.LastBattleDate = time.Now()
	if battle.Duration > defenderStats.LongestBattleTime {
		defenderStats.LongestBattleTime = battle.Duration
	}
	if defenderStats.ShortestBattleTime == 0 || battle.Duration < defenderStats.ShortestBattleTime {
		defenderStats.ShortestBattleTime = battle.Duration
	}

	if err := s.battleRepo.UpdateBattleStatistics(defenderStats); err != nil {
		s.logger.Error("Error actualizando estadísticas del defensor", zap.Error(err))
	}
}

// GetBattleStatistics obtiene las estadísticas de batalla de un jugador (con cache)
func (s *BattleService) GetBattleStatistics(playerID uuid.UUID) (*models.BattleStatistics, error) {
	// Intentar obtener del cache primero
	cacheKey := fmt.Sprintf("battle_stats:%s", playerID.String())
	var cachedStats models.BattleStatistics
	if err := s.redisService.GetCache(cacheKey, &cachedStats); err == nil {
		return &cachedStats, nil
	}

	stats, err := s.battleRepo.GetBattleStatistics(playerID)
	if err != nil {
		return nil, err
	}

	// Cache por 15 minutos
	s.redisService.SetCache(cacheKey, stats, 15*time.Minute)

	return stats, nil
}

// GetBattleRankings obtiene los rankings de batalla (con cache)
func (s *BattleService) GetBattleRankings(limit int) ([]models.BattleRanking, error) {
	// Intentar obtener del cache primero
	cacheKey := fmt.Sprintf("battle_rankings:%d", limit)
	var cachedRankings []models.BattleRanking
	if err := s.redisService.GetCache(cacheKey, &cachedRankings); err == nil {
		return cachedRankings, nil
	}

	rankings, err := s.battleRepo.GetBattleRankings(limit)
	if err != nil {
		return nil, err
	}

	// Cache por 10 minutos
	s.redisService.SetCache(cacheKey, rankings, 10*time.Minute)

	return rankings, nil
}

// GetActiveBattles obtiene las batallas activas (con cache)
func (s *BattleService) GetActiveBattles(limit int) ([]models.Battle, error) {
	// Intentar obtener del cache primero
	cacheKey := fmt.Sprintf("active_battles:%d", limit)
	var cachedBattles []models.Battle
	if err := s.redisService.GetCache(cacheKey, &cachedBattles); err == nil {
		return cachedBattles, nil
	}

	// Obtener batallas activas de la base de datos
	battles, err := s.battleRepo.GetBattlesByStatus("active", limit)
	if err != nil {
		return nil, err
	}

	// Cache por 5 minutos
	s.redisService.SetCache(cacheKey, battles, 5*time.Minute)

	return battles, nil
}

// GetPendingBattles obtiene las batallas pendientes
func (s *BattleService) GetPendingBattles(limit int) ([]models.Battle, error) {
	return s.battleRepo.GetBattlesByStatus("pending", limit)
}

// CancelBattle cancela una batalla con Redis
func (s *BattleService) CancelBattle(battleID uuid.UUID, playerID uuid.UUID) error {
	// Obtener batalla
	battle, err := s.battleRepo.GetBattle(battleID)
	if err != nil {
		return fmt.Errorf("batalla no encontrada: %w", err)
	}

	// Verificar autorización
	if battle.AttackerID != playerID && battle.DefenderID != playerID {
		return fmt.Errorf("no autorizado para cancelar esta batalla")
	}

	if battle.Status != "pending" {
		return fmt.Errorf("solo se pueden cancelar batallas pendientes")
	}

	// Actualizar estado
	battle.Status = "cancelled"
	now := time.Now()
	battle.EndTime = &now

	// Guardar en base de datos
	if err := s.battleRepo.UpdateBattle(battle); err != nil {
		return fmt.Errorf("error actualizando batalla: %w", err)
	}

	// Actualizar cache
	battleKey := fmt.Sprintf("battle:%s", battleID.String())
	s.redisService.SetCache(battleKey, battle, time.Hour)

	// Remover de batallas activas usando cola
	ctx := context.Background()
	s.redisService.RemoveFromQueue(ctx, "active_battles", int64(battleID.ID()))

	// Notificar a los jugadores
	s.notifyBattleCancelled(battle)

	return nil
}

// notifyBattleCreated notifica a los jugadores sobre una nueva batalla
func (s *BattleService) notifyBattleCreated(battle *models.Battle) {
	if s.wsManager == nil {
		s.logger.Warn("WebSocket Manager no disponible para notificaciones de batalla")
		return
	}

	// Notificar al atacante
	attackerMessage := map[string]interface{}{
		"type": "battle_created",
		"data": map[string]interface{}{
			"battle_id":   battle.ID.String(),
			"defender_id": battle.DefenderID.String(),
			"battle_type": battle.BattleType,
			"status":      battle.Status,
			"created_at":  battle.CreatedAt.Unix(),
			"message":     "Tu ataque ha sido iniciado",
		},
	}

	if err := s.wsManager.SendToUser(battle.AttackerID.String(), "battle_notification", attackerMessage); err != nil {
		s.logger.Warn("Error enviando notificación de batalla al atacante", zap.Error(err))
	}

	// Notificar al defensor
	defenderMessage := map[string]interface{}{
		"type": "battle_incoming",
		"data": map[string]interface{}{
			"battle_id":   battle.ID.String(),
			"attacker_id": battle.AttackerID.String(),
			"battle_type": battle.BattleType,
			"status":      battle.Status,
			"created_at":  battle.CreatedAt.Unix(),
			"message":     "¡Estás siendo atacado!",
		},
	}

	if err := s.wsManager.SendToUser(battle.DefenderID.String(), "battle_notification", defenderMessage); err != nil {
		s.logger.Warn("Error enviando notificación de batalla al defensor", zap.Error(err))
	}

	s.logger.Info("Notificaciones de batalla enviadas",
		zap.String("battle_id", battle.ID.String()),
		zap.String("attacker_id", battle.AttackerID.String()),
		zap.String("defender_id", battle.DefenderID.String()),
	)
}

// notifyBattleCompleted notifica a los jugadores sobre el resultado de una batalla
func (s *BattleService) notifyBattleCompleted(battle *models.Battle, result *BattleResult) {
	if s.wsManager == nil {
		s.logger.Warn("WebSocket Manager no disponible para notificaciones de batalla")
		return
	}

	// Notificar al atacante
	attackerMessage := map[string]interface{}{
		"type": "battle_completed",
		"data": map[string]interface{}{
			"battle_id":       battle.ID.String(),
			"result":          result.Winner,
			"attacker_losses": result.AttackerLosses,
			"defender_losses": result.DefenderLosses,
			"duration":        battle.Duration,
			"completed_at":    battle.UpdatedAt.Unix(),
			"message":         fmt.Sprintf("Batalla completada. Resultado: %s", result.Winner),
		},
	}

	if err := s.wsManager.SendToUser(battle.AttackerID.String(), "battle_notification", attackerMessage); err != nil {
		s.logger.Warn("Error enviando notificación de batalla completada al atacante", zap.Error(err))
	}

	// Notificar al defensor
	defenderMessage := map[string]interface{}{
		"type": "battle_completed",
		"data": map[string]interface{}{
			"battle_id":       battle.ID.String(),
			"result":          result.Winner,
			"attacker_losses": result.AttackerLosses,
			"defender_losses": result.DefenderLosses,
			"duration":        battle.Duration,
			"completed_at":    battle.UpdatedAt.Unix(),
			"message":         fmt.Sprintf("Batalla completada. Resultado: %s", result.Winner),
		},
	}

	if err := s.wsManager.SendToUser(battle.DefenderID.String(), "battle_notification", defenderMessage); err != nil {
		s.logger.Warn("Error enviando notificación de batalla completada al defensor", zap.Error(err))
	}

	s.logger.Info("Notificaciones de batalla completada enviadas",
		zap.String("battle_id", battle.ID.String()),
		zap.String("winner", result.Winner),
		zap.Int("duration", battle.Duration),
	)
}

// notifyBattleCancelled notifica a los jugadores sobre la cancelación de una batalla
func (s *BattleService) notifyBattleCancelled(battle *models.Battle) {
	if s.wsManager == nil {
		s.logger.Warn("WebSocket Manager no disponible para notificaciones de batalla")
		return
	}

	// Notificar al atacante
	attackerMessage := map[string]interface{}{
		"type": "battle_cancelled",
		"data": map[string]interface{}{
			"battle_id":    battle.ID.String(),
			"defender_id":  battle.DefenderID.String(),
			"cancelled_at": battle.UpdatedAt.Unix(),
			"message":      "Tu ataque ha sido cancelado",
		},
	}

	if err := s.wsManager.SendToUser(battle.AttackerID.String(), "battle_notification", attackerMessage); err != nil {
		s.logger.Warn("Error enviando notificación de batalla cancelada al atacante", zap.Error(err))
	}

	// Notificar al defensor
	defenderMessage := map[string]interface{}{
		"type": "battle_cancelled",
		"data": map[string]interface{}{
			"battle_id":    battle.ID.String(),
			"attacker_id":  battle.AttackerID.String(),
			"cancelled_at": battle.UpdatedAt.Unix(),
			"message":      "El ataque ha sido cancelado",
		},
	}

	if err := s.wsManager.SendToUser(battle.DefenderID.String(), "battle_notification", defenderMessage); err != nil {
		s.logger.Warn("Error enviando notificación de batalla cancelada al defensor", zap.Error(err))
	}

	s.logger.Info("Notificaciones de batalla cancelada enviadas",
		zap.String("battle_id", battle.ID.String()),
		zap.String("attacker_id", battle.AttackerID.String()),
	)
}

// SetWebSocketManager establece el manager de WebSocket
func (s *BattleService) SetWebSocketManager(wsManager *websocket.Manager) {
	s.wsManager = wsManager
}

// BattleResult representa el resultado de una batalla
type BattleResult struct {
	Winner         string `json:"winner"`
	AttackerLosses string `json:"attacker_losses"`
	DefenderLosses string `json:"defender_losses"`
}

// RequestBattle solicita una batalla PvP (matchmaking)
func (s *BattleService) RequestBattle(ctx context.Context, playerID int64, username string, level int) error {
	request := &MatchmakingRequest{
		PlayerID:    playerID,
		Username:    username,
		Level:       level,
		RequestedAt: time.Now(),
	}

	// Agregar a la cola de matchmaking
	queueKey := fmt.Sprintf("matchmaking_queue:%d", level)
	requestJSON, _ := json.Marshal(request)

	if err := s.redisService.AddToQueue(ctx, queueKey, string(requestJSON)); err != nil {
		return fmt.Errorf("error agregando a cola de matchmaking: %w", err)
	}

	// Notificar al jugador que está en cola
	s.notifyMatchmakingStatus(playerID, "en_cola")

	return nil
}

// ProcessMatchmaking procesa la cola de matchmaking
func (s *BattleService) ProcessMatchmaking(ctx context.Context) error {
	// Procesar colas por nivel
	for level := 1; level <= 100; level++ {
		if err := s.processMatchmakingQueue(ctx, fmt.Sprintf("matchmaking_queue:%d", level)); err != nil {
			s.logger.Error("Error procesando cola de matchmaking", zap.Int("level", level), zap.Error(err))
		}
	}
	return nil
}

// processMatchmakingQueue procesa una cola específica de matchmaking
func (s *BattleService) processMatchmakingQueue(ctx context.Context, queueKey string) error {
	// Obtener todos los jugadores en la cola
	players, err := s.redisService.GetQueue(ctx, queueKey)
	if err != nil {
		return err
	}

	if len(players) < 2 {
		return nil // No hay suficientes jugadores
	}

	// Parsear jugadores
	var requests []*MatchmakingRequest
	for _, playerJSON := range players {
		var request MatchmakingRequest
		if err := json.Unmarshal([]byte(playerJSON), &request); err != nil {
			s.logger.Error("Error parseando request de matchmaking", zap.Error(err))
			continue
		}
		requests = append(requests, &request)
	}

	// Emparejar jugadores
	for i := 0; i < len(requests)-1; i += 2 {
		attacker := requests[i]
		defender := requests[i+1]

		// Crear batalla
		if err := s.createBattle(ctx, attacker, defender); err != nil {
			s.logger.Error("Error creando batalla desde matchmaking", zap.Error(err))
			continue
		}

		// Remover de la cola
		s.redisService.RemoveFromQueue(ctx, queueKey, attacker.PlayerID)
		s.redisService.RemoveFromQueue(ctx, queueKey, defender.PlayerID)
	}

	return nil
}

// createBattle crea una batalla desde matchmaking
func (s *BattleService) createBattle(ctx context.Context, attacker, defender *MatchmakingRequest) error {
	// Obtener IDs reales de los jugadores
	attackerID, err := s.getPlayerUUIDFromID(attacker.PlayerID)
	if err != nil {
		return fmt.Errorf("error obteniendo ID del atacante: %w", err)
	}

	defenderID, err := s.getPlayerUUIDFromID(defender.PlayerID)
	if err != nil {
		return fmt.Errorf("error obteniendo ID del defensor: %w", err)
	}

	// Obtener aldea del defensor
	defenderVillageID, err := s.getDefenderVillageID(defenderID)
	if err != nil {
		return fmt.Errorf("error obteniendo aldea del defensor: %w", err)
	}

	// Crear batalla PvP con IDs reales
	battleRequest := &models.BattleRequest{
		AttackerID:        attackerID,
		DefenderVillageID: defenderVillageID,
		BattleType:        "pvp",
		Mode:              "basic",
		Units:             map[string]int{"default": 100}, // Unidades por defecto
	}

	// Crear la batalla
	battle, err := s.CreateBattle(battleRequest)
	if err != nil {
		return err
	}

	// Notificar a ambos jugadores
	s.notifyBattleCreated(battle)

	return nil
}

// updateBattleRankingsCache actualiza el cache de rankings
func (s *BattleService) updateBattleRankingsCache() {
	// Limpiar cache de rankings
	keys := []string{"battle_rankings:10", "battle_rankings:20", "battle_rankings:50", "battle_rankings:100"}
	for _, key := range keys {
		s.redisService.DeleteCache(key)
	}
}

// notifyMatchmakingStatus notifica el estado del matchmaking
func (s *BattleService) notifyMatchmakingStatus(playerID int64, status string) {
	notification := map[string]interface{}{
		"type":      "matchmaking_status",
		"player_id": playerID,
		"status":    status,
		"timestamp": time.Now(),
	}

	// Enviar por WebSocket si está disponible
	if s.wsManager != nil {
		s.wsManager.SendToUser(fmt.Sprintf("%d", playerID), "matchmaking_status", notification)
	}

	// También guardar en Redis para persistencia
	notificationKey := fmt.Sprintf("matchmaking_notification:%d", playerID)
	s.redisService.SetCache(notificationKey, notification, 5*time.Minute) // 5 minutos
}

// GetBattle obtiene datos de una batalla desde cache o BD
func (s *BattleService) GetBattle(ctx context.Context, battleID int64) (*BattleData, error) {
	// Intentar obtener desde cache
	battleKey := fmt.Sprintf("battle:%d", battleID)
	var battleData BattleData

	err := s.redisService.GetCache(battleKey, &battleData)
	if err == nil {
		return &battleData, nil
	}

	// Convertir int64 a UUID correctamente
	battleUUID, err := s.convertInt64ToUUID(battleID)
	if err != nil {
		return nil, fmt.Errorf("error convirtiendo ID de batalla: %w", err)
	}

	// Obtener desde BD
	battle, err := s.battleRepo.GetBattle(battleUUID)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo batalla: %v", err)
	}

	// Convertir a BattleData
	battleData = BattleData{
		ID:           int64(battle.ID.ID()), // Convertir UUID a int64
		AttackerID:   int64(battle.AttackerID.ID()),
		DefenderID:   int64(battle.DefenderID.ID()),
		AttackerName: "", // Campo no disponible en el modelo
		DefenderName: "", // Campo no disponible en el modelo
		Status:       battle.Status,
		StartTime:    *battle.StartTime,
		EndTime:      *battle.EndTime,
		Result:       battle.Winner,
		Units:        make(map[string]interface{}),
		Rewards:      make(map[string]interface{}),
	}

	// Cachear resultado por 10 minutos
	err = s.redisService.SetCache(battleKey, battleData, 10*time.Minute)
	if err != nil {
		log.Printf("Error cacheando batalla: %v", err)
	}

	return &battleData, nil
}

// StartBattle inicia una batalla
func (s *BattleService) StartBattle(ctx context.Context, battleID int64) error {
	// Convertir int64 a UUID correctamente
	battleUUID, err := s.convertInt64ToUUID(battleID)
	if err != nil {
		return fmt.Errorf("error convirtiendo ID de batalla: %w", err)
	}

	// Obtener batalla
	battle, err := s.battleRepo.GetBattle(battleUUID)
	if err != nil {
		return fmt.Errorf("error obteniendo batalla: %v", err)
	}

	// Actualizar estado
	battle.Status = "in_progress"
	now := time.Now()
	battle.StartTime = &now
	err = s.battleRepo.UpdateBattle(battle)
	if err != nil {
		return fmt.Errorf("error actualizando batalla: %v", err)
	}

	// Actualizar cache
	battleKey := fmt.Sprintf("battle:%d", battleID)
	err = s.redisService.DeleteCache(battleKey)
	if err != nil {
		log.Printf("Error invalidando cache de batalla: %v", err)
	}

	// Programar fin de batalla (simulación de 5 minutos)
	timerKey := fmt.Sprintf("battle:end:%d", battleID)
	err = s.redisService.SetCache(timerKey, battleID, 5*time.Minute)
	if err != nil {
		log.Printf("Error programando fin de batalla: %v", err)
	}

	return nil
}

// ProcessBattleResults procesa los resultados de batallas (llamado por un worker)
func (s *BattleService) ProcessBattleResults(ctx context.Context) error {
	// Obtener batallas que deben finalizar
	endKeys, err := s.redisService.GetKeys(ctx, "battle:end:*")
	if err != nil {
		return fmt.Errorf("error obteniendo timers de batalla: %v", err)
	}

	for _, key := range endKeys {
		// Extraer battleID del key
		var battleID int64
		err := s.redisService.GetCache(key, &battleID)
		if err != nil {
			continue
		}

		// Finalizar batalla
		err = s.finishBattle(ctx, battleID)
		if err != nil {
			log.Printf("Error finalizando batalla %d: %v", battleID, err)
			continue
		}

		// Remover timer
		s.redisService.DeleteCache(key)
	}

	return nil
}

// finishBattle finaliza una batalla
func (s *BattleService) finishBattle(ctx context.Context, battleID int64) error {
	// Convertir int64 a UUID correctamente
	battleUUID, err := s.convertInt64ToUUID(battleID)
	if err != nil {
		return fmt.Errorf("error convirtiendo ID de batalla: %w", err)
	}

	// Obtener batalla
	battle, err := s.battleRepo.GetBattle(battleUUID)
	if err != nil {
		return fmt.Errorf("error obteniendo batalla: %v", err)
	}

	// Simular resultado
	result := &BattleResult{
		Winner:         "attacker", // Simplificado
		AttackerLosses: "10%",
		DefenderLosses: "15%",
	}

	// Actualizar batalla
	battle.Status = "completed"
	now := time.Now()
	battle.EndTime = &now
	battle.Winner = result.Winner

	err = s.battleRepo.UpdateBattle(battle)
	if err != nil {
		return fmt.Errorf("error actualizando batalla: %v", err)
	}

	// Actualizar estadísticas
	s.updatePlayerBattleStatistics(battle, result)

	// Notificar a jugadores
	s.notifyBattleCompleted(battle, result)

	// Actualizar cache
	battleKey := fmt.Sprintf("battle:%d", battleID)
	s.redisService.DeleteCache(battleKey)

	return nil
}

// GetPlayerBattles obtiene las batallas de un jugador
func (s *BattleService) GetPlayerBattles(ctx context.Context, playerID int64, limit int) ([]*BattleData, error) {
	// Convertir int64 a UUID correctamente
	playerUUID, err := s.convertInt64ToUUID(playerID)
	if err != nil {
		return nil, fmt.Errorf("error convirtiendo ID de jugador: %w", err)
	}

	battles, err := s.battleRepo.GetBattlesByPlayer(playerUUID, limit)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo batallas del jugador: %v", err)
	}

	var battleData []*BattleData
	for _, battle := range battles {
		data := &BattleData{
			ID:         int64(battle.ID.ID()),
			AttackerID: int64(battle.AttackerID.ID()),
			DefenderID: int64(battle.DefenderID.ID()),
			Status:     battle.Status,
			StartTime:  *battle.StartTime,
			EndTime:    *battle.EndTime,
			Result:     battle.Winner,
			Units:      make(map[string]interface{}),
			Rewards:    make(map[string]interface{}),
		}
		battleData = append(battleData, data)
	}

	return battleData, nil
}

// ============================================================================
// FUNCIONES AUXILIARES PARA MANEJO CORRECTO DE IDs
// ============================================================================

// convertInt64ToUUID convierte un int64 a UUID de manera segura
func (s *BattleService) convertInt64ToUUID(id int64) (uuid.UUID, error) {
	// Crear un UUID a partir del int64 usando el método correcto
	// Esto mantiene la consistencia con el sistema de IDs
	uuidBytes := make([]byte, 16)
	for i := 0; i < 8; i++ {
		uuidBytes[i] = byte(id >> (8 * i))
	}

	// Generar el resto del UUID de manera determinística
	for i := 8; i < 16; i++ {
		uuidBytes[i] = byte(id >> (8 * (i - 8)))
	}

	return uuid.FromBytes(uuidBytes)
}

// getPlayerUUIDFromID obtiene el UUID de un jugador desde su ID int64
func (s *BattleService) getPlayerUUIDFromID(playerID int64) (uuid.UUID, error) {
	// En un sistema real, esto consultaría la base de datos
	// Por ahora, usamos la conversión directa
	return s.convertInt64ToUUID(playerID)
}

// getDefenderVillageID obtiene el ID de la aldea del defensor
func (s *BattleService) getDefenderVillageID(defenderID uuid.UUID) (uuid.UUID, error) {
	// En un sistema real, esto consultaría la base de datos para obtener
	// la aldea principal del jugador defensor
	// Por ahora, generamos un ID determinístico basado en el jugador
	defenderIDInt := int64(defenderID.ID())
	villageIDInt := defenderIDInt + 1000000 // Offset para distinguir aldeas
	return s.convertInt64ToUUID(villageIDInt)
}
