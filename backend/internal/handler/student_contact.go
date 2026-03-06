package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jiaxiang-medical-assistant/backend/internal/response"
	"github.com/jiaxiang-medical-assistant/backend/internal/service"
)

type StudentContactHandler struct {
	studentContactService *service.StudentContactService
}

type UpdateStudentContactRequest struct {
	StudentName      *string `json:"student_name"`
	GuardianName     *string `json:"guardian_name"`
	GuardianPhone    *string `json:"guardian_phone"`
	GuardianRelation *string `json:"guardian_relation"`
}

func NewStudentContactHandler(studentContactService *service.StudentContactService) *StudentContactHandler {
	return &StudentContactHandler{studentContactService: studentContactService}
}

func (h *StudentContactHandler) List(c *gin.Context) {
	pageParams := service.ParsePageParams(c)
	result, err := h.studentContactService.List(c.Request.Context(), service.StudentContactListInput{
		PageParams: pageParams,
		Keyword:    c.Query("keyword"),
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

func (h *StudentContactHandler) Update(c *gin.Context) {
	var req UpdateStudentContactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, 1001, "请求参数无效")
		return
	}

	contact, err := h.studentContactService.UpdateByStudentID(c.Request.Context(), c.Param("studentId"), service.UpdateStudentContactInput{
		StudentName:      req.StudentName,
		GuardianName:     req.GuardianName,
		GuardianPhone:    req.GuardianPhone,
		GuardianRelation: req.GuardianRelation,
	})
	if err != nil {
		handleDomainError(c, err)
		return
	}

	response.Success(c, contact)
}
