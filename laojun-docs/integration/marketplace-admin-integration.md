# 插件市场与总后台集成方案

## 1. 集成架构概述

### 1.1 系统组件
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   插件市场前端   │    │   总后台前端     │    │   开发者工具     │
│   (React)      │    │   (React)      │    │   (CLI)        │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
┌─────────────────────────────────┼─────────────────────────────────┐
│                    API Gateway                                    │
└─────────────────────────────────┼─────────────────────────────────┘
                                 │
         ┌───────────────────────┼───────────────────────┐
         │                       │                       │
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  插件市场API     │    │   总后台API     │    │   共享服务       │
│  (marketplace)  │    │   (admin)      │    │   (shared)      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
                    ┌─────────────────┐
                    │   PostgreSQL    │
                    │   数据库        │
                    └─────────────────┘
```

### 1.2 数据流向
- **插件市场 → 总后台**: 插件数据同步、审核状态更新
- **总后台 → 插件市场**: 审核结果、配置更新、状态管理
- **双向同步**: 用户数据、权限信息、统计数据

## 2. 数据模型映射

### 2.1 核心表结构
```sql
-- 插件市场核心表
mp_plugins (插件基础信息)
mp_plugin_versions (插件版本)
mp_categories (插件分类)
mp_developers (开发者信息)
mp_plugin_reviews (插件评价)

-- 总后台管理表
admin_plugin_management (插件管理记录)
admin_review_queue (审核队列)
admin_plugin_configs (插件配置)
admin_plugin_stats (插件统计)
```

### 2.2 数据映射关系
```yaml
插件市场数据 -> 总后台数据:
  mp_plugins.id -> admin_plugin_management.plugin_id
  mp_plugins.status -> admin_plugin_management.status
  mp_plugins.developer_id -> admin_plugin_management.developer_id
  mp_plugin_reviews -> admin_plugin_stats.review_data
```

## 3. API接口设计

### 3.1 插件市场API接口

#### 插件管理接口
```http
# 获取插件列表
GET /api/v1/marketplace/plugins
Query: page, limit, category, status, search, sort

# 获取插件详情
GET /api/v1/marketplace/plugins/{id}

# 创建插件
POST /api/v1/marketplace/plugins
Body: name, description, category_id, version, files

# 更新插件
PUT /api/v1/marketplace/plugins/{id}
Body: description, metadata, status

# 删除插件
DELETE /api/v1/marketplace/plugins/{id}
```

#### 插件版本管理
```http
# 发布新版本
POST /api/v1/marketplace/plugins/{id}/versions
Body: version, changelog, files

# 获取版本列表
GET /api/v1/marketplace/plugins/{id}/versions

# 获取版本详情
GET /api/v1/marketplace/plugins/{id}/versions/{version}
```

#### 插件分类管理
```http
# 获取分类列表
GET /api/v1/marketplace/categories

# 获取分类下的插件
GET /api/v1/marketplace/categories/{id}/plugins
```

### 3.2 总后台API接口

#### 插件审核管理
```http
# 获取审核队列
GET /api/v1/admin/plugins/review-queue
Query: status, priority, reviewer, page, limit

# 分配审核员
POST /api/v1/admin/plugins/{id}/assign-reviewer
Body: reviewer_id, priority

# 执行审核
POST /api/v1/admin/plugins/{id}/review
Body: status, notes, reviewer_id

# 批量审核
POST /api/v1/admin/plugins/batch-review
Body: plugin_ids[], action, notes
```

#### 插件管理
```http
# 获取插件管理列表
GET /api/v1/admin/plugins
Query: status, category, developer, page, limit

# 更新插件状态
PATCH /api/v1/admin/plugins/{id}/status
Body: status, reason

# 获取插件统计
GET /api/v1/admin/plugins/{id}/stats

# 获取插件配置
GET /api/v1/admin/plugins/{id}/config
PUT /api/v1/admin/plugins/{id}/config
```

#### 开发者管理
```http
# 获取开发者列表
GET /api/v1/admin/developers
Query: status, verified, page, limit

# 开发者认证
POST /api/v1/admin/developers/{id}/verify
Body: verified, notes

# 获取开发者统计
GET /api/v1/admin/developers/{id}/stats
```

### 3.3 数据同步接口

#### 插件数据同步
```http
# 同步插件到总后台
POST /api/v1/sync/plugins/{id}/to-admin
Body: sync_type, force_update

# 从总后台同步状态
POST /api/v1/sync/plugins/{id}/from-admin
Body: status, config, metadata

# 批量同步
POST /api/v1/sync/plugins/batch
Body: plugin_ids[], sync_direction
```

## 4. 服务层设计

### 4.1 插件市场服务

#### PluginService
```go
type PluginService struct {
    db           *gorm.DB
    redis        *redis.Client
    adminClient  *AdminAPIClient
}

func (s *PluginService) CreatePlugin(plugin *models.Plugin) error
func (s *PluginService) UpdatePlugin(id string, updates map[string]interface{}) error
func (s *PluginService) GetPlugins(filters PluginFilters) (*PaginatedPlugins, error)
func (s *PluginService) SyncToAdmin(pluginID string) error
```

#### CategoryService
```go
type CategoryService struct {
    db    *gorm.DB
    cache *redis.Client
}

func (s *CategoryService) GetCategories() ([]models.Category, error)
func (s *CategoryService) GetCategoryPlugins(categoryID string) ([]models.Plugin, error)
```

### 4.2 总后台服务

#### AdminPluginService
```go
type AdminPluginService struct {
    db              *gorm.DB
    marketplaceClient *MarketplaceAPIClient
    reviewService   *ReviewService
}

func (s *AdminPluginService) GetPluginsForReview() ([]models.PluginReview, error)
func (s *AdminPluginService) ReviewPlugin(pluginID string, review ReviewRequest) error
func (s *AdminPluginService) UpdatePluginStatus(pluginID string, status string) error
func (s *AdminPluginService) SyncFromMarketplace(pluginID string) error
```

#### ReviewService
```go
type ReviewService struct {
    db           *gorm.DB
    autoReviewer *AutoReviewer
}

func (s *ReviewService) AssignReviewer(pluginID, reviewerID string) error
func (s *ReviewService) ProcessReview(review ReviewRequest) error
func (s *ReviewService) GetReviewQueue(filters ReviewFilters) (*PaginatedReviews, error)
```

### 4.3 同步服务

#### SyncService
```go
type SyncService struct {
    marketplaceDB *gorm.DB
    adminDB       *gorm.DB
    eventBus      *EventBus
}

func (s *SyncService) SyncPluginToAdmin(pluginID string) error
func (s *SyncService) SyncStatusFromAdmin(pluginID string, status string) error
func (s *SyncService) BatchSync(pluginIDs []string, direction string) error
```

## 5. 事件驱动架构

### 5.1 事件定义
```go
type PluginEvent struct {
    Type      string    `json:"type"`
    PluginID  string    `json:"plugin_id"`
    Data      interface{} `json:"data"`
    Timestamp time.Time `json:"timestamp"`
    Source    string    `json:"source"`
}

// 事件类型
const (
    PluginCreated    = "plugin.created"
    PluginUpdated    = "plugin.updated"
    PluginDeleted    = "plugin.deleted"
    PluginReviewed   = "plugin.reviewed"
    PluginPublished  = "plugin.published"
    PluginSuspended  = "plugin.suspended"
)
```

### 5.2 事件处理器
```go
type EventHandler interface {
    Handle(event PluginEvent) error
}

type PluginSyncHandler struct {
    syncService *SyncService
}

func (h *PluginSyncHandler) Handle(event PluginEvent) error {
    switch event.Type {
    case PluginCreated:
        return h.syncService.SyncPluginToAdmin(event.PluginID)
    case PluginReviewed:
        return h.syncService.SyncStatusFromAdmin(event.PluginID, event.Data.(string))
    }
    return nil
}
```

## 6. 缓存策略

### 6.1 Redis缓存设计
```yaml
缓存键设计:
  plugin:{id}:info          # 插件基础信息 (TTL: 1h)
  plugin:{id}:versions      # 插件版本列表 (TTL: 30m)
  category:{id}:plugins     # 分类插件列表 (TTL: 15m)
  review:queue:{status}     # 审核队列 (TTL: 5m)
  stats:plugin:{id}         # 插件统计数据 (TTL: 1h)
  sync:status:{id}          # 同步状态 (TTL: 10m)
```

### 6.2 缓存更新策略
- **写入时更新**: 插件创建/更新时清除相关缓存
- **定时刷新**: 统计数据每小时刷新
- **事件驱动**: 审核状态变更时立即更新缓存

## 7. 权限控制

### 7.1 角色定义
```yaml
roles:
  marketplace_admin:
    permissions:
      - marketplace.plugin.view_all
      - marketplace.plugin.manage
      - marketplace.category.manage
      
  plugin_reviewer:
    permissions:
      - admin.plugin.review
      - admin.plugin.view_queue
      - admin.plugin.assign
      
  developer:
    permissions:
      - marketplace.plugin.create
      - marketplace.plugin.update_own
      - marketplace.plugin.view_own
```

### 7.2 权限验证
```go
type PermissionService struct {
    userService *UserService
    roleService *RoleService
}

func (s *PermissionService) CheckPermission(userID, resource, action string) bool
func (s *PermissionService) GetUserPermissions(userID string) []string
```

## 8. 监控和日志

### 8.1 监控指标
- API响应时间和成功率
- 插件同步成功率
- 审核队列长度
- 缓存命中率
- 数据库连接池状态

### 8.2 日志记录
```go
type AuditLog struct {
    ID        string    `json:"id"`
    UserID    string    `json:"user_id"`
    Action    string    `json:"action"`
    Resource  string    `json:"resource"`
    Details   string    `json:"details"`
    Timestamp time.Time `json:"timestamp"`
    IP        string    `json:"ip"`
}
```

## 9. 部署配置

### 9.1 环境变量
```env
# 数据库配置
DB_HOST=localhost
DB_PORT=5432
DB_NAME=laojun
DB_USER=laojun
DB_PASSWORD=laojun123

# Redis配置
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# API配置
MARKETPLACE_API_URL=http://localhost:8081
ADMIN_API_URL=http://localhost:8082

# 同步配置
SYNC_INTERVAL=300
SYNC_BATCH_SIZE=100
```

### 9.2 Docker配置
```yaml
version: '3.8'
services:
  marketplace-api:
    build: ./cmd/marketplace-api
    ports:
      - "8081:8080"
    environment:
      - DB_HOST=postgres
      - REDIS_HOST=redis
    depends_on:
      - postgres
      - redis

  admin-api:
    build: ./cmd/admin-api
    ports:
      - "8082:8080"
    environment:
      - DB_HOST=postgres
      - REDIS_HOST=redis
    depends_on:
      - postgres
      - redis
```

## 10. 实施计划

### 10.1 第一阶段：基础集成
1. 创建数据同步服务
2. 实现基础API接口
3. 建立事件驱动机制

### 10.2 第二阶段：功能完善
1. 实现插件审核流程
2. 添加权限控制
3. 完善缓存策略

### 10.3 第三阶段：优化提升
1. 性能优化
2. 监控告警
3. 自动化测试

## 11. 测试策略

### 11.1 单元测试
- 服务层逻辑测试
- API接口测试
- 数据同步测试

### 11.2 集成测试
- 跨服务调用测试
- 数据一致性测试
- 事件处理测试

### 11.3 性能测试
- API响应时间测试
- 并发处理能力测试
- 数据同步性能测试