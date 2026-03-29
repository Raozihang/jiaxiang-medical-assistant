package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jiaxiang-medical-assistant/backend/internal/response"
	"github.com/jiaxiang-medical-assistant/backend/internal/service"
)

type AuthHandler struct {
	authService *service.AuthService
}

type LoginRequest struct {
	Account  string `json:"account" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, 1001, "请求参数无效")
		return
	}

	result, err := h.authService.Login(service.LoginInput{
		Account:  req.Account,
		Password: req.Password,
	})
	if errors.Is(err, service.ErrInvalidCredentials) {
		response.Fail(c, http.StatusUnauthorized, 1002, "账号或密码错误")
		return
	}
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, 5000, "服务内部错误")
		return
	}

	response.Success(c, result)
}
