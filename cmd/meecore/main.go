package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nxwex/meecore/internal/config"
	"github.com/nxwex/meecore/internal/docker"
	"github.com/nxwex/meecore/internal/node"
	"github.com/nxwex/meecore/internal/server"
	"github.com/nxwex/meecore/internal/storage/postgres"
)

func main() {
	cfg := config.Load()

	ctx := context.Background()

	db, err := postgres.New(ctx, cfg.DatabaseDSN)
	if err != nil {
		log.Fatalf("db error: %v", err)
	}
	defer db.Close()

	nodes := node.NewService(db)

	dockerClient, err := docker.New(ctx)
	if err != nil {
		log.Fatalf("docker error: %v", err)
	}
	defer dockerClient.Close()

	httpServer := server.New(cfg.HTTPAddr, nodes, dockerClient)

	serverErr := make(chan error, 1)

	go func() {
		log.Printf("meecore started on %s", cfg.HTTPAddr)
		serverErr <- httpServer.Start()
	}()

	signalCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case err := <-serverErr:
		log.Fatalf("server error: %v", err)
	case <-signalCtx.Done():
		log.Println("shutting down meecore")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("http server shutdown: %v", err)
	}

	log.Println("meecore stopped")
}
