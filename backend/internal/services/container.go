package services

import (
	"database/sql"
	"url-manager-system/backend/internal/config"
	"url-manager-system/backend/internal/k8s"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

// Container 服务容器
type Container struct {
	AuthService     *AuthService
	ProjectService  *ProjectService
	URLService      *URLService
	TemplateService *TemplateService
	CleanupService  *CleanupService
}

// StartWorkers 启动所有后台工作线程
func (c *Container) StartWorkers() {
	// 启动清理工作线程
	go c.CleanupService.StartWorker()

	// 启动Pod状态监控工作线程
	go c.URLService.StartPodMonitor()
}

// NewContainer 创建服务容器
func NewContainer(db *sql.DB, redis *redis.Client, k8sClient *k8s.Client, cfg *config.Config) *Container {
	var resourceManager *k8s.ResourceManager
	var ingressManager *k8s.IngressManager

	// 只有在k8sClient不为nil时才创建资源管理器
	if k8sClient != nil {
		resourceManager = k8s.NewResourceManager(k8sClient, cfg.K8s.Namespace)
		ingressManager = k8s.NewIngressManager(k8sClient, cfg.K8s.Namespace, cfg.K8s.IngressClass, cfg.K8s.DefaultDomain)
	}

	// 为 TemplateService 创建 sqlx.DB 实例
	sqlxDB := sqlx.NewDb(db, "postgres")

	// 创建服务实例
	authService := NewAuthService(sqlxDB, cfg.Security.JWTSecret)
	projectService := NewProjectService(db)
	templateService := NewTemplateService(sqlxDB)
	urlService := NewURLService(db, resourceManager, ingressManager, templateService, cfg)
	cleanupService := NewCleanupService(db, redis, resourceManager, ingressManager, cfg)

	return &Container{
		AuthService:     authService,
		ProjectService:  projectService,
		URLService:      urlService,
		TemplateService: templateService,
		CleanupService:  cleanupService,
	}
}
