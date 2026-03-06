package repository

import "context"

type StudentContact struct {
	StudentID        string `json:"student_id"`
	StudentName      string `json:"student_name"`
	GuardianName     string `json:"guardian_name"`
	GuardianPhone    string `json:"guardian_phone"`
	GuardianRelation string `json:"guardian_relation"`
}

type StudentContactListParams struct {
	PageParams
	Keyword string
}

type UpdateStudentContactInput struct {
	StudentName      *string
	GuardianName     *string
	GuardianPhone    *string
	GuardianRelation *string
}

type StudentContactRepository interface {
	List(ctx context.Context, params StudentContactListParams) (PageResult[StudentContact], error)
	GetByStudentID(ctx context.Context, studentID string) (StudentContact, error)
	UpdateByStudentID(ctx context.Context, studentID string, input UpdateStudentContactInput) (StudentContact, error)
}
