package routes

import (
	"server-backend/handlers"
	"server-backend/middleware"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// SetupChatRoutes configura todas las rutas relacionadas con chat
func SetupChatRoutes(r *gin.RouterGroup, chatHandler *handlers.ChatHandler, authMiddleware *middleware.AuthMiddleware, logger *zap.Logger) {
	// Grupo de rutas de chat (ya protegido por el grupo padre)
	chatGroup := r.Group("/chat")
	chatGroup.Use(middleware.WebSocketLoggingMiddleware(logger))

	// Canales
	chatGroup.GET("/channels", chatHandler.GetChannels)
	chatGroup.POST("/channels/alliance", chatHandler.CreateAllianceChannel)
	chatGroup.GET("/channels/info", chatHandler.GetChannelInfo)

	// Mensajes
	chatGroup.POST("/messages", chatHandler.SendMessage)
	chatGroup.GET("/messages", chatHandler.GetRecentMessages)

	// Usuarios
	chatGroup.POST("/join", chatHandler.JoinChannel)
	chatGroup.POST("/leave", chatHandler.LeaveChannel)
	chatGroup.GET("/users", chatHandler.GetOnlineUsers)

	// WebSocket para tiempo real (con middleware específico)
	chatGroup.GET("/ws", authMiddleware.RequireAuthGinWebSocket(), chatHandler.WebSocketChat)

	// Moderación
	chatGroup.POST("/ban", chatHandler.BanUser)
	chatGroup.POST("/system", chatHandler.SendSystemMessage)

	// Estadísticas
	chatGroup.GET("/stats", chatHandler.GetChatStats)

	logger.Info("✅ Rutas de chat configuradas exitosamente")
}
