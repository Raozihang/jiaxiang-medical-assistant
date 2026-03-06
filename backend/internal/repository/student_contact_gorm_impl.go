package repository

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jiaxiang-medical-assistant/backend/internal/model"
	"gorm.io/gorm"
)

type GormStudentContactRepository struct {
	db *gorm.DB
}

func NewGormStudentContactRepository(db *gorm.DB) *GormStudentContactRepository {
	return &GormStudentContactRepository{db: db}
}

func (r *GormStudentContactRepository) List(ctx context.Context, params StudentContactListParams) (PageResult[StudentContact], error) {
	query := r.db.WithContext(ctx).Model(&model.Student{})
	keyword := strings.TrimSpace(params.Keyword)
	if keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("student_id ILIKE ? OR name ILIKE ? OR guardian_name ILIKE ? OR guardian_phone ILIKE ?", like, like, like, like)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return PageResult[StudentContact]{}, err
	}

	var rows []model.Student
	if err := query.Order("updated_at desc").Offset((params.Page - 1) * params.PageSize).Limit(params.PageSize).Find(&rows).Error; err != nil {
		return PageResult[StudentContact]{}, err
	}

	items := make([]StudentContact, 0, len(rows))
	for _, row := range rows {
		items = append(items, toStudentContactDTO(row))
	}

	return PageResult[StudentContact]{Items: items, Page: params.Page, PageSize: params.PageSize, Total: total}, nil
}

func (r *GormStudentContactRepository) GetByStudentID(ctx context.Context, studentID string) (StudentContact, error) {
	var row model.Student
	if err := r.db.WithContext(ctx).Where("student_id = ?", strings.TrimSpace(studentID)).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return StudentContact{}, ErrNotFound
		}
		return StudentContact{}, err
	}

	return toStudentContactDTO(row), nil
}

func (r *GormStudentContactRepository) UpdateByStudentID(ctx context.Context, studentID string, input UpdateStudentContactInput) (StudentContact, error) {
	code := strings.TrimSpace(studentID)
	var row model.Student
	err := r.db.WithContext(ctx).Where("student_id = ?", code).First(&row).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return StudentContact{}, err
		}
		name := code
		if input.StudentName != nil && strings.TrimSpace(*input.StudentName) != "" {
			name = strings.TrimSpace(*input.StudentName)
		}
		row = model.Student{
			ID:        uuid.New(),
			StudentID: code,
			Name:      name,
			ClassID:   uuid.New(),
			Grade:     "未知",
		}
	}

	if input.StudentName != nil && strings.TrimSpace(*input.StudentName) != "" {
		row.Name = strings.TrimSpace(*input.StudentName)
	}
	if input.GuardianName != nil {
		row.GuardianName = strings.TrimSpace(*input.GuardianName)
	}
	if input.GuardianPhone != nil {
		row.GuardianPhone = strings.TrimSpace(*input.GuardianPhone)
	}
	if input.GuardianRelation != nil {
		row.GuardianRelation = strings.TrimSpace(*input.GuardianRelation)
	}

	if err := r.db.WithContext(ctx).Save(&row).Error; err != nil {
		return StudentContact{}, err
	}

	return toStudentContactDTO(row), nil
}

func toStudentContactDTO(row model.Student) StudentContact {
	return StudentContact{
		StudentID:        row.StudentID,
		StudentName:      row.Name,
		GuardianName:     row.GuardianName,
		GuardianPhone:    row.GuardianPhone,
		GuardianRelation: row.GuardianRelation,
	}
}
