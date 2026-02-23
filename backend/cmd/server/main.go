package main

import (
	"fmt"
	"log"

	"github.com/jiaxiang-medical-assistant/backend/internal/bootstrap"
	"github.com/jiaxiang-medical-assistant/backend/internal/config"
)

func main() {
	cfg := config.Load()

	engine, cleanup, err := bootstrap.NewServer(cfg)
	if err != nil {
		log.Fatalf("failed to initialize server: %v", err)
	}
	defer cleanup()

	addr := fmt.Sprintf(":%d", cfg.AppPort)
	log.Printf("starting %s on %s (%s)", cfg.AppName, addr, cfg.AppEnv)

	if err := engine.Run(addr); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
