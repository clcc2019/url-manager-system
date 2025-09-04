package services

import (
	"context"
	"database/sql"
	"fmt"
	"time"
	"url-manager-system/backend/internal/config"
	"url-manager-system/backend/internal/db/models"
	"url-manager-system/backend/internal/k8s"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

const (
	cleanupLockKey  = "url_manager:cleanup_lock"
	lockTTL         = 30 * time.Minute
	cleanupInterval = 5 * time.Minute
)

// CleanupService 清理服务
type CleanupService struct {
	db              *sql.DB
	redis           *redis.Client
	resourceManager *k8s.ResourceManager
	ingressManager  *k8s.IngressManager
	config          *config.Config
}

// NewCleanupService 创建清理服务
func NewCleanupService(db *sql.DB, redis *redis.Client, resourceManager *k8s.ResourceManager, ingressManager *k8s.IngressManager, cfg *config.Config) *CleanupService {
	return &CleanupService{
		db:              db,
		redis:           redis,
		resourceManager: resourceManager,
		ingressManager:  ingressManager,
		config:          cfg,
	}
}

// StartWorker 启动清理工作线程
func (s *CleanupService) StartWorker() {
	logrus.Info("Starting cleanup worker")

	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()

	// 立即执行一次清理
	s.runCleanup()

	for {
		select {
		case <-ticker.C:
			s.runCleanup()
		}
	}
}

// runCleanup 执行清理操作
func (s *CleanupService) runCleanup() {
	ctx := context.Background()

	// 尝试获取分布式锁
	lock, err := s.acquireLock(ctx)
	if err != nil {
		logrus.WithError(err).Debug("Failed to acquire cleanup lock, skipping")
		return
	}
	if !lock {
		logrus.Debug("Another instance is running cleanup, skipping")
		return
	}
	defer s.releaseLock(ctx)

	logrus.Info("Starting cleanup process")

	// 获取过期的URL
	expiredURLs, err := s.getExpiredURLs(ctx)
	if err != nil {
		logrus.WithError(err).Error("Failed to get expired URLs")
		return
	}

	if len(expiredURLs) == 0 {
		logrus.Debug("No expired URLs found")
		return
	}

	logrus.WithField("count", len(expiredURLs)).Info("Found expired URLs")

	// 清理每个过期的URL
	for _, url := range expiredURLs {
		if err := s.cleanupURL(ctx, &url); err != nil {
			logrus.WithError(err).WithField("url_id", url.ID).Error("Failed to cleanup URL")
		}
	}

	logrus.Info("Cleanup process completed")
}

// acquireLock 获取分布式锁
func (s *CleanupService) acquireLock(ctx context.Context) (bool, error) {
	result, err := s.redis.SetNX(ctx, cleanupLockKey, "locked", lockTTL).Result()
	if err != nil {
		return false, fmt.Errorf("failed to acquire lock: %w", err)
	}
	return result, nil
}

// releaseLock 释放分布式锁
func (s *CleanupService) releaseLock(ctx context.Context) {
	err := s.redis.Del(ctx, cleanupLockKey).Err()
	if err != nil {
		logrus.WithError(err).Error("Failed to release cleanup lock")
	}
}

// getExpiredURLs 获取过期的URL
func (s *CleanupService) getExpiredURLs(ctx context.Context) ([]models.EphemeralURL, error) {
	query := `
		SELECT eu.id, eu.project_id, eu.path, eu.image, eu.env, eu.replicas, eu.resources,
		       eu.status, eu.k8s_deployment_name, eu.k8s_service_name, eu.k8s_secret_name,
		       eu.error_message, eu.expire_at, eu.created_at, eu.updated_at,
		       p.id, p.name, p.description, p.created_at, p.updated_at
		FROM ephemeral_urls eu
		INNER JOIN projects p ON eu.project_id = p.id
		WHERE eu.expire_at <= NOW() 
		  AND eu.status IN ('creating', 'active', 'failed')
		ORDER BY eu.expire_at ASC
		LIMIT 50
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query expired URLs: %w", err)
	}
	defer rows.Close()

	var urls []models.EphemeralURL
	for rows.Next() {
		var url models.EphemeralURL
		url.Project = &models.Project{}

		err := rows.Scan(
			&url.ID, &url.ProjectID, &url.Path, &url.Image, &url.Env, &url.Replicas, &url.Resources,
			&url.Status, &url.K8sDeploymentName, &url.K8sServiceName, &url.K8sSecretName,
			&url.ErrorMessage, &url.ExpireAt, &url.CreatedAt, &url.UpdatedAt,
			&url.Project.ID, &url.Project.Name, &url.Project.Description, &url.Project.CreatedAt, &url.Project.UpdatedAt,
		)
		if err != nil {
			logrus.WithError(err).Error("Failed to scan expired URL")
			continue
		}
		urls = append(urls, url)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating expired URLs: %w", err)
	}

	return urls, nil
}

// cleanupURL 清理单个URL
func (s *CleanupService) cleanupURL(ctx context.Context, url *models.EphemeralURL) error {
	logrus.WithFields(logrus.Fields{
		"url_id": url.ID,
		"path":   url.Path,
		"status": url.Status,
	}).Info("Cleaning up expired URL")

	// 更新状态为删除中
	if err := s.updateURLStatus(ctx, url.ID, models.StatusDeleting, ""); err != nil {
		return fmt.Errorf("failed to update status to deleting: %w", err)
	}

	// 删除Kubernetes资源
	if err := s.deleteKubernetesResources(ctx, url); err != nil {
		logrus.WithError(err).Error("Failed to delete Kubernetes resources")
		// 更新状态为失败，但继续清理数据库记录
		s.updateURLStatus(ctx, url.ID, models.StatusFailed, err.Error())
	}

	// 更新状态为已删除
	if err := s.updateURLStatus(ctx, url.ID, models.StatusDeleted, ""); err != nil {
		return fmt.Errorf("failed to update status to deleted: %w", err)
	}

	logrus.WithField("url_id", url.ID).Info("URL cleanup completed")
	return nil
}

// deleteKubernetesResources 删除Kubernetes资源
func (s *CleanupService) deleteKubernetesResources(ctx context.Context, url *models.EphemeralURL) error {
	var errors []error

	// 从Ingress移除路径
	if url.Project != nil {
		if err := s.ingressManager.RemovePath(ctx, url.Project.Name, url.Path); err != nil {
			logrus.WithError(err).Warn("Failed to remove ingress path")
			errors = append(errors, fmt.Errorf("failed to remove ingress path: %w", err))
		}
	}

	// 删除Deployment
	if url.K8sDeploymentName != nil {
		if err := s.resourceManager.DeleteDeployment(ctx, *url.K8sDeploymentName); err != nil {
			logrus.WithError(err).Warn("Failed to delete deployment")
			errors = append(errors, fmt.Errorf("failed to delete deployment: %w", err))
		}
	}

	// 删除Service
	if url.K8sServiceName != nil {
		if err := s.resourceManager.DeleteService(ctx, *url.K8sServiceName); err != nil {
			logrus.WithError(err).Warn("Failed to delete service")
			errors = append(errors, fmt.Errorf("failed to delete service: %w", err))
		}
	}

	// 删除Secret
	if url.K8sSecretName != nil {
		if err := s.resourceManager.DeleteSecret(ctx, *url.K8sSecretName); err != nil {
			logrus.WithError(err).Warn("Failed to delete secret")
			errors = append(errors, fmt.Errorf("failed to delete secret: %w", err))
		}
	}

	// 如果有错误，返回第一个错误
	if len(errors) > 0 {
		return errors[0]
	}

	return nil
}

// updateURLStatus 更新URL状态
func (s *CleanupService) updateURLStatus(ctx context.Context, id uuid.UUID, status, errorMessage string) error {
	query := `
		UPDATE ephemeral_urls 
		SET status = $2, error_message = $3, updated_at = $4
		WHERE id = $1
	`

	var errMsg *string
	if errorMessage != "" {
		errMsg = &errorMessage
	}

	_, err := s.db.ExecContext(ctx, query, id, status, errMsg, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update URL status: %w", err)
	}

	return nil
}

// ForceCleanupURL 强制清理指定URL（用于手动删除）
func (s *CleanupService) ForceCleanupURL(ctx context.Context, id uuid.UUID) error {
	// 获取URL信息
	url, err := s.getURLWithProject(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get URL: %w", err)
	}

	// 如果已经删除，直接返回
	if url.Status == models.StatusDeleted {
		return nil
	}

	// 执行清理
	return s.cleanupURL(ctx, url)
}

// getURLWithProject 获取URL及其项目信息
func (s *CleanupService) getURLWithProject(ctx context.Context, id uuid.UUID) (*models.EphemeralURL, error) {
	query := `
		SELECT eu.id, eu.project_id, eu.path, eu.image, eu.env, eu.replicas, eu.resources,
		       eu.status, eu.k8s_deployment_name, eu.k8s_service_name, eu.k8s_secret_name,
		       eu.error_message, eu.expire_at, eu.created_at, eu.updated_at,
		       p.id, p.name, p.description, p.created_at, p.updated_at
		FROM ephemeral_urls eu
		INNER JOIN projects p ON eu.project_id = p.id
		WHERE eu.id = $1
	`

	url := &models.EphemeralURL{Project: &models.Project{}}
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&url.ID, &url.ProjectID, &url.Path, &url.Image, &url.Env, &url.Replicas, &url.Resources,
		&url.Status, &url.K8sDeploymentName, &url.K8sServiceName, &url.K8sSecretName,
		&url.ErrorMessage, &url.ExpireAt, &url.CreatedAt, &url.UpdatedAt,
		&url.Project.ID, &url.Project.Name, &url.Project.Description, &url.Project.CreatedAt, &url.Project.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("URL not found")
		}
		return nil, fmt.Errorf("failed to get URL: %w", err)
	}

	return url, nil
}
