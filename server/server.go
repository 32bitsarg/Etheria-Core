package server

import (
	"fmt"
	"net/http"
	"time"

	"server-backend/config"
	"server-backend/handlers"
	"server-backend/middleware"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type Server struct {
	router         *chi.Mux
	logger         *zap.Logger
	upgrader       websocket.Upgrader
	clients        map[*websocket.Conn]bool
	broadcast      chan []byte
	register       chan *websocket.Conn
	unregister     chan *websocket.Conn
	authHandler    *handlers.AuthHandler
	worldHandler   *handlers.WorldHandler
	villageHandler *handlers.VillageHandler
	authMiddleware *middleware.AuthMiddleware
}

func NewServer(
	logger *zap.Logger,
	authHandler *handlers.AuthHandler,
	worldHandler *handlers.WorldHandler,
	villageHandler *handlers.VillageHandler,
	authMiddleware *middleware.AuthMiddleware,
) *Server {
	s := &Server{
		router:         chi.NewRouter(),
		logger:         logger,
		authHandler:    authHandler,
		worldHandler:   worldHandler,
		villageHandler: villageHandler,
		authMiddleware: authMiddleware,
		upgrader:       config.GetWebSocketUpgrader(),
		clients:        make(map[*websocket.Conn]bool),
		broadcast:      make(chan []byte),
		register:       make(chan *websocket.Conn),
		unregister:     make(chan *websocket.Conn),
	}

	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	// Middleware
	s.router.Use(chimiddleware.Logger)
	s.router.Use(chimiddleware.Recoverer)
	s.router.Use(chimiddleware.Timeout(60 * time.Second))

	// CORS
	s.router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Rutas
	s.router.Get("/", s.handleHome)
	s.router.Get("/ws", s.handleWebSocket)
	s.router.Route("/api", func(r chi.Router) {
		r.Route("/v1", func(r chi.Router) {
			// Rutas de autenticación
			r.Post("/auth/register", s.handleRegister)
			r.Post("/auth/login", s.handleLogin)
			r.Post("/auth/logout", s.handleLogout)

			// Rutas de mundos (públicas)
			r.Get("/worlds", s.worldHandler.GetWorlds)

			// Rutas protegidas
			r.Group(func(r chi.Router) {
				r.Use(s.authMiddleware.RequireAuth)

				// Rutas de mundos
				r.Post("/worlds/{worldID}/join", s.worldHandler.JoinWorld)

				// Rutas de jugador
				r.Route("/player", func(r chi.Router) {
					r.Get("/", s.handleGetPlayer)
					r.Put("/", s.handleUpdatePlayer)
				})

				// Rutas de aldeas/ciudades
				r.Route("/cities", func(r chi.Router) {
					r.Get("/", s.handleGetCities)
					r.Post("/", s.handleCreateCity)
					r.Get("/{id}", s.handleGetCity)
					r.Put("/{id}", s.handleUpdateCity)
				})

				// Rutas de aldeas (alias para compatibilidad)
				r.Route("/village", func(r chi.Router) {
					r.Get("/{id}", s.villageHandler.GetVillage)
					r.Get("/", s.villageHandler.GetPlayerVillages)
				})

				// Rutas de edificios
				r.Route("/buildings", func(r chi.Router) {
					r.Get("/", s.handleGetBuildings)
					r.Post("/", s.handleCreateBuilding)
					r.Put("/{id}", s.handleUpdateBuilding)
				})

				// Rutas de unidades
				r.Route("/units", func(r chi.Router) {
					r.Get("/", s.handleGetUnits)
					r.Post("/train", s.handleTrainUnits)
				})
			})
		})
	})
}

func (s *Server) Start(port int) error {
	// Iniciar el manejador de WebSocket
	go s.handleWebSocketMessages()

	// Iniciar el servidor HTTP
	addr := fmt.Sprintf(":%d", port)
	s.logger.Info("Iniciando servidor", zap.String("addr", addr))
	return http.ListenAndServe(addr, s.router)
}

func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte("Servidor MMO funcionando correctamente"))
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logger.Error("Error al actualizar conexión a WebSocket", zap.Error(err))
		return
	}

	s.register <- conn

	// Configurar el manejador de mensajes para esta conexión
	go s.handleClientMessages(conn)
}

func (s *Server) handleWebSocketMessages() {
	for {
		select {
		case client := <-s.register:
			s.clients[client] = true
			s.logger.Info("Nuevo cliente conectado")

		case client := <-s.unregister:
			if _, ok := s.clients[client]; ok {
				delete(s.clients, client)
				client.Close()
				s.logger.Info("Cliente desconectado")
			}

		case message := <-s.broadcast:
			for client := range s.clients {
				err := client.WriteMessage(websocket.TextMessage, message)
				if err != nil {
					s.logger.Error("Error al enviar mensaje", zap.Error(err))
					client.Close()
					delete(s.clients, client)
				}
			}
		}
	}
}

func (s *Server) handleClientMessages(conn *websocket.Conn) {
	defer func() {
		s.unregister <- conn
	}()

	// Crear validador de mensajes
	validator := middleware.NewWebSocketValidator(s.logger)
	clientInfo := conn.RemoteAddr().String()

	// Configurar timeouts
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				s.logger.Error("Error al leer mensaje", zap.Error(err))
			}
			break
		}

		// Log del mensaje recibido
		validator.LogMessageReceived(message, clientInfo)

		// Validar tipo de mensaje
		if !validator.IsValidMessageType(messageType) {
			validator.SendError(conn, "Solo se permiten mensajes de texto")
			continue
		}

		// Validar mensaje
		_, err = validator.ValidateMessage(message)
		if err != nil {
			s.logger.Warn("Mensaje inválido recibido",
				zap.Error(err),
				zap.String("message", string(message)),
				zap.String("client", clientInfo))
			validator.SendError(conn, fmt.Sprintf("Mensaje inválido: %s", err.Error()))
			continue
		}

		// Procesar el mensaje recibido
		s.broadcast <- message

		// Actualizar deadline de lectura
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	}
}

// Manejadores de rutas HTTP implementados

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	s.authHandler.Register(w, r)
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	s.authHandler.Login(w, r)
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	// Limpiar token del cliente
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Sesión cerrada exitosamente"}`))
}

func (s *Server) handleGetPlayer(w http.ResponseWriter, r *http.Request) {
	s.authHandler.GetProfile(w, r)
}

func (s *Server) handleUpdatePlayer(w http.ResponseWriter, r *http.Request) {
	s.authHandler.UpdateProfile(w, r)
}

func (s *Server) handleGetCities(w http.ResponseWriter, r *http.Request) {
	// Obtener todas las aldeas del jugador
	s.villageHandler.GetPlayerVillages(w, r)
}

func (s *Server) handleCreateCity(w http.ResponseWriter, r *http.Request) {
	// Crear una nueva aldea para el jugador
	w.Header().Set("Content-Type", "application/json")

	// Obtener el ID del jugador del contexto
	playerIDStr := r.Context().Value("player_id").(string)

	// TODO: Implementar lógica para crear aldea
	// Por ahora, devolver respuesta de placeholder
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"message": "Aldea creada exitosamente", "player_id": "` + playerIDStr + `"}`))
}

func (s *Server) handleGetCity(w http.ResponseWriter, r *http.Request) {
	// Obtener información de una aldea específica
	s.villageHandler.GetVillage(w, r)
}

func (s *Server) handleUpdateCity(w http.ResponseWriter, r *http.Request) {
	// Actualizar información de una aldea
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Aldea actualizada exitosamente"}`))
}

func (s *Server) handleGetBuildings(w http.ResponseWriter, r *http.Request) {
	// Obtener edificios de una aldea
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"buildings": []}`))
}

func (s *Server) handleCreateBuilding(w http.ResponseWriter, r *http.Request) {
	// Crear un nuevo edificio en una aldea
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"message": "Edificio creado exitosamente"}`))
}

func (s *Server) handleUpdateBuilding(w http.ResponseWriter, r *http.Request) {
	// Actualizar/mejorar un edificio
	s.villageHandler.UpgradeBuilding(w, r)
}

func (s *Server) handleGetUnits(w http.ResponseWriter, r *http.Request) {
	// Obtener unidades de una aldea
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"units": []}`))
}

func (s *Server) handleTrainUnits(w http.ResponseWriter, r *http.Request) {
	// Entrenar unidades en una aldea
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Unidades en entrenamiento"}`))
}
