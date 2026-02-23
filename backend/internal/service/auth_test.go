package service

import (
	"errors"
	"testing"

	"github.com/jiaxiang-medical-assistant/backend/internal/config"
)

func TestAuthServiceLoginAndVerifyToken(t *testing.T) {
	cfg := config.Config{
		Auth: config.AuthConfig{
			JWTSecret:      "test-secret",
			JWTExpiresIn:   3600,
			DoctorAccount:  "doctor",
			DoctorPassword: "dev",
			AdminAccount:   "admin",
			AdminPassword:  "admin123",
		},
	}

	svc := NewAuthService(cfg, "mock")
	result, err := svc.Login(LoginInput{
		Account:  "doctor",
		Password: "dev",
	})
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	if result.Token == "" {
		t.Fatalf("expected non-empty token")
	}

	claims, err := svc.VerifyToken(result.Token)
	if err != nil {
		t.Fatalf("verify token failed: %v", err)
	}

	if claims.Subject != "doctor" {
		t.Fatalf("unexpected subject: %s", claims.Subject)
	}
	if claims.Role != "doctor" {
		t.Fatalf("unexpected role: %s", claims.Role)
	}
}

func TestAuthServiceRejectsInvalidCredentials(t *testing.T) {
	cfg := config.Config{
		Auth: config.AuthConfig{
			JWTSecret:      "test-secret",
			DoctorAccount:  "doctor",
			DoctorPassword: "dev",
		},
	}

	svc := NewAuthService(cfg, "mock")
	_, err := svc.Login(LoginInput{
		Account:  "doctor",
		Password: "wrong-password",
	})
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}
