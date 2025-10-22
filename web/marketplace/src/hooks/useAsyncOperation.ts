import { useState, useCallback, useRef } from 'react';
import { message } from 'antd';

interface AsyncOperationState<T> {
  data: T | null;
  loading: boolean;
  error: string | null;
}

interface AsyncOperationOptions {
  showSuccessMessage?: boolean;
  showErrorMessage?: boolean;
  successMessage?: string;
  errorMessage?: string;
  retryCount?: number;
  retryDelay?: number;
}

interface UseAsyncOperationReturn<T> {
  data: T | null;
  loading: boolean;
  error: string | null;
  execute: (...args: any[]) => Promise<T | null>;
  retry: () => Promise<T | null>;
  reset: () => void;
}

export function useAsyncOperation<T>(
  asyncFunction: (...args: any[]) => Promise<T>,
  options: AsyncOperationOptions = {}
): UseAsyncOperationReturn<T> {
  const {
    showSuccessMessage = false,
    showErrorMessage = true,
    successMessage = '操作成功',
    errorMessage = '操作失败',
    retryCount = 0,
    retryDelay = 1000
  } = options;

  const [state, setState] = useState<AsyncOperationState<T>>({
    data: null,
    loading: false,
    error: null
  });

  const lastArgsRef = useRef<any[]>([]);
  const retryCountRef = useRef(0);

  const execute = useCallback(async (...args: any[]): Promise<T | null> => {
    lastArgsRef.current = args;
    retryCountRef.current = 0;

    setState(prev => ({ ...prev, loading: true, error: null }));

    try {
      const result = await asyncFunction(...args);
      setState({ data: result, loading: false, error: null });
      
      if (showSuccessMessage) {
        message.success(successMessage);
      }
      
      return result;
    } catch (error) {
      const errorMsg = error instanceof Error ? error.message : errorMessage;
      setState(prev => ({ ...prev, loading: false, error: errorMsg }));
      
      if (showErrorMessage) {
        message.error(errorMsg);
      }
      
      return null;
    }
  }, [asyncFunction, showSuccessMessage, showErrorMessage, successMessage, errorMessage]);

  const retry = useCallback(async (): Promise<T | null> => {
    if (retryCountRef.current >= retryCount) {
      if (showErrorMessage) {
        message.error('重试次数已达上限');
      }
      return null;
    }

    retryCountRef.current += 1;
    
    if (retryDelay > 0) {
      await new Promise(resolve => setTimeout(resolve, retryDelay));
    }

    return execute(...lastArgsRef.current);
  }, [execute, retryCount, retryDelay, showErrorMessage]);

  const reset = useCallback(() => {
    setState({ data: null, loading: false, error: null });
    retryCountRef.current = 0;
    lastArgsRef.current = [];
  }, []);

  return {
    data: state.data,
    loading: state.loading,
    error: state.error,
    execute,
    retry,
    reset
  };
}

export default useAsyncOperation;