package service

import (
	"context"
	"sort"
	"strings"
)

type AIProvider interface {
	Analyze(context.Context, AnalyzeInput) (AnalyzeResult, error)
	Triage(context.Context, TriageInput) (TriageResult, error)
	Recommend(context.Context, RecommendInput) (RecommendResult, error)
	InteractionCheck(context.Context, InteractionCheckInput) (InteractionCheckResult, error)
}

type AIService struct {
	provider AIProvider
}

type AnalyzeInput struct {
	Symptoms    []string `json:"symptoms"`
	Description string   `json:"description"`
	Temperature float64  `json:"temperature"`
}

type AnalyzeResult struct {
	RiskLevel        string   `json:"risk_level"`
	Confidence       float64  `json:"confidence"`
	MatchedSignals   []string `json:"matched_signals"`
	PossibleCauses   []string `json:"possible_causes"`
	SuggestedActions []string `json:"suggested_actions"`
}

type TriageInput struct {
	Symptoms    []string `json:"symptoms"`
	Description string   `json:"description"`
	Temperature float64  `json:"temperature"`
}

type TriageResult struct {
	TriageLevel      string   `json:"triage_level"`
	Destination      string   `json:"destination"`
	Reason           string   `json:"reason"`
	ReviewInMinutes  int      `json:"review_in_minutes"`
	SuggestedActions []string `json:"suggested_actions"`
}

type RecommendInput struct {
	Diagnosis   string   `json:"diagnosis"`
	Symptoms    []string `json:"symptoms"`
	Destination string   `json:"destination"`
}

type RecommendResult struct {
	PlanVersion   string   `json:"plan_version"`
	CarePlan      []string `json:"care_plan"`
	MedicineHints []string `json:"medicine_hints"`
	FollowUp      string   `json:"follow_up"`
}

type InteractionCheckInput struct {
	Medicines []string `json:"medicines"`
}

type InteractionCheckResult struct {
	HasInteraction bool                    `json:"has_interaction"`
	RiskLevel      string                  `json:"risk_level"`
	Interactions   []MedicationInteraction `json:"interactions"`
	Advice         []string                `json:"advice"`
}

type MedicationInteraction struct {
	Pair     []string `json:"pair"`
	Severity string   `json:"severity"`
	Effect   string   `json:"effect"`
}

type ruleBasedAIProvider struct{}

var defaultAIProvider AIProvider = &ruleBasedAIProvider{}

func NewAIService() *AIService {
	return NewAIServiceWithProvider(nil)
}

func NewAIServiceWithProvider(provider AIProvider) *AIService {
	if provider == nil {
		provider = defaultAIProvider
	}

	return &AIService{provider: provider}
}

func (s *AIService) resolveProvider() AIProvider {
	if s == nil || s.provider == nil {
		return defaultAIProvider
	}

	return s.provider
}

func (s *AIService) Analyze(ctx context.Context, input AnalyzeInput) (AnalyzeResult, error) {
	return s.resolveProvider().Analyze(ctx, input)
}

func (s *AIService) Triage(ctx context.Context, input TriageInput) (TriageResult, error) {
	return s.resolveProvider().Triage(ctx, input)
}

func (s *AIService) Recommend(ctx context.Context, input RecommendInput) (RecommendResult, error) {
	return s.resolveProvider().Recommend(ctx, input)
}

func (s *AIService) InteractionCheck(ctx context.Context, input InteractionCheckInput) (InteractionCheckResult, error) {
	return s.resolveProvider().InteractionCheck(ctx, input)
}

func (p *ruleBasedAIProvider) Analyze(_ context.Context, input AnalyzeInput) (AnalyzeResult, error) {
	matched := collectSignals(input.Symptoms, input.Description)
	risk, confidence := estimateRisk(input.Temperature, matched)

	causes := []string{"common_cold"}
	actions := []string{"hydrate", "rest", "recheck_if_worse"}

	if containsAny(matched, highRiskSignals()) {
		causes = []string{"acute_condition", "urgent_assessment_needed"}
		actions = []string{"notify_guardian", "refer_to_hospital"}
	} else if risk == "medium" {
		causes = []string{"upper_respiratory_infection", "gastrointestinal_irritation"}
		actions = []string{"observe_in_clinic", "check_temperature_every_30m"}
	}

	return AnalyzeResult{
		RiskLevel:        risk,
		Confidence:       confidence,
		MatchedSignals:   matched,
		PossibleCauses:   causes,
		SuggestedActions: actions,
	}, nil
}

func (p *ruleBasedAIProvider) Triage(ctx context.Context, input TriageInput) (TriageResult, error) {
	analysis, err := p.Analyze(ctx, AnalyzeInput{
		Symptoms:    input.Symptoms,
		Description: input.Description,
		Temperature: input.Temperature,
	})
	if err != nil {
		return TriageResult{}, err
	}

	result := TriageResult{
		TriageLevel:      "routine",
		Destination:      "classroom",
		Reason:           "no_high_risk_signal",
		ReviewInMinutes:  120,
		SuggestedActions: []string{"return_to_class", "self_report_if_unwell"},
	}

	switch analysis.RiskLevel {
	case "high":
		result.TriageLevel = "urgent"
		result.Destination = "hospital"
		result.Reason = "high_risk_signal_detected"
		result.ReviewInMinutes = 10
		result.SuggestedActions = []string{"contact_guardian_now", "prepare_transfer"}
	case "medium":
		result.TriageLevel = "priority"
		result.Destination = "observation"
		result.Reason = "needs_observation"
		result.ReviewInMinutes = 30
		result.SuggestedActions = []string{"stay_in_observation", "repeat_vitals"}
	}

	return result, nil
}

func (p *ruleBasedAIProvider) Recommend(_ context.Context, input RecommendInput) (RecommendResult, error) {
	destination := strings.TrimSpace(strings.ToLower(input.Destination))
	if destination == "" {
		destination = "observation"
	}

	plan := []string{"record_symptoms", "ensure_hydration"}
	medicineHints := []string{"acetaminophen_if_fever", "oral_rehydration_if_needed"}
	followUp := "review_in_2_hours"

	if strings.Contains(strings.ToLower(strings.TrimSpace(input.Diagnosis)), "allergy") {
		plan = append(plan, "avoid_known_allergen")
		medicineHints = []string{"cetirizine_consideration"}
		followUp = "review_in_1_hour"
	}

	if destination == "hospital" {
		plan = []string{"stabilize_patient", "handover_to_hospital"}
		medicineHints = []string{"follow_transfer_protocol"}
		followUp = "follow_hospital_feedback"
	}

	return RecommendResult{
		PlanVersion:   "mock-v1",
		CarePlan:      plan,
		MedicineHints: medicineHints,
		FollowUp:      followUp,
	}, nil
}

func (p *ruleBasedAIProvider) InteractionCheck(_ context.Context, input InteractionCheckInput) (InteractionCheckResult, error) {
	interactions := make([]MedicationInteraction, 0)

	normalized := make([]string, 0, len(input.Medicines))
	for _, name := range input.Medicines {
		trimmed := strings.TrimSpace(strings.ToLower(name))
		if trimmed == "" {
			continue
		}
		normalized = append(normalized, trimmed)
	}

	for i := 0; i < len(normalized); i++ {
		for j := i + 1; j < len(normalized); j++ {
			if interaction, ok := lookupInteraction(normalized[i], normalized[j]); ok {
				interactions = append(interactions, interaction)
			}
		}
	}

	result := InteractionCheckResult{
		HasInteraction: len(interactions) > 0,
		RiskLevel:      "low",
		Interactions:   interactions,
		Advice:         []string{"keep_medication_record"},
	}
	if len(interactions) > 0 {
		result.RiskLevel = "medium"
		result.Advice = []string{"verify_with_doctor", "avoid_coadministration_until_confirmed"}
	}

	return result, nil
}

func highRiskSignals() []string {
	return []string{"chest pain", "difficulty breathing", "convulsion", "fainting", "vomit blood"}
}

func collectSignals(symptoms []string, description string) []string {
	matches := map[string]struct{}{}
	for _, symptom := range symptoms {
		value := strings.TrimSpace(strings.ToLower(symptom))
		if value == "" {
			continue
		}
		matches[value] = struct{}{}
	}

	desc := strings.ToLower(strings.TrimSpace(description))
	for _, key := range highRiskSignals() {
		if strings.Contains(desc, key) {
			matches[key] = struct{}{}
		}
	}

	result := make([]string, 0, len(matches))
	for key := range matches {
		result = append(result, key)
	}
	sort.Strings(result)
	return result
}

func estimateRisk(temperature float64, matched []string) (string, float64) {
	if temperature >= 39.5 || containsAny(matched, highRiskSignals()) {
		return "high", 0.91
	}
	if temperature >= 38.0 || len(matched) >= 3 {
		return "medium", 0.78
	}
	return "low", 0.62
}

func containsAny(values []string, keywords []string) bool {
	keywordSet := map[string]struct{}{}
	for _, keyword := range keywords {
		keywordSet[strings.ToLower(strings.TrimSpace(keyword))] = struct{}{}
	}

	for _, value := range values {
		if _, ok := keywordSet[strings.ToLower(strings.TrimSpace(value))]; ok {
			return true
		}
	}

	return false
}

func lookupInteraction(medicineA string, medicineB string) (MedicationInteraction, bool) {
	pair := []string{medicineA, medicineB}
	sort.Strings(pair)
	key := strings.Join(pair, "+")

	interactions := map[string]MedicationInteraction{
		"aspirin+ibuprofen": {
			Pair:     []string{"aspirin", "ibuprofen"},
			Severity: "medium",
			Effect:   "increased_gastrointestinal_risk",
		},
		"cetirizine+chlorpheniramine": {
			Pair:     []string{"cetirizine", "chlorpheniramine"},
			Severity: "medium",
			Effect:   "excessive_drowsiness",
		},
		"acetaminophen+ibuprofen": {
			Pair:     []string{"acetaminophen", "ibuprofen"},
			Severity: "low",
			Effect:   "duplicate_analgesic_monitoring_needed",
		},
	}

	interaction, ok := interactions[key]
	return interaction, ok
}
