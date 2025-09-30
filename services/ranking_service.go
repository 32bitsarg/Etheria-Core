package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"server-backend/models"
	"server-backend/repository"
)

type RankingService struct {
	rankingRepo  *repository.RankingRepository
	redisService *RedisService
}

type RankingEntry struct {
	PlayerID  int       `json:"player_id"`
	Username  string    `json:"username"`
	Score     int       `json:"score"`
	Rank      int       `json:"rank"`
	UpdatedAt time.Time `json:"updated_at"`
}

type RankingStats struct {
	TotalPlayers int       `json:"total_players"`
	TopScore     int       `json:"top_score"`
	AvgScore     float64   `json:"avg_score"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func NewRankingService(rankingRepo *repository.RankingRepository, redisService *RedisService) *RankingService {
	return &RankingService{
		rankingRepo:  rankingRepo,
		redisService: redisService,
	}
}

// UpdatePlayerScore actualiza el score de un jugador y refresca el cache
func (s *RankingService) UpdatePlayerScore(ctx context.Context, playerID int, categoryID int, score int) error {
	// Actualizar en base de datos
	entry := &models.RankingEntry{
		EntityType:  "player",
		EntityID:    playerID,
		CategoryID:  categoryID,
		Score:       score,
		LastUpdated: time.Now(),
	}
	err := s.rankingRepo.UpdateRankingEntry(entry)
	if err != nil {
		return fmt.Errorf("error actualizando score: %v", err)
	}

	// Invalidar cache de rankings
	_ = s.invalidateRankingCache(ctx)
	return nil
}

// GetTopPlayers obtiene los mejores jugadores desde cache o BD
func (s *RankingService) GetTopPlayers(ctx context.Context, categoryID int, limit int) ([]*RankingEntry, error) {
	if limit <= 0 {
		limit = 100
	}

	cacheKey := fmt.Sprintf("ranking:top:%d:%d", categoryID, limit)
	var entries []*RankingEntry
	err := s.redisService.GetCache(cacheKey, &entries)
	if err == nil {
		return entries, nil
	}

	rankings, err := s.rankingRepo.GetRankingEntries(categoryID, nil, limit)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo top players: %v", err)
	}

	entries = make([]*RankingEntry, len(rankings))
	for i, ranking := range rankings {
		entries[i] = &RankingEntry{
			PlayerID:  ranking.EntityID,
			Username:  ranking.EntityName,
			Score:     ranking.Score,
			Rank:      ranking.Position,
			UpdatedAt: ranking.LastUpdated,
		}
	}

	_ = s.redisService.SetCache(cacheKey, entries, 5*time.Minute)
	return entries, nil
}

// GetPlayerRank obtiene el ranking de un jugador específico
func (s *RankingService) GetPlayerRank(ctx context.Context, playerID int, categoryID int) (*RankingEntry, error) {
	cacheKey := fmt.Sprintf("ranking:player:%d:%d", playerID, categoryID)
	var entry RankingEntry
	err := s.redisService.GetCache(cacheKey, &entry)
	if err == nil {
		return &entry, nil
	}

	rankings, err := s.rankingRepo.GetRankingEntries(categoryID, nil, 0)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo ranking: %v", err)
	}

	for _, ranking := range rankings {
		if ranking.EntityID == playerID {
			entry = RankingEntry{
				PlayerID:  ranking.EntityID,
				Username:  ranking.EntityName,
				Score:     ranking.Score,
				Rank:      ranking.Position,
				UpdatedAt: ranking.LastUpdated,
			}
			_ = s.redisService.SetCache(cacheKey, entry, 2*time.Minute)
			return &entry, nil
		}
	}

	return nil, fmt.Errorf("jugador no encontrado en ranking")
}

// GetRankingStats obtiene estadísticas del ranking
func (s *RankingService) GetRankingStats(ctx context.Context) (*RankingStats, error) {
	cacheKey := "ranking:stats"
	var stats RankingStats
	err := s.redisService.GetCache(cacheKey, &stats)
	if err == nil {
		return &stats, nil
	}

	summary, err := s.rankingRepo.GetStatisticsSummary()
	if err != nil {
		return nil, fmt.Errorf("error obteniendo estadísticas: %v", err)
	}

	stats = RankingStats{
		TotalPlayers: int(summary.TotalPlayers),
		TopScore:     0, // No hay campo directo, se puede calcular si es necesario
		AvgScore:     0, // No hay campo directo, se puede calcular si es necesario
		UpdatedAt:    summary.LastUpdated,
	}

	_ = s.redisService.SetCache(cacheKey, stats, 10*time.Minute)
	return &stats, nil
}

// GetRankingByCategory obtiene rankings por categoría
func (s *RankingService) GetRankingByCategory(ctx context.Context, categoryID int, limit int) ([]*RankingEntry, error) {
	return s.GetTopPlayers(ctx, categoryID, limit)
}

// invalidateRankingCache invalida todos los caches de ranking
func (s *RankingService) invalidateRankingCache(ctx context.Context) error {
	// Obtener todas las claves de ranking
	keys, err := s.redisService.GetKeys(ctx, "ranking:*")
	if err != nil {
		return fmt.Errorf("error obteniendo claves de ranking: %v", err)
	}

	// Eliminar cada clave
	for _, key := range keys {
		err = s.redisService.DeleteCache(key)
		if err != nil {
			log.Printf("Error eliminando cache %s: %v", key, err)
		}
	}

	return nil
}

// RefreshRankingCache refresca todos los caches de ranking
func (s *RankingService) RefreshRankingCache(ctx context.Context) error {
	// Invalidar cache existente
	err := s.invalidateRankingCache(ctx)
	if err != nil {
		return fmt.Errorf("error invalidando cache: %v", err)
	}

	// Pre-cargar rankings más populares
	_, err = s.GetTopPlayers(ctx, 0, 100)
	if err != nil {
		log.Printf("Error pre-cargando top players: %v", err)
	}

	_, err = s.GetRankingStats(ctx)
	if err != nil {
		log.Printf("Error pre-cargando estadísticas: %v", err)
	}

	return nil
}
