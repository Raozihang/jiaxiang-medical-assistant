package repository

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
)

type MockMedicineRepository struct {
	mu        sync.RWMutex
	medicines map[string]Medicine
}

func NewMockMedicineRepository() *MockMedicineRepository {
	return &MockMedicineRepository{medicines: map[string]Medicine{}}
}

func (r *MockMedicineRepository) EnsureSeedData(_ context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.medicines) > 0 {
		return nil
	}

	now := time.Now().UTC()
	id1 := uuid.NewString()
	r.medicines[id1] = Medicine{
		ID:                   id1,
		Name:                 "布洛芬片",
		Specification:        "0.2g*24片",
		Stock:                120,
		SafeStock:            50,
		ExpiryDate:           now.AddDate(1, 0, 0),
		Warnings:             []string{"对布洛芬过敏者禁用"},
		RecommendedDosage:    "0.2g",
		RecommendedFrequency: "必要时每6到8小时一次",
		RecommendedDuration:  "最多连续使用3天",
		UsageInstructions:    "建议饭后服用；如出现胃部不适应立即停用",
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	id2 := uuid.NewString()
	r.medicines[id2] = Medicine{
		ID:                id2,
		Name:              "医用纱布",
		Specification:     "10cm*10cm",
		Stock:             30,
		SafeStock:         40,
		ExpiryDate:        now.AddDate(0, 1, 0),
		Warnings:          []string{},
		UsageInstructions: "仅限外用",
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	return nil
}

func (r *MockMedicineRepository) List(_ context.Context, params MedicineListParams) (PageResult[Medicine], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	items := make([]Medicine, 0, len(r.medicines))
	threshold := time.Now().UTC().AddDate(0, 0, 30)
	for _, medicine := range r.medicines {
		medicine.IsLowStock = medicine.Stock < medicine.SafeStock
		medicine.IsExpiringSoon = !medicine.ExpiryDate.After(threshold)
		items = append(items, medicine)
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].Name < items[j].Name
	})

	start, end := pageWindow(params.Page, params.PageSize, len(items))
	return PageResult[Medicine]{
		Items:    items[start:end],
		Page:     params.Page,
		PageSize: params.PageSize,
		Total:    int64(len(items)),
	}, nil
}

func (r *MockMedicineRepository) ListAll(_ context.Context) ([]Medicine, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	items := make([]Medicine, 0, len(r.medicines))
	threshold := time.Now().UTC().AddDate(0, 0, 30)
	for _, medicine := range r.medicines {
		medicine.IsLowStock = medicine.Stock < medicine.SafeStock
		medicine.IsExpiringSoon = !medicine.ExpiryDate.After(threshold)
		items = append(items, medicine)
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].Name < items[j].Name
	})

	return items, nil
}

func (r *MockMedicineRepository) Create(_ context.Context, input CreateMedicineInput) (Medicine, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now().UTC()
	threshold := now.AddDate(0, 0, 30)
	medicine := Medicine{
		ID:                   uuid.NewString(),
		Name:                 input.Name,
		Specification:        input.Specification,
		Stock:                input.Stock,
		SafeStock:            input.SafeStock,
		ExpiryDate:           input.ExpiryDate.UTC(),
		Warnings:             append([]string(nil), input.Warnings...),
		RecommendedDosage:    input.RecommendedDosage,
		RecommendedFrequency: input.RecommendedFrequency,
		RecommendedDuration:  input.RecommendedDuration,
		UsageInstructions:    input.UsageInstructions,
		IsLowStock:           input.Stock < input.SafeStock,
		IsExpiringSoon:       !input.ExpiryDate.After(threshold),
		CreatedAt:            now,
		UpdatedAt:            now,
	}
	r.medicines[medicine.ID] = medicine

	return medicine, nil
}

func (r *MockMedicineRepository) Inbound(_ context.Context, input StockChangeInput) (Medicine, error) {
	return r.changeStock(input, true)
}

func (r *MockMedicineRepository) Outbound(_ context.Context, input StockChangeInput) (Medicine, error) {
	return r.changeStock(input, false)
}

func (r *MockMedicineRepository) changeStock(input StockChangeInput, inbound bool) (Medicine, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	medicine, ok := r.medicines[input.MedicineID]
	if !ok {
		return Medicine{}, ErrNotFound
	}

	if !inbound && medicine.Stock < input.Quantity {
		return Medicine{}, ErrInsufficientStock
	}

	if inbound {
		medicine.Stock += input.Quantity
	} else {
		medicine.Stock -= input.Quantity
	}

	now := time.Now().UTC()
	threshold := now.AddDate(0, 0, 30)
	medicine.UpdatedAt = now
	medicine.IsLowStock = medicine.Stock < medicine.SafeStock
	medicine.IsExpiringSoon = !medicine.ExpiryDate.After(threshold)
	r.medicines[input.MedicineID] = medicine

	return medicine, nil
}

func (r *MockMedicineRepository) UpdateInventory(_ context.Context, id string, input UpdateMedicineInventoryInput) (Medicine, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	medicine, ok := r.medicines[id]
	if !ok {
		return Medicine{}, ErrNotFound
	}

	if input.Stock != nil {
		medicine.Stock = *input.Stock
	}
	if input.SafeStock != nil {
		medicine.SafeStock = *input.SafeStock
	}

	now := time.Now().UTC()
	threshold := now.AddDate(0, 0, 30)
	medicine.UpdatedAt = now
	medicine.IsLowStock = medicine.Stock < medicine.SafeStock
	medicine.IsExpiringSoon = !medicine.ExpiryDate.After(threshold)
	r.medicines[id] = medicine

	return medicine, nil
}

func (r *MockMedicineRepository) CountWarnings(_ context.Context, now time.Time) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	threshold := now.UTC().AddDate(0, 0, 30)
	var count int64
	for _, medicine := range r.medicines {
		if medicine.Stock < medicine.SafeStock || !medicine.ExpiryDate.After(threshold) {
			count++
		}
	}

	return count, nil
}
