import { createBrowserRouter, Navigate } from 'react-router-dom';
import { lazy, Suspense } from 'react';
import { Spin } from 'antd';
import Layout from '@/components/Layout';
import AuthGuard from '@/components/AuthGuard';
import ErrorBoundary from '@/components/ErrorBoundary';

// 懒加载组件
const Login = lazy(() => import('@/pages/Login'));
const Dashboard = lazy(() => import('@/pages/Dashboard'));
const UserManagement = lazy(() => import('@/pages/UserManagement'));
const RoleManagement = lazy(() => import('@/pages/RoleManagement'));
const PermissionManagement = lazy(() => import('@/pages/PermissionManagement'));
const MenuManagement = lazy(() => import('@/pages/MenuManagement'));
const PluginManagement = lazy(() => import('@/pages/PluginManagement'));
const PluginReviewManagement = lazy(() => import('@/pages/PluginReviewManagement'));
const SystemSettings = lazy(() => import('@/pages/SystemSettings'));
const Profile = lazy(() => import('@/pages/Profile'));
const NotFound = lazy(() => import('@/pages/NotFound'));

// 加载组件包装器
const LoadingWrapper = ({ children }: { children: React.ReactNode }) => (
  <Suspense
    fallback={
      <div style={{ 
        display: 'flex', 
        justifyContent: 'center', 
        alignItems: 'center', 
        height: '200px' 
      }}>
        <Spin size="large" />
      </div>
    }
  >
    <ErrorBoundary>
      {children}
    </ErrorBoundary>
  </Suspense>
);

export const router = createBrowserRouter([
  {
    path: '/login',
    element: (
      <LoadingWrapper>
        <Login />
      </LoadingWrapper>
    ),
  },
  {
    path: '/',
    element: (
      <AuthGuard>
        <Layout />
      </AuthGuard>
    ),
    children: [
      {
        index: true,
        element: <Navigate to="/dashboard" replace />,
      },
      {
        path: 'dashboard',
        element: (
          <LoadingWrapper>
            <Dashboard />
          </LoadingWrapper>
        ),
      },
      {
        path: 'users',
        element: (
          <LoadingWrapper>
            <UserManagement />
          </LoadingWrapper>
        ),
      },
      {
        path: 'roles',
        element: (
          <LoadingWrapper>
            <RoleManagement />
          </LoadingWrapper>
        ),
      },
      {
        path: 'permissions',
        element: (
          <LoadingWrapper>
            <PermissionManagement />
          </LoadingWrapper>
        ),
      },
      {
        path: 'menus',
        element: (
          <LoadingWrapper>
            <MenuManagement />
          </LoadingWrapper>
        ),
      },
      {
        path: 'plugins',
        element: (
          <LoadingWrapper>
            <PluginManagement />
          </LoadingWrapper>
        ),
      },
      {
        path: 'plugin-review',
        element: (
          <LoadingWrapper>
            <PluginReviewManagement />
          </LoadingWrapper>
        ),
      },
      {
        path: 'settings',
        element: (
          <LoadingWrapper>
            <SystemSettings />
          </LoadingWrapper>
        ),
      },
      {
        path: 'profile',
        element: (
          <LoadingWrapper>
            <Profile />
          </LoadingWrapper>
        ),
      },
    ],
  },
  {
    path: '*',
    element: (
      <LoadingWrapper>
        <NotFound />
      </LoadingWrapper>
    ),
  },
]);

export default router;