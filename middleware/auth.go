package middleware

import (
	"context"
	"net/http"
	"strings"

	"server-backend/auth"
	"server-backend/repository"

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
		if authHeader == "" {
			http.Error(w, "No se proporcionó token de autenticación", http.StatusUnauthorized)
			return
		}

		// Verificar que el header tenga el formato correcto
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

		// Agregar la información del jugador al contexto
		ctx := r.Context()
		ctx = context.WithValue(ctx, "player_id", claims.UserID)
		ctx = context.WithValue(ctx, "username", claims.Username)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
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
