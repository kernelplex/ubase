package ubsecurity

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"strings"
)

type EncryptionServiceImpl struct {
	key []byte
}

type EncryptionService interface {
	Encrypt(data []byte) ([]byte, error)
	Decrypt(data []byte) ([]byte, error)
	Encrypt64(data string) (string, error)
	Decrypt64(data string) ([]byte, error)
}

func NewEncryptionService(key []byte) EncryptionService {
	// Validate key size - must be 16, 24 or 32 bytes for AES
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		panic(fmt.Errorf("invalid key size - must be 16, 24 or 32 bytes"))
	}
	service := EncryptionServiceImpl{
		key: key,
	}
	return service
}

func (s EncryptionServiceImpl) Encrypt(data []byte) ([]byte, error) {
	if s.key == nil {
		return nil, fmt.Errorf("invalid encryption service: nil key")
	}
	encrypted, err := Encrypt(s.key, data)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt data: %w", err)
	}
	return encrypted, nil
}

func (s EncryptionServiceImpl) Decrypt(data []byte) ([]byte, error) {
	decrypted, err := Decrypt(s.key, data)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data: %w", err)
	}
	return decrypted, nil
}

func (s EncryptionServiceImpl) Encrypt64(data string) (string, error) {
	encrypted, err := Encrypt64(s.key, []byte(data))
	if err != nil {
		return "", fmt.Errorf("failed to encrypt64 data: %w", err)
	}
	return encrypted, nil
}

func (s EncryptionServiceImpl) Decrypt64(data string) ([]byte, error) {
	decrypted, err := Decrypt64(s.key, data)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt64 data: %w", err)
	}
	return decrypted, nil
}

func Encrypt(key []byte, data []byte) ([]byte, error) {

	c, err := aes.NewCipher(key)

	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	out := gcm.Seal(nonce, nonce, data, nil)

	return out, nil
}

func Decrypt(key []byte, data []byte) ([]byte, error) {
    c, err := aes.NewCipher(key)

    if err != nil {
        return nil, fmt.Errorf("failed to create AES cipher: %w", err)
    }

    gcm, err := cipher.NewGCM(c)
    if err != nil {
        return nil, fmt.Errorf("failed to create GCM: %w", err)
    }

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, fmt.Errorf("data is too small - must be at least %d bytes", nonceSize)
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt with GCM: %w", err)
	}

	return plaintext, nil
}

func Encrypt64(key []byte, data []byte) (string, error) {
	encrypted, err := Encrypt(key, data)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt data: %w", err)
	}

	encoded := base64.StdEncoding.EncodeToString(encrypted)
	return encoded, nil
}

func Decrypt64(key []byte, data string) ([]byte, error) {
	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 data: %w", err)
	}
	decrypted, err := Decrypt(key, decoded)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt decoded data: %w", err)
	}
	return decrypted, nil
}

// Utility to generate a secure random bytes.
func GenerateSecureRandom(length uint32) []byte {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		panic(fmt.Errorf("failed to generate secure random bytes: %w", err)) // This should not happen since 'rand' is initialized with 'crypto/rand'
	}
	return bytes
}

func GenerateSecureRandomStringWithChars(length uint32, chars []rune) string {
	var b strings.Builder
	randBytes := GenerateSecureRandom(length)
	for _, c := range randBytes {
		b.WriteRune(chars[c%uint8(len(chars))])
	}
	return b.String()
}

func GenerateSecureRandomString(length uint32) string {
	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	return GenerateSecureRandomStringWithChars(length, chars)
}
