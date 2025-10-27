import React from 'react';
import { Dropdown, Button, Space } from 'antd';
import { GlobalOutlined } from '@ant-design/icons';
import { useTranslation, Locale, localeNames } from '../../locales';
import type { MenuProps } from 'antd';

const LanguageSwitch: React.FC = () => {
  const { locale, changeLocale, t } = useTranslation();

  const items: MenuProps['items'] = Object.entries(localeNames).map(([key, name]) => ({
    key,
    label: (
      <Space>
        <span>{name}</span>
        {locale === key && <span style={{ color: '#1890ff' }}>✓</span>}
      </Space>
    ),
    onClick: () => changeLocale(key as Locale),
  }));

  const currentLanguageName = localeNames[locale];

  return (
    <Dropdown
      menu={{ items }}
      placement="bottomRight"
      trigger={['click']}
      arrow
    >
      <Button
        type="text"
        icon={<GlobalOutlined />}
        style={{
          display: 'flex',
          alignItems: 'center',
          gap: '4px',
          color: 'inherit',
        }}
      >
        <span className="hidden lg:inline">{currentLanguageName}</span>
      </Button>
    </Dropdown>
  );
};

export default LanguageSwitch;