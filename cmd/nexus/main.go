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
)

func main() {
	configPath := flag.String("config", "config.yaml", "path to config file")
	initAdmin := flag.Bool("init-admin", false, "create initial admin user (interactive)")
	flag.Parse()

	// Load config
	if err := config.Load(*configPath); err != nil {
		log.Fatalf("加载配置文件失败: %v", err)
	}

	// Init JWT
	jwt.Init(config.Global.JWT.Secret)

	// Init database
	if err := database.Init(config.Global.Database); err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}
	log.Println("数据库初始化完成")

	// Create data directory
	dir := filepath.Dir(config.Global.Database.DSN)
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Fatalf("创建数据目录失败: %v", err)
	}

	// Handle init-admin flag
	if *initAdmin {
		if err := database.CreateInitialAdmin(); err != nil {
			log.Fatalf("创建管理员失败: %v", err)
		}
		return
	}

	// Start HTTP server
	r := router.Setup()
	addr := fmt.Sprintf("%s:%d", config.Global.Server.Host, config.Global.Server.Port)
	log.Printf("Nexus Panel 启动在 %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("启动服务器失败: %v", err)
	}
}
