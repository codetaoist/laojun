import React, { useEffect, useCallback } from 'react';
import { message, Modal } from 'antd';

export interface KeyboardShortcut {
  key: string;
  description: string;
  action: () => void;
  ctrlKey?: boolean;
  altKey?: boolean;
  shiftKey?: boolean;
  metaKey?: boolean;
}

export interface KeyboardShortcutsProps {
  shortcuts: KeyboardShortcut[];
  enabled?: boolean;
  showHelp?: boolean;
  onShowHelp?: () => void;
}

const KeyboardShortcuts: React.FC<KeyboardShortcutsProps> = ({
  shortcuts,
  enabled = true,
  showHelp = false,
  onShowHelp
}) => {
  // 处理键盘事件
  const handleKeyDown = useCallback((event: KeyboardEvent) => {
    if (!enabled) return;

    // 忽略在输入框中的按键
    const target = event.target as HTMLElement;
    if (target.tagName === 'INPUT' || target.tagName === 'TEXTAREA' || target.contentEditable === 'true') {
      return;
    }

    // 查找匹配的快捷键
    const matchedShortcut = shortcuts.find(shortcut => {
      const keyMatch = shortcut.key.toLowerCase() === event.key.toLowerCase();
      const ctrlMatch = !!shortcut.ctrlKey === event.ctrlKey;
      const altMatch = !!shortcut.altKey === event.altKey;
      const shiftMatch = !!shortcut.shiftKey === event.shiftKey;
      const metaMatch = !!shortcut.metaKey === event.metaKey;

      return keyMatch && ctrlMatch && altMatch && shiftMatch && metaMatch;
    });

    if (matchedShortcut) {
      event.preventDefault();
      event.stopPropagation();
      
      try {
        matchedShortcut.action();
      } catch (error) {
        console.error('快捷键执行失败:', error);
        message.error('快捷键执行失败');
      }
    }
  }, [shortcuts, enabled]);

  // 注册键盘事件监听器
  useEffect(() => {
    if (enabled) {
      document.addEventListener('keydown', handleKeyDown);
      return () => {
        document.removeEventListener('keydown', handleKeyDown);
      };
    }
  }, [handleKeyDown, enabled]);

  // 显示快捷键帮助
  const showShortcutHelp = () => {
    const formatShortcut = (shortcut: KeyboardShortcut) => {
      const keys = [];
      if (shortcut.ctrlKey) keys.push('Ctrl');
      if (shortcut.altKey) keys.push('Alt');
      if (shortcut.shiftKey) keys.push('Shift');
      if (shortcut.metaKey) keys.push('Cmd');
      keys.push(shortcut.key.toUpperCase());
      return keys.join(' + ');
    };

    Modal.info({
      title: '快捷键帮助',
      width: 600,
      content: (
        <div style={{ maxHeight: '400px', overflowY: 'auto' }}>
          <table style={{ width: '100%', borderCollapse: 'collapse' }}>
            <thead>
              <tr>
                <th style={{ textAlign: 'left', padding: '8px', borderBottom: '1px solid #d9d9d9' }}>
                  快捷键
                </th>
                <th style={{ textAlign: 'left', padding: '8px', borderBottom: '1px solid #d9d9d9' }}>
                  功能描述
                </th>
              </tr>
            </thead>
            <tbody>
              {shortcuts.map((shortcut, index) => (
                <tr key={index}>
                  <td style={{ 
                    padding: '8px', 
                    borderBottom: '1px solid #f0f0f0',
                    fontFamily: 'monospace',
                    fontWeight: 'bold'
                  }}>
                    {formatShortcut(shortcut)}
                  </td>
                  <td style={{ padding: '8px', borderBottom: '1px solid #f0f0f0' }}>
                    {shortcut.description}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      ),
      okText: '知道了'
    });
  };

  // 如果需要显示帮助，则显示
  useEffect(() => {
    if (showHelp) {
      showShortcutHelp();
      onShowHelp?.();
    }
  }, [showHelp, onShowHelp]);

  return null; // 这个组件不渲染任何内容
};

export default KeyboardShortcuts;

// 预定义的常用快捷键
export const commonShortcuts = {
  // 基础操作
  save: { key: 's', ctrlKey: true, description: '保存' },
  copy: { key: 'c', ctrlKey: true, description: '复制' },
  paste: { key: 'v', ctrlKey: true, description: '粘贴' },
  cut: { key: 'x', ctrlKey: true, description: '剪切' },
  undo: { key: 'z', ctrlKey: true, description: '撤销' },
  redo: { key: 'y', ctrlKey: true, description: '重做' },
  selectAll: { key: 'a', ctrlKey: true, description: '全选' },
  
  // 导航操作
  refresh: { key: 'F5', description: '刷新' },
  search: { key: 'f', ctrlKey: true, description: '搜索' },
  help: { key: 'F1', description: '帮助' },
  
  // 菜单操作
  newItem: { key: 'n', ctrlKey: true, description: '新建' },
  editItem: { key: 'e', ctrlKey: true, description: '编辑' },
  deleteItem: { key: 'Delete', description: '删除' },
  
  // 视图操作
  toggleView: { key: 'Tab', ctrlKey: true, description: '切换视图' },
  zoomIn: { key: '=', ctrlKey: true, description: '放大' },
  zoomOut: { key: '-', ctrlKey: true, description: '缩小' },
  
  // ESC键
  escape: { key: 'Escape', description: '取消/关闭' }
};

// 快捷键组合工具函数
export const createShortcut = (
  key: string,
  action: () => void,
  description: string,
  modifiers: {
    ctrl?: boolean;
    alt?: boolean;
    shift?: boolean;
    meta?: boolean;
  } = {}
): KeyboardShortcut => ({
  key,
  action,
  description,
  ctrlKey: modifiers.ctrl,
  altKey: modifiers.alt,
  shiftKey: modifiers.shift,
  metaKey: modifiers.meta
});