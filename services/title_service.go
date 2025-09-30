package services

import (
	"fmt"
	"server-backend/models"
	"server-backend/repository"
	"server-backend/websocket"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type TitleService struct {
	titleRepo  *repository.TitleRepository
	playerRepo *repository.PlayerRepository
	wsManager  *websocket.Manager
	logger     *zap.Logger
}

func NewTitleService(
	titleRepo *repository.TitleRepository,
	playerRepo *repository.PlayerRepository,
	wsManager *websocket.Manager,
	logger *zap.Logger,
) *TitleService {
	return &TitleService{
		titleRepo:  titleRepo,
		playerRepo: playerRepo,
		wsManager:  wsManager,
		logger:     logger,
	}
}

// GetTitleDashboard obtiene el dashboard principal de títulos
func (s *TitleService) GetTitleDashboard(playerID string) (*models.TitleDashboard, error) {
	// Convertir playerID string a UUID
	playerUUID, err := uuid.Parse(playerID)
	if err != nil {
		return nil, fmt.Errorf("playerID inválido: %w", err)
	}

	// Obtener títulos equipados del jugador
	equippedTitles, err := s.titleRepo.GetPlayerTitles(playerUUID, nil, true) // equippedOnly = true
	if err != nil {
		return nil, fmt.Errorf("error obteniendo títulos equipados: %w", err)
	}

	// Obtener títulos recientemente desbloqueados
	recentUnlocks, err := s.titleRepo.GetPlayerTitles(playerUUID, nil, false)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo títulos recientes: %w", err)
	}

	// Obtener categorías de títulos
	categories, err := s.titleRepo.GetTitleCategories(true) // activeOnly = true
	if err != nil {
		return nil, fmt.Errorf("error obteniendo categorías: %w", err)
	}

	// Obtener notificaciones
	notifications, err := s.titleRepo.GetPlayerTitleNotifications(playerUUID)
	if err != nil {
		s.logger.Warn("Error obteniendo notificaciones", zap.Error(err))
		notifications = []models.TitleNotification{}
	}

	// TODO: Implementar obtención de otros datos del dashboard
	return &models.TitleDashboard{
		EquippedTitles: equippedTitles,
		RecentUnlocks:  recentUnlocks,
		Categories:     categories,
		Notifications:  notifications,
		LastUpdated:    time.Now(),
	}, nil
}

// GetTitleCategories obtiene todas las categorías de títulos
func (s *TitleService) GetTitleCategories() ([]*models.TitleCategory, error) {
	categories, err := s.titleRepo.GetTitleCategories(true) // activeOnly = true
	if err != nil {
		return nil, fmt.Errorf("error obteniendo categorías: %w", err)
	}

	// Convertir a punteros
	var categoriesPtr []*models.TitleCategory
	for i := range categories {
		categoriesPtr = append(categoriesPtr, &categories[i])
	}

	return categoriesPtr, nil
}

// GetTitleCategory obtiene una categoría específica
func (s *TitleService) GetTitleCategory(categoryID string) (*models.TitleCategory, error) {
	// Convertir categoryID string a UUID
	categoryUUID, err := uuid.Parse(categoryID)
	if err != nil {
		return nil, fmt.Errorf("categoryID inválido: %w", err)
	}

	category, err := s.titleRepo.GetTitleCategory(categoryUUID)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo categoría: %w", err)
	}

	return category, nil
}

// CreateTitleCategory crea una nueva categoría de títulos
func (s *TitleService) CreateTitleCategory(category *models.TitleCategory) error {
	// Validar categoría
	if err := s.validateTitleCategory(category); err != nil {
		return fmt.Errorf("categoría inválida: %w", err)
	}

	// Crear categoría
	if err := s.titleRepo.CreateTitleCategory(category); err != nil {
		return fmt.Errorf("error creando categoría: %w", err)
	}

	s.logger.Info("Nueva categoría de título creada",
		zap.String("name", category.Name),
		zap.String("category_id", category.ID.String()),
	)

	return nil
}

// GetTitles obtiene todos los títulos
func (s *TitleService) GetTitles(filters map[string]interface{}) ([]*models.Title, error) {
	// Convertir filtros de string a UUID si es necesario
	if categoryIDStr, ok := filters["category_id"].(string); ok {
		if categoryID, err := uuid.Parse(categoryIDStr); err == nil {
			filters["category_id"] = categoryID
		}
	}

	titles, err := s.titleRepo.GetTitles(filters)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo títulos: %w", err)
	}

	// Convertir a punteros
	var titlesPtr []*models.Title
	for i := range titles {
		titlesPtr = append(titlesPtr, &titles[i])
	}

	return titlesPtr, nil
}

// GetTitle obtiene un título específico
func (s *TitleService) GetTitle(titleID string) (*models.Title, error) {
	// Convertir titleID string a UUID
	titleUUID, err := uuid.Parse(titleID)
	if err != nil {
		return nil, fmt.Errorf("titleID inválido: %w", err)
	}

	title, err := s.titleRepo.GetTitle(titleUUID)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo título: %w", err)
	}

	return title, nil
}

// CreateTitle crea un nuevo título
func (s *TitleService) CreateTitle(title *models.Title) error {
	// Validar título
	if err := s.validateTitle(title); err != nil {
		return fmt.Errorf("título inválido: %w", err)
	}

	// Crear título
	if err := s.titleRepo.CreateTitle(title); err != nil {
		return fmt.Errorf("error creando título: %w", err)
	}

	s.logger.Info("Nuevo título creado",
		zap.String("name", title.Name),
		zap.String("title_id", title.ID.String()),
	)

	return nil
}

// GetPlayerTitles obtiene los títulos de un jugador
func (s *TitleService) GetPlayerTitles(playerID string, filters map[string]interface{}) ([]*models.PlayerTitle, error) {
	// Convertir playerID string a UUID
	playerUUID, err := uuid.Parse(playerID)
	if err != nil {
		return nil, fmt.Errorf("playerID inválido: %w", err)
	}

	var categoryID *uuid.UUID
	equippedOnly := false

	if catIDStr, ok := filters["category_id"].(string); ok {
		if catID, err := uuid.Parse(catIDStr); err == nil {
			categoryID = &catID
		}
	}
	if equipped, ok := filters["equipped_only"].(bool); ok {
		equippedOnly = equipped
	}

	titles, err := s.titleRepo.GetPlayerTitles(playerUUID, categoryID, equippedOnly)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo títulos del jugador: %w", err)
	}

	// Convertir a punteros
	var titlesPtr []*models.PlayerTitle
	for i := range titles {
		titlesPtr = append(titlesPtr, &titles[i])
	}

	return titlesPtr, nil
}

// GrantTitle otorga un título a un jugador
func (s *TitleService) GrantTitle(playerID, titleID string, reason string) error {
	// Convertir IDs string a UUID
	playerUUID, err := uuid.Parse(playerID)
	if err != nil {
		return fmt.Errorf("playerID inválido: %w", err)
	}

	titleUUID, err := uuid.Parse(titleID)
	if err != nil {
		return fmt.Errorf("titleID inválido: %w", err)
	}

	// Obtener el título para verificar requisitos
	title, err := s.titleRepo.GetTitle(titleUUID)
	if err != nil {
		return fmt.Errorf("error obteniendo título: %w", err)
	}

	// Verificar requisitos
	if err := s.verifyTitleRequirements(playerUUID, title); err != nil {
		return fmt.Errorf("no cumple requisitos para el título: %w", err)
	}

	// Otorgar el título
	if err := s.titleRepo.GrantTitle(playerUUID, titleUUID, reason); err != nil {
		return fmt.Errorf("error otorgando título: %w", err)
	}

	// Enviar notificación
	if err := s.sendTitleNotification(playerID, &titleID, "title_unlocked", map[string]interface{}{
		"title_id":   titleID,
		"title_name": title.Name,
		"reason":     reason,
	}); err != nil {
		s.logger.Warn("Error enviando notificación de título", zap.Error(err))
	}

	s.logger.Info("Título otorgado",
		zap.String("player_id", playerID),
		zap.String("title_id", titleID),
		zap.String("reason", reason),
	)

	return nil
}

// EquipTitle equipa un título
func (s *TitleService) EquipTitle(playerID, titleID string) error {
	// Convertir IDs string a UUID
	playerUUID, err := uuid.Parse(playerID)
	if err != nil {
		return fmt.Errorf("playerID inválido: %w", err)
	}

	titleUUID, err := uuid.Parse(titleID)
	if err != nil {
		return fmt.Errorf("titleID inválido: %w", err)
	}

	// Verificar que el jugador tiene el título
	playerTitles, err := s.titleRepo.GetPlayerTitles(playerUUID, nil, false)
	if err != nil {
		return fmt.Errorf("error verificando títulos del jugador: %w", err)
	}

	hasTitle := false
	for _, pt := range playerTitles {
		if pt.TitleID == titleUUID {
			hasTitle = true
			break
		}
	}

	if !hasTitle {
		return fmt.Errorf("el jugador no tiene este título")
	}

	// Equipar el título
	if err := s.titleRepo.EquipTitle(playerUUID, titleUUID); err != nil {
		return fmt.Errorf("error equipando título: %w", err)
	}

	// Enviar notificación
	if err := s.sendTitleNotification(playerID, nil, "title_equipped", map[string]interface{}{
		"title_id": titleID,
		"equipped": true,
	}); err != nil {
		s.logger.Warn("Error enviando notificación de equipamiento", zap.Error(err))
	}

	s.logger.Info("Título equipado",
		zap.String("player_id", playerID),
		zap.String("title_id", titleID),
	)

	return nil
}

// UnequipTitle desequipa el título actual
func (s *TitleService) UnequipTitle(playerID string) error {
	// Convertir playerID string a UUID
	playerUUID, err := uuid.Parse(playerID)
	if err != nil {
		return fmt.Errorf("playerID inválido: %w", err)
	}

	// Desequipar título
	if err := s.titleRepo.UnequipTitle(playerUUID); err != nil {
		return fmt.Errorf("error desequipando título: %w", err)
	}

	// Enviar notificación
	if err := s.sendTitleNotification(playerID, nil, "title_unequipped", map[string]interface{}{
		"equipped": false,
	}); err != nil {
		s.logger.Warn("Error enviando notificación de desequipamiento", zap.Error(err))
	}

	s.logger.Info("Título desequipado",
		zap.String("player_id", playerID),
	)

	return nil
}

// GetTitleLeaderboard obtiene el leaderboard de títulos
func (s *TitleService) GetTitleLeaderboard(limit int) ([]*models.TitleLeaderboard, error) {
	leaderboard, err := s.titleRepo.GetTitleLeaderboard(limit)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo leaderboard: %w", err)
	}

	// Convertir a punteros
	var leaderboardPtr []*models.TitleLeaderboard
	for i := range leaderboard {
		leaderboardPtr = append(leaderboardPtr, &leaderboard[i])
	}

	return leaderboardPtr, nil
}

// GetTitleStatistics obtiene las estadísticas de títulos
func (s *TitleService) GetTitleStatistics() (*models.TitleStatistics, error) {
	stats, err := s.titleRepo.GetTitleStatistics()
	if err != nil {
		return nil, fmt.Errorf("error obteniendo estadísticas: %w", err)
	}

	return stats, nil
}

// SetWebSocketManager establece el WebSocket manager
func (s *TitleService) SetWebSocketManager(wsManager *websocket.Manager) {
	s.wsManager = wsManager
}

// validateTitleCategory valida una categoría de título
func (s *TitleService) validateTitleCategory(category *models.TitleCategory) error {
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

// validateTitle valida un título
func (s *TitleService) validateTitle(title *models.Title) error {
	if title.Name == "" {
		return fmt.Errorf("el nombre del título es requerido")
	}
	if title.Description == "" {
		return fmt.Errorf("la descripción del título es requerida")
	}
	if title.Rarity == "" {
		return fmt.Errorf("la rareza del título es requerida")
	}
	// Validar rareza válida
	validRarities := []string{"common", "rare", "epic", "legendary", "mythic", "divine"}
	isValid := false
	for _, rarity := range validRarities {
		if title.Rarity == rarity {
			isValid = true
			break
		}
	}
	if !isValid {
		return fmt.Errorf("rareza inválida: %s", title.Rarity)
	}
	return nil
}

// verifyTitleRequirements verifica que el jugador cumple los requisitos para un título
func (s *TitleService) verifyTitleRequirements(playerID uuid.UUID, title *models.Title) error {
	// Obtener datos del jugador (por ahora solo verificamos que existe)
	_, err := s.playerRepo.GetPlayerByID(playerID)
	if err != nil {
		return fmt.Errorf("error obteniendo datos del jugador: %w", err)
	}

	// Verificar nivel requerido
	if title.LevelRequired > 0 {
		player, err := s.playerRepo.GetPlayerByID(playerID)
		if err != nil {
			return fmt.Errorf("error obteniendo información del jugador: %w", err)
		}
		if player == nil {
			return fmt.Errorf("jugador no encontrado")
		}

		playerLevel := player.Level
		if playerLevel < title.LevelRequired {
			return fmt.Errorf("nivel insuficiente: requiere nivel %d, tienes nivel %d", title.LevelRequired, playerLevel)
		}
	}

	// Verificar prestigio requerido
	if title.PrestigeRequired > 0 {
		prestige, err := s.titleRepo.GetPlayerPrestige(playerID)
		if err != nil {
			return fmt.Errorf("error obteniendo prestigio del jugador: %w", err)
		}
		if prestige.CurrentPrestige < title.PrestigeRequired {
			return fmt.Errorf("prestigio insuficiente: requiere %d, tiene %d", title.PrestigeRequired, prestige.CurrentPrestige)
		}
	}

	// Verificar alianza requerida
	if title.AllianceRequired != nil {
		// TODO: Implementar verificación de alianza cuando se complete el sistema de alianzas
		// Por ahora, asumimos que no hay requisito de alianza
	}

	// Verificar si el título ya está desbloqueado
	playerTitles, err := s.titleRepo.GetPlayerTitles(playerID, nil, false)
	if err != nil {
		return fmt.Errorf("error verificando títulos del jugador: %w", err)
	}

	for _, pt := range playerTitles {
		if pt.TitleID == title.ID {
			return fmt.Errorf("el jugador ya tiene este título")
		}
	}

	// Verificar límite de propietarios
	if title.MaxOwners > 0 {
		// TODO: Implementar verificación de límite de propietarios cuando se complete el repositorio
		s.logger.Info("Verificación de límite de propietarios pendiente",
			zap.String("title_id", title.ID.String()),
			zap.Int("max_owners", title.MaxOwners),
		)
	}

	return nil
}

// sendTitleNotification envía una notificación de título
func (s *TitleService) sendTitleNotification(playerID string, titleID *string, notificationType string, data map[string]interface{}) error {
	if s.wsManager == nil {
		s.logger.Warn("WebSocket Manager no disponible para notificaciones de título")
		return nil
	}

	// Crear mensaje de notificación
	message := map[string]interface{}{
		"type": "title_notification",
		"data": map[string]interface{}{
			"notification_type": notificationType,
			"title_id":          titleID,
			"timestamp":         time.Now().Unix(),
			"data":              data,
		},
	}

	// Enviar notificación por WebSocket
	if err := s.wsManager.SendToUser(playerID, "title_notification", message); err != nil {
		s.logger.Warn("Error enviando notificación de título por WebSocket",
			zap.String("player_id", playerID),
			zap.String("notification_type", notificationType),
			zap.Error(err),
		)
		return err
	}

	s.logger.Info("Notificación de título enviada por WebSocket",
		zap.String("player_id", playerID),
		zap.String("notification_type", notificationType),
		zap.Any("data", data),
	)

	return nil
}

// processTitleRewards procesa las recompensas de un título
func (s *TitleService) processTitleRewards(playerID string, rewards []*models.TitleReward) error {
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
		case "items":
			if err := s.grantItemReward(playerID, reward); err != nil {
				return fmt.Errorf("error otorgando recompensa de items: %w", err)
			}
		case "resources":
			if err := s.grantResourceReward(playerID, reward); err != nil {
				return fmt.Errorf("error otorgando recompensa de recursos: %w", err)
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

// grantCurrencyReward otorga recompensa de moneda
func (s *TitleService) grantCurrencyReward(playerID string, reward *models.TitleReward) error {
	// TODO: Implementar integración con el sistema de economía
	// Por ahora, solo registramos la recompensa
	s.logger.Info("Recompensa de moneda otorgada",
		zap.String("player_id", playerID),
		zap.String("reward_id", reward.ID.String()),
		zap.String("reward_type", reward.RewardType),
		zap.Int("quantity", reward.Quantity),
	)
	return nil
}

// grantExperienceReward otorga recompensa de experiencia
func (s *TitleService) grantExperienceReward(playerID string, reward *models.TitleReward) error {
	// TODO: Implementar integración con el sistema de experiencia
	// Por ahora, solo registramos la recompensa
	s.logger.Info("Recompensa de experiencia otorgada",
		zap.String("player_id", playerID),
		zap.String("reward_id", reward.ID.String()),
		zap.String("reward_type", reward.RewardType),
		zap.Int("quantity", reward.Quantity),
	)
	return nil
}

// grantItemReward otorga recompensa de items
func (s *TitleService) grantItemReward(playerID string, reward *models.TitleReward) error {
	// TODO: Implementar integración con el sistema de inventario
	// Por ahora, solo registramos la recompensa
	s.logger.Info("Recompensa de items otorgada",
		zap.String("player_id", playerID),
		zap.String("reward_id", reward.ID.String()),
		zap.String("reward_type", reward.RewardType),
		zap.Int("quantity", reward.Quantity),
	)
	return nil
}

// grantResourceReward otorga recompensa de recursos
func (s *TitleService) grantResourceReward(playerID string, reward *models.TitleReward) error {
	// TODO: Implementar integración con el sistema de recursos
	// Por ahora, solo registramos la recompensa
	s.logger.Info("Recompensa de recursos otorgada",
		zap.String("player_id", playerID),
		zap.String("reward_id", reward.ID.String()),
		zap.String("reward_type", reward.RewardType),
		zap.Int("quantity", reward.Quantity),
	)
	return nil
}
