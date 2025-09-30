package models

import (
	"time"

	"github.com/google/uuid"
)

type Unit struct {
	ID                     uuid.UUID  `json:"id"`
	VillageID              uuid.UUID  `json:"village_id"`
	Type                   string     `json:"type"`
	Quantity               int        `json:"quantity"`
	InTraining             int        `json:"in_training"`
	TrainingCompletionTime *time.Time `json:"training_completion_time,omitempty"`
	CreatedAt              time.Time  `json:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at"`
}

type UnitType struct {
	Type        string `json:"type"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Attack      int    `json:"attack"`
	Defense     int    `json:"defense"`
	Speed       int    `json:"speed"`
	Capacity    int    `json:"capacity"`
	Cost        struct {
		Wood  int `json:"wood"`
		Stone int `json:"stone"`
		Food  int `json:"food"`
		Gold  int `json:"gold"`
	} `json:"cost"`
	TrainingTime int `json:"training_time"` // en segundos
}

// Definir tipos de unidades disponibles
var UnitTypes = map[string]UnitType{
	"warrior": {
		Type:        "warrior",
		Name:        "Guerrero",
		Description: "Unidad básica de combate",
		Attack:      10,
		Defense:     8,
		Speed:       5,
		Capacity:    10,
		Cost: struct {
			Wood  int `json:"wood"`
			Stone int `json:"stone"`
			Food  int `json:"food"`
			Gold  int `json:"gold"`
		}{
			Wood:  50,
			Stone: 30,
			Food:  20,
			Gold:  10,
		},
		TrainingTime: 60,
	},
	"archer": {
		Type:        "archer",
		Name:        "Arquero",
		Description: "Unidad de ataque a distancia",
		Attack:      15,
		Defense:     5,
		Speed:       6,
		Capacity:    8,
		Cost: struct {
			Wood  int `json:"wood"`
			Stone int `json:"stone"`
			Food  int `json:"food"`
			Gold  int `json:"gold"`
		}{
			Wood:  60,
			Stone: 20,
			Food:  25,
			Gold:  15,
		},
		TrainingTime: 90,
	},
	"knight": {
		Type:        "knight",
		Name:        "Caballero",
		Description: "Unidad de élite con alta defensa",
		Attack:      12,
		Defense:     15,
		Speed:       4,
		Capacity:    15,
		Cost: struct {
			Wood  int `json:"wood"`
			Stone int `json:"stone"`
			Food  int `json:"food"`
			Gold  int `json:"gold"`
		}{
			Wood:  80,
			Stone: 60,
			Food:  40,
			Gold:  30,
		},
		TrainingTime: 120,
	},
	"scout": {
		Type:        "scout",
		Name:        "Explorador",
		Description: "Unidad rápida para exploración",
		Attack:      5,
		Defense:     3,
		Speed:       10,
		Capacity:    5,
		Cost: struct {
			Wood  int `json:"wood"`
			Stone int `json:"stone"`
			Food  int `json:"food"`
			Gold  int `json:"gold"`
		}{
			Wood:  30,
			Stone: 10,
			Food:  15,
			Gold:  5,
		},
		TrainingTime: 45,
	},
}
