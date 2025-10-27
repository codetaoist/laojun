package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/scrypt"
)

// EncryptionConfig encryption configuration
type EncryptionConfig struct {
	SecretKey string
	Salt      string
}

// Encryptor encryption tool
type Encryptor struct {
	key []byte
}

// NewEncryptor creates a new encryption tool
func NewEncryptor(config EncryptionConfig) (*Encryptor, error) {
	// Use scrypt to generate fixed-length key from secret key and salt
	key, err := scrypt.Key([]byte(config.SecretKey), []byte(config.Salt), 32768, 8, 1, 32)
	if err != nil {
		return nil, fmt.Errorf("failed to derive key: %w", err)
	}

	return &Encryptor{key: key}, nil
}

// Encrypt encrypts data
func (e *Encryptor) Encrypt(plaintext string) (string, error) {
	return e.EncryptString(plaintext)
}

// EncryptString encrypts string data
func (e *Encryptor) EncryptString(plaintext string) (string, error) {
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate random nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt data
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)

	// Return base64 encoded result
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts data
func (e *Encryptor) Decrypt(ciphertext string) (string, error) {
	return e.DecryptString(ciphertext)
}

// DecryptString decrypts string data
func (e *Encryptor) DecryptString(ciphertext string) (string, error) {
	// Decode base64
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

	nonce, ciphertext_bytes := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext_bytes, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}

// EncryptBytes encrypts byte data
func (e *Encryptor) EncryptBytes(data []byte) ([]byte, error) {
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

// DecryptBytes decrypts byte data
func (e *Encryptor) DecryptBytes(data []byte) ([]byte, error) {
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

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

// PasswordHasher password hashing tool
type PasswordHasher struct {
	cost int
}

// NewPasswordHasher creates a new password hasher
func NewPasswordHasher(cost int) *PasswordHasher {
	if cost < 4 || cost > 31 {
		cost = bcrypt.DefaultCost
	}
	return &PasswordHasher{cost: cost}
}

// HashPassword hashes password
func (ph *PasswordHasher) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), ph.cost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hash), nil
}

// CheckPassword verifies password
func (ph *PasswordHasher) CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateRandomKey generates random key
func GenerateRandomKey(length int) (string, error) {
	bytes, err := GenerateRandomBytes(length)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// GenerateRandomBytes generates random bytes
func GenerateRandomBytes(length int) ([]byte, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	return bytes, err
}

// SHA256Hash calculates SHA256 hash
func SHA256Hash(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// SHA256HashBytes calculates SHA256 hash of byte data
func SHA256HashBytes(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// EncryptionHelper encryption helper
type EncryptionHelper struct {
	encryptor      *Encryptor
	passwordHasher *PasswordHasher
}

// NewEncryptionHelper creates a new encryption helper
func NewEncryptionHelper(secretKey string) *EncryptionHelper {
	config := EncryptionConfig{
		SecretKey: secretKey,
		Salt:      "laojun-salt", // Default salt
	}
	
	encryptor, _ := NewEncryptor(config)
	passwordHasher := NewPasswordHasher(bcrypt.DefaultCost)
	
	return &EncryptionHelper{
		encryptor:      encryptor,
		passwordHasher: passwordHasher,
	}
}

// Encrypt encrypts sensitive data
func (eh *EncryptionHelper) Encrypt(data string) (string, error) {
	return eh.encryptor.Encrypt(data)
}

// Decrypt decrypts sensitive data
func (eh *EncryptionHelper) Decrypt(encryptedData string) (string, error) {
	return eh.encryptor.Decrypt(encryptedData)
}

// HashPassword hashes password
func (eh *EncryptionHelper) HashPassword(password string) (string, error) {
	return eh.passwordHasher.HashPassword(password)
}

// CheckPassword verifies password
func (eh *EncryptionHelper) CheckPassword(password, hash string) bool {
	return eh.passwordHasher.CheckPassword(password, hash)
}

// EncryptConfig encrypts configuration value
func (eh *EncryptionHelper) EncryptConfig(value string, shouldEncrypt bool) (string, error) {
	if !shouldEncrypt {
		return value, nil
	}
	return eh.encryptor.Encrypt(value)
}

// DecryptConfig decrypts configuration value
func (eh *EncryptionHelper) DecryptConfig(value string, isEncrypted bool) (string, error) {
	if !isEncrypted {
		return value, nil
	}
	return eh.encryptor.Decrypt(value)
}

// GenerateAPIKey generates API key
func (eh *EncryptionHelper) GenerateAPIKey() (string, error) {
	// Generate 32 bytes of random data
	bytes, err := GenerateRandomBytes(32)
	if err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Use base64 encoding
	apiKey := base64.URLEncoding.EncodeToString(bytes)

	// Remove padding characters
	apiKey = strings.TrimRight(apiKey, "=")

	return apiKey, nil
}

// GenerateSecretToken generates secret token
func (eh *EncryptionHelper) GenerateSecretToken(length int) (string, error) {
	return GenerateRandomKey(length)
}

// ValidateEncryptedData validates encrypted data integrity
func (eh *EncryptionHelper) ValidateEncryptedData(encryptedData string) bool {
	if encryptedData == "" {
		return true // Empty data is considered valid
	}

	// Try to decrypt data to validate its integrity
	_, err := eh.encryptor.Decrypt(encryptedData)
	return err == nil
}

// RotateKey rotates encryption key
func (eh *EncryptionHelper) RotateKey(newSecretKey string) error {
	config := EncryptionConfig{
		SecretKey: newSecretKey,
		Salt:      "laojun-salt",
	}
	
	newEncryptor, err := NewEncryptor(config)
	if err != nil {
		return fmt.Errorf("failed to create new encryptor: %w", err)
	}
	
	eh.encryptor = newEncryptor
	return nil
}

// ReencryptData re-encrypts data with new key
func (eh *EncryptionHelper) ReencryptData(oldEncryptedData string, oldEncryptor *Encryptor) (string, error) {
	// Decrypt data with old encryptor
	plaintext, err := oldEncryptor.Decrypt(oldEncryptedData)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt with old key: %w", err)
	}

	// Encrypt data with new encryptor
	newEncryptedData, err := eh.encryptor.Encrypt(plaintext)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt with new key: %w", err)
	}

	return newEncryptedData, nil
}

// SecureCompare securely compares two strings (prevents timing attacks)
func SecureCompare(a, b string) bool {
	return SecureCompareBytes([]byte(a), []byte(b))
}

// SecureCompareBytes securely compares two byte slices
func SecureCompareBytes(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	return subtle.ConstantTimeCompare(a, b) == 1
}
