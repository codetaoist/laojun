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
  Drawer,
  Divider
} from 'antd';
import { 
  PlusOutlined, 
  MessageOutlined, 
  EyeOutlined, 
  LikeOutlined,
  SearchOutlined,
  FireOutlined,
  ClockCircleOutlined,
  FilterOutlined,
  MenuOutlined
} from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { communityService } from '@/services/community';
import { ForumPost, ForumCategory } from '@/types';
import dayjs from 'dayjs';
import ErrorBoundary from '@/components/ErrorBoundary';
import LoadingState from '@/components/LoadingState';
import useAsyncOperation from '@/hooks/useAsyncOperation';
import './index.css';

const { Title, Text } = Typography;
const { Option } = Select;
const { TextArea } = Input;

const Forum: React.FC = () => {
  const navigate = useNavigate();
  const [form] = Form.useForm();
  
  const [posts, setPosts] = useState<ForumPost[]>([]);
  const [categories, setCategories] = useState<ForumCategory[]>([]);
  const [createModalVisible, setCreateModalVisible] = useState(false);
  const [filterDrawerVisible, setFilterDrawerVisible] = useState(false);
  
  // 筛选和搜索状态
  const [selectedCategory, setSelectedCategory] = useState<number | undefined>();
  const [searchKeyword, setSearchKeyword] = useState('');
  const [sortBy, setSortBy] = useState<'latest' | 'hot'>('latest');
  
  // 分页状态
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 10,
    total: 0
  });

  // 使用自定义Hook处理帖子列表加载
  const {
    loading: postsLoading,
    error: postsError,
    execute: loadPosts,
    retry: retryLoadPosts
  } = useAsyncOperation(
    async (page = 1) => {
      const params = {
        page,
        limit: pagination.pageSize,
        category_id: selectedCategory,
        keyword: searchKeyword,
        sort: sortBy
      };
      
      const response = await communityService.getForumPosts(params);
      setPosts(response.data || []);
      setPagination(prev => ({
        ...prev,
        current: page,
        total: response.meta?.total || 0
      }));
      return response;
    },
    {
      showErrorMessage: true,
      errorMessage: '加载帖子失败，请稍后重试'
    }
  );

  // 使用自定义Hook处理分类加载
  const {
    loading: categoriesLoading,
    error: categoriesError,
    execute: loadCategories
  } = useAsyncOperation(
    async () => {
      const response = await communityService.getForumCategories();
      setCategories(response.data);
      return response;
    },
    {
      showErrorMessage: true,
      errorMessage: '加载分类失败'
    }
  );

  // 使用自定义Hook处理帖子创建
  const {
    loading: createLoading,
    execute: createPost
  } = useAsyncOperation(
    async (values: any) => {
      const response = await communityService.createForumPost(values);
      setCreateModalVisible(false);
      form.resetFields();
      loadPosts(1);
      return response;
    },
    {
      showSuccessMessage: true,
      successMessage: '帖子发布成功！',
      showErrorMessage: true,
      errorMessage: '发布失败，请稍后重试'
    }
  );

  // 页面初始化
  useEffect(() => {
    loadCategories();
    loadPosts(1);
  }, []);

  // 当筛选条件改变时重新加载
  useEffect(() => {
    loadPosts(1);
  }, [selectedCategory, searchKeyword, sortBy]);

  // 创建帖子
  const handleCreatePost = async (values: any) => {
    await createPost(values);
  };

  // 点赞帖子
  const handleLike = async (postId: number) => {
    try {
      await communityService.toggleLike('forum_post', postId);
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
  const handleCategoryChange = (categoryId: number | undefined) => {
    setSelectedCategory(categoryId);
    loadPosts(1);
  };

  // 排序切换
  const handleSortChange = (sort: 'latest' | 'hot') => {
    setSortBy(sort);
    loadPosts(1);
  };

  return (
    <ErrorBoundary>
      <div className="forum-page">
        <Card>
          <Row justify="space-between" align="middle" style={{ marginBottom: 24 }}>
            <Col flex="auto">
              <Title level={3} style={{ margin: 0 }}>
                开发者论坛
              </Title>
            </Col>
            <Col>
              <Space>
                <Button 
                  className="mobile-only"
                  icon={<FilterOutlined />}
                  onClick={() => setFilterDrawerVisible(true)}
                >
                  筛选
                </Button>
                <Button 
                  type="primary" 
                  icon={<PlusOutlined />}
                  onClick={() => setCreateModalVisible(true)}
                >
                  <span className="desktop-only">发布帖子</span>
                  <span className="mobile-only">发布</span>
                </Button>
              </Space>
            </Col>
          </Row>

          {/* 桌面端筛选和搜索栏 */}
          <Row gutter={16} style={{ marginBottom: 24 }} className="desktop-only">
            <Col span={8}>
              <Input.Search
                placeholder="搜索帖子..."
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
                {categories.map(category => (
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
                <Option value="hot">
                  <FireOutlined /> 热门讨论
                </Option>
              </Select>
            </Col>
          </Row>

          {/* 帖子列表 */}
          <LoadingState
            loading={postsLoading}
            error={postsError}
            type="list"
            onRetry={retryLoadPosts}
            empty={posts.length === 0}
            emptyText="暂无帖子"
            emptyDescription="还没有人发布帖子，快来发布第一个吧！"
          >
            <List
              dataSource={posts}
              pagination={{
                current: pagination.current,
                pageSize: pagination.pageSize,
                total: pagination.total,
                showSizeChanger: false,
                showQuickJumper: true,
                onChange: (page) => loadPosts(page),
                showTotal: (total, range) => `第 ${range[0]}-${range[1]} 条，共 ${total} 条`
              }}
              renderItem={(post) => (
                <List.Item
                  key={post.id}
                  actions={[
                    <Button
                      type="text"
                      icon={<LikeOutlined />}
                      onClick={() => handleLike(post.id)}
                    >
                      {post.likes_count}
                    </Button>,
                    <Button
                      type="text"
                      icon={<MessageOutlined />}
                      onClick={() => navigate(`/community/forum/${post.id}`)}
                    >
                      {post.replies_count}
                    </Button>,
                    <Button
                      type="text"
                      icon={<EyeOutlined />}
                    >
                      {post.views_count}
                    </Button>
                  ]}
                >
                  <List.Item.Meta
                    avatar={<Avatar src={post.avatar_url} />}
                    title={
                      <Space>
                        <a onClick={() => navigate(`/community/forum/${post.id}`)}>
                          {post.title}
                        </a>
                        {post.is_pinned && <Tag color="red">置顶</Tag>}
                        {post.category && <Tag color="blue">{post.category.name}</Tag>}
                      </Space>
                    }
                    description={
                      <Space direction="vertical" size={4}>
                        <Text type="secondary">
                          {post.username} · {dayjs(post.created_at).fromNow()}
                        </Text>
                        <Text ellipsis={{ rows: 2 }}>
                          {post.content}
                        </Text>
                      </Space>
                    }
                  />
                </List.Item>
              )}
            />
          </LoadingState>
        </Card>

        {/* 创建帖子模态框 */}
        <Modal
          title="发布新帖子"
          open={createModalVisible}
          onCancel={() => setCreateModalVisible(false)}
          footer={null}
          width={600}
        >
          <Form
            form={form}
            layout="vertical"
            onFinish={handleCreatePost}
          >
            <Form.Item
              name="title"
              label="标题"
              rules={[{ required: true, message: '请输入帖子标题' }]}
            >
              <Input placeholder="请输入帖子标题" />
            </Form.Item>
            
            <Form.Item
              name="category_id"
              label="分类"
              rules={[{ required: true, message: '请选择分类' }]}
            >
              <Select placeholder="请选择分类">
                {categories.map(category => (
                  <Option key={category.id} value={category.id}>
                    {category.name}
                  </Option>
                ))}
              </Select>
            </Form.Item>
            
            <Form.Item
              name="content"
              label="内容"
              rules={[{ required: true, message: '请输入帖子内容' }]}
            >
              <TextArea
                rows={8}
                placeholder="请输入帖子内容"
              />
            </Form.Item>
            
            <Form.Item>
              <Space>
                <Button type="primary" htmlType="submit" loading={createLoading}>
                  发布
                </Button>
                <Button onClick={() => setCreateModalVisible(false)}>
                  取消
                </Button>
              </Space>
            </Form.Item>
          </Form>
        </Modal>

        {/* 移动端筛选抽屉 */}
        <Drawer
          title="筛选和搜索"
          placement="right"
          onClose={() => setFilterDrawerVisible(false)}
          open={filterDrawerVisible}
          width={300}
        >
          <Space direction="vertical" style={{ width: '100%' }} size="large">
            <div>
              <Text strong>搜索</Text>
              <Input.Search
                placeholder="搜索帖子..."
                allowClear
                onSearch={handleSearch}
                style={{ marginTop: 8 }}
              />
            </div>
            
            <div>
              <Text strong>分类</Text>
              <Select
                placeholder="选择分类"
                allowClear
                style={{ width: '100%', marginTop: 8 }}
                value={selectedCategory}
                onChange={handleCategoryChange}
              >
                {categories.map(category => (
                  <Option key={category.id} value={category.id}>
                    {category.name}
                  </Option>
                ))}
              </Select>
            </div>
            
            <div>
              <Text strong>排序</Text>
              <Select
                value={sortBy}
                onChange={handleSortChange}
                style={{ width: '100%', marginTop: 8 }}
              >
                <Option value="latest">
                  <ClockCircleOutlined /> 最新发布
                </Option>
                <Option value="hot">
                  <FireOutlined /> 热门讨论
                </Option>
              </Select>
            </div>
            
            <Divider />
            
            <Space>
              <Button 
                type="primary" 
                onClick={() => setFilterDrawerVisible(false)}
              >
                应用筛选
              </Button>
              <Button 
                onClick={() => {
                  setSelectedCategory(undefined);
                  setSearchKeyword('');
                  setSortBy('latest');
                }}
              >
                重置筛选
              </Button>
            </Space>
          </Space>
        </Drawer>
      </div>
    </ErrorBoundary>
  );
};

export default Forum;