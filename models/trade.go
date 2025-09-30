package models

import (
	"time"

	"github.com/google/uuid"
)

// TradeOffer representa una oferta en el mercado
type TradeOffer struct {
	ID           uuid.UUID `json:"id" db:"id"`
	SellerID     uuid.UUID `json:"seller_id" db:"seller_id"`
	VillageID    uuid.UUID `json:"village_id" db:"village_id"`
	ResourceType string    `json:"resource_type" db:"resource_type"` // wood, stone, food, gold
	Amount       int       `json:"amount" db:"amount"`
	PricePerUnit int       `json:"price_per_unit" db:"price_per_unit"`
	Status       string    `json:"status" db:"status"` // active, completed, cancelled
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// TradeTransaction representa una transacción de comercio
type TradeTransaction struct {
	ID              uuid.UUID `json:"id" db:"id"`
	OfferID         uuid.UUID `json:"offer_id" db:"offer_id"`
	BuyerID         uuid.UUID `json:"buyer_id" db:"buyer_id"`
	SellerID        uuid.UUID `json:"seller_id" db:"seller_id"`
	BuyerVillageID  uuid.UUID `json:"buyer_village_id" db:"buyer_village_id"`
	SellerVillageID uuid.UUID `json:"seller_village_id" db:"seller_village_id"`
	ResourceType    string    `json:"resource_type" db:"resource_type"`
	Amount          int       `json:"amount" db:"amount"`
	PricePerUnit    int       `json:"price_per_unit" db:"price_per_unit"`
	TotalPrice      int       `json:"total_price" db:"total_price"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
}

// DirectTrade representa un intercambio directo entre jugadores
type DirectTrade struct {
	ID                 uuid.UUID `json:"id" db:"id"`
	InitiatorID        uuid.UUID `json:"initiator_id" db:"initiator_id"`
	InitiatorVillageID uuid.UUID `json:"initiator_village_id" db:"initiator_village_id"`
	TargetID           uuid.UUID `json:"target_id" db:"target_id"`
	TargetVillageID    uuid.UUID `json:"target_village_id" db:"target_village_id"`
	OfferedResource    string    `json:"offered_resource" db:"offered_resource"`
	OfferedAmount      int       `json:"offered_amount" db:"offered_amount"`
	RequestedResource  string    `json:"requested_resource" db:"requested_resource"`
	RequestedAmount    int       `json:"requested_amount" db:"requested_amount"`
	Status             string    `json:"status" db:"status"` // pending, accepted, declined, expired
	Message            string    `json:"message" db:"message"`
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
	ExpiresAt          time.Time `json:"expires_at" db:"expires_at"`
}

// MarketStats representa estadísticas del mercado
type MarketStats struct {
	ResourceType string    `json:"resource_type" db:"resource_type"`
	AveragePrice float64   `json:"average_price" db:"average_price"`
	MinPrice     int       `json:"min_price" db:"min_price"`
	MaxPrice     int       `json:"max_price" db:"max_price"`
	TotalVolume  int       `json:"total_volume" db:"total_volume"`
	ActiveOffers int       `json:"active_offers" db:"active_offers"`
	LastUpdated  time.Time `json:"last_updated" db:"last_updated"`
}

// ResourcePrice representa el precio actual de un recurso
type ResourcePrice struct {
	ResourceType string    `json:"resource_type" db:"resource_type"`
	CurrentPrice int       `json:"current_price" db:"current_price"`
	Change24h    float64   `json:"change_24h" db:"change_24h"`
	Volume24h    int       `json:"volume_24h" db:"volume_24h"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// TradeHistory representa el historial de transacciones de un jugador
type TradeHistory struct {
	ID            uuid.UUID `json:"id" db:"id"`
	TransactionID uuid.UUID `json:"transaction_id" db:"transaction_id"`
	PlayerID      uuid.UUID `json:"player_id" db:"player_id"`
	Type          string    `json:"type" db:"type"` // buy, sell
	ResourceType  string    `json:"resource_type" db:"resource_type"`
	Amount        int       `json:"amount" db:"amount"`
	PricePerUnit  int       `json:"price_per_unit" db:"price_per_unit"`
	TotalValue    int       `json:"total_value" db:"total_value"`
	Counterparty  string    `json:"counterparty" db:"counterparty"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

// TradeNotification representa una notificación de comercio
type TradeNotification struct {
	ID        uuid.UUID `json:"id" db:"id"`
	PlayerID  uuid.UUID `json:"player_id" db:"player_id"`
	Type      string    `json:"type" db:"type"` // offer_sold, offer_bought, direct_trade_received, etc.
	Message   string    `json:"message" db:"message"`
	Data      string    `json:"data" db:"data"` // JSON con datos adicionales
	Read      bool      `json:"read" db:"read"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
