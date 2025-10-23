# 插件市场 API 接口设计

## 1. API 概述

### 1.1 基础信息

- **Base URL**: `https://api.example.com/v1`
- **认证方式**: JWT Token
- **数据格式**: JSON
- **字符编码**: UTF-8
- **API版本**: v1

### 1.2 通用响应格式

```json
{
  "success": true,
  "code": 200,
  "message": "操作成功",
  "data": {},
  "timestamp": "2024-01-01T00:00:00Z",
  "request_id": "uuid"
}
```

### 1.3 错误响应格式

```json
{
  "success": false,
  "code": 400,
  "message": "请求参数错误",
  "error": {
    "type": "ValidationError",
    "details": [
      {
        "field": "name",
        "message": "插件名称不能为空"
      }
    ]
  },
  "timestamp": "2024-01-01T00:00:00Z",
  "request_id": "uuid"
}
```

### 1.4 状态码定义

| 状态码 | 说明 |
|--------|------|
| 200 | 请求成功 |
| 201 | 创建成功 |
| 400 | 请求参数错误 |
| 401 | 未授权 |
| 403 | 权限不足 |
| 404 | 资源不存在 |
| 409 | 资源冲突 |
| 422 | 数据验证失败 |
| 500 | 服务器内部错误 |

## 2. 认证和授权

### 2.1 用户认证

#### 登录
```http
POST /auth/login
Content-Type: application/json

{
  "username": "developer@example.com",
  "password": "password123",
  "remember_me": true
}
```

**响应:**
```json
{
  "success": true,
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_in": 3600,
    "user": {
      "id": "uuid",
      "username": "developer",
      "email": "developer@example.com",
      "role": "developer",
      "permissions": ["plugin.create", "plugin.update"]
    }
  }
}
```

#### 刷新Token
```http
POST /auth/refresh
Authorization: Bearer {refresh_token}
```

#### 登出
```http
POST /auth/logout
Authorization: Bearer {access_token}
```

### 2.2 权限系统

#### 角色定义
```yaml
roles:
  developer:
    permissions:
      - plugin.create
      - plugin.update
      - plugin.delete
      - plugin.view_own
  
  admin:
    permissions:
      - plugin.*
      - review.*
      - user.*
      - system.*
  
  reviewer:
    permissions:
      - plugin.view_all
      - review.create
      - review.update
      - review.view
```

## 3. 插件管理 API

### 3.1 插件基础操作

#### 获取插件列表
```http
GET /marketplace/plugins?page=1&limit=20&category=tools&status=published&search=keyword
Authorization: Bearer {access_token}
```

**查询参数:**
- `page`: 页码 (默认: 1)
- `limit`: 每页数量 (默认: 20, 最大: 100)
- `category`: 分类ID
- `status`: 插件状态
- `search`: 搜索关键词
- `sort`: 排序方式 (created_at, updated_at, downloads, rating)
- `order`: 排序顺序 (asc, desc)

**响应:**
```json
{
  "success": true,
  "data": {
    "plugins": [
      {
        "id": "uuid",
        "name": "插件名称",
        "description": "插件描述",
        "version": "1.0.0",
        "status": "published",
        "category": {
          "id": "uuid",
          "name": "工具类"
        },
        "developer": {
          "id": "uuid",
          "username": "developer",
          "display_name": "开发者名称"
        },
        "downloads": 1000,
        "rating": 4.5,
        "created_at": "2024-01-01T00:00:00Z",
        "updated_at": "2024-01-01T00:00:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 100,
      "pages": 5
    }
  }
}
```

#### 获取插件详情
```http
GET /marketplace/plugins/{plugin_id}
Authorization: Bearer {access_token}
```

**响应:**
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "name": "插件名称",
    "description": "详细描述",
    "long_description": "长描述支持Markdown",
    "version": "1.0.0",
    "status": "published",
    "category": {
      "id": "uuid",
      "name": "工具类",
      "description": "分类描述"
    },
    "developer": {
      "id": "uuid",
      "username": "developer",
      "display_name": "开发者名称",
      "email": "developer@example.com",
      "website": "https://developer.com"
    },
    "metadata": {
      "tags": ["tool", "productivity"],
      "homepage": "https://plugin.com",
      "repository": "https://github.com/dev/plugin",
      "license": "MIT",
      "min_version": "1.0.0",
      "max_version": "2.0.0"
    },
    "files": {
      "icon": "https://cdn.example.com/icons/plugin.png",
      "screenshots": [
        "https://cdn.example.com/screenshots/1.png"
      ],
      "download_url": "https://cdn.example.com/plugins/plugin-1.0.0.zip"
    },
    "stats": {
      "downloads": 1000,
      "rating": 4.5,
      "reviews_count": 50,
      "size": 1024000
    },
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

#### 创建插件
```http
POST /marketplace/plugins
Authorization: Bearer {access_token}
Content-Type: multipart/form-data

{
  "name": "插件名称",
  "description": "插件描述",
  "category_id": "uuid",
  "version": "1.0.0",
  "metadata": {
    "tags": ["tool"],
    "license": "MIT"
  },
  "plugin_file": "binary_data",
  "icon": "binary_data",
  "screenshots": ["binary_data"]
}
```

#### 更新插件
```http
PUT /marketplace/plugins/{plugin_id}
Authorization: Bearer {access_token}
Content-Type: application/json

{
  "description": "更新的描述",
  "metadata": {
    "tags": ["tool", "updated"]
  }
}
```

#### 删除插件
```http
DELETE /marketplace/plugins/{plugin_id}
Authorization: Bearer {access_token}
```

### 3.2 插件版本管理

#### 发布新版本
```http
POST /marketplace/plugins/{plugin_id}/versions
Authorization: Bearer {access_token}
Content-Type: multipart/form-data

{
  "version": "1.1.0",
  "changelog": "版本更新说明",
  "plugin_file": "binary_data"
}
```

#### 获取版本列表
```http
GET /marketplace/plugins/{plugin_id}/versions
Authorization: Bearer {access_token}
```

#### 获取版本详情
```http
GET /marketplace/plugins/{plugin_id}/versions/{version}
Authorization: Bearer {access_token}
```

## 4. 插件审核 API

### 4.1 审核队列管理

#### 获取待审核插件列表
```http
GET /admin/reviews/queue?status=pending&priority=high&reviewer=uuid
Authorization: Bearer {access_token}
```

**响应:**
```json
{
  "success": true,
  "data": {
    "reviews": [
      {
        "id": "uuid",
        "plugin": {
          "id": "uuid",
          "name": "插件名称",
          "version": "1.0.0",
          "developer": "开发者名称"
        },
        "status": "pending",
        "priority": "high",
        "submitted_at": "2024-01-01T00:00:00Z",
        "assigned_reviewer": null,
        "auto_review_result": {
          "status": "passed",
          "score": 85,
          "issues": []
        },
        "estimated_time": 240
      }
    ],
    "stats": {
      "total_pending": 50,
      "high_priority": 10,
      "medium_priority": 30,
      "low_priority": 10
    }
  }
}
```

#### 分配审核任务
```http
POST /admin/reviews/{review_id}/assign
Authorization: Bearer {access_token}
Content-Type: application/json

{
  "reviewer_id": "uuid",
  "priority": "high",
  "notes": "需要重点关注安全性"
}
```

### 4.2 审核操作

#### 开始审核
```http
POST /admin/reviews/{review_id}/start
Authorization: Bearer {access_token}
```

#### 提交审核结果
```http
POST /admin/reviews/{review_id}/submit
Authorization: Bearer {access_token}
Content-Type: application/json

{
  "result": "approved",
  "notes": "审核通过，插件质量良好",
  "checklist": {
    "security": {
      "malware_scan": "pass",
      "vulnerability_check": "pass",
      "permission_review": "pass"
    },
    "quality": {
      "code_quality": "pass",
      "performance": "pass",
      "documentation": "pass"
    },
    "compliance": {
      "content_policy": "pass",
      "legal_compliance": "pass"
    }
  },
  "recommendations": [
    "建议优化启动性能",
    "建议完善错误处理"
  ]
}
```

#### 批量审核操作
```http
POST /admin/reviews/batch
Authorization: Bearer {access_token}
Content-Type: application/json

{
  "review_ids": ["uuid1", "uuid2"],
  "action": "approve",
  "notes": "批量审核通过"
}
```

### 4.3 自动审核

#### 触发自动审核
```http
POST /admin/plugins/{plugin_id}/auto-review
Authorization: Bearer {access_token}
```

#### 获取自动审核结果
```http
GET /admin/plugins/{plugin_id}/auto-review/result
Authorization: Bearer {access_token}
```

**响应:**
```json
{
  "success": true,
  "data": {
    "status": "completed",
    "overall_score": 85,
    "checks": {
      "security": {
        "score": 90,
        "status": "pass",
        "details": {
          "malware_scan": "clean",
          "vulnerability_scan": "no_issues",
          "permission_check": "appropriate"
        }
      },
      "quality": {
        "score": 80,
        "status": "pass",
        "details": {
          "code_complexity": "acceptable",
          "test_coverage": "good",
          "documentation": "complete"
        }
      },
      "performance": {
        "score": 85,
        "status": "pass",
        "details": {
          "startup_time": "2.1s",
          "memory_usage": "45MB",
          "cpu_usage": "low"
        }
      }
    },
    "issues": [
      {
        "type": "warning",
        "category": "performance",
        "message": "启动时间略长，建议优化",
        "severity": "low"
      }
    ],
    "recommendations": [
      "考虑使用懒加载优化启动性能",
      "建议添加更多单元测试"
    ]
  }
}
```

## 5. 开发者管理 API

### 5.1 开发者注册和认证

#### 开发者注册申请
```http
POST /developers/register
Content-Type: application/json

{
  "username": "developer",
  "email": "developer@example.com",
  "password": "password123",
  "display_name": "开发者名称",
  "company": "公司名称",
  "website": "https://developer.com",
  "bio": "开发者简介",
  "verification_documents": [
    {
      "type": "identity",
      "file_url": "https://cdn.example.com/docs/id.pdf"
    }
  ]
}
```

#### 获取开发者信息
```http
GET /developers/{developer_id}
Authorization: Bearer {access_token}
```

#### 更新开发者信息
```http
PUT /developers/{developer_id}
Authorization: Bearer {access_token}
Content-Type: application/json

{
  "display_name": "新的显示名称",
  "bio": "更新的简介",
  "website": "https://newwebsite.com"
}
```

### 5.2 开发者统计

#### 获取开发者统计信息
```http
GET /developers/{developer_id}/stats
Authorization: Bearer {access_token}
```

**响应:**
```json
{
  "success": true,
  "data": {
    "plugins": {
      "total": 10,
      "published": 8,
      "pending": 1,
      "rejected": 1
    },
    "downloads": {
      "total": 10000,
      "this_month": 1500,
      "growth_rate": 15.5
    },
    "ratings": {
      "average": 4.2,
      "total_reviews": 150,
      "distribution": {
        "5": 60,
        "4": 45,
        "3": 30,
        "2": 10,
        "1": 5
      }
    },
    "revenue": {
      "total": 5000.00,
      "this_month": 800.00,
      "currency": "USD"
    }
  }
}
```

## 6. 分类管理 API

### 6.1 分类操作

#### 获取分类列表
```http
GET /marketplace/categories
```

#### 创建分类
```http
POST /admin/categories
Authorization: Bearer {access_token}
Content-Type: application/json

{
  "name": "分类名称",
  "description": "分类描述",
  "parent_id": "uuid",
  "icon": "category-icon",
  "sort_order": 1
}
```

#### 更新分类
```http
PUT /admin/categories/{category_id}
Authorization: Bearer {access_token}
Content-Type: application/json

{
  "name": "更新的分类名称",
  "description": "更新的描述"
}
```

## 7. 用户评价 API

### 7.1 评价管理

#### 获取插件评价
```http
GET /marketplace/plugins/{plugin_id}/reviews?page=1&limit=20&sort=created_at
```

#### 创建评价
```http
POST /marketplace/plugins/{plugin_id}/reviews
Authorization: Bearer {access_token}
Content-Type: application/json

{
  "rating": 5,
  "title": "评价标题",
  "content": "评价内容",
  "pros": ["优点1", "优点2"],
  "cons": ["缺点1"]
}
```

#### 更新评价
```http
PUT /marketplace/plugins/{plugin_id}/reviews/{review_id}
Authorization: Bearer {access_token}
Content-Type: application/json

{
  "rating": 4,
  "content": "更新的评价内容"
}
```

## 8. 申诉管理 API

### 8.1 申诉操作

#### 提交申诉
```http
POST /developers/appeals
Authorization: Bearer {access_token}
Content-Type: application/json

{
  "plugin_id": "uuid",
  "type": "review_decision",
  "reason": "申诉原因",
  "description": "详细说明",
  "evidence": [
    {
      "type": "document",
      "url": "https://cdn.example.com/evidence.pdf",
      "description": "证据说明"
    }
  ]
}
```

#### 获取申诉列表
```http
GET /developers/appeals?status=pending&plugin_id=uuid
Authorization: Bearer {access_token}
```

#### 处理申诉
```http
POST /admin/appeals/{appeal_id}/process
Authorization: Bearer {access_token}
Content-Type: application/json

{
  "decision": "approved",
  "response": "申诉处理结果说明",
  "actions": [
    {
      "type": "update_plugin_status",
      "plugin_id": "uuid",
      "new_status": "approved"
    }
  ]
}
```

## 9. 系统管理 API

### 9.1 系统配置

#### 获取系统配置
```http
GET /admin/system/config
Authorization: Bearer {access_token}
```

#### 更新系统配置
```http
PUT /admin/system/config
Authorization: Bearer {access_token}
Content-Type: application/json

{
  "review_settings": {
    "auto_review_enabled": true,
    "manual_review_required": true,
    "max_review_time": 5
  },
  "upload_settings": {
    "max_file_size": 104857600,
    "allowed_extensions": [".zip", ".tar.gz"],
    "virus_scan_enabled": true
  }
}
```

### 9.2 统计和监控

#### 获取系统统计
```http
GET /admin/system/stats?period=30d
Authorization: Bearer {access_token}
```

**响应:**
```json
{
  "success": true,
  "data": {
    "plugins": {
      "total": 1000,
      "published": 800,
      "pending_review": 50,
      "rejected": 150
    },
    "reviews": {
      "completed": 950,
      "pending": 50,
      "average_time": 2.5,
      "approval_rate": 0.85
    },
    "developers": {
      "total": 200,
      "active": 150,
      "new_this_month": 20
    },
    "downloads": {
      "total": 100000,
      "this_month": 15000,
      "top_plugins": [
        {
          "plugin_id": "uuid",
          "name": "热门插件",
          "downloads": 5000
        }
      ]
    }
  }
}
```

## 10. WebSocket 实时通知

### 10.1 连接建立

```javascript
const ws = new WebSocket('wss://api.example.com/ws');
ws.onopen = function() {
  // 发送认证信息
  ws.send(JSON.stringify({
    type: 'auth',
    token: 'jwt_token'
  }));
};
```

### 10.2 消息格式

```json
{
  "type": "notification",
  "event": "plugin_review_completed",
  "data": {
    "plugin_id": "uuid",
    "review_result": "approved",
    "message": "您的插件已通过审核"
  },
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 10.3 事件类型

- `plugin_review_completed`: 插件审核完成
- `plugin_status_changed`: 插件状态变更
- `new_review_assigned`: 新的审核任务分配
- `appeal_processed`: 申诉处理完成
- `system_maintenance`: 系统维护通知

## 11. API 限流和安全

### 11.1 限流策略

```yaml
rate_limits:
  public_api:
    requests_per_minute: 100
    requests_per_hour: 1000
  
  authenticated_api:
    requests_per_minute: 500
    requests_per_hour: 5000
  
  admin_api:
    requests_per_minute: 1000
    requests_per_hour: 10000
```

### 11.2 安全措施

- **HTTPS**: 所有API必须使用HTTPS
- **CORS**: 配置适当的跨域策略
- **输入验证**: 严格验证所有输入参数
- **SQL注入防护**: 使用参数化查询
- **XSS防护**: 对输出进行适当编码
- **文件上传安全**: 验证文件类型和大小，病毒扫描

这个API设计提供了完整的插件市场功能，包括插件管理、审核流程、开发者管理等核心功能。