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
	"gorm.io/driver/sqlite"
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

	return DB.AutoMigrate(
		&model.User{},
		&model.Plan{},
		&model.Node{},
		&model.TrafficLog{},
		&model.SystemConfig{},
		&model.AliveIP{},
		&model.RouteRule{},
	)
}

func CreateInitialAdmin() error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("请输入管理员邮箱: ")
	email, _ := reader.ReadString('\n')
	email = strings.TrimSpace(email)

	fmt.Printf("请输入管理员密码: ")
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)

	if email == "" || password == "" {
		return fmt.Errorf("邮箱和密码不能为空")
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
	fmt.Printf("管理员账号 %s 创建成功\n", email)
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