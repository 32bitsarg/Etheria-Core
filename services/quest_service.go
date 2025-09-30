package services

import (
	"encoding/json"
	"fmt"
	"time"

	"server-backend/models"
	"server-backend/repository"
	"server-backend/websocket"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type QuestService struct {
	questRepo  *repository.QuestRepository
	playerRepo *repository.PlayerRepository
	wsManager  *websocket.Manager
	logger     *zap.Logger
}

func NewQuestService(
	questRepo *repository.QuestRepository,
	playerRepo *repository.PlayerRepository,
	wsManager *websocket.Manager,
	logger *zap.Logger,
) *QuestService {
	return &QuestService{
		questRepo:  questRepo,
		playerRepo: playerRepo,
		wsManager:  wsManager,
		logger:     logger,
	}
}

// GetQuestDashboard obtiene el dashboard principal de quests
func (s *QuestService) GetQuestDashboard(playerID string) (*models.QuestDashboard, error) {
	// Convertir playerID string a UUID
	playerUUID, err := uuid.Parse(playerID)
	if err != nil {
		return nil, fmt.Errorf("playerID inválido: %w", err)
	}

	// Obtener quests activas del jugador
	activePlayerQuests, err := s.questRepo.GetPlayerActiveQuests(playerUUID, nil, false)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo quests activas: %w", err)
	}

	// Convertir PlayerQuest a Quest para el dashboard
	var activeQuests []models.Quest
	for _, pq := range activePlayerQuests {
		quest, err := s.questRepo.GetQuest(pq.QuestID)
		if err != nil {
			s.logger.Warn("Error obteniendo quest para dashboard", zap.Error(err))
			continue
		}
		activeQuests = append(activeQuests, *quest)
	}

	// Obtener categorías de quests
	categories, err := s.questRepo.GetQuestCategories(true) // activeOnly = true
	if err != nil {
		return nil, fmt.Errorf("error obteniendo categorías: %w", err)
	}

	// Obtener quests disponibles
	availableQuests, err := s.questRepo.GetAvailableQuests(playerUUID, nil, 1, false)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo quests disponibles: %w", err)
	}

	// Obtener estadísticas del jugador
	playerStats, err := s.questRepo.GetPlayerQuestStatistics(playerUUID)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo estadísticas: %w", err)
	}

	return &models.QuestDashboard{
		ActiveQuests:    activeQuests,
		Categories:      categories,
		AvailableQuests: availableQuests,
		PlayerStats:     playerStats,
		LastUpdated:     time.Now(),
	}, nil
}

// GetAllQuestCategories obtiene todas las categorías de quests
func (s *QuestService) GetAllQuestCategories() ([]*models.QuestCategory, error) {
	categories, err := s.questRepo.GetAllQuestCategories()
	if err != nil {
		return nil, fmt.Errorf("error obteniendo categorías: %w", err)
	}

	// Convertir a punteros
	var categoriesPtr []*models.QuestCategory
	for i := range categories {
		categoriesPtr = append(categoriesPtr, &categories[i])
	}

	return categoriesPtr, nil
}

// GetQuestCategory obtiene una categoría específica
func (s *QuestService) GetQuestCategory(categoryID string) (*models.QuestCategory, error) {
	// Convertir categoryID string a UUID
	categoryUUID, err := uuid.Parse(categoryID)
	if err != nil {
		return nil, fmt.Errorf("categoryID inválido: %w", err)
	}

	category, err := s.questRepo.GetQuestCategory(categoryUUID)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo categoría: %w", err)
	}

	return category, nil
}

// CreateQuestCategory crea una nueva categoría de quests
func (s *QuestService) CreateQuestCategory(category *models.QuestCategory) error {
	// Validar categoría
	if err := s.validateQuestCategory(category); err != nil {
		return fmt.Errorf("categoría inválida: %w", err)
	}

	// Crear categoría
	if err := s.questRepo.CreateQuestCategory(category); err != nil {
		return fmt.Errorf("error creando categoría: %w", err)
	}

	s.logger.Info("Nueva categoría de quest creada",
		zap.String("name", category.Name),
		zap.String("category_id", category.ID.String()),
	)

	return nil
}

// GetAvailableQuests obtiene las quests disponibles para un jugador
func (s *QuestService) GetAvailableQuests(playerID string, filters map[string]interface{}) ([]*models.Quest, error) {
	// Convertir playerID string a UUID
	playerUUID, err := uuid.Parse(playerID)
	if err != nil {
		return nil, fmt.Errorf("playerID inválido: %w", err)
	}

	// Obtener nivel del jugador
	playerLevel := 1
	player, err := s.playerRepo.GetPlayerByID(playerUUID)
	if err == nil && player != nil {
		playerLevel = player.Level
	}

	var categoryID *uuid.UUID
	includeCompleted := false

	if catIDStr, ok := filters["category_id"].(string); ok {
		if catID, err := uuid.Parse(catIDStr); err == nil {
			categoryID = &catID
		}
	}
	if completed, ok := filters["include_completed"].(bool); ok {
		includeCompleted = completed
	}

	quests, err := s.questRepo.GetAvailableQuests(playerUUID, categoryID, playerLevel, includeCompleted)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo quests disponibles: %w", err)
	}

	// Convertir a punteros
	var questsPtr []*models.Quest
	for i := range quests {
		questsPtr = append(questsPtr, &quests[i])
	}

	return questsPtr, nil
}

// GetQuest obtiene una quest específica
func (s *QuestService) GetQuest(questID string) (*models.Quest, error) {
	// Convertir questID string a UUID
	questUUID, err := uuid.Parse(questID)
	if err != nil {
		return nil, fmt.Errorf("questID inválido: %w", err)
	}

	quest, err := s.questRepo.GetQuest(questUUID)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo quest: %w", err)
	}

	return quest, nil
}

// GetQuestWithDetails obtiene una quest con detalles completos
func (s *QuestService) GetQuestWithDetails(questID string) (*models.QuestWithDetails, error) {
	quest, err := s.GetQuest(questID)
	if err != nil {
		return nil, err
	}

	// Obtener categoría
	category, _ := s.questRepo.GetQuestCategory(quest.CategoryID)

	// Obtener prerequisitos
	var prerequisites []models.Quest
	if quest.Prerequisites != "" {
		// Los prerequisitos están en formato JSON string
		// Por ahora los dejamos vacíos hasta que se defina el formato
	}

	// Obtener recompensas (están en RewardsConfig como JSON string)
	var rewards []models.QuestReward
	if quest.RewardsConfig != "" {
		// Por ahora las dejamos vacías hasta que se implemente el parsing
	}

	// Obtener hitos (están en el campo Tiers como JSON string)
	var milestones []models.QuestMilestone
	if quest.Tiers != "" {
		// Por ahora los dejamos vacíos hasta que se implemente el parsing
	}

	// Obtener estadísticas
	statistics, err := s.questRepo.GetQuestStatistics(quest.ID)
	if err != nil {
		// Si no hay estadísticas, crear unas por defecto
		statistics = &models.QuestStatistics{
			TotalQuests:     0,
			CompletedQuests: 0,
			CompletionRate:  0.0,
		}
	}

	return &models.QuestWithDetails{
		Quest:         quest,
		Category:      category,
		Prerequisites: prerequisites,
		Rewards:       rewards,
		Milestones:    milestones,
		Statistics:    statistics,
	}, nil
}

// GetPlayerActiveQuests obtiene las quests activas de un jugador
func (s *QuestService) GetPlayerActiveQuests(playerID string, filters map[string]interface{}) ([]*models.PlayerQuest, error) {
	// Convertir playerID string a UUID
	playerUUID, err := uuid.Parse(playerID)
	if err != nil {
		return nil, fmt.Errorf("playerID inválido: %w", err)
	}

	var categoryID *uuid.UUID
	includeCompleted := false

	if catIDStr, ok := filters["category_id"].(string); ok {
		if catID, err := uuid.Parse(catIDStr); err == nil {
			categoryID = &catID
		}
	}
	if completed, ok := filters["include_completed"].(bool); ok {
		includeCompleted = completed
	}

	quests, err := s.questRepo.GetPlayerActiveQuests(playerUUID, categoryID, includeCompleted)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo quests activas: %w", err)
	}

	// Convertir a punteros
	var questsPtr []*models.PlayerQuest
	for i := range quests {
		questsPtr = append(questsPtr, &quests[i])
	}

	return questsPtr, nil
}

// UpdateQuestProgress actualiza el progreso de una quest
func (s *QuestService) UpdateQuestProgress(playerID, questID string, progress int, eventData map[string]interface{}) error {
	// Convertir IDs string a UUID
	playerUUID, err := uuid.Parse(playerID)
	if err != nil {
		return fmt.Errorf("playerID inválido: %w", err)
	}

	questUUID, err := uuid.Parse(questID)
	if err != nil {
		return fmt.Errorf("questID inválido: %w", err)
	}

	// Actualizar progreso usando el método del repositorio
	if err := s.questRepo.UpdateQuestProgress(playerUUID, questUUID, progress, eventData); err != nil {
		return fmt.Errorf("error actualizando progreso: %w", err)
	}

	s.logger.Info("Progreso de quest actualizado",
		zap.String("player_id", playerID),
		zap.String("quest_id", questID),
		zap.Int("progress", progress),
	)

	return nil
}

// ClaimQuestRewards reclama las recompensas de una quest completada
func (s *QuestService) ClaimQuestRewards(playerID, questID string) error {
	// Convertir IDs string a UUID
	playerUUID, err := uuid.Parse(playerID)
	if err != nil {
		return fmt.Errorf("playerID inválido: %w", err)
	}

	questUUID, err := uuid.Parse(questID)
	if err != nil {
		return fmt.Errorf("questID inválido: %w", err)
	}

	// Marcar recompensas como reclamadas
	if err := s.questRepo.MarkQuestRewardsClaimed(playerUUID, questUUID); err != nil {
		return fmt.Errorf("error marcando recompensas como reclamadas: %w", err)
	}

	s.logger.Info("Recompensas de quest reclamadas",
		zap.String("player_id", playerID),
		zap.String("quest_id", questID),
	)

	return nil
}

// ProcessGameEvent procesa un evento del juego para actualizar quests
func (s *QuestService) ProcessGameEvent(playerID string, eventType string, eventData map[string]interface{}) error {
	// Convertir playerID string a UUID
	playerUUID, err := uuid.Parse(playerID)
	if err != nil {
		return fmt.Errorf("playerID inválido: %w", err)
	}

	// Obtener quests activas del jugador
	activeQuests, err := s.questRepo.GetPlayerActiveQuests(playerUUID, nil, false)
	if err != nil {
		return fmt.Errorf("error obteniendo quests activas: %w", err)
	}

	// Procesar cada quest activa
	for _, playerQuest := range activeQuests {
		quest, err := s.questRepo.GetQuest(playerQuest.QuestID)
		if err != nil {
			s.logger.Warn("Error obteniendo quest para procesar evento", zap.Error(err))
			continue
		}

		// Verificar si el evento afecta a esta quest
		if s.eventAffectsQuest(eventType, quest) {
			if err := s.processQuestForEvent(playerID, quest, eventType, eventData); err != nil {
				s.logger.Warn("Error procesando quest para evento", zap.Error(err))
			}
		}
	}

	s.logger.Info("Evento del juego procesado",
		zap.String("player_id", playerID),
		zap.String("event_type", eventType),
	)

	return nil
}

// StartQuest inicia una quest para un jugador
func (s *QuestService) StartQuest(playerID, questID string) error {
	// Convertir IDs string a UUID
	playerUUID, err := uuid.Parse(playerID)
	if err != nil {
		return fmt.Errorf("playerID inválido: %w", err)
	}

	questUUID, err := uuid.Parse(questID)
	if err != nil {
		return fmt.Errorf("questID inválido: %w", err)
	}

	// Verificar que la quest existe y está disponible
	quest, err := s.questRepo.GetQuest(questUUID)
	if err != nil {
		return fmt.Errorf("error obteniendo quest: %w", err)
	}

	// Verificar que la quest esté activa
	if !quest.IsActive {
		return fmt.Errorf("la quest no está disponible")
	}

	// Verificar nivel requerido
	if quest.LevelRequired > 0 {
		playerLevel := 1
		player, err := s.playerRepo.GetPlayerByID(playerUUID)
		if err == nil && player != nil {
			playerLevel = player.Level
		}
		if playerLevel < quest.LevelRequired {
			return fmt.Errorf("nivel insuficiente para esta quest")
		}
	}

	// Verificar que no tenga ya esta quest activa
	existingQuests, err := s.questRepo.GetPlayerActiveQuests(playerUUID, nil, false)
	if err != nil {
		return fmt.Errorf("error verificando quests existentes: %w", err)
	}

	for _, pq := range existingQuests {
		if pq.QuestID == questUUID {
			return fmt.Errorf("ya tienes esta quest activa")
		}
	}

	// Crear PlayerQuest
	playerQuest := &models.PlayerQuest{
		PlayerID:        playerUUID,
		QuestID:         questUUID,
		CurrentProgress: 0,
		TargetProgress:  quest.RequiredProgress,
		ProgressPercent: 0.0,
		IsCompleted:     false,
		IsClaimed:       false,
		IsFailed:        false,
		CurrentTier:     1,
		CompletionCount: 0,
		RewardsClaimed:  false,
		PointsEarned:    0,
		StartedAt:       time.Now(),
		LastUpdated:     time.Now(),
		CreatedAt:       time.Now(),
	}

	if err := s.questRepo.CreatePlayerQuest(playerQuest); err != nil {
		return fmt.Errorf("error creando quest del jugador: %w", err)
	}

	s.logger.Info("Quest iniciada",
		zap.String("player_id", playerID),
		zap.String("quest_id", questID),
		zap.String("quest_name", quest.Name),
	)

	return nil
}

// AbandonQuest abandona una quest
func (s *QuestService) AbandonQuest(playerID, questID string) error {
	// Convertir IDs string a UUID
	playerUUID, err := uuid.Parse(playerID)
	if err != nil {
		return fmt.Errorf("playerID inválido: %w", err)
	}

	questUUID, err := uuid.Parse(questID)
	if err != nil {
		return fmt.Errorf("questID inválido: %w", err)
	}

	// Obtener la quest del jugador
	playerQuest, err := s.questRepo.GetPlayerQuest(playerUUID, questUUID)
	if err != nil {
		return fmt.Errorf("error obteniendo quest del jugador: %w", err)
	}

	if playerQuest == nil {
		return fmt.Errorf("no tienes esta quest activa")
	}

	// Marcar como fallida
	playerQuest.IsFailed = true
	playerQuest.FailedAt = &time.Time{}
	*playerQuest.FailedAt = time.Now()
	playerQuest.LastUpdated = time.Now()

	// Actualizar en el repositorio
	if err := s.questRepo.UpdatePlayerQuest(playerQuest); err != nil {
		return fmt.Errorf("error actualizando quest: %w", err)
	}

	s.logger.Info("Quest abandonada",
		zap.String("player_id", playerID),
		zap.String("quest_id", questID),
	)

	return nil
}

// GetQuestHistory obtiene el historial de quests de un jugador
func (s *QuestService) GetQuestHistory(playerID string, limit int) ([]*models.PlayerQuest, error) {
	// Convertir playerID string a UUID
	playerUUID, err := uuid.Parse(playerID)
	if err != nil {
		return nil, fmt.Errorf("playerID inválido: %w", err)
	}

	// Obtener historial del repositorio
	history, err := s.questRepo.GetPlayerQuestHistory(playerUUID, limit)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo historial: %w", err)
	}

	// Convertir a punteros
	var historyPtr []*models.PlayerQuest
	for i := range history {
		historyPtr = append(historyPtr, &history[i])
	}

	s.logger.Info("Historial de quests obtenido",
		zap.String("player_id", playerID),
		zap.Int("limit", limit),
		zap.Int("count", len(historyPtr)),
	)

	return historyPtr, nil
}

// GetDailyQuests obtiene las quests diarias de un jugador
func (s *QuestService) GetDailyQuests(playerID string) ([]*models.Quest, error) {
	// Convertir playerID string a UUID
	playerUUID, err := uuid.Parse(playerID)
	if err != nil {
		return nil, fmt.Errorf("playerID inválido: %w", err)
	}

	// Obtener quests diarias del repositorio
	quests, err := s.questRepo.GetDailyQuests(playerUUID)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo quests diarias: %w", err)
	}

	// Convertir a punteros
	var questsPtr []*models.Quest
	for i := range quests {
		questsPtr = append(questsPtr, &quests[i])
	}

	s.logger.Info("Quests diarias obtenidas",
		zap.String("player_id", playerID),
		zap.Int("count", len(questsPtr)),
	)

	return questsPtr, nil
}

// RefreshDailyQuests refresca las quests diarias de un jugador
func (s *QuestService) RefreshDailyQuests(playerID string) error {
	// Convertir playerID string a UUID
	playerUUID, err := uuid.Parse(playerID)
	if err != nil {
		return fmt.Errorf("playerID inválido: %w", err)
	}

	// Refrescar quests diarias en el repositorio
	if err := s.questRepo.RefreshDailyQuests(playerUUID); err != nil {
		return fmt.Errorf("error refrescando quests diarias: %w", err)
	}

	s.logger.Info("Quests diarias refrescadas",
		zap.String("player_id", playerID),
	)

	return nil
}

// SetWebSocketManager establece el WebSocket manager
func (s *QuestService) SetWebSocketManager(wsManager *websocket.Manager) {
	s.wsManager = wsManager
}

// Métodos privados

func (s *QuestService) validateQuestCategory(category *models.QuestCategory) error {
	if category == nil {
		return fmt.Errorf("categoría no puede ser nula")
	}

	if category.Name == "" {
		return fmt.Errorf("nombre de categoría es requerido")
	}

	if category.DisplayOrder < 0 {
		return fmt.Errorf("orden de visualización debe ser mayor o igual a 0")
	}

	return nil
}

// completeQuest completa una quest
func (s *QuestService) completeQuest(playerID, questID string) error {
	// Convertir IDs string a UUID
	playerUUID, err := uuid.Parse(playerID)
	if err != nil {
		return fmt.Errorf("playerID inválido: %w", err)
	}

	questUUID, err := uuid.Parse(questID)
	if err != nil {
		return fmt.Errorf("questID inválido: %w", err)
	}

	// Obtener quest
	quest, err := s.GetQuest(questID)
	if err != nil {
		return fmt.Errorf("error obteniendo quest: %w", err)
	}

	// Obtener la quest del jugador
	playerQuest, err := s.questRepo.GetPlayerQuest(playerUUID, questUUID)
	if err != nil {
		return fmt.Errorf("error obteniendo quest del jugador: %w", err)
	}

	if playerQuest == nil {
		return fmt.Errorf("no tienes esta quest activa")
	}

	// Marcar como completada
	playerQuest.IsCompleted = true
	playerQuest.CompletionTime = &time.Time{}
	*playerQuest.CompletionTime = time.Now()
	playerQuest.LastUpdated = time.Now()

	// Actualizar en el repositorio
	if err := s.questRepo.UpdatePlayerQuest(playerQuest); err != nil {
		return fmt.Errorf("error actualizando quest: %w", err)
	}

	s.logger.Info("Quest completada",
		zap.String("player_id", playerID),
		zap.String("quest_id", questID),
		zap.String("quest_name", quest.Name),
	)

	return nil
}

// processQuestRewards procesa las recompensas de una quest
func (s *QuestService) processQuestRewards(playerID, questID string) error {
	// Convertir IDs string a UUID
	playerUUID, err := uuid.Parse(playerID)
	if err != nil {
		return fmt.Errorf("playerID inválido: %w", err)
	}

	questUUID, err := uuid.Parse(questID)
	if err != nil {
		return fmt.Errorf("questID inválido: %w", err)
	}

	// Obtener recompensas de la quest
	rewards, err := s.questRepo.GetQuestRewards(questUUID)
	if err != nil {
		return fmt.Errorf("error obteniendo recompensas: %w", err)
	}

	// Procesar cada recompensa
	for _, reward := range rewards {
		if err := s.grantQuestReward(playerID, &reward); err != nil {
			s.logger.Warn("Error otorgando recompensa de quest",
				zap.String("player_id", playerID),
				zap.String("quest_id", questID),
				zap.String("reward_id", reward.ID.String()),
				zap.Error(err),
			)
		}
	}

	// Marcar recompensas como reclamadas
	if err := s.questRepo.MarkRewardsAsClaimed(playerUUID, questUUID); err != nil {
		return fmt.Errorf("error marcando recompensas como reclamadas: %w", err)
	}

	s.logger.Info("Recompensas de quest procesadas",
		zap.String("player_id", playerID),
		zap.String("quest_id", questID),
		zap.Int("rewards_count", len(rewards)),
	)

	return nil
}

// grantQuestReward otorga una recompensa de quest a un jugador
func (s *QuestService) grantQuestReward(playerID string, reward *models.QuestReward) error {
	// Parsear datos de la recompensa
	var rewardData map[string]interface{}
	if err := json.Unmarshal([]byte(reward.RewardData), &rewardData); err != nil {
		return fmt.Errorf("error parseando datos de recompensa: %w", err)
	}

	// Procesar recompensa según el tipo
	switch reward.RewardType {
	case "currency":
		if err := s.grantCurrencyReward(playerID, reward, rewardData); err != nil {
			return fmt.Errorf("error otorgando recompensa de moneda: %w", err)
		}
	case "experience":
		if err := s.grantExperienceReward(playerID, reward, rewardData); err != nil {
			return fmt.Errorf("error otorgando recompensa de experiencia: %w", err)
		}
	case "resources":
		if err := s.grantResourceReward(playerID, reward, rewardData); err != nil {
			return fmt.Errorf("error otorgando recompensa de recursos: %w", err)
		}
	case "items":
		if err := s.grantItemReward(playerID, reward, rewardData); err != nil {
			return fmt.Errorf("error otorgando recompensa de items: %w", err)
		}
	case "title":
		if err := s.grantTitleReward(playerID, reward, rewardData); err != nil {
			return fmt.Errorf("error otorgando recompensa de título: %w", err)
		}
	default:
		s.logger.Warn("Tipo de recompensa no reconocido",
			zap.String("reward_type", reward.RewardType),
			zap.String("player_id", playerID),
		)
	}

	return nil
}

// grantCurrencyReward otorga recompensa de moneda
func (s *QuestService) grantCurrencyReward(playerID string, reward *models.QuestReward, rewardData map[string]interface{}) error {
	// TODO: Implementar integración con el sistema de economía
	s.logger.Info("Recompensa de moneda otorgada",
		zap.String("player_id", playerID),
		zap.String("reward_id", reward.ID.String()),
		zap.Int("quantity", reward.Quantity),
	)
	return nil
}

// grantExperienceReward otorga recompensa de experiencia
func (s *QuestService) grantExperienceReward(playerID string, reward *models.QuestReward, rewardData map[string]interface{}) error {
	// TODO: Implementar integración con el sistema de experiencia
	s.logger.Info("Recompensa de experiencia otorgada",
		zap.String("player_id", playerID),
		zap.String("reward_id", reward.ID.String()),
		zap.Int("quantity", reward.Quantity),
	)
	return nil
}

// grantResourceReward otorga recompensa de recursos
func (s *QuestService) grantResourceReward(playerID string, reward *models.QuestReward, rewardData map[string]interface{}) error {
	// TODO: Implementar integración con el sistema de recursos
	s.logger.Info("Recompensa de recursos otorgada",
		zap.String("player_id", playerID),
		zap.String("reward_id", reward.ID.String()),
		zap.Int("quantity", reward.Quantity),
	)
	return nil
}

// grantItemReward otorga recompensa de item
func (s *QuestService) grantItemReward(playerID string, reward *models.QuestReward, rewardData map[string]interface{}) error {
	// TODO: Implementar integración con el sistema de inventario
	s.logger.Info("Recompensa de item otorgada",
		zap.String("player_id", playerID),
		zap.String("reward_id", reward.ID.String()),
		zap.Int("quantity", reward.Quantity),
	)
	return nil
}

// grantTitleReward otorga recompensa de título
func (s *QuestService) grantTitleReward(playerID string, reward *models.QuestReward, rewardData map[string]interface{}) error {
	// TODO: Implementar integración con el sistema de títulos
	s.logger.Info("Recompensa de título otorgada",
		zap.String("player_id", playerID),
		zap.String("reward_id", reward.ID.String()),
		zap.Int("quantity", reward.Quantity),
	)
	return nil
}

// processQuestForEvent procesa una quest para un evento específico
func (s *QuestService) processQuestForEvent(playerID string, quest *models.Quest, eventType string, eventData map[string]interface{}) error {
	// Verificar si el evento coincide con los requisitos de la quest
	if s.eventMatchesQuestRequirements(eventType, quest, eventData) {
		// Actualizar progreso
		if err := s.UpdateQuestProgress(playerID, quest.ID.String(), 1, eventData); err != nil {
			return fmt.Errorf("error actualizando progreso: %w", err)
		}
	}

	return nil
}

// verifyQuestRequirements verifica que el jugador cumple los requisitos para una quest
func (s *QuestService) verifyQuestRequirements(playerID, questID string) error {
	// Obtener quest
	quest, err := s.GetQuest(questID)
	if err != nil {
		return fmt.Errorf("error obteniendo quest: %w", err)
	}

	// Verificar nivel requerido
	if quest.LevelRequired > 0 {
		playerLevel := 1
		playerUUID, err := uuid.Parse(playerID)
		if err == nil {
			player, err := s.playerRepo.GetPlayerByID(playerUUID)
			if err == nil && player != nil {
				playerLevel = player.Level
			}
		}
		if playerLevel < quest.LevelRequired {
			return fmt.Errorf("nivel insuficiente: requiere nivel %d, tienes nivel %d", quest.LevelRequired, playerLevel)
		}
	}

	// TODO: Verificar prerequisitos cuando se implemente en el repositorio
	// Por ahora, asumimos que no hay prerequisitos

	return nil
}

// eventAffectsQuest verifica si un evento afecta a una quest específica
func (s *QuestService) eventAffectsQuest(eventType string, quest *models.Quest) bool {
	// Verificar si el tipo de evento coincide con los requisitos de la quest
	// Esta es una implementación simplificada
	switch eventType {
	case "building_completed":
		return quest.QuestType == "building" || quest.QuestType == "construction"
	case "unit_trained":
		return quest.QuestType == "training" || quest.QuestType == "military"
	case "battle_won":
		return quest.QuestType == "combat" || quest.QuestType == "military"
	case "resource_collected":
		return quest.QuestType == "gathering" || quest.QuestType == "economy"
	default:
		return false
	}
}

// eventMatchesQuestRequirements verifica si un evento coincide con los requisitos de una quest
func (s *QuestService) eventMatchesQuestRequirements(eventType string, quest *models.Quest, eventData map[string]interface{}) bool {
	// Implementación simplificada - en un sistema real, esto sería más complejo
	return s.eventAffectsQuest(eventType, quest)
}
