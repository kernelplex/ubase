package ubsecurity

import (
	"testing"
)

func TestArgon2IdHashGenerator(t *testing.T) {
	t.Run("Default generator", func(t *testing.T) {
		generator := DefaultArgon2Id
		testPassword := "testPassword123!"

		// Test GenerateHashBytes
		hashBytes := generator.GenerateHashBytes(testPassword)
		if len(hashBytes) != int(generator.SaltLength+generator.Keylength) {
			t.Errorf("Expected hash length %d, got %d",
				generator.SaltLength+generator.Keylength, len(hashBytes))
		}

		// Test Verify
		if !generator.Verify([]byte(testPassword), hashBytes) {
			t.Error("Failed to verify correct password")
		}

		// Test Verify with wrong password
		if generator.Verify([]byte("wrongPassword"), hashBytes) {
			t.Error("Incorrectly verified wrong password")
		}

		// Test GenerateHashBase64
		hashBase64, err := generator.GenerateHashBase64(testPassword)
		if err != nil {
			t.Fatalf("GenerateHashBase64 failed: %v", err)
		}
		if len(hashBase64) == 0 {
			t.Error("Empty base64 hash returned")
		}

		// Test VerifyBase64
		valid, err := generator.VerifyBase64(testPassword, hashBase64)
		if err != nil {
			t.Fatalf("VerifyBase64 failed: %v", err)
		}
		if !valid {
			t.Error("Failed to verify correct password via base64")
		}
	})

	t.Run("With pepper", func(t *testing.T) {
		pepper := []byte("testPepper")
		generator := Argon2IdHashGenerator{
			Memory:      64 * 1024,
			Iterations:  3,
			Parallelism: 2,
			SaltLength:  16,
			Keylength:   32,
			Pepper:      pepper,
		}
		testPassword := "testPassword123!"

		hashBytes := generator.GenerateHashBytes(testPassword)
		if !generator.Verify([]byte(testPassword), hashBytes) {
			t.Error("Failed to verify correct password with pepper")
		}
	})

	t.Run("GenerateKeyFromBytes", func(t *testing.T) {
		generator := DefaultArgon2Id
		testInput := []byte("testInput")
		salt := GenerateSecureRandom(generator.SaltLength)

		key := generator.GenerateKeyFromBytes(testInput, 32, salt)
		if len(key) != 32 {
			t.Errorf("Expected key length 32, got %d", len(key))
		}
	})
}
