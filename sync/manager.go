package sync

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
)

// SyncManagerImpl implementación principal del gestor de sincronización
type SyncManagerImpl struct {
	mu sync.RWMutex

	// Dependencias
	chatRepo    ChatRepository
	redisClient RedisClient
	logger      *zap.Logger

	// Configuración
	config *SyncConfig

	// Componentes
	strategy       SyncStrategy
	circuitBreaker *CircuitBreaker
	metrics        *Metrics
	retryManager   *RetryManager

	// Estado
	isRunning bool
	stopCh    chan struct{}

	// Timers
	syncTimer        *time.Timer
	consistencyTimer *time.Timer
}

// NewSyncManager crea una nueva instancia del gestor de sincronización
func NewSyncManager(
	chatRepo ChatRepository,
	redisClient RedisClient,
	logger *zap.Logger,
	config *SyncConfig,
) *SyncManagerImpl {
	if config == nil {
		config = DefaultSyncConfig()
	}

	// Crear circuit breaker
	circuitBreaker := NewCircuitBreaker(config.CircuitBreaker)

	// Crear métricas
	metrics := NewMetrics()

	// Crear retry manager
	retryConfig := RetryConfig{
		MaxRetries:    config.MaxRetries,
		RetryDelay:    config.RetryDelay,
		MaxRetryDelay: config.MaxRetryDelay,
		BackoffFactor: 2.0,
	}
	retryManager := NewRetryManager(retryConfig)

	// Crear estrategia por defecto (batch)
	strategy := NewBatchSyncStrategy(chatRepo, redisClient, config)

	return &SyncManagerImpl{
		chatRepo:       chatRepo,
		redisClient:    redisClient,
		logger:         logger,
		config:         config,
		strategy:       strategy,
		circuitBreaker: circuitBreaker,
		metrics:        metrics,
		retryManager:   retryManager,
		stopCh:         make(chan struct{}),
	}
}

// Start inicia el gestor de sincronización
func (sm *SyncManagerImpl) Start(ctx context.Context) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.isRunning {
		return ErrSyncFailed("start", ErrSyncAborted("already running"))
	}

	sm.isRunning = true

	// Sincronización inicial
	go sm.initialSync(ctx)

	// Sincronización periódica
	sm.syncTimer = time.NewTimer(sm.config.SyncInterval)
	go sm.periodicSync(ctx)

	// Verificación de consistencia periódica
	sm.consistencyTimer = time.NewTimer(sm.config.ConsistencyCheck)
	go sm.consistencyCheck(ctx)

	sm.logger.Info("SyncManager iniciado exitosamente")
	return nil
}

// Stop detiene el gestor de sincronización
func (sm *SyncManagerImpl) Stop() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if !sm.isRunning {
		return nil
	}

	sm.isRunning = false
	close(sm.stopCh)

	// Detener timers
	if sm.syncTimer != nil {
		sm.syncTimer.Stop()
	}
	if sm.consistencyTimer != nil {
		sm.consistencyTimer.Stop()
	}

	sm.logger.Info("SyncManager detenido exitosamente")
	return nil
}

// SyncAll sincroniza todos los canales
func (sm *SyncManagerImpl) SyncAll(ctx context.Context) error {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if !sm.isRunning {
		return ErrSyncFailed("sync_all", ErrSyncAborted("manager not running"))
	}

	// Obtener todos los canales de BD
	channels, err := sm.chatRepo.GetAllChannels()
	if err != nil {
		return ErrSyncFailed("sync_all", err)
	}

	// Ejecutar sincronización con circuit breaker
	return sm.circuitBreaker.Execute(func() error {
		start := time.Now()
		defer func() {
			sm.metrics.RecordSyncDuration("sync_all", time.Since(start))
		}()

		err := sm.strategy.Sync(ctx, channels)
		if err != nil {
			sm.metrics.RecordSyncFailure("sync_all", err)
			return err
		}

		sm.metrics.RecordSyncSuccess("sync_all")
		sm.metrics.RecordBatchSize(len(channels))

		return nil
	})
}

// SyncChannel sincroniza un canal específico
func (sm *SyncManagerImpl) SyncChannel(ctx context.Context, channelID string) error {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if !sm.isRunning {
		return ErrSyncFailed("sync_channel", ErrSyncAborted("manager not running"))
	}

	// Obtener canal de BD
	channel, err := sm.chatRepo.GetChannelByID(channelID)
	if err != nil {
		return ErrSyncFailed("sync_channel", err)
	}

	// Ejecutar sincronización con circuit breaker
	return sm.circuitBreaker.Execute(func() error {
		start := time.Now()
		defer func() {
			sm.metrics.RecordSyncDuration("sync_channel", time.Since(start))
		}()

		err := sm.strategy.SyncChannel(ctx, channel)
		if err != nil {
			sm.metrics.RecordSyncFailure("sync_channel", err)
			return err
		}

		sm.metrics.RecordSyncSuccess("sync_channel")

		return nil
	})
}

// SyncChannels sincroniza múltiples canales específicos
func (sm *SyncManagerImpl) SyncChannels(ctx context.Context, channelIDs []string) error {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if !sm.isRunning {
		return ErrSyncFailed("sync_channels", ErrSyncAborted("manager not running"))
	}

	// Obtener canales de BD
	var channels []*ChatChannel
	for _, channelID := range channelIDs {
		channel, err := sm.chatRepo.GetChannelByID(channelID)
		if err != nil {
			sm.logger.Warn("Error obteniendo canal", zap.String("channel_id", channelID), zap.Error(err))
			continue
		}
		channels = append(channels, channel)
	}

	if len(channels) == 0 {
		return nil
	}

	// Ejecutar sincronización con circuit breaker
	return sm.circuitBreaker.Execute(func() error {
		start := time.Now()
		defer func() {
			sm.metrics.RecordSyncDuration("sync_channels", time.Since(start))
		}()

		err := sm.strategy.Sync(ctx, channels)
		if err != nil {
			sm.metrics.RecordSyncFailure("sync_channels", err)
			return err
		}

		sm.metrics.RecordSyncSuccess("sync_channels")
		sm.metrics.RecordBatchSize(len(channels))

		return nil
	})
}

// GetStatus obtiene el estado actual del gestor
func (sm *SyncManagerImpl) GetStatus() SyncStatus {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	stats := sm.metrics.GetSyncStats()

	return SyncStatus{
		IsRunning:           sm.isRunning,
		LastSyncTime:        stats.LastSyncTime,
		LastSyncDuration:    stats.AvgSyncDuration,
		LastSyncError:       stats.LastSyncError,
		CircuitBreakerState: sm.circuitBreaker.GetState(),
		TotalSyncs:          stats.TotalSyncs,
		SuccessfulSyncs:     stats.SuccessfulSyncs,
		FailedSyncs:         stats.FailedSyncs,
		RetryAttempts:       stats.RetryAttempts,
		ChannelsSynced:      0, // TODO: Implementar contador
		ChannelsPending:     0, // TODO: Implementar contador
		ChannelsFailed:      0, // TODO: Implementar contador
	}
}

// GetMetrics obtiene las métricas del gestor
func (sm *SyncManagerImpl) GetMetrics() SyncMetrics {
	return sm.metrics
}

// IsHealthy verifica si el gestor está saludable
func (sm *SyncManagerImpl) IsHealthy() bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	// Verificar si está corriendo
	if !sm.isRunning {
		return false
	}

	// Verificar estado del circuit breaker
	if sm.circuitBreaker.GetState() == CircuitBreakerOpen {
		return false
	}

	// Verificar métricas de error
	stats := sm.metrics.GetSyncStats()
	if stats.TotalSyncs > 0 {
		errorRate := float64(stats.FailedSyncs) / float64(stats.TotalSyncs)
		if errorRate > 0.5 { // Más del 50% de errores
			return false
		}
	}

	return true
}

// initialSync realiza la sincronización inicial
func (sm *SyncManagerImpl) initialSync(ctx context.Context) {
	sm.logger.Info("Iniciando sincronización inicial")

	if err := sm.SyncAll(ctx); err != nil {
		sm.logger.Error("Error en sincronización inicial", zap.Error(err))
	} else {
		sm.logger.Info("Sincronización inicial completada exitosamente")
	}
}

// periodicSync realiza la sincronización periódica
func (sm *SyncManagerImpl) periodicSync(ctx context.Context) {
	for {
		select {
		case <-sm.stopCh:
			return
		case <-sm.syncTimer.C:
			sm.logger.Debug("Ejecutando sincronización periódica")

			if err := sm.SyncAll(ctx); err != nil {
				sm.logger.Error("Error en sincronización periódica", zap.Error(err))
			}

			// Resetear timer
			sm.syncTimer.Reset(sm.config.SyncInterval)
		}
	}
}

// consistencyCheck realiza la verificación de consistencia periódica
func (sm *SyncManagerImpl) consistencyCheck(ctx context.Context) {
	for {
		select {
		case <-sm.stopCh:
			return
		case <-sm.consistencyTimer.C:
			sm.logger.Debug("Ejecutando verificación de consistencia")

			// TODO: Implementar verificación de consistencia
			sm.metrics.RecordConsistencyCheck(true, 0)

			// Resetear timer
			sm.consistencyTimer.Reset(sm.config.ConsistencyCheck)
		}
	}
}
