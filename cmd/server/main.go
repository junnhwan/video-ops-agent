package main

import (
	"errors"
	"log"
	"net/http"
	"os"

	"video-ops-agent/internal/config"
	httpapi "video-ops-agent/internal/http"
)

func main() {
	cfg, err := config.Load(os.Getenv("CONFIG_PATH"))
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	server := &http.Server{
		Addr:    cfg.Server.Address,
		Handler: httpapi.NewRouter(),
	}

	log.Printf("video-ops-agent listening on %s", cfg.Server.Address)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("serve: %v", err)
	}
}
