# 数据库表前缀命名约定

## 概述
为了更好地组织和管理数据库表，我们采用基于业务域的前缀命名约定。每个业务域使用2-3个字符的前缀来标识相关表。

## 域分类和前缀

### 1. 用户认证域 (User Authentication) - `ua_`
负责用户身份验证、会话管理和基础配置
- `ua_users` - 用户基础信息
- `ua_sessions` - 用户会话
- `ua_configs` - 用户配置
- `ua_audit_logs` - 审计日志
- `ua_system_settings` - 系统设置

### 2. 插件市场域 (Marketplace) - `mp_`
负责插件市场的核心功能
- `mp_users` - 市场用户信息
- `mp_categories` - 插件分类
- `mp_plugins` - 插件信息
- `mp_plugin_versions` - 插件版本
- `mp_forum_categories` - 市场论坛分类
- `mp_forum_posts` - 市场论坛帖子

### 3. 社区功能域 (Community) - `cm_`
负责社区交流和内容分享
- `cm_forum_categories` - 论坛分类
- `cm_forum_posts` - 论坛帖子
- `cm_forum_replies` - 论坛回复
- `cm_blog_categories` - 博客分类
- `cm_blog_posts` - 博客文章
- `cm_blog_comments` - 博客评论
- `cm_code_snippets` - 代码片段

### 4. 交互功能域 (Interaction) - `ix_`
负责用户间的交互功能
- `ix_reviews` - 评论评价
- `ix_purchases` - 购买记录
- `ix_likes` - 点赞记录
- `ix_user_follows` - 用户关注
- `ix_bookmarks` - 收藏记录
- `ix_messages` - 私信
- `ix_notifications` - 通知

### 5. 积分徽章域 (Gamification) - `gm_`
负责积分系统和徽章管理
- `gm_user_points` - 用户积分
- `gm_point_records` - 积分记录
- `gm_badges` - 徽章
- `gm_user_badges` - 用户徽章

### 6. 权限管理域 (Authorization) - `az_`
负责角色和权限管理
- `az_roles` - 角色
- `az_permissions` - 权限
- `az_user_roles` - 用户角色关联
- `az_role_permissions` - 角色权限关联

## 命名规则

1. **前缀格式**: 使用2-3个小写字母，后跟下划线
2. **表名**: 使用小写字母和下划线分隔
3. **一致性**: 同一域内的所有表必须使用相同前缀
4. **可读性**: 前缀应该能够清晰表达业务域的含义

## 索引命名约定

索引名称应该包含表前缀，格式为：
- 主键索引: `pk_{table_name}`
- 唯一索引: `uk_{table_name}_{column_name}`
- 普通索引: `idx_{table_name}_{column_name}`
- 外键索引: `fk_{table_name}_{referenced_table}`

## 迁移策略

1. 创建新的迁移文件来重命名现有表
2. 更新所有相关的Go模型和代码引用
3. 更新种子数据文件
4. 确保所有测试通过
5. 更新文档和API文档

## 优势

1. **清晰的域分离**: 通过前缀可以快速识别表所属的业务域
2. **更好的组织**: 相关表在数据库管理工具中会自动分组
3. **避免命名冲突**: 不同域可以有相似的表名而不冲突
4. **便于维护**: 开发者可以快速定位相关表
5. **扩展性**: 新增业务域时可以轻松添加新前缀