package main

import (
	"context"
	"database/sql"
	"encoding/json"
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
	"server-backend/services"
	"server-backend/websocket"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

func main() {
	// Inicializar logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Cargar configuración
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatal("Error cargando configuración", zap.Error(err))
	}

	// Conectar a la base de datos
	databaseURL := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.Password, cfg.Database.DBName, cfg.Database.SSLMode)

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		logger.Fatal("Error conectando a la base de datos", zap.Error(err))
	}
	defer db.Close()

	// Verificar conexión a la base de datos
	if err := db.Ping(); err != nil {
		logger.Fatal("Error verificando conexión a la base de datos", zap.Error(err))
	}

	logger.Info("Conectado a la base de datos exitosamente")

	// Inicializar Redis (opcional)
	var redisService *services.RedisService
	redisService, err = services.NewRedisService(cfg, logger)
	if err != nil {
		logger.Warn("Redis no disponible, ejecutando sin funcionalidades de tiempo real", zap.Error(err))
		redisService = nil
	} else {
		defer redisService.Close()
		logger.Info("Conectado a Redis exitosamente")
	}

	// Inicializar repositorios
	playerRepo := repository.NewPlayerRepository(db, logger)
	villageRepo := repository.NewVillageRepository(db, logger)
	worldRepo := repository.NewWorldRepository(db, logger)
	unitRepo := repository.NewUnitRepository(db, logger)
	chatRepo := repository.NewChatRepository(db, logger)
	allianceRepo := repository.NewAllianceRepository(db, logger)
	researchRepo := repository.NewResearchRepository(db, logger)
	heroRepo := repository.NewHeroRepository(db, logger)
	battleRepo := repository.NewBattleRepository(db, logger)
	economyRepo := repository.NewEconomyRepository(db, logger)
	rankingRepo := repository.NewRankingRepository(db, logger)
	achievementRepo := repository.NewAchievementRepository(db, logger)
	questRepo := repository.NewQuestRepository(db, logger)
	eventRepo := repository.NewEventRepository(db, logger)
	titleRepo := repository.NewTitleRepository(db, logger)
	tradeRepo := repository.NewTradeRepository(db)
	notificationRepo := repository.NewNotificationRepository(db)
	buildingConfigRepo := repository.NewBuildingConfigRepository(db, logger)
	currencyRepo := repository.NewCurrencyRepository(db, logger)

	// Inicializar servicios
	resourceService := services.NewResourceService(villageRepo, buildingConfigRepo, logger, redisService)
	achievementService := services.NewAchievementService(achievementRepo, playerRepo, nil, logger, redisService)
	questService := services.NewQuestService(questRepo, playerRepo, nil, logger)
	titleService := services.NewTitleService(titleRepo, playerRepo, nil, logger)
	notificationService := services.NewNotificationService(notificationRepo, playerRepo, nil, logger, redisService)
	rateLimitService := services.NewRateLimitService(redisService, cfg)
	configCacheService := services.NewConfigCacheService(redisService)
	constructionService := services.NewConstructionService(villageRepo, buildingConfigRepo, redisService, logger, cfg.TimeZone)
	gameMechanicsService := services.NewGameMechanicsService(villageRepo, playerRepo, logger)
	battleService := services.NewBattleService(battleRepo, villageRepo, unitRepo, logger, redisService)
	chatService := services.NewChatService(chatRepo, redisService, logger)
	rankingService := services.NewRankingService(rankingRepo, redisService)
	worldService := services.NewWorldService(worldRepo, playerRepo, villageRepo, allianceRepo, battleRepo, economyRepo, logger)
	economyService := services.NewEconomyService(economyRepo, playerRepo, villageRepo, nil, logger)
	inventoryService := services.NewInventoryService(redisService)
	researchService := services.NewResearchService(researchRepo, villageRepo, economyRepo, battleService, logger, redisService)

	// Crear WebSocket Manager
	wsManager := websocket.NewManager(chatRepo, villageRepo, unitRepo, logger, redisService)

	// Crear EventService después de wsManager
	eventService := services.NewEventService(eventRepo, playerRepo, wsManager, redisService, economyRepo, titleRepo, villageRepo, logger)

	// Inicializar JWT Manager
	jwtManager := auth.NewJWTManager(cfg.JWT.SecretKey, cfg.JWT.TokenDuration)

	// Configurar WebSocket Manager en servicios (usando métodos públicos)
	battleService.SetWebSocketManager(wsManager)
	economyService.SetWebSocketManager(wsManager)
	achievementService.SetWebSocketManager(wsManager)
	questService.SetWebSocketManager(wsManager)
	eventService.SetWebSocketManager(wsManager)
	titleService.SetWebSocketManager(wsManager)
	notificationService.SetWebSocketManager(wsManager)

	// Inicializar handlers
	authHandler := handlers.NewAuthHandler(playerRepo, jwtManager, logger, redisService, villageRepo)
	villageHandler := handlers.NewVillageHandler(villageRepo, constructionService, logger)
	worldHandler := handlers.NewWorldHandler(worldRepo, playerRepo, villageRepo, allianceRepo, battleRepo, economyRepo, logger)
	worldClientHandler := handlers.NewWorldClientHandler(worldRepo, playerRepo, villageRepo, allianceRepo, battleRepo, economyRepo, worldService, logger)
	unitHandler := handlers.NewUnitHandler(unitRepo, villageRepo, logger)
	chatHandler := handlers.NewChatHandler(chatService, logger)
	allianceHandler := handlers.NewAllianceHandler(allianceRepo, logger)
	researchHandler := handlers.NewResearchHandler(researchRepo, villageRepo, researchService, logger)
	heroHandler := handlers.NewHeroHandler(heroRepo, logger)
	battleHandler := handlers.NewBattleHandler(battleRepo, villageRepo, unitRepo, logger, redisService)
	economyHandler := handlers.NewEconomyHandler(economyService, economyRepo, logger)
	rankingHandler := handlers.NewRankingHandler(rankingRepo, logger)
	achievementHandler := handlers.NewAchievementHandler(achievementRepo, logger)
	questHandler := handlers.NewQuestHandler(questRepo, questService)
	eventHandler := handlers.NewEventHandler(eventRepo)

	// Crear TitleService y TitleHandler
	titleService = services.NewTitleService(titleRepo, playerRepo, wsManager, logger)
	titleHandler := handlers.NewTitleHandler(titleService, logger)

	tradeHandler := handlers.NewTradeHandler(tradeRepo, villageRepo, logger)
	redisHandler := handlers.NewRedisHandler(redisService, constructionService, chatService, rankingService, eventService, battleService, rateLimitService, inventoryService, configCacheService, villageRepo, logger)
	adminHandler := handlers.NewAdminHandler(playerRepo, worldRepo, villageRepo, battleRepo, economyRepo, allianceRepo, achievementRepo, eventRepo, currencyRepo, logger)

	// Crear el nuevo handler de mecánicas avanzadas del juego
	gameMechanicsHandler := handlers.NewGameMechanicsHandler(gameMechanicsService, logger)

	// Inicializar middleware
	authMiddleware := middleware.NewAuthMiddleware(jwtManager, playerRepo, logger)

	// Configurar router
	r := chi.NewRouter()

	// Middleware global
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Timeout(60 * time.Second))

	// CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Rutas públicas
	r.Group(func(r chi.Router) {
		r.Post("/auth/register", authHandler.Register)
		r.Post("/auth/login", authHandler.Login)
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})
		r.Get("/api/health", func(w http.ResponseWriter, r *http.Request) {
			// Verificar conexión a Redis
			redisStatus := "disconnected"
			if redisService != nil {
				if err := redisService.Ping(); err == nil {
					redisStatus = "connected"
				}
			}

			response := map[string]interface{}{
				"status":    "healthy",
				"timestamp": time.Now().UTC(),
				"version":   "1.0.0",
				"database":  "connected",
				"redis":     redisStatus,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		})
		r.Get("/api/server-time", handlers.ServerTimeHandler)
	})

	// Rutas protegidas
	r.Group(func(r chi.Router) {
		r.Use(authMiddleware.RequireAuth)

		// Rutas de autenticación
		r.Get("/auth/profile", authHandler.GetProfile)
		r.Put("/auth/profile", authHandler.UpdateProfile)

		// Rutas de mundos para el cliente
		r.Route("/client/worlds", func(r chi.Router) {
			// Rutas públicas
			r.Get("/", worldClientHandler.GetWorlds)
			r.Get("/{id}", worldClientHandler.GetWorld)
			r.Get("/{id}/stats", worldClientHandler.GetWorldStats)
			r.Get("/{id}/players", worldClientHandler.GetWorldPlayers)
			r.Get("/{id}/status", worldClientHandler.GetWorldStatus)

			// Rutas autenticadas
			r.Group(func(r chi.Router) {
				r.Use(authMiddleware.RequireAuth)
				r.Post("/{id}/join", worldClientHandler.JoinWorld)
				r.Delete("/{id}/leave", worldClientHandler.LeaveWorld)
				r.Get("/current", worldClientHandler.GetCurrentWorld)
			})
		})

		// Rutas de mundos
		r.Route("/worlds", func(r chi.Router) {
			r.Get("/", worldHandler.GetWorlds)
			r.Get("/{worldID}", worldHandler.GetWorld)
			r.Post("/assign", worldHandler.AssignToWorld)
			r.Post("/{worldID}/join", worldHandler.JoinWorld)
			r.Get("/player/villages", worldHandler.GetPlayerVillages)
		})

		// Rutas de aldeas
		r.Route("/api/villages", func(r chi.Router) {
			r.Use(authMiddleware.RequireAuth)
			r.Get("/", villageHandler.GetPlayerVillages)
			r.Get("/{villageID}", villageHandler.GetVillage)

			// Edificios
			r.Get("/{villageID}/buildings/{buildingType}/upgrade-info", villageHandler.GetBuildingUpgradeInfo)
			r.Post("/{villageID}/buildings/{buildingType}/upgrade", villageHandler.UpgradeBuilding)
			r.Post("/{villageID}/buildings/{buildingType}/complete", villageHandler.CompleteBuildingUpgrade)

			// Nuevas funciones avanzadas de construcción
			r.Get("/{villageID}/buildings/{buildingType}/requirements", villageHandler.CheckBuildingRequirements)
			r.Post("/{villageID}/construction-queue/process", villageHandler.ProcessConstructionQueue)
			r.Get("/{villageID}/construction-queue", villageHandler.GetConstructionQueue)
			r.Delete("/{villageID}/buildings/{buildingType}/upgrade", villageHandler.CancelUpgrade)
		})

		// Rutas de unidades
		r.Route("/api/units", func(r chi.Router) {
			r.Use(authMiddleware.RequireAuth)
			r.Get("/", unitHandler.GetUnits)
			r.Post("/train", unitHandler.TrainUnits)
		})

		// Rutas de chat con Redis Pub/Sub
		r.Route("/chat", func(r chi.Router) {
			r.Use(authMiddleware.RequireAuth)

			// Canales
			r.Get("/channels", chatHandler.GetChannels)
			r.Post("/channels/alliance", chatHandler.CreateAllianceChannel)
			r.Get("/channels/info", chatHandler.GetChannelInfo)

			// Mensajes
			r.Post("/messages", chatHandler.SendMessage)
			r.Get("/messages", chatHandler.GetRecentMessages)

			// Usuarios
			r.Post("/join", chatHandler.JoinChannel)
			r.Post("/leave", chatHandler.LeaveChannel)
			r.Get("/users", chatHandler.GetOnlineUsers)

			// WebSocket para tiempo real
			r.Get("/ws", chatHandler.WebSocketChat)

			// Moderación
			r.Post("/ban", chatHandler.BanUser)
			r.Post("/system", chatHandler.SendSystemMessage)

			// Estadísticas
			r.Get("/stats", chatHandler.GetChatStats)
		})

		// Rutas de alianzas
		r.Route("/api/alliances", func(r chi.Router) {
			r.Use(authMiddleware.RequireAuth)
			r.Get("/", allianceHandler.GetAlliances)
			r.Post("/", allianceHandler.CreateAlliance)
			r.Get("/{id}", allianceHandler.GetAlliance)
			r.Post("/{id}/join", allianceHandler.JoinAlliance)
			r.Post("/{id}/leave", allianceHandler.LeaveAlliance)
		})

		// Rutas de recursos
		r.Route("/resources", func(r chi.Router) {
			r.Get("/village/{villageID}/production", func(w http.ResponseWriter, r *http.Request) {
				villageIDStr := chi.URLParam(r, "villageID")
				villageID, err := uuid.Parse(villageIDStr)
				if err != nil {
					http.Error(w, "ID de aldea inválido", http.StatusBadRequest)
					return
				}
				production, err := resourceService.GetVillageProduction(villageID)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(production)
			})
			r.Get("/village/{villageID}/storage", func(w http.ResponseWriter, r *http.Request) {
				villageIDStr := chi.URLParam(r, "villageID")
				villageID, err := uuid.Parse(villageIDStr)
				if err != nil {
					http.Error(w, "ID de aldea inválido", http.StatusBadRequest)
					return
				}
				storage, err := resourceService.GetVillageStorage(villageID)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(storage)
			})
		})

		// Rutas de jugadores
		r.Route("/players", func(r chi.Router) {
			r.Get("/", func(w http.ResponseWriter, r *http.Request) {
				players, err := playerRepo.GetAllPlayers()
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(players)
			})
			r.Get("/{playerID}", func(w http.ResponseWriter, r *http.Request) {
				playerIDStr := chi.URLParam(r, "playerID")
				playerID, err := uuid.Parse(playerIDStr)
				if err != nil {
					http.Error(w, "ID de jugador inválido", http.StatusBadRequest)
					return
				}
				player, err := playerRepo.GetPlayerByID(playerID)
				if err != nil {
					http.Error(w, err.Error(), http.StatusNotFound)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(player)
			})
			r.Get("/{playerID}/villages", func(w http.ResponseWriter, r *http.Request) {
				playerIDStr := chi.URLParam(r, "playerID")
				playerID, err := uuid.Parse(playerIDStr)
				if err != nil {
					http.Error(w, "ID de jugador inválido", http.StatusBadRequest)
					return
				}
				villages, err := villageRepo.GetVillagesByPlayerID(playerID)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(villages)
			})
		})

		// Rutas de edificios
		r.Route("/buildings", func(r chi.Router) {
			r.Get("/", func(w http.ResponseWriter, r *http.Request) {
				buildings, err := villageRepo.GetBuildingTypes()
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(buildings)
			})
			r.Get("/{buildingID}", func(w http.ResponseWriter, r *http.Request) {
				buildingID := chi.URLParam(r, "buildingID")
				building, err := villageRepo.GetBuildingType(buildingID)
				if err != nil {
					http.Error(w, err.Error(), http.StatusNotFound)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(building)
			})
		})

		// Rutas de logros
		r.Route("/api/achievements", func(r chi.Router) {
			// Dashboard principal
			r.Get("/dashboard", achievementHandler.GetAchievementDashboard)

			// Categorías de logros
			r.Get("/categories", achievementHandler.GetAchievementCategories)
			r.Get("/categories/{id}", achievementHandler.GetAchievementCategory)
			r.Post("/categories", achievementHandler.CreateAchievementCategory)

			// Logros
			r.Get("/", achievementHandler.GetAchievements)
			r.Get("/{id}", achievementHandler.GetAchievement)
			r.Get("/{id}/details", achievementHandler.GetAchievementWithDetails)

			// Logros del jugador
			r.Get("/player/achievements", achievementHandler.GetPlayerAchievements)
			r.Get("/player/achievements/{id}", achievementHandler.GetPlayerAchievement)
			r.Put("/player/achievements/{id}/progress", achievementHandler.UpdateAchievementProgress)
			r.Post("/player/achievements/{id}/claim", achievementHandler.ClaimAchievementReward)

			// Estadísticas y rankings
			r.Get("/player/statistics", achievementHandler.GetAchievementStatistics)
			r.Get("/leaderboard", achievementHandler.GetAchievementLeaderboard)

			// Notificaciones
			r.Get("/notifications", achievementHandler.GetAchievementNotifications)
			r.Put("/notifications/{id}/read", achievementHandler.MarkNotificationAsRead)

			// Utilidades
			r.Get("/{id}/calculate-progress", achievementHandler.CalculateAchievementProgress)
		})

		// Rutas de eventos
		r.Route("/events", func(r chi.Router) {
			r.Get("/", func(w http.ResponseWriter, r *http.Request) {
				events, err := eventService.GetActiveEvents(r.Context())
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(events)
			})
			r.Get("/{eventID}", func(w http.ResponseWriter, r *http.Request) {
				event, err := eventService.GetEvent(r.Context(), chi.URLParam(r, "eventID"))
				if err != nil {
					http.Error(w, err.Error(), http.StatusNotFound)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(event)
			})
			r.Post("/{eventID}/participate", func(w http.ResponseWriter, r *http.Request) {
				eventID := chi.URLParam(r, "eventID")
				playerID := r.Context().Value("player_id").(string)
				err := eventService.RegisterPlayerForEvent(playerID, eventID)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				w.WriteHeader(http.StatusOK)
			})
		})

		// Rutas del sistema de rankings
		r.Route("/api/rankings", func(r chi.Router) {
			// Dashboard principal
			r.Get("/dashboard", rankingHandler.GetRankingsDashboard)
			r.Get("/summary", rankingHandler.GetStatisticsSummary)

			// Categorías de ranking
			r.Get("/categories", rankingHandler.GetRankingCategories)
			r.Get("/categories/{categoryID}", rankingHandler.GetRankingCategory)
			r.Post("/categories", rankingHandler.CreateRankingCategory)
			r.Put("/categories/{categoryID}", rankingHandler.UpdateRankingCategory)

			// Temporadas de ranking
			r.Get("/seasons", rankingHandler.GetRankingSeasons)

			// Entradas de ranking
			r.Get("/categories/{categoryID}/entries", rankingHandler.GetRankingEntries)

			// Rankings específicos
			r.Get("/top/players", rankingHandler.GetTopPlayers)
			r.Get("/top/alliances", rankingHandler.GetTopAlliances)
			r.Get("/comparison", rankingHandler.GetRankingComparison)

			// Historial de rankings
			r.Get("/categories/{categoryID}/history", rankingHandler.GetRankingHistory)
		})

		// Rutas de estadísticas
		r.Route("/api/statistics", func(r chi.Router) {
			r.Get("/players/{playerID}", rankingHandler.GetPlayerStatistics)
			r.Get("/alliances/{allianceID}", rankingHandler.GetAllianceStatistics)
			r.Get("/villages/{villageID}", rankingHandler.GetVillageStatistics)
			r.Get("/worlds/{worldID}", rankingHandler.GetWorldStatistics)
			r.Get("/players/{playerID}/rankings", rankingHandler.GetPlayerRankings)
		})

		// Rutas de notificaciones
		r.Route("/notifications", func(r chi.Router) {
			r.Get("/", func(w http.ResponseWriter, r *http.Request) {
				playerID := r.Context().Value("player_id").(string)
				limit := 50 // Por defecto
				notifications, err := achievementService.GetAchievementNotifications(playerID, limit)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(notifications)
			})
			r.Put("/{notificationID}/read", func(w http.ResponseWriter, r *http.Request) {
				playerID := r.Context().Value("player_id").(string)
				notificationID := chi.URLParam(r, "notificationID")
				err := achievementService.MarkNotificationAsRead(playerID, notificationID)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				w.WriteHeader(http.StatusOK)
			})
			r.Delete("/{notificationID}", func(w http.ResponseWriter, r *http.Request) {
				playerID := r.Context().Value("player_id").(string)
				notificationID := chi.URLParam(r, "notificationID")
				err := achievementService.MarkNotificationAsRead(playerID, notificationID)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				w.WriteHeader(http.StatusOK)
			})
		})

		// Rutas de reportes
		r.Route("/reports", func(r chi.Router) {
			r.Get("/", func(w http.ResponseWriter, r *http.Request) {
				reports := []interface{}{} // Stub temporal
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(reports)
			})
			r.Get("/{reportID}", func(w http.ResponseWriter, r *http.Request) {
				reportID := chi.URLParam(r, "reportID")
				report := map[string]interface{}{"id": reportID, "status": "pending"} // Stub temporal
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(report)
			})
			r.Delete("/{reportID}", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})
		})

		// Rutas de investigación
		r.Route("/api/research", func(r chi.Router) {
			r.Use(authMiddleware.RequireAuth)

			// Obtener tecnologías disponibles
			r.Get("/technologies", researchHandler.GetTechnologies)
			r.Get("/technologies/{id}", researchHandler.GetTechnology)
			r.Get("/technologies/{id}/details", researchHandler.GetTechnologyDetails)

			// Tecnologías del jugador
			r.Get("/player/technologies", researchHandler.GetPlayerTechnologies)
			r.Get("/player/queue", researchHandler.GetResearchQueue)
			r.Get("/player/statistics", researchHandler.GetResearchStatistics)
			r.Get("/player/recommendations", researchHandler.GetResearchRecommendations)
			r.Get("/player/history", researchHandler.GetResearchHistory)
			r.Get("/player/bonuses", researchHandler.GetResearchBonuses)

			// Acciones de investigación
			r.Post("/start", researchHandler.StartResearch)
			r.Post("/complete", researchHandler.CompleteResearch)
			r.Post("/cancel", researchHandler.CancelResearch)

			// Información general
			r.Get("/tree", researchHandler.GetTechnologyTree)
			r.Get("/rankings", researchHandler.GetTechnologyRankings)
		})

		// Rutas de héroes
		r.Route("/api/heroes", func(r chi.Router) {
			r.Get("/", heroHandler.GetHeroes)
			r.Get("/rankings", heroHandler.GetHeroRankings)
			r.Get("/config", heroHandler.GetHeroSystemConfig)
			r.Put("/config", heroHandler.UpdateHeroSystemConfig)
			r.Get("/{id}", heroHandler.GetHero)
			r.Post("/{id}/recruit", heroHandler.RecruitHero)
			r.Post("/{id}/upgrade", heroHandler.UpgradeHero)
			r.Post("/{id}/activate", heroHandler.ActivateHero)
			r.Post("/{id}/deactivate", heroHandler.DeactivateHero)
			r.Get("/{id}/progress", heroHandler.GetHeroProgress)
		})

		// Rutas de héroes de jugador
		r.Route("/api/player/heroes", func(r chi.Router) {
			r.Get("/", heroHandler.GetPlayerHeroes)
			r.Get("/active", heroHandler.GetActiveHeroes)
			r.Get("/{id}", heroHandler.GetPlayerHero)
		})

		// Rutas de batallas
		r.Route("/api/battles", func(r chi.Router) {
			r.Get("/", battleHandler.GetPlayerBattles)
			r.Get("/active", battleHandler.GetActiveBattles)
			r.Get("/incoming", battleHandler.GetIncomingAttacks)
			r.Get("/outgoing", battleHandler.GetOutgoingAttacks)
			r.Get("/rankings", battleHandler.GetBattleRankings)
			r.Get("/statistics", battleHandler.GetBattleStatistics)
			r.Post("/attack", battleHandler.AttackVillage)
			r.Get("/{battleID}", battleHandler.GetBattle)
			r.Get("/{battleID}/report", battleHandler.GetBattleReport)
			r.Get("/{battleID}/log", battleHandler.GetBattleLog)
			r.Delete("/{battleID}", battleHandler.CancelBattle)
		})

		// Rutas de defensas de aldeas
		r.Route("/api/villages/{villageID}", func(r chi.Router) {
			r.Get("/defenses", battleHandler.GetVillageDefenses)
		})

		// Rutas del sistema de economía
		r.Route("/api/economy", func(r chi.Router) {
			// Dashboard y configuración
			r.Get("/dashboard", economyHandler.GetEconomyDashboard)
			r.Get("/config", economyHandler.GetEconomyConfig)
			r.Put("/config", economyHandler.UpdateEconomyConfig)

			// Items del mercado
			r.Get("/market/items", economyHandler.GetMarketItems)
			r.Get("/market/items/{itemID}", economyHandler.GetMarketItemDetails)
			r.Get("/market/listings", economyHandler.GetMarketListings)
			r.Post("/market/listings", economyHandler.CreateMarketListing)

			// Transacciones y estadísticas
			r.Get("/market/statistics", economyHandler.GetMarketStatistics)
			r.Get("/market/trends", economyHandler.GetMarketTrends)

			// Intercambio de monedas
			r.Post("/exchange", economyHandler.ExchangeCurrency)

			// Economía de jugadores
			r.Get("/player/{playerID}", economyHandler.GetPlayerEconomy)
			r.Get("/player/{playerID}/activity", economyHandler.GetPlayerMarketActivity)

			// Recursos de jugadores
			r.Get("/player/{playerID}/resources", economyHandler.GetPlayerResources)
			r.Post("/player/{playerID}/resources/add", economyHandler.AddResources)
			r.Post("/player/{playerID}/resources/remove", economyHandler.RemoveResources)

			// Historial de transacciones
			r.Get("/player/{playerID}/transactions", economyHandler.GetTransactionHistory)
		})

		// Rutas del sistema de misiones
		r.Route("/api/quests", func(r chi.Router) {
			// Dashboard principal
			r.Get("/dashboard", questHandler.GetQuestDashboard)

			// Categorías de misiones
			r.Get("/categories", questHandler.GetAllQuestCategories)
			r.Get("/categories/{id}", questHandler.GetQuestCategory)
			r.Post("/categories", questHandler.CreateQuestCategory)

			// Misiones
			r.Get("/available", questHandler.GetAvailableQuests)
			r.Get("/{id}", questHandler.GetQuest)
			r.Get("/{id}/details", questHandler.GetQuestWithDetails)

			// Misiones del jugador
			r.Get("/player/active", questHandler.GetPlayerActiveQuests)
			r.Post("/player/progress", questHandler.UpdateQuestProgress)
			r.Post("/player/claim", questHandler.ClaimQuestRewards)

			// Procesamiento de eventos del juego
			r.Post("/process-event", questHandler.ProcessGameEvent)
		})

		// Rutas del sistema de eventos
		r.Route("/api/events", func(r chi.Router) {
			// Dashboard principal
			r.Get("/dashboard", eventHandler.GetEventDashboard)

			// Categorías de eventos
			r.Get("/categories", eventHandler.GetAllEventCategories)
			r.Get("/categories/{id}", eventHandler.GetEventCategory)
			r.Post("/categories", eventHandler.CreateEventCategory)

			// Eventos
			r.Get("/by-category", eventHandler.GetEventsByCategory)
			r.Get("/active", eventHandler.GetActiveEvents)
			r.Get("/upcoming", eventHandler.GetUpcomingEvents)
			r.Get("/{id}", eventHandler.GetEvent)
			r.Post("/", eventHandler.CreateEvent)

			// Participantes
			r.Get("/{eventID}/participants", eventHandler.GetEventParticipants)
			r.Post("/register", eventHandler.RegisterPlayerForEvent)
			r.Post("/progress", eventHandler.UpdateEventProgress)

			// Partidas
			r.Get("/{eventID}/matches", eventHandler.GetEventMatches)
			r.Post("/matches", eventHandler.CreateEventMatch)
		})

		// Rutas del sistema de títulos y prestigio
		r.Route("/api/titles", func(r chi.Router) {
			// Dashboard de títulos
			r.Get("/dashboard", titleHandler.GetTitleDashboard)

			// Categorías de títulos
			r.Get("/categories", titleHandler.GetTitleCategories)
			r.Get("/categories/{categoryID}", titleHandler.GetTitleCategory)
			r.Post("/categories", titleHandler.CreateTitleCategory)

			// Títulos
			r.Get("/", titleHandler.GetTitles)
			r.Get("/{titleID}", titleHandler.GetTitle)
			r.Post("/", titleHandler.CreateTitle)

			// Títulos del jugador
			r.Get("/player/titles", titleHandler.GetPlayerTitles)
			r.Post("/grant", titleHandler.GrantTitle)
			r.Post("/equip", titleHandler.EquipTitle)
			r.Post("/unequip", titleHandler.UnequipTitle)

			// Leaderboard y estadísticas
			r.Get("/leaderboard", titleHandler.GetTitleLeaderboard)
			r.Get("/statistics", titleHandler.GetTitleStatistics)
		})

		// Rutas de comercio
		r.Route("/api/trade", func(r chi.Router) {
			r.Get("/offers", tradeHandler.GetTradeOffers)
			r.Get("/offers/{offerID}", tradeHandler.GetTradeOffer)
			r.Post("/offers", tradeHandler.CreateTradeOffer)
			r.Post("/offers/{offerID}/buy", tradeHandler.BuyTradeOffer)
			r.Delete("/offers/{offerID}", tradeHandler.CancelTradeOffer)
			r.Get("/player/offers", tradeHandler.GetPlayerTradeOffers)
			r.Get("/player/history", tradeHandler.GetTradeHistory)
			r.Get("/stats", tradeHandler.GetMarketStats)
			r.Get("/prices", tradeHandler.GetResourcePrices)
			r.Post("/direct", tradeHandler.CreateDirectTrade)
			r.Post("/direct/{tradeID}/accept", tradeHandler.AcceptDirectTrade)
			r.Post("/direct/{tradeID}/decline", tradeHandler.DeclineDirectTrade)
			r.Get("/direct", tradeHandler.GetDirectTrades)
		})

		// Rutas de mecánicas avanzadas del juego
		r.Route("/api/game-mechanics", func(r chi.Router) {
			r.Use(authMiddleware.RequireAuth)
			r.Get("/village/{villageID}/resources/production", gameMechanicsHandler.CalculateResourceProduction)
			r.Post("/village/{villageID}/resources/update", gameMechanicsHandler.UpdateVillageResources)
			r.Post("/battle/calculate", gameMechanicsHandler.CalculateBattleOutcome)
			r.Get("/trade/rates", gameMechanicsHandler.CalculateTradeRates)
			r.Get("/alliance/{allianceID}/benefits", gameMechanicsHandler.ProcessAllianceBenefits)
			r.Get("/player/{playerID}/score", gameMechanicsHandler.CalculatePlayerScore)
			r.Post("/player/daily-rewards", gameMechanicsHandler.GenerateDailyRewards)
			r.Post("/admin/cleanup", gameMechanicsHandler.CleanupInactiveData)
		})

		// Rutas de Redis (solo para administradores)
		r.Route("/api/redis", func(r chi.Router) {
			r.Use(authMiddleware.RequireAuth)

			// Verificación de conexión
			r.Get("/ping", redisHandler.PingRedis)

			// Estadísticas
			r.Get("/stats", redisHandler.GetRedisStats)

			// Usuarios online
			r.Get("/users/online", redisHandler.GetOnlineUsers)

			// Sesiones de usuario
			r.Get("/sessions/{userID}", redisHandler.GetUserSession)
			r.Delete("/sessions/{userID}", redisHandler.DeleteUserSession)

			// Recursos de jugadores
			r.Get("/resources/{playerID}", redisHandler.GetPlayerResources)

			// Progreso de investigación
			r.Get("/research/{playerID}", redisHandler.GetResearchProgress)

			// Notificaciones
			r.Get("/notifications/{userID}", redisHandler.GetNotifications)
			r.Post("/notifications/{userID}/read/{notificationID}", redisHandler.MarkNotificationAsRead)
			r.Post("/notifications", redisHandler.AddNotification)

			// Cache general
			r.Get("/cache/{key}", redisHandler.GetCache)
			r.Post("/cache/{key}", redisHandler.SetCache)
			r.Delete("/cache/{key}", redisHandler.DeleteCache)

			// Contadores
			r.Get("/counters/{key}", redisHandler.GetCounter)
			r.Post("/counters/{key}/increment", redisHandler.IncrementCounter)

			// Limpieza (solo desarrollo)
			r.Delete("/flush", redisHandler.FlushRedis)
		})

		// Rutas administrativas
		r.Route("/admin", func(r chi.Router) {
			r.Get("/stats", adminHandler.GetServerStats)
			r.Get("/players", adminHandler.GetPlayers)
			r.Get("/villages", adminHandler.GetVillages)
			r.Get("/battles", adminHandler.GetBattles)
			r.Get("/alliances", adminHandler.GetAlliances)
			r.Get("/events", adminHandler.GetEvents)

			// Rutas de gestión de mundos
			r.Route("/worlds", func(r chi.Router) {
				r.Get("/", adminHandler.GetWorlds)             // GET /admin/worlds
				r.Post("/", adminHandler.CreateWorld)          // POST /admin/worlds
				r.Put("/{id}", adminHandler.UpdateWorld)       // PUT /admin/worlds/{id}
				r.Delete("/{id}", adminHandler.DeleteWorld)    // DELETE /admin/worlds/{id}
				r.Post("/{id}/start", adminHandler.StartWorld) // POST /admin/worlds/{id}/start
				r.Post("/{id}/stop", adminHandler.StopWorld)   // POST /admin/worlds/{id}/stop

				// Rutas de gestión de jugadores en mundos
				r.Get("/{id}/status", adminHandler.GetWorldStatus)                       // GET /admin/worlds/{id}/status
				r.Get("/{id}/players", adminHandler.GetWorldPlayers)                     // GET /admin/worlds/{id}/players
				r.Post("/{id}/players", adminHandler.AddPlayerToWorld)                   // POST /admin/worlds/{id}/players
				r.Delete("/{id}/players/{playerId}", adminHandler.RemovePlayerFromWorld) // DELETE /admin/worlds/{id}/players/{playerId}
			})

			// Rutas de gestión de monedas
			r.Route("/currency", func(r chi.Router) {
				r.Get("/config", adminHandler.GetCurrencyConfig)                                // GET /admin/currency/config
				r.Put("/config", adminHandler.UpdateCurrencyConfig)                             // PUT /admin/currency/config
				r.Get("/stats", adminHandler.GetCurrencyStats)                                  // GET /admin/currency/stats
				r.Post("/add", adminHandler.AddCurrency)                                        // POST /admin/currency/add
				r.Post("/transfer", adminHandler.TransferCurrency)                              // POST /admin/currency/transfer
				r.Get("/players/{playerId}/balance", adminHandler.GetPlayerCurrencyBalance)     // GET /admin/currency/players/{playerId}/balance
				r.Get("/players/{playerId}/transactions", adminHandler.GetCurrencyTransactions) // GET /admin/currency/players/{playerId}/transactions
			})
		})
	})

	// Rutas de aldeas para el cliente
	r.Route("/client/villages", func(r chi.Router) {
		r.Use(authMiddleware.RequireAuth)
		r.Get("/{id}", villageHandler.GetVillage)
		r.Get("/{id}/production", func(w http.ResponseWriter, r *http.Request) {
			villageIDStr := chi.URLParam(r, "id")
			villageID, err := uuid.Parse(villageIDStr)
			if err != nil {
				http.Error(w, "ID de aldea inválido", http.StatusBadRequest)
				return
			}
			production, err := resourceService.GetVillageProduction(villageID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(production)
		})
		r.Get("/{id}/storage", func(w http.ResponseWriter, r *http.Request) {
			villageIDStr := chi.URLParam(r, "id")
			villageID, err := uuid.Parse(villageIDStr)
			if err != nil {
				http.Error(w, "ID de aldea inválido", http.StatusBadRequest)
				return
			}
			storage, err := resourceService.GetVillageStorage(villageID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(storage)
		})
	})

	// WebSocket endpoint
	r.HandleFunc("/ws", wsManager.HandleWebSocket)

	// Inicializar canales por defecto
	if err := chatRepo.InitializeDefaultChannels(); err != nil {
		logger.Error("Error inicializando canales por defecto", zap.Error(err))
	}

	// Iniciar el gestor de sincronización robusto
	if redisService != nil {
		go func() {
			// Esperar un poco para que Redis esté completamente listo
			time.Sleep(2 * time.Second)

			// Crear contexto con timeout para la inicialización
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			if err := chatService.StartSyncManager(ctx); err != nil {
				logger.Error("Error iniciando gestor de sincronización", zap.Error(err))
			} else {
				logger.Info("Gestor de sincronización iniciado exitosamente")
			}
		}()
	}

	// Iniciar servicio de generación de recursos en background
	go resourceService.StartResourceGeneration()

	// Iniciar WebSocket manager en background
	go wsManager.Start()

	// Tarea en segundo plano: procesar cola de construcción
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// Procesar todas las aldeas
				villages, err := villageRepo.GetAllVillages()
				if err != nil {
					logger.Error("Error obteniendo aldeas para procesar cola de construcción", zap.Error(err))
					continue
				}

				for _, village := range villages {
					// Procesar cola de construcción para cada aldea
					_, err := constructionService.ProcessConstructionQueue(village.Village.ID)
					if err != nil {
						logger.Error("Error procesando cola de construcción",
							zap.String("village_id", village.Village.ID.String()),
							zap.Error(err))
					}
				}
			}
		}
	}()

	// Configurar servidor HTTP
	srv := &http.Server{
		Addr:         fmt.Sprintf("0.0.0.0:%d", cfg.Server.Port), // ← CAMBIADO: De ":%d" a "0.0.0.0:%d"
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

	// Detener el gestor de sincronización
	if redisService != nil {
		if err := chatService.StopSyncManager(); err != nil {
			logger.Error("Error deteniendo gestor de sincronización", zap.Error(err))
		} else {
			logger.Info("Gestor de sincronización detenido exitosamente")
		}
	}

	// Dar tiempo para que las conexiones se cierren
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Error cerrando servidor", zap.Error(err))
	}

	logger.Info("Servidor cerrado exitosamente")
}
