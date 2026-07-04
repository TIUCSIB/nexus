package database

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"nexus/internal/config"
	"nexus/internal/model"
	"nexus/internal/pkg/crypto"

	"github.com/google/uuid"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Init(cfg config.DatabaseConfig) error {
	dir := filepath.Dir(cfg.DSN)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create data directory: %w", err)
	}

	var err error
	logLevel := logger.Warn
	if config.Global.App.Debug {
		logLevel = logger.Info
	}
	DB, err = gorm.Open(sqlite.Open(cfg.DSN), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return err
	}

	DB.Exec("PRAGMA journal_mode=WAL")
	DB.Exec("PRAGMA busy_timeout=5000")

	if err := DB.AutoMigrate(
		&model.User{},
		&model.Plan{},
		&model.Node{},
		&model.ServerGroup{},
		&model.TrafficLog{},
		&model.SystemConfig{},
		&model.AliveIP{},
		&model.RouteRule{},
	); err != nil {
		return err
	}

	// Initialize default server_token if not exists
	initDefaultServerToken()

	return nil
}

func initDefaultServerToken() {
	var cfg model.SystemConfig
	if err := DB.Where("key = ?", "server_token").First(&cfg).Error; err != nil {
		// Not found, create default
		DB.Create(&model.SystemConfig{
			Key:   "server_token",
			Value: uuid.New().String(),
		})
		fmt.Printf("Generated default server_token: %s\n", GetSetting("server_token"))
	}
}

func CreateInitialAdmin() error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("\u8bf7\u8f93\u5165\u7ba1\u7406\u5458\u90ae\u7bb1: ")
	email, _ := reader.ReadString('\n')
	email = strings.TrimSpace(email)

	fmt.Print("\u8bf7\u8f93\u5165\u7ba1\u7406\u5458\u5bc6\u7801: ")
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)

	if email == "" || password == "" {
		return fmt.Errorf("email and password cannot be empty")
	}

	hash, err := crypto.HashPassword(password)
	if err != nil {
		return err
	}

	admin := model.User{
		UUID:         uuid.New().String(),
		Email:        email,
		PasswordHash: hash,
		IsAdmin:      true,
		Token:        uuid.New().String(),
		Status:       1,
		TrafficLimit: 0,
	}

	result := DB.Where("email = ?", email).FirstOrCreate(&admin)
	if result.Error != nil {
		return result.Error
	}

	_ = time.Now()
	fmt.Printf("admin %s created successfully\n", email)
	return nil
}

func CreateAdminByEmail(email, password string) error {
	hash, err := crypto.HashPassword(password)
	if err != nil {
		return err
	}

	admin := model.User{
		UUID:         uuid.New().String(),
		Email:        email,
		PasswordHash: hash,
		IsAdmin:      true,
		Token:        uuid.New().String(),
		Status:       1,
		TrafficLimit: 0,
	}

	result := DB.Where("email = ?", email).FirstOrCreate(&admin)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		fmt.Printf("admin %s already exists\n", email)
	} else {
		fmt.Printf("admin %s created successfully\n", email)
	}
	return nil
}
