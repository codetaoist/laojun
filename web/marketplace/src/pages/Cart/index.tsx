import React, { useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Card,
  List,
  Button,
  InputNumber,
  Avatar,
  Typography,
  Space,
  Divider,
  Empty,
  Row,
  Col,
  Statistic,
  message,
} from 'antd';
import {
  DeleteOutlined,
  ShoppingCartOutlined,
  CreditCardOutlined,
  UserOutlined,
} from '@ant-design/icons';
import { useCartStore } from '@/stores/cart';

const { Title, Text } = Typography;

const Cart: React.FC = () => {
  const navigate = useNavigate();
  const {
    items,
    total,
    currency,
    removeFromCart,
    updateQuantity,
    clearCart,
    getItemCount,
    calculateTotal,
  } = useCartStore();

  // 确保在组件加载时计算总价
  useEffect(() => {
    calculateTotal();
  }, [items, calculateTotal]);

  const handleQuantityChange = (pluginId: string, quantity: number) => {
    if (quantity <= 0) {
      removeFromCart(pluginId);
    } else {
      updateQuantity(pluginId, quantity);
    }
  };

  const handleRemove = (pluginId: string) => {
    removeFromCart(pluginId);
    message.success('已从购物车移除');
  };

  const handleClearCart = () => {
    clearCart();
    message.success('购物车已清空');
  };

  const handleCheckout = () => {
    if (items.length === 0) {
      message.warning('购物车为空');
      return;
    }
    
    // 这里应该跳转到结算页面
    message.info('跳转到结算页面...');
    navigate('/checkout');
  };

  const handleContinueShopping = () => {
    navigate('/');
  };

  if (items.length === 0) {
    return (
      <div style={{ padding: '24px', maxWidth: '1200px', margin: '0 auto' }}>
        <Card>
          <Empty
            image={<ShoppingCartOutlined style={{ fontSize: '64px', color: '#d9d9d9' }} />}
            description={
              <div>
                <Title level={4}>购物车为空</Title>
                <Text type="secondary">还没有添加任何插件到购物车</Text>
              </div>
            }
          >
            <Button type="primary" onClick={handleContinueShopping}>
              去逛逛
            </Button>
          </Empty>
        </Card>
      </div>
    );
  }

  return (
    <div style={{ padding: '24px', maxWidth: '1200px', margin: '0 auto' }}>
      <Title level={2} style={{ marginBottom: '24px' }}>
        <ShoppingCartOutlined /> 购物车 ({getItemCount()} 件商品)
      </Title>

      <Row gutter={[24, 24]}>
        {/* 左侧商品列表 */}
        <Col xs={24} lg={16}>
          <Card
            title="商品列表"
            extra={
              <Button type="link" danger onClick={handleClearCart}>
                清空购物车
              </Button>
            }
          >
            <List
              dataSource={items}
              renderItem={(item) => (
                <List.Item
                  actions={[
                    <InputNumber
                      min={1}
                      max={10}
                      value={item.quantity}
                      onChange={(value) => handleQuantityChange(item.plugin.id, value || 1)}
                      size="small"
                    />,
                    <Button
                      type="text"
                      danger
                      icon={<DeleteOutlined />}
                      onClick={() => handleRemove(item.plugin.id)}
                    >
                      移除
                    </Button>,
                  ]}
                >
                  <List.Item.Meta
                    avatar={
                      <Avatar
                        size={64}
                        src={item.plugin.icon}
                        icon={<UserOutlined />}
                        style={{ cursor: 'pointer' }}
                        onClick={() => navigate(`/plugin/${item.plugin.id}`)}
                      />
                    }
                    title={
                      <div>
                        <Text
                          strong
                          style={{ cursor: 'pointer' }}
                          onClick={() => navigate(`/plugin/${item.plugin.id}`)}
                        >
                          {item.plugin.name || '未知插件'}
                        </Text>
                        <div style={{ marginTop: '4px' }}>
                          <Text type="secondary" style={{ fontSize: '12px' }}>
                            开发者: {item.plugin.author || '未知'}
                          </Text>
                        </div>
                      </div>
                    }
                    description={
                      <div>
                        <Text type="secondary">{item.plugin.description || '暂无描述'}</Text>
                        <div style={{ marginTop: '8px' }}>
                          <Space>
                            <Text strong style={{ color: '#ff4d4f', fontSize: '16px' }}>
                              ¥{item.plugin.price || 0}
                            </Text>
                            {item.quantity > 1 && (
                              <Text type="secondary">
                                x {item.quantity} = ¥{((item.plugin.price || 0) * item.quantity).toFixed(2)}
                              </Text>
                            )}
                          </Space>
                        </div>
                      </div>
                    }
                  />
                </List.Item>
              )}
            />
          </Card>
        </Col>

        {/* 右侧结算信息 */}
        <Col xs={24} lg={8}>
          <Card title="订单摘要">
            <div style={{ marginBottom: '16px' }}>
              <Row justify="space-between">
                <Col>商品数量:</Col>
                <Col>{getItemCount()} 件</Col>
              </Row>
            </div>

            <div style={{ marginBottom: '16px' }}>
              <Row justify="space-between">
                <Col>商品总价:</Col>
                <Col>¥{(total || 0).toFixed(2)}</Col>
              </Row>
            </div>

            <div style={{ marginBottom: '16px' }}>
              <Row justify="space-between">
                <Col>优惠券:</Col>
                <Col>
                  <Button type="link" size="small">
                    选择优惠券
                  </Button>
                </Col>
              </Row>
            </div>

            <Divider />

            <div style={{ marginBottom: '24px' }}>
              <Row justify="space-between">
                <Col>
                  <Text strong style={{ fontSize: '16px' }}>
                    合计:
                  </Text>
                </Col>
                <Col>
                  <Text strong style={{ fontSize: '18px', color: '#ff4d4f' }}>
                    ¥{(total || 0).toFixed(2)}
                  </Text>
                </Col>
              </Row>
            </div>

            <Space direction="vertical" style={{ width: '100%' }} size="middle">
              <Button
                type="primary"
                size="large"
                block
                icon={<CreditCardOutlined />}
                onClick={handleCheckout}
              >
                立即结算
              </Button>
              
              <Button
                size="large"
                block
                onClick={handleContinueShopping}
              >
                继续购物
              </Button>
            </Space>

            <Divider />

            <div style={{ fontSize: '12px', color: '#999', lineHeight: '1.5' }}>
              <div>• 支持支付宝、微信支付</div>
              <div>• 购买后可立即下载使用</div>
              <div>• 7天无理由退款</div>
              <div>• 终身免费更新</div>
            </div>
          </Card>

          <Card title="推荐插件" style={{ marginTop: '16px' }}>
            <div style={{ textAlign: 'center', color: '#999' }}>
              <Text type="secondary">暂无推荐</Text>
            </div>
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default Cart;