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

	now := time.Now().UTC()
	existingNames := make(map[string]struct{}, len(r.medicines))
	for _, medicine := range r.medicines {
		if medicine.Name != "" {
			existingNames[medicine.Name] = struct{}{}
		}
	}

	for _, input := range defaultMedicineSeedInputs(now) {
		if _, ok := existingNames[input.Name]; ok {
			continue
		}

		id := uuid.NewString()
		r.medicines[id] = mockMedicineFromInput(input, id, now)
		existingNames[input.Name] = struct{}{}
	}

	return nil
}

func mockMedicineFromInput(input CreateMedicineInput, id string, now time.Time) Medicine {
	threshold := now.AddDate(0, 0, 30)
	return Medicine{
		ID:                   id,
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
	medicine := mockMedicineFromInput(input, uuid.NewString(), now)
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
