package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"server-backend/services"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type ChatHandler struct {
	chatService *services.ChatService
	logger      *zap.Logger
}

func NewChatHandler(chatService *services.ChatService, logger *zap.Logger) *ChatHandler {
	return &ChatHandler{
		chatService: chatService,
		logger:      logger,
	}
}

// SendMessage envía un mensaje al chat
func (h *ChatHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Channel string `json:"channel"`
		Message string `json:"message"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Error decodificando request", http.StatusBadRequest)
		return
	}

	// Obtener datos del usuario desde el contexto
	playerID := r.Context().Value("player_id").(int64)
	username := r.Context().Value("username").(string)

	// Validar request
	if req.Channel == "" {
		http.Error(w, "Canal requerido", http.StatusBadRequest)
		return
	}

	if req.Message == "" {
		http.Error(w, "Mensaje requerido", http.StatusBadRequest)
		return
	}

	// Verificar si el usuario está baneado
	if h.chatService.IsUserBanned(r.Context(), username, req.Channel) {
		http.Error(w, "Has sido baneado de este canal", http.StatusForbidden)
		return
	}

	// Enviar mensaje
	err := h.chatService.SendMessage(r.Context(), playerID, username, req.Channel, req.Message)
	if err != nil {
		h.logger.Error("Error enviando mensaje", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Mensaje enviado",
	})
}

// GetRecentMessages obtiene mensajes recientes
func (h *ChatHandler) GetRecentMessages(w http.ResponseWriter, r *http.Request) {
	channel := r.URL.Query().Get("channel")
	if channel == "" {
		http.Error(w, "Canal requerido", http.StatusBadRequest)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 20 // default
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	messages, err := h.chatService.GetRecentMessages(r.Context(), channel, limit)
	if err != nil {
		h.logger.Error("Error obteniendo mensajes", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"messages": messages,
		"count":    len(messages),
	})
}

// JoinChannel une a un usuario a un canal
func (h *ChatHandler) JoinChannel(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Channel string `json:"channel"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Error decodificando request", http.StatusBadRequest)
		return
	}

	// Obtener datos del usuario
	playerID := r.Context().Value("player_id").(int64)
	username := r.Context().Value("username").(string)

	// Verificar si el usuario está baneado
	if h.chatService.IsUserBanned(r.Context(), username, req.Channel) {
		http.Error(w, "Has sido baneado de este canal", http.StatusForbidden)
		return
	}

	// Unirse al canal
	err := h.chatService.JoinChannel(r.Context(), playerID, username, req.Channel)
	if err != nil {
		h.logger.Error("Error uniéndose al canal", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Te has unido al canal",
	})
}

// LeaveChannel saca a un usuario de un canal
func (h *ChatHandler) LeaveChannel(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Channel string `json:"channel"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Error decodificando request", http.StatusBadRequest)
		return
	}

	// Obtener datos del usuario
	playerID := r.Context().Value("player_id").(int64)
	username := r.Context().Value("username").(string)

	// Salir del canal
	err := h.chatService.LeaveChannel(r.Context(), playerID, username, req.Channel)
	if err != nil {
		h.logger.Error("Error saliendo del canal", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Has salido del canal",
	})
}

// GetOnlineUsers obtiene usuarios online en un canal
func (h *ChatHandler) GetOnlineUsers(w http.ResponseWriter, r *http.Request) {
	channel := r.URL.Query().Get("channel")
	if channel == "" {
		http.Error(w, "Canal requerido", http.StatusBadRequest)
		return
	}

	users, err := h.chatService.GetOnlineUsers(r.Context(), channel)
	if err != nil {
		h.logger.Error("Error obteniendo usuarios online", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"users":   users,
		"count":   len(users),
	})
}

// GetChannelInfo obtiene información de un canal
func (h *ChatHandler) GetChannelInfo(w http.ResponseWriter, r *http.Request) {
	channel := r.URL.Query().Get("channel")
	if channel == "" {
		http.Error(w, "Canal requerido", http.StatusBadRequest)
		return
	}

	info, err := h.chatService.GetChannelInfo(r.Context(), channel)
	if err != nil {
		h.logger.Error("Error obteniendo información del canal", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"channel": info,
	})
}

// GetChatStats obtiene estadísticas del chat
func (h *ChatHandler) GetChatStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.chatService.GetChatStats(r.Context())
	if err != nil {
		h.logger.Error("Error obteniendo estadísticas", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"stats":   stats,
	})
}

// CreateAllianceChannel crea un canal de alianza
func (h *ChatHandler) CreateAllianceChannel(w http.ResponseWriter, r *http.Request) {
	var req struct {
		AllianceID   string `json:"alliance_id"`
		AllianceName string `json:"alliance_name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Error decodificando request", http.StatusBadRequest)
		return
	}

	if req.AllianceID == "" || req.AllianceName == "" {
		http.Error(w, "AllianceID y AllianceName requeridos", http.StatusBadRequest)
		return
	}

	err := h.chatService.CreateAllianceChannel(r.Context(), req.AllianceID, req.AllianceName)
	if err != nil {
		h.logger.Error("Error creando canal de alianza", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Canal de alianza creado",
		"channel": fmt.Sprintf("alliance:%s", req.AllianceID),
	})
}

// BanUser banea a un usuario de un canal (solo moderadores)
func (h *ChatHandler) BanUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Channel  string `json:"channel"`
		Duration int    `json:"duration"` // en minutos, 0 = permanente
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Error decodificando request", http.StatusBadRequest)
		return
	}

	// Verificar permisos (aquí podrías verificar si es moderador)
	// Por ahora, permitimos que cualquier usuario pueda banear (para pruebas)

	var duration time.Duration
	if req.Duration > 0 {
		duration = time.Duration(req.Duration) * time.Minute
	}

	err := h.chatService.BanUser(r.Context(), req.Username, req.Channel, duration)
	if err != nil {
		h.logger.Error("Error baneando usuario", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Usuario %s baneado", req.Username),
	})
}

// SendSystemMessage envía un mensaje del sistema (solo admins)
func (h *ChatHandler) SendSystemMessage(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Channel string `json:"channel"`
		Message string `json:"message"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Error decodificando request", http.StatusBadRequest)
		return
	}

	// Verificar permisos de admin (aquí podrías verificar el rol)
	// Por ahora, permitimos que cualquier usuario pueda enviar mensajes del sistema

	err := h.chatService.SendSystemMessage(r.Context(), req.Channel, req.Message)
	if err != nil {
		h.logger.Error("Error enviando mensaje del sistema", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Mensaje del sistema enviado",
	})
}

// GetChannels obtiene todos los canales disponibles
func (h *ChatHandler) GetChannels(w http.ResponseWriter, r *http.Request) {
	// Por ahora, retornamos canales predefinidos
	// En el futuro, esto podría obtener canales dinámicos desde Redis
	channels := []map[string]interface{}{
		{
			"id":           "global",
			"name":         "Chat Global",
			"type":         "global",
			"member_count": 0,
			"is_active":    true,
		},
		{
			"id":           "alliance:1",
			"name":         "Alianza Ejemplo",
			"type":         "alliance",
			"alliance_id":  "1",
			"member_count": 0,
			"is_active":    true,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"channels": channels,
		"count":    len(channels),
	})
}

// WebSocket endpoint para chat en tiempo real
func (h *ChatHandler) WebSocketChat(w http.ResponseWriter, r *http.Request) {
	// Obtener canal de query params
	channel := r.URL.Query().Get("channel")
	if channel == "" {
		http.Error(w, "Canal requerido", http.StatusBadRequest)
		return
	}

	// Obtener datos del usuario
	playerID := r.Context().Value("player_id").(int64)
	username := r.Context().Value("username").(string)

	// Verificar si el usuario está baneado
	if h.chatService.IsUserBanned(r.Context(), username, channel) {
		http.Error(w, "Has sido baneado de este canal", http.StatusForbidden)
		return
	}

	// Unirse al canal
	err := h.chatService.JoinChannel(r.Context(), playerID, username, channel)
	if err != nil {
		h.logger.Error("Error uniéndose al canal", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Suscribirse al canal de Redis
	pubsub, err := h.chatService.SubscribeToChannel(r.Context(), channel)
	if err != nil {
		h.logger.Error("Error suscribiéndose al canal", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer pubsub.Close()

	// Configurar WebSocket
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // En producción, configurar esto adecuadamente
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("Error actualizando a WebSocket", zap.Error(err))
		return
	}
	defer conn.Close()

	// Canal para mensajes de Redis
	redisChan := pubsub.Channel()

	// Goroutine para enviar mensajes de Redis al WebSocket
	go func() {
		for msg := range redisChan {
			err := conn.WriteMessage(websocket.TextMessage, []byte(msg.Payload))
			if err != nil {
				h.logger.Error("Error enviando mensaje WebSocket", zap.Error(err))
				break
			}
		}
	}()

	// Leer mensajes del WebSocket y enviarlos a Redis
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			h.logger.Error("Error leyendo mensaje WebSocket", zap.Error(err))
			break
		}

		// Parsear mensaje
		var chatMsg struct {
			Message string `json:"message"`
		}

		if err := json.Unmarshal(message, &chatMsg); err != nil {
			h.logger.Error("Error parseando mensaje", zap.Error(err))
			continue
		}

		// Enviar mensaje
		err = h.chatService.SendMessage(r.Context(), playerID, username, channel, chatMsg.Message)
		if err != nil {
			h.logger.Error("Error enviando mensaje", zap.Error(err))
			// Enviar error al cliente
			errorMsg := map[string]interface{}{
				"type":    "error",
				"message": err.Error(),
			}
			errorJSON, _ := json.Marshal(errorMsg)
			conn.WriteMessage(websocket.TextMessage, errorJSON)
		}
	}

	// Salir del canal cuando se desconecte
	h.chatService.LeaveChannel(r.Context(), playerID, username, channel)
}
