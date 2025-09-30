package repository

import (
	"database/sql"
	"server-backend/models"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type UnitRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewUnitRepository(db *sql.DB, logger *zap.Logger) *UnitRepository {
	return &UnitRepository{
		db:     db,
		logger: logger,
	}
}

func (r *UnitRepository) GetUnitsByVillageID(villageID uuid.UUID) ([]*models.Unit, error) {
	rows, err := r.db.Query(`
		SELECT id, village_id, type, quantity, in_training, training_completion_time, created_at, updated_at
		FROM units
		WHERE village_id = $1
		ORDER BY type
	`, villageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var units []*models.Unit
	for rows.Next() {
		var unit models.Unit
		err := rows.Scan(
			&unit.ID,
			&unit.VillageID,
			&unit.Type,
			&unit.Quantity,
			&unit.InTraining,
			&unit.TrainingCompletionTime,
			&unit.CreatedAt,
			&unit.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		units = append(units, &unit)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return units, nil
}

func (r *UnitRepository) GetUnitByVillageAndType(villageID uuid.UUID, unitType string) (*models.Unit, error) {
	var unit models.Unit
	err := r.db.QueryRow(`
		SELECT id, village_id, type, quantity, in_training, training_completion_time, created_at, updated_at
		FROM units
		WHERE village_id = $1 AND type = $2
	`, villageID, unitType).Scan(
		&unit.ID,
		&unit.VillageID,
		&unit.Type,
		&unit.Quantity,
		&unit.InTraining,
		&unit.TrainingCompletionTime,
		&unit.CreatedAt,
		&unit.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &unit, nil
}

func (r *UnitRepository) CreateUnit(villageID uuid.UUID, unitType string) (*models.Unit, error) {
	unit := &models.Unit{
		ID:         uuid.New(),
		VillageID:  villageID,
		Type:       unitType,
		Quantity:   0,
		InTraining: 0,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	_, err := r.db.Exec(`
		INSERT INTO units (id, village_id, type, quantity, in_training, training_completion_time, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, unit.ID, unit.VillageID, unit.Type, unit.Quantity, unit.InTraining, unit.TrainingCompletionTime, unit.CreatedAt, unit.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return unit, nil
}

func (r *UnitRepository) UpdateUnit(unit *models.Unit) error {
	unit.UpdatedAt = time.Now()
	_, err := r.db.Exec(`
		UPDATE units
		SET quantity = $1, in_training = $2, training_completion_time = $3, updated_at = $4
		WHERE id = $5
	`, unit.Quantity, unit.InTraining, unit.TrainingCompletionTime, unit.UpdatedAt, unit.ID)
	return err
}

func (r *UnitRepository) StartTraining(villageID uuid.UUID, unitType string, quantity int) error {
	// Obtener o crear la unidad
	unit, err := r.GetUnitByVillageAndType(villageID, unitType)
	if err != nil {
		return err
	}

	if unit == nil {
		unit, err = r.CreateUnit(villageID, unitType)
		if err != nil {
			return err
		}
	}

	// Verificar que el tipo de unidad existe
	unitTypeInfo, exists := models.UnitTypes[unitType]
	if !exists {
		return sql.ErrNoRows
	}

	// Calcular tiempo de entrenamiento
	trainingTime := time.Duration(unitTypeInfo.TrainingTime) * time.Second
	completionTime := time.Now().Add(trainingTime)

	// Actualizar unidad
	unit.InTraining += quantity
	unit.TrainingCompletionTime = &completionTime

	return r.UpdateUnit(unit)
}

func (r *UnitRepository) CompleteTraining(villageID uuid.UUID, unitType string) error {
	unit, err := r.GetUnitByVillageAndType(villageID, unitType)
	if err != nil {
		return err
	}
	if unit == nil {
		return sql.ErrNoRows
	}

	// Verificar si hay entrenamiento completado
	if unit.TrainingCompletionTime != nil && time.Now().After(*unit.TrainingCompletionTime) {
		unit.Quantity += unit.InTraining
		unit.InTraining = 0
		unit.TrainingCompletionTime = nil
		return r.UpdateUnit(unit)
	}

	return nil
}

func (r *UnitRepository) GetUnitsInTraining(villageID uuid.UUID) ([]*models.Unit, error) {
	rows, err := r.db.Query(`
		SELECT id, village_id, type, quantity, in_training, training_completion_time, created_at, updated_at
		FROM units
		WHERE village_id = $1 AND in_training > 0
		ORDER BY training_completion_time ASC
	`, villageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var units []*models.Unit
	for rows.Next() {
		var unit models.Unit
		err := rows.Scan(
			&unit.ID,
			&unit.VillageID,
			&unit.Type,
			&unit.Quantity,
			&unit.InTraining,
			&unit.TrainingCompletionTime,
			&unit.CreatedAt,
			&unit.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		units = append(units, &unit)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return units, nil
}

func (r *UnitRepository) GetTotalUnitsByType(villageID uuid.UUID) (map[string]int, error) {
	rows, err := r.db.Query(`
		SELECT type, SUM(quantity) as total
		FROM units
		WHERE village_id = $1
		GROUP BY type
	`, villageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	totals := make(map[string]int)
	for rows.Next() {
		var unitType string
		var total int
		err := rows.Scan(&unitType, &total)
		if err != nil {
			return nil, err
		}
		totals[unitType] = total
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return totals, nil
}
