package repository

import (
	"database/sql"
	"fmt"
	"time"

	"server-backend/models"

	"github.com/google/uuid"
)

type TradeRepository struct {
	db *sql.DB
}

func NewTradeRepository(db *sql.DB) *TradeRepository {
	return &TradeRepository{db: db}
}

// CreateTradeOffer crea una nueva oferta de comercio
func (r *TradeRepository) CreateTradeOffer(offer *models.TradeOffer) (*models.TradeOffer, error) {
	query := `
		INSERT INTO trade_offers (
			id, seller_id, village_id, resource_type, amount, price_per_unit,
			status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	now := time.Now()
	offer.ID = uuid.New()
	offer.CreatedAt = now
	offer.UpdatedAt = now

	_, err := r.db.Exec(query,
		offer.ID, offer.SellerID, offer.VillageID, offer.ResourceType, offer.Amount,
		offer.PricePerUnit, offer.Status, offer.CreatedAt, offer.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("error creating trade offer: %v", err)
	}

	return offer, nil
}

// GetTradeOffer obtiene una oferta específica
func (r *TradeRepository) GetTradeOffer(offerID uuid.UUID) (*models.TradeOffer, error) {
	query := `SELECT id, seller_id, village_id, resource_type, amount, price_per_unit, status, created_at, updated_at FROM trade_offers WHERE id = $1`

	var offer models.TradeOffer
	err := r.db.QueryRow(query, offerID).Scan(
		&offer.ID, &offer.SellerID, &offer.VillageID, &offer.ResourceType,
		&offer.Amount, &offer.PricePerUnit, &offer.Status,
		&offer.CreatedAt, &offer.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("error getting trade offer: %v", err)
	}

	return &offer, nil
}

// GetTradeOffers obtiene ofertas con filtros
func (r *TradeRepository) GetTradeOffers(resourceType, sellerID, priceMin, priceMax string) ([]models.TradeOffer, error) {
	query := `SELECT id, seller_id, village_id, resource_type, amount, price_per_unit, status, created_at, updated_at FROM trade_offers WHERE status = 'active'`
	args := []interface{}{}

	if resourceType != "" {
		query += ` AND resource_type = $1`
		args = append(args, resourceType)
	}

	if sellerID != "" {
		query += ` AND seller_id = $2`
		args = append(args, sellerID)
	}

	if priceMin != "" {
		query += ` AND price_per_unit >= $3`
		args = append(args, priceMin)
	}

	if priceMax != "" {
		query += ` AND price_per_unit <= $4`
		args = append(args, priceMax)
	}

	query += ` ORDER BY created_at DESC`

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("error getting trade offers: %v", err)
	}
	defer rows.Close()

	var offers []models.TradeOffer
	for rows.Next() {
		var offer models.TradeOffer
		err := rows.Scan(
			&offer.ID, &offer.SellerID, &offer.VillageID, &offer.ResourceType,
			&offer.Amount, &offer.PricePerUnit, &offer.Status,
			&offer.CreatedAt, &offer.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning trade offer: %v", err)
		}
		offers = append(offers, offer)
	}

	return offers, nil
}

// GetPlayerTradeOffers obtiene las ofertas de un jugador
func (r *TradeRepository) GetPlayerTradeOffers(playerID uuid.UUID) ([]models.TradeOffer, error) {
	query := `SELECT id, seller_id, village_id, resource_type, amount, price_per_unit, status, created_at, updated_at FROM trade_offers WHERE seller_id = $1 ORDER BY created_at DESC`

	rows, err := r.db.Query(query, playerID)
	if err != nil {
		return nil, fmt.Errorf("error getting player trade offers: %v", err)
	}
	defer rows.Close()

	var offers []models.TradeOffer
	for rows.Next() {
		var offer models.TradeOffer
		err := rows.Scan(
			&offer.ID, &offer.SellerID, &offer.VillageID, &offer.ResourceType,
			&offer.Amount, &offer.PricePerUnit, &offer.Status,
			&offer.CreatedAt, &offer.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning trade offer: %v", err)
		}
		offers = append(offers, offer)
	}

	return offers, nil
}

// ProcessTrade procesa una transacción de comercio
func (r *TradeRepository) ProcessTrade(offerID, buyerID, buyerVillageID uuid.UUID, amount int) (*models.TradeTransaction, error) {
	// Obtener la oferta
	offer, err := r.GetTradeOffer(offerID)
	if err != nil {
		return nil, fmt.Errorf("error getting trade offer: %v", err)
	}

	if offer.Status != "active" {
		return nil, fmt.Errorf("offer is not active")
	}

	if offer.Amount < amount {
		return nil, fmt.Errorf("insufficient amount in offer")
	}

	// Crear la transacción
	transaction := &models.TradeTransaction{
		ID:              uuid.New(),
		OfferID:         offerID,
		SellerID:        offer.SellerID,
		BuyerID:         buyerID,
		SellerVillageID: offer.VillageID,
		BuyerVillageID:  buyerVillageID,
		ResourceType:    offer.ResourceType,
		Amount:          amount,
		PricePerUnit:    offer.PricePerUnit,
		TotalPrice:      offer.PricePerUnit * amount,
		CreatedAt:       time.Now(),
	}

	query := `
		INSERT INTO trade_transactions (
			id, offer_id, seller_id, buyer_id, seller_village_id, buyer_village_id,
			resource_type, amount, price_per_unit, total_price, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err = r.db.Exec(query,
		transaction.ID, transaction.OfferID, transaction.SellerID, transaction.BuyerID,
		transaction.SellerVillageID, transaction.BuyerVillageID,
		transaction.ResourceType, transaction.Amount, transaction.PricePerUnit,
		transaction.TotalPrice, transaction.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("error creating trade transaction: %v", err)
	}

	// Actualizar la oferta
	if offer.Amount == amount {
		// Oferta completamente vendida
		_, err = r.db.Exec("UPDATE trade_offers SET status = 'completed', updated_at = $1 WHERE id = $2", time.Now(), offerID)
	} else {
		// Oferta parcialmente vendida
		_, err = r.db.Exec("UPDATE trade_offers SET amount = amount - $1, updated_at = $2 WHERE id = $3", amount, time.Now(), offerID)
	}
	if err != nil {
		return nil, fmt.Errorf("error updating trade offer: %v", err)
	}

	return transaction, nil
}

// CancelTradeOffer cancela una oferta de comercio
func (r *TradeRepository) CancelTradeOffer(offerID uuid.UUID) error {
	query := `UPDATE trade_offers SET status = 'cancelled', updated_at = $1 WHERE id = $2`
	_, err := r.db.Exec(query, time.Now(), offerID)
	if err != nil {
		return fmt.Errorf("error cancelling trade offer: %v", err)
	}
	return nil
}

// DeleteTradeOffer elimina una oferta de comercio
func (r *TradeRepository) DeleteTradeOffer(offerID uuid.UUID) error {
	query := `DELETE FROM trade_offers WHERE id = $1`
	_, err := r.db.Exec(query, offerID)
	if err != nil {
		return fmt.Errorf("error deleting trade offer: %v", err)
	}
	return nil
}

// GetTradeHistory obtiene el historial de transacciones de un jugador
func (r *TradeRepository) GetTradeHistory(playerID uuid.UUID, limit int) ([]models.TradeTransaction, error) {
	query := `
		SELECT id, offer_id, seller_id, buyer_id, seller_village_id, buyer_village_id,
		       resource_type, amount, price_per_unit, total_price, created_at
		FROM trade_transactions
		WHERE seller_id = $1 OR buyer_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := r.db.Query(query, playerID, limit)
	if err != nil {
		return nil, fmt.Errorf("error getting trade history: %v", err)
	}
	defer rows.Close()

	var transactions []models.TradeTransaction
	for rows.Next() {
		var transaction models.TradeTransaction
		err := rows.Scan(
			&transaction.ID, &transaction.OfferID, &transaction.SellerID, &transaction.BuyerID,
			&transaction.SellerVillageID, &transaction.BuyerVillageID,
			&transaction.ResourceType, &transaction.Amount, &transaction.PricePerUnit,
			&transaction.TotalPrice, &transaction.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning trade transaction: %v", err)
		}
		transactions = append(transactions, transaction)
	}

	return transactions, nil
}

// GetTradeOffersByVillage obtiene las ofertas de una aldea específica
func (r *TradeRepository) GetTradeOffersByVillage(villageID uuid.UUID) ([]models.TradeOffer, error) {
	query := `
		SELECT id, seller_id, village_id, resource_type, amount, price_per_unit, status, created_at, updated_at
		FROM trade_offers
		WHERE village_id = $1 AND status = 'active'
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, villageID)
	if err != nil {
		return nil, fmt.Errorf("error getting trade offers by village: %v", err)
	}
	defer rows.Close()

	var offers []models.TradeOffer
	for rows.Next() {
		var offer models.TradeOffer
		err := rows.Scan(
			&offer.ID, &offer.SellerID, &offer.VillageID, &offer.ResourceType,
			&offer.Amount, &offer.PricePerUnit, &offer.Status,
			&offer.CreatedAt, &offer.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning trade offer: %v", err)
		}
		offers = append(offers, offer)
	}

	return offers, nil
}

// GetTradeOffersByResourceType obtiene las ofertas por tipo de recurso
func (r *TradeRepository) GetTradeOffersByResourceType(resourceType string) ([]models.TradeOffer, error) {
	query := `
		SELECT id, seller_id, village_id, resource_type, amount, price_per_unit, status, created_at, updated_at
		FROM trade_offers
		WHERE resource_type = $1 AND status = 'active'
		ORDER BY price_per_unit ASC, created_at DESC
	`

	rows, err := r.db.Query(query, resourceType)
	if err != nil {
		return nil, fmt.Errorf("error getting trade offers by resource type: %v", err)
	}
	defer rows.Close()

	var offers []models.TradeOffer
	for rows.Next() {
		var offer models.TradeOffer
		err := rows.Scan(
			&offer.ID, &offer.SellerID, &offer.VillageID, &offer.ResourceType,
			&offer.Amount, &offer.PricePerUnit, &offer.Status,
			&offer.CreatedAt, &offer.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning trade offer: %v", err)
		}
		offers = append(offers, offer)
	}

	return offers, nil
}

// GetTradeOffersByPriceRange obtiene las ofertas por rango de precio
func (r *TradeRepository) GetTradeOffersByPriceRange(minPrice, maxPrice int) ([]models.TradeOffer, error) {
	query := `
		SELECT id, seller_id, village_id, resource_type, amount, price_per_unit, status, created_at, updated_at
		FROM trade_offers
		WHERE price_per_unit >= $1 AND price_per_unit <= $2 AND status = 'active'
		ORDER BY price_per_unit ASC, created_at DESC
	`

	rows, err := r.db.Query(query, minPrice, maxPrice)
	if err != nil {
		return nil, fmt.Errorf("error getting trade offers by price range: %v", err)
	}
	defer rows.Close()

	var offers []models.TradeOffer
	for rows.Next() {
		var offer models.TradeOffer
		err := rows.Scan(
			&offer.ID, &offer.SellerID, &offer.VillageID, &offer.ResourceType,
			&offer.Amount, &offer.PricePerUnit, &offer.Status,
			&offer.CreatedAt, &offer.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning trade offer: %v", err)
		}
		offers = append(offers, offer)
	}

	return offers, nil
}

// GetTradeOffersBySeller obtiene las ofertas de un vendedor específico
func (r *TradeRepository) GetTradeOffersBySeller(sellerID uuid.UUID) ([]models.TradeOffer, error) {
	query := `
		SELECT id, seller_id, village_id, resource_type, amount, price_per_unit, status, created_at, updated_at
		FROM trade_offers
		WHERE seller_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, sellerID)
	if err != nil {
		return nil, fmt.Errorf("error getting trade offers by seller: %v", err)
	}
	defer rows.Close()

	var offers []models.TradeOffer
	for rows.Next() {
		var offer models.TradeOffer
		err := rows.Scan(
			&offer.ID, &offer.SellerID, &offer.VillageID, &offer.ResourceType,
			&offer.Amount, &offer.PricePerUnit, &offer.Status,
			&offer.CreatedAt, &offer.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning trade offer: %v", err)
		}
		offers = append(offers, offer)
	}

	return offers, nil
}
