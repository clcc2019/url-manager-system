package services

import (
	"crypto/rand"
	"fmt"
	"time"
	"url-manager-system/backend/internal/db/models"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	db     *sqlx.DB
	jwtKey []byte
}

func NewAuthService(db *sqlx.DB, jwtSecret string) *AuthService {
	// 如果没有提供JWT密钥，生成一个随机密钥
	var key []byte
	if jwtSecret != "" {
		key = []byte(jwtSecret)
	} else {
		key = generateRandomKey()
	}
	
	return &AuthService{
		db:     db,
		jwtKey: key,
	}
}

// generateRandomKey 生成随机JWT密钥
func generateRandomKey() []byte {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return bytes
}

// HashPassword 对密码进行bcrypt加密
func (s *AuthService) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPassword 验证密码
func (s *AuthService) CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateJWT 生成JWT token
func (s *AuthService) GenerateJWT(user *models.User) (string, time.Time, error) {
	expirationTime := time.Now().Add(24 * time.Hour) // 24小时有效期
	
	claims := &models.JWTClaims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		Exp:      expirationTime.Unix(),
		Iat:      time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  claims.UserID.String(),
		"username": claims.Username,
		"role":     claims.Role,
		"exp":      claims.Exp,
		"iat":      claims.Iat,
	})

	tokenString, err := token.SignedString(s.jwtKey)
	return tokenString, expirationTime, err
}

// ValidateJWT 验证JWT token
func (s *AuthService) ValidateJWT(tokenString string) (*models.JWTClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userIDStr, ok := claims["user_id"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid user_id in token")
		}
		
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return nil, fmt.Errorf("invalid user_id format: %v", err)
		}

		username, ok := claims["username"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid username in token")
		}

		role, ok := claims["role"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid role in token")
		}

		exp, ok := claims["exp"].(float64)
		if !ok {
			return nil, fmt.Errorf("invalid exp in token")
		}

		iat, ok := claims["iat"].(float64)
		if !ok {
			return nil, fmt.Errorf("invalid iat in token")
		}

		return &models.JWTClaims{
			UserID:   userID,
			Username: username,
			Role:     role,
			Exp:      int64(exp),
			Iat:      int64(iat),
		}, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// Login 用户登录
func (s *AuthService) Login(req *models.LoginRequest) (*models.LoginResponse, error) {
	var user models.User
	query := `
		SELECT id, username, password_hash, role, email, created_at, updated_at, last_login_at
		FROM users 
		WHERE username = $1
	`
	
	err := s.db.Get(&user, query, req.Username)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	if !s.CheckPassword(req.Password, user.PasswordHash) {
		return nil, fmt.Errorf("invalid password")
	}

	// 更新最后登录时间
	now := time.Now()
	user.LastLoginAt = &now
	_, err = s.db.Exec("UPDATE users SET last_login_at = $1 WHERE id = $2", now, user.ID)
	if err != nil {
		// 登录时间更新失败不影响登录流程，只记录日志
		// TODO: 添加日志记录
	}

	token, expiresAt, err := s.GenerateJWT(&user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %v", err)
	}

	return &models.LoginResponse{
		Token:     token,
		User:      user,
		ExpiresAt: expiresAt,
	}, nil
}

// Register 用户注册（管理员功能）
func (s *AuthService) Register(req *models.RegisterRequest) (*models.User, error) {
	// 检查用户名是否已存在
	var count int
	err := s.db.Get(&count, "SELECT COUNT(*) FROM users WHERE username = $1", req.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to check username: %v", err)
	}
	if count > 0 {
		return nil, fmt.Errorf("username already exists")
	}

	// 密码加密
	hashedPassword, err := s.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %v", err)
	}

	// 创建用户
	user := &models.User{
		ID:           uuid.New(),
		Username:     req.Username,
		PasswordHash: hashedPassword,
		Role:         models.RoleUser, // 默认为普通用户
		Email:        req.Email,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	query := `
		INSERT INTO users (id, username, password_hash, role, email, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	
	_, err = s.db.Exec(query, user.ID, user.Username, user.PasswordHash, user.Role, user.Email, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %v", err)
	}

	return user, nil
}

// GetUserByID 根据ID获取用户
func (s *AuthService) GetUserByID(userID uuid.UUID) (*models.User, error) {
	var user models.User
	query := `
		SELECT id, username, password_hash, role, email, created_at, updated_at, last_login_at
		FROM users 
		WHERE id = $1
	`
	
	err := s.db.Get(&user, query, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	return &user, nil
}

// ChangePassword 修改密码
func (s *AuthService) ChangePassword(userID uuid.UUID, req *models.ChangePasswordRequest) error {
	// 获取用户当前密码
	var currentPasswordHash string
	err := s.db.Get(&currentPasswordHash, "SELECT password_hash FROM users WHERE id = $1", userID)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	// 验证旧密码
	if !s.CheckPassword(req.OldPassword, currentPasswordHash) {
		return fmt.Errorf("current password is incorrect")
	}

	// 加密新密码
	newPasswordHash, err := s.HashPassword(req.NewPassword)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %v", err)
	}

	// 更新密码
	_, err = s.db.Exec("UPDATE users SET password_hash = $1, updated_at = $2 WHERE id = $3", 
		newPasswordHash, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to update password: %v", err)
	}

	return nil
}

// GetJWTKey 获取JWT密钥（用于中间件）
func (s *AuthService) GetJWTKey() []byte {
	return s.jwtKey
}

// ListUsers 获取用户列表（管理员功能）
func (s *AuthService) ListUsers(limit, offset int) ([]models.User, int, error) {
	var users []models.User
	var total int

	// 获取总数
	err := s.db.Get(&total, "SELECT COUNT(*) FROM users")
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %v", err)
	}

	// 获取用户列表
	query := `
		SELECT id, username, role, email, created_at, updated_at, last_login_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	
	err = s.db.Select(&users, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %v", err)
	}

	return users, total, nil
}