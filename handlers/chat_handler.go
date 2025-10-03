package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"server-backend/config"
	"server-backend/middleware"
	"server-backend/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
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
func (h *ChatHandler) SendMessage(c *gin.Context) {
	var req struct {
		Channel string `json:"channel"`
		Message string `json:"message"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error decodificando request"})
		return
	}

	// Obtener datos del usuario desde el contexto
	playerIDStr := c.GetString("player_id")
	username := c.GetString("username")

	// Convertir playerID de string a UUID
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de jugador inválido"})
		return
	}

	// Validar request
	if req.Channel == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Canal requerido"})
		return
	}

	if req.Message == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Mensaje requerido"})
		return
	}

	// Verificar si el usuario está baneado
	if h.chatService.IsUserBanned(c.Request.Context(), username, req.Channel) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Has sido baneado de este canal"})
		return
	}

	// Enviar mensaje
	err = h.chatService.SendMessage(c.Request.Context(), playerID, username, req.Channel, req.Message)
	if err != nil {
		h.logger.Error("Error enviando mensaje", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"message":  "Mensaje enviado",
		"username": username, // ← AGREGADO: Incluir username en respuesta
		"channel":  req.Channel,
	})
}

// GetRecentMessages obtiene mensajes recientes
func (h *ChatHandler) GetRecentMessages(c *gin.Context) {
	channel := c.Query("channel")
	if channel == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Canal requerido"})
		return
	}

	limitStr := c.Query("limit")
	limit := 20 // default
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	messages, err := h.chatService.GetRecentMessages(c.Request.Context(), channel, limit)
	if err != nil {
		h.logger.Error("Error obteniendo mensajes", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"messages": messages,
		"count":    len(messages),
	})
}

// JoinChannel une a un usuario a un canal
func (h *ChatHandler) JoinChannel(c *gin.Context) {
	var req struct {
		Channel string `json:"channel"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error decodificando request"})
		return
	}

	// Obtener datos del usuario
	playerIDStr := c.GetString("player_id")
	username := c.GetString("username")

	// Convertir playerID de string a UUID
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de jugador inválido"})
		return
	}

	// Verificar si el usuario está baneado
	if h.chatService.IsUserBanned(c.Request.Context(), username, req.Channel) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Has sido baneado de este canal"})
		return
	}

	// Unirse al canal
	err = h.chatService.JoinChannel(c.Request.Context(), playerID, username, req.Channel)
	if err != nil {
		h.logger.Error("Error uniéndose al canal", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Te has unido al canal",
	})
}

// LeaveChannel saca a un usuario de un canal
func (h *ChatHandler) LeaveChannel(c *gin.Context) {
	var req struct {
		Channel string `json:"channel"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error decodificando request"})
		return
	}

	// Obtener datos del usuario
	playerIDStr := c.GetString("player_id")
	username := c.GetString("username")

	// Convertir playerID de string a UUID
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de jugador inválido"})
		return
	}

	// Salir del canal
	err = h.chatService.LeaveChannel(c.Request.Context(), playerID, username, req.Channel)
	if err != nil {
		h.logger.Error("Error saliendo del canal", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Has salido del canal",
	})
}

// GetOnlineUsers obtiene usuarios online en un canal
func (h *ChatHandler) GetOnlineUsers(c *gin.Context) {
	channel := c.Query("channel")
	if channel == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Canal requerido"})
		return
	}

	users, err := h.chatService.GetOnlineUsers(c.Request.Context(), channel)
	if err != nil {
		h.logger.Error("Error obteniendo usuarios online", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"users":   users,
		"count":   len(users),
	})
}

// GetChannelInfo obtiene información de un canal
func (h *ChatHandler) GetChannelInfo(c *gin.Context) {
	channel := c.Query("channel")
	if channel == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Canal requerido"})
		return
	}

	info, err := h.chatService.GetChannelInfo(c.Request.Context(), channel)
	if err != nil {
		h.logger.Error("Error obteniendo información del canal", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"channel": info,
	})
}

// GetChatStats obtiene estadísticas del chat
func (h *ChatHandler) GetChatStats(c *gin.Context) {
	stats, err := h.chatService.GetChatStats(c.Request.Context())
	if err != nil {
		h.logger.Error("Error obteniendo estadísticas", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"stats":   stats,
	})
}

// CreateAllianceChannel crea un canal de alianza
func (h *ChatHandler) CreateAllianceChannel(c *gin.Context) {
	var req struct {
		AllianceID   string `json:"alliance_id"`
		AllianceName string `json:"alliance_name"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error decodificando request"})
		return
	}

	if req.AllianceID == "" || req.AllianceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "AllianceID y AllianceName requeridos"})
		return
	}

	err := h.chatService.CreateAllianceChannel(c.Request.Context(), req.AllianceID, req.AllianceName)
	if err != nil {
		h.logger.Error("Error creando canal de alianza", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Canal de alianza creado",
		"channel": fmt.Sprintf("alliance:%s", req.AllianceID),
	})
}

// BanUser banea a un usuario de un canal (solo moderadores)
func (h *ChatHandler) BanUser(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Channel  string `json:"channel"`
		Duration int    `json:"duration"` // en minutos, 0 = permanente
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error decodificando request"})
		return
	}

	// Verificar permisos (aquí podrías verificar si es moderador)
	// Por ahora, permitimos que cualquier usuario pueda banear (para pruebas)

	var duration time.Duration
	if req.Duration > 0 {
		duration = time.Duration(req.Duration) * time.Minute
	}

	err := h.chatService.BanUser(c.Request.Context(), req.Username, req.Channel, duration)
	if err != nil {
		h.logger.Error("Error baneando usuario", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": fmt.Sprintf("Usuario %s baneado", req.Username),
	})
}

// SendSystemMessage envía un mensaje del sistema (solo admins)
func (h *ChatHandler) SendSystemMessage(c *gin.Context) {
	var req struct {
		Channel string `json:"channel"`
		Message string `json:"message"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error decodificando request"})
		return
	}

	// Verificar permisos de admin (aquí podrías verificar el rol)
	// Por ahora, permitimos que cualquier usuario pueda enviar mensajes del sistema

	err := h.chatService.SendSystemMessage(c.Request.Context(), req.Channel, req.Message)
	if err != nil {
		h.logger.Error("Error enviando mensaje del sistema", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Mensaje del sistema enviado",
	})
}

// GetChannels obtiene todos los canales disponibles desde BD
func (h *ChatHandler) GetChannels(c *gin.Context) {
	// Obtener canales desde el servicio (que los obtiene de BD)
	channels, err := h.chatService.GetAllChannels(c.Request.Context())
	if err != nil {
		h.logger.Error("Error obteniendo canales", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convertir a formato de respuesta
	var responseChannels []map[string]interface{}
	for _, channel := range channels {
		channelData := map[string]interface{}{
			"id":           channel.ID,
			"name":         channel.Name,
			"type":         channel.Type,
			"member_count": channel.MemberCount,
			"is_active":    channel.IsActive,
			"created_at":   channel.CreatedAt,
		}

		// Agregar campos específicos según el tipo
		if channel.WorldID != "" {
			channelData["world_id"] = channel.WorldID
		}

		if channel.Type == "alliance" && channel.AllianceID != "" {
			channelData["alliance_id"] = channel.AllianceID
		}

		responseChannels = append(responseChannels, channelData)
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"channels": responseChannels,
		"count":    len(responseChannels),
	})
}

// WebSocket endpoint para chat en tiempo real
func (h *ChatHandler) WebSocketChat(c *gin.Context) {
	// Obtener canal de query params
	channel := c.Query("channel")
	if channel == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Canal requerido"})
		return
	}

	// Obtener datos del usuario
	playerIDStr := c.GetString("player_id")
	username := c.GetString("username")

	// Convertir playerID de string a UUID
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de jugador inválido"})
		return
	}

	// Verificar si el usuario está baneado
	if h.chatService.IsUserBanned(c.Request.Context(), username, channel) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Has sido baneado de este canal"})
		return
	}

	// Unirse al canal
	err = h.chatService.JoinChannel(c.Request.Context(), playerID, username, channel)
	if err != nil {
		h.logger.Error("Error uniéndose al canal", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Suscribirse al canal de Redis
	pubsub, err := h.chatService.SubscribeToChannel(c.Request.Context(), channel)
	if err != nil {
		h.logger.Error("Error suscribiéndose al canal", zap.Error(err))
		// Si Redis no está disponible, continuar sin suscripción
		if err.Error() == "Redis no disponible - suscripción no disponible" {
			h.logger.Warn("Redis no disponible - WebSocket funcionará sin tiempo real")
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	defer func() {
		if pubsub != nil {
			pubsub.Close()
		}
	}()

	// Configurar WebSocket con configuración robusta
	upgrader := config.GetWebSocketUpgrader()

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Error("Error actualizando a WebSocket", zap.Error(err))
		return
	}
	defer conn.Close()

	// Crear validador de mensajes
	validator := middleware.NewWebSocketValidator(h.logger)
	clientInfo := fmt.Sprintf("%s@%s", username, channel)

	// Configurar timeouts para la conexión
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

	// Canal para mensajes de Redis (solo si Redis está disponible)
	var redisChan <-chan *redis.Message
	if pubsub != nil {
		redisChan = pubsub.Channel()

		// Goroutine segura para enviar mensajes de Redis al WebSocket
		go func() {
			defer func() {
				if r := recover(); r != nil {
					h.logger.Error("Panic en goroutine de Redis", zap.Any("panic", r))
				}
			}()

			for msg := range redisChan {
				// Validar mensaje de Redis antes de enviar
				if len(msg.Payload) == 0 {
					h.logger.Warn("Mensaje vacío de Redis ignorado")
					continue
				}

				// Enviar con timeout
				conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
				err := conn.WriteMessage(websocket.TextMessage, []byte(msg.Payload))
				if err != nil {
					h.logger.Error("Error enviando mensaje WebSocket", zap.Error(err))
					break
				}

				// Log del mensaje enviado
				validator.LogMessageSent([]byte(msg.Payload), clientInfo)
			}
		}()
	}

	// Leer mensajes del WebSocket con validación robusta
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				h.logger.Error("Error leyendo mensaje WebSocket", zap.Error(err))
			}
			break
		}

		// Log del mensaje recibido
		validator.LogMessageReceived(message, clientInfo)

		// Validar tipo de mensaje
		if !validator.IsValidMessageType(messageType) {
			validator.SendError(conn, "Solo se permiten mensajes de texto")
			continue
		}

		// Validar y parsear mensaje
		data, err := validator.ValidateMessage(message)
		if err != nil {
			h.logger.Warn("Mensaje inválido recibido",
				zap.Error(err),
				zap.String("message", string(message)),
				zap.String("client", clientInfo))
			validator.SendError(conn, fmt.Sprintf("Mensaje inválido: %s", err.Error()))
			continue
		}

		// Extraer mensaje validado
		messageText := data["message"].(string)

		// Enviar mensaje al servicio
		err = h.chatService.SendMessage(c.Request.Context(), playerID, username, channel, messageText)
		if err != nil {
			h.logger.Error("Error enviando mensaje", zap.Error(err))
			validator.SendError(conn, fmt.Sprintf("Error enviando mensaje: %s", err.Error()))
			continue
		}

		// Enviar confirmación al cliente
		validator.SendSuccess(conn, "Mensaje enviado exitosamente")

		// Actualizar deadline de lectura
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	}

	// Salir del canal cuando se desconecte
	h.chatService.LeaveChannel(c.Request.Context(), playerID, username, channel)
}
