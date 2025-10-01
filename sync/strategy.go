package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// ImmediateSyncStrategy estrategia de sincronización inmediata
type ImmediateSyncStrategy struct {
	chatRepo     ChatRepository
	redisClient  RedisClient
	retryManager *RetryManager
	config       *SyncConfig
}

// NewImmediateSyncStrategy crea una nueva estrategia de sincronización inmediata
func NewImmediateSyncStrategy(
	chatRepo ChatRepository,
	redisClient RedisClient,
	config *SyncConfig,
) *ImmediateSyncStrategy {
	retryConfig := RetryConfig{
		MaxRetries:    config.MaxRetries,
		RetryDelay:    config.RetryDelay,
		MaxRetryDelay: config.MaxRetryDelay,
		BackoffFactor: 2.0,
	}

	return &ImmediateSyncStrategy{
		chatRepo:     chatRepo,
		redisClient:  redisClient,
		retryManager: NewRetryManager(retryConfig),
		config:       config,
	}
}

// Sync sincroniza múltiples canales
func (s *ImmediateSyncStrategy) Sync(ctx context.Context, channels []*ChatChannel) error {
	if len(channels) == 0 {
		return nil
	}

	// Sincronizar cada canal individualmente
	for _, channel := range channels {
		if err := s.SyncChannel(ctx, channel); err != nil {
			// Log error pero continuar con los demás canales
			continue
		}
	}

	return nil
}

// SyncChannel sincroniza un canal específico
func (s *ImmediateSyncStrategy) SyncChannel(ctx context.Context, channel *ChatChannel) error {
	return s.retryManager.ExecuteWithRetry(ctx, "sync_channel", func() error {
		return s.syncChannelToRedis(ctx, channel)
	})
}

// syncChannelToRedis sincroniza un canal específico a Redis
func (s *ImmediateSyncStrategy) syncChannelToRedis(ctx context.Context, channel *ChatChannel) error {
	if s.redisClient == nil {
		return ErrRedisUnavailable("sync_channel")
	}

	// Convertir modelo de BD a estructura de Redis
	redisChannel := &ChatChannel{
		ID:          channel.ID,
		Name:        channel.Name,
		Type:        channel.Type,
		MemberCount: 0, // Se calculará dinámicamente
		CreatedAt:   channel.CreatedAt,
		IsActive:    channel.IsActive,
	}

	// Agregar campos específicos según el tipo
	if channel.WorldID != nil {
		redisChannel.WorldID = channel.WorldID
	}

	if channel.Type == "alliance" {
		// Extraer alliance_id del nombre o usar un ID por defecto
		redisChannel.AllianceID = stringPtr("1") // Por ahora, se puede mejorar
	}

	// Serializar canal
	channelJSON, err := json.Marshal(redisChannel)
	if err != nil {
		return ErrSyncFailed("serialize_channel", err)
	}

	// Guardar en Redis
	key := fmt.Sprintf("chat:channel:%s", redisChannel.ID)
	err = s.redisClient.Set(ctx, key, channelJSON, 0).Err()
	if err != nil {
		return ErrRedisConnection("set_channel", err)
	}

	return nil
}

// stringPtr helper para crear un puntero a string
func stringPtr(s string) *string {
	return &s
}

// ShouldRetry determina si un error debe ser reintentado
func (s *ImmediateSyncStrategy) ShouldRetry(err error) bool {
	return s.retryManager.ShouldRetry(err)
}

// GetRetryDelay obtiene el delay para un intento específico
func (s *ImmediateSyncStrategy) GetRetryDelay(attempt int) time.Duration {
	return s.retryManager.GetRetryDelay(attempt)
}

// GetMaxRetries obtiene el número máximo de reintentos
func (s *ImmediateSyncStrategy) GetMaxRetries() int {
	return s.retryManager.GetMaxRetries()
}

// GetName obtiene el nombre de la estrategia
func (s *ImmediateSyncStrategy) GetName() string {
	return "immediate"
}

// GetDescription obtiene la descripción de la estrategia
func (s *ImmediateSyncStrategy) GetDescription() string {
	return "Sincronización inmediata de canales individuales"
}

// BatchSyncStrategy estrategia de sincronización por lotes
type BatchSyncStrategy struct {
	chatRepo     ChatRepository
	redisClient  RedisClient
	retryManager *RetryManager
	config       *SyncConfig
}

// NewBatchSyncStrategy crea una nueva estrategia de sincronización por lotes
func NewBatchSyncStrategy(
	chatRepo ChatRepository,
	redisClient RedisClient,
	config *SyncConfig,
) *BatchSyncStrategy {
	retryConfig := RetryConfig{
		MaxRetries:    config.MaxRetries,
		RetryDelay:    config.RetryDelay,
		MaxRetryDelay: config.MaxRetryDelay,
		BackoffFactor: 2.0,
	}

	return &BatchSyncStrategy{
		chatRepo:     chatRepo,
		redisClient:  redisClient,
		retryManager: NewRetryManager(retryConfig),
		config:       config,
	}
}

// Sync sincroniza múltiples canales en lotes
func (s *BatchSyncStrategy) Sync(ctx context.Context, channels []*ChatChannel) error {
	if len(channels) == 0 {
		return nil
	}

	// Dividir en lotes
	batches := s.createBatches(channels, s.config.BatchSize)

	// Sincronizar cada lote
	for _, batch := range batches {
		if err := s.syncBatch(ctx, batch); err != nil {
			// Log error pero continuar con los demás lotes
			continue
		}
	}

	return nil
}

// SyncChannel sincroniza un canal específico (fallback a inmediato)
func (s *BatchSyncStrategy) SyncChannel(ctx context.Context, channel *ChatChannel) error {
	return s.retryManager.ExecuteWithRetry(ctx, "sync_channel", func() error {
		return s.syncChannelToRedis(ctx, channel)
	})
}

// createBatches crea lotes de canales
func (s *BatchSyncStrategy) createBatches(channels []*ChatChannel, batchSize int) [][]*ChatChannel {
	var batches [][]*ChatChannel

	for i := 0; i < len(channels); i += batchSize {
		end := i + batchSize
		if end > len(channels) {
			end = len(channels)
		}
		batches = append(batches, channels[i:end])
	}

	return batches
}

// syncBatch sincroniza un lote de canales
func (s *BatchSyncStrategy) syncBatch(ctx context.Context, batch []*ChatChannel) error {
	return s.retryManager.ExecuteWithRetry(ctx, "sync_batch", func() error {
		return s.syncBatchToRedis(ctx, batch)
	})
}

// syncBatchToRedis sincroniza un lote de canales a Redis
func (s *BatchSyncStrategy) syncBatchToRedis(ctx context.Context, batch []*ChatChannel) error {
	if s.redisClient == nil {
		return ErrRedisUnavailable("sync_batch")
	}

	// Usar pipeline de Redis para mejor rendimiento
	pipe := s.redisClient.Pipeline()

	for _, channel := range batch {
		// Convertir modelo de BD a estructura de Redis
		redisChannel := &ChatChannel{
			ID:          channel.ID,
			Name:        channel.Name,
			Type:        channel.Type,
			MemberCount: 0,
			CreatedAt:   channel.CreatedAt,
			IsActive:    channel.IsActive,
		}

		if channel.WorldID != nil {
			redisChannel.WorldID = channel.WorldID
		}

		if channel.Type == "alliance" {
			redisChannel.AllianceID = stringPtr("1")
		}

		// Serializar canal
		channelJSON, err := json.Marshal(redisChannel)
		if err != nil {
			return ErrSyncFailed("serialize_channel", err)
		}

		// Agregar al pipeline
		key := fmt.Sprintf("chat:channel:%s", redisChannel.ID)
		pipe.Set(ctx, key, channelJSON, 0)
	}

	// Ejecutar pipeline
	_, err := pipe.Exec(ctx)
	if err != nil {
		return ErrRedisConnection("exec_pipeline", err)
	}

	return nil
}

// syncChannelToRedis sincroniza un canal específico a Redis
func (s *BatchSyncStrategy) syncChannelToRedis(ctx context.Context, channel *ChatChannel) error {
	if s.redisClient == nil {
		return ErrRedisUnavailable("sync_channel")
	}

	// Convertir modelo de BD a estructura de Redis
	redisChannel := &ChatChannel{
		ID:          channel.ID,
		Name:        channel.Name,
		Type:        channel.Type,
		MemberCount: 0,
		CreatedAt:   channel.CreatedAt,
		IsActive:    channel.IsActive,
	}

	if channel.WorldID != nil {
		redisChannel.WorldID = channel.WorldID
	}

	if channel.Type == "alliance" {
		redisChannel.AllianceID = stringPtr("1")
	}

	// Serializar canal
	channelJSON, err := json.Marshal(redisChannel)
	if err != nil {
		return ErrSyncFailed("serialize_channel", err)
	}

	// Guardar en Redis
	key := fmt.Sprintf("chat:channel:%s", redisChannel.ID)
	err = s.redisClient.Set(ctx, key, channelJSON, 0).Err()
	if err != nil {
		return ErrRedisConnection("set_channel", err)
	}

	return nil
}

// ShouldRetry determina si un error debe ser reintentado
func (s *BatchSyncStrategy) ShouldRetry(err error) bool {
	return s.retryManager.ShouldRetry(err)
}

// GetRetryDelay obtiene el delay para un intento específico
func (s *BatchSyncStrategy) GetRetryDelay(attempt int) time.Duration {
	return s.retryManager.GetRetryDelay(attempt)
}

// GetMaxRetries obtiene el número máximo de reintentos
func (s *BatchSyncStrategy) GetMaxRetries() int {
	return s.retryManager.GetMaxRetries()
}

// GetName obtiene el nombre de la estrategia
func (s *BatchSyncStrategy) GetName() string {
	return "batch"
}

// GetDescription obtiene la descripción de la estrategia
func (s *BatchSyncStrategy) GetDescription() string {
	return "Sincronización por lotes para mejor rendimiento"
}
