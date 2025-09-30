package repository

import (
	"database/sql"
	"fmt"
	"time"

	"server-backend/models"

	"go.uber.org/zap"
)

type AllianceRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewAllianceRepository(db *sql.DB, logger *zap.Logger) *AllianceRepository {
	return &AllianceRepository{
		db:     db,
		logger: logger,
	}
}

// CreateAlliance crea una nueva alianza
func (r *AllianceRepository) CreateAlliance(alliance *models.Alliance) (*models.Alliance, error) {
	query := `
		INSERT INTO alliances (name, description, tag, leader_id, level, experience, max_members, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at
	`

	now := time.Now()
	alliance.CreatedAt = now
	alliance.UpdatedAt = now
	alliance.Level = 1
	alliance.Experience = 0
	alliance.MaxMembers = 20

	err := r.db.QueryRow(
		query,
		alliance.Name,
		alliance.Description,
		alliance.Tag,
		alliance.LeaderID,
		alliance.Level,
		alliance.Experience,
		alliance.MaxMembers,
		alliance.CreatedAt,
		alliance.UpdatedAt,
	).Scan(&alliance.ID, &alliance.CreatedAt, &alliance.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("error creando alianza: %w", err)
	}

	return alliance, nil
}

// GetAlliances obtiene todas las alianzas
func (r *AllianceRepository) GetAlliances() ([]models.Alliance, error) {
	query := `
		SELECT id, name, description, tag, leader_id, level, experience, max_members, created_at, updated_at
		FROM alliances
		ORDER BY level DESC, experience DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo alianzas: %w", err)
	}
	defer rows.Close()

	var alliances []models.Alliance
	for rows.Next() {
		var alliance models.Alliance
		err := rows.Scan(
			&alliance.ID,
			&alliance.Name,
			&alliance.Description,
			&alliance.Tag,
			&alliance.LeaderID,
			&alliance.Level,
			&alliance.Experience,
			&alliance.MaxMembers,
			&alliance.CreatedAt,
			&alliance.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando alianza: %w", err)
		}
		alliances = append(alliances, alliance)
	}

	return alliances, nil
}

// GetAlliance obtiene una alianza específica
func (r *AllianceRepository) GetAlliance(allianceID int) (*models.Alliance, error) {
	query := `
		SELECT id, name, description, tag, leader_id, level, experience, max_members, created_at, updated_at
		FROM alliances
		WHERE id = $1
	`

	var alliance models.Alliance
	err := r.db.QueryRow(query, allianceID).Scan(
		&alliance.ID,
		&alliance.Name,
		&alliance.Description,
		&alliance.Tag,
		&alliance.LeaderID,
		&alliance.Level,
		&alliance.Experience,
		&alliance.MaxMembers,
		&alliance.CreatedAt,
		&alliance.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("alianza no encontrada")
		}
		return nil, fmt.Errorf("error obteniendo alianza: %w", err)
	}

	return &alliance, nil
}

// UpdateAlliance actualiza una alianza
func (r *AllianceRepository) UpdateAlliance(alliance *models.Alliance) (*models.Alliance, error) {
	query := `
		UPDATE alliances
		SET name = $1, description = $2, tag = $3, updated_at = $4
		WHERE id = $5
		RETURNING id, name, description, tag, leader_id, level, experience, max_members, created_at, updated_at
	`

	alliance.UpdatedAt = time.Now()

	var updatedAlliance models.Alliance
	err := r.db.QueryRow(
		query,
		alliance.Name,
		alliance.Description,
		alliance.Tag,
		alliance.UpdatedAt,
		alliance.ID,
	).Scan(
		&updatedAlliance.ID,
		&updatedAlliance.Name,
		&updatedAlliance.Description,
		&updatedAlliance.Tag,
		&updatedAlliance.LeaderID,
		&updatedAlliance.Level,
		&updatedAlliance.Experience,
		&updatedAlliance.MaxMembers,
		&updatedAlliance.CreatedAt,
		&updatedAlliance.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("error actualizando alianza: %w", err)
	}

	return &updatedAlliance, nil
}

// DeleteAlliance elimina una alianza
func (r *AllianceRepository) DeleteAlliance(allianceID int) error {
	// Primero eliminar todos los miembros
	_, err := r.db.Exec("DELETE FROM alliance_members WHERE alliance_id = $1", allianceID)
	if err != nil {
		return fmt.Errorf("error eliminando miembros de la alianza: %w", err)
	}

	// Luego eliminar la alianza
	_, err = r.db.Exec("DELETE FROM alliances WHERE id = $1", allianceID)
	if err != nil {
		return fmt.Errorf("error eliminando alianza: %w", err)
	}

	return nil
}

// AddMember agrega un miembro a una alianza
func (r *AllianceRepository) AddMember(member *models.AllianceMember) error {
	query := `
		INSERT INTO alliance_members (alliance_id, player_id, role, joined_at, contribution)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	member.JoinedAt = time.Now()
	member.Contribution = 0

	err := r.db.QueryRow(
		query,
		member.AllianceID,
		member.PlayerID,
		member.Role,
		member.JoinedAt,
		member.Contribution,
	).Scan(&member.ID)

	if err != nil {
		return fmt.Errorf("error agregando miembro: %w", err)
	}

	return nil
}

// RemoveMember elimina un miembro de una alianza
func (r *AllianceRepository) RemoveMember(allianceID, playerID int) error {
	_, err := r.db.Exec(
		"DELETE FROM alliance_members WHERE alliance_id = $1 AND player_id = $2",
		allianceID, playerID,
	)
	if err != nil {
		return fmt.Errorf("error eliminando miembro: %w", err)
	}

	return nil
}

// GetAllianceMembers obtiene todos los miembros de una alianza
func (r *AllianceRepository) GetAllianceMembers(allianceID int) ([]models.AllianceMember, error) {
	query := `
		SELECT id, alliance_id, player_id, role, joined_at, contribution
		FROM alliance_members
		WHERE alliance_id = $1
		ORDER BY 
			CASE role 
				WHEN 'leader' THEN 1 
				WHEN 'officer' THEN 2 
				ELSE 3 
			END,
			joined_at ASC
	`

	rows, err := r.db.Query(query, allianceID)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo miembros: %w", err)
	}
	defer rows.Close()

	var members []models.AllianceMember
	for rows.Next() {
		var member models.AllianceMember
		err := rows.Scan(
			&member.ID,
			&member.AllianceID,
			&member.PlayerID,
			&member.Role,
			&member.JoinedAt,
			&member.Contribution,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando miembro: %w", err)
		}
		members = append(members, member)
	}

	return members, nil
}

// IsPlayerMember verifica si un jugador es miembro de una alianza
func (r *AllianceRepository) IsPlayerMember(allianceID, playerID int) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM alliance_members WHERE alliance_id = $1 AND player_id = $2)"

	err := r.db.QueryRow(query, allianceID, playerID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("error verificando membresía: %w", err)
	}

	return exists, nil
}

// IsPlayerLeader verifica si un jugador es líder de una alianza
func (r *AllianceRepository) IsPlayerLeader(allianceID, playerID int) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM alliance_members WHERE alliance_id = $1 AND player_id = $2 AND role = 'leader')"

	err := r.db.QueryRow(query, allianceID, playerID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("error verificando liderazgo: %w", err)
	}

	return exists, nil
}

// GetPlayerRole obtiene el rol de un jugador en una alianza
func (r *AllianceRepository) GetPlayerRole(allianceID, playerID int) (string, error) {
	var role string
	query := "SELECT role FROM alliance_members WHERE alliance_id = $1 AND player_id = $2"

	err := r.db.QueryRow(query, allianceID, playerID).Scan(&role)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("jugador no es miembro de la alianza")
		}
		return "", fmt.Errorf("error obteniendo rol: %w", err)
	}

	return role, nil
}

// GetPlayerAlliance obtiene la alianza de un jugador
func (r *AllianceRepository) GetPlayerAlliance(playerID int) (*models.Alliance, error) {
	query := `
		SELECT a.id, a.name, a.description, a.tag, a.leader_id, a.level, a.experience, a.max_members, a.created_at, a.updated_at
		FROM alliances a
		JOIN alliance_members am ON a.id = am.alliance_id
		WHERE am.player_id = $1
	`

	var alliance models.Alliance
	err := r.db.QueryRow(query, playerID).Scan(
		&alliance.ID,
		&alliance.Name,
		&alliance.Description,
		&alliance.Tag,
		&alliance.LeaderID,
		&alliance.Level,
		&alliance.Experience,
		&alliance.MaxMembers,
		&alliance.CreatedAt,
		&alliance.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No está en ninguna alianza
		}
		return nil, fmt.Errorf("error obteniendo alianza del jugador: %w", err)
	}

	return &alliance, nil
}

// PromoteMember promueve a un miembro a oficial
func (r *AllianceRepository) PromoteMember(allianceID, playerID int) error {
	_, err := r.db.Exec(
		"UPDATE alliance_members SET role = 'officer' WHERE alliance_id = $1 AND player_id = $2",
		allianceID, playerID,
	)
	if err != nil {
		return fmt.Errorf("error promoviendo miembro: %w", err)
	}

	return nil
}

// DemoteMember degrada a un oficial a miembro
func (r *AllianceRepository) DemoteMember(allianceID, playerID int) error {
	_, err := r.db.Exec(
		"UPDATE alliance_members SET role = 'member' WHERE alliance_id = $1 AND player_id = $2",
		allianceID, playerID,
	)
	if err != nil {
		return fmt.Errorf("error degradando miembro: %w", err)
	}

	return nil
}

// GetMemberCount obtiene el número de miembros de una alianza
func (r *AllianceRepository) GetMemberCount(allianceID int) (int, error) {
	var count int
	query := "SELECT COUNT(*) FROM alliance_members WHERE alliance_id = $1"

	err := r.db.QueryRow(query, allianceID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("error contando miembros: %w", err)
	}

	return count, nil
}

// AddExperience agrega experiencia a una alianza
func (r *AllianceRepository) AddExperience(allianceID, experience int) error {
	_, err := r.db.Exec(
		"UPDATE alliances SET experience = experience + $1, updated_at = $2 WHERE id = $3",
		experience, time.Now(), allianceID,
	)
	if err != nil {
		return fmt.Errorf("error agregando experiencia: %w", err)
	}

	return nil
}

// GetAllianceRankings obtiene el ranking de alianzas
func (r *AllianceRepository) GetAllianceRankings(limit int) ([]models.AllianceRanking, error) {
	query := `
		WITH alliance_stats AS (
			SELECT 
				a.id as alliance_id,
				a.name as alliance_name,
				a.tag as alliance_tag,
				COUNT(am.player_id) as member_count,
				COALESCE(SUM(p.total_power), 0) as total_power,
				COALESCE(SUM(p.total_score), 0) as total_score
			FROM alliances a
			LEFT JOIN alliance_members am ON a.id = am.alliance_id
			LEFT JOIN players p ON am.player_id = p.id
			GROUP BY a.id, a.name, a.tag
		)
		SELECT 
			alliance_id,
			alliance_name,
			alliance_tag,
			total_power,
			total_score,
			member_count,
			RANK() OVER (ORDER BY total_power DESC, total_score DESC) as rank
		FROM alliance_stats
		ORDER BY rank
		LIMIT $1
	`

	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo rankings: %w", err)
	}
	defer rows.Close()

	var rankings []models.AllianceRanking
	for rows.Next() {
		var ranking models.AllianceRanking
		err := rows.Scan(
			&ranking.AllianceID,
			&ranking.AllianceName,
			&ranking.AllianceTag,
			&ranking.TotalPower,
			&ranking.TotalScore,
			&ranking.MemberCount,
			&ranking.Rank,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando ranking: %w", err)
		}
		rankings = append(rankings, ranking)
	}

	return rankings, nil
}
