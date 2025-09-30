package models

import (
	"time"

	"github.com/google/uuid"
)

// AchievementCategory representa una categoría de logros
type AchievementCategory struct {
	ID              uuid.UUID `json:"id" db:"id"`
	Name            string    `json:"name" db:"name"`
	Description     string    `json:"description" db:"description"`
	Icon            string    `json:"icon" db:"icon"`
	Color           string    `json:"color" db:"color"`
	BackgroundColor string    `json:"background_color" db:"background_color"`

	// Configuración
	DisplayOrder    int  `json:"display_order" db:"display_order"`
	IsPublic        bool `json:"is_public" db:"is_public"`
	ShowInDashboard bool `json:"show_in_dashboard" db:"show_in_dashboard"`

	// Estadísticas
	TotalAchievements int     `json:"total_achievements" db:"total_achievements"`
	CompletedCount    int     `json:"completed_count" db:"completed_count"`
	CompletionRate    float64 `json:"completion_rate" db:"completion_rate"`

	// Estado
	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Achievement representa un logro individual
type Achievement struct {
	ID              uuid.UUID `json:"id" db:"id"`
	CategoryID      uuid.UUID `json:"category_id" db:"category_id"`
	Name            string    `json:"name" db:"name"`
	Description     string    `json:"description" db:"description"`
	LongDescription string    `json:"long_description" db:"long_description"`

	// Visualización
	Icon            string `json:"icon" db:"icon"`
	Color           string `json:"color" db:"color"`
	BackgroundColor string `json:"background_color" db:"background_color"`
	Rarity          string `json:"rarity" db:"rarity"` // common, rare, epic, legendary, mythic

	// Progresión
	ProgressType     string `json:"progress_type" db:"progress_type"` // single, cumulative, tiered
	TargetValue      int    `json:"target_value" db:"target_value"`
	CurrentValue     int    `json:"current_value" db:"current_value"`
	ProgressFormula  string `json:"progress_formula" db:"progress_formula"` // JSON con fórmula de progreso
	RequiredProgress int    `json:"required_progress" db:"required_progress"`

	// Niveles (para logros tiered)
	Tiers       string `json:"tiers" db:"tiers"` // JSON con niveles y requisitos
	CurrentTier int    `json:"current_tier" db:"current_tier"`
	MaxTier     int    `json:"max_tier" db:"max_tier"`

	// Recompensas
	RewardsEnabled bool                `json:"rewards_enabled" db:"rewards_enabled"`
	RewardsConfig  string              `json:"rewards_config" db:"rewards_config"` // JSON con configuración de recompensas
	Rewards        []AchievementReward `json:"rewards" db:"-"`

	// Condiciones
	Prerequisites string     `json:"prerequisites" db:"prerequisites"` // JSON con logros requeridos
	TimeLimit     *time.Time `json:"time_limit" db:"time_limit"`
	EventRequired *uuid.UUID `json:"event_required" db:"event_required"`

	// Estadísticas
	Points           int     `json:"points" db:"points"`
	Difficulty       int     `json:"difficulty" db:"difficulty"` // 1-10
	CompletionRate   float64 `json:"completion_rate" db:"completion_rate"`
	TotalCompletions int     `json:"total_completions" db:"total_completions"`

	// Estado
	IsActive     bool `json:"is_active" db:"is_active"`
	IsHidden     bool `json:"is_hidden" db:"is_hidden"`
	IsSecret     bool `json:"is_secret" db:"is_secret"`
	DisplayOrder int  `json:"display_order" db:"display_order"`

	// Fechas
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// PlayerAchievement representa el progreso de un jugador en un logro
type PlayerAchievement struct {
	PlayerID      uuid.UUID `json:"player_id" db:"player_id"`
	AchievementID uuid.UUID `json:"achievement_id" db:"achievement_id"`

	// Progreso
	CurrentProgress int     `json:"current_progress" db:"current_progress"`
	TargetProgress  int     `json:"target_progress" db:"target_progress"`
	ProgressPercent float64 `json:"progress_percent" db:"progress_percent"`

	// Estado
	IsCompleted bool `json:"is_completed" db:"is_completed"`
	IsClaimed   bool `json:"is_claimed" db:"is_claimed"`
	CurrentTier int  `json:"current_tier" db:"current_tier"`

	// Recompensas
	RewardsClaimed bool   `json:"rewards_claimed" db:"rewards_claimed"`
	RewardsData    string `json:"rewards_data" db:"rewards_data"` // JSON con recompensas recibidas

	// Estadísticas
	PointsEarned   int        `json:"points_earned" db:"points_earned"`
	CompletionTime *time.Time `json:"completion_time" db:"completion_time"`

	// Fechas
	StartedAt   time.Time  `json:"started_at" db:"started_at"`
	LastUpdated time.Time  `json:"last_updated" db:"last_updated"`
	CompletedAt *time.Time `json:"completed_at" db:"completed_at"`
	ClaimedAt   *time.Time `json:"claimed_at" db:"claimed_at"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
}

// AchievementProgress representa el progreso detallado de un logro
type AchievementProgress struct {
	ID            uuid.UUID `json:"id" db:"id"`
	PlayerID      uuid.UUID `json:"player_id" db:"player_id"`
	AchievementID uuid.UUID `json:"achievement_id" db:"achievement_id"`

	// Progreso detallado
	CurrentProgress  int    `json:"current_progress" db:"current_progress"`
	RequiredProgress int    `json:"required_progress" db:"required_progress"`
	ProgressData     string `json:"progress_data" db:"progress_data"` // JSON con datos de progreso
	Milestones       string `json:"milestones" db:"milestones"`       // JSON con hitos alcanzados
	Breakdown        string `json:"breakdown" db:"breakdown"`         // JSON con desglose de progreso

	// Metadatos
	LastActivity  time.Time `json:"last_activity" db:"last_activity"`
	ActivityCount int       `json:"activity_count" db:"activity_count"`
	LastUpdated   time.Time `json:"last_updated" db:"last_updated"`

	// Fechas
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// AchievementReward representa una recompensa de logro
type AchievementReward struct {
	ID            uuid.UUID `json:"id" db:"id"`
	AchievementID uuid.UUID `json:"achievement_id" db:"achievement_id"`

	// Configuración de recompensa
	RewardType string `json:"reward_type" db:"reward_type"` // currency, items, title, experience, etc.
	RewardData string `json:"reward_data" db:"reward_data"` // JSON con datos de la recompensa
	Quantity   int    `json:"quantity" db:"quantity"`

	// Condiciones
	TierRequired int  `json:"tier_required" db:"tier_required"` // 0 = cualquier tier
	IsRepeatable bool `json:"is_repeatable" db:"is_repeatable"`

	// Estado
	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// AchievementEvent representa un evento que puede desbloquear logros
type AchievementEvent struct {
	ID        uuid.UUID `json:"id" db:"id"`
	EventType string    `json:"event_type" db:"event_type"` // battle_won, building_upgraded, etc.
	EventData string    `json:"event_data" db:"event_data"` // JSON con datos del evento

	// Entidad relacionada
	EntityType string    `json:"entity_type" db:"entity_type"` // player, alliance, village
	EntityID   uuid.UUID `json:"entity_id" db:"entity_id"`

	// Progreso
	ProgressValue int    `json:"progress_value" db:"progress_value"`
	ProgressData  string `json:"progress_data" db:"progress_data"` // JSON con datos de progreso

	// Metadatos
	Source      string `json:"source" db:"source"` // system, manual, event
	Description string `json:"description" db:"description"`

	// Fechas
	OccurredAt time.Time `json:"occurred_at" db:"occurred_at"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// AchievementStatistics representa estadísticas de logros
type AchievementStatistics struct {
	ID       uuid.UUID `json:"id" db:"id"`
	PlayerID uuid.UUID `json:"player_id" db:"player_id"`

	// Estadísticas generales
	TotalAchievements     int     `json:"total_achievements" db:"total_achievements"`
	CompletedAchievements int     `json:"completed_achievements" db:"completed_achievements"`
	CompletionRate        float64 `json:"completion_rate" db:"completion_rate"`

	// Puntos y progreso
	TotalPoints     int `json:"total_points" db:"total_points"`
	PointsThisWeek  int `json:"points_this_week" db:"points_this_week"`
	PointsThisMonth int `json:"points_this_month" db:"points_this_month"`

	// Categorías
	CategoryStats string `json:"category_stats" db:"category_stats"` // JSON con estadísticas por categoría
	RarityStats   string `json:"rarity_stats" db:"rarity_stats"`     // JSON con estadísticas por rareza

	// Rankings
	GlobalRank   int `json:"global_rank" db:"global_rank"`
	AllianceRank int `json:"alliance_rank" db:"alliance_rank"`
	WorldRank    int `json:"world_rank" db:"world_rank"`

	// Progreso temporal
	LastAchievement *time.Time `json:"last_achievement" db:"last_achievement"`
	StreakDays      int        `json:"streak_days" db:"streak_days"`
	LongestStreak   int        `json:"longest_streak" db:"longest_streak"`

	// Recompensas
	TotalRewardsClaimed int `json:"total_rewards_claimed" db:"total_rewards_claimed"`
	RewardsValue        int `json:"rewards_value" db:"rewards_value"`

	// Campos adicionales faltantes
	AverageDifficulty float64    `json:"average_difficulty" db:"average_difficulty"`
	RarestAchievement string     `json:"rarest_achievement" db:"rarest_achievement"`
	FastestCompletion string     `json:"fastest_completion" db:"fastest_completion"`
	SlowestCompletion string     `json:"slowest_completion" db:"slowest_completion"`
	LastCompletion    *time.Time `json:"last_completion" db:"last_completion"`

	// Fechas
	FirstAchievement time.Time `json:"first_achievement" db:"first_achievement"`
	LastUpdated      time.Time `json:"last_updated" db:"last_updated"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
}

// AchievementLeaderboard representa el ranking de logros
type AchievementLeaderboard struct {
	ID         uuid.UUID `json:"id" db:"id"`
	PlayerID   uuid.UUID `json:"player_id" db:"player_id"`
	PlayerName string    `json:"player_name" db:"player_name"`

	// Puntuación
	TotalPoints    int     `json:"total_points" db:"total_points"`
	CompletedCount int     `json:"completed_count" db:"completed_count"`
	CompletionRate float64 `json:"completion_rate" db:"completion_rate"`

	// Posición
	Position         int `json:"position" db:"position"`
	PreviousPosition int `json:"previous_position" db:"previous_position"`
	PositionChange   int `json:"position_change" db:"position_change"`

	// Campos adicionales faltantes
	TotalAchievements     int `json:"total_achievements" db:"total_achievements"`
	CompletedAchievements int `json:"completed_achievements" db:"completed_achievements"`
	Rank                  int `json:"rank" db:"rank"`

	// Metadatos
	AllianceID   *uuid.UUID `json:"alliance_id" db:"alliance_id"`
	AllianceName string     `json:"alliance_name" db:"alliance_name"`
	WorldID      uuid.UUID  `json:"world_id" db:"world_id"`

	// Fechas
	LastUpdated time.Time `json:"last_updated" db:"last_updated"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// AchievementMilestone representa un hito alcanzado en un logro
type AchievementMilestone struct {
	ID            uuid.UUID `json:"id" db:"id"`
	AchievementID uuid.UUID `json:"achievement_id" db:"achievement_id"`
	PlayerID      uuid.UUID `json:"player_id" db:"player_id"`

	// Hito
	MilestoneType  string `json:"milestone_type" db:"milestone_type"` // progress, tier, completion
	MilestoneValue int    `json:"milestone_value" db:"milestone_value"`
	MilestoneData  string `json:"milestone_data" db:"milestone_data"` // JSON con datos del hito

	// Recompensa del hito
	RewardClaimed bool   `json:"reward_claimed" db:"reward_claimed"`
	RewardData    string `json:"reward_data" db:"reward_data"` // JSON con recompensa del hito

	// Fechas
	AchievedAt time.Time  `json:"achieved_at" db:"achieved_at"`
	ClaimedAt  *time.Time `json:"claimed_at" db:"claimed_at"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
}

// AchievementNotification representa una notificación de logro
type AchievementNotification struct {
	ID            uuid.UUID `json:"id" db:"id"`
	PlayerID      uuid.UUID `json:"player_id" db:"player_id"`
	AchievementID uuid.UUID `json:"achievement_id" db:"achievement_id"`

	// Notificación
	NotificationType string `json:"notification_type" db:"notification_type"` // progress, milestone, completion, reward
	Title            string `json:"title" db:"title"`
	Message          string `json:"message" db:"message"`
	Data             string `json:"data" db:"data"` // JSON con datos adicionales

	// Estado
	IsRead      bool `json:"is_read" db:"is_read"`
	IsDismissed bool `json:"is_dismissed" db:"is_dismissed"`

	// Fechas
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	ReadAt      *time.Time `json:"read_at" db:"read_at"`
	DismissedAt *time.Time `json:"dismissed_at" db:"dismissed_at"`
}

// AchievementWithDetails representa un logro con todos sus detalles
type AchievementWithDetails struct {
	Achievement    *Achievement           `json:"achievement"`
	Category       *AchievementCategory   `json:"category"`
	PlayerProgress *PlayerAchievement     `json:"player_progress"`
	Rewards        []AchievementReward    `json:"rewards"`
	Prerequisites  []Achievement          `json:"prerequisites"`
	Milestones     []AchievementMilestone `json:"milestones"`
	Statistics     *AchievementStatistics `json:"statistics"`
}

// AchievementDashboard representa el dashboard de logros
type AchievementDashboard struct {
	PlayerStats          *AchievementStatistics    `json:"player_stats"`
	Categories           []AchievementCategory     `json:"categories"`
	RecentAchievements   []Achievement             `json:"recent_achievements"`
	UpcomingAchievements []Achievement             `json:"upcoming_achievements"`
	Leaderboard          []AchievementLeaderboard  `json:"leaderboard"`
	Notifications        []AchievementNotification `json:"notifications"`
	GlobalStats          map[string]interface{}    `json:"global_stats"`
	LastUpdated          time.Time                 `json:"last_updated"`
}

// AchievementProgressUpdate representa una actualización de progreso
type AchievementProgressUpdate struct {
	PlayerID       uuid.UUID `json:"player_id"`
	AchievementID  uuid.UUID `json:"achievement_id"`
	OldProgress    int       `json:"old_progress"`
	NewProgress    int       `json:"new_progress"`
	ProgressChange int       `json:"progress_change"`
	IsCompleted    bool      `json:"is_completed"`
	IsMilestone    bool      `json:"is_milestone"`
	MilestoneData  string    `json:"milestone_data"`
	RewardsEarned  string    `json:"rewards_earned"`
	Timestamp      time.Time `json:"timestamp"`
}
