package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"url-manager-system/backend/internal/db/models"
	"url-manager-system/backend/internal/services"
	"url-manager-system/backend/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// ProjectHandler 项目处理器
type ProjectHandler struct {
	projectService *services.ProjectService
}

// NewProjectHandler 创建项目处理器
func NewProjectHandler(projectService *services.ProjectService) *ProjectHandler {
	return &ProjectHandler{
		projectService: projectService,
	}
}

// validateProjectName 验证项目名称
func (h *ProjectHandler) validateProjectName(name string) error {
	// 基本验证
	name = strings.TrimSpace(name)
	if len(name) == 0 {
		return fmt.Errorf("项目名称不能为空")
	}
	if len(name) > 100 {
		return fmt.Errorf("项目名称不能超过100个字符")
	}

	// 检查清理后的名称是否有效
	sanitized := utils.SanitizeKubernetesName(name)
	if sanitized == "unnamed" {
		return fmt.Errorf("项目名称包含无效字符，请使用字母、数字、连字符或下划线")
	}

	// 如果清理后的名称与原名称差异太大，给出警告信息
	if len(sanitized) < len(name)/2 {
		return fmt.Errorf("项目名称包含过多特殊字符，建议使用更多的字母和数字")
	}

	return nil
}

// CreateProject 创建项目
func (h *ProjectHandler) CreateProject(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required,min=1,max=100"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证项目名称
	if err := h.validateProjectName(req.Name); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	project, err := h.projectService.CreateProject(c.Request.Context(), req.Name, req.Description)
	if err != nil {
		logrus.WithError(err).Error("Failed to create project")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create project"})
		return
	}

	c.JSON(http.StatusCreated, project)
}

// GetProject 获取项目
func (h *ProjectHandler) GetProject(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	project, err := h.projectService.GetProject(c.Request.Context(), id)
	if err != nil {
		if err.Error() == "project not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
			return
		}
		logrus.WithError(err).Error("Failed to get project")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get project"})
		return
	}

	c.JSON(http.StatusOK, project)
}

// ListProjects 列出项目
func (h *ProjectHandler) ListProjects(c *gin.Context) {
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

	projects, total, err := h.projectService.ListProjects(c.Request.Context(), limit, offset)
	if err != nil {
		logrus.WithError(err).Error("Failed to list projects")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list projects"})
		return
	}

	response := models.ListProjectsResponse{
		Projects: projects,
		Total:    total,
	}

	c.JSON(http.StatusOK, response)
}

// UpdateProject 更新项目
func (h *ProjectHandler) UpdateProject(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	var req struct {
		Name        string `json:"name" binding:"required,min=1,max=100"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证项目名称
	if err := h.validateProjectName(req.Name); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	project, err := h.projectService.UpdateProject(c.Request.Context(), id, req.Name, req.Description)
	if err != nil {
		if err.Error() == "project not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
			return
		}
		logrus.WithError(err).Error("Failed to update project")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update project"})
		return
	}

	c.JSON(http.StatusOK, project)
}

// DeleteProject 删除项目
func (h *ProjectHandler) DeleteProject(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	err = h.projectService.DeleteProject(c.Request.Context(), id)
	if err != nil {
		if err.Error() == "project not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
			return
		}
		if err.Error() == "cannot delete project with active URLs" {
			c.JSON(http.StatusConflict, gin.H{"error": "Cannot delete project with active URLs"})
			return
		}
		logrus.WithError(err).Error("Failed to delete project")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete project"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
