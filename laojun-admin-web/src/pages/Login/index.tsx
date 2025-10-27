import { useState, useEffect, useRef } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { Form, Input, Button, Card, App, Row, Col, Image, Space, Alert, Checkbox } from 'antd';
import { UserOutlined, LockOutlined, SafetyOutlined, EyeInvisibleOutlined, EyeTwoTone } from '@ant-design/icons';
import { useAuthStore } from '@/stores/auth';
import { authService } from '@/services/auth';

interface LoginForm {
  username: string;
  password: string;
  captcha?: string;
}

const Login: React.FC = () => {
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [captcha, setCaptcha] = useState<{ image: string; key: string } | null>(null);
  const [captchaRefreshing, setCaptchaRefreshing] = useState(false);
  const [captchaEnabled, setCaptchaEnabled] = useState(true); // 验证码是否启用
  const hasFetchedCaptchaRef = useRef(false);
  const navigate = useNavigate();
  const location = useLocation();
  const { login, isAuthenticated } = useAuthStore();
  const { message } = App.useApp();
  const [rateLimitSeconds, setRateLimitSeconds] = useState<number>(0)
  const rateLimitTimerRef = useRef<number | null>(null)

  // 如果已经登录，重定向到目标页面
  useEffect(() => {
    if (isAuthenticated) {
      const from = (location.state as any)?.from || '/dashboard';
      navigate(from, { replace: true });
    }
  }, [isAuthenticated, navigate, location]);

  // 获取验证码配置
  useEffect(() => {
    const fetchCaptchaConfig = async () => {
      try {
        const config = await authService.getCaptchaConfig();
        setCaptchaEnabled(config.enabled);
        
        // 如果验证码启用，则获取验证码
        if (config.enabled && !hasFetchedCaptchaRef.current) {
          hasFetchedCaptchaRef.current = true;
          await fetchCaptcha();
        }
      } catch (error) {
        console.error('获取验证码配置失败:', error);
        // 出错时默认启用验证码，保持向后兼容
        setCaptchaEnabled(true);
        if (!hasFetchedCaptchaRef.current) {
          hasFetchedCaptchaRef.current = true;
          await fetchCaptcha();
        }
      }
    };

    fetchCaptchaConfig();
  }, []);

  // 获取验证码
  const fetchCaptcha = async () => {
    try {
      setCaptchaRefreshing(true)
      console.log('开始获取验证码...')
      const captchaData = await authService.getCaptcha()
      console.log('验证码数据:', captchaData)
      console.log('验证码图片数据长度:', captchaData?.image?.length)
      console.log('验证码Key:', captchaData?.key)
      setCaptcha(captchaData)
      // 刷新时清空验证码输入并聚焦，避免旧输入与新key不匹配
      form.setFieldsValue({ captcha: '' })
      const el = document.getElementById('captcha-input') as HTMLInputElement | null
      el?.focus()
    } catch (error) {
      console.error('Failed to fetch captcha:', error)
      console.error('错误详情:', error)
    } finally {
      setCaptchaRefreshing(false)
    }
  }


  // 添加captcha状态变化的监听
  useEffect(() => {
    console.log('captcha 状态变化:', captcha);
  }, [captcha]);

  // 处理登录
  const handleLogin = async (values: LoginForm) => {
    if (rateLimitSeconds > 0) {
      const mins = Math.floor(rateLimitSeconds / 60)
      const secs = rateLimitSeconds % 60
      message.warning(`请求过于频繁，请在 ${mins}分${secs}秒 后再试`)
      return
    }
    setLoading(true)
    try {
      // 规范化输入，避免隐藏空格导致认证失败
      const username = (values.username || '').trim()
      const captchaInput = captchaEnabled ? (values.captcha || '').trim() : ''

      // 构建登录请求数据，只在启用验证码时包含验证码参数
      const loginData = {
        username,
        password: values.password, // 密码不裁剪，尊重后端真实性能
        ...(captchaEnabled && {
          captcha: captchaInput,
          captcha_key: captcha?.key,
        }),
      }
      
      await login(
        loginData.username, 
        loginData.password, 
        captchaEnabled ? loginData.captcha : undefined, 
        captchaEnabled ? loginData.captcha_key : undefined
      )
      
      // 记住我：在成功登录后处理持久化
      const rememberMe = !!form.getFieldValue('rememberMe')
      if (rememberMe) {
        localStorage.setItem('rememberMe', 'true')
        localStorage.setItem('rememberedUsername', loginData.username)
      } else {
        localStorage.removeItem('rememberMe')
        localStorage.removeItem('rememberedUsername')
      }

      message.success('登录成功')
    } catch (error: any) {
      console.error('Login failed:', error)
      
      // 解析错误信息，支持多种后端错误格式
      let errorMessage = '登录失败，请检查用户名和密码'
      
      if (error.response?.data) {
        const data = error.response.data
        // 支持字符串或 JSON 对象的错误响应格式
        if (typeof data === 'string') {
          try {
            const parsed = JSON.parse(data)
            errorMessage = parsed.message || parsed.error || parsed.msg || data
          } catch {
            errorMessage = data
          }
        } else {
          errorMessage = data.message || data.error || data.msg || errorMessage
        }
      } else if (error.message) {
        errorMessage = error.message
      }

      // 友好化通用后端错误文案 + 429 冷却
      const lower = String(errorMessage).toLowerCase()
      if (lower.includes('invalid captcha')) {
        errorMessage = '验证码错误或已过期，请重新输入'
      } else if (lower.includes('invalid username or password')) {
        errorMessage = '用户名或密码不正确'
      }

      // 429: Too Many Requests，显示冷却倒计时
      const status = error.response?.status
      const retryAfter = parseInt(error.response?.data?.retry_after) || parseInt(error.response?.headers?.['retry-after']) || 0
      if (status === 429 && retryAfter > 0) {
        setRateLimitSeconds(retryAfter)
        const mins = Math.floor(retryAfter / 60)
        const secs = retryAfter % 60
        errorMessage = `请求过于频繁，请在 ${mins}分${secs}秒 后再试`
      }
      
      message.error(errorMessage)
      
      // 登录失败后刷新验证码并清空输入
      fetchCaptcha()
      form.setFieldsValue({ captcha: '' })
    } finally {
      setLoading(false)
    }
  }

  return (
    <div style={{
      minHeight: '100vh',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      background: 'linear-gradient(135deg, #f0f5ff 0%, #ffffff 100%)',
      padding: '24px'
    }}>
      <Row justify="center" style={{ width: '100%' }}>
        <Col xs={24} sm={20} md={16} lg={12} xl={8}>
          <Card
            style={{
              borderRadius: '12px',
              boxShadow: '0 8px 32px rgba(0, 0, 0, 0.1)',
              border: 'none',
            }}
          >
            {/* 头部 */}
            <div style={{ textAlign: 'center', marginBottom: '32px' }}>
              <div
                style={{
                  fontSize: '32px',
                  fontWeight: 'bold',
                  color: '#1890ff',
                  marginBottom: '8px',
                }}
              >
                太上老君
              </div>
              <div style={{ color: '#666', fontSize: '16px' }}>
                管理后台系统
              </div>
            </div>

            {/* 登录表单 */}
            <Form
              form={form}
              onFinish={handleLogin}
              autoComplete="off"
              size="large"
            >
              <Form.Item
                name="username"
                rules={[
                  { required: true, message: '请输入用户名' },
                  { min: 3, message: '用户名至少3个字符' },
                ]}
              >
                <Input
                  prefix={<UserOutlined />}
                  placeholder="用户名"
                  autoComplete="username"
                />
              </Form.Item>

              <Form.Item
                name="password"
                rules={[
                  { required: true, message: '请输入密码' },
                  { min: 6, message: '密码至少6个字符' },
                ]}
              >
                <Input.Password
                  prefix={<LockOutlined />}
                  placeholder="密码"
                  autoComplete="current-password"
                  iconRender={(visible) => (visible ? <EyeTwoTone /> : <EyeInvisibleOutlined />)}
                />
              </Form.Item>

              {/* 验证码 - 根据配置决定是否显示 */}
              {captchaEnabled && (
                <Form.Item
                  name="captcha"
                  rules={[{ required: true, message: '请输入验证码' }]}
                >
                  <Row gutter={8}>
                    <Col span={14}>
                      <Input
                        id="captcha-input"
                        prefix={<SafetyOutlined />}
                        placeholder="验证码"
                        autoComplete="off"
                      />
                    </Col>
                    <Col span={10}>
                      {captcha ? (
                        <Image
                          src={captcha.image}
                          alt="验证码"
                          style={{
                            width: '100%',
                            height: '40px',
                            cursor: 'pointer',
                            borderRadius: '6px',
                          }}
                          preview={false}
                          onClick={fetchCaptcha}
                        />
                      ) : (
                        <div
                          style={{
                            width: '100%',
                            height: '40px',
                            cursor: 'pointer',
                            borderRadius: '6px',
                            border: '1px dashed #d9d9d9',
                            display: 'flex',
                            alignItems: 'center',
                            justifyContent: 'center',
                            color: '#999',
                            fontSize: '12px',
                          }}
                          onClick={fetchCaptcha}
                        >
                          {captchaRefreshing ? '加载中...' : '点击获取验证码'}
                        </div>
                      )}
                    </Col>
                  </Row>
                </Form.Item>
              )}

              <Form.Item name="rememberMe" valuePropName="checked" style={{ marginBottom: 12 }}>
                <Checkbox>记住我</Checkbox>
              </Form.Item>
              <Form.Item style={{ marginBottom: '16px' }}>
                <Button
                  type="primary"
                  htmlType="submit"
                  loading={loading}
                  disabled={loading || rateLimitSeconds > 0}
                  style={{
                    width: '100%',
                    height: '44px',
                    borderRadius: '8px',
                    fontSize: '16px',
                    fontWeight: 500,
                  }}
                >
                  {rateLimitSeconds > 0 ? '冷却中' : '登录'}
                </Button>
              </Form.Item>
              {rateLimitSeconds > 0 && (
                <Alert
                  type="warning"
                  showIcon
                  message="请求过于频繁"
                  description={`请稍后再试（剩余 ${Math.floor(rateLimitSeconds/60)}分${rateLimitSeconds%60}秒）`}
                  style={{ marginBottom: 16 }}
                />
              )}
            </Form>

            {/* 底部信息 */}
            <div style={{ textAlign: 'center', color: '#999', fontSize: '12px' }}>
              <Space direction="vertical" size={4}>
                <div>请使用后端配置的账号登录（常见：admin / admin123）</div>
                {captchaEnabled && <div>点击验证码可刷新</div>}
              </Space>
            </div>
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default Login;