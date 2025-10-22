package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/scrypt"
)

// EncryptionConfig 加密配置
type EncryptionConfig struct {
	SecretKey string
	Salt      string
}

// Encryptor 加密工具
type Encryptor struct {
	key []byte
}

// NewEncryptor 创建新的加密工具
func NewEncryptor(config EncryptionConfig) (*Encryptor, error) {
	// 使用 scrypt 从密钥和盐生成固定长度的密钥
	key, err := scrypt.Key([]byte(config.SecretKey), []byte(config.Salt), 32768, 8, 1, 32)
	if err != nil {
		return nil, fmt.Errorf("failed to derive key: %w", err)
	}

	return &Encryptor{key: key}, nil
}

// Encrypt 加密数据
func (e *Encryptor) Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// 创建 GCM 模式
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// 生成随机 nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	// 加密数据
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)

	// 返回 base64 编码的结果
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt 解密数据
func (e *Encryptor) Decrypt(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}

	// 解码 base64
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, cipherData := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, cipherData, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}

// EncryptBytes 加密字节数据
func (e *Encryptor) EncryptBytes(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, nil
	}

	block, err := aes.NewCipher(e.key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	return gcm.Seal(nonce, nonce, data, nil), nil
}

// DecryptBytes 解密字节数据
func (e *Encryptor) DecryptBytes(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, nil
	}

	block, err := aes.NewCipher(e.key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, cipherData := data[:nonceSize], data[nonceSize:]
	return gcm.Open(nil, nonce, cipherData, nil)
}

// PasswordHasher 密码哈希工具
type PasswordHasher struct {
	cost int
}

// NewPasswordHasher 创建新的密码哈希工具
func NewPasswordHasher(cost int) *PasswordHasher {
	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		cost = bcrypt.DefaultCost
	}
	return &PasswordHasher{cost: cost}
}

// HashPassword 哈希密码
func (ph *PasswordHasher) HashPassword(password string) (string, error) {
	if password == "" {
		return "", errors.New("password cannot be empty")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), ph.cost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(hash), nil
}

// VerifyPassword 验证密码
func (ph *PasswordHasher) VerifyPassword(password, hash string) bool {
	if password == "" || hash == "" {
		return false
	}

	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateRandomKey 生成随机密钥
func GenerateRandomKey(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random key: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// GenerateRandomBytes 生成随机字节
func GenerateRandomBytes(length int) ([]byte, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return bytes, nil
}

// HashSHA256 计算 SHA256 哈希
func HashSHA256(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// HashSHA256Bytes 计算字节数据的 SHA256 哈希
func HashSHA256Bytes(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// EncryptionHelper 加密助手
type EncryptionHelper struct {
	encryptor      *Encryptor
	passwordHasher *PasswordHasher
}

// NewEncryptionHelper 创建新的加密助手
func NewEncryptionHelper(config EncryptionConfig) (*EncryptionHelper, error) {
	encryptor, err := NewEncryptor(config)
	if err != nil {
		return nil, err
	}

	passwordHasher := NewPasswordHasher(bcrypt.DefaultCost)

	return &EncryptionHelper{
		encryptor:      encryptor,
		passwordHasher: passwordHasher,
	}, nil
}

// EncryptSensitiveData 加密敏感数据
func (eh *EncryptionHelper) EncryptSensitiveData(data string) (string, error) {
	return eh.encryptor.Encrypt(data)
}

// DecryptSensitiveData 解密敏感数据
func (eh *EncryptionHelper) DecryptSensitiveData(encryptedData string) (string, error) {
	return eh.encryptor.Decrypt(encryptedData)
}

// HashPassword 哈希密码
func (eh *EncryptionHelper) HashPassword(password string) (string, error) {
	return eh.passwordHasher.HashPassword(password)
}

// VerifyPassword 验证密码
func (eh *EncryptionHelper) VerifyPassword(password, hash string) bool {
	return eh.passwordHasher.VerifyPassword(password, hash)
}

// EncryptConfig 加密配置值
func (eh *EncryptionHelper) EncryptConfig(value string, shouldEncrypt bool) (string, error) {
	if !shouldEncrypt {
		return value, nil
	}
	return eh.encryptor.Encrypt(value)
}

// DecryptConfig 解密配置值
func (eh *EncryptionHelper) DecryptConfig(value string, isEncrypted bool) (string, error) {
	if !isEncrypted {
		return value, nil
	}
	return eh.encryptor.Decrypt(value)
}

// GenerateAPIKey 生成 API 密钥
func (eh *EncryptionHelper) GenerateAPIKey() (string, error) {
	// 生成 32 字节的随机数
	randomBytes, err := GenerateRandomBytes(32)
	if err != nil {
		return "", err
	}

	// 使用 base64 编码
	apiKey := base64.URLEncoding.EncodeToString(randomBytes)

	// 移除填充字符
	apiKey = strings.TrimRight(apiKey, "=")

	return apiKey, nil
}

// GenerateSecretToken 生成密钥令牌
func (eh *EncryptionHelper) GenerateSecretToken(length int) (string, error) {
	return GenerateRandomKey(length)
}

// ValidateEncryptedData 验证加密数据的完整性
func (eh *EncryptionHelper) ValidateEncryptedData(encryptedData string) bool {
	if encryptedData == "" {
		return true // 空数据被认为是有效的
	}

	// 尝试解密数据来验证其完整性
	_, err := eh.encryptor.Decrypt(encryptedData)
	return err == nil
}

// RotateKey 轮换加密密钥
func (eh *EncryptionHelper) RotateKey(newConfig EncryptionConfig) error {
	newEncryptor, err := NewEncryptor(newConfig)
	if err != nil {
		return fmt.Errorf("failed to create new encryptor: %w", err)
	}

	eh.encryptor = newEncryptor
	return nil
}

// ReencryptData 使用新密钥重新加密数据
func (eh *EncryptionHelper) ReencryptData(oldEncryptedData string, oldEncryptor *Encryptor) (string, error) {
	// 使用旧加密器解密数据
	plaintext, err := oldEncryptor.Decrypt(oldEncryptedData)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt with old key: %w", err)
	}

	// 使用新加密器加密数据
	newEncryptedData, err := eh.encryptor.Encrypt(plaintext)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt with new key: %w", err)
	}

	return newEncryptedData, nil
}

// SecureCompare 安全比较两个字符串（防止时序攻击）
func SecureCompare(a, b string) bool {
	if len(a) != len(b) {
		return false
	}

	var result byte
	for i := 0; i < len(a); i++ {
		result |= a[i] ^ b[i]
	}

	return result == 0
}

// SecureCompareBytes 安全比较两个字节切片
func SecureCompareBytes(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}

	var result byte
	for i := 0; i < len(a); i++ {
		result |= a[i] ^ b[i]
	}

	return result == 0
}
