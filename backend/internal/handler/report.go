package handler

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

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

func (h *ReportHandler) ExportDaily(c *gin.Context) {
	h.serveExcel(c, h.reportService.ExportDaily)
}

func (h *ReportHandler) ExportWeekly(c *gin.Context) {
	h.serveExcel(c, h.reportService.ExportWeekly)
}

func (h *ReportHandler) ExportMonthly(c *gin.Context) {
	h.serveExcel(c, h.reportService.ExportMonthly)
}

func (h *ReportHandler) serveExcel(c *gin.Context, exportFn func(ctx context.Context) (*service.ExcelExportResult, error)) {
	result, err := exportFn(c.Request.Context())
	if err != nil {
		handleDomainError(c, err)
		return
	}
	defer result.File.Close()

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"; filename*=UTF-8''%s`, result.Filename, url.PathEscape(result.Filename)))
	c.Header("Cache-Control", "no-cache")

	if err := result.File.Write(c.Writer); err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
	}
}
