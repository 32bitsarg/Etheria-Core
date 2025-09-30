package services

import (
	"fmt"
	"time"

	"server-backend/models"
	"server-backend/repository"
	"server-backend/websocket"

	"go.uber.org/zap"
)

type NotificationService struct {
	notificationRepo *repository.NotificationRepository
	playerRepo       *repository.PlayerRepository
	wsManager        *websocket.Manager
	logger           *zap.Logger
	redisService     *RedisService
}

func NewNotificationService(
	notificationRepo *repository.NotificationRepository,
	playerRepo *repository.PlayerRepository,
	wsManager *websocket.Manager,
	logger *zap.Logger,
	redisService *RedisService,
) *NotificationService {
	return &NotificationService{
		notificationRepo: notificationRepo,
		playerRepo:       playerRepo,
		wsManager:        wsManager,
		logger:           logger,
		redisService:     redisService,
	}
}

// GetPlayerNotifications obtiene las notificaciones de un jugador
func (s *NotificationService) GetPlayerNotifications(playerID string, limit int) ([]*models.Notification, error) {
	// Primero intentar obtener de Redis
	if s.redisService != nil {
		notifs, err := s.redisService.GetNotifications(playerID, limit)
		if err == nil && len(notifs) > 0 {
			return notifs, nil
		}
	}
	// Si no está en Redis, obtener de la base de datos y cachear
	notifications, err := s.notificationRepo.GetPlayerNotifications(playerID, limit)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo notificaciones: %w", err)
	}
	// Guardar en Redis
	if s.redisService != nil {
		for _, n := range notifications {
			s.redisService.AddNotification(playerID, n)
		}
	}
	return notifications, nil
}

// MarkNotificationAsRead marca una notificación como leída
func (s *NotificationService) MarkNotificationAsRead(playerID, notificationID string) error {
	if err := s.notificationRepo.MarkNotificationAsRead(playerID, notificationID); err != nil {
		return fmt.Errorf("error marcando notificación como leída: %w", err)
	}

	return nil
}

// DeleteNotification elimina una notificación
func (s *NotificationService) DeleteNotification(playerID, notificationID string) error {
	if err := s.notificationRepo.DeleteNotification(playerID, notificationID); err != nil {
		return fmt.Errorf("error eliminando notificación: %w", err)
	}

	return nil
}

// CreateNotification crea una nueva notificación
func (s *NotificationService) CreateNotification(notification *models.Notification) error {
	// Validar notificación
	if err := s.validateNotification(notification); err != nil {
		return fmt.Errorf("notificación inválida: %w", err)
	}

	// Crear notificación en la base de datos
	if err := s.notificationRepo.CreateNotification(notification); err != nil {
		return fmt.Errorf("error creando notificación: %w", err)
	}

	// Guardar en Redis
	if s.redisService != nil {
		s.redisService.AddNotification(notification.PlayerID, notification)
	}

	// Enviar por WebSocket
	if s.wsManager != nil {
		s.logger.Info("Enviando notificación por WebSocket",
			zap.String("player_id", notification.PlayerID),
			zap.String("notification_id", notification.ID),
			zap.String("type", notification.Type),
		)
	}

	s.logger.Info("Notificación creada",
		zap.String("player_id", notification.PlayerID),
		zap.String("notification_id", notification.ID),
		zap.String("type", notification.Type),
	)

	return nil
}

// CreateSystemNotification crea una notificación del sistema
func (s *NotificationService) CreateSystemNotification(message string, data map[string]interface{}) error {
	// Obtener todos los jugadores activos
	players, err := s.playerRepo.GetAllPlayers()
	if err != nil {
		return fmt.Errorf("error obteniendo jugadores: %w", err)
	}

	// Crear notificación para cada jugador
	for _, player := range players {
		notification := &models.Notification{
			PlayerID:  player.ID.String(),
			Type:      "system",
			Title:     "Notificación del Sistema",
			Message:   message,
			Data:      data,
			IsRead:    false,
			IsDeleted: false,
			CreatedAt: time.Now(),
		}

		if err := s.CreateNotification(notification); err != nil {
			s.logger.Error("Error creando notificación del sistema",
				zap.String("player_id", player.ID.String()),
				zap.Error(err),
			)
		}
	}

	s.logger.Info("Notificaciones del sistema enviadas",
		zap.Int("total_players", len(players)),
		zap.String("message", message),
	)

	return nil
}

// GetUnreadCount obtiene el número de notificaciones no leídas
func (s *NotificationService) GetUnreadCount(playerID string) (int, error) {
	count, err := s.notificationRepo.GetUnreadCount(playerID)
	if err != nil {
		return 0, fmt.Errorf("error obteniendo conteo de no leídas: %w", err)
	}

	return count, nil
}

// MarkAllAsRead marca todas las notificaciones como leídas
func (s *NotificationService) MarkAllAsRead(playerID string) error {
	if err := s.notificationRepo.MarkAllAsRead(playerID); err != nil {
		return fmt.Errorf("error marcando todas como leídas: %w", err)
	}

	return nil
}

// CleanOldNotifications limpia notificaciones antiguas
func (s *NotificationService) CleanOldNotifications(daysOld int) error {
	if err := s.notificationRepo.CleanOldNotifications(daysOld); err != nil {
		return fmt.Errorf("error limpiando notificaciones antiguas: %w", err)
	}

	s.logger.Info("Notificaciones antiguas limpiadas",
		zap.Int("days_old", daysOld),
	)

	return nil
}

// Métodos privados

func (s *NotificationService) validateNotification(notification *models.Notification) error {
	if notification == nil {
		return fmt.Errorf("notificación no puede ser nula")
	}

	if notification.PlayerID == "" {
		return fmt.Errorf("ID del jugador es requerido")
	}

	if notification.Type == "" {
		return fmt.Errorf("tipo de notificación es requerido")
	}

	if notification.Message == "" {
		return fmt.Errorf("mensaje es requerido")
	}

	return nil
}

// SetWebSocketManager establece el manager de WebSocket
func (s *NotificationService) SetWebSocketManager(wsManager *websocket.Manager) {
	s.wsManager = wsManager
}
