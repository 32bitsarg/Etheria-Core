package models

import (
	"time"

	"github.com/google/uuid"
)

type ResourceProduction struct {
	VillageID  uuid.UUID `json:"village_id"`
	Wood       int       `json:"wood_per_hour"`
	Stone      int       `json:"stone_per_hour"`
	Food       int       `json:"food_per_hour"`
	Gold       int       `json:"gold_per_hour"`
	LastUpdate time.Time `json:"last_update"`
}

type ResourceStorage struct {
	VillageID    uuid.UUID `json:"village_id"`
	WoodStorage  int       `json:"wood_storage"`
	StoneStorage int       `json:"stone_storage"`
	FoodStorage  int       `json:"food_storage"`
	GoldStorage  int       `json:"gold_storage"`
}

type ResourceUpdate struct {
	VillageID uuid.UUID `json:"village_id"`
	Wood      int       `json:"wood"`
	Stone     int       `json:"stone"`
	Food      int       `json:"food"`
	Gold      int       `json:"gold"`
	Timestamp time.Time `json:"timestamp"`
}

type ResourceData struct {
	Wood  int `json:"wood"`
	Stone int `json:"stone"`
	Food  int `json:"food"`
	Gold  int `json:"gold"`
	Gems  int `json:"gems"`
}

// Configuración de producción base por edificio
var BaseProductionRates = map[string]map[string]int{
	"town_hall": {
		"wood":  5,
		"stone": 3,
		"food":  8,
		"gold":  2,
	},
	"marketplace": {
		"wood":  0,
		"stone": 0,
		"food":  0,
		"gold":  5,
	},
}

// Configuración de almacenamiento por edificio
var BaseStorageRates = map[string]map[string]int{
	"warehouse": {
		"wood":  200,
		"stone": 200,
		"food":  0,
		"gold":  0,
	},
	"granary": {
		"wood":  0,
		"stone": 0,
		"food":  200,
		"gold":  0,
	},
	"marketplace": {
		"wood":  0,
		"stone": 0,
		"food":  0,
		"gold":  100,
	},
}

// Calcular producción de recursos basada en edificios
func CalculateResourceProduction(buildings map[string]*Building) ResourceProduction {
	production := ResourceProduction{
		Wood:       10, // Producción base
		Stone:      5,  // Producción base
		Food:       15, // Producción base
		Gold:       1,  // Producción base
		LastUpdate: time.Now(),
	}

	for buildingType, building := range buildings {
		if buildingType == "town_hall" && building.Level > 0 {
			// El ayuntamiento aumenta la producción base
			multiplier := building.Level
			production.Wood += BaseProductionRates[buildingType]["wood"] * multiplier
			production.Stone += BaseProductionRates[buildingType]["stone"] * multiplier
			production.Food += BaseProductionRates[buildingType]["food"] * multiplier
			production.Gold += BaseProductionRates[buildingType]["gold"] * multiplier
		} else if buildingType == "marketplace" && building.Level > 0 {
			// El mercado genera oro
			production.Gold += BaseProductionRates[buildingType]["gold"] * building.Level
		}
	}

	return production
}

// Calcular capacidad de almacenamiento basada en edificios
func CalculateResourceStorage(buildings map[string]*Building) ResourceStorage {
	storage := ResourceStorage{
		WoodStorage:  1000, // Almacenamiento base
		StoneStorage: 1000, // Almacenamiento base
		FoodStorage:  1000, // Almacenamiento base
		GoldStorage:  500,  // Almacenamiento base
	}

	for buildingType, building := range buildings {
		if buildingType == "warehouse" && building.Level > 0 {
			storage.WoodStorage += BaseStorageRates[buildingType]["wood"] * building.Level
			storage.StoneStorage += BaseStorageRates[buildingType]["stone"] * building.Level
		} else if buildingType == "granary" && building.Level > 0 {
			storage.FoodStorage += BaseStorageRates[buildingType]["food"] * building.Level
		} else if buildingType == "marketplace" && building.Level > 0 {
			storage.GoldStorage += BaseStorageRates[buildingType]["gold"] * building.Level
		}
	}

	return storage
}

// Calcular recursos generados desde la última actualización
func CalculateGeneratedResources(production ResourceProduction, lastUpdate time.Time) ResourceUpdate {
	now := time.Now()
	hoursElapsed := now.Sub(lastUpdate).Hours()

	if hoursElapsed < 0 {
		hoursElapsed = 0
	}

	return ResourceUpdate{
		Wood:      int(float64(production.Wood) * hoursElapsed),
		Stone:     int(float64(production.Stone) * hoursElapsed),
		Food:      int(float64(production.Food) * hoursElapsed),
		Gold:      int(float64(production.Gold) * hoursElapsed),
		Timestamp: now,
	}
}
