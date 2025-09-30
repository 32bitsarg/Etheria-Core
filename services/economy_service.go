package services

import (
	"fmt"
	"server-backend/models"
	"server-backend/repository"
	"server-backend/websocket"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type EconomyService struct {
	economyRepo *repository.EconomyRepository
	playerRepo  *repository.PlayerRepository
	villageRepo *repository.VillageRepository
	wsManager   *websocket.Manager
	logger      *zap.Logger
}

func NewEconomyService(
	economyRepo *repository.EconomyRepository,
	playerRepo *repository.PlayerRepository,
	villageRepo *repository.VillageRepository,
	wsManager *websocket.Manager,
	logger *zap.Logger,
) *EconomyService {
	return &EconomyService{
		economyRepo: economyRepo,
		playerRepo:  playerRepo,
		villageRepo: villageRepo,
		wsManager:   wsManager,
		logger:      logger,
	}
}

// GetEconomyDashboard obtiene el dashboard principal de economía
func (s *EconomyService) GetEconomyDashboard(playerID uuid.UUID) (*models.EconomyDashboard, error) {
	// Obtener estadísticas del jugador
	playerStats, err := s.economyRepo.GetPlayerEconomyStatistics(playerID)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo estadísticas: %w", err)
	}

	// Obtener items del mercado
	marketItems, err := s.economyRepo.GetMarketItems("", "", 100) // category="", rarity="", limit=100
	if err != nil {
		return nil, fmt.Errorf("error obteniendo items del mercado: %w", err)
	}

	// Obtener ofertas activas
	activeListings, err := s.economyRepo.GetMarketListings("", 0, 0, "active", 100) // itemType="", itemID=0, sellerID=0, status="active", limit=100
	if err != nil {
		return nil, fmt.Errorf("error obteniendo ofertas activas: %w", err)
	}

	// Obtener transacciones recientes
	recentTransactions, err := s.economyRepo.GetRecentTransactions()
	if err != nil {
		return nil, fmt.Errorf("error obteniendo transacciones recientes: %w", err)
	}

	// Obtener estadísticas del mercado
	marketStats, err := s.economyRepo.GetMarketStatistics(time.Now())
	if err != nil {
		return nil, fmt.Errorf("error obteniendo estadísticas del mercado: %w", err)
	}

	return &models.EconomyDashboard{
		PlayerStats:        *playerStats,
		MarketItems:        marketItems,
		ActiveListings:     activeListings,
		RecentTransactions: recentTransactions,
		MarketStatistics:   *marketStats,
		LastUpdated:        time.Now(),
	}, nil
}

// GetEconomyConfig obtiene la configuración del sistema de economía
func (s *EconomyService) GetEconomyConfig() (*models.EconomySystemConfig, error) {
	config, err := s.economyRepo.GetEconomyConfig()
	if err != nil {
		return nil, fmt.Errorf("error obteniendo configuración: %w", err)
	}

	return config, nil
}

// UpdateEconomyConfig actualiza la configuración del sistema de economía
func (s *EconomyService) UpdateEconomyConfig(config *models.EconomySystemConfig) error {
	// Validar configuración
	if err := s.validateEconomyConfig(config); err != nil {
		return fmt.Errorf("configuración inválida: %w", err)
	}

	// Actualizar configuración
	if err := s.economyRepo.UpdateEconomyConfig(config); err != nil {
		return fmt.Errorf("error actualizando configuración: %w", err)
	}

	s.logger.Info("Configuración de economía actualizada",
		zap.String("updated_by", "admin"), // TODO: Obtener del contexto
		zap.Time("updated_at", time.Now()),
	)

	return nil
}

// GetMarketItems obtiene los items del mercado
func (s *EconomyService) GetMarketItems(filters map[string]interface{}) ([]*models.MarketItem, error) {
	category := ""
	rarity := ""
	limit := 100

	if cat, ok := filters["category"].(string); ok {
		category = cat
	}
	if rar, ok := filters["rarity"].(string); ok {
		rarity = rar
	}
	if lim, ok := filters["limit"].(int); ok {
		limit = lim
	}

	items, err := s.economyRepo.GetMarketItems(category, rarity, limit)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo items del mercado: %w", err)
	}

	// Convertir a punteros
	var itemsPtr []*models.MarketItem
	for i := range items {
		itemsPtr = append(itemsPtr, &items[i])
	}

	return itemsPtr, nil
}

// GetMarketItemDetails obtiene los detalles de un item del mercado
func (s *EconomyService) GetMarketItemDetails(itemID uuid.UUID) (*models.MarketItemDetails, error) {
	item, err := s.economyRepo.GetMarketItem(itemID)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo item: %w", err)
	}

	// Obtener historial de precios (implementación básica)
	priceHistory := []*models.PriceHistory{
		{
			Date:  time.Now().Add(-24 * time.Hour),
			Price: item.BasePrice * 0.95,
		},
		{
			Date:  time.Now(),
			Price: item.BasePrice,
		},
	}

	return &models.MarketItemDetails{
		Item:         item,
		PriceHistory: priceHistory,
	}, nil
}

// GetMarketListings obtiene las ofertas del mercado
func (s *EconomyService) GetMarketListings(filters map[string]interface{}) ([]*models.MarketListing, error) {
	itemType := ""
	itemID := 0
	sellerID := 0
	status := "active"
	limit := 100

	if it, ok := filters["item_type"].(string); ok {
		itemType = it
	}
	if id, ok := filters["item_id"].(int); ok {
		itemID = id
	}
	if sid, ok := filters["seller_id"].(int); ok {
		sellerID = sid
	}
	if st, ok := filters["status"].(string); ok {
		status = st
	}
	if lim, ok := filters["limit"].(int); ok {
		limit = lim
	}

	listings, err := s.economyRepo.GetMarketListings(itemType, itemID, sellerID, status, limit)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo ofertas: %w", err)
	}

	// Convertir a punteros
	var listingsPtr []*models.MarketListing
	for i := range listings {
		listingsPtr = append(listingsPtr, &listings[i])
	}

	return listingsPtr, nil
}

// CreateMarketListing crea una nueva oferta en el mercado
func (s *EconomyService) CreateMarketListing(playerID uuid.UUID, listing *models.MarketListing) error {
	// Validar oferta
	if err := s.validateMarketListing(listing); err != nil {
		return fmt.Errorf("oferta inválida: %w", err)
	}

	// Verificar recursos del jugador
	if err := s.verifyPlayerResources(playerID, listing); err != nil {
		return fmt.Errorf("recursos insuficientes: %w", err)
	}

	// Consumir recursos del jugador
	if err := s.consumePlayerResources(playerID, listing); err != nil {
		return fmt.Errorf("error consumiendo recursos: %w", err)
	}

	// Crear oferta
	if err := s.economyRepo.CreateMarketListing(listing); err != nil {
		return fmt.Errorf("error creando oferta: %w", err)
	}

	// Enviar notificación
	if err := s.sendMarketNotification(playerID.String(), "listing_created", map[string]interface{}{
		"listing_id": listing.ID,
		"item_name":  listing.ItemName,
		"quantity":   listing.Quantity,
		"price":      listing.PricePerUnit,
	}); err != nil {
		s.logger.Warn("Error enviando notificación de oferta creada", zap.Error(err))
	}

	return nil
}

// GetMarketStatistics obtiene estadísticas del mercado
func (s *EconomyService) GetMarketStatistics() (*models.MarketStatistics, error) {
	stats, err := s.economyRepo.GetMarketStatistics(time.Now())
	if err != nil {
		return nil, fmt.Errorf("error obteniendo estadísticas: %w", err)
	}

	return stats, nil
}

// GetMarketTrends obtiene las tendencias del mercado
func (s *EconomyService) GetMarketTrends(timeframe string) ([]*models.MarketTrend, error) {
	// Implementación básica de tendencias
	trends := []*models.MarketTrend{
		{
			ID:             uuid.New(),
			ResourceID:     uuid.New(),
			TrendType:      "rising",
			Direction:      "up",
			Strength:       0.7,
			Confidence:     0.8,
			Duration:       24,
			StartPrice:     100.0,
			CurrentPrice:   120.0,
			PredictedPrice: 130.0,
			Factors:        `{"demand": "high", "supply": "low"}`,
			IsActive:       true,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
	}

	return trends, nil
}

// ExchangeCurrency intercambia monedas
func (s *EconomyService) ExchangeCurrency(playerID uuid.UUID, exchange *models.CurrencyExchange) error {
	// Validar intercambio
	if err := s.validateCurrencyExchange(exchange); err != nil {
		return fmt.Errorf("intercambio inválido: %w", err)
	}

	// Verificar monedas del jugador
	if err := s.verifyPlayerCurrency(playerID, exchange); err != nil {
		return fmt.Errorf("monedas insuficientes: %w", err)
	}

	// Procesar intercambio
	if err := s.economyRepo.ProcessCurrencyExchange(exchange); err != nil {
		return fmt.Errorf("error procesando intercambio: %w", err)
	}

	// Enviar notificación
	if err := s.sendMarketNotification(playerID.String(), "currency_exchanged", map[string]interface{}{
		"from_currency": exchange.FromCurrencyID,
		"to_currency":   exchange.ToCurrencyID,
		"from_amount":   exchange.FromAmount,
		"to_amount":     exchange.ToAmount,
		"fee":           exchange.Fee,
	}); err != nil {
		s.logger.Warn("Error enviando notificación de intercambio", zap.Error(err))
	}

	return nil
}

// GetPlayerEconomy obtiene la economía de un jugador
func (s *EconomyService) GetPlayerEconomy(playerID uuid.UUID) (*models.PlayerEconomy, error) {
	economy, err := s.economyRepo.GetPlayerEconomy(playerID)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo economía: %w", err)
	}

	return economy, nil
}

// GetPlayerMarketActivity obtiene la actividad de mercado de un jugador
func (s *EconomyService) GetPlayerMarketActivity(playerID uuid.UUID, limit int) ([]*models.MarketActivity, error) {
	// Implementación básica de actividad de mercado
	activities := []*models.MarketActivity{
		{
			ID:            uuid.New(),
			ResourceID:    uuid.New(),
			ActivityType:  "trade",
			Volume:        1000,
			Price:         150.0,
			Change:        5.2,
			Buyers:        5,
			Sellers:       3,
			Data:          `{"transaction_type": "buy"}`,
			IsSignificant: true,
			CreatedAt:     time.Now(),
		},
	}

	return activities, nil
}

// SetWebSocketManager establece el manager de WebSocket
func (s *EconomyService) SetWebSocketManager(wsManager *websocket.Manager) {
	s.wsManager = wsManager
}

// validateEconomyConfig valida la configuración de economía
func (s *EconomyService) validateEconomyConfig(config *models.EconomySystemConfig) error {
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
	return nil
}

// validateMarketListing valida una oferta del mercado
func (s *EconomyService) validateMarketListing(listing *models.MarketListing) error {
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

// validateCurrencyExchange valida un intercambio de monedas
func (s *EconomyService) validateCurrencyExchange(exchange *models.CurrencyExchange) error {
	if exchange.FromCurrencyID == "" || exchange.ToCurrencyID == "" {
		return fmt.Errorf("monedas de origen y destino son requeridas")
	}
	if exchange.FromCurrencyID == exchange.ToCurrencyID {
		return fmt.Errorf("no se puede intercambiar la misma moneda")
	}
	if exchange.FromAmount <= 0 {
		return fmt.Errorf("cantidad de origen debe ser mayor a 0")
	}
	if exchange.ToAmount <= 0 {
		return fmt.Errorf("cantidad de destino debe ser mayor a 0")
	}
	return nil
}

// verifyPlayerResources verifica que el jugador tenga los recursos necesarios
func (s *EconomyService) verifyPlayerResources(playerID uuid.UUID, listing *models.MarketListing) error {
	resources, err := s.economyRepo.GetPlayerResources(playerID)
	if err != nil {
		return fmt.Errorf("error obteniendo recursos: %w", err)
	}

	// Verificar según el tipo de recurso
	switch listing.ItemName {
	case "wood":
		if resources.Wood < listing.Quantity {
			return fmt.Errorf("madera insuficiente")
		}
	case "stone":
		if resources.Stone < listing.Quantity {
			return fmt.Errorf("piedra insuficiente")
		}
	case "iron":
		if resources.Iron < listing.Quantity {
			return fmt.Errorf("hierro insuficiente")
		}
	case "food":
		if resources.Food < listing.Quantity {
			return fmt.Errorf("comida insuficiente")
		}
	default:
		return fmt.Errorf("tipo de recurso no soportado: %s", listing.ItemName)
	}

	return nil
}

// verifyPlayerCurrency verifica que el jugador tenga las monedas necesarias
func (s *EconomyService) verifyPlayerCurrency(playerID uuid.UUID, exchange *models.CurrencyExchange) error {
	economy, err := s.economyRepo.GetPlayerEconomy(playerID)
	if err != nil {
		return fmt.Errorf("error obteniendo economía: %w", err)
	}

	if exchange.FromCurrencyID == "primary" && economy.PrimaryCurrency < int(exchange.FromAmount) {
		return fmt.Errorf("moneda primaria insuficiente")
	}
	if exchange.FromCurrencyID == "secondary" && economy.SecondaryCurrency < int(exchange.FromAmount) {
		return fmt.Errorf("moneda secundaria insuficiente")
	}

	return nil
}

// consumePlayerResources consume los recursos del jugador
func (s *EconomyService) consumePlayerResources(playerID uuid.UUID, listing *models.MarketListing) error {
	return s.economyRepo.RemoveResources(playerID, listing.ItemName, listing.Quantity, "market_listing")
}

// calculateExchangeRate calcula la tasa de intercambio
func (s *EconomyService) calculateExchangeRate(fromCurrency, toCurrency string) (float64, error) {
	config, err := s.economyRepo.GetEconomySystemConfig()
	if err != nil {
		return 0, fmt.Errorf("error obteniendo configuración: %w", err)
	}

	if fromCurrency == "primary" && toCurrency == "secondary" {
		return config.ExchangeRate, nil
	}
	if fromCurrency == "secondary" && toCurrency == "primary" {
		return 1.0 / config.ExchangeRate, nil
	}

	return 0, fmt.Errorf("combinación de monedas no soportada")
}

// sendMarketNotification envía una notificación de mercado
func (s *EconomyService) sendMarketNotification(playerID string, notificationType string, data map[string]interface{}) error {
	if s.wsManager == nil {
		s.logger.Warn("WebSocket Manager no disponible para notificaciones de mercado")
		return nil
	}

	// Crear mensaje de notificación
	message := map[string]interface{}{
		"type": "market_notification",
		"data": map[string]interface{}{
			"notification_type": notificationType,
			"timestamp":        time.Now().Unix(),
			"data":            data,
		},
	}

	// Enviar notificación por WebSocket
	if err := s.wsManager.SendToUser(playerID, "market_notification", message); err != nil {
		s.logger.Warn("Error enviando notificación de mercado por WebSocket",
			zap.String("player_id", playerID),
			zap.String("type", notificationType),
			zap.Error(err),
		)
		return err
	}

	s.logger.Info("Notificación de mercado enviada por WebSocket",
		zap.String("player_id", playerID),
		zap.String("type", notificationType),
		zap.Any("data", data),
	)

	return nil
}
