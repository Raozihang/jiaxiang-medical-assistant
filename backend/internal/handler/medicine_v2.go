package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jiaxiang-medical-assistant/backend/internal/response"
	"github.com/jiaxiang-medical-assistant/backend/internal/service"
)

type MedicineHandler struct {
	medicineService *service.MedicineService
}

type StockChangeRequest struct {
	MedicineID string `json:"medicine_id" binding:"required"`
	Quantity   int    `json:"quantity" binding:"required"`
}

func NewMedicineHandler(medicineService *service.MedicineService) *MedicineHandler {
	return &MedicineHandler{medicineService: medicineService}
}

func (h *MedicineHandler) List(c *gin.Context) {
	pageParams := service.ParsePageParams(c)
	result, err := h.medicineService.List(c.Request.Context(), service.MedicineListInput{
		PageParams: pageParams,
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

func (h *MedicineHandler) Inbound(c *gin.Context) {
	var req StockChangeRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Quantity <= 0 {
		response.Fail(c, http.StatusBadRequest, 1001, "请求参数无效")
		return
	}

	result, err := h.medicineService.Inbound(c.Request.Context(), service.StockChangeInput{
		MedicineID: req.MedicineID,
		Quantity:   req.Quantity,
	})
	if err != nil {
		handleDomainError(c, err)
		return
	}

	response.Success(c, result)
}

func (h *MedicineHandler) Outbound(c *gin.Context) {
	var req StockChangeRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Quantity <= 0 {
		response.Fail(c, http.StatusBadRequest, 1001, "请求参数无效")
		return
	}

	result, err := h.medicineService.Outbound(c.Request.Context(), service.StockChangeInput{
		MedicineID: req.MedicineID,
		Quantity:   req.Quantity,
	})
	if err != nil {
		handleDomainError(c, err)
		return
	}

	response.Success(c, result)
}
