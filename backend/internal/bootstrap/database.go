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
		log.Printf("未配置数据库连接，跳过数据库初始化")
		return nil, func() {}
	}

	db, err := gorm.Open(postgres.Open(cfg.DB.DSN()), &gorm.Config{})
	if err != nil {
		log.Printf("数据库初始化失败: %v", err)
		return nil, func() {}
	}

	if err := db.AutoMigrate(
		&model.Student{},
		&model.Visit{},
		&model.Medicine{},
		&model.ImportTask{},
		&model.NotificationLog{},
		&model.OutboundCall{},
		&model.SafetyAlertState{},
	); err != nil {
		log.Printf("数据库迁移失败: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Printf("获取数据库句柄失败: %v", err)
		return db, func() {}
	}

	return db, func() {
		if closeErr := closeDatabase(sqlDB); closeErr != nil {
			log.Printf("数据库关闭失败: %v", closeErr)
		}
	}
}

func closeDatabase(db *sql.DB) error {
	return db.Close()
}
