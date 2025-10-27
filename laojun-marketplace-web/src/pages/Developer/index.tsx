import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import {
  Row,
  Col,
  Card,
  Typography,
  Space,
  Button,
  Avatar,
  Tag,
  Rate,
  Spin,
  message,
  Tabs,
  List,
  Statistic,
  Breadcrumb,
  Empty,
  Divider,
} from 'antd';
import './Developer.css';
import {
  UserOutlined,
  MailOutlined,
  GlobalOutlined,
  AppstoreOutlined,
  DownloadOutlined,
  StarOutlined,
  EyeOutlined,
  ShoppingCartOutlined,
  HomeOutlined,
  TeamOutlined,
  CalendarOutlined,
  TrophyOutlined,
} from '@ant-design/icons';
import { Plugin, Developer as DeveloperType } from '@/types';
import { pluginService } from '@/services/plugin';
import { useCartStore } from '@/stores/cart';

const { Title, Text, Paragraph } = Typography;
const { TabPane } = Tabs;

const Developer: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { addToCart, isInCart } = useCartStore();
  
  const [loading, setLoading] = useState(true);
  const [developer, setDeveloper] = useState<DeveloperType | null>(null);
  const [plugins, setPlugins] = useState<Plugin[]>([]);
  const [activeTab, setActiveTab] = useState('plugins');

  // 模拟开发者数据
  const mockDeveloper: DeveloperType = {
    id: 'devtools-inc',
    name: 'DevTools Inc.',
    email: 'contact@devtools.com',
    website: 'https://devtools.com',
    avatar: 'https://via.placeholder.com/128',
    bio: '专注于开发工具和效率软件的创新公司，致力于为开发者提供最优质的工具和服务。',
    location: '美国 加利福尼亚州',
    joinDate: '2020-01-15',
    verified: true,
    pluginCount: 12,
    totalDownloads: 156420,
    averageRating: 4.7,
    followers: 2340,
    socialLinks: {
      github: 'https://github.com/devtools-inc',
      twitter: 'https://twitter.com/devtools_inc',
      linkedin: 'https://linkedin.com/company/devtools-inc',
    },
  };

  // 模拟开发者的插件
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
      name: 'Debug Assistant',
      description: '智能调试助手，帮助快速定位和解决代码问题',
      version: '1.3.2',
      author: 'DevTools Inc.',
      authorEmail: 'contact@devtools.com',
      category: 'development',
      tags: ['debug', 'assistant', 'analysis'],
      icon: 'https://via.placeholder.com/64',
      screenshots: [],
      downloadUrl: '',
      license: 'Commercial',
      price: 39.99,
      currency: 'USD',
      downloads: 8932,
      rating: 4.6,
      reviewCount: 156,
      size: 1536000,
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

  useEffect(() => {
    if (id) {
      loadDeveloperData();
    }
  }, [id]);

  const loadDeveloperData = async () => {
    try {
      setLoading(true);
      
      // 模拟加载开发者数据
      await new Promise(resolve => setTimeout(resolve, 800));
      setDeveloper(mockDeveloper);
      setPlugins(mockPlugins);
    } catch (error) {
      console.error('加载开发者数据失败:', error);
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

  const handleFollow = () => {
    message.success('已关注该开发者');
  };

  if (loading) {
    return (
      <div className="developer-loading">
        <Spin size="large" />
      </div>
    );
  }

  if (!developer) {
    return (
      <div className="developer-empty">
        <Empty description="开发者不存在" />
      </div>
    );
  }

  return (
    <div className="developer-page">
      {/* 面包屑导航 */}
      <Breadcrumb className="developer-breadcrumb">
        <Breadcrumb.Item>
          <HomeOutlined />
          <span onClick={() => navigate('/')} style={{ cursor: 'pointer' }}>
            首页
          </span>
        </Breadcrumb.Item>
        <Breadcrumb.Item>
          <TeamOutlined />
          <span onClick={() => navigate('/developers')} style={{ cursor: 'pointer' }}>
            开发者
          </span>
        </Breadcrumb.Item>
        <Breadcrumb.Item>
          <UserOutlined />
          {developer.name}
        </Breadcrumb.Item>
      </Breadcrumb>

      {/* 开发者信息卡片 */}
      <Card className="developer-info-card">
        <Row gutter={24}>
          <Col xs={24} md={6}>
            <div className="developer-avatar-section">
              <Avatar
                size={128}
                src={developer.avatar}
                icon={<UserOutlined />}
                className="developer-avatar"
              />
              {developer.verified && (
                <div className="developer-verified-badge">
                  <Tag color="blue" icon={<TrophyOutlined />}>
                    认证开发者
                  </Tag>
                </div>
              )}
              <Button 
                type="primary" 
                block 
                onClick={handleFollow}
                className="developer-follow-btn"
              >
                关注开发者
              </Button>
            </div>
          </Col>
          
          <Col xs={24} md={18}>
            <div className="developer-details">
              <Title level={2} className="developer-name">
                {developer.name}
              </Title>
              
              <Space direction="vertical" size="small" className="developer-contact-info">
                <Space>
                  <MailOutlined />
                  <Text>{developer.email}</Text>
                </Space>
                {developer.website && (
                  <Space>
                    <GlobalOutlined />
                    <a href={developer.website} target="_blank" rel="noopener noreferrer">
                      {developer.website}
                    </a>
                  </Space>
                )}
                <Space>
                  <CalendarOutlined />
                  <Text>加入时间: {new Date(developer.joinDate).toLocaleDateString()}</Text>
                </Space>
              </Space>
              
              <Paragraph className="developer-bio">
                {developer.bio}
              </Paragraph>
              
              {/* 统计数据 */}
              <div className="developer-stats">
                <Row gutter={16}>
                  <Col span={6}>
                    <Statistic
                      title="插件数量"
                      value={developer.pluginCount}
                      prefix={<AppstoreOutlined />}
                    />
                  </Col>
                  <Col span={6}>
                    <Statistic
                      title="总下载量"
                      value={developer.totalDownloads}
                      prefix={<DownloadOutlined />}
                    />
                  </Col>
                  <Col span={6}>
                    <Statistic
                      title="平均评分"
                      value={developer.averageRating}
                      precision={1}
                      prefix={<StarOutlined />}
                    />
                  </Col>
                  <Col span={6}>
                    <Statistic
                      title="关注者"
                      value={developer.followers}
                      prefix={<UserOutlined />}
                    />
                  </Col>
                </Row>
              </div>
            </div>
          </Col>
        </Row>
      </Card>

      {/* 标签页内容 */}
      <Tabs activeKey={activeTab} onChange={setActiveTab} className="developer-tabs">
        <TabPane tab={`插件 (${plugins.length})`} key="plugins">
          {plugins.length === 0 ? (
            <Empty description="该开发者暂无插件" />
          ) : (
            <Row gutter={[24, 24]} className="developer-plugins-grid">
              {plugins.map((plugin) => (
                <Col xs={24} sm={12} md={8} lg={6} key={plugin.id}>
                  <Card
                    hoverable
                    className="developer-plugin-card"
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
                        <div className="plugin-card-header">
                          <Text strong className="plugin-card-title">
                            {plugin.name}
                          </Text>
                          {plugin.featured && (
                            <Tag className="plugin-featured-tag" size="small">推荐</Tag>
                          )}
                        </div>
                      }
                      description={
                        <div className="plugin-card-meta">
                          <Paragraph 
                            ellipsis={{ rows: 2 }} 
                            className="plugin-card-description"
                          >
                            {plugin.description}
                          </Paragraph>
                          
                          <Space direction="vertical" size="small" style={{ width: '100%' }}>
                            <div className="plugin-rating-section">
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
                              <Text className="plugin-downloads">
                                {plugin.downloads.toLocaleString()} 下载
                              </Text>
                            </div>
                            
                            <div className="plugin-version-price">
                              <Text className="plugin-version">
                                v{plugin.version}
                              </Text>
                              <Text 
                                strong 
                                className={`plugin-price ${plugin.price === 0 ? 'free' : 'paid'}`}
                              >
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
          )}
        </TabPane>
        
        <TabPane tab="关于" key="about">
          <Card className="developer-about-card">
            <Title level={4}>开发者简介</Title>
            <Paragraph>
              {developer.bio}
            </Paragraph>
            
            <Divider />
            
            <Title level={4}>联系方式</Title>
            <Space direction="vertical" size="middle" className="developer-contact-section">
              <Space>
                <MailOutlined />
                <Text>邮箱: {developer.email}</Text>
              </Space>
              {developer.website && (
                <Space>
                  <GlobalOutlined />
                  <Text>网站: </Text>
                  <a href={developer.website} target="_blank" rel="noopener noreferrer">
                    {developer.website}
                  </a>
                </Space>
              )}
              {developer.socialLinks && (
                <>
                  {developer.socialLinks.github && (
                    <Space>
                      <Text>GitHub: </Text>
                      <a href={developer.socialLinks.github} target="_blank" rel="noopener noreferrer">
                        {developer.socialLinks.github}
                      </a>
                    </Space>
                  )}
                  {developer.socialLinks.twitter && (
                    <Space>
                      <Text>Twitter: </Text>
                      <a href={developer.socialLinks.twitter} target="_blank" rel="noopener noreferrer">
                        {developer.socialLinks.twitter}
                      </a>
                    </Space>
                  )}
                </>
              )}
            </Space>
          </Card>
        </TabPane>
      </Tabs>
    </div>
  );
};

export default Developer;