package services

import (
	"context"
	"fmt"
	"time"

	"server-backend/config"

	"github.com/redis/go-redis/v9"
)

type RateLimitService struct {
	redisService *RedisService
	config       *config.Config
}

type RateLimitInfo struct {
	Remaining int           `json:"remaining"`
	ResetTime time.Time     `json:"reset_time"`
	Limit     int           `json:"limit"`
	Window    time.Duration `json:"window"`
}

func NewRateLimitService(redisService *RedisService, config *config.Config) *RateLimitService {
	return &RateLimitService{
		redisService: redisService,
		config:       config,
	}
}

// CheckRateLimit verifica si una solicitud está dentro del límite de rate
func (r *RateLimitService) CheckRateLimit(ctx context.Context, key string, limit int, window time.Duration) (*RateLimitInfo, error) {
	// Crear clave única para el rate limit
	rateKey := fmt.Sprintf("rate_limit:%s", key)

	// Obtener el tiempo actual
	now := time.Now()

	// Calcular el tiempo de inicio de la ventana
	windowStart := now.Add(-window)

	// Obtener todas las solicitudes en la ventana actual
	requests, err := r.redisService.client.ZRangeByScore(ctx, rateKey,
		&redis.ZRangeBy{
			Min: fmt.Sprintf("%d", windowStart.Unix()),
			Max: fmt.Sprintf("%d", now.Unix()),
		}).Result()

	if err != nil {
		return nil, fmt.Errorf("error obteniendo solicitudes: %w", err)
	}

	// Contar solicitudes en la ventana
	currentCount := len(requests)

	// Verificar si se excedió el límite
	if currentCount >= limit {
		// Calcular tiempo de reset
		if len(requests) > 0 {
			oldestRequest := requests[0]
			// Convertir string a int64 para el timestamp
			var oldestTime int64
			fmt.Sscanf(oldestRequest, "%d", &oldestTime)
			resetTime := time.Unix(oldestTime, 0).Add(window)

			return &RateLimitInfo{
				Remaining: 0,
				ResetTime: resetTime,
				Limit:     limit,
				Window:    window,
			}, nil
		}
	}

	// Agregar nueva solicitud
	err = r.redisService.client.ZAdd(ctx, rateKey,
		redis.Z{
			Score:  float64(now.Unix()),
			Member: fmt.Sprintf("%d", now.UnixNano()),
		}).Err()

	if err != nil {
		return nil, fmt.Errorf("error agregando solicitud: %w", err)
	}

	// Configurar expiración para limpiar automáticamente
	err = r.redisService.client.Expire(ctx, rateKey, window).Err()
	if err != nil {
		return nil, fmt.Errorf("error configurando expiración: %w", err)
	}

	// Calcular tiempo de reset
	resetTime := now.Add(window)

	return &RateLimitInfo{
		Remaining: limit - currentCount - 1,
		ResetTime: resetTime,
		Limit:     limit,
		Window:    window,
	}, nil
}

// CheckIPRateLimit verifica rate limit por IP
func (r *RateLimitService) CheckIPRateLimit(ctx context.Context, ip string) (*RateLimitInfo, error) {
	// Configuración por defecto: 100 solicitudes por minuto
	limit := 100
	window := 1 * time.Minute

	// Verificar configuración personalizada
	if r.config.RateLimit.IPLimit > 0 {
		limit = r.config.RateLimit.IPLimit
	}
	if r.config.RateLimit.IPWindow > 0 {
		window = time.Duration(r.config.RateLimit.IPWindow) * time.Second
	}

	key := fmt.Sprintf("ip:%s", ip)
	return r.CheckRateLimit(ctx, key, limit, window)
}

// CheckUserRateLimit verifica rate limit por usuario
func (r *RateLimitService) CheckUserRateLimit(ctx context.Context, userID string) (*RateLimitInfo, error) {
	// Configuración por defecto: 1000 solicitudes por hora
	limit := 1000
	window := 1 * time.Hour

	// Verificar configuración personalizada
	if r.config.RateLimit.UserLimit > 0 {
		limit = r.config.RateLimit.UserLimit
	}
	if r.config.RateLimit.UserWindow > 0 {
		window = time.Duration(r.config.RateLimit.UserWindow) * time.Second
	}

	key := fmt.Sprintf("user:%s", userID)
	return r.CheckRateLimit(ctx, key, limit, window)
}

// CheckEndpointRateLimit verifica rate limit por endpoint
func (r *RateLimitService) CheckEndpointRateLimit(ctx context.Context, endpoint string, userID string) (*RateLimitInfo, error) {
	// Configuración por defecto: 50 solicitudes por minuto por endpoint
	limit := 50
	window := 1 * time.Minute

	// Verificar configuración personalizada
	if r.config.RateLimit.EndpointLimit > 0 {
		limit = r.config.RateLimit.EndpointLimit
	}
	if r.config.RateLimit.EndpointWindow > 0 {
		window = time.Duration(r.config.RateLimit.EndpointWindow) * time.Second
	}

	key := fmt.Sprintf("endpoint:%s:%s", endpoint, userID)
	return r.CheckRateLimit(ctx, key, limit, window)
}

// GetRateLimitInfo obtiene información de rate limit sin incrementar el contador
func (r *RateLimitService) GetRateLimitInfo(ctx context.Context, key string, limit int, window time.Duration) (*RateLimitInfo, error) {
	rateKey := fmt.Sprintf("rate_limit:%s", key)
	now := time.Now()
	windowStart := now.Add(-window)

	requests, err := r.redisService.client.ZRangeByScore(ctx, rateKey,
		&redis.ZRangeBy{
			Min: fmt.Sprintf("%d", windowStart.Unix()),
			Max: fmt.Sprintf("%d", now.Unix()),
		}).Result()

	if err != nil {
		return nil, fmt.Errorf("error obteniendo solicitudes: %w", err)
	}

	currentCount := len(requests)
	resetTime := now.Add(window)

	return &RateLimitInfo{
		Remaining: limit - currentCount,
		ResetTime: resetTime,
		Limit:     limit,
		Window:    window,
	}, nil
}

// ResetRateLimit resetea el rate limit para una clave específica
func (r *RateLimitService) ResetRateLimit(ctx context.Context, key string) error {
	rateKey := fmt.Sprintf("rate_limit:%s", key)
	return r.redisService.client.Del(ctx, rateKey).Err()
}

// GetRateLimitStats obtiene estadísticas de rate limiting
func (r *RateLimitService) GetRateLimitStats(ctx context.Context) (map[string]interface{}, error) {
	// Obtener todas las claves de rate limit
	keys, err := r.redisService.GetKeys(ctx, "rate_limit:*")
	if err != nil {
		return nil, fmt.Errorf("error obteniendo claves de rate limit: %w", err)
	}

	stats := map[string]interface{}{
		"total_keys":    len(keys),
		"timestamp":     time.Now(),
		"active_limits": make(map[string]int),
	}

	// Contar solicitudes por tipo
	for _, key := range keys {
		count, err := r.redisService.client.ZCard(ctx, key).Result()
		if err != nil {
			continue
		}

		// Extraer tipo de rate limit de la clave
		limitType := "unknown"
		if len(key) > 12 { // "rate_limit:" tiene 12 caracteres
			limitType = key[12:]
		}

		stats["active_limits"].(map[string]int)[limitType] = int(count)
	}

	return stats, nil
}

// CleanupExpiredRateLimits limpia rate limits expirados
func (r *RateLimitService) CleanupExpiredRateLimits(ctx context.Context) error {
	// Obtener todas las claves de rate limit
	keys, err := r.redisService.GetKeys(ctx, "rate_limit:*")
	if err != nil {
		return fmt.Errorf("error obteniendo claves de rate limit: %w", err)
	}

	cleaned := 0

	for _, key := range keys {
		// Verificar si la clave tiene TTL
		ttl, err := r.redisService.client.TTL(ctx, key).Result()
		if err != nil {
			continue
		}

		// Si no tiene TTL o está expirada, eliminarla
		if ttl == -1 || ttl == 0 {
			err = r.redisService.client.Del(ctx, key).Err()
			if err == nil {
				cleaned++
			}
		}
	}

	return nil
}

// IsRateLimited verifica si una clave está siendo rate limited
func (r *RateLimitService) IsRateLimited(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	info, err := r.GetRateLimitInfo(ctx, key, limit, window)
	if err != nil {
		return false, err
	}

	return info.Remaining <= 0, nil
}

// GetRemainingRequests obtiene el número de solicitudes restantes
func (r *RateLimitService) GetRemainingRequests(ctx context.Context, key string, limit int, window time.Duration) (int, error) {
	info, err := r.GetRateLimitInfo(ctx, key, limit, window)
	if err != nil {
		return 0, err
	}

	return info.Remaining, nil
}
