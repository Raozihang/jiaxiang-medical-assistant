package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jiaxiang-medical-assistant/backend/internal/repository"
	"github.com/jiaxiang-medical-assistant/backend/internal/response"
	"github.com/jiaxiang-medical-assistant/backend/internal/service"
)

type VisitHandler struct {
	visitService *service.VisitService
}

type CreateVisitRequest struct {
	StudentID   string   `json:"student_id" binding:"required"`
	Symptoms    []string `json:"symptoms"`
	Description string   `json:"description"`
}

type UpdateVisitRequest struct {
	Diagnosis    *string  `json:"diagnosis"`
	Prescription []string `json:"prescription"`
	Destination  *string  `json:"destination"`
}

func NewVisitHandler(visitService *service.VisitService) *VisitHandler {
	return &VisitHandler{visitService: visitService}
}

func (h *VisitHandler) List(c *gin.Context) {
	pageParams := service.ParsePageParams(c)
	result, err := h.visitService.List(c.Request.Context(), service.VisitListInput{
		PageParams: pageParams,
		StudentID:  c.Query("student_id"),
	})
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

func (h *VisitHandler) Create(c *gin.Context) {
	var req CreateVisitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, 1001, "invalid request body")
		return
	}
	if strings.TrimSpace(req.StudentID) == "" {
		response.Fail(c, http.StatusBadRequest, 1001, "student_id is required")
		return
	}

	visit, err := h.visitService.Create(c.Request.Context(), service.CreateVisitInput{
		StudentID:   req.StudentID,
		Symptoms:    req.Symptoms,
		Description: req.Description,
	})
	if err != nil {
		handleDomainError(c, err)
		return
	}

	response.Success(c, visit)
}

func (h *VisitHandler) Detail(c *gin.Context) {
	visit, err := h.visitService.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		handleDomainError(c, err)
		return
	}

	response.Success(c, visit)
}

func (h *VisitHandler) Update(c *gin.Context) {
	var req UpdateVisitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, 1001, "invalid request body")
		return
	}

	var prescription *[]string
	if req.Prescription != nil {
		prescription = &req.Prescription
	}

	visit, err := h.visitService.Update(c.Request.Context(), c.Param("id"), service.UpdateVisitInput{
		Diagnosis:    req.Diagnosis,
		Prescription: prescription,
		Destination:  req.Destination,
	})
	if err != nil {
		handleDomainError(c, err)
		return
	}

	response.Success(c, visit)
}

func handleDomainError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		response.Fail(c, http.StatusNotFound, 2001, "resource not found")
	case errors.Is(err, repository.ErrInsufficientStock):
		response.Fail(c, http.StatusBadRequest, 3001, "insufficient stock")
	default:
		response.Fail(c, http.StatusInternalServerError, 5000, "internal error")
	}
}
