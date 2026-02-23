package bootstrap

import (
	"github.com/gin-gonic/gin"
	"github.com/jiaxiang-medical-assistant/backend/internal/config"
	"github.com/jiaxiang-medical-assistant/backend/internal/middleware"
)

func NewServer(cfg config.Config) (*gin.Engine, func()) {
	gin.SetMode(resolveGinMode(cfg.AppEnv))

	engine := gin.New()
	engine.Use(
		middleware.CORS(),
		middleware.RequestID(),
		middleware.AccessLog(),
		gin.Recovery(),
	)

	database, cleanupDB := InitDatabase(cfg)
	registerRoutes(engine, cfg, database)

	return engine, cleanupDB
}

func resolveGinMode(appEnv string) string {
	if appEnv == "production" {
		return gin.ReleaseMode
	}

	return gin.DebugMode
}
