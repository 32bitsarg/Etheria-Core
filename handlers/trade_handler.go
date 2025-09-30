package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"server-backend/models"
	"server-backend/repository"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type TradeHandler struct {
	tradeRepo   *repository.TradeRepository
	villageRepo *repository.VillageRepository
	logger      *zap.Logger
}

func NewTradeHandler(tradeRepo *repository.TradeRepository, villageRepo *repository.VillageRepository, logger *zap.Logger) *TradeHandler {
	return &TradeHandler{
		tradeRepo:   tradeRepo,
		villageRepo: villageRepo,
		logger:      logger,
	}
}

// CreateTradeOffer crea una nueva oferta de comercio
func (h *TradeHandler) CreateTradeOffer(w http.ResponseWriter, r *http.Request) {
	// Obtener playerID del contexto
	playerIDStr := r.Context().Value("player_id").(string)
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		h.logger.Error("ID de jugador inválido", zap.Error(err))
		http.Error(w, "ID de jugador inválido", http.StatusBadRequest)
		return
	}

	var offer models.TradeOffer
	if err := json.NewDecoder(r.Body).Decode(&offer); err != nil {
		h.logger.Error("Error decodificando request", zap.Error(err))
		http.Error(w, "Error decodificando request", http.StatusBadRequest)
		return
	}

	// Configurar valores por defecto
	offer.SellerID = playerID
	offer.Status = "active"

	// Verificar que la oferta tenga datos válidos
	if offer.VillageID == uuid.Nil || offer.ResourceType == "" || offer.Amount <= 0 || offer.PricePerUnit <= 0 {
		http.Error(w, "Datos de oferta inválidos", http.StatusBadRequest)
		return
	}

	createdOffer, err := h.tradeRepo.CreateTradeOffer(&offer)
	if err != nil {
		h.logger.Error("Error creando oferta", zap.Error(err))
		http.Error(w, "Error creando oferta", http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusCreated, createdOffer)
}

// GetTradeOffers obtiene todas las ofertas de comercio
func (h *TradeHandler) GetTradeOffers(w http.ResponseWriter, r *http.Request) {
	// Parámetros de filtrado
	resourceType := r.URL.Query().Get("resource_type")
	sellerID := r.URL.Query().Get("seller_id")
	priceMin := r.URL.Query().Get("price_min")
	priceMax := r.URL.Query().Get("price_max")

	offers, err := h.tradeRepo.GetTradeOffers(resourceType, sellerID, priceMin, priceMax)
	if err != nil {
		h.logger.Error("Error obteniendo ofertas", zap.Error(err))
		http.Error(w, "Error obteniendo ofertas", http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, offers)
}

// GetTradeOffer obtiene una oferta específica
func (h *TradeHandler) GetTradeOffer(w http.ResponseWriter, r *http.Request) {
	offerIDStr := chi.URLParam(r, "offerID")
	offerID, err := uuid.Parse(offerIDStr)
	if err != nil {
		http.Error(w, "ID de oferta inválido", http.StatusBadRequest)
		return
	}

	offer, err := h.tradeRepo.GetTradeOffer(offerID)
	if err != nil {
		h.logger.Error("Error obteniendo oferta", zap.Error(err))
		http.Error(w, "Oferta no encontrada", http.StatusNotFound)
		return
	}

	respondWithJSON(w, http.StatusOK, offer)
}

// BuyTradeOffer compra una oferta de comercio
func (h *TradeHandler) BuyTradeOffer(w http.ResponseWriter, r *http.Request) {
	offerIDStr := chi.URLParam(r, "offerID")
	offerID, err := uuid.Parse(offerIDStr)
	if err != nil {
		http.Error(w, "ID de oferta inválido", http.StatusBadRequest)
		return
	}

	playerIDStr := r.Context().Value("player_id").(string)
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		h.logger.Error("ID de jugador inválido", zap.Error(err))
		http.Error(w, "ID de jugador inválido", http.StatusBadRequest)
		return
	}

	var request struct {
		VillageID uuid.UUID `json:"village_id"`
		Amount    int       `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Error decodificando request", http.StatusBadRequest)
		return
	}

	// Verificar que el jugador no esté comprando su propia oferta
	offer, err := h.tradeRepo.GetTradeOffer(offerID)
	if err != nil {
		h.logger.Error("Error obteniendo oferta", zap.Error(err))
		http.Error(w, "Oferta no encontrada", http.StatusNotFound)
		return
	}

	if offer.SellerID == playerID {
		http.Error(w, "No puedes comprar tu propia oferta", http.StatusBadRequest)
		return
	}

	// Procesar la transacción
	transaction, err := h.tradeRepo.ProcessTrade(offerID, playerID, request.VillageID, request.Amount)
	if err != nil {
		h.logger.Error("Error procesando transacción", zap.Error(err))
		http.Error(w, "Error procesando transacción", http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, transaction)
}

// CancelTradeOffer cancela una oferta de comercio
func (h *TradeHandler) CancelTradeOffer(w http.ResponseWriter, r *http.Request) {
	offerIDStr := chi.URLParam(r, "offerID")
	offerID, err := uuid.Parse(offerIDStr)
	if err != nil {
		http.Error(w, "ID de oferta inválido", http.StatusBadRequest)
		return
	}

	playerIDStr := r.Context().Value("player_id").(string)
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		h.logger.Error("ID de jugador inválido", zap.Error(err))
		http.Error(w, "ID de jugador inválido", http.StatusBadRequest)
		return
	}

	// Verificar que el jugador sea el vendedor
	offer, err := h.tradeRepo.GetTradeOffer(offerID)
	if err != nil {
		h.logger.Error("Error obteniendo oferta", zap.Error(err))
		http.Error(w, "Oferta no encontrada", http.StatusNotFound)
		return
	}

	if offer.SellerID != playerID {
		http.Error(w, "No autorizado", http.StatusForbidden)
		return
	}

	if err := h.tradeRepo.CancelTradeOffer(offerID); err != nil {
		h.logger.Error("Error cancelando oferta", zap.Error(err))
		http.Error(w, "Error cancelando oferta", http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Oferta cancelada exitosamente"})
}

// GetPlayerTradeOffers obtiene las ofertas del jugador
func (h *TradeHandler) GetPlayerTradeOffers(w http.ResponseWriter, r *http.Request) {
	playerIDStr := r.Context().Value("player_id").(string)
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		h.logger.Error("ID de jugador inválido", zap.Error(err))
		http.Error(w, "ID de jugador inválido", http.StatusBadRequest)
		return
	}

	offers, err := h.tradeRepo.GetPlayerTradeOffers(playerID)
	if err != nil {
		h.logger.Error("Error obteniendo ofertas del jugador", zap.Error(err))
		http.Error(w, "Error obteniendo ofertas", http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, offers)
}

// GetTradeHistory obtiene el historial de transacciones
func (h *TradeHandler) GetTradeHistory(w http.ResponseWriter, r *http.Request) {
	playerIDStr := r.Context().Value("player_id").(string)
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		h.logger.Error("ID de jugador inválido", zap.Error(err))
		http.Error(w, "ID de jugador inválido", http.StatusBadRequest)
		return
	}

	// Parámetros de paginación
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	transactions, err := h.tradeRepo.GetTradeHistory(playerID, limit)
	if err != nil {
		h.logger.Error("Error obteniendo historial", zap.Error(err))
		http.Error(w, "Error obteniendo historial", http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, transactions)
}

// GetMarketStats obtiene estadísticas del mercado
func (h *TradeHandler) GetMarketStats(w http.ResponseWriter, r *http.Request) {
	// Implementación básica de estadísticas del mercado
	stats := []models.MarketStats{
		{
			ResourceType: "wood",
			AveragePrice: 150.0,
			MinPrice:     100,
			MaxPrice:     200,
			TotalVolume:  10000,
			ActiveOffers: 25,
			LastUpdated:  time.Now(),
		},
		{
			ResourceType: "stone",
			AveragePrice: 200.0,
			MinPrice:     150,
			MaxPrice:     250,
			TotalVolume:  8000,
			ActiveOffers: 20,
			LastUpdated:  time.Now(),
		},
	}

	respondWithJSON(w, http.StatusOK, stats)
}

// GetResourcePrices obtiene precios de recursos
func (h *TradeHandler) GetResourcePrices(w http.ResponseWriter, r *http.Request) {
	// Implementación básica de precios de recursos
	prices := []models.ResourcePrice{
		{
			ResourceType: "wood",
			CurrentPrice: 150,
			Change24h:    5.2,
			Volume24h:    10000,
			UpdatedAt:    time.Now(),
		},
		{
			ResourceType: "stone",
			CurrentPrice: 200,
			Change24h:    -2.1,
			Volume24h:    8000,
			UpdatedAt:    time.Now(),
		},
	}

	respondWithJSON(w, http.StatusOK, prices)
}

// CreateDirectTrade crea un intercambio directo
func (h *TradeHandler) CreateDirectTrade(w http.ResponseWriter, r *http.Request) {
	playerIDStr := r.Context().Value("player_id").(string)
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		h.logger.Error("ID de jugador inválido", zap.Error(err))
		http.Error(w, "ID de jugador inválido", http.StatusBadRequest)
		return
	}

	var trade models.DirectTrade
	if err := json.NewDecoder(r.Body).Decode(&trade); err != nil {
		h.logger.Error("Error decodificando request", zap.Error(err))
		http.Error(w, "Error decodificando request", http.StatusBadRequest)
		return
	}

	// Configurar valores por defecto
	trade.ID = uuid.New()
	trade.InitiatorID = playerID
	trade.Status = "pending"
	trade.CreatedAt = time.Now()
	trade.ExpiresAt = time.Now().Add(24 * time.Hour)

	// TODO: Implementar creación de intercambio directo en el repositorio
	respondWithJSON(w, http.StatusCreated, trade)
}

// AcceptDirectTrade acepta un intercambio directo
func (h *TradeHandler) AcceptDirectTrade(w http.ResponseWriter, r *http.Request) {
	tradeIDStr := chi.URLParam(r, "tradeID")
	tradeID, err := uuid.Parse(tradeIDStr)
	if err != nil {
		http.Error(w, "ID de intercambio inválido", http.StatusBadRequest)
		return
	}

	playerIDStr := r.Context().Value("player_id").(string)
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		h.logger.Error("ID de jugador inválido", zap.Error(err))
		http.Error(w, "ID de jugador inválido", http.StatusBadRequest)
		return
	}

	// TODO: Implementar aceptación de intercambio directo usando tradeID y playerID
	_ = tradeID
	_ = playerID
	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Intercambio aceptado"})
}

// DeclineDirectTrade rechaza un intercambio directo
func (h *TradeHandler) DeclineDirectTrade(w http.ResponseWriter, r *http.Request) {
	tradeIDStr := chi.URLParam(r, "tradeID")
	tradeID, err := uuid.Parse(tradeIDStr)
	if err != nil {
		http.Error(w, "ID de intercambio inválido", http.StatusBadRequest)
		return
	}

	playerIDStr := r.Context().Value("player_id").(string)
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		h.logger.Error("ID de jugador inválido", zap.Error(err))
		http.Error(w, "ID de jugador inválido", http.StatusBadRequest)
		return
	}

	// TODO: Implementar rechazo de intercambio directo usando tradeID y playerID
	_ = tradeID
	_ = playerID
	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Intercambio rechazado"})
}

// GetDirectTrades obtiene los intercambios directos
func (h *TradeHandler) GetDirectTrades(w http.ResponseWriter, r *http.Request) {
	playerIDStr := r.Context().Value("player_id").(string)
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		h.logger.Error("ID de jugador inválido", zap.Error(err))
		http.Error(w, "ID de jugador inválido", http.StatusBadRequest)
		return
	}

	// TODO: Implementar obtención de intercambios directos usando playerID
	_ = playerID
	trades := []models.DirectTrade{}
	respondWithJSON(w, http.StatusOK, trades)
}
