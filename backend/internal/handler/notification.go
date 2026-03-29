package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jiaxiang-medical-assistant/backend/internal/response"
	"github.com/jiaxiang-medical-assistant/backend/internal/service"
)

type NotificationHandler struct {
	notificationService *service.NotificationService
}

func NewNotificationHandler(notificationService *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{notificationService: notificationService}
}

func (h *NotificationHandler) Send(c *gin.Context) {
	var req service.SendNotificationInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, 1001, "请求参数无效")
		return
	}

	log, err := h.notificationService.Send(c.Request.Context(), req)
	if err != nil {
		handleDomainError(c, err)
		return
	}

	response.Success(c, log)
}

func (h *NotificationHandler) Dispatch(c *gin.Context) {
	var req service.DispatchScenarioInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, 1001, "请求参数无效")
		return
	}

	log, err := h.notificationService.DispatchScenario(c.Request.Context(), req)
	if err != nil {
		handleDomainError(c, err)
		return
	}

	response.Success(c, log)
}

func (h *NotificationHandler) Logs(c *gin.Context) {
	pageParams := service.ParsePageParams(c)
	logs, err := h.notificationService.ListLogs(c.Request.Context())
	if err != nil {
		handleDomainError(c, err)
		return
	}

	start, end := paginateWindow(pageParams.Page, pageParams.PageSize, len(logs))
	response.Success(c, gin.H{
		"items":     logs[start:end],
		"page":      pageParams.Page,
		"page_size": pageParams.PageSize,
		"total":     len(logs),
	})
}

func paginateWindow(page int, pageSize int, total int) (int, int) {
	if total == 0 {
		return 0, 0
	}

	start := (page - 1) * pageSize
	if start >= total {
		return total, total
	}

	end := start + pageSize
	if end > total {
		end = total
	}

	return start, end
}
