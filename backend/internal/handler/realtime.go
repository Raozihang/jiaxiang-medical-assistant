package handler

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/jiaxiang-medical-assistant/backend/internal/repository"
	"github.com/jiaxiang-medical-assistant/backend/internal/service"
)

type RealtimeHandler struct {
	hub          *service.RealtimeHub
	visitService *service.VisitService
	authService  *service.AuthService
	upgrader     websocket.Upgrader
}

func NewRealtimeHandler(hub *service.RealtimeHub, visitService *service.VisitService, authService *service.AuthService) *RealtimeHandler {
	return &RealtimeHandler{
		hub:          hub,
		visitService: visitService,
		authService:  authService,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(_ *http.Request) bool {
				return true
			},
		},
	}
}

func (h *RealtimeHandler) Doctor(c *gin.Context) {
	if !h.authorizeDoctor(c) {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "unauthorized"})
		return
	}

	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	client, unsubscribe := h.hub.SubscribeDoctor()
	defer unsubscribe()

	done := make(chan struct{})
	go h.readDoctorMessages(conn, done)

	if snapshot, err := h.visitsSnapshot(c.Request.Context(), "connected", nil); err == nil {
		if err := conn.WriteJSON(service.NewRealtimeMessage("visits_snapshot", snapshot)); err != nil {
			return
		}
	}

	for {
		select {
		case <-done:
			return
		case message, ok := <-client.Send():
			if !ok {
				return
			}
			if err := conn.WriteJSON(message); err != nil {
				return
			}
		}
	}
}

func (h *RealtimeHandler) CheckIn(c *gin.Context) {
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	client, unsubscribe := h.hub.SubscribeCheckIn()
	defer unsubscribe()

	done := make(chan struct{})
	go h.readCheckInMessages(conn, done)

	if err := conn.WriteJSON(service.NewRealtimeMessage("connected", gin.H{"channel": "checkin"})); err != nil {
		return
	}

	for {
		select {
		case <-done:
			return
		case message, ok := <-client.Send():
			if !ok {
				return
			}
			if err := conn.WriteJSON(message); err != nil {
				return
			}
		}
	}
}

func (h *RealtimeHandler) readDoctorMessages(conn *websocket.Conn, done chan<- struct{}) {
	defer close(done)
	for {
		var message service.RealtimeMessage
		if err := conn.ReadJSON(&message); err != nil {
			return
		}
		messageType := strings.TrimSpace(message.Type)
		switch messageType {
		case "temperature_requested", "temperature_recorded":
			h.hub.BroadcastToCheckIns(service.NewRealtimeMessage(messageType, message.Payload))
			h.hub.BroadcastToDoctors(service.NewRealtimeMessage(messageType, message.Payload))
		}
	}
}

func (h *RealtimeHandler) readCheckInMessages(conn *websocket.Conn, done chan<- struct{}) {
	defer close(done)
	for {
		var message service.RealtimeMessage
		if err := conn.ReadJSON(&message); err != nil {
			return
		}
		messageType := strings.TrimSpace(message.Type)
		switch messageType {
		case "checkin_progress", "temperature_due":
			h.hub.BroadcastToDoctors(service.NewRealtimeMessage(messageType, message.Payload))
		}
	}
}

func (h *RealtimeHandler) visitsSnapshot(ctx context.Context, reason string, changedVisit *repository.Visit) (service.VisitSnapshotPayload, error) {
	result, err := h.visitService.List(ctx, service.VisitListInput{
		PageParams: repository.PageParams{Page: 1, PageSize: 100},
	})
	if err != nil {
		log.Printf("initial realtime visit snapshot failed: %v", err)
		return service.VisitSnapshotPayload{}, err
	}

	return service.VisitSnapshotPayload{
		Reason:       reason,
		ChangedVisit: changedVisit,
		Items:        result.Items,
		Page:         result.Page,
		PageSize:     result.PageSize,
		Total:        result.Total,
	}, nil
}

func (h *RealtimeHandler) authorizeDoctor(c *gin.Context) bool {
	if h.authService == nil {
		return false
	}
	token := strings.TrimSpace(c.Query("token"))
	if token == "" {
		return false
	}
	claims, err := h.authService.VerifyToken(token)
	if err != nil {
		return false
	}
	return claims.Role == "doctor" || claims.Role == "admin"
}
