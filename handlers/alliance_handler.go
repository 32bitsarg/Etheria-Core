package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"server-backend/models"
	"server-backend/repository"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type AllianceHandler struct {
	allianceRepo *repository.AllianceRepository
	logger       *zap.Logger
}

func NewAllianceHandler(allianceRepo *repository.AllianceRepository, logger *zap.Logger) *AllianceHandler {
	return &AllianceHandler{
		allianceRepo: allianceRepo,
		logger:       logger,
	}
}

// CreateAlliance crea una nueva alianza
func (h *AllianceHandler) CreateAlliance(w http.ResponseWriter, r *http.Request) {
	var alliance models.Alliance
	if err := json.NewDecoder(r.Body).Decode(&alliance); err != nil {
		h.logger.Error("Error decodificando request", zap.Error(err))
		http.Error(w, "Error decodificando request", http.StatusBadRequest)
		return
	}

	// Obtener playerID del contexto (seteado por el middleware de auth)
	playerID := r.Context().Value("playerID").(int)
	alliance.LeaderID = playerID

	createdAlliance, err := h.allianceRepo.CreateAlliance(&alliance)
	if err != nil {
		h.logger.Error("Error creando alianza", zap.Error(err))
		http.Error(w, "Error creando alianza", http.StatusInternalServerError)
		return
	}

	// Agregar al líder como miembro
	member := models.AllianceMember{
		AllianceID: createdAlliance.ID,
		PlayerID:   playerID,
		Role:       "leader",
		JoinedAt:   createdAlliance.CreatedAt,
	}

	if err := h.allianceRepo.AddMember(&member); err != nil {
		h.logger.Error("Error agregando líder como miembro", zap.Error(err))
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdAlliance)
}

// GetAlliances lista todas las alianzas
func (h *AllianceHandler) GetAlliances(w http.ResponseWriter, r *http.Request) {
	alliances, err := h.allianceRepo.GetAlliances()
	if err != nil {
		h.logger.Error("Error obteniendo alianzas", zap.Error(err))
		http.Error(w, "Error obteniendo alianzas", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(alliances)
}

// GetAlliance obtiene una alianza específica
func (h *AllianceHandler) GetAlliance(w http.ResponseWriter, r *http.Request) {
	allianceID, err := strconv.Atoi(chi.URLParam(r, "allianceID"))
	if err != nil {
		http.Error(w, "ID de alianza inválido", http.StatusBadRequest)
		return
	}

	alliance, err := h.allianceRepo.GetAlliance(allianceID)
	if err != nil {
		h.logger.Error("Error obteniendo alianza", zap.Error(err))
		http.Error(w, "Alianza no encontrada", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(alliance)
}

// UpdateAlliance actualiza una alianza
func (h *AllianceHandler) UpdateAlliance(w http.ResponseWriter, r *http.Request) {
	allianceID, err := strconv.Atoi(chi.URLParam(r, "allianceID"))
	if err != nil {
		http.Error(w, "ID de alianza inválido", http.StatusBadRequest)
		return
	}

	playerID := r.Context().Value("playerID").(int)

	// Verificar que el jugador sea líder de la alianza
	isLeader, err := h.allianceRepo.IsPlayerLeader(allianceID, playerID)
	if err != nil || !isLeader {
		http.Error(w, "No autorizado", http.StatusForbidden)
		return
	}

	var alliance models.Alliance
	if err := json.NewDecoder(r.Body).Decode(&alliance); err != nil {
		h.logger.Error("Error decodificando request", zap.Error(err))
		http.Error(w, "Error decodificando request", http.StatusBadRequest)
		return
	}

	alliance.ID = allianceID
	updatedAlliance, err := h.allianceRepo.UpdateAlliance(&alliance)
	if err != nil {
		h.logger.Error("Error actualizando alianza", zap.Error(err))
		http.Error(w, "Error actualizando alianza", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedAlliance)
}

// DeleteAlliance elimina una alianza
func (h *AllianceHandler) DeleteAlliance(w http.ResponseWriter, r *http.Request) {
	allianceID, err := strconv.Atoi(chi.URLParam(r, "allianceID"))
	if err != nil {
		http.Error(w, "ID de alianza inválido", http.StatusBadRequest)
		return
	}

	playerID := r.Context().Value("playerID").(int)

	// Verificar que el jugador sea líder de la alianza
	isLeader, err := h.allianceRepo.IsPlayerLeader(allianceID, playerID)
	if err != nil || !isLeader {
		http.Error(w, "No autorizado", http.StatusForbidden)
		return
	}

	if err := h.allianceRepo.DeleteAlliance(allianceID); err != nil {
		h.logger.Error("Error eliminando alianza", zap.Error(err))
		http.Error(w, "Error eliminando alianza", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetMembers obtiene los miembros de una alianza
func (h *AllianceHandler) GetMembers(w http.ResponseWriter, r *http.Request) {
	allianceID, err := strconv.Atoi(chi.URLParam(r, "allianceID"))
	if err != nil {
		http.Error(w, "ID de alianza inválido", http.StatusBadRequest)
		return
	}

	members, err := h.allianceRepo.GetAllianceMembers(allianceID)
	if err != nil {
		h.logger.Error("Error obteniendo miembros", zap.Error(err))
		http.Error(w, "Error obteniendo miembros", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(members)
}

// JoinAlliance permite a un jugador unirse a una alianza
func (h *AllianceHandler) JoinAlliance(w http.ResponseWriter, r *http.Request) {
	allianceID, err := strconv.Atoi(chi.URLParam(r, "allianceID"))
	if err != nil {
		http.Error(w, "ID de alianza inválido", http.StatusBadRequest)
		return
	}

	playerID := r.Context().Value("playerID").(int)

	// Verificar que el jugador no esté ya en una alianza
	currentAlliance, err := h.allianceRepo.GetPlayerAlliance(playerID)
	if err == nil && currentAlliance != nil {
		http.Error(w, "Ya perteneces a una alianza", http.StatusBadRequest)
		return
	}

	member := models.AllianceMember{
		AllianceID: allianceID,
		PlayerID:   playerID,
		Role:       "member",
	}

	if err := h.allianceRepo.AddMember(&member); err != nil {
		h.logger.Error("Error uniéndose a la alianza", zap.Error(err))
		http.Error(w, "Error uniéndose a la alianza", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// LeaveAlliance permite a un jugador salir de una alianza
func (h *AllianceHandler) LeaveAlliance(w http.ResponseWriter, r *http.Request) {
	allianceID, err := strconv.Atoi(chi.URLParam(r, "allianceID"))
	if err != nil {
		http.Error(w, "ID de alianza inválido", http.StatusBadRequest)
		return
	}

	playerID := r.Context().Value("playerID").(int)

	// Verificar que el jugador sea miembro de la alianza
	isMember, err := h.allianceRepo.IsPlayerMember(allianceID, playerID)
	if err != nil || !isMember {
		http.Error(w, "No eres miembro de esta alianza", http.StatusForbidden)
		return
	}

	// Verificar que no sea el líder (los líderes no pueden salir)
	isLeader, err := h.allianceRepo.IsPlayerLeader(allianceID, playerID)
	if err == nil && isLeader {
		http.Error(w, "Los líderes no pueden salir de la alianza", http.StatusBadRequest)
		return
	}

	if err := h.allianceRepo.RemoveMember(allianceID, playerID); err != nil {
		h.logger.Error("Error saliendo de la alianza", zap.Error(err))
		http.Error(w, "Error saliendo de la alianza", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// PromoteMember promueve a un miembro a oficial
func (h *AllianceHandler) PromoteMember(w http.ResponseWriter, r *http.Request) {
	allianceID, err := strconv.Atoi(chi.URLParam(r, "allianceID"))
	if err != nil {
		http.Error(w, "ID de alianza inválido", http.StatusBadRequest)
		return
	}

	playerID := r.Context().Value("playerID").(int)

	// Verificar que el jugador sea líder de la alianza
	isLeader, err := h.allianceRepo.IsPlayerLeader(allianceID, playerID)
	if err != nil || !isLeader {
		http.Error(w, "No autorizado", http.StatusForbidden)
		return
	}

	var request struct {
		MemberID int `json:"member_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Error decodificando request", http.StatusBadRequest)
		return
	}

	if err := h.allianceRepo.PromoteMember(allianceID, request.MemberID); err != nil {
		h.logger.Error("Error promoviendo miembro", zap.Error(err))
		http.Error(w, "Error promoviendo miembro", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// DemoteMember degrada a un oficial a miembro
func (h *AllianceHandler) DemoteMember(w http.ResponseWriter, r *http.Request) {
	allianceID, err := strconv.Atoi(chi.URLParam(r, "allianceID"))
	if err != nil {
		http.Error(w, "ID de alianza inválido", http.StatusBadRequest)
		return
	}

	playerID := r.Context().Value("playerID").(int)

	// Verificar que el jugador sea líder de la alianza
	isLeader, err := h.allianceRepo.IsPlayerLeader(allianceID, playerID)
	if err != nil || !isLeader {
		http.Error(w, "No autorizado", http.StatusForbidden)
		return
	}

	var request struct {
		MemberID int `json:"member_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Error decodificando request", http.StatusBadRequest)
		return
	}

	if err := h.allianceRepo.DemoteMember(allianceID, request.MemberID); err != nil {
		h.logger.Error("Error degradando miembro", zap.Error(err))
		http.Error(w, "Error degradando miembro", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// KickMember expulsa a un miembro de la alianza
func (h *AllianceHandler) KickMember(w http.ResponseWriter, r *http.Request) {
	allianceID, err := strconv.Atoi(chi.URLParam(r, "allianceID"))
	if err != nil {
		http.Error(w, "ID de alianza inválido", http.StatusBadRequest)
		return
	}

	playerID := r.Context().Value("playerID").(int)

	// Verificar que el jugador sea líder u oficial de la alianza
	role, err := h.allianceRepo.GetPlayerRole(allianceID, playerID)
	if err != nil || (role != "leader" && role != "officer") {
		http.Error(w, "No autorizado", http.StatusForbidden)
		return
	}

	var request struct {
		MemberID int `json:"member_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Error decodificando request", http.StatusBadRequest)
		return
	}

	// Los oficiales no pueden expulsar a otros oficiales
	if role == "officer" {
		memberRole, err := h.allianceRepo.GetPlayerRole(allianceID, request.MemberID)
		if err == nil && memberRole == "officer" {
			http.Error(w, "Los oficiales no pueden expulsar a otros oficiales", http.StatusForbidden)
			return
		}
	}

	if err := h.allianceRepo.RemoveMember(allianceID, request.MemberID); err != nil {
		h.logger.Error("Error expulsando miembro", zap.Error(err))
		http.Error(w, "Error expulsando miembro", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetPlayerAlliance obtiene la alianza del jugador actual
func (h *AllianceHandler) GetPlayerAlliance(w http.ResponseWriter, r *http.Request) {
	playerID := r.Context().Value("playerID").(int)

	alliance, err := h.allianceRepo.GetPlayerAlliance(playerID)
	if err != nil {
		// Si no está en ninguna alianza, devolver null
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(alliance)
}
