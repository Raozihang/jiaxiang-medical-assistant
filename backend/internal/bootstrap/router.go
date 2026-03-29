package bootstrap

import (
	"context"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jiaxiang-medical-assistant/backend/internal/config"
	"github.com/jiaxiang-medical-assistant/backend/internal/handler"
	"github.com/jiaxiang-medical-assistant/backend/internal/middleware"
	"github.com/jiaxiang-medical-assistant/backend/internal/repository"
	"github.com/jiaxiang-medical-assistant/backend/internal/service"
	"gorm.io/gorm"
)

func registerRoutes(engine *gin.Engine, cfg config.Config, db *gorm.DB) (func(), error) {
	dataMode := cfg.ResolveDataMode(db != nil)
	visitRepo, medicineRepo := buildRepositories(dataMode, db)
	importTaskRepo, notificationLogRepo, safetyAlertStateRepo, studentContactRepo, outboundCallRepo := buildStateRepositories(dataMode, db)
	reportTplRepo, reportSchedRepo := buildReportTemplateRepositories(dataMode, db)
	outboundProvider, err := buildOutboundCallProvider(cfg)
	if err != nil {
		return func() {}, err
	}
	outboundCallService := service.NewOutboundCallService(outboundCallRepo, visitRepo, studentContactRepo, outboundProvider, cfg.Outbound.AliyunTemplateCode)

	visitService := service.NewVisitService(visitRepo, outboundCallService)
	medicineService := service.NewMedicineService(medicineRepo)
	reportService := service.NewReportService(visitRepo, medicineRepo)
	aiService := buildAIService(cfg)
	importService := service.NewImportService(visitRepo, importTaskRepo)
	notificationService := service.NewNotificationService(notificationLogRepo)
	studentContactService := service.NewStudentContactService(studentContactRepo)
	safetyService := service.NewSafetyService(visitRepo, safetyAlertStateRepo)
	reportTemplateService := service.NewReportTemplateService(reportTplRepo, reportSchedRepo, reportService)
	authService, err := service.NewAuthService(cfg, dataMode)
	if err != nil {
		return func() {}, err
	}

	seedContext := context.Background()
	if err := visitService.EnsureSeedData(seedContext); err != nil {
		log.Printf("就诊种子数据初始化失败: %v", err)
	}
	if err := medicineService.EnsureSeedData(seedContext); err != nil {
		log.Printf("药品种子数据初始化失败: %v", err)
	}

	runnerContext, cancelRunner := context.WithCancel(context.Background())
	service.NewReportScheduleRunner(
		reportTemplateService,
		time.Minute,
		service.DefaultReportScheduleOutputDir(),
		cfg.Report.ScheduleRetentionDays,
	).Start(runnerContext)

	healthHandler := handler.NewHealthHandler(cfg, dataMode)
	authHandler := handler.NewAuthHandler(authService)
	visitHandler := handler.NewVisitHandler(visitService)
	medicineHandler := handler.NewMedicineHandler(medicineService)
	reportHandler := handler.NewReportHandler(reportService)
	aiHandler := handler.NewAIHandler(aiService)
	importHandler := handler.NewImportHandler(importService)
	notificationHandler := handler.NewNotificationHandler(notificationService)
	studentContactHandler := handler.NewStudentContactHandler(studentContactService)
	outboundCallHandler := handler.NewOutboundCallHandler(outboundCallService, cfg.Outbound.AliyunCallbackSecret)
	safetyHandler := handler.NewSafetyHandler(safetyService)
	reportTemplateHandler := handler.NewReportTemplateHandler(reportTemplateService)

	api := engine.Group("/api/v1")
	{
		api.GET("/healthz", healthHandler.Healthz)
		api.POST("/auth/login", authHandler.Login)
		api.POST("/outbound-calls/callback/aliyun", outboundCallHandler.AliyunCallback)

		api.POST("/visits", visitHandler.Create)

		protected := api.Group("")
		protected.Use(middleware.AuthRequired(authService))
		{
			protected.GET("/visits", visitHandler.List)
			protected.GET("/visits/:id", visitHandler.Detail)
			protected.PATCH("/visits/:id", visitHandler.Update)

			protected.GET("/medicines", medicineHandler.List)
			protected.POST("/medicines/inbound", medicineHandler.Inbound)
			protected.POST("/medicines/outbound", medicineHandler.Outbound)

			protected.GET("/reports/overview", reportHandler.Overview)
			protected.GET("/reports/daily", reportHandler.Daily)
			protected.GET("/reports/weekly", reportHandler.Weekly)
			protected.GET("/reports/monthly", reportHandler.Monthly)
			protected.GET("/reports/export/daily", reportHandler.ExportDaily)
			protected.GET("/reports/export/weekly", reportHandler.ExportWeekly)
			protected.GET("/reports/export/monthly", reportHandler.ExportMonthly)

			protected.POST("/ai/analyze", aiHandler.Analyze)
			protected.POST("/ai/triage", aiHandler.Triage)
			protected.POST("/ai/recommend", aiHandler.Recommend)
			protected.POST("/ai/interaction-check", aiHandler.InteractionCheck)

			protected.POST("/import/visits", importHandler.ImportVisits)
			protected.GET("/import/tasks", importHandler.Tasks)
			protected.GET("/import/tasks/:id", importHandler.TaskDetail)

			protected.POST("/notifications/send", notificationHandler.Send)
			protected.POST("/notifications/dispatch", notificationHandler.Dispatch)
			protected.GET("/notifications/logs", notificationHandler.Logs)
			protected.GET("/students/contacts", studentContactHandler.List)
			protected.PUT("/students/:studentId/contact", studentContactHandler.Update)
			protected.GET("/outbound-calls", outboundCallHandler.List)
			protected.POST("/outbound-calls/:id/retry", outboundCallHandler.Retry)

			protected.GET("/safety/alerts", safetyHandler.Alerts)
			protected.PATCH("/safety/alerts/:id", safetyHandler.UpdateAlert)

			protected.GET("/report-templates/columns", reportTemplateHandler.ColumnOptions)
			protected.POST("/report-templates", reportTemplateHandler.CreateTemplate)
			protected.GET("/report-templates", reportTemplateHandler.ListTemplates)
			protected.GET("/report-templates/:id", reportTemplateHandler.GetTemplate)
			protected.PATCH("/report-templates/:id", reportTemplateHandler.UpdateTemplate)
			protected.DELETE("/report-templates/:id", reportTemplateHandler.DeleteTemplate)
			protected.GET("/report-templates/:id/export", reportTemplateHandler.ExportWithTemplate)

			protected.POST("/report-schedules", reportTemplateHandler.CreateSchedule)
			protected.GET("/report-schedules", reportTemplateHandler.ListSchedules)
			protected.POST("/report-schedules/:id/run", reportTemplateHandler.TriggerSchedule)
			protected.GET("/report-schedules/:id/files", reportTemplateHandler.ListScheduleFiles)
			protected.GET("/report-schedules/:id/files/*filename", reportTemplateHandler.DownloadScheduleFile)
			protected.PATCH("/report-schedules/:id", reportTemplateHandler.UpdateSchedule)
			protected.DELETE("/report-schedules/:id", reportTemplateHandler.DeleteSchedule)
		}
	}

	return cancelRunner, nil
}

func buildAIService(cfg config.Config) *service.AIService {
	if cfg.AI.Provider == "bailian" && cfg.AI.APIKey != "" {
		provider := service.NewBailianProvider(cfg.AI.APIKey, cfg.AI.Model, cfg.AI.BaseURL)
		log.Printf("AI provider: bailian (model=%s)", cfg.AI.Model)
		return service.NewAIServiceWithProvider(provider)
	}
	log.Printf("AI provider: rule-based (set AI_PROVIDER=bailian to use LLM)")
	return service.NewAIService()
}

func buildRepositories(dataMode string, db *gorm.DB) (repository.VisitRepository, repository.MedicineRepository) {
	if dataMode == "db" && db != nil {
		return repository.NewGormVisitRepository(db), repository.NewGormMedicineRepository(db)
	}

	return repository.NewMockVisitRepository(), repository.NewMockMedicineRepository()
}

func buildStateRepositories(dataMode string, db *gorm.DB) (
	repository.ImportTaskRepository,
	repository.NotificationLogRepository,
	repository.SafetyAlertStateRepository,
	repository.StudentContactRepository,
	repository.OutboundCallRepository,
) {
	if dataMode == "db" && db != nil {
		return repository.NewGormImportTaskRepository(db),
			repository.NewGormNotificationLogRepository(db),
			repository.NewGormSafetyAlertStateRepository(db),
			repository.NewGormStudentContactRepository(db),
			repository.NewGormOutboundCallRepository(db)
	}

	return repository.NewMemoryImportTaskRepository(),
		repository.NewMemoryNotificationLogRepository(),
		repository.NewMemorySafetyAlertStateRepository(),
		repository.NewMemoryStudentContactRepository(),
		repository.NewMemoryOutboundCallRepository()
}

func buildReportTemplateRepositories(dataMode string, db *gorm.DB) (repository.ReportTemplateRepository, repository.ReportScheduleRepository) {
	if dataMode == "db" && db != nil {
		return repository.NewGormReportTemplateRepository(db), repository.NewGormReportScheduleRepository(db)
	}

	return repository.NewMemoryReportTemplateRepository(), repository.NewMemoryReportScheduleRepository()
}

func buildOutboundCallProvider(cfg config.Config) (service.OutboundCallProvider, error) {
	if cfg.Outbound.Provider == "aliyun" {
		provider, err := service.NewAliyunOutboundCallProvider(cfg.Outbound)
		if err != nil {
			return nil, err
		}
		log.Printf("outbound call provider: aliyun")
		return provider, nil
	}

	log.Printf("outbound call provider: mock")
	return service.NewMockOutboundCallProvider(), nil
}
