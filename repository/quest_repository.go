package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"server-backend/models"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type QuestRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewQuestRepository(db *sql.DB, logger *zap.Logger) *QuestRepository {
	return &QuestRepository{
		db:     db,
		logger: logger,
	}
}

// GetQuestCategories obtiene todas las categorías de quests
func (r *QuestRepository) GetQuestCategories(activeOnly bool) ([]models.QuestCategory, error) {
	query := `
		SELECT id, name, description, icon, color, background_color,
		       display_order, is_public, show_in_dashboard, total_quests,
		       completed_count, completion_rate, is_active, created_at, updated_at
		FROM quest_categories
	`

	if activeOnly {
		query += " WHERE is_active = true"
	}

	query += " ORDER BY display_order ASC, name ASC"

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo categorías de quests: %w", err)
	}
	defer rows.Close()

	var categories []models.QuestCategory
	for rows.Next() {
		var category models.QuestCategory
		err := rows.Scan(
			&category.ID, &category.Name, &category.Description, &category.Icon,
			&category.Color, &category.BackgroundColor, &category.DisplayOrder,
			&category.IsPublic, &category.ShowInDashboard, &category.TotalQuests,
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

// GetQuestCategory obtiene una categoría específica
func (r *QuestRepository) GetQuestCategory(categoryID uuid.UUID) (*models.QuestCategory, error) {
	query := `
		SELECT id, name, description, icon, color, background_color,
		       display_order, is_public, show_in_dashboard, total_quests,
		       completed_count, completion_rate, is_active, created_at, updated_at
		FROM quest_categories
		WHERE id = $1
	`

	var category models.QuestCategory
	err := r.db.QueryRow(query, categoryID).Scan(
		&category.ID, &category.Name, &category.Description, &category.Icon,
		&category.Color, &category.BackgroundColor, &category.DisplayOrder,
		&category.IsPublic, &category.ShowInDashboard, &category.TotalQuests,
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

// CreateQuestCategory crea una nueva categoría de quests
func (r *QuestRepository) CreateQuestCategory(category *models.QuestCategory) error {
	query := `
		INSERT INTO quest_categories (
			id, name, description, icon, color, background_color,
			display_order, is_public, show_in_dashboard, total_quests,
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
		category.ShowInDashboard, category.TotalQuests, category.CompletedCount,
		category.CompletionRate, category.IsActive, now, now,
	)

	if err != nil {
		return fmt.Errorf("error creando categoría: %w", err)
	}

	return nil
}

// GetAvailableQuests obtiene las quests disponibles para un jugador
func (r *QuestRepository) GetAvailableQuests(playerID uuid.UUID, categoryID *uuid.UUID, playerLevel int, includeCompleted bool) ([]models.Quest, error) {
	query := `
		SELECT id, category_id, name, description, long_description, story_text, icon, color, background_color, rarity, quest_type, progress_type, target_value, current_value, progress_formula, tiers, current_tier, max_tier, chain_id, chain_order, next_quest_id, previous_quest_id, rewards_enabled, rewards_config, prerequisites, level_required, alliance_required, time_limit, event_required, is_repeatable, repeat_interval, max_completions, points, difficulty, completion_rate, total_completions, average_time, is_active, is_hidden, is_secret, display_order, created_at, updated_at
		FROM quests
		WHERE is_active = true AND level_required <= $1
	`

	args := []interface{}{playerLevel}
	argCount := 2

	if categoryID != nil {
		query += fmt.Sprintf(" AND category_id = $%d", argCount)
		args = append(args, *categoryID)
		argCount++
	}

	if !includeCompleted {
		query += fmt.Sprintf(" AND id NOT IN (SELECT quest_id FROM player_quests WHERE player_id = $%d AND is_completed = true)", argCount)
		args = append(args, playerID)
		argCount++
	}

	query += " ORDER BY display_order ASC, name ASC"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo quests disponibles: %w", err)
	}
	defer rows.Close()

	var quests []models.Quest
	for rows.Next() {
		var quest models.Quest
		err := rows.Scan(
			&quest.ID, &quest.CategoryID, &quest.Name, &quest.Description, &quest.LongDescription, &quest.StoryText, &quest.Icon, &quest.Color, &quest.BackgroundColor, &quest.Rarity, &quest.QuestType, &quest.ProgressType, &quest.TargetValue, &quest.CurrentValue, &quest.ProgressFormula, &quest.Tiers, &quest.CurrentTier, &quest.MaxTier, &quest.ChainID, &quest.ChainOrder, &quest.NextQuestID, &quest.PreviousQuestID, &quest.RewardsEnabled, &quest.RewardsConfig, &quest.Prerequisites, &quest.LevelRequired, &quest.AllianceRequired, &quest.TimeLimit, &quest.EventRequired, &quest.IsRepeatable, &quest.RepeatInterval, &quest.MaxCompletions, &quest.Points, &quest.Difficulty, &quest.CompletionRate, &quest.TotalCompletions, &quest.AverageTime, &quest.IsActive, &quest.IsHidden, &quest.IsSecret, &quest.DisplayOrder, &quest.CreatedAt, &quest.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando quest: %w", err)
		}
		quests = append(quests, quest)
	}

	return quests, nil
}

// GetQuest obtiene una quest específica
func (r *QuestRepository) GetQuest(questID uuid.UUID) (*models.Quest, error) {
	query := `
		SELECT id, category_id, name, description, long_description, story_text, icon, color, background_color, rarity, quest_type, progress_type, target_value, current_value, progress_formula, tiers, current_tier, max_tier, chain_id, chain_order, next_quest_id, previous_quest_id, rewards_enabled, rewards_config, prerequisites, level_required, alliance_required, time_limit, event_required, is_repeatable, repeat_interval, max_completions, points, difficulty, completion_rate, total_completions, average_time, is_active, is_hidden, is_secret, display_order, created_at, updated_at
		FROM quests
		WHERE id = $1
	`

	var quest models.Quest
	err := r.db.QueryRow(query, questID).Scan(
		&quest.ID, &quest.CategoryID, &quest.Name, &quest.Description, &quest.LongDescription, &quest.StoryText, &quest.Icon, &quest.Color, &quest.BackgroundColor, &quest.Rarity, &quest.QuestType, &quest.ProgressType, &quest.TargetValue, &quest.CurrentValue, &quest.ProgressFormula, &quest.Tiers, &quest.CurrentTier, &quest.MaxTier, &quest.ChainID, &quest.ChainOrder, &quest.NextQuestID, &quest.PreviousQuestID, &quest.RewardsEnabled, &quest.RewardsConfig, &quest.Prerequisites, &quest.LevelRequired, &quest.AllianceRequired, &quest.TimeLimit, &quest.EventRequired, &quest.IsRepeatable, &quest.RepeatInterval, &quest.MaxCompletions, &quest.Points, &quest.Difficulty, &quest.CompletionRate, &quest.TotalCompletions, &quest.AverageTime, &quest.IsActive, &quest.IsHidden, &quest.IsSecret, &quest.DisplayOrder, &quest.CreatedAt, &quest.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("quest no encontrada")
		}
		return nil, fmt.Errorf("error obteniendo quest: %w", err)
	}

	return &quest, nil
}

// GetPlayerActiveQuests obtiene las quests activas de un jugador
func (r *QuestRepository) GetPlayerActiveQuests(playerID uuid.UUID, categoryID *uuid.UUID, includeCompleted bool) ([]models.PlayerQuest, error) {
	query := `
		SELECT player_id, quest_id, current_progress, target_progress,
		       progress_percent, is_completed, rewards_claimed, current_tier,
		       rewards_data, points_earned, completion_time,
		       started_at, last_updated, completed_at, claimed_at, created_at
		FROM player_quests
		WHERE player_id = $1
	`

	args := []interface{}{playerID}
	argCount := 2

	if categoryID != nil {
		query += fmt.Sprintf(" AND quest_id IN (SELECT id FROM quests WHERE category_id = $%d)", argCount)
		args = append(args, *categoryID)
		argCount++
	}

	if !includeCompleted {
		query += fmt.Sprintf(" AND is_completed = false", argCount)
		argCount++
	}

	query += " ORDER BY created_at DESC"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo quests activas: %w", err)
	}
	defer rows.Close()

	var quests []models.PlayerQuest
	for rows.Next() {
		var quest models.PlayerQuest
		err := rows.Scan(
			&quest.PlayerID, &quest.QuestID, &quest.CurrentProgress,
			&quest.TargetProgress, &quest.ProgressPercent, &quest.IsCompleted,
			&quest.RewardsClaimed, &quest.CurrentTier, &quest.RewardsData,
			&quest.PointsEarned, &quest.CompletionTime, &quest.StartedAt,
			&quest.LastUpdated, &quest.CompletedAt, &quest.ClaimedAt, &quest.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando quest del jugador: %w", err)
		}
		quests = append(quests, quest)
	}

	return quests, nil
}

// GetPlayerQuest obtiene una quest específica de un jugador
func (r *QuestRepository) GetPlayerQuest(playerID, questID uuid.UUID) (*models.PlayerQuest, error) {
	query := `
		SELECT id, player_id, quest_id, current_progress, target_progress,
		       progress_percent, is_completed, is_claimed, is_failed,
		       current_tier, completion_count, rewards_claimed, rewards_data,
		       points_earned, completion_time, time_spent, started_at,
		       last_updated, completed_at, claimed_at, failed_at,
		       expires_at, created_at
		FROM player_quests
		WHERE player_id = $1 AND quest_id = $2
	`

	var playerQuest models.PlayerQuest
	err := r.db.QueryRow(query, playerID, questID).Scan(
		&playerQuest.ID, &playerQuest.PlayerID, &playerQuest.QuestID,
		&playerQuest.CurrentProgress, &playerQuest.TargetProgress,
		&playerQuest.ProgressPercent, &playerQuest.IsCompleted,
		&playerQuest.IsClaimed, &playerQuest.IsFailed, &playerQuest.CurrentTier,
		&playerQuest.CompletionCount, &playerQuest.RewardsClaimed,
		&playerQuest.RewardsData, &playerQuest.PointsEarned,
		&playerQuest.CompletionTime, &playerQuest.TimeSpent, &playerQuest.StartedAt,
		&playerQuest.LastUpdated, &playerQuest.CompletedAt, &playerQuest.ClaimedAt,
		&playerQuest.FailedAt, &playerQuest.ExpiresAt, &playerQuest.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No se encontró la quest del jugador
		}
		return nil, fmt.Errorf("error obteniendo quest del jugador: %w", err)
	}

	return &playerQuest, nil
}

// CreatePlayerQuest crea una nueva quest para un jugador
func (r *QuestRepository) CreatePlayerQuest(playerQuest *models.PlayerQuest) error {
	query := `
		INSERT INTO player_quests (
			id, player_id, quest_id, current_progress, target_progress,
			progress_percent, is_completed, is_claimed, is_failed,
			current_tier, completion_count, rewards_claimed, rewards_data,
			points_earned, completion_time, time_spent, started_at,
			last_updated, completed_at, claimed_at, failed_at,
			expires_at, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13,
			$14, $15, $16, $17, $18, $19, $20, $21, $22, $23
		)
	`

	// Generar ID si no existe
	if playerQuest.ID == uuid.Nil {
		playerQuest.ID = uuid.New()
	}

	_, err := r.db.Exec(query,
		playerQuest.ID, playerQuest.PlayerID, playerQuest.QuestID,
		playerQuest.CurrentProgress, playerQuest.TargetProgress,
		playerQuest.ProgressPercent, playerQuest.IsCompleted,
		playerQuest.IsClaimed, playerQuest.IsFailed, playerQuest.CurrentTier,
		playerQuest.CompletionCount, playerQuest.RewardsClaimed,
		playerQuest.RewardsData, playerQuest.PointsEarned,
		playerQuest.CompletionTime, playerQuest.TimeSpent, playerQuest.StartedAt,
		playerQuest.LastUpdated, playerQuest.CompletedAt, playerQuest.ClaimedAt,
		playerQuest.FailedAt, playerQuest.ExpiresAt, playerQuest.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("error creando quest del jugador: %w", err)
	}

	return nil
}

// UpdatePlayerQuest actualiza una quest del jugador
func (r *QuestRepository) UpdatePlayerQuest(playerQuest *models.PlayerQuest) error {
	query := `
		UPDATE player_quests SET
			current_progress = $1, target_progress = $2, progress_percent = $3,
			is_completed = $4, is_claimed = $5, is_failed = $6,
			current_tier = $7, completion_count = $8, rewards_claimed = $9,
			rewards_data = $10, points_earned = $11, completion_time = $12,
			time_spent = $13, last_updated = $14, completed_at = $15,
			claimed_at = $16, failed_at = $17, expires_at = $18
		WHERE player_id = $19 AND quest_id = $20
	`

	_, err := r.db.Exec(query,
		playerQuest.CurrentProgress, playerQuest.TargetProgress, playerQuest.ProgressPercent,
		playerQuest.IsCompleted, playerQuest.IsClaimed, playerQuest.IsFailed,
		playerQuest.CurrentTier, playerQuest.CompletionCount, playerQuest.RewardsClaimed,
		playerQuest.RewardsData, playerQuest.PointsEarned, playerQuest.CompletionTime,
		playerQuest.TimeSpent, playerQuest.LastUpdated, playerQuest.CompletedAt,
		playerQuest.ClaimedAt, playerQuest.FailedAt, playerQuest.ExpiresAt,
		playerQuest.PlayerID, playerQuest.QuestID,
	)

	if err != nil {
		return fmt.Errorf("error actualizando quest del jugador: %w", err)
	}

	return nil
}

// GetPlayerQuestHistory obtiene el historial de quests de un jugador
func (r *QuestRepository) GetPlayerQuestHistory(playerID uuid.UUID, limit int) ([]models.PlayerQuest, error) {
	query := `
		SELECT id, player_id, quest_id, current_progress, target_progress,
		       progress_percent, is_completed, is_claimed, is_failed,
		       current_tier, completion_count, rewards_claimed, rewards_data,
		       points_earned, completion_time, time_spent, started_at,
		       last_updated, completed_at, claimed_at, failed_at,
		       expires_at, created_at
		FROM player_quests
		WHERE player_id = $1
		ORDER BY created_at DESC
	`

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := r.db.Query(query, playerID)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo historial de quests: %w", err)
	}
	defer rows.Close()

	var history []models.PlayerQuest
	for rows.Next() {
		var playerQuest models.PlayerQuest
		err := rows.Scan(
			&playerQuest.ID, &playerQuest.PlayerID, &playerQuest.QuestID,
			&playerQuest.CurrentProgress, &playerQuest.TargetProgress,
			&playerQuest.ProgressPercent, &playerQuest.IsCompleted,
			&playerQuest.IsClaimed, &playerQuest.IsFailed, &playerQuest.CurrentTier,
			&playerQuest.CompletionCount, &playerQuest.RewardsClaimed,
			&playerQuest.RewardsData, &playerQuest.PointsEarned,
			&playerQuest.CompletionTime, &playerQuest.TimeSpent, &playerQuest.StartedAt,
			&playerQuest.LastUpdated, &playerQuest.CompletedAt, &playerQuest.ClaimedAt,
			&playerQuest.FailedAt, &playerQuest.ExpiresAt, &playerQuest.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando historial de quests: %w", err)
		}
		history = append(history, playerQuest)
	}

	return history, nil
}

// GetDailyQuests obtiene las quests diarias de un jugador
func (r *QuestRepository) GetDailyQuests(playerID uuid.UUID) ([]models.Quest, error) {
	query := `
		SELECT DISTINCT q.id, q.category_id, q.name, q.description, q.long_description,
		       q.story_text, q.icon, q.color, q.background_color, q.rarity,
		       q.quest_type, q.progress_type, q.target_value, q.current_value,
		       q.required_progress, q.progress_formula, q.tiers, q.current_tier,
		       q.max_tier, q.chain_id, q.chain_order, q.next_quest_id,
		       q.previous_quest_id, q.rewards_enabled, q.rewards_config,
		       q.prerequisites, q.level_required, q.alliance_required,
		       q.time_limit, q.event_required, q.is_repeatable, q.repeat_interval,
		       q.max_completions, q.points, q.difficulty, q.completion_rate,
		       q.total_completions, q.average_time, q.is_active, q.is_hidden,
		       q.is_secret, q.display_order, q.created_at, q.updated_at
		FROM quests q
		INNER JOIN player_quests pq ON q.id = pq.quest_id
		WHERE pq.player_id = $1 
		  AND q.quest_type = 'daily'
		  AND q.is_active = true
		  AND (pq.is_completed = false OR q.is_repeatable = true)
		ORDER BY q.display_order ASC, q.name ASC
	`

	rows, err := r.db.Query(query, playerID)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo quests diarias: %w", err)
	}
	defer rows.Close()

	var quests []models.Quest
	for rows.Next() {
		var quest models.Quest
		err := rows.Scan(
			&quest.ID, &quest.CategoryID, &quest.Name, &quest.Description,
			&quest.LongDescription, &quest.StoryText, &quest.Icon, &quest.Color,
			&quest.BackgroundColor, &quest.Rarity, &quest.QuestType,
			&quest.ProgressType, &quest.TargetValue, &quest.CurrentValue,
			&quest.RequiredProgress, &quest.ProgressFormula, &quest.Tiers,
			&quest.CurrentTier, &quest.MaxTier, &quest.ChainID, &quest.ChainOrder,
			&quest.NextQuestID, &quest.PreviousQuestID, &quest.RewardsEnabled,
			&quest.RewardsConfig, &quest.Prerequisites, &quest.LevelRequired,
			&quest.AllianceRequired, &quest.TimeLimit, &quest.EventRequired,
			&quest.IsRepeatable, &quest.RepeatInterval, &quest.MaxCompletions,
			&quest.Points, &quest.Difficulty, &quest.CompletionRate,
			&quest.TotalCompletions, &quest.AverageTime, &quest.IsActive,
			&quest.IsHidden, &quest.IsSecret, &quest.DisplayOrder,
			&quest.CreatedAt, &quest.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando quests diarias: %w", err)
		}
		quests = append(quests, quest)
	}

	return quests, nil
}

// RefreshDailyQuests refresca las quests diarias de un jugador
func (r *QuestRepository) RefreshDailyQuests(playerID uuid.UUID) error {
	// Primero, marcar como expiradas las quests diarias completadas
	expireQuery := `
		UPDATE player_quests 
		SET is_failed = true, failed_at = $2
		WHERE player_id = $1 
		  AND quest_id IN (
		    SELECT id FROM quests 
		    WHERE quest_type = 'daily' AND is_active = true
		  )
		  AND is_completed = true
		  AND created_at < CURRENT_DATE
	`

	now := time.Now()
	_, err := r.db.Exec(expireQuery, playerID, now)
	if err != nil {
		return fmt.Errorf("error expirando quests diarias: %w", err)
	}

	// Luego, crear nuevas quests diarias disponibles
	refreshQuery := `
		INSERT INTO player_quests (
			id, player_id, quest_id, current_progress, target_progress,
			progress_percent, is_completed, is_claimed, is_failed,
			current_tier, completion_count, rewards_claimed, rewards_data,
			points_earned, completion_time, time_spent, started_at,
			last_updated, completed_at, claimed_at, failed_at,
			expires_at, created_at
		)
		SELECT 
			uuid_generate_v4(), $1, q.id, 0, q.required_progress,
			0.0, false, false, false, 1, 0, false, '{}',
			0, NULL, 0, $2, $2, NULL, NULL, NULL,
			$2 + INTERVAL '1 day', $2
		FROM quests q
		WHERE q.quest_type = 'daily' 
		  AND q.is_active = true
		  AND q.level_required <= (
		    SELECT COALESCE(level, 1) FROM players WHERE id = $1
		  )
		  AND NOT EXISTS (
		    SELECT 1 FROM player_quests pq 
		    WHERE pq.player_id = $1 AND pq.quest_id = q.id
		      AND pq.created_at >= CURRENT_DATE
		  )
	`

	_, err = r.db.Exec(refreshQuery, playerID, now)
	if err != nil {
		return fmt.Errorf("error refrescando quests diarias: %w", err)
	}

	return nil
}

// GetQuestRewards obtiene las recompensas de una quest
func (r *QuestRepository) GetQuestRewards(questID uuid.UUID) ([]models.QuestReward, error) {
	query := `
		SELECT id, quest_id, reward_type, reward_data, quantity,
		       tier_required, is_repeatable, is_guaranteed,
		       is_active, created_at
		FROM quest_rewards
		WHERE quest_id = $1 AND is_active = true
		ORDER BY tier_required ASC, quantity DESC
	`

	rows, err := r.db.Query(query, questID)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo recompensas de quest: %w", err)
	}
	defer rows.Close()

	var rewards []models.QuestReward
	for rows.Next() {
		var reward models.QuestReward
		err := rows.Scan(
			&reward.ID, &reward.QuestID, &reward.RewardType, &reward.RewardData,
			&reward.Quantity, &reward.TierRequired, &reward.IsRepeatable,
			&reward.IsGuaranteed, &reward.IsActive, &reward.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando recompensa de quest: %w", err)
		}
		rewards = append(rewards, reward)
	}

	return rewards, nil
}

// MarkRewardsAsClaimed marca las recompensas de una quest como reclamadas
func (r *QuestRepository) MarkRewardsAsClaimed(playerID, questID uuid.UUID) error {
	query := `
		UPDATE player_quests
		SET rewards_claimed = true, claimed_at = $3
		WHERE player_id = $1 AND quest_id = $2
	`

	now := time.Now()
	_, err := r.db.Exec(query, playerID, questID, now)
	if err != nil {
		return fmt.Errorf("error marcando recompensas como reclamadas: %w", err)
	}

	return nil
}

// UpdateQuestProgress actualiza el progreso de una quest
func (r *QuestRepository) UpdateQuestProgress(playerID, questID uuid.UUID, progress int, eventData map[string]interface{}) error {
	query := `
		INSERT INTO player_quests (
			player_id, quest_id, current_progress, target_progress,
			progress_percent, is_completed, rewards_claimed, current_tier,
			rewards_data, points_earned, completion_time,
			started_at, last_updated, completed_at, claimed_at, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
		)
		ON CONFLICT (player_id, quest_id) DO UPDATE SET
			current_progress = EXCLUDED.current_progress,
			progress_percent = EXCLUDED.progress_percent,
			is_completed = EXCLUDED.is_completed,
			last_updated = EXCLUDED.last_updated,
			completed_at = EXCLUDED.completed_at
	`

	// Obtener la quest para obtener información adicional
	quest, err := r.GetQuest(questID)
	if err != nil {
		return fmt.Errorf("error obteniendo quest: %w", err)
	}

	now := time.Now()
	isCompleted := progress >= quest.TargetValue
	progressPercent := float64(progress) / float64(quest.TargetValue) * 100

	var completionTime *time.Time
	if isCompleted {
		completionTime = &now
	}

	// Convertir eventData a JSON
	eventDataJSON := ""
	if eventData != nil {
		if jsonData, err := json.Marshal(eventData); err == nil {
			eventDataJSON = string(jsonData)
		}
	}

	_, err = r.db.Exec(query,
		playerID, questID, progress, quest.TargetValue,
		progressPercent, isCompleted, false, 1, eventDataJSON, 0,
		completionTime, now, now, completionTime, nil, now,
	)

	if err != nil {
		return fmt.Errorf("error actualizando progreso: %w", err)
	}

	return nil
}

// CompleteQuest marca una quest como completada
func (r *QuestRepository) CompleteQuest(playerID, questID uuid.UUID) error {
	query := `
		UPDATE player_quests
		SET is_completed = true, completed_at = $3, last_updated = $3
		WHERE player_id = $1 AND quest_id = $2
	`

	now := time.Now()
	_, err := r.db.Exec(query, playerID, questID, now)
	if err != nil {
		return fmt.Errorf("error completando quest: %w", err)
	}

	return nil
}

// MarkQuestRewardsClaimed marca las recompensas de una quest como reclamadas
func (r *QuestRepository) MarkQuestRewardsClaimed(playerID, questID uuid.UUID) error {
	query := `
		UPDATE player_quests
		SET rewards_claimed = true, claimed_at = $3, last_updated = $3
		WHERE player_id = $1 AND quest_id = $2
	`

	now := time.Now()
	_, err := r.db.Exec(query, playerID, questID, now)
	if err != nil {
		return fmt.Errorf("error marcando recompensas como reclamadas: %w", err)
	}

	return nil
}

// GetPlayerQuestStatistics obtiene las estadísticas de quests de un jugador
func (r *QuestRepository) GetPlayerQuestStatistics(playerID uuid.UUID) (*models.QuestStatistics, error) {
	query := `
		SELECT id, player_id, total_quests, completed_quests, failed_quests,
		       completion_rate, total_points, points_this_week, points_this_month,
		       category_stats, type_stats, chains_completed, current_chains,
		       last_quest, streak_days, longest_streak, average_time,
		       total_rewards_claimed, rewards_value, average_difficulty,
		       fastest_completion, slowest_completion, last_completion,
		       first_quest, last_updated, created_at, updated_at
		FROM quest_statistics
		WHERE player_id = $1
	`

	var stats models.QuestStatistics
	err := r.db.QueryRow(query, playerID).Scan(
		&stats.ID, &stats.PlayerID, &stats.TotalQuests, &stats.CompletedQuests,
		&stats.FailedQuests, &stats.CompletionRate, &stats.TotalPoints,
		&stats.PointsThisWeek, &stats.PointsThisMonth, &stats.CategoryStats,
		&stats.TypeStats, &stats.ChainsCompleted, &stats.CurrentChains,
		&stats.LastQuest, &stats.StreakDays, &stats.LongestStreak,
		&stats.AverageTime, &stats.TotalRewardsClaimed, &stats.RewardsValue,
		&stats.AverageDifficulty, &stats.FastestCompletion, &stats.SlowestCompletion,
		&stats.LastCompletion, &stats.FirstQuest, &stats.LastUpdated,
		&stats.CreatedAt, &stats.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			// Crear estadísticas por defecto
			return r.createDefaultQuestStatistics(playerID)
		}
		return nil, fmt.Errorf("error obteniendo estadísticas de quests: %w", err)
	}

	return &stats, nil
}

// GetQuestStatistics obtiene las estadísticas globales de una quest específica
func (r *QuestRepository) GetQuestStatistics(questID uuid.UUID) (*models.QuestStatistics, error) {
	query := `
		SELECT 
			COUNT(*) as total_quests,
			COUNT(CASE WHEN is_completed = true THEN 1 END) as completed_quests,
			COUNT(CASE WHEN is_failed = true THEN 1 END) as failed_quests,
			CASE 
				WHEN COUNT(*) > 0 THEN 
					ROUND((COUNT(CASE WHEN is_completed = true THEN 1 END)::float / COUNT(*)::float) * 100, 2)
				ELSE 0 
			END as completion_rate,
			AVG(CASE WHEN is_completed = true THEN EXTRACT(EPOCH FROM (completed_at - started_at))/60 END) as average_time
		FROM player_quests
		WHERE quest_id = $1
	`

	var stats models.QuestStatistics
	var avgTime sql.NullFloat64

	err := r.db.QueryRow(query, questID).Scan(
		&stats.TotalQuests, &stats.CompletedQuests, &stats.FailedQuests,
		&stats.CompletionRate, &avgTime,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			// Retornar estadísticas por defecto
			return &models.QuestStatistics{
				TotalQuests:     0,
				CompletedQuests: 0,
				FailedQuests:    0,
				CompletionRate:  0.0,
				AverageTime:     0,
			}, nil
		}
		return nil, fmt.Errorf("error obteniendo estadísticas de quest: %w", err)
	}

	// Convertir tiempo promedio a minutos
	if avgTime.Valid {
		stats.AverageTime = int(avgTime.Float64)
	}

	return &stats, nil
}

// createDefaultQuestStatistics crea estadísticas por defecto para un jugador
func (r *QuestRepository) createDefaultQuestStatistics(playerID uuid.UUID) (*models.QuestStatistics, error) {
	now := time.Now()
	stats := &models.QuestStatistics{
		ID:                  uuid.New(),
		PlayerID:            playerID,
		TotalQuests:         0,
		CompletedQuests:     0,
		FailedQuests:        0,
		CompletionRate:      0.0,
		TotalPoints:         0,
		PointsThisWeek:      0,
		PointsThisMonth:     0,
		CategoryStats:       "{}",
		TypeStats:           "{}",
		ChainsCompleted:     0,
		CurrentChains:       "{}",
		LastQuest:           nil,
		StreakDays:          0,
		LongestStreak:       0,
		AverageTime:         0,
		TotalRewardsClaimed: 0,
		RewardsValue:        0,
		AverageDifficulty:   0.0,
		FastestCompletion:   nil,
		SlowestCompletion:   nil,
		LastCompletion:      nil,
		FirstQuest:          now,
		LastUpdated:         now,
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	query := `
		INSERT INTO quest_statistics (
			id, player_id, total_quests, completed_quests, failed_quests,
			completion_rate, total_points, points_this_week, points_this_month,
			category_stats, type_stats, chains_completed, current_chains,
			last_quest, streak_days, longest_streak, average_time,
			total_rewards_claimed, rewards_value, average_difficulty,
			fastest_completion, slowest_completion, last_completion,
			first_quest, last_updated, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27
		)
	`

	_, err := r.db.Exec(query,
		stats.ID, stats.PlayerID, stats.TotalQuests, stats.CompletedQuests, stats.FailedQuests,
		stats.CompletionRate, stats.TotalPoints, stats.PointsThisWeek, stats.PointsThisMonth,
		stats.CategoryStats, stats.TypeStats, stats.ChainsCompleted, stats.CurrentChains,
		stats.LastQuest, stats.StreakDays, stats.LongestStreak, stats.AverageTime,
		stats.TotalRewardsClaimed, stats.RewardsValue, stats.AverageDifficulty,
		stats.FastestCompletion, stats.SlowestCompletion, stats.LastCompletion,
		stats.FirstQuest, stats.LastUpdated, stats.CreatedAt, stats.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("error creando estadísticas por defecto: %w", err)
	}

	return stats, nil
}

// GetAllQuestCategories obtiene todas las categorías de quests
func (r *QuestRepository) GetAllQuestCategories() ([]models.QuestCategory, error) {
	return r.GetQuestCategories(false) // activeOnly = false
}

// GetQuestCategoryByID obtiene una categoría por ID (compatibilidad con handler)
func (r *QuestRepository) GetQuestCategoryByID(id int) (*models.QuestCategory, error) {
	// Convertir int a UUID (esto es temporal, debería usar UUID)
	// Por ahora, buscamos por el primer UUID que encontremos
	query := `
		SELECT id, name, description, icon, color, background_color,
		       display_order, is_public, show_in_dashboard, total_quests,
		       completed_count, completion_rate, is_active, created_at, updated_at
		FROM quest_categories
		LIMIT 1
	`

	var category models.QuestCategory
	err := r.db.QueryRow(query).Scan(
		&category.ID, &category.Name, &category.Description, &category.Icon,
		&category.Color, &category.BackgroundColor, &category.DisplayOrder,
		&category.IsPublic, &category.ShowInDashboard, &category.TotalQuests,
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

// UpdateQuestCategory actualiza una categoría de quests
func (r *QuestRepository) UpdateQuestCategory(category *models.QuestCategory) error {
	query := `
		UPDATE quest_categories
		SET name = $2, description = $3, icon = $4, color = $5, background_color = $6,
		    display_order = $7, is_public = $8, show_in_dashboard = $9, total_quests = $10,
		    completed_count = $11, completion_rate = $12, is_active = $13, updated_at = $14
		WHERE id = $1
	`

	now := time.Now()
	_, err := r.db.Exec(query,
		category.ID, category.Name, category.Description, category.Icon, category.Color,
		category.BackgroundColor, category.DisplayOrder, category.IsPublic,
		category.ShowInDashboard, category.TotalQuests, category.CompletedCount,
		category.CompletionRate, category.IsActive, now,
	)

	if err != nil {
		return fmt.Errorf("error actualizando categoría: %w", err)
	}

	return nil
}

// DeleteQuestCategory elimina una categoría de quests
func (r *QuestRepository) DeleteQuestCategory(id int) error {
	// Por ahora, marcamos como inactiva en lugar de eliminar
	query := `
		UPDATE quest_categories
		SET is_active = false, updated_at = $2
		WHERE id = $1
	`

	now := time.Now()
	_, err := r.db.Exec(query, id, now)

	if err != nil {
		return fmt.Errorf("error eliminando categoría: %w", err)
	}

	return nil
}

// CreateQuest crea una nueva quest
func (r *QuestRepository) CreateQuest(quest *models.Quest) error {
	query := `
		INSERT INTO quests (
			id, category_id, name, description, long_description, story_text,
			icon, color, background_color, rarity, quest_type, progress_type,
			target_value, current_value, required_progress, progress_formula,
			tiers, current_tier, max_tier, chain_id, chain_order, next_quest_id,
			previous_quest_id, rewards_enabled, rewards_config, prerequisites,
			level_required, alliance_required, time_limit, event_required,
			is_repeatable, repeat_interval, max_completions, points, difficulty,
			completion_rate, total_completions, average_time, is_active,
			is_hidden, is_secret, display_order, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16,
			$17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30,
			$31, $32, $33, $34, $35, $36, $37, $38, $39, $40, $41, $42, $43
		)
	`

	now := time.Now()
	if quest.ID == uuid.Nil {
		quest.ID = uuid.New()
	}

	_, err := r.db.Exec(query,
		quest.ID, quest.CategoryID, quest.Name, quest.Description, quest.LongDescription,
		quest.StoryText, quest.Icon, quest.Color, quest.BackgroundColor, quest.Rarity,
		quest.QuestType, quest.ProgressType, quest.TargetValue, quest.CurrentValue,
		quest.RequiredProgress, quest.ProgressFormula, quest.Tiers, quest.CurrentTier,
		quest.MaxTier, quest.ChainID, quest.ChainOrder, quest.NextQuestID,
		quest.PreviousQuestID, quest.RewardsEnabled, quest.RewardsConfig,
		quest.Prerequisites, quest.LevelRequired, quest.AllianceRequired,
		quest.TimeLimit, quest.EventRequired, quest.IsRepeatable, quest.RepeatInterval,
		quest.MaxCompletions, quest.Points, quest.Difficulty, quest.CompletionRate,
		quest.TotalCompletions, quest.AverageTime, quest.IsActive, quest.IsHidden,
		quest.IsSecret, quest.DisplayOrder, now, now,
	)

	if err != nil {
		return fmt.Errorf("error creando quest: %w", err)
	}

	return nil
}

// GetQuestByID obtiene una quest por ID (compatibilidad con handler)
func (r *QuestRepository) GetQuestByID(id int) (*models.Quest, error) {
	// Convertir int a UUID (esto es temporal, debería usar UUID)
	// Por ahora, buscamos por el primer UUID que encontremos
	query := `
		SELECT id, category_id, name, description, long_description, story_text, icon, color, background_color, rarity, quest_type, progress_type, target_value, current_value, required_progress, progress_formula, tiers, current_tier, max_tier, chain_id, chain_order, next_quest_id, previous_quest_id, rewards_enabled, rewards_config, prerequisites, level_required, alliance_required, time_limit, event_required, is_repeatable, repeat_interval, max_completions, points, difficulty, completion_rate, total_completions, average_time, is_active, is_hidden, is_secret, display_order, created_at, updated_at
		FROM quests
		LIMIT 1
	`

	var quest models.Quest
	err := r.db.QueryRow(query).Scan(
		&quest.ID, &quest.CategoryID, &quest.Name, &quest.Description, &quest.LongDescription, &quest.StoryText, &quest.Icon, &quest.Color, &quest.BackgroundColor, &quest.Rarity, &quest.QuestType, &quest.ProgressType, &quest.TargetValue, &quest.CurrentValue, &quest.RequiredProgress, &quest.ProgressFormula, &quest.Tiers, &quest.CurrentTier, &quest.MaxTier, &quest.ChainID, &quest.ChainOrder, &quest.NextQuestID, &quest.PreviousQuestID, &quest.RewardsEnabled, &quest.RewardsConfig, &quest.Prerequisites, &quest.LevelRequired, &quest.AllianceRequired, &quest.TimeLimit, &quest.EventRequired, &quest.IsRepeatable, &quest.RepeatInterval, &quest.MaxCompletions, &quest.Points, &quest.Difficulty, &quest.CompletionRate, &quest.TotalCompletions, &quest.AverageTime, &quest.IsActive, &quest.IsHidden, &quest.IsSecret, &quest.DisplayOrder, &quest.CreatedAt, &quest.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("quest no encontrada")
		}
		return nil, fmt.Errorf("error obteniendo quest: %w", err)
	}

	return &quest, nil
}

// UpdateQuest actualiza una quest
func (r *QuestRepository) UpdateQuest(quest *models.Quest) error {
	query := `
		UPDATE quests
		SET category_id = $2, name = $3, description = $4, long_description = $5,
		    story_text = $6, icon = $7, color = $8, background_color = $9,
		    rarity = $10, quest_type = $11, progress_type = $12, target_value = $13,
		    current_value = $14, required_progress = $15, progress_formula = $16,
		    tiers = $17, current_tier = $18, max_tier = $19, chain_id = $20,
		    chain_order = $21, next_quest_id = $22, previous_quest_id = $23,
		    rewards_enabled = $24, rewards_config = $25, prerequisites = $26,
		    level_required = $27, alliance_required = $28, time_limit = $29,
		    event_required = $30, is_repeatable = $31, repeat_interval = $32,
		    max_completions = $33, points = $34, difficulty = $35,
		    completion_rate = $36, total_completions = $37, average_time = $38,
		    is_active = $39, is_hidden = $40, is_secret = $41, display_order = $42,
		    updated_at = $43
		WHERE id = $1
	`

	now := time.Now()
	_, err := r.db.Exec(query,
		quest.ID, quest.CategoryID, quest.Name, quest.Description, quest.LongDescription,
		quest.StoryText, quest.Icon, quest.Color, quest.BackgroundColor, quest.Rarity,
		quest.QuestType, quest.ProgressType, quest.TargetValue, quest.CurrentValue,
		quest.RequiredProgress, quest.ProgressFormula, quest.Tiers, quest.CurrentTier,
		quest.MaxTier, quest.ChainID, quest.ChainOrder, quest.NextQuestID,
		quest.PreviousQuestID, quest.RewardsEnabled, quest.RewardsConfig,
		quest.Prerequisites, quest.LevelRequired, quest.AllianceRequired,
		quest.TimeLimit, quest.EventRequired, quest.IsRepeatable, quest.RepeatInterval,
		quest.MaxCompletions, quest.Points, quest.Difficulty, quest.CompletionRate,
		quest.TotalCompletions, quest.AverageTime, quest.IsActive, quest.IsHidden,
		quest.IsSecret, quest.DisplayOrder, now,
	)

	if err != nil {
		return fmt.Errorf("error actualizando quest: %w", err)
	}

	return nil
}

// DeleteQuest elimina una quest
func (r *QuestRepository) DeleteQuest(id int) error {
	// Por ahora, marcamos como inactiva en lugar de eliminar
	query := `
		UPDATE quests
		SET is_active = false, updated_at = $2
		WHERE id = $1
	`

	now := time.Now()
	_, err := r.db.Exec(query, id, now)

	if err != nil {
		return fmt.Errorf("error eliminando quest: %w", err)
	}

	return nil
}

// GetQuestsByCategory obtiene quests por categoría
func (r *QuestRepository) GetQuestsByCategory(categoryID int) ([]models.Quest, error) {
	// Convertir int a UUID (esto es temporal, debería usar UUID)
	// Por ahora, obtenemos todas las quests activas
	query := `
		SELECT id, category_id, name, description, long_description, story_text, icon, color, background_color, rarity, quest_type, progress_type, target_value, current_value, required_progress, progress_formula, tiers, current_tier, max_tier, chain_id, chain_order, next_quest_id, previous_quest_id, rewards_enabled, rewards_config, prerequisites, level_required, alliance_required, time_limit, event_required, is_repeatable, repeat_interval, max_completions, points, difficulty, completion_rate, total_completions, average_time, is_active, is_hidden, is_secret, display_order, created_at, updated_at
		FROM quests
		WHERE is_active = true
		ORDER BY display_order ASC, name ASC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo quests por categoría: %w", err)
	}
	defer rows.Close()

	var quests []models.Quest
	for rows.Next() {
		var quest models.Quest
		err := rows.Scan(
			&quest.ID, &quest.CategoryID, &quest.Name, &quest.Description, &quest.LongDescription, &quest.StoryText, &quest.Icon, &quest.Color, &quest.BackgroundColor, &quest.Rarity, &quest.QuestType, &quest.ProgressType, &quest.TargetValue, &quest.CurrentValue, &quest.RequiredProgress, &quest.ProgressFormula, &quest.Tiers, &quest.CurrentTier, &quest.MaxTier, &quest.ChainID, &quest.ChainOrder, &quest.NextQuestID, &quest.PreviousQuestID, &quest.RewardsEnabled, &quest.RewardsConfig, &quest.Prerequisites, &quest.LevelRequired, &quest.AllianceRequired, &quest.TimeLimit, &quest.EventRequired, &quest.IsRepeatable, &quest.RepeatInterval, &quest.MaxCompletions, &quest.Points, &quest.Difficulty, &quest.CompletionRate, &quest.TotalCompletions, &quest.AverageTime, &quest.IsActive, &quest.IsHidden, &quest.IsSecret, &quest.DisplayOrder, &quest.CreatedAt, &quest.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando quest: %w", err)
		}
		quests = append(quests, quest)
	}

	return quests, nil
}
