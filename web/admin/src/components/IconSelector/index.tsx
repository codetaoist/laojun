import React, { useState, useEffect, useMemo } from 'react';
import {
  Modal,
  Input,
  Tabs,
  Row,
  Col,
  Button,
  Upload,
  message,
  Spin,
  Empty,
  Tooltip,
  Pagination,
  Select,
  Space,
} from 'antd';
import {
  SearchOutlined,
  UploadOutlined,
  DeleteOutlined,
  StarOutlined,
  StarFilled,
  AppstoreOutlined,
  UnorderedListOutlined,
} from '@ant-design/icons';
import * as AntdIcons from '@ant-design/icons';
import { iconService, IconLibraryItem } from '@/services/menu';
import './index.less';

const { Search } = Input;
const { TabPane } = Tabs;
const { Option } = Select;

export interface IconSelectorProps {
  visible: boolean;
  onCancel: () => void;
  onSelect: (icon: string) => void;
  selectedIcon?: string;
  allowUpload?: boolean;
}

interface IconCategory {
  key: string;
  label: string;
  icons: string[];
}

const IconSelector: React.FC<IconSelectorProps> = ({
  visible,
  onCancel,
  onSelect,
  selectedIcon,
  allowUpload = true,
}) => {
  const [activeTab, setActiveTab] = useState('antd');
  const [searchKeyword, setSearchKeyword] = useState('');
  const [customIcons, setCustomIcons] = useState<IconLibraryItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [uploading, setUploading] = useState(false);
  const [viewMode, setViewMode] = useState<'grid' | 'list'>('grid');
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(24);
  const [sortBy, setSortBy] = useState<'name' | 'usage' | 'created'>('name');
  const [favorites, setFavorites] = useState<string[]>([]);

  // Ant Design 图标分类
  const antdIconCategories: IconCategory[] = [
    {
      key: 'direction',
      label: '方向性图标',
      icons: [
        'StepBackwardOutlined', 'StepForwardOutlined', 'FastBackwardOutlined', 'FastForwardOutlined',
        'ShrinkOutlined', 'ArrowsAltOutlined', 'DownOutlined', 'UpOutlined', 'LeftOutlined', 'RightOutlined',
        'CaretUpOutlined', 'CaretDownOutlined', 'CaretLeftOutlined', 'CaretRightOutlined',
        'UpCircleOutlined', 'DownCircleOutlined', 'LeftCircleOutlined', 'RightCircleOutlined',
        'DoubleRightOutlined', 'DoubleLeftOutlined', 'VerticalRightOutlined', 'VerticalLeftOutlined',
        'VerticalAlignTopOutlined', 'VerticalAlignMiddleOutlined', 'VerticalAlignBottomOutlined',
      ],
    },
    {
      key: 'suggestion',
      label: '提示建议性图标',
      icons: [
        'QuestionOutlined', 'QuestionCircleOutlined', 'PlusOutlined', 'PlusCircleOutlined',
        'PauseOutlined', 'PauseCircleOutlined', 'MinusOutlined', 'MinusCircleOutlined',
        'PlusSquareOutlined', 'MinusSquareOutlined', 'InfoOutlined', 'InfoCircleOutlined',
        'ExclamationOutlined', 'ExclamationCircleOutlined', 'CloseOutlined', 'CloseCircleOutlined',
        'CloseSquareOutlined', 'CheckOutlined', 'CheckCircleOutlined', 'CheckSquareOutlined',
      ],
    },
    {
      key: 'editor',
      label: '编辑类图标',
      icons: [
        'EditOutlined', 'FormOutlined', 'CopyOutlined', 'ScissorOutlined', 'DeleteOutlined',
        'SnippetsOutlined', 'DiffOutlined', 'HighlightOutlined', 'AlignCenterOutlined',
        'AlignLeftOutlined', 'AlignRightOutlined', 'BgColorsOutlined', 'BoldOutlined',
        'ItalicOutlined', 'UnderlineOutlined', 'StrikethroughOutlined', 'RedoOutlined', 'UndoOutlined',
        'ZoomInOutlined', 'ZoomOutOutlined', 'FontColorsOutlined', 'FontSizeOutlined',
      ],
    },
    {
      key: 'data',
      label: '数据类图标',
      icons: [
        'AreaChartOutlined', 'PieChartOutlined', 'BarChartOutlined', 'DotChartOutlined',
        'LineChartOutlined', 'RadarChartOutlined', 'HeatMapOutlined', 'FallOutlined', 'RiseOutlined',
        'StockOutlined', 'BoxPlotOutlined', 'FundOutlined', 'SlidersOutlined',
      ],
    },
    {
      key: 'brand',
      label: '品牌和标识',
      icons: [
        'AndroidOutlined', 'AppleOutlined', 'WindowsOutlined', 'IeOutlined', 'ChromeOutlined',
        'GithubOutlined', 'AliwangwangOutlined', 'DingdingOutlined', 'WeiboSquareOutlined',
        'WeiboCircleOutlined', 'TaobaoCircleOutlined', 'Html5Outlined', 'WechatOutlined',
        'YoutubeOutlined', 'AlipayCircleOutlined', 'TaobaoOutlined', 'SkypeOutlined',
        'QqOutlined', 'MediumWorkmarkOutlined', 'GitlabOutlined', 'MediumOutlined',
      ],
    },
    {
      key: 'application',
      label: '网站通用图标',
      icons: [
        'AccountBookOutlined', 'AimOutlined', 'AlertOutlined', 'ApartmentOutlined',
        'ApiOutlined', 'AppstoreAddOutlined', 'AppstoreOutlined', 'AudioOutlined',
        'AudioMutedOutlined', 'AuditOutlined', 'BackwardOutlined', 'BankOutlined',
        'BarcodeOutlined', 'BarsOutlined', 'BellOutlined', 'BlockOutlined',
        'BookOutlined', 'BorderOutlined', 'BranchesOutlined', 'BugOutlined',
        'BuildOutlined', 'BulbOutlined', 'CalculatorOutlined', 'CalendarOutlined',
        'CameraOutlined', 'CarOutlined', 'CarryOutOutlined', 'CloudOutlined',
        'CodeOutlined', 'CompassOutlined', 'CompressOutlined', 'ContactsOutlined',
        'ContainerOutlined', 'ControlOutlined', 'CopyrightCircleOutlined', 'CopyrightOutlined',
        'CreditCardOutlined', 'CrownOutlined', 'CustomerServiceOutlined', 'DashboardOutlined',
        'DatabaseOutlined', 'DeleteColumnOutlined', 'DeleteRowOutlined', 'DeliveredProcedureOutlined',
        'DeploymentUnitOutlined', 'DesktopOutlined', 'DisconnectOutlined', 'DislikeOutlined',
        'DollarCircleOutlined', 'DollarOutlined', 'DownloadOutlined', 'EllipsisOutlined',
        'EnvironmentOutlined', 'EuroCircleOutlined', 'EuroOutlined', 'ExceptionOutlined',
        'ExpandAltOutlined', 'ExpandOutlined', 'ExperimentOutlined', 'ExportOutlined',
        'EyeOutlined', 'EyeInvisibleOutlined', 'FieldBinaryOutlined', 'FieldNumberOutlined',
        'FieldStringOutlined', 'FieldTimeOutlined', 'FileAddOutlined', 'FileDoneOutlined',
        'FileExcelOutlined', 'FileExclamationOutlined', 'FileImageOutlined', 'FileJpgOutlined',
        'FileMarkdownOutlined', 'FileOutlined', 'FilePdfOutlined', 'FilePptOutlined',
        'FileProtectOutlined', 'FileSearchOutlined', 'FileSyncOutlined', 'FileTextOutlined',
        'FileUnknownOutlined', 'FileWordOutlined', 'FileZipOutlined', 'FilterOutlined',
        'FireOutlined', 'FlagOutlined', 'FolderAddOutlined', 'FolderOpenOutlined',
        'FolderOutlined', 'ForkOutlined', 'FormatPainterOutlined', 'ForwardOutlined',
        'FunctionOutlined', 'FunnelPlotOutlined', 'GatewayOutlined', 'GifOutlined',
        'GiftOutlined', 'GlobalOutlined', 'GoldOutlined', 'GroupOutlined',
        'HddOutlined', 'HeartOutlined', 'HistoryOutlined', 'HomeOutlined',
        'HourglassOutlined', 'IdcardOutlined', 'ImportOutlined', 'InboxOutlined',
        'InsertRowAboveOutlined', 'InsertRowBelowOutlined', 'InsertRowLeftOutlined', 'InsertRowRightOutlined',
        'InsuranceOutlined', 'InteractionOutlined', 'KeyOutlined', 'LaptopOutlined',
        'LayoutOutlined', 'LikeOutlined', 'LineOutlined', 'LinkOutlined',
        'LoadingOutlined', 'LockOutlined', 'MailOutlined', 'ManOutlined',
        'MedicineBoxOutlined', 'MehOutlined', 'MenuFoldOutlined', 'MenuUnfoldOutlined',
        'MenuOutlined', 'MergeCellsOutlined', 'MessageOutlined', 'MobileOutlined',
        'MoneyCollectOutlined', 'MonitorOutlined', 'MoreOutlined', 'NodeCollapseOutlined',
        'NodeExpandOutlined', 'NodeIndexOutlined', 'NotificationOutlined', 'NumberOutlined',
        'PaperClipOutlined', 'PartitionOutlined', 'PayCircleOutlined', 'PercentageOutlined',
        'PhoneOutlined', 'PictureOutlined', 'PlayCircleOutlined', 'PlaySquareOutlined',
        'PoundCircleOutlined', 'PoundOutlined', 'PoweroffOutlined', 'PrinterOutlined',
        'ProfileOutlined', 'ProjectOutlined', 'PropertySafetyOutlined', 'PullRequestOutlined',
        'PushpinOutlined', 'QrcodeOutlined', 'ReadOutlined', 'ReconciliationOutlined',
        'RedEnvelopeOutlined', 'ReloadOutlined', 'RestOutlined', 'RobotOutlined',
        'RocketOutlined', 'SafetyCertificateOutlined', 'SafetyOutlined', 'ScanOutlined',
        'ScheduleOutlined', 'SearchOutlined', 'SecurityScanOutlined', 'SelectOutlined',
        'SendOutlined', 'SettingOutlined', 'ShakeOutlined', 'ShareAltOutlined',
        'ShopOutlined', 'ShoppingCartOutlined', 'ShoppingOutlined', 'SisternodeOutlined',
        'SkinOutlined', 'SmileOutlined', 'SolutionOutlined', 'SoundOutlined',
        'SplitCellsOutlined', 'StarOutlined', 'SubnodeOutlined', 'SyncOutlined',
        'TableOutlined', 'TabletOutlined', 'TagOutlined', 'TagsOutlined',
        'TeamOutlined', 'ThunderboltOutlined', 'ToTopOutlined', 'ToolOutlined',
        'TrademarkCircleOutlined', 'TrademarkOutlined', 'TransactionOutlined', 'TrophyOutlined',
        'UngroupOutlined', 'UnlockOutlined', 'UploadOutlined', 'UsbOutlined',
        'UserAddOutlined', 'UserDeleteOutlined', 'UserOutlined', 'UserSwitchOutlined',
        'UsergroupAddOutlined', 'UsergroupDeleteOutlined', 'VideoCameraOutlined', 'WalletOutlined',
        'WifiOutlined', 'BorderlessTableOutlined', 'WomanOutlined', 'BehanceOutlined',
        'DropboxOutlined', 'DeploymentUnitOutlined', 'UpCircleOutlined', 'DownCircleOutlined',
        'LeftCircleOutlined', 'RightCircleOutlined', 'PlayCircleOutlined', 'QuestionCircleOutlined',
        'PlusCircleOutlined', 'PlusSquareOutlined', 'MinusSquareOutlined', 'MinusCircleOutlined',
        'InfoCircleOutlined', 'ExclamationCircleOutlined', 'CloseCircleOutlined', 'CloseSquareOutlined',
        'CheckCircleOutlined', 'CheckSquareOutlined', 'ClockCircleOutlined', 'WarningOutlined',
        'IssuesCloseOutlined', 'StopOutlined',
      ],
    },
  ];

  // 获取所有 Ant Design 图标
  const allAntdIcons = useMemo(() => {
    const iconNames = Object.keys(AntdIcons).filter(
      (name) => name.endsWith('Outlined') || name.endsWith('Filled') || name.endsWith('TwoTone')
    );
    return iconNames;
  }, []);

  // 过滤图标
  const filteredAntdIcons = useMemo(() => {
    if (!searchKeyword) return allAntdIcons;
    return allAntdIcons.filter((iconName) =>
      iconName.toLowerCase().includes(searchKeyword.toLowerCase())
    );
  }, [allAntdIcons, searchKeyword]);

  // 过滤自定义图标
  const filteredCustomIcons = useMemo(() => {
    if (!searchKeyword) return customIcons;
    return customIcons.filter((icon) =>
      icon.name.toLowerCase().includes(searchKeyword.toLowerCase()) ||
      (icon.tags && icon.tags.some(tag => tag.toLowerCase().includes(searchKeyword.toLowerCase())))
    );
  }, [customIcons, searchKeyword]);

  // 分页处理
  const paginatedIcons = useMemo(() => {
    const icons = activeTab === 'antd' ? filteredAntdIcons : filteredCustomIcons;
    const startIndex = (currentPage - 1) * pageSize;
    const endIndex = startIndex + pageSize;
    return icons.slice(startIndex, endIndex);
  }, [activeTab, filteredAntdIcons, filteredCustomIcons, currentPage, pageSize]);

  // 加载自定义图标
  const loadCustomIcons = async () => {
    try {
      setLoading(true);
      const response = await iconService.getIcons({
        keyword: searchKeyword,
        sort_by: sortBy,
        page: 1,
        page_size: 1000,
      });
      setCustomIcons(response.data.items);
    } catch (error: any) {
      // 如果是404错误，说明后端还没有实现图标API，静默处理
      if (error?.response?.status === 404) {
        console.warn('图标API尚未实现，暂时只支持Ant Design图标');
        setCustomIcons([]);
      } else {
        message.error('加载自定义图标失败');
      }
    } finally {
      setLoading(false);
    }
  };

  // 上传图标
  const handleUpload = async (file: File) => {
    try {
      setUploading(true);
      await iconService.uploadIcon(file, file.name.replace(/\.[^/.]+$/, ''), undefined, []);
      message.success('图标上传成功');
      loadCustomIcons();
    } catch (error: any) {
      if (error?.response?.status === 404) {
        message.warning('图标上传功能暂未开放，请联系管理员');
      } else {
        message.error('图标上传失败');
      }
    } finally {
      setUploading(false);
    }
  };

  // 删除图标
  const handleDeleteIcon = async (iconId: string) => {
    try {
      await iconService.deleteIcon(iconId);
      message.success('图标删除成功');
      loadCustomIcons();
    } catch (error: any) {
      if (error?.response?.status === 404) {
        message.warning('图标删除功能暂未开放，请联系管理员');
      } else {
        message.error('图标删除失败');
      }
    }
  };

  // 切换收藏
  const toggleFavorite = (iconName: string) => {
    const newFavorites = favorites.includes(iconName)
      ? favorites.filter(name => name !== iconName)
      : [...favorites, iconName];
    setFavorites(newFavorites);
    localStorage.setItem('icon-favorites', JSON.stringify(newFavorites));
  };

  // 渲染 Ant Design 图标
  const renderAntdIcon = (iconName: string) => {
    const IconComponent = (AntdIcons as any)[iconName];
    if (!IconComponent) return null;

    const isFavorite = favorites.includes(iconName);
    const isSelected = selectedIcon === iconName;

    return (
      <div
        key={iconName}
        className={`icon-item ${isSelected ? 'selected' : ''}`}
        onClick={() => onSelect(iconName)}
      >
        <div className="icon-content">
          <IconComponent className="icon" />
          <div className="icon-actions">
            <Tooltip title={isFavorite ? '取消收藏' : '收藏'}>
              <Button
                type="text"
                size="small"
                icon={isFavorite ? <StarFilled /> : <StarOutlined />}
                onClick={(e) => {
                  e.stopPropagation();
                  toggleFavorite(iconName);
                }}
                className={isFavorite ? 'favorite-active' : ''}
              />
            </Tooltip>
          </div>
        </div>
        <div className="icon-name">{iconName}</div>
      </div>
    );
  };

  // 渲染自定义图标
  const renderCustomIcon = (icon: IconLibraryItem) => {
    const isSelected = selectedIcon === icon.name;

    return (
      <div
        key={icon.id}
        className={`icon-item ${isSelected ? 'selected' : ''}`}
        onClick={() => onSelect(icon.name)}
      >
        <div className="icon-content">
          <img src={icon.url} alt={icon.name} className="icon custom-icon" />
          <div className="icon-actions">
            <Tooltip title="删除图标">
              <Button
                type="text"
                size="small"
                icon={<DeleteOutlined />}
                onClick={(e) => {
                  e.stopPropagation();
                  handleDeleteIcon(icon.id);
                }}
                danger
              />
            </Tooltip>
          </div>
        </div>
        <div className="icon-name">{icon.name}</div>
      </div>
    );
  };

  useEffect(() => {
    if (visible) {
      loadCustomIcons();
      // 加载收藏列表
      const savedFavorites = localStorage.getItem('icon-favorites');
      if (savedFavorites) {
        setFavorites(JSON.parse(savedFavorites));
      }
    }
  }, [visible]);

  useEffect(() => {
    setCurrentPage(1);
  }, [searchKeyword, activeTab]);

  return (
    <Modal
      title="选择图标"
      open={visible}
      onCancel={onCancel}
      footer={null}
      width={800}
      className="icon-selector-modal"
    >
      <div className="icon-selector">
        {/* 搜索和工具栏 */}
        <div className="icon-toolbar">
          <Search
            placeholder="搜索图标..."
            value={searchKeyword}
            onChange={(e) => setSearchKeyword(e.target.value)}
            style={{ width: 300 }}
            allowClear
          />
          <Space>
            <Select
              value={sortBy}
              onChange={setSortBy}
              style={{ width: 120 }}
            >
              <Option value="name">按名称</Option>
              <Option value="usage">按使用量</Option>
              <Option value="created">按创建时间</Option>
            </Select>
            <Button
              type={viewMode === 'grid' ? 'primary' : 'default'}
              icon={<AppstoreOutlined />}
              onClick={() => setViewMode('grid')}
            />
            <Button
              type={viewMode === 'list' ? 'primary' : 'default'}
              icon={<UnorderedListOutlined />}
              onClick={() => setViewMode('list')}
            />
          </Space>
        </div>

        {/* 标签页 */}
        <Tabs activeKey={activeTab} onChange={setActiveTab}>
          <TabPane tab="Ant Design 图标" key="antd">
            <Spin spinning={loading}>
              {filteredAntdIcons.length > 0 ? (
                <>
                  <div className={`icon-grid ${viewMode}`}>
                    {paginatedIcons.map((iconName) => renderAntdIcon(iconName as string))}
                  </div>
                  <div className="icon-pagination">
                    <Pagination
                      current={currentPage}
                      pageSize={pageSize}
                      total={filteredAntdIcons.length}
                      onChange={setCurrentPage}
                      onShowSizeChange={(current, size) => {
                        setCurrentPage(1);
                        setPageSize(size);
                      }}
                      showSizeChanger
                      showQuickJumper
                      showTotal={(total, range) =>
                        `第 ${range[0]}-${range[1]} 项，共 ${total} 项`
                      }
                    />
                  </div>
                </>
              ) : (
                <Empty description="未找到匹配的图标" />
              )}
            </Spin>
          </TabPane>

          <TabPane
            tab="自定义图标"
            key="custom"
            tabBarExtraContent={
              allowUpload && (
                <Upload
                  accept=".svg,.png,.jpg,.jpeg,.gif"
                  showUploadList={false}
                  beforeUpload={(file) => {
                    handleUpload(file);
                    return false;
                  }}
                  disabled={uploading}
                >
                  <Button
                    type="primary"
                    icon={<UploadOutlined />}
                    loading={uploading}
                    size="small"
                  >
                    上传图标
                  </Button>
                </Upload>
              )
            }
          >
            <Spin spinning={loading}>
              {filteredCustomIcons.length > 0 ? (
                <>
                  <div className={`icon-grid ${viewMode}`}>
                    {(paginatedIcons as IconLibraryItem[]).map((icon) => renderCustomIcon(icon))}
                  </div>
                  <div className="icon-pagination">
                    <Pagination
                      current={currentPage}
                      pageSize={pageSize}
                      total={filteredCustomIcons.length}
                      onChange={setCurrentPage}
                      onShowSizeChange={(current, size) => {
                        setCurrentPage(1);
                        setPageSize(size);
                      }}
                      showSizeChanger
                      showQuickJumper
                      showTotal={(total, range) =>
                        `第 ${range[0]}-${range[1]} 项，共 ${total} 项`
                      }
                    />
                  </div>
                </>
              ) : (
                <Empty description="暂无自定义图标" />
              )}
            </Spin>
          </TabPane>
        </Tabs>
      </div>
    </Modal>
  );
};

export default IconSelector;