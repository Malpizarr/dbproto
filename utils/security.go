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
	return &Utils{
		aesKey: []byte(os.Getenv("AES_KEY")),
	}
}

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
