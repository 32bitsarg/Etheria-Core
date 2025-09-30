package models

import (
	"time"

	"github.com/google/uuid"
)

type Player struct {
	ID           uuid.UUID  `json:"id"`
	Username     string     `json:"username"`
	Password     string     `json:"-"` // No se incluye en JSON
	Email        string     `json:"email"`
	Role         string     `json:"role"`
	Level        int        `json:"level"`
	Experience   int        `json:"experience"`
	IsActive     bool       `json:"isActive"`
	IsOnline     bool       `json:"isOnline"`
	IsBanned     bool       `json:"isBanned"`
	BanReason    *string    `json:"banReason,omitempty"`
	BanExpiresAt *time.Time `json:"banExpiresAt,omitempty"`
	Gold         int        `json:"gold"`
	Gems         int        `json:"gems"`
	AllianceID   *uuid.UUID `json:"allianceId,omitempty"`
	WorldID      *uuid.UUID `json:"worldId,omitempty"`
	RaceID       *uuid.UUID `json:"raceId,omitempty"`
	LastLogin    time.Time  `json:"lastLogin"`
	LastActive   *time.Time `json:"lastActive,omitempty"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`

	// Campos relacionados (opcionales)
	Villages     []VillageWithDetails `json:"villages,omitempty"`
	Achievements []SimpleAchievement  `json:"achievements,omitempty"`
	Titles       []Title              `json:"titles,omitempty"`
}

// SimpleAchievement representa un logro simple para el dashboard
type SimpleAchievement struct {
	ID               uuid.UUID `json:"id"`
	Name             string    `json:"name"`
	Description      string    `json:"description"`
	Category         string    `json:"category"`
	RequirementType  string    `json:"requirementType"`
	RequirementValue int       `json:"requirementValue"`
	RewardType       string    `json:"rewardType"`
	RewardValue      int       `json:"rewardValue"`
	IconURL          string    `json:"iconUrl"`
	CreatedAt        time.Time `json:"createdAt"`
}

func NewPlayer(username, password, email string) *Player {
	now := time.Now()
	return &Player{
		ID:         uuid.New(),
		Username:   username,
		Password:   password,
		Email:      email,
		Role:       "user",
		Level:      1,
		Experience: 0,
		IsActive:   true,
		IsOnline:   false,
		IsBanned:   false,
		Gold:       0,
		Gems:       0,
		LastLogin:  now,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

type PlayerRepository interface {
	Create(player *Player) error
	GetByID(id uuid.UUID) (*Player, error)
	GetByEmail(email string) (*Player, error)
	Update(player *Player) error
	Delete(id uuid.UUID) error
}
