package handlers

import (
	"net/http"
	"strconv"
	"url-manager-system/backend/internal/api/middleware"
	"url-manager-system/backend/internal/db/models"
	"url-manager-system/backend/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Login 用户登录
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.authService.Login(&req)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"username": req.Username,
			"error":    err.Error(),
		}).Error("Login failed")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	logrus.WithFields(logrus.Fields{
		"user_id":  resp.User.ID,
		"username": resp.User.Username,
		"role":     resp.User.Role,
	}).Info("User logged in successfully")

	c.JSON(http.StatusOK, resp)
}

// Register 用户注册（仅管理员可用）
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.authService.Register(&req)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"username": req.Username,
			"error":    err.Error(),
		}).Error("Registration failed")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	logrus.WithFields(logrus.Fields{
		"user_id":  user.ID,
		"username": user.Username,
		"role":     user.Role,
	}).Info("User registered successfully")

	// 不返回密码hash
	user.PasswordHash = ""
	c.JSON(http.StatusCreated, gin.H{"user": user})
}

// GetProfile 获取当前用户信息
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID, err := middleware.GetCurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	user, err := h.authService.GetUserByID(userID)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"user_id": userID,
			"error":   err.Error(),
		}).Error("Failed to get user profile")
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// 不返回密码hash
	user.PasswordHash = ""
	c.JSON(http.StatusOK, &models.UserProfileResponse{User: *user})
}

// ChangePassword 修改密码
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID, err := middleware.GetCurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	var req models.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.authService.ChangePassword(userID, &req)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"user_id": userID,
			"error":   err.Error(),
		}).Error("Failed to change password")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	logrus.WithFields(logrus.Fields{
		"user_id": userID,
	}).Info("Password changed successfully")

	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}

// ListUsers 获取用户列表（仅管理员可用）
func (h *AuthHandler) ListUsers(c *gin.Context) {
	// 分页参数
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100 // 限制最大每页数量
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	users, total, err := h.authService.ListUsers(limit, offset)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("Failed to list users")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list users"})
		return
	}

	// 清除密码hash
	for i := range users {
		users[i].PasswordHash = ""
	}

	c.JSON(http.StatusOK, gin.H{
		"users": users,
		"total": total,
	})
}

// Logout 用户登出（前端处理，后端只需返回成功）
func (h *AuthHandler) Logout(c *gin.Context) {
	userID, _ := middleware.GetCurrentUserID(c)
	
	logrus.WithFields(logrus.Fields{
		"user_id": userID,
	}).Info("User logged out")

	c.JSON(http.StatusOK, gin.H{"message": "Logout successful"})
}