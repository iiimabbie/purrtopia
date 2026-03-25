package database

import (
	"fmt"
	"log"
	"time"

	"purrtopia/internal/config"
	"purrtopia/internal/database/models"
	"purrtopia/internal/database/repository"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB is the global GORM database connection
var DB *gorm.DB

// Repositories (類似 Spring 的 @Autowired Repository)
var (
	ProxySellRepo       *repository.SnowProxySellRepository
	GroupActivityRepo   *repository.GroupActivityRepository
	GroupThreadNameRepo *repository.GroupThreadNameRepository
)

// Connect establishes a connection to the PostgreSQL database using GORM
func Connect(cfg *config.DBConfig) error {
	var err error

	// GORM 設定
	gormConfig := &gorm.Config{
		Logger:                                   logger.Default.LogMode(logger.Silent), // 生產環境靜音
		DisableForeignKeyConstraintWhenMigrating: true,                                  // 避免外鍵衝突
	}

	// 連接資料庫
	DB, err = gorm.Open(postgres.Open(cfg.DSN()), gormConfig)
	if err != nil {
		return fmt.Errorf("failed to connect database: %w", err)
	}

	// 取得底層 sql.DB 來設定連線池
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	// Verify connection
	if err = sqlDB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Database connected successfully (GORM)")

	// Auto migrate (自動建表，類似 JPA ddl-auto)
	if err = DB.AutoMigrate(
		&models.SnowProxySell{},
		&models.GroupActivity{},
		&models.GroupThreadName{},
	); err != nil {
		return fmt.Errorf("failed to auto migrate: %w", err)
	}

	log.Println("Database schema migrated successfully")

	// 初始化 Repositories (類似 Spring Bean)
	ProxySellRepo = repository.NewSnowProxySellRepository(DB)
	GroupActivityRepo = repository.NewGroupActivityRepository(DB)
	GroupThreadNameRepo = repository.NewGroupThreadNameRepository(DB)

	log.Println("Repositories initialized")

	// 植入初始資料（僅首次執行）
	SeedGroupThreadNames()

	return nil
}

// Close closes the database connection
func Close() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}
