package handlers

import (
	"encoding/json"
	"net/http"

	"server-backend/models"
	"server-backend/repository"

	"github.com/google/uuid"
)

type EventHandler struct {
	eventRepo *repository.EventRepository
}

func NewEventHandler(eventRepo *repository.EventRepository) *EventHandler {
	return &EventHandler{
		eventRepo: eventRepo,
	}
}

// ==================== CATEGORÍAS DE EVENTOS ====================

// CreateEventCategory crea una nueva categoría de eventos
func (h *EventHandler) CreateEventCategory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	var category models.EventCategory
	if err := json.NewDecoder(r.Body).Decode(&category); err != nil {
		http.Error(w, "Error al decodificar datos: "+err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.eventRepo.CreateEventCategory(&category); err != nil {
		http.Error(w, "Error al crear categoría: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Categoría de evento creada exitosamente",
		"data":    category,
	})
}

// GetEventCategory obtiene una categoría por ID
func (h *EventHandler) GetEventCategory(w http.ResponseWriter, r *http.Request) {
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

	category, err := h.eventRepo.GetEventCategory(id)
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

// GetAllEventCategories obtiene todas las categorías
func (h *EventHandler) GetAllEventCategories(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	categories, err := h.eventRepo.GetAllEventCategories()
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

// UpdateEventCategory actualiza una categoría
func (h *EventHandler) UpdateEventCategory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	var category models.EventCategory
	if err := json.NewDecoder(r.Body).Decode(&category); err != nil {
		http.Error(w, "Error al decodificar datos: "+err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.eventRepo.UpdateEventCategory(&category); err != nil {
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

// DeleteEventCategory elimina una categoría
func (h *EventHandler) DeleteEventCategory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
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

	if err := h.eventRepo.DeleteEventCategory(id); err != nil {
		http.Error(w, "Error al eliminar categoría: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Categoría eliminada exitosamente",
	})
}

// ==================== EVENTOS ====================

// CreateEvent crea un nuevo evento
func (h *EventHandler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	var event models.Event
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		http.Error(w, "Error al decodificar datos: "+err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.eventRepo.CreateEvent(&event); err != nil {
		http.Error(w, "Error al crear evento: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Evento creado exitosamente",
		"data":    event,
	})
}

// GetEvent obtiene un evento por ID
func (h *EventHandler) GetEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "ID de evento requerido", http.StatusBadRequest)
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	event, err := h.eventRepo.GetEventByID(id)
	if err != nil {
		http.Error(w, "Error al obtener evento: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    event,
	})
}

// GetEventsByCategory obtiene eventos por categoría
func (h *EventHandler) GetEventsByCategory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	categoryIDStr := r.URL.Query().Get("category_id")
	if categoryIDStr == "" {
		http.Error(w, "ID de categoría requerido", http.StatusBadRequest)
		return
	}

	categoryID, err := uuid.Parse(categoryIDStr)
	if err != nil {
		http.Error(w, "ID de categoría inválido", http.StatusBadRequest)
		return
	}

	activeOnly := r.URL.Query().Get("active_only") == "true"
	includePast := r.URL.Query().Get("include_past") == "true"

	events, err := h.eventRepo.GetEventsByCategory(categoryID, activeOnly, includePast)
	if err != nil {
		http.Error(w, "Error al obtener eventos: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    events,
	})
}

// GetActiveEvents obtiene eventos activos
func (h *EventHandler) GetActiveEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	events, err := h.eventRepo.GetActiveEvents()
	if err != nil {
		http.Error(w, "Error al obtener eventos activos: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    events,
	})
}

// GetUpcomingEvents obtiene eventos próximos
func (h *EventHandler) GetUpcomingEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	events, err := h.eventRepo.GetUpcomingEvents()
	if err != nil {
		http.Error(w, "Error al obtener eventos próximos: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    events,
	})
}

// UpdateEvent actualiza un evento
func (h *EventHandler) UpdateEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	var event models.Event
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		http.Error(w, "Error al decodificar datos: "+err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.eventRepo.UpdateEvent(&event); err != nil {
		http.Error(w, "Error al actualizar evento: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Evento actualizado exitosamente",
		"data":    event,
	})
}

// DeleteEvent elimina un evento
func (h *EventHandler) DeleteEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "ID de evento requerido", http.StatusBadRequest)
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	if err := h.eventRepo.DeleteEvent(id); err != nil {
		http.Error(w, "Error al eliminar evento: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Evento eliminado exitosamente",
	})
}

// ==================== PARTICIPANTES ====================

// GetEventParticipant obtiene la participación de un jugador en un evento
func (h *EventHandler) GetEventParticipant(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	eventIDStr := r.URL.Query().Get("event_id")
	playerIDStr := r.URL.Query().Get("player_id")

	if eventIDStr == "" || playerIDStr == "" {
		http.Error(w, "event_id y player_id son requeridos", http.StatusBadRequest)
		return
	}

	eventID, err := uuid.Parse(eventIDStr)
	if err != nil {
		http.Error(w, "event_id inválido", http.StatusBadRequest)
		return
	}

	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		http.Error(w, "player_id inválido", http.StatusBadRequest)
		return
	}

	participant, err := h.eventRepo.GetEventParticipant(eventID, playerID)
	if err != nil {
		http.Error(w, "Error al obtener participación: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    participant,
	})
}

// GetEventParticipants obtiene todos los participantes de un evento
func (h *EventHandler) GetEventParticipants(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	eventIDStr := r.URL.Query().Get("event_id")
	if eventIDStr == "" {
		http.Error(w, "event_id es requerido", http.StatusBadRequest)
		return
	}

	eventID, err := uuid.Parse(eventIDStr)
	if err != nil {
		http.Error(w, "event_id inválido", http.StatusBadRequest)
		return
	}

	participants, err := h.eventRepo.GetEventParticipants(eventID)
	if err != nil {
		http.Error(w, "Error al obtener participantes: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    participants,
	})
}

// RegisterPlayerForEvent registra un jugador en un evento
func (h *EventHandler) RegisterPlayerForEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		EventID      uuid.UUID `json:"event_id"`
		PlayerID     uuid.UUID `json:"player_id"`
		EntryFeePaid bool      `json:"entry_fee_paid"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Error al decodificar datos: "+err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.eventRepo.RegisterPlayerForEvent(request.EventID, request.PlayerID, request.EntryFeePaid); err != nil {
		http.Error(w, "Error al registrar jugador: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Jugador registrado exitosamente en el evento",
	})
}

// UpdateEventProgress actualiza el progreso de un participante
func (h *EventHandler) UpdateEventProgress(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		EventID  uuid.UUID `json:"event_id"`
		PlayerID uuid.UUID `json:"player_id"`
		NewScore int       `json:"new_score"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Error al decodificar datos: "+err.Error(), http.StatusBadRequest)
		return
	}

	update, err := h.eventRepo.UpdateEventProgress(request.EventID, request.PlayerID, request.NewScore)
	if err != nil {
		http.Error(w, "Error al actualizar progreso: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Progreso actualizado exitosamente",
		"data":    update,
	})
}

// ==================== PARTIDAS ====================

// CreateEventMatch crea una nueva partida
func (h *EventHandler) CreateEventMatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	var match models.EventMatch
	if err := json.NewDecoder(r.Body).Decode(&match); err != nil {
		http.Error(w, "Error al decodificar datos: "+err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.eventRepo.CreateEventMatch(&match); err != nil {
		http.Error(w, "Error al crear partida: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Partida creada exitosamente",
		"data":    match,
	})
}

// GetEventMatches obtiene las partidas de un evento
func (h *EventHandler) GetEventMatches(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	eventIDStr := r.URL.Query().Get("event_id")
	if eventIDStr == "" {
		http.Error(w, "event_id es requerido", http.StatusBadRequest)
		return
	}

	eventID, err := uuid.Parse(eventIDStr)
	if err != nil {
		http.Error(w, "event_id inválido", http.StatusBadRequest)
		return
	}

	matches, err := h.eventRepo.GetEventMatches(eventID)
	if err != nil {
		http.Error(w, "Error al obtener partidas: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    matches,
	})
}

// ==================== RANKING ====================

// GetEventLeaderboard obtiene el ranking de un evento
func (h *EventHandler) GetEventLeaderboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	eventIDStr := r.URL.Query().Get("event_id")
	if eventIDStr == "" {
		http.Error(w, "event_id es requerido", http.StatusBadRequest)
		return
	}

	eventID, err := uuid.Parse(eventIDStr)
	if err != nil {
		http.Error(w, "event_id inválido", http.StatusBadRequest)
		return
	}

	leaderboard, err := h.eventRepo.GetEventLeaderboard(eventID)
	if err != nil {
		http.Error(w, "Error al obtener ranking: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    leaderboard,
	})
}

// ==================== ESTADÍSTICAS ====================

// GetEventStatistics obtiene las estadísticas de eventos de un jugador
func (h *EventHandler) GetEventStatistics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	playerIDStr := r.URL.Query().Get("player_id")
	if playerIDStr == "" {
		http.Error(w, "player_id es requerido", http.StatusBadRequest)
		return
	}

	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		http.Error(w, "player_id inválido", http.StatusBadRequest)
		return
	}

	stats, err := h.eventRepo.GetEventStatistics(playerID)
	if err != nil {
		http.Error(w, "Error al obtener estadísticas: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    stats,
	})
}

// ==================== NOTIFICACIONES ====================

// GetPlayerEventNotifications obtiene las notificaciones de eventos de un jugador
func (h *EventHandler) GetPlayerEventNotifications(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	playerIDStr := r.URL.Query().Get("player_id")
	if playerIDStr == "" {
		http.Error(w, "player_id es requerido", http.StatusBadRequest)
		return
	}

	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		http.Error(w, "player_id inválido", http.StatusBadRequest)
		return
	}

	notifications, err := h.eventRepo.GetPlayerEventNotifications(playerID)
	if err != nil {
		http.Error(w, "Error al obtener notificaciones: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    notifications,
	})
}

// ==================== DASHBOARD ====================

// GetEventDashboard obtiene el dashboard completo de eventos
func (h *EventHandler) GetEventDashboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	playerIDStr := r.URL.Query().Get("player_id")
	if playerIDStr == "" {
		http.Error(w, "player_id es requerido", http.StatusBadRequest)
		return
	}

	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		http.Error(w, "player_id inválido", http.StatusBadRequest)
		return
	}

	dashboard, err := h.eventRepo.GetEventDashboard(playerID)
	if err != nil {
		http.Error(w, "Error al obtener dashboard: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    dashboard,
	})
}

// GetEventWithDetails obtiene un evento con todos sus detalles
func (h *EventHandler) GetEventWithDetails(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	eventIDStr := r.URL.Query().Get("event_id")
	playerIDStr := r.URL.Query().Get("player_id")

	if eventIDStr == "" || playerIDStr == "" {
		http.Error(w, "event_id y player_id son requeridos", http.StatusBadRequest)
		return
	}

	eventID, err := uuid.Parse(eventIDStr)
	if err != nil {
		http.Error(w, "event_id inválido", http.StatusBadRequest)
		return
	}

	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		http.Error(w, "player_id inválido", http.StatusBadRequest)
		return
	}

	details, err := h.eventRepo.GetEventWithDetails(eventID, playerID)
	if err != nil {
		http.Error(w, "Error al obtener detalles del evento: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    details,
	})
}
