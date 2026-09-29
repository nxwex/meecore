package server

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/nxwex/meecore/internal/docker"
	"github.com/nxwex/meecore/internal/node"
)

type Server struct {
	httpServer *http.Server
	nodes      *node.Service
	docker     *docker.Client
}

func New(addr string, nodeService *node.Service, dockerClient *docker.Client) *Server {
	s := &Server{
		nodes:  nodeService,
		docker: dockerClient,
	}

	s.httpServer = &http.Server{
		Addr:    addr,
		Handler: s.setupRouter(),
	}

	return s
}

func (s *Server) setupRouter() http.Handler {
	router := chi.NewRouter()

	router.Get("/health", s.health)
	router.Get("/api/nodes/{id}", s.getNode)

	router.Get("/api/containers/{id}", s.getContainer)
	router.Get("/api/containers", s.getContainers)

	router.Post("/api/containers/{id}/start", s.startContainer)
	router.Post("/api/containers/{id}/stop", s.stopContainer)
	router.Post("/api/containers/{id}/restart", s.restartContainer)

	router.Delete("/api/containers/{id}", s.deleteContainer)

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
