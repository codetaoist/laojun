import { createBrowserRouter, Navigate } from 'react-router-dom';
import { lazy, Suspense } from 'react';
import { Spin } from 'antd';
import Layout from '@/components/Layout';
import ErrorBoundary from '@/components/ErrorBoundary';

// 懒加载页面组件
const Home = lazy(() => import('@/pages/Home'));
const Search = lazy(() => import('@/pages/Search'));
const PluginDetail = lazy(() => import('@/pages/PluginDetail'));
const Category = lazy(() => import('@/pages/Category'));
const Categories = lazy(() => import('@/pages/Categories'));
const Developer = lazy(() => import('@/pages/Developer'));
const Developers = lazy(() => import('@/pages/Developers'));
const Cart = lazy(() => import('@/pages/Cart'));
const Checkout = lazy(() => import('@/pages/Checkout'));
const MyPlugins = lazy(() => import('@/pages/MyPlugins'));
const UploadPlugin = lazy(() => import('@/pages/UploadPlugin'));
const Favorites = lazy(() => import('@/pages/Favorites'));
const Profile = lazy(() => import('@/pages/Profile'));
const Settings = lazy(() => import('@/pages/Settings'));
const Login = lazy(() => import('@/pages/Login'));
const About = lazy(() => import('@/pages/About'));
const Privacy = lazy(() => import('@/pages/Privacy'));
const Terms = lazy(() => import('@/pages/Terms'));
const DownloadTest = lazy(() => import('@/pages/DownloadTest'));
const NotFound = lazy(() => import('@/pages/NotFound'));

// 社区相关页面
const Community = lazy(() => import('@/pages/Community'));
const Forum = lazy(() => import('@/pages/Community/Forum'));
const Blog = lazy(() => import('@/pages/Community/Blog'));
const CodeShare = lazy(() => import('@/pages/Community/CodeShare'));
const PostDetail = lazy(() => import('@/pages/Community/PostDetail'));

// 加载组件包装器
const LoadingWrapper = ({ children }: { children: React.ReactNode }) => (
  <Suspense
    fallback={
      <div style={{ 
        display: 'flex', 
        justifyContent: 'center', 
        alignItems: 'center', 
        height: '50vh' 
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

// 路由配置
export const router = createBrowserRouter([
  {
    path: '/',
    element: <Layout />,
    errorElement: <LoadingWrapper><NotFound /></LoadingWrapper>,
    children: [
      {
        index: true,
        element: <LoadingWrapper><Home /></LoadingWrapper>,
      },
      {
        path: 'search',
        element: <LoadingWrapper><Search /></LoadingWrapper>,
      },
      {
        path: 'plugin/:id',
        element: <LoadingWrapper><PluginDetail /></LoadingWrapper>,
      },
      {
        path: 'categories',
        element: <LoadingWrapper><Categories /></LoadingWrapper>,
      },
      {
        path: 'category/:id',
        element: <LoadingWrapper><Category /></LoadingWrapper>,
      },
      {
        path: 'developers',
        element: <LoadingWrapper><Developers /></LoadingWrapper>,
      },
      {
        path: 'developer/:id',
        element: <LoadingWrapper><Developer /></LoadingWrapper>,
      },
      {
        path: 'cart',
        element: <LoadingWrapper><Cart /></LoadingWrapper>,
      },
      {
        path: 'checkout',
        element: <LoadingWrapper><Checkout /></LoadingWrapper>,
      },
      {
        path: 'my-plugins',
        element: <LoadingWrapper><MyPlugins /></LoadingWrapper>,
      },
      {
        path: 'upload-plugin',
        element: <LoadingWrapper><UploadPlugin /></LoadingWrapper>,
      },
      {
        path: 'upload',
        element: <Navigate to="/upload-plugin" replace />,
      },
      {
        path: 'favorites',
        element: <LoadingWrapper><Favorites /></LoadingWrapper>,
      },
      {
        path: 'profile',
        element: <LoadingWrapper><Profile /></LoadingWrapper>,
      },
      {
        path: 'settings',
        element: <LoadingWrapper><Settings /></LoadingWrapper>,
      },
      {
        path: 'login',
        element: <LoadingWrapper><Login /></LoadingWrapper>,
      },
      {
        path: 'about',
        element: <LoadingWrapper><About /></LoadingWrapper>,
      },
      {
        path: 'privacy',
        element: <LoadingWrapper><Privacy /></LoadingWrapper>,
      },
      {
        path: 'terms',
        element: <LoadingWrapper><Terms /></LoadingWrapper>,
      },
      // 社区路由
      {
        path: 'community',
        element: <LoadingWrapper><Community /></LoadingWrapper>,
      },
      {
        path: 'community/forum',
        element: <LoadingWrapper><Forum /></LoadingWrapper>,
      },
      {
        path: 'community/forum/post/:id',
        element: <LoadingWrapper><PostDetail /></LoadingWrapper>,
      },
      {
        path: 'community/blog',
        element: <LoadingWrapper><Blog /></LoadingWrapper>,
      },
      {
        path: 'community/code',
        element: <LoadingWrapper><CodeShare /></LoadingWrapper>,
      },
      {
        path: 'download-test',
        element: <LoadingWrapper><DownloadTest /></LoadingWrapper>,
      },
      {
        path: '404',
        element: <LoadingWrapper><NotFound /></LoadingWrapper>,
      },
      {
        path: '*',
        element: <Navigate to="/404" replace />,
      },
    ],
  },
]);

export default router;