package models

import (
	"time"

	"github.com/google/uuid"
)

// QuestCategory representa una categoría de misiones
type QuestCategory struct {
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
	TotalQuests    int     `json:"total_quests" db:"total_quests"`
	CompletedCount int     `json:"completed_count" db:"completed_count"`
	CompletionRate float64 `json:"completion_rate" db:"completion_rate"`

	// Estado
	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Quest representa una misión individual
type Quest struct {
	ID              uuid.UUID `json:"id" db:"id"`
	CategoryID      uuid.UUID `json:"category_id" db:"category_id"`
	Name            string    `json:"name" db:"name"`
	Description     string    `json:"description" db:"description"`
	LongDescription string    `json:"long_description" db:"long_description"`
	StoryText       string    `json:"story_text" db:"story_text"` // Texto narrativo de la misión

	// Visualización
	Icon            string `json:"icon" db:"icon"`
	Color           string `json:"color" db:"color"`
	BackgroundColor string `json:"background_color" db:"background_color"`
	Rarity          string `json:"rarity" db:"rarity"` // common, rare, epic, legendary, mythic

	// Tipo y Progresión
	QuestType        string `json:"quest_type" db:"quest_type"`       // daily, weekly, story, alliance, exploration, research
	ProgressType     string `json:"progress_type" db:"progress_type"` // single, cumulative, tiered, chain
	TargetValue      int    `json:"target_value" db:"target_value"`
	CurrentValue     int    `json:"current_value" db:"current_value"`
	RequiredProgress int    `json:"required_progress" db:"required_progress"` // Progreso requerido para completar
	ProgressFormula  string `json:"progress_formula" db:"progress_formula"`   // JSON con fórmula de progreso

	// Niveles (para misiones tiered)
	Tiers       string `json:"tiers" db:"tiers"` // JSON con niveles y requisitos
	CurrentTier int    `json:"current_tier" db:"current_tier"`
	MaxTier     int    `json:"max_tier" db:"max_tier"`

	// Cadena de misiones
	ChainID         *uuid.UUID `json:"chain_id" db:"chain_id"`
	ChainOrder      int        `json:"chain_order" db:"chain_order"`
	NextQuestID     *uuid.UUID `json:"next_quest_id" db:"next_quest_id"`
	PreviousQuestID *uuid.UUID `json:"previous_quest_id" db:"previous_quest_id"`

	// Recompensas
	RewardsEnabled bool   `json:"rewards_enabled" db:"rewards_enabled"`
	RewardsConfig  string `json:"rewards_config" db:"rewards_config"` // JSON con configuración de recompensas

	// Condiciones
	Prerequisites    string     `json:"prerequisites" db:"prerequisites"` // JSON con misiones requeridas
	LevelRequired    int        `json:"level_required" db:"level_required"`
	AllianceRequired *uuid.UUID `json:"alliance_required" db:"alliance_required"`
	TimeLimit        *time.Time `json:"time_limit" db:"time_limit"`
	EventRequired    *uuid.UUID `json:"event_required" db:"event_required"`

	// Repetición
	IsRepeatable   bool   `json:"is_repeatable" db:"is_repeatable"`
	RepeatInterval string `json:"repeat_interval" db:"repeat_interval"` // daily, weekly, monthly
	MaxCompletions int    `json:"max_completions" db:"max_completions"` // 0 = infinito

	// Estadísticas
	Points           int     `json:"points" db:"points"`
	Difficulty       int     `json:"difficulty" db:"difficulty"` // 1-10
	CompletionRate   float64 `json:"completion_rate" db:"completion_rate"`
	TotalCompletions int     `json:"total_completions" db:"total_completions"`
	AverageTime      int     `json:"average_time" db:"average_time"` // en minutos

	// Estado
	IsActive     bool `json:"is_active" db:"is_active"`
	IsHidden     bool `json:"is_hidden" db:"is_hidden"`
	IsSecret     bool `json:"is_secret" db:"is_secret"`
	DisplayOrder int  `json:"display_order" db:"display_order"`

	// Fechas
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// PlayerQuest representa el progreso de un jugador en una misión
type PlayerQuest struct {
	ID       uuid.UUID `json:"id" db:"id"`
	PlayerID uuid.UUID `json:"player_id" db:"player_id"`
	QuestID  uuid.UUID `json:"quest_id" db:"quest_id"`

	// Progreso
	CurrentProgress int     `json:"current_progress" db:"current_progress"`
	TargetProgress  int     `json:"target_progress" db:"target_progress"`
	ProgressPercent float64 `json:"progress_percent" db:"progress_percent"`

	// Estado
	IsCompleted     bool `json:"is_completed" db:"is_completed"`
	IsClaimed       bool `json:"is_claimed" db:"is_claimed"`
	IsFailed        bool `json:"is_failed" db:"is_failed"`
	CurrentTier     int  `json:"current_tier" db:"current_tier"`
	CompletionCount int  `json:"completion_count" db:"completion_count"`

	// Recompensas
	RewardsClaimed bool   `json:"rewards_claimed" db:"rewards_claimed"`
	RewardsData    string `json:"rewards_data" db:"rewards_data"` // JSON con recompensas recibidas

	// Estadísticas
	PointsEarned   int        `json:"points_earned" db:"points_earned"`
	CompletionTime *time.Time `json:"completion_time" db:"completion_time"`
	TimeSpent      int        `json:"time_spent" db:"time_spent"` // en minutos

	// Fechas
	StartedAt   time.Time  `json:"started_at" db:"started_at"`
	LastUpdated time.Time  `json:"last_updated" db:"last_updated"`
	CompletedAt *time.Time `json:"completed_at" db:"completed_at"`
	ClaimedAt   *time.Time `json:"claimed_at" db:"claimed_at"`
	FailedAt    *time.Time `json:"failed_at" db:"failed_at"`
	ExpiresAt   *time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
}

// QuestProgress representa el progreso detallado de una misión
type QuestProgress struct {
	ID       uuid.UUID `json:"id" db:"id"`
	PlayerID uuid.UUID `json:"player_id" db:"player_id"`
	QuestID  uuid.UUID `json:"quest_id" db:"quest_id"`

	// Progreso detallado
	ProgressData  string `json:"progress_data" db:"progress_data"`   // JSON con datos de progreso
	Milestones    string `json:"milestones" db:"milestones"`         // JSON con hitos alcanzados
	Breakdown     string `json:"breakdown" db:"breakdown"`           // JSON con desglose de progreso
	StoryProgress string `json:"story_progress" db:"story_progress"` // JSON con progreso narrativo

	// Metadatos
	LastActivity  time.Time `json:"last_activity" db:"last_activity"`
	ActivityCount int       `json:"activity_count" db:"activity_count"`
	StoryChoices  string    `json:"story_choices" db:"story_choices"` // JSON con decisiones del jugador

	// Fechas
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// QuestReward representa una recompensa de misión
type QuestReward struct {
	ID      uuid.UUID `json:"id" db:"id"`
	QuestID uuid.UUID `json:"quest_id" db:"quest_id"`

	// Configuración de recompensa
	RewardType string `json:"reward_type" db:"reward_type"` // currency, items, title, experience, resources, etc.
	RewardData string `json:"reward_data" db:"reward_data"` // JSON con datos de la recompensa
	Quantity   int    `json:"quantity" db:"quantity"`

	// Condiciones
	TierRequired int  `json:"tier_required" db:"tier_required"` // 0 = cualquier tier
	IsRepeatable bool `json:"is_repeatable" db:"is_repeatable"`
	IsGuaranteed bool `json:"is_guaranteed" db:"is_guaranteed"` // recompensa garantizada o aleatoria

	// Estado
	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// QuestChain representa una cadena de misiones
type QuestChain struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	StoryArc    string    `json:"story_arc" db:"story_arc"` // JSON con arco narrativo

	// Configuración
	CategoryID   uuid.UUID `json:"category_id" db:"category_id"`
	DisplayOrder int       `json:"display_order" db:"display_order"`
	IsActive     bool      `json:"is_active" db:"is_active"`

	// Progresión
	TotalQuests     int     `json:"total_quests" db:"total_quests"`
	CompletedQuests int     `json:"completed_quests" db:"completed_quests"`
	ProgressPercent float64 `json:"progress_percent" db:"progress_percent"`

	// Recompensas de cadena
	ChainRewards string `json:"chain_rewards" db:"chain_rewards"` // JSON con recompensas por completar la cadena

	// Estado
	IsRepeatable   bool   `json:"is_repeatable" db:"is_repeatable"`
	RepeatInterval string `json:"repeat_interval" db:"repeat_interval"`

	// Fechas
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// QuestStatistics representa las estadísticas de misiones de un jugador
type QuestStatistics struct {
	ID       uuid.UUID `json:"id" db:"id"`
	PlayerID uuid.UUID `json:"player_id" db:"player_id"`

	// Estadísticas generales
	TotalQuests     int     `json:"total_quests" db:"total_quests"`
	CompletedQuests int     `json:"completed_quests" db:"completed_quests"`
	FailedQuests    int     `json:"failed_quests" db:"failed_quests"`
	CompletionRate  float64 `json:"completion_rate" db:"completion_rate"`

	// Puntos y progreso
	TotalPoints     int `json:"total_points" db:"total_points"`
	PointsThisWeek  int `json:"points_this_week" db:"points_this_week"`
	PointsThisMonth int `json:"points_this_month" db:"points_this_month"`

	// Categorías
	CategoryStats string `json:"category_stats" db:"category_stats"` // JSON con estadísticas por categoría
	TypeStats     string `json:"type_stats" db:"type_stats"`         // JSON con estadísticas por tipo

	// Cadenas de misiones
	ChainsCompleted int    `json:"chains_completed" db:"chains_completed"`
	CurrentChains   string `json:"current_chains" db:"current_chains"` // JSON con cadenas en progreso

	// Actividad
	LastQuest     *time.Time `json:"last_quest" db:"last_quest"`
	StreakDays    int        `json:"streak_days" db:"streak_days"`
	LongestStreak int        `json:"longest_streak" db:"longest_streak"`
	AverageTime   int        `json:"average_time" db:"average_time"` // tiempo promedio en minutos

	// Recompensas
	TotalRewardsClaimed int `json:"total_rewards_claimed" db:"total_rewards_claimed"`
	RewardsValue        int `json:"rewards_value" db:"rewards_value"`

	// Campos adicionales para compatibilidad con el repositorio
	AverageDifficulty float64    `json:"average_difficulty" db:"average_difficulty"`
	FastestCompletion *time.Time `json:"fastest_completion" db:"fastest_completion"`
	SlowestCompletion *time.Time `json:"slowest_completion" db:"slowest_completion"`
	LastCompletion    *time.Time `json:"last_completion" db:"last_completion"`

	// Fechas
	FirstQuest  time.Time `json:"first_quest" db:"first_quest"`
	LastUpdated time.Time `json:"last_updated" db:"last_updated"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// QuestNotification representa una notificación de misión
type QuestNotification struct {
	ID       uuid.UUID `json:"id" db:"id"`
	PlayerID uuid.UUID `json:"player_id" db:"player_id"`
	QuestID  uuid.UUID `json:"quest_id" db:"quest_id"`

	// Notificación
	Type    string `json:"type" db:"type"` // progress, milestone, completion, reward, failure, expiration
	Title   string `json:"title" db:"title"`
	Message string `json:"message" db:"message"`
	Data    string `json:"data" db:"data"` // JSON con datos adicionales

	// Estado
	IsRead      bool `json:"is_read" db:"is_read"`
	IsDismissed bool `json:"is_dismissed" db:"is_dismissed"`

	// Fechas
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	ReadAt      *time.Time `json:"read_at" db:"read_at"`
	DismissedAt *time.Time `json:"dismissed_at" db:"dismissed_at"`
}

// QuestMilestone representa un hito de una misión
type QuestMilestone struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	QuestID     uuid.UUID  `json:"quest_id" db:"quest_id"`
	Name        string     `json:"name" db:"name"`
	Description string     `json:"description" db:"description"`
	Progress    int        `json:"progress" db:"progress"` // Progreso requerido para este hito
	Rewards     string     `json:"rewards" db:"rewards"`   // JSON con recompensas del hito
	IsCompleted bool       `json:"is_completed" db:"is_completed"`
	CompletedAt *time.Time `json:"completed_at" db:"completed_at"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
}

// QuestWithDetails representa una misión con todos sus detalles
type QuestWithDetails struct {
	Quest          *Quest           `json:"quest"`
	Category       *QuestCategory   `json:"category"`
	PlayerProgress *PlayerQuest     `json:"player_progress"`
	Rewards        []QuestReward    `json:"rewards"`
	Prerequisites  []Quest          `json:"prerequisites"`
	NextQuests     []Quest          `json:"next_quests"`
	Chain          *QuestChain      `json:"chain"`
	Statistics     *QuestStatistics `json:"statistics"`
	Milestones     []QuestMilestone `json:"milestones"`
}

// QuestDashboard representa el dashboard de misiones de un jugador
type QuestDashboard struct {
	PlayerStats     *QuestStatistics       `json:"player_stats"`
	Categories      []QuestCategory        `json:"categories"`
	ActiveQuests    []Quest                `json:"active_quests"`
	AvailableQuests []Quest                `json:"available_quests"`
	CompletedQuests []Quest                `json:"completed_quests"`
	QuestChains     []QuestChain           `json:"quest_chains"`
	Notifications   []QuestNotification    `json:"notifications"`
	DailyProgress   map[string]interface{} `json:"daily_progress"`
	WeeklyProgress  map[string]interface{} `json:"weekly_progress"`
	LastUpdated     time.Time              `json:"last_updated"`
}

// QuestProgressUpdate representa una actualización de progreso de misión
type QuestProgressUpdate struct {
	PlayerID       uuid.UUID `json:"player_id"`
	QuestID        uuid.UUID `json:"quest_id"`
	OldProgress    int       `json:"old_progress"`
	NewProgress    int       `json:"new_progress"`
	ProgressChange int       `json:"progress_change"`
	IsCompleted    bool      `json:"is_completed"`
	IsFailed       bool      `json:"is_failed"`
	IsMilestone    bool      `json:"is_milestone"`
	MilestoneData  string    `json:"milestone_data"`
	RewardsEarned  string    `json:"rewards_earned"`
	StoryProgress  string    `json:"story_progress"`
	Timestamp      time.Time `json:"timestamp"`
}

// QuestDetails representa los detalles completos de una misión
type QuestDetails struct {
	Quest        *Quest              `json:"quest"`
	Rewards      []*QuestReward      `json:"rewards"`
	Requirements []*QuestRequirement `json:"requirements"`
}

// QuestRequirement representa un requisito de misión
type QuestRequirement struct {
	ID      string                 `json:"id"`
	QuestID string                 `json:"quest_id"`
	Type    string                 `json:"type"`
	Value   int                    `json:"value"`
	Data    map[string]interface{} `json:"data,omitempty"`
}
