package main

import (
	"fmt"
	"log"

	"github.com/jiaxiang-medical-assistant/backend/internal/bootstrap"
	"github.com/jiaxiang-medical-assistant/backend/internal/config"
)

func main() {
	cfg := config.Load()

	engine, cleanup := bootstrap.NewServer(cfg)
	defer cleanup()

	addr := fmt.Sprintf(":%d", cfg.AppPort)
	log.Printf("starting %s on %s (%s)", cfg.AppName, addr, cfg.AppEnv)

	if err := engine.Run(addr); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
