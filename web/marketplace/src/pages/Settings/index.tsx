import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuthStore } from '@/stores/auth';
import {
  Card,
  Typography,
  Form,
  Input,
  Button,
  Switch,
  Select,
  Radio,
  Divider,
  Space,
  message,
  Tabs,
  Upload,
  Avatar,
  Row,
  Col,
  Breadcrumb,
  Alert,
  Modal,
  List,
  Tag,
} from 'antd';
import {
  UserOutlined,
  SettingOutlined,
  BellOutlined,
  SecurityScanOutlined,
  GlobalOutlined,
  UploadOutlined,
  HomeOutlined,
  ExclamationCircleOutlined,
  DeleteOutlined,
  KeyOutlined,
  ShieldOutlined,
  EyeInvisibleOutlined,
  EyeOutlined,
} from '@ant-design/icons';

const { Title, Text, Paragraph } = Typography;
const { Option } = Select;
const { TabPane } = Tabs;
const { confirm } = Modal;

interface UserSettings {
  profile: {
    username: string;
    email: string;
    fullName: string;
    bio: string;
    avatar: string;
    website: string;
    location: string;
  };
  preferences: {
    language: string;
    theme: string;
    timezone: string;
    currency: string;
    autoUpdate: boolean;
    betaFeatures: boolean;
  };
  notifications: {
    emailNotifications: boolean;
    pushNotifications: boolean;
    marketingEmails: boolean;
    securityAlerts: boolean;
    pluginUpdates: boolean;
    newReleases: boolean;
  };
  privacy: {
    profileVisibility: string;
    showEmail: boolean;
    showDownloads: boolean;
    allowAnalytics: boolean;
    shareUsageData: boolean;
  };
  security: {
    twoFactorEnabled: boolean;
    loginAlerts: boolean;
    sessionTimeout: number;
  };
}

const Settings: React.FC = () => {
  const navigate = useNavigate();
  const { isAuthenticated } = useAuthStore();
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [activeTab, setActiveTab] = useState('profile');
  const [showPassword, setShowPassword] = useState(false);
  
  const [settings, setSettings] = useState<UserSettings>({
    profile: {
      username: 'john_doe',
      email: 'john.doe@example.com',
      fullName: 'John Doe',
      bio: '热爱编程的开发者，专注于前端技术',
      avatar: 'https://via.placeholder.com/128',
      website: 'https://johndoe.dev',
      location: '中国 上海',
    },
    preferences: {
      language: 'zh-CN',
      theme: 'light',
      timezone: 'Asia/Shanghai',
      currency: 'CNY',
      autoUpdate: true,
      betaFeatures: false,
    },
    notifications: {
      emailNotifications: true,
      pushNotifications: true,
      marketingEmails: false,
      securityAlerts: true,
      pluginUpdates: true,
      newReleases: true,
    },
    privacy: {
      profileVisibility: 'public',
      showEmail: false,
      showDownloads: true,
      allowAnalytics: true,
      shareUsageData: false,
    },
    security: {
      twoFactorEnabled: false,
      loginAlerts: true,
      sessionTimeout: 30,
    },
  });

  // 检查用户登录状态
  useEffect(() => {
    if (!isAuthenticated) {
      navigate('/login');
    }
  }, [isAuthenticated, navigate]);

  useEffect(() => {
    form.setFieldsValue(settings);
  }, [settings, form]);

  const handleSave = async (values: any) => {
    try {
      setLoading(true);
      
      // 模拟保存设置
      await new Promise(resolve => setTimeout(resolve, 1000));
      
      setSettings(prev => ({
        ...prev,
        ...values,
      }));
      
      message.success('设置已保存');
    } catch (error) {
      message.error('保存失败，请重试');
    } finally {
      setLoading(false);
    }
  };

  const handleAvatarUpload = (info: any) => {
    if (info.file.status === 'done') {
      message.success('头像上传成功');
      setSettings(prev => ({
        ...prev,
        profile: {
          ...prev.profile,
          avatar: info.file.response?.url || prev.profile.avatar,
        },
      }));
    } else if (info.file.status === 'error') {
      message.error('头像上传失败');
    }
  };

  const handleDeleteAccount = () => {
    confirm({
      title: '确认删除账户',
      icon: <ExclamationCircleOutlined />,
      content: '删除账户后，您的所有数据将被永久删除且无法恢复。请确认您要继续此操作。',
      okText: '确认删除',
      okType: 'danger',
      cancelText: '取消',
      onOk() {
        message.success('账户删除请求已提交，我们将在24小时内处理');
      },
    });
  };

  const handleChangePassword = () => {
    Modal.info({
      title: '修改密码',
      content: (
        <div>
          <Paragraph>
            为了您的账户安全，我们将向您的邮箱发送密码重置链接。
          </Paragraph>
          <Paragraph>
            请检查您的邮箱：<Text strong>{settings.profile.email}</Text>
          </Paragraph>
        </div>
      ),
      onOk() {
        message.success('密码重置邮件已发送');
      },
    });
  };

  const renderProfileTab = () => (
    <Form
      form={form}
      layout="vertical"
      onFinish={handleSave}
      initialValues={settings.profile}
    >
      <Row gutter={24}>
        <Col xs={24} md={8}>
          <div style={{ textAlign: 'center', marginBottom: '24px' }}>
            <Avatar
              size={128}
              src={settings.profile.avatar}
              icon={<UserOutlined />}
              style={{ marginBottom: '16px' }}
            />
            <div>
              <Upload
                name="avatar"
                showUploadList={false}
                action="/api/upload/avatar"
                onChange={handleAvatarUpload}
              >
                <Button icon={<UploadOutlined />}>更换头像</Button>
              </Upload>
            </div>
          </div>
        </Col>
        
        <Col xs={24} md={16}>
          <Form.Item
            label="用户名"
            name={['profile', 'username']}
            rules={[{ required: true, message: '请输入用户名' }]}
          >
            <Input placeholder="请输入用户名" />
          </Form.Item>
          
          <Form.Item
            label="邮箱"
            name={['profile', 'email']}
            rules={[
              { required: true, message: '请输入邮箱' },
              { type: 'email', message: '请输入有效的邮箱地址' },
            ]}
          >
            <Input placeholder="请输入邮箱" />
          </Form.Item>
          
          <Form.Item
            label="姓名"
            name={['profile', 'fullName']}
          >
            <Input placeholder="请输入姓名" />
          </Form.Item>
          
          <Form.Item
            label="个人简介"
            name={['profile', 'bio']}
          >
            <Input.TextArea
              rows={3}
              placeholder="介绍一下自己..."
              maxLength={200}
              showCount
            />
          </Form.Item>
          
          <Form.Item
            label="个人网站"
            name={['profile', 'website']}
          >
            <Input placeholder="https://example.com" />
          </Form.Item>
          
          <Form.Item
            label="所在地区"
            name={['profile', 'location']}
          >
            <Input placeholder="请输入所在地区" />
          </Form.Item>
        </Col>
      </Row>
      
      <Form.Item>
        <Button type="primary" htmlType="submit" loading={loading}>
          保存个人信息
        </Button>
      </Form.Item>
    </Form>
  );

  const renderPreferencesTab = () => (
    <Form
      form={form}
      layout="vertical"
      onFinish={handleSave}
      initialValues={settings.preferences}
    >
      <Row gutter={24}>
        <Col xs={24} md={12}>
          <Form.Item
            label="语言"
            name={['preferences', 'language']}
          >
            <Select>
              <Option value="zh-CN">简体中文</Option>
              <Option value="zh-TW">繁體中文</Option>
              <Option value="en-US">English</Option>
              <Option value="ja-JP">日本語</Option>
              <Option value="ko-KR">한국어</Option>
            </Select>
          </Form.Item>
          
          <Form.Item
            label="主题"
            name={['preferences', 'theme']}
          >
            <Radio.Group>
              <Radio value="light">浅色主题</Radio>
              <Radio value="dark">深色主题</Radio>
              <Radio value="auto">跟随系统</Radio>
            </Radio.Group>
          </Form.Item>
          
          <Form.Item
            label="时区"
            name={['preferences', 'timezone']}
          >
            <Select>
              <Option value="Asia/Shanghai">Asia/Shanghai (UTC+8)</Option>
              <Option value="Asia/Tokyo">Asia/Tokyo (UTC+9)</Option>
              <Option value="America/New_York">America/New_York (UTC-5)</Option>
              <Option value="Europe/London">Europe/London (UTC+0)</Option>
            </Select>
          </Form.Item>
        </Col>
        
        <Col xs={24} md={12}>
          <Form.Item
            label="货币"
            name={['preferences', 'currency']}
          >
            <Select>
              <Option value="CNY">人民币 (CNY)</Option>
              <Option value="USD">美元 (USD)</Option>
              <Option value="EUR">欧元 (EUR)</Option>
              <Option value="JPY">日元 (JPY)</Option>
            </Select>
          </Form.Item>
          
          <Form.Item
            label="自动更新"
            name={['preferences', 'autoUpdate']}
            valuePropName="checked"
          >
            <Switch checkedChildren="开启" unCheckedChildren="关闭" />
          </Form.Item>
          
          <Form.Item
            label="测试版功能"
            name={['preferences', 'betaFeatures']}
            valuePropName="checked"
          >
            <Switch checkedChildren="开启" unCheckedChildren="关闭" />
          </Form.Item>
        </Col>
      </Row>
      
      <Form.Item>
        <Button type="primary" htmlType="submit" loading={loading}>
          保存偏好设置
        </Button>
      </Form.Item>
    </Form>
  );

  const renderNotificationsTab = () => (
    <Form
      form={form}
      layout="vertical"
      onFinish={handleSave}
      initialValues={settings.notifications}
    >
      <Title level={4}>通知设置</Title>
      <Paragraph type="secondary">
        选择您希望接收的通知类型
      </Paragraph>
      
      <Space direction="vertical" size="large" style={{ width: '100%' }}>
        <Card size="small">
          <Form.Item
            label="邮件通知"
            name={['notifications', 'emailNotifications']}
            valuePropName="checked"
            style={{ marginBottom: '8px' }}
          >
            <Switch />
          </Form.Item>
          <Text type="secondary">接收重要的邮件通知</Text>
        </Card>
        
        <Card size="small">
          <Form.Item
            label="推送通知"
            name={['notifications', 'pushNotifications']}
            valuePropName="checked"
            style={{ marginBottom: '8px' }}
          >
            <Switch />
          </Form.Item>
          <Text type="secondary">接收浏览器推送通知</Text>
        </Card>
        
        <Card size="small">
          <Form.Item
            label="营销邮件"
            name={['notifications', 'marketingEmails']}
            valuePropName="checked"
            style={{ marginBottom: '8px' }}
          >
            <Switch />
          </Form.Item>
          <Text type="secondary">接收产品更新和促销信息</Text>
        </Card>
        
        <Card size="small">
          <Form.Item
            label="安全警报"
            name={['notifications', 'securityAlerts']}
            valuePropName="checked"
            style={{ marginBottom: '8px' }}
          >
            <Switch />
          </Form.Item>
          <Text type="secondary">接收账户安全相关通知</Text>
        </Card>
        
        <Card size="small">
          <Form.Item
            label="插件更新"
            name={['notifications', 'pluginUpdates']}
            valuePropName="checked"
            style={{ marginBottom: '8px' }}
          >
            <Switch />
          </Form.Item>
          <Text type="secondary">已安装插件有更新时通知</Text>
        </Card>
        
        <Card size="small">
          <Form.Item
            label="新版本发布"
            name={['notifications', 'newReleases']}
            valuePropName="checked"
            style={{ marginBottom: '8px' }}
          >
            <Switch />
          </Form.Item>
          <Text type="secondary">有新插件发布时通知</Text>
        </Card>
      </Space>
      
      <Form.Item style={{ marginTop: '24px' }}>
        <Button type="primary" htmlType="submit" loading={loading}>
          保存通知设置
        </Button>
      </Form.Item>
    </Form>
  );

  const renderSecurityTab = () => (
    <div>
      <Title level={4}>安全设置</Title>
      
      <Space direction="vertical" size="large" style={{ width: '100%' }}>
        <Card>
          <Row justify="space-between" align="middle">
            <Col>
              <Space direction="vertical" size="small">
                <Text strong>修改密码</Text>
                <Text type="secondary">定期更新密码以保护账户安全</Text>
              </Space>
            </Col>
            <Col>
              <Button icon={<KeyOutlined />} onClick={handleChangePassword}>
                修改密码
              </Button>
            </Col>
          </Row>
        </Card>
        
        <Card>
          <Row justify="space-between" align="middle">
            <Col>
              <Space direction="vertical" size="small">
                <Text strong>两步验证</Text>
                <Text type="secondary">
                  为您的账户添加额外的安全保护
                  {settings.security.twoFactorEnabled && (
                    <Tag color="green" style={{ marginLeft: '8px' }}>已启用</Tag>
                  )}
                </Text>
              </Space>
            </Col>
            <Col>
              <Switch
                checked={settings.security.twoFactorEnabled}
                onChange={(checked) => {
                  setSettings(prev => ({
                    ...prev,
                    security: {
                      ...prev.security,
                      twoFactorEnabled: checked,
                    },
                  }));
                  message.success(checked ? '两步验证已启用' : '两步验证已关闭');
                }}
              />
            </Col>
          </Row>
        </Card>
        
        <Card>
          <Row justify="space-between" align="middle">
            <Col>
              <Space direction="vertical" size="small">
                <Text strong>登录警报</Text>
                <Text type="secondary">新设备登录时发送邮件通知</Text>
              </Space>
            </Col>
            <Col>
              <Switch
                checked={settings.security.loginAlerts}
                onChange={(checked) => {
                  setSettings(prev => ({
                    ...prev,
                    security: {
                      ...prev.security,
                      loginAlerts: checked,
                    },
                  }));
                }}
              />
            </Col>
          </Row>
        </Card>
        
        <Card>
          <Space direction="vertical" size="small" style={{ width: '100%' }}>
            <Text strong>会话超时</Text>
            <Text type="secondary">设置自动登出时间（分钟）</Text>
            <Select
              style={{ width: '200px' }}
              value={settings.security.sessionTimeout}
              onChange={(value) => {
                setSettings(prev => ({
                  ...prev,
                  security: {
                    ...prev.security,
                    sessionTimeout: value,
                  },
                }));
              }}
            >
              <Option value={15}>15分钟</Option>
              <Option value={30}>30分钟</Option>
              <Option value={60}>1小时</Option>
              <Option value={120}>2小时</Option>
              <Option value={0}>永不超时</Option>
            </Select>
          </Space>
        </Card>
        
        <Card>
          <Alert
            message="危险操作"
            description="以下操作将永久删除您的账户和所有相关数据"
            type="warning"
            showIcon
            style={{ marginBottom: '16px' }}
          />
          <Button
            danger
            icon={<DeleteOutlined />}
            onClick={handleDeleteAccount}
          >
            删除账户
          </Button>
        </Card>
      </Space>
    </div>
  );

  // 如果用户未登录，不渲染内容（会被重定向到登录页面）
  if (!isAuthenticated) {
    return null;
  }

  return (
    <div style={{ padding: '24px', maxWidth: '1000px', margin: '0 auto' }}>
      {/* 面包屑导航 */}
      <Breadcrumb style={{ marginBottom: '24px' }}>
        <Breadcrumb.Item>
          <HomeOutlined />
          <span onClick={() => navigate('/')} style={{ cursor: 'pointer' }}>
            首页
          </span>
        </Breadcrumb.Item>
        <Breadcrumb.Item>
          <SettingOutlined />
          设置
        </Breadcrumb.Item>
      </Breadcrumb>

      <Title level={2} style={{ marginBottom: '24px' }}>
        <SettingOutlined style={{ marginRight: '8px' }} />
        账户设置
      </Title>

      <Tabs activeKey={activeTab} onChange={setActiveTab}>
        <TabPane
          tab={
            <span>
              <UserOutlined />
              个人资料
            </span>
          }
          key="profile"
        >
          {renderProfileTab()}
        </TabPane>
        
        <TabPane
          tab={
            <span>
              <GlobalOutlined />
              偏好设置
            </span>
          }
          key="preferences"
        >
          {renderPreferencesTab()}
        </TabPane>
        
        <TabPane
          tab={
            <span>
              <BellOutlined />
              通知设置
            </span>
          }
          key="notifications"
        >
          {renderNotificationsTab()}
        </TabPane>
        
        <TabPane
          tab={
            <span>
              <SecurityScanOutlined />
              安全设置
            </span>
          }
          key="security"
        >
          {renderSecurityTab()}
        </TabPane>
      </Tabs>
    </div>
  );
};

export default Settings;