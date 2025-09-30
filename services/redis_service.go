package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"server-backend/config"
	"server-backend/models"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// RedisService maneja todas las operaciones con Redis
type RedisService struct {
	client *redis.Client
	logger *zap.Logger
}

// NewRedisService crea una nueva instancia del servicio de Redis
func NewRedisService(cfg *config.Config, logger *zap.Logger) (*RedisService, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		PoolSize:     cfg.Redis.PoolSize,
		MinIdleConns: cfg.Redis.MinIdleConns,
		MaxRetries:   cfg.Redis.MaxRetries,
		DialTimeout:  cfg.Redis.DialTimeout,
		ReadTimeout:  cfg.Redis.ReadTimeout,
		WriteTimeout: cfg.Redis.WriteTimeout,
		PoolTimeout:  cfg.Redis.PoolTimeout,
	})

	// Verificar conexión
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("error conectando a Redis: %w", err)
	}

	logger.Info("Conectado a Redis exitosamente",
		zap.String("host", cfg.Redis.Host),
		zap.Int("port", cfg.Redis.Port),
		zap.Int("db", cfg.Redis.DB))

	return &RedisService{
		client: client,
		logger: logger,
	}, nil
}

// Close cierra la conexión con Redis
func (r *RedisService) Close() error {
	return r.client.Close()
}

// ========================================
// SISTEMA DE SESIONES DE USUARIO
// ========================================

// SessionData representa los datos de sesión de un usuario
type SessionData struct {
	UserID      string                 `json:"user_id"`
	Username    string                 `json:"username"`
	Role        string                 `json:"role"`
	IsOnline    bool                   `json:"is_online"`
	LastActive  time.Time              `json:"last_active"`
	WorldID     string                 `json:"world_id,omitempty"`
	VillageID   string                 `json:"village_id,omitempty"`
	Preferences map[string]interface{} `json:"preferences,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
}

// StoreUserSession almacena la sesión de un usuario
func (r *RedisService) StoreUserSession(userID string, sessionData *SessionData) error {
	ctx := context.Background()
	key := fmt.Sprintf("session:user:%s", userID)

	data, err := json.Marshal(sessionData)
	if err != nil {
		return fmt.Errorf("error serializando sesión: %w", err)
	}

	// Almacenar sesión por 24 horas
	err = r.client.SetEx(ctx, key, data, 24*time.Hour).Err()
	if err != nil {
		return fmt.Errorf("error almacenando sesión: %w", err)
	}

	r.logger.Debug("Sesión almacenada",
		zap.String("user_id", userID),
		zap.String("username", sessionData.Username))

	return nil
}

// GetUserSession obtiene la sesión de un usuario
func (r *RedisService) GetUserSession(userID string) (*SessionData, error) {
	ctx := context.Background()
	key := fmt.Sprintf("session:user:%s", userID)

	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("sesión no encontrada")
		}
		return nil, fmt.Errorf("error obteniendo sesión: %w", err)
	}

	var session SessionData
	if err := json.Unmarshal([]byte(data), &session); err != nil {
		return nil, fmt.Errorf("error deserializando sesión: %w", err)
	}

	return &session, nil
}

// DeleteUserSession elimina la sesión de un usuario
func (r *RedisService) DeleteUserSession(userID string) error {
	ctx := context.Background()
	key := fmt.Sprintf("session:user:%s", userID)

	err := r.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("error eliminando sesión: %w", err)
	}

	r.logger.Debug("Sesión eliminada", zap.String("user_id", userID))
	return nil
}

// ========================================
// SISTEMA DE USUARIOS ONLINE
// ========================================

// SetUserOnline marca un usuario como online
func (r *RedisService) SetUserOnline(userID string, username string) error {
	ctx := context.Background()
	key := fmt.Sprintf("user:online:%s", userID)

	onlineData := map[string]interface{}{
		"user_id":   userID,
		"username":  username,
		"online_at": time.Now().Unix(),
	}

	data, _ := json.Marshal(onlineData)

	// Marcar como online por 5 minutos
	err := r.client.SetEx(ctx, key, data, 5*time.Minute).Err()
	if err != nil {
		return fmt.Errorf("error marcando usuario online: %w", err)
	}

	// Agregar a la lista de usuarios online
	err = r.client.SAdd(ctx, "users:online", userID).Err()
	if err != nil {
		return fmt.Errorf("error agregando a lista online: %w", err)
	}

	r.logger.Debug("Usuario marcado como online",
		zap.String("user_id", userID),
		zap.String("username", username))

	return nil
}

// SetUserOffline marca un usuario como offline
func (r *RedisService) SetUserOffline(userID string) error {
	ctx := context.Background()
	key := fmt.Sprintf("user:online:%s", userID)

	// Eliminar estado online
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("error eliminando estado online: %w", err)
	}

	// Remover de la lista de usuarios online
	err = r.client.SRem(ctx, "users:online", userID).Err()
	if err != nil {
		return fmt.Errorf("error removiendo de lista online: %w", err)
	}

	r.logger.Debug("Usuario marcado como offline", zap.String("user_id", userID))
	return nil
}

// IsUserOnline verifica si un usuario está online
func (r *RedisService) IsUserOnline(userID string) (bool, error) {
	ctx := context.Background()
	key := fmt.Sprintf("user:online:%s", userID)

	exists, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("error verificando estado online: %w", err)
	}

	return exists > 0, nil
}

// GetOnlineUsers obtiene la lista de usuarios online
func (r *RedisService) GetOnlineUsers() ([]string, error) {
	ctx := context.Background()

	userIDs, err := r.client.SMembers(ctx, "users:online").Result()
	if err != nil {
		return nil, fmt.Errorf("error obteniendo usuarios online: %w", err)
	}

	// Filtrar usuarios que realmente están online (no expirados)
	var onlineUsers []string
	for _, userID := range userIDs {
		key := fmt.Sprintf("user:online:%s", userID)
		exists, _ := r.client.Exists(ctx, key).Result()
		if exists > 0 {
			onlineUsers = append(onlineUsers, userID)
		}
	}

	return onlineUsers, nil
}

// ========================================
// SISTEMA DE RECURSOS EN TIEMPO REAL
// ========================================

// StorePlayerResources almacena los recursos de un jugador
func (r *RedisService) StorePlayerResources(playerID string, resources *models.ResourceData) error {
	ctx := context.Background()
	key := fmt.Sprintf("player:resources:%s", playerID)

	data, err := json.Marshal(resources)
	if err != nil {
		return fmt.Errorf("error serializando recursos: %w", err)
	}

	// Almacenar por 10 minutos
	err = r.client.SetEx(ctx, key, data, 10*time.Minute).Err()
	if err != nil {
		return fmt.Errorf("error almacenando recursos: %w", err)
	}

	return nil
}

// GetPlayerResources obtiene los recursos de un jugador
func (r *RedisService) GetPlayerResources(playerID string) (*models.ResourceData, error) {
	ctx := context.Background()
	key := fmt.Sprintf("player:resources:%s", playerID)

	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("recursos no encontrados")
		}
		return nil, fmt.Errorf("error obteniendo recursos: %w", err)
	}

	var resources models.ResourceData
	if err := json.Unmarshal([]byte(data), &resources); err != nil {
		return nil, fmt.Errorf("error deserializando recursos: %w", err)
	}

	return &resources, nil
}

// ========================================
// SISTEMA DE INVESTIGACIÓN EN TIEMPO REAL
// ========================================

// ResearchData representa el estado de investigación de un jugador

// StoreResearchProgress almacena el progreso de investigación
func (r *RedisService) StoreResearchProgress(playerID string, research *models.ResearchData) error {
	ctx := context.Background()
	key := fmt.Sprintf("research:active:%s", playerID)

	data, err := json.Marshal(research)
	if err != nil {
		return fmt.Errorf("error serializando investigación: %w", err)
	}

	// Calcular tiempo restante
	remaining := time.Until(research.EndsAt)
	if remaining > 0 {
		err = r.client.SetEx(ctx, key, data, remaining).Err()
	} else {
		err = r.client.Set(ctx, key, data, 1*time.Hour).Err()
	}

	if err != nil {
		return fmt.Errorf("error almacenando investigación: %w", err)
	}

	return nil
}

// GetResearchProgress obtiene el progreso de investigación
func (r *RedisService) GetResearchProgress(playerID string) (*models.ResearchData, error) {
	ctx := context.Background()
	key := fmt.Sprintf("research:active:%s", playerID)

	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("investigación no encontrada")
		}
		return nil, fmt.Errorf("error obteniendo investigación: %w", err)
	}

	var research models.ResearchData
	if err := json.Unmarshal([]byte(data), &research); err != nil {
		return nil, fmt.Errorf("error deserializando investigación: %w", err)
	}

	return &research, nil
}

// ========================================
// SISTEMA DE NOTIFICACIONES
// ========================================

// Notification representa una notificación
type Notification struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Title     string                 `json:"title"`
	Message   string                 `json:"message"`
	Data      map[string]interface{} `json:"data,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
	Read      bool                   `json:"read"`
}

// AddNotification agrega una notificación a un usuario
func (r *RedisService) AddNotification(userID string, notification *models.Notification) error {
	ctx := context.Background()
	key := fmt.Sprintf("notifications:user:%s", userID)

	data, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("error serializando notificación: %w", err)
	}

	// Agregar al inicio de la lista (más reciente primero)
	err = r.client.LPush(ctx, key, data).Err()
	if err != nil {
		return fmt.Errorf("error agregando notificación: %w", err)
	}

	// Mantener solo las últimas 50 notificaciones
	err = r.client.LTrim(ctx, key, 0, 49).Err()
	if err != nil {
		return fmt.Errorf("error recortando notificaciones: %w", err)
	}

	// Expirar después de 7 días
	err = r.client.Expire(ctx, key, 7*24*time.Hour).Err()
	if err != nil {
		return fmt.Errorf("error configurando expiración: %w", err)
	}

	return nil
}

// GetNotifications obtiene las notificaciones de un usuario
func (r *RedisService) GetNotifications(userID string, limit int) ([]*models.Notification, error) {
	ctx := context.Background()
	key := fmt.Sprintf("notifications:user:%s", userID)

	if limit <= 0 {
		limit = 20
	}

	// Obtener las notificaciones más recientes
	dataList, err := r.client.LRange(ctx, key, 0, int64(limit-1)).Result()
	if err != nil {
		return nil, fmt.Errorf("error obteniendo notificaciones: %w", err)
	}

	var notifications []*models.Notification
	for _, data := range dataList {
		var notification models.Notification
		if err := json.Unmarshal([]byte(data), &notification); err != nil {
			r.logger.Warn("Error deserializando notificación", zap.Error(err))
			continue
		}
		notifications = append(notifications, &notification)
	}

	return notifications, nil
}

// MarkNotificationAsRead marca una notificación como leída
func (r *RedisService) MarkNotificationAsRead(userID string, notificationID string) error {
	ctx := context.Background()
	key := fmt.Sprintf("notifications:user:%s", userID)

	// Obtener todas las notificaciones
	dataList, err := r.client.LRange(ctx, key, 0, -1).Result()
	if err != nil {
		return fmt.Errorf("error obteniendo notificaciones: %w", err)
	}

	// Buscar y actualizar la notificación específica
	for i, data := range dataList {
		var notification Notification
		if err := json.Unmarshal([]byte(data), &notification); err != nil {
			continue
		}

		if notification.ID == notificationID {
			notification.Read = true
			updatedData, _ := json.Marshal(notification)

			// Reemplazar en la lista
			err = r.client.LSet(ctx, key, int64(i), updatedData).Err()
			if err != nil {
				return fmt.Errorf("error actualizando notificación: %w", err)
			}
			break
		}
	}

	return nil
}

// ========================================
// SISTEMA DE CACHE GENERAL
// ========================================

// SetCache almacena datos en cache
func (r *RedisService) SetCache(key string, data interface{}, expiration time.Duration) error {
	ctx := context.Background()

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("error serializando datos: %w", err)
	}

	err = r.client.SetEx(ctx, key, jsonData, expiration).Err()
	if err != nil {
		return fmt.Errorf("error almacenando en cache: %w", err)
	}

	return nil
}

// GetCache obtiene datos del cache
func (r *RedisService) GetCache(key string, target interface{}) error {
	ctx := context.Background()

	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("dato no encontrado en cache")
		}
		return fmt.Errorf("error obteniendo de cache: %w", err)
	}

	if err := json.Unmarshal([]byte(data), target); err != nil {
		return fmt.Errorf("error deserializando datos: %w", err)
	}

	return nil
}

// DeleteCache elimina datos del cache
func (r *RedisService) DeleteCache(key string) error {
	ctx := context.Background()

	err := r.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("error eliminando de cache: %w", err)
	}

	return nil
}

// ========================================
// SISTEMA DE ESTADÍSTICAS
// ========================================

// IncrementCounter incrementa un contador
func (r *RedisService) IncrementCounter(key string) error {
	ctx := context.Background()

	err := r.client.Incr(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("error incrementando contador: %w", err)
	}

	return nil
}

// GetCounter obtiene el valor de un contador
func (r *RedisService) GetCounter(key string) (int64, error) {
	ctx := context.Background()

	value, err := r.client.Get(ctx, key).Int64()
	if err != nil {
		if err == redis.Nil {
			return 0, nil
		}
		return 0, fmt.Errorf("error obteniendo contador: %w", err)
	}

	return value, nil
}

// SetCounter establece el valor de un contador
func (r *RedisService) SetCounter(key string, value int64, expiration time.Duration) error {
	ctx := context.Background()

	err := r.client.SetEx(ctx, key, value, expiration).Err()
	if err != nil {
		return fmt.Errorf("error estableciendo contador: %w", err)
	}

	return nil
}

// ========================================
// UTILIDADES
// ========================================

// Ping verifica la conexión con Redis
func (r *RedisService) Ping() error {
	ctx := context.Background()
	return r.client.Ping(ctx).Err()
}

// FlushAll limpia toda la base de datos (solo para desarrollo)
func (r *RedisService) FlushAll() error {
	ctx := context.Background()
	return r.client.FlushAll(ctx).Err()
}

// GetStats obtiene estadísticas de Redis
func (r *RedisService) GetStats() (map[string]interface{}, error) {
	ctx := context.Background()

	info, err := r.client.Info(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("error obteniendo estadísticas: %w", err)
	}

	stats := map[string]interface{}{
		"info":              info,
		"db_size":           r.client.DBSize(ctx).Val(),
		"connected_clients": 0, // Se puede extraer del info
	}

	return stats, nil
}

// ========================================
// SISTEMA DE COLAS
// ========================================

// AddToQueue agrega un item a una cola en Redis
func (r *RedisService) AddToQueue(ctx context.Context, queueKey string, item interface{}) error {
	itemJSON, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("error serializando item: %w", err)
	}

	err = r.client.RPush(ctx, queueKey, itemJSON).Err()
	if err != nil {
		return fmt.Errorf("error agregando a cola: %w", err)
	}
	return nil
}

// GetQueue obtiene todos los items de una cola
func (r *RedisService) GetQueue(ctx context.Context, queueKey string) ([]string, error) {
	items, err := r.client.LRange(ctx, queueKey, 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("error obteniendo cola: %w", err)
	}
	return items, nil
}

// RemoveFromQueue remueve un item específico de una cola
func (r *RedisService) RemoveFromQueue(ctx context.Context, queueKey string, itemID int64) error {
	// Obtener todos los items
	items, err := r.client.LRange(ctx, queueKey, 0, -1).Result()
	if err != nil {
		return fmt.Errorf("error obteniendo items de cola: %w", err)
	}

	// Encontrar y remover el item específico
	for _, item := range items {
		var itemData map[string]interface{}
		err := json.Unmarshal([]byte(item), &itemData)
		if err != nil {
			continue
		}

		if id, ok := itemData["id"].(float64); ok && int64(id) == itemID {
			err = r.client.LRem(ctx, queueKey, 1, item).Err()
			if err != nil {
				return fmt.Errorf("error removiendo item de cola: %w", err)
			}
			break
		}
	}

	return nil
}

// GetKeys obtiene todas las claves que coinciden con un patrón
func (r *RedisService) GetKeys(ctx context.Context, pattern string) ([]string, error) {
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("error obteniendo claves: %w", err)
	}
	return keys, nil
}

// ========================================
// SISTEMA DE PUB/SUB PARA WEBSOCKET
// ========================================

// Publish publica un mensaje en un canal Redis
func (r *RedisService) Publish(channel string, message interface{}) error {
	ctx := context.Background()

	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("error serializando mensaje: %w", err)
	}

	err = r.client.Publish(ctx, channel, data).Err()
	if err != nil {
		return fmt.Errorf("error publicando mensaje: %w", err)
	}

	return nil
}

// Subscribe se suscribe a un canal Redis (retorna un pubsub para compatibilidad)
func (r *RedisService) Subscribe(channel string) interface{} {
	ctx := context.Background()
	return r.client.Subscribe(ctx, channel)
}
