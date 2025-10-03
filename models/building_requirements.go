package models

import (
	"time"
)

// CancelUpgradeResult representa el resultado de cancelar una mejora de edificio
type CancelUpgradeResult struct {
	BuildingType     string              `json:"building_type"`
	RefundAmount     ResourceCostsLegacy `json:"refund_amount"`
	RefundPercentage float64             `json:"refund_percentage"`
	OriginalCost     ResourceCostsLegacy `json:"original_cost"`
	CancelledAt      time.Time           `json:"cancelled_at"`
	TimeRemaining    time.Duration       `json:"time_remaining"`
	RefundReason     string              `json:"refund_reason"`
}

// BuildingUpgradeResultLegacy representa el resultado de una mejora de edificio (formato legacy)
type BuildingUpgradeResultLegacy struct {
	BuildingType   string              `json:"building_type"`
	NewLevel       int                 `json:"new_level"`
	UpgradeTime    time.Duration       `json:"upgrade_time"`
	CompletionTime time.Time           `json:"completion_time"`
	Costs          ResourceCostsLegacy `json:"costs"`
	ResourcesSpent ResourceCostsLegacy `json:"resources_spent"`
}

// ResourceCostsLegacy representa los costos de recursos (formato legacy)
type ResourceCostsLegacy struct {
	Wood  int `json:"wood"`
	Stone int `json:"stone"`
	Food  int `json:"food"`
	Gold  int `json:"gold"`
}

// BuildingUpgradeStatus representa el estado de una mejora de edificio
type BuildingUpgradeStatus struct {
	BuildingType     string        `json:"building_type"`
	CurrentLevel     int           `json:"current_level"`
	TargetLevel      int           `json:"target_level"`
	IsUpgrading      bool          `json:"is_upgrading"`
	StartTime        time.Time     `json:"start_time,omitempty"`
	CompletionTime   time.Time     `json:"completion_time,omitempty"`
	TimeRemaining    time.Duration `json:"time_remaining,omitempty"`
	ProgressPercent  float64       `json:"progress_percent,omitempty"`
	CanCancel        bool          `json:"can_cancel"`
	RefundPercentage float64       `json:"refund_percentage"`
}
