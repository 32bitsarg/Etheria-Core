package repository

import (
	"database/sql"
	"fmt"
	"time"

	"server-backend/models"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type BattleRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewBattleRepository(db *sql.DB, logger *zap.Logger) *BattleRepository {
	return &BattleRepository{
		db:     db,
		logger: logger,
	}
}

// GetBattleSystemConfig obtiene la configuración del sistema de batallas
func (r *BattleRepository) GetBattleSystemConfig() (*models.BattleSystemConfig, error) {
	query := `
		SELECT id, is_enabled, advanced_mode, max_battle_duration, max_waves,
		       auto_resolve, advanced_config, created_at, updated_at
		FROM battle_system_config
		ORDER BY created_at DESC
		LIMIT 1
	`

	var config models.BattleSystemConfig
	err := r.db.QueryRow(query).Scan(
		&config.ID, &config.IsEnabled, &config.AdvancedMode, &config.MaxBattleDuration,
		&config.MaxWaves, &config.AutoResolve, &config.AdvancedConfig,
		&config.CreatedAt, &config.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			// Crear configuración por defecto
			return r.createDefaultBattleConfig()
		}
		return nil, fmt.Errorf("error obteniendo configuración: %w", err)
	}

	return &config, nil
}

// createDefaultBattleConfig crea una configuración por defecto
func (r *BattleRepository) createDefaultBattleConfig() (*models.BattleSystemConfig, error) {
	query := `
		INSERT INTO battle_system_config (
			id, is_enabled, advanced_mode, max_battle_duration, max_waves,
			auto_resolve, advanced_config, created_at, updated_at
		) VALUES (
			$1, true, false, 300, 10, true,
			'{"enable_terrain": false, "enable_weather": false, "enable_formations": false}',
			$2, $2
		) RETURNING id, is_enabled, advanced_mode, max_battle_duration, max_waves,
		            auto_resolve, advanced_config, created_at, updated_at
	`

	var config models.BattleSystemConfig
	now := time.Now()
	configID := uuid.New()
	err := r.db.QueryRow(query, configID, now).Scan(
		&config.ID, &config.IsEnabled, &config.AdvancedMode, &config.MaxBattleDuration,
		&config.MaxWaves, &config.AutoResolve, &config.AdvancedConfig,
		&config.CreatedAt, &config.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("error creando configuración por defecto: %w", err)
	}

	return &config, nil
}

// UpdateBattleSystemConfig actualiza la configuración del sistema de batallas
func (r *BattleRepository) UpdateBattleSystemConfig(config *models.BattleSystemConfig) error {
	query := `
		UPDATE battle_system_config 
		SET is_enabled = $1, advanced_mode = $2, max_battle_duration = $3,
		    max_waves = $4, auto_resolve = $5, advanced_config = $6, updated_at = $7
		WHERE id = $8
	`

	_, err := r.db.Exec(query,
		config.IsEnabled, config.AdvancedMode, config.MaxBattleDuration,
		config.MaxWaves, config.AutoResolve, config.AdvancedConfig,
		time.Now(), config.ID,
	)

	if err != nil {
		return fmt.Errorf("error actualizando configuración: %w", err)
	}

	return nil
}

// GetMilitaryUnits obtiene todas las unidades militares
func (r *BattleRepository) GetMilitaryUnits(unitType, category string) ([]models.MilitaryUnit, error) {
	query := `
		SELECT id, name, description, type, category, tier, health, max_health,
		       attack_speed, range, physical_attack, physical_penetration,
		       magic_attack, magic_penetration, physical_defense, physical_resistance,
		       magic_defense, magic_resistance, critical_chance, critical_damage,
		       dodge_chance, block_chance, moral, training_cost, upgrade_cost,
		       requirements, advantages, disadvantages, icon, model, animation,
		       is_active, is_elite, created_at
		FROM military_units
		WHERE is_active = true
	`

	args := []interface{}{}
	argCount := 1

	if unitType != "" {
		query += fmt.Sprintf(" AND type = $%d", argCount)
		args = append(args, unitType)
		argCount++
	}
	if category != "" {
		query += fmt.Sprintf(" AND category = $%d", argCount)
		args = append(args, category)
		argCount++
	}

	query += " ORDER BY tier ASC, type ASC, name ASC"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo unidades militares: %w", err)
	}
	defer rows.Close()

	var units []models.MilitaryUnit
	for rows.Next() {
		var unit models.MilitaryUnit
		err := rows.Scan(
			&unit.ID, &unit.Name, &unit.Description, &unit.Type, &unit.Category, &unit.Tier,
			&unit.Health, &unit.MaxHealth, &unit.AttackSpeed, &unit.Range,
			&unit.PhysicalAttack, &unit.PhysicalPenetration, &unit.MagicAttack, &unit.MagicPenetration,
			&unit.PhysicalDefense, &unit.PhysicalResistance, &unit.MagicDefense, &unit.MagicResistance,
			&unit.CriticalChance, &unit.CriticalDamage, &unit.DodgeChance, &unit.BlockChance,
			&unit.Moral, &unit.TrainingCost, &unit.UpgradeCost, &unit.Requirements,
			&unit.Advantages, &unit.Disadvantages, &unit.Icon, &unit.Model, &unit.Animation,
			&unit.IsActive, &unit.IsElite, &unit.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando unidad militar: %w", err)
		}
		units = append(units, unit)
	}

	return units, nil
}

// GetMilitaryUnit obtiene una unidad militar específica
func (r *BattleRepository) GetMilitaryUnit(unitID uuid.UUID) (*models.MilitaryUnit, error) {
	query := `
		SELECT id, name, description, type, category, tier, health, max_health,
		       attack_speed, range, physical_attack, physical_penetration,
		       magic_attack, magic_penetration, physical_defense, physical_resistance,
		       magic_defense, magic_resistance, critical_chance, critical_damage,
		       dodge_chance, block_chance, moral, training_cost, upgrade_cost,
		       requirements, advantages, disadvantages, icon, model, animation,
		       is_active, is_elite, created_at
		FROM military_units
		WHERE id = $1 AND is_active = true
	`

	var unit models.MilitaryUnit
	err := r.db.QueryRow(query, unitID).Scan(
		&unit.ID, &unit.Name, &unit.Description, &unit.Type, &unit.Category, &unit.Tier,
		&unit.Health, &unit.MaxHealth, &unit.AttackSpeed, &unit.Range,
		&unit.PhysicalAttack, &unit.PhysicalPenetration, &unit.MagicAttack, &unit.MagicPenetration,
		&unit.PhysicalDefense, &unit.PhysicalResistance, &unit.MagicDefense, &unit.MagicResistance,
		&unit.CriticalChance, &unit.CriticalDamage, &unit.DodgeChance, &unit.BlockChance,
		&unit.Moral, &unit.TrainingCost, &unit.UpgradeCost, &unit.Requirements,
		&unit.Advantages, &unit.Disadvantages, &unit.Icon, &unit.Model, &unit.Animation,
		&unit.IsActive, &unit.IsElite, &unit.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("unidad militar no encontrada")
		}
		return nil, fmt.Errorf("error obteniendo unidad militar: %w", err)
	}

	return &unit, nil
}

// GetPlayerUnits obtiene las unidades de un jugador
func (r *BattleRepository) GetPlayerUnits(playerID uuid.UUID) ([]models.PlayerUnit, error) {
	query := `
		SELECT id, player_id, unit_id, quantity, level, experience,
		       current_health, current_attack, current_defense, current_moral,
		       equipment, bonuses, is_training, training_end_time, is_injured,
		       injury_time, battles_won, battles_lost, units_killed, units_lost,
		       created_at, updated_at
		FROM player_units
		WHERE player_id = $1
		ORDER BY level DESC, quantity DESC
	`

	rows, err := r.db.Query(query, playerID)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo unidades del jugador: %w", err)
	}
	defer rows.Close()

	var units []models.PlayerUnit
	for rows.Next() {
		var unit models.PlayerUnit
		err := rows.Scan(
			&unit.ID, &unit.PlayerID, &unit.UnitID, &unit.Quantity, &unit.Level, &unit.Experience,
			&unit.CurrentHealth, &unit.CurrentAttack, &unit.CurrentDefense, &unit.CurrentMoral,
			&unit.Equipment, &unit.Bonuses, &unit.IsTraining, &unit.TrainingEndTime, &unit.IsInjured,
			&unit.InjuryTime, &unit.BattlesWon, &unit.BattlesLost, &unit.UnitsKilled, &unit.UnitsLost,
			&unit.CreatedAt, &unit.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando unidad del jugador: %w", err)
		}
		units = append(units, unit)
	}

	return units, nil
}

// GetPlayerUnit obtiene una unidad específica de un jugador
func (r *BattleRepository) GetPlayerUnit(playerID, unitID uuid.UUID) (*models.PlayerUnit, error) {
	query := `
		SELECT id, player_id, unit_id, quantity, level, experience,
		       current_health, current_attack, current_defense, current_moral,
		       equipment, bonuses, is_training, training_end_time, is_injured,
		       injury_time, battles_won, battles_lost, units_killed, units_lost,
		       created_at, updated_at
		FROM player_units
		WHERE player_id = $1 AND unit_id = $2
	`

	var unit models.PlayerUnit
	err := r.db.QueryRow(query, playerID, unitID).Scan(
		&unit.ID, &unit.PlayerID, &unit.UnitID, &unit.Quantity, &unit.Level, &unit.Experience,
		&unit.CurrentHealth, &unit.CurrentAttack, &unit.CurrentDefense, &unit.CurrentMoral,
		&unit.Equipment, &unit.Bonuses, &unit.IsTraining, &unit.TrainingEndTime, &unit.IsInjured,
		&unit.InjuryTime, &unit.BattlesWon, &unit.BattlesLost, &unit.UnitsKilled, &unit.UnitsLost,
		&unit.CreatedAt, &unit.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("unidad del jugador no encontrada")
		}
		return nil, fmt.Errorf("error obteniendo unidad del jugador: %w", err)
	}

	return &unit, nil
}

// TrainUnits entrena unidades para un jugador
func (r *BattleRepository) TrainUnits(playerID, unitID uuid.UUID, quantity int) error {
	// Obtener la unidad militar para verificar costos y requisitos
	militaryUnit, err := r.GetMilitaryUnit(unitID)
	if err != nil {
		return fmt.Errorf("error obteniendo unidad militar: %w", err)
	}

	// Verificar si el jugador ya tiene esta unidad
	existingUnit, err := r.GetPlayerUnit(playerID, unitID)
	if err != nil && err.Error() != "unidad del jugador no encontrada" {
		return fmt.Errorf("error verificando unidad existente: %w", err)
	}

	now := time.Now()
	trainingTime := r.calculateTrainingTime(militaryUnit.Tier, quantity)
	trainingEndTime := now.Add(trainingTime)

	if existingUnit != nil {
		// Actualizar unidad existente
		query := `
			UPDATE player_units 
			SET quantity = quantity + $1, is_training = true, training_end_time = $2,
			    updated_at = $3
			WHERE player_id = $4 AND unit_id = $5
		`
		_, err = r.db.Exec(query, quantity, trainingEndTime, now, playerID, unitID)
	} else {
		// Crear nueva unidad
		query := `
			INSERT INTO player_units (
				id, player_id, unit_id, quantity, level, experience,
				current_health, current_attack, current_defense, current_moral,
				equipment, bonuses, is_training, training_end_time, is_injured,
				battles_won, battles_lost, units_killed, units_lost,
				created_at, updated_at
			) VALUES (
				$1, $2, $3, $4, 1, 0,
				$5, $6, $7, $8,
				'{}', '{}', true, $9, false,
				0, 0, 0, 0,
				$10, $10
			)
		`
		unitID := uuid.New()
		_, err = r.db.Exec(query,
			unitID, playerID, unitID, quantity,
			militaryUnit.Health, militaryUnit.PhysicalAttack, militaryUnit.PhysicalDefense, militaryUnit.Moral,
			trainingEndTime, now,
		)
	}

	if err != nil {
		return fmt.Errorf("error entrenando unidades: %w", err)
	}

	return nil
}

// CreateBattle crea una nueva batalla
func (r *BattleRepository) CreateBattle(attackerID, defenderID uuid.UUID, battleType, mode string, config map[string]interface{}) (*models.Battle, error) {
	// Obtener configuración del sistema
	systemConfig, err := r.GetBattleSystemConfig()
	if err != nil {
		return nil, fmt.Errorf("error obteniendo configuración del sistema: %w", err)
	}

	if !systemConfig.IsEnabled {
		return nil, fmt.Errorf("sistema de batallas deshabilitado")
	}

	// Crear la batalla
	battle := &models.Battle{
		ID:          uuid.New(),
		AttackerID:  attackerID,
		DefenderID:  defenderID,
		BattleType:  battleType,
		Mode:        mode,
		Status:      "pending",
		MaxWaves:    r.getIntFromConfig(config, "max_waves", systemConfig.MaxWaves),
		MaxDuration: r.getIntFromConfig(config, "max_duration", systemConfig.MaxBattleDuration),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Configuración avanzada si es necesario
	if mode == "advanced" {
		battle.Terrain = r.getStringFromConfig(config, "terrain", "")
		battle.Weather = r.getStringFromConfig(config, "weather", "")
		battle.AttackerFormation = r.getStringFromConfig(config, "attacker_formation", "")
		battle.DefenderFormation = r.getStringFromConfig(config, "defender_formation", "")
		battle.AttackerTactics = r.getStringFromConfig(config, "attacker_tactics", "")
		battle.DefenderTactics = r.getStringFromConfig(config, "defender_tactics", "")
	}

	query := `
		INSERT INTO battles (
			id, attacker_id, defender_id, battle_type, mode, max_waves, max_duration,
			status, current_wave, terrain, weather, attacker_formation, defender_formation,
			attacker_tactics, defender_tactics, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $16
		)
	`

	_, err = r.db.Exec(query,
		battle.ID, battle.AttackerID, battle.DefenderID, battle.BattleType, battle.Mode,
		battle.MaxWaves, battle.MaxDuration, battle.Status, battle.CurrentWave,
		battle.Terrain, battle.Weather, battle.AttackerFormation, battle.DefenderFormation,
		battle.AttackerTactics, battle.DefenderTactics, battle.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("error creando batalla: %w", err)
	}

	return battle, nil
}

// GetBattle obtiene una batalla específica
func (r *BattleRepository) GetBattle(battleID uuid.UUID) (*models.Battle, error) {
	query := `
		SELECT id, attacker_id, defender_id, battle_type, mode, max_waves, max_duration,
		       status, current_wave, start_time, end_time, duration, winner,
		       attacker_losses, defender_losses, terrain, weather, attacker_formation,
		       defender_formation, attacker_tactics, defender_tactics, created_at, updated_at
		FROM battles
		WHERE id = $1
	`

	var battle models.Battle
	err := r.db.QueryRow(query, battleID).Scan(
		&battle.ID, &battle.AttackerID, &battle.DefenderID, &battle.BattleType, &battle.Mode,
		&battle.MaxWaves, &battle.MaxDuration, &battle.Status, &battle.CurrentWave,
		&battle.StartTime, &battle.EndTime, &battle.Duration, &battle.Winner,
		&battle.AttackerLosses, &battle.DefenderLosses, &battle.Terrain, &battle.Weather,
		&battle.AttackerFormation, &battle.DefenderFormation, &battle.AttackerTactics,
		&battle.DefenderTactics, &battle.CreatedAt, &battle.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("batalla no encontrada")
		}
		return nil, fmt.Errorf("error obteniendo batalla: %w", err)
	}

	return &battle, nil
}

// GetBattleWaves obtiene las oleadas de una batalla
func (r *BattleRepository) GetBattleWaves(battleID uuid.UUID) ([]models.BattleWave, error) {
	query := `
		SELECT id, battle_id, wave_number, attacker_units, defender_units,
		       attacker_damage, defender_damage, attacker_losses, defender_losses,
		       combat_log, duration, created_at
		FROM battle_waves
		WHERE battle_id = $1
		ORDER BY wave_number ASC
	`

	rows, err := r.db.Query(query, battleID)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo oleadas: %w", err)
	}
	defer rows.Close()

	var waves []models.BattleWave
	for rows.Next() {
		var wave models.BattleWave
		err := rows.Scan(
			&wave.ID, &wave.BattleID, &wave.WaveNumber, &wave.AttackerUnits, &wave.DefenderUnits,
			&wave.AttackerDamage, &wave.DefenderDamage, &wave.AttackerLosses, &wave.DefenderLosses,
			&wave.CombatLog, &wave.Duration, &wave.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando oleada: %w", err)
		}
		waves = append(waves, wave)
	}

	return waves, nil
}

// GetBattleRankings obtiene los rankings de batalla
func (r *BattleRepository) GetBattleRankings(limit int) ([]models.BattleRanking, error) {
	query := `
		SELECT 
			p.id as player_id,
			p.username as player_name,
			COALESCE(bs.battles_won, 0) as battles_won,
			COALESCE(bs.battles_lost, 0) as battles_lost,
			CASE 
				WHEN COALESCE(bs.battles_won, 0) + COALESCE(bs.battles_lost, 0) = 0 THEN 0
				ELSE ROUND((COALESCE(bs.battles_won, 0)::float / (COALESCE(bs.battles_won, 0) + COALESCE(bs.battles_lost, 0))::float) * 100, 2)
			END as win_rate,
			COALESCE(bs.total_damage_dealt, 0) as total_damage,
			COALESCE(bs.units_killed, 0) as units_killed,
			COALESCE(bs.units_lost, 0) as units_lost,
			ROW_NUMBER() OVER (ORDER BY COALESCE(bs.battles_won, 0) DESC, COALESCE(bs.total_damage_dealt, 0) DESC) as rank
		FROM players p
		LEFT JOIN battle_statistics bs ON p.id = bs.player_id
		ORDER BY COALESCE(bs.battles_won, 0) DESC, COALESCE(bs.total_damage_dealt, 0) DESC
		LIMIT $1
	`

	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo rankings: %w", err)
	}
	defer rows.Close()

	var rankings []models.BattleRanking
	for rows.Next() {
		var ranking models.BattleRanking
		err := rows.Scan(
			&ranking.PlayerID, &ranking.PlayerName, &ranking.BattlesWon, &ranking.BattlesLost,
			&ranking.WinRate, &ranking.TotalDamage, &ranking.UnitsKilled, &ranking.UnitsLost, &ranking.Rank,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando ranking: %w", err)
		}
		rankings = append(rankings, ranking)
	}

	return rankings, nil
}

// calculateTrainingTime calcula el tiempo de entrenamiento basado en el tier y cantidad
func (r *BattleRepository) calculateTrainingTime(tier, quantity int) time.Duration {
	baseTime := time.Duration(tier*30) * time.Second // 30 segundos por tier
	totalTime := baseTime * time.Duration(quantity)
	return totalTime
}

// getStringFromConfig obtiene un string de la configuración
func (r *BattleRepository) getStringFromConfig(config map[string]interface{}, key string, defaultValue string) string {
	if value, exists := config[key]; exists {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return defaultValue
}

// getIntFromConfig obtiene un int de la configuración
func (r *BattleRepository) getIntFromConfig(config map[string]interface{}, key string, defaultValue int) int {
	if value, exists := config[key]; exists {
		if num, ok := value.(int); ok {
			return num
		}
		if num, ok := value.(float64); ok {
			return int(num)
		}
	}
	return defaultValue
}

// UpdateBattle actualiza una batalla
func (r *BattleRepository) UpdateBattle(battle *models.Battle) error {
	query := `
		UPDATE battles 
		SET status = $1, current_wave = $2, start_time = $3, end_time = $4,
		    duration = $5, winner = $6, attacker_losses = $7, defender_losses = $8,
		    terrain = $9, weather = $10, attacker_formation = $11, defender_formation = $12,
		    attacker_tactics = $13, defender_tactics = $14, updated_at = $15
		WHERE id = $16
	`

	_, err := r.db.Exec(query,
		battle.Status, battle.CurrentWave, battle.StartTime, battle.EndTime,
		battle.Duration, battle.Winner, battle.AttackerLosses, battle.DefenderLosses,
		battle.Terrain, battle.Weather, battle.AttackerFormation, battle.DefenderFormation,
		battle.AttackerTactics, battle.DefenderTactics, time.Now(), battle.ID,
	)

	if err != nil {
		return fmt.Errorf("error actualizando batalla: %w", err)
	}

	return nil
}

// GetBattlesByPlayer obtiene las batallas de un jugador
func (r *BattleRepository) GetBattlesByPlayer(playerID uuid.UUID, limit int) ([]models.Battle, error) {
	query := `
		SELECT id, attacker_id, defender_id, battle_type, mode, max_waves, max_duration,
		       status, current_wave, start_time, end_time, duration, winner,
		       attacker_losses, defender_losses, terrain, weather, attacker_formation,
		       defender_formation, attacker_tactics, defender_tactics, created_at, updated_at
		FROM battles
		WHERE attacker_id = $1 OR defender_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := r.db.Query(query, playerID, limit)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo batallas del jugador: %w", err)
	}
	defer rows.Close()

	var battles []models.Battle
	for rows.Next() {
		var battle models.Battle
		err := rows.Scan(
			&battle.ID, &battle.AttackerID, &battle.DefenderID, &battle.BattleType, &battle.Mode,
			&battle.MaxWaves, &battle.MaxDuration, &battle.Status, &battle.CurrentWave,
			&battle.StartTime, &battle.EndTime, &battle.Duration, &battle.Winner,
			&battle.AttackerLosses, &battle.DefenderLosses, &battle.Terrain, &battle.Weather,
			&battle.AttackerFormation, &battle.DefenderFormation, &battle.AttackerTactics,
			&battle.DefenderTactics, &battle.CreatedAt, &battle.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando batalla: %w", err)
		}
		battles = append(battles, battle)
	}

	return battles, nil
}

// GetBattlesByStatus obtiene batallas por estado
func (r *BattleRepository) GetBattlesByStatus(status string, limit int) ([]models.Battle, error) {
	query := `
		SELECT id, attacker_id, defender_id, battle_type, mode, max_waves, max_duration,
		       status, current_wave, start_time, end_time, duration, winner,
		       attacker_losses, defender_losses, terrain, weather, attacker_formation,
		       defender_formation, attacker_tactics, defender_tactics, created_at, updated_at
		FROM battles
		WHERE status = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := r.db.Query(query, status, limit)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo batallas por estado: %w", err)
	}
	defer rows.Close()

	var battles []models.Battle
	for rows.Next() {
		var battle models.Battle
		err := rows.Scan(
			&battle.ID, &battle.AttackerID, &battle.DefenderID, &battle.BattleType, &battle.Mode,
			&battle.MaxWaves, &battle.MaxDuration, &battle.Status, &battle.CurrentWave,
			&battle.StartTime, &battle.EndTime, &battle.Duration, &battle.Winner,
			&battle.AttackerLosses, &battle.DefenderLosses, &battle.Terrain, &battle.Weather,
			&battle.AttackerFormation, &battle.DefenderFormation, &battle.AttackerTactics,
			&battle.DefenderTactics, &battle.CreatedAt, &battle.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando batalla: %w", err)
		}
		battles = append(battles, battle)
	}

	return battles, nil
}

// GetBattlesByType obtiene batallas por tipo
func (r *BattleRepository) GetBattlesByType(battleType string, limit int) ([]models.Battle, error) {
	query := `
		SELECT id, attacker_id, defender_id, battle_type, mode, max_waves, max_duration,
		       status, current_wave, start_time, end_time, duration, winner,
		       attacker_losses, defender_losses, terrain, weather, attacker_formation,
		       defender_formation, attacker_tactics, defender_tactics, created_at, updated_at
		FROM battles
		WHERE battle_type = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := r.db.Query(query, battleType, limit)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo batallas por tipo: %w", err)
	}
	defer rows.Close()

	var battles []models.Battle
	for rows.Next() {
		var battle models.Battle
		err := rows.Scan(
			&battle.ID, &battle.AttackerID, &battle.DefenderID, &battle.BattleType, &battle.Mode,
			&battle.MaxWaves, &battle.MaxDuration, &battle.Status, &battle.CurrentWave,
			&battle.StartTime, &battle.EndTime, &battle.Duration, &battle.Winner,
			&battle.AttackerLosses, &battle.DefenderLosses, &battle.Terrain, &battle.Weather,
			&battle.AttackerFormation, &battle.DefenderFormation, &battle.AttackerTactics,
			&battle.DefenderTactics, &battle.CreatedAt, &battle.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando batalla: %w", err)
		}
		battles = append(battles, battle)
	}

	return battles, nil
}

// GetBattlesByDateRange obtiene batallas en un rango de fechas
func (r *BattleRepository) GetBattlesByDateRange(startDate, endDate time.Time, limit int) ([]models.Battle, error) {
	query := `
		SELECT id, attacker_id, defender_id, battle_type, mode, max_waves, max_duration,
		       status, current_wave, start_time, end_time, duration, winner,
		       attacker_losses, defender_losses, terrain, weather, attacker_formation,
		       defender_formation, attacker_tactics, defender_tactics, created_at, updated_at
		FROM battles
		WHERE created_at >= $1 AND created_at <= $2
		ORDER BY created_at DESC
		LIMIT $3
	`

	rows, err := r.db.Query(query, startDate, endDate, limit)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo batallas por rango de fechas: %w", err)
	}
	defer rows.Close()

	var battles []models.Battle
	for rows.Next() {
		var battle models.Battle
		err := rows.Scan(
			&battle.ID, &battle.AttackerID, &battle.DefenderID, &battle.BattleType, &battle.Mode,
			&battle.MaxWaves, &battle.MaxDuration, &battle.Status, &battle.CurrentWave,
			&battle.StartTime, &battle.EndTime, &battle.Duration, &battle.Winner,
			&battle.AttackerLosses, &battle.DefenderLosses, &battle.Terrain, &battle.Weather,
			&battle.AttackerFormation, &battle.DefenderFormation, &battle.AttackerTactics,
			&battle.DefenderTactics, &battle.CreatedAt, &battle.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando batalla: %w", err)
		}
		battles = append(battles, battle)
	}

	return battles, nil
}

// GetBattleStatistics obtiene las estadísticas de batalla de un jugador
func (r *BattleRepository) GetBattleStatistics(playerID uuid.UUID) (*models.BattleStatistics, error) {
	query := `
		SELECT player_id, total_battles, battles_won, battles_lost, win_rate,
		       total_damage_dealt, total_damage_taken, units_killed, units_lost,
		       kill_death_ratio, average_battle_time, longest_battle_time,
		       shortest_battle_time, last_battle_date, created_at, updated_at
		FROM battle_statistics
		WHERE player_id = $1
	`

	var stats models.BattleStatistics
	err := r.db.QueryRow(query, playerID).Scan(
		&stats.PlayerID, &stats.TotalBattles, &stats.BattlesWon, &stats.BattlesLost, &stats.WinRate,
		&stats.TotalDamageDealt, &stats.TotalDamageTaken, &stats.UnitsKilled, &stats.UnitsLost,
		&stats.KillDeathRatio, &stats.AverageBattleTime, &stats.LongestBattleTime,
		&stats.ShortestBattleTime, &stats.LastBattleDate, &stats.CreatedAt, &stats.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			// Crear estadísticas por defecto
			return r.createDefaultBattleStatistics(playerID)
		}
		return nil, fmt.Errorf("error obteniendo estadísticas: %w", err)
	}

	return &stats, nil
}

// createDefaultBattleStatistics crea estadísticas por defecto para un jugador
func (r *BattleRepository) createDefaultBattleStatistics(playerID uuid.UUID) (*models.BattleStatistics, error) {
	stats := &models.BattleStatistics{
		PlayerID:           playerID,
		TotalBattles:       0,
		BattlesWon:         0,
		BattlesLost:        0,
		WinRate:            0.0,
		TotalDamageDealt:   0,
		TotalDamageTaken:   0,
		UnitsKilled:        0,
		UnitsLost:          0,
		KillDeathRatio:     0.0,
		AverageBattleTime:  0,
		LongestBattleTime:  0,
		ShortestBattleTime: 0,
		LastBattleDate:     time.Now(),
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	query := `
		INSERT INTO battle_statistics (
			player_id, total_battles, battles_won, battles_lost, win_rate,
			total_damage_dealt, total_damage_taken, units_killed, units_lost,
			kill_death_ratio, average_battle_time, longest_battle_time,
			shortest_battle_time, last_battle_date, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $15
		)
	`

	_, err := r.db.Exec(query,
		stats.PlayerID, stats.TotalBattles, stats.BattlesWon, stats.BattlesLost, stats.WinRate,
		stats.TotalDamageDealt, stats.TotalDamageTaken, stats.UnitsKilled, stats.UnitsLost,
		stats.KillDeathRatio, stats.AverageBattleTime, stats.LongestBattleTime,
		stats.ShortestBattleTime, stats.LastBattleDate, stats.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("error creando estadísticas por defecto: %w", err)
	}

	return stats, nil
}

// UpdateBattleStatistics actualiza las estadísticas de batalla de un jugador
func (r *BattleRepository) UpdateBattleStatistics(stats *models.BattleStatistics) error {
	query := `
		UPDATE battle_statistics 
		SET total_battles = $1, battles_won = $2, battles_lost = $3, win_rate = $4,
		    total_damage_dealt = $5, total_damage_taken = $6, units_killed = $7, units_lost = $8,
		    kill_death_ratio = $9, average_battle_time = $10, longest_battle_time = $11,
		    shortest_battle_time = $12, last_battle_date = $13, updated_at = $14
		WHERE player_id = $15
	`

	_, err := r.db.Exec(query,
		stats.TotalBattles, stats.BattlesWon, stats.BattlesLost, stats.WinRate,
		stats.TotalDamageDealt, stats.TotalDamageTaken, stats.UnitsKilled, stats.UnitsLost,
		stats.KillDeathRatio, stats.AverageBattleTime, stats.LongestBattleTime,
		stats.ShortestBattleTime, stats.LastBattleDate, time.Now(), stats.PlayerID,
	)

	if err != nil {
		return fmt.Errorf("error actualizando estadísticas: %w", err)
	}

	return nil
}

// GetTotalBattles obtiene el total de batallas registradas
func (r *BattleRepository) GetTotalBattles() (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM battles`
	err := r.db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("error contando batallas: %w", err)
	}
	return count, nil
}

// GetBattlesToday obtiene el número de batallas de hoy
func (r *BattleRepository) GetBattlesToday() (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM battles WHERE DATE(created_at) = CURRENT_DATE`
	err := r.db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("error contando batallas de hoy: %w", err)
	}
	return count, nil
}

// GetAllBattles obtiene todas las batallas para el dashboard
func (r *BattleRepository) GetAllBattles() ([]models.Battle, error) {
	query := `
		SELECT id, attacker_id, defender_id, battle_type, mode, max_waves, max_duration,
		       status, current_wave, start_time, end_time, duration, winner,
		       attacker_losses, defender_losses, terrain, weather, attacker_formation,
		       defender_formation, attacker_tactics, defender_tactics, created_at, updated_at
		FROM battles
		ORDER BY created_at DESC
		LIMIT 100
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo batallas: %w", err)
	}
	defer rows.Close()

	var battles []models.Battle
	for rows.Next() {
		var battle models.Battle
		err := rows.Scan(
			&battle.ID, &battle.AttackerID, &battle.DefenderID, &battle.BattleType, &battle.Mode,
			&battle.MaxWaves, &battle.MaxDuration, &battle.Status, &battle.CurrentWave,
			&battle.StartTime, &battle.EndTime, &battle.Duration, &battle.Winner,
			&battle.AttackerLosses, &battle.DefenderLosses, &battle.Terrain, &battle.Weather,
			&battle.AttackerFormation, &battle.DefenderFormation, &battle.AttackerTactics,
			&battle.DefenderTactics, &battle.CreatedAt, &battle.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando batalla: %w", err)
		}
		battles = append(battles, battle)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterando batallas: %w", err)
	}

	return battles, nil
}
