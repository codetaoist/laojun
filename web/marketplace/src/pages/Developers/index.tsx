import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Row,
  Col,
  Card,
  Typography,
  Space,
  Button,
  Avatar,
  Tag,
  Input,
  Select,
  Spin,
  message,
  Pagination,
  Empty,
  Breadcrumb,
  Statistic,
} from 'antd';
import {
  UserOutlined,
  SearchOutlined,
  TeamOutlined,
  TrophyOutlined,
  AppstoreOutlined,
  DownloadOutlined,
  StarOutlined,
  HomeOutlined,
  FilterOutlined,
} from '@ant-design/icons';
import { Developer } from '@/types';

const { Title, Text } = Typography;
const { Option } = Select;

const Developers: React.FC = () => {
  const navigate = useNavigate();
  
  const [loading, setLoading] = useState(true);
  const [developers, setDevelopers] = useState<Developer[]>([]);
  const [filteredDevelopers, setFilteredDevelopers] = useState<Developer[]>([]);
  const [searchText, setSearchText] = useState('');
  const [sortBy, setSortBy] = useState<string>('downloads');
  const [filterBy, setFilterBy] = useState<string>('all');
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize] = useState(12);

  // 模拟开发者数据
  const mockDevelopers: Developer[] = [
    {
      id: 'devtools-inc',
      name: 'DevTools Inc.',
      email: 'contact@devtools.com',
      website: 'https://devtools.com',
      avatar: 'https://via.placeholder.com/128',
      bio: '专注于开发工具和效率软件的创新公司',
      location: '美国 加利福尼亚州',
      joinDate: '2020-01-15',
      verified: true,
      pluginCount: 12,
      totalDownloads: 156420,
      averageRating: 4.7,
      followers: 2340,
    },
    {
      id: 'code-masters',
      name: 'Code Masters',
      email: 'hello@codemasters.dev',
      website: 'https://codemasters.dev',
      avatar: 'https://via.placeholder.com/128',
      bio: '代码大师团队，致力于提供最优质的编程工具',
      location: '中国 北京',
      joinDate: '2019-08-20',
      verified: true,
      pluginCount: 8,
      totalDownloads: 89320,
      averageRating: 4.5,
      followers: 1890,
    },
    {
      id: 'ui-wizards',
      name: 'UI Wizards',
      email: 'team@uiwizards.com',
      website: 'https://uiwizards.com',
      avatar: 'https://via.placeholder.com/128',
      bio: 'UI设计专家，专注于用户界面和用户体验',
      location: '英国 伦敦',
      joinDate: '2021-03-10',
      verified: false,
      pluginCount: 15,
      totalDownloads: 234560,
      averageRating: 4.8,
      followers: 3120,
    },
    {
      id: 'data-scientists',
      name: 'Data Scientists',
      email: 'info@datascientists.org',
      website: 'https://datascientists.org',
      avatar: 'https://via.placeholder.com/128',
      bio: '数据科学家团队，专注于数据分析和机器学习工具',
      location: '德国 柏林',
      joinDate: '2020-11-05',
      verified: true,
      pluginCount: 6,
      totalDownloads: 67890,
      averageRating: 4.6,
      followers: 1560,
    },
    {
      id: 'security-experts',
      name: 'Security Experts',
      email: 'contact@securityexperts.io',
      website: 'https://securityexperts.io',
      avatar: 'https://via.placeholder.com/128',
      bio: '网络安全专家，提供安全相关的开发工具',
      location: '以色列 特拉维夫',
      joinDate: '2018-06-12',
      verified: true,
      pluginCount: 9,
      totalDownloads: 123450,
      averageRating: 4.4,
      followers: 2100,
    },
    {
      id: 'mobile-devs',
      name: 'Mobile Developers',
      email: 'hello@mobiledevs.app',
      website: 'https://mobiledevs.app',
      avatar: 'https://via.placeholder.com/128',
      bio: '移动应用开发专家，专注于跨平台开发工具',
      location: '加拿大 多伦多',
      joinDate: '2021-09-18',
      verified: false,
      pluginCount: 11,
      totalDownloads: 98760,
      averageRating: 4.3,
      followers: 1780,
    },
  ];

  useEffect(() => {
    loadDevelopers();
  }, []);

  useEffect(() => {
    filterAndSortDevelopers();
  }, [developers, searchText, sortBy, filterBy]);

  const loadDevelopers = async () => {
    try {
      setLoading(true);
      
      // 模拟加载数据
      await new Promise(resolve => setTimeout(resolve, 800));
      setDevelopers(mockDevelopers);
    } catch (error) {
      console.error('加载开发者数据失败:', error);
      message.error('加载数据失败，请刷新页面重试');
    } finally {
      setLoading(false);
    }
  };

  const filterAndSortDevelopers = () => {
    let filtered = [...developers];

    // 搜索过滤
    if (searchText) {
      filtered = filtered.filter(dev =>
        dev.name.toLowerCase().includes(searchText.toLowerCase()) ||
        dev.bio.toLowerCase().includes(searchText.toLowerCase())
      );
    }

    // 认证状态过滤
    if (filterBy === 'verified') {
      filtered = filtered.filter(dev => dev.verified);
    } else if (filterBy === 'unverified') {
      filtered = filtered.filter(dev => !dev.verified);
    }

    // 排序
    filtered.sort((a, b) => {
      switch (sortBy) {
        case 'downloads':
          return b.totalDownloads - a.totalDownloads;
        case 'rating':
          return b.averageRating - a.averageRating;
        case 'plugins':
          return b.pluginCount - a.pluginCount;
        case 'followers':
          return b.followers - a.followers;
        case 'name':
          return a.name.localeCompare(b.name);
        case 'joinDate':
          return new Date(b.joinDate).getTime() - new Date(a.joinDate).getTime();
        default:
          return 0;
      }
    });

    setFilteredDevelopers(filtered);
    setCurrentPage(1);
  };

  const handleDeveloperClick = (developerId: string) => {
    navigate(`/developer/${developerId}`);
  };

  const handleFollow = (developerId: string, e: React.MouseEvent) => {
    e.stopPropagation();
    message.success('已关注该开发者');
  };

  const getCurrentPageData = () => {
    const startIndex = (currentPage - 1) * pageSize;
    const endIndex = startIndex + pageSize;
    return filteredDevelopers.slice(startIndex, endIndex);
  };

  const totalStats = {
    totalDevelopers: developers.length,
    verifiedDevelopers: developers.filter(dev => dev.verified).length,
    totalPlugins: developers.reduce((sum, dev) => sum + dev.pluginCount, 0),
    totalDownloads: developers.reduce((sum, dev) => sum + dev.totalDownloads, 0),
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
          <TeamOutlined />
          开发者
        </Breadcrumb.Item>
      </Breadcrumb>

      {/* 页面标题和统计 */}
      <div style={{ marginBottom: '24px' }}>
        <Title level={2} style={{ marginBottom: '16px' }}>
          <TeamOutlined style={{ marginRight: '8px' }} />
          开发者
        </Title>
        
        <Row gutter={16} style={{ marginBottom: '24px' }}>
          <Col xs={12} sm={6}>
            <Card>
              <Statistic
                title="开发者总数"
                value={totalStats.totalDevelopers}
                prefix={<UserOutlined />}
              />
            </Card>
          </Col>
          <Col xs={12} sm={6}>
            <Card>
              <Statistic
                title="认证开发者"
                value={totalStats.verifiedDevelopers}
                prefix={<TrophyOutlined />}
              />
            </Card>
          </Col>
          <Col xs={12} sm={6}>
            <Card>
              <Statistic
                title="插件总数"
                value={totalStats.totalPlugins}
                prefix={<AppstoreOutlined />}
              />
            </Card>
          </Col>
          <Col xs={12} sm={6}>
            <Card>
              <Statistic
                title="总下载量"
                value={totalStats.totalDownloads}
                prefix={<DownloadOutlined />}
              />
            </Card>
          </Col>
        </Row>
      </div>

      {/* 搜索和过滤 */}
      <Card style={{ marginBottom: '24px' }}>
        <Row gutter={16} align="middle">
          <Col xs={24} sm={12} md={8}>
            <Input
              placeholder="搜索开发者..."
              prefix={<SearchOutlined />}
              value={searchText}
              onChange={(e) => setSearchText(e.target.value)}
              allowClear
            />
          </Col>
          <Col xs={12} sm={6} md={4}>
            <Select
              style={{ width: '100%' }}
              placeholder="认证状态"
              value={filterBy}
              onChange={setFilterBy}
              suffixIcon={<FilterOutlined />}
            >
              <Option value="all">全部</Option>
              <Option value="verified">已认证</Option>
              <Option value="unverified">未认证</Option>
            </Select>
          </Col>
          <Col xs={12} sm={6} md={4}>
            <Select
              style={{ width: '100%' }}
              placeholder="排序方式"
              value={sortBy}
              onChange={setSortBy}
            >
              <Option value="downloads">下载量</Option>
              <Option value="rating">评分</Option>
              <Option value="plugins">插件数量</Option>
              <Option value="followers">关注者</Option>
              <Option value="name">名称</Option>
              <Option value="joinDate">加入时间</Option>
            </Select>
          </Col>
          <Col xs={24} sm={24} md={8}>
            <Text style={{ color: '#666' }}>
              找到 {filteredDevelopers.length} 个开发者
            </Text>
          </Col>
        </Row>
      </Card>

      {/* 开发者列表 */}
      {filteredDevelopers.length === 0 ? (
        <Empty description="没有找到匹配的开发者" />
      ) : (
        <>
          <Row gutter={[24, 24]}>
            {getCurrentPageData().map((developer) => (
              <Col xs={24} sm={12} md={8} lg={6} key={developer.id}>
                <Card
                  hoverable
                  style={{ height: '100%' }}
                  onClick={() => handleDeveloperClick(developer.id)}
                  actions={[
                    <Button
                      type="text"
                      onClick={(e) => handleFollow(developer.id, e)}
                    >
                      关注
                    </Button>,
                    <Button
                      type="text"
                      onClick={(e) => {
                        e.stopPropagation();
                        handleDeveloperClick(developer.id);
                      }}
                    >
                      查看详情
                    </Button>,
                  ]}
                >
                  <div style={{ textAlign: 'center', marginBottom: '16px' }}>
                    <Avatar
                      size={64}
                      src={developer.avatar}
                      icon={<UserOutlined />}
                      style={{ marginBottom: '8px' }}
                    />
                    <div>
                      <Text strong style={{ fontSize: '16px' }}>
                        {developer.name}
                      </Text>
                      {developer.verified && (
                        <Tag 
                          color="blue" 
                          size="small" 
                          style={{ marginLeft: '4px' }}
                          icon={<TrophyOutlined />}
                        >
                          认证
                        </Tag>
                      )}
                    </div>
                  </div>
                  
                  <div style={{ marginBottom: '16px' }}>
                    <Text style={{ fontSize: '12px', color: '#666', display: 'block' }}>
                      {developer.bio}
                    </Text>
                  </div>
                  
                  <Space direction="vertical" size="small" style={{ width: '100%' }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                      <Text style={{ fontSize: '12px' }}>插件数量</Text>
                      <Text strong style={{ fontSize: '12px' }}>
                        {developer.pluginCount}
                      </Text>
                    </div>
                    <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                      <Text style={{ fontSize: '12px' }}>总下载量</Text>
                      <Text strong style={{ fontSize: '12px' }}>
                        {developer.totalDownloads.toLocaleString()}
                      </Text>
                    </div>
                    <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                      <Text style={{ fontSize: '12px' }}>平均评分</Text>
                      <Space size="small">
                        <StarOutlined style={{ color: '#faad14', fontSize: '12px' }} />
                        <Text strong style={{ fontSize: '12px' }}>
                          {developer.averageRating}
                        </Text>
                      </Space>
                    </div>
                    <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                      <Text style={{ fontSize: '12px' }}>关注者</Text>
                      <Text strong style={{ fontSize: '12px' }}>
                        {developer.followers.toLocaleString()}
                      </Text>
                    </div>
                  </Space>
                </Card>
              </Col>
            ))}
          </Row>

          {/* 分页 */}
          <div style={{ textAlign: 'center', marginTop: '32px' }}>
            <Pagination
              current={currentPage}
              pageSize={pageSize}
              total={filteredDevelopers.length}
              onChange={setCurrentPage}
              showSizeChanger={false}
              showQuickJumper
              showTotal={(total, range) =>
                `第 ${range[0]}-${range[1]} 项，共 ${total} 项`
              }
            />
          </div>
        </>
      )}
    </div>
  );
};

export default Developers;