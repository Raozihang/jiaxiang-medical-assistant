package service

import (
	"context"
	"strings"
	"time"

	"github.com/jiaxiang-medical-assistant/backend/internal/repository"
)

type MedicineService struct {
	repo repository.MedicineRepository
}

type MedicineListInput struct {
	PageParams repository.PageParams
}

type StockChangeInput struct {
	MedicineID string
	Quantity   int
}

type CreateMedicineInput struct {
	Name          string
	Specification string
	Stock         int
	SafeStock     int
	ExpiryDate    string
}

type UpdateMedicineInventoryInput struct {
	MedicineID string
	Stock      *int
	SafeStock  *int
}

func NewMedicineService(repo repository.MedicineRepository) *MedicineService {
	return &MedicineService{repo: repo}
}

func (s *MedicineService) EnsureSeedData(ctx context.Context) error {
	return s.repo.EnsureSeedData(ctx)
}

func (s *MedicineService) List(ctx context.Context, input MedicineListInput) (repository.PageResult[repository.Medicine], error) {
	return s.repo.List(ctx, repository.MedicineListParams{
		PageParams: input.PageParams,
	})
}

func (s *MedicineService) ListAll(ctx context.Context) ([]repository.Medicine, error) {
	return s.repo.ListAll(ctx)
}

func (s *MedicineService) Create(ctx context.Context, input CreateMedicineInput) (repository.Medicine, error) {
	name := strings.TrimSpace(input.Name)
	specification := strings.TrimSpace(input.Specification)
	if name == "" || specification == "" || input.Stock < 0 || input.SafeStock < 0 {
		return repository.Medicine{}, ErrInvalidInput
	}

	expiryDate, err := time.Parse("2006-01-02", strings.TrimSpace(input.ExpiryDate))
	if err != nil {
		return repository.Medicine{}, ErrInvalidInput
	}

	return s.repo.Create(ctx, repository.CreateMedicineInput{
		Name:          name,
		Specification: specification,
		Stock:         input.Stock,
		SafeStock:     input.SafeStock,
		ExpiryDate:    expiryDate.UTC(),
	})
}

func (s *MedicineService) Inbound(ctx context.Context, input StockChangeInput) (repository.Medicine, error) {
	if strings.TrimSpace(input.MedicineID) == "" || input.Quantity <= 0 {
		return repository.Medicine{}, ErrInvalidInput
	}
	return s.repo.Inbound(ctx, repository.StockChangeInput{
		MedicineID: strings.TrimSpace(input.MedicineID),
		Quantity:   input.Quantity,
	})
}

func (s *MedicineService) Outbound(ctx context.Context, input StockChangeInput) (repository.Medicine, error) {
	if strings.TrimSpace(input.MedicineID) == "" || input.Quantity <= 0 {
		return repository.Medicine{}, ErrInvalidInput
	}
	return s.repo.Outbound(ctx, repository.StockChangeInput{
		MedicineID: strings.TrimSpace(input.MedicineID),
		Quantity:   input.Quantity,
	})
}

func (s *MedicineService) UpdateInventory(ctx context.Context, input UpdateMedicineInventoryInput) (repository.Medicine, error) {
	if strings.TrimSpace(input.MedicineID) == "" {
		return repository.Medicine{}, ErrInvalidInput
	}
	if input.Stock == nil && input.SafeStock == nil {
		return repository.Medicine{}, ErrInvalidInput
	}
	if input.Stock != nil && *input.Stock < 0 {
		return repository.Medicine{}, ErrInvalidInput
	}
	if input.SafeStock != nil && *input.SafeStock < 0 {
		return repository.Medicine{}, ErrInvalidInput
	}

	return s.repo.UpdateInventory(ctx, strings.TrimSpace(input.MedicineID), repository.UpdateMedicineInventoryInput{
		Stock:     input.Stock,
		SafeStock: input.SafeStock,
	})
}
