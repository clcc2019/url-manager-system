package services
package services

import (
	"database/sql"
	"url-manager-system/backend/internal/config"
	"url-manager-system/backend/internal/k8s"

	"github.com/redis/go-redis/v9"
)

// Container 服务容器
type Container struct {
	ProjectService  *ProjectService
	URLService      *URLService
	CleanupService  *CleanupService
}

// NewContainer 创建服务容器
func NewContainer(db *sql.DB, redis *redis.Client, k8sClient *k8s.Client, cfg *config.Config) *Container {
	// 创建资源管理器
	resourceManager := k8s.NewResourceManager(k8sClient, cfg.K8s.Namespace)
	ingressManager := k8s.NewIngressManager(k8sClient, cfg.K8s.Namespace, cfg.K8s.IngressClass, cfg.K8s.DefaultDomain)

	// 创建服务实例
	projectService := NewProjectService(db)
	urlService := NewURLService(db, resourceManager, ingressManager, cfg)
	cleanupService := NewCleanupService(db, redis, resourceManager, ingressManager, cfg)

	return &Container{
		ProjectService:  projectService,
		URLService:      urlService,
		CleanupService:  cleanupService,
	}
}