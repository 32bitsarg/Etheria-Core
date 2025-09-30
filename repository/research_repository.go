package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"server-backend/models"

	"go.uber.org/zap"
)

type ResearchRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewResearchRepository(db *sql.DB, logger *zap.Logger) *ResearchRepository {
	return &ResearchRepository{
		db:     db,
		logger: logger,
	}
}

// GetTechnologies obtiene todas las tecnologías disponibles
func (r *ResearchRepository) GetTechnologies(category, subCategory string) ([]models.Technology, error) {
	query := `
		SELECT id, name, description, category, sub_category, level, max_level, 
		       research_time, research_cost, requirements, effects, icon, color, 
		       is_active, is_special, created_at, updated_at
		FROM technologies
		WHERE is_active = true
	`

	args := []interface{}{}
	if category != "" {
		query += " AND category = $1"
		args = append(args, category)
	}
	if subCategory != "" {
		query += " AND sub_category = $2"
		args = append(args, subCategory)
	}

	query += " ORDER BY category, sub_category, level"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo tecnologías: %w", err)
	}
	defer rows.Close()

	var technologies []models.Technology
	for rows.Next() {
		var tech models.Technology
		err := rows.Scan(
			&tech.ID, &tech.Name, &tech.Description, &tech.Category, &tech.SubCategory,
			&tech.Level, &tech.MaxLevel, &tech.ResearchTime, &tech.ResearchCost,
			&tech.Requirements, &tech.Effects, &tech.Icon, &tech.Color,
			&tech.IsActive, &tech.IsSpecial, &tech.CreatedAt, &tech.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando tecnología: %w", err)
		}
		technologies = append(technologies, tech)
	}

	return technologies, nil
}

// GetTechnology obtiene una tecnología específica
func (r *ResearchRepository) GetTechnology(technologyID string) (*models.Technology, error) {
	query := `
		SELECT id, name, description, category, sub_category, level, max_level, 
		       research_time, research_cost, requirements, effects, icon, color, 
		       is_active, is_special, created_at, updated_at
		FROM technologies
		WHERE id = $1 AND is_active = true
	`

	var tech models.Technology
	err := r.db.QueryRow(query, technologyID).Scan(
		&tech.ID, &tech.Name, &tech.Description, &tech.Category, &tech.SubCategory,
		&tech.Level, &tech.MaxLevel, &tech.ResearchTime, &tech.ResearchCost,
		&tech.Requirements, &tech.Effects, &tech.Icon, &tech.Color,
		&tech.IsActive, &tech.IsSpecial, &tech.CreatedAt, &tech.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("tecnología no encontrada")
		}
		return nil, fmt.Errorf("error obteniendo tecnología: %w", err)
	}

	return &tech, nil
}

// GetPlayerTechnologies obtiene las tecnologías de un jugador
func (r *ResearchRepository) GetPlayerTechnologies(playerID string) ([]models.PlayerTechnology, error) {
	query := `
		SELECT id, player_id, technology_id, level, is_researching, started_at, 
		       completed_at, progress, created_at, updated_at
		FROM player_technologies
		WHERE player_id = $1
		ORDER BY technology_id, level
	`

	rows, err := r.db.Query(query, playerID)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo tecnologías del jugador: %w", err)
	}
	defer rows.Close()

	var playerTechs []models.PlayerTechnology
	for rows.Next() {
		var pt models.PlayerTechnology
		err := rows.Scan(
			&pt.ID, &pt.PlayerID, &pt.TechnologyID, &pt.Level, &pt.IsResearching,
			&pt.StartedAt, &pt.CompletedAt, &pt.Progress, &pt.CreatedAt, &pt.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando tecnología del jugador: %w", err)
		}
		playerTechs = append(playerTechs, pt)
	}

	return playerTechs, nil
}

// GetPlayerTechnology obtiene una tecnología específica de un jugador
func (r *ResearchRepository) GetPlayerTechnology(playerID string, technologyID string) (*models.PlayerTechnology, error) {
	query := `
		SELECT id, player_id, technology_id, level, is_researching, started_at, 
		       completed_at, progress, created_at, updated_at
		FROM player_technologies
		WHERE player_id = $1 AND technology_id = $2
	`

	var pt models.PlayerTechnology
	err := r.db.QueryRow(query, playerID, technologyID).Scan(
		&pt.ID, &pt.PlayerID, &pt.TechnologyID, &pt.Level, &pt.IsResearching,
		&pt.StartedAt, &pt.CompletedAt, &pt.Progress, &pt.CreatedAt, &pt.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No tiene esta tecnología
		}
		return nil, fmt.Errorf("error obteniendo tecnología del jugador: %w", err)
	}

	return &pt, nil
}

// StartResearch inicia la investigación de una tecnología
func (r *ResearchRepository) StartResearch(playerID string, technologyID string) error {
	// Verificar que no esté investigando otra tecnología
	var count int
	err := r.db.QueryRow(
		"SELECT COUNT(*) FROM player_technologies WHERE player_id = $1 AND is_researching = true",
		playerID,
	).Scan(&count)
	if err != nil {
		return fmt.Errorf("error verificando investigación activa: %w", err)
	}

	if count > 0 {
		return fmt.Errorf("ya tienes una investigación en progreso")
	}

	// Obtener la tecnología
	tech, err := r.GetTechnology(technologyID)
	if err != nil {
		return fmt.Errorf("error obteniendo tecnología: %w", err)
	}

	// Obtener el nivel actual del jugador
	playerTech, err := r.GetPlayerTechnology(playerID, technologyID)
	if err != nil {
		return fmt.Errorf("error obteniendo tecnología del jugador: %w", err)
	}

	currentLevel := 0
	if playerTech != nil {
		currentLevel = playerTech.Level
	}

	if currentLevel >= tech.MaxLevel {
		return fmt.Errorf("ya tienes el nivel máximo de esta tecnología")
	}

	// Verificar requisitos
	canResearch, reasons, err := r.CheckTechnologyRequirements(playerID, technologyID)
	if err != nil {
		return fmt.Errorf("error verificando requisitos: %w", err)
	}
	if !canResearch {
		return fmt.Errorf("no cumples los requisitos: %v", reasons)
	}

	// Verificar costos y recursos
	costs, err := r.GetTechnologyCosts(technologyID, currentLevel+1)
	if err != nil {
		return fmt.Errorf("error obteniendo costos: %w", err)
	}

	// Verificar que tenga suficientes recursos
	for _, cost := range costs {
		// Aquí se implementaría la verificación de recursos del jugador
		// Por ahora asumimos que tiene suficientes
		_ = cost
	}

	// Calcular tiempo de investigación con bonificaciones
	researchTime := r.CalculateResearchTime(tech.ResearchTime, playerID, tech.Category)

	now := time.Now()
	completedAt := now.Add(time.Duration(researchTime) * time.Second)

	// Insertar o actualizar la tecnología del jugador
	if playerTech == nil {
		query := `
			INSERT INTO player_technologies (player_id, technology_id, level, is_researching, 
			                                started_at, completed_at, progress, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`
		_, err = r.db.Exec(query, playerID, technologyID, 0, true, now, completedAt, 0, now, now)
	} else {
		query := `
			UPDATE player_technologies 
			SET is_researching = true, started_at = $1, completed_at = $2, progress = 0, updated_at = $3
			WHERE player_id = $4 AND technology_id = $5
		`
		_, err = r.db.Exec(query, now, completedAt, now, playerID, technologyID)
	}

	if err != nil {
		return fmt.Errorf("error iniciando investigación: %w", err)
	}

	// Agregar a la cola de investigación
	query := `
		INSERT INTO research_queue (player_id, technology_id, priority, added_at, status, progress, time_remaining)
		VALUES ($1, $2, 1, $3, 'researching', 0, $4)
		ON CONFLICT (player_id) DO UPDATE SET
		technology_id = $2, added_at = $3, status = 'researching', progress = 0, time_remaining = $4
	`
	_, err = r.db.Exec(query, playerID, technologyID, now, researchTime)
	if err != nil {
		r.logger.Warn("Error agregando a cola de investigación", zap.Error(err))
		// No fallamos la investigación por este error
	}

	return nil
}

// CompleteResearch completa la investigación de una tecnología
func (r *ResearchRepository) CompleteResearch(playerID string, technologyID string) error {
	// Obtener la tecnología del jugador
	playerTech, err := r.GetPlayerTechnology(playerID, technologyID)
	if err != nil {
		return fmt.Errorf("error obteniendo tecnología del jugador: %w", err)
	}

	if playerTech == nil || !playerTech.IsResearching {
		return fmt.Errorf("no hay investigación activa para esta tecnología")
	}

	// Verificar que el tiempo haya pasado
	if playerTech.CompletedAt != nil && time.Now().Before(*playerTech.CompletedAt) {
		return fmt.Errorf("la investigación aún no ha terminado")
	}

	// Obtener la tecnología base
	_, err = r.GetTechnology(technologyID)
	if err != nil {
		return fmt.Errorf("error obteniendo tecnología: %w", err)
	}

	// Actualizar nivel
	newLevel := playerTech.Level + 1
	now := time.Now()

	query := `
		UPDATE player_technologies 
		SET level = $1, is_researching = false, started_at = NULL, completed_at = NULL, 
		    progress = 0, updated_at = $2
		WHERE player_id = $3 AND technology_id = $4
	`
	_, err = r.db.Exec(query, newLevel, now, playerID, technologyID)
	if err != nil {
		return fmt.Errorf("error completando investigación: %w", err)
	}

	// Actualizar cola de investigación
	query = `
		UPDATE research_queue
		SET status = 'completed', time_remaining = 0
		WHERE player_id = $1 AND technology_id = $2 AND status = 'researching'
	`
	_, err = r.db.Exec(query, playerID, technologyID)
	if err != nil {
		r.logger.Warn("Error actualizando cola", zap.Error(err))
	}

	// Registrar en historial
	if playerTech.StartedAt != nil {
		duration := int(time.Since(*playerTech.StartedAt).Seconds())
		costs, _ := r.GetTechnologyCosts(technologyID, newLevel)
		costsJSON, _ := json.Marshal(costs)

		query = `
			INSERT INTO research_history (player_id, technology_id, level, start_time, end_time, duration, cost, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`
		_, err = r.db.Exec(query, playerID, technologyID, newLevel, playerTech.StartedAt, now, duration, string(costsJSON), now)
		if err != nil {
			r.logger.Warn("Error registrando historial", zap.Error(err))
		}
	}

	// Verificar logros
	r.checkResearchAchievements(playerID)

	return nil
}

// CancelResearch cancela la investigación actual
func (r *ResearchRepository) CancelResearch(playerID string) error {
	// Obtener investigación activa
	var technologyID string
	var startedAt time.Time
	err := r.db.QueryRow(
		"SELECT technology_id, started_at FROM player_technologies WHERE player_id = $1 AND is_researching = true",
		playerID,
	).Scan(&technologyID, &startedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("no hay investigación activa")
		}
		return fmt.Errorf("error obteniendo investigación activa: %w", err)
	}

	// Calcular reembolso (porcentaje del tiempo transcurrido)
	elapsed := time.Since(startedAt)
	// Aquí se implementaría la lógica de reembolso de recursos
	// Por ahora solo registramos el tiempo transcurrido
	r.logger.Info("Investigación cancelada",
		zap.String("playerID", playerID),
		zap.String("technologyID", technologyID),
		zap.Duration("elapsed", elapsed))

	// Cancelar investigación
	query := `
		UPDATE player_technologies 
		SET is_researching = false, started_at = NULL, completed_at = NULL, progress = 0, updated_at = $1
		WHERE player_id = $2 AND technology_id = $3
	`
	_, err = r.db.Exec(query, time.Now(), playerID, technologyID)
	if err != nil {
		return fmt.Errorf("error cancelando investigación: %w", err)
	}

	// Actualizar cola
	query = `
		UPDATE research_queue
		SET status = 'cancelled', time_remaining = 0
		WHERE player_id = $1 AND technology_id = $2 AND status = 'researching'
	`
	_, err = r.db.Exec(query, playerID, technologyID)
	if err != nil {
		r.logger.Warn("Error actualizando cola", zap.Error(err))
	}

	return nil
}

// GetResearchQueue obtiene la cola de investigación de un jugador
func (r *ResearchRepository) GetResearchQueue(playerID string) ([]models.ResearchQueueItem, error) {
	query := `
		SELECT id, player_id, technology_id, priority, added_at, started_at, status, progress, time_remaining
		FROM research_queue
		WHERE player_id = $1
		ORDER BY priority, added_at
	`

	rows, err := r.db.Query(query, playerID)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo cola de investigación: %w", err)
	}
	defer rows.Close()

	var queue []models.ResearchQueueItem
	for rows.Next() {
		var item models.ResearchQueueItem
		err := rows.Scan(
			&item.ID, &item.PlayerID, &item.TechnologyID, &item.Priority,
			&item.AddedAt, &item.StartedAt, &item.Status, &item.Progress, &item.TimeRemaining,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando item de cola: %w", err)
		}
		queue = append(queue, item)
	}

	return queue, nil
}

// GetTechnologyEffects obtiene los efectos de una tecnología
func (r *ResearchRepository) GetTechnologyEffects(technologyID string) ([]models.TechnologyEffect, error) {
	query := `
		SELECT id, technology_id, effect_type, target, value, is_percentage, level, data
		FROM technology_effects
		WHERE technology_id = $1
		ORDER BY level, effect_type
	`

	rows, err := r.db.Query(query, technologyID)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo efectos: %w", err)
	}
	defer rows.Close()

	var effects []models.TechnologyEffect
	for rows.Next() {
		var effect models.TechnologyEffect
		var dataJSON string
		err := rows.Scan(
			&effect.ID, &effect.TechnologyID, &effect.EffectType, &effect.Target,
			&effect.Value, &effect.IsPercentage, &effect.Level, &dataJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando efecto: %w", err)
		}

		// Parsear data JSON
		if dataJSON != "" {
			json.Unmarshal([]byte(dataJSON), &effect.Data)
		}

		effects = append(effects, effect)
	}

	return effects, nil
}

// GetTechnologyCosts obtiene los costos de una tecnología
func (r *ResearchRepository) GetTechnologyCosts(technologyID string, level int) ([]models.TechnologyCost, error) {
	query := `
		SELECT id, technology_id, level, resource_type, amount
		FROM technology_costs
		WHERE technology_id = $1 AND level = $2
		ORDER BY resource_type
	`

	rows, err := r.db.Query(query, technologyID, level)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo costos: %w", err)
	}
	defer rows.Close()

	var costs []models.TechnologyCost
	for rows.Next() {
		var cost models.TechnologyCost
		err := rows.Scan(
			&cost.ID, &cost.TechnologyID, &cost.Level, &cost.ResourceType, &cost.Amount,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando costo: %w", err)
		}
		costs = append(costs, cost)
	}

	return costs, nil
}

// GetTechnologyRequirements obtiene los requisitos de una tecnología
func (r *ResearchRepository) GetTechnologyRequirements(technologyID string) ([]models.TechnologyRequirement, error) {
	query := `
		SELECT id, technology_id, required_tech_id, required_level, required_village
		FROM technology_requirements
		WHERE technology_id = $1
		ORDER BY required_tech_id
	`

	rows, err := r.db.Query(query, technologyID)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo requisitos: %w", err)
	}
	defer rows.Close()

	var requirements []models.TechnologyRequirement
	for rows.Next() {
		var req models.TechnologyRequirement
		err := rows.Scan(
			&req.ID, &req.TechnologyID, &req.RequiredTechID, &req.RequiredLevel, &req.RequiredVillage,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando requisito: %w", err)
		}
		requirements = append(requirements, req)
	}

	return requirements, nil
}

// CheckTechnologyRequirements verifica si un jugador cumple los requisitos para una tecnología
func (r *ResearchRepository) CheckTechnologyRequirements(playerID string, technologyID string) (bool, []string, error) {
	requirements, err := r.GetTechnologyRequirements(technologyID)
	if err != nil {
		return false, nil, fmt.Errorf("error obteniendo requisitos: %w", err)
	}

	var reasons []string
	for _, req := range requirements {
		// Verificar tecnología requerida - convertir int a string para la consulta
		requiredTechID := fmt.Sprintf("%d", req.RequiredTechID)
		playerTech, err := r.GetPlayerTechnology(playerID, requiredTechID)
		if err != nil {
			return false, nil, fmt.Errorf("error verificando tecnología requerida: %w", err)
		}

		if playerTech == nil || playerTech.Level < req.RequiredLevel {
			tech, _ := r.GetTechnology(requiredTechID)
			techName := "Tecnología desconocida"
			if tech != nil {
				techName = tech.Name
			}
			reasons = append(reasons, fmt.Sprintf("Requiere %s nivel %d", techName, req.RequiredLevel))
		}
	}

	return len(reasons) == 0, reasons, nil
}

// GetResearchStatistics obtiene estadísticas de investigación de un jugador
func (r *ResearchRepository) GetResearchStatistics(playerID string) (*models.ResearchStatistics, error) {
	// Obtener tecnologías del jugador
	playerTechs, err := r.GetPlayerTechnologies(playerID)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo tecnologías del jugador: %w", err)
	}

	// Calcular estadísticas
	stats := &models.ResearchStatistics{
		PlayerID:           playerID,
		TotalTechnologies:  len(playerTechs),
		MilitaryTechs:      0,
		EconomicTechs:      0,
		SocialTechs:        0,
		ScientificTechs:    0,
		TotalLevels:        0,
		ResearchTime:       0,
		ResourcesSpent:     "{}",
		LastResearchAt:     nil,
		ResearchEfficiency: 0.0,
	}

	for _, pt := range playerTechs {
		tech, err := r.GetTechnology(pt.TechnologyID)
		if err != nil {
			continue
		}

		stats.TotalLevels += pt.Level

		switch tech.Category {
		case "military":
			stats.MilitaryTechs++
		case "economic":
			stats.EconomicTechs++
		case "social":
			stats.SocialTechs++
		case "scientific":
			stats.ScientificTechs++
		}

		if pt.CompletedAt != nil && (stats.LastResearchAt == nil || pt.CompletedAt.After(*stats.LastResearchAt)) {
			stats.LastResearchAt = pt.CompletedAt
		}
	}

	// Calcular eficiencia (tecnologías completadas vs tiempo total)
	if len(playerTechs) > 0 {
		stats.ResearchEfficiency = float64(stats.TotalLevels) / float64(len(playerTechs))
	}

	return stats, nil
}

// GetTechnologyRankings obtiene el ranking de tecnologías
func (r *ResearchRepository) GetTechnologyRankings(limit int) ([]models.TechnologyRanking, error) {
	query := `
		SELECT 
			pt.player_id,
			p.username as player_name,
			COUNT(pt.technology_id) as total_techs,
			SUM(pt.level) as total_levels,
			COUNT(CASE WHEN t.category = 'military' THEN 1 END) as military_score,
			COUNT(CASE WHEN t.category = 'economic' THEN 1 END) as economic_score,
			COUNT(CASE WHEN t.category = 'social' THEN 1 END) as social_score,
			COUNT(CASE WHEN t.category = 'scientific' THEN 1 END) as scientific_score
		FROM player_technologies pt
		JOIN players p ON pt.player_id = p.id
		JOIN technologies t ON pt.technology_id = t.id
		GROUP BY pt.player_id, p.username
		ORDER BY total_levels DESC, total_techs DESC
		LIMIT $1
	`

	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo rankings: %w", err)
	}
	defer rows.Close()

	var rankings []models.TechnologyRanking
	rank := 1
	for rows.Next() {
		var ranking models.TechnologyRanking
		var playerIDStr string
		err := rows.Scan(
			&playerIDStr, &ranking.PlayerName, &ranking.TotalTechs, &ranking.TotalLevels,
			&ranking.MilitaryScore, &ranking.EconomicScore, &ranking.SocialScore, &ranking.ScientificScore,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando ranking: %w", err)
		}

		// Convertir string a int para el modelo
		playerID, _ := strconv.Atoi(playerIDStr)
		ranking.PlayerID = playerID
		ranking.Rank = rank
		rankings = append(rankings, ranking)
		rank++
	}

	return rankings, nil
}

// GetResearchRecommendations obtiene recomendaciones de investigación
func (r *ResearchRepository) GetResearchRecommendations(playerID string) ([]models.ResearchRecommendation, error) {
	// Obtener tecnologías disponibles
	technologies, err := r.GetTechnologies("", "")
	if err != nil {
		return nil, fmt.Errorf("error obteniendo tecnologías: %w", err)
	}

	var recommendations []models.ResearchRecommendation
	for _, tech := range technologies {
		// Verificar si ya tiene esta tecnología al máximo nivel
		playerTech, err := r.GetPlayerTechnology(playerID, tech.ID)
		if err != nil {
			continue
		}

		if playerTech != nil && playerTech.Level >= tech.MaxLevel {
			continue
		}

		// Verificar requisitos
		canResearch, _, err := r.CheckTechnologyRequirements(playerID, tech.ID)
		if err != nil || !canResearch {
			continue
		}

		// Calcular prioridad
		priority := r.calculatePriority(&tech, playerID)
		reason := r.generateRecommendationReason(&tech)

		recommendations = append(recommendations, models.ResearchRecommendation{
			TechnologyID:  tech.ID,
			Priority:      priority,
			Reason:        reason,
			EstimatedTime: tech.ResearchTime,
		})
	}

	// Ordenar por prioridad
	// (Aquí se implementaría el ordenamiento)

	return recommendations, nil
}

// CalculateResearchTime calcula el tiempo de investigación con bonificaciones
func (r *ResearchRepository) CalculateResearchTime(baseTime int, playerID string, category string) int {
	// Obtener bonificaciones del jugador
	// Por ahora implementamos bonificaciones básicas
	bonus := 1.0

	// Bonificación por tecnologías de la misma categoría
	playerTechs, err := r.GetPlayerTechnologies(playerID)
	if err == nil {
		for _, pt := range playerTechs {
			tech, err := r.GetTechnology(pt.TechnologyID)
			if err == nil && tech.Category == category {
				bonus -= 0.05 // 5% de reducción por cada tecnología de la misma categoría
			}
		}
	}

	// Aplicar límites
	if bonus < 0.5 {
		bonus = 0.5 // Mínimo 50% del tiempo original
	}

	return int(float64(baseTime) * bonus)
}

// calculatePriority calcula la prioridad de una tecnología
func (r *ResearchRepository) calculatePriority(tech *models.Technology, playerID string) int {
	priority := 1

	// Prioridad por categoría
	switch tech.Category {
	case "military":
		priority += 10
	case "economic":
		priority += 8
	case "scientific":
		priority += 6
	case "social":
		priority += 4
	}

	// Prioridad por nivel (tecnologías de nivel bajo son más prioritarias)
	priority += (10 - tech.Level) * 2

	// Prioridad por si es especial
	if tech.IsSpecial {
		priority += 5
	}

	return priority
}

// generateRecommendationReason genera la razón de una recomendación
func (r *ResearchRepository) generateRecommendationReason(tech *models.Technology) string {
	switch tech.Category {
	case "military":
		return "Mejora tus capacidades militares"
	case "economic":
		return "Aumenta tu producción de recursos"
	case "scientific":
		return "Desbloquea nuevas tecnologías"
	case "social":
		return "Mejora tu aldea y población"
	default:
		return "Tecnología recomendada"
	}
}

// checkResearchAchievements verifica logros de investigación
func (r *ResearchRepository) checkResearchAchievements(playerID string) {
	// Obtener estadísticas del jugador
	stats, err := r.GetResearchStatistics(playerID)
	if err != nil {
		return
	}

	// Verificar logros básicos
	achievements := []struct {
		condition bool
		message   string
	}{
		{stats.TotalTechnologies >= 5, "Investigador Novato"},
		{stats.TotalTechnologies >= 10, "Investigador Experto"},
		{stats.TotalTechnologies >= 20, "Maestro Investigador"},
		{stats.TotalLevels >= 50, "Nivel Alto"},
		{stats.MilitaryTechs >= 5, "Estratega Militar"},
		{stats.EconomicTechs >= 5, "Economista"},
	}

	for _, achievement := range achievements {
		if achievement.condition {
			r.logger.Info("Logro de investigación desbloqueado",
				zap.String("playerID", playerID),
				zap.String("achievement", achievement.message))
			// Aquí se enviaría la notificación al jugador
		}
	}
}

// GetTechnologiesByCategory obtiene tecnologías por categoría
func (r *ResearchRepository) GetTechnologiesByCategory(category string) ([]models.Technology, error) {
	return r.GetTechnologies(category, "")
}

// GetTechnologiesByLevel obtiene tecnologías por nivel
func (r *ResearchRepository) GetTechnologiesByLevel(level int) ([]models.Technology, error) {
	query := `
		SELECT id, name, description, category, sub_category, level, max_level, 
		       research_time, research_cost, requirements, effects, icon, color, 
		       is_active, is_special, created_at, updated_at
		FROM technologies
		WHERE is_active = true AND level = $1
		ORDER BY category, sub_category
	`

	rows, err := r.db.Query(query, level)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo tecnologías por nivel: %w", err)
	}
	defer rows.Close()

	var technologies []models.Technology
	for rows.Next() {
		var tech models.Technology
		err := rows.Scan(
			&tech.ID, &tech.Name, &tech.Description, &tech.Category, &tech.SubCategory,
			&tech.Level, &tech.MaxLevel, &tech.ResearchTime, &tech.ResearchCost,
			&tech.Requirements, &tech.Effects, &tech.Icon, &tech.Color,
			&tech.IsActive, &tech.IsSpecial, &tech.CreatedAt, &tech.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando tecnología: %w", err)
		}
		technologies = append(technologies, tech)
	}

	return technologies, nil
}

// GetTechnologiesByCost obtiene tecnologías por rango de costo
func (r *ResearchRepository) GetTechnologiesByCost(minCost, maxCost int) ([]models.Technology, error) {
	query := `
		SELECT DISTINCT t.id, t.name, t.description, t.category, t.sub_category, t.level, t.max_level, 
		       t.research_time, t.research_cost, t.requirements, t.effects, t.icon, t.color, 
		       t.is_active, t.is_special, t.created_at, t.updated_at
		FROM technologies t
		JOIN technology_costs tc ON t.id = tc.technology_id
		WHERE t.is_active = true AND tc.amount BETWEEN $1 AND $2
		ORDER BY t.category, t.sub_category, t.level
	`

	rows, err := r.db.Query(query, minCost, maxCost)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo tecnologías por costo: %w", err)
	}
	defer rows.Close()

	var technologies []models.Technology
	for rows.Next() {
		var tech models.Technology
		err := rows.Scan(
			&tech.ID, &tech.Name, &tech.Description, &tech.Category, &tech.SubCategory,
			&tech.Level, &tech.MaxLevel, &tech.ResearchTime, &tech.ResearchCost,
			&tech.Requirements, &tech.Effects, &tech.Icon, &tech.Color,
			&tech.IsActive, &tech.IsSpecial, &tech.CreatedAt, &tech.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando tecnología: %w", err)
		}
		technologies = append(technologies, tech)
	}

	return technologies, nil
}

// GetTechnologiesByRequirements obtiene tecnologías que requieren una tecnología específica
func (r *ResearchRepository) GetTechnologiesByRequirements(requiredTechnologyID string) ([]models.Technology, error) {
	query := `
		SELECT DISTINCT t.id, t.name, t.description, t.category, t.sub_category, t.level, t.max_level, 
		       t.research_time, t.research_cost, t.requirements, t.effects, t.icon, t.color, 
		       t.is_active, t.is_special, t.created_at, t.updated_at
		FROM technologies t
		JOIN technology_requirements tr ON t.id = tr.technology_id
		WHERE t.is_active = true AND tr.required_tech_id = $1
		ORDER BY t.category, t.sub_category, t.level
	`

	rows, err := r.db.Query(query, requiredTechnologyID)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo tecnologías por requisitos: %w", err)
	}
	defer rows.Close()

	var technologies []models.Technology
	for rows.Next() {
		var tech models.Technology
		err := rows.Scan(
			&tech.ID, &tech.Name, &tech.Description, &tech.Category, &tech.SubCategory,
			&tech.Level, &tech.MaxLevel, &tech.ResearchTime, &tech.ResearchCost,
			&tech.Requirements, &tech.Effects, &tech.Icon, &tech.Color,
			&tech.IsActive, &tech.IsSpecial, &tech.CreatedAt, &tech.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando tecnología: %w", err)
		}
		technologies = append(technologies, tech)
	}

	return technologies, nil
}

// GetTechnologyDetails obtiene los detalles completos de una tecnología
func (r *ResearchRepository) GetTechnologyDetails(technologyID string) (*models.TechnologyDetails, error) {
	// Obtener la tecnología base
	tech, err := r.GetTechnology(technologyID)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo tecnología: %w", err)
	}

	// Obtener efectos
	effects, err := r.GetTechnologyEffects(technologyID)
	if err != nil {
		r.logger.Warn("Error obteniendo efectos de tecnología", zap.Error(err))
		effects = []models.TechnologyEffect{}
	}

	// Obtener requisitos
	requirements, err := r.GetTechnologyRequirements(technologyID)
	if err != nil {
		r.logger.Warn("Error obteniendo requisitos de tecnología", zap.Error(err))
		requirements = []models.TechnologyRequirement{}
	}

	// Obtener costos para diferentes niveles
	var costs []models.TechnologyCost
	for level := 1; level <= 5; level++ {
		levelCosts, err := r.GetTechnologyCosts(technologyID, level)
		if err != nil {
			r.logger.Warn("Error obteniendo costos de tecnología", zap.Error(err))
			continue
		}
		costs = append(costs, levelCosts...)
	}

	return &models.TechnologyDetails{
		Technology:   *tech,
		Effects:      effects,
		Requirements: requirements,
		Costs:        costs,
		LastUpdated:  time.Now(),
	}, nil
}
