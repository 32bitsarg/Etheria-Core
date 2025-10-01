package sync

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// SyncManager interface principal para el manejo de sincronización
type SyncManager interface {
	// Lifecycle
	Start(ctx context.Context) error
	Stop() error

	// Sincronización
	SyncAll(ctx context.Context) error
	SyncChannel(ctx context.Context, channelID string) error
	SyncChannels(ctx context.Context, channelIDs []string) error

	// Estado y métricas
	GetStatus() SyncStatus
	GetMetrics() SyncMetrics
	IsHealthy() bool
}

// SyncStrategy interface para estrategias de sincronización
type SyncStrategy interface {
	// Sincronización
	Sync(ctx context.Context, channels []*ChatChannel) error
	SyncChannel(ctx context.Context, channel *ChatChannel) error

	// Configuración de reintentos
	ShouldRetry(err error) bool
	GetRetryDelay(attempt int) time.Duration
	GetMaxRetries() int

	// Información
	GetName() string
	GetDescription() string
}

// SyncMetrics interface para métricas de sincronización
type SyncMetrics interface {
	// Métricas de sincronización
	RecordSyncDuration(operation string, duration time.Duration)
	RecordSyncSuccess(operation string)
	RecordSyncFailure(operation string, err error)
	RecordRetryAttempt(operation string)

	// Métricas de circuit breaker
	RecordCircuitBreakerStateChange(state CircuitBreakerState)
	GetCircuitBreakerState() CircuitBreakerState

	// Métricas de rendimiento
	RecordBatchSize(size int)
	RecordSyncLatency(operation string, latency time.Duration)

	// Métricas de consistencia
	RecordConsistencyCheck(success bool, inconsistencies int)
	RecordDataValidation(success bool, errors int)

	// Obtener métricas
	GetSyncStats() SyncStats
	GetPerformanceStats() PerformanceStats
	GetConsistencyStats() ConsistencyStats
}

// RedisClient interface para el cliente de Redis
type RedisClient interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Get(ctx context.Context, key string) *redis.StringCmd
	Pipeline() redis.Pipeliner
	Close() error
}

// ChatRepository interface para el repositorio de chat
type ChatRepository interface {
	GetAllChannels() ([]*ChatChannel, error)
	GetChannelByID(channelID string) (*ChatChannel, error)
	GetChannelByName(name string) (*ChatChannel, error)
	CreateChannel(channel *ChatChannel) error
	UpdateChannel(channel *ChatChannel) error
	DeleteChannel(channelID string) error
}

// ChatChannel modelo de canal de chat
type ChatChannel struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	WorldID     *string   `json:"world_id,omitempty"`
	AllianceID  *string   `json:"alliance_id,omitempty"`
	MemberCount int       `json:"member_count"`
	CreatedAt   time.Time `json:"created_at"`
	IsActive    bool      `json:"is_active"`
}

// SyncStatus estado del sistema de sincronización
type SyncStatus struct {
	IsRunning           bool                `json:"is_running"`
	LastSyncTime        time.Time           `json:"last_sync_time"`
	LastSyncDuration    time.Duration       `json:"last_sync_duration"`
	LastSyncError       string              `json:"last_sync_error,omitempty"`
	CircuitBreakerState CircuitBreakerState `json:"circuit_breaker_state"`
	TotalSyncs          int64               `json:"total_syncs"`
	SuccessfulSyncs     int64               `json:"successful_syncs"`
	FailedSyncs         int64               `json:"failed_syncs"`
	RetryAttempts       int64               `json:"retry_attempts"`
	ChannelsSynced      int                 `json:"channels_synced"`
	ChannelsPending     int                 `json:"channels_pending"`
	ChannelsFailed      int                 `json:"channels_failed"`
}

// CircuitBreakerState estado del circuit breaker
type CircuitBreakerState int

const (
	CircuitBreakerClosed CircuitBreakerState = iota
	CircuitBreakerOpen
	CircuitBreakerHalfOpen
)

func (s CircuitBreakerState) String() string {
	switch s {
	case CircuitBreakerClosed:
		return "closed"
	case CircuitBreakerOpen:
		return "open"
	case CircuitBreakerHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// SyncStats estadísticas de sincronización
type SyncStats struct {
	TotalSyncs      int64         `json:"total_syncs"`
	SuccessfulSyncs int64         `json:"successful_syncs"`
	FailedSyncs     int64         `json:"failed_syncs"`
	RetryAttempts   int64         `json:"retry_attempts"`
	AvgSyncDuration time.Duration `json:"avg_sync_duration"`
	LastSyncTime    time.Time     `json:"last_sync_time"`
	LastSyncError   string        `json:"last_sync_error,omitempty"`
}

// PerformanceStats estadísticas de rendimiento
type PerformanceStats struct {
	AvgLatency       time.Duration `json:"avg_latency"`
	MaxLatency       time.Duration `json:"max_latency"`
	MinLatency       time.Duration `json:"min_latency"`
	AvgBatchSize     float64       `json:"avg_batch_size"`
	MaxBatchSize     int           `json:"max_batch_size"`
	MinBatchSize     int           `json:"min_batch_size"`
	ThroughputPerSec float64       `json:"throughput_per_sec"`
	ErrorRate        float64       `json:"error_rate"`
}

// ConsistencyStats estadísticas de consistencia
type ConsistencyStats struct {
	TotalChecks      int64     `json:"total_checks"`
	SuccessfulChecks int64     `json:"successful_checks"`
	FailedChecks     int64     `json:"failed_checks"`
	Inconsistencies  int64     `json:"inconsistencies"`
	ValidationErrors int64     `json:"validation_errors"`
	ConsistencyRate  float64   `json:"consistency_rate"`
	LastCheckTime    time.Time `json:"last_check_time"`
	LastCheckError   string    `json:"last_check_error,omitempty"`
}

// SyncEvent evento de sincronización
type SyncEvent struct {
	Type       string        `json:"type"`
	Operation  string        `json:"operation"`
	ChannelID  string        `json:"channel_id,omitempty"`
	Success    bool          `json:"success"`
	Error      string        `json:"error,omitempty"`
	Duration   time.Duration `json:"duration"`
	Timestamp  time.Time     `json:"timestamp"`
	RetryCount int           `json:"retry_count,omitempty"`
	BatchSize  int           `json:"batch_size,omitempty"`
}

// SyncEventHandler función para manejar eventos de sincronización
type SyncEventHandler func(event SyncEvent)

// SyncEventBus interface para el bus de eventos
type SyncEventBus interface {
	Subscribe(handler SyncEventHandler) string
	Unsubscribe(id string)
	Publish(event SyncEvent)
	Close()
}
