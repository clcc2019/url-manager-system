package services

import (
	"context"
	"crypto/md5"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"url-manager-system/backend/internal/config"
	"url-manager-system/backend/internal/db/models"
	"url-manager-system/backend/internal/k8s"
	"url-manager-system/backend/internal/utils"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

const (
	// 用于生成随机路径的字符集
	pathChars  = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	pathLength = 8
)

// URLService URL服务
type URLService struct {
	db              *sql.DB
	resourceManager *k8s.ResourceManager
	ingressManager  *k8s.IngressManager
	templateService *TemplateService
	config          *config.Config
}

// NewURLService 创建URL服务
func NewURLService(db *sql.DB, resourceManager *k8s.ResourceManager, ingressManager *k8s.IngressManager, templateService *TemplateService, cfg *config.Config) *URLService {
	return &URLService{
		db:              db,
		resourceManager: resourceManager,
		ingressManager:  ingressManager,
		templateService: templateService,
		config:          cfg,
	}
}

// CreateEphemeralURL 创建临时URL
func (s *URLService) CreateEphemeralURL(ctx context.Context, projectID uuid.UUID, req *models.CreateEphemeralURLRequest) (*models.CreateEphemeralURLResponse, error) {
	// 验证请求
	if err := s.validateCreateRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// 验证容器配置
	if err := utils.ValidateContainerConfig(req.ContainerConfig); err != nil {
		return nil, fmt.Errorf("container config validation failed: %w", err)
	}

	// 验证设备映射安全性
	for _, device := range req.ContainerConfig.Devices {
		if err := utils.ValidateDeviceMapping(device); err != nil {
			return nil, fmt.Errorf("device mapping validation failed: %w", err)
		}
	}

	// 获取项目信息
	project, err := s.getProject(ctx, projectID)
	if err != nil {
		return nil, err
	}

	// 生成随机路径
	path, err := s.generateUniquePath(ctx, projectID, req.Image)
	if err != nil {
		return nil, fmt.Errorf("failed to generate unique path: %w", err)
	}

	// 创建URL记录
	url := &models.EphemeralURL{
		ID:              uuid.New(),
		ProjectID:       projectID,
		Path:            path,
		Image:           req.Image,
		Env:             req.Env,
		Replicas:        req.Replicas,
		Resources:       req.Resources,
		ContainerConfig: req.ContainerConfig,
		Status:          models.StatusCreating,
		TTLSeconds:      req.TTLSeconds,  // 保存TTL值
		IngressHost:     req.IngressHost, // 保存自定义ingress host
		// 注意：只有在active状态时才开始计算过期时间
		ExpireAt:  time.Now().Add(365 * 24 * time.Hour), // 设置为1年后的时间，在active前不会过期
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Logs: []models.LogEntry{
			{
				Timestamp: time.Now(),
				Level:     "info",
				Message:   "URL创建请求已提交",
				Details:   fmt.Sprintf("镜像: %s, 副本数: %d", req.Image, req.Replicas),
			},
		},
	}

	// 设置默认值
	if url.Replicas == 0 {
		url.Replicas = 1
	}
	if url.Resources.Requests.CPU == "" {
		url.Resources.Requests.CPU = "100m"
	}
	if url.Resources.Requests.Memory == "" {
		url.Resources.Requests.Memory = "128Mi"
	}
	if url.Resources.Limits.CPU == "" {
		url.Resources.Limits.CPU = s.config.Security.DefaultCPULimit
	}
	if url.Resources.Limits.Memory == "" {
		url.Resources.Limits.Memory = s.config.Security.DefaultMemLimit
	}

	// 生成K8s资源名称
	url.K8sDeploymentName = stringPtr(fmt.Sprintf("ephemeral-%s", url.ID.String()[:8]))
	url.K8sServiceName = stringPtr(fmt.Sprintf("svc-ephemeral-%s", url.ID.String()[:8]))

	// 如果有环境变量，创建Secret名称
	if len(url.Env) > 0 {
		url.K8sSecretName = stringPtr(fmt.Sprintf("secret-ephemeral-%s", url.ID.String()[:8]))
	}

	// 开始事务处理
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 插入数据库记录
	if err := s.insertURLRecord(ctx, tx, url); err != nil {
		return nil, fmt.Errorf("failed to insert URL record: %w", err)
	}

	// 检查Kubernetes资源管理器是否可用
	if s.resourceManager == nil || s.ingressManager == nil {
		// Kubernetes不可用，设置为draft状态
		s.updateURLStatus(ctx, url.ID, "draft", "Kubernetes not available")
		logrus.Warn("Kubernetes not available, URL created in draft mode")
	} else {
		// Kubernetes可用，创建资源并设置为waiting状态
		if err := s.createKubernetesResources(ctx, url, project.Name); err != nil {
			logrus.WithError(err).Error("Failed to create Kubernetes resources")
			// 更新状态为失败
			s.updateURLStatus(ctx, url.ID, models.StatusFailed, err.Error())
			return nil, fmt.Errorf("failed to create Kubernetes resources: %w", err)
		}
		// 资源创建成功后，设置为等待状态（等待Pod Ready）
		s.updateURLStatus(ctx, url.ID, models.StatusWaiting, "")
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// 异步验证部署状态
	go s.verifyDeployment(url)

	// 构建返回URL
	fullURL := fmt.Sprintf("https://%s%s", s.config.K8s.DefaultDomain, path)

	logrus.WithFields(logrus.Fields{
		"url_id":     url.ID,
		"project_id": projectID,
		"path":       path,
	}).Info("Ephemeral URL created successfully")

	return &models.CreateEphemeralURLResponse{
		URL: fullURL,
		ID:  url.ID,
	}, nil
}

// DeployURL 部署URL到Kubernetes集群
func (s *URLService) DeployURL(ctx context.Context, urlID uuid.UUID) error {
	// 获取URL信息
	url, err := s.GetEphemeralURL(ctx, urlID)
	if err != nil {
		return fmt.Errorf("failed to get URL: %w", err)
	}

	// 检查状态 - 允许draft、failed、active状态进行部署（active状态用于更新）
	allowedStatuses := []string{"draft", models.StatusFailed, models.StatusActive, models.StatusWaiting, models.StatusCreating}
	isAllowed := false
	for _, status := range allowedStatuses {
		if url.Status == status {
			isAllowed = true
			break
		}
	}

	if !isAllowed {
		return fmt.Errorf("URL is not in deployable status, current status: %s", url.Status)
	}

	// 获取项目信息
	var projectName string
	err = s.db.QueryRowContext(ctx, "SELECT name FROM projects WHERE id = $1", url.ProjectID).Scan(&projectName)
	if err != nil {
		return fmt.Errorf("failed to get project name: %w", err)
	}

	// 更新状态为创建中（如果不是active状态）
	if url.Status != models.StatusActive {
		s.updateURLStatus(ctx, url.ID, models.StatusCreating, "")
	}

	// 创建或更新Kubernetes资源
	if err := s.createKubernetesResources(ctx, url, projectName); err != nil {
		logrus.WithError(err).Error("Failed to deploy Kubernetes resources")
		s.updateURLStatus(ctx, url.ID, models.StatusFailed, err.Error())
		return fmt.Errorf("failed to deploy Kubernetes resources: %w", err)
	}

	// 更新部署状态
	_, err = s.db.ExecContext(ctx,
		"UPDATE ephemeral_urls SET deployed = true, deployment_requested_at = NOW() WHERE id = $1",
		url.ID)
	if err != nil {
		logrus.WithError(err).Error("Failed to update deployment status")
	}

	// 如果原来是active状态，保持active状态；否则设置为waiting状态
	if url.Status == models.StatusActive {
		logrus.WithField("url_id", url.ID).Info("URL updated successfully")
	} else {
		s.updateURLStatus(ctx, url.ID, models.StatusWaiting, "")
		logrus.WithField("url_id", url.ID).Info("URL deployed successfully")
	}

	return nil
}

// GetEphemeralURL 获取临时URL
func (s *URLService) GetEphemeralURL(ctx context.Context, id uuid.UUID) (*models.EphemeralURL, error) {
	query := `
		SELECT eu.id, eu.project_id, eu.path, eu.image, eu.env, eu.replicas, eu.resources,
		       eu.container_config, eu.status, eu.k8s_deployment_name, eu.k8s_service_name, eu.k8s_secret_name,
		       eu.error_message, eu.expire_at, eu.created_at, eu.updated_at,
		       p.id, p.name, p.description, p.created_at, p.updated_at
		FROM ephemeral_urls eu
		INNER JOIN projects p ON eu.project_id = p.id
		WHERE eu.id = $1
	`

	url := &models.EphemeralURL{Project: &models.Project{}}
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&url.ID, &url.ProjectID, &url.Path, &url.Image, &url.Env, &url.Replicas, &url.Resources,
		&url.ContainerConfig, &url.Status, &url.K8sDeploymentName, &url.K8sServiceName, &url.K8sSecretName,
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

// UpdateEphemeralURL 更新临时URL
func (s *URLService) UpdateEphemeralURL(ctx context.Context, id uuid.UUID, req *models.UpdateEphemeralURLRequest) (*models.EphemeralURL, error) {
	logrus.WithFields(logrus.Fields{
		"url_id":  id.String(),
		"request": req,
	}).Info("Starting URL update")

	// 获取现有URL
	existingURL, err := s.GetEphemeralURL(ctx, id)
	if err != nil {
		logrus.WithError(err).WithField("url_id", id.String()).Error("Failed to get existing URL")
		return nil, err
	}

	logrus.WithFields(logrus.Fields{
		"url_id":            id.String(),
		"existing_image":    existingURL.Image,
		"existing_replicas": existingURL.Replicas,
		"new_image":         req.Image,
		"new_replicas":      req.Replicas,
	}).Info("Current vs new values")

	// 验证更新请求
	if err := s.validateUpdateRequest(req); err != nil {
		logrus.WithError(err).WithField("url_id", id.String()).Error("Validation failed")
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// 构建更新语句
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1

	if req.Image != "" {
		setParts = append(setParts, fmt.Sprintf("image = $%d", argIndex))
		args = append(args, req.Image)
		argIndex++
	}

	if req.Env != nil {
		setParts = append(setParts, fmt.Sprintf("env = $%d", argIndex))
		args = append(args, req.Env)
		argIndex++
	}

	if req.TTLSeconds > 0 {
		setParts = append(setParts, fmt.Sprintf("ttl_seconds = $%d", argIndex))
		args = append(args, req.TTLSeconds)
		argIndex++
	}

	if req.Replicas > 0 {
		setParts = append(setParts, fmt.Sprintf("replicas = $%d", argIndex))
		args = append(args, req.Replicas)
		argIndex++
	}

	if req.Resources.Requests.CPU != "" || req.Resources.Requests.Memory != "" ||
		req.Resources.Limits.CPU != "" || req.Resources.Limits.Memory != "" {
		setParts = append(setParts, fmt.Sprintf("resources = $%d", argIndex))
		args = append(args, req.Resources)
		argIndex++
	}

	// 总是更新container_config，因为前端会发送完整的配置
	// 这样可以支持清空字段（如设置command为空数组）
	setParts = append(setParts, fmt.Sprintf("container_config = $%d", argIndex))
	args = append(args, req.ContainerConfig)
	argIndex++

	logrus.WithFields(logrus.Fields{
		"url_id":           id.String(),
		"container_config": req.ContainerConfig,
	}).Info("Updating container_config")

	if req.IngressHost != nil {
		setParts = append(setParts, fmt.Sprintf("ingress_host = $%d", argIndex))
		args = append(args, *req.IngressHost)
		argIndex++
	}

	if len(setParts) == 0 {
		return nil, fmt.Errorf("no fields to update")
	}

	// 添加更新时间
	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	// 执行更新
	query := fmt.Sprintf("UPDATE ephemeral_urls SET %s WHERE id = $%d",
		strings.Join(setParts, ", "), argIndex)
	args = append(args, id)

	logrus.WithFields(logrus.Fields{
		"url_id": id.String(),
		"query":  query,
		"args":   args,
	}).Info("Executing update query")

	_, err = s.db.ExecContext(ctx, query, args...)
	if err != nil {
		logrus.WithError(err).WithField("url_id", id.String()).Error("Failed to execute update query")
		return nil, fmt.Errorf("failed to update URL: %w", err)
	}

	logrus.WithField("url_id", id.String()).Info("Database update completed successfully")

	// 如果状态为active且TTL被更新，重新计算过期时间
	if existingURL.Status == models.StatusActive && req.TTLSeconds > 0 {
		newExpireAt := time.Now().Add(time.Duration(req.TTLSeconds) * time.Second)
		err = s.updateURLExpireAt(ctx, id, newExpireAt)
		if err != nil {
			logrus.WithError(err).WithField("url_id", id).Error("Failed to update expire time")
		}
	}

	// 获取更新后的URL
	updatedURL, err := s.GetEphemeralURL(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated URL: %w", err)
	}

	logrus.WithField("url_id", id).Info("URL updated successfully")
	return updatedURL, nil
}

// GetURLContainerStatus 获取URL容器状态
func (s *URLService) GetURLContainerStatus(ctx context.Context, id uuid.UUID) ([]*models.ContainerStatus, error) {
	// 获取URL信息
	url, err := s.GetEphemeralURL(ctx, id)
	if err != nil {
		return nil, err
	}

	// 检查Kubernetes资源管理器是否可用
	if s.resourceManager == nil {
		return nil, fmt.Errorf("Kubernetes client not available")
	}

	// 检查URL是否有部署名称
	if url.K8sDeploymentName == nil {
		return []*models.ContainerStatus{}, nil
	}

	// 获取容器状态
	statuses, err := s.resourceManager.GetContainerStatus(ctx, *url.K8sDeploymentName)
	if err != nil {
		logrus.WithError(err).WithField("url_id", id).Error("Failed to get container status")
		return nil, fmt.Errorf("failed to get container status: %w", err)
	}

	return statuses, nil
}

// GetURLPodEvents 获取URL Pod事件
func (s *URLService) GetURLPodEvents(ctx context.Context, id uuid.UUID) ([]*models.PodEvent, error) {
	// 获取URL信息
	url, err := s.GetEphemeralURL(ctx, id)
	if err != nil {
		return nil, err
	}

	// 检查Kubernetes资源管理器是否可用
	if s.resourceManager == nil {
		return nil, fmt.Errorf("Kubernetes client not available")
	}

	// 检查URL是否有部署名称
	if url.K8sDeploymentName == nil {
		return []*models.PodEvent{}, nil
	}

	// 获取Pod事件
	events, err := s.resourceManager.GetPodEvents(ctx, *url.K8sDeploymentName)
	if err != nil {
		logrus.WithError(err).WithField("url_id", id).Error("Failed to get pod events")
		return nil, fmt.Errorf("failed to get pod events: %w", err)
	}

	return events, nil
}

// GetURLContainerLogs 获取URL容器日志
func (s *URLService) GetURLContainerLogs(ctx context.Context, id uuid.UUID, containerName string, lines int) ([]*models.ContainerLog, error) {
	// 获取URL信息
	url, err := s.GetEphemeralURL(ctx, id)
	if err != nil {
		return nil, err
	}

	// 检查Kubernetes资源管理器是否可用
	if s.resourceManager == nil {
		return nil, fmt.Errorf("Kubernetes client not available")
	}

	// 检查URL是否有部署名称
	if url.K8sDeploymentName == nil {
		return []*models.ContainerLog{}, nil
	}

	// 获取容器日志
	logs, err := s.resourceManager.GetContainerLogs(ctx, *url.K8sDeploymentName, containerName, lines)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"url_id":    id,
			"container": containerName,
			"lines":     lines,
		}).Error("Failed to get container logs")
		return nil, fmt.Errorf("failed to get container logs: %w", err)
	}

	return logs, nil
}

// validateUpdateRequest 验证更新请求
func (s *URLService) validateUpdateRequest(req *models.UpdateEphemeralURLRequest) error {
	if req.TTLSeconds > 0 && (req.TTLSeconds < 60 || req.TTLSeconds > 604800) {
		return fmt.Errorf("TTL seconds must be between 60 and 604800")
	}
	if req.Replicas > 0 && (req.Replicas < 1 || req.Replicas > 10) {
		return fmt.Errorf("replicas must be between 1 and 10")
	}
	return nil
}

// updateURLExpireAt 更新URL过期时间
func (s *URLService) updateURLExpireAt(ctx context.Context, id uuid.UUID, expireAt time.Time) error {
	query := "UPDATE ephemeral_urls SET expire_at = $1, updated_at = $2 WHERE id = $3"
	_, err := s.db.ExecContext(ctx, query, expireAt, time.Now(), id)
	return err
}

// ListEphemeralURLs 列出项目的临时URL
func (s *URLService) ListEphemeralURLs(ctx context.Context, projectID uuid.UUID, limit, offset int) ([]models.EphemeralURL, int, error) {
	// 获取总数
	var total int
	countQuery := `SELECT COUNT(*) FROM ephemeral_urls WHERE project_id = $1`
	err := s.db.QueryRowContext(ctx, countQuery, projectID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count URLs: %w", err)
	}

	// 获取URL列表
	query := `
		SELECT id, project_id, path, image, env, replicas, resources,
		       status, k8s_deployment_name, k8s_service_name, k8s_secret_name,
		       error_message, expire_at, created_at, updated_at
		FROM ephemeral_urls
		WHERE project_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := s.db.QueryContext(ctx, query, projectID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list URLs: %w", err)
	}
	defer rows.Close()

	var urls []models.EphemeralURL
	for rows.Next() {
		var url models.EphemeralURL
		err := rows.Scan(
			&url.ID, &url.ProjectID, &url.Path, &url.Image, &url.Env, &url.Replicas, &url.Resources,
			&url.Status, &url.K8sDeploymentName, &url.K8sServiceName, &url.K8sSecretName,
			&url.ErrorMessage, &url.ExpireAt, &url.CreatedAt, &url.UpdatedAt,
		)
		if err != nil {
			logrus.WithError(err).Error("Failed to scan URL")
			continue
		}
		urls = append(urls, url)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating URLs: %w", err)
	}

	return urls, total, nil
}

// DeleteEphemeralURL 删除临时URL
func (s *URLService) DeleteEphemeralURL(ctx context.Context, id uuid.UUID) error {
	// 获取URL信息
	url, err := s.GetEphemeralURL(ctx, id)
	if err != nil {
		return err
	}

	// 更新状态为删除中
	if err := s.updateURLStatus(ctx, id, models.StatusDeleting, ""); err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	// 删除Kubernetes资源
	if err := s.deleteKubernetesResources(ctx, url); err != nil {
		logrus.WithError(err).Error("Failed to delete Kubernetes resources")
		// 更新状态为失败
		s.updateURLStatus(ctx, id, models.StatusFailed, err.Error())
		return fmt.Errorf("failed to delete Kubernetes resources: %w", err)
	}

	// 更新状态为已删除
	if err := s.updateURLStatus(ctx, id, models.StatusDeleted, ""); err != nil {
		return fmt.Errorf("failed to update final status: %w", err)
	}

	logrus.WithField("url_id", id).Info("Ephemeral URL deleted successfully")
	return nil
}

// validateCreateRequest 验证创建请求
func (s *URLService) validateCreateRequest(req *models.CreateEphemeralURLRequest) error {
	// 验证镜像格式
	if !utils.ValidateImageName(req.Image) {
		return fmt.Errorf("invalid image name format: %s", req.Image)
	}

	// 验证镜像白名单
	// if !s.isImageAllowed(req.Image) {
	// 	return fmt.Errorf("image %s is not in allowed list", req.Image)
	// }

	// 验证副本数
	if req.Replicas > s.config.Security.MaxReplicas {
		return fmt.Errorf("replicas cannot exceed %d", s.config.Security.MaxReplicas)
	}

	// 验证TTL
	if req.TTLSeconds > s.config.Security.MaxTTLSeconds {
		return fmt.Errorf("TTL cannot exceed %d seconds", s.config.Security.MaxTTLSeconds)
	}

	// 验证环境变量
	for _, env := range req.Env {
		if !utils.ValidateEnvironmentVariableName(env.Name) {
			return fmt.Errorf("invalid environment variable name: %s", env.Name)
		}
		// 清理环境变量值
		env.Value = utils.SanitizeInput(env.Value)
	}

	// 验证资源配置
	if req.Resources.Requests.CPU != "" && !utils.ValidateResourceString(req.Resources.Requests.CPU) {
		return fmt.Errorf("invalid CPU request format: %s", req.Resources.Requests.CPU)
	}
	if req.Resources.Requests.Memory != "" && !utils.ValidateResourceString(req.Resources.Requests.Memory) {
		return fmt.Errorf("invalid memory request format: %s", req.Resources.Requests.Memory)
	}
	if req.Resources.Limits.CPU != "" && !utils.ValidateResourceString(req.Resources.Limits.CPU) {
		return fmt.Errorf("invalid CPU limit format: %s", req.Resources.Limits.CPU)
	}
	if req.Resources.Limits.Memory != "" && !utils.ValidateResourceString(req.Resources.Limits.Memory) {
		return fmt.Errorf("invalid memory limit format: %s", req.Resources.Limits.Memory)
	}

	return nil
}

// isImageAllowed 检查镜像是否在白名单中
func (s *URLService) isImageAllowed(image string) bool {
	for _, allowedImage := range s.config.Security.AllowedImages {
		if strings.HasPrefix(image, allowedImage) {
			return true
		}
	}
	return false
}

// generateHashPath 生成基于哈希的路径
func (s *URLService) generateHashPath(projectID uuid.UUID, image string) string {
	// 使用项目ID、镜像名和时间戳生成哈希
	data := fmt.Sprintf("%s-%s-%d", projectID.String(), image, time.Now().UnixNano())
	hash := md5.Sum([]byte(data))
	// 取前8位作为路径
	return fmt.Sprintf("/%x", hash[:4])
}

// generateUniquePath 生成唯一路径
func (s *URLService) generateUniquePath(ctx context.Context, projectID uuid.UUID, image string) (string, error) {
	maxRetries := 10
	for i := 0; i < maxRetries; i++ {
		path := s.generateHashPath(projectID, image)

		// 检查路径是否已存在
		var count int
		query := `SELECT COUNT(*) FROM ephemeral_urls WHERE project_id = $1 AND path = $2`
		err := s.db.QueryRowContext(ctx, query, projectID, path).Scan(&count)
		if err != nil {
			return "", err
		}

		if count == 0 {
			return path, nil
		}
	}

	return "", fmt.Errorf("failed to generate unique path after %d retries", maxRetries)
}

// getProject 获取项目信息
func (s *URLService) getProject(ctx context.Context, projectID uuid.UUID) (*models.Project, error) {
	project := &models.Project{}
	query := `SELECT id, name, description, created_at, updated_at FROM projects WHERE id = $1`
	err := s.db.QueryRowContext(ctx, query, projectID).Scan(
		&project.ID, &project.Name, &project.Description, &project.CreatedAt, &project.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("project not found")
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}
	return project, nil
}

// insertURLRecord 插入URL记录
func (s *URLService) insertURLRecord(ctx context.Context, tx *sql.Tx, url *models.EphemeralURL) error {
	query := `
		INSERT INTO ephemeral_urls (
			id, project_id, path, image, env, replicas, resources, status, ttl_seconds,
			k8s_deployment_name, k8s_service_name, k8s_secret_name,
			expire_at, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
		)
	`

	_, err := tx.ExecContext(ctx, query,
		url.ID, url.ProjectID, url.Path, url.Image, url.Env, url.Replicas, url.Resources, url.Status, url.TTLSeconds,
		url.K8sDeploymentName, url.K8sServiceName, url.K8sSecretName,
		url.ExpireAt, url.CreatedAt, url.UpdatedAt,
	)

	return err
}

// createKubernetesResources 创建Kubernetes资源
func (s *URLService) createKubernetesResources(ctx context.Context, url *models.EphemeralURL, projectName string) error {
	// 检查Kubernetes资源管理器是否可用
	if s.resourceManager == nil || s.ingressManager == nil {
		logrus.Warn("Kubernetes managers not available, skipping Kubernetes resource creation")
		return nil
	}

	// 创建Secret（如果需要）
	if url.K8sSecretName != nil {
		if err := s.resourceManager.CreateSecret(ctx, url); err != nil {
			return fmt.Errorf("failed to create secret: %w", err)
		}
	}

	// 创建或更新Deployment
	if err := s.resourceManager.CreateOrUpdateDeployment(ctx, url); err != nil {
		return fmt.Errorf("failed to create or update deployment: %w", err)
	}

	// 创建或更新Service
	if err := s.resourceManager.CreateOrUpdateService(ctx, url); err != nil {
		return fmt.Errorf("failed to create or update service: %w", err)
	}

	// 添加Ingress路径
	if err := s.ingressManager.AddPath(ctx, url, projectName); err != nil {
		return fmt.Errorf("failed to add ingress path: %w", err)
	}

	return nil
}

// deleteKubernetesResources 删除Kubernetes资源
func (s *URLService) deleteKubernetesResources(ctx context.Context, url *models.EphemeralURL) error {
	// 从Ingress移除路径
	if err := s.ingressManager.RemovePath(ctx, url.Project.Name, url.Path); err != nil {
		logrus.WithError(err).Warn("Failed to remove ingress path")
	}

	// 删除Deployment
	if url.K8sDeploymentName != nil {
		if err := s.resourceManager.DeleteDeployment(ctx, *url.K8sDeploymentName); err != nil {
			logrus.WithError(err).Warn("Failed to delete deployment")
		}
	}

	// 删除Service
	if url.K8sServiceName != nil {
		if err := s.resourceManager.DeleteService(ctx, *url.K8sServiceName); err != nil {
			logrus.WithError(err).Warn("Failed to delete service")
		}
	}

	// 删除Secret
	if url.K8sSecretName != nil {
		if err := s.resourceManager.DeleteSecret(ctx, *url.K8sSecretName); err != nil {
			logrus.WithError(err).Warn("Failed to delete secret")
		}
	}

	return nil
}

// updateURLStatus 更新URL状态
func (s *URLService) updateURLStatus(ctx context.Context, id uuid.UUID, status, errorMessage string) error {
	// 获取当前URL信息
	url, err := s.GetEphemeralURL(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get URL: %w", err)
	}

	// 构建更新查询
	var query string
	var args []interface{}

	if status == models.StatusActive {
		// 当状态变为active时，开始计算过期时间
		newExpireAt := time.Now().Add(time.Duration(url.TTLSeconds) * time.Second)
		query = `
			UPDATE ephemeral_urls
			SET status = $2, error_message = $3, started_at = NOW(), expire_at = $4, updated_at = NOW(),
				logs = logs || $5::jsonb
			WHERE id = $1
		`
		logEntry := models.LogEntry{
			Timestamp: time.Now(),
			Level:     "info",
			Message:   "URL已成功部署并开始运行",
			Details:   fmt.Sprintf("过期时间: %s", newExpireAt.Format("2006-01-02 15:04:05")),
		}
		logsJSON, _ := json.Marshal([]models.LogEntry{logEntry})

		var errMsg *string
		if errorMessage != "" {
			errMsg = &errorMessage
		}
		args = []interface{}{id, status, errMsg, newExpireAt, string(logsJSON)}
	} else {
		// 其他状态更新
		query = `
			UPDATE ephemeral_urls
			SET status = $2, error_message = $3, updated_at = $4,
				logs = logs || $5::jsonb
			WHERE id = $1
		`

		var logEntry models.LogEntry
		switch status {
		case models.StatusFailed:
			logEntry = models.LogEntry{
				Timestamp: time.Now(),
				Level:     "error",
				Message:   "URL部署失败",
				Details:   errorMessage,
			}
		case models.StatusDeleting:
			logEntry = models.LogEntry{
				Timestamp: time.Now(),
				Level:     "info",
				Message:   "开始删除URL",
			}
		case models.StatusDeleted:
			logEntry = models.LogEntry{
				Timestamp: time.Now(),
				Level:     "info",
				Message:   "URL已成功删除",
			}
		default:
			logEntry = models.LogEntry{
				Timestamp: time.Now(),
				Level:     "info",
				Message:   fmt.Sprintf("状态更新为: %s", status),
			}
		}
		logsJSON, _ := json.Marshal([]models.LogEntry{logEntry})

		var errMsg *string
		if errorMessage != "" {
			errMsg = &errorMessage
		}
		args = []interface{}{id, status, errMsg, time.Now(), string(logsJSON)}
	}

	_, err = s.db.ExecContext(ctx, query, args...)
	return err
}

// verifyDeployment 异步验证部署状态
func (s *URLService) verifyDeployment(url *models.EphemeralURL) {
	ctx := context.Background()
	timeout := time.After(5 * time.Minute)
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	logrus.WithFields(logrus.Fields{
		"url_id":          url.ID,
		"deployment_name": url.K8sDeploymentName,
	}).Info("Starting deployment verification")

	for {
		select {
		case <-timeout:
			logrus.WithField("url_id", url.ID).Warn("Deployment verification timeout")
			s.updateURLStatus(ctx, url.ID, models.StatusFailed, "deployment verification timeout")
			return
		case <-ticker.C:
			if url.K8sDeploymentName != nil {
				ready, err := s.resourceManager.CheckDeploymentReady(ctx, *url.K8sDeploymentName)
				if err != nil {
					// 检查是否是deployment不存在的错误
					if strings.Contains(err.Error(), "not found") {
						logrus.WithFields(logrus.Fields{
							"url_id":          url.ID,
							"deployment_name": *url.K8sDeploymentName,
							"error":           err.Error(),
						}).Error("Deployment not found, marking as failed")
						s.updateURLStatus(ctx, url.ID, models.StatusFailed, fmt.Sprintf("deployment not found: %s", err.Error()))
						return
					}
					logrus.WithError(err).Error("Failed to check deployment status")
					continue
				}

				if ready {
					logrus.WithField("url_id", url.ID).Info("Deployment is ready, marking as active")
					s.updateURLStatus(ctx, url.ID, models.StatusActive, "")
					return
				}
			}
		}
	}
}

// stringPtr 辅助函数，返回字符串指针
func stringPtr(s string) *string {
	return &s
}

// CreateEphemeralURLFromTemplate 基于模版创建临时URL
func (s *URLService) CreateEphemeralURLFromTemplate(ctx context.Context, projectID uuid.UUID, req *models.CreateEphemeralURLFromTemplateRequest) (*models.CreateEphemeralURLResponse, error) {
	// 获取项目信息
	project, err := s.getProject(ctx, projectID)
	if err != nil {
		return nil, err
	}

	// 生成路径（优先使用用户指定的路径）
	var path string
	if req.Path != "" {
		// 验证自定义路径格式
		if !strings.HasPrefix(req.Path, "/") {
			req.Path = "/" + req.Path
		}
		// 检查路径是否已存在
		var count int
		query := `SELECT COUNT(*) FROM ephemeral_urls WHERE project_id = $1 AND path = $2`
		err := s.db.QueryRowContext(ctx, query, projectID, req.Path).Scan(&count)
		if err != nil {
			return nil, fmt.Errorf("failed to check path uniqueness: %w", err)
		}
		if count > 0 {
			return nil, fmt.Errorf("path '%s' already exists in this project", req.Path)
		}
		path = req.Path
	} else {
		// 生成随机路径
		path, err = s.generateUniquePath(ctx, projectID, fmt.Sprintf("template-%s", req.TemplateID.String()))
		if err != nil {
			return nil, fmt.Errorf("failed to generate unique path: %w", err)
		}
	}

	// 生成全局唯一的资源名称
	baseID := uuid.New().String()[:8]

	// 生成模版变量
	variables := map[string]string{
		"PATH":            strings.TrimPrefix(path, "/"),
		"SERVICE_NAME":    fmt.Sprintf("svc-ephemeral-%s", baseID),
		"DEPLOYMENT_NAME": fmt.Sprintf("ephemeral-%s", baseID),
		"PROJECT_NAME":    project.Name,
		"UUID":            baseID,
	}

	// 获取模板信息
	template, err := s.templateService.GetTemplate(ctx, req.TemplateID)
	if err != nil {
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	// 处理模版，获取处理后的YAML
	processedYAML, err := s.templateService.ProcessTemplate(ctx, req.TemplateID, variables)
	if err != nil {
		return nil, fmt.Errorf("failed to process template: %w", err)
	}

	// 从模板解析规格创建URL记录
	url := &models.EphemeralURL{
		ID:         uuid.New(),
		ProjectID:  projectID,
		TemplateID: &req.TemplateID,
		Path:       path,
		Image:      template.ParsedSpec.Image,     // 使用模板中解析的镜像
		Env:        template.ParsedSpec.Env,       // 使用模板中的环境变量
		Replicas:   1,                             // 默认1个副本
		Resources:  template.ParsedSpec.Resources, // 使用模板中的资源配置
		ContainerConfig: models.ContainerConfig{
			Command:    template.ParsedSpec.Command,    // 使用模板中的命令
			Args:       template.ParsedSpec.Args,       // 使用模板中的参数
			WorkingDir: template.ParsedSpec.WorkingDir, // 使用模板中的工作目录
		},
		Status:     models.StatusCreating,
		TTLSeconds: req.TTLSeconds,                 // 保存TTL值
		ExpireAt:   time.Now().Add(24 * time.Hour), // 临时过期时间
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// 生成K8s资源名称
	deploymentName := variables["DEPLOYMENT_NAME"]
	serviceName := variables["SERVICE_NAME"]
	url.K8sDeploymentName = &deploymentName
	url.K8sServiceName = &serviceName

	// 开始事务处理
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 插入数据库记录
	if err := s.insertURLRecordWithTemplate(ctx, tx, url); err != nil {
		return nil, fmt.Errorf("failed to insert URL record: %w", err)
	}

	// 在开发环境中，不实际创建Kubernetes资源，只保存到数据库
	if s.config.Environment == "development" {
		// 开发环境：设置为draft状态，不部署
		s.updateURLStatus(ctx, url.ID, "draft", "")
		logrus.Info("URL created from template in draft mode (development environment)")
	} else {
		// 生产环境：实际创建Kubernetes资源
		if err := s.createKubernetesResourcesFromYAML(ctx, url, processedYAML); err != nil {
			logrus.WithError(err).Error("Failed to create Kubernetes resources from template")
			// 更新状态为失败
			s.updateURLStatus(ctx, url.ID, models.StatusFailed, err.Error())
			return nil, fmt.Errorf("failed to create Kubernetes resources: %w", err)
		}
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// 异步验证部署状态
	go s.verifyDeployment(url)

	// 构建URL
	fullURL := fmt.Sprintf("https://%s%s", s.config.K8s.DefaultDomain, path)

	logrus.WithFields(logrus.Fields{
		"url_id":      url.ID,
		"project_id":  projectID,
		"template_id": req.TemplateID,
		"path":        path,
	}).Info("Ephemeral URL created from template successfully")

	return &models.CreateEphemeralURLResponse{
		URL: fullURL,
		ID:  url.ID,
	}, nil
}

// insertURLRecordWithTemplate 插入包含模版ID的URL记录
func (s *URLService) insertURLRecordWithTemplate(ctx context.Context, tx *sql.Tx, url *models.EphemeralURL) error {
	query := `
		INSERT INTO ephemeral_urls (
			id, project_id, template_id, path, image, env, replicas, resources, container_config, status, ttl_seconds,
			k8s_deployment_name, k8s_service_name, k8s_secret_name,
			expire_at, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17
		)
	`

	_, err := tx.ExecContext(ctx, query,
		url.ID, url.ProjectID, url.TemplateID, url.Path, url.Image,
		url.Env, url.Replicas, url.Resources, url.ContainerConfig,
		url.Status, url.TTLSeconds, url.K8sDeploymentName, url.K8sServiceName, url.K8sSecretName,
		url.ExpireAt, url.CreatedAt, url.UpdatedAt,
	)

	return err
}

// createKubernetesResourcesFromYAML 从处理后的YAML创建 Kubernetes 资源
func (s *URLService) createKubernetesResourcesFromYAML(ctx context.Context, url *models.EphemeralURL, yamlSpec string) error {
	if s.resourceManager == nil {
		return fmt.Errorf("kubernetes resource manager not available")
	}

	logrus.WithField("url_id", url.ID).Info("Creating Kubernetes resources from YAML template")

	// 解析YAML并创建资源
	if err := s.resourceManager.CreateResourcesFromYAML(ctx, yamlSpec); err != nil {
		return fmt.Errorf("failed to create resources from YAML: %w", err)
	}

	// 更新状态为waiting（等待Pod Ready）
	s.updateURLStatus(ctx, url.ID, models.StatusWaiting, "")

	return nil
}

// StartPodMonitor 启动Pod状态监控服务
func (s *URLService) StartPodMonitor() {
	logrus.Info("Starting Pod status monitor")

	ticker := time.NewTicker(30 * time.Second) // 每30秒检查一次
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.monitorPodStatus()
		}
	}
}

// monitorPodStatus 监控Pod状态
func (s *URLService) monitorPodStatus() {
	ctx := context.Background()

	// 查找所有处于waiting状态的URL
	query := `
		SELECT id, k8s_deployment_name, ttl_seconds, created_at
		FROM ephemeral_urls 
		WHERE status = $1 AND k8s_deployment_name IS NOT NULL
	`

	rows, err := s.db.QueryContext(ctx, query, models.StatusWaiting)
	if err != nil {
		logrus.WithError(err).Error("Failed to query waiting URLs")
		return
	}
	defer rows.Close()

	for rows.Next() {
		var (
			urlID          uuid.UUID
			deploymentName sql.NullString
			ttlSeconds     int
			createdAt      time.Time
		)

		if err := rows.Scan(&urlID, &deploymentName, &ttlSeconds, &createdAt); err != nil {
			logrus.WithError(err).Error("Failed to scan waiting URL")
			continue
		}

		// 检查是否超时（创建后15分钟还没Ready）
		if time.Since(createdAt) > 15*time.Minute {
			logrus.WithField("url_id", urlID).Warn("URL has been waiting too long, marking as failed")
			s.updateURLStatus(ctx, urlID, models.StatusFailed, "Pod failed to become ready within 15 minutes")
			continue
		}

		// 检查Pod是否Ready
		if deploymentName.Valid && s.resourceManager != nil {
			ready, err := s.resourceManager.CheckDeploymentReady(ctx, deploymentName.String)
			if err != nil {
				// 如果是deployment不存在的错误，说明URL可能在draft状态，不应该处于waiting状态
				if strings.Contains(err.Error(), "not found") {
					logrus.WithFields(logrus.Fields{
						"url_id":          urlID,
						"deployment_name": deploymentName.String,
					}).Warn("Deployment not found for waiting URL, this might indicate a state inconsistency")
					// 将URL标记为失败状态
					s.updateURLStatus(ctx, urlID, models.StatusFailed, "Deployment not found")
				} else {
					logrus.WithError(err).WithField("url_id", urlID).Error("Failed to check deployment readiness")
				}
				continue
			}

			if ready {
				// Pod已经 Ready，计算真正的过期时间
				now := time.Now()
				expireAt := now.Add(time.Duration(ttlSeconds) * time.Second)

				// 更新数据库：设置 started_at, expire_at 和 status
				updateQuery := `
					UPDATE ephemeral_urls 
					SET started_at = $1, expire_at = $2, status = $3, updated_at = $4
					WHERE id = $5
				`
				_, err = s.db.ExecContext(ctx, updateQuery, now, expireAt, models.StatusActive, now, urlID)
				if err != nil {
					logrus.WithError(err).WithField("url_id", urlID).Error("Failed to update URL with ready status")
					continue
				}

				logrus.WithFields(logrus.Fields{
					"url_id":     urlID,
					"started_at": now,
					"expire_at":  expireAt,
				}).Info("URL is now active and TTL countdown started")
			}
		}
	}

	if err = rows.Err(); err != nil {
		logrus.WithError(err).Error("Error iterating waiting URLs")
	}
}
