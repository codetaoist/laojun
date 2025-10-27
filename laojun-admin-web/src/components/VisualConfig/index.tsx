import React, { useState, useEffect } from 'react';
import {
  Modal,
  Tabs,
  Card,
  Form,
  Input,
  Select,
  Switch,
  InputNumber,
  ColorPicker,
  Slider,
  Button,
  Space,
  Row,
  Col,
  Divider,
  Typography,
  Alert,
  Tooltip,
  Upload,
  message,
} from 'antd';
import {
  SettingOutlined,
  EyeOutlined,
  SaveOutlined,
  ReloadOutlined,
  UploadOutlined,
  DownloadOutlined,
  BgColorsOutlined,
  FontSizeOutlined,
  BorderOutlined,
  RadiusSettingOutlined,
} from '@ant-design/icons';
import { menuConfigService } from '@/services/menu';
import './index.less';

const { Title, Text } = Typography;
const { TextArea } = Input;

export interface VisualConfigData {
  theme: {
    primaryColor: string;
    backgroundColor: string;
    textColor: string;
    borderColor: string;
    hoverColor: string;
    activeColor: string;
  };
  layout: {
    menuWidth: number;
    itemHeight: number;
    fontSize: number;
    iconSize: number;
    padding: number;
    margin: number;
    borderRadius: number;
    borderWidth: number;
  };
  animation: {
    enabled: boolean;
    duration: number;
    easing: string;
    hoverEffect: boolean;
    expandAnimation: boolean;
  };
  typography: {
    fontFamily: string;
    fontWeight: number;
    lineHeight: number;
    letterSpacing: number;
  };
  effects: {
    shadow: boolean;
    shadowColor: string;
    shadowBlur: number;
    gradient: boolean;
    gradientStart: string;
    gradientEnd: string;
    backdrop: boolean;
    backdropBlur: number;
  };
  responsive: {
    breakpoints: {
      mobile: number;
      tablet: number;
      desktop: number;
    };
    mobileLayout: string;
    tabletLayout: string;
    desktopLayout: string;
  };
  customCSS: string;
}

interface VisualConfigProps {
  visible: boolean;
  onSave: (config: VisualConfigData) => void;
  onCancel: () => void;
}

const defaultConfig: VisualConfigData = {
  theme: {
    primaryColor: '#1890ff',
    backgroundColor: '#ffffff',
    textColor: '#000000',
    borderColor: '#d9d9d9',
    hoverColor: '#f5f5f5',
    activeColor: '#e6f7ff',
  },
  layout: {
    menuWidth: 256,
    itemHeight: 40,
    fontSize: 14,
    iconSize: 16,
    padding: 12,
    margin: 4,
    borderRadius: 6,
    borderWidth: 1,
  },
  animation: {
    enabled: true,
    duration: 300,
    easing: 'ease-in-out',
    hoverEffect: true,
    expandAnimation: true,
  },
  typography: {
    fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto',
    fontWeight: 400,
    lineHeight: 1.5,
    letterSpacing: 0,
  },
  effects: {
    shadow: true,
    shadowColor: 'rgba(0, 0, 0, 0.1)',
    shadowBlur: 4,
    gradient: false,
    gradientStart: '#ffffff',
    gradientEnd: '#f0f0f0',
    backdrop: false,
    backdropBlur: 10,
  },
  responsive: {
    breakpoints: {
      mobile: 768,
      tablet: 1024,
      desktop: 1200,
    },
    mobileLayout: 'drawer',
    tabletLayout: 'collapsed',
    desktopLayout: 'expanded',
  },
  customCSS: '',
};

const VisualConfig: React.FC<VisualConfigProps> = ({
  visible,
  onSave,
  onCancel,
}) => {
  const [form] = Form.useForm();
  const [config, setConfig] = useState<VisualConfigData>(defaultConfig);
  const [loading, setLoading] = useState(false);
  const [previewMode, setPreviewMode] = useState(false);

  useEffect(() => {
    if (visible) {
      loadConfig();
    }
  }, [visible]);

  const loadConfig = async () => {
    try {
      setLoading(true);
      const savedConfig = await menuConfigService.getConfig('visual');
      if (savedConfig) {
        setConfig({ ...defaultConfig, ...savedConfig });
        form.setFieldsValue({ ...defaultConfig, ...savedConfig });
      }
    } catch (error) {
      message.error('加载配置失败');
    } finally {
      setLoading(false);
    }
  };

  const handleSave = async () => {
    try {
      const values = await form.validateFields();
      const newConfig = { ...config, ...values };
      setConfig(newConfig);
      onSave(newConfig);
      message.success('配置保存成功');
    } catch (error) {
      message.error('配置保存失败');
    }
  };

  const handleReset = () => {
    setConfig(defaultConfig);
    form.setFieldsValue(defaultConfig);
    message.success('配置已重置');
  };

  const handleExport = () => {
    const blob = new Blob([JSON.stringify(config, null, 2)], {
      type: 'application/json',
    });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `visual-config-${new Date().toISOString().split('T')[0]}.json`;
    a.click();
    URL.revokeObjectURL(url);
    message.success('配置导出成功');
  };

  const handleImport = (file: File) => {
    const reader = new FileReader();
    reader.onload = (e) => {
      try {
        const importedConfig = JSON.parse(e.target?.result as string);
        setConfig({ ...defaultConfig, ...importedConfig });
        form.setFieldsValue({ ...defaultConfig, ...importedConfig });
        message.success('配置导入成功');
      } catch (error) {
        message.error('配置文件格式错误');
      }
    };
    reader.readAsText(file);
    return false;
  };

  const generatePreviewCSS = () => {
    return `
      .menu-preview {
        width: ${config.layout.menuWidth}px;
        background: ${config.effects.gradient 
          ? `linear-gradient(to bottom, ${config.effects.gradientStart}, ${config.effects.gradientEnd})`
          : config.theme.backgroundColor
        };
        color: ${config.theme.textColor};
        font-family: ${config.typography.fontFamily};
        font-size: ${config.layout.fontSize}px;
        font-weight: ${config.typography.fontWeight};
        line-height: ${config.typography.lineHeight};
        letter-spacing: ${config.typography.letterSpacing}px;
        border: ${config.layout.borderWidth}px solid ${config.theme.borderColor};
        border-radius: ${config.layout.borderRadius}px;
        ${config.effects.shadow ? `box-shadow: 0 ${config.effects.shadowBlur}px ${config.effects.shadowBlur * 2}px ${config.effects.shadowColor};` : ''}
        ${config.effects.backdrop ? `backdrop-filter: blur(${config.effects.backdropBlur}px);` : ''}
      }
      
      .menu-preview-item {
        height: ${config.layout.itemHeight}px;
        padding: ${config.layout.padding}px;
        margin: ${config.layout.margin}px;
        border-radius: ${config.layout.borderRadius}px;
        display: flex;
        align-items: center;
        transition: ${config.animation.enabled ? `all ${config.animation.duration}ms ${config.animation.easing}` : 'none'};
      }
      
      .menu-preview-item:hover {
        background: ${config.theme.hoverColor};
        ${config.animation.hoverEffect ? 'transform: translateX(4px);' : ''}
      }
      
      .menu-preview-item.active {
        background: ${config.theme.activeColor};
        color: ${config.theme.primaryColor};
      }
      
      .menu-preview-icon {
        width: ${config.layout.iconSize}px;
        height: ${config.layout.iconSize}px;
        margin-right: 8px;
      }
      
      ${config.customCSS}
    `;
  };

  return (
    <Modal
      title={
        <Space>
          <SettingOutlined />
          可视化配置
        </Space>
      }
      open={visible}
      onCancel={onCancel}
      width={1200}
      footer={
        <Space>
          <Button onClick={onCancel}>取消</Button>
          <Button onClick={handleReset}>重置</Button>
          <Button icon={<DownloadOutlined />} onClick={handleExport}>
            导出
          </Button>
          <Upload
            accept=".json"
            showUploadList={false}
            beforeUpload={handleImport}
          >
            <Button icon={<UploadOutlined />}>导入</Button>
          </Upload>
          <Button
            type={previewMode ? 'default' : 'primary'}
            icon={<EyeOutlined />}
            onClick={() => setPreviewMode(!previewMode)}
          >
            {previewMode ? '退出预览' : '预览'}
          </Button>
          <Button type="primary" icon={<SaveOutlined />} onClick={handleSave}>
            保存
          </Button>
        </Space>
      }
      className="visual-config-modal"
    >
      <Row gutter={24}>
        <Col span={previewMode ? 12 : 24}>
          <Form form={form} layout="vertical" initialValues={config}>
            <Tabs
              items={[
                {
                  key: 'theme',
                  label: (
                    <span>
                      <BgColorsOutlined />
                      主题配色
                    </span>
                  ),
                  children: (
                    <Card size="small">
                      <Row gutter={16}>
                        <Col span={12}>
                          <Form.Item name={['theme', 'primaryColor']} label="主色调">
                            <ColorPicker showText />
                          </Form.Item>
                        </Col>
                        <Col span={12}>
                          <Form.Item name={['theme', 'backgroundColor']} label="背景色">
                            <ColorPicker showText />
                          </Form.Item>
                        </Col>
                        <Col span={12}>
                          <Form.Item name={['theme', 'textColor']} label="文字颜色">
                            <ColorPicker showText />
                          </Form.Item>
                        </Col>
                        <Col span={12}>
                          <Form.Item name={['theme', 'borderColor']} label="边框颜色">
                            <ColorPicker showText />
                          </Form.Item>
                        </Col>
                        <Col span={12}>
                          <Form.Item name={['theme', 'hoverColor']} label="悬停颜色">
                            <ColorPicker showText />
                          </Form.Item>
                        </Col>
                        <Col span={12}>
                          <Form.Item name={['theme', 'activeColor']} label="激活颜色">
                            <ColorPicker showText />
                          </Form.Item>
                        </Col>
                      </Row>
                    </Card>
                  ),
                },
                {
                  key: 'layout',
                  label: (
                    <span>
                      <BorderOutlined />
                      布局尺寸
                    </span>
                  ),
                  children: (
                    <Card size="small">
                      <Row gutter={16}>
                        <Col span={12}>
                          <Form.Item name={['layout', 'menuWidth']} label="菜单宽度">
                            <Slider min={200} max={400} />
                          </Form.Item>
                        </Col>
                        <Col span={12}>
                          <Form.Item name={['layout', 'itemHeight']} label="项目高度">
                            <Slider min={32} max={64} />
                          </Form.Item>
                        </Col>
                        <Col span={12}>
                          <Form.Item name={['layout', 'fontSize']} label="字体大小">
                            <Slider min={12} max={20} />
                          </Form.Item>
                        </Col>
                        <Col span={12}>
                          <Form.Item name={['layout', 'iconSize']} label="图标大小">
                            <Slider min={12} max={24} />
                          </Form.Item>
                        </Col>
                        <Col span={12}>
                          <Form.Item name={['layout', 'padding']} label="内边距">
                            <Slider min={4} max={24} />
                          </Form.Item>
                        </Col>
                        <Col span={12}>
                          <Form.Item name={['layout', 'margin']} label="外边距">
                            <Slider min={0} max={16} />
                          </Form.Item>
                        </Col>
                        <Col span={12}>
                          <Form.Item name={['layout', 'borderRadius']} label="圆角半径">
                            <Slider min={0} max={16} />
                          </Form.Item>
                        </Col>
                        <Col span={12}>
                          <Form.Item name={['layout', 'borderWidth']} label="边框宽度">
                            <Slider min={0} max={4} />
                          </Form.Item>
                        </Col>
                      </Row>
                    </Card>
                  ),
                },
                {
                  key: 'animation',
                  label: (
                    <span>
                      <RadiusSettingOutlined />
                      动画效果
                    </span>
                  ),
                  children: (
                    <Card size="small">
                      <Row gutter={16}>
                        <Col span={12}>
                          <Form.Item name={['animation', 'enabled']} label="启用动画" valuePropName="checked">
                            <Switch />
                          </Form.Item>
                        </Col>
                        <Col span={12}>
                          <Form.Item name={['animation', 'duration']} label="动画时长(ms)">
                            <InputNumber min={100} max={1000} />
                          </Form.Item>
                        </Col>
                        <Col span={12}>
                          <Form.Item name={['animation', 'easing']} label="缓动函数">
                            <Select>
                              <Select.Option value="ease">ease</Select.Option>
                              <Select.Option value="ease-in">ease-in</Select.Option>
                              <Select.Option value="ease-out">ease-out</Select.Option>
                              <Select.Option value="ease-in-out">ease-in-out</Select.Option>
                              <Select.Option value="linear">linear</Select.Option>
                            </Select>
                          </Form.Item>
                        </Col>
                        <Col span={12}>
                          <Form.Item name={['animation', 'hoverEffect']} label="悬停效果" valuePropName="checked">
                            <Switch />
                          </Form.Item>
                        </Col>
                        <Col span={12}>
                          <Form.Item name={['animation', 'expandAnimation']} label="展开动画" valuePropName="checked">
                            <Switch />
                          </Form.Item>
                        </Col>
                      </Row>
                    </Card>
                  ),
                },
                {
                  key: 'typography',
                  label: (
                    <span>
                      <FontSizeOutlined />
                      字体排版
                    </span>
                  ),
                  children: (
                    <Card size="small">
                      <Row gutter={16}>
                        <Col span={24}>
                          <Form.Item name={['typography', 'fontFamily']} label="字体族">
                            <Input />
                          </Form.Item>
                        </Col>
                        <Col span={12}>
                          <Form.Item name={['typography', 'fontWeight']} label="字体粗细">
                            <Slider min={100} max={900} step={100} />
                          </Form.Item>
                        </Col>
                        <Col span={12}>
                          <Form.Item name={['typography', 'lineHeight']} label="行高">
                            <Slider min={1} max={2} step={0.1} />
                          </Form.Item>
                        </Col>
                        <Col span={12}>
                          <Form.Item name={['typography', 'letterSpacing']} label="字间距">
                            <Slider min={-2} max={4} step={0.1} />
                          </Form.Item>
                        </Col>
                      </Row>
                    </Card>
                  ),
                },
                {
                  key: 'effects',
                  label: '视觉效果',
                  children: (
                    <Card size="small">
                      <Row gutter={16}>
                        <Col span={12}>
                          <Form.Item name={['effects', 'shadow']} label="阴影效果" valuePropName="checked">
                            <Switch />
                          </Form.Item>
                        </Col>
                        <Col span={12}>
                          <Form.Item name={['effects', 'shadowColor']} label="阴影颜色">
                            <ColorPicker showText />
                          </Form.Item>
                        </Col>
                        <Col span={12}>
                          <Form.Item name={['effects', 'shadowBlur']} label="阴影模糊">
                            <Slider min={0} max={20} />
                          </Form.Item>
                        </Col>
                        <Col span={12}>
                          <Form.Item name={['effects', 'gradient']} label="渐变背景" valuePropName="checked">
                            <Switch />
                          </Form.Item>
                        </Col>
                        <Col span={12}>
                          <Form.Item name={['effects', 'gradientStart']} label="渐变起始">
                            <ColorPicker showText />
                          </Form.Item>
                        </Col>
                        <Col span={12}>
                          <Form.Item name={['effects', 'gradientEnd']} label="渐变结束">
                            <ColorPicker showText />
                          </Form.Item>
                        </Col>
                        <Col span={12}>
                          <Form.Item name={['effects', 'backdrop']} label="背景模糊" valuePropName="checked">
                            <Switch />
                          </Form.Item>
                        </Col>
                        <Col span={12}>
                          <Form.Item name={['effects', 'backdropBlur']} label="模糊程度">
                            <Slider min={0} max={20} />
                          </Form.Item>
                        </Col>
                      </Row>
                    </Card>
                  ),
                },
                {
                  key: 'responsive',
                  label: '响应式',
                  children: (
                    <Card size="small">
                      <Row gutter={16}>
                        <Col span={8}>
                          <Form.Item name={['responsive', 'breakpoints', 'mobile']} label="移动端断点">
                            <InputNumber min={320} max={1024} />
                          </Form.Item>
                        </Col>
                        <Col span={8}>
                          <Form.Item name={['responsive', 'breakpoints', 'tablet']} label="平板断点">
                            <InputNumber min={768} max={1440} />
                          </Form.Item>
                        </Col>
                        <Col span={8}>
                          <Form.Item name={['responsive', 'breakpoints', 'desktop']} label="桌面断点">
                            <InputNumber min={1024} max={2560} />
                          </Form.Item>
                        </Col>
                        <Col span={8}>
                          <Form.Item name={['responsive', 'mobileLayout']} label="移动端布局">
                            <Select>
                              <Select.Option value="drawer">抽屉</Select.Option>
                              <Select.Option value="bottom">底部</Select.Option>
                              <Select.Option value="top">顶部</Select.Option>
                            </Select>
                          </Form.Item>
                        </Col>
                        <Col span={8}>
                          <Form.Item name={['responsive', 'tabletLayout']} label="平板布局">
                            <Select>
                              <Select.Option value="collapsed">折叠</Select.Option>
                              <Select.Option value="expanded">展开</Select.Option>
                              <Select.Option value="drawer">抽屉</Select.Option>
                            </Select>
                          </Form.Item>
                        </Col>
                        <Col span={8}>
                          <Form.Item name={['responsive', 'desktopLayout']} label="桌面布局">
                            <Select>
                              <Select.Option value="expanded">展开</Select.Option>
                              <Select.Option value="collapsed">折叠</Select.Option>
                              <Select.Option value="floating">浮动</Select.Option>
                            </Select>
                          </Form.Item>
                        </Col>
                      </Row>
                    </Card>
                  ),
                },
                {
                  key: 'custom',
                  label: '自定义CSS',
                  children: (
                    <Card size="small">
                      <Form.Item name="customCSS" label="自定义CSS代码">
                        <TextArea
                          rows={10}
                          placeholder="在这里输入自定义CSS代码..."
                          style={{ fontFamily: 'Monaco, Consolas, monospace' }}
                        />
                      </Form.Item>
                      <Alert
                        message="提示"
                        description="自定义CSS将会覆盖默认样式，请谨慎使用。建议使用CSS类名前缀避免样式冲突。"
                        type="info"
                        showIcon
                      />
                    </Card>
                  ),
                },
              ]}
            />
          </Form>
        </Col>
        
        {previewMode && (
          <Col span={12}>
            <Card title="实时预览" size="small">
              <style>{generatePreviewCSS()}</style>
              <div className="menu-preview">
                <div className="menu-preview-item active">
                  <div className="menu-preview-icon">🏠</div>
                  <span>首页</span>
                </div>
                <div className="menu-preview-item">
                  <div className="menu-preview-icon">📊</div>
                  <span>数据分析</span>
                </div>
                <div className="menu-preview-item">
                  <div className="menu-preview-icon">👥</div>
                  <span>用户管理</span>
                </div>
                <div className="menu-preview-item">
                  <div className="menu-preview-icon">⚙️</div>
                  <span>系统设置</span>
                </div>
                <div className="menu-preview-item">
                  <div className="menu-preview-icon">📝</div>
                  <span>菜单管理</span>
                </div>
              </div>
            </Card>
          </Col>
        )}
      </Row>
    </Modal>
  );
};

export default VisualConfig;