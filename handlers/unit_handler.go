package handlers

import (
	"encoding/json"
	"net/http"
	"server-backend/models"
	"server-backend/repository"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type UnitHandler struct {
	unitRepo    *repository.UnitRepository
	villageRepo *repository.VillageRepository
	logger      *zap.Logger
}

func NewUnitHandler(unitRepo *repository.UnitRepository, villageRepo *repository.VillageRepository, logger *zap.Logger) *UnitHandler {
	return &UnitHandler{
		unitRepo:    unitRepo,
		villageRepo: villageRepo,
		logger:      logger,
	}
}

type TrainUnitsRequest struct {
	UnitType string `json:"unit_type"`
	Quantity int    `json:"quantity"`
}

type UnitResponse struct {
	Type                   string  `json:"type"`
	Name                   string  `json:"name"`
	Description            string  `json:"description"`
	Quantity               int     `json:"quantity"`
	InTraining             int     `json:"in_training"`
	TrainingCompletionTime *string `json:"training_completion_time,omitempty"`
	Cost                   struct {
		Wood  int `json:"wood"`
		Stone int `json:"stone"`
		Food  int `json:"food"`
		Gold  int `json:"gold"`
	} `json:"cost"`
	TrainingTime int `json:"training_time"`
}

func (h *UnitHandler) GetUnits(w http.ResponseWriter, r *http.Request) {
	// Obtener el ID del jugador del contexto
	playerIDStr := r.Context().Value("player_id").(string)
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		h.logger.Error("Error parseando ID de jugador", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	// Obtener el ID de la aldea de los query params
	villageIDStr := r.URL.Query().Get("village_id")
	if villageIDStr == "" {
		http.Error(w, "ID de aldea requerido", http.StatusBadRequest)
		return
	}

	villageID, err := uuid.Parse(villageIDStr)
	if err != nil {
		http.Error(w, "ID de aldea inválido", http.StatusBadRequest)
		return
	}

	// Verificar que la aldea pertenece al jugador
	village, err := h.villageRepo.GetVillageByID(villageID)
	if err != nil {
		h.logger.Error("Error obteniendo aldea", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}
	if village == nil {
		http.Error(w, "Aldea no encontrada", http.StatusNotFound)
		return
	}
	if village.Village.PlayerID != playerID {
		http.Error(w, "No autorizado", http.StatusForbidden)
		return
	}

	// Obtener unidades de la aldea
	units, err := h.unitRepo.GetUnitsByVillageID(villageID)
	if err != nil {
		h.logger.Error("Error obteniendo unidades", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	// Crear respuesta con información completa de unidades
	var response []UnitResponse
	for _, unit := range units {
		unitType, exists := models.UnitTypes[unit.Type]
		if !exists {
			continue
		}

		unitResp := UnitResponse{
			Type:         unit.Type,
			Name:         unitType.Name,
			Description:  unitType.Description,
			Quantity:     unit.Quantity,
			InTraining:   unit.InTraining,
			Cost:         unitType.Cost,
			TrainingTime: unitType.TrainingTime,
		}

		if unit.TrainingCompletionTime != nil {
			timeStr := unit.TrainingCompletionTime.Format("2006-01-02T15:04:05Z")
			unitResp.TrainingCompletionTime = &timeStr
		}

		response = append(response, unitResp)
	}

	// Agregar tipos de unidades disponibles que no están en la aldea
	existingTypes := make(map[string]bool)
	for _, unit := range units {
		existingTypes[unit.Type] = true
	}

	for unitType, unitTypeInfo := range models.UnitTypes {
		if !existingTypes[unitType] {
			unitResp := UnitResponse{
				Type:         unitType,
				Name:         unitTypeInfo.Name,
				Description:  unitTypeInfo.Description,
				Quantity:     0,
				InTraining:   0,
				Cost:         unitTypeInfo.Cost,
				TrainingTime: unitTypeInfo.TrainingTime,
			}
			response = append(response, unitResp)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *UnitHandler) TrainUnits(w http.ResponseWriter, r *http.Request) {
	// Obtener el ID del jugador del contexto
	playerIDStr := r.Context().Value("player_id").(string)
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		h.logger.Error("Error parseando ID de jugador", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	// Obtener el ID de la aldea de los query params
	villageIDStr := r.URL.Query().Get("village_id")
	if villageIDStr == "" {
		http.Error(w, "ID de aldea requerido", http.StatusBadRequest)
		return
	}

	villageID, err := uuid.Parse(villageIDStr)
	if err != nil {
		http.Error(w, "ID de aldea inválido", http.StatusBadRequest)
		return
	}

	// Verificar que la aldea pertenece al jugador
	village, err := h.villageRepo.GetVillageByID(villageID)
	if err != nil {
		h.logger.Error("Error obteniendo aldea", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}
	if village == nil {
		http.Error(w, "Aldea no encontrada", http.StatusNotFound)
		return
	}
	if village.Village.PlayerID != playerID {
		http.Error(w, "No autorizado", http.StatusForbidden)
		return
	}

	// Decodificar la solicitud
	var req TrainUnitsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Error decodificando la solicitud", http.StatusBadRequest)
		return
	}

	// Validar el tipo de unidad
	unitType, exists := models.UnitTypes[req.UnitType]
	if !exists {
		http.Error(w, "Tipo de unidad inválido", http.StatusBadRequest)
		return
	}

	// Validar cantidad
	if req.Quantity <= 0 {
		http.Error(w, "Cantidad debe ser mayor a 0", http.StatusBadRequest)
		return
	}

	// Verificar que hay suficientes recursos
	totalCost := struct {
		Wood  int
		Stone int
		Food  int
		Gold  int
	}{
		Wood:  unitType.Cost.Wood * req.Quantity,
		Stone: unitType.Cost.Stone * req.Quantity,
		Food:  unitType.Cost.Food * req.Quantity,
		Gold:  unitType.Cost.Gold * req.Quantity,
	}

	if village.Resources.Wood < totalCost.Wood ||
		village.Resources.Stone < totalCost.Stone ||
		village.Resources.Food < totalCost.Food ||
		village.Resources.Gold < totalCost.Gold {
		http.Error(w, "Recursos insuficientes", http.StatusBadRequest)
		return
	}

	// Verificar que hay cuartel para entrenar unidades
	barracks, exists := village.Buildings["barracks"]
	if !exists || barracks.Level < 1 {
		http.Error(w, "Se requiere cuartel nivel 1 para entrenar unidades", http.StatusBadRequest)
		return
	}

	// Iniciar entrenamiento
	err = h.unitRepo.StartTraining(villageID, req.UnitType, req.Quantity)
	if err != nil {
		h.logger.Error("Error iniciando entrenamiento", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	// Descontar recursos
	newWood := village.Resources.Wood - totalCost.Wood
	newStone := village.Resources.Stone - totalCost.Stone
	newFood := village.Resources.Food - totalCost.Food
	newGold := village.Resources.Gold - totalCost.Gold

	err = h.villageRepo.UpdateResources(villageID, newWood, newStone, newFood, newGold)
	if err != nil {
		h.logger.Error("Error actualizando recursos", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	// Respuesta exitosa
	response := map[string]interface{}{
		"message":       "Entrenamiento iniciado exitosamente",
		"unit_type":     req.UnitType,
		"quantity":      req.Quantity,
		"training_time": unitType.TrainingTime,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *UnitHandler) GetUnitTypes(w http.ResponseWriter, r *http.Request) {
	// Devolver todos los tipos de unidades disponibles
	var response []UnitResponse
	for unitType, unitTypeInfo := range models.UnitTypes {
		unitResp := UnitResponse{
			Type:         unitType,
			Name:         unitTypeInfo.Name,
			Description:  unitTypeInfo.Description,
			Cost:         unitTypeInfo.Cost,
			TrainingTime: unitTypeInfo.TrainingTime,
		}
		response = append(response, unitResp)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
