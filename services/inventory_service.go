package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

type InventoryService struct {
	redisService *RedisService
}

type InventoryItem struct {
	ID         string                 `json:"id"`
	PlayerID   int64                  `json:"player_id"`
	ItemType   string                 `json:"item_type"`
	ItemID     string                 `json:"item_id"`
	Quantity   int                    `json:"quantity"`
	Quality    int                    `json:"quality"`
	Attributes map[string]interface{} `json:"attributes"`
	ExpiresAt  *time.Time             `json:"expires_at,omitempty"`
	CreatedAt  time.Time              `json:"created_at"`
}

type InventoryTransaction struct {
	ID        string    `json:"id"`
	PlayerID  int64     `json:"player_id"`
	Type      string    `json:"type"` // "add", "remove", "transfer"
	ItemType  string    `json:"item_type"`
	ItemID    string    `json:"item_id"`
	Quantity  int       `json:"quantity"`
	Reason    string    `json:"reason"`
	Timestamp time.Time `json:"timestamp"`
}

func NewInventoryService(redisService *RedisService) *InventoryService {
	return &InventoryService{
		redisService: redisService,
	}
}

// AddItem agrega un item al inventario temporal
func (s *InventoryService) AddItem(ctx context.Context, playerID int64, itemType, itemID string, quantity, quality int, attributes map[string]interface{}) error {
	// Crear item
	item := &InventoryItem{
		ID:         fmt.Sprintf("%d_%s_%d", playerID, itemID, time.Now().UnixNano()),
		PlayerID:   playerID,
		ItemType:   itemType,
		ItemID:     itemID,
		Quantity:   quantity,
		Quality:    quality,
		Attributes: attributes,
		CreatedAt:  time.Now(),
	}

	// Cachear item
	itemKey := fmt.Sprintf("inventory:item:%s", item.ID)
	err := s.redisService.SetCache(itemKey, item, 24*time.Hour)
	if err != nil {
		return fmt.Errorf("error cacheando item: %v", err)
	}

	// Agregar a lista de items del jugador
	playerKey := fmt.Sprintf("inventory:player:%d", playerID)
	err = s.redisService.client.SAdd(ctx, playerKey, item.ID).Err()
	if err != nil {
		return fmt.Errorf("error agregando item a jugador: %v", err)
	}

	// Configurar expiración
	err = s.redisService.client.Expire(ctx, playerKey, 24*time.Hour).Err()
	if err != nil {
		return fmt.Errorf("error configurando expiración: %v", err)
	}

	// Registrar transacción
	err = s.recordTransaction(ctx, playerID, "add", itemType, itemID, quantity, "item_added")
	if err != nil {
		log.Printf("Error registrando transacción: %v", err)
	}

	return nil
}

// RemoveItem remueve un item del inventario
func (s *InventoryService) RemoveItem(ctx context.Context, playerID int64, itemID string, quantity int) error {
	// Buscar item
	item, err := s.findItem(ctx, playerID, itemID)
	if err != nil {
		return fmt.Errorf("item no encontrado: %v", err)
	}

	if item.Quantity < quantity {
		return fmt.Errorf("cantidad insuficiente: %d disponible, %d solicitado", item.Quantity, quantity)
	}

	// Actualizar cantidad
	item.Quantity -= quantity

	if item.Quantity <= 0 {
		// Eliminar item completamente
		itemKey := fmt.Sprintf("inventory:item:%s", item.ID)
		err = s.redisService.DeleteCache(itemKey)
		if err != nil {
			log.Printf("Error eliminando item: %v", err)
		}

		// Remover de lista del jugador
		playerKey := fmt.Sprintf("inventory:player:%d", playerID)
		err = s.redisService.client.SRem(ctx, playerKey, item.ID).Err()
		if err != nil {
			log.Printf("Error removiendo item de jugador: %v", err)
		}
	} else {
		// Actualizar item
		itemKey := fmt.Sprintf("inventory:item:%s", item.ID)
		err = s.redisService.SetCache(itemKey, item, 24*time.Hour)
		if err != nil {
			log.Printf("Error actualizando item: %v", err)
		}
	}

	// Registrar transacción
	err = s.recordTransaction(ctx, playerID, "remove", item.ItemType, item.ItemID, quantity, "item_removed")
	if err != nil {
		log.Printf("Error registrando transacción: %v", err)
	}

	return nil
}

// GetPlayerInventory obtiene el inventario completo de un jugador
func (s *InventoryService) GetPlayerInventory(ctx context.Context, playerID int64) ([]*InventoryItem, error) {
	// Obtener IDs de items del jugador
	playerKey := fmt.Sprintf("inventory:player:%d", playerID)
	itemIDs, err := s.redisService.client.SMembers(ctx, playerKey).Result()
	if err != nil {
		return nil, fmt.Errorf("error obteniendo items del jugador: %v", err)
	}

	var items []*InventoryItem
	for _, itemID := range itemIDs {
		item, err := s.getItem(ctx, itemID)
		if err != nil {
			log.Printf("Error obteniendo item %s: %v", itemID, err)
			continue
		}
		items = append(items, item)
	}

	return items, nil
}

// GetItem obtiene un item específico
func (s *InventoryService) GetItem(ctx context.Context, playerID int64, itemID string) (*InventoryItem, error) {
	return s.findItem(ctx, playerID, itemID)
}

// TransferItem transfiere un item entre jugadores
func (s *InventoryService) TransferItem(ctx context.Context, fromPlayerID, toPlayerID int64, itemID string, quantity int) error {
	// Remover del jugador origen
	err := s.RemoveItem(ctx, fromPlayerID, itemID, quantity)
	if err != nil {
		return fmt.Errorf("error removiendo item del origen: %v", err)
	}

	// Obtener item para transferir
	item, err := s.findItem(ctx, fromPlayerID, itemID)
	if err != nil {
		// Si no se encuentra, crear uno nuevo
		item = &InventoryItem{
			ItemType:   "unknown",
			ItemID:     itemID,
			Quantity:   quantity,
			Quality:    1,
			Attributes: make(map[string]interface{}),
		}
	}

	// Agregar al jugador destino
	err = s.AddItem(ctx, toPlayerID, item.ItemType, item.ItemID, quantity, item.Quality, item.Attributes)
	if err != nil {
		return fmt.Errorf("error agregando item al destino: %v", err)
	}

	// Registrar transacción de transferencia
	err = s.recordTransaction(ctx, fromPlayerID, "transfer", item.ItemType, item.ItemID, quantity, fmt.Sprintf("transferred_to_%d", toPlayerID))
	if err != nil {
		log.Printf("Error registrando transacción de transferencia: %v", err)
	}

	return nil
}

// GetItemCount obtiene la cantidad de un item específico
func (s *InventoryService) GetItemCount(ctx context.Context, playerID int64, itemID string) (int, error) {
	item, err := s.findItem(ctx, playerID, itemID)
	if err != nil {
		return 0, nil // Item no encontrado, cantidad 0
	}

	return item.Quantity, nil
}

// GetTransactionHistory obtiene el historial de transacciones de un jugador
func (s *InventoryService) GetTransactionHistory(ctx context.Context, playerID int64, limit int) ([]*InventoryTransaction, error) {
	if limit <= 0 {
		limit = 50
	}

	// Obtener transacciones desde Redis
	historyKey := fmt.Sprintf("inventory:history:%d", playerID)
	transactions, err := s.redisService.client.LRange(ctx, historyKey, 0, int64(limit-1)).Result()
	if err != nil {
		return nil, fmt.Errorf("error obteniendo historial: %v", err)
	}

	var history []*InventoryTransaction
	for _, txData := range transactions {
		var transaction InventoryTransaction
		err := json.Unmarshal([]byte(txData), &transaction)
		if err != nil {
			log.Printf("Error deserializando transacción: %v", err)
			continue
		}
		history = append(history, &transaction)
	}

	return history, nil
}

// CleanupExpiredItems limpia items expirados
func (s *InventoryService) CleanupExpiredItems(ctx context.Context) error {
	// Obtener todas las claves de items
	itemKeys, err := s.redisService.GetKeys(ctx, "inventory:item:*")
	if err != nil {
		return fmt.Errorf("error obteniendo items: %v", err)
	}

	now := time.Now()
	cleaned := 0

	for _, itemKey := range itemKeys {
		var item InventoryItem
		err := s.redisService.GetCache(itemKey, &item)
		if err != nil {
			continue
		}

		// Verificar si el item ha expirado
		if item.ExpiresAt != nil && now.After(*item.ExpiresAt) {
			// Eliminar item
			err = s.redisService.DeleteCache(itemKey)
			if err != nil {
				log.Printf("Error eliminando item expirado: %v", err)
				continue
			}

			// Remover de lista del jugador
			playerKey := fmt.Sprintf("inventory:player:%d", item.PlayerID)
			s.redisService.client.SRem(ctx, playerKey, item.ID)

			cleaned++
		}
	}

	log.Printf("Limpieza completada: %d items expirados eliminados", cleaned)
	return nil
}

// findItem busca un item específico en el inventario de un jugador
func (s *InventoryService) findItem(ctx context.Context, playerID int64, itemID string) (*InventoryItem, error) {
	// Obtener IDs de items del jugador
	playerKey := fmt.Sprintf("inventory:player:%d", playerID)
	itemIDs, err := s.redisService.client.SMembers(ctx, playerKey).Result()
	if err != nil {
		return nil, fmt.Errorf("error obteniendo items del jugador: %v", err)
	}

	// Buscar item específico
	for _, id := range itemIDs {
		item, err := s.getItem(ctx, id)
		if err != nil {
			continue
		}

		if item.ItemID == itemID {
			return item, nil
		}
	}

	return nil, fmt.Errorf("item no encontrado")
}

// getItem obtiene un item por su ID
func (s *InventoryService) getItem(ctx context.Context, itemID string) (*InventoryItem, error) {
	itemKey := fmt.Sprintf("inventory:item:%s", itemID)
	var item InventoryItem

	err := s.redisService.GetCache(itemKey, &item)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo item: %v", err)
	}

	return &item, nil
}

// recordTransaction registra una transacción de inventario
func (s *InventoryService) recordTransaction(ctx context.Context, playerID int64, txType, itemType, itemID string, quantity int, reason string) error {
	transaction := &InventoryTransaction{
		ID:        fmt.Sprintf("%d_%d", playerID, time.Now().UnixNano()),
		PlayerID:  playerID,
		Type:      txType,
		ItemType:  itemType,
		ItemID:    itemID,
		Quantity:  quantity,
		Reason:    reason,
		Timestamp: time.Now(),
	}

	// Agregar al historial
	historyKey := fmt.Sprintf("inventory:history:%d", playerID)
	transactionJSON, _ := json.Marshal(transaction)

	err := s.redisService.client.LPush(ctx, historyKey, transactionJSON).Err()
	if err != nil {
		return fmt.Errorf("error agregando transacción: %v", err)
	}

	// Mantener solo las últimas 100 transacciones
	err = s.redisService.client.LTrim(ctx, historyKey, 0, 99).Err()
	if err != nil {
		return fmt.Errorf("error recortando historial: %v", err)
	}

	// Expirar después de 7 días
	err = s.redisService.client.Expire(ctx, historyKey, 7*24*time.Hour).Err()
	if err != nil {
		return fmt.Errorf("error configurando expiración: %v", err)
	}

	return nil
}

// GetInventoryStats obtiene estadísticas del inventario
func (s *InventoryService) GetInventoryStats(ctx context.Context) (map[string]interface{}, error) {
	// Obtener todas las claves de inventario
	playerKeys, err := s.redisService.GetKeys(ctx, "inventory:player:*")
	if err != nil {
		return nil, fmt.Errorf("error obteniendo claves de inventario: %v", err)
	}

	stats := map[string]interface{}{
		"total_players": len(playerKeys),
		"total_items":   0,
		"timestamp":     time.Now(),
		"items_by_type": make(map[string]int),
	}

	// Contar items por tipo
	for _, playerKey := range playerKeys {
		itemIDs, err := s.redisService.client.SMembers(ctx, playerKey).Result()
		if err != nil {
			continue
		}

		for _, itemID := range itemIDs {
			item, err := s.getItem(ctx, itemID)
			if err != nil {
				continue
			}

			stats["total_items"] = stats["total_items"].(int) + item.Quantity

			itemTypeCount := stats["items_by_type"].(map[string]int)[item.ItemType]
			stats["items_by_type"].(map[string]int)[item.ItemType] = itemTypeCount + item.Quantity
		}
	}

	return stats, nil
}
