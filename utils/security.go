package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"
	"os"
)

type Utils struct {
	masterKey        []byte
	encryptedDataKey string
}

func NewUtils() *Utils {
	masterKey := []byte(os.Getenv("MASTER_AES_KEY"))
	if len(masterKey) != 32 {
		panic("Master AES key must be exactly 32 bytes long")
	}

	u := &Utils{masterKey: masterKey}
	u.initDataKey()
	return u
}

func (u *Utils) initDataKey() {
	dataKey := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, dataKey); err != nil {
		panic("failed to generate data key: " + err.Error())
	}
	encryptedKey, err := u.encryptDataKey(dataKey)
	if err != nil {
		panic("failed to encrypt data key: " + err.Error())
	}
	u.encryptedDataKey = encryptedKey
}

func (u *Utils) encryptDataKey(dataKey []byte) (string, error) {
	block, err := aes.NewCipher(u.masterKey)
	if err != nil {
		return "", err
	}
	cipherText := make([]byte, aes.BlockSize+len(dataKey))
	iv := cipherText[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}
	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(cipherText[aes.BlockSize:], dataKey)
	return base64.StdEncoding.EncodeToString(cipherText), nil
}

func (u *Utils) Encrypt(data []byte) (string, error) {
	dataKey, err := u.decryptDataKey(u.encryptedDataKey)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(dataKey)
	if err != nil {
		return "", err
	}
	cipherText := make([]byte, aes.BlockSize+len(data))
	iv := cipherText[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}
	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(cipherText[aes.BlockSize:], data)
	return base64.StdEncoding.EncodeToString(cipherText), nil
}

func (u *Utils) Decrypt(encryptedData string) ([]byte, error) {
	dataKey, err := u.decryptDataKey(u.encryptedDataKey)
	if err != nil {
		return nil, err
	}
	cipherText, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(dataKey)
	if err != nil {
		return nil, err
	}
	plainText := make([]byte, len(cipherText)-aes.BlockSize)
	iv := cipherText[:aes.BlockSize]
	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(plainText, cipherText[aes.BlockSize:])
	return plainText, nil
}

func (u *Utils) decryptDataKey(encryptedDataKey string) ([]byte, error) {
	cipherText, err := base64.StdEncoding.DecodeString(encryptedDataKey)
	if err != nil {
		return nil, err
	}
	if len(cipherText) < aes.BlockSize {
		return nil, err
	}
	block, err := aes.NewCipher(u.masterKey)
	if err != nil {
		return nil, err
	}
	iv := cipherText[:aes.BlockSize]
	dataKey := make([]byte, len(cipherText)-aes.BlockSize)
	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(dataKey, cipherText[aes.BlockSize:])
	return dataKey, nil
}
