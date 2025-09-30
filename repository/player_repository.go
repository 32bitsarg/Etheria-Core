package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"server-backend/models"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

var ErrPlayerNotFound = errors.New("jugador no encontrado")

type PlayerRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewPlayerRepository(db *sql.DB, logger *zap.Logger) *PlayerRepository {
	return &PlayerRepository{
		db:     db,
		logger: logger,
	}
}

func (r *PlayerRepository) GetPlayerByUsername(username string) (*models.Player, error) {
	var player models.Player
	err := r.db.QueryRow(`
		SELECT id, username, password, email, role, level, experience,
		       is_active, is_online, is_banned, ban_reason, ban_expires_at,
		       gold, gems, alliance_id, world_id, race_id,
		       last_login, last_active, created_at, updated_at
		FROM players
		WHERE username = $1
	`, username).Scan(
		&player.ID,
		&player.Username,
		&player.Password,
		&player.Email,
		&player.Role,
		&player.Level,
		&player.Experience,
		&player.IsActive,
		&player.IsOnline,
		&player.IsBanned,
		&player.BanReason,
		&player.BanExpiresAt,
		&player.Gold,
		&player.Gems,
		&player.AllianceID,
		&player.WorldID,
		&player.RaceID,
		&player.LastLogin,
		&player.LastActive,
		&player.CreatedAt,
		&player.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &player, nil
}

func (r *PlayerRepository) GetPlayerByID(id uuid.UUID) (*models.Player, error) {
	var player models.Player
	err := r.db.QueryRow(`
		SELECT id, username, password, email, role, level, experience,
		       is_active, is_online, is_banned, ban_reason, ban_expires_at,
		       gold, gems, alliance_id, world_id, race_id,
		       last_login, last_active, created_at, updated_at
		FROM players
		WHERE id = $1
	`, id).Scan(
		&player.ID,
		&player.Username,
		&player.Password,
		&player.Email,
		&player.Role,
		&player.Level,
		&player.Experience,
		&player.IsActive,
		&player.IsOnline,
		&player.IsBanned,
		&player.BanReason,
		&player.BanExpiresAt,
		&player.Gold,
		&player.Gems,
		&player.AllianceID,
		&player.WorldID,
		&player.RaceID,
		&player.LastLogin,
		&player.LastActive,
		&player.CreatedAt,
		&player.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &player, nil
}

func (r *PlayerRepository) CreatePlayer(username, password, email string) (uuid.UUID, error) {
	id := uuid.New()
	now := time.Now()
	_, err := r.db.Exec(`
		INSERT INTO players (id, username, password, email, role, level, experience,
		                    is_active, is_online, is_banned, gold, gems,
		                    last_login, last_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`, id, username, password, email, "user", 1, 0, true, false, false, 0, 0, now, now, now, now)
	if err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

func (r *PlayerRepository) UpdatePlayer(id uuid.UUID, username, email string) error {
	_, err := r.db.Exec(`
		UPDATE players
		SET username = $1, email = $2
		WHERE id = $3
	`, username, email, id)
	return err
}

func (r *PlayerRepository) UpdateLastLogin(id uuid.UUID) error {
	_, err := r.db.Exec(`
		UPDATE players
		SET last_login = $1
		WHERE id = $2
	`, time.Now(), id)
	return err
}

func (r *PlayerRepository) UsernameExists(username string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM players WHERE username = $1)`
	err := r.db.QueryRow(query, username).Scan(&exists)
	return exists, err
}

func (r *PlayerRepository) Update(player *models.Player) error {
	query := `
		UPDATE players
		SET username = $1, password = $2, email = $3, last_login = $4
		WHERE id = $5
	`
	_, err := r.db.Exec(query,
		player.Username,
		player.Password,
		player.Email,
		player.LastLogin,
		player.ID,
	)
	return err
}

func (r *PlayerRepository) Delete(id string) error {
	query := `DELETE FROM players WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

// GetAllPlayers obtiene todos los jugadores
func (r *PlayerRepository) GetAllPlayers() ([]*models.Player, error) {
	query := `
		SELECT id, username, email, created_at, last_login
		FROM players
		ORDER BY username
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error consultando jugadores: %w", err)
	}
	defer rows.Close()

	var players []*models.Player
	for rows.Next() {
		var player models.Player
		err := rows.Scan(
			&player.ID,
			&player.Username,
			&player.Email,
			&player.CreatedAt,
			&player.LastLogin,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando jugador: %w", err)
		}
		players = append(players, &player)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterando jugadores: %w", err)
	}

	return players, nil
}

// GetTotalPlayers obtiene el total de jugadores registrados
func (r *PlayerRepository) GetTotalPlayers() (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM players`
	err := r.db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("error contando jugadores: %w", err)
	}
	return count, nil
}

// GetActivePlayers obtiene el número de jugadores activos (que han iniciado sesión en las últimas 24 horas)
func (r *PlayerRepository) GetActivePlayers() (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM players WHERE last_login > NOW() - INTERVAL '24 hours'`
	err := r.db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("error contando jugadores activos: %w", err)
	}
	return count, nil
}

// GetAllPlayersForAdmin obtiene todos los jugadores con datos completos para el dashboard de administración
func (r *PlayerRepository) GetAllPlayersForAdmin() ([]*models.Player, error) {
	query := `
		SELECT 
			p.id, p.username, p.email, p.role, p.level, p.experience,
			p.is_active, p.is_online, p.is_banned, p.ban_reason, p.ban_expires_at,
			p.gold, p.gems, p.alliance_id, p.world_id, p.race_id,
			p.last_login, p.last_active, p.created_at, p.updated_at
		FROM players p
		ORDER BY p.created_at DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error consultando jugadores: %w", err)
	}
	defer rows.Close()

	var players []*models.Player
	for rows.Next() {
		var player models.Player
		err := rows.Scan(
			&player.ID,
			&player.Username,
			&player.Email,
			&player.Role,
			&player.Level,
			&player.Experience,
			&player.IsActive,
			&player.IsOnline,
			&player.IsBanned,
			&player.BanReason,
			&player.BanExpiresAt,
			&player.Gold,
			&player.Gems,
			&player.AllianceID,
			&player.WorldID,
			&player.RaceID,
			&player.LastLogin,
			&player.LastActive,
			&player.CreatedAt,
			&player.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando jugador: %w", err)
		}
		players = append(players, &player)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterando jugadores: %w", err)
	}

	return players, nil
}

// GetPlayerVillages obtiene las aldeas de un jugador
func (r *PlayerRepository) GetPlayerVillages(playerID uuid.UUID) ([]models.Village, error) {
	query := `
		SELECT id, player_id, world_id, name, x_coordinate, y_coordinate, created_at
		FROM villages 
		WHERE player_id = $1
		ORDER BY created_at
	`

	rows, err := r.db.Query(query, playerID)
	if err != nil {
		return nil, fmt.Errorf("error consultando aldeas: %w", err)
	}
	defer rows.Close()

	var villages []models.Village
	for rows.Next() {
		var village models.Village
		err := rows.Scan(
			&village.ID,
			&village.PlayerID,
			&village.WorldID,
			&village.Name,
			&village.XCoordinate,
			&village.YCoordinate,
			&village.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando aldea: %w", err)
		}
		villages = append(villages, village)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterando aldeas: %w", err)
	}

	return villages, nil
}

// GetPlayerAchievements obtiene los logros de un jugador
func (r *PlayerRepository) GetPlayerAchievements(playerID uuid.UUID) ([]models.SimpleAchievement, error) {
	// Por ahora, devolvemos un array vacío ya que la tabla player_achievements puede no existir
	// TODO: Implementar cuando se tenga la tabla player_achievements configurada
	return []models.SimpleAchievement{}, nil
}

// GetPlayerTitles obtiene los títulos de un jugador
func (r *PlayerRepository) GetPlayerTitles(playerID uuid.UUID) ([]models.Title, error) {
	// Por ahora, devolvemos un array vacío ya que la tabla player_titles puede no existir
	// TODO: Implementar cuando se tenga la tabla player_titles configurada
	return []models.Title{}, nil
}
