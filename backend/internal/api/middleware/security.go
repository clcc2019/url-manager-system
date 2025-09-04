package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// SecurityHeaders 添加安全响应头
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 防止点击劫持
		c.Header("X-Frame-Options", "DENY")
		
		// 防止MIME嗅探
		c.Header("X-Content-Type-Options", "nosniff")
		
		// XSS保护
		c.Header("X-XSS-Protection", "1; mode=block")
		
		// 强制HTTPS（仅在生产环境）
		if gin.Mode() == gin.ReleaseMode {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		
		// 内容安全策略
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' https:; connect-src 'self'")
		
		// 引用策略
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		
		// 权限策略
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		
		c.Next()
	}
}

// RateLimiter 简单的速率限制中间件
func RateLimiter() gin.HandlerFunc {
	// 这里可以集成更复杂的速率限制器，如redis-based limiter
	return gin.HandlerFunc(func(c *gin.Context) {
		// 简单实现：检查User-Agent和IP
		userAgent := c.GetHeader("User-Agent")
		if userAgent == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "User-Agent required"})
			c.Abort()
			return
		}
		
		// 阻止明显的机器人请求
		suspiciousAgents := []string{"bot", "crawler", "spider", "scraper"}
		lowerUserAgent := strings.ToLower(userAgent)
		for _, agent := range suspiciousAgents {
			if strings.Contains(lowerUserAgent, agent) {
				c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
				c.Abort()
				return
			}
		}
		
		c.Next()
	})
}

// RequestSize 限制请求体大小
func RequestSize(maxSize int64) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSize)
		c.Next()
	})
}