package models

import (
	"time"

	"github.com/google/uuid"
)

// BattleSystemConfig representa la configuración del sistema de batallas
type BattleSystemConfig struct {
	ID                uuid.UUID `json:"id" db:"id"`
	IsEnabled         bool      `json:"is_enabled" db:"is_enabled"`
	AdvancedMode      bool      `json:"advanced_mode" db:"advanced_mode"`             // true = modo avanzado, false = modo básico
	MaxBattleDuration int       `json:"max_battle_duration" db:"max_battle_duration"` // en segundos
	MaxWaves          int       `json:"max_waves" db:"max_waves"`                     // máximo de oleadas por batalla
	AutoResolve       bool      `json:"auto_resolve" db:"auto_resolve"`               // resolver automáticamente batallas simples

	// Configuraciones avanzadas
	AdvancedConfig string    `json:"advanced_config" db:"advanced_config"` // JSON con configuraciones avanzadas
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// MilitaryUnit representa una unidad militar
type MilitaryUnit struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Type        string    `json:"type" db:"type"`         // infantry, cavalry, archer, siege, magic, etc.
	Category    string    `json:"category" db:"category"` // light, medium, heavy, elite, etc.
	Tier        int       `json:"tier" db:"tier"`         // 1-10, nivel de la unidad

	// Estadísticas de combate
	Health      int     `json:"health" db:"health"`
	MaxHealth   int     `json:"max_health" db:"max_health"`
	AttackSpeed float64 `json:"attack_speed" db:"attack_speed"` // ataques por segundo
	Range       int     `json:"range" db:"range"`               // alcance de ataque

	// Ataque físico
	PhysicalAttack      int `json:"physical_attack" db:"physical_attack"`
	PhysicalPenetration int `json:"physical_penetration" db:"physical_penetration"`

	// Ataque mágico
	MagicAttack      int `json:"magic_attack" db:"magic_attack"`
	MagicPenetration int `json:"magic_penetration" db:"magic_penetration"`

	// Defensa física
	PhysicalDefense    int `json:"physical_defense" db:"physical_defense"`
	PhysicalResistance int `json:"physical_resistance" db:"physical_resistance"`

	// Defensa mágica
	MagicDefense    int `json:"magic_defense" db:"magic_defense"`
	MagicResistance int `json:"magic_resistance" db:"magic_resistance"`

	// Estadísticas especiales
	CriticalChance float64 `json:"critical_chance" db:"critical_chance"` // probabilidad de crítico
	CriticalDamage float64 `json:"critical_damage" db:"critical_damage"` // multiplicador de daño crítico
	DodgeChance    float64 `json:"dodge_chance" db:"dodge_chance"`       // probabilidad de esquivar
	BlockChance    float64 `json:"block_chance" db:"block_chance"`       // probabilidad de bloquear
	Moral          int     `json:"moral" db:"moral"`                     // moral base (0-100)

	// Costos y requisitos
	TrainingCost string `json:"training_cost" db:"training_cost"` // JSON con costos de entrenamiento
	UpgradeCost  string `json:"upgrade_cost" db:"upgrade_cost"`   // JSON con costos de mejora
	Requirements string `json:"requirements" db:"requirements"`   // JSON con requisitos

	// Ventajas y desventajas
	Advantages    string `json:"advantages" db:"advantages"`       // JSON con ventajas contra otros tipos
	Disadvantages string `json:"disadvantages" db:"disadvantages"` // JSON con desventajas

	// Visual
	Icon      string `json:"icon" db:"icon"`
	Model     string `json:"model" db:"model"`
	Animation string `json:"animation" db:"animation"`

	// Estado
	IsActive  bool      `json:"is_active" db:"is_active"`
	IsElite   bool      `json:"is_elite" db:"is_elite"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// PlayerUnit representa una unidad de un jugador
type PlayerUnit struct {
	ID         uuid.UUID `json:"id" db:"id"`
	PlayerID   uuid.UUID `json:"player_id" db:"player_id"`
	UnitID     uuid.UUID `json:"unit_id" db:"unit_id"`
	Quantity   int       `json:"quantity" db:"quantity"`
	Level      int       `json:"level" db:"level"`
	Experience int       `json:"experience" db:"experience"`

	// Estadísticas actuales (con bonificaciones)
	CurrentHealth  int `json:"current_health" db:"current_health"`
	CurrentAttack  int `json:"current_attack" db:"current_attack"`
	CurrentDefense int `json:"current_defense" db:"current_defense"`
	CurrentMoral   int `json:"current_moral" db:"current_moral"`

	// Equipamiento y bonificaciones
	Equipment string `json:"equipment" db:"equipment"` // JSON con equipamiento
	Bonuses   string `json:"bonuses" db:"bonuses"`     // JSON con bonificaciones

	// Estado
	IsTraining      bool       `json:"is_training" db:"is_training"`
	TrainingEndTime *time.Time `json:"training_end_time" db:"training_end_time"`
	IsInjured       bool       `json:"is_injured" db:"is_injured"`
	InjuryTime      *time.Time `json:"injury_time" db:"injury_time"`

	// Estadísticas de batalla
	BattlesWon  int `json:"battles_won" db:"battles_won"`
	BattlesLost int `json:"battles_lost" db:"battles_lost"`
	UnitsKilled int `json:"units_killed" db:"units_killed"`
	UnitsLost   int `json:"units_lost" db:"units_lost"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Battle representa una batalla
type Battle struct {
	ID         uuid.UUID `json:"id" db:"id"`
	AttackerID uuid.UUID `json:"attacker_id" db:"attacker_id"`
	DefenderID uuid.UUID `json:"defender_id" db:"defender_id"`
	BattleType string    `json:"battle_type" db:"battle_type"` // pvp, pve, siege, raid, etc.
	Mode       string    `json:"mode" db:"mode"`               // basic, advanced

	// Configuración de la batalla
	MaxWaves    int `json:"max_waves" db:"max_waves"`
	MaxDuration int `json:"max_duration" db:"max_duration"` // en segundos

	// Estado de la batalla
	Status      string     `json:"status" db:"status"` // pending, active, completed, cancelled
	CurrentWave int        `json:"current_wave" db:"current_wave"`
	StartTime   *time.Time `json:"start_time" db:"start_time"`
	EndTime     *time.Time `json:"end_time" db:"end_time"`
	Duration    int        `json:"duration" db:"duration"` // en segundos

	// Resultado
	Winner         string `json:"winner" db:"winner"`                   // attacker, defender, draw
	AttackerLosses string `json:"attacker_losses" db:"attacker_losses"` // JSON con pérdidas
	DefenderLosses string `json:"defender_losses" db:"defender_losses"` // JSON con pérdidas

	// Configuración avanzada (solo si mode = advanced)
	Terrain           string `json:"terrain" db:"terrain"`                       // plain, forest, mountain, etc.
	Weather           string `json:"weather" db:"weather"`                       // sunny, rainy, snowy, etc.
	AttackerFormation string `json:"attacker_formation" db:"attacker_formation"` // JSON con formación
	DefenderFormation string `json:"defender_formation" db:"defender_formation"` // JSON con formación
	AttackerTactics   string `json:"attacker_tactics" db:"attacker_tactics"`     // JSON con tácticas
	DefenderTactics   string `json:"defender_tactics" db:"defender_tactics"`     // JSON con tácticas

	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// BattleWave representa una oleada de batalla
type BattleWave struct {
	ID         uuid.UUID `json:"id" db:"id"`
	BattleID   uuid.UUID `json:"battle_id" db:"battle_id"`
	WaveNumber int       `json:"wave_number" db:"wave_number"`

	// Unidades participantes
	AttackerUnits string `json:"attacker_units" db:"attacker_units"` // JSON con unidades atacantes
	DefenderUnits string `json:"defender_units" db:"defender_units"` // JSON con unidades defensoras

	// Resultado de la oleada
	AttackerDamage int    `json:"attacker_damage" db:"attacker_damage"`
	DefenderDamage int    `json:"defender_damage" db:"defender_damage"`
	AttackerLosses string `json:"attacker_losses" db:"attacker_losses"` // JSON con pérdidas
	DefenderLosses string `json:"defender_losses" db:"defender_losses"` // JSON con pérdidas

	// Detalles del combate
	CombatLog string `json:"combat_log" db:"combat_log"` // JSON con log detallado
	Duration  int    `json:"duration" db:"duration"`     // en segundos

	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// BattleFormation representa una formación de batalla
type BattleFormation struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Type        string    `json:"type" db:"type"` // offensive, defensive, balanced

	// Configuración de la formación
	Layout    string `json:"layout" db:"layout"`       // JSON con disposición de unidades
	Bonuses   string `json:"bonuses" db:"bonuses"`     // JSON con bonificaciones
	Penalties string `json:"penalties" db:"penalties"` // JSON con penalizaciones

	// Requisitos
	Requirements string `json:"requirements" db:"requirements"` // JSON con requisitos
	MinUnits     int    `json:"min_units" db:"min_units"`
	MaxUnits     int    `json:"max_units" db:"max_units"`

	// Visual
	Icon    string `json:"icon" db:"icon"`
	Preview string `json:"preview" db:"preview"`

	// Estado
	IsActive   bool      `json:"is_active" db:"is_active"`
	IsAdvanced bool      `json:"is_advanced" db:"is_advanced"` // solo para modo avanzado
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// BattleTactic representa una táctica de batalla
type BattleTactic struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Type        string    `json:"type" db:"type"` // offensive, defensive, support

	// Efectos de la táctica
	Effects  string `json:"effects" db:"effects"`   // JSON con efectos
	Target   string `json:"target" db:"target"`     // self, enemy, ally, all
	Duration int    `json:"duration" db:"duration"` // en oleadas

	// Costos y requisitos
	Cost         string `json:"cost" db:"cost"`                 // JSON con costos
	Requirements string `json:"requirements" db:"requirements"` // JSON con requisitos

	// Visual
	Icon      string `json:"icon" db:"icon"`
	Animation string `json:"animation" db:"animation"`

	// Estado
	IsActive   bool      `json:"is_active" db:"is_active"`
	IsAdvanced bool      `json:"is_advanced" db:"is_advanced"` // solo para modo avanzado
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// BattleTerrain representa un tipo de terreno
type BattleTerrain struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Type        string    `json:"type" db:"type"` // plain, forest, mountain, water, etc.

	// Efectos del terreno
	Bonuses   string `json:"bonuses" db:"bonuses"`     // JSON con bonificaciones por tipo de unidad
	Penalties string `json:"penalties" db:"penalties"` // JSON con penalizaciones por tipo de unidad

	// Configuración visual
	Icon  string `json:"icon" db:"icon"`
	Model string `json:"model" db:"model"`
	Color string `json:"color" db:"color"`

	// Estado
	IsActive   bool      `json:"is_active" db:"is_active"`
	IsAdvanced bool      `json:"is_advanced" db:"is_advanced"` // solo para modo avanzado
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// BattleWeather representa un tipo de clima
type BattleWeather struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Type        string    `json:"type" db:"type"` // sunny, rainy, snowy, foggy, etc.

	// Efectos del clima
	Effects       string `json:"effects" db:"effects"`               // JSON con efectos globales
	UnitModifiers string `json:"unit_modifiers" db:"unit_modifiers"` // JSON con modificadores por tipo de unidad

	// Configuración visual
	Icon           string `json:"icon" db:"icon"`
	ParticleEffect string `json:"particle_effect" db:"particle_effect"`

	// Estado
	IsActive   bool      `json:"is_active" db:"is_active"`
	IsAdvanced bool      `json:"is_advanced" db:"is_advanced"` // solo para modo avanzado
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// BattleSimulation representa una simulación de batalla
type BattleSimulation struct {
	ID       uuid.UUID `json:"id" db:"id"`
	PlayerID uuid.UUID `json:"player_id" db:"player_id"`
	BattleID uuid.UUID `json:"battle_id" db:"battle_id"`

	// Configuración de la simulación
	AttackerUnits string `json:"attacker_units" db:"attacker_units"` // JSON con unidades atacantes
	DefenderUnits string `json:"defender_units" db:"defender_units"` // JSON con unidades defensoras
	Mode          string `json:"mode" db:"mode"`                     // basic, advanced

	// Configuración avanzada
	Terrain           string `json:"terrain" db:"terrain"`
	Weather           string `json:"weather" db:"weather"`
	AttackerFormation string `json:"attacker_formation" db:"attacker_formation"`
	DefenderFormation string `json:"defender_formation" db:"defender_formation"`
	AttackerTactics   string `json:"attacker_tactics" db:"attacker_tactics"`
	DefenderTactics   string `json:"defender_tactics" db:"defender_tactics"`

	// Resultado de la simulación
	Result         string `json:"result" db:"result"` // attacker_win, defender_win, draw
	AttackerLosses string `json:"attacker_losses" db:"attacker_losses"`
	DefenderLosses string `json:"defender_losses" db:"defender_losses"`
	BattleLog      string `json:"battle_log" db:"battle_log"` // JSON con log completo

	// Estadísticas
	Duration    int `json:"duration" db:"duration"`
	Waves       int `json:"waves" db:"waves"`
	TotalDamage int `json:"total_damage" db:"total_damage"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// BattleRanking representa el ranking de batallas de un jugador
type BattleRanking struct {
	PlayerID    uuid.UUID `json:"player_id" db:"player_id"`
	PlayerName  string    `json:"player_name" db:"player_name"`
	BattlesWon  int       `json:"battles_won" db:"battles_won"`
	BattlesLost int       `json:"battles_lost" db:"battles_lost"`
	WinRate     float64   `json:"win_rate" db:"win_rate"`
	TotalDamage int       `json:"total_damage" db:"total_damage"`
	UnitsKilled int       `json:"units_killed" db:"units_killed"`
	UnitsLost   int       `json:"units_lost" db:"units_lost"`
	Rank        int       `json:"rank" db:"rank"`
}

// BattleWithDetails representa una batalla con todos sus detalles
type BattleWithDetails struct {
	Battle            *Battle          `json:"battle"`
	Waves             []BattleWave     `json:"waves"`
	AttackerUnits     []PlayerUnit     `json:"attacker_units"`
	DefenderUnits     []PlayerUnit     `json:"defender_units"`
	Terrain           *BattleTerrain   `json:"terrain,omitempty"`
	Weather           *BattleWeather   `json:"weather,omitempty"`
	AttackerFormation *BattleFormation `json:"attacker_formation,omitempty"`
	DefenderFormation *BattleFormation `json:"defender_formation,omitempty"`
	AttackerTactics   []BattleTactic   `json:"attacker_tactics,omitempty"`
	DefenderTactics   []BattleTactic   `json:"defender_tactics,omitempty"`
}

// BattleRequest representa una solicitud de batalla
type BattleRequest struct {
	AttackerID        uuid.UUID              `json:"attacker_id"`
	AttackerVillageID uuid.UUID              `json:"attacker_village_id"`
	DefenderVillageID uuid.UUID              `json:"defender_village_id"`
	BattleType        string                 `json:"battle_type"` // pvp, pve, siege, raid
	Mode              string                 `json:"mode"`        // basic, advanced
	Units             map[string]int         `json:"units"`       // tipo_unidad -> cantidad
	Formation         string                 `json:"formation,omitempty"`
	Tactics           []string               `json:"tactics,omitempty"`
	Terrain           string                 `json:"terrain,omitempty"`
	Weather           string                 `json:"weather,omitempty"`
	MaxWaves          int                    `json:"max_waves,omitempty"`
	MaxDuration       int                    `json:"max_duration,omitempty"`
	AdvancedConfig    map[string]interface{} `json:"advanced_config,omitempty"`
}

// BattleStatistics representa estadísticas de batalla de un jugador
type BattleStatistics struct {
	PlayerID           uuid.UUID `json:"player_id" db:"player_id"`
	TotalBattles       int       `json:"total_battles" db:"total_battles"`
	BattlesWon         int       `json:"battles_won" db:"battles_won"`
	BattlesLost        int       `json:"battles_lost" db:"battles_lost"`
	WinRate            float64   `json:"win_rate" db:"win_rate"`
	TotalDamageDealt   int       `json:"total_damage_dealt" db:"total_damage_dealt"`
	TotalDamageTaken   int       `json:"total_damage_taken" db:"total_damage_taken"`
	UnitsKilled        int       `json:"units_killed" db:"units_killed"`
	UnitsLost          int       `json:"units_lost" db:"units_lost"`
	KillDeathRatio     float64   `json:"kill_death_ratio" db:"kill_death_ratio"`
	AverageBattleTime  int       `json:"average_battle_time" db:"average_battle_time"`
	LongestBattleTime  int       `json:"longest_battle_time" db:"longest_battle_time"`
	ShortestBattleTime int       `json:"shortest_battle_time" db:"shortest_battle_time"`
	LastBattleDate     time.Time `json:"last_battle_date" db:"last_battle_date"`
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time `json:"updated_at" db:"updated_at"`
}
