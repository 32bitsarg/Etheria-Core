package models

import (
	"time"

	"github.com/google/uuid"
)

type Village struct {
	ID          uuid.UUID `json:"id"`
	PlayerID    uuid.UUID `json:"player_id"`
	WorldID     uuid.UUID `json:"world_id"`
	Name        string    `json:"name"`
	XCoordinate int       `json:"x_coordinate"`
	YCoordinate int       `json:"y_coordinate"`
	CreatedAt   time.Time `json:"created_at"`
}

type Resources struct {
	ID          uuid.UUID `json:"id"`
	VillageID   uuid.UUID `json:"village_id"`
	Wood        int       `json:"wood"`
	Stone       int       `json:"stone"`
	Food        int       `json:"food"`
	Gold        int       `json:"gold"`
	LastUpdated time.Time `json:"last_updated"`
}

type Building struct {
	ID                    uuid.UUID  `json:"id"`
	VillageID             uuid.UUID  `json:"village_id"`
	Type                  string     `json:"type"`
	Level                 int        `json:"level"`
	IsUpgrading           bool       `json:"is_upgrading"`
	UpgradeCompletionTime *time.Time `json:"upgrade_completion_time,omitempty"`
}

type VillageWithDetails struct {
	Village   Village              `json:"village"`
	Resources Resources            `json:"resources"`
	Buildings map[string]*Building `json:"buildings"`
}
