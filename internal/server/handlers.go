package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/nxwex/meecore/internal/node"
)

func (s *Server) getNode(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid node id", http.StatusBadRequest)
		return
	}

	n, err := s.nodes.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, node.ErrNotFound) {
			http.Error(w, "node not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if err := writeJSON(w, http.StatusOK, n); err != nil {
		log.Printf("writeJSON error: %v", err)
	}
}

func (s *Server) getContainer(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	container, err := s.docker.GetContainer(r.Context(), id)
	if err != nil {
		http.Error(w, "container not found", http.StatusNotFound)
		return
	}

	if err := writeJSON(w, http.StatusOK, container); err != nil {
		log.Printf("writeJSON error: %v", err)
	}
}

func (s *Server) getContainers(w http.ResponseWriter, r *http.Request) {
	containers, err := s.docker.GetConatiners(r.Context())
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if err := writeJSON(w, http.StatusOK, containers); err != nil {
		log.Printf("writeJSON error: %v", err)
	}
}

func (s *Server) startContainer(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := s.docker.StartContainer(r.Context(), id); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if err := writeJSON(w, http.StatusAccepted, "accepted"); err != nil {
		log.Printf("writeJSON error: %v", err)
	}
}

func (s *Server) stopContainer(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := s.docker.StopContainer(r.Context(), id); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if err := writeJSON(w, http.StatusAccepted, "accepted"); err != nil {
		log.Printf("writeJSON error: %v", err)
	}
}

func (s *Server) restartContainer(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := s.docker.RestartContainer(r.Context(), id); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if err := writeJSON(w, http.StatusAccepted, "accepted"); err != nil {
		log.Printf("writeJSON error: %v", err)
	}
}

func (s *Server) deleteContainer(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := s.docker.RemoveContainer(r.Context(), id); err != nil {
		log.Printf("remove container error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if err := writeJSON(w, http.StatusOK, "deleted"); err != nil {
		log.Printf("writeJSON error: %v", err)
	}
}

func writeJSON(w http.ResponseWriter, statusCode int, data any) error {
	body, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal response: %w", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if _, err := w.Write(body); err != nil {
		return fmt.Errorf("write response: %w", err)
	}

	return nil
}
