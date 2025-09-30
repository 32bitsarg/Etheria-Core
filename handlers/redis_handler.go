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
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type RedisHandler struct {
	redisService        *services.RedisService
	constructionService *services.ConstructionService
	chatService         *services.ChatService
	rankingService      *services.RankingService
	eventService        *services.EventService
	battleService       *services.BattleService
	rateLimitService    *services.RateLimitService
	inventoryService    *services.InventoryService
	configCacheService  *services.ConfigCacheService
	villageRepo         *repository.VillageRepository
	logger              *zap.Logger
}

func NewRedisHandler(
	redisService *services.RedisService,
	constructionService *services.ConstructionService,
	chatService *services.ChatService,
	rankingService *services.RankingService,
	eventService *services.EventService,
	battleService *services.BattleService,
	rateLimitService *services.RateLimitService,
	inventoryService *services.InventoryService,
	configCacheService *services.ConfigCacheService,
	villageRepo *repository.VillageRepository,
	logger *zap.Logger,
) *RedisHandler {
	return &RedisHandler{
		redisService:        redisService,
		constructionService: constructionService,
		chatService:         chatService,
		rankingService:      rankingService,
		eventService:        eventService,
		battleService:       battleService,
		rateLimitService:    rateLimitService,
		inventoryService:    inventoryService,
		configCacheService:  configCacheService,
		villageRepo:         villageRepo,
		logger:              logger,
	}
}

// ========================================
// ENDPOINTS BÁSICOS DE REDIS
// ========================================

// GetRedisStats obtiene estadísticas de Redis
func (h *RedisHandler) GetRedisStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.redisService.GetStats()
	if err != nil {
		h.logger.Error("Error obteniendo estadísticas de Redis", zap.Error(err))
		http.Error(w, "Error obteniendo estadísticas", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    stats,
	})
}

// GetOnlineUsers obtiene la lista de usuarios online
func (h *RedisHandler) GetOnlineUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.redisService.GetOnlineUsers()
	if err != nil {
		h.logger.Error("Error obteniendo usuarios online", zap.Error(err))
		http.Error(w, "Error obteniendo usuarios online", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"online_users": users,
			"count":        len(users),
		},
	})
}

// GetUserSession obtiene la sesión de un usuario
func (h *RedisHandler) GetUserSession(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")
	if userID == "" {
		http.Error(w, "UserID requerido", http.StatusBadRequest)
		return
	}

	session, err := h.redisService.GetUserSession(userID)
	if err != nil {
		h.logger.Error("Error obteniendo sesión de usuario", zap.Error(err), zap.String("user_id", userID))
		http.Error(w, "Error obteniendo sesión", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    session,
	})
}

// DeleteUserSession elimina la sesión de un usuario
func (h *RedisHandler) DeleteUserSession(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")
	if userID == "" {
		http.Error(w, "UserID requerido", http.StatusBadRequest)
		return
	}

	err := h.redisService.DeleteUserSession(userID)
	if err != nil {
		h.logger.Error("Error eliminando sesión de usuario", zap.Error(err), zap.String("user_id", userID))
		http.Error(w, "Error eliminando sesión", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Sesión eliminada exitosamente",
	})
}

// GetPlayerResources obtiene los recursos de un jugador
func (h *RedisHandler) GetPlayerResources(w http.ResponseWriter, r *http.Request) {
	playerID := chi.URLParam(r, "playerID")
	if playerID == "" {
		http.Error(w, "PlayerID requerido", http.StatusBadRequest)
		return
	}

	resources, err := h.redisService.GetPlayerResources(playerID)
	if err != nil {
		h.logger.Error("Error obteniendo recursos del jugador", zap.Error(err), zap.String("player_id", playerID))
		http.Error(w, "Error obteniendo recursos", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    resources,
	})
}

// GetResearchProgress obtiene el progreso de investigación de un jugador
func (h *RedisHandler) GetResearchProgress(w http.ResponseWriter, r *http.Request) {
	playerID := chi.URLParam(r, "playerID")
	if playerID == "" {
		http.Error(w, "PlayerID requerido", http.StatusBadRequest)
		return
	}

	research, err := h.redisService.GetResearchProgress(playerID)
	if err != nil {
		h.logger.Error("Error obteniendo progreso de investigación", zap.Error(err), zap.String("player_id", playerID))
		http.Error(w, "Error obteniendo progreso de investigación", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    research,
	})
}

// GetNotifications obtiene las notificaciones de un usuario
func (h *RedisHandler) GetNotifications(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")
	if userID == "" {
		http.Error(w, "UserID requerido", http.StatusBadRequest)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 20 // default
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	notifications, err := h.redisService.GetNotifications(userID, limit)
	if err != nil {
		h.logger.Error("Error obteniendo notificaciones", zap.Error(err), zap.String("user_id", userID))
		http.Error(w, "Error obteniendo notificaciones", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"notifications": notifications,
			"count":         len(notifications),
		},
	})
}

// MarkNotificationAsRead marca una notificación como leída
func (h *RedisHandler) MarkNotificationAsRead(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")
	notificationID := chi.URLParam(r, "notificationID")

	if userID == "" || notificationID == "" {
		http.Error(w, "UserID y NotificationID requeridos", http.StatusBadRequest)
		return
	}

	err := h.redisService.MarkNotificationAsRead(userID, notificationID)
	if err != nil {
		h.logger.Error("Error marcando notificación como leída",
			zap.Error(err),
			zap.String("user_id", userID),
			zap.String("notification_id", notificationID))
		http.Error(w, "Error marcando notificación como leída", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Notificación marcada como leída",
	})
}

// AddNotification agrega una nueva notificación
func (h *RedisHandler) AddNotification(w http.ResponseWriter, r *http.Request) {
	var notification struct {
		UserID  string `json:"user_id"`
		Type    string `json:"type"`
		Title   string `json:"title"`
		Message string `json:"message"`
	}

	if err := json.NewDecoder(r.Body).Decode(&notification); err != nil {
		http.Error(w, "Error decodificando notificación", http.StatusBadRequest)
		return
	}

	if notification.UserID == "" || notification.Title == "" || notification.Message == "" {
		http.Error(w, "UserID, Title y Message son requeridos", http.StatusBadRequest)
		return
	}

	// Crear notificación
	notif := &models.Notification{
		ID:        notification.Type + "_" + strconv.FormatInt(time.Now().UnixNano(), 10),
		Type:      notification.Type,
		Title:     notification.Title,
		Message:   notification.Message,
		CreatedAt: time.Now(),
		IsRead:    false,
	}

	err := h.redisService.AddNotification(notification.UserID, notif)
	if err != nil {
		h.logger.Error("Error agregando notificación", zap.Error(err))
		http.Error(w, "Error agregando notificación", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Notificación agregada exitosamente",
		"data":    notif,
	})
}

// PingRedis verifica la conectividad de Redis
func (h *RedisHandler) PingRedis(w http.ResponseWriter, r *http.Request) {
	err := h.redisService.Ping()
	if err != nil {
		h.logger.Error("Error haciendo ping a Redis", zap.Error(err))
		http.Error(w, "Error conectando con Redis", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"message":   "Redis conectado",
		"timestamp": time.Now(),
	})
}

// FlushRedis limpia toda la base de datos de Redis
func (h *RedisHandler) FlushRedis(w http.ResponseWriter, r *http.Request) {
	// Verificar que sea una solicitud POST
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	err := h.redisService.FlushAll()
	if err != nil {
		h.logger.Error("Error limpiando Redis", zap.Error(err))
		http.Error(w, "Error limpiando Redis", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Redis limpiado exitosamente",
	})
}

// GetCache obtiene un valor del cache
func (h *RedisHandler) GetCache(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "Key requerida", http.StatusBadRequest)
		return
	}

	var data interface{}
	err := h.redisService.GetCache(key, &data)
	if err != nil {
		h.logger.Error("Error obteniendo cache", zap.Error(err), zap.String("key", key))
		http.Error(w, "Error obteniendo cache", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    data,
	})
}

// SetCache establece un valor en el cache
func (h *RedisHandler) SetCache(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Key        string      `json:"key"`
		Data       interface{} `json:"data"`
		Expiration int         `json:"expiration"` // en segundos
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Error decodificando request", http.StatusBadRequest)
		return
	}

	if request.Key == "" {
		http.Error(w, "Key requerida", http.StatusBadRequest)
		return
	}

	expiration := time.Duration(request.Expiration) * time.Second
	if expiration <= 0 {
		expiration = 1 * time.Hour // default
	}

	err := h.redisService.SetCache(request.Key, request.Data, expiration)
	if err != nil {
		h.logger.Error("Error estableciendo cache", zap.Error(err), zap.String("key", request.Key))
		http.Error(w, "Error estableciendo cache", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Cache establecido exitosamente",
	})
}

// DeleteCache elimina un valor del cache
func (h *RedisHandler) DeleteCache(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "Key requerida", http.StatusBadRequest)
		return
	}

	err := h.redisService.DeleteCache(key)
	if err != nil {
		h.logger.Error("Error eliminando cache", zap.Error(err), zap.String("key", key))
		http.Error(w, "Error eliminando cache", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Cache eliminado exitosamente",
	})
}

// GetCounter obtiene el valor de un contador
func (h *RedisHandler) GetCounter(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "Key requerida", http.StatusBadRequest)
		return
	}

	value, err := h.redisService.GetCounter(key)
	if err != nil {
		h.logger.Error("Error obteniendo contador", zap.Error(err), zap.String("key", key))
		http.Error(w, "Error obteniendo contador", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"key":   key,
			"value": value,
		},
	})
}

// IncrementCounter incrementa un contador
func (h *RedisHandler) IncrementCounter(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "Key requerida", http.StatusBadRequest)
		return
	}

	err := h.redisService.IncrementCounter(key)
	if err != nil {
		h.logger.Error("Error incrementando contador", zap.Error(err), zap.String("key", key))
		http.Error(w, "Error incrementando contador", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Contador incrementado exitosamente",
	})
}

// ========================================
// ENDPOINTS DE CONSTRUCCIÓN
// ========================================

// GetConstructionQueue obtiene la cola de construcción del jugador
func (h *RedisHandler) GetConstructionQueue(w http.ResponseWriter, r *http.Request) {
	// Obtener el ID del jugador del contexto de autenticación
	playerIDStr := r.Context().Value("player_id").(string)
	if playerIDStr == "" {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		h.logger.Error("Error parseando ID de jugador", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	// Obtener todas las aldeas del jugador
	allVillages, err := h.villageRepo.GetAllVillages()
	if err != nil {
		h.logger.Error("Error obteniendo aldeas", zap.Error(err), zap.String("player_id", playerID.String()))
		http.Error(w, "Error obteniendo aldeas", http.StatusInternalServerError)
		return
	}

	// Filtrar aldeas del jugador específico
	var playerVillages []*models.VillageWithDetails
	for _, village := range allVillages {
		if village.Village.PlayerID == playerID {
			playerVillages = append(playerVillages, village)
		}
	}

	// Obtener colas de construcción de todas las aldeas del jugador
	var allConstructions []interface{}
	for _, village := range playerVillages {
		// Verificar si hay edificios en construcción
		for _, building := range village.Buildings {
			if building.IsUpgrading {
				constructionItem := map[string]interface{}{
					"village_id":           village.Village.ID.String(),
					"village_name":         village.Village.Name,
					"building_type":        building.Type,
					"current_level":        building.Level,
					"target_level":         building.Level + 1,
					"start_time":           building.UpgradeCompletionTime,
					"estimated_completion": building.UpgradeCompletionTime,
					"status":               "upgrading",
				}
				allConstructions = append(allConstructions, constructionItem)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"constructions":    allConstructions,
			"count":            len(allConstructions),
			"player_id":        playerID.String(),
			"villages_checked": len(playerVillages),
		},
	})
}

// ========================================
// ENDPOINTS DE CHAT
// ========================================

// GetRecentMessages obtiene mensajes recientes de un canal
func (h *RedisHandler) GetRecentMessages(w http.ResponseWriter, r *http.Request) {
	channel := chi.URLParam(r, "channel")
	if channel == "" {
		http.Error(w, "Channel requerido", http.StatusBadRequest)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 20 // default
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	ctx := r.Context()
	messages, err := h.chatService.GetRecentMessages(ctx, channel, limit)
	if err != nil {
		h.logger.Error("Error obteniendo mensajes recientes", zap.Error(err), zap.String("channel", channel))
		http.Error(w, "Error obteniendo mensajes", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"messages": messages,
			"count":    len(messages),
			"channel":  channel,
		},
	})
}

// GetOnlineUsersInChannel obtiene usuarios online en un canal
func (h *RedisHandler) GetOnlineUsersInChannel(w http.ResponseWriter, r *http.Request) {
	channel := chi.URLParam(r, "channel")
	if channel == "" {
		http.Error(w, "Channel requerido", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	users, err := h.chatService.GetOnlineUsers(ctx, channel)
	if err != nil {
		h.logger.Error("Error obteniendo usuarios online en canal", zap.Error(err), zap.String("channel", channel))
		http.Error(w, "Error obteniendo usuarios online", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"users":   users,
			"count":   len(users),
			"channel": channel,
		},
	})
}

// ========================================
// ENDPOINTS DE RANKING
// ========================================

// GetTopPlayers obtiene los mejores jugadores
func (h *RedisHandler) GetTopPlayers(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 100 // default
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	// Obtener categoryID de query params, default 1 (general)
	categoryIDStr := r.URL.Query().Get("category")
	categoryID := 1 // default
	if categoryIDStr != "" {
		if c, err := strconv.Atoi(categoryIDStr); err == nil && c > 0 {
			categoryID = c
		}
	}

	ctx := r.Context()
	players, err := h.rankingService.GetTopPlayers(ctx, categoryID, limit)
	if err != nil {
		h.logger.Error("Error obteniendo top players", zap.Error(err))
		http.Error(w, "Error obteniendo top players", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"players": players,
			"count":   len(players),
		},
	})
}

// GetPlayerRank obtiene el ranking de un jugador específico
func (h *RedisHandler) GetPlayerRank(w http.ResponseWriter, r *http.Request) {
	playerIDStr := chi.URLParam(r, "playerID")
	if playerIDStr == "" {
		http.Error(w, "PlayerID requerido", http.StatusBadRequest)
		return
	}

	playerID, err := strconv.ParseInt(playerIDStr, 10, 64)
	if err != nil {
		http.Error(w, "PlayerID inválido", http.StatusBadRequest)
		return
	}

	// Obtener categoryID de query params, default 1 (general)
	categoryIDStr := r.URL.Query().Get("category")
	categoryID := 1 // default
	if categoryIDStr != "" {
		if c, err := strconv.Atoi(categoryIDStr); err == nil && c > 0 {
			categoryID = c
		}
	}

	ctx := r.Context()
	rank, err := h.rankingService.GetPlayerRank(ctx, int(playerID), categoryID)
	if err != nil {
		h.logger.Error("Error obteniendo ranking del jugador", zap.Error(err), zap.Int64("player_id", playerID))
		http.Error(w, "Error obteniendo ranking", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    rank,
	})
}

// GetRankingStats obtiene estadísticas del ranking
func (h *RedisHandler) GetRankingStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	stats, err := h.rankingService.GetRankingStats(ctx)
	if err != nil {
		h.logger.Error("Error obteniendo estadísticas de ranking", zap.Error(err))
		http.Error(w, "Error obteniendo estadísticas", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    stats,
	})
}

// ========================================
// ENDPOINTS DE EVENTOS
// ========================================

// GetActiveEvents obtiene eventos activos
func (h *RedisHandler) GetActiveEvents(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	events, err := h.eventService.GetActiveEvents(ctx)
	if err != nil {
		h.logger.Error("Error obteniendo eventos activos", zap.Error(err))
		http.Error(w, "Error obteniendo eventos", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"events": events,
			"count":  len(events),
		},
	})
}

// GetEventParticipants obtiene participantes de un evento
func (h *RedisHandler) GetEventParticipants(w http.ResponseWriter, r *http.Request) {
	eventIDStr := chi.URLParam(r, "eventID")
	if eventIDStr == "" {
		http.Error(w, "EventID requerido", http.StatusBadRequest)
		return
	}

	eventID, err := strconv.ParseInt(eventIDStr, 10, 64)
	if err != nil {
		http.Error(w, "EventID inválido", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	participants, err := h.eventService.GetEventParticipants(ctx, eventIDStr)
	if err != nil {
		h.logger.Error("Error obteniendo participantes del evento", zap.Error(err), zap.Int64("event_id", eventID))
		http.Error(w, "Error obteniendo participantes", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"participants": participants,
			"count":        len(participants),
			"event_id":     eventID,
		},
	})
}

// ========================================
// ENDPOINTS DE BATALLAS
// ========================================

// GetBattle obtiene datos de una batalla
func (h *RedisHandler) GetBattle(w http.ResponseWriter, r *http.Request) {
	battleIDStr := chi.URLParam(r, "battleID")
	if battleIDStr == "" {
		http.Error(w, "BattleID requerido", http.StatusBadRequest)
		return
	}

	battleID, err := strconv.ParseInt(battleIDStr, 10, 64)
	if err != nil {
		http.Error(w, "BattleID inválido", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	battle, err := h.battleService.GetBattle(ctx, battleID)
	if err != nil {
		h.logger.Error("Error obteniendo batalla", zap.Error(err), zap.Int64("battle_id", battleID))
		http.Error(w, "Error obteniendo batalla", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    battle,
	})
}

// GetPlayerBattles obtiene las batallas de un jugador
func (h *RedisHandler) GetPlayerBattles(w http.ResponseWriter, r *http.Request) {
	playerIDStr := chi.URLParam(r, "playerID")
	if playerIDStr == "" {
		http.Error(w, "PlayerID requerido", http.StatusBadRequest)
		return
	}

	playerID, err := strconv.ParseInt(playerIDStr, 10, 64)
	if err != nil {
		http.Error(w, "PlayerID inválido", http.StatusBadRequest)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 20 // default
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	ctx := r.Context()
	battles, err := h.battleService.GetPlayerBattles(ctx, playerID, limit)
	if err != nil {
		h.logger.Error("Error obteniendo batallas del jugador", zap.Error(err), zap.Int64("player_id", playerID))
		http.Error(w, "Error obteniendo batallas", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"battles": battles,
			"count":   len(battles),
		},
	})
}

// ========================================
// ENDPOINTS DE RATE LIMITING
// ========================================

// GetRateLimitInfo obtiene información de rate limit
func (h *RedisHandler) GetRateLimitInfo(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "Key requerida", http.StatusBadRequest)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 100 // default
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	windowStr := r.URL.Query().Get("window")
	window := 1 * time.Minute // default
	if windowStr != "" {
		if w, err := strconv.Atoi(windowStr); err == nil && w > 0 {
			window = time.Duration(w) * time.Second
		}
	}

	ctx := r.Context()
	info, err := h.rateLimitService.GetRateLimitInfo(ctx, key, limit, window)
	if err != nil {
		h.logger.Error("Error obteniendo información de rate limit", zap.Error(err), zap.String("key", key))
		http.Error(w, "Error obteniendo información de rate limit", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    info,
	})
}

// GetRateLimitStats obtiene estadísticas de rate limiting
func (h *RedisHandler) GetRateLimitStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	stats, err := h.rateLimitService.GetRateLimitStats(ctx)
	if err != nil {
		h.logger.Error("Error obteniendo estadísticas de rate limiting", zap.Error(err))
		http.Error(w, "Error obteniendo estadísticas", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    stats,
	})
}

// ========================================
// ENDPOINTS DE INVENTARIO
// ========================================

// GetPlayerInventory obtiene el inventario de un jugador
func (h *RedisHandler) GetPlayerInventory(w http.ResponseWriter, r *http.Request) {
	playerIDStr := chi.URLParam(r, "playerID")
	if playerIDStr == "" {
		http.Error(w, "PlayerID requerido", http.StatusBadRequest)
		return
	}

	playerID, err := strconv.ParseInt(playerIDStr, 10, 64)
	if err != nil {
		http.Error(w, "PlayerID inválido", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	inventory, err := h.inventoryService.GetPlayerInventory(ctx, playerID)
	if err != nil {
		h.logger.Error("Error obteniendo inventario del jugador", zap.Error(err), zap.Int64("player_id", playerID))
		http.Error(w, "Error obteniendo inventario", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"inventory": inventory,
			"count":     len(inventory),
		},
	})
}

// GetInventoryStats obtiene estadísticas del inventario
func (h *RedisHandler) GetInventoryStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	stats, err := h.inventoryService.GetInventoryStats(ctx)
	if err != nil {
		h.logger.Error("Error obteniendo estadísticas de inventario", zap.Error(err))
		http.Error(w, "Error obteniendo estadísticas", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    stats,
	})
}

// ========================================
// ENDPOINTS DE CONFIGURACIONES
// ========================================

// GetBuildingConfigs obtiene configuraciones de edificios
func (h *RedisHandler) GetBuildingConfigs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	configs, err := h.configCacheService.GetBuildingConfigs(ctx)
	if err != nil {
		h.logger.Error("Error obteniendo configuraciones de edificios", zap.Error(err))
		http.Error(w, "Error obteniendo configuraciones", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"configs": configs,
			"count":   len(configs),
		},
	})
}

// GetUnitConfigs obtiene configuraciones de unidades
func (h *RedisHandler) GetUnitConfigs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	configs, err := h.configCacheService.GetUnitConfigs(ctx)
	if err != nil {
		h.logger.Error("Error obteniendo configuraciones de unidades", zap.Error(err))
		http.Error(w, "Error obteniendo configuraciones", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"configs": configs,
			"count":   len(configs),
		},
	})
}

// GetTechnologyConfigs obtiene configuraciones de tecnologías
func (h *RedisHandler) GetTechnologyConfigs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	configs, err := h.configCacheService.GetTechnologyConfigs(ctx)
	if err != nil {
		h.logger.Error("Error obteniendo configuraciones de tecnologías", zap.Error(err))
		http.Error(w, "Error obteniendo configuraciones", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"configs": configs,
			"count":   len(configs),
		},
	})
}

// GetConfigStats obtiene estadísticas de configuraciones
func (h *RedisHandler) GetConfigStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	stats, err := h.configCacheService.GetConfigStats(ctx)
	if err != nil {
		h.logger.Error("Error obteniendo estadísticas de configuraciones", zap.Error(err))
		http.Error(w, "Error obteniendo estadísticas", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    stats,
	})
}
