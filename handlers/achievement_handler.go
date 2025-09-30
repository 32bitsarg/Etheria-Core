package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"server-backend/models"
	"server-backend/repository"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type AchievementHandler struct {
	achievementRepo *repository.AchievementRepository
	logger          *zap.Logger
}

func NewAchievementHandler(achievementRepo *repository.AchievementRepository, logger *zap.Logger) *AchievementHandler {
	return &AchievementHandler{
		achievementRepo: achievementRepo,
		logger:          logger,
	}
}

// GetAchievementDashboard obtiene el dashboard completo de logros
func (h *AchievementHandler) GetAchievementDashboard(w http.ResponseWriter, r *http.Request) {
	playerIDStr := r.Context().Value("player_id").(string)
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		http.Error(w, "Player ID inválido", http.StatusBadRequest)
		return
	}

	// Obtener estadísticas del jugador
	stats, err := h.achievementRepo.GetAchievementStatistics(playerID)
	if err != nil {
		h.logger.Error("Error obteniendo estadísticas de logros", zap.Error(err))
		http.Error(w, "Error obteniendo estadísticas", http.StatusInternalServerError)
		return
	}

	// Obtener categorías
	categories, err := h.achievementRepo.GetAchievementCategories(true)
	if err != nil {
		h.logger.Error("Error obteniendo categorías", zap.Error(err))
		http.Error(w, "Error obteniendo categorías", http.StatusInternalServerError)
		return
	}

	// Obtener logros recientes
	_, err = h.achievementRepo.GetPlayerAchievements(playerID, nil, true)
	if err != nil {
		h.logger.Error("Error obteniendo logros recientes", zap.Error(err))
		http.Error(w, "Error obteniendo logros recientes", http.StatusInternalServerError)
		return
	}

	// Obtener logros próximos (no completados)
	_, err = h.achievementRepo.GetPlayerAchievements(playerID, nil, false)
	if err != nil {
		h.logger.Error("Error obteniendo logros próximos", zap.Error(err))
		http.Error(w, "Error obteniendo logros próximos", http.StatusInternalServerError)
		return
	}

	// Obtener ranking
	leaderboard, err := h.achievementRepo.GetAchievementLeaderboard(10, nil)
	if err != nil {
		h.logger.Error("Error obteniendo ranking", zap.Error(err))
		http.Error(w, "Error obteniendo ranking", http.StatusInternalServerError)
		return
	}

	// Obtener notificaciones
	notifications, err := h.achievementRepo.GetAchievementNotifications(playerID, true)
	if err != nil {
		h.logger.Error("Error obteniendo notificaciones", zap.Error(err))
		http.Error(w, "Error obteniendo notificaciones", http.StatusInternalServerError)
		return
	}

	// Crear dashboard
	dashboard := models.AchievementDashboard{
		PlayerStats:          stats,
		Categories:           categories,
		RecentAchievements:   []models.Achievement{}, // TODO: Convertir PlayerAchievement a Achievement
		UpcomingAchievements: []models.Achievement{}, // TODO: Convertir PlayerAchievement a Achievement
		Leaderboard:          leaderboard,
		Notifications:        notifications,
		GlobalStats:          make(map[string]interface{}),
		LastUpdated:          time.Now(),
	}

	// Agregar estadísticas globales
	dashboard.GlobalStats["total_players"] = len(leaderboard)
	dashboard.GlobalStats["average_completion_rate"] = 0.0
	if len(leaderboard) > 0 {
		totalRate := 0.0
		for _, entry := range leaderboard {
			totalRate += entry.CompletionRate
		}
		dashboard.GlobalStats["average_completion_rate"] = totalRate / float64(len(leaderboard))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dashboard)
}

// GetAchievementCategories obtiene todas las categorías de logros
func (h *AchievementHandler) GetAchievementCategories(w http.ResponseWriter, r *http.Request) {
	activeOnly := r.URL.Query().Get("active_only") == "true"

	categories, err := h.achievementRepo.GetAchievementCategories(activeOnly)
	if err != nil {
		h.logger.Error("Error obteniendo categorías", zap.Error(err))
		http.Error(w, "Error obteniendo categorías", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(categories)
}

// GetAchievementCategory obtiene una categoría específica
func (h *AchievementHandler) GetAchievementCategory(w http.ResponseWriter, r *http.Request) {
	vars := chi.URLParam(r, "id")
	categoryID, err := uuid.Parse(vars)
	if err != nil {
		http.Error(w, "ID de categoría inválido", http.StatusBadRequest)
		return
	}

	category, err := h.achievementRepo.GetAchievementCategory(categoryID)
	if err != nil {
		h.logger.Error("Error obteniendo categoría", zap.Error(err))
		http.Error(w, "Categoría no encontrada", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(category)
}

// CreateAchievementCategory crea una nueva categoría
func (h *AchievementHandler) CreateAchievementCategory(w http.ResponseWriter, r *http.Request) {
	var category models.AchievementCategory
	if err := json.NewDecoder(r.Body).Decode(&category); err != nil {
		http.Error(w, "Datos inválidos", http.StatusBadRequest)
		return
	}

	err := h.achievementRepo.CreateAchievementCategory(&category)
	if err != nil {
		h.logger.Error("Error creando categoría", zap.Error(err))
		http.Error(w, "Error creando categoría", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(category)
}

// GetAchievements obtiene logros con filtros
func (h *AchievementHandler) GetAchievements(w http.ResponseWriter, r *http.Request) {
	categoryIDStr := r.URL.Query().Get("category_id")
	activeOnly := r.URL.Query().Get("active_only") == "true"
	includeHidden := r.URL.Query().Get("include_hidden") == "true"

	var categoryID *uuid.UUID
	if categoryIDStr != "" {
		if id, err := uuid.Parse(categoryIDStr); err == nil {
			categoryID = &id
		}
	}

	achievements, err := h.achievementRepo.GetAchievements(categoryID, activeOnly, includeHidden)
	if err != nil {
		h.logger.Error("Error obteniendo logros", zap.Error(err))
		http.Error(w, "Error obteniendo logros", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(achievements)
}

// GetAchievement obtiene un logro específico
func (h *AchievementHandler) GetAchievement(w http.ResponseWriter, r *http.Request) {
	vars := chi.URLParam(r, "id")
	achievementID, err := uuid.Parse(vars)
	if err != nil {
		http.Error(w, "ID de logro inválido", http.StatusBadRequest)
		return
	}

	achievement, err := h.achievementRepo.GetAchievement(achievementID)
	if err != nil {
		h.logger.Error("Error obteniendo logro", zap.Error(err))
		http.Error(w, "Logro no encontrado", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(achievement)
}

// GetPlayerAchievements obtiene los logros de un jugador
func (h *AchievementHandler) GetPlayerAchievements(w http.ResponseWriter, r *http.Request) {
	playerIDStr := r.Context().Value("player_id").(string)
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		http.Error(w, "Player ID inválido", http.StatusBadRequest)
		return
	}

	categoryIDStr := r.URL.Query().Get("category_id")
	completedOnly := r.URL.Query().Get("completed_only") == "true"

	var categoryID *uuid.UUID
	if categoryIDStr != "" {
		if id, err := uuid.Parse(categoryIDStr); err == nil {
			categoryID = &id
		}
	}

	achievements, err := h.achievementRepo.GetPlayerAchievements(playerID, categoryID, completedOnly)
	if err != nil {
		h.logger.Error("Error obteniendo logros del jugador", zap.Error(err))
		http.Error(w, "Error obteniendo logros", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(achievements)
}

// GetPlayerAchievement obtiene el progreso de un jugador en un logro específico
func (h *AchievementHandler) GetPlayerAchievement(w http.ResponseWriter, r *http.Request) {
	playerIDStr := r.Context().Value("player_id").(string)
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		http.Error(w, "Player ID inválido", http.StatusBadRequest)
		return
	}

	vars := chi.URLParam(r, "id")
	achievementID, err := uuid.Parse(vars)
	if err != nil {
		http.Error(w, "ID de logro inválido", http.StatusBadRequest)
		return
	}

	achievement, err := h.achievementRepo.GetPlayerAchievement(playerID, achievementID)
	if err != nil {
		h.logger.Error("Error obteniendo logro del jugador", zap.Error(err))
		http.Error(w, "Logro no encontrado", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(achievement)
}

// UpdateAchievementProgress actualiza el progreso de un logro
func (h *AchievementHandler) UpdateAchievementProgress(w http.ResponseWriter, r *http.Request) {
	playerIDStr := r.Context().Value("player_id").(string)
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		http.Error(w, "Player ID inválido", http.StatusBadRequest)
		return
	}

	vars := chi.URLParam(r, "id")
	achievementID, err := uuid.Parse(vars)
	if err != nil {
		http.Error(w, "ID de logro inválido", http.StatusBadRequest)
		return
	}

	var request struct {
		Progress int `json:"progress"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Datos inválidos", http.StatusBadRequest)
		return
	}

	err = h.achievementRepo.UpdateAchievementProgress(playerID, achievementID, request.Progress)
	if err != nil {
		h.logger.Error("Error actualizando progreso", zap.Error(err))
		http.Error(w, "Error actualizando progreso", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GetAchievementStatistics obtiene las estadísticas de logros de un jugador
func (h *AchievementHandler) GetAchievementStatistics(w http.ResponseWriter, r *http.Request) {
	playerIDStr := r.Context().Value("player_id").(string)
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		http.Error(w, "Player ID inválido", http.StatusBadRequest)
		return
	}

	stats, err := h.achievementRepo.GetAchievementStatistics(playerID)
	if err != nil {
		h.logger.Error("Error obteniendo estadísticas", zap.Error(err))
		http.Error(w, "Error obteniendo estadísticas", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// GetAchievementLeaderboard obtiene el ranking de logros
func (h *AchievementHandler) GetAchievementLeaderboard(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 10
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	worldIDStr := r.URL.Query().Get("world_id")

	var worldID *uuid.UUID
	if worldIDStr != "" {
		if id, err := uuid.Parse(worldIDStr); err == nil {
			worldID = &id
		}
	}

	leaderboard, err := h.achievementRepo.GetAchievementLeaderboard(limit, worldID)
	if err != nil {
		h.logger.Error("Error obteniendo ranking", zap.Error(err))
		http.Error(w, "Error obteniendo ranking", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(leaderboard)
}

// GetAchievementNotifications obtiene las notificaciones de logros
func (h *AchievementHandler) GetAchievementNotifications(w http.ResponseWriter, r *http.Request) {
	playerIDStr := r.Context().Value("player_id").(string)
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		http.Error(w, "Player ID inválido", http.StatusBadRequest)
		return
	}

	unreadOnly := r.URL.Query().Get("unread_only") == "true"

	notifications, err := h.achievementRepo.GetAchievementNotifications(playerID, unreadOnly)
	if err != nil {
		h.logger.Error("Error obteniendo notificaciones", zap.Error(err))
		http.Error(w, "Error obteniendo notificaciones", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notifications)
}

// MarkNotificationAsRead marca una notificación como leída
func (h *AchievementHandler) MarkNotificationAsRead(w http.ResponseWriter, r *http.Request) {
	playerIDStr := r.Context().Value("player_id").(string)
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		http.Error(w, "Player ID inválido", http.StatusBadRequest)
		return
	}

	vars := chi.URLParam(r, "id")
	notificationID, err := uuid.Parse(vars)
	if err != nil {
		http.Error(w, "ID de notificación inválido", http.StatusBadRequest)
		return
	}

	err = h.achievementRepo.MarkNotificationAsRead(playerID, notificationID)
	if err != nil {
		h.logger.Error("Error marcando notificación como leída", zap.Error(err))
		http.Error(w, "Error marcando notificación", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// ClaimAchievementReward reclama la recompensa de un logro
func (h *AchievementHandler) ClaimAchievementReward(w http.ResponseWriter, r *http.Request) {
	playerIDStr := r.Context().Value("player_id").(string)
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		http.Error(w, "Player ID inválido", http.StatusBadRequest)
		return
	}

	vars := chi.URLParam(r, "id")
	achievementID, err := uuid.Parse(vars)
	if err != nil {
		http.Error(w, "ID de logro inválido", http.StatusBadRequest)
		return
	}

	// Verificar que el jugador tiene el logro completado
	playerAchievement, err := h.achievementRepo.GetPlayerAchievement(playerID, achievementID)
	if err != nil {
		h.logger.Error("Error obteniendo logro del jugador", zap.Error(err))
		http.Error(w, "Logro no encontrado", http.StatusNotFound)
		return
	}

	if !playerAchievement.IsCompleted {
		http.Error(w, "El logro no está completado", http.StatusBadRequest)
		return
	}

	if playerAchievement.RewardsClaimed {
		http.Error(w, "Las recompensas ya han sido reclamadas", http.StatusBadRequest)
		return
	}

	// TODO: Implementar reclamación de recompensas
	w.WriteHeader(http.StatusOK)
}

// GetAchievementWithDetails obtiene un logro con todos sus detalles
func (h *AchievementHandler) GetAchievementWithDetails(w http.ResponseWriter, r *http.Request) {
	playerIDStr := r.Context().Value("player_id").(string)
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		http.Error(w, "Player ID inválido", http.StatusBadRequest)
		return
	}

	vars := chi.URLParam(r, "id")
	achievementID, err := uuid.Parse(vars)
	if err != nil {
		http.Error(w, "ID de logro inválido", http.StatusBadRequest)
		return
	}

	// Obtener logro
	achievement, err := h.achievementRepo.GetAchievement(achievementID)
	if err != nil {
		h.logger.Error("Error obteniendo logro", zap.Error(err))
		http.Error(w, "Logro no encontrado", http.StatusNotFound)
		return
	}

	// Obtener progreso del jugador
	playerAchievement, err := h.achievementRepo.GetPlayerAchievement(playerID, achievementID)
	if err != nil {
		h.logger.Error("Error obteniendo progreso del jugador", zap.Error(err))
		// No es un error crítico, puede ser que el jugador no tenga progreso
	}

	// TODO: Implementar obtención de detalles completos
	details := models.AchievementWithDetails{
		Achievement:    achievement,
		PlayerProgress: playerAchievement,
		Category:       nil,
		Prerequisites:  []models.Achievement{},
		Rewards:        []models.AchievementReward{},
		Statistics:     nil,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(details)
}

// CalculateAchievementProgress calcula el progreso actual de un logro
func (h *AchievementHandler) CalculateAchievementProgress(w http.ResponseWriter, r *http.Request) {
	playerIDStr := r.Context().Value("player_id").(string)
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		http.Error(w, "Player ID inválido", http.StatusBadRequest)
		return
	}

	vars := chi.URLParam(r, "id")
	achievementID, err := uuid.Parse(vars)
	if err != nil {
		http.Error(w, "ID de logro inválido", http.StatusBadRequest)
		return
	}

	// Obtener progreso actual
	playerAchievement, err := h.achievementRepo.GetPlayerAchievement(playerID, achievementID)
	if err != nil {
		h.logger.Error("Error obteniendo progreso", zap.Error(err))
		http.Error(w, "Error obteniendo progreso", http.StatusInternalServerError)
		return
	}

	// TODO: Implementar cálculo de progreso
	progress := map[string]interface{}{
		"current_progress": playerAchievement.CurrentProgress,
		"target_progress":  playerAchievement.TargetProgress,
		"percentage":       float64(playerAchievement.CurrentProgress) / float64(playerAchievement.TargetProgress) * 100,
		"is_completed":     playerAchievement.IsCompleted,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(progress)
}
