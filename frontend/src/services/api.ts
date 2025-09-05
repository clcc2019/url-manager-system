import axios from 'axios';
import type {
  Project,
  EphemeralURL,
  CreateProjectRequest,
  CreateURLRequest,
  CreateURLResponse,
  UpdateURLRequest,
  ListProjectsResponse,
  ListURLsResponse,
  PaginationParams,
  AppTemplate,
  CreateTemplateRequest,
  UpdateTemplateRequest,
  CreateURLFromTemplateRequest,
  ListTemplatesResponse,
  TemplateVariablesResponse,
  TemplatePreviewResponse,
  User,
  LoginRequest,
  LoginResponse,
  RegisterRequest,
  ChangePasswordRequest,
} from '../types/api.js';

// 在生产环境中使用相对路径，通过Nginx代理访问后端API
// 在开发环境中使用环境变量或默认的开发服务器地址
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 
  (import.meta.env.PROD ? '/api/v1' : 'http://localhost:8080/api/v1');

// 创建axios实例
const apiClient = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Token管理
export class TokenManager {
  private static readonly TOKEN_KEY = 'auth_token';
  private static readonly USER_KEY = 'auth_user';

  static getToken(): string | null {
    return localStorage.getItem(this.TOKEN_KEY);
  }

  static setToken(token: string): void {
    localStorage.setItem(this.TOKEN_KEY, token);
  }

  static removeToken(): void {
    localStorage.removeItem(this.TOKEN_KEY);
  }

  static getUser(): User | null {
    const userStr = localStorage.getItem(this.USER_KEY);
    return userStr ? JSON.parse(userStr) : null;
  }

  static setUser(user: User): void {
    localStorage.setItem(this.USER_KEY, JSON.stringify(user));
  }

  static removeUser(): void {
    localStorage.removeItem(this.USER_KEY);
  }

  static clear(): void {
    this.removeToken();
    this.removeUser();
  }
}

// 请求拦截器 - 自动添加Authorization header
apiClient.interceptors.request.use(
  (config) => {
    const token = TokenManager.getToken();
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// 响应拦截器
apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    // 如果是401未授权错误，清除token并跳转到登录页
    if (error.response?.status === 401) {
      TokenManager.clear();
      // 触发自定义事件，通知应用需要重新登录
      window.dispatchEvent(new CustomEvent('auth:unauthorized'));
    }
    console.error('API Error:', error.response?.data || error.message);
    return Promise.reject(error);
  }
);

export class ApiService {
  // 项目管理 API
  static async getProjects(params?: PaginationParams): Promise<ListProjectsResponse> {
    const response = await apiClient.get('/projects', {
      params,
    });
    return response.data;
  }

  static async getProject(id: string): Promise<Project> {
    const response = await apiClient.get(`/projects/${id}`);
    return response.data;
  }

  static async createProject(data: CreateProjectRequest): Promise<Project> {
    const response = await apiClient.post('/projects', data);
    return response.data;
  }

  static async updateProject(id: string, data: CreateProjectRequest): Promise<Project> {
    const response = await apiClient.put(`/projects/${id}`, data);
    return response.data;
  }

  static async deleteProject(id: string): Promise<void> {
    await apiClient.delete(`/projects/${id}`);
  }

  // URL管理 API
  static async getProjectURLs(
    projectId: string,
    params?: PaginationParams
  ): Promise<ListURLsResponse> {
    const response = await apiClient.get(
      `/projects/${projectId}/urls`,
      { params }
    );
    return response.data;
  }

  static async createURL(
    projectId: string,
    data: CreateURLRequest
  ): Promise<CreateURLResponse> {
    const response = await apiClient.post(
      `/projects/${projectId}/urls`,
      data
    );
    return response.data;
  }

  static async createURLFromTemplate(
    projectId: string,
    data: CreateURLFromTemplateRequest
  ): Promise<CreateURLResponse> {
    const response = await apiClient.post(
      `/projects/${projectId}/urls/from-template`,
      data
    );
    return response.data;
  }

  static async getURL(id: string): Promise<EphemeralURL> {
    const response = await apiClient.get(`/urls/${id}`);
    return response.data;
  }

  static async deleteURL(id: string): Promise<void> {
    await apiClient.delete(`/urls/${id}`);
  }

  static async updateURL(
    id: string,
    data: UpdateURLRequest
  ): Promise<EphemeralURL> {
    const response = await apiClient.put(`/urls/${id}`, data);
    return response.data;
  }

  static async deployURL(id: string): Promise<void> {
    await apiClient.post(`/urls/${id}/deploy`);
  }

  // 健康检查
  static async healthCheck(): Promise<{ status: string; service: string }> {
    const response = await apiClient.get('/health');
    return response.data;
  }

  // 数据校验和清理
  static async validateAndCleanupData(): Promise<{ message: string }> {
    const response = await apiClient.post('/urls/validate-cleanup');
    return response.data;
  }

  // 模版管理 API
  static async getTemplates(params?: PaginationParams): Promise<ListTemplatesResponse> {
    const response = await apiClient.get('/templates', {
      params,
    });
    return response.data;
  }

  static async getTemplate(id: string): Promise<AppTemplate> {
    const response = await apiClient.get(`/templates/${id}`);
    return response.data;
  }

  static async createTemplate(data: CreateTemplateRequest): Promise<AppTemplate> {
    const response = await apiClient.post('/templates', data);
    return response.data;
  }

  static async updateTemplate(id: string, data: UpdateTemplateRequest): Promise<AppTemplate> {
    const response = await apiClient.put(`/templates/${id}`, data);
    return response.data;
  }

  static async deleteTemplate(id: string): Promise<void> {
    await apiClient.delete(`/templates/${id}`);
  }

  static async getTemplateVariables(id: string): Promise<TemplateVariablesResponse> {
    const response = await apiClient.get(`/templates/${id}/variables`);
    return response.data;
  }

  static async previewTemplate(
    id: string, 
    variables?: Record<string, string>
  ): Promise<TemplatePreviewResponse> {
    const response = await apiClient.post(`/templates/${id}/preview`, variables || {});
    return response.data;
  }

  // 认证API
  static async login(data: LoginRequest): Promise<LoginResponse> {
    const response = await apiClient.post('/auth/login', data);
    return response.data;
  }

  static async logout(): Promise<void> {
    await apiClient.post('/auth/logout');
  }

  static async register(data: RegisterRequest): Promise<User> {
    const response = await apiClient.post('/users/register', data);
    return response.data;
  }

  static async getProfile(): Promise<User> {
    const response = await apiClient.get('/users/profile');
    return response.data;
  }

  static async changePassword(data: ChangePasswordRequest): Promise<void> {
    await apiClient.put('/users/password', data);
  }

  static async getUsers(params?: PaginationParams): Promise<{ users: User[]; total: number }> {
    const response = await apiClient.get('/users', { params });
    return response.data;
  }
}