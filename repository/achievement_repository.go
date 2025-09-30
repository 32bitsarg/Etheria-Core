package repository

import (
	"database/sql"
	"fmt"
	"time"

	"server-backend/models"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type AchievementRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewAchievementRepository(db *sql.DB, logger *zap.Logger) *AchievementRepository {
	return &AchievementRepository{
		db:     db,
		logger: logger,
	}
}

// GetAchievementCategories obtiene todas las categorías de logros
func (r *AchievementRepository) GetAchievementCategories(activeOnly bool) ([]models.AchievementCategory, error) {
	query := `
		SELECT id, name, description, icon, color, background_color,
		       display_order, is_public, show_in_dashboard, total_achievements,
		       completed_count, completion_rate, is_active, created_at, updated_at
		FROM achievement_categories
	`

	if activeOnly {
		query += " WHERE is_active = true"
	}

	query += " ORDER BY display_order ASC, name ASC"

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo categorías de logros: %w", err)
	}
	defer rows.Close()

	var categories []models.AchievementCategory
	for rows.Next() {
		var category models.AchievementCategory
		err := rows.Scan(
			&category.ID, &category.Name, &category.Description, &category.Icon,
			&category.Color, &category.BackgroundColor, &category.DisplayOrder,
			&category.IsPublic, &category.ShowInDashboard, &category.TotalAchievements,
			&category.CompletedCount, &category.CompletionRate, &category.IsActive,
			&category.CreatedAt, &category.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando categoría: %w", err)
		}
		categories = append(categories, category)
	}

	return categories, nil
}

// GetAchievementCategory obtiene una categoría específica
func (r *AchievementRepository) GetAchievementCategory(categoryID uuid.UUID) (*models.AchievementCategory, error) {
	query := `
		SELECT id, name, description, icon, color, background_color,
		       display_order, is_public, show_in_dashboard, total_achievements,
		       completed_count, completion_rate, is_active, created_at, updated_at
		FROM achievement_categories
		WHERE id = $1
	`

	var category models.AchievementCategory
	err := r.db.QueryRow(query, categoryID).Scan(
		&category.ID, &category.Name, &category.Description, &category.Icon,
		&category.Color, &category.BackgroundColor, &category.DisplayOrder,
		&category.IsPublic, &category.ShowInDashboard, &category.TotalAchievements,
		&category.CompletedCount, &category.CompletionRate, &category.IsActive,
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

// CreateAchievementCategory crea una nueva categoría de logros
func (r *AchievementRepository) CreateAchievementCategory(category *models.AchievementCategory) error {
	query := `
		INSERT INTO achievement_categories (
			id, name, description, icon, color, background_color,
			display_order, is_public, show_in_dashboard, total_achievements,
			completed_count, completion_rate, is_active, created_at, updated_at
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
		category.ShowInDashboard, category.TotalAchievements, category.CompletedCount,
		category.CompletionRate, category.IsActive, now, now,
	)

	if err != nil {
		return fmt.Errorf("error creando categoría: %w", err)
	}

	return nil
}

// GetAchievements obtiene logros con filtros opcionales
func (r *AchievementRepository) GetAchievements(categoryID *uuid.UUID, activeOnly bool, includeHidden bool) ([]models.Achievement, error) {
	query := `
		SELECT id, category_id, name, description, long_description,
		       icon, color, background_color, rarity, progress_type,
		       target_value, current_value, progress_formula, tiers,
		       current_tier, max_tier, rewards_enabled, rewards_config,
		       prerequisites, time_limit, event_required, points,
		       difficulty, completion_rate, total_completions, is_active,
		       is_hidden, is_secret, display_order, created_at, updated_at
		FROM achievements
		WHERE 1=1
	`

	args := []interface{}{}
	argCount := 1

	if categoryID != nil {
		query += fmt.Sprintf(" AND category_id = $%d", argCount)
		args = append(args, *categoryID)
		argCount++
	}

	if activeOnly {
		query += fmt.Sprintf(" AND is_active = true", argCount)
		argCount++
	}

	if !includeHidden {
		query += fmt.Sprintf(" AND is_hidden = false", argCount)
		argCount++
	}

	query += " ORDER BY display_order ASC, name ASC"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo logros: %w", err)
	}
	defer rows.Close()

	var achievements []models.Achievement
	for rows.Next() {
		var achievement models.Achievement
		err := rows.Scan(
			&achievement.ID, &achievement.CategoryID, &achievement.Name,
			&achievement.Description, &achievement.LongDescription, &achievement.Icon,
			&achievement.Color, &achievement.BackgroundColor, &achievement.Rarity,
			&achievement.ProgressType, &achievement.TargetValue, &achievement.CurrentValue,
			&achievement.ProgressFormula, &achievement.Tiers, &achievement.CurrentTier,
			&achievement.MaxTier, &achievement.RewardsEnabled, &achievement.RewardsConfig,
			&achievement.Prerequisites, &achievement.TimeLimit, &achievement.EventRequired,
			&achievement.Points, &achievement.Difficulty, &achievement.CompletionRate,
			&achievement.TotalCompletions, &achievement.IsActive, &achievement.IsHidden,
			&achievement.IsSecret, &achievement.DisplayOrder, &achievement.CreatedAt,
			&achievement.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando logro: %w", err)
		}
		achievements = append(achievements, achievement)
	}

	return achievements, nil
}

// GetAchievement obtiene un logro específico
func (r *AchievementRepository) GetAchievement(achievementID uuid.UUID) (*models.Achievement, error) {
	query := `
		SELECT id, category_id, name, description, long_description,
		       icon, color, background_color, rarity, progress_type,
		       target_value, current_value, progress_formula, tiers,
		       current_tier, max_tier, rewards_enabled, rewards_config,
		       prerequisites, time_limit, event_required, points,
		       difficulty, completion_rate, total_completions, is_active,
		       is_hidden, is_secret, display_order, created_at, updated_at
		FROM achievements
		WHERE id = $1
	`

	var achievement models.Achievement
	err := r.db.QueryRow(query, achievementID).Scan(
		&achievement.ID, &achievement.CategoryID, &achievement.Name,
		&achievement.Description, &achievement.LongDescription, &achievement.Icon,
		&achievement.Color, &achievement.BackgroundColor, &achievement.Rarity,
		&achievement.ProgressType, &achievement.TargetValue, &achievement.CurrentValue,
		&achievement.ProgressFormula, &achievement.Tiers, &achievement.CurrentTier,
		&achievement.MaxTier, &achievement.RewardsEnabled, &achievement.RewardsConfig,
		&achievement.Prerequisites, &achievement.TimeLimit, &achievement.EventRequired,
		&achievement.Points, &achievement.Difficulty, &achievement.CompletionRate,
		&achievement.TotalCompletions, &achievement.IsActive, &achievement.IsHidden,
		&achievement.IsSecret, &achievement.DisplayOrder, &achievement.CreatedAt,
		&achievement.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("logro no encontrado")
		}
		return nil, fmt.Errorf("error obteniendo logro: %w", err)
	}

	return &achievement, nil
}

// GetPlayerAchievements obtiene los logros de un jugador
func (r *AchievementRepository) GetPlayerAchievements(playerID uuid.UUID, categoryID *uuid.UUID, completedOnly bool) ([]models.PlayerAchievement, error) {
	query := `
		SELECT player_id, achievement_id, current_progress, target_progress,
		       progress_percent, is_completed, is_claimed, current_tier,
		       rewards_claimed, rewards_data, points_earned, completion_time,
		       started_at, last_updated, completed_at, claimed_at, created_at
		FROM player_achievements
		WHERE player_id = $1
	`

	args := []interface{}{playerID}
	argCount := 2

	if categoryID != nil {
		query += fmt.Sprintf(" AND achievement_id IN (SELECT id FROM achievements WHERE category_id = $%d)", argCount)
		args = append(args, *categoryID)
		argCount++
	}

	if completedOnly {
		query += fmt.Sprintf(" AND is_completed = true", argCount)
		argCount++
	}

	query += " ORDER BY created_at DESC"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo logros del jugador: %w", err)
	}
	defer rows.Close()

	var achievements []models.PlayerAchievement
	for rows.Next() {
		var achievement models.PlayerAchievement
		err := rows.Scan(
			&achievement.PlayerID, &achievement.AchievementID, &achievement.CurrentProgress,
			&achievement.TargetProgress, &achievement.ProgressPercent, &achievement.IsCompleted,
			&achievement.IsClaimed, &achievement.CurrentTier, &achievement.RewardsClaimed,
			&achievement.RewardsData, &achievement.PointsEarned, &achievement.CompletionTime,
			&achievement.StartedAt, &achievement.LastUpdated, &achievement.CompletedAt,
			&achievement.ClaimedAt, &achievement.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando logro del jugador: %w", err)
		}
		achievements = append(achievements, achievement)
	}

	return achievements, nil
}

// GetPlayerAchievement obtiene un logro específico de un jugador
func (r *AchievementRepository) GetPlayerAchievement(playerID, achievementID uuid.UUID) (*models.PlayerAchievement, error) {
	query := `
		SELECT player_id, achievement_id, current_progress, target_progress,
		       progress_percent, is_completed, is_claimed, current_tier,
		       rewards_claimed, rewards_data, points_earned, completion_time,
		       started_at, last_updated, completed_at, claimed_at, created_at
		FROM player_achievements
		WHERE player_id = $1 AND achievement_id = $2
	`

	var achievement models.PlayerAchievement
	err := r.db.QueryRow(query, playerID, achievementID).Scan(
		&achievement.PlayerID, &achievement.AchievementID, &achievement.CurrentProgress,
		&achievement.TargetProgress, &achievement.ProgressPercent, &achievement.IsCompleted,
		&achievement.IsClaimed, &achievement.CurrentTier, &achievement.RewardsClaimed,
		&achievement.RewardsData, &achievement.PointsEarned, &achievement.CompletionTime,
		&achievement.StartedAt, &achievement.LastUpdated, &achievement.CompletedAt,
		&achievement.ClaimedAt, &achievement.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("logro del jugador no encontrado")
		}
		return nil, fmt.Errorf("error obteniendo logro del jugador: %w", err)
	}

	return &achievement, nil
}

// UpdateAchievementProgress actualiza el progreso de un logro
func (r *AchievementRepository) UpdateAchievementProgress(playerID, achievementID uuid.UUID, progress int) error {
	query := `
		INSERT INTO player_achievements (
			player_id, achievement_id, current_progress, target_progress,
			progress_percent, is_completed, is_claimed, current_tier,
			rewards_claimed, rewards_data, points_earned, completion_time,
			started_at, last_updated, completed_at, claimed_at, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17
		)
		ON CONFLICT (player_id, achievement_id) DO UPDATE SET
			current_progress = EXCLUDED.current_progress,
			progress_percent = EXCLUDED.progress_percent,
			is_completed = EXCLUDED.is_completed,
			last_updated = EXCLUDED.last_updated,
			completed_at = EXCLUDED.completed_at
	`

	// Obtener el logro para obtener información adicional
	achievement, err := r.GetAchievement(achievementID)
	if err != nil {
		return fmt.Errorf("error obteniendo logro: %w", err)
	}

	now := time.Now()
	isCompleted := progress >= achievement.TargetValue
	progressPercent := float64(progress) / float64(achievement.TargetValue) * 100

	var completionTime *time.Time
	if isCompleted {
		completionTime = &now
	}

	_, err = r.db.Exec(query,
		playerID, achievementID, progress, achievement.TargetValue,
		progressPercent, isCompleted, false, 1, false, "", 0,
		completionTime, now, now, completionTime, nil, now,
	)

	if err != nil {
		return fmt.Errorf("error actualizando progreso: %w", err)
	}

	return nil
}

// CompleteAchievement marca un logro como completado
func (r *AchievementRepository) CompleteAchievement(playerID, achievementID uuid.UUID) error {
	query := `
		UPDATE player_achievements
		SET is_completed = true, completed_at = $3, last_updated = $3
		WHERE player_id = $1 AND achievement_id = $2
	`

	now := time.Now()
	_, err := r.db.Exec(query, playerID, achievementID, now)
	if err != nil {
		return fmt.Errorf("error completando logro: %w", err)
	}

	return nil
}

// GetAchievementStatistics obtiene las estadísticas de logros de un jugador
func (r *AchievementRepository) GetAchievementStatistics(playerID uuid.UUID) (*models.AchievementStatistics, error) {
	query := `
		SELECT player_id, total_achievements, completed_achievements,
		       completion_rate, total_points, average_difficulty,
		       rarest_achievement, fastest_completion, slowest_completion,
		       last_completion, created_at, updated_at
		FROM achievement_statistics
		WHERE player_id = $1
	`

	var stats models.AchievementStatistics
	err := r.db.QueryRow(query, playerID).Scan(
		&stats.PlayerID, &stats.TotalAchievements, &stats.CompletedAchievements,
		&stats.CompletionRate, &stats.TotalPoints, &stats.AverageDifficulty,
		&stats.RarestAchievement, &stats.FastestCompletion, &stats.SlowestCompletion,
		&stats.LastCompletion, &stats.CreatedAt, &stats.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			// Crear estadísticas si no existen
			return r.createDefaultStatistics(playerID)
		}
		return nil, fmt.Errorf("error obteniendo estadísticas: %w", err)
	}

	return &stats, nil
}

// GetAchievementLeaderboard obtiene el leaderboard de logros
func (r *AchievementRepository) GetAchievementLeaderboard(limit int, worldID *uuid.UUID) ([]models.AchievementLeaderboard, error) {
	query := `
		SELECT player_id, player_name, total_achievements, completed_achievements,
		       completion_rate, total_points, rank, world_id, created_at, updated_at
		FROM achievement_leaderboard
		WHERE 1=1
	`

	args := []interface{}{}
	argCount := 1

	if worldID != nil {
		query += fmt.Sprintf(" AND world_id = $%d", argCount)
		args = append(args, *worldID)
		argCount++
	}

	query += fmt.Sprintf(" ORDER BY total_points DESC, completion_rate DESC LIMIT $%d", argCount)
	args = append(args, limit)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo leaderboard: %w", err)
	}
	defer rows.Close()

	var leaderboard []models.AchievementLeaderboard
	for rows.Next() {
		var entry models.AchievementLeaderboard
		err := rows.Scan(
			&entry.PlayerID, &entry.PlayerName, &entry.TotalAchievements,
			&entry.CompletedAchievements, &entry.CompletionRate, &entry.TotalPoints,
			&entry.Rank, &entry.WorldID, &entry.CreatedAt, &entry.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando entrada del leaderboard: %w", err)
		}
		leaderboard = append(leaderboard, entry)
	}

	return leaderboard, nil
}

// GetAchievementNotifications obtiene las notificaciones de logros de un jugador
func (r *AchievementRepository) GetAchievementNotifications(playerID uuid.UUID, unreadOnly bool) ([]models.AchievementNotification, error) {
	query := `
		SELECT id, player_id, achievement_id, notification_type, title,
		       message, is_read, created_at, read_at
		FROM achievement_notifications
		WHERE player_id = $1
	`

	if unreadOnly {
		query += " AND is_read = false"
	}

	query += " ORDER BY created_at DESC"

	rows, err := r.db.Query(query, playerID)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo notificaciones: %w", err)
	}
	defer rows.Close()

	var notifications []models.AchievementNotification
	for rows.Next() {
		var notification models.AchievementNotification
		err := rows.Scan(
			&notification.ID, &notification.PlayerID, &notification.AchievementID,
			&notification.NotificationType, &notification.Title, &notification.Message,
			&notification.IsRead, &notification.CreatedAt, &notification.ReadAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando notificación: %w", err)
		}
		notifications = append(notifications, notification)
	}

	return notifications, nil
}

// MarkNotificationAsRead marca una notificación como leída
func (r *AchievementRepository) MarkNotificationAsRead(playerID, notificationID uuid.UUID) error {
	query := `
		UPDATE achievement_notifications
		SET is_read = true, read_at = $3
		WHERE id = $1 AND player_id = $2
	`

	now := time.Now()
	_, err := r.db.Exec(query, notificationID, playerID, now)
	if err != nil {
		return fmt.Errorf("error marcando notificación como leída: %w", err)
	}

	return nil
}

// MarkRewardsAsClaimed marca las recompensas de un achievement como reclamadas
func (r *AchievementRepository) MarkRewardsAsClaimed(playerID, achievementID uuid.UUID) error {
	query := `
		UPDATE player_achievements
		SET rewards_claimed = true, claimed_at = $3
		WHERE player_id = $1 AND achievement_id = $2
	`

	now := time.Now()
	_, err := r.db.Exec(query, playerID, achievementID, now)
	if err != nil {
		return fmt.Errorf("error marcando recompensas como reclamadas: %w", err)
	}

	return nil
}

// createDefaultStatistics crea estadísticas por defecto para un jugador
func (r *AchievementRepository) createDefaultStatistics(playerID uuid.UUID) (*models.AchievementStatistics, error) {
	now := time.Now()
	stats := &models.AchievementStatistics{
		PlayerID:              playerID,
		TotalAchievements:     0,
		CompletedAchievements: 0,
		CompletionRate:        0.0,
		TotalPoints:           0,
		AverageDifficulty:     0.0,
		RarestAchievement:     "",
		FastestCompletion:     "",
		SlowestCompletion:     "",
		LastCompletion:        nil,
		CreatedAt:             now,
		UpdatedAt:             now,
	}

	query := `
		INSERT INTO achievement_statistics (
			player_id, total_achievements, completed_achievements,
			completion_rate, total_points, average_difficulty,
			rarest_achievement, fastest_completion, slowest_completion,
			last_completion, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
		)
	`

	_, err := r.db.Exec(query,
		stats.PlayerID, stats.TotalAchievements, stats.CompletedAchievements,
		stats.CompletionRate, stats.TotalPoints, stats.AverageDifficulty,
		stats.RarestAchievement, stats.FastestCompletion, stats.SlowestCompletion,
		stats.LastCompletion, stats.CreatedAt, stats.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("error creando estadísticas por defecto: %w", err)
	}

	return stats, nil
}
