package models

import (
	"time"

	"github.com/google/uuid"
)

// ResourceUpdate representa una actualización de recursos para notificaciones WebSocket
type ResourceUpdate struct {
	VillageID      uuid.UUID `json:"village_id"`
	Wood           int       `json:"wood"`
	Stone          int       `json:"stone"`
	Food           int       `json:"food"`
	Gold           int       `json:"gold"`
	WoodGenerated  int       `json:"wood_generated"`
	StoneGenerated int       `json:"stone_generated"`
	FoodGenerated  int       `json:"food_generated"`
	GoldGenerated  int       `json:"gold_generated"`
	Capacity       Resources `json:"capacity"`
	LastUpdate     time.Time `json:"last_update"`
	ElapsedHours   float64   `json:"elapsed_hours"`
}

// ResourceNotification representa una notificación de recursos para el cliente
type ResourceNotification struct {
	Type      string        `json:"type"`
	VillageID uuid.UUID     `json:"village_id"`
	Resources ResourceUpdate `json:"resources"`
	Timestamp time.Time     `json:"timestamp"`
}

// ResourceMetrics representa métricas del sistema de recursos
type ResourceMetrics struct {
	WebSocketNotifications int       `json:"websocket_notifications"`
	PollingRequests        int       `json:"polling_requests"`
	UpdateErrors          int       `json:"update_errors"`
	LastUpdate            time.Time `json:"last_update"`
	TotalUpdates          int       `json:"total_updates"`
	SuccessRate           float64   `json:"success_rate"`
}

// ResourceProduction representa la producción de recursos de una aldea
type ResourceProduction struct {
	VillageID  uuid.UUID `json:"village_id"`
	Wood       int       `json:"wood"`
	Stone      int       `json:"stone"`
	Food       int       `json:"food"`
	Gold       int       `json:"gold"`
	LastUpdate time.Time `json:"last_update"`
}

// ResourceStorage representa la capacidad de almacenamiento de una aldea
type ResourceStorage struct {
	VillageID    uuid.UUID `json:"village_id"`
	WoodStorage  int       `json:"wood_storage"`
	StoneStorage int       `json:"stone_storage"`
	FoodStorage  int       `json:"food_storage"`
	GoldStorage  int       `json:"gold_storage"`
}

// ResourceData representa los datos de recursos para Redis
type ResourceData struct {
	Wood  int `json:"wood"`
	Stone int `json:"stone"`
	Food  int `json:"food"`
	Gold  int `json:"gold"`
	Gems  int `json:"gems"`
}