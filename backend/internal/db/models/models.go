package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// User 用户模型
type User struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	Username     string     `json:"username" db:"username" binding:"required,min=3,max=50"`
	PasswordHash string     `json:"-" db:"password_hash"` // 不返回到前端
	Role         string     `json:"role" db:"role"`
	Email        *string    `json:"email" db:"email"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
	LastLoginAt  *time.Time `json:"last_login_at" db:"last_login_at"`
}

// UserRole 用户角色常量
const (
	RoleAdmin = "admin"
	RoleUser  = "user"
)

// AppTemplate 应用模版模型
type AppTemplate struct {
	ID          uuid.UUID `json:"id" db:"id"`
	UserID      uuid.UUID `json:"user_id" db:"user_id"`
	Name        string    `json:"name" db:"name" binding:"required,min=1,max=100"`
	Description string    `json:"description" db:"description"`
	YamlSpec    string    `json:"yaml_spec" db:"yaml_spec" binding:"required"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// Project 项目模型
type Project struct {
	ID          uuid.UUID `json:"id" db:"id"`
	UserID      uuid.UUID `json:"user_id" db:"user_id"`
	Name        string    `json:"name" db:"name" binding:"required,min=1,max=100"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// EphemeralURL 临时URL模型
type EphemeralURL struct {
	ID                uuid.UUID       `json:"id" db:"id"`
	UserID            uuid.UUID       `json:"user_id" db:"user_id"`
	ProjectID         uuid.UUID       `json:"project_id" db:"project_id"`
	TemplateID        *uuid.UUID      `json:"template_id" db:"template_id"`
	Path              string          `json:"path" db:"path"`
	Image             string          `json:"image" db:"image" binding:"required"`
	Env               EnvironmentVars `json:"env" db:"env"`
	Replicas          int             `json:"replicas" db:"replicas" binding:"min=1,max=10"`
	Resources         ResourceLimits  `json:"resources" db:"resources"`
	ContainerConfig   ContainerConfig `json:"container_config" db:"container_config"`
	Status            string          `json:"status" db:"status"`
	TTLSeconds        int             `json:"ttl_seconds" db:"ttl_seconds"`
	K8sDeploymentName *string         `json:"k8s_deployment_name" db:"k8s_deployment_name"`
	K8sServiceName    *string         `json:"k8s_service_name" db:"k8s_service_name"`
	K8sSecretName     *string         `json:"k8s_secret_name" db:"k8s_secret_name"`
	ErrorMessage      *string         `json:"error_message" db:"error_message"`
	StartedAt         *time.Time      `json:"started_at" db:"started_at"`
	ExpireAt          time.Time       `json:"expire_at" db:"expire_at"`
	CreatedAt         time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at" db:"updated_at"`

	// 关联项目信息(用于查询时连表获取)
	Project *Project `json:"project,omitempty"`
	// 关联模版信息(用于查询时连表获取)
	Template *AppTemplate `json:"template,omitempty"`
}

// EnvironmentVar 环境变量
type EnvironmentVar struct {
	Name  string `json:"name" binding:"required"`
	Value string `json:"value" binding:"required"`
}

// EnvironmentVars 环境变量列表
type EnvironmentVars []EnvironmentVar

// Value 实现driver.Valuer接口，用于数据库存储
func (e EnvironmentVars) Value() (driver.Value, error) {
	if e == nil {
		return nil, nil
	}
	return json.Marshal(e)
}

// Scan 实现sql.Scanner接口，用于数据库读取
func (e *EnvironmentVars) Scan(value interface{}) error {
	if value == nil {
		*e = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	return json.Unmarshal(bytes, e)
}

// ResourceRequests 资源请求
type ResourceRequests struct {
	CPU    string `json:"cpu"`
	Memory string `json:"memory"`
}

// ResourceLimits 资源限制
type ResourceLimits struct {
	Requests ResourceRequests `json:"requests"`
	Limits   ResourceRequests `json:"limits"`
}

// DeviceMapping 设备映射
type DeviceMapping struct {
	HostPath      string `json:"host_path" binding:"required"`
	ContainerPath string `json:"container_path" binding:"required"`
	Permissions   string `json:"permissions"` // r, w, rw, m
}

// DeviceMappings 设备映射列表
type DeviceMappings []DeviceMapping

// Value 实现driver.Valuer接口
func (d DeviceMappings) Value() (driver.Value, error) {
	if d == nil {
		return nil, nil
	}
	return json.Marshal(d)
}

// Scan 实现sql.Scanner接口
func (d *DeviceMappings) Scan(value interface{}) error {
	if value == nil {
		*d = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	return json.Unmarshal(bytes, d)
}

// ContainerConfig 容器配置
type ContainerConfig struct {
	ContainerName string         `json:"container_name,omitempty"` // 容器名称
	Devices       DeviceMappings `json:"devices,omitempty"`        // 设备映射
	Command       []string       `json:"command,omitempty"`        // 启动命令
	Args          []string       `json:"args,omitempty"`           // 启动参数
	WorkingDir    string         `json:"working_dir,omitempty"`    // 工作目录
	TTY           bool           `json:"tty"`                      // 是否分配TTY
	Stdin         bool           `json:"stdin"`                    // 是否打开stdin
}

// Value 实现driver.Valuer接口
func (c ContainerConfig) Value() (driver.Value, error) {
	return json.Marshal(c)
}

// Scan 实现sql.Scanner接口
func (c *ContainerConfig) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	return json.Unmarshal(bytes, c)
}

// Value 实现driver.Valuer接口
func (r ResourceLimits) Value() (driver.Value, error) {
	return json.Marshal(r)
}

// Scan 实现sql.Scanner接口
func (r *ResourceLimits) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	return json.Unmarshal(bytes, r)
}

// URLStatus 状态常量
const (
	StatusCreating = "creating"
	StatusWaiting  = "waiting"
	StatusActive   = "active"
	StatusDeleting = "deleting"
	StatusDeleted  = "deleted"
	StatusFailed   = "failed"
)

// CreateEphemeralURLRequest 创建URL请求
type CreateEphemeralURLRequest struct {
	Image           string          `json:"image" binding:"required"`
	Env             EnvironmentVars `json:"env"`
	TTLSeconds      int             `json:"ttl_seconds" binding:"required,min=60,max=604800"` // 1分钟到7天
	Replicas        int             `json:"replicas" binding:"min=1,max=10"`
	Resources       ResourceLimits  `json:"resources"`
	ContainerConfig ContainerConfig `json:"container_config"`
}

// CreateEphemeralURLFromTemplateRequest 基于模版创建URL请求
type CreateEphemeralURLFromTemplateRequest struct {
	TemplateID uuid.UUID `json:"template_id" binding:"required"`
	TTLSeconds int       `json:"ttl_seconds" binding:"required,min=60,max=604800"` // 1分钟到7天
	Path       string    `json:"path,omitempty"`                                   // 可选，为空时系统生成
}

// CreateAppTemplateRequest 创建应用模版请求
type CreateAppTemplateRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=100"`
	Description string `json:"description"`
	YamlSpec    string `json:"yaml_spec" binding:"required"`
}

// UpdateAppTemplateRequest 更新应用模版请求
type UpdateAppTemplateRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=100"`
	Description string `json:"description"`
	YamlSpec    string `json:"yaml_spec" binding:"required"`
}

// CreateEphemeralURLResponse 创建URL响应
type CreateEphemeralURLResponse struct {
	URL string    `json:"url"`
	ID  uuid.UUID `json:"id"`
}

// ListProjectsResponse 项目列表响应
type ListProjectsResponse struct {
	Projects []Project `json:"projects"`
	Total    int       `json:"total"`
}

// ListURLsResponse URL列表响应
type ListURLsResponse struct {
	URLs  []EphemeralURL `json:"urls"`
	Total int            `json:"total"`
}

// ListAppTemplatesResponse 模版列表响应
type ListAppTemplatesResponse struct {
	Templates []AppTemplate `json:"templates"`
	Total     int           `json:"total"`
}

// 用户认证相关的请求和响应类型

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6,max=50"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token     string    `json:"token"`
	User      User      `json:"user"`
	ExpiresAt time.Time `json:"expires_at"`
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string  `json:"username" binding:"required,min=3,max=50"`
	Password string  `json:"password" binding:"required,min=6,max=50"`
	Email    *string `json:"email"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6,max=50"`
}

// UserProfileResponse 用户信息响应
type UserProfileResponse struct {
	User User `json:"user"`
}

// JWTClaims JWT声明
type JWTClaims struct {
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	Role     string    `json:"role"`
	Exp      int64     `json:"exp"`
	Iat      int64     `json:"iat"`
}
