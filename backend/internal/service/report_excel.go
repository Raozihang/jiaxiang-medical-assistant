package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jiaxiang-medical-assistant/backend/internal/repository"
	"github.com/xuri/excelize/v2"
)

type ExcelExportResult struct {
	File     *excelize.File
	Filename string
}

func (s *ReportService) ExportDaily(ctx context.Context) (*ExcelExportResult, error) {
	now := time.Now().UTC()
	start := dayStart(now)
	end := start.Add(24 * time.Hour)
	dateStr := start.Format("2006-01-02")

	visits, err := s.listVisitsInRange(ctx, start, end)
	if err != nil {
		return nil, err
	}

	report, err := s.buildVisitPeriodReport(ctx, "daily", start, end)
	if err != nil {
		return nil, err
	}

	f := excelize.NewFile()
	writeVisitSheet(f, "就诊明细", visits)
	writeSummarySheet(f, "统计概览", report, dateStr)

	return &ExcelExportResult{
		File:     f,
		Filename: fmt.Sprintf("日报_%s.xlsx", dateStr),
	}, nil
}

func (s *ReportService) ExportWeekly(ctx context.Context) (*ExcelExportResult, error) {
	now := time.Now().UTC()
	start := weekStart(now)
	end := start.AddDate(0, 0, 7)
	label := start.Format("2006-01-02") + "_" + end.AddDate(0, 0, -1).Format("2006-01-02")

	visits, err := s.listVisitsInRange(ctx, start, end)
	if err != nil {
		return nil, err
	}

	report, err := s.buildVisitPeriodReport(ctx, "weekly", start, end)
	if err != nil {
		return nil, err
	}

	f := excelize.NewFile()
	writeVisitSheet(f, "就诊明细", visits)
	writeSummarySheet(f, "统计概览", report, label)

	return &ExcelExportResult{
		File:     f,
		Filename: fmt.Sprintf("周报_%s.xlsx", label),
	}, nil
}

func (s *ReportService) ExportMonthly(ctx context.Context) (*ExcelExportResult, error) {
	now := time.Now().UTC()
	start := monthStart(now)
	end := start.AddDate(0, 1, 0)
	label := start.Format("2006-01")

	visits, err := s.listVisitsInRange(ctx, start, end)
	if err != nil {
		return nil, err
	}

	report, err := s.buildVisitPeriodReport(ctx, "monthly", start, end)
	if err != nil {
		return nil, err
	}

	f := excelize.NewFile()
	writeVisitSheet(f, "就诊明细", visits)
	writeSummarySheet(f, "统计概览", report, label)

	return &ExcelExportResult{
		File:     f,
		Filename: fmt.Sprintf("月报_%s.xlsx", label),
	}, nil
}

func (s *ReportService) listVisitsInRange(ctx context.Context, start, end time.Time) ([]repository.Visit, error) {
	all, err := s.listAllVisits(ctx)
	if err != nil {
		return nil, err
	}

	filtered := make([]repository.Visit, 0)
	for _, v := range all {
		if !v.CreatedAt.Before(start) && v.CreatedAt.Before(end) {
			filtered = append(filtered, v)
		}
	}
	return filtered, nil
}

// ---- Shared styles ----

var thinBorder = []excelize.Border{
	{Type: "left", Style: 1, Color: "#B4C6E7"},
	{Type: "right", Style: 1, Color: "#B4C6E7"},
	{Type: "top", Style: 1, Color: "#B4C6E7"},
	{Type: "bottom", Style: 1, Color: "#B4C6E7"},
}

// ---- Visit detail sheet ----

var visitHeaders = []string{
	"序号", "学生姓名", "班级", "症状", "描述", "诊断",
	"处方", "去向", "复诊时间", "复诊备注", "就诊时间",
}

var visitColWidths = []float64{6, 12, 12, 22, 28, 22, 22, 10, 18, 22, 20}

var destMap = map[string]string{
	"classroom":   "返回教室",
	"dormitory":   "返回寝室",
	"observation": "留观",
	"hospital":    "转院",
	"home":        "离校回家",
}

func writeVisitSheet(f *excelize.File, sheet string, visits []repository.Visit) {
	if sheet != "Sheet1" {
		idx, _ := f.NewSheet(sheet)
		f.SetActiveSheet(idx)
	}

	// --- Styles ---
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
	numCenterStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 10},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border:    thinBorder,
	})
	numCenterEvenStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 10},
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#D9E2F3"}},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border:    thinBorder,
	})

	// --- Header row ---
	_ = f.SetRowHeight(sheet, 1, 28)
	for i, h := range visitHeaders {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellValue(sheet, cell, h)
		_ = f.SetCellStyle(sheet, cell, cell, headerStyle)
	}

	// --- Data rows ---
	for i, v := range visits {
		row := i + 2
		isEven := i%2 == 0

		dest := v.Destination
		if label, ok := destMap[dest]; ok {
			dest = label
		}
		followUp := ""
		if v.FollowUpAt != nil {
			followUp = v.FollowUpAt.Format("2006-01-02 15:04")
		}
		followUpNote := ""
		if v.FollowUpNote != nil {
			followUpNote = *v.FollowUpNote
		}

		vals := []any{
			i + 1,
			v.StudentName,
			v.ClassName,
			strings.Join(v.Symptoms, "、"),
			v.Description,
			v.Diagnosis,
			strings.Join(v.Prescription, "、"),
			dest,
			followUp,
			followUpNote,
			v.CreatedAt.Format("2006-01-02 15:04:05"),
		}

		for j, val := range vals {
			cell, _ := excelize.CoordinatesToCellName(j+1, row)
			_ = f.SetCellValue(sheet, cell, val)

			// First column (序号) uses center style; others use row style
			if j == 0 {
				if isEven {
					_ = f.SetCellStyle(sheet, cell, cell, numCenterEvenStyle)
				} else {
					_ = f.SetCellStyle(sheet, cell, cell, numCenterStyle)
				}
			} else {
				if isEven {
					_ = f.SetCellStyle(sheet, cell, cell, evenRowStyle)
				} else {
					_ = f.SetCellStyle(sheet, cell, cell, oddRowStyle)
				}
			}
		}
	}

	// --- Column widths ---
	for i, w := range visitColWidths {
		col, _ := excelize.ColumnNumberToName(i + 1)
		_ = f.SetColWidth(sheet, col, col, w)
	}

	// --- Freeze header row ---
	_ = f.SetPanes(sheet, &excelize.Panes{
		Freeze:      true,
		Split:       false,
		XSplit:      0,
		YSplit:      1,
		TopLeftCell: "A2",
		ActivePane:  "bottomLeft",
	})

	// --- Auto filter ---
	lastCol, _ := excelize.ColumnNumberToName(len(visitHeaders))
	lastRow := len(visits) + 1
	_ = f.AutoFilter(sheet, fmt.Sprintf("A1:%s%d", lastCol, lastRow), nil)
}

// ---- Summary sheet ----

func writeSummarySheet(f *excelize.File, sheet string, report VisitPeriodReport, label string) {
	writeSummarySheetWithTitle(f, sheet, report, label, "")
}

func writeSummarySheetWithTitle(f *excelize.File, sheet string, report VisitPeriodReport, label string, customTitle string) {
	_, _ = f.NewSheet(sheet)

	// --- Styles ---
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 16, Color: "#1F4E79"},
		Alignment: &excelize.Alignment{Vertical: "center"},
	})
	sectionStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 12, Color: "#2E75B6"},
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#D6E4F0"}},
		Alignment: &excelize.Alignment{Vertical: "center"},
		Border: []excelize.Border{
			{Type: "bottom", Style: 2, Color: "#2E75B6"},
		},
	})
	labelStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 11},
		Alignment: &excelize.Alignment{Vertical: "center"},
		Border:    thinBorder,
	})
	valueStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 11},
		Alignment: &excelize.Alignment{Vertical: "center"},
		Border:    thinBorder,
	})
	numValueStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 11, Bold: true, Color: "#2E75B6"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border:    thinBorder,
	})

	periodNames := map[string]string{
		"daily":   "日报",
		"weekly":  "周报",
		"monthly": "月报",
	}
	title := strings.TrimSpace(customTitle)
	if title == "" {
		title = periodNames[report.Period]
		if title != "" {
			title = fmt.Sprintf("嘉祥智能医务室 %s", title)
		}
	}
	if title == "" {
		title = "报表"
	}

	// --- Title ---
	_ = f.SetRowHeight(sheet, 1, 36)
	_ = f.SetCellValue(sheet, "A1", title)
	_ = f.SetCellStyle(sheet, "A1", "A1", titleStyle)
	_ = f.MergeCell(sheet, "A1", "C1")

	_ = f.SetCellValue(sheet, "A2", fmt.Sprintf("报表周期：%s", label))
	_ = f.SetCellValue(sheet, "B2", fmt.Sprintf("生成时间：%s", time.Now().Format("2006-01-02 15:04")))

	// --- Section: 基本统计 ---
	_ = f.SetCellValue(sheet, "A4", "基本统计")
	_ = f.SetCellStyle(sheet, "A4", "B4", sectionStyle)
	_ = f.MergeCell(sheet, "A4", "B4")

	stats := []struct {
		label string
		value any
	}{
		{"统计范围", fmt.Sprintf("%s ~ %s", report.StartAt.Format("2006-01-02"), report.EndAt.Format("2006-01-02"))},
		{"总就诊人次", report.TotalVisits},
		{"留观人次", report.ObservationVisits},
	}

	for i, s := range stats {
		row := 5 + i
		_ = f.SetCellValue(sheet, fmt.Sprintf("A%d", row), s.label)
		_ = f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), labelStyle)
		_ = f.SetCellValue(sheet, fmt.Sprintf("B%d", row), s.value)

		switch s.value.(type) {
		case int64, int:
			_ = f.SetCellStyle(sheet, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), numValueStyle)
		default:
			_ = f.SetCellStyle(sheet, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), valueStyle)
		}
	}

	// --- Section: 去向分布 ---
	distStart := 5 + len(stats) + 1
	_ = f.SetCellValue(sheet, fmt.Sprintf("A%d", distStart), "去向分布")
	_ = f.SetCellStyle(sheet, fmt.Sprintf("A%d", distStart), fmt.Sprintf("B%d", distStart), sectionStyle)
	_ = f.MergeCell(sheet, fmt.Sprintf("A%d", distStart), fmt.Sprintf("B%d", distStart))

	destLabels := map[string]string{
		"classroom":   "返回教室",
		"dormitory":   "返回寝室",
		"observation": "留观",
		"hospital":    "转院",
		"home":        "离校回家",
		"unknown":     "未登记",
	}

	i := 0
	for dest, count := range report.DestinationDistribution {
		row := distStart + 1 + i
		destLabel := dest
		if l, ok := destLabels[dest]; ok {
			destLabel = l
		}
		_ = f.SetCellValue(sheet, fmt.Sprintf("A%d", row), destLabel)
		_ = f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), labelStyle)
		_ = f.SetCellValue(sheet, fmt.Sprintf("B%d", row), count)
		_ = f.SetCellStyle(sheet, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), numValueStyle)
		i++
	}

	// --- Column widths ---
	_ = f.SetColWidth(sheet, "A", "A", 18)
	_ = f.SetColWidth(sheet, "B", "B", 32)
	_ = f.SetColWidth(sheet, "C", "C", 20)
}
