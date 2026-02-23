package service

import (
	"strings"
	"time"
)

type AuthService struct {
	dataMode string
}

type LoginInput struct {
	Account  string
	Password string
}

type LoginResult struct {
	Token     string `json:"token"`
	ExpiresIn int    `json:"expires_in"`
	User      struct {
		Name string `json:"name"`
		Role string `json:"role"`
	} `json:"user"`
	DataMode string `json:"data_mode"`
	IssuedAt string `json:"issued_at"`
}

func NewAuthService(dataMode string) *AuthService {
	return &AuthService{dataMode: dataMode}
}

func (s *AuthService) Login(input LoginInput) LoginResult {
	name := strings.TrimSpace(input.Account)
	if name == "" {
		name = "doctor"
	}

	result := LoginResult{
		Token:     "mock-token-for-dev",
		ExpiresIn: 7200,
		DataMode:  s.dataMode,
		IssuedAt:  time.Now().UTC().Format(time.RFC3339),
	}
	result.User.Name = name
	result.User.Role = "doctor"

	return result
}
