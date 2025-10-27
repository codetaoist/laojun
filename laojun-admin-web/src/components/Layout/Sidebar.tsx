import { useState, useEffect } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { Layout, Menu, theme } from 'antd';
import {
  DashboardOutlined,
  UserOutlined,
  TeamOutlined,
  SafetyOutlined,
  MenuOutlined,
  AppstoreOutlined,
  SettingOutlined,
  AuditOutlined,
} from '@ant-design/icons';
import { useAppStore } from '@/stores/app';

const { Sider } = Layout;

interface MenuItem {
  key: string;
  icon: React.ReactNode;
  label: string;
  path: string;
  children?: MenuItem[];
}

const menuItems: MenuItem[] = [
  {
    key: 'dashboard',
    icon: <DashboardOutlined />,
    label: '仪表盘',
    path: '/dashboard',
  },
  {
    key: 'system',
    icon: <SettingOutlined />,
    label: '系统管理',
    path: '/system',
    children: [
      {
        key: 'users',
        icon: <UserOutlined />,
        label: '用户管理',
        path: '/users',
      },
      {
        key: 'roles',
        icon: <TeamOutlined />,
        label: '角色管理',
        path: '/roles',
      },
      {
        key: 'permissions',
        icon: <SafetyOutlined />,
        label: '权限管理',
        path: '/permissions',
      },
      {
        key: 'menus',
        icon: <MenuOutlined />,
        label: '菜单管理',
        path: '/menus',
      },
    ],
  },
  {
    key: 'plugins',
    icon: <AppstoreOutlined />,
    label: '插件管理',
    path: '/plugins',
  },
  {
    key: 'plugin-review',
    icon: <AuditOutlined />,
    label: '插件审核',
    path: '/plugin-review',
  },
  {
    key: 'settings',
    icon: <SettingOutlined />,
    label: '系统设置',
    path: '/settings',
  },
];

const Sidebar: React.FC = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const { sidebarCollapsed } = useAppStore();
  const { token } = theme.useToken();
  
  const [selectedKeys, setSelectedKeys] = useState<string[]>([]);
  const [openKeys, setOpenKeys] = useState<string[]>([]);

  // 根据当前路径设置选中的菜单项
  useEffect(() => {
    const currentPath = location.pathname;
    
    // 查找匹配的菜单项
    const findMenuItem = (items: MenuItem[], path: string): { key: string; parentKey?: string } | null => {
      for (const item of items) {
        if (item.path === path) {
          return { key: item.key };
        }
        if (item.children) {
          for (const child of item.children) {
            if (child.path === path) {
              return { key: child.key, parentKey: item.key };
            }
          }
        }
      }
      return null;
    };

    const matched = findMenuItem(menuItems, currentPath);
    if (matched) {
      setSelectedKeys([matched.key]);
      if (matched.parentKey && !sidebarCollapsed) {
        setOpenKeys([matched.parentKey]);
      }
    }
  }, [location.pathname, sidebarCollapsed]);

  // 处理菜单点击
  const handleMenuClick = ({ key }: { key: string }) => {
    const findPath = (items: MenuItem[], targetKey: string): string | null => {
      for (const item of items) {
        if (item.key === targetKey) {
          return item.path;
        }
        if (item.children) {
          for (const child of item.children) {
            if (child.key === targetKey) {
              return child.path;
            }
          }
        }
      }
      return null;
    };

    const path = findPath(menuItems, key);
    if (path) {
      navigate(path);
    }
  };

  // 处理子菜单展开/收起
  const handleOpenChange = (keys: string[]) => {
    setOpenKeys(keys);
  };

  // 转换菜单数据格式
  const convertMenuItems = (items: MenuItem[]) => {
    return items.map((item) => {
      if (item.children) {
        return {
          key: item.key,
          icon: item.icon,
          label: item.label,
          children: item.children.map((child) => ({
            key: child.key,
            icon: child.icon,
            label: child.label,
          })),
        };
      }
      return {
        key: item.key,
        icon: item.icon,
        label: item.label,
      };
    });
  };

  return (
    <Sider
      trigger={null}
      collapsible
      collapsed={sidebarCollapsed}
      style={{
        position: 'fixed',
        left: 0,
        top: 0,
        bottom: 0,
        zIndex: 1000,
        background: token.colorBgContainer,
        borderRight: `1px solid ${token.colorBorderSecondary}`,
      }}
      width={200}
      collapsedWidth={80}
    >
      {/* Logo 区域 */}
      <div
        style={{
          height: '64px',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          borderBottom: `1px solid ${token.colorBorderSecondary}`,
          background: token.colorBgContainer,
        }}
      >
        {sidebarCollapsed ? (
          <div style={{ fontSize: '24px', fontWeight: 'bold', color: token.colorPrimary }}>
            老
          </div>
        ) : (
          <div style={{ fontSize: '18px', fontWeight: 'bold', color: token.colorPrimary }}>
            太上老君
          </div>
        )}
      </div>

      {/* 菜单 */}
      <Menu
        mode="inline"
        selectedKeys={selectedKeys}
        openKeys={sidebarCollapsed ? [] : openKeys}
        onOpenChange={handleOpenChange}
        onClick={handleMenuClick}
        style={{
          height: 'calc(100vh - 64px)',
          borderRight: 0,
          overflow: 'auto',
        }}
        items={convertMenuItems(menuItems)}
      />
    </Sider>
  );
};

export default Sidebar;