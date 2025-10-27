import React, { Component, ErrorInfo, ReactNode } from 'react';
import { Result, Button } from 'antd';

// 错误边界状态
interface ErrorBoundaryState {
  hasError: boolean;
  error?: Error;
  errorInfo?: ErrorInfo;
}

// 错误边界属性
export interface ErrorBoundaryProps {
  children: ReactNode;
  fallback?: ReactNode;
  onError?: (error: Error, errorInfo: ErrorInfo) => void;
  showDetails?: boolean;
  title?: string;
  subTitle?: string;
  extra?: ReactNode;
}

// 错误边界组件
export class ErrorBoundary extends Component<ErrorBoundaryProps, ErrorBoundaryState> {
  constructor(props: ErrorBoundaryProps) {
    super(props);
    this.state = { hasError: false };
  }

  static getDerivedStateFromError(error: Error): ErrorBoundaryState {
    // 更新 state 使下一次渲染能够显示降级后的 UI
    return {
      hasError: true,
      error,
    };
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    // 记录错误信息
    console.error('ErrorBoundary caught an error:', error, errorInfo);
    
    // 更新状态
    this.setState({
      error,
      errorInfo,
    });

    // 调用错误回调
    this.props.onError?.(error, errorInfo);

    // 可以在这里上报错误到监控系统
    this.reportError(error, errorInfo);
  }

  private reportError = (error: Error, errorInfo: ErrorInfo) => {
    // 这里可以集成错误监控服务，如 Sentry
    try {
      const errorReport = {
        message: error.message,
        stack: error.stack,
        componentStack: errorInfo.componentStack,
        timestamp: new Date().toISOString(),
        userAgent: navigator.userAgent,
        url: window.location.href,
      };

      // 发送错误报告到监控服务
      // 例如：sendToMonitoringService(errorReport);
      console.log('Error report:', errorReport);
    } catch (reportError) {
      console.error('Failed to report error:', reportError);
    }
  };

  private handleRetry = () => {
    this.setState({ hasError: false, error: undefined, errorInfo: undefined });
  };

  private handleReload = () => {
    window.location.reload();
  };

  render() {
    if (this.state.hasError) {
      // 如果有自定义的 fallback，使用它
      if (this.props.fallback) {
        return this.props.fallback;
      }

      // 默认错误页面
      const {
        title = '页面出现错误',
        subTitle = '抱歉，页面出现了意外错误。请尝试刷新页面或联系技术支持。',
        showDetails = typeof process !== 'undefined' && process.env?.NODE_ENV === 'development',
        extra,
      } = this.props;

      const defaultExtra = (
        <div style={{ display: 'flex', gap: '8px', justifyContent: 'center' }}>
          <Button type="primary" onClick={this.handleRetry}>
            重试
          </Button>
          <Button onClick={this.handleReload}>
            刷新页面
          </Button>
        </div>
      );

      return (
        <div style={{ padding: '50px', textAlign: 'center' }}>
          <Result
            status="error"
            title={title}
            subTitle={subTitle}
            extra={extra || defaultExtra}
          />
          
          {/* 开发环境显示错误详情 */}
          {showDetails && this.state.error && (
            <details style={{ marginTop: '20px', textAlign: 'left' }}>
              <summary style={{ cursor: 'pointer', marginBottom: '10px' }}>
                <strong>错误详情（开发模式）</strong>
              </summary>
              <div style={{ 
                background: '#f5f5f5', 
                padding: '15px', 
                borderRadius: '4px',
                fontSize: '12px',
                fontFamily: 'monospace',
                whiteSpace: 'pre-wrap',
                overflow: 'auto',
                maxHeight: '300px',
              }}>
                <div style={{ marginBottom: '10px' }}>
                  <strong>错误信息:</strong>
                  <br />
                  {this.state.error.message}
                </div>
                
                <div style={{ marginBottom: '10px' }}>
                  <strong>错误堆栈:</strong>
                  <br />
                  {this.state.error.stack}
                </div>
                
                {this.state.errorInfo && (
                  <div>
                    <strong>组件堆栈:</strong>
                    <br />
                    {this.state.errorInfo.componentStack}
                  </div>
                )}
              </div>
            </details>
          )}
        </div>
      );
    }

    return this.props.children;
  }
}

// 函数式错误边界 Hook（React 18+）
export const useErrorHandler = () => {
  const [error, setError] = React.useState<Error | null>(null);

  const resetError = React.useCallback(() => {
    setError(null);
  }, []);

  const captureError = React.useCallback((error: Error) => {
    setError(error);
  }, []);

  React.useEffect(() => {
    if (error) {
      throw error;
    }
  }, [error]);

  return { captureError, resetError };
};

// 异步错误边界组件
export interface AsyncErrorBoundaryProps extends ErrorBoundaryProps {
  onReset?: () => void;
  resetKeys?: Array<string | number>;
}

export const AsyncErrorBoundary: React.FC<AsyncErrorBoundaryProps> = ({
  onReset,
  resetKeys = [],
  ...props
}) => {
  const prevResetKeysRef = React.useRef(resetKeys);
  const errorBoundaryRef = React.useRef<ErrorBoundary>(null);

  React.useEffect(() => {
    const prevResetKeys = prevResetKeysRef.current;
    const hasResetKeyChanged = resetKeys.some((key, index) => key !== prevResetKeys[index]);

    if (hasResetKeyChanged) {
      onReset?.();
      // 重置错误边界
      if (errorBoundaryRef.current) {
        (errorBoundaryRef.current as any).handleRetry();
      }
    }

    prevResetKeysRef.current = resetKeys;
  }, [resetKeys, onReset]);

  return <ErrorBoundary ref={errorBoundaryRef} {...props} />;
};

// 错误边界 HOC
export const withErrorBoundary = <P extends object>(
  Component: React.ComponentType<P>,
  errorBoundaryProps?: Omit<ErrorBoundaryProps, 'children'>
) => {
  const WrappedComponent = (props: P) => (
    <ErrorBoundary {...errorBoundaryProps}>
      <Component {...props} />
    </ErrorBoundary>
  );

  WrappedComponent.displayName = `withErrorBoundary(${Component.displayName || Component.name})`;

  return WrappedComponent;
};