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

  // 所有登录用户都有所有权限，不再检查角色限制
  return <>{children}</>;
}

// 便捷的权限检查hooks
export function useRoleCheck() {
  const { user } = useAuth();

  return {
    // 所有用户都被视为管理员
    isAdmin: !!user,
    isUser: !!user,
    hasRole: (role: 'admin' | 'user') => !!user,
    hasAnyRole: (roles: ('admin' | 'user')[]) => !!user,
  };
}