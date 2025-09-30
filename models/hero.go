package models

import (
	"time"
)

// Hero representa un héroe en el juego
type Hero struct {
	ID               int    `json:"id" db:"id"`
	Name             string `json:"name" db:"name"`
	Title            string `json:"title" db:"title"`
	Description      string `json:"description" db:"description"`
	Race             string `json:"race" db:"race"`     // human, elf, dwarf, orc, etc.
	Class            string `json:"class" db:"class"`   // warrior, mage, archer, etc.
	Rarity           string `json:"rarity" db:"rarity"` // common, rare, epic, legendary, mythical
	Level            int    `json:"level" db:"level"`
	MaxLevel         int    `json:"max_level" db:"max_level"`
	Experience       int    `json:"experience" db:"experience"`
	ExperienceToNext int    `json:"experience_to_next" db:"experience_to_next"`

	// Estadísticas base
	Health       int `json:"health" db:"health"`
	Attack       int `json:"attack" db:"attack"`
	Defense      int `json:"defense" db:"defense"`
	Speed        int `json:"speed" db:"speed"`
	Intelligence int `json:"intelligence" db:"intelligence"`
	Charisma     int `json:"charisma" db:"charisma"`

	// Habilidades
	ActiveSkills  string `json:"active_skills" db:"active_skills"`   // JSON con habilidades activas
	PassiveSkills string `json:"passive_skills" db:"passive_skills"` // JSON con habilidades pasivas
	UltimateSkill string `json:"ultimate_skill" db:"ultimate_skill"` // Habilidad definitiva

	// Equipamiento
	Equipment string `json:"equipment" db:"equipment"` // JSON con slots de equipamiento

	// Artefactos y tesoros
	Artifacts string `json:"artifacts" db:"artifacts"` // JSON con artefactos equipados

	// Costos y requisitos
	RecruitCost  string `json:"recruit_cost" db:"recruit_cost"` // JSON con costos de reclutamiento
	UpgradeCost  string `json:"upgrade_cost" db:"upgrade_cost"` // JSON con costos de mejora
	Requirements string `json:"requirements" db:"requirements"` // JSON con requisitos para desbloquear

	// Visual y presentación
	Icon     string `json:"icon" db:"icon"`
	Portrait string `json:"portrait" db:"portrait"`
	Model    string `json:"model" db:"model"`
	Color    string `json:"color" db:"color"`

	// Estado del sistema
	IsActive  bool `json:"is_active" db:"is_active"`
	IsSpecial bool `json:"is_special" db:"is_special"` // Héroe especial/evento
	IsLimited bool `json:"is_limited" db:"is_limited"` // Héroe de tiempo limitado

	// Fechas
	ReleaseDate *time.Time `json:"release_date" db:"release_date"`
	ExpiryDate  *time.Time `json:"expiry_date" db:"expiry_date"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

// PlayerHero representa un héroe de un jugador
type PlayerHero struct {
	ID               int `json:"id" db:"id"`
	PlayerID         int `json:"player_id" db:"player_id"`
	HeroID           int `json:"hero_id" db:"hero_id"`
	Level            int `json:"level" db:"level"`
	Experience       int `json:"experience" db:"experience"`
	ExperienceToNext int `json:"experience_to_next" db:"experience_to_next"`

	// Estadísticas actuales (con bonificaciones)
	CurrentHealth       int `json:"current_health" db:"current_health"`
	MaxHealth           int `json:"max_health" db:"max_health"`
	CurrentAttack       int `json:"current_attack" db:"current_attack"`
	CurrentDefense      int `json:"current_defense" db:"current_defense"`
	CurrentSpeed        int `json:"current_speed" db:"current_speed"`
	CurrentIntelligence int `json:"current_intelligence" db:"current_intelligence"`
	CurrentCharisma     int `json:"current_charisma" db:"current_charisma"`

	// Estado del héroe
	IsRecruited bool       `json:"is_recruited" db:"is_recruited"`
	IsActive    bool       `json:"is_active" db:"is_active"` // Si está en uso
	IsInjured   bool       `json:"is_injured" db:"is_injured"`
	InjuryTime  *time.Time `json:"injury_time" db:"injury_time"`

	// Progreso y logros
	BattlesWon       int `json:"battles_won" db:"battles_won"`
	BattlesLost      int `json:"battles_lost" db:"battles_lost"`
	QuestsCompleted  int `json:"quests_completed" db:"quests_completed"`
	ExperienceGained int `json:"experience_gained" db:"experience_gained"`

	// Equipamiento actual
	Equipment string `json:"equipment" db:"equipment"` // JSON con equipamiento actual
	Artifacts string `json:"artifacts" db:"artifacts"` // JSON con artefactos actuales

	// Habilidades desbloqueadas
	UnlockedSkills string `json:"unlocked_skills" db:"unlocked_skills"` // JSON con habilidades desbloqueadas
	SkillLevels    string `json:"skill_levels" db:"skill_levels"`       // JSON con niveles de habilidades

	// Fechas
	RecruitedAt *time.Time `json:"recruited_at" db:"recruited_at"`
	LastUsedAt  *time.Time `json:"last_used_at" db:"last_used_at"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

// HeroSkill representa una habilidad de héroe
type HeroSkill struct {
	ID          int    `json:"id" db:"id"`
	HeroID      int    `json:"hero_id" db:"hero_id"`
	Name        string `json:"name" db:"name"`
	Description string `json:"description" db:"description"`
	Type        string `json:"type" db:"type"`         // active, passive, ultimate
	Category    string `json:"category" db:"category"` // combat, utility, support, etc.
	Level       int    `json:"level" db:"level"`       // Nivel requerido para desbloquear
	MaxLevel    int    `json:"max_level" db:"max_level"`

	// Efectos de la habilidad
	Effects  string `json:"effects" db:"effects"` // JSON con efectos
	Target   string `json:"target" db:"target"`   // self, enemy, ally, all
	Range    int    `json:"range" db:"range"`
	Cooldown int    `json:"cooldown" db:"cooldown"` // en segundos
	Duration int    `json:"duration" db:"duration"` // en segundos

	// Costos
	ManaCost   int `json:"mana_cost" db:"mana_cost"`
	HealthCost int `json:"health_cost" db:"health_cost"`

	// Visual
	Icon      string `json:"icon" db:"icon"`
	Animation string `json:"animation" db:"animation"`
	Sound     string `json:"sound" db:"sound"`

	// Estado
	IsActive   bool      `json:"is_active" db:"is_active"`
	IsUltimate bool      `json:"is_ultimate" db:"is_ultimate"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// HeroEquipment representa equipamiento de héroe
type HeroEquipment struct {
	ID          int    `json:"id" db:"id"`
	HeroID      int    `json:"hero_id" db:"hero_id"`
	Slot        string `json:"slot" db:"slot"` // weapon, armor, helmet, boots, etc.
	Name        string `json:"name" db:"name"`
	Description string `json:"description" db:"description"`
	Type        string `json:"type" db:"type"` // weapon, armor, accessory
	Rarity      string `json:"rarity" db:"rarity"`
	Level       int    `json:"level" db:"level"`
	MaxLevel    int    `json:"max_level" db:"max_level"`

	// Estadísticas del equipamiento
	Stats        string `json:"stats" db:"stats"`               // JSON con estadísticas
	Requirements string `json:"requirements" db:"requirements"` // JSON con requisitos

	// Costos
	Cost        string `json:"cost" db:"cost"`                 // JSON con costos
	UpgradeCost string `json:"upgrade_cost" db:"upgrade_cost"` // JSON con costos de mejora

	// Visual
	Icon  string `json:"icon" db:"icon"`
	Model string `json:"model" db:"model"`
	Color string `json:"color" db:"color"`

	// Estado
	IsActive  bool      `json:"is_active" db:"is_active"`
	IsSpecial bool      `json:"is_special" db:"is_special"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// HeroArtifact representa un artefacto de héroe
type HeroArtifact struct {
	ID          int    `json:"id" db:"id"`
	HeroID      int    `json:"hero_id" db:"hero_id"`
	Name        string `json:"name" db:"name"`
	Description string `json:"description" db:"description"`
	Type        string `json:"type" db:"type"` // weapon, armor, accessory, special
	Rarity      string `json:"rarity" db:"rarity"`
	Level       int    `json:"level" db:"level"`
	MaxLevel    int    `json:"max_level" db:"max_level"`

	// Poderes del artefacto
	Powers       string `json:"powers" db:"powers"`             // JSON con poderes
	Effects      string `json:"effects" db:"effects"`           // JSON con efectos
	Requirements string `json:"requirements" db:"requirements"` // JSON con requisitos

	// Costos y mejora
	Cost        string `json:"cost" db:"cost"`                 // JSON con costos
	UpgradeCost string `json:"upgrade_cost" db:"upgrade_cost"` // JSON con costos de mejora

	// Visual
	Icon           string `json:"icon" db:"icon"`
	Model          string `json:"model" db:"model"`
	ParticleEffect string `json:"particle_effect" db:"particle_effect"`

	// Estado
	IsActive   bool      `json:"is_active" db:"is_active"`
	IsMythical bool      `json:"is_mythical" db:"is_mythical"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// HeroQuest representa una misión específica de héroe
type HeroQuest struct {
	ID          int    `json:"id" db:"id"`
	HeroID      int    `json:"hero_id" db:"hero_id"`
	Name        string `json:"name" db:"name"`
	Description string `json:"description" db:"description"`
	Type        string `json:"type" db:"type"`             // personal, world, event
	Category    string `json:"category" db:"category"`     // combat, exploration, social, etc.
	Level       int    `json:"level" db:"level"`           // Nivel requerido del héroe
	Difficulty  string `json:"difficulty" db:"difficulty"` // easy, medium, hard, epic

	// Objetivos
	Objectives   string `json:"objectives" db:"objectives"`     // JSON con objetivos
	Requirements string `json:"requirements" db:"requirements"` // JSON con requisitos

	// Recompensas
	Rewards          string `json:"rewards" db:"rewards"` // JSON con recompensas
	ExperienceReward int    `json:"experience_reward" db:"experience_reward"`

	// Tiempo
	Duration  int        `json:"duration" db:"duration"` // en segundos
	TimeLimit *time.Time `json:"time_limit" db:"time_limit"`

	// Estado
	IsActive     bool      `json:"is_active" db:"is_active"`
	IsRepeatable bool      `json:"is_repeatable" db:"is_repeatable"`
	IsEvent      bool      `json:"is_event" db:"is_event"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// PlayerHeroQuest representa una misión de héroe de un jugador
type PlayerHeroQuest struct {
	ID             int        `json:"id" db:"id"`
	PlayerID       int        `json:"player_id" db:"player_id"`
	HeroID         int        `json:"hero_id" db:"hero_id"`
	QuestID        int        `json:"quest_id" db:"quest_id"`
	Status         string     `json:"status" db:"status"`     // available, active, completed, failed
	Progress       string     `json:"progress" db:"progress"` // JSON con progreso de objetivos
	StartTime      *time.Time `json:"start_time" db:"start_time"`
	EndTime        *time.Time `json:"end_time" db:"end_time"`
	CompletedAt    *time.Time `json:"completed_at" db:"completed_at"`
	RewardsClaimed bool       `json:"rewards_claimed" db:"rewards_claimed"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

// HeroBattle representa una batalla de héroe
type HeroBattle struct {
	ID         int    `json:"id" db:"id"`
	PlayerID   int    `json:"player_id" db:"player_id"`
	HeroID     int    `json:"hero_id" db:"hero_id"`
	EnemyType  string `json:"enemy_type" db:"enemy_type"` // monster, player, boss, etc.
	EnemyID    int    `json:"enemy_id" db:"enemy_id"`
	BattleType string `json:"battle_type" db:"battle_type"` // pve, pvp, boss, event
	Result     string `json:"result" db:"result"`           // victory, defeat, draw
	Duration   int    `json:"duration" db:"duration"`       // en segundos

	// Estadísticas de la batalla
	DamageDealt      int    `json:"damage_dealt" db:"damage_dealt"`
	DamageReceived   int    `json:"damage_received" db:"damage_received"`
	SkillsUsed       string `json:"skills_used" db:"skills_used"` // JSON con habilidades usadas
	ExperienceGained int    `json:"experience_gained" db:"experience_gained"`

	// Recompensas
	Rewards string `json:"rewards" db:"rewards"` // JSON con recompensas
	Loot    string `json:"loot" db:"loot"`       // JSON con botín

	// Estado
	IsInjured      bool      `json:"is_injured" db:"is_injured"`
	InjuryDuration int       `json:"injury_duration" db:"injury_duration"` // en segundos
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

// HeroRanking representa ranking de héroes
type HeroRanking struct {
	PlayerID    int     `json:"player_id" db:"player_id"`
	PlayerName  string  `json:"player_name" db:"player_name"`
	HeroID      int     `json:"hero_id" db:"hero_id"`
	HeroName    string  `json:"hero_name" db:"hero_name"`
	Level       int     `json:"level" db:"level"`
	Experience  int     `json:"experience" db:"experience"`
	BattlesWon  int     `json:"battles_won" db:"battles_won"`
	BattlesLost int     `json:"battles_lost" db:"battles_lost"`
	WinRate     float64 `json:"win_rate" db:"win_rate"`
	TotalPower  int     `json:"total_power" db:"total_power"`
	Rank        int     `json:"rank" db:"rank"`
}

// HeroEvent representa eventos especiales de héroes
type HeroEvent struct {
	ID          int       `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Type        string    `json:"type" db:"type"`         // tournament, recruitment, training, etc.
	Category    string    `json:"category" db:"category"` // combat, social, economic, etc.
	StartTime   time.Time `json:"start_time" db:"start_time"`
	EndTime     time.Time `json:"end_time" db:"end_time"`
	IsActive    bool      `json:"is_active" db:"is_active"`

	// Configuración del evento
	Config       string `json:"config" db:"config"`             // JSON con configuración
	Rewards      string `json:"rewards" db:"rewards"`           // JSON con recompensas
	Requirements string `json:"requirements" db:"requirements"` // JSON con requisitos

	// Participación
	MaxParticipants     int `json:"max_participants" db:"max_participants"`
	CurrentParticipants int `json:"current_participants" db:"current_participants"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// HeroSystemConfig representa la configuración del sistema de héroes
type HeroSystemConfig struct {
	ID                   int     `json:"id" db:"id"`
	IsEnabled            bool    `json:"is_enabled" db:"is_enabled"`
	MaxHeroesPerPlayer   int     `json:"max_heroes_per_player" db:"max_heroes_per_player"`
	MaxActiveHeroes      int     `json:"max_active_heroes" db:"max_active_heroes"`
	ExperienceMultiplier float64 `json:"experience_multiplier" db:"experience_multiplier"`
	InjuryDuration       int     `json:"injury_duration" db:"injury_duration"` // en segundos
	RecoveryCost         string  `json:"recovery_cost" db:"recovery_cost"`     // JSON con costos de recuperación

	// Configuraciones avanzadas
	AdvancedConfig string    `json:"advanced_config" db:"advanced_config"` // JSON con configuraciones avanzadas
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// HeroWithDetails representa un héroe con todos sus detalles
type HeroWithDetails struct {
	Hero                *Hero           `json:"hero"`
	PlayerHero          *PlayerHero     `json:"player_hero,omitempty"`
	Skills              []HeroSkill     `json:"skills"`
	Equipment           []HeroEquipment `json:"equipment"`
	Artifacts           []HeroArtifact  `json:"artifacts"`
	Quests              []HeroQuest     `json:"quests"`
	CanRecruit          bool            `json:"can_recruit"`
	CanUpgrade          bool            `json:"can_upgrade"`
	RecruitCost         string          `json:"recruit_cost"`
	UpgradeCost         string          `json:"upgrade_cost"`
	Requirements        string          `json:"requirements"`
	MissingRequirements []string        `json:"missing_requirements"`
}

// HeroProgress representa el progreso de un héroe
type HeroProgress struct {
	HeroID              int     `json:"hero_id" db:"hero_id"`
	HeroName            string  `json:"hero_name" db:"hero_name"`
	Level               int     `json:"level" db:"level"`
	Experience          int     `json:"experience" db:"experience"`
	ExperienceToNext    int     `json:"experience_to_next" db:"experience_to_next"`
	Progress            float64 `json:"progress" db:"progress"` // Porcentaje de progreso al siguiente nivel
	TotalPower          int     `json:"total_power" db:"total_power"`
	BattlesWon          int     `json:"battles_won" db:"battles_won"`
	BattlesLost         int     `json:"battles_lost" db:"battles_lost"`
	WinRate             float64 `json:"win_rate" db:"win_rate"`
	QuestsCompleted     int     `json:"quests_completed" db:"quests_completed"`
	IsActive            bool    `json:"is_active" db:"is_active"`
	IsInjured           bool    `json:"is_injured" db:"is_injured"`
	InjuryTimeRemaining int     `json:"injury_time_remaining" db:"injury_time_remaining"`
}
