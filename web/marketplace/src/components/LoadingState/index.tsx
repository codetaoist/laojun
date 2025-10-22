import React from 'react';
import { Spin, Skeleton, Card, List, Result } from 'antd';
import { LoadingOutlined, WifiOutlined } from '@ant-design/icons';

interface LoadingStateProps {
  loading: boolean;
  error?: string | null;
  empty?: boolean;
  emptyText?: string;
  emptyDescription?: string;
  type?: 'spin' | 'skeleton' | 'card' | 'list';
  size?: 'small' | 'default' | 'large';
  children: React.ReactNode;
  onRetry?: () => void;
}

const LoadingState: React.FC<LoadingStateProps> = ({
  loading,
  error,
  empty = false,
  emptyText = '暂无数据',
  emptyDescription = '当前没有可显示的内容',
  type = 'spin',
  size = 'default',
  children,
  onRetry
}) => {
  // 错误状态
  if (error) {
    return (
      <Result
        status="error"
        title="加载失败"
        subTitle={error}
        extra={
          onRetry && (
            <button 
              onClick={onRetry}
              className="ant-btn ant-btn-primary"
            >
              重新加载
            </button>
          )
        }
      />
    );
  }

  // 空数据状态
  if (!loading && empty) {
    return (
      <Result
        icon={<WifiOutlined />}
        title={emptyText}
        subTitle={emptyDescription}
        extra={
          onRetry && (
            <button 
              onClick={onRetry}
              className="ant-btn ant-btn-primary"
            >
              刷新
            </button>
          )
        }
      />
    );
  }

  // 加载状态
  if (loading) {
    const antIcon = <LoadingOutlined style={{ fontSize: 24 }} spin />;

    switch (type) {
      case 'skeleton':
        return (
          <div style={{ padding: '20px' }}>
            <Skeleton active paragraph={{ rows: 4 }} />
          </div>
        );

      case 'card':
        return (
          <div style={{ padding: '20px' }}>
            <Card>
              <Skeleton active paragraph={{ rows: 3 }} />
            </Card>
          </div>
        );

      case 'list':
        return (
          <div style={{ padding: '20px' }}>
            <List
              itemLayout="horizontal"
              dataSource={[1, 2, 3]}
              renderItem={() => (
                <List.Item>
                  <Skeleton avatar active paragraph={{ rows: 1 }} />
                </List.Item>
              )}
            />
          </div>
        );

      case 'spin':
      default:
        return (
          <div style={{ 
            display: 'flex', 
            justifyContent: 'center', 
            alignItems: 'center', 
            minHeight: '200px',
            padding: '20px'
          }}>
            <Spin indicator={antIcon} size={size} />
          </div>
        );
    }
  }

  return <>{children}</>;
};

export default LoadingState;