import React from 'react';
import { Button, Dropdown, Space, theme } from 'antd';
import {
  SunOutlined,
  MoonOutlined,
  DesktopOutlined,
  BgColorsOutlined,
} from '@ant-design/icons';
import { useAppStore } from '@/stores/app';

const ThemeSwitch: React.FC = () => {
  const { theme: currentTheme, setTheme } = useAppStore();
  const { token } = theme.useToken();

  const themeOptions = [
    {
      key: 'light',
      label: '浅色主题',
      icon: <SunOutlined />,
      description: '明亮清爽的界面风格',
    },
    {
      key: 'dark',
      label: '深色主题',
      icon: <MoonOutlined />,
      description: '护眼的深色界面风格',
    },
    {
      key: 'auto',
      label: '跟随系统',
      icon: <DesktopOutlined />,
      description: '根据系统设置自动切换',
    },
  ];

  const getCurrentThemeIcon = () => {
    switch (currentTheme) {
      case 'light':
        return <SunOutlined />;
      case 'dark':
        return <MoonOutlined />;
      case 'auto':
        return <DesktopOutlined />;
      default:
        return <BgColorsOutlined />;
    }
  };

  const getCurrentThemeLabel = () => {
    const option = themeOptions.find(opt => opt.key === currentTheme);
    return option?.label || '主题';
  };

  const menuItems = themeOptions.map(option => ({
    key: option.key,
    label: (
      <div style={{ padding: '8px 0' }}>
        <Space>
          {option.icon}
          <div>
            <div style={{ fontWeight: 500 }}>{option.label}</div>
            <div style={{ 
              fontSize: '12px', 
              color: token.colorTextSecondary,
              marginTop: '2px'
            }}>
              {option.description}
            </div>
          </div>
          {currentTheme === option.key && (
            <div style={{
              width: '6px',
              height: '6px',
              borderRadius: '50%',
              backgroundColor: token.colorPrimary,
              marginLeft: '8px'
            }} />
          )}
        </Space>
      </div>
    ),
    onClick: () => setTheme(option.key as any),
  }));

  return (
    <Dropdown
      menu={{ items: menuItems }}
      placement="bottomRight"
      trigger={['click']}
      overlayStyle={{ minWidth: '240px' }}
    >
      <Button
        type="text"
        icon={getCurrentThemeIcon()}
        style={{
          display: 'flex',
          alignItems: 'center',
          gap: '6px',
          height: '40px',
          padding: '0 12px',
          borderRadius: token.borderRadius,
        }}
      >
        <span className="mobile-hidden">{getCurrentThemeLabel()}</span>
      </Button>
    </Dropdown>
  );
};

export default ThemeSwitch;