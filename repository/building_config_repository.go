package repository

import (
	"database/sql"
	"server-backend/models"

	"go.uber.org/zap"
)

type BuildingConfigRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewBuildingConfigRepository(db *sql.DB, logger *zap.Logger) *BuildingConfigRepository {
	return &BuildingConfigRepository{
		db:     db,
		logger: logger,
	}
}

// GetBuildingConfig obtiene la configuración de un edificio específico y nivel
func (r *BuildingConfigRepository) GetBuildingConfig(buildingType string, level int) (*models.BuildingConfig, error) {
	var config models.BuildingConfig
	err := r.db.QueryRow(`
		SELECT id, type, level, wood_cost, stone_cost, food_cost, gold_cost, 
		       build_time_seconds, production_per_hour, storage_capacity, 
		       training_speed_modifier, construction_speed_modifier
		FROM building_configs
		WHERE type = $1 AND level = $2
	`, buildingType, level).Scan(
		&config.ID,
		&config.Type,
		&config.Level,
		&config.WoodCost,
		&config.StoneCost,
		&config.FoodCost,
		&config.GoldCost,
		&config.BuildTimeSeconds,
		&config.ProductionPerHour,
		&config.StorageCapacity,
		&config.TrainingSpeedModifier,
		&config.ConstructionSpeedModifier,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// GetBuildingConfigsByType obtiene todas las configuraciones de un tipo de edificio
func (r *BuildingConfigRepository) GetBuildingConfigsByType(buildingType string) ([]*models.BuildingConfig, error) {
	rows, err := r.db.Query(`
		SELECT id, type, level, wood_cost, stone_cost, food_cost, gold_cost, 
		       build_time_seconds, production_per_hour, storage_capacity, 
		       training_speed_modifier, construction_speed_modifier
		FROM building_configs
		WHERE type = $1
		ORDER BY level
	`, buildingType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var configs []*models.BuildingConfig
	for rows.Next() {
		var config models.BuildingConfig
		err := rows.Scan(
			&config.ID,
			&config.Type,
			&config.Level,
			&config.WoodCost,
			&config.StoneCost,
			&config.FoodCost,
			&config.GoldCost,
			&config.BuildTimeSeconds,
			&config.ProductionPerHour,
			&config.StorageCapacity,
			&config.TrainingSpeedModifier,
			&config.ConstructionSpeedModifier,
		)
		if err != nil {
			return nil, err
		}
		configs = append(configs, &config)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return configs, nil
}

// GetMaxLevel obtiene el nivel máximo disponible para un tipo de edificio
func (r *BuildingConfigRepository) GetMaxLevel(buildingType string) (int, error) {
	var maxLevel int
	err := r.db.QueryRow(`
		SELECT COALESCE(MAX(level), 0)
		FROM building_configs
		WHERE type = $1
	`, buildingType).Scan(&maxLevel)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return maxLevel, nil
}

// GetAllBuildingTypes obtiene todos los tipos de edificios disponibles
func (r *BuildingConfigRepository) GetAllBuildingTypes() ([]string, error) {
	rows, err := r.db.Query(`
		SELECT DISTINCT type
		FROM building_configs
		ORDER BY type
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var types []string
	for rows.Next() {
		var buildingType string
		err := rows.Scan(&buildingType)
		if err != nil {
			return nil, err
		}
		types = append(types, buildingType)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return types, nil
}
