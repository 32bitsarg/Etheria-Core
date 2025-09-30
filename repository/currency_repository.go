package repository

import (
	"database/sql"
	"fmt"
	"server-backend/models"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type CurrencyRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewCurrencyRepository(db *sql.DB, logger *zap.Logger) *CurrencyRepository {
	return &CurrencyRepository{
		db:     db,
		logger: logger,
	}
}

// GetCurrencyConfig obtiene la configuración de monedas
func (r *CurrencyRepository) GetCurrencyConfig() (*models.CurrencyConfig, error) {
	var config models.CurrencyConfig
	err := r.db.QueryRow(`
		SELECT id, global_coin, world_coin, created_at, updated_at
		FROM currency_config
		ORDER BY created_at DESC
		LIMIT 1
	`).Scan(
		&config.ID,
		&config.GlobalCoin,
		&config.WorldCoin,
		&config.CreatedAt,
		&config.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		// Crear configuración por defecto si no existe
		return r.createDefaultConfig()
	}
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// UpdateCurrencyConfig actualiza la configuración de monedas
func (r *CurrencyRepository) UpdateCurrencyConfig(globalCoin, worldCoin string) (*models.CurrencyConfig, error) {
	config, err := r.GetCurrencyConfig()
	if err != nil {
		return nil, err
	}

	_, err = r.db.Exec(`
		UPDATE currency_config 
		SET global_coin = $1, world_coin = $2, updated_at = $3
		WHERE id = $4
	`, globalCoin, worldCoin, time.Now(), config.ID)
	if err != nil {
		return nil, err
	}

	// Retornar configuración actualizada
	return &models.CurrencyConfig{
		ID:         config.ID,
		GlobalCoin: globalCoin,
		WorldCoin:  worldCoin,
		CreatedAt:  config.CreatedAt,
		UpdatedAt:  time.Now(),
	}, nil
}

// createDefaultConfig crea una configuración por defecto
func (r *CurrencyRepository) createDefaultConfig() (*models.CurrencyConfig, error) {
	id := uuid.New()
	now := time.Now()

	_, err := r.db.Exec(`
		INSERT INTO currency_config (id, global_coin, world_coin, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`, id, "Oro", "Plata", now, now)
	if err != nil {
		return nil, err
	}

	return &models.CurrencyConfig{
		ID:         id,
		GlobalCoin: "Oro",
		WorldCoin:  "Plata",
		CreatedAt:  now,
		UpdatedAt:  now,
	}, nil
}

// GetPlayerCurrencyBalance obtiene el balance completo de monedas de un jugador
func (r *CurrencyRepository) GetPlayerCurrencyBalance(playerID uuid.UUID, worldID *uuid.UUID) (*models.CurrencyBalance, error) {
	var balance models.CurrencyBalance
	var currentWorld sql.NullString

	query := `
		SELECT 
			p.id as player_id,
			COALESCE(pgc.amount, 0) as global_amount,
			COALESCE(pwc.amount, 0) as world_amount,
			pwe.world_id as current_world,
			cc.global_coin,
			cc.world_coin
		FROM players p
		LEFT JOIN player_global_currency pgc ON p.id = pgc.player_id
		LEFT JOIN player_world_currency pwc ON p.id = pwc.player_id AND pwc.world_id = COALESCE($2, pwe.world_id)
		LEFT JOIN player_world_entries pwe ON p.id = pwe.player_id AND pwe.is_active = true
		CROSS JOIN currency_config cc
		WHERE p.id = $1
	`

	err := r.db.QueryRow(query, playerID, worldID).Scan(
		&balance.PlayerID,
		&balance.GlobalAmount,
		&balance.WorldAmount,
		&currentWorld,
		&balance.GlobalCoin,
		&balance.WorldCoin,
	)
	if err != nil {
		return nil, err
	}

	if currentWorld.Valid {
		if id, err := uuid.Parse(currentWorld.String); err == nil {
			balance.CurrentWorld = &id
		}
	}

	return &balance, nil
}

// AddGlobalCurrency agrega moneda global a un jugador
func (r *CurrencyRepository) AddGlobalCurrency(playerID uuid.UUID, amount int64, description string) error {
	_, err := r.db.Exec(`
		SELECT add_global_currency($1, $2, $3)
	`, playerID, amount, description)
	return err
}

// AddWorldCurrency agrega moneda de mundo a un jugador
func (r *CurrencyRepository) AddWorldCurrency(playerID, worldID uuid.UUID, amount int64, description string) error {
	_, err := r.db.Exec(`
		SELECT add_world_currency($1, $2, $3, $4)
	`, playerID, worldID, amount, description)
	return err
}

// SpendGlobalCurrency gasta moneda global de un jugador
func (r *CurrencyRepository) SpendGlobalCurrency(playerID uuid.UUID, amount int64, description string) error {
	// Verificar que tenga suficientes fondos
	var currentBalance int64
	err := r.db.QueryRow(`
		SELECT COALESCE(amount, 0) FROM player_global_currency WHERE player_id = $1
	`, playerID).Scan(&currentBalance)
	if err == sql.ErrNoRows {
		return fmt.Errorf("el jugador no tiene moneda global")
	}
	if err != nil {
		return err
	}

	if currentBalance < amount {
		return fmt.Errorf("fondos insuficientes: tiene %d, necesita %d", currentBalance, amount)
	}

	// Realizar el gasto
	_, err = r.db.Exec(`
		UPDATE player_global_currency 
		SET amount = amount - $1, updated_at = NOW()
		WHERE player_id = $2
	`, amount, playerID)
	if err != nil {
		return err
	}

	// Registrar transacción
	newBalance := currentBalance - amount
	_, err = r.db.Exec(`
		INSERT INTO currency_transactions (player_id, currency_type, amount, type, description, balance)
		VALUES ($1, 'global', $2, 'spend', $3, $4)
	`, playerID, amount, description, newBalance)

	return err
}

// SpendWorldCurrency gasta moneda de mundo de un jugador
func (r *CurrencyRepository) SpendWorldCurrency(playerID, worldID uuid.UUID, amount int64, description string) error {
	// Verificar que tenga suficientes fondos
	var currentBalance int64
	err := r.db.QueryRow(`
		SELECT COALESCE(amount, 0) FROM player_world_currency WHERE player_id = $1 AND world_id = $2
	`, playerID, worldID).Scan(&currentBalance)
	if err == sql.ErrNoRows {
		return fmt.Errorf("el jugador no tiene moneda en este mundo")
	}
	if err != nil {
		return err
	}

	if currentBalance < amount {
		return fmt.Errorf("fondos insuficientes: tiene %d, necesita %d", currentBalance, amount)
	}

	// Realizar el gasto
	_, err = r.db.Exec(`
		UPDATE player_world_currency 
		SET amount = amount - $1, updated_at = NOW()
		WHERE player_id = $2 AND world_id = $3
	`, amount, playerID, worldID)
	if err != nil {
		return err
	}

	// Registrar transacción
	newBalance := currentBalance - amount
	_, err = r.db.Exec(`
		INSERT INTO currency_transactions (player_id, world_id, currency_type, amount, type, description, balance)
		VALUES ($1, $2, 'world', $3, 'spend', $4, $5)
	`, playerID, worldID, amount, description, newBalance)

	return err
}

// TransferGlobalCurrency transfiere moneda global entre jugadores
func (r *CurrencyRepository) TransferGlobalCurrency(fromPlayerID, toPlayerID uuid.UUID, amount int64, description string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Verificar que el jugador origen tenga suficientes fondos
	var fromBalance int64
	err = tx.QueryRow(`
		SELECT COALESCE(amount, 0) FROM player_global_currency WHERE player_id = $1
	`, fromPlayerID).Scan(&fromBalance)
	if err == sql.ErrNoRows {
		return fmt.Errorf("el jugador origen no tiene moneda global")
	}
	if err != nil {
		return err
	}

	if fromBalance < amount {
		return fmt.Errorf("fondos insuficientes: tiene %d, necesita %d", fromBalance, amount)
	}

	// Retirar del jugador origen
	_, err = tx.Exec(`
		UPDATE player_global_currency 
		SET amount = amount - $1, updated_at = NOW()
		WHERE player_id = $2
	`, amount, fromPlayerID)
	if err != nil {
		return err
	}

	// Agregar al jugador destino
	_, err = tx.Exec(`
		INSERT INTO player_global_currency (player_id, amount)
		VALUES ($1, $2)
		ON CONFLICT (player_id)
		DO UPDATE SET amount = player_global_currency.amount + $2, updated_at = NOW()
	`, toPlayerID, amount)
	if err != nil {
		return err
	}

	// Registrar transacciones
	fromNewBalance := fromBalance - amount
	_, err = tx.Exec(`
		INSERT INTO currency_transactions (player_id, currency_type, amount, type, description, balance)
		VALUES ($1, 'global', $2, 'transfer', $3, $4)
	`, fromPlayerID, -amount, "Transferencia enviada: "+description, fromNewBalance)
	if err != nil {
		return err
	}

	// Obtener balance del jugador destino
	var toNewBalance int64
	err = tx.QueryRow(`
		SELECT amount FROM player_global_currency WHERE player_id = $1
	`, toPlayerID).Scan(&toNewBalance)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		INSERT INTO currency_transactions (player_id, currency_type, amount, type, description, balance)
		VALUES ($1, 'global', $2, 'transfer', $3, $4)
	`, toPlayerID, amount, "Transferencia recibida: "+description, toNewBalance)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// TransferWorldCurrency transfiere moneda de mundo entre jugadores
func (r *CurrencyRepository) TransferWorldCurrency(fromPlayerID, toPlayerID, worldID uuid.UUID, amount int64, description string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Verificar que el jugador origen tenga suficientes fondos
	var fromBalance int64
	err = tx.QueryRow(`
		SELECT COALESCE(amount, 0) FROM player_world_currency WHERE player_id = $1 AND world_id = $2
	`, fromPlayerID, worldID).Scan(&fromBalance)
	if err == sql.ErrNoRows {
		return fmt.Errorf("el jugador origen no tiene moneda en este mundo")
	}
	if err != nil {
		return err
	}

	if fromBalance < amount {
		return fmt.Errorf("fondos insuficientes: tiene %d, necesita %d", fromBalance, amount)
	}

	// Retirar del jugador origen
	_, err = tx.Exec(`
		UPDATE player_world_currency 
		SET amount = amount - $1, updated_at = NOW()
		WHERE player_id = $2 AND world_id = $3
	`, amount, fromPlayerID, worldID)
	if err != nil {
		return err
	}

	// Agregar al jugador destino
	_, err = tx.Exec(`
		INSERT INTO player_world_currency (player_id, world_id, amount)
		VALUES ($1, $2, $3)
		ON CONFLICT (player_id, world_id)
		DO UPDATE SET amount = player_world_currency.amount + $3, updated_at = NOW()
	`, toPlayerID, worldID, amount)
	if err != nil {
		return err
	}

	// Registrar transacciones
	fromNewBalance := fromBalance - amount
	_, err = tx.Exec(`
		INSERT INTO currency_transactions (player_id, world_id, currency_type, amount, type, description, balance)
		VALUES ($1, $2, 'world', $3, 'transfer', $4, $5)
	`, fromPlayerID, worldID, -amount, "Transferencia enviada: "+description, fromNewBalance)
	if err != nil {
		return err
	}

	// Obtener balance del jugador destino
	var toNewBalance int64
	err = tx.QueryRow(`
		SELECT amount FROM player_world_currency WHERE player_id = $1 AND world_id = $2
	`, toPlayerID, worldID).Scan(&toNewBalance)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		INSERT INTO currency_transactions (player_id, world_id, currency_type, amount, type, description, balance)
		VALUES ($1, $2, 'world', $3, 'transfer', $4, $5)
	`, toPlayerID, worldID, amount, "Transferencia recibida: "+description, toNewBalance)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// GetCurrencyTransactions obtiene el historial de transacciones de un jugador
func (r *CurrencyRepository) GetCurrencyTransactions(playerID uuid.UUID, limit int) ([]models.CurrencyTransaction, error) {
	rows, err := r.db.Query(`
		SELECT id, player_id, world_id, currency_type, amount, type, description, balance, created_at
		FROM currency_transactions
		WHERE player_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, playerID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []models.CurrencyTransaction
	for rows.Next() {
		var transaction models.CurrencyTransaction
		var worldID sql.NullString
		err := rows.Scan(
			&transaction.ID,
			&transaction.PlayerID,
			&worldID,
			&transaction.CurrencyType,
			&transaction.Amount,
			&transaction.Type,
			&transaction.Description,
			&transaction.Balance,
			&transaction.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if worldID.Valid {
			if id, err := uuid.Parse(worldID.String); err == nil {
				transaction.WorldID = &id
			}
		}

		transactions = append(transactions, transaction)
	}

	return transactions, nil
}

// GetWorldCurrencyStats obtiene estadísticas de moneda de mundo para un mundo específico
func (r *CurrencyRepository) GetWorldCurrencyStats(worldID uuid.UUID) (map[string]interface{}, error) {
	var totalCurrency, totalPlayers, avgCurrency int64

	err := r.db.QueryRow(`
		SELECT 
			COALESCE(SUM(amount), 0) as total_currency,
			COUNT(*) as total_players,
			COALESCE(AVG(amount), 0) as avg_currency
		FROM player_world_currency
		WHERE world_id = $1
	`, worldID).Scan(&totalCurrency, &totalPlayers, &avgCurrency)
	if err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"totalCurrency": totalCurrency,
		"totalPlayers":  totalPlayers,
		"avgCurrency":   avgCurrency,
		"worldID":       worldID,
	}

	return stats, nil
}

// GetGlobalCurrencyStats obtiene estadísticas de moneda global
func (r *CurrencyRepository) GetGlobalCurrencyStats() (map[string]interface{}, error) {
	var totalCurrency, totalPlayers, avgCurrency int64

	err := r.db.QueryRow(`
		SELECT 
			COALESCE(SUM(amount), 0) as total_currency,
			COUNT(*) as total_players,
			COALESCE(AVG(amount), 0) as avg_currency
		FROM player_global_currency
	`).Scan(&totalCurrency, &totalPlayers, &avgCurrency)
	if err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"totalCurrency": totalCurrency,
		"totalPlayers":  totalPlayers,
		"avgCurrency":   avgCurrency,
	}

	return stats, nil
}
