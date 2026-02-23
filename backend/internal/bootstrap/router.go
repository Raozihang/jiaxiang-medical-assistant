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

	visitService := service.NewVisitService(visitRepo)
	medicineService := service.NewMedicineService(medicineRepo)
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
	reportHandler := handler.NewReportHandler(reportService)

	api := engine.Group("/api/v1")
	{
		api.GET("/healthz", healthHandler.Healthz)
		api.POST("/auth/login", authHandler.Login)

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
