package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/jiaxiang-medical-assistant/backend/internal/repository"
)

const (
	OutboundCallScenarioExternalMedicalFollowup = "external_medical_followup"
	visitDestinationLeaveSchool                 = "leave_school"
	visitDestinationReferred                    = "referred"
)

type OutboundCallService struct {
	repo         repository.OutboundCallRepository
	visitRepo    repository.VisitRepository
	studentRepo  repository.StudentContactRepository
	provider     OutboundCallProvider
	templateCode string
}

type OutboundCallListInput struct {
	PageParams repository.PageParams
	Status     string
	StudentID  string
	Keyword    string
}

type AliyunCallbackInput struct {
	RequestID string `json:"request_id"`
	Status    string `json:"status"`
	Error     string `json:"error"`
	CallID    string `json:"call_id"`
	Payload   string `json:"payload"`
}

func NewOutboundCallService(repo repository.OutboundCallRepository, visitRepo repository.VisitRepository, studentRepo repository.StudentContactRepository, provider OutboundCallProvider, templateCode string) *OutboundCallService {
	if provider == nil {
		provider = NewMockOutboundCallProvider()
	}

	return &OutboundCallService{
		repo:         repo,
		visitRepo:    visitRepo,
		studentRepo:  studentRepo,
		provider:     provider,
		templateCode: strings.TrimSpace(templateCode),
	}
}

func (s *OutboundCallService) List(ctx context.Context, input OutboundCallListInput) (repository.PageResult[repository.OutboundCall], error) {
	return s.repo.List(ctx, repository.OutboundCallListParams{
		PageParams: input.PageParams,
		Status:     strings.TrimSpace(input.Status),
		StudentID:  strings.TrimSpace(input.StudentID),
		Keyword:    strings.TrimSpace(input.Keyword),
	})
}

func (s *OutboundCallService) TrackVisitUpdate(ctx context.Context, visit repository.Visit) {
	if !shouldTrackExternalMedical(visit.Destination) {
		return
	}

	_, err := s.repo.FindLatestByVisitAndScenario(ctx, visit.ID, OutboundCallScenarioExternalMedicalFollowup)
	switch {
	case err == nil:
		return
	case !errors.Is(err, repository.ErrNotFound):
		return
	}

	contact, err := s.studentRepo.GetByStudentID(ctx, visit.StudentID)
	if err != nil {
		_, _ = s.createSkippedCall(ctx, visit, fallbackContact(visit, repository.StudentContact{}), nil, "system", "guardian contact not found")
		return
	}

	if strings.TrimSpace(contact.GuardianPhone) == "" {
		_, _ = s.createSkippedCall(ctx, visit, contact, nil, "system", "guardian phone is empty")
		return
	}

	_, _ = s.placeAndPersistCall(ctx, visit, contact, nil, "system")
}

func (s *OutboundCallService) Retry(ctx context.Context, id string) (repository.OutboundCall, error) {
	current, err := s.repo.GetByID(ctx, strings.TrimSpace(id))
	if err != nil {
		return repository.OutboundCall{}, err
	}

	visit, err := s.visitRepo.GetByID(ctx, current.VisitID)
	if err != nil {
		return repository.OutboundCall{}, err
	}

	contact, err := s.studentRepo.GetByStudentID(ctx, current.StudentID)
	if err != nil {
		if !errors.Is(err, repository.ErrNotFound) {
			return repository.OutboundCall{}, err
		}
		contact = repository.StudentContact{
			StudentID:        current.StudentID,
			StudentName:      current.StudentName,
			GuardianName:     current.GuardianName,
			GuardianPhone:    current.GuardianPhone,
			GuardianRelation: current.GuardianRelation,
		}
	}
	contact = fallbackContact(visit, contact)

	if strings.TrimSpace(contact.GuardianPhone) == "" {
		return s.createSkippedCall(ctx, visit, contact, &current.ID, "manual", "guardian phone is empty")
	}

	return s.placeAndPersistCall(ctx, visit, contact, &current.ID, "manual")
}

func (s *OutboundCallService) HandleAliyunCallback(ctx context.Context, input AliyunCallbackInput) (repository.OutboundCall, error) {
	requestID := strings.TrimSpace(input.RequestID)
	if requestID == "" {
		return repository.OutboundCall{}, ErrInvalidInput
	}

	call, err := s.repo.FindByRequestID(ctx, requestID)
	if err != nil {
		return repository.OutboundCall{}, err
	}

	status := normalizeCallStatus(input.Status)
	now := time.Now().UTC()
	errText := strings.TrimSpace(input.Error)
	callID := strings.TrimSpace(input.CallID)
	payload := strings.TrimSpace(input.Payload)

	return s.repo.UpdateStatus(ctx, call.ID, repository.UpdateOutboundCallStatusInput{
		Status:      status,
		CallID:      optionalStringPtr(callID),
		Error:       optionalStringPtr(errText),
		ResponseRaw: optionalStringPtr(payload),
		CompletedAt: &now,
	})
}

func (s *OutboundCallService) createSkippedCall(ctx context.Context, visit repository.Visit, contact repository.StudentContact, retryOfID *string, triggerSource string, reason string) (repository.OutboundCall, error) {
	contact = fallbackContact(visit, contact)
	now := time.Now().UTC()

	return s.repo.Create(ctx, repository.CreateOutboundCallInput{
		VisitID:          visit.ID,
		StudentID:        visit.StudentID,
		StudentName:      visit.StudentName,
		GuardianName:     contact.GuardianName,
		GuardianPhone:    contact.GuardianPhone,
		GuardianRelation: contact.GuardianRelation,
		Scenario:         OutboundCallScenarioExternalMedicalFollowup,
		Provider:         s.provider.ProviderName(),
		TriggerSource:    triggerSource,
		Status:           "failed",
		Message:          buildExternalMedicalCallMessage(visit, contact),
		TemplateCode:     s.templateCode,
		TemplateParams:   mustJSON(map[string]string{"reason": reason, "destination": visit.Destination}),
		Error:            reason,
		RetryOfID:        retryOfID,
		RequestedAt:      now,
		CompletedAt:      timePtr(now),
	})
}

func (s *OutboundCallService) placeAndPersistCall(ctx context.Context, visit repository.Visit, contact repository.StudentContact, retryOfID *string, triggerSource string) (repository.OutboundCall, error) {
	contact = fallbackContact(visit, contact)
	message := buildExternalMedicalCallMessage(visit, contact)
	vars := map[string]string{
		"student_name":  visit.StudentName,
		"guardian_name": contact.GuardianName,
		"destination":   visit.Destination,
		"message":       message,
	}
	result, err := s.provider.PlaceCall(ctx, PlaceOutboundCallInput{
		Phone:        contact.GuardianPhone,
		TemplateCode: s.templateCode,
		Message:      message,
		TemplateVars: vars,
	})

	requestedAt := time.Now().UTC()
	status := "failed"
	responseRaw := ""
	requestID := ""
	callID := ""
	errText := ""
	completedAt := timePtr(requestedAt)
	if err == nil {
		status = normalizeCallStatus(result.Status)
		responseRaw = result.ResponseRaw
		requestID = result.RequestID
		callID = result.CallID
		completedAt = nil
	} else {
		errText = err.Error()
	}

	created, createErr := s.repo.Create(ctx, repository.CreateOutboundCallInput{
		VisitID:          visit.ID,
		StudentID:        visit.StudentID,
		StudentName:      visit.StudentName,
		GuardianName:     contact.GuardianName,
		GuardianPhone:    contact.GuardianPhone,
		GuardianRelation: contact.GuardianRelation,
		Scenario:         OutboundCallScenarioExternalMedicalFollowup,
		Provider:         s.provider.ProviderName(),
		TriggerSource:    triggerSource,
		Status:           status,
		Message:          message,
		TemplateCode:     s.templateCode,
		TemplateParams:   mustJSON(vars),
		RequestID:        requestID,
		CallID:           callID,
		Error:            errText,
		ResponseRaw:      responseRaw,
		RetryOfID:        retryOfID,
		RequestedAt:      requestedAt,
		CompletedAt:      completedAt,
	})
	if createErr != nil {
		return repository.OutboundCall{}, createErr
	}

	return created, nil
}

func fallbackContact(visit repository.Visit, contact repository.StudentContact) repository.StudentContact {
	if strings.TrimSpace(contact.StudentID) == "" {
		contact.StudentID = visit.StudentID
	}
	if strings.TrimSpace(contact.StudentName) == "" {
		contact.StudentName = visit.StudentName
	}
	if strings.TrimSpace(contact.GuardianName) == "" {
		contact.GuardianName = "Guardian"
	}
	return contact
}

func buildExternalMedicalCallMessage(visit repository.Visit, contact repository.StudentContact) string {
	studentName := strings.TrimSpace(visit.StudentName)
	if studentName == "" {
		studentName = visit.StudentID
	}
	guardianName := strings.TrimSpace(contact.GuardianName)
	if guardianName == "" {
		guardianName = "Guardian"
	}

	destinationText := visit.Destination
	switch strings.TrimSpace(visit.Destination) {
	case visitDestinationLeaveSchool:
		destinationText = "leave school for medical care"
	case visitDestinationReferred:
		destinationText = "be referred to an external hospital"
	}

	return guardianName + ", hello. " + studentName + " has been advised by the school clinic to " + destinationText + ". Please follow up promptly."
}

func shouldTrackExternalMedical(destination string) bool {
	switch strings.TrimSpace(strings.ToLower(destination)) {
	case visitDestinationLeaveSchool, visitDestinationReferred:
		return true
	default:
		return false
	}
}

func normalizeCallStatus(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "connected", "success", "completed":
		return "connected"
	case "requested", "accepted", "queued", "processing":
		return "requested"
	case "busy":
		return "busy"
	case "no_answer", "noanswer":
		return "no_answer"
	case "cancelled", "canceled":
		return "cancelled"
	case "failed", "error":
		return "failed"
	default:
		return "requested"
	}
}

func mustJSON(payload any) string {
	raw, err := json.Marshal(payload)
	if err != nil {
		return "{}"
	}
	return string(raw)
}

func optionalStringPtr(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func timePtr(value time.Time) *time.Time {
	clone := value.UTC()
	return &clone
}
