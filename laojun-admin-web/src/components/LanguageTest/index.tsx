import React from 'react';
import { Card, Space, Typography, Button } from 'antd';
import { useTranslation } from '../../locales';

const { Title, Text } = Typography;

const LanguageTest: React.FC = () => {
  const { t } = useTranslation();

  return (
    <Card title="语言切换测试" style={{ margin: '20px' }}>
      <Space direction="vertical" size="large" style={{ width: '100%' }}>
        <div>
          <Title level={4}>通用翻译测试</Title>
          <Space wrap>
            <Button type="primary">{t('common.confirm')}</Button>
            <Button>{t('common.cancel')}</Button>
            <Button type="dashed">{t('common.save')}</Button>
            <Button type="link">{t('common.edit')}</Button>
          </Space>
        </div>

        <div>
          <Title level={4}>菜单翻译测试</Title>
          <Space direction="vertical">
            <Text>• {t('menu.dashboard')}</Text>
            <Text>• {t('menu.systemManagement')}</Text>
            <Text>• {t('menu.userManagement')}</Text>
            <Text>• {t('menu.roleManagement')}</Text>
          </Space>
        </div>

        <div>
          <Title level={4}>认证翻译测试</Title>
          <Space direction="vertical">
            <Text>• {t('auth.logout')}</Text>
            <Text>• {t('auth.logoutSuccess')}</Text>
            <Text>• {t('auth.expired')}</Text>
          </Space>
        </div>

        <div>
          <Title level={4}>仪表板翻译测试</Title>
          <Space direction="vertical">
            <Text>• {t('dashboard.welcome', { name: '测试用户' })}</Text>
            <Text>• {t('dashboard.totalUsers')}</Text>
            <Text>• {t('dashboard.systemStatus')}</Text>
            <Text>• {t('dashboard.recentActivity')}</Text>
          </Space>
        </div>

        <div>
          <Title level={4}>用户类型测试</Title>
          <Space wrap>
            <Text>• {t('common.user')}</Text>
            <Text>• {t('common.admin')}</Text>
            <Text>• {t('common.devMode')}</Text>
          </Space>
        </div>
      </Space>
    </Card>
  );
};

export default LanguageTest;