package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"server-backend/models"

	"go.uber.org/zap"
)

type RankingRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewRankingRepository(db *sql.DB, logger *zap.Logger) *RankingRepository {
	return &RankingRepository{
		db:     db,
		logger: logger,
	}
}

// GetRankingCategories obtiene todas las categorías de ranking
func (r *RankingRepository) GetRankingCategories(activeOnly bool) ([]models.RankingCategory, error) {
	query := `
		SELECT id, name, description, type, sub_type, icon, color,
		       update_interval, max_positions, min_score, score_formula,
		       rewards_enabled, rewards_config, display_order, is_public,
		       show_in_dashboard, is_active, last_updated, created_at
		FROM ranking_categories
	`

	if activeOnly {
		query += " WHERE is_active = true"
	}

	query += " ORDER BY display_order ASC, name ASC"

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo categorías de ranking: %w", err)
	}
	defer rows.Close()

	var categories []models.RankingCategory
	for rows.Next() {
		var category models.RankingCategory
		err := rows.Scan(
			&category.ID, &category.Name, &category.Description, &category.Type,
			&category.SubType, &category.Icon, &category.Color, &category.UpdateInterval,
			&category.MaxPositions, &category.MinScore, &category.ScoreFormula,
			&category.RewardsEnabled, &category.RewardsConfig, &category.DisplayOrder,
			&category.IsPublic, &category.ShowInDashboard, &category.IsActive,
			&category.LastUpdated, &category.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando categoría: %w", err)
		}
		categories = append(categories, category)
	}

	return categories, nil
}

// GetRankingCategory obtiene una categoría específica
func (r *RankingRepository) GetRankingCategory(categoryID int) (*models.RankingCategory, error) {
	query := `
		SELECT id, name, description, type, sub_type, icon, color,
		       update_interval, max_positions, min_score, score_formula,
		       rewards_enabled, rewards_config, display_order, is_public,
		       show_in_dashboard, is_active, last_updated, created_at
		FROM ranking_categories
		WHERE id = $1
	`

	var category models.RankingCategory
	err := r.db.QueryRow(query, categoryID).Scan(
		&category.ID, &category.Name, &category.Description, &category.Type,
		&category.SubType, &category.Icon, &category.Color, &category.UpdateInterval,
		&category.MaxPositions, &category.MinScore, &category.ScoreFormula,
		&category.RewardsEnabled, &category.RewardsConfig, &category.DisplayOrder,
		&category.IsPublic, &category.ShowInDashboard, &category.IsActive,
		&category.LastUpdated, &category.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("categoría no encontrada")
		}
		return nil, fmt.Errorf("error obteniendo categoría: %w", err)
	}

	return &category, nil
}

// CreateRankingCategory crea una nueva categoría de ranking
func (r *RankingRepository) CreateRankingCategory(category *models.RankingCategory) error {
	query := `
		INSERT INTO ranking_categories (
			name, description, type, sub_type, icon, color,
			update_interval, max_positions, min_score, score_formula,
			rewards_enabled, rewards_config, display_order, is_public,
			show_in_dashboard, is_active, last_updated, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
			$11, $12, $13, $14, $15, $16, $17, $18
		) RETURNING id
	`

	now := time.Now()
	err := r.db.QueryRow(query,
		category.Name, category.Description, category.Type, category.SubType,
		category.Icon, category.Color, category.UpdateInterval, category.MaxPositions,
		category.MinScore, category.ScoreFormula, category.RewardsEnabled,
		category.RewardsConfig, category.DisplayOrder, category.IsPublic,
		category.ShowInDashboard, category.IsActive, now, now,
	).Scan(&category.ID)

	if err != nil {
		return fmt.Errorf("error creando categoría: %w", err)
	}

	return nil
}

// UpdateRankingCategory actualiza una categoría de ranking
func (r *RankingRepository) UpdateRankingCategory(category *models.RankingCategory) error {
	query := `
		UPDATE ranking_categories 
		SET name = $1, description = $2, type = $3, sub_type = $4,
		    icon = $5, color = $6, update_interval = $7, max_positions = $8,
		    min_score = $9, score_formula = $10, rewards_enabled = $11,
		    rewards_config = $12, display_order = $13, is_public = $14,
		    show_in_dashboard = $15, is_active = $16, last_updated = $17
		WHERE id = $18
	`

	_, err := r.db.Exec(query,
		category.Name, category.Description, category.Type, category.SubType,
		category.Icon, category.Color, category.UpdateInterval, category.MaxPositions,
		category.MinScore, category.ScoreFormula, category.RewardsEnabled,
		category.RewardsConfig, category.DisplayOrder, category.IsPublic,
		category.ShowInDashboard, category.IsActive, time.Now(), category.ID,
	)

	if err != nil {
		return fmt.Errorf("error actualizando categoría: %w", err)
	}

	return nil
}

// GetRankingSeasons obtiene las temporadas de ranking
func (r *RankingRepository) GetRankingSeasons(status string) ([]models.RankingSeason, error) {
	query := `
		SELECT id, name, description, season_number, start_date, end_date,
		       is_active, categories, rewards_enabled, rewards_config,
		       total_participants, total_alliances, total_villages, status,
		       created_at, updated_at
		FROM ranking_seasons
	`

	if status != "" {
		query += " WHERE status = $1"
	}

	query += " ORDER BY season_number DESC"

	var rows *sql.Rows
	var err error

	if status != "" {
		rows, err = r.db.Query(query, status)
	} else {
		rows, err = r.db.Query(query)
	}

	if err != nil {
		return nil, fmt.Errorf("error obteniendo temporadas: %w", err)
	}
	defer rows.Close()

	var seasons []models.RankingSeason
	for rows.Next() {
		var season models.RankingSeason
		err := rows.Scan(
			&season.ID, &season.Name, &season.Description, &season.SeasonNumber,
			&season.StartDate, &season.EndDate, &season.IsActive, &season.Categories,
			&season.RewardsEnabled, &season.RewardsConfig, &season.TotalParticipants,
			&season.TotalAlliances, &season.TotalVillages, &season.Status,
			&season.CreatedAt, &season.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando temporada: %w", err)
		}
		seasons = append(seasons, season)
	}

	return seasons, nil
}

// GetRankingEntries obtiene las entradas de un ranking
func (r *RankingRepository) GetRankingEntries(categoryID int, seasonID *int, limit int) ([]models.RankingEntry, error) {
	query := `
		SELECT id, category_id, season_id, entity_type, entity_id, entity_name,
		       position, score, previous_position, position_change, stats,
		       breakdown, last_updated, created_at
		FROM ranking_entries
		WHERE category_id = $1
	`

	args := []interface{}{categoryID}
	argCount := 2

	if seasonID != nil {
		query += fmt.Sprintf(" AND season_id = $%d", argCount)
		args = append(args, *seasonID)
		argCount++
	} else {
		query += " AND season_id IS NULL"
	}

	query += " ORDER BY position ASC"

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, limit)
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo entradas de ranking: %w", err)
	}
	defer rows.Close()

	var entries []models.RankingEntry
	for rows.Next() {
		var entry models.RankingEntry
		err := rows.Scan(
			&entry.ID, &entry.CategoryID, &entry.SeasonID, &entry.EntityType,
			&entry.EntityID, &entry.EntityName, &entry.Position, &entry.Score,
			&entry.PreviousPosition, &entry.PositionChange, &entry.Stats,
			&entry.Breakdown, &entry.LastUpdated, &entry.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando entrada: %w", err)
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

// GetPlayerStatistics obtiene las estadísticas de un jugador
func (r *RankingRepository) GetPlayerStatistics(playerID int) (*models.PlayerStatistics, error) {
	query := `
		SELECT id, player_id, total_play_time, days_active, last_active,
		       villages_owned, total_population, buildings_built, buildings_upgraded,
		       battles_won, battles_lost, battles_total, win_rate,
		       units_trained, units_lost, units_killed,
		       heroes_recruited, heroes_upgraded, heroes_active,
		       technologies_researched, research_points,
		       total_earned, total_spent, market_transactions, items_sold, items_bought,
		       alliances_joined, current_alliance, alliance_contribution,
		       achievements_earned, total_points,
		       best_ranking, current_ranking, rankings_won,
		       events_participated, events_won,
		       resources_gathered, resources_spent,
		       first_login, last_updated, created_at
		FROM player_statistics
		WHERE player_id = $1
	`

	var stats models.PlayerStatistics
	err := r.db.QueryRow(query, playerID).Scan(
		&stats.ID, &stats.PlayerID, &stats.TotalPlayTime, &stats.DaysActive,
		&stats.LastActive, &stats.VillagesOwned, &stats.TotalPopulation,
		&stats.BuildingsBuilt, &stats.BuildingsUpgraded, &stats.BattlesWon,
		&stats.BattlesLost, &stats.BattlesTotal, &stats.WinRate,
		&stats.UnitsTrained, &stats.UnitsLost, &stats.UnitsKilled,
		&stats.HeroesRecruited, &stats.HeroesUpgraded, &stats.HeroesActive,
		&stats.TechnologiesResearched, &stats.ResearchPoints,
		&stats.TotalEarned, &stats.TotalSpent, &stats.MarketTransactions,
		&stats.ItemsSold, &stats.ItemsBought, &stats.AlliancesJoined,
		&stats.CurrentAlliance, &stats.AllianceContribution,
		&stats.AchievementsEarned, &stats.TotalPoints,
		&stats.BestRanking, &stats.CurrentRanking, &stats.RankingsWon,
		&stats.EventsParticipated, &stats.EventsWon,
		&stats.ResourcesGathered, &stats.ResourcesSpent,
		&stats.FirstLogin, &stats.LastUpdated, &stats.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("estadísticas de jugador no encontradas")
		}
		return nil, fmt.Errorf("error obteniendo estadísticas de jugador: %w", err)
	}

	return &stats, nil
}

// CreateOrUpdatePlayerStatistics crea o actualiza estadísticas de jugador
func (r *RankingRepository) CreateOrUpdatePlayerStatistics(stats *models.PlayerStatistics) error {
	query := `
		INSERT INTO player_statistics (
			player_id, total_play_time, days_active, last_active,
			villages_owned, total_population, buildings_built, buildings_upgraded,
			battles_won, battles_lost, battles_total, win_rate,
			units_trained, units_lost, units_killed,
			heroes_recruited, heroes_upgraded, heroes_active,
			technologies_researched, research_points,
			total_earned, total_spent, market_transactions, items_sold, items_bought,
			alliances_joined, current_alliance, alliance_contribution,
			achievements_earned, total_points,
			best_ranking, current_ranking, rankings_won,
			events_participated, events_won,
			resources_gathered, resources_spent,
			first_login, last_updated, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15,
			$16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28,
			$29, $30, $31, $32, $33, $34, $35, $36, $37, $38, $39, $40
		) ON CONFLICT (player_id) DO UPDATE SET
			total_play_time = EXCLUDED.total_play_time,
			days_active = EXCLUDED.days_active,
			last_active = EXCLUDED.last_active,
			villages_owned = EXCLUDED.villages_owned,
			total_population = EXCLUDED.total_population,
			buildings_built = EXCLUDED.buildings_built,
			buildings_upgraded = EXCLUDED.buildings_upgraded,
			battles_won = EXCLUDED.battles_won,
			battles_lost = EXCLUDED.battles_lost,
			battles_total = EXCLUDED.battles_total,
			win_rate = EXCLUDED.win_rate,
			units_trained = EXCLUDED.units_trained,
			units_lost = EXCLUDED.units_lost,
			units_killed = EXCLUDED.units_killed,
			heroes_recruited = EXCLUDED.heroes_recruited,
			heroes_upgraded = EXCLUDED.heroes_upgraded,
			heroes_active = EXCLUDED.heroes_active,
			technologies_researched = EXCLUDED.technologies_researched,
			research_points = EXCLUDED.research_points,
			total_earned = EXCLUDED.total_earned,
			total_spent = EXCLUDED.total_spent,
			market_transactions = EXCLUDED.market_transactions,
			items_sold = EXCLUDED.items_sold,
			items_bought = EXCLUDED.items_bought,
			alliances_joined = EXCLUDED.alliances_joined,
			current_alliance = EXCLUDED.current_alliance,
			alliance_contribution = EXCLUDED.alliance_contribution,
			achievements_earned = EXCLUDED.achievements_earned,
			total_points = EXCLUDED.total_points,
			best_ranking = EXCLUDED.best_ranking,
			current_ranking = EXCLUDED.current_ranking,
			rankings_won = EXCLUDED.rankings_won,
			events_participated = EXCLUDED.events_participated,
			events_won = EXCLUDED.events_won,
			resources_gathered = EXCLUDED.resources_gathered,
			resources_spent = EXCLUDED.resources_spent,
			last_updated = EXCLUDED.last_updated
		RETURNING id
	`

	now := time.Now()
	err := r.db.QueryRow(query,
		stats.PlayerID, stats.TotalPlayTime, stats.DaysActive, stats.LastActive,
		stats.VillagesOwned, stats.TotalPopulation, stats.BuildingsBuilt,
		stats.BuildingsUpgraded, stats.BattlesWon, stats.BattlesLost,
		stats.BattlesTotal, stats.WinRate, stats.UnitsTrained, stats.UnitsLost,
		stats.UnitsKilled, stats.HeroesRecruited, stats.HeroesUpgraded,
		stats.HeroesActive, stats.TechnologiesResearched, stats.ResearchPoints,
		stats.TotalEarned, stats.TotalSpent, stats.MarketTransactions,
		stats.ItemsSold, stats.ItemsBought, stats.AlliancesJoined,
		stats.CurrentAlliance, stats.AllianceContribution, stats.AchievementsEarned,
		stats.TotalPoints, stats.BestRanking, stats.CurrentRanking,
		stats.RankingsWon, stats.EventsParticipated, stats.EventsWon,
		stats.ResourcesGathered, stats.ResourcesSpent, stats.FirstLogin,
		now, now,
	).Scan(&stats.ID)

	if err != nil {
		return fmt.Errorf("error creando/actualizando estadísticas: %w", err)
	}

	return nil
}

// GetAllianceStatistics obtiene las estadísticas de una alianza
func (r *RankingRepository) GetAllianceStatistics(allianceID int) (*models.AllianceStatistics, error) {
	query := `
		SELECT id, alliance_id, total_members, active_members, new_members,
		       total_villages, total_population, battles_won, battles_lost, win_rate,
		       total_units, units_trained, units_lost, total_heroes, heroes_active,
		       technologies_researched, research_points, total_earned, total_spent,
		       market_volume, best_ranking, current_ranking, rankings_won,
		       events_participated, events_won, average_activity, last_activity,
		       founded_date, last_updated, created_at
		FROM alliance_statistics
		WHERE alliance_id = $1
	`

	var stats models.AllianceStatistics
	err := r.db.QueryRow(query, allianceID).Scan(
		&stats.ID, &stats.AllianceID, &stats.TotalMembers, &stats.ActiveMembers,
		&stats.NewMembers, &stats.TotalVillages, &stats.TotalPopulation,
		&stats.BattlesWon, &stats.BattlesLost, &stats.WinRate, &stats.TotalUnits,
		&stats.UnitsTrained, &stats.UnitsLost, &stats.TotalHeroes,
		&stats.HeroesActive, &stats.TechnologiesResearched, &stats.ResearchPoints,
		&stats.TotalEarned, &stats.TotalSpent, &stats.MarketVolume,
		&stats.BestRanking, &stats.CurrentRanking, &stats.RankingsWon,
		&stats.EventsParticipated, &stats.EventsWon, &stats.AverageActivity,
		&stats.LastActivity, &stats.FoundedDate, &stats.LastUpdated, &stats.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("estadísticas de alianza no encontradas")
		}
		return nil, fmt.Errorf("error obteniendo estadísticas de alianza: %w", err)
	}

	return &stats, nil
}

// GetVillageStatistics obtiene las estadísticas de una aldea
func (r *RankingRepository) GetVillageStatistics(villageID int) (*models.VillageStatistics, error) {
	query := `
		SELECT id, village_id, player_id, population, max_population, happiness,
		       buildings_built, buildings_upgraded, total_building_levels,
		       resources_produced, resources_consumed, resources_stored,
		       units_trained, units_stationed, units_lost, defense_strength,
		       attacks_defended, attacks_suffered, technologies_researched,
		       research_points, market_transactions, items_sold, items_bought,
		       heroes_assigned, heroes_active, last_activity, activity_score,
		       founded_date, last_updated, created_at
		FROM village_statistics
		WHERE village_id = $1
	`

	var stats models.VillageStatistics
	err := r.db.QueryRow(query, villageID).Scan(
		&stats.ID, &stats.VillageID, &stats.PlayerID, &stats.Population,
		&stats.MaxPopulation, &stats.Happiness, &stats.BuildingsBuilt,
		&stats.BuildingsUpgraded, &stats.TotalBuildingLevels,
		&stats.ResourcesProduced, &stats.ResourcesConsumed, &stats.ResourcesStored,
		&stats.UnitsTrained, &stats.UnitsStationed, &stats.UnitsLost,
		&stats.DefenseStrength, &stats.AttacksDefended, &stats.AttacksSuffered,
		&stats.TechnologiesResearched, &stats.ResearchPoints,
		&stats.MarketTransactions, &stats.ItemsSold, &stats.ItemsBought,
		&stats.HeroesAssigned, &stats.HeroesActive, &stats.LastActivity,
		&stats.ActivityScore, &stats.FoundedDate, &stats.LastUpdated, &stats.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("estadísticas de aldea no encontradas")
		}
		return nil, fmt.Errorf("error obteniendo estadísticas de aldea: %w", err)
	}

	return &stats, nil
}

// GetWorldStatistics obtiene las estadísticas de un mundo
func (r *RankingRepository) GetWorldStatistics(worldID int) (*models.WorldStatistics, error) {
	query := `
		SELECT id, world_id, total_players, active_players, new_players,
		       total_alliances, active_alliances, total_villages, active_villages,
		       total_battles, battles_today, average_battle_duration,
		       total_units, units_trained, units_lost, total_heroes,
		       heroes_recruited, heroes_active, technologies_researched,
		       research_points, total_market_volume, market_transactions,
		       total_taxes, resources_produced, resources_consumed,
		       active_events, events_completed, average_activity, peak_activity,
		       world_start_date, last_updated, created_at
		FROM world_statistics
		WHERE world_id = $1
	`

	var stats models.WorldStatistics
	err := r.db.QueryRow(query, worldID).Scan(
		&stats.ID, &stats.WorldID, &stats.TotalPlayers, &stats.ActivePlayers,
		&stats.NewPlayers, &stats.TotalAlliances, &stats.ActiveAlliances,
		&stats.TotalVillages, &stats.ActiveVillages, &stats.TotalBattles,
		&stats.BattlesToday, &stats.AverageBattleDuration, &stats.TotalUnits,
		&stats.UnitsTrained, &stats.UnitsLost, &stats.TotalHeroes,
		&stats.HeroesRecruited, &stats.HeroesActive, &stats.TechnologiesResearched,
		&stats.ResearchPoints, &stats.TotalMarketVolume, &stats.MarketTransactions,
		&stats.TotalTaxes, &stats.ResourcesProduced, &stats.ResourcesConsumed,
		&stats.ActiveEvents, &stats.EventsCompleted, &stats.AverageActivity,
		&stats.PeakActivity, &stats.WorldStartDate, &stats.LastUpdated, &stats.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("estadísticas de mundo no encontradas")
		}
		return nil, fmt.Errorf("error obteniendo estadísticas de mundo: %w", err)
	}

	return &stats, nil
}

// GetStatisticsSummary obtiene un resumen de estadísticas globales
func (r *RankingRepository) GetStatisticsSummary() (*models.StatisticsSummary, error) {
	query := `
		SELECT 
			COUNT(DISTINCT ps.player_id) as total_players,
			COUNT(DISTINCT CASE WHEN ps.last_active > NOW() - INTERVAL '7 days' THEN ps.player_id END) as active_players,
			COUNT(DISTINCT als.alliance_id) as total_alliances,
			COUNT(DISTINCT vs.village_id) as total_villages,
			SUM(ps.battles_total) as total_battles,
			SUM(ps.units_trained) as total_units,
			SUM(ps.heroes_recruited) as total_heroes,
			SUM(ps.total_earned) as market_volume,
			SUM(ps.research_points) as research_points,
			MAX(ps.last_updated) as last_updated
		FROM player_statistics ps
		LEFT JOIN alliance_statistics als ON true
		LEFT JOIN village_statistics vs ON true
	`

	var summary models.StatisticsSummary
	err := r.db.QueryRow(query).Scan(
		&summary.TotalPlayers, &summary.ActivePlayers, &summary.TotalAlliances,
		&summary.TotalVillages, &summary.TotalBattles, &summary.TotalUnits,
		&summary.TotalHeroes, &summary.MarketVolume, &summary.ResearchPoints,
		&summary.LastUpdated,
	)

	if err != nil {
		return nil, fmt.Errorf("error obteniendo resumen de estadísticas: %w", err)
	}

	return &summary, nil
}

// UpdateRankingEntry actualiza una entrada de ranking
func (r *RankingRepository) UpdateRankingEntry(entry *models.RankingEntry) error {
	query := `
		INSERT INTO ranking_entries (
			category_id, season_id, entity_type, entity_id, entity_name,
			position, score, previous_position, position_change, stats,
			breakdown, last_updated, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		) ON CONFLICT (category_id, season_id, entity_type, entity_id) DO UPDATE SET
			entity_name = EXCLUDED.entity_name,
			position = EXCLUDED.position,
			score = EXCLUDED.score,
			previous_position = EXCLUDED.previous_position,
			position_change = EXCLUDED.position_change,
			stats = EXCLUDED.stats,
			breakdown = EXCLUDED.breakdown,
			last_updated = EXCLUDED.last_updated
		RETURNING id
	`

	now := time.Now()
	err := r.db.QueryRow(query,
		entry.CategoryID, entry.SeasonID, entry.EntityType, entry.EntityID,
		entry.EntityName, entry.Position, entry.Score, entry.PreviousPosition,
		entry.PositionChange, entry.Stats, entry.Breakdown, now, now,
	).Scan(&entry.ID)

	if err != nil {
		return fmt.Errorf("error actualizando entrada de ranking: %w", err)
	}

	return nil
}

// GetRankingHistory obtiene el historial de posiciones
func (r *RankingRepository) GetRankingHistory(categoryID int, entityType string, entityID int, days int) ([]models.RankingHistory, error) {
	query := `
		SELECT id, category_id, season_id, entity_type, entity_id,
		       position, score, recorded_at, created_at
		FROM ranking_history
		WHERE category_id = $1 AND entity_type = $2 AND entity_id = $3
		AND recorded_at > NOW() - INTERVAL '1 day' * $4
		ORDER BY recorded_at ASC
	`

	rows, err := r.db.Query(query, categoryID, entityType, entityID, days)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo historial: %w", err)
	}
	defer rows.Close()

	var history []models.RankingHistory
	for rows.Next() {
		var record models.RankingHistory
		err := rows.Scan(
			&record.ID, &record.CategoryID, &record.SeasonID, &record.EntityType,
			&record.EntityID, &record.Position, &record.Score, &record.RecordedAt,
			&record.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando historial: %w", err)
		}
		history = append(history, record)
	}

	return history, nil
}

// CalculatePlayerScore calcula la puntuación de un jugador para una categoría
func (r *RankingRepository) CalculatePlayerScore(playerID int, categoryID int) (int, error) {
	// Obtener la fórmula de puntuación de la categoría
	category, err := r.GetRankingCategory(categoryID)
	if err != nil {
		return 0, fmt.Errorf("error obteniendo categoría: %w", err)
	}

	// Obtener estadísticas del jugador
	stats, err := r.GetPlayerStatistics(playerID)
	if err != nil {
		return 0, fmt.Errorf("error obteniendo estadísticas: %w", err)
	}

	// Calcular puntuación basada en la fórmula
	score := r.calculateScoreFromFormula(category.ScoreFormula, stats)
	return score, nil
}

// calculateScoreFromFormula calcula la puntuación usando la fórmula JSON
func (r *RankingRepository) calculateScoreFromFormula(formulaJSON string, stats *models.PlayerStatistics) int {
	// Parsear la fórmula JSON
	var formula map[string]interface{}
	if err := json.Unmarshal([]byte(formulaJSON), &formula); err != nil {
		r.logger.Error("Error parseando fórmula de puntuación", zap.Error(err))
		return 0
	}

	// Implementar cálculo de puntuación basado en la fórmula
	// Por ahora, un cálculo simple basado en el tipo de categoría
	score := 0

	// Ejemplo de fórmulas por tipo
	switch formula["type"] {
	case "combat":
		score = stats.BattlesWon*100 + stats.UnitsKilled*10 + stats.HeroesActive*1000
	case "economy":
		score = stats.TotalEarned + stats.MarketTransactions*100 + stats.ItemsSold*50
	case "research":
		score = stats.ResearchPoints + stats.TechnologiesResearched*1000
	case "building":
		score = stats.BuildingsBuilt*100 + stats.BuildingsUpgraded*200 + stats.TotalPopulation*10
	case "activity":
		score = stats.DaysActive*1000 + stats.TotalPlayTime/60 // convertir a horas
	default:
		score = stats.TotalPoints
	}

	return score
}

// GetRankingsByType obtiene rankings por tipo
func (r *RankingRepository) GetRankingsByType(rankingType string, limit int) ([]models.RankingEntry, error) {
	query := `
		SELECT re.id, re.category_id, re.season_id, re.entity_type, re.entity_id,
		       re.entity_name, re.position, re.score, re.previous_position, re.position_change,
		       re.stats, re.breakdown, re.last_updated, re.created_at,
		       rc.name as category_name, rc.type as category_type
		FROM ranking_entries re
		JOIN ranking_categories rc ON re.category_id = rc.id
		WHERE rc.type = $1
		ORDER BY re.score DESC
		LIMIT $2
	`

	rows, err := r.db.Query(query, rankingType, limit)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo rankings por tipo: %w", err)
	}
	defer rows.Close()

	var entries []models.RankingEntry
	for rows.Next() {
		var entry models.RankingEntry
		var categoryName, categoryType string
		err := rows.Scan(
			&entry.ID, &entry.CategoryID, &entry.SeasonID, &entry.EntityType,
			&entry.EntityID, &entry.EntityName, &entry.Position, &entry.Score,
			&entry.PreviousPosition, &entry.PositionChange, &entry.Stats,
			&entry.Breakdown, &entry.LastUpdated, &entry.CreatedAt, &categoryName, &categoryType,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando entrada de ranking: %w", err)
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

// GetRankingsBySeason obtiene rankings por temporada
func (r *RankingRepository) GetRankingsBySeason(seasonID int, limit int) ([]models.RankingEntry, error) {
	query := `
		SELECT re.id, re.category_id, re.season_id, re.entity_type, re.entity_id,
		       re.entity_name, re.position, re.score, re.previous_position, re.position_change,
		       re.stats, re.breakdown, re.last_updated, re.created_at,
		       rc.name as category_name, rc.type as category_type
		FROM ranking_entries re
		JOIN ranking_categories rc ON re.category_id = rc.id
		WHERE re.season_id = $1
		ORDER BY re.category_id, re.score DESC
		LIMIT $2
	`

	rows, err := r.db.Query(query, seasonID, limit)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo rankings por temporada: %w", err)
	}
	defer rows.Close()

	var entries []models.RankingEntry
	for rows.Next() {
		var entry models.RankingEntry
		var categoryName, categoryType string
		err := rows.Scan(
			&entry.ID, &entry.CategoryID, &entry.SeasonID, &entry.EntityType,
			&entry.EntityID, &entry.EntityName, &entry.Position, &entry.Score,
			&entry.PreviousPosition, &entry.PositionChange, &entry.Stats,
			&entry.Breakdown, &entry.LastUpdated, &entry.CreatedAt, &categoryName, &categoryType,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando entrada de ranking: %w", err)
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

// GetRankingsByDateRange obtiene rankings en un rango de fechas
func (r *RankingRepository) GetRankingsByDateRange(startDate, endDate time.Time, limit int) ([]models.RankingEntry, error) {
	query := `
		SELECT re.id, re.category_id, re.season_id, re.entity_type, re.entity_id,
		       re.entity_name, re.position, re.score, re.previous_position, re.position_change,
		       re.stats, re.breakdown, re.last_updated, re.created_at,
		       rc.name as category_name, rc.type as category_type
		FROM ranking_entries re
		JOIN ranking_categories rc ON re.category_id = rc.id
		WHERE re.last_updated >= $1 AND re.last_updated <= $2
		ORDER BY re.last_updated DESC, re.score DESC
		LIMIT $3
	`

	rows, err := r.db.Query(query, startDate, endDate, limit)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo rankings por rango de fechas: %w", err)
	}
	defer rows.Close()

	var entries []models.RankingEntry
	for rows.Next() {
		var entry models.RankingEntry
		var categoryName, categoryType string
		err := rows.Scan(
			&entry.ID, &entry.CategoryID, &entry.SeasonID, &entry.EntityType,
			&entry.EntityID, &entry.EntityName, &entry.Position, &entry.Score,
			&entry.PreviousPosition, &entry.PositionChange, &entry.Stats,
			&entry.Breakdown, &entry.LastUpdated, &entry.CreatedAt, &categoryName, &categoryType,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando entrada de ranking: %w", err)
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

// GetRankingsByScoreRange obtiene rankings por rango de puntuación
func (r *RankingRepository) GetRankingsByScoreRange(minScore, maxScore int, limit int) ([]models.RankingEntry, error) {
	query := `
		SELECT re.id, re.category_id, re.season_id, re.entity_type, re.entity_id,
		       re.entity_name, re.position, re.score, re.previous_position, re.position_change,
		       re.stats, re.breakdown, re.last_updated, re.created_at,
		       rc.name as category_name, rc.type as category_type
		FROM ranking_entries re
		JOIN ranking_categories rc ON re.category_id = rc.id
		WHERE re.score >= $1 AND re.score <= $2
		ORDER BY re.score DESC
		LIMIT $3
	`

	rows, err := r.db.Query(query, minScore, maxScore, limit)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo rankings por rango de puntuación: %w", err)
	}
	defer rows.Close()

	var entries []models.RankingEntry
	for rows.Next() {
		var entry models.RankingEntry
		var categoryName, categoryType string
		err := rows.Scan(
			&entry.ID, &entry.CategoryID, &entry.SeasonID, &entry.EntityType,
			&entry.EntityID, &entry.EntityName, &entry.Position, &entry.Score,
			&entry.PreviousPosition, &entry.PositionChange, &entry.Stats,
			&entry.Breakdown, &entry.LastUpdated, &entry.CreatedAt, &categoryName, &categoryType,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando entrada de ranking: %w", err)
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

// GetRankingsDashboard obtiene el dashboard completo de rankings
func (r *RankingRepository) GetRankingsDashboard(playerID int) (*models.RankingDashboard, error) {
	// Obtener categorías activas
	categories, err := r.GetRankingCategories(true)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo categorías: %w", err)
	}

	// Obtener temporada activa
	seasons, err := r.GetRankingSeasons("active")
	if err != nil {
		return nil, fmt.Errorf("error obteniendo temporada activa: %w", err)
	}

	var activeSeason *models.RankingSeason
	if len(seasons) > 0 {
		activeSeason = &seasons[0]
	}

	// Obtener top rankings (primeros 10 de cada categoría)
	var topRankings []models.RankingEntry
	for _, category := range categories {
		var seasonID *int
		if activeSeason != nil {
			seasonID = &activeSeason.ID
		}
		entries, err := r.GetRankingEntries(category.ID, seasonID, 10)
		if err != nil {
			r.logger.Warn("Error obteniendo top rankings para categoría",
				zap.Int("category_id", category.ID), zap.Error(err))
			continue
		}
		topRankings = append(topRankings, entries...)
	}

	// Obtener rankings del jugador
	var playerRankings []models.RankingEntry
	for _, category := range categories {
		var seasonID *int
		if activeSeason != nil {
			seasonID = &activeSeason.ID
		}
		entries, err := r.GetRankingEntries(category.ID, seasonID, 100)
		if err != nil {
			r.logger.Warn("Error obteniendo rankings del jugador para categoría",
				zap.Int("category_id", category.ID), zap.Error(err))
			continue
		}

		// Filtrar solo las entradas del jugador
		for _, entry := range entries {
			if entry.EntityType == "player" && entry.EntityID == playerID {
				playerRankings = append(playerRankings, entry)
				break
			}
		}
	}

	// Obtener estadísticas resumidas
	statistics, err := r.GetStatisticsSummary()
	if err != nil {
		return nil, fmt.Errorf("error obteniendo estadísticas: %w", err)
	}

	dashboard := &models.RankingDashboard{
		Categories:     categories,
		ActiveSeason:   activeSeason,
		TopRankings:    topRankings,
		PlayerRankings: playerRankings,
		Statistics:     *statistics,
		LastUpdated:    time.Now(),
	}

	return dashboard, nil
}
