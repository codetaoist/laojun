import { useState, useEffect } from 'react';
import { Outlet, useLocation } from 'react-router-dom';
import { Layout as AntLayout, theme, Drawer } from 'antd';
import { useAppStore } from '@/stores/app';
import Sidebar from './Sidebar';
import Header from './Header';
import Breadcrumb from './Breadcrumb';

const { Content } = AntLayout;

const Layout: React.FC = () => {
  const location = useLocation();
  const { sidebarCollapsed, setSidebarCollapsed, theme: appTheme } = useAppStore();
  const { token } = theme.useToken();
  const [isMobile, setIsMobile] = useState(false);

  // 检测屏幕尺寸
  useEffect(() => {
    const checkMobile = () => {
      const mobile = window.innerWidth <= 768;
      setIsMobile(mobile);
      
      // 移动端默认收起侧边栏
      if (mobile && !sidebarCollapsed) {
        setSidebarCollapsed(true);
      }
    };

    checkMobile();
    window.addEventListener('resize', checkMobile);
    return () => window.removeEventListener('resize', checkMobile);
  }, [sidebarCollapsed, setSidebarCollapsed]);

  // 根据路由更新面包屑
  useEffect(() => {
    const { setBreadcrumbs, setPageTitle } = useAppStore.getState();
    
    const pathMap: Record<string, { title: string; breadcrumbs: Array<{ title: string; path?: string }> }> = {
      '/dashboard': {
        title: '仪表盘',
        breadcrumbs: [{ title: '首页' }, { title: '仪表盘' }]
      },
      '/users': {
        title: '用户管理',
        breadcrumbs: [{ title: '首页', path: '/' }, { title: '系统管理' }, { title: '用户管理' }]
      },
      '/roles': {
        title: '角色管理',
        breadcrumbs: [{ title: '首页', path: '/' }, { title: '系统管理' }, { title: '角色管理' }]
      },
      '/permissions': {
        title: '权限管理',
        breadcrumbs: [{ title: '首页', path: '/' }, { title: '系统管理' }, { title: '权限管理' }]
      },
      '/menus': {
        title: '菜单管理',
        breadcrumbs: [{ title: '首页', path: '/' }, { title: '系统管理' }, { title: '菜单管理' }]
      },
      '/plugins': {
        title: '插件管理',
        breadcrumbs: [{ title: '首页', path: '/' }, { title: '插件管理' }]
      },
      '/plugin-review': {
        title: '插件审核',
        breadcrumbs: [{ title: '首页', path: '/' }, { title: '插件审核' }]
      },
      '/settings': {
        title: '系统设置',
        breadcrumbs: [{ title: '首页', path: '/' }, { title: '系统设置' }]
      },
      '/profile': {
        title: '个人资料',
        breadcrumbs: [{ title: '首页', path: '/' }, { title: '个人资料' }]
      },
    };

    const currentPath = pathMap[location.pathname];
    if (currentPath) {
      setPageTitle(currentPath.title);
      setBreadcrumbs(currentPath.breadcrumbs);
    }
  }, [location.pathname]);

  return (
    <AntLayout style={{ minHeight: '100vh' }}>
      {/* 桌面端侧边栏 */}
      {!isMobile && <Sidebar />}
      
      {/* 移动端抽屉式侧边栏 */}
      {isMobile && (
        <Drawer
          title={
            <div style={{ 
              fontSize: '18px', 
              fontWeight: 'bold', 
              color: token.colorPrimary 
            }}>
              太上老君
            </div>
          }
          placement="left"
          onClose={() => setSidebarCollapsed(true)}
          open={!sidebarCollapsed}
          width={250}
          bodyStyle={{ padding: 0 }}
          headerStyle={{ 
            borderBottom: `1px solid ${token.colorBorderSecondary}`,
            padding: '16px 24px'
          }}
        >
          <Sidebar />
        </Drawer>
      )}
      
      <AntLayout 
        style={{ 
          marginLeft: isMobile ? 0 : (sidebarCollapsed ? 80 : 200), 
          transition: 'margin-left 0.2s' 
        }}
      >
        <Header />
        <Content 
          style={{ 
            margin: isMobile ? '8px' : '16px', 
            overflow: 'auto' 
          }}
        >
          <Breadcrumb />
          <div
            className="fade-in"
            style={{
              padding: isMobile ? '16px' : '24px',
              minHeight: 'calc(100vh - 112px)',
              background: token.colorBgContainer,
              borderRadius: token.borderRadiusLG,
              marginTop: '16px',
            }}
          >
            <Outlet />
          </div>
        </Content>
      </AntLayout>
    </AntLayout>
  );
};

export default Layout;