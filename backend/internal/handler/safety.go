package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jiaxiang-medical-assistant/backend/internal/response"
	"github.com/jiaxiang-medical-assistant/backend/internal/service"
)

type SafetyHandler struct {
	safetyService *service.SafetyService
}

type UpdateSafetyAlertRequest struct {
	Status string `json:"status"`
}

func NewSafetyHandler(safetyService *service.SafetyService) *SafetyHandler {
	return &SafetyHandler{safetyService: safetyService}
}

func (h *SafetyHandler) Alerts(c *gin.Context) {
	pageParams := service.ParsePageParams(c)
	alerts, err := h.safetyService.ListAlerts(c.Request.Context(), c.Query("status"))
	if err != nil {
		handleDomainError(c, err)
		return
	}

	start, end := paginateWindow(pageParams.Page, pageParams.PageSize, len(alerts))
	response.Success(c, gin.H{
		"items":     alerts[start:end],
		"page":      pageParams.Page,
		"page_size": pageParams.PageSize,
		"total":     len(alerts),
	})
}

func (h *SafetyHandler) UpdateAlert(c *gin.Context) {
	var req UpdateSafetyAlertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, 1001, "请求参数无效")
		return
	}

	if strings.ToLower(strings.TrimSpace(req.Status)) != "resolved" {
		response.Fail(c, http.StatusBadRequest, 1001, "状态必须为 resolved")
		return
	}

	alert, err := h.safetyService.ResolveAlert(c.Request.Context(), c.Param("id"))
	if err != nil {
		handleDomainError(c, err)
		return
	}

	response.Success(c, alert)
}
