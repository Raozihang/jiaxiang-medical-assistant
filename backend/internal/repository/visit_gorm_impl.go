package repository

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jiaxiang-medical-assistant/backend/internal/model"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type GormVisitRepository struct {
	db *gorm.DB
}

func NewGormVisitRepository(db *gorm.DB) *GormVisitRepository {
	return &GormVisitRepository{db: db}
}

func (r *GormVisitRepository) EnsureSeedData(ctx context.Context) error {
	var total int64
	if err := r.db.WithContext(ctx).Model(&model.Visit{}).Count(&total).Error; err != nil {
		return err
	}
	if total > 0 {
		return nil
	}

	student, err := r.ensureStudentByCode(ctx, "20260001")
	if err != nil {
		return err
	}

	symptoms, _ := json.Marshal([]string{"fever"})
	now := time.Now().UTC()
	return r.db.WithContext(ctx).Create(&model.Visit{
		StudentID:   student.ID,
		DoctorID:    uuid.New(),
		Symptoms:    datatypes.JSON(symptoms),
		Description: "felt unwell after PE class",
		Destination: "observation",
		CreatedAt:   now.Add(-10 * time.Minute),
		UpdatedAt:   now.Add(-10 * time.Minute),
	}).Error
}

func (r *GormVisitRepository) List(ctx context.Context, params VisitListParams) (PageResult[Visit], error) {
	query := r.db.WithContext(ctx).Model(&model.Visit{})

	if params.StudentID != "" {
		student, err := r.findStudentByCode(ctx, params.StudentID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return PageResult[Visit]{Items: []Visit{}, Page: params.Page, PageSize: params.PageSize, Total: 0}, nil
			}
			return PageResult[Visit]{}, err
		}
		query = query.Where("student_id = ?", student.ID)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return PageResult[Visit]{}, err
	}

	var rows []model.Visit
	if err := query.Order("created_at desc").Offset((params.Page - 1) * params.PageSize).Limit(params.PageSize).Find(&rows).Error; err != nil {
		return PageResult[Visit]{}, err
	}

	students, err := r.studentsByIDs(ctx, rows)
	if err != nil {
		return PageResult[Visit]{}, err
	}

	items := make([]Visit, 0, len(rows))
	for _, row := range rows {
		items = append(items, toVisitDTO(row, students[row.StudentID]))
	}

	return PageResult[Visit]{
		Items:    items,
		Page:     params.Page,
		PageSize: params.PageSize,
		Total:    total,
	}, nil
}

func (r *GormVisitRepository) Create(ctx context.Context, input CreateVisitInput) (Visit, error) {
	student, err := r.ensureStudentByCode(ctx, input.StudentID)
	if err != nil {
		return Visit{}, err
	}

	symptoms, err := json.Marshal(input.Symptoms)
	if err != nil {
		return Visit{}, err
	}

	now := time.Now().UTC()
	if input.CreatedAt != nil {
		now = input.CreatedAt.UTC()
	}
	row := model.Visit{
		StudentID:   student.ID,
		DoctorID:    uuid.New(),
		Symptoms:    datatypes.JSON(symptoms),
		Description: strings.TrimSpace(input.Description),
		Destination: "observation",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		return Visit{}, err
	}

	return toVisitDTO(row, student), nil
}

func (r *GormVisitRepository) GetByID(ctx context.Context, id string) (Visit, error) {
	rowID, err := uuid.Parse(id)
	if err != nil {
		return Visit{}, ErrNotFound
	}

	var row model.Visit
	if err := r.db.WithContext(ctx).First(&row, "id = ?", rowID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return Visit{}, ErrNotFound
		}
		return Visit{}, err
	}

	var student model.Student
	if err := r.db.WithContext(ctx).First(&student, "id = ?", row.StudentID).Error; err != nil {
		return Visit{}, err
	}

	return toVisitDTO(row, student), nil
}

func (r *GormVisitRepository) Update(ctx context.Context, id string, input UpdateVisitInput) (Visit, error) {
	rowID, err := uuid.Parse(id)
	if err != nil {
		return Visit{}, ErrNotFound
	}

	var row model.Visit
	if err := r.db.WithContext(ctx).First(&row, "id = ?", rowID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return Visit{}, ErrNotFound
		}
		return Visit{}, err
	}

	if input.Diagnosis != nil {
		row.Diagnosis = strings.TrimSpace(*input.Diagnosis)
	}
	if input.Prescription != nil {
		prescriptionRaw, marshalErr := json.Marshal(*input.Prescription)
		if marshalErr != nil {
			return Visit{}, marshalErr
		}
		row.Prescription = datatypes.JSON(prescriptionRaw)
	}
	if input.Destination != nil {
		row.Destination = strings.TrimSpace(*input.Destination)
	}
	if input.SetFollowUpAt {
		if input.FollowUpAt == nil {
			row.FollowUpAt = nil
		} else {
			followUpAt := input.FollowUpAt.UTC()
			row.FollowUpAt = &followUpAt
		}
	}
	if input.FollowUpNote != nil {
		note := strings.TrimSpace(*input.FollowUpNote)
		if note == "" {
			row.FollowUpNote = nil
		} else {
			row.FollowUpNote = &note
		}
	}
	row.UpdatedAt = time.Now().UTC()

	if err := r.db.WithContext(ctx).Save(&row).Error; err != nil {
		return Visit{}, err
	}

	var student model.Student
	if err := r.db.WithContext(ctx).First(&student, "id = ?", row.StudentID).Error; err != nil {
		return Visit{}, err
	}

	return toVisitDTO(row, student), nil
}

func (r *GormVisitRepository) CountToday(ctx context.Context, now time.Time) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.Visit{}).
		Where("created_at >= ?", dayStart(now)).
		Count(&count).Error
	return count, err
}

func (r *GormVisitRepository) CountObservationToday(ctx context.Context, now time.Time) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.Visit{}).
		Where("created_at >= ?", dayStart(now)).
		Where("destination = ?", "observation").
		Count(&count).Error
	return count, err
}

func (r *GormVisitRepository) CountDueFollowUps(ctx context.Context, now time.Time) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.Visit{}).
		Where("follow_up_at IS NOT NULL").
		Where("follow_up_at <= ?", now.UTC()).
		Count(&count).Error
	return count, err
}

func (r *GormVisitRepository) studentsByIDs(ctx context.Context, rows []model.Visit) (map[uuid.UUID]model.Student, error) {
	ids := make([]uuid.UUID, 0, len(rows))
	seen := map[uuid.UUID]struct{}{}
	for _, row := range rows {
		if _, ok := seen[row.StudentID]; ok {
			continue
		}
		seen[row.StudentID] = struct{}{}
		ids = append(ids, row.StudentID)
	}

	result := map[uuid.UUID]model.Student{}
	if len(ids) == 0 {
		return result, nil
	}

	var students []model.Student
	if err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&students).Error; err != nil {
		return nil, err
	}
	for _, student := range students {
		result[student.ID] = student
	}

	return result, nil
}

func (r *GormVisitRepository) findStudentByCode(ctx context.Context, studentID string) (model.Student, error) {
	var student model.Student
	err := r.db.WithContext(ctx).Where("student_id = ?", studentID).First(&student).Error
	return student, err
}

func (r *GormVisitRepository) ensureStudentByCode(ctx context.Context, studentID string) (model.Student, error) {
	student, err := r.findStudentByCode(ctx, studentID)
	if err == nil {
		return student, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return model.Student{}, err
	}

	student = model.Student{
		ID:        uuid.New(),
		StudentID: studentID,
		Name:      "Student-" + studentID,
		ClassID:   uuid.New(),
		Grade:     "Unknown",
	}
	if createErr := r.db.WithContext(ctx).Create(&student).Error; createErr != nil {
		return model.Student{}, createErr
	}

	return student, nil
}

func toVisitDTO(row model.Visit, student model.Student) Visit {
	var symptoms []string
	_ = json.Unmarshal(row.Symptoms, &symptoms)

	var prescription []string
	_ = json.Unmarshal(row.Prescription, &prescription)

	studentCode := row.StudentID.String()
	studentName := row.StudentID.String()
	className := "Unknown Class"
	if student.ID != uuid.Nil {
		studentCode = student.StudentID
		studentName = student.Name
		className = student.Grade
	}

	return Visit{
		ID:           row.ID.String(),
		StudentID:    studentCode,
		StudentName:  studentName,
		ClassName:    className,
		Symptoms:     symptoms,
		Description:  row.Description,
		Diagnosis:    row.Diagnosis,
		Prescription: prescription,
		Destination:  row.Destination,
		FollowUpAt:   row.FollowUpAt,
		FollowUpNote: row.FollowUpNote,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}
}
