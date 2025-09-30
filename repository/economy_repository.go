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
	query := `
		SELECT id, name, description, category, base_price, is_active
		FROM market_items 
		WHERE is_active = true
	`

	args := []interface{}{}
	argIndex := 1

	if category != "" {
		query += fmt.Sprintf(" AND category = $%d", argIndex)
		args = append(args, category)
		argIndex++
	}

	query += " ORDER BY base_price ASC"

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, limit)
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo items del mercado: %w", err)
	}
	defer rows.Close()

	var items []models.MarketItem
	for rows.Next() {
		var item models.MarketItem
		err := rows.Scan(
			&item.ID, &item.Name, &item.Description, &item.Category,
			&item.BasePrice, &item.IsActive,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando item del mercado: %w", err)
		}
		items = append(items, item)
	}

	return items, nil
}

// GetMarketListings obtiene las listas de venta del mercado
func (r *EconomyRepository) GetMarketListings(itemType string, itemID int, sellerID int, status string, limit int) ([]models.MarketListing, error) {
	query := `
		SELECT id, seller_id, item_name, quantity, 
		       price_per_unit, currency_id, created_at, expires_at, is_active
		FROM market_listings 
		WHERE 1=1
	`

	args := []interface{}{}
	argIndex := 1

	if itemType != "" {
		query += fmt.Sprintf(" AND item_type = $%d", argIndex)
		args = append(args, itemType)
		argIndex++
	}

	if itemID > 0 {
		query += fmt.Sprintf(" AND item_id = $%d", argIndex)
		args = append(args, itemID)
		argIndex++
	}

	if sellerID > 0 {
		query += fmt.Sprintf(" AND seller_id = $%d", argIndex)
		args = append(args, sellerID)
		argIndex++
	}

	query += " ORDER BY created_at DESC"

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, limit)
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo listas de venta: %w", err)
	}
	defer rows.Close()

	var listings []models.MarketListing
	for rows.Next() {
		var listing models.MarketListing
		err := rows.Scan(
			&listing.ID, &listing.SellerID, &listing.ItemName,
			&listing.Quantity, &listing.PricePerUnit,
			&listing.CurrencyID, &listing.CreatedAt, &listing.ExpiresAt, &listing.IsActive,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando lista de venta: %w", err)
		}
		listings = append(listings, listing)
	}

	return listings, nil
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
	// Obtener estadísticas del día
	query := `
		SELECT 
			COUNT(*) as total_transactions,
			SUM(total_price) as total_volume,
			AVG(total_price) as average_price
		FROM market_transactions 
		WHERE DATE(completed_at) = DATE($1)
	`

	var stats models.MarketStatistics
	err := r.db.QueryRow(query, date).Scan(
		&stats.TotalTransactions, &stats.TotalVolume,
		&stats.AveragePrice,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			// No hay transacciones para esta fecha
			stats.TotalTransactions = 0
			stats.TotalVolume = 0
			stats.AveragePrice = 0
		} else {
			return nil, fmt.Errorf("error obteniendo estadísticas del mercado: %w", err)
		}
	}

	stats.LastUpdated = time.Now()

	return &stats, nil
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
	// Obtener estadísticas de transacciones del jugador
	query := `
		SELECT 
			COUNT(*) as total_transactions,
			SUM(CASE WHEN buyer_id = $1 THEN total_price ELSE 0 END) as total_spent,
			SUM(CASE WHEN seller_id = $1 THEN total_price ELSE 0 END) as total_earned,
			AVG(total_price) as average_transaction_value
		FROM market_transactions 
		WHERE buyer_id = $1 OR seller_id = $1
	`

	var stats models.EconomyStatistics
	err := r.db.QueryRow(query, playerID).Scan(
		&stats.TotalTransactions, &stats.TotalIncome,
		&stats.TotalExpenses, &stats.AverageTransaction,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			// No hay transacciones para este jugador
			stats.TotalTransactions = 0
			stats.TotalIncome = 0
			stats.TotalExpenses = 0
			stats.AverageTransaction = 0
		} else {
			return nil, fmt.Errorf("error obteniendo estadísticas del jugador: %w", err)
		}
	}

	// Calcular net worth
	stats.NetWorth = stats.TotalIncome - stats.TotalExpenses

	stats.PlayerID = uuid.MustParse(playerID)
	stats.UpdatedAt = time.Now()

	return &stats, nil
}

// GetRecentTransactions obtiene las transacciones recientes
func (r *EconomyRepository) GetRecentTransactions() ([]models.MarketTransaction, error) {
	query := `
		SELECT id, buyer_id, seller_id, item_name, quantity,
		       price_per_unit, total_price, completed_at
		FROM market_transactions 
		ORDER BY completed_at DESC
		LIMIT 50
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo transacciones recientes: %w", err)
	}
	defer rows.Close()

	var transactions []models.MarketTransaction
	for rows.Next() {
		var transaction models.MarketTransaction
		err := rows.Scan(
			&transaction.ID, &transaction.BuyerID, &transaction.SellerID,
			&transaction.ItemName, &transaction.Quantity,
			&transaction.PricePerUnit, &transaction.TotalPrice,
			&transaction.CompletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando transacción: %w", err)
		}
		transactions = append(transactions, transaction)
	}

	return transactions, nil
}

// GetMarketItem obtiene un item específico del mercado
func (r *EconomyRepository) GetMarketItem(itemID uuid.UUID) (*models.MarketItem, error) {
	query := `
		SELECT id, name, description, category, base_price, is_active
		FROM market_items 
		WHERE id = $1 AND is_active = true
	`

	var item models.MarketItem
	err := r.db.QueryRow(query, itemID).Scan(
		&item.ID, &item.Name, &item.Description, &item.Category,
		&item.BasePrice, &item.IsActive,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("item del mercado no encontrado")
		}
		return nil, fmt.Errorf("error obteniendo item del mercado: %w", err)
	}

	return &item, nil
}

// GetTotalTrades obtiene el total de transacciones en el sistema
func (r *EconomyRepository) GetTotalTrades() (int, error) {
	query := `
		SELECT COUNT(*) 
		FROM market_transactions 
		WHERE created_at IS NOT NULL
	`
	
	var totalTrades int
	err := r.db.QueryRow(query).Scan(&totalTrades)
	if err != nil {
		return 0, fmt.Errorf("error obteniendo total de transacciones: %w", err)
	}
	
	return totalTrades, nil
}

// GetTradesToday obtiene el número de transacciones realizadas hoy
func (r *EconomyRepository) GetTradesToday() (int, error) {
	query := `
		SELECT COUNT(*) 
		FROM market_transactions 
		WHERE DATE(created_at) = CURRENT_DATE
	`
	
	var tradesToday int
	err := r.db.QueryRow(query).Scan(&tradesToday)
	if err != nil {
		return 0, fmt.Errorf("error obteniendo transacciones de hoy: %w", err)
	}
	
	return tradesToday, nil
}

// ProcessCurrencyExchange procesa un intercambio de monedas
func (r *EconomyRepository) ProcessCurrencyExchange(exchange *models.CurrencyExchange) error {
	// Implementar lógica de procesamiento de intercambio
	// Por ahora, solo actualizar el estado
	return nil
}

// ============================================================================
// FUNCIONES AUXILIARES PARA SISTEMA ECONÓMICO PROFESIONAL
// ============================================================================

// InitializePlayerResources inicializa los recursos de un jugador recién creado
func (r *EconomyRepository) InitializePlayerResources(playerID uuid.UUID, worldID uuid.UUID) error {
	// Recursos iniciales estándar
	initialResources := map[string]int{
		"gold":  1000,
		"wood":  500,
		"stone": 300,
		"food":  200,
	}

	// Crear registro de recursos del jugador
	query := `
		INSERT INTO player_resources (
			player_id, world_id, gold, wood, stone, food,
			last_updated, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8
		)
		ON CONFLICT (player_id, world_id) 
		DO UPDATE SET
			gold = EXCLUDED.gold,
			wood = EXCLUDED.wood,
			stone = EXCLUDED.stone,
			food = EXCLUDED.food,
			last_updated = EXCLUDED.last_updated
	`

	now := time.Now()
	_, err := r.db.Exec(query,
		playerID, worldID,
		initialResources["gold"], initialResources["wood"],
		initialResources["stone"], initialResources["food"],
		now, now,
	)

	if err != nil {
		return fmt.Errorf("error inicializando recursos del jugador: %w", err)
	}

	// Registrar transacción inicial
	err = r.RecordResourceTransaction(playerID, "initial_resources", initialResources, "system")
	if err != nil {
		r.logger.Warn("Error registrando transacción inicial", zap.Error(err))
	}

	return nil
}

// UpdatePlayerResourcesSafe actualiza los recursos de un jugador de manera segura
func (r *EconomyRepository) UpdatePlayerResourcesSafe(playerID uuid.UUID, worldID uuid.UUID, resourceChanges map[string]int) error {
	// Obtener recursos actuales
	currentResources, err := r.GetPlayerResources(playerID)
	if err != nil {
		return fmt.Errorf("error obteniendo recursos actuales: %w", err)
	}

	// Calcular nuevos recursos
	newResources := map[string]int{
		"gold":  currentResources.Gold + resourceChanges["gold"],
		"wood":  currentResources.Wood + resourceChanges["wood"],
		"stone": currentResources.Stone + resourceChanges["stone"],
		"food":  currentResources.Food + resourceChanges["food"],
	}

	// Validar que no haya recursos negativos
	for resource, amount := range newResources {
		if amount < 0 {
			return fmt.Errorf("recursos insuficientes: %s (disponible: %d, requerido: %d)",
				resource, currentResources.Gold, -resourceChanges["gold"])
		}
	}

	// Actualizar recursos
	query := `
		UPDATE player_resources 
		SET gold = $1, wood = $2, stone = $3, food = $4, last_updated = $5
		WHERE player_id = $6
	`

	_, err = r.db.Exec(query,
		newResources["gold"], newResources["wood"],
		newResources["stone"], newResources["food"],
		time.Now(), playerID,
	)

	if err != nil {
		return fmt.Errorf("error actualizando recursos: %w", err)
	}

	// Registrar transacción
	err = r.RecordResourceTransaction(playerID, "resource_update", resourceChanges, "system")
	if err != nil {
		r.logger.Warn("Error registrando transacción de recursos", zap.Error(err))
	}

	return nil
}

// RecordResourceTransaction registra una transacción de recursos
func (r *EconomyRepository) RecordResourceTransaction(playerID uuid.UUID, transactionType string, resourceChanges map[string]int, source string) error {
	query := `
		INSERT INTO resource_transactions (
			id, player_id, transaction_type, gold_change, wood_change,
			stone_change, food_change, source, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		)
	`

	_, err := r.db.Exec(query,
		uuid.New(), playerID, transactionType,
		resourceChanges["gold"], resourceChanges["wood"],
		resourceChanges["stone"], resourceChanges["food"],
		source, time.Now(),
	)

	if err != nil {
		return fmt.Errorf("error registrando transacción de recursos: %w", err)
	}

	return nil
}

// ValidateResourceTransaction valida si un jugador tiene suficientes recursos
func (r *EconomyRepository) ValidateResourceTransaction(playerID uuid.UUID, requiredResources map[string]int) error {
	resources, err := r.GetPlayerResources(playerID)
	if err != nil {
		return fmt.Errorf("error obteniendo recursos: %w", err)
	}

	// Verificar cada recurso requerido
	if requiredResources["gold"] > 0 && resources.Gold < requiredResources["gold"] {
		return fmt.Errorf("oro insuficiente: disponible %d, requerido %d", resources.Gold, requiredResources["gold"])
	}
	if requiredResources["wood"] > 0 && resources.Wood < requiredResources["wood"] {
		return fmt.Errorf("madera insuficiente: disponible %d, requerido %d", resources.Wood, requiredResources["wood"])
	}
	if requiredResources["stone"] > 0 && resources.Stone < requiredResources["stone"] {
		return fmt.Errorf("piedra insuficiente: disponible %d, requerido %d", resources.Stone, requiredResources["stone"])
	}
	if requiredResources["food"] > 0 && resources.Food < requiredResources["food"] {
		return fmt.Errorf("comida insuficiente: disponible %d, requerido %d", resources.Food, requiredResources["food"])
	}

	return nil
}

// ProcessMarketTransaction procesa una transacción de mercado completa
func (r *EconomyRepository) ProcessMarketTransaction(buyerID, sellerID uuid.UUID, itemName string, quantity int, pricePerUnit float64) error {
	totalPrice := float64(quantity) * pricePerUnit

	// Iniciar transacción de base de datos
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("error iniciando transacción: %w", err)
	}
	defer tx.Rollback()

	// Verificar que el comprador tiene suficientes recursos
	buyerResources, err := r.getPlayerResourcesTx(tx, buyerID)
	if err != nil {
		return fmt.Errorf("error obteniendo recursos del comprador: %w", err)
	}

	if buyerResources.Gold < int(totalPrice) {
		return fmt.Errorf("oro insuficiente: disponible %d, requerido %d", buyerResources.Gold, int(totalPrice))
	}

	// Actualizar recursos del comprador (restar oro)
	err = r.updatePlayerResourcesTx(tx, buyerID, map[string]int{"gold": -int(totalPrice)})
	if err != nil {
		return fmt.Errorf("error actualizando recursos del comprador: %w", err)
	}

	// Actualizar recursos del vendedor (sumar oro)
	err = r.updatePlayerResourcesTx(tx, sellerID, map[string]int{"gold": int(totalPrice)})
	if err != nil {
		return fmt.Errorf("error actualizando recursos del vendedor: %w", err)
	}

	// Registrar transacción de mercado
	transactionQuery := `
		INSERT INTO market_transactions (
			id, buyer_id, seller_id, item_name, quantity,
			price_per_unit, total_price, completed_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8
		)
	`

	_, err = tx.Exec(transactionQuery,
		uuid.New(), buyerID, sellerID, itemName, quantity,
		pricePerUnit, totalPrice, time.Now(),
	)

	if err != nil {
		return fmt.Errorf("error registrando transacción de mercado: %w", err)
	}

	// Confirmar transacción
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("error confirmando transacción: %w", err)
	}

	return nil
}

// Funciones auxiliares para transacciones de base de datos
func (r *EconomyRepository) getPlayerResourcesTx(tx *sql.Tx, playerID uuid.UUID) (*models.PlayerResources, error) {
	query := `
		SELECT player_id, gold, wood, stone, food,
		       last_updated, created_at
		FROM player_resources 
		WHERE player_id = $1
	`

	var resources models.PlayerResources
	err := tx.QueryRow(query, playerID).Scan(
		&resources.PlayerID, &resources.Gold,
		&resources.Wood, &resources.Stone, &resources.Food,
		&resources.LastUpdated, &resources.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &resources, nil
}

func (r *EconomyRepository) updatePlayerResourcesTx(tx *sql.Tx, playerID uuid.UUID, resourceChanges map[string]int) error {
	// Obtener recursos actuales
	currentResources, err := r.getPlayerResourcesTx(tx, playerID)
	if err != nil {
		return err
	}

	// Calcular nuevos recursos
	newGold := currentResources.Gold + resourceChanges["gold"]
	newWood := currentResources.Wood + resourceChanges["wood"]
	newStone := currentResources.Stone + resourceChanges["stone"]
	newFood := currentResources.Food + resourceChanges["food"]

	// Actualizar recursos
	query := `
		UPDATE player_resources 
		SET gold = $1, wood = $2, stone = $3, food = $4, last_updated = $5
		WHERE player_id = $6
	`

	_, err = tx.Exec(query,
		newGold, newWood, newStone, newFood,
		time.Now(), playerID,
	)

	return err
}
