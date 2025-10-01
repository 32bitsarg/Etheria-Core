package sync

import (
	"sync"
	"time"
)

// CircuitBreaker implementación de circuit breaker para sincronización
type CircuitBreaker struct {
	mu sync.RWMutex

	config CircuitBreakerConfig

	// Estado actual
	state CircuitBreakerState

	// Contadores
	failures    int
	successes   int
	lastFailure time.Time

	// Configuración de half-open
	halfOpenCalls int
}

// NewCircuitBreaker crea una nueva instancia de circuit breaker
func NewCircuitBreaker(config CircuitBreakerConfig) *CircuitBreaker {
	return &CircuitBreaker{
		config: config,
		state:  CircuitBreakerClosed,
	}
}

// Execute ejecuta una función con protección del circuit breaker
func (cb *CircuitBreaker) Execute(fn func() error) error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	// Verificar si el circuit breaker está abierto
	if cb.state == CircuitBreakerOpen {
		// Verificar si es tiempo de intentar half-open
		if time.Since(cb.lastFailure) >= cb.config.ResetTimeout {
			cb.state = CircuitBreakerHalfOpen
			cb.halfOpenCalls = 0
		} else {
			return ErrCircuitBreakerOpen("execute")
		}
	}

	// Verificar límite de llamadas en half-open
	if cb.state == CircuitBreakerHalfOpen {
		if cb.halfOpenCalls >= cb.config.HalfOpenMaxCalls {
			return ErrCircuitBreakerHalfOpen("execute")
		}
		cb.halfOpenCalls++
	}

	// Ejecutar la función
	err := fn()

	// Procesar el resultado
	if err != nil {
		cb.recordFailure()
	} else {
		cb.recordSuccess()
	}

	return err
}

// recordFailure registra un fallo
func (cb *CircuitBreaker) recordFailure() {
	cb.failures++
	cb.lastFailure = time.Now()

	// Verificar si se debe abrir el circuit breaker
	if cb.failures >= cb.config.MaxFailures {
		cb.state = CircuitBreakerOpen
	}
}

// recordSuccess registra un éxito
func (cb *CircuitBreaker) recordSuccess() {
	cb.successes++

	// Si estamos en half-open y tenemos suficientes éxitos, cerrar
	if cb.state == CircuitBreakerHalfOpen {
		if cb.successes >= cb.config.HalfOpenMaxCalls {
			cb.state = CircuitBreakerClosed
			cb.failures = 0
			cb.successes = 0
		}
	}
}

// GetState obtiene el estado actual del circuit breaker
func (cb *CircuitBreaker) GetState() CircuitBreakerState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return cb.state
}

// GetStats obtiene estadísticas del circuit breaker
func (cb *CircuitBreaker) GetStats() CircuitBreakerStats {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return CircuitBreakerStats{
		State:         cb.state,
		Failures:      cb.failures,
		Successes:     cb.successes,
		LastFailure:   cb.lastFailure,
		HalfOpenCalls: cb.halfOpenCalls,
	}
}

// Reset resetea el circuit breaker al estado cerrado
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.state = CircuitBreakerClosed
	cb.failures = 0
	cb.successes = 0
	cb.halfOpenCalls = 0
}

// CircuitBreakerStats estadísticas del circuit breaker
type CircuitBreakerStats struct {
	State         CircuitBreakerState `json:"state"`
	Failures      int                 `json:"failures"`
	Successes     int                 `json:"successes"`
	LastFailure   time.Time           `json:"last_failure"`
	HalfOpenCalls int                 `json:"half_open_calls"`
}
