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
  EditOutlined
} from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { communityService } from '@/services/community';
import { BlogPost, BlogCategory } from '@/types';
import dayjs from 'dayjs';

const { Title, Text, Paragraph } = Typography;
const { Option } = Select;
const { TextArea } = Input;

const Blog: React.FC = () => {
  const navigate = useNavigate();
  const [form] = Form.useForm();
  
  const [posts, setPosts] = useState<BlogPost[]>([]);
  const [categories, setCategories] = useState<BlogCategory[]>([]);
  const [loading, setLoading] = useState(false);
  const [createModalVisible, setCreateModalVisible] = useState(false);
  const [createLoading, setCreateLoading] = useState(false);
  
  // 筛选和搜索状态
  const [selectedCategory, setSelectedCategory] = useState<string | undefined>();
  const [searchKeyword, setSearchKeyword] = useState('');
  const [sortBy, setSortBy] = useState<'latest' | 'popular' | 'views'>('latest');
  
  // 分页状态
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 10,
    total: 0
  });

  // 加载博客分类
  const loadCategories = async () => {
    try {
      const response = await communityService.getBlogCategories();
      setCategories(response.data || []);
    } catch (error) {
      message.error('加载分类失败');
      setCategories([]);
    }
  };

  // 加载文章列表
  const loadPosts = async (page = 1) => {
    setLoading(true);
    try {
      const params = {
        page,
        limit: pagination.pageSize,
        category_id: selectedCategory,
        query: searchKeyword,
        sort_by: sortBy
      };
      
      const response = await communityService.getBlogPosts(params);
      setPosts(response.data || []);
      setPagination(prev => ({
        ...prev,
        current: page,
        total: response.meta?.total || 0
      }));
    } catch (error) {
      message.error('加载文章失败');
      setPosts([]);
    } finally {
      setLoading(false);
    }
  };

  // 创建文章
  const handleCreatePost = async (values: any) => {
    setCreateLoading(true);
    try {
      await communityService.createBlogPost({
        title: values.title,
        content: values.content,
        summary: values.summary,
        category_id: values.category_id,
        tags: values.tags ? String(values.tags).trim() : undefined,
        is_published: true,
      });
      
      message.success('文章发布成功');
      setCreateModalVisible(false);
      form.resetFields();
      loadPosts(1);
    } catch (error) {
      message.error('发布失败');
    } finally {
      setCreateLoading(false);
    }
  };

  // 点赞文章
  const handleLike = async (postId: string) => {
    try {
      await communityService.toggleLike('blog_post', postId as any);
      // 重新加载当前页数据
      loadPosts(pagination.current);
    } catch (error) {
      message.error('操作失败');
    }
  };

  // 搜索处理
  const handleSearch = (value: string) => {
    setSearchKeyword(value);
    loadPosts(1);
  };

  // 分类筛选
  const handleCategoryChange = (categoryId: string | undefined) => {
    setSelectedCategory(categoryId);
    loadPosts(1);
  };

  // 排序切换
  const handleSortChange = (sort: 'latest' | 'popular' | 'views') => {
    setSortBy(sort);
    loadPosts(1);
  };

  // 页面初始化
  useEffect(() => {
    loadCategories();
    loadPosts();
  }, []);

  // 筛选条件变化时重新加载
  useEffect(() => {
    loadPosts(1);
  }, [selectedCategory, sortBy]);

  return (
    <div className="blog-page">
      <Card>
        <Row justify="space-between" align="middle" style={{ marginBottom: 24 }}>
          <Col>
            <Title level={3} style={{ margin: 0 }}>
              开发者博客
            </Title>
          </Col>
          <Col>
            <Button 
              type="primary" 
              icon={<EditOutlined />}
              onClick={() => setCreateModalVisible(true)}
            >
              写博客
            </Button>
          </Col>
        </Row>

        {/* 筛选和搜索栏 */}
        <Row gutter={16} style={{ marginBottom: 24 }}>
          <Col span={8}>
            <Input.Search
              placeholder="搜索文章..."
              allowClear
              enterButton={<SearchOutlined />}
              onSearch={handleSearch}
            />
          </Col>
          <Col span={6}>
            <Select
              placeholder="选择分类"
              allowClear
              style={{ width: '100%' }}
              value={selectedCategory}
              onChange={handleCategoryChange}
            >
              {(categories || []).map(category => (
                <Option key={category.id} value={category.id}>
                  {category.name}
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
                <ClockCircleOutlined /> 最新发布
              </Option>
              <Option value="popular">
                <FireOutlined /> 热门文章
              </Option>
              <Option value="views">
                <EyeOutlined /> 浏览最多
              </Option>
            </Select>
          </Col>
        </Row>

        {/* 文章列表 */}
        <Spin spinning={loading}>
          <List
            itemLayout="vertical"
            size="large"
            dataSource={posts || []}
            pagination={{
              current: pagination.current,
              pageSize: pagination.pageSize,
              total: pagination.total,
              showSizeChanger: true,
              showQuickJumper: true,
              showTotal: (total, range) => 
                `第 ${range[0]}-${range[1]} 条，共 ${total} 条`,
              onChange: loadPosts,
              onShowSizeChange: (current, size) => {
                setPagination(prev => ({ ...prev, pageSize: size }));
                loadPosts(current);
              }
            }}
            renderItem={(post) => (
              <List.Item
                key={post.id}
                actions={[
                  <Space key="stats">
                    <Button 
                      type="text" 
                      icon={<LikeOutlined />}
                      onClick={() => handleLike(post.id)}
                    >
                      {(post as any).likes_count ?? (post as any).like_count}
                    </Button>
                    <Button type="text" icon={<EyeOutlined />}>
                      {(post as any).views_count ?? (post as any).view_count}
                    </Button>
                  </Space>
                ]}
                extra={
                  post.cover_image && (
                    <img
                      width={272}
                      height={180}
                      alt={post.title}
                      src={post.cover_image}
                      style={{ borderRadius: 8, objectFit: 'cover' }}
                    />
                  )
                }
              >
                <List.Item.Meta
                  avatar={
                    <Avatar src={post.author?.avatar} size={48}>
                      {post.author?.username?.[0]?.toUpperCase()}
                    </Avatar>
                  }
                  title={
                    <Space direction="vertical" size={4}>
                      <a 
                        onClick={() => navigate(`/community/blog/${post.id}`)}
                        style={{ fontSize: 18, fontWeight: 600 }}
                      >
                        {post.title}
                      </a>
                      <Space>
                        {post.category && (
                          <Tag color="blue">{post.category.name}</Tag>
                        )}
                        {post.tags && String(post.tags).split(',').filter(Boolean).map(tag => (
                          <Tag key={tag} color="default">{tag}</Tag>
                        ))}
                      </Space>
                    </Space>
                  }
                  description={
                    <Space direction="vertical" size={8}>
                      <Text type="secondary">
                        {post.author?.username} · {dayjs(post.created_at).fromNow()}
                      </Text>
                      <Paragraph 
                        ellipsis={{ rows: 3 }}
                        style={{ margin: 0, color: '#595959' }}
                      >
                        {post.summary || post.content}
                      </Paragraph>
                    </Space>
                  }
                />
              </List.Item>
            )}
          />
        </Spin>
      </Card>

      {/* 创建文章弹窗 */}
      <Modal
        title="写新文章"
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
          onFinish={handleCreatePost}
        >
          <Form.Item
            name="title"
            label="文章标题"
            rules={[
              { required: true, message: '请输入文章标题' },
              { max: 100, message: '标题不能超过100个字符' }
            ]}
          >
            <Input placeholder="请输入文章标题" />
          </Form.Item>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                name="category_id"
                label="选择分类"
                rules={[{ required: true, message: '请选择分类' }]}
              >
                <Select placeholder="请选择分类">
                  {(categories || []).map(category => (
                    <Option key={category.id} value={category.id}>
                      {category.name}
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
                <Input placeholder="如：React, TypeScript, 前端" />
              </Form.Item>
            </Col>
          </Row>

          <Form.Item
            name="summary"
            label="文章摘要"
            extra="简短描述文章内容，用于列表展示"
          >
            <TextArea
              rows={3}
              placeholder="请输入文章摘要"
              maxLength={200}
              showCount
            />
          </Form.Item>

          <Form.Item
            name="content"
            label="文章内容"
            rules={[
              { required: true, message: '请输入文章内容' },
              { min: 50, message: '内容至少50个字符' }
            ]}
          >
            <TextArea
              rows={12}
              placeholder="请输入文章内容，支持Markdown格式"
            />
          </Form.Item>

          <Form.Item>
            <Space>
              <Button 
                type="primary" 
                htmlType="submit"
                loading={createLoading}
              >
                发布文章
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

export default Blog;