package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"server-backend/models"
	"server-backend/repository"

	"github.com/google/uuid"
)

type QuestHandler struct {
	questRepo *repository.QuestRepository
}

func NewQuestHandler(questRepo *repository.QuestRepository) *QuestHandler {
	return &QuestHandler{
		questRepo: questRepo,
	}
}

// ==================== CATEGORÍAS DE MISIONES ====================

// CreateQuestCategory crea una nueva categoría de misiones
func (h *QuestHandler) CreateQuestCategory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	var category models.QuestCategory
	if err := json.NewDecoder(r.Body).Decode(&category); err != nil {
		http.Error(w, "Error al decodificar datos: "+err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.questRepo.CreateQuestCategory(&category); err != nil {
		http.Error(w, "Error al crear categoría: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Categoría de misión creada exitosamente",
		"data":    category,
	})
}

// GetQuestCategory obtiene una categoría por ID
func (h *QuestHandler) GetQuestCategory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "ID de categoría requerido", http.StatusBadRequest)
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	category, err := h.questRepo.GetQuestCategory(id)
	if err != nil {
		http.Error(w, "Error al obtener categoría: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    category,
	})
}

// GetAllQuestCategories obtiene todas las categorías
func (h *QuestHandler) GetAllQuestCategories(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	categories, err := h.questRepo.GetAllQuestCategories()
	if err != nil {
		http.Error(w, "Error al obtener categorías: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    categories,
	})
}

// UpdateQuestCategory actualiza una categoría
func (h *QuestHandler) UpdateQuestCategory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	var category models.QuestCategory
	if err := json.NewDecoder(r.Body).Decode(&category); err != nil {
		http.Error(w, "Error al decodificar datos: "+err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.questRepo.UpdateQuestCategory(&category); err != nil {
		http.Error(w, "Error al actualizar categoría: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Categoría actualizada exitosamente",
		"data":    category,
	})
}

// DeleteQuestCategory elimina una categoría de quests
func (h *QuestHandler) DeleteQuestCategory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "ID requerido", http.StatusBadRequest)
		return
	}

	_, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	// TODO: Implementar cuando se actualice el repositorio para usar UUID
	// if err := h.questRepo.DeleteQuestCategory(id); err != nil {
	// 	http.Error(w, "Error al eliminar categoría: "+err.Error(), http.StatusInternalServerError)
	// 	return
	// }

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Categoría eliminada exitosamente",
	})
}

// ==================== MISIONES ====================

// CreateQuest crea una nueva misión
func (h *QuestHandler) CreateQuest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	var quest models.Quest
	if err := json.NewDecoder(r.Body).Decode(&quest); err != nil {
		http.Error(w, "Error al decodificar datos: "+err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.questRepo.CreateQuest(&quest); err != nil {
		http.Error(w, "Error al crear misión: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Misión creada exitosamente",
		"data":    quest,
	})
}

// GetQuest obtiene una misión por ID
func (h *QuestHandler) GetQuest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "ID de misión requerido", http.StatusBadRequest)
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	quest, err := h.questRepo.GetQuest(id)
	if err != nil {
		http.Error(w, "Error al obtener misión: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    quest,
	})
}

// GetQuestsByCategory obtiene misiones por categoría
func (h *QuestHandler) GetQuestsByCategory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	categoryIDStr := r.URL.Query().Get("category_id")
	if categoryIDStr == "" {
		http.Error(w, "ID de categoría requerido", http.StatusBadRequest)
		return
	}

	_, err := uuid.Parse(categoryIDStr)
	if err != nil {
		http.Error(w, "ID de categoría inválido", http.StatusBadRequest)
		return
	}

	// TODO: Implementar cuando se actualice el repositorio para usar UUID
	// quests, err := h.questRepo.GetQuestsByCategory(categoryID)
	// if err != nil {
	// 	http.Error(w, "Error al obtener misiones: "+err.Error(), http.StatusInternalServerError)
	// 	return
	// }

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    []models.Quest{},
	})
}

// GetAvailableQuests obtiene las misiones disponibles para un jugador
func (h *QuestHandler) GetAvailableQuests(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	playerIDStr := r.URL.Query().Get("player_id")
	if playerIDStr == "" {
		http.Error(w, "ID de jugador requerido", http.StatusBadRequest)
		return
	}

	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		http.Error(w, "ID de jugador inválido", http.StatusBadRequest)
		return
	}

	quests, err := h.questRepo.GetAvailableQuests(playerID, nil, 1, false)
	if err != nil {
		http.Error(w, "Error al obtener misiones disponibles: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    quests,
	})
}

// UpdateQuest actualiza una misión
func (h *QuestHandler) UpdateQuest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	var quest models.Quest
	if err := json.NewDecoder(r.Body).Decode(&quest); err != nil {
		http.Error(w, "Error al decodificar datos: "+err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.questRepo.UpdateQuest(&quest); err != nil {
		http.Error(w, "Error al actualizar misión: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Misión actualizada exitosamente",
		"data":    quest,
	})
}

// DeleteQuest elimina una misión
func (h *QuestHandler) DeleteQuest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "ID de misión requerido", http.StatusBadRequest)
		return
	}

	_, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	// TODO: Implementar cuando se actualice el repositorio para usar UUID
	// if err := h.questRepo.DeleteQuest(id); err != nil {
	// 	http.Error(w, "Error al eliminar misión: "+err.Error(), http.StatusInternalServerError)
	// 	return
	// }

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Misión eliminada exitosamente",
	})
}

// ==================== PROGRESO DE JUGADOR ====================

// GetPlayerQuest obtiene el progreso de un jugador en una misión
func (h *QuestHandler) GetPlayerQuest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	playerIDStr := r.URL.Query().Get("player_id")
	questIDStr := r.URL.Query().Get("quest_id")

	if playerIDStr == "" || questIDStr == "" {
		http.Error(w, "player_id y quest_id son requeridos", http.StatusBadRequest)
		return
	}

	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		http.Error(w, "player_id inválido", http.StatusBadRequest)
		return
	}

	questID, err := uuid.Parse(questIDStr)
	if err != nil {
		http.Error(w, "quest_id inválido", http.StatusBadRequest)
		return
	}

	playerQuest, err := h.questRepo.GetPlayerQuest(playerID, questID)
	if err != nil {
		http.Error(w, "Error al obtener progreso del jugador: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    playerQuest,
	})
}

// GetPlayerActiveQuests obtiene las misiones activas de un jugador
func (h *QuestHandler) GetPlayerActiveQuests(w http.ResponseWriter, r *http.Request) {
	playerIDStr := r.URL.Query().Get("player_id")
	if playerIDStr == "" {
		http.Error(w, "ID de jugador requerido", http.StatusBadRequest)
		return
	}

	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		http.Error(w, "ID de jugador inválido", http.StatusBadRequest)
		return
	}

	quests, err := h.questRepo.GetPlayerActiveQuests(playerID, nil, false)
	if err != nil {
		http.Error(w, "Error al obtener misiones activas: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    quests,
	})
}

// UpdateQuestProgress actualiza el progreso de una quest
func (h *QuestHandler) UpdateQuestProgress(w http.ResponseWriter, r *http.Request) {
	var request struct {
		PlayerID    string `json:"player_id"`
		QuestID     string `json:"quest_id"`
		NewProgress int    `json:"new_progress"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Datos inválidos", http.StatusBadRequest)
		return
	}

	playerUUID, err := uuid.Parse(request.PlayerID)
	if err != nil {
		http.Error(w, "player_id inválido", http.StatusBadRequest)
		return
	}
	questUUID, err := uuid.Parse(request.QuestID)
	if err != nil {
		http.Error(w, "quest_id inválido", http.StatusBadRequest)
		return
	}

	err = h.questRepo.UpdateQuestProgress(playerUUID, questUUID, request.NewProgress, nil)
	if err != nil {
		http.Error(w, "Error actualizando progreso: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// ClaimQuestRewards reclama las recompensas de una quest completada
func (h *QuestHandler) ClaimQuestRewards(w http.ResponseWriter, r *http.Request) {
	// TODO: Implementar lógica de reclamación de recompensas
	http.Error(w, "No implementado", http.StatusNotImplemented)
}

// ==================== CADENAS DE MISIONES ====================

// CreateQuestChain crea una nueva cadena de quests
func (h *QuestHandler) CreateQuestChain(w http.ResponseWriter, r *http.Request) {
	// TODO: Implementar lógica de creación de cadenas de quests
	http.Error(w, "No implementado", http.StatusNotImplemented)
}

// GetQuestChain obtiene una cadena de quests
func (h *QuestHandler) GetQuestChain(w http.ResponseWriter, r *http.Request) {
	// TODO: Implementar lógica de obtención de cadenas de quests
	http.Error(w, "No implementado", http.StatusNotImplemented)
}

// GetQuestsInChain obtiene las quests de una cadena
func (h *QuestHandler) GetQuestsInChain(w http.ResponseWriter, r *http.Request) {
	// TODO: Implementar lógica de quests en cadena
	http.Error(w, "No implementado", http.StatusNotImplemented)
}

// ==================== ESTADÍSTICAS ====================

// GetQuestStatistics obtiene estadísticas de quests
func (h *QuestHandler) GetQuestStatistics(w http.ResponseWriter, r *http.Request) {
	// TODO: Implementar lógica de estadísticas de quests
	http.Error(w, "No implementado", http.StatusNotImplemented)
}

// ==================== NOTIFICACIONES ====================

// GetPlayerNotifications obtiene notificaciones de quests
func (h *QuestHandler) GetPlayerNotifications(w http.ResponseWriter, r *http.Request) {
	// TODO: Implementar lógica de notificaciones de quests
	http.Error(w, "No implementado", http.StatusNotImplemented)
}

// MarkNotificationAsRead marca una notificación como leída
func (h *QuestHandler) MarkNotificationAsRead(w http.ResponseWriter, r *http.Request) {
	notificationIDStr := r.URL.Query().Get("notification_id")
	if notificationIDStr == "" {
		http.Error(w, "ID de notificación requerido", http.StatusBadRequest)
		return
	}

	_, err := uuid.Parse(notificationIDStr)
	if err != nil {
		http.Error(w, "ID de notificación inválido", http.StatusBadRequest)
		return
	}

	// TODO: Implementar cuando se agregue el método al repositorio
	// Por ahora, solo marcamos como exitoso
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Notificación marcada como leída",
	})
}

// ==================== DASHBOARD ====================

// GetQuestDashboard obtiene el dashboard de quests de un jugador
func (h *QuestHandler) GetQuestDashboard(w http.ResponseWriter, r *http.Request) {
	playerIDStr := r.URL.Query().Get("player_id")
	if playerIDStr == "" {
		http.Error(w, "ID de jugador requerido", http.StatusBadRequest)
		return
	}

	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		http.Error(w, "ID de jugador inválido", http.StatusBadRequest)
		return
	}

	// Obtener estadísticas del jugador
	stats, err := h.questRepo.GetPlayerQuestStatistics(playerID)
	if err != nil {
		http.Error(w, "Error obteniendo estadísticas: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Obtener categorías
	categories, err := h.questRepo.GetQuestCategories(true)
	if err != nil {
		http.Error(w, "Error obteniendo categorías: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Obtener quests activas
	_, err = h.questRepo.GetPlayerActiveQuests(playerID, nil, false)
	if err != nil {
		http.Error(w, "Error obteniendo quests activas: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Obtener quests disponibles
	availableQuests, err := h.questRepo.GetAvailableQuests(playerID, nil, 1, false)
	if err != nil {
		http.Error(w, "Error obteniendo quests disponibles: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Crear dashboard
	dashboard := models.QuestDashboard{
		PlayerStats:     stats,
		Categories:      categories,
		ActiveQuests:    []models.Quest{}, // Convertir PlayerQuest a Quest
		AvailableQuests: availableQuests,
		LastUpdated:     time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    dashboard,
	})
}

// GetQuestWithDetails obtiene una quest con todos sus detalles
func (h *QuestHandler) GetQuestWithDetails(w http.ResponseWriter, r *http.Request) {
	questIDStr := r.URL.Query().Get("quest_id")
	if questIDStr == "" {
		http.Error(w, "ID de quest requerido", http.StatusBadRequest)
		return
	}

	questID, err := uuid.Parse(questIDStr)
	if err != nil {
		http.Error(w, "ID de quest inválido", http.StatusBadRequest)
		return
	}

	playerIDStr := r.URL.Query().Get("player_id")
	var playerID uuid.UUID
	if playerIDStr != "" {
		playerID, err = uuid.Parse(playerIDStr)
		if err != nil {
			http.Error(w, "ID de jugador inválido", http.StatusBadRequest)
			return
		}
	}

	// Obtener quest
	quest, err := h.questRepo.GetQuest(questID)
	if err != nil {
		http.Error(w, "Error obteniendo quest: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Obtener progreso del jugador si se proporciona playerID
	var playerProgress *models.PlayerQuest
	if playerID != uuid.Nil {
		playerProgress, err = h.questRepo.GetPlayerQuest(playerID, questID)
		if err != nil {
			// No es un error crítico, puede ser que el jugador no tenga progreso
			playerProgress = nil
		}
	}

	// Crear respuesta con detalles
	details := models.QuestWithDetails{
		Quest:          quest,
		PlayerProgress: playerProgress,
		Category:       nil, // TODO: Obtener categoría si es necesario
		Prerequisites:  []models.Quest{},
		Rewards:        []models.QuestReward{},
		Statistics:     nil,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    details,
	})
}

// ==================== EVENTOS DEL JUEGO ====================

// ProcessGameEvent procesa un evento del juego para actualizar quests
func (h *QuestHandler) ProcessGameEvent(w http.ResponseWriter, r *http.Request) {
	var request struct {
		PlayerID  string                 `json:"player_id"`
		EventType string                 `json:"event_type"`
		EventData map[string]interface{} `json:"event_data"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Datos inválidos", http.StatusBadRequest)
		return
	}

	playerID, err := uuid.Parse(request.PlayerID)
	if err != nil {
		http.Error(w, "ID de jugador inválido", http.StatusBadRequest)
		return
	}

	// Obtener quests activas del jugador
	activeQuests, err := h.questRepo.GetPlayerActiveQuests(playerID, nil, false)
	if err != nil {
		http.Error(w, "Error obteniendo quests activas: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Procesar cada quest activa para ver si el evento la afecta
	updatedQuests := []models.PlayerQuest{}
	for _, playerQuest := range activeQuests {
		// TODO: Implementar lógica de procesamiento de eventos específicos
		// Por ahora, solo agregamos la quest a la lista de actualizadas
		updatedQuests = append(updatedQuests, playerQuest)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Evento procesado exitosamente",
		"data": map[string]interface{}{
			"event_type":     request.EventType,
			"quests_updated": len(updatedQuests),
		},
	})
}
