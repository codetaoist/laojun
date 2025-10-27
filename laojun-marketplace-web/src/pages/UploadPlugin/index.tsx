import React, { useState } from 'react';
import {
  Card,
  Form,
  Input,
  Button,
  Upload,
  Select,
  InputNumber,
  message,
  Steps,
  Row,
  Col,
  Typography,
  Space,
  Tag,
  Divider,
  Alert,
} from 'antd';
import {
  UploadOutlined,
  InboxOutlined,
  CheckCircleOutlined,
  FileZipOutlined,
  InfoCircleOutlined,
  DownloadOutlined,
} from '@ant-design/icons';
import type { UploadProps, UploadFile } from 'antd';
import { useNavigate } from 'react-router-dom';
import { pluginService } from '@/services/plugin';
import './index.scss';

const { Title, Text, Paragraph } = Typography;
const { TextArea } = Input;
const { Option } = Select;
const { Dragger } = Upload;

interface PluginFormData {
  name: string;
  version: string;
  description: string;
  category: string;
  price: number;
  tags: string[];
  author: string;
  homepage?: string;
  repository?: string;
  license: string;
  requirements?: string;
}

const UploadPlugin: React.FC = () => {
  const [form] = Form.useForm();
  const navigate = useNavigate();
  const [currentStep, setCurrentStep] = useState(0);
  const [fileList, setFileList] = useState<UploadFile[]>([]);
  const [uploading, setUploading] = useState(false);
  const [formData, setFormData] = useState<PluginFormData>();

  const categories = [
    '开发工具',
    '代码编辑',
    '调试工具',
    '版本控制',
    '数据库',
    '网络工具',
    '安全工具',
    '性能优化',
    'UI/UX',
    '测试工具',
    '其他',
  ];

  const licenses = [
    'MIT',
    'Apache-2.0',
    'GPL-3.0',
    'BSD-3-Clause',
    'ISC',
    'MPL-2.0',
    'LGPL-3.0',
    '专有许可',
  ];

  const uploadProps: UploadProps = {
    name: 'file',
    multiple: false,
    fileList,
    beforeUpload: (file) => {
      const isZip = file.type === 'application/zip' || file.name.endsWith('.zip');
      if (!isZip) {
        message.error('只能上传 ZIP 格式的插件包！');
        return false;
      }
      const isLt50M = file.size / 1024 / 1024 < 50;
      if (!isLt50M) {
        message.error('插件包大小不能超过 50MB！');
        return false;
      }
      setFileList([file]);
      return false; // 阻止自动上传
    },
    onRemove: () => {
      setFileList([]);
    },
  };

  const handleStepNext = async () => {
    if (currentStep === 0) {
      if (fileList.length === 0) {
        message.error('请先上传插件包文件');
        return;
      }
      setCurrentStep(1);
    } else if (currentStep === 1) {
      try {
        const values = await form.validateFields();
        setFormData(values);
        setCurrentStep(2);
      } catch (error) {
        message.error('请完善插件信息');
      }
    }
  };

  const handleStepPrev = () => {
    setCurrentStep(currentStep - 1);
  };

  const handleSubmit = async () => {
    if (!formData || fileList.length === 0) {
      message.error('请完善所有信息');
      return;
    }

    setUploading(true);
    try {
      const formDataToSend = new FormData();
      formDataToSend.append('file', fileList[0] as any);
      formDataToSend.append('pluginInfo', JSON.stringify(formData));

      const response = await pluginService.uploadPlugin(formDataToSend);
      
      message.success(response.message || '插件上传成功！');
      navigate('/my-plugins');
    } catch (error: any) {
      console.error('上传失败:', error);
      const errorMessage = error?.response?.data?.message || error?.message || '上传失败，请重试';
      message.error(errorMessage);
    } finally {
      setUploading(false);
    }
  };

  const steps = [
    {
      title: '上传文件',
      icon: <UploadOutlined />,
    },
    {
      title: '填写信息',
      icon: <InfoCircleOutlined />,
    },
    {
      title: '确认提交',
      icon: <CheckCircleOutlined />,
    },
  ];

  const renderStepContent = () => {
    switch (currentStep) {
      case 0:
        return (
          <Card title="上传插件包" className="upload-card">
            <Alert
              message="上传要求"
              description={
                <ul>
                  <li>文件格式：ZIP 压缩包</li>
                  <li>文件大小：不超过 50MB</li>
                  <li>包含：插件代码、配置文件、说明文档</li>
                  <li>建议：包含 plugin.json 配置文件</li>
                </ul>
              }
              type="info"
              showIcon
              style={{ marginBottom: 24 }}
            />
            <Dragger {...uploadProps} style={{ padding: '40px 20px' }}>
              <p className="ant-upload-drag-icon">
                <FileZipOutlined style={{ fontSize: 48, color: '#1890ff' }} />
              </p>
              <p className="ant-upload-text">点击或拖拽文件到此区域上传</p>
              <p className="ant-upload-hint">
                支持单个文件上传，仅支持 ZIP 格式的插件包
              </p>
            </Dragger>
            
            <Divider>示例插件包</Divider>
            
            <Alert
              message="快速开始"
              description="下载示例插件包，了解插件开发结构和最佳实践"
              type="success"
              showIcon
              style={{ marginBottom: 16 }}
            />
            
            <Row gutter={[16, 16]}>
              <Col xs={24} sm={12}>
                <Card 
                  size="small" 
                  title={
                    <Space>
                      <FileZipOutlined style={{ color: '#1890ff' }} />
                      <span>Hello World 插件</span>
                    </Space>
                  }
                  extra={
                    <Button 
                      type="primary" 
                      size="small" 
                      icon={<DownloadOutlined />}
                      href="/examples/hello-world-plugin.zip"
                      download="hello-world-plugin.zip"
                    >
                      下载
                    </Button>
                  }
                >
                  <Text type="secondary" style={{ fontSize: '12px' }}>
                    进程内插件示例，展示基本的插件结构和命令注册
                  </Text>
                  <div style={{ marginTop: 8 }}>
                    <Tag color="blue" size="small">JavaScript</Tag>
                    <Tag color="green" size="small">进程内</Tag>
                    <Tag color="orange" size="small">入门</Tag>
                  </div>
                </Card>
              </Col>
              
              <Col xs={24} sm={12}>
                <Card 
                  size="small" 
                  title={
                    <Space>
                      <FileZipOutlined style={{ color: '#52c41a' }} />
                      <span>微服务插件</span>
                    </Space>
                  }
                  extra={
                    <Button 
                      type="primary" 
                      size="small" 
                      icon={<DownloadOutlined />}
                      href="/examples/microservice-plugin.zip"
                      download="microservice-plugin.zip"
                    >
                      下载
                    </Button>
                  }
                >
                  <Text type="secondary" style={{ fontSize: '12px' }}>
                    微服务插件示例，展示独立服务和 Docker 部署
                  </Text>
                  <div style={{ marginTop: 8 }}>
                    <Tag color="blue" size="small">Node.js</Tag>
                    <Tag color="purple" size="small">微服务</Tag>
                    <Tag color="red" size="small">Docker</Tag>
                  </div>
                </Card>
              </Col>
            </Row>
          </Card>
        );

      case 1:
        return (
          <Card title="插件信息" className="form-card">
            <Form
              form={form}
              layout="vertical"
              initialValues={{
                license: 'MIT',
                price: 0,
              }}
            >
              <Row gutter={16}>
                <Col span={12}>
                  <Form.Item
                    label="插件名称"
                    name="name"
                    rules={[{ required: true, message: '请输入插件名称' }]}
                  >
                    <Input placeholder="请输入插件名称" />
                  </Form.Item>
                </Col>
                <Col span={12}>
                  <Form.Item
                    label="版本号"
                    name="version"
                    rules={[{ required: true, message: '请输入版本号' }]}
                  >
                    <Input placeholder="例如：1.0.0" />
                  </Form.Item>
                </Col>
              </Row>

              <Form.Item
                label="插件描述"
                name="description"
                rules={[{ required: true, message: '请输入插件描述' }]}
              >
                <TextArea
                  rows={4}
                  placeholder="请详细描述插件的功能和特点"
                />
              </Form.Item>

              <Row gutter={16}>
                <Col span={12}>
                  <Form.Item
                    label="分类"
                    name="category"
                    rules={[{ required: true, message: '请选择分类' }]}
                  >
                    <Select placeholder="请选择插件分类">
                      {categories.map(category => (
                        <Option key={category} value={category}>
                          {category}
                        </Option>
                      ))}
                    </Select>
                  </Form.Item>
                </Col>
                <Col span={12}>
                  <Form.Item
                    label="价格 (¥)"
                    name="price"
                    rules={[{ required: true, message: '请输入价格' }]}
                  >
                    <InputNumber
                      min={0}
                      precision={2}
                      style={{ width: '100%' }}
                      placeholder="0 表示免费"
                    />
                  </Form.Item>
                </Col>
              </Row>

              <Row gutter={16}>
                <Col span={12}>
                  <Form.Item
                    label="作者"
                    name="author"
                    rules={[{ required: true, message: '请输入作者名称' }]}
                  >
                    <Input placeholder="请输入作者名称" />
                  </Form.Item>
                </Col>
                <Col span={12}>
                  <Form.Item
                    label="许可证"
                    name="license"
                    rules={[{ required: true, message: '请选择许可证' }]}
                  >
                    <Select placeholder="请选择许可证">
                      {licenses.map(license => (
                        <Option key={license} value={license}>
                          {license}
                        </Option>
                      ))}
                    </Select>
                  </Form.Item>
                </Col>
              </Row>

              <Form.Item label="标签" name="tags">
                <Select
                  mode="tags"
                  placeholder="请输入标签，按回车添加"
                  style={{ width: '100%' }}
                />
              </Form.Item>

              <Row gutter={16}>
                <Col span={12}>
                  <Form.Item label="主页地址" name="homepage">
                    <Input placeholder="https://example.com" />
                  </Form.Item>
                </Col>
                <Col span={12}>
                  <Form.Item label="代码仓库" name="repository">
                    <Input placeholder="https://github.com/user/repo" />
                  </Form.Item>
                </Col>
              </Row>

              <Form.Item label="系统要求" name="requirements">
                <TextArea
                  rows={3}
                  placeholder="请描述插件的系统要求和依赖"
                />
              </Form.Item>
            </Form>
          </Card>
        );

      case 2:
        return (
          <Card title="确认信息" className="confirm-card">
            <Row gutter={24}>
              <Col span={12}>
                <Title level={4}>文件信息</Title>
                <Space direction="vertical" style={{ width: '100%' }}>
                  <div>
                    <Text strong>文件名：</Text>
                    <Text>{fileList[0]?.name}</Text>
                  </div>
                  <div>
                    <Text strong>文件大小：</Text>
                    <Text>{((fileList[0]?.size || 0) / 1024 / 1024).toFixed(2)} MB</Text>
                  </div>
                </Space>
              </Col>
              <Col span={12}>
                <Title level={4}>插件信息</Title>
                <Space direction="vertical" style={{ width: '100%' }}>
                  <div>
                    <Text strong>名称：</Text>
                    <Text>{formData?.name}</Text>
                  </div>
                  <div>
                    <Text strong>版本：</Text>
                    <Text>{formData?.version}</Text>
                  </div>
                  <div>
                    <Text strong>分类：</Text>
                    <Tag color="blue">{formData?.category}</Tag>
                  </div>
                  <div>
                    <Text strong>价格：</Text>
                    <Text>
                      {formData?.price === 0 ? '免费' : `¥${formData?.price}`}
                    </Text>
                  </div>
                  <div>
                    <Text strong>作者：</Text>
                    <Text>{formData?.author}</Text>
                  </div>
                  <div>
                    <Text strong>许可证：</Text>
                    <Text>{formData?.license}</Text>
                  </div>
                </Space>
              </Col>
            </Row>
            <Divider />
            <div>
              <Title level={4}>描述</Title>
              <Paragraph>{formData?.description}</Paragraph>
            </div>
            {formData?.tags && formData.tags.length > 0 && (
              <div>
                <Title level={4}>标签</Title>
                <Space wrap>
                  {formData.tags.map(tag => (
                    <Tag key={tag} color="geekblue">{tag}</Tag>
                  ))}
                </Space>
              </div>
            )}
          </Card>
        );

      default:
        return null;
    }
  };

  return (
    <div className="upload-plugin-page">
      <div className="page-header">
        <Title level={2}>上传插件</Title>
        <Paragraph type="secondary">
          将您的插件分享给更多用户，让创意发光发热
        </Paragraph>
      </div>

      <Card className="steps-card">
        <Steps current={currentStep} items={steps} />
      </Card>

      <div className="step-content">
        {renderStepContent()}
      </div>

      <Card className="action-card">
        <Space>
          {currentStep > 0 && (
            <Button onClick={handleStepPrev}>
              上一步
            </Button>
          )}
          {currentStep < 2 && (
            <Button type="primary" onClick={handleStepNext}>
              下一步
            </Button>
          )}
          {currentStep === 2 && (
            <Button
              type="primary"
              loading={uploading}
              onClick={handleSubmit}
            >
              提交上传
            </Button>
          )}
          <Button onClick={() => navigate('/my-plugins')}>
            取消
          </Button>
        </Space>
      </Card>
    </div>
  );
};

export default UploadPlugin;