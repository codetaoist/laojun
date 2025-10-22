import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Row,
  Col,
  Card,
  Typography,
  Space,
  Button,
  Empty,
  message,
  Tag,
  Rate,
  Avatar,
  Spin,
  Select,
  Input,
  Tabs,
  Progress,
  Modal,
  List,
  Tooltip,
  Badge,
  Dropdown,
  Menu,
} from 'antd';
import {
  AppstoreOutlined,
  DownloadOutlined,
  EyeOutlined,
  SettingOutlined,
  DeleteOutlined,
  SearchOutlined,
  SyncOutlined,
  PlayCircleOutlined,
  PauseCircleOutlined,
  ExclamationCircleOutlined,
  CheckCircleOutlined,
  ClockCircleOutlined,
  MoreOutlined,
  CloudDownloadOutlined,
  BugOutlined,
  StarOutlined,
  ShareAltOutlined,
  UploadOutlined,
} from '@ant-design/icons';
import { Plugin } from '@/types';
import { useAuthStore } from '@/stores/auth';

const { Title, Text, Paragraph } = Typography;
const { Option } = Select;
const { Search } = Input;
const { TabPane } = Tabs;

interface InstalledPlugin extends Plugin {
  installDate: string;
  lastUsed: string;
  status: 'active' | 'inactive' | 'error' | 'updating';
  autoUpdate: boolean;
  usageCount: number;
}

const MyPlugins: React.FC = () => {
  const navigate = useNavigate();
  const { user, isAuthenticated } = useAuthStore();
  
  const [loading, setLoading] = useState(true);
  const [installedPlugins, setInstalledPlugins] = useState<InstalledPlugin[]>([]);
  const [filteredPlugins, setFilteredPlugins] = useState<InstalledPlugin[]>([]);
  const [searchText, setSearchText] = useState('');
  const [sortBy, setSortBy] = useState<string>('installDate');
  const [filterStatus, setFilterStatus] = useState<string>('all');
  const [activeTab, setActiveTab] = useState<string>('installed');
  const [updateProgress, setUpdateProgress] = useState<Record<string, number>>({});

  // 模拟已安装的插件数据
  const mockInstalledPlugins: InstalledPlugin[] = [
    {
      id: '1',
      name: 'Code Editor Pro',
      description: '强大的代码编辑器，支持多种编程语言的语法高亮和智能提示',
      version: '2.1.0',
      author: 'DevTools Inc.',
      authorEmail: 'contact@devtools.com',
      category: 'development',
      tags: ['editor', 'syntax', 'autocomplete'],
      icon: 'https://via.placeholder.com/64',
      screenshots: [],
      downloadUrl: '',
      license: 'MIT',
      price: 29.99,
      currency: 'USD',
      downloads: 15420,
      rating: 4.8,
      reviewCount: 234,
      size: 2048000,
      requirements: {
        minVersion: '1.0.0',
        dependencies: [],
      },
      status: 'active',
      featured: true,
      createdAt: '2024-01-01T00:00:00Z',
      updatedAt: '2024-01-15T10:30:00Z',
      publishedAt: '2024-01-01T00:00:00Z',
      installDate: '2024-01-10T08:30:00Z',
      lastUsed: '2024-01-20T14:20:00Z',
      autoUpdate: true,
      usageCount: 156,
    },
    {
      id: '2',
      name: 'Task Manager',
      description: '高效的任务管理工具，帮助您更好地组织和跟踪工作进度',
      version: '1.8.3',
      author: 'Productivity Team',
      authorEmail: 'team@productivity.com',
      category: 'productivity',
      tags: ['task', 'management', 'productivity'],
      icon: 'https://via.placeholder.com/64',
      screenshots: [],
      downloadUrl: '',
      license: 'MIT',
      price: 0,
      currency: 'USD',
      downloads: 12456,
      rating: 4.7,
      reviewCount: 89,
      size: 1536000,
      requirements: {
        minVersion: '1.0.0',
        dependencies: [],
      },
      status: 'active',
      featured: true,
      createdAt: '2024-01-03T00:00:00Z',
      updatedAt: '2024-01-12T09:45:00Z',
      publishedAt: '2024-01-03T00:00:00Z',
      installDate: '2024-01-12T16:15:00Z',
      lastUsed: '2024-01-19T11:30:00Z',
      autoUpdate: false,
      usageCount: 89,
    },
    {
      id: '3',
      name: 'File Organizer',
      description: '智能文件整理工具，自动分类和管理您的文件',
      version: '2.3.1',
      author: 'FileTools Co.',
      authorEmail: 'support@filetools.com',
      category: 'productivity',
      tags: ['file', 'organize', 'automation'],
      icon: 'https://via.placeholder.com/64',
      screenshots: [],
      downloadUrl: '',
      license: 'Commercial',
      price: 19.99,
      currency: 'USD',
      downloads: 7834,
      rating: 4.5,
      reviewCount: 67,
      size: 1024000,
      requirements: {
        minVersion: '1.0.0',
        dependencies: [],
      },
      status: 'inactive',
      featured: false,
      createdAt: '2024-01-05T00:00:00Z',
      updatedAt: '2024-01-18T12:00:00Z',
      publishedAt: '2024-01-05T00:00:00Z',
      installDate: '2024-01-15T10:45:00Z',
      lastUsed: '2024-01-17T09:20:00Z',
      autoUpdate: true,
      usageCount: 23,
    },
  ];

  const sortOptions = [
    { value: 'installDate', label: '安装时间' },
    { value: 'lastUsed', label: '最后使用' },
    { value: 'name', label: '名称' },
    { value: 'usageCount', label: '使用次数' },
    { value: 'rating', label: '评分' },
  ];

  const statusOptions = [
    { value: 'all', label: '全部状态' },
    { value: 'active', label: '运行中' },
    { value: 'inactive', label: '已停用' },
    { value: 'error', label: '错误' },
    { value: 'updating', label: '更新中' },
  ];

  // 检查用户登录状态
  useEffect(() => {
    if (!isAuthenticated) {
      navigate('/login', { state: { from: { pathname: '/my-plugins' } } });
      return;
    }
  }, [isAuthenticated, navigate]);

  useEffect(() => {
    if (isAuthenticated) {
      loadInstalledPlugins();
    }
  }, [isAuthenticated]);

  useEffect(() => {
    filterAndSortPlugins();
  }, [installedPlugins, searchText, sortBy, filterStatus]);

  const loadInstalledPlugins = async () => {
    try {
      setLoading(true);
      // 模拟加载已安装插件数据
      await new Promise(resolve => setTimeout(resolve, 1000));
      setInstalledPlugins(mockInstalledPlugins);
    } catch (error) {
      console.error('加载插件失败:', error);
      message.error('加载插件失败，请刷新页面重试');
    } finally {
      setLoading(false);
    }
  };

  const filterAndSortPlugins = () => {
    let filtered = [...installedPlugins];

    // 搜索过滤
    if (searchText) {
      filtered = filtered.filter(plugin =>
        plugin.name.toLowerCase().includes(searchText.toLowerCase()) ||
        plugin.description.toLowerCase().includes(searchText.toLowerCase()) ||
        plugin.author.toLowerCase().includes(searchText.toLowerCase())
      );
    }

    // 状态过滤
    if (filterStatus !== 'all') {
      filtered = filtered.filter(plugin => plugin.status === filterStatus);
    }

    // 排序
    filtered.sort((a, b) => {
      switch (sortBy) {
        case 'name':
          return a.name.localeCompare(b.name);
        case 'lastUsed':
          return new Date(b.lastUsed).getTime() - new Date(a.lastUsed).getTime();
        case 'usageCount':
          return b.usageCount - a.usageCount;
        case 'rating':
          return b.rating - a.rating;
        case 'installDate':
        default:
          return new Date(b.installDate).getTime() - new Date(a.installDate).getTime();
      }
    });

    setFilteredPlugins(filtered);
  };

  const handlePluginAction = async (pluginId: string, action: string) => {
    const plugin = installedPlugins.find(p => p.id === pluginId);
    if (!plugin) return;

    try {
      switch (action) {
        case 'start':
          setInstalledPlugins(prev => prev.map(p => 
            p.id === pluginId ? { ...p, status: 'active' } : p
          ));
          message.success(`${plugin.name} 已启动`);
          break;
        
        case 'stop':
          setInstalledPlugins(prev => prev.map(p => 
            p.id === pluginId ? { ...p, status: 'inactive' } : p
          ));
          message.success(`${plugin.name} 已停止`);
          break;
        
        case 'update':
          setInstalledPlugins(prev => prev.map(p => 
            p.id === pluginId ? { ...p, status: 'updating' } : p
          ));
          setUpdateProgress({ [pluginId]: 0 });
          
          // 模拟更新进度
          for (let i = 0; i <= 100; i += 10) {
            await new Promise(resolve => setTimeout(resolve, 200));
            setUpdateProgress({ [pluginId]: i });
          }
          
          setInstalledPlugins(prev => prev.map(p => 
            p.id === pluginId ? { ...p, status: 'active', version: '2.2.0' } : p
          ));
          setUpdateProgress({});
          message.success(`${plugin.name} 更新完成`);
          break;
        
        case 'uninstall':
          Modal.confirm({
            title: '确认卸载',
            content: `确定要卸载 ${plugin.name} 吗？此操作不可撤销。`,
            icon: <ExclamationCircleOutlined />,
            onOk: () => {
              setInstalledPlugins(prev => prev.filter(p => p.id !== pluginId));
              message.success(`${plugin.name} 已卸载`);
            },
          });
          break;
        
        case 'settings':
          message.info(`打开 ${plugin.name} 设置`);
          break;
        
        case 'details':
          navigate(`/plugin/${pluginId}`);
          break;
      }
    } catch (error) {
      console.error(`执行操作失败:`, error);
      message.error('操作失败，请重试');
    }
  };

  const handleToggleAutoUpdate = (pluginId: string) => {
    setInstalledPlugins(prev => prev.map(p => 
      p.id === pluginId ? { ...p, autoUpdate: !p.autoUpdate } : p
    ));
    const plugin = installedPlugins.find(p => p.id === pluginId);
    message.success(`${plugin?.name} 自动更新已${plugin?.autoUpdate ? '关闭' : '开启'}`);
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'active':
        return <CheckCircleOutlined style={{ color: '#52c41a' }} />;
      case 'inactive':
        return <PauseCircleOutlined style={{ color: '#faad14' }} />;
      case 'error':
        return <ExclamationCircleOutlined style={{ color: '#ff4d4f' }} />;
      case 'updating':
        return <SyncOutlined spin style={{ color: '#1890ff' }} />;
      default:
        return <ClockCircleOutlined style={{ color: '#d9d9d9' }} />;
    }
  };

  const getStatusText = (status: string) => {
    switch (status) {
      case 'active':
        return '运行中';
      case 'inactive':
        return '已停用';
      case 'error':
        return '错误';
      case 'updating':
        return '更新中';
      default:
        return '未知';
    }
  };

  const renderPluginCard = (plugin: InstalledPlugin) => {
    const actionMenu = (
      <Menu>
        <Menu.Item
          key="details"
          icon={<EyeOutlined />}
          onClick={() => handlePluginAction(plugin.id, 'details')}
        >
          查看详情
        </Menu.Item>
        <Menu.Item
          key="settings"
          icon={<SettingOutlined />}
          onClick={() => handlePluginAction(plugin.id, 'settings')}
        >
          插件设置
        </Menu.Item>
        <Menu.Item
          key="update"
          icon={<CloudDownloadOutlined />}
          onClick={() => handlePluginAction(plugin.id, 'update')}
          disabled={plugin.status === 'updating'}
        >
          检查更新
        </Menu.Item>
        <Menu.Divider />
        <Menu.Item
          key="report"
          icon={<BugOutlined />}
        >
          报告问题
        </Menu.Item>
        <Menu.Item
          key="rate"
          icon={<StarOutlined />}
        >
          评价插件
        </Menu.Item>
        <Menu.Item
          key="share"
          icon={<ShareAltOutlined />}
        >
          分享插件
        </Menu.Item>
        <Menu.Divider />
        <Menu.Item
          key="uninstall"
          icon={<DeleteOutlined />}
          danger
          onClick={() => handlePluginAction(plugin.id, 'uninstall')}
        >
          卸载插件
        </Menu.Item>
      </Menu>
    );

    return (
      <Card
        key={plugin.id}
        style={{ marginBottom: '16px' }}
        actions={[
          plugin.status === 'active' ? (
            <Button
              type="text"
              icon={<PauseCircleOutlined />}
              onClick={() => handlePluginAction(plugin.id, 'stop')}
            >
              停止
            </Button>
          ) : (
            <Button
              type="text"
              icon={<PlayCircleOutlined />}
              onClick={() => handlePluginAction(plugin.id, 'start')}
              disabled={plugin.status === 'updating'}
            >
              启动
            </Button>
          ),
          <Button
            type="text"
            icon={<SettingOutlined />}
            onClick={() => handlePluginAction(plugin.id, 'settings')}
          >
            设置
          </Button>,
          <Dropdown overlay={actionMenu} trigger={['click']}>
            <Button type="text" icon={<MoreOutlined />}>
              更多
            </Button>
          </Dropdown>,
        ]}
      >
        <Card.Meta
          avatar={
            <Badge
              status={plugin.status === 'active' ? 'success' : 'default'}
              offset={[-8, 8]}
            >
              <Avatar
                src={plugin.icon}
                icon={<AppstoreOutlined />}
                size={64}
              />
            </Badge>
          }
          title={
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <Space>
                <Text strong>{plugin.name}</Text>
                <Tag color="blue">v{plugin.version}</Tag>
                {getStatusIcon(plugin.status)}
                <Text type="secondary">{getStatusText(plugin.status)}</Text>
              </Space>
              {plugin.autoUpdate && (
                <Tooltip title="自动更新已开启">
                  <SyncOutlined style={{ color: '#1890ff' }} />
                </Tooltip>
              )}
            </div>
          }
          description={
            <Space direction="vertical" size="small" style={{ width: '100%' }}>
              <Paragraph ellipsis={{ rows: 2 }} style={{ marginBottom: '8px' }}>
                {plugin.description}
              </Paragraph>
              
              {plugin.status === 'updating' && updateProgress[plugin.id] !== undefined && (
                <Progress
                  percent={updateProgress[plugin.id]}
                  size="small"
                  status="active"
                />
              )}
              
              <Row gutter={16}>
                <Col span={12}>
                  <Space direction="vertical" size="small">
                    <Text type="secondary">安装时间</Text>
                    <Text style={{ fontSize: '12px' }}>
                      {new Date(plugin.installDate).toLocaleDateString()}
                    </Text>
                  </Space>
                </Col>
                <Col span={12}>
                  <Space direction="vertical" size="small">
                    <Text type="secondary">最后使用</Text>
                    <Text style={{ fontSize: '12px' }}>
                      {new Date(plugin.lastUsed).toLocaleDateString()}
                    </Text>
                  </Space>
                </Col>
              </Row>
              
              <Row gutter={16}>
                <Col span={12}>
                  <Space direction="vertical" size="small">
                    <Text type="secondary">使用次数</Text>
                    <Text style={{ fontSize: '12px' }}>
                      {plugin.usageCount} 次
                    </Text>
                  </Space>
                </Col>
                <Col span={12}>
                  <Space direction="vertical" size="small">
                    <Text type="secondary">评分</Text>
                    <Space size="small">
                      <Rate disabled defaultValue={plugin.rating} style={{ fontSize: '12px' }} />
                      <Text style={{ fontSize: '12px' }}>({plugin.reviewCount})</Text>
                    </Space>
                  </Space>
                </Col>
              </Row>
              
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <Space>
                  <Tag>{plugin.category}</Tag>
                  <Text type="secondary" style={{ fontSize: '12px' }}>
                    by {plugin.author}
                  </Text>
                </Space>
                <Button
                  type="link"
                  size="small"
                  onClick={() => handleToggleAutoUpdate(plugin.id)}
                >
                  自动更新: {plugin.autoUpdate ? '开启' : '关闭'}
                </Button>
              </div>
            </Space>
          }
        />
      </Card>
    );
  };

  const renderStatistics = () => {
    const activeCount = installedPlugins.filter(p => p.status === 'active').length;
    const inactiveCount = installedPlugins.filter(p => p.status === 'inactive').length;
    const errorCount = installedPlugins.filter(p => p.status === 'error').length;
    const totalUsage = installedPlugins.reduce((sum, p) => sum + p.usageCount, 0);

    return (
      <Row gutter={16} style={{ marginBottom: '24px' }}>
        <Col span={6}>
          <Card size="small">
            <div style={{ textAlign: 'center' }}>
              <Text type="secondary">总插件数</Text>
              <div style={{ fontSize: '24px', fontWeight: 'bold', color: '#1890ff' }}>
                {installedPlugins.length}
              </div>
            </div>
          </Card>
        </Col>
        <Col span={6}>
          <Card size="small">
            <div style={{ textAlign: 'center' }}>
              <Text type="secondary">运行中</Text>
              <div style={{ fontSize: '24px', fontWeight: 'bold', color: '#52c41a' }}>
                {activeCount}
              </div>
            </div>
          </Card>
        </Col>
        <Col span={6}>
          <Card size="small">
            <div style={{ textAlign: 'center' }}>
              <Text type="secondary">已停用</Text>
              <div style={{ fontSize: '24px', fontWeight: 'bold', color: '#faad14' }}>
                {inactiveCount}
              </div>
            </div>
          </Card>
        </Col>
        <Col span={6}>
          <Card size="small">
            <div style={{ textAlign: 'center' }}>
              <Text type="secondary">总使用次数</Text>
              <div style={{ fontSize: '24px', fontWeight: 'bold', color: '#722ed1' }}>
                {totalUsage}
              </div>
            </div>
          </Card>
        </Col>
      </Row>
    );
  };

  if (loading) {
    return (
      <div style={{ 
        display: 'flex', 
        justifyContent: 'center', 
        alignItems: 'center', 
        height: '50vh' 
      }}>
        <Spin size="large" />
      </div>
    );
  }

  // 如果未登录，不渲染内容（会被重定向到登录页面）
  if (!isAuthenticated) {
    return null;
  }

  return (
    <div style={{ padding: '24px', maxWidth: '1200px', margin: '0 auto' }}>
      <div style={{ marginBottom: '24px', display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
        <div>
          <Title level={2}>
            <AppstoreOutlined /> 我的插件
          </Title>
          <Text type="secondary">
            管理您已安装的插件
          </Text>
        </div>
        <Button 
          type="primary" 
          icon={<UploadOutlined />}
          size="large"
          onClick={() => navigate('/upload-plugin')}
          style={{ 
            background: 'linear-gradient(135deg, #1890ff 0%, #096dd9 100%)',
            border: 'none',
            boxShadow: '0 2px 8px rgba(24, 144, 255, 0.3)',
          }}
        >
          上传插件
        </Button>
      </div>

      {renderStatistics()}

      <Tabs activeKey={activeTab} onChange={setActiveTab}>
        <TabPane tab="已安装插件" key="installed">
          {installedPlugins.length === 0 ? (
            <Empty
              description="您还没有安装任何插件"
              image={Empty.PRESENTED_IMAGE_SIMPLE}
            >
              <Button type="primary" onClick={() => navigate('/')}>
                去发现插件
              </Button>
            </Empty>
          ) : (
            <>
              {/* 工具栏 */}
              <div style={{ 
                marginBottom: '24px',
                padding: '16px',
                background: '#f5f5f5',
                borderRadius: '8px',
              }}>
                <Row gutter={[16, 16]} align="middle">
                  <Col xs={24} sm={12} md={8}>
                    <Search
                      placeholder="搜索已安装的插件..."
                      value={searchText}
                      onChange={(e) => setSearchText(e.target.value)}
                      style={{ width: '100%' }}
                    />
                  </Col>
                  <Col xs={12} sm={6} md={4}>
                    <Select
                      value={filterStatus}
                      onChange={setFilterStatus}
                      style={{ width: '100%' }}
                      placeholder="状态"
                    >
                      {statusOptions.map(option => (
                        <Option key={option.value} value={option.value}>
                          {option.label}
                        </Option>
                      ))}
                    </Select>
                  </Col>
                  <Col xs={12} sm={6} md={4}>
                    <Select
                      value={sortBy}
                      onChange={setSortBy}
                      style={{ width: '100%' }}
                      placeholder="排序"
                    >
                      {sortOptions.map(option => (
                        <Option key={option.value} value={option.value}>
                          {option.label}
                        </Option>
                      ))}
                    </Select>
                  </Col>
                  <Col xs={24} sm={12} md={8}>
                    <Space>
                      <Button
                        icon={<SyncOutlined />}
                        onClick={loadInstalledPlugins}
                      >
                        刷新
                      </Button>
                      <Button
                        icon={<CloudDownloadOutlined />}
                        onClick={() => message.info('检查所有插件更新...')}
                      >
                        检查更新
                      </Button>
                    </Space>
                  </Col>
                </Row>
              </div>

              {/* 结果统计 */}
              <div style={{ marginBottom: '16px' }}>
                <Text>
                  显示 {filteredPlugins.length} / {installedPlugins.length} 个插件
                </Text>
              </div>

              {/* 插件列表 */}
              {filteredPlugins.length === 0 ? (
                <Empty description="没有找到匹配的插件" />
              ) : (
                <div>
                  {filteredPlugins.map(plugin => renderPluginCard(plugin))}
                </div>
              )}
            </>
          )}
        </TabPane>
        
        <TabPane tab="更新历史" key="updates">
          <Empty description="更新历史功能开发中..." />
        </TabPane>
        
        <TabPane tab="插件设置" key="settings">
          <Empty description="插件设置功能开发中..." />
        </TabPane>
      </Tabs>
    </div>
  );
};

export default MyPlugins;