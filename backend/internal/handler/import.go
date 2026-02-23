package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jiaxiang-medical-assistant/backend/internal/response"
	"github.com/jiaxiang-medical-assistant/backend/internal/service"
)

type ImportHandler struct {
	importService *service.ImportService
}

func NewImportHandler(importService *service.ImportService) *ImportHandler {
	return &ImportHandler{importService: importService}
}

func (h *ImportHandler) ImportVisits(c *gin.Context) {
	var req []service.VisitImportItem
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, 1001, "invalid request body")
		return
	}

	task, err := h.importService.SubmitVisits(c.Request.Context(), req)
	if err != nil {
		handleDomainError(c, err)
		return
	}

	response.Success(c, task)
}

func (h *ImportHandler) TaskDetail(c *gin.Context) {
	task, err := h.importService.GetTask(c.Request.Context(), c.Param("id"))
	if err != nil {
		handleDomainError(c, err)
		return
	}

	response.Success(c, task)
}

func (h *ImportHandler) Tasks(c *gin.Context) {
	pageParams := service.ParsePageParams(c)
	result, err := h.importService.ListTasks(c.Request.Context(), pageParams)
	if err != nil {
		handleDomainError(c, err)
		return
	}

	response.Success(c, gin.H{
		"items":     result.Items,
		"page":      result.Page,
		"page_size": result.PageSize,
		"total":     result.Total,
	})
}
