import React, { useState } from 'react';
import { Drawer, Switch, Radio, Button, Space, Typography, Divider } from 'antd';
import { SettingOutlined, EyeOutlined, FontSizeOutlined, SoundOutlined } from '@ant-design/icons';
import { useAccessibility } from '../AccessibilityProvider';
import { useTranslation } from '../../locales';

const { Title, Text } = Typography;

interface AccessibilityPanelProps {
  visible?: boolean;
  onClose?: () => void;
}

const AccessibilityPanel: React.FC<AccessibilityPanelProps> = ({ 
  visible: controlledVisible, 
  onClose: controlledOnClose 
}) => {
  const [internalVisible, setInternalVisible] = useState(false);
  const { t } = useTranslation();
  const {
    highContrast,
    reducedMotion,
    fontSize,
    screenReader,
    toggleHighContrast,
    toggleReducedMotion,
    setFontSize,
    announceToScreenReader,
  } = useAccessibility();

  const isControlled = controlledVisible !== undefined;
  const visible = isControlled ? controlledVisible : internalVisible;
  
  const handleClose = () => {
    if (isControlled && controlledOnClose) {
      controlledOnClose();
    } else {
      setInternalVisible(false);
    }
  };

  const handleOpen = () => {
    if (!isControlled) {
      setInternalVisible(true);
    }
  };

  const handleHighContrastChange = (checked: boolean) => {
    toggleHighContrast();
    announceToScreenReader(
      checked ? '已启用高对比度模式' : '已禁用高对比度模式'
    );
  };

  const handleReducedMotionChange = (checked: boolean) => {
    toggleReducedMotion();
    announceToScreenReader(
      checked ? '已启用减少动画模式' : '已禁用减少动画模式'
    );
  };

  const handleFontSizeChange = (size: 'small' | 'medium' | 'large') => {
    setFontSize(size);
    const sizeNames = {
      small: '小',
      medium: '中',
      large: '大'
    };
    announceToScreenReader(`字体大小已设置为${sizeNames[size]}`);
  };

  return (
    <>
      {!isControlled && (
        <Button
          type="text"
          icon={<SettingOutlined />}
          onClick={handleOpen}
          aria-label="打开无障碍设置"
          style={{
            position: 'fixed',
            right: '20px',
            bottom: '20px',
            zIndex: 1000,
            backgroundColor: '#1890ff',
            color: 'white',
            borderRadius: '50%',
            width: '48px',
            height: '48px',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
          }}
        />
      )}

      <Drawer
        title={
          <Space>
            <SettingOutlined />
            <span>无障碍设置</span>
          </Space>
        }
        placement="right"
        onClose={handleClose}
        open={visible}
        width={320}
        styles={{
          body: { padding: '16px' }
        }}
      >
        <Space direction="vertical" style={{ width: '100%' }} size="large">
          {/* 视觉设置 */}
          <div>
            <Title level={5}>
              <EyeOutlined /> 视觉设置
            </Title>
            
            <Space direction="vertical" style={{ width: '100%' }} size="middle">
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <Text>高对比度模式</Text>
                <Switch
                  checked={highContrast}
                  onChange={handleHighContrastChange}
                  aria-label="切换高对比度模式"
                />
              </div>
              
              <div>
                <Text style={{ display: 'block', marginBottom: '8px' }}>
                  <FontSizeOutlined /> 字体大小
                </Text>
                <Radio.Group
                  value={fontSize}
                  onChange={(e) => handleFontSizeChange(e.target.value)}
                  style={{ width: '100%' }}
                >
                  <Space direction="vertical" style={{ width: '100%' }}>
                    <Radio value="small">小</Radio>
                    <Radio value="medium">中（默认）</Radio>
                    <Radio value="large">大</Radio>
                  </Space>
                </Radio.Group>
              </div>
            </Space>
          </div>

          <Divider />

          {/* 动画设置 */}
          <div>
            <Title level={5}>动画设置</Title>
            
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <div>
                <Text>减少动画</Text>
                <br />
                <Text type="secondary" style={{ fontSize: '12px' }}>
                  减少页面动画效果
                </Text>
              </div>
              <Switch
                checked={reducedMotion}
                onChange={handleReducedMotionChange}
                aria-label="切换减少动画模式"
              />
            </div>
          </div>

          <Divider />

          {/* 屏幕阅读器信息 */}
          <div>
            <Title level={5}>
              <SoundOutlined /> 屏幕阅读器
            </Title>
            
            <Space direction="vertical" style={{ width: '100%' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <Text>检测状态</Text>
                <Text type={screenReader ? 'success' : 'secondary'}>
                  {screenReader ? '已检测到' : '未检测到'}
                </Text>
              </div>
              
              <Button
                type="primary"
                ghost
                size="small"
                onClick={() => announceToScreenReader('这是一个测试公告')}
                style={{ width: '100%' }}
              >
                测试屏幕阅读器公告
              </Button>
            </Space>
          </div>

          <Divider />

          {/* 键盘导航提示 */}
          <div>
            <Title level={5}>键盘导航</Title>
            <Text type="secondary" style={{ fontSize: '12px', lineHeight: '1.5' }}>
              • 使用 Tab 键在元素间导航<br />
              • 使用 Enter 或空格键激活按钮<br />
              • 使用方向键在菜单中导航<br />
              • 使用 Esc 键关闭弹窗
            </Text>
          </div>

          <Divider />

          {/* 重置按钮 */}
          <Button
            type="default"
            block
            onClick={() => {
              setFontSize('medium');
              if (highContrast) toggleHighContrast();
              if (reducedMotion) toggleReducedMotion();
              announceToScreenReader('无障碍设置已重置为默认值');
            }}
          >
            重置为默认设置
          </Button>
        </Space>
      </Drawer>
    </>
  );
};

export default AccessibilityPanel;