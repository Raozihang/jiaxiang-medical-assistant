package service

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/jiaxiang-medical-assistant/backend/internal/repository"
)

type ReportScheduleRunner struct {
	svc       *ReportTemplateService
	interval  time.Duration
	outputDir string
	retentionDays int
}

type ScheduledReportFile struct {
	Name       string    `json:"name"`
	SizeBytes  int64     `json:"size_bytes"`
	ModifiedAt time.Time `json:"modified_at"`
}

func DefaultReportScheduleOutputDir() string {
	return filepath.Join("output", "reports")
}

func NewReportScheduleRunner(svc *ReportTemplateService, interval time.Duration, outputDir string, retentionDays int) *ReportScheduleRunner {
	if interval <= 0 {
		interval = time.Minute
	}
	if strings.TrimSpace(outputDir) == "" {
		outputDir = DefaultReportScheduleOutputDir()
	}

	return &ReportScheduleRunner{svc: svc, interval: interval, outputDir: outputDir, retentionDays: retentionDays}
}

func (r *ReportScheduleRunner) Start(ctx context.Context) {
	r.runOnce(ctx)

	ticker := time.NewTicker(r.interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				r.runOnce(ctx)
			}
		}
	}()
}

func (r *ReportScheduleRunner) runOnce(ctx context.Context) {
	count, err := r.svc.RunDueSchedules(ctx, time.Now().UTC(), r.outputDir)
	if err != nil {
		log.Printf("report schedule runner failed: %v", err)
	}
	if count > 0 {
		log.Printf("report schedule runner processed %d task(s)", count)
	}

	deletedCount, cleanupErr := cleanupExpiredScheduleFiles(r.outputDir, time.Now().UTC(), r.retentionDays)
	if cleanupErr != nil {
		log.Printf("report schedule cleanup failed: %v", cleanupErr)
		return
	}
	if deletedCount > 0 {
		log.Printf("report schedule cleanup deleted %d expired file(s)", deletedCount)
	}
}

func (s *ReportTemplateService) RunDueSchedules(ctx context.Context, now time.Time, outputDir string) (int, error) {
	dueSchedules, err := s.schedRepo.ListDue(ctx, now)
	if err != nil {
		return 0, err
	}

	runCount := 0
	var firstErr error
	for _, sched := range dueSchedules {
		tpl, err := s.tplRepo.Get(ctx, sched.TemplateID)
		if err != nil {
			log.Printf("report schedule %s skipped: load template %s failed: %v", sched.ID, sched.TemplateID, err)
			if firstErr == nil {
				firstErr = err
			}
			continue
		}

		result, err := s.runSchedule(ctx, tpl, &sched, now, outputDir)
		if err != nil {
			log.Printf("report schedule %s skipped: run failed: %v", sched.ID, err)
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		result.File.Close()
		runCount++
	}

	return runCount, firstErr
}

func (s *ReportTemplateService) RunScheduleNow(ctx context.Context, scheduleID string, now time.Time, outputDir string) (*ExcelExportResult, error) {
	sched, err := s.schedRepo.Get(ctx, scheduleID)
	if err != nil {
		return nil, err
	}
	tpl, err := s.tplRepo.Get(ctx, sched.TemplateID)
	if err != nil {
		return nil, err
	}

	return s.runSchedule(ctx, tpl, &sched, now, outputDir)
}

func (s *ReportTemplateService) ListScheduleFiles(ctx context.Context, scheduleID string, outputDir string) ([]ScheduledReportFile, error) {
	if _, err := s.schedRepo.Get(ctx, scheduleID); err != nil {
		return nil, err
	}

	files, err := readScheduleFiles(scheduleID, outputDir)
	if err != nil {
		return nil, err
	}
	return files, nil
}

func (s *ReportTemplateService) ResolveScheduleFile(ctx context.Context, scheduleID string, fileName string, outputDir string) (string, error) {
	if _, err := s.schedRepo.Get(ctx, scheduleID); err != nil {
		return "", err
	}

	cleanName := filepath.Base(strings.TrimSpace(fileName))
	if cleanName == "." || cleanName == "" {
		return "", ErrInvalidInput
	}

	files, err := readScheduleFiles(scheduleID, outputDir)
	if err != nil {
		return "", err
	}
	for _, file := range files {
		if file.Name == cleanName {
			return filepath.Join(normalizeOutputDir(outputDir), cleanName), nil
		}
	}

	return "", repository.ErrNotFound
}

func (s *ReportTemplateService) runSchedule(ctx context.Context, tpl repository.ReportTemplate, sched *repository.ReportSchedule, now time.Time, outputDir string) (*ExcelExportResult, error) {
	result, err := s.ExportWithTemplate(ctx, tpl.ID)
	if err != nil {
		return nil, err
	}

	normalizedOutputDir := normalizeOutputDir(outputDir)
	if err := os.MkdirAll(normalizedOutputDir, 0o755); err != nil {
		result.File.Close()
		return nil, err
	}

	filename := buildScheduledFilename(sched.ID, result.Filename, now)
	fullPath := filepath.Join(normalizedOutputDir, filename)
	if err := result.File.SaveAs(fullPath); err != nil {
		result.File.Close()
		return nil, err
	}

	sched.LastRunAt = &now
	if sched.Enabled {
		sched.NextRunAt = calcNextRunForPeriod(tpl.Period, sched.CronExpr, now)
	}
	sched.UpdatedAt = now
	if err := s.schedRepo.Update(ctx, *sched); err != nil {
		result.File.Close()
		return nil, err
	}

	result.Filename = filename
	return result, nil
}

var invalidFilenameChars = regexp.MustCompile(`[\\/:*?"<>|\s]+`)

func buildScheduledFilename(scheduleID string, sourceFilename string, now time.Time) string {
	cleanName := invalidFilenameChars.ReplaceAllString(strings.TrimSpace(sourceFilename), "_")
	cleanName = strings.Trim(cleanName, "._")
	if cleanName == "" {
		cleanName = "report.xlsx"
	}

	return fmt.Sprintf("%s_%09d_%s_%s", now.Format("20060102_150405"), now.Nanosecond(), shortScheduleID(scheduleID), cleanName)
}

func normalizeOutputDir(outputDir string) string {
	if strings.TrimSpace(outputDir) == "" {
		return DefaultReportScheduleOutputDir()
	}
	return outputDir
}

func readScheduleFiles(scheduleID string, outputDir string) ([]ScheduledReportFile, error) {
	entries, err := os.ReadDir(normalizeOutputDir(outputDir))
	if err != nil {
		if os.IsNotExist(err) {
			return []ScheduledReportFile{}, nil
		}
		return nil, err
	}

	prefix := fmt.Sprintf("_%s_", shortScheduleID(scheduleID))
	files := make([]ScheduledReportFile, 0)
	for _, entry := range entries {
		if entry.IsDir() || !strings.Contains(entry.Name(), prefix) {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			return nil, err
		}
		files = append(files, ScheduledReportFile{
			Name:       entry.Name(),
			SizeBytes:  info.Size(),
			ModifiedAt: info.ModTime().UTC(),
		})
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].ModifiedAt.After(files[j].ModifiedAt)
	})
	return files, nil
}

func shortScheduleID(scheduleID string) string {
	shortID := scheduleID
	if len(shortID) > 8 {
		shortID = shortID[:8]
	}
	return shortID
}

func cleanupExpiredScheduleFiles(outputDir string, now time.Time, retentionDays int) (int, error) {
	if retentionDays <= 0 {
		return 0, nil
	}

	entries, err := os.ReadDir(normalizeOutputDir(outputDir))
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}

	cutoff := now.UTC().AddDate(0, 0, -retentionDays)
	deletedCount := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			return deletedCount, err
		}
		if !info.ModTime().UTC().Before(cutoff) {
			continue
		}
		if err := os.Remove(filepath.Join(normalizeOutputDir(outputDir), entry.Name())); err != nil {
			return deletedCount, err
		}
		deletedCount++
	}

	return deletedCount, nil
}
