package service

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jiaxiang-medical-assistant/backend/internal/repository"
)

type NotificationService struct {
	logRepo repository.NotificationLogRepository
}

type SendNotificationInput struct {
	Channel  string `json:"channel"`
	Receiver string `json:"receiver"`
	Message  string `json:"message"`
}

type DispatchScenarioInput struct {
	Scenario    string `json:"scenario"`
	Channel     string `json:"channel"`
	Receiver    string `json:"receiver"`
	StudentName string `json:"student_name"`
	Destination string `json:"destination"`
	FollowUpAt  string `json:"follow_up_at"`
	Note        string `json:"note"`
}

func NewNotificationService(logRepo repository.NotificationLogRepository) *NotificationService {
	return &NotificationService{logRepo: logRepo}
}

func (s *NotificationService) DispatchScenario(ctx context.Context, input DispatchScenarioInput) (repository.NotificationLog, error) {
	message, err := buildScenarioMessage(input)
	if err != nil {
		return repository.NotificationLog{}, err
	}

	return s.Send(ctx, SendNotificationInput{
		Channel:  input.Channel,
		Receiver: input.Receiver,
		Message:  message,
	})
}

func (s *NotificationService) Send(ctx context.Context, input SendNotificationInput) (repository.NotificationLog, error) {
	channel := strings.ToLower(strings.TrimSpace(input.Channel))
	receiver := strings.TrimSpace(input.Receiver)
	message := strings.TrimSpace(input.Message)
	now := time.Now().UTC()

	log := repository.NotificationLog{
		ID:       uuid.NewString(),
		Channel:  channel,
		Receiver: receiver,
		Message:  message,
		Status:   "failed",
		SentAt:   now,
	}

	if channel != "wechat" && channel != "dingtalk" {
		log.Error = "不支持的通知渠道"
		_ = s.logRepo.Append(ctx, log)
		return repository.NotificationLog{}, ErrInvalidInput
	}
	if receiver == "" || message == "" {
		log.Error = "接收人和消息内容不能为空"
		_ = s.logRepo.Append(ctx, log)
		return repository.NotificationLog{}, ErrInvalidInput
	}

	log.Status = "sent"

	if err := s.logRepo.Append(ctx, log); err != nil {
		return repository.NotificationLog{}, err
	}

	return log, nil
}

func (s *NotificationService) ListLogs(ctx context.Context) ([]repository.NotificationLog, error) {
	logs, err := s.logRepo.List(ctx)
	if err != nil {
		return nil, err
	}

	sort.Slice(logs, func(i, j int) bool {
		return logs[i].SentAt.After(logs[j].SentAt)
	})

	return logs, nil
}

func buildScenarioMessage(input DispatchScenarioInput) (string, error) {
	scenario := strings.ToLower(strings.TrimSpace(input.Scenario))
	studentName := strings.TrimSpace(input.StudentName)
	if studentName == "" {
		studentName = "该学生"
	}
	destination := strings.TrimSpace(input.Destination)
	followUpAt := strings.TrimSpace(input.FollowUpAt)
	note := strings.TrimSpace(input.Note)

	switch scenario {
	case "visit_completed":
		parts := []string{studentName + "已完成本次就诊。"}
		if destination != "" {
			parts = append(parts, "去向："+destination+"。")
		}
		if note != "" {
			parts = append(parts, "备注："+note+"。")
		}
		return strings.Join(parts, " "), nil
	case "observation_notice":
		parts := []string{studentName + "目前需要留观，请及时关注。"}
		if destination != "" {
			parts = append(parts, "留观点："+destination+"。")
		}
		if note != "" {
			parts = append(parts, "说明："+note+"。")
		}
		return strings.Join(parts, " "), nil
	case "follow_up_reminder":
		parts := []string{"请提醒" + studentName + "按时复诊。"}
		if followUpAt != "" {
			parts = append(parts, "复诊时间："+followUpAt+"。")
		}
		if destination != "" {
			parts = append(parts, "复诊地点："+destination+"。")
		}
		if note != "" {
			parts = append(parts, "备注："+note+"。")
		}
		return strings.Join(parts, " "), nil
	default:
		return "", ErrInvalidInput
	}
}
