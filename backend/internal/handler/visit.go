//go:build ignore

package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jiaxiang-medical-assistant/backend/internal/response"
	"gorm.io/gorm"
)

type VisitHandler struct {
	db *gorm.DB
}

type CreateVisitRequest struct {
	StudentID string `json:"student_id" binding:"required"`
	Symptoms  string `json:"symptoms"`
}

func NewVisitHandler(db *gorm.DB) *VisitHandler {
	return &VisitHandler{db: db}
}

func (h *VisitHandler) List(c *gin.Context) {
	response.Success(c, []gin.H{
		{
			"id":         "visit-001",
			"student_id": "20260001",
			"symptoms":   "发热",
			"created_at": time.Now().Add(-10 * time.Minute).Format(time.RFC3339),
		},
	})
}

func (h *VisitHandler) Create(c *gin.Context) {
	var req CreateVisitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, 1001, "invalid request body")
		return
	}

	response.Success(c, gin.H{
		"id":         uuid.NewString(),
		"student_id": req.StudentID,
		"symptoms":   req.Symptoms,
		"created_at": time.Now().UTC().Format(time.RFC3339),
		"db_ready":   h.db != nil,
	})
}

func (h *VisitHandler) Detail(c *gin.Context) {
	response.Success(c, gin.H{
		"id":         c.Param("id"),
		"student_id": "20260001",
		"diagnosis":  "普通感冒",
		"db_ready":   h.db != nil,
	})
}

func (h *VisitHandler) Update(c *gin.Context) {
	response.Success(c, gin.H{
		"id":       c.Param("id"),
		"updated":  true,
		"db_ready": h.db != nil,
	})
}
