import React, { useState, useEffect, useMemo } from 'react';
import {
  Card,
  Table,
  Button,
  Space,
  Input,
  Tag,
  message,
  Modal,
  Form,
  Select,
  Popconfirm,
  Drawer,
  Row,
  Col,
  Divider,
  Statistic,
  Badge,
  Tabs,
  Upload,
  Progress,
  Tooltip,
  Dropdown
} from 'antd';
import type { ColumnsType } from 'antd/es/table';
import {
  PlusOutlined,
  SearchOutlined,
  EditOutlined,
  DeleteOutlined,
  EyeOutlined,
  DownloadOutlined,
  UploadOutlined,
  SyncOutlined,
  MoreOutlined,
  FilterOutlined,
  ReloadOutlined,
  SettingOutlined
} from '@ant-design/icons';
import { Permission, QueryParams } from '@/types';
import { permissionService, PermissionTemplate, PermissionStats } from '@/services/permission';

const { Option } = Select;
const { TabPane } = Tabs;
const { TextArea } = Input;

// 权限表单接口
interface PermissionFormData {
  name: string;
  resource: string;
  action: string;
  description?: string;
}

// 权限模板表单接口
interface TemplateFormData {
  name: string;
  description: string;
  permissionIds: string[];
}

const PermissionManagement: React.FC = () => {
  // 基础状态
  const [permissions, setPermissions] = useState<Permission[]>([]);
  const [loading, setLoading] = useState(false);
  const [selectedRowKeys, setSelectedRowKeys] = useState<string[]>([]);
  const [searchText, setSearchText] = useState('');
  const [resourceFilter, setResourceFilter] = useState<string>('');
  const [actionFilter, setActionFilter] = useState<string>('');
  const [systemFilter, setSystemFilter] = useState<string>('');

  // 分页状态
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 10,
    total: 0,
  });

  // 模态框状态
  const [modalVisible, setModalVisible] = useState(false);
  const [editingPermission, setEditingPermission] = useState<Permission | null>(null);
  const [detailDrawerVisible, setDetailDrawerVisible] = useState(false);
  const [selectedPermission, setSelectedPermission] = useState<Permission | null>(null);

  // 权限模板状态
  const [templateModalVisible, setTemplateModalVisible] = useState(false);
  const [templates, setTemplates] = useState<PermissionTemplate[]>([]);
  const [editingTemplate, setEditingTemplate] = useState<PermissionTemplate | null>(null);

  // 统计信息状态
  const [stats, setStats] = useState<PermissionStats | null>(null);
  const [statsLoading, setStatsLoading] = useState(false);

  // 资源和操作选项
  const [resources, setResources] = useState<string[]>([]);
  const [actions, setActions] = useState<string[]>([]);

  // 表单实例
  const [form] = Form.useForm();
  const [templateForm] = Form.useForm();

  // 加载权限列表
  const loadPermissions = async () => {
    setLoading(true);
    try {
      const params: QueryParams = {
        page: pagination.current,
        pageSize: pagination.pageSize,
        search: searchText || undefined,
        resource: resourceFilter || undefined,
        action: actionFilter || undefined,
        isSystem: systemFilter ? systemFilter === 'system' : undefined,
      };
      
      const response = await permissionService.getPermissions(params);
      // 确保 response.data 是数组
      setPermissions(Array.isArray(response.data) ? response.data : []);
      setPagination(prev => ({
        ...prev,
        total: response.total || 0
      }));
    } catch (error) {
      message.error('加载权限列表失败');
      console.error('Load permissions error:', error);
      // 设置空数组作为默认值
      setPermissions([]);
      setPagination(prev => ({
        ...prev,
        total: 0
      }));
    } finally {
      setLoading(false);
    }
  };

  // 加载统计信息
  const loadStats = async () => {
    setStatsLoading(true);
    try {
      const statsData = await permissionService.getPermissionStats();
      setStats(statsData);
    } catch (error) {
      message.error('加载统计信息失败');
      console.error('Load stats error:', error);
      // 设置默认统计信息
      setStats({
        totalPermissions: 0,
        systemPermissions: 0,
        customPermissions: 0,
        byResource: {},
        byAction: {},
        recentlyCreated: [],
        mostUsed: []
      });
    } finally {
      setStatsLoading(false);
    }
  };

  // 加载权限模板
  const loadTemplates = async () => {
    try {
      const templatesData = await permissionService.getPermissionTemplates();
      const normalized = (Array.isArray(templatesData) ? templatesData : (templatesData?.items ?? []))
        .map((t: any) => {
          let ids: string[] = Array.isArray(t?.permissionIds) ? t.permissionIds : [];
          const raw = t?.template_data ?? t?.templateData;
          if (ids.length === 0 && raw) {
            try {
              const parsed = typeof raw === 'string' ? JSON.parse(raw) : raw;
              if (Array.isArray(parsed?.permissionIds)) {
                ids = parsed.permissionIds;
              } else if (Array.isArray(parsed?.permissions)) {
                ids = parsed.permissions;
              }
            } catch {
              // ignore parse errors, fallback to empty array
            }
          }
          return { ...t, permissionIds: ids };
        });
      setTemplates(normalized as PermissionTemplate[]);
    } catch (error) {
      message.error('加载权限模板失败');
      console.error('Load templates error:', error);
      // 设置空数组作为默认值
      setTemplates([]);
    }
  };

  // 加载资源和操作选项
  const loadOptions = async () => {
    try {
      const [resourcesData, actionsData] = await Promise.all([
        permissionService.getResources(),
        permissionService.getActions()
      ]);
      // 确保返回的数据是数组
      setResources(Array.isArray(resourcesData) ? resourcesData : []);
      setActions(Array.isArray(actionsData) ? actionsData : []);
    } catch (error) {
      message.error('加载选项失败');
      console.error('Load options error:', error);
      // 设置空数组作为默认值
      setResources([]);
      setActions([]);
    }
  };

  useEffect(() => {
    loadPermissions();
  }, [pagination.current, pagination.pageSize, searchText, resourceFilter, actionFilter, systemFilter]);

  useEffect(() => {
    loadStats();
    loadTemplates();
    loadOptions();
  }, []);

  // 处理搜索
  const handleSearch = (value: string) => {
    setSearchText(value);
    setPagination(prev => ({ ...prev, current: 1 }));
  };

  // 处理创建权限
  const handleCreate = () => {
    setEditingPermission(null);
    form.resetFields();
    setModalVisible(true);
  };

  // 处理编辑权限
  const handleEdit = (permission: Permission) => {
    setEditingPermission(permission);
    form.setFieldsValue(permission);
    setModalVisible(true);
  };

  // 处理查看详情
  const handleViewDetail = (permission: Permission) => {
    setSelectedPermission(permission);
    setDetailDrawerVisible(true);
  };

  // 处理删除权限
  const handleDelete = async (permission: Permission) => {
    try {
      await permissionService.deletePermission(permission.id);
      message.success('权限删除成功');
      loadPermissions();
      loadStats();
    } catch (error) {
      message.error('权限删除失败');
      console.error('Delete permission error:', error);
    }
  };

  // 处理批量删除
  const handleBatchDelete = async () => {
    if (selectedRowKeys.length === 0) {
      message.warning('请选择要删除的权限');
      return;
    }

    Modal.confirm({
      title: '确认删除',
      content: `确定要删除选中的 ${selectedRowKeys.length} 个权限吗？`,
      onOk: async () => {
        try {
          await permissionService.batchDeletePermissions(selectedRowKeys);
          setSelectedRowKeys([]);
          message.success('批量删除成功');
          loadPermissions();
          loadStats();
        } catch (error) {
          message.error('批量删除失败');
          console.error('Batch delete error:', error);
        }
      }
    });
  };

  // 处理表单提交
  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();

      if (editingPermission) {
        // 更新权限
        await permissionService.updatePermission(editingPermission.id, values);
        message.success('权限更新成功');
      } else {
        // 创建权限
        await permissionService.createPermission(values);
        message.success('权限创建成功');
      }

      setModalVisible(false);
      form.resetFields();
      loadPermissions();
      loadStats();
    } catch (error) {
      message.error('操作失败');
      console.error('Submit error:', error);
    }
  };

  // 处理权限模板创建
  const handleCreateTemplate = () => {
    setEditingTemplate(null);
    templateForm.resetFields();
    setTemplateModalVisible(true);
  };

  // 处理权限模板编辑
  const handleEditTemplate = (template: PermissionTemplate) => {
    setEditingTemplate(template);
    templateForm.setFieldsValue(template);
    setTemplateModalVisible(true);
  };

  // 处理权限模板删除
  const handleDeleteTemplate = async (template: PermissionTemplate) => {
    try {
      await permissionService.deletePermissionTemplate(template.id);
      message.success('权限模板删除成功');
      loadTemplates();
    } catch (error) {
      message.error('权限模板删除失败');
      console.error('Delete template error:', error);
    }
  };

  // 处理权限模板提交
  const handleTemplateSubmit = async () => {
    try {
      const values = await templateForm.validateFields();

      if (editingTemplate) {
        // 更新模板
        await permissionService.updatePermissionTemplate(editingTemplate.id, values);
        message.success('权限模板更新成功');
      } else {
        // 创建模板
        await permissionService.createPermissionTemplate(values);
        message.success('权限模板创建成功');
      }

      setTemplateModalVisible(false);
      templateForm.resetFields();
      loadTemplates();
    } catch (error) {
      message.error('操作失败');
      console.error('Template submit error:', error);
    }
  };

  // 处理同步系统权限
  const handleSyncSystemPermissions = async () => {
    try {
      setLoading(true);
      await permissionService.syncSystemPermissions();
      message.success('系统权限同步成功');
      loadPermissions();
      loadStats();
    } catch (error) {
      message.error('系统权限同步失败');
      console.error('Sync permissions error:', error);
    } finally {
      setLoading(false);
    }
  };

  // 表格列配置
  const columns: ColumnsType<Permission> = [
    {
      title: '权限名称',
      dataIndex: 'name',
      key: 'name',
      width: 200,
      ellipsis: true,
      render: (text, record) => (
        <Space>
          <span>{text}</span>
          {record.isSystem && <Tag color="blue" size="small">系统</Tag>}
        </Space>
      )
    },
    {
      title: '资源',
      dataIndex: 'resource',
      key: 'resource',
      width: 120,
      render: (text) => <Tag color="geekblue">{text}</Tag>
    },
    {
      title: '操作',
      dataIndex: 'action',
      key: 'action',
      width: 120,
      render: (text) => <Tag color="green">{text}</Tag>
    },
    {
      title: '描述',
      dataIndex: 'description',
      key: 'description',
      ellipsis: true,
      render: (text) => text || '-'
    },
    {
      title: '操作',
      key: 'actions',
      width: 120,
      fixed: 'right',
      render: (_, record) => (
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
            menu={{
              items: [
                {
                  key: 'edit',
                  label: '编辑',
                  icon: <EditOutlined />,
                  onClick: () => handleEdit(record),
                  disabled: record.isSystem
                },
                {
                  key: 'delete',
                  label: '删除',
                  icon: <DeleteOutlined />,
                  danger: true,
                  disabled: record.isSystem,
                  onClick: () => {
                    Modal.confirm({
                      title: '确认删除',
                      content: `确定要删除权限 "${record.name}" 吗？`,
                      onOk: () => handleDelete(record)
                    });
                  }
                }
              ]
            }}
            trigger={['click']}
          >
            <Button type="link" size="small" icon={<MoreOutlined />} />
          </Dropdown>
        </Space>
      )
    }
  ];

  // 行选择配置
  const rowSelection = {
    selectedRowKeys,
    onChange: (keys: React.Key[]) => {
      setSelectedRowKeys(keys as string[]);
    },
    getCheckboxProps: (record: Permission) => ({
      disabled: record.isSystem
    })
  };

  // 过滤后的权限列表
  const filteredPermissions = useMemo(() => {
    return permissions;
  }, [permissions]);

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
                  onClick={handleCreate}
                >
                  新建权限
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
                  icon={<SyncOutlined />}
                  onClick={handleSyncSystemPermissions}
                  loading={loading}
                >
                  同步系统权限
                </Button>
                <Button
                  icon={<SettingOutlined />}
                  onClick={handleCreateTemplate}
                >
                  权限模板
                </Button>
              </Space>
            </Col>
            <Col>
              <Space>
                <Input.Search
                  placeholder="搜索权限名称、资源或操作"
                  allowClear
                  style={{ width: 250 }}
                  onSearch={handleSearch}
                  onChange={(e) => !e.target.value && handleSearch('')}
                />
                <Select
                  placeholder="资源"
                  allowClear
                  style={{ width: 120 }}
                  value={resourceFilter}
                  onChange={setResourceFilter}
                >
                  {(resources || []).map(resource => (
                    <Option key={resource} value={resource}>{resource}</Option>
                  ))}
                </Select>
                <Select
                  placeholder="操作"
                  allowClear
                  style={{ width: 120 }}
                  value={actionFilter}
                  onChange={setActionFilter}
                >
                  {(actions || []).map(action => (
                    <Option key={action} value={action}>{action}</Option>
                  ))}
                </Select>
                <Select
                  placeholder="类型"
                  allowClear
                  style={{ width: 120 }}
                  value={systemFilter}
                  onChange={setSystemFilter}
                >
                  <Option value="system">系统权限</Option>
                  <Option value="custom">自定义权限</Option>
                </Select>
                <Button
                  icon={<ReloadOutlined />}
                  onClick={() => {
                    setSearchText('');
                    setResourceFilter('');
                    setActionFilter('');
                    setSystemFilter('');
                    loadPermissions();
                  }}
                >
                  重置
                </Button>
              </Space>
            </Col>
          </Row>
        </div>

        {/* 统计信息 */}
        {stats && (
          <Row gutter={16} style={{ marginBottom: 16 }}>
            <Col span={6}>
              <Statistic
                title="总权限数"
                value={stats.totalPermissions}
                suffix="个"
                valueStyle={{ color: '#1890ff' }}
              />
            </Col>
            <Col span={6}>
              <Statistic
                title="系统权限"
                value={stats.systemPermissions}
                suffix="个"
                valueStyle={{ color: '#52c41a' }}
              />
            </Col>
            <Col span={6}>
              <Statistic
                title="自定义权限"
                value={stats.customPermissions}
                suffix="个"
                valueStyle={{ color: '#faad14' }}
              />
            </Col>
            <Col span={6}>
              <Statistic
                title="已选择"
                value={selectedRowKeys.length}
                suffix="个"
                valueStyle={{ color: '#f5222d' }}
              />
            </Col>
          </Row>
        )}

        {/* 权限表格 */}
        <Table
          columns={columns}
          dataSource={filteredPermissions}
          rowKey="id"
          loading={loading}
          rowSelection={rowSelection}
          pagination={{
            ...pagination,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total, range) => `第 ${range[0]}-${range[1]} 条，共 ${total} 条`,
          }}
          scroll={{ x: 1000 }}
        />
      </Card>

      {/* 权限编辑模态框 */}
      <Modal
        title={editingPermission ? '编辑权限' : '新建权限'}
        open={modalVisible}
        onOk={handleSubmit}
        onCancel={() => setModalVisible(false)}
        destroyOnClose
      >
        <Form
          form={form}
          layout="vertical"
          initialValues={{ isSystem: false }}
        >
          <Form.Item
            name="name"
            label="权限名称"
            rules={[{ required: true, message: '请输入权限名称' }]}
          >
            <Input placeholder="请输入权限名称" />
          </Form.Item>
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                name="resource"
                label="资源"
                rules={[{ required: true, message: '请选择或输入资源' }]}
              >
                <Select
                  placeholder="请选择或输入资源"
                  showSearch
                  allowClear
                  mode="combobox"
                >
                  {(resources || []).map(resource => (
                    <Option key={resource} value={resource}>{resource}</Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="action"
                label="操作"
                rules={[{ required: true, message: '请选择或输入操作' }]}
              >
                <Select
                  placeholder="请选择或输入操作"
                  showSearch
                  allowClear
                  mode="combobox"
                >
                  {(actions || []).map(action => (
                    <Option key={action} value={action}>{action}</Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>
          </Row>
          <Form.Item
            name="description"
            label="描述"
          >
            <TextArea
              placeholder="请输入权限描述"
              rows={3}
            />
          </Form.Item>
        </Form>
      </Modal>

      {/* 权限详情抽屉 */}
      <Drawer
        title="权限详情"
        placement="right"
        width={600}
        open={detailDrawerVisible}
        onClose={() => setDetailDrawerVisible(false)}
        destroyOnClose
      >
        {selectedPermission && (
          <div>
            <Row gutter={16}>
              <Col span={12}>
                <Statistic title="权限名称" value={selectedPermission.name} />
              </Col>
              <Col span={12}>
                <Statistic 
                  title="权限类型" 
                  value={selectedPermission.isSystem ? '系统权限' : '自定义权限'}
                  valueStyle={{ 
                    color: selectedPermission.isSystem ? '#52c41a' : '#faad14' 
                  }}
                />
              </Col>
            </Row>
            <Divider />
            <Row gutter={16}>
              <Col span={12}>
                <Statistic title="资源" value={selectedPermission.resource} />
              </Col>
              <Col span={12}>
                <Statistic title="操作" value={selectedPermission.action} />
              </Col>
            </Row>
            <Divider />
            <div>
              <h4>描述</h4>
              <p>{selectedPermission.description || '暂无描述'}</p>
            </div>
          </div>
        )}
      </Drawer>

      {/* 权限模板管理模态框 */}
      <Modal
        title="权限模板管理"
        open={templateModalVisible}
        onCancel={() => setTemplateModalVisible(false)}
        footer={null}
        width={800}
        destroyOnClose
      >
        <Tabs defaultActiveKey="list">
          <TabPane tab="模板列表" key="list">
            <div style={{ marginBottom: 16 }}>
              <Button
                type="primary"
                icon={<PlusOutlined />}
                onClick={() => {
                  setEditingTemplate(null);
                  templateForm.resetFields();
                  // 这里可以打开创建模板的表单
                }}
              >
                新建模板
              </Button>
            </div>
            <Table
              dataSource={templates || []}
              rowKey="id"
              pagination={false}
              size="small"
              columns={[
                {
                  title: '模板名称',
                  dataIndex: 'name',
                  key: 'name',
                  render: (text, record) => (
                    <Space>
                      <span>{text}</span>
                      {record.isSystem && <Tag color="blue" size="small">系统</Tag>}
                    </Space>
                  )
                },
                {
                  title: '描述',
                  dataIndex: 'description',
                  key: 'description',
                  ellipsis: true
                },
                {
                  title: '权限数量',
                  dataIndex: 'permissionIds',
                  key: 'permissionCount',
                  render: (permissionIds) => `${Array.isArray(permissionIds) ? permissionIds.length : 0} 个`
                },
                {
                  title: '操作',
                  key: 'actions',
                  width: 150,
                  render: (_, record) => (
                    <Space size="small">
                      <Button
                        type="link"
                        size="small"
                        onClick={() => handleEditTemplate(record)}
                        disabled={record.isSystem}
                      >
                        编辑
                      </Button>
                      <Popconfirm
                        title="确定要删除这个权限模板吗？"
                        onConfirm={() => handleDeleteTemplate(record)}
                        disabled={record.isSystem}
                      >
                        <Button
                          type="link"
                          size="small"
                          danger
                          disabled={record.isSystem}
                        >
                          删除
                        </Button>
                      </Popconfirm>
                    </Space>
                  )
                }
              ]}
            />
          </TabPane>
          <TabPane tab="创建模板" key="create">
            <Form
              form={templateForm}
              layout="vertical"
              onFinish={handleTemplateSubmit}
            >
              <Form.Item
                name="name"
                label="模板名称"
                rules={[{ required: true, message: '请输入模板名称' }]}
              >
                <Input placeholder="请输入模板名称" />
              </Form.Item>
              <Form.Item
                name="description"
                label="模板描述"
                rules={[{ required: true, message: '请输入模板描述' }]}
              >
                <TextArea placeholder="请输入模板描述" rows={3} />
              </Form.Item>
              <Form.Item
                name="permissionIds"
                label="选择权限"
                rules={[{ required: true, message: '请选择权限' }]}
              >
                <Select
                  mode="multiple"
                  placeholder="请选择权限"
                  optionFilterProp="children"
                  showSearch
                  filterOption={(input, option) =>
                    (option?.children as string)?.toLowerCase().includes(input.toLowerCase())
                  }
                >
                  {(permissions || []).map(permission => (
                    <Option key={permission.id} value={permission.id}>
                      {permission.name} ({permission.resource}:{permission.action})
                    </Option>
                  ))}
                </Select>
              </Form.Item>
              <Form.Item>
                <Space>
                  <Button type="primary" htmlType="submit">
                    {editingTemplate ? '更新模板' : '创建模板'}
                  </Button>
                  <Button onClick={() => templateForm.resetFields()}>
                    重置
                  </Button>
                </Space>
              </Form.Item>
            </Form>
          </TabPane>
        </Tabs>
      </Modal>
    </div>
  );
};

export default PermissionManagement;