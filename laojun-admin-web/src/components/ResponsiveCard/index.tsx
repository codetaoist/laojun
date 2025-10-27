import React from 'react';
import { Card, CardProps } from 'antd';
import { useMediaQuery } from '@/hooks/useMediaQuery';

interface ResponsiveCardProps extends CardProps {
  mobileStyle?: React.CSSProperties;
  tabletStyle?: React.CSSProperties;
  desktopStyle?: React.CSSProperties;
  animation?: 'fade-in' | 'slide-in' | 'slide-up' | 'zoom-in' | 'bounce-in';
  hover?: boolean;
}

const ResponsiveCard: React.FC<ResponsiveCardProps> = ({
  children,
  mobileStyle,
  tabletStyle,
  desktopStyle,
  animation = 'fade-in',
  hover = false,
  style,
  className,
  ...props
}) => {
  const isMobile = useMediaQuery('(max-width: 768px)');
  const isTablet = useMediaQuery('(max-width: 1024px)');

  // 根据屏幕尺寸选择样式
  const getResponsiveStyle = (): React.CSSProperties => {
    if (isMobile && mobileStyle) {
      return { ...style, ...mobileStyle };
    }
    if (isTablet && tabletStyle) {
      return { ...style, ...tabletStyle };
    }
    if (desktopStyle) {
      return { ...style, ...desktopStyle };
    }
    return style || {};
  };

  // 组合类名
  const getClassName = (): string => {
    const classes = [className, animation];
    if (hover) {
      classes.push('hover-lift');
    }
    return classes.filter(Boolean).join(' ');
  };

  return (
    <Card
      {...props}
      style={getResponsiveStyle()}
      className={getClassName()}
      bodyStyle={{
        padding: isMobile ? '16px' : '24px',
        ...props.bodyStyle,
      }}
    >
      {children}
    </Card>
  );
};

export default ResponsiveCard;