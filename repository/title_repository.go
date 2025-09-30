package repository

import (
	"database/sql"
	"fmt"
	"server-backend/models"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type TitleRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewTitleRepository(db *sql.DB, logger *zap.Logger) *TitleRepository {
	return &TitleRepository{
		db:     db,
		logger: logger,
	}
}

// ==================== CATEGORÍAS DE TÍTULOS ====================

func (r *TitleRepository) GetTitleCategories(activeOnly bool) ([]models.TitleCategory, error) {
	query := `
		SELECT id, name, description, icon, color, background_color,
		       display_order, is_public, show_in_dashboard, total_titles,
		       unlocked_count, completion_rate, is_active, created_at, updated_at
		FROM title_categories
	`

	if activeOnly {
		query += " WHERE is_active = true"
	}

	query += " ORDER BY display_order ASC, name ASC"

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo categorías de títulos: %w", err)
	}
	defer rows.Close()

	var categories []models.TitleCategory
	for rows.Next() {
		var category models.TitleCategory
		err := rows.Scan(
			&category.ID, &category.Name, &category.Description, &category.Icon,
			&category.Color, &category.BackgroundColor, &category.DisplayOrder,
			&category.IsPublic, &category.ShowInDashboard, &category.TotalTitles,
			&category.UnlockedCount, &category.CompletionRate, &category.IsActive,
			&category.CreatedAt, &category.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando categoría: %w", err)
		}
		categories = append(categories, category)
	}

	return categories, nil
}

func (r *TitleRepository) GetTitleCategory(categoryID uuid.UUID) (*models.TitleCategory, error) {
	query := `
		SELECT id, name, description, icon, color, background_color,
		       display_order, is_public, show_in_dashboard, total_titles,
		       unlocked_count, completion_rate, is_active, created_at, updated_at
		FROM title_categories
		WHERE id = $1
	`

	var category models.TitleCategory
	err := r.db.QueryRow(query, categoryID).Scan(
		&category.ID, &category.Name, &category.Description, &category.Icon,
		&category.Color, &category.BackgroundColor, &category.DisplayOrder,
		&category.IsPublic, &category.ShowInDashboard, &category.TotalTitles,
		&category.UnlockedCount, &category.CompletionRate, &category.IsActive,
		&category.CreatedAt, &category.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("categoría no encontrada")
		}
		return nil, fmt.Errorf("error obteniendo categoría: %w", err)
	}

	return &category, nil
}

func (r *TitleRepository) CreateTitleCategory(category *models.TitleCategory) error {
	query := `
		INSERT INTO title_categories (
			id, name, description, icon, color, background_color,
			display_order, is_public, show_in_dashboard, total_titles,
			unlocked_count, completion_rate, is_active, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
		)
	`

	now := time.Now()
	if category.ID == uuid.Nil {
		category.ID = uuid.New()
	}

	_, err := r.db.Exec(query,
		category.ID, category.Name, category.Description, category.Icon, category.Color,
		category.BackgroundColor, category.DisplayOrder, category.IsPublic,
		category.ShowInDashboard, category.TotalTitles, category.UnlockedCount,
		category.CompletionRate, category.IsActive, now, now,
	)

	if err != nil {
		return fmt.Errorf("error creando categoría: %w", err)
	}

	return nil
}

// ==================== TÍTULOS ====================

func (r *TitleRepository) CreateTitle(title *models.Title) error {
	query := `
		INSERT INTO titles (
			id, category_id, name, description, long_description, story_text,
			icon, color, background_color, border_color, rarity, title_type,
			title_format, display_format, level_required, prestige_required,
			alliance_required, prerequisites, unlock_conditions, effects,
			bonuses, special_abilities, prestige_value, reputation_bonus,
			social_status, max_owners, time_limit, is_exclusive, is_temporary,
			status, unlock_date, retire_date, total_unlocked, current_owners,
			unlock_rate, is_repeatable, repeat_interval, next_unlock_date,
			is_active, is_hidden, is_featured, display_order, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16,
			$17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30,
			$31, $32, $33, $34, $35, $36, $37, $38, $39, $40, $41, $42, $43
		)
	`

	now := time.Now()
	if title.ID == uuid.Nil {
		title.ID = uuid.New()
	}

	_, err := r.db.Exec(query,
		title.ID, title.CategoryID, title.Name, title.Description, title.LongDescription,
		title.StoryText, title.Icon, title.Color, title.BackgroundColor, title.BorderColor,
		title.Rarity, title.TitleType, title.TitleFormat, title.DisplayFormat,
		title.LevelRequired, title.PrestigeRequired, title.AllianceRequired,
		title.Prerequisites, title.UnlockConditions, title.Effects, title.Bonuses,
		title.SpecialAbilities, title.PrestigeValue, title.ReputationBonus,
		title.SocialStatus, title.MaxOwners, title.TimeLimit, title.IsExclusive,
		title.IsTemporary, title.Status, title.UnlockDate, title.RetireDate,
		title.TotalUnlocked, title.CurrentOwners, title.UnlockRate, title.IsRepeatable,
		title.RepeatInterval, title.NextUnlockDate, title.IsActive, title.IsHidden,
		title.IsFeatured, title.DisplayOrder, now, now,
	)

	if err != nil {
		return fmt.Errorf("error creating title: %v", err)
	}

	return nil
}

func (r *TitleRepository) GetAllTitles() ([]models.Title, error) {
	query := `SELECT * FROM titles WHERE is_active = true ORDER BY display_order, name`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error getting titles: %v", err)
	}
	defer rows.Close()
	var titles []models.Title
	for rows.Next() {
		var t models.Title
		err := rows.Scan(&t.ID, &t.CategoryID, &t.Name, &t.Description, &t.LongDescription, &t.StoryText, &t.Icon, &t.Color, &t.BackgroundColor, &t.BorderColor, &t.Rarity, &t.TitleType, &t.TitleFormat, &t.DisplayFormat, &t.LevelRequired, &t.PrestigeRequired, &t.AllianceRequired, &t.Prerequisites, &t.UnlockConditions, &t.Effects, &t.Bonuses, &t.SpecialAbilities, &t.PrestigeValue, &t.ReputationBonus, &t.SocialStatus, &t.MaxOwners, &t.TimeLimit, &t.IsExclusive, &t.IsTemporary, &t.Status, &t.UnlockDate, &t.RetireDate, &t.TotalUnlocked, &t.CurrentOwners, &t.UnlockRate, &t.IsRepeatable, &t.RepeatInterval, &t.NextUnlockDate, &t.IsActive, &t.IsHidden, &t.IsFeatured, &t.DisplayOrder, &t.CreatedAt, &t.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("error scanning title: %v", err)
		}
		titles = append(titles, t)
	}
	return titles, nil
}

// ==================== PRESTIGIO ====================

func (r *TitleRepository) GetPlayerPrestige(playerID uuid.UUID) (*models.PlayerPrestige, error) {
	query := `SELECT * FROM player_prestige WHERE player_id = $1`
	var p models.PlayerPrestige
	err := r.db.QueryRow(query, playerID).Scan(&p.ID, &p.PlayerID, &p.CurrentPrestige, &p.TotalPrestige, &p.PrestigeLevel, &p.PrestigeToNext, &p.ProgressPercent, &p.TitlesUnlocked, &p.TitlesEquipped, &p.AchievementsCompleted, &p.PrestigeHistory, &p.LastGainDate, &p.LargestGain, &p.GlobalRank, &p.CategoryRank, &p.AllianceRank, &p.FirstPrestige, &p.LastUpdated, &p.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("error getting player prestige: %v", err)
	}
	return &p, nil
}

// GetPrestigeLevel obtiene un nivel de prestigio específico
func (r *TitleRepository) GetPrestigeLevel(level int) (*models.PrestigeLevel, error) {
	query := `
		SELECT id, name, description, icon, color, background_color,
		       level, prestige_required, experience_multiplier,
		       bonuses, special_effects, unlock_features, is_active, created_at
		FROM prestige_levels
		WHERE level = $1 AND is_active = true
	`

	var prestigeLevel models.PrestigeLevel
	err := r.db.QueryRow(query, level).Scan(
		&prestigeLevel.ID, &prestigeLevel.Name, &prestigeLevel.Description,
		&prestigeLevel.Icon, &prestigeLevel.Color, &prestigeLevel.BackgroundColor,
		&prestigeLevel.Level, &prestigeLevel.PrestigeRequired, &prestigeLevel.ExperienceMultiplier,
		&prestigeLevel.Bonuses, &prestigeLevel.SpecialEffects, &prestigeLevel.UnlockFeatures,
		&prestigeLevel.IsActive, &prestigeLevel.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("nivel de prestigio no encontrado")
		}
		return nil, fmt.Errorf("error obteniendo nivel de prestigio: %w", err)
	}

	return &prestigeLevel, nil
}

// GetPlayerEquippedTitles obtiene los títulos equipados de un jugador
func (r *TitleRepository) GetPlayerEquippedTitles(playerID uuid.UUID) ([]models.PlayerTitle, error) {
	query := `
		SELECT pt.id, pt.player_id, pt.title_id, pt.status, pt.is_unlocked,
		       pt.is_equipped, pt.is_favorite, pt.unlock_date, pt.equipped_date,
		       pt.unlock_method, pt.unlock_data, pt.expiry_date, pt.days_remaining,
		       pt.is_permanent, pt.times_equipped, pt.total_time_equipped,
		       pt.last_equipped, pt.progress, pt.max_progress, pt.level,
		       pt.max_level, pt.prestige_level, pt.rewards_claimed,
		       pt.rewards_data, pt.points_earned, pt.created_at, pt.updated_at
		FROM player_titles pt
		WHERE pt.player_id = $1 AND pt.is_equipped = true
		ORDER BY pt.equipped_date DESC
	`

	rows, err := r.db.Query(query, playerID)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo títulos equipados: %w", err)
	}
	defer rows.Close()

	var titles []models.PlayerTitle
	for rows.Next() {
		var title models.PlayerTitle
		err := rows.Scan(
			&title.ID, &title.PlayerID, &title.TitleID, &title.Status, &title.IsUnlocked,
			&title.IsEquipped, &title.IsFavorite, &title.UnlockDate, &title.EquippedDate,
			&title.UnlockMethod, &title.UnlockData, &title.ExpiryDate, &title.DaysRemaining,
			&title.IsPermanent, &title.TimesEquipped, &title.TotalTimeEquipped,
			&title.LastEquipped, &title.Progress, &title.MaxProgress, &title.Level,
			&title.MaxLevel, &title.PrestigeLevel, &title.RewardsClaimed,
			&title.RewardsData, &title.PointsEarned, &title.CreatedAt, &title.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando título equipado: %w", err)
		}
		titles = append(titles, title)
	}

	return titles, nil
}

// GetPlayerRecentUnlocks obtiene los títulos recientemente desbloqueados de un jugador
func (r *TitleRepository) GetPlayerRecentUnlocks(playerID uuid.UUID, limit int) ([]models.PlayerTitle, error) {
	query := `
		SELECT pt.id, pt.player_id, pt.title_id, pt.status, pt.is_unlocked,
		       pt.is_equipped, pt.is_favorite, pt.unlock_date, pt.equipped_date,
		       pt.unlock_method, pt.unlock_data, pt.expiry_date, pt.days_remaining,
		       pt.is_permanent, pt.times_equipped, pt.total_time_equipped,
		       pt.last_equipped, pt.progress, pt.max_progress, pt.level,
		       pt.max_level, pt.prestige_level, pt.rewards_claimed,
		       pt.rewards_data, pt.points_earned, pt.created_at, pt.updated_at
		FROM player_titles pt
		WHERE pt.player_id = $1 AND pt.is_unlocked = true
		ORDER BY pt.unlock_date DESC
		LIMIT $2
	`

	rows, err := r.db.Query(query, playerID, limit)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo títulos recientes: %w", err)
	}
	defer rows.Close()

	var titles []models.PlayerTitle
	for rows.Next() {
		var title models.PlayerTitle
		err := rows.Scan(
			&title.ID, &title.PlayerID, &title.TitleID, &title.Status, &title.IsUnlocked,
			&title.IsEquipped, &title.IsFavorite, &title.UnlockDate, &title.EquippedDate,
			&title.UnlockMethod, &title.UnlockData, &title.ExpiryDate, &title.DaysRemaining,
			&title.IsPermanent, &title.TimesEquipped, &title.TotalTimeEquipped,
			&title.LastEquipped, &title.Progress, &title.MaxProgress, &title.Level,
			&title.MaxLevel, &title.PrestigeLevel, &title.RewardsClaimed,
			&title.RewardsData, &title.PointsEarned, &title.CreatedAt, &title.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando título reciente: %w", err)
		}
		titles = append(titles, title)
	}

	return titles, nil
}

// GetAvailableTitles obtiene los títulos disponibles para un jugador
func (r *TitleRepository) GetAvailableTitles(playerID uuid.UUID) ([]models.Title, error) {
	query := `
		SELECT t.id, t.category_id, t.name, t.description, t.long_description,
		       t.story_text, t.icon, t.color, t.background_color, t.border_color,
		       t.rarity, t.title_type, t.title_format, t.display_format,
		       t.level_required, t.prestige_required, t.alliance_required,
		       t.prerequisites, t.unlock_conditions, t.effects, t.bonuses,
		       t.special_abilities, t.prestige_value, t.reputation_bonus,
		       t.social_status, t.max_owners, t.time_limit, t.is_exclusive,
		       t.is_temporary, t.status, t.unlock_date, t.retire_date,
		       t.total_unlocked, t.current_owners, t.unlock_rate, t.is_repeatable,
		       t.repeat_interval, t.next_unlock_date, t.is_active, t.is_hidden,
		       t.is_featured, t.display_order, t.created_at, t.updated_at
		FROM titles t
		WHERE t.is_active = true AND t.status = 'available'
		  AND t.id NOT IN (
		    SELECT pt.title_id FROM player_titles pt WHERE pt.player_id = $1
		  )
		ORDER BY t.display_order ASC, t.name ASC
	`

	rows, err := r.db.Query(query, playerID)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo títulos disponibles: %w", err)
	}
	defer rows.Close()

	var titles []models.Title
	for rows.Next() {
		var title models.Title
		err := rows.Scan(
			&title.ID, &title.CategoryID, &title.Name, &title.Description, &title.LongDescription,
			&title.StoryText, &title.Icon, &title.Color, &title.BackgroundColor, &title.BorderColor,
			&title.Rarity, &title.TitleType, &title.TitleFormat, &title.DisplayFormat,
			&title.LevelRequired, &title.PrestigeRequired, &title.AllianceRequired,
			&title.Prerequisites, &title.UnlockConditions, &title.Effects, &title.Bonuses,
			&title.SpecialAbilities, &title.PrestigeValue, &title.ReputationBonus,
			&title.SocialStatus, &title.MaxOwners, &title.TimeLimit, &title.IsExclusive,
			&title.IsTemporary, &title.Status, &title.UnlockDate, &title.RetireDate,
			&title.TotalUnlocked, &title.CurrentOwners, &title.UnlockRate, &title.IsRepeatable,
			&title.RepeatInterval, &title.NextUnlockDate, &title.IsActive, &title.IsHidden,
			&title.IsFeatured, &title.DisplayOrder, &title.CreatedAt, &title.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando título disponible: %w", err)
		}
		titles = append(titles, title)
	}

	return titles, nil
}

// GetPlayerTitleRankings obtiene los rankings de títulos de un jugador
func (r *TitleRepository) GetPlayerTitleRankings(playerID uuid.UUID) (*models.TitleRanking, error) {
	query := `
		SELECT id, player_id, prestige_rank, titles_rank, achievements_rank,
		       overall_rank, prestige_score, titles_score, achievements_score,
		       overall_score, rare_titles, epic_titles, legendary_titles,
		       mythic_titles, divine_titles, last_updated, created_at
		FROM title_rankings
		WHERE player_id = $1
	`

	var ranking models.TitleRanking
	err := r.db.QueryRow(query, playerID).Scan(
		&ranking.ID, &ranking.PlayerID, &ranking.PrestigeRank, &ranking.TitlesRank,
		&ranking.AchievementsRank, &ranking.OverallRank, &ranking.PrestigeScore,
		&ranking.TitlesScore, &ranking.AchievementsScore, &ranking.OverallScore,
		&ranking.RareTitles, &ranking.EpicTitles, &ranking.LegendaryTitles,
		&ranking.MythicTitles, &ranking.DivineTitles, &ranking.LastUpdated, &ranking.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			// Crear un ranking por defecto si no existe
			ranking = models.TitleRanking{
				ID:                uuid.New(),
				PlayerID:          playerID,
				PrestigeRank:      0,
				TitlesRank:        0,
				AchievementsRank:  0,
				OverallRank:       0,
				PrestigeScore:     0,
				TitlesScore:       0,
				AchievementsScore: 0,
				OverallScore:      0,
				RareTitles:        0,
				EpicTitles:        0,
				LegendaryTitles:   0,
				MythicTitles:      0,
				DivineTitles:      0,
				LastUpdated:       time.Now(),
				CreatedAt:         time.Now(),
			}
			return &ranking, nil
		}
		return nil, fmt.Errorf("error obteniendo rankings: %w", err)
	}

	return &ranking, nil
}

// GetPlayerTitleStatistics obtiene las estadísticas de títulos de un jugador
func (r *TitleRepository) GetPlayerTitleStatistics(playerID uuid.UUID) (*models.TitleStatistics, error) {
	query := `
		SELECT id, player_id, total_titles_unlocked, total_titles_equipped,
		       total_prestige_gained, total_achievements_completed, category_stats,
		       type_stats, rarity_stats, last_title_unlocked, last_title_equipped,
		       last_prestige_gain, longest_equipped_title, most_prestigious_title,
		       rarest_title, first_title_unlocked, last_updated, created_at
		FROM title_statistics
		WHERE player_id = $1
	`

	var stats models.TitleStatistics
	err := r.db.QueryRow(query, playerID).Scan(
		&stats.ID, &stats.PlayerID, &stats.TotalTitlesUnlocked, &stats.TotalTitlesEquipped,
		&stats.TotalPrestigeGained, &stats.TotalAchievementsCompleted, &stats.CategoryStats,
		&stats.TypeStats, &stats.RarityStats, &stats.LastTitleUnlocked, &stats.LastTitleEquipped,
		&stats.LastPrestigeGain, &stats.LongestEquippedTitle, &stats.MostPrestigiousTitle,
		&stats.RarestTitle, &stats.FirstTitleUnlocked, &stats.LastUpdated, &stats.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			// Crear estadísticas por defecto si no existen
			stats = models.TitleStatistics{
				ID:                         uuid.New(),
				PlayerID:                   playerID,
				TotalTitlesUnlocked:        0,
				TotalTitlesEquipped:        0,
				TotalPrestigeGained:        0,
				TotalAchievementsCompleted: 0,
				FirstTitleUnlocked:         time.Now(),
				LastUpdated:                time.Now(),
				CreatedAt:                  time.Now(),
			}
			return &stats, nil
		}
		return nil, fmt.Errorf("error obteniendo estadísticas: %w", err)
	}

	return &stats, nil
}

// GetPlayerTitleNotifications obtiene las notificaciones de títulos de un jugador
func (r *TitleRepository) GetPlayerTitleNotifications(playerID uuid.UUID) ([]models.TitleNotification, error) {
	query := `
		SELECT id, player_id, title_id, type, title, message, data,
		       is_read, is_dismissed, created_at, read_at, dismissed_at
		FROM title_notifications
		WHERE player_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, playerID)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo notificaciones: %w", err)
	}
	defer rows.Close()

	var notifications []models.TitleNotification
	for rows.Next() {
		var notification models.TitleNotification
		err := rows.Scan(
			&notification.ID, &notification.PlayerID, &notification.TitleID,
			&notification.Type, &notification.Title, &notification.Message, &notification.Data,
			&notification.IsRead, &notification.IsDismissed, &notification.CreatedAt,
			&notification.ReadAt, &notification.DismissedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando notificación: %w", err)
		}
		notifications = append(notifications, notification)
	}

	return notifications, nil
}

// GetActiveTitleEvents obtiene los eventos de títulos activos
func (r *TitleRepository) GetActiveTitleEvents() ([]models.TitleEvent, error) {
	query := `
		SELECT id, name, description, icon, event_type, start_date, end_date,
		       effects, prestige_multiplier, unlock_chance, total_participants,
		       active_participants, status, is_active, created_at
		FROM title_events
		WHERE is_active = true AND status IN ('upcoming', 'active')
		ORDER BY start_date ASC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo eventos activos: %w", err)
	}
	defer rows.Close()

	var events []models.TitleEvent
	for rows.Next() {
		var event models.TitleEvent
		err := rows.Scan(
			&event.ID, &event.Name, &event.Description, &event.Icon, &event.EventType,
			&event.StartDate, &event.EndDate, &event.Effects, &event.PrestigeMultiplier,
			&event.UnlockChance, &event.TotalParticipants, &event.ActiveParticipants,
			&event.Status, &event.IsActive, &event.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando evento: %w", err)
		}
		events = append(events, event)
	}

	return events, nil
}

// CreateTitleNotification crea una nueva notificación de título
func (r *TitleRepository) CreateTitleNotification(notification *models.TitleNotification) error {
	query := `
		INSERT INTO title_notifications (
			id, player_id, title_id, type, title, message, data,
			is_read, is_dismissed, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		)
	`

	_, err := r.db.Exec(query,
		notification.ID, notification.PlayerID, notification.TitleID,
		notification.Type, notification.Title, notification.Message, notification.Data,
		notification.IsRead, notification.IsDismissed, notification.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("error creando notificación: %w", err)
	}

	return nil
}

// GetTitleOwnersCount obtiene el número de propietarios de un título
func (r *TitleRepository) GetTitleOwnersCount(titleID uuid.UUID) (int, error) {
	query := `
		SELECT COUNT(*) FROM player_titles
		WHERE title_id = $1 AND is_unlocked = true
	`

	var count int
	err := r.db.QueryRow(query, titleID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("error contando propietarios: %w", err)
	}

	return count, nil
}

// ==================== PLAYER TITLE ====================

func (r *TitleRepository) GetPlayerTitles(playerID uuid.UUID, categoryID *uuid.UUID, unlockedOnly bool) ([]models.PlayerTitle, error) {
	query := `
		SELECT player_id, title_id, is_unlocked, is_equipped,
		       unlock_date, equipped_date, prestige_level,
		       rewards_claimed, rewards_data, points_earned,
		       created_at, updated_at
		FROM player_titles
		WHERE player_id = $1
	`

	args := []interface{}{playerID}
	argCount := 2

	if categoryID != nil {
		query += fmt.Sprintf(" AND title_id IN (SELECT id FROM titles WHERE category_id = $%d)", argCount)
		args = append(args, *categoryID)
		argCount++
	}

	if unlockedOnly {
		query += fmt.Sprintf(" AND is_unlocked = true", argCount)
		argCount++
	}

	query += " ORDER BY created_at DESC"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo títulos del jugador: %w", err)
	}
	defer rows.Close()

	var titles []models.PlayerTitle
	for rows.Next() {
		var title models.PlayerTitle
		err := rows.Scan(
			&title.PlayerID, &title.TitleID, &title.IsUnlocked,
			&title.IsEquipped, &title.UnlockDate, &title.EquippedDate,
			&title.PrestigeLevel, &title.RewardsClaimed, &title.RewardsData,
			&title.PointsEarned, &title.CreatedAt, &title.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando título del jugador: %w", err)
		}
		titles = append(titles, title)
	}

	return titles, nil
}

// ==================== RANKING ====================

func (r *TitleRepository) GetTitleRanking() ([]models.TitleRanking, error) {
	query := `SELECT * FROM title_ranking ORDER BY overall_rank ASC LIMIT 100`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error getting title ranking: %v", err)
	}
	defer rows.Close()
	var rankings []models.TitleRanking
	for rows.Next() {
		var tr models.TitleRanking
		err := rows.Scan(&tr.ID, &tr.PlayerID, &tr.PrestigeRank, &tr.TitlesRank, &tr.AchievementsRank, &tr.OverallRank, &tr.PrestigeScore, &tr.TitlesScore, &tr.AchievementsScore, &tr.OverallScore, &tr.RareTitles, &tr.EpicTitles, &tr.LegendaryTitles, &tr.MythicTitles, &tr.DivineTitles, &tr.LastUpdated, &tr.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("error scanning title ranking: %v", err)
		}
		rankings = append(rankings, tr)
	}
	return rankings, nil
}

// ==================== NOTIFICACIONES ====================

func (r *TitleRepository) GetTitleNotification(notificationID uuid.UUID) (*models.TitleNotification, error) {
	query := `
		SELECT id, player_id, title_id, type, title, message, data,
		       is_read, is_dismissed, created_at, read_at, dismissed_at
		FROM title_notifications
		WHERE id = $1
	`

	var notification models.TitleNotification
	err := r.db.QueryRow(query, notificationID).Scan(
		&notification.ID, &notification.PlayerID, &notification.TitleID,
		&notification.Type, &notification.Title, &notification.Message, &notification.Data,
		&notification.IsRead, &notification.IsDismissed, &notification.CreatedAt,
		&notification.ReadAt, &notification.DismissedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("notificación no encontrada")
		}
		return nil, fmt.Errorf("error obteniendo notificación: %w", err)
	}

	return &notification, nil
}

// ==================== DASHBOARD ====================

func (r *TitleRepository) GetPrestigeDashboard(playerID uuid.UUID) (*models.PrestigeDashboard, error) {
	dashboard := &models.PrestigeDashboard{
		LastUpdated: time.Now(),
	}
	// Obtener prestigio
	prestige, err := r.GetPlayerPrestige(playerID)
	if err == nil {
		dashboard.PlayerPrestige = prestige
	}
	// Obtener títulos equipados
	titles, err := r.GetPlayerTitles(playerID, nil, false)
	if err == nil {
		dashboard.EquippedTitles = titles
	}
	// Obtener ranking
	rankings, err := r.GetTitleRanking()
	if err == nil && len(rankings) > 0 {
		for _, rnk := range rankings {
			if rnk.PlayerID == playerID {
				dashboard.Rankings = &rnk
				break
			}
		}
	}
	// Obtener notificaciones
	notifs, err := r.GetPlayerTitleNotifications(playerID)
	if err == nil {
		dashboard.Notifications = notifs
	}
	return dashboard, nil
}

// GetTitles obtiene todos los títulos con filtros opcionales
func (r *TitleRepository) GetTitles(filters map[string]interface{}) ([]models.Title, error) {
	query := `
		SELECT id, category_id, name, description, long_description, story_text,
		       icon, color, background_color, border_color, rarity, title_type,
		       title_format, display_format, level_required, prestige_required,
		       alliance_required, prerequisites, unlock_conditions, effects,
		       bonuses, special_abilities, prestige_value, reputation_bonus,
		       social_status, max_owners, time_limit, is_exclusive, is_temporary,
		       status, unlock_date, retire_date, total_unlocked, current_owners,
		       unlock_rate, is_repeatable, repeat_interval, next_unlock_date,
		       is_active, is_hidden, is_featured, display_order, created_at, updated_at
		FROM titles
		WHERE is_active = true
	`

	args := []interface{}{}
	argCount := 1

	// Aplicar filtros
	if categoryID, ok := filters["category_id"].(uuid.UUID); ok {
		query += fmt.Sprintf(" AND category_id = $%d", argCount)
		args = append(args, categoryID)
		argCount++
	}

	if rarity, ok := filters["rarity"].(string); ok {
		query += fmt.Sprintf(" AND rarity = $%d", argCount)
		args = append(args, rarity)
		argCount++
	}

	if titleType, ok := filters["title_type"].(string); ok {
		query += fmt.Sprintf(" AND title_type = $%d", argCount)
		args = append(args, titleType)
		argCount++
	}

	query += " ORDER BY display_order ASC, name ASC"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo títulos: %w", err)
	}
	defer rows.Close()

	var titles []models.Title
	for rows.Next() {
		var title models.Title
		err := rows.Scan(
			&title.ID, &title.CategoryID, &title.Name, &title.Description, &title.LongDescription,
			&title.StoryText, &title.Icon, &title.Color, &title.BackgroundColor, &title.BorderColor,
			&title.Rarity, &title.TitleType, &title.TitleFormat, &title.DisplayFormat,
			&title.LevelRequired, &title.PrestigeRequired, &title.AllianceRequired,
			&title.Prerequisites, &title.UnlockConditions, &title.Effects, &title.Bonuses,
			&title.SpecialAbilities, &title.PrestigeValue, &title.ReputationBonus,
			&title.SocialStatus, &title.MaxOwners, &title.TimeLimit, &title.IsExclusive,
			&title.IsTemporary, &title.Status, &title.UnlockDate, &title.RetireDate,
			&title.TotalUnlocked, &title.CurrentOwners, &title.UnlockRate, &title.IsRepeatable,
			&title.RepeatInterval, &title.NextUnlockDate, &title.IsActive, &title.IsHidden,
			&title.IsFeatured, &title.DisplayOrder, &title.CreatedAt, &title.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando título: %w", err)
		}
		titles = append(titles, title)
	}

	return titles, nil
}

// GetTitle obtiene un título específico por ID
func (r *TitleRepository) GetTitle(titleID uuid.UUID) (*models.Title, error) {
	query := `
		SELECT id, category_id, name, description, long_description, story_text,
		       icon, color, background_color, border_color, rarity, title_type,
		       title_format, display_format, level_required, prestige_required,
		       alliance_required, prerequisites, unlock_conditions, effects,
		       bonuses, special_abilities, prestige_value, reputation_bonus,
		       social_status, max_owners, time_limit, is_exclusive, is_temporary,
		       status, unlock_date, retire_date, total_unlocked, current_owners,
		       unlock_rate, is_repeatable, repeat_interval, next_unlock_date,
		       is_active, is_hidden, is_featured, display_order, created_at, updated_at
		FROM titles
		WHERE id = $1
	`

	var title models.Title
	err := r.db.QueryRow(query, titleID).Scan(
		&title.ID, &title.CategoryID, &title.Name, &title.Description, &title.LongDescription,
		&title.StoryText, &title.Icon, &title.Color, &title.BackgroundColor, &title.BorderColor,
		&title.Rarity, &title.TitleType, &title.TitleFormat, &title.DisplayFormat,
		&title.LevelRequired, &title.PrestigeRequired, &title.AllianceRequired,
		&title.Prerequisites, &title.UnlockConditions, &title.Effects, &title.Bonuses,
		&title.SpecialAbilities, &title.PrestigeValue, &title.ReputationBonus,
		&title.SocialStatus, &title.MaxOwners, &title.TimeLimit, &title.IsExclusive,
		&title.IsTemporary, &title.Status, &title.UnlockDate, &title.RetireDate,
		&title.TotalUnlocked, &title.CurrentOwners, &title.UnlockRate, &title.IsRepeatable,
		&title.RepeatInterval, &title.NextUnlockDate, &title.IsActive, &title.IsHidden,
		&title.IsFeatured, &title.DisplayOrder, &title.CreatedAt, &title.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("título no encontrado")
		}
		return nil, fmt.Errorf("error obteniendo título: %w", err)
	}

	return &title, nil
}

// GetTitleLeaderboard obtiene el leaderboard de títulos
func (r *TitleRepository) GetTitleLeaderboard(limit int) ([]models.TitleLeaderboard, error) {
	// Por ahora, retornamos una lista vacía ya que TitleLeaderboard es una estructura de configuración
	// no de datos de ranking. El ranking real se obtiene a través de otras consultas
	var leaderboard []models.TitleLeaderboard
	return leaderboard, nil
}

// GetTitleStatistics obtiene las estadísticas de títulos
func (r *TitleRepository) GetTitleStatistics() (*models.TitleStatistics, error) {
	// Obtener estadísticas generales
	statsQuery := `
		SELECT 
			COUNT(*) as total_titles,
			COUNT(CASE WHEN is_active = true THEN 1 END) as active_titles,
			COUNT(CASE WHEN rarity = 'common' THEN 1 END) as common_titles,
			COUNT(CASE WHEN rarity = 'rare' THEN 1 END) as rare_titles,
			COUNT(CASE WHEN rarity = 'epic' THEN 1 END) as epic_titles,
			COUNT(CASE WHEN rarity = 'legendary' THEN 1 END) as legendary_titles,
			COUNT(CASE WHEN rarity = 'mythic' THEN 1 END) as mythic_titles,
			COUNT(CASE WHEN rarity = 'divine' THEN 1 END) as divine_titles
		FROM titles
	`

	var totalTitles, activeTitles, commonTitles, rareTitles, epicTitles, legendaryTitles, mythicTitles, divineTitles int
	err := r.db.QueryRow(statsQuery).Scan(
		&totalTitles, &activeTitles, &commonTitles,
		&rareTitles, &epicTitles, &legendaryTitles,
		&mythicTitles, &divineTitles,
	)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo estadísticas: %w", err)
	}

	// Crear estadísticas con datos básicos
	stats := &models.TitleStatistics{
		ID:                         uuid.New(),
		PlayerID:                   uuid.Nil, // Estadísticas globales
		TotalTitlesUnlocked:        totalTitles,
		TotalTitlesEquipped:        0,    // Se calculará por separado
		TotalPrestigeGained:        0,    // Se calculará por separado
		TotalAchievementsCompleted: 0,    // Se calculará por separado
		CategoryStats:              "{}", // JSON vacío por ahora
		TypeStats:                  "{}", // JSON vacío por ahora
		RarityStats:                "{}", // JSON vacío por ahora
		LastTitleUnlocked:          nil,
		LastTitleEquipped:          nil,
		LastPrestigeGain:           nil,
		LongestEquippedTitle:       "",
		MostPrestigiousTitle:       "",
		RarestTitle:                "",
		FirstTitleUnlocked:         time.Now(),
		LastUpdated:                time.Now(),
		CreatedAt:                  time.Now(),
	}

	return stats, nil
}

// GrantTitle otorga un título a un jugador
func (r *TitleRepository) GrantTitle(playerID, titleID uuid.UUID, reason string) error {
	query := `
		INSERT INTO player_titles (
			id, player_id, title_id, unlocked_at, is_equipped, unlock_reason,
			equipped_at, unequipped_at, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		)
	`

	now := time.Now()
	playerTitleID := uuid.New()

	_, err := r.db.Exec(query,
		playerTitleID, playerID, titleID, now, false, reason,
		nil, nil, now, now,
	)

	if err != nil {
		return fmt.Errorf("error otorgando título: %w", err)
	}

	return nil
}

// EquipTitle equipa un título para un jugador
func (r *TitleRepository) EquipTitle(playerID, titleID uuid.UUID) error {
	// Primero desequipar todos los títulos del jugador
	unequipQuery := `
		UPDATE player_titles 
		SET is_equipped = false, unequipped_at = $1, updated_at = $1
		WHERE player_id = $2 AND is_equipped = true
	`

	now := time.Now()
	_, err := r.db.Exec(unequipQuery, now, playerID)
	if err != nil {
		return fmt.Errorf("error desequipando títulos anteriores: %w", err)
	}

	// Luego equipar el título específico
	equipQuery := `
		UPDATE player_titles 
		SET is_equipped = true, equipped_at = $1, updated_at = $1
		WHERE player_id = $2 AND title_id = $3
	`

	_, err = r.db.Exec(equipQuery, now, playerID, titleID)
	if err != nil {
		return fmt.Errorf("error equipando título: %w", err)
	}

	return nil
}

// UnequipTitle desequipa todos los títulos de un jugador
func (r *TitleRepository) UnequipTitle(playerID uuid.UUID) error {
	query := `
		UPDATE player_titles 
		SET is_equipped = false, unequipped_at = $1, updated_at = $1
		WHERE player_id = $2 AND is_equipped = true
	`

	now := time.Now()
	_, err := r.db.Exec(query, now, playerID)
	if err != nil {
		return fmt.Errorf("error desequipando títulos: %w", err)
	}

	return nil
}
