package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jiaxiang-medical-assistant/backend/internal/response"
	"github.com/jiaxiang-medical-assistant/backend/internal/service"
)

type AIHandler struct {
	aiService *service.AIService
}

func NewAIHandler(aiService *service.AIService) *AIHandler {
	return &AIHandler{aiService: aiService}
}

func (h *AIHandler) Analyze(c *gin.Context) {
	var req service.AnalyzeInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, 1001, "请求参数无效")
		return
	}

	result, err := h.aiService.Analyze(c.Request.Context(), req)
	if err != nil {
		handleDomainError(c, err)
		return
	}

	response.Success(c, result)
}

func (h *AIHandler) Triage(c *gin.Context) {
	var req service.TriageInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, 1001, "请求参数无效")
		return
	}

	result, err := h.aiService.Triage(c.Request.Context(), req)
	if err != nil {
		handleDomainError(c, err)
		return
	}

	response.Success(c, result)
}

func (h *AIHandler) Recommend(c *gin.Context) {
	var req service.RecommendInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, 1001, "请求参数无效")
		return
	}

	result, err := h.aiService.Recommend(c.Request.Context(), req)
	if err != nil {
		handleDomainError(c, err)
		return
	}

	response.Success(c, result)
}

func (h *AIHandler) InteractionCheck(c *gin.Context) {
	var req service.InteractionCheckInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, 1001, "请求参数无效")
		return
	}

	result, err := h.aiService.InteractionCheck(c.Request.Context(), req)
	if err != nil {
		handleDomainError(c, err)
		return
	}

	response.Success(c, result)
}
