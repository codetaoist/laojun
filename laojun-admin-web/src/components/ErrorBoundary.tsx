import React from 'react';
import { Result, Button } from 'antd';

interface ErrorBoundaryState {
  hasError: boolean;
  error?: Error;
}

interface ErrorBoundaryProps {
  children: React.ReactNode;
}

class ErrorBoundary extends React.Component<ErrorBoundaryProps, ErrorBoundaryState> {
  constructor(props: ErrorBoundaryProps) {
    super(props);
    this.state = { hasError: false };
  }

  static getDerivedStateFromError(error: Error): ErrorBoundaryState {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    console.error('ErrorBoundary caught an error:', error, errorInfo);
    
    // 这里可以添加错误上报逻辑
    // reportError(error, errorInfo);
  }

  handleReload = () => {
    window.location.reload();
  };

  handleGoHome = () => {
    window.location.href = '/';
  };

  render() {
    if (this.state.hasError) {
      return (
        <div style={{ padding: '50px' }}>
          <Result
            status="500"
            title="页面出错了"
            subTitle="抱歉，页面发生了意外错误。"
            extra={
              <div style={{ display: 'flex', gap: '8px', justifyContent: 'center' }}>
                <Button type="primary" onClick={this.handleReload}>
                  重新加载
                </Button>
                <Button onClick={this.handleGoHome}>
                  返回首页
                </Button>
              </div>
            }
          >
            {process.env.NODE_ENV === 'development' && this.state.error && (
              <div style={{ 
                marginTop: '20px', 
                padding: '16px', 
                backgroundColor: '#f5f5f5', 
                borderRadius: '4px',
                textAlign: 'left'
              }}>
                <h4>错误详情（开发模式）：</h4>
                <pre style={{ 
                  fontSize: '12px', 
                  color: '#d32f2f',
                  whiteSpace: 'pre-wrap',
                  wordBreak: 'break-word'
                }}>
                  {this.state.error.stack}
                </pre>
              </div>
            )}
          </Result>
        </div>
      );
    }

    return this.props.children;
  }
}

export default ErrorBoundary;