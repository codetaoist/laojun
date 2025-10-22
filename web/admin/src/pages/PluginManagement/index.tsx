import { useState, useEffect } from 'react';
import { 
  Table, Card, Button, Space, Input, Tag, message, Modal, Form, 
  Select, Popconfirm, Upload, Drawer, Descriptions, Switch, 
  Progress, Alert, Row, Col, Statistic, Divider, Tabs, List,
  Avatar, Rate, Badge, Tooltip, Steps, Timeline, Empty
} from 'antd';
import { 
  PlusOutlined, SearchOutlined, EditOutlined, DeleteOutlined, 
  ReloadOutlined, UploadOutlined, DownloadOutlined, SettingOutlined,
  AppstoreOutlined, CloudUploadOutlined, InfoCircleOutlined,
  CheckCircleOutlined, ExclamationCircleOutlined, StopOutlined,
  ShopOutlined, LinkOutlined, WarningOutlined, ClockCircleOutlined,
  StarOutlined, HeartOutlined, EyeOutlined, TeamOutlined
} from '@ant-design/icons';
import { Plugin } from '@/types';
import { pluginService } from '@/services/plugin';

const { Search } = Input;
const { Option } = Select;
const { Dragger } = Upload;
const { TabPane } = Tabs;

interface PluginStats {
  total: number;
  enabled: number;
  disabled: number;
  system: number;
}

// 插件市场插件接口
interface MarketplacePlugin {
  id: string;
  name: string;
  displayName: string;
  description: string;
  version: string;
  author: string;
  category: string;
  rating: number;
  downloads: number;
  price: number;
  isFree: boolean;
  screenshots: string[];
  dependencies: string[];
  tags: string[];
  lastUpdated: string;
  isInstalled: boolean;
}

// 安装进度接口
interface InstallProgress {
  pluginId: string;
  step: number;
  totalSteps: number;
  currentStep: string;
  progress: number;
  status: 'installing' | 'success' | 'error';
  error?: string;
}

const PluginManagement: React.FC = () => {
  const [plugins, setPlugins] = useState<Plugin[]>([]);
  const [loading, setLoading] = useState(false);
  const [searchText, setSearchText] = useState('');
  const [selectedRowKeys, setSelectedRowKeys] = useState<string[]>([]);
  const [uploadModalVisible, setUploadModalVisible] = useState(false);
  const [detailDrawerVisible, setDetailDrawerVisible] = useState(false);
  const [configModalVisible, setConfigModalVisible] = useState(false);
  const [currentPlugin, setCurrentPlugin] = useState<Plugin | null>(null);
  const [uploadProgress, setUploadProgress] = useState(0);
  const [stats, setStats] = useState<PluginStats>({
    total: 0,
    enabled: 0,
    disabled: 0,
    system: 0,
  });
  
  // 插件市场相关状态
  const [marketplaceVisible, setMarketplaceVisible] = useState(false);
  const [marketplacePlugins, setMarketplacePlugins] = useState<MarketplacePlugin[]>([]);
  const [marketplaceLoading, setMarketplaceLoading] = useState(false);
  const [selectedCategory, setSelectedCategory] = useState<string>('all');
  const [categories, setCategories] = useState<Array<{ id: string; name: string }>>([]);
  const [installProgress, setInstallProgress] = useState<InstallProgress | null>(null);
  const [installModalVisible, setInstallModalVisible] = useState(false);

  // 模拟插件数据
  const mockPlugins: Plugin[] = [
    {
      id: '1',
      name: 'user-auth',
      displayName: '用户认证插件',
      description: '提供用户认证和授权功能',
      version: '1.0.0',
      author: 'System',
      status: 'enabled',
      isSystem: true,
      config: {
        enableTwoFactor: true,
        sessionTimeout: 3600,
      },
      routes: [
        { path: '/auth/login', component: 'LoginComponent' },
        { path: '/auth/register', component: 'RegisterComponent' },
      ],
      menus: [
        { title: '认证管理', path: '/auth', icon: 'UserOutlined' },
      ],
      permissions: ['auth:read', 'auth:write'],
      createdAt: '2024-01-01T00:00:00Z',
      updatedAt: '2024-01-15T10:30:00Z',
    },
    {
      id: '2',
      name: 'file-manager',
      displayName: '文件管理器',
      description: '文件上传、下载和管理功能',
      version: '2.1.0',
      author: 'Admin',
      status: 'enabled',
      isSystem: false,
      config: {
        maxFileSize: 10485760,
        allowedTypes: ['jpg', 'png', 'pdf', 'doc'],
      },
      routes: [
        { path: '/files', component: 'FileManagerComponent' },
      ],
      menus: [
        { title: '文件管理', path: '/files', icon: 'FolderOutlined' },
      ],
      permissions: ['file:read', 'file:write', 'file:delete'],
      createdAt: '2024-01-02T00:00:00Z',
      updatedAt: '2024-01-10T15:20:00Z',
    },
    {
      id: '3',
      name: 'notification',
      displayName: '通知系统',
      description: '系统通知和消息推送功能',
      version: '1.5.0',
      author: 'Developer',
      status: 'disabled',
      isSystem: false,
      config: {
        enableEmail: true,
        enableSMS: false,
        enablePush: true,
      },
      routes: [
        { path: '/notifications', component: 'NotificationComponent' },
      ],
      menus: [
        { title: '通知中心', path: '/notifications', icon: 'BellOutlined' },
      ],
      permissions: ['notification:read', 'notification:send'],
      createdAt: '2024-01-03T00:00:00Z',
      updatedAt: '2024-01-05T09:15:00Z',
    },
  ];

  // 模拟插件市场数据
  const mockMarketplacePlugins: MarketplacePlugin[] = [
    {
      id: 'mp-1',
      name: 'advanced-analytics',
      displayName: '高级数据分析',
      description: '提供强大的数据分析和可视化功能，支持多种图表类型和实时数据监控',
      version: '2.0.0',
      author: 'Analytics Team',
      category: 'analytics',
      rating: 4.8,
      downloads: 15420,
      price: 99.99,
      isFree: false,
      screenshots: ['/screenshots/analytics1.png', '/screenshots/analytics2.png'],
      dependencies: ['chart-lib', 'data-processor'],
      tags: ['数据分析', '图表', '可视化', '监控'],
      lastUpdated: '2024-01-20T00:00:00Z',
      isInstalled: false,
    },
    {
      id: 'mp-2',
      name: 'backup-manager',
      displayName: '备份管理器',
      description: '自动化数据备份和恢复解决方案，支持增量备份和云存储',
      version: '1.3.0',
      author: 'Backup Solutions',
      category: 'utility',
      rating: 4.6,
      downloads: 8930,
      price: 0,
      isFree: true,
      screenshots: ['/screenshots/backup1.png'],
      dependencies: ['cloud-sdk'],
      tags: ['备份', '恢复', '云存储', '自动化'],
      lastUpdated: '2024-01-18T00:00:00Z',
      isInstalled: false,
    },
    {
      id: 'mp-3',
      name: 'security-scanner',
      displayName: '安全扫描器',
      description: '全面的安全漏洞扫描和威胁检测工具',
      version: '3.1.0',
      author: 'Security Corp',
      category: 'security',
      rating: 4.9,
      downloads: 12650,
      price: 199.99,
      isFree: false,
      screenshots: ['/screenshots/security1.png', '/screenshots/security2.png'],
      dependencies: ['security-lib', 'threat-db'],
      tags: ['安全', '扫描', '威胁检测', '漏洞'],
      lastUpdated: '2024-01-22T00:00:00Z',
      isInstalled: false,
    },
    {
      id: 'mp-4',
      name: 'workflow-automation',
      displayName: '工作流自动化',
      description: '可视化工作流设计和自动化执行平台',
      version: '1.8.0',
      author: 'Workflow Inc',
      category: 'automation',
      rating: 4.7,
      downloads: 6780,
      price: 149.99,
      isFree: false,
      screenshots: ['/screenshots/workflow1.png'],
      dependencies: ['workflow-engine'],
      tags: ['工作流', '自动化', '可视化', '流程'],
      lastUpdated: '2024-01-19T00:00:00Z',
      isInstalled: false,
    },
  ];

  // 加载插件数据
  const loadPlugins = async () => {
    setLoading(true);
    try {
      // 模拟 API 调用
      await new Promise(resolve => setTimeout(resolve, 1000));
      setPlugins(mockPlugins);
      
      // 计算统计数据
      const newStats = {
        total: mockPlugins.length,
        enabled: mockPlugins.filter(p => p.status === 'enabled').length,
        disabled: mockPlugins.filter(p => p.status === 'disabled').length,
        system: mockPlugins.filter(p => p.isSystem).length,
      };
      setStats(newStats);
    } catch (error) {
      message.error('加载插件数据失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadPlugins();
  }, []);

  // 表格列配置
  const columns = [
    {
      title: '插件信息',
      key: 'pluginInfo',
      render: (_, record: Plugin) => (
        <Space>
          <AppstoreOutlined style={{ fontSize: 24, color: '#1890ff' }} />
          <div>
            <div style={{ fontWeight: 500 }}>
              {record.displayName}
              {record.isSystem && <Tag color="red" style={{ marginLeft: 8 }}>系统</Tag>}
            </div>
            <div style={{ fontSize: '12px', color: '#999' }}>
              {record.name} v{record.version}
            </div>
          </div>
        </Space>
      ),
    },
    {
      title: '描述',
      dataIndex: 'description',
      key: 'description',
      ellipsis: true,
    },
    {
      title: '作者',
      dataIndex: 'author',
      key: 'author',
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string, record: Plugin) => (
        <Switch
          checked={status === 'enabled'}
          onChange={(checked) => handleStatusChange(record.id, checked)}
          checkedChildren="启用"
          unCheckedChildren="禁用"
          disabled={record.isSystem}
        />
      ),
    },
    {
      title: '权限',
      dataIndex: 'permissions',
      key: 'permissions',
      render: (permissions: string[]) => (
        <Tag color="blue">{permissions?.length || 0} 个权限</Tag>
      ),
    },
    {
      title: '更新时间',
      dataIndex: 'updatedAt',
      key: 'updatedAt',
      render: (date: string) => new Date(date).toLocaleString(),
      sorter: true,
    },
    {
      title: '操作',
      key: 'action',
      render: (_, record: Plugin) => (
        <Space size="middle">
          <Button
            type="link"
            icon={<InfoCircleOutlined />}
            onClick={() => handleViewDetail(record)}
          >
            详情
          </Button>
          <Button
            type="link"
            icon={<SettingOutlined />}
            onClick={() => handleConfig(record)}
          >
            配置
          </Button>
          <Popconfirm
            title="确定要卸载这个插件吗？"
            onConfirm={() => handleUninstall(record.id)}
            disabled={record.isSystem}
          >
            <Button
              type="link"
              danger
              icon={<DeleteOutlined />}
              disabled={record.isSystem}
            >
              卸载
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  // 处理状态切换
  const handleStatusChange = async (pluginId: string, checked: boolean) => {
    try {
      const status = checked ? 'enabled' : 'disabled';
      setPlugins(prev => prev.map(p => 
        p.id === pluginId ? { ...p, status, updatedAt: new Date().toISOString() } : p
      ));
      message.success(`插件已${checked ? '启用' : '禁用'}`);
      
      // 重新计算统计数据
      const updatedPlugins = plugins.map(p => 
        p.id === pluginId ? { ...p, status } : p
      );
      const newStats = {
        total: updatedPlugins.length,
        enabled: updatedPlugins.filter(p => p.status === 'enabled').length,
        disabled: updatedPlugins.filter(p => p.status === 'disabled').length,
        system: updatedPlugins.filter(p => p.isSystem).length,
      };
      setStats(newStats);
    } catch (error) {
      message.error('状态更新失败');
    }
  };

  // 处理查看详情
  const handleViewDetail = (plugin: Plugin) => {
    setCurrentPlugin(plugin);
    setDetailDrawerVisible(true);
  };

  // 处理配置
  const handleConfig = (plugin: Plugin) => {
    setCurrentPlugin(plugin);
    setConfigModalVisible(true);
  };

  // 处理卸载
  const handleUninstall = async (pluginId: string) => {
    try {
      await new Promise(resolve => setTimeout(resolve, 500));
      setPlugins(prev => prev.filter(p => p.id !== pluginId));
      message.success('插件卸载成功');
      loadPlugins(); // 重新加载以更新统计数据
    } catch (error) {
      message.error('插件卸载失败');
    }
  };

  // 搜索插件
  const handleSearch = (value: string) => {
    setSearchText(value);
  };

  // 加载插件市场数据
  const loadMarketplaceCategories = async () => {
    try {
      const items = await pluginService.getMarketplaceCategories();
      const mapped = (items || []).map((c: any) => ({
        id: String(c.id ?? c.name ?? Math.random().toString(36).slice(2)),
        name: c.name ?? String(c.id ?? ''),
      }));
      setCategories(mapped);
    } catch (error) {
      message.warning('加载分类失败，使用默认选项');
      setCategories([]);
    }
  };

  const loadMarketplacePlugins = async () => {
    setMarketplaceLoading(true);
    try {
      const items = await pluginService.getMarketplacePlugins({
        category: selectedCategory === 'all' ? undefined : selectedCategory,
        pageSize: 12,
      });
      const mapped = (items || []).map((p: any) => ({
        id: p.id,
        name: p.name,
        displayName: p.name,
        description: p.description || '',
        version: p.version || '1.0.0',
        author: p.author || 'Unknown',
        category: selectedCategory === 'all' ? 'general' : selectedCategory,
        rating: 0,
        downloads: 0,
        price: 0,
        isFree: true,
        screenshots: [],
        dependencies: p.dependencies || [],
        tags: p.permissions || [],
        lastUpdated: p.updatedAt || new Date().toISOString(),
        isInstalled: false,
      }));
      setMarketplacePlugins(mapped);
    } catch (error) {
      message.error('加载插件市场失败');
    } finally {
      setMarketplaceLoading(false);
    }
  };

  // 从市场安装插件
  const handleInstallFromMarket = async (marketPlugin: MarketplacePlugin) => {
    setInstallProgress({
      pluginId: marketPlugin.id,
      step: 1,
      totalSteps: 2,
      currentStep: '触发安装',
      progress: 30,
      status: 'installing',
    });
    setInstallModalVisible(true);

    try {
      await pluginService.installFromMarketplace(marketPlugin.id);

      setInstallProgress(prev => prev ? {
        ...prev,
        step: 2,
        currentStep: '安装完成',
        progress: 100,
        status: 'success',
      } : null);

      const newPlugin: Plugin = {
        id: marketPlugin.id,
        name: marketPlugin.name,
        displayName: marketPlugin.displayName,
        description: marketPlugin.description,
        version: marketPlugin.version,
        author: marketPlugin.author,
        status: 'disabled',
        isSystem: false,
        config: {},
        routes: [],
        menus: [],
        permissions: [],
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      };

      setPlugins(prev => [...prev, newPlugin]);
      setMarketplacePlugins(prev => prev.map(p => 
        p.id === marketPlugin.id ? { ...p, isInstalled: true } : p
      ));

      message.success('插件安装成功');

      setTimeout(() => {
        setInstallModalVisible(false);
        setInstallProgress(null);
      }, 1500);
    } catch (error) {
      setInstallProgress(prev => prev ? {
        ...prev,
        status: 'error',
        currentStep: '安装失败',
        error: '安装失败，请重试',
      } : null);
      message.error('插件安装失败');
    }
  };

  // 打开插件市场
  const handleOpenMarketplace = () => {
    setMarketplaceVisible(true);
    loadMarketplaceCategories();
    if (marketplacePlugins.length === 0) {
      loadMarketplacePlugins();
    }
  };

  // 处理上传
  const handleUpload = {
    name: 'plugin',
    multiple: false,
    accept: '.zip,.tar.gz',
    beforeUpload: () => false, // 阻止自动上传
    onChange: (info: any) => {
      const { status } = info.file;
      if (status === 'uploading') {
        setUploadProgress(info.file.percent || 0);
      } else if (status === 'done') {
        message.success('插件上传成功');
        setUploadModalVisible(false);
        setUploadProgress(0);
        loadPlugins();
      } else if (status === 'error') {
        message.error('插件上传失败');
        setUploadProgress(0);
      }
    },
  };

  // 行选择配置
  const rowSelection = {
    selectedRowKeys,
    onChange: (keys: React.Key[]) => {
      setSelectedRowKeys(keys as string[]);
    },
    getCheckboxProps: (record: Plugin) => ({
      disabled: record.isSystem,
    }),
  };

  return (
    <div>
      {/* 统计卡片 */}
      <Row gutter={16} style={{ marginBottom: 16 }}>
        <Col span={6}>
          <Card>
            <Statistic
              title="总插件数"
              value={stats.total}
              prefix={<AppstoreOutlined />}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="已启用"
              value={stats.enabled}
              prefix={<CheckCircleOutlined />}
              valueStyle={{ color: '#3f8600' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="已禁用"
              value={stats.disabled}
              prefix={<StopOutlined />}
              valueStyle={{ color: '#cf1322' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="系统插件"
              value={stats.system}
              prefix={<ExclamationCircleOutlined />}
              valueStyle={{ color: '#fa8c16' }}
            />
          </Card>
        </Col>
      </Row>

      <Card>
        {/* 操作栏 */}
        <div style={{ marginBottom: 16 }}>
          <Row gutter={16}>
            <Col flex="auto">
              <Space>
                <Button
                  type="primary"
                  icon={<ShopOutlined />}
                  onClick={handleOpenMarketplace}
                >
                  插件市场
                </Button>
                <Button
                  icon={<CloudUploadOutlined />}
                  onClick={() => setUploadModalVisible(true)}
                >
                  本地安装
                </Button>
                <Button
                  danger
                  icon={<DeleteOutlined />}
                  disabled={selectedRowKeys.length === 0}
                  onClick={() => {
                    Modal.confirm({
                      title: '批量卸载插件',
                      content: `确定要卸载选中的 ${selectedRowKeys.length} 个插件吗？`,
                      onOk: async () => {
                        try {
                          await new Promise(resolve => setTimeout(resolve, 1000));
                          setPlugins(prev => prev.filter(p => !selectedRowKeys.includes(p.id)));
                          setSelectedRowKeys([]);
                          message.success('批量卸载成功');
                          loadPlugins();
                        } catch (error) {
                          message.error('批量卸载失败');
                        }
                      },
                    });
                  }}
                >
                  批量卸载
                </Button>
                <Button
                  icon={<ReloadOutlined />}
                  onClick={loadPlugins}
                >
                  刷新
                </Button>
              </Space>
            </Col>
            <Col>
              <Search
                placeholder="搜索插件名称或描述"
                allowClear
                style={{ width: 300 }}
                onSearch={handleSearch}
              />
            </Col>
          </Row>
        </div>

        {/* 表格 */}
        <Table
          columns={columns}
          dataSource={plugins}
          rowKey="id"
          loading={loading}
          rowSelection={rowSelection}
          pagination={{
            total: plugins.length,
            pageSize: 10,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total, range) =>
              `第 ${range[0]}-${range[1]} 条/共 ${total} 条`,
          }}
        />
      </Card>

      {/* 上传插件模态框 */}
      <Modal
        title="安装插件"
        open={uploadModalVisible}
        onCancel={() => {
          setUploadModalVisible(false);
          setUploadProgress(0);
        }}
        footer={null}
        width={600}
      >
        <Alert
          message="插件安装说明"
          description="请上传 .zip 或 .tar.gz 格式的插件包。插件包应包含插件配置文件和相关资源。"
          type="info"
          showIcon
          style={{ marginBottom: 16 }}
        />
        
        <Dragger {...handleUpload}>
          <p className="ant-upload-drag-icon">
            <UploadOutlined />
          </p>
          <p className="ant-upload-text">点击或拖拽文件到此区域上传</p>
          <p className="ant-upload-hint">
            支持单个文件上传，仅支持 .zip 和 .tar.gz 格式
          </p>
        </Dragger>

        {uploadProgress > 0 && (
          <div style={{ marginTop: 16 }}>
            <Progress percent={uploadProgress} />
          </div>
        )}
      </Modal>

      {/* 插件详情抽屉 */}
      <Drawer
        title="插件详情"
        placement="right"
        onClose={() => setDetailDrawerVisible(false)}
        open={detailDrawerVisible}
        width={600}
      >
        {currentPlugin && (
          <div>
            <Descriptions title={currentPlugin.displayName} bordered>
              <Descriptions.Item label="插件名称" span={3}>
                {currentPlugin.name}
              </Descriptions.Item>
              <Descriptions.Item label="版本" span={3}>
                {currentPlugin.version}
              </Descriptions.Item>
              <Descriptions.Item label="作者" span={3}>
                {currentPlugin.author}
              </Descriptions.Item>
              <Descriptions.Item label="状态" span={3}>
                <Tag color={currentPlugin.status === 'enabled' ? 'green' : 'red'}>
                  {currentPlugin.status === 'enabled' ? '已启用' : '已禁用'}
                </Tag>
                {currentPlugin.isSystem && <Tag color="red">系统插件</Tag>}
              </Descriptions.Item>
              <Descriptions.Item label="描述" span={3}>
                {currentPlugin.description}
              </Descriptions.Item>
              <Descriptions.Item label="权限" span={3}>
                <Space wrap>
                  {currentPlugin.permissions?.map(permission => (
                    <Tag key={permission} color="blue">{permission}</Tag>
                  ))}
                </Space>
              </Descriptions.Item>
              <Descriptions.Item label="路由" span={3}>
                {currentPlugin.routes?.map(route => (
                  <div key={route.path}>
                    <Tag color="cyan">{route.path}</Tag> → {route.component}
                  </div>
                ))}
              </Descriptions.Item>
              <Descriptions.Item label="菜单" span={3}>
                {currentPlugin.menus?.map(menu => (
                  <div key={menu.path}>
                    <Tag color="purple">{menu.title}</Tag> ({menu.path})
                  </div>
                ))}
              </Descriptions.Item>
              <Descriptions.Item label="创建时间" span={3}>
                {new Date(currentPlugin.createdAt).toLocaleString()}
              </Descriptions.Item>
              <Descriptions.Item label="更新时间" span={3}>
                {new Date(currentPlugin.updatedAt).toLocaleString()}
              </Descriptions.Item>
            </Descriptions>
          </div>
        )}
      </Drawer>

      {/* 插件配置模态框 */}
      <Modal
        title={`配置插件 - ${currentPlugin?.displayName}`}
        open={configModalVisible}
        onCancel={() => setConfigModalVisible(false)}
        onOk={async () => {
          try {
            await new Promise(resolve => setTimeout(resolve, 500));
            message.success('配置保存成功');
            setConfigModalVisible(false);
          } catch (error) {
            message.error('配置保存失败');
          }
        }}
        width={600}
      >
        {currentPlugin && (
          <div>
            <Alert
              message="插件配置"
              description="修改插件配置可能需要重启插件才能生效。"
              type="warning"
              showIcon
              style={{ marginBottom: 16 }}
            />
            
            <Form layout="vertical">
              {currentPlugin.config && Object.entries(currentPlugin.config).map(([key, value]) => (
                <Form.Item key={key} label={key} name={key} initialValue={value}>
                  {typeof value === 'boolean' ? (
                    <Switch defaultChecked={value} />
                  ) : typeof value === 'number' ? (
                    <Input type="number" defaultValue={value} />
                  ) : Array.isArray(value) ? (
                    <Select mode="tags" defaultValue={value} style={{ width: '100%' }} />
                  ) : (
                    <Input defaultValue={value} />
                  )}
                </Form.Item>
              ))}
            </Form>
          </div>
        )}
      </Modal>

      {/* 插件市场模态框 */}
      <Modal
        title="插件市场"
        open={marketplaceVisible}
        onCancel={() => setMarketplaceVisible(false)}
        footer={null}
        width={1200}
        style={{ top: 20 }}
      >
        <Tabs defaultActiveKey="all">
          <TabPane tab="全部" key="all">
            <div style={{ marginBottom: 16 }}>
              <Row gutter={16}>
                <Col span={8}>
                  <Select
                     placeholder="选择分类"
                     style={{ width: '100%' }}
                     value={selectedCategory}
                     onChange={setSelectedCategory}
                   >
                     <Option value="all">全部分类</Option>
                     {categories.map(c => (
                       <Option key={c.id} value={c.id}>{c.name}</Option>
                     ))}
                   </Select>
                </Col>
                <Col span={16}>
                  <Search placeholder="搜索插件..." />
                </Col>
              </Row>
            </div>
            
            <List
              loading={marketplaceLoading}
              grid={{ gutter: 16, column: 2 }}
              dataSource={marketplacePlugins.filter(plugin => 
                selectedCategory === 'all' || plugin.category === selectedCategory
              )}
              renderItem={(plugin) => (
                <List.Item>
                  <Card
                    hoverable
                    actions={[
                      <Tooltip title="查看详情">
                        <EyeOutlined key="view" />
                      </Tooltip>,
                      <Tooltip title="收藏">
                        <HeartOutlined key="favorite" />
                      </Tooltip>,
                      plugin.isInstalled ? (
                        <Button type="text" disabled>已安装</Button>
                      ) : (
                        <Button 
                          type="primary" 
                          size="small"
                          onClick={() => handleInstallFromMarket(plugin)}
                        >
                          {plugin.isFree ? '免费安装' : `¥${plugin.price}`}
                        </Button>
                      ),
                    ]}
                  >
                    <Card.Meta
                      avatar={<Avatar icon={<AppstoreOutlined />} />}
                      title={
                        <div>
                          <span>{plugin.displayName}</span>
                          <div style={{ float: 'right' }}>
                            <Rate disabled defaultValue={plugin.rating} style={{ fontSize: 12 }} />
                            <span style={{ marginLeft: 4, fontSize: 12, color: '#999' }}>
                              ({plugin.rating})
                            </span>
                          </div>
                        </div>
                      }
                      description={
                        <div>
                          <p style={{ marginBottom: 8 }}>{plugin.description}</p>
                          <div style={{ fontSize: 12, color: '#999' }}>
                            <Space split={<span>•</span>}>
                              <span>v{plugin.version}</span>
                              <span>{plugin.author}</span>
                              <span><TeamOutlined /> {plugin.downloads.toLocaleString()}</span>
                            </Space>
                          </div>
                          <div style={{ marginTop: 8 }}>
                            {plugin.tags.map(tag => (
                              <Tag key={tag} size="small">{tag}</Tag>
                            ))}
                          </div>
                        </div>
                      }
                    />
                  </Card>
                </List.Item>
              )}
            />
          </TabPane>
          <TabPane tab="已安装" key="installed">
            <Empty description="暂无已安装的市场插件" />
          </TabPane>
        </Tabs>
      </Modal>

      {/* 安装进度模态框 */}
      <Modal
        title="安装插件"
        open={installModalVisible}
        footer={null}
        closable={false}
        width={500}
      >
        {installProgress && (
          <div>
            <Steps
              current={installProgress.step - 1}
              status={installProgress.status === 'error' ? 'error' : 'process'}
              size="small"
              style={{ marginBottom: 24 }}
            >
              <Steps.Step title="检查依赖" />
              <Steps.Step title="下载插件" />
              <Steps.Step title="安装插件" />
              <Steps.Step title="配置插件" />
            </Steps>
            
            <div style={{ textAlign: 'center', marginBottom: 16 }}>
              <div style={{ fontSize: 16, marginBottom: 8 }}>
                {installProgress.currentStep}
              </div>
              <Progress 
                percent={installProgress.progress} 
                status={installProgress.status === 'error' ? 'exception' : 'active'}
              />
            </div>
            
            {installProgress.status === 'success' && (
              <Alert
                message="安装成功"
                description="插件已成功安装，您可以在插件列表中查看和配置。"
                type="success"
                showIcon
              />
            )}
            
            {installProgress.status === 'error' && (
              <Alert
                message="安装失败"
                description={installProgress.error}
                type="error"
                showIcon
              />
            )}
          </div>
        )}
      </Modal>
    </div>
  );
};

export default PluginManagement;