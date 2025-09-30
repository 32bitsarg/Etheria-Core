package services

import (
	"fmt"
	"time"

	"server-backend/models"
	"server-backend/repository"

	"go.uber.org/zap"
)

type ResearchService struct {
	researchRepo *repository.ResearchRepository
	villageRepo  *repository.VillageRepository
	logger       *zap.Logger
	redisService *RedisService
}

func NewResearchService(researchRepo *repository.ResearchRepository, villageRepo *repository.VillageRepository, logger *zap.Logger, redisService *RedisService) *ResearchService {
	return &ResearchService{
		researchRepo: researchRepo,
		villageRepo:  villageRepo,
		logger:       logger,
		redisService: redisService,
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

	// TODO: Implementar verificación de recursos cuando se complete el sistema de aldeas
	// Por ahora, asumimos que el jugador tiene suficientes recursos

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

	// TODO: Obtener costos cuando se implemente el método en el repositorio
	var costs []models.TechnologyCost

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
		Prerequisites: []*models.Technology{}, // TODO: Implementar
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
	// TODO: Implementar cuando se complete el sistema de aldeas
	s.logger.Info("Aplicando efecto de producción",
		zap.Int("player_id", playerID),
		zap.String("target", effect.Target),
		zap.Float64("value", effect.Value),
	)
}

// applyCombatEffect aplica efectos de combate
func (s *ResearchService) applyCombatEffect(playerID int, effect models.TechnologyEffect) {
	// TODO: Implementar cuando se complete el sistema de batallas
	s.logger.Info("Aplicando efecto de combate",
		zap.Int("player_id", playerID),
		zap.String("target", effect.Target),
		zap.Float64("value", effect.Value),
	)
}

// applyBuildingEffect aplica efectos de construcción
func (s *ResearchService) applyBuildingEffect(playerID int, effect models.TechnologyEffect) {
	// TODO: Implementar cuando se complete el sistema de aldeas
	s.logger.Info("Aplicando efecto de construcción",
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
