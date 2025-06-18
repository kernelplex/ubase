package ubsecurity

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

type EncryptionServiceX struct {
	key []byte
}

type EncryptionService interface {
	Encrypt(data []byte) ([]byte, error)
	Decrypt(data []byte) ([]byte, error)
	Encrypt64(data string) (string, error)
	Decrypt64(data string) ([]byte, error)
}

func CreateEncryptionService(key []byte) EncryptionService {
	// Validate key size - must be 16, 24 or 32 bytes for AES
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return nil
	}
	service := EncryptionServiceX{
		key: key,
	}
	return service
}

func (s EncryptionServiceX) Encrypt(data []byte) ([]byte, error) {
	if s.key == nil {
		return nil, fmt.Errorf("invalid encryption service: nil key")
	}
	return Encrypt(s.key, data)
}

func (s EncryptionServiceX) Decrypt(data []byte) ([]byte, error) {
	return Decrypt(s.key, data)
}

func (s EncryptionServiceX) Encrypt64(data string) (string, error) {
	return Encrypt64(s.key, []byte(data))
}

func (s EncryptionServiceX) Decrypt64(data string) ([]byte, error) {
	return Decrypt64(s.key, data)
}

func Encrypt(key []byte, data []byte) ([]byte, error) {

	c, err := aes.NewCipher(key)

	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	out := gcm.Seal(nonce, nonce, data, nil)

	return out, nil
}

func Decrypt(key []byte, data []byte) ([]byte, error) {
	c, err := aes.NewCipher(key)

	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		err = fmt.Errorf("Data is too small - must be at lease %d", nonceSize)
		return nil, err
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

func Encrypt64(key []byte, data []byte) (string, error) {
	encrypted, err := Encrypt(key, data)
	if err != nil {
		return "", err
	}

	encoded := base64.StdEncoding.EncodeToString(encrypted)
	return encoded, nil
}

func Decrypt64(key []byte, data string) ([]byte, error) {
	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}
	decrypted, err := Decrypt(key, decoded)
	if err != nil {
		return nil, err
	}
	return decrypted, nil
}

// Utility to generate a secure random bytes.
func GenerateSecureRandom(length uint32) []byte {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		panic(err) // This should not happen since 'rand' is initialized with 'crypto/rand'
	}
	return bytes
}
