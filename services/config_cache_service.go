package services

import (
	"context"
	"fmt"
	"log"
	"time"
)

type ConfigCacheService struct {
	redisService *RedisService
}

type BuildingConfig struct {
	ID           int64                  `json:"id"`
	Name         string                 `json:"name"`
	Type         string                 `json:"type"`
	MaxLevel     int                    `json:"max_level"`
	Costs        map[string]interface{} `json:"costs"`
	Benefits     map[string]interface{} `json:"benefits"`
	Requirements map[string]interface{} `json:"requirements"`
}

type UnitConfig struct {
	ID           int64                  `json:"id"`
	Name         string                 `json:"name"`
	Type         string                 `json:"type"`
	Attack       int                    `json:"attack"`
	Defense      int                    `json:"defense"`
	Health       int                    `json:"health"`
	Speed        int                    `json:"speed"`
	Costs        map[string]interface{} `json:"costs"`
	Requirements map[string]interface{} `json:"requirements"`
}

type TechnologyConfig struct {
	ID           int64                  `json:"id"`
	Name         string                 `json:"name"`
	Type         string                 `json:"type"`
	MaxLevel     int                    `json:"max_level"`
	Costs        map[string]interface{} `json:"costs"`
	Benefits     map[string]interface{} `json:"benefits"`
	Requirements map[string]interface{} `json:"requirements"`
}

func NewConfigCacheService(redisService *RedisService) *ConfigCacheService {
	return &ConfigCacheService{
		redisService: redisService,
	}
}

// CacheBuildingConfigs cachea configuraciones de edificios
func (s *ConfigCacheService) CacheBuildingConfigs(ctx context.Context, configs []*BuildingConfig) error {
	// Cachear lista completa
	configsKey := "config:buildings:all"
	err := s.redisService.SetCache(configsKey, configs, 24*time.Hour)
	if err != nil {
		return fmt.Errorf("error cacheando configuraciones de edificios: %v", err)
	}

	// Cachear configuraciones individuales
	for _, config := range configs {
		configKey := fmt.Sprintf("config:building:%d", config.ID)
		err = s.redisService.SetCache(configKey, config, 24*time.Hour)
		if err != nil {
			log.Printf("Error cacheando configuración de edificio %d: %v", config.ID, err)
		}
	}

	// Cachear por tipo
	configsByType := make(map[string][]*BuildingConfig)
	for _, config := range configs {
		configsByType[config.Type] = append(configsByType[config.Type], config)
	}

	for buildingType, typeConfigs := range configsByType {
		typeKey := fmt.Sprintf("config:buildings:type:%s", buildingType)
		err = s.redisService.SetCache(typeKey, typeConfigs, 24*time.Hour)
		if err != nil {
			log.Printf("Error cacheando edificios por tipo %s: %v", buildingType, err)
		}
	}

	return nil
}

// GetBuildingConfig obtiene una configuración de edificio desde cache
func (s *ConfigCacheService) GetBuildingConfig(ctx context.Context, buildingID int64) (*BuildingConfig, error) {
	configKey := fmt.Sprintf("config:building:%d", buildingID)
	var config BuildingConfig

	err := s.redisService.GetCache(configKey, &config)
	if err != nil {
		return nil, fmt.Errorf("configuración de edificio no encontrada: %v", err)
	}

	return &config, nil
}

// GetBuildingConfigs obtiene todas las configuraciones de edificios
func (s *ConfigCacheService) GetBuildingConfigs(ctx context.Context) ([]*BuildingConfig, error) {
	configsKey := "config:buildings:all"
	var configs []*BuildingConfig

	err := s.redisService.GetCache(configsKey, &configs)
	if err != nil {
		return nil, fmt.Errorf("configuraciones de edificios no encontradas: %v", err)
	}

	return configs, nil
}

// GetBuildingConfigsByType obtiene configuraciones de edificios por tipo
func (s *ConfigCacheService) GetBuildingConfigsByType(ctx context.Context, buildingType string) ([]*BuildingConfig, error) {
	typeKey := fmt.Sprintf("config:buildings:type:%s", buildingType)
	var configs []*BuildingConfig

	err := s.redisService.GetCache(typeKey, &configs)
	if err != nil {
		return nil, fmt.Errorf("configuraciones de edificios por tipo no encontradas: %v", err)
	}

	return configs, nil
}

// CacheUnitConfigs cachea configuraciones de unidades
func (s *ConfigCacheService) CacheUnitConfigs(ctx context.Context, configs []*UnitConfig) error {
	// Cachear lista completa
	configsKey := "config:units:all"
	err := s.redisService.SetCache(configsKey, configs, 24*time.Hour)
	if err != nil {
		return fmt.Errorf("error cacheando configuraciones de unidades: %v", err)
	}

	// Cachear configuraciones individuales
	for _, config := range configs {
		configKey := fmt.Sprintf("config:unit:%d", config.ID)
		err = s.redisService.SetCache(configKey, config, 24*time.Hour)
		if err != nil {
			log.Printf("Error cacheando configuración de unidad %d: %v", config.ID, err)
		}
	}

	// Cachear por tipo
	configsByType := make(map[string][]*UnitConfig)
	for _, config := range configs {
		configsByType[config.Type] = append(configsByType[config.Type], config)
	}

	for unitType, typeConfigs := range configsByType {
		typeKey := fmt.Sprintf("config:units:type:%s", unitType)
		err = s.redisService.SetCache(typeKey, typeConfigs, 24*time.Hour)
		if err != nil {
			log.Printf("Error cacheando unidades por tipo %s: %v", unitType, err)
		}
	}

	return nil
}

// GetUnitConfig obtiene una configuración de unidad desde cache
func (s *ConfigCacheService) GetUnitConfig(ctx context.Context, unitID int64) (*UnitConfig, error) {
	configKey := fmt.Sprintf("config:unit:%d", unitID)
	var config UnitConfig

	err := s.redisService.GetCache(configKey, &config)
	if err != nil {
		return nil, fmt.Errorf("configuración de unidad no encontrada: %v", err)
	}

	return &config, nil
}

// GetUnitConfigs obtiene todas las configuraciones de unidades
func (s *ConfigCacheService) GetUnitConfigs(ctx context.Context) ([]*UnitConfig, error) {
	configsKey := "config:units:all"
	var configs []*UnitConfig

	err := s.redisService.GetCache(configsKey, &configs)
	if err != nil {
		return nil, fmt.Errorf("configuraciones de unidades no encontradas: %v", err)
	}

	return configs, nil
}

// GetUnitConfigsByType obtiene configuraciones de unidades por tipo
func (s *ConfigCacheService) GetUnitConfigsByType(ctx context.Context, unitType string) ([]*UnitConfig, error) {
	typeKey := fmt.Sprintf("config:units:type:%s", unitType)
	var configs []*UnitConfig

	err := s.redisService.GetCache(typeKey, &configs)
	if err != nil {
		return nil, fmt.Errorf("configuraciones de unidades por tipo no encontradas: %v", err)
	}

	return configs, nil
}

// CacheTechnologyConfigs cachea configuraciones de tecnologías
func (s *ConfigCacheService) CacheTechnologyConfigs(ctx context.Context, configs []*TechnologyConfig) error {
	// Cachear lista completa
	configsKey := "config:technologies:all"
	err := s.redisService.SetCache(configsKey, configs, 24*time.Hour)
	if err != nil {
		return fmt.Errorf("error cacheando configuraciones de tecnologías: %v", err)
	}

	// Cachear configuraciones individuales
	for _, config := range configs {
		configKey := fmt.Sprintf("config:technology:%d", config.ID)
		err = s.redisService.SetCache(configKey, config, 24*time.Hour)
		if err != nil {
			log.Printf("Error cacheando configuración de tecnología %d: %v", config.ID, err)
		}
	}

	// Cachear por tipo
	configsByType := make(map[string][]*TechnologyConfig)
	for _, config := range configs {
		configsByType[config.Type] = append(configsByType[config.Type], config)
	}

	for techType, typeConfigs := range configsByType {
		typeKey := fmt.Sprintf("config:technologies:type:%s", techType)
		err = s.redisService.SetCache(typeKey, typeConfigs, 24*time.Hour)
		if err != nil {
			log.Printf("Error cacheando tecnologías por tipo %s: %v", techType, err)
		}
	}

	return nil
}

// GetTechnologyConfig obtiene una configuración de tecnología desde cache
func (s *ConfigCacheService) GetTechnologyConfig(ctx context.Context, techID int64) (*TechnologyConfig, error) {
	configKey := fmt.Sprintf("config:technology:%d", techID)
	var config TechnologyConfig

	err := s.redisService.GetCache(configKey, &config)
	if err != nil {
		return nil, fmt.Errorf("configuración de tecnología no encontrada: %v", err)
	}

	return &config, nil
}

// GetTechnologyConfigs obtiene todas las configuraciones de tecnologías
func (s *ConfigCacheService) GetTechnologyConfigs(ctx context.Context) ([]*TechnologyConfig, error) {
	configsKey := "config:technologies:all"
	var configs []*TechnologyConfig

	err := s.redisService.GetCache(configsKey, &configs)
	if err != nil {
		return nil, fmt.Errorf("configuraciones de tecnologías no encontradas: %v", err)
	}

	return configs, nil
}

// GetTechnologyConfigsByType obtiene configuraciones de tecnologías por tipo
func (s *ConfigCacheService) GetTechnologyConfigsByType(ctx context.Context, techType string) ([]*TechnologyConfig, error) {
	typeKey := fmt.Sprintf("config:technologies:type:%s", techType)
	var configs []*TechnologyConfig

	err := s.redisService.GetCache(typeKey, &configs)
	if err != nil {
		return nil, fmt.Errorf("configuraciones de tecnologías por tipo no encontradas: %v", err)
	}

	return configs, nil
}

// CacheGameSettings cachea configuraciones generales del juego
func (s *ConfigCacheService) CacheGameSettings(ctx context.Context, settings map[string]interface{}) error {
	settingsKey := "config:game:settings"
	err := s.redisService.SetCache(settingsKey, settings, 24*time.Hour)
	if err != nil {
		return fmt.Errorf("error cacheando configuraciones del juego: %v", err)
	}

	return nil
}

// GetGameSettings obtiene configuraciones generales del juego
func (s *ConfigCacheService) GetGameSettings(ctx context.Context) (map[string]interface{}, error) {
	settingsKey := "config:game:settings"
	var settings map[string]interface{}

	err := s.redisService.GetCache(settingsKey, &settings)
	if err != nil {
		return nil, fmt.Errorf("configuraciones del juego no encontradas: %v", err)
	}

	return settings, nil
}

// CacheResourceConfigs cachea configuraciones de recursos
func (s *ConfigCacheService) CacheResourceConfigs(ctx context.Context, configs map[string]interface{}) error {
	configsKey := "config:resources"
	err := s.redisService.SetCache(configsKey, configs, 24*time.Hour)
	if err != nil {
		return fmt.Errorf("error cacheando configuraciones de recursos: %v", err)
	}

	return nil
}

// GetResourceConfigs obtiene configuraciones de recursos
func (s *ConfigCacheService) GetResourceConfigs(ctx context.Context) (map[string]interface{}, error) {
	configsKey := "config:resources"
	var configs map[string]interface{}

	err := s.redisService.GetCache(configsKey, &configs)
	if err != nil {
		return nil, fmt.Errorf("configuraciones de recursos no encontradas: %v", err)
	}

	return configs, nil
}

// InvalidateConfigCache invalida todo el cache de configuraciones
func (s *ConfigCacheService) InvalidateConfigCache(ctx context.Context) error {
	// Obtener todas las claves de configuración
	keys, err := s.redisService.GetKeys(ctx, "config:*")
	if err != nil {
		return fmt.Errorf("error obteniendo claves de configuración: %v", err)
	}

	// Eliminar cada clave
	for _, key := range keys {
		err = s.redisService.DeleteCache(key)
		if err != nil {
			log.Printf("Error eliminando cache %s: %v", key, err)
		}
	}

	return nil
}

// RefreshConfigCache refresca todo el cache de configuraciones
func (s *ConfigCacheService) RefreshConfigCache(ctx context.Context) error {
	// Invalidar cache existente
	err := s.InvalidateConfigCache(ctx)
	if err != nil {
		return fmt.Errorf("error invalidando cache: %v", err)
	}

	// Aquí se cargarían las configuraciones desde la base de datos
	// y se cachearían nuevamente
	log.Printf("Cache de configuraciones refrescado")

	return nil
}

// GetConfigStats obtiene estadísticas del cache de configuraciones
func (s *ConfigCacheService) GetConfigStats(ctx context.Context) (map[string]interface{}, error) {
	// Obtener todas las claves de configuración
	keys, err := s.redisService.GetKeys(ctx, "config:*")
	if err != nil {
		return nil, fmt.Errorf("error obteniendo claves de configuración: %v", err)
	}

	stats := map[string]interface{}{
		"total_configs": len(keys),
		"timestamp":     time.Now(),
		"configs_by_type": map[string]int{
			"buildings":    0,
			"units":        0,
			"technologies": 0,
			"resources":    0,
			"game":         0,
		},
	}

	// Contar por tipo
	for _, key := range keys {
		if len(key) > 7 { // "config:" tiene 7 caracteres
			configType := key[7:]
			if len(configType) > 0 {
				// Extraer tipo principal
				if len(configType) > 9 && configType[:9] == "buildings" {
					stats["configs_by_type"].(map[string]int)["buildings"]++
				} else if len(configType) > 5 && configType[:5] == "units" {
					stats["configs_by_type"].(map[string]int)["units"]++
				} else if len(configType) > 12 && configType[:12] == "technologies" {
					stats["configs_by_type"].(map[string]int)["technologies"]++
				} else if len(configType) > 9 && configType[:9] == "resources" {
					stats["configs_by_type"].(map[string]int)["resources"]++
				} else if len(configType) > 4 && configType[:4] == "game" {
					stats["configs_by_type"].(map[string]int)["game"]++
				}
			}
		}
	}

	return stats, nil
}
