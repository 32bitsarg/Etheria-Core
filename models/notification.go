package models

import (
	"time"
)

// Notification representa una notificación
type Notification struct {
	ID        string                 `json:"id"`
	PlayerID  string                 `json:"player_id"`
	Type      string                 `json:"type"`
	Title     string                 `json:"title"`
	Message   string                 `json:"message"`
	Data      map[string]interface{} `json:"data,omitempty"`
	IsRead    bool                   `json:"is_read"`
	IsDeleted bool                   `json:"is_deleted"`
	CreatedAt time.Time              `json:"created_at"`
	ReadAt    *time.Time             `json:"read_at,omitempty"`
}

// NotificationTemplate representa una plantilla de notificación
type NotificationTemplate struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Title    string `json:"title"`
	Message  string `json:"message"`
	IsActive bool   `json:"is_active"`
}

// NotificationSettings representa la configuración de notificaciones de un jugador
type NotificationSettings struct {
	PlayerID            string          `json:"player_id"`
	EmailNotifications  bool            `json:"email_notifications"`
	PushNotifications   bool            `json:"push_notifications"`
	InGameNotifications bool            `json:"in_game_notifications"`
	Preferences         map[string]bool `json:"preferences"`
}

// NotificationBatch representa un lote de notificaciones
type NotificationBatch struct {
	ID          string                 `json:"id"`
	TemplateID  string                 `json:"template_id"`
	PlayerIDs   []string               `json:"player_ids"`
	Data        map[string]interface{} `json:"data,omitempty"`
	ScheduledAt *time.Time             `json:"scheduled_at,omitempty"`
	SentAt      *time.Time             `json:"sent_at,omitempty"`
	Status      string                 `json:"status"`
	CreatedAt   time.Time              `json:"created_at"`
}

// NotificationStats representa estadísticas de notificaciones
type NotificationStats struct {
	PlayerID         string     `json:"player_id"`
	TotalReceived    int        `json:"total_received"`
	TotalRead        int        `json:"total_read"`
	TotalUnread      int        `json:"total_unread"`
	ReadRate         float64    `json:"read_rate"`
	LastNotification *time.Time `json:"last_notification,omitempty"`
}
