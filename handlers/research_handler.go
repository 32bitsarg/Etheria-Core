package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"server-backend/models"
	"server-backend/repository"
	"server-backend/services"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type ResearchHandler struct {
	researchRepo  *repository.ResearchRepository
	villageRepo   *repository.VillageRepository
	researchService *services.ResearchService
	logger        *zap.Logger
}

func NewResearchHandler(researchRepo *repository.ResearchRepository, villageRepo *repository.VillageRepository, researchService *services.ResearchService, logger *zap.Logger) *ResearchHandler {
	return &ResearchHandler{
		researchRepo: researchRepo,
		villageRepo:  villageRepo,
		researchService: researchService,
		logger:       logger,
	}
}

// GetTechnologies obtiene todas las tecnologías disponibles
func (h *ResearchHandler) GetTechnologies(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")
	subCategory := r.URL.Query().Get("sub_category")

	technologies, err := h.researchRepo.GetTechnologies(category, subCategory)
	if err != nil {
		h.logger.Error("error obteniendo tecnologías", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success":      true,
		"technologies": technologies,
		"count":        len(technologies),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetTechnology obtiene una tecnología específica
func (h *ResearchHandler) GetTechnology(w http.ResponseWriter, r *http.Request) {
	technologyID := chi.URLParam(r, "id")

	technology, err := h.researchRepo.GetTechnology(technologyID)
	if err != nil {
		h.logger.Error("error obteniendo tecnología", zap.Error(err))
		http.Error(w, "Tecnología no encontrada", http.StatusNotFound)
		return
	}

	// Obtener efectos y costos
	effects, _ := h.researchRepo.GetTechnologyEffects(technologyID)
	costs, _ := h.researchRepo.GetTechnologyCosts(technologyID, 1)

	response := map[string]interface{}{
		"success":    true,
		"technology": technology,
		"effects":    effects,
		"costs":      costs,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetPlayerTechnologies obtiene las tecnologías de un jugador
func (h *ResearchHandler) GetPlayerTechnologies(w http.ResponseWriter, r *http.Request) {
	playerID := r.Context().Value("player_id").(int)
	playerIDStr := strconv.Itoa(playerID)

	playerTechs, err := h.researchRepo.GetPlayerTechnologies(playerIDStr)
	if err != nil {
		h.logger.Error("error obteniendo tecnologías del jugador", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	// Obtener detalles completos de cada tecnología
	var technologiesWithDetails []models.TechnologyWithDetails
	for _, pt := range playerTechs {
		tech, err := h.researchRepo.GetTechnology(pt.TechnologyID)
		if err != nil {
			continue
		}

		effects, _ := h.researchRepo.GetTechnologyEffects(pt.TechnologyID)
		costs, _ := h.researchRepo.GetTechnologyCosts(pt.TechnologyID, pt.Level+1)
		requirements, _ := h.researchRepo.GetTechnologyRequirements(pt.TechnologyID)

		canResearch := pt.Level < tech.MaxLevel && !pt.IsResearching
		if canResearch {
			canResearch, _, _ = h.researchRepo.CheckTechnologyRequirements(playerIDStr, pt.TechnologyID)
		}

		techWithDetails := models.TechnologyWithDetails{
			Technology:    tech,
			PlayerLevel:   pt.Level,
			IsResearching: pt.IsResearching,
			CanResearch:   canResearch,
			Requirements:  requirements,
			Effects:       effects,
			Costs:         costs,
			Progress:      pt.Progress,
		}

		// Calcular tiempo de investigación si está investigando
		if pt.IsResearching && pt.CompletedAt != nil {
			techWithDetails.ResearchTime = int(time.Until(*pt.CompletedAt).Seconds())
			if techWithDetails.ResearchTime < 0 {
				techWithDetails.ResearchTime = 0
			}
		}

		technologiesWithDetails = append(technologiesWithDetails, techWithDetails)
	}

	response := map[string]interface{}{
		"success":      true,
		"technologies": technologiesWithDetails,
		"count":        len(technologiesWithDetails),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// StartResearch inicia la investigación de una tecnología
func (h *ResearchHandler) StartResearch(w http.ResponseWriter, r *http.Request) {
	playerID := r.Context().Value("player_id").(int)
	playerIDStr := strconv.Itoa(playerID)

	var request struct {
		TechnologyID string `json:"technology_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Datos de entrada inválidos", http.StatusBadRequest)
		return
	}

	// Verificar que la tecnología existe
	technology, err := h.researchRepo.GetTechnology(request.TechnologyID)
	if err != nil {
		http.Error(w, "Tecnología no encontrada", http.StatusNotFound)
		return
	}

	// Verificar requisitos
	canResearch, missingReqs, err := h.researchRepo.CheckTechnologyRequirements(playerIDStr, request.TechnologyID)
	if err != nil {
		h.logger.Error("error verificando requisitos", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	if !canResearch {
		response := map[string]interface{}{
			"success":              false,
			"error":                "No cumples los requisitos para esta tecnología",
			"missing_requirements": missingReqs,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Obtener costos
	playerTech, _ := h.researchRepo.GetPlayerTechnology(playerIDStr, request.TechnologyID)
	currentLevel := 0
	if playerTech != nil {
		currentLevel = playerTech.Level
	}

	costs, err := h.researchRepo.GetTechnologyCosts(request.TechnologyID, currentLevel+1)
	if err != nil {
		h.logger.Error("error obteniendo costos", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	// Iniciar investigación
	err = h.researchRepo.StartResearch(playerIDStr, request.TechnologyID)
	if err != nil {
		h.logger.Error("error iniciando investigación", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success":    true,
		"technology": technology,
		"costs":      costs,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CompleteResearch completa la investigación de una tecnología
func (h *ResearchHandler) CompleteResearch(w http.ResponseWriter, r *http.Request) {
	playerID := r.Context().Value("player_id").(int)
	playerIDStr := strconv.Itoa(playerID)

	var request struct {
		TechnologyID string `json:"technology_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Datos de entrada inválidos", http.StatusBadRequest)
		return
	}

	err := h.researchRepo.CompleteResearch(playerIDStr, request.TechnologyID)
	if err != nil {
		h.logger.Error("error completando investigación", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	playerTech, _ := h.researchRepo.GetPlayerTechnology(playerIDStr, request.TechnologyID)
	response := map[string]interface{}{
		"success": true,
		"level":   playerTech.Level,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CancelResearch cancela la investigación actual
func (h *ResearchHandler) CancelResearch(w http.ResponseWriter, r *http.Request) {
	playerID := r.Context().Value("player_id").(int)
	playerIDStr := strconv.Itoa(playerID)

	err := h.researchRepo.CancelResearch(playerIDStr)
	if err != nil {
		h.logger.Error("error cancelando investigación", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetResearchQueue obtiene la cola de investigación de un jugador
func (h *ResearchHandler) GetResearchQueue(w http.ResponseWriter, r *http.Request) {
	playerID := r.Context().Value("player_id").(int)
	playerIDStr := strconv.Itoa(playerID)

	queue, err := h.researchRepo.GetResearchQueue(playerIDStr)
	if err != nil {
		h.logger.Error("error obteniendo cola de investigación", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"queue":   queue,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetResearchStatistics obtiene estadísticas de investigación de un jugador
func (h *ResearchHandler) GetResearchStatistics(w http.ResponseWriter, r *http.Request) {
	playerID := r.Context().Value("player_id").(int)
	playerIDStr := strconv.Itoa(playerID)

	stats, err := h.researchRepo.GetResearchStatistics(playerIDStr)
	if err != nil {
		h.logger.Error("error obteniendo estadísticas de investigación", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"stats":   stats,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetTechnologyRankings obtiene el ranking de tecnologías
func (h *ResearchHandler) GetTechnologyRankings(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	rankings, err := h.researchRepo.GetTechnologyRankings(limit)
	if err != nil {
		h.logger.Error("error obteniendo ranking de tecnologías", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success":  true,
		"rankings": rankings,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetResearchRecommendations obtiene recomendaciones de investigación
func (h *ResearchHandler) GetResearchRecommendations(w http.ResponseWriter, r *http.Request) {
	playerID := r.Context().Value("player_id").(int)
	playerIDStr := strconv.Itoa(playerID)

	recommendations, err := h.researchRepo.GetResearchRecommendations(playerIDStr)
	if err != nil {
		h.logger.Error("error obteniendo recomendaciones de investigación", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success":         true,
		"recommendations": recommendations,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetTechnologyTree obtiene el árbol de tecnologías
func (h *ResearchHandler) GetTechnologyTree(w http.ResponseWriter, r *http.Request) {
	playerID := r.Context().Value("player_id").(int)
	playerIDStr := strconv.Itoa(playerID)

	// Obtener todas las tecnologías
	technologies, err := h.researchRepo.GetTechnologies("", "")
	if err != nil {
		h.logger.Error("error obteniendo tecnologías", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	// Obtener tecnologías del jugador
	playerTechs, err := h.researchRepo.GetPlayerTechnologies(playerIDStr)
	if err != nil {
		h.logger.Error("error obteniendo tecnologías del jugador", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	// Crear mapa de tecnologías del jugador para acceso rápido
	playerTechMap := make(map[string]models.PlayerTechnology)
	for _, pt := range playerTechs {
		playerTechMap[pt.TechnologyID] = pt
	}

	// Construir árbol de tecnologías
	var tree []map[string]interface{}
	for _, tech := range technologies {
		playerTech, hasTech := playerTechMap[tech.ID]

		node := map[string]interface{}{
			"technology": tech,
			"unlocked":   hasTech,
		}

		if hasTech {
			node["player_level"] = playerTech.Level
			node["is_researching"] = playerTech.IsResearching
		}

		tree = append(tree, node)
	}

	response := map[string]interface{}{
		"success": true,
		"tree":    tree,
		"count":   len(tree),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetTechnologyDetails obtiene los detalles completos de una tecnología
func (h *ResearchHandler) GetTechnologyDetails(w http.ResponseWriter, r *http.Request) {
	technologyID := chi.URLParam(r, "id")

	technology, err := h.researchRepo.GetTechnology(technologyID)
	if err != nil {
		http.Error(w, "Tecnología no encontrada", http.StatusNotFound)
		return
	}

	effects, _ := h.researchRepo.GetTechnologyEffects(technologyID)
	costs, _ := h.researchRepo.GetTechnologyCosts(technologyID, 1)
	requirements, _ := h.researchRepo.GetTechnologyRequirements(technologyID)

	response := map[string]interface{}{
		"success":      true,
		"technology":   technology,
		"effects":      effects,
		"costs":        costs,
		"requirements": requirements,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetResearchHistory obtiene el historial de investigación de un jugador
func (h *ResearchHandler) GetResearchHistory(w http.ResponseWriter, r *http.Request) {
	playerID := r.Context().Value("player_id").(int)
	playerIDStr := strconv.Itoa(playerID)

	history, err := h.researchRepo.GetPlayerTechnologies(playerIDStr)
	if err != nil {
		h.logger.Error("error obteniendo historial de investigación", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"history": history,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetResearchBonuses obtiene los bonos de investigación
func (h *ResearchHandler) GetResearchBonuses(w http.ResponseWriter, r *http.Request) {
	// TODO: Implementar obtención de bonos de investigación
	// Por ahora retornamos una lista vacía
	response := map[string]interface{}{
		"success": true,
		"bonuses": []interface{}{},
		"count":   0,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
