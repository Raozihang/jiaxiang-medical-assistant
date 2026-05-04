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

	realtimeHub := service.NewRealtimeHub()
	visitService := service.NewVisitService(visitRepo, outboundCallService)
	visitService.SetRealtimeHub(realtimeHub)
	medicineService := service.NewMedicineService(medicineRepo)
	reportService := service.NewReportService(visitRepo, medicineRepo)
	aiService := buildAIService(cfg, medicineRepo)
	aiAnalysisService := service.NewAIAnalysisService(visitRepo, aiService)
	visitService.SetAIAnalysisQueue(aiAnalysisService)
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
	aiAnalysisService.Start(runnerContext, 2)
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
	realtimeHandler := handler.NewRealtimeHandler(realtimeHub, visitService, authService)

	api := engine.Group("/api/v1")
	{
		api.GET("/healthz", healthHandler.Healthz)
		api.POST("/auth/login", authHandler.Login)
		api.POST("/outbound-calls/callback/aliyun", outboundCallHandler.AliyunCallback)

		api.POST("/visits", visitHandler.Create)
		api.GET("/realtime/checkin", realtimeHandler.CheckIn)
		api.GET("/realtime/doctor", realtimeHandler.Doctor)

		protected := api.Group("")
		protected.Use(middleware.AuthRequired(authService))
		{
			medical := protected.Group("")
			medical.Use(middleware.RequireRoles("doctor", "admin"))
			{
				medical.GET("/visits", visitHandler.List)
				medical.GET("/visits/:id", visitHandler.Detail)
				medical.PATCH("/visits/:id", visitHandler.Update)
				medical.POST("/visits/:id/ai-analysis/regenerate", visitHandler.RegenerateAIAnalysis)

				medical.GET("/medicines", medicineHandler.List)
				medical.POST("/medicines", medicineHandler.Create)
				medical.POST("/medicines/inbound", medicineHandler.Inbound)
				medical.POST("/medicines/outbound", medicineHandler.Outbound)
				medical.PATCH("/medicines/:id/inventory", medicineHandler.UpdateInventory)

				medical.POST("/ai/analyze", aiHandler.Analyze)
				medical.POST("/ai/triage", aiHandler.Triage)
				medical.POST("/ai/recommend", aiHandler.Recommend)
				medical.POST("/ai/interaction-check", aiHandler.InteractionCheck)
			}

			admin := protected.Group("")
			admin.Use(middleware.RequireRoles("admin"))
			{
				admin.GET("/reports/overview", reportHandler.Overview)
				admin.GET("/reports/daily", reportHandler.Daily)
				admin.GET("/reports/weekly", reportHandler.Weekly)
				admin.GET("/reports/monthly", reportHandler.Monthly)
				admin.GET("/reports/export/daily", reportHandler.ExportDaily)
				admin.GET("/reports/export/weekly", reportHandler.ExportWeekly)
				admin.GET("/reports/export/monthly", reportHandler.ExportMonthly)

				admin.POST("/import/visits", importHandler.ImportVisits)
				admin.GET("/import/tasks", importHandler.Tasks)
				admin.GET("/import/tasks/:id", importHandler.TaskDetail)

				admin.POST("/notifications/send", notificationHandler.Send)
				admin.POST("/notifications/dispatch", notificationHandler.Dispatch)
				admin.GET("/notifications/logs", notificationHandler.Logs)
				admin.GET("/students/contacts", studentContactHandler.List)
				admin.PUT("/students/:studentId/contact", studentContactHandler.Update)
				admin.GET("/outbound-calls", outboundCallHandler.List)
				admin.POST("/outbound-calls/:id/retry", outboundCallHandler.Retry)

				admin.GET("/safety/alerts", safetyHandler.Alerts)
				admin.PATCH("/safety/alerts/:id", safetyHandler.UpdateAlert)

				admin.GET("/report-templates/columns", reportTemplateHandler.ColumnOptions)
				admin.POST("/report-templates", reportTemplateHandler.CreateTemplate)
				admin.GET("/report-templates", reportTemplateHandler.ListTemplates)
				admin.GET("/report-templates/:id", reportTemplateHandler.GetTemplate)
				admin.PATCH("/report-templates/:id", reportTemplateHandler.UpdateTemplate)
				admin.DELETE("/report-templates/:id", reportTemplateHandler.DeleteTemplate)
				admin.GET("/report-templates/:id/export", reportTemplateHandler.ExportWithTemplate)

				admin.POST("/report-schedules", reportTemplateHandler.CreateSchedule)
				admin.GET("/report-schedules", reportTemplateHandler.ListSchedules)
				admin.POST("/report-schedules/:id/run", reportTemplateHandler.TriggerSchedule)
				admin.GET("/report-schedules/:id/files", reportTemplateHandler.ListScheduleFiles)
				admin.GET("/report-schedules/:id/files/*filename", reportTemplateHandler.DownloadScheduleFile)
				admin.PATCH("/report-schedules/:id", reportTemplateHandler.UpdateSchedule)
				admin.DELETE("/report-schedules/:id", reportTemplateHandler.DeleteSchedule)
			}
		}
	}

	return cancelRunner, nil
}

func buildAIService(cfg config.Config, medicineRepo repository.MedicineRepository) *service.AIService {
	if cfg.AI.Provider == "bailian" && cfg.AI.APIKey != "" {
		provider := service.NewBailianProvider(cfg.AI.APIKey, cfg.AI.Model, cfg.AI.BaseURL)
		log.Printf("AI provider: bailian (model=%s)", cfg.AI.Model)
		return service.NewAIServiceWithDependencies(provider, medicineRepo)
	}
	log.Printf("AI provider: rule-based (set AI_PROVIDER=bailian to use LLM)")
	return service.NewAIServiceWithDependencies(nil, medicineRepo)
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
