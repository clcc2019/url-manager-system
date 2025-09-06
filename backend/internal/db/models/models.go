package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"k8s.io/apimachinery/pkg/util/intstr"
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
	ID          uuid.UUID    `json:"id" db:"id"`
	UserID      uuid.UUID    `json:"user_id" db:"user_id"`
	Name        string       `json:"name" db:"name" binding:"required,min=1,max=100"`
	Description string       `json:"description" db:"description"`
	YamlSpec    string       `json:"yaml_spec" db:"yaml_spec" binding:"required"`
	ParsedSpec  TemplateSpec `json:"parsed_spec" db:"parsed_spec"` // 解析后的结构化数据
	CreatedAt   time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at" db:"updated_at"`
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
	Logs              []LogEntry      `json:"logs" db:"logs"`
	IngressHost       *string         `json:"ingress_host" db:"ingress_host"`
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

// LogEntry 日志条目
type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"` // info, warn, error
	Message   string    `json:"message"`
	Details   string    `json:"details,omitempty"`
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

// TemplateSpec 模板规格（解析后的YAML结构）
type TemplateSpec struct {
	// Deployment 级别配置
	DeploymentName string            `json:"deployment_name,omitempty"` // Deployment名称
	Namespace      string            `json:"namespace,omitempty"`       // 命名空间
	Replicas       int32             `json:"replicas,omitempty"`        // 副本数
	Labels         map[string]string `json:"labels,omitempty"`          // 标签
	Annotations    map[string]string `json:"annotations,omitempty"`     // 注解

	// Pod 级别配置
	PodLabels      map[string]string `json:"pod_labels,omitempty"`      // Pod标签
	PodAnnotations map[string]string `json:"pod_annotations,omitempty"` // Pod注解
	RestartPolicy  string            `json:"restart_policy,omitempty"`  // 重启策略
	ServiceAccount string            `json:"service_account,omitempty"` // 服务账号
	NodeSelector   map[string]string `json:"node_selector,omitempty"`   // 节点选择器
	Tolerations    []Toleration      `json:"tolerations,omitempty"`     // 容忍度
	Affinity       *Affinity         `json:"affinity,omitempty"`        // 亲和性

	// Container 级别配置
	ContainerName   string           `json:"container_name,omitempty"`    // 容器名称
	Image           string           `json:"image"`                       // 镜像地址
	ImagePullPolicy string           `json:"image_pull_policy,omitempty"` // 镜像拉取策略
	Env             EnvironmentVars  `json:"env,omitempty"`               // 环境变量
	Command         []string         `json:"command,omitempty"`           // 启动命令
	Args            []string         `json:"args,omitempty"`              // 启动参数
	Ports           []ContainerPort  `json:"ports,omitempty"`             // 容器端口
	Resources       ResourceLimits   `json:"resources,omitempty"`         // 资源配额
	VolumeMounts    []VolumeMount    `json:"volume_mounts,omitempty"`     // 卷挂载
	LivenessProbe   *Probe           `json:"liveness_probe,omitempty"`    // 存活检测
	ReadinessProbe  *Probe           `json:"readiness_probe,omitempty"`   // 就绪检测
	StartupProbe    *Probe           `json:"startup_probe,omitempty"`     // 启动探针
	WorkingDir      string           `json:"working_dir,omitempty"`       // 工作目录
	SecurityContext *SecurityContext `json:"security_context,omitempty"`  // 安全上下文

	// Volume 配置
	Volumes    []Volume    `json:"volumes,omitempty"`     // 卷定义
	ConfigMaps []ConfigMap `json:"config_maps,omitempty"` // 配置挂载
	Secrets    []Secret    `json:"secrets,omitempty"`     // 密钥挂载

	// 网络配置
	HostNetwork bool   `json:"host_network,omitempty"` // 使用主机网络
	DNSPolicy   string `json:"dns_policy,omitempty"`   // DNS策略

	// 其他配置
	ImagePullSecrets []string    `json:"image_pull_secrets,omitempty"` // 镜像拉取密钥
	InitContainers   []Container `json:"init_containers,omitempty"`    // 初始化容器
}

// ContainerPort 容器端口
type ContainerPort struct {
	Name          string `json:"name,omitempty"`
	ContainerPort int32  `json:"container_port"`
	Protocol      string `json:"protocol,omitempty"` // TCP, UDP, SCTP
}

// VolumeMount 卷挂载
type VolumeMount struct {
	Name      string `json:"name"`
	MountPath string `json:"mount_path"`
	SubPath   string `json:"sub_path,omitempty"`
	ReadOnly  bool   `json:"read_only,omitempty"`
}

// ConfigMap 配置映射
type ConfigMap struct {
	Name  string          `json:"name"`
	Items []ConfigMapItem `json:"items,omitempty"`
}

// ConfigMapItem 配置项
type ConfigMapItem struct {
	Key  string `json:"key"`
	Path string `json:"path"`
}

// Secret 密钥
type Secret struct {
	Name  string       `json:"name"`
	Items []SecretItem `json:"items,omitempty"`
}

// SecretItem 密钥项
type SecretItem struct {
	Key  string `json:"key"`
	Path string `json:"path"`
}

// Probe 探针配置
type Probe struct {
	InitialDelaySeconds int32            `json:"initial_delay_seconds,omitempty"`
	PeriodSeconds       int32            `json:"period_seconds,omitempty"`
	TimeoutSeconds      int32            `json:"timeout_seconds,omitempty"`
	SuccessThreshold    int32            `json:"success_threshold,omitempty"`
	FailureThreshold    int32            `json:"failure_threshold,omitempty"`
	HTTPGet             *HTTPGetAction   `json:"http_get,omitempty"`
	TCPSocket           *TCPSocketAction `json:"tcp_socket,omitempty"`
	Exec                *ExecAction      `json:"exec,omitempty"`
}

// HTTPGetAction HTTP GET 探针动作
type HTTPGetAction struct {
	Path        string             `json:"path,omitempty"`
	Port        intstr.IntOrString `json:"port"`
	Host        string             `json:"host,omitempty"`
	Scheme      string             `json:"scheme,omitempty"`
	HTTPHeaders []HTTPHeader       `json:"http_headers,omitempty"`
}

// HTTPHeader HTTP 头
type HTTPHeader struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// TCPSocketAction TCP Socket 探针动作
type TCPSocketAction struct {
	Port intstr.IntOrString `json:"port"`
	Host string             `json:"host,omitempty"`
}

// ExecAction 执行探针动作
type ExecAction struct {
	Command []string `json:"command"`
}

// SecurityContext 安全上下文
type SecurityContext struct {
	RunAsUser                *int64        `json:"run_as_user,omitempty"`
	RunAsGroup               *int64        `json:"run_as_group,omitempty"`
	RunAsNonRoot             *bool         `json:"run_as_non_root,omitempty"`
	ReadOnlyRootFilesystem   *bool         `json:"read_only_root_filesystem,omitempty"`
	AllowPrivilegeEscalation *bool         `json:"allow_privilege_escalation,omitempty"`
	Capabilities             *Capabilities `json:"capabilities,omitempty"`
}

// Capabilities 能力
type Capabilities struct {
	Add  []string `json:"add,omitempty"`
	Drop []string `json:"drop,omitempty"`
}

// Toleration 容忍度
type Toleration struct {
	Key               string `json:"key,omitempty"`
	Operator          string `json:"operator,omitempty"` // Equal, Exists
	Value             string `json:"value,omitempty"`
	Effect            string `json:"effect,omitempty"` // NoSchedule, PreferNoSchedule, NoExecute
	TolerationSeconds *int64 `json:"toleration_seconds,omitempty"`
}

// Affinity 亲和性
type Affinity struct {
	NodeAffinity    *NodeAffinity    `json:"node_affinity,omitempty"`
	PodAffinity     *PodAffinity     `json:"pod_affinity,omitempty"`
	PodAntiAffinity *PodAntiAffinity `json:"pod_anti_affinity,omitempty"`
}

// NodeAffinity 节点亲和性
type NodeAffinity struct {
	RequiredDuringSchedulingIgnoredDuringExecution  *NodeSelector             `json:"required_during_scheduling_ignored_during_execution,omitempty"`
	PreferredDuringSchedulingIgnoredDuringExecution []PreferredSchedulingTerm `json:"preferred_during_scheduling_ignored_during_execution,omitempty"`
}

// NodeSelector 节点选择器
type NodeSelector struct {
	NodeSelectorTerms []NodeSelectorTerm `json:"node_selector_terms"`
}

// NodeSelectorTerm 节点选择器条件
type NodeSelectorTerm struct {
	MatchExpressions []NodeSelectorRequirement `json:"match_expressions,omitempty"`
	MatchFields      []NodeSelectorRequirement `json:"match_fields,omitempty"`
}

// NodeSelectorRequirement 节点选择器要求
type NodeSelectorRequirement struct {
	Key      string   `json:"key"`
	Operator string   `json:"operator"` // In, NotIn, Exists, DoesNotExist, Gt, Lt
	Values   []string `json:"values,omitempty"`
}

// PreferredSchedulingTerm 首选调度条件
type PreferredSchedulingTerm struct {
	Weight     int32            `json:"weight"`
	Preference NodeSelectorTerm `json:"preference"`
}

// PodAffinity Pod亲和性
type PodAffinity struct {
	RequiredDuringSchedulingIgnoredDuringExecution  []PodAffinityTerm         `json:"required_during_scheduling_ignored_during_execution,omitempty"`
	PreferredDuringSchedulingIgnoredDuringExecution []WeightedPodAffinityTerm `json:"preferred_during_scheduling_ignored_during_execution,omitempty"`
}

// PodAntiAffinity Pod反亲和性
type PodAntiAffinity struct {
	RequiredDuringSchedulingIgnoredDuringExecution  []PodAffinityTerm         `json:"required_during_scheduling_ignored_during_execution,omitempty"`
	PreferredDuringSchedulingIgnoredDuringExecution []WeightedPodAffinityTerm `json:"preferred_during_scheduling_ignored_during_execution,omitempty"`
}

// PodAffinityTerm Pod亲和性条件
type PodAffinityTerm struct {
	LabelSelector *LabelSelector `json:"label_selector,omitempty"`
	Namespaces    []string       `json:"namespaces,omitempty"`
	TopologyKey   string         `json:"topology_key"`
}

// WeightedPodAffinityTerm 加权Pod亲和性条件
type WeightedPodAffinityTerm struct {
	Weight          int32           `json:"weight"`
	PodAffinityTerm PodAffinityTerm `json:"pod_affinity_term"`
}

// LabelSelector 标签选择器
type LabelSelector struct {
	MatchLabels      map[string]string          `json:"match_labels,omitempty"`
	MatchExpressions []LabelSelectorRequirement `json:"match_expressions,omitempty"`
}

// LabelSelectorRequirement 标签选择器要求
type LabelSelectorRequirement struct {
	Key      string   `json:"key"`
	Operator string   `json:"operator"` // In, NotIn, Exists, DoesNotExist
	Values   []string `json:"values,omitempty"`
}

// Volume 卷定义
type Volume struct {
	Name         string        `json:"name"`
	VolumeSource *VolumeSource `json:"volume_source,omitempty"`
}

// VolumeSource 卷源
type VolumeSource struct {
	EmptyDir  *EmptyDirVolumeSource  `json:"empty_dir,omitempty"`
	HostPath  *HostPathVolumeSource  `json:"host_path,omitempty"`
	ConfigMap *ConfigMapVolumeSource `json:"config_map,omitempty"`
	Secret    *SecretVolumeSource    `json:"secret,omitempty"`
	PVC       *PVCVolumeSource       `json:"persistent_volume_claim,omitempty"`
}

// EmptyDirVolumeSource 空目录卷源
type EmptyDirVolumeSource struct {
	Medium    string `json:"medium,omitempty"`     // "", "Memory"
	SizeLimit string `json:"size_limit,omitempty"` // 例如: "1Gi"
}

// HostPathVolumeSource 主机路径卷源
type HostPathVolumeSource struct {
	Path string `json:"path"`
	Type string `json:"type,omitempty"` // "", "DirectoryOrCreate", "Directory", "FileOrCreate", "File", "Socket", "CharDevice", "BlockDevice"
}

// ConfigMapVolumeSource 配置映射卷源
type ConfigMapVolumeSource struct {
	LocalObjectReference `json:",inline"`
	Items                []KeyToPath `json:"items,omitempty"`
	DefaultMode          *int32      `json:"default_mode,omitempty"`
	Optional             *bool       `json:"optional,omitempty"`
}

// SecretVolumeSource 密钥卷源
type SecretVolumeSource struct {
	SecretName  string      `json:"secret_name"`
	Items       []KeyToPath `json:"items,omitempty"`
	DefaultMode *int32      `json:"default_mode,omitempty"`
	Optional    *bool       `json:"optional,omitempty"`
}

// PVCVolumeSource 持久卷声明卷源
type PVCVolumeSource struct {
	ClaimName string `json:"claim_name"`
	ReadOnly  bool   `json:"read_only,omitempty"`
}

// LocalObjectReference 本地对象引用
type LocalObjectReference struct {
	Name string `json:"name,omitempty"`
}

// KeyToPath 键到路径映射
type KeyToPath struct {
	Key  string `json:"key"`
	Path string `json:"path"`
	Mode *int32 `json:"mode,omitempty"`
}

// Container 容器定义
type Container struct {
	Name            string           `json:"name"`
	Image           string           `json:"image"`
	ImagePullPolicy string           `json:"image_pull_policy,omitempty"`
	Command         []string         `json:"command,omitempty"`
	Args            []string         `json:"args,omitempty"`
	Env             EnvironmentVars  `json:"env,omitempty"`
	Ports           []ContainerPort  `json:"ports,omitempty"`
	Resources       ResourceLimits   `json:"resources,omitempty"`
	VolumeMounts    []VolumeMount    `json:"volume_mounts,omitempty"`
	LivenessProbe   *Probe           `json:"liveness_probe,omitempty"`
	ReadinessProbe  *Probe           `json:"readiness_probe,omitempty"`
	StartupProbe    *Probe           `json:"startup_probe,omitempty"`
	WorkingDir      string           `json:"working_dir,omitempty"`
	SecurityContext *SecurityContext `json:"security_context,omitempty"`
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

// Value 实现driver.Valuer接口
func (t TemplateSpec) Value() (driver.Value, error) {
	return json.Marshal(t)
}

// Scan 实现sql.Scanner接口
func (t *TemplateSpec) Scan(value interface{}) error {
	if value == nil {
		*t = TemplateSpec{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	return json.Unmarshal(bytes, t)
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

// ContainerStatus 容器状态
type ContainerStatus struct {
	Name         string         `json:"name"`
	Image        string         `json:"image"`
	Ready        bool           `json:"ready"`
	Started      bool           `json:"started"`
	RestartCount int32          `json:"restart_count"`
	ContainerID  string         `json:"container_id"`
	State        ContainerState `json:"state"`
}

// ContainerState 容器状态详情
type ContainerState struct {
	Waiting    *ContainerStateWaiting    `json:"waiting,omitempty"`
	Running    *ContainerStateRunning    `json:"running,omitempty"`
	Terminated *ContainerStateTerminated `json:"terminated,omitempty"`
}

// ContainerStateWaiting 容器等待状态
type ContainerStateWaiting struct {
	Reason  string `json:"reason,omitempty"`
	Message string `json:"message,omitempty"`
}

// ContainerStateRunning 容器运行状态
type ContainerStateRunning struct {
	StartedAt time.Time `json:"started_at"`
}

// ContainerStateTerminated 容器终止状态
type ContainerStateTerminated struct {
	ExitCode   int32     `json:"exit_code"`
	Signal     int32     `json:"signal,omitempty"`
	Reason     string    `json:"reason,omitempty"`
	Message    string    `json:"message,omitempty"`
	StartedAt  time.Time `json:"started_at"`
	FinishedAt time.Time `json:"finished_at"`
}

// PodEvent Pod事件
type PodEvent struct {
	Type            string    `json:"type"`
	Reason          string    `json:"reason"`
	Message         string    `json:"message"`
	Count           int32     `json:"count"`
	FirstTimestamp  time.Time `json:"first_timestamp"`
	LastTimestamp   time.Time `json:"last_timestamp"`
	SourceComponent string    `json:"source_component"`
}

// ContainerLog 容器日志
type ContainerLog struct {
	Timestamp time.Time `json:"timestamp"`
	Log       string    `json:"log"`
}

// CreateEphemeralURLRequest 创建URL请求
type CreateEphemeralURLRequest struct {
	Image           string          `json:"image" binding:"required"`
	Env             EnvironmentVars `json:"env"`
	TTLSeconds      int             `json:"ttl_seconds" binding:"required,min=60,max=604800"` // 1分钟到7天
	Replicas        int             `json:"replicas" binding:"min=1,max=10"`
	Resources       ResourceLimits  `json:"resources"`
	ContainerConfig ContainerConfig `json:"container_config"`
	IngressHost     *string         `json:"ingress_host,omitempty"` // 可选，自定义ingress host
}

// CreateEphemeralURLFromTemplateRequest 基于模版创建URL请求
type CreateEphemeralURLFromTemplateRequest struct {
	TemplateID uuid.UUID `json:"template_id" binding:"required"`
	TTLSeconds int       `json:"ttl_seconds" binding:"required,min=60,max=604800"` // 1分钟到7天
	Path       string    `json:"path,omitempty"`                                   // 可选，为空时系统生成
}

// UpdateEphemeralURLRequest 更新URL请求
type UpdateEphemeralURLRequest struct {
	Image           string          `json:"image,omitempty"`
	Env             EnvironmentVars `json:"env,omitempty"`
	TTLSeconds      int             `json:"ttl_seconds,omitempty"`
	Replicas        int             `json:"replicas,omitempty"`
	Resources       ResourceLimits  `json:"resources,omitempty"`
	ContainerConfig ContainerConfig `json:"container_config,omitempty"`
	IngressHost     *string         `json:"ingress_host,omitempty"`
}

// CreateAppTemplateRequest 创建应用模版请求
type CreateAppTemplateRequest struct {
	Name        string       `json:"name" binding:"required,min=1,max=100"`
	Description string       `json:"description"`
	YamlSpec    string       `json:"yaml_spec" binding:"required"`
	ParsedSpec  TemplateSpec `json:"parsed_spec,omitempty"` // 可选，解析后的规格
}

// UpdateAppTemplateRequest 更新应用模版请求
type UpdateAppTemplateRequest struct {
	Name        string        `json:"name" binding:"required,min=1,max=100"`
	Description string        `json:"description"`
	YamlSpec    string        `json:"yaml_spec"`             // 可选，YAML编辑模式时使用
	ParsedSpec  *TemplateSpec `json:"parsed_spec,omitempty"` // 可选，结构化编辑模式时使用
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
