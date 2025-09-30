package repository

import (
	"database/sql"
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

func (r *ChatRepository) CreateChannel(channel *models.ChatChannel) error {
	channel.ID = uuid.New()
	channel.CreatedAt = time.Now()

	_, err := r.db.Exec(`
		INSERT INTO chat_channels (id, name, description, type, world_id, is_active, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, channel.ID, channel.Name, channel.Description, channel.Type, channel.WorldID, channel.IsActive, channel.CreatedAt)

	return err
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

func (r *ChatRepository) GetChannelByID(channelID uuid.UUID) (*models.ChatChannel, error) {
	var channel models.ChatChannel
	err := r.db.QueryRow(`
		SELECT id, name, description, type, world_id, is_active, created_at
		FROM chat_channels
		WHERE id = $1
	`, channelID).Scan(
		&channel.ID,
		&channel.Name,
		&channel.Description,
		&channel.Type,
		&channel.WorldID,
		&channel.IsActive,
		&channel.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &channel, nil
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
	for _, channel := range models.DefaultChannels {
		// Verificar si el canal ya existe
		existingChannels, err := r.db.Query(`
			SELECT id FROM chat_channels WHERE name = $1
		`, channel.Name)
		if err != nil {
			return err
		}

		if !existingChannels.Next() {
			// Canal no existe, crearlo
			channel.ID = uuid.New()
			channel.CreatedAt = time.Now()

			_, err = r.db.Exec(`
				INSERT INTO chat_channels (id, name, description, type, world_id, is_active, created_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7)
			`, channel.ID, channel.Name, channel.Description, channel.Type, channel.WorldID, channel.IsActive, channel.CreatedAt)

			if err != nil {
				return err
			}
		}
		existingChannels.Close()
	}

	return nil
}
