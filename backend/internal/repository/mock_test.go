package repository

import (
	"context"
	"testing"
	"time"
)

func TestMockVisitRepositoryCRUD(t *testing.T) {
	repo := NewMockVisitRepository()
	if err := repo.EnsureSeedData(context.Background()); err != nil {
		t.Fatalf("seed failed: %v", err)
	}

	created, err := repo.Create(context.Background(), CreateVisitInput{
		StudentID:   "20261111",
		Symptoms:    []string{"cough"},
		Description: "test",
	})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	detail, err := repo.GetByID(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("detail failed: %v", err)
	}
	if detail.StudentID != "20261111" {
		t.Fatalf("unexpected student_id: %s", detail.StudentID)
	}
}

func TestMockMedicineOutboundInsufficientStock(t *testing.T) {
	repo := NewMockMedicineRepository()
	if err := repo.EnsureSeedData(context.Background()); err != nil {
		t.Fatalf("seed failed: %v", err)
	}

	list, err := repo.List(context.Background(), MedicineListParams{
		PageParams: PageParams{Page: 1, PageSize: 20},
	})
	if err != nil || len(list.Items) == 0 {
		t.Fatalf("list failed: %v", err)
	}

	_, err = repo.Outbound(context.Background(), StockChangeInput{
		MedicineID: list.Items[0].ID,
		Quantity:   99999,
	})
	if err != ErrInsufficientStock {
		t.Fatalf("expected ErrInsufficientStock, got %v", err)
	}
}

func TestMockMedicineExpiryBoundaryConsistency(t *testing.T) {
	repo := NewMockMedicineRepository()
	now := time.Now().UTC()

	repo.medicines["boundary"] = Medicine{
		ID:         "boundary",
		Name:       "Boundary Medicine",
		Stock:      10,
		SafeStock:  5,
		ExpiryDate: now.AddDate(0, 0, 30),
	}

	list, err := repo.List(context.Background(), MedicineListParams{
		PageParams: PageParams{Page: 1, PageSize: 10},
	})
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}

	if len(list.Items) != 1 || !list.Items[0].IsExpiringSoon {
		t.Fatalf("expected boundary medicine to be expiring soon")
	}

	warnings, err := repo.CountWarnings(context.Background(), now)
	if err != nil {
		t.Fatalf("count warnings failed: %v", err)
	}
	if warnings != 1 {
		t.Fatalf("expected 1 warning, got %d", warnings)
	}
}
