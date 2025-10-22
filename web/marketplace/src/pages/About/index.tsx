import React from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Typography,
  Card,
  Row,
  Col,
  Space,
  Avatar,
  Divider,
  Timeline,
  Statistic,
  Button,
  Breadcrumb,
  Tag,
} from 'antd';
import {
  HomeOutlined,
  InfoCircleOutlined,
  TeamOutlined,
  RocketOutlined,
  TrophyOutlined,
  GlobalOutlined,
  MailOutlined,
  PhoneOutlined,
  EnvironmentOutlined,
  GithubOutlined,
  TwitterOutlined,
  LinkedinOutlined,
  HeartOutlined,
  StarOutlined,
  AppstoreOutlined,
  DownloadOutlined,
  UserOutlined,
} from '@ant-design/icons';

const { Title, Paragraph, Text } = Typography;

const About: React.FC = () => {
  const navigate = useNavigate();

  const teamMembers = [
    {
      name: '张三',
      role: '创始人 & CEO',
      avatar: 'https://via.placeholder.com/80',
      bio: '资深软件工程师，拥有15年的技术管理经验',
      skills: ['产品策略', '团队管理', '技术架构'],
    },
    {
      name: '李四',
      role: '技术总监',
      avatar: 'https://via.placeholder.com/80',
      bio: '全栈开发专家，专注于高性能系统设计',
      skills: ['系统架构', '性能优化', '云计算'],
    },
    {
      name: '王五',
      role: '产品经理',
      avatar: 'https://via.placeholder.com/80',
      bio: '用户体验专家，致力于打造优秀的产品体验',
      skills: ['产品设计', '用户研究', '数据分析'],
    },
    {
      name: '赵六',
      role: '设计总监',
      avatar: 'https://via.placeholder.com/80',
      bio: 'UI/UX设计师，专注于创造美观实用的界面',
      skills: ['UI设计', 'UX设计', '品牌设计'],
    },
  ];

  const milestones = [
    {
      date: '2024年1月',
      title: '平台正式上线',
      description: '插件市场正式对外开放，开始为开发者和用户提供服务',
      color: 'green',
    },
    {
      date: '2023年12月',
      title: '内测版本发布',
      description: '邀请核心用户参与内测，收集反馈并优化产品',
      color: 'blue',
    },
    {
      date: '2023年10月',
      title: '开发者计划启动',
      description: '启动开发者招募计划，建立插件开发生态',
      color: 'orange',
    },
    {
      date: '2023年8月',
      title: '项目立项',
      description: '确定产品方向，组建核心团队，开始产品开发',
      color: 'purple',
    },
  ];

  const stats = [
    {
      title: '注册用户',
      value: 50000,
      prefix: <UserOutlined />,
      suffix: '+',
    },
    {
      title: '插件数量',
      value: 1200,
      prefix: <AppstoreOutlined />,
      suffix: '+',
    },
    {
      title: '总下载量',
      value: 2500000,
      prefix: <DownloadOutlined />,
      suffix: '+',
    },
    {
      title: '开发者',
      value: 800,
      prefix: <TeamOutlined />,
      suffix: '+',
    },
  ];

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
          <InfoCircleOutlined />
          关于我们
        </Breadcrumb.Item>
      </Breadcrumb>

      {/* 页面标题 */}
      <div style={{ textAlign: 'center', marginBottom: '48px' }}>
        <Title level={1}>
          <InfoCircleOutlined style={{ marginRight: '12px', color: '#1890ff' }} />
          关于我们
        </Title>
        <Paragraph style={{ fontSize: '18px', color: '#666', maxWidth: '600px', margin: '0 auto' }}>
          我们致力于打造最优秀的插件市场平台，为开发者和用户搭建桥梁，推动软件生态的繁荣发展
        </Paragraph>
      </div>

      {/* 统计数据 */}
      <Row gutter={24} style={{ marginBottom: '48px' }}>
        {stats.map((stat, index) => (
          <Col xs={12} sm={6} key={index}>
            <Card>
              <Statistic
                title={stat.title}
                value={stat.value}
                prefix={stat.prefix}
                suffix={stat.suffix}
                valueStyle={{ color: '#1890ff' }}
              />
            </Card>
          </Col>
        ))}
      </Row>

      {/* 公司介绍 */}
      <Card style={{ marginBottom: '48px' }}>
        <Row gutter={48} align="middle">
          <Col xs={24} md={12}>
            <Title level={2}>
              <RocketOutlined style={{ marginRight: '8px', color: '#1890ff' }} />
              我们的使命
            </Title>
            <Paragraph style={{ fontSize: '16px', lineHeight: '1.8' }}>
              我们相信优秀的工具能够释放创造力，提升工作效率。我们的使命是为全球开发者提供一个
              开放、安全、高效的插件分发平台，让每个人都能轻松找到和分享优质的软件工具。
            </Paragraph>
            <Paragraph style={{ fontSize: '16px', lineHeight: '1.8' }}>
              通过我们的平台，开发者可以专注于创造，用户可以轻松获得所需的功能扩展，
              共同构建一个更加美好的数字世界。
            </Paragraph>
          </Col>
          <Col xs={24} md={12}>
            <div style={{ 
              background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
              borderRadius: '12px',
              padding: '40px',
              textAlign: 'center',
              color: 'white',
            }}>
              <RocketOutlined style={{ fontSize: '64px', marginBottom: '16px' }} />
              <Title level={3} style={{ color: 'white', marginBottom: '8px' }}>
                创新驱动
              </Title>
              <Text style={{ color: 'rgba(255,255,255,0.9)', fontSize: '16px' }}>
                持续创新，追求卓越
              </Text>
            </div>
          </Col>
        </Row>
      </Card>

      {/* 核心价值 */}
      <Card style={{ marginBottom: '48px' }}>
        <Title level={2} style={{ textAlign: 'center', marginBottom: '32px' }}>
          <HeartOutlined style={{ marginRight: '8px', color: '#ff4d4f' }} />
          核心价值
        </Title>
        <Row gutter={24}>
          <Col xs={24} sm={8}>
            <div style={{ textAlign: 'center', padding: '24px' }}>
              <StarOutlined style={{ fontSize: '48px', color: '#faad14', marginBottom: '16px' }} />
              <Title level={4}>品质至上</Title>
              <Paragraph style={{ color: '#666' }}>
                严格的审核标准，确保每个插件都具有高品质和安全性
              </Paragraph>
            </div>
          </Col>
          <Col xs={24} sm={8}>
            <div style={{ textAlign: 'center', padding: '24px' }}>
              <GlobalOutlined style={{ fontSize: '48px', color: '#52c41a', marginBottom: '16px' }} />
              <Title level={4}>开放共享</Title>
              <Paragraph style={{ color: '#666' }}>
                开放的生态系统，鼓励知识分享和技术交流
              </Paragraph>
            </div>
          </Col>
          <Col xs={24} sm={8}>
            <div style={{ textAlign: 'center', padding: '24px' }}>
              <TeamOutlined style={{ fontSize: '48px', color: '#1890ff', marginBottom: '16px' }} />
              <Title level={4}>用户第一</Title>
              <Paragraph style={{ color: '#666' }}>
                以用户需求为中心，持续改进产品体验
              </Paragraph>
            </div>
          </Col>
        </Row>
      </Card>

      {/* 发展历程 */}
      <Card style={{ marginBottom: '48px' }}>
        <Title level={2} style={{ marginBottom: '32px' }}>
          <TrophyOutlined style={{ marginRight: '8px', color: '#faad14' }} />
          发展历程
        </Title>
        <Timeline mode="left">
          {milestones.map((milestone, index) => (
            <Timeline.Item key={index} color={milestone.color}>
              <div>
                <Title level={4} style={{ marginBottom: '8px' }}>
                  {milestone.title}
                </Title>
                <Text style={{ color: '#666', fontSize: '14px', display: 'block', marginBottom: '8px' }}>
                  {milestone.date}
                </Text>
                <Paragraph style={{ marginBottom: 0 }}>
                  {milestone.description}
                </Paragraph>
              </div>
            </Timeline.Item>
          ))}
        </Timeline>
      </Card>

      {/* 团队介绍 */}
      <Card style={{ marginBottom: '48px' }}>
        <Title level={2} style={{ textAlign: 'center', marginBottom: '32px' }}>
          <TeamOutlined style={{ marginRight: '8px', color: '#1890ff' }} />
          核心团队
        </Title>
        <Row gutter={24}>
          {teamMembers.map((member, index) => (
            <Col xs={24} sm={12} md={6} key={index}>
              <Card hoverable style={{ textAlign: 'center', height: '100%' }}>
                <Avatar
                  size={80}
                  src={member.avatar}
                  icon={<UserOutlined />}
                  style={{ marginBottom: '16px' }}
                />
                <Title level={4} style={{ marginBottom: '8px' }}>
                  {member.name}
                </Title>
                <Text style={{ color: '#1890ff', fontSize: '14px', display: 'block', marginBottom: '12px' }}>
                  {member.role}
                </Text>
                <Paragraph style={{ fontSize: '12px', color: '#666', marginBottom: '16px' }}>
                  {member.bio}
                </Paragraph>
                <Space wrap>
                  {member.skills.map((skill, skillIndex) => (
                    <Tag key={skillIndex} color="blue">
                      {skill}
                    </Tag>
                  ))}
                </Space>
              </Card>
            </Col>
          ))}
        </Row>
      </Card>

      {/* 联系我们 */}
      <Card>
        <Title level={2} style={{ textAlign: 'center', marginBottom: '32px' }}>
          <MailOutlined style={{ marginRight: '8px', color: '#52c41a' }} />
          联系我们
        </Title>
        <Row gutter={48}>
          <Col xs={24} md={12}>
            <Space direction="vertical" size="large" style={{ width: '100%' }}>
              <div>
                <Title level={4}>联系方式</Title>
                <Space direction="vertical" size="middle">
                  <Space>
                    <MailOutlined style={{ color: '#1890ff' }} />
                    <Text>邮箱: contact@pluginmarket.com</Text>
                  </Space>
                  <Space>
                    <PhoneOutlined style={{ color: '#1890ff' }} />
                    <Text>电话: +86 400-123-4567</Text>
                  </Space>
                  <Space>
                    <EnvironmentOutlined style={{ color: '#1890ff' }} />
                    <Text>地址: 中国上海市浦东新区张江高科技园区</Text>
                  </Space>
                </Space>
              </div>
              
              <Divider />
              
              <div>
                <Title level={4}>关注我们</Title>
                <Space size="large">
                  <Button
                    type="text"
                    icon={<GithubOutlined />}
                    size="large"
                    href="https://github.com/pluginmarket"
                    target="_blank"
                  >
                    GitHub
                  </Button>
                  <Button
                    type="text"
                    icon={<TwitterOutlined />}
                    size="large"
                    href="https://twitter.com/pluginmarket"
                    target="_blank"
                  >
                    Twitter
                  </Button>
                  <Button
                    type="text"
                    icon={<LinkedinOutlined />}
                    size="large"
                    href="https://linkedin.com/company/pluginmarket"
                    target="_blank"
                  >
                    LinkedIn
                  </Button>
                </Space>
              </div>
            </Space>
          </Col>
          
          <Col xs={24} md={12}>
            <div style={{
              background: '#f5f5f5',
              borderRadius: '8px',
              padding: '32px',
              textAlign: 'center',
            }}>
              <Title level={4}>加入我们</Title>
              <Paragraph style={{ marginBottom: '24px' }}>
                我们正在寻找优秀的人才加入我们的团队，
                一起打造更好的插件市场平台。
              </Paragraph>
              <Button type="primary" size="large">
                查看职位
              </Button>
            </div>
          </Col>
        </Row>
      </Card>
    </div>
  );
};

export default About;