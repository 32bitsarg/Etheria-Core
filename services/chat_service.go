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
	"server-backend/sync"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type ChatService struct {
	chatRepo     *repository.ChatRepository
	redisService *RedisService
	syncManager  sync.SyncManager
	logger       *zap.Logger
}

type ChatMessage struct {
	ID        string                 `json:"id"`
	PlayerID  uuid.UUID              `json:"player_id"`
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
	WorldID     string    `json:"world_id,omitempty"`
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

// NewChatService crea una nueva instancia del servicio de chat
func NewChatService(chatRepo *repository.ChatRepository, redisService *RedisService, logger *zap.Logger) *ChatService {
	// Crear adaptadores para el paquete sync
	chatRepoAdapter := &ChatRepositoryAdapter{chatRepo: chatRepo}
	redisClientAdapter := &RedisClientAdapter{redisService: redisService}

	// Crear SyncManager completo usando el paquete sync
	syncManager := sync.NewSyncManager(chatRepoAdapter, redisClientAdapter, logger, nil)

	return &ChatService{
		chatRepo:     chatRepo,
		redisService: redisService,
		syncManager:  syncManager,
		logger:       logger,
	}
}

// ChatRepositoryAdapter adaptador para el repositorio de chat
type ChatRepositoryAdapter struct {
	chatRepo *repository.ChatRepository
}

func (a *ChatRepositoryAdapter) GetAllChannels() ([]*sync.ChatChannel, error) {
	channels, err := a.chatRepo.GetAllChannels()
	if err != nil {
		return nil, err
	}

	var syncChannels []*sync.ChatChannel
	for _, channel := range channels {
		syncChannel := &sync.ChatChannel{
			ID:          channel.ID.String(),
			Name:        channel.Name,
			Type:        channel.Type,
			MemberCount: 0, // Se calculará dinámicamente
			CreatedAt:   channel.CreatedAt,
			IsActive:    channel.IsActive,
		}

		if channel.WorldID != nil {
			worldIDStr := channel.WorldID.String()
			syncChannel.WorldID = &worldIDStr
		}

		// No hay AllianceID en models.ChatChannel, se puede agregar después
		syncChannels = append(syncChannels, syncChannel)
	}

	return syncChannels, nil
}

func (a *ChatRepositoryAdapter) GetChannelByID(channelID string) (*sync.ChatChannel, error) {
	channel, err := a.chatRepo.GetChannelByID(channelID)
	if err != nil {
		return nil, err
	}

	syncChannel := &sync.ChatChannel{
		ID:          channel.ID.String(),
		Name:        channel.Name,
		Type:        channel.Type,
		MemberCount: 0,
		CreatedAt:   channel.CreatedAt,
		IsActive:    channel.IsActive,
	}

	if channel.WorldID != nil {
		worldIDStr := channel.WorldID.String()
		syncChannel.WorldID = &worldIDStr
	}

	return syncChannel, nil
}

func (a *ChatRepositoryAdapter) GetChannelByName(name string) (*sync.ChatChannel, error) {
	channel, err := a.chatRepo.GetChannelByName(name)
	if err != nil {
		return nil, err
	}

	syncChannel := &sync.ChatChannel{
		ID:          channel.ID.String(),
		Name:        channel.Name,
		Type:        channel.Type,
		MemberCount: 0,
		CreatedAt:   channel.CreatedAt,
		IsActive:    channel.IsActive,
	}

	if channel.WorldID != nil {
		worldIDStr := channel.WorldID.String()
		syncChannel.WorldID = &worldIDStr
	}

	return syncChannel, nil
}

func (a *ChatRepositoryAdapter) CreateChannel(channel *sync.ChatChannel) error {
	// Convertir sync.ChatChannel a models.ChatChannel
	modelChannel := &models.ChatChannel{
		Name:      channel.Name,
		Type:      channel.Type,
		CreatedAt: channel.CreatedAt,
		IsActive:  channel.IsActive,
	}

	if channel.WorldID != nil {
		worldID, err := uuid.Parse(*channel.WorldID)
		if err != nil {
			return err
		}
		modelChannel.WorldID = &worldID
	}

	return a.chatRepo.CreateChannel(modelChannel)
}

func (a *ChatRepositoryAdapter) UpdateChannel(channel *sync.ChatChannel) error {
	// Convertir sync.ChatChannel a models.ChatChannel
	modelChannel := &models.ChatChannel{
		Name:      channel.Name,
		Type:      channel.Type,
		CreatedAt: channel.CreatedAt,
		IsActive:  channel.IsActive,
	}

	if channel.WorldID != nil {
		worldID, err := uuid.Parse(*channel.WorldID)
		if err != nil {
			return err
		}
		modelChannel.WorldID = &worldID
	}

	return a.chatRepo.UpdateChannel(modelChannel)
}

func (a *ChatRepositoryAdapter) DeleteChannel(channelID string) error {
	return a.chatRepo.DeleteChannel(channelID)
}

// RedisClientAdapter adaptador para el cliente de Redis
type RedisClientAdapter struct {
	redisService *RedisService
}

func (a *RedisClientAdapter) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	return a.redisService.GetClient().Set(ctx, key, value, expiration)
}

func (a *RedisClientAdapter) Get(ctx context.Context, key string) *redis.StringCmd {
	return a.redisService.GetClient().Get(ctx, key)
}

func (a *RedisClientAdapter) Pipeline() redis.Pipeliner {
	return a.redisService.GetClient().Pipeline()
}

func (a *RedisClientAdapter) Close() error {
	return a.redisService.Close()
}

// StartSyncManager inicia el gestor de sincronización
func (s *ChatService) StartSyncManager(ctx context.Context) error {
	return s.syncManager.Start(ctx)
}

// StopSyncManager detiene el gestor de sincronización
func (s *ChatService) StopSyncManager() error {
	return s.syncManager.Stop()
}

// GetSyncStatus obtiene el estado del gestor de sincronización
func (s *ChatService) GetSyncStatus() sync.SyncStatus {
	return s.syncManager.GetStatus()
}

// GetSyncMetrics obtiene las métricas del gestor de sincronización
func (s *ChatService) GetSyncMetrics() sync.SyncMetrics {
	return s.syncManager.GetMetrics()
}

// IsSyncHealthy verifica si el gestor de sincronización está saludable
func (s *ChatService) IsSyncHealthy() bool {
	return s.syncManager.IsHealthy()
}

// SendMessage envía un mensaje y lo almacena en Redis para tiempo real
func (s *ChatService) SendMessage(ctx context.Context, playerID uuid.UUID, username, channel, message string) error {
	// Validar mensaje
	if strings.TrimSpace(message) == "" {
		return fmt.Errorf("mensaje no puede estar vacío")
	}

	if len(message) > 500 {
		return fmt.Errorf("mensaje demasiado largo (máximo 500 caracteres)")
	}

	// Si Redis no está disponible, solo guardar en base de datos
	if s.redisService == nil {
		log.Printf("Redis no disponible - guardando mensaje solo en base de datos")
		return s.saveMessageToDatabase(ctx, playerID, username, channel, message)
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
	// Buscar el canal en la base de datos (case insensitive)
	channelInfo, err := s.chatRepo.GetChannelByName(channel)
	if err != nil {
		return fmt.Errorf("canal '%s' no encontrado: %v", channel, err)
	}

	dbMessage := &models.ChatMessage{
		PlayerID:  playerID,       // Usar el playerID real
		ChannelID: channelInfo.ID, // Usar el ID real del canal
		Username:  username,
		Message:   message,
		Type:      "text",
		CreatedAt: time.Now(),
	}

	err = s.chatRepo.SaveMessage(dbMessage)
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
		PlayerID:  uuid.Nil, // UUID cero para mensajes del sistema
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

	// Validar que Redis esté disponible
	if s.redisService == nil {
		log.Printf("Redis no disponible - retornando mensajes vacíos para canal: %s", channel)
		return []*ChatMessage{}, nil
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
	// Validar que Redis esté disponible
	if s.redisService == nil {
		return nil, fmt.Errorf("Redis no disponible - suscripción no disponible")
	}

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
	// Validar que Redis esté disponible
	if s.redisService == nil {
		log.Printf("Redis no disponible - retornando lista vacía de usuarios online")
		return []string{}, nil
	}

	onlineKey := fmt.Sprintf("chat:online:%s", channel)

	users, err := s.redisService.client.SMembers(ctx, onlineKey).Result()
	if err != nil {
		return nil, fmt.Errorf("error obteniendo usuarios online: %v", err)
	}

	return users, nil
}

// JoinChannel marca a un usuario como online en un canal
func (s *ChatService) JoinChannel(ctx context.Context, playerID uuid.UUID, username, channel string) error {
	// Si Redis no está disponible, solo loggear la acción
	if s.redisService == nil {
		log.Printf("Usuario %s se unió al canal %s (Redis no disponible)", username, channel)
		return nil
	}

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
func (s *ChatService) LeaveChannel(ctx context.Context, playerID uuid.UUID, username, channel string) error {
	// Si Redis no está disponible, solo loggear la acción
	if s.redisService == nil {
		log.Printf("Usuario %s abandonó el canal %s (Redis no disponible)", username, channel)
		return nil
	}

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

// SyncChannelsToRedis sincroniza todos los canales de BD a Redis usando el nuevo SyncManager
func (s *ChatService) SyncChannelsToRedis(ctx context.Context) error {
	return s.syncManager.SyncAll(ctx)
}

// syncChannelToRedis sincroniza un canal específico a Redis usando el nuevo SyncManager
func (s *ChatService) syncChannelToRedis(ctx context.Context, channel *models.ChatChannel) error {
	return s.syncManager.SyncChannel(ctx, channel.ID.String())
}

// GetAllChannels obtiene todos los canales desde BD y los sincroniza con Redis
func (s *ChatService) GetAllChannels(ctx context.Context) ([]*ChatChannel, error) {
	// Obtener canales de BD
	dbChannels, err := s.chatRepo.GetAllChannels()
	if err != nil {
		return nil, fmt.Errorf("error obteniendo canales de BD: %v", err)
	}

	// Convertir a estructura de Redis
	var channels []*ChatChannel
	for _, dbChannel := range dbChannels {
		redisChannel := &ChatChannel{
			ID:          dbChannel.ID.String(),
			Name:        dbChannel.Name,
			Type:        dbChannel.Type,
			MemberCount: 0, // Se calculará dinámicamente
			CreatedAt:   dbChannel.CreatedAt,
			IsActive:    dbChannel.IsActive,
		}

		if dbChannel.WorldID != nil {
			redisChannel.WorldID = dbChannel.WorldID.String()
		}

		if dbChannel.Type == "alliance" {
			redisChannel.AllianceID = "1" // Por ahora
		}

		channels = append(channels, redisChannel)
	}

	// Sincronizar a Redis si está disponible usando el nuevo SyncManager
	if s.syncManager != nil {
		go func() {
			if err := s.syncManager.SyncAll(ctx); err != nil {
				s.logger.Error("Error en sincronización automática", zap.Error(err))
			}
		}()
	}

	return channels, nil
}

// GetChannelInfo obtiene información de un canal desde BD o Redis
func (s *ChatService) GetChannelInfo(ctx context.Context, channel string) (*ChatChannel, error) {
	// Primero intentar obtener desde Redis
	if s.redisService != nil {
		key := fmt.Sprintf("chat:channel:%s", channel)
		data, err := s.redisService.client.Get(ctx, key).Result()
		if err == nil {
			var channelInfo ChatChannel
			err = json.Unmarshal([]byte(data), &channelInfo)
			if err == nil {
				return &channelInfo, nil
			}
		}
	}

	// Si no está en Redis o Redis no está disponible, obtener desde BD
	dbChannel, err := s.chatRepo.GetChannelByID(channel)
	if err != nil {
		// Intentar por nombre si no se encuentra por ID
		dbChannel, err = s.chatRepo.GetChannelByName(channel)
		if err != nil {
			return nil, fmt.Errorf("canal no encontrado: %s", channel)
		}
	}

	// Convertir a estructura de Redis
	redisChannel := &ChatChannel{
		ID:          dbChannel.ID.String(),
		Name:        dbChannel.Name,
		Type:        dbChannel.Type,
		MemberCount: 0,
		CreatedAt:   dbChannel.CreatedAt,
		IsActive:    dbChannel.IsActive,
	}

	if dbChannel.WorldID != nil {
		redisChannel.WorldID = dbChannel.WorldID.String()
	}

	if dbChannel.Type == "alliance" {
		redisChannel.AllianceID = "1"
	}

	// Sincronizar a Redis si está disponible usando el nuevo SyncManager
	if s.syncManager != nil {
		go func() {
			if err := s.syncManager.SyncChannel(ctx, dbChannel.ID.String()); err != nil {
				s.logger.Error("Error en sincronización automática de canal",
					zap.String("channel_id", dbChannel.ID.String()),
					zap.Error(err))
			}
		}()
	}

	return redisChannel, nil
}

// GetChatStats obtiene estadísticas del chat
func (s *ChatService) GetChatStats(ctx context.Context) (*ChatStats, error) {
	stats := &ChatStats{}

	// Si Redis no está disponible, retornar estadísticas vacías
	if s.redisService == nil {
		log.Printf("Redis no disponible - retornando estadísticas vacías")
		return stats, nil
	}

	// Contar mensajes totales (aproximado)
	keys, err := s.redisService.GetClient().Keys(ctx, "chat:recent:*").Result()
	if err == nil {
		stats.ChannelsCount = int64(len(keys))
	}

	// Contar usuarios activos
	onlineKeys, err := s.redisService.GetClient().Keys(ctx, "chat:online:*").Result()
	if err == nil {
		for _, key := range onlineKeys {
			count, _ := s.redisService.GetClient().SCard(ctx, key).Result()
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
	// Si Redis no está disponible, solo loggear la acción
	if s.redisService == nil {
		log.Printf("Usuario %s baneado del canal %s por %v (Redis no disponible)", username, channel, duration)
		return nil
	}

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
	s.LeaveChannel(ctx, uuid.Nil, username, channel)

	// Enviar mensaje de sistema
	s.SendSystemMessage(ctx, channel, fmt.Sprintf("%s ha sido baneado del chat", username))

	return nil
}

// IsUserBanned verifica si un usuario está baneado
func (s *ChatService) IsUserBanned(ctx context.Context, username, channel string) bool {
	// Si Redis no está disponible, no hay baneos
	if s.redisService == nil {
		return false
	}

	banKey := fmt.Sprintf("chat:banned:%s", channel)

	isBanned, err := s.redisService.client.SIsMember(ctx, banKey, username).Result()
	if err != nil {
		return false
	}

	return isBanned
}

// saveMessageToDatabase guarda un mensaje solo en la base de datos (fallback cuando Redis no está disponible)
func (s *ChatService) saveMessageToDatabase(ctx context.Context, playerID uuid.UUID, username, channel, message string) error {
	// Buscar el canal en la base de datos (case insensitive)
	channelInfo, err := s.chatRepo.GetChannelByName(channel)
	if err != nil {
		return fmt.Errorf("canal '%s' no encontrado: %v", channel, err)
	}

	// Crear mensaje para la base de datos
	dbMessage := &models.ChatMessage{
		PlayerID:  playerID,       // Usar el playerID real
		ChannelID: channelInfo.ID, // Usar el ID real del canal
		Username:  username,
		Message:   message,
		Type:      "text",
		CreatedAt: time.Now(),
	}

	err = s.chatRepo.SaveMessage(dbMessage)
	if err != nil {
		return fmt.Errorf("error guardando mensaje en BD: %v", err)
	}

	log.Printf("Mensaje guardado en base de datos: %s en canal %s (ID: %s)", username, channel, channelInfo.ID)
	return nil
}
