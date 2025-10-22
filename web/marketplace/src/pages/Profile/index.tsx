import React, { useEffect, useState } from 'react';
import {
  Card,
  Tabs,
  List,
  Avatar,
  Button,
  Rate,
  Tag,
  Empty,
  Typography,
  Space,
  Statistic,
  Row,
  Col,
  Badge,
} from 'antd';
import {
  UserOutlined,
  HeartOutlined,
  DownloadOutlined,
  StarOutlined,
  SettingOutlined,
  ShoppingOutlined,
} from '@ant-design/icons';
import { pluginService } from '@/services/plugin';
import { userService } from '@/services/user';
import { request } from '@/services/api';
import { useAuthStore } from '@/stores/auth';
import { useNavigate } from 'react-router-dom';
import type { User, Plugin, Purchase } from '@/types';

const { Title, Text } = Typography;
const { TabPane } = Tabs;

// 移除模拟数据，使用真实数据
const Profile: React.FC = () => {
  const navigate = useNavigate();
  const { user: authUser, isAuthenticated } = useAuthStore();
  const [activeTab, setActiveTab] = useState('installed');
  const [user, setUser] = useState<User | null>(null);
  const [favorites, setFavorites] = useState<Plugin[]>([]);
  const [purchases, setPurchases] = useState<Purchase[]>([]);
  const [favoriteTotal, setFavoriteTotal] = useState(0);
  const [purchaseTotal, setPurchaseTotal] = useState(0);

  // 如果未登录，重定向到登录页面
  useEffect(() => {
    if (!isAuthenticated) {
      navigate('/login', { state: { from: { pathname: '/profile' } } });
      return;
    }
  }, [isAuthenticated, navigate]);

  useEffect(() => {
    const loadData = async () => {
      try {
        // 获取用户资料
        const profile = await userService.getProfile();
        setUser({
          id: profile.id || '',
          username: profile.username || '用户',
          email: profile.email || '',
          name: profile.full_name || profile.username || '',
          avatar: profile.avatar || '',
          joinedAt: profile.created_at || '',
        });

        // 获取收藏列表
        const favRes = await userService.getFavoritePlugins();
        setFavorites(favRes.data || []);
        setFavoriteTotal(favRes.total || (favRes.data?.length ?? 0));

        // 获取购买记录
        const purchaseRes = await userService.getPurchases();
        setPurchases(purchaseRes.data || []);
        setPurchaseTotal(purchaseRes.total || (purchaseRes.data?.length ?? 0));
      } catch (e) {
        // 静默失败，页面展示 Empty
        console.error('加载个人数据失败', e);
      }
    };

    loadData();
  }, []);

  const renderPurchasedPlugins = () => (
    <List
      dataSource={purchases}
      renderItem={(purchase) => {
        const plugin = (purchase as any).plugin as Plugin | undefined;
        return (
          <List.Item
            actions={[
              <Button key="view" type="link">查看</Button>,
            ]}
          >
            <List.Item.Meta
              avatar={<Avatar src={plugin?.icon} icon={<UserOutlined />} />}
              title={plugin?.name || `插件 #${purchase.pluginId}`}
              description={
                <div>
                  <Text type="secondary">订单时间: {purchase.createdAt}</Text>
                  <br />
                  {plugin ? (
                    <Space>
                      <Rate disabled value={plugin.rating} />
                      <Text type="secondary">{plugin.downloads} 下载</Text>
                      <Text strong style={{ color: plugin.price === 0 ? '#52c41a' : '#ff4d4f' }}>
                        {plugin.price === 0 ? '免费' : `¥${plugin.price}`}
                      </Text>
                    </Space>
                  ) : null}
                </div>
              }
            />
          </List.Item>
        );
      }}
    />
  );

  const renderFavorites = () => (
    <List
      dataSource={favorites}
      renderItem={(plugin) => (
        <List.Item
          actions={[
            <Button key="view" type="link">查看</Button>,
            <Button 
              key="unfavorite" 
              type="link" 
              danger 
              onClick={async () => {
                try {
                  await userService.toggleFavorite(plugin.id);
                  // 重新加载收藏列表
                  const favRes = await userService.getFavoritePlugins();
                  setFavorites(favRes.data || []);
                  setFavoriteTotal(favRes.total || (favRes.data?.length ?? 0));
                } catch (error) {
                  console.error('取消收藏失败', error);
                }
              }}
            >
              取消收藏
            </Button>,
          ]}
        >
          <List.Item.Meta
            avatar={<Avatar src={plugin.icon} icon={<UserOutlined />} />}
            title={plugin.name}
            description={
              <div>
                <Text type="secondary">{plugin.description}</Text>
                <br />
                <Space>
                  <Rate disabled value={plugin.rating} />
                  <Text type="secondary">{plugin.downloads} 下载</Text>
                  <Text strong style={{ color: plugin.price === 0 ? '#52c41a' : '#ff4d4f' }}>
                    {plugin.price === 0 ? '免费' : `¥${plugin.price}`}
                  </Text>
                </Space>
              </div>
            }
          />
        </List.Item>
      )}
    />
  );

  const renderReviews = () => (
    <Empty description="还没有发表任何评价" />
  );

  // 如果未登录，不渲染内容（会被重定向到登录页面）
  if (!isAuthenticated) {
    return null;
  }

  return (
    <div style={{ padding: '24px', maxWidth: '1200px', margin: '0 auto' }}>
      <Row gutter={[24, 24]}>
        {/* 左侧用户信息 */}
        <Col xs={24} lg={8}>
          <Card>
            <div style={{ textAlign: 'center', marginBottom: '24px' }}>
              <Avatar size={80} src={user?.avatar} icon={<UserOutlined />} />
              <Title level={3} style={{ marginTop: '16px', marginBottom: '8px' }}>
                {user?.username || '未登录'}
              </Title>
              <Text type="secondary">{user?.email || ''}</Text>
              <div style={{ marginTop: '8px' }}>
                <Text type="secondary">加入时间: {user?.joinedAt || '-'}</Text>
              </div>
            </div>

            <Row gutter={[16, 16]}>
              <Col span={8}>
                <Statistic
                  title="购买"
                  value={purchaseTotal}
                  prefix={<ShoppingOutlined />}
                  valueStyle={{ fontSize: '18px' }}
                />
              </Col>
              <Col span={8}>
                <Statistic
                  title="收藏"
                  value={favoriteTotal}
                  prefix={<HeartOutlined />}
                  valueStyle={{ fontSize: '18px' }}
                />
              </Col>
              <Col span={8}>
                <Statistic
                  title="评价"
                  value={0}
                  prefix={<StarOutlined />}
                  valueStyle={{ fontSize: '18px' }}
                />
              </Col>
            </Row>

            <div style={{ marginTop: '24px' }}>
              <Button
                type="primary"
                block
                icon={<SettingOutlined />}
                style={{ marginBottom: '12px' }}
              >
                编辑资料
              </Button>
              <Button block icon={<ShoppingOutlined />}>
                购买历史
              </Button>
            </div>
          </Card>
        </Col>

        {/* 右侧内容区域 */}
        <Col xs={24} lg={16}>
          <Card>
            <Tabs activeKey={activeTab} onChange={setActiveTab}>
              <TabPane
                tab={
                  <span>
                    <ShoppingOutlined />
                    已购买插件 ({purchaseTotal})
                  </span>
                }
                key="installed"
              >
                {purchases.length > 0 ? (
                  renderPurchasedPlugins()
                ) : (
                  <Empty description="还没有购买任何插件" />
                )}
              </TabPane>

              <TabPane
                tab={
                  <span>
                    <HeartOutlined />
                    我的收藏 ({favoriteTotal})
                  </span>
                }
                key="favorites"
              >
                {favorites.length > 0 ? (
                  renderFavorites()
                ) : (
                  <Empty description="还没有收藏任何插件" />
                )}
              </TabPane>

              <TabPane
                tab={
                  <span>
                    <StarOutlined />
                    我的评价 (0)
                  </span>
                }
                key="reviews"
              >
                {renderReviews()}
              </TabPane>
            </Tabs>
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default Profile;