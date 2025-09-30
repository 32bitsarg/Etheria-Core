package models

// BuildingConfig representa la configuración de un nivel específico de un edificio
type BuildingConfig struct {
	ID                        int     `json:"id" db:"id"`
	Type                      string  `json:"type" db:"type"`
	Level                     int     `json:"level" db:"level"`
	WoodCost                  int     `json:"wood_cost" db:"wood_cost"`
	StoneCost                 int     `json:"stone_cost" db:"stone_cost"`
	FoodCost                  int     `json:"food_cost" db:"food_cost"`
	GoldCost                  int     `json:"gold_cost" db:"gold_cost"`
	BuildTimeSeconds          int     `json:"build_time_seconds" db:"build_time_seconds"`
	ProductionPerHour         int     `json:"production_per_hour" db:"production_per_hour"`
	StorageCapacity           int     `json:"storage_capacity" db:"storage_capacity"`
	TrainingSpeedModifier     float64 `json:"training_speed_modifier" db:"training_speed_modifier"`
	ConstructionSpeedModifier float64 `json:"construction_speed_modifier" db:"construction_speed_modifier"`
}

// BuildingConfigResponse para respuestas de API
type BuildingConfigResponse struct {
	Type                      string  `json:"type"`
	Level                     int     `json:"level"`
	WoodCost                  int     `json:"wood_cost"`
	StoneCost                 int     `json:"stone_cost"`
	FoodCost                  int     `json:"food_cost"`
	GoldCost                  int     `json:"gold_cost"`
	BuildTimeSeconds          int     `json:"build_time_seconds"`
	ProductionPerHour         int     `json:"production_per_hour"`
	StorageCapacity           int     `json:"storage_capacity"`
	TrainingSpeedModifier     float64 `json:"training_speed_modifier"`
	ConstructionSpeedModifier float64 `json:"construction_speed_modifier"`
}
