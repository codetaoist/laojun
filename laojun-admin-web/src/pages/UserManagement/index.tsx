import { useState, useEffect, useMemo } from 'react';
import { 
  Table, Card, Button, Space, Input, Tag, message, Modal, Form, 
  Select, Popconfirm, Avatar, Switch, Drawer, Transfer, Divider,
  Descriptions, Timeline, Tabs, Badge, Row, Col, Statistic, DatePicker,
  Dropdown
} from 'antd';
import { 
  PlusOutlined, SearchOutlined, EditOutlined, DeleteOutlined, 
  ReloadOutlined, UserOutlined, TeamOutlined, LockOutlined,
  EyeOutlined, HistoryOutlined, SafetyOutlined, InfoCircleOutlined,
  FilterOutlined, MoreOutlined
} from '@ant-design/icons';
import { User, Role } from '@/types';
import { userService } from '@/services/user';
import { roleService } from '@/services/role';
import { RoleUtils } from '@/utils/roleUtils';
import { RoleManagementTest } from '@/utils/roleManagementTest';
import { RoleAssignmentTester } from '@/utils/testRoleAssignment';

const { Search } = Input;
const { Option } = Select;
const { TabPane } = Tabs;

// 用户活动日志类型
interface UserActivity {
  id: string;
  action: string;
  description: string;
  ip: string;
  userAgent: string;
  createdAt: string;
  status: 'success' | 'failed' | 'warning';
}

const UserManagement: React.FC = () => {
  const [users, setUsers] = useState<User[]>([]);
  const [roles, setRoles] = useState<Role[]>([]);
  const [loading, setLoading] = useState(false);
  const [searchText, setSearchText] = useState('');
  const [selectedRowKeys, setSelectedRowKeys] = useState<string[]>([]);
  const [modalVisible, setModalVisible] = useState(false);
  const [roleDrawerVisible, setRoleDrawerVisible] = useState(false);
  const [detailDrawerVisible, setDetailDrawerVisible] = useState(false);
  const [editingUser, setEditingUser] = useState<User | null>(null);
  const [currentUser, setCurrentUser] = useState<User | null>(null);
  const [targetKeys, setTargetKeys] = useState<string[]>([]);
  const [userActivities, setUserActivities] = useState<UserActivity[]>([]);
  const [activitiesLoading, setActivitiesLoading] = useState(false);
  const [form] = Form.useForm();
  const [searchForm] = Form.useForm();
  const [pagination, setPagination] = useState({ page: 1, pageSize: 10, total: 0 });
  const [advancedSearchVisible, setAdvancedSearchVisible] = useState(false);
  const [searchFilters, setSearchFilters] = useState<{
    status?: string;
    roleId?: string;
    dateRange?: [string, string];
  }>({});
  
  // 批量操作相关状态
  const [bulkRoleModalVisible, setBulkRoleModalVisible] = useState(false);
  const [bulkRoleForm] = Form.useForm();

  // 加载用户数据（直连后端）
  const loadUsers = async (
    page = pagination.page,
    pageSize = pagination.pageSize,
    search: string = searchText,
    filters = searchFilters
  ) => {
    setLoading(true);
    try {
      const params: any = { page, pageSize, search };
      
      // 添加筛选条件
      if (filters.status) {
        params.status = filters.status;
      }
      if (filters.roleId) {
        params.roleId = filters.roleId;
      }
      if (filters.dateRange && filters.dateRange[0] && filters.dateRange[1]) {
        params.startDate = filters.dateRange[0];
        params.endDate = filters.dateRange[1];
      }
      
      const res = await userService.getUsers(params);
      setUsers(res.items);
      setPagination({ page: res.page, pageSize: res.pageSize, total: res.total });
    } catch (error) {
      message.error('加载用户数据失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadUsers();
  }, []);

  // 加载角色列表（用于分配角色）
  const loadRoles = async () => {
    try {
      const res: any = await roleService.getRoles({ page: 1, pageSize: 100 });
      const list: Role[] = res?.items || res?.data || [];
      setRoles(list);
    } catch (error) {
      message.error('加载角色列表失败');
    }
  };

  // 处理状态切换（直连后端）
  const handleStatusChange = async (userId: string, checked: boolean) => {
    try {
      const updated = await userService.setActive(userId, checked);
      setUsers(prev => prev.map(u => u.id === userId ? updated : u));
      message.success(`用户状态已${checked ? '启用' : '禁用'}`);
    } catch (error) {
      message.error('状态更新失败');
    }
  };

  // 打开角色管理抽屉
  const handleRoleManagement = async (user: User) => {
    setCurrentUser(user);
    // 确保角色ID数组的数据类型一致性
    const userRoleIds = user.roles?.map(r => r.id) || [];
    setTargetKeys(userRoleIds);
    
    // 如果角色列表还没加载，先加载
    if (!roles || roles.length === 0) {
      await loadRoles();
    }
    setRoleDrawerVisible(true);
  };

  // 保存用户角色分配（增强版本）
  const handleAssignRoles = async () => {
    if (!currentUser) return;
    
    try {
      // 使用工具函数验证角色ID
      const validation = RoleUtils.validateRoleIds(targetKeys);
      if (!validation.isValid) {
        message.error(`角色验证失败: ${validation.errors.join(', ')}`);
        return;
      }

      // 验证角色是否存在于可用角色列表中
      const existenceCheck = RoleUtils.validateRolesExist(targetKeys, roles);
      if (!existenceCheck.isValid) {
        message.error(`以下角色不存在: ${existenceCheck.missingRoles.join(', ')}`);
        return;
      }

      // 获取角色变更信息
      const oldRoleIds = RoleUtils.extractRoleIds(currentUser);
      const roleChanges = RoleUtils.getRoleChanges(oldRoleIds, targetKeys);
      const changeDescription = RoleUtils.formatRoleChanges(roleChanges, roles);
      
      console.log('角色变更:', changeDescription);

      // 调用API分配角色
       await userService.assignRoles(currentUser.id, targetKeys);
      
      // 重新获取该用户详情以刷新列表中的角色显示
      try {
        const freshUser = await userService.getUser(currentUser.id);
        setUsers(prev => prev.map(u => (u.id === freshUser.id ? freshUser : u)));
        
        // 验证角色分配是否成功
         const assignedRoleIds = freshUser.roles?.map(r => r.id) || [];
         const expectedRoleIds = [...targetKeys].sort();
         const actualRoleIds = assignedRoleIds.sort();
         
         if (!RoleUtils.compareRoleArrays(expectedRoleIds, actualRoleIds)) {
          message.warning('角色分配可能不完整，请检查结果');
        } else {
          message.success('角色分配成功');
        }
      } catch (refreshError) {
        message.success('角色分配成功，但刷新用户信息失败');
        console.error('Failed to refresh user data:', refreshError);
      }
      
      setRoleDrawerVisible(false);
      setCurrentUser(null);
      setTargetKeys([]);
    } catch (error: any) {
      console.error('Role assignment failed:', error);
      
      // 根据错误类型提供更具体的错误信息
      if (error?.response?.status === 400) {
        message.error(`角色分配失败: ${error.response.data?.details || '请求参数无效'}`);
      } else if (error?.response?.status === 403) {
        message.error('权限不足，无法分配角色');
      } else if (error?.response?.status === 404) {
        message.error('用户或角色不存在');
      } else {
        message.error('角色分配失败，请稍后重试');
      }
    }
  };

  // 查看用户详情
  const handleViewDetail = async (user: User) => {
    setCurrentUser(user);
    setDetailDrawerVisible(true);
    // 加载用户活动日志
    await loadUserActivities(user.id);
  };

  // 加载用户活动日志
  const loadUserActivities = async (userId: string) => {
    setActivitiesLoading(true);
    try {
      // 模拟活动日志数据，实际应该从后端API获取
      const mockActivities: UserActivity[] = [
        {
          id: '1',
          action: 'login',
          description: '用户登录系统',
          ip: '192.168.1.100',
          userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36',
          createdAt: new Date(Date.now() - 1000 * 60 * 30).toISOString(),
          status: 'success'
        },
        {
          id: '2',
          action: 'profile_update',
          description: '更新个人资料',
          ip: '192.168.1.100',
          userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36',
          createdAt: new Date(Date.now() - 1000 * 60 * 60 * 2).toISOString(),
          status: 'success'
        },
        {
          id: '3',
          action: 'password_change',
          description: '修改密码',
          ip: '192.168.1.100',
          userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36',
          createdAt: new Date(Date.now() - 1000 * 60 * 60 * 24).toISOString(),
          status: 'success'
        },
        {
          id: '4',
          action: 'login_failed',
          description: '登录失败 - 密码错误',
          ip: '192.168.1.101',
          userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36',
          createdAt: new Date(Date.now() - 1000 * 60 * 60 * 24 * 2).toISOString(),
          status: 'failed'
        }
      ];
      setUserActivities(mockActivities);
    } catch (error) {
      message.error('加载用户活动日志失败');
    } finally {
      setActivitiesLoading(false);
    }
  };

  // 表格列配置
  const columns = [
    {
      title: '用户信息',
      key: 'userInfo',
      width: 180,
      fixed: 'left',
      render: (_: any, record: User) => (
        <Space>
          <Avatar 
            src={record.avatar} 
            icon={<UserOutlined />}
            size="small"
          />
          <div>
            <div style={{ fontWeight: 500 }}>{record.username}</div>
            <div style={{ fontSize: '12px', color: '#999' }}>{record.name || record.username}</div>
          </div>
        </Space>
      ),
    },
    {
      title: '邮箱',
      dataIndex: 'email',
      key: 'email',
      width: 200,
      ellipsis: true,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      align: 'center',
      render: (status: string, record: User) => (
        <Switch
          checked={status === 'active'}
          onChange={(checked) => handleStatusChange(record.id, checked)}
          checkedChildren="启用"
          unCheckedChildren="禁用"
        />
      ),
    },
    {
      title: '角色',
      dataIndex: 'roles',
      key: 'roles',
      width: 200,
      render: (roles: User['roles']) => (
        <Space wrap>
          {roles.map(role => (
            <Tag 
              key={role.id} 
              color={role.isSystem ? 'red' : 'blue'}
              icon={<TeamOutlined />}
            >
              {role.name}
            </Tag>
          ))}
        </Space>
      ),
    },
    {
      title: '最后登录',
      dataIndex: 'lastLoginAt',
      key: 'lastLoginAt',
      width: 160,
      render: (date: string) => date ? new Date(date).toLocaleString('zh-CN') : '从未登录',
      sorter: true,
    },
    {
      title: '创建时间',
      dataIndex: 'createdAt',
      key: 'createdAt',
      width: 160,
      render: (date: string) => new Date(date).toLocaleString('zh-CN'),
      sorter: true,
    },
    {
      title: '操作',
      key: 'action',
      width: 200,
      fixed: 'right',
      render: (_: any, record: User) => {
        const items = [
          {
            key: 'detail',
            icon: <EyeOutlined />,
            label: '查看详情',
            onClick: () => handleViewDetail(record),
          },
          {
            key: 'role',
            icon: <TeamOutlined />,
            label: '管理角色',
            onClick: () => handleRoleManagement(record),
          },
          {
            key: 'edit',
            icon: <EditOutlined />,
            label: '编辑用户',
            onClick: () => handleEdit(record),
          },
          {
            key: 'reset',
            icon: <LockOutlined />,
            label: '重置密码',
            onClick: () => handleResetPassword(record),
          },
          {
            type: 'divider',
          },
          {
            key: 'delete',
            icon: <DeleteOutlined />,
            label: '删除用户',
            danger: true,
            onClick: () => {
              Modal.confirm({
                title: '确认删除',
                content: `确定要删除用户 "${record.username}" 吗？此操作不可恢复。`,
                okText: '确认删除',
                okType: 'danger',
                cancelText: '取消',
                onOk: () => handleDelete(record),
              });
            },
          },
        ];

        return (
          <Space size={4}>
            <Button
              type="primary"
              size="small"
              icon={<EyeOutlined />}
              onClick={() => handleViewDetail(record)}
              style={{ minWidth: '60px' }}
            >
              详情
            </Button>
            <Dropdown
              menu={{ items }}
              placement="bottomRight"
              trigger={['click']}
              arrow={{ pointAtCenter: true }}
            >
              <Button 
                size="small" 
                icon={<MoreOutlined />}
                style={{ 
                  minWidth: '60px',
                  borderColor: '#d9d9d9',
                  color: '#666'
                }}
              >
                更多
              </Button>
            </Dropdown>
          </Space>
        );
      },
    },
  ];

  // 处理搜索
  const handleSearch = (value: string) => {
    setSearchText(value);
    // 直接使用传入的 value 进行查询，避免使用旧的 searchText
    loadUsers(1, pagination.pageSize, value, searchFilters);
  };

  // 处理高级搜索
  const handleAdvancedSearch = (values: any) => {
    const filters = {
      status: values.status,
      roleId: values.roleId,
      dateRange: values.dateRange ? [
        values.dateRange[0].format('YYYY-MM-DD'),
        values.dateRange[1].format('YYYY-MM-DD')
      ] : undefined
    };
    setSearchFilters(filters);
    loadUsers(1, pagination.pageSize, searchText, filters);
    setAdvancedSearchVisible(false);
  };

  // 重置搜索条件
  const handleResetSearch = () => {
    setSearchFilters({});
    setSearchText('');
    searchForm.resetFields();
    loadUsers(1, pagination.pageSize, '', {});
    setAdvancedSearchVisible(false);
  };

  // 批量状态变更
  const handleBulkStatusChange = async (status: boolean) => {
    try {
      // 这里应该调用后端API进行批量状态更新
      // await userService.bulkUpdateStatus(selectedRowKeys, status);
      
      // 模拟API调用
      await new Promise(resolve => setTimeout(resolve, 1000));
      
      message.success(`批量${status ? '启用' : '禁用'}成功`);
      setSelectedRowKeys([]);
      loadUsers(pagination.page, pagination.pageSize);
    } catch (error) {
      message.error(`批量${status ? '启用' : '禁用'}失败`);
    }
  };

  // 批量分配角色
  const handleBulkRoleAssign = async (values: { roleIds: string[] }) => {
    try {
      // 这里应该调用后端API进行批量角色分配
      // await userService.bulkAssignRoles(selectedRowKeys, values.roleIds);
      
      // 模拟API调用
      await new Promise(resolve => setTimeout(resolve, 1000));
      
      message.success('批量分配角色成功');
      setBulkRoleModalVisible(false);
      bulkRoleForm.resetFields();
      setSelectedRowKeys([]);
      loadUsers(pagination.page, pagination.pageSize);
    } catch (error) {
      message.error('批量分配角色失败');
    }
  };

  // 处理新增
  const handleAdd = () => {
    setEditingUser(null);
    form.resetFields();
    setModalVisible(true);
  };

  // 处理编辑
  const handleEdit = (user: User) => {
    setEditingUser(user);
    form.setFieldsValue({
      username: user.username,
      email: user.email,
      status: user.status,
    });
    setModalVisible(true);
  };

  // 处理重置密码（直连后端）
  const handleResetPassword = (user: User) => {
    Modal.confirm({
      title: '重置密码',
      content: `确定要重置用户 "${user.username}" 的密码吗？`,
      onOk: async () => {
        try {
          const newPwd = Math.random().toString(36).slice(2, 10);
          await userService.resetPassword(user.id, newPwd);
          message.success('密码重置成功');
        } catch (error) {
          message.error('密码重置失败');
        }
      },
    });
  };

  // 处理删除（直连后端）
  const handleDelete = async (user: User) => {
    try {
      await userService.deleteUser(user.id);
      message.success('删除成功');
      loadUsers(pagination.page, pagination.pageSize);
    } catch (error) {
      message.error('删除失败');
    }
  };

  // 处理表单提交（直连后端）
  const handleSubmit = async (values: any) => {
    try {
      if (editingUser) {
        const isActive = values.status ? values.status === 'active' : undefined;
        const updated = await userService.updateUser(editingUser.id, {
          email: values.email,
          isActive,
        });
        message.success('更新成功');
        setModalVisible(false);
        setEditingUser(null);
        // 更新列表
        setUsers(prev => prev.map(u => (u.id === updated.id ? updated : u)));
      } else {
        // 新增用户
        const created = await userService.createUser({
          username: values.username,
          email: values.email,
          password: values.password,
        });
        message.success('创建成功');
        setModalVisible(false);
        // 重新加载第一页，确保可见
        loadUsers(1, pagination.pageSize);
      }
    } catch (error) {
      message.error('保存失败');
    }
  };

  // Transfer组件的数据源处理（增强版本）
  const transferData = useMemo(() => {
    if (!roles || roles.length === 0) return [];
    
    return roles
      .filter(role => role.id && role.name) // 过滤无效数据
      .map(role => ({
        key: role.id,
        title: role.name,
        description: role.description || '',
        disabled: false, // 可以根据业务逻辑设置某些角色不可选
      }));
  }, [roles]);

  return (
    <div>
      <Card>
        <div style={{ marginBottom: '16px' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '12px' }}>
            <Space>
              <Search
                placeholder="搜索用户名或邮箱"
                allowClear
                style={{ width: 300 }}
                onSearch={handleSearch}
              />
              <Button 
                icon={<FilterOutlined />} 
                onClick={() => setAdvancedSearchVisible(true)}
              >
                高级搜索
              </Button>
              <Button icon={<ReloadOutlined />} onClick={() => loadUsers(pagination.page, pagination.pageSize)}>
                刷新
              </Button>
            </Space>
            <Space>
              <Button type="primary" icon={<PlusOutlined />} onClick={handleAdd}>
                新增用户
              </Button>
            </Space>
          </div>
          
          {/* 当前筛选条件显示 */}
          {(searchFilters.status || searchFilters.roleId || searchFilters.dateRange) && (
            <div style={{ marginBottom: '12px' }}>
              <Space wrap>
                <span style={{ color: '#666' }}>当前筛选:</span>
                {searchFilters.status && (
                  <Tag closable onClose={() => {
                    const newFilters = { ...searchFilters };
                    delete newFilters.status;
                    setSearchFilters(newFilters);
                    loadUsers(1, pagination.pageSize, searchText, newFilters);
                  }}>
                    状态: {searchFilters.status === 'active' ? '正常' : '禁用'}
                  </Tag>
                )}
                {searchFilters.roleId && (
                  <Tag closable onClose={() => {
                    const newFilters = { ...searchFilters };
                    delete newFilters.roleId;
                    setSearchFilters(newFilters);
                    loadUsers(1, pagination.pageSize, searchText, newFilters);
                  }}>
                    角色: {roles.find(r => r.id === searchFilters.roleId)?.name || '未知'}
                  </Tag>
                )}
                {searchFilters.dateRange && (
                  <Tag closable onClose={() => {
                    const newFilters = { ...searchFilters };
                    delete newFilters.dateRange;
                    setSearchFilters(newFilters);
                    loadUsers(1, pagination.pageSize, searchText, newFilters);
                  }}>
                    创建时间: {searchFilters.dateRange[0]} ~ {searchFilters.dateRange[1]}
                  </Tag>
                )}
                <Button type="link" size="small" onClick={handleResetSearch}>
                  清除所有筛选
                </Button>
              </Space>
            </div>
          )}
        </div>

        {/* 批量操作 */}
        {selectedRowKeys.length > 0 && (
          <div style={{ 
            marginBottom: '16px', 
            padding: '12px 16px', 
            backgroundColor: '#e6f7ff', 
            borderRadius: '8px',
            border: '1px solid #91d5ff',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between'
          }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
              <Badge count={selectedRowKeys.length} color="#1890ff">
                <span style={{ color: '#1890ff', fontWeight: 500 }}>
                  已选择 {selectedRowKeys.length} 项
                </span>
              </Badge>
              <Space size="small">
                <Button 
                  size="small" 
                  type="primary"
                  ghost
                  onClick={() => handleBulkStatusChange(true)}
                >
                  批量启用
                </Button>
                <Button 
                  size="small" 
                  onClick={() => handleBulkStatusChange(false)}
                >
                  批量禁用
                </Button>
                <Button 
                  size="small" 
                  onClick={() => setBulkRoleModalVisible(true)}
                >
                  批量分配角色
                </Button>
              </Space>
            </div>
            <Button 
              size="small" 
              type="text"
              onClick={() => setSelectedRowKeys([])}
              style={{ color: '#666' }}
            >
              取消选择
            </Button>
          </div>
        )}

        <Table
          columns={columns}
          dataSource={users}
          loading={loading}
          rowKey="id"
          scroll={{ x: 1200 }}
          rowSelection={{
            selectedRowKeys,
            onChange: setSelectedRowKeys,
          }}
          pagination={{
            current: pagination.page,
            total: pagination.total,
            pageSize: pagination.pageSize,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total: number, range: [number, number]) => `第 ${range[0]}-${range[1]} 条/共 ${total} 条`,
            onChange: (page: number, pageSize?: number) => {
              setPagination(prev => ({ ...prev, page, pageSize: pageSize || prev.pageSize }));
              loadUsers(page, pageSize || pagination.pageSize);
            },
          }}
        />
      </Card>

      {/* 用户编辑模态框 */}
      <Modal
        title={editingUser ? '编辑用户' : '新增用户'}
        open={modalVisible}
        onCancel={() => setModalVisible(false)}
        footer={null}
        width={600}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSubmit}
        >
          <Form.Item
            name="username"
            label="用户名"
            rules={[
              { required: true, message: '请输入用户名' },
              { min: 3, message: '用户名至少3个字符' },
            ]}
          >
            <Input placeholder="请输入用户名" disabled={!!editingUser} />
          </Form.Item>

          <Form.Item
            name="email"
            label="邮箱"
            rules={[
              { required: true, message: '请输入邮箱' },
              { type: 'email', message: '请输入有效的邮箱地址' },
            ]}
          >
            <Input placeholder="请输入邮箱" />
          </Form.Item>

          {!editingUser && (
            <Form.Item
              name="password"
              label="密码"
              rules={[
                { required: true, message: '请输入密码' },
                { min: 6, message: '密码至少6个字符' },
              ]}
            >
              <Input.Password placeholder="请输入密码" />
            </Form.Item>
          )}

          <Form.Item
            name="status"
            label="状态"
            rules={[{ required: true, message: '请选择状态' }]}
          >
            <Select placeholder="请选择状态">
              <Option value="active">正常</Option>
              <Option value="inactive">禁用</Option>
            </Select>
          </Form.Item>

          {/* 角色选择暂不接通，后续实现 */}
          {/* <Form.Item name="roleIds" label="角色" rules={[{ required: true, message: '请选择角色' }]}>...
          </Form.Item> */}

          <Form.Item style={{ marginBottom: 0, textAlign: 'right' }}>
            <Space>
              <Button onClick={() => setModalVisible(false)}>
                取消
              </Button>
              <Button type="primary" htmlType="submit">
                {editingUser ? '更新' : '创建'}
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>

      {/* 角色管理抽屉 */}
      <Drawer
        title={`管理用户角色 - ${currentUser?.username || ''}`}
        placement="right"
        onClose={() => setRoleDrawerVisible(false)}
        open={roleDrawerVisible}
        width={600}
        extra={
          <Space>
            <Button onClick={() => setRoleDrawerVisible(false)}>取消</Button>
            <Button type="primary" onClick={handleAssignRoles}>保存</Button>
          </Space>
        }
      >
        <div style={{ marginBottom: 16 }}>
          <p>选择要为该用户分配的角色。</p>
        </div>
        <Transfer
          dataSource={transferData}
          titles={['可选角色', '已分配']}
          targetKeys={targetKeys}
          onChange={(nextTargetKeys) => setTargetKeys(nextTargetKeys as string[])}
          render={item => item.title}
          listStyle={{ width: 250, height: 360 }}
        />
        <Divider />
      </Drawer>

      {/* 用户详情抽屉 */}
      <Drawer
        title={
          <Space>
            <Avatar src={currentUser?.avatar} icon={<UserOutlined />} />
            <span>{currentUser?.username} - 用户详情</span>
          </Space>
        }
        placement="right"
        onClose={() => setDetailDrawerVisible(false)}
        open={detailDrawerVisible}
        width={800}
      >
        {currentUser && (
          <Tabs defaultActiveKey="basic">
            <TabPane 
              tab={
                <span>
                  <InfoCircleOutlined />
                  基本信息
                </span>
              } 
              key="basic"
            >
              <Row gutter={[16, 16]}>
                <Col span={24}>
                  <Card title="用户统计" size="small">
                    <Row gutter={16}>
                      <Col span={6}>
                        <Statistic 
                          title="登录次数" 
                          value={Math.floor(Math.random() * 100) + 50} 
                          prefix={<UserOutlined />}
                        />
                      </Col>
                      <Col span={6}>
                        <Statistic 
                          title="在线时长" 
                          value={Math.floor(Math.random() * 500) + 100} 
                          suffix="小时"
                        />
                      </Col>
                      <Col span={6}>
                        <Statistic 
                          title="角色数量" 
                          value={currentUser.roles.length} 
                          prefix={<TeamOutlined />}
                        />
                      </Col>
                      <Col span={6}>
                        <Statistic 
                          title="状态" 
                          value={currentUser.status === 'active' ? '正常' : '禁用'}
                          valueStyle={{ color: currentUser.status === 'active' ? '#3f8600' : '#cf1322' }}
                        />
                      </Col>
                    </Row>
                  </Card>
                </Col>
                <Col span={24}>
                  <Descriptions title="详细信息" bordered column={2}>
                    <Descriptions.Item label="用户名">{currentUser.username}</Descriptions.Item>
                    <Descriptions.Item label="邮箱">{currentUser.email}</Descriptions.Item>
                    <Descriptions.Item label="姓名">{currentUser.name || '未设置'}</Descriptions.Item>
                    <Descriptions.Item label="状态">
                      <Badge 
                        status={currentUser.status === 'active' ? 'success' : 'error'} 
                        text={currentUser.status === 'active' ? '正常' : '禁用'} 
                      />
                    </Descriptions.Item>
                    <Descriptions.Item label="创建时间">
                      {new Date(currentUser.createdAt).toLocaleString('zh-CN')}
                    </Descriptions.Item>
                    <Descriptions.Item label="最后登录">
                      {currentUser.lastLoginAt 
                        ? new Date(currentUser.lastLoginAt).toLocaleString('zh-CN') 
                        : '从未登录'
                      }
                    </Descriptions.Item>
                    <Descriptions.Item label="个人简介" span={2}>
                      {currentUser.bio || '暂无简介'}
                    </Descriptions.Item>
                  </Descriptions>
                </Col>
              </Row>
            </TabPane>
            
            <TabPane 
              tab={
                <span>
                  <SafetyOutlined />
                  权限详情
                </span>
              } 
              key="permissions"
            >
              <Card title="角色信息" size="small" style={{ marginBottom: 16 }}>
                <Space wrap>
                  {currentUser.roles.map(role => (
                    <Tag 
                      key={role.id} 
                      color={role.isSystem ? 'red' : 'blue'}
                      icon={<TeamOutlined />}
                      style={{ padding: '4px 8px', fontSize: '14px' }}
                    >
                      {role.name}
                      {role.isSystem && <span style={{ marginLeft: 4 }}>(系统)</span>}
                    </Tag>
                  ))}
                </Space>
              </Card>
              
              <Card title="权限列表" size="small">
                {currentUser.roles.length > 0 ? (
                  <div>
                    {currentUser.roles.map(role => (
                      <div key={role.id} style={{ marginBottom: 16 }}>
                        <h4 style={{ marginBottom: 8, color: '#1890ff' }}>
                          <TeamOutlined /> {role.name}
                        </h4>
                        <Space wrap>
                          {role.permissions?.map(permission => (
                            <Tag key={permission.id} color="green">
                              {permission.name}
                            </Tag>
                          )) || <span style={{ color: '#999' }}>暂无权限</span>}
                        </Space>
                        <Divider style={{ margin: '12px 0' }} />
                      </div>
                    ))}
                  </div>
                ) : (
                  <div style={{ textAlign: 'center', color: '#999', padding: '20px' }}>
                    该用户暂未分配任何角色
                  </div>
                )}
              </Card>
            </TabPane>
            
            <TabPane 
              tab={
                <span>
                  <HistoryOutlined />
                  活动日志
                </span>
              } 
              key="activities"
            >
              <Card title="最近活动" size="small">
                <Timeline loading={activitiesLoading}>
                  {userActivities.map(activity => (
                    <Timeline.Item
                      key={activity.id}
                      color={
                        activity.status === 'success' ? 'green' : 
                        activity.status === 'failed' ? 'red' : 'orange'
                      }
                    >
                      <div>
                        <div style={{ fontWeight: 500, marginBottom: 4 }}>
                          {activity.description}
                        </div>
                        <div style={{ fontSize: '12px', color: '#999' }}>
                          <div>时间: {new Date(activity.createdAt).toLocaleString('zh-CN')}</div>
                          <div>IP: {activity.ip}</div>
                          <div>状态: 
                            <Badge 
                              status={
                                activity.status === 'success' ? 'success' : 
                                activity.status === 'failed' ? 'error' : 'warning'
                              } 
                              text={
                                activity.status === 'success' ? '成功' : 
                                activity.status === 'failed' ? '失败' : '警告'
                              }
                              style={{ marginLeft: 4 }}
                            />
                          </div>
                        </div>
                      </div>
                    </Timeline.Item>
                  ))}
                </Timeline>
                {userActivities.length === 0 && !activitiesLoading && (
                  <div style={{ textAlign: 'center', color: '#999', padding: '20px' }}>
                    暂无活动记录
                  </div>
                )}
              </Card>
            </TabPane>
          </Tabs>
        )}
      </Drawer>

      {/* 高级搜索Modal */}
      <Modal
        title="高级搜索"
        open={advancedSearchVisible}
        onCancel={() => setAdvancedSearchVisible(false)}
        footer={null}
        width={600}
      >
        <Form
          form={searchForm}
          layout="vertical"
          onFinish={handleAdvancedSearch}
          initialValues={searchFilters}
        >
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item name="status" label="用户状态">
                <Select placeholder="选择状态" allowClear>
                  <Option value="active">正常</Option>
                  <Option value="inactive">禁用</Option>
                </Select>
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="roleId" label="用户角色">
                <Select placeholder="选择角色" allowClear>
                  {roles.map(role => (
                    <Option key={role.id} value={role.id}>
                      {role.name}
                    </Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>
          </Row>
          
          <Form.Item name="dateRange" label="创建时间范围">
            <DatePicker.RangePicker 
              style={{ width: '100%' }}
              placeholder={['开始日期', '结束日期']}
            />
          </Form.Item>

          <Form.Item style={{ marginBottom: 0, textAlign: 'right' }}>
            <Space>
              <Button onClick={() => setAdvancedSearchVisible(false)}>
                取消
              </Button>
              <Button onClick={handleResetSearch}>
                重置
              </Button>
              <Button type="primary" htmlType="submit">
                搜索
              </Button>
            </Space>
          </Form.Item>
        </Form>
       </Modal>

      {/* 批量分配角色Modal */}
      <Modal
        title="批量分配角色"
        open={bulkRoleModalVisible}
        onCancel={() => setBulkRoleModalVisible(false)}
        footer={null}
        width={500}
      >
        <Form
          form={bulkRoleForm}
          layout="vertical"
          onFinish={handleBulkRoleAssign}
        >
          <Form.Item 
            name="roleIds" 
            label="选择角色"
            rules={[{ required: true, message: '请选择至少一个角色' }]}
          >
            <Select
              mode="multiple"
              placeholder="选择要分配的角色"
              style={{ width: '100%' }}
            >
              {roles.map(role => (
                <Option key={role.id} value={role.id}>
                  {role.name}
                </Option>
              ))}
            </Select>
          </Form.Item>

          <div style={{ marginBottom: '16px', padding: '12px', backgroundColor: '#f6f8fa', borderRadius: '6px' }}>
            <div style={{ fontSize: '14px', color: '#666', marginBottom: '8px' }}>
              将为以下 {selectedRowKeys.length} 个用户分配角色：
            </div>
            <div style={{ fontSize: '12px', color: '#999' }}>
              {users
                .filter(user => selectedRowKeys.includes(user.id))
                .map(user => user.username)
                .join('、')}
            </div>
          </div>

          <Form.Item style={{ marginBottom: 0, textAlign: 'right' }}>
            <Space>
              <Button onClick={() => setBulkRoleModalVisible(false)}>
                取消
              </Button>
              <Button type="primary" htmlType="submit">
                确认分配
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default UserManagement;