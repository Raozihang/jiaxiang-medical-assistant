package config

import "testing"

func TestAuthConfigValidationRejectsUnsafeDefaults(t *testing.T) {
	cfg := AuthConfig{
		JWTSecret:       "replace-with-a-long-random-secret",
		StudentAccount:  "student",
		StudentPassword: "replace-with-student-password",
		DoctorAccount:   "doctor",
		DoctorPassword:  "replace-with-doctor-password",
		AdminAccount:    "admin",
		AdminPassword:   "replace-with-admin-password",
	}

	if err := cfg.Validate(); err == nil {
		t.Fatalf("expected validation error for placeholder credentials")
	}
}

func TestAuthConfigValidationRejectsAccountCollision(t *testing.T) {
	cfg := AuthConfig{
		JWTSecret:       "very-strong-secret",
		StudentAccount:  "student",
		StudentPassword: "student-password",
		DoctorAccount:   "same",
		DoctorPassword:  "doctor-password",
		AdminAccount:    "same",
		AdminPassword:   "admin-password",
	}

	if err := cfg.Validate(); err == nil {
		t.Fatalf("expected account collision validation error")
	}
}

func TestAuthConfigValidationAcceptsSecureValues(t *testing.T) {
	cfg := AuthConfig{
		JWTSecret:       "very-strong-secret",
		StudentAccount:  "student",
		StudentPassword: "student-password",
		DoctorAccount:   "doctor",
		DoctorPassword:  "doctor-password",
		AdminAccount:    "admin",
		AdminPassword:   "admin-password",
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}
}
