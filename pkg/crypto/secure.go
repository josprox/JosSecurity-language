package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"io"

	"golang.org/x/crypto/pbkdf2"
)

// DeriveKey generates a 32-byte key from a password and salt using PBKDF2.
func DeriveKey(password, salt []byte) []byte {
	return pbkdf2.Key(password, salt, 100000, 32, sha256.New)
}

// EncryptAES encrypts data using AES-256-GCM.
// It returns nonce + ciphertext.
func EncryptAES(data, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, data, nil), nil
}

// DecryptAES decrypts data using AES-256-GCM.
// It expects data to be nonce + ciphertext.
func DecryptAES(data, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

// ObfuscateKey XORs the key with a simple mask to avoid plain strings in binary.
// This is NOT encryption, just obfuscation to prevent `strings` command from finding it easily.
func ObfuscateKey(key []byte) []byte {
	mask := []byte{0x5A, 0xA5, 0x12, 0x34} // Simple repeating pattern
	obfuscated := make([]byte, len(key))
	for i, b := range key {
		obfuscated[i] = b ^ mask[i%len(mask)]
	}
	return obfuscated
}

// DeobfuscateKey reverses ObfuscateKey.
func DeobfuscateKey(obfuscated []byte) []byte {
	return ObfuscateKey(obfuscated) // XOR is its own inverse
}
