import React, { useState, useCallback, useMemo } from 'react';
import { Tree, Input, Button, Dropdown, Space, Tag, Tooltip, Modal, message } from 'antd';
import {
  SearchOutlined,
  StarOutlined,
  StarFilled,
  EyeOutlined,
  EyeInvisibleOutlined,
  EditOutlined,
  DeleteOutlined,
  PlusOutlined,
  DragOutlined,
  MoreOutlined,
  FilterOutlined,
  ExpandAltOutlined,
  CompressOutlined,
} from '@ant-design/icons';
import { DndProvider, useDrag, useDrop } from 'react-dnd';
import { HTML5Backend } from 'react-dnd-html5-backend';
import type { TreeProps, TreeDataNode } from 'antd/es/tree';
import { Menu, DeviceType } from '@/services/menu';
import IconSelector from '../IconSelector';
import './index.less';

const { Search } = Input;

export interface MenuTreeNode extends TreeDataNode {
  key: string;
  title: React.ReactNode;
  children?: MenuTreeNode[];
  menu: Menu;
  level: number;
}

interface MenuTreeProps {
  menus: Menu[];
  loading?: boolean;
  selectedKeys?: string[];
  expandedKeys?: string[];
  onSelect?: (selectedKeys: string[], info: any) => void;
  onExpand?: (expandedKeys: string[]) => void;
  onEdit?: (menu: Menu) => void;
  onDelete?: (menu: Menu) => void;
  onCreate?: (parentId?: string) => void;
  onToggleVisibility?: (menu: Menu) => void;
  onToggleFavorite?: (menu: Menu) => void;
  onDragSort?: (dragKey: string, dropKey: string, position: 'before' | 'after' | 'inside') => void;
  showSearch?: boolean;
  showFilter?: boolean;
  showActions?: boolean;
  deviceFilter?: DeviceType;
  favoriteFilter?: boolean;
  className?: string;
}

const DragableTreeNode: React.FC<{
  node: MenuTreeNode;
  onDragSort?: (dragKey: string, dropKey: string, position: 'before' | 'after' | 'inside') => void;
  children: React.ReactNode;
}> = ({ node, onDragSort, children }) => {
  const [{ isDragging }, drag] = useDrag({
    type: 'menu-node',
    item: { key: node.key, menu: node.menu },
    collect: (monitor) => ({
      isDragging: monitor.isDragging(),
    }),
  });

  const [{ isOver }, drop] = useDrop({
    accept: 'menu-node',
    drop: (item: { key: string; menu: Menu }, monitor) => {
      if (!monitor.isOver({ shallow: true })) return;
      
      const dragKey = item.key;
      const dropKey = node.key;
      
      if (dragKey === dropKey) return;
      
      // 确定放置位置
      const position = dropPosition < 0.5 ? 'before' : dropPosition > 0.5 ? 'after' : 'inside';
      onDragSort?.(dragKey, dropKey, position);
    },
    hover: (item, monitor) => {
      const hoverBoundingRect = (drop.current as any)?.getBoundingClientRect();
      if (!hoverBoundingRect) return;
      
      const hoverMiddleY = (hoverBoundingRect.bottom - hoverBoundingRect.top) / 2;
      const clientOffset = monitor.getClientOffset();
      const hoverClientY = (clientOffset as any).y - hoverBoundingRect.top;
      
      // 设置放置位置
      setDropPosition(hoverClientY / hoverBoundingRect.height);
    },
    collect: (monitor) => ({
      isOver: monitor.isOver({ shallow: true }),
    }),
  });

  const [dropPosition, setDropPosition] = useState(0.5);

  const ref = (el: HTMLDivElement) => {
    drag(el);
    drop(el);
  };

  return (
    <div
      ref={ref}
      className={`menu-tree-node ${isDragging ? 'dragging' : ''} ${isOver ? 'drop-over' : ''}`}
      style={{ opacity: isDragging ? 0.5 : 1 }}
    >
      {children}
      {isOver && (
        <div
          className={`drop-indicator ${
            dropPosition < 0.3 ? 'before' : dropPosition > 0.7 ? 'after' : 'inside'
          }`}
        />
      )}
    </div>
  );
};

export const MenuTree: React.FC<MenuTreeProps> = ({
  menus,
  loading = false,
  selectedKeys = [],
  expandedKeys = [],
  onSelect,
  onExpand,
  onEdit,
  onDelete,
  onCreate,
  onToggleVisibility,
  onToggleFavorite,
  onDragSort,
  showSearch = true,
  showFilter = true,
  showActions = true,
  deviceFilter,
  favoriteFilter,
  className,
}) => {
  const [searchValue, setSearchValue] = useState('');
  const [autoExpandParent, setAutoExpandParent] = useState(true);
  const [expandAll, setExpandAll] = useState(false);
  const [filterVisible, setFilterVisible] = useState(false);

  // 构建树形数据
  const treeData = useMemo(() => {
    const buildTree = (items: Menu[], parentId?: string, level = 0): MenuTreeNode[] => {
      return items
        .filter(item => item.parentId === parentId)
        .filter(item => {
          // 设备类型过滤
          if (deviceFilter && item.deviceTypes) {
            const deviceTypes = item.deviceTypes.split(',');
            if (!deviceTypes.includes(deviceFilter)) return false;
          }
          
          // 收藏过滤
          if (favoriteFilter !== undefined && item.isFavorite !== favoriteFilter) {
            return false;
          }
          
          // 搜索过滤
          if (searchValue) {
            const searchLower = searchValue.toLowerCase();
            return (
              item.title.toLowerCase().includes(searchLower) ||
              item.path?.toLowerCase().includes(searchLower) ||
              item.keywords?.toLowerCase().includes(searchLower)
            );
          }
          
          return true;
        })
        .sort((a, b) => a.sortOrder - b.sortOrder)
        .map(item => {
          const children = buildTree(items, item.id, level + 1);
          
          return {
            key: item.id,
            title: renderNodeTitle(item, level),
            children: children.length > 0 ? children : undefined,
            menu: item,
            level,
            isLeaf: children.length === 0,
          };
        });
    };

    return buildTree(menus);
  }, [menus, searchValue, deviceFilter, favoriteFilter]);

  // 渲染节点标题
  function renderNodeTitle(menu: Menu, level: number) {
    const handleMenuAction = (action: string, e: React.MouseEvent) => {
      e.stopPropagation();
      
      switch (action) {
        case 'edit':
          onEdit?.(menu);
          break;
        case 'delete':
          Modal.confirm({
            title: '确认删除',
            content: `确定要删除菜单"${menu.title}"吗？`,
            onOk: () => onDelete?.(menu),
          });
          break;
        case 'toggle-visibility':
          onToggleVisibility?.(menu);
          break;
        case 'toggle-favorite':
          onToggleFavorite?.(menu);
          break;
        case 'add-child':
          onCreate?.(menu.id);
          break;
      }
    };

    const actionItems = [
      {
        key: 'edit',
        icon: <EditOutlined />,
        label: '编辑',
        onClick: (e: any) => handleMenuAction('edit', e.domEvent),
      },
      {
        key: 'add-child',
        icon: <PlusOutlined />,
        label: '添加子菜单',
        onClick: (e: any) => handleMenuAction('add-child', e.domEvent),
      },
      {
        key: 'toggle-visibility',
        icon: menu.isHidden ? <EyeOutlined /> : <EyeInvisibleOutlined />,
        label: menu.isHidden ? '显示' : '隐藏',
        onClick: (e: any) => handleMenuAction('toggle-visibility', e.domEvent),
      },
      {
        type: 'divider' as const,
      },
      {
        key: 'delete',
        icon: <DeleteOutlined />,
        label: '删除',
        danger: true,
        onClick: (e: any) => handleMenuAction('delete', e.domEvent),
      },
    ];

    return (
      <div className="menu-tree-node-title">
        <div className="menu-info">
          <DragOutlined className="drag-handle" />
          
          {menu.icon && (
            <span className="menu-icon">
              <IconSelector value={menu.icon} disabled />
            </span>
          )}
          
          <span className="menu-title">{menu.title}</span>
          
          <div className="menu-tags">
            {menu.isFavorite && (
              <Tooltip title="收藏">
                <StarFilled 
                  className="favorite-icon active" 
                  onClick={(e) => handleMenuAction('toggle-favorite', e)}
                />
              </Tooltip>
            )}
            
            {menu.isHidden && (
              <Tag color="red" size="small">隐藏</Tag>
            )}
            
            {menu.deviceTypes && (
              <Tooltip title={`适配设备: ${menu.deviceTypes}`}>
                <Tag color="blue" size="small">
                  {menu.deviceTypes.split(',').length}端
                </Tag>
              </Tooltip>
            )}
            
            <Tag color="default" size="small">L{level}</Tag>
          </div>
        </div>
        
        {showActions && (
          <div className="menu-actions">
            <Button
              type="text"
              size="small"
              icon={menu.isFavorite ? <StarFilled /> : <StarOutlined />}
              onClick={(e) => handleMenuAction('toggle-favorite', e)}
              className={menu.isFavorite ? 'favorite-active' : ''}
            />
            
            <Dropdown
              menu={{ items: actionItems }}
              trigger={['click']}
              placement="bottomRight"
            >
              <Button
                type="text"
                size="small"
                icon={<MoreOutlined />}
                onClick={(e) => e.stopPropagation()}
              />
            </Dropdown>
          </div>
        )}
      </div>
    );
  }

  // 搜索处理
  const handleSearch = (value: string) => {
    setSearchValue(value);
    if (value) {
      // 搜索时自动展开所有节点
      const allKeys = menus.map(menu => menu.id);
      onExpand?.(allKeys);
      setAutoExpandParent(true);
    }
  };

  // 展开/收起所有
  const handleExpandAll = () => {
    if (expandAll) {
      onExpand?.([]);
    } else {
      const allKeys = menus.map(menu => menu.id);
      onExpand?.(allKeys);
    }
    setExpandAll(!expandAll);
  };

  // 过滤器菜单
  const filterMenu = {
    items: [
      {
        key: 'device-pc',
        label: 'PC端',
        onClick: () => {/* 处理设备过滤 */},
      },
      {
        key: 'device-mobile',
        label: '移动端',
        onClick: () => {/* 处理设备过滤 */},
      },
      {
        type: 'divider' as const,
      },
      {
        key: 'favorite-only',
        label: '仅收藏',
        onClick: () => {/* 处理收藏过滤 */},
      },
      {
        key: 'hidden-only',
        label: '仅隐藏',
        onClick: () => {/* 处理隐藏过滤 */},
      },
    ],
  };

  return (
    <DndProvider backend={HTML5Backend}>
      <div className={`menu-tree-container ${className || ''}`}>
        {(showSearch || showFilter) && (
          <div className="menu-tree-toolbar">
            {showSearch && (
              <Search
                placeholder="搜索菜单名称、路径或关键词"
                allowClear
                onSearch={handleSearch}
                onChange={(e) => handleSearch(e.target.value)}
                style={{ flex: 1 }}
              />
            )}
            
            <Space>
              {showFilter && (
                <Dropdown menu={filterMenu} trigger={['click']}>
                  <Button icon={<FilterOutlined />}>过滤</Button>
                </Dropdown>
              )}
              
              <Button
                icon={expandAll ? <CompressOutlined /> : <ExpandAltOutlined />}
                onClick={handleExpandAll}
              >
                {expandAll ? '收起' : '展开'}全部
              </Button>
              
              {onCreate && (
                <Button
                  type="primary"
                  icon={<PlusOutlined />}
                  onClick={() => onCreate()}
                >
                  新增菜单
                </Button>
              )}
            </Space>
          </div>
        )}
        
        <div className="menu-tree-content">
          <Tree
            treeData={treeData}
            selectedKeys={selectedKeys}
            expandedKeys={expandedKeys}
            autoExpandParent={autoExpandParent}
            onSelect={onSelect}
            onExpand={(keys) => {
              onExpand?.(keys as string[]);
              setAutoExpandParent(false);
            }}
            loading={loading}
            showLine={{ showLeafIcon: false }}
            blockNode
            titleRender={(node) => (
              <DragableTreeNode
                node={node as MenuTreeNode}
                onDragSort={onDragSort}
              >
                {node.title}
              </DragableTreeNode>
            )}
          />
        </div>
      </div>
    </DndProvider>
  );
};

export default MenuTree;