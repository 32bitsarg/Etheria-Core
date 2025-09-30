package websocket

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"server-backend/models"
	"server-backend/repository"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// RedisInterface define los métodos de Redis que necesita el WebSocket manager
type RedisInterface interface {
	SetUserOnline(userID string, username string) error
	SetUserOffline(userID string) error
	Publish(channel string, message interface{}) error
	GetOnlineUsers() ([]string, error)
	IsUserOnline(userID string) (bool, error)
	Subscribe(channel string) interface{}
}

type Client struct {
	ID       string
	PlayerID string
	Username string
	Conn     *websocket.Conn
	Channels map[string]bool // Canales a los que está suscrito
	Send     chan []byte
	Manager  *Manager
}

type Manager struct {
	clients      map[string]*Client
	broadcast    chan []byte
	register     chan *Client
	unregister   chan *Client
	chatRepo     *repository.ChatRepository
	villageRepo  *repository.VillageRepository
	unitRepo     *repository.UnitRepository
	logger       *zap.Logger
	mutex        sync.RWMutex
	upgrader     websocket.Upgrader
	redisService RedisInterface
}

type WSMessage struct {
	Type   string                 `json:"type"`
	Data   map[string]interface{} `json:"data"`
	UserID string                 `json:"user_id,omitempty"`
	Time   time.Time              `json:"time"`
}

func NewManager(chatRepo *repository.ChatRepository, villageRepo *repository.VillageRepository, unitRepo *repository.UnitRepository, logger *zap.Logger, redisService RedisInterface) *Manager {
	return &Manager{
		clients:     make(map[string]*Client),
		broadcast:   make(chan []byte),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		chatRepo:    chatRepo,
		villageRepo: villageRepo,
		unitRepo:    unitRepo,
		logger:      logger,
		mutex:       sync.RWMutex{},
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true // En producción, configurar esto adecuadamente
			},
		},
		redisService: redisService,
	}
}

func (m *Manager) Start() {
	// Iniciar worker de Redis Pub/Sub
	go m.startRedisSubscriber()

	for {
		select {
		case client := <-m.register:
			m.registerClient(client)

		case client := <-m.unregister:
			m.unregisterClient(client)

		case message := <-m.broadcast:
			m.broadcastMessage(message)
		}
	}
}

func (m *Manager) registerClient(client *Client) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.clients[client.ID] = client

	// Marcar usuario como online en Redis
	err := m.redisService.SetUserOnline(client.PlayerID, client.Username)
	if err != nil {
		log.Printf("Error marcando usuario online: %v", err)
	}

	// Publicar evento de usuario online
	event := WSMessage{
		Type: "user_online",
		Data: map[string]interface{}{
			"user_id":  client.PlayerID,
			"username": client.Username,
		},
		Time: time.Now(),
	}

	m.publishToRedis(event)

	log.Printf("Cliente registrado: %s (Usuario: %s)", client.ID, client.Username)
}

func (m *Manager) unregisterClient(client *Client) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, ok := m.clients[client.ID]; ok {
		delete(m.clients, client.ID)
		close(client.Send)

		// Marcar usuario como offline en Redis
		err := m.redisService.SetUserOffline(client.PlayerID)
		if err != nil {
			log.Printf("Error marcando usuario offline: %v", err)
		}

		// Publicar evento de usuario offline
		event := WSMessage{
			Type: "user_offline",
			Data: map[string]interface{}{
				"user_id":  client.PlayerID,
				"username": client.Username,
			},
			Time: time.Now(),
		}

		m.publishToRedis(event)

		log.Printf("Cliente desregistrado: %s (Usuario: %s)", client.ID, client.Username)
	}
}

func (m *Manager) broadcastMessage(message []byte) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	for _, client := range m.clients {
		select {
		case client.Send <- message:
		default:
			close(client.Send)
			delete(m.clients, client.ID)
		}
	}
}

func (m *Manager) SendToUser(userID string, messageType string, data map[string]interface{}) error {
	message := WSMessage{
		Type:   messageType,
		Data:   data,
		UserID: userID,
		Time:   time.Now(),
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("error serializando mensaje: %v", err)
	}

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Enviar a todos los clientes del usuario
	for _, client := range m.clients {
		if client.PlayerID == userID {
			select {
			case client.Send <- messageBytes:
			default:
				// Cliente desconectado, remover
				close(client.Send)
				delete(m.clients, client.ID)
			}
		}
	}

	return nil
}

func (m *Manager) SendToAll(messageType string, data map[string]interface{}) error {
	message := WSMessage{
		Type: messageType,
		Data: data,
		Time: time.Now(),
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("error serializando mensaje: %v", err)
	}

	m.broadcast <- messageBytes
	return nil
}

func (m *Manager) GetOnlineUsers() []string {
	users, err := m.redisService.GetOnlineUsers()
	if err != nil {
		m.logger.Error("Error obteniendo usuarios online", zap.Error(err))
		return []string{}
	}
	return users
}

func (m *Manager) IsUserOnline(userID string) bool {
	online, err := m.redisService.IsUserOnline(userID)
	if err != nil {
		m.logger.Error("Error verificando estado online", zap.Error(err))
		return false
	}
	return online
}

func (m *Manager) publishToRedis(message WSMessage) {
	messageBytes, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error serializando mensaje para Redis: %v", err)
		return
	}

	// Publicar en canal de WebSocket
	err = m.redisService.Publish("websocket:broadcast", messageBytes)
	if err != nil {
		log.Printf("Error publicando en Redis: %v", err)
	}
}

func (m *Manager) startRedisSubscriber() {
	// Nota: La implementación específica dependerá del tipo de pubsub retornado
	// Por ahora, esto es un placeholder
	m.logger.Info("Redis subscriber iniciado")
}

func (m *Manager) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := m.upgrader.Upgrade(w, r, nil)
	if err != nil {
		m.logger.Error("Error al actualizar conexión WebSocket", zap.Error(err))
		return
	}

	// Obtener información del jugador del contexto (si está autenticado)
	playerID := r.Context().Value("player_id")
	username := r.Context().Value("username")

	clientID := ""
	playerIDStr := ""
	usernameStr := ""

	if playerID != nil {
		playerIDStr = playerID.(string)
		usernameStr = username.(string)
		clientID = playerIDStr
	} else {
		clientID = generateClientID()
	}

	client := &Client{
		ID:       clientID,
		PlayerID: playerIDStr,
		Username: usernameStr,
		Conn:     conn,
		Channels: make(map[string]bool),
		Send:     make(chan []byte, 256),
		Manager:  m,
	}

	m.register <- client

	// Iniciar goroutines para manejar el cliente
	go client.writePump()
	go client.readPump()
}

func (c *Client) readPump() {
	defer func() {
		c.Manager.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(512)
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.Manager.logger.Error("Error al leer mensaje WebSocket", zap.Error(err))
			}
			break
		}

		c.handleMessage(message)
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PongMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) handleMessage(message []byte) {
	var wsMessage WSMessage
	err := json.Unmarshal(message, &wsMessage)
	if err != nil {
		c.Manager.logger.Error("Error al decodificar mensaje WebSocket", zap.Error(err))
		return
	}

	switch wsMessage.Type {
	case "ping":
		// Responder con pong
		response := WSMessage{
			Type: "pong",
			Data: map[string]interface{}{
				"timestamp": time.Now().Unix(),
			},
			Time: time.Now(),
		}
		responseBytes, _ := json.Marshal(response)
		c.Send <- responseBytes

	case "chat":
		// Reenviar mensaje de chat a todos
		wsMessage.UserID = c.PlayerID
		wsMessage.Time = time.Now()
		messageBytes, _ := json.Marshal(wsMessage)
		c.Manager.broadcast <- messageBytes

	case "private_message":
		// Enviar mensaje privado
		if targetPlayerID, ok := wsMessage.Data["target_player_id"].(string); ok {
			wsMessage.UserID = c.PlayerID
			wsMessage.Time = time.Now()
			c.Manager.SendToUser(targetPlayerID, "private_message", wsMessage.Data)
		}

	default:
		c.Manager.logger.Warn("Tipo de mensaje WebSocket desconocido", zap.String("type", wsMessage.Type))
	}
}

// Funciones auxiliares
func generateClientID() string {
	return "anon_" + time.Now().Format("20060102150405")
}

// Métodos para enviar mensajes específicos
func (m *Manager) SendResourceUpdate(villageID string, resources models.ResourceUpdate) {
	message := WSMessage{
		Type: "resource_update",
		Data: map[string]interface{}{
			"village_id": villageID,
			"resources":  resources,
		},
		Time: time.Now(),
	}

	messageBytes, _ := json.Marshal(message)
	m.broadcast <- messageBytes
}

func (m *Manager) SendBuildingUpdate(villageID string, building models.Building) {
	message := WSMessage{
		Type: "building_update",
		Data: map[string]interface{}{
			"village_id": villageID,
			"building":   building,
		},
		Time: time.Now(),
	}

	messageBytes, _ := json.Marshal(message)
	m.broadcast <- messageBytes
}

func (m *Manager) SendUnitUpdate(villageID string, unit models.Unit) {
	message := WSMessage{
		Type: "unit_update",
		Data: map[string]interface{}{
			"village_id": villageID,
			"unit":       unit,
		},
		Time: time.Now(),
	}

	messageBytes, _ := json.Marshal(message)
	m.broadcast <- messageBytes
}

func (m *Manager) GetClientCount() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return len(m.clients)
}

func (m *Manager) GetUserClientCount(userID string) int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	count := 0
	for _, client := range m.clients {
		if client.PlayerID == userID {
			count++
		}
	}

	return count
}

func (m *Manager) DisconnectUser(userID string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for _, client := range m.clients {
		if client.PlayerID == userID {
			client.Conn.Close()
		}
	}
}

// Métodos específicos para notificaciones del sistema
func (m *Manager) SendAchievementNotification(userID string, achievement models.Achievement) {
	message := WSMessage{
		Type: "achievement_unlocked",
		Data: map[string]interface{}{
			"achievement_id":   achievement.ID.String(),
			"achievement_name": achievement.Name,
			"description":      achievement.Description,
			"category":         achievement.CategoryID.String(),
			"points":           achievement.Points,
		},
		Time: time.Now(),
	}

	m.SendToUser(userID, "achievement_notification", message.Data)
}

func (m *Manager) SendQuestNotification(userID string, quest models.Quest) {
	message := WSMessage{
		Type: "quest_update",
		Data: map[string]interface{}{
			"quest_id":   quest.ID.String(),
			"quest_name": quest.Name,
			"status":     "available",
		},
		Time: time.Now(),
	}

	m.SendToUser(userID, "quest_notification", message.Data)
}

func (m *Manager) SendEventNotification(userID string, event models.Event) {
	message := WSMessage{
		Type: "event_notification",
		Data: map[string]interface{}{
			"event_id":   event.ID.String(),
			"event_name": event.Name,
			"status":     event.Status,
			"phase":      event.Phase,
		},
		Time: time.Now(),
	}

	m.SendToUser(userID, "event_notification", message.Data)
}

func (m *Manager) SendTitleNotification(userID string, title models.Title) {
	message := WSMessage{
		Type: "title_unlocked",
		Data: map[string]interface{}{
			"title_id":   title.ID.String(),
			"title_name": title.Name,
			"rarity":     title.Rarity,
			"category":   title.CategoryID.String(),
		},
		Time: time.Now(),
	}

	m.SendToUser(userID, "title_notification", message.Data)
}

func (m *Manager) SendBattleNotification(userID string, battle models.Battle) {
	message := WSMessage{
		Type: "battle_notification",
		Data: map[string]interface{}{
			"battle_id":   battle.ID.String(),
			"battle_type": battle.BattleType,
			"status":      battle.Status,
			"attacker_id": battle.AttackerID.String(),
			"defender_id": battle.DefenderID.String(),
		},
		Time: time.Now(),
	}

	m.SendToUser(userID, "battle_notification", message.Data)
}

func (m *Manager) SendEconomyNotification(userID string, notificationType string, data map[string]interface{}) {
	message := WSMessage{
		Type: "economy_notification",
		Data: map[string]interface{}{
			"notification_type": notificationType,
			"data":             data,
		},
		Time: time.Now(),
	}

	m.SendToUser(userID, "economy_notification", message.Data)
}

func (m *Manager) SendConstructionNotification(userID string, buildingType string, level int, status string) {
	message := WSMessage{
		Type: "construction_notification",
		Data: map[string]interface{}{
			"building_type": buildingType,
			"level":         level,
			"status":        status,
		},
		Time: time.Now(),
	}

	m.SendToUser(userID, "construction_notification", message.Data)
}

func (m *Manager) SendResearchNotification(userID string, technology models.Technology, status string) {
	message := WSMessage{
		Type: "research_notification",
		Data: map[string]interface{}{
			"technology_id":   technology.ID,
			"technology_name": technology.Name,
			"status":          status,
			"level":           technology.Level,
		},
		Time: time.Now(),
	}

	m.SendToUser(userID, "research_notification", message.Data)
}

func (m *Manager) SendAllianceNotification(userID string, alliance models.Alliance, notificationType string, data map[string]interface{}) {
	message := WSMessage{
		Type: "alliance_notification",
		Data: map[string]interface{}{
			"alliance_id":      alliance.ID,
			"alliance_name":    alliance.Name,
			"notification_type": notificationType,
			"data":             data,
		},
		Time: time.Now(),
	}

	m.SendToUser(userID, "alliance_notification", message.Data)
}

func (m *Manager) SendSystemNotification(userID string, title string, message string, notificationType string) {
	wsMessage := WSMessage{
		Type: "system_notification",
		Data: map[string]interface{}{
			"title":             title,
			"message":           message,
			"notification_type": notificationType,
			"timestamp":         time.Now().Unix(),
		},
		Time: time.Now(),
	}

	m.SendToUser(userID, "system_notification", wsMessage.Data)
}

func (m *Manager) SendGlobalNotification(title string, message string, notificationType string) {
	wsMessage := WSMessage{
		Type: "global_notification",
		Data: map[string]interface{}{
			"title":             title,
			"message":           message,
			"notification_type": notificationType,
			"timestamp":         time.Now().Unix(),
		},
		Time: time.Now(),
	}

	messageBytes, _ := json.Marshal(wsMessage)
	m.broadcast <- messageBytes
}
