package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"nexus/internal/config"
	"nexus/internal/database"
	"nexus/internal/http/router"
	"nexus/internal/pkg/jwt"
	"nexus/internal/service"
)

const defaultJWTSecret = "change-me-jwt-secret"
const defaultSecretKey = "change-me-to-a-random-string-in-production"

func main() {
	configPath := flag.String("config", "config.yaml", "path to config file")
	initAdminEmail := flag.String("admin-email", "", "create admin user with this email")
	initAdminPass := flag.String("admin-pass", "", "admin password (use with -admin-email)")
	flag.Parse()

if err := config.Load(*configPath); err != nil {
			log.Fatalf("load config failed: %v", err)
		}

		// Security: detect default secret keys
		if config.Global.JWT.Secret == defaultJWTSecret || config.Global.App.SecretKey == defaultSecretKey {
			log.Fatal("[安全] 检测到默认密钥！请修改 config.yaml 中的 jwt.secret 和 app.secret_key 为随机字符串，否则面板存在严重安全风险")
		}

	jwt.Init(config.Global.JWT.Secret)

	if err := database.Init(config.Global.Database); err != nil {
		log.Fatalf("init database failed: %v", err)
	}
	log.Println("database ready")

	if *initAdminEmail != "" && *initAdminPass != "" {
		if err := database.CreateAdminByEmail(*initAdminEmail, *initAdminPass); err != nil {
			log.Fatalf("create admin failed: %v", err)
		}
		return
	}

	dir := filepath.Dir(config.Global.Database.DSN)
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Fatalf("create data dir failed: %v", err)
	}

	r := router.Setup()

	// Start background scheduler (monthly traffic reset etc.)
	service.StartScheduler()

	addr := fmt.Sprintf("%s:%d", config.Global.Server.Host, config.Global.Server.Port)
	log.Printf("Nexus Panel started at %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("start server failed: %v", err)
	}
}