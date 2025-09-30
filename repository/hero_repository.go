package repository

import (
	"database/sql"
	"fmt"
	"time"

	"server-backend/models"

	"go.uber.org/zap"
)

type HeroRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewHeroRepository(db *sql.DB, logger *zap.Logger) *HeroRepository {
	return &HeroRepository{
		db:     db,
		logger: logger,
	}
}

// GetHeroes obtiene todos los héroes disponibles
func (r *HeroRepository) GetHeroes(race, class, rarity string) ([]models.Hero, error) {
	query := `
		SELECT id, name, title, description, race, class, rarity, level, max_level, 
		       experience, experience_to_next, health, attack, defense, speed, 
		       intelligence, charisma, active_skills, passive_skills, ultimate_skill,
		       equipment, artifacts, recruit_cost, upgrade_cost, requirements,
		       icon, portrait, model, color, is_active, is_special, is_limited,
		       release_date, expiry_date, created_at, updated_at
		FROM heroes
		WHERE is_active = true
	`

	args := []interface{}{}
	argCount := 1

	if race != "" {
		query += fmt.Sprintf(" AND race = $%d", argCount)
		args = append(args, race)
		argCount++
	}
	if class != "" {
		query += fmt.Sprintf(" AND class = $%d", argCount)
		args = append(args, class)
		argCount++
	}
	if rarity != "" {
		query += fmt.Sprintf(" AND rarity = $%d", argCount)
		args = append(args, rarity)
		argCount++
	}

	query += " ORDER BY rarity DESC, level ASC, name ASC"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo héroes: %w", err)
	}
	defer rows.Close()

	var heroes []models.Hero
	for rows.Next() {
		var hero models.Hero
		err := rows.Scan(
			&hero.ID, &hero.Name, &hero.Title, &hero.Description, &hero.Race, &hero.Class,
			&hero.Rarity, &hero.Level, &hero.MaxLevel, &hero.Experience, &hero.ExperienceToNext,
			&hero.Health, &hero.Attack, &hero.Defense, &hero.Speed, &hero.Intelligence, &hero.Charisma,
			&hero.ActiveSkills, &hero.PassiveSkills, &hero.UltimateSkill, &hero.Equipment, &hero.Artifacts,
			&hero.RecruitCost, &hero.UpgradeCost, &hero.Requirements, &hero.Icon, &hero.Portrait,
			&hero.Model, &hero.Color, &hero.IsActive, &hero.IsSpecial, &hero.IsLimited,
			&hero.ReleaseDate, &hero.ExpiryDate, &hero.CreatedAt, &hero.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando héroe: %w", err)
		}
		heroes = append(heroes, hero)
	}

	return heroes, nil
}

// GetHero obtiene un héroe específico
func (r *HeroRepository) GetHero(heroID int) (*models.Hero, error) {
	query := `
		SELECT id, name, title, description, race, class, rarity, level, max_level, 
		       experience, experience_to_next, health, attack, defense, speed, 
		       intelligence, charisma, active_skills, passive_skills, ultimate_skill,
		       equipment, artifacts, recruit_cost, upgrade_cost, requirements,
		       icon, portrait, model, color, is_active, is_special, is_limited,
		       release_date, expiry_date, created_at, updated_at
		FROM heroes
		WHERE id = $1 AND is_active = true
	`

	var hero models.Hero
	err := r.db.QueryRow(query, heroID).Scan(
		&hero.ID, &hero.Name, &hero.Title, &hero.Description, &hero.Race, &hero.Class,
		&hero.Rarity, &hero.Level, &hero.MaxLevel, &hero.Experience, &hero.ExperienceToNext,
		&hero.Health, &hero.Attack, &hero.Defense, &hero.Speed, &hero.Intelligence, &hero.Charisma,
		&hero.ActiveSkills, &hero.PassiveSkills, &hero.UltimateSkill, &hero.Equipment, &hero.Artifacts,
		&hero.RecruitCost, &hero.UpgradeCost, &hero.Requirements, &hero.Icon, &hero.Portrait,
		&hero.Model, &hero.Color, &hero.IsActive, &hero.IsSpecial, &hero.IsLimited,
		&hero.ReleaseDate, &hero.ExpiryDate, &hero.CreatedAt, &hero.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("héroe no encontrado")
		}
		return nil, fmt.Errorf("error obteniendo héroe: %w", err)
	}

	return &hero, nil
}

// GetPlayerHeroes obtiene los héroes de un jugador
func (r *HeroRepository) GetPlayerHeroes(playerID int) ([]models.PlayerHero, error) {
	query := `
		SELECT id, player_id, hero_id, level, experience, experience_to_next,
		       current_health, max_health, current_attack, current_defense, current_speed,
		       current_intelligence, current_charisma, is_recruited, is_active, is_injured,
		       injury_time, battles_won, battles_lost, quests_completed, experience_gained,
		       equipment, artifacts, unlocked_skills, skill_levels, recruited_at, last_used_at,
		       created_at, updated_at
		FROM player_heroes
		WHERE player_id = $1
		ORDER BY level DESC, experience DESC
	`

	rows, err := r.db.Query(query, playerID)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo héroes del jugador: %w", err)
	}
	defer rows.Close()

	var playerHeroes []models.PlayerHero
	for rows.Next() {
		var ph models.PlayerHero
		err := rows.Scan(
			&ph.ID, &ph.PlayerID, &ph.HeroID, &ph.Level, &ph.Experience, &ph.ExperienceToNext,
			&ph.CurrentHealth, &ph.MaxHealth, &ph.CurrentAttack, &ph.CurrentDefense, &ph.CurrentSpeed,
			&ph.CurrentIntelligence, &ph.CurrentCharisma, &ph.IsRecruited, &ph.IsActive, &ph.IsInjured,
			&ph.InjuryTime, &ph.BattlesWon, &ph.BattlesLost, &ph.QuestsCompleted, &ph.ExperienceGained,
			&ph.Equipment, &ph.Artifacts, &ph.UnlockedSkills, &ph.SkillLevels, &ph.RecruitedAt,
			&ph.LastUsedAt, &ph.CreatedAt, &ph.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando héroe del jugador: %w", err)
		}
		playerHeroes = append(playerHeroes, ph)
	}

	return playerHeroes, nil
}

// GetPlayerHero obtiene un héroe específico de un jugador
func (r *HeroRepository) GetPlayerHero(playerID, heroID int) (*models.PlayerHero, error) {
	query := `
		SELECT id, player_id, hero_id, level, experience, experience_to_next,
		       current_health, max_health, current_attack, current_defense, current_speed,
		       current_intelligence, current_charisma, is_recruited, is_active, is_injured,
		       injury_time, battles_won, battles_lost, quests_completed, experience_gained,
		       equipment, artifacts, unlocked_skills, skill_levels, recruited_at, last_used_at,
		       created_at, updated_at
		FROM player_heroes
		WHERE player_id = $1 AND hero_id = $2
	`

	var ph models.PlayerHero
	err := r.db.QueryRow(query, playerID, heroID).Scan(
		&ph.ID, &ph.PlayerID, &ph.HeroID, &ph.Level, &ph.Experience, &ph.ExperienceToNext,
		&ph.CurrentHealth, &ph.MaxHealth, &ph.CurrentAttack, &ph.CurrentDefense, &ph.CurrentSpeed,
		&ph.CurrentIntelligence, &ph.CurrentCharisma, &ph.IsRecruited, &ph.IsActive, &ph.IsInjured,
		&ph.InjuryTime, &ph.BattlesWon, &ph.BattlesLost, &ph.QuestsCompleted, &ph.ExperienceGained,
		&ph.Equipment, &ph.Artifacts, &ph.UnlockedSkills, &ph.SkillLevels, &ph.RecruitedAt,
		&ph.LastUsedAt, &ph.CreatedAt, &ph.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No tiene este héroe
		}
		return nil, fmt.Errorf("error obteniendo héroe del jugador: %w", err)
	}

	return &ph, nil
}

// RecruitHero recluta un héroe para un jugador
func (r *HeroRepository) RecruitHero(playerID, heroID int) error {
	// Verificar que el héroe existe y está activo
	hero, err := r.GetHero(heroID)
	if err != nil {
		return fmt.Errorf("error obteniendo héroe: %w", err)
	}

	// Verificar que no lo tenga ya
	existingHero, err := r.GetPlayerHero(playerID, heroID)
	if err != nil {
		return fmt.Errorf("error verificando héroe existente: %w", err)
	}

	if existingHero != nil {
		return fmt.Errorf("ya tienes este héroe")
	}

	// Verificar límite de héroes
	config, err := r.GetHeroSystemConfig()
	if err != nil {
		return fmt.Errorf("error obteniendo configuración: %w", err)
	}

	if !config.IsEnabled {
		return fmt.Errorf("el sistema de héroes está deshabilitado")
	}

	playerHeroes, err := r.GetPlayerHeroes(playerID)
	if err != nil {
		return fmt.Errorf("error obteniendo héroes del jugador: %w", err)
	}

	if len(playerHeroes) >= config.MaxHeroesPerPlayer {
		return fmt.Errorf("has alcanzado el límite de héroes (%d)", config.MaxHeroesPerPlayer)
	}

	// Crear el héroe del jugador
	now := time.Now()
	query := `
		INSERT INTO player_heroes (
			player_id, hero_id, level, experience, experience_to_next,
			current_health, max_health, current_attack, current_defense, current_speed,
			current_intelligence, current_charisma, is_recruited, is_active, is_injured,
			battles_won, battles_lost, quests_completed, experience_gained,
			equipment, artifacts, unlocked_skills, skill_levels, recruited_at,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15,
			$16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27
		)
	`

	_, err = r.db.Exec(query,
		playerID, heroID, 1, 0, hero.ExperienceToNext,
		hero.Health, hero.Health, hero.Attack, hero.Defense, hero.Speed,
		hero.Intelligence, hero.Charisma, true, false, false,
		0, 0, 0, 0,
		"{}", "{}", "[]", "{}", &now,
		now, now,
	)

	if err != nil {
		return fmt.Errorf("error reclutando héroe: %w", err)
	}

	return nil
}

// UpgradeHero mejora un héroe
func (r *HeroRepository) UpgradeHero(playerID, heroID int) error {
	// Obtener el héroe del jugador
	playerHero, err := r.GetPlayerHero(playerID, heroID)
	if err != nil {
		return fmt.Errorf("error obteniendo héroe del jugador: %w", err)
	}

	if playerHero == nil {
		return fmt.Errorf("no tienes este héroe")
	}

	// Obtener el héroe base
	hero, err := r.GetHero(heroID)
	if err != nil {
		return fmt.Errorf("error obteniendo héroe base: %w", err)
	}

	// Verificar que no esté en el nivel máximo
	if playerHero.Level >= hero.MaxLevel {
		return fmt.Errorf("el héroe ya está en el nivel máximo")
	}

	// Verificar que tenga suficiente experiencia
	if playerHero.Experience < playerHero.ExperienceToNext {
		return fmt.Errorf("no tienes suficiente experiencia para subir de nivel")
	}

	// Calcular nuevas estadísticas
	newLevel := playerHero.Level + 1
	remainingExp := playerHero.Experience - playerHero.ExperienceToNext
	expToNext := r.calculateExperienceToNext(newLevel)

	// Mejorar estadísticas
	statIncrease := r.calculateStatIncrease(hero, newLevel)
	newMaxHealth := hero.Health + (statIncrease * newLevel)
	newAttack := hero.Attack + (statIncrease * newLevel)
	newDefense := hero.Defense + (statIncrease * newLevel)
	newSpeed := hero.Speed + (statIncrease * newLevel)
	newIntelligence := hero.Intelligence + (statIncrease * newLevel)
	newCharisma := hero.Charisma + (statIncrease * newLevel)

	// Actualizar el héroe
	query := `
		UPDATE player_heroes 
		SET level = $1, experience = $2, experience_to_next = $3,
		    max_health = $4, current_health = $5, current_attack = $6,
		    current_defense = $7, current_speed = $8, current_intelligence = $9,
		    current_charisma = $10, updated_at = $11
		WHERE player_id = $12 AND hero_id = $13
	`

	_, err = r.db.Exec(query,
		newLevel, remainingExp, expToNext,
		newMaxHealth, newMaxHealth, newAttack,
		newDefense, newSpeed, newIntelligence,
		newCharisma, time.Now(),
		playerID, heroID,
	)

	if err != nil {
		return fmt.Errorf("error mejorando héroe: %w", err)
	}

	return nil
}

// ActivateHero activa un héroe
func (r *HeroRepository) ActivateHero(playerID, heroID int) error {
	// Verificar configuración
	config, err := r.GetHeroSystemConfig()
	if err != nil {
		return fmt.Errorf("error obteniendo configuración: %w", err)
	}

	if !config.IsEnabled {
		return fmt.Errorf("el sistema de héroes está deshabilitado")
	}

	// Obtener el héroe del jugador
	playerHero, err := r.GetPlayerHero(playerID, heroID)
	if err != nil {
		return fmt.Errorf("error obteniendo héroe del jugador: %w", err)
	}

	if playerHero == nil {
		return fmt.Errorf("no tienes este héroe")
	}

	if !playerHero.IsRecruited {
		return fmt.Errorf("debes reclutar el héroe primero")
	}

	if playerHero.IsInjured {
		return fmt.Errorf("el héroe está herido y no puede ser activado")
	}

	// Verificar límite de héroes activos
	if config.MaxActiveHeroes > 0 {
		activeHeroes, err := r.GetActiveHeroes(playerID)
		if err != nil {
			return fmt.Errorf("error obteniendo héroes activos: %w", err)
		}

		if len(activeHeroes) >= config.MaxActiveHeroes {
			return fmt.Errorf("has alcanzado el límite de héroes activos (%d)", config.MaxActiveHeroes)
		}
	}

	// Activar el héroe
	query := `
		UPDATE player_heroes 
		SET is_active = true, last_used_at = $1, updated_at = $2
		WHERE player_id = $3 AND hero_id = $4
	`

	_, err = r.db.Exec(query, time.Now(), time.Now(), playerID, heroID)
	if err != nil {
		return fmt.Errorf("error activando héroe: %w", err)
	}

	return nil
}

// DeactivateHero desactiva un héroe
func (r *HeroRepository) DeactivateHero(playerID, heroID int) error {
	query := `
		UPDATE player_heroes 
		SET is_active = false, updated_at = $1
		WHERE player_id = $2 AND hero_id = $3
	`

	_, err := r.db.Exec(query, time.Now(), playerID, heroID)
	if err != nil {
		return fmt.Errorf("error desactivando héroe: %w", err)
	}

	return nil
}

// GetActiveHeroes obtiene los héroes activos de un jugador
func (r *HeroRepository) GetActiveHeroes(playerID int) ([]models.PlayerHero, error) {
	query := `
		SELECT id, player_id, hero_id, level, experience, experience_to_next,
		       current_health, max_health, current_attack, current_defense, current_speed,
		       current_intelligence, current_charisma, is_recruited, is_active, is_injured,
		       injury_time, battles_won, battles_lost, quests_completed, experience_gained,
		       equipment, artifacts, unlocked_skills, skill_levels, recruited_at, last_used_at,
		       created_at, updated_at
		FROM player_heroes
		WHERE player_id = $1 AND is_active = true
		ORDER BY level DESC, experience DESC
	`

	rows, err := r.db.Query(query, playerID)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo héroes activos: %w", err)
	}
	defer rows.Close()

	var activeHeroes []models.PlayerHero
	for rows.Next() {
		var ph models.PlayerHero
		err := rows.Scan(
			&ph.ID, &ph.PlayerID, &ph.HeroID, &ph.Level, &ph.Experience, &ph.ExperienceToNext,
			&ph.CurrentHealth, &ph.MaxHealth, &ph.CurrentAttack, &ph.CurrentDefense, &ph.CurrentSpeed,
			&ph.CurrentIntelligence, &ph.CurrentCharisma, &ph.IsRecruited, &ph.IsActive, &ph.IsInjured,
			&ph.InjuryTime, &ph.BattlesWon, &ph.BattlesLost, &ph.QuestsCompleted, &ph.ExperienceGained,
			&ph.Equipment, &ph.Artifacts, &ph.UnlockedSkills, &ph.SkillLevels, &ph.RecruitedAt,
			&ph.LastUsedAt, &ph.CreatedAt, &ph.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando héroe activo: %w", err)
		}
		activeHeroes = append(activeHeroes, ph)
	}

	return activeHeroes, nil
}

// GetHeroSkills obtiene las habilidades de un héroe
func (r *HeroRepository) GetHeroSkills(heroID int) ([]models.HeroSkill, error) {
	query := `
		SELECT id, hero_id, name, description, type, category, level, max_level,
		       effects, target, range, cooldown, duration, mana_cost, health_cost,
		       icon, animation, sound, is_active, is_ultimate, created_at
		FROM hero_skills
		WHERE hero_id = $1 AND is_active = true
		ORDER BY type, level
	`

	rows, err := r.db.Query(query, heroID)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo habilidades: %w", err)
	}
	defer rows.Close()

	var skills []models.HeroSkill
	for rows.Next() {
		var skill models.HeroSkill
		err := rows.Scan(
			&skill.ID, &skill.HeroID, &skill.Name, &skill.Description, &skill.Type,
			&skill.Category, &skill.Level, &skill.MaxLevel, &skill.Effects, &skill.Target,
			&skill.Range, &skill.Cooldown, &skill.Duration, &skill.ManaCost, &skill.HealthCost,
			&skill.Icon, &skill.Animation, &skill.Sound, &skill.IsActive, &skill.IsUltimate,
			&skill.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando habilidad: %w", err)
		}
		skills = append(skills, skill)
	}

	return skills, nil
}

// GetHeroEquipment obtiene el equipamiento de un héroe
func (r *HeroRepository) GetHeroEquipment(heroID int) ([]models.HeroEquipment, error) {
	query := `
		SELECT id, hero_id, slot, name, description, type, rarity, level, max_level,
		       stats, requirements, cost, upgrade_cost, icon, model, color,
		       is_active, is_special, created_at
		FROM hero_equipment
		WHERE hero_id = $1 AND is_active = true
		ORDER BY slot, level
	`

	rows, err := r.db.Query(query, heroID)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo equipamiento: %w", err)
	}
	defer rows.Close()

	var equipment []models.HeroEquipment
	for rows.Next() {
		var eq models.HeroEquipment
		err := rows.Scan(
			&eq.ID, &eq.HeroID, &eq.Slot, &eq.Name, &eq.Description, &eq.Type,
			&eq.Rarity, &eq.Level, &eq.MaxLevel, &eq.Stats, &eq.Requirements,
			&eq.Cost, &eq.UpgradeCost, &eq.Icon, &eq.Model, &eq.Color,
			&eq.IsActive, &eq.IsSpecial, &eq.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando equipamiento: %w", err)
		}
		equipment = append(equipment, eq)
	}

	return equipment, nil
}

// GetHeroQuests obtiene las misiones de un héroe
func (r *HeroRepository) GetHeroQuests(heroID int) ([]models.HeroQuest, error) {
	query := `
		SELECT id, hero_id, name, description, type, category, level, difficulty,
		       objectives, requirements, rewards, experience_reward, duration,
		       time_limit, is_active, is_repeatable, is_event, created_at
		FROM hero_quests
		WHERE hero_id = $1 AND is_active = true
		ORDER BY level, difficulty
	`

	rows, err := r.db.Query(query, heroID)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo misiones: %w", err)
	}
	defer rows.Close()

	var quests []models.HeroQuest
	for rows.Next() {
		var quest models.HeroQuest
		err := rows.Scan(
			&quest.ID, &quest.HeroID, &quest.Name, &quest.Description, &quest.Type,
			&quest.Category, &quest.Level, &quest.Difficulty, &quest.Objectives,
			&quest.Requirements, &quest.Rewards, &quest.ExperienceReward, &quest.Duration,
			&quest.TimeLimit, &quest.IsActive, &quest.IsRepeatable, &quest.IsEvent,
			&quest.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando misión: %w", err)
		}
		quests = append(quests, quest)
	}

	return quests, nil
}

// GetHeroRankings obtiene los rankings de héroes
func (r *HeroRepository) GetHeroRankings(limit int) ([]models.HeroRanking, error) {
	query := `
		WITH hero_stats AS (
			SELECT 
				ph.player_id,
				p.username as player_name,
				ph.hero_id,
				h.name as hero_name,
				ph.level,
				ph.experience,
				ph.battles_won,
				ph.battles_lost,
				CASE 
					WHEN (ph.battles_won + ph.battles_lost) > 0 
					THEN (ph.battles_won::float / (ph.battles_won + ph.battles_lost) * 100)
					ELSE 0 
				END as win_rate,
				(ph.current_attack + ph.current_defense + ph.current_speed + 
				 ph.current_intelligence + ph.current_charisma) as total_power
			FROM player_heroes ph
			JOIN players p ON ph.player_id = p.id
			JOIN heroes h ON ph.hero_id = h.id
			WHERE ph.is_recruited = true
		)
		SELECT 
			player_id, player_name, hero_id, hero_name, level, experience,
			battles_won, battles_lost, win_rate, total_power,
			RANK() OVER (ORDER BY total_power DESC, level DESC, experience DESC) as rank
		FROM hero_stats
		ORDER BY rank
		LIMIT $1
	`

	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo rankings: %w", err)
	}
	defer rows.Close()

	var rankings []models.HeroRanking
	for rows.Next() {
		var ranking models.HeroRanking
		err := rows.Scan(
			&ranking.PlayerID, &ranking.PlayerName, &ranking.HeroID, &ranking.HeroName,
			&ranking.Level, &ranking.Experience, &ranking.BattlesWon, &ranking.BattlesLost,
			&ranking.WinRate, &ranking.TotalPower, &ranking.Rank,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando ranking: %w", err)
		}
		rankings = append(rankings, ranking)
	}

	return rankings, nil
}

// GetHeroSystemConfig obtiene la configuración del sistema de héroes
func (r *HeroRepository) GetHeroSystemConfig() (*models.HeroSystemConfig, error) {
	query := `
		SELECT id, is_enabled, max_heroes_per_player, max_active_heroes,
		       experience_multiplier, injury_duration, recovery_cost,
		       advanced_config, created_at, updated_at
		FROM hero_system_config
		ORDER BY id DESC
		LIMIT 1
	`

	var config models.HeroSystemConfig
	err := r.db.QueryRow(query).Scan(
		&config.ID, &config.IsEnabled, &config.MaxHeroesPerPlayer, &config.MaxActiveHeroes,
		&config.ExperienceMultiplier, &config.InjuryDuration, &config.RecoveryCost,
		&config.AdvancedConfig, &config.CreatedAt, &config.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			// Crear configuración por defecto
			return r.createDefaultHeroConfig()
		}
		return nil, fmt.Errorf("error obteniendo configuración: %w", err)
	}

	return &config, nil
}

// createDefaultHeroConfig crea una configuración por defecto
func (r *HeroRepository) createDefaultHeroConfig() (*models.HeroSystemConfig, error) {
	query := `
		INSERT INTO hero_system_config (
			is_enabled, max_heroes_per_player, max_active_heroes,
			experience_multiplier, injury_duration, recovery_cost,
			advanced_config, created_at, updated_at
		) VALUES (
			true, 10, 3, 1.0, 3600, '{"gold": 100}',
			'{"enable_auto_recovery": true, "enable_hero_events": true}',
			$1, $1
		) RETURNING id, is_enabled, max_heroes_per_player, max_active_heroes,
		            experience_multiplier, injury_duration, recovery_cost,
		            advanced_config, created_at, updated_at
	`

	var config models.HeroSystemConfig
	now := time.Now()
	err := r.db.QueryRow(query, now).Scan(
		&config.ID, &config.IsEnabled, &config.MaxHeroesPerPlayer, &config.MaxActiveHeroes,
		&config.ExperienceMultiplier, &config.InjuryDuration, &config.RecoveryCost,
		&config.AdvancedConfig, &config.CreatedAt, &config.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("error creando configuración por defecto: %w", err)
	}

	return &config, nil
}

// calculateExperienceToNext calcula la experiencia necesaria para el siguiente nivel
func (r *HeroRepository) calculateExperienceToNext(level int) int {
	// Fórmula: 100 * level^1.5
	return int(100 * float64(level) * float64(level) * 0.5)
}

// calculateStatIncrease calcula el aumento de estadísticas por nivel
func (r *HeroRepository) calculateStatIncrease(hero *models.Hero, level int) int {
	// Aumento base de 5 puntos por nivel
	return 5
}

// UpdateHeroSystemConfig actualiza la configuración del sistema de héroes
func (r *HeroRepository) UpdateHeroSystemConfig(config *models.HeroSystemConfig) error {
	query := `
		UPDATE hero_system_config 
		SET is_enabled = $1, max_heroes_per_player = $2, max_active_heroes = $3,
		    experience_multiplier = $4, injury_duration = $5, recovery_cost = $6,
		    advanced_config = $7, updated_at = $8
		WHERE id = $9
	`

	_, err := r.db.Exec(query,
		config.IsEnabled, config.MaxHeroesPerPlayer, config.MaxActiveHeroes,
		config.ExperienceMultiplier, config.InjuryDuration, config.RecoveryCost,
		config.AdvancedConfig, time.Now(), config.ID,
	)

	if err != nil {
		return fmt.Errorf("error actualizando configuración: %w", err)
	}

	return nil
}
