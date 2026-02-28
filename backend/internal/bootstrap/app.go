package bootstrap

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/jiaxiang-medical-assistant/backend/internal/config"
	"github.com/jiaxiang-medical-assistant/backend/internal/middleware"
)

func NewServer(cfg config.Config) (*gin.Engine, func(), error) {
	if err := cfg.Validate(); err != nil {
		return nil, func() {}, fmt.Errorf("配置无效: %w", err)
	}

	gin.SetMode(resolveGinMode(cfg.AppEnv))

	engine := gin.New()
	engine.Use(
		middleware.CORS(),
		middleware.RequestID(),
		middleware.AccessLog(),
		gin.Recovery(),
	)

	database, cleanupDB := InitDatabase(cfg)
	if err := registerRoutes(engine, cfg, database); err != nil {
		cleanupDB()
		return nil, func() {}, fmt.Errorf("路由注册失败: %w", err)
	}

	return engine, cleanupDB, nil
}

func resolveGinMode(appEnv string) string {
	if appEnv == "production" {
		return gin.ReleaseMode
	}

	return gin.DebugMode
}
