package services

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
	"url-manager-system/backend/internal/db/models"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

// TemplateService 模版服务
type TemplateService struct {
	db *sqlx.DB
}

// NewTemplateService 创建模版服务
func NewTemplateService(db *sqlx.DB) *TemplateService {
	return &TemplateService{
		db: db,
	}
}

// CreateTemplate 创建应用模版
func (s *TemplateService) CreateTemplate(ctx context.Context, userID uuid.UUID, req *models.CreateAppTemplateRequest) (*models.AppTemplate, error) {
	// 验证模版名称唯一性（同一用户下）
	var count int
	err := s.db.GetContext(ctx, &count, "SELECT COUNT(*) FROM app_templates WHERE name = $1 AND user_id = $2", req.Name, userID)
	if err != nil {
		logrus.WithError(err).Error("Failed to check template name uniqueness")
		return nil, fmt.Errorf("failed to check template name uniqueness: %w", err)
	}
	if count > 0 {
		return nil, fmt.Errorf("template name '%s' already exists", req.Name)
	}

	// 验证YAML规范
	if err := s.validateYamlSpec(req.YamlSpec); err != nil {
		return nil, fmt.Errorf("invalid YAML specification: %w", err)
	}

	// 创建模版记录
	template := &models.AppTemplate{
		ID:          uuid.New(),
		UserID:      userID,
		Name:        req.Name,
		Description: req.Description,
		YamlSpec:    req.YamlSpec,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	query := `
		INSERT INTO app_templates (id, user_id, name, description, yaml_spec, created_at, updated_at)
		VALUES (:id, :user_id, :name, :description, :yaml_spec, :created_at, :updated_at)
	`

	_, err = s.db.NamedExecContext(ctx, query, template)
	if err != nil {
		logrus.WithError(err).Error("Failed to create template")
		return nil, fmt.Errorf("failed to create template: %w", err)
	}

	logrus.WithField("template_id", template.ID).Info("Template created successfully")
	return template, nil
}

// GetTemplate 获取单个模版
func (s *TemplateService) GetTemplate(ctx context.Context, id uuid.UUID) (*models.AppTemplate, error) {
	var template models.AppTemplate
	query := "SELECT * FROM app_templates WHERE id = $1"

	err := s.db.GetContext(ctx, &template, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("template not found")
		}
		logrus.WithError(err).Error("Failed to get template")
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	return &template, nil
}

// ListTemplates 列出模版（支持用户过滤）
func (s *TemplateService) ListTemplates(ctx context.Context, userID *uuid.UUID, isAdmin bool, limit, offset int) ([]models.AppTemplate, int, error) {
	var countQuery, listQuery string
	var args []interface{}

	if isAdmin {
		// 管理员可以查看所有模版
		countQuery = "SELECT COUNT(*) FROM app_templates"
		listQuery = `
			SELECT * FROM app_templates 
			ORDER BY created_at DESC 
			LIMIT $1 OFFSET $2
		`
		args = []interface{}{limit, offset}
	} else {
		// 普通用户只能查看自己的模版
		countQuery = "SELECT COUNT(*) FROM app_templates WHERE user_id = $1"
		listQuery = `
			SELECT * FROM app_templates 
			WHERE user_id = $1
			ORDER BY created_at DESC 
			LIMIT $2 OFFSET $3
		`
		args = []interface{}{*userID, limit, offset}
	}

	// 获取总数
	var total int
	if isAdmin {
		err := s.db.GetContext(ctx, &total, countQuery)
		if err != nil {
			logrus.WithError(err).Error("Failed to count templates")
			return nil, 0, fmt.Errorf("failed to count templates: %w", err)
		}
	} else {
		err := s.db.GetContext(ctx, &total, countQuery, *userID)
		if err != nil {
			logrus.WithError(err).Error("Failed to count templates")
			return nil, 0, fmt.Errorf("failed to count templates: %w", err)
		}
	}

	// 获取模版列表
	var templates []models.AppTemplate
	err := s.db.SelectContext(ctx, &templates, listQuery, args...)
	if err != nil {
		logrus.WithError(err).Error("Failed to list templates")
		return nil, 0, fmt.Errorf("failed to list templates: %w", err)
	}

	return templates, total, nil
}

// UpdateTemplate 更新模版
func (s *TemplateService) UpdateTemplate(ctx context.Context, id uuid.UUID, userID uuid.UUID, isAdmin bool, req *models.UpdateAppTemplateRequest) (*models.AppTemplate, error) {
	// 检查模版是否存在
	existingTemplate, err := s.GetTemplate(ctx, id)
	if err != nil {
		return nil, err
	}

	// 检查权限：普通用户只能更新自己的模版
	if !isAdmin && existingTemplate.UserID != userID {
		return nil, fmt.Errorf("access denied: template belongs to another user")
	}

	// 如果名称发生变化，检查新名称在同一用户下的唯一性
	if req.Name != existingTemplate.Name {
		var count int
		var checkQuery string
		var checkArgs []interface{}

		if isAdmin {
			// 管理员可以使用全局唯一的模版名
			checkQuery = "SELECT COUNT(*) FROM app_templates WHERE name = $1 AND id != $2"
			checkArgs = []interface{}{req.Name, id}
		} else {
			// 普通用户只需要在自己的模版中唯一
			checkQuery = "SELECT COUNT(*) FROM app_templates WHERE name = $1 AND user_id = $2 AND id != $3"
			checkArgs = []interface{}{req.Name, userID, id}
		}

		err := s.db.GetContext(ctx, &count, checkQuery, checkArgs...)
		if err != nil {
			logrus.WithError(err).Error("Failed to check template name uniqueness")
			return nil, fmt.Errorf("failed to check template name uniqueness: %w", err)
		}
		if count > 0 {
			return nil, fmt.Errorf("template name '%s' already exists", req.Name)
		}
	}

	// 验证YAML规范
	if err := s.validateYamlSpec(req.YamlSpec); err != nil {
		return nil, fmt.Errorf("invalid YAML specification: %w", err)
	}

	// 更新模版
	query := `
		UPDATE app_templates 
		SET name = $1, description = $2, yaml_spec = $3, updated_at = $4
		WHERE id = $5
	`

	_, err = s.db.ExecContext(ctx, query, req.Name, req.Description, req.YamlSpec, time.Now(), id)
	if err != nil {
		logrus.WithError(err).Error("Failed to update template")
		return nil, fmt.Errorf("failed to update template: %w", err)
	}

	// 返回更新后的模版
	return s.GetTemplate(ctx, id)
}

// DeleteTemplate 删除模版
func (s *TemplateService) DeleteTemplate(ctx context.Context, id uuid.UUID, userID uuid.UUID, isAdmin bool) error {
	// 检查模版是否存在并验证权限
	existingTemplate, err := s.GetTemplate(ctx, id)
	if err != nil {
		return err
	}

	// 检查权限：普通用户只能删除自己的模版
	if !isAdmin && existingTemplate.UserID != userID {
		return fmt.Errorf("access denied: template belongs to another user")
	}

	// 检查模版是否被URL使用
	var urlCount int
	err = s.db.GetContext(ctx, &urlCount, "SELECT COUNT(*) FROM ephemeral_urls WHERE template_id = $1", id)
	if err != nil {
		logrus.WithError(err).Error("Failed to check template usage")
		return fmt.Errorf("failed to check template usage: %w", err)
	}

	if urlCount > 0 {
		return fmt.Errorf("template is being used by %d URLs, cannot delete", urlCount)
	}

	// 删除模版
	result, err := s.db.ExecContext(ctx, "DELETE FROM app_templates WHERE id = $1", id)
	if err != nil {
		logrus.WithError(err).Error("Failed to delete template")
		return fmt.Errorf("failed to delete template: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("template not found")
	}

	logrus.WithField("template_id", id).Info("Template deleted successfully")
	return nil
}

// ProcessTemplate 处理模版，替换占位符
func (s *TemplateService) ProcessTemplate(ctx context.Context, templateID uuid.UUID, variables map[string]string) (string, error) {
	template, err := s.GetTemplate(ctx, templateID)
	if err != nil {
		return "", err
	}

	yamlSpec := template.YamlSpec

	// 替换占位符
	for key, value := range variables {
		placeholder := fmt.Sprintf("${%s}", key)
		yamlSpec = strings.ReplaceAll(yamlSpec, placeholder, value)
	}

	return yamlSpec, nil
}

// validateYamlSpec 验证YAML规范基本格式
func (s *TemplateService) validateYamlSpec(yamlSpec string) error {
	if strings.TrimSpace(yamlSpec) == "" {
		return fmt.Errorf("YAML specification cannot be empty")
	}

	// 基本的YAML格式检查
	if !strings.Contains(yamlSpec, "apiVersion") {
		return fmt.Errorf("YAML specification must contain 'apiVersion' field")
	}

	if !strings.Contains(yamlSpec, "kind") {
		return fmt.Errorf("YAML specification must contain 'kind' field")
	}

	// TODO: 添加更严格的YAML解析验证
	return nil
}

// GetTemplateVariables 获取模版中的占位符变量
func (s *TemplateService) GetTemplateVariables(ctx context.Context, templateID uuid.UUID) ([]string, error) {
	template, err := s.GetTemplate(ctx, templateID)
	if err != nil {
		return nil, err
	}

	var variables []string
	yamlSpec := template.YamlSpec

	// 简单的占位符解析（查找 ${VARIABLE_NAME} 模式）
	variableMap := make(map[string]bool)

	start := 0
	for {
		startIdx := strings.Index(yamlSpec[start:], "${")
		if startIdx == -1 {
			break
		}
		startIdx += start

		endIdx := strings.Index(yamlSpec[startIdx:], "}")
		if endIdx == -1 {
			break
		}
		endIdx += startIdx

		variable := yamlSpec[startIdx+2 : endIdx]
		if variable != "" && !variableMap[variable] {
			variables = append(variables, variable)
			variableMap[variable] = true
		}

		start = endIdx + 1
	}

	return variables, nil
}
