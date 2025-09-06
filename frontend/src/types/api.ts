export interface Project {
  id: string;
  user_id: string;
  name: string;
  description: string;
  created_at: string;
  updated_at: string;
}

export interface EnvironmentVar {
  name: string;
  value: string;
}

export interface LogEntry {
  timestamp: string;
  level: 'info' | 'warn' | 'error';
  message: string;
  details?: string;
}

export interface ContainerStatus {
  name: string;
  ready: boolean;
  restart_count: number;
  image: string;
  image_id: string;
  container_id?: string;
  started: boolean;
  state: ContainerState;
  last_state?: ContainerState;
}

export interface ContainerState {
  waiting?: ContainerStateWaiting;
  running?: ContainerStateRunning;
  terminated?: ContainerStateTerminated;
}

export interface ContainerStateWaiting {
  reason?: string;
  message?: string;
}

export interface ContainerStateRunning {
  started_at?: string;
}

export interface ContainerStateTerminated {
  exit_code: number;
  signal?: number;
  reason?: string;
  message?: string;
  started_at?: string;
  finished_at?: string;
  container_id?: string;
}

export interface PodEvent {
  type: string;
  reason: string;
  message: string;
  first_timestamp: string;
  last_timestamp: string;
  count: number;
  source_component?: string;
  source_host?: string;
}

export interface ContainerLog {
  timestamp: string;
  log: string;
}

export interface DeviceMapping {
  host_path: string;
  container_path: string;
  permissions?: string;
}

export interface ContainerConfig {
  container_name?: string;
  devices?: DeviceMapping[];
  command?: string[];
  args?: string[];
  working_dir?: string;
  tty?: boolean;
  stdin?: boolean;
}

export interface ResourceRequests {
  cpu: string;
  memory: string;
}

export interface ResourceLimits {
  requests: ResourceRequests;
  limits: ResourceRequests;
}

export interface EphemeralURL {
  id: string;
  project_id: string;
  template_id?: string;
  path: string;
  image: string;
  env: EnvironmentVar[];
  replicas: number;
  resources: ResourceLimits;
  container_config?: ContainerConfig;
  status: 'draft' | 'creating' | 'waiting' | 'active' | 'deleting' | 'deleted' | 'failed';
  ttl_seconds: number;
  k8s_deployment_name?: string;
  k8s_service_name?: string;
  k8s_secret_name?: string;
  error_message?: string;
  logs?: LogEntry[];
  container_statuses?: ContainerStatus[];
  pod_events?: PodEvent[];
  container_logs?: ContainerLog[];
  ingress_host?: string;
  started_at?: string;
  expire_at: string;
  created_at: string;
  updated_at: string;
  project?: Project;
  template?: AppTemplate;
}

export interface CreateProjectRequest {
  name: string;
  description?: string;
}

export interface CreateURLRequest {
  image: string;
  env?: EnvironmentVar[];
  ttl_seconds: number;
  replicas?: number;
  resources?: ResourceLimits;
  container_config?: ContainerConfig;
  ingress_host?: string;
}

export interface CreateURLResponse {
  url: string;
  id: string;
}

export interface UpdateURLRequest {
  image?: string;
  env?: EnvironmentVar[];
  ttl_seconds?: number;
  replicas?: number;
  resources?: ResourceLimits;
  container_config?: ContainerConfig;
  ingress_host?: string;
}

export interface ListProjectsResponse {
  projects: Project[];
  total: number;
}

export interface ListURLsResponse {
  urls: EphemeralURL[];
  total: number;
}

export interface ApiResponse<T> {
  data?: T;
  error?: string;
}

export interface PaginationParams {
  limit?: number;
  offset?: number;
}

// 应用模版相关类型
export interface TemplateSpec {
  // Deployment 级别配置
  deployment_name?: string;
  namespace?: string;
  replicas?: number;
  labels?: Record<string, string>;
  annotations?: Record<string, string>;
  
  // Pod 级别配置
  pod_labels?: Record<string, string>;
  pod_annotations?: Record<string, string>;
  restart_policy?: string;
  service_account?: string;
  node_selector?: Record<string, string>;
  tolerations?: Toleration[];
  affinity?: Affinity;
  
  // Container 级别配置
  container_name?: string;
  image: string;
  image_pull_policy?: string;
  env?: EnvironmentVar[];
  command?: string[];
  args?: string[];
  ports?: ContainerPort[];
  resources?: ResourceLimits;
  volume_mounts?: VolumeMount[];
  liveness_probe?: Probe;
  readiness_probe?: Probe;
  startup_probe?: Probe;
  working_dir?: string;
  security_context?: SecurityContext;
  
  // Volume 配置
  volumes?: Volume[];
  config_maps?: ConfigMap[];
  secrets?: Secret[];
  
  // 网络配置
  host_network?: boolean;
  dns_policy?: string;
  
  // 其他配置
  image_pull_secrets?: string[];
  init_containers?: Container[];
}

export interface ContainerPort {
  name?: string;
  container_port: number;
  protocol?: string;
}

export interface VolumeMount {
  name: string;
  mount_path: string;
  sub_path?: string;
  read_only?: boolean;
}

export interface ConfigMap {
  name: string;
  items?: ConfigMapItem[];
}

export interface ConfigMapItem {
  key: string;
  path: string;
}

export interface Secret {
  name: string;
  items?: SecretItem[];
}

export interface SecretItem {
  key: string;
  path: string;
}

export interface Probe {
  initial_delay_seconds?: number;
  period_seconds?: number;
  timeout_seconds?: number;
  success_threshold?: number;
  failure_threshold?: number;
  http_get?: HTTPGetAction;
  tcp_socket?: TCPSocketAction;
  exec?: ExecAction;
}

export interface HTTPGetAction {
  path?: string;
  port: string | number;
  host?: string;
  scheme?: string;
  http_headers?: HTTPHeader[];
}

export interface HTTPHeader {
  name: string;
  value: string;
}

export interface TCPSocketAction {
  port: string | number;
  host?: string;
}

export interface ExecAction {
  command: string[];
}

export interface SecurityContext {
  run_as_user?: number;
  run_as_group?: number;
  run_as_non_root?: boolean;
  read_only_root_filesystem?: boolean;
  allow_privilege_escalation?: boolean;
  capabilities?: Capabilities;
}

export interface Capabilities {
  add?: string[];
  drop?: string[];
}

// 新增的类型定义
export interface Toleration {
  key?: string;
  operator?: string;
  value?: string;
  effect?: string;
  toleration_seconds?: number;
}

export interface Affinity {
  node_affinity?: NodeAffinity;
  pod_affinity?: PodAffinity;
  pod_anti_affinity?: PodAntiAffinity;
}

export interface NodeAffinity {
  required_during_scheduling_ignored_during_execution?: NodeSelector;
  preferred_during_scheduling_ignored_during_execution?: PreferredSchedulingTerm[];
}

export interface NodeSelector {
  node_selector_terms: NodeSelectorTerm[];
}

export interface NodeSelectorTerm {
  match_expressions?: NodeSelectorRequirement[];
  match_fields?: NodeSelectorRequirement[];
}

export interface NodeSelectorRequirement {
  key: string;
  operator: string;
  values?: string[];
}

export interface PreferredSchedulingTerm {
  weight: number;
  preference: NodeSelectorTerm;
}

export interface PodAffinity {
  required_during_scheduling_ignored_during_execution?: PodAffinityTerm[];
  preferred_during_scheduling_ignored_during_execution?: WeightedPodAffinityTerm[];
}

export interface PodAntiAffinity {
  required_during_scheduling_ignored_during_execution?: PodAffinityTerm[];
  preferred_during_scheduling_ignored_during_execution?: WeightedPodAffinityTerm[];
}

export interface PodAffinityTerm {
  label_selector?: LabelSelector;
  namespaces?: string[];
  topology_key: string;
}

export interface WeightedPodAffinityTerm {
  weight: number;
  pod_affinity_term: PodAffinityTerm;
}

export interface LabelSelector {
  match_labels?: Record<string, string>;
  match_expressions?: LabelSelectorRequirement[];
}

export interface LabelSelectorRequirement {
  key: string;
  operator: string;
  values?: string[];
}

export interface Volume {
  name: string;
  volume_source?: VolumeSource;
}

export interface VolumeSource {
  empty_dir?: EmptyDirVolumeSource;
  host_path?: HostPathVolumeSource;
  config_map?: ConfigMapVolumeSource;
  secret?: SecretVolumeSource;
  persistent_volume_claim?: PVCVolumeSource;
}

export interface EmptyDirVolumeSource {
  medium?: string;
  size_limit?: string;
}

export interface HostPathVolumeSource {
  path: string;
  type?: string;
}

export interface ConfigMapVolumeSource {
  name?: string;
  items?: KeyToPath[];
  default_mode?: number;
  optional?: boolean;
}

export interface SecretVolumeSource {
  secret_name: string;
  items?: KeyToPath[];
  default_mode?: number;
  optional?: boolean;
}

export interface PVCVolumeSource {
  claim_name: string;
  read_only?: boolean;
}

export interface KeyToPath {
  key: string;
  path: string;
  mode?: number;
}

export interface Container {
  name: string;
  image: string;
  image_pull_policy?: string;
  command?: string[];
  args?: string[];
  env?: EnvironmentVar[];
  ports?: ContainerPort[];
  resources?: ResourceLimits;
  volume_mounts?: VolumeMount[];
  liveness_probe?: Probe;
  readiness_probe?: Probe;
  startup_probe?: Probe;
  working_dir?: string;
  security_context?: SecurityContext;
}

export interface AppTemplate {
  id: string;
  user_id: string;
  name: string;
  description: string;
  yaml_spec: string;
  parsed_spec: TemplateSpec;
  created_at: string;
  updated_at: string;
}

export interface CreateTemplateRequest {
  name: string;
  description?: string;
  yaml_spec: string;
  parsed_spec?: TemplateSpec;
}

export interface UpdateTemplateRequest {
  name: string;
  description?: string;
  yaml_spec?: string;
  parsed_spec?: TemplateSpec;
}

export interface CreateURLFromTemplateRequest {
  template_id: string;
  ttl_seconds: number;
  path?: string;
  variables?: Record<string, string>;
}

export interface ListTemplatesResponse {
  templates: AppTemplate[];
  total: number;
}

export interface TemplateVariablesResponse {
  variables: string[];
}

export interface TemplatePreviewResponse {
  processed_yaml: string;
  variables: Record<string, string>;
}

// 认证相关类型
export interface User {
  id: string;
  username: string;
  role: 'admin' | 'user';
  email?: string;
  created_at: string;
  updated_at: string;
  last_login_at?: string;
}

export interface LoginRequest {
  username: string;
  password: string;
}

export interface LoginResponse {
  token: string;
  user: User;
}

export interface RegisterRequest {
  username: string;
  password: string;
  email?: string;
}

export interface ChangePasswordRequest {
  old_password: string;
  new_password: string;
}

export interface AuthContextType {
  user: User | null;
  token: string | null;
  login: (username: string, password: string) => Promise<void>;
  logout: () => void;
  isLoading: boolean;
  isAuthenticated: boolean;
}
