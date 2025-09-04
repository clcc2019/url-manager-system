package middleware

import (
	"net/http"
	"strings"
	"url-manager-system/backend/internal/config"

	"github.com/gin-gonic/gin"
)

// ImageValidator 镜像验证中间件
func ImageValidator(cfg *config.Config) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// 只对创建URL的API进行验证
		if c.Request.Method == "POST" && strings.Contains(c.Request.URL.Path, "/urls") {
			// 这里可以添加更复杂的验证逻辑
			// 例如从请求体中解析image字段并验证
		}
		c.Next()
	})
}

// ResourceValidator 资源限制验证中间件
func ResourceValidator(cfg *config.Config) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// 验证资源配额等
		c.Next()
	})
}

// PodSecurityPolicy Pod安全策略验证
func PodSecurityPolicy() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// 这里可以添加Pod安全策略验证
		// 确保创建的Pod符合安全要求
		c.Next()
	})
}