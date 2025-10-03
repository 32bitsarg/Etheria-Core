package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"server-backend/auth"
	"server-backend/config"
	"server-backend/handlers"
	"server-backend/middleware"
	"server-backend/repository"
	"server-backend/routes"
	"server-backend/services"
	"server-backend/websocket"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

func main() {
	// Inicializar logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	logger.Info("🚀 Iniciando Etheria Core Server")

	// Cargar configuración
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatal("Error cargando configuración", zap.Error(err))
	}

	// Conectar a la base de datos
	connectionString := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.DBName,
		cfg.Database.SSLMode,
	)

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		logger.Fatal("Error conectando a la base de datos", zap.Error(err))
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		logger.Fatal("Error haciendo ping a la base de datos", zap.Error(err))
	}
	logger.Info("Conectado a la base de datos exitosamente")

	// Inicializar servicios
	services, constructionService, chatService := initializeServices(db, cfg, logger)

	// Inicializar repositorios
	repos := initializeRepositories(db, logger)

	// Inicializar handlers
	handlers := initializeHandlers(repos, services, constructionService, chatService, logger)

	// Inicializar middleware
	authMiddleware := middleware.NewAuthMiddleware(services.JWT, repos.Player, logger)

	// Configurar router Gin
	r := gin.Default()

	// Configurar todas las rutas
	routes.SetupAllRoutes(r, handlers, repos, services, authMiddleware, logger)

	// Iniciar servicios en background
	startBackgroundServices(services, logger)

	// Configurar servidor HTTP
	srv := &http.Server{
		Addr:         fmt.Sprintf("0.0.0.0:%d", cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Canal para manejar señales de terminación
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Iniciar servidor en goroutine
	go func() {
		logger.Info("Iniciando servidor", zap.Int("port", cfg.Server.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Error iniciando servidor", zap.Error(err))
		}
	}()

	// Esperar señal de terminación
	<-stop
	logger.Info("Cerrando servidor...")

	// Detener servicios
	stopBackgroundServices(services, logger)

	// Dar tiempo para que las conexiones se cierren
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Error cerrando servidor", zap.Error(err))
	}

	logger.Info("Servidor cerrado exitosamente")
}

// initializeServices inicializa todos los servicios
func initializeServices(db *sql.DB, cfg *config.Config, logger *zap.Logger) (*routes.Services, *services.ConstructionService, *services.ChatService) {
	// Servicios básicos
	jwtManager := auth.NewJWTManager(cfg.JWT.SecretKey, cfg.JWT.TokenDuration)
	redisService, err := services.NewRedisService(cfg, logger)
	if err != nil {
		logger.Fatal("Error inicializando Redis", zap.Error(err))
	}

	logger.Info("Redis service inicializado")

	// Repositorios necesarios para servicios
	villageRepo := repository.NewVillageRepository(db, logger)
	buildingConfigRepo := repository.NewBuildingConfigRepository(db, logger)
	researchRepo := repository.NewResearchRepository(db, logger)
	allianceRepo := repository.NewAllianceRepository(db, logger)
	chatRepo := repository.NewChatRepository(db, logger)
	unitRepo := repository.NewUnitRepository(db, logger)

	// WebSocket Manager
	wsManager := websocket.NewManager(chatRepo, villageRepo, unitRepo, logger, redisService)

	// Servicios de dominio
	resourceService := services.NewResourceService(villageRepo, buildingConfigRepo, logger, redisService)
	constructionService := services.NewConstructionService(villageRepo, buildingConfigRepo, researchRepo, allianceRepo, redisService, logger, cfg.JWT.SecretKey)
	chatService := services.NewChatService(chatRepo, redisService, logger)

	// Configurar WebSocket en servicios
	resourceService.SetWebSocketManager(wsManager)
	constructionService.SetWebSocketManager(wsManager)

	return &routes.Services{
		Resource: resourceService,
		JWT:      jwtManager,
		Redis:    redisService,
		Chat:     chatService,
	}, constructionService, chatService
}

// initializeRepositories inicializa todos los repositorios
func initializeRepositories(db *sql.DB, logger *zap.Logger) *routes.Repositories {
	return &routes.Repositories{
		Player:   repository.NewPlayerRepository(db, logger),
		Village:  repository.NewVillageRepository(db, logger),
		Alliance: repository.NewAllianceRepository(db, logger),
		Unit:     repository.NewUnitRepository(db, logger),
	}
}

// initializeHandlers inicializa todos los handlers
func initializeHandlers(repos *routes.Repositories, services *routes.Services, constructionService *services.ConstructionService, chatService *services.ChatService, logger *zap.Logger) *routes.Handlers {
	// Usar repositorios existentes (con db válido) en lugar de crear nuevos
	return &routes.Handlers{
		Auth:     handlers.NewAuthHandler(repos.Player, services.JWT, logger, services.Redis, repos.Village),
		Village:  handlers.NewVillageHandler(repos.Village, constructionService, logger),
		Chat:     handlers.NewChatHandler(chatService, logger),
		Alliance: handlers.NewAllianceHandler(repos.Alliance, logger),
		Unit:     handlers.NewUnitHandler(repos.Unit, repos.Village, logger),
	}
}

// startBackgroundServices inicia servicios en background
func startBackgroundServices(services *routes.Services, logger *zap.Logger) {
	// Iniciar SyncManager del ChatService
	if services.Chat != nil {
		ctx := context.Background()
		if err := services.Chat.StartSyncManager(ctx); err != nil {
			logger.Error("Error iniciando SyncManager del chat", zap.Error(err))
		} else {
			logger.Info("✅ SyncManager del chat iniciado exitosamente")
		}
	}

	// Iniciar generación de recursos
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				logger.Info("Ejecutando ciclo de generación de recursos híbrido")
				// Aquí iría la lógica de generación de recursos
			}
		}
	}()

	logger.Info("Servicios en background iniciados")
}

// stopBackgroundServices detiene servicios en background
func stopBackgroundServices(services *routes.Services, logger *zap.Logger) {
	logger.Info("Deteniendo servicios en background...")
	
	// Detener SyncManager del ChatService
	if services.Chat != nil {
		if err := services.Chat.StopSyncManager(); err != nil {
			logger.Error("Error deteniendo SyncManager del chat", zap.Error(err))
		} else {
			logger.Info("✅ SyncManager del chat detenido exitosamente")
		}
	}
	
	logger.Info("Servicios en background detenidos")
}
