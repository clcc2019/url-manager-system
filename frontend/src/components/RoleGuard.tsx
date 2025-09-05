import React from 'react';
import { useAuth } from '../contexts/AuthContext';

interface RoleGuardProps {
  children: React.ReactNode;
  allowedRoles?: ('admin' | 'user')[];
  requireAdmin?: boolean;
  fallback?: React.ReactNode;
}

export default function RoleGuard({ 
  children, 
  allowedRoles, 
  requireAdmin = false, 
  fallback = null 
}: RoleGuardProps) {
  const { user } = useAuth();

  if (!user) {
    return <>{fallback}</>;
  }

  // 如果需要管理员权限但用户不是管理员
  if (requireAdmin && user.role !== 'admin') {
    return <>{fallback}</>;
  }

  // 如果指定了允许的角色列表
  if (allowedRoles && !allowedRoles.includes(user.role)) {
    return <>{fallback}</>;
  }

  return <>{children}</>;
}

// 便捷的权限检查hooks
export function useRoleCheck() {
  const { user } = useAuth();
  
  const isAdmin = user?.role === 'admin';
  
  // 添加调试日志
  React.useEffect(() => {
    console.log('RoleGuard: useRoleCheck', { user: user?.username, role: user?.role, isAdmin });
  }, [user, isAdmin]);

  return {
    isAdmin,
    isUser: user?.role === 'user',
    hasRole: (role: 'admin' | 'user') => user?.role === role,
    hasAnyRole: (roles: ('admin' | 'user')[]) => user ? roles.includes(user.role) : false,
  };
}