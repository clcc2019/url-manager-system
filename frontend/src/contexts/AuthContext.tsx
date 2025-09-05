import React, { createContext, useContext, useEffect, useState } from 'react';
import type { ReactNode } from 'react';
import { message } from 'antd';
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
      console.log('AuthContext: 初始化认证状态...');
      const savedToken = TokenManager.getToken();
      const savedUser = TokenManager.getUser();
      console.log('AuthContext: 本地存储', { hasToken: !!savedToken, hasUser: !!savedUser, userRole: savedUser?.role });

      if (savedToken && savedUser) {
        setToken(savedToken);
        setUser(savedUser);
        console.log('AuthContext: 设置初始状态', { userRole: savedUser.role });
        
        // 验证token是否仍然有效
        try {
          const currentUser = await ApiService.getProfile();
          console.log('AuthContext: 验证成功', { userRole: currentUser.role });
          // 更新用户信息（可能有变更）
          setUser(currentUser);
          TokenManager.setUser(currentUser);
        } catch (error) {
          console.warn('AuthContext: Token validation failed:', error);
          // 只有在明确是认证错误时才清除状态
          const isAuthError = (error as any)?.response?.status === 401;
          if (isAuthError) {
            console.log('AuthContext: 401错误，清除认证状态');
            TokenManager.clear();
            setToken(null);
            setUser(null);
          } else {
            console.log('AuthContext: 非认证错误，保持本地状态');
          }
          // 其他网络错误不清除状态，保持本地状态
        }
      } else {
        console.log('AuthContext: 没有本地认证信息');
      }
      
      console.log('AuthContext: 初始化完成');
      setIsLoading(false);
    };

    initAuth();
  }, []);

  // 监听401错误，自动登出
  useEffect(() => {
    const handleUnauthorized = () => {
      logout();
      message.error('登录已过期，请重新登录');
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
      
      message.success('登录成功');
    } catch (error: any) {
      const errorMessage = error.response?.data?.error || '登录失败';
      message.error(errorMessage);
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
      message.success('已登出');
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