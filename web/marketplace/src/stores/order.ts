import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import { 
  Order, 
  OrderState, 
  OrderStatus, 
  PaymentStatus,
  PaymentMethod,
  BillingInfo,
  CartItem 
} from '@/types';
import { orderService } from '@/services/order';
import { paymentService } from '@/services/payment';
import { message } from 'antd';

interface OrderStore extends OrderState {
  // 订单操作
  createOrder: (
    cartItems: CartItem[],
    billingInfo: BillingInfo,
    paymentMethod: PaymentMethod,
    userId: string
  ) => Promise<Order | null>;
  
  // 获取用户订单列表
  fetchUserOrders: (
    userId: string,
    page?: number,
    pageSize?: number,
    status?: OrderStatus
  ) => Promise<void>;
  
  // 获取订单详情
  fetchOrderById: (orderId: string) => Promise<void>;
  
  // 更新订单状态
  updateOrderStatus: (orderId: string, status: OrderStatus, notes?: string) => Promise<void>;
  
  // 取消订单
  cancelOrder: (orderId: string, reason?: string) => Promise<void>;
  
  // 申请退款
  requestRefund: (orderId: string, reason: string, amount?: number) => Promise<void>;
  
  // 处理支付
  processPayment: (orderId: string, paymentMethod: PaymentMethod) => Promise<boolean>;
  
  // 查询支付状态
  checkPaymentStatus: (paymentId: string) => Promise<void>;
  
  // 设置当前订单
  setCurrentOrder: (order: Order | undefined) => void;
  
  // 清除错误
  clearError: () => void;
  
  // 重置状态
  reset: () => void;
}

const initialState: OrderState = {
  orders: [],
  currentOrder: undefined,
  loading: false,
  error: undefined,
};

export const useOrderStore = create<OrderStore>()(
  persist(
    (set, get) => ({
      ...initialState,

      createOrder: async (cartItems, billingInfo, paymentMethod, userId) => {
        set({ loading: true, error: undefined });
        
        try {
          const response = await orderService.createOrder(
            cartItems,
            billingInfo,
            paymentMethod,
            userId
          );

          if (response.success && response.data) {
            const newOrder = response.data;
            
            // 更新订单列表
            const { orders } = get();
            set({ 
              orders: [newOrder, ...orders],
              currentOrder: newOrder,
              loading: false 
            });

            message.success(response.message || '订单创建成功');
            return newOrder;
          } else {
            throw new Error(response.message || '订单创建失败');
          }
        } catch (error) {
          const errorMessage = error instanceof Error ? error.message : '订单创建失败';
          set({ error: errorMessage, loading: false });
          message.error(errorMessage);
          return null;
        }
      },

      fetchUserOrders: async (userId, page = 1, pageSize = 10, status) => {
        set({ loading: true, error: undefined });
        
        try {
          const response = await orderService.getUserOrders(userId, page, pageSize, status);

          if (response.success && response.data) {
            set({ 
              orders: response.data.data,
              loading: false 
            });
          } else {
            throw new Error(response.message || '获取订单列表失败');
          }
        } catch (error) {
          const errorMessage = error instanceof Error ? error.message : '获取订单列表失败';
          set({ error: errorMessage, loading: false });
          message.error(errorMessage);
        }
      },

      fetchOrderById: async (orderId) => {
        set({ loading: true, error: undefined });
        
        try {
          const response = await orderService.getOrderById(orderId);

          if (response.success && response.data) {
            set({ 
              currentOrder: response.data,
              loading: false 
            });
          } else {
            throw new Error(response.message || '获取订单详情失败');
          }
        } catch (error) {
          const errorMessage = error instanceof Error ? error.message : '获取订单详情失败';
          set({ error: errorMessage, loading: false });
          message.error(errorMessage);
        }
      },

      updateOrderStatus: async (orderId, status, notes) => {
        set({ loading: true, error: undefined });
        
        try {
          const response = await orderService.updateOrderStatus(orderId, status, notes);

          if (response.success) {
            // 更新本地订单状态
            const { orders, currentOrder } = get();
            const updatedOrders = orders.map(order =>
              order.id === orderId 
                ? { 
                    ...order, 
                    status, 
                    updatedAt: new Date().toISOString(),
                    ...(status === 'completed' && { completedAt: new Date().toISOString() }),
                    ...(status === 'cancelled' && { cancelledAt: new Date().toISOString() }),
                  }
                : order
            );

            const updatedCurrentOrder = currentOrder?.id === orderId
              ? { 
                  ...currentOrder, 
                  status, 
                  updatedAt: new Date().toISOString(),
                  ...(status === 'completed' && { completedAt: new Date().toISOString() }),
                  ...(status === 'cancelled' && { cancelledAt: new Date().toISOString() }),
                }
              : currentOrder;

            set({ 
              orders: updatedOrders,
              currentOrder: updatedCurrentOrder,
              loading: false 
            });

            message.success(response.message || '订单状态更新成功');
          } else {
            throw new Error(response.message || '订单状态更新失败');
          }
        } catch (error) {
          const errorMessage = error instanceof Error ? error.message : '订单状态更新失败';
          set({ error: errorMessage, loading: false });
          message.error(errorMessage);
        }
      },

      cancelOrder: async (orderId, reason) => {
        set({ loading: true, error: undefined });
        
        try {
          const response = await orderService.cancelOrder(orderId, reason);

          if (response.success) {
            // 更新订单状态为已取消
            await get().updateOrderStatus(orderId, 'cancelled');
            message.success(response.message || '订单取消成功');
          } else {
            throw new Error(response.message || '订单取消失败');
          }
        } catch (error) {
          const errorMessage = error instanceof Error ? error.message : '订单取消失败';
          set({ error: errorMessage, loading: false });
          message.error(errorMessage);
        }
      },

      requestRefund: async (orderId, reason, amount) => {
        set({ loading: true, error: undefined });
        
        try {
          const response = await orderService.requestRefund(orderId, reason, amount);

          if (response.success) {
            // 可以选择更新订单状态或添加退款记录
            message.success(response.message || '退款申请已提交');
            set({ loading: false });
          } else {
            throw new Error(response.message || '退款申请失败');
          }
        } catch (error) {
          const errorMessage = error instanceof Error ? error.message : '退款申请失败';
          set({ error: errorMessage, loading: false });
          message.error(errorMessage);
        }
      },

      processPayment: async (orderId, paymentMethod) => {
        set({ loading: true, error: undefined });
        
        try {
          // 获取订单信息
          const { currentOrder } = get();
          if (!currentOrder || currentOrder.id !== orderId) {
            await get().fetchOrderById(orderId);
          }

          const order = get().currentOrder;
          if (!order) {
            throw new Error('订单不存在');
          }

          // 创建支付
          const paymentResponse = await paymentService.createPayment({
            orderId: order.id,
            amount: order.total,
            currency: order.currency,
            method: paymentMethod,
            description: `订单支付 - ${order.items.map(item => item.plugin.name).join(', ')}`,
          });

          if (paymentResponse.success && paymentResponse.data) {
            const { paymentId, paymentUrl, qrCode } = paymentResponse.data;

            // 根据支付方式处理不同的支付流程
            if (paymentMethod === 'paypal' && paymentUrl) {
              // 跳转到PayPal支付页面
              window.open(paymentUrl, '_blank');
            } else if ((paymentMethod === 'alipay' || paymentMethod === 'wechat') && qrCode) {
              // 显示二维码供用户扫描
              message.info('请使用手机扫描二维码完成支付');
            }

            // 模拟支付成功（在实际应用中，这应该通过回调或轮询来处理）
            setTimeout(async () => {
              await get().updateOrderStatus(orderId, 'completed');
              message.success('支付成功！');
            }, 3000);

            set({ loading: false });
            return true;
          } else {
            throw new Error(paymentResponse.message || '支付创建失败');
          }
        } catch (error) {
          const errorMessage = error instanceof Error ? error.message : '支付处理失败';
          set({ error: errorMessage, loading: false });
          message.error(errorMessage);
          return false;
        }
      },

      checkPaymentStatus: async (paymentId) => {
        try {
          const response = await paymentService.queryPaymentStatus(paymentId);

          if (response.success && response.data) {
            const paymentInfo = response.data;
            
            // 根据支付状态更新订单状态
            if (paymentInfo.status === 'completed') {
              await get().updateOrderStatus(paymentInfo.orderId, 'completed');
            } else if (paymentInfo.status === 'failed') {
              message.error('支付失败，请重试');
            }
          }
        } catch (error) {
          console.error('查询支付状态失败:', error);
        }
      },

      setCurrentOrder: (order) => {
        set({ currentOrder: order });
      },

      clearError: () => {
        set({ error: undefined });
      },

      reset: () => {
        set(initialState);
      },
    }),
    {
      name: 'marketplace-order-store',
      partialize: (state) => ({
        orders: state.orders,
        currentOrder: state.currentOrder,
      }),
    }
  )
);

export default useOrderStore;