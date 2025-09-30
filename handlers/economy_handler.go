package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"server-backend/models"
	"server-backend/repository"
	"server-backend/services"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type EconomyHandler struct {
	economyService *services.EconomyService
	economyRepo    *repository.EconomyRepository
	logger         *zap.Logger
}

func NewEconomyHandler(economyService *services.EconomyService, economyRepo *repository.EconomyRepository, logger *zap.Logger) *EconomyHandler {
	return &EconomyHandler{
		economyService: economyService,
		economyRepo:    economyRepo,
		logger:         logger,
	}
}

// GetEconomyDashboard obtiene el dashboard completo de economía
func (h *EconomyHandler) GetEconomyDashboard(w http.ResponseWriter, r *http.Request) {
	// Obtener playerID del contexto (asumiendo que viene del middleware de autenticación)
	playerIDStr := r.Context().Value("player_id").(string)
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		h.logger.Error("ID de jugador inválido", zap.Error(err))
		http.Error(w, "ID de jugador inválido", http.StatusBadRequest)
		return
	}

	dashboard, err := h.economyService.GetEconomyDashboard(playerID)
	if err != nil {
		h.logger.Error("Error obteniendo dashboard de economía", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dashboard)
}

// GetEconomyConfig obtiene la configuración del sistema de economía
func (h *EconomyHandler) GetEconomyConfig(w http.ResponseWriter, r *http.Request) {
	config, err := h.economyService.GetEconomyConfig()
	if err != nil {
		h.logger.Error("Error obteniendo configuración", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
}

// UpdateEconomyConfig actualiza la configuración del sistema de economía
func (h *EconomyHandler) UpdateEconomyConfig(w http.ResponseWriter, r *http.Request) {
	var config models.EconomySystemConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		h.logger.Error("Error decodificando configuración", zap.Error(err))
		http.Error(w, "Datos inválidos", http.StatusBadRequest)
		return
	}

	// Validar configuración
	if err := h.validateEconomyConfig(&config); err != nil {
		h.logger.Error("Configuración inválida", zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Obtener configuración actual para preservar el ID
	currentConfig, err := h.economyRepo.GetEconomySystemConfig()
	if err != nil {
		h.logger.Error("Error obteniendo configuración actual", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	config.ID = currentConfig.ID
	config.UpdatedAt = time.Now()

	err = h.economyRepo.UpdateEconomySystemConfig(&config)
	if err != nil {
		h.logger.Error("Error actualizando configuración", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Configuración actualizada exitosamente"})
}

// GetPlayerEconomy obtiene la economía de un jugador
func (h *EconomyHandler) GetPlayerEconomy(w http.ResponseWriter, r *http.Request) {
	playerIDStr := chi.URLParam(r, "playerID")
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		h.logger.Error("ID de jugador inválido", zap.Error(err))
		http.Error(w, "ID de jugador inválido", http.StatusBadRequest)
		return
	}

	economy, err := h.economyService.GetPlayerEconomy(playerID)
	if err != nil {
		h.logger.Error("Error obteniendo economía del jugador", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, economy)
}

// GetMarketItems obtiene los items del mercado
func (h *EconomyHandler) GetMarketItems(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	filters := map[string]interface{}{
		"category": query.Get("category"),
		"rarity":   query.Get("rarity"),
	}

	if limitStr := query.Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			filters["limit"] = limit
		}
	}

	items, err := h.economyService.GetMarketItems(filters)
	if err != nil {
		h.logger.Error("Error obteniendo items del mercado", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, items)
}

// GetMarketItemDetails obtiene detalles completos de un item del mercado
func (h *EconomyHandler) GetMarketItemDetails(w http.ResponseWriter, r *http.Request) {
	itemIDStr := chi.URLParam(r, "itemID")
	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		h.logger.Error("ID de item inválido", zap.Error(err))
		http.Error(w, "ID de item inválido", http.StatusBadRequest)
		return
	}

	details, err := h.economyService.GetMarketItemDetails(itemID)
	if err != nil {
		h.logger.Error("Error obteniendo detalles del item", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, details)
}

// CreateMarketListing crea una nueva lista de venta
func (h *EconomyHandler) CreateMarketListing(w http.ResponseWriter, r *http.Request) {
	// Obtener playerID del contexto
	playerIDStr := r.Context().Value("player_id").(string)
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		h.logger.Error("ID de jugador inválido", zap.Error(err))
		http.Error(w, "ID de jugador inválido", http.StatusBadRequest)
		return
	}

	var listing models.MarketListing
	if err := json.NewDecoder(r.Body).Decode(&listing); err != nil {
		h.logger.Error("Error decodificando lista", zap.Error(err))
		http.Error(w, "Datos inválidos", http.StatusBadRequest)
		return
	}

	// Configurar valores por defecto
	listing.ID = uuid.New().String()
	listing.SellerID = playerID.String()
	listing.CreatedAt = time.Now()
	listing.ExpiresAt = listing.CreatedAt.Add(24 * time.Hour) // Por defecto 24 horas
	listing.IsActive = true

	err = h.economyService.CreateMarketListing(playerID, &listing)
	if err != nil {
		h.logger.Error("Error creando lista", zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	respondWithJSON(w, http.StatusCreated, listing)
}

// GetMarketListings obtiene las ofertas del mercado
func (h *EconomyHandler) GetMarketListings(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	filters := map[string]interface{}{
		"item_type": query.Get("item_type"),
		"status":    query.Get("status"),
	}

	if itemIDStr := query.Get("item_id"); itemIDStr != "" {
		if itemID, err := strconv.Atoi(itemIDStr); err == nil {
			filters["item_id"] = itemID
		}
	}

	if sellerIDStr := query.Get("seller_id"); sellerIDStr != "" {
		if sellerID, err := strconv.Atoi(sellerIDStr); err == nil {
			filters["seller_id"] = sellerID
		}
	}

	if limitStr := query.Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			filters["limit"] = limit
		}
	}

	listings, err := h.economyService.GetMarketListings(filters)
	if err != nil {
		h.logger.Error("Error obteniendo ofertas", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, listings)
}

// ExchangeCurrency intercambia monedas
func (h *EconomyHandler) ExchangeCurrency(w http.ResponseWriter, r *http.Request) {
	// Obtener playerID del contexto
	playerIDStr := r.Context().Value("player_id").(string)
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		h.logger.Error("ID de jugador inválido", zap.Error(err))
		http.Error(w, "ID de jugador inválido", http.StatusBadRequest)
		return
	}

	var exchange models.CurrencyExchange
	if err := json.NewDecoder(r.Body).Decode(&exchange); err != nil {
		h.logger.Error("Error decodificando intercambio", zap.Error(err))
		http.Error(w, "Datos inválidos", http.StatusBadRequest)
		return
	}

	// Configurar valores por defecto
	exchange.ID = uuid.New().String()
	exchange.PlayerID = playerID.String()
	exchange.CompletedAt = time.Now()

	err = h.economyService.ExchangeCurrency(playerID, &exchange)
	if err != nil {
		h.logger.Error("Error procesando intercambio", zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	respondWithJSON(w, http.StatusOK, exchange)
}

// GetMarketStatistics obtiene estadísticas del mercado
func (h *EconomyHandler) GetMarketStatistics(w http.ResponseWriter, r *http.Request) {
	stats, err := h.economyService.GetMarketStatistics()
	if err != nil {
		h.logger.Error("Error obteniendo estadísticas", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, stats)
}

// GetMarketTrends obtiene tendencias del mercado
func (h *EconomyHandler) GetMarketTrends(w http.ResponseWriter, r *http.Request) {
	timeframe := r.URL.Query().Get("timeframe")
	if timeframe == "" {
		timeframe = "24h"
	}

	trends, err := h.economyService.GetMarketTrends(timeframe)
	if err != nil {
		h.logger.Error("Error obteniendo tendencias", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, trends)
}

// GetPlayerMarketActivity obtiene la actividad de mercado de un jugador
func (h *EconomyHandler) GetPlayerMarketActivity(w http.ResponseWriter, r *http.Request) {
	// Obtener playerID del contexto
	playerIDStr := r.Context().Value("player_id").(string)
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		h.logger.Error("ID de jugador inválido", zap.Error(err))
		http.Error(w, "ID de jugador inválido", http.StatusBadRequest)
		return
	}

	limit := 20 // Por defecto
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	activities, err := h.economyService.GetPlayerMarketActivity(playerID, limit)
	if err != nil {
		h.logger.Error("Error obteniendo actividad de mercado", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, activities)
}

// GetPlayerResources obtiene los recursos de un jugador
func (h *EconomyHandler) GetPlayerResources(w http.ResponseWriter, r *http.Request) {
	playerIDStr := chi.URLParam(r, "playerID")
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		h.logger.Error("ID de jugador inválido", zap.Error(err))
		http.Error(w, "ID de jugador inválido", http.StatusBadRequest)
		return
	}

	resources, err := h.economyRepo.GetPlayerResources(playerID)
	if err != nil {
		h.logger.Error("Error obteniendo recursos", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, resources)
}

// AddResources agrega recursos a un jugador
func (h *EconomyHandler) AddResources(w http.ResponseWriter, r *http.Request) {
	playerIDStr := chi.URLParam(r, "playerID")
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		h.logger.Error("ID de jugador inválido", zap.Error(err))
		http.Error(w, "ID de jugador inválido", http.StatusBadRequest)
		return
	}

	var request struct {
		ResourceType string `json:"resource_type"`
		Amount       int    `json:"amount"`
		Reason       string `json:"reason"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.logger.Error("Error decodificando request", zap.Error(err))
		http.Error(w, "Datos inválidos", http.StatusBadRequest)
		return
	}

	err = h.economyRepo.AddResources(playerID, request.ResourceType, request.Amount, request.Reason)
	if err != nil {
		h.logger.Error("Error agregando recursos", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Recursos agregados exitosamente"})
}

// RemoveResources remueve recursos de un jugador
func (h *EconomyHandler) RemoveResources(w http.ResponseWriter, r *http.Request) {
	playerIDStr := chi.URLParam(r, "playerID")
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		h.logger.Error("ID de jugador inválido", zap.Error(err))
		http.Error(w, "ID de jugador inválido", http.StatusBadRequest)
		return
	}

	var request struct {
		ResourceType string `json:"resource_type"`
		Amount       int    `json:"amount"`
		Reason       string `json:"reason"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.logger.Error("Error decodificando request", zap.Error(err))
		http.Error(w, "Datos inválidos", http.StatusBadRequest)
		return
	}

	err = h.economyRepo.RemoveResources(playerID, request.ResourceType, request.Amount, request.Reason)
	if err != nil {
		h.logger.Error("Error removiendo recursos", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Recursos removidos exitosamente"})
}

// GetTransactionHistory obtiene el historial de transacciones
func (h *EconomyHandler) GetTransactionHistory(w http.ResponseWriter, r *http.Request) {
	playerIDStr := chi.URLParam(r, "playerID")
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		h.logger.Error("ID de jugador inválido", zap.Error(err))
		http.Error(w, "ID de jugador inválido", http.StatusBadRequest)
		return
	}

	limit := 50 // Por defecto
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	transactions, err := h.economyRepo.GetTransactionHistory(playerID, limit)
	if err != nil {
		h.logger.Error("Error obteniendo historial", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, transactions)
}

// validateEconomyConfig valida la configuración de economía
func (h *EconomyHandler) validateEconomyConfig(config *models.EconomySystemConfig) error {
	if config.PrimaryCurrencyName == "" {
		return fmt.Errorf("nombre de moneda primaria es requerido")
	}
	if config.SecondaryCurrencyName == "" {
		return fmt.Errorf("nombre de moneda secundaria es requerido")
	}
	if config.ExchangeRate <= 0 {
		return fmt.Errorf("tasa de intercambio debe ser mayor a 0")
	}
	if config.ExchangeFee < 0 || config.ExchangeFee > 1 {
		return fmt.Errorf("comisión de intercambio debe estar entre 0 y 1")
	}
	if config.MinExchangeAmount < 0 {
		return fmt.Errorf("cantidad mínima de intercambio debe ser mayor o igual a 0")
	}
	if config.MaxExchangeAmount <= config.MinExchangeAmount {
		return fmt.Errorf("cantidad máxima debe ser mayor a la mínima")
	}
	if config.MarketTax < 0 || config.MarketTax > 1 {
		return fmt.Errorf("impuesto del mercado debe estar entre 0 y 1")
	}
	if config.TransactionFee < 0 || config.TransactionFee > 1 {
		return fmt.Errorf("comisión de transacción debe estar entre 0 y 1")
	}
	return nil
}

// validateMarketListing valida una oferta del mercado
func (h *EconomyHandler) validateMarketListing(listing *models.MarketListing) error {
	if listing.ItemName == "" {
		return fmt.Errorf("nombre del item es requerido")
	}
	if listing.Quantity <= 0 {
		return fmt.Errorf("cantidad debe ser mayor a 0")
	}
	if listing.PricePerUnit <= 0 {
		return fmt.Errorf("precio por unidad debe ser mayor a 0")
	}
	if listing.CurrencyID == "" {
		return fmt.Errorf("moneda es requerida")
	}
	return nil
}

// respondWithJSON es una función helper para responder con JSON
func respondWithJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}
