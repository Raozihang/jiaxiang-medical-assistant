package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/jiaxiang-medical-assistant/backend/internal/config"
)

type OutboundCallProvider interface {
	ProviderName() string
	PlaceCall(ctx context.Context, input PlaceOutboundCallInput) (PlaceOutboundCallResult, error)
}

type PlaceOutboundCallInput struct {
	Phone        string
	TemplateCode string
	Message      string
	TemplateVars map[string]string
}

type PlaceOutboundCallResult struct {
	Provider    string
	Status      string
	RequestID   string
	CallID      string
	ResponseRaw string
}

type MockOutboundCallProvider struct{}

func NewMockOutboundCallProvider() *MockOutboundCallProvider {
	return &MockOutboundCallProvider{}
}

func (p *MockOutboundCallProvider) ProviderName() string {
	return "mock"
}

func (p *MockOutboundCallProvider) PlaceCall(_ context.Context, input PlaceOutboundCallInput) (PlaceOutboundCallResult, error) {
	raw, _ := json.Marshal(map[string]any{
		"called_number": input.Phone,
		"template_code": input.TemplateCode,
		"message":       input.Message,
		"template_vars": input.TemplateVars,
	})
	return PlaceOutboundCallResult{
		Provider:    p.ProviderName(),
		Status:      "requested",
		RequestID:   fmt.Sprintf("mock-%d", time.Now().UTC().UnixNano()),
		CallID:      fmt.Sprintf("mock-call-%d", time.Now().UTC().UnixNano()),
		ResponseRaw: string(raw),
	}, nil
}

type AliyunOutboundCallProvider struct {
	client           *sdk.Client
	calledShowNumber string
	ttsCode          string
	playTimes        int
}

func NewAliyunOutboundCallProvider(cfg config.OutboundCallConfig) (*AliyunOutboundCallProvider, error) {
	if strings.TrimSpace(cfg.AliyunAccessKeyID) == "" || strings.TrimSpace(cfg.AliyunAccessKeySecret) == "" {
		return nil, errors.New("aliyun outbound call credentials are not configured")
	}
	if strings.TrimSpace(cfg.AliyunCalledShowNumber) == "" {
		return nil, errors.New("aliyun outbound call number is not configured")
	}
	if strings.TrimSpace(cfg.AliyunTTSCode) == "" {
		return nil, errors.New("aliyun outbound call TTS code is not configured")
	}

	client, err := sdk.NewClientWithAccessKey(strings.TrimSpace(cfg.AliyunRegionID), strings.TrimSpace(cfg.AliyunAccessKeyID), strings.TrimSpace(cfg.AliyunAccessKeySecret))
	if err != nil {
		return nil, err
	}

	playTimes := cfg.AliyunPlayTimes
	if playTimes <= 0 {
		playTimes = 2
	}

	return &AliyunOutboundCallProvider{
		client:           client,
		calledShowNumber: strings.TrimSpace(cfg.AliyunCalledShowNumber),
		ttsCode:          strings.TrimSpace(cfg.AliyunTTSCode),
		playTimes:        playTimes,
	}, nil
}

func (p *AliyunOutboundCallProvider) ProviderName() string {
	return "aliyun"
}

func (p *AliyunOutboundCallProvider) PlaceCall(ctx context.Context, input PlaceOutboundCallInput) (PlaceOutboundCallResult, error) {
	ttsParamBytes, err := json.Marshal(input.TemplateVars)
	if err != nil {
		return PlaceOutboundCallResult{}, err
	}

	request := requests.NewCommonRequest()
	request.Method = "POST"
	request.Scheme = "https"
	request.Domain = "dyvmsapi.aliyuncs.com"
	request.Version = "2017-05-25"
	request.ApiName = "SingleCallByTts"
	request.Product = "Dyvmsapi"
	request.QueryParams["CalledShowNumber"] = p.calledShowNumber
	request.QueryParams["CalledNumber"] = normalizePhone(input.Phone)
	request.QueryParams["TtsCode"] = p.ttsCode
	request.QueryParams["TtsParam"] = string(ttsParamBytes)
	request.QueryParams["PlayTimes"] = fmt.Sprintf("%d", p.playTimes)

	response, err := p.client.ProcessCommonRequest(request)
	if err != nil {
		return PlaceOutboundCallResult{}, err
	}

	body := response.GetHttpContentString()
	var payload struct {
		Code      string `json:"Code"`
		Message   string `json:"Message"`
		RequestID string `json:"RequestId"`
		CallID    string `json:"CallId"`
	}
	if err := json.Unmarshal([]byte(body), &payload); err != nil {
		return PlaceOutboundCallResult{}, err
	}
	if !strings.EqualFold(payload.Code, "OK") {
		if strings.TrimSpace(payload.Message) == "" {
			payload.Message = payload.Code
		}
		return PlaceOutboundCallResult{}, errors.New(strings.TrimSpace(payload.Message))
	}

	return PlaceOutboundCallResult{
		Provider:    p.ProviderName(),
		Status:      "requested",
		RequestID:   strings.TrimSpace(payload.RequestID),
		CallID:      strings.TrimSpace(payload.CallID),
		ResponseRaw: body,
	}, nil
}

func normalizePhone(phone string) string {
	trimmed := strings.TrimSpace(phone)
	trimmed = strings.TrimPrefix(trimmed, "+86")
	return strings.TrimSpace(trimmed)
}
