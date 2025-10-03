package services

import (
	"fmt"
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
	metrics            *models.ResourceMetrics
}

func NewResourceService(villageRepo *repository.VillageRepository, buildingConfigRepo *repository.BuildingConfigRepository, logger *zap.Logger, redisService *RedisService) *ResourceService {
	return &ResourceService{
		villageRepo:        villageRepo,
		buildingConfigRepo: buildingConfigRepo,
		logger:             logger,
		redisService:       redisService,
		metrics: &models.ResourceMetrics{
			WebSocketNotifications: 0,
			PollingRequests:        0,
			UpdateErrors:          0,
			LastUpdate:            time.Now(),
			TotalUpdates:         0,
			SuccessRate:          0.0,
		},
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
		s.logger.Error("Error obteniendo aldea para actualizar recursos",
			zap.String("village_id", villageID.String()),
			zap.Error(err),
		)
		return err
	}
	if village == nil {
		s.logger.Warn("Aldea no encontrada para actualizar recursos",
			zap.String("village_id", villageID.String()),
		)
		return nil
	}

	// Calcular producción por hora
	production := s.CalculateProduction(village)

	// Log de diagnóstico de producción
	s.logger.Info("Calculando producción de recursos",
		zap.String("village_id", villageID.String()),
		zap.String("village_name", village.Village.Name),
		zap.Int("wood_production", production.Wood),
		zap.Int("stone_production", production.Stone),
		zap.Int("food_production", production.Food),
		zap.Int("gold_production", production.Gold),
		zap.Int("buildings_count", len(village.Buildings)),
	)

	// Calcular tiempo transcurrido desde la última actualización
	now := time.Now()
	lastUpdate := village.Resources.LastUpdated
	if lastUpdate.IsZero() {
		lastUpdate = now
		s.logger.Info("Primera actualización de recursos - inicializando last_updated",
			zap.String("village_id", villageID.String()),
		)
	}

	elapsedHours := now.Sub(lastUpdate).Hours()

	// Log de diagnóstico de tiempo
	s.logger.Info("Verificando tiempo transcurrido",
		zap.String("village_id", villageID.String()),
		zap.Time("last_update", lastUpdate),
		zap.Time("now", now),
		zap.Float64("elapsed_hours", elapsedHours),
		zap.Float64("min_required", 0.1),
	)

	if elapsedHours < 0.1 { // Mínimo 6 minutos (más realista)
		s.logger.Info("Recursos no actualizados - tiempo insuficiente",
			zap.String("village_id", villageID.String()),
			zap.Float64("elapsed_hours", elapsedHours),
			zap.Float64("min_required", 0.1),
			zap.String("reason", "Esperando más tiempo para generar recursos"),
		)
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

	// ✅ NUEVO: Notificación WebSocket
	if s.wsManager != nil {
		// Obtener información del jugador para notificación específica
		village, err := s.villageRepo.GetVillageByID(villageID)
		if err == nil && village != nil {
			resourceUpdate := models.ResourceUpdate{
				VillageID:      villageID,
				Wood:           newWood,
				Stone:          newStone,
				Food:           newFood,
				Gold:           newGold,
				WoodGenerated:  woodGenerated,
				StoneGenerated: stoneGenerated,
				FoodGenerated:  foodGenerated,
				GoldGenerated:  goldGenerated,
				Capacity:       capacity,
				LastUpdate:     time.Now(),
				ElapsedHours:   elapsedHours,
			}
			
			// Notificar al jugador específico
			if wsManager, ok := s.wsManager.(interface {
				SendResourceUpdateToUser(userID string, villageID string, resources models.ResourceUpdate) error
			}); ok {
				err := wsManager.SendResourceUpdateToUser(village.Village.PlayerID.String(), villageID.String(), resourceUpdate)
				if err != nil {
					s.logger.Warn("Error enviando notificación WebSocket de recursos",
						zap.String("village_id", villageID.String()),
						zap.String("player_id", village.Village.PlayerID.String()),
						zap.Error(err),
					)
				} else {
					s.metrics.WebSocketNotifications++
					s.logger.Debug("Notificación WebSocket de recursos enviada",
						zap.String("village_id", villageID.String()),
						zap.String("player_id", village.Village.PlayerID.String()),
					)
				}
			}
		}
	}

	// Actualizar métricas
	s.metrics.TotalUpdates++
	s.metrics.LastUpdate = time.Now()
	s.metrics.SuccessRate = float64(s.metrics.TotalUpdates-s.metrics.UpdateErrors) / float64(s.metrics.TotalUpdates) * 100

	s.logger.Info("Recursos actualizados",
		zap.String("village_id", villageID.String()),
		zap.Int("wood_generated", woodGenerated),
		zap.Int("stone_generated", stoneGenerated),
		zap.Int("food_generated", foodGenerated),
		zap.Int("gold_generated", goldGenerated),
		zap.Float64("elapsed_hours", elapsedHours),
		zap.Int("wood_total", newWood),
		zap.Int("stone_total", newStone),
		zap.Int("food_total", newFood),
		zap.Int("gold_total", newGold),
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
	s.logger.Info("Iniciando actualización de recursos para todas las aldeas")

	villages, err := s.villageRepo.GetAllVillages()
	if err != nil {
		s.logger.Error("Error obteniendo aldeas para actualizar recursos", zap.Error(err))
		return err
	}

	s.logger.Info("Aldeas encontradas para actualizar recursos",
		zap.Int("total_villages", len(villages)),
	)

	updatedCount := 0
	errorCount := 0

	for _, village := range villages {
		s.logger.Debug("Procesando aldea para actualización de recursos",
			zap.String("village_id", village.Village.ID.String()),
			zap.String("village_name", village.Village.Name),
			zap.String("player_id", village.Village.PlayerID.String()),
		)

		err := s.UpdateResources(village.Village.ID)
		if err != nil {
			s.logger.Error("Error al actualizar recursos de aldea",
				zap.String("village_id", village.Village.ID.String()),
				zap.String("village_name", village.Village.Name),
				zap.Error(err),
			)
			errorCount++
		} else {
			updatedCount++
		}
	}

	s.logger.Info("Actualización de recursos completada",
		zap.Int("total_villages", len(villages)),
		zap.Int("updated_successfully", updatedCount),
		zap.Int("errors", errorCount),
	)

	return nil
}

// Iniciar el servicio de generación de recursos
func (s *ResourceService) StartResourceGeneration() {
	ticker := time.NewTicker(5 * time.Minute) // Actualizar cada 5 minutos
	defer ticker.Stop()

	s.logger.Info("Servicio de generación de recursos iniciado",
		zap.Duration("interval", 5*time.Minute),
	)

	for {
		select {
		case <-ticker.C:
			s.logger.Info("Ejecutando ciclo de generación de recursos")
			err := s.UpdateAllVillageResources()
			if err != nil {
				s.logger.Error("Error en la generación de recursos", zap.Error(err))
			} else {
				s.logger.Info("Ciclo de generación de recursos completado exitosamente")
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

// GetResourceInfoWithFallback obtiene información de recursos con fallback inteligente
func (s *ResourceService) GetResourceInfoWithFallback(villageID uuid.UUID) (*models.ResourceProduction, error) {
	// Incrementar contador de polling
	s.metrics.PollingRequests++
	
	// Intentar WebSocket primero si está disponible
	if s.wsManager != nil {
		if wsManager, ok := s.wsManager.(interface {
			IsUserOnline(userID string) bool
		}); ok {
			village, err := s.villageRepo.GetVillageByID(villageID)
			if err == nil && village != nil {
				if wsManager.IsUserOnline(village.Village.PlayerID.String()) {
					s.logger.Debug("Usuario conectado, usando WebSocket para recursos",
						zap.String("village_id", villageID.String()),
						zap.String("player_id", village.Village.PlayerID.String()),
					)
				} else {
					s.logger.Debug("Usuario desconectado, usando polling para recursos",
						zap.String("village_id", villageID.String()),
						zap.String("player_id", village.Village.PlayerID.String()),
					)
				}
			}
		}
	}
	
	// Usar método estándar (que incluye actualización automática)
	return s.GetResourceInfo(villageID)
}

// UpdateResourcesWithRetry actualiza recursos con reintentos automáticos
func (s *ResourceService) UpdateResourcesWithRetry(villageID uuid.UUID) error {
	maxRetries := 3
	var lastErr error
	
	for i := 0; i < maxRetries; i++ {
		err := s.UpdateResources(villageID)
		if err == nil {
			return nil
		}
		
		lastErr = err
		s.metrics.UpdateErrors++
		
		if i < maxRetries-1 {
			backoffDuration := time.Duration(i+1) * time.Second
			s.logger.Warn("Error actualizando recursos, reintentando",
				zap.String("village_id", villageID.String()),
				zap.Int("attempt", i+1),
				zap.Duration("backoff", backoffDuration),
				zap.Error(err),
			)
			time.Sleep(backoffDuration)
		}
	}
	
	s.logger.Error("Error actualizando recursos después de múltiples intentos",
		zap.String("village_id", villageID.String()),
		zap.Int("max_retries", maxRetries),
		zap.Error(lastErr),
	)
	
	return fmt.Errorf("failed to update resources after %d retries: %v", maxRetries, lastErr)
}

// GetResourceMetrics obtiene métricas del sistema de recursos
func (s *ResourceService) GetResourceMetrics() *models.ResourceMetrics {
	return s.metrics
}

// ResetResourceMetrics reinicia las métricas del sistema
func (s *ResourceService) ResetResourceMetrics() {
	s.metrics = &models.ResourceMetrics{
		WebSocketNotifications: 0,
		PollingRequests:        0,
		UpdateErrors:          0,
		LastUpdate:            time.Now(),
		TotalUpdates:         0,
		SuccessRate:          0.0,
	}
	s.logger.Info("Métricas de recursos reiniciadas")
}

// UpdateAllVillageResourcesWithNotifications actualiza todas las aldeas y envía notificaciones masivas
func (s *ResourceService) UpdateAllVillageResourcesWithNotifications() error {
	s.logger.Info("Iniciando actualización masiva de recursos con notificaciones")

	villages, err := s.villageRepo.GetAllVillages()
	if err != nil {
		s.logger.Error("Error obteniendo aldeas para actualización masiva", zap.Error(err))
		return err
	}

	s.logger.Info("Aldeas encontradas para actualización masiva",
		zap.Int("total_villages", len(villages)),
	)

	updatedCount := 0
	errorCount := 0
	resourceUpdates := make(map[string]models.ResourceUpdate)

	for _, village := range villages {
		s.logger.Debug("Procesando aldea para actualización masiva",
			zap.String("village_id", village.Village.ID.String()),
			zap.String("village_name", village.Village.Name),
			zap.String("player_id", village.Village.PlayerID.String()),
		)

		err := s.UpdateResources(village.Village.ID)
		if err != nil {
			s.logger.Error("Error al actualizar recursos de aldea",
				zap.String("village_id", village.Village.ID.String()),
				zap.String("village_name", village.Village.Name),
				zap.Error(err),
			)
			errorCount++
		} else {
			updatedCount++
			
			// Preparar notificación masiva
			if s.wsManager != nil {
				villageInfo, err := s.villageRepo.GetVillageByID(village.Village.ID)
				if err == nil && villageInfo != nil {
					resourceUpdate := models.ResourceUpdate{
						VillageID:      village.Village.ID,
						Wood:           villageInfo.Resources.Wood,
						Stone:          villageInfo.Resources.Stone,
						Food:           villageInfo.Resources.Food,
						Gold:           villageInfo.Resources.Gold,
						LastUpdate:     time.Now(),
					}
					resourceUpdates[village.Village.ID.String()] = resourceUpdate
				}
			}
		}
	}

	// Enviar notificaciones masivas si hay WebSocket Manager
	if s.wsManager != nil && len(resourceUpdates) > 0 {
		if wsManager, ok := s.wsManager.(interface {
			SendResourceUpdateToAllVillages(resources map[string]models.ResourceUpdate) error
		}); ok {
			err := wsManager.SendResourceUpdateToAllVillages(resourceUpdates)
			if err != nil {
				s.logger.Warn("Error enviando notificaciones masivas de recursos", zap.Error(err))
			} else {
				s.logger.Info("Notificaciones masivas de recursos enviadas",
					zap.Int("villages_notified", len(resourceUpdates)),
				)
			}
		}
	}

	s.logger.Info("Actualización masiva de recursos completada",
		zap.Int("total_villages", len(villages)),
		zap.Int("updated_successfully", updatedCount),
		zap.Int("errors", errorCount),
		zap.Int("notifications_sent", len(resourceUpdates)),
	)

	return nil
}
