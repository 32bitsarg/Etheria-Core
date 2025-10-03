package middleware

import (
	"context"
	"net/http"
	"strings"

	"server-backend/auth"
	"server-backend/repository"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type AuthMiddleware struct {
	jwtManager *auth.JWTManager
	playerRepo *repository.PlayerRepository
	logger     *zap.Logger
}

func NewAuthMiddleware(jwtManager *auth.JWTManager, playerRepo *repository.PlayerRepository, logger *zap.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		jwtManager: jwtManager,
		playerRepo: playerRepo,
		logger:     logger,
	}
}

func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Obtener el token del header Authorization
		authHeader := r.Header.Get("Authorization")
		
		// 🔍 DEBUG: Log de la request entrante
		m.logger.Info("🔍 DEBUG MIDDLEWARE - Request entrante",
			zap.String("method", r.Method),
			zap.String("url", r.URL.String()),
			zap.String("path", r.URL.Path),
			zap.String("remote_addr", r.RemoteAddr),
			zap.String("auth_header", authHeader))

		if authHeader == "" {
			m.logger.Error("🔍 DEBUG MIDDLEWARE - No se proporcionó token de autenticación",
				zap.String("path", r.URL.Path))
			http.Error(w, "No se proporcionó token de autenticación", http.StatusUnauthorized)
			return
		}

		// Verificar que el header tenga el formato correcto
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			m.logger.Error("🔍 DEBUG MIDDLEWARE - Formato de token inválido",
				zap.String("path", r.URL.Path),
				zap.String("auth_header", authHeader))
			http.Error(w, "Formato de token inválido", http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]

		// Validar el token JWT
		claims, err := m.jwtManager.VerifyToken(tokenString)
		if err != nil {
			m.logger.Error("🔍 DEBUG MIDDLEWARE - Error validando token",
				zap.String("path", r.URL.Path),
				zap.Error(err))
			http.Error(w, "Token inválido", http.StatusUnauthorized)
			return
		}

		// 🔍 DEBUG: Token válido
		m.logger.Info("🔍 DEBUG MIDDLEWARE - Token válido, continuando",
			zap.String("path", r.URL.Path),
			zap.String("user_id", claims.UserID),
			zap.String("username", claims.Username))

		// Agregar la información del jugador al contexto
		ctx := r.Context()
		ctx = context.WithValue(ctx, "player_id", claims.UserID)
		ctx = context.WithValue(ctx, "username", claims.Username)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireAuthGin - Middleware de autenticación nativo para Gin
// RequireAuthGinWebSocket es un middleware específico para WebSockets con mejor manejo de errores
func (m *AuthMiddleware) RequireAuthGinWebSocket() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Obtener el token del header Authorization
		authHeader := c.GetHeader("Authorization")
		
		// 🔍 DEBUG: Log de la request entrante
		m.logger.Info("🔍 DEBUG MIDDLEWARE WEBSOCKET - Request entrante",
			zap.String("method", c.Request.Method),
			zap.String("url", c.Request.URL.String()),
			zap.String("path", c.Request.URL.Path),
			zap.String("remote_addr", c.Request.RemoteAddr),
			zap.String("auth_header", authHeader))

		if authHeader == "" {
			m.logger.Error("🔍 DEBUG MIDDLEWARE WEBSOCKET - No se proporcionó token de autenticación",
				zap.String("path", c.Request.URL.Path))
			c.JSON(401, gin.H{"error": "No se proporcionó token de autenticación"})
			c.Abort()
			return
		}

		// Verificar que el header tenga el formato correcto
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			m.logger.Error("🔍 DEBUG MIDDLEWARE WEBSOCKET - Formato de token inválido",
				zap.String("path", c.Request.URL.Path),
				zap.String("auth_header", authHeader))
			c.JSON(401, gin.H{"error": "Formato de token inválido"})
			c.Abort()
			return
		}

		token := parts[1]
		
		// Verificar si es un token dummy
		if token == "dummy_token" {
			m.logger.Error("🔍 DEBUG MIDDLEWARE WEBSOCKET - Token dummy detectado",
				zap.String("path", c.Request.URL.Path))
			c.JSON(401, gin.H{"error": "Token de prueba no válido. Por favor, inicie sesión nuevamente."})
			c.Abort()
			return
		}

		// Validar el token
		claims, err := m.jwtManager.VerifyToken(token)
		if err != nil {
			m.logger.Error("🔍 DEBUG MIDDLEWARE WEBSOCKET - Error validando token",
				zap.String("path", c.Request.URL.Path),
				zap.Error(err))
			c.JSON(401, gin.H{"error": "Token inválido"})
			c.Abort()
			return
		}

		// Obtener el player_id del token
		playerIDStr := claims.UserID

		// Obtener el username del token
		username := claims.Username

		// Establecer valores en el contexto de Gin
		c.Set("player_id", playerIDStr)
		c.Set("username", username)
		c.Set("playerID", playerIDStr) // Para compatibilidad con handlers existentes

		m.logger.Info("🔍 DEBUG MIDDLEWARE WEBSOCKET - Token válido, continuando",
			zap.String("path", c.Request.URL.Path),
			zap.String("user_id", playerIDStr),
			zap.String("username", username))

		c.Next()
	}
}

func (m *AuthMiddleware) RequireAuthGin() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Obtener el token del header Authorization
		authHeader := c.GetHeader("Authorization")
		
		// 🔍 DEBUG: Log de la request entrante
		m.logger.Info("🔍 DEBUG MIDDLEWARE GIN - Request entrante",
			zap.String("method", c.Request.Method),
			zap.String("url", c.Request.URL.String()),
			zap.String("path", c.Request.URL.Path),
			zap.String("remote_addr", c.Request.RemoteAddr),
			zap.String("auth_header", authHeader))

		if authHeader == "" {
			m.logger.Error("🔍 DEBUG MIDDLEWARE GIN - No se proporcionó token de autenticación",
				zap.String("path", c.Request.URL.Path))
			c.JSON(401, gin.H{"error": "No se proporcionó token de autenticación"})
			c.Abort()
			return
		}

		// Verificar que el header tenga el formato correcto
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			m.logger.Error("🔍 DEBUG MIDDLEWARE GIN - Formato de token inválido",
				zap.String("path", c.Request.URL.Path),
				zap.String("auth_header", authHeader))
			c.JSON(401, gin.H{"error": "Formato de token inválido"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Validar el token JWT
		claims, err := m.jwtManager.VerifyToken(tokenString)
		if err != nil {
			m.logger.Error("🔍 DEBUG MIDDLEWARE GIN - Error validando token",
				zap.String("path", c.Request.URL.Path),
				zap.Error(err))
			c.JSON(401, gin.H{"error": "Token inválido"})
			c.Abort()
			return
		}

		// 🔍 DEBUG: Token válido
		m.logger.Info("🔍 DEBUG MIDDLEWARE GIN - Token válido, continuando",
			zap.String("path", c.Request.URL.Path),
			zap.String("user_id", claims.UserID),
			zap.String("username", claims.Username))

		// Establecer la información del usuario en el contexto Gin
		c.Set("player_id", claims.UserID)  // Mantener compatibilidad con handlers existentes
		c.Set("user_id", claims.UserID)    // Nuevo formato estándar
		c.Set("username", claims.Username)

		c.Next()
	}
}

// RequireAdmin verifica que el usuario tenga rol de administrador
func (m *AuthMiddleware) RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Primero verificar autenticación
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "No se proporcionó token de autenticación", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Formato de token inválido", http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]

		// Validar el token JWT
		claims, err := m.jwtManager.VerifyToken(tokenString)
		if err != nil {
			m.logger.Error("Error validando token", zap.Error(err))
			http.Error(w, "Token inválido", http.StatusUnauthorized)
			return
		}

		// Obtener el usuario de la base de datos para verificar su rol
		playerID, err := uuid.Parse(claims.UserID)
		if err != nil {
			m.logger.Error("Error parseando UUID del usuario", zap.Error(err))
			http.Error(w, "Token inválido", http.StatusUnauthorized)
			return
		}

		player, err := m.playerRepo.GetPlayerByID(playerID)
		if err != nil {
			m.logger.Error("Error obteniendo usuario", zap.Error(err))
			http.Error(w, "Usuario no encontrado", http.StatusUnauthorized)
			return
		}

		// Verificar que el usuario tenga rol de administrador
		if player.Role != "admin" {
			m.logger.Warn("Usuario intentó acceder a endpoint de administración sin permisos",
				zap.String("username", player.Username),
				zap.String("role", player.Role))
			http.Error(w, "Acceso denegado. Se requieren permisos de administrador", http.StatusForbidden)
			return
		}

		// Agregar la información del jugador al contexto
		ctx := r.Context()
		ctx = context.WithValue(ctx, "player_id", claims.UserID)
		ctx = context.WithValue(ctx, "username", claims.Username)
		ctx = context.WithValue(ctx, "role", player.Role)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
