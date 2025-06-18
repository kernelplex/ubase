// Security related tools.
package ubsecurity

import (
	"bytes"
	"encoding/base64"
	"slices"

	"golang.org/x/crypto/argon2"
)

// Use to generate hashes using argon2id.
type Argon2IdHashGenerator struct {
	// Amount of memory to use
	Memory uint32

	// Number of argon2id iterations
	Iterations uint32

	// Amount of threads or parallelism.
	Parallelism uint8

	// Length of the salt.
	SaltLength uint32

	// Length of the resulting key.
	Keylength uint32

	// If set, this value is appended to the password/target before hashing.
	Pepper []byte
}

// An Argon2IdHashGenerator with default settings which can be used as is.
var DefaultArgon2Id = Argon2IdHashGenerator{
	Memory:      64 * 1024,
	Iterations:  3,
	Parallelism: 2,
	SaltLength:  16,
	Keylength:   32,
	Pepper:      nil,
}

// Base interface for hash generators.
type HashGenerator interface {
	GenerateHashBytes(string) []byte
	GenerateHashBase64(string) (string, error)
	VerifyBase64(string, string) (bool, error)
	Verify([]byte, []byte) bool
}

// Generates a hash of the specified string.
func (i Argon2IdHashGenerator) GenerateHashBytes(target string) []byte {
	saltBytes := GenerateSecureRandom(i.SaltLength)

	var targetBytes []byte = []byte(target)
	hashedBytes := i.GenerateHashWithSalt(targetBytes, saltBytes)
	return hashedBytes
}

// Generates the raw hash without including the salt in the output bytes.
func (i Argon2IdHashGenerator) GenerateKeyFromBytes(targetBytes []byte, length uint32, saltBytes []byte) []byte {
	if i.Pepper != nil {
		targetBytes = slices.Concat(targetBytes, i.Pepper)
	}
	return argon2.IDKey(targetBytes, saltBytes, i.Iterations, i.Memory, i.Parallelism, length)
}

// Genrate bytes with the specified salt.
// The salt is the first n bytes of the returned value
func (i Argon2IdHashGenerator) GenerateHashWithSalt(targetBytes []byte, saltBytes []byte) []byte {
	hashedBytes := i.GenerateKeyFromBytes(targetBytes, i.Keylength, saltBytes)
	return slices.Concat(saltBytes, hashedBytes)
}

// Generates the hash of a string and returns the base64 encoded hash.
func (i Argon2IdHashGenerator) GenerateHashBase64(target string) (string, error) {
	hashedBytes := i.GenerateHashBytes(target)
	hashedString := base64.StdEncoding.EncodeToString(hashedBytes)
	return hashedString, nil
}

// Verifies the specified target bytes against the hash.
func (i Argon2IdHashGenerator) Verify(target []byte, hashed []byte) bool {
	saltBytes := hashed[:i.SaltLength]
	generated := i.GenerateHashWithSalt([]byte(target), saltBytes)
	return bytes.Equal(generated, hashed)
}

// Verifies the specified target against a hash encoded in base64.
func (i Argon2IdHashGenerator) VerifyBase64(target string, base64Hash string) (bool, error) {
	hash, err := base64.StdEncoding.DecodeString(base64Hash)
	if err != nil {
		return false, err
	}
	return i.Verify([]byte(target), hash), nil
}
