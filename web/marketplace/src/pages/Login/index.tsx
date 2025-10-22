import React, { useState, useEffect } from 'react';
import { Form, Input, Button, Card, Tabs, Checkbox, Divider, Space, Typography, message, Row, Col, Image } from 'antd';
import { UserOutlined, LockOutlined, MailOutlined, EyeInvisibleOutlined, EyeTwoTone, SafetyOutlined } from '@ant-design/icons';
import { Link, useNavigate, useLocation } from 'react-router-dom';
import { useAuthStore, LoginCredentials, RegisterData } from '@/stores/auth';
import './index.css';
import { request } from '@/services/api';

const { Title, Text } = Typography;
const { TabPane } = Tabs;

const Login: React.FC = () => {
  const [loginForm] = Form.useForm();
  const [registerForm] = Form.useForm();
  const [activeTab, setActiveTab] = useState('login');
  const navigate = useNavigate();
  const location = useLocation();
  
  const { login, register, loading, error, isAuthenticated, clearError } = useAuthStore();
  const [captcha, setCaptcha] = useState<{ image: string; key: string } | null>(null);
  const [captchaRefreshing, setCaptchaRefreshing] = useState(false);

  // 根据URL参数设置默认标签页
  useEffect(() => {
    const searchParams = new URLSearchParams(location.search);
    const tab = searchParams.get('tab');
    if (tab === 'register') {
      setActiveTab('register');
    }
  }, [location.search]);

  // 如果已登录，重定向到首页或原来的页面
  useEffect(() => {
    if (isAuthenticated) {
      const from = (location.state as any)?.from?.pathname || '/';
      navigate(from, { replace: true });
    }
  }, [isAuthenticated, navigate, location]);

  // 首次加载验证码
  useEffect(() => {
    fetchCaptcha();
  }, []);

  const fetchCaptcha = async () => {
    try {
      setCaptchaRefreshing(true);
      // 统一使用 axios 客户端调用验证码端点
      const resp: any = await request.get('/auth/captcha');
      const data = resp?.data || resp;
      setCaptcha({ image: data.image, key: data.key });
    } catch (err: any) {
      message.error(err?.message || '获取验证码失败');
    } finally {
      setCaptchaRefreshing(false);
    }
  };

  // 清除错误信息
  useEffect(() => {
    if (error) {
      const timer = setTimeout(() => {
        clearError();
      }, 5000);
      return () => clearTimeout(timer);
    }
  }, [error, clearError]);

  const handleLogin = async (values: LoginCredentials) => {
    const payload: LoginCredentials = { ...values, captcha_key: captcha?.key };
    const success = await login(payload);
    if (success) {
      const from = (location.state as any)?.from?.pathname || '/';
      navigate(from, { replace: true });
    }
  };

  const handleRegister = async (values: RegisterData) => {
    const success = await register(values);
    if (success) {
      setActiveTab('login');
      registerForm.resetFields();
      message.success('注册成功，请登录');
    }
  };

  const handleTabChange = (key: string) => {
    setActiveTab(key);
    clearError();
  };

  return (
    <div className="login-container">
      <div className="login-background">
        <div className="login-overlay" />
      </div>
      
      <div className="login-content">
        <Card className="login-card" bordered={false}>
          <div className="login-header">
            <Title level={2} className="login-title">
              插件市场
            </Title>
            <Text type="secondary" className="login-subtitle">
              发现、安装和管理您的插件
            </Text>
          </div>

          <Tabs 
            activeKey={activeTab} 
            onChange={handleTabChange}
            centered
            className="login-tabs"
          >
            <TabPane tab="登录" key="login">
              <Form
                form={loginForm}
                name="login"
                onFinish={handleLogin}
                autoComplete="off"
                layout="vertical"
                size="large"
                className="login-form"
              >
                <Form.Item
                  name="username"
                  rules={[
                    { required: true, message: '请输入用户名或邮箱' },
                    { min: 3, message: '用户名至少3个字符' }
                  ]}
                >
                  <Input
                    prefix={<UserOutlined />}
                    placeholder="用户名或邮箱"
                    autoComplete="username"
                  />
                </Form.Item
                >
                <Form.Item
                  name="password"
                  rules={[
                    { required: true, message: '请输入密码' },
                    { min: 6, message: '密码至少6个字符' }
                  ]}
                >
                  <Input.Password
                    prefix={<LockOutlined />}
                    placeholder="密码"
                    iconRender={(visible) => (visible ? <EyeTwoTone /> : <EyeInvisibleOutlined />)}
                    autoComplete="current-password"
                  />
                </Form.Item>

                <Row gutter={12} align="middle">
                  <Col span={16}>
                    <Form.Item name="captcha" rules={[{ required: true, message: '请输入验证码' }]}>
                      <Input
                        prefix={<SafetyOutlined />}
                        placeholder="请输入图片中的验证码"
                      />
                    </Form.Item>
                  </Col>
                  <Col span={8}>
                    <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'flex-end' }}>
                      <Image
                        src={captcha?.image}
                        alt="验证码"
                        width={150}
                        height={45}
                        style={{ borderRadius: 6, objectFit: 'cover' }}
                        preview={false}
                      />
                      <Button type="link" onClick={fetchCaptcha} loading={captchaRefreshing}>
                        刷新
                      </Button>
                    </div>
                  </Col>
                </Row>

                <Form.Item>
                  <Checkbox defaultChecked>记住我</Checkbox>
                </Form.Item>

                <Form.Item>
                  <Button type="primary" htmlType="submit" loading={loading} block>
                    登录
                  </Button>
                </Form.Item>

                <Divider plain>
                  <Space>
                    <Link to="/forgot-password">忘记密码？</Link>
                    <span>或</span>
                    <a onClick={() => setActiveTab('register')}>创建账号</a>
                  </Space>
                </Divider>
              </Form>
            </TabPane>

            <TabPane tab="注册" key="register">
              <Form
                form={registerForm}
                name="register"
                onFinish={handleRegister}
                autoComplete="off"
                layout="vertical"
                size="large"
                className="login-form"
              >
                <Form.Item
                  name="username"
                  rules={[
                    { required: true, message: '请输入用户名' },
                    { min: 3, message: '用户名至少3个字符' }
                  ]}
                >
                  <Input prefix={<UserOutlined />} placeholder="用户名" autoComplete="username" />
                </Form.Item>

                <Form.Item
                  name="email"
                  rules={[
                    { required: true, message: '请输入邮箱' },
                    { type: 'email', message: '请输入有效的邮箱地址' }
                  ]}
                >
                  <Input prefix={<MailOutlined />} placeholder="邮箱" autoComplete="email" />
                </Form.Item>

                <Form.Item
                  name="password"
                  rules={[
                    { required: true, message: '请输入密码' },
                    { min: 6, message: '密码至少6个字符' }
                  ]}
                >
                  <Input.Password
                    prefix={<LockOutlined />}
                    placeholder="密码"
                    iconRender={(visible) => (visible ? <EyeTwoTone /> : <EyeInvisibleOutlined />)}
                    autoComplete="new-password"
                  />
                </Form.Item>

                <Form.Item
                  name="confirmPassword"
                  rules={[
                    { required: true, message: '请确认密码' },
                    ({ getFieldValue }) => ({
                      validator(_, value) {
                        if (!value || getFieldValue('password') === value) {
                          return Promise.resolve();
                        }
                        return Promise.reject(new Error('两次输入的密码不一致'));
                      },
                    }),
                  ]}
                >
                  <Input.Password
                    prefix={<LockOutlined />}
                    placeholder="确认密码"
                    iconRender={(visible) => (visible ? <EyeTwoTone /> : <EyeInvisibleOutlined />)}
                    autoComplete="new-password"
                  />
                </Form.Item>

                <Form.Item>
                  <Space>
                    <Button type="primary" htmlType="submit" loading={loading}>
                      创建账号
                    </Button>
                    <Button onClick={() => registerForm.resetFields()}>重置</Button>
                  </Space>
                </Form.Item>
              </Form>
            </TabPane>
          </Tabs>
        </Card>
      </div>
    </div>
  );
};

export default Login;