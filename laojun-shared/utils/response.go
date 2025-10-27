package utils

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/codetaoist/laojun-shared/models"
)

// SuccessResponse 成功响应
func SuccessResponse(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, models.APIResponse{
		Code:      http.StatusOK,
		Message:   "Success",
		Data:      data,
		Timestamp: time.Now(),
	})
}

// SuccessResponseWithMessage 带消息的成功响应
func SuccessResponseWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, models.APIResponse{
		Code:      http.StatusOK,
		Message:   message,
		Data:      data,
		Timestamp: time.Now(),
	})
}

// ErrorResponse 错误响应
func ErrorResponse(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, models.APIResponse{
		Code:      statusCode,
		Message:   message,
		Timestamp: time.Now(),
	})
}

// BadRequestResponse 400错误响应
func BadRequestResponse(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusBadRequest, message)
}

func UnauthorizedResponse(c *gin.Context) {
	c.JSON(http.StatusUnauthorized, models.APIResponse{
		Code:      http.StatusUnauthorized,
		Message:   "Unauthorized",
		Timestamp: time.Now(),
	})
}

// ForbiddenResponse 403错误响应
func ForbiddenResponse(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusForbidden, message)
}

// NotFoundResponse 404错误响应
func NotFoundResponse(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusNotFound, message)
}

// InternalServerErrorResponse 500错误响应
func InternalServerErrorResponse(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusInternalServerError, message)
}

// PaginatedResponse 分页响应
func PaginatedResponse(c *gin.Context, data interface{}, meta models.PaginationMeta) {
	c.JSON(http.StatusOK, models.PaginatedResponse{
		Data: data,
		Meta: meta,
	})
}

// ValidationErrorResponse 验证错误响应
func ValidationErrorResponse(c *gin.Context, errors interface{}) {
	c.JSON(http.StatusBadRequest, models.APIResponse{
		Code:      http.StatusBadRequest,
		Message:   "Validation failed",
		Data:      errors,
		Timestamp: time.Now(),
	})
}

// CreatedResponse 201创建成功响应
func CreatedResponse(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, models.APIResponse{
		Code:      http.StatusCreated,
		Message:   "Created successfully",
		Data:      data,
		Timestamp: time.Now(),
	})
}

// NotImplementedResponse 501未实现响应
func NotImplementedResponse(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, models.APIResponse{
		Code:      http.StatusNotImplemented,
		Message:   "Not implemented",
		Timestamp: time.Now(),
	})
}

// GetUserIDFromContext 从Gin上下文中获取用户ID并转换为uuid.UUID类型
func GetUserIDFromContext(c *gin.Context) (uuid.UUID, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil, false
	}

	switch v := userID.(type) {
	case uuid.UUID:
		return v, true
	case string:
		if parsed, err := uuid.Parse(v); err == nil {
			return parsed, true
		}
		return uuid.Nil, false
	default:
		return uuid.Nil, false
	}
}
