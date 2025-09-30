package models

import (
	"time"

	"github.com/google/uuid"
)

// EconomySystemConfig representa la configuración del sistema de economía
type EconomySystemConfig struct {
	ID                      int                    `json:"id"`
	IsEnabled               bool                   `json:"is_enabled"`
	PrimaryCurrencyName     string                 `json:"primary_currency_name"`
	SecondaryCurrencyName   string                 `json:"secondary_currency_name"`
	PrimaryCurrencySymbol   string                 `json:"primary_currency_symbol"`
	SecondaryCurrencySymbol string                 `json:"secondary_currency_symbol"`
	ExchangeRate            float64                `json:"exchange_rate"`
	ExchangeFee             float64                `json:"exchange_fee"`
	MinExchangeAmount       int                    `json:"min_exchange_amount"`
	MaxExchangeAmount       int                    `json:"max_exchange_amount"`
	MarketTax               float64                `json:"market_tax"`
	TransactionFee          float64                `json:"transaction_fee"`
	MaxPriceFluctuation     float64                `json:"max_price_fluctuation"`
	PriceUpdateInterval     int                    `json:"price_update_interval"`
	MaxItemsPerPlayer       int                    `json:"max_items_per_player"`
	MaxActiveListings       int                    `json:"max_active_listings"`
	MinListingDuration      int                    `json:"min_listing_duration"`
	MaxListingDuration      int                    `json:"max_listing_duration"`
	AdvancedConfig          map[string]interface{} `json:"advanced_config"`
	CreatedAt               time.Time              `json:"created_at"`
	UpdatedAt               time.Time              `json:"updated_at"`
}

// EconomyDashboard representa el dashboard de economía
type EconomyDashboard struct {
	SystemConfig       EconomySystemConfig `json:"system_config"`
	PlayerEconomy      PlayerEconomy       `json:"player_economy"`
	PlayerStats        EconomyStatistics   `json:"player_stats"`
	ActiveListings     []MarketListing     `json:"active_listings"`
	MarketStats        MarketStatistics    `json:"market_stats"`
	MarketItems        []MarketItem        `json:"market_items"`
	MarketListings     []MarketListing     `json:"market_listings"`
	RecentTransactions []MarketTransaction `json:"recent_transactions"`
	MarketStatistics   MarketStatistics    `json:"market_statistics"`
	LastUpdated        time.Time           `json:"last_updated"`
}

// EconomyConfig representa la configuración del sistema económico
type EconomyConfig struct {
	Currencies []*Currency            `json:"currencies"`
	Taxes      []*Tax                 `json:"taxes"`
	Config     map[string]interface{} `json:"config"`
}

// Currency representa una moneda
type Currency struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Symbol       string  `json:"symbol"`
	ExchangeRate float64 `json:"exchange_rate"`
	IsActive     bool    `json:"is_active"`
}

// Tax representa un impuesto
type Tax struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Rate        float64 `json:"rate"`
	Description string  `json:"description"`
	IsActive    bool    `json:"is_active"`
}

// MarketItem representa un item del mercado
type MarketItem struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Category    string  `json:"category"`
	BasePrice   float64 `json:"base_price"`
	IsActive    bool    `json:"is_active"`
}

// MarketItemDetails representa detalles de un item del mercado
type MarketItemDetails struct {
	Item         *MarketItem     `json:"item"`
	PriceHistory []*PriceHistory `json:"price_history"`
}

// PriceHistory representa el historial de precios
type PriceHistory struct {
	Date  time.Time `json:"date"`
	Price float64   `json:"price"`
}

// MarketListing representa un listado del mercado
type MarketListing struct {
	ID           string    `json:"id"`
	SellerID     string    `json:"seller_id"`
	ItemName     string    `json:"item_name"`
	Quantity     int       `json:"quantity"`
	PricePerUnit float64   `json:"price_per_unit"`
	CurrencyID   string    `json:"currency_id"`
	CreatedAt    time.Time `json:"created_at"`
	ExpiresAt    time.Time `json:"expires_at"`
	IsActive     bool      `json:"is_active"`
}

// MarketTransaction representa una transacción del mercado
type MarketTransaction struct {
	ID             string    `json:"id"`
	BuyerID        string    `json:"buyer_id"`
	SellerID       string    `json:"seller_id"`
	ListingID      string    `json:"listing_id"`
	ItemName       string    `json:"item_name"`
	Quantity       int       `json:"quantity"`
	PricePerUnit   float64   `json:"price_per_unit"`
	TotalPrice     float64   `json:"total_price"`
	TransactionFee float64   `json:"transaction_fee"`
	CurrencyID     string    `json:"currency_id"`
	CompletedAt    time.Time `json:"completed_at"`
}

// MarketStatistics representa estadísticas del mercado
type MarketStatistics struct {
	TotalListings     int       `json:"total_listings"`
	TotalTransactions int       `json:"total_transactions"`
	AveragePrice      float64   `json:"average_price"`
	TotalVolume       float64   `json:"total_volume"`
	LastUpdated       time.Time `json:"last_updated"`
}

// MarketTrends representa tendencias del mercado
type MarketTrends struct {
	Period        string          `json:"period"`
	PriceChanges  []*PriceChange  `json:"price_changes"`
	VolumeChanges []*VolumeChange `json:"volume_changes"`
}

// PriceChange representa un cambio de precio
type PriceChange struct {
	ItemName   string  `json:"item_name"`
	Change     float64 `json:"change"`
	Percentage float64 `json:"percentage"`
}

// VolumeChange representa un cambio de volumen
type VolumeChange struct {
	ItemName   string  `json:"item_name"`
	Change     float64 `json:"change"`
	Percentage float64 `json:"percentage"`
}

// CurrencyExchange representa un intercambio de monedas
type CurrencyExchange struct {
	ID             string    `json:"id"`
	PlayerID       string    `json:"player_id"`
	FromCurrencyID string    `json:"from_currency_id"`
	ToCurrencyID   string    `json:"to_currency_id"`
	FromAmount     float64   `json:"from_amount"`
	ToAmount       float64   `json:"to_amount"`
	ExchangeRate   float64   `json:"exchange_rate"`
	Fee            float64   `json:"fee"`
	CompletedAt    time.Time `json:"completed_at"`
}

// PlayerEconomy representa la economía de un jugador
type PlayerEconomy struct {
	ID                 int        `json:"id"`
	PlayerID           uuid.UUID  `json:"player_id"`
	PrimaryCurrency    int        `json:"primary_currency"`
	SecondaryCurrency  int        `json:"secondary_currency"`
	TotalSpent         int        `json:"total_spent"`
	TotalEarned        int        `json:"total_earned"`
	ItemsSold          int        `json:"items_sold"`
	ItemsBought        int        `json:"items_bought"`
	Reputation         int        `json:"reputation"`
	TrustLevel         string     `json:"trust_level"`
	DailySpendingLimit int        `json:"daily_spending_limit"`
	DailyEarningLimit  int        `json:"daily_earning_limit"`
	CurrentDailySpent  int        `json:"current_daily_spent"`
	CurrentDailyEarned int        `json:"current_daily_earned"`
	LastTransaction    *time.Time `json:"last_transaction,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

// PlayerMarketActivity representa la actividad de mercado de un jugador
type PlayerMarketActivity struct {
	PlayerID           string  `json:"player_id"`
	ActiveListings     int     `json:"active_listings"`
	CompletedSales     int     `json:"completed_sales"`
	TotalSalesValue    float64 `json:"total_sales_value"`
	CompletedPurchases int     `json:"completed_purchases"`
	TotalPurchaseValue float64 `json:"total_purchase_value"`
}

// PlayerResources representa los recursos de un jugador
type PlayerResources struct {
	ID          uuid.UUID `json:"id" db:"id"`
	VillageID   uuid.UUID `json:"village_id" db:"village_id"`
	Wood        int       `json:"wood" db:"wood"`
	Stone       int       `json:"stone" db:"stone"`
	Food        int       `json:"food" db:"food"`
	Gold        int       `json:"gold" db:"gold"`
	LastUpdated time.Time `json:"last_updated" db:"last_updated"`
}

// ResourceTransaction representa una transacción de recursos
type ResourceTransaction struct {
	ID              uuid.UUID `json:"id" db:"id"`
	PlayerID        uuid.UUID `json:"player_id" db:"player_id"`
	ResourceType    string    `json:"resource_type" db:"resource_type"`
	Amount          int       `json:"amount" db:"amount"`
	TransactionType string    `json:"transaction_type" db:"transaction_type"` // add, remove, transfer
	Reason          string    `json:"reason" db:"reason"`
	Metadata        string    `json:"metadata" db:"metadata"` // JSON con datos adicionales
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
}

// EconomyStatistics representa estadísticas económicas de un jugador
type EconomyStatistics struct {
	PlayerID           uuid.UUID  `json:"player_id" db:"player_id"`
	TotalTransactions  int        `json:"total_transactions" db:"total_transactions"`
	TotalIncome        int        `json:"total_income" db:"total_income"`
	TotalExpenses      int        `json:"total_expenses" db:"total_expenses"`
	NetWorth           int        `json:"net_worth" db:"net_worth"`
	MostUsedResource   string     `json:"most_used_resource" db:"most_used_resource"`
	HighestTransaction int        `json:"highest_transaction" db:"highest_transaction"`
	AverageTransaction float64    `json:"average_transaction" db:"average_transaction"`
	LastTransaction    *time.Time `json:"last_transaction" db:"last_transaction"`
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at" db:"updated_at"`
}

// MarketPrice representa el precio de mercado de un recurso
type MarketPrice struct {
	ID           uuid.UUID `json:"id" db:"id"`
	ResourceType string    `json:"resource_type" db:"resource_type"`
	Currency     string    `json:"currency" db:"currency"`
	CurrentPrice int       `json:"current_price" db:"current_price"`
	MinPrice     int       `json:"min_price" db:"min_price"`
	MaxPrice     int       `json:"max_price" db:"max_price"`
	Volume24h    int       `json:"volume_24h" db:"volume_24h"`
	Change24h    float64   `json:"change_24h" db:"change_24h"`
	LastUpdated  time.Time `json:"last_updated" db:"last_updated"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// MarketTrend representa una tendencia del mercado
type MarketTrend struct {
	ID             uuid.UUID `json:"id" db:"id"`
	ResourceID     uuid.UUID `json:"resource_id" db:"resource_id"`
	TrendType      string    `json:"trend_type" db:"trend_type"` // rising, falling, stable, volatile
	Direction      string    `json:"direction" db:"direction"`   // up, down, sideways
	Strength       float64   `json:"strength" db:"strength"`     // 0.0 - 1.0
	Confidence     float64   `json:"confidence" db:"confidence"` // 0.0 - 1.0
	Duration       int       `json:"duration" db:"duration"`     // en horas
	StartPrice     float64   `json:"start_price" db:"start_price"`
	CurrentPrice   float64   `json:"current_price" db:"current_price"`
	PredictedPrice float64   `json:"predicted_price" db:"predicted_price"`
	Factors        string    `json:"factors" db:"factors"` // JSON con factores que influyen
	IsActive       bool      `json:"is_active" db:"is_active"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// MarketActivity representa la actividad del mercado
type MarketActivity struct {
	ID            uuid.UUID `json:"id" db:"id"`
	ResourceID    uuid.UUID `json:"resource_id" db:"resource_id"`
	ActivityType  string    `json:"activity_type" db:"activity_type"` // trade, price_change, volume_spike, etc.
	Volume        int       `json:"volume" db:"volume"`
	Price         float64   `json:"price" db:"price"`
	Change        float64   `json:"change" db:"change"` // cambio porcentual
	Buyers        int       `json:"buyers" db:"buyers"`
	Sellers       int       `json:"sellers" db:"sellers"`
	Data          string    `json:"data" db:"data"` // JSON con datos adicionales
	IsSignificant bool      `json:"is_significant" db:"is_significant"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

// CurrencyConfig representa la configuración de monedas del juego
type CurrencyConfig struct {
	ID         uuid.UUID `json:"id" db:"id"`
	GlobalCoin string    `json:"globalCoin" db:"global_coin"` // Nombre de la moneda global (ej: "Oro")
	WorldCoin  string    `json:"worldCoin" db:"world_coin"`   // Nombre de la moneda de mundo (ej: "Plata")
	CreatedAt  time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt  time.Time `json:"updatedAt" db:"updated_at"`
}

// PlayerGlobalCurrency representa la moneda global de un jugador
type PlayerGlobalCurrency struct {
	ID        uuid.UUID `json:"id" db:"id"`
	PlayerID  uuid.UUID `json:"playerId" db:"player_id"`
	Amount    int64     `json:"amount" db:"amount"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
}

// PlayerWorldCurrency representa la moneda de mundo de un jugador
type PlayerWorldCurrency struct {
	ID        uuid.UUID `json:"id" db:"id"`
	PlayerID  uuid.UUID `json:"playerId" db:"player_id"`
	WorldID   uuid.UUID `json:"worldId" db:"world_id"`
	Amount    int64     `json:"amount" db:"amount"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
}

// CurrencyTransaction representa una transacción de monedas
type CurrencyTransaction struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	PlayerID     uuid.UUID  `json:"playerId" db:"player_id"`
	WorldID      *uuid.UUID `json:"worldId,omitempty" db:"world_id"` // NULL para transacciones globales
	CurrencyType string     `json:"currencyType" db:"currency_type"` // "global" o "world"
	Amount       int64      `json:"amount" db:"amount"`
	Type         string     `json:"type" db:"type"` // "earn", "spend", "transfer"
	Description  string     `json:"description" db:"description"`
	Balance      int64      `json:"balance" db:"balance"` // Balance después de la transacción
	CreatedAt    time.Time  `json:"createdAt" db:"created_at"`
}

// CurrencyBalance representa el balance completo de un jugador
type CurrencyBalance struct {
	PlayerID     uuid.UUID  `json:"playerId"`
	GlobalAmount int64      `json:"globalAmount"`
	WorldAmount  int64      `json:"worldAmount"`
	CurrentWorld *uuid.UUID `json:"currentWorld,omitempty"`
	GlobalCoin   string     `json:"globalCoin"`
	WorldCoin    string     `json:"worldCoin"`
}

// UpdateCurrencyConfigRequest representa la solicitud para actualizar configuración de monedas
type UpdateCurrencyConfigRequest struct {
	GlobalCoin string `json:"globalCoin" validate:"required,min=1,max=50"`
	WorldCoin  string `json:"worldCoin" validate:"required,min=1,max=50"`
}

// AddCurrencyRequest representa la solicitud para agregar monedas
type AddCurrencyRequest struct {
	PlayerID     string `json:"playerId" validate:"required"`
	Amount       int64  `json:"amount" validate:"required"`
	CurrencyType string `json:"currencyType" validate:"required,oneof=global world"`
	WorldID      string `json:"worldId,omitempty"` // Requerido solo para world currency
	Description  string `json:"description" validate:"max=200"`
}

// TransferCurrencyRequest representa la solicitud para transferir monedas
type TransferCurrencyRequest struct {
	FromPlayerID string `json:"fromPlayerId" validate:"required"`
	ToPlayerID   string `json:"toPlayerId" validate:"required"`
	Amount       int64  `json:"amount" validate:"required,gt=0"`
	CurrencyType string `json:"currencyType" validate:"required,oneof=global world"`
	WorldID      string `json:"worldId,omitempty"` // Requerido solo para world currency
	Description  string `json:"description" validate:"max=200"`
}
