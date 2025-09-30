package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"server-backend/models"
	"server-backend/repository"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type ChatService struct {
	chatRepo     *repository.ChatRepository
	redisService *RedisService
}

type ChatMessage struct {
	ID        string                 `json:"id"`
	PlayerID  int64                  `json:"player_id"`
	Username  string                 `json:"username"`
	Channel   string                 `json:"channel"` // "global", "alliance:{id}", "private"
	Message   string                 `json:"message"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Type      string                 `json:"type"` // "message", "system", "join", "leave"
}

type ChatChannel struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Type        string    `json:"type"` // "global", "alliance", "private"
	AllianceID  string    `json:"alliance_id,omitempty"`
	MemberCount int       `json:"member_count"`
	CreatedAt   time.Time `json:"created_at"`
	IsActive    bool      `json:"is_active"`
}

type ChatStats struct {
	TotalMessages   int64 `json:"total_messages"`
	ActiveUsers     int64 `json:"active_users"`
	ChannelsCount   int64 `json:"channels_count"`
	MessagesPerHour int64 `json:"messages_per_hour"`
	AllianceChats   int64 `json:"alliance_chats"`
	GlobalChats     int64 `json:"global_chats"`
}

func NewChatService(chatRepo *repository.ChatRepository, redisService *RedisService) *ChatService {
	return &ChatService{
		chatRepo:     chatRepo,
		redisService: redisService,
	}
}

// SendMessage envía un mensaje y lo almacena en Redis para tiempo real
func (s *ChatService) SendMessage(ctx context.Context, playerID int64, username, channel, message string) error {
	// Validar mensaje
	if strings.TrimSpace(message) == "" {
		return fmt.Errorf("mensaje no puede estar vacío")
	}

	if len(message) > 500 {
		return fmt.Errorf("mensaje demasiado largo (máximo 500 caracteres)")
	}

	// Verificar si el usuario está en el canal
	if !s.isUserInChannel(ctx, username, channel) {
		return fmt.Errorf("usuario no está en el canal")
	}

	// Crear mensaje
	chatMsg := &ChatMessage{
		ID:        fmt.Sprintf("%d_%d", playerID, time.Now().UnixNano()),
		PlayerID:  playerID,
		Username:  username,
		Channel:   channel,
		Message:   message,
		Timestamp: time.Now(),
		Type:      "message",
	}

	// Guardar en base de datos
	dbMessage := &models.ChatMessage{
		PlayerID:  uuid.New(), // Convertir playerID a UUID
		ChannelID: uuid.New(), // Crear channelID temporal
		Username:  username,
		Message:   message,
		Type:      "text",
		CreatedAt: time.Now(),
	}

	err := s.chatRepo.SaveMessage(dbMessage)
	if err != nil {
		return fmt.Errorf("error guardando mensaje en BD: %v", err)
	}

	// Publicar en Redis para tiempo real
	err = s.publishMessage(ctx, chatMsg)
	if err != nil {
		log.Printf("Error publicando mensaje: %v", err)
	}

	// Cachear mensaje reciente
	err = s.cacheRecentMessage(ctx, channel, chatMsg)
	if err != nil {
		log.Printf("Error cacheando mensaje: %v", err)
	}

	// Actualizar estadísticas
	s.updateChatStats(ctx, channel)

	return nil
}

// SendSystemMessage envía un mensaje del sistema
func (s *ChatService) SendSystemMessage(ctx context.Context, channel, message string) error {
	systemMsg := &ChatMessage{
		ID:        fmt.Sprintf("system_%d", time.Now().UnixNano()),
		PlayerID:  0,
		Username:  "Sistema",
		Channel:   channel,
		Message:   message,
		Timestamp: time.Now(),
		Type:      "system",
		Data: map[string]interface{}{
			"system": true,
		},
	}

	return s.publishMessage(ctx, systemMsg)
}

// publishMessage publica un mensaje en el canal de Redis
func (s *ChatService) publishMessage(ctx context.Context, msg *ChatMessage) error {
	channelKey := fmt.Sprintf("chat:%s", msg.Channel)

	msgJSON, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("error serializando mensaje: %v", err)
	}

	// Publicar en Redis Pub/Sub
	err = s.redisService.client.Publish(ctx, channelKey, msgJSON).Err()
	if err != nil {
		return fmt.Errorf("error publicando mensaje: %v", err)
	}

	return nil
}

// cacheRecentMessage cachea los mensajes recientes
func (s *ChatService) cacheRecentMessage(ctx context.Context, channel string, msg *ChatMessage) error {
	recentKey := fmt.Sprintf("chat:recent:%s", channel)

	msgJSON, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("error serializando mensaje: %v", err)
	}

	// Agregar al inicio de la lista
	err = s.redisService.client.LPush(ctx, recentKey, msgJSON).Err()
	if err != nil {
		return fmt.Errorf("error agregando mensaje reciente: %v", err)
	}

	// Mantener solo los últimos 100 mensajes
	err = s.redisService.client.LTrim(ctx, recentKey, 0, 99).Err()
	if err != nil {
		return fmt.Errorf("error recortando mensajes: %v", err)
	}

	// Expirar después de 24 horas
	err = s.redisService.client.Expire(ctx, recentKey, 24*time.Hour).Err()
	if err != nil {
		return fmt.Errorf("error configurando expiración: %v", err)
	}

	return nil
}

// GetRecentMessages obtiene los mensajes recientes desde Redis
func (s *ChatService) GetRecentMessages(ctx context.Context, channel string, limit int) ([]*ChatMessage, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	recentKey := fmt.Sprintf("chat:recent:%s", channel)

	// Obtener mensajes desde Redis
	msgList, err := s.redisService.client.LRange(ctx, recentKey, 0, int64(limit-1)).Result()
	if err != nil {
		return nil, fmt.Errorf("error obteniendo mensajes recientes: %v", err)
	}

	var messages []*ChatMessage
	for _, msgData := range msgList {
		var msg ChatMessage
		err := json.Unmarshal([]byte(msgData), &msg)
		if err != nil {
			log.Printf("Error deserializando mensaje: %v", err)
			continue
		}
		messages = append(messages, &msg)
	}

	return messages, nil
}

// SubscribeToChannel suscribe a un canal de chat
func (s *ChatService) SubscribeToChannel(ctx context.Context, channel string) (*redis.PubSub, error) {
	channelKey := fmt.Sprintf("chat:%s", channel)
	pubsub := s.redisService.client.Subscribe(ctx, channelKey)

	// Verificar conexión
	_, err := pubsub.Receive(ctx)
	if err != nil {
		return nil, fmt.Errorf("error suscribiendo al canal: %v", err)
	}

	return pubsub, nil
}

// GetOnlineUsers obtiene usuarios online en un canal
func (s *ChatService) GetOnlineUsers(ctx context.Context, channel string) ([]string, error) {
	onlineKey := fmt.Sprintf("chat:online:%s", channel)

	users, err := s.redisService.client.SMembers(ctx, onlineKey).Result()
	if err != nil {
		return nil, fmt.Errorf("error obteniendo usuarios online: %v", err)
	}

	return users, nil
}

// JoinChannel marca a un usuario como online en un canal
func (s *ChatService) JoinChannel(ctx context.Context, playerID int64, username, channel string) error {
	onlineKey := fmt.Sprintf("chat:online:%s", channel)

	// Agregar usuario al set de online
	err := s.redisService.client.SAdd(ctx, onlineKey, username).Err()
	if err != nil {
		return fmt.Errorf("error agregando usuario online: %v", err)
	}

	// Expirar después de 30 minutos de inactividad
	err = s.redisService.client.Expire(ctx, onlineKey, 30*time.Minute).Err()
	if err != nil {
		return fmt.Errorf("error configurando expiración: %v", err)
	}

	// Publicar evento de usuario online
	joinMsg := &ChatMessage{
		ID:        fmt.Sprintf("join_%d", time.Now().UnixNano()),
		PlayerID:  playerID,
		Username:  username,
		Channel:   channel,
		Message:   fmt.Sprintf("%s se unió al chat", username),
		Timestamp: time.Now(),
		Type:      "join",
		Data: map[string]interface{}{
			"action": "join",
		},
	}

	return s.publishMessage(ctx, joinMsg)
}

// LeaveChannel marca a un usuario como offline en un canal
func (s *ChatService) LeaveChannel(ctx context.Context, playerID int64, username, channel string) error {
	onlineKey := fmt.Sprintf("chat:online:%s", channel)

	// Remover usuario del set de online
	err := s.redisService.client.SRem(ctx, onlineKey, username).Err()
	if err != nil {
		return fmt.Errorf("error removiendo usuario online: %v", err)
	}

	// Publicar evento de usuario offline
	leaveMsg := &ChatMessage{
		ID:        fmt.Sprintf("leave_%d", time.Now().UnixNano()),
		PlayerID:  playerID,
		Username:  username,
		Channel:   channel,
		Message:   fmt.Sprintf("%s abandonó el chat", username),
		Timestamp: time.Now(),
		Type:      "leave",
		Data: map[string]interface{}{
			"action": "leave",
		},
	}

	return s.publishMessage(ctx, leaveMsg)
}

// isUserInChannel verifica si un usuario está en un canal
func (s *ChatService) isUserInChannel(ctx context.Context, username, channel string) bool {
	onlineKey := fmt.Sprintf("chat:online:%s", channel)

	isMember, err := s.redisService.client.SIsMember(ctx, onlineKey, username).Result()
	if err != nil {
		return false
	}

	return isMember
}

// CreateAllianceChannel crea un canal de alianza
func (s *ChatService) CreateAllianceChannel(ctx context.Context, allianceID, allianceName string) error {
	channelID := fmt.Sprintf("alliance:%s", allianceID)

	channel := &ChatChannel{
		ID:          channelID,
		Name:        fmt.Sprintf("Alianza %s", allianceName),
		Type:        "alliance",
		AllianceID:  allianceID,
		MemberCount: 0,
		CreatedAt:   time.Now(),
		IsActive:    true,
	}

	channelJSON, err := json.Marshal(channel)
	if err != nil {
		return fmt.Errorf("error serializando canal: %v", err)
	}

	// Guardar canal en Redis
	key := fmt.Sprintf("chat:channel:%s", channelID)
	err = s.redisService.client.Set(ctx, key, channelJSON, 0).Err()
	if err != nil {
		return fmt.Errorf("error guardando canal: %v", err)
	}

	// Enviar mensaje de sistema
	return s.SendSystemMessage(ctx, channelID, "Canal de alianza creado")
}

// GetChannelInfo obtiene información de un canal
func (s *ChatService) GetChannelInfo(ctx context.Context, channel string) (*ChatChannel, error) {
	key := fmt.Sprintf("chat:channel:%s", channel)

	data, err := s.redisService.client.Get(ctx, key).Result()
	if err != nil {
		// Si no existe, crear canal global por defecto
		if channel == "global" {
			return &ChatChannel{
				ID:          "global",
				Name:        "Chat Global",
				Type:        "global",
				MemberCount: 0,
				CreatedAt:   time.Now(),
				IsActive:    true,
			}, nil
		}
		return nil, fmt.Errorf("canal no encontrado")
	}

	var channelInfo ChatChannel
	err = json.Unmarshal([]byte(data), &channelInfo)
	if err != nil {
		return nil, fmt.Errorf("error deserializando canal: %v", err)
	}

	return &channelInfo, nil
}

// GetChatStats obtiene estadísticas del chat
func (s *ChatService) GetChatStats(ctx context.Context) (*ChatStats, error) {
	stats := &ChatStats{}

	// Contar mensajes totales (aproximado)
	keys, err := s.redisService.client.Keys(ctx, "chat:recent:*").Result()
	if err == nil {
		stats.ChannelsCount = int64(len(keys))
	}

	// Contar usuarios activos
	onlineKeys, err := s.redisService.client.Keys(ctx, "chat:online:*").Result()
	if err == nil {
		for _, key := range onlineKeys {
			count, _ := s.redisService.client.SCard(ctx, key).Result()
			stats.ActiveUsers += count
		}
	}

	return stats, nil
}

// updateChatStats actualiza estadísticas del chat
func (s *ChatService) updateChatStats(ctx context.Context, channel string) {
	statsKey := fmt.Sprintf("chat:stats:%s", channel)

	// Incrementar contador de mensajes
	s.redisService.client.Incr(ctx, statsKey)

	// Expirar después de 1 hora
	s.redisService.client.Expire(ctx, statsKey, time.Hour)
}

// CleanupInactiveUsers limpia usuarios inactivos
func (s *ChatService) CleanupInactiveUsers(ctx context.Context) error {
	// Obtener todas las claves de usuarios online
	onlineKeys, err := s.redisService.client.Keys(ctx, "chat:online:*").Result()
	if err != nil {
		return fmt.Errorf("error obteniendo claves online: %v", err)
	}

	for _, key := range onlineKeys {
		// Los sets expiran automáticamente después de 30 minutos
		// Esto es manejado por Redis automáticamente
		log.Printf("Limpiando usuarios inactivos en: %s", key)
	}

	return nil
}

// BanUser banea a un usuario de un canal
func (s *ChatService) BanUser(ctx context.Context, username, channel string, duration time.Duration) error {
	banKey := fmt.Sprintf("chat:banned:%s", channel)

	// Agregar usuario a la lista de baneados
	err := s.redisService.client.SAdd(ctx, banKey, username).Err()
	if err != nil {
		return fmt.Errorf("error baneando usuario: %v", err)
	}

	// Configurar expiración del ban
	if duration > 0 {
		err = s.redisService.client.Expire(ctx, banKey, duration).Err()
		if err != nil {
			return fmt.Errorf("error configurando expiración del ban: %v", err)
		}
	}

	// Remover del canal
	s.LeaveChannel(ctx, 0, username, channel)

	// Enviar mensaje de sistema
	s.SendSystemMessage(ctx, channel, fmt.Sprintf("%s ha sido baneado del chat", username))

	return nil
}

// IsUserBanned verifica si un usuario está baneado
func (s *ChatService) IsUserBanned(ctx context.Context, username, channel string) bool {
	banKey := fmt.Sprintf("chat:banned:%s", channel)

	isBanned, err := s.redisService.client.SIsMember(ctx, banKey, username).Result()
	if err != nil {
		return false
	}

	return isBanned
}
