import { useState, useEffect } from 'react';

export interface BreakpointValues {
  isMobile: boolean;
  isTablet: boolean;
  isDesktop: boolean;
  isLargeScreen: boolean;
}

// 断点定义
const BREAKPOINTS = {
  mobile: 768,
  tablet: 1024,
  desktop: 1280,
} as const;

/**
 * 响应式断点 Hook
 * @returns {BreakpointValues} 各断点的状态
 */
export const useBreakpoint = (): BreakpointValues => {
  const [breakpoint, setBreakpoint] = useState<BreakpointValues>({
    isMobile: false,
    isTablet: false,
    isDesktop: false,
    isLargeScreen: false,
  });

  useEffect(() => {
    const calculateBreakpoint = () => {
      const width = window.innerWidth;
      
      setBreakpoint({
        isMobile: width < BREAKPOINTS.mobile,
        isTablet: width >= BREAKPOINTS.mobile && width < BREAKPOINTS.tablet,
        isDesktop: width >= BREAKPOINTS.tablet && width < BREAKPOINTS.desktop,
        isLargeScreen: width >= BREAKPOINTS.desktop,
      });
    };

    // 初始计算
    calculateBreakpoint();

    // 监听窗口大小变化
    window.addEventListener('resize', calculateBreakpoint);
    
    // 清理监听器
    return () => window.removeEventListener('resize', calculateBreakpoint);
  }, []);

  return breakpoint;
};

/**
 * 获取当前屏幕尺寸类型
 * @returns {'mobile' | 'tablet' | 'desktop' | 'large'} 屏幕尺寸类型
 */
export const useScreenSize = () => {
  const { isMobile, isTablet, isDesktop, isLargeScreen } = useBreakpoint();
  
  if (isMobile) return 'mobile';
  if (isTablet) return 'tablet';
  if (isDesktop) return 'desktop';
  if (isLargeScreen) return 'large';
  return 'desktop'; // 默认值
};

/**
 * 判断是否为移动设备（手机和平板）
 * @returns {boolean} 是否为移动设备
 */
export const useIsMobile = () => {
  const { isMobile, isTablet } = useBreakpoint();
  return isMobile || isTablet;
};

/**
 * 响应式值 Hook - 根据屏幕大小返回不同的值
 * @param values 不同断点对应的值
 * @returns 当前断点对应的值
 */
export const useResponsiveValue = <T>(values: {
  mobile?: T;
  tablet?: T;
  desktop?: T;
  large?: T;
  default: T;
}): T => {
  const { isMobile, isTablet, isDesktop, isLargeScreen } = useBreakpoint();
  
  if (isMobile && values.mobile !== undefined) return values.mobile;
  if (isTablet && values.tablet !== undefined) return values.tablet;
  if (isDesktop && values.desktop !== undefined) return values.desktop;
  if (isLargeScreen && values.large !== undefined) return values.large;
  
  return values.default;
};