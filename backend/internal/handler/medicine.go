//go:build ignore

package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jiaxiang-medical-assistant/backend/internal/response"
	"gorm.io/gorm"
)

type MedicineHandler struct {
	db *gorm.DB
}

type StockChangeRequest struct {
	MedicineID string `json:"medicine_id" binding:"required"`
	Quantity   int    `json:"quantity" binding:"required"`
}

func NewMedicineHandler(db *gorm.DB) *MedicineHandler {
	return &MedicineHandler{db: db}
}

func (h *MedicineHandler) List(c *gin.Context) {
	response.Success(c, []gin.H{
		{"id": "med-001", "name": "布洛芬", "stock": 120},
		{"id": "med-002", "name": "医用纱布", "stock": 30},
	})
}

func (h *MedicineHandler) Inbound(c *gin.Context) {
	var req StockChangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, 1001, "invalid request body")
		return
	}

	response.Success(c, gin.H{
		"medicine_id": req.MedicineID,
		"quantity":    req.Quantity,
		"action":      "inbound",
		"db_ready":    h.db != nil,
	})
}

func (h *MedicineHandler) Outbound(c *gin.Context) {
	var req StockChangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, 1001, "invalid request body")
		return
	}

	response.Success(c, gin.H{
		"medicine_id": req.MedicineID,
		"quantity":    req.Quantity,
		"action":      "outbound",
		"db_ready":    h.db != nil,
	})
}
