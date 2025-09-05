package handlers

import (
	"net/http"
	"strconv"
	"url-manager-system/backend/internal/api/middleware"
	"url-manager-system/backend/internal/db/models"
	"url-manager-system/backend/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// TemplateHandler 模版处理器
type TemplateHandler struct {
	templateService *services.TemplateService
}

// NewTemplateHandler 创建模版处理器
func NewTemplateHandler(templateService *services.TemplateService) *TemplateHandler {
	return &TemplateHandler{
		templateService: templateService,
	}
}

// CreateTemplate 创建应用模版
func (h *TemplateHandler) CreateTemplate(c *gin.Context) {
	// 获取当前用户ID
	userID, err := middleware.GetCurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User authentication required"})
		return
	}

	var req models.CreateAppTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	template, err := h.templateService.CreateTemplate(c.Request.Context(), userID, &req)
	if err != nil {
		logrus.WithError(err).Error("Failed to create template")

		// 根据错误类型返回不同的状态码
		if err.Error() == "template name '"+req.Name+"' already exists" {
			c.JSON(http.StatusConflict, gin.H{"error": "Template name already exists"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create template"})
		}
		return
	}

	c.JSON(http.StatusCreated, template)
}

// GetTemplate 获取单个模版
func (h *TemplateHandler) GetTemplate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template ID"})
		return
	}

	// 获取当前用户信息
	userID, err := middleware.GetCurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User authentication required"})
		return
	}

	userRole, err := middleware.GetCurrentUserRole(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User role not found"})
		return
	}

	template, err := h.templateService.GetTemplate(c.Request.Context(), id)
	if err != nil {
		if err.Error() == "template not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
			return
		}
		logrus.WithError(err).Error("Failed to get template")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get template"})
		return
	}

	// 检查权限：普通用户只能查看自己的模版
	if userRole != models.RoleAdmin && template.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	c.JSON(http.StatusOK, template)
}

// ListTemplates 列出所有模版
func (h *TemplateHandler) ListTemplates(c *gin.Context) {
	// 解析分页参数
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 20
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// 获取当前用户信息
	userID, err := middleware.GetCurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User authentication required"})
		return
	}

	userRole, err := middleware.GetCurrentUserRole(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User role not found"})
		return
	}

	isAdmin := userRole == models.RoleAdmin

	templates, total, err := h.templateService.ListTemplates(c.Request.Context(), &userID, isAdmin, limit, offset)
	if err != nil {
		logrus.WithError(err).Error("Failed to list templates")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list templates"})
		return
	}

	response := models.ListAppTemplatesResponse{
		Templates: templates,
		Total:     total,
	}

	c.JSON(http.StatusOK, response)
}

// UpdateTemplate 更新模版
func (h *TemplateHandler) UpdateTemplate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template ID"})
		return
	}

	// 获取当前用户信息
	userID, err := middleware.GetCurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User authentication required"})
		return
	}

	userRole, err := middleware.GetCurrentUserRole(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User role not found"})
		return
	}

	isAdmin := userRole == models.RoleAdmin

	var req models.UpdateAppTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	template, err := h.templateService.UpdateTemplate(c.Request.Context(), id, userID, isAdmin, &req)
	if err != nil {
		logrus.WithError(err).Error("Failed to update template")

		// 根据错误类型返回不同的状态码
		switch {
		case err.Error() == "template not found":
			c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
		case err.Error() == "access denied: template belongs to another user":
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		case err.Error() == "template name '"+req.Name+"' already exists":
			c.JSON(http.StatusConflict, gin.H{"error": "Template name already exists"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update template"})
		}
		return
	}

	c.JSON(http.StatusOK, template)
}

// DeleteTemplate 删除模版
func (h *TemplateHandler) DeleteTemplate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template ID"})
		return
	}

	// 获取当前用户信息
	userID, err := middleware.GetCurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User authentication required"})
		return
	}

	userRole, err := middleware.GetCurrentUserRole(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User role not found"})
		return
	}

	isAdmin := userRole == models.RoleAdmin

	err = h.templateService.DeleteTemplate(c.Request.Context(), id, userID, isAdmin)
	if err != nil {
		logrus.WithError(err).Error("Failed to delete template")

		// 根据错误类型返回不同的状态码
		switch {
		case err.Error() == "template not found":
			c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
		case err.Error() == "access denied: template belongs to another user":
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		case err.Error()[:32] == "template is being used by":
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete template"})
		}
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// GetTemplateVariables 获取模版中的占位符变量
func (h *TemplateHandler) GetTemplateVariables(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template ID"})
		return
	}

	variables, err := h.templateService.GetTemplateVariables(c.Request.Context(), id)
	if err != nil {
		if err.Error() == "template not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
			return
		}
		logrus.WithError(err).Error("Failed to get template variables")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get template variables"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"variables": variables,
	})
}

// PreviewTemplate 预览处理后的模版
func (h *TemplateHandler) PreviewTemplate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template ID"})
		return
	}

	// 解析变量参数
	var variables map[string]string
	if err := c.ShouldBindJSON(&variables); err != nil {
		// 如果没有传递变量，使用默认示例变量
		variables = map[string]string{
			"PATH":            "example-path",
			"SERVICE_NAME":    "example-service",
			"DEPLOYMENT_NAME": "example-deployment",
			"PROJECT_NAME":    "example-project",
			"UUID":            "abc12345",
		}
	}

	processedYAML, err := h.templateService.ProcessTemplate(c.Request.Context(), id, variables)
	if err != nil {
		if err.Error() == "template not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
			return
		}
		logrus.WithError(err).Error("Failed to process template")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process template"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"processed_yaml": processedYAML,
		"variables":      variables,
	})
}
