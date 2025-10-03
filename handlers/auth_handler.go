package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"server-backend/auth"
	"server-backend/models"
	"server-backend/repository"
	"server-backend/services"
)

type AuthHandler struct {
	playerRepo   *repository.PlayerRepository
	jwtManager   *auth.JWTManager
	logger       *zap.Logger
	redisService *services.RedisService
	villageRepo  *repository.VillageRepository
}

func NewAuthHandler(playerRepo *repository.PlayerRepository, jwtManager *auth.JWTManager, logger *zap.Logger, redisService *services.RedisService, villageRepo *repository.VillageRepository) *AuthHandler {
	return &AuthHandler{
		playerRepo:   playerRepo,
		jwtManager:   jwtManager,
		logger:       logger,
		redisService: redisService,
		villageRepo:  villageRepo,
	}
}

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type UpdateProfileRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
}

type AuthResponse struct {
	Token    string `json:"token"`
	UserID   string `json:"user_id"`
	Username string `json:"username"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error decodificando la solicitud"})
		return
	}

	// Verificar si el usuario ya existe
	exists, err := h.playerRepo.UsernameExists(req.Username)
	if err != nil {
		h.logger.Error("Error verificando existencia de usuario",
			zap.Error(err),
			zap.String("username", req.Username),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno del servidor"})
		return
	}
	if exists {
		c.JSON(http.StatusConflict, gin.H{"error": "El nombre de usuario ya está en uso"})
		return
	}

	// Hash de la contraseña
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		h.logger.Error("Error hasheando contraseña",
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno del servidor"})
		return
	}

	// Crear el jugador
	playerID, err := h.playerRepo.CreatePlayer(req.Username, hashedPassword, req.Email)
	if err != nil {
		h.logger.Error("Error creando jugador",
			zap.Error(err),
			zap.String("username", req.Username),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno del servidor"})
		return
	}

	// Generar token
	token, err := h.jwtManager.GenerateToken(playerID.String(), req.Username)
	if err != nil {
		h.logger.Error("Error generando token",
			zap.Error(err),
			zap.String("username", req.Username),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno del servidor"})
		return
	}

	// Enviar respuesta
	response := AuthResponse{
		Token:    token,
		UserID:   playerID.String(),
		Username: req.Username,
	}

	c.JSON(http.StatusOK, response)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error decodificando la solicitud"})
		return
	}

	// Obtener el jugador
	player, err := h.playerRepo.GetPlayerByUsername(req.Username)
	if err != nil {
		h.logger.Error("Error obteniendo jugador",
			zap.Error(err),
			zap.String("username", req.Username),
		)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Credenciales inválidas"})
		return
	}

	// Verificar que el jugador existe
	if player == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Credenciales inválidas"})
		return
	}

	// Verificar contraseña
	if !auth.CheckPasswordHash(req.Password, player.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Credenciales inválidas"})
		return
	}

	// Generar token
	token, err := h.jwtManager.GenerateToken(player.ID.String(), player.Username)
	if err != nil {
		h.logger.Error("Error generando token",
			zap.Error(err),
			zap.String("username", req.Username),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno del servidor"})
		return
	}

	// Guardar sesión en Redis y marcar usuario online
	if h.redisService != nil {
		session := &services.SessionData{
			UserID:     player.ID.String(),
			Username:   player.Username,
			Role:       player.Role,
			IsOnline:   true,
			LastActive: time.Now(),
			WorldID:    "",
			VillageID:  "",
			CreatedAt:  time.Now(),
		}
		h.redisService.StoreUserSession(player.ID.String(), session)
		h.redisService.SetUserOnline(player.ID.String(), player.Username)
	}

	// Enviar respuesta
	response := AuthResponse{
		Token:    token,
		UserID:   player.ID.String(),
		Username: player.Username,
	}

	c.JSON(http.StatusOK, response)
}

func (h *AuthHandler) GetProfile(c *gin.Context) {
	// Obtener el ID del jugador del contexto
	playerIDStr := c.GetString("player_id")
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		h.logger.Error("Error parseando ID de jugador", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno del servidor"})
		return
	}

	// Obtener el perfil del jugador
	player, err := h.playerRepo.GetPlayerByID(playerID)
	if err != nil {
		h.logger.Error("Error obteniendo perfil", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno del servidor"})
		return
	}
	if player == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Jugador no encontrado"})
		return
	}

	// Obtener aldeas del jugador
	villages, err := h.villageRepo.GetVillagesByPlayerID(playerID)
	if err != nil {
		h.logger.Error("Error obteniendo aldeas del jugador", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno del servidor"})
		return
	}

	h.logger.Info("Aldeas encontradas para el jugador",
		zap.String("player_id", playerID.String()),
		zap.Int("count", len(villages)))

	player.Villages = nil
	if villages != nil {
		h.logger.Info("Procesando aldeas", zap.Int("total", len(villages)))
		for i, v := range villages {
			if v == nil || v.Village.ID == uuid.Nil {
				h.logger.Warn("Aldea nula o sin ID, se omite", zap.Int("index", i))
				continue
			}
			// Validar resources
			if v.Resources.ID == uuid.Nil {
				v.Resources = models.Resources{
					Wood:  0,
					Stone: 0,
					Food:  0,
					Gold:  0,
				}
			}
			// Validar buildings
			if v.Buildings == nil {
				v.Buildings = make(map[string]*models.Building)
			}
			player.Villages = append(player.Villages, *v)
		}
	} else {
		h.logger.Warn("No se encontraron aldeas para el jugador", zap.String("player_id", playerID.String()))
	}

	c.JSON(http.StatusOK, player)
}

func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	// Obtener el ID del jugador del contexto
	playerIDStr := c.GetString("player_id")
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		h.logger.Error("Error parseando ID de jugador", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno del servidor"})
		return
	}

	// Decodificar la solicitud
	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error decodificando la solicitud"})
		return
	}

	// Validar datos de entrada
	if strings.TrimSpace(req.Username) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "El nombre de usuario es requerido"})
		return
	}
	if strings.TrimSpace(req.Email) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "El email es requerido"})
		return
	}

	// Actualizar el perfil
	err = h.playerRepo.UpdatePlayer(playerID, req.Username, req.Email)
	if err != nil {
		h.logger.Error("Error actualizando perfil", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno del servidor"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Perfil actualizado exitosamente"})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	playerIDStr := c.GetString("player_id")
	if h.redisService != nil {
		h.redisService.DeleteUserSession(playerIDStr)
		h.redisService.SetUserOffline(playerIDStr)
	}
	c.JSON(http.StatusOK, gin.H{"message": "Logout exitoso"})
}
