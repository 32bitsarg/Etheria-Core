package repository

import (
	"database/sql"
	"fmt"
	"time"

	"server-backend/models"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type EconomyRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewEconomyRepository(db *sql.DB, logger *zap.Logger) *EconomyRepository {
	return &EconomyRepository{
		db:     db,
		logger: logger,
	}
}

// GetEconomySystemConfig obtiene la configuración del sistema de economía
func (r *EconomyRepository) GetEconomySystemConfig() (*models.EconomySystemConfig, error) {
	query := `
		SELECT id, is_enabled, primary_currency_name, secondary_currency_name,
		       primary_currency_symbol, secondary_currency_symbol, exchange_rate,
		       exchange_fee, min_exchange_amount, max_exchange_amount, market_tax,
		       transaction_fee, max_price_fluctuation, price_update_interval,
		       max_items_per_player, max_active_listings, min_listing_duration,
		       max_listing_duration, advanced_config, created_at, updated_at
		FROM economy_system_config
		ORDER BY id DESC
		LIMIT 1
	`

	var config models.EconomySystemConfig
	err := r.db.QueryRow(query).Scan(
		&config.ID, &config.IsEnabled, &config.PrimaryCurrencyName, &config.SecondaryCurrencyName,
		&config.PrimaryCurrencySymbol, &config.SecondaryCurrencySymbol, &config.ExchangeRate,
		&config.ExchangeFee, &config.MinExchangeAmount, &config.MaxExchangeAmount, &config.MarketTax,
		&config.TransactionFee, &config.MaxPriceFluctuation, &config.PriceUpdateInterval,
		&config.MaxItemsPerPlayer, &config.MaxActiveListings, &config.MinListingDuration,
		&config.MaxListingDuration, &config.AdvancedConfig, &config.CreatedAt, &config.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			// Crear configuración por defecto
			return r.createDefaultEconomyConfig()
		}
		return nil, fmt.Errorf("error obteniendo configuración: %w", err)
	}

	return &config, nil
}

// createDefaultEconomyConfig crea una configuración por defecto
func (r *EconomyRepository) createDefaultEconomyConfig() (*models.EconomySystemConfig, error) {
	query := `
		INSERT INTO economy_system_config (
			is_enabled, primary_currency_name, secondary_currency_name,
			primary_currency_symbol, secondary_currency_symbol, exchange_rate,
			exchange_fee, min_exchange_amount, max_exchange_amount, market_tax,
			transaction_fee, max_price_fluctuation, price_update_interval,
			max_items_per_player, max_active_listings, min_listing_duration,
			max_listing_duration, advanced_config, created_at, updated_at
		) VALUES (
			true, 'Silver', 'Gold', 'S', 'G', 100.0, 0.05, 100, 1000000,
			0.02, 0.01, 0.50, 15, 100, 10, 1, 168,
			'{"enable_auctions": true, "enable_alerts": true, "enable_trends": true}',
			$1, $1
		) RETURNING id, is_enabled, primary_currency_name, secondary_currency_name,
		            primary_currency_symbol, secondary_currency_symbol, exchange_rate,
		            exchange_fee, min_exchange_amount, max_exchange_amount, market_tax,
		            transaction_fee, max_price_fluctuation, price_update_interval,
		            max_items_per_player, max_active_listings, min_listing_duration,
		            max_listing_duration, advanced_config, created_at, updated_at
	`

	var config models.EconomySystemConfig
	now := time.Now()
	err := r.db.QueryRow(query, now).Scan(
		&config.ID, &config.IsEnabled, &config.PrimaryCurrencyName, &config.SecondaryCurrencyName,
		&config.PrimaryCurrencySymbol, &config.SecondaryCurrencySymbol, &config.ExchangeRate,
		&config.ExchangeFee, &config.MinExchangeAmount, &config.MaxExchangeAmount, &config.MarketTax,
		&config.TransactionFee, &config.MaxPriceFluctuation, &config.PriceUpdateInterval,
		&config.MaxItemsPerPlayer, &config.MaxActiveListings, &config.MinListingDuration,
		&config.MaxListingDuration, &config.AdvancedConfig, &config.CreatedAt, &config.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("error creando configuración por defecto: %w", err)
	}

	return &config, nil
}

// UpdateEconomySystemConfig actualiza la configuración del sistema de economía
func (r *EconomyRepository) UpdateEconomySystemConfig(config *models.EconomySystemConfig) error {
	query := `
		UPDATE economy_system_config 
		SET is_enabled = $1, primary_currency_name = $2, secondary_currency_name = $3,
		    primary_currency_symbol = $4, secondary_currency_symbol = $5, exchange_rate = $6,
		    exchange_fee = $7, min_exchange_amount = $8, max_exchange_amount = $9, market_tax = $10,
		    transaction_fee = $11, max_price_fluctuation = $12, price_update_interval = $13,
		    max_items_per_player = $14, max_active_listings = $15, min_listing_duration = $16,
		    max_listing_duration = $17, advanced_config = $18, updated_at = $19
		WHERE id = $20
	`

	_, err := r.db.Exec(query,
		config.IsEnabled, config.PrimaryCurrencyName, config.SecondaryCurrencyName,
		config.PrimaryCurrencySymbol, config.SecondaryCurrencySymbol, config.ExchangeRate,
		config.ExchangeFee, config.MinExchangeAmount, config.MaxExchangeAmount, config.MarketTax,
		config.TransactionFee, config.MaxPriceFluctuation, config.PriceUpdateInterval,
		config.MaxItemsPerPlayer, config.MaxActiveListings, config.MinListingDuration,
		config.MaxListingDuration, config.AdvancedConfig, time.Now(), config.ID,
	)

	if err != nil {
		return fmt.Errorf("error actualizando configuración: %w", err)
	}

	return nil
}

// GetPlayerEconomy obtiene la economía de un jugador
func (r *EconomyRepository) GetPlayerEconomy(playerID uuid.UUID) (*models.PlayerEconomy, error) {
	query := `
		SELECT id, player_id, primary_currency, secondary_currency, total_spent,
		       total_earned, items_sold, items_bought, reputation, trust_level,
		       daily_spending_limit, daily_earning_limit, current_daily_spent,
		       current_daily_earned, last_transaction, created_at, updated_at
		FROM player_economy
		WHERE player_id = $1
	`

	var economy models.PlayerEconomy
	err := r.db.QueryRow(query, playerID).Scan(
		&economy.ID, &economy.PlayerID, &economy.PrimaryCurrency, &economy.SecondaryCurrency,
		&economy.TotalSpent, &economy.TotalEarned, &economy.ItemsSold, &economy.ItemsBought,
		&economy.Reputation, &economy.TrustLevel, &economy.DailySpendingLimit,
		&economy.DailyEarningLimit, &economy.CurrentDailySpent, &economy.CurrentDailyEarned,
		&economy.LastTransaction, &economy.CreatedAt, &economy.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			// Crear economía por defecto
			return r.createDefaultPlayerEconomy(playerID)
		}
		return nil, fmt.Errorf("error obteniendo economía del jugador: %w", err)
	}

	return &economy, nil
}

// createDefaultPlayerEconomy crea una economía por defecto para un jugador
func (r *EconomyRepository) createDefaultPlayerEconomy(playerID uuid.UUID) (*models.PlayerEconomy, error) {
	query := `
		INSERT INTO player_economy (
			player_id, primary_currency, secondary_currency, total_spent,
			total_earned, items_sold, items_bought, reputation, trust_level,
			daily_spending_limit, daily_earning_limit, current_daily_spent,
			current_daily_earned, created_at, updated_at
		) VALUES (
			$1, 1000, 0, 0, 0, 0, 0, 0, 'new',
			10000, 50000, 0, 0, $2, $3
		) RETURNING id, player_id, primary_currency, secondary_currency, total_spent,
		            total_earned, items_sold, items_bought, reputation, trust_level,
		            daily_spending_limit, daily_earning_limit, current_daily_spent,
		            current_daily_earned, last_transaction, created_at, updated_at
	`

	var economy models.PlayerEconomy
	now := time.Now()
	err := r.db.QueryRow(query, playerID, now, now).Scan(
		&economy.ID, &economy.PlayerID, &economy.PrimaryCurrency, &economy.SecondaryCurrency,
		&economy.TotalSpent, &economy.TotalEarned, &economy.ItemsSold, &economy.ItemsBought,
		&economy.Reputation, &economy.TrustLevel, &economy.DailySpendingLimit,
		&economy.DailyEarningLimit, &economy.CurrentDailySpent, &economy.CurrentDailyEarned,
		&economy.LastTransaction, &economy.CreatedAt, &economy.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("error creando economía por defecto: %w", err)
	}

	return &economy, nil
}

// UpdatePlayerCurrency actualiza las monedas de un jugador
func (r *EconomyRepository) UpdatePlayerCurrency(playerID uuid.UUID, primaryDelta, secondaryDelta int) error {
	query := `
		UPDATE player_economy 
		SET primary_currency = primary_currency + $1,
		    secondary_currency = secondary_currency + $2,
		    updated_at = $3
		WHERE player_id = $4
	`

	_, err := r.db.Exec(query, primaryDelta, secondaryDelta, time.Now(), playerID)
	if err != nil {
		return fmt.Errorf("error actualizando monedas: %w", err)
	}

	return nil
}

// GetMarketItems obtiene los items del mercado
func (r *EconomyRepository) GetMarketItems(category string, rarity string, limit int) ([]models.MarketItem, error) {
	// Mock temporal
	return []models.MarketItem{}, nil
}

// GetMarketListings obtiene las listas de venta del mercado
func (r *EconomyRepository) GetMarketListings(itemType string, itemID int, sellerID int, status string, limit int) ([]models.MarketListing, error) {
	// Mock temporal
	return []models.MarketListing{}, nil
}

// CreateMarketListing crea una nueva lista de venta
func (r *EconomyRepository) CreateMarketListing(listing *models.MarketListing) error {
	query := `
		INSERT INTO market_listings (
			seller_id, item_name, quantity, price_per_unit, currency_id,
			created_at, expires_at, is_active
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8
		) RETURNING id
	`

	now := time.Now()
	expiresAt := now.Add(time.Hour * 24 * 7) // 7 días por defecto

	err := r.db.QueryRow(query,
		listing.SellerID, listing.ItemName, listing.Quantity,
		listing.PricePerUnit, listing.CurrencyID, now, expiresAt, true,
	).Scan(&listing.ID)

	if err != nil {
		return fmt.Errorf("error creando lista de venta: %w", err)
	}

	// Actualizar fechas
	listing.CreatedAt = now
	listing.ExpiresAt = expiresAt
	listing.IsActive = true

	return nil
}

// ExchangeCurrency intercambia monedas
func (r *EconomyRepository) ExchangeCurrency(playerID uuid.UUID, exchangeType string, amount int) (*models.CurrencyExchange, error) {
	// Obtener configuración del sistema
	config, err := r.GetEconomySystemConfig()
	if err != nil {
		return nil, fmt.Errorf("error obteniendo configuración: %w", err)
	}

	// Validar cantidad mínima y máxima
	if amount < config.MinExchangeAmount {
		return nil, fmt.Errorf("cantidad menor al mínimo permitido: %d", config.MinExchangeAmount)
	}
	if amount > config.MaxExchangeAmount {
		return nil, fmt.Errorf("cantidad mayor al máximo permitido: %d", config.MaxExchangeAmount)
	}

	// Calcular tasas y montos
	var fromCurrencyID, toCurrencyID string
	var fromAmount, toAmount float64
	var exchangeRate float64

	if exchangeType == "primary_to_secondary" {
		fromCurrencyID = "primary"
		toCurrencyID = "secondary"
		exchangeRate = config.ExchangeRate
		fromAmount = float64(amount)
		toAmount = fromAmount * exchangeRate * (1 - config.ExchangeFee)
	} else if exchangeType == "secondary_to_primary" {
		fromCurrencyID = "secondary"
		toCurrencyID = "primary"
		exchangeRate = 1.0 / config.ExchangeRate
		fromAmount = float64(amount)
		toAmount = fromAmount * exchangeRate * (1 - config.ExchangeFee)
	} else {
		return nil, fmt.Errorf("tipo de intercambio inválido: %s", exchangeType)
	}

	// Crear registro de intercambio
	exchange := &models.CurrencyExchange{
		ID:             uuid.New().String(),
		PlayerID:       playerID.String(),
		FromCurrencyID: fromCurrencyID,
		ToCurrencyID:   toCurrencyID,
		FromAmount:     fromAmount,
		ToAmount:       toAmount,
		ExchangeRate:   exchangeRate,
		Fee:            fromAmount * config.ExchangeFee,
		CompletedAt:    time.Now(),
	}

	// Actualizar monedas del jugador
	var primaryDelta, secondaryDelta int
	if exchangeType == "primary_to_secondary" {
		primaryDelta = -amount
		secondaryDelta = int(toAmount)
	} else {
		primaryDelta = int(toAmount)
		secondaryDelta = -amount
	}

	err = r.UpdatePlayerCurrency(playerID, primaryDelta, secondaryDelta)
	if err != nil {
		return nil, fmt.Errorf("error actualizando monedas: %w", err)
	}

	// Obtener economía del jugador
	playerEconomy, err := r.GetPlayerEconomy(playerID)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo economía del jugador: %w", err)
	}

	// Actualizar límites diarios
	if exchangeType == "primary_to_secondary" {
		playerEconomy.CurrentDailySpent += amount
	} else {
		playerEconomy.CurrentDailyEarned += int(toAmount)
	}

	// Guardar intercambio en base de datos
	query := `
		INSERT INTO currency_exchanges (
			id, player_id, from_currency_id, to_currency_id, from_amount,
			to_amount, exchange_rate, fee, completed_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err = r.db.Exec(query,
		exchange.ID, exchange.PlayerID, exchange.FromCurrencyID, exchange.ToCurrencyID,
		exchange.FromAmount, exchange.ToAmount, exchange.ExchangeRate, exchange.Fee, exchange.CompletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("error guardando intercambio: %w", err)
	}

	return exchange, nil
}

// GetMarketStatistics obtiene estadísticas del mercado
func (r *EconomyRepository) GetMarketStatistics(date time.Time) (*models.MarketStatistics, error) {
	// Mock temporal
	return &models.MarketStatistics{}, nil
}

// GetEconomyDashboard obtiene el dashboard completo de economía
func (r *EconomyRepository) GetEconomyDashboard(playerID uuid.UUID) (*models.EconomyDashboard, error) {
	// Obtener configuración del sistema
	config, err := r.GetEconomySystemConfig()
	if err != nil {
		return nil, fmt.Errorf("error obteniendo configuración: %w", err)
	}

	// Obtener economía del jugador
	playerEconomy, err := r.GetPlayerEconomy(playerID)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo economía del jugador: %w", err)
	}

	// Obtener listados activos del jugador
	listings, err := r.GetMarketListings("", 0, 0, "active", 10)
	if err != nil {
		r.logger.Warn("Error obteniendo listados del jugador", zap.Error(err))
		listings = []models.MarketListing{}
	}

	// Obtener estadísticas del mercado
	marketStats, err := r.GetMarketStatistics(time.Now())
	if err != nil {
		r.logger.Warn("Error obteniendo estadísticas del mercado", zap.Error(err))
		marketStats = &models.MarketStatistics{}
	}

	dashboard := &models.EconomyDashboard{
		SystemConfig:   *config,
		PlayerEconomy:  *playerEconomy,
		ActiveListings: listings,
		MarketStats:    *marketStats,
		LastUpdated:    time.Now(),
	}

	return dashboard, nil
}

// GetEconomyConfig obtiene la configuración del sistema de economía
func (r *EconomyRepository) GetEconomyConfig() (*models.EconomySystemConfig, error) {
	return r.GetEconomySystemConfig()
}

// UpdateEconomyConfig actualiza la configuración del sistema de economía
func (r *EconomyRepository) UpdateEconomyConfig(config *models.EconomySystemConfig) error {
	return r.UpdateEconomySystemConfig(config)
}

// GetPlayerMarketActivity obtiene la actividad de mercado de un jugador
func (r *EconomyRepository) GetPlayerMarketActivity(playerID uuid.UUID) (*models.PlayerMarketActivity, error) {
	query := `
		SELECT 
			COUNT(CASE WHEN status = 'active' THEN 1 END) as active_listings,
			COUNT(CASE WHEN status = 'completed' THEN 1 END) as completed_sales,
			COALESCE(SUM(CASE WHEN status = 'completed' THEN quantity * price_per_unit ELSE 0 END), 0) as total_sales_value,
			COUNT(CASE WHEN buyer_id = $1 THEN 1 END) as completed_purchases,
			COALESCE(SUM(CASE WHEN buyer_id = $1 THEN quantity * price_per_unit ELSE 0 END), 0) as total_purchase_value
		FROM market_listings
		WHERE seller_id = $1 OR buyer_id = $1
	`

	var activity models.PlayerMarketActivity
	activity.PlayerID = playerID.String()

	err := r.db.QueryRow(query, playerID).Scan(
		&activity.ActiveListings, &activity.CompletedSales, &activity.TotalSalesValue,
		&activity.CompletedPurchases, &activity.TotalPurchaseValue,
	)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo actividad de mercado: %w", err)
	}

	return &activity, nil
}

// GetMarketTrends obtiene las tendencias del mercado
func (r *EconomyRepository) GetMarketTrends(period string) (*models.MarketTrends, error) {
	// TODO: Implementar obtención de tendencias del mercado
	// Por ahora retornamos tendencias básicas
	return &models.MarketTrends{
		Period:        period,
		PriceChanges:  []*models.PriceChange{},
		VolumeChanges: []*models.VolumeChange{},
	}, nil
}

// UpdatePlayerReputation actualiza la reputación de un jugador
func (r *EconomyRepository) UpdatePlayerReputation(playerID int, reputationDelta int) error {
	query := `
		UPDATE player_economy 
		SET reputation = GREATEST(0, reputation + $1), updated_at = $2
		WHERE player_id = $3
	`

	_, err := r.db.Exec(query, reputationDelta, time.Now(), playerID)
	if err != nil {
		return fmt.Errorf("error actualizando reputación: %w", err)
	}

	return nil
}

// GetCurrencyExchangeHistory obtiene el historial de intercambios de un jugador
func (r *EconomyRepository) GetCurrencyExchangeHistory(playerID uuid.UUID, limit int) ([]models.CurrencyExchange, error) {
	query := `
		SELECT id, player_id, from_currency_id, to_currency_id, from_amount,
		       to_amount, exchange_rate, fee, completed_at
		FROM currency_exchanges
		WHERE player_id = $1
		ORDER BY completed_at DESC
		LIMIT $2
	`

	rows, err := r.db.Query(query, playerID.String(), limit)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo historial: %w", err)
	}
	defer rows.Close()

	var exchanges []models.CurrencyExchange
	for rows.Next() {
		var exchange models.CurrencyExchange
		err := rows.Scan(
			&exchange.ID, &exchange.PlayerID, &exchange.FromCurrencyID, &exchange.ToCurrencyID,
			&exchange.FromAmount, &exchange.ToAmount, &exchange.ExchangeRate, &exchange.Fee, &exchange.CompletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando intercambio: %w", err)
		}
		exchanges = append(exchanges, exchange)
	}

	return exchanges, nil
}

// GetPlayerResources obtiene los recursos de un jugador
func (r *EconomyRepository) GetPlayerResources(playerID uuid.UUID) (*models.PlayerResources, error) {
	query := `
		SELECT player_id, gold, silver, copper, gems, premium_currency,
		       wood, stone, iron, food, population, max_population,
		       storage_capacity, last_updated, created_at
		FROM player_resources
		WHERE player_id = $1
	`

	var resources models.PlayerResources
	err := r.db.QueryRow(query, playerID).Scan(
		&resources.PlayerID, &resources.Gold, &resources.Silver, &resources.Copper,
		&resources.Gems, &resources.PremiumCurrency, &resources.Wood, &resources.Stone,
		&resources.Iron, &resources.Food, &resources.Population, &resources.MaxPopulation,
		&resources.StorageCapacity, &resources.LastUpdated, &resources.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			// Crear recursos por defecto si no existen
			return r.createDefaultResources(playerID)
		}
		return nil, fmt.Errorf("error obteniendo recursos: %w", err)
	}

	return &resources, nil
}

// UpdatePlayerResources actualiza los recursos de un jugador
func (r *EconomyRepository) UpdatePlayerResources(resources *models.PlayerResources) error {
	query := `
		INSERT INTO player_resources (
			player_id, gold, silver, copper, gems, premium_currency,
			wood, stone, iron, food, population, max_population,
			storage_capacity, last_updated, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
		)
		ON CONFLICT (player_id) DO UPDATE SET
			gold = EXCLUDED.gold,
			silver = EXCLUDED.silver,
			copper = EXCLUDED.copper,
			gems = EXCLUDED.gems,
			premium_currency = EXCLUDED.premium_currency,
			wood = EXCLUDED.wood,
			stone = EXCLUDED.stone,
			iron = EXCLUDED.iron,
			food = EXCLUDED.food,
			population = EXCLUDED.population,
			max_population = EXCLUDED.max_population,
			storage_capacity = EXCLUDED.storage_capacity,
			last_updated = EXCLUDED.last_updated
	`

	now := time.Now()
	resources.LastUpdated = now

	_, err := r.db.Exec(query,
		resources.PlayerID, resources.Gold, resources.Silver, resources.Copper,
		resources.Gems, resources.PremiumCurrency, resources.Wood, resources.Stone,
		resources.Iron, resources.Food, resources.Population, resources.MaxPopulation,
		resources.StorageCapacity, now, resources.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("error actualizando recursos: %w", err)
	}

	return nil
}

// AddResources añade recursos a un jugador
func (r *EconomyRepository) AddResources(playerID uuid.UUID, resourceType string, amount int, reason string) error {
	// Obtener recursos actuales
	resources, err := r.GetPlayerResources(playerID)
	if err != nil {
		return fmt.Errorf("error obteniendo recursos: %w", err)
	}

	// Actualizar el recurso específico
	switch resourceType {
	case "gold":
		resources.Gold += amount
	case "silver":
		resources.Silver += amount
	case "copper":
		resources.Copper += amount
	case "gems":
		resources.Gems += amount
	case "premium_currency":
		resources.PremiumCurrency += amount
	case "wood":
		resources.Wood += amount
	case "stone":
		resources.Stone += amount
	case "iron":
		resources.Iron += amount
	case "food":
		resources.Food += amount
	default:
		return fmt.Errorf("tipo de recurso no válido: %s", resourceType)
	}

	// Actualizar en la base de datos
	err = r.UpdatePlayerResources(resources)
	if err != nil {
		return fmt.Errorf("error actualizando recursos: %w", err)
	}

	// Registrar la transacción
	err = r.recordTransaction(playerID, resourceType, amount, "add", reason)
	if err != nil {
		return fmt.Errorf("error registrando transacción: %w", err)
	}

	return nil
}

// RemoveResources remueve recursos de un jugador
func (r *EconomyRepository) RemoveResources(playerID uuid.UUID, resourceType string, amount int, reason string) error {
	// Obtener recursos actuales
	resources, err := r.GetPlayerResources(playerID)
	if err != nil {
		return fmt.Errorf("error obteniendo recursos: %w", err)
	}

	// Verificar que hay suficientes recursos
	var currentAmount int
	switch resourceType {
	case "gold":
		currentAmount = resources.Gold
	case "silver":
		currentAmount = resources.Silver
	case "copper":
		currentAmount = resources.Copper
	case "gems":
		currentAmount = resources.Gems
	case "premium_currency":
		currentAmount = resources.PremiumCurrency
	case "wood":
		currentAmount = resources.Wood
	case "stone":
		currentAmount = resources.Stone
	case "iron":
		currentAmount = resources.Iron
	case "food":
		currentAmount = resources.Food
	default:
		return fmt.Errorf("tipo de recurso no válido: %s", resourceType)
	}

	if currentAmount < amount {
		return fmt.Errorf("recursos insuficientes: %s (tiene %d, necesita %d)", resourceType, currentAmount, amount)
	}

	// Actualizar el recurso específico
	switch resourceType {
	case "gold":
		resources.Gold -= amount
	case "silver":
		resources.Silver -= amount
	case "copper":
		resources.Copper -= amount
	case "gems":
		resources.Gems -= amount
	case "premium_currency":
		resources.PremiumCurrency -= amount
	case "wood":
		resources.Wood -= amount
	case "stone":
		resources.Stone -= amount
	case "iron":
		resources.Iron -= amount
	case "food":
		resources.Food -= amount
	}

	// Actualizar en la base de datos
	err = r.UpdatePlayerResources(resources)
	if err != nil {
		return fmt.Errorf("error actualizando recursos: %w", err)
	}

	// Registrar la transacción
	err = r.recordTransaction(playerID, resourceType, -amount, "remove", reason)
	if err != nil {
		return fmt.Errorf("error registrando transacción: %w", err)
	}

	return nil
}

// GetTransactionHistory obtiene el historial de transacciones de un jugador
func (r *EconomyRepository) GetTransactionHistory(playerID uuid.UUID, limit int) ([]models.ResourceTransaction, error) {
	query := `
		SELECT id, player_id, resource_type, amount, transaction_type,
		       reason, metadata, created_at
		FROM resource_transactions
		WHERE player_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := r.db.Query(query, playerID, limit)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo historial: %w", err)
	}
	defer rows.Close()

	var transactions []models.ResourceTransaction
	for rows.Next() {
		var transaction models.ResourceTransaction
		err := rows.Scan(
			&transaction.ID, &transaction.PlayerID, &transaction.ResourceType,
			&transaction.Amount, &transaction.TransactionType, &transaction.Reason,
			&transaction.Metadata, &transaction.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando transacción: %w", err)
		}
		transactions = append(transactions, transaction)
	}

	return transactions, nil
}

// GetPlayerEconomyStatistics obtiene las estadísticas económicas de un jugador
func (r *EconomyRepository) GetPlayerEconomyStatistics(playerID uuid.UUID) (*models.EconomyStatistics, error) {
	query := `
		SELECT player_id, total_transactions, total_income, total_expenses,
		       net_worth, most_used_resource, highest_transaction,
		       average_transaction, last_transaction, created_at, updated_at
		FROM economy_statistics
		WHERE player_id = $1
	`

	var stats models.EconomyStatistics
	err := r.db.QueryRow(query, playerID).Scan(
		&stats.PlayerID, &stats.TotalTransactions, &stats.TotalIncome,
		&stats.TotalExpenses, &stats.NetWorth, &stats.MostUsedResource,
		&stats.HighestTransaction, &stats.AverageTransaction,
		&stats.LastTransaction, &stats.CreatedAt, &stats.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			// Crear estadísticas si no existen
			return r.createDefaultEconomyStatistics(playerID)
		}
		return nil, fmt.Errorf("error obteniendo estadísticas: %w", err)
	}

	return &stats, nil
}

// createDefaultResources crea recursos por defecto para un jugador
func (r *EconomyRepository) createDefaultResources(playerID uuid.UUID) (*models.PlayerResources, error) {
	now := time.Now()
	resources := &models.PlayerResources{
		PlayerID:        playerID,
		Gold:            1000,
		Silver:          5000,
		Copper:          10000,
		Gems:            50,
		PremiumCurrency: 10,
		Wood:            100,
		Stone:           100,
		Iron:            50,
		Food:            200,
		Population:      10,
		MaxPopulation:   50,
		StorageCapacity: 1000,
		LastUpdated:     now,
		CreatedAt:       now,
	}

	err := r.UpdatePlayerResources(resources)
	if err != nil {
		return nil, fmt.Errorf("error creando recursos por defecto: %w", err)
	}

	return resources, nil
}

// createDefaultEconomyStatistics crea estadísticas por defecto para un jugador
func (r *EconomyRepository) createDefaultEconomyStatistics(playerID uuid.UUID) (*models.EconomyStatistics, error) {
	now := time.Now()
	stats := &models.EconomyStatistics{
		PlayerID:           playerID,
		TotalTransactions:  0,
		TotalIncome:        0,
		TotalExpenses:      0,
		NetWorth:           0,
		MostUsedResource:   "",
		HighestTransaction: 0,
		AverageTransaction: 0.0,
		LastTransaction:    nil,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	query := `
		INSERT INTO economy_statistics (
			player_id, total_transactions, total_income, total_expenses,
			net_worth, most_used_resource, highest_transaction,
			average_transaction, last_transaction, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
		)
	`

	_, err := r.db.Exec(query,
		stats.PlayerID, stats.TotalTransactions, stats.TotalIncome,
		stats.TotalExpenses, stats.NetWorth, stats.MostUsedResource,
		stats.HighestTransaction, stats.AverageTransaction,
		stats.LastTransaction, stats.CreatedAt, stats.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("error creando estadísticas por defecto: %w", err)
	}

	return stats, nil
}

// recordTransaction registra una transacción en el historial
func (r *EconomyRepository) recordTransaction(playerID uuid.UUID, resourceType string, amount int, transactionType, reason string) error {
	query := `
		INSERT INTO resource_transactions (
			id, player_id, resource_type, amount, transaction_type,
			reason, metadata, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8
		)
	`

	now := time.Now()
	metadata := fmt.Sprintf(`{"reason": "%s", "timestamp": "%s"}`, reason, now.Format(time.RFC3339))

	_, err := r.db.Exec(query,
		uuid.New(), playerID, resourceType, amount, transactionType,
		reason, metadata, now,
	)

	if err != nil {
		return fmt.Errorf("error registrando transacción: %w", err)
	}

	return nil
}

// GetPlayerEconomyStats obtiene las estadísticas económicas de un jugador
func (r *EconomyRepository) GetPlayerEconomyStats(playerID string) (*models.EconomyStatistics, error) {
	// Mock temporal
	return &models.EconomyStatistics{}, nil
}

// GetRecentTransactions obtiene las transacciones recientes
func (r *EconomyRepository) GetRecentTransactions() ([]models.MarketTransaction, error) {
	// Mock temporal
	return []models.MarketTransaction{}, nil
}

// GetMarketItem obtiene un item específico del mercado
func (r *EconomyRepository) GetMarketItem(itemID uuid.UUID) (*models.MarketItem, error) {
	// Mock temporal
	return &models.MarketItem{
		ID:          itemID.String(),
		Name:        "Mock Item",
		Description: "Mock description",
		Category:    "mock",
		BasePrice:   100.0,
		IsActive:    true,
	}, nil
}

// ProcessCurrencyExchange procesa un intercambio de monedas
func (r *EconomyRepository) ProcessCurrencyExchange(exchange *models.CurrencyExchange) error {
	// Implementar lógica de procesamiento de intercambio
	// Por ahora, solo actualizar el estado
	query := `
		UPDATE currency_exchanges 
		SET status = $1, processed_at = $2, updated_at = $3
		WHERE id = $4
	`

	_, err := r.db.Exec(query, "completed", time.Now(), time.Now(), exchange.ID)
	if err != nil {
		return fmt.Errorf("error procesando intercambio: %w", err)
	}

	return nil
}

// GetTotalTrades obtiene el total de transacciones registradas
func (r *EconomyRepository) GetTotalTrades() (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM market_transactions`
	err := r.db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("error contando transacciones: %w", err)
	}
	return count, nil
}

// GetTradesToday obtiene el número de transacciones de hoy
func (r *EconomyRepository) GetTradesToday() (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM market_transactions WHERE DATE(created_at) = CURRENT_DATE`
	err := r.db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("error contando transacciones de hoy: %w", err)
	}
	return count, nil
}
