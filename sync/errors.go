package sync

import "fmt"

// Errores personalizados para el sistema de sincronización
var (
	// Errores de configuración
	ErrInvalidConfig = func(msg string) error {
		return fmt.Errorf("configuración inválida: %s", msg)
	}

	// Errores de sincronización
	ErrSyncFailed = func(operation string, err error) error {
		return fmt.Errorf("sincronización falló en %s: %v", operation, err)
	}

	ErrSyncTimeout = func(operation string) error {
		return fmt.Errorf("timeout en sincronización %s", operation)
	}

	ErrSyncAborted = func(operation string) error {
		return fmt.Errorf("sincronización abortada en %s", operation)
	}

	// Errores de circuit breaker
	ErrCircuitBreakerOpen = func(operation string) error {
		return fmt.Errorf("circuit breaker abierto para %s", operation)
	}

	ErrCircuitBreakerHalfOpen = func(operation string) error {
		return fmt.Errorf("circuit breaker en estado half-open para %s", operation)
	}

	// Errores de Redis
	ErrRedisUnavailable = func(operation string) error {
		return fmt.Errorf("Redis no disponible para %s", operation)
	}

	ErrRedisConnection = func(operation string, err error) error {
		return fmt.Errorf("error de conexión Redis en %s: %v", operation, err)
	}

	// Errores de base de datos
	ErrDatabaseUnavailable = func(operation string) error {
		return fmt.Errorf("base de datos no disponible para %s", operation)
	}

	ErrDatabaseConnection = func(operation string, err error) error {
		return fmt.Errorf("error de conexión BD en %s: %v", operation, err)
	}

	// Errores de consistencia
	ErrDataInconsistency = func(operation string, details string) error {
		return fmt.Errorf("inconsistencia de datos en %s: %s", operation, details)
	}

	ErrValidationFailed = func(operation string, details string) error {
		return fmt.Errorf("validación falló en %s: %s", operation, details)
	}

	// Errores de contexto
	ErrContextCanceled = func(operation string) error {
		return fmt.Errorf("contexto cancelado en %s", operation)
	}

	ErrContextTimeout = func(operation string) error {
		return fmt.Errorf("timeout de contexto en %s", operation)
	}
)

// contains verifica si una cadena contiene otra
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr ||
		len(s) > len(substr) && contains(s[1:], substr)
}
