package models

import (
	"time"

	"github.com/google/uuid"
)

type ChatMessage struct {
	ID        uuid.UUID `json:"id"`
	ChannelID uuid.UUID `json:"channel_id"`
	PlayerID  uuid.UUID `json:"player_id"`
	Username  string    `json:"username"`
	Message   string    `json:"message"`
	Type      string    `json:"type"` // "text", "system", "private"
	CreatedAt time.Time `json:"created_at"`
}

type ChatChannel struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Type        string     `json:"type"` // "global", "world", "alliance", "private"
	WorldID     *uuid.UUID `json:"world_id,omitempty"`
	IsActive    bool       `json:"is_active"`
	CreatedAt   time.Time  `json:"created_at"`
}

type ChatChannelMember struct {
	ID        uuid.UUID `json:"id"`
	ChannelID uuid.UUID `json:"channel_id"`
	PlayerID  uuid.UUID `json:"player_id"`
	Username  string    `json:"username"`
	IsAdmin   bool      `json:"is_admin"`
	JoinedAt  time.Time `json:"joined_at"`
}

type WebSocketMessage struct {
	Type     string      `json:"type"`
	Channel  string      `json:"channel,omitempty"`
	Data     interface{} `json:"data"`
	PlayerID string      `json:"player_id,omitempty"`
	Username string      `json:"username,omitempty"`
}

// Tipos de mensajes WebSocket
const (
	WSMessageTypeChat     = "chat"
	WSMessageTypeSystem   = "system"
	WSMessageTypePrivate  = "private"
	WSMessageTypeJoin     = "join"
	WSMessageTypeLeave    = "leave"
	WSMessageTypeResource = "resource"
	WSMessageTypeBuilding = "building"
	WSMessageTypeUnit     = "unit"
	WSMessageTypeAttack   = "attack"
	WSMessageTypeDefense  = "defense"
)

// Canales de chat predefinidos
var DefaultChannels = map[string]ChatChannel{
	"global": {
		Name:        "Global",
		Description: "Chat global para todos los jugadores",
		Type:        "global",
		IsActive:    true,
	},
	"help": {
		Name:        "Ayuda",
		Description: "Canal de ayuda y soporte",
		Type:        "global",
		IsActive:    true,
	},
	"trade": {
		Name:        "Comercio",
		Description: "Canal para intercambios y comercio",
		Type:        "global",
		IsActive:    true,
	},
}
