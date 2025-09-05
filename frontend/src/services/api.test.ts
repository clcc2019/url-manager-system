import { ApiService } from './api';
import { describe, it, expect, vi, beforeEach } from 'vitest';

// Mock axios
vi.mock('axios', () => ({
  default: {
    create: vi.fn(() => ({
      get: vi.fn(),
      post: vi.fn(),
      put: vi.fn(),
      delete: vi.fn(),
      interceptors: {
        response: {
          use: vi.fn(),
        },
      },
    })),
  },
}));

describe('ApiService', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Project APIs', () => {
    it('should have getProjects method', () => {
      expect(typeof ApiService.getProjects).toBe('function');
    });

    it('should have createProject method', () => {
      expect(typeof ApiService.createProject).toBe('function');
    });

    it('should have updateProject method', () => {
      expect(typeof ApiService.updateProject).toBe('function');
    });

    it('should have deleteProject method', () => {
      expect(typeof ApiService.deleteProject).toBe('function');
    });
  });

  describe('URL APIs', () => {
    it('should have getProjectURLs method', () => {
      expect(typeof ApiService.getProjectURLs).toBe('function');
    });

    it('should have createURL method', () => {
      expect(typeof ApiService.createURL).toBe('function');
    });

    it('should have getURL method', () => {
      expect(typeof ApiService.getURL).toBe('function');
    });

    it('should have deleteURL method', () => {
      expect(typeof ApiService.deleteURL).toBe('function');
    });
  });

  describe('Health Check', () => {
    it('should have healthCheck method', () => {
      expect(typeof ApiService.healthCheck).toBe('function');
    });
  });
});