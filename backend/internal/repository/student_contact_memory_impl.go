package repository

import (
	"context"
	"sort"
	"strings"
	"sync"
)

type MemoryStudentContactRepository struct {
	mu       sync.RWMutex
	contacts map[string]StudentContact
}

func NewMemoryStudentContactRepository() *MemoryStudentContactRepository {
	return &MemoryStudentContactRepository{
		contacts: map[string]StudentContact{
			"20260001": {
				StudentID:        "20260001",
				StudentName:      "张三",
				GuardianName:     "张家长",
				GuardianPhone:    "13800000001",
				GuardianRelation: "父亲",
			},
		},
	}
}

func (r *MemoryStudentContactRepository) List(_ context.Context, params StudentContactListParams) (PageResult[StudentContact], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	keyword := strings.TrimSpace(params.Keyword)
	items := make([]StudentContact, 0, len(r.contacts))
	for _, item := range r.contacts {
		if keyword != "" {
			target := strings.Join([]string{item.StudentID, item.StudentName, item.GuardianName, item.GuardianPhone}, " ")
			if !strings.Contains(target, keyword) {
				continue
			}
		}
		items = append(items, item)
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].StudentID < items[j].StudentID
	})

	start, end := pageWindow(params.Page, params.PageSize, len(items))
	return PageResult[StudentContact]{
		Items:    items[start:end],
		Page:     params.Page,
		PageSize: params.PageSize,
		Total:    int64(len(items)),
	}, nil
}

func (r *MemoryStudentContactRepository) GetByStudentID(_ context.Context, studentID string) (StudentContact, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	item, ok := r.contacts[strings.TrimSpace(studentID)]
	if !ok {
		return StudentContact{}, ErrNotFound
	}
	return item, nil
}

func (r *MemoryStudentContactRepository) UpdateByStudentID(_ context.Context, studentID string, input UpdateStudentContactInput) (StudentContact, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	code := strings.TrimSpace(studentID)
	item := r.contacts[code]
	if item.StudentID == "" {
		item.StudentID = code
		item.StudentName = code
	}
	if input.StudentName != nil && strings.TrimSpace(*input.StudentName) != "" {
		item.StudentName = strings.TrimSpace(*input.StudentName)
	}
	if input.GuardianName != nil {
		item.GuardianName = strings.TrimSpace(*input.GuardianName)
	}
	if input.GuardianPhone != nil {
		item.GuardianPhone = strings.TrimSpace(*input.GuardianPhone)
	}
	if input.GuardianRelation != nil {
		item.GuardianRelation = strings.TrimSpace(*input.GuardianRelation)
	}
	r.contacts[code] = item

	return item, nil
}
