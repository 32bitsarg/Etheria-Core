package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"server-backend/models"
	"server-backend/repository"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type RankingHandler struct {
	rankingRepo *repository.RankingRepository
	logger      *zap.Logger
}

func NewRankingHandler(rankingRepo *repository.RankingRepository, logger *zap.Logger) *RankingHandler {
	return &RankingHandler{
		rankingRepo: rankingRepo,
		logger:      logger,
	}
}

// GetRankingsDashboard obtiene el dashboard completo de rankings
func (h *RankingHandler) GetRankingsDashboard(w http.ResponseWriter, r *http.Request) {
	// Obtener categorías activas
	categories, err := h.rankingRepo.GetRankingCategories(true)
	if err != nil {
		h.logger.Error("Error obteniendo categorías", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	// Obtener temporada activa
	seasons, err := h.rankingRepo.GetRankingSeasons("active")
	if err != nil {
		h.logger.Error("Error obteniendo temporadas", zap.Error(err))
		seasons = []models.RankingSeason{}
	}

	// Obtener resumen de estadísticas
	summary, err := h.rankingRepo.GetStatisticsSummary()
	if err != nil {
		h.logger.Error("Error obteniendo resumen", zap.Error(err))
		summary = &models.StatisticsSummary{}
	}

	// Obtener rankings destacados
	featuredRankings, err := h.getFeaturedRankings()
	if err != nil {
		h.logger.Error("Error obteniendo rankings destacados", zap.Error(err))
		featuredRankings = []models.RankingEntry{}
	}

	dashboard := map[string]interface{}{
		"categories":         categories,
		"active_season":      seasons,
		"statistics_summary": summary,
		"featured_rankings":  featuredRankings,
		"last_updated":       time.Now(),
	}

	respondWithJSON(w, http.StatusOK, dashboard)
}

// GetRankingCategories obtiene todas las categorías de ranking
func (h *RankingHandler) GetRankingCategories(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	activeOnly := query.Get("active") == "true"

	categories, err := h.rankingRepo.GetRankingCategories(activeOnly)
	if err != nil {
		h.logger.Error("Error obteniendo categorías", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, categories)
}

// GetRankingCategory obtiene una categoría específica
func (h *RankingHandler) GetRankingCategory(w http.ResponseWriter, r *http.Request) {
	vars := chi.URLParam(r, "categoryID")
	categoryID, err := strconv.Atoi(vars)
	if err != nil {
		h.logger.Error("ID de categoría inválido", zap.Error(err))
		http.Error(w, "ID de categoría inválido", http.StatusBadRequest)
		return
	}

	category, err := h.rankingRepo.GetRankingCategory(categoryID)
	if err != nil {
		h.logger.Error("Error obteniendo categoría", zap.Error(err))
		http.Error(w, "Categoría no encontrada", http.StatusNotFound)
		return
	}

	respondWithJSON(w, http.StatusOK, category)
}

// CreateRankingCategory crea una nueva categoría de ranking
func (h *RankingHandler) CreateRankingCategory(w http.ResponseWriter, r *http.Request) {
	var category models.RankingCategory
	if err := json.NewDecoder(r.Body).Decode(&category); err != nil {
		h.logger.Error("Error decodificando categoría", zap.Error(err))
		http.Error(w, "Datos inválidos", http.StatusBadRequest)
		return
	}

	// Validar categoría
	if err := h.validateRankingCategory(&category); err != nil {
		h.logger.Error("Categoría inválida", zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := h.rankingRepo.CreateRankingCategory(&category)
	if err != nil {
		h.logger.Error("Error creando categoría", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusCreated, category)
}

// UpdateRankingCategory actualiza una categoría de ranking
func (h *RankingHandler) UpdateRankingCategory(w http.ResponseWriter, r *http.Request) {
	vars := chi.URLParam(r, "categoryID")
	categoryID, err := strconv.Atoi(vars)
	if err != nil {
		h.logger.Error("ID de categoría inválido", zap.Error(err))
		http.Error(w, "ID de categoría inválido", http.StatusBadRequest)
		return
	}

	var category models.RankingCategory
	if err := json.NewDecoder(r.Body).Decode(&category); err != nil {
		h.logger.Error("Error decodificando categoría", zap.Error(err))
		http.Error(w, "Datos inválidos", http.StatusBadRequest)
		return
	}

	category.ID = categoryID

	// Validar categoría
	if err := h.validateRankingCategory(&category); err != nil {
		h.logger.Error("Categoría inválida", zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.rankingRepo.UpdateRankingCategory(&category)
	if err != nil {
		h.logger.Error("Error actualizando categoría", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, category)
}

// GetRankingSeasons obtiene las temporadas de ranking
func (h *RankingHandler) GetRankingSeasons(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	status := query.Get("status")

	seasons, err := h.rankingRepo.GetRankingSeasons(status)
	if err != nil {
		h.logger.Error("Error obteniendo temporadas", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, seasons)
}

// GetRankingEntries obtiene las entradas de un ranking
func (h *RankingHandler) GetRankingEntries(w http.ResponseWriter, r *http.Request) {
	vars := chi.URLParam(r, "categoryID")
	categoryID, err := strconv.Atoi(vars)
	if err != nil {
		h.logger.Error("ID de categoría inválido", zap.Error(err))
		http.Error(w, "ID de categoría inválido", http.StatusBadRequest)
		return
	}

	query := r.URL.Query()
	seasonIDStr := query.Get("season_id")
	limitStr := query.Get("limit")

	var seasonID *int
	if seasonIDStr != "" {
		if sid, err := strconv.Atoi(seasonIDStr); err == nil {
			seasonID = &sid
		}
	}

	limit := 100 // Por defecto
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	entries, err := h.rankingRepo.GetRankingEntries(categoryID, seasonID, limit)
	if err != nil {
		h.logger.Error("Error obteniendo entradas", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, entries)
}

// GetPlayerStatistics obtiene las estadísticas de un jugador
func (h *RankingHandler) GetPlayerStatistics(w http.ResponseWriter, r *http.Request) {
	vars := chi.URLParam(r, "player_id")
	playerID, err := strconv.Atoi(vars)
	if err != nil {
		h.logger.Error("ID de jugador inválido", zap.Error(err))
		http.Error(w, "ID de jugador inválido", http.StatusBadRequest)
		return
	}

	stats, err := h.rankingRepo.GetPlayerStatistics(playerID)
	if err != nil {
		h.logger.Error("Error obteniendo estadísticas", zap.Error(err))
		http.Error(w, "Estadísticas no encontradas", http.StatusNotFound)
		return
	}

	respondWithJSON(w, http.StatusOK, stats)
}

// GetAllianceStatistics obtiene las estadísticas de una alianza
func (h *RankingHandler) GetAllianceStatistics(w http.ResponseWriter, r *http.Request) {
	vars := chi.URLParam(r, "allianceID")
	allianceID, err := strconv.Atoi(vars)
	if err != nil {
		h.logger.Error("ID de alianza inválido", zap.Error(err))
		http.Error(w, "ID de alianza inválido", http.StatusBadRequest)
		return
	}

	stats, err := h.rankingRepo.GetAllianceStatistics(allianceID)
	if err != nil {
		h.logger.Error("Error obteniendo estadísticas", zap.Error(err))
		http.Error(w, "Estadísticas no encontradas", http.StatusNotFound)
		return
	}

	respondWithJSON(w, http.StatusOK, stats)
}

// GetVillageStatistics obtiene las estadísticas de una aldea
func (h *RankingHandler) GetVillageStatistics(w http.ResponseWriter, r *http.Request) {
	vars := chi.URLParam(r, "villageID")
	villageID, err := strconv.Atoi(vars)
	if err != nil {
		h.logger.Error("ID de aldea inválido", zap.Error(err))
		http.Error(w, "ID de aldea inválido", http.StatusBadRequest)
		return
	}

	stats, err := h.rankingRepo.GetVillageStatistics(villageID)
	if err != nil {
		h.logger.Error("Error obteniendo estadísticas", zap.Error(err))
		http.Error(w, "Estadísticas no encontradas", http.StatusNotFound)
		return
	}

	respondWithJSON(w, http.StatusOK, stats)
}

// GetWorldStatistics obtiene las estadísticas de un mundo
func (h *RankingHandler) GetWorldStatistics(w http.ResponseWriter, r *http.Request) {
	vars := chi.URLParam(r, "worldID")
	worldID, err := strconv.Atoi(vars)
	if err != nil {
		h.logger.Error("ID de mundo inválido", zap.Error(err))
		http.Error(w, "ID de mundo inválido", http.StatusBadRequest)
		return
	}

	stats, err := h.rankingRepo.GetWorldStatistics(worldID)
	if err != nil {
		h.logger.Error("Error obteniendo estadísticas", zap.Error(err))
		http.Error(w, "Estadísticas no encontradas", http.StatusNotFound)
		return
	}

	respondWithJSON(w, http.StatusOK, stats)
}

// GetStatisticsSummary obtiene el resumen de estadísticas globales
func (h *RankingHandler) GetStatisticsSummary(w http.ResponseWriter, r *http.Request) {
	summary, err := h.rankingRepo.GetStatisticsSummary()
	if err != nil {
		h.logger.Error("Error obteniendo resumen", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, summary)
}

// GetRankingHistory obtiene el historial de posiciones
func (h *RankingHandler) GetRankingHistory(w http.ResponseWriter, r *http.Request) {
	vars := chi.URLParam(r, "categoryID")
	categoryID, err := strconv.Atoi(vars)
	if err != nil {
		h.logger.Error("ID de categoría inválido", zap.Error(err))
		http.Error(w, "ID de categoría inválido", http.StatusBadRequest)
		return
	}

	query := r.URL.Query()
	entityType := query.Get("entity_type")
	entityIDStr := query.Get("entity_id")
	daysStr := query.Get("days")

	entityID, err := strconv.Atoi(entityIDStr)
	if err != nil {
		h.logger.Error("ID de entidad inválido", zap.Error(err))
		http.Error(w, "ID de entidad inválido", http.StatusBadRequest)
		return
	}

	days := 30 // Por defecto
	if daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 {
			days = d
		}
	}

	history, err := h.rankingRepo.GetRankingHistory(categoryID, entityType, entityID, days)
	if err != nil {
		h.logger.Error("Error obteniendo historial", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, history)
}

// GetPlayerRankings obtiene todos los rankings de un jugador
func (h *RankingHandler) GetPlayerRankings(w http.ResponseWriter, r *http.Request) {
	vars := chi.URLParam(r, "player_id")
	playerID, err := strconv.Atoi(vars)
	if err != nil {
		h.logger.Error("ID de jugador inválido", zap.Error(err))
		http.Error(w, "ID de jugador inválido", http.StatusBadRequest)
		return
	}

	// Obtener categorías
	categories, err := h.rankingRepo.GetRankingCategories(true)
	if err != nil {
		h.logger.Error("Error obteniendo categorías", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	// Obtener posiciones del jugador en cada categoría
	playerRankings := make(map[string]interface{})
	for _, category := range categories {
		if category.Type == "player" {
			entries, err := h.rankingRepo.GetRankingEntries(category.ID, nil, 1000)
			if err != nil {
				h.logger.Error("Error obteniendo entradas", zap.Error(err))
				continue
			}

			// Buscar la posición del jugador
			for _, entry := range entries {
				if entry.EntityType == "player" && entry.EntityID == playerID {
					playerRankings[category.Name] = map[string]interface{}{
						"category": category,
						"position": entry.Position,
						"score":    entry.Score,
						"change":   entry.PositionChange,
					}
					break
				}
			}
		}
	}

	respondWithJSON(w, http.StatusOK, playerRankings)
}

// GetTopPlayers obtiene los mejores jugadores
func (h *RankingHandler) GetTopPlayers(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	categoryStr := query.Get("category")
	limitStr := query.Get("limit")

	limit := 10 // Por defecto
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	// Obtener categoría de jugadores
	categories, err := h.rankingRepo.GetRankingCategories(true)
	if err != nil {
		h.logger.Error("Error obteniendo categorías", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	var targetCategory *models.RankingCategory
	for _, category := range categories {
		if category.Type == "player" && (categoryStr == "" || category.Name == categoryStr) {
			targetCategory = &category
			break
		}
	}

	if targetCategory == nil {
		http.Error(w, "Categoría no encontrada", http.StatusNotFound)
		return
	}

	entries, err := h.rankingRepo.GetRankingEntries(targetCategory.ID, nil, limit)
	if err != nil {
		h.logger.Error("Error obteniendo entradas", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"category": targetCategory,
		"players":  entries,
	})
}

// GetTopAlliances obtiene las mejores alianzas
func (h *RankingHandler) GetTopAlliances(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	categoryStr := query.Get("category")
	limitStr := query.Get("limit")

	limit := 10 // Por defecto
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	// Obtener categoría de alianzas
	categories, err := h.rankingRepo.GetRankingCategories(true)
	if err != nil {
		h.logger.Error("Error obteniendo categorías", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	var targetCategory *models.RankingCategory
	for _, category := range categories {
		if category.Type == "alliance" && (categoryStr == "" || category.Name == categoryStr) {
			targetCategory = &category
			break
		}
	}

	if targetCategory == nil {
		http.Error(w, "Categoría no encontrada", http.StatusNotFound)
		return
	}

	entries, err := h.rankingRepo.GetRankingEntries(targetCategory.ID, nil, limit)
	if err != nil {
		h.logger.Error("Error obteniendo entradas", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"category":  targetCategory,
		"alliances": entries,
	})
}

// GetRankingComparison compara rankings entre entidades
func (h *RankingHandler) GetRankingComparison(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	categoryIDStr := query.Get("category_id")
	entityType := query.Get("entity_type")
	entityIDs := query.Get("entity_ids")

	if categoryIDStr == "" || entityType == "" || entityIDs == "" {
		http.Error(w, "Parámetros requeridos: category_id, entity_type, entity_ids", http.StatusBadRequest)
		return
	}

	categoryID, err := strconv.Atoi(categoryIDStr)
	if err != nil {
		h.logger.Error("ID de categoría inválido", zap.Error(err))
		http.Error(w, "ID de categoría inválido", http.StatusBadRequest)
		return
	}

	// Obtener entradas para comparación
	entries, err := h.rankingRepo.GetRankingEntries(categoryID, nil, 1000)
	if err != nil {
		h.logger.Error("Error obteniendo entradas", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	// Filtrar entradas por entidades solicitadas
	var comparisonEntries []models.RankingEntry
	for _, entry := range entries {
		if entry.EntityType == entityType {
			comparisonEntries = append(comparisonEntries, entry)
		}
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"category": categoryID,
		"entities": comparisonEntries,
	})
}

// Funciones auxiliares

func (h *RankingHandler) validateRankingCategory(category *models.RankingCategory) error {
	if category.Name == "" {
		return fmt.Errorf("nombre de categoría no puede estar vacío")
	}

	if category.Type == "" {
		return fmt.Errorf("tipo de categoría no puede estar vacío")
	}

	if category.UpdateInterval <= 0 {
		return fmt.Errorf("intervalo de actualización debe ser mayor a 0")
	}

	if category.MaxPositions <= 0 {
		return fmt.Errorf("número máximo de posiciones debe ser mayor a 0")
	}

	return nil
}

func (h *RankingHandler) getFeaturedRankings() ([]models.RankingEntry, error) {
	// Obtener rankings destacados (primeros lugares de categorías importantes)
	categories, err := h.rankingRepo.GetRankingCategories(true)
	if err != nil {
		return nil, err
	}

	var featured []models.RankingEntry
	for _, category := range categories {
		if category.ShowInDashboard {
			entries, err := h.rankingRepo.GetRankingEntries(category.ID, nil, 3)
			if err != nil {
				h.logger.Error("Error obteniendo entradas destacadas", zap.Error(err))
				continue
			}
			featured = append(featured, entries...)
		}
	}

	return featured, nil
}
