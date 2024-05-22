package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"os"
)

// Utils is a utility structure that holds the AES key.
type Utils struct {
	aesKey []byte
}

// NewUtils creates a new Utils instance with the AES key from the environment variable.
// The AES key must be exactly 32 bytes (256 bits) long.
func NewUtils() (*Utils, error) {
	key := os.Getenv("AES_KEY")
	if len(key) != 32 {
		return nil, errors.New("AES key must be exactly 32 bytes (256 bits) long")
	}
	return &Utils{
		aesKey: []byte(key),
	}, nil
}

// Encrypt encrypts the given data using AES encryption in CTR mode.
// A random Initialization Vector (IV) is generated for each encryption operation.
// The IV is prepended to the ciphertext and base64 encoded.
func (u *Utils) Encrypt(data []byte) (string, error) {
	block, err := aes.NewCipher(u.aesKey)
	if err != nil {
		return "", err
	}

	// Create a byte slice to hold the IV and the ciphertext.
	cipherText := make([]byte, aes.BlockSize+len(data))

	// Generate a random IV.
	iv := cipherText[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	// Create a new CTR stream cipher using the block cipher and the IV.
	stream := cipher.NewCTR(block, iv)

	// Encrypt the data using the CTR stream cipher.
	stream.XORKeyStream(cipherText[aes.BlockSize:], data)

	// Encode the ciphertext to base64.
	return base64.StdEncoding.EncodeToString(cipherText), nil
}

// Decrypt decrypts the given base64 encoded data using AES encryption in CTR mode.
// The IV is extracted from the ciphertext and used to initialize the cipher.
func (u *Utils) Decrypt(data string) ([]byte, error) {
	// Decode the base64 encoded data.
	cipherText, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}

	// Ensure the ciphertext is at least as long as the AES block size.
	if len(cipherText) < aes.BlockSize {
		return nil, errors.New("cipherText too short")
	}

	block, err := aes.NewCipher(u.aesKey)
	if err != nil {
		return nil, err
	}

	// Extract the IV from the beginning of the ciphertext.
	iv := cipherText[:aes.BlockSize]

	// Create a byte slice to hold the plaintext.
	plainText := make([]byte, len(cipherText)-aes.BlockSize)

	// Create a new CTR stream cipher using the block cipher and the IV.
	stream := cipher.NewCTR(block, iv)

	// Decrypt the data using the CTR stream cipher.
	stream.XORKeyStream(plainText, cipherText[aes.BlockSize:])

	return plainText, nil
}
