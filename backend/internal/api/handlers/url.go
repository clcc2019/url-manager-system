package handlers

import (
	"net/http"
	"strconv"
	"url-manager-system/backend/internal/db/models"
	"url-manager-system/backend/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// URLHandler URL处理器
type URLHandler struct {
	urlService     *services.URLService
	cleanupService *services.CleanupService
}

// NewURLHandler 创建URL处理器
func NewURLHandler(urlService *services.URLService, cleanupService *services.CleanupService) *URLHandler {
	return &URLHandler{
		urlService:     urlService,
		cleanupService: cleanupService,
	}
}

// CreateEphemeralURL 创建临时URL
func (h *URLHandler) CreateEphemeralURL(c *gin.Context) {
	projectIDStr := c.Param("id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	var req models.CreateEphemeralURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.urlService.CreateEphemeralURL(c.Request.Context(), projectID, &req)
	if err != nil {
		logrus.WithError(err).Error("Failed to create ephemeral URL")

		// 根据错误类型返回不同的状态码
		switch {
		case err.Error() == "project not found":
			c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		case err.Error() == "validation failed: image not in allowed list":
			c.JSON(http.StatusBadRequest, gin.H{"error": "Image not allowed"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create ephemeral URL"})
		}
		return
	}

	c.JSON(http.StatusCreated, response)
}

// GetEphemeralURL 获取临时URL
func (h *URLHandler) GetEphemeralURL(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid URL ID"})
		return
	}

	url, err := h.urlService.GetEphemeralURL(c.Request.Context(), id)
	if err != nil {
		if err.Error() == "URL not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "URL not found"})
			return
		}
		logrus.WithError(err).Error("Failed to get ephemeral URL")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get URL"})
		return
	}

	c.JSON(http.StatusOK, url)
}

// ListEphemeralURLs 列出项目的临时URL
func (h *URLHandler) ListEphemeralURLs(c *gin.Context) {
	projectIDStr := c.Param("id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

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

	urls, total, err := h.urlService.ListEphemeralURLs(c.Request.Context(), projectID, limit, offset)
	if err != nil {
		logrus.WithError(err).Error("Failed to list ephemeral URLs")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list URLs"})
		return
	}

	response := models.ListURLsResponse{
		URLs:  urls,
		Total: total,
	}

	c.JSON(http.StatusOK, response)
}

// DeleteEphemeralURL 删除临时URL
func (h *URLHandler) DeleteEphemeralURL(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid URL ID"})
		return
	}

	err = h.cleanupService.ForceCleanupURL(c.Request.Context(), id)
	if err != nil {
		if err.Error() == "URL not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "URL not found"})
			return
		}
		logrus.WithError(err).Error("Failed to delete ephemeral URL")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete URL"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// GetURLByPath 根据路径获取URL（用于内部查询）
func (h *URLHandler) GetURLByPath(c *gin.Context) {
	path := c.Param("path")
	if path == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Path is required"})
		return
	}

	// 这里需要实现根据路径查询URL的逻辑
	// 为了简化，暂时返回未实现
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

// DeployURL 部署URL到Kubernetes集群
func (h *URLHandler) DeployURL(c *gin.Context) {
	urlIDStr := c.Param("id")
	urlID, err := uuid.Parse(urlIDStr)
	if err != nil {
		logrus.WithError(err).Error("Invalid URL ID format")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid URL ID format"})
		return
	}

	err = h.urlService.DeployURL(c.Request.Context(), urlID)
	if err != nil {
		logrus.WithError(err).WithField("url_id", urlID).Error("Failed to deploy URL")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logrus.WithField("url_id", urlID).Info("URL deployment initiated")
	c.JSON(http.StatusOK, gin.H{"message": "URL deployment initiated successfully"})
}

// ValidateAndCleanupData 校验并清理数据
func (h *URLHandler) ValidateAndCleanupData(c *gin.Context) {
	err := h.cleanupService.ValidateAndCleanupData(c.Request.Context())
	if err != nil {
		logrus.WithError(err).Error("Failed to validate and cleanup data")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate and cleanup data"})
		return
	}

	logrus.Info("Data validation and cleanup completed successfully")
	c.JSON(http.StatusOK, gin.H{"message": "Data validation and cleanup completed successfully"})
}

// CreateEphemeralURLFromTemplate 基于模版创建临时URL
func (h *URLHandler) CreateEphemeralURLFromTemplate(c *gin.Context) {
	projectIDStr := c.Param("id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	var req models.CreateEphemeralURLFromTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.urlService.CreateEphemeralURLFromTemplate(c.Request.Context(), projectID, &req)
	if err != nil {
		logrus.WithError(err).Error("Failed to create ephemeral URL from template")

		// 根据错误类型返回不同的状态码
		switch {
		case err.Error() == "project not found":
			c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		case err.Error() == "template not found":
			c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
		case err.Error()[:4] == "path":
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create ephemeral URL"})
		}
		return
	}

	c.JSON(http.StatusCreated, response)
}
