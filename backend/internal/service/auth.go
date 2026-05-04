package service

import (
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jiaxiang-medical-assistant/backend/internal/config"
)

var ErrInvalidCredentials = errors.New("账号或密码错误")

type AuthService struct {
	dataMode    string
	secret      []byte
	expiresIn   int
	credentials map[string]credential
}

type credential struct {
	password string
	role     string
	name     string
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

type TokenClaims struct {
	Name string `json:"name"`
	Role string `json:"role"`
	jwt.RegisteredClaims
}

func NewAuthService(cfg config.Config, dataMode string) (*AuthService, error) {
	studentAccount := strings.TrimSpace(cfg.Auth.StudentAccount)
	doctorAccount := strings.TrimSpace(cfg.Auth.DoctorAccount)
	adminAccount := strings.TrimSpace(cfg.Auth.AdminAccount)
	if studentAccount == "" || doctorAccount == "" || adminAccount == "" {
		return nil, errors.New("学生、医生和管理员账号配置不能为空")
	}
	if hasDuplicateAccounts(studentAccount, doctorAccount, adminAccount) {
		return nil, errors.New("学生、医生和管理员账号不能相同")
	}

	studentPassword := strings.TrimSpace(cfg.Auth.StudentPassword)
	doctorPassword := strings.TrimSpace(cfg.Auth.DoctorPassword)
	adminPassword := strings.TrimSpace(cfg.Auth.AdminPassword)
	if studentPassword == "" || doctorPassword == "" || adminPassword == "" {
		return nil, errors.New("学生、医生和管理员密码配置不能为空")
	}

	secret := []byte(strings.TrimSpace(cfg.Auth.JWTSecret))
	if len(secret) == 0 {
		return nil, errors.New("JWT 密钥配置不能为空")
	}

	expiresIn := cfg.Auth.JWTExpiresIn
	if expiresIn <= 0 {
		expiresIn = 7200
	}

	return &AuthService{
		dataMode:  dataMode,
		secret:    secret,
		expiresIn: expiresIn,
		credentials: map[string]credential{
			studentAccount: {
				password: studentPassword,
				role:     "student",
				name:     studentAccount,
			},
			doctorAccount: {
				password: doctorPassword,
				role:     "doctor",
				name:     doctorAccount,
			},
			adminAccount: {
				password: adminPassword,
				role:     "admin",
				name:     adminAccount,
			},
		},
	}, nil
}

func hasDuplicateAccounts(accounts ...string) bool {
	seen := make(map[string]struct{}, len(accounts))
	for _, account := range accounts {
		if _, ok := seen[account]; ok {
			return true
		}
		seen[account] = struct{}{}
	}
	return false
}

func (s *AuthService) Login(input LoginInput) (LoginResult, error) {
	account := strings.TrimSpace(input.Account)
	password := strings.TrimSpace(input.Password)

	cred, ok := s.credentials[account]
	if !ok || cred.password != password {
		return LoginResult{}, ErrInvalidCredentials
	}

	now := time.Now().UTC()
	expiresAt := now.Add(time.Duration(s.expiresIn) * time.Second)
	claims := TokenClaims{
		Name: cred.name,
		Role: cred.role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   account,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.secret)
	if err != nil {
		return LoginResult{}, err
	}

	result := LoginResult{
		Token:     token,
		ExpiresIn: s.expiresIn,
		DataMode:  s.dataMode,
		IssuedAt:  now.Format(time.RFC3339),
	}
	result.User.Name = cred.name
	result.User.Role = cred.role

	return result, nil
}

func (s *AuthService) VerifyToken(token string) (TokenClaims, error) {
	parser := jwt.NewParser(jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	parsed, err := parser.ParseWithClaims(token, &TokenClaims{}, func(_ *jwt.Token) (any, error) {
		return s.secret, nil
	})
	if err != nil {
		return TokenClaims{}, err
	}

	claims, ok := parsed.Claims.(*TokenClaims)
	if !ok || !parsed.Valid {
		return TokenClaims{}, errors.New("invalid token")
	}

	return *claims, nil
}
