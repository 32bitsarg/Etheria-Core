package sync

import (
	"context"
	"math"
	"time"
)

// RetryConfig configuración para el sistema de reintentos
type RetryConfig struct {
	MaxRetries    int           `yaml:"max_retries"`
	RetryDelay    time.Duration `yaml:"retry_delay"`
	MaxRetryDelay time.Duration `yaml:"max_retry_delay"`
	BackoffFactor float64       `yaml:"backoff_factor"`
}

// RetryManager maneja la lógica de reintentos
type RetryManager struct {
	config RetryConfig
}

// NewRetryManager crea una nueva instancia del gestor de reintentos
func NewRetryManager(config RetryConfig) *RetryManager {
	return &RetryManager{
		config: config,
	}
}

// ExecuteWithRetry ejecuta una función con reintentos automáticos
func (rm *RetryManager) ExecuteWithRetry(ctx context.Context, operation string, fn func() error) error {
	var lastErr error

	for attempt := 0; attempt <= rm.config.MaxRetries; attempt++ {
		// Verificar si el contexto fue cancelado
		select {
		case <-ctx.Done():
			return ErrContextCanceled(operation)
		default:
		}

		// Ejecutar la función
		err := fn()
		if err == nil {
			return nil // Éxito
		}

		lastErr = err

		// Verificar si el error es reintentable
		if !IsRetryableError(err) {
			return err
		}

		// Si no es el último intento, esperar antes del siguiente
		if attempt < rm.config.MaxRetries {
			delay := rm.calculateDelay(attempt)

			// Verificar timeout del contexto
			select {
			case <-ctx.Done():
				return ErrContextCanceled(operation)
			case <-time.After(delay):
				// Continuar con el siguiente intento
			}
		}
	}

	return ErrSyncFailed(operation, lastErr)
}

// calculateDelay calcula el delay para el siguiente intento
func (rm *RetryManager) calculateDelay(attempt int) time.Duration {
	// Calcular delay con backoff exponencial
	delay := time.Duration(float64(rm.config.RetryDelay) * math.Pow(rm.config.BackoffFactor, float64(attempt)))

	// Limitar al máximo delay configurado
	if delay > rm.config.MaxRetryDelay {
		delay = rm.config.MaxRetryDelay
	}

	return delay
}

// ShouldRetry determina si un error debe ser reintentado
func (rm *RetryManager) ShouldRetry(err error) bool {
	return IsRetryableError(err)
}

// GetRetryDelay obtiene el delay para un intento específico
func (rm *RetryManager) GetRetryDelay(attempt int) time.Duration {
	return rm.calculateDelay(attempt)
}

// GetMaxRetries obtiene el número máximo de reintentos
func (rm *RetryManager) GetMaxRetries() int {
	return rm.config.MaxRetries
}

// RetryStats estadísticas de reintentos
type RetryStats struct {
	TotalAttempts     int64         `json:"total_attempts"`
	SuccessfulRetries int64         `json:"successful_retries"`
	FailedRetries     int64         `json:"failed_retries"`
	AvgRetryDelay     time.Duration `json:"avg_retry_delay"`
	MaxRetryDelay     time.Duration `json:"max_retry_delay"`
	LastRetryTime     time.Time     `json:"last_retry_time"`
}

// RetryableError error que puede ser reintentado
type RetryableError struct {
	Err        error
	Attempt    int
	MaxRetries int
}

func (e *RetryableError) Error() string {
	return e.Err.Error()
}

func (e *RetryableError) Unwrap() error {
	return e.Err
}

// IsRetryableError determina si un error es reintentable
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Verificar si es un error de circuit breaker
	if _, ok := err.(*CircuitBreakerError); ok {
		return false
	}

	// Verificar si es un error de contexto
	if err == context.Canceled || err == context.DeadlineExceeded {
		return false
	}

	// Verificar el tipo de error
	errStr := err.Error()

	// Errores que NO son reintentables
	nonRetryableErrors := []string{
		"configuración inválida",
		"circuit breaker abierto",
		"contexto cancelado",
		"timeout de contexto",
		"validación falló",
	}

	for _, nonRetryable := range nonRetryableErrors {
		if contains(errStr, nonRetryable) {
			return false
		}
	}

	// Errores que SÍ son reintentables
	retryableErrors := []string{
		"error de conexión",
		"Redis no disponible",
		"base de datos no disponible",
		"sincronización falló",
		"timeout en sincronización",
	}

	for _, retryable := range retryableErrors {
		if contains(errStr, retryable) {
			return true
		}
	}

	// Por defecto, asumir que es reintentable
	return true
}

// CircuitBreakerError error específico del circuit breaker
type CircuitBreakerError struct {
	State     CircuitBreakerState
	Operation string
}

func (e *CircuitBreakerError) Error() string {
	return ErrCircuitBreakerOpen(e.Operation).Error()
}
