-- 太上老君系统数据库初始化脚本
-- Database initialization script for Taishang Laojun System

-- 创建扩展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 用户表
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    nickname VARCHAR(100),
    avatar VARCHAR(255),
    phone VARCHAR(20),
    status INTEGER DEFAULT 1, -- 1: 正常, 0: 禁用
    role VARCHAR(20) DEFAULT 'user', -- admin, user
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 分类表
CREATE TABLE IF NOT EXISTS categories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    parent_id UUID REFERENCES categories(id),
    sort_order INTEGER DEFAULT 0,
    status INTEGER DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 插件表
CREATE TABLE IF NOT EXISTS plugins (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    version VARCHAR(20) NOT NULL,
    author VARCHAR(100),
    category_id UUID REFERENCES categories(id),
    price DECIMAL(10,2) DEFAULT 0.00,
    download_count INTEGER DEFAULT 0,
    rating DECIMAL(3,2) DEFAULT 0.00,
    status INTEGER DEFAULT 1, -- 1: 上架, 0: 下架, 2: 审核中
    file_path VARCHAR(255),
    file_size BIGINT,
    screenshots TEXT[], -- JSON array of screenshot URLs
    tags TEXT[],
    requirements TEXT,
    changelog TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 订单表
CREATE TABLE IF NOT EXISTS orders (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id),
    plugin_id UUID NOT NULL REFERENCES plugins(id),
    amount DECIMAL(10,2) NOT NULL,
    status INTEGER DEFAULT 0, -- 0: 待支付, 1: 已支付, 2: 已取消, 3: 已退款
    payment_method VARCHAR(20),
    payment_id VARCHAR(100),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 评论表
CREATE TABLE IF NOT EXISTS reviews (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id),
    plugin_id UUID NOT NULL REFERENCES plugins(id),
    rating INTEGER NOT NULL CHECK (rating >= 1 AND rating <= 5),
    content TEXT,
    status INTEGER DEFAULT 1, -- 1: 显示, 0: 隐藏
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 下载记录表
CREATE TABLE IF NOT EXISTS downloads (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id),
    plugin_id UUID NOT NULL REFERENCES plugins(id),
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 配置表
CREATE TABLE IF NOT EXISTS configs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    key VARCHAR(100) UNIQUE NOT NULL,
    value TEXT,
    description TEXT,
    type VARCHAR(20) DEFAULT 'string', -- string, number, boolean, json
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);

CREATE INDEX IF NOT EXISTS idx_categories_parent_id ON categories(parent_id);
CREATE INDEX IF NOT EXISTS idx_categories_status ON categories(status);

CREATE INDEX IF NOT EXISTS idx_plugins_category_id ON plugins(category_id);
CREATE INDEX IF NOT EXISTS idx_plugins_status ON plugins(status);
CREATE INDEX IF NOT EXISTS idx_plugins_created_at ON plugins(created_at);

CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id);
CREATE INDEX IF NOT EXISTS idx_orders_plugin_id ON orders(plugin_id);
CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);

CREATE INDEX IF NOT EXISTS idx_reviews_plugin_id ON reviews(plugin_id);
CREATE INDEX IF NOT EXISTS idx_reviews_user_id ON reviews(user_id);
CREATE INDEX IF NOT EXISTS idx_reviews_status ON reviews(status);

CREATE INDEX IF NOT EXISTS idx_downloads_plugin_id ON downloads(plugin_id);
CREATE INDEX IF NOT EXISTS idx_downloads_user_id ON downloads(user_id);
CREATE INDEX IF NOT EXISTS idx_downloads_created_at ON downloads(created_at);

CREATE INDEX IF NOT EXISTS idx_configs_key ON configs(key);

-- 插入默认数据
INSERT INTO configs (key, value, description, type) VALUES
('site_name', '太上老君插件市场', '网站名称', 'string'),
('site_description', '专业的插件市场平台', '网站描述', 'string'),
('default_category', '工具', '默认分类', 'string'),
('max_file_size', '104857600', '最大文件大小(字节)', 'number'),
('allowed_file_types', '["zip", "tar.gz", "rar"]', '允许的文件类型', 'json')
ON CONFLICT (key) DO NOTHING;

-- 插入默认分类
INSERT INTO categories (name, description, sort_order) VALUES
('工具', '实用工具类插件', 1),
('游戏', '游戏相关插件', 2),
('娱乐', '娱乐类插件', 3),
('教育', '教育学习类插件', 4),
('商务', '商务办公类插件', 5)
ON CONFLICT DO NOTHING;

-- 插入默认管理员用户 (密码: admin123)
INSERT INTO users (username, email, password_hash, nickname, role) VALUES
('admin', 'admin@laojun.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', '管理员', 'admin')
ON CONFLICT (username) DO NOTHING;

-- 创建更新时间触发器函数
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 为所有表创建更新时间触发器
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_categories_updated_at BEFORE UPDATE ON categories FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_plugins_updated_at BEFORE UPDATE ON plugins FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_orders_updated_at BEFORE UPDATE ON orders FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_reviews_updated_at BEFORE UPDATE ON reviews FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_configs_updated_at BEFORE UPDATE ON configs FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();