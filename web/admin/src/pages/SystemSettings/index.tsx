import React, { useState, useEffect, useCallback, useMemo } from 'react';
import {
  Card,
  Tabs,
  Form,
  Input,
  Switch,
  Button,
  Select,
  InputNumber,
  Table,
  Space,
  Row,
  Col,
  Statistic,
  Progress,
  Alert,
  Tag,
  Modal,
  Upload,
  Divider,
  Timeline,
  Badge,
  Tooltip,
  DatePicker,
  Radio,
  Slider,
  Typography,
  List,
  Avatar,
  Descriptions,
  Empty,
  App,
} from 'antd';
import {
  SettingOutlined,
  DatabaseOutlined,
  SecurityScanOutlined,
  FileTextOutlined,
  MonitorOutlined,
  CloudServerOutlined,
  ExclamationCircleOutlined,
  CheckCircleOutlined,
  WarningOutlined,
  InfoCircleOutlined,
  DownloadOutlined,
  ReloadOutlined,
  DeleteOutlined,
  SearchOutlined,
  FilterOutlined,
  ExportOutlined,
  ImportOutlined,
  SaveOutlined,
  UndoOutlined,
  EyeOutlined,
  PlusOutlined,
  ClockCircleOutlined,
  BarChartOutlined,
  LineChartOutlined,
  PieChartOutlined,
  DashboardOutlined,
  BugOutlined,
  ThunderboltOutlined,
  HeartOutlined,
  // CpuOutlined,
  HddOutlined,
  WifiOutlined,
} from '@ant-design/icons';
import dayjs from 'dayjs';
import { configAPI, logAPI, metricsAPI, systemAPI } from '@/services/systemService';
import type { SystemConfig as SystemConfigType, SystemLog, SystemMetrics } from '@/services/systemService';


const { Option } = Select;
const { TextArea } = Input;
const { Title, Text, Paragraph } = Typography;
const { RangePicker } = DatePicker;

// 系统配置接口
interface SystemConfig {
  id: string;
  key: string;
  value: string;
  description: string;
  category: string;
  type: 'string' | 'number' | 'boolean' | 'json';
  isPublic: boolean;
  updatedAt: string;
}

// 日志条目接口
interface LogEntry {
  id: string;
  timestamp: string;
  level: 'debug' | 'info' | 'warn' | 'error' | 'fatal';
  message: string;
  module: string;
  userId?: string;
  ip?: string;
  userAgent?: string;
  details?: any;
}

// 系统信息接口
interface SystemInfo {
  systemName: string;
  version: string;
  buildTime: string;
  gitCommit: string;
  environment: string;
  os: string;
  architecture: string;
  goVersion: string;
  nodeVersion: string;
  database: string;
  redis: string;
  nginx: string;
  license: {
    type: string;
    description: string;
    expiryDate: string;
    isValid: boolean;
  };
}

// 性能指标接口
interface PerformanceMetrics {
  cpu: {
    usage: number;
    cores: number;
    loadAverage: number[];
  };
  memory: {
    total: number;
    used: number;
    free: number;
    usage: number;
  };
  disk: {
    total: number;
    used: number;
    free: number;
    usage: number;
  };
  network: {
    bytesIn: number;
    bytesOut: number;
    packetsIn: number;
    packetsOut: number;
  };
  database: {
    connections: number;
    queries: number;
    slowQueries: number;
    avgResponseTime: number;
  };
  uptime: number;
  timestamp: string;
}

const SystemSettings: React.FC = () => {
  const { message } = App.useApp();
  const [activeTab, setActiveTab] = useState('general');
  const [loading, setLoading] = useState(false);
  const [logsLoading, setLogsLoading] = useState(false);
  const [metricsLoading, setMetricsLoading] = useState(false);
  const [configs, setConfigs] = useState<SystemConfig[]>([]);
  const [logs, setLogs] = useState<LogEntry[]>([]);
  const [metrics, setMetrics] = useState<PerformanceMetrics | null>(null);
  const [systemInfo, setSystemInfo] = useState<SystemInfo | null>(null);
  const [systemInfoLoading, setSystemInfoLoading] = useState(false);
  const [configModalVisible, setConfigModalVisible] = useState(false);
  const [editingConfig, setEditingConfig] = useState<SystemConfig | null>(null);
  const [viewingConfig, setViewingConfig] = useState<SystemConfig | null>(null);
  const [configForm] = Form.useForm();
  const [logLevel, setLogLevel] = useState<string>('info');
  const [logModule, setLogModule] = useState<string>('all');
  const [logDateRange, setLogDateRange] = useState<[dayjs.Dayjs, dayjs.Dayjs] | null>(null);
  
  // 缓存相关状态
  const [configsLastFetch, setConfigsLastFetch] = useState<number>(0);
  const [metricsLastFetch, setMetricsLastFetch] = useState<number>(0);
  
  // 缓存有效期（毫秒）
  const CACHE_DURATION = 30000; // 30秒



  // 加载系统配置
  const loadConfigs = useCallback(async (forceRefresh = false) => {
    const now = Date.now();
    
    // 检查缓存是否有效（使用函数式状态检查）
    if (!forceRefresh && (now - configsLastFetch) < CACHE_DURATION) {
      // 检查当前configs状态是否有数据
      let shouldLoadFromAPI = false;
      setConfigs(currentConfigs => {
        if (currentConfigs && currentConfigs.length > 0) {
          console.log('使用缓存的配置数据');
          return currentConfigs; // 返回当前状态，不触发更新
        } else {
          // 如果没有数据，标记需要从API加载
          shouldLoadFromAPI = true;
          return currentConfigs;
        }
      });
      
      if (!shouldLoadFromAPI) {
        return;
      }
    }
    
    // 从API加载数据
    try {
      setLoading(true);
      const data = await configAPI.getConfigs();
      // 确保data是数组，如果为null或undefined则设置为空数组
      setConfigs(Array.isArray(data) ? data : []);
      setConfigsLastFetch(now);
    } catch (err: any) {
      console.error('加载系统配置失败', err);
      message.error('加载系统配置失败');
      // 发生错误时确保configs不为null
      setConfigs([]);
    } finally {
      setLoading(false);
    }
  }, [message, configsLastFetch, CACHE_DURATION]);

  // 加载系统信息
  const loadSystemInfo = useCallback(async () => {
    try {
      setSystemInfoLoading(true);
      const response = await fetch('/api/v1/system/info');
      if (!response.ok) {
        throw new Error('获取系统信息失败');
      }
      const result = await response.json();
      setSystemInfo(result.data);
    } catch (err: any) {
      console.error('加载系统信息失败', err);
      message.error('加载系统信息失败');
    } finally {
      setSystemInfoLoading(false);
    }
  }, [message]);

  // 提交保存配置
  const handleSaveConfigs = async () => {
    try {
      setLoading(true);
      // 批量更新配置
      for (const config of configs) {
        await configAPI.updateConfig(config.id, {
          key: config.key,
          value: config.value,
          description: config.description,
          category: config.category,
          type: config.type,
          isPublic: config.isPublic,
        });
      }
      message.success('配置已保存');
      await loadConfigs();
    } catch (err: any) {
      console.error('保存配置失败', err);
      message.error('保存配置失败: ' + (err.message || '未知错误'));
    } finally {
      setLoading(false);
    }
  };

  // 加载日志
  const loadLogs = useCallback(async () => {
    try {
      setLogsLoading(true);
      const params: any = {
        level: logLevel === 'all' ? undefined : logLevel,
        source: logModule === 'all' ? undefined : logModule,
        startTime: logDateRange?.[0]?.toISOString(),
        endTime: logDateRange?.[1]?.toISOString(),
        page: 1,
        pageSize: 100,
      };
      const res = await logAPI.getLogs(params);
      setLogs(res?.logs || []);
    } catch (err: any) {
      console.error('加载日志失败', err);
      message.error('加载日志失败: ' + (err.message || '未知错误'));
      // 发生错误时确保logs不为null
      setLogs([]);
    } finally {
      setLogsLoading(false);
    }
  }, [logLevel, logModule, logDateRange, message]);

  // 导出日志
  const handleExportLogs = async () => {
    try {
      const params = {
        level: logLevel === 'all' ? undefined : logLevel,
        source: logModule === 'all' ? undefined : logModule,
        startTime: logDateRange?.[0]?.toISOString(),
        endTime: logDateRange?.[1]?.toISOString(),
      };
      const blob = await logAPI.exportLogs(params);
      const url = URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = `system-logs-${dayjs().format('YYYY-MM-DD')}.json`;
      link.click();
      URL.revokeObjectURL(url);
      message.success('日志导出成功');
    } catch (err: any) {
      console.error('导出日志失败', err);
      message.error('导出日志失败');
    }
  };

  // 清理日志
  const handleClearLogs = async () => {
    Modal.confirm({
      title: '确认清理日志',
      content: '此操作将删除符合过滤条件的日志记录，是否继续？',
      onOk: async () => {
        try {
          const params = {
            level: logLevel === 'all' ? undefined : logLevel,
            source: logModule === 'all' ? undefined : logModule,
            beforeTime: logDateRange?.[1]?.toISOString(),
          };
          const result = await logAPI.clearLogs(params);
          message.success(`已清理 ${result.deletedCount} 条日志`);
          await loadLogs();
        } catch (err: any) {
          console.error('清理日志失败', err);
          message.error('清理日志失败');
        }
      },
    });
  };

  // 加载性能指标
  const loadMetrics = useCallback(async (forceRefresh = false) => {
    const now = Date.now();
    
    // 检查缓存是否有效（性能指标缓存时间较短，10秒）
    if (!forceRefresh && (now - metricsLastFetch) < 10000) {
      // 检查当前metrics状态是否有数据
      let shouldLoadFromAPI = false;
      setMetrics(currentMetrics => {
        if (currentMetrics) {
          console.log('使用缓存的性能指标数据');
          return currentMetrics; // 返回当前状态，不触发更新
        } else {
          // 如果没有数据，标记需要从API加载
          shouldLoadFromAPI = true;
          return currentMetrics;
        }
      });
      
      if (!shouldLoadFromAPI) {
        return;
      }
    }
    
    try {
      setMetricsLoading(true);
      const data = await metricsAPI.getCurrentMetrics();
      if (data) {
        setMetrics({
          cpu: { 
            usage: data.cpu?.usage || 0, 
            cores: data.cpu?.cores || 0, 
            loadAverage: data.cpu?.loadAverage || [] 
          },
          memory: { 
            total: data.memory?.total || 0, 
            used: data.memory?.used || 0, 
            free: (data.memory?.total || 0) - (data.memory?.used || 0), 
            usage: data.memory?.usage || 0 
          },
          disk: { 
            total: data.disk?.total || 0, 
            used: data.disk?.used || 0, 
            free: (data.disk?.total || 0) - (data.disk?.used || 0), 
            usage: data.disk?.usage || 0 
          },
          network: { 
            bytesIn: data.network?.bytesIn || 0, 
            bytesOut: data.network?.bytesOut || 0, 
            packetsIn: data.network?.packetsIn || 0, 
            packetsOut: data.network?.packetsOut || 0 
          },
          database: { 
            connections: data.database?.connections || 0, 
            queries: data.database?.queries || 0, 
            slowQueries: data.database?.slowQueries || 0, 
            avgResponseTime: data.database?.avgResponseTime || 0 
          },
          uptime: data.uptime || 0,
          timestamp: data.timestamp || new Date().toISOString(),
        });
        setMetricsLastFetch(now);
      }
    } catch (err: any) {
      console.error('加载性能指标失败', err);
      message.error('加载性能指标失败: ' + (err.message || '未知错误'));
    } finally {
      setMetricsLoading(false);
    }
  }, [message, metricsLastFetch]);

  // 格式化工具（页面展示使用）
  const formatBytes = (bytes: number) => {
    if (bytes === 0 || !Number.isFinite(bytes)) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.min(sizes.length - 1, Math.floor(Math.log(Math.abs(bytes)) / Math.log(k)));
    const value = bytes / Math.pow(k, i);
    return `${value.toFixed(2)} ${sizes[i]}`;
  };

  const formatUptime = (seconds: number) => {
    if (!Number.isFinite(seconds) || seconds < 0) seconds = 0;
    const days = Math.floor(seconds / 86400);
    const hours = Math.floor((seconds % 86400) / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    return `${days}天 ${hours}小时 ${minutes}分钟`;
  };

  // 查看配置详情
  const handleViewConfig = (record: SystemConfig) => {
    setViewingConfig(record);
  };

  // 编辑配置
  const handleEditConfig = (record: SystemConfig) => {
    setEditingConfig(record);
    configForm.setFieldsValue({
      key: record.key,
      value: record.value,
      description: record.description,
      category: record.category,
      type: record.type,
      isPublic: record.isPublic,
    });
    setConfigModalVisible(true);
  };

  // 新增配置
  const handleCreateConfig = () => {
    setEditingConfig(null);
    configForm.resetFields();
    setConfigModalVisible(true);
  };

  // 保存配置（新增或编辑）
  const handleSaveConfig = async (values: any) => {
    try {
      setLoading(true);
      if (editingConfig) {
        // 编辑模式：更新现有配置
        await configAPI.updateConfig(editingConfig.id, values);
        message.success('配置更新成功');
      } else {
        // 新增模式：添加新配置
        await configAPI.createConfig(values);
        message.success('配置创建成功');
      }
      setConfigModalVisible(false);
      configForm.resetFields();
      // 重新加载配置列表
      await loadConfigs();
    } catch (err: any) {
      console.error('保存配置失败', err);
      message.error('保存配置失败');
    } finally {
      setLoading(false);
    }
  };

  // 关闭配置弹窗
  const handleCloseConfigModal = () => {
    setConfigModalVisible(false);
    setEditingConfig(null);
    configForm.resetFields();
  };

  // 关闭查看弹窗
  const handleCloseViewModal = () => {
    setViewingConfig(null);
  };

  useEffect(() => {
    // 页面首次加载时读取配置、日志与指标
    loadConfigs();
    loadLogs();
    loadMetrics();
    loadSystemInfo();
  }, [loadConfigs, loadLogs, loadMetrics, loadSystemInfo]);

  // 使用防抖机制避免频繁请求日志
  useEffect(() => {
    const timer = setTimeout(() => {
      loadLogs();
    }, 1000); // 1000ms 防抖延迟

    return () => clearTimeout(timer);
  }, [logLevel, logModule, logDateRange, loadLogs]);

  // 配置表格列
  const configColumns = [
    {
      title: '配置项',
      dataIndex: 'key',
      key: 'key',
      render: (key: string, record: SystemConfig) => (
        <div>
          <div style={{ fontWeight: 500 }}>{key}</div>
          <div style={{ fontSize: '12px', color: '#999' }}>{record.description}</div>
        </div>
      ),
    },
    {
      title: '当前值',
      dataIndex: 'value',
      key: 'value',
      render: (value: string, record: SystemConfig) => {
        if (record.type === 'boolean') {
          return <Tag color={value === 'true' ? 'green' : 'red'}>{value === 'true' ? '启用' : '禁用'}</Tag>;
        }
        return <Text code>{value}</Text>;
      },
    },
    {
      title: '类型',
      dataIndex: 'type',
      key: 'type',
      render: (type: string) => <Tag>{type}</Tag>,
    },
    {
      title: '分类',
      dataIndex: 'category',
      key: 'category',
      render: (category: string) => {
        const categoryMap: Record<string, { color: string; label: string }> = {
          general: { color: 'blue', label: '常规' },
          security: { color: 'red', label: '安全' },
          database: { color: 'green', label: '数据库' },
          logging: { color: 'orange', label: '日志' },
          features: { color: 'purple', label: '功能' },
          email: { color: 'cyan', label: '邮件' },
        };
        const config = categoryMap[category] || { color: 'default', label: category };
        return <Tag color={config.color}>{config.label}</Tag>;
      },
    },
    {
      title: '可见性',
      dataIndex: 'isPublic',
      key: 'isPublic',
      render: (isPublic: boolean) => (
        <Tag color={isPublic ? 'green' : 'orange'}>
          {isPublic ? '公开' : '私有'}
        </Tag>
      ),
    },
    {
      title: '更新时间',
      dataIndex: 'updatedAt',
      key: 'updatedAt',
      render: (time: string) => dayjs(time).format('YYYY-MM-DD HH:mm:ss'),
    },
    {
      title: '操作',
      key: 'action',
      render: (_, record: SystemConfig) => (
        <Space>
          <Button 
            type="link" 
            size="small" 
            icon={<EyeOutlined />}
            onClick={() => handleViewConfig(record)}
          >
            查看
          </Button>
          <Button 
            type="link" 
            size="small" 
            icon={<SettingOutlined />}
            onClick={() => handleEditConfig(record)}
          >
            编辑
          </Button>
        </Space>
      ),
    },
  ];

  // 日志表格列
  const logColumns = [
    {
      title: '时间',
      dataIndex: 'timestamp',
      key: 'timestamp',
      width: 180,
      render: (time: string) => dayjs(time).format('YYYY-MM-DD HH:mm:ss'),
    },
    {
      title: '级别',
      dataIndex: 'level',
      key: 'level',
      width: 80,
      render: (level: string) => {
        const levelMap: Record<string, { color: string; icon: React.ReactNode }> = {
          debug: { color: 'default', icon: <BugOutlined /> },
          info: { color: 'blue', icon: <InfoCircleOutlined /> },
          warn: { color: 'orange', icon: <WarningOutlined /> },
          error: { color: 'red', icon: <ExclamationCircleOutlined /> },
          fatal: { color: 'red', icon: <ThunderboltOutlined /> },
        };
        const config = levelMap[level] || { color: 'default', icon: <InfoCircleOutlined /> };
        return (
          <Tag color={config.color} icon={config.icon}>
            {level.toUpperCase()}
          </Tag>
        );
      },
    },
    {
      title: '模块',
      dataIndex: 'module',
      key: 'module',
      width: 100,
      render: (module: string) => <Tag>{module}</Tag>,
    },
    {
      title: '消息',
      dataIndex: 'message',
      key: 'message',
      ellipsis: true,
    },
    {
      title: '用户',
      dataIndex: 'userId',
      key: 'userId',
      width: 100,
      render: (userId: string) => userId ? <Text code>{userId}</Text> : '-',
    },
    {
      title: 'IP地址',
      dataIndex: 'ip',
      key: 'ip',
      width: 120,
      render: (ip: string) => ip ? <Text code>{ip}</Text> : '-',
    },
    {
      title: '操作',
      key: 'action',
      width: 100,
      render: (_, record: LogEntry) => (
        <Button
          type="link"
          size="small"
          icon={<EyeOutlined />}
          onClick={() => {
            Modal.info({
              title: '日志详情',
              width: 800,
              content: (
                <div>
                  <Descriptions column={1} bordered size="small">
                    <Descriptions.Item label="时间">
                      {dayjs(record.timestamp).format('YYYY-MM-DD HH:mm:ss')}
                    </Descriptions.Item>
                    <Descriptions.Item label="级别">
                      <Tag color={record.level === 'error' ? 'red' : record.level === 'warn' ? 'orange' : 'blue'}>
                        {record.level.toUpperCase()}
                      </Tag>
                    </Descriptions.Item>
                    <Descriptions.Item label="模块">{record.module}</Descriptions.Item>
                    <Descriptions.Item label="消息">{record.message}</Descriptions.Item>
                    {record.userId && (
                      <Descriptions.Item label="用户ID">{record.userId}</Descriptions.Item>
                    )}
                    {record.ip && (
                      <Descriptions.Item label="IP地址">{record.ip}</Descriptions.Item>
                    )}
                    {record.userAgent && (
                      <Descriptions.Item label="User Agent">{record.userAgent}</Descriptions.Item>
                    )}
                  </Descriptions>
                  {record.details && (
                    <div style={{ marginTop: 16 }}>
                      <Title level={5}>详细信息</Title>
                      <pre style={{ background: '#f5f5f5', padding: 12, borderRadius: 4 }}>
                        {JSON.stringify(record.details, null, 2)}
                      </pre>
                    </div>
                  )}
                </div>
              ),
            });
          }}
        >
          详情
        </Button>
      ),
    },
  ];

  const tabItems = [
    {
      key: 'general',
      label: (
        <span>
          <SettingOutlined />
          系统配置
        </span>
      ),
      children: (
        <>
          <div style={{ marginBottom: 16 }}>
            <Space>
              <Button
                type="primary"
                icon={<PlusOutlined />}
                onClick={handleCreateConfig}
              >
                新增配置
              </Button>
              <Button
                type="primary"
                icon={<SaveOutlined />}
                onClick={() => configForm.submit()}
                loading={loading}
              >
                保存配置
              </Button>
              <Button icon={<ReloadOutlined />} onClick={loadConfigs} loading={loading}>
                刷新
              </Button>
              <Button icon={<ImportOutlined />}>
                导入配置
              </Button>
              <Button icon={<ExportOutlined />}>
                导出配置
              </Button>
            </Space>
          </div>

          <Table
            columns={configColumns}
            dataSource={configs}
            rowKey="id"
            loading={loading}
            pagination={{
              pageSize: 10,
              showSizeChanger: true,
              showQuickJumper: true,
              showTotal: (total) => `共 ${total} 项配置`,
            }}
          />
        </>
      ),
    },
    {
      key: 'logs',
      label: (
        <span>
          <FileTextOutlined />
          日志管理
        </span>
      ),
      children: (
        <>
          <div style={{ marginBottom: 16 }}>
            <Row gutter={16}>
              <Col span={6}>
                <Select
                  placeholder="选择日志级别"
                  style={{ width: '100%' }}
                  value={logLevel}
                  onChange={setLogLevel}
                >
                  <Option value="all">全部级别</Option>
                  <Option value="debug">DEBUG</Option>
                  <Option value="info">INFO</Option>
                  <Option value="warn">WARN</Option>
                  <Option value="error">ERROR</Option>
                  <Option value="fatal">FATAL</Option>
                </Select>
              </Col>
              <Col span={6}>
                <Select
                  placeholder="选择模块"
                  style={{ width: '100%' }}
                  value={logModule}
                  onChange={setLogModule}
                >
                  <Option value="all">全部模块</Option>
                  <Option value="auth">认证</Option>
                  <Option value="database">数据库</Option>
                  <Option value="plugin">插件</Option>
                  <Option value="config">配置</Option>
                  <Option value="api">API</Option>
                </Select>
              </Col>
              <Col span={8}>
                <RangePicker
                  style={{ width: '100%' }}
                  value={logDateRange}
                  onChange={setLogDateRange}
                  placeholder={['开始时间', '结束时间']}
                />
              </Col>
              <Col span={4}>
                <Space>
                  <Button icon={<SearchOutlined />} onClick={loadLogs} loading={logsLoading}>
                    搜索
                  </Button>
                  <Button icon={<ExportOutlined />} onClick={handleExportLogs}>
                    导出
                  </Button>
                  <Button danger icon={<DeleteOutlined />} onClick={handleClearLogs}>
                    清理
                  </Button>
                </Space>
              </Col>
            </Row>
          </div>

          <Table
            columns={logColumns}
            dataSource={logs}
            rowKey="id"
            loading={logsLoading}
            size="small"
            pagination={{
              pageSize: 20,
              showSizeChanger: true,
              showQuickJumper: true,
              showTotal: (total) => `共 ${total} 条日志`,
            }}
          />
        </>
      ),
    },
    {
      key: 'performance',
      label: (
        <span>
          <MonitorOutlined />
          性能监控
        </span>
      ),
      children: (
        <>
          {metrics && (
            <div>
              <Row gutter={16} style={{ marginBottom: 24 }}>
                <Col span={6}>
                  <Card>
                    <Statistic
                      title="CPU使用率"
                      value={metrics.cpu.usage}
                      precision={1}
                      suffix="%"
                      prefix={<DashboardOutlined />}
                      valueStyle={{ color: metrics.cpu.usage > 80 ? '#cf1322' : '#3f8600' }}
                    />
                    <Progress
                      percent={metrics.cpu.usage}
                      size="small"
                      status={metrics.cpu.usage > 80 ? 'exception' : 'active'}
                      style={{ marginTop: 8 }}
                    />
                  </Card>
                </Col>
                <Col span={6}>
                  <Card>
                    <Statistic
                      title="内存使用率"
                      value={metrics.memory.usage}
                      precision={1}
                      suffix="%"
                      prefix={<HddOutlined />}
                      valueStyle={{ color: metrics.memory.usage > 80 ? '#cf1322' : '#3f8600' }}
                    />
                    <Progress
                      percent={metrics.memory.usage}
                      size="small"
                      status={metrics.memory.usage > 80 ? 'exception' : 'active'}
                      style={{ marginTop: 8 }}
                    />
                    <div style={{ fontSize: '12px', color: '#999', marginTop: 4 }}>
                      {formatBytes(metrics.memory.used)} / {formatBytes(metrics.memory.total)}
                    </div>
                  </Card>
                </Col>
                <Col span={6}>
                  <Card>
                    <Statistic
                      title="磁盘使用率"
                      value={metrics.disk.usage}
                      precision={1}
                      suffix="%"
                      prefix={<DatabaseOutlined />}
                      valueStyle={{ color: metrics.disk.usage > 80 ? '#cf1322' : '#3f8600' }}
                    />
                    <Progress
                      percent={metrics.disk.usage}
                      size="small"
                      status={metrics.disk.usage > 80 ? 'exception' : 'active'}
                      style={{ marginTop: 8 }}
                    />
                    <div style={{ fontSize: '12px', color: '#999', marginTop: 4 }}>
                      {formatBytes(metrics.disk.used)} / {formatBytes(metrics.disk.total)}
                    </div>
                  </Card>
                </Col>
                <Col span={6}>
                  <Card>
                    <Statistic
                      title="系统运行时间"
                      value={formatUptime(metrics.uptime)}
                      prefix={<ClockCircleOutlined />}
                      valueStyle={{ color: '#1890ff' }}
                    />
                    <div style={{ fontSize: '12px', color: '#999', marginTop: 8 }}>
                      启动时间: {dayjs().subtract(metrics.uptime, 'second').format('YYYY-MM-DD HH:mm:ss')}
                    </div>
                  </Card>
                </Col>
              </Row>

              <Row gutter={16} style={{ marginBottom: 24 }}>
                <Col span={12}>
                  <Card title="网络流量" size="small">
                    <Row gutter={16}>
                      <Col span={12}>
                        <Statistic
                          title="入站流量"
                          value={formatBytes(metrics.network.bytesIn)}
                          prefix={<WifiOutlined />}
                          valueStyle={{ color: '#52c41a' }}
                        />
                      </Col>
                      <Col span={12}>
                        <Statistic
                          title="出站流量"
                          value={formatBytes(metrics.network.bytesOut)}
                          prefix={<WifiOutlined />}
                          valueStyle={{ color: '#1890ff' }}
                        />
                      </Col>
                    </Row>
                    <Divider />
                    <Row gutter={16}>
                      <Col span={12}>
                        <Text type="secondary">入站数据包: {metrics.network.packetsIn.toLocaleString()}</Text>
                      </Col>
                      <Col span={12}>
                        <Text type="secondary">出站数据包: {metrics.network.packetsOut.toLocaleString()}</Text>
                      </Col>
                    </Row>
                  </Card>
                </Col>
                <Col span={12}>
                  <Card title="数据库性能" size="small">
                    <Row gutter={16}>
                      <Col span={12}>
                        <Statistic
                          title="活跃连接"
                          value={metrics.database.connections}
                          prefix={<DatabaseOutlined />}
                          valueStyle={{ color: '#722ed1' }}
                        />
                      </Col>
                      <Col span={12}>
                        <Statistic
                          title="查询总数"
                          value={metrics.database.queries}
                          prefix={<BarChartOutlined />}
                          valueStyle={{ color: '#fa8c16' }}
                        />
                      </Col>
                    </Row>
                    <Divider />
                    <Row gutter={16}>
                      <Col span={12}>
                        <Text type="secondary">慢查询: {metrics.database.slowQueries}</Text>
                      </Col>
                      <Col span={12}>
                        <Text type="secondary">平均响应: {metrics.database.avgResponseTime}ms</Text>
                      </Col>
                    </Row>
                  </Card>
                </Col>
              </Row>

              <Card title="系统信息" size="small">
                <Descriptions column={3} bordered>
                  <Descriptions.Item label="CPU核心数">{metrics.cpu.cores}</Descriptions.Item>
                  <Descriptions.Item label="负载均衡">
                    {metrics.cpu.loadAverage && Array.isArray(metrics.cpu.loadAverage) 
                      ? metrics.cpu.loadAverage.map((load, index) => (
                          <Tag key={index} color="blue">{load.toFixed(2)}</Tag>
                        ))
                      : <Tag color="gray">暂无数据</Tag>
                    }
                  </Descriptions.Item>
                  <Descriptions.Item label="最后更新">
                    {dayjs(metrics.timestamp).format('YYYY-MM-DD HH:mm:ss')}
                  </Descriptions.Item>
                </Descriptions>
              </Card>

              <div style={{ marginTop: 16, textAlign: 'center' }}>
                <Button
                  type="primary"
                  icon={<ReloadOutlined />}
                  onClick={() => loadMetrics(true)}
                  loading={metricsLoading}
                >
                  刷新监控数据
                </Button>
              </div>
            </div>
          )}
        </>
      ),
    },
    {
      key: 'info',
      label: (
        <span>
          <CloudServerOutlined />
          系统信息
        </span>
      ),
      children: (
        <>
          <Row gutter={16}>
            <Col span={12}>
              <Card title="系统版本" size="small" loading={systemInfoLoading}>
                <Descriptions column={1} bordered>
                  <Descriptions.Item label="系统名称">
                    {systemInfo?.systemName || '加载中...'}
                  </Descriptions.Item>
                  <Descriptions.Item label="版本号">
                    {systemInfo?.version || '加载中...'}
                  </Descriptions.Item>
                  <Descriptions.Item label="构建时间">
                    {systemInfo?.buildTime || '加载中...'}
                  </Descriptions.Item>
                  <Descriptions.Item label="Git提交">
                    {systemInfo?.gitCommit || '加载中...'}
                  </Descriptions.Item>
                  <Descriptions.Item label="环境">
                    {systemInfo?.environment || '加载中...'}
                  </Descriptions.Item>
                </Descriptions>
              </Card>
            </Col>
            <Col span={12}>
              <Card title="运行环境" size="small" loading={systemInfoLoading}>
                <Descriptions column={1} bordered>
                  <Descriptions.Item label="操作系统">
                    {systemInfo?.os || '加载中...'}
                  </Descriptions.Item>
                  <Descriptions.Item label="Go版本">
                    {systemInfo?.goVersion || '加载中...'}
                  </Descriptions.Item>
                  <Descriptions.Item label="Node.js版本">
                    {systemInfo?.nodeVersion || '加载中...'}
                  </Descriptions.Item>
                  <Descriptions.Item label="数据库">
                    {systemInfo?.database || '加载中...'}
                  </Descriptions.Item>
                  <Descriptions.Item label="Redis版本">
                    {systemInfo?.redis || '加载中...'}
                  </Descriptions.Item>
                </Descriptions>
              </Card>
            </Col>
          </Row>

          <Card title="许可证信息" style={{ marginTop: 16 }} size="small" loading={systemInfoLoading}>
            {systemInfo?.license ? (
              <Alert
                message={systemInfo.license.type === 'commercial' ? '商业许可证' : '开源许可证'}
                description={`${systemInfo.license.description} 许可证有效期至 ${systemInfo.license.expiryDate}。`}
                type={systemInfo.license.isValid ? 'info' : 'warning'}
                showIcon
              />
            ) : (
              <Alert
                message="加载中..."
                description="正在获取许可证信息..."
                type="info"
                showIcon
              />
            )}
          </Card>
        </>
      ),
    },
  ];

  return (
    <div>
      <Card>
        <Tabs activeKey={activeTab} onChange={setActiveTab} items={tabItems} />
      </Card>

      {/* 配置编辑弹窗 */}
      <Modal
        title={editingConfig ? '编辑配置' : '新增配置'}
        open={configModalVisible}
        onCancel={handleCloseConfigModal}
        footer={null}
        width={600}
      >
        <Form
          form={configForm}
          layout="vertical"
          onFinish={handleSaveConfig}
        >
          <Form.Item
            label="配置键"
            name="key"
            rules={[{ required: true, message: '请输入配置键' }]}
          >
            <Input placeholder="例如：app.name" />
          </Form.Item>
          
          <Form.Item
            label="配置值"
            name="value"
            rules={[{ required: true, message: '请输入配置值' }]}
          >
            <Input.TextArea rows={3} placeholder="请输入配置值" />
          </Form.Item>
          
          <Form.Item
            label="描述"
            name="description"
          >
            <Input placeholder="请输入配置描述" />
          </Form.Item>
          
          <Form.Item
            label="分类"
            name="category"
            rules={[{ required: true, message: '请选择配置分类' }]}
          >
            <Select placeholder="请选择分类">
              <Select.Option value="system">系统配置</Select.Option>
              <Select.Option value="database">数据库配置</Select.Option>
              <Select.Option value="cache">缓存配置</Select.Option>
              <Select.Option value="security">安全配置</Select.Option>
              <Select.Option value="performance">性能配置</Select.Option>
              <Select.Option value="other">其他配置</Select.Option>
            </Select>
          </Form.Item>
          
          <Form.Item
            label="类型"
            name="type"
            rules={[{ required: true, message: '请选择配置类型' }]}
          >
            <Select placeholder="请选择类型">
              <Select.Option value="string">字符串</Select.Option>
              <Select.Option value="number">数字</Select.Option>
              <Select.Option value="boolean">布尔值</Select.Option>
              <Select.Option value="json">JSON</Select.Option>
            </Select>
          </Form.Item>
          
          <Form.Item
            label="是否公开"
            name="isPublic"
            valuePropName="checked"
          >
            <Switch checkedChildren="公开" unCheckedChildren="私有" />
          </Form.Item>
          
          <Form.Item style={{ marginBottom: 0, textAlign: 'right' }}>
            <Space>
              <Button onClick={handleCloseConfigModal}>
                取消
              </Button>
              <Button type="primary" htmlType="submit" loading={loading}>
                {editingConfig ? '更新' : '创建'}
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>

      {/* 配置查看弹窗 */}
      <Modal
        title="配置详情"
        open={!!viewingConfig}
        onCancel={handleCloseViewModal}
        footer={[
          <Button key="close" onClick={handleCloseViewModal}>
            关闭
          </Button>
        ]}
        width={600}
      >
        {viewingConfig && (
          <Descriptions column={1} bordered>
            <Descriptions.Item label="配置键">{viewingConfig.key}</Descriptions.Item>
            <Descriptions.Item label="配置值">
              <Typography.Text copyable>{viewingConfig.value}</Typography.Text>
            </Descriptions.Item>
            <Descriptions.Item label="描述">{viewingConfig.description || '无'}</Descriptions.Item>
            <Descriptions.Item label="分类">
              <Tag color="blue">{viewingConfig.category}</Tag>
            </Descriptions.Item>
            <Descriptions.Item label="类型">
              <Tag color="green">{viewingConfig.type}</Tag>
            </Descriptions.Item>
            <Descriptions.Item label="是否公开">
              <Tag color={viewingConfig.isPublic ? 'success' : 'warning'}>
                {viewingConfig.isPublic ? '公开' : '私有'}
              </Tag>
            </Descriptions.Item>
            <Descriptions.Item label="更新时间">
              {dayjs(viewingConfig.updatedAt).format('YYYY-MM-DD HH:mm:ss')}
            </Descriptions.Item>
          </Descriptions>
        )}
      </Modal>
    </div>
  );
};

export default SystemSettings;