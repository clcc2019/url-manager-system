import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { describe, it, expect } from 'vitest';
import App from './App';

// Mock API service
vi.mock('../services/api', () => ({
  ApiService: {
    getProjects: vi.fn(() => Promise.resolve({ projects: [], total: 0 })),
    healthCheck: vi.fn(() => Promise.resolve({ status: 'ok', service: 'url-manager-system' })),
  },
}));

describe('App', () => {
  it('renders without crashing', () => {
    render(
      <MemoryRouter>
        <App />
      </MemoryRouter>
    );
    
    expect(screen.getByText('URL管理系统')).toBeInTheDocument();
  });

  it('displays project management menu item', () => {
    render(
      <MemoryRouter>
        <App />
      </MemoryRouter>
    );
    
    expect(screen.getByText('项目管理')).toBeInTheDocument();
  });

  it('navigates to projects page by default', () => {
    render(
      <MemoryRouter initialEntries={['/']}>
        <App />
      </MemoryRouter>
    );
    
    // Should redirect to /projects and show project list
    expect(screen.getByText('项目管理')).toBeInTheDocument();
  });
});