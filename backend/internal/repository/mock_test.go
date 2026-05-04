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

func TestMockMedicineCreateAndUpdateInventory(t *testing.T) {
	repo := NewMockMedicineRepository()
	expiryDate := time.Now().UTC().AddDate(0, 6, 0)

	created, err := repo.Create(context.Background(), CreateMedicineInput{
		Name:          "Test Medicine",
		Specification: "10ml",
		Stock:         12,
		SafeStock:     5,
		ExpiryDate:    expiryDate,
	})
	if err != nil {
		t.Fatalf("create medicine failed: %v", err)
	}
	if created.ID == "" || created.Stock != 12 || created.SafeStock != 5 {
		t.Fatalf("unexpected created medicine: %+v", created)
	}

	stock := 3
	safeStock := 8
	updated, err := repo.UpdateInventory(context.Background(), created.ID, UpdateMedicineInventoryInput{
		Stock:     &stock,
		SafeStock: &safeStock,
	})
	if err != nil {
		t.Fatalf("update inventory failed: %v", err)
	}
	if updated.Stock != stock || updated.SafeStock != safeStock || !updated.IsLowStock {
		t.Fatalf("unexpected updated medicine: %+v", updated)
	}
}

func TestMockMedicineEnsureSeedDataBackfillsMissingCatalog(t *testing.T) {
	repo := NewMockMedicineRepository()
	repo.medicines["custom"] = Medicine{
		ID:         "custom",
		Name:       "自定义药品",
		Stock:      5,
		SafeStock:  2,
		ExpiryDate: time.Now().UTC().AddDate(0, 6, 0),
	}

	if err := repo.EnsureSeedData(context.Background()); err != nil {
		t.Fatalf("seed failed: %v", err)
	}

	all, err := repo.ListAll(context.Background())
	if err != nil {
		t.Fatalf("list all failed: %v", err)
	}

	if len(all) < 51 {
		t.Fatalf("expected at least 51 medicines after backfill, got %d", len(all))
	}

	found := false
	for _, item := range all {
		if item.Name == "布洛芬片" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected seeded OTC catalog to include 布洛芬片")
	}
}
