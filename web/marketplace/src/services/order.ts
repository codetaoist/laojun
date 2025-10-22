import { 
  Order, 
  OrderItem, 
  OrderStatus, 
  PaymentStatus, 
  PaymentMethod, 
  BillingInfo,
  CartItem,
  ApiResponse,
  PaginatedResponse 
} from '@/types';

// 模拟API基础URL
const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api';

class OrderService {
  // 创建订单
  async createOrder(
    cartItems: CartItem[],
    billingInfo: BillingInfo,
    paymentMethod: PaymentMethod,
    userId: string
  ): Promise<ApiResponse<Order>> {
    try {
      // 计算订单金额
      const subtotal = cartItems.reduce((sum, item) => sum + (item.plugin.price * item.quantity), 0);
      const tax = subtotal * 0.1; // 10% 税率
      const discount = 0; // 暂时没有折扣
      const total = subtotal + tax - discount;

      // 创建订单项目
      const orderItems: OrderItem[] = cartItems.map((item, index) => ({
        id: `item-${Date.now()}-${index}`,
        pluginId: item.pluginId,
        plugin: item.plugin,
        quantity: item.quantity,
        unitPrice: item.plugin.price,
        totalPrice: item.plugin.price * item.quantity,
      }));

      // 创建订单对象
      const order: Order = {
        id: `ORDER-${Date.now()}`,
        userId,
        items: orderItems,
        subtotal,
        tax,
        discount,
        total,
        currency: 'CNY',
        status: 'pending',
        paymentStatus: 'pending',
        paymentMethod,
        billingInfo,
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      };

      // 模拟API调用
      await new Promise(resolve => setTimeout(resolve, 1000));

      // 在实际应用中，这里应该调用真实的API
      // const response = await fetch(`${API_BASE_URL}/orders`, {
      //   method: 'POST',
      //   headers: {
      //     'Content-Type': 'application/json',
      //     'Authorization': `Bearer ${getAuthToken()}`,
      //   },
      //   body: JSON.stringify(order),
      // });

      return {
        success: true,
        data: order,
        message: '订单创建成功',
      };
    } catch (error) {
      console.error('创建订单失败:', error);
      return {
        success: false,
        message: '创建订单失败，请重试',
      };
    }
  }

  // 获取用户订单列表
  async getUserOrders(
    userId: string,
    page: number = 1,
    pageSize: number = 10,
    status?: OrderStatus
  ): Promise<ApiResponse<PaginatedResponse<Order>>> {
    try {
      // 模拟API调用
      await new Promise(resolve => setTimeout(resolve, 500));

      // 模拟订单数据
      const mockOrders: Order[] = [
        {
          id: 'ORDER-1703123456789',
          userId,
          items: [
            {
              id: 'item-1',
              pluginId: 'plugin-1',
              plugin: {
                id: 'plugin-1',
                name: 'Code Formatter Pro',
                description: '专业的代码格式化工具',
                version: '2.1.0',
                author: 'DevTools Inc',
                category: 'development',
                tags: ['formatter', 'code', 'productivity'],
                icon: 'https://via.placeholder.com/64',
                downloadUrl: '',
                license: 'MIT',
                price: 29.99,
                currency: 'CNY',
                downloads: 15420,
                rating: 4.8,
                reviewCount: 234,
                size: 2048000,
                requirements: { minVersion: '1.0.0' },
                status: 'active',
                featured: true,
                createdAt: '2023-12-01T00:00:00Z',
                updatedAt: '2023-12-15T00:00:00Z',
              },
              quantity: 1,
              unitPrice: 29.99,
              totalPrice: 29.99,
            },
          ],
          subtotal: 29.99,
          tax: 3.00,
          discount: 0,
          total: 32.99,
          currency: 'CNY',
          status: 'completed',
          paymentStatus: 'completed',
          paymentMethod: 'alipay',
          billingInfo: {
            firstName: '张',
            lastName: '三',
            email: 'zhangsan@example.com',
            phone: '13800138000',
            address: '北京市朝阳区某某街道123号',
            city: '北京',
            state: '北京市',
            zipCode: '100000',
            country: 'CN',
          },
          createdAt: '2023-12-21T10:30:00Z',
          updatedAt: '2023-12-21T10:35:00Z',
          completedAt: '2023-12-21T10:35:00Z',
        },
        {
          id: 'ORDER-1703023456789',
          userId,
          items: [
            {
              id: 'item-2',
              pluginId: 'plugin-2',
              plugin: {
                id: 'plugin-2',
                name: 'Theme Designer',
                description: '强大的主题设计工具',
                version: '1.5.2',
                author: 'UI Masters',
                category: 'design',
                tags: ['theme', 'design', 'ui'],
                icon: 'https://via.placeholder.com/64',
                downloadUrl: '',
                license: 'Commercial',
                price: 49.99,
                currency: 'CNY',
                downloads: 8930,
                rating: 4.6,
                reviewCount: 156,
                size: 5120000,
                requirements: { minVersion: '1.2.0' },
                status: 'active',
                featured: false,
                createdAt: '2023-11-15T00:00:00Z',
                updatedAt: '2023-12-10T00:00:00Z',
              },
              quantity: 1,
              unitPrice: 49.99,
              totalPrice: 49.99,
            },
          ],
          subtotal: 49.99,
          tax: 5.00,
          discount: 0,
          total: 54.99,
          currency: 'CNY',
          status: 'processing',
          paymentStatus: 'completed',
          paymentMethod: 'wechat',
          billingInfo: {
            firstName: '张',
            lastName: '三',
            email: 'zhangsan@example.com',
            phone: '13800138000',
            address: '北京市朝阳区某某街道123号',
            city: '北京',
            state: '北京市',
            zipCode: '100000',
            country: 'CN',
          },
          createdAt: '2023-12-20T15:20:00Z',
          updatedAt: '2023-12-20T15:25:00Z',
        },
      ];

      // 根据状态过滤
      let filteredOrders = mockOrders;
      if (status) {
        filteredOrders = mockOrders.filter(order => order.status === status);
      }

      // 分页
      const startIndex = (page - 1) * pageSize;
      const endIndex = startIndex + pageSize;
      const paginatedOrders = filteredOrders.slice(startIndex, endIndex);

      return {
        success: true,
        data: {
          data: paginatedOrders,
          total: filteredOrders.length,
          page,
          pageSize,
          totalPages: Math.ceil(filteredOrders.length / pageSize),
        },
        message: '获取订单列表成功',
      };
    } catch (error) {
      console.error('获取订单列表失败:', error);
      return {
        success: false,
        message: '获取订单列表失败，请重试',
      };
    }
  }

  // 获取订单详情
  async getOrderById(orderId: string): Promise<ApiResponse<Order>> {
    try {
      // 模拟API调用
      await new Promise(resolve => setTimeout(resolve, 300));

      // 在实际应用中，这里应该调用真实的API
      // const response = await fetch(`${API_BASE_URL}/orders/${orderId}`, {
      //   headers: {
      //     'Authorization': `Bearer ${getAuthToken()}`,
      //   },
      // });

      // 模拟返回订单数据
      const mockOrder: Order = {
        id: orderId,
        userId: 'user-123',
        items: [
          {
            id: 'item-1',
            pluginId: 'plugin-1',
            plugin: {
              id: 'plugin-1',
              name: 'Code Formatter Pro',
              description: '专业的代码格式化工具',
              version: '2.1.0',
              author: 'DevTools Inc',
              category: 'development',
              tags: ['formatter', 'code', 'productivity'],
              icon: 'https://via.placeholder.com/64',
              downloadUrl: '',
              license: 'MIT',
              price: 29.99,
              currency: 'CNY',
              downloads: 15420,
              rating: 4.8,
              reviewCount: 234,
              size: 2048000,
              requirements: { minVersion: '1.0.0' },
              status: 'active',
              featured: true,
              createdAt: '2023-12-01T00:00:00Z',
              updatedAt: '2023-12-15T00:00:00Z',
            },
            quantity: 1,
            unitPrice: 29.99,
            totalPrice: 29.99,
          },
        ],
        subtotal: 29.99,
        tax: 3.00,
        discount: 0,
        total: 32.99,
        currency: 'CNY',
        status: 'completed',
        paymentStatus: 'completed',
        paymentMethod: 'alipay',
        billingInfo: {
          firstName: '张',
          lastName: '三',
          email: 'zhangsan@example.com',
          phone: '13800138000',
          address: '北京市朝阳区某某街道123号',
          city: '北京',
          state: '北京市',
          zipCode: '100000',
          country: 'CN',
        },
        createdAt: '2023-12-21T10:30:00Z',
        updatedAt: '2023-12-21T10:35:00Z',
        completedAt: '2023-12-21T10:35:00Z',
      };

      return {
        success: true,
        data: mockOrder,
        message: '获取订单详情成功',
      };
    } catch (error) {
      console.error('获取订单详情失败:', error);
      return {
        success: false,
        message: '获取订单详情失败，请重试',
      };
    }
  }

  // 更新订单状态
  async updateOrderStatus(
    orderId: string,
    status: OrderStatus,
    notes?: string
  ): Promise<ApiResponse<Order>> {
    try {
      // 模拟API调用
      await new Promise(resolve => setTimeout(resolve, 500));

      // 在实际应用中，这里应该调用真实的API
      // const response = await fetch(`${API_BASE_URL}/orders/${orderId}/status`, {
      //   method: 'PATCH',
      //   headers: {
      //     'Content-Type': 'application/json',
      //     'Authorization': `Bearer ${getAuthToken()}`,
      //   },
      //   body: JSON.stringify({ status, notes }),
      // });

      return {
        success: true,
        message: '订单状态更新成功',
      };
    } catch (error) {
      console.error('更新订单状态失败:', error);
      return {
        success: false,
        message: '更新订单状态失败，请重试',
      };
    }
  }

  // 取消订单
  async cancelOrder(orderId: string, reason?: string): Promise<ApiResponse<void>> {
    try {
      // 模拟API调用
      await new Promise(resolve => setTimeout(resolve, 500));

      // 在实际应用中，这里应该调用真实的API
      // const response = await fetch(`${API_BASE_URL}/orders/${orderId}/cancel`, {
      //   method: 'POST',
      //   headers: {
      //     'Content-Type': 'application/json',
      //     'Authorization': `Bearer ${getAuthToken()}`,
      //   },
      //   body: JSON.stringify({ reason }),
      // });

      return {
        success: true,
        message: '订单取消成功',
      };
    } catch (error) {
      console.error('取消订单失败:', error);
      return {
        success: false,
        message: '取消订单失败，请重试',
      };
    }
  }

  // 申请退款
  async requestRefund(
    orderId: string,
    reason: string,
    amount?: number
  ): Promise<ApiResponse<void>> {
    try {
      // 模拟API调用
      await new Promise(resolve => setTimeout(resolve, 1000));

      // 在实际应用中，这里应该调用真实的API
      // const response = await fetch(`${API_BASE_URL}/orders/${orderId}/refund`, {
      //   method: 'POST',
      //   headers: {
      //     'Content-Type': 'application/json',
      //     'Authorization': `Bearer ${getAuthToken()}`,
      //   },
      //   body: JSON.stringify({ reason, amount }),
      // });

      return {
        success: true,
        message: '退款申请已提交，我们将在3-5个工作日内处理',
      };
    } catch (error) {
      console.error('申请退款失败:', error);
      return {
        success: false,
        message: '申请退款失败，请重试',
      };
    }
  }
}

// 创建单例实例
export const orderService = new OrderService();
export default orderService;