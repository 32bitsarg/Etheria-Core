package handlers

import (
	"encoding/json"
	"net/http"
	"server-backend/models"
	"server-backend/services"
	"strconv"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type TitleHandler struct {
	titleService *services.TitleService
	logger       *zap.Logger
}

func NewTitleHandler(titleService *services.TitleService, logger *zap.Logger) *TitleHandler {
	return &TitleHandler{
		titleService: titleService,
		logger:       logger,
	}
}

// GetTitleDashboard obtiene el dashboard de títulos de un jugador
func (h *TitleHandler) GetTitleDashboard(w http.ResponseWriter, r *http.Request) {
	// Obtener playerID del contexto (asumiendo que está autenticado)
	playerID := r.Context().Value("player_id").(string)
	if playerID == "" {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	dashboard, err := h.titleService.GetTitleDashboard(playerID)
	if err != nil {
		h.logger.Error("Error obteniendo dashboard de títulos", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dashboard)
}

// GetTitleCategories obtiene todas las categorías de títulos
func (h *TitleHandler) GetTitleCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.titleService.GetTitleCategories()
	if err != nil {
		h.logger.Error("Error obteniendo categorías de títulos", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(categories)
}

// GetTitleCategory obtiene una categoría específica
func (h *TitleHandler) GetTitleCategory(w http.ResponseWriter, r *http.Request) {
	categoryID := chi.URLParam(r, "categoryID")
	if categoryID == "" {
		http.Error(w, "ID de categoría requerido", http.StatusBadRequest)
		return
	}

	category, err := h.titleService.GetTitleCategory(categoryID)
	if err != nil {
		h.logger.Error("Error obteniendo categoría", zap.Error(err))
		http.Error(w, "Categoría no encontrada", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(category)
}

// CreateTitleCategory crea una nueva categoría de títulos
func (h *TitleHandler) CreateTitleCategory(w http.ResponseWriter, r *http.Request) {
	var category models.TitleCategory
	if err := json.NewDecoder(r.Body).Decode(&category); err != nil {
		http.Error(w, "Datos inválidos", http.StatusBadRequest)
		return
	}

	if err := h.titleService.CreateTitleCategory(&category); err != nil {
		h.logger.Error("Error creando categoría", zap.Error(err))
		http.Error(w, "Error creando categoría", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(category)
}

// GetTitles obtiene todos los títulos con filtros opcionales
func (h *TitleHandler) GetTitles(w http.ResponseWriter, r *http.Request) {
	// Parsear filtros de query parameters
	filters := make(map[string]interface{})
	if categoryID := r.URL.Query().Get("category_id"); categoryID != "" {
		filters["category_id"] = categoryID
	}
	if rarity := r.URL.Query().Get("rarity"); rarity != "" {
		filters["rarity"] = rarity
	}
	if titleType := r.URL.Query().Get("type"); titleType != "" {
		filters["title_type"] = titleType
	}

	titles, err := h.titleService.GetTitles(filters)
	if err != nil {
		h.logger.Error("Error obteniendo títulos", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(titles)
}

// GetTitle obtiene un título específico
func (h *TitleHandler) GetTitle(w http.ResponseWriter, r *http.Request) {
	titleID := chi.URLParam(r, "titleID")
	if titleID == "" {
		http.Error(w, "ID de título requerido", http.StatusBadRequest)
		return
	}

	title, err := h.titleService.GetTitle(titleID)
	if err != nil {
		h.logger.Error("Error obteniendo título", zap.Error(err))
		http.Error(w, "Título no encontrado", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(title)
}

// CreateTitle crea un nuevo título
func (h *TitleHandler) CreateTitle(w http.ResponseWriter, r *http.Request) {
	var title models.Title
	if err := json.NewDecoder(r.Body).Decode(&title); err != nil {
		http.Error(w, "Datos inválidos", http.StatusBadRequest)
		return
	}

	if err := h.titleService.CreateTitle(&title); err != nil {
		h.logger.Error("Error creando título", zap.Error(err))
		http.Error(w, "Error creando título", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(title)
}

// GetPlayerTitles obtiene los títulos de un jugador
func (h *TitleHandler) GetPlayerTitles(w http.ResponseWriter, r *http.Request) {
	// Obtener playerID del contexto (asumiendo que está autenticado)
	playerID := r.Context().Value("player_id").(string)
	if playerID == "" {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	// Parsear filtros de query parameters
	filters := make(map[string]interface{})
	if categoryID := r.URL.Query().Get("category_id"); categoryID != "" {
		filters["category_id"] = categoryID
	}
	if equippedOnly := r.URL.Query().Get("equipped_only"); equippedOnly != "" {
		equipped, _ := strconv.ParseBool(equippedOnly)
		filters["equipped_only"] = equipped
	}

	titles, err := h.titleService.GetPlayerTitles(playerID, filters)
	if err != nil {
		h.logger.Error("Error obteniendo títulos del jugador", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(titles)
}

// GrantTitle otorga un título a un jugador
func (h *TitleHandler) GrantTitle(w http.ResponseWriter, r *http.Request) {
	var request struct {
		TitleID string `json:"title_id"`
		Reason  string `json:"reason"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Datos inválidos", http.StatusBadRequest)
		return
	}

	// Obtener playerID del contexto (asumiendo que está autenticado)
	playerID := r.Context().Value("player_id").(string)
	if playerID == "" {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	if err := h.titleService.GrantTitle(playerID, request.TitleID, request.Reason); err != nil {
		h.logger.Error("Error otorgando título", zap.Error(err))
		http.Error(w, "Error otorgando título", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Título otorgado exitosamente"})
}

// EquipTitle equipa un título
func (h *TitleHandler) EquipTitle(w http.ResponseWriter, r *http.Request) {
	var request struct {
		TitleID string `json:"title_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Datos inválidos", http.StatusBadRequest)
		return
	}

	// Obtener playerID del contexto (asumiendo que está autenticado)
	playerID := r.Context().Value("player_id").(string)
	if playerID == "" {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	if err := h.titleService.EquipTitle(playerID, request.TitleID); err != nil {
		h.logger.Error("Error equipando título", zap.Error(err))
		http.Error(w, "Error equipando título", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Título equipado exitosamente"})
}

// UnequipTitle desequipa el título actual
func (h *TitleHandler) UnequipTitle(w http.ResponseWriter, r *http.Request) {
	// Obtener playerID del contexto (asumiendo que está autenticado)
	playerID := r.Context().Value("player_id").(string)
	if playerID == "" {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	if err := h.titleService.UnequipTitle(playerID); err != nil {
		h.logger.Error("Error desequipando título", zap.Error(err))
		http.Error(w, "Error desequipando título", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Título desequipado exitosamente"})
}

// GetTitleLeaderboard obtiene el leaderboard de títulos
func (h *TitleHandler) GetTitleLeaderboard(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 10 // valor por defecto
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	leaderboard, err := h.titleService.GetTitleLeaderboard(limit)
	if err != nil {
		h.logger.Error("Error obteniendo leaderboard", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(leaderboard)
}

// GetTitleStatistics obtiene las estadísticas de títulos
func (h *TitleHandler) GetTitleStatistics(w http.ResponseWriter, r *http.Request) {
	stats, err := h.titleService.GetTitleStatistics()
	if err != nil {
		h.logger.Error("Error obteniendo estadísticas", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
