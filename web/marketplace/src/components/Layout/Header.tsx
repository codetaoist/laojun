import React, { useState } from 'react';
import { Link, useNavigate, useLocation } from 'react-router-dom';
import {
  Layout,
  Input,
  Button,
  Badge,
  Dropdown,
  Avatar,
  Space,
  Menu,
  AutoComplete,
  Drawer,
} from 'antd';
import {
  SearchOutlined,
  ShoppingCartOutlined,
  UserOutlined,
  HeartOutlined,
  SettingOutlined,
  LogoutOutlined,
  MenuOutlined,
  AppstoreOutlined,
  TeamOutlined,
  HomeOutlined,
  DownloadOutlined,
  MessageOutlined,
  EditOutlined,
  CodeOutlined,
  UploadOutlined,
} from '@ant-design/icons';
import { useCartStore } from '@/stores/cart';
import { useAppStore } from '@/stores/app';
import { useDownloadStore } from '@/stores/download';
import { useAuthStore } from '@/stores/auth';
import DownloadManager from '@/components/DownloadManager';
import './Header.css';

const { Header: AntHeader } = Layout;
const { Search } = Input;

const Header: React.FC = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const [searchValue, setSearchValue] = useState('');
  const [searchOptions, setSearchOptions] = useState<{ value: string }[]>([]);
  const [mobileMenuVisible, setMobileMenuVisible] = useState(false);
  const [downloadManagerVisible, setDownloadManagerVisible] = useState(false);
  
  const { getItemCount } = useCartStore();
  const { searchHistory, addSearchHistory } = useAppStore();
  const { getActiveTasksCount } = useDownloadStore();
  const { user, isAuthenticated, logout } = useAuthStore();
  
  const cartItemCount = getItemCount();
  const activeDownloadsCount = getActiveTasksCount();

  // 搜索处理
  const handleSearch = (value: string) => {
    const trimmedValue = value.trim();
    if (trimmedValue) {
      addSearchHistory(trimmedValue);
      navigate(`/search?q=${encodeURIComponent(trimmedValue)}`);
    }
  };

  // 需要登录的功能处理
  const handleAuthRequiredAction = (action: () => void, redirectPath?: string) => {
    if (isAuthenticated) {
      action();
    } else {
      // 未登录时跳转到登录页面，并记录原本要访问的页面
      const returnUrl = redirectPath || location.pathname;
      navigate(`/login?returnUrl=${encodeURIComponent(returnUrl)}`);
    }
  };

  // 搜索建议
  const handleSearchChange = (value: string) => {
    setSearchValue(value);
    
    if (value) {
      // 从搜索历史中筛选建议
      const suggestions = searchHistory
        .filter(item => item.toLowerCase().includes(value.toLowerCase()))
        .slice(0, 5)
        .map(item => ({ value: item }));
      
      setSearchOptions(suggestions);
    } else {
      setSearchOptions([]);
    }
  };

  // 用户菜单
  const userMenuItems = [
    {
      key: 'upload-plugin',
      icon: <UploadOutlined />,
      label: '上传插件',
      onClick: () => navigate('/upload-plugin'),
    },
    {
      type: 'divider' as const,
    },
    {
      key: 'my-plugins',
      icon: <AppstoreOutlined />,
      label: '我的插件',
      onClick: () => navigate('/my-plugins'),
    },
    {
      key: 'favorites',
      icon: <HeartOutlined />,
      label: '我的收藏',
      onClick: () => navigate('/favorites'),
    },
    {
      type: 'divider' as const,
    },
    {
      key: 'profile',
      icon: <UserOutlined />,
      label: '个人资料',
      onClick: () => navigate('/profile'),
    },
    {
      key: 'settings',
      icon: <SettingOutlined />,
      label: '设置',
      onClick: () => navigate('/settings'),
    },
    {
      type: 'divider' as const,
    },
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: '退出登录',
      onClick: () => {
        logout();
        navigate('/');
      },
    },
  ];

  // 分类数据（模拟数据，实际应从API获取）
  const categories = [
    { key: 'productivity', label: '效率工具', icon: '⚡' },
    { key: 'development', label: '开发工具', icon: '💻' },
    { key: 'design', label: '设计工具', icon: '🎨' },
    { key: 'communication', label: '沟通协作', icon: '💬' },
    { key: 'entertainment', label: '娱乐休闲', icon: '🎮' },
    { key: 'education', label: '教育学习', icon: '📚' },
    { key: 'business', label: '商务办公', icon: '💼' },
    { key: 'security', label: '安全防护', icon: '🔒' },
  ];

  // 分类菜单项
  const categoryMenuItems = categories.map(category => ({
    key: `/categories/${category.key}`,
    label: (
      <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
        <span style={{ fontSize: '16px' }}>{category.icon}</span>
        <span>{category.label}</span>
      </div>
    ),
    onClick: () => navigate(`/categories/${category.key}`),
  }));

  // 社区菜单项
  const communityMenuItems = [
    {
      key: '/community',
      label: (
        <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
          <MessageOutlined />
          <span>社区首页</span>
        </div>
      ),
      onClick: () => navigate('/community'),
    },
    {
      key: '/community/forum',
      label: (
        <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
          <MessageOutlined />
          <span>开发者论坛</span>
        </div>
      ),
      onClick: () => navigate('/community/forum'),
    },
    {
      key: '/community/blog',
      label: (
        <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
          <EditOutlined />
          <span>技术博客</span>
        </div>
      ),
      onClick: () => navigate('/community/blog'),
    },
    {
      key: '/community/code',
      label: (
        <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
          <CodeOutlined />
          <span>代码分享</span>
        </div>
      ),
      onClick: () => navigate('/community/code'),
    },
  ];

  // 导航菜单项
  const navItems = [
    { key: '/', label: '首页', icon: <HomeOutlined /> },
    { 
      key: '/categories', 
      label: '分类', 
      icon: <AppstoreOutlined />,
      hasDropdown: true,
      dropdownItems: categoryMenuItems
    },
    { 
      key: '/community', 
      label: '社区', 
      icon: <MessageOutlined />,
      hasDropdown: true,
      dropdownItems: communityMenuItems
    },
    { key: '/developers', label: '开发者', icon: <TeamOutlined /> },
  ];

  // 移动端菜单
  const mobileMenuItems = [
    ...navItems,
    { type: 'divider' as const },
    { key: '/community', label: '社区首页', icon: <MessageOutlined /> },
    { key: '/community/forum', label: '开发者论坛', icon: <MessageOutlined /> },
    { key: '/community/blog', label: '技术博客', icon: <EditOutlined /> },
    { key: '/community/code', label: '代码分享', icon: <CodeOutlined /> },
    { type: 'divider' as const },
    // 需要登录的功能
    ...(isAuthenticated ? [
      { key: '/upload-plugin', label: '上传插件', icon: <UploadOutlined /> },
      { key: '/my-plugins', label: '我的插件', icon: <AppstoreOutlined /> },
      { key: '/favorites', label: '我的收藏', icon: <HeartOutlined /> },
      { key: '/cart', label: '购物车', icon: <ShoppingCartOutlined /> },
      { 
        key: 'downloads', 
        label: `下载管理${activeDownloadsCount > 0 ? ` (${activeDownloadsCount})` : ''}`, 
        icon: <DownloadOutlined />,
        onClick: () => setDownloadManagerVisible(true)
      },
      { type: 'divider' as const },
      { key: '/profile', label: '个人资料', icon: <UserOutlined /> },
      { key: '/settings', label: '设置', icon: <SettingOutlined /> },
    ] : [
      { key: '/login', label: '登录', icon: <UserOutlined /> },
      { key: '/login?tab=register', label: '注册', icon: <UserOutlined /> },
    ]),
  ];

  return (
    <AntHeader className="marketplace-header">
      <div className="header-container">
        {/* Logo */}
        <div className="header-logo">
          <Link to="/" className="logo-link">
            <img src="/logo.svg" alt="太上老君插件市场" className="logo-image" />
            <span className="logo-text">插件市场</span>
          </Link>
        </div>

        {/* 导航菜单 - 桌面端 */}
        <nav className="header-nav desktop-only">
          <Space size="large">
            {navItems.map(item => (
              item.hasDropdown ? (
                <Dropdown
                  key={item.key}
                  menu={{ items: item.dropdownItems }}
                  placement="bottomLeft"
                  trigger={['hover']}
                  overlayClassName="nav-dropdown"
                >
                  <Link
                    to={item.key}
                    className={`nav-link ${location.pathname.startsWith(item.key) ? 'active' : ''}`}
                    onClick={(e) => {
                      // 如果有下拉菜单，点击时不跳转，而是显示下拉菜单
                      if (item.hasDropdown) {
                        e.preventDefault();
                      }
                    }}
                  >
                    {item.icon}
                    <span>{item.label}</span>
                  </Link>
                </Dropdown>
              ) : (
                <Link
                  key={item.key}
                  to={item.key}
                  className={`nav-link ${location.pathname === item.key ? 'active' : ''}`}
                >
                  {item.icon}
                  <span>{item.label}</span>
                </Link>
              )
            ))}
          </Space>
        </nav>

        {/* 搜索框 */}
        <div className="header-search">
          <AutoComplete
            options={searchOptions}
            value={searchValue}
            onChange={handleSearchChange}
            onSelect={handleSearch}
            style={{ width: '100%' }}
          >
            <Search
              placeholder="搜索插件..."
              allowClear
              enterButton={<SearchOutlined />}
              size="middle"
              onSearch={handleSearch}
            />
          </AutoComplete>
        </div>

        {/* 右侧操作区 */}
        <div className="header-actions">
          {/* 下载管理器 - 桌面端 */}
          <div className="desktop-only">
            <Badge count={isAuthenticated ? activeDownloadsCount : 0} size="small">
              <Button
                type="text"
                icon={<DownloadOutlined />}
                onClick={() => handleAuthRequiredAction(() => setDownloadManagerVisible(true))}
                className="action-button"
                title={isAuthenticated ? "下载管理" : "登录后查看下载管理"}
              >
                下载
              </Button>
            </Badge>
          </div>

          {/* 购物车 - 桌面端 */}
          <div className="desktop-only">
            <Badge count={isAuthenticated ? cartItemCount : 0} size="small">
              <Button
                type="text"
                icon={<ShoppingCartOutlined />}
                onClick={() => handleAuthRequiredAction(() => navigate('/cart'), '/cart')}
                className="action-button"
                title={isAuthenticated ? "购物车" : "登录后查看购物车"}
              >
                购物车
              </Button>
            </Badge>
          </div>

          {/* 上传插件 - 桌面端 */}
          <div className="desktop-only">
            <Button
              type="primary"
              icon={<UploadOutlined />}
              onClick={() => handleAuthRequiredAction(() => navigate('/upload-plugin'), '/upload-plugin')}
              size="small"
              title={isAuthenticated ? "上传插件" : "登录后上传插件"}
            >
              上传插件
            </Button>
          </div>

          {/* 用户菜单 - 桌面端 */}
          <div className="desktop-only">
            {isAuthenticated ? (
              <Dropdown
                menu={{ items: userMenuItems }}
                placement="bottomRight"
                trigger={['click']}
              >
                <Button type="text" className="user-button">
                  <Avatar 
                    size="small" 
                    src={user?.avatar} 
                    icon={<UserOutlined />} 
                  />
                  <span className="user-name">{user?.username || user?.fullName || '用户'}</span>
                </Button>
              </Dropdown>
            ) : (
              <Space>
                <Button 
                  type="text" 
                  onClick={() => navigate('/login')}
                  className="action-button"
                >
                  登录
                </Button>
                <Button 
                  type="primary" 
                  onClick={() => navigate('/login?tab=register')}
                  size="small"
                >
                  注册
                </Button>
              </Space>
            )}
          </div>

          {/* 移动端菜单按钮 */}
          <div className="mobile-only">
            <Button
              type="text"
              icon={<MenuOutlined />}
              onClick={() => setMobileMenuVisible(true)}
              className="mobile-menu-button"
            />
          </div>
        </div>
      </div>

      {/* 移动端抽屉菜单 */}
      <Drawer
        title="菜单"
        placement="right"
        onClose={() => setMobileMenuVisible(false)}
        open={mobileMenuVisible}
        className="mobile-menu-drawer"
      >
        <Menu
          mode="vertical"
          selectedKeys={[location.pathname]}
          items={mobileMenuItems}
          onClick={({ key, domEvent }) => {
            // 处理下载管理器特殊情况
            const item = mobileMenuItems.find(item => item.key === key);
            if (item && 'onClick' in item && item.onClick) {
              item.onClick();
              setMobileMenuVisible(false);
              return;
            }
            
            navigate(key);
            setMobileMenuVisible(false);
          }}
        />
      </Drawer>

      {/* 全局下载管理器 */}
      <DownloadManager
        visible={downloadManagerVisible}
        onClose={() => setDownloadManagerVisible(false)}
      />
    </AntHeader>
  );
};

export default Header;