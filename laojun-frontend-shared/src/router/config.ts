import { RouteConfig, UserRole } from '../types';

// 路由配置基类
export interface BaseRouteConfig extends RouteConfig {
  name?: string;
  component?: string; // 组件路径，用于懒加载
  redirect?: string;
  children?: BaseRouteConfig[];
}

// 管理后台路由配置
export const adminRoutes: BaseRouteConfig[] = [
  {
    path: '/login',
    name: 'AdminLogin',
    component: 'pages/Login',
    element: null as any, // 将在运行时设置
    meta: {
      title: '管理员登录',
      requireAuth: false,
      hideInMenu: true,
    },
  },
  {
    path: '/',
    name: 'AdminLayout',
    component: 'layouts/AdminLayout',
    element: null as any,
    meta: {
      requireAuth: true,
      roles: [UserRole.ADMIN],
    },
    children: [
      {
        path: '',
        redirect: '/dashboard',
        element: null as any,
      },
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: 'pages/Dashboard',
        element: null as any,
        meta: {
          title: '仪表盘',
          icon: 'DashboardOutlined',
        },
      },
      {
        path: 'users',
        name: 'UserManagement',
        component: 'pages/UserManagement',
        element: null as any,
        meta: {
          title: '用户管理',
          icon: 'UserOutlined',
        },
      },
      {
        path: 'roles',
        name: 'RoleManagement',
        component: 'pages/RoleManagement',
        element: null as any,
        meta: {
          title: '角色管理',
          icon: 'TeamOutlined',
        },
      },
      {
        path: 'plugins',
        name: 'PluginManagement',
        component: 'pages/PluginManagement',
        element: null as any,
        meta: {
          title: '插件管理',
          icon: 'AppstoreOutlined',
        },
      },
      {
        path: 'plugins/review',
        name: 'PluginReview',
        component: 'pages/PluginReview',
        element: null as any,
        meta: {
          title: '插件审核',
          icon: 'AuditOutlined',
        },
      },
      {
        path: 'categories',
        name: 'CategoryManagement',
        component: 'pages/CategoryManagement',
        element: null as any,
        meta: {
          title: '分类管理',
          icon: 'TagsOutlined',
        },
      },
      {
        path: 'statistics',
        name: 'Statistics',
        component: 'pages/Statistics',
        element: null as any,
        meta: {
          title: '统计分析',
          icon: 'BarChartOutlined',
        },
      },
      {
        path: 'settings',
        name: 'SystemSettings',
        component: 'pages/SystemSettings',
        element: null as any,
        meta: {
          title: '系统设置',
          icon: 'SettingOutlined',
        },
      },
      {
        path: 'profile',
        name: 'AdminProfile',
        component: 'pages/Profile',
        element: null as any,
        meta: {
          title: '个人资料',
          hideInMenu: true,
        },
      },
    ],
  },
  {
    path: '/404',
    name: 'NotFound',
    component: 'pages/NotFound',
    element: null as any,
    meta: {
      title: '页面不存在',
      hideInMenu: true,
    },
  },
  {
    path: '*',
    redirect: '/404',
    element: null as any,
  },
];

// 插件市场路由配置
export const marketplaceRoutes: BaseRouteConfig[] = [
  {
    path: '/login',
    name: 'MarketplaceLogin',
    component: 'pages/Login',
    element: null as any,
    meta: {
      title: '登录',
      requireAuth: false,
      hideInMenu: true,
    },
  },
  {
    path: '/register',
    name: 'Register',
    component: 'pages/Register',
    element: null as any,
    meta: {
      title: '注册',
      requireAuth: false,
      hideInMenu: true,
    },
  },
  {
    path: '/',
    name: 'MarketplaceLayout',
    component: 'layouts/MarketplaceLayout',
    element: null as any,
    children: [
      {
        path: '',
        name: 'Home',
        component: 'pages/Home',
        element: null as any,
        meta: {
          title: '首页',
        },
      },
      {
        path: 'search',
        name: 'Search',
        component: 'pages/Search',
        element: null as any,
        meta: {
          title: '搜索',
        },
      },
      {
        path: 'plugins/:id',
        name: 'PluginDetail',
        component: 'pages/PluginDetail',
        element: null as any,
        meta: {
          title: '插件详情',
          hideInMenu: true,
        },
      },
      {
        path: 'categories',
        name: 'Categories',
        component: 'pages/Categories',
        element: null as any,
        meta: {
          title: '分类',
        },
      },
      {
        path: 'categories/:id',
        name: 'CategoryDetail',
        component: 'pages/CategoryDetail',
        element: null as any,
        meta: {
          title: '分类详情',
          hideInMenu: true,
        },
      },
      {
        path: 'developers',
        name: 'Developers',
        component: 'pages/Developers',
        element: null as any,
        meta: {
          title: '开发者',
        },
      },
      {
        path: 'developers/:id',
        name: 'DeveloperDetail',
        component: 'pages/DeveloperDetail',
        element: null as any,
        meta: {
          title: '开发者详情',
          hideInMenu: true,
        },
      },
      {
        path: 'cart',
        name: 'Cart',
        component: 'pages/Cart',
        element: null as any,
        meta: {
          title: '购物车',
          requireAuth: true,
          hideInMenu: true,
        },
      },
      {
        path: 'checkout',
        name: 'Checkout',
        component: 'pages/Checkout',
        element: null as any,
        meta: {
          title: '结账',
          requireAuth: true,
          hideInMenu: true,
        },
      },
      {
        path: 'my',
        name: 'MyCenter',
        component: 'layouts/UserLayout',
        element: null as any,
        meta: {
          title: '个人中心',
          requireAuth: true,
        },
        children: [
          {
            path: '',
            redirect: '/my/plugins',
            element: null as any,
          },
          {
            path: 'plugins',
            name: 'MyPlugins',
            component: 'pages/MyPlugins',
            element: null as any,
            meta: {
              title: '我的插件',
            },
          },
          {
            path: 'upload',
            name: 'UploadPlugin',
            component: 'pages/UploadPlugin',
            element: null as any,
            meta: {
              title: '上传插件',
              roles: [UserRole.DEVELOPER],
            },
          },
          {
            path: 'favorites',
            name: 'Favorites',
            component: 'pages/Favorites',
            element: null as any,
            meta: {
              title: '收藏夹',
            },
          },
          {
            path: 'orders',
            name: 'Orders',
            component: 'pages/Orders',
            element: null as any,
            meta: {
              title: '订单记录',
            },
          },
          {
            path: 'profile',
            name: 'UserProfile',
            component: 'pages/Profile',
            element: null as any,
            meta: {
              title: '个人资料',
            },
          },
          {
            path: 'settings',
            name: 'UserSettings',
            component: 'pages/Settings',
            element: null as any,
            meta: {
              title: '账户设置',
            },
          },
        ],
      },
      {
        path: 'about',
        name: 'About',
        component: 'pages/About',
        element: null as any,
        meta: {
          title: '关于我们',
        },
      },
      {
        path: 'privacy',
        name: 'Privacy',
        component: 'pages/Privacy',
        element: null as any,
        meta: {
          title: '隐私政策',
        },
      },
      {
        path: 'terms',
        name: 'Terms',
        component: 'pages/Terms',
        element: null as any,
        meta: {
          title: '服务条款',
        },
      },
      {
        path: 'community',
        name: 'Community',
        component: 'pages/Community',
        element: null as any,
        meta: {
          title: '社区',
        },
      },
      {
        path: 'help',
        name: 'Help',
        component: 'pages/Help',
        element: null as any,
        meta: {
          title: '帮助中心',
        },
      },
    ],
  },
  {
    path: '/404',
    name: 'NotFound',
    component: 'pages/NotFound',
    element: null as any,
    meta: {
      title: '页面不存在',
      hideInMenu: true,
    },
  },
  {
    path: '*',
    redirect: '/404',
    element: null as any,
  },
];

// 路由配置工厂
export interface RouteConfigOptions {
  type: 'admin' | 'marketplace';
  componentLoader: (path: string) => React.ComponentType;
}

export const createRouteConfig = (options: RouteConfigOptions): BaseRouteConfig[] => {
  const routes = options.type === 'admin' ? adminRoutes : marketplaceRoutes;
  
  // 递归设置组件
  const setComponents = (routes: BaseRouteConfig[]): BaseRouteConfig[] => {
    return routes.map(route => {
      const newRoute = { ...route };
      
      if (newRoute.component) {
        newRoute.element = options.componentLoader(newRoute.component);
      }
      
      if (newRoute.children) {
        newRoute.children = setComponents(newRoute.children);
      }
      
      return newRoute;
    });
  };
  
  return setComponents(routes);
};

// 路由权限检查
export const checkRoutePermission = (route: BaseRouteConfig, userRole?: UserRole): boolean => {
  if (!route.meta?.requireAuth) {
    return true;
  }
  
  if (!userRole) {
    return false;
  }
  
  if (route.meta.roles && route.meta.roles.length > 0) {
    return route.meta.roles.includes(userRole);
  }
  
  return true;
};

// 生成面包屑
export const generateBreadcrumb = (
  routes: BaseRouteConfig[],
  pathname: string
): Array<{ title: string; path?: string }> => {
  const breadcrumb: Array<{ title: string; path?: string }> = [];
  
  const findRoute = (routes: BaseRouteConfig[], path: string): BaseRouteConfig | null => {
    for (const route of routes) {
      if (route.path === path) {
        return route;
      }
      if (route.children) {
        const found = findRoute(route.children, path);
        if (found) {
          return found;
        }
      }
    }
    return null;
  };
  
  const pathSegments = pathname.split('/').filter(Boolean);
  let currentPath = '';
  
  for (const segment of pathSegments) {
    currentPath += `/${segment}`;
    const route = findRoute(routes, currentPath);
    
    if (route?.meta?.title) {
      breadcrumb.push({
        title: route.meta.title,
        path: currentPath,
      });
    }
  }
  
  return breadcrumb;
};