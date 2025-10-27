import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Row,
  Col,
  Card,
  Typography,
  Space,
  Button,
  Form,
  Input,
  Select,
  Radio,
  Divider,
  List,
  Avatar,
  Tag,
  message,
  Steps,
  Alert,
  Spin,
  Modal,
  QRCode,
  Result,
} from 'antd';
import {
  CreditCardOutlined,
  SafetyOutlined,
  CheckCircleOutlined,
  ShoppingCartOutlined,
  UserOutlined,
  MailOutlined,
  PhoneOutlined,
  HomeOutlined,
  PayCircleOutlined,
  AppstoreOutlined,
  AlipayOutlined,
  WechatOutlined,
  LoadingOutlined,
  ArrowLeftOutlined,
  GiftOutlined,
} from '@ant-design/icons';
import { useCartStore } from '@/stores/cart';
import { useOrderStore } from '@/stores/order';
import { Plugin, PaymentMethod, BillingInfo } from '@/types';

const { Title, Text, Paragraph } = Typography;
const { Option } = Select;
const { Step } = Steps;

interface PaymentMethod {
  id: string;
  name: string;
  icon: React.ReactNode;
  description: string;
}

interface OrderInfo {
  orderId: string;
  items: Plugin[];
  subtotal: number;
  tax: number;
  total: number;
  paymentMethod: string;
  billingInfo: any;
}

const Checkout: React.FC = () => {
  const navigate = useNavigate();
  const { items, total, clearCart } = useCartStore();
  const { 
    createOrder, 
    processPayment, 
    currentOrder, 
    loading: orderLoading, 
    error: orderError,
    clearError 
  } = useOrderStore();
  
  const [currentStep, setCurrentStep] = useState(0);
  const [loading, setLoading] = useState(false);
  const [paymentLoading, setPaymentLoading] = useState(false);
  const [orderInfo, setOrderInfo] = useState<OrderInfo | null>(null);
  const [billingInfo, setBillingInfo] = useState<BillingInfo>({
    firstName: '',
    lastName: '',
    email: '',
    phone: '',
    address: '',
    city: '',
    state: '',
    zipCode: '',
    country: 'CN',
  });
  const [selectedPaymentMethod, setSelectedPaymentMethod] = useState<PaymentMethod>('alipay');
  const [paymentModalVisible, setPaymentModalVisible] = useState(false);
  const [qrCodeData, setQrCodeData] = useState('');
  const [form] = Form.useForm();

  // 支付方式配置
  const paymentMethods = [
    {
      id: 'alipay' as PaymentMethod,
      name: '支付宝',
      icon: <AlipayOutlined style={{ fontSize: 24, color: '#1677ff' }} />,
      description: '使用支付宝安全支付',
      fee: 0,
    },
    {
      id: 'wechat' as PaymentMethod,
      name: '微信支付',
      icon: <WechatOutlined style={{ fontSize: 24, color: '#52c41a' }} />,
      description: '使用微信安全支付',
      fee: 0,
    },
    {
      id: 'paypal' as PaymentMethod,
      name: 'PayPal',
      icon: <PayCircleOutlined style={{ fontSize: 24, color: '#faad14' }} />,
      description: '国际支付，支持多种货币',
      fee: total * 0.029, // 2.9% 手续费
    },
    {
      id: 'credit_card' as PaymentMethod,
      name: '信用卡',
      icon: <CreditCardOutlined style={{ fontSize: 24, color: '#722ed1' }} />,
      description: '支持 Visa、MasterCard 等',
      fee: total * 0.025, // 2.5% 手续费
    },
  ];

  const steps = [
    {
      title: '确认订单',
      icon: <ShoppingCartOutlined />,
      description: '确认购买的插件',
    },
    {
      title: '填写信息',
      icon: <SafetyOutlined />,
      description: '填写账单信息',
    },
    {
      title: '选择支付',
      icon: <CreditCardOutlined />,
      description: '选择支付方式',
    },
    {
      title: '完成支付',
      icon: <CheckCircleOutlined />,
      description: '支付并完成订单',
    },
  ];

  useEffect(() => {
    if (items.length === 0) {
      message.warning('购物车为空，请先添加插件');
      navigate('/cart');
    }
  }, [items, navigate]);

  const calculateTotal = () => {
    const subtotal = total;
    const selectedMethod = paymentMethods.find(method => method.id === selectedPaymentMethod);
    const paymentFee = selectedMethod?.fee || 0;
    const tax = subtotal * 0.1; // 10% 税费
    const discount = 0; // 暂无折扣
    return {
      subtotal,
      tax,
      discount,
      paymentFee,
      total: subtotal + tax + paymentFee - discount,
    };
  };

  const handleNext = async () => {
    if (currentStep === 1) {
      // 验证账单信息
      try {
        await form.validateFields();
        const values = form.getFieldsValue();
        setBillingInfo(values);
      } catch (error) {
        message.error('请填写完整的账单信息');
        return;
      }
    }
    
    if (currentStep < steps.length - 1) {
      setCurrentStep(currentStep + 1);
    }
  };

  const handlePrev = () => {
    if (currentStep > 0) {
      setCurrentStep(currentStep - 1);
    }
  };

  const handleCreateOrder = async () => {
    setLoading(true);
    clearError();
    
    try {
      // 创建订单
      const order = await createOrder(
        items,
        billingInfo,
        selectedPaymentMethod,
        'user-123' // 模拟用户ID，实际应该从用户状态获取
      );

      if (order) {
        setCurrentStep(3); // 跳转到支付步骤
      }
    } catch (error) {
      message.error('创建订单失败，请重试');
    } finally {
      setLoading(false);
    }
  };

  const handlePayment = async () => {
    if (!currentOrder) {
      message.error('订单信息不存在');
      return;
    }

    setPaymentLoading(true);
    
    try {
      // 处理支付
      const success = await processPayment(currentOrder.id, selectedPaymentMethod);
      
      if (success) {
        // 显示支付二维码或跳转支付页面
        if (selectedPaymentMethod === 'alipay' || selectedPaymentMethod === 'wechat') {
          setQrCodeData(`payment://${selectedPaymentMethod}/${currentOrder.id}`);
          setPaymentModalVisible(true);
        }
        
        // 模拟支付完成
        setTimeout(() => {
          clearCart();
          setPaymentModalVisible(false);
          message.success('支付成功！');
          navigate('/orders');
        }, 5000);
      }
    } catch (error) {
      message.error('支付失败，请重试');
    } finally {
      setPaymentLoading(false);
    }
  };

  const handleFinish = () => {
    navigate('/profile/orders');
  };

  const renderOrderSummary = () => {
    const totals = calculateTotal();
    
    return (
      <Card title="订单摘要" className="mb-4">
        <List
          dataSource={items}
          renderItem={(item) => (
            <List.Item>
              <List.Item.Meta
                avatar={<Avatar src={item.plugin.icon} size={48} />}
                title={item.plugin.name}
                description={
                  <Space direction="vertical" size="small">
                    <Text type="secondary">{item.plugin.developer}</Text>
                    <Text>数量: {item.quantity}</Text>
                  </Space>
                }
              />
              <div className="text-right">
                <Text strong>¥{(item.plugin.price * item.quantity).toFixed(2)}</Text>
              </div>
            </List.Item>
          )}
        />
        <Divider />
        <div className="space-y-2">
          <div className="flex justify-between">
            <Text>小计:</Text>
            <Text>¥{totals.subtotal.toFixed(2)}</Text>
          </div>
          <div className="flex justify-between">
            <Text>税费:</Text>
            <Text>¥{totals.tax.toFixed(2)}</Text>
          </div>
          {totals.paymentFee > 0 && (
            <div className="flex justify-between">
              <Text>支付手续费:</Text>
              <Text>¥{totals.paymentFee.toFixed(2)}</Text>
            </div>
          )}
          {totals.discount > 0 && (
            <div className="flex justify-between">
              <Text>折扣:</Text>
              <Text type="success">-¥{totals.discount.toFixed(2)}</Text>
            </div>
          )}
          <Divider />
          <div className="flex justify-between">
            <Title level={4}>总计:</Title>
            <Title level={4} type="danger">¥{totals.total.toFixed(2)}</Title>
          </div>
        </div>
      </Card>
    );
  };

  const renderBillingForm = () => (
    <Card title="账单信息" className="mb-4">
      <Form 
        form={form} 
        layout="vertical"
        initialValues={billingInfo}
        onValuesChange={(_, allValues) => setBillingInfo(allValues)}
      >
        <Row gutter={16}>
          <Col span={12}>
            <Form.Item
              name="firstName"
              label="名"
              rules={[{ required: true, message: '请输入名' }]}
            >
              <Input placeholder="请输入名" />
            </Form.Item>
          </Col>
          <Col span={12}>
            <Form.Item
              name="lastName"
              label="姓"
              rules={[{ required: true, message: '请输入姓' }]}
            >
              <Input placeholder="请输入姓" />
            </Form.Item>
          </Col>
        </Row>
        
        <Form.Item
          name="email"
          label="邮箱"
          rules={[
            { required: true, message: '请输入邮箱' },
            { type: 'email', message: '请输入有效的邮箱地址' }
          ]}
        >
          <Input placeholder="请输入邮箱地址" />
        </Form.Item>
        
        <Form.Item
          name="phone"
          label="电话"
          rules={[{ required: true, message: '请输入电话号码' }]}
        >
          <Input placeholder="请输入电话号码" />
        </Form.Item>
        
        <Form.Item
          name="address"
          label="地址"
          rules={[{ required: true, message: '请输入地址' }]}
        >
          <Input placeholder="请输入详细地址" />
        </Form.Item>
        
        <Row gutter={16}>
          <Col span={8}>
            <Form.Item
              name="city"
              label="城市"
              rules={[{ required: true, message: '请输入城市' }]}
            >
              <Input placeholder="城市" />
            </Form.Item>
          </Col>
          <Col span={8}>
            <Form.Item
              name="state"
              label="省份"
              rules={[{ required: true, message: '请选择省份' }]}
            >
              <Select placeholder="选择省份">
                <Option value="beijing">北京</Option>
                <Option value="shanghai">上海</Option>
                <Option value="guangdong">广东</Option>
                <Option value="zhejiang">浙江</Option>
                <Option value="jiangsu">江苏</Option>
                <Option value="sichuan">四川</Option>
                <Option value="hubei">湖北</Option>
                <Option value="hunan">湖南</Option>
              </Select>
            </Form.Item>
          </Col>
          <Col span={8}>
            <Form.Item
              name="zipCode"
              label="邮编"
              rules={[{ required: true, message: '请输入邮编' }]}
            >
              <Input placeholder="邮编" />
            </Form.Item>
          </Col>
        </Row>
        
        <Form.Item
          name="country"
          label="国家"
          rules={[{ required: true, message: '请选择国家' }]}
          initialValue="CN"
        >
          <Select>
            <Option value="CN">中国</Option>
            <Option value="US">美国</Option>
            <Option value="UK">英国</Option>
            <Option value="JP">日本</Option>
            <Option value="DE">德国</Option>
            <Option value="FR">法国</Option>
          </Select>
        </Form.Item>
      </Form>
    </Card>
  );

  const renderPaymentMethods = () => {
    const totals = calculateTotal();
    
    return (
      <Card title="支付方式" className="mb-4">
        <Radio.Group
          value={selectedPaymentMethod}
          onChange={(e) => setSelectedPaymentMethod(e.target.value)}
          className="w-full"
        >
          <Space direction="vertical" size="middle" className="w-full">
            {paymentMethods.map((method) => (
              <Radio key={method.id} value={method.id} className="w-full">
                <Card
                  size="small"
                  className={`cursor-pointer transition-all ${
                    selectedPaymentMethod === method.id
                      ? 'border-blue-500 bg-blue-50'
                      : 'border-gray-200'
                  }`}
                  bodyStyle={{ padding: '12px 16px' }}
                >
                  <div className="flex items-center justify-between">
                    <div className="flex items-center space-x-3">
                      <div>{method.icon}</div>
                      <div>
                        <div className="font-medium">{method.name}</div>
                        <div className="text-sm text-gray-500">{method.description}</div>
                        {method.fee > 0 && (
                          <div className="text-xs text-orange-500">
                            手续费: ¥{method.fee.toFixed(2)}
                          </div>
                        )}
                      </div>
                    </div>
                    <div className="text-right">
                      <SafetyOutlined className="text-green-500" />
                      {method.fee === 0 && (
                        <div className="text-xs text-green-500 mt-1">免手续费</div>
                      )}
                    </div>
                  </div>
                </Card>
              </Radio>
            ))}
          </Space>
        </Radio.Group>
        
        <Alert
          message="安全提示"
          description="我们使用业界标准的加密技术保护您的支付信息安全，所有交易均通过SSL加密传输"
          type="info"
          showIcon
          className="mt-4"
        />
        
        <div className="mt-4 p-3 bg-gray-50 rounded">
          <div className="flex justify-between items-center">
            <Text strong>应付金额:</Text>
            <Title level={4} type="danger" className="mb-0">
              ¥{totals.total.toFixed(2)}
            </Title>
          </div>
        </div>
      </Card>
    );
  };

  const renderPaymentProcessing = () => {
    if (paymentLoading) {
      return (
        <div className="text-center py-8">
          <Spin size="large" indicator={<LoadingOutlined style={{ fontSize: 48 }} spin />} />
          <Title level={3} className="mt-4">正在处理支付...</Title>
          <Paragraph type="secondary">
            请不要关闭页面，支付完成后将自动跳转
          </Paragraph>
          
          {(selectedPaymentMethod === 'alipay' || selectedPaymentMethod === 'wechat') && qrCodeData && (
            <div className="mt-6">
              <Card className="inline-block">
                <QRCode value={qrCodeData} size={200} />
                <div className="mt-2 text-center">
                  <Text type="secondary">
                    请使用{selectedPaymentMethod === 'alipay' ? '支付宝' : '微信'}扫码支付
                  </Text>
                </div>
              </Card>
            </div>
          )}
        </div>
      );
    }

    if (currentOrder) {
      return (
        <Result
          status="success"
          title="支付成功！"
          subTitle={`订单号: ${currentOrder.id} | 支付金额: ¥${currentOrder.total.toFixed(2)}`}
          extra={[
            <Button type="primary" key="orders" onClick={() => navigate('/orders')}>
              查看订单
            </Button>,
            <Button key="continue" onClick={() => navigate('/')}>
              继续购物
            </Button>,
          ]}
        />
      );
    }

    return null;
  };

  const renderStepContent = () => {
    switch (currentStep) {
      case 0:
        return renderOrderSummary();
      case 1:
        return renderBillingForm();
      case 2:
        return renderPaymentMethods();
      case 3:
        return renderPaymentProcessing();
      default:
        return null;
    }
  };

  // 错误处理
  useEffect(() => {
    if (orderError) {
      message.error(orderError);
    }
  }, [orderError]);

  if (items.length === 0) {
    return (
      <div className="container mx-auto px-4 py-8">
        <Result
          status="404"
          title="购物车为空"
          subTitle="您的购物车中没有商品，请先添加一些插件"
          extra={
            <Button type="primary" onClick={() => navigate('/')}>
              去购物
            </Button>
          }
        />
      </div>
    );
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="max-w-4xl mx-auto">
        <div className="mb-8">
          <Button
            icon={<ArrowLeftOutlined />}
            onClick={() => navigate('/cart')}
            className="mb-4"
          >
            返回购物车
          </Button>
          <Title level={2}>结账</Title>
        </div>

        <Steps current={currentStep} className="mb-8">
          {steps.map((step, index) => (
            <Steps.Step
              key={index}
              title={step.title}
              description={step.description}
              icon={step.icon}
            />
          ))}
        </Steps>

        <Row gutter={24}>
          <Col span={16}>
            {renderStepContent()}
          </Col>
          <Col span={8}>
            {currentStep < 3 && (
              <Card title="订单摘要" className="sticky top-4">
                <List
                  dataSource={items}
                  renderItem={(item) => (
                    <List.Item>
                      <List.Item.Meta
                        avatar={<Avatar src={item.plugin.icon} size={40} />}
                        title={item.plugin.name}
                        description={`数量: ${item.quantity}`}
                      />
                      <Text strong>¥{(item.plugin.price * item.quantity).toFixed(2)}</Text>
                    </List.Item>
                  )}
                />
                <Divider />
                <div className="flex justify-between">
                  <Text strong>总计:</Text>
                  <Text strong type="danger">¥{calculateTotal().total.toFixed(2)}</Text>
                </div>
              </Card>
            )}
          </Col>
        </Row>

        {currentStep < 3 && (
          <div className="mt-8 text-center">
            <Space>
              {currentStep > 0 && (
                <Button onClick={handlePrev} disabled={loading || orderLoading}>
                  上一步
                </Button>
              )}
              {currentStep < 2 && (
                <Button 
                  type="primary" 
                  onClick={handleNext}
                  loading={loading || orderLoading}
                >
                  下一步
                </Button>
              )}
              {currentStep === 2 && (
                <Button
                  type="primary"
                  loading={loading || orderLoading}
                  onClick={handleCreateOrder}
                >
                  创建订单并支付
                </Button>
              )}
            </Space>
          </div>
        )}

        {/* 支付二维码模态框 */}
        <Modal
          title={`${selectedPaymentMethod === 'alipay' ? '支付宝' : '微信'}支付`}
          open={paymentModalVisible}
          footer={null}
          onCancel={() => setPaymentModalVisible(false)}
          centered
        >
          <div className="text-center py-4">
            <QRCode value={qrCodeData} size={250} />
            <div className="mt-4">
              <Text type="secondary">
                请使用{selectedPaymentMethod === 'alipay' ? '支付宝' : '微信'}扫描二维码完成支付
              </Text>
            </div>
            <div className="mt-2">
              <Text strong>支付金额: ¥{currentOrder?.total.toFixed(2)}</Text>
            </div>
          </div>
        </Modal>
      </div>
    </div>
  );
};

export default Checkout;