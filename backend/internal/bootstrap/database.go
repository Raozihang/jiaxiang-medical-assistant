package bootstrap

import (
	"database/sql"
	"log"

	"github.com/jiaxiang-medical-assistant/backend/internal/config"
	"github.com/jiaxiang-medical-assistant/backend/internal/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitDatabase(cfg config.Config) (*gorm.DB, func()) {
	if !cfg.DB.IsConfigured() {
		log.Printf("database config not found, skip database bootstrap")
		return nil, func() {}
	}

	db, err := gorm.Open(postgres.Open(cfg.DB.DSN()), &gorm.Config{})
	if err != nil {
		log.Printf("database bootstrap failed: %v", err)
		return nil, func() {}
	}

	if err := db.AutoMigrate(&model.Student{}, &model.Visit{}, &model.Medicine{}); err != nil {
		log.Printf("database migration failed: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Printf("database sql handle failed: %v", err)
		return db, func() {}
	}

	return db, func() {
		if closeErr := closeDatabase(sqlDB); closeErr != nil {
			log.Printf("database close failed: %v", closeErr)
		}
	}
}

func closeDatabase(db *sql.DB) error {
	return db.Close()
}
