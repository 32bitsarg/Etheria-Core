package repository

import (
	"database/sql"
	"fmt"
	"server-backend/models"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ChatRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewChatRepository(db *sql.DB, logger *zap.Logger) *ChatRepository {
	return &ChatRepository{
		db:     db,
		logger: logger,
	}
}

func (r *ChatRepository) GetChannels() ([]*models.ChatChannel, error) {
	rows, err := r.db.Query(`
		SELECT id, name, description, type, world_id, is_active, created_at
		FROM chat_channels
		WHERE is_active = true
		ORDER BY created_at ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var channels []*models.ChatChannel
	for rows.Next() {
		var channel models.ChatChannel
		err := rows.Scan(
			&channel.ID,
			&channel.Name,
			&channel.Description,
			&channel.Type,
			&channel.WorldID,
			&channel.IsActive,
			&channel.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		channels = append(channels, &channel)
	}
	return channels, nil
}

func (r *ChatRepository) SaveMessage(message *models.ChatMessage) error {
	message.ID = uuid.New()
	message.CreatedAt = time.Now()

	_, err := r.db.Exec(`
		INSERT INTO chat_messages (id, channel_id, player_id, username, message, type, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, message.ID, message.ChannelID, message.PlayerID, message.Username, message.Message, message.Type, message.CreatedAt)

	return err
}

func (r *ChatRepository) GetChannelMessages(channelID uuid.UUID, limit int) ([]*models.ChatMessage, error) {
	if limit <= 0 {
		limit = 50
	}

	rows, err := r.db.Query(`
		SELECT id, channel_id, player_id, username, message, type, created_at
		FROM chat_messages
		WHERE channel_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, channelID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*models.ChatMessage
	for rows.Next() {
		var message models.ChatMessage
		err := rows.Scan(
			&message.ID,
			&message.ChannelID,
			&message.PlayerID,
			&message.Username,
			&message.Message,
			&message.Type,
			&message.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		messages = append(messages, &message)
	}

	// Invertir el orden para mostrar los más antiguos primero
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

func (r *ChatRepository) AddChannelMember(channelID, playerID uuid.UUID, username string) error {
	memberID := uuid.New()
	_, err := r.db.Exec(`
		INSERT INTO chat_channel_members (id, channel_id, player_id, username, is_admin, joined_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (channel_id, player_id) DO NOTHING
	`, memberID, channelID, playerID, username, false, time.Now())

	return err
}

func (r *ChatRepository) RemoveChannelMember(channelID, playerID uuid.UUID) error {
	_, err := r.db.Exec(`
		DELETE FROM chat_channel_members
		WHERE channel_id = $1 AND player_id = $2
	`, channelID, playerID)

	return err
}

func (r *ChatRepository) GetChannelMembers(channelID uuid.UUID) ([]*models.ChatChannelMember, error) {
	rows, err := r.db.Query(`
		SELECT id, channel_id, player_id, username, is_admin, joined_at
		FROM chat_channel_members
		WHERE channel_id = $1
		ORDER BY joined_at ASC
	`, channelID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []*models.ChatChannelMember
	for rows.Next() {
		var member models.ChatChannelMember
		err := rows.Scan(
			&member.ID,
			&member.ChannelID,
			&member.PlayerID,
			&member.Username,
			&member.IsAdmin,
			&member.JoinedAt,
		)
		if err != nil {
			return nil, err
		}
		members = append(members, &member)
	}
	return members, nil
}

func (r *ChatRepository) IsChannelMember(channelID, playerID uuid.UUID) (bool, error) {
	var exists bool
	err := r.db.QueryRow(`
		SELECT EXISTS(SELECT 1 FROM chat_channel_members WHERE channel_id = $1 AND player_id = $2)
	`, channelID, playerID).Scan(&exists)

	return exists, err
}

func (r *ChatRepository) InitializeDefaultChannels() error {
	// Canales predefinidos con valores específicos
	defaultChannels := []struct {
		Name        string
		Description string
		Type        string
		MaxMembers  int
	}{
		{
			Name:        "Global",
			Description: "Chat global para todos los jugadores",
			Type:        "global",
			MaxMembers:  10000,
		},
		{
			Name:        "Ayuda",
			Description: "Canal de ayuda para nuevos jugadores",
			Type:        "help",
			MaxMembers:  1000,
		},
		{
			Name:        "Comercio",
			Description: "Canal para intercambios y comercio",
			Type:        "trade",
			MaxMembers:  5000,
		},
	}

	for _, channelData := range defaultChannels {
		// Verificar si el canal ya existe
		var exists bool
		err := r.db.QueryRow(`
			SELECT EXISTS(SELECT 1 FROM chat_channels WHERE name = $1)
		`, channelData.Name).Scan(&exists)

		if err != nil {
			return fmt.Errorf("error verificando canal %s: %v", channelData.Name, err)
		}

		if !exists {
			// Canal no existe, crearlo
			channelID := uuid.New()
			createdAt := time.Now()

			_, err = r.db.Exec(`
				INSERT INTO chat_channels (id, name, description, type, world_id, is_active, max_members, created_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			`, channelID, channelData.Name, channelData.Description, channelData.Type, nil, true, channelData.MaxMembers, createdAt)

			if err != nil {
				return fmt.Errorf("error creando canal %s: %v", channelData.Name, err)
			}

			r.logger.Info("Canal creado",
				zap.String("name", channelData.Name),
				zap.String("type", channelData.Type),
				zap.Int("max_members", channelData.MaxMembers))
		} else {
			r.logger.Debug("Canal ya existe", zap.String("name", channelData.Name))
		}
	}

	return nil
}

// GetAllChannels obtiene todos los canales desde la base de datos
func (r *ChatRepository) GetAllChannels() ([]*models.ChatChannel, error) {
	rows, err := r.db.Query(`
		SELECT id, name, description, type, world_id, is_active, max_members, created_at
		FROM chat_channels
		WHERE is_active = true
		ORDER BY created_at ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var channels []*models.ChatChannel
	for rows.Next() {
		var channel models.ChatChannel
		var worldID sql.NullString

		err := rows.Scan(
			&channel.ID,
			&channel.Name,
			&channel.Description,
			&channel.Type,
			&worldID,
			&channel.IsActive,
			&channel.MaxMembers,
			&channel.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if worldID.Valid {
			worldUUID, err := uuid.Parse(worldID.String)
			if err == nil {
				channel.WorldID = &worldUUID
			}
		}

		channels = append(channels, &channel)
	}

	return channels, nil
}

// GetChannelByID obtiene un canal específico por ID
func (r *ChatRepository) GetChannelByID(channelID string) (*models.ChatChannel, error) {
	var channel models.ChatChannel
	var worldID sql.NullString

	err := r.db.QueryRow(`
		SELECT id, name, description, type, world_id, is_active, max_members, created_at
		FROM chat_channels
		WHERE id = $1 AND is_active = true
	`, channelID).Scan(
		&channel.ID,
		&channel.Name,
		&channel.Description,
		&channel.Type,
		&worldID,
		&channel.IsActive,
		&channel.MaxMembers,
		&channel.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	if worldID.Valid {
		worldUUID, err := uuid.Parse(worldID.String)
		if err == nil {
			channel.WorldID = &worldUUID
		}
	}

	return &channel, nil
}

// GetChannelByName obtiene un canal específico por nombre (case insensitive)
func (r *ChatRepository) GetChannelByName(name string) (*models.ChatChannel, error) {
	var channel models.ChatChannel
	var worldID sql.NullString

	err := r.db.QueryRow(`
		SELECT id, name, description, type, world_id, is_active, max_members, created_at
		FROM chat_channels
		WHERE LOWER(name) = LOWER($1) AND is_active = true
	`, name).Scan(
		&channel.ID,
		&channel.Name,
		&channel.Description,
		&channel.Type,
		&worldID,
		&channel.IsActive,
		&channel.MaxMembers,
		&channel.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	if worldID.Valid {
		worldUUID, err := uuid.Parse(worldID.String)
		if err == nil {
			channel.WorldID = &worldUUID
		}
	}

	return &channel, nil
}

// CreateChannel crea un nuevo canal en la base de datos
func (r *ChatRepository) CreateChannel(channel *models.ChatChannel) error {
	channel.ID = uuid.New()
	channel.CreatedAt = time.Now()

	_, err := r.db.Exec(`
		INSERT INTO chat_channels (id, name, description, type, world_id, is_active, max_members, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, channel.ID, channel.Name, channel.Description, channel.Type, channel.WorldID, channel.IsActive, channel.MaxMembers, channel.CreatedAt)

	return err
}

// UpdateChannel actualiza un canal existente
func (r *ChatRepository) UpdateChannel(channel *models.ChatChannel) error {
	_, err := r.db.Exec(`
		UPDATE chat_channels 
		SET name = $2, description = $3, type = $4, world_id = $5, is_active = $6, max_members = $7
		WHERE id = $1
	`, channel.ID, channel.Name, channel.Description, channel.Type, channel.WorldID, channel.IsActive, channel.MaxMembers)

	return err
}

// DeleteChannel marca un canal como inactivo
func (r *ChatRepository) DeleteChannel(channelID string) error {
	_, err := r.db.Exec(`
		UPDATE chat_channels 
		SET is_active = false
		WHERE id = $1
	`, channelID)

	return err
}
