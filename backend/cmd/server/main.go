package main

import (
	"fmt"
	"log"
	"url-manager-system/backend/internal/api/routes"
	"url-manager-system/backend/internal/config"
	"url-manager-system/backend/internal/db"
	"url-manager-system/backend/internal/k8s"
	"url-manager-system/backend/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// 设置日志级别
	logrus.SetLevel(logrus.InfoLevel)
	if cfg.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	// 初始化数据库连接
	database, err := db.NewConnection(cfg.Database)
	if err != nil {
		logrus.Fatal("Failed to connect to database:", err)
	}
	defer database.Close()

	// 运行数据库迁移
	var migrationDSN string
	if cfg.Database.URL != "" {
		migrationDSN = cfg.Database.URL
	} else {
		migrationDSN = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
			cfg.Database.User, cfg.Database.Password, cfg.Database.Host,
			cfg.Database.Port, cfg.Database.Database, cfg.Database.SSLMode)
	}
	if err := db.RunMigrations(migrationDSN); err != nil {
		logrus.Fatal("Failed to run migrations:", err)
	}

	// 初始化Redis连接
	redisClient, err := db.NewRedisConnection(cfg.Redis)
	if err != nil {
		logrus.Fatal("Failed to connect to Redis:", err)
	}
	defer redisClient.Close()

	// 初始化Kubernetes客户端（开发环境可选项）
	k8sClient, err := k8s.NewClient()
	if err != nil {
		logrus.Warn("Failed to create Kubernetes client (this is normal in development):", err)
		k8sClient = nil
	}

	// 初始化服务层
	serviceContainer := services.NewContainer(database, redisClient, k8sClient, cfg)

	// 启动清理worker
	go serviceContainer.CleanupService.StartWorker()

	// 设置Gin模式
	if !cfg.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建路由
	router := routes.SetupRoutes(serviceContainer)

	// 启动服务器
	logrus.Infof("Starting server on port %s", cfg.Server.Port)
	if err := router.Run(":" + cfg.Server.Port); err != nil {
		logrus.Fatal("Failed to start server:", err)
	}
}
