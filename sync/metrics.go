package sync

import (
	"sync"
	"time"
)

// Metrics implementación básica de métricas de sincronización
type Metrics struct {
	mu sync.RWMutex

	// Métricas de sincronización
	totalSyncs      int64
	successfulSyncs int64
	failedSyncs     int64
	retryAttempts   int64

	// Métricas de rendimiento
	syncDurations []time.Duration
	latencies     map[string][]time.Duration
	batchSizes    []int

	// Métricas de consistencia
	totalChecks      int64
	successfulChecks int64
	failedChecks     int64
	inconsistencies  int64
	validationErrors int64

	// Estado del circuit breaker
	circuitBreakerState CircuitBreakerState

	// Timestamps
	lastSyncTime   time.Time
	lastCheckTime  time.Time
	lastSyncError  string
	lastCheckError string
}

// NewMetrics crea una nueva instancia de métricas
func NewMetrics() *Metrics {
	return &Metrics{
		latencies: make(map[string][]time.Duration),
	}
}

// RecordSyncDuration registra la duración de una sincronización
func (m *Metrics) RecordSyncDuration(operation string, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.syncDurations = append(m.syncDurations, duration)

	// Mantener solo los últimos 1000 registros
	if len(m.syncDurations) > 1000 {
		m.syncDurations = m.syncDurations[1:]
	}
}

// RecordSyncSuccess registra una sincronización exitosa
func (m *Metrics) RecordSyncSuccess(operation string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.totalSyncs++
	m.successfulSyncs++
	m.lastSyncTime = time.Now()
	m.lastSyncError = ""
}

// RecordSyncFailure registra una sincronización fallida
func (m *Metrics) RecordSyncFailure(operation string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.totalSyncs++
	m.failedSyncs++
	m.lastSyncTime = time.Now()
	if err != nil {
		m.lastSyncError = err.Error()
	}
}

// RecordRetryAttempt registra un intento de reintento
func (m *Metrics) RecordRetryAttempt(operation string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.retryAttempts++
}

// RecordCircuitBreakerStateChange registra un cambio de estado del circuit breaker
func (m *Metrics) RecordCircuitBreakerStateChange(state CircuitBreakerState) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.circuitBreakerState = state
}

// GetCircuitBreakerState obtiene el estado actual del circuit breaker
func (m *Metrics) GetCircuitBreakerState() CircuitBreakerState {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.circuitBreakerState
}

// RecordBatchSize registra el tamaño de un lote
func (m *Metrics) RecordBatchSize(size int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.batchSizes = append(m.batchSizes, size)

	// Mantener solo los últimos 1000 registros
	if len(m.batchSizes) > 1000 {
		m.batchSizes = m.batchSizes[1:]
	}
}

// RecordSyncLatency registra la latencia de una operación
func (m *Metrics) RecordSyncLatency(operation string, latency time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.latencies[operation] == nil {
		m.latencies[operation] = make([]time.Duration, 0)
	}

	m.latencies[operation] = append(m.latencies[operation], latency)

	// Mantener solo los últimos 1000 registros por operación
	if len(m.latencies[operation]) > 1000 {
		m.latencies[operation] = m.latencies[operation][1:]
	}
}

// RecordConsistencyCheck registra el resultado de una verificación de consistencia
func (m *Metrics) RecordConsistencyCheck(success bool, inconsistencies int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.totalChecks++
	if success {
		m.successfulChecks++
	} else {
		m.failedChecks++
	}

	m.inconsistencies += int64(inconsistencies)
	m.lastCheckTime = time.Now()
}

// RecordDataValidation registra el resultado de una validación de datos
func (m *Metrics) RecordDataValidation(success bool, errors int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !success {
		m.validationErrors += int64(errors)
	}
}

// GetSyncStats obtiene las estadísticas de sincronización
func (m *Metrics) GetSyncStats() SyncStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var avgDuration time.Duration
	if len(m.syncDurations) > 0 {
		var total time.Duration
		for _, d := range m.syncDurations {
			total += d
		}
		avgDuration = total / time.Duration(len(m.syncDurations))
	}

	return SyncStats{
		TotalSyncs:      m.totalSyncs,
		SuccessfulSyncs: m.successfulSyncs,
		FailedSyncs:     m.failedSyncs,
		RetryAttempts:   m.retryAttempts,
		AvgSyncDuration: avgDuration,
		LastSyncTime:    m.lastSyncTime,
		LastSyncError:   m.lastSyncError,
	}
}

// GetPerformanceStats obtiene las estadísticas de rendimiento
func (m *Metrics) GetPerformanceStats() PerformanceStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var avgLatency, maxLatency, minLatency time.Duration
	var totalLatency time.Duration
	var count int

	for _, latencies := range m.latencies {
		for _, latency := range latencies {
			totalLatency += latency
			count++
			if latency > maxLatency {
				maxLatency = latency
			}
			if minLatency == 0 || latency < minLatency {
				minLatency = latency
			}
		}
	}

	if count > 0 {
		avgLatency = totalLatency / time.Duration(count)
	}

	var avgBatchSize float64
	var maxBatchSize, minBatchSize int

	if len(m.batchSizes) > 0 {
		var totalBatchSize int
		for _, size := range m.batchSizes {
			totalBatchSize += size
			if size > maxBatchSize {
				maxBatchSize = size
			}
			if minBatchSize == 0 || size < minBatchSize {
				minBatchSize = size
			}
		}
		avgBatchSize = float64(totalBatchSize) / float64(len(m.batchSizes))
	}

	var errorRate float64
	if m.totalSyncs > 0 {
		errorRate = float64(m.failedSyncs) / float64(m.totalSyncs)
	}

	return PerformanceStats{
		AvgLatency:       avgLatency,
		MaxLatency:       maxLatency,
		MinLatency:       minLatency,
		AvgBatchSize:     avgBatchSize,
		MaxBatchSize:     maxBatchSize,
		MinBatchSize:     minBatchSize,
		ThroughputPerSec: 0, // TODO: Implementar cálculo de throughput
		ErrorRate:        errorRate,
	}
}

// GetConsistencyStats obtiene las estadísticas de consistencia
func (m *Metrics) GetConsistencyStats() ConsistencyStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var consistencyRate float64
	if m.totalChecks > 0 {
		consistencyRate = float64(m.successfulChecks) / float64(m.totalChecks)
	}

	return ConsistencyStats{
		TotalChecks:      m.totalChecks,
		SuccessfulChecks: m.successfulChecks,
		FailedChecks:     m.failedChecks,
		Inconsistencies:  m.inconsistencies,
		ValidationErrors: m.validationErrors,
		ConsistencyRate:  consistencyRate,
		LastCheckTime:    m.lastCheckTime,
		LastCheckError:   m.lastCheckError,
	}
}
