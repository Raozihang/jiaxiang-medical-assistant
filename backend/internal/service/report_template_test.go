package service

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jiaxiang-medical-assistant/backend/internal/repository"
)

func newTestReportTemplateService() *ReportTemplateService {
	tplRepo := repository.NewMemoryReportTemplateRepository()
	schedRepo := repository.NewMemoryReportScheduleRepository()
	visitRepo := repository.NewMockVisitRepository()
	medRepo := repository.NewMockMedicineRepository()
	reportSvc := NewReportService(visitRepo, medRepo)
	return NewReportTemplateService(tplRepo, schedRepo, reportSvc)
}

func TestCreateAndListTemplates(t *testing.T) {
	svc := newTestReportTemplateService()
	ctx := context.Background()

	tpl, err := svc.CreateTemplate(ctx, CreateTemplateInput{
		Name:    "每日简报",
		Period:  "daily",
		Columns: []string{"student_name", "class_name", "diagnosis", "created_at"},
		Title:   "嘉祥医务室每日简报",
	})
	if err != nil {
		t.Fatalf("create template: %v", err)
	}
	if tpl.ID == "" {
		t.Fatal("expected non-empty ID")
	}
	if len(tpl.Columns) != 4 {
		t.Fatalf("expected 4 columns, got %d", len(tpl.Columns))
	}

	list, err := svc.ListTemplates(ctx)
	if err != nil {
		t.Fatalf("list templates: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 template, got %d", len(list))
	}
}

func TestUpdateTemplate(t *testing.T) {
	svc := newTestReportTemplateService()
	ctx := context.Background()

	tpl, _ := svc.CreateTemplate(ctx, CreateTemplateInput{
		Name:    "test",
		Period:  "weekly",
		Columns: []string{"student_name"},
	})

	newName := "updated"
	updated, err := svc.UpdateTemplate(ctx, tpl.ID, UpdateTemplateInput{
		Name:    &newName,
		Columns: []string{"student_name", "diagnosis", "destination"},
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if updated.Name != "updated" {
		t.Fatalf("expected updated name, got %s", updated.Name)
	}
	if len(updated.Columns) != 3 {
		t.Fatalf("expected 3 columns, got %d", len(updated.Columns))
	}
}

func TestCreateTemplateRejectsInvalidInput(t *testing.T) {
	svc := newTestReportTemplateService()
	ctx := context.Background()

	_, err := svc.CreateTemplate(ctx, CreateTemplateInput{Name: "", Period: "daily"})
	if err == nil {
		t.Fatal("expected error on empty name")
	}

	_, err = svc.CreateTemplate(ctx, CreateTemplateInput{Name: "test", Period: "invalid"})
	if err == nil {
		t.Fatal("expected error on invalid period")
	}
}

func TestExportWithTemplate(t *testing.T) {
	svc := newTestReportTemplateService()
	ctx := context.Background()

	tpl, _ := svc.CreateTemplate(ctx, CreateTemplateInput{
		Name:    "export-test",
		Period:  "daily",
		Columns: []string{"student_name", "diagnosis"},
	})

	result, err := svc.ExportWithTemplate(ctx, tpl.ID)
	if err != nil {
		t.Fatalf("export: %v", err)
	}
	defer result.File.Close()

	if result.Filename == "" {
		t.Fatal("expected non-empty filename")
	}

	title, err := result.File.GetCellValue("统计概览", "A1")
	if err != nil {
		t.Fatalf("get summary title: %v", err)
	}
	if title != "嘉祥智能医务室 日报" {
		t.Fatalf("expected default title, got %q", title)
	}
}

func TestExportWithTemplateUsesCustomTitle(t *testing.T) {
	svc := newTestReportTemplateService()
	ctx := context.Background()

	tpl, _ := svc.CreateTemplate(ctx, CreateTemplateInput{
		Name:    "custom-title",
		Period:  "daily",
		Columns: []string{"student_name", "diagnosis"},
		Title:   "自定义医务室日报",
	})

	result, err := svc.ExportWithTemplate(ctx, tpl.ID)
	if err != nil {
		t.Fatalf("export: %v", err)
	}
	defer result.File.Close()

	title, err := result.File.GetCellValue("统计概览", "A1")
	if err != nil {
		t.Fatalf("get summary title: %v", err)
	}
	if title != "自定义医务室日报" {
		t.Fatalf("expected custom title, got %q", title)
	}
}

func TestCreateAndListSchedules(t *testing.T) {
	svc := newTestReportTemplateService()
	ctx := context.Background()

	tpl, _ := svc.CreateTemplate(ctx, CreateTemplateInput{
		Name:   "sched-test",
		Period: "daily",
	})

	sched, err := svc.CreateSchedule(ctx, CreateScheduleInput{
		TemplateID: tpl.ID,
		CronExpr:   "18:00",
	})
	if err != nil {
		t.Fatalf("create schedule: %v", err)
	}
	if !sched.Enabled {
		t.Fatal("expected enabled by default")
	}
	if sched.NextRunAt == nil {
		t.Fatal("expected next run time")
	}

	list, err := svc.ListSchedules(ctx)
	if err != nil {
		t.Fatalf("list schedules: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 schedule, got %d", len(list))
	}
}

func TestDeleteTemplateAlsoDeletesSchedules(t *testing.T) {
	svc := newTestReportTemplateService()
	ctx := context.Background()

	tpl, _ := svc.CreateTemplate(ctx, CreateTemplateInput{Name: "delete-test", Period: "daily"})
	if _, err := svc.CreateSchedule(ctx, CreateScheduleInput{TemplateID: tpl.ID, CronExpr: "18:00"}); err != nil {
		t.Fatalf("create schedule: %v", err)
	}

	if err := svc.DeleteTemplate(ctx, tpl.ID); err != nil {
		t.Fatalf("delete template: %v", err)
	}

	list, err := svc.ListSchedules(ctx)
	if err != nil {
		t.Fatalf("list schedules: %v", err)
	}
	if len(list) != 0 {
		t.Fatalf("expected schedules to be deleted, got %d", len(list))
	}
}

func TestUpdateScheduleToggleResetsNextRun(t *testing.T) {
	svc := newTestReportTemplateService()
	ctx := context.Background()

	tpl, _ := svc.CreateTemplate(ctx, CreateTemplateInput{Name: "toggle-test", Period: "daily"})
	sched, err := svc.CreateSchedule(ctx, CreateScheduleInput{TemplateID: tpl.ID, CronExpr: "18:00"})
	if err != nil {
		t.Fatalf("create schedule: %v", err)
	}
	if sched.NextRunAt == nil {
		t.Fatal("expected next run time")
	}

	disabled := false
	sched, err = svc.UpdateSchedule(ctx, sched.ID, UpdateScheduleInput{Enabled: &disabled})
	if err != nil {
		t.Fatalf("disable schedule: %v", err)
	}
	if sched.NextRunAt != nil {
		t.Fatal("expected next run to be cleared when disabled")
	}

	enabled := true
	sched, err = svc.UpdateSchedule(ctx, sched.ID, UpdateScheduleInput{Enabled: &enabled})
	if err != nil {
		t.Fatalf("enable schedule: %v", err)
	}
	if sched.NextRunAt == nil {
		t.Fatal("expected next run to be recalculated when enabled")
	}
}

func TestRunDueSchedules(t *testing.T) {
	svc := newTestReportTemplateService()
	ctx := context.Background()
	tempDir := t.TempDir()

	tpl, _ := svc.CreateTemplate(ctx, CreateTemplateInput{Name: "runner-test", Period: "daily"})
	sched, err := svc.CreateSchedule(ctx, CreateScheduleInput{TemplateID: tpl.ID, CronExpr: "18:00"})
	if err != nil {
		t.Fatalf("create schedule: %v", err)
	}

	now := fixedTime()
	past := now.Add(-time.Minute)
	sched.NextRunAt = &past
	if err := svc.schedRepo.Update(ctx, sched); err != nil {
		t.Fatalf("force due schedule: %v", err)
	}

	runCount, err := svc.RunDueSchedules(ctx, now, tempDir)
	if err != nil {
		t.Fatalf("run due schedules: %v", err)
	}
	if runCount != 1 {
		t.Fatalf("expected 1 due schedule to run, got %d", runCount)
	}

	updated, err := svc.schedRepo.Get(ctx, sched.ID)
	if err != nil {
		t.Fatalf("get updated schedule: %v", err)
	}
	if updated.LastRunAt == nil || !updated.LastRunAt.Equal(now) {
		t.Fatalf("expected last run at %v, got %+v", now, updated.LastRunAt)
	}
	if updated.NextRunAt == nil || !updated.NextRunAt.After(now) {
		t.Fatalf("expected next run after now, got %+v", updated.NextRunAt)
	}

	entries, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("read output dir: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 generated file, got %d", len(entries))
	}
	if filepath.Ext(entries[0].Name()) != ".xlsx" {
		t.Fatalf("expected xlsx output, got %s", entries[0].Name())
	}
}

func TestRunScheduleNowAndListFiles(t *testing.T) {
	svc := newTestReportTemplateService()
	ctx := context.Background()
	tempDir := t.TempDir()
	now := fixedTime()

	tpl, _ := svc.CreateTemplate(ctx, CreateTemplateInput{Name: "manual-run-test", Period: "daily"})
	sched, err := svc.CreateSchedule(ctx, CreateScheduleInput{TemplateID: tpl.ID, CronExpr: "18:00"})
	if err != nil {
		t.Fatalf("create schedule: %v", err)
	}

	result, err := svc.RunScheduleNow(ctx, sched.ID, now, tempDir)
	if err != nil {
		t.Fatalf("run schedule now: %v", err)
	}
	defer result.File.Close()

	files, err := svc.ListScheduleFiles(ctx, sched.ID, tempDir)
	if err != nil {
		t.Fatalf("list schedule files: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}
	if files[0].Name != result.Filename {
		t.Fatalf("expected file %s, got %s", result.Filename, files[0].Name)
	}

	fullPath, err := svc.ResolveScheduleFile(ctx, sched.ID, files[0].Name, tempDir)
	if err != nil {
		t.Fatalf("resolve schedule file: %v", err)
	}
	if _, err := os.Stat(fullPath); err != nil {
		t.Fatalf("expected resolved file to exist: %v", err)
	}
}

func TestRunScheduleNowKeepsDisabledScheduleUnscheduled(t *testing.T) {
	svc := newTestReportTemplateService()
	ctx := context.Background()
	now := fixedTime()

	tpl, _ := svc.CreateTemplate(ctx, CreateTemplateInput{Name: "disabled-run-test", Period: "daily"})
	sched, err := svc.CreateSchedule(ctx, CreateScheduleInput{TemplateID: tpl.ID, CronExpr: "18:00"})
	if err != nil {
		t.Fatalf("create schedule: %v", err)
	}

	disabled := false
	sched, err = svc.UpdateSchedule(ctx, sched.ID, UpdateScheduleInput{Enabled: &disabled})
	if err != nil {
		t.Fatalf("disable schedule: %v", err)
	}
	if sched.NextRunAt != nil {
		t.Fatal("expected disabled schedule to have no next run")
	}

	result, err := svc.RunScheduleNow(ctx, sched.ID, now, t.TempDir())
	if err != nil {
		t.Fatalf("run disabled schedule manually: %v", err)
	}
	defer result.File.Close()

	updated, err := svc.schedRepo.Get(ctx, sched.ID)
	if err != nil {
		t.Fatalf("get updated schedule: %v", err)
	}
	if updated.NextRunAt != nil {
		t.Fatal("expected manual run not to repopulate next_run_at for disabled schedule")
	}
}

func TestBuildScheduledFilenameIsUniqueWithinSameSecond(t *testing.T) {
	now := time.Date(2026, 3, 7, 8, 0, 0, 123, time.UTC)
	filename1 := buildScheduledFilename("schedule-12345678", "日报.xlsx", now)
	filename2 := buildScheduledFilename("schedule-12345678", "日报.xlsx", now.Add(100*time.Nanosecond))
	if filename1 == filename2 {
		t.Fatalf("expected unique filenames, got %s", filename1)
	}
}

func TestRunDueSchedulesContinuesAfterError(t *testing.T) {
	svc := newTestReportTemplateService()
	ctx := context.Background()
	tempDir := t.TempDir()
	now := fixedTime()
	past := now.Add(-time.Minute)

	validTpl, _ := svc.CreateTemplate(ctx, CreateTemplateInput{Name: "valid-schedule", Period: "daily"})
	validSched, err := svc.CreateSchedule(ctx, CreateScheduleInput{TemplateID: validTpl.ID, CronExpr: "18:00"})
	if err != nil {
		t.Fatalf("create valid schedule: %v", err)
	}
	validSched.NextRunAt = &past
	if err := svc.schedRepo.Update(ctx, validSched); err != nil {
		t.Fatalf("update valid schedule: %v", err)
	}

	brokenSched, err := svc.CreateSchedule(ctx, CreateScheduleInput{TemplateID: validTpl.ID, CronExpr: "18:00"})
	if err != nil {
		t.Fatalf("create broken schedule: %v", err)
	}
	brokenSched.TemplateID = "missing-template"
	brokenSched.NextRunAt = &past
	if err := svc.schedRepo.Update(ctx, brokenSched); err != nil {
		t.Fatalf("update broken schedule: %v", err)
	}

	runCount, err := svc.RunDueSchedules(ctx, now, tempDir)
	if err == nil {
		t.Fatal("expected aggregated error from broken due schedule")
	}
	if runCount != 1 {
		t.Fatalf("expected valid schedule to still run, got %d", runCount)
	}

	updated, err := svc.schedRepo.Get(ctx, validSched.ID)
	if err != nil {
		t.Fatalf("get valid schedule: %v", err)
	}
	if updated.LastRunAt == nil {
		t.Fatal("expected valid schedule to be processed despite another failing schedule")
	}
}

func TestCleanupExpiredScheduleFiles(t *testing.T) {
	tempDir := t.TempDir()
	now := fixedTime()

	expiredPath := filepath.Join(tempDir, "expired.xlsx")
	keptPath := filepath.Join(tempDir, "recent.xlsx")
	if err := os.WriteFile(expiredPath, []byte("old"), 0o644); err != nil {
		t.Fatalf("write expired file: %v", err)
	}
	if err := os.WriteFile(keptPath, []byte("new"), 0o644); err != nil {
		t.Fatalf("write recent file: %v", err)
	}
	oldTime := now.AddDate(0, 0, -31)
	recentTime := now.AddDate(0, 0, -5)
	if err := os.Chtimes(expiredPath, oldTime, oldTime); err != nil {
		t.Fatalf("set expired times: %v", err)
	}
	if err := os.Chtimes(keptPath, recentTime, recentTime); err != nil {
		t.Fatalf("set recent times: %v", err)
	}

	deletedCount, err := cleanupExpiredScheduleFiles(tempDir, now, 30)
	if err != nil {
		t.Fatalf("cleanup expired files: %v", err)
	}
	if deletedCount != 1 {
		t.Fatalf("expected 1 deleted file, got %d", deletedCount)
	}
	if _, err := os.Stat(expiredPath); !os.IsNotExist(err) {
		t.Fatalf("expected expired file removed, got err=%v", err)
	}
	if _, err := os.Stat(keptPath); err != nil {
		t.Fatalf("expected recent file kept: %v", err)
	}
}

func TestCleanupExpiredScheduleFilesDisabledWhenRetentionZero(t *testing.T) {
	tempDir := t.TempDir()
	now := fixedTime()
	filePath := filepath.Join(tempDir, "old.xlsx")
	if err := os.WriteFile(filePath, []byte("old"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	oldTime := now.AddDate(0, 0, -90)
	if err := os.Chtimes(filePath, oldTime, oldTime); err != nil {
		t.Fatalf("set old time: %v", err)
	}

	deletedCount, err := cleanupExpiredScheduleFiles(tempDir, now, 0)
	if err != nil {
		t.Fatalf("cleanup with disabled retention: %v", err)
	}
	if deletedCount != 0 {
		t.Fatalf("expected no deletion when retention disabled, got %d", deletedCount)
	}
	if _, err := os.Stat(filePath); err != nil {
		t.Fatalf("expected file kept when retention disabled: %v", err)
	}
}

func TestFilterValidColumns(t *testing.T) {
	cols := filterValidColumns([]string{"student_name", "invalid_col", "diagnosis", ""})
	if len(cols) != 2 {
		t.Fatalf("expected 2 valid columns, got %d: %v", len(cols), cols)
	}
}

func TestCalcNextRun(t *testing.T) {
	result := calcNextRun("18:00", parseHHMM("08:00", fixedTime()))
	if result == nil {
		t.Fatal("expected non-nil next run")
	}
	if result.Hour() != 18 {
		t.Fatalf("expected hour 18, got %d", result.Hour())
	}
}

func TestCalcNextRunForPeriod(t *testing.T) {
	weekly := calcNextRunForPeriod("weekly", "18:00", fixedTime())
	if weekly == nil || weekly.Weekday() != time.Sunday || weekly.Hour() != 18 {
		t.Fatalf("expected weekly schedule on Sunday 18:00, got %+v", weekly)
	}

	monthly := calcNextRunForPeriod("monthly", "18:00", fixedTime())
	if monthly == nil || monthly.Day() != 31 || monthly.Hour() != 18 {
		t.Fatalf("expected monthly schedule on month end 18:00, got %+v", monthly)
	}
}

func fixedTime() time.Time {
	return time.Date(2026, 3, 7, 8, 0, 0, 0, time.UTC)
}
