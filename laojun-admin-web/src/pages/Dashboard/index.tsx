import { useState, useEffect } from 'react';
import { Row, Col, Card, Statistic, Table, Progress, Tag, Space, Button } from 'antd';
import {
  UserOutlined,
  TeamOutlined,
  AppstoreOutlined,
  SettingOutlined,
  ArrowUpOutlined,
  ArrowDownOutlined,
  ReloadOutlined,
} from '@ant-design/icons';
import { useAuthStore } from '@/stores/auth';
import { useTranslation } from '@/locales';
import LanguageTest from '../../components/LanguageTest';

interface SystemStats {
  userCount: number;
  roleCount: number;
  pluginCount: number;
  activePlugins: number;
}

interface RecentActivity {
  id: string;
  type: 'login' | 'plugin' | 'user' | 'system';
  description: string;
  time: string;
  status: 'success' | 'warning' | 'error';
}

const Dashboard: React.FC = () => {
  const { user } = useAuthStore();
  const { t } = useTranslation();
  const [loading, setLoading] = useState(false);
  const [stats, setStats] = useState<SystemStats>({
    userCount: 156,
    roleCount: 8,
    pluginCount: 24,
    activePlugins: 18,
  });

  const [activities] = useState<RecentActivity[]>([
    {
      id: '1',
      type: 'login',
      description: '用户 admin 登录系统',
      time: '2024-01-15 10:30:00',
      status: 'success',
    },
    {
      id: '2',
      type: 'plugin',
      description: '插件 "用户管理增强" 已启用',
      time: '2024-01-15 09:45:00',
      status: 'success',
    },
    {
      id: '3',
      type: 'user',
      description: '新增用户 "张三"',
      time: '2024-01-15 09:20:00',
      status: 'success',
    },
    {
      id: '4',
      type: 'system',
      description: '系统配置已更新',
      time: '2024-01-15 08:55:00',
      status: 'warning',
    },
    {
      id: '5',
      type: 'plugin',
      description: '插件 "数据导出" 安装失败',
      time: '2024-01-15 08:30:00',
      status: 'error',
    },
  ]);

  // 刷新数据
  const refreshData = async () => {
    setLoading(true);
    try {
      // 模拟 API 调用
      await new Promise(resolve => setTimeout(resolve, 1000));
      
      // 更新统计数据
      setStats(prev => ({
        ...prev,
        userCount: prev.userCount + Math.floor(Math.random() * 5),
        activePlugins: Math.min(prev.pluginCount, prev.activePlugins + Math.floor(Math.random() * 3)),
      }));
    } catch (error) {
      console.error('Failed to refresh data:', error);
    } finally {
      setLoading(false);
    }
  };

  // 活动表格列配置
  const activityColumns = [
    {
      title: t('dashboard.activityType'),
      dataIndex: 'type',
      key: 'type',
      width: 80,
      render: (type: string) => {
        const typeMap = {
          login: { icon: <UserOutlined />, color: 'blue', text: t('dashboard.login') },
          plugin: { icon: <AppstoreOutlined />, color: 'green', text: t('dashboard.plugin') },
          user: { icon: <TeamOutlined />, color: 'purple', text: t('dashboard.user') },
          system: { icon: <SettingOutlined />, color: 'orange', text: t('dashboard.system') },
        };
        const config = typeMap[type as keyof typeof typeMap];
        return (
          <Tag icon={config.icon} color={config.color}>
            {config.text}
          </Tag>
        );
      },
    },
    {
      title: t('dashboard.description'),
      dataIndex: 'description',
      key: 'description',
    },
    {
      title: t('dashboard.time'),
      dataIndex: 'time',
      key: 'time',
      width: 160,
    },
    {
      title: t('dashboard.status'),
      dataIndex: 'status',
      key: 'status',
      width: 80,
      render: (status: string) => {
        const statusMap = {
          success: { color: 'success', text: t('dashboard.success') },
          warning: { color: 'warning', text: t('dashboard.warning') },
          error: { color: 'error', text: t('dashboard.error') },
        };
        const config = statusMap[status as keyof typeof statusMap];
        return <Tag color={config.color}>{config.text}</Tag>;
      },
    },
  ];

  return (
    <div>
      {/* 欢迎信息 */}
      <Card style={{ marginBottom: '24px' }}>
        <Row justify="space-between" align="middle">
          <Col>
            <Space direction="vertical" size={4}>
              <h2 style={{ margin: 0 }}>
                {t('dashboard.welcome', { name: '管理员' })}
              </h2>
              <p style={{ margin: 0, color: '#666' }}>
                {t('dashboard.today')} {new Date().toLocaleDateString('zh-CN', { 
                  year: 'numeric', 
                  month: 'long', 
                  day: 'numeric',
                  weekday: 'long'
                })}
              </p>
            </Space>
          </Col>
          <Col>
            <Button 
              icon={<ReloadOutlined />} 
              onClick={refreshData}
              loading={loading}
            >
              {t('dashboard.refreshData')}
            </Button>
          </Col>
        </Row>
      </Card>

      {/* 统计卡片 */}
      <Row gutter={[16, 16]} style={{ marginBottom: '24px' }}>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title={t('dashboard.totalUsers')}
              value={stats.userCount}
              prefix={<UserOutlined />}
              valueStyle={{ color: '#3f8600' }}
              suffix={
                <span style={{ fontSize: '14px' }}>
                  <ArrowUpOutlined /> 12%
                </span>
              }
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title={t('dashboard.roleCount')}
              value={stats.roleCount}
              prefix={<TeamOutlined />}
              valueStyle={{ color: '#1890ff' }}
              suffix={
                <span style={{ fontSize: '14px' }}>
                  <ArrowUpOutlined /> 3%
                </span>
              }
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title={t('dashboard.totalPlugins')}
              value={stats.pluginCount}
              prefix={<AppstoreOutlined />}
              valueStyle={{ color: '#722ed1' }}
              suffix={
                <span style={{ fontSize: '14px' }}>
                  <ArrowDownOutlined /> 2%
                </span>
              }
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title={t('dashboard.activePlugins')}
              value={stats.activePlugins}
              prefix={<SettingOutlined />}
              valueStyle={{ color: '#fa8c16' }}
            />
            <Progress
              percent={Math.round((stats.activePlugins / stats.pluginCount) * 100)}
              size="small"
              style={{ marginTop: '8px' }}
            />
          </Card>
        </Col>
      </Row>

      {/* 系统状态和最近活动 */}
      <Row gutter={[16, 16]}>
        <Col xs={24} lg={12}>
          <Card title={t('dashboard.systemStatus')} style={{ height: '400px' }}>
            <Space direction="vertical" style={{ width: '100%' }} size="large">
              <div>
                <div style={{ marginBottom: '8px' }}>{t('dashboard.cpuUsage')}</div>
                <Progress percent={45} status="active" />
              </div>
              <div>
                <div style={{ marginBottom: '8px' }}>{t('dashboard.memoryUsage')}</div>
                <Progress percent={67} status="active" />
              </div>
              <div>
                <div style={{ marginBottom: '8px' }}>{t('dashboard.diskUsage')}</div>
                <Progress percent={23} />
              </div>
              <div>
                <div style={{ marginBottom: '8px' }}>{t('dashboard.networkStatus')}</div>
                <Progress percent={89} status="active" />
              </div>
            </Space>
          </Card>
        </Col>
        <Col xs={24} lg={12}>
          <Card title={t('dashboard.recentActivity')} style={{ height: '400px' }}>
            <Table
              dataSource={activities}
              columns={activityColumns}
              pagination={false}
              size="small"
              scroll={{ y: 280 }}
              rowKey="id"
            />
          </Card>
        </Col>
      </Row>

      {/* 语言切换测试组件 */}
      <LanguageTest />
    </div>
  );
};

export default Dashboard;