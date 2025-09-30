package services

import (
	"encoding/json"
	"fmt"
	"server-backend/repository"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// GameMechanicsService maneja la lógica avanzada del juego usando las funciones de la BD
type GameMechanicsService struct {
	villageRepo *repository.VillageRepository
	playerRepo  *repository.PlayerRepository
	logger      *zap.Logger
}

// ResourceProductionResult representa el resultado del cálculo de producción de recursos
type ResourceProductionResult struct {
	WoodProduction  int     `json:"wood_production"`
	StoneProduction int     `json:"stone_production"`
	FoodProduction  int     `json:"food_production"`
	GoldProduction  int     `json:"gold_production"`
	TechBonus       float64 `json:"tech_bonus"`
	AllianceBonus   float64 `json:"alliance_bonus"`
	EventBonus      float64 `json:"event_bonus"`
	WorldBonus      float64 `json:"world_bonus"`
}

// GameBattleResult representa el resultado de una batalla del sistema de mecánicas
type GameBattleResult struct {
	AttackerVictory    bool            `json:"attacker_victory"`
	AttackerLosses     json.RawMessage `json:"attacker_losses"`
	DefenderLosses     json.RawMessage `json:"defender_losses"`
	BattleDuration     int             `json:"battle_duration"`
	ExperienceGained   int             `json:"experience_gained"`
	ResourcesPlundered json.RawMessage `json:"resources_plundered"`
}

// TradeAnalysisResult representa el análisis de comercio
type TradeAnalysisResult struct {
	ResourceType      string `json:"resource_type"`
	CurrentPrice      int    `json:"current_price"`
	PriceTrend        string `json:"price_trend"`
	SupplyLevel       string `json:"supply_level"`
	DemandLevel       string `json:"demand_level"`
	RecommendedAction string `json:"recommended_action"`
}

// AllianceBenefitsResult representa los beneficios de una alianza
type AllianceBenefitsResult struct {
	MemberID          uuid.UUID       `json:"member_id"`
	ResourceBonus     json.RawMessage `json:"resource_bonus"`
	MilitaryBonus     json.RawMessage `json:"military_bonus"`
	ConstructionBonus json.RawMessage `json:"construction_bonus"`
	TotalBenefits     json.RawMessage `json:"total_benefits"`
}

// PlayerScoreResult representa la puntuación del jugador
type PlayerScoreResult struct {
	TotalScore       int64 `json:"total_score"`
	LevelScore       int64 `json:"level_score"`
	BuildingScore    int64 `json:"building_score"`
	MilitaryScore    int64 `json:"military_score"`
	AchievementScore int64 `json:"achievement_score"`
	ActivityScore    int64 `json:"activity_score"`
	RankPosition     int   `json:"rank_position"`
}

// DailyRewardResult representa las recompensas diarias
type DailyRewardResult struct {
	RewardType        string `json:"reward_type"`
	RewardValue       int    `json:"reward_value"`
	RewardDescription string `json:"reward_description"`
	ConsecutiveDays   int    `json:"consecutive_days"`
	TotalRewardsToday int    `json:"total_rewards_today"`
}

// CleanupResult representa el resultado de la limpieza de datos
type CleanupResult struct {
	TableName      string `json:"table_name"`
	RecordsDeleted int    `json:"records_deleted"`
	CleanupType    string `json:"cleanup_type"`
}

func NewGameMechanicsService(villageRepo *repository.VillageRepository, playerRepo *repository.PlayerRepository, logger *zap.Logger) *GameMechanicsService {
	return &GameMechanicsService{
		villageRepo: villageRepo,
		playerRepo:  playerRepo,
		logger:      logger,
	}
}

// CalculateResourceProduction calcula la producción de recursos usando la función avanzada
func (s *GameMechanicsService) CalculateResourceProduction(villageID uuid.UUID) (*ResourceProductionResult, error) {
	rows, err := s.villageRepo.CalculateResourceProductionAdvanced(villageID)
	if err != nil {
		return nil, fmt.Errorf("error calculando producción de recursos: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, fmt.Errorf("no se pudo calcular la producción de recursos")
	}

	var result ResourceProductionResult
	err = rows.Scan(
		&result.WoodProduction,
		&result.StoneProduction,
		&result.FoodProduction,
		&result.GoldProduction,
	)
	if err != nil {
		return nil, fmt.Errorf("error escaneando resultado: %w", err)
	}

	// Calcular bonificaciones adicionales (estos valores podrían venir de la función)
	result.TechBonus = 0.1      // Ejemplo: 10% de bonificación por tecnologías
	result.AllianceBonus = 0.05 // Ejemplo: 5% de bonificación por alianza
	result.EventBonus = 0.0     // Ejemplo: sin bonificación por eventos
	result.WorldBonus = 0.0     // Ejemplo: sin bonificación por mundo

	s.logger.Info("Producción de recursos calculada",
		zap.String("village_id", villageID.String()),
		zap.Int("wood", result.WoodProduction),
		zap.Int("stone", result.StoneProduction),
		zap.Int("food", result.FoodProduction),
		zap.Int("gold", result.GoldProduction),
	)

	return &result, nil
}

// CalculateBattleOutcome calcula el resultado de una batalla
func (s *GameMechanicsService) CalculateBattleOutcome(
	attackerUnits json.RawMessage,
	defenderUnits json.RawMessage,
	attackerHeroes json.RawMessage,
	defenderHeroes json.RawMessage,
	terrain string,
	weather string,
	attackerTechnologies json.RawMessage,
	defenderTechnologies json.RawMessage,
) (*GameBattleResult, error) {
	rows, err := s.villageRepo.CalculateBattleOutcome(
		attackerUnits, defenderUnits, attackerHeroes, defenderHeroes,
		terrain, weather, attackerTechnologies, defenderTechnologies,
	)
	if err != nil {
		return nil, fmt.Errorf("error calculando resultado de batalla: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, fmt.Errorf("no se pudo calcular el resultado de la batalla")
	}

	var result GameBattleResult
	err = rows.Scan(
		&result.AttackerVictory,
		&result.AttackerLosses,
		&result.DefenderLosses,
		&result.BattleDuration,
		&result.ExperienceGained,
		&result.ResourcesPlundered,
	)
	if err != nil {
		return nil, fmt.Errorf("error escaneando resultado: %w", err)
	}

	s.logger.Info("Resultado de batalla calculado",
		zap.Bool("attacker_victory", result.AttackerVictory),
		zap.Int("battle_duration", result.BattleDuration),
		zap.Int("experience_gained", result.ExperienceGained),
	)

	return &result, nil
}

// CalculateTradeRates calcula las tasas de intercambio
func (s *GameMechanicsService) CalculateTradeRates(resourceType string, worldID *uuid.UUID) (*TradeAnalysisResult, error) {
	rows, err := s.villageRepo.CalculateTradeRates(resourceType, worldID)
	if err != nil {
		return nil, fmt.Errorf("error calculando tasas de intercambio: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, fmt.Errorf("no se pudo calcular las tasas de intercambio")
	}

	var result TradeAnalysisResult
	err = rows.Scan(
		&result.ResourceType,
		&result.CurrentPrice,
		&result.PriceTrend,
		&result.SupplyLevel,
		&result.DemandLevel,
		&result.RecommendedAction,
	)
	if err != nil {
		return nil, fmt.Errorf("error escaneando resultado: %w", err)
	}

	s.logger.Info("Tasas de intercambio calculadas",
		zap.String("resource_type", result.ResourceType),
		zap.Int("current_price", result.CurrentPrice),
		zap.String("price_trend", result.PriceTrend),
		zap.String("recommended_action", result.RecommendedAction),
	)

	return &result, nil
}

// ProcessAllianceBenefits procesa los beneficios de una alianza
func (s *GameMechanicsService) ProcessAllianceBenefits(allianceID uuid.UUID) ([]AllianceBenefitsResult, error) {
	rows, err := s.villageRepo.ProcessAllianceBenefits(allianceID)
	if err != nil {
		return nil, fmt.Errorf("error procesando beneficios de alianza: %w", err)
	}
	defer rows.Close()

	var results []AllianceBenefitsResult
	for rows.Next() {
		var result AllianceBenefitsResult
		err := rows.Scan(
			&result.MemberID,
			&result.ResourceBonus,
			&result.MilitaryBonus,
			&result.ConstructionBonus,
			&result.TotalBenefits,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando resultado: %w", err)
		}
		results = append(results, result)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterando resultados: %w", err)
	}

	s.logger.Info("Beneficios de alianza procesados",
		zap.String("alliance_id", allianceID.String()),
		zap.Int("members_processed", len(results)),
	)

	return results, nil
}

// CalculatePlayerScore calcula la puntuación del jugador
func (s *GameMechanicsService) CalculatePlayerScore(playerID uuid.UUID) (*PlayerScoreResult, error) {
	rows, err := s.villageRepo.CalculatePlayerScore(playerID)
	if err != nil {
		return nil, fmt.Errorf("error calculando puntuación del jugador: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, fmt.Errorf("no se pudo calcular la puntuación del jugador")
	}

	var result PlayerScoreResult
	err = rows.Scan(
		&result.TotalScore,
		&result.LevelScore,
		&result.BuildingScore,
		&result.MilitaryScore,
		&result.AchievementScore,
		&result.ActivityScore,
		&result.RankPosition,
	)
	if err != nil {
		return nil, fmt.Errorf("error escaneando resultado: %w", err)
	}

	s.logger.Info("Puntuación del jugador calculada",
		zap.String("player_id", playerID.String()),
		zap.Int64("total_score", result.TotalScore),
		zap.Int("rank_position", result.RankPosition),
	)

	return &result, nil
}

// GenerateDailyRewards genera recompensas diarias
func (s *GameMechanicsService) GenerateDailyRewards(playerID uuid.UUID) (*DailyRewardResult, error) {
	rows, err := s.villageRepo.GenerateDailyRewards(playerID)
	if err != nil {
		return nil, fmt.Errorf("error generando recompensas diarias: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, fmt.Errorf("no se pudo generar las recompensas diarias")
	}

	var result DailyRewardResult
	err = rows.Scan(
		&result.RewardType,
		&result.RewardValue,
		&result.RewardDescription,
		&result.ConsecutiveDays,
		&result.TotalRewardsToday,
	)
	if err != nil {
		return nil, fmt.Errorf("error escaneando resultado: %w", err)
	}

	s.logger.Info("Recompensas diarias generadas",
		zap.String("player_id", playerID.String()),
		zap.String("reward_type", result.RewardType),
		zap.Int("reward_value", result.RewardValue),
		zap.Int("consecutive_days", result.ConsecutiveDays),
	)

	return &result, nil
}

// CleanupInactiveData limpia datos inactivos
func (s *GameMechanicsService) CleanupInactiveData(daysOld int) ([]CleanupResult, error) {
	rows, err := s.villageRepo.CleanupInactiveData(daysOld)
	if err != nil {
		return nil, fmt.Errorf("error limpiando datos inactivos: %w", err)
	}
	defer rows.Close()

	var results []CleanupResult
	for rows.Next() {
		var result CleanupResult
		err := rows.Scan(
			&result.TableName,
			&result.RecordsDeleted,
			&result.CleanupType,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando resultado: %w", err)
		}
		results = append(results, result)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterando resultados: %w", err)
	}

	s.logger.Info("Limpieza de datos inactivos completada",
		zap.Int("days_old", daysOld),
		zap.Int("tables_processed", len(results)),
	)

	return results, nil
}

// UpdateVillageResources actualiza los recursos de una aldea basándose en la producción calculada
func (s *GameMechanicsService) UpdateVillageResources(villageID uuid.UUID) error {
	// Calcular producción actual
	production, err := s.CalculateResourceProduction(villageID)
	if err != nil {
		return fmt.Errorf("error calculando producción: %w", err)
	}

	// Obtener recursos actuales
	village, err := s.villageRepo.GetVillageByID(villageID)
	if err != nil {
		return fmt.Errorf("error obteniendo aldea: %w", err)
	}

	// Calcular nuevos recursos (producción por hora)
	now := time.Now()
	lastUpdate := village.Resources.LastUpdated
	hoursElapsed := now.Sub(lastUpdate).Hours()

	if hoursElapsed < 0.1 { // Menos de 6 minutos
		return nil // No actualizar si no ha pasado suficiente tiempo
	}

	newWood := village.Resources.Wood + int(float64(production.WoodProduction)*hoursElapsed)
	newStone := village.Resources.Stone + int(float64(production.StoneProduction)*hoursElapsed)
	newFood := village.Resources.Food + int(float64(production.FoodProduction)*hoursElapsed)
	newGold := village.Resources.Gold + int(float64(production.GoldProduction)*hoursElapsed)

	// Actualizar recursos
	err = s.villageRepo.UpdateResources(villageID, newWood, newStone, newFood, newGold)
	if err != nil {
		return fmt.Errorf("error actualizando recursos: %w", err)
	}

	s.logger.Info("Recursos de aldea actualizados",
		zap.String("village_id", villageID.String()),
		zap.Float64("hours_elapsed", hoursElapsed),
		zap.Int("wood_produced", int(float64(production.WoodProduction)*hoursElapsed)),
		zap.Int("stone_produced", int(float64(production.StoneProduction)*hoursElapsed)),
		zap.Int("food_produced", int(float64(production.FoodProduction)*hoursElapsed)),
		zap.Int("gold_produced", int(float64(production.GoldProduction)*hoursElapsed)),
	)

	return nil
}
