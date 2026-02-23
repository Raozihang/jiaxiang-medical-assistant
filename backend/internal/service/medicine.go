package service

import (
	"context"
	"strings"

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

func (s *MedicineService) Inbound(ctx context.Context, input StockChangeInput) (repository.Medicine, error) {
	return s.repo.Inbound(ctx, repository.StockChangeInput{
		MedicineID: strings.TrimSpace(input.MedicineID),
		Quantity:   input.Quantity,
	})
}

func (s *MedicineService) Outbound(ctx context.Context, input StockChangeInput) (repository.Medicine, error) {
	return s.repo.Outbound(ctx, repository.StockChangeInput{
		MedicineID: strings.TrimSpace(input.MedicineID),
		Quantity:   input.Quantity,
	})
}
