package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"os"
)

type Utils struct {
	aesKey []byte
}

func NewUtils() *Utils {
	key := os.Getenv("AES_KEY")
	if len(key) != 32 {
		panic("AES key must be exactly 32 bytes (256 bits) long")
	}
	return &Utils{
		aesKey: []byte(key),
	}
}

// Encrypt encrypts the given data using AES encryption.
func (u *Utils) Encrypt(data []byte) (string, error) {
	block, err := aes.NewCipher(u.aesKey)
	if err != nil {
		return "", err
	}
	cipherText := make([]byte, aes.BlockSize+len(data))
	iv := cipherText[:aes.BlockSize]
	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(cipherText[aes.BlockSize:], data)
	return base64.StdEncoding.EncodeToString(cipherText), nil
}

// Decrypt decrypts the given data using AES encryption.
func (u *Utils) Decrypt(data string) ([]byte, error) {
	cipherText, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(u.aesKey)
	if err != nil {
		return nil, err
	}
	plainText := make([]byte, len(cipherText)-aes.BlockSize)
	iv := cipherText[:aes.BlockSize]
	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(plainText, cipherText[aes.BlockSize:])
	return plainText, nil
}
