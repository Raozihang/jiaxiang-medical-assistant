package handler

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jiaxiang-medical-assistant/backend/internal/response"
	"github.com/jiaxiang-medical-assistant/backend/internal/service"
)

type OutboundCallHandler struct {
	outboundCallService *service.OutboundCallService
	callbackSecret      string
}

func NewOutboundCallHandler(outboundCallService *service.OutboundCallService, callbackSecret string) *OutboundCallHandler {
	return &OutboundCallHandler{
		outboundCallService: outboundCallService,
		callbackSecret:      strings.TrimSpace(callbackSecret),
	}
}

func (h *OutboundCallHandler) List(c *gin.Context) {
	pageParams := service.ParsePageParams(c)
	result, err := h.outboundCallService.List(c.Request.Context(), service.OutboundCallListInput{
		PageParams: pageParams,
		Status:     c.Query("status"),
		StudentID:  c.Query("student_id"),
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

func (h *OutboundCallHandler) Retry(c *gin.Context) {
	call, err := h.outboundCallService.Retry(c.Request.Context(), c.Param("id"))
	if err != nil {
		handleDomainError(c, err)
		return
	}

	response.Success(c, call)
}

func (h *OutboundCallHandler) AliyunCallback(c *gin.Context) {
	if !h.authorizeCallback(c.GetHeader("X-Webhook-Secret")) {
		response.Fail(c, http.StatusUnauthorized, 1003, "callback authorization failed")
		return
	}

	var req service.AliyunCallbackInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, 1001, "invalid request body")
		return
	}

	call, err := h.outboundCallService.HandleAliyunCallback(c.Request.Context(), req)
	if err != nil {
		handleDomainError(c, err)
		return
	}

	response.Success(c, call)
}

func (h *OutboundCallHandler) authorizeCallback(secret string) bool {
	expected := strings.TrimSpace(h.callbackSecret)
	provided := strings.TrimSpace(secret)
	if expected == "" || provided == "" {
		return false
	}

	return subtle.ConstantTimeCompare([]byte(expected), []byte(provided)) == 1
}
