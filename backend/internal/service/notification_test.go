package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/jiaxiang-medical-assistant/backend/internal/repository"
)

func TestNotificationServiceDispatchScenarioSuccess(t *testing.T) {
	tests := []struct {
		name     string
		input    DispatchScenarioInput
		contains []string
	}{
		{
			name: "visit completed",
			input: DispatchScenarioInput{
				Scenario:    "visit_completed",
				Channel:     "wechat",
				Receiver:    "teacher-1",
				StudentName: "张三",
				Destination: "教室",
				Note:        "状态稳定",
			},
			contains: []string{"张三已完成本次就诊", "去向：教室", "备注：状态稳定"},
		},
		{
			name: "follow up reminder",
			input: DispatchScenarioInput{
				Scenario:    "follow_up_reminder",
				Channel:     "dingtalk",
				Receiver:    "parent-1",
				StudentName: "李四",
				FollowUpAt:  "2026-02-25 14:00",
				Note:        "请家长陪同",
			},
			contains: []string{"请提醒李四按时复诊", "复诊时间：2026-02-25 14:00", "备注：请家长陪同"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := repository.NewMemoryNotificationLogRepository()
			svc := NewNotificationService(repo)

			log, err := svc.DispatchScenario(context.Background(), tt.input)
			if err != nil {
				t.Fatalf("dispatch scenario failed: %v", err)
			}
			if log.Status != "sent" {
				t.Fatalf("expected sent status, got %s", log.Status)
			}
			for _, want := range tt.contains {
				if !strings.Contains(log.Message, want) {
					t.Fatalf("expected message to contain %q, got %q", want, log.Message)
				}
			}

			logs, err := repo.List(context.Background())
			if err != nil {
				t.Fatalf("list logs failed: %v", err)
			}
			if len(logs) != 1 {
				t.Fatalf("expected 1 log, got %d", len(logs))
			}
		})
	}
}

func TestNotificationServiceDispatchScenarioInvalidScenario(t *testing.T) {
	repo := repository.NewMemoryNotificationLogRepository()
	svc := NewNotificationService(repo)

	_, err := svc.DispatchScenario(context.Background(), DispatchScenarioInput{
		Scenario: "unknown_scenario",
		Channel:  "wechat",
		Receiver: "teacher-1",
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}
