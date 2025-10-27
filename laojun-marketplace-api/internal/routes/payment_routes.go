package routes

import (
	"github.com/codetaoist/laojun-marketplace-api/internal/handlers"
	"github.com/codetaoist/laojun-marketplace-api/internal/middleware"
	"github.com/codetaoist/laojun-marketplace-api/internal/services"
	"github.com/gin-gonic/gin"
)

// SetupPaymentRoutes 设置支付路由
func SetupPaymentRoutes(router *gin.Engine, handler *handlers.PaymentHandler, authService *services.AuthService) {
	// 需要认证的支付路由
	authAPI := router.Group("/api/v1/payment")
	authAPI.Use(middleware.AuthMiddleware(authService))
	{
		// 订单管理
		orders := authAPI.Group("/orders")
		{
			// 创建支付订单
			orders.POST("", handler.CreatePaymentOrder)
			
			// 获取用户订单列表
			orders.GET("", handler.GetUserOrders)
			
			// 获取单个订单详情
			orders.GET("/:order_id", handler.GetPaymentOrder)
			
			// 取消订单
			orders.POST("/:order_id/cancel", handler.CancelOrder)
			
			// 处理支付回调（这个可能需要特殊处理，因为是第三方调用）
			orders.POST("/:order_id/process", handler.ProcessPayment)
		}
		
		// 退款管理
		authAPI.POST("/refund", handler.RefundOrder)
	}
	
	// 支付回调路由（不需要认证，但需要验证签名）
	callbackAPI := router.Group("/api/v1/payment/callback")
	{
		// 支付宝回调
		callbackAPI.POST("/alipay", handler.ProcessPayment)
		
		// 微信支付回调
		callbackAPI.POST("/wechat", handler.ProcessPayment)
		
		// 银行卡支付回调（Stripe等）
		callbackAPI.POST("/stripe", handler.ProcessPayment)
		
		// 加密货币支付回调
		callbackAPI.POST("/crypto", handler.ProcessPayment)
	}
	
	// 支付状态查询（公开接口，但需要订单号和验证码）
	publicAPI := router.Group("/api/v1/payment/public")
	{
		// 查询订单状态（通过订单号）
		publicAPI.GET("/orders/:order_number/status", handler.GetPaymentOrderByNumber)
	}
}