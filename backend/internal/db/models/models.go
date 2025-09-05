package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Project 项目模型
type Project struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name" binding:"required,min=1,max=100"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// EphemeralURL 临时URL模型
type EphemeralURL struct {
	ID                uuid.UUID       `json:"id" db:"id"`
	ProjectID         uuid.UUID       `json:"project_id" db:"project_id"`
	Path              string          `json:"path" db:"path"`
	Image             string          `json:"image" db:"image" binding:"required"`
	Env               EnvironmentVars `json:"env" db:"env"`
	Replicas          int             `json:"replicas" db:"replicas" binding:"min=1,max=10"`
	Resources         ResourceLimits  `json:"resources" db:"resources"`
	Status            string          `json:"status" db:"status"`
	K8sDeploymentName *string         `json:"k8s_deployment_name" db:"k8s_deployment_name"`
	K8sServiceName    *string         `json:"k8s_service_name" db:"k8s_service_name"`
	K8sSecretName     *string         `json:"k8s_secret_name" db:"k8s_secret_name"`
	ErrorMessage      *string         `json:"error_message" db:"error_message"`
	ExpireAt          time.Time       `json:"expire_at" db:"expire_at"`
	CreatedAt         time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at" db:"updated_at"`

	// 关联项目信息(用于查询时连表获取)
	Project *Project `json:"project,omitempty"`
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
	StatusActive   = "active"
	StatusDeleting = "deleting"
	StatusDeleted  = "deleted"
	StatusFailed   = "failed"
)

// CreateEphemeralURLRequest 创建URL请求
type CreateEphemeralURLRequest struct {
	Image      string          `json:"image" binding:"required"`
	Env        EnvironmentVars `json:"env"`
	TTLSeconds int             `json:"ttl_seconds" binding:"required,min=60,max=604800"` // 1分钟到7天
	Replicas   int             `json:"replicas" binding:"min=1,max=10"`
	Resources  ResourceLimits  `json:"resources"`
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
