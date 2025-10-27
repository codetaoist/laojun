package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/codetaoist/laojun-marketplace-api/internal/services"
	"github.com/codetaoist/laojun-shared/logger"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// PaymentHandler 支付处理器
type PaymentHandler struct {
	paymentService *services.PaymentService
}

// NewPaymentHandler 创建支付处理器
func NewPaymentHandler(paymentService *services.PaymentService) *PaymentHandler {
	return &PaymentHandler{
		paymentService: paymentService,
	}
}

// CreatePaymentOrder 创建支付订单
// @Summary 创建支付订单
// @Description 为插件购买创建支付订单
// @Tags payment
// @Accept json
// @Produce json
// @Param request body services.PaymentRequest true "支付请求"
// @Success 200 {object} services.PaymentOrder "支付订单"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 401 {object} ErrorResponse "未授权"
// @Failure 500 {object} ErrorResponse "服务器错误"
// @Router /api/v1/payment/orders [post]
func (h *PaymentHandler) CreatePaymentOrder(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Unauthorized",
			Message: "User not authenticated",
		})
		return
	}

	var req services.PaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Failed to bind payment request: ", err.Error())
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request",
			Message: err.Error(),
		})
		return
	}

	// 验证支付方式
	validMethods := map[services.PaymentMethod]bool{
		services.PaymentMethodAlipay:  true,
		services.PaymentMethodWechat:  true,
		services.PaymentMethodCard:    true,
		services.PaymentMethodBalance: true,
		services.PaymentMethodCrypto:  true,
	}

	if !validMethods[req.PaymentMethod] {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid payment method",
			Message: "Unsupported payment method",
		})
		return
	}

	order, err := h.paymentService.CreatePaymentOrder(userID.(uuid.UUID), req)
	if err != nil {
		logger.Error("Failed to create payment order: ", err.Error())
		
		// 根据错误类型返回不同的状态码
		if err.Error() == "plugin not found" {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "Plugin not found",
				Message: "The specified plugin does not exist",
			})
			return
		}
		
		if err.Error() == "plugin is not active" {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "Plugin not available",
				Message: "The plugin is not currently available for purchase",
			})
			return
		}
		
		if err.Error() == "plugin already purchased" {
			c.JSON(http.StatusConflict, ErrorResponse{
				Error:   "Already purchased",
				Message: "You have already purchased this plugin",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to create order",
			Message: "Unable to create payment order",
		})
		return
	}

	c.JSON(http.StatusOK, order)
}

// GetPaymentOrder 获取支付订单
// @Summary 获取支付订单
// @Description 根据订单ID获取支付订单详情
// @Tags payment
// @Accept json
// @Produce json
// @Param order_id path string true "订单ID"
// @Success 200 {object} services.PaymentOrder "支付订单"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 401 {object} ErrorResponse "未授权"
// @Failure 404 {object} ErrorResponse "订单不存在"
// @Failure 500 {object} ErrorResponse "服务器错误"
// @Router /api/v1/payment/orders/{order_id} [get]
func (h *PaymentHandler) GetPaymentOrder(c *gin.Context) {
	orderIDStr := c.Param("order_id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid order ID",
			Message: "Order ID must be a valid UUID",
		})
		return
	}

	order, err := h.paymentService.GetPaymentOrder(orderID)
	if err != nil {
		logger.Error("Failed to get payment order: ", err.Error())
		
		if err.Error() == "order not found" {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "Order not found",
				Message: "The specified order does not exist",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to get order",
			Message: "Unable to retrieve payment order",
		})
		return
	}

	// 检查用户权限（只能查看自己的订单）
	userID, exists := c.Get("user_id")
	if !exists || userID.(uuid.UUID) != order.UserID {
		c.JSON(http.StatusForbidden, ErrorResponse{
			Error:   "Access denied",
			Message: "You can only view your own orders",
		})
		return
	}

	c.JSON(http.StatusOK, order)
}

// GetUserOrders 获取用户订单列表
// @Summary 获取用户订单列表
// @Description 获取当前用户的所有订单
// @Tags payment
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param limit query int false "每页数量" default(20)
// @Success 200 {object} ListResponse "订单列表"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 401 {object} ErrorResponse "未授权"
// @Failure 500 {object} ErrorResponse "服务器错误"
// @Router /api/v1/payment/orders [get]
func (h *PaymentHandler) GetUserOrders(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Unauthorized",
			Message: "User not authenticated",
		})
		return
	}

	// 解析分页参数
	page := 1
	limit := 20

	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	orders, meta, err := h.paymentService.GetUserOrders(userID.(uuid.UUID), page, limit)
	if err != nil {
		logger.Error("Failed to get user orders: ", err.Error())
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to get orders",
			Message: "Unable to retrieve user orders",
		})
		return
	}

	c.JSON(http.StatusOK, ListResponse{
		Data: orders,
		Meta: meta,
	})
}

// ProcessPayment 处理支付回调
// @Summary 处理支付回调
// @Description 处理第三方支付平台的支付回调
// @Tags payment
// @Accept json
// @Produce json
// @Param order_id path string true "订单ID"
// @Param request body ProcessPaymentRequest true "支付回调数据"
// @Success 200 {object} SuccessResponse "处理成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 404 {object} ErrorResponse "订单不存在"
// @Failure 500 {object} ErrorResponse "服务器错误"
// @Router /api/v1/payment/orders/{order_id}/process [post]
func (h *PaymentHandler) ProcessPayment(c *gin.Context) {
	orderIDStr := c.Param("order_id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid order ID",
			Message: "Order ID must be a valid UUID",
		})
		return
	}

	var req ProcessPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Failed to bind process payment request: ", err.Error())
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request",
			Message: err.Error(),
		})
		return
	}

	err = h.paymentService.ProcessPayment(orderID, req.TransactionID)
	if err != nil {
		logger.Error("Failed to process payment: ", err.Error())
		
		if err.Error() == "failed to get order: sql: no rows in result set" {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "Order not found",
				Message: "The specified order does not exist",
			})
			return
		}
		
		if err.Error() == "order is not in pending status" {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "Invalid order status",
				Message: "Order cannot be processed in current status",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to process payment",
			Message: "Unable to process payment",
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Payment processed successfully",
	})
}

// CancelOrder 取消订单
// @Summary 取消订单
// @Description 取消待支付的订单
// @Tags payment
// @Accept json
// @Produce json
// @Param order_id path string true "订单ID"
// @Param request body CancelOrderRequest true "取消原因"
// @Success 200 {object} SuccessResponse "取消成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 401 {object} ErrorResponse "未授权"
// @Failure 404 {object} ErrorResponse "订单不存在"
// @Failure 500 {object} ErrorResponse "服务器错误"
// @Router /api/v1/payment/orders/{order_id}/cancel [post]
func (h *PaymentHandler) CancelOrder(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Unauthorized",
			Message: "User not authenticated",
		})
		return
	}

	orderIDStr := c.Param("order_id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid order ID",
			Message: "Order ID must be a valid UUID",
		})
		return
	}

	var req CancelOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Failed to bind cancel order request: ", err.Error())
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request",
			Message: err.Error(),
		})
		return
	}

	// 验证用户权限
	order, err := h.paymentService.GetPaymentOrder(orderID)
	if err != nil {
		if err.Error() == "order not found" {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "Order not found",
				Message: "The specified order does not exist",
			})
			return
		}
		
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to get order",
			Message: "Unable to retrieve order information",
		})
		return
	}

	if order.UserID != userID.(uuid.UUID) {
		c.JSON(http.StatusForbidden, ErrorResponse{
			Error:   "Access denied",
			Message: "You can only cancel your own orders",
		})
		return
	}

	err = h.paymentService.CancelOrder(orderID, req.Reason)
	if err != nil {
		logger.Error("Failed to cancel order: ", err.Error())
		
		if err.Error() == "order not found or cannot be cancelled" {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "Cannot cancel order",
				Message: "Order cannot be cancelled in current status",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to cancel order",
			Message: "Unable to cancel order",
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Order cancelled successfully",
	})
}

// RefundOrder 申请退款
// @Summary 申请退款
// @Description 为已支付的订单申请退款
// @Tags payment
// @Accept json
// @Produce json
// @Param request body services.RefundRequest true "退款请求"
// @Success 200 {object} SuccessResponse "退款申请成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 401 {object} ErrorResponse "未授权"
// @Failure 404 {object} ErrorResponse "订单不存在"
// @Failure 500 {object} ErrorResponse "服务器错误"
// @Router /api/v1/payment/refund [post]
func (h *PaymentHandler) RefundOrder(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Unauthorized",
			Message: "User not authenticated",
		})
		return
	}

	var req services.RefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Failed to bind refund request: ", err.Error())
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request",
			Message: err.Error(),
		})
		return
	}

	// 验证用户权限
	order, err := h.paymentService.GetPaymentOrder(req.OrderID)
	if err != nil {
		if err.Error() == "order not found" {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "Order not found",
				Message: "The specified order does not exist",
			})
			return
		}
		
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to get order",
			Message: "Unable to retrieve order information",
		})
		return
	}

	if order.UserID != userID.(uuid.UUID) {
		c.JSON(http.StatusForbidden, ErrorResponse{
			Error:   "Access denied",
			Message: "You can only refund your own orders",
		})
		return
	}

	err = h.paymentService.RefundOrder(req)
	if err != nil {
		logger.Error("Failed to refund order: ", err.Error())
		
		if err.Error() == "order is not paid" {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "Cannot refund order",
				Message: "Order is not in paid status",
			})
			return
		}
		
		if err.Error() == "invalid refund amount" {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "Invalid refund amount",
				Message: "Refund amount must be greater than 0",
			})
			return
		}
		
		if err.Error() == "refund amount exceeds order amount" {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "Invalid refund amount",
				Message: "Refund amount exceeds available amount",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to process refund",
			Message: "Unable to process refund request",
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Refund processed successfully",
	})
}

// ProcessPaymentRequest 处理支付请求结构
type ProcessPaymentRequest struct {
	TransactionID string `json:"transaction_id" binding:"required"`
	Gateway       string `json:"gateway"`
	Status        string `json:"status"`
	Signature     string `json:"signature"`
}

// CancelOrderRequest 取消订单请求结构
type CancelOrderRequest struct {
	Reason string `json:"reason" binding:"required"`
}

// GetPaymentOrderByNumber 通过订单号获取支付订单状态
// @Summary 通过订单号获取支付订单状态
// @Description 通过订单号查询支付订单的状态信息（公开接口）
// @Tags payment
// @Accept json
// @Produce json
// @Param order_number path string true "订单号"
// @Success 200 {object} OrderStatusResponse "订单状态"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 404 {object} ErrorResponse "订单不存在"
// @Failure 500 {object} ErrorResponse "服务器错误"
// @Router /api/v1/payment/public/orders/{order_number}/status [get]
func (h *PaymentHandler) GetPaymentOrderByNumber(c *gin.Context) {
	orderNumber := c.Param("order_number")
	if orderNumber == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid order number",
			Message: "Order number is required",
		})
		return
	}

	order, err := h.paymentService.GetPaymentOrderByNumber(orderNumber)
	if err != nil {
		logger.Error("Failed to get payment order by number: ", err.Error())
		
		if err.Error() == "order not found" {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "Order not found",
				Message: "The specified order does not exist",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to get order",
			Message: "Unable to retrieve payment order",
		})
		return
	}

	// 只返回状态信息，不返回敏感数据
	response := OrderStatusResponse{
		OrderNumber:   order.OrderNumber,
		PaymentStatus: string(order.PaymentStatus),
		OrderStatus:   string(order.OrderStatus),
		Amount:        order.Amount,
		Currency:      order.Currency,
		CreatedAt:     order.CreatedAt,
		PaidAt:        order.PaidAt,
		ExpiresAt:     order.ExpiresAt,
	}

	if order.Plugin != nil {
		response.PluginName = order.Plugin.Name
	}

	c.JSON(http.StatusOK, response)
}

// OrderStatusResponse 订单状态响应
type OrderStatusResponse struct {
	OrderNumber   string     `json:"order_number"`
	PaymentStatus string     `json:"payment_status"`
	OrderStatus   string     `json:"order_status"`
	Amount        float64    `json:"amount"`
	Currency      string     `json:"currency"`
	PluginName    string     `json:"plugin_name,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	PaidAt        *time.Time `json:"paid_at,omitempty"`
	ExpiresAt     *time.Time `json:"expires_at,omitempty"`
}