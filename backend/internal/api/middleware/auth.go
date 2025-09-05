package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"url-manager-system/backend/internal/db/models"
	"url-manager-system/backend/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AuthMiddleware JWT认证中间件
func AuthMiddleware(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// 检查Bearer token格式
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		tokenString := tokenParts[1]

		// 验证JWT token
		claims, err := authService.ValidateJWT(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// 将用户信息存储到context中
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("user_role", claims.Role)

		c.Next()
	}
}

// AdminMiddleware 管理员权限中间件
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User role not found"})
			c.Abort()
			return
		}

		if role != models.RoleAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin role required"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetCurrentUserID 从context中获取当前用户ID
func GetCurrentUserID(c *gin.Context) (uuid.UUID, error) {
	userID, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil, fmt.Errorf("user ID not found in context")
	}

	uid, ok := userID.(uuid.UUID)
	if !ok {
		return uuid.Nil, fmt.Errorf("invalid user ID format")
	}

	return uid, nil
}

// GetCurrentUserRole 从context中获取当前用户角色
func GetCurrentUserRole(c *gin.Context) (string, error) {
	role, exists := c.Get("user_role")
	if !exists {
		return "", fmt.Errorf("user role not found in context")
	}

	userRole, ok := role.(string)
	if !ok {
		return "", fmt.Errorf("invalid user role format")
	}

	return userRole, nil
}

// IsAdmin 检查当前用户是否为管理员
func IsAdmin(c *gin.Context) bool {
	role, err := GetCurrentUserRole(c)
	if err != nil {
		return false
	}
	return role == models.RoleAdmin
}