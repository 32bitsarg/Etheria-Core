package models

import (
	"time"
)

type BuildingType struct {
	Type        string `json:"type"`
	Name        string `json:"name"`
	Description string `json:"description"`
	MaxLevel    int    `json:"max_level"`
	Cost        struct {
		Wood  int `json:"wood"`
		Stone int `json:"stone"`
		Food  int `json:"food"`
		Gold  int `json:"gold"`
	} `json:"cost"`
	UpgradeTime int `json:"upgrade_time"` // en segundos
	Effects     struct {
		PopulationBonus    int `json:"population_bonus"`
		ResourceProduction struct {
			Wood  int `json:"wood"`
			Stone int `json:"stone"`
			Food  int `json:"food"`
			Gold  int `json:"gold"`
		} `json:"resource_production"`
		StorageBonus struct {
			Wood  int `json:"wood"`
			Stone int `json:"stone"`
			Food  int `json:"food"`
			Gold  int `json:"gold"`
		} `json:"storage_bonus"`
		MilitaryBonus struct {
			Attack  int `json:"attack"`
			Defense int `json:"defense"`
		} `json:"military_bonus"`
	} `json:"effects"`
}

// Definir tipos de edificios disponibles
var BuildingTypes = map[string]BuildingType{
	"town_hall": {
		Type:        "town_hall",
		Name:        "Ayuntamiento",
		Description: "Centro de la aldea, determina el nivel máximo de otros edificios",
		MaxLevel:    20,
		Cost: struct {
			Wood  int `json:"wood"`
			Stone int `json:"stone"`
			Food  int `json:"food"`
			Gold  int `json:"gold"`
		}{
			Wood:  100,
			Stone: 100,
			Food:  50,
			Gold:  20,
		},
		UpgradeTime: 300,
		Effects: struct {
			PopulationBonus    int `json:"population_bonus"`
			ResourceProduction struct {
				Wood  int `json:"wood"`
				Stone int `json:"stone"`
				Food  int `json:"food"`
				Gold  int `json:"gold"`
			} `json:"resource_production"`
			StorageBonus struct {
				Wood  int `json:"wood"`
				Stone int `json:"stone"`
				Food  int `json:"food"`
				Gold  int `json:"gold"`
			} `json:"storage_bonus"`
			MilitaryBonus struct {
				Attack  int `json:"attack"`
				Defense int `json:"defense"`
			} `json:"military_bonus"`
		}{
			PopulationBonus: 10,
			ResourceProduction: struct {
				Wood  int `json:"wood"`
				Stone int `json:"stone"`
				Food  int `json:"food"`
				Gold  int `json:"gold"`
			}{
				Wood:  5,
				Stone: 3,
				Food:  8,
				Gold:  2,
			},
			StorageBonus: struct {
				Wood  int `json:"wood"`
				Stone int `json:"stone"`
				Food  int `json:"food"`
				Gold  int `json:"gold"`
			}{
				Wood:  100,
				Stone: 100,
				Food:  100,
				Gold:  50,
			},
			MilitaryBonus: struct {
				Attack  int `json:"attack"`
				Defense int `json:"defense"`
			}{
				Attack:  2,
				Defense: 2,
			},
		},
	},
	"warehouse": {
		Type:        "warehouse",
		Name:        "Almacén",
		Description: "Almacena madera y piedra",
		MaxLevel:    20,
		Cost: struct {
			Wood  int `json:"wood"`
			Stone int `json:"stone"`
			Food  int `json:"food"`
			Gold  int `json:"gold"`
		}{
			Wood:  80,
			Stone: 60,
			Food:  30,
			Gold:  10,
		},
		UpgradeTime: 180,
		Effects: struct {
			PopulationBonus    int `json:"population_bonus"`
			ResourceProduction struct {
				Wood  int `json:"wood"`
				Stone int `json:"stone"`
				Food  int `json:"food"`
				Gold  int `json:"gold"`
			} `json:"resource_production"`
			StorageBonus struct {
				Wood  int `json:"wood"`
				Stone int `json:"stone"`
				Food  int `json:"food"`
				Gold  int `json:"gold"`
			} `json:"storage_bonus"`
			MilitaryBonus struct {
				Attack  int `json:"attack"`
				Defense int `json:"defense"`
			} `json:"military_bonus"`
		}{
			PopulationBonus: 0,
			ResourceProduction: struct {
				Wood  int `json:"wood"`
				Stone int `json:"stone"`
				Food  int `json:"food"`
				Gold  int `json:"gold"`
			}{
				Wood:  0,
				Stone: 0,
				Food:  0,
				Gold:  0,
			},
			StorageBonus: struct {
				Wood  int `json:"wood"`
				Stone int `json:"stone"`
				Food  int `json:"food"`
				Gold  int `json:"gold"`
			}{
				Wood:  200,
				Stone: 200,
				Food:  0,
				Gold:  0,
			},
			MilitaryBonus: struct {
				Attack  int `json:"attack"`
				Defense int `json:"defense"`
			}{
				Attack:  0,
				Defense: 0,
			},
		},
	},
	"granary": {
		Type:        "granary",
		Name:        "Granero",
		Description: "Almacena comida",
		MaxLevel:    20,
		Cost: struct {
			Wood  int `json:"wood"`
			Stone int `json:"stone"`
			Food  int `json:"food"`
			Gold  int `json:"gold"`
		}{
			Wood:  60,
			Stone: 80,
			Food:  40,
			Gold:  10,
		},
		UpgradeTime: 180,
		Effects: struct {
			PopulationBonus    int `json:"population_bonus"`
			ResourceProduction struct {
				Wood  int `json:"wood"`
				Stone int `json:"stone"`
				Food  int `json:"food"`
				Gold  int `json:"gold"`
			} `json:"resource_production"`
			StorageBonus struct {
				Wood  int `json:"wood"`
				Stone int `json:"stone"`
				Food  int `json:"food"`
				Gold  int `json:"gold"`
			} `json:"storage_bonus"`
			MilitaryBonus struct {
				Attack  int `json:"attack"`
				Defense int `json:"defense"`
			} `json:"military_bonus"`
		}{
			PopulationBonus: 0,
			ResourceProduction: struct {
				Wood  int `json:"wood"`
				Stone int `json:"stone"`
				Food  int `json:"food"`
				Gold  int `json:"gold"`
			}{
				Wood:  0,
				Stone: 0,
				Food:  0,
				Gold:  0,
			},
			StorageBonus: struct {
				Wood  int `json:"wood"`
				Stone int `json:"stone"`
				Food  int `json:"food"`
				Gold  int `json:"gold"`
			}{
				Wood:  0,
				Stone: 0,
				Food:  200,
				Gold:  0,
			},
			MilitaryBonus: struct {
				Attack  int `json:"attack"`
				Defense int `json:"defense"`
			}{
				Attack:  0,
				Defense: 0,
			},
		},
	},
	"marketplace": {
		Type:        "marketplace",
		Name:        "Mercado",
		Description: "Permite el comercio entre aldeas",
		MaxLevel:    10,
		Cost: struct {
			Wood  int `json:"wood"`
			Stone int `json:"stone"`
			Food  int `json:"food"`
			Gold  int `json:"gold"`
		}{
			Wood:  120,
			Stone: 100,
			Food:  80,
			Gold:  50,
		},
		UpgradeTime: 240,
		Effects: struct {
			PopulationBonus    int `json:"population_bonus"`
			ResourceProduction struct {
				Wood  int `json:"wood"`
				Stone int `json:"stone"`
				Food  int `json:"food"`
				Gold  int `json:"gold"`
			} `json:"resource_production"`
			StorageBonus struct {
				Wood  int `json:"wood"`
				Stone int `json:"stone"`
				Food  int `json:"food"`
				Gold  int `json:"gold"`
			} `json:"storage_bonus"`
			MilitaryBonus struct {
				Attack  int `json:"attack"`
				Defense int `json:"defense"`
			} `json:"military_bonus"`
		}{
			PopulationBonus: 5,
			ResourceProduction: struct {
				Wood  int `json:"wood"`
				Stone int `json:"stone"`
				Food  int `json:"food"`
				Gold  int `json:"gold"`
			}{
				Wood:  0,
				Stone: 0,
				Food:  0,
				Gold:  5,
			},
			StorageBonus: struct {
				Wood  int `json:"wood"`
				Stone int `json:"stone"`
				Food  int `json:"food"`
				Gold  int `json:"gold"`
			}{
				Wood:  0,
				Stone: 0,
				Food:  0,
				Gold:  100,
			},
			MilitaryBonus: struct {
				Attack  int `json:"attack"`
				Defense int `json:"defense"`
			}{
				Attack:  0,
				Defense: 0,
			},
		},
	},
	"barracks": {
		Type:        "barracks",
		Name:        "Cuartel",
		Description: "Entrena unidades de infantería",
		MaxLevel:    20,
		Cost: struct {
			Wood  int `json:"wood"`
			Stone int `json:"stone"`
			Food  int `json:"food"`
			Gold  int `json:"gold"`
		}{
			Wood:  100,
			Stone: 80,
			Food:  60,
			Gold:  30,
		},
		UpgradeTime: 300,
		Effects: struct {
			PopulationBonus    int `json:"population_bonus"`
			ResourceProduction struct {
				Wood  int `json:"wood"`
				Stone int `json:"stone"`
				Food  int `json:"food"`
				Gold  int `json:"gold"`
			} `json:"resource_production"`
			StorageBonus struct {
				Wood  int `json:"wood"`
				Stone int `json:"stone"`
				Food  int `json:"food"`
				Gold  int `json:"gold"`
			} `json:"storage_bonus"`
			MilitaryBonus struct {
				Attack  int `json:"attack"`
				Defense int `json:"defense"`
			} `json:"military_bonus"`
		}{
			PopulationBonus: 0,
			ResourceProduction: struct {
				Wood  int `json:"wood"`
				Stone int `json:"stone"`
				Food  int `json:"food"`
				Gold  int `json:"gold"`
			}{
				Wood:  0,
				Stone: 0,
				Food:  0,
				Gold:  0,
			},
			StorageBonus: struct {
				Wood  int `json:"wood"`
				Stone int `json:"stone"`
				Food  int `json:"food"`
				Gold  int `json:"gold"`
			}{
				Wood:  0,
				Stone: 0,
				Food:  0,
				Gold:  0,
			},
			MilitaryBonus: struct {
				Attack  int `json:"attack"`
				Defense int `json:"defense"`
			}{
				Attack:  3,
				Defense: 2,
			},
		},
	},
	"wood_cutter": {
		Type:        "wood_cutter",
		Name:        "Aserradero",
		Description: "Produce madera para construcción",
		MaxLevel:    20,
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
		UpgradeTime: 180,
		Effects: struct {
			PopulationBonus    int `json:"population_bonus"`
			ResourceProduction struct {
				Wood  int `json:"wood"`
				Stone int `json:"stone"`
				Food  int `json:"food"`
				Gold  int `json:"gold"`
			} `json:"resource_production"`
			StorageBonus struct {
				Wood  int `json:"wood"`
				Stone int `json:"stone"`
				Food  int `json:"food"`
				Gold  int `json:"gold"`
			} `json:"storage_bonus"`
			MilitaryBonus struct {
				Attack  int `json:"attack"`
				Defense int `json:"defense"`
			} `json:"military_bonus"`
		}{
			PopulationBonus: 0,
			ResourceProduction: struct {
				Wood  int `json:"wood"`
				Stone int `json:"stone"`
				Food  int `json:"food"`
				Gold  int `json:"gold"`
			}{
				Wood:  10,
				Stone: 0,
				Food:  0,
				Gold:  0,
			},
			StorageBonus: struct {
				Wood  int `json:"wood"`
				Stone int `json:"stone"`
				Food  int `json:"food"`
				Gold  int `json:"gold"`
			}{
				Wood:  0,
				Stone: 0,
				Food:  0,
				Gold:  0,
			},
			MilitaryBonus: struct {
				Attack  int `json:"attack"`
				Defense int `json:"defense"`
			}{
				Attack:  0,
				Defense: 0,
			},
		},
	},
	"stone_quarry": {
		Type:        "stone_quarry",
		Name:        "Cantera",
		Description: "Extrae piedra para construcción",
		MaxLevel:    20,
		Cost: struct {
			Wood  int `json:"wood"`
			Stone int `json:"stone"`
			Food  int `json:"food"`
			Gold  int `json:"gold"`
		}{
			Wood:  40,
			Stone: 50,
			Food:  25,
			Gold:  10,
		},
		UpgradeTime: 180,
		Effects: struct {
			PopulationBonus    int `json:"population_bonus"`
			ResourceProduction struct {
				Wood  int `json:"wood"`
				Stone int `json:"stone"`
				Food  int `json:"food"`
				Gold  int `json:"gold"`
			} `json:"resource_production"`
			StorageBonus struct {
				Wood  int `json:"wood"`
				Stone int `json:"stone"`
				Food  int `json:"food"`
				Gold  int `json:"gold"`
			} `json:"storage_bonus"`
			MilitaryBonus struct {
				Attack  int `json:"attack"`
				Defense int `json:"defense"`
			} `json:"military_bonus"`
		}{
			PopulationBonus: 0,
			ResourceProduction: struct {
				Wood  int `json:"wood"`
				Stone int `json:"stone"`
				Food  int `json:"food"`
				Gold  int `json:"gold"`
			}{
				Wood:  0,
				Stone: 8,
				Food:  0,
				Gold:  0,
			},
			StorageBonus: struct {
				Wood  int `json:"wood"`
				Stone int `json:"stone"`
				Food  int `json:"food"`
				Gold  int `json:"gold"`
			}{
				Wood:  0,
				Stone: 0,
				Food:  0,
				Gold:  0,
			},
			MilitaryBonus: struct {
				Attack  int `json:"attack"`
				Defense int `json:"defense"`
			}{
				Attack:  0,
				Defense: 0,
			},
		},
	},
	"farm": {
		Type:        "farm",
		Name:        "Granja",
		Description: "Produce comida para la población",
		MaxLevel:    20,
		Cost: struct {
			Wood  int `json:"wood"`
			Stone int `json:"stone"`
			Food  int `json:"food"`
			Gold  int `json:"gold"`
		}{
			Wood:  60,
			Stone: 40,
			Food:  30,
			Gold:  10,
		},
		UpgradeTime: 180,
		Effects: struct {
			PopulationBonus    int `json:"population_bonus"`
			ResourceProduction struct {
				Wood  int `json:"wood"`
				Stone int `json:"stone"`
				Food  int `json:"food"`
				Gold  int `json:"gold"`
			} `json:"resource_production"`
			StorageBonus struct {
				Wood  int `json:"wood"`
				Stone int `json:"stone"`
				Food  int `json:"food"`
				Gold  int `json:"gold"`
			} `json:"storage_bonus"`
			MilitaryBonus struct {
				Attack  int `json:"attack"`
				Defense int `json:"defense"`
			} `json:"military_bonus"`
		}{
			PopulationBonus: 0,
			ResourceProduction: struct {
				Wood  int `json:"wood"`
				Stone int `json:"stone"`
				Food  int `json:"food"`
				Gold  int `json:"gold"`
			}{
				Wood:  0,
				Stone: 0,
				Food:  12,
				Gold:  0,
			},
			StorageBonus: struct {
				Wood  int `json:"wood"`
				Stone int `json:"stone"`
				Food  int `json:"food"`
				Gold  int `json:"gold"`
			}{
				Wood:  0,
				Stone: 0,
				Food:  0,
				Gold:  0,
			},
			MilitaryBonus: struct {
				Attack  int `json:"attack"`
				Defense int `json:"defense"`
			}{
				Attack:  0,
				Defense: 0,
			},
		},
	},
	"gold_mine": {
		Type:        "gold_mine",
		Name:        "Mina de Oro",
		Description: "Extrae oro para el comercio",
		MaxLevel:    20,
		Cost: struct {
			Wood  int `json:"wood"`
			Stone int `json:"stone"`
			Food  int `json:"food"`
			Gold  int `json:"gold"`
		}{
			Wood:  80,
			Stone: 60,
			Food:  40,
			Gold:  20,
		},
		UpgradeTime: 240,
		Effects: struct {
			PopulationBonus    int `json:"population_bonus"`
			ResourceProduction struct {
				Wood  int `json:"wood"`
				Stone int `json:"stone"`
				Food  int `json:"food"`
				Gold  int `json:"gold"`
			} `json:"resource_production"`
			StorageBonus struct {
				Wood  int `json:"wood"`
				Stone int `json:"stone"`
				Food  int `json:"food"`
				Gold  int `json:"gold"`
			} `json:"storage_bonus"`
			MilitaryBonus struct {
				Attack  int `json:"attack"`
				Defense int `json:"defense"`
			} `json:"military_bonus"`
		}{
			PopulationBonus: 0,
			ResourceProduction: struct {
				Wood  int `json:"wood"`
				Stone int `json:"stone"`
				Food  int `json:"food"`
				Gold  int `json:"gold"`
			}{
				Wood:  0,
				Stone: 0,
				Food:  0,
				Gold:  3,
			},
			StorageBonus: struct {
				Wood  int `json:"wood"`
				Stone int `json:"stone"`
				Food  int `json:"food"`
				Gold  int `json:"gold"`
			}{
				Wood:  0,
				Stone: 0,
				Food:  0,
				Gold:  0,
			},
			MilitaryBonus: struct {
				Attack  int `json:"attack"`
				Defense int `json:"defense"`
			}{
				Attack:  0,
				Defense: 0,
			},
		},
	},
}

// ResourceCosts representa los costos de recursos
type ResourceCosts struct {
	Wood  int `json:"wood"`
	Stone int `json:"stone"`
	Food  int `json:"food"`
	Gold  int `json:"gold"`
}

// BuildingUpgradeResult representa el resultado de una mejora de edificio
type BuildingUpgradeResult struct {
	BuildingType   string        `json:"building_type"`
	NewLevel       int           `json:"new_level"`
	UpgradeTime    time.Duration `json:"upgrade_time"`
	CompletionTime time.Time     `json:"completion_time"`
	Costs          ResourceCosts `json:"costs"`
	ResourcesSpent ResourceCosts `json:"resources_spent"`
}

// BuildingUpgradeInfo representa la información de mejora de un edificio
type BuildingUpgradeInfo struct {
	BuildingType                 string        `json:"building_type"`
	CurrentLevel                 int           `json:"current_level"`
	NextLevel                    int           `json:"next_level"`
	MaxLevel                     int           `json:"max_level"`
	UpgradeCosts                 ResourceCosts `json:"upgrade_costs"`
	UpgradeTime                  time.Duration `json:"upgrade_time"`
	CanAfford                    bool          `json:"can_afford"`
	IsUpgrading                  bool          `json:"is_upgrading"`
	UpgradeCompletionTime        *time.Time    `json:"upgrade_completion_time"`
	TownHallRequirement          int           `json:"town_hall_requirement"`
	MeetsTownHallRequirement     bool          `json:"meets_town_hall_requirement"`
	CanUpgrade                   bool          `json:"can_upgrade"`
	ProductionIncrease           int           `json:"production_increase,omitempty"`
	StorageIncrease              int           `json:"storage_increase,omitempty"`
	TrainingSpeedImprovement     float64       `json:"training_speed_improvement,omitempty"`
	ConstructionSpeedImprovement float64       `json:"construction_speed_improvement,omitempty"`
}
