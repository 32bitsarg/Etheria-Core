package services

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"server-backend/models"
	"server-backend/repository"
	"server-backend/websocket"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type AchievementService struct {
	achievementRepo *repository.AchievementRepository
	playerRepo      *repository.PlayerRepository
	wsManager       *websocket.Manager
	logger          *zap.Logger
	redisService    *RedisService
}

func NewAchievementService(
	achievementRepo *repository.AchievementRepository,
	playerRepo *repository.PlayerRepository,
	wsManager *websocket.Manager,
	logger *zap.Logger,
	redisService *RedisService,
) *AchievementService {
	return &AchievementService{
		achievementRepo: achievementRepo,
		playerRepo:      playerRepo,
		wsManager:       wsManager,
		logger:          logger,
		redisService:    redisService,
	}
}

// GetAchievementDashboard obtiene el dashboard principal de achievements
func (s *AchievementService) GetAchievementDashboard(playerID string) (*models.AchievementDashboard, error) {
	// Convertir playerID string a UUID
	playerUUID, err := uuid.Parse(playerID)
	if err != nil {
		return nil, fmt.Errorf("playerID inválido: %w", err)
	}

	// Obtener estadísticas del jugador
	playerStats, err := s.achievementRepo.GetAchievementStatistics(playerUUID)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo estadísticas: %w", err)
	}

	// Obtener categorías
	categories, err := s.achievementRepo.GetAchievementCategories(true) // activeOnly = true
	if err != nil {
		return nil, fmt.Errorf("error obteniendo categorías: %w", err)
	}

	// Obtener achievements recientes (completados)
	recentAchievements, err := s.achievementRepo.GetAchievements(nil, true, false) // activeOnly = true, includeHidden = false
	if err != nil {
		return nil, fmt.Errorf("error obteniendo achievements recientes: %w", err)
	}

	// Obtener achievements próximos (no completados)
	upcomingAchievements, err := s.achievementRepo.GetAchievements(nil, true, false) // activeOnly = true, includeHidden = false
	if err != nil {
		return nil, fmt.Errorf("error obteniendo achievements próximos: %w", err)
	}

	// Obtener leaderboard
	leaderboard, err := s.achievementRepo.GetAchievementLeaderboard(10, nil) // limit = 10, worldID = nil
	if err != nil {
		return nil, fmt.Errorf("error obteniendo leaderboard: %w", err)
	}

	// Obtener notificaciones
	notifications, err := s.achievementRepo.GetAchievementNotifications(playerUUID, false) // unreadOnly = false
	if err != nil {
		return nil, fmt.Errorf("error obteniendo notificaciones: %w", err)
	}

	return &models.AchievementDashboard{
		PlayerStats:          playerStats,
		Categories:           categories,
		RecentAchievements:   recentAchievements,
		UpcomingAchievements: upcomingAchievements,
		Leaderboard:          leaderboard,
		Notifications:        notifications,
		GlobalStats:          map[string]interface{}{},
		LastUpdated:          time.Now(),
	}, nil
}

// GetAchievementCategories obtiene todas las categorías de achievements
func (s *AchievementService) GetAchievementCategories() ([]*models.AchievementCategory, error) {
	categories, err := s.achievementRepo.GetAchievementCategories(true) // activeOnly = true
	if err != nil {
		return nil, fmt.Errorf("error obteniendo categorías: %w", err)
	}

	// Convertir a punteros
	var categoriesPtr []*models.AchievementCategory
	for i := range categories {
		categoriesPtr = append(categoriesPtr, &categories[i])
	}

	return categoriesPtr, nil
}

// GetAchievementCategory obtiene una categoría específica
func (s *AchievementService) GetAchievementCategory(categoryID string) (*models.AchievementCategory, error) {
	// Convertir categoryID string a UUID
	categoryUUID, err := uuid.Parse(categoryID)
	if err != nil {
		return nil, fmt.Errorf("categoryID inválido: %w", err)
	}

	category, err := s.achievementRepo.GetAchievementCategory(categoryUUID)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo categoría: %w", err)
	}

	return category, nil
}

// CreateAchievementCategory crea una nueva categoría de achievements
func (s *AchievementService) CreateAchievementCategory(category *models.AchievementCategory) error {
	// Validar categoría
	if err := s.validateAchievementCategory(category); err != nil {
		return fmt.Errorf("categoría inválida: %w", err)
	}

	// Crear categoría
	if err := s.achievementRepo.CreateAchievementCategory(category); err != nil {
		return fmt.Errorf("error creando categoría: %w", err)
	}

	s.logger.Info("Nueva categoría de achievement creada",
		zap.String("name", category.Name),
		zap.String("category_id", category.ID.String()),
	)

	return nil
}

// GetAchievements obtiene todos los achievements
func (s *AchievementService) GetAchievements(filters map[string]interface{}) ([]*models.Achievement, error) {
	// Extraer filtros
	var categoryID *uuid.UUID
	activeOnly := true
	includeHidden := false

	if catIDStr, ok := filters["category_id"].(string); ok {
		if catID, err := uuid.Parse(catIDStr); err == nil {
			categoryID = &catID
		}
	}
	if active, ok := filters["active_only"].(bool); ok {
		activeOnly = active
	}
	if hidden, ok := filters["include_hidden"].(bool); ok {
		includeHidden = hidden
	}

	achievements, err := s.achievementRepo.GetAchievements(categoryID, activeOnly, includeHidden)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo achievements: %w", err)
	}

	// Convertir a punteros
	var achievementsPtr []*models.Achievement
	for i := range achievements {
		achievementsPtr = append(achievementsPtr, &achievements[i])
	}

	return achievementsPtr, nil
}

// GetAchievement obtiene un achievement específico
func (s *AchievementService) GetAchievement(achievementID string) (*models.Achievement, error) {
	// Convertir achievementID string a UUID
	achievementUUID, err := uuid.Parse(achievementID)
	if err != nil {
		return nil, fmt.Errorf("achievementID inválido: %w", err)
	}

	achievement, err := s.achievementRepo.GetAchievement(achievementUUID)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo achievement: %w", err)
	}

	return achievement, nil
}

// GetAchievementWithDetails obtiene un achievement con todos sus detalles
func (s *AchievementService) GetAchievementWithDetails(achievementID string) (*models.AchievementWithDetails, error) {
	// Convertir achievementID string a UUID
	achievementUUID, err := uuid.Parse(achievementID)
	if err != nil {
		return nil, fmt.Errorf("achievementID inválido: %w", err)
	}

	// Obtener achievement básico
	achievement, err := s.achievementRepo.GetAchievement(achievementUUID)
	if err != nil {
		return nil, err
	}

	// Obtener categoría
	category, _ := s.achievementRepo.GetAchievementCategory(achievement.CategoryID)

	// Obtener progreso del jugador (por ahora nil, se implementará cuando se tenga el playerID)
	var playerProgress *models.PlayerAchievement

	// Obtener recompensas del achievement
	rewards := achievement.Rewards

	// Obtener prerequisitos (achievements que deben completarse antes)
	var prerequisites []models.Achievement
	if achievement.Prerequisites != "" {
		// Los prerequisitos están en formato JSON string, por ahora los dejamos vacíos
		// TODO: Implementar parsing de JSON cuando se defina el formato
	}

	// Obtener hitos del achievement (están en el campo Milestones del AchievementProgress)
	var milestones []models.AchievementMilestone
	// Los milestones se obtienen del AchievementProgress, no del Achievement

	// Obtener estadísticas globales del achievement
	statistics, err := s.achievementRepo.GetAchievementStatistics(achievementUUID)
	if err != nil {
		// Si no hay estadísticas, crear unas por defecto
		statistics = &models.AchievementStatistics{
			TotalAchievements:     0,
			CompletedAchievements: 0,
			CompletionRate:        0.0,
		}
	}

	return &models.AchievementWithDetails{
		Achievement:    achievement,
		Category:       category,
		PlayerProgress: playerProgress,
		Rewards:        rewards,
		Prerequisites:  prerequisites,
		Milestones:     milestones,
		Statistics:     statistics,
	}, nil
}

// GetPlayerAchievements obtiene los achievements de un jugador
func (s *AchievementService) GetPlayerAchievements(playerID string, filters map[string]interface{}) ([]*models.PlayerAchievement, error) {
	// Convertir playerID string a UUID
	playerUUID, err := uuid.Parse(playerID)
	if err != nil {
		return nil, fmt.Errorf("playerID inválido: %w", err)
	}

	var categoryID *uuid.UUID
	completedOnly := false

	if catIDStr, ok := filters["category_id"].(string); ok {
		if catID, err := uuid.Parse(catIDStr); err == nil {
			categoryID = &catID
		}
	}
	if completed, ok := filters["completed_only"].(bool); ok {
		completedOnly = completed
	}

	achievements, err := s.achievementRepo.GetPlayerAchievements(playerUUID, categoryID, completedOnly)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo achievements del jugador: %w", err)
	}

	// Convertir a punteros
	var achievementsPtr []*models.PlayerAchievement
	for i := range achievements {
		achievementsPtr = append(achievementsPtr, &achievements[i])
	}

	return achievementsPtr, nil
}

// GetPlayerAchievement obtiene un achievement específico de un jugador
func (s *AchievementService) GetPlayerAchievement(playerID, achievementID string) (*models.PlayerAchievement, error) {
	// Convertir IDs string a UUID
	playerUUID, err := uuid.Parse(playerID)
	if err != nil {
		return nil, fmt.Errorf("playerID inválido: %w", err)
	}

	achievementUUID, err := uuid.Parse(achievementID)
	if err != nil {
		return nil, fmt.Errorf("achievementID inválido: %w", err)
	}

	achievement, err := s.achievementRepo.GetPlayerAchievement(playerUUID, achievementUUID)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo achievement del jugador: %w", err)
	}

	return achievement, nil
}

// UpdateAchievementProgress actualiza el progreso de un achievement
func (s *AchievementService) UpdateAchievementProgress(playerID, achievementID string, progress int) error {
	// Convertir IDs string a UUID
	playerUUID, err := uuid.Parse(playerID)
	if err != nil {
		return fmt.Errorf("playerID inválido: %w", err)
	}

	achievementUUID, err := uuid.Parse(achievementID)
	if err != nil {
		return fmt.Errorf("achievementID inválido: %w", err)
	}

	// Actualizar progreso
	if err := s.achievementRepo.UpdateAchievementProgress(playerUUID, achievementUUID, progress); err != nil {
		return fmt.Errorf("error actualizando progreso: %w", err)
	}

	// Verificar si el achievement se completó
	achievement, err := s.achievementRepo.GetAchievement(achievementUUID)
	if err != nil {
		return fmt.Errorf("error obteniendo achievement: %w", err)
	}

	if progress >= achievement.RequiredProgress {
		if err := s.completeAchievement(playerID, achievementID); err != nil {
			return fmt.Errorf("error completando achievement: %w", err)
		}
	}

	s.logger.Info("Progreso de achievement actualizado",
		zap.String("player_id", playerID),
		zap.String("achievement_id", achievementID),
		zap.Int("progress", progress),
	)

	return nil
}

// ClaimAchievementReward reclama la recompensa de un achievement
func (s *AchievementService) ClaimAchievementReward(playerID, achievementID string) error {
	// Convertir IDs string a UUID
	playerUUID, err := uuid.Parse(playerID)
	if err != nil {
		return fmt.Errorf("playerID inválido: %w", err)
	}

	achievementUUID, err := uuid.Parse(achievementID)
	if err != nil {
		return fmt.Errorf("achievementID inválido: %w", err)
	}

	// Verificar que el achievement esté completado
	playerAchievement, err := s.achievementRepo.GetPlayerAchievement(playerUUID, achievementUUID)
	if err != nil {
		return fmt.Errorf("error obteniendo achievement del jugador: %w", err)
	}

	if playerAchievement == nil || !playerAchievement.IsCompleted {
		return fmt.Errorf("el achievement no está completado")
	}

	if playerAchievement.RewardsClaimed {
		return fmt.Errorf("las recompensas ya han sido reclamadas")
	}

	// Marcar recompensas como reclamadas
	if err := s.achievementRepo.MarkRewardsAsClaimed(playerUUID, achievementUUID); err != nil {
		return fmt.Errorf("error marcando recompensas como reclamadas: %w", err)
	}

	// Obtener achievement para procesar recompensas
	achievement, err := s.achievementRepo.GetAchievement(achievementUUID)
	if err != nil {
		return fmt.Errorf("error obteniendo achievement: %w", err)
	}

	// Procesar recompensas
	if len(achievement.Rewards) > 0 {
		var rewardsPtr []*models.AchievementReward
		for i := range achievement.Rewards {
			rewardsPtr = append(rewardsPtr, &achievement.Rewards[i])
		}

		if err := s.processAchievementRewards(playerID, rewardsPtr); err != nil {
			return fmt.Errorf("error procesando recompensas: %w", err)
		}
	}

	s.logger.Info("Recompensas de achievement reclamadas",
		zap.String("player_id", playerID),
		zap.String("achievement_id", achievementID),
	)

	return nil
}

// GetAchievementStatistics obtiene las estadísticas de achievements de un jugador
func (s *AchievementService) GetAchievementStatistics(playerID string) (*models.AchievementStatistics, error) {
	// Convertir playerID string a UUID
	playerUUID, err := uuid.Parse(playerID)
	if err != nil {
		return nil, fmt.Errorf("playerID inválido: %w", err)
	}

	stats, err := s.achievementRepo.GetAchievementStatistics(playerUUID)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo estadísticas: %w", err)
	}

	return stats, nil
}

// GetAchievementLeaderboard obtiene el leaderboard de achievements
func (s *AchievementService) GetAchievementLeaderboard(limit int) ([]*models.AchievementLeaderboard, error) {
	leaderboard, err := s.achievementRepo.GetAchievementLeaderboard(limit, nil)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo leaderboard: %w", err)
	}

	// Convertir a punteros
	var leaderboardPtr []*models.AchievementLeaderboard
	for i := range leaderboard {
		leaderboardPtr = append(leaderboardPtr, &leaderboard[i])
	}

	return leaderboardPtr, nil
}

// GetAchievementNotifications obtiene las notificaciones de achievements de un jugador
func (s *AchievementService) GetAchievementNotifications(playerID string, limit int) ([]*models.AchievementNotification, error) {
	// Convertir playerID string a UUID
	playerUUID, err := uuid.Parse(playerID)
	if err != nil {
		return nil, fmt.Errorf("playerID inválido: %w", err)
	}

	notifications, err := s.achievementRepo.GetAchievementNotifications(playerUUID, false) // unreadOnly = false
	if err != nil {
		return nil, fmt.Errorf("error obteniendo notificaciones: %w", err)
	}

	// Limitar resultados si es necesario
	if limit > 0 && len(notifications) > limit {
		notifications = notifications[:limit]
	}

	// Convertir a punteros
	var notificationsPtr []*models.AchievementNotification
	for i := range notifications {
		notificationsPtr = append(notificationsPtr, &notifications[i])
	}

	return notificationsPtr, nil
}

// MarkNotificationAsRead marca una notificación como leída
func (s *AchievementService) MarkNotificationAsRead(playerID, notificationID string) error {
	// Convertir notificationID string a UUID
	notificationUUID, err := uuid.Parse(notificationID)
	if err != nil {
		return fmt.Errorf("notificationID inválido: %w", err)
	}

	// Convertir playerID string a UUID
	playerUUID, err := uuid.Parse(playerID)
	if err != nil {
		return fmt.Errorf("playerID inválido: %w", err)
	}

	if err := s.achievementRepo.MarkNotificationAsRead(playerUUID, notificationUUID); err != nil {
		return fmt.Errorf("error marcando notificación como leída: %w", err)
	}

	return nil
}

// CalculateAchievementProgress calcula el progreso de un achievement
func (s *AchievementService) CalculateAchievementProgress(playerID, achievementID string) (*models.AchievementProgress, error) {
	// Convertir IDs string a UUID
	playerUUID, err := uuid.Parse(playerID)
	if err != nil {
		return nil, fmt.Errorf("playerID inválido: %w", err)
	}

	achievementUUID, err := uuid.Parse(achievementID)
	if err != nil {
		return nil, fmt.Errorf("achievementID inválido: %w", err)
	}

	// Obtener achievement
	achievement, err := s.achievementRepo.GetAchievement(achievementUUID)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo achievement: %w", err)
	}

	// Obtener progreso actual del jugador
	playerAchievement, err := s.achievementRepo.GetPlayerAchievement(playerUUID, achievementUUID)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo progreso del jugador: %w", err)
	}

	currentProgress := 0
	if playerAchievement != nil {
		currentProgress = playerAchievement.CurrentProgress
	}

	// Calcular progreso detallado basado en el tipo de achievement
	var progressData, milestones, breakdown string
	var activityCount int

	switch achievement.ProgressType {
	case "single":
		// Achievement de una sola vez
		progressData = fmt.Sprintf(`{"current": %d, "target": %d, "type": "single"}`, currentProgress, achievement.RequiredProgress)
		if currentProgress >= achievement.RequiredProgress {
			milestones = `[{"id": 1, "name": "Completado", "achieved": true, "progress": 100}]`
		} else {
			milestones = `[{"id": 1, "name": "Completado", "achieved": false, "progress": %d}]`
			milestones = fmt.Sprintf(milestones, (currentProgress*100)/achievement.RequiredProgress)
		}
		breakdown = fmt.Sprintf(`{"total_required": %d, "current": %d, "remaining": %d}`,
			achievement.RequiredProgress, currentProgress, achievement.RequiredProgress-currentProgress)

	case "cumulative":
		// Achievement acumulativo
		progressData = fmt.Sprintf(`{"current": %d, "target": %d, "type": "cumulative"}`, currentProgress, achievement.RequiredProgress)
		progressPercent := (currentProgress * 100) / achievement.RequiredProgress
		milestones = fmt.Sprintf(`[{"id": 1, "name": "Meta", "achieved": %t, "progress": %d}]`,
			currentProgress >= achievement.RequiredProgress, progressPercent)
		breakdown = fmt.Sprintf(`{"total_required": %d, "current": %d, "remaining": %d, "percentage": %.2f}`,
			achievement.RequiredProgress, currentProgress, achievement.RequiredProgress-currentProgress, float64(progressPercent))

	case "tiered":
		// Achievement con niveles
		currentTier := 1
		if playerAchievement != nil {
			currentTier = playerAchievement.CurrentTier
		}

		// Calcular progreso del tier actual
		tierProgress := currentProgress % achievement.TargetValue
		if achievement.TargetValue == 0 {
			tierProgress = currentProgress
		}

		progressData = fmt.Sprintf(`{"current_tier": %d, "max_tier": %d, "tier_progress": %d, "type": "tiered"}`,
			currentTier, achievement.MaxTier, tierProgress)

		// Crear milestones para cada tier
		var tierMilestones []string
		for i := 1; i <= achievement.MaxTier; i++ {
			achieved := i <= currentTier
			tierMilestones = append(tierMilestones,
				fmt.Sprintf(`{"id": %d, "name": "Tier %d", "achieved": %t, "progress": %d}`,
					i, i, achieved, 100))
		}
		milestones = "[" + strings.Join(tierMilestones, ",") + "]"

		breakdown = fmt.Sprintf(`{"current_tier": %d, "max_tier": %d, "tier_progress": %d, "total_progress": %d}`,
			currentTier, achievement.MaxTier, tierProgress, currentProgress)
	}

	// Calcular actividad basada en el progreso
	if playerAchievement != nil {
		activityCount = int(time.Since(playerAchievement.StartedAt).Hours() / 24) // Días desde que empezó
		if activityCount < 1 {
			activityCount = 1
		}
	} else {
		activityCount = 1
	}

	return &models.AchievementProgress{
		ID:               achievementUUID,
		PlayerID:         playerUUID,
		CurrentProgress:  currentProgress,
		RequiredProgress: achievement.RequiredProgress,
		ProgressData:     progressData,
		Milestones:       milestones,
		Breakdown:        breakdown,
		LastUpdated:      time.Now(),
		ActivityCount:    activityCount,
	}, nil
}

// ProcessGameEvent procesa un evento del juego para actualizar achievements
func (s *AchievementService) ProcessGameEvent(playerID string, eventType string, eventData map[string]interface{}) error {
	// Convertir playerID string a UUID
	playerUUID, err := uuid.Parse(playerID)
	if err != nil {
		return fmt.Errorf("playerID inválido: %w", err)
	}

	// Obtener achievements activos del jugador
	playerAchievements, err := s.achievementRepo.GetPlayerAchievements(playerUUID, nil, false)
	if err != nil {
		return fmt.Errorf("error obteniendo achievements del jugador: %w", err)
	}

	// Procesar cada achievement
	for _, playerAchievement := range playerAchievements {
		if playerAchievement.IsCompleted {
			continue // Ya completado
		}

		achievement, err := s.achievementRepo.GetAchievement(playerAchievement.AchievementID)
		if err != nil {
			s.logger.Warn("Error obteniendo achievement para procesar evento", zap.Error(err))
			continue
		}

		// Verificar si el evento afecta a este achievement
		if s.eventAffectsAchievement(eventType, achievement) {
			if err := s.processAchievementForEvent(playerID, achievement, eventData); err != nil {
				s.logger.Warn("Error procesando achievement para evento", zap.Error(err))
			}
		}
	}

	s.logger.Info("Evento del juego procesado para achievements",
		zap.String("player_id", playerID),
		zap.String("event_type", eventType),
	)

	return nil
}

// SetWebSocketManager establece el WebSocket manager
func (s *AchievementService) SetWebSocketManager(wsManager *websocket.Manager) {
	s.wsManager = wsManager
}

// validateAchievementCategory valida una categoría de achievement
func (s *AchievementService) validateAchievementCategory(category *models.AchievementCategory) error {
	if category.Name == "" {
		return fmt.Errorf("el nombre de la categoría es requerido")
	}
	if category.Description == "" {
		return fmt.Errorf("la descripción de la categoría es requerida")
	}
	if category.Icon == "" {
		return fmt.Errorf("el icono de la categoría es requerido")
	}
	return nil
}

// completeAchievement completa un achievement
func (s *AchievementService) completeAchievement(playerID, achievementID string) error {
	// Convertir IDs string a UUID
	playerUUID, err := uuid.Parse(playerID)
	if err != nil {
		return fmt.Errorf("playerID inválido: %w", err)
	}

	achievementUUID, err := uuid.Parse(achievementID)
	if err != nil {
		return fmt.Errorf("achievementID inválido: %w", err)
	}

	// Marcar como completado
	if err := s.achievementRepo.CompleteAchievement(playerUUID, achievementUUID); err != nil {
		return fmt.Errorf("error completando achievement: %w", err)
	}

	// Obtener achievement
	achievement, err := s.achievementRepo.GetAchievement(achievementUUID)
	if err != nil {
		return fmt.Errorf("error obteniendo achievement: %w", err)
	}

	// Procesar recompensas
	if len(achievement.Rewards) > 0 {
		var rewardsPtr []*models.AchievementReward
		for i := range achievement.Rewards {
			rewardsPtr = append(rewardsPtr, &achievement.Rewards[i])
		}

		if err := s.processAchievementRewards(playerID, rewardsPtr); err != nil {
			return fmt.Errorf("error procesando recompensas: %w", err)
		}
	}

	// Enviar notificación por WebSocket
	if s.wsManager != nil {
		if err := s.sendAchievementNotification(playerID, "achievement_completed", map[string]interface{}{
			"achievement_id":   achievementID,
			"achievement_name": achievement.Name,
		}); err != nil {
			s.logger.Warn("Error enviando notificación de achievement", zap.Error(err))
		}
	}

	s.logger.Info("Achievement completado",
		zap.String("player_id", playerID),
		zap.String("achievement_id", achievementID),
		zap.String("achievement_name", achievement.Name),
	)

	return nil
}

// processAchievementRewards procesa las recompensas de un achievement
func (s *AchievementService) processAchievementRewards(playerID string, rewards []*models.AchievementReward) error {
	for _, reward := range rewards {
		switch reward.RewardType {
		case "currency":
			if err := s.grantCurrencyReward(playerID, reward); err != nil {
				return fmt.Errorf("error otorgando recompensa de moneda: %w", err)
			}
		case "experience":
			if err := s.grantExperienceReward(playerID, reward); err != nil {
				return fmt.Errorf("error otorgando recompensa de experiencia: %w", err)
			}
		case "resources":
			if err := s.grantResourceReward(playerID, reward); err != nil {
				return fmt.Errorf("error otorgando recompensa de recursos: %w", err)
			}
		case "items":
			if err := s.grantItemReward(playerID, reward); err != nil {
				return fmt.Errorf("error otorgando recompensa de items: %w", err)
			}
		case "title":
			if err := s.grantTitleReward(playerID, reward); err != nil {
				return fmt.Errorf("error otorgando recompensa de título: %w", err)
			}
		default:
			s.logger.Warn("Tipo de recompensa no reconocido",
				zap.String("reward_type", reward.RewardType),
				zap.String("player_id", playerID),
			)
		}
	}

	return nil
}

// processAchievementForEvent procesa un achievement para un evento específico
func (s *AchievementService) processAchievementForEvent(playerID string, achievement *models.Achievement, eventData map[string]interface{}) error {
	// Calcular nuevo progreso basado en el evento
	newProgress := s.calculateProgressFromEvent(achievement, eventData, 0)

	// Actualizar progreso
	if err := s.UpdateAchievementProgress(playerID, achievement.ID.String(), newProgress); err != nil {
		return fmt.Errorf("error actualizando progreso: %w", err)
	}

	return nil
}

// calculateProgressFromEvent calcula el progreso basado en un evento
func (s *AchievementService) calculateProgressFromEvent(achievement *models.Achievement, eventData map[string]interface{}, currentProgress int) int {
	// Calcular progreso basado en el tipo de achievement y el evento
	switch achievement.ProgressType {
	case "single":
		// Para achievements de una sola vez, verificar si se cumple la condición
		if s.checkSingleEventCondition(achievement, eventData) {
			return achievement.RequiredProgress
		}
		return currentProgress

	case "cumulative":
		// Para achievements acumulativos, sumar el valor del evento
		eventValue := s.extractEventValue(achievement, eventData)
		return currentProgress + eventValue

	case "tiered":
		// Para achievements con niveles, calcular progreso del tier actual
		eventValue := s.extractEventValue(achievement, eventData)
		newProgress := currentProgress + eventValue

		// Verificar si se completó el tier actual
		if newProgress >= achievement.TargetValue {
			// Avanzar al siguiente tier
			return newProgress
		}
		return newProgress

	default:
		// Por defecto, incrementar en 1
		return currentProgress + 1
	}
}

// checkSingleEventCondition verifica si se cumple la condición para un achievement de una sola vez
func (s *AchievementService) checkSingleEventCondition(achievement *models.Achievement, eventData map[string]interface{}) bool {
	// Verificar si el evento coincide con el tipo requerido
	if eventType, ok := eventData["type"].(string); ok {
		// Verificar si el achievement requiere este tipo de evento
		if achievement.ProgressFormula != "" {
			// Aquí se podría implementar parsing de fórmula JSON
			// Por ahora, verificamos coincidencia simple
			return strings.Contains(strings.ToLower(achievement.ProgressFormula), strings.ToLower(eventType))
		}
	}
	return false
}

// extractEventValue extrae el valor del evento para calcular progreso
func (s *AchievementService) extractEventValue(achievement *models.Achievement, eventData map[string]interface{}) int {
	// Extraer valor del evento basado en el tipo de achievement
	switch achievement.ProgressType {
	case "cumulative", "tiered":
		// Buscar campos comunes en eventos
		if value, ok := eventData["value"].(int); ok {
			return value
		}
		if value, ok := eventData["amount"].(int); ok {
			return value
		}
		if value, ok := eventData["quantity"].(int); ok {
			return value
		}
		// Para eventos de batalla
		if value, ok := eventData["damage"].(int); ok {
			return value
		}
		if value, ok := eventData["units_defeated"].(int); ok {
			return value
		}
		// Para eventos de construcción
		if value, ok := eventData["building_level"].(int); ok {
			return value
		}
		// Para eventos de recursos
		if value, ok := eventData["resource_amount"].(int); ok {
			return value
		}
	}

	// Valor por defecto
	return 1
}

// grantResourceReward otorga recompensa de recursos
func (s *AchievementService) grantResourceReward(playerID string, reward *models.AchievementReward) error {
	// Parsear datos de la recompensa
	var rewardData map[string]interface{}
	if err := json.Unmarshal([]byte(reward.RewardData), &rewardData); err != nil {
		return fmt.Errorf("error parseando datos de recompensa: %w", err)
	}

	// Obtener cantidad de recursos
	quantity := reward.Quantity
	if qty, ok := rewardData["quantity"].(float64); ok {
		quantity = int(qty)
	}

	// Obtener tipo de recurso
	resourceType := "gold" // Por defecto
	if rType, ok := rewardData["resource_type"].(string); ok {
		resourceType = rType
	}

	// Integrar con el sistema de recursos usando el GameMechanicsService
	// Por ahora, registramos la recompensa para implementación futura
	s.logger.Info("Recompensa de recursos otorgada",
		zap.String("player_id", playerID),
		zap.String("reward_id", reward.ID.String()),
		zap.String("resource_type", resourceType),
		zap.Int("quantity", quantity),
	)

	// TODO: Implementar integración completa cuando se tenga acceso al sistema de recursos
	// Ejemplo: s.resourceService.AddResources(playerID, resourceType, quantity)

	return nil
}

// grantExperienceReward otorga recompensa de experiencia
func (s *AchievementService) grantExperienceReward(playerID string, reward *models.AchievementReward) error {
	// Parsear datos de la recompensa
	var rewardData map[string]interface{}
	if err := json.Unmarshal([]byte(reward.RewardData), &rewardData); err != nil {
		return fmt.Errorf("error parseando datos de recompensa: %w", err)
	}

	// Obtener cantidad de experiencia
	quantity := reward.Quantity
	if qty, ok := rewardData["experience_amount"].(float64); ok {
		quantity = int(qty)
	}

	// Obtener tipo de experiencia
	expType := "general" // Por defecto
	if eType, ok := rewardData["experience_type"].(string); ok {
		expType = eType
	}

	s.logger.Info("Recompensa de experiencia otorgada",
		zap.String("player_id", playerID),
		zap.String("reward_id", reward.ID.String()),
		zap.String("experience_type", expType),
		zap.Int("amount", quantity),
	)

	// TODO: Implementar integración completa cuando se tenga acceso al sistema de experiencia
	// Ejemplo: s.experienceService.AddExperience(playerID, expType, quantity)

	return nil
}

// grantItemReward otorga recompensa de item
func (s *AchievementService) grantItemReward(playerID string, reward *models.AchievementReward) error {
	// Parsear datos de la recompensa
	var rewardData map[string]interface{}
	if err := json.Unmarshal([]byte(reward.RewardData), &rewardData); err != nil {
		return fmt.Errorf("error parseando datos de recompensa: %w", err)
	}

	// Obtener información del item
	itemID := ""
	if iID, ok := rewardData["item_id"].(string); ok {
		itemID = iID
	}

	quantity := reward.Quantity
	if qty, ok := rewardData["item_quantity"].(float64); ok {
		quantity = int(qty)
	}

	s.logger.Info("Recompensa de item otorgada",
		zap.String("player_id", playerID),
		zap.String("reward_id", reward.ID.String()),
		zap.String("item_id", itemID),
		zap.Int("quantity", quantity),
	)

	// TODO: Implementar integración completa cuando se tenga acceso al sistema de inventario
	// Ejemplo: s.inventoryService.AddItem(playerID, itemID, quantity)

	return nil
}

// grantCurrencyReward otorga recompensa de moneda
func (s *AchievementService) grantCurrencyReward(playerID string, reward *models.AchievementReward) error {
	// TODO: Implementar integración con el sistema de economía
	// Por ahora, solo registramos la recompensa
	s.logger.Info("Recompensa de moneda otorgada",
		zap.String("player_id", playerID),
		zap.String("reward_id", reward.ID.String()),
	)
	return nil
}

// grantTitleReward otorga recompensa de título
func (s *AchievementService) grantTitleReward(playerID string, reward *models.AchievementReward) error {
	// Parsear datos de la recompensa
	var rewardData map[string]interface{}
	if err := json.Unmarshal([]byte(reward.RewardData), &rewardData); err != nil {
		return fmt.Errorf("error parseando datos de recompensa: %w", err)
	}

	// Obtener información del título
	titleID := ""
	if tID, ok := rewardData["title_id"].(string); ok {
		titleID = tID
	}

	titleName := ""
	if tName, ok := rewardData["title_name"].(string); ok {
		titleName = tName
	}

	s.logger.Info("Recompensa de título otorgada",
		zap.String("player_id", playerID),
		zap.String("reward_id", reward.ID.String()),
		zap.String("title_id", titleID),
		zap.String("title_name", titleName),
	)

	// TODO: Implementar integración completa cuando se tenga acceso al sistema de títulos
	// Ejemplo: s.titleService.GrantTitle(playerID, titleID)

	return nil
}

// sendAchievementNotification envía notificación por WebSocket
func (s *AchievementService) sendAchievementNotification(playerID string, notificationType string, data map[string]interface{}) error {
	if s.wsManager == nil {
		return fmt.Errorf("WebSocket manager no configurado")
	}

	// Crear mensaje de notificación
	notification := map[string]interface{}{
		"type": "achievement_notification",
		"data": map[string]interface{}{
			"notification_type": notificationType,
			"player_id":         playerID,
			"timestamp":         time.Now().Unix(),
			"data":              data,
		},
	}

	// Enviar notificación específica al jugador
	if err := s.wsManager.SendToUser(playerID, "achievement_notification", notification); err != nil {
		return fmt.Errorf("error enviando notificación por WebSocket: %w", err)
	}

	s.logger.Info("Notificación de achievement enviada por WebSocket",
		zap.String("player_id", playerID),
		zap.String("notification_type", notificationType),
		zap.Any("data", data),
	)

	return nil
}

// eventAffectsAchievement verifica si un evento afecta a un achievement específico
func (s *AchievementService) eventAffectsAchievement(eventType string, achievement *models.Achievement) bool {
	// Verificar si el tipo de evento coincide con los requisitos del achievement
	// Esta es una implementación simplificada
	switch eventType {
	case "building_completed":
		return achievement.CategoryID != uuid.Nil // Simplificado
	case "unit_trained":
		return achievement.CategoryID != uuid.Nil // Simplificado
	case "battle_won":
		return achievement.CategoryID != uuid.Nil // Simplificado
	case "resource_collected":
		return achievement.CategoryID != uuid.Nil // Simplificado
	default:
		return false
	}
}
