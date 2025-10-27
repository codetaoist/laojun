// 导出路由配置
export {
  adminRoutes,
  marketplaceRoutes,
  createRouteConfig,
  checkRoutePermission,
  generateBreadcrumb,
} from './config';
export type { BaseRouteConfig, RouteConfigOptions } from './config';

// 导出路由守卫
export {
  AuthGuard,
  RoleGuard,
  LoginGuard,
  usePermission,
  ConditionalRender,
  PermissionWrapper,
} from './guards';
export type {
  RouteGuardConfig,
  AuthGuardProps,
  RoleGuardProps,
  LoginGuardProps,
  UsePermissionOptions,
  ConditionalRenderProps,
  PermissionWrapperProps,
} from './guards';

// 导出类型
export type { RouteConfig, RouteMeta, UserRole } from '@/types';