package sync

import "time"

// SyncConfig configuración para el sistema de sincronización
type SyncConfig struct {
	// Configuración de reintentos
	MaxRetries    int           `yaml:"max_retries" default:"3"`
	RetryDelay    time.Duration `yaml:"retry_delay" default:"1s"`
	MaxRetryDelay time.Duration `yaml:"max_retry_delay" default:"30s"`

	// Configuración de lotes
	BatchSize    int           `yaml:"batch_size" default:"10"`
	BatchTimeout time.Duration `yaml:"batch_timeout" default:"5s"`

	// Sincronización periódica
	SyncInterval     time.Duration `yaml:"sync_interval" default:"5m"`
	ConsistencyCheck time.Duration `yaml:"consistency_check" default:"1h"`

	// Circuit breaker
	CircuitBreaker CircuitBreakerConfig `yaml:"circuit_breaker"`

	// Configuración de contexto
	ContextTimeout time.Duration `yaml:"context_timeout" default:"30s"`
}

// CircuitBreakerConfig configuración del circuit breaker
type CircuitBreakerConfig struct {
	MaxFailures      int           `yaml:"max_failures" default:"5"`
	ResetTimeout     time.Duration `yaml:"reset_timeout" default:"1m"`
	HalfOpenMaxCalls int           `yaml:"half_open_max_calls" default:"3"`
}

// DefaultSyncConfig retorna la configuración por defecto
func DefaultSyncConfig() *SyncConfig {
	return &SyncConfig{
		MaxRetries:       3,
		RetryDelay:       time.Second,
		MaxRetryDelay:    30 * time.Second,
		BatchSize:        10,
		BatchTimeout:     5 * time.Second,
		SyncInterval:     5 * time.Minute,
		ConsistencyCheck: time.Hour,
		CircuitBreaker: CircuitBreakerConfig{
			MaxFailures:      5,
			ResetTimeout:     time.Minute,
			HalfOpenMaxCalls: 3,
		},
		ContextTimeout: 30 * time.Second,
	}
}

// Validate valida la configuración
func (c *SyncConfig) Validate() error {
	if c.MaxRetries < 0 {
		return ErrInvalidConfig("MaxRetries debe ser >= 0")
	}
	if c.RetryDelay <= 0 {
		return ErrInvalidConfig("RetryDelay debe ser > 0")
	}
	if c.MaxRetryDelay <= 0 {
		return ErrInvalidConfig("MaxRetryDelay debe ser > 0")
	}
	if c.BatchSize <= 0 {
		return ErrInvalidConfig("BatchSize debe ser > 0")
	}
	if c.BatchTimeout <= 0 {
		return ErrInvalidConfig("BatchTimeout debe ser > 0")
	}
	if c.SyncInterval <= 0 {
		return ErrInvalidConfig("SyncInterval debe ser > 0")
	}
	if c.ContextTimeout <= 0 {
		return ErrInvalidConfig("ContextTimeout debe ser > 0")
	}
	return nil
}
