package models

import (
	"time"

	"github.com/google/uuid"
)

type World struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	Name           string     `json:"name" db:"name"`
	Description    string     `json:"description" db:"description"`
	MaxPlayers     int        `json:"maxPlayers" db:"max_players"`
	CurrentPlayers int        `json:"currentPlayers" db:"current_players"`
	IsActive       bool       `json:"isActive" db:"is_active"`
	IsOnline       bool       `json:"isOnline" db:"is_online"`
	WorldType      string     `json:"worldType" db:"world_type"`
	Status         string     `json:"status" db:"status"`
	LastStartedAt  *time.Time `json:"lastStartedAt" db:"last_started_at"`
	LastStoppedAt  *time.Time `json:"lastStoppedAt" db:"last_stopped_at"`
	CreatedAt      time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt      time.Time  `json:"updatedAt" db:"updated_at"`
}

type WorldResponse struct {
	ID             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	MaxPlayers     int       `json:"maxPlayers"`
	CurrentPlayers int       `json:"currentPlayers"`
	IsActive       bool      `json:"isActive"`
	IsOnline       bool      `json:"isOnline"`
	WorldType      string    `json:"worldType"`
	Status         string    `json:"status"`
	PlayerCount    int       `json:"playerCount"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

// CreateWorldRequest representa la solicitud para crear un mundo
type CreateWorldRequest struct {
	Name        string `json:"name" validate:"required,min=3,max=100"`
	Description string `json:"description" validate:"max=500"`
	MaxPlayers  int    `json:"maxPlayers" validate:"required,min=1,max=10000"`
	WorldType   string `json:"worldType" validate:"required,oneof=normal pvp peaceful"`
}

// UpdateWorldRequest representa la solicitud para actualizar un mundo
type UpdateWorldRequest struct {
	Name        string `json:"name" validate:"required,min=3,max=100"`
	Description string `json:"description" validate:"max=500"`
	MaxPlayers  int    `json:"maxPlayers" validate:"required,min=1,max=10000"`
	WorldType   string `json:"worldType" validate:"required,oneof=normal pvp peaceful"`
}

// WorldActionResponse representa la respuesta de acciones sobre mundos
type WorldActionResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	WorldID string `json:"worldId,omitempty"`
	Status  string `json:"status,omitempty"`
}

// PlayerWorldEntry representa la entrada de un jugador a un mundo
type PlayerWorldEntry struct {
	ID        uuid.UUID `json:"id" db:"id"`
	PlayerID  uuid.UUID `json:"playerId" db:"player_id"`
	WorldID   uuid.UUID `json:"worldId" db:"world_id"`
	EnteredAt time.Time `json:"enteredAt" db:"entered_at"`
	IsActive  bool      `json:"isActive" db:"is_active"`
	LastSeen  time.Time `json:"lastSeen" db:"last_seen"`
}

// WorldPlayerInfo representa la información de un jugador en un mundo específico
type WorldPlayerInfo struct {
	PlayerID     uuid.UUID  `json:"playerId"`
	Username     string     `json:"username"`
	Level        int        `json:"level"`
	EnteredAt    time.Time  `json:"enteredAt"`
	LastSeen     time.Time  `json:"lastSeen"`
	IsActive     bool       `json:"isActive"`
	VillageCount int        `json:"villageCount"`
	AllianceID   *uuid.UUID `json:"allianceId,omitempty"`
}

// WorldStatus representa el estado actual de un mundo
type WorldStatus struct {
	WorldID        uuid.UUID         `json:"worldId"`
	Name           string            `json:"name"`
	Status         string            `json:"status"`
	IsOnline       bool              `json:"isOnline"`
	CurrentPlayers int               `json:"currentPlayers"`
	MaxPlayers     int               `json:"maxPlayers"`
	LastStartedAt  *time.Time        `json:"lastStartedAt,omitempty"`
	LastStoppedAt  *time.Time        `json:"lastStoppedAt,omitempty"`
	Uptime         string            `json:"uptime,omitempty"`
	PlayerList     []WorldPlayerInfo `json:"playerList,omitempty"`
}
