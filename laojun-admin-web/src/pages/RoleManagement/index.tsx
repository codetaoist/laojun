import { useState, useEffect, useMemo } from 'react';
import { 
  Table, Card, Button, Space, Input, Tag, message, Modal, Form, 
  Select, Popconfirm, Drawer, Tree, Checkbox, Row, Col, Divider,
  Descriptions, Statistic, Badge, Timeline, Tabs, Dropdown
} from 'antd';
import type { ColumnsType } from 'antd/es/table';
import type { QueryParams } from '@/types';
import { 
  PlusOutlined, SearchOutlined, EditOutlined, DeleteOutlined, 
  ReloadOutlined, SafetyOutlined, TeamOutlined, EyeOutlined,
  UserOutlined, HistoryOutlined, InfoCircleOutlined, ExclamationCircleOutlined,
  MoreOutlined
} from '@ant-design/icons';
import { roleService } from '@/services/role';
import { Role, Permission } from '@/types';

const { Search } = Input;
const { Option } = Select;
const { TabPane } = Tabs;

// 角色活动记录接口
interface RoleActivity {
  id: string;
  action: string;
  description: string;
  timestamp: string;
  operator: string;
  details?: any;
}

// 权限模板接口
interface PermissionTemplate {
  id: string;
  name: string;
  description: string;
  permissionIds: string[];
}

// 角色统计信息接口
interface RoleStats {
  userCount: number;
  permissionCount: number;
  lastUsed: string;
  usageFrequency: number;
  riskLevel: 'low' | 'medium' | 'high';
}

// 权限分析接口
interface PermissionAnalysis {
  byResource: { [key: string]: number };
  byAction: { [key: string]: number };
  systemPermissions: number;
  customPermissions: number;
  riskScore: number;
}

// 角色依赖信息接口
interface RoleDependency {
  userCount: number;
  users: Array<{ id: string; name: string; email: string }>;
  hasSystemPermissions: boolean;
  systemPermissions: string[];
  relatedRoles: Array<{ id: string; name: string; relation: string }>;
  lastUsed: string;
}

const RoleManagement: React.FC = () => {
  const [roles, setRoles] = useState<Role[]>([]);
  const [loading, setLoading] = useState(false);
  const [searchText, setSearchText] = useState('');
  const [selectedRowKeys, setSelectedRowKeys] = useState<string[]>([]);
  const [modalVisible, setModalVisible] = useState(false);
  const [permissionDrawerVisible, setPermissionDrawerVisible] = useState(false);
  const [editingRole, setEditingRole] = useState<Role | null>(null);
  const [currentRole, setCurrentRole] = useState<Role | null>(null);
  const [permissions, setPermissions] = useState<Permission[]>([]);
  const [selectedPermissions, setSelectedPermissions] = useState<string[]>([]);
  const [permissionSearch, setPermissionSearch] = useState('');
  const [form] = Form.useForm();
  
  // 分页状态
  const [page, setPage] = useState<number>(1);
  const [pageSize, setPageSize] = useState<number>(10);
  const [total, setTotal] = useState<number>(0);
  
  // 角色详情相关状态
  const [detailDrawerVisible, setDetailDrawerVisible] = useState(false);
  const [detailRole, setDetailRole] = useState<Role | null>(null);
  const [roleActivities, setRoleActivities] = useState<RoleActivity[]>([]);
  const [activitiesLoading, setActivitiesLoading] = useState(false);
  const [roleStats, setRoleStats] = useState<RoleStats | null>(null);
  const [permissionAnalysis, setPermissionAnalysis] = useState<PermissionAnalysis | null>(null);
  const [statsLoading, setStatsLoading] = useState(false);
  
  // 权限模板相关状态
  const [permissionTemplates, setPermissionTemplates] = useState<PermissionTemplate[]>([]);
  const [selectedTemplate, setSelectedTemplate] = useState<string>('');

  // 权限管理相关状态
  const [permissionRole, setPermissionRole] = useState<Role | null>(null);
  const [allPermissions, setAllPermissions] = useState<Permission[]>([]);
  const [permissionsLoading, setPermissionsLoading] = useState(false);
  const [permissionSearchText, setPermissionSearchText] = useState('');
  const [permissionFilter, setPermissionFilter] = useState<'all' | 'assigned' | 'unassigned'>('all');
  const [permissionGroupBy, setPermissionGroupBy] = useState<'none' | 'resource' | 'action'>('resource');

  // 模拟权限模板数据
  const mockPermissionTemplates: PermissionTemplate[] = [
    {
      id: '1',
      name: '基础用户',
      description: '只能查看个人信息的基础权限',
      permissionIds: ['3']
    },
    {
      id: '2',
      name: '管理员',
      description: '拥有用户和角色管理权限',
      permissionIds: ['1', '2', '3']
    },
    {
      id: '3',
      name: '系统管理员',
      description: '拥有所有系统权限',
      permissionIds: ['1', '2', '3', '4', '5']
    },
    {
      id: '4',
      name: '插件管理员',
      description: '专门负责插件管理的权限',
      permissionIds: ['3', '4']
    }
  ];

  // 加载角色数据
  const loadRoles = async (query?: Partial<QueryParams>) => {
    setLoading(true);
    try {
      const res = await roleService.getRoles({
        page: query?.page ?? page,
        pageSize: query?.pageSize ?? pageSize,
        search: query?.search ?? searchText,
      });
      setRoles(res.items);
      setPage(res.page);
      setPageSize(res.pageSize);
      setTotal(res.total);
    } catch (err) {
      message.error('加载角色列表失败');
    } finally {
      setLoading(false);
    }
  };
  
  // 加载权限模板
  const loadPermissionTemplates = async () => {
    try {
      await new Promise(resolve => setTimeout(resolve, 300));
      setPermissionTemplates(mockPermissionTemplates);
    } catch (err) {
      message.error('加载权限模板失败');
    }
  };

  // 初始化加载
  useEffect(() => {
    loadRoles({ page: 1, pageSize });
    loadPermissionTemplates();
  }, []);
  
  // 搜索处理
  const handleSearch = (value: string) => {
    setSearchText(value);
    setPage(1);
    loadRoles({ page: 1, pageSize, search: value });
  };
  
  // 表格列定义
  const columns: ColumnsType<Role> = [
    {
      title: '角色名称',
      dataIndex: 'displayName',
      key: 'displayName',
      render: (text: string, record: Role) => (
        <span>
          {text || record.name}
          {record.name && (
            <Tag style={{ marginLeft: 8 }} color="blue">{record.name}</Tag>
          )}
          {record.isSystem && (
            <Tag style={{ marginLeft: 8 }} color="red">系统</Tag>
          )}
        </span>
      ),
    },
    {
      title: '描述',
      dataIndex: 'description',
      key: 'description',
      ellipsis: true,
    },
    {
      title: '权限数量',
      key: 'permissionCount',
      render: (_, record: Role) => (
        <Tag color="green">{record.permissions?.length || 0} 个权限</Tag>
      ),
    },
    {
      title: '创建时间',
      dataIndex: 'createdAt',
      key: 'createdAt',
      render: (date: string) => new Date(date).toLocaleString(),
      sorter: true,
    },
    {
      title: '操作',
      key: 'action',
      width: 120,
      fixed: 'right',
      render: (_, record: Role) => {
        const menuItems = [
          {
            key: 'view',
            icon: <EyeOutlined />,
            label: '查看详情',
            onClick: () => handleViewDetail(record),
          },
          {
            key: 'permission',
            icon: <SafetyOutlined />,
            label: '权限管理',
            onClick: () => handlePermissionManagement(record),
          },
          {
            key: 'edit',
            icon: <EditOutlined />,
            label: '编辑角色',
            disabled: record.isSystem,
            onClick: () => handleEdit(record),
          },
          {
            type: 'divider',
          },
          {
            key: 'delete',
            icon: <DeleteOutlined />,
            label: '删除角色',
            danger: true,
            disabled: record.isSystem,
            onClick: () => {
              Modal.confirm({
                title: '确定要删除这个角色吗？',
                content: `删除后将无法恢复角色"${record.displayName || record.name}"`,
                okText: '确定删除',
                okType: 'danger',
                cancelText: '取消',
                onOk: () => handleDelete(record.id),
              });
            },
          },
        ];

        return (
          <Space size="small">
            <Button
              type="link"
              size="small"
              icon={<EyeOutlined />}
              onClick={() => handleViewDetail(record)}
            >
              详情
            </Button>
            <Dropdown
              menu={{ items: menuItems }}
              trigger={['click']}
              placement="bottomRight"
            >
              <Button
                type="link"
                size="small"
                icon={<MoreOutlined />}
                onClick={(e) => e.preventDefault()}
              />
            </Dropdown>
          </Space>
        );
      },
    },
  ];

  // 处理编辑
  const handleEdit = (role: Role) => {
    setEditingRole(role);
    form.setFieldsValue({
      name: role.name,
      displayName: role.displayName || role.name,
      description: role.description,
    });
    setModalVisible(true);
  };

  // 检查角色依赖
  const checkRoleDependencies = async (roleId: string): Promise<RoleDependency> => {
    await new Promise(resolve => setTimeout(resolve, 500));
    
    const role = roles.find(r => r.id === roleId);
    if (!role) throw new Error('角色不存在');

    const mockDependency: RoleDependency = {
      userCount: Math.floor(Math.random() * 20),
      users: Array.from({ length: Math.min(5, Math.floor(Math.random() * 20)) }, (_, i) => ({
        id: `user_${i + 1}`,
        name: `用户${i + 1}`,
        email: `user${i + 1}@example.com`
      })),
      hasSystemPermissions: role.permissions?.some(p => p.isSystem) || false,
      systemPermissions: role.permissions?.filter(p => p.isSystem).map(p => p.name) || [],
      relatedRoles: [],
      lastUsed: new Date(Date.now() - Math.random() * 30 * 24 * 60 * 60 * 1000).toISOString()
    };

    return mockDependency;
  };

  // 处理删除
  const handleDelete = async (id: string) => {
    const role = roles.find(r => r.id === id);
    if (!role) {
      message.error('角色不存在');
      return;
    }

    if (role.isSystem) {
      message.error('系统角色不能删除');
      return;
    }

    try {
      const dependencies = await checkRoleDependencies(id);
      
      let confirmContent = (
        <div>
          <p><strong>确定要删除角色 "{role.displayName || role.name}" 吗？</strong></p>
          <div style={{ marginTop: '16px' }}>
            <p><strong>删除影响分析：</strong></p>
            <ul style={{ paddingLeft: '20px', margin: '8px 0' }}>
              <li>关联用户数量: <span style={{ color: dependencies.userCount > 0 ? '#f5222d' : '#52c41a' }}>
                {dependencies.userCount} 个用户
              </span></li>
              {dependencies.userCount > 0 && (
                <li style={{ color: '#f5222d' }}>
                  这些用户将失去此角色的所有权限
                </li>
              )}
              <li>系统权限: <span style={{ color: dependencies.hasSystemPermissions ? '#f5222d' : '#52c41a' }}>
                {dependencies.hasSystemPermissions ? `包含 ${dependencies.systemPermissions.length} 个系统权限` : '无系统权限'}
              </span></li>
              <li>最后使用时间: {new Date(dependencies.lastUsed).toLocaleString()}</li>
            </ul>
            
            {dependencies.userCount > 0 && (
              <div style={{ marginTop: '12px' }}>
                <p><strong>受影响的用户（前5个）：</strong></p>
                <div style={{ maxHeight: '100px', overflowY: 'auto', border: '1px solid #d9d9d9', padding: '8px', borderRadius: '4px' }}>
                  {dependencies.users.map(user => (
                    <div key={user.id} style={{ fontSize: '12px', marginBottom: '4px' }}>
                      {user.name} ({user.email})
                    </div>
                  ))}
                  {dependencies.userCount > 5 && (
                    <div style={{ fontSize: '12px', color: '#999' }}>
                      还有 {dependencies.userCount - 5} 个用户...
                    </div>
                  )}
                </div>
              </div>
            )}

            {dependencies.hasSystemPermissions && (
              <div style={{ marginTop: '12px' }}>
                <p><strong>包含的系统权限：</strong></p>
                <div style={{ maxHeight: '80px', overflowY: 'auto' }}>
                  {dependencies.systemPermissions.map(permission => (
                    <Tag key={permission} color="red" size="small" style={{ margin: '2px' }}>
                      {permission}
                    </Tag>
                  ))}
                </div>
              </div>
            )}
          </div>
        </div>
      );

      Modal.confirm({
        title: '删除角色确认',
        content: confirmContent,
        width: 600,
        okText: '确认删除',
        okType: 'danger',
        cancelText: '取消',
        onOk: async () => {
          try {
            await roleService.deleteRole(id);
            message.success('删除成功');
            loadRoles({ page, pageSize, search: searchText });
          } catch (error) {
            message.error('删除失败');
          }
        },
        onCancel: () => {
          message.info('已取消删除');
        }
      });

    } catch (error) {
      message.error('检查角色依赖失败');
    }
  };

  // 批量检查角色依赖
  const checkBatchRoleDependencies = async (roleIds: string[]): Promise<{ [key: string]: RoleDependency }> => {
    const dependencies: { [key: string]: RoleDependency } = {};
    
    for (const roleId of roleIds) {
      try {
        dependencies[roleId] = await checkRoleDependencies(roleId);
      } catch (error) {
        console.error(`检查角色 ${roleId} 依赖失败:`, error);
      }
    }
    
    return dependencies;
  };

  // 处理批量删除
  const handleBatchDelete = async () => {
    if (selectedRowKeys.length === 0) {
      message.warning('请选择要删除的角色');
      return;
    }
    
    const selectedRecords = roles.filter(r => selectedRowKeys.includes(r.id));
    const deletable = selectedRecords.filter(r => !r.isSystem);
    const systemRoles = selectedRecords.filter(r => r.isSystem);
    
    if (deletable.length === 0) {
      message.warning('没有可删除的角色（所选角色均为系统角色）');
      return;
    }

    try {
      const dependencies = await checkBatchRoleDependencies(deletable.map(r => r.id));
      
      const totalUserCount = Object.values(dependencies).reduce((sum, dep) => sum + dep.userCount, 0);
      const hasSystemPermissions = Object.values(dependencies).some(dep => dep.hasSystemPermissions);
      const systemPermissionCount = Object.values(dependencies).reduce((sum, dep) => sum + dep.systemPermissions.length, 0);
      
      let confirmContent = (
        <div>
          <p><strong>确定要批量删除 {deletable.length} 个角色吗？</strong></p>
          
          {systemRoles.length > 0 && (
            <div style={{ marginBottom: '16px', padding: '8px', backgroundColor: '#fff7e6', border: '1px solid #ffd591', borderRadius: '4px' }}>
              <p style={{ margin: 0, color: '#d46b08' }}>
                <ExclamationCircleOutlined style={{ marginRight: '4px' }} />
                将自动跳过 {systemRoles.length} 个系统角色：
              </p>
              <div style={{ marginTop: '4px', fontSize: '12px' }}>
                {systemRoles.map(role => (
                  <Tag key={role.id} color="orange" size="small" style={{ margin: '2px' }}>
                    {role.displayName || role.name}
                  </Tag>
                ))}
              </div>
            </div>
          )}

          <div style={{ marginTop: '16px' }}>
            <p><strong>批量删除影响分析：</strong></p>
            <ul style={{ paddingLeft: '20px', margin: '8px 0' }}>
              <li>总计影响用户: <span style={{ color: totalUserCount > 0 ? '#f5222d' : '#52c41a' }}>
                {totalUserCount} 个用户
              </span></li>
              <li>包含系统权限: <span style={{ color: hasSystemPermissions ? '#f5222d' : '#52c41a' }}>
                {hasSystemPermissions ? `${systemPermissionCount} 个系统权限` : '无系统权限'}
              </span></li>
              <li>删除角色数量: {deletable.length} 个</li>
            </ul>
            
            <div style={{ marginTop: '12px' }}>
              <p><strong>待删除角色详情：</strong></p>
              <div style={{ maxHeight: '200px', overflowY: 'auto', border: '1px solid #d9d9d9', padding: '8px', borderRadius: '4px' }}>
                {deletable.map(role => {
                  const dep = dependencies[role.id];
                  return (
                    <div key={role.id} style={{ marginBottom: '8px', padding: '8px', backgroundColor: '#fafafa', borderRadius: '4px' }}>
                      <div style={{ fontWeight: 'bold', marginBottom: '4px' }}>
                        {role.displayName || role.name}
                      </div>
                      <div style={{ fontSize: '12px', color: '#666' }}>
                        关联用户: {dep?.userCount || 0} 个 | 
                        系统权限: {dep?.systemPermissions?.length || 0} 个 |
                        最后使用: {dep ? new Date(dep.lastUsed).toLocaleDateString() : '未知'}
                      </div>
                      {dep && dep.userCount > 0 && (
                        <div style={{ fontSize: '11px', color: '#f5222d', marginTop: '2px' }}>
                          ⚠️ 将影响 {dep.userCount} 个用户的权限
                        </div>
                      )}
                    </div>
                  );
                })}
              </div>
            </div>

            {totalUserCount > 0 && (
              <div style={{ marginTop: '12px', padding: '8px', backgroundColor: '#fff2f0', border: '1px solid #ffccc7', borderRadius: '4px' }}>
                <p style={{ margin: 0, color: '#cf1322', fontSize: '12px' }}>
                  <ExclamationCircleOutlined style={{ marginRight: '4px' }} />
                  警告：此操作将影响 {totalUserCount} 个用户，这些用户将失去相关角色的所有权限！
                </p>
              </div>
            )}
          </div>
        </div>
      );

      Modal.confirm({
        title: '批量删除角色确认',
        content: confirmContent,
        width: 700,
        okText: '确认删除',
        okType: 'danger',
        cancelText: '取消',
        onOk: async () => {
          try {
            setLoading(true);
            let successCount = 0;
            let failCount = 0;
            
            for (const role of deletable) {
              try {
                await roleService.deleteRole(role.id);
                successCount++;
              } catch (error) {
                failCount++;
                console.error(`删除角色 ${role.name} 失败:`, error);
              }
            }
            
            if (successCount > 0) {
              message.success(`成功删除 ${successCount} 个角色${failCount > 0 ? `，${failCount} 个失败` : ''}`);
            }
            if (failCount > 0 && successCount === 0) {
              message.error(`批量删除失败，${failCount} 个角色删除失败`);
            }
            
            setSelectedRowKeys([]);
            loadRoles({ page, pageSize, search: searchText });
          } catch (error) {
            message.error('批量删除失败');
          } finally {
            setLoading(false);
          }
        },
        onCancel: () => {
          message.info('已取消批量删除');
        }
      });

    } catch (error) {
      message.error('检查角色依赖失败');
    }
  };

  // 处理表单提交
  const handleSubmit = async (values: any) => {
    try {
      if (editingRole) {
        await roleService.updateRole(editingRole.id, {
          displayName: values.displayName || values.name,
          description: values.description || '',
        });
        message.success('更新成功');
      } else {
        await roleService.createRole({
          name: values.name,
          displayName: values.displayName || values.name,
          description: values.description || '',
        });
        message.success('创建成功');
      }
      setModalVisible(false);
      loadRoles({ page, pageSize, search: searchText });
    } catch (error) {
      message.error(editingRole ? '更新失败' : '创建失败');
    }
  };

  // 处理详情查看
  const handleViewDetail = async (role: Role) => {
    setDetailRole(role);
    setDetailDrawerVisible(true);
    loadRoleActivities(role.id);
    loadRoleStats(role.id);
    loadPermissionAnalysis(role);
  };

  // 加载角色活动日志
  const loadRoleActivities = async (roleId: string) => {
    setActivitiesLoading(true);
    try {
      await new Promise(resolve => setTimeout(resolve, 1000));
      
      const mockActivities: RoleActivity[] = [
        {
          id: '1',
          action: 'permission_assigned',
          description: '分配了用户管理权限',
          timestamp: '2024-01-15T10:30:00Z',
          operator: 'admin'
        },
        {
          id: '2',
          action: 'role_updated',
          description: '更新了角色描述',
          timestamp: '2024-01-10T15:20:00Z',
          operator: 'super_admin'
        },
        {
          id: '3',
          action: 'role_created',
          description: '创建了角色',
          timestamp: '2024-01-01T00:00:00Z',
          operator: 'system'
        }
      ];
      
      setRoleActivities(mockActivities);
    } catch (err) {
      message.error('加载活动日志失败');
    } finally {
      setActivitiesLoading(false);
    }
  };

  // 加载角色统计信息
  const loadRoleStats = async (roleId: string) => {
    setStatsLoading(true);
    try {
      await new Promise(resolve => setTimeout(resolve, 800));
      
      const mockStats: RoleStats = {
        userCount: Math.floor(Math.random() * 50) + 1,
        permissionCount: detailRole?.permissions?.length || 0,
        lastUsed: new Date(Date.now() - Math.random() * 7 * 24 * 60 * 60 * 1000).toISOString(),
        usageFrequency: Math.floor(Math.random() * 100),
        riskLevel: ['low', 'medium', 'high'][Math.floor(Math.random() * 3)] as 'low' | 'medium' | 'high'
      };
      
      setRoleStats(mockStats);
    } catch (err) {
      message.error('加载角色统计失败');
    } finally {
      setStatsLoading(false);
    }
  };

  // 加载权限分析
  const loadPermissionAnalysis = async (role: Role) => {
    try {
      if (!role.permissions || role.permissions.length === 0) {
        setPermissionAnalysis(null);
        return;
      }

      const byResource: { [key: string]: number } = {};
      const byAction: { [key: string]: number } = {};
      let systemPermissions = 0;
      let customPermissions = 0;

      role.permissions.forEach(permission => {
        byResource[permission.resource] = (byResource[permission.resource] || 0) + 1;
        byAction[permission.action] = (byAction[permission.action] || 0) + 1;
        
        if (permission.isSystem) {
          systemPermissions++;
        } else {
          customPermissions++;
        }
      });

      const riskScore = Math.min(100, (systemPermissions * 10 + customPermissions * 5) / role.permissions.length * 10);

      setPermissionAnalysis({
        byResource,
        byAction,
        systemPermissions,
        customPermissions,
        riskScore
      });
    } catch (err) {
      message.error('权限分析失败');
    }
  };

  // 获取风险等级颜色
  const getRiskLevelColor = (level: string) => {
    switch (level) {
      case 'low': return '#52c41a';
      case 'medium': return '#faad14';
      case 'high': return '#f5222d';
      default: return '#d9d9d9';
    }
  };

  // 获取风险等级文本
  const getRiskLevelText = (level: string) => {
    switch (level) {
      case 'low': return '低风险';
      case 'medium': return '中风险';
      case 'high': return '高风险';
      default: return '未知';
    }
  };

  // 导出角色配置
  const handleExportRole = (role: Role) => {
    const exportData = {
      name: role.name,
      displayName: role.displayName,
      description: role.description,
      permissions: role.permissions?.map(p => ({
        resource: p.resource,
        action: p.action,
        name: p.name
      })) || [],
      exportTime: new Date().toISOString(),
      version: '1.0'
    };

    const blob = new Blob([JSON.stringify(exportData, null, 2)], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `role_${role.name}_${new Date().getTime()}.json`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
    
    message.success('角色配置导出成功');
  };

  // 复制角色
  const handleCloneRole = (role: Role) => {
    setEditingRole(null);
    form.setFieldsValue({
      name: `${role.name}_copy`,
      displayName: `${role.displayName || role.name} (副本)`,
      description: `${role.description} (复制自 ${role.name})`,
    });
    setModalVisible(true);
    message.info('已预填充角色信息，请修改后保存');
  };

  // 处理权限模板选择
  const handleTemplateSelect = (templateId: string) => {
    setSelectedTemplate(templateId);
    if (templateId) {
      const template = permissionTemplates.find(t => t.id === templateId);
      if (template) {
        setSelectedPermissions(template.permissionIds);
        message.success(`已应用权限模板：${template.name}`);
      }
    }
  };

  // 清除权限模板选择
  const handleClearTemplate = () => {
    setSelectedTemplate('');
    setSelectedPermissions([]);
    message.success('已清除权限选择');
  };

  // 处理权限管理
  const handlePermissionManagement = async (role: Role) => {
    setPermissionRole(role);
    setPermissionDrawerVisible(true);
    setSelectedPermissions(role.permissions?.map(p => p.id) || []);
    loadAllPermissions();
  };

  // 加载所有权限
  const loadAllPermissions = async () => {
    setPermissionsLoading(true);
    try {
      await new Promise(resolve => setTimeout(resolve, 1000));
      
      const mockPermissions: Permission[] = [
        // 用户管理权限
        { id: '1', name: '用户列表', resource: 'user', action: 'list', description: '查看用户列表', isSystem: true },
        { id: '2', name: '用户详情', resource: 'user', action: 'view', description: '查看用户详细信息', isSystem: true },
        { id: '3', name: '创建用户', resource: 'user', action: 'create', description: '创建新用户', isSystem: true },
        { id: '4', name: '编辑用户', resource: 'user', action: 'update', description: '编辑用户信息', isSystem: true },
        { id: '5', name: '删除用户', resource: 'user', action: 'delete', description: '删除用户', isSystem: true },
        
        // 角色管理权限
        { id: '6', name: '角色列表', resource: 'role', action: 'list', description: '查看角色列表', isSystem: true },
        { id: '7', name: '角色详情', resource: 'role', action: 'view', description: '查看角色详细信息', isSystem: true },
        { id: '8', name: '创建角色', resource: 'role', action: 'create', description: '创建新角色', isSystem: true },
        { id: '9', name: '编辑角色', resource: 'role', action: 'update', description: '编辑角色信息', isSystem: true },
        { id: '10', name: '删除角色', resource: 'role', action: 'delete', description: '删除角色', isSystem: true },
        
        // 权限管理权限
        { id: '11', name: '权限列表', resource: 'permission', action: 'list', description: '查看权限列表', isSystem: true },
        { id: '12', name: '分配权限', resource: 'permission', action: 'assign', description: '分配权限给角色', isSystem: true },
        
        // 系统管理权限
        { id: '13', name: '系统设置', resource: 'system', action: 'config', description: '系统配置管理', isSystem: true },
        { id: '14', name: '日志查看', resource: 'system', action: 'log', description: '查看系统日志', isSystem: true },
        { id: '15', name: '数据备份', resource: 'system', action: 'backup', description: '数据备份操作', isSystem: true },
        
        // 自定义权限示例
        { id: '16', name: '报表查看', resource: 'report', action: 'view', description: '查看业务报表', isSystem: false },
        { id: '17', name: '报表导出', resource: 'report', action: 'export', description: '导出报表数据', isSystem: false },
        { id: '18', name: '数据分析', resource: 'analytics', action: 'view', description: '查看数据分析', isSystem: false },
      ];
      
      setAllPermissions(mockPermissions);
    } catch (err) {
      message.error('加载权限列表失败');
    } finally {
      setPermissionsLoading(false);
    }
  };

  // 保存权限分配
  const handleSavePermissions = async () => {
    if (!permissionRole) return;
    
    try {
      await new Promise(resolve => setTimeout(resolve, 500));
      
      const updatedPermissions = allPermissions.filter(p => selectedPermissions.includes(p.id));
      const updatedRole = { ...permissionRole, permissions: updatedPermissions };
      
      setRoles(prev => prev.map(role => 
        role.id === permissionRole.id ? updatedRole : role
      ));
      
      setPermissionDrawerVisible(false);
      message.success('权限分配成功');
    } catch (err) {
      message.error('权限分配失败');
    }
  };

  // 权限搜索过滤
  const getFilteredPermissions = () => {
    let filtered = allPermissions;
    
    if (permissionSearchText) {
      filtered = filtered.filter(p => 
        p.name.toLowerCase().includes(permissionSearchText.toLowerCase()) ||
        p.resource.toLowerCase().includes(permissionSearchText.toLowerCase()) ||
        p.action.toLowerCase().includes(permissionSearchText.toLowerCase()) ||
        p.description.toLowerCase().includes(permissionSearchText.toLowerCase())
      );
    }
    
    if (permissionFilter === 'assigned') {
      filtered = filtered.filter(p => selectedPermissions.includes(p.id));
    } else if (permissionFilter === 'unassigned') {
      filtered = filtered.filter(p => !selectedPermissions.includes(p.id));
    }
    
    return filtered;
  };

  // 按分组组织权限
  const getGroupedPermissions = () => {
    const filtered = getFilteredPermissions();
    
    if (permissionGroupBy === 'none') {
      return { '所有权限': filtered };
    }
    
    const grouped: { [key: string]: Permission[] } = {};
    
    filtered.forEach(permission => {
      const key = permissionGroupBy === 'resource' ? permission.resource : permission.action;
      if (!grouped[key]) {
        grouped[key] = [];
      }
      grouped[key].push(permission);
    });
    
    return grouped;
  };

  // 快速选择权限
  const handleQuickSelect = (type: 'all' | 'none' | 'system' | 'custom') => {
    switch (type) {
      case 'all':
        setSelectedPermissions(allPermissions.map(p => p.id));
        break;
      case 'none':
        setSelectedPermissions([]);
        break;
      case 'system':
        setSelectedPermissions(allPermissions.filter(p => p.isSystem).map(p => p.id));
        break;
      case 'custom':
        setSelectedPermissions(allPermissions.filter(p => !p.isSystem).map(p => p.id));
        break;
    }
  };

  // 权限模板应用
  const handleApplyTemplate = (templateId: string) => {
    const template = permissionTemplates.find(t => t.id === templateId);
    if (template) {
      setSelectedPermissions(template.permissionIds);
      message.success(`已应用权限模板: ${template.name}`);
    }
  };

  // 行选择配置
  const rowSelection = {
    selectedRowKeys,
    onChange: (keys: React.Key[]) => {
      setSelectedRowKeys(keys as string[]);
    },
    getCheckboxProps: (record: Role) => ({
      disabled: record.isSystem,
    }),
  };
  
  const filteredPermissions = useMemo(() => {
    const keyword = permissionSearch.trim().toLowerCase();
    if (!keyword) return permissions;
    return permissions.filter(p => {
      const text = `${p.name} ${p.description ?? ''} ${p.resource ?? ''} ${p.action ?? ''}`.toLowerCase();
      return text.includes(keyword);
    });
  }, [permissions, permissionSearch]);

  const groupedPermissions = useMemo(() => {
    const groups: Record<string, Permission[]> = {};
    filteredPermissions.forEach(p => {
      const key = p.resource || '未分类';
      if (!groups[key]) groups[key] = [];
      groups[key].push(p);
    });
    return Object.entries(groups).sort(([a], [b]) => a.localeCompare(b));
  }, [filteredPermissions]);

  return (
    <div>
      <Card>
        {/* 操作栏 */}
        <div style={{ marginBottom: 16 }}>
          <Row gutter={16}>
            <Col flex="auto">
              <Space>
                <Button
                  type="primary"
                  icon={<PlusOutlined />}
                  onClick={() => {
                    setEditingRole(null);
                    form.resetFields();
                    setModalVisible(true);
                  }}
                >
                  新建角色
                </Button>
                <Button
                  danger
                  icon={<DeleteOutlined />}
                  onClick={handleBatchDelete}
                  disabled={selectedRowKeys.length === 0}
                >
                  批量删除
                </Button>
                <Button
                  icon={<ReloadOutlined />}
                  onClick={() => loadRoles({ page, pageSize, search: searchText })}
                >
                  刷新
                </Button>
              </Space>
            </Col>
            <Col>
              <Search
                placeholder="搜索角色名称或描述"
                allowClear
                style={{ width: 300 }}
                onSearch={handleSearch}
              />
            </Col>
          </Row>
        </div>

        {/* 角色表格 */}
        <Table
          columns={columns}
          dataSource={roles}
          rowKey="id"
          loading={loading}
          rowSelection={rowSelection}
          scroll={{ x: 1200 }}
          pagination={{
            current: page,
            pageSize: pageSize,
            total: total,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total, range) => `第 ${range[0]}-${range[1]} 条，共 ${total} 条`,
            onChange: (newPage, newPageSize) => {
              setPage(newPage);
              setPageSize(newPageSize || 10);
              loadRoles({ page: newPage, pageSize: newPageSize || 10, search: searchText });
            },
          }}
        />
      </Card>

      {/* 角色编辑模态框 */}
      <Modal
        title={editingRole ? '编辑角色' : '新建角色'}
        open={modalVisible}
        onCancel={() => setModalVisible(false)}
        onOk={() => form.submit()}
        destroyOnClose
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSubmit}
        >
          <Form.Item
            name="name"
            label="角色标识"
            rules={[{ required: true, message: '请输入角色标识' }]}
          >
            <Input placeholder="请输入角色标识（英文）" disabled={!!editingRole} />
          </Form.Item>
          <Form.Item
            name="displayName"
            label="显示名称"
            rules={[{ required: true, message: '请输入显示名称' }]}
          >
            <Input placeholder="请输入角色显示名称" />
          </Form.Item>
          <Form.Item
            name="description"
            label="描述"
          >
            <Input.TextArea placeholder="请输入角色描述" rows={3} />
          </Form.Item>
        </Form>
      </Modal>

      {/* 角色详情抽屉 */}
      <Drawer
        title={`角色详情 - ${detailRole?.displayName || detailRole?.name}`}
        placement="right"
        width={800}
        open={detailDrawerVisible}
        onClose={() => setDetailDrawerVisible(false)}
        destroyOnClose
      >
        {detailRole && (
          <Tabs defaultActiveKey="basic">
            <TabPane tab="基本信息" key="basic">
              <Descriptions column={2} bordered>
                <Descriptions.Item label="角色标识">{detailRole.name}</Descriptions.Item>
                <Descriptions.Item label="显示名称">{detailRole.displayName || detailRole.name}</Descriptions.Item>
                <Descriptions.Item label="角色类型" span={2}>
                  <Tag color={detailRole.isSystem ? 'red' : 'blue'}>
                    {detailRole.isSystem ? '系统角色' : '自定义角色'}
                  </Tag>
                </Descriptions.Item>
                <Descriptions.Item label="描述" span={2}>{detailRole.description || '暂无描述'}</Descriptions.Item>
                <Descriptions.Item label="权限数量">
                  <Badge count={detailRole.permissions?.length || 0} showZero color="green" />
                </Descriptions.Item>
                <Descriptions.Item label="创建时间">
                  {new Date(detailRole.createdAt).toLocaleString()}
                </Descriptions.Item>
                <Descriptions.Item label="更新时间" span={2}>
                  {new Date(detailRole.updatedAt).toLocaleString()}
                </Descriptions.Item>
              </Descriptions>

              <Divider>操作</Divider>
              <Space>
                <Button icon={<EditOutlined />} onClick={() => handleEdit(detailRole)} disabled={detailRole.isSystem}>
                  编辑角色
                </Button>
                <Button icon={<SafetyOutlined />} onClick={() => handlePermissionManagement(detailRole)}>
                  权限管理
                </Button>
                <Button icon={<UserOutlined />} onClick={() => handleExportRole(detailRole)}>
                  导出配置
                </Button>
                <Button icon={<PlusOutlined />} onClick={() => handleCloneRole(detailRole)}>
                  复制角色
                </Button>
              </Space>
            </TabPane>

            <TabPane tab="统计信息" key="stats">
              {statsLoading ? (
                <div style={{ textAlign: 'center', padding: '50px' }}>加载中...</div>
              ) : roleStats ? (
                <Row gutter={16}>
                  <Col span={6}>
                    <Statistic
                      title="关联用户数"
                      value={roleStats.userCount}
                      prefix={<UserOutlined />}
                      suffix="个"
                    />
                  </Col>
                  <Col span={6}>
                    <Statistic
                      title="权限数量"
                      value={roleStats.permissionCount}
                      prefix={<SafetyOutlined />}
                      suffix="个"
                    />
                  </Col>
                  <Col span={6}>
                    <Statistic
                      title="使用频率"
                      value={roleStats.usageFrequency}
                      suffix="%"
                    />
                  </Col>
                  <Col span={6}>
                    <Statistic
                      title="风险等级"
                      value={getRiskLevelText(roleStats.riskLevel)}
                      valueStyle={{ color: getRiskLevelColor(roleStats.riskLevel) }}
                    />
                  </Col>
                  <Col span={24} style={{ marginTop: 16 }}>
                    <Descriptions column={1} bordered>
                      <Descriptions.Item label="最后使用时间">
                        {new Date(roleStats.lastUsed).toLocaleString()}
                      </Descriptions.Item>
                    </Descriptions>
                  </Col>
                </Row>
              ) : (
                <div style={{ textAlign: 'center', padding: '50px', color: '#999' }}>
                  暂无统计数据
                </div>
              )}
            </TabPane>

            <TabPane tab="权限分析" key="analysis">
              {permissionAnalysis ? (
                <div>
                  <Row gutter={16} style={{ marginBottom: 24 }}>
                    <Col span={8}>
                      <Statistic
                        title="系统权限"
                        value={permissionAnalysis.systemPermissions}
                        suffix="个"
                        valueStyle={{ color: '#f5222d' }}
                      />
                    </Col>
                    <Col span={8}>
                      <Statistic
                        title="自定义权限"
                        value={permissionAnalysis.customPermissions}
                        suffix="个"
                        valueStyle={{ color: '#52c41a' }}
                      />
                    </Col>
                    <Col span={8}>
                      <Statistic
                        title="风险评分"
                        value={permissionAnalysis.riskScore}
                        suffix="/100"
                        valueStyle={{ 
                          color: permissionAnalysis.riskScore > 70 ? '#f5222d' : 
                                 permissionAnalysis.riskScore > 40 ? '#faad14' : '#52c41a'
                        }}
                      />
                    </Col>
                  </Row>

                  <Divider>按资源分布</Divider>
                  <Row gutter={16}>
                    {Object.entries(permissionAnalysis.byResource).map(([resource, count]) => (
                      <Col span={6} key={resource} style={{ marginBottom: 16 }}>
                        <Card size="small">
                          <Statistic
                            title={resource}
                            value={count}
                            suffix="个权限"
                            valueStyle={{ fontSize: 16 }}
                          />
                        </Card>
                      </Col>
                    ))}
                  </Row>

                  <Divider>按操作分布</Divider>
                  <Row gutter={16}>
                    {Object.entries(permissionAnalysis.byAction).map(([action, count]) => (
                      <Col span={6} key={action} style={{ marginBottom: 16 }}>
                        <Card size="small">
                          <Statistic
                            title={action}
                            value={count}
                            suffix="个权限"
                            valueStyle={{ fontSize: 16 }}
                          />
                        </Card>
                      </Col>
                    ))}
                  </Row>
                </div>
              ) : (
                <div style={{ textAlign: 'center', padding: '50px', color: '#999' }}>
                  该角色暂无权限
                </div>
              )}
            </TabPane>

            <TabPane tab="活动日志" key="activities">
              {activitiesLoading ? (
                <div style={{ textAlign: 'center', padding: '50px' }}>加载中...</div>
              ) : (
                <Timeline>
                  {roleActivities.map(activity => (
                    <Timeline.Item key={activity.id}>
                      <div>
                        <div style={{ fontWeight: 'bold', marginBottom: 4 }}>
                          {activity.description}
                        </div>
                        <div style={{ fontSize: 12, color: '#999' }}>
                          操作人: {activity.operator} | 时间: {new Date(activity.timestamp).toLocaleString()}
                        </div>
                      </div>
                    </Timeline.Item>
                  ))}
                </Timeline>
              )}
            </TabPane>
          </Tabs>
        )}
      </Drawer>

      {/* 权限管理抽屉 */}
      <Drawer
        title={`权限管理 - ${permissionRole?.displayName || permissionRole?.name}`}
        placement="right"
        width={800}
        open={permissionDrawerVisible}
        onClose={() => setPermissionDrawerVisible(false)}
        destroyOnClose
        extra={
          <Space>
            <Button onClick={() => setPermissionDrawerVisible(false)}>取消</Button>
            <Button type="primary" onClick={handleSavePermissions}>保存</Button>
          </Space>
        }
      >
        <Space direction="vertical" style={{ width: '100%' }} size="large">
          {/* 权限模板选择 */}
          <Card size="small" title="权限模板">
            <Space style={{ width: '100%' }} direction="vertical">
              <Select
                style={{ width: '100%' }}
                placeholder="选择权限模板快速配置"
                value={selectedTemplate}
                onChange={handleTemplateSelect}
                allowClear
                onClear={handleClearTemplate}
              >
                {permissionTemplates.map(template => (
                  <Option key={template.id} value={template.id}>
                    {template.name} - {template.description}
                  </Option>
                ))}
              </Select>
            </Space>
          </Card>

          {/* 快速操作 */}
          <Card size="small" title="快速操作">
            <Space wrap>
              <Button size="small" onClick={() => handleQuickSelect('all')}>全选</Button>
              <Button size="small" onClick={() => handleQuickSelect('none')}>清空</Button>
              <Button size="small" onClick={() => handleQuickSelect('system')}>选择系统权限</Button>
              <Button size="small" onClick={() => handleQuickSelect('custom')}>选择自定义权限</Button>
            </Space>
          </Card>

          {/* 搜索和过滤 */}
          <Card size="small" title="搜索过滤">
            <Space style={{ width: '100%' }} direction="vertical">
              <Input
                placeholder="搜索权限名称、资源或操作"
                value={permissionSearchText}
                onChange={(e) => setPermissionSearchText(e.target.value)}
                prefix={<SearchOutlined />}
                allowClear
              />
              <Row gutter={16}>
                <Col span={12}>
                  <Select
                    style={{ width: '100%' }}
                    value={permissionFilter}
                    onChange={setPermissionFilter}
                  >
                    <Option value="all">全部权限</Option>
                    <Option value="assigned">已分配</Option>
                    <Option value="unassigned">未分配</Option>
                  </Select>
                </Col>
                <Col span={12}>
                  <Select
                    style={{ width: '100%' }}
                    value={permissionGroupBy}
                    onChange={setPermissionGroupBy}
                  >
                    <Option value="resource">按资源分组</Option>
                    <Option value="action">按操作分组</Option>
                    <Option value="none">不分组</Option>
                  </Select>
                </Col>
              </Row>
            </Space>
          </Card>

          {/* 权限列表 */}
          <Card 
            size="small" 
            title={`权限列表 (已选择 ${selectedPermissions.length} 个)`}
            loading={permissionsLoading}
          >
            <Checkbox.Group
              value={selectedPermissions}
              onChange={setSelectedPermissions}
              style={{ width: '100%' }}
            >
              {Object.entries(getGroupedPermissions()).map(([groupName, permissions]) => (
                <div key={groupName} style={{ marginBottom: 16 }}>
                  <div style={{ fontWeight: 'bold', marginBottom: 8, color: '#1890ff' }}>
                    {groupName} ({permissions.length})
                  </div>
                  <Row gutter={[8, 8]}>
                    {permissions.map(permission => (
                      <Col span={24} key={permission.id}>
                        <Checkbox value={permission.id} style={{ width: '100%' }}>
                          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', width: '100%' }}>
                            <div>
                              <div style={{ fontWeight: 'bold' }}>{permission.name}</div>
                              <div style={{ fontSize: 12, color: '#666' }}>
                                {permission.resource}.{permission.action} - {permission.description}
                              </div>
                            </div>
                            <Tag color={permission.isSystem ? 'red' : 'blue'} size="small">
                              {permission.isSystem ? '系统' : '自定义'}
                            </Tag>
                          </div>
                        </Checkbox>
                      </Col>
                    ))}
                  </Row>
                </div>
              ))}
            </Checkbox.Group>
            
            {getFilteredPermissions().length === 0 && (
              <div style={{ textAlign: 'center', color: '#999', padding: '40px' }}>
                {permissionSearchText ? '没有找到匹配的权限' : '暂无权限数据'}
              </div>
            )}
          </Card>
        </Space>
      </Drawer>
    </div>
  );
};

export default RoleManagement;