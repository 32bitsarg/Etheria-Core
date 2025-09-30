package repository

import (
	"database/sql"
	"fmt"
	"time"

	"server-backend/models"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type EventRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewEventRepository(db *sql.DB, logger *zap.Logger) *EventRepository {
	return &EventRepository{
		db:     db,
		logger: logger,
	}
}

// ==================== CATEGORÍAS DE EVENTOS ====================

func (r *EventRepository) CreateEventCategory(category *models.EventCategory) error {
	query := `
		INSERT INTO event_categories (
			id, name, description, icon, color, background_color,
			display_order, is_public, show_in_dashboard, total_events,
			active_count, completion_rate, is_active, created_at, updated_at
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
		category.ShowInDashboard, category.TotalEvents, category.ActiveCount,
		category.CompletionRate, category.IsActive, now, now,
	)

	if err != nil {
		return fmt.Errorf("error creando categoría: %w", err)
	}

	return nil
}

func (r *EventRepository) GetEventCategory(categoryID uuid.UUID) (*models.EventCategory, error) {
	query := `
		SELECT id, name, description, icon, color, background_color,
		       display_order, is_public, show_in_dashboard, total_events,
		       active_count, completion_rate, is_active, created_at, updated_at
		FROM event_categories
		WHERE id = $1
	`

	var category models.EventCategory
	err := r.db.QueryRow(query, categoryID).Scan(
		&category.ID, &category.Name, &category.Description, &category.Icon,
		&category.Color, &category.BackgroundColor, &category.DisplayOrder,
		&category.IsPublic, &category.ShowInDashboard, &category.TotalEvents,
		&category.ActiveCount, &category.CompletionRate, &category.IsActive,
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

func (r *EventRepository) GetAllEventCategories() ([]models.EventCategory, error) {
	query := `SELECT * FROM event_categories WHERE is_active = true ORDER BY display_order, name`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error getting event categories: %v", err)
	}
	defer rows.Close()

	var categories []models.EventCategory
	for rows.Next() {
		var category models.EventCategory
		err := rows.Scan(
			&category.ID, &category.Name, &category.Description, &category.Icon,
			&category.Color, &category.BackgroundColor, &category.DisplayOrder,
			&category.IsPublic, &category.ShowInDashboard, &category.TotalEvents,
			&category.ActiveCount, &category.CompletionRate, &category.IsActive,
			&category.CreatedAt, &category.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning event category: %v", err)
		}
		categories = append(categories, category)
	}

	return categories, nil
}

func (r *EventRepository) UpdateEventCategory(category *models.EventCategory) error {
	query := `
		UPDATE event_categories SET 
			name = ?, description = ?, icon = ?, color = ?, background_color = ?,
			display_order = ?, is_public = ?, show_in_dashboard = ?, is_active = ?, updated_at = ?
		WHERE id = ?
	`

	category.UpdatedAt = time.Now()

	_, err := r.db.Exec(query,
		category.Name, category.Description, category.Icon, category.Color,
		category.BackgroundColor, category.DisplayOrder, category.IsPublic,
		category.ShowInDashboard, category.IsActive, category.UpdatedAt, category.ID,
	)
	if err != nil {
		return fmt.Errorf("error updating event category: %v", err)
	}

	return nil
}

func (r *EventRepository) DeleteEventCategory(id uuid.UUID) error {
	query := `DELETE FROM event_categories WHERE id = $1`

	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("error deleting event category: %v", err)
	}

	return nil
}

// ==================== EVENTOS ====================

func (r *EventRepository) CreateEvent(event *models.Event) error {
	query := `
		INSERT INTO events (
			id, category_id, name, description, long_description, story_text, icon, color,
			background_color, banner_image, rarity, event_type, event_format, max_participants,
			min_participants, start_date, end_date, registration_start, registration_end,
			duration, level_required, alliance_required, prerequisites, entry_fee,
			entry_currency, event_rules, scoring_system, rewards_config, special_effects,
			status, phase, current_round, total_rounds, total_participants, active_participants,
			completion_rate, average_score, is_repeatable, repeat_interval, next_event_id,
			is_hidden, is_featured, display_order, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31, $32, $33, $34, $35, $36, $37, $38, $39, $40, $41, $42, $43
		)
	`

	now := time.Now()
	if event.ID == uuid.Nil {
		event.ID = uuid.New()
	}
	event.CreatedAt = now
	event.UpdatedAt = now

	_, err := r.db.Exec(query,
		event.ID, event.CategoryID, event.Name, event.Description, event.LongDescription,
		event.StoryText, event.Icon, event.Color, event.BackgroundColor, event.BannerImage,
		event.Rarity, event.EventType, event.EventFormat, event.MaxParticipants,
		event.MinParticipants, event.StartDate, event.EndDate, event.RegistrationStart,
		event.RegistrationEnd, event.Duration, event.LevelRequired, event.AllianceRequired,
		event.Prerequisites, event.EntryFee, event.EntryCurrency, event.EventRules,
		event.ScoringSystem, event.RewardsConfig, event.SpecialEffects, event.Status,
		event.Phase, event.CurrentRound, event.TotalRounds, event.TotalParticipants,
		event.ActiveParticipants, event.CompletionRate, event.AverageScore,
		event.IsRepeatable, event.RepeatInterval, event.NextEventID, event.IsHidden,
		event.IsFeatured, event.DisplayOrder, event.CreatedAt, event.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("error creating event: %v", err)
	}

	return nil
}

func (r *EventRepository) GetEventByID(id uuid.UUID) (*models.Event, error) {
	query := `SELECT * FROM events WHERE id = $1`

	var event models.Event
	err := r.db.QueryRow(query, id).Scan(
		&event.ID, &event.CategoryID, &event.Name, &event.Description, &event.LongDescription,
		&event.StoryText, &event.Icon, &event.Color, &event.BackgroundColor, &event.BannerImage,
		&event.Rarity, &event.EventType, &event.EventFormat, &event.MaxParticipants,
		&event.MinParticipants, &event.StartDate, &event.EndDate, &event.RegistrationStart,
		&event.RegistrationEnd, &event.Duration, &event.LevelRequired, &event.AllianceRequired,
		&event.Prerequisites, &event.EntryFee, &event.EntryCurrency, &event.EventRules,
		&event.ScoringSystem, &event.RewardsConfig, &event.SpecialEffects, &event.Status,
		&event.Phase, &event.CurrentRound, &event.TotalRounds, &event.TotalParticipants,
		&event.ActiveParticipants, &event.CompletionRate, &event.AverageScore,
		&event.IsRepeatable, &event.RepeatInterval, &event.NextEventID, &event.IsHidden,
		&event.IsFeatured, &event.DisplayOrder, &event.CreatedAt, &event.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("error getting event: %v", err)
	}

	return &event, nil
}

func (r *EventRepository) GetEventsByCategory(categoryID uuid.UUID, activeOnly bool, includePast bool) ([]models.Event, error) {
	query := `SELECT * FROM events WHERE category_id = $1`

	if activeOnly {
		query += ` AND status = 'active'`
	}

	if !includePast {
		query += ` AND end_date > NOW()`
	}

	query += ` ORDER BY start_date ASC`

	rows, err := r.db.Query(query, categoryID)
	if err != nil {
		return nil, fmt.Errorf("error getting events by category: %v", err)
	}
	defer rows.Close()

	var events []models.Event
	for rows.Next() {
		var event models.Event
		err := rows.Scan(
			&event.ID, &event.CategoryID, &event.Name, &event.Description, &event.LongDescription,
			&event.StoryText, &event.Icon, &event.Color, &event.BackgroundColor, &event.BannerImage,
			&event.Rarity, &event.EventType, &event.EventFormat, &event.MaxParticipants,
			&event.MinParticipants, &event.StartDate, &event.EndDate, &event.RegistrationStart,
			&event.RegistrationEnd, &event.Duration, &event.LevelRequired, &event.AllianceRequired,
			&event.Prerequisites, &event.EntryFee, &event.EntryCurrency, &event.EventRules,
			&event.ScoringSystem, &event.RewardsConfig, &event.SpecialEffects, &event.Status,
			&event.Phase, &event.CurrentRound, &event.TotalRounds, &event.TotalParticipants,
			&event.ActiveParticipants, &event.CompletionRate, &event.AverageScore,
			&event.IsRepeatable, &event.RepeatInterval, &event.NextEventID, &event.IsHidden,
			&event.IsFeatured, &event.DisplayOrder, &event.CreatedAt, &event.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning event: %v", err)
		}
		events = append(events, event)
	}

	return events, nil
}

func (r *EventRepository) GetActiveEvents() ([]models.Event, error) {
	query := `
		SELECT * FROM events 
		WHERE status = 'active' AND is_hidden = false
		ORDER BY start_date ASC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error getting active events: %v", err)
	}
	defer rows.Close()

	var events []models.Event
	for rows.Next() {
		var event models.Event
		err := rows.Scan(
			&event.ID, &event.CategoryID, &event.Name, &event.Description, &event.LongDescription,
			&event.StoryText, &event.Icon, &event.Color, &event.BackgroundColor, &event.BannerImage,
			&event.Rarity, &event.EventType, &event.EventFormat, &event.MaxParticipants,
			&event.MinParticipants, &event.StartDate, &event.EndDate, &event.RegistrationStart,
			&event.RegistrationEnd, &event.Duration, &event.LevelRequired, &event.AllianceRequired,
			&event.Prerequisites, &event.EntryFee, &event.EntryCurrency, &event.EventRules,
			&event.ScoringSystem, &event.RewardsConfig, &event.SpecialEffects, &event.Status,
			&event.Phase, &event.CurrentRound, &event.TotalRounds, &event.TotalParticipants,
			&event.ActiveParticipants, &event.CompletionRate, &event.AverageScore,
			&event.IsRepeatable, &event.RepeatInterval, &event.NextEventID, &event.IsHidden,
			&event.IsFeatured, &event.DisplayOrder, &event.CreatedAt, &event.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning event: %v", err)
		}
		events = append(events, event)
	}

	return events, nil
}

func (r *EventRepository) GetUpcomingEvents() ([]models.Event, error) {
	query := `
		SELECT * FROM events 
		WHERE status = 'upcoming' 
		ORDER BY start_date ASC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error getting upcoming events: %v", err)
	}
	defer rows.Close()

	var events []models.Event
	for rows.Next() {
		var event models.Event
		err := rows.Scan(
			&event.ID, &event.CategoryID, &event.Name, &event.Description, &event.LongDescription,
			&event.StoryText, &event.Icon, &event.Color, &event.BackgroundColor, &event.BannerImage,
			&event.Rarity, &event.EventType, &event.EventFormat, &event.MaxParticipants,
			&event.MinParticipants, &event.StartDate, &event.EndDate, &event.RegistrationStart,
			&event.RegistrationEnd, &event.Duration, &event.LevelRequired, &event.AllianceRequired,
			&event.Prerequisites, &event.EntryFee, &event.EntryCurrency, &event.EventRules,
			&event.ScoringSystem, &event.RewardsConfig, &event.SpecialEffects, &event.Status,
			&event.Phase, &event.CurrentRound, &event.TotalRounds, &event.TotalParticipants,
			&event.ActiveParticipants, &event.CompletionRate, &event.AverageScore,
			&event.IsRepeatable, &event.RepeatInterval, &event.NextEventID, &event.IsHidden,
			&event.IsFeatured, &event.DisplayOrder, &event.CreatedAt, &event.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning event: %v", err)
		}
		events = append(events, event)
	}

	return events, nil
}

func (r *EventRepository) UpdateEvent(event *models.Event) error {
	query := `
		UPDATE events SET 
			category_id = ?, name = ?, description = ?, long_description = ?, story_text = ?,
			icon = ?, color = ?, background_color = ?, banner_image = ?, rarity = ?,
			event_type = ?, event_format = ?, max_participants = ?, min_participants = ?,
			start_date = ?, end_date = ?, registration_start = ?, registration_end = ?,
			duration = ?, level_required = ?, alliance_required = ?, prerequisites = ?,
			entry_fee = ?, entry_currency = ?, event_rules = ?, scoring_system = ?,
			rewards_config = ?, special_effects = ?, status = ?, phase = ?,
			current_round = ?, total_rounds = ?, total_participants = ?, active_participants = ?,
			completion_rate = ?, average_score = ?, is_repeatable = ?, repeat_interval = ?,
			next_event_id = ?, is_hidden = ?, is_featured = ?, display_order = ?,
			updated_at = ?
		WHERE id = ?
	`

	event.UpdatedAt = time.Now()

	_, err := r.db.Exec(query,
		event.CategoryID, event.Name, event.Description, event.LongDescription,
		event.StoryText, event.Icon, event.Color, event.BackgroundColor, event.BannerImage,
		event.Rarity, event.EventType, event.EventFormat, event.MaxParticipants,
		event.MinParticipants, event.StartDate, event.EndDate, event.RegistrationStart,
		event.RegistrationEnd, event.Duration, event.LevelRequired, event.AllianceRequired,
		event.Prerequisites, event.EntryFee, event.EntryCurrency, event.EventRules,
		event.ScoringSystem, event.RewardsConfig, event.SpecialEffects, event.Status,
		event.Phase, event.CurrentRound, event.TotalRounds, event.TotalParticipants,
		event.ActiveParticipants, event.CompletionRate, event.AverageScore,
		event.IsRepeatable, event.RepeatInterval, event.NextEventID, event.IsHidden,
		event.IsFeatured, event.DisplayOrder, event.UpdatedAt, event.ID,
	)
	if err != nil {
		return fmt.Errorf("error updating event: %v", err)
	}

	return nil
}

func (r *EventRepository) DeleteEvent(id uuid.UUID) error {
	query := `DELETE FROM events WHERE id = $1`

	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("error deleting event: %v", err)
	}

	return nil
}

// ==================== PARTICIPANTES ====================

func (r *EventRepository) CreateEventParticipant(participant *models.EventParticipant) error {
	query := `
		INSERT INTO event_participants (
			event_id, player_id, status, registration_date, entry_fee_paid,
			current_score, total_score, rank, final_rank, matches_played,
			matches_won, matches_lost, matches_drawn, rewards_earned,
			rewards_data, points_earned, last_activity, time_spent,
			joined_at, completed_at, eliminated_at, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	participant.RegistrationDate = now
	participant.JoinedAt = now
	participant.LastActivity = now
	participant.CreatedAt = now

	_, err := r.db.Exec(query,
		participant.EventID, participant.PlayerID, participant.Status,
		participant.RegistrationDate, participant.EntryFeePaid, participant.CurrentScore,
		participant.TotalScore, participant.Rank, participant.FinalRank,
		participant.MatchesPlayed, participant.MatchesWon, participant.MatchesLost,
		participant.MatchesDrawn, participant.RewardsEarned, participant.RewardsData,
		participant.PointsEarned, participant.LastActivity, participant.TimeSpent,
		participant.JoinedAt, participant.CompletedAt, participant.EliminatedAt,
		participant.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("error creating event participant: %v", err)
	}

	return nil
}

func (r *EventRepository) GetEventParticipant(eventID, playerID uuid.UUID) (*models.EventParticipant, error) {
	query := `SELECT * FROM event_participants WHERE event_id = $1 AND player_id = $2`

	var participant models.EventParticipant
	err := r.db.QueryRow(query, eventID, playerID).Scan(
		&participant.ID, &participant.EventID, &participant.PlayerID, &participant.Status,
		&participant.RegistrationDate, &participant.EntryFeePaid, &participant.CurrentScore,
		&participant.TotalScore, &participant.Rank, &participant.FinalRank,
		&participant.MatchesPlayed, &participant.MatchesWon, &participant.MatchesLost,
		&participant.MatchesDrawn, &participant.RewardsEarned, &participant.RewardsData,
		&participant.PointsEarned, &participant.LastActivity, &participant.TimeSpent,
		&participant.JoinedAt, &participant.CompletedAt, &participant.EliminatedAt,
		&participant.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("error getting event participant: %v", err)
	}

	return &participant, nil
}

func (r *EventRepository) GetEventParticipants(eventID uuid.UUID) ([]models.EventParticipant, error) {
	query := `SELECT * FROM event_participants WHERE event_id = $1 ORDER BY rank, total_score DESC`

	rows, err := r.db.Query(query, eventID)
	if err != nil {
		return nil, fmt.Errorf("error getting event participants: %v", err)
	}
	defer rows.Close()

	var participants []models.EventParticipant
	for rows.Next() {
		var participant models.EventParticipant
		err := rows.Scan(
			&participant.ID, &participant.EventID, &participant.PlayerID, &participant.Status,
			&participant.RegistrationDate, &participant.EntryFeePaid, &participant.CurrentScore,
			&participant.TotalScore, &participant.Rank, &participant.FinalRank,
			&participant.MatchesPlayed, &participant.MatchesWon, &participant.MatchesLost,
			&participant.MatchesDrawn, &participant.RewardsEarned, &participant.RewardsData,
			&participant.PointsEarned, &participant.LastActivity, &participant.TimeSpent,
			&participant.JoinedAt, &participant.CompletedAt, &participant.EliminatedAt,
			&participant.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning event participant: %v", err)
		}
		participants = append(participants, participant)
	}

	return participants, nil
}

func (r *EventRepository) UpdateEventParticipant(participant *models.EventParticipant) error {
	query := `
		UPDATE event_participants SET 
			status = ?, current_score = ?, total_score = ?, rank = ?, final_rank = ?,
			matches_played = ?, matches_won = ?, matches_lost = ?, matches_drawn = ?,
			rewards_earned = ?, rewards_data = ?, points_earned = ?, last_activity = ?,
			time_spent = ?, completed_at = ?, eliminated_at = ?
		WHERE event_id = ? AND player_id = ?
	`

	participant.LastActivity = time.Now()

	_, err := r.db.Exec(query,
		participant.Status, participant.CurrentScore, participant.TotalScore,
		participant.Rank, participant.FinalRank, participant.MatchesPlayed,
		participant.MatchesWon, participant.MatchesLost, participant.MatchesDrawn,
		participant.RewardsEarned, participant.RewardsData, participant.PointsEarned,
		participant.LastActivity, participant.TimeSpent, participant.CompletedAt,
		participant.EliminatedAt, participant.EventID, participant.PlayerID,
	)
	if err != nil {
		return fmt.Errorf("error updating event participant: %v", err)
	}

	return nil
}

// ==================== PARTIDAS ====================

func (r *EventRepository) CreateEventMatch(match *models.EventMatch) error {
	query := `
		INSERT INTO event_matches (
			event_id, round, match_number, player1_id, player2_id, winner_id,
			player1_score, player2_score, match_data, status, start_time, end_time,
			duration, match_type, match_rules
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.Exec(query,
		match.EventID, match.Round, match.MatchNumber, match.Player1ID, match.Player2ID,
		match.WinnerID, match.Player1Score, match.Player2Score, match.MatchData,
		match.Status, match.StartTime, match.EndTime, match.Duration,
		match.MatchType, match.MatchRules,
	)
	if err != nil {
		return fmt.Errorf("error creating event match: %v", err)
	}

	return nil
}

func (r *EventRepository) GetEventMatches(eventID uuid.UUID) ([]models.EventMatch, error) {
	query := `SELECT * FROM event_matches WHERE event_id = $1 ORDER BY round, match_number`

	rows, err := r.db.Query(query, eventID)
	if err != nil {
		return nil, fmt.Errorf("error getting event matches: %v", err)
	}
	defer rows.Close()

	var matches []models.EventMatch
	for rows.Next() {
		var match models.EventMatch
		err := rows.Scan(
			&match.ID, &match.EventID, &match.Round, &match.MatchNumber,
			&match.Player1ID, &match.Player2ID, &match.WinnerID, &match.Player1Score,
			&match.Player2Score, &match.MatchData, &match.Status, &match.StartTime,
			&match.EndTime, &match.Duration, &match.MatchType, &match.MatchRules,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning event match: %v", err)
		}
		matches = append(matches, match)
	}

	return matches, nil
}

// ==================== RECOMPENSAS ====================

func (r *EventRepository) GetEventRewards(eventID uuid.UUID) ([]models.EventReward, error) {
	query := `SELECT * FROM event_rewards WHERE event_id = $1 AND is_active = true`

	rows, err := r.db.Query(query, eventID)
	if err != nil {
		return nil, fmt.Errorf("error getting event rewards: %v", err)
	}
	defer rows.Close()

	var rewards []models.EventReward
	for rows.Next() {
		var reward models.EventReward
		err := rows.Scan(
			&reward.ID, &reward.EventID, &reward.Name, &reward.Description,
			&reward.Type, &reward.MinRank, &reward.MaxRank, &reward.MinScore,
			&reward.Quantity, &reward.ResourceType, &reward.ItemID,
			&reward.CurrencyType, &reward.TitleID, &reward.IsActive,
			&reward.CreatedAt, &reward.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning event reward: %v", err)
		}
		rewards = append(rewards, reward)
	}

	return rewards, nil
}

// ==================== RANKING ====================

func (r *EventRepository) GetEventLeaderboard(eventID uuid.UUID) ([]models.EventLeaderboard, error) {
	query := `
		SELECT event_id, player_id, player_name, rank, score, matches_won, matches_lost, win_rate, last_activity
		FROM event_leaderboard 
		WHERE event_id = $1 
		ORDER BY rank, score DESC
	`

	rows, err := r.db.Query(query, eventID)
	if err != nil {
		return nil, fmt.Errorf("error getting event leaderboard: %v", err)
	}
	defer rows.Close()

	var leaderboard []models.EventLeaderboard
	for rows.Next() {
		var entry models.EventLeaderboard
		err := rows.Scan(
			&entry.EventID, &entry.PlayerID, &entry.PlayerName, &entry.Rank,
			&entry.Score, &entry.MatchesWon, &entry.MatchesLost, &entry.WinRate,
			&entry.LastActivity,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning leaderboard entry: %v", err)
		}
		leaderboard = append(leaderboard, entry)
	}

	return leaderboard, nil
}

// ==================== ESTADÍSTICAS ====================

func (r *EventRepository) GetEventStatistics(playerID uuid.UUID) (*models.EventStatistics, error) {
	query := `
		SELECT 
			player_id,
			COUNT(DISTINCT event_id) as total_events_joined,
			SUM(CASE WHEN status = 'active' THEN 1 ELSE 0 END) as active_events_joined,
			SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END) as completed_events,
			SUM(CASE WHEN final_rank = 1 THEN 1 ELSE 0 END) as events_won,
			SUM(matches_played) as total_matches_played,
			SUM(matches_won) as total_matches_won,
			SUM(matches_lost) as total_matches_lost,
			SUM(current_score) as total_score,
			SUM(points_earned) as total_points_earned,
			SUM(time_spent) as total_time_spent,
			SUM(CASE WHEN final_rank = 1 THEN 1 ELSE 0 END) as first_place_finishes,
			SUM(CASE WHEN final_rank <= 3 THEN 1 ELSE 0 END) as top_three_finishes,
			MIN(joined_at) as first_event_date,
			MAX(last_activity) as last_event_date
		FROM event_participants
		WHERE player_id = $1
		GROUP BY player_id
	`

	var stats models.EventStatistics
	err := r.db.QueryRow(query, playerID).Scan(
		&stats.PlayerID, &stats.TotalEventsJoined, &stats.ActiveEventsJoined,
		&stats.CompletedEvents, &stats.EventsWon, &stats.TotalMatchesPlayed,
		&stats.TotalMatchesWon, &stats.TotalMatchesLost, &stats.TotalScore,
		&stats.TotalPointsEarned, &stats.TotalTimeSpent, &stats.FirstPlaceFinishes,
		&stats.TopThreeFinishes, &stats.FirstEventDate, &stats.LastEventDate,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			// El jugador no tiene estadísticas, crear unas básicas
			stats.PlayerID = playerID
			stats.CreatedAt = time.Now()
			stats.UpdatedAt = time.Now()
			return &stats, nil
		}
		return nil, fmt.Errorf("error getting event statistics: %v", err)
	}

	// Calcular estadísticas derivadas
	if stats.TotalMatchesPlayed > 0 {
		stats.WinRate = float64(stats.TotalMatchesWon) / float64(stats.TotalMatchesPlayed)
	}
	if stats.TotalEventsJoined > 0 {
		stats.AverageScore = float64(stats.TotalScore) / float64(stats.TotalEventsJoined)
	}

	stats.CreatedAt = time.Now()
	stats.UpdatedAt = time.Now()

	return &stats, nil
}

func (r *EventRepository) CreateEventStatistics(stats *models.EventStatistics) error {
	query := `
		INSERT INTO event_statistics (
			player_id, total_events_joined, active_events_joined, completed_events,
			events_won, total_matches_played, total_matches_won, total_matches_lost,
			win_rate, total_score, average_score, highest_score, total_rewards_earned,
			total_points_earned, total_time_spent, first_place_finishes,
			top_three_finishes, perfect_scores, first_event_date, last_event_date,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			total_events_joined = VALUES(total_events_joined),
			active_events_joined = VALUES(active_events_joined),
			completed_events = VALUES(completed_events),
			events_won = VALUES(events_won),
			total_matches_played = VALUES(total_matches_played),
			total_matches_won = VALUES(total_matches_won),
			total_matches_lost = VALUES(total_matches_lost),
			win_rate = VALUES(win_rate),
			total_score = VALUES(total_score),
			average_score = VALUES(average_score),
			highest_score = VALUES(highest_score),
			total_rewards_earned = VALUES(total_rewards_earned),
			total_points_earned = VALUES(total_points_earned),
			total_time_spent = VALUES(total_time_spent),
			first_place_finishes = VALUES(first_place_finishes),
			top_three_finishes = VALUES(top_three_finishes),
			perfect_scores = VALUES(perfect_scores),
			first_event_date = VALUES(first_event_date),
			last_event_date = VALUES(last_event_date),
			updated_at = VALUES(updated_at)
	`

	stats.UpdatedAt = time.Now()
	if stats.CreatedAt.IsZero() {
		stats.CreatedAt = time.Now()
	}

	_, err := r.db.Exec(query,
		stats.PlayerID, stats.TotalEventsJoined, stats.ActiveEventsJoined,
		stats.CompletedEvents, stats.EventsWon, stats.TotalMatchesPlayed,
		stats.TotalMatchesWon, stats.TotalMatchesLost, stats.WinRate,
		stats.TotalScore, stats.AverageScore, stats.HighestScore,
		stats.TotalRewardsEarned, stats.TotalPointsEarned, stats.TotalTimeSpent,
		stats.FirstPlaceFinishes, stats.TopThreeFinishes, stats.PerfectScores,
		stats.FirstEventDate, stats.LastEventDate, stats.CreatedAt, stats.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("error creating/updating event statistics: %v", err)
	}

	return nil
}

// ==================== NOTIFICACIONES ====================

func (r *EventRepository) CreateEventNotification(notification *models.EventNotification) error {
	query := `
		INSERT INTO event_notifications (
			player_id, event_id, type, title, message, data, is_read, is_dismissed, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.Exec(query,
		notification.PlayerID, notification.EventID, notification.Type,
		notification.Title, notification.Message, notification.Data,
		notification.IsRead, notification.IsDismissed, notification.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("error creating event notification: %v", err)
	}

	return nil
}

func (r *EventRepository) GetPlayerEventNotifications(playerID uuid.UUID) ([]models.EventNotification, error) {
	query := `
		SELECT * FROM event_notifications 
		WHERE player_id = ? AND is_dismissed = false 
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, playerID)
	if err != nil {
		return nil, fmt.Errorf("error getting player event notifications: %v", err)
	}
	defer rows.Close()

	var notifications []models.EventNotification
	for rows.Next() {
		var notification models.EventNotification
		err := rows.Scan(
			&notification.ID, &notification.PlayerID, &notification.EventID,
			&notification.Type, &notification.Title, &notification.Message,
			&notification.Data, &notification.IsRead, &notification.IsDismissed,
			&notification.CreatedAt, &notification.ReadAt, &notification.DismissedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning notification: %v", err)
		}
		notifications = append(notifications, notification)
	}

	return notifications, nil
}

// ==================== DASHBOARD ====================

func (r *EventRepository) GetEventDashboard(playerID uuid.UUID) (*models.EventDashboard, error) {
	dashboard := &models.EventDashboard{
		LastUpdated: time.Now(),
	}

	// Obtener estadísticas del jugador
	stats, err := r.GetEventStatistics(playerID)
	if err != nil {
		return nil, fmt.Errorf("error getting event statistics: %v", err)
	}
	dashboard.PlayerStats = stats

	// Obtener categorías
	categories, err := r.GetAllEventCategories()
	if err != nil {
		return nil, fmt.Errorf("error getting event categories: %v", err)
	}
	dashboard.Categories = categories

	// Obtener eventos activos
	activeEvents, err := r.GetActiveEvents()
	if err != nil {
		return nil, fmt.Errorf("error getting active events: %v", err)
	}
	dashboard.ActiveEvents = activeEvents

	// Obtener eventos próximos
	upcomingEvents, err := r.GetUpcomingEvents()
	if err != nil {
		return nil, fmt.Errorf("error getting upcoming events: %v", err)
	}
	dashboard.UpcomingEvents = upcomingEvents

	// Obtener notificaciones
	notifications, err := r.GetPlayerEventNotifications(playerID)
	if err != nil {
		return nil, fmt.Errorf("error getting notifications: %v", err)
	}
	dashboard.Notifications = notifications

	// Inicializar calendario
	dashboard.Calendar = make(map[string]interface{})

	return dashboard, nil
}

// ==================== FUNCIONES UTILITARIAS ====================

func (r *EventRepository) GetEventWithDetails(eventID, playerID uuid.UUID) (*models.EventWithDetails, error) {
	details := &models.EventWithDetails{}

	// Obtener el evento
	event, err := r.GetEventByID(eventID)
	if err != nil {
		return nil, fmt.Errorf("error getting event: %v", err)
	}
	details.Event = event

	// Obtener categoría
	category, err := r.GetEventCategory(event.CategoryID)
	if err != nil {
		return nil, fmt.Errorf("error getting event category: %v", err)
	}
	details.Category = category

	// Obtener participantes
	participants, err := r.GetEventParticipants(eventID)
	if err == nil {
		details.Participants = participants
	}

	// Obtener partidas
	matches, err := r.GetEventMatches(eventID)
	if err == nil {
		details.Matches = matches
	}

	// Obtener recompensas
	rewards, err := r.GetEventRewards(eventID)
	if err == nil {
		details.Rewards = rewards
	}

	// Obtener ranking
	leaderboard, err := r.GetEventLeaderboard(eventID)
	if err == nil {
		details.Leaderboard = leaderboard
	}

	return details, nil
}

// Función para registrar un jugador en un evento
func (r *EventRepository) RegisterPlayerForEvent(eventID, playerID uuid.UUID, entryFeePaid bool) error {
	// Verificar si ya está registrado
	existing, err := r.GetEventParticipant(eventID, playerID)
	if err == nil && existing != nil {
		return fmt.Errorf("player already registered for this event")
	}

	// Crear nueva participación
	participant := &models.EventParticipant{
		EventID:       eventID,
		PlayerID:      playerID,
		Status:        "registered",
		EntryFeePaid:  entryFeePaid,
		CurrentScore:  0,
		TotalScore:    0,
		Rank:          0,
		FinalRank:     0,
		MatchesPlayed: 0,
		MatchesWon:    0,
		MatchesLost:   0,
		MatchesDrawn:  0,
		RewardsEarned: false,
		RewardsData:   "{}",
		PointsEarned:  0,
		TimeSpent:     0,
	}

	return r.CreateEventParticipant(participant)
}

// Función para actualizar el progreso de un participante
func (r *EventRepository) UpdateEventProgress(eventID, playerID uuid.UUID, newScore int) (*models.EventProgressUpdate, error) {
	participant, err := r.GetEventParticipant(eventID, playerID)
	if err != nil {
		return nil, fmt.Errorf("error getting event participant: %v", err)
	}

	oldScore := participant.CurrentScore
	scoreChange := newScore - oldScore

	// Actualizar puntuación
	participant.CurrentScore = newScore
	participant.TotalScore += scoreChange
	participant.LastActivity = time.Now()

	// Actualizar en la base de datos
	if err := r.UpdateEventParticipant(participant); err != nil {
		return nil, fmt.Errorf("error updating event participant: %v", err)
	}

	// Calcular cambio de ranking (simplificado)
	oldRank := participant.Rank
	newRank := oldRank // En una implementación real, se recalcularía el ranking

	return &models.EventProgressUpdate{
		EventID:       eventID,
		PlayerID:      playerID,
		OldScore:      oldScore,
		NewScore:      newScore,
		ScoreChange:   scoreChange,
		OldRank:       oldRank,
		NewRank:       newRank,
		RankChange:    newRank - oldRank,
		IsCompleted:   false,
		RewardsEarned: "{}",
		Timestamp:     time.Now(),
	}, nil
}

// GetEventCategories obtiene categorías de eventos
func (r *EventRepository) GetEventCategories(activeOnly bool) ([]models.EventCategory, error) {
	query := `SELECT * FROM event_categories`
	if activeOnly {
		query += ` WHERE is_active = true`
	}
	query += ` ORDER BY display_order, name`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error getting event categories: %v", err)
	}
	defer rows.Close()

	var categories []models.EventCategory
	for rows.Next() {
		var category models.EventCategory
		err := rows.Scan(
			&category.ID, &category.Name, &category.Description, &category.Icon,
			&category.Color, &category.BackgroundColor, &category.DisplayOrder,
			&category.IsPublic, &category.ShowInDashboard, &category.TotalEvents,
			&category.ActiveCount, &category.CompletionRate, &category.IsActive,
			&category.CreatedAt, &category.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning event category: %v", err)
		}
		categories = append(categories, category)
	}

	return categories, nil
}

// GetEvent obtiene un evento específico
func (r *EventRepository) GetEvent(eventID uuid.UUID) (*models.Event, error) {
	return r.GetEventByID(eventID)
}

// GetPlayerEvents obtiene los eventos de un jugador
func (r *EventRepository) GetPlayerEvents(playerID uuid.UUID) ([]models.Event, error) {
	query := `
		SELECT e.* FROM events e
		INNER JOIN event_participants ep ON e.id = ep.event_id
		WHERE ep.player_id = $1
		ORDER BY e.start_date DESC
	`

	rows, err := r.db.Query(query, playerID)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo eventos del jugador: %w", err)
	}
	defer rows.Close()

	var events []models.Event
	for rows.Next() {
		var event models.Event
		err := rows.Scan(
			&event.ID, &event.CategoryID, &event.Name, &event.Description, &event.LongDescription,
			&event.StoryText, &event.Icon, &event.Color, &event.BackgroundColor, &event.BannerImage,
			&event.Rarity, &event.EventType, &event.EventFormat, &event.MaxParticipants,
			&event.MinParticipants, &event.StartDate, &event.EndDate, &event.RegistrationStart,
			&event.RegistrationEnd, &event.Duration, &event.LevelRequired, &event.AllianceRequired,
			&event.Prerequisites, &event.EntryFee, &event.EntryCurrency, &event.EventRules,
			&event.ScoringSystem, &event.RewardsConfig, &event.SpecialEffects, &event.Status,
			&event.Phase, &event.CurrentRound, &event.TotalRounds, &event.TotalParticipants,
			&event.ActiveParticipants, &event.CompletionRate, &event.AverageScore,
			&event.IsRepeatable, &event.RepeatInterval, &event.NextEventID, &event.IsHidden,
			&event.IsFeatured, &event.DisplayOrder, &event.CreatedAt, &event.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando evento: %w", err)
		}
		events = append(events, event)
	}

	return events, nil
}

// GetAllEvents obtiene todos los eventos para el dashboard
func (r *EventRepository) GetAllEvents() ([]models.Event, error) {
	query := `
		SELECT id, category_id, name, description, long_description, story_text, icon, color,
		       background_color, banner_image, rarity, event_type, event_format, max_participants,
		       min_participants, start_date, end_date, registration_start, registration_end,
		       duration, level_required, alliance_required, prerequisites, entry_fee,
		       entry_currency, event_rules, scoring_system, rewards_config, special_effects,
		       status, phase, current_round, total_rounds, total_participants, active_participants,
		       completion_rate, average_score, is_repeatable, repeat_interval, next_event_id,
		       is_hidden, is_featured, display_order, created_at, updated_at
		FROM events
		ORDER BY created_at DESC
		LIMIT 100
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo eventos: %w", err)
	}
	defer rows.Close()

	var events []models.Event
	for rows.Next() {
		var event models.Event
		err := rows.Scan(
			&event.ID, &event.CategoryID, &event.Name, &event.Description, &event.LongDescription,
			&event.StoryText, &event.Icon, &event.Color, &event.BackgroundColor, &event.BannerImage,
			&event.Rarity, &event.EventType, &event.EventFormat, &event.MaxParticipants,
			&event.MinParticipants, &event.StartDate, &event.EndDate, &event.RegistrationStart,
			&event.RegistrationEnd, &event.Duration, &event.LevelRequired, &event.AllianceRequired,
			&event.Prerequisites, &event.EntryFee, &event.EntryCurrency, &event.EventRules,
			&event.ScoringSystem, &event.RewardsConfig, &event.SpecialEffects, &event.Status,
			&event.Phase, &event.CurrentRound, &event.TotalRounds, &event.TotalParticipants,
			&event.ActiveParticipants, &event.CompletionRate, &event.AverageScore,
			&event.IsRepeatable, &event.RepeatInterval, &event.NextEventID, &event.IsHidden,
			&event.IsFeatured, &event.DisplayOrder, &event.CreatedAt, &event.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando evento: %w", err)
		}
		events = append(events, event)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterando eventos: %w", err)
	}

	return events, nil
}
