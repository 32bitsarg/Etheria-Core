package models

import (
	"time"

	"github.com/google/uuid"
)

// EventCategory representa una categoría de eventos
type EventCategory struct {
	ID              uuid.UUID `json:"id" db:"id"`
	Name            string    `json:"name" db:"name"`
	Description     string    `json:"description" db:"description"`
	Icon            string    `json:"icon" db:"icon"`
	Color           string    `json:"color" db:"color"`
	BackgroundColor string    `json:"background_color" db:"background_color"`

	// Configuración
	DisplayOrder    int  `json:"display_order" db:"display_order"`
	IsPublic        bool `json:"is_public" db:"is_public"`
	ShowInCalendar  bool `json:"show_in_calendar" db:"show_in_calendar"`
	ShowInDashboard bool `json:"show_in_dashboard" db:"show_in_dashboard"`

	// Estadísticas
	TotalEvents       int     `json:"total_events" db:"total_events"`
	ActiveEvents      int     `json:"active_events" db:"active_events"`
	ActiveCount       int     `json:"active_count" db:"active_count"`
	TotalParticipants int     `json:"total_participants" db:"total_participants"`
	CompletionRate    float64 `json:"completion_rate" db:"completion_rate"`

	// Estado
	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Event representa un evento individual
type Event struct {
	ID              uuid.UUID `json:"id" db:"id"`
	CategoryID      uuid.UUID `json:"category_id" db:"category_id"`
	Name            string    `json:"name" db:"name"`
	Description     string    `json:"description" db:"description"`
	LongDescription string    `json:"long_description" db:"long_description"`
	StoryText       string    `json:"story_text" db:"story_text"` // Narrativa del evento

	// Visualización
	Icon            string `json:"icon" db:"icon"`
	Color           string `json:"color" db:"color"`
	BackgroundColor string `json:"background_color" db:"background_color"`
	BannerImage     string `json:"banner_image" db:"banner_image"`
	Rarity          string `json:"rarity" db:"rarity"` // common, rare, epic, legendary, mythic

	// Tipo y Configuración
	EventType       string `json:"event_type" db:"event_type"`             // tournament, seasonal, special, alliance, world, limited
	EventFormat     string `json:"event_format" db:"event_format"`         // single_elimination, round_robin, points_based, cooperative
	MaxParticipants int    `json:"max_participants" db:"max_participants"` // 0 = sin límite
	MinParticipants int    `json:"min_participants" db:"min_participants"`

	// Fechas y Duración
	StartDate         time.Time  `json:"start_date" db:"start_date"`
	EndDate           time.Time  `json:"end_date" db:"end_date"`
	RegistrationStart *time.Time `json:"registration_start" db:"registration_start"`
	RegistrationEnd   *time.Time `json:"registration_end" db:"registration_end"`
	Duration          int        `json:"duration" db:"duration"` // en minutos

	// Condiciones de Participación
	LevelRequired    int        `json:"level_required" db:"level_required"`
	AllianceRequired *uuid.UUID `json:"alliance_required" db:"alliance_required"`
	Prerequisites    string     `json:"prerequisites" db:"prerequisites"`   // JSON con requisitos
	EntryFee         int        `json:"entry_fee" db:"entry_fee"`           // costo de entrada
	EntryCurrency    string     `json:"entry_currency" db:"entry_currency"` // silver, gold, etc.

	// Mecánicas del Evento
	EventRules     string `json:"event_rules" db:"event_rules"`         // JSON con reglas
	ScoringSystem  string `json:"scoring_system" db:"scoring_system"`   // JSON con sistema de puntuación
	RewardsConfig  string `json:"rewards_config" db:"rewards_config"`   // JSON con configuración de recompensas
	SpecialEffects string `json:"special_effects" db:"special_effects"` // JSON con efectos especiales

	// Estado del Evento
	Status       string `json:"status" db:"status"` // upcoming, active, completed, cancelled
	Phase        string `json:"phase" db:"phase"`   // registration, preparation, active, results, rewards
	CurrentRound int    `json:"current_round" db:"current_round"`
	TotalRounds  int    `json:"total_rounds" db:"total_rounds"`

	// Estadísticas
	TotalParticipants  int     `json:"total_participants" db:"total_participants"`
	ActiveParticipants int     `json:"active_participants" db:"active_participants"`
	CompletionRate     float64 `json:"completion_rate" db:"completion_rate"`
	AverageScore       float64 `json:"average_score" db:"average_score"`

	// Repetición
	IsRepeatable   bool       `json:"is_repeatable" db:"is_repeatable"`
	RepeatInterval string     `json:"repeat_interval" db:"repeat_interval"` // daily, weekly, monthly, yearly
	NextEventID    *uuid.UUID `json:"next_event_id" db:"next_event_id"`

	// Estado
	IsHidden     bool `json:"is_hidden" db:"is_hidden"`
	IsFeatured   bool `json:"is_featured" db:"is_featured"`
	DisplayOrder int  `json:"display_order" db:"display_order"`

	// Fechas
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// EventParticipant representa la participación de un jugador en un evento
type EventParticipant struct {
	ID       uuid.UUID `json:"id" db:"id"`
	EventID  uuid.UUID `json:"event_id" db:"event_id"`
	PlayerID uuid.UUID `json:"player_id" db:"player_id"`

	// Estado de Participación
	Status           string    `json:"status" db:"status"` // registered, active, eliminated, completed, disqualified
	RegistrationDate time.Time `json:"registration_date" db:"registration_date"`
	EntryFeePaid     bool      `json:"entry_fee_paid" db:"entry_fee_paid"`

	// Progreso y Rendimiento
	CurrentScore int `json:"current_score" db:"current_score"`
	TotalScore   int `json:"total_score" db:"total_score"`
	Rank         int `json:"rank" db:"rank"`
	FinalRank    int `json:"final_rank" db:"final_rank"`

	// Estadísticas Detalladas
	MatchesPlayed int `json:"matches_played" db:"matches_played"`
	MatchesWon    int `json:"matches_won" db:"matches_won"`
	MatchesLost   int `json:"matches_lost" db:"matches_lost"`
	MatchesDrawn  int `json:"matches_drawn" db:"matches_drawn"`

	// Recompensas
	RewardsEarned bool   `json:"rewards_earned" db:"rewards_earned"`
	RewardsData   string `json:"rewards_data" db:"rewards_data"` // JSON con recompensas
	PointsEarned  int    `json:"points_earned" db:"points_earned"`

	// Actividad
	LastActivity time.Time `json:"last_activity" db:"last_activity"`
	TimeSpent    int       `json:"time_spent" db:"time_spent"` // en minutos

	// Fechas
	JoinedAt     time.Time  `json:"joined_at" db:"joined_at"`
	CompletedAt  *time.Time `json:"completed_at" db:"completed_at"`
	EliminatedAt *time.Time `json:"eliminated_at" db:"eliminated_at"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
}

// EventMatch representa una partida/encuentro en un evento
type EventMatch struct {
	ID          uuid.UUID `json:"id" db:"id"`
	EventID     uuid.UUID `json:"event_id" db:"event_id"`
	Round       int       `json:"round" db:"round"`
	MatchNumber int       `json:"match_number" db:"match_number"`

	// Participantes
	Player1ID uuid.UUID  `json:"player1_id" db:"player1_id"`
	Player2ID uuid.UUID  `json:"player2_id" db:"player2_id"`
	WinnerID  *uuid.UUID `json:"winner_id" db:"winner_id"`

	// Resultados
	Player1Score int    `json:"player1_score" db:"player1_score"`
	Player2Score int    `json:"player2_score" db:"player2_score"`
	MatchData    string `json:"match_data" db:"match_data"` // JSON con datos del partido

	// Estado
	Status    string     `json:"status" db:"status"` // scheduled, in_progress, completed, cancelled
	StartTime *time.Time `json:"start_time" db:"start_time"`
	EndTime   *time.Time `json:"end_time" db:"end_time"`
	Duration  int        `json:"duration" db:"duration"` // en minutos

	// Configuración
	MatchType  string `json:"match_type" db:"match_type"` // battle, race, puzzle, etc.
	MatchRules string `json:"match_rules" db:"match_rules"`
}

// EventDashboard representa el dashboard completo de eventos para un jugador
type EventDashboard struct {
	// Eventos
	ActiveEvents   []Event `json:"active_events"`
	UpcomingEvents []Event `json:"upcoming_events"`
	PlayerEvents   []Event `json:"player_events"` // Eventos del jugador

	// Categorías
	Categories []EventCategory `json:"categories"`

	// Estadísticas del jugador
	PlayerStats *EventStatistics `json:"player_stats"`

	// Notificaciones
	Notifications []EventNotification `json:"notifications"`

	// Calendario
	Calendar map[string]interface{} `json:"calendar"`

	// Metadatos
	LastUpdated time.Time `json:"last_updated"`
}

// EventStatistics representa las estadísticas de eventos de un jugador
type EventStatistics struct {
	PlayerID uuid.UUID `json:"player_id"`

	// Participación
	TotalEventsJoined  int `json:"total_events_joined"`
	ActiveEventsJoined int `json:"active_events_joined"`
	CompletedEvents    int `json:"completed_events"`
	EventsWon          int `json:"events_won"`

	// Rendimiento
	TotalMatchesPlayed int     `json:"total_matches_played"`
	TotalMatchesWon    int     `json:"total_matches_won"`
	TotalMatchesLost   int     `json:"total_matches_lost"`
	WinRate            float64 `json:"win_rate"`

	// Puntuación
	TotalScore   int     `json:"total_score"`
	AverageScore float64 `json:"average_score"`
	HighestScore int     `json:"highest_score"`

	// Recompensas
	TotalRewardsEarned int `json:"total_rewards_earned"`
	TotalPointsEarned  int `json:"total_points_earned"`

	// Tiempo
	TotalTimeSpent int `json:"total_time_spent"` // en minutos

	// Logros
	FirstPlaceFinishes int `json:"first_place_finishes"`
	TopThreeFinishes   int `json:"top_three_finishes"`
	PerfectScores      int `json:"perfect_scores"`

	// Fechas
	FirstEventDate *time.Time `json:"first_event_date"`
	LastEventDate  *time.Time `json:"last_event_date"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// EventReward representa una recompensa de evento
type EventReward struct {
	ID          uuid.UUID `json:"id" db:"id"`
	EventID     uuid.UUID `json:"event_id" db:"event_id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Type        string    `json:"type" db:"type"` // resource, experience, item, currency, title

	// Configuración
	MinRank  int `json:"min_rank" db:"min_rank"`
	MaxRank  int `json:"max_rank" db:"max_rank"`
	MinScore int `json:"min_score" db:"min_score"`
	Quantity int `json:"quantity" db:"quantity"`

	// Datos específicos del tipo
	ResourceType string     `json:"resource_type" db:"resource_type"`
	ItemID       *uuid.UUID `json:"item_id" db:"item_id"`
	CurrencyType string     `json:"currency_type" db:"currency_type"`
	TitleID      *uuid.UUID `json:"title_id" db:"title_id"`

	// Estado
	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// EventLeaderboard representa el ranking de un evento
type EventLeaderboard struct {
	EventID      uuid.UUID `json:"event_id" db:"event_id"`
	PlayerID     uuid.UUID `json:"player_id" db:"player_id"`
	PlayerName   string    `json:"player_name" db:"player_name"`
	Rank         int       `json:"rank" db:"rank"`
	Score        int       `json:"score" db:"score"`
	MatchesWon   int       `json:"matches_won" db:"matches_won"`
	MatchesLost  int       `json:"matches_lost" db:"matches_lost"`
	WinRate      float64   `json:"win_rate" db:"win_rate"`
	LastActivity time.Time `json:"last_activity" db:"last_activity"`
}

// EventNotification representa una notificación de evento
type EventNotification struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	PlayerID    uuid.UUID  `json:"player_id" db:"player_id"`
	EventID     uuid.UUID  `json:"event_id" db:"event_id"`
	Type        string     `json:"type" db:"type"`
	Title       string     `json:"title" db:"title"`
	Message     string     `json:"message" db:"message"`
	Data        string     `json:"data" db:"data"` // JSON con datos adicionales
	IsRead      bool       `json:"is_read" db:"is_read"`
	IsDismissed bool       `json:"is_dismissed" db:"is_dismissed"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	ReadAt      *time.Time `json:"read_at" db:"read_at"`
	DismissedAt *time.Time `json:"dismissed_at" db:"dismissed_at"`
}

// EventWithDetails representa un evento con todos sus detalles
type EventWithDetails struct {
	Event        *Event             `json:"event"`
	Category     *EventCategory     `json:"category"`
	Participants []EventParticipant `json:"participants"`
	Matches      []EventMatch       `json:"matches"`
	Rewards      []EventReward      `json:"rewards"`
	Leaderboard  []EventLeaderboard `json:"leaderboard"`
}

// EventProgressUpdate representa una actualización de progreso de evento
type EventProgressUpdate struct {
	EventID       uuid.UUID `json:"event_id"`
	PlayerID      uuid.UUID `json:"player_id"`
	OldScore      int       `json:"old_score"`
	NewScore      int       `json:"new_score"`
	ScoreChange   int       `json:"score_change"`
	OldRank       int       `json:"old_rank"`
	NewRank       int       `json:"new_rank"`
	RankChange    int       `json:"rank_change"`
	IsCompleted   bool      `json:"is_completed"`
	RewardsEarned string    `json:"rewards_earned"`
	Timestamp     time.Time `json:"timestamp"`
}
