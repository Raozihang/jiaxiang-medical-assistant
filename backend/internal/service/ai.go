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
	Temperature float64  `json:"temperature"`
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
			Warnings:             localizeTextList(uniqueStrings(row.Warnings)),
			RecommendedDosage:    localizeMedicalText(strings.TrimSpace(row.RecommendedDosage)),
			RecommendedFrequency: localizeMedicalText(strings.TrimSpace(row.RecommendedFrequency)),
			RecommendedDuration:  localizeMedicalText(strings.TrimSpace(row.RecommendedDuration)),
			UsageInstructions:    localizeMedicalText(strings.TrimSpace(row.UsageInstructions)),
			IsLowStock:           row.IsLowStock,
			IsExpiringSoon:       row.IsExpiringSoon,
		})
	}
	return items
}

func (p *ruleBasedAIProvider) Analyze(_ context.Context, input AnalyzeInput) (AnalyzeResult, error) {
	matched := collectSignals(input.Symptoms, input.Description)
	risk, confidence := estimateRisk(input.Temperature, matched)
	causes := []string{"普通感冒或轻度上呼吸道不适"}
	actions := []string{"注意补水", "适当休息", "如症状加重及时复诊"}
	if containsAny(matched, highRiskSignals()) {
		causes = []string{"可能存在急性风险情况", "需要立即进一步评估"}
		actions = []string{"立即通知监护人", "尽快转诊医院"}
	} else if risk == "medium" {
		causes = []string{"可能为上呼吸道感染", "可能存在胃肠道不适"}
		actions = []string{"留观观察", "每30分钟复测体温或生命体征"}
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
		Reason:           "暂未发现高风险信号，可先回班并继续观察。",
		ReviewInMinutes:  120,
		SuggestedActions: []string{"返回班级后继续观察", "如再次不适立即报告老师或校医"},
	}
	switch analysis.RiskLevel {
	case "high":
		result.TriageLevel = "urgent"
		result.Destination = "hospital"
		result.Reason = "已识别高风险信号，需要立即送医进一步处置。"
		result.ReviewInMinutes = 10
		result.SuggestedActions = []string{"立即联系监护人", "准备转运或送医"}
	case "medium":
		result.TriageLevel = "priority"
		result.Destination = "observation"
		result.Reason = "当前存在一定风险，建议先留观并复评。"
		result.ReviewInMinutes = 30
		result.SuggestedActions = []string{"安排留观", "复测体温及其他生命体征"}
	}
	return result, nil
}

func (p *ruleBasedAIProvider) Recommend(_ context.Context, input RecommendInput) (RecommendResult, error) {
	result := RecommendResult{
		PlanVersion: "mock-rag-v1",
		Advice:      []string{"记录症状变化", "注意补水"},
		FollowUp:    "2小时后复评",
	}
	if isHighRiskTriage(input.TriageLevel, input.Destination) {
		result.Advice = []string{"先稳定学生当前状况", "尽快与医院完成交接"}
		result.FollowUp = "持续跟进医院处置反馈"
		return syncRecommendLegacy(result), nil
	}
	if strings.Contains(strings.ToLower(strings.TrimSpace(input.Diagnosis)), "allergy") {
		result.Advice = append(result.Advice, "避免接触已知过敏原")
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
			Reason:    "基于本地库存、推荐用法和安全信息筛选出的可用药品",
			Caution:   strings.Join(item.Warnings, "; "),
		})
		break
	}
	if len(result.Medicines) == 0 {
		result.Contraindications = []string{"本地库存中暂无符合当前条件的安全用药选项"}
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
		Advice:         []string{"请保留完整用药记录"},
		Safe:           len(interactions) == 0,
	}
	if len(interactions) > 0 {
		result.RiskLevel = highestSeverity(interactions)
		result.Advice = []string{"请由医生进一步复核", "在确认前避免联合用药"}
	}
	return normalizeInteractionResult(result, result, nil, nil), nil
}

func highRiskSignals() []string {
	return []string{"chest pain", "difficulty breathing", "convulsion", "fainting", "vomit blood", "胸痛", "呼吸困难", "抽搐", "晕厥", "吐血"}
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
		"aspirin+ibuprofen":           {Pair: []string{"aspirin", "ibuprofen"}, Severity: "medium", Effect: "胃肠道不良反应风险增加"},
		"cetirizine+chlorpheniramine": {Pair: []string{"cetirizine", "chlorpheniramine"}, Severity: "medium", Effect: "嗜睡风险增加"},
		"acetaminophen+ibuprofen":     {Pair: []string{"acetaminophen", "ibuprofen"}, Severity: "low", Effect: "存在重复止痛用药风险，需加强观察"},
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
			"药品=" + item.Name,
			"规格=" + item.Specification,
			"库存=" + strconv.Itoa(item.Stock),
			"安全库存=" + strconv.Itoa(item.SafeStock),
		}
		if item.RecommendedDosage != "" {
			parts = append(parts, "推荐剂量="+item.RecommendedDosage)
		}
		if item.RecommendedFrequency != "" {
			parts = append(parts, "推荐频次="+item.RecommendedFrequency)
		}
		if item.RecommendedDuration != "" {
			parts = append(parts, "推荐疗程="+item.RecommendedDuration)
		}
		if item.UsageInstructions != "" {
			parts = append(parts, "用法="+item.UsageInstructions)
		}
		if len(item.Warnings) > 0 {
			parts = append(parts, "警示="+strings.Join(item.Warnings, "; "))
		}
		if item.IsLowStock {
			parts = append(parts, "低库存=是")
		}
		lines = append(lines, strings.Join(parts, " | "))
	}
	return strings.Join(lines, "\n")
}

func selectedMedicineRAGContext(names []string, items []MedicineKnowledge, riskFlags []string) string {
	lines := []string{}
	if len(names) > 0 {
		lines = append(lines, "已选药品="+strings.Join(names, "，"))
	}
	if len(items) > 0 {
		lines = append(lines, medicineRAGContext(items))
	}
	if len(riskFlags) > 0 {
		lines = append(lines, "风险标记="+strings.Join(riskFlags, ", "))
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
			result.Contraindications = append(result.Contraindications, fmt.Sprintf("%s 不在本地库存中，已移除。", strings.TrimSpace(item.Name)))
			result.RiskFlags = append(result.RiskFlags, "inventory_unavailable")
			continue
		}
		if medicine.Stock <= 0 || medicine.IsExpiringSoon || allergyHit(medicine, input.Allergies) {
			result.Contraindications = append(result.Contraindications, fmt.Sprintf("%s 因本地安全规则限制已移除。", medicine.Name))
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
		item.Dosage = localizeMedicalText(item.Dosage)
		item.Frequency = localizeMedicalText(item.Frequency)
		item.Duration = localizeMedicalText(item.Duration)
		item.Reason = localizeMedicalText(item.Reason)
		item.Caution = joinWithSemicolon(item.Caution, medicine.UsageInstructions, strings.Join(medicine.Warnings, "; "))
		if medicine.IsLowStock {
			item.Caution = joinWithSemicolon(item.Caution, fmt.Sprintf("库存偏低：剩余 %d", medicine.Stock))
			result.RiskFlags = append(result.RiskFlags, "medicine_low_stock")
		}
		if len(medicine.Warnings) > 0 {
			result.RiskFlags = append(result.RiskFlags, "medicine_warning_present")
		}
		if item.Reason == "" {
			item.Reason = "基于本地库存、推荐用法与RAG知识综合推荐"
		}
		item.Caution = localizeMedicalText(item.Caution)
		item.Reason = localizeMedicalText(item.Reason)
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
				result.Contraindications = append(result.Contraindications, fmt.Sprintf("%s：%s", strings.Join(interaction.Pair, " + "), interaction.Effect))
			}
		}
	}
	result.Advice = uniqueStrings(append(result.Advice, result.CarePlan...))
	result.Advice = localizeTextList(result.Advice)
	result.RiskFlags = uniqueStrings(result.RiskFlags)
	result.Contraindications = localizeTextList(uniqueStrings(result.Contraindications))
	result.InventoryBasis = localizeTextList(uniqueStrings(result.InventoryBasis))
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
	result.Advice = localizeTextList(result.Advice)
	result.RiskFlags = uniqueStrings(append(result.RiskFlags, riskFlags...))
	if len(result.Warnings) == 0 {
		for _, interaction := range result.Interactions {
			result.Warnings = append(result.Warnings, InteractionWarning{
				Title:       strings.Join(interaction.Pair, " + "),
				Severity:    interaction.Severity,
				Description: interaction.Effect,
				Suggestion:  "发药前请由医生复核当前处方。",
			})
		}
	}
	for _, item := range inventory {
		if len(item.Warnings) > 0 {
			result.Warnings = append(result.Warnings, InteractionWarning{
				Title:       item.Name,
				Severity:    "medium",
				Description: strings.Join(item.Warnings, "; "),
				Suggestion:  "发药前请核对该药品的本地警示信息。",
			})
		}
	}
	for i := range result.Interactions {
		result.Interactions[i].Effect = localizeMedicalText(result.Interactions[i].Effect)
	}
	for i := range result.Warnings {
		result.Warnings[i].Title = localizeMedicalText(result.Warnings[i].Title)
		result.Warnings[i].Description = localizeMedicalText(result.Warnings[i].Description)
		result.Warnings[i].Suggestion = localizeMedicalText(result.Warnings[i].Suggestion)
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
		result.FollowUp = "发药前请由医生复核"
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
	return strings.Contains(text, "gauze") || strings.Contains(text, "bandage") || strings.Contains(text, "dressing") ||
		strings.Contains(text, "纱布") || strings.Contains(text, "绷带") || strings.Contains(text, "敷料")
}

func medicineBasis(item MedicineKnowledge) string {
	parts := []string{item.Name, "库存=" + strconv.Itoa(item.Stock)}
	if item.Specification != "" {
		parts = append(parts, "规格="+item.Specification)
	}
	if item.RecommendedDosage != "" {
		parts = append(parts, "推荐剂量="+item.RecommendedDosage)
	}
	if item.RecommendedFrequency != "" {
		parts = append(parts, "推荐频次="+item.RecommendedFrequency)
	}
	if item.RecommendedDuration != "" {
		parts = append(parts, "推荐疗程="+item.RecommendedDuration)
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

func localizeTextList(items []string) []string {
	result := make([]string, 0, len(items))
	for _, item := range items {
		if value := localizeMedicalText(item); strings.TrimSpace(value) != "" {
			result = append(result, value)
		}
	}
	return uniqueStrings(result)
}

func localizeMedicalText(value string) string {
	text := strings.TrimSpace(value)
	if text == "" {
		return ""
	}

	replacements := map[string]string{
		"Do not use in students allergic to ibuprofen":                                           "对布洛芬过敏者禁用",
		"Do not administer Ibuprofen solely for cough relief as it is clinically inappropriate.": "不建议仅因咳嗽症状使用布洛芬，该用药场景并不适宜。",
		"take after meals; stop if stomach discomfort occurs":                                    "建议饭后服用；如出现胃部不适应立即停用",
		"every 6-8 hours as needed":                                                              "必要时每6到8小时一次",
		"up to 3 days":                                                                           "最多连续使用3天",
		"external use only":                                                                      "仅限外用",
		"review the prescription with a doctor before dispensing":                                "发药前请由医生复核当前处方。",
		"review local medicine warnings before dispensing":                                       "发药前请核对该药品的本地警示信息。",
		"increased gastrointestinal risk":                                                        "胃肠道不良反应风险增加",
		"increased gastrointestinal bleeding risk":                                               "胃肠道出血风险增加",
		"excessive drowsiness":                                                                   "嗜睡风险增加",
		"duplicate analgesic monitoring needed":                                                  "存在重复止痛用药风险，需加强观察",
	}
	if localized, ok := replacements[text]; ok {
		return localized
	}

	text = strings.ReplaceAll(text, "Do not use in students allergic to ibuprofen", "对布洛芬过敏者禁用")
	text = strings.ReplaceAll(text, "take after meals", "建议饭后服用")
	text = strings.ReplaceAll(text, "stop if stomach discomfort occurs", "如出现胃部不适应立即停用")
	text = strings.ReplaceAll(text, "every 6-8 hours as needed", "必要时每6到8小时一次")
	text = strings.ReplaceAll(text, "up to 3 days", "最多连续使用3天")
	text = strings.ReplaceAll(text, "external use only", "仅限外用")
	text = strings.ReplaceAll(text, "clinically inappropriate", "不适宜")
	text = strings.ReplaceAll(text, "inappropriate", "不适宜")
	text = strings.ReplaceAll(text, "solely for cough relief", "仅用于缓解咳嗽")
	text = strings.ReplaceAll(text, "Do not administer", "不建议使用")
	text = strings.ReplaceAll(text, "Ibuprofen", "布洛芬")
	text = strings.ReplaceAll(text, "ibuprofen", "布洛芬")
	text = strings.ReplaceAll(text, "Gauze", "纱布")
	text = strings.ReplaceAll(text, "Tablets", "片")
	text = strings.ReplaceAll(text, "tablets", "片")
	text = strings.ReplaceAll(text, " + ", " + ")
	text = strings.TrimSpace(strings.TrimSuffix(text, "."))

	return text
}
