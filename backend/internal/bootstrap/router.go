package bootstrap

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/jiaxiang-medical-assistant/backend/internal/config"
	"github.com/jiaxiang-medical-assistant/backend/internal/handler"
	"github.com/jiaxiang-medical-assistant/backend/internal/middleware"
	"github.com/jiaxiang-medical-assistant/backend/internal/repository"
	"github.com/jiaxiang-medical-assistant/backend/internal/service"
	"gorm.io/gorm"
)

func registerRoutes(engine *gin.Engine, cfg config.Config, db *gorm.DB) error {
	dataMode := cfg.ResolveDataMode(db != nil)
	visitRepo, medicineRepo := buildRepositories(dataMode, db)
	importTaskRepo, notificationLogRepo, safetyAlertStateRepo, studentContactRepo, outboundCallRepo := buildStateRepositories(dataMode, db)
	outboundProvider, err := buildOutboundCallProvider(cfg)
	if err != nil {
		return err
	}

	outboundCallService := service.NewOutboundCallService(outboundCallRepo, visitRepo, studentContactRepo, outboundProvider, cfg.Outbound.AliyunTemplateCode)
	visitService := service.NewVisitService(visitRepo, outboundCallService)
	medicineService := service.NewMedicineService(medicineRepo)
	reportService := service.NewReportService(visitRepo, medicineRepo)
	aiService := service.NewAIService()
	importService := service.NewImportService(visitRepo, importTaskRepo)
	notificationService := service.NewNotificationService(notificationLogRepo)
	studentContactService := service.NewStudentContactService(studentContactRepo)
	safetyService := service.NewSafetyService(visitRepo, safetyAlertStateRepo)
	authService, err := service.NewAuthService(cfg, dataMode)
	if err != nil {
		return err
	}

	seedContext := context.Background()
	if err := visitService.EnsureSeedData(seedContext); err != nil {
		log.Printf("就诊种子数据初始化失败: %v", err)
	}
	if err := medicineService.EnsureSeedData(seedContext); err != nil {
		log.Printf("药品种子数据初始化失败: %v", err)
	}

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
		}
	}

	return nil
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
