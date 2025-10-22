import React from 'react';
import { Card, Descriptions, Avatar, Space, Tag, Typography } from 'antd';
import { useAuthStore } from '@/stores/auth';
import dayjs from 'dayjs';

const Profile: React.FC = () => {
  const { user, token, expiresAt } = useAuthStore();

  const isDevToken = token?.startsWith('dev.');
  const expiresText = expiresAt ? dayjs(expiresAt).format('YYYY-MM-DD HH:mm:ss') : '未知';
  const isExpired = expiresAt ? dayjs(expiresAt).isBefore(dayjs()) : false;

  return (
    <Card title="个人资料" extra={isDevToken ? <Tag color="purple">开发模式</Tag> : undefined}>
      {user ? (
        <Space align="start" size={24}>
          <Avatar size={80} src={user.avatar}>
            {(user.name || user.username || 'A').charAt(0)}
          </Avatar>
          <Descriptions column={1} bordered size="small">
            <Descriptions.Item label="姓名">{user.name || '-'}</Descriptions.Item>
            <Descriptions.Item label="用户名">{user.username || '-'}</Descriptions.Item>
            <Descriptions.Item label="邮箱">{user.email || '-'}</Descriptions.Item>
            <Descriptions.Item label="状态">{user.status || '-'}</Descriptions.Item>
            <Descriptions.Item label="角色">
              {user.roles?.length ? user.roles.map(r => r.name).join(', ') : '-'}
            </Descriptions.Item>
            <Descriptions.Item label="上次登录时间">{user.lastLoginAt || '-'}</Descriptions.Item>
            <Descriptions.Item label="创建时间">{user.createdAt || '-'}</Descriptions.Item>
            <Descriptions.Item label="令牌到期">
              <Space>
                <Tag color={isExpired ? 'red' : 'green'}>{isExpired ? '已过期' : '有效'}</Tag>
                <Typography.Text>{expiresText}</Typography.Text>
              </Space>
            </Descriptions.Item>
            <Descriptions.Item label="令牌">
              <Typography.Text copyable>
                {token ? (token.length > 32 ? token.slice(0, 32) + '...' : token) : '-'}
              </Typography.Text>
            </Descriptions.Item>
          </Descriptions>
        </Space>
      ) : (
        <Typography.Text type="secondary">暂无用户信息，请登录后查看。</Typography.Text>
      )}
    </Card>
  );
};

export default Profile;