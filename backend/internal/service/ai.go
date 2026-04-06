package service

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jiaxiang-medical-assistant/backend/internal/repository"
)

type AIProvider interface {
	Analyze(context.Context, AnalyzeInput) (AnalyzeResult, error)
	Triage(context.Context, TriageInput) (TriageResult, error)
	Recommend(context.Context, RecommendInput) (RecommendResult, error)
	InteractionCheck(context.Context, InteractionCheckInput) (InteractionCheckResult, error)
}

type AIService struct {
	provider     AIProvider
	medicineRepo repository.MedicineRepository
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

type MedicineKnowledge struct {
	Name                 string    `json:"name"`
	Specification        string    `json:"specification"`
	Stock                int       `json:"stock"`
	SafeStock            int       `json:"safe_stock"`
	ExpiryDate           time.Time `json:"expiry_date"`
	Warnings             []string  `json:"warnings"`
	RecommendedDosage    string    `json:"recommended_dosage"`
	RecommendedFrequency string    `json:"recommended_frequency"`
	RecommendedDuration  string    `json:"recommended_duration"`
	UsageInstructions    string    `json:"usage_instructions"`
	IsLowStock           bool      `json:"is_low_stock"`
	IsExpiringSoon       bool      `json:"is_expiring_soon"`
}

type RecommendInput struct {
	Diagnosis   string   `json:"diagnosis"`
	Symptoms    []string `json:"symptoms"`
	Destination string   `json:"destination"`
	TriageLevel string   `json:"triage_level"`
	Allergies   []string `json:"allergies"`

	AvailableMedicines []MedicineKnowledge `json:"-"`
	RAGContext         string              `json:"-"`
	UseWebSearch       bool                `json:"-"`
	RiskFlags          []string            `json:"-"`
}

type MedicineRecommendation struct {
	Name      string `json:"name"`
	Dosage    string `json:"dosage"`
	Frequency string `json:"frequency"`
	Duration  string `json:"duration"`
	Reason    string `json:"reason"`
	Caution   string `json:"caution"`
}

type RecommendResult struct {
	PlanVersion       string                   `json:"plan_version"`
	CarePlan          []string                 `json:"care_plan"`
	MedicineHints     []string                 `json:"medicine_hints"`
	FollowUp          string                   `json:"follow_up"`
	Medicines         []MedicineRecommendation `json:"medicines"`
	Advice            []string                 `json:"advice"`
	Contraindications []string                 `json:"contraindications"`
	RiskFlags         []string                 `json:"risk_flags"`
	InventoryBasis    []string                 `json:"inventory_basis"`
	UsedWebSearch     bool                     `json:"used_web_search"`
}

type InteractionCheckInput struct {
	Medicines []string `json:"medicines"`
	StudentID string   `json:"student_id"`

	AvailableMedicines []MedicineKnowledge `json:"-"`
	RAGContext         string              `json:"-"`
	UseWebSearch       bool                `json:"-"`
	RiskFlags          []string            `json:"-"`
}

type InteractionWarning struct {
	Title       string `json:"title"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
	Suggestion  string `json:"suggestion"`
}

type InteractionCheckResult struct {
	HasInteraction bool                    `json:"has_interaction"`
	RiskLevel      string                  `json:"risk_level"`
	Interactions   []MedicationInteraction `json:"interactions"`
	Advice         []string                `json:"advice"`
	Warnings       []InteractionWarning    `json:"warnings"`
	Safe           bool                    `json:"safe"`
	RiskFlags      []string                `json:"risk_flags"`
	UsedWebSearch  bool                    `json:"used_web_search"`
}

type MedicationInteraction struct {
	Pair     []string `json:"pair"`
	Severity string   `json:"severity"`
	Effect   string   `json:"effect"`
}

type ruleBasedAIProvider struct{}

var defaultAIProvider AIProvider = &ruleBasedAIProvider{}

func NewAIService() *AIService {
	return NewAIServiceWithDependencies(nil, nil)
}

func NewAIServiceWithProvider(provider AIProvider) *AIService {
	return NewAIServiceWithDependencies(provider, nil)
}

func NewAIServiceWithDependencies(provider AIProvider, medicineRepo repository.MedicineRepository) *AIService {
	if provider == nil {
		provider = defaultAIProvider
	}
	return &AIService{provider: provider, medicineRepo: medicineRepo}
}

func (s *AIService) resolveProvider() AIProvider {
	if s == nil || s.provider == nil {
		return defaultAIProvider
	}
	return s.provider
}

func (s *AIService) canUseWebSearch() bool {
	return s.resolveProvider() != nil && s.resolveProvider() != defaultAIProvider
}

func (s *AIService) Analyze(ctx context.Context, input AnalyzeInput) (AnalyzeResult, error) {
	provider := s.resolveProvider()
	result, err := provider.Analyze(ctx, input)
	if err == nil {
		return result, nil
	}
	log.Printf("AI analyze failed with provider %T, falling back to rule-based provider: %v", provider, err)
	return defaultAIProvider.Analyze(ctx, input)
}

func (s *AIService) Triage(ctx context.Context, input TriageInput) (TriageResult, error) {
	provider := s.resolveProvider()
	result, err := provider.Triage(ctx, input)
	if err == nil {
		return result, nil
	}
	log.Printf("AI triage failed with provider %T, falling back to rule-based provider: %v", provider, err)
	return defaultAIProvider.Triage(ctx, input)
}

func (s *AIService) Recommend(ctx context.Context, input RecommendInput) (RecommendResult, error) {
	inventory := s.loadMedicineKnowledge(ctx)
	input.AvailableMedicines = recommendableMedicines(inventory)
	input.RAGContext = medicineRAGContext(input.AvailableMedicines)

	result, err := s.callRecommend(ctx, input)
	if err != nil {
		return RecommendResult{}, err
	}
	result = normalizeRecommendResult(result, input, inventory)

	if s.canUseWebSearch() && recommendNeedsWebSearch(input, result) {
		input.UseWebSearch = true
		input.RiskFlags = uniqueStrings(append(input.RiskFlags, result.RiskFlags...))
		webResult, webErr := s.resolveProvider().Recommend(ctx, input)
		if webErr != nil {
			log.Printf("AI recommend web search retry failed, keep local RAG result: %v", webErr)
		} else {
			webResult.UsedWebSearch = true
			result = normalizeRecommendResult(webResult, input, inventory)
			result.UsedWebSearch = true
		}
	}

	return result, nil
}

func (s *AIService) InteractionCheck(ctx context.Context, input InteractionCheckInput) (InteractionCheckResult, error) {
	inventory := s.loadMedicineKnowledge(ctx)
	input.AvailableMedicines, input.RiskFlags = selectedMedicines(input.Medicines, inventory)
	input.RAGContext = selectedMedicineRAGContext(input.Medicines, input.AvailableMedicines, input.RiskFlags)

	local, err := defaultAIProvider.InteractionCheck(ctx, input)
	if err != nil {
		return InteractionCheckResult{}, err
	}
	if s.canUseWebSearch() && interactionNeedsWebSearch(local, input.RiskFlags) {
		input.UseWebSearch = true
	}

	result, err := s.callInteractionCheck(ctx, input)
	if err != nil {
		return InteractionCheckResult{}, err
	}
	result = normalizeInteractionResult(result, local, input.AvailableMedicines, input.RiskFlags)
	if input.UseWebSearch {
		result.UsedWebSearch = true
	}
	return result, nil
}

func (s *AIService) callRecommend(ctx context.Context, input RecommendInput) (RecommendResult, error) {
	provider := s.resolveProvider()
	result, err := provider.Recommend(ctx, input)
	if err == nil {
		return result, nil
	}
	if provider == defaultAIProvider {
		return RecommendResult{}, err
	}
	log.Printf("AI recommend failed with provider %T, falling back to rule-based provider: %v", provider, err)
	return defaultAIProvider.Recommend(ctx, input)
}

func (s *AIService) callInteractionCheck(ctx context.Context, input InteractionCheckInput) (InteractionCheckResult, error) {
	provider := s.resolveProvider()
	result, err := provider.InteractionCheck(ctx, input)
	if err == nil {
		return result, nil
	}
	if provider == defaultAIProvider {
		return InteractionCheckResult{}, err
	}
	log.Printf("AI interaction check failed with provider %T, falling back to rule-based provider: %v", provider, err)
	return defaultAIProvider.InteractionCheck(ctx, input)
}

func (s *AIService) loadMedicineKnowledge(ctx context.Context) []MedicineKnowledge {
	if s == nil || s.medicineRepo == nil {
		return nil
	}
	rows, err := s.medicineRepo.ListAll(ctx)
	if err != nil {
		log.Printf("load medicine knowledge failed: %v", err)
		return nil
	}
	items := make([]MedicineKnowledge, 0, len(rows))
	for _, row := range rows {
		items = append(items, MedicineKnowledge{
			Name:                 strings.TrimSpace(row.Name),
			Specification:        strings.TrimSpace(row.Specification),
			Stock:                row.Stock,
			SafeStock:            row.SafeStock,
			ExpiryDate:           row.ExpiryDate,
			Warnings:             uniqueStrings(row.Warnings),
			RecommendedDosage:    strings.TrimSpace(row.RecommendedDosage),
			RecommendedFrequency: strings.TrimSpace(row.RecommendedFrequency),
			RecommendedDuration:  strings.TrimSpace(row.RecommendedDuration),
			UsageInstructions:    strings.TrimSpace(row.UsageInstructions),
			IsLowStock:           row.IsLowStock,
			IsExpiringSoon:       row.IsExpiringSoon,
		})
	}
	return items
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
	result := RecommendResult{
		PlanVersion: "mock-rag-v1",
		Advice:      []string{"record_symptoms", "ensure_hydration"},
		FollowUp:    "review_in_2_hours",
	}
	if isHighRiskTriage(input.TriageLevel, input.Destination) {
		result.Advice = []string{"stabilize_patient", "handover_to_hospital"}
		result.FollowUp = "follow_hospital_feedback"
		return syncRecommendLegacy(result), nil
	}
	if strings.Contains(strings.ToLower(strings.TrimSpace(input.Diagnosis)), "allergy") {
		result.Advice = append(result.Advice, "avoid_known_allergen")
	}
	for _, item := range recommendableMedicines(input.AvailableMedicines) {
		if isSupply(item) {
			continue
		}
		result.Medicines = append(result.Medicines, MedicineRecommendation{
			Name:      item.Name,
			Dosage:    item.RecommendedDosage,
			Frequency: item.RecommendedFrequency,
			Duration:  item.RecommendedDuration,
			Reason:    "local_inventory_safe_candidate",
			Caution:   strings.Join(item.Warnings, "; "),
		})
		break
	}
	if len(result.Medicines) == 0 {
		result.Contraindications = []string{"no_safe_local_medicine_candidate"}
	}
	return syncRecommendLegacy(result), nil
}

func (p *ruleBasedAIProvider) InteractionCheck(_ context.Context, input InteractionCheckInput) (InteractionCheckResult, error) {
	interactions := make([]MedicationInteraction, 0)
	names := make([]string, 0, len(input.Medicines))
	for _, name := range input.Medicines {
		if key := canonicalMedicineName(name); key != "" {
			names = append(names, key)
		}
	}
	for i := 0; i < len(names); i++ {
		for j := i + 1; j < len(names); j++ {
			if interaction, ok := lookupInteraction(names[i], names[j]); ok {
				interactions = append(interactions, interaction)
			}
		}
	}
	result := InteractionCheckResult{
		HasInteraction: len(interactions) > 0,
		RiskLevel:      "low",
		Interactions:   interactions,
		Advice:         []string{"keep_medication_record"},
		Safe:           len(interactions) == 0,
	}
	if len(interactions) > 0 {
		result.RiskLevel = highestSeverity(interactions)
		result.Advice = []string{"verify_with_doctor", "avoid_coadministration_until_confirmed"}
	}
	return normalizeInteractionResult(result, result, nil, nil), nil
}

func highRiskSignals() []string {
	return []string{"chest pain", "difficulty breathing", "convulsion", "fainting", "vomit blood"}
}

func collectSignals(symptoms []string, description string) []string {
	matches := map[string]struct{}{}
	for _, symptom := range symptoms {
		value := strings.TrimSpace(strings.ToLower(symptom))
		if value != "" {
			matches[value] = struct{}{}
		}
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
	set := map[string]struct{}{}
	for _, keyword := range keywords {
		set[strings.ToLower(strings.TrimSpace(keyword))] = struct{}{}
	}
	for _, value := range values {
		if _, ok := set[strings.ToLower(strings.TrimSpace(value))]; ok {
			return true
		}
	}
	return false
}

func lookupInteraction(a string, b string) (MedicationInteraction, bool) {
	pair := []string{canonicalMedicineName(a), canonicalMedicineName(b)}
	sort.Strings(pair)
	key := strings.Join(pair, "+")
	m := map[string]MedicationInteraction{
		"aspirin+ibuprofen":                 {Pair: []string{"aspirin", "ibuprofen"}, Severity: "medium", Effect: "increased_gastrointestinal_risk"},
		"cetirizine+chlorpheniramine":       {Pair: []string{"cetirizine", "chlorpheniramine"}, Severity: "medium", Effect: "excessive_drowsiness"},
		"acetaminophen+ibuprofen":           {Pair: []string{"acetaminophen", "ibuprofen"}, Severity: "low", Effect: "duplicate_analgesic_monitoring_needed"},
	}
	item, ok := m[key]
	return item, ok
}

func recommendableMedicines(items []MedicineKnowledge) []MedicineKnowledge {
	result := make([]MedicineKnowledge, 0, len(items))
	for _, item := range items {
		if item.Stock <= 0 || item.IsExpiringSoon || strings.TrimSpace(item.Name) == "" {
			continue
		}
		result = append(result, item)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].IsLowStock != result[j].IsLowStock {
			return !result[i].IsLowStock
		}
		return normalizeKey(result[i].Name) < normalizeKey(result[j].Name)
	})
	return result
}

func selectedMedicines(names []string, inventory []MedicineKnowledge) ([]MedicineKnowledge, []string) {
	items := make([]MedicineKnowledge, 0, len(names))
	flags := make([]string, 0)
	for _, name := range names {
		item, ok := findMedicine(name, inventory)
		if !ok {
			flags = append(flags, "inventory_unavailable")
			continue
		}
		items = append(items, item)
		if item.IsLowStock {
			flags = append(flags, "medicine_low_stock")
		}
		if item.IsExpiringSoon {
			flags = append(flags, "medicine_expiring_soon")
		}
		if len(item.Warnings) > 0 {
			flags = append(flags, "medicine_warning_present")
		}
	}
	return items, uniqueStrings(flags)
}

func findMedicine(name string, inventory []MedicineKnowledge) (MedicineKnowledge, bool) {
	key := normalizeKey(name)
	canonical := canonicalMedicineName(name)
	for _, item := range inventory {
		if normalizeKey(item.Name) == key || canonicalMedicineName(item.Name) == canonical {
			return item, true
		}
	}
	for _, item := range inventory {
		itemKey := normalizeKey(item.Name)
		if strings.Contains(itemKey, key) || strings.Contains(key, itemKey) {
			return item, true
		}
	}
	return MedicineKnowledge{}, false
}

func medicineRAGContext(items []MedicineKnowledge) string {
	lines := make([]string, 0, len(items))
	for _, item := range items {
		parts := []string{
			"name=" + item.Name,
			"spec=" + item.Specification,
			"stock=" + strconv.Itoa(item.Stock),
			"safe_stock=" + strconv.Itoa(item.SafeStock),
		}
		if item.RecommendedDosage != "" {
			parts = append(parts, "dosage="+item.RecommendedDosage)
		}
		if item.RecommendedFrequency != "" {
			parts = append(parts, "frequency="+item.RecommendedFrequency)
		}
		if item.RecommendedDuration != "" {
			parts = append(parts, "duration="+item.RecommendedDuration)
		}
		if item.UsageInstructions != "" {
			parts = append(parts, "usage="+item.UsageInstructions)
		}
		if len(item.Warnings) > 0 {
			parts = append(parts, "warnings="+strings.Join(item.Warnings, "; "))
		}
		if item.IsLowStock {
			parts = append(parts, "low_stock=true")
		}
		lines = append(lines, strings.Join(parts, " | "))
	}
	return strings.Join(lines, "\n")
}

func selectedMedicineRAGContext(names []string, items []MedicineKnowledge, riskFlags []string) string {
	lines := []string{}
	if len(names) > 0 {
		lines = append(lines, "selected="+strings.Join(names, ", "))
	}
	if len(items) > 0 {
		lines = append(lines, medicineRAGContext(items))
	}
	if len(riskFlags) > 0 {
		lines = append(lines, "risk_flags="+strings.Join(riskFlags, ", "))
	}
	return strings.Join(lines, "\n")
}

func normalizeRecommendResult(result RecommendResult, input RecommendInput, inventory []MedicineKnowledge) RecommendResult {
	if len(result.Medicines) == 0 {
		for _, hint := range result.MedicineHints {
			if name := strings.TrimSpace(hint); name != "" {
				result.Medicines = append(result.Medicines, MedicineRecommendation{Name: name})
			}
		}
	}
	final := make([]MedicineRecommendation, 0, len(result.Medicines))
	for _, item := range result.Medicines {
		medicine, ok := findMedicine(item.Name, inventory)
		if !ok {
			result.Contraindications = append(result.Contraindications, fmt.Sprintf("%s not found in local inventory and was removed", strings.TrimSpace(item.Name)))
			result.RiskFlags = append(result.RiskFlags, "inventory_unavailable")
			continue
		}
		if medicine.Stock <= 0 || medicine.IsExpiringSoon || allergyHit(medicine, input.Allergies) {
			result.Contraindications = append(result.Contraindications, fmt.Sprintf("%s was removed by local safety rules", medicine.Name))
			if medicine.IsExpiringSoon {
				result.RiskFlags = append(result.RiskFlags, "medicine_expiring_soon")
			}
			if allergyHit(medicine, input.Allergies) {
				result.RiskFlags = append(result.RiskFlags, "medicine_allergy_warning")
			}
			continue
		}
		item.Name = medicine.Name
		if item.Dosage == "" {
			item.Dosage = medicine.RecommendedDosage
		}
		if item.Frequency == "" {
			item.Frequency = medicine.RecommendedFrequency
		}
		if item.Duration == "" {
			item.Duration = medicine.RecommendedDuration
		}
		item.Caution = joinWithSemicolon(item.Caution, medicine.UsageInstructions, strings.Join(medicine.Warnings, "; "))
		if medicine.IsLowStock {
			item.Caution = joinWithSemicolon(item.Caution, fmt.Sprintf("low stock: %d remaining", medicine.Stock))
			result.RiskFlags = append(result.RiskFlags, "medicine_low_stock")
		}
		if len(medicine.Warnings) > 0 {
			result.RiskFlags = append(result.RiskFlags, "medicine_warning_present")
		}
		if item.Reason == "" {
			item.Reason = "recommended from local inventory RAG context"
		}
		final = append(final, item)
		result.InventoryBasis = append(result.InventoryBasis, medicineBasis(medicine))
	}
	result.Medicines = final
	if isHighRiskTriage(input.TriageLevel, input.Destination) {
		result.RiskFlags = append(result.RiskFlags, "high_risk_triage")
	}
	if len(final) > 1 {
		check, err := defaultAIProvider.InteractionCheck(context.Background(), InteractionCheckInput{Medicines: medicineNames(final)})
		if err == nil && check.HasInteraction {
			result.RiskFlags = append(result.RiskFlags, "local_interaction_detected")
			result.Advice = append(result.Advice, check.Advice...)
			for _, interaction := range check.Interactions {
				result.Contraindications = append(result.Contraindications, fmt.Sprintf("%s: %s", strings.Join(interaction.Pair, " + "), interaction.Effect))
			}
		}
	}
	result.Advice = uniqueStrings(append(result.Advice, result.CarePlan...))
	result.RiskFlags = uniqueStrings(result.RiskFlags)
	result.Contraindications = uniqueStrings(result.Contraindications)
	result.InventoryBasis = uniqueStrings(result.InventoryBasis)
	return syncRecommendLegacy(result)
}

func normalizeInteractionResult(result InteractionCheckResult, local InteractionCheckResult, inventory []MedicineKnowledge, riskFlags []string) InteractionCheckResult {
	result.HasInteraction = result.HasInteraction || local.HasInteraction || len(result.Interactions) > 0 || len(local.Interactions) > 0
	if len(result.Interactions) == 0 {
		result.Interactions = append(result.Interactions, local.Interactions...)
	}
	if result.RiskLevel == "" {
		result.RiskLevel = local.RiskLevel
	}
	if result.RiskLevel == "" && result.HasInteraction {
		result.RiskLevel = highestSeverity(result.Interactions)
	}
	if result.RiskLevel == "" {
		result.RiskLevel = "low"
	}
	result.Advice = uniqueStrings(append(result.Advice, local.Advice...))
	result.RiskFlags = uniqueStrings(append(result.RiskFlags, riskFlags...))
	if len(result.Warnings) == 0 {
		for _, interaction := range result.Interactions {
			result.Warnings = append(result.Warnings, InteractionWarning{
				Title:       strings.Join(interaction.Pair, " + "),
				Severity:    interaction.Severity,
				Description: interaction.Effect,
				Suggestion:  "review the prescription with a doctor before dispensing",
			})
		}
	}
	for _, item := range inventory {
		if len(item.Warnings) > 0 {
			result.Warnings = append(result.Warnings, InteractionWarning{
				Title:       item.Name,
				Severity:    "medium",
				Description: strings.Join(item.Warnings, "; "),
				Suggestion:  "review local medicine warnings before dispensing",
			})
		}
	}
	result.Safe = !result.HasInteraction
	return result
}

func recommendNeedsWebSearch(input RecommendInput, result RecommendResult) bool {
	if isHighRiskTriage(input.TriageLevel, input.Destination) {
		return true
	}
	for _, flag := range result.RiskFlags {
		switch flag {
		case "local_interaction_detected", "medicine_allergy_warning", "medicine_warning_present", "medicine_low_stock", "medicine_expiring_soon", "inventory_unavailable":
			return true
		}
	}
	return false
}

func interactionNeedsWebSearch(local InteractionCheckResult, flags []string) bool {
	return local.HasInteraction || len(flags) > 0
}

func syncRecommendLegacy(result RecommendResult) RecommendResult {
	if result.PlanVersion == "" {
		result.PlanVersion = "ai-rag-v1"
	}
	if len(result.CarePlan) == 0 {
		result.CarePlan = append([]string{}, result.Advice...)
	}
	if len(result.MedicineHints) == 0 {
		result.MedicineHints = medicineNames(result.Medicines)
	}
	if result.FollowUp == "" {
		result.FollowUp = "doctor_review_before_dispense"
	}
	return result
}

func medicineNames(items []MedicineRecommendation) []string {
	names := make([]string, 0, len(items))
	for _, item := range items {
		if name := strings.TrimSpace(item.Name); name != "" {
			names = append(names, name)
		}
	}
	return uniqueStrings(names)
}

func isHighRiskTriage(triageLevel string, destination string) bool {
	triageLevel = strings.ToLower(strings.TrimSpace(triageLevel))
	destination = strings.ToLower(strings.TrimSpace(destination))
	return triageLevel == "priority" || triageLevel == "urgent" || destination == "hospital" || destination == "urgent"
}

func allergyHit(item MedicineKnowledge, allergies []string) bool {
	if len(allergies) == 0 {
		return false
	}
	nameKey := canonicalMedicineName(item.Name)
	warnings := strings.ToLower(strings.Join(item.Warnings, " "))
	for _, allergy := range allergies {
		a := strings.ToLower(strings.TrimSpace(allergy))
		if a == "" {
			continue
		}
		if canonicalMedicineName(allergy) == nameKey || strings.Contains(strings.ToLower(item.Name), a) || strings.Contains(warnings, a) {
			return true
		}
	}
	return false
}

func isSupply(item MedicineKnowledge) bool {
	text := strings.ToLower(strings.Join([]string{item.Name, item.Specification, item.UsageInstructions}, " "))
	return strings.Contains(text, "gauze") || strings.Contains(text, "bandage") || strings.Contains(text, "dressing")
}

func medicineBasis(item MedicineKnowledge) string {
	parts := []string{item.Name, "stock=" + strconv.Itoa(item.Stock)}
	if item.Specification != "" {
		parts = append(parts, "spec="+item.Specification)
	}
	if item.RecommendedDosage != "" {
		parts = append(parts, "dosage="+item.RecommendedDosage)
	}
	if item.RecommendedFrequency != "" {
		parts = append(parts, "frequency="+item.RecommendedFrequency)
	}
	if item.RecommendedDuration != "" {
		parts = append(parts, "duration="+item.RecommendedDuration)
	}
	return strings.Join(parts, " | ")
}

func normalizeKey(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	replacer := strings.NewReplacer(" ", "", "-", "", "_", "", "*", "", "(", "", ")", "", "（", "", "）", "", "片", "", "tablet", "", "tablets", "")
	return replacer.Replace(value)
}

func canonicalMedicineName(name string) string {
	switch normalizeKey(name) {
	case "ibuprofen", "布洛芬", "布洛芬片":
		return "ibuprofen"
	case "ibuprofentablets":
		return "ibuprofen"
	case "acetaminophen", "paracetamol", "对乙酰氨基酚":
		return "acetaminophen"
	case "aspirin", "阿司匹林":
		return "aspirin"
	case "cetirizine", "西替利嗪":
		return "cetirizine"
	case "chlorpheniramine", "扑尔敏", "氯苯那敏":
		return "chlorpheniramine"
	default:
		return normalizeKey(name)
	}
}

func highestSeverity(items []MedicationInteraction) string {
	level := ""
	for _, item := range items {
		if severityScore(item.Severity) > severityScore(level) {
			level = item.Severity
		}
	}
	return level
}

func severityScore(level string) int {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "high", "critical", "severe":
		return 3
	case "medium", "warning":
		return 2
	case "low":
		return 1
	default:
		return 0
	}
}

func joinWithSemicolon(parts ...string) string {
	items := make([]string, 0, len(parts))
	for _, part := range parts {
		if value := strings.TrimSpace(part); value != "" {
			items = append(items, value)
		}
	}
	return strings.Join(uniqueStrings(items), "; ")
}

func uniqueStrings(items []string) []string {
	seen := map[string]struct{}{}
	result := make([]string, 0, len(items))
	for _, item := range items {
		value := strings.TrimSpace(item)
		if value == "" {
			continue
		}
		key := strings.ToLower(value)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, value)
	}
	return result
}
