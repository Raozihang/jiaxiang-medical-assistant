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
		log.Fatalf("服务初始化失败: %v", err)
	}
	defer cleanup()

	addr := fmt.Sprintf(":%d", cfg.AppPort)
	log.Printf("正在启动 %s，监听 %s（%s）", cfg.AppName, addr, cfg.AppEnv)

	if err := engine.Run(addr); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
