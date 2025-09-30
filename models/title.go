package models

import (
	"time"

	"github.com/google/uuid"
)

// TitleCategory representa una categoría de títulos
type TitleCategory struct {
	ID              uuid.UUID `json:"id" db:"id"`
	Name            string    `json:"name" db:"name"`
	Description     string    `json:"description" db:"description"`
	Icon            string    `json:"icon" db:"icon"`
	Color           string    `json:"color" db:"color"`
	BackgroundColor string    `json:"background_color" db:"background_color"`

	// Configuración
	DisplayOrder    int  `json:"display_order" db:"display_order"`
	IsPublic        bool `json:"is_public" db:"is_public"`
	ShowInProfile   bool `json:"show_in_profile" db:"show_in_profile"`
	ShowInDashboard bool `json:"show_in_dashboard" db:"show_in_dashboard"`

	// Estadísticas
	TotalTitles    int     `json:"total_titles" db:"total_titles"`
	UnlockedTitles int     `json:"unlocked_titles" db:"unlocked_titles"`
	UnlockedCount  int     `json:"unlocked_count" db:"unlocked_count"`
	TotalPlayers   int     `json:"total_players" db:"total_players"`
	CompletionRate float64 `json:"completion_rate" db:"completion_rate"`

	// Estado
	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Title representa un título individual
type Title struct {
	ID              uuid.UUID `json:"id" db:"id"`
	CategoryID      uuid.UUID `json:"category_id" db:"category_id"`
	Name            string    `json:"name" db:"name"`
	Description     string    `json:"description" db:"description"`
	LongDescription string    `json:"long_description" db:"long_description"`
	StoryText       string    `json:"story_text" db:"story_text"` // Narrativa del título

	// Visualización
	Icon            string `json:"icon" db:"icon"`
	Color           string `json:"color" db:"color"`
	BackgroundColor string `json:"background_color" db:"background_color"`
	BorderColor     string `json:"border_color" db:"border_color"`
	Rarity          string `json:"rarity" db:"rarity"` // common, rare, epic, legendary, mythic, divine

	// Tipo y Configuración
	TitleType     string `json:"title_type" db:"title_type"`         // achievement, event, social, military, economic, cultural, seasonal, special
	TitleFormat   string `json:"title_format" db:"title_format"`     // prefix, suffix, standalone, dynamic
	DisplayFormat string `json:"display_format" db:"display_format"` // "{title} {name}", "{name} {title}", etc.

	// Requisitos
	LevelRequired    int        `json:"level_required" db:"level_required"`
	PrestigeRequired int        `json:"prestige_required" db:"prestige_required"`
	AllianceRequired *uuid.UUID `json:"alliance_required" db:"alliance_required"`
	Prerequisites    string     `json:"prerequisites" db:"prerequisites"`         // JSON con requisitos específicos
	UnlockConditions string     `json:"unlock_conditions" db:"unlock_conditions"` // JSON con condiciones de desbloqueo

	// Efectos y Bonificaciones
	Effects          string `json:"effects" db:"effects"`                     // JSON con efectos del título
	Bonuses          string `json:"bonuses" db:"bonuses"`                     // JSON con bonificaciones
	SpecialAbilities string `json:"special_abilities" db:"special_abilities"` // JSON con habilidades especiales

	// Prestigio y Reputación
	PrestigeValue   int    `json:"prestige_value" db:"prestige_value"`     // Valor de prestigio que otorga
	ReputationBonus int    `json:"reputation_bonus" db:"reputation_bonus"` // Bonus de reputación
	SocialStatus    string `json:"social_status" db:"social_status"`       // noble, commoner, outlaw, etc.

	// Limitaciones
	MaxOwners   int  `json:"max_owners" db:"max_owners"`     // 0 = sin límite
	TimeLimit   int  `json:"time_limit" db:"time_limit"`     // en días, 0 = permanente
	IsExclusive bool `json:"is_exclusive" db:"is_exclusive"` // solo un jugador puede tenerlo
	IsTemporary bool `json:"is_temporary" db:"is_temporary"` // título temporal

	// Estado del Título
	Status     string     `json:"status" db:"status"` // available, locked, retired, seasonal
	UnlockDate *time.Time `json:"unlock_date" db:"unlock_date"`
	RetireDate *time.Time `json:"retire_date" db:"retire_date"`

	// Estadísticas
	TotalUnlocked int     `json:"total_unlocked" db:"total_unlocked"`
	CurrentOwners int     `json:"current_owners" db:"current_owners"`
	UnlockRate    float64 `json:"unlock_rate" db:"unlock_rate"` // porcentaje de jugadores que lo tienen

	// Repetición
	IsRepeatable   bool       `json:"is_repeatable" db:"is_repeatable"`
	RepeatInterval string     `json:"repeat_interval" db:"repeat_interval"` // daily, weekly, monthly, yearly
	NextUnlockDate *time.Time `json:"next_unlock_date" db:"next_unlock_date"`

	// Estado
	IsActive     bool `json:"is_active" db:"is_active"`
	IsHidden     bool `json:"is_hidden" db:"is_hidden"`
	IsFeatured   bool `json:"is_featured" db:"is_featured"`
	DisplayOrder int  `json:"display_order" db:"display_order"`

	// Fechas
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// PlayerTitle representa la posesión de un título por un jugador
type PlayerTitle struct {
	ID       uuid.UUID `json:"id" db:"id"`
	PlayerID uuid.UUID `json:"player_id" db:"player_id"`
	TitleID  uuid.UUID `json:"title_id" db:"title_id"`

	// Estado de Posesión
	Status     string `json:"status" db:"status"` // active, inactive, expired, revoked
	IsUnlocked bool   `json:"is_unlocked" db:"is_unlocked"`
	IsEquipped bool   `json:"is_equipped" db:"is_equipped"`
	IsFavorite bool   `json:"is_favorite" db:"is_favorite"`

	// Información de Desbloqueo
	UnlockDate   time.Time  `json:"unlock_date" db:"unlock_date"`
	EquippedDate *time.Time `json:"equipped_date" db:"equipped_date"`
	UnlockMethod string     `json:"unlock_method" db:"unlock_method"` // achievement, event, purchase, etc.
	UnlockData   string     `json:"unlock_data" db:"unlock_data"`     // JSON con datos del desbloqueo

	// Duración y Expiración
	ExpiryDate    *time.Time `json:"expiry_date" db:"expiry_date"`
	DaysRemaining int        `json:"days_remaining" db:"days_remaining"`
	IsPermanent   bool       `json:"is_permanent" db:"is_permanent"`

	// Estadísticas de Uso
	TimesEquipped     int        `json:"times_equipped" db:"times_equipped"`
	TotalTimeEquipped int        `json:"total_time_equipped" db:"total_time_equipped"` // en minutos
	LastEquipped      *time.Time `json:"last_equipped" db:"last_equipped"`

	// Progreso y Logros
	Progress      int `json:"progress" db:"progress"` // progreso hacia el siguiente nivel
	MaxProgress   int `json:"max_progress" db:"max_progress"`
	Level         int `json:"level" db:"level"` // nivel del título (para títulos con niveles)
	MaxLevel      int `json:"max_level" db:"max_level"`
	PrestigeLevel int `json:"prestige_level" db:"prestige_level"`

	// Recompensas
	RewardsClaimed bool   `json:"rewards_claimed" db:"rewards_claimed"`
	RewardsData    string `json:"rewards_data" db:"rewards_data"` // JSON con recompensas recibidas
	PointsEarned   int    `json:"points_earned" db:"points_earned"`

	// Fechas
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// PrestigeLevel representa un nivel de prestigio
type PrestigeLevel struct {
	ID              uuid.UUID `json:"id" db:"id"`
	Name            string    `json:"name" db:"name"`
	Description     string    `json:"description" db:"description"`
	Icon            string    `json:"icon" db:"icon"`
	Color           string    `json:"color" db:"color"`
	BackgroundColor string    `json:"background_color" db:"background_color"`

	// Configuración del Nivel
	Level                int     `json:"level" db:"level"`
	PrestigeRequired     int     `json:"prestige_required" db:"prestige_required"`
	ExperienceMultiplier float64 `json:"experience_multiplier" db:"experience_multiplier"`

	// Bonificaciones
	Bonuses        string `json:"bonuses" db:"bonuses"`                 // JSON con bonificaciones del nivel
	SpecialEffects string `json:"special_effects" db:"special_effects"` // JSON con efectos especiales
	UnlockFeatures string `json:"unlock_features" db:"unlock_features"` // JSON con características desbloqueadas

	// Estado
	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// PlayerPrestige representa el prestigio de un jugador
type PlayerPrestige struct {
	ID       uuid.UUID `json:"id" db:"id"`
	PlayerID uuid.UUID `json:"player_id" db:"player_id"`

	// Prestigio Actual
	CurrentPrestige int `json:"current_prestige" db:"current_prestige"`
	TotalPrestige   int `json:"total_prestige" db:"total_prestige"`
	PrestigeLevel   int `json:"prestige_level" db:"prestige_level"`

	// Progreso
	PrestigeToNext  int     `json:"prestige_to_next" db:"prestige_to_next"`
	ProgressPercent float64 `json:"progress_percent" db:"progress_percent"`

	// Estadísticas
	TitlesUnlocked        int `json:"titles_unlocked" db:"titles_unlocked"`
	TitlesEquipped        int `json:"titles_equipped" db:"titles_equipped"`
	AchievementsCompleted int `json:"achievements_completed" db:"achievements_completed"`

	// Historial
	PrestigeHistory string    `json:"prestige_history" db:"prestige_history"` // JSON con historial de ganancias
	LastGainDate    time.Time `json:"last_gain_date" db:"last_gain_date"`
	LargestGain     int       `json:"largest_gain" db:"largest_gain"`

	// Rankings
	GlobalRank   int `json:"global_rank" db:"global_rank"`
	CategoryRank int `json:"category_rank" db:"category_rank"`
	AllianceRank int `json:"alliance_rank" db:"alliance_rank"`

	// Fechas
	FirstPrestige time.Time `json:"first_prestige" db:"first_prestige"`
	LastUpdated   time.Time `json:"last_updated" db:"last_updated"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

// TitleAchievement representa un logro relacionado con títulos
type TitleAchievement struct {
	ID      uuid.UUID `json:"id" db:"id"`
	TitleID uuid.UUID `json:"title_id" db:"title_id"`

	// Configuración del Logro
	Name            string `json:"name" db:"name"`
	Description     string `json:"description" db:"description"`
	Icon            string `json:"icon" db:"icon"`
	AchievementType string `json:"achievement_type" db:"achievement_type"` // unlock, equip, level, etc.

	// Requisitos
	Requirements string `json:"requirements" db:"requirements"`   // JSON con requisitos
	ProgressType string `json:"progress_type" db:"progress_type"` // count, time, score, etc.
	TargetValue  int    `json:"target_value" db:"target_value"`

	// Recompensas
	Rewards        string `json:"rewards" db:"rewards"` // JSON con recompensas
	PrestigeReward int    `json:"prestige_reward" db:"prestige_reward"`

	// Estado
	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// PlayerTitleAchievement representa el progreso de un jugador en un logro de título
type PlayerTitleAchievement struct {
	ID            uuid.UUID `json:"id" db:"id"`
	PlayerID      uuid.UUID `json:"player_id" db:"player_id"`
	AchievementID uuid.UUID `json:"achievement_id" db:"achievement_id"`

	// Progreso
	CurrentProgress int        `json:"current_progress" db:"current_progress"`
	IsCompleted     bool       `json:"is_completed" db:"is_completed"`
	CompletionDate  *time.Time `json:"completion_date" db:"completion_date"`

	// Recompensas
	RewardsClaimed bool       `json:"rewards_claimed" db:"rewards_claimed"`
	ClaimDate      *time.Time `json:"claim_date" db:"claim_date"`

	// Fechas
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// TitleEvent representa un evento relacionado con títulos
type TitleEvent struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Icon        string    `json:"icon" db:"icon"`

	// Configuración del Evento
	EventType string    `json:"event_type" db:"event_type"` // title_race, prestige_boost, unlock_fest, etc.
	StartDate time.Time `json:"start_date" db:"start_date"`
	EndDate   time.Time `json:"end_date" db:"end_date"`

	// Efectos del Evento
	Effects            string  `json:"effects" db:"effects"` // JSON con efectos del evento
	PrestigeMultiplier float64 `json:"prestige_multiplier" db:"prestige_multiplier"`
	UnlockChance       float64 `json:"unlock_chance" db:"unlock_chance"` // probabilidad aumentada de desbloqueo

	// Participación
	TotalParticipants  int `json:"total_participants" db:"total_participants"`
	ActiveParticipants int `json:"active_participants" db:"active_participants"`

	// Estado
	Status    string    `json:"status" db:"status"` // upcoming, active, completed
	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// TitleRanking representa el ranking de títulos de un jugador
type TitleRanking struct {
	ID       uuid.UUID `json:"id" db:"id"`
	PlayerID uuid.UUID `json:"player_id" db:"player_id"`

	// Rankings
	PrestigeRank     int `json:"prestige_rank" db:"prestige_rank"`
	TitlesRank       int `json:"titles_rank" db:"titles_rank"`
	AchievementsRank int `json:"achievements_rank" db:"achievements_rank"`
	OverallRank      int `json:"overall_rank" db:"overall_rank"`

	// Puntuaciones
	PrestigeScore     int `json:"prestige_score" db:"prestige_score"`
	TitlesScore       int `json:"titles_score" db:"titles_score"`
	AchievementsScore int `json:"achievements_score" db:"achievements_score"`
	OverallScore      int `json:"overall_score" db:"overall_score"`

	// Estadísticas
	RareTitles      int `json:"rare_titles" db:"rare_titles"`
	EpicTitles      int `json:"epic_titles" db:"epic_titles"`
	LegendaryTitles int `json:"legendary_titles" db:"legendary_titles"`
	MythicTitles    int `json:"mythic_titles" db:"mythic_titles"`
	DivineTitles    int `json:"divine_titles" db:"divine_titles"`

	// Metadatos
	LastUpdated time.Time `json:"last_updated" db:"last_updated"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// TitleStatistics representa las estadísticas de títulos de un jugador
type TitleStatistics struct {
	ID       uuid.UUID `json:"id" db:"id"`
	PlayerID uuid.UUID `json:"player_id" db:"player_id"`

	// Estadísticas Generales
	TotalTitlesUnlocked        int `json:"total_titles_unlocked" db:"total_titles_unlocked"`
	TotalTitlesEquipped        int `json:"total_titles_equipped" db:"total_titles_equipped"`
	TotalPrestigeGained        int `json:"total_prestige_gained" db:"total_prestige_gained"`
	TotalAchievementsCompleted int `json:"total_achievements_completed" db:"total_achievements_completed"`

	// Por Categoría
	CategoryStats string `json:"category_stats" db:"category_stats"` // JSON con estadísticas por categoría
	TypeStats     string `json:"type_stats" db:"type_stats"`         // JSON con estadísticas por tipo
	RarityStats   string `json:"rarity_stats" db:"rarity_stats"`     // JSON con estadísticas por rareza

	// Actividad
	LastTitleUnlocked *time.Time `json:"last_title_unlocked" db:"last_title_unlocked"`
	LastTitleEquipped *time.Time `json:"last_title_equipped" db:"last_title_equipped"`
	LastPrestigeGain  *time.Time `json:"last_prestige_gain" db:"last_prestige_gain"`

	// Logros
	LongestEquippedTitle string `json:"longest_equipped_title" db:"longest_equipped_title"`
	MostPrestigiousTitle string `json:"most_prestigious_title" db:"most_prestigious_title"`
	RarestTitle          string `json:"rarest_title" db:"rarest_title"`

	// Fechas
	FirstTitleUnlocked time.Time `json:"first_title_unlocked" db:"first_title_unlocked"`
	LastUpdated        time.Time `json:"last_updated" db:"last_updated"`
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
}

// TitleNotification representa una notificación de título
type TitleNotification struct {
	ID       uuid.UUID  `json:"id" db:"id"`
	PlayerID uuid.UUID  `json:"player_id" db:"player_id"`
	TitleID  *uuid.UUID `json:"title_id" db:"title_id"`

	// Notificación
	Type    string `json:"type" db:"type"` // unlock, equip, achievement, prestige, event
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

// TitleWithDetails representa un título con todos sus detalles
type TitleWithDetails struct {
	Title              *Title                   `json:"title"`
	Category           *TitleCategory           `json:"category"`
	PlayerTitle        *PlayerTitle             `json:"player_title"`
	Achievements       []TitleAchievement       `json:"achievements"`
	PlayerAchievements []PlayerTitleAchievement `json:"player_achievements"`
	Statistics         *TitleStatistics         `json:"statistics"`
}

// PrestigeDashboard representa el dashboard de prestigio de un jugador
type PrestigeDashboard struct {
	PlayerPrestige  *PlayerPrestige     `json:"player_prestige"`
	PrestigeLevel   *PrestigeLevel      `json:"prestige_level"`
	EquippedTitles  []PlayerTitle       `json:"equipped_titles"`
	RecentUnlocks   []PlayerTitle       `json:"recent_unlocks"`
	AvailableTitles []Title             `json:"available_titles"`
	Rankings        *TitleRanking       `json:"rankings"`
	Statistics      *TitleStatistics    `json:"statistics"`
	Notifications   []TitleNotification `json:"notifications"`
	ActiveEvents    []TitleEvent        `json:"active_events"`
	LastUpdated     time.Time           `json:"last_updated"`
}

// TitleProgressUpdate representa una actualización de progreso de título
type TitleProgressUpdate struct {
	PlayerID       uuid.UUID `json:"player_id"`
	TitleID        uuid.UUID `json:"title_id"`
	OldProgress    int       `json:"old_progress"`
	NewProgress    int       `json:"new_progress"`
	ProgressChange int       `json:"progress_change"`
	IsUnlocked     bool      `json:"is_unlocked"`
	PrestigeGained int       `json:"prestige_gained"`
	Timestamp      time.Time `json:"timestamp"`
}

// TitleRankingEntry representa una entrada en el ranking de títulos
type TitleRankingEntry struct {
	PlayerID      string `json:"player_id"`
	Username      string `json:"username"`
	TitlesCount   int    `json:"titles_count"`
	PrestigeLevel int    `json:"prestige_level"`
	Rank          int    `json:"rank"`
}

// PrestigeLeaderboardEntry representa una entrada en el leaderboard de prestigio
type PrestigeLeaderboardEntry struct {
	PlayerID       string    `json:"player_id"`
	Username       string    `json:"username"`
	PrestigeLevel  int       `json:"prestige_level"`
	PrestigePoints int       `json:"prestige_points"`
	LastUpdated    time.Time `json:"last_updated"`
	Rank           int       `json:"rank"`
}

// PrestigeStatistics representa estadísticas de prestigio
type PrestigeStatistics struct {
	PlayerID            string         `json:"player_id"`
	TitlesEarned        int            `json:"titles_earned"`
	AchievementsEarned  int            `json:"achievements_earned"`
	TotalPrestigeGained int            `json:"total_prestige_gained"`
	PrestigeSources     map[string]int `json:"prestige_sources"`
}

// PrestigeAchievement representa un logro de prestigio
type PrestigeAchievement struct {
	ID             string                 `json:"id"`
	Name           string                 `json:"name"`
	Description    string                 `json:"description"`
	PrestigePoints int                    `json:"prestige_points"`
	TriggerType    string                 `json:"trigger_type"`
	Requirements   map[string]interface{} `json:"requirements,omitempty"`
}

// PlayerPrestigeAchievement representa un logro de prestigio de un jugador
type PlayerPrestigeAchievement struct {
	PlayerID      string    `json:"player_id"`
	AchievementID string    `json:"achievement_id"`
	EarnedAt      time.Time `json:"earned_at"`
}

// TitleReward representa una recompensa de título
type TitleReward struct {
	ID      uuid.UUID `json:"id" db:"id"`
	TitleID uuid.UUID `json:"title_id" db:"title_id"`

	// Configuración de recompensa
	RewardType string `json:"reward_type" db:"reward_type"` // currency, items, experience, resources, etc.
	RewardData string `json:"reward_data" db:"reward_data"` // JSON con datos de la recompensa
	Quantity   int    `json:"quantity" db:"quantity"`

	// Condiciones
	LevelRequired int  `json:"level_required" db:"level_required"` // 0 = cualquier nivel
	IsRepeatable  bool `json:"is_repeatable" db:"is_repeatable"`
	IsGuaranteed  bool `json:"is_guaranteed" db:"is_guaranteed"` // recompensa garantizada o aleatoria

	// Estado
	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// TitleDashboard representa el dashboard principal de títulos
type TitleDashboard struct {
	PlayerPrestige  *PlayerPrestige     `json:"player_prestige"`
	PrestigeLevel   *PrestigeLevel      `json:"prestige_level"`
	EquippedTitles  []PlayerTitle       `json:"equipped_titles"`
	RecentUnlocks   []PlayerTitle       `json:"recent_unlocks"`
	AvailableTitles []Title             `json:"available_titles"`
	Rankings        *TitleRanking       `json:"rankings"`
	Statistics      *TitleStatistics    `json:"statistics"`
	Notifications   []TitleNotification `json:"notifications"`
	ActiveEvents    []TitleEvent        `json:"active_events"`
	Categories      []TitleCategory     `json:"categories"`
	LastUpdated     time.Time           `json:"last_updated"`
}

// TitleLeaderboard representa una tabla de clasificación de títulos
type TitleLeaderboard struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Type        string    `json:"type" db:"type"` // prestige, titles, achievements, overall

	// Configuración
	CategoryID   *uuid.UUID `json:"category_id" db:"category_id"`
	DisplayOrder int        `json:"display_order" db:"display_order"`
	IsActive     bool       `json:"is_active" db:"is_active"`

	// Estadísticas
	TotalParticipants int       `json:"total_participants" db:"total_participants"`
	LastUpdated       time.Time `json:"last_updated" db:"last_updated"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
}
