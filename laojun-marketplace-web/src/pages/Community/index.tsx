import React, { useState, useEffect } from 'react';
import { Card, Tabs, List, Button, Row, Col, Statistic, Tag, Typography, Space, Avatar, message, Drawer } from 'antd';
import { PlusOutlined, MessageOutlined, EyeOutlined, LikeOutlined, CodeOutlined, EditOutlined, UserOutlined, CommentOutlined, MenuOutlined, FireOutlined } from '@ant-design/icons';
import { communityService } from '@/services/community';
import { ForumPost, BlogPost, CodeSnippet } from '@/types';
import { useNavigate, Link } from 'react-router-dom';
import dayjs from 'dayjs';
import ErrorBoundary from '@/components/ErrorBoundary';
import LoadingState from '@/components/LoadingState';
import useAsyncOperation from '@/hooks/useAsyncOperation';
import './index.css';

const { Title, Text, Paragraph } = Typography;
const { TabPane } = Tabs;

const Community: React.FC = () => {
  const navigate = useNavigate();
  const [activeTab, setActiveTab] = useState('forum');
  const [mobileMenuVisible, setMobileMenuVisible] = useState(false);
  
  // 数据状态
  const [forumPosts, setForumPosts] = useState<ForumPost[]>([]);
  const [blogPosts, setBlogPosts] = useState<BlogPost[]>([]);
  const [codeSnippets, setCodeSnippets] = useState<CodeSnippet[]>([]);
  const [stats, setStats] = useState<any>({});
  const [drawerVisible, setDrawerVisible] = useState(false);

  // 使用自定义Hook处理数据加载
  const {
    loading: communityLoading,
    error: communityError,
    execute: loadCommunityData,
    retry: retryCommunityData
  } = useAsyncOperation(
    async () => {
      const [forumRes, blogRes, codeRes, statsRes] = await Promise.allSettled([
        communityService.getForumPosts({ limit: 5, sort: 'latest' }),
        communityService.getBlogPosts({ limit: 5, sort_by: 'latest' }),
        communityService.getCodeSnippets({ limit: 5, sort_by: 'latest' }),
        communityService.getCommunityStats()
      ]);

      const results = {
        forumPosts: forumRes.status === 'fulfilled' ? (forumRes.value.data || []) : [],
        blogPosts: blogRes.status === 'fulfilled' ? (blogRes.value.data || []) : [],
        codeSnippets: codeRes.status === 'fulfilled' ? (codeRes.value.data || []) : [],
        stats: statsRes.status === 'fulfilled' ? statsRes.value : {}
      };

      setForumPosts(results.forumPosts);
      setBlogPosts(results.blogPosts);
      setCodeSnippets(results.codeSnippets);
      setStats(results.stats);

      return results;
    },
    {
      showErrorMessage: true,
      errorMessage: '加载社区数据失败，请稍后重试'
    }
  );

  // 页面初始化
  useEffect(() => {
    loadCommunityData();
  }, []);

  // 格式化时间
  const formatTime = (timeStr: string) => {
    const time = new Date(timeStr);
    const now = new Date();
    const diff = now.getTime() - time.getTime();
    
    if (diff < 60000) return '刚刚';
    if (diff < 3600000) return `${Math.floor(diff / 60000)}分钟前`;
    if (diff < 86400000) return `${Math.floor(diff / 3600000)}小时前`;
    if (diff < 604800000) return `${Math.floor(diff / 86400000)}天前`;
    
    return time.toLocaleDateString();
  };

  // 渲染论坛帖子列表
  const renderForumPosts = () => (
    <List
      itemLayout="vertical"
      dataSource={forumPosts}
      renderItem={(post) => (
        <List.Item
          key={post.id}
          actions={[
            <Space key="stats">
              <EyeOutlined /> {post.view_count}
              <CommentOutlined /> {post.reply_count}
              <LikeOutlined /> {post.like_count}
            </Space>
          ]}
        >
          <List.Item.Meta
            avatar={<Avatar icon={<UserOutlined />} />}
            title={
              <Link to={`/community/forum/post/${post.id}`}>
                {post.is_pinned && <Tag color="red">置顶</Tag>}
                {post.is_featured && <Tag color="gold">精华</Tag>}
                {post.title}
              </Link>
            }
            description={
              <Space>
                <Text type="secondary">{post.author?.username || '匿名用户'}</Text>
                <Text type="secondary">•</Text>
                <Text type="secondary">{formatTime(post.created_at)}</Text>
                {post.category && (
                  <>
                    <Text type="secondary">•</Text>
                    <Tag color={post.category.color}>{post.category.name}</Tag>
                  </>
                )}
              </Space>
            }
          />
          <Paragraph ellipsis={{ rows: 2 }}>
            {post.content.replace(/[#*`]/g, '').substring(0, 200)}
          </Paragraph>
        </List.Item>
      )}
    />
  );

  // 渲染博客文章列表
  const renderBlogPosts = () => (
    <List
      itemLayout="vertical"
      dataSource={blogPosts}
      renderItem={(post) => (
        <List.Item
          key={post.id}
          actions={[
            <Space key="stats">
              <EyeOutlined /> {post.view_count}
              <LikeOutlined /> {post.like_count}
              <CommentOutlined /> {post.comment_count}
            </Space>
          ]}
          extra={
            post.cover_image && (
              <img
                width={200}
                alt={post.title}
                src={post.cover_image}
                style={{ borderRadius: 8 }}
              />
            )
          }
        >
          <List.Item.Meta
            avatar={<Avatar icon={<UserOutlined />} />}
            title={
              <Link to={`/community/blog/post/${post.id}`}>
                {post.title}
              </Link>
            }
            description={
              <Space>
                <Text type="secondary">{post.author?.username || '匿名用户'}</Text>
                <Text type="secondary">•</Text>
                <Text type="secondary">{formatTime(post.created_at)}</Text>
                {post.category && (
                  <>
                    <Text type="secondary">•</Text>
                    <Tag color={post.category.color}>{post.category.name}</Tag>
                  </>
                )}
              </Space>
            }
          />
          <Paragraph ellipsis={{ rows: 2 }}>
            {post.summary || post.content.replace(/[#*`]/g, '').substring(0, 200)}
          </Paragraph>
        </List.Item>
      )}
    />
  );

  // 渲染代码片段列表
  const renderCodeSnippets = () => (
    <List
      itemLayout="vertical"
      dataSource={codeSnippets}
      renderItem={(snippet) => (
        <List.Item
          key={snippet.id}
          actions={[
            <Space key="stats">
              <EyeOutlined /> {snippet.view_count}
              <LikeOutlined /> {snippet.like_count}
              <CodeOutlined /> {snippet.language}
            </Space>
          ]}
        >
          <List.Item.Meta
            avatar={<Avatar icon={<UserOutlined />} />}
            title={
              <Link to={`/community/code/snippet/${snippet.id}`}>
                {snippet.title}
              </Link>
            }
            description={
              <Space>
                <Text type="secondary">{snippet.user?.username || '匿名用户'}</Text>
                <Text type="secondary">•</Text>
                <Text type="secondary">{formatTime(snippet.created_at)}</Text>
                <Tag color="blue">{snippet.language}</Tag>
              </Space>
            }
          />
          {snippet.description && (
            <Paragraph ellipsis={{ rows: 2 }}>
              {snippet.description}
            </Paragraph>
          )}
        </List.Item>
      )}
    />
  );

  return (
    <ErrorBoundary>
      <div className="community-page">
        <LoadingState
          loading={communityLoading}
          error={communityError}
          type="skeleton"
          onRetry={retryCommunityData}
          empty={!forumPosts.length && !blogPosts.length && !codeSnippets.length}
          emptyText="社区暂无内容"
          emptyDescription="成为第一个发布内容的用户吧！"
        >
      {/* 页面头部 */}
      <div className="community-header">
        <Row gutter={[24, 24]}>
          <Col span={24}>
            <Card>
              <Row align="middle" justify="space-between">
                <Col>
                  <Title level={2} style={{ margin: 0 }}>
                    <FireOutlined style={{ color: '#ff4d4f', marginRight: 8 }} />
                    开发者社区
                  </Title>
                  <Text type="secondary">
                    分享知识，交流经验，共同成长
                  </Text>
                </Col>
                <Col>
                  <Space>
                    <Button 
                      type="primary" 
                      icon={<PlusOutlined />}
                      onClick={() => navigate('/community/forum/create')}
                    >
                      发布帖子
                    </Button>
                    <Button 
                      icon={<EditOutlined />}
                      onClick={() => navigate('/community/blog/create')}
                    >
                      写博客
                    </Button>
                    <Button 
                      icon={<CodeOutlined />}
                      onClick={() => navigate('/community/code/create')}
                    >
                      分享代码
                    </Button>
                  </Space>
                </Col>
              </Row>
            </Card>
          </Col>
        </Row>
      </div>

      {/* 统计数据 */}
      <Row gutter={[24, 24]} style={{ marginTop: 24 }}>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title="论坛帖子"
              value={stats.total_posts}
              prefix={<MessageOutlined />}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title="博客文章"
              value={stats.total_blogs}
              prefix={<EditOutlined />}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title="代码片段"
              value={stats.total_snippets}
              prefix={<CodeOutlined />}
              valueStyle={{ color: '#722ed1' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title="活跃用户"
              value={stats.active_users}
              prefix={<UserOutlined />}
              valueStyle={{ color: '#fa8c16' }}
            />
          </Card>
        </Col>
      </Row>

      {/* 内容区域 */}
      <Row gutter={[24, 24]} style={{ marginTop: 24 }}>
        <Col span={24}>
          <Card>
            <Tabs 
              activeKey={activeTab} 
              onChange={setActiveTab}
              tabBarExtraContent={
                <div className="community-tab-extra">
                  {/* 桌面端显示链接 */}
                  <Space className="desktop-only">
                    <Link to="/community/forum">查看全部论坛</Link>
                    <Link to="/community/blog">查看全部博客</Link>
                    <Link to="/community/code">查看全部代码</Link>
                  </Space>
                  {/* 移动端显示菜单按钮 */}
                  <Button 
                    className="mobile-only"
                    icon={<MenuOutlined />}
                    onClick={() => setMobileMenuVisible(true)}
                  />
                </div>
              }
            >
              <TabPane 
                tab={
                  <span>
                    <MessageOutlined />
                    <span className="tab-text">最新帖子</span>
                  </span>
                } 
                key="forum"
              >
                {renderForumPosts()}
              </TabPane>
              <TabPane 
                tab={
                  <span>
                    <EditOutlined />
                    <span className="tab-text">最新博客</span>
                  </span>
                } 
                key="blog"
              >
                {renderBlogPosts()}
              </TabPane>
              <TabPane 
                tab={
                  <span>
                    <CodeOutlined />
                    <span className="tab-text">最新代码</span>
                  </span>
                } 
                key="code"
              >
                {renderCodeSnippets()}
              </TabPane>
            </Tabs>
          </Card>
        </Col>
      </Row>

      {/* 移动端快捷操作抽屉 */}
      <Drawer
        title="快捷操作"
        placement="bottom"
        onClose={() => setMobileMenuVisible(false)}
        open={mobileMenuVisible}
        height={300}
      >
        <Space direction="vertical" size="large" style={{ width: '100%' }}>
          <Button 
            type="primary" 
            icon={<PlusOutlined />}
            block
            size="large"
            onClick={() => {
              setMobileMenuVisible(false);
              navigate('/community/forum');
            }}
          >
            发布帖子
          </Button>
          <Button 
            icon={<EditOutlined />}
            block
            size="large"
            onClick={() => {
              setMobileMenuVisible(false);
              navigate('/community/blog');
            }}
          >
            写博客
          </Button>
          <Button 
            icon={<CodeOutlined />}
            block
            size="large"
            onClick={() => {
              setMobileMenuVisible(false);
              navigate('/community/code');
            }}
          >
            分享代码
          </Button>
          <div style={{ borderTop: '1px solid #f0f0f0', paddingTop: 16 }}>
            <Space direction="vertical" style={{ width: '100%' }}>
              <Link 
                to="/community/forum"
                onClick={() => setMobileMenuVisible(false)}
              >
                查看全部论坛
              </Link>
              <Link 
                to="/community/blog"
                onClick={() => setMobileMenuVisible(false)}
              >
                查看全部博客
              </Link>
              <Link 
                to="/community/code"
                onClick={() => setMobileMenuVisible(false)}
              >
                查看全部代码
              </Link>
            </Space>
          </div>
        </Space>
      </Drawer>
        </LoadingState>
      </div>
    </ErrorBoundary>
  );
};

export default Community;