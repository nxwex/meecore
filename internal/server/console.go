package server

import (
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/go-chi/chi"
	"github.com/gorilla/websocket"
	"github.com/moby/moby/api/pkg/stdcopy"
)

var consoleUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type consoleWriter struct {
	mu   sync.Mutex
	conn *websocket.Conn
}

func (w *consoleWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if err := w.conn.WriteMessage(websocket.TextMessage, p); err != nil {
		return 0, err
	}

	return len(p), nil
}

func (s *Server) instanceConsole(w http.ResponseWriter, r *http.Request) {
	instanceID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid instance id", http.StatusBadRequest)
		return
	}

	instance, err := s.instances.GetInstance(r.Context(), instanceID)
	if err != nil {
		if errors.Is(err, ErrInstanceNotFound) {
			http.Error(w, "instance not found", http.StatusNotFound)
			return
		}

		log.Printf("get instance for console error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	conn, err := consoleUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("upgrade console websocket error: %v", err)
		return
	}
	defer conn.Close()

	attach, err := s.docker.AttachContainer(
		r.Context(),
		instance.ContainerID,
	)
	if err != nil {
		log.Printf(
			"attach container %s for instance %d error: %v",
			instance.ContainerID,
			instance.ID,
			err,
		)
		return
	}
	defer attach.Close()

	writer := &consoleWriter{
		conn: conn,
	}

	errCh := make(chan error, 2)

	go func() {
		_, err := stdcopy.StdCopy(
			writer,
			writer,
			attach.Reader,
		)

		errCh <- err
	}()

	go func() {
		for {
			messageType, message, err := conn.ReadMessage()
			if err != nil {
				errCh <- err
				return
			}

			if messageType != websocket.TextMessage {
				continue
			}

			command := strings.TrimSpace(string(message))
			if command == "" {
				continue
			}

			log.Printf(
				"console command for instance %d: %q",
				instance.ID,
				command,
			)

			output, err := s.docker.ExecuteCommand(
				r.Context(),
				instance.ContainerID,
				command,
			)
			if err != nil {
				log.Printf(
					"execute command for instance %d error: %v",
					instance.ID,
					err,
				)

				if err := conn.WriteMessage(
					websocket.TextMessage,
					[]byte("\n[command error] "+err.Error()+"\n"),
				); err != nil {
					errCh <- err
					return
				}

				continue
			}

			if output == "" {
				continue
			}

			if err := conn.WriteMessage(
				websocket.TextMessage,
				[]byte(output),
			); err != nil {
				errCh <- err
				return
			}
		}
	}()

	if err := <-errCh; err != nil && !errors.Is(err, io.EOF) {
		log.Printf(
			"instance %d console error: %v",
			instance.ID,
			err,
		)
	}
}
