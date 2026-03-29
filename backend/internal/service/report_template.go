package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jiaxiang-medical-assistant/backend/internal/repository"
	"github.com/xuri/excelize/v2"
)

type ReportTemplateService struct {
	tplRepo   repository.ReportTemplateRepository
	schedRepo repository.ReportScheduleRepository
	reportSvc *ReportService
}

func NewReportTemplateService(
	tplRepo repository.ReportTemplateRepository,
	schedRepo repository.ReportScheduleRepository,
	reportSvc *ReportService,
) *ReportTemplateService {
	return &ReportTemplateService{tplRepo: tplRepo, schedRepo: schedRepo, reportSvc: reportSvc}
}

// ---- Template CRUD ----

type CreateTemplateInput struct {
	Name    string   `json:"name"`
	Period  string   `json:"period"`
	Columns []string `json:"columns"`
	Title   string   `json:"title"`
}

func (s *ReportTemplateService) CreateTemplate(ctx context.Context, input CreateTemplateInput) (repository.ReportTemplate, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return repository.ReportTemplate{}, ErrInvalidInput
	}
	period := strings.ToLower(strings.TrimSpace(input.Period))
	if period != "daily" && period != "weekly" && period != "monthly" {
		return repository.ReportTemplate{}, ErrInvalidInput
	}
	columns := filterValidColumns(input.Columns)
	if len(columns) == 0 {
		columns = allColumnKeys()
	}

	now := time.Now().UTC()
	tpl := repository.ReportTemplate{
		ID:        uuid.NewString(),
		Name:      name,
		Period:    period,
		Columns:   columns,
		Title:     strings.TrimSpace(input.Title),
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.tplRepo.Create(ctx, tpl); err != nil {
		return repository.ReportTemplate{}, err
	}
	return tpl, nil
}

func (s *ReportTemplateService) GetTemplate(ctx context.Context, id string) (repository.ReportTemplate, error) {
	return s.tplRepo.Get(ctx, id)
}

func (s *ReportTemplateService) ListTemplates(ctx context.Context) ([]repository.ReportTemplate, error) {
	return s.tplRepo.List(ctx)
}

type UpdateTemplateInput struct {
	Name    *string  `json:"name"`
	Columns []string `json:"columns"`
	Title   *string  `json:"title"`
}

func (s *ReportTemplateService) UpdateTemplate(ctx context.Context, id string, input UpdateTemplateInput) (repository.ReportTemplate, error) {
	tpl, err := s.tplRepo.Get(ctx, id)
	if err != nil {
		return repository.ReportTemplate{}, err
	}

	if input.Name != nil {
		name := strings.TrimSpace(*input.Name)
		if name == "" {
			return repository.ReportTemplate{}, ErrInvalidInput
		}
		tpl.Name = name
	}
	if input.Columns != nil {
		cols := filterValidColumns(input.Columns)
		if len(cols) == 0 {
			return repository.ReportTemplate{}, ErrInvalidInput
		}
		tpl.Columns = cols
	}
	if input.Title != nil {
		tpl.Title = strings.TrimSpace(*input.Title)
	}

	tpl.UpdatedAt = time.Now().UTC()
	if err := s.tplRepo.Update(ctx, tpl); err != nil {
		return repository.ReportTemplate{}, err
	}
	return tpl, nil
}

func (s *ReportTemplateService) DeleteTemplate(ctx context.Context, id string) error {
	if _, err := s.tplRepo.Get(ctx, id); err != nil {
		return err
	}
	if err := repositoryDeleteSchedulesByTemplate(ctx, s.schedRepo, id); err != nil {
		return err
	}
	return s.tplRepo.Delete(ctx, id)
}

// ---- Export with template ----

func (s *ReportTemplateService) ExportWithTemplate(ctx context.Context, templateID string) (*ExcelExportResult, error) {
	tpl, err := s.tplRepo.Get(ctx, templateID)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	var start, end time.Time
	var dateLabel string

	switch tpl.Period {
	case "daily":
		start = dayStart(now)
		end = start.Add(24 * time.Hour)
		dateLabel = start.Format("2006-01-02")
	case "weekly":
		start = weekStart(now)
		end = start.AddDate(0, 0, 7)
		dateLabel = start.Format("2006-01-02") + "_" + end.AddDate(0, 0, -1).Format("2006-01-02")
	case "monthly":
		start = monthStart(now)
		end = start.AddDate(0, 1, 0)
		dateLabel = start.Format("2006-01")
	default:
		return nil, ErrInvalidInput
	}

	visits, err := s.reportSvc.listVisitsInRange(ctx, start, end)
	if err != nil {
		return nil, err
	}

	report, err := s.reportSvc.buildVisitPeriodReport(ctx, tpl.Period, start, end)
	if err != nil {
		return nil, err
	}

	f := excelize.NewFile()
	writeVisitSheetWithColumns(f, "就诊明细", visits, tpl.Columns)

	writeSummarySheetWithTitle(f, "统计概览", report, dateLabel, tpl.Title)

	filename := fmt.Sprintf("%s_%s_%s.xlsx", tpl.Name, tpl.Period, dateLabel)

	return &ExcelExportResult{File: f, Filename: filename}, nil
}

// ---- Schedule CRUD ----

type CreateScheduleInput struct {
	TemplateID string `json:"template_id"`
	CronExpr   string `json:"cron_expr"`
}

func (s *ReportTemplateService) CreateSchedule(ctx context.Context, input CreateScheduleInput) (repository.ReportSchedule, error) {
	templateID := strings.TrimSpace(input.TemplateID)
	cronExpr := strings.TrimSpace(input.CronExpr)
	if templateID == "" || cronExpr == "" {
		return repository.ReportSchedule{}, ErrInvalidInput
	}

	tpl, err := s.tplRepo.Get(ctx, templateID)
	if err != nil {
		return repository.ReportSchedule{}, err
	}

	nextRun := calcNextRunForPeriod(tpl.Period, cronExpr, time.Now().UTC())
	now := time.Now().UTC()

	sched := repository.ReportSchedule{
		ID:         uuid.NewString(),
		TemplateID: templateID,
		CronExpr:   cronExpr,
		Enabled:    true,
		NextRunAt:  nextRun,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := s.schedRepo.Create(ctx, sched); err != nil {
		return repository.ReportSchedule{}, err
	}
	return sched, nil
}

func (s *ReportTemplateService) ListSchedules(ctx context.Context) ([]repository.ReportSchedule, error) {
	return s.schedRepo.List(ctx)
}

type UpdateScheduleInput struct {
	CronExpr *string `json:"cron_expr"`
	Enabled  *bool   `json:"enabled"`
}

func (s *ReportTemplateService) UpdateSchedule(ctx context.Context, id string, input UpdateScheduleInput) (repository.ReportSchedule, error) {
	sched, err := s.schedRepo.Get(ctx, id)
	if err != nil {
		return repository.ReportSchedule{}, err
	}

	if input.CronExpr != nil {
		expr := strings.TrimSpace(*input.CronExpr)
		if expr == "" {
			return repository.ReportSchedule{}, ErrInvalidInput
		}
		tpl, err := s.tplRepo.Get(ctx, sched.TemplateID)
		if err != nil {
			return repository.ReportSchedule{}, err
		}
		sched.CronExpr = expr
		sched.NextRunAt = calcNextRunForPeriod(tpl.Period, expr, time.Now().UTC())
	}
	if input.Enabled != nil {
		wasEnabled := sched.Enabled
		sched.Enabled = *input.Enabled
		if !sched.Enabled {
			sched.NextRunAt = nil
		} else if !wasEnabled || sched.NextRunAt == nil {
			tpl, err := s.tplRepo.Get(ctx, sched.TemplateID)
			if err != nil {
				return repository.ReportSchedule{}, err
			}
			sched.NextRunAt = calcNextRunForPeriod(tpl.Period, sched.CronExpr, time.Now().UTC())
		}
	}

	sched.UpdatedAt = time.Now().UTC()
	if err := s.schedRepo.Update(ctx, sched); err != nil {
		return repository.ReportSchedule{}, err
	}
	return sched, nil
}

func (s *ReportTemplateService) DeleteSchedule(ctx context.Context, id string) error {
	return s.schedRepo.Delete(ctx, id)
}

// ---- Column definitions ----

type columnDef struct {
	Key    string
	Header string
	Width  float64
	Extract func(v repository.Visit, idx int) string
}

var columnRegistry = []columnDef{
	{"index", "序号", 6, func(_ repository.Visit, idx int) string { return fmt.Sprintf("%d", idx+1) }},
	{"student_name", "学生姓名", 12, func(v repository.Visit, _ int) string { return v.StudentName }},
	{"class_name", "班级", 12, func(v repository.Visit, _ int) string { return v.ClassName }},
	{"symptoms", "症状", 22, func(v repository.Visit, _ int) string { return strings.Join(v.Symptoms, "、") }},
	{"description", "描述", 28, func(v repository.Visit, _ int) string { return v.Description }},
	{"diagnosis", "诊断", 22, func(v repository.Visit, _ int) string { return v.Diagnosis }},
	{"prescription", "处方", 22, func(v repository.Visit, _ int) string { return strings.Join(v.Prescription, "、") }},
	{"destination", "去向", 10, func(v repository.Visit, _ int) string {
		if label, ok := destMap[v.Destination]; ok {
			return label
		}
		return v.Destination
	}},
	{"follow_up_at", "复诊时间", 18, func(v repository.Visit, _ int) string {
		if v.FollowUpAt != nil {
			return v.FollowUpAt.Format("2006-01-02 15:04")
		}
		return ""
	}},
	{"follow_up_note", "复诊备注", 22, func(v repository.Visit, _ int) string {
		if v.FollowUpNote != nil {
			return *v.FollowUpNote
		}
		return ""
	}},
	{"created_at", "就诊时间", 20, func(v repository.Visit, _ int) string { return v.CreatedAt.Format("2006-01-02 15:04:05") }},
}

func allColumnKeys() []string {
	keys := make([]string, len(columnRegistry))
	for i, col := range columnRegistry {
		keys[i] = col.Key
	}
	return keys
}

var validColumnSet = func() map[string]bool {
	m := make(map[string]bool, len(columnRegistry))
	for _, col := range columnRegistry {
		m[col.Key] = true
	}
	return m
}()

func filterValidColumns(cols []string) []string {
	result := make([]string, 0, len(cols))
	for _, c := range cols {
		c = strings.TrimSpace(c)
		if validColumnSet[c] {
			result = append(result, c)
		}
	}
	return result
}

func selectedColumns(keys []string) []columnDef {
	colMap := make(map[string]columnDef, len(columnRegistry))
	for _, col := range columnRegistry {
		colMap[col.Key] = col
	}
	result := make([]columnDef, 0, len(keys))
	for _, key := range keys {
		if col, ok := colMap[key]; ok {
			result = append(result, col)
		}
	}
	return result
}

func writeVisitSheetWithColumns(f *excelize.File, sheet string, visits []repository.Visit, columns []string) {
	cols := selectedColumns(columns)
	if len(cols) == 0 {
		cols = selectedColumns(allColumnKeys())
	}

	if sheet != "Sheet1" {
		idx, _ := f.NewSheet(sheet)
		f.SetActiveSheet(idx)
	}

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 11, Color: "#FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#4472C4"}},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Border:    thinBorder,
	})
	evenRowStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 10},
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#D9E2F3"}},
		Alignment: &excelize.Alignment{Vertical: "center", WrapText: true},
		Border:    thinBorder,
	})
	oddRowStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 10},
		Alignment: &excelize.Alignment{Vertical: "center", WrapText: true},
		Border:    thinBorder,
	})

	_ = f.SetRowHeight(sheet, 1, 28)
	for i, col := range cols {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellValue(sheet, cell, col.Header)
		_ = f.SetCellStyle(sheet, cell, cell, headerStyle)
	}

	for i, v := range visits {
		row := i + 2
		isEven := i%2 == 0
		style := oddRowStyle
		if isEven {
			style = evenRowStyle
		}
		for j, col := range cols {
			cell, _ := excelize.CoordinatesToCellName(j+1, row)
			_ = f.SetCellValue(sheet, cell, col.Extract(v, i))
			_ = f.SetCellStyle(sheet, cell, cell, style)
		}
	}

	for i, col := range cols {
		colName, _ := excelize.ColumnNumberToName(i + 1)
		_ = f.SetColWidth(sheet, colName, colName, col.Width)
	}

	_ = f.SetPanes(sheet, &excelize.Panes{
		Freeze:      true,
		Split:       false,
		XSplit:      0,
		YSplit:      1,
		TopLeftCell: "A2",
		ActivePane:  "bottomLeft",
	})

	if len(cols) > 0 {
		lastCol, _ := excelize.ColumnNumberToName(len(cols))
		lastRow := len(visits) + 1
		_ = f.AutoFilter(sheet, fmt.Sprintf("A1:%s%d", lastCol, lastRow), nil)
	}
}

// ---- Simple cron next-run calculator ----
// Supports: "HH:MM" (daily), "dow HH:MM" (weekly, dow=1-7), "day HH:MM" (monthly, day=1-28)

func calcNextRun(cronExpr string, now time.Time) *time.Time {
	parts := strings.Fields(cronExpr)

	switch len(parts) {
	case 1:
		// "HH:MM" — daily
		t := parseHHMM(parts[0], now)
		if !t.After(now) {
			t = t.Add(24 * time.Hour)
		}
		return &t
	case 2:
		// "DAY HH:MM" — interpret DAY as day-of-month (1-28) or day-of-week (mon-sun)
		hhmm := parseHHMM(parts[1], now)
		var dayNum int
		if _, err := fmt.Sscanf(parts[0], "%d", &dayNum); err == nil {
			if dayNum >= 1 && dayNum <= 28 {
				// Monthly
				t := time.Date(now.Year(), now.Month(), dayNum, hhmm.Hour(), hhmm.Minute(), 0, 0, time.UTC)
				if !t.After(now) {
					t = t.AddDate(0, 1, 0)
				}
				return &t
			}
		}
	}

	// Fallback: next hour
	t := now.Add(time.Hour).Truncate(time.Hour)
	return &t
}

func calcNextRunForPeriod(period string, cronExpr string, now time.Time) *time.Time {
	parts := strings.Fields(strings.TrimSpace(cronExpr))
	if len(parts) != 1 {
		return calcNextRun(cronExpr, now)
	}

	hhmm := parseHHMM(parts[0], now)
	switch period {
	case "weekly":
		return nextWeekEndAt(hhmm.Hour(), hhmm.Minute(), now)
	case "monthly":
		return nextMonthEndAt(hhmm.Hour(), hhmm.Minute(), now)
	default:
		return calcNextRun(cronExpr, now)
	}
}

func nextWeekEndAt(hour int, minute int, now time.Time) *time.Time {
	start := weekStart(now)
	target := time.Date(start.Year(), start.Month(), start.Day(), hour, minute, 0, 0, time.UTC).AddDate(0, 0, 6)
	if !target.After(now.UTC()) {
		target = target.AddDate(0, 0, 7)
	}
	return &target
}

func nextMonthEndAt(hour int, minute int, now time.Time) *time.Time {
	t := now.UTC()
	firstOfNextMonth := time.Date(t.Year(), t.Month(), 1, hour, minute, 0, 0, time.UTC).AddDate(0, 1, 0)
	target := firstOfNextMonth.AddDate(0, 0, -1)
	if !target.After(t) {
		target = firstOfNextMonth.AddDate(0, 1, -1)
	}
	return &target
}

func repositoryDeleteSchedulesByTemplate(ctx context.Context, schedRepo repository.ReportScheduleRepository, templateID string) error {
	schedules, err := schedRepo.List(ctx)
	if err != nil {
		return err
	}
	for _, sched := range schedules {
		if sched.TemplateID != templateID {
			continue
		}
		if err := schedRepo.Delete(ctx, sched.ID); err != nil {
			return err
		}
	}
	return nil
}

func parseHHMM(s string, now time.Time) time.Time {
	var h, m int
	if _, err := fmt.Sscanf(s, "%d:%d", &h, &m); err != nil {
		return now
	}
	return time.Date(now.Year(), now.Month(), now.Day(), h, m, 0, 0, time.UTC)
}
