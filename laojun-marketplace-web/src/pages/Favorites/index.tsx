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
  Checkbox,
  Modal,
  List,
} from 'antd';
import {
  HeartOutlined,
  HeartFilled,
  AppstoreOutlined,
  DownloadOutlined,
  EyeOutlined,
  ShoppingCartOutlined,
  DeleteOutlined,
  SearchOutlined,
  FilterOutlined,
  ShareAltOutlined,
  ExportOutlined,
} from '@ant-design/icons';
import { Plugin } from '@/types';
import { useCartStore } from '@/stores/cart';
import { useAuthStore } from '@/stores/auth';

const { Title, Text, Paragraph } = Typography;
const { Option } = Select;
const { Search } = Input;

const Favorites: React.FC = () => {
  const navigate = useNavigate();
  const { addToCart, isInCart } = useCartStore();
  const { user, isAuthenticated } = useAuthStore();
  
  const [loading, setLoading] = useState(true);
  const [favorites, setFavorites] = useState<Plugin[]>([]);
  const [filteredFavorites, setFilteredFavorites] = useState<Plugin[]>([]);
  const [selectedItems, setSelectedItems] = useState<string[]>([]);
  const [searchText, setSearchText] = useState('');
  const [sortBy, setSortBy] = useState<string>('added');
  const [filterCategory, setFilterCategory] = useState<string>('all');
  const [viewMode, setViewMode] = useState<'grid' | 'list'>('grid');

  // 模拟收藏的插件数据
  const mockFavorites: Plugin[] = [
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
    },
    {
      id: '2',
      name: 'Design Tools Suite',
      description: '完整的设计工具套件，包含矢量编辑、图像处理等功能',
      version: '3.0.1',
      author: 'Design Studio',
      authorEmail: 'hello@designstudio.com',
      category: 'design',
      tags: ['design', 'vector', 'image'],
      icon: 'https://via.placeholder.com/64',
      screenshots: [],
      downloadUrl: '',
      license: 'Commercial',
      price: 49.99,
      currency: 'USD',
      downloads: 8932,
      rating: 4.6,
      reviewCount: 156,
      size: 5120000,
      requirements: {
        minVersion: '1.0.0',
        dependencies: [],
      },
      status: 'active',
      featured: false,
      createdAt: '2024-01-02T00:00:00Z',
      updatedAt: '2024-01-10T15:20:00Z',
      publishedAt: '2024-01-02T00:00:00Z',
    },
    {
      id: '3',
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
    },
  ];

  const sortOptions = [
    { value: 'added', label: '添加时间' },
    { value: 'name', label: '名称' },
    { value: 'rating', label: '评分' },
    { value: 'downloads', label: '下载量' },
    { value: 'updated', label: '更新时间' },
    { value: 'price', label: '价格' },
  ];

  const categories = [
    { value: 'all', label: '全部分类' },
    { value: 'development', label: '开发工具' },
    { value: 'design', label: '设计工具' },
    { value: 'productivity', label: '效率工具' },
    { value: 'entertainment', label: '娱乐' },
    { value: 'education', label: '教育' },
  ];

  // 如果未登录，重定向到登录页面
  useEffect(() => {
    if (!isAuthenticated) {
      navigate('/login', { state: { from: { pathname: '/favorites' } } });
      return;
    }
  }, [isAuthenticated, navigate]);

  useEffect(() => {
    if (isAuthenticated) {
      loadFavorites();
    }
  }, [isAuthenticated]);

  useEffect(() => {
    filterAndSortFavorites();
  }, [favorites, searchText, sortBy, filterCategory]);

  const loadFavorites = async () => {
    try {
      setLoading(true);
      // 模拟加载收藏数据
      await new Promise(resolve => setTimeout(resolve, 1000));
      setFavorites(mockFavorites);
    } catch (error) {
      console.error('加载收藏失败:', error);
      message.error('加载收藏失败，请刷新页面重试');
    } finally {
      setLoading(false);
    }
  };

  const filterAndSortFavorites = () => {
    let filtered = [...favorites];

    // 搜索过滤
    if (searchText) {
      filtered = filtered.filter(plugin =>
        plugin.name.toLowerCase().includes(searchText.toLowerCase()) ||
        plugin.description.toLowerCase().includes(searchText.toLowerCase()) ||
        plugin.author.toLowerCase().includes(searchText.toLowerCase())
      );
    }

    // 分类过滤
    if (filterCategory !== 'all') {
      filtered = filtered.filter(plugin => plugin.category === filterCategory);
    }

    // 排序
    filtered.sort((a, b) => {
      switch (sortBy) {
        case 'name':
          return a.name.localeCompare(b.name);
        case 'rating':
          return b.rating - a.rating;
        case 'downloads':
          return b.downloads - a.downloads;
        case 'updated':
          return new Date(b.updatedAt).getTime() - new Date(a.updatedAt).getTime();
        case 'price':
          return a.price - b.price;
        case 'added':
        default:
          return new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime();
      }
    });

    setFilteredFavorites(filtered);
  };

  const handleRemoveFromFavorites = (pluginId: string) => {
    Modal.confirm({
      title: '确认移除',
      content: '确定要从收藏夹中移除这个插件吗？',
      onOk: () => {
        setFavorites(prev => prev.filter(p => p.id !== pluginId));
        setSelectedItems(prev => prev.filter(id => id !== pluginId));
        message.success('已从收藏夹移除');
      },
    });
  };

  const handleBatchRemove = () => {
    if (selectedItems.length === 0) {
      message.warning('请先选择要移除的插件');
      return;
    }

    Modal.confirm({
      title: '批量移除',
      content: `确定要从收藏夹中移除选中的 ${selectedItems.length} 个插件吗？`,
      onOk: () => {
        setFavorites(prev => prev.filter(p => !selectedItems.includes(p.id)));
        setSelectedItems([]);
        message.success(`已移除 ${selectedItems.length} 个插件`);
      },
    });
  };

  const handleSelectAll = (checked: boolean) => {
    if (checked) {
      setSelectedItems(filteredFavorites.map(p => p.id));
    } else {
      setSelectedItems([]);
    }
  };

  const handleSelectItem = (pluginId: string, checked: boolean) => {
    if (checked) {
      setSelectedItems(prev => [...prev, pluginId]);
    } else {
      setSelectedItems(prev => prev.filter(id => id !== pluginId));
    }
  };

  const handlePluginClick = (pluginId: string) => {
    navigate(`/plugin/${pluginId}`);
  };

  const handleAddToCart = (plugin: Plugin, e: React.MouseEvent) => {
    e.stopPropagation();
    addToCart(plugin);
  };

  const handleInstall = (plugin: Plugin, e: React.MouseEvent) => {
    e.stopPropagation();
    if (plugin.price === 0) {
      message.success(`正在安装 ${plugin.name}...`);
    } else {
      addToCart(plugin);
    }
  };

  const handleExport = () => {
    const exportData = {
      exportTime: new Date().toISOString(),
      totalCount: favorites.length,
      favorites: favorites.map(plugin => ({
        id: plugin.id,
        name: plugin.name,
        version: plugin.version,
        author: plugin.author,
        category: plugin.category,
        price: plugin.price,
        addedAt: plugin.createdAt,
      })),
    };

    const blob = new Blob([JSON.stringify(exportData, null, 2)], {
      type: 'application/json',
    });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `favorites-${new Date().toISOString().split('T')[0]}.json`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
    
    message.success('收藏夹已导出');
  };

  const renderGridView = () => (
    <Row gutter={[24, 24]}>
      {filteredFavorites.map((plugin) => (
        <Col xs={24} sm={12} md={8} lg={6} key={plugin.id}>
          <Card
            hoverable
            style={{ height: '100%' }}
            onClick={() => handlePluginClick(plugin.id)}
            cover={
              <div style={{ 
                height: '120px', 
                display: 'flex', 
                alignItems: 'center', 
                justifyContent: 'center',
                background: '#f5f5f5',
                position: 'relative',
              }}>
                <Checkbox
                  checked={selectedItems.includes(plugin.id)}
                  onChange={(e) => handleSelectItem(plugin.id, e.target.checked)}
                  onClick={(e) => e.stopPropagation()}
                  style={{ position: 'absolute', top: '8px', left: '8px' }}
                />
                <Avatar
                  size={64}
                  src={plugin.icon}
                  icon={<AppstoreOutlined />}
                />
              </div>
            }
            actions={[
              <Button
                type="text"
                icon={<EyeOutlined />}
                onClick={(e) => {
                  e.stopPropagation();
                  handlePluginClick(plugin.id);
                }}
              >
                查看
              </Button>,
              plugin.price === 0 ? (
                <Button
                  type="text"
                  icon={<DownloadOutlined />}
                  onClick={(e) => handleInstall(plugin, e)}
                >
                  安装
                </Button>
              ) : (
                <Button
                  type="text"
                  icon={<ShoppingCartOutlined />}
                  onClick={(e) => handleAddToCart(plugin, e)}
                  disabled={isInCart(plugin.id)}
                >
                  {isInCart(plugin.id) ? '已在购物车' : '加入购物车'}
                </Button>
              ),
              <Button
                type="text"
                danger
                icon={<HeartFilled />}
                onClick={(e) => {
                  e.stopPropagation();
                  handleRemoveFromFavorites(plugin.id);
                }}
              >
                移除
              </Button>,
            ]}
          >
            <Card.Meta
              title={
                <div style={{ 
                  display: 'flex', 
                  justifyContent: 'space-between',
                  alignItems: 'flex-start',
                }}>
                  <Text strong style={{ fontSize: '14px' }}>
                    {plugin.name}
                  </Text>
                  {plugin.featured && (
                    <Tag color="gold" size="small">推荐</Tag>
                  )}
                </div>
              }
              description={
                <div>
                  <Paragraph 
                    ellipsis={{ rows: 2 }} 
                    style={{ fontSize: '12px', marginBottom: '8px' }}
                  >
                    {plugin.description}
                  </Paragraph>
                  
                  <Space direction="vertical" size="small" style={{ width: '100%' }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                      <Space size="small">
                        <Rate 
                          disabled 
                          defaultValue={plugin.rating} 
                          style={{ fontSize: '12px' }} 
                        />
                        <Text style={{ fontSize: '12px' }}>
                          ({plugin.reviewCount})
                        </Text>
                      </Space>
                      <Text style={{ fontSize: '12px', color: '#666' }}>
                        {plugin.downloads.toLocaleString()} 下载
                      </Text>
                    </div>
                    
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                      <Text style={{ fontSize: '12px', color: '#666' }}>
                        by {plugin.author}
                      </Text>
                      <Text strong style={{ color: plugin.price === 0 ? '#52c41a' : '#1890ff' }}>
                        {plugin.price === 0 ? '免费' : `$${plugin.price}`}
                      </Text>
                    </div>
                  </Space>
                </div>
              }
            />
          </Card>
        </Col>
      ))}
    </Row>
  );

  const renderListView = () => (
    <List
      dataSource={filteredFavorites}
      renderItem={(plugin) => (
        <List.Item
          actions={[
            <Button
              type="text"
              icon={<EyeOutlined />}
              onClick={() => handlePluginClick(plugin.id)}
            >
              查看
            </Button>,
            plugin.price === 0 ? (
              <Button
                type="text"
                icon={<DownloadOutlined />}
                onClick={(e) => handleInstall(plugin, e)}
              >
                安装
              </Button>
            ) : (
              <Button
                type="text"
                icon={<ShoppingCartOutlined />}
                onClick={(e) => handleAddToCart(plugin, e)}
                disabled={isInCart(plugin.id)}
              >
                {isInCart(plugin.id) ? '已在购物车' : '加入购物车'}
              </Button>
            ),
            <Button
              type="text"
              danger
              icon={<HeartFilled />}
              onClick={() => handleRemoveFromFavorites(plugin.id)}
            >
              移除
            </Button>,
          ]}
        >
          <List.Item.Meta
            avatar={
              <Space>
                <Checkbox
                  checked={selectedItems.includes(plugin.id)}
                  onChange={(e) => handleSelectItem(plugin.id, e.target.checked)}
                />
                <Avatar
                  src={plugin.icon}
                  icon={<AppstoreOutlined />}
                  size={48}
                />
              </Space>
            }
            title={
              <Space>
                <Text strong>{plugin.name}</Text>
                <Text type="secondary">v{plugin.version}</Text>
                {plugin.featured && <Tag color="gold" size="small">推荐</Tag>}
              </Space>
            }
            description={
              <Space direction="vertical" size="small" style={{ width: '100%' }}>
                <Text>{plugin.description}</Text>
                <Space>
                  <Tag>{plugin.category}</Tag>
                  <Text type="secondary">by {plugin.author}</Text>
                  <Rate disabled defaultValue={plugin.rating} style={{ fontSize: '12px' }} />
                  <Text type="secondary">({plugin.reviewCount})</Text>
                  <Text type="secondary">{plugin.downloads.toLocaleString()} 下载</Text>
                  <Text strong style={{ color: plugin.price === 0 ? '#52c41a' : '#1890ff' }}>
                    {plugin.price === 0 ? '免费' : `$${plugin.price}`}
                  </Text>
                </Space>
              </Space>
            }
          />
        </List.Item>
      )}
    />
  );

  // 如果未登录，不渲染内容（会被重定向到登录页面）
  if (!isAuthenticated) {
    return null;
  }

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

  return (
    <div style={{ padding: '24px', maxWidth: '1200px', margin: '0 auto' }}>
      <div style={{ marginBottom: '24px' }}>
        <Title level={2}>
          <HeartOutlined /> 我的收藏
        </Title>
        <Text type="secondary">
          您收藏了 {favorites.length} 个插件
        </Text>
      </div>

      {favorites.length === 0 ? (
        <Empty
          description="您还没有收藏任何插件"
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
                  placeholder="搜索收藏的插件..."
                  value={searchText}
                  onChange={(e) => setSearchText(e.target.value)}
                  style={{ width: '100%' }}
                />
              </Col>
              <Col xs={12} sm={6} md={4}>
                <Select
                  value={filterCategory}
                  onChange={setFilterCategory}
                  style={{ width: '100%' }}
                  placeholder="分类"
                >
                  {categories.map(cat => (
                    <Option key={cat.value} value={cat.value}>
                      {cat.label}
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
                  <Checkbox
                    checked={selectedItems.length === filteredFavorites.length && filteredFavorites.length > 0}
                    indeterminate={selectedItems.length > 0 && selectedItems.length < filteredFavorites.length}
                    onChange={(e) => handleSelectAll(e.target.checked)}
                  >
                    全选
                  </Checkbox>
                  <Button
                    danger
                    icon={<DeleteOutlined />}
                    onClick={handleBatchRemove}
                    disabled={selectedItems.length === 0}
                  >
                    批量移除
                  </Button>
                  <Button
                    icon={<ExportOutlined />}
                    onClick={handleExport}
                  >
                    导出
                  </Button>
                </Space>
              </Col>
            </Row>
          </div>

          {/* 结果统计 */}
          <div style={{ marginBottom: '16px' }}>
            <Text>
              显示 {filteredFavorites.length} / {favorites.length} 个插件
              {selectedItems.length > 0 && ` (已选择 ${selectedItems.length} 个)`}
            </Text>
          </div>

          {/* 插件列表 */}
          {filteredFavorites.length === 0 ? (
            <Empty description="没有找到匹配的插件" />
          ) : (
            renderGridView()
          )}
        </>
      )}
    </div>
  );
};

export default Favorites;