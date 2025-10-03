package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// LoggingConfig configuración para el middleware de logging
type LoggingConfig struct {
	Logger       *zap.Logger
	SkipPaths    []string
	EnableColors bool
}

// CustomLoggingMiddleware crea un middleware de logging personalizado y legible
func CustomLoggingMiddleware(config LoggingConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Saltar paths específicos si es necesario
		for _, skipPath := range config.SkipPaths {
			if path == skipPath {
				c.Next()
				return
			}
		}

		// Procesar request
		c.Next()

		// Calcular duración
		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		bodySize := c.Writer.Size()
		userAgent := c.Request.UserAgent()

		// Determinar color basado en status code
		var statusColor string
		if config.EnableColors {
			statusColor = getStatusColor(statusCode)
		}

		// Formatear query string
		if raw != "" {
			path = path + "?" + raw
		}

		// Crear mensaje de log estructurado
		logMessage := fmt.Sprintf(
			"🌐 %s %s %s %d %s %s %s %s",
			getMethodIcon(method),
			method,
			path,
			statusCode,
			statusColor,
			formatLatency(latency),
			clientIP,
			formatBodySize(bodySize),
		)

		// Log basado en status code
		switch {
		case statusCode >= 500:
			config.Logger.Error(logMessage,
				zap.String("method", method),
				zap.String("path", path),
				zap.Int("status", statusCode),
				zap.Duration("latency", latency),
				zap.String("ip", clientIP),
				zap.Int("size", bodySize),
				zap.String("user_agent", userAgent),
			)
		case statusCode >= 400:
			config.Logger.Warn(logMessage,
				zap.String("method", method),
				zap.String("path", path),
				zap.Int("status", statusCode),
				zap.Duration("latency", latency),
				zap.String("ip", clientIP),
				zap.Int("size", bodySize),
			)
		case statusCode >= 300:
			config.Logger.Info(logMessage,
				zap.String("method", method),
				zap.String("path", path),
				zap.Int("status", statusCode),
				zap.Duration("latency", latency),
				zap.String("ip", clientIP),
			)
		default:
			config.Logger.Info(logMessage,
				zap.String("method", method),
				zap.String("path", path),
				zap.Int("status", statusCode),
				zap.Duration("latency", latency),
				zap.String("ip", clientIP),
			)
		}
	}
}

// getMethodIcon retorna un emoji para el método HTTP
func getMethodIcon(method string) string {
	switch method {
	case "GET":
		return "📥"
	case "POST":
		return "📤"
	case "PUT":
		return "🔄"
	case "DELETE":
		return "🗑️"
	case "PATCH":
		return "🔧"
	default:
		return "❓"
	}
}

// getStatusColor retorna el color ANSI para el status code
func getStatusColor(statusCode int) string {
	switch {
	case statusCode >= 200 && statusCode < 300:
		return "✅" // Verde
	case statusCode >= 300 && statusCode < 400:
		return "🔄" // Amarillo
	case statusCode >= 400 && statusCode < 500:
		return "⚠️" // Naranja
	case statusCode >= 500:
		return "❌" // Rojo
	default:
		return "❓" // Gris
	}
}

// formatLatency formatea la duración de manera legible
func formatLatency(latency time.Duration) string {
	if latency < time.Millisecond {
		return fmt.Sprintf("%.2fμs", float64(latency.Nanoseconds())/1000)
	} else if latency < time.Second {
		return fmt.Sprintf("%.2fms", float64(latency.Nanoseconds())/1000000)
	} else {
		return fmt.Sprintf("%.2fs", latency.Seconds())
	}
}

// formatBodySize formatea el tamaño del body de manera legible
func formatBodySize(size int) string {
	if size == 0 {
		return "0B"
	} else if size < 1024 {
		return fmt.Sprintf("%dB", size)
	} else if size < 1024*1024 {
		return fmt.Sprintf("%.1fKB", float64(size)/1024)
	} else {
		return fmt.Sprintf("%.1fMB", float64(size)/(1024*1024))
	}
}

// AuthLoggingMiddleware middleware específico para logging de autenticación
func AuthLoggingMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Log de request entrante
		logger.Info("🔐 AUTH REQUEST",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
		)

		c.Next()

		// Log de response
		latency := time.Since(start)
		statusCode := c.Writer.Status()

		switch {
		case statusCode >= 500:
			logger.Error("❌ AUTH RESPONSE",
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
				zap.Int("status", statusCode),
				zap.Duration("latency", latency),
				zap.String("ip", c.ClientIP()),
			)
		case statusCode >= 400:
			logger.Warn("⚠️ AUTH RESPONSE",
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
				zap.Int("status", statusCode),
				zap.Duration("latency", latency),
				zap.String("ip", c.ClientIP()),
			)
		default:
			logger.Info("✅ AUTH RESPONSE",
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
				zap.Int("status", statusCode),
				zap.Duration("latency", latency),
				zap.String("ip", c.ClientIP()),
			)
		}
	}
}

// WebSocketLoggingMiddleware middleware específico para logging de WebSocket
func WebSocketLoggingMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger.Info("🔌 WEBSOCKET CONNECTION",
			zap.String("path", c.Request.URL.Path),
			zap.String("ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
		)
		c.Next()
	}
}
