import { Layout, Button, Dropdown, Avatar, Space, theme, message, Tag } from 'antd';
import React, { useEffect, useState } from 'react';
import {
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  UserOutlined,
  SettingOutlined,
  LogoutOutlined,
  BellOutlined,
  EyeOutlined,
} from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { useAppStore } from '@/stores/app';
import { useAuthStore } from '@/stores/auth';
import ThemeSwitch from '@/components/ThemeSwitch';
import LanguageSwitch from '@/components/LanguageSwitch';
import { useAccessibility } from '@/components/AccessibilityProvider';
import { useTranslation } from '@/locales';

const { Header: AntHeader } = Layout;

const Header: React.FC = () => {
  const navigate = useNavigate();
  const { sidebarCollapsed, toggleSidebar } = useAppStore();
  const { user, logout, token: authToken, expiresAt } = useAuthStore();
  const { token } = theme.useToken();
  const { togglePanel } = useAccessibility();
  const { t } = useTranslation();

  const [remaining, setRemaining] = useState<string | null>(null);
  useEffect(() => {
    const update = () => {
      if (!expiresAt) { setRemaining(null); return; }
      const diff = Date.parse(expiresAt) - Date.now();
      if (diff <= 0) { setRemaining(t('auth.expired')); return; }
      const mins = Math.floor(diff / 60000);
      const secs = Math.floor((diff % 60000) / 1000);
      setRemaining(t('auth.remaining', { mins, secs }));
    };
    update();
    const timer = setInterval(update, 1000);
    return () => clearInterval(timer);
  }, [expiresAt, t]);

  // 处理登出
  const handleLogout = async () => {
    try {
      await logout();
      message.success(t('auth.logoutSuccess'));
      navigate('/login');
    } catch (error) {
      console.error('Logout error:', error);
      message.error(t('auth.logoutFailed'));
    }
  };

  // 用户下拉菜单
  const userMenuItems = [
    {
      key: 'profile',
      icon: <UserOutlined />,
      label: t('menu.profile'),
      onClick: () => navigate('/profile'),
    },
    {
      key: 'settings',
      icon: <SettingOutlined />,
      label: t('menu.settings'),
      onClick: () => navigate('/settings'),
    },
    {
      type: 'divider' as const,
    },
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: t('auth.logout'),
      onClick: handleLogout,
    },
  ];

  const expiryColor = (() => {
    if (!expiresAt || remaining === null) return undefined;
    if (remaining === '已过期') return 'red';
    const diff = Date.parse(expiresAt) - Date.now();
    if (diff <= 10 * 60 * 1000) return 'orange';
    return 'geekblue';
  })();

  return (
    <AntHeader
      style={{
        position: 'sticky',
        top: 0,
        zIndex: 999,
        width: '100%',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
        padding: '0 16px',
        background: token.colorBgContainer,
        borderBottom: `1px solid ${token.colorBorderSecondary}`,
        height: '64px',
      }}
    >
      {/* 左侧：折叠按钮 */}
      <div style={{ display: 'flex', alignItems: 'center' }}>
        <Button
          type="text"
          icon={sidebarCollapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
          onClick={toggleSidebar}
          style={{
            fontSize: '16px',
            width: 40,
            height: 40,
          }}
        />
      </div>

      {/* 右侧：用户信息和操作 */}
      <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
        {/* 主题切换 */}
        <ThemeSwitch />
        
        {/* 语言切换 */}
        <LanguageSwitch />
        
        {/* 无障碍访问 */}
        <Button
          type="text"
          icon={<EyeOutlined />}
          style={{
            fontSize: '16px',
            width: 40,
            height: 40,
          }}
          onClick={togglePanel}
          title="无障碍访问设置"
        />
        
        {/* 通知铃铛 */}
        <Button
          type="text"
          icon={<BellOutlined />}
          style={{
            fontSize: '16px',
            width: 40,
            height: 40,
          }}
          onClick={() => message.info(t('notifications.noNew'))}
        />

        {/* 用户信息下拉菜单 */}
        <Dropdown
          menu={{ items: userMenuItems }}
          placement="bottomRight"
          arrow
        >
          <div
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: '8px',
              padding: '4px 8px',
              borderRadius: token.borderRadius,
              cursor: 'pointer',
              transition: 'background-color 0.2s',
            }}
            onMouseEnter={(e) => {
              e.currentTarget.style.backgroundColor = token.colorBgTextHover;
            }}
            onMouseLeave={(e) => {
              e.currentTarget.style.backgroundColor = 'transparent';
            }}
          >
            <Avatar
              size="small"
              icon={<UserOutlined />}
              src={user?.avatar}
              style={{ backgroundColor: token.colorPrimary }}
            />
            <Space direction="vertical" size={0} style={{ lineHeight: 1 }}>
              <div style={{ fontSize: '14px', fontWeight: 500 }}>
                {user?.name || user?.username || t('common.user')}
              </div>
              <div style={{ fontSize: '12px', color: token.colorTextSecondary }}>
                {user?.email || t('common.admin')}
              </div>
            </Space>
            {authToken?.startsWith('dev.') && (
              <Tag color="purple" style={{ marginLeft: 4 }}>{t('common.devMode')}</Tag>
            )}
            {expiresAt && remaining && (
              <Tag color={expiryColor} style={{ marginLeft: 4 }}>{t('auth.remainingLabel')}: {remaining}</Tag>
            )}
          </div>
        </Dropdown>
      </div>
    </AntHeader>
  );
};

export default Header;