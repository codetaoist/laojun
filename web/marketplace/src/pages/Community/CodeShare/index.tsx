import React, { useState, useEffect } from 'react';
import { 
  Card, 
  List, 
  Button, 
  Row, 
  Col, 
  Tag, 
  Typography, 
  Space, 
  Avatar, 
  message,
  Select,
  Input,
  Modal,
  Form,
  Spin
} from 'antd';
import { 
  PlusOutlined, 
  EyeOutlined, 
  LikeOutlined,
  SearchOutlined,
  FireOutlined,
  ClockCircleOutlined,
  CodeOutlined,
  CopyOutlined
} from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { communityService } from '@/services/community';
import { CodeSnippet } from '@/types';
import dayjs from 'dayjs';

const { Title, Text, Paragraph } = Typography;
const { Option } = Select;
const { TextArea } = Input;

const CodeShare: React.FC = () => {
  const navigate = useNavigate();
  const [form] = Form.useForm();
  
  const [snippets, setSnippets] = useState<CodeSnippet[]>([]);
  const [loading, setLoading] = useState(false);
  const [createModalVisible, setCreateModalVisible] = useState(false);
  const [createLoading, setCreateLoading] = useState(false);
  
  // 筛选和搜索状态
  const [selectedLanguage, setSelectedLanguage] = useState<string | undefined>();
  const [searchKeyword, setSearchKeyword] = useState('');
  const [sortBy, setSortBy] = useState<'latest' | 'hot'>('latest');
  
  // 分页状态
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 10,
    total: 0
  });

  // 编程语言选项
  const languages = [
    'JavaScript', 'TypeScript', 'Python', 'Java', 'Go', 'Rust',
    'C++', 'C#', 'PHP', 'Ruby', 'Swift', 'Kotlin', 'Dart',
    'HTML', 'CSS', 'SQL', 'Shell', 'PowerShell'
  ];

  // 加载代码片段列表
  const loadSnippets = async (page = 1) => {
    setLoading(true);
    try {
      const params = {
        page,
        limit: pagination.pageSize,
        language: selectedLanguage,
        keyword: searchKeyword,
        sort: sortBy
      };
      
      const response = await communityService.getCodeSnippets(params);
      setSnippets(response.data || []);
      setPagination(prev => ({
        ...prev,
        current: page,
        total: response.meta?.total || 0
      }));
    } catch (error) {
      message.error('加载代码片段失败');
    } finally {
      setLoading(false);
    }
  };

  // 创建代码片段
  const handleCreateSnippet = async (values: any) => {
    setCreateLoading(true);
    try {
      await communityService.createCodeSnippet({
        title: values.title,
        description: values.description,
        code: values.code,
        language: values.language,
        tags: values.tags ? values.tags.split(',').map((tag: string) => tag.trim()) : []
      });
      
      message.success('代码片段分享成功');
      setCreateModalVisible(false);
      form.resetFields();
      loadSnippets(1);
    } catch (error) {
      message.error('分享失败');
    } finally {
      setCreateLoading(false);
    }
  };

  // 点赞代码片段
  const handleLike = async (snippetId: number) => {
    try {
      await communityService.toggleLike('code_snippet', snippetId);
      // 重新加载当前页数据
      loadSnippets(pagination.current);
    } catch (error) {
      message.error('操作失败');
    }
  };

  // 复制代码
  const handleCopyCode = async (code: string) => {
    try {
      await navigator.clipboard.writeText(code);
      message.success('代码已复制到剪贴板');
    } catch (error) {
      message.error('复制失败');
    }
  };

  // 搜索处理
  const handleSearch = (value: string) => {
    setSearchKeyword(value);
    loadSnippets(1);
  };

  // 语言筛选
  const handleLanguageChange = (language: string | undefined) => {
    setSelectedLanguage(language);
    loadSnippets(1);
  };

  // 排序切换
  const handleSortChange = (sort: 'latest' | 'hot') => {
    setSortBy(sort);
    loadSnippets(1);
  };

  // 页面初始化
  useEffect(() => {
    loadSnippets();
  }, []);

  // 筛选条件变化时重新加载
  useEffect(() => {
    loadSnippets(1);
  }, [selectedLanguage, sortBy]);

  return (
    <div className="code-share-page">
      <Card>
        <Row justify="space-between" align="middle" style={{ marginBottom: 24 }}>
          <Col>
            <Title level={3} style={{ margin: 0 }}>
              代码分享
            </Title>
          </Col>
          <Col>
            <Button 
              type="primary" 
              icon={<CodeOutlined />}
              onClick={() => setCreateModalVisible(true)}
            >
              分享代码
            </Button>
          </Col>
        </Row>

        {/* 筛选和搜索栏 */}
        <Row gutter={16} style={{ marginBottom: 24 }}>
          <Col span={8}>
            <Input.Search
              placeholder="搜索代码片段..."
              allowClear
              enterButton={<SearchOutlined />}
              onSearch={handleSearch}
            />
          </Col>
          <Col span={6}>
            <Select
              placeholder="选择语言"
              allowClear
              style={{ width: '100%' }}
              value={selectedLanguage}
              onChange={handleLanguageChange}
            >
              {languages.map(language => (
                <Option key={language} value={language}>
                  {language}
                </Option>
              ))}
            </Select>
          </Col>
          <Col span={6}>
            <Select
              value={sortBy}
              onChange={handleSortChange}
              style={{ width: '100%' }}
            >
              <Option value="latest">
                <ClockCircleOutlined /> 最新分享
              </Option>
              <Option value="hot">
                <FireOutlined /> 热门代码
              </Option>
            </Select>
          </Col>
        </Row>

        {/* 代码片段列表 */}
        <Spin spinning={loading}>
          <List
            itemLayout="vertical"
            size="large"
            dataSource={snippets}
            pagination={{
              current: pagination.current,
              pageSize: pagination.pageSize,
              total: pagination.total,
              showSizeChanger: true,
              showQuickJumper: true,
              showTotal: (total, range) => 
                `第 ${range[0]}-${range[1]} 条，共 ${total} 条`,
              onChange: loadSnippets,
              onShowSizeChange: (current, size) => {
                setPagination(prev => ({ ...prev, pageSize: size }));
                loadSnippets(current);
              }
            }}
            renderItem={(snippet) => (
              <List.Item
                key={snippet.id}
                actions={[
                  <Space key="stats">
                    <Button 
                      type="text" 
                      icon={<LikeOutlined />}
                      onClick={() => handleLike(snippet.id)}
                    >
                      {snippet.likes_count}
                    </Button>
                    <Button type="text" icon={<EyeOutlined />}>
                      {snippet.views_count}
                    </Button>
                    <Button 
                      type="text" 
                      icon={<CopyOutlined />}
                      onClick={() => handleCopyCode(snippet.code)}
                    >
                      复制
                    </Button>
                  </Space>
                ]}
              >
                <List.Item.Meta
                  avatar={
                    <Avatar src={snippet.author?.avatar} size={48}>
                      {snippet.author?.username?.[0]?.toUpperCase()}
                    </Avatar>
                  }
                  title={
                    <Space direction="vertical" size={4}>
                      <a 
                        onClick={() => navigate(`/community/code/${snippet.id}`)}
                        style={{ fontSize: 16, fontWeight: 500 }}
                      >
                        {snippet.title}
                      </a>
                      <Space>
                        <Tag color="green">{snippet.language}</Tag>
                        {snippet.tags?.map(tag => (
                          <Tag key={tag} color="default">{tag}</Tag>
                        ))}
                      </Space>
                    </Space>
                  }
                  description={
                    <Space direction="vertical" size={8}>
                      <Text type="secondary">
                        {snippet.author?.username} · {dayjs(snippet.created_at).fromNow()}
                      </Text>
                      {snippet.description && (
                        <Paragraph 
                          ellipsis={{ rows: 2 }}
                          style={{ margin: 0, color: '#595959' }}
                        >
                          {snippet.description}
                        </Paragraph>
                      )}
                    </Space>
                  }
                />
                
                {/* 代码预览 */}
                <Card 
                  size="small" 
                  style={{ 
                    marginTop: 12, 
                    backgroundColor: '#f6f8fa',
                    border: '1px solid #e1e4e8'
                  }}
                >
                  <pre 
                    style={{ 
                      margin: 0, 
                      fontSize: 13,
                      fontFamily: 'Monaco, Menlo, "Ubuntu Mono", monospace',
                      overflow: 'auto',
                      maxHeight: 200,
                      whiteSpace: 'pre-wrap',
                      wordBreak: 'break-all'
                    }}
                  >
                    <code>{snippet.code}</code>
                  </pre>
                </Card>
              </List.Item>
            )}
          />
        </Spin>
      </Card>

      {/* 创建代码片段弹窗 */}
      <Modal
        title="分享代码片段"
        open={createModalVisible}
        onCancel={() => {
          setCreateModalVisible(false);
          form.resetFields();
        }}
        footer={null}
        width={900}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleCreateSnippet}
        >
          <Form.Item
            name="title"
            label="代码标题"
            rules={[
              { required: true, message: '请输入代码标题' },
              { max: 100, message: '标题不能超过100个字符' }
            ]}
          >
            <Input placeholder="请输入代码标题" />
          </Form.Item>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                name="language"
                label="编程语言"
                rules={[{ required: true, message: '请选择编程语言' }]}
              >
                <Select placeholder="请选择编程语言">
                  {languages.map(language => (
                    <Option key={language} value={language}>
                      {language}
                    </Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="tags"
                label="标签"
                extra="多个标签用逗号分隔"
              >
                <Input placeholder="如：算法, 工具函数, 组件" />
              </Form.Item>
            </Col>
          </Row>

          <Form.Item
            name="description"
            label="代码描述"
            extra="简短描述代码的功能和用途"
          >
            <TextArea
              rows={3}
              placeholder="请输入代码描述"
              maxLength={500}
              showCount
            />
          </Form.Item>

          <Form.Item
            name="code"
            label="代码内容"
            rules={[
              { required: true, message: '请输入代码内容' },
              { min: 10, message: '代码内容至少10个字符' }
            ]}
          >
            <TextArea
              rows={15}
              placeholder="请输入代码内容"
              style={{ 
                fontFamily: 'Monaco, Menlo, "Ubuntu Mono", monospace',
                fontSize: 13
              }}
            />
          </Form.Item>

          <Form.Item>
            <Space>
              <Button 
                type="primary" 
                htmlType="submit"
                loading={createLoading}
              >
                分享代码
              </Button>
              <Button 
                onClick={() => {
                  setCreateModalVisible(false);
                  form.resetFields();
                }}
              >
                取消
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default CodeShare;