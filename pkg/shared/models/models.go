package models

import (
	"time"

	"github.com/google/uuid"
)

// User 用户模型
type User struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	Username     string     `json:"username" db:"username"`
	Email        string     `json:"email" db:"email"`
	PasswordHash string     `json:"-" db:"password_hash"`
	Avatar       *string    `json:"avatar" db:"avatar"`
	Bio          *string    `json:"bio" db:"bio"`
	IsActive     bool       `json:"is_active" db:"is_active"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
	LastLoginAt  *time.Time `json:"last_login_at" db:"last_login_at"`
}

// Role 角色模型
type Role struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	DisplayName string    `json:"display_name" db:"display_name"`
	Description *string   `json:"description" db:"description"`
	IsSystem    bool      `json:"is_system" db:"is_system"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// Permission 权限模型
type Permission struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Code        string    `json:"code" db:"code"`
	Description *string   `json:"description" db:"description"`
	Resource    string    `json:"resource" db:"resource"`
	Action      string    `json:"action" db:"action"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// Category 插件分类模型
type Category struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description *string   `json:"description" db:"description"`
	Icon        *string   `json:"icon" db:"icon"`
	Color       string    `json:"color" db:"color"`
	SortOrder   int       `json:"sort_order" db:"sort_order"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// Plugin 插件模型
type Plugin struct {
	ID               uuid.UUID  `json:"id" db:"id"`
	Name             string     `json:"name" db:"name"`
	Description      *string    `json:"description" db:"description"`
	ShortDescription *string    `json:"short_description" db:"short_description"`
	Author           string     `json:"author" db:"author"`
	DeveloperID      *uuid.UUID `json:"developer_id" db:"developer_id"`
	Version          string     `json:"version" db:"version"`
	CategoryID       *uuid.UUID `json:"category_id" db:"category_id"`
	Price            float64    `json:"price" db:"price"`
	IsFree           bool       `json:"is_free" db:"is_free"`
	IsFeatured       bool       `json:"is_featured" db:"is_featured"`
	IsActive         bool       `json:"is_active" db:"is_active"`
	DownloadCount    int        `json:"download_count" db:"download_count"`
	Rating           float64    `json:"rating" db:"rating"`
	ReviewCount      int        `json:"review_count" db:"review_count"`
	IconURL          *string    `json:"icon_url" db:"icon_url"`
	BannerURL        *string    `json:"banner_url" db:"banner_url"`
	Screenshots      []string   `json:"screenshots" db:"screenshots"`
	Tags             []string   `json:"tags" db:"tags"`
	Requirements     *string    `json:"requirements" db:"requirements"`
	Changelog        *string    `json:"changelog" db:"changelog"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`

	// 关联数据
	Category *Category `json:"category,omitempty"`
}

// PluginWithCategory 带分类信息的插件
type PluginWithCategory struct {
	Plugin
	CategoryName  *string `json:"category_name" db:"category_name"`
	CategoryIcon  *string `json:"category_icon" db:"category_icon"`
	CategoryColor *string `json:"category_color" db:"category_color"`
}

// Review 评论模型
type Review struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	PluginID     uuid.UUID  `json:"plugin_id" db:"plugin_id"`
	UserID       *uuid.UUID `json:"user_id" db:"user_id"`
	Rating       int        `json:"rating" db:"rating"`
	Title        *string    `json:"title" db:"title"`
	Content      *string    `json:"content" db:"content"`
	IsVerified   bool       `json:"is_verified" db:"is_verified"`
	HelpfulCount int        `json:"helpful_count" db:"helpful_count"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`

	// 关联数据
	User   *User   `json:"user,omitempty"`
	Plugin *Plugin `json:"plugin,omitempty"`
}

// Purchase 购买记录模型
type Purchase struct {
	ID            uuid.UUID `json:"id" db:"id"`
	UserID        uuid.UUID `json:"user_id" db:"user_id"`
	PluginID      uuid.UUID `json:"plugin_id" db:"plugin_id"`
	Amount        float64   `json:"amount" db:"amount"`
	Status        string    `json:"status" db:"status"`
	PaymentMethod *string   `json:"payment_method" db:"payment_method"`
	TransactionID *string   `json:"transaction_id" db:"transaction_id"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`

	// 关联数据
	User   *User   `json:"user,omitempty"`
	Plugin *Plugin `json:"plugin,omitempty"`
}

// APIResponse 统一API响应格式
type APIResponse struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// PaginationMeta 分页元数据
type PaginationMeta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// ========== 社区相关模型 ==========

// ForumCategory 论坛分类模型
type ForumCategory struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description *string   `json:"description" db:"description"`
	Icon        *string   `json:"icon" db:"icon"`
	SortOrder   int       `json:"sort_order" db:"sort_order"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// ForumPost 论坛帖子模型
type ForumPost struct {
	ID           uuid.UUID `json:"id" db:"id"`
	CategoryID   uuid.UUID `json:"category_id" db:"category_id"`
	UserID       uuid.UUID `json:"user_id" db:"user_id"`
	Title        string    `json:"title" db:"title"`
	Content      string    `json:"content" db:"content"`
	LikesCount   int       `json:"likes_count" db:"likes_count"`
	RepliesCount int       `json:"replies_count" db:"replies_count"`
	ViewsCount   int       `json:"views_count" db:"views_count"`
	IsPinned     bool      `json:"is_pinned" db:"is_pinned"`
	IsLocked     bool      `json:"is_locked" db:"is_locked"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`

	// 关联数据
	Username     string  `json:"username,omitempty" db:"username"`
	AvatarURL    *string `json:"avatar_url,omitempty" db:"avatar_url"`
	CategoryName string  `json:"category_name,omitempty" db:"category_name"`
}

// ForumReply 论坛回复模型
type ForumReply struct {
	ID         uuid.UUID `json:"id" db:"id"`
	PostID     uuid.UUID `json:"post_id" db:"post_id"`
	UserID     uuid.UUID `json:"user_id" db:"user_id"`
	Content    string    `json:"content" db:"content"`
	LikesCount int       `json:"likes_count" db:"likes_count"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`

	// 关联数据
	Username  string  `json:"username,omitempty" db:"username"`
	AvatarURL *string `json:"avatar_url,omitempty" db:"avatar_url"`
}

// BlogCategory 博客分类模型
type BlogCategory struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description *string   `json:"description" db:"description"`
	Icon        *string   `json:"icon" db:"icon"`
	SortOrder   int       `json:"sort_order" db:"sort_order"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// BlogPost 博客文章模型
type BlogPost struct {
	ID            uuid.UUID `json:"id" db:"id"`
	CategoryID    uuid.UUID `json:"category_id" db:"category_id"`
	UserID        uuid.UUID `json:"user_id" db:"user_id"`
	Title         string    `json:"title" db:"title"`
	Content       string    `json:"content" db:"content"`
	Summary       *string   `json:"summary" db:"summary"`
	CoverImage    *string   `json:"cover_image" db:"cover_image"`
	Tags          *string   `json:"tags" db:"tags"`
	LikesCount    int       `json:"likes_count" db:"likes_count"`
	CommentsCount int       `json:"comments_count" db:"comments_count"`
	ViewsCount    int       `json:"views_count" db:"views_count"`
	IsPublished   bool      `json:"is_published" db:"is_published"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`

	// 关联数据
	Username     string  `json:"username,omitempty" db:"username"`
	AvatarURL    *string `json:"avatar_url,omitempty" db:"avatar_url"`
	CategoryName string  `json:"category_name,omitempty" db:"category_name"`
}

// BlogComment 博客评论模型
type BlogComment struct {
	ID         uuid.UUID `json:"id" db:"id"`
	PostID     uuid.UUID `json:"post_id" db:"post_id"`
	UserID     uuid.UUID `json:"user_id" db:"user_id"`
	Content    string    `json:"content" db:"content"`
	LikesCount int       `json:"likes_count" db:"likes_count"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`

	// 关联数据
	Username  string  `json:"username,omitempty" db:"username"`
	AvatarURL *string `json:"avatar_url,omitempty" db:"avatar_url"`
}

// CodeSnippet 代码片段模型
type CodeSnippet struct {
	ID          uuid.UUID `json:"id" db:"id"`
	UserID      uuid.UUID `json:"user_id" db:"user_id"`
	Title       string    `json:"title" db:"title"`
	Description *string   `json:"description" db:"description"`
	Code        string    `json:"code" db:"code"`
	Language    string    `json:"language" db:"language"`
	Tags        *string   `json:"tags" db:"tags"`
	LikesCount  int       `json:"likes_count" db:"likes_count"`
	ViewsCount  int       `json:"views_count" db:"views_count"`
	IsPublic    bool      `json:"is_public" db:"is_public"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`

	// 关联数据
	Username  string  `json:"username,omitempty" db:"username"`
	AvatarURL *string `json:"avatar_url,omitempty" db:"avatar_url"`
}

// Like 点赞模型
type Like struct {
	ID         uuid.UUID `json:"id" db:"id"`
	UserID     uuid.UUID `json:"user_id" db:"user_id"`
	TargetType string    `json:"target_type" db:"target_type"` // forum_post, forum_reply, blog_post, code_snippet
	TargetID   uuid.UUID `json:"target_id" db:"target_id"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// UserFollow 用户关注模型
type UserFollow struct {
	ID          uuid.UUID `json:"id" db:"id"`
	FollowerID  uuid.UUID `json:"follower_id" db:"follower_id"`
	FollowingID uuid.UUID `json:"following_id" db:"following_id"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// Bookmark 收藏模型
type Bookmark struct {
	ID         uuid.UUID `json:"id" db:"id"`
	UserID     uuid.UUID `json:"user_id" db:"user_id"`
	TargetType string    `json:"target_type" db:"target_type"` // forum_post, blog_post, code_snippet
	TargetID   uuid.UUID `json:"target_id" db:"target_id"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// UserPoints 用户积分模型
type UserPoints struct {
	ID           uuid.UUID `json:"id" db:"id"`
	UserID       uuid.UUID `json:"user_id" db:"user_id"`
	TotalPoints  int       `json:"total_points" db:"total_points"`
	Level        int       `json:"level" db:"level"`
	LevelName    string    `json:"level_name" db:"level_name"`
	NextLevelExp int       `json:"next_level_exp" db:"next_level_exp"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// PointRecord 积分记录模型
type PointRecord struct {
	ID          uuid.UUID `json:"id" db:"id"`
	UserID      uuid.UUID `json:"user_id" db:"user_id"`
	Points      int       `json:"points" db:"points"`
	Action      string    `json:"action" db:"action"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// Badge 徽章模型
type Badge struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Icon        string    `json:"icon" db:"icon"`
	Color       string    `json:"color" db:"color"`
	Condition   string    `json:"condition" db:"condition"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// UserBadge 用户徽章模型
type UserBadge struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	BadgeID   uuid.UUID `json:"badge_id" db:"badge_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`

	// 关联数据
	Badge *Badge `json:"badge,omitempty"`
}

// Message 私信模型
type Message struct {
	ID         uuid.UUID `json:"id" db:"id"`
	SenderID   uuid.UUID `json:"sender_id" db:"sender_id"`
	ReceiverID uuid.UUID `json:"receiver_id" db:"receiver_id"`
	Title      string    `json:"title" db:"title"`
	Content    string    `json:"content" db:"content"`
	IsRead     bool      `json:"is_read" db:"is_read"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`

	// 关联数据
	SenderUsername   string  `json:"sender_username,omitempty" db:"sender_username"`
	SenderAvatarURL  *string `json:"sender_avatar_url,omitempty" db:"sender_avatar_url"`
	ReceiverUsername string  `json:"receiver_username,omitempty" db:"receiver_username"`
}

// Notification 通知模型
type Notification struct {
	ID         uuid.UUID  `json:"id" db:"id"`
	UserID     uuid.UUID  `json:"user_id" db:"user_id"`
	Type       string     `json:"type" db:"type"`
	Title      string     `json:"title" db:"title"`
	Content    string     `json:"content" db:"content"`
	TargetType *string    `json:"target_type" db:"target_type"`
	TargetID   *uuid.UUID `json:"target_id" db:"target_id"`
	IsRead     bool       `json:"is_read" db:"is_read"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
}

// PaginatedResponse 分页响应
type PaginatedResponse struct {
	Data interface{}    `json:"data"`
	Meta PaginationMeta `json:"meta"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username   string `json:"username" binding:"required"`
	Password   string `json:"password" binding:"required"`
	Captcha    string `json:"captcha" binding:"required"`
	CaptchaKey string `json:"captcha_key" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	User      User      `json:"user"`
}

// PluginSearchParams 插件搜索参数
type PluginSearchParams struct {
	Query      string     `form:"q"`
	CategoryID *uuid.UUID `form:"category_id"`
	Featured   *bool      `form:"featured"`
	Free       *bool      `form:"free"`
	MinPrice   *float64   `form:"min_price"`
	MaxPrice   *float64   `form:"max_price"`
	MinRating  *float64   `form:"min_rating"`
	SortBy     string     `form:"sort_by"`    // name, rating, downloads, created_at
	SortOrder  string     `form:"sort_order"` // asc, desc
	Page       int        `form:"page"`
	Limit      int        `form:"page_size"`
}

// ReviewSearchParams 评论搜索参数
type ReviewSearchParams struct {
	PluginID  uuid.UUID `form:"plugin_id" binding:"required"`
	Rating    *int      `form:"rating"`
	Verified  *bool     `form:"verified"`
	SortBy    string    `form:"sort_by"`    // rating, helpful_count, created_at
	SortOrder string    `form:"sort_order"` // asc, desc
	Page      int       `form:"page"`
	Limit     int       `form:"page_size"`
}
