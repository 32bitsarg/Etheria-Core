package services

import (
	"fmt"
	"server-backend/models"
	"server-backend/repository"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// BuildingRequirementsEngine maneja toda la lógica de requisitos de construcción en Go
type BuildingRequirementsEngine struct {
	villageRepo        *repository.VillageRepository
	buildingConfigRepo *repository.BuildingConfigRepository
	researchRepo       *repository.ResearchRepository
	allianceRepo       *repository.AllianceRepository
	logger             *zap.Logger
}

// BuildingRequirement define los requisitos para construir un edificio
type BuildingRequirement struct {
	Type                 string                  `json:"type"`
	RequiredLevel        int                     `json:"required_level"`
	RequiredTechnologies []TechnologyRequirement `json:"required_technologies"`
	RequiredBuildings    []BuildingDependency    `json:"required_buildings"`
	TownHallLevel        int                     `json:"town_hall_level"`
}

// TechnologyRequirement define requisitos de tecnología
type TechnologyRequirement struct {
	TechnologyID  string `json:"technology_id"`
	RequiredLevel int    `json:"required_level"`
}

// BuildingDependency define dependencias de edificios
type BuildingDependency struct {
	BuildingType  string `json:"building_type"`
	RequiredLevel int    `json:"required_level"`
}

// BuildingRequirementsResultGo representa el resultado de verificar requisitos (formato Go)
type BuildingRequirementsResultGo struct {
	CanBuild            bool              `json:"can_build"`
	MissingRequirements []string          `json:"missing_requirements"`
	CostWood            int               `json:"cost_wood"`
	CostStone           int               `json:"cost_stone"`
	CostFood            int               `json:"cost_food"`
	CostGold            int               `json:"cost_gold"`
	AllianceBonuses     AllianceBonuses   `json:"alliance_bonuses"`
	TechnologyBonuses   TechnologyBonuses `json:"technology_bonuses"`
}

// AllianceBonuses representa bonificaciones de alianza
type AllianceBonuses struct {
	ConstructionSpeed    float64 `json:"construction_speed"`
	UpgradeCostReduction float64 `json:"upgrade_cost_reduction"`
	ResourceBonus        float64 `json:"resource_bonus"`
}

// TechnologyBonuses representa bonificaciones de tecnología
type TechnologyBonuses struct {
	ConstructionSpeed    float64 `json:"construction_speed"`
	UpgradeCostReduction float64 `json:"upgrade_cost_reduction"`
	ResourceBonus        float64 `json:"resource_bonus"`
}

// NewBuildingRequirementsEngine crea una nueva instancia del motor de requisitos
func NewBuildingRequirementsEngine(
	villageRepo *repository.VillageRepository,
	buildingConfigRepo *repository.BuildingConfigRepository,
	researchRepo *repository.ResearchRepository,
	allianceRepo *repository.AllianceRepository,
	logger *zap.Logger,
) *BuildingRequirementsEngine {
	return &BuildingRequirementsEngine{
		villageRepo:        villageRepo,
		buildingConfigRepo: buildingConfigRepo,
		researchRepo:       researchRepo,
		allianceRepo:       allianceRepo,
		logger:             logger,
	}
}

// CheckBuildingRequirements verifica los requisitos para construir un edificio usando lógica Go
func (e *BuildingRequirementsEngine) CheckBuildingRequirements(villageID uuid.UUID, buildingType string, targetLevel int) (*BuildingRequirementsResultGo, error) {
	e.logger.Info("Verificando requisitos de construcción",
		zap.String("village_id", villageID.String()),
		zap.String("building_type", buildingType),
		zap.Int("target_level", targetLevel),
	)

	// 1. Obtener información de la aldea
	village, err := e.villageRepo.GetVillageByID(villageID)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo aldea: %w", err)
	}
	if village == nil {
		return nil, fmt.Errorf("aldea no encontrada")
	}

	// 2. Obtener configuración del edificio
	config, err := e.buildingConfigRepo.GetBuildingConfig(buildingType, targetLevel)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo configuración del edificio: %w", err)
	}
	if config == nil {
		return nil, fmt.Errorf("configuración de edificio no encontrada para %s nivel %d", buildingType, targetLevel)
	}

	// 3. Verificar requisitos básicos
	missingRequirements := []string{}

	// Verificar requisitos de ayuntamiento
	if err := e.checkTownHallRequirements(village, buildingType, targetLevel); err != nil {
		missingRequirements = append(missingRequirements, err.Error())
	}

	// Verificar requisitos de edificios dependientes
	if missing := e.checkBuildingDependencies(village, buildingType, targetLevel); len(missing) > 0 {
		missingRequirements = append(missingRequirements, missing...)
	}

	// Verificar requisitos de tecnologías
	if missing := e.checkTechnologyRequirements(village.Village.PlayerID, buildingType, targetLevel); len(missing) > 0 {
		missingRequirements = append(missingRequirements, missing...)
	}

	// 4. Calcular bonificaciones
	allianceBonuses := e.calculateAllianceBonuses(village.Village.PlayerID)
	technologyBonuses := e.calculateTechnologyBonuses(village.Village.PlayerID)

	// 5. Aplicar bonificaciones a los costos
	finalCosts := e.applyBonusesToCosts(config, allianceBonuses, technologyBonuses)

	// 6. Verificar recursos disponibles
	if !e.hasEnoughResources(village.Resources, finalCosts) {
		missingRequirements = append(missingRequirements, "recursos insuficientes")
	}

	result := &BuildingRequirementsResultGo{
		CanBuild:            len(missingRequirements) == 0,
		MissingRequirements: missingRequirements,
		CostWood:            finalCosts.Wood,
		CostStone:           finalCosts.Stone,
		CostFood:            finalCosts.Food,
		CostGold:            finalCosts.Gold,
		AllianceBonuses:     allianceBonuses,
		TechnologyBonuses:   technologyBonuses,
	}

	e.logger.Info("Verificación de requisitos completada",
		zap.String("village_id", villageID.String()),
		zap.String("building_type", buildingType),
		zap.Int("target_level", targetLevel),
		zap.Bool("can_build", result.CanBuild),
		zap.Strings("missing_requirements", missingRequirements),
	)

	return result, nil
}

// checkTownHallRequirements verifica los requisitos del ayuntamiento
func (e *BuildingRequirementsEngine) checkTownHallRequirements(village *models.VillageWithDetails, buildingType string, targetLevel int) error {
	townHall, exists := village.Buildings["town_hall"]
	if !exists {
		return fmt.Errorf("se requiere ayuntamiento")
	}

	// Requisitos básicos por tipo de edificio
	var requiredLevel int
	switch buildingType {
	case "warehouse", "granary":
		requiredLevel = 1
	case "marketplace":
		requiredLevel = 3
	case "barracks":
		requiredLevel = 5
	case "wood_cutter", "stone_quarry", "farm", "gold_mine":
		requiredLevel = 2
	default:
		requiredLevel = 1
	}

	// Verificar nivel mínimo del ayuntamiento
	if townHall.Level < requiredLevel {
		return fmt.Errorf("se requiere ayuntamiento nivel %d, actual: %d", requiredLevel, townHall.Level)
	}

	// Verificar que el ayuntamiento permita el nivel objetivo
	maxBuildingLevel := townHall.Level * 2 // Regla: nivel máximo = ayuntamiento * 2
	if targetLevel > maxBuildingLevel {
		return fmt.Errorf("el ayuntamiento nivel %d solo permite edificios hasta nivel %d", townHall.Level, maxBuildingLevel)
	}

	return nil
}

// checkBuildingDependencies verifica las dependencias de otros edificios
func (e *BuildingRequirementsEngine) checkBuildingDependencies(village *models.VillageWithDetails, buildingType string, targetLevel int) []string {
	missing := []string{}

	// Definir dependencias específicas
	dependencies := e.getBuildingDependencies(buildingType, targetLevel)

	for _, dep := range dependencies {
		building, exists := village.Buildings[dep.BuildingType]
		if !exists || building.Level < dep.RequiredLevel {
			missing = append(missing, fmt.Sprintf("se requiere %s nivel %d", dep.BuildingType, dep.RequiredLevel))
		}
	}

	return missing
}

// getBuildingDependencies obtiene las dependencias específicas de un edificio
func (e *BuildingRequirementsEngine) getBuildingDependencies(buildingType string, targetLevel int) []BuildingDependency {
	dependencies := []BuildingDependency{}

	switch buildingType {
	case "marketplace":
		if targetLevel >= 2 {
			dependencies = append(dependencies, BuildingDependency{
				BuildingType:  "warehouse",
				RequiredLevel: 2,
			})
		}
	case "barracks":
		dependencies = append(dependencies, BuildingDependency{
			BuildingType:  "warehouse",
			RequiredLevel: 1,
		})
	case "wood_cutter", "stone_quarry", "farm", "gold_mine":
		if targetLevel >= 3 {
			dependencies = append(dependencies, BuildingDependency{
				BuildingType:  "warehouse",
				RequiredLevel: 2,
			})
		}
	}

	return dependencies
}

// checkTechnologyRequirements verifica los requisitos de tecnologías
func (e *BuildingRequirementsEngine) checkTechnologyRequirements(playerID uuid.UUID, buildingType string, targetLevel int) []string {
	missing := []string{}

	// Obtener tecnologías requeridas para este edificio y nivel
	requiredTechs := e.getRequiredTechnologies(buildingType, targetLevel)

	for _, techReq := range requiredTechs {
		// Verificar si el jugador tiene la tecnología al nivel requerido
		playerTech, err := e.researchRepo.GetPlayerTechnology(playerID.String(), techReq.TechnologyID)
		if err != nil || playerTech == nil || playerTech.Level < techReq.RequiredLevel {
			// Obtener nombre de la tecnología para el mensaje de error
			tech, _ := e.researchRepo.GetTechnology(techReq.TechnologyID)
			techName := "Tecnología desconocida"
			if tech != nil {
				techName = tech.Name
			}
			missing = append(missing, fmt.Sprintf("se requiere %s nivel %d", techName, techReq.RequiredLevel))
		}
	}

	return missing
}

// getRequiredTechnologies obtiene las tecnologías requeridas para un edificio
func (e *BuildingRequirementsEngine) getRequiredTechnologies(buildingType string, targetLevel int) []TechnologyRequirement {
	requirements := []TechnologyRequirement{}

	// Definir tecnologías requeridas por tipo de edificio
	switch buildingType {
	case "barracks":
		if targetLevel >= 2 {
			requirements = append(requirements, TechnologyRequirement{
				TechnologyID:  "military_tactics",
				RequiredLevel: 1,
			})
		}
	case "marketplace":
		if targetLevel >= 3 {
			requirements = append(requirements, TechnologyRequirement{
				TechnologyID:  "trade_routes",
				RequiredLevel: 2,
			})
		}
	case "warehouse", "granary":
		if targetLevel >= 5 {
			requirements = append(requirements, TechnologyRequirement{
				TechnologyID:  "advanced_storage",
				RequiredLevel: 3,
			})
		}
	}

	return requirements
}

// calculateAllianceBonuses calcula las bonificaciones de alianza
func (e *BuildingRequirementsEngine) calculateAllianceBonuses(playerID uuid.UUID) AllianceBonuses {
	bonuses := AllianceBonuses{
		ConstructionSpeed:    0.0,
		UpgradeCostReduction: 0.0,
		ResourceBonus:        0.0,
	}

	// Obtener alianza del jugador
	alliance, err := e.allianceRepo.GetPlayerAlliance(int(playerID.ID()))
	if err != nil || alliance == nil {
		return bonuses // Sin alianza = sin bonificaciones
	}

	// Calcular bonificaciones basadas en nivel de alianza
	bonuses.ConstructionSpeed = float64(alliance.Level) * 0.015   // 1.5% por nivel
	bonuses.UpgradeCostReduction = float64(alliance.Level) * 0.01 // 1% por nivel
	bonuses.ResourceBonus = float64(alliance.Level) * 0.02        // 2% por nivel

	return bonuses
}

// calculateTechnologyBonuses calcula las bonificaciones de tecnología
func (e *BuildingRequirementsEngine) calculateTechnologyBonuses(playerID uuid.UUID) TechnologyBonuses {
	bonuses := TechnologyBonuses{
		ConstructionSpeed:    0.0,
		UpgradeCostReduction: 0.0,
		ResourceBonus:        0.0,
	}

	// Obtener tecnologías del jugador que afectan construcción
	technologies, err := e.researchRepo.GetPlayerTechnologies(playerID.String())
	if err != nil {
		return bonuses
	}

	for _, tech := range technologies {
		// Procesar efectos de tecnologías específicas
		// Por ahora, usar bonificaciones básicas basadas en el nivel
		bonuses.ConstructionSpeed += float64(tech.Level) * 0.01     // 1% por nivel
		bonuses.UpgradeCostReduction += float64(tech.Level) * 0.005 // 0.5% por nivel
		bonuses.ResourceBonus += float64(tech.Level) * 0.01         // 1% por nivel
	}

	return bonuses
}

// applyBonusesToCosts aplica las bonificaciones a los costos
func (e *BuildingRequirementsEngine) applyBonusesToCosts(config *models.BuildingConfig, allianceBonuses AllianceBonuses, technologyBonuses TechnologyBonuses) models.ResourceCostsLegacy {
	// Calcular reducción total de costos
	totalCostReduction := allianceBonuses.UpgradeCostReduction + technologyBonuses.UpgradeCostReduction

	// Aplicar reducción (máximo 50% de reducción)
	if totalCostReduction > 0.5 {
		totalCostReduction = 0.5
	}

	costMultiplier := 1.0 - totalCostReduction

	return models.ResourceCostsLegacy{
		Wood:  int(float64(config.WoodCost) * costMultiplier),
		Stone: int(float64(config.StoneCost) * costMultiplier),
		Food:  int(float64(config.FoodCost) * costMultiplier),
		Gold:  int(float64(config.GoldCost) * costMultiplier),
	}
}

// hasEnoughResources verifica si hay suficientes recursos
func (e *BuildingRequirementsEngine) hasEnoughResources(resources models.Resources, costs models.ResourceCostsLegacy) bool {
	return resources.Wood >= costs.Wood &&
		resources.Stone >= costs.Stone &&
		resources.Food >= costs.Food &&
		resources.Gold >= costs.Gold
}

// GetBuildingRequirementsInfo obtiene información detallada sobre los requisitos de un edificio
func (e *BuildingRequirementsEngine) GetBuildingRequirementsInfo(buildingType string, targetLevel int) (*BuildingRequirement, error) {
	requirement := &BuildingRequirement{
		Type:                 buildingType,
		RequiredLevel:        targetLevel,
		RequiredTechnologies: e.getRequiredTechnologies(buildingType, targetLevel),
		RequiredBuildings:    e.getBuildingDependencies(buildingType, targetLevel),
		TownHallLevel:        e.getRequiredTownHallLevel(buildingType, targetLevel),
	}

	return requirement, nil
}

// getRequiredTownHallLevel obtiene el nivel mínimo requerido del ayuntamiento
func (e *BuildingRequirementsEngine) getRequiredTownHallLevel(buildingType string, targetLevel int) int {
	switch buildingType {
	case "warehouse", "granary":
		return 1
	case "marketplace":
		return 3
	case "barracks":
		return 5
	case "wood_cutter", "stone_quarry", "farm", "gold_mine":
		return 2
	default:
		return 1
	}
}
