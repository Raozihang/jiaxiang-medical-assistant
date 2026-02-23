package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/jiaxiang-medical-assistant/backend/internal/config"
	"github.com/jiaxiang-medical-assistant/backend/internal/response"
)

type HealthHandler struct {
	cfg      config.Config
	dataMode string
}

func NewHealthHandler(cfg config.Config, dataMode string) *HealthHandler {
	return &HealthHandler{cfg: cfg, dataMode: dataMode}
}

func (h *HealthHandler) Healthz(c *gin.Context) {
	response.Success(c, gin.H{
		"service": h.cfg.AppName,
		"env":     h.cfg.AppEnv,
		"mode":    h.dataMode,
		"status":  "ok",
	})
}
