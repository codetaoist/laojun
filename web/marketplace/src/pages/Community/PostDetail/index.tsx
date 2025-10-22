import React, { useState, useEffect } from 'react';
import { 
  Card, 
  Button, 
  Row, 
  Col, 
  Tag, 
  Typography, 
  Space, 
  Avatar, 
  message,
  Divider,
  Form,
  Input,
  List,
  Spin,
  Breadcrumb
} from 'antd';
import { 
  LikeOutlined, 
  MessageOutlined, 
  EyeOutlined,
  ArrowLeftOutlined,
  SendOutlined
} from '@ant-design/icons';
import { useParams, useNavigate } from 'react-router-dom';
import { communityService } from '@/services/community';
import { ForumPost, ForumReply } from '@/types';
import dayjs from 'dayjs';
import ErrorBoundary from '@/components/ErrorBoundary';
import LoadingState from '@/components/LoadingState';
import useAsyncOperation from '@/hooks/useAsyncOperation';
import './index.css';

const { Title, Text, Paragraph } = Typography;
const { TextArea } = Input;

const PostDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [form] = Form.useForm();
  
  const [post, setPost] = useState<ForumPost | null>(null);
  const [replies, setReplies] = useState<ForumReply[]>([]);
  const [replyLoading, setReplyLoading] = useState(false);
  
  // 分页状态
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 20,
    total: 0
  });

  // 使用自定义Hook处理帖子详情加载
  const {
    loading: postLoading,
    error: postError,
    execute: loadPost,
    retry: retryLoadPost
  } = useAsyncOperation(
    async () => {
      const response = await communityService.getForumPost(Number(id));
      setPost(response.data);
      return response;
    },
    {
      showErrorMessage: true,
      errorMessage: '加载帖子详情失败'
    }
  );

  // 使用自定义Hook处理回复列表加载
  const {
    loading: repliesLoading,
    error: repliesError,
    execute: loadReplies,
    retry: retryLoadReplies
  } = useAsyncOperation(
    async (page = 1) => {
      const response = await communityService.getForumReplies(Number(id), {
        page,
        limit: pagination.pageSize
      });
      setReplies(response.data || []);
      setPagination(prev => ({
        ...prev,
        current: page,
        total: response.meta?.total || 0
      }));
      return response;
    },
    {
      showErrorMessage: true,
      errorMessage: '加载回复失败'
    }
  );

  // 使用自定义Hook处理回复提交
  const {
    loading: replySubmitting,
    execute: submitReply
  } = useAsyncOperation(
    async (values: any) => {
      const response = await communityService.createForumReply(Number(id), values);
      form.resetFields();
      loadReplies(1); // 重新加载第一页回复
      return response;
    },
    {
      showSuccessMessage: true,
      successMessage: '回复发布成功！',
      showErrorMessage: true,
      errorMessage: '回复失败，请稍后重试'
    }
  );

  // 提交回复
  const handleReplySubmit = async (values: any) => {
    await submitReply(values);
  };

  // 点赞帖子
  const handleLike = async () => {
    if (!post) return;
    
    try {
      await communityService.toggleLike('forum_post', post.id);
      loadPost(); // 重新加载帖子以更新点赞数
    } catch (error) {
      message.error('操作失败');
    }
  };

  // 页面初始化
  useEffect(() => {
    if (id) {
      loadPost();
      loadReplies();
    }
  }, [id]);

  return (
    <ErrorBoundary>
      <LoadingState
        loading={postLoading}
        error={postError}
        type="card"
        onRetry={retryLoadPost}
        empty={!post}
        emptyText="帖子不存在"
        emptyDescription="该帖子可能已被删除或不存在"
      >
        {post && (
    <div className="post-detail-page" style={{ padding: '24px', maxWidth: '1200px', margin: '0 auto' }}>
      {/* 面包屑导航 - 桌面端显示 */}
      <Breadcrumb style={{ marginBottom: '24px' }} className="desktop-only">
        <Breadcrumb.Item>
          <a onClick={() => navigate('/community')}>社区</a>
        </Breadcrumb.Item>
        <Breadcrumb.Item>
          <a onClick={() => navigate('/community/forum')}>论坛</a>
        </Breadcrumb.Item>
        <Breadcrumb.Item>{post.title}</Breadcrumb.Item>
      </Breadcrumb>

      {/* 返回按钮 */}
      <Button 
        icon={<ArrowLeftOutlined />} 
        onClick={() => navigate('/community/forum')}
        style={{ marginBottom: '16px' }}
        block
        className="mobile-only"
      >
        返回论坛
      </Button>
      <Button 
        icon={<ArrowLeftOutlined />} 
        onClick={() => navigate('/community/forum')}
        style={{ marginBottom: '16px' }}
        className="desktop-only"
      >
        返回论坛
      </Button>

      {/* 帖子内容 */}
      <Card>
        <Row gutter={16}>
          <Col xs={24} lg={18}>
            <Space direction="vertical" size="large" style={{ width: '100%' }}>
              {/* 帖子标题和标签 */}
              <div>
                <Title level={2} style={{ marginBottom: '8px', fontSize: 'clamp(18px, 4vw, 24px)' }}>
                  {post.title}
                </Title>
                <Space wrap>
                  {post.is_pinned && (
                    <Tag color="red">置顶</Tag>
                  )}
                  {post.category && (
                    <Tag color="blue">{post.category.name}</Tag>
                  )}
                </Space>
              </div>

              {/* 作者信息 */}
              <Space>
                <Avatar src={post.author?.avatar} size={40}>
                  {post.author?.username?.[0]?.toUpperCase()}
                </Avatar>
                <div>
                  <div style={{ fontWeight: 500 }}>{post.author?.username}</div>
                  <Text type="secondary">
                    {dayjs(post.created_at).format('YYYY-MM-DD HH:mm')}
                  </Text>
                </div>
              </Space>

              {/* 帖子内容 */}
              <Paragraph style={{ fontSize: 'clamp(14px, 3vw, 16px)', lineHeight: '1.6' }}>
                {post.content}
              </Paragraph>

              {/* 操作按钮 */}
              <Space size="large" wrap>
                <Button 
                  type="text" 
                  icon={<LikeOutlined />}
                  onClick={handleLike}
                >
                  <span className="desktop-only">{post.likes_count} 点赞</span>
                  <span className="mobile-only">{post.likes_count}</span>
                </Button>
                <Button type="text" icon={<MessageOutlined />}>
                  <span className="desktop-only">{post.replies_count} 回复</span>
                  <span className="mobile-only">{post.replies_count}</span>
                </Button>
                <Button type="text" icon={<EyeOutlined />}>
                  <span className="desktop-only">{post.views_count} 浏览</span>
                  <span className="mobile-only">{post.views_count}</span>
                </Button>
              </Space>
            </Space>
          </Col>
        </Row>
      </Card>

      <Divider />

      {/* 回复表单 */}
      <Card title="发表回复" style={{ marginBottom: '24px' }}>
        <Form form={form} onFinish={handleReplySubmit}>
          <Form.Item
            name="content"
            rules={[
              { required: true, message: '请输入回复内容' },
              { min: 5, message: '回复内容至少5个字符' }
            ]}
          >
            <TextArea
              rows={4}
              placeholder="请输入回复内容..."
            />
          </Form.Item>
          <Form.Item>
            <Button 
              type="primary" 
              htmlType="submit"
              icon={<SendOutlined />}
              loading={replySubmitting}
            >
              发表回复
            </Button>
          </Form.Item>
        </Form>
      </Card>

      {/* 回复列表 */}
      <Card title={`回复 (${pagination.total})`}>
        <LoadingState
          loading={repliesLoading}
          error={repliesError}
          type="list"
          onRetry={retryLoadReplies}
          empty={replies.length === 0}
          emptyText="暂无回复"
          emptyDescription="成为第一个回复的人吧！"
        >
          <List
            itemLayout="vertical"
            dataSource={replies}
            pagination={{
              current: pagination.current,
              pageSize: pagination.pageSize,
              total: pagination.total,
              showSizeChanger: false,
              showQuickJumper: true,
              showTotal: (total, range) => 
                `第 ${range[0]}-${range[1]} 条，共 ${total} 条`,
              onChange: loadReplies
            }}
            renderItem={(reply, index) => (
              <List.Item key={reply.id}>
                <List.Item.Meta
                  avatar={
                    <Avatar src={reply.author?.avatar}>
                      {reply.author?.username?.[0]?.toUpperCase()}
                    </Avatar>
                  }
                  title={
                    <Space>
                      <span style={{ fontWeight: 500 }}>
                        {reply.author?.username}
                      </span>
                      <Text type="secondary">
                        #{pagination.pageSize * (pagination.current - 1) + index + 1}
                      </Text>
                      <Text type="secondary">
                        {dayjs(reply.created_at).fromNow()}
                      </Text>
                    </Space>
                  }
                  description={
                    <Paragraph style={{ marginTop: '8px', marginBottom: 0 }}>
                      {reply.content}
                    </Paragraph>
                  }
                />
              </List.Item>
            )}
          />
        </LoadingState>
      </Card>
    </div>
        )}
      </LoadingState>
    </ErrorBoundary>
  );
};

export default PostDetail;