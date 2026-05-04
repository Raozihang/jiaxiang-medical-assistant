package service

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/jiaxiang-medical-assistant/backend/internal/repository"
)

const defaultRealtimeVisitPageSize = 100

type RealtimeHub struct {
	mu             sync.RWMutex
	doctorClients  map[*RealtimeClient]struct{}
	checkInClients map[*RealtimeClient]struct{}
}

type RealtimeClient struct {
	send chan RealtimeMessage
}

type RealtimeMessage struct {
	Type    string `json:"type"`
	Payload any    `json:"payload,omitempty"`
	SentAt  string `json:"sent_at"`
}

type VisitSnapshotPayload struct {
	Reason       string             `json:"reason"`
	ChangedVisit *repository.Visit  `json:"changed_visit,omitempty"`
	Items        []repository.Visit `json:"items"`
	Page         int                `json:"page"`
	PageSize     int                `json:"page_size"`
	Total        int64              `json:"total"`
}

func NewRealtimeHub() *RealtimeHub {
	return &RealtimeHub{
		doctorClients:  map[*RealtimeClient]struct{}{},
		checkInClients: map[*RealtimeClient]struct{}{},
	}
}

func NewRealtimeMessage(messageType string, payload any) RealtimeMessage {
	return RealtimeMessage{
		Type:    messageType,
		Payload: payload,
		SentAt:  time.Now().UTC().Format(time.RFC3339),
	}
}

func (h *RealtimeHub) SubscribeDoctor() (*RealtimeClient, func()) {
	return h.subscribe(&h.doctorClients)
}

func (h *RealtimeHub) SubscribeCheckIn() (*RealtimeClient, func()) {
	return h.subscribe(&h.checkInClients)
}

func (h *RealtimeHub) subscribe(target *map[*RealtimeClient]struct{}) (*RealtimeClient, func()) {
	client := &RealtimeClient{send: make(chan RealtimeMessage, 16)}

	h.mu.Lock()
	(*target)[client] = struct{}{}
	h.mu.Unlock()

	return client, func() {
		h.mu.Lock()
		if _, ok := (*target)[client]; ok {
			delete(*target, client)
			close(client.send)
		}
		h.mu.Unlock()
	}
}

func (c *RealtimeClient) Send() <-chan RealtimeMessage {
	return c.send
}

func (h *RealtimeHub) BroadcastToDoctors(message RealtimeMessage) {
	h.broadcast(message, h.doctorClients)
}

func (h *RealtimeHub) BroadcastToCheckIns(message RealtimeMessage) {
	h.broadcast(message, h.checkInClients)
}

func (h *RealtimeHub) broadcast(message RealtimeMessage, target map[*RealtimeClient]struct{}) {
	h.mu.RLock()
	clients := make([]*RealtimeClient, 0, len(target))
	for client := range target {
		clients = append(clients, client)
	}
	h.mu.RUnlock()

	for _, client := range clients {
		select {
		case client.send <- message:
		default:
			log.Printf("realtime doctor client channel full; dropping %s message", message.Type)
		}
	}
}

func (s *VisitService) SetRealtimeHub(hub *RealtimeHub) {
	s.realtimeHub = hub
}

func (s *VisitService) SetAIAnalysisQueue(queue AIAnalysisQueue) {
	s.aiAnalysisQueue = queue
}

func (s *VisitService) broadcastVisitsSnapshot(ctx context.Context, reason string, changedVisit *repository.Visit) {
	if s.realtimeHub == nil {
		return
	}

	result, err := s.List(ctx, VisitListInput{
		PageParams: repository.PageParams{
			Page:     1,
			PageSize: defaultRealtimeVisitPageSize,
		},
	})
	if err != nil {
		log.Printf("build realtime visit snapshot failed: %v", err)
		return
	}

	s.realtimeHub.BroadcastToDoctors(NewRealtimeMessage("visits_snapshot", VisitSnapshotPayload{
		Reason:       reason,
		ChangedVisit: changedVisit,
		Items:        result.Items,
		Page:         result.Page,
		PageSize:     result.PageSize,
		Total:        result.Total,
	}))
}
