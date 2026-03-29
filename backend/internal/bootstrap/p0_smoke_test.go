package bootstrap

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"
	"time"

	"github.com/jiaxiang-medical-assistant/backend/internal/config"
)

type apiEnvelope struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func TestP0SmokeFlow(t *testing.T) {
	engine, cleanup := mustBuildSmokeServer(t)
	defer cleanup()

	doctorToken := loginAndGetToken(t, engine, "doctor", "doctor-pass-2026", "doctor")
	adminToken := loginAndGetToken(t, engine, "admin", "admin-pass-2026", "admin")

	healthResp := doJSONRequest(t, engine, http.MethodGet, "/api/v1/healthz", nil, "")
	requireStatus(t, healthResp, http.StatusOK)
	healthEnv := decodeEnvelope(t, healthResp)
	if healthEnv.Code != 0 {
		t.Fatalf("unexpected health code: %d", healthEnv.Code)
	}
	var health struct {
		Status string `json:"status"`
		Mode   string `json:"mode"`
	}
	unmarshalData(t, healthEnv, &health)
	if health.Status != "ok" || health.Mode != "mock" {
		t.Fatalf("unexpected health payload: status=%q mode=%q", health.Status, health.Mode)
	}

	unauthorizedResp := doJSONRequest(t, engine, http.MethodGet, "/api/v1/visits?page=1&page_size=10", nil, "")
	requireStatus(t, unauthorizedResp, http.StatusUnauthorized)
	unauthorizedEnv := decodeEnvelope(t, unauthorizedResp)
	if unauthorizedEnv.Code != 1002 {
		t.Fatalf("expected unauthorized code 1002, got %d", unauthorizedEnv.Code)
	}

	invalidCreateResp := doJSONRequest(t, engine, http.MethodPost, "/api/v1/visits", map[string]any{
		"symptoms": []string{"fever"},
	}, "")
	requireStatus(t, invalidCreateResp, http.StatusBadRequest)
	invalidCreateEnv := decodeEnvelope(t, invalidCreateResp)
	if invalidCreateEnv.Code != 1001 {
		t.Fatalf("expected invalid input code 1001, got %d", invalidCreateEnv.Code)
	}

	createResp := doJSONRequest(t, engine, http.MethodPost, "/api/v1/visits", map[string]any{
		"student_id":   "20269999",
		"symptoms":     []string{"headache", "cough"},
		"description":  "smoke test visit",
		"destination":  "observation",
		"follow_up_at": "",
	}, "")
	requireStatus(t, createResp, http.StatusOK)
	createEnv := decodeEnvelope(t, createResp)
	if createEnv.Code != 0 {
		t.Fatalf("unexpected create visit code: %d", createEnv.Code)
	}
	var createdVisit struct {
		ID        string `json:"id"`
		StudentID string `json:"student_id"`
	}
	unmarshalData(t, createEnv, &createdVisit)
	if createdVisit.ID == "" || createdVisit.StudentID != "20269999" {
		t.Fatalf("invalid created visit payload: %+v", createdVisit)
	}

	listResp := doJSONRequest(t, engine, http.MethodGet, "/api/v1/visits?page=1&page_size=20", nil, doctorToken)
	requireStatus(t, listResp, http.StatusOK)
	listEnv := decodeEnvelope(t, listResp)
	var listPayload struct {
		Items []struct {
			ID string `json:"id"`
		} `json:"items"`
	}
	unmarshalData(t, listEnv, &listPayload)
	visitIDs := make([]string, 0, len(listPayload.Items))
	for _, item := range listPayload.Items {
		visitIDs = append(visitIDs, item.ID)
	}
	if !slices.Contains(visitIDs, createdVisit.ID) {
		t.Fatalf("new visit %s not found in list: %v", createdVisit.ID, visitIDs)
	}

	followUpAt := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	updateResp := doJSONRequest(t, engine, http.MethodPatch, "/api/v1/visits/"+createdVisit.ID, map[string]any{
		"diagnosis":      "common cold",
		"prescription":   []string{"Ibuprofen Tablets"},
		"destination":    "home",
		"follow_up_at":   followUpAt,
		"follow_up_note": "recheck if fever continues",
	}, doctorToken)
	requireStatus(t, updateResp, http.StatusOK)
	updateEnv := decodeEnvelope(t, updateResp)
	if updateEnv.Code != 0 {
		t.Fatalf("unexpected update visit code: %d", updateEnv.Code)
	}

	detailResp := doJSONRequest(t, engine, http.MethodGet, "/api/v1/visits/"+createdVisit.ID, nil, doctorToken)
	requireStatus(t, detailResp, http.StatusOK)
	detailEnv := decodeEnvelope(t, detailResp)
	var detailPayload struct {
		ID          string     `json:"id"`
		Diagnosis   string     `json:"diagnosis"`
		Destination string     `json:"destination"`
		FollowUpAt  *time.Time `json:"follow_up_at"`
	}
	unmarshalData(t, detailEnv, &detailPayload)
	if detailPayload.Diagnosis != "common cold" || detailPayload.Destination != "home" || detailPayload.FollowUpAt == nil {
		t.Fatalf("visit update not persisted: %+v", detailPayload)
	}

	medicineListResp := doJSONRequest(t, engine, http.MethodGet, "/api/v1/medicines?page=1&page_size=20", nil, doctorToken)
	requireStatus(t, medicineListResp, http.StatusOK)
	medicineListEnv := decodeEnvelope(t, medicineListResp)
	var medicineListPayload struct {
		Items []struct {
			ID    string `json:"id"`
			Stock int    `json:"stock"`
		} `json:"items"`
	}
	unmarshalData(t, medicineListEnv, &medicineListPayload)
	if len(medicineListPayload.Items) == 0 {
		t.Fatalf("expected seeded medicines, got empty list")
	}
	medicineID := medicineListPayload.Items[0].ID
	stockBefore := medicineListPayload.Items[0].Stock

	outboundResp := doJSONRequest(t, engine, http.MethodPost, "/api/v1/medicines/outbound", map[string]any{
		"medicine_id": medicineID,
		"quantity":    1,
	}, doctorToken)
	requireStatus(t, outboundResp, http.StatusOK)
	outboundEnv := decodeEnvelope(t, outboundResp)
	var outboundPayload struct {
		ID    string `json:"id"`
		Stock int    `json:"stock"`
	}
	unmarshalData(t, outboundEnv, &outboundPayload)
	if outboundPayload.ID != medicineID || outboundPayload.Stock != stockBefore-1 {
		t.Fatalf("unexpected outbound result: before=%d after=%d payload=%+v", stockBefore, outboundPayload.Stock, outboundPayload)
	}

	overviewResp := doJSONRequest(t, engine, http.MethodGet, "/api/v1/reports/overview", nil, adminToken)
	requireStatus(t, overviewResp, http.StatusOK)
	overviewEnv := decodeEnvelope(t, overviewResp)
	if overviewEnv.Code != 0 {
		t.Fatalf("unexpected overview code: %d", overviewEnv.Code)
	}
	var overviewPayload struct {
		TodayVisits int64 `json:"today_visits"`
	}
	unmarshalData(t, overviewEnv, &overviewPayload)
	if overviewPayload.TodayVisits < 1 {
		t.Fatalf("unexpected today_visits: %d", overviewPayload.TodayVisits)
	}
}

func mustBuildSmokeServer(t *testing.T) (http.Handler, func()) {
	t.Helper()

	cfg := config.Config{
		AppName:  "medical-assistant-backend",
		AppEnv:   "test",
		AppPort:  8080,
		DataMode: "mock",
		Auth: config.AuthConfig{
			JWTSecret:      "smoke-test-secret-2026-please-change-in-prod",
			JWTExpiresIn:   3600,
			DoctorAccount:  "doctor",
			DoctorPassword: "doctor-pass-2026",
			AdminAccount:   "admin",
			AdminPassword:  "admin-pass-2026",
		},
	}

	engine, cleanup, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("build server failed: %v", err)
	}

	return engine, cleanup
}

func loginAndGetToken(t *testing.T, handler http.Handler, account string, password string, expectedRole string) string {
	t.Helper()

	resp := doJSONRequest(t, handler, http.MethodPost, "/api/v1/auth/login", map[string]string{
		"account":  account,
		"password": password,
	}, "")
	requireStatus(t, resp, http.StatusOK)
	env := decodeEnvelope(t, resp)
	if env.Code != 0 {
		t.Fatalf("unexpected login code: %d", env.Code)
	}

	var loginPayload struct {
		Token string `json:"token"`
		User  struct {
			Role string `json:"role"`
		} `json:"user"`
	}
	unmarshalData(t, env, &loginPayload)
	if loginPayload.Token == "" {
		t.Fatalf("empty login token for account %q", account)
	}
	if loginPayload.User.Role != expectedRole {
		t.Fatalf("unexpected role for %q: got %q want %q", account, loginPayload.User.Role, expectedRole)
	}

	return loginPayload.Token
}

func doJSONRequest(t *testing.T, handler http.Handler, method string, path string, body any, token string) *httptest.ResponseRecorder {
	t.Helper()

	var payload []byte
	if body != nil {
		var err error
		payload, err = json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal body failed: %v", err)
		}
	}

	req := httptest.NewRequest(method, path, bytes.NewReader(payload))
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func requireStatus(t *testing.T, rec *httptest.ResponseRecorder, expected int) {
	t.Helper()

	if rec.Code != expected {
		t.Fatalf("unexpected status: got %d want %d body=%s", rec.Code, expected, rec.Body.String())
	}
}

func decodeEnvelope(t *testing.T, rec *httptest.ResponseRecorder) apiEnvelope {
	t.Helper()

	var env apiEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("decode response failed: %v body=%s", err, rec.Body.String())
	}
	return env
}

func unmarshalData(t *testing.T, env apiEnvelope, target any) {
	t.Helper()

	if len(env.Data) == 0 {
		t.Fatalf("empty envelope data")
	}
	if err := json.Unmarshal(env.Data, target); err != nil {
		t.Fatalf("decode envelope data failed: %v data=%s", err, string(env.Data))
	}
}
