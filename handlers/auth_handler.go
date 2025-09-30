package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"server-backend/auth"
	"server-backend/models"
	"server-backend/repository"
	"server-backend/services"

	"github.com/google/uuid"
	"go.uber.org/zap"
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

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Error decodificando la solicitud", http.StatusBadRequest)
		return
	}

	// Verificar si el usuario ya existe
	exists, err := h.playerRepo.UsernameExists(req.Username)
	if err != nil {
		h.logger.Error("Error verificando existencia de usuario",
			zap.Error(err),
			zap.String("username", req.Username),
		)
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}
	if exists {
		http.Error(w, "El nombre de usuario ya está en uso", http.StatusConflict)
		return
	}

	// Hash de la contraseña
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		h.logger.Error("Error hasheando contraseña",
			zap.Error(err),
		)
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	// Crear el jugador
	playerID, err := h.playerRepo.CreatePlayer(req.Username, hashedPassword, req.Email)
	if err != nil {
		h.logger.Error("Error creando jugador",
			zap.Error(err),
			zap.String("username", req.Username),
		)
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	// Generar token
	token, err := h.jwtManager.GenerateToken(playerID.String(), req.Username)
	if err != nil {
		h.logger.Error("Error generando token",
			zap.Error(err),
			zap.String("username", req.Username),
		)
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	// Enviar respuesta
	response := AuthResponse{
		Token:    token,
		UserID:   playerID.String(),
		Username: req.Username,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Error decodificando la solicitud", http.StatusBadRequest)
		return
	}

	// Obtener el jugador
	player, err := h.playerRepo.GetPlayerByUsername(req.Username)
	if err != nil {
		h.logger.Error("Error obteniendo jugador",
			zap.Error(err),
			zap.String("username", req.Username),
		)
		http.Error(w, "Credenciales inválidas", http.StatusUnauthorized)
		return
	}

	// Verificar que el jugador existe
	if player == nil {
		http.Error(w, "Credenciales inválidas", http.StatusUnauthorized)
		return
	}

	// Verificar contraseña
	if !auth.CheckPasswordHash(req.Password, player.Password) {
		http.Error(w, "Credenciales inválidas", http.StatusUnauthorized)
		return
	}

	// Generar token
	token, err := h.jwtManager.GenerateToken(player.ID.String(), player.Username)
	if err != nil {
		h.logger.Error("Error generando token",
			zap.Error(err),
			zap.String("username", req.Username),
		)
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *AuthHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	// Obtener el ID del jugador del contexto
	playerIDStr := r.Context().Value("player_id").(string)
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		h.logger.Error("Error parseando ID de jugador", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	// Obtener el perfil del jugador
	player, err := h.playerRepo.GetPlayerByID(playerID)
	if err != nil {
		h.logger.Error("Error obteniendo perfil", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}
	if player == nil {
		http.Error(w, "Jugador no encontrado", http.StatusNotFound)
		return
	}

	// Obtener aldeas del jugador
	villages, err := h.villageRepo.GetVillagesByPlayerID(playerID)
	if err != nil {
		h.logger.Error("Error obteniendo aldeas del jugador", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(player)
}

func (h *AuthHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	// Obtener el ID del jugador del contexto
	playerIDStr := r.Context().Value("player_id").(string)
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		h.logger.Error("Error parseando ID de jugador", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	// Decodificar la solicitud
	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Error decodificando la solicitud", http.StatusBadRequest)
		return
	}

	// Validar datos de entrada
	if strings.TrimSpace(req.Username) == "" {
		http.Error(w, "El nombre de usuario es requerido", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.Email) == "" {
		http.Error(w, "El email es requerido", http.StatusBadRequest)
		return
	}

	// Actualizar el perfil
	err = h.playerRepo.UpdatePlayer(playerID, req.Username, req.Email)
	if err != nil {
		h.logger.Error("Error actualizando perfil", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Perfil actualizado exitosamente",
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	playerIDStr := r.Context().Value("player_id").(string)
	if h.redisService != nil {
		h.redisService.DeleteUserSession(playerIDStr)
		h.redisService.SetUserOffline(playerIDStr)
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Logout exitoso"))
}
