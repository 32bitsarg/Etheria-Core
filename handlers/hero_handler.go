package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"server-backend/models"
	"server-backend/repository"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type HeroHandler struct {
	heroRepo *repository.HeroRepository
	logger   *zap.Logger
}

func NewHeroHandler(heroRepo *repository.HeroRepository, logger *zap.Logger) *HeroHandler {
	return &HeroHandler{
		heroRepo: heroRepo,
		logger:   logger,
	}
}

// GetHeroes obtiene todos los héroes disponibles
func (h *HeroHandler) GetHeroes(w http.ResponseWriter, r *http.Request) {
	// Verificar si el sistema está habilitado
	config, err := h.heroRepo.GetHeroSystemConfig()
	if err != nil {
		h.logger.Error("Error obteniendo configuración del sistema de héroes", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	if !config.IsEnabled {
		http.Error(w, "El sistema de héroes está deshabilitado", http.StatusServiceUnavailable)
		return
	}

	// Obtener filtros de query
	race := r.URL.Query().Get("race")
	class := r.URL.Query().Get("class")
	rarity := r.URL.Query().Get("rarity")

	heroes, err := h.heroRepo.GetHeroes(race, class, rarity)
	if err != nil {
		h.logger.Error("Error obteniendo héroes", zap.Error(err))
		http.Error(w, "Error obteniendo héroes", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"data":    heroes,
		"count":   len(heroes),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetHero obtiene un héroe específico
func (h *HeroHandler) GetHero(w http.ResponseWriter, r *http.Request) {
	// Verificar si el sistema está habilitado
	config, err := h.heroRepo.GetHeroSystemConfig()
	if err != nil {
		h.logger.Error("Error obteniendo configuración del sistema de héroes", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	if !config.IsEnabled {
		http.Error(w, "El sistema de héroes está deshabilitado", http.StatusServiceUnavailable)
		return
	}

	heroID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "ID de héroe inválido", http.StatusBadRequest)
		return
	}

	hero, err := h.heroRepo.GetHero(heroID)
	if err != nil {
		h.logger.Error("Error obteniendo héroe", zap.Error(err), zap.Int("hero_id", heroID))
		http.Error(w, "Héroe no encontrado", http.StatusNotFound)
		return
	}

	// Obtener detalles adicionales
	skills, _ := h.heroRepo.GetHeroSkills(heroID)
	equipment, _ := h.heroRepo.GetHeroEquipment(heroID)
	quests, _ := h.heroRepo.GetHeroQuests(heroID)

	heroWithDetails := models.HeroWithDetails{
		Hero:      hero,
		Skills:    skills,
		Equipment: equipment,
		Quests:    quests,
	}

	response := map[string]interface{}{
		"success": true,
		"data":    heroWithDetails,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetPlayerHeroes obtiene los héroes de un jugador
func (h *HeroHandler) GetPlayerHeroes(w http.ResponseWriter, r *http.Request) {
	// Verificar si el sistema está habilitado
	config, err := h.heroRepo.GetHeroSystemConfig()
	if err != nil {
		h.logger.Error("Error obteniendo configuración del sistema de héroes", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	if !config.IsEnabled {
		http.Error(w, "El sistema de héroes está deshabilitado", http.StatusServiceUnavailable)
		return
	}

	// Obtener playerID del contexto (asumiendo que viene del middleware de autenticación)
	playerID := r.Context().Value("player_id").(int)

	playerHeroes, err := h.heroRepo.GetPlayerHeroes(playerID)
	if err != nil {
		h.logger.Error("Error obteniendo héroes del jugador", zap.Error(err), zap.Int("player_id", playerID))
		http.Error(w, "Error obteniendo héroes", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"data":    playerHeroes,
		"count":   len(playerHeroes),
		"config": map[string]interface{}{
			"max_heroes":        config.MaxHeroesPerPlayer,
			"max_active_heroes": config.MaxActiveHeroes,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetPlayerHero obtiene un héroe específico de un jugador
func (h *HeroHandler) GetPlayerHero(w http.ResponseWriter, r *http.Request) {
	// Verificar si el sistema está habilitado
	config, err := h.heroRepo.GetHeroSystemConfig()
	if err != nil {
		h.logger.Error("Error obteniendo configuración del sistema de héroes", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	if !config.IsEnabled {
		http.Error(w, "El sistema de héroes está deshabilitado", http.StatusServiceUnavailable)
		return
	}

	playerID := r.Context().Value("player_id").(int)
	heroID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "ID de héroe inválido", http.StatusBadRequest)
		return
	}

	playerHero, err := h.heroRepo.GetPlayerHero(playerID, heroID)
	if err != nil {
		h.logger.Error("Error obteniendo héroe del jugador", zap.Error(err), zap.Int("player_id", playerID), zap.Int("hero_id", heroID))
		http.Error(w, "Error obteniendo héroe", http.StatusInternalServerError)
		return
	}

	if playerHero == nil {
		http.Error(w, "No tienes este héroe", http.StatusNotFound)
		return
	}

	// Obtener información del héroe base
	hero, err := h.heroRepo.GetHero(heroID)
	if err != nil {
		h.logger.Error("Error obteniendo héroe base", zap.Error(err), zap.Int("hero_id", heroID))
		http.Error(w, "Error obteniendo información del héroe", http.StatusInternalServerError)
		return
	}

	heroWithDetails := models.HeroWithDetails{
		Hero:       hero,
		PlayerHero: playerHero,
	}

	response := map[string]interface{}{
		"success": true,
		"data":    heroWithDetails,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// RecruitHero recluta un héroe
func (h *HeroHandler) RecruitHero(w http.ResponseWriter, r *http.Request) {
	// Verificar si el sistema está habilitado
	config, err := h.heroRepo.GetHeroSystemConfig()
	if err != nil {
		h.logger.Error("Error obteniendo configuración del sistema de héroes", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	if !config.IsEnabled {
		http.Error(w, "El sistema de héroes está deshabilitado", http.StatusServiceUnavailable)
		return
	}

	playerID := r.Context().Value("player_id").(int)
	heroID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "ID de héroe inválido", http.StatusBadRequest)
		return
	}

	err = h.heroRepo.RecruitHero(playerID, heroID)
	if err != nil {
		h.logger.Error("Error reclutando héroe", zap.Error(err), zap.Int("player_id", playerID), zap.Int("hero_id", heroID))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Héroe reclutado exitosamente",
		"data": map[string]interface{}{
			"hero_id":   heroID,
			"player_id": playerID,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpgradeHero mejora un héroe
func (h *HeroHandler) UpgradeHero(w http.ResponseWriter, r *http.Request) {
	// Verificar si el sistema está habilitado
	config, err := h.heroRepo.GetHeroSystemConfig()
	if err != nil {
		h.logger.Error("Error obteniendo configuración del sistema de héroes", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	if !config.IsEnabled {
		http.Error(w, "El sistema de héroes está deshabilitado", http.StatusServiceUnavailable)
		return
	}

	playerID := r.Context().Value("player_id").(int)
	heroID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "ID de héroe inválido", http.StatusBadRequest)
		return
	}

	err = h.heroRepo.UpgradeHero(playerID, heroID)
	if err != nil {
		h.logger.Error("Error mejorando héroe", zap.Error(err), zap.Int("player_id", playerID), zap.Int("hero_id", heroID))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Héroe mejorado exitosamente",
		"data": map[string]interface{}{
			"hero_id":   heroID,
			"player_id": playerID,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ActivateHero activa un héroe
func (h *HeroHandler) ActivateHero(w http.ResponseWriter, r *http.Request) {
	// Verificar si el sistema está habilitado
	config, err := h.heroRepo.GetHeroSystemConfig()
	if err != nil {
		h.logger.Error("Error obteniendo configuración del sistema de héroes", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	if !config.IsEnabled {
		http.Error(w, "El sistema de héroes está deshabilitado", http.StatusServiceUnavailable)
		return
	}

	playerID := r.Context().Value("player_id").(int)
	heroID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "ID de héroe inválido", http.StatusBadRequest)
		return
	}

	err = h.heroRepo.ActivateHero(playerID, heroID)
	if err != nil {
		h.logger.Error("Error activando héroe", zap.Error(err), zap.Int("player_id", playerID), zap.Int("hero_id", heroID))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Héroe activado exitosamente",
		"data": map[string]interface{}{
			"hero_id":   heroID,
			"player_id": playerID,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DeactivateHero desactiva un héroe
func (h *HeroHandler) DeactivateHero(w http.ResponseWriter, r *http.Request) {
	// Verificar si el sistema está habilitado
	config, err := h.heroRepo.GetHeroSystemConfig()
	if err != nil {
		h.logger.Error("Error obteniendo configuración del sistema de héroes", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	if !config.IsEnabled {
		http.Error(w, "El sistema de héroes está deshabilitado", http.StatusServiceUnavailable)
		return
	}

	playerID := r.Context().Value("player_id").(int)
	heroID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "ID de héroe inválido", http.StatusBadRequest)
		return
	}

	err = h.heroRepo.DeactivateHero(playerID, heroID)
	if err != nil {
		h.logger.Error("Error desactivando héroe", zap.Error(err), zap.Int("player_id", playerID), zap.Int("hero_id", heroID))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Héroe desactivado exitosamente",
		"data": map[string]interface{}{
			"hero_id":   heroID,
			"player_id": playerID,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetActiveHeroes obtiene los héroes activos de un jugador
func (h *HeroHandler) GetActiveHeroes(w http.ResponseWriter, r *http.Request) {
	// Verificar si el sistema está habilitado
	config, err := h.heroRepo.GetHeroSystemConfig()
	if err != nil {
		h.logger.Error("Error obteniendo configuración del sistema de héroes", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	if !config.IsEnabled {
		http.Error(w, "El sistema de héroes está deshabilitado", http.StatusServiceUnavailable)
		return
	}

	playerID := r.Context().Value("player_id").(int)

	activeHeroes, err := h.heroRepo.GetActiveHeroes(playerID)
	if err != nil {
		h.logger.Error("Error obteniendo héroes activos", zap.Error(err), zap.Int("player_id", playerID))
		http.Error(w, "Error obteniendo héroes activos", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"data":    activeHeroes,
		"count":   len(activeHeroes),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetHeroRankings obtiene los rankings de héroes
func (h *HeroHandler) GetHeroRankings(w http.ResponseWriter, r *http.Request) {
	// Verificar si el sistema está habilitado
	config, err := h.heroRepo.GetHeroSystemConfig()
	if err != nil {
		h.logger.Error("Error obteniendo configuración del sistema de héroes", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	if !config.IsEnabled {
		http.Error(w, "El sistema de héroes está deshabilitado", http.StatusServiceUnavailable)
		return
	}

	limit := 100 // Por defecto
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 1000 {
			limit = l
		}
	}

	rankings, err := h.heroRepo.GetHeroRankings(limit)
	if err != nil {
		h.logger.Error("Error obteniendo rankings de héroes", zap.Error(err))
		http.Error(w, "Error obteniendo rankings", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"data":    rankings,
		"count":   len(rankings),
		"limit":   limit,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetHeroSystemConfig obtiene la configuración del sistema de héroes
func (h *HeroHandler) GetHeroSystemConfig(w http.ResponseWriter, r *http.Request) {
	config, err := h.heroRepo.GetHeroSystemConfig()
	if err != nil {
		h.logger.Error("Error obteniendo configuración del sistema de héroes", zap.Error(err))
		http.Error(w, "Error obteniendo configuración", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"data":    config,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateHeroSystemConfig actualiza la configuración del sistema de héroes (solo admin)
func (h *HeroHandler) UpdateHeroSystemConfig(w http.ResponseWriter, r *http.Request) {
	// Verificar si es admin (asumiendo que viene del middleware)
	isAdmin := r.Context().Value("is_admin").(bool)
	if !isAdmin {
		http.Error(w, "Acceso denegado", http.StatusForbidden)
		return
	}

	var config models.HeroSystemConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		http.Error(w, "Datos inválidos", http.StatusBadRequest)
		return
	}

	// Obtener la configuración actual para mantener el ID
	currentConfig, err := h.heroRepo.GetHeroSystemConfig()
	if err != nil {
		h.logger.Error("Error obteniendo configuración actual", zap.Error(err))
		http.Error(w, "Error obteniendo configuración actual", http.StatusInternalServerError)
		return
	}

	config.ID = currentConfig.ID
	config.UpdatedAt = time.Now()

	err = h.heroRepo.UpdateHeroSystemConfig(&config)
	if err != nil {
		h.logger.Error("Error actualizando configuración del sistema de héroes", zap.Error(err))
		http.Error(w, "Error actualizando configuración", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Configuración actualizada exitosamente",
		"data":    config,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetHeroProgress obtiene el progreso de un héroe
func (h *HeroHandler) GetHeroProgress(w http.ResponseWriter, r *http.Request) {
	// Verificar si el sistema está habilitado
	config, err := h.heroRepo.GetHeroSystemConfig()
	if err != nil {
		h.logger.Error("Error obteniendo configuración del sistema de héroes", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	if !config.IsEnabled {
		http.Error(w, "El sistema de héroes está deshabilitado", http.StatusServiceUnavailable)
		return
	}

	playerID := r.Context().Value("player_id").(int)
	heroID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "ID de héroe inválido", http.StatusBadRequest)
		return
	}

	playerHero, err := h.heroRepo.GetPlayerHero(playerID, heroID)
	if err != nil {
		h.logger.Error("Error obteniendo héroe del jugador", zap.Error(err), zap.Int("player_id", playerID), zap.Int("hero_id", heroID))
		http.Error(w, "Error obteniendo héroe", http.StatusInternalServerError)
		return
	}

	if playerHero == nil {
		http.Error(w, "No tienes este héroe", http.StatusNotFound)
		return
	}

	// Obtener información del héroe base
	hero, err := h.heroRepo.GetHero(heroID)
	if err != nil {
		h.logger.Error("Error obteniendo héroe base", zap.Error(err), zap.Int("hero_id", heroID))
		http.Error(w, "Error obteniendo información del héroe", http.StatusInternalServerError)
		return
	}

	// Calcular progreso
	progress := float64(playerHero.Experience) / float64(playerHero.ExperienceToNext) * 100
	if progress > 100 {
		progress = 100
	}

	// Calcular tiempo restante de lesión
	injuryTimeRemaining := 0
	if playerHero.IsInjured && playerHero.InjuryTime != nil {
		recoveryTime := playerHero.InjuryTime.Add(time.Duration(config.InjuryDuration) * time.Second)
		if time.Now().Before(recoveryTime) {
			injuryTimeRemaining = int(recoveryTime.Sub(time.Now()).Seconds())
		}
	}

	// Calcular win rate
	winRate := 0.0
	if playerHero.BattlesWon+playerHero.BattlesLost > 0 {
		winRate = float64(playerHero.BattlesWon) / float64(playerHero.BattlesWon+playerHero.BattlesLost) * 100
	}

	// Calcular poder total
	totalPower := playerHero.CurrentAttack + playerHero.CurrentDefense + playerHero.CurrentSpeed +
		playerHero.CurrentIntelligence + playerHero.CurrentCharisma

	heroProgress := models.HeroProgress{
		HeroID:              heroID,
		HeroName:            hero.Name,
		Level:               playerHero.Level,
		Experience:          playerHero.Experience,
		ExperienceToNext:    playerHero.ExperienceToNext,
		Progress:            progress,
		TotalPower:          totalPower,
		BattlesWon:          playerHero.BattlesWon,
		BattlesLost:         playerHero.BattlesLost,
		WinRate:             winRate,
		QuestsCompleted:     playerHero.QuestsCompleted,
		IsActive:            playerHero.IsActive,
		IsInjured:           playerHero.IsInjured,
		InjuryTimeRemaining: injuryTimeRemaining,
	}

	response := map[string]interface{}{
		"success": true,
		"data":    heroProgress,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
