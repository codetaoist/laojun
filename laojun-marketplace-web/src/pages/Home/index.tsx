import React, { useEffect, useState } from 'react';
import {
  Row,
  Col,
  Card,
  Button,
  Typography,
  Space,
  Carousel,
  Statistic,
  Tag,
  Rate,
  Avatar,
  Skeleton,
  App,
  message,
} from 'antd';
import {
  DownloadOutlined,
  StarOutlined,
  EyeOutlined,
  RightOutlined,
  AppstoreOutlined,
  TeamOutlined,
  TrophyOutlined,
} from '@ant-design/icons';
import { Link, useNavigate } from 'react-router-dom';
import { Plugin, Category } from '@/types';
import { pluginService } from '@/services/plugin';
import { useCartStore } from '@/stores/cart';
import './index.css';

const { Title, Paragraph, Text } = Typography;
const { Meta } = Card;

const Home: React.FC = () => {
  const navigate = useNavigate();
  const { addToCart, isInCart } = useCartStore();
  const { message } = App.useApp();
  
  const [loading, setLoading] = useState(true);
  const [featuredPlugins, setFeaturedPlugins] = useState<Plugin[]>([]);
  const [latestPlugins, setLatestPlugins] = useState<Plugin[]>([]);
  const [popularCategories, setPopularCategories] = useState<Category[]>([]);
  const [stats, setStats] = useState({
    totalPlugins: 0,
    totalDownloads: 0,
    totalDevelopers: 0,
    averageRating: 0,
  });

  useEffect(() => {
    loadHomeData();
  }, []);

  const loadHomeData = async () => {
    try {
      setLoading(true);
      
      // 并行加载数据
      const [featured, latest, categories] = await Promise.all([
        pluginService.getFeaturedPlugins(8),
        pluginService.getLatestPlugins(8),
        pluginService.getCategories(),
      ]);

      setFeaturedPlugins(featured);
      setLatestPlugins(latest);
      setPopularCategories(categories.slice(0, 6));
      
      // 模拟统计数据
      setStats({
        totalPlugins: 1234,
        totalDownloads: 56789,
        totalDevelopers: 123,
        averageRating: 4.6,
      });
    } catch (error) {
      console.error('加载首页数据失败:', error);
      message.error('加载数据失败，请刷新页面重试');
    } finally {
      setLoading(false);
    }
  };

  const handleAddToCart = (plugin: Plugin) => {
    addToCart(plugin);
  };

  const handleInstallPlugin = (plugin: Plugin) => {
    if (plugin.price === 0) {
      // 免费插件直接安装
      message.success(`正在安装 ${plugin.name}...`);
      // 这里调用安装 API
    } else {
      // 付费插件加入购物车
      handleAddToCart(plugin);
    }
  };

  // 轮播图数据
  const bannerData = [
    {
      title: '发现优质插件',
      description: '探索丰富的插件生态，扩展您的系统功能',
      image: '/images/banner1.jpg',
      action: { text: '浏览插件', link: '/categories' },
    },
    {
      title: '开发者社区',
      description: '加入开发者社区，分享您的创意和技能',
      image: '/images/banner2.jpg',
      action: { text: '了解更多', link: '/developers' },
    },
    {
      title: '安全可靠',
      description: '所有插件经过严格审核，确保安全性和质量',
      image: '/images/banner3.jpg',
      action: { text: '查看详情', link: '/about' },
    },
  ];

  return (
    <div className="home-page">
      {/* 轮播横幅 */}
      <section className="hero-section">
        <Carousel autoplay effect="fade" className="hero-carousel">
          {bannerData.map((banner, index) => (
            <div key={index} className="hero-slide">
              <div className="hero-content">
                <div className="hero-text">
                  <Title level={1} className="hero-title">
                    {banner.title}
                  </Title>
                  <Paragraph className="hero-description">
                    {banner.description}
                  </Paragraph>
                  <Button
                    type="primary"
                    size="large"
                    onClick={() => navigate(banner.action.link)}
                    className="hero-button"
                  >
                    {banner.action.text}
                    <RightOutlined />
                  </Button>
                </div>
                <div className="hero-image">
                  <img src={banner.image} alt={banner.title} />
                </div>
              </div>
            </div>
          ))}
        </Carousel>
      </section>

      {/* 统计数据 */}
      <section className="stats-section">
        <Row gutter={[24, 24]}>
          <Col xs={12} sm={6}>
            <Card className="stat-card">
              <Statistic
                title="插件总数"
                value={stats.totalPlugins}
                prefix={<AppstoreOutlined />}
                valueStyle={{ color: '#1890ff' }}
              />
            </Card>
          </Col>
          <Col xs={12} sm={6}>
            <Card className="stat-card">
              <Statistic
                title="总下载量"
                value={stats.totalDownloads}
                prefix={<DownloadOutlined />}
                valueStyle={{ color: '#52c41a' }}
              />
            </Card>
          </Col>
          <Col xs={12} sm={6}>
            <Card className="stat-card">
              <Statistic
                title="开发者"
                value={stats.totalDevelopers}
                prefix={<TeamOutlined />}
                valueStyle={{ color: '#722ed1' }}
              />
            </Card>
          </Col>
          <Col xs={12} sm={6}>
            <Card className="stat-card">
              <Statistic
                title="平均评分"
                value={stats.averageRating}
                precision={1}
                prefix={<TrophyOutlined />}
                valueStyle={{ color: '#fa8c16' }}
              />
            </Card>
          </Col>
        </Row>
      </section>

      {/* 热门分类 */}
      <section className="categories-section">
        <div className="section-header">
          <Title level={2}>热门分类</Title>
          <Link to="/categories">
            查看全部 <RightOutlined />
          </Link>
        </div>
        
        <Row gutter={[16, 16]}>
          {loading ? (
            Array.from({ length: 6 }).map((_, index) => (
              <Col key={index} xs={12} sm={8} md={6} lg={4}>
                <Card loading />
              </Col>
            ))
          ) : (
            popularCategories.map((category) => (
              <Col key={category.id} xs={12} sm={8} md={6} lg={4}>
                <Card
                  hoverable
                  className="category-card"
                  onClick={() => navigate(`/category/${category.id}`)}
                >
                  <div className="category-icon">
                    {category.icon ? (
                      <img src={category.icon} alt={category.name} />
                    ) : (
                      <AppstoreOutlined />
                    )}
                  </div>
                  <div className="category-info">
                    <Text strong>{category.name}</Text>
                    <Text type="secondary" className="category-count">
                      {category.pluginCount} 个插件
                    </Text>
                  </div>
                </Card>
              </Col>
            ))
          )}
        </Row>
      </section>

      {/* 精选插件 */}
      <section className="featured-section">
        <div className="section-header">
          <Title level={2}>精选插件</Title>
          <Link to="/search?featured=true">
            查看全部 <RightOutlined />
          </Link>
        </div>
        
        <Row gutter={[16, 16]}>
          {loading ? (
            Array.from({ length: 8 }).map((_, index) => (
              <Col key={index} xs={24} sm={12} md={8} lg={6}>
                <Card loading />
              </Col>
            ))
          ) : (
            featuredPlugins.map((plugin) => (
              <Col key={plugin.id} xs={24} sm={12} md={8} lg={6}>
                <Card
                  hoverable
                  className="plugin-card"
                  cover={
                    plugin.icon ? (
                      <img alt={plugin.name} src={plugin.icon} />
                    ) : (
                      <div className="plugin-placeholder">
                        <AppstoreOutlined />
                      </div>
                    )
                  }
                  actions={[
                    <Button
                      key="view"
                      type="text"
                      icon={<EyeOutlined />}
                      onClick={() => navigate(`/plugin/${plugin.id}`)}
                    >
                      查看
                    </Button>,
                    <Button
                      key="install"
                      type="primary"
                      size="small"
                      disabled={plugin.price > 0 && isInCart(plugin.id)}
                      onClick={() => handleInstallPlugin(plugin)}
                    >
                      {plugin.price === 0 ? '安装' : isInCart(plugin.id) ? '已在购物车' : '加入购物车'}
                    </Button>,
                  ]}
                >
                  <Meta
                    title={
                      <div className="plugin-title">
                        <Text strong ellipsis={{ tooltip: plugin.name }}>
                          {plugin.name}
                        </Text>
                        {plugin.price > 0 && (
                          <Text type="danger" className="plugin-price">
                            ¥{plugin.price}
                          </Text>
                        )}
                      </div>
                    }
                    description={
                      <div className="plugin-meta">
                        <Paragraph
                          ellipsis={{ rows: 2, tooltip: plugin.description }}
                          className="plugin-description"
                        >
                          {plugin.description}
                        </Paragraph>
                        <div className="plugin-stats">
                          <Space size="small">
                            <Rate disabled defaultValue={plugin.rating || 0} allowHalf />
                            <Text type="secondary">({plugin.reviewCount || 0})</Text>
                          </Space>
                          <Text type="secondary">
                            {(plugin.downloads || 0).toLocaleString()} 下载
                          </Text>
                        </div>
                        <div className="plugin-tags">
                          {(plugin.tags || []).slice(0, 2).map((tag) => (
                            <Tag key={tag} size="small">
                              {tag}
                            </Tag>
                          ))}
                        </div>
                      </div>
                    }
                  />
                </Card>
              </Col>
            ))
          )}
        </Row>
      </section>

      {/* 最新插件 */}
      <section className="latest-section">
        <div className="section-header">
          <Title level={2}>最新插件</Title>
          <Link to="/search?sort=created&order=desc">
            查看全部 <RightOutlined />
          </Link>
        </div>
        
        <Row gutter={[16, 16]}>
          {loading ? (
            Array.from({ length: 8 }).map((_, index) => (
              <Col key={index} xs={24} sm={12} md={8} lg={6}>
                <Skeleton loading active avatar />
              </Col>
            ))
          ) : (
            latestPlugins.map((plugin) => (
              <Col key={plugin.id} xs={24} sm={12} md={8} lg={6}>
                <Card
                  hoverable
                  className="latest-plugin-card"
                  onClick={() => navigate(`/plugin/${plugin.id}`)}
                >
                  <Meta
                    avatar={
                      <Avatar
                        size={48}
                        src={plugin.icon}
                        icon={<AppstoreOutlined />}
                      />
                    }
                    title={plugin.name}
                    description={
                      <div>
                        <Paragraph
                          ellipsis={{ rows: 2 }}
                          className="latest-plugin-description"
                        >
                          {plugin.description}
                        </Paragraph>
                        <div className="latest-plugin-info">
                          <Text type="secondary">by {plugin.author}</Text>
                          <Text type="secondary">
                            {plugin.createdAt ? new Date(plugin.createdAt).toLocaleDateString() : '未知日期'}
                          </Text>
                        </div>
                      </div>
                    }
                  />
                </Card>
              </Col>
            ))
          )}
        </Row>
      </section>
    </div>
  );
};

export default Home;