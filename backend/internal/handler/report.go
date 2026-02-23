package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/jiaxiang-medical-assistant/backend/internal/response"
	"github.com/jiaxiang-medical-assistant/backend/internal/service"
)

type ReportHandler struct {
	reportService *service.ReportService
}

func NewReportHandler(reportService *service.ReportService) *ReportHandler {
	return &ReportHandler{reportService: reportService}
}

func (h *ReportHandler) Overview(c *gin.Context) {
	overview, err := h.reportService.Overview(c.Request.Context())
	if err != nil {
		handleDomainError(c, err)
		return
	}

	response.Success(c, overview)
}

func (h *ReportHandler) Daily(c *gin.Context) {
	report, err := h.reportService.Daily(c.Request.Context())
	if err != nil {
		handleDomainError(c, err)
		return
	}

	response.Success(c, report)
}

func (h *ReportHandler) Weekly(c *gin.Context) {
	report, err := h.reportService.Weekly(c.Request.Context())
	if err != nil {
		handleDomainError(c, err)
		return
	}

	response.Success(c, report)
}

func (h *ReportHandler) Monthly(c *gin.Context) {
	report, err := h.reportService.Monthly(c.Request.Context())
	if err != nil {
		handleDomainError(c, err)
		return
	}

	response.Success(c, report)
}
