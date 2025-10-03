package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"server-backend/models"
	"server-backend/repository"
	"server-backend/websocket"
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
	ErrConstructionQueueFull  = errors.New("la cola de construcción está llena")
)

// Constantes para límites de construcción
const (
	MaxConstructionSlots = 4  // Máximo de slots de construcción por aldea
	ActiveConstructionSlots = 2 // Slots activos (los otros 2 se habilitan con micropagos)
)

type ConstructionService struct {
	villageRepo        *repository.VillageRepository
	buildingConfigRepo *repository.BuildingConfigRepository
	researchRepo       *repository.ResearchRepository
	allianceRepo       *repository.AllianceRepository
	redisService       *RedisService
	logger             *zap.Logger
	timeZone           string
	requirementsEngine *BuildingRequirementsEngine
	wsManager          *websocket.Manager
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

// BuildingRequirementsResultLegacy representa el resultado de verificar requisitos (formato legacy)
type BuildingRequirementsResultLegacy struct {
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

func NewConstructionService(
	villageRepo *repository.VillageRepository,
	buildingConfigRepo *repository.BuildingConfigRepository,
	researchRepo *repository.ResearchRepository,
	allianceRepo *repository.AllianceRepository,
	redisService *RedisService,
	logger *zap.Logger,
	timeZone string,
) *ConstructionService {
	// Crear el motor de requisitos
	requirementsEngine := NewBuildingRequirementsEngine(
		villageRepo,
		buildingConfigRepo,
		researchRepo,
		allianceRepo,
		logger,
	)

	return &ConstructionService{
		villageRepo:        villageRepo,
		buildingConfigRepo: buildingConfigRepo,
		researchRepo:       researchRepo,
		allianceRepo:       allianceRepo,
		redisService:       redisService,
		logger:             logger,
		timeZone:           timeZone,
		requirementsEngine: requirementsEngine,
		wsManager:          nil, // Se establecerá después con SetWebSocketManager
	}
}

// SetWebSocketManager establece el WebSocket manager para notificaciones en tiempo real
func (s *ConstructionService) SetWebSocketManager(wsManager *websocket.Manager) {
	s.wsManager = wsManager
	s.logger.Info("WebSocket manager configurado en ConstructionService")
}

// CheckBuildingRequirements verifica los requisitos para construir usando la nueva lógica Go
func (s *ConstructionService) CheckBuildingRequirements(villageID uuid.UUID, buildingType string, targetLevel int) (*BuildingRequirementsResultLegacy, error) {
	// Usar el nuevo motor de requisitos en Go
	result, err := s.requirementsEngine.CheckBuildingRequirements(villageID, buildingType, targetLevel)
	if err != nil {
		return nil, fmt.Errorf("error verificando requisitos: %w", err)
	}

	// Convertir el resultado al formato esperado por el handler
	return &BuildingRequirementsResultLegacy{
		CanBuild:            result.CanBuild,
		MissingRequirements: result.MissingRequirements,
		CostWood:            result.CostWood,
		CostStone:           result.CostStone,
		CostFood:            result.CostFood,
		CostGold:            result.CostGold,
	}, nil
}

// UpgradeBuilding maneja la mejora de un edificio usando la nueva función avanzada
func (s *ConstructionService) UpgradeBuilding(villageID uuid.UUID, buildingType string) (*models.BuildingUpgradeResultLegacy, error) {
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

	// Verificar límite de cola de construcción
	canStart, err := s.canStartConstruction(villageID)
	if err != nil {
		return nil, err
	}
	if !canStart {
		return nil, ErrConstructionQueueFull
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
	if !s.hasEnoughResources(village.Resources, models.ResourceCostsLegacy{
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

	// Enviar notificación de inicio de mejora
	costs := models.ResourceCostsLegacy{
		Wood:  requirements.CostWood,
		Stone: requirements.CostStone,
		Food:  requirements.CostFood,
		Gold:  requirements.CostGold,
	}
	s.sendBuildingUpgradeStarted(villageID, buildingType, nextLevel, upgradeTime, completionTime, costs)

	return &models.BuildingUpgradeResultLegacy{
		BuildingType:   buildingType,
		NewLevel:       nextLevel,
		UpgradeTime:    upgradeTime,
		CompletionTime: completionTime,
		Costs: models.ResourceCostsLegacy{
			Wood:  requirements.CostWood,
			Stone: requirements.CostStone,
			Food:  requirements.CostFood,
			Gold:  requirements.CostGold,
		},
		ResourcesSpent: models.ResourceCostsLegacy{
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

	// Enviar notificación de finalización de mejora
	s.sendBuildingUpgradeCompleted(villageID, buildingType, building.Level)

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
func (s *ConstructionService) hasEnoughResources(resources models.Resources, costs models.ResourceCostsLegacy) bool {
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

// CancelUpgradeWithRefund cancela la mejora de un edificio y devuelve el 50% de los recursos
func (s *ConstructionService) CancelUpgradeWithRefund(villageID uuid.UUID, buildingType string) (*models.CancelUpgradeResult, error) {
	s.logger.Info("Iniciando cancelación de mejora con reembolso",
		zap.String("village_id", villageID.String()),
		zap.String("building_type", buildingType),
	)

	// 1. Obtener información de la aldea
	village, err := s.villageRepo.GetVillageByID(villageID)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo aldea: %w", err)
	}
	if village == nil {
		return nil, errors.New("aldea no encontrada")
	}

	// 2. Verificar que el edificio existe y está siendo mejorado
	building, exists := village.Buildings[buildingType]
	if !exists {
		return nil, ErrInvalidBuildingType
	}

	if !building.IsUpgrading {
		return nil, errors.New("el edificio no está siendo mejorado")
	}

	// 3. Calcular recursos gastados (nivel actual + 1)
	targetLevel := building.Level + 1
	config, err := s.buildingConfigRepo.GetBuildingConfig(buildingType, targetLevel)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo configuración del edificio: %w", err)
	}
	if config == nil {
		return nil, fmt.Errorf("configuración de edificio no encontrada para %s nivel %d", buildingType, targetLevel)
	}

	// 4. Calcular reembolso (50% de recursos gastados)
	refundPercentage := 0.5
	refundAmount := models.ResourceCostsLegacy{
		Wood:  int(float64(config.WoodCost) * refundPercentage),
		Stone: int(float64(config.StoneCost) * refundPercentage),
		Food:  int(float64(config.FoodCost) * refundPercentage),
		Gold:  int(float64(config.GoldCost) * refundPercentage),
	}

	originalCost := models.ResourceCostsLegacy{
		Wood:  config.WoodCost,
		Stone: config.StoneCost,
		Food:  config.FoodCost,
		Gold:  config.GoldCost,
	}

	// 5. Calcular tiempo restante
	var timeRemaining time.Duration
	if building.UpgradeCompletionTime != nil {
		timeRemaining = time.Until(*building.UpgradeCompletionTime)
		if timeRemaining < 0 {
			timeRemaining = 0
		}
	}

	// 6. Actualizar recursos (agregar reembolso)
	newWood := village.Resources.Wood + refundAmount.Wood
	newStone := village.Resources.Stone + refundAmount.Stone
	newFood := village.Resources.Food + refundAmount.Food
	newGold := village.Resources.Gold + refundAmount.Gold

	err = s.villageRepo.UpdateResources(villageID, newWood, newStone, newFood, newGold)
	if err != nil {
		return nil, fmt.Errorf("error actualizando recursos: %w", err)
	}

	// 7. Cancelar la mejora (volver al nivel anterior)
	err = s.villageRepo.UpdateBuilding(
		villageID,
		buildingType,
		building.Level, // Mantener nivel actual
		false,          // isUpgrading = false
		nil,            // completionTime = nil
	)
	if err != nil {
		return nil, fmt.Errorf("error cancelando mejora: %w", err)
	}

	// 8. Crear resultado
	result := &models.CancelUpgradeResult{
		BuildingType:     buildingType,
		RefundAmount:     refundAmount,
		RefundPercentage: refundPercentage * 100, // Convertir a porcentaje
		OriginalCost:     originalCost,
		CancelledAt:      time.Now(),
		TimeRemaining:    timeRemaining,
		RefundReason:     "Cancelación voluntaria por el jugador",
	}

	s.logger.Info("Mejora cancelada exitosamente con reembolso",
		zap.String("village_id", villageID.String()),
		zap.String("building_type", buildingType),
		zap.Int("refund_wood", refundAmount.Wood),
		zap.Int("refund_stone", refundAmount.Stone),
		zap.Int("refund_food", refundAmount.Food),
		zap.Int("refund_gold", refundAmount.Gold),
		zap.Float64("refund_percentage", refundPercentage*100),
	)

	// Enviar notificación de cancelación de mejora
	s.sendBuildingUpgradeCancelled(villageID, buildingType, refundAmount, refundPercentage*100)

	return result, nil
}

// ===== MÉTODOS DE NOTIFICACIÓN WEBSOCKET =====

// sendBuildingUpgradeStarted envía notificación de inicio de mejora
func (s *ConstructionService) sendBuildingUpgradeStarted(villageID uuid.UUID, buildingType string, targetLevel int, upgradeTime time.Duration, completionTime time.Time, costs models.ResourceCostsLegacy) {
	if s.wsManager == nil {
		return
	}

	// Obtener información del jugador propietario
	village, err := s.villageRepo.GetVillageByID(villageID)
	if err != nil || village == nil {
		s.logger.Error("Error obteniendo aldea para notificación", zap.Error(err))
		return
	}

	message := websocket.WSMessage{
		Type: "building_upgrade_started",
		Data: map[string]interface{}{
			"village_id":       villageID.String(),
			"building_type":    buildingType,
			"target_level":     targetLevel,
			"upgrade_time":     upgradeTime.String(),
			"completion_time":  completionTime.Format(time.RFC3339),
			"costs":            costs,
			"timestamp":        time.Now().Unix(),
		},
		Time: time.Now(),
	}

	s.wsManager.SendToUser(village.Village.PlayerID.String(), "building_upgrade_started", message.Data)

	s.logger.Info("Notificación de inicio de mejora enviada",
		zap.String("village_id", villageID.String()),
		zap.String("building_type", buildingType),
		zap.String("player_id", village.Village.PlayerID.String()),
	)
}

// sendBuildingProgressUpdate envía actualización de progreso de mejora
func (s *ConstructionService) sendBuildingProgressUpdate(villageID uuid.UUID, buildingType string, timeRemaining time.Duration, progressPercent float64) {
	if s.wsManager == nil {
		return
	}

	// Obtener información del jugador propietario
	village, err := s.villageRepo.GetVillageByID(villageID)
	if err != nil || village == nil {
		s.logger.Error("Error obteniendo aldea para notificación de progreso", zap.Error(err))
		return
	}

	message := websocket.WSMessage{
		Type: "building_progress",
		Data: map[string]interface{}{
			"village_id":       villageID.String(),
			"building_type":    buildingType,
			"time_remaining":   timeRemaining.String(),
			"progress_percent": progressPercent,
			"timestamp":        time.Now().Unix(),
		},
		Time: time.Now(),
	}

	s.wsManager.SendToUser(village.Village.PlayerID.String(), "building_progress", message.Data)
}

// sendBuildingUpgradeCompleted envía notificación de finalización de mejora
func (s *ConstructionService) sendBuildingUpgradeCompleted(villageID uuid.UUID, buildingType string, newLevel int) {
	if s.wsManager == nil {
		return
	}

	// Obtener información del jugador propietario
	village, err := s.villageRepo.GetVillageByID(villageID)
	if err != nil || village == nil {
		s.logger.Error("Error obteniendo aldea para notificación de finalización", zap.Error(err))
		return
	}

	// Obtener efectos del edificio mejorado
	config, err := s.buildingConfigRepo.GetBuildingConfig(buildingType, newLevel)
	var effects map[string]interface{}
	if err == nil && config != nil {
		effects = map[string]interface{}{
			"storage_capacity": config.StorageCapacity,
			"production_per_hour": config.ProductionPerHour,
			"build_time_seconds": config.BuildTimeSeconds,
			"construction_speed_modifier": config.ConstructionSpeedModifier,
		}
	}

	message := websocket.WSMessage{
		Type: "building_upgrade_completed",
		Data: map[string]interface{}{
			"village_id":       villageID.String(),
			"building_type":    buildingType,
			"new_level":        newLevel,
			"completion_time":  time.Now().Format(time.RFC3339),
			"effects":          effects,
			"timestamp":        time.Now().Unix(),
		},
		Time: time.Now(),
	}

	s.wsManager.SendToUser(village.Village.PlayerID.String(), "building_upgrade_completed", message.Data)

	s.logger.Info("Notificación de finalización de mejora enviada",
		zap.String("village_id", villageID.String()),
		zap.String("building_type", buildingType),
		zap.Int("new_level", newLevel),
		zap.String("player_id", village.Village.PlayerID.String()),
	)
}

// sendBuildingUpgradeCancelled envía notificación de cancelación de mejora
func (s *ConstructionService) sendBuildingUpgradeCancelled(villageID uuid.UUID, buildingType string, refundAmount models.ResourceCostsLegacy, refundPercentage float64) {
	if s.wsManager == nil {
		return
	}

	// Obtener información del jugador propietario
	village, err := s.villageRepo.GetVillageByID(villageID)
	if err != nil || village == nil {
		s.logger.Error("Error obteniendo aldea para notificación de cancelación", zap.Error(err))
		return
	}

	message := websocket.WSMessage{
		Type: "building_upgrade_cancelled",
		Data: map[string]interface{}{
			"village_id":        villageID.String(),
			"building_type":    buildingType,
			"refund_amount":    refundAmount,
			"refund_percentage": refundPercentage,
			"cancelled_at":     time.Now().Format(time.RFC3339),
			"timestamp":        time.Now().Unix(),
		},
		Time: time.Now(),
	}

	s.wsManager.SendToUser(village.Village.PlayerID.String(), "building_upgrade_cancelled", message.Data)

	s.logger.Info("Notificación de cancelación de mejora enviada",
		zap.String("village_id", villageID.String()),
		zap.String("building_type", buildingType),
		zap.String("player_id", village.Village.PlayerID.String()),
	)
}

// calculateProgressPercent calcula el porcentaje de progreso de una mejora
func (s *ConstructionService) calculateProgressPercent(building *models.Building, timeRemaining time.Duration) float64 {
	if building.UpgradeCompletionTime == nil {
		return 0
	}

	// Calcular tiempo total de la mejora
	totalTime := time.Until(*building.UpgradeCompletionTime) + timeRemaining
	if totalTime <= 0 {
		return 100
	}

	// Calcular porcentaje completado
	elapsedTime := totalTime - timeRemaining
	progressPercent := (float64(elapsedTime) / float64(totalTime)) * 100

	if progressPercent < 0 {
		return 0
	}
	if progressPercent > 100 {
		return 100
	}

	return progressPercent
}

// ===== SISTEMA DE ACTUALIZACIONES PERIÓDICAS =====

// StartProgressUpdates inicia el sistema de actualizaciones periódicas de progreso
func (s *ConstructionService) StartProgressUpdates() {
	go func() {
		ticker := time.NewTicker(30 * time.Second) // Actualizar cada 30 segundos
		defer ticker.Stop()

		s.logger.Info("Sistema de actualizaciones de progreso de mejoras iniciado",
			zap.Duration("interval", 30*time.Second),
		)

		for {
			select {
			case <-ticker.C:
				s.updateAllBuildingProgress()
			}
		}
	}()
}

// updateAllBuildingProgress actualiza el progreso de todas las mejoras activas
func (s *ConstructionService) updateAllBuildingProgress() {
	// Obtener todas las aldeas con mejoras en progreso
	villages, err := s.villageRepo.GetAllVillages()
	if err != nil {
		s.logger.Error("Error obteniendo aldeas para actualización de progreso", zap.Error(err))
		return
	}

	updatedCount := 0
	for _, village := range villages {
		for buildingType, building := range village.Buildings {
			if building.IsUpgrading && building.UpgradeCompletionTime != nil {
				timeRemaining := time.Until(*building.UpgradeCompletionTime)
				
				// Solo enviar actualización si aún queda tiempo
				if timeRemaining > 0 {
					progressPercent := s.calculateProgressPercent(building, timeRemaining)
					s.sendBuildingProgressUpdate(village.Village.ID, buildingType, timeRemaining, progressPercent)
					updatedCount++
				}
			}
		}
	}

	if updatedCount > 0 {
		s.logger.Debug("Actualizaciones de progreso enviadas",
			zap.Int("buildings_updated", updatedCount),
		)
	}
}

// GetVillagesWithActiveUpgrades obtiene aldeas con mejoras activas (método auxiliar)
func (s *ConstructionService) GetVillagesWithActiveUpgrades() ([]*models.VillageWithDetails, error) {
	villages, err := s.villageRepo.GetAllVillages()
	if err != nil {
		return nil, err
	}

	var activeVillages []*models.VillageWithDetails
	for _, village := range villages {
		hasActiveUpgrades := false
		for _, building := range village.Buildings {
			if building.IsUpgrading {
				hasActiveUpgrades = true
				break
			}
		}
		if hasActiveUpgrades {
			activeVillages = append(activeVillages, village)
		}
	}

	return activeVillages, nil
}

// ===== MÉTODOS DE GESTIÓN DE COLA DE CONSTRUCCIÓN =====

// getActiveConstructionCount cuenta cuántos edificios están siendo mejorados en una aldea
func (s *ConstructionService) getActiveConstructionCount(villageID uuid.UUID) (int, error) {
	village, err := s.villageRepo.GetVillageByID(villageID)
	if err != nil {
		return 0, err
	}
	if village == nil {
		return 0, errors.New("aldea no encontrada")
	}

	count := 0
	for _, building := range village.Buildings {
		if building.IsUpgrading {
			count++
		}
	}

	return count, nil
}

// canStartConstruction verifica si se puede iniciar una nueva construcción
func (s *ConstructionService) canStartConstruction(villageID uuid.UUID) (bool, error) {
	activeCount, err := s.getActiveConstructionCount(villageID)
	if err != nil {
		return false, err
	}

	// Verificar si hay slots disponibles
	return activeCount < ActiveConstructionSlots, nil
}

// GetConstructionQueueStatus obtiene el estado actual de la cola de construcción
func (s *ConstructionService) GetConstructionQueueStatus(villageID uuid.UUID) (*ConstructionQueueStatus, error) {
	activeCount, err := s.getActiveConstructionCount(villageID)
	if err != nil {
		return nil, err
	}

	village, err := s.villageRepo.GetVillageByID(villageID)
	if err != nil {
		return nil, err
	}
	if village == nil {
		return nil, errors.New("aldea no encontrada")
	}

	// Obtener edificios en construcción con detalles
	buildingsInConstruction := []BuildingConstructionInfo{}
	for buildingType, building := range village.Buildings {
		if building.IsUpgrading {
			timeRemaining := time.Duration(0)
			if building.UpgradeCompletionTime != nil {
				timeRemaining = time.Until(*building.UpgradeCompletionTime)
				if timeRemaining < 0 {
					timeRemaining = 0
				}
			}

			buildingsInConstruction = append(buildingsInConstruction, BuildingConstructionInfo{
				BuildingType:     buildingType,
				CurrentLevel:     building.Level,
				TargetLevel:      building.Level + 1,
				TimeRemaining:    timeRemaining,
				ProgressPercent:  s.calculateProgressPercent(building, timeRemaining),
				CompletionTime:   building.UpgradeCompletionTime,
			})
		}
	}

	return &ConstructionQueueStatus{
		ActiveSlots:           ActiveConstructionSlots,
		MaxSlots:              MaxConstructionSlots,
		AvailableSlots:       ActiveConstructionSlots - activeCount,
		BuildingsInConstruction: buildingsInConstruction,
		CanStartConstruction:  activeCount < ActiveConstructionSlots,
	}, nil
}

// ConstructionQueueStatus representa el estado de la cola de construcción
type ConstructionQueueStatus struct {
	ActiveSlots            int                      `json:"active_slots"`
	MaxSlots               int                      `json:"max_slots"`
	AvailableSlots         int                      `json:"available_slots"`
	BuildingsInConstruction []BuildingConstructionInfo `json:"buildings_in_construction"`
	CanStartConstruction   bool                     `json:"can_start_construction"`
}

// BuildingConstructionInfo representa información de un edificio en construcción
type BuildingConstructionInfo struct {
	BuildingType     string         `json:"building_type"`
	CurrentLevel     int            `json:"current_level"`
	TargetLevel      int            `json:"target_level"`
	TimeRemaining    time.Duration  `json:"time_remaining"`
	ProgressPercent  float64        `json:"progress_percent"`
	CompletionTime   *time.Time     `json:"completion_time,omitempty"`
}
