package service

import (
	"errors"
	"testing"

	"github.com/jiaxiang-medical-assistant/backend/internal/config"
)

func TestAuthServiceLoginAndVerifyToken(t *testing.T) {
	cfg := config.Config{
		Auth: config.AuthConfig{
			JWTSecret:       "test-secret",
			JWTExpiresIn:    3600,
			StudentAccount:  "student",
			StudentPassword: "student-password-123",
			DoctorAccount:   "doctor",
			DoctorPassword:  "doctor-password-123",
			AdminAccount:    "admin",
			AdminPassword:   "admin-password-123",
		},
	}

	svc, err := NewAuthService(cfg, "mock")
	if err != nil {
		t.Fatalf("new auth service failed: %v", err)
	}
	result, err := svc.Login(LoginInput{
		Account:  "doctor",
		Password: "doctor-password-123",
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
			JWTSecret:       "test-secret",
			StudentAccount:  "student",
			StudentPassword: "student-password-123",
			DoctorAccount:   "doctor",
			DoctorPassword:  "doctor-password-123",
			AdminAccount:    "admin",
			AdminPassword:   "admin-password-123",
		},
	}

	svc, err := NewAuthService(cfg, "mock")
	if err != nil {
		t.Fatalf("new auth service failed: %v", err)
	}
	_, err = svc.Login(LoginInput{
		Account:  "doctor",
		Password: "wrong-password",
	})
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestAuthServiceRejectsCollidingAccounts(t *testing.T) {
	cfg := config.Config{
		Auth: config.AuthConfig{
			JWTSecret:       "test-secret",
			StudentAccount:  "student",
			StudentPassword: "student-password-123",
			DoctorAccount:   "same-account",
			DoctorPassword:  "doctor-password-123",
			AdminAccount:    "same-account",
			AdminPassword:   "admin-password-123",
		},
	}

	_, err := NewAuthService(cfg, "mock")
	if err == nil {
		t.Fatalf("expected account collision error")
	}
}
