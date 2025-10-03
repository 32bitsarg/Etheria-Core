package config

import (
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type WebSocketConfig struct {
	ReadBufferSize    int
	WriteBufferSize   int
	HandshakeTimeout  time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	MaxMessageSize    int64
	EnableCompression bool
	CheckOrigin       func(*http.Request) bool
}

func GetWebSocketConfig() *WebSocketConfig {
	return &WebSocketConfig{
		ReadBufferSize:    4096, // Aumentado de 1024 a 4096
		WriteBufferSize:   4096, // Aumentado de 1024 a 4096
		HandshakeTimeout:  10 * time.Second,
		ReadTimeout:       60 * time.Second,
		WriteTimeout:      10 * time.Second,
		MaxMessageSize:    1024, // Límite de 1KB por mensaje
		EnableCompression: true,
		CheckOrigin: func(r *http.Request) bool {
			// Validar origen para seguridad
			origin := r.Header.Get("Origin")
			allowedOrigins := []string{
				"http://localhost:3000",
				"http://127.0.0.1:3000",
				"http://10.0.2.2:3000",
				"http://localhost:8080",
				"http://127.0.0.1:8080",
				"http://10.0.2.2:8080",
			}

			for _, allowed := range allowedOrigins {
				if origin == allowed {
					return true
				}
			}

			// En desarrollo, permitir cualquier origen
			return true
		},
	}
}

// GetWebSocketUpgrader crea un upgrader WebSocket con configuración robusta
func GetWebSocketUpgrader() websocket.Upgrader {
	config := GetWebSocketConfig()
	return websocket.Upgrader{
		ReadBufferSize:    config.ReadBufferSize,
		WriteBufferSize:   config.WriteBufferSize,
		HandshakeTimeout:  config.HandshakeTimeout,
		EnableCompression: config.EnableCompression,
		CheckOrigin:       config.CheckOrigin,
	}
}
