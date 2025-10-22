import React, { useState, useEffect, useCallback, useMemo } from 'react';
import {
  Card,
  Table,
  Button,
  Space,
  Modal,
  Form,
  Input,
  Select,
  Switch,
  InputNumber,
  App,
  Popconfirm,
  Tree,
  Tabs,
  Row,
  Col,
  Statistic,
  Tag,
  Tooltip,
  Dropdown,
  MenuProps,
  Checkbox,
  Divider,
  Alert,
} from 'antd';
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  EyeOutlined,
  EyeInvisibleOutlined,
  MenuOutlined,
  AppstoreOutlined,
  MoreOutlined,
  ReloadOutlined,
  UpOutlined,
  DownOutlined,
  SettingOutlined,
  StarOutlined,
  StarFilled,
  DragOutlined,
  BulbOutlined,
  MobileOutlined,
  ExportOutlined,
  ImportOutlined,
  SearchOutlined,
} from '@ant-design/icons';
import * as AntdIcons from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import type { DataNode } from 'antd/es/tree';
import { 
  menuService, 
  Menu, 
  CreateMenuRequest, 
  UpdateMenuRequest, 
  MenuStats,
  DeviceType,
  menuConfigService,
  menuOperationService,
} from '@/services/menu';
import { permissionService } from '@/services/permission';
import MenuTree from '@/components/MenuTree';
import IconSelector from '@/components/IconSelector';
import DeviceConfig, { DeviceConfigData } from '@/components/DeviceConfig';
import VisualConfig, { VisualConfigData } from '@/components/VisualConfig';
import KeyboardShortcuts, { KeyboardShortcut, createShortcut } from '@/components/KeyboardShortcuts';

// 移除TabPane的解构，使用新的items属性
const { Option } = Select;

interface MenuFormData {
  title: string;
  path: string;
  icon?: string;
  component?: string;
  parentId?: string;
  sortOrder: number;
  isHidden: boolean;
  isFavorite: boolean;
  deviceTypes: DeviceType[];
  permissions: string[];
  customIcon?: string;
  description?: string;
  keywords?: string;
}

const MenuManagement: React.FC = () => {
  const { message } = App.useApp();
  const [menus, setMenus] = useState<Menu[]>([]);
  const [treeData, setTreeData] = useState<DataNode[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [editingMenu, setEditingMenu] = useState<Menu | null>(null);
  const [form] = Form.useForm<MenuFormData>();
  const [stats, setStats] = useState<MenuStats | null>(null);
  const [activeTab, setActiveTab] = useState('tree');
  const [expandedKeys, setExpandedKeys] = useState<React.Key[]>([]);
  
  // 新增状态
  const [iconSelectorVisible, setIconSelectorVisible] = useState(false);
  const [deviceConfigVisible, setDeviceConfigVisible] = useState(false);
  const [visualConfigVisible, setVisualConfigVisible] = useState(false);
  const [selectedMenus, setSelectedMenus] = useState<string[]>([]);
  const [searchKeyword, setSearchKeyword] = useState('');
  const [filterDeviceType, setFilterDeviceType] = useState<DeviceType | 'all'>('all');
  const [showHidden, setShowHidden] = useState(false);
  const [showFavoriteOnly, setShowFavoriteOnly] = useState(false);
  const [dragMode, setDragMode] = useState(false);
  const [shortcutsEnabled, setShortcutsEnabled] = useState(true);
  const [showShortcutHelp, setShowShortcutHelp] = useState(false);

  // 权限门控状态
  const [canCreateMenu, setCanCreateMenu] = useState(true);
  const [canEditMenu, setCanEditMenu] = useState(true);
  const [canDeleteMenu, setCanDeleteMenu] = useState(true);

  useEffect(() => {
    (async () => {
      try {
        const [createRes, editRes, deleteRes] = await Promise.all([
          permissionService.checkCurrentUserPermission({ module: 'system', resource: 'menu', action: 'create' }),
          permissionService.checkCurrentUserPermission({ module: 'system', resource: 'menu', action: 'edit' }),
          permissionService.checkCurrentUserPermission({ module: 'system', resource: 'menu', action: 'delete' }),
        ]);
        setCanCreateMenu(!!createRes?.hasPermission);
        setCanEditMenu(!!editRes?.hasPermission);
        setCanDeleteMenu(!!deleteRes?.hasPermission);
      } catch (e) {
        // 接口异常时不阻断操作，保持默认 true
      }
    })();
  }, []);
  // 加载菜单列表
  const loadMenus = useCallback(async () => {
    setLoading(true);
    try {
      const response = await menuService.getMenus({ tree: true });
      setMenus(response.items);
      setTreeData(convertToTreeData(response.items));
    } catch (error) {
      message.error('加载菜单列表失败');
    } finally {
      setLoading(false);
    }
  }, []);

  // 加载统计信息
  const loadStats = useCallback(async () => {
    try {
      const statsData = await menuService.getMenuStats();
      setStats(statsData);
    } catch (error) {
      message.error('加载统计信息失败');
    }
  }, []);

  // 将菜单数据转换为树形数据
  const convertToTreeData = (menuList: Menu[]): DataNode[] => {
    return menuList.map(menu => ({
      key: menu.id,
      title: (
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <span>
            {menu.icon && <span style={{ marginRight: 8 }}>{menu.icon}</span>}
            {menu.title}
            {menu.isHidden && <Tag color="orange" size="small" style={{ marginLeft: 8 }}>隐藏</Tag>}
          </span>
          <Space size="small">
            <Tooltip title="编辑">
              <Button
                type="text"
                size="small"
                icon={<EditOutlined />}
                onClick={(e) => {
                  e.stopPropagation();
                  handleEdit(menu);
                }}
                disabled={!canEditMenu}
              />
            </Tooltip>
            <Tooltip title="删除">
              <Popconfirm
                title="确定要删除这个菜单吗？"
                onConfirm={(e) => {
                  e?.stopPropagation();
                  handleDelete(menu.id);
                }}
                onCancel={(e) => e?.stopPropagation()}
              >
                <Button
                  type="text"
                  size="small"
                  danger
                  icon={<DeleteOutlined />}
                  onClick={(e) => e.stopPropagation()}
                  disabled={!canDeleteMenu}
                />
              </Popconfirm>
            </Tooltip>
          </Space>
        </div>
      ),
      children: menu.children ? convertToTreeData(menu.children) : undefined,
    }));
  };

  // 获取所有菜单的扁平列表（用于父菜单选择）
  const getFlatMenus = (menuList: Menu[], level = 0): Array<Menu & { level: number }> => {
    const result: Array<Menu & { level: number }> = [];
    menuList.forEach(menu => {
      result.push({ ...menu, level });
      if (menu.children) {
        result.push(...getFlatMenus(menu.children, level + 1));
      }
    });
    return result;
  };

  // 表格列定义
  const columns: ColumnsType<Menu> = [
    {
      title: '菜单名称',
      dataIndex: 'title',
      key: 'title',
      width: 200,
      render: (title: string, record) => (
        <Space>
          {record.isFavorite && <StarFilled style={{ color: '#faad14' }} />}
          {/* 移除图标名称文本，避免在名称列显示 */}
          <span>{title}</span>
          {record.isHidden && <Tag color="orange" size="small">隐藏</Tag>}
        </Space>
      ),
    },
    {
      title: '图标',
      dataIndex: 'icon',
      key: 'icon',
      width: 80,
      render: (icon: string, record) => {
        if (record.customIcon) {
          return <img src={record.customIcon} alt="custom icon" style={{ width: 16, height: 16 }} />;
        }
        if (!icon) return '-';
        const IconComp = (AntdIcons as any)[icon];
        return IconComp ? <IconComp style={{ fontSize: 16 }} /> : icon;
      },
    },
    {
      title: '路径',
      dataIndex: 'path',
      key: 'path',
      width: 200,
      render: (text) => text || '-',
    },
    {
      title: '设备类型',
      dataIndex: 'deviceTypes',
      key: 'deviceTypes',
      width: 120,
      render: (deviceTypes: any) => {
        let types: string[] = [];
        if (Array.isArray(deviceTypes)) {
          types = deviceTypes as string[];
        } else if (typeof deviceTypes === 'string' && deviceTypes) {
          try {
            const parsed = JSON.parse(deviceTypes);
            types = Array.isArray(parsed) ? parsed : String(deviceTypes).split(',').map(s => s.trim()).filter(Boolean);
          } catch {
            types = String(deviceTypes).split(',').map(s => s.trim()).filter(Boolean);
          }
        }
        const labelMap: Record<string, string> = {
          [DeviceType.PC]: '桌面',
          [DeviceType.WEB]: 'Web',
          [DeviceType.MOBILE]: '移动',
          [DeviceType.WATCH]: '手表',
          [DeviceType.IOT]: '物联网',
          [DeviceType.ROBOT]: '机器人',
        };
        return (
          <Space wrap>
            {types.length > 0 ? types.map(t => <Tag key={t} size="small">{labelMap[t] || t}</Tag>) : '-'}
          </Space>
        );
      },
    },
    {
      title: '组件',
      dataIndex: 'component',
      key: 'component',
      render: (text) => text || '-',
    },
    {
      title: '排序',
      dataIndex: 'sortOrder',
      key: 'sortOrder',
      width: 80,
      sorter: (a, b) => a.sortOrder - b.sortOrder,
    },
    {
      title: '状态',
      dataIndex: 'isHidden',
      key: 'isHidden',
      width: 80,
      render: (isHidden) => (
        <Tag color={isHidden ? 'orange' : 'green'}>
          {isHidden ? '隐藏' : '显示'}
        </Tag>
      ),
    },
    {
      title: '创建时间',
      dataIndex: 'createdAt',
      key: 'createdAt',
      width: 180,
      render: (text) => new Date(text).toLocaleString(),
    },
    {
      title: '操作',
      key: 'action',
      width: 250,
      render: (_, record) => (
        <Space size="small">
          <Tooltip title="编辑">
            <Button
              type="text"
              size="small"
              icon={<EditOutlined />}
              onClick={() => handleEdit(record)}
              disabled={!canEditMenu}
            />
          </Tooltip>
          <Tooltip title={record.isFavorite ? '取消收藏' : '收藏'}>
            <Button
              type="text"
              size="small"
              icon={record.isFavorite ? <StarFilled /> : <StarOutlined />}
              onClick={() => handleToggleFavorite(record.id)}
              style={{ color: record.isFavorite ? '#faad14' : undefined }}
            />
          </Tooltip>
          <Tooltip title={record.isHidden ? '显示' : '隐藏'}>
            <Button
              type="text"
              size="small"
              icon={record.isHidden ? <EyeOutlined /> : <EyeInvisibleOutlined />}
              onClick={() => handleToggleVisibility(record)}
              disabled={!canEditMenu}
            />
          </Tooltip>
          <Dropdown
            menu={{
              items: [
                {
                  key: 'moveUp',
                  label: '上移',
                  icon: <UpOutlined />,
                  onClick: () => handleMove(record, 'up'),
                },
                {
                  key: 'moveDown',
                  label: '下移',
                  icon: <DownOutlined />,
                  onClick: () => handleMove(record, 'down'),
                },
                {
                  type: 'divider',
                },
                {
                  key: 'delete',
                  label: '删除',
                  icon: <DeleteOutlined />,
                  danger: true,
                  onClick: () => handleDelete(record),
                  disabled: !canDeleteMenu,
                },
              ],
            }}
            trigger={['click']}
          >
            <Button type="text" size="small" icon={<MoreOutlined />} />
          </Dropdown>
        </Space>
      ),
    },
  ];

  // 处理新增菜单
  const handleAdd = () => {
    setEditingMenu(null);
    form.resetFields();
    form.setFieldsValue({
      sortOrder: 0,
      isHidden: false,
    });
    setModalVisible(true);
  };

  // 处理编辑菜单
  const handleEdit = (menu: Menu) => {
    setEditingMenu(menu);
    // 解析 deviceTypes/permissions 为数组以填充表单
    const parseToArray = (val: any): string[] => {
      if (!val) return [];
      if (Array.isArray(val)) return val as string[];
      if (typeof val === 'string') {
        try {
          const parsed = JSON.parse(val);
          return Array.isArray(parsed) ? parsed : String(val).split(',').map(s => s.trim()).filter(Boolean);
        } catch {
          return String(val).split(',').map(s => s.trim()).filter(Boolean);
        }
      }
      return [];
    };

    form.setFieldsValue({
      title: menu.title,
      path: menu.path,
      icon: menu.icon,
      component: menu.component,
      parentId: menu.parentId,
      sortOrder: menu.sortOrder,
      isHidden: menu.isHidden,
      isFavorite: menu.isFavorite,
      deviceTypes: parseToArray(menu.deviceTypes),
      permissions: parseToArray(menu.permissions),
      description: menu.description,
      keywords: menu.keywords,
    });
    setModalVisible(true);
  };

  // 处理删除菜单
  const handleDelete = async (menu: Menu) => {
    try {
      await menuService.deleteMenu(menu.id);
      message.success('删除成功');
      loadMenus();
      loadStats();
    } catch (error) {
      message.error('删除失败');
    }
  };

  // 处理切换显示/隐藏状态
  const handleToggleVisibility = async (menu: Menu) => {
    try {
      await menuService.updateMenu(menu.id, {
        isHidden: !menu.isHidden,
      });
      message.success('状态更新成功');
      loadMenus();
      loadStats();
    } catch (error) {
      message.error('状态更新失败');
    }
  };

  // 处理移动菜单
  const handleMove = async (menu: Menu, direction: 'up' | 'down') => {
    try {
      const newSortOrder = direction === 'up' ? menu.sortOrder - 1 : menu.sortOrder + 1;
      await menuService.updateMenu(menu.id, {
        sortOrder: newSortOrder,
      });
      message.success('移动成功');
      loadMenus();
    } catch (error) {
      message.error('移动失败');
    }
  };

  // 处理表单提交
  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();

      const payload: UpdateMenuRequest | CreateMenuRequest = {
        title: values.title,
        path: values.path,
        icon: values.icon,
        component: values.component,
        parentId: values.parentId,
        sortOrder: values.sortOrder,
        isHidden: values.isHidden,
        isFavorite: values.isFavorite,
        deviceTypes: Array.isArray(values.deviceTypes) && values.deviceTypes.length > 0
          ? JSON.stringify(values.deviceTypes)
          : undefined,
        permissions: Array.isArray(values.permissions) && values.permissions.length > 0
          ? JSON.stringify(values.permissions)
          : undefined,
        customIcon: values.customIcon,
        description: values.description,
        keywords: values.keywords,
      };

      if (editingMenu) {
        await menuService.updateMenu(editingMenu.id, payload as UpdateMenuRequest);
        message.success('更新成功');
      } else {
        await menuService.createMenu(payload as CreateMenuRequest);
        message.success('创建成功');
      }

      setModalVisible(false);
      loadMenus();
      loadStats();
    } catch (error) {
      message.error(editingMenu ? '更新失败' : '创建失败');
    }
  };

  // 处理拖拽排序
  const handleDragSort = async (draggedId: string, targetId: string, position: 'before' | 'after' | 'inside') => {
    try {
      await menuOperationService.dragSort({
        draggedMenuId: draggedId,
        targetMenuId: targetId,
        position,
      });
      message.success('菜单排序成功');
      await loadMenus();
    } catch (error) {
      message.error('菜单排序失败');
    }
  };

  // 处理收藏切换
  const handleToggleFavorite = async (menuId: string) => {
    try {
      await menuOperationService.toggleFavorite(menuId);
      message.success('收藏状态更新成功');
      await loadMenus();
    } catch (error) {
      message.error('收藏状态更新失败');
    }
  };

  // 处理批量操作
  const handleBatchOperation = async (operation: 'delete' | 'hide' | 'show' | 'favorite' | 'unfavorite') => {
    if (selectedMenus.length === 0) {
      message.warning('请先选择要操作的菜单');
      return;
    }

    try {
      setLoading(true);
      await menuOperationService.batchOperation({
        menuIds: selectedMenus,
        operation,
      });
      message.success('批量操作成功');
      setSelectedMenus([]);
      await loadMenus();
    } catch (error) {
      message.error('批量操作失败');
    } finally {
      setLoading(false);
    }
  };

  // 处理图标选择
  const handleIconSelect = (icon: string) => {
    form.setFieldsValue({ icon });
    setIconSelectorVisible(false);
  };

  // 处理设备配置保存
  const handleDeviceConfigSave = async (config: DeviceConfigData) => {
    try {
      await menuConfigService.updateConfig('device', config);
      message.success('设备配置保存成功');
      setDeviceConfigVisible(false);
    } catch (error) {
      message.error('设备配置保存失败');
    }
  };

  // 导出菜单配置
  const handleExportConfig = async () => {
    try {
      const config = await menuConfigService.exportConfig();
      const blob = new Blob([JSON.stringify(config, null, 2)], { type: 'application/json' });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `menu-config-${new Date().toISOString().split('T')[0]}.json`;
      a.click();
      URL.revokeObjectURL(url);
      message.success('配置导出成功');
    } catch (error) {
      message.error('配置导出失败');
    }
  };

  // 导入菜单配置
  const handleImportConfig = async (file: File) => {
    try {
      const text = await file.text();
      const config = JSON.parse(text);
      await menuConfigService.importConfig(config);
      message.success('配置导入成功');
      await loadMenus();
    } catch (error) {
      message.error('配置导入失败');
    }
  };

  // 处理可视化配置保存
  const handleVisualConfigSave = async (config: VisualConfigData) => {
    try {
      await menuConfigService.updateConfig('visual', config);
      message.success('可视化配置保存成功');
      setVisualConfigVisible(false);
    } catch (error) {
      message.error('可视化配置保存失败');
    }
  };

  // 定义快捷键配置
  const shortcuts: KeyboardShortcut[] = [
    createShortcut('n', () => handleAdd(), '新建菜单', { ctrl: true }),
    createShortcut('s', () => form.submit(), '保存表单', { ctrl: true }),
    createShortcut('f', () => document.querySelector<HTMLInputElement>('.ant-input')?.focus(), '聚焦搜索框', { ctrl: true }),
    createShortcut('r', () => loadMenus(), '刷新列表', { ctrl: true }),
    createShortcut('d', () => setDragMode(!dragMode), '切换拖拽模式', { ctrl: true }),
    createShortcut('t', () => setActiveTab(activeTab === 'tree' ? 'table' : 'tree'), '切换视图', { ctrl: true }),
    createShortcut('h', () => setShowHidden(!showHidden), '切换显示隐藏项', { ctrl: true }),
    createShortcut('b', () => setShowFavoriteOnly(!showFavoriteOnly), '切换仅显示收藏', { ctrl: true }),
    createShortcut('a', () => {
      const allMenuIds = filteredMenus.map(menu => menu.id);
      setSelectedMenus(selectedMenus.length === allMenuIds.length ? [] : allMenuIds);
    }, '全选/取消全选', { ctrl: true }),
    createShortcut('Delete', () => {
      if (selectedMenus.length > 0) {
        handleBatchOperation('delete');
      }
    }, '删除选中项'),
    createShortcut('Escape', () => {
      setModalVisible(false);
      setIconSelectorVisible(false);
      setDeviceConfigVisible(false);
      setVisualConfigVisible(false);
      setSelectedMenus([]);
    }, '取消/关闭'),
    createShortcut('F1', () => setShowShortcutHelp(true), '显示快捷键帮助'),
    createShortcut('i', () => setIconSelectorVisible(true), '打开图标选择器', { ctrl: true }),
    createShortcut('m', () => setDeviceConfigVisible(true), '打开设备配置', { ctrl: true }),
    createShortcut('v', () => setVisualConfigVisible(true), '打开可视化配置', { ctrl: true })
  ];

  // 获取扁平菜单列表用于表格显示
  const getFlatMenusForTable = (menuList: Menu[]): Menu[] => {
    const result: Menu[] = [];
    const traverse = (menus: Menu[], level = 0) => {
      menus.forEach(menu => {
        result.push({ ...menu, level } as Menu & { level: number });
        if (menu.children) {
          traverse(menu.children, level + 1);
        }
      });
    };
    traverse(menuList);
    return result;
  };

  // 过滤菜单数据
  const filteredMenus = useMemo(() => {
    let filtered = menus;

    // 搜索过滤
    if (searchKeyword) {
      filtered = filtered.filter(menu => 
        menu.title.toLowerCase().includes(searchKeyword.toLowerCase()) ||
        menu.path?.toLowerCase().includes(searchKeyword.toLowerCase()) ||
        menu.keywords?.toLowerCase().includes(searchKeyword.toLowerCase())
      );
    }

    // 设备类型过滤（当选择“全部”时不做过滤）
    if (filterDeviceType !== 'all') {
      filtered = filtered.filter(menu => {
        const types = Array.isArray(menu.deviceTypes)
          ? (menu.deviceTypes as unknown as string[])
          : (typeof menu.deviceTypes === 'string'
              ? (() => { try { return JSON.parse(menu.deviceTypes as string); } catch { return []; } })()
              : []);
        return types.includes(filterDeviceType as DeviceType);
      });
    }

    // 隐藏菜单过滤
    if (!showHidden) {
      filtered = filtered.filter(menu => !menu.isHidden);
    }

    // 收藏过滤
    if (showFavoriteOnly) {
      filtered = filtered.filter(menu => menu.isFavorite);
    }

    return filtered;
  }, [menus, searchKeyword, filterDeviceType, showHidden, showFavoriteOnly]);

  useEffect(() => {
    loadMenus();
    loadStats();
  }, [loadMenus, loadStats]);

  return (
    <div style={{ padding: '24px' }}>
      {/* 统计信息 */}
      {stats && (
        <Row gutter={16} style={{ marginBottom: 24 }}>
          <Col span={4}>
            <Card>
              <Statistic title="总菜单数" value={stats.totalMenus || 0} />
            </Card>
          </Col>
          <Col span={4}>
            <Card>
              <Statistic title="显示菜单" value={stats.visibleMenus || 0} />
            </Card>
          </Col>
          <Col span={4}>
            <Card>
              <Statistic title="隐藏菜单" value={stats.hiddenMenus || 0} />
            </Card>
          </Col>
          <Col span={4}>
            <Card>
              <Statistic title="收藏菜单" value={menus.filter(m => m.isFavorite).length} />
            </Card>
          </Col>
          <Col span={4}>
            <Card>
              <Statistic title="最大层级" value={stats.maxDepth || 0} />
            </Card>
          </Col>
          <Col span={4}>
            <Card>
              <Statistic title="已选择" value={selectedMenus.length} />
            </Card>
          </Col>
        </Row>
      )}

      <Card>
        {/* 工具栏 */}
        <div style={{ marginBottom: 16 }}>
          <Row gutter={[16, 16]}>
            <Col span={12}>
              <Space wrap>
                <Button
                  type="primary"
                  icon={<PlusOutlined />}
                  onClick={handleAdd}
                  disabled={!canCreateMenu}
                >
                  新增菜单
                </Button>
                <Button
                  icon={<ReloadOutlined />}
                  onClick={loadMenus}
                  loading={loading}
                >
                  刷新
                </Button>
                <Button
                  icon={<DragOutlined />}
                  type={dragMode ? 'primary' : 'default'}
                  onClick={() => setDragMode(!dragMode)}
                >
                  {dragMode ? '退出拖拽' : '拖拽排序'}
                </Button>
                <Button
                  icon={<SettingOutlined />}
                  onClick={() => setDeviceConfigVisible(true)}
                >
                  设备配置
                </Button>
                <Button
              icon={<BulbOutlined />}
              onClick={() => setVisualConfigVisible(true)}
            >
              可视化配置
            </Button>
            <Button
              icon={<SearchOutlined />}
              onClick={() => setShowShortcutHelp(true)}
              title="快捷键帮助 (F1)"
            >
              快捷键
            </Button>
                <Button
                  icon={<ExportOutlined />}
                  onClick={handleExportConfig}
                >
                  导出配置
                </Button>
                <Button
                  icon={<ImportOutlined />}
                  onClick={() => {
                    const input = document.createElement('input');
                    input.type = 'file';
                    input.accept = '.json';
                    input.onchange = (e) => {
                      const file = (e.target as HTMLInputElement).files?.[0];
                      if (file) handleImportConfig(file);
                    };
                    input.click();
                  }}
                >
                  导入配置
                </Button>
              </Space>
            </Col>
            <Col span={12}>
              <Space style={{ float: 'right' }} wrap>
                <Input
                  placeholder="搜索菜单..."
                  prefix={<SearchOutlined />}
                  value={searchKeyword}
                  onChange={(e) => setSearchKeyword(e.target.value)}
                  style={{ width: 200 }}
                  allowClear
                />
                <Select
                  placeholder="设备类型"
                  value={filterDeviceType}
                  onChange={setFilterDeviceType}
                  style={{ width: 120 }}
                  allowClear
                >
                  <Select.Option value={DeviceType.PC}>桌面</Select.Option>
                  <Select.Option value={DeviceType.WEB}>Web</Select.Option>
                  <Select.Option value={DeviceType.MOBILE}>移动</Select.Option>
                  <Select.Option value={DeviceType.WATCH}>手表</Select.Option>
                  <Select.Option value={DeviceType.IOT}>物联网</Select.Option>
                  <Select.Option value={DeviceType.ROBOT}>机器人</Select.Option>
                </Select>
                <Checkbox
                  checked={showHidden}
                  onChange={(e) => setShowHidden(e.target.checked)}
                >
                  显示隐藏
                </Checkbox>
                <Checkbox
                  checked={showFavoriteOnly}
                  onChange={(e) => setShowFavoriteOnly(e.target.checked)}
                >
                  仅收藏
                </Checkbox>
              </Space>
            </Col>
          </Row>
        </div>

        {/* 批量操作栏 */}
        {selectedMenus.length > 0 && (
          <Alert
            message={
              <Space>
                <span>已选择 {selectedMenus.length} 个菜单</span>
                <Button size="small" onClick={() => setSelectedMenus([])}>
                  清空选择
                </Button>
                <Divider type="vertical" />
                <Button 
                  size="small" 
                  onClick={() => handleBatchOperation('hide')}
                >
                  批量隐藏
                </Button>
                <Button 
                  size="small" 
                  onClick={() => handleBatchOperation('show')}
                >
                  批量显示
                </Button>
                <Button 
                  size="small" 
                  onClick={() => handleBatchOperation('favorite')}
                >
                  批量收藏
                </Button>
                <Button 
                  size="small" 
                  onClick={() => handleBatchOperation('unfavorite')}
                >
                  取消收藏
                </Button>
                <Button 
                  size="small" 
                  danger
                  onClick={() => handleBatchOperation('delete')}
                >
                  批量删除
                </Button>
              </Space>
            }
            type="info"
            style={{ marginBottom: 16 }}
          />
        )}

        <Tabs
          activeKey={activeTab}
          onChange={setActiveTab}
          items={[
            {
              key: 'tree',
              label: (
                <span>
                  <MenuOutlined />
                  树形视图
                </span>
              ),
              children: (
                <MenuTree
                  menus={filteredMenus}
                  onEdit={handleEdit}
                  onDelete={handleDelete}
                  onToggleVisibility={handleToggleVisibility}
                  onToggleFavorite={handleToggleFavorite}
                  onDragSort={handleDragSort}
                  selectedKeys={selectedMenus}
                  onSelect={(selectedKeys) => setSelectedMenus(selectedKeys)}
                  loading={loading}
                />
              ),
            },
            {
              key: 'table',
              label: (
                <span>
                  <AppstoreOutlined />
                  表格视图
                </span>
              ),
              children: (
                <Table
                  columns={columns}
                  dataSource={getFlatMenusForTable(filteredMenus)}
                  rowKey="id"
                  loading={loading}
                  rowSelection={{
                    selectedRowKeys: selectedMenus,
                    onChange: setSelectedMenus,
                  }}
                  pagination={{
                    showSizeChanger: true,
                    showQuickJumper: true,
                    showTotal: (total) => `共 ${total} 条记录`,
                  }}
                />
              ),
            },
          ]}
        />
      </Card>

      {/* 新增/编辑菜单模态框 */}
      <Modal
        title={editingMenu ? '编辑菜单' : '新增菜单'}
        open={modalVisible}
        onOk={handleSubmit}
        onCancel={() => setModalVisible(false)}
        width={800}
      >
        <Form form={form} layout="vertical">
          <Form.Item
            name="title"
            label="菜单名称"
            rules={[{ required: true, message: '请输入菜单名称' }]}
          >
            <Input placeholder="请输入菜单名称" />
          </Form.Item>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item name="path" label="路径">
                <Input placeholder="请输入路径" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="icon" label="图标">
                <Input
                  placeholder="请输入图标"
                  suffix={
                    <Button
                      type="text"
                      size="small"
                      icon={<BulbOutlined />}
                      onClick={() => setIconSelectorVisible(true)}
                    />
                  }
                />
              </Form.Item>
            </Col>
          </Row>

          <Form.Item name="component" label="组件">
            <Input placeholder="请输入组件路径" />
          </Form.Item>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item name="parentId" label="父菜单">
                <Select placeholder="请选择父菜单" allowClear>
                  {getFlatMenus(menus).map(menu => (
                    <Option key={menu.id} value={menu.id}>
                      {'　'.repeat(menu.level)}{menu.title}
                    </Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="sortOrder" label="排序">
                <InputNumber min={0} style={{ width: '100%' }} />
              </Form.Item>
            </Col>
          </Row>

          <Form.Item name="deviceTypes" label="设备类型">
            <Checkbox.Group>
              <Row>
                <Col span={8}>
                  <Checkbox value={DeviceType.PC}>桌面</Checkbox>
                </Col>
                <Col span={8}>
                  <Checkbox value={DeviceType.WEB}>Web</Checkbox>
                </Col>
                <Col span={8}>
                  <Checkbox value={DeviceType.MOBILE}>移动</Checkbox>
                </Col>
                <Col span={8}>
                  <Checkbox value={DeviceType.WATCH}>手表</Checkbox>
                </Col>
                <Col span={8}>
                  <Checkbox value={DeviceType.IOT}>物联网</Checkbox>
                </Col>
                <Col span={8}>
                  <Checkbox value={DeviceType.ROBOT}>机器人</Checkbox>
                </Col>
              </Row>
            </Checkbox.Group>
          </Form.Item>

          <Form.Item name="permissions" label="权限">
            <Select mode="tags" placeholder="请输入权限标识" />
          </Form.Item>

          <Form.Item name="description" label="描述">
            <Input.TextArea placeholder="请输入菜单描述" rows={3} />
          </Form.Item>

          <Form.Item name="keywords" label="关键词">
            <Input placeholder="请输入关键词，用于搜索" />
          </Form.Item>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item name="isHidden" label="是否隐藏" valuePropName="checked">
                <Switch />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="isFavorite" label="是否收藏" valuePropName="checked">
                <Switch />
              </Form.Item>
            </Col>
          </Row>
        </Form>
      </Modal>

      {/* 图标选择器 */}
      <IconSelector
        visible={iconSelectorVisible}
        onSelect={handleIconSelect}
        onCancel={() => setIconSelectorVisible(false)}
      />

      {/* 设备配置 */}
      <DeviceConfig
        visible={deviceConfigVisible}
        onSave={handleDeviceConfigSave}
        onCancel={() => setDeviceConfigVisible(false)}
      />

      {/* 可视化配置 */}
      <VisualConfig
        visible={visualConfigVisible}
        onSave={handleVisualConfigSave}
        onCancel={() => setVisualConfigVisible(false)}
      />

      {/* 快捷键支持 */}
      <KeyboardShortcuts
        shortcuts={shortcuts}
        enabled={shortcutsEnabled}
        showHelp={showShortcutHelp}
        onShowHelp={() => setShowShortcutHelp(false)}
      />

      <style>{`
        .menu-level-0 {
          font-weight: bold;
        }
        .menu-level-1 td:first-child {
          padding-left: 32px !important;
        }
        .menu-level-2 td:first-child {
          padding-left: 48px !important;
        }
        .menu-level-3 td:first-child {
          padding-left: 64px !important;
        }
      `}</style>
    </div>
  );
};

export default MenuManagement;