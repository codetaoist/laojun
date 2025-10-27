import React, { useState, useEffect } from 'react';
import {
  Card,
  Table,
  Button,
  Space,
  Tag,
  Modal,
  Form,
  Input,
  Select,
  message,
  Tabs,
  Statistic,
  Row,
  Col,
  DatePicker,
  Tooltip,
  Popconfirm,
  Badge,
  Typography,
  Divider,
  Timeline,
  Rate,
  Progress
} from 'antd';
import {
  EyeOutlined,
  CheckOutlined,
  CloseOutlined,
  ClockCircleOutlined,
  ExclamationCircleOutlined,
  HistoryOutlined,
  UserOutlined,
  FileTextOutlined,
  BarChartOutlined,
  ReloadOutlined,
  FilterOutlined
} from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import dayjs from 'dayjs';
import { pluginReviewService } from '../../services/pluginReview';

const { TabPane } = Tabs;
const { TextArea } = Input;
const { Option } = Select;
const { Title, Text, Paragraph } = Typography;
const { RangePicker } = DatePicker;

interface Plugin {
  id: string;
  name: string;
  version: string;
  developer: string;
  description: string;
  reviewStatus: 'pending' | 'in_review' | 'approved' | 'rejected';
  reviewPriority: 'high' | 'medium' | 'low';
  submittedAt: string;
  reviewedAt?: string;
  reviewerId?: string;
  reviewerName?: string;
  autoReviewScore?: number;
  rejectionReason?: string;
  appealCount: number;
}

interface ReviewStats {
  totalPlugins: number;
  pendingReviews: number;
  approvedToday: number;
  rejectedToday: number;
  averageReviewTime: number;
  reviewerWorkload: Array<{
    reviewerId: string;
    reviewerName: string;
    assignedCount: number;
    completedCount: number;
  }>;
}

interface Appeal {
  id: string;
  pluginId: string;
  pluginName: string;
  reason: string;
  status: 'pending' | 'in_review' | 'approved' | 'rejected';
  createdAt: string;
  processedAt?: string;
  processorName?: string;
}

const PluginReviewManagement: React.FC = () => {
  const [plugins, setPlugins] = useState<Plugin[]>([]);
  const [appeals, setAppeals] = useState<Appeal[]>([]);
  const [stats, setStats] = useState<ReviewStats | null>(null);
  const [loading, setLoading] = useState(false);
  const [reviewModalVisible, setReviewModalVisible] = useState(false);
  const [appealModalVisible, setAppealModalVisible] = useState(false);
  const [historyModalVisible, setHistoryModalVisible] = useState(false);
  const [selectedPlugin, setSelectedPlugin] = useState<Plugin | null>(null);
  const [selectedAppeal, setAppeal] = useState<Appeal | null>(null);
  const [reviewHistory, setReviewHistory] = useState<any[]>([]);
  const [activeTab, setActiveTab] = useState('queue');
  const [filters, setFilters] = useState({
    status: '',
    priority: '',
    dateRange: null as any,
    reviewer: ''
  });

  const [reviewForm] = Form.useForm();
  const [appealForm] = Form.useForm();

  // 获取审核队列
  const fetchReviewQueue = async () => {
    setLoading(true);
    try {
      const response = await pluginReviewService.getReviewQueue(filters);
      setPlugins(response.data);
    } catch (error) {
      message.error('获取审核队列失败');
    } finally {
      setLoading(false);
    }
  };

  // 获取申诉列表
  const fetchAppeals = async () => {
    setLoading(true);
    try {
      const response = await pluginReviewService.getAppeals();
      setAppeals(response.data);
    } catch (error) {
      message.error('获取申诉列表失败');
    } finally {
      setLoading(false);
    }
  };

  // 获取统计数据
  const fetchStats = async () => {
    try {
      const response = await pluginReviewService.getReviewStats();
      setStats(response.data);
    } catch (error) {
      message.error('获取统计数据失败');
    }
  };

  useEffect(() => {
    fetchReviewQueue();
    fetchStats();
  }, [filters]);

  useEffect(() => {
    if (activeTab === 'appeals') {
      fetchAppeals();
    }
  }, [activeTab]);

  // 分配审核员
  const handleAssignReviewer = async (pluginId: string, reviewerId: string) => {
    try {
      await pluginReviewService.assignReviewer(pluginId, reviewerId);
      message.success('分配审核员成功');
      fetchReviewQueue();
    } catch (error) {
      message.error('分配审核员失败');
    }
  };

  // 审核插件
  const handleReviewPlugin = async (values: any) => {
    if (!selectedPlugin) return;

    try {
      await pluginReviewService.reviewPlugin(selectedPlugin.id, {
        result: values.result,
        notes: values.notes,
        rejectionReason: values.rejectionReason
      });
      message.success('审核完成');
      setReviewModalVisible(false);
      reviewForm.resetFields();
      fetchReviewQueue();
      fetchStats();
    } catch (error) {
      message.error('审核失败');
    }
  };

  // 处理申诉
  const handleProcessAppeal = async (values: any) => {
    if (!selectedAppeal) return;

    try {
      await pluginReviewService.processAppeal(selectedAppeal.id, {
        decision: values.decision,
        notes: values.notes
      });
      message.success('申诉处理完成');
      setAppealModalVisible(false);
      appealForm.resetFields();
      fetchAppeals();
    } catch (error) {
      message.error('申诉处理失败');
    }
  };

  // 查看审核历史
  const handleViewHistory = async (pluginId: string) => {
    try {
      const response = await pluginReviewService.getPluginReviewHistory(pluginId);
      setReviewHistory(response.data);
      setHistoryModalVisible(true);
    } catch (error) {
      message.error('获取审核历史失败');
    }
  };

  // 自动审核
  const handleAutoReview = async (pluginId: string) => {
    try {
      await pluginReviewService.autoReviewPlugin(pluginId);
      message.success('自动审核已启动');
      fetchReviewQueue();
    } catch (error) {
      message.error('自动审核失败');
    }
  };

  // 批量审核
  const handleBatchReview = async (pluginIds: string[], result: string) => {
    try {
      await pluginReviewService.batchReviewPlugins({
        pluginIds,
        result,
        notes: '批量审核'
      });
      message.success('批量审核完成');
      fetchReviewQueue();
      fetchStats();
    } catch (error) {
      message.error('批量审核失败');
    }
  };

  const getStatusColor = (status: string) => {
    const colors = {
      pending: 'orange',
      in_review: 'blue',
      approved: 'green',
      rejected: 'red'
    };
    return colors[status as keyof typeof colors] || 'default';
  };

  const getPriorityColor = (priority: string) => {
    const colors = {
      high: 'red',
      medium: 'orange',
      low: 'green'
    };
    return colors[priority as keyof typeof colors] || 'default';
  };

  const pluginColumns: ColumnsType<Plugin> = [
    {
      title: '插件信息',
      key: 'plugin',
      render: (_, record) => (
        <div>
          <div style={{ fontWeight: 'bold' }}>{record.name}</div>
          <div style={{ fontSize: '12px', color: '#666' }}>
            v{record.version} | {record.developer}
          </div>
        </div>
      ),
    },
    {
      title: '审核状态',
      dataIndex: 'reviewStatus',
      key: 'reviewStatus',
      render: (status) => (
        <Tag color={getStatusColor(status)}>
          {status === 'pending' && '待审核'}
          {status === 'in_review' && '审核中'}
          {status === 'approved' && '已通过'}
          {status === 'rejected' && '已拒绝'}
        </Tag>
      ),
      filters: [
        { text: '待审核', value: 'pending' },
        { text: '审核中', value: 'in_review' },
        { text: '已通过', value: 'approved' },
        { text: '已拒绝', value: 'rejected' },
      ],
    },
    {
      title: '优先级',
      dataIndex: 'reviewPriority',
      key: 'reviewPriority',
      render: (priority) => (
        <Tag color={getPriorityColor(priority)}>
          {priority === 'high' && '高'}
          {priority === 'medium' && '中'}
          {priority === 'low' && '低'}
        </Tag>
      ),
    },
    {
      title: '自动评分',
      dataIndex: 'autoReviewScore',
      key: 'autoReviewScore',
      render: (score) => score ? (
        <div>
          <Progress 
            percent={score} 
            size="small" 
            status={score >= 80 ? 'success' : score >= 60 ? 'normal' : 'exception'}
          />
          <Text style={{ fontSize: '12px' }}>{score}/100</Text>
        </div>
      ) : '-',
    },
    {
      title: '提交时间',
      dataIndex: 'submittedAt',
      key: 'submittedAt',
      render: (date) => dayjs(date).format('YYYY-MM-DD HH:mm'),
      sorter: (a, b) => dayjs(a.submittedAt).unix() - dayjs(b.submittedAt).unix(),
    },
    {
      title: '审核员',
      dataIndex: 'reviewerName',
      key: 'reviewerName',
      render: (name) => name || '-',
    },
    {
      title: '申诉次数',
      dataIndex: 'appealCount',
      key: 'appealCount',
      render: (count) => count > 0 ? <Badge count={count} /> : '-',
    },
    {
      title: '操作',
      key: 'actions',
      render: (_, record) => (
        <Space>
          <Tooltip title="查看详情">
            <Button 
              type="text" 
              icon={<EyeOutlined />}
              onClick={() => {
                setSelectedPlugin(record);
                setReviewModalVisible(true);
              }}
            />
          </Tooltip>
          <Tooltip title="查看历史">
            <Button 
              type="text" 
              icon={<HistoryOutlined />}
              onClick={() => handleViewHistory(record.id)}
            />
          </Tooltip>
          {record.reviewStatus === 'pending' && (
            <Tooltip title="自动审核">
              <Button 
                type="text" 
                icon={<ReloadOutlined />}
                onClick={() => handleAutoReview(record.id)}
              />
            </Tooltip>
          )}
        </Space>
      ),
    },
  ];

  const appealColumns: ColumnsType<Appeal> = [
    {
      title: '插件名称',
      dataIndex: 'pluginName',
      key: 'pluginName',
    },
    {
      title: '申诉原因',
      dataIndex: 'reason',
      key: 'reason',
      ellipsis: true,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status) => (
        <Tag color={getStatusColor(status)}>
          {status === 'pending' && '待处理'}
          {status === 'in_review' && '处理中'}
          {status === 'approved' && '已通过'}
          {status === 'rejected' && '已拒绝'}
        </Tag>
      ),
    },
    {
      title: '提交时间',
      dataIndex: 'createdAt',
      key: 'createdAt',
      render: (date) => dayjs(date).format('YYYY-MM-DD HH:mm'),
    },
    {
      title: '处理人',
      dataIndex: 'processorName',
      key: 'processorName',
      render: (name) => name || '-',
    },
    {
      title: '操作',
      key: 'actions',
      render: (_, record) => (
        <Space>
          <Button 
            type="primary" 
            size="small"
            onClick={() => {
              setAppeal(record);
              setAppealModalVisible(true);
            }}
          >
            处理
          </Button>
        </Space>
      ),
    },
  ];

  return (
    <div style={{ padding: '24px' }}>
      <Title level={2}>插件审核管理</Title>
      
      {/* 统计卡片 */}
      {stats && (
        <Row gutter={16} style={{ marginBottom: '24px' }}>
          <Col span={6}>
            <Card>
              <Statistic
                title="待审核插件"
                value={stats.pendingReviews}
                prefix={<ClockCircleOutlined />}
                valueStyle={{ color: '#faad14' }}
              />
            </Card>
          </Col>
          <Col span={6}>
            <Card>
              <Statistic
                title="今日通过"
                value={stats.approvedToday}
                prefix={<CheckOutlined />}
                valueStyle={{ color: '#52c41a' }}
              />
            </Card>
          </Col>
          <Col span={6}>
            <Card>
              <Statistic
                title="今日拒绝"
                value={stats.rejectedToday}
                prefix={<CloseOutlined />}
                valueStyle={{ color: '#ff4d4f' }}
              />
            </Card>
          </Col>
          <Col span={6}>
            <Card>
              <Statistic
                title="平均审核时间"
                value={stats.averageReviewTime}
                suffix="小时"
                prefix={<BarChartOutlined />}
              />
            </Card>
          </Col>
        </Row>
      )}

      <Tabs activeKey={activeTab} onChange={setActiveTab}>
        <TabPane tab="审核队列" key="queue">
          <Card>
            {/* 筛选器 */}
            <div style={{ marginBottom: '16px' }}>
              <Space>
                <Select
                  placeholder="审核状态"
                  style={{ width: 120 }}
                  allowClear
                  value={filters.status}
                  onChange={(value) => setFilters({ ...filters, status: value || '' })}
                >
                  <Option value="pending">待审核</Option>
                  <Option value="in_review">审核中</Option>
                  <Option value="approved">已通过</Option>
                  <Option value="rejected">已拒绝</Option>
                </Select>
                <Select
                  placeholder="优先级"
                  style={{ width: 100 }}
                  allowClear
                  value={filters.priority}
                  onChange={(value) => setFilters({ ...filters, priority: value || '' })}
                >
                  <Option value="high">高</Option>
                  <Option value="medium">中</Option>
                  <Option value="low">低</Option>
                </Select>
                <RangePicker
                  placeholder={['开始日期', '结束日期']}
                  value={filters.dateRange}
                  onChange={(dates) => setFilters({ ...filters, dateRange: dates })}
                />
                <Button icon={<FilterOutlined />} onClick={fetchReviewQueue}>
                  刷新
                </Button>
              </Space>
            </div>

            <Table
              columns={pluginColumns}
              dataSource={plugins}
              rowKey="id"
              loading={loading}
              pagination={{
                showSizeChanger: true,
                showQuickJumper: true,
                showTotal: (total) => `共 ${total} 条记录`,
              }}
              rowSelection={{
                type: 'checkbox',
                onChange: (selectedRowKeys) => {
                  // 处理批量选择
                },
              }}
            />
          </Card>
        </TabPane>

        <TabPane tab="申诉管理" key="appeals">
          <Card>
            <Table
              columns={appealColumns}
              dataSource={appeals}
              rowKey="id"
              loading={loading}
              pagination={{
                showSizeChanger: true,
                showQuickJumper: true,
                showTotal: (total) => `共 ${total} 条记录`,
              }}
            />
          </Card>
        </TabPane>

        <TabPane tab="审核员工作负载" key="workload">
          <Card>
            {stats?.reviewerWorkload.map((reviewer) => (
              <Card key={reviewer.reviewerId} style={{ marginBottom: '16px' }}>
                <Row>
                  <Col span={8}>
                    <div>
                      <UserOutlined /> {reviewer.reviewerName}
                    </div>
                  </Col>
                  <Col span={8}>
                    <Statistic
                      title="分配任务"
                      value={reviewer.assignedCount}
                      valueStyle={{ fontSize: '16px' }}
                    />
                  </Col>
                  <Col span={8}>
                    <Statistic
                      title="完成任务"
                      value={reviewer.completedCount}
                      valueStyle={{ fontSize: '16px' }}
                    />
                  </Col>
                </Row>
                <Progress
                  percent={reviewer.assignedCount > 0 ? 
                    Math.round((reviewer.completedCount / reviewer.assignedCount) * 100) : 0}
                  status="active"
                />
              </Card>
            ))}
          </Card>
        </TabPane>
      </Tabs>

      {/* 审核模态框 */}
      <Modal
        title="插件审核"
        open={reviewModalVisible}
        onCancel={() => {
          setReviewModalVisible(false);
          reviewForm.resetFields();
        }}
        footer={null}
        width={800}
      >
        {selectedPlugin && (
          <div>
            <div style={{ marginBottom: '24px' }}>
              <Title level={4}>{selectedPlugin.name}</Title>
              <Paragraph>{selectedPlugin.description}</Paragraph>
              <Space>
                <Tag>版本: {selectedPlugin.version}</Tag>
                <Tag>开发者: {selectedPlugin.developer}</Tag>
                <Tag color={getPriorityColor(selectedPlugin.reviewPriority)}>
                  优先级: {selectedPlugin.reviewPriority}
                </Tag>
              </Space>
            </div>

            <Form
              form={reviewForm}
              layout="vertical"
              onFinish={handleReviewPlugin}
            >
              <Form.Item
                name="result"
                label="审核结果"
                rules={[{ required: true, message: '请选择审核结果' }]}
              >
                <Select placeholder="请选择审核结果">
                  <Option value="approved">通过</Option>
                  <Option value="rejected">拒绝</Option>
                  <Option value="needs_modification">需要修改</Option>
                </Select>
              </Form.Item>

              <Form.Item
                name="notes"
                label="审核备注"
                rules={[{ required: true, message: '请输入审核备注' }]}
              >
                <TextArea rows={4} placeholder="请输入审核备注" />
              </Form.Item>

              <Form.Item
                noStyle
                shouldUpdate={(prevValues, currentValues) =>
                  prevValues.result !== currentValues.result
                }
              >
                {({ getFieldValue }) =>
                  getFieldValue('result') === 'rejected' ? (
                    <Form.Item
                      name="rejectionReason"
                      label="拒绝原因"
                      rules={[{ required: true, message: '请输入拒绝原因' }]}
                    >
                      <TextArea rows={3} placeholder="请详细说明拒绝原因" />
                    </Form.Item>
                  ) : null
                }
              </Form.Item>

              <Form.Item>
                <Space>
                  <Button type="primary" htmlType="submit">
                    提交审核
                  </Button>
                  <Button onClick={() => setReviewModalVisible(false)}>
                    取消
                  </Button>
                </Space>
              </Form.Item>
            </Form>
          </div>
        )}
      </Modal>

      {/* 申诉处理模态框 */}
      <Modal
        title="处理申诉"
        open={appealModalVisible}
        onCancel={() => {
          setAppealModalVisible(false);
          appealForm.resetFields();
        }}
        footer={null}
      >
        {selectedAppeal && (
          <div>
            <div style={{ marginBottom: '24px' }}>
              <Title level={4}>{selectedAppeal.pluginName}</Title>
              <Paragraph>
                <strong>申诉原因：</strong>
                {selectedAppeal.reason}
              </Paragraph>
            </div>

            <Form
              form={appealForm}
              layout="vertical"
              onFinish={handleProcessAppeal}
            >
              <Form.Item
                name="decision"
                label="处理决定"
                rules={[{ required: true, message: '请选择处理决定' }]}
              >
                <Select placeholder="请选择处理决定">
                  <Option value="approved">接受申诉</Option>
                  <Option value="rejected">拒绝申诉</Option>
                </Select>
              </Form.Item>

              <Form.Item
                name="notes"
                label="处理说明"
                rules={[{ required: true, message: '请输入处理说明' }]}
              >
                <TextArea rows={4} placeholder="请输入处理说明" />
              </Form.Item>

              <Form.Item>
                <Space>
                  <Button type="primary" htmlType="submit">
                    提交处理
                  </Button>
                  <Button onClick={() => setAppealModalVisible(false)}>
                    取消
                  </Button>
                </Space>
              </Form.Item>
            </Form>
          </div>
        )}
      </Modal>

      {/* 审核历史模态框 */}
      <Modal
        title="审核历史"
        open={historyModalVisible}
        onCancel={() => setHistoryModalVisible(false)}
        footer={[
          <Button key="close" onClick={() => setHistoryModalVisible(false)}>
            关闭
          </Button>
        ]}
        width={600}
      >
        <Timeline>
          {reviewHistory.map((item, index) => (
            <Timeline.Item
              key={index}
              color={item.result === 'approved' ? 'green' : 
                     item.result === 'rejected' ? 'red' : 'blue'}
            >
              <div>
                <div style={{ fontWeight: 'bold' }}>
                  {item.result === 'approved' && '审核通过'}
                  {item.result === 'rejected' && '审核拒绝'}
                  {item.result === 'needs_modification' && '需要修改'}
                </div>
                <div style={{ color: '#666', fontSize: '12px' }}>
                  {dayjs(item.createdAt).format('YYYY-MM-DD HH:mm')} | {item.reviewerName}
                </div>
                <div style={{ marginTop: '8px' }}>
                  {item.notes}
                </div>
                {item.rejectionReason && (
                  <div style={{ marginTop: '4px', color: '#ff4d4f' }}>
                    拒绝原因：{item.rejectionReason}
                  </div>
                )}
              </div>
            </Timeline.Item>
          ))}
        </Timeline>
      </Modal>
    </div>
  );
};

export default PluginReviewManagement;