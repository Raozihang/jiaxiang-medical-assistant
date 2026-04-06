package service

import (
	"context"
	"errors"
	"testing"

	"github.com/jiaxiang-medical-assistant/backend/internal/repository"
)

func TestMedicineServiceCreateRejectsInvalidInput(t *testing.T) {
	svc := NewMedicineService(repository.NewMockMedicineRepository())

	_, err := svc.Create(context.Background(), CreateMedicineInput{
		Name:          "  ",
		Specification: "tablet",
		Stock:         1,
		SafeStock:     1,
		ExpiryDate:    "2026-12-31",
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}

	_, err = svc.Create(context.Background(), CreateMedicineInput{
		Name:          "Ibuprofen",
		Specification: "tablet",
		Stock:         -1,
		SafeStock:     1,
		ExpiryDate:    "2026-12-31",
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput for negative stock, got %v", err)
	}
}

func TestMedicineServiceCreateAndUpdateInventory(t *testing.T) {
	ctx := context.Background()
	svc := NewMedicineService(repository.NewMockMedicineRepository())

	created, err := svc.Create(ctx, CreateMedicineInput{
		Name:          "Ibuprofen",
		Specification: "0.2g*24",
		Stock:         10,
		SafeStock:     6,
		ExpiryDate:    "2026-12-31",
	})
	if err != nil {
		t.Fatalf("create medicine failed: %v", err)
	}

	stock := 4
	safeStock := 9
	updated, err := svc.UpdateInventory(ctx, UpdateMedicineInventoryInput{
		MedicineID: created.ID,
		Stock:      &stock,
		SafeStock:  &safeStock,
	})
	if err != nil {
		t.Fatalf("update inventory failed: %v", err)
	}
	if updated.Stock != stock || updated.SafeStock != safeStock || !updated.IsLowStock {
		t.Fatalf("unexpected updated medicine: %+v", updated)
	}
}
