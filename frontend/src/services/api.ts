import axios from 'axios';
import type {
  Project,
  EphemeralURL,
  CreateProjectRequest,
  CreateURLRequest,
  CreateURLResponse,
  ListProjectsResponse,
  ListURLsResponse,
  PaginationParams,
} from '../types/api.js';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://127.0.0.1:8080/api/v1';

// 创建axios实例
const apiClient = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

// 响应拦截器
apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
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

  static async getURL(id: string): Promise<EphemeralURL> {
    const response = await apiClient.get(`/urls/${id}`);
    return response.data;
  }

  static async deleteURL(id: string): Promise<void> {
    await apiClient.delete(`/urls/${id}`);
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
}