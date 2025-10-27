import React, { useState, useEffect } from 'react';
import {
  Card,
  Tabs,
  Form,
  Switch,
  InputNumber,
  Select,
  Button,
  Space,
  Divider,
  Row,
  Col,
  Alert,
  Tooltip,
  Badge,
  message,
  Modal,
  Input,
} from 'antd';
import {
  DesktopOutlined,
  GlobalOutlined,
  MobileOutlined,
  ClockCircleOutlined,
  ApiOutlined,
  RobotOutlined,
  SettingOutlined,
  EyeOutlined,
  SaveOutlined,
  ReloadOutlined,
  ExportOutlined,
  ImportOutlined,
  QuestionCircleOutlined,
} from '@ant-design/icons';
import { DeviceType } from '@/services/menu';
import './index.less';

const { TabPane } = Tabs;
const { Option } = Select;
const { TextArea } = Input;

export interface DeviceConfigProps {
  visible: boolean;
  onCancel: () => void;
  onSave: (configs: DeviceConfigData) => void;
  initialConfigs?: DeviceConfigData;
}

export interface DeviceConfigData {
  [DeviceType.DESKTOP]: DeviceSettings;
  [DeviceType.WEB]: DeviceSettings;
  [DeviceType.MOBILE]: DeviceSettings;
  [DeviceType.WATCH]: DeviceSettings;
  [DeviceType.IOT]: DeviceSettings;
  [DeviceType.ROBOT]: DeviceSettings;
}

export interface DeviceSettings {
  enabled: boolean;
  maxMenuDepth: number;
  showIcons: boolean;
  showBadges: boolean;
  collapsible: boolean;
  defaultCollapsed: boolean;
  menuWidth: number;
  itemHeight: number;
  fontSize: number;
  theme: 'light' | 'dark' | 'auto';
  layout: 'vertical' | 'horizontal' | 'mixed';
  animation: boolean;
  customCSS?: string;
  hiddenMenus: string[];
  menuOrder: string[];
  responsiveBreakpoint?: number;
}

const DeviceConfig: React.FC<DeviceConfigProps> = ({
  visible,
  onCancel,
  onSave,
  initialConfigs,
}) => {
  const [form] = Form.useForm();
  const [activeDevice, setActiveDevice] = useState<DeviceType>(DeviceType.DESKTOP);
  const [configs, setConfigs] = useState<DeviceConfigData>({} as DeviceConfigData);
  const [previewMode, setPreviewMode] = useState(false);
  const [hasChanges, setHasChanges] = useState(false);

  // 设备配置信息
  const deviceInfo = {
    [DeviceType.DESKTOP]: {
      name: '桌面端',
      icon: <DesktopOutlined />,
      description: '适用于桌面应用程序，支持完整的菜单功能',
      color: '#1890ff',
      defaultSettings: {
        enabled: true,
        maxMenuDepth: 4,
        showIcons: true,
        showBadges: true,
        collapsible: true,
        defaultCollapsed: false,
        menuWidth: 240,
        itemHeight: 40,
        fontSize: 14,
        theme: 'light' as const,
        layout: 'vertical' as const,
        animation: true,
        hiddenMenus: [],
        menuOrder: [],
      },
    },
    [DeviceType.WEB]: {
      name: 'Web端',
      icon: <GlobalOutlined />,
      description: '适用于Web浏览器，响应式设计',
      color: '#52c41a',
      defaultSettings: {
        enabled: true,
        maxMenuDepth: 3,
        showIcons: true,
        showBadges: true,
        collapsible: true,
        defaultCollapsed: false,
        menuWidth: 220,
        itemHeight: 36,
        fontSize: 14,
        theme: 'auto' as const,
        layout: 'vertical' as const,
        animation: true,
        responsiveBreakpoint: 768,
        hiddenMenus: [],
        menuOrder: [],
      },
    },
    [DeviceType.MOBILE]: {
      name: '移动端',
      icon: <MobileOutlined />,
      description: '适用于手机和平板设备，触摸优化',
      color: '#fa8c16',
      defaultSettings: {
        enabled: true,
        maxMenuDepth: 2,
        showIcons: true,
        showBadges: false,
        collapsible: false,
        defaultCollapsed: true,
        menuWidth: 280,
        itemHeight: 48,
        fontSize: 16,
        theme: 'auto' as const,
        layout: 'vertical' as const,
        animation: true,
        hiddenMenus: [],
        menuOrder: [],
      },
    },
    [DeviceType.WATCH]: {
      name: '手表端',
      icon: <ClockCircleOutlined />,
      description: '适用于智能手表，极简设计',
      color: '#eb2f96',
      defaultSettings: {
        enabled: false,
        maxMenuDepth: 1,
        showIcons: false,
        showBadges: false,
        collapsible: false,
        defaultCollapsed: false,
        menuWidth: 200,
        itemHeight: 32,
        fontSize: 12,
        theme: 'dark' as const,
        layout: 'vertical' as const,
        animation: false,
        hiddenMenus: [],
        menuOrder: [],
      },
    },
    [DeviceType.IOT]: {
      name: '物联网端',
      icon: <ApiOutlined />,
      description: '适用于物联网设备，API优先',
      color: '#722ed1',
      defaultSettings: {
        enabled: false,
        maxMenuDepth: 2,
        showIcons: false,
        showBadges: false,
        collapsible: false,
        defaultCollapsed: false,
        menuWidth: 180,
        itemHeight: 28,
        fontSize: 12,
        theme: 'dark' as const,
        layout: 'horizontal' as const,
        animation: false,
        hiddenMenus: [],
        menuOrder: [],
      },
    },
    [DeviceType.ROBOT]: {
      name: '机器人端',
      icon: <RobotOutlined />,
      description: '适用于机器人系统，语音交互优化',
      color: '#13c2c2',
      defaultSettings: {
        enabled: false,
        maxMenuDepth: 2,
        showIcons: false,
        showBadges: false,
        collapsible: false,
        defaultCollapsed: false,
        menuWidth: 160,
        itemHeight: 24,
        fontSize: 10,
        theme: 'dark' as const,
        layout: 'horizontal' as const,
        animation: false,
        hiddenMenus: [],
        menuOrder: [],
      },
    },
  };

  // 初始化配置
  useEffect(() => {
    if (visible) {
      const defaultConfigs = Object.keys(deviceInfo).reduce((acc, device) => {
        acc[device as DeviceType] = {
          ...deviceInfo[device as DeviceType].defaultSettings,
          ...initialConfigs?.[device as DeviceType],
        };
        return acc;
      }, {} as DeviceConfigData);
      
      setConfigs(defaultConfigs);
      form.setFieldsValue(defaultConfigs[activeDevice]);
      setHasChanges(false);
    }
  }, [visible, initialConfigs, activeDevice]);

  // 处理表单变化
  const handleFormChange = (changedValues: any, allValues: any) => {
    setConfigs(prev => ({
      ...prev,
      [activeDevice]: allValues,
    }));
    setHasChanges(true);
  };

  // 切换设备
  const handleDeviceChange = (device: DeviceType) => {
    setActiveDevice(device);
    form.setFieldsValue(configs[device]);
  };

  // 保存配置
  const handleSave = async () => {
    try {
      await form.validateFields();
      onSave(configs);
      setHasChanges(false);
      message.success('配置保存成功');
    } catch (error) {
      message.error('请检查配置项是否正确');
    }
  };

  // 重置配置
  const handleReset = () => {
    Modal.confirm({
      title: '确认重置',
      content: '是否重置当前设备的配置为默认值？',
      onOk: () => {
        const defaultSettings = deviceInfo[activeDevice].defaultSettings;
        setConfigs(prev => ({
          ...prev,
          [activeDevice]: defaultSettings,
        }));
        form.setFieldsValue(defaultSettings);
        setHasChanges(true);
        message.success('配置已重置');
      },
    });
  };

  // 导出配置
  const handleExport = () => {
    const dataStr = JSON.stringify(configs, null, 2);
    const dataBlob = new Blob([dataStr], { type: 'application/json' });
    const url = URL.createObjectURL(dataBlob);
    const link = document.createElement('a');
    link.href = url;
    link.download = 'device-configs.json';
    link.click();
    URL.revokeObjectURL(url);
    message.success('配置导出成功');
  };

  // 导入配置
  const handleImport = () => {
    const input = document.createElement('input');
    input.type = 'file';
    input.accept = '.json';
    input.onchange = (e) => {
      const file = (e.target as HTMLInputElement).files?.[0];
      if (file) {
        const reader = new FileReader();
        reader.onload = (e) => {
          try {
            const importedConfigs = JSON.parse(e.target?.result as string);
            setConfigs(importedConfigs);
            form.setFieldsValue(importedConfigs[activeDevice]);
            setHasChanges(true);
            message.success('配置导入成功');
          } catch (error) {
            message.error('配置文件格式错误');
          }
        };
        reader.readAsText(file);
      }
    };
    input.click();
  };

  // 渲染设备标签页
  const renderDeviceTabs = () => {
    return Object.entries(deviceInfo).map(([device, info]) => {
      const isEnabled = configs[device as DeviceType]?.enabled;
      return (
        <TabPane
          tab={
            <Space>
              <Badge
                status={isEnabled ? 'success' : 'default'}
                dot
              />
              {info.icon}
              {info.name}
            </Space>
          }
          key={device}
        >
          <div className="device-config-content">
            <Alert
              message={info.description}
              type="info"
              showIcon
              style={{ marginBottom: 16 }}
            />
            
            <Form
              form={form}
              layout="vertical"
              onValuesChange={handleFormChange}
              initialValues={configs[device as DeviceType]}
            >
              <Row gutter={16}>
                <Col span={12}>
                  <Form.Item
                    name="enabled"
                    label={
                      <Space>
                        启用此设备
                        <Tooltip title="是否为此设备类型启用菜单功能">
                          <QuestionCircleOutlined />
                        </Tooltip>
                      </Space>
                    }
                    valuePropName="checked"
                  >
                    <Switch />
                  </Form.Item>
                </Col>
                <Col span={12}>
                  <Form.Item
                    name="maxMenuDepth"
                    label="最大菜单层级"
                    rules={[{ required: true, message: '请输入最大菜单层级' }]}
                  >
                    <InputNumber min={1} max={5} style={{ width: '100%' }} />
                  </Form.Item>
                </Col>
              </Row>

              <Row gutter={16}>
                <Col span={8}>
                  <Form.Item
                    name="showIcons"
                    label="显示图标"
                    valuePropName="checked"
                  >
                    <Switch />
                  </Form.Item>
                </Col>
                <Col span={8}>
                  <Form.Item
                    name="showBadges"
                    label="显示徽章"
                    valuePropName="checked"
                  >
                    <Switch />
                  </Form.Item>
                </Col>
                <Col span={8}>
                  <Form.Item
                    name="animation"
                    label="启用动画"
                    valuePropName="checked"
                  >
                    <Switch />
                  </Form.Item>
                </Col>
              </Row>

              <Row gutter={16}>
                <Col span={8}>
                  <Form.Item
                    name="collapsible"
                    label="可折叠"
                    valuePropName="checked"
                  >
                    <Switch />
                  </Form.Item>
                </Col>
                <Col span={8}>
                  <Form.Item
                    name="defaultCollapsed"
                    label="默认折叠"
                    valuePropName="checked"
                  >
                    <Switch />
                  </Form.Item>
                </Col>
                <Col span={8}>
                  <Form.Item
                    name="theme"
                    label="主题"
                    rules={[{ required: true, message: '请选择主题' }]}
                  >
                    <Select>
                      <Option value="light">浅色</Option>
                      <Option value="dark">深色</Option>
                      <Option value="auto">自动</Option>
                    </Select>
                  </Form.Item>
                </Col>
              </Row>

              <Row gutter={16}>
                <Col span={8}>
                  <Form.Item
                    name="layout"
                    label="布局方式"
                    rules={[{ required: true, message: '请选择布局方式' }]}
                  >
                    <Select>
                      <Option value="vertical">垂直</Option>
                      <Option value="horizontal">水平</Option>
                      <Option value="mixed">混合</Option>
                    </Select>
                  </Form.Item>
                </Col>
                <Col span={8}>
                  <Form.Item
                    name="menuWidth"
                    label="菜单宽度(px)"
                    rules={[{ required: true, message: '请输入菜单宽度' }]}
                  >
                    <InputNumber min={120} max={400} style={{ width: '100%' }} />
                  </Form.Item>
                </Col>
                <Col span={8}>
                  <Form.Item
                    name="itemHeight"
                    label="菜单项高度(px)"
                    rules={[{ required: true, message: '请输入菜单项高度' }]}
                  >
                    <InputNumber min={20} max={60} style={{ width: '100%' }} />
                  </Form.Item>
                </Col>
              </Row>

              <Row gutter={16}>
                <Col span={12}>
                  <Form.Item
                    name="fontSize"
                    label="字体大小(px)"
                    rules={[{ required: true, message: '请输入字体大小' }]}
                  >
                    <InputNumber min={10} max={20} style={{ width: '100%' }} />
                  </Form.Item>
                </Col>
                {device === DeviceType.WEB && (
                  <Col span={12}>
                    <Form.Item
                      name="responsiveBreakpoint"
                      label="响应式断点(px)"
                    >
                      <InputNumber min={320} max={1920} style={{ width: '100%' }} />
                    </Form.Item>
                  </Col>
                )}
              </Row>

              <Form.Item
                name="customCSS"
                label="自定义CSS"
              >
                <TextArea
                  rows={4}
                  placeholder="输入自定义CSS样式..."
                />
              </Form.Item>
            </Form>
          </div>
        </TabPane>
      );
    });
  };

  return (
    <Modal
      title="多终端适配配置"
      open={visible}
      onCancel={onCancel}
      width={1000}
      footer={
        <Space>
          <Button onClick={handleImport} icon={<ImportOutlined />}>
            导入配置
          </Button>
          <Button onClick={handleExport} icon={<ExportOutlined />}>
            导出配置
          </Button>
          <Button onClick={handleReset} icon={<ReloadOutlined />}>
            重置当前
          </Button>
          <Button
            type="primary"
            onClick={handleSave}
            icon={<SaveOutlined />}
            disabled={!hasChanges}
          >
            保存配置
          </Button>
        </Space>
      }
      className="device-config-modal"
    >
      <div className="device-config">
        <Tabs
          activeKey={activeDevice}
          onChange={handleDeviceChange}
          tabPosition="left"
          className="device-tabs"
        >
          {renderDeviceTabs()}
        </Tabs>
      </div>
    </Modal>
  );
};

export default DeviceConfig;