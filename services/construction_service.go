package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"server-backend/models"
	"server-backend/repository"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

var (
	ErrInsufficientResources  = errors.New("recursos insuficientes")
	ErrBuildingMaxLevel       = errors.New("el edificio ya está en su nivel máximo")
	ErrBuildingUpgrading      = errors.New("el edificio ya está siendo mejorado")
	ErrInvalidBuildingType    = errors.New("tipo de edificio inválido")
	ErrTownHallRequired       = errors.New("se requiere un ayuntamiento de nivel superior")
	ErrBuildingConfigNotFound = errors.New("configuración de edificio no encontrada")
	ErrRequirementsNotMet     = errors.New("no se cumplen los requisitos para construir")
)

type ConstructionService struct {
	villageRepo        *repository.VillageRepository
	buildingConfigRepo *repository.BuildingConfigRepository
	redisService       *RedisService
	logger             *zap.Logger
	timeZone           string
}

type ConstructionQueueItem struct {
	ID         int64     `json:"id"`
	PlayerID   int64     `json:"player_id"`
	VillageID  int64     `json:"village_id"`
	BuildingID int64     `json:"building_id"`
	Level      int       `json:"level"`
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`
	Status     string    `json:"status"` // "queued", "in_progress", "completed", "cancelled"
}

// BuildingRequirementsResult representa el resultado de verificar requisitos
type BuildingRequirementsResult struct {
	CanBuild            bool     `json:"can_build"`
	MissingRequirements []string `json:"missing_requirements"`
	CostWood            int      `json:"cost_wood"`
	CostStone           int      `json:"cost_stone"`
	CostFood            int      `json:"cost_food"`
	CostGold            int      `json:"cost_gold"`
}

// ConstructionResult representa el resultado de procesar la cola de construcción
type ConstructionResult struct {
	BuildingType     string          `json:"building_type"`
	OldLevel         int             `json:"old_level"`
	NewLevel         int             `json:"new_level"`
	ConstructionTime int             `json:"construction_time"`
	ResourcesSpent   json.RawMessage `json:"resources_spent"`
}

func NewConstructionService(villageRepo *repository.VillageRepository, buildingConfigRepo *repository.BuildingConfigRepository, redisService *RedisService, logger *zap.Logger, timeZone string) *ConstructionService {
	return &ConstructionService{
		villageRepo:        villageRepo,
		buildingConfigRepo: buildingConfigRepo,
		redisService:       redisService,
		logger:             logger,
		timeZone:           timeZone,
	}
}

// CheckBuildingRequirements verifica los requisitos para construir usando la función avanzada de la BD
func (s *ConstructionService) CheckBuildingRequirements(villageID uuid.UUID, buildingType string, targetLevel int) (*BuildingRequirementsResult, error) {
	// Usar la función avanzada de la base de datos a través del repositorio
	rows, err := s.villageRepo.CheckBuildingRequirementsAdvanced(villageID, buildingType, targetLevel)
	if err != nil {
		return nil, fmt.Errorf("error verificando requisitos: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, errors.New("no se pudo verificar los requisitos")
	}

	var result BuildingRequirementsResult
	var missingReqs []string
	err = rows.Scan(
		&result.CanBuild,
		&missingReqs,
		&result.CostWood,
		&result.CostStone,
		&result.CostFood,
		&result.CostGold,
	)
	if err != nil {
		return nil, fmt.Errorf("error escaneando resultado: %w", err)
	}

	result.MissingRequirements = missingReqs
	return &result, nil
}

// UpgradeBuilding maneja la mejora de un edificio usando la nueva función avanzada
func (s *ConstructionService) UpgradeBuilding(villageID uuid.UUID, buildingType string) (*models.BuildingUpgradeResult, error) {
	// Obtener la aldea con todos sus detalles
	village, err := s.villageRepo.GetVillageByID(villageID)
	if err != nil {
		return nil, err
	}
	if village == nil {
		return nil, errors.New("aldea no encontrada")
	}

	// Verificar que el edificio existe
	building, exists := village.Buildings[buildingType]
	if !exists {
		return nil, ErrInvalidBuildingType
	}

	// Verificar que no esté ya siendo mejorado
	if building.IsUpgrading {
		return nil, ErrBuildingUpgrading
	}

	// Obtener el nivel máximo disponible
	maxLevel, err := s.buildingConfigRepo.GetMaxLevel(buildingType)
	if err != nil {
		return nil, err
	}

	// Verificar que no esté en el nivel máximo
	if building.Level >= maxLevel {
		return nil, ErrBuildingMaxLevel
	}

	// Verificar requisitos usando la función avanzada
	nextLevel := building.Level + 1
	requirements, err := s.CheckBuildingRequirements(villageID, buildingType, nextLevel)
	if err != nil {
		return nil, err
	}

	if !requirements.CanBuild {
		return nil, fmt.Errorf("%w: %v", ErrRequirementsNotMet, requirements.MissingRequirements)
	}

	// Verificar que hay suficientes recursos
	if !s.hasEnoughResources(village.Resources, models.ResourceCosts{
		Wood:  requirements.CostWood,
		Stone: requirements.CostStone,
		Food:  requirements.CostFood,
		Gold:  requirements.CostGold,
	}) {
		return nil, ErrInsufficientResources
	}

	// Calcular tiempo de construcción con modificadores del ayuntamiento
	baseTime := s.calculateConstructionTime(buildingType, nextLevel)
	townHallLevel := s.getTownHallLevel(village)
	constructionSpeedModifier := s.getConstructionSpeedModifier(townHallLevel)
	upgradeTime := time.Duration(float64(baseTime) * constructionSpeedModifier)

	// Usar la zona horaria configurada
	loc, err := time.LoadLocation(s.timeZone)
	if err != nil {
		loc = time.UTC
	}
	now := time.Now().In(loc)
	completionTime := now.Add(upgradeTime)

	// Iniciar transacción para actualizar edificio y consumir recursos
	err = s.villageRepo.UpdateBuilding(
		villageID,
		buildingType,
		nextLevel,
		true, // isUpgrading
		&completionTime,
	)
	if err != nil {
		return nil, err
	}

	// Consumir recursos
	newWood := village.Resources.Wood - requirements.CostWood
	newStone := village.Resources.Stone - requirements.CostStone
	newFood := village.Resources.Food - requirements.CostFood
	newGold := village.Resources.Gold - requirements.CostGold

	err = s.villageRepo.UpdateResources(villageID, newWood, newStone, newFood, newGold)
	if err != nil {
		return nil, err
	}

	s.logger.Info("Edificio mejorado iniciado",
		zap.String("village_id", villageID.String()),
		zap.String("building_type", buildingType),
		zap.Int("new_level", nextLevel),
		zap.Duration("upgrade_time", upgradeTime),
	)

	return &models.BuildingUpgradeResult{
		BuildingType:   buildingType,
		NewLevel:       nextLevel,
		UpgradeTime:    upgradeTime,
		CompletionTime: completionTime,
		Costs: models.ResourceCosts{
			Wood:  requirements.CostWood,
			Stone: requirements.CostStone,
			Food:  requirements.CostFood,
			Gold:  requirements.CostGold,
		},
		ResourcesSpent: models.ResourceCosts{
			Wood:  requirements.CostWood,
			Stone: requirements.CostStone,
			Food:  requirements.CostFood,
			Gold:  requirements.CostGold,
		},
	}, nil
}

// ProcessConstructionQueue procesa la cola de construcción usando la nueva función de la BD
func (s *ConstructionService) ProcessConstructionQueue(villageID uuid.UUID) ([]ConstructionResult, error) {
	// Usar la función avanzada de la base de datos a través del repositorio
	rows, err := s.villageRepo.ProcessConstructionQueue(villageID)
	if err != nil {
		return nil, fmt.Errorf("error procesando cola de construcción: %w", err)
	}
	defer rows.Close()

	var results []ConstructionResult
	for rows.Next() {
		var result ConstructionResult
		err := rows.Scan(
			&result.BuildingType,
			&result.OldLevel,
			&result.NewLevel,
			&result.ConstructionTime,
			&result.ResourcesSpent,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando resultado: %w", err)
		}
		results = append(results, result)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterando resultados: %w", err)
	}

	s.logger.Info("Cola de construcción procesada",
		zap.String("village_id", villageID.String()),
		zap.Int("buildings_completed", len(results)),
	)

	return results, nil
}

// CompleteUpgrade completa la mejora de un edificio
func (s *ConstructionService) CompleteUpgrade(villageID uuid.UUID, buildingType string) error {
	village, err := s.villageRepo.GetVillageByID(villageID)
	if err != nil {
		return err
	}
	if village == nil {
		return errors.New("aldea no encontrada")
	}

	building, exists := village.Buildings[buildingType]
	if !exists {
		return ErrInvalidBuildingType
	}

	if !building.IsUpgrading {
		return errors.New("el edificio no está siendo mejorado")
	}

	if building.UpgradeCompletionTime == nil || time.Now().Before(*building.UpgradeCompletionTime) {
		return errors.New("la mejora aún no ha terminado")
	}

	// Completar la mejora
	err = s.villageRepo.UpdateBuilding(
		villageID,
		buildingType,
		building.Level,
		false, // isUpgrading
		nil,   // completionTime
	)
	if err != nil {
		return err
	}

	s.logger.Info("Mejora de edificio completada",
		zap.String("village_id", villageID.String()),
		zap.String("building_type", buildingType),
		zap.Int("new_level", building.Level),
	)

	return nil
}

// GetUpgradeInfo obtiene información sobre la mejora de un edificio
func (s *ConstructionService) GetUpgradeInfo(villageID uuid.UUID, buildingType string) (*models.BuildingUpgradeInfo, error) {
	village, err := s.villageRepo.GetVillageByID(villageID)
	if err != nil {
		return nil, err
	}
	if village == nil {
		return nil, errors.New("aldea no encontrada")
	}

	building, exists := village.Buildings[buildingType]
	if !exists {
		return nil, ErrInvalidBuildingType
	}

	// Obtener requisitos para el siguiente nivel
	nextLevel := building.Level + 1
	requirements, err := s.CheckBuildingRequirements(villageID, buildingType, nextLevel)
	if err != nil {
		return nil, err
	}

	// Calcular tiempo de construcción
	baseTime := s.calculateConstructionTime(buildingType, nextLevel)
	townHallLevel := s.getTownHallLevel(village)
	constructionSpeedModifier := s.getConstructionSpeedModifier(townHallLevel)
	upgradeTime := time.Duration(float64(baseTime) * constructionSpeedModifier)

	return &models.BuildingUpgradeInfo{
		BuildingType: buildingType,
		CurrentLevel: building.Level,
		NextLevel:    nextLevel,
		CanUpgrade:   requirements.CanBuild,
		UpgradeTime:  upgradeTime,
		UpgradeCosts: models.ResourceCosts{
			Wood:  requirements.CostWood,
			Stone: requirements.CostStone,
			Food:  requirements.CostFood,
			Gold:  requirements.CostGold,
		},
	}, nil
}

// calculateConstructionTime calcula el tiempo base de construcción
func (s *ConstructionService) calculateConstructionTime(buildingType string, level int) time.Duration {
	// Obtener tiempo base del tipo de edificio
	baseTime := 60 * time.Second // Tiempo base por defecto

	// Aplicar multiplicador por nivel
	levelMultiplier := float64(level) * 0.2
	totalTime := float64(baseTime) * (1 + levelMultiplier)

	return time.Duration(totalTime)
}

// hasEnoughResources verifica si hay suficientes recursos
func (s *ConstructionService) hasEnoughResources(resources models.Resources, costs models.ResourceCosts) bool {
	return resources.Wood >= costs.Wood &&
		resources.Stone >= costs.Stone &&
		resources.Food >= costs.Food &&
		resources.Gold >= costs.Gold
}

// getTownHallLevel obtiene el nivel del ayuntamiento
func (s *ConstructionService) getTownHallLevel(village *models.VillageWithDetails) int {
	if townHall, exists := village.Buildings["town_hall"]; exists {
		return townHall.Level
	}
	return 0
}

// getConstructionSpeedModifier calcula el modificador de velocidad de construcción
func (s *ConstructionService) getConstructionSpeedModifier(townHallLevel int) float64 {
	// Base: 1.0 (sin modificador)
	// Cada nivel del ayuntamiento reduce el tiempo en 5%
	reduction := float64(townHallLevel) * 0.05
	return 1.0 - reduction
}

// validateTownHallRequirement valida el requisito del ayuntamiento
func (s *ConstructionService) validateTownHallRequirement(village *models.VillageWithDetails, buildingType string, targetLevel int) error {
	townHallLevel := s.getTownHallLevel(village)

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

	if townHallLevel < requiredLevel {
		return fmt.Errorf("%w: se requiere ayuntamiento nivel %d, actual: %d",
			ErrTownHallRequired, requiredLevel, townHallLevel)
	}

	return nil
}

// GetConstructionQueue obtiene la cola de construcción de una aldea
func (s *ConstructionService) GetConstructionQueue(villageID uuid.UUID) ([]ConstructionQueueItem, error) {
	// Esta función podría implementarse para obtener la cola desde Redis o la BD
	// Por ahora retornamos una lista vacía
	return []ConstructionQueueItem{}, nil
}

// CancelUpgrade cancela la mejora de un edificio
func (s *ConstructionService) CancelUpgrade(villageID uuid.UUID, buildingType string) error {
	village, err := s.villageRepo.GetVillageByID(villageID)
	if err != nil {
		return err
	}
	if village == nil {
		return errors.New("aldea no encontrada")
	}

	building, exists := village.Buildings[buildingType]
	if !exists {
		return ErrInvalidBuildingType
	}

	if !building.IsUpgrading {
		return errors.New("el edificio no está siendo mejorado")
	}

	// Cancelar la mejora
	err = s.villageRepo.UpdateBuilding(
		villageID,
		buildingType,
		building.Level,
		false, // isUpgrading
		nil,   // completionTime
	)
	if err != nil {
		return err
	}

	s.logger.Info("Mejora de edificio cancelada",
		zap.String("village_id", villageID.String()),
		zap.String("building_type", buildingType),
	)

	return nil
}
