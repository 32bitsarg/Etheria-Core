package models

import (
	"time"
)

// Technology representa una tecnología en el juego
type Technology struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Category     string                 `json:"category"`     // military, economic, social, scientific
	SubCategory  string                 `json:"sub_category"` // infantry, cavalry, trade, etc.
	Level        int                    `json:"level"`
	MaxLevel     int                    `json:"max_level"`
	ResearchTime int                    `json:"research_time"` // en segundos
	ResearchCost string                 `json:"research_cost"` // JSON con costos por nivel
	Requirements string                 `json:"requirements"`  // JSON con tecnologías requeridas
	Effects      map[string]interface{} `json:"effects"`       // JSON con efectos de la tecnología
	Icon         string                 `json:"icon"`
	Color        string                 `json:"color"`
	IsActive     bool                   `json:"is_active"`
	IsSpecial    bool                   `json:"is_special"` // Tecnología especial/única
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
	BaseCost     map[string]int         `json:"base_cost"`
}

// PlayerTechnology representa una tecnología de un jugador
type PlayerTechnology struct {
	ID            string     `json:"id"`
	PlayerID      string     `json:"player_id"`
	TechnologyID  string     `json:"technology_id"`
	Level         int        `json:"level"`
	IsResearching bool       `json:"is_researching"`
	StartedAt     *time.Time `json:"started_at,omitempty"`
	CompletedAt   *time.Time `json:"completed_at,omitempty"`
	Progress      int        `json:"progress"` // Progreso en segundos
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// ResearchQueue representa la cola de investigación de un jugador
type ResearchQueue struct {
	ID           string               `json:"id"`
	PlayerID     string               `json:"player_id"`
	QueueItems   []*ResearchQueueItem `json:"queue_items"`
	MaxQueueSize int                  `json:"max_queue_size"`
	CreatedAt    time.Time            `json:"created_at"`
}

// TechnologyEffect representa un efecto de una tecnología
type TechnologyEffect struct {
	ID           string                 `json:"id"`
	TechnologyID string                 `json:"technology_id"`
	EffectType   string                 `json:"effect_type"`   // production, combat, building, etc.
	Target       string                 `json:"target"`        // wood_production, infantry_attack, etc.
	Value        float64                `json:"value"`         // Valor del efecto
	IsPercentage bool                   `json:"is_percentage"` // Si es porcentaje o valor fijo
	Level        int                    `json:"level"`         // Nivel de la tecnología para este efecto
	Data         map[string]interface{} `json:"data,omitempty"`
}

// TechnologyRequirement representa un requisito de una tecnología
type TechnologyRequirement struct {
	ID              int `json:"id"`
	TechnologyID    int `json:"technology_id"`
	RequiredTechID  int `json:"required_tech_id"`
	RequiredLevel   int `json:"required_level"`
	RequiredVillage int `json:"required_village"` // Nivel mínimo de aldea
}

// TechnologyCost representa el costo de investigación de una tecnología
type TechnologyCost struct {
	ID           int    `json:"id"`
	TechnologyID int    `json:"technology_id"`
	Level        int    `json:"level"`
	ResourceType string `json:"resource_type"` // wood, stone, food, gold
	Amount       int    `json:"amount"`
}

// ResearchBonus representa bonificaciones de investigación
type ResearchBonus struct {
	ID        int        `json:"id"`
	PlayerID  int        `json:"player_id"`
	Source    string     `json:"source"`     // building, event, item, etc.
	BonusType string     `json:"bonus_type"` // speed, cost, experience
	Value     float64    `json:"value"`
	Category  string     `json:"category"` // military, economic, etc.
	ExpiresAt *time.Time `json:"expires_at"`
	CreatedAt time.Time  `json:"created_at"`
}

// TechnologyTree representa el árbol de tecnologías
type TechnologyTree struct {
	ID           int          `json:"id"`
	Name         string       `json:"name"`
	Description  string       `json:"description"`
	Category     string       `json:"category"`
	Technologies []Technology `json:"technologies"`
}

// ResearchHistory representa el historial de investigación
type ResearchHistory struct {
	ID           int       `json:"id"`
	PlayerID     int       `json:"player_id"`
	TechnologyID int       `json:"technology_id"`
	Level        int       `json:"level"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
	Duration     int       `json:"duration"` // Duración real en segundos
	Cost         string    `json:"cost"`     // JSON con costos reales
	CreatedAt    time.Time `json:"created_at"`
}

// TechnologyUnlock representa desbloqueos de tecnologías
type TechnologyUnlock struct {
	ID           int       `json:"id"`
	PlayerID     int       `json:"player_id"`
	TechnologyID int       `json:"technology_id"`
	UnlockType   string    `json:"unlock_type"` // research, event, achievement, etc.
	UnlockedAt   time.Time `json:"unlocked_at"`
	Source       string    `json:"source"` // Detalles del desbloqueo
}

// ResearchAchievement representa logros de investigación
type ResearchAchievement struct {
	ID              int       `json:"id"`
	PlayerID        int       `json:"player_id"`
	AchievementType string    `json:"achievement_type"`
	Value           int       `json:"value"`
	CompletedAt     time.Time `json:"completed_at"`
	Reward          string    `json:"reward"` // JSON con recompensas
}

// TechnologyEvent representa eventos de tecnología
type TechnologyEvent struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	EventType   string    `json:"event_type"` // research_boost, cost_reduction, etc.
	Category    string    `json:"category"`
	BonusValue  float64   `json:"bonus_value"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
}

// ResearchStatistics representa estadísticas de investigación
type ResearchStatistics struct {
	PlayerID           string     `json:"player_id"`
	TotalTechnologies  int        `json:"total_technologies"`
	MilitaryTechs      int        `json:"military_techs"`
	EconomicTechs      int        `json:"economic_techs"`
	SocialTechs        int        `json:"social_techs"`
	ScientificTechs    int        `json:"scientific_techs"`
	TotalLevels        int        `json:"total_levels"`
	ResearchTime       int        `json:"research_time"`   // Tiempo total de investigación
	ResourcesSpent     string     `json:"resources_spent"` // JSON con recursos gastados
	LastResearchAt     *time.Time `json:"last_research_at"`
	ResearchEfficiency float64    `json:"research_efficiency"`
}

// TechnologyWithDetails representa una tecnología con todos sus detalles
type TechnologyWithDetails struct {
	Technology    *Technology             `json:"technology"`
	PlayerLevel   int                     `json:"player_level"`
	IsResearching bool                    `json:"is_researching"`
	CanResearch   bool                    `json:"can_research"`
	Requirements  []TechnologyRequirement `json:"requirements"`
	Effects       []TechnologyEffect      `json:"effects"`
	Costs         []TechnologyCost        `json:"costs"`
	ResearchTime  int                     `json:"research_time"`
	Progress      int                     `json:"progress"`
	StartTime     *time.Time              `json:"start_time"`
	EndTime       *time.Time              `json:"end_time"`
	Prerequisites []*Technology           `json:"prerequisites"`
}

// ResearchQueueItem representa un item en la cola de investigación
type ResearchQueueItem struct {
	ID            string     `json:"id"`
	PlayerID      string     `json:"player_id"`
	TechnologyID  string     `json:"technology_id"`
	Priority      int        `json:"priority"`
	AddedAt       time.Time  `json:"added_at"`
	StartedAt     *time.Time `json:"started_at,omitempty"`
	Status        string     `json:"status"`
	Progress      int        `json:"progress"`
	TimeRemaining int        `json:"time_remaining"`
}

// TechnologyRanking representa ranking de tecnologías
type TechnologyRanking struct {
	PlayerID        int    `json:"player_id"`
	PlayerName      string `json:"player_name"`
	TotalTechs      int    `json:"total_techs"`
	TotalLevels     int    `json:"total_levels"`
	MilitaryScore   int    `json:"military_score"`
	EconomicScore   int    `json:"economic_score"`
	SocialScore     int    `json:"social_score"`
	ScientificScore int    `json:"scientific_score"`
	Rank            int    `json:"rank"`
}

// ResearchRecommendation representa recomendaciones de investigación
type ResearchRecommendation struct {
	TechnologyID  string `json:"technology_id"`
	Priority      int    `json:"priority"`
	Reason        string `json:"reason"`
	EstimatedTime int    `json:"estimated_time"`
}

// ResearchProgress representa el progreso de investigación
type ResearchProgress struct {
	TechnologyID  string  `json:"technology_id"`
	CurrentLevel  int     `json:"current_level"`
	MaxLevel      int     `json:"max_level"`
	Progress      float64 `json:"progress"`
	TimeRemaining int     `json:"time_remaining"`
	IsResearching bool    `json:"is_researching"`
}

// ResearchCost representa el costo de investigación
type ResearchCost struct {
	TechnologyID string         `json:"technology_id"`
	Level        int            `json:"level"`
	Costs        map[string]int `json:"costs"`
	TimeRequired int            `json:"time_required"`
}

// TechnologyDetails representa los detalles completos de una tecnología
type TechnologyDetails struct {
	Technology   Technology              `json:"technology"`
	Effects      []TechnologyEffect      `json:"effects"`
	Requirements []TechnologyRequirement `json:"requirements"`
	Costs        []TechnologyCost        `json:"costs"`
	LastUpdated  time.Time               `json:"last_updated"`
}

// ResearchData representa datos de investigación para uso con Redis
type ResearchData struct {
	TechnologyID   string    `json:"technology_id"`
	TechnologyName string    `json:"technology_name"`
	Level          int       `json:"level"`
	Progress       int       `json:"progress"`   // segundos
	TotalTime      int       `json:"total_time"` // segundos
	StartedAt      time.Time `json:"started_at"`
	EndsAt         time.Time `json:"ends_at"`
	IsActive       bool      `json:"is_active"`
}
