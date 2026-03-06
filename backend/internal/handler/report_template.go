package handler

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jiaxiang-medical-assistant/backend/internal/response"
	"github.com/jiaxiang-medical-assistant/backend/internal/service"
)

type ReportTemplateHandler struct {
	svc       *service.ReportTemplateService
	outputDir string
}

func NewReportTemplateHandler(svc *service.ReportTemplateService) *ReportTemplateHandler {
	return &ReportTemplateHandler{svc: svc, outputDir: service.DefaultReportScheduleOutputDir()}
}

// ---- Templates ----

func (h *ReportTemplateHandler) CreateTemplate(c *gin.Context) {
	var req service.CreateTemplateInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, 1001, "请求参数无效")
		return
	}
	tpl, err := h.svc.CreateTemplate(c.Request.Context(), req)
	if err != nil {
		handleDomainError(c, err)
		return
	}
	response.Success(c, tpl)
}

func (h *ReportTemplateHandler) ListTemplates(c *gin.Context) {
	list, err := h.svc.ListTemplates(c.Request.Context())
	if err != nil {
		handleDomainError(c, err)
		return
	}
	response.Success(c, list)
}

func (h *ReportTemplateHandler) GetTemplate(c *gin.Context) {
	tpl, err := h.svc.GetTemplate(c.Request.Context(), c.Param("id"))
	if err != nil {
		handleDomainError(c, err)
		return
	}
	response.Success(c, tpl)
}

func (h *ReportTemplateHandler) UpdateTemplate(c *gin.Context) {
	var req service.UpdateTemplateInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, 1001, "请求参数无效")
		return
	}
	tpl, err := h.svc.UpdateTemplate(c.Request.Context(), c.Param("id"), req)
	if err != nil {
		handleDomainError(c, err)
		return
	}
	response.Success(c, tpl)
}

func (h *ReportTemplateHandler) DeleteTemplate(c *gin.Context) {
	if err := h.svc.DeleteTemplate(c.Request.Context(), c.Param("id")); err != nil {
		handleDomainError(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *ReportTemplateHandler) ExportWithTemplate(c *gin.Context) {
	result, err := h.svc.ExportWithTemplate(c.Request.Context(), c.Param("id"))
	if err != nil {
		handleDomainError(c, err)
		return
	}
	defer result.File.Close()

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, result.Filename))
	c.Header("Cache-Control", "no-cache")

	if err := result.File.Write(c.Writer); err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
	}
}

// ---- Column definitions (for frontend) ----

type columnOption struct {
	Key   string `json:"key"`
	Label string `json:"label"`
}

func (h *ReportTemplateHandler) ColumnOptions(c *gin.Context) {
	options := []columnOption{
		{"index", "序号"},
		{"student_name", "学生姓名"},
		{"class_name", "班级"},
		{"symptoms", "症状"},
		{"description", "描述"},
		{"diagnosis", "诊断"},
		{"prescription", "处方"},
		{"destination", "去向"},
		{"follow_up_at", "复诊时间"},
		{"follow_up_note", "复诊备注"},
		{"created_at", "就诊时间"},
	}
	response.Success(c, options)
}

// ---- Schedules ----

func (h *ReportTemplateHandler) CreateSchedule(c *gin.Context) {
	var req service.CreateScheduleInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, 1001, "请求参数无效")
		return
	}
	sched, err := h.svc.CreateSchedule(c.Request.Context(), req)
	if err != nil {
		handleDomainError(c, err)
		return
	}
	response.Success(c, sched)
}

func (h *ReportTemplateHandler) ListSchedules(c *gin.Context) {
	list, err := h.svc.ListSchedules(c.Request.Context())
	if err != nil {
		handleDomainError(c, err)
		return
	}
	response.Success(c, list)
}

func (h *ReportTemplateHandler) UpdateSchedule(c *gin.Context) {
	var req service.UpdateScheduleInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, 1001, "请求参数无效")
		return
	}
	sched, err := h.svc.UpdateSchedule(c.Request.Context(), c.Param("id"), req)
	if err != nil {
		handleDomainError(c, err)
		return
	}
	response.Success(c, sched)
}

func (h *ReportTemplateHandler) DeleteSchedule(c *gin.Context) {
	if err := h.svc.DeleteSchedule(c.Request.Context(), c.Param("id")); err != nil {
		handleDomainError(c, err)
		return
	}
	response.Success(c, nil)
}

// ---- Manual trigger ----

func (h *ReportTemplateHandler) TriggerSchedule(c *gin.Context) {
	result, err := h.svc.RunScheduleNow(c.Request.Context(), c.Param("id"), time.Now().UTC(), h.outputDir)
	if err != nil {
		handleDomainError(c, err)
		return
	}
	defer result.File.Close()

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, result.Filename))
	c.Header("Cache-Control", "no-cache")

	if err := result.File.Write(c.Writer); err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
	}
}

func (h *ReportTemplateHandler) ListScheduleFiles(c *gin.Context) {
	files, err := h.svc.ListScheduleFiles(c.Request.Context(), c.Param("id"), h.outputDir)
	if err != nil {
		handleDomainError(c, err)
		return
	}
	response.Success(c, files)
}

func (h *ReportTemplateHandler) DownloadScheduleFile(c *gin.Context) {
	fileName := strings.TrimPrefix(c.Param("filename"), "/")
	fullPath, err := h.svc.ResolveScheduleFile(c.Request.Context(), c.Param("id"), fileName, h.outputDir)
	if err != nil {
		handleDomainError(c, err)
		return
	}
	c.FileAttachment(fullPath, filepath.Base(fullPath))
}

