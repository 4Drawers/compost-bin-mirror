package two_factor_auth_service

import (
	"compost-bin/logger"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"
	"os"
)

func Encrypt(plain string) string {
	block, err := aes.NewCipher([]byte(getEnv("MASTER_KEY", "")))
	if err != nil {
		logger.Fatalf("Counldn't create cipher block: %v", err)
		return ""
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		logger.Fatalf("Failed to initialize gcm algorithm: %v", err)
		return ""
	}

	nonce := make([]byte, gcm.NonceSize())
	io.ReadFull(rand.Reader, nonce)

	cipherText := gcm.Seal(nonce, nonce, []byte(plain), nil)

	return base64.StdEncoding.EncodeToString(cipherText)
}

func Decrypt(cryptoText string) string {
	cipherText, err := base64.StdEncoding.DecodeString(cryptoText)
	if err != nil {
		logger.Fatalf("Failed to decode base64 string: %v", err)
		return ""
	}

	block, err := aes.NewCipher([]byte(getEnv("MASTER_KEY", "")))
	if err != nil {
		logger.Fatalf("Failed to create cipher block: %v", err)
		return ""
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		logger.Fatalf("Failed to initialize gcm algorithm: %v", err)
		return ""
	}

	nonceSize := gcm.NonceSize()
	nonce, cipherText := cipherText[:nonceSize], cipherText[nonceSize:]

	plain, err := gcm.Open(nil, nonce, cipherText, nil)
	if err != nil {
		logger.Fatalf("Failed to decrypt: %v", err)
		return ""
	}

	return string(plain)
}

func getEnv(variableName, defaultValue string) string {
	if value := os.Getenv(variableName); value != "" {
		return value
	}
	return defaultValue
}
