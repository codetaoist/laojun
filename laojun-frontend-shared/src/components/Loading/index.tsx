import React, { Suspense } from 'react';
import { Spin, Skeleton, Card } from 'antd';
import { LoadingOutlined } from '@ant-design/icons';
import { ErrorBoundary } from '../ErrorBoundary';

// 加载指示器大小
export type LoadingSize = 'small' | 'default' | 'large';

// 基础加载组件属性
export interface LoadingSpinnerProps {
  size?: LoadingSize;
  tip?: string;
  spinning?: boolean;
  children?: React.ReactNode;
  style?: React.CSSProperties;
  className?: string;
  indicator?: React.ReactElement;
  delay?: number;
}

// 基础加载组件
export const LoadingSpinner: React.FC<LoadingSpinnerProps> = ({
  size = 'default',
  tip,
  spinning = true,
  children,
  style,
  className,
  indicator,
  delay = 0,
}) => {
  const customIndicator = indicator || <LoadingOutlined style={{ fontSize: 24 }} spin />;

  if (children) {
    return (
      <Spin
        size={size}
        tip={tip}
        spinning={spinning}
        indicator={customIndicator}
        delay={delay}
        className={className}
        style={style}
      >
        {children}
      </Spin>
    );
  }

  return (
    <div
      className={className}
      style={{
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center',
        minHeight: '200px',
        ...style,
      }}
    >
      <Spin
        size={size}
        tip={tip}
        spinning={spinning}
        indicator={customIndicator}
        delay={delay}
      />
    </div>
  );
};

// 页面加载组件
export interface PageLoadingProps {
  tip?: string;
  height?: number | string;
  style?: React.CSSProperties;
}

export const PageLoading: React.FC<PageLoadingProps> = ({
  tip = '页面加载中...',
  height = '100vh',
  style,
}) => {
  return (
    <div
      style={{
        display: 'flex',
        flexDirection: 'column',
        justifyContent: 'center',
        alignItems: 'center',
        height,
        background: '#f0f2f5',
        ...style,
      }}
    >
      <Spin size="large" tip={tip} />
    </div>
  );
};

// 内容加载组件
export interface ContentLoadingProps {
  rows?: number;
  avatar?: boolean;
  title?: boolean;
  paragraph?: boolean;
  active?: boolean;
  style?: React.CSSProperties;
  className?: string;
}

export const ContentLoading: React.FC<ContentLoadingProps> = ({
  rows = 3,
  avatar = false,
  title = true,
  paragraph = true,
  active = true,
  style,
  className,
}) => {
  return (
    <div className={className} style={style}>
      <Skeleton
        avatar={avatar}
        title={title}
        paragraph={paragraph ? { rows } : false}
        active={active}
      />
    </div>
  );
};

// 卡片加载组件
export interface CardLoadingProps {
  count?: number;
  avatar?: boolean;
  title?: boolean;
  paragraph?: boolean;
  actions?: boolean;
  style?: React.CSSProperties;
  className?: string;
}

export const CardLoading: React.FC<CardLoadingProps> = ({
  count = 1,
  avatar = true,
  title = true,
  paragraph = true,
  actions = false,
  style,
  className,
}) => {
  const cards = Array.from({ length: count }, (_, index) => (
    <Card key={index} style={{ marginBottom: 16 }}>
      <Skeleton
        avatar={avatar}
        title={title}
        paragraph={paragraph ? { rows: 2 } : false}
        active
      />
      {actions && (
        <div style={{ marginTop: 16 }}>
          <Skeleton.Button size="small" style={{ marginRight: 8 }} />
          <Skeleton.Button size="small" />
        </div>
      )}
    </Card>
  ));

  return (
    <div className={className} style={style}>
      {cards}
    </div>
  );
};

// 表格加载组件
export interface TableLoadingProps {
  columns?: number;
  rows?: number;
  style?: React.CSSProperties;
  className?: string;
}

export const TableLoading: React.FC<TableLoadingProps> = ({
  columns = 4,
  rows = 5,
  style,
  className,
}) => {
  return (
    <div className={className} style={style}>
      <Skeleton
        title={false}
        paragraph={{ rows: rows * columns, width: Array(rows * columns).fill('100%') }}
        active
      />
    </div>
  );
};

// 懒加载包装器
export interface LazyWrapperProps {
  children: React.ReactNode;
  fallback?: React.ReactNode;
  errorFallback?: React.ReactNode;
  onError?: (error: Error) => void;
}

export const LazyWrapper: React.FC<LazyWrapperProps> = ({
  children,
  fallback,
  errorFallback,
  onError,
}) => {
  const defaultFallback = fallback || <PageLoading tip="组件加载中..." />;
  
  const defaultErrorFallback = errorFallback || (
    <div style={{ 
      display: 'flex', 
      justifyContent: 'center', 
      alignItems: 'center', 
      height: '200px',
      color: '#999',
    }}>
      组件加载失败
    </div>
  );

  return (
    <Suspense fallback={defaultFallback}>
      <ErrorBoundary
        fallback={defaultErrorFallback}
        onError={onError}
      >
        {children}
      </ErrorBoundary>
    </Suspense>
  );
};

// 加载状态 Hook
export interface UseLoadingOptions {
  initialLoading?: boolean;
  delay?: number;
}

export const useLoading = (options: UseLoadingOptions = {}) => {
  const { initialLoading = false, delay = 0 } = options;
  const [loading, setLoading] = React.useState(initialLoading);
  const [delayedLoading, setDelayedLoading] = React.useState(initialLoading);

  React.useEffect(() => {
    let timer: NodeJS.Timeout;

    if (loading && delay > 0) {
      timer = setTimeout(() => {
        setDelayedLoading(true);
      }, delay);
    } else {
      setDelayedLoading(loading);
    }

    return () => {
      if (timer) {
        clearTimeout(timer);
      }
    };
  }, [loading, delay]);

  const startLoading = React.useCallback(() => {
    setLoading(true);
  }, []);

  const stopLoading = React.useCallback(() => {
    setLoading(false);
    setDelayedLoading(false);
  }, []);

  const withLoading = React.useCallback(async <T,>(promise: Promise<T>): Promise<T> => {
    startLoading();
    try {
      const result = await promise;
      return result;
    } finally {
      stopLoading();
    }
  }, [startLoading, stopLoading]);

  return {
    loading: delayedLoading,
    startLoading,
    stopLoading,
    withLoading,
  };
};