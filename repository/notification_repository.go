package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"server-backend/models"
)

type NotificationRepository struct {
	db *sql.DB
}

func NewNotificationRepository(db *sql.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

// CreateNotification crea una nueva notificación
func (r *NotificationRepository) CreateNotification(notification *models.Notification) error {
	query := `
		INSERT INTO notifications (
			player_id, type, title, message, data, is_read, is_deleted, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	notification.CreatedAt = now

	// Convertir data a JSON string
	dataJSON := "{}"
	if notification.Data != nil {
		if jsonBytes, err := json.Marshal(notification.Data); err == nil {
			dataJSON = string(jsonBytes)
		}
	}

	result, err := r.db.Exec(query,
		notification.PlayerID, notification.Type, notification.Title,
		notification.Message, dataJSON, notification.IsRead,
		notification.IsDeleted, notification.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("error creating notification: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("error getting last insert id: %v", err)
	}

	notification.ID = fmt.Sprintf("%d", id)
	return nil
}

// GetPlayerNotifications obtiene las notificaciones de un jugador
func (r *NotificationRepository) GetPlayerNotifications(playerID string, limit int) ([]*models.Notification, error) {
	query := `
		SELECT id, player_id, type, title, message, data, is_read, is_deleted, created_at, read_at
		FROM notifications 
		WHERE player_id = ? AND is_deleted = false 
		ORDER BY created_at DESC
		LIMIT ?
	`

	rows, err := r.db.Query(query, playerID, limit)
	if err != nil {
		return nil, fmt.Errorf("error getting player notifications: %v", err)
	}
	defer rows.Close()

	var notifications []*models.Notification
	for rows.Next() {
		var notification models.Notification
		var dataJSON string
		err := rows.Scan(
			&notification.ID, &notification.PlayerID, &notification.Type,
			&notification.Title, &notification.Message, &dataJSON,
			&notification.IsRead, &notification.IsDeleted, &notification.CreatedAt,
			&notification.ReadAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning notification: %v", err)
		}

		// Convertir dataJSON a map[string]interface{}
		notification.Data = make(map[string]interface{})
		if dataJSON != "" && dataJSON != "{}" {
			if err := json.Unmarshal([]byte(dataJSON), &notification.Data); err != nil {
				// Si hay error en el parsing, mantener el map vacío
				notification.Data = make(map[string]interface{})
			}
		}

		notifications = append(notifications, &notification)
	}

	return notifications, nil
}

// MarkNotificationAsRead marca una notificación como leída
func (r *NotificationRepository) MarkNotificationAsRead(playerID, notificationID string) error {
	query := `
		UPDATE notifications 
		SET is_read = true, read_at = ? 
		WHERE id = ? AND player_id = ?
	`

	now := time.Now()
	_, err := r.db.Exec(query, now, notificationID, playerID)
	if err != nil {
		return fmt.Errorf("error marking notification as read: %v", err)
	}

	return nil
}

// DeleteNotification elimina una notificación
func (r *NotificationRepository) DeleteNotification(playerID, notificationID string) error {
	query := `
		UPDATE notifications 
		SET is_deleted = true 
		WHERE id = ? AND player_id = ?
	`

	_, err := r.db.Exec(query, notificationID, playerID)
	if err != nil {
		return fmt.Errorf("error deleting notification: %v", err)
	}

	return nil
}

// GetUnreadCount obtiene el número de notificaciones no leídas
func (r *NotificationRepository) GetUnreadCount(playerID string) (int, error) {
	query := `
		SELECT COUNT(*) 
		FROM notifications 
		WHERE player_id = ? AND is_read = false AND is_deleted = false
	`

	var count int
	err := r.db.QueryRow(query, playerID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("error getting unread count: %v", err)
	}

	return count, nil
}

// MarkAllAsRead marca todas las notificaciones como leídas
func (r *NotificationRepository) MarkAllAsRead(playerID string) error {
	query := `
		UPDATE notifications 
		SET is_read = true, read_at = ? 
		WHERE player_id = ? AND is_read = false AND is_deleted = false
	`

	now := time.Now()
	_, err := r.db.Exec(query, now, playerID)
	if err != nil {
		return fmt.Errorf("error marking all as read: %v", err)
	}

	return nil
}

// CleanOldNotifications limpia notificaciones antiguas
func (r *NotificationRepository) CleanOldNotifications(daysOld int) error {
	query := `
		UPDATE notifications 
		SET is_deleted = true 
		WHERE created_at < ? AND is_deleted = false
	`

	cutoffDate := time.Now().AddDate(0, 0, -daysOld)
	_, err := r.db.Exec(query, cutoffDate)
	if err != nil {
		return fmt.Errorf("error cleaning old notifications: %v", err)
	}

	return nil
}

// GetNotificationByID obtiene una notificación por ID
func (r *NotificationRepository) GetNotificationByID(notificationID string) (*models.Notification, error) {
	query := `
		SELECT id, player_id, type, title, message, data, is_read, is_deleted, created_at, read_at
		FROM notifications 
		WHERE id = ? AND is_deleted = false
	`

	var notification models.Notification
	var dataJSON string
	err := r.db.QueryRow(query, notificationID).Scan(
		&notification.ID, &notification.PlayerID, &notification.Type,
		&notification.Title, &notification.Message, &dataJSON,
		&notification.IsRead, &notification.IsDeleted, &notification.CreatedAt,
		&notification.ReadAt,
	)
	if err != nil {
		return nil, fmt.Errorf("error getting notification: %v", err)
	}

	// Convertir dataJSON a map[string]interface{}
	notification.Data = make(map[string]interface{})
	if dataJSON != "" && dataJSON != "{}" {
		if err := json.Unmarshal([]byte(dataJSON), &notification.Data); err != nil {
			// Si hay error en el parsing, mantener el map vacío
			notification.Data = make(map[string]interface{})
		}
	}

	return &notification, nil
}

// GetNotificationStats obtiene estadísticas de notificaciones de un jugador
func (r *NotificationRepository) GetNotificationStats(playerID string) (*models.NotificationStats, error) {
	query := `
		SELECT 
			COUNT(*) as total_received,
			SUM(CASE WHEN is_read = true THEN 1 ELSE 0 END) as total_read,
			SUM(CASE WHEN is_read = false THEN 1 ELSE 0 END) as total_unread,
			MAX(created_at) as last_notification
		FROM notifications 
		WHERE player_id = ? AND is_deleted = false
	`

	var stats models.NotificationStats
	stats.PlayerID = playerID

	var lastNotification *time.Time
	err := r.db.QueryRow(query, playerID).Scan(
		&stats.TotalReceived, &stats.TotalRead, &stats.TotalUnread, &lastNotification,
	)
	if err != nil {
		return nil, fmt.Errorf("error getting notification stats: %v", err)
	}

	stats.LastNotification = lastNotification

	// Calcular tasa de lectura
	if stats.TotalReceived > 0 {
		stats.ReadRate = float64(stats.TotalRead) / float64(stats.TotalReceived)
	}

	return &stats, nil
}
