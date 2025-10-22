import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import {
  Row,
  Col,
  Card,
  Button,
  Tag,
  Rate,
  Divider,
  Tabs,
  Avatar,
  List,
  Spin,
  message,
  Modal,
  Image,
  Statistic,
  Space,
  Typography,
  Badge,
  Form,
  Input,
  Progress,
  Tooltip,
  Alert,
  Timeline,
  Breadcrumb,
} from 'antd';
import {
  DownloadOutlined,
  HeartOutlined,
  HeartFilled,
  ShareAltOutlined,
  FlagOutlined,
  UserOutlined,
  CalendarOutlined,
  EyeOutlined,
  StarOutlined,
  LikeOutlined,
  DislikeOutlined,
  HomeOutlined,
  AppstoreOutlined,
  CheckCircleOutlined,
  ExclamationCircleOutlined,
  ClockCircleOutlined,
  BugOutlined,
  GiftOutlined,
  SafetyOutlined,
} from '@ant-design/icons';
import { Plugin, Review, InstallationStatus } from '@/types';
import { pluginService } from '@/services/plugin';
import userService from '@/services/user';
import { useCartStore } from '@/stores/cart';
import { useDownloadStore } from '@/stores/download';
import { useAuthStore } from '@/stores/auth';
import DownloadManager from '@/components/DownloadManager';

const { Title, Paragraph, Text } = Typography;
const { TabPane } = Tabs;
const { TextArea } = Input;

// 扩展的插件类型，包含更多详细信息
interface ExtendedPlugin extends Plugin {
  longDescription?: string;
  screenshots?: string[];
  versions?: PluginVersion[];
  relatedPlugins?: Plugin[];
  developer: {
    id: string;
    name: string;
    email: string;
    avatar?: string;
    verified?: boolean;
  };
  downloadCount: number;
  compatibility: {
    minVersion: string;
    maxVersion?: string;
    platforms: string[];
  };
  security: {
    verified: boolean;
    lastScan: string;
    issues: number;
  };
  changelog?: string;
}

interface PluginVersion {
  version: string;
  releaseDate: string;
  changelog: string;
  downloadCount: number;
  size: number;
}

interface ExtendedReview extends Review {
  user: {
    id: string;
    username: string;
    avatar?: string;
    verified?: boolean;
  };
  helpful: number;
  reported: number;
  reply?: {
    content: string;
    author: string;
    createdAt: string;
  };
}

const PluginDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { addToCart, isInCart } = useCartStore();
  const { 
    downloadPlugin, 
    installPlugin, 
    updatePlugin, 
    uninstallPlugin,
    isPluginDownloading,
    isPluginInstalling,
    getInstallationStatus,
    getTasksByPlugin
  } = useDownloadStore();
  const { isAuthenticated } = useAuthStore();
  
  const [loading, setLoading] = useState(true);
  const [plugin, setPlugin] = useState<ExtendedPlugin | null>(null);
  const [reviews, setReviews] = useState<ExtendedReview[]>([]);
  const [isFavorited, setIsFavorited] = useState(false);
  const [isPurchased, setIsPurchased] = useState(false);
  const [installing, setInstalling] = useState(false);
  const [reviewForm] = Form.useForm();
  const [showReviewForm, setShowReviewForm] = useState(false);
  const [submittingReview, setSubmittingReview] = useState(false);
  const [downloadManagerVisible, setDownloadManagerVisible] = useState(false);
  
  // 获取插件的安装状态和下载状态
  const installationStatus = plugin ? getInstallationStatus(plugin.id) : 'not_installed';
  const isDownloading = plugin ? isPluginDownloading(plugin.id) : false;
  const isInstalling = plugin ? isPluginInstalling(plugin.id) : false;
  const pluginTasks = plugin ? getTasksByPlugin(plugin.id) : [];

  // Mock数据
  const mockPlugin: ExtendedPlugin = {
    id: id || '1',
    name: 'Advanced Code Editor',
    description: '功能强大的代码编辑器，支持多种编程语言，提供智能代码补全、语法高亮、代码格式化等功能',
    longDescription: `Advanced Code Editor 是一款专为开发者设计的高级代码编辑器插件。它集成了最新的语言服务器协议（LSP），为超过50种编程语言提供智能代码补全、实时错误检测、代码重构等功能。

主要特性：
• 智能代码补全：基于上下文的智能提示，支持函数签名、文档预览
• 语法高亮：支持自定义主题，提供丰富的颜色方案
• 代码格式化：一键格式化代码，支持多种代码风格配置
• 实时错误检测：即时发现语法错误和潜在问题
• 代码重构：安全的重命名、提取方法等重构操作
• 多光标编辑：提高编辑效率的多光标支持
• 代码折叠：智能代码折叠，支持自定义折叠规则
• 插件生态：丰富的扩展插件，满足不同开发需求

技术规格：
• 基于 Monaco Editor 核心
• 支持 Language Server Protocol (LSP)
• 内存占用优化，支持大文件编辑
• 跨平台兼容，支持 Windows、macOS、Linux`,
    version: '3.2.1',
    author: 'CodeMaster Studio',
    authorEmail: 'contact@codemaster.dev',
    category: 'development',
    tags: ['editor', 'ide', 'syntax-highlighting', 'autocomplete', 'refactoring'],
    icon: 'https://via.placeholder.com/128/4CAF50/FFFFFF?text=ACE',
    screenshots: [
      'https://via.placeholder.com/800x600/f0f0f0/333?text=Code+Editor+Interface',
      'https://via.placeholder.com/800x600/e8f5e8/333?text=Syntax+Highlighting',
      'https://via.placeholder.com/800x600/e3f2fd/333?text=Auto+Completion',
      'https://via.placeholder.com/800x600/fff3e0/333?text=Error+Detection',
    ],
    downloadUrl: '/api/plugins/1/download',
    documentationUrl: 'https://docs.codemaster.dev/ace',
    repositoryUrl: 'https://github.com/codemaster/ace-editor',
    license: 'MIT',
    price: 49.99,
    currency: 'CNY',
    downloads: 0,
    downloadCount: 28547,
    rating: 4.7,
    reviewCount: 342,
    size: 15728640, // 15MB
    requirements: {
      minVersion: '1.0.0',
      dependencies: ['typescript-language-server', 'eslint'],
    },
    compatibility: {
      minVersion: '1.0.0',
      maxVersion: '2.0.0',
      platforms: ['Windows', 'macOS', 'Linux'],
    },
    security: {
      verified: true,
      lastScan: '2024-01-20',
      issues: 0,
    },
    status: 'active',
    featured: true,
    createdAt: '2023-06-15T08:00:00Z',
    updatedAt: '2024-01-20T14:30:00Z',
    publishedAt: '2023-06-20T10:00:00Z',
    developer: {
      id: 'codemaster-studio',
      name: 'CodeMaster Studio',
      email: 'contact@codemaster.dev',
      avatar: 'https://via.placeholder.com/64/2196F3/FFFFFF?text=CM',
      verified: true,
    },
    versions: [
      {
        version: '3.2.1',
        releaseDate: '2024-01-20',
        changelog: '修复了代码补全的性能问题，优化了内存使用',
        downloadCount: 1247,
        size: 15728640,
      },
      {
        version: '3.2.0',
        releaseDate: '2024-01-15',
        changelog: '新增了 TypeScript 5.0 支持，改进了错误检测算法',
        downloadCount: 2156,
        size: 15654400,
      },
      {
        version: '3.1.5',
        releaseDate: '2024-01-10',
        changelog: '修复了在大文件中的性能问题，优化了语法高亮渲染',
        downloadCount: 3421,
        size: 15532800,
      },
    ],
    relatedPlugins: [
      {
        id: '2',
        name: 'Git Integration Pro',
        description: '强大的Git集成工具',
        version: '2.1.0',
        author: 'DevTools Inc.',
        category: 'development',
        tags: ['git', 'version-control'],
        icon: 'https://via.placeholder.com/64/FF9800/FFFFFF?text=GIT',
        downloadUrl: '',
        license: 'MIT',
        price: 29.99,
        currency: 'CNY',
        downloads: 15420,
        rating: 4.5,
        reviewCount: 189,
        size: 8388608,
        requirements: { minVersion: '1.0.0' },
        status: 'active',
        featured: false,
        createdAt: '2023-08-01T00:00:00Z',
        updatedAt: '2024-01-18T00:00:00Z',
      },
      {
        id: '3',
        name: 'Theme Designer',
        description: '自定义编辑器主题设计工具',
        version: '1.5.2',
        author: 'UI Masters',
        category: 'design',
        tags: ['theme', 'customization', 'ui'],
        icon: 'https://via.placeholder.com/64/9C27B0/FFFFFF?text=TD',
        downloadUrl: '',
        license: 'Apache-2.0',
        price: 0,
        currency: 'CNY',
        downloads: 9876,
        rating: 4.3,
        reviewCount: 76,
        size: 4194304,
        requirements: { minVersion: '1.0.0' },
        status: 'active',
        featured: false,
        createdAt: '2023-09-15T00:00:00Z',
        updatedAt: '2024-01-12T00:00:00Z',
      },
    ],
    changelog: '版本 3.2.1 更新内容：\n• 修复了代码补全在某些情况下的性能问题\n• 优化了内存使用，减少了50%的内存占用\n• 改进了错误提示的准确性\n• 新增了对 Python 3.12 的支持\n• 修复了主题切换时的显示问题',
  };

  const mockReviews: ExtendedReview[] = [
    {
      id: '1',
      pluginId: id || '1',
      userId: 'user1',
      userName: 'developer_pro',
      userAvatar: 'https://via.placeholder.com/40/4CAF50/FFFFFF?text=DP',
      rating: 5,
      title: '非常优秀的代码编辑器',
      content: '这是我用过最好的代码编辑器插件！智能补全功能非常强大，大大提高了我的开发效率。界面设计也很美观，推荐给所有开发者。',
      helpful: 23,
      reported: 0,
      createdAt: '2024-01-18T10:30:00Z',
      updatedAt: '2024-01-18T10:30:00Z',
      user: {
        id: 'user1',
        username: 'developer_pro',
        avatar: 'https://via.placeholder.com/40/4CAF50/FFFFFF?text=DP',
        verified: true,
      },
      reply: {
        content: '感谢您的反馈！我们会继续努力改进产品。',
        author: 'CodeMaster Studio',
        createdAt: '2024-01-18T15:20:00Z',
      },
    },
    {
      id: '2',
      pluginId: id || '1',
      userId: 'user2',
      userName: 'frontend_ninja',
      userAvatar: 'https://via.placeholder.com/40/2196F3/FFFFFF?text=FN',
      rating: 4,
      title: '功能丰富，但还有改进空间',
      content: '整体来说是一个很不错的编辑器，语法高亮和代码补全都很好用。不过在处理大文件时偶尔会有些卡顿，希望后续版本能优化一下性能。',
      helpful: 15,
      reported: 0,
      createdAt: '2024-01-15T14:20:00Z',
      updatedAt: '2024-01-15T14:20:00Z',
      user: {
        id: 'user2',
        username: 'frontend_ninja',
        avatar: 'https://via.placeholder.com/40/2196F3/FFFFFF?text=FN',
        verified: false,
      },
    },
    {
      id: '3',
      pluginId: id || '1',
      userId: 'user3',
      userName: 'backend_master',
      userAvatar: 'https://via.placeholder.com/40/FF9800/FFFFFF?text=BM',
      rating: 5,
      title: '完美的开发工具',
      content: '作为一个后端开发者，这个编辑器完全满足了我的需求。多语言支持很全面，错误检测也很准确。值得购买！',
      helpful: 31,
      reported: 0,
      createdAt: '2024-01-12T09:15:00Z',
      updatedAt: '2024-01-12T09:15:00Z',
      user: {
        id: 'user3',
        username: 'backend_master',
        avatar: 'https://via.placeholder.com/40/FF9800/FFFFFF?text=BM',
        verified: true,
      },
    },
  ];

  useEffect(() => {
    if (id) {
      loadPluginDetail();
    }
  }, [id]);

  const loadPluginDetail = async () => {
    try {
      setLoading(true);
      // 获取插件详情
      const pluginData = await pluginService.getPlugin(id!);
      setPlugin(pluginData);
      setReviews(mockReviews); // 暂时使用模拟评论数据
      
      // 检查是否已收藏（仅在用户已登录时）
      if (isAuthenticated) {
        try {
          const favoritesData = await userService.getFavoritePlugins();
          const isFav = favoritesData.data.some((fav: any) => fav.id === id);
          setIsFavorited(isFav);
        } catch (error) {
          // 如果获取收藏状态失败，默认为未收藏
          setIsFavorited(false);
        }

        // 检查是否已购买
        try {
          const purchasesData = await userService.getPurchases();
          const isPurch = purchasesData.data.some((purchase: any) => purchase.plugin_id === id);
          setIsPurchased(isPurch);
        } catch (error) {
          // 如果获取购买状态失败，默认为未购买
          setIsPurchased(false);
        }
      } else {
        setIsFavorited(false);
        setIsPurchased(false);
      }
    } catch (error) {
      message.error('加载插件详情失败');
      navigate('/404');
    } finally {
      setLoading(false);
    }
  };

  const handleDownload = async () => {
    if (!plugin) return;
    
    try {
      await downloadPlugin(plugin.id, plugin.name, plugin.icon);
      setDownloadManagerVisible(true);
    } catch (error) {
      message.error('下载失败，请稍后重试');
    }
  };

  const handleInstall = async () => {
    if (!plugin) return;
    
    if (plugin.price === 0) {
      // 免费插件直接下载安装
      await handleDownload();
    } else {
      // 付费插件需要先购买
      addToCart(plugin);
      message.success('已添加到购物车');
    }
  };

  const handleUpdate = async () => {
    if (!plugin) return;
    
    try {
      await updatePlugin(plugin.id, plugin.name, plugin.icon);
      setDownloadManagerVisible(true);
    } catch (error) {
      message.error('更新失败，请稍后重试');
    }
  };

  const handleUninstall = async () => {
    if (!plugin) return;
    
    Modal.confirm({
      title: '确认卸载',
      content: `您确定要卸载 ${plugin.name} 吗？`,
      onOk: async () => {
        try {
          await uninstallPlugin(plugin.id, plugin.name);
          setDownloadManagerVisible(true);
        } catch (error) {
          message.error('卸载失败，请稍后重试');
        }
      },
    });
  };

  const handleAddToCart = async () => {
    if (!plugin) return;
    
    if (plugin.price === 0) {
      handleInstall();
    } else {
      // 付费插件直接购买
      if (!isAuthenticated) {
        message.warning('请先登录');
        navigate('/login');
        return;
      }
      
      try {
        setLoading(true);
        await userService.purchasePlugin(plugin.id);
        setIsPurchased(true); // 更新购买状态
        message.success('购买成功！');
        // 购买成功后可以直接下载安装
        await handleDownload();
      } catch (error: any) {
        if (error.response?.status === 400) {
          message.warning('您已购买过此插件');
        } else {
          message.error('购买失败，请稍后重试');
        }
      } finally {
        setLoading(false);
      }
    }
  };

  const handleFavorite = async () => {
    if (!plugin) return;
    
    try {
      const result = await userService.toggleFavorite(plugin.id);
      setIsFavorited(result.is_favorited);
      message.success(result.message || (result.is_favorited ? '已添加到收藏' : '已取消收藏'));
    } catch (error) {
      message.error('操作失败，请稍后重试');
    }
  };

  const handleShare = () => {
    const url = window.location.href;
    navigator.clipboard.writeText(url).then(() => {
      message.success('链接已复制到剪贴板');
    });
  };

  const handleReport = () => {
    Modal.confirm({
      title: '举报插件',
      content: '您确定要举报这个插件吗？我们会尽快处理您的举报。',
      onOk: async () => {
        try {
          message.success('举报已提交，感谢您的反馈');
        } catch (error) {
          message.error('举报失败，请稍后重试');
        }
      },
    });
  };

  const handleSubmitReview = async (values: any) => {
    try {
      setSubmittingReview(true);
      // 模拟提交评论
      await new Promise(resolve => setTimeout(resolve, 1000));
      
      const newReview: ExtendedReview = {
        id: Date.now().toString(),
        pluginId: plugin!.id,
        userId: 'current-user',
        userName: 'current_user',
        userAvatar: 'https://via.placeholder.com/40/607D8B/FFFFFF?text=CU',
        rating: values.rating,
        title: values.title,
        content: values.content,
        helpful: 0,
        reported: 0,
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
        user: {
          id: 'current-user',
          username: 'current_user',
          avatar: 'https://via.placeholder.com/40/607D8B/FFFFFF?text=CU',
          verified: false,
        },
      };
      
      setReviews([newReview, ...reviews]);
      setShowReviewForm(false);
      reviewForm.resetFields();
      message.success('评价提交成功！');
    } catch (error) {
      message.error('提交失败，请稍后重试');
    } finally {
      setSubmittingReview(false);
    }
  };

  const handleHelpful = (reviewId: string, helpful: boolean) => {
    setReviews(reviews.map(review => 
      review.id === reviewId 
        ? { ...review, helpful: helpful ? review.helpful + 1 : Math.max(0, review.helpful - 1) }
        : review
    ));
    message.success(helpful ? '感谢您的反馈' : '已取消');
  };

  if (loading) {
    return (
      <div style={{ 
        display: 'flex', 
        justifyContent: 'center', 
        alignItems: 'center', 
        minHeight: '400px',
        flexDirection: 'column',
        gap: '16px'
      }}>
        <Spin size="large" />
        <Text type="secondary">加载插件详情中...</Text>
      </div>
    );
  }

  if (!plugin) {
    return null;
  }

  return (
    <div style={{ padding: '24px', maxWidth: '1200px', margin: '0 auto' }}>
      {/* 面包屑导航 */}
      <Breadcrumb style={{ marginBottom: '24px' }}>
        <Breadcrumb.Item>
          <HomeOutlined />
          <span onClick={() => navigate('/')} style={{ cursor: 'pointer' }}>首页</span>
        </Breadcrumb.Item>
        <Breadcrumb.Item>
          <AppstoreOutlined />
          <span onClick={() => navigate(`/category/${plugin.category}`)} style={{ cursor: 'pointer' }}>
            {plugin.category === 'development' ? '开发工具' : plugin.category}
          </span>
        </Breadcrumb.Item>
        <Breadcrumb.Item>{plugin.name}</Breadcrumb.Item>
      </Breadcrumb>

      <Row gutter={[24, 24]}>
        {/* 左侧主要内容 */}
        <Col xs={24} lg={16}>
          <Card>
            {/* 插件头部信息 */}
            <div style={{ display: 'flex', gap: '16px', marginBottom: '24px' }}>
              <Avatar
                size={80}
                src={plugin.icon}
                icon={<UserOutlined />}
                style={{ flexShrink: 0 }}
              />
              <div style={{ flex: 1 }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: '8px', marginBottom: '8px' }}>
                  <Title level={2} style={{ margin: 0 }}>
                    {plugin.name}
                  </Title>
                  {plugin.security.verified && (
                    <Tooltip title="已通过安全验证">
                      <Badge count={<SafetyOutlined style={{ color: '#52c41a' }} />} />
                    </Tooltip>
                  )}
                  {plugin.featured && (
                    <Tooltip title="精选插件">
                      <Badge count={<GiftOutlined style={{ color: '#faad14' }} />} />
                    </Tooltip>
                  )}
                </div>
                <Text type="secondary" style={{ fontSize: '16px', display: 'block', marginBottom: '12px' }}>
                  {plugin.description}
                </Text>
                <div style={{ marginBottom: '12px' }}>
                  <Space>
                    <Rate disabled value={plugin.rating} allowHalf />
                    <Text>({plugin.reviewCount} 评价)</Text>
                    <Text type="secondary">•</Text>
                    <Text>{plugin.downloadCount.toLocaleString()} 下载</Text>
                    <Text type="secondary">•</Text>
                    <Text>版本 {plugin.version}</Text>
                  </Space>
                </div>
                <div>
                  {plugin.tags?.map(tag => (
                    <Tag key={tag} color="blue" style={{ marginBottom: '4px' }}>{tag}</Tag>
                  ))}
                </div>
              </div>
            </div>

            {/* 操作按钮 */}
            <Space size="middle" style={{ marginBottom: '24px', flexWrap: 'wrap' }}>
              {/* 主要操作按钮 */}
              {installationStatus === 'not_installed' && (
                <Button
                  type="primary"
                  size="large"
                  icon={<DownloadOutlined />}
                  loading={loading || isDownloading || isInstalling}
                  onClick={handleAddToCart}
                  disabled={isInCart(plugin.id) || (plugin.price > 0 && isPurchased)}
                >
                  {plugin.price === 0 ? 
                    (isDownloading ? '下载中...' : isInstalling ? '安装中...' : '免费安装') : 
                    isPurchased ? '已购买' :
                    loading ? '购买中...' : isInCart(plugin.id) ? '已在购物车' : `¥${plugin.price} 购买`}
                </Button>
              )}
              
              {/* 已购买但未安装的插件显示安装按钮 */}
              {installationStatus === 'not_installed' && plugin.price > 0 && isPurchased && (
                <Button
                  type="default"
                  size="large"
                  icon={<DownloadOutlined />}
                  loading={isDownloading || isInstalling}
                  onClick={handleDownload}
                  style={{ marginLeft: '8px' }}
                >
                  {isDownloading ? '下载中...' : isInstalling ? '安装中...' : '安装'}
                </Button>
              )}
              
              {installationStatus === 'installed' && (
                <>
                  <Button
                    type="primary"
                    size="large"
                    icon={<CheckCircleOutlined />}
                    disabled
                  >
                    已安装
                  </Button>
                  <Button
                    size="large"
                    icon={<DownloadOutlined />}
                    loading={isDownloading || isInstalling}
                    onClick={handleUpdate}
                  >
                    {isDownloading || isInstalling ? '更新中...' : '更新'}
                  </Button>
                  <Button
                    size="large"
                    danger
                    onClick={handleUninstall}
                  >
                    卸载
                  </Button>
                </>
              )}
              
              {installationStatus === 'installing' && (
                <Button
                  type="primary"
                  size="large"
                  icon={<ClockCircleOutlined />}
                  loading
                  disabled
                >
                  安装中...
                </Button>
              )}
              
              {installationStatus === 'updating' && (
                <Button
                  type="primary"
                  size="large"
                  icon={<ClockCircleOutlined />}
                  loading
                  disabled
                >
                  更新中...
                </Button>
              )}
              
              {installationStatus === 'failed' && (
                <Button
                  type="primary"
                  size="large"
                  icon={<ExclamationCircleOutlined />}
                  danger
                  onClick={handleInstall}
                >
                  重新安装
                </Button>
              )}
              
              {/* 下载管理器按钮 */}
              {pluginTasks.length > 0 && (
                <Button
                  icon={<DownloadOutlined />}
                  onClick={() => setDownloadManagerVisible(true)}
                >
                  查看进度 ({pluginTasks.length})
                </Button>
              )}
              
              <Button
                icon={isFavorited ? <HeartFilled /> : <HeartOutlined />}
                onClick={handleFavorite}
                style={{ color: isFavorited ? '#ff4d4f' : undefined }}
              >
                {isFavorited ? '已收藏' : '收藏'}
              </Button>
              
              <Button icon={<ShareAltOutlined />} onClick={handleShare}>
                分享
              </Button>
              
              <Button icon={<FlagOutlined />} onClick={handleReport}>
                举报
              </Button>
            </Space>

            {/* 安全信息提示 */}
            {plugin.security.verified ? (
              <Alert
                message="安全验证"
                description={`此插件已通过安全扫描，最后扫描时间：${plugin.security.lastScan}，未发现安全问题。`}
                type="success"
                showIcon
                style={{ marginBottom: '24px' }}
              />
            ) : (
              <Alert
                message="安全提示"
                description="此插件尚未通过安全验证，请谨慎使用。"
                type="warning"
                showIcon
                style={{ marginBottom: '24px' }}
              />
            )}

            <Divider />

            {/* 详细信息标签页 */}
            <Tabs defaultActiveKey="description">
              <TabPane tab="详细介绍" key="description">
                <div style={{ lineHeight: '1.8' }}>
                  <Paragraph style={{ whiteSpace: 'pre-line', fontSize: '14px' }}>
                    {plugin.longDescription || plugin.description}
                  </Paragraph>
                  
                  {plugin.screenshots && plugin.screenshots.length > 0 && (
                    <div style={{ marginTop: '32px' }}>
                      <Title level={4}>截图预览</Title>
                      <Image.PreviewGroup>
                        <Row gutter={[16, 16]}>
                          {plugin.screenshots.map((screenshot, index) => (
                            <Col key={index} xs={24} sm={12} md={8}>
                              <Image
                                src={screenshot}
                                alt={`截图 ${index + 1}`}
                                style={{ 
                                  width: '100%', 
                                  borderRadius: '8px',
                                  border: '1px solid #f0f0f0'
                                }}
                              />
                            </Col>
                          ))}
                        </Row>
                      </Image.PreviewGroup>
                    </div>
                  )}

                  {/* 兼容性信息 */}
                  <div style={{ marginTop: '32px' }}>
                    <Title level={4}>兼容性信息</Title>
                    <Row gutter={[16, 16]}>
                      <Col span={12}>
                        <Text strong>支持平台：</Text>
                        <div style={{ marginTop: '8px' }}>
                          {plugin.compatibility.platforms.map(platform => (
                            <Tag key={platform} color="green">{platform}</Tag>
                          ))}
                        </div>
                      </Col>
                      <Col span={12}>
                        <Text strong>版本要求：</Text>
                        <div style={{ marginTop: '8px' }}>
                          <Text>{plugin.compatibility.minVersion}+</Text>
                        </div>
                      </Col>
                    </Row>
                  </div>
                </div>
              </TabPane>
              
              <TabPane tab={`评价 (${reviews.length})`} key="reviews">
                <div style={{ marginBottom: '24px' }}>
                  <Button 
                    type="primary" 
                    onClick={() => setShowReviewForm(!showReviewForm)}
                    style={{ marginBottom: '16px' }}
                  >
                    {showReviewForm ? '取消评价' : '写评价'}
                  </Button>
                  
                  {showReviewForm && (
                    <Card style={{ marginBottom: '24px' }}>
                      <Form
                        form={reviewForm}
                        layout="vertical"
                        onFinish={handleSubmitReview}
                      >
                        <Form.Item
                          name="rating"
                          label="评分"
                          rules={[{ required: true, message: '请选择评分' }]}
                        >
                          <Rate />
                        </Form.Item>
                        <Form.Item
                          name="title"
                          label="标题"
                          rules={[{ required: true, message: '请输入评价标题' }]}
                        >
                          <Input placeholder="请输入评价标题" />
                        </Form.Item>
                        <Form.Item
                          name="content"
                          label="内容"
                          rules={[{ required: true, message: '请输入评价内容' }]}
                        >
                          <TextArea 
                            rows={4} 
                            placeholder="请详细描述您的使用体验..."
                            showCount
                            maxLength={500}
                          />
                        </Form.Item>
                        <Form.Item>
                          <Space>
                            <Button type="primary" htmlType="submit" loading={submittingReview}>
                              提交评价
                            </Button>
                            <Button onClick={() => setShowReviewForm(false)}>
                              取消
                            </Button>
                          </Space>
                        </Form.Item>
                      </Form>
                    </Card>
                  )}
                </div>

                <List
                  dataSource={reviews}
                  renderItem={(review) => (
                    <Card 
                      style={{ marginBottom: '16px' }}
                      size="small"
                    >
                      <div style={{ display: 'flex', alignItems: 'flex-start', gap: '12px' }}>
                        <Avatar src={review.user.avatar} icon={<UserOutlined />} />
                        <div style={{ flex: 1 }}>
                          <div style={{ marginBottom: '8px' }}>
                            <Space>
                              <Text strong>{review.user.username}</Text>
                              {review.user.verified && (
                                <Badge count={<CheckCircleOutlined style={{ color: '#52c41a' }} />} />
                              )}
                            </Space>
                          </div>
                          <div style={{ marginBottom: '8px' }}>
                            <Rate disabled value={review.rating} size="small" />
                            <Text type="secondary" style={{ marginLeft: '8px' }}>
                              {new Date(review.createdAt).toLocaleDateString()}
                            </Text>
                          </div>
                          <div style={{ marginBottom: '8px' }}>
                            <Text strong style={{ fontSize: '16px' }}>{review.title}</Text>
                          </div>
                          <Paragraph>{review.content}</Paragraph>
                          {review.reply && (
                            <div style={{ 
                              background: '#f9f9f9', 
                              padding: '12px', 
                              borderRadius: '6px',
                              marginTop: '12px'
                            }}>
                              <Text strong>{review.reply.author} 回复：</Text>
                              <Paragraph style={{ margin: '4px 0 0 0' }}>{review.reply.content}</Paragraph>
                              <Text type="secondary" style={{ fontSize: '12px' }}>
                                {new Date(review.reply.createdAt).toLocaleDateString()}
                              </Text>
                            </div>
                          )}
                          <div style={{ marginTop: '12px' }}>
                            <Space>
                              <Button 
                                type="text" 
                                icon={<LikeOutlined />}
                                onClick={() => handleHelpful(review.id, true)}
                              >
                                有用 ({review.helpful})
                              </Button>
                              <Button 
                                type="text" 
                                icon={<FlagOutlined />}
                                onClick={() => message.info('举报功能开发中')}
                              >
                                举报
                              </Button>
                            </Space>
                          </div>
                        </div>
                      </div>
                    </Card>
                  )}
                />
              </TabPane>
              
              <TabPane tab="版本历史" key="versions">
                <Timeline>
                  {plugin.versions?.map((version, index) => (
                    <Timeline.Item
                      key={version.version}
                      dot={index === 0 ? <ClockCircleOutlined style={{ color: '#1890ff' }} /> : undefined}
                      color={index === 0 ? 'blue' : 'gray'}
                    >
                      <div style={{ marginBottom: '16px' }}>
                        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '8px' }}>
                          <Text strong style={{ fontSize: '16px' }}>版本 {version.version}</Text>
                          {index === 0 && <Tag color="blue">最新版本</Tag>}
                        </div>
                        <Text type="secondary" style={{ display: 'block', marginBottom: '8px' }}>
                          发布时间：{version.releaseDate} • 下载：{version.downloadCount.toLocaleString()} • 大小：{(version.size / 1024 / 1024).toFixed(1)}MB
                        </Text>
                        <Paragraph>{version.changelog}</Paragraph>
                      </div>
                    </Timeline.Item>
                  ))}
                </Timeline>
              </TabPane>

              <TabPane tab="更新日志" key="changelog">
                <div style={{ lineHeight: '1.8' }}>
                  <Title level={4}>最新更新 (v{plugin.version})</Title>
                  <Paragraph style={{ whiteSpace: 'pre-line' }}>
                    {plugin.changelog}
                  </Paragraph>
                </div>
              </TabPane>
            </Tabs>
          </Card>
        </Col>

        {/* 右侧信息栏 */}
        <Col xs={24} lg={8}>
          {/* 插件统计信息 */}
          <Card title="插件信息" style={{ marginBottom: '16px' }}>
            <Row gutter={[16, 16]}>
              <Col span={12}>
                <Statistic
                  title="下载量"
                  value={plugin.downloadCount}
                  prefix={<DownloadOutlined />}
                />
              </Col>
              <Col span={12}>
                <Statistic
                  title="评分"
                  value={plugin.rating}
                  precision={1}
                  suffix="/ 5"
                  prefix={<StarOutlined />}
                />
              </Col>
              <Col span={12}>
                <Statistic
                  title="版本"
                  value={plugin.version}
                />
              </Col>
              <Col span={12}>
                <Statistic
                  title="大小"
                  value={(plugin.size / 1024 / 1024).toFixed(1)}
                  suffix="MB"
                />
              </Col>
            </Row>
            
            <Divider />
            
            {/* 开发者信息 */}
            <div>
              <Text strong>开发者：</Text>
              <div style={{ marginTop: '8px', marginBottom: '16px' }}>
                <Space>
                  <Avatar src={plugin.developer.avatar} icon={<UserOutlined />} />
                  <div>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '4px' }}>
                      <Text strong>{plugin.developer.name}</Text>
                      {plugin.developer.verified && (
                        <CheckCircleOutlined style={{ color: '#52c41a', fontSize: '12px' }} />
                      )}
                    </div>
                    <Text type="secondary" style={{ fontSize: '12px' }}>
                      {plugin.developer.email}
                    </Text>
                  </div>
                </Space>
              </div>
              <Button 
                size="small" 
                onClick={() => navigate(`/developer/${plugin.developer.id}`)}
              >
                查看开发者
              </Button>
            </div>
            
            <Divider />
            
            {/* 发布信息 */}
            <div>
              <div style={{ marginBottom: '8px' }}>
                <CalendarOutlined /> <Text>发布时间：{new Date(plugin.createdAt).toLocaleDateString()}</Text>
              </div>
              <div style={{ marginBottom: '8px' }}>
                <CalendarOutlined /> <Text>更新时间：{new Date(plugin.updatedAt).toLocaleDateString()}</Text>
              </div>
              <div style={{ marginBottom: '8px' }}>
                <Text>许可证：</Text>
                <Tag color="blue">{plugin.license}</Tag>
              </div>
              <div>
                <Text>分类：</Text>
                <Tag 
                  color="blue" 
                  style={{ cursor: 'pointer' }}
                  onClick={() => navigate(`/category/${plugin.category}`)}
                >
                  {plugin.category === 'development' ? '开发工具' : plugin.category}
                </Tag>
              </div>
            </div>
          </Card>

          {/* 相关插件推荐 */}
          {plugin.relatedPlugins && plugin.relatedPlugins.length > 0 && (
            <Card title="相关插件推荐">
              <List
                dataSource={plugin.relatedPlugins}
                renderItem={(relatedPlugin) => (
                  <List.Item
                    style={{ cursor: 'pointer', padding: '12px 0' }}
                    onClick={() => navigate(`/plugin/${relatedPlugin.id}`)}
                  >
                    <List.Item.Meta
                      avatar={<Avatar src={relatedPlugin.icon} icon={<UserOutlined />} />}
                      title={
                        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                          <Text strong ellipsis style={{ flex: 1 }}>{relatedPlugin.name}</Text>
                          <Text type={relatedPlugin.price === 0 ? 'success' : 'danger'} style={{ fontSize: '12px' }}>
                            {relatedPlugin.price === 0 ? '免费' : `¥${relatedPlugin.price}`}
                          </Text>
                        </div>
                      }
                      description={
                        <div>
                          <Text type="secondary" style={{ fontSize: '12px', display: 'block' }} ellipsis>
                            {relatedPlugin.description}
                          </Text>
                          <div style={{ marginTop: '4px', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                            <Rate disabled value={relatedPlugin.rating} size="small" />
                            <Text type="secondary" style={{ fontSize: '11px' }}>
                              {relatedPlugin.downloads.toLocaleString()} 下载
                            </Text>
                          </div>
                        </div>
                      }
                    />
                  </List.Item>
                )}
              />
            </Card>
          )}
        </Col>
      </Row>
      
      {/* 下载管理器 */}
      <DownloadManager
        visible={downloadManagerVisible}
        onClose={() => setDownloadManagerVisible(false)}
      />
    </div>
  );
};

export default PluginDetail;