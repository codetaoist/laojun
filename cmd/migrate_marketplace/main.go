package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	// 从环境变量获取数据库连接信息
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "5432"
	}
	user := os.Getenv("DB_USER")
	if user == "" {
		user = "laojun"
	}
	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		password = "laojun123"
	}
	dbname := os.Getenv("DB_NAME")
	if dbname == "" {
		dbname = "laojun"
	}

	// 构建连接字符串
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	// 连接数据库
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("连接数据库失败:", err)
	}
	defer db.Close()

	// 测试连接
	if err := db.Ping(); err != nil {
		log.Fatal("数据库连接测试失败:", err)
	}

	fmt.Println("数据库连接成功")

	// 开始事务
	tx, err := db.Begin()
	if err != nil {
		log.Fatal("开始事务失败:", err)
	}
	defer tx.Rollback()

	// 0. 创建必要的函数
	if err := createRequiredFunctions(tx); err != nil {
		log.Fatal("创建必要函数失败:", err)
	}

	// 1. 检查并重命名plugins表为mp_plugins
	if err := renamePluginsTable(tx); err != nil {
		log.Fatal("重命名plugins表失败:", err)
	}

	// 2. 检查并创建mp_categories表
	if err := createCategoriesTable(tx); err != nil {
		log.Fatal("创建mp_categories表失败:", err)
	}

	// 3. 更新mp_plugins表结构
	if err := updatePluginsTableStructure(tx); err != nil {
		log.Fatal("更新mp_plugins表结构失败:", err)
	}

	// 4. 创建其他marketplace相关表
	if err := createMarketplaceTables(tx); err != nil {
		log.Fatal("创建marketplace相关表失败:", err)
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		log.Fatal("提交事务失败:", err)
	}

	fmt.Println("Marketplace数据库迁移完成!")
}

func createRequiredFunctions(tx *sql.Tx) error {
	fmt.Println("创建必要的数据库函数...")
	
	// 检查函数是否存在
	var exists bool
	err := tx.QueryRow("SELECT EXISTS (SELECT FROM pg_proc WHERE proname = 'update_updated_at_column')").Scan(&exists)
	if err != nil {
		return err
	}
	
	if !exists {
		fmt.Println("创建update_updated_at_column函数...")
		createFunctionSQL := `
		CREATE OR REPLACE FUNCTION update_updated_at_column()
		RETURNS TRIGGER AS $$
		BEGIN
			NEW.updated_at = CURRENT_TIMESTAMP;
			RETURN NEW;
		END;
		$$ language 'plpgsql';
		`
		_, err = tx.Exec(createFunctionSQL)
		if err != nil {
			return err
		}
		fmt.Println("✓ update_updated_at_column函数创建成功")
	} else {
		fmt.Println("✓ update_updated_at_column函数已存在")
	}
	
	return nil
}

func renamePluginsTable(tx *sql.Tx) error {
	// 检查plugins表是否存在
	var pluginsExists bool
	err := tx.QueryRow("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'plugins')").Scan(&pluginsExists)
	if err != nil {
		return err
	}

	// 检查mp_plugins表是否存在
	var mpPluginsExists bool
	err = tx.QueryRow("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'mp_plugins')").Scan(&mpPluginsExists)
	if err != nil {
		return err
	}

	if pluginsExists && !mpPluginsExists {
		fmt.Println("重命名plugins表为mp_plugins...")
		_, err = tx.Exec("ALTER TABLE plugins RENAME TO mp_plugins")
		if err != nil {
			return err
		}
		fmt.Println("✓ plugins表已重命名为mp_plugins")
	} else if mpPluginsExists {
		fmt.Println("✓ mp_plugins表已存在")
	} else {
		fmt.Println("⚠ plugins表不存在，将创建新的mp_plugins表")
	}

	return nil
}

func createCategoriesTable(tx *sql.Tx) error {
	// 检查mp_categories表是否存在
	var exists bool
	err := tx.QueryRow("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'mp_categories')").Scan(&exists)
	if err != nil {
		return err
	}

	if !exists {
		fmt.Println("创建mp_categories表...")
		createCategoriesSQL := `
		CREATE TABLE mp_categories (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			name VARCHAR(255) NOT NULL,
			slug VARCHAR(255) UNIQUE NOT NULL,
			description TEXT,
			icon VARCHAR(255),
			color VARCHAR(7) DEFAULT '#6366f1',
			parent_id UUID REFERENCES mp_categories(id) ON DELETE CASCADE,
			sort_order INTEGER DEFAULT 0,
			is_active BOOLEAN DEFAULT TRUE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);

		-- 创建索引
		CREATE INDEX idx_mp_categories_slug ON mp_categories(slug);
		CREATE INDEX idx_mp_categories_parent_id ON mp_categories(parent_id);
		CREATE INDEX idx_mp_categories_sort_order ON mp_categories(sort_order);
		CREATE INDEX idx_mp_categories_is_active ON mp_categories(is_active);

		-- 创建触发器
		CREATE TRIGGER update_mp_categories_updated_at 
			BEFORE UPDATE ON mp_categories 
			FOR EACH ROW 
			EXECUTE FUNCTION update_updated_at_column();
		`
		_, err = tx.Exec(createCategoriesSQL)
		if err != nil {
			return err
		}
		fmt.Println("✓ mp_categories表创建成功")
	} else {
		fmt.Println("✓ mp_categories表已存在")
	}

	return nil
}

func updatePluginsTableStructure(tx *sql.Tx) error {
	fmt.Println("更新mp_plugins表结构...")

	// 检查mp_plugins表是否存在
	var exists bool
	err := tx.QueryRow("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'mp_plugins')").Scan(&exists)
	if err != nil {
		return err
	}

	if !exists {
		// 如果mp_plugins表不存在，创建新表
		fmt.Println("创建新的mp_plugins表...")
		createPluginsSQL := `
		CREATE TABLE mp_plugins (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			name VARCHAR(255) NOT NULL,
			slug VARCHAR(255) UNIQUE NOT NULL,
			description TEXT,
			short_description TEXT,
			author VARCHAR(255) NOT NULL,
			developer_id UUID,
			version VARCHAR(50) NOT NULL,
			icon_url VARCHAR(500),
			banner_url VARCHAR(500),
			price DECIMAL(10,2) DEFAULT 0.00,
			is_free BOOLEAN DEFAULT TRUE,
			rating DECIMAL(3,2) DEFAULT 0.00,
			download_count INTEGER DEFAULT 0,
			review_count INTEGER DEFAULT 0,
			is_featured BOOLEAN DEFAULT FALSE,
			is_active BOOLEAN DEFAULT TRUE,
			category_id UUID REFERENCES mp_categories(id) ON DELETE SET NULL,
			tags TEXT[],
			requirements JSONB DEFAULT '{}',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			
			UNIQUE(name, version)
		);

		-- 创建索引
		CREATE INDEX idx_mp_plugins_slug ON mp_plugins(slug);
		CREATE INDEX idx_mp_plugins_name ON mp_plugins(name);
		CREATE INDEX idx_mp_plugins_author ON mp_plugins(author);
		CREATE INDEX idx_mp_plugins_category_id ON mp_plugins(category_id);
		CREATE INDEX idx_mp_plugins_is_featured ON mp_plugins(is_featured);
		CREATE INDEX idx_mp_plugins_is_active ON mp_plugins(is_active);
		CREATE INDEX idx_mp_plugins_download_count ON mp_plugins(download_count);
		CREATE INDEX idx_mp_plugins_rating ON mp_plugins(rating);
		CREATE INDEX idx_mp_plugins_tags ON mp_plugins USING GIN(tags);
		CREATE INDEX idx_mp_plugins_name_version ON mp_plugins(name, version);

		-- 创建触发器
		CREATE TRIGGER update_mp_plugins_updated_at 
			BEFORE UPDATE ON mp_plugins 
			FOR EACH ROW 
			EXECUTE FUNCTION update_updated_at_column();
		`
		_, err = tx.Exec(createPluginsSQL)
		if err != nil {
			return err
		}
		fmt.Println("✓ mp_plugins表创建成功")
		return nil
	}

	// 检查并添加缺失的字段
	fields := []struct {
		name         string
		dataType     string
		nullable     bool
		defaultValue string
	}{
		{"short_description", "TEXT", true, ""},
		{"developer_id", "UUID", true, ""},
		{"category_id", "UUID", true, ""},
		{"banner_url", "VARCHAR(500)", true, ""},
		{"is_free", "BOOLEAN", false, "DEFAULT TRUE"},
		{"review_count", "INTEGER", false, "DEFAULT 0"},
		{"is_active", "BOOLEAN", false, "DEFAULT TRUE"},
		{"tags", "TEXT[]", true, ""},
		{"requirements", "JSONB", false, "DEFAULT '{}'"},
	}

	for _, field := range fields {
		if !columnExists(tx, "mp_plugins", field.name) {
			fmt.Printf("添加字段 %s...\n", field.name)
			
			var sql string
			if field.nullable {
				sql = fmt.Sprintf("ALTER TABLE mp_plugins ADD COLUMN %s %s", field.name, field.dataType)
			} else {
				sql = fmt.Sprintf("ALTER TABLE mp_plugins ADD COLUMN %s %s %s", field.name, field.dataType, field.defaultValue)
			}
			
			_, err = tx.Exec(sql)
			if err != nil {
				return fmt.Errorf("添加字段 %s 失败: %v", field.name, err)
			}
			fmt.Printf("✓ 字段 %s 添加成功\n", field.name)
		}
	}

	// 检查并修改category字段为category_id
	if columnExists(tx, "mp_plugins", "category") && !columnExists(tx, "mp_plugins", "category_id") {
		fmt.Println("将category字段重命名为category_id并修改类型...")
		
		// 先添加新的category_id字段
		_, err = tx.Exec("ALTER TABLE mp_plugins ADD COLUMN category_id UUID")
		if err != nil {
			return err
		}
		
		// 删除旧的category字段
		_, err = tx.Exec("ALTER TABLE mp_plugins DROP COLUMN category")
		if err != nil {
			return err
		}
		
		fmt.Println("✓ category字段已更新为category_id")
	}

	return nil
}

func createMarketplaceTables(tx *sql.Tx) error {
	// 创建其他marketplace相关表
	tables := []struct {
		name string
		sql  string
	}{
		{
			"mp_favorites",
			`CREATE TABLE IF NOT EXISTS mp_favorites (
				id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
				user_id UUID NOT NULL,
				plugin_id UUID NOT NULL REFERENCES mp_plugins(id) ON DELETE CASCADE,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				UNIQUE(user_id, plugin_id)
			);
			CREATE INDEX IF NOT EXISTS idx_mp_favorites_user_id ON mp_favorites(user_id);
			CREATE INDEX IF NOT EXISTS idx_mp_favorites_plugin_id ON mp_favorites(plugin_id);`,
		},
		{
			"mp_purchases",
			`CREATE TABLE IF NOT EXISTS mp_purchases (
				id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
				user_id UUID NOT NULL,
				plugin_id UUID NOT NULL REFERENCES mp_plugins(id) ON DELETE CASCADE,
				amount DECIMAL(10,2) NOT NULL,
				status VARCHAR(50) DEFAULT 'completed',
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				UNIQUE(user_id, plugin_id)
			);
			CREATE INDEX IF NOT EXISTS idx_mp_purchases_user_id ON mp_purchases(user_id);
			CREATE INDEX IF NOT EXISTS idx_mp_purchases_plugin_id ON mp_purchases(plugin_id);`,
		},
		{
			"mp_reviews",
			`CREATE TABLE IF NOT EXISTS mp_reviews (
				id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
				plugin_id UUID NOT NULL REFERENCES mp_plugins(id) ON DELETE CASCADE,
				user_id UUID NOT NULL,
				rating INTEGER NOT NULL CHECK (rating >= 1 AND rating <= 5),
				comment TEXT,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				UNIQUE(plugin_id, user_id)
			);
			CREATE INDEX IF NOT EXISTS idx_mp_reviews_plugin_id ON mp_reviews(plugin_id);
			CREATE INDEX IF NOT EXISTS idx_mp_reviews_user_id ON mp_reviews(user_id);
			CREATE INDEX IF NOT EXISTS idx_mp_reviews_rating ON mp_reviews(rating);`,
		},
	}

	for _, table := range tables {
		fmt.Printf("创建表 %s...\n", table.name)
		_, err := tx.Exec(table.sql)
		if err != nil {
			return fmt.Errorf("创建表 %s 失败: %v", table.name, err)
		}
		fmt.Printf("✓ 表 %s 创建成功\n", table.name)
	}

	return nil
}

func columnExists(tx *sql.Tx, tableName, columnName string) bool {
	var exists bool
	query := `SELECT EXISTS (
		SELECT FROM information_schema.columns 
		WHERE table_name = $1 AND column_name = $2
	)`
	err := tx.QueryRow(query, tableName, columnName).Scan(&exists)
	return err == nil && exists
}