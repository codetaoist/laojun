import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import {
  Row,
  Col,
  Card,
  Typography,
  Space,
  Spin,
  message,
  Tag,
  Rate,
  Button,
  Select,
  Pagination,
  Avatar,
  Breadcrumb,
  Empty,
} from 'antd';
import {
  AppstoreOutlined,
  DownloadOutlined,
  StarOutlined,
  EyeOutlined,
  HomeOutlined,
  FolderOutlined,
  UserOutlined,
  ShoppingCartOutlined,
} from '@ant-design/icons';
import { Plugin, Category as CategoryType } from '@/types';
import { pluginService } from '@/services/plugin';
import { useCartStore } from '@/stores/cart';

const { Title, Paragraph, Text } = Typography;
const { Option } = Select;

const Category: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { addToCart, isInCart } = useCartStore();
  
  const [loading, setLoading] = useState(true);
  const [category, setCategory] = useState<CategoryType | null>(null);
  const [plugins, setPlugins] = useState<Plugin[]>([]);
  const [total, setTotal] = useState(0);
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(12);
  const [sortBy, setSortBy] = useState<string>('downloads');

  // 模拟分类数据
  const mockCategories: Record<string, CategoryType> = {
    development: {
      id: 'development',
      name: '开发工具',
      description: '代码编辑器、调试工具、版本控制等开发相关插件',
      icon: 'development',
      pluginCount: 156,
    },
    productivity: {
      id: 'productivity',
      name: '效率工具',
      description: '提升工作效率的各类工具和插件',
      icon: 'productivity',
      pluginCount: 89,
    },
    design: {
      id: 'design',
      name: '设计工具',
      description: '图形设计、UI/UX设计相关的插件和工具',
      icon: 'design',
      pluginCount: 67,
    },
  };

  // 模拟插件数据
  const mockPlugins: Plugin[] = [
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
      name: 'Git Helper',
      description: '简化Git操作的可视化工具，让版本控制变得更加简单',
      version: '1.5.2',
      author: 'GitTools Team',
      authorEmail: 'team@gittools.com',
      category: 'development',
      tags: ['git', 'version-control', 'visual'],
      icon: 'https://via.placeholder.com/64',
      screenshots: [],
      downloadUrl: '',
      license: 'Apache 2.0',
      price: 0,
      currency: 'USD',
      downloads: 8932,
      rating: 4.6,
      reviewCount: 156,
      size: 1024000,
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
  ];

  const sortOptions = [
    { value: 'downloads', label: '下载量' },
    { value: 'rating', label: '评分' },
    { value: 'updated', label: '更新时间' },
    { value: 'created', label: '发布时间' },
    { value: 'name', label: '名称' },
    { value: 'price', label: '价格' },
  ];

  useEffect(() => {
    if (id) {
      loadCategoryData();
    }
  }, [id, currentPage, pageSize, sortBy]);

  const loadCategoryData = async () => {
    try {
      setLoading(true);
      
      // 加载分类信息
      const categoryData = mockCategories[id!];
      if (!categoryData) {
        message.error('分类不存在');
        navigate('/404');
        return;
      }
      setCategory(categoryData);

      // 模拟加载插件数据
      await new Promise(resolve => setTimeout(resolve, 800));
      setPlugins(mockPlugins);
      setTotal(mockPlugins.length);
    } catch (error) {
      console.error('加载分类数据失败:', error);
      message.error('加载数据失败，请刷新页面重试');
    } finally {
      setLoading(false);
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

  const handlePageChange = (page: number, size?: number) => {
    setCurrentPage(page);
    if (size) setPageSize(size);
  };

  const handleSortChange = (value: string) => {
    setSortBy(value);
    setCurrentPage(1);
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

  if (!category) {
    return (
      <div style={{ padding: '24px', textAlign: 'center' }}>
        <Empty description="分类不存在" />
      </div>
    );
  }

  return (
    <div style={{ padding: '24px', maxWidth: '1200px', margin: '0 auto' }}>
      {/* 面包屑导航 */}
      <Breadcrumb style={{ marginBottom: '24px' }}>
        <Breadcrumb.Item>
          <HomeOutlined />
          <span onClick={() => navigate('/')} style={{ cursor: 'pointer' }}>
            首页
          </span>
        </Breadcrumb.Item>
        <Breadcrumb.Item>
          <FolderOutlined />
          <span onClick={() => navigate('/categories')} style={{ cursor: 'pointer' }}>
            分类
          </span>
        </Breadcrumb.Item>
        <Breadcrumb.Item>
          <AppstoreOutlined />
          {category.name}
        </Breadcrumb.Item>
      </Breadcrumb>

      {/* 分类头部 */}
      <div style={{ marginBottom: '32px' }}>
        <Title level={1}>
          <AppstoreOutlined /> {category.name}
        </Title>
        <Paragraph style={{ fontSize: '16px', color: '#666', marginBottom: '16px' }}>
          {category.description}
        </Paragraph>
        <Tag color="blue">{category.pluginCount} 个插件</Tag>
      </div>

      {/* 排序和筛选 */}
      <div style={{ 
        display: 'flex', 
        justifyContent: 'space-between', 
        alignItems: 'center',
        marginBottom: '24px',
        padding: '16px',
        background: '#f5f5f5',
        borderRadius: '8px',
      }}>
        <Text>共找到 {total} 个插件</Text>
        <Space>
          <Text>排序方式：</Text>
          <Select
            value={sortBy}
            onChange={handleSortChange}
            style={{ width: 120 }}
          >
            {sortOptions.map(option => (
              <Option key={option.value} value={option.value}>
                {option.label}
              </Option>
            ))}
          </Select>
        </Space>
      </div>

      {/* 插件网格 */}
      {plugins.length === 0 ? (
        <Empty description="暂无插件" />
      ) : (
        <>
          <Row gutter={[24, 24]}>
            {plugins.map((plugin) => (
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
                    }}>
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

          {/* 分页 */}
          <div style={{ textAlign: 'center', marginTop: '32px' }}>
            <Pagination
              current={currentPage}
              total={total}
              pageSize={pageSize}
              showSizeChanger
              showQuickJumper
              showTotal={(total, range) =>
                `第 ${range[0]}-${range[1]} 条/共 ${total} 条`
              }
              onChange={handlePageChange}
            />
          </div>
        </>
      )}
    </div>
  );
};

export default Category;