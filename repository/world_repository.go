package repository

import (
	"database/sql"
	"fmt"
	"server-backend/models"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type WorldRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewWorldRepository(db *sql.DB, logger *zap.Logger) *WorldRepository {
	return &WorldRepository{
		db:     db,
		logger: logger,
	}
}

func (r *WorldRepository) GetWorldByID(id uuid.UUID) (*models.World, error) {
	var world models.World
	err := r.db.QueryRow(`
		SELECT id, name, description, max_players, current_players, is_active, is_online, world_type, status, last_started_at, last_stopped_at, created_at, updated_at
		FROM worlds
		WHERE id = $1
	`, id).Scan(
		&world.ID,
		&world.Name,
		&world.Description,
		&world.MaxPlayers,
		&world.CurrentPlayers,
		&world.IsActive,
		&world.IsOnline,
		&world.WorldType,
		&world.Status,
		&world.LastStartedAt,
		&world.LastStoppedAt,
		&world.CreatedAt,
		&world.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &world, nil
}

func (r *WorldRepository) GetWorlds() ([]*models.World, error) {
	rows, err := r.db.Query(`
		SELECT id, name, max_players, is_active, created_at
		FROM worlds
		WHERE is_active = true
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var worlds []*models.World
	for rows.Next() {
		var world models.World
		err := rows.Scan(
			&world.ID,
			&world.Name,
			&world.MaxPlayers,
			&world.IsActive,
			&world.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		worlds = append(worlds, &world)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return worlds, nil
}

func (r *WorldRepository) CreateWorld(name, description, worldType string, maxPlayers int) (*models.World, error) {
	id := uuid.New()
	now := time.Now()

	world := &models.World{
		ID:             id,
		Name:           name,
		Description:    description,
		MaxPlayers:     maxPlayers,
		CurrentPlayers: 0,
		IsActive:       false, // Por defecto inactivo
		IsOnline:       false, // Por defecto offline
		WorldType:      worldType,
		Status:         "offline", // Por defecto offline
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	_, err := r.db.Exec(`
		INSERT INTO worlds (id, name, description, max_players, current_players, is_active, is_online, world_type, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`, world.ID, world.Name, world.Description, world.MaxPlayers, world.CurrentPlayers, world.IsActive, world.IsOnline, world.WorldType, world.Status, world.CreatedAt, world.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return world, nil
}

func (r *WorldRepository) GetWorldPlayerCount(worldID uuid.UUID) (int, error) {
	var count int
	err := r.db.QueryRow(`
		SELECT COUNT(DISTINCT player_id)
		FROM villages
		WHERE world_id = $1
	`, worldID).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// GetTotalWorlds obtiene el total de mundos registrados
func (r *WorldRepository) GetTotalWorlds() (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM worlds`
	err := r.db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// GetActiveWorlds obtiene el número de mundos activos
func (r *WorldRepository) GetActiveWorlds() (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM worlds WHERE is_active = true`
	err := r.db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// GetAllWorlds obtiene todos los mundos (activos e inactivos)
func (r *WorldRepository) GetAllWorlds() ([]*models.World, error) {
	rows, err := r.db.Query(`
		SELECT id, name, description, max_players, current_players, is_active, is_online, world_type, status, last_started_at, last_stopped_at, created_at, updated_at
		FROM worlds
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var worlds []*models.World
	for rows.Next() {
		var world models.World
		err := rows.Scan(
			&world.ID,
			&world.Name,
			&world.Description,
			&world.MaxPlayers,
			&world.CurrentPlayers,
			&world.IsActive,
			&world.IsOnline,
			&world.WorldType,
			&world.Status,
			&world.LastStartedAt,
			&world.LastStoppedAt,
			&world.CreatedAt,
			&world.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		worlds = append(worlds, &world)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return worlds, nil
}

// UpdateWorld actualiza un mundo existente
func (r *WorldRepository) UpdateWorld(id uuid.UUID, name, description, worldType string, maxPlayers int) (*models.World, error) {
	// Verificar que el mundo existe
	existingWorld, err := r.GetWorldByID(id)
	if err != nil {
		return nil, err
	}
	if existingWorld == nil {
		return nil, sql.ErrNoRows
	}

	now := time.Now()
	_, err = r.db.Exec(`
		UPDATE worlds 
		SET name = $1, description = $2, max_players = $3, world_type = $4, updated_at = $5
		WHERE id = $6
	`, name, description, maxPlayers, worldType, now, id)

	if err != nil {
		return nil, err
	}

	// Retornar el mundo actualizado
	updatedWorld := &models.World{
		ID:             id,
		Name:           name,
		Description:    description,
		MaxPlayers:     maxPlayers,
		CurrentPlayers: existingWorld.CurrentPlayers, // Mantener el conteo actual
		IsActive:       existingWorld.IsActive,       // Mantener el estado actual
		WorldType:      worldType,
		CreatedAt:      existingWorld.CreatedAt,
		UpdatedAt:      now,
	}

	return updatedWorld, nil
}

// DeleteWorld elimina un mundo
func (r *WorldRepository) DeleteWorld(id uuid.UUID) error {
	// Verificar que el mundo existe
	world, err := r.GetWorldByID(id)
	if err != nil {
		return err
	}
	if world == nil {
		return sql.ErrNoRows
	}

	// Verificar que no hay jugadores activos
	playerCount, err := r.GetWorldPlayerCount(id)
	if err != nil {
		return err
	}
	if playerCount > 0 {
		return fmt.Errorf("no se puede eliminar un mundo con jugadores activos (%d jugadores)", playerCount)
	}

	// Eliminar el mundo
	_, err = r.db.Exec(`DELETE FROM worlds WHERE id = $1`, id)
	return err
}

// StartWorld activa un mundo (inicia la instancia)
func (r *WorldRepository) StartWorld(id uuid.UUID) error {
	// Verificar que el mundo existe
	world, err := r.GetWorldByID(id)
	if err != nil {
		return err
	}
	if world == nil {
		return sql.ErrNoRows
	}

	// Verificar que no esté ya online
	if world.IsOnline {
		return fmt.Errorf("el mundo ya está online")
	}

	now := time.Now()
	// Iniciar el mundo
	_, err = r.db.Exec(`
		UPDATE worlds 
		SET is_active = true, is_online = true, status = 'online', last_started_at = $1, updated_at = $2
		WHERE id = $3
	`, now, now, id)

	return err
}

// StopWorld desactiva un mundo (detiene la instancia)
func (r *WorldRepository) StopWorld(id uuid.UUID) error {
	// Verificar que el mundo existe
	world, err := r.GetWorldByID(id)
	if err != nil {
		return err
	}
	if world == nil {
		return sql.ErrNoRows
	}

	// Verificar que esté online
	if !world.IsOnline {
		return fmt.Errorf("el mundo ya está offline")
	}

	now := time.Now()
	// Detener el mundo
	_, err = r.db.Exec(`
		UPDATE worlds 
		SET is_active = false, is_online = false, status = 'offline', last_stopped_at = $1, updated_at = $2
		WHERE id = $3
	`, now, now, id)

	return err
}

// UpdateWorldPlayerCount actualiza el conteo de jugadores de un mundo
func (r *WorldRepository) UpdateWorldPlayerCount(worldID uuid.UUID) error {
	_, err := r.db.Exec(`
		UPDATE worlds 
		SET current_players = (
			SELECT COUNT(*) 
			FROM player_world_entries 
			WHERE world_id = $1 AND is_active = true
		), updated_at = NOW()
		WHERE id = $1
	`, worldID)

	return err
}

// PlayerWorldEntry methods

// AddPlayerToWorld agrega un jugador a un mundo específico
func (r *WorldRepository) AddPlayerToWorld(playerID, worldID uuid.UUID) error {
	// Verificar que el mundo existe y está online
	world, err := r.GetWorldByID(worldID)
	if err != nil {
		return err
	}
	if world == nil {
		return fmt.Errorf("mundo no encontrado")
	}
	if !world.IsOnline {
		return fmt.Errorf("el mundo no está online")
	}

	// Verificar que no esté lleno
	if world.CurrentPlayers >= world.MaxPlayers {
		return fmt.Errorf("el mundo está lleno")
	}

	// Iniciar transacción
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Desactivar entrada previa del jugador en otros mundos
	_, err = tx.Exec(`
		UPDATE player_world_entries 
		SET is_active = false, updated_at = NOW()
		WHERE player_id = $1 AND is_active = true
	`, playerID)
	if err != nil {
		return err
	}

	// Agregar entrada al nuevo mundo
	_, err = tx.Exec(`
		INSERT INTO player_world_entries (player_id, world_id, entered_at, is_active, last_seen)
		VALUES ($1, $2, NOW(), true, NOW())
		ON CONFLICT (player_id, world_id) 
		DO UPDATE SET is_active = true, last_seen = NOW(), updated_at = NOW()
	`, playerID, worldID)
	if err != nil {
		return err
	}

	// Actualizar conteo de jugadores del mundo
	_, err = tx.Exec(`
		UPDATE worlds 
		SET current_players = (
			SELECT COUNT(*) 
			FROM player_world_entries 
			WHERE world_id = $1 AND is_active = true
		), updated_at = NOW()
		WHERE id = $1
	`, worldID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// RemovePlayerFromWorld remueve un jugador de un mundo
func (r *WorldRepository) RemovePlayerFromWorld(playerID, worldID uuid.UUID) error {
	// Iniciar transacción
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Desactivar entrada del jugador
	_, err = tx.Exec(`
		UPDATE player_world_entries 
		SET is_active = false, updated_at = NOW()
		WHERE player_id = $1 AND world_id = $2
	`, playerID, worldID)
	if err != nil {
		return err
	}

	// Actualizar conteo de jugadores del mundo
	_, err = tx.Exec(`
		UPDATE worlds 
		SET current_players = (
			SELECT COUNT(*) 
			FROM player_world_entries 
			WHERE world_id = $1 AND is_active = true
		), updated_at = NOW()
		WHERE id = $1
	`, worldID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// GetWorldPlayers obtiene la lista de jugadores en un mundo específico
func (r *WorldRepository) GetWorldPlayers(worldID uuid.UUID) ([]models.WorldPlayerInfo, error) {
	rows, err := r.db.Query(`
		SELECT 
			p.id as player_id,
			p.username,
			p.level,
			pwe.entered_at,
			pwe.last_seen,
			pwe.is_active,
			(SELECT COUNT(*) FROM villages v WHERE v.player_id = p.id AND v.world_id = $1) as village_count,
			p.alliance_id
		FROM player_world_entries pwe
		JOIN players p ON p.id = pwe.player_id
		WHERE pwe.world_id = $1
		ORDER BY pwe.entered_at DESC
	`, worldID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var players []models.WorldPlayerInfo
	for rows.Next() {
		var player models.WorldPlayerInfo
		var allianceID sql.NullString
		err := rows.Scan(
			&player.PlayerID,
			&player.Username,
			&player.Level,
			&player.EnteredAt,
			&player.LastSeen,
			&player.IsActive,
			&player.VillageCount,
			&allianceID,
		)
		if err != nil {
			return nil, err
		}

		if allianceID.Valid {
			if id, err := uuid.Parse(allianceID.String); err == nil {
				player.AllianceID = &id
			}
		}

		players = append(players, player)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return players, nil
}

// GetPlayerCurrentWorld obtiene el mundo actual de un jugador
func (r *WorldRepository) GetPlayerCurrentWorld(playerID uuid.UUID) (*models.World, error) {
	var world models.World
	err := r.db.QueryRow(`
		SELECT w.id, w.name, w.description, w.max_players, w.current_players, w.is_active, w.is_online, w.world_type, w.status, w.last_started_at, w.last_stopped_at, w.created_at, w.updated_at
		FROM worlds w
		JOIN player_world_entries pwe ON w.id = pwe.world_id
		WHERE pwe.player_id = $1 AND pwe.is_active = true
	`, playerID).Scan(
		&world.ID,
		&world.Name,
		&world.Description,
		&world.MaxPlayers,
		&world.CurrentPlayers,
		&world.IsActive,
		&world.IsOnline,
		&world.WorldType,
		&world.Status,
		&world.LastStartedAt,
		&world.LastStoppedAt,
		&world.CreatedAt,
		&world.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &world, nil
}

// GetWorldStatus obtiene el estado completo de un mundo
func (r *WorldRepository) GetWorldStatus(worldID uuid.UUID) (*models.WorldStatus, error) {
	world, err := r.GetWorldByID(worldID)
	if err != nil {
		return nil, err
	}
	if world == nil {
		return nil, sql.ErrNoRows
	}

	players, err := r.GetWorldPlayers(worldID)
	if err != nil {
		return nil, err
	}

	// Calcular uptime si está online
	var uptime string
	if world.IsOnline && world.LastStartedAt != nil {
		duration := time.Since(*world.LastStartedAt)
		uptime = duration.String()
	}

	status := &models.WorldStatus{
		WorldID:        world.ID,
		Name:           world.Name,
		Status:         world.Status,
		IsOnline:       world.IsOnline,
		CurrentPlayers: world.CurrentPlayers,
		MaxPlayers:     world.MaxPlayers,
		LastStartedAt:  world.LastStartedAt,
		LastStoppedAt:  world.LastStoppedAt,
		Uptime:         uptime,
		PlayerList:     players,
	}

	return status, nil
}
