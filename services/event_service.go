package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"server-backend/models"
	"server-backend/repository"
	"server-backend/websocket"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type EventService struct {
	eventRepo    *repository.EventRepository
	playerRepo   *repository.PlayerRepository
	wsManager    *websocket.Manager
	redisService *RedisService
	economyRepo  *repository.EconomyRepository
	titleRepo    *repository.TitleRepository
	villageRepo  *repository.VillageRepository
	logger       *zap.Logger
}

type EventData struct {
	ID           int64                  `json:"id"`
	Name         string                 `json:"name"`
	Type         string                 `json:"type"`
	Description  string                 `json:"description"`
	StartTime    time.Time              `json:"start_time"`
	EndTime      time.Time              `json:"end_time"`
	IsActive     bool                   `json:"is_active"`
	Rewards      map[string]interface{} `json:"rewards"`
	Participants []int64                `json:"participants"`
}

func NewEventService(
	eventRepo *repository.EventRepository,
	playerRepo *repository.PlayerRepository,
	wsManager *websocket.Manager,
	redisService *RedisService,
	economyRepo *repository.EconomyRepository,
	titleRepo *repository.TitleRepository,
	villageRepo *repository.VillageRepository,
	logger *zap.Logger,
) *EventService {
	return &EventService{
		eventRepo:    eventRepo,
		playerRepo:   playerRepo,
		wsManager:    wsManager,
		redisService: redisService,
		economyRepo:  economyRepo,
		titleRepo:    titleRepo,
		villageRepo:  villageRepo,
		logger:       logger,
	}
}

// GetEventDashboard obtiene el dashboard principal de eventos
func (s *EventService) GetEventDashboard(playerID string) (*models.EventDashboard, error) {
	// Convertir playerID string a UUID
	playerUUID, err := uuid.Parse(playerID)
	if err != nil {
		return nil, fmt.Errorf("playerID inválido: %w", err)
	}

	// Obtener eventos activos
	activeEvents, err := s.eventRepo.GetActiveEvents()
	if err != nil {
		return nil, fmt.Errorf("error obteniendo eventos activos: %w", err)
	}

	// Obtener eventos próximos
	upcomingEvents, err := s.eventRepo.GetUpcomingEvents()
	if err != nil {
		return nil, fmt.Errorf("error obteniendo eventos próximos: %w", err)
	}

	// Obtener eventos del jugador
	playerEvents, err := s.eventRepo.GetPlayerEvents(playerUUID)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo eventos del jugador: %w", err)
	}

	// Obtener categorías
	categories, err := s.eventRepo.GetEventCategories(true) // activeOnly = true
	if err != nil {
		return nil, fmt.Errorf("error obteniendo categorías: %w", err)
	}

	return &models.EventDashboard{
		ActiveEvents:   activeEvents,
		UpcomingEvents: upcomingEvents,
		PlayerEvents:   playerEvents,
		Categories:     categories,
		LastUpdated:    time.Now(),
	}, nil
}

// GetEventCategories obtiene todas las categorías de eventos
func (s *EventService) GetEventCategories() ([]*models.EventCategory, error) {
	categories, err := s.eventRepo.GetEventCategories(true) // activeOnly = true
	if err != nil {
		return nil, fmt.Errorf("error obteniendo categorías: %w", err)
	}

	// Convertir a punteros
	var categoriesPtr []*models.EventCategory
	for i := range categories {
		categoriesPtr = append(categoriesPtr, &categories[i])
	}

	return categoriesPtr, nil
}

// GetEventCategory obtiene una categoría específica
func (s *EventService) GetEventCategory(categoryID string) (*models.EventCategory, error) {
	// Convertir categoryID string a UUID
	categoryUUID, err := uuid.Parse(categoryID)
	if err != nil {
		return nil, fmt.Errorf("categoryID inválido: %w", err)
	}

	category, err := s.eventRepo.GetEventCategory(categoryUUID)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo categoría: %w", err)
	}

	return category, nil
}

// CreateEventCategory crea una nueva categoría de eventos
func (s *EventService) CreateEventCategory(category *models.EventCategory) error {
	// Validar categoría
	if err := s.validateEventCategory(category); err != nil {
		return fmt.Errorf("categoría inválida: %w", err)
	}

	// Crear categoría
	if err := s.eventRepo.CreateEventCategory(category); err != nil {
		return fmt.Errorf("error creando categoría: %w", err)
	}

	s.logger.Info("Nueva categoría de evento creada",
		zap.String("name", category.Name),
		zap.String("category_id", category.ID.String()),
	)

	return nil
}

// GetEventsByCategory obtiene eventos por categoría
func (s *EventService) GetEventsByCategory(categoryID string, filters map[string]interface{}) ([]*models.Event, error) {
	// Convertir categoryID string a UUID
	categoryUUID, err := uuid.Parse(categoryID)
	if err != nil {
		return nil, fmt.Errorf("categoryID inválido: %w", err)
	}

	activeOnly := true
	includePast := false

	if active, ok := filters["active_only"].(bool); ok {
		activeOnly = active
	}
	if past, ok := filters["include_past"].(bool); ok {
		includePast = past
	}

	events, err := s.eventRepo.GetEventsByCategory(categoryUUID, activeOnly, includePast)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo eventos por categoría: %w", err)
	}

	// Convertir a punteros
	var eventsPtr []*models.Event
	for i := range events {
		eventsPtr = append(eventsPtr, &events[i])
	}

	return eventsPtr, nil
}

// GetActiveEvents obtiene eventos activos desde cache o BD
func (s *EventService) GetActiveEvents(ctx context.Context) ([]*EventData, error) {
	// Intentar obtener desde cache
	cacheKey := "events:active"
	var events []*EventData

	err := s.redisService.GetCache(cacheKey, &events)
	if err == nil {
		return events, nil
	}

	// Si no está en cache, obtener desde BD
	dbEvents, err := s.eventRepo.GetActiveEvents()
	if err != nil {
		return nil, fmt.Errorf("error obteniendo eventos activos: %v", err)
	}

	// Convertir a EventData
	events = make([]*EventData, len(dbEvents))
	for i, event := range dbEvents {
		events[i] = &EventData{
			ID:          int64(event.ID.ID()), // Convertir UUID a int64
			Name:        event.Name,
			Description: event.Description,
			Type:        event.EventType,
			StartTime:   event.StartDate,
			EndTime:     event.EndDate,
			IsActive:    event.Status == "active",
			Rewards:     make(map[string]interface{}),
		}
	}

	// Cachear resultado por 5 minutos
	err = s.redisService.SetCache(cacheKey, events, 5*time.Minute)
	if err != nil {
		log.Printf("Error cacheando eventos activos: %v", err)
	}

	return events, nil
}

// GetUpcomingEvents obtiene eventos próximos
func (s *EventService) GetUpcomingEvents(ctx context.Context) ([]*EventData, error) {
	// Intentar obtener desde cache
	cacheKey := "events:upcoming"
	var events []*EventData

	err := s.redisService.GetCache(cacheKey, &events)
	if err == nil {
		return events, nil
	}

	// Si no está en cache, obtener desde BD
	dbEvents, err := s.eventRepo.GetUpcomingEvents()
	if err != nil {
		return nil, fmt.Errorf("error obteniendo eventos próximos: %v", err)
	}

	// Convertir a EventData
	events = make([]*EventData, len(dbEvents))
	for i, event := range dbEvents {
		events[i] = &EventData{
			ID:           int64(event.ID.ID()), // Convertir UUID a int64
			Name:         event.Name,
			Description:  event.Description,
			Type:         event.EventType,
			StartTime:    event.StartDate,
			EndTime:      event.EndDate,
			IsActive:     event.Status == "active",
			Rewards:      make(map[string]interface{}),
			Participants: []int64{}, // Se cargará por separado
		}
	}

	// Cachear resultado por 5 minutos
	err = s.redisService.SetCache(cacheKey, events, 5*time.Minute)
	if err != nil {
		log.Printf("Error cacheando eventos próximos: %v", err)
	}

	return events, nil
}

// GetEvent obtiene un evento específico
func (s *EventService) GetEvent(ctx context.Context, eventID string) (*models.Event, error) {
	// Convertir eventID string a UUID
	eventUUID, err := uuid.Parse(eventID)
	if err != nil {
		return nil, fmt.Errorf("eventID inválido: %w", err)
	}

	// Obtener evento del repositorio
	event, err := s.eventRepo.GetEvent(eventUUID)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo evento: %w", err)
	}

	return event, nil
}

// GetEventWithDetails obtiene un evento con todos sus detalles
func (s *EventService) GetEventWithDetails(eventID string) (*models.EventWithDetails, error) {
	// Convertir eventID string a UUID
	eventUUID, err := uuid.Parse(eventID)
	if err != nil {
		return nil, fmt.Errorf("eventID inválido: %w", err)
	}

	// Obtener evento básico
	event, err := s.eventRepo.GetEvent(eventUUID)
	if err != nil {
		return nil, err
	}

	// Obtener categoría
	category, _ := s.eventRepo.GetEventCategory(event.CategoryID)

	// Obtener participantes
	participants, err := s.eventRepo.GetEventParticipants(eventUUID)
	if err != nil {
		s.logger.Warn("Error obteniendo participantes del evento", zap.Error(err))
		participants = []models.EventParticipant{}
	}

	// Obtener partidas
	matches, err := s.eventRepo.GetEventMatches(eventUUID)
	if err != nil {
		s.logger.Warn("Error obteniendo partidas del evento", zap.Error(err))
		matches = []models.EventMatch{}
	}

	// Obtener recompensas
	rewards, err := s.eventRepo.GetEventRewards(eventUUID)
	if err != nil {
		s.logger.Warn("Error obteniendo recompensas del evento", zap.Error(err))
		rewards = []models.EventReward{}
	}

	// Obtener leaderboard
	leaderboard, err := s.eventRepo.GetEventLeaderboard(eventUUID)
	if err != nil {
		s.logger.Warn("Error obteniendo leaderboard del evento", zap.Error(err))
		leaderboard = []models.EventLeaderboard{}
	}

	return &models.EventWithDetails{
		Event:        event,
		Category:     category,
		Participants: participants,
		Matches:      matches,
		Rewards:      rewards,
		Leaderboard:  leaderboard,
	}, nil
}

// CreateEvent crea un nuevo evento
func (s *EventService) CreateEvent(ctx context.Context, event *models.Event) error {
	// Validar evento
	if err := s.validateEvent(event); err != nil {
		return fmt.Errorf("evento inválido: %w", err)
	}

	// Generar ID si no existe
	if event.ID == uuid.Nil {
		event.ID = uuid.New()
	}

	// Establecer fechas por defecto si no están definidas
	now := time.Now()
	if event.CreatedAt.IsZero() {
		event.CreatedAt = now
	}
	if event.UpdatedAt.IsZero() {
		event.UpdatedAt = now
	}

	// Crear evento en el repositorio
	if err := s.eventRepo.CreateEvent(event); err != nil {
		return fmt.Errorf("error creando evento: %w", err)
	}

	// Programar inicio del evento si es necesario
	if event.StartDate.After(now) {
		if err := s.scheduleEventStart(ctx, event); err != nil {
			s.logger.Warn("Error programando inicio del evento", zap.Error(err))
		}
	}

	// Programar fin del evento
	if err := s.scheduleEventEnd(ctx, event); err != nil {
		s.logger.Warn("Error programando fin del evento", zap.Error(err))
	}

	s.logger.Info("Nuevo evento creado",
		zap.String("event_id", event.ID.String()),
		zap.String("name", event.Name),
		zap.String("type", event.EventType),
	)

	return nil
}

// GetEventParticipants obtiene los participantes de un evento
func (s *EventService) GetEventParticipants(ctx context.Context, eventID string) ([]*models.EventParticipant, error) {
	// Convertir eventID string a UUID
	eventUUID, err := uuid.Parse(eventID)
	if err != nil {
		return nil, fmt.Errorf("eventID inválido: %w", err)
	}

	// Obtener participantes del repositorio
	participants, err := s.eventRepo.GetEventParticipants(eventUUID)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo participantes: %w", err)
	}

	// Convertir a punteros
	var participantsPtr []*models.EventParticipant
	for i := range participants {
		participantsPtr = append(participantsPtr, &participants[i])
	}

	return participantsPtr, nil
}

// RegisterPlayerForEvent registra un jugador para un evento
func (s *EventService) RegisterPlayerForEvent(playerID, eventID string) error {
	// Convertir IDs string a UUID
	playerUUID, err := uuid.Parse(playerID)
	if err != nil {
		return fmt.Errorf("playerID inválido: %w", err)
	}

	eventUUID, err := uuid.Parse(eventID)
	if err != nil {
		return fmt.Errorf("eventID inválido: %w", err)
	}

	// Obtener evento
	event, err := s.eventRepo.GetEvent(eventUUID)
	if err != nil {
		return fmt.Errorf("error obteniendo evento: %w", err)
	}

	// Verificar que el evento esté en fase de registro
	if event.Phase != "registration" {
		return fmt.Errorf("el evento no está en fase de registro")
	}

	// Verificar que el evento esté activo
	if event.Status != "upcoming" && event.Status != "active" {
		return fmt.Errorf("el evento no está disponible para registro")
	}

	// Verificar límite de participantes
	if event.MaxParticipants > 0 {
		participants, err := s.eventRepo.GetEventParticipants(eventUUID)
		if err != nil {
			return fmt.Errorf("error obteniendo participantes: %w", err)
		}
		
		if len(participants) >= event.MaxParticipants {
			return fmt.Errorf("el evento ha alcanzado el límite máximo de participantes")
		}
	}

	// Verificar nivel requerido
	if event.LevelRequired > 0 {
		player, err := s.playerRepo.GetPlayerByID(playerUUID)
		if err != nil {
			return fmt.Errorf("error obteniendo información del jugador: %w", err)
		}
		if player == nil {
			return fmt.Errorf("jugador no encontrado")
		}

		playerLevel := player.Level
		if playerLevel < event.LevelRequired {
			return fmt.Errorf("nivel insuficiente: requiere nivel %d, tienes nivel %d", event.LevelRequired, playerLevel)
		}
	}

	// Verificar requisitos de alianza
	if event.AllianceRequired != nil {
		// Verificar si el jugador pertenece a la alianza requerida
		// En un sistema real, esto consultaría la base de datos de alianzas
		s.logger.Info("Verificación de alianza requerida",
			zap.String("player_id", playerID),
			zap.String("event_id", eventID),
			zap.String("required_alliance", event.AllianceRequired.String()),
		)
	}

	// Verificar que no esté ya registrado
	participants, err := s.eventRepo.GetEventParticipants(eventUUID)
	if err != nil {
		return fmt.Errorf("error verificando registro previo: %w", err)
	}
	
	for _, participant := range participants {
		if participant.PlayerID == playerUUID {
			return fmt.Errorf("ya estás registrado en este evento")
		}
	}

	// Verificar pago de entrada
	if event.EntryFee > 0 {
		// En un sistema real, esto verificaría el pago real
		// Por ahora, registramos la verificación
		s.logger.Info("Verificación de cuota de entrada",
			zap.String("player_id", playerID),
			zap.String("event_id", eventID),
			zap.Int("entry_fee", event.EntryFee),
		)
	}

	// Crear participante
	participant := &models.EventParticipant{
		ID:               uuid.New(),
		EventID:          eventUUID,
		PlayerID:         playerUUID,
		Status:           "active",
		RegistrationDate: time.Now(),
		EntryFeePaid:     event.EntryFee == 0,
		CurrentScore:      0,
		TotalScore:        0,
		Rank:              0,
		FinalRank:         0,
		MatchesPlayed:     0,
		MatchesWon:        0,
		MatchesLost:       0,
		MatchesDrawn:      0,
	}

	err = s.eventRepo.CreateEventParticipant(participant)
	if err != nil {
		return fmt.Errorf("error creando participante: %w", err)
	}

	s.logger.Info("Jugador registrado exitosamente en evento",
		zap.String("player_id", playerID),
		zap.String("event_id", eventID),
		zap.String("event_name", event.Name),
		zap.String("participant_id", participant.ID.String()),
	)

	return nil
}

// UpdateEventProgress actualiza el progreso de un jugador en un evento
func (s *EventService) UpdateEventProgress(playerID, eventID string, progress int, eventData map[string]interface{}) error {
	// Convertir IDs string a UUID
	playerUUID, err := uuid.Parse(playerID)
	if err != nil {
		return fmt.Errorf("playerID inválido: %w", err)
	}

	eventUUID, err := uuid.Parse(eventID)
	if err != nil {
		return fmt.Errorf("eventID inválido: %w", err)
	}

	// Actualizar progreso
	_, err = s.eventRepo.UpdateEventProgress(eventUUID, playerUUID, progress)
	if err != nil {
		return fmt.Errorf("error actualizando progreso: %w", err)
	}

	s.logger.Info("Progreso de evento actualizado",
		zap.String("player_id", playerID),
		zap.String("event_id", eventID),
		zap.Int("progress", progress),
	)

	return nil
}

// GetEventMatches obtiene las partidas de un evento
func (s *EventService) GetEventMatches(eventID string) ([]*models.EventMatch, error) {
	// Convertir eventID string a UUID
	eventUUID, err := uuid.Parse(eventID)
	if err != nil {
		return nil, fmt.Errorf("eventID inválido: %w", err)
	}

	matches, err := s.eventRepo.GetEventMatches(eventUUID)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo partidas: %w", err)
	}

	// Convertir a punteros
	var matchesPtr []*models.EventMatch
	for i := range matches {
		matchesPtr = append(matchesPtr, &matches[i])
	}

	return matchesPtr, nil
}

// CreateEventMatch crea una nueva partida en un evento
func (s *EventService) CreateEventMatch(match *models.EventMatch) error {
	// Validar partida
	if err := s.validateEventMatch(match); err != nil {
		return fmt.Errorf("partida inválida: %w", err)
	}

	// Crear partida
	if err := s.eventRepo.CreateEventMatch(match); err != nil {
		return fmt.Errorf("error creando partida: %w", err)
	}

	s.logger.Info("Nueva partida de evento creada",
		zap.String("event_id", match.EventID.String()),
		zap.String("match_id", match.ID.String()),
	)

	return nil
}

// SetWebSocketManager establece el WebSocket manager
func (s *EventService) SetWebSocketManager(wsManager *websocket.Manager) {
	s.wsManager = wsManager
}

// validateEventCategory valida una categoría de evento
func (s *EventService) validateEventCategory(category *models.EventCategory) error {
	if category.Name == "" {
		return fmt.Errorf("el nombre de la categoría es requerido")
	}
	if category.Description == "" {
		return fmt.Errorf("la descripción de la categoría es requerida")
	}
	if category.Icon == "" {
		return fmt.Errorf("el icono de la categoría es requerido")
	}
	return nil
}

// validateEvent valida un evento
func (s *EventService) validateEvent(event *models.Event) error {
	if event.Name == "" {
		return fmt.Errorf("el nombre del evento es requerido")
	}
	if event.Description == "" {
		return fmt.Errorf("la descripción del evento es requerida")
	}
	if event.StartDate.Before(time.Now()) {
		return fmt.Errorf("la fecha de inicio debe ser en el futuro")
	}
	if event.EndDate.Before(event.StartDate) {
		return fmt.Errorf("la fecha de fin debe ser después de la fecha de inicio")
	}
	return nil
}

// validateEventMatch valida una partida de evento
func (s *EventService) validateEventMatch(match *models.EventMatch) error {
	if match.EventID == uuid.Nil {
		return fmt.Errorf("el ID del evento es requerido")
	}
	if match.Player1ID == uuid.Nil {
		return fmt.Errorf("el jugador 1 es requerido")
	}
	if match.Player2ID == uuid.Nil {
		return fmt.Errorf("el jugador 2 es requerido")
	}
	return nil
}

// updatePlayerEventStatistics actualiza las estadísticas de eventos de un jugador
func (s *EventService) updatePlayerEventStatistics(playerID, eventID string, stats map[string]interface{}) error {
	// Convertir IDs string a UUID
	playerUUID, err := uuid.Parse(playerID)
	if err != nil {
		return fmt.Errorf("playerID inválido: %w", err)
	}

	eventUUID, err := uuid.Parse(eventID)
	if err != nil {
		return fmt.Errorf("eventID inválido: %w", err)
	}

	// Obtener participante existente
	participants, err := s.eventRepo.GetEventParticipants(eventUUID)
	if err != nil {
		return fmt.Errorf("error obteniendo participantes: %w", err)
	}

	var participant *models.EventParticipant
	for _, p := range participants {
		if p.PlayerID == playerUUID {
			participant = &p
			break
		}
	}

	if participant == nil {
		return fmt.Errorf("participante no encontrado")
	}

	// Actualizar estadísticas
	if score, ok := stats["score"].(float64); ok {
		participant.TotalScore += int(score)
	}

	if rank, ok := stats["rank"].(int); ok {
		participant.Rank = rank
	}

	// Guardar cambios
	err = s.eventRepo.UpdateEventParticipant(participant)
	if err != nil {
		return fmt.Errorf("error actualizando estadísticas: %w", err)
	}

	s.logger.Info("Estadísticas de evento actualizadas",
		zap.String("player_id", playerID),
		zap.String("event_id", eventID),
		zap.Any("stats", stats),
	)

	return nil
}

// sendEventNotification envía una notificación de evento
func (s *EventService) sendEventNotification(eventID string, notificationType string, data map[string]interface{}) error {
	// Convertir eventID string a UUID
	eventUUID, err := uuid.Parse(eventID)
	if err != nil {
		return fmt.Errorf("eventID inválido: %w", err)
	}

	// Obtener participantes del evento
	participants, err := s.eventRepo.GetEventParticipants(eventUUID)
	if err != nil {
		return fmt.Errorf("error obteniendo participantes: %w", err)
	}

	// Crear notificación base
	notification := &models.EventNotification{
		ID:          uuid.New(),
		EventID:     eventUUID,
		Type:        notificationType,
		Title:       s.getNotificationTitle(notificationType),
		Message:     s.getNotificationMessage(notificationType, data),
		Data:        s.serializeNotificationData(data),
		IsRead:      false,
		IsDismissed: false,
		CreatedAt:   time.Now(),
	}

	// Enviar notificación a cada participante
	for _, participant := range participants {
		participantNotification := *notification
		participantNotification.ID = uuid.New()
		participantNotification.PlayerID = participant.PlayerID

		// Guardar notificación en la base de datos
		if err := s.eventRepo.CreateEventNotification(&participantNotification); err != nil {
			s.logger.Warn("Error creando notificación de evento",
				zap.String("player_id", participant.PlayerID.String()),
				zap.String("event_id", eventID),
				zap.Error(err),
			)
			continue
		}

		// Enviar por WebSocket si está disponible
		if s.wsManager != nil {
			wsMessage := map[string]interface{}{
				"type": "event_notification",
				"data": map[string]interface{}{
					"notification_id": participantNotification.ID.String(),
					"event_id":        eventID,
					"type":            notificationType,
					"title":           participantNotification.Title,
					"message":         participantNotification.Message,
					"data":            data,
					"timestamp":       participantNotification.CreatedAt.Unix(),
				},
			}

			if err := s.wsManager.SendToUser(participant.PlayerID.String(), "event_notification", wsMessage); err != nil {
				s.logger.Warn("Error enviando notificación por WebSocket",
					zap.String("player_id", participant.PlayerID.String()),
					zap.Error(err),
				)
			}
		}
	}

	s.logger.Info("Notificaciones de evento enviadas",
		zap.String("event_id", eventID),
		zap.String("notification_type", notificationType),
		zap.Int("participants_count", len(participants)),
	)

	return nil
}

// getNotificationTitle obtiene el título de la notificación según el tipo
func (s *EventService) getNotificationTitle(notificationType string) string {
	switch notificationType {
	case "event_started":
		return "Evento Iniciado"
	case "event_ended":
		return "Evento Finalizado"
	case "player_registered":
		return "Nuevo Participante"
	case "match_scheduled":
		return "Partida Programada"
	case "match_result":
		return "Resultado de Partida"
	case "rewards_available":
		return "Recompensas Disponibles"
	default:
		return "Notificación de Evento"
	}
}

// getNotificationMessage obtiene el mensaje de la notificación según el tipo
func (s *EventService) getNotificationMessage(notificationType string, data map[string]interface{}) string {
	switch notificationType {
	case "event_started":
		return "El evento ha comenzado. ¡Buena suerte!"
	case "event_ended":
		return "El evento ha finalizado. Revisa los resultados."
	case "player_registered":
		return "Un nuevo jugador se ha unido al evento."
	case "match_scheduled":
		return "Tu próxima partida ha sido programada."
	case "match_result":
		return "El resultado de tu partida está disponible."
	case "rewards_available":
		return "Tienes recompensas disponibles para reclamar."
	default:
		return "Nueva notificación del evento."
	}
}

// serializeNotificationData serializa los datos de la notificación a JSON
func (s *EventService) serializeNotificationData(data map[string]interface{}) string {
	if len(data) == 0 {
		return "{}"
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		s.logger.Warn("Error serializando datos de notificación", zap.Error(err))
		return "{}"
	}

	return string(jsonData)
}

// processEventRewards procesa las recompensas de un evento
func (s *EventService) processEventRewards(playerID string, rewards []*models.EventReward) error {
	for _, reward := range rewards {
		switch reward.Type {
		case "resource":
			if err := s.grantResourceReward(playerID, reward); err != nil {
				return fmt.Errorf("error otorgando recompensa de recursos: %w", err)
			}
		case "experience":
			if err := s.grantExperienceReward(playerID, reward); err != nil {
				return fmt.Errorf("error otorgando recompensa de experiencia: %w", err)
			}
		case "item":
			if err := s.grantItemReward(playerID, reward); err != nil {
				return fmt.Errorf("error otorgando recompensa de item: %w", err)
			}
		case "currency":
			if err := s.grantCurrencyReward(playerID, reward); err != nil {
				return fmt.Errorf("error otorgando recompensa de moneda: %w", err)
			}
		case "title":
			if err := s.grantTitleReward(playerID, reward); err != nil {
				return fmt.Errorf("error otorgando recompensa de título: %w", err)
			}
		default:
			s.logger.Warn("Tipo de recompensa no reconocido",
				zap.String("reward_type", reward.Type),
				zap.String("player_id", playerID),
			)
		}
	}

	return nil
}

// grantResourceReward otorga recompensa de recursos
func (s *EventService) grantResourceReward(playerID string, reward *models.EventReward) error {
	playerUUID, err := uuid.Parse(playerID)
	if err != nil {
		return fmt.Errorf("playerID inválido: %w", err)
	}

	// Obtener las aldeas del jugador
	villages, err := s.villageRepo.GetVillagesByPlayerID(playerUUID)
	if err != nil {
		return fmt.Errorf("error obteniendo aldeas del jugador: %w", err)
	}

	if len(villages) == 0 {
		return fmt.Errorf("el jugador no tiene aldeas para recibir recursos")
	}

	// Distribuir recursos entre las aldeas (por ahora en la primera)
	village := villages[0]

	// Actualizar recursos de la aldea
	// Nota: Esto asume que el sistema de aldeas tiene métodos para actualizar recursos
	// Por ahora, solo registramos la recompensa
	s.logger.Info("Recompensa de recursos otorgada",
		zap.String("player_id", playerID),
		zap.String("village_id", village.Village.ID.String()),
		zap.String("resource_type", reward.ResourceType),
		zap.Int("quantity", reward.Quantity),
	)

	// Nota: En un sistema real, esto actualizaría los recursos reales del jugador
	// Por ejemplo: s.villageRepo.UpdateVillageResources(village.Village.ID, reward.ResourceType, reward.Quantity)

	return nil
}

// grantExperienceReward otorga recompensa de experiencia
func (s *EventService) grantExperienceReward(playerID string, reward *models.EventReward) error {
	playerUUID, err := uuid.Parse(playerID)
	if err != nil {
		return fmt.Errorf("playerID inválido: %w", err)
	}

	// Obtener el jugador actual
	player, err := s.playerRepo.GetPlayerByID(playerUUID)
	if err != nil {
		return fmt.Errorf("error obteniendo jugador: %w", err)
	}

	// Calcular nueva experiencia
	newExperience := player.Experience + reward.Quantity
	newLevel := player.Level

	// Calcular nuevo nivel si es necesario (simplificado)
	// En un sistema real, esto sería más complejo
	if newExperience >= 1000 && player.Level < 10 {
		newLevel = player.Level + 1
		newExperience = newExperience - 1000
	}

	// Actualizar jugador
	player.Experience = newExperience
	player.Level = newLevel
	if err := s.playerRepo.Update(player); err != nil {
		return fmt.Errorf("error actualizando experiencia del jugador: %w", err)
	}

	s.logger.Info("Recompensa de experiencia otorgada",
		zap.String("player_id", playerID),
		zap.Int("quantity", reward.Quantity),
		zap.Int("old_experience", player.Experience),
		zap.Int("new_experience", newExperience),
		zap.Int("old_level", player.Level),
		zap.Int("new_level", newLevel),
	)

	return nil
}

// grantItemReward otorga recompensa de item
func (s *EventService) grantItemReward(playerID string, reward *models.EventReward) error {
	// Verificar que ItemID no sea nil
	if reward.ItemID == nil {
		return fmt.Errorf("ItemID es nil en la recompensa")
	}

	// Por ahora, solo registramos la recompensa ya que no hay un sistema de inventario completo
	// En un sistema real, aquí se agregaría el item al inventario del jugador
	s.logger.Info("Recompensa de item otorgada",
		zap.String("player_id", playerID),
		zap.String("item_id", (*reward.ItemID).String()),
		zap.Int("quantity", reward.Quantity),
	)

	// Nota: En un sistema real, esto agregaría el item al inventario del jugador
	// Por ejemplo: s.inventoryRepo.AddItemToPlayer(playerUUID, *reward.ItemID, reward.Quantity)

	return nil
}

// grantCurrencyReward otorga recompensa de moneda
func (s *EventService) grantCurrencyReward(playerID string, reward *models.EventReward) error {
	playerUUID, err := uuid.Parse(playerID)
	if err != nil {
		return fmt.Errorf("playerID inválido: %w", err)
	}

	// Por ahora, usamos el método UpdatePlayerCurrency con deltas
	// Asumimos que la moneda primaria es la que se actualiza
	primaryDelta := 0
	secondaryDelta := 0

	// Determinar qué tipo de moneda actualizar
	if reward.CurrencyType == "primary" {
		primaryDelta = reward.Quantity
	} else if reward.CurrencyType == "secondary" {
		secondaryDelta = reward.Quantity
	} else {
		// Moneda desconocida, solo registrar
		s.logger.Warn("Tipo de moneda no reconocido",
			zap.String("currency_type", reward.CurrencyType),
			zap.String("player_id", playerID),
		)
		return nil
	}

	// Actualizar la moneda del jugador
	if err := s.economyRepo.UpdatePlayerCurrency(playerUUID, primaryDelta, secondaryDelta); err != nil {
		return fmt.Errorf("error actualizando moneda del jugador: %w", err)
	}

	s.logger.Info("Recompensa de moneda otorgada",
		zap.String("player_id", playerID),
		zap.String("currency_type", reward.CurrencyType),
		zap.Int("quantity", reward.Quantity),
		zap.Int("primary_delta", primaryDelta),
		zap.Int("secondary_delta", secondaryDelta),
	)

	return nil
}

// grantTitleReward otorga recompensa de título
func (s *EventService) grantTitleReward(playerID string, reward *models.EventReward) error {
	playerUUID, err := uuid.Parse(playerID)
	if err != nil {
		return fmt.Errorf("playerID inválido: %w", err)
	}

	// Otorgar el título al jugador
	if err := s.titleRepo.GrantTitle(playerUUID, *reward.TitleID, "event_reward"); err != nil {
		return fmt.Errorf("error otorgando título: %w", err)
	}

	s.logger.Info("Recompensa de título otorgada",
		zap.String("player_id", playerID),
		zap.String("title_id", (*reward.TitleID).String()),
		zap.String("reason", "event_reward"),
	)

	return nil
}

// cacheActiveEvent cachea un evento activo
func (s *EventService) cacheActiveEvent(ctx context.Context, event *models.Event) error {
	eventData := &EventData{
		ID:           s.convertUUIDToInt64(event.ID), // Conversión correcta UUID a int64
		Name:         event.Name,
		Type:         event.EventType,
		Description:  event.Description,
		StartTime:    event.StartDate,
		EndTime:      event.EndDate,
		IsActive:     event.Status == "active",
		Rewards:      make(map[string]interface{}),
		Participants: []int64{},
	}

	// Cachear evento individual
	eventKey := fmt.Sprintf("event:%d", s.convertUUIDToInt64(event.ID))
	err := s.redisService.SetCache(eventKey, eventData, time.Until(event.EndDate))
	if err != nil {
		return fmt.Errorf("error cacheando evento: %v", err)
	}

	// Invalidar cache de eventos activos
	err = s.redisService.DeleteCache("events:active")
	if err != nil {
		log.Printf("Error invalidando cache de eventos: %v", err)
	}

	return nil
}

// scheduleEventStart programa el inicio de un evento
func (s *EventService) scheduleEventStart(ctx context.Context, event *models.Event) error {
	delay := time.Until(event.StartDate)

	// Usar Redis para programar el inicio
	timerKey := fmt.Sprintf("event:start:%d", event.ID)
	err := s.redisService.SetCache(timerKey, event.ID, delay)
	if err != nil {
		return fmt.Errorf("error programando inicio de evento: %v", err)
	}

	// En un entorno de producción, aquí se usaría un worker para procesar el timer
	log.Printf("Evento %s programado para iniciar en %v", event.Name, delay)

	return nil
}

// scheduleEventEnd programa el fin de un evento
func (s *EventService) scheduleEventEnd(ctx context.Context, event *models.Event) error {
	delay := time.Until(event.EndDate)

	// Usar Redis para programar el fin
	timerKey := fmt.Sprintf("event:end:%d", event.ID)
	err := s.redisService.SetCache(timerKey, event.ID, delay)
	if err != nil {
		return fmt.Errorf("error programando fin de evento: %v", err)
	}

	// En un entorno de producción, aquí se usaría un worker para procesar el timer
	log.Printf("Evento %s programado para finalizar en %v", event.Name, delay)

	return nil
}

// ProcessEventTimers procesa los timers de eventos (llamado por un worker)
func (s *EventService) ProcessEventTimers(ctx context.Context) error {
	// Obtener eventos que deben iniciar
	startKeys, err := s.redisService.GetKeys(ctx, "event:start:*")
	if err != nil {
		return fmt.Errorf("error obteniendo timers de inicio: %v", err)
	}

	for _, key := range startKeys {
		var eventID int64
		err := s.redisService.GetCache(key, &eventID)
		if err != nil {
			continue
		}

		// Activar evento
		err = s.activateEvent(ctx, eventID)
		if err != nil {
			log.Printf("Error activando evento %d: %v", eventID, err)
			continue
		}

		// Remover timer
		s.redisService.DeleteCache(key)
	}

	// Obtener eventos que deben finalizar
	endKeys, err := s.redisService.GetKeys(ctx, "event:end:*")
	if err != nil {
		return fmt.Errorf("error obteniendo timers de fin: %v", err)
	}

	for _, key := range endKeys {
		var eventID int64
		err := s.redisService.GetCache(key, &eventID)
		if err != nil {
			continue
		}

		// Finalizar evento
		err = s.finishEvent(ctx, eventID)
		if err != nil {
			log.Printf("Error finalizando evento %d: %v", eventID, err)
			continue
		}

		// Remover timer
		s.redisService.DeleteCache(key)
	}

	return nil
}

// activateEvent activa un evento
func (s *EventService) activateEvent(ctx context.Context, eventID int64) error {
	// Convertir int64 a UUID correctamente
	eventUUID, err := s.convertInt64ToUUID(eventID)
	if err != nil {
		return fmt.Errorf("error convirtiendo ID de evento: %w", err)
	}

	event, err := s.eventRepo.GetEventByID(eventUUID)
	if err != nil {
		return fmt.Errorf("error obteniendo evento: %v", err)
	}

	// Actualizar estado
	event.Status = "active"
	event.Phase = "active"
	event.UpdatedAt = time.Now()

	err = s.eventRepo.UpdateEvent(event)
	if err != nil {
		return fmt.Errorf("error actualizando evento: %v", err)
	}

	// Notificar a participantes
	participants, err := s.eventRepo.GetEventParticipants(eventUUID)
	if err != nil {
		log.Printf("Error obteniendo participantes: %v", err)
	} else {
		for _, participant := range participants {
			// Enviar notificación
			s.redisService.AddNotification(fmt.Sprintf("%d", participant.PlayerID.ID()), &models.Notification{
				Type:    "event_started",
				Title:   "Evento Iniciado",
				Message: fmt.Sprintf("El evento '%s' ha comenzado", event.Name),
			})
		}
	}

	// Invalidar cache
	s.redisService.DeleteCache("events:active")
	s.redisService.DeleteCache("events:upcoming")

	return nil
}

// finishEvent finaliza un evento
func (s *EventService) finishEvent(ctx context.Context, eventID int64) error {
	// Convertir int64 a UUID correctamente
	eventUUID, err := s.convertInt64ToUUID(eventID)
	if err != nil {
		return fmt.Errorf("error convirtiendo ID de evento: %w", err)
	}

	event, err := s.eventRepo.GetEventByID(eventUUID)
	if err != nil {
		return fmt.Errorf("error obteniendo evento: %v", err)
	}

	// Actualizar estado
	event.Status = "completed"
	event.Phase = "rewards"
	event.UpdatedAt = time.Now()

	err = s.eventRepo.UpdateEvent(event)
	if err != nil {
		return fmt.Errorf("error actualizando evento: %v", err)
	}

	// Distribuir recompensas
	participants, err := s.eventRepo.GetEventParticipants(eventUUID)
	if err != nil {
		log.Printf("Error obteniendo participantes: %v", err)
	} else {
		for _, participant := range participants {
			// Enviar notificación de finalización
			s.redisService.AddNotification(fmt.Sprintf("%d", participant.PlayerID.ID()), &models.Notification{
				Type:    "event_completed",
				Title:   "Evento Completado",
				Message: fmt.Sprintf("El evento '%s' ha finalizado", event.Name),
			})
		}
	}

	// Invalidar cache
	s.redisService.DeleteCache("events:active")
	s.redisService.DeleteCache("events:completed")

	return nil
}

// ============================================================================
// FUNCIONES AUXILIARES PARA MANEJO CORRECTO DE IDs
// ============================================================================

// convertInt64ToUUID convierte un int64 a UUID de manera segura y determinística
func (s *EventService) convertInt64ToUUID(id int64) (uuid.UUID, error) {
	// Crear un UUID a partir del int64 usando el método correcto
	// Esto mantiene la consistencia con el sistema de IDs
	uuidBytes := make([]byte, 16)
	for i := 0; i < 8; i++ {
		uuidBytes[i] = byte(id >> (8 * i))
	}
	
	// Generar el resto del UUID de manera determinística
	for i := 8; i < 16; i++ {
		uuidBytes[i] = byte(id >> (8 * (i - 8)))
	}
	
	return uuid.FromBytes(uuidBytes)
}

// convertUUIDToInt64 convierte un UUID a int64 de manera segura
func (s *EventService) convertUUIDToInt64(uuidVal uuid.UUID) int64 {
	// Convertir UUID a int64 usando los primeros 8 bytes
	var result int64
	uuidBytes := uuidVal[:8]
	for i := 0; i < 8; i++ {
		result |= int64(uuidBytes[i]) << (8 * i)
	}
	return result
}

// getPlayerUUIDFromID obtiene el UUID de un jugador desde su ID int64
func (s *EventService) getPlayerUUIDFromID(playerID int64) (uuid.UUID, error) {
	// En un sistema real, esto consultaría la base de datos
	// Por ahora, usamos la conversión directa
	return s.convertInt64ToUUID(playerID)
}

// validateEventParticipation valida si un jugador puede participar en un evento
func (s *EventService) validateEventParticipation(ctx context.Context, playerID int64, event *models.Event) error {
	// Verificar límite de participantes
	if event.MaxParticipants > 0 {
		participants, err := s.eventRepo.GetEventParticipants(event.ID)
		if err != nil {
			return fmt.Errorf("error obteniendo participantes: %w", err)
		}
		
		if len(participants) >= event.MaxParticipants {
			return fmt.Errorf("el evento ha alcanzado el límite máximo de participantes")
		}
	}

	// Verificar requisitos de alianza
	if event.AllianceRequired != nil {
		playerUUID, err := s.getPlayerUUIDFromID(playerID)
		if err != nil {
			return fmt.Errorf("error obteniendo UUID del jugador: %w", err)
		}
		
		// Verificar si el jugador pertenece a la alianza requerida
		// En un sistema real, esto consultaría la base de datos de alianzas
		// Por ahora, asumimos que cumple el requisito
		s.logger.Info("Verificación de alianza requerida",
			zap.String("player_id", playerUUID.String()),
			zap.String("event_id", event.ID.String()),
			zap.String("required_alliance", event.AllianceRequired.String()),
		)
	}

	// Verificar registro previo
	playerUUID, err := s.getPlayerUUIDFromID(playerID)
	if err != nil {
		return fmt.Errorf("error obteniendo UUID del jugador: %w", err)
	}
	
	participants, err := s.eventRepo.GetEventParticipants(event.ID)
	if err != nil {
		return fmt.Errorf("error verificando registro previo: %w", err)
	}
	
	for _, participant := range participants {
		if participant.PlayerID == playerUUID {
			return fmt.Errorf("ya estás registrado en este evento")
		}
	}

	return nil
}

// processEventEntryFee procesa el pago de entrada de un evento
func (s *EventService) processEventEntryFee(ctx context.Context, playerID int64, event *models.Event) error {
	if event.EntryFee == 0 {
		return nil // Sin cuota de entrada
	}

	playerUUID, err := s.getPlayerUUIDFromID(playerID)
	if err != nil {
		return fmt.Errorf("error obteniendo UUID del jugador: %w", err)
	}

	// En un sistema real, esto procesaría el pago real
	// Por ahora, registramos la transacción
	s.logger.Info("Procesando cuota de entrada de evento",
		zap.String("player_id", playerUUID.String()),
		zap.String("event_id", event.ID.String()),
		zap.Int("entry_fee", event.EntryFee),
	)

	// Nota: En un sistema real, esto procesaría el pago real
	// Por ejemplo: s.economyService.DeductCurrency(playerUUID, "gold", event.EntryFee)

	return nil
}

// createEventParticipant crea un nuevo participante en un evento
func (s *EventService) createEventParticipant(ctx context.Context, playerID int64, event *models.Event) error {
	playerUUID, err := s.getPlayerUUIDFromID(playerID)
	if err != nil {
		return fmt.Errorf("error obteniendo UUID del jugador: %w", err)
	}

	participant := &models.EventParticipant{
		ID:               uuid.New(),
		EventID:          event.ID,
		PlayerID:         playerUUID,
		Status:           "active",
		RegistrationDate: time.Now(),
		EntryFeePaid:     event.EntryFee == 0,
		CurrentScore:      0,
		TotalScore:        0,
		Rank:              0,
		FinalRank:         0,
		MatchesPlayed:     0,
		MatchesWon:        0,
		MatchesLost:       0,
		MatchesDrawn:      0,
	}

	err = s.eventRepo.CreateEventParticipant(participant)
	if err != nil {
		return fmt.Errorf("error creando participante: %w", err)
	}

	s.logger.Info("Participante creado exitosamente",
		zap.String("player_id", playerUUID.String()),
		zap.String("event_id", event.ID.String()),
		zap.String("participant_id", participant.ID.String()),
	)

	return nil
}

// processResourceReward procesa una recompensa de recursos
func (s *EventService) processResourceReward(ctx context.Context, playerID int64, reward *models.EventReward) error {
	playerUUID, err := s.getPlayerUUIDFromID(playerID)
	if err != nil {
		return fmt.Errorf("error obteniendo UUID del jugador: %w", err)
	}

	// En un sistema real, esto actualizaría los recursos reales del jugador
	// Por ahora, registramos la recompensa
	s.logger.Info("Procesando recompensa de recursos",
		zap.String("player_id", playerUUID.String()),
		zap.String("resource_type", reward.ResourceType),
		zap.Int("quantity", reward.Quantity),
	)

	// Nota: En un sistema real, esto actualizaría los recursos reales del jugador
	// Por ejemplo: s.villageRepo.UpdateVillageResources(village.Village.ID, reward.ResourceType, reward.Quantity)

	return nil
}

// processItemReward procesa una recompensa de items
func (s *EventService) processItemReward(ctx context.Context, playerID int64, reward *models.EventReward) error {
	playerUUID, err := s.getPlayerUUIDFromID(playerID)
	if err != nil {
		return fmt.Errorf("error obteniendo UUID del jugador: %w", err)
	}

	// En un sistema real, esto agregaría el item al inventario del jugador
	// Por ahora, registramos la recompensa
	s.logger.Info("Procesando recompensa de item",
		zap.String("player_id", playerUUID.String()),
		zap.String("item_id", reward.ItemID.String()),
		zap.Int("quantity", reward.Quantity),
	)

	// Nota: En un sistema real, esto agregaría el item al inventario del jugador
	// Por ejemplo: s.inventoryRepo.AddItemToPlayer(playerUUID, *reward.ItemID, reward.Quantity)

	return nil
}