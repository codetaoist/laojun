import React, { createContext, useContext, useEffect, useState } from 'react';

interface AccessibilityContextType {
  highContrast: boolean;
  reducedMotion: boolean;
  fontSize: 'small' | 'medium' | 'large';
  keyboardNavigation: boolean;
  screenReader: boolean;
  toggleHighContrast: () => void;
  toggleReducedMotion: () => void;
  setFontSize: (size: 'small' | 'medium' | 'large') => void;
  announceToScreenReader: (message: string) => void;
}

const AccessibilityContext = createContext<AccessibilityContextType | undefined>(undefined);

export const useAccessibility = () => {
  const context = useContext(AccessibilityContext);
  if (!context) {
    throw new Error('useAccessibility must be used within AccessibilityProvider');
  }
  return context;
};

interface AccessibilityProviderProps {
  children: React.ReactNode;
}

export const AccessibilityProvider: React.FC<AccessibilityProviderProps> = ({ children }) => {
  const [highContrast, setHighContrast] = useState(false);
  const [reducedMotion, setReducedMotion] = useState(false);
  const [fontSize, setFontSizeState] = useState<'small' | 'medium' | 'large'>('medium');
  const [keyboardNavigation, setKeyboardNavigation] = useState(false);
  const [screenReader, setScreenReader] = useState(false);

  // 检测用户偏好设置
  useEffect(() => {
    // 检测高对比度偏好
    const highContrastQuery = window.matchMedia('(prefers-contrast: high)');
    setHighContrast(highContrastQuery.matches);
    
    const handleHighContrastChange = (e: MediaQueryListEvent) => {
      setHighContrast(e.matches);
    };
    highContrastQuery.addEventListener('change', handleHighContrastChange);

    // 检测减少动画偏好
    const reducedMotionQuery = window.matchMedia('(prefers-reduced-motion: reduce)');
    setReducedMotion(reducedMotionQuery.matches);
    
    const handleReducedMotionChange = (e: MediaQueryListEvent) => {
      setReducedMotion(e.matches);
    };
    reducedMotionQuery.addEventListener('change', handleReducedMotionChange);

    // 检测屏幕阅读器
    const detectScreenReader = () => {
      // 简单的屏幕阅读器检测
      const isScreenReader = window.navigator.userAgent.includes('NVDA') ||
                            window.navigator.userAgent.includes('JAWS') ||
                            window.speechSynthesis !== undefined;
      setScreenReader(isScreenReader);
    };
    detectScreenReader();

    // 键盘导航检测
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Tab') {
        setKeyboardNavigation(true);
      }
    };

    const handleMouseDown = () => {
      setKeyboardNavigation(false);
    };

    document.addEventListener('keydown', handleKeyDown);
    document.addEventListener('mousedown', handleMouseDown);

    // 从 localStorage 恢复设置
    const savedSettings = localStorage.getItem('accessibility-settings');
    if (savedSettings) {
      try {
        const settings = JSON.parse(savedSettings);
        if (settings.fontSize) setFontSizeState(settings.fontSize);
        if (settings.highContrast !== undefined) setHighContrast(settings.highContrast);
        if (settings.reducedMotion !== undefined) setReducedMotion(settings.reducedMotion);
      } catch (error) {
        console.warn('Failed to parse accessibility settings:', error);
      }
    }

    return () => {
      highContrastQuery.removeEventListener('change', handleHighContrastChange);
      reducedMotionQuery.removeEventListener('change', handleReducedMotionChange);
      document.removeEventListener('keydown', handleKeyDown);
      document.removeEventListener('mousedown', handleMouseDown);
    };
  }, []);

  // 保存设置到 localStorage
  useEffect(() => {
    const settings = {
      fontSize,
      highContrast,
      reducedMotion,
    };
    localStorage.setItem('accessibility-settings', JSON.stringify(settings));
  }, [fontSize, highContrast, reducedMotion]);

  // 应用样式类
  useEffect(() => {
    const root = document.documentElement;
    
    // 高对比度
    if (highContrast) {
      root.classList.add('high-contrast');
    } else {
      root.classList.remove('high-contrast');
    }

    // 减少动画
    if (reducedMotion) {
      root.classList.add('reduced-motion');
    } else {
      root.classList.remove('reduced-motion');
    }

    // 字体大小
    root.classList.remove('font-small', 'font-medium', 'font-large');
    root.classList.add(`font-${fontSize}`);

    // 键盘导航
    if (keyboardNavigation) {
      root.classList.add('keyboard-navigation');
    } else {
      root.classList.remove('keyboard-navigation');
    }
  }, [highContrast, reducedMotion, fontSize, keyboardNavigation]);

  const toggleHighContrast = () => {
    setHighContrast(!highContrast);
  };

  const toggleReducedMotion = () => {
    setReducedMotion(!reducedMotion);
  };

  const setFontSize = (size: 'small' | 'medium' | 'large') => {
    setFontSizeState(size);
  };

  // 屏幕阅读器公告
  const announceToScreenReader = (message: string) => {
    const announcement = document.createElement('div');
    announcement.setAttribute('aria-live', 'polite');
    announcement.setAttribute('aria-atomic', 'true');
    announcement.className = 'sr-only';
    announcement.textContent = message;
    
    document.body.appendChild(announcement);
    
    // 清理
    setTimeout(() => {
      document.body.removeChild(announcement);
    }, 1000);
  };

  const value: AccessibilityContextType = {
    highContrast,
    reducedMotion,
    fontSize,
    keyboardNavigation,
    screenReader,
    toggleHighContrast,
    toggleReducedMotion,
    setFontSize,
    announceToScreenReader,
  };

  return (
    <AccessibilityContext.Provider value={value}>
      {children}
      {/* 跳转到主内容的链接 */}
      <a
        href="#main-content"
        className="sr-only focus:not-sr-only focus:absolute focus:top-4 focus:left-4 focus:z-50 focus:px-4 focus:py-2 focus:bg-blue-600 focus:text-white focus:rounded"
      >
        跳转到主内容
      </a>
    </AccessibilityContext.Provider>
  );
};

export default AccessibilityProvider;