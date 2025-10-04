package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand"
	"server-backend/models"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"go.uber.org/zap"
)

type VillageRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewVillageRepository(db *sql.DB, logger *zap.Logger) *VillageRepository {
	return &VillageRepository{
		db:     db,
		logger: logger,
	}
}

func (r *VillageRepository) CreateVillage(playerID, worldID uuid.UUID, name string, x, y int) (*models.VillageWithDetails, error) {
	// Iniciar transacción
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Crear aldea
	villageID := uuid.New()
	_, err = tx.Exec(`
		INSERT INTO villages (id, player_id, world_id, name, x_coordinate, y_coordinate, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, villageID, playerID, worldID, name, x, y, time.Now())
	if err != nil {
		return nil, err
	}

	// Calcular recursos iniciales basados en la capacidad del almacén nivel 1
	initialWood, initialStone, initialFood, initialGold := r.CalculateInitialResources()

	// Inicializar recursos con valores calculados
	resourcesID := uuid.New()
	_, err = tx.Exec(`
		INSERT INTO resources (id, village_id, wood, stone, food, gold, last_updated)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, resourcesID, villageID, initialWood, initialStone, initialFood, initialGold, time.Now())
	if err != nil {
		return nil, err
	}

	// Inicializar edificios básicos
	buildings := []string{"town_hall", "warehouse", "granary", "marketplace", "barracks", "wood_cutter", "stone_quarry", "farm", "gold_mine"}
	for _, buildingType := range buildings {
		buildingID := uuid.New()
		_, err = tx.Exec(`
			INSERT INTO buildings (id, village_id, type, level, is_upgrading, upgrade_completion_time)
			VALUES ($1, $2, $3, 1, false, NULL)
		`, buildingID, villageID, buildingType)
		if err != nil {
			return nil, err
		}
	}

	// Confirmar transacción
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// Obtener la aldea con sus detalles
	return r.GetVillageByID(villageID)
}

func (r *VillageRepository) GetVillageByID(id uuid.UUID) (*models.VillageWithDetails, error) {
	var village models.VillageWithDetails
	var resources models.Resources

	// Obtener aldea y recursos
	err := r.db.QueryRow(`
		SELECT v.id, v.player_id, v.world_id, v.name, v.x_coordinate, v.y_coordinate, v.created_at,
			   r.id, r.village_id, r.wood, r.stone, r.food, r.gold, r.last_updated
		FROM villages v
		LEFT JOIN resources r ON r.village_id = v.id
		WHERE v.id = $1
	`, id).Scan(
		&village.Village.ID,
		&village.Village.PlayerID,
		&village.Village.WorldID,
		&village.Village.Name,
		&village.Village.XCoordinate,
		&village.Village.YCoordinate,
		&village.Village.CreatedAt,
		&resources.ID,
		&resources.VillageID,
		&resources.Wood,
		&resources.Stone,
		&resources.Food,
		&resources.Gold,
		&resources.LastUpdated,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	village.Resources = resources

	// Obtener edificios
	rows, err := r.db.Query(`
		SELECT id, village_id, type, level, is_upgrading, upgrade_completion_time
		FROM buildings
		WHERE village_id = $1
	`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	village.Buildings = make(map[string]*models.Building)
	for rows.Next() {
		var building models.Building
		err := rows.Scan(
			&building.ID,
			&building.VillageID,
			&building.Type,
			&building.Level,
			&building.IsUpgrading,
			&building.UpgradeCompletionTime,
		)
		if err != nil {
			return nil, err
		}
		village.Buildings[building.Type] = &building
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &village, nil
}

func (r *VillageRepository) GetVillagesByPlayerID(playerID uuid.UUID) ([]*models.VillageWithDetails, error) {
	rows, err := r.db.Query(`
		SELECT v.id, v.player_id, v.world_id, v.name, v.x_coordinate, v.y_coordinate, v.created_at,
			   r.id, r.village_id, r.wood, r.stone, r.food, r.gold, r.last_updated
		FROM villages v
		LEFT JOIN resources r ON r.village_id = v.id
		WHERE v.player_id = $1
	`, playerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var villages []*models.VillageWithDetails
	for rows.Next() {
		var village models.VillageWithDetails
		var resources models.Resources
		err := rows.Scan(
			&village.Village.ID,
			&village.Village.PlayerID,
			&village.Village.WorldID,
			&village.Village.Name,
			&village.Village.XCoordinate,
			&village.Village.YCoordinate,
			&village.Village.CreatedAt,
			&resources.ID,
			&resources.VillageID,
			&resources.Wood,
			&resources.Stone,
			&resources.Food,
			&resources.Gold,
			&resources.LastUpdated,
		)
		if err != nil {
			return nil, err
		}
		village.Resources = resources

		// Obtener edificios
		buildingRows, err := r.db.Query(`
			SELECT id, village_id, type, level, is_upgrading, upgrade_completion_time
			FROM buildings
			WHERE village_id = $1
		`, village.Village.ID)
		if err != nil {
			return nil, err
		}
		defer buildingRows.Close()

		village.Buildings = make(map[string]*models.Building)
		for buildingRows.Next() {
			var building models.Building
			err := buildingRows.Scan(
				&building.ID,
				&building.VillageID,
				&building.Type,
				&building.Level,
				&building.IsUpgrading,
				&building.UpgradeCompletionTime,
			)
			if err != nil {
				return nil, err
			}
			if &building == nil {
				r.logger.Warn("Intento de agregar edificio nil al mapa buildings", zap.String("village_id", village.Village.ID.String()), zap.String("type", building.Type))
				continue
			}
			village.Buildings[building.Type] = &building
		}
		if err := buildingRows.Err(); err != nil {
			return nil, err
		}

		villages = append(villages, &village)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return villages, nil
}

func (r *VillageRepository) UpdateResources(villageID uuid.UUID, wood, stone, food, gold int) error {
	_, err := r.db.Exec(`
		UPDATE resources
		SET wood = $1, stone = $2, food = $3, gold = $4, last_updated = $5
		WHERE village_id = $6
	`, wood, stone, food, gold, time.Now(), villageID)
	return err
}

func (r *VillageRepository) UpdateBuilding(villageID uuid.UUID, buildingType string, level int, isUpgrading bool, upgradeCompletionTime *time.Time) error {
	_, err := r.db.Exec(`
		UPDATE buildings
		SET level = $1, is_upgrading = $2, upgrade_completion_time = $3
		WHERE village_id = $4 AND type = $5
	`, level, isUpgrading, upgradeCompletionTime, villageID, buildingType)
	return err
}

// CheckBuildingRequirementsAdvanced verifica los requisitos para construir usando la función avanzada de la BD
func (r *VillageRepository) CheckBuildingRequirementsAdvanced(villageID uuid.UUID, buildingType string, targetLevel int) (*sql.Rows, error) {
	return r.db.Query(`
		SELECT can_build, missing_requirements, cost_wood, cost_stone, cost_food, cost_gold
		FROM check_building_requirements_advanced($1, $2, $3)
	`, villageID, buildingType, targetLevel)
}

// ProcessConstructionQueue procesa la cola de construcción usando la nueva función de la BD
func (r *VillageRepository) ProcessConstructionQueue(villageID uuid.UUID) (*sql.Rows, error) {
	return r.db.Query(`
		SELECT building_type, old_level, new_level, construction_time, resources_spent
		FROM process_construction_queue($1)
	`, villageID)
}

// CalculateResourceProductionAdvanced calcula la producción de recursos usando la función avanzada
func (r *VillageRepository) CalculateResourceProductionAdvanced(villageID uuid.UUID) (*sql.Rows, error) {
	return r.db.Query(`
		SELECT wood_production, stone_production, food_production, gold_production
		FROM calculate_resource_production_advanced($1)
	`, villageID)
}

// CalculateBattleOutcome calcula el resultado de una batalla
func (r *VillageRepository) CalculateBattleOutcome(
	attackerUnits json.RawMessage,
	defenderUnits json.RawMessage,
	attackerHeroes json.RawMessage,
	defenderHeroes json.RawMessage,
	terrain string,
	weather string,
	attackerTechnologies json.RawMessage,
	defenderTechnologies json.RawMessage,
) (*sql.Rows, error) {
	return r.db.Query(`
		SELECT attacker_victory, attacker_losses, defender_losses, battle_duration, experience_gained, resources_plundered
		FROM calculate_battle_outcome($1, $2, $3, $4, $5, $6, $7, $8)
	`, attackerUnits, defenderUnits, attackerHeroes, defenderHeroes, terrain, weather, attackerTechnologies, defenderTechnologies)
}

// CalculateTradeRates calcula las tasas de intercambio
func (r *VillageRepository) CalculateTradeRates(resourceType string, worldID *uuid.UUID) (*sql.Rows, error) {
	if worldID != nil {
		return r.db.Query(`
			SELECT resource_type, current_price, price_trend, supply_level, demand_level, recommended_action
			FROM calculate_trade_rates($1, $2)
		`, resourceType, worldID)
	}
	return r.db.Query(`
		SELECT resource_type, current_price, price_trend, supply_level, demand_level, recommended_action
		FROM calculate_trade_rates($1, NULL)
	`, resourceType)
}

// ProcessAllianceBenefits procesa los beneficios de una alianza
func (r *VillageRepository) ProcessAllianceBenefits(allianceID uuid.UUID) (*sql.Rows, error) {
	return r.db.Query(`
		SELECT member_id, resource_bonus, military_bonus, construction_bonus, total_benefits
		FROM process_alliance_benefits($1)
	`, allianceID)
}

// CalculatePlayerScore calcula la puntuación del jugador
func (r *VillageRepository) CalculatePlayerScore(playerID uuid.UUID) (*sql.Rows, error) {
	return r.db.Query(`
		SELECT total_score, level_score, building_score, military_score, achievement_score, activity_score, rank_position
		FROM calculate_player_score($1)
	`, playerID)
}

// GenerateDailyRewards genera recompensas diarias
func (r *VillageRepository) GenerateDailyRewards(playerID uuid.UUID) (*sql.Rows, error) {
	return r.db.Query(`
		SELECT reward_type, reward_value, reward_description, consecutive_days, total_rewards_today
		FROM generate_daily_rewards($1)
	`, playerID)
}

// CleanupInactiveData limpia datos inactivos
func (r *VillageRepository) CleanupInactiveData(daysOld int) (*sql.Rows, error) {
	return r.db.Query(`
		SELECT table_name, records_deleted, cleanup_type
		FROM cleanup_inactive_data($1)
	`, daysOld)
}

func (r *VillageRepository) GetAllVillages() ([]*models.VillageWithDetails, error) {
	rows, err := r.db.Query(`
		SELECT v.id, v.player_id, v.world_id, v.name, v.x_coordinate, v.y_coordinate, v.created_at,
			   r.id, r.village_id, r.wood, r.stone, r.food, r.gold, r.last_updated
		FROM villages v
		LEFT JOIN resources r ON r.village_id = v.id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var villages []*models.VillageWithDetails
	for rows.Next() {
		var village models.VillageWithDetails
		var resources models.Resources
		err := rows.Scan(
			&village.Village.ID,
			&village.Village.PlayerID,
			&village.Village.WorldID,
			&village.Village.Name,
			&village.Village.XCoordinate,
			&village.Village.YCoordinate,
			&village.Village.CreatedAt,
			&resources.ID,
			&resources.VillageID,
			&resources.Wood,
			&resources.Stone,
			&resources.Food,
			&resources.Gold,
			&resources.LastUpdated,
		)
		if err != nil {
			return nil, err
		}
		village.Resources = resources

		// Obtener edificios
		buildingRows, err := r.db.Query(`
			SELECT id, village_id, type, level, is_upgrading, upgrade_completion_time
			FROM buildings
			WHERE village_id = $1
		`, village.Village.ID)
		if err != nil {
			return nil, err
		}
		defer buildingRows.Close()

		village.Buildings = make(map[string]*models.Building)
		for buildingRows.Next() {
			var building models.Building
			err := buildingRows.Scan(
				&building.ID,
				&building.VillageID,
				&building.Type,
				&building.Level,
				&building.IsUpgrading,
				&building.UpgradeCompletionTime,
			)
			if err != nil {
				return nil, err
			}
			village.Buildings[building.Type] = &building
		}
		if err := buildingRows.Err(); err != nil {
			return nil, err
		}

		villages = append(villages, &village)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return villages, nil
}

func (r *VillageRepository) GetVillageByCoordinates(worldID uuid.UUID, x, y int) (*models.VillageWithDetails, error) {
	var village models.VillageWithDetails
	var resources models.Resources

	// Obtener aldea y recursos por coordenadas
	err := r.db.QueryRow(`
		SELECT v.id, v.player_id, v.world_id, v.name, v.x_coordinate, v.y_coordinate, v.created_at,
			   r.id, r.village_id, r.wood, r.stone, r.food, r.gold, r.last_updated
		FROM villages v
		LEFT JOIN resources r ON r.village_id = v.id
		WHERE v.world_id = $1 AND v.x_coordinate = $2 AND v.y_coordinate = $3
	`, worldID, x, y).Scan(
		&village.Village.ID,
		&village.Village.PlayerID,
		&village.Village.WorldID,
		&village.Village.Name,
		&village.Village.XCoordinate,
		&village.Village.YCoordinate,
		&village.Village.CreatedAt,
		&resources.ID,
		&resources.VillageID,
		&resources.Wood,
		&resources.Stone,
		&resources.Food,
		&resources.Gold,
		&resources.LastUpdated,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	village.Resources = resources

	// Obtener edificios
	rows, err := r.db.Query(`
		SELECT id, village_id, type, level, is_upgrading, upgrade_completion_time
		FROM buildings
		WHERE village_id = $1
	`, village.Village.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	village.Buildings = make(map[string]*models.Building)
	for rows.Next() {
		var building models.Building
		err := rows.Scan(
			&building.ID,
			&building.VillageID,
			&building.Type,
			&building.Level,
			&building.IsUpgrading,
			&building.UpgradeCompletionTime,
		)
		if err != nil {
			return nil, err
		}
		village.Buildings[building.Type] = &building
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &village, nil
}

func (r *VillageRepository) GetVillagesInRange(worldID uuid.UUID, centerX, centerY, rangeDistance int) ([]*models.VillageWithDetails, error) {
	rows, err := r.db.Query(`
		SELECT v.id, v.player_id, v.world_id, v.name, v.x_coordinate, v.y_coordinate, v.created_at,
			   r.id, r.village_id, r.wood, r.stone, r.food, r.gold, r.last_updated
		FROM villages v
		LEFT JOIN resources r ON r.village_id = v.id
		WHERE v.world_id = $1 
		AND ABS(v.x_coordinate - $2) <= $4 
		AND ABS(v.y_coordinate - $3) <= $4
	`, worldID, centerX, centerY, rangeDistance)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var villages []*models.VillageWithDetails
	for rows.Next() {
		var village models.VillageWithDetails
		var resources models.Resources
		err := rows.Scan(
			&village.Village.ID,
			&village.Village.PlayerID,
			&village.Village.WorldID,
			&village.Village.Name,
			&village.Village.XCoordinate,
			&village.Village.YCoordinate,
			&village.Village.CreatedAt,
			&resources.ID,
			&resources.VillageID,
			&resources.Wood,
			&resources.Stone,
			&resources.Food,
			&resources.Gold,
			&resources.LastUpdated,
		)
		if err != nil {
			return nil, err
		}
		village.Resources = resources

		// Obtener edificios
		buildingRows, err := r.db.Query(`
			SELECT id, village_id, type, level, is_upgrading, upgrade_completion_time
			FROM buildings
			WHERE village_id = $1
		`, village.Village.ID)
		if err != nil {
			return nil, err
		}
		defer buildingRows.Close()

		village.Buildings = make(map[string]*models.Building)
		for buildingRows.Next() {
			var building models.Building
			err := buildingRows.Scan(
				&building.ID,
				&building.VillageID,
				&building.Type,
				&building.Level,
				&building.IsUpgrading,
				&building.UpgradeCompletionTime,
			)
			if err != nil {
				return nil, err
			}
			village.Buildings[building.Type] = &building
		}
		if err := buildingRows.Err(); err != nil {
			return nil, err
		}

		villages = append(villages, &village)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return villages, nil
}

// GetBuildingTypes obtiene todos los tipos de edificios disponibles
func (r *VillageRepository) GetBuildingTypes() ([]*models.BuildingType, error) {
	// Como BuildingTypes está definido como un map en models/building.go,
	// vamos a retornar los valores del map
	var buildingTypes []*models.BuildingType
	for _, buildingType := range models.BuildingTypes {
		// Crear una copia del buildingType para evitar problemas con punteros
		bt := buildingType
		buildingTypes = append(buildingTypes, &bt)
	}
	return buildingTypes, nil
}

// GetBuildingType obtiene un tipo de edificio específico
func (r *VillageRepository) GetBuildingType(buildingTypeID string) (*models.BuildingType, error) {
	// Buscar en el map de BuildingTypes
	if buildingType, exists := models.BuildingTypes[buildingTypeID]; exists {
		return &buildingType, nil
	}
	return nil, nil // No encontrado
}

// DeleteVillagesByPlayerAndWorld elimina todas las aldeas de un jugador en un mundo específico
func (r *VillageRepository) DeleteVillagesByPlayerAndWorld(playerID, worldID uuid.UUID) error {
	// Iniciar transacción
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Obtener IDs de aldeas del jugador en el mundo
	rows, err := tx.Query(`
		SELECT id FROM villages 
		WHERE player_id = $1 AND world_id = $2
	`, playerID, worldID)
	if err != nil {
		return err
	}
	defer rows.Close()

	var villageIDs []uuid.UUID
	for rows.Next() {
		var villageID uuid.UUID
		if err := rows.Scan(&villageID); err != nil {
			return err
		}
		villageIDs = append(villageIDs, villageID)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	// Si no hay aldeas, no hacer nada
	if len(villageIDs) == 0 {
		return tx.Commit()
	}

	// Eliminar recursos de las aldeas
	_, err = tx.Exec(`
		DELETE FROM resources 
		WHERE village_id = ANY($1::uuid[])
	`, pq.Array(villageIDs))
	if err != nil {
		return err
	}

	// Eliminar edificios de las aldeas
	_, err = tx.Exec(`
		DELETE FROM buildings 
		WHERE village_id = ANY($1::uuid[])
	`, pq.Array(villageIDs))
	if err != nil {
		return err
	}

	// Eliminar las aldeas
	_, err = tx.Exec(`
		DELETE FROM villages 
		WHERE id = ANY($1::uuid[])
	`, pq.Array(villageIDs))
	if err != nil {
		return err
	}

	// Confirmar transacción
	return tx.Commit()
}

// GenerateRandomCoordinates genera coordenadas aleatorias únicas para una aldea
func (r *VillageRepository) GenerateRandomCoordinates(worldID uuid.UUID) (int, int, error) {
	const maxAttempts = 100
	const maxCoordinate = 1000 // Límite de coordenadas según la base de datos

	for attempt := 0; attempt < maxAttempts; attempt++ {
		// Generar coordenadas aleatorias
		x := rand.Intn(maxCoordinate)
		y := rand.Intn(maxCoordinate)

		// Verificar si las coordenadas están disponibles
		existingVillage, err := r.GetVillageByCoordinates(worldID, x, y)
		if err != nil {
			return 0, 0, err
		}

		// Si no hay aldea en estas coordenadas, son válidas
		if existingVillage == nil {
			return x, y, nil
		}
	}

	// Si no se encontraron coordenadas después de maxAttempts intentos
	return 0, 0, fmt.Errorf("no se pudieron encontrar coordenadas únicas después de %d intentos", maxAttempts)
}

// CalculateInitialResources calcula los recursos iniciales basados en la capacidad del almacén nivel 1
func (r *VillageRepository) CalculateInitialResources() (int, int, int, int) {
	// Capacidad base del almacén nivel 1 según el modelo BuildingTypes
	// warehouse: Wood: 200, Stone: 200, Food: 0, Gold: 0
	// granary: Wood: 0, Stone: 0, Food: 200, Gold: 0
	// Capacidad base: Wood: 100, Stone: 100, Food: 100, Gold: 100

	// Total con almacén y granero nivel 1:
	// Wood: 100 (base) + 200 (warehouse) = 300
	// Stone: 100 (base) + 200 (warehouse) = 300
	// Food: 100 (base) + 200 (granary) = 300
	// Gold: 100 (base) + 0 = 100

	// Iniciar con 80% de la capacidad para dar margen de crecimiento
	initialWood := int(300 * 0.8)  // 240
	initialStone := int(300 * 0.8) // 240
	initialFood := int(300 * 0.8)  // 240
	initialGold := int(100 * 0.8)  // 80

	return initialWood, initialStone, initialFood, initialGold
}

// GetConstructionQueue obtiene la cola de construcción de una aldea (solo mejoras activas)
func (r *VillageRepository) GetConstructionQueue(villageID uuid.UUID) ([]models.ConstructionQueueItem, error) {
	query := `
		SELECT 
			type as building_type,
			level,
			is_upgrading,
			upgrade_completion_time,
			created_at as started_at
		FROM buildings 
		WHERE village_id = $1 
		AND is_upgrading = true
		AND upgrade_completion_time > NOW()
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(query, villageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var queue []models.ConstructionQueueItem
	for rows.Next() {
		var item models.ConstructionQueueItem
		err := rows.Scan(
			&item.BuildingType,
			&item.Level,
			&item.IsUpgrading,
			&item.CompletionTime,
			&item.StartedAt,
		)
		if err != nil {
			return nil, err
		}
		queue = append(queue, item)
	}

	return queue, nil
}

// CleanupCompletedUpgrades limpia las mejoras completadas en la base de datos
func (r *VillageRepository) CleanupCompletedUpgrades(villageID uuid.UUID) error {
	query := `
		UPDATE buildings 
		SET is_upgrading = false, upgrade_completion_time = NULL
		WHERE village_id = $1 
		AND is_upgrading = true 
		AND upgrade_completion_time <= NOW()
	`

	_, err := r.db.Exec(query, villageID)
	return err
}
