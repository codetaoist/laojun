package services

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// JWTKeyService JWT密钥管理服务
type JWTKeyService struct {
	db *sql.DB
}

// JWTKey JWT密钥结构
type JWTKey struct {
	ID        uuid.UUID `json:"id" db:"id"`
	KeyHash   string    `json:"-" db:"key_hash"`
	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
}

func NewJWTKeyService(db *sql.DB) *JWTKeyService {
	return &JWTKeyService{db: db}
}

// GenerateSecureKey 生成安全的JWT密钥
func (s *JWTKeyService) GenerateSecureKey() (string, error) {
	// 生成256位随机密钥
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		return "", fmt.Errorf("failed to generate random key: %w", err)
	}

	return hex.EncodeToString(key), nil
}

// CreateNewKey 创建新的JWT密钥
func (s *JWTKeyService) CreateNewKey(expiresIn time.Duration) (*JWTKey, string, error) {
	// 生成新密钥
	keyString, err := s.GenerateSecureKey()
	if err != nil {
		return nil, "", err
	}

	// 计算密钥哈希
	hash := sha256.Sum256([]byte(keyString))
	keyHash := hex.EncodeToString(hash[:])

	// 保存到数据库
	id := uuid.New()
	now := time.Now()
	expiresAt := now.Add(expiresIn)

	query := `
		INSERT INTO ua_jwt_keys (id, key_hash, is_active, created_at, expires_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err = s.db.Exec(query, id, keyHash, true, now, expiresAt)
	if err != nil {
		return nil, "", fmt.Errorf("failed to save JWT key: %w", err)
	}

	jwtKey := &JWTKey{
		ID:        id,
		KeyHash:   keyHash,
		IsActive:  true,
		CreatedAt: now,
		ExpiresAt: expiresAt,
	}

	return jwtKey, keyString, nil
}

// GetActiveKey 获取当前活跃的JWT密钥
func (s *JWTKeyService) GetActiveKey() (*JWTKey, error) {
	query := `
		SELECT id, key_hash, is_active, created_at, expires_at
		FROM ua_jwt_keys
		WHERE is_active = true AND expires_at > NOW()
		ORDER BY created_at DESC
		LIMIT 1
	`

	var key JWTKey
	err := s.db.QueryRow(query).Scan(
		&key.ID, &key.KeyHash, &key.IsActive, &key.CreatedAt, &key.ExpiresAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no active JWT key found")
		}
		return nil, fmt.Errorf("failed to get active JWT key: %w", err)
	}

	return &key, nil
}

// ValidateKeyHash 验证密钥哈希
func (s *JWTKeyService) ValidateKeyHash(keyString string) (bool, error) {
	hash := sha256.Sum256([]byte(keyString))
	keyHash := hex.EncodeToString(hash[:])

	query := `
		SELECT EXISTS(
			SELECT 1 FROM ua_jwt_keys 
			WHERE key_hash = $1 AND is_active = true AND expires_at > NOW()
		)
	`

	var exists bool
	err := s.db.QueryRow(query, keyHash).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to validate key hash: %w", err)
	}

	return exists, nil
}

// RotateKey 轮换JWT密钥
func (s *JWTKeyService) RotateKey(newKeyExpiresIn time.Duration) (*JWTKey, string, error) {
	// 创建新密钥
	newKey, keyString, err := s.CreateNewKey(newKeyExpiresIn)
	if err != nil {
		return nil, "", err
	}

	// 将旧密钥标记为非活跃（但不删除，以便验证现有token）
	updateQuery := `
		UPDATE ua_jwt_keys 
		SET is_active = false 
		WHERE is_active = true AND id != $1
	`

	_, err = s.db.Exec(updateQuery, newKey.ID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to deactivate old keys: %w", err)
	}

	return newKey, keyString, nil
}

// CleanExpiredKeys 清理过期的JWT密钥
func (s *JWTKeyService) CleanExpiredKeys() error {
	// 只删除过期超过7天的密钥，保留一段时间以便验证旧token
	query := `
		DELETE FROM ua_jwt_keys 
		WHERE expires_at < NOW() - INTERVAL '7 days'
	`

	result, err := s.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to clean expired keys: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		fmt.Printf("Cleaned %d expired JWT keys\n", rowsAffected)
	}

	return nil
}

// GetKeyStatistics 获取密钥统计信息
func (s *JWTKeyService) GetKeyStatistics() (map[string]int, error) {
	query := `
		SELECT 
			COUNT(*) as total,
			COUNT(CASE WHEN is_active = true THEN 1 END) as active,
			COUNT(CASE WHEN expires_at < NOW() THEN 1 END) as expired
		FROM ua_jwt_keys
	`

	var total, active, expired int
	err := s.db.QueryRow(query).Scan(&total, &active, &expired)
	if err != nil {
		return nil, fmt.Errorf("failed to get key statistics: %w", err)
	}

	return map[string]int{
		"total":   total,
		"active":  active,
		"expired": expired,
	}, nil
}
