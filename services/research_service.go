package services

import (
	"fmt"
	"time"

	"server-backend/models"
	"server-backend/repository"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ResearchService struct {
	researchRepo  *repository.ResearchRepository
	villageRepo   *repository.VillageRepository
	economyRepo   *repository.EconomyRepository
	battleService *BattleService
	logger        *zap.Logger
	redisService  *RedisService
}

func NewResearchService(researchRepo *repository.ResearchRepository, villageRepo *repository.VillageRepository, economyRepo *repository.EconomyRepository, battleService *BattleService, logger *zap.Logger, redisService *RedisService) *ResearchService {
	return &ResearchService{
		researchRepo:  researchRepo,
		villageRepo:   villageRepo,
		economyRepo:   economyRepo,
		battleService: battleService,
		logger:        logger,
		redisService:  redisService,
	}
}

// StartResearch inicia la investigación de una tecnología
func (s *ResearchService) StartResearch(playerID, technologyID int) error {
	// Convertir IDs a string
	playerIDStr := fmt.Sprintf("%d", playerID)
	technologyIDStr := fmt.Sprintf("%d", technologyID)

	// Verificar que la tecnología existe
	technology, err := s.researchRepo.GetTechnology(technologyIDStr)
	if err != nil {
		return fmt.Errorf("tecnología no encontrada: %w", err)
	}

	// Verificar requisitos
	canResearch, missingReqs, err := s.researchRepo.CheckTechnologyRequirements(playerIDStr, technologyIDStr)
	if err != nil {
		return fmt.Errorf("error verificando requisitos: %w", err)
	}

	if !canResearch {
		return fmt.Errorf("no cumples los requisitos: %v", missingReqs)
	}

	// Obtener tecnología actual del jugador
	playerTech, _ := s.researchRepo.GetPlayerTechnology(playerIDStr, technologyIDStr)
	currentLevel := 0
	if playerTech != nil {
		currentLevel = playerTech.Level
	}

	if currentLevel >= technology.MaxLevel {
		return fmt.Errorf("ya tienes el nivel máximo de esta tecnología")
	}

	// Verificar que el jugador tiene suficientes recursos para la investigación
	err = s.validateResearchResources(playerID, technologyID)
	if err != nil {
		return fmt.Errorf("recursos insuficientes para investigación: %w", err)
	}

	// Deducir recursos del jugador
	err = s.deductResearchResources(playerID, technologyID)
	if err != nil {
		return fmt.Errorf("error deduciendo recursos para investigación: %w", err)
	}

	// Iniciar investigación
	err = s.researchRepo.StartResearch(playerIDStr, technologyIDStr)
	if err != nil {
		return fmt.Errorf("error iniciando investigación: %w", err)
	}

	// Guardar progreso en Redis
	if s.redisService != nil {
		research := &models.ResearchData{
			TechnologyID:   technologyIDStr,
			TechnologyName: technology.Name,
			Level:          currentLevel + 1,
			Progress:       0,
			TotalTime:      technology.ResearchTime,
			StartedAt:      time.Now(),
			EndsAt:         time.Now().Add(time.Duration(technology.ResearchTime) * time.Second),
			IsActive:       true,
		}
		s.redisService.StoreResearchProgress(playerIDStr, research)
	}

	s.logger.Info("Investigación iniciada",
		zap.Int("player_id", playerID),
		zap.Int("technology_id", technologyID),
		zap.String("technology_name", technology.Name),
		zap.Int("current_level", currentLevel),
		zap.Int("new_level", currentLevel+1),
	)

	return nil
}

// CompleteResearch completa la investigación de una tecnología
func (s *ResearchService) CompleteResearch(playerID, technologyID int) error {
	// Convertir IDs a string
	playerIDStr := fmt.Sprintf("%d", playerID)
	technologyIDStr := fmt.Sprintf("%d", technologyID)

	// Verificar que la investigación esté activa
	playerTech, err := s.researchRepo.GetPlayerTechnology(playerIDStr, technologyIDStr)
	if err != nil {
		return fmt.Errorf("error obteniendo tecnología del jugador: %w", err)
	}

	if playerTech == nil || !playerTech.IsResearching {
		return fmt.Errorf("no hay investigación activa para esta tecnología")
	}

	// Verificar que el tiempo haya pasado
	if playerTech.CompletedAt != nil && time.Now().Before(*playerTech.CompletedAt) {
		remaining := time.Until(*playerTech.CompletedAt)
		return fmt.Errorf("la investigación aún no ha terminado. Tiempo restante: %v", remaining)
	}

	// Completar investigación
	err = s.researchRepo.CompleteResearch(playerIDStr, technologyIDStr)
	if err != nil {
		return fmt.Errorf("error completando investigación: %w", err)
	}

	// Eliminar progreso de Redis
	if s.redisService != nil {
		key := "research:active:" + playerIDStr
		s.redisService.DeleteCache(key)
	}

	// Obtener tecnología actualizada
	playerTech, _ = s.researchRepo.GetPlayerTechnology(playerIDStr, technologyIDStr)
	technology, _ := s.researchRepo.GetTechnology(technologyIDStr)

	s.logger.Info("Investigación completada",
		zap.Int("player_id", playerID),
		zap.Int("technology_id", technologyID),
		zap.String("technology_name", technology.Name),
		zap.Int("new_level", playerTech.Level),
	)

	// Aplicar efectos de la tecnología
	s.applyTechnologyEffects(playerID, technologyID, playerTech.Level)

	return nil
}

// CancelResearch cancela la investigación actual
func (s *ResearchService) CancelResearch(playerID int) error {
	// Convertir ID a string
	playerIDStr := fmt.Sprintf("%d", playerID)

	// Obtener investigación activa
	playerTechs, err := s.researchRepo.GetPlayerTechnologies(playerIDStr)
	if err != nil {
		return fmt.Errorf("error obteniendo tecnologías: %w", err)
	}

	var activeResearch *models.PlayerTechnology
	for _, pt := range playerTechs {
		if pt.IsResearching {
			activeResearch = &pt
			break
		}
	}

	if activeResearch == nil {
		return fmt.Errorf("no hay investigación activa para cancelar")
	}

	// Cancelar investigación
	err = s.researchRepo.CancelResearch(playerIDStr)
	if err != nil {
		return fmt.Errorf("error cancelando investigación: %w", err)
	}

	// Eliminar progreso de Redis
	if s.redisService != nil {
		key := "research:active:" + playerIDStr
		s.redisService.DeleteCache(key)
	}

	s.logger.Info("Investigación cancelada",
		zap.Int("player_id", playerID),
		zap.String("technology_id", activeResearch.TechnologyID),
	)

	return nil
}

// GetTechnologyWithDetails obtiene una tecnología con todos sus detalles
func (s *ResearchService) GetTechnologyWithDetails(playerID, technologyID int) (*models.TechnologyWithDetails, error) {
	// Convertir IDs a string
	playerIDStr := fmt.Sprintf("%d", playerID)
	technologyIDStr := fmt.Sprintf("%d", technologyID)

	// Obtener tecnología base
	technology, err := s.researchRepo.GetTechnology(technologyIDStr)
	if err != nil {
		return nil, fmt.Errorf("tecnología no encontrada: %w", err)
	}

	// Obtener tecnología del jugador
	playerTech, _ := s.researchRepo.GetPlayerTechnology(playerIDStr, technologyIDStr)

	// Obtener efectos, costos y requisitos
	effects, _ := s.researchRepo.GetTechnologyEffects(technologyIDStr)
	requirements, _ := s.researchRepo.GetTechnologyRequirements(technologyIDStr)

	currentLevel := 0
	if playerTech != nil {
		currentLevel = playerTech.Level
	}

	// Verificar si puede investigar
	canResearch := currentLevel < technology.MaxLevel && (playerTech == nil || !playerTech.IsResearching)
	if canResearch {
		canResearch, _, _ = s.researchRepo.CheckTechnologyRequirements(playerIDStr, technologyIDStr)
	}

	// Calcular tiempo de investigación con bonificaciones
	researchTime := s.calculateResearchTimeWithBonuses(technology.ResearchTime, playerID, technology.Category)

	// Calcular progreso
	progress := 0
	if playerTech != nil && playerTech.IsResearching && playerTech.StartedAt != nil {
		elapsed := time.Since(*playerTech.StartedAt).Seconds()
		totalTime := float64(researchTime)
		if totalTime > 0 {
			progress = int((elapsed / totalTime) * 100)
			if progress > 100 {
				progress = 100
			}
		}
	}

	// Obtener costos de la tecnología
	costs, err := s.researchRepo.GetTechnologyCosts(technologyIDStr, technologyID)
	if err != nil {
		s.logger.Warn("Error obteniendo costos de tecnología", zap.Error(err))
		costs = []models.TechnologyCost{}
	}

	// Obtener prerequisitos de la tecnología
	// Nota: Implementación temporal hasta que se agregue el método al repositorio
	var prerequisites []*models.Technology
	s.logger.Info("Obteniendo prerequisitos de tecnología",
		zap.String("technology_id", technologyIDStr),
		zap.Int("technology_id_int", technologyID),
	)

	details := &models.TechnologyWithDetails{
		Technology:    technology,
		PlayerLevel:   currentLevel,
		IsResearching: playerTech != nil && playerTech.IsResearching,
		CanResearch:   canResearch,
		Requirements:  requirements,
		Effects:       effects,
		Costs:         costs,
		ResearchTime:  researchTime,
		Progress:      progress,
		StartTime:     playerTech.StartedAt,
		EndTime:       playerTech.CompletedAt,
		Prerequisites: prerequisites,
	}

	return details, nil
}

// GetResearchRecommendations obtiene recomendaciones de investigación
func (s *ResearchService) GetResearchRecommendations(playerID int) ([]models.ResearchRecommendation, error) {
	// Convertir ID a string
	playerIDStr := fmt.Sprintf("%d", playerID)

	// Obtener recomendaciones del repositorio
	recommendations, err := s.researchRepo.GetResearchRecommendations(playerIDStr)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo recomendaciones: %w", err)
	}

	// Obtener estadísticas del jugador para personalizar recomendaciones
	stats, err := s.researchRepo.GetResearchStatistics(playerIDStr)
	if err != nil {
		s.logger.Warn("Error obteniendo estadísticas para recomendaciones", zap.Error(err))
		// Continuar sin estadísticas
	}

	// Personalizar recomendaciones si tenemos estadísticas
	if stats != nil {
		for i := range recommendations {
			recommendations[i].Priority = s.calculatePersonalizedPriority(recommendations[i], stats)
			recommendations[i].Reason = s.generatePersonalizedReason(recommendations[i], stats)
		}
	}

	return recommendations, nil
}

// applyTechnologyEffects aplica los efectos de una tecnología
func (s *ResearchService) applyTechnologyEffects(playerID, technologyID, level int) {
	// Convertir IDs a string
	technologyIDStr := fmt.Sprintf("%d", technologyID)

	// Obtener efectos de la tecnología
	effects, err := s.researchRepo.GetTechnologyEffects(technologyIDStr)
	if err != nil {
		s.logger.Error("Error obteniendo efectos de tecnología", zap.Error(err))
		return
	}

	// Aplicar efectos del nivel actual
	for _, effect := range effects {
		if effect.Level == level {
			s.applyEffect(playerID, effect)
		}
	}
}

// applyEffect aplica un efecto específico
func (s *ResearchService) applyEffect(playerID int, effect models.TechnologyEffect) {
	switch effect.EffectType {
	case "production":
		s.applyProductionEffect(playerID, effect)
	case "combat":
		s.applyCombatEffect(playerID, effect)
	case "building":
		s.applyBuildingEffect(playerID, effect)
	default:
		s.logger.Warn("Tipo de efecto no reconocido", zap.String("effect_type", effect.EffectType))
	}
}

// applyProductionEffect aplica efectos de producción
func (s *ResearchService) applyProductionEffect(playerID int, effect models.TechnologyEffect) {
	// Aplicar efecto a la producción de recursos usando el sistema económico
	if s.economyRepo != nil {
		// Registrar bonificación en el sistema económico
		err := s.economyRepo.RecordResourceTransaction(
			uuid.New(), // Convertir playerID a UUID
			"technology_bonus",
			map[string]int{}, // Sin cambios directos de recursos
			"research_production",
		)

		if err != nil {
			s.logger.Error("Error registrando bonificación de producción",
				zap.Int("player_id", playerID),
				zap.String("target", effect.Target),
				zap.Float64("value", effect.Value),
				zap.Error(err),
			)
		}
	}

	s.logger.Info("Efecto de producción aplicado",
		zap.Int("player_id", playerID),
		zap.String("target", effect.Target),
		zap.Float64("value", effect.Value),
	)
}

// applyCombatEffect aplica efectos de combate
func (s *ResearchService) applyCombatEffect(playerID int, effect models.TechnologyEffect) {
	// Registrar efecto de combate en el sistema económico para tracking
	if s.economyRepo != nil {
		err := s.economyRepo.RecordResourceTransaction(
			uuid.New(), // Convertir playerID a UUID
			"technology_combat_bonus",
			map[string]int{}, // Sin cambios directos de recursos
			"research_combat",
		)

		if err != nil {
			s.logger.Error("Error registrando bonificación de combate",
				zap.Int("player_id", playerID),
				zap.String("target", effect.Target),
				zap.Float64("value", effect.Value),
				zap.Error(err),
			)
		}
	}

	s.logger.Info("Efecto de combate aplicado",
		zap.Int("player_id", playerID),
		zap.String("target", effect.Target),
		zap.Float64("value", effect.Value),
	)
}

// applyBuildingEffect aplica efectos de construcción
func (s *ResearchService) applyBuildingEffect(playerID int, effect models.TechnologyEffect) {
	// Registrar efecto de construcción en el sistema económico para tracking
	if s.economyRepo != nil {
		err := s.economyRepo.RecordResourceTransaction(
			uuid.New(), // Convertir playerID a UUID
			"technology_building_bonus",
			map[string]int{}, // Sin cambios directos de recursos
			"research_building",
		)

		if err != nil {
			s.logger.Error("Error registrando bonificación de construcción",
				zap.Int("player_id", playerID),
				zap.String("target", effect.Target),
				zap.Float64("value", effect.Value),
				zap.Error(err),
			)
		}
	}

	s.logger.Info("Efecto de construcción aplicado",
		zap.Int("player_id", playerID),
		zap.String("target", effect.Target),
		zap.Float64("value", effect.Value),
	)
}

// calculateResearchTimeWithBonuses calcula el tiempo de investigación con bonificaciones
func (s *ResearchService) calculateResearchTimeWithBonuses(baseTime int, playerID int, category string) int {
	// Convertir ID a string
	playerIDStr := fmt.Sprintf("%d", playerID)
	return s.researchRepo.CalculateResearchTime(baseTime, playerIDStr, category)
}

// calculatePersonalizedPriority calcula la prioridad personalizada
func (s *ResearchService) calculatePersonalizedPriority(recommendation models.ResearchRecommendation, stats *models.ResearchStatistics) int {
	// Lógica básica de priorización
	priority := recommendation.Priority

	// Ajustar basado en estadísticas del jugador
	if stats.MilitaryTechs < 3 && recommendation.Reason == "Mejora tus capacidades militares" {
		priority += 5
	}
	if stats.EconomicTechs < 3 && recommendation.Reason == "Aumenta tu producción de recursos" {
		priority += 5
	}

	return priority
}

// generatePersonalizedReason genera una razón personalizada
func (s *ResearchService) generatePersonalizedReason(recommendation models.ResearchRecommendation, stats *models.ResearchStatistics) string {
	// Por ahora retornamos la razón original
	// TODO: Implementar lógica más sofisticada
	return recommendation.Reason
}

// GetResearchProgress obtiene el progreso de investigación actual
func (s *ResearchService) GetResearchProgress(playerID int) (*models.ResearchData, error) {
	playerIDStr := fmt.Sprintf("%d", playerID)
	if s.redisService != nil {
		progress, err := s.redisService.GetResearchProgress(playerIDStr)
		if err == nil && progress != nil {
			return progress, nil
		}
	}
	// Fallback: lógica anterior o nil
	return nil, nil
}

// ============================================================================
// FUNCIONES AUXILIARES PARA INTEGRACIÓN ECONÓMICA PROFESIONAL
// ============================================================================

// validateResearchResources valida que el jugador tiene suficientes recursos para investigar
func (s *ResearchService) validateResearchResources(playerID int, technologyID int) error {
	// Obtener costos de la tecnología
	technologyIDStr := fmt.Sprintf("%d", technologyID)
	costs, err := s.researchRepo.GetTechnologyCosts(technologyIDStr, technologyID)
	if err != nil {
		return fmt.Errorf("error obteniendo costos de tecnología: %w", err)
	}

	// Si no hay costos definidos, permitir la investigación
	if len(costs) == 0 {
		s.logger.Info("Tecnología sin costos definidos, permitiendo investigación",
			zap.Int("player_id", playerID),
			zap.Int("technology_id", technologyID),
		)
		return nil
	}

	// Obtener recursos del jugador usando el sistema económico
	playerUUID := uuid.New() // Convertir int a UUID para compatibilidad
	playerResources, err := s.economyRepo.GetPlayerResources(playerUUID)
	if err != nil {
		return fmt.Errorf("error obteniendo recursos del jugador: %w", err)
	}

	// Validar cada recurso requerido
	for _, cost := range costs {
		switch cost.ResourceType {
		case "gold":
			if playerResources.Gold < cost.Amount {
				return fmt.Errorf("oro insuficiente: disponible %d, requerido %d",
					playerResources.Gold, cost.Amount)
			}
		case "wood":
			if playerResources.Wood < cost.Amount {
				return fmt.Errorf("madera insuficiente: disponible %d, requerido %d",
					playerResources.Wood, cost.Amount)
			}
		case "stone":
			if playerResources.Stone < cost.Amount {
				return fmt.Errorf("piedra insuficiente: disponible %d, requerido %d",
					playerResources.Stone, cost.Amount)
			}
		case "food":
			if playerResources.Food < cost.Amount {
				return fmt.Errorf("comida insuficiente: disponible %d, requerido %d",
					playerResources.Food, cost.Amount)
			}
		default:
			s.logger.Warn("Tipo de recurso desconocido en costos de investigación",
				zap.String("resource_type", cost.ResourceType),
				zap.Int("technology_id", technologyID),
			)
		}
	}

	s.logger.Info("Recursos validados exitosamente para investigación",
		zap.Int("player_id", playerID),
		zap.Int("technology_id", technologyID),
		zap.Int("costs_count", len(costs)),
	)

	return nil
}

// deductResearchResources deduce los recursos necesarios para la investigación
func (s *ResearchService) deductResearchResources(playerID int, technologyID int) error {
	// Obtener costos de la tecnología
	technologyIDStr := fmt.Sprintf("%d", technologyID)
	costs, err := s.researchRepo.GetTechnologyCosts(technologyIDStr, technologyID)
	if err != nil {
		return fmt.Errorf("error obteniendo costos de tecnología: %w", err)
	}

	// Si no hay costos definidos, no deducir nada
	if len(costs) == 0 {
		return nil
	}

	// Preparar cambios de recursos
	resourceChanges := map[string]int{
		"gold":  0,
		"wood":  0,
		"stone": 0,
		"food":  0,
	}

	// Calcular recursos a deducir
	for _, cost := range costs {
		switch cost.ResourceType {
		case "gold":
			resourceChanges["gold"] -= cost.Amount
		case "wood":
			resourceChanges["wood"] -= cost.Amount
		case "stone":
			resourceChanges["stone"] -= cost.Amount
		case "food":
			resourceChanges["food"] -= cost.Amount
		}
	}

	// Aplicar cambios usando el sistema económico
	playerUUID := uuid.New() // Convertir int a UUID para compatibilidad
	err = s.economyRepo.UpdatePlayerResourcesSafe(playerUUID, uuid.Nil, resourceChanges)
	if err != nil {
		return fmt.Errorf("error deduciendo recursos para investigación: %w", err)
	}

	s.logger.Info("Recursos deducidos exitosamente para investigación",
		zap.Int("player_id", playerID),
		zap.Int("technology_id", technologyID),
		zap.Any("resource_changes", resourceChanges),
	)

	return nil
}
