package services

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	shareddb "github.com/codetaoist/laojun-shared/database"
	"github.com/codetaoist/laojun-shared/logger"
	"github.com/codetaoist/laojun-shared/models"
	"github.com/google/uuid"
)

// PaymentMethod 支付方式枚举
type PaymentMethod string

const (
	PaymentMethodAlipay   PaymentMethod = "alipay"   // 支付宝
	PaymentMethodWechat   PaymentMethod = "wechat"   // 微信支付
	PaymentMethodCard     PaymentMethod = "card"     // 银行卡
	PaymentMethodBalance  PaymentMethod = "balance"  // 余额支付
	PaymentMethodCrypto   PaymentMethod = "crypto"   // 加密货币
)

// PaymentStatus 支付状态枚举
type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "pending"   // 待支付
	PaymentStatusPaid      PaymentStatus = "paid"      // 已支付
	PaymentStatusFailed    PaymentStatus = "failed"    // 支付失败
	PaymentStatusCancelled PaymentStatus = "cancelled" // 已取消
	PaymentStatusRefunded  PaymentStatus = "refunded"  // 已退款
	PaymentStatusPartialRefunded PaymentStatus = "partial_refunded" // 部分退款
)

// OrderStatus 订单状态枚举
type OrderStatus string

const (
	OrderStatusPending    OrderStatus = "pending"    // 待支付
	OrderStatusPaid       OrderStatus = "paid"       // 已支付
	OrderStatusDelivered  OrderStatus = "delivered"  // 已交付
	OrderStatusCompleted  OrderStatus = "completed"  // 已完成
	OrderStatusCancelled  OrderStatus = "cancelled"  // 已取消
	OrderStatusRefunded   OrderStatus = "refunded"   // 已退款
)

// PaymentOrder 支付订单
type PaymentOrder struct {
	ID              uuid.UUID     `json:"id" db:"id"`
	OrderNumber     string        `json:"order_number" db:"order_number"`
	UserID          uuid.UUID     `json:"user_id" db:"user_id"`
	PluginID        uuid.UUID     `json:"plugin_id" db:"plugin_id"`
	Amount          float64       `json:"amount" db:"amount"`
	Currency        string        `json:"currency" db:"currency"`
	PaymentMethod   PaymentMethod `json:"payment_method" db:"payment_method"`
	PaymentStatus   PaymentStatus `json:"payment_status" db:"payment_status"`
	OrderStatus     OrderStatus   `json:"order_status" db:"order_status"`
	PaymentGateway  string        `json:"payment_gateway" db:"payment_gateway"`
	TransactionID   *string       `json:"transaction_id" db:"transaction_id"`
	PaymentURL      *string       `json:"payment_url" db:"payment_url"`
	ExpiresAt       *time.Time    `json:"expires_at" db:"expires_at"`
	PaidAt          *time.Time    `json:"paid_at" db:"paid_at"`
	DeliveredAt     *time.Time    `json:"delivered_at" db:"delivered_at"`
	CompletedAt     *time.Time    `json:"completed_at" db:"completed_at"`
	CancelledAt     *time.Time    `json:"cancelled_at" db:"cancelled_at"`
	RefundedAt      *time.Time    `json:"refunded_at" db:"refunded_at"`
	RefundAmount    *float64      `json:"refund_amount" db:"refund_amount"`
	RefundReason    *string       `json:"refund_reason" db:"refund_reason"`
	Metadata        *string       `json:"metadata" db:"metadata"`
	CreatedAt       time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at" db:"updated_at"`
	
	// 关联数据
	Plugin *models.Plugin `json:"plugin,omitempty"`
	User   *models.User   `json:"user,omitempty"`
}

// PaymentRequest 支付请求
type PaymentRequest struct {
	PluginID      uuid.UUID     `json:"plugin_id" binding:"required"`
	PaymentMethod PaymentMethod `json:"payment_method" binding:"required"`
	Currency      string        `json:"currency"`
	ReturnURL     string        `json:"return_url"`
	CancelURL     string        `json:"cancel_url"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// RefundRequest 退款请求
type RefundRequest struct {
	OrderID      uuid.UUID `json:"order_id" binding:"required"`
	Amount       *float64  `json:"amount"` // 如果为空则全额退款
	Reason       string    `json:"reason" binding:"required"`
	RefundMethod string    `json:"refund_method"` // 退款方式
}

// PaymentService 支付服务
type PaymentService struct {
	db *shareddb.DB
}

// NewPaymentService 创建支付服务
func NewPaymentService(db *shareddb.DB) *PaymentService {
	return &PaymentService{db: db}
}

// CreatePaymentOrder 创建支付订单
func (s *PaymentService) CreatePaymentOrder(userID uuid.UUID, req PaymentRequest) (*PaymentOrder, error) {
	// 检查插件是否存在
	var plugin models.Plugin
	pluginQuery := "SELECT id, name, price, is_active FROM mp_plugins WHERE id = $1"
	err := s.db.QueryRow(pluginQuery, req.PluginID).Scan(
		&plugin.ID, &plugin.Name, &plugin.Price, &plugin.IsActive,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("plugin not found")
		}
		return nil, fmt.Errorf("failed to get plugin: %w", err)
	}

	if !plugin.IsActive {
		return nil, fmt.Errorf("plugin is not active")
	}

	// 检查用户是否已购买
	var exists bool
	checkQuery := "SELECT EXISTS(SELECT 1 FROM mp_purchases WHERE user_id = $1 AND plugin_id = $2)"
	err = s.db.QueryRow(checkQuery, userID, req.PluginID).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("failed to check purchase: %w", err)
	}

	if exists {
		return nil, fmt.Errorf("plugin already purchased")
	}

	// 生成订单号
	orderNumber, err := s.generateOrderNumber()
	if err != nil {
		return nil, fmt.Errorf("failed to generate order number: %w", err)
	}

	// 设置默认货币
	currency := req.Currency
	if currency == "" {
		currency = "CNY"
	}

	// 序列化元数据
	var metadataJSON *string
	if req.Metadata != nil {
		metadataBytes, err := json.Marshal(req.Metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal metadata: %w", err)
		}
		metadataStr := string(metadataBytes)
		metadataJSON = &metadataStr
	}

	// 创建订单
	order := &PaymentOrder{
		ID:            uuid.New(),
		OrderNumber:   orderNumber,
		UserID:        userID,
		PluginID:      req.PluginID,
		Amount:        plugin.Price,
		Currency:      currency,
		PaymentMethod: req.PaymentMethod,
		PaymentStatus: PaymentStatusPending,
		OrderStatus:   OrderStatusPending,
		ExpiresAt:     timePtr(time.Now().Add(30 * time.Minute)), // 30分钟过期
		Metadata:      metadataJSON,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// 根据支付方式设置支付网关
	switch req.PaymentMethod {
	case PaymentMethodAlipay:
		order.PaymentGateway = "alipay"
	case PaymentMethodWechat:
		order.PaymentGateway = "wechat"
	case PaymentMethodCard:
		order.PaymentGateway = "stripe"
	case PaymentMethodBalance:
		order.PaymentGateway = "internal"
	case PaymentMethodCrypto:
		order.PaymentGateway = "coinbase"
	}

	// 插入订单
	insertQuery := `
		INSERT INTO mp_payment_orders 
		(id, order_number, user_id, plugin_id, amount, currency, payment_method, 
		 payment_status, order_status, payment_gateway, expires_at, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`

	_, err = s.db.Exec(insertQuery,
		order.ID, order.OrderNumber, order.UserID, order.PluginID, order.Amount,
		order.Currency, order.PaymentMethod, order.PaymentStatus, order.OrderStatus,
		order.PaymentGateway, order.ExpiresAt, order.Metadata, order.CreatedAt, order.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// 生成支付URL（模拟）
	paymentURL := s.generatePaymentURL(order)
	order.PaymentURL = &paymentURL

	// 更新支付URL
	_, err = s.db.Exec("UPDATE mp_payment_orders SET payment_url = $1 WHERE id = $2", paymentURL, order.ID)
	if err != nil {
		logger.Error("Failed to update payment URL: ", err.Error())
	}

	order.Plugin = &plugin
	return order, nil
}

// ProcessPayment 处理支付
func (s *PaymentService) ProcessPayment(orderID uuid.UUID, transactionID string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 获取订单
	var order PaymentOrder
	orderQuery := `
		SELECT id, user_id, plugin_id, amount, payment_status, order_status
		FROM mp_payment_orders WHERE id = $1`
	
	err = tx.QueryRow(orderQuery, orderID).Scan(
		&order.ID, &order.UserID, &order.PluginID, &order.Amount,
		&order.PaymentStatus, &order.OrderStatus,
	)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}

	if order.PaymentStatus != PaymentStatusPending {
		return fmt.Errorf("order is not in pending status")
	}

	now := time.Now()

	// 更新订单状态
	updateOrderQuery := `
		UPDATE mp_payment_orders 
		SET payment_status = $1, order_status = $2, transaction_id = $3, 
		    paid_at = $4, updated_at = $5
		WHERE id = $6`

	_, err = tx.Exec(updateOrderQuery,
		PaymentStatusPaid, OrderStatusPaid, transactionID, now, now, orderID,
	)
	if err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}

	// 创建购买记录
	purchaseID := uuid.New()
	insertPurchaseQuery := `
		INSERT INTO mp_purchases (id, user_id, plugin_id, amount, status, order_id, created_at) 
		VALUES ($1, $2, $3, $4, 'completed', $5, $6)`

	_, err = tx.Exec(insertPurchaseQuery,
		purchaseID, order.UserID, order.PluginID, order.Amount, orderID, now,
	)
	if err != nil {
		return fmt.Errorf("failed to create purchase: %w", err)
	}

	// 更新插件下载次数
	_, err = tx.Exec("UPDATE mp_plugins SET download_count = download_count + 1 WHERE id = $1", order.PluginID)
	if err != nil {
		return fmt.Errorf("failed to update download count: %w", err)
	}

	return tx.Commit()
}

// GetPaymentOrderByNumber 通过订单号获取支付订单
func (s *PaymentService) GetPaymentOrderByNumber(orderNumber string) (*PaymentOrder, error) {
	var order PaymentOrder
	query := `
		SELECT o.id, o.order_number, o.user_id, o.plugin_id, o.amount, o.currency,
		       o.payment_method, o.payment_status, o.order_status, o.payment_gateway,
		       o.transaction_id, o.payment_url, o.expires_at, o.paid_at, o.delivered_at,
		       o.completed_at, o.cancelled_at, o.refunded_at, o.refund_amount,
		       o.refund_reason, o.metadata, o.created_at, o.updated_at,
		       p.name as plugin_name, p.version as plugin_version, p.icon_url as plugin_icon
		FROM mp_payment_orders o
		LEFT JOIN mp_plugins p ON o.plugin_id = p.id
		WHERE o.order_number = $1`

	var pluginName, pluginVersion sql.NullString
	var pluginIcon sql.NullString

	err := s.db.QueryRow(query, orderNumber).Scan(
		&order.ID, &order.OrderNumber, &order.UserID, &order.PluginID,
		&order.Amount, &order.Currency, &order.PaymentMethod, &order.PaymentStatus,
		&order.OrderStatus, &order.PaymentGateway, &order.TransactionID,
		&order.PaymentURL, &order.ExpiresAt, &order.PaidAt, &order.DeliveredAt,
		&order.CompletedAt, &order.CancelledAt, &order.RefundedAt,
		&order.RefundAmount, &order.RefundReason, &order.Metadata,
		&order.CreatedAt, &order.UpdatedAt,
		&pluginName, &pluginVersion, &pluginIcon,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("order not found")
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	// 设置插件信息
	if pluginName.Valid {
		order.Plugin = &models.Plugin{
			ID:      order.PluginID,
			Name:    pluginName.String,
			Version: pluginVersion.String,
		}
		if pluginIcon.Valid {
			order.Plugin.IconURL = pluginIcon.String
		}
	}

	return &order, nil
}

// GetPaymentOrder 获取支付订单
func (s *PaymentService) GetPaymentOrder(orderID uuid.UUID) (*PaymentOrder, error) {
	var order PaymentOrder
	query := `
		SELECT o.id, o.order_number, o.user_id, o.plugin_id, o.amount, o.currency,
		       o.payment_method, o.payment_status, o.order_status, o.payment_gateway,
		       o.transaction_id, o.payment_url, o.expires_at, o.paid_at, o.delivered_at,
		       o.completed_at, o.cancelled_at, o.refunded_at, o.refund_amount,
		       o.refund_reason, o.metadata, o.created_at, o.updated_at,
		       p.name as plugin_name, p.version as plugin_version, p.icon_url as plugin_icon
		FROM mp_payment_orders o
		LEFT JOIN mp_plugins p ON o.plugin_id = p.id
		WHERE o.id = $1`

	var pluginName, pluginVersion sql.NullString
	var pluginIcon sql.NullString

	err := s.db.QueryRow(query, orderID).Scan(
		&order.ID, &order.OrderNumber, &order.UserID, &order.PluginID,
		&order.Amount, &order.Currency, &order.PaymentMethod, &order.PaymentStatus,
		&order.OrderStatus, &order.PaymentGateway, &order.TransactionID,
		&order.PaymentURL, &order.ExpiresAt, &order.PaidAt, &order.DeliveredAt,
		&order.CompletedAt, &order.CancelledAt, &order.RefundedAt,
		&order.RefundAmount, &order.RefundReason, &order.Metadata,
		&order.CreatedAt, &order.UpdatedAt,
		&pluginName, &pluginVersion, &pluginIcon,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("order not found")
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	// 设置插件信息
	if pluginName.Valid {
		order.Plugin = &models.Plugin{
			ID:      order.PluginID,
			Name:    pluginName.String,
			Version: pluginVersion.String,
		}
		if pluginIcon.Valid {
			order.Plugin.IconURL = pluginIcon.String
		}
	}

	return &order, nil
}

// GetUserOrders 获取用户订单列表
func (s *PaymentService) GetUserOrders(userID uuid.UUID, page, limit int) ([]PaymentOrder, models.PaginationMeta, error) {
	var orders []PaymentOrder
	var totalCount int

	// 获取总数
	countQuery := "SELECT COUNT(*) FROM mp_payment_orders WHERE user_id = $1"
	err := s.db.QueryRow(countQuery, userID).Scan(&totalCount)
	if err != nil {
		return nil, models.PaginationMeta{}, err
	}

	// 计算偏移量
	offset := (page - 1) * limit

	// 获取订单列表
	query := `
		SELECT o.id, o.order_number, o.user_id, o.plugin_id, o.amount, o.currency,
		       o.payment_method, o.payment_status, o.order_status, o.payment_gateway,
		       o.transaction_id, o.payment_url, o.expires_at, o.paid_at, o.delivered_at,
		       o.completed_at, o.cancelled_at, o.refunded_at, o.refund_amount,
		       o.refund_reason, o.metadata, o.created_at, o.updated_at,
		       p.name as plugin_name, p.version as plugin_version, p.icon_url as plugin_icon
		FROM mp_payment_orders o
		LEFT JOIN mp_plugins p ON o.plugin_id = p.id
		WHERE o.user_id = $1
		ORDER BY o.created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := s.db.Query(query, userID, limit, offset)
	if err != nil {
		return nil, models.PaginationMeta{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var order PaymentOrder
		var pluginName, pluginVersion sql.NullString
		var pluginIcon sql.NullString

		err := rows.Scan(
			&order.ID, &order.OrderNumber, &order.UserID, &order.PluginID,
			&order.Amount, &order.Currency, &order.PaymentMethod, &order.PaymentStatus,
			&order.OrderStatus, &order.PaymentGateway, &order.TransactionID,
			&order.PaymentURL, &order.ExpiresAt, &order.PaidAt, &order.DeliveredAt,
			&order.CompletedAt, &order.CancelledAt, &order.RefundedAt,
			&order.RefundAmount, &order.RefundReason, &order.Metadata,
			&order.CreatedAt, &order.UpdatedAt,
			&pluginName, &pluginVersion, &pluginIcon,
		)
		if err != nil {
			return nil, models.PaginationMeta{}, err
		}

		// 设置插件信息
		if pluginName.Valid {
			order.Plugin = &models.Plugin{
				ID:      order.PluginID,
				Name:    pluginName.String,
				Version: pluginVersion.String,
			}
			if pluginIcon.Valid {
				order.Plugin.IconURL = pluginIcon.String
			}
		}

		orders = append(orders, order)
	}

	// 计算分页信息
	totalPages := (totalCount + limit - 1) / limit
	meta := models.PaginationMeta{
		Page:       page,
		Limit:      limit,
		Total:      totalCount,
		TotalPages: totalPages,
	}

	return orders, meta, nil
}

// CancelOrder 取消订单
func (s *PaymentService) CancelOrder(orderID uuid.UUID, reason string) error {
	now := time.Now()
	
	query := `
		UPDATE mp_payment_orders 
		SET order_status = $1, payment_status = $2, cancelled_at = $3, 
		    refund_reason = $4, updated_at = $5
		WHERE id = $6 AND payment_status = $7`

	result, err := s.db.Exec(query,
		OrderStatusCancelled, PaymentStatusCancelled, now, reason, now,
		orderID, PaymentStatusPending,
	)
	if err != nil {
		return fmt.Errorf("failed to cancel order: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("order not found or cannot be cancelled")
	}

	return nil
}

// RefundOrder 退款订单
func (s *PaymentService) RefundOrder(req RefundRequest) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 获取订单信息
	var order PaymentOrder
	orderQuery := `
		SELECT id, amount, payment_status, order_status, refund_amount
		FROM mp_payment_orders WHERE id = $1`
	
	err = tx.QueryRow(orderQuery, req.OrderID).Scan(
		&order.ID, &order.Amount, &order.PaymentStatus, &order.OrderStatus, &order.RefundAmount,
	)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}

	if order.PaymentStatus != PaymentStatusPaid {
		return fmt.Errorf("order is not paid")
	}

	// 计算退款金额
	refundAmount := req.Amount
	if refundAmount == nil {
		// 全额退款
		alreadyRefunded := float64(0)
		if order.RefundAmount != nil {
			alreadyRefunded = *order.RefundAmount
		}
		remainingAmount := order.Amount - alreadyRefunded
		refundAmount = &remainingAmount
	}

	if *refundAmount <= 0 {
		return fmt.Errorf("invalid refund amount")
	}

	// 检查退款金额是否超过可退款金额
	alreadyRefunded := float64(0)
	if order.RefundAmount != nil {
		alreadyRefunded = *order.RefundAmount
	}
	
	if alreadyRefunded + *refundAmount > order.Amount {
		return fmt.Errorf("refund amount exceeds order amount")
	}

	now := time.Now()
	totalRefunded := alreadyRefunded + *refundAmount

	// 确定新的支付状态
	var newPaymentStatus PaymentStatus
	var newOrderStatus OrderStatus
	
	if totalRefunded >= order.Amount {
		// 全额退款
		newPaymentStatus = PaymentStatusRefunded
		newOrderStatus = OrderStatusRefunded
	} else {
		// 部分退款
		newPaymentStatus = PaymentStatusPartialRefunded
		newOrderStatus = OrderStatusRefunded
	}

	// 更新订单
	updateQuery := `
		UPDATE mp_payment_orders 
		SET payment_status = $1, order_status = $2, refund_amount = $3, 
		    refund_reason = $4, refunded_at = $5, updated_at = $6
		WHERE id = $7`

	_, err = tx.Exec(updateQuery,
		newPaymentStatus, newOrderStatus, totalRefunded, req.Reason, now, now, req.OrderID,
	)
	if err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}

	// 如果是全额退款，删除购买记录
	if totalRefunded >= order.Amount {
		_, err = tx.Exec("DELETE FROM mp_purchases WHERE order_id = $1", req.OrderID)
		if err != nil {
			return fmt.Errorf("failed to delete purchase: %w", err)
		}
	}

	return tx.Commit()
}

// generateOrderNumber 生成订单号
func (s *PaymentService) generateOrderNumber() (string, error) {
	// 格式: MP + 年月日 + 6位随机数
	now := time.Now()
	dateStr := now.Format("20060102")
	
	// 生成6位随机数
	randomBytes := make([]byte, 3)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}
	
	randomStr := hex.EncodeToString(randomBytes)
	randomStr = strings.ToUpper(randomStr)
	
	return fmt.Sprintf("MP%s%s", dateStr, randomStr), nil
}

// generatePaymentURL 生成支付URL（模拟）
func (s *PaymentService) generatePaymentURL(order *PaymentOrder) string {
	// 这里应该调用真实的支付网关API生成支付URL
	// 现在只是模拟返回一个URL
	baseURL := "https://pay.laojun.com"
	return fmt.Sprintf("%s/pay/%s?method=%s&amount=%.2f&currency=%s",
		baseURL, order.OrderNumber, order.PaymentMethod, order.Amount, order.Currency)
}

// timePtr 返回时间指针
func timePtr(t time.Time) *time.Time {
	return &t
}