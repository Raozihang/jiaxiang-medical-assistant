package response

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jiaxiang-medical-assistant/backend/internal/middleware"
)

type Envelope struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	Data      any    `json:"data"`
	RequestID string `json:"request_id"`
	Timestamp string `json:"timestamp"`
}

func Success(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Envelope{
		Code:      0,
		Message:   "ok",
		Data:      data,
		RequestID: middleware.GetRequestID(c),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
}

func Fail(c *gin.Context, statusCode int, code int, message string) {
	c.JSON(statusCode, Envelope{
		Code:      code,
		Message:   message,
		Data:      gin.H{},
		RequestID: middleware.GetRequestID(c),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
}
