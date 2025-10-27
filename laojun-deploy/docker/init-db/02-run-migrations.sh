#!/bin/bash
# 数据库迁移执行脚本
# 在数据库初始化后运行所有迁移文件

set -e

echo "Starting database migrations..."

# 等待数据库完全就绪
# 在初始化脚本中，数据库已经在运行，我们只需要确保可以连接
echo "Database should be ready during initialization, proceeding with migrations..."

echo "Database is ready, running migrations..."

# 按顺序执行迁移文件
MIGRATION_DIR="/docker-entrypoint-initdb.d/migrations"

if [ -d "$MIGRATION_DIR" ]; then
    for migration_file in $(ls $MIGRATION_DIR/*.up.sql 2>/dev/null | sort); do
        echo "Running migration: $(basename $migration_file)"
        psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" -f "$migration_file"
    done
    echo "All migrations completed successfully"
else
    echo "No migration directory found, skipping migrations"
fi

echo "Database initialization and migration completed"