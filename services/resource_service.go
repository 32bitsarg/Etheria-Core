package services

import (
	"server-backend/models"
	"server-backend/repository"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ResourceService struct {
	villageRepo        *repository.VillageRepository
	buildingConfigRepo *repository.BuildingConfigRepository
	logger             *zap.Logger
	wsManager          interface{}
	redisService       *RedisService
}

func NewResourceService(villageRepo *repository.VillageRepository, buildingConfigRepo *repository.BuildingConfigRepository, logger *zap.Logger, redisService *RedisService) *ResourceService {
	return &ResourceService{
		villageRepo:        villageRepo,
		buildingConfigRepo: buildingConfigRepo,
		logger:             logger,
		redisService:       redisService,
	}
}

func (s *ResourceService) SetWebSocketManager(wsManager interface{}) {
	s.wsManager = wsManager
}

// CalculateProduction calcula la producción de recursos basada en los edificios actuales
func (s *ResourceService) CalculateProduction(village *models.VillageWithDetails) models.Resources {
	production := models.Resources{
		Wood:  0,
		Stone: 0,
		Food:  0,
		Gold:  0,
	}

	// Calcular producción de cada edificio de recursos
	for buildingType, building := range village.Buildings {
		if building.Level > 0 {
			// Obtener configuración del edificio para su nivel actual
			config, err := s.buildingConfigRepo.GetBuildingConfig(buildingType, building.Level)
			if err != nil {
				s.logger.Error("Error obteniendo configuración de edificio",
					zap.String("building_type", buildingType),
					zap.Int("level", building.Level),
					zap.Error(err),
				)
				continue
			}
			if config == nil {
				continue
			}

			// Sumar producción según el tipo de edificio
			switch buildingType {
			case "wood_cutter":
				production.Wood += config.ProductionPerHour
			case "stone_quarry":
				production.Stone += config.ProductionPerHour
			case "farm":
				production.Food += config.ProductionPerHour
			case "gold_mine":
				production.Gold += config.ProductionPerHour
			}
		}
	}

	return production
}

// CalculateStorageCapacity calcula la capacidad de almacenamiento basada en los edificios
func (s *ResourceService) CalculateStorageCapacity(village *models.VillageWithDetails) models.Resources {
	capacity := models.Resources{
		Wood:  100, // Capacidad base
		Stone: 100,
		Food:  100,
		Gold:  100,
	}

	// Sumar capacidad de almacenes y graneros
	for buildingType, building := range village.Buildings {
		if building.Level > 0 {
			config, err := s.buildingConfigRepo.GetBuildingConfig(buildingType, building.Level)
			if err != nil {
				s.logger.Error("Error obteniendo configuración de edificio",
					zap.String("building_type", buildingType),
					zap.Int("level", building.Level),
					zap.Error(err),
				)
				continue
			}
			if config == nil {
				continue
			}

			switch buildingType {
			case "warehouse":
				// El almacén aumenta capacidad de madera y piedra
				capacity.Wood += config.StorageCapacity
				capacity.Stone += config.StorageCapacity
			case "granary":
				// El granero aumenta capacidad de comida
				capacity.Food += config.StorageCapacity
			}
		}
	}

	return capacity
}

// UpdateResources actualiza los recursos de una aldea
func (s *ResourceService) UpdateResources(villageID uuid.UUID) error {
	village, err := s.villageRepo.GetVillageByID(villageID)
	if err != nil {
		return err
	}
	if village == nil {
		return nil
	}

	// Calcular producción por hora
	production := s.CalculateProduction(village)

	// Calcular tiempo transcurrido desde la última actualización
	now := time.Now()
	lastUpdate := village.Resources.LastUpdated
	if lastUpdate.IsZero() {
		lastUpdate = now
	}

	elapsedHours := now.Sub(lastUpdate).Hours()
	if elapsedHours < 0.01 { // Mínimo 36 segundos
		return nil
	}

	// Calcular recursos generados
	woodGenerated := int(float64(production.Wood) * elapsedHours)
	stoneGenerated := int(float64(production.Stone) * elapsedHours)
	foodGenerated := int(float64(production.Food) * elapsedHours)
	goldGenerated := int(float64(production.Gold) * elapsedHours)

	// Calcular capacidad de almacenamiento
	capacity := s.CalculateStorageCapacity(village)

	// Actualizar recursos respetando límites de capacidad
	newWood := village.Resources.Wood + woodGenerated
	if newWood > capacity.Wood {
		newWood = capacity.Wood
	}

	newStone := village.Resources.Stone + stoneGenerated
	if newStone > capacity.Stone {
		newStone = capacity.Stone
	}

	newFood := village.Resources.Food + foodGenerated
	if newFood > capacity.Food {
		newFood = capacity.Food
	}

	newGold := village.Resources.Gold + goldGenerated
	if newGold > capacity.Gold {
		newGold = capacity.Gold
	}

	// Actualizar en la base de datos (esto también actualiza last_updated automáticamente)
	err = s.villageRepo.UpdateResources(villageID, newWood, newStone, newFood, newGold)
	if err != nil {
		return err
	}

	// Actualizar en Redis
	if s.redisService != nil {
		res := &models.ResourceData{
			Wood:  newWood,
			Stone: newStone,
			Food:  newFood,
			Gold:  newGold,
			Gems:  0,
		}
		s.redisService.StorePlayerResources(villageID.String(), res)
	}

	s.logger.Debug("Recursos actualizados",
		zap.String("village_id", villageID.String()),
		zap.Int("wood_generated", woodGenerated),
		zap.Int("stone_generated", stoneGenerated),
		zap.Int("food_generated", foodGenerated),
		zap.Int("gold_generated", goldGenerated),
		zap.Float64("elapsed_hours", elapsedHours),
	)

	return nil
}

// GetResourceInfo obtiene información detallada de recursos
func (s *ResourceService) GetResourceInfo(villageID uuid.UUID) (*models.ResourceProduction, error) {
	// Primero intentar obtener de Redis
	if s.redisService != nil {
		resources, err := s.redisService.GetPlayerResources(villageID.String())
		if err == nil && resources != nil {
			return &models.ResourceProduction{
				VillageID:  villageID,
				Wood:       resources.Wood,
				Stone:      resources.Stone,
				Food:       resources.Food,
				Gold:       resources.Gold,
				LastUpdate: time.Now(),
			}, nil
		}
	}

	// Si no está en Redis, obtener de la base de datos y cachear
	village, err := s.villageRepo.GetVillageByID(villageID)
	if err != nil {
		return nil, err
	}
	if village == nil {
		return nil, nil
	}

	// Actualizar recursos antes de obtener información
	err = s.UpdateResources(villageID)
	if err != nil {
		return nil, err
	}

	// Guardar en Redis
	if s.redisService != nil {
		res := &models.ResourceData{
			Wood:  village.Resources.Wood,
			Stone: village.Resources.Stone,
			Food:  village.Resources.Food,
			Gold:  village.Resources.Gold,
			Gems:  0,
		}
		s.redisService.StorePlayerResources(villageID.String(), res)
	}

	return &models.ResourceProduction{
		VillageID:  villageID,
		Wood:       village.Resources.Wood,
		Stone:      village.Resources.Stone,
		Food:       village.Resources.Food,
		Gold:       village.Resources.Gold,
		LastUpdate: time.Now(),
	}, nil
}

// Actualizar recursos de todas las aldeas
func (s *ResourceService) UpdateAllVillageResources() error {
	villages, err := s.villageRepo.GetAllVillages()
	if err != nil {
		return err
	}

	for _, village := range villages {
		err := s.UpdateResources(village.Village.ID)
		if err != nil {
			s.logger.Error("Error al actualizar recursos de aldea",
				zap.String("village_id", village.Village.ID.String()),
				zap.Error(err),
			)
		}
	}

	return nil
}

// Iniciar el servicio de generación de recursos
func (s *ResourceService) StartResourceGeneration() {
	ticker := time.NewTicker(5 * time.Minute) // Actualizar cada 5 minutos
	defer ticker.Stop()

	s.logger.Info("Servicio de generación de recursos iniciado")

	for {
		select {
		case <-ticker.C:
			err := s.UpdateAllVillageResources()
			if err != nil {
				s.logger.Error("Error en la generación de recursos", zap.Error(err))
			}
		}
	}
}

// Consumir recursos para construcción o entrenamiento
func (s *ResourceService) ConsumeResources(villageID uuid.UUID, wood, stone, food, gold int) error {
	village, err := s.villageRepo.GetVillageByID(villageID)
	if err != nil {
		return err
	}

	if village == nil {
		return nil
	}

	// Verificar si hay suficientes recursos
	if village.Resources.Wood < wood || village.Resources.Stone < stone || village.Resources.Food < food || village.Resources.Gold < gold {
		return nil // No hay suficientes recursos
	}

	// Consumir recursos
	newWood := village.Resources.Wood - wood
	newStone := village.Resources.Stone - stone
	newFood := village.Resources.Food - food
	newGold := village.Resources.Gold - gold

	// Actualizar en la base de datos
	err = s.villageRepo.UpdateResources(villageID, newWood, newStone, newFood, newGold)
	if err != nil {
		return err
	}

	s.logger.Info("Recursos consumidos",
		zap.String("village_id", villageID.String()),
		zap.Int("wood_consumed", wood),
		zap.Int("stone_consumed", stone),
		zap.Int("food_consumed", food),
		zap.Int("gold_consumed", gold),
	)

	return nil
}

// Verificar si hay suficientes recursos
func (s *ResourceService) HasEnoughResources(villageID uuid.UUID, wood, stone, food, gold int) (bool, error) {
	village, err := s.villageRepo.GetVillageByID(villageID)
	if err != nil {
		return false, err
	}

	if village == nil {
		return false, nil
	}

	hasEnough := village.Resources.Wood >= wood && village.Resources.Stone >= stone && village.Resources.Food >= food && village.Resources.Gold >= gold
	return hasEnough, nil
}

// Obtener información de producción de una aldea
func (s *ResourceService) GetVillageProduction(villageID uuid.UUID) (*models.ResourceProduction, error) {
	village, err := s.villageRepo.GetVillageByID(villageID)
	if err != nil {
		return nil, err
	}

	if village == nil {
		return nil, nil
	}

	production := s.CalculateProduction(village)
	resourceProduction := &models.ResourceProduction{
		VillageID:  villageID,
		Wood:       production.Wood,
		Stone:      production.Stone,
		Food:       production.Food,
		Gold:       production.Gold,
		LastUpdate: village.Resources.LastUpdated,
	}

	return resourceProduction, nil
}

// Obtener información de almacenamiento de una aldea
func (s *ResourceService) GetVillageStorage(villageID uuid.UUID) (*models.ResourceStorage, error) {
	village, err := s.villageRepo.GetVillageByID(villageID)
	if err != nil {
		return nil, err
	}

	if village == nil {
		return nil, nil
	}

	capacity := s.CalculateStorageCapacity(village)
	storage := &models.ResourceStorage{
		VillageID:    villageID,
		WoodStorage:  capacity.Wood,
		StoneStorage: capacity.Stone,
		FoodStorage:  capacity.Food,
		GoldStorage:  capacity.Gold,
	}

	return storage, nil
}

// GetResourceProduction obtiene la producción de recursos
func (s *ResourceService) GetResourceProduction(villageID uuid.UUID) (*models.ResourceProduction, error) {
	village, err := s.villageRepo.GetVillageByID(villageID)
	if err != nil {
		return nil, err
	}
	if village == nil {
		return nil, nil
	}

	production := s.CalculateProduction(village)
	resourceProduction := &models.ResourceProduction{
		VillageID:  villageID,
		Wood:       production.Wood,
		Stone:      production.Stone,
		Food:       production.Food,
		Gold:       production.Gold,
		LastUpdate: village.Resources.LastUpdated,
	}

	return resourceProduction, nil
}

// GetResourceStorage obtiene la capacidad de almacenamiento
func (s *ResourceService) GetResourceStorage(villageID uuid.UUID) (*models.ResourceStorage, error) {
	village, err := s.villageRepo.GetVillageByID(villageID)
	if err != nil {
		return nil, err
	}
	if village == nil {
		return nil, nil
	}

	capacity := s.CalculateStorageCapacity(village)
	storage := &models.ResourceStorage{
		VillageID:    villageID,
		WoodStorage:  capacity.Wood,
		StoneStorage: capacity.Stone,
		FoodStorage:  capacity.Food,
		GoldStorage:  capacity.Gold,
	}

	return storage, nil
}
