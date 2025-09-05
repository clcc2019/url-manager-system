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
}

export interface CreateURLResponse {
  url: string;
  id: string;
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
export interface AppTemplate {
  id: string;
  user_id: string;
  name: string;
  description: string;
  yaml_spec: string;
  created_at: string;
  updated_at: string;
}

export interface CreateTemplateRequest {
  name: string;
  description?: string;
  yaml_spec: string;
}

export interface UpdateTemplateRequest {
  name: string;
  description?: string;
  yaml_spec: string;
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
