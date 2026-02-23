package repository

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jiaxiang-medical-assistant/backend/internal/model"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GormMedicineRepository struct {
	db *gorm.DB
}

func NewGormMedicineRepository(db *gorm.DB) *GormMedicineRepository {
	return &GormMedicineRepository{db: db}
}

func (r *GormMedicineRepository) EnsureSeedData(ctx context.Context) error {
	var total int64
	if err := r.db.WithContext(ctx).Model(&model.Medicine{}).Count(&total).Error; err != nil {
		return err
	}
	if total > 0 {
		return nil
	}

	now := time.Now().UTC()
	rows := []model.Medicine{
		{
			ID:            uuid.New(),
			Name:          "Ibuprofen Tablets",
			Specification: "0.2g*24",
			Stock:         120,
			SafeStock:     50,
			ExpiryDate:    now.AddDate(1, 0, 0),
			Warnings:      datatypes.JSON([]byte(`["allergy risk"]`)),
		},
		{
			ID:            uuid.New(),
			Name:          "Medical Gauze",
			Specification: "10cm*10cm",
			Stock:         30,
			SafeStock:     40,
			ExpiryDate:    now.AddDate(0, 1, 0),
			Warnings:      datatypes.JSON([]byte(`[]`)),
		},
	}

	return r.db.WithContext(ctx).Create(&rows).Error
}

func (r *GormMedicineRepository) List(ctx context.Context, params MedicineListParams) (PageResult[Medicine], error) {
	query := r.db.WithContext(ctx).Model(&model.Medicine{})

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return PageResult[Medicine]{}, err
	}

	var rows []model.Medicine
	if err := query.Order("name asc").Offset((params.Page - 1) * params.PageSize).Limit(params.PageSize).Find(&rows).Error; err != nil {
		return PageResult[Medicine]{}, err
	}

	items := make([]Medicine, 0, len(rows))
	for _, row := range rows {
		items = append(items, toMedicineDTO(row))
	}

	return PageResult[Medicine]{
		Items:    items,
		Page:     params.Page,
		PageSize: params.PageSize,
		Total:    total,
	}, nil
}

func (r *GormMedicineRepository) Inbound(ctx context.Context, input StockChangeInput) (Medicine, error) {
	return r.changeStock(ctx, input, true)
}

func (r *GormMedicineRepository) Outbound(ctx context.Context, input StockChangeInput) (Medicine, error) {
	return r.changeStock(ctx, input, false)
}

func (r *GormMedicineRepository) changeStock(ctx context.Context, input StockChangeInput, inbound bool) (Medicine, error) {
	medicineID, err := uuid.Parse(input.MedicineID)
	if err != nil {
		return Medicine{}, ErrNotFound
	}

	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return Medicine{}, tx.Error
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	var row model.Medicine
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&row, "id = ?", medicineID).Error; err != nil {
		_ = tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return Medicine{}, ErrNotFound
		}
		return Medicine{}, err
	}

	if !inbound && row.Stock < input.Quantity {
		_ = tx.Rollback()
		return Medicine{}, ErrInsufficientStock
	}

	if inbound {
		row.Stock += input.Quantity
	} else {
		row.Stock -= input.Quantity
	}
	row.UpdatedAt = time.Now().UTC()

	if err := tx.Save(&row).Error; err != nil {
		_ = tx.Rollback()
		return Medicine{}, err
	}

	if err := tx.Commit().Error; err != nil {
		return Medicine{}, err
	}

	return toMedicineDTO(row), nil
}

func (r *GormMedicineRepository) CountWarnings(ctx context.Context, now time.Time) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.Medicine{}).
		Where("stock < safe_stock OR expiry_date <= ?", now.UTC().AddDate(0, 0, 30)).
		Count(&count).Error
	return count, err
}

func toMedicineDTO(row model.Medicine) Medicine {
	var warnings []string
	_ = json.Unmarshal(row.Warnings, &warnings)

	threshold := time.Now().UTC().AddDate(0, 0, 30)
	return Medicine{
		ID:             row.ID.String(),
		Name:           row.Name,
		Specification:  row.Specification,
		Stock:          row.Stock,
		SafeStock:      row.SafeStock,
		ExpiryDate:     row.ExpiryDate,
		Warnings:       warnings,
		IsLowStock:     row.Stock < row.SafeStock,
		IsExpiringSoon: !row.ExpiryDate.After(threshold),
		CreatedAt:      row.CreatedAt,
		UpdatedAt:      row.UpdatedAt,
	}
}
