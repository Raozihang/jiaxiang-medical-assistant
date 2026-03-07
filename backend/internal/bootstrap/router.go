package bootstrap

import (
	"context"
	"fmt"
	"log"
	"strings"

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
	visitRepo, medicineRepo, studentContactRepo, outboundCallRepo, importTaskRepo := buildRepositories(dataMode, db)

	outboundProvider, err := buildOutboundCallProvider(cfg)
	if err != nil {
		return err
	}

	outboundCallService := service.NewOutboundCallService(
		outboundCallRepo,
		visitRepo,
		studentContactRepo,
		outboundProvider,
		cfg.Outbound.AliyunTemplateCode,
	)
	visitService := service.NewVisitService(visitRepo, outboundCallService)
	medicineService := service.NewMedicineService(medicineRepo)
	studentContactService := service.NewStudentContactService(studentContactRepo)
	importService := service.NewImportService(visitRepo, importTaskRepo)
	reportService := service.NewReportService(visitRepo, medicineRepo)
	authService, err := service.NewAuthService(cfg, dataMode)
	if err != nil {
		return err
	}

	seedContext := context.Background()
	if err := visitService.EnsureSeedData(seedContext); err != nil {
		log.Printf("visit seed failed: %v", err)
	}
	if err := medicineService.EnsureSeedData(seedContext); err != nil {
		log.Printf("medicine seed failed: %v", err)
	}

	healthHandler := handler.NewHealthHandler(cfg, dataMode)
	authHandler := handler.NewAuthHandler(authService)
	visitHandler := handler.NewVisitHandler(visitService)
	medicineHandler := handler.NewMedicineHandler(medicineService)
	studentContactHandler := handler.NewStudentContactHandler(studentContactService)
	importHandler := handler.NewImportHandler(importService)
	outboundCallHandler := handler.NewOutboundCallHandler(outboundCallService, cfg.Outbound.AliyunCallbackSecret)
	reportHandler := handler.NewReportHandler(reportService)

	api := engine.Group("/api/v1")
	{
		api.GET("/healthz", healthHandler.Healthz)
		api.POST("/auth/login", authHandler.Login)
		api.POST("/visits", visitHandler.Create)
		api.POST("/outbound-calls/callback/aliyun", outboundCallHandler.AliyunCallback)

		protected := api.Group("")
		protected.Use(middleware.AuthRequired(authService))
		{
			protected.GET("/visits", visitHandler.List)
			protected.GET("/visits/:id", visitHandler.Detail)
			protected.PATCH("/visits/:id", visitHandler.Update)

			protected.GET("/medicines", medicineHandler.List)
			protected.POST("/medicines/inbound", medicineHandler.Inbound)
			protected.POST("/medicines/outbound", medicineHandler.Outbound)

			protected.POST("/import/visits", importHandler.ImportVisits)
			protected.GET("/import/tasks", importHandler.TaskList)
			protected.GET("/import/tasks/:id", importHandler.TaskDetail)

			protected.GET("/students/contacts", studentContactHandler.List)
			protected.PUT("/students/:studentId/contact", studentContactHandler.Update)

			protected.GET("/outbound-calls", outboundCallHandler.List)
			protected.POST("/outbound-calls/:id/retry", outboundCallHandler.Retry)

			protected.GET("/reports/overview", reportHandler.Overview)
		}
	}

	return nil
}

func buildRepositories(dataMode string, db *gorm.DB) (repository.VisitRepository, repository.MedicineRepository, repository.StudentContactRepository, repository.OutboundCallRepository, repository.ImportTaskRepository) {
	if dataMode == "db" && db != nil {
		return repository.NewGormVisitRepository(db), repository.NewGormMedicineRepository(db), repository.NewGormStudentContactRepository(db), repository.NewGormOutboundCallRepository(db), repository.NewGormImportTaskRepository(db)
	}

	return repository.NewMockVisitRepository(), repository.NewMockMedicineRepository(), repository.NewMemoryStudentContactRepository(), repository.NewMemoryOutboundCallRepository(), repository.NewMemoryImportTaskRepository()
}

func buildOutboundCallProvider(cfg config.Config) (service.OutboundCallProvider, error) {
	switch strings.ToLower(strings.TrimSpace(cfg.Outbound.Provider)) {
	case "", "mock":
		return service.NewMockOutboundCallProvider(), nil
	case "aliyun":
		return service.NewAliyunOutboundCallProvider(cfg.Outbound)
	default:
		return nil, fmt.Errorf("unsupported outbound call provider: %s", cfg.Outbound.Provider)
	}
}
