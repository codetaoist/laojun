# 认证接口

本文档详细介绍太上老君系统的认证相关 API 接口。

## 基础信息

- **模块路径**: `/auth`
- **认证要求**: 部分接口需要认证
- **限流策略**: 登录接口每分钟最多 10 次尝试

## 接口列表

### 1. 用户登录

```http
POST /auth/login
```

**描述**: 用户登录获取访问令牌

**请求参数**:
```json
{
  "username": "string",     // 用户名或邮箱
  "password": "string",     // 密码
  "remember": "boolean"     // 是否记住登录状态（可选）
}
```

**响应示例**:
```json
{
  "code": 200,
  "message": "登录成功",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "token_type": "Bearer",
    "expires_in": 3600,
    "user": {
      "id": 1,
      "username": "admin",
      "email": "admin@example.com",
      "role": "admin",
      "avatar": "https://cdn.laojun.dev/avatars/1.jpg"
    }
  }
}
```

**错误响应**:
```json
{
  "code": 401,
  "message": "用户名或密码错误",
  "timestamp": "2023-12-01T10:00:00Z"
}
```

### 2. 用户注册

```http
POST /auth/register
```

**描述**: 新用户注册

**请求参数**:
```json
{
  "username": "string",     // 用户名（3-20字符）
  "email": "string",        // 邮箱地址
  "password": "string",     // 密码（8-50字符）
  "confirm_password": "string", // 确认密码
  "invite_code": "string"   // 邀请码（可选）
}
```

**响应示例**:
```json
{
  "code": 200,
  "message": "注册成功",
  "data": {
    "user_id": 123,
    "username": "newuser",
    "email": "newuser@example.com",
    "status": "pending_verification"
  }
}
```

### 3. 刷新令牌

```http
POST /auth/refresh
```

**描述**: 使用刷新令牌获取新的访问令牌

**请求头**:
```http
Authorization: Bearer <refresh_token>
```

**响应示例**:
```json
{
  "code": 200,
  "message": "令牌刷新成功",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "token_type": "Bearer",
    "expires_in": 3600
  }
}
```

### 4. 用户登出

```http
POST /auth/logout
```

**描述**: 用户登出，使令牌失效

**请求头**:
```http
Authorization: Bearer <access_token>
```

**响应示例**:
```json
{
  "code": 200,
  "message": "登出成功"
}
```

### 5. 忘记密码

```http
POST /auth/forgot-password
```

**描述**: 发送密码重置邮件

**请求参数**:
```json
{
  "email": "string"         // 注册邮箱
}
```

**响应示例**:
```json
{
  "code": 200,
  "message": "密码重置邮件已发送",
  "data": {
    "email": "user@example.com",
    "expires_at": "2023-12-01T11:00:00Z"
  }
}
```

### 6. 重置密码

```http
POST /auth/reset-password
```

**描述**: 使用重置令牌重置密码

**请求参数**:
```json
{
  "token": "string",        // 重置令牌
  "password": "string",     // 新密码
  "confirm_password": "string" // 确认新密码
}
```

**响应示例**:
```json
{
  "code": 200,
  "message": "密码重置成功"
}
```

### 7. 验证邮箱

```http
POST /auth/verify-email
```

**描述**: 验证用户邮箱

**请求参数**:
```json
{
  "token": "string"         // 验证令牌
}
```

**响应示例**:
```json
{
  "code": 200,
  "message": "邮箱验证成功",
  "data": {
    "user_id": 123,
    "email": "user@example.com",
    "verified_at": "2023-12-01T10:00:00Z"
  }
}
```

### 8. 重发验证邮件

```http
POST /auth/resend-verification
```

**描述**: 重新发送邮箱验证邮件

**请求头**:
```http
Authorization: Bearer <access_token>
```

**响应示例**:
```json
{
  "code": 200,
  "message": "验证邮件已重新发送"
}
```

### 9. 检查令牌有效性

```http
GET /auth/verify-token
```

**描述**: 验证当前令牌是否有效

**请求头**:
```http
Authorization: Bearer <access_token>
```

**响应示例**:
```json
{
  "code": 200,
  "message": "令牌有效",
  "data": {
    "user_id": 1,
    "username": "admin",
    "role": "admin",
    "expires_at": "2023-12-01T11:00:00Z"
  }
}
```

### 10. 获取用户会话

```http
GET /auth/sessions
```

**描述**: 获取用户的活跃会话列表

**请求头**:
```http
Authorization: Bearer <access_token>
```

**响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "sessions": [
      {
        "id": "session_123",
        "device": "Chrome on Windows",
        "ip_address": "192.168.1.100",
        "location": "北京, 中国",
        "created_at": "2023-12-01T09:00:00Z",
        "last_active": "2023-12-01T10:00:00Z",
        "is_current": true
      }
    ]
  }
}
```

### 11. 撤销会话

```http
DELETE /auth/sessions/{session_id}
```

**描述**: 撤销指定的用户会话

**请求头**:
```http
Authorization: Bearer <access_token>
```

**路径参数**:
- `session_id`: 会话 ID

**响应示例**:
```json
{
  "code": 200,
  "message": "会话已撤销"
}
```

## 错误码说明

| 错误码 | HTTP状态码 | 说明 |
|--------|------------|------|
| E001 | 400 | 请求参数错误 |
| E002 | 401 | 用户名或密码错误 |
| E003 | 401 | 令牌无效或已过期 |
| E004 | 403 | 账户被禁用 |
| E005 | 409 | 用户名或邮箱已存在 |
| E006 | 429 | 请求过于频繁 |
| E007 | 422 | 邮箱未验证 |
| E008 | 410 | 重置令牌已过期 |

## 安全注意事项

### 1. 密码策略
- 最少 8 个字符
- 包含大小写字母、数字和特殊字符
- 不能与用户名相同
- 不能是常见弱密码

### 2. 令牌管理
- 访问令牌有效期：1 小时
- 刷新令牌有效期：30 天
- 令牌在登出时立即失效
- 支持令牌黑名单机制

### 3. 限流保护
- 登录失败 5 次后锁定账户 15 分钟
- 密码重置邮件每小时最多发送 3 次
- 验证邮件每天最多发送 10 次

### 4. 安全日志
- 记录所有认证相关操作
- 监控异常登录行为
- 支持多因素认证（MFA）

## 示例代码

### JavaScript (Axios)

```javascript
// 用户登录
const login = async (username, password) => {
  try {
    const response = await axios.post('/auth/login', {
      username,
      password
    });
    
    // 保存令牌
    localStorage.setItem('access_token', response.data.data.access_token);
    localStorage.setItem('refresh_token', response.data.data.refresh_token);
    
    return response.data;
  } catch (error) {
    console.error('登录失败:', error.response.data);
    throw error;
  }
};

// 刷新令牌
const refreshToken = async () => {
  const refreshToken = localStorage.getItem('refresh_token');
  
  try {
    const response = await axios.post('/auth/refresh', {}, {
      headers: {
        'Authorization': `Bearer ${refreshToken}`
      }
    });
    
    localStorage.setItem('access_token', response.data.data.access_token);
    return response.data;
  } catch (error) {
    // 刷新失败，跳转到登录页
    localStorage.clear();
    window.location.href = '/login';
  }
};
```

### Go

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
)

type LoginRequest struct {
    Username string `json:"username"`
    Password string `json:"password"`
}

type LoginResponse struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Data    struct {
        AccessToken  string `json:"access_token"`
        RefreshToken string `json:"refresh_token"`
        TokenType    string `json:"token_type"`
        ExpiresIn    int    `json:"expires_in"`
    } `json:"data"`
}

func login(username, password string) (*LoginResponse, error) {
    loginReq := LoginRequest{
        Username: username,
        Password: password,
    }
    
    jsonData, _ := json.Marshal(loginReq)
    
    resp, err := http.Post("http://localhost:8080/auth/login", 
        "application/json", bytes.NewBuffer(jsonData))
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var loginResp LoginResponse
    if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
        return nil, err
    }
    
    return &loginResp, nil
}
```

### Python (Requests)

```python
import requests
import json

class AuthClient:
    def __init__(self, base_url):
        self.base_url = base_url
        self.session = requests.Session()
    
    def login(self, username, password):
        """用户登录"""
        url = f"{self.base_url}/auth/login"
        data = {
            "username": username,
            "password": password
        }
        
        response = self.session.post(url, json=data)
        response.raise_for_status()
        
        result = response.json()
        if result['code'] == 200:
            # 保存令牌到会话头
            token = result['data']['access_token']
            self.session.headers.update({
                'Authorization': f'Bearer {token}'
            })
        
        return result
    
    def refresh_token(self, refresh_token):
        """刷新令牌"""
        url = f"{self.base_url}/auth/refresh"
        headers = {'Authorization': f'Bearer {refresh_token}'}
        
        response = self.session.post(url, headers=headers)
        response.raise_for_status()
        
        return response.json()

# 使用示例
client = AuthClient("http://localhost:8080")
result = client.login("admin", "password123")
print(f"登录成功: {result['data']['user']['username']}")
```

更多示例和最佳实践，请参考 [API 使用指南](../usage-guide.md)。