import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Row,
  Col,
  Card,
  Typography,
  Space,
  Spin,
  message,
  Tag,
  Avatar,
  Statistic,
} from 'antd';
import {
  AppstoreOutlined,
  FolderOutlined,
  RightOutlined,
  CodeOutlined,
  ToolOutlined,
  BgColorsOutlined,
  PlayCircleOutlined,
  SettingOutlined,
  DatabaseOutlined,
  SecurityScanOutlined,
  ApiOutlined,
} from '@ant-design/icons';
import { Category } from '@/types';
import { pluginService } from '@/services/plugin';

const { Title, Paragraph } = Typography;

const Categories: React.FC = () => {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(true);
  const [categories, setCategories] = useState<Category[]>([]);

  // 分类图标映射
  const categoryIcons: Record<string, React.ReactNode> = {
    development: <CodeOutlined style={{ fontSize: 32, color: '#1890ff' }} />,
    productivity: <ToolOutlined style={{ fontSize: 32, color: '#52c41a' }} />,
    design: <BgColorsOutlined style={{ fontSize: 32, color: '#eb2f96' }} />,
    entertainment: <PlayCircleOutlined style={{ fontSize: 32, color: '#fa8c16' }} />,
    utility: <SettingOutlined style={{ fontSize: 32, color: '#722ed1' }} />,
    database: <DatabaseOutlined style={{ fontSize: 32, color: '#13c2c2' }} />,
    security: <SecurityScanOutlined style={{ fontSize: 32, color: '#f5222d' }} />,
    api: <ApiOutlined style={{ fontSize: 32, color: '#faad14' }} />,
  };

  // 模拟分类数据
  const mockCategories: Category[] = [
    {
      id: 'development',
      name: '开发工具',
      description: '代码编辑器、调试工具、版本控制等开发相关插件',
      icon: 'development',
      pluginCount: 156,
    },
    {
      id: 'productivity',
      name: '效率工具',
      description: '提升工作效率的各类工具和插件',
      icon: 'productivity',
      pluginCount: 89,
    },
    {
      id: 'design',
      name: '设计工具',
      description: '图形设计、UI/UX设计相关的插件和工具',
      icon: 'design',
      pluginCount: 67,
    },
    {
      id: 'entertainment',
      name: '娱乐工具',
      description: '游戏、音乐、视频等娱乐相关插件',
      icon: 'entertainment',
      pluginCount: 45,
    },
    {
      id: 'utility',
      name: '实用工具',
      description: '系统工具、文件管理、网络工具等实用插件',
      icon: 'utility',
      pluginCount: 78,
    },
    {
      id: 'database',
      name: '数据库工具',
      description: '数据库管理、数据分析相关插件',
      icon: 'database',
      pluginCount: 34,
    },
    {
      id: 'security',
      name: '安全工具',
      description: '安全扫描、加密解密、权限管理等安全插件',
      icon: 'security',
      pluginCount: 23,
    },
    {
      id: 'api',
      name: 'API工具',
      description: 'API测试、接口文档、数据交互相关插件',
      icon: 'api',
      pluginCount: 41,
    },
  ];

  useEffect(() => {
    loadCategories();
  }, []);

  const loadCategories = async () => {
    try {
      setLoading(true);
      // 模拟 API 调用
      await new Promise(resolve => setTimeout(resolve, 800));
      setCategories(mockCategories);
    } catch (error) {
      console.error('加载分类失败:', error);
      message.error('加载分类失败，请刷新页面重试');
    } finally {
      setLoading(false);
    }
  };

  const handleCategoryClick = (categoryId: string) => {
    navigate(`/category/${categoryId}`);
  };

  const totalPlugins = categories.reduce((sum, cat) => sum + cat.pluginCount, 0);

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
      {/* 页面头部 */}
      <div style={{ textAlign: 'center', marginBottom: '48px' }}>
        <Title level={1}>
          <AppstoreOutlined /> 插件分类
        </Title>
        <Paragraph style={{ fontSize: '16px', color: '#666' }}>
          浏览所有插件分类，找到适合您需求的插件
        </Paragraph>
        
        {/* 统计信息 */}
        <Row gutter={16} style={{ marginTop: '32px' }}>
          <Col span={8}>
            <Statistic
              title="总分类数"
              value={categories.length}
              prefix={<FolderOutlined />}
            />
          </Col>
          <Col span={8}>
            <Statistic
              title="总插件数"
              value={totalPlugins}
              prefix={<AppstoreOutlined />}
            />
          </Col>
          <Col span={8}>
            <Statistic
              title="平均每类"
              value={Math.round(totalPlugins / categories.length)}
              suffix="个插件"
            />
          </Col>
        </Row>
      </div>

      {/* 分类网格 */}
      <Row gutter={[24, 24]}>
        {categories.map((category) => (
          <Col xs={24} sm={12} md={8} lg={6} key={category.id}>
            <Card
              hoverable
              style={{ height: '100%' }}
              onClick={() => handleCategoryClick(category.id)}
              bodyStyle={{ 
                padding: '24px',
                display: 'flex',
                flexDirection: 'column',
                alignItems: 'center',
                textAlign: 'center',
                height: '100%',
              }}
            >
              <div style={{ marginBottom: '16px' }}>
                {categoryIcons[category.icon] || (
                  <FolderOutlined style={{ fontSize: 32, color: '#1890ff' }} />
                )}
              </div>
              
              <Title level={4} style={{ marginBottom: '8px' }}>
                {category.name}
              </Title>
              
              <Paragraph 
                style={{ 
                  color: '#666', 
                  fontSize: '14px',
                  marginBottom: '16px',
                  flex: 1,
                  display: 'flex',
                  alignItems: 'center',
                }}
              >
                {category.description}
              </Paragraph>
              
              <Space direction="vertical" align="center">
                <Tag color="blue" style={{ margin: 0 }}>
                  {category.pluginCount} 个插件
                </Tag>
                
                <div style={{ 
                  display: 'flex', 
                  alignItems: 'center', 
                  color: '#1890ff',
                  fontSize: '14px',
                  marginTop: '8px',
                }}>
                  浏览插件 <RightOutlined style={{ marginLeft: '4px' }} />
                </div>
              </Space>
            </Card>
          </Col>
        ))}
      </Row>

      {/* 底部提示 */}
      <div style={{ 
        textAlign: 'center', 
        marginTop: '48px',
        padding: '24px',
        background: '#f5f5f5',
        borderRadius: '8px',
      }}>
        <Title level={4}>找不到合适的分类？</Title>
        <Paragraph style={{ marginBottom: '16px' }}>
          您可以使用搜索功能查找特定的插件，或者联系我们建议新的分类。
        </Paragraph>
        <Space>
          <Card
            size="small"
            style={{ cursor: 'pointer' }}
            onClick={() => navigate('/search')}
          >
            <Space>
              <AppstoreOutlined />
              搜索插件
            </Space>
          </Card>
        </Space>
      </div>
    </div>
  );
};

export default Categories;