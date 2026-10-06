package server

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/nxwex/meecore/internal/docker"
	"github.com/nxwex/meecore/internal/node"
	"github.com/nxwex/meecore/internal/service"
)

type Server struct {
	httpServer *http.Server
	nodes      *node.Service
	docker     *docker.Client
	templates  map[string]service.Template
	instances  InstanceStorage
}

func New(addr string, nodeService *node.Service, dockerClient *docker.Client, templates map[string]service.Template, instanceStorage InstanceStorage) *Server {
	s := &Server{
		nodes:     nodeService,
		docker:    dockerClient,
		templates: templates,
		instances: instanceStorage,
	}

	s.httpServer = &http.Server{
		Addr:    addr,
		Handler: s.setupRouter(),
	}

	return s
}

func (s *Server) setupRouter() http.Handler {
	router := chi.NewRouter()

	router.Use(
		middleware.RequestID,
		middleware.RealIP,
		logger,
		middleware.Recoverer,
	)

	router.Get("/health", s.health)
	router.Get("/api/nodes/{id}", s.getNode)
	router.Get("/api/nodes", s.getNodes)

	router.Get("/api/templates", s.getTemplates)

	router.Get("/api/instances", s.getInstances)
	router.Get("/api/instances/{id}", s.getInstance)

	// консоль
	router.Get("/api/instances/{id}/console", s.instanceConsole)
	router.Get("/console/{id}", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/static/console.html")
	})

	router.Post("/api/instances", s.createInstance)

	router.Delete("/api/instances/{id}", s.deleteInstance)

	router.Get("/api/containers", s.getContainers)
	router.Get("/api/containers/{id≠}", s.getContainer)

	router.Post("/api/containers", s.createContainer)
	router.Post("/api/containers/{id}/start", s.startContainer)
	router.Post("/api/containers/{id}/stop", s.stopContainer)
	router.Post("/api/containers/{id}/restart", s.restartContainer)

	router.Delete("/api/containers/{id}", s.deleteContainer)

	router.Handle("/*", http.FileServer(http.Dir("web/static")))

	return router
}

func (s *Server) Start() error {
	err := s.httpServer.ListenAndServe()

	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	if err := writeJSON(w, http.StatusOK, "ok"); err != nil {
		log.Printf("write healthcheck error: %v", err)
	}
}
