package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jiaxiang-medical-assistant/backend/internal/repository"
)

type ImportService struct {
	visitRepo repository.VisitRepository
	taskRepo  repository.ImportTaskRepository
}

type VisitImportItem struct {
	StudentID    string     `json:"student_id"`
	Symptoms     []string   `json:"symptoms"`
	Description  string     `json:"description"`
	Diagnosis    string     `json:"diagnosis"`
	Prescription []string   `json:"prescription"`
	Destination  string     `json:"destination"`
	CreatedAt    *time.Time `json:"created_at"`
}

func NewImportService(visitRepo repository.VisitRepository, taskRepo repository.ImportTaskRepository) *ImportService {
	return &ImportService{visitRepo: visitRepo, taskRepo: taskRepo}
}

func (s *ImportService) SubmitVisits(ctx context.Context, items []VisitImportItem) (repository.ImportTask, error) {
	now := time.Now().UTC()
	task := repository.ImportTask{
		ID:        uuid.NewString(),
		Status:    "processing",
		Total:     len(items),
		Success:   0,
		Failed:    0,
		Errors:    []repository.ImportTaskError{},
		CreatedAt: now,
		UpdatedAt: now,
	}

	task, err := s.taskRepo.Create(ctx, task)
	if err != nil {
		return repository.ImportTask{}, err
	}

	for idx, item := range items {
		if strings.TrimSpace(item.StudentID) == "" {
			task.Failed++
			task.Errors = append(task.Errors, repository.ImportTaskError{
				Index:   idx,
				Message: "学号不能为空",
			})
			continue
		}

		created, createErr := s.visitRepo.Create(ctx, repository.CreateVisitInput{
			StudentID:   strings.TrimSpace(item.StudentID),
			Symptoms:    item.Symptoms,
			Description: strings.TrimSpace(item.Description),
			CreatedAt:   item.CreatedAt,
		})
		if createErr != nil {
			task.Failed++
			task.Errors = append(task.Errors, repository.ImportTaskError{
				Index:   idx,
				Message: fmt.Sprintf("创建就诊记录失败: %v", createErr),
			})
			continue
		}

		if hasVisitUpdatePayload(item) {
			input := repository.UpdateVisitInput{}
			if diagnosis := strings.TrimSpace(item.Diagnosis); diagnosis != "" {
				input.Diagnosis = &diagnosis
			}
			if destination := strings.TrimSpace(item.Destination); destination != "" {
				input.Destination = &destination
			}
			if len(item.Prescription) > 0 {
				prescription := item.Prescription
				input.Prescription = &prescription
			}

			if _, updateErr := s.visitRepo.Update(ctx, created.ID, input); updateErr != nil {
				task.Failed++
				task.Errors = append(task.Errors, repository.ImportTaskError{
					Index:   idx,
					Message: fmt.Sprintf("更新就诊记录失败: %v", updateErr),
				})
				continue
			}
		}

		task.Success++
	}

	task.UpdatedAt = time.Now().UTC()
	switch {
	case task.Failed == 0:
		task.Status = "completed"
	case task.Success == 0:
		task.Status = "failed"
	default:
		task.Status = "completed_with_errors"
	}

	return s.taskRepo.Update(ctx, task)
}

func (s *ImportService) GetTask(ctx context.Context, id string) (repository.ImportTask, error) {
	return s.taskRepo.GetByID(ctx, strings.TrimSpace(id))
}

func (s *ImportService) ListTasks(ctx context.Context, pageParams repository.PageParams) (repository.PageResult[repository.ImportTask], error) {
	items, err := s.taskRepo.List(ctx)
	if err != nil {
		return repository.PageResult[repository.ImportTask]{}, err
	}

	start, end := pageWindow(pageParams.Page, pageParams.PageSize, len(items))
	return repository.PageResult[repository.ImportTask]{
		Items:    items[start:end],
		Page:     pageParams.Page,
		PageSize: pageParams.PageSize,
		Total:    int64(len(items)),
	}, nil
}

func pageWindow(page int, pageSize int, total int) (int, int) {
	if total == 0 {
		return 0, 0
	}

	start := (page - 1) * pageSize
	if start >= total {
		return total, total
	}

	end := start + pageSize
	if end > total {
		end = total
	}

	return start, end
}

func hasVisitUpdatePayload(item VisitImportItem) bool {
	return strings.TrimSpace(item.Diagnosis) != "" || strings.TrimSpace(item.Destination) != "" || len(item.Prescription) > 0
}
