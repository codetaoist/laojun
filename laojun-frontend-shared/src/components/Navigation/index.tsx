import React from 'react';
import { Menu, Breadcrumb, Dropdown, Avatar, Badge, Space } from 'antd';
import {
  UserOutlined,
  BellOutlined,
  SettingOutlined,
  LogoutOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
} from '@ant-design/icons';
import { useNavigate, useLocation } from 'react-router-dom';
import { User, UserRole } from '../../types';
import { BaseRouteConfig } from '../../router';

// 菜单项类型
export interface MenuItem {
  key: string;
  label: string;
  icon?: React.ReactNode;
  path?: string;
  children?: MenuItem[];
  roles?: UserRole[];
  disabled?: boolean;
}

// 侧边栏菜单属性
export interface SideMenuProps {
  items: MenuItem[];
  user?: User | null;
  collapsed?: boolean;
  selectedKeys?: string[];
  openKeys?: string[];
  onSelect?: (key: string, path?: string) => void;
  onOpenChange?: (openKeys: string[]) => void;
  theme?: 'light' | 'dark';
  mode?: 'vertical' | 'horizontal' | 'inline';
  style?: React.CSSProperties;
  className?: string;
}

// 侧边栏菜单组件
export const SideMenu: React.FC<SideMenuProps> = ({
  items,
  user,
  collapsed = false,
  selectedKeys,
  openKeys,
  onSelect,
  onOpenChange,
  theme = 'dark',
  mode = 'inline',
  style,
  className,
}) => {
  const navigate = useNavigate();
  const location = useLocation();

  // 过滤有权限的菜单项
  const filterMenuItems = (items: MenuItem[]): MenuItem[] => {
    return items.filter(item => {
      // 检查角色权限
      if (item.roles && item.roles.length > 0) {
        if (!user || !item.roles.includes(user.role)) {
          return false;
        }
      }

      // 递归过滤子菜单
      if (item.children) {
        item.children = filterMenuItems(item.children);
      }

      return true;
    });
  };

  // 转换为 Ant Design Menu 项格式
  const convertToMenuItems = (items: MenuItem[]): any[] => {
    return items.map(item => ({
      key: item.key,
      label: item.label,
      icon: item.icon,
      disabled: item.disabled,
      children: item.children ? convertToMenuItems(item.children) : undefined,
    }));
  };

  const handleMenuClick = ({ key }: { key: string }) => {
    const findMenuItem = (items: MenuItem[], targetKey: string): MenuItem | null => {
      for (const item of items) {
        if (item.key === targetKey) {
          return item;
        }
        if (item.children) {
          const found = findMenuItem(item.children, targetKey);
          if (found) return found;
        }
      }
      return null;
    };

    const menuItem = findMenuItem(items, key);
    if (menuItem?.path) {
      navigate(menuItem.path);
    }
    onSelect?.(key, menuItem?.path);
  };

  const filteredItems = filterMenuItems(items);
  const menuItems = convertToMenuItems(filteredItems);

  // 自动计算选中的菜单项
  const currentSelectedKeys = selectedKeys || [location.pathname];

  return (
    <Menu
      theme={theme}
      mode={mode}
      selectedKeys={currentSelectedKeys}
      openKeys={openKeys}
      onSelect={handleMenuClick}
      onOpenChange={onOpenChange}
      items={menuItems}
      inlineCollapsed={collapsed}
      style={style}
      className={className}
    />
  );
};

// 顶部导航栏属性
export interface TopNavProps {
  user?: User | null;
  collapsed?: boolean;
  onToggleCollapse?: () => void;
  onLogout?: () => void;
  onProfileClick?: () => void;
  onSettingsClick?: () => void;
  notificationCount?: number;
  onNotificationClick?: () => void;
  logo?: React.ReactNode;
  title?: string;
  extra?: React.ReactNode;
  style?: React.CSSProperties;
  className?: string;
}

// 顶部导航栏组件
export const TopNav: React.FC<TopNavProps> = ({
  user,
  collapsed = false,
  onToggleCollapse,
  onLogout,
  onProfileClick,
  onSettingsClick,
  notificationCount = 0,
  onNotificationClick,
  logo,
  title,
  extra,
  style,
  className,
}) => {
  const userMenuItems = [
    {
      key: 'profile',
      label: '个人资料',
      icon: <UserOutlined />,
      onClick: onProfileClick,
    },
    {
      key: 'settings',
      label: '设置',
      icon: <SettingOutlined />,
      onClick: onSettingsClick,
    },
    {
      type: 'divider' as const,
    },
    {
      key: 'logout',
      label: '退出登录',
      icon: <LogoutOutlined />,
      onClick: onLogout,
    },
  ];

  return (
    <div
      className={className}
      style={{
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
        padding: '0 24px',
        height: '100%',
        background: '#fff',
        borderBottom: '1px solid #f0f0f0',
        ...style,
      }}
    >
      {/* 左侧 */}
      <div style={{ display: 'flex', alignItems: 'center' }}>
        {onToggleCollapse && (
          <div
            style={{ cursor: 'pointer', marginRight: 16 }}
            onClick={onToggleCollapse}
          >
            {collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
          </div>
        )}
        
        {logo && (
          <div style={{ marginRight: 16 }}>
            {logo}
          </div>
        )}
        
        {title && (
          <h1 style={{ margin: 0, fontSize: 18, fontWeight: 500 }}>
            {title}
          </h1>
        )}
      </div>

      {/* 右侧 */}
      <div style={{ display: 'flex', alignItems: 'center' }}>
        {extra}
        
        <Space size="middle">
          {/* 通知 */}
          {onNotificationClick && (
            <Badge count={notificationCount} size="small">
              <BellOutlined
                style={{ fontSize: 16, cursor: 'pointer' }}
                onClick={onNotificationClick}
              />
            </Badge>
          )}

          {/* 用户菜单 */}
          {user && (
            <Dropdown
              menu={{ items: userMenuItems }}
              placement="bottomRight"
              trigger={['click']}
            >
              <div style={{ cursor: 'pointer', display: 'flex', alignItems: 'center' }}>
                <Avatar
                  size="small"
                  src={user.avatar}
                  icon={<UserOutlined />}
                  style={{ marginRight: 8 }}
                />
                <span>{user.username}</span>
              </div>
            </Dropdown>
          )}
        </Space>
      </div>
    </div>
  );
};

// 面包屑导航属性
export interface BreadcrumbNavProps {
  items: Array<{
    title: string;
    path?: string;
  }>;
  separator?: string;
  style?: React.CSSProperties;
  className?: string;
}

// 面包屑导航组件
export const BreadcrumbNav: React.FC<BreadcrumbNavProps> = ({
  items,
  separator = '/',
  style,
  className,
}) => {
  const navigate = useNavigate();

  const breadcrumbItems = items.map((item) => ({
    title: item.path ? (
      <a onClick={() => navigate(item.path!)}>{item.title}</a>
    ) : (
      item.title
    ),
  }));

  return (
    <Breadcrumb
      separator={separator}
      items={breadcrumbItems}
      style={style}
      className={className}
    />
  );
};

// 从路由配置生成菜单项
export const generateMenuFromRoutes = (routes: BaseRouteConfig[]): MenuItem[] => {
  const convertRoute = (route: BaseRouteConfig): MenuItem | null => {
    // 跳过隐藏的菜单项
    if (route.meta?.hideInMenu) {
      return null;
    }

    const menuItem: MenuItem = {
      key: route.path,
      label: route.meta?.title || route.name || route.path,
      icon: route.meta?.icon ? React.createElement(route.meta.icon as any) : undefined,
      path: route.path,
      roles: route.meta?.roles,
    };

    // 处理子路由
    if (route.children) {
      const children = route.children
        .map(convertRoute)
        .filter((item): item is MenuItem => item !== null);
      
      if (children.length > 0) {
        menuItem.children = children;
      }
    }

    return menuItem;
  };

  return routes
    .map(convertRoute)
    .filter((item): item is MenuItem => item !== null);
};