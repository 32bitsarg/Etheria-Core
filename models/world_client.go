package models

import (
	"time"

	"github.com/google/uuid"
)

// WorldClientResponse representa la respuesta para el cliente
type WorldClientResponse struct {
	ID             uuid.UUID     `json:"id"`
	Name           string        `json:"name"`
	Description    string        `json:"description"`
	MaxPlayers     int           `json:"maxPlayers"`
	CurrentPlayers int           `json:"currentPlayers"`
	IsOnline       bool          `json:"isOnline"`
	WorldType      string        `json:"worldType"`
	Status         string        `json:"status"`
	PlayerCount    int           `json:"playerCount"`
	IsFull         bool          `json:"isFull"`
	CanJoin        bool          `json:"canJoin"`
	LastStartedAt  *time.Time    `json:"lastStartedAt,omitempty"`
	Uptime         string        `json:"uptime,omitempty"`
	Features       WorldFeatures `json:"features"`
}

// WorldFeatures representa las características del mundo
type WorldFeatures struct {
	PvPEnabled       bool `json:"pvpEnabled"`
	AlliancesEnabled bool `json:"alliancesEnabled"`
	TradingEnabled   bool `json:"tradingEnabled"`
	EventsEnabled    bool `json:"eventsEnabled"`
}

// WorldJoinRequest representa la solicitud para unirse a un mundo
type WorldJoinRequest struct {
	VillageName      string `json:"villageName" validate:"required,min=3,max=50"`
	StartingLocation string `json:"startingLocation" validate:"oneof=random center edge"`
}

// WorldJoinResponse representa la respuesta al unirse a un mundo
type WorldJoinResponse struct {
	Success           bool        `json:"success"`
	Message           string      `json:"message"`
	WorldID           string      `json:"worldId"`
	VillageID         string      `json:"villageId"`
	StartingResources ResourceSet `json:"startingResources"`
	RedirectUrl       string      `json:"redirectUrl"`
}

// ResourceSet representa un conjunto de recursos
type ResourceSet struct {
	Gold  int `json:"gold"`
	Wood  int `json:"wood"`
	Stone int `json:"stone"`
	Food  int `json:"food"`
}

// WorldStats representa las estadísticas de un mundo
type WorldStats struct {
	WorldID        string         `json:"worldId"`
	Name           string         `json:"name"`
	PlayerCount    int            `json:"playerCount"`
	MaxPlayers     int            `json:"maxPlayers"`
	AllianceCount  int            `json:"allianceCount"`
	VillageCount   int            `json:"villageCount"`
	BattleCount    int            `json:"battleCount"`
	TradeCount     int            `json:"tradeCount"`
	TopPlayers     []TopPlayer    `json:"topPlayers"`
	RecentActivity []ActivityItem `json:"recentActivity"`
}

// TopPlayer representa un jugador top
type TopPlayer struct {
	Username     string `json:"username"`
	Level        int    `json:"level"`
	VillageCount int    `json:"villageCount"`
}

// ActivityItem representa una actividad reciente
type ActivityItem struct {
	Type        string    `json:"type"`
	Description string    `json:"description"`
	Timestamp   time.Time `json:"timestamp"`
}

// WorldStatusResponse representa el estado de un mundo
type WorldStatusResponse struct {
	WorldID           string  `json:"worldId"`
	IsOnline          bool    `json:"isOnline"`
	IsFull            bool    `json:"isFull"`
	CanJoin           bool    `json:"canJoin"`
	MaintenanceMode   bool    `json:"maintenanceMode"`
	EstimatedWaitTime int     `json:"estimatedWaitTime"`
	ServerLoad        float64 `json:"serverLoad"`
}

// PlayerWorldInfo representa la información del mundo actual del jugador
type PlayerWorldInfo struct {
	WorldID      string    `json:"worldId"`
	WorldName    string    `json:"worldName"`
	JoinedAt     time.Time `json:"joinedAt"`
	LastSeen     time.Time `json:"lastSeen"`
	VillageCount int       `json:"villageCount"`
	IsActive     bool      `json:"isActive"`
	CanLeave     bool      `json:"canLeave"`
}

// WorldPlayerListItem representa un jugador en la lista del mundo
type WorldPlayerListItem struct {
	PlayerID     string    `json:"playerId"`
	Username     string    `json:"username"`
	Level        int       `json:"level"`
	VillageCount int       `json:"villageCount"`
	AllianceName *string   `json:"allianceName,omitempty"`
	LastSeen     time.Time `json:"lastSeen"`
	IsOnline     bool      `json:"isOnline"`
}
