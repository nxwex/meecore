package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/nxwex/meecore/internal/node"
	"github.com/nxwex/meecore/internal/service"
)

type InstanceStorage interface {
	CreateInstance(ctx context.Context, instance *Instance) error
	GetInstance(ctx context.Context, id int64) (*Instance, error)
	ListInstances(ctx context.Context) ([]Instance, error)
	DeleteInstance(ctx context.Context, id int64) error
}

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

func (s *Server) getNodes(w http.ResponseWriter, r *http.Request) {
	nodes, err := s.nodes.GetAll(r.Context())
	if err != nil {
		log.Printf("get all nodes error: %v", err)

		if err := writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to get nodes",
		}); err != nil {
			log.Printf("write nodes error response: %v", err)
		}

		return
	}

	if err := writeJSON(w, http.StatusOK, nodes); err != nil {
		log.Printf("write nodes response error: %v", err)
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

func (s *Server) createContainer(w http.ResponseWriter, r *http.Request) {
	contRequest := CreateContainerRequest{}

	if err := json.NewDecoder(r.Body).Decode(&contRequest); err != nil {
		log.Printf("decode json error: %v", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	template, ok := s.templates[contRequest.Template]
	if !ok {
		http.Error(w, "template not found", http.StatusBadRequest)
		return
	}

	id, err := s.docker.CreateContainer(r.Context(), contRequest.Name, template)
	if err != nil {
		log.Printf("createContainer error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if err := writeJSON(w, http.StatusCreated, id); err != nil {
		log.Printf("writeJSON error: %v", err)
	}

}

func (s *Server) getTemplates(w http.ResponseWriter, r *http.Request) {
	templates := make([]service.Template, 0, len(s.templates))

	for _, template := range s.templates {
		templates = append(templates, template)
	}

	if err := writeJSON(w, http.StatusOK, templates); err != nil {
		log.Printf("writeJSON error: %v", err)
	}
}

func (s *Server) getInstances(w http.ResponseWriter, r *http.Request) {
	instances, err := s.instances.ListInstances(r.Context())
	if err != nil {
		log.Printf("list instances error: %v", err)

		if err := writeJSON(w, http.StatusInternalServerError, "internal server error"); err != nil {
			log.Printf("write error response error: %v", err)
		}

		return
	}

	if err := writeJSON(w, http.StatusOK, instances); err != nil {
		log.Printf("write instances response error: %v", err)
	}
}

func (s *Server) getInstance(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid instance id", http.StatusBadRequest)
		return
	}

	instance, err := s.instances.GetInstance(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrInstanceNotFound) {
			if err := writeJSON(w, http.StatusNotFound, "instance not found"); err != nil {
				log.Printf("write not found response error: %v", err)
			}

			return
		}

		log.Printf("get instance %d error: %v", id, err)

		if err := writeJSON(w, http.StatusInternalServerError, "internal server error"); err != nil {
			log.Printf("write error response error: %v", err)
		}

		return
	}

	if err := writeJSON(w, http.StatusOK, instance); err != nil {
		log.Printf("write instance response error: %v", err)
	}
}

func (s *Server) createInstance(w http.ResponseWriter, r *http.Request) {
	var req CreateInstanceRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		if err := writeJSON(w, http.StatusBadRequest, "invalid request body"); err != nil {
			log.Printf("write error response error: %v", err)
		}

		return
	}

	if req.NodeID == 0 {
		if err := writeJSON(w, http.StatusBadRequest, "node_id is required"); err != nil {
			log.Printf("write error response error: %v", err)
		}

		return
	}

	if req.Name == "" {
		if err := writeJSON(w, http.StatusBadRequest, "name is required"); err != nil {
			log.Printf("write error response error: %v", err)
		}

		return
	}

	template, ok := s.templates[req.Template]
	if !ok {
		if err := writeJSON(w, http.StatusBadRequest, "template not found"); err != nil {
			log.Printf("write error response error: %v", err)
		}

		return
	}

	if _, err := s.nodes.Get(r.Context(), req.NodeID); err != nil {
		if errors.Is(err, node.ErrNotFound) {
			if err := writeJSON(w, http.StatusNotFound, "node not found"); err != nil {
				log.Printf("write error response error: %v", err)
			}

			return
		}

		log.Printf("get node %d error: %v", req.NodeID, err)

		if err := writeJSON(w, http.StatusInternalServerError, "internal server error"); err != nil {
			log.Printf("write error response error: %v", err)
		}

		return
	}

	containerID, err := s.docker.CreateContainer(r.Context(), req.Name, template)
	if err != nil {
		log.Printf("create container error: %v", err)

		if err := writeJSON(w, http.StatusInternalServerError, "failed to create container"); err != nil {
			log.Printf("write error response error: %v", err)
		}

		return
	}

	// сразу запускаем контейнер
	if err := s.docker.StartContainer(r.Context(), containerID); err != nil {
		log.Printf("start container %s error: %v", containerID, err)
	}

	instance := &Instance{
		NodeID:      req.NodeID,
		Name:        req.Name,
		Template:    req.Template,
		ContainerID: containerID,
	}

	if err := s.instances.CreateInstance(r.Context(), instance); err != nil {
		log.Printf("create instance error: %v", err)

		if err := s.docker.RemoveContainer(r.Context(), containerID); err != nil {
			log.Printf("remove orphan container %s error: %v", containerID, err)
		}

		if err := writeJSON(w, http.StatusInternalServerError, "failed to create instance"); err != nil {
			log.Printf("write error response error: %v", err)
		}

		return
	}

	if err := writeJSON(w, http.StatusCreated, instance); err != nil {
		log.Printf("write instance response error: %v", err)
	}
}

func (s *Server) deleteInstance(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid instance id", http.StatusBadRequest)
		return
	}

	instance, err := s.instances.GetInstance(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrInstanceNotFound) {
			if err := writeJSON(w, http.StatusNotFound, "instance not found"); err != nil {
				log.Printf("write not found response error: %v", err)
			}

			return
		}

		log.Printf("get instance %d error: %v", id, err)

		if err := writeJSON(w, http.StatusInternalServerError, "internal server error"); err != nil {
			log.Printf("write error response error: %v", err)
		}

		return
	}

	if err := s.instances.DeleteInstance(r.Context(), id); err != nil {
		if errors.Is(err, ErrInstanceNotFound) {
			if err := writeJSON(w, http.StatusNotFound, "instance not found"); err != nil {
				log.Printf("write not found response error: %v", err)
			}

			return
		}

		log.Printf("delete instance %d error: %v", id, err)

		if err := writeJSON(w, http.StatusInternalServerError, "internal server error"); err != nil {
			log.Printf("write error response error: %v", err)
		}

		return
	}

	if err := s.docker.RemoveContainer(r.Context(), instance.ContainerID); err != nil {
		log.Printf("remove container %s for instance %d error: %v", instance.ContainerID, id, err)
	}

	w.WriteHeader(http.StatusNoContent)
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
