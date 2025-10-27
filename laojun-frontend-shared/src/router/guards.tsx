import React, { useEffect } from 'react';
import { Navigate, useLocation } from 'react-router-dom';
import { User, UserRole } from '../types';

// 路由守卫配置
export interface RouteGuardConfig {
  requireAuth?: boolean;
  roles?: UserRole[];
  redirectTo?: string;
  onUnauthorized?: () => void;
}

// 认证守卫组件属性
export interface AuthGuardProps {
  children: React.ReactNode;
  user?: User | null;
  isLoading?: boolean;
  config?: RouteGuardConfig;
}

// 认证守卫组件
export const AuthGuard: React.FC<AuthGuardProps> = ({
  children,
  user,
  isLoading = false,
  config = {},
}) => {
  const location = useLocation();
  const {
    requireAuth = true,
    roles = [],
    redirectTo = '/login',
    onUnauthorized,
  } = config;

  useEffect(() => {
    // 如果需要认证但用户未登录，执行未授权回调
    if (requireAuth && !isLoading && !user) {
      onUnauthorized?.();
    }
  }, [requireAuth, isLoading, user, onUnauthorized]);

  // 加载中状态
  if (isLoading) {
    return (
      <div style={{ 
        display: 'flex', 
        justifyContent: 'center', 
        alignItems: 'center', 
        height: '100vh' 
      }}>
        <div>加载中...</div>
      </div>
    );
  }

  // 不需要认证，直接渲染
  if (!requireAuth) {
    return <>{children}</>;
  }

  // 需要认证但用户未登录
  if (!user) {
    return (
      <Navigate
        to={redirectTo}
        state={{ from: location }}
        replace
      />
    );
  }

  // 检查角色权限
  if (roles.length > 0 && !roles.includes(user.role)) {
    return (
      <Navigate
        to="/403"
        state={{ from: location }}
        replace
      />
    );
  }

  // 权限检查通过，渲染子组件
  return <>{children}</>;
};

// 角色守卫组件属性
export interface RoleGuardProps {
  children: React.ReactNode;
  user?: User | null;
  allowedRoles: UserRole[];
  fallback?: React.ReactNode;
}

// 角色守卫组件
export const RoleGuard: React.FC<RoleGuardProps> = ({
  children,
  user,
  allowedRoles,
  fallback = <div>没有权限访问此内容</div>,
}) => {
  if (!user || !allowedRoles.includes(user.role)) {
    return <>{fallback}</>;
  }

  return <>{children}</>;
};

// 登录状态守卫组件属性
export interface LoginGuardProps {
  children: React.ReactNode;
  user?: User | null;
  redirectTo?: string;
}

// 登录状态守卫（已登录用户重定向）
export const LoginGuard: React.FC<LoginGuardProps> = ({
  children,
  user,
  redirectTo = '/',
}) => {
  const location = useLocation();

  // 如果用户已登录，重定向到指定页面
  if (user) {
    const from = (location.state as any)?.from?.pathname || redirectTo;
    return <Navigate to={from} replace />;
  }

  return <>{children}</>;
};

// 权限检查 Hook
export interface UsePermissionOptions {
  user?: User | null;
  requiredRoles?: UserRole[];
  requiredPermissions?: string[];
}

export const usePermission = (options: UsePermissionOptions = {}) => {
  const { user, requiredRoles = [], requiredPermissions = [] } = options;

  const hasRole = (role: UserRole): boolean => {
    return user?.role === role;
  };

  const hasAnyRole = (roles: UserRole[]): boolean => {
    return roles.length === 0 || (user ? roles.includes(user.role) : false);
  };

  const hasPermission = (permission: string): boolean => {
    // 这里可以根据实际的权限系统实现
    // 目前简化为基于角色的权限检查
    if (!user) return false;
    
    // 管理员拥有所有权限
    if (user.role === UserRole.ADMIN) return true;
    
    // 根据具体权限进行检查
    // 这里需要根据实际的权限系统进行实现
    // 暂时返回true，实际应该根据permission参数进行权限检查
    console.debug('Checking permission:', permission);
    return true;
  };

  const hasAnyPermission = (permissions: string[]): boolean => {
    return permissions.length === 0 || permissions.some(permission => hasPermission(permission));
  };

  const canAccess = (): boolean => {
    return hasAnyRole(requiredRoles) && hasAnyPermission(requiredPermissions);
  };

  return {
    user,
    hasRole,
    hasAnyRole,
    hasPermission,
    hasAnyPermission,
    canAccess,
    isAuthenticated: !!user,
    isAdmin: user?.role === UserRole.ADMIN,
    isDeveloper: user?.role === UserRole.DEVELOPER,
    isUser: user?.role === UserRole.USER,
  };
};

// 条件渲染组件
export interface ConditionalRenderProps {
  children: React.ReactNode;
  condition: boolean;
  fallback?: React.ReactNode;
}

export const ConditionalRender: React.FC<ConditionalRenderProps> = ({
  children,
  condition,
  fallback = null,
}) => {
  return condition ? <>{children}</> : <>{fallback}</>;
};

// 权限包装组件
export interface PermissionWrapperProps {
  children: React.ReactNode;
  user?: User | null;
  roles?: UserRole[];
  permissions?: string[];
  fallback?: React.ReactNode;
  mode?: 'hide' | 'disable' | 'replace';
}

export const PermissionWrapper: React.FC<PermissionWrapperProps> = ({
  children,
  user,
  roles = [],
  permissions = [],
  fallback = null,
  mode = 'hide',
}) => {
  const { canAccess } = usePermission({
    user,
    requiredRoles: roles,
    requiredPermissions: permissions,
  });

  if (!canAccess()) {
    switch (mode) {
      case 'hide':
        return <>{fallback}</>;
      case 'disable':
        return (
          <div style={{ opacity: 0.5, pointerEvents: 'none' }}>
            {children}
          </div>
        );
      case 'replace':
        return <>{fallback}</>;
      default:
        return <>{fallback}</>;
    }
  }

  return <>{children}</>;
};