import React, { createContext, useContext, useEffect, useState } from 'react';
import type { ReactNode } from 'react';
import { toast } from '@/hooks/use-toast';
import { ApiService, TokenManager } from '../services/api';
import type { User, AuthContextType } from '../types/api';

const AuthContext = createContext<AuthContextType | undefined>(undefined);

interface AuthProviderProps {
  children: ReactNode;
}

export function AuthProvider({ children }: AuthProviderProps) {
  const [user, setUser] = useState<User | null>(null);
  const [token, setToken] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  // 初始化：检查本地存储的token和用户信息
  useEffect(() => {
    const initAuth = async () => {
      const savedToken = TokenManager.getToken();
      const savedUser = TokenManager.getUser();

      if (savedToken && savedUser) {
        // 立即设置本地状态，不等待API验证
        setToken(savedToken);
        setUser(savedUser);

        // 异步验证token是否仍然有效，但不阻塞UI渲染
        try {
          const currentUser = await ApiService.getProfile();
          // 更新用户信息（可能有变更）
          setUser(currentUser);
          TokenManager.setUser(currentUser);
        } catch (error) {
          console.warn('Token validation failed:', error);
          // 只有在明确是认证错误时才清除状态
          const isAuthError = (error as any)?.response?.status === 401;
          if (isAuthError) {
            TokenManager.clear();
            setToken(null);
            setUser(null);
          }
          // 其他网络错误不清除状态，保持本地状态
        }
      }

      // 无论如何都要设置加载完成状态，确保UI能正常渲染
      setIsLoading(false);
    };

    // 添加超时机制，确保即使网络问题也能正常渲染
    const timeoutId = setTimeout(() => {
      if (isLoading) {
        console.warn('Authentication initialization timeout, proceeding with current state');
        setIsLoading(false);
      }
    }, 3000); // 3秒超时

    initAuth();

    return () => clearTimeout(timeoutId);
  }, []);

  // 监听401错误，自动登出
  useEffect(() => {
    const handleUnauthorized = () => {
      logout();
      toast({
        title: '登录已过期',
        description: '请重新登录',
        variant: 'destructive',
      });
    };

    window.addEventListener('auth:unauthorized', handleUnauthorized);
    
    return () => {
      window.removeEventListener('auth:unauthorized', handleUnauthorized);
    };
  }, []);

  const login = async (username: string, password: string): Promise<void> => {
    try {
      setIsLoading(true);
      const response = await ApiService.login({ username, password });
      
      setToken(response.token);
      setUser(response.user);
      TokenManager.setToken(response.token);
      TokenManager.setUser(response.user);
      
      toast({
        title: '登录成功',
        description: '欢迎回来！',
      });
    } catch (error: any) {
      const errorMessage = error.response?.data?.error || '登录失败';
      toast({
        title: '登录失败',
        description: errorMessage,
        variant: 'destructive',
      });
      throw error;
    } finally {
      setIsLoading(false);
    }
  };

  const logout = async (): Promise<void> => {
    try {
      // 尝试调用后端登出API
      if (token) {
        await ApiService.logout();
      }
    } catch (error) {
      console.warn('Logout API call failed:', error);
    } finally {
      // 无论API调用是否成功，都清除本地状态
      setToken(null);
      setUser(null);
      TokenManager.clear();
      toast({
        title: '退出成功',
        description: '您已安全退出系统',
      });
    }
  };

  const value: AuthContextType = {
    user,
    token,
    login,
    logout,
    isLoading,
    isAuthenticated: !!token && !!user,
  };

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth(): AuthContextType {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}

// 导出默认的Context以便测试
export { AuthContext };