package models

import (
	"time"
)

// RankingCategory representa una categoría de ranking
type RankingCategory struct {
	ID          int    `json:"id" db:"id"`
	Name        string `json:"name" db:"name"`
	Description string `json:"description" db:"description"`
	Type        string `json:"type" db:"type"`         // player, alliance, village, world
	SubType     string `json:"sub_type" db:"sub_type"` // combat, economy, research, etc.
	Icon        string `json:"icon" db:"icon"`
	Color       string `json:"color" db:"color"`

	// Configuración del ranking
	UpdateInterval int    `json:"update_interval" db:"update_interval"` // en minutos
	MaxPositions   int    `json:"max_positions" db:"max_positions"`
	MinScore       int    `json:"min_score" db:"min_score"`
	ScoreFormula   string `json:"score_formula" db:"score_formula"` // JSON con fórmula de cálculo

	// Recompensas
	RewardsEnabled bool   `json:"rewards_enabled" db:"rewards_enabled"`
	RewardsConfig  string `json:"rewards_config" db:"rewards_config"` // JSON con configuración de recompensas

	// Visualización
	DisplayOrder    int  `json:"display_order" db:"display_order"`
	IsPublic        bool `json:"is_public" db:"is_public"`
	ShowInDashboard bool `json:"show_in_dashboard" db:"show_in_dashboard"`

	// Estado
	IsActive    bool      `json:"is_active" db:"is_active"`
	LastUpdated time.Time `json:"last_updated" db:"last_updated"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// RankingSeason representa una temporada de rankings
type RankingSeason struct {
	ID           int    `json:"id" db:"id"`
	Name         string `json:"name" db:"name"`
	Description  string `json:"description" db:"description"`
	SeasonNumber int    `json:"season_number" db:"season_number"`

	// Período de la temporada
	StartDate time.Time `json:"start_date" db:"start_date"`
	EndDate   time.Time `json:"end_date" db:"end_date"`
	IsActive  bool      `json:"is_active" db:"is_active"`

	// Configuración
	Categories     string `json:"categories" db:"categories"` // JSON con categorías incluidas
	RewardsEnabled bool   `json:"rewards_enabled" db:"rewards_enabled"`
	RewardsConfig  string `json:"rewards_config" db:"rewards_config"`

	// Estadísticas de la temporada
	TotalParticipants int `json:"total_participants" db:"total_participants"`
	TotalAlliances    int `json:"total_alliances" db:"total_alliances"`
	TotalVillages     int `json:"total_villages" db:"total_villages"`

	// Estado
	Status    string    `json:"status" db:"status"` // upcoming, active, finished, archived
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// RankingEntry representa una entrada en un ranking
type RankingEntry struct {
	ID         int  `json:"id" db:"id"`
	CategoryID int  `json:"category_id" db:"category_id"`
	SeasonID   *int `json:"season_id" db:"season_id"`

	// Entidad rankeada
	EntityType string `json:"entity_type" db:"entity_type"` // player, alliance, village, world
	EntityID   int    `json:"entity_id" db:"entity_id"`
	EntityName string `json:"entity_name" db:"entity_name"`

	// Posición y puntuación
	Position         int `json:"position" db:"position"`
	Score            int `json:"score" db:"score"`
	PreviousPosition int `json:"previous_position" db:"previous_position"`
	PositionChange   int `json:"position_change" db:"position_change"`

	// Estadísticas detalladas
	Stats     string `json:"stats" db:"stats"`         // JSON con estadísticas específicas
	Breakdown string `json:"breakdown" db:"breakdown"` // JSON con desglose de puntuación

	// Metadatos
	LastUpdated time.Time `json:"last_updated" db:"last_updated"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// PlayerStatistics representa estadísticas detalladas de un jugador
type PlayerStatistics struct {
	ID       int `json:"id" db:"id"`
	PlayerID int `json:"player_id" db:"player_id"`

	// Estadísticas generales
	TotalPlayTime int       `json:"total_play_time" db:"total_play_time"` // en minutos
	DaysActive    int       `json:"days_active" db:"days_active"`
	LastActive    time.Time `json:"last_active" db:"last_active"`

	// Estadísticas de aldeas
	VillagesOwned     int `json:"villages_owned" db:"villages_owned"`
	TotalPopulation   int `json:"total_population" db:"total_population"`
	BuildingsBuilt    int `json:"buildings_built" db:"buildings_built"`
	BuildingsUpgraded int `json:"buildings_upgraded" db:"buildings_upgraded"`

	// Estadísticas de combate
	BattlesWon   int     `json:"battles_won" db:"battles_won"`
	BattlesLost  int     `json:"battles_lost" db:"battles_lost"`
	BattlesTotal int     `json:"battles_total" db:"battles_total"`
	WinRate      float64 `json:"win_rate" db:"win_rate"`

	// Unidades
	UnitsTrained int `json:"units_trained" db:"units_trained"`
	UnitsLost    int `json:"units_lost" db:"units_lost"`
	UnitsKilled  int `json:"units_killed" db:"units_killed"`

	// Héroes
	HeroesRecruited int `json:"heroes_recruited" db:"heroes_recruited"`
	HeroesUpgraded  int `json:"heroes_upgraded" db:"heroes_upgraded"`
	HeroesActive    int `json:"heroes_active" db:"heroes_active"`

	// Investigación
	TechnologiesResearched int `json:"technologies_researched" db:"technologies_researched"`
	ResearchPoints         int `json:"research_points" db:"research_points"`

	// Economía
	TotalEarned        int `json:"total_earned" db:"total_earned"`
	TotalSpent         int `json:"total_spent" db:"total_spent"`
	MarketTransactions int `json:"market_transactions" db:"market_transactions"`
	ItemsSold          int `json:"items_sold" db:"items_sold"`
	ItemsBought        int `json:"items_bought" db:"items_bought"`

	// Alianzas
	AlliancesJoined      int  `json:"alliances_joined" db:"alliances_joined"`
	CurrentAlliance      *int `json:"current_alliance" db:"current_alliance"`
	AllianceContribution int  `json:"alliance_contribution" db:"alliance_contribution"`

	// Logros
	AchievementsEarned int `json:"achievements_earned" db:"achievements_earned"`
	TotalPoints        int `json:"total_points" db:"total_points"`

	// Rankings
	BestRanking    int `json:"best_ranking" db:"best_ranking"`
	CurrentRanking int `json:"current_ranking" db:"current_ranking"`
	RankingsWon    int `json:"rankings_won" db:"rankings_won"`

	// Eventos
	EventsParticipated int `json:"events_participated" db:"events_participated"`
	EventsWon          int `json:"events_won" db:"events_won"`

	// Recursos
	ResourcesGathered int `json:"resources_gathered" db:"resources_gathered"`
	ResourcesSpent    int `json:"resources_spent" db:"resources_spent"`

	// Fechas
	FirstLogin  time.Time `json:"first_login" db:"first_login"`
	LastUpdated time.Time `json:"last_updated" db:"last_updated"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// AllianceStatistics representa estadísticas de una alianza
type AllianceStatistics struct {
	ID         int `json:"id" db:"id"`
	AllianceID int `json:"alliance_id" db:"alliance_id"`

	// Miembros
	TotalMembers  int `json:"total_members" db:"total_members"`
	ActiveMembers int `json:"active_members" db:"active_members"`
	NewMembers    int `json:"new_members" db:"new_members"`

	// Aldeas
	TotalVillages   int `json:"total_villages" db:"total_villages"`
	TotalPopulation int `json:"total_population" db:"total_population"`

	// Combate
	BattlesWon  int     `json:"battles_won" db:"battles_won"`
	BattlesLost int     `json:"battles_lost" db:"battles_lost"`
	WinRate     float64 `json:"win_rate" db:"win_rate"`

	// Unidades
	TotalUnits   int `json:"total_units" db:"total_units"`
	UnitsTrained int `json:"units_trained" db:"units_trained"`
	UnitsLost    int `json:"units_lost" db:"units_lost"`

	// Héroes
	TotalHeroes  int `json:"total_heroes" db:"total_heroes"`
	HeroesActive int `json:"heroes_active" db:"heroes_active"`

	// Investigación
	TechnologiesResearched int `json:"technologies_researched" db:"technologies_researched"`
	ResearchPoints         int `json:"research_points" db:"research_points"`

	// Economía
	TotalEarned  int `json:"total_earned" db:"total_earned"`
	TotalSpent   int `json:"total_spent" db:"total_spent"`
	MarketVolume int `json:"market_volume" db:"market_volume"`

	// Rankings
	BestRanking    int `json:"best_ranking" db:"best_ranking"`
	CurrentRanking int `json:"current_ranking" db:"current_ranking"`
	RankingsWon    int `json:"rankings_won" db:"rankings_won"`

	// Eventos
	EventsParticipated int `json:"events_participated" db:"events_participated"`
	EventsWon          int `json:"events_won" db:"events_won"`

	// Actividad
	AverageActivity float64   `json:"average_activity" db:"average_activity"`
	LastActivity    time.Time `json:"last_activity" db:"last_activity"`

	// Fechas
	FoundedDate time.Time `json:"founded_date" db:"founded_date"`
	LastUpdated time.Time `json:"last_updated" db:"last_updated"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// VillageStatistics representa estadísticas de una aldea
type VillageStatistics struct {
	ID        int `json:"id" db:"id"`
	VillageID int `json:"village_id" db:"village_id"`
	PlayerID  int `json:"player_id" db:"player_id"`

	// Población y desarrollo
	Population    int `json:"population" db:"population"`
	MaxPopulation int `json:"max_population" db:"max_population"`
	Happiness     int `json:"happiness" db:"happiness"`

	// Edificios
	BuildingsBuilt      int `json:"buildings_built" db:"buildings_built"`
	BuildingsUpgraded   int `json:"buildings_upgraded" db:"buildings_upgraded"`
	TotalBuildingLevels int `json:"total_building_levels" db:"total_building_levels"`

	// Recursos
	ResourcesProduced int `json:"resources_produced" db:"resources_produced"`
	ResourcesConsumed int `json:"resources_consumed" db:"resources_consumed"`
	ResourcesStored   int `json:"resources_stored" db:"resources_stored"`

	// Unidades
	UnitsTrained   int `json:"units_trained" db:"units_trained"`
	UnitsStationed int `json:"units_stationed" db:"units_stationed"`
	UnitsLost      int `json:"units_lost" db:"units_lost"`

	// Defensa
	DefenseStrength int `json:"defense_strength" db:"defense_strength"`
	AttacksDefended int `json:"attacks_defended" db:"attacks_defended"`
	AttacksSuffered int `json:"attacks_suffered" db:"attacks_suffered"`

	// Investigación
	TechnologiesResearched int `json:"technologies_researched" db:"technologies_researched"`
	ResearchPoints         int `json:"research_points" db:"research_points"`

	// Economía
	MarketTransactions int `json:"market_transactions" db:"market_transactions"`
	ItemsSold          int `json:"items_sold" db:"items_sold"`
	ItemsBought        int `json:"items_bought" db:"items_bought"`

	// Héroes
	HeroesAssigned int `json:"heroes_assigned" db:"heroes_assigned"`
	HeroesActive   int `json:"heroes_active" db:"heroes_active"`

	// Actividad
	LastActivity  time.Time `json:"last_activity" db:"last_activity"`
	ActivityScore int       `json:"activity_score" db:"activity_score"`

	// Fechas
	FoundedDate time.Time `json:"founded_date" db:"founded_date"`
	LastUpdated time.Time `json:"last_updated" db:"last_updated"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// WorldStatistics representa estadísticas de un mundo
type WorldStatistics struct {
	ID      int `json:"id" db:"id"`
	WorldID int `json:"world_id" db:"world_id"`

	// Población
	TotalPlayers  int `json:"total_players" db:"total_players"`
	ActivePlayers int `json:"active_players" db:"active_players"`
	NewPlayers    int `json:"new_players" db:"new_players"`

	// Alianzas
	TotalAlliances  int `json:"total_alliances" db:"total_alliances"`
	ActiveAlliances int `json:"active_alliances" db:"active_alliances"`

	// Aldeas
	TotalVillages  int `json:"total_villages" db:"total_villages"`
	ActiveVillages int `json:"active_villages" db:"active_villages"`

	// Combate
	TotalBattles          int `json:"total_battles" db:"total_battles"`
	BattlesToday          int `json:"battles_today" db:"battles_today"`
	AverageBattleDuration int `json:"average_battle_duration" db:"average_battle_duration"`

	// Unidades
	TotalUnits   int `json:"total_units" db:"total_units"`
	UnitsTrained int `json:"units_trained" db:"units_trained"`
	UnitsLost    int `json:"units_lost" db:"units_lost"`

	// Héroes
	TotalHeroes     int `json:"total_heroes" db:"total_heroes"`
	HeroesRecruited int `json:"heroes_recruited" db:"heroes_recruited"`
	HeroesActive    int `json:"heroes_active" db:"heroes_active"`

	// Investigación
	TechnologiesResearched int `json:"technologies_researched" db:"technologies_researched"`
	ResearchPoints         int `json:"research_points" db:"research_points"`

	// Economía
	TotalMarketVolume  int `json:"total_market_volume" db:"total_market_volume"`
	MarketTransactions int `json:"market_transactions" db:"market_transactions"`
	TotalTaxes         int `json:"total_taxes" db:"total_taxes"`

	// Recursos
	ResourcesProduced int `json:"resources_produced" db:"resources_produced"`
	ResourcesConsumed int `json:"resources_consumed" db:"resources_consumed"`

	// Eventos
	ActiveEvents    int `json:"active_events" db:"active_events"`
	EventsCompleted int `json:"events_completed" db:"events_completed"`

	// Actividad
	AverageActivity float64   `json:"average_activity" db:"average_activity"`
	PeakActivity    time.Time `json:"peak_activity" db:"peak_activity"`

	// Fechas
	WorldStartDate time.Time `json:"world_start_date" db:"world_start_date"`
	LastUpdated    time.Time `json:"last_updated" db:"last_updated"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

// RankingHistory representa el historial de posiciones en rankings
type RankingHistory struct {
	ID         int    `json:"id" db:"id"`
	CategoryID int    `json:"category_id" db:"category_id"`
	SeasonID   *int   `json:"season_id" db:"season_id"`
	EntityType string `json:"entity_type" db:"entity_type"`
	EntityID   int    `json:"entity_id" db:"entity_id"`

	// Posición histórica
	Position int `json:"position" db:"position"`
	Score    int `json:"score" db:"score"`

	// Fecha
	RecordedAt time.Time `json:"recorded_at" db:"recorded_at"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// RankingReward representa una recompensa de ranking
type RankingReward struct {
	ID         int  `json:"id" db:"id"`
	CategoryID int  `json:"category_id" db:"category_id"`
	SeasonID   *int `json:"season_id" db:"season_id"`

	// Configuración de la recompensa
	Position   int    `json:"position" db:"position"`       // 0 = todos, 1 = primer lugar, etc.
	RewardType string `json:"reward_type" db:"reward_type"` // currency, items, title, etc.
	RewardData string `json:"reward_data" db:"reward_data"` // JSON con datos de la recompensa

	// Estado
	IsClaimed bool       `json:"is_claimed" db:"is_claimed"`
	ClaimedBy *int       `json:"claimed_by" db:"claimed_by"`
	ClaimedAt *time.Time `json:"claimed_at" db:"claimed_at"`

	// Fechas
	ExpiresAt *time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
}

// StatisticsSnapshot representa una instantánea de estadísticas
type StatisticsSnapshot struct {
	ID         int    `json:"id" db:"id"`
	EntityType string `json:"entity_type" db:"entity_type"` // player, alliance, village, world
	EntityID   int    `json:"entity_id" db:"entity_id"`

	// Datos de la instantánea
	SnapshotData string `json:"snapshot_data" db:"snapshot_data"` // JSON con datos completos
	SnapshotType string `json:"snapshot_type" db:"snapshot_type"` // daily, weekly, monthly

	// Fecha
	SnapshotDate time.Time `json:"snapshot_date" db:"snapshot_date"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// RankingWithDetails representa un ranking con detalles completos
type RankingWithDetails struct {
	Category      *RankingCategory     `json:"category"`
	Season        *RankingSeason       `json:"season"`
	Entries       []RankingEntry       `json:"entries"`
	PlayerStats   []PlayerStatistics   `json:"player_stats"`
	AllianceStats []AllianceStatistics `json:"alliance_stats"`
	VillageStats  []VillageStatistics  `json:"village_stats"`
	WorldStats    *WorldStatistics     `json:"world_stats"`
	LastUpdated   time.Time            `json:"last_updated"`
}

// StatisticsSummary representa un resumen de estadísticas
type StatisticsSummary struct {
	TotalPlayers   int       `json:"total_players"`
	ActivePlayers  int       `json:"active_players"`
	TotalAlliances int       `json:"total_alliances"`
	TotalVillages  int       `json:"total_villages"`
	TotalBattles   int       `json:"total_battles"`
	TotalUnits     int       `json:"total_units"`
	TotalHeroes    int       `json:"total_heroes"`
	MarketVolume   int       `json:"market_volume"`
	ResearchPoints int       `json:"research_points"`
	LastUpdated    time.Time `json:"last_updated"`
}

// RankingDashboard representa el dashboard completo de rankings
type RankingDashboard struct {
	Categories     []RankingCategory `json:"categories"`
	ActiveSeason   *RankingSeason    `json:"active_season"`
	TopRankings    []RankingEntry    `json:"top_rankings"`
	PlayerRankings []RankingEntry    `json:"player_rankings"`
	Statistics     StatisticsSummary `json:"statistics"`
	LastUpdated    time.Time         `json:"last_updated"`
}
