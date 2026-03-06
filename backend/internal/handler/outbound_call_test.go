package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jiaxiang-medical-assistant/backend/internal/repository"
	"github.com/jiaxiang-medical-assistant/backend/internal/service"
)

func TestOutboundCallHandlerAliyunCallbackRequiresSecret(t *testing.T) {
	gin.SetMode(gin.TestMode)

	visitRepo := repository.NewMockVisitRepository()
	contactRepo := repository.NewMemoryStudentContactRepository()
	callRepo := repository.NewMemoryOutboundCallRepository()
	callService := service.NewOutboundCallService(callRepo, visitRepo, contactRepo, service.NewMockOutboundCallProvider(), "external_medical_followup")
	handler := NewOutboundCallHandler(callService, "top-secret")

	engine := gin.New()
	engine.POST("/callback", handler.AliyunCallback)

	body, _ := json.Marshal(gin.H{
		"request_id": "mock-1",
		"status":     "connected",
	})
	req := httptest.NewRequest(http.MethodPost, "/callback", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	engine.ServeHTTP(resp, req)

	if resp.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, resp.Code)
	}
}
