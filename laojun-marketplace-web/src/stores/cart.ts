import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import { CartState, CartItem, Plugin } from '@/types';
import { message } from 'antd';

interface CartStore extends CartState {
  // 添加到购物车
  addToCart: (plugin: Plugin) => void;
  
  // 从购物车移除
  removeFromCart: (pluginId: string) => void;
  
  // 更新数量
  updateQuantity: (pluginId: string, quantity: number) => void;
  
  // 清空购物车
  clearCart: () => void;
  
  // 获取购物车项目数量
  getItemCount: () => number;
  
  // 检查插件是否在购物车中
  isInCart: (pluginId: string) => boolean;
  
  // 计算总价
  calculateTotal: () => void;
}

const initialState: CartState = {
  items: [],
  total: 0,
  currency: 'CNY',
};

export const useCartStore = create<CartStore>()(
  persist(
    (set, get) => ({
      ...initialState,

      addToCart: (plugin) => {
        const { items } = get();
        
        // 检查是否为免费插件
        if (plugin.price === 0) {
          message.info('免费插件无需加入购物车，可直接安装');
          return;
        }

        // 检查是否已在购物车中
        const existingItem = items.find(item => item.pluginId === plugin.id);
        
        if (existingItem) {
          message.info('该插件已在购物车中');
          return;
        }

        const newItem: CartItem = {
          pluginId: plugin.id,
          plugin,
          quantity: 1,
          addedAt: new Date().toISOString(),
        };

        const newItems = [...items, newItem];
        set({ items: newItems });
        get().calculateTotal();
        
        message.success('已添加到购物车');
      },

      removeFromCart: (pluginId) => {
        const { items } = get();
        const newItems = items.filter(item => item.pluginId !== pluginId);
        set({ items: newItems });
        get().calculateTotal();
        
        message.success('已从购物车移除');
      },

      updateQuantity: (pluginId, quantity) => {
        if (quantity <= 0) {
          get().removeFromCart(pluginId);
          return;
        }

        const { items } = get();
        const newItems = items.map(item =>
          item.pluginId === pluginId
            ? { ...item, quantity }
            : item
        );
        
        set({ items: newItems });
        get().calculateTotal();
      },

      clearCart: () => {
        set({ items: [], total: 0 });
        message.success('购物车已清空');
      },

      getItemCount: () => {
        const { items } = get();
        return items.reduce((count, item) => count + item.quantity, 0);
      },

      isInCart: (pluginId) => {
        const { items } = get();
        return items.some(item => item.pluginId === pluginId);
      },

      calculateTotal: () => {
        const { items } = get();
        const total = items.reduce((sum, item) => {
          return sum + (item.plugin.price * item.quantity);
        }, 0);
        
        set({ total });
      },
    }),
    {
      name: 'marketplace-cart-store',
      partialize: (state) => ({
        items: state.items,
        total: state.total,
        currency: state.currency,
      }),
    }
  )
);

export default useCartStore;