package middleware

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type WebSocketValidator struct {
	logger *zap.Logger
}

func NewWebSocketValidator(logger *zap.Logger) *WebSocketValidator {
	return &WebSocketValidator{logger: logger}
}

// ValidateMessage valida y parsea mensajes WebSocket
func (w *WebSocketValidator) ValidateMessage(message []byte) (map[string]interface{}, error) {
	// Validar tamaño
	if len(message) == 0 {
		return nil, fmt.Errorf("mensaje vacío")
	}

	if len(message) > 1024 {
		return nil, fmt.Errorf("mensaje demasiado grande: %d bytes", len(message))
	}

	// Validar JSON
	var data map[string]interface{}
	if err := json.Unmarshal(message, &data); err != nil {
		return nil, fmt.Errorf("JSON inválido: %w", err)
	}

	// Validar estructura requerida
	if _, exists := data["message"]; !exists {
		return nil, fmt.Errorf("campo 'message' requerido")
	}

	if messageText, ok := data["message"].(string); !ok {
		return nil, fmt.Errorf("campo 'message' debe ser string")
	} else if len(messageText) == 0 {
		return nil, fmt.Errorf("mensaje no puede estar vacío")
	}

	return data, nil
}

// SendError envía error al cliente WebSocket
func (w *WebSocketValidator) SendError(conn *websocket.Conn, errorMsg string) {
	errorResponse := map[string]interface{}{
		"type":      "error",
		"message":   errorMsg,
		"timestamp": time.Now().Format(time.RFC3339),
	}

	if err := conn.WriteJSON(errorResponse); err != nil {
		w.logger.Error("Error enviando error al cliente", zap.Error(err))
	}
}

// SendSuccess envía confirmación de éxito al cliente WebSocket
func (w *WebSocketValidator) SendSuccess(conn *websocket.Conn, message string) {
	successResponse := map[string]interface{}{
		"type":      "success",
		"message":   message,
		"timestamp": time.Now().Format(time.RFC3339),
	}

	if err := conn.WriteJSON(successResponse); err != nil {
		w.logger.Error("Error enviando confirmación al cliente", zap.Error(err))
	}
}

// IsValidMessageType verifica si el tipo de mensaje es válido
func (w *WebSocketValidator) IsValidMessageType(messageType int) bool {
	return messageType == websocket.TextMessage
}

// LogMessageReceived registra la recepción de un mensaje para debugging
func (w *WebSocketValidator) LogMessageReceived(message []byte, clientInfo string) {
	w.logger.Debug("Mensaje WebSocket recibido",
		zap.String("client", clientInfo),
		zap.String("message", string(message)),
		zap.Int("size", len(message)),
	)
}

// LogMessageSent registra el envío de un mensaje para debugging
func (w *WebSocketValidator) LogMessageSent(message []byte, clientInfo string) {
	w.logger.Debug("Mensaje WebSocket enviado",
		zap.String("client", clientInfo),
		zap.String("message", string(message)),
		zap.Int("size", len(message)),
	)
}
