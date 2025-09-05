export interface Project {
  id: string;
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
  path: string;
  image: string;
  env: EnvironmentVar[];
  replicas: number;
  resources: ResourceLimits;
  container_config?: ContainerConfig;
  status: 'draft' | 'creating' | 'active' | 'deleting' | 'deleted' | 'failed';
  k8s_deployment_name?: string;
  k8s_service_name?: string;
  k8s_secret_name?: string;
  error_message?: string;
  expire_at: string;
  created_at: string;
  updated_at: string;
  project?: Project;
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
