package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB is the global GORM database instance.
var DB *gorm.DB

// JWTSecret is the HMAC key used to sign and verify JWTs.
// Set via JWT_SECRET environment variable; falls back to a fixed dev default.
var JWTSecret string

// ConnectDatabase opens the MySQL connection and panics if it fails.
// DSN can be overridden via the MYSQL_DSN environment variable.
func ConnectDatabase() {
	// Resolve JWT secret
	JWTSecret = os.Getenv("JWT_SECRET")
	if JWTSecret == "" {
		JWTSecret = "smart-mt-dev-secret-change-in-production"
	}

	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		// Default DSN for local development.
		// Format: user:pass@tcp(host:port)/dbname?params
		dsn = "root:@tcp(127.0.0.1:3306)/maintenance_pnc?charset=utf8mb4&parseTime=True&loc=Local"
	}

	gormLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Gagal koneksi ke database MySQL: %v\n", err)
		log.Fatalf("database connection failed: %v", err)
	}

	// Configure the underlying sql.DB connection pool.
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("failed to get underlying sql.DB: %v", err)
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	DB = db
	fmt.Println("✅ Berhasil terhubung ke database MySQL!")
}
