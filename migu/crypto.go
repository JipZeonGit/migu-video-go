package migu

import (
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

const (
	defaultAESKey = "MQDUjI19MGe3BhaqTlpc9g=="
	defaultAESIV  = "abcdefghijklmnop"
	rsaPrivateKeyPKCS8 = "MIICdQIBADANBgkqhkiG9w0BAQEFAASCAl8wggJbAgEAAoGBAOhvWsrglBpQGpjB" +
		"8okxLUCaaiKKOytn9EtvytB5tKDchmgkSaXpreWcDy/9imsuOiVCSdBr6hHjrTN7" +
		"QKkA4/QYS8ptiFv1ap61PiAyRFDI1b8wp2haJ6HF1rDShG2XdfWIhLk4Hj6efVZA" +
		"Sfa3taM7C8NseWoWh05Cp26g4hXZAgMBAAECgYBzqZXghsisH1hc04ZBRrth/nT6" +
		"Ixc2jlA+ia6+9xEvSw2HHSeY7COgsnvMQbpzg1lj2QyqLkkYBdfWWmrerpa/mb7j" +
		"m6w95YKs5Ndii8NhFWvC0eGK8Ygt02DeLohmkQu3B+Yq8JszjB7tQJRR2kdG6cPt" +
		"Kp99ZTyyPom/9uD+AQJBAPxCwajHAkCuH4+aKdZhH6n7oDAxZoMH/mihDRxHZJof" +
		"nT+K662QCCIx0kVCl64s/wZ4YMYbP8/PWDvLMNNWC7ECQQDr4V23KRT9fAPAN8vB" +
		"q2NqjLAmEx+tVnd4maJ16Xjy5Q4PSRiAXYLSr9uGtneSPP2fd/tja0IyawlP5UPL" +
		"l76pAkAeXqMWAK+CvfPKxBKZXqQDQOnuI2RmDgZQ7mK3rtirvXae+ciZ4qc4Bqt7" +
		"7yJ3s68YRlHQR+OMzzeeKz47kzZhAkAPteH1ChJw06q4Sb8TdiPX++jbkFiCxgiN" +
		"CsaMTfGVU/Y8xGSSYCgPelEHxu1t2wwVa/tdYs505zYmkSGT1NaJAkBCS5hymXsA" +
		"B92Fx8eGW5WpLfnpvxl8nOcP+eNXobi8Sc6q1FmoHi8snbcmBhidcDdcieKn+DbX" +
		"GG3BQE/OCOkM"
)

// GetStringMD5 returns the lowercase hex MD5 hash of the input string.
func GetStringMD5(s string) string {
	h := md5.Sum([]byte(s))
	return hex.EncodeToString(h[:])
}

// Base64Encrypt encodes a string to base64.
func Base64Encrypt(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

// Base64Decrypt decodes a base64 string.
func Base64Decrypt(s string) (string, error) {
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func padKey(key []byte, size int) []byte {
	if len(key) >= size {
		return key[:size]
	}
	padded := make([]byte, size)
	copy(padded, key)
	return padded
}

// AESEncrypt encrypts data using AES-256-CBC with PKCS7 padding.
func AESEncrypt(data string, baseKey string, ivStr string) (string, error) {
	if baseKey == "" {
		baseKey = defaultAESKey
	}
	if ivStr == "" {
		ivStr = defaultAESIV
	}
	key := padKey([]byte(baseKey), 32)
	iv := padKey([]byte(ivStr), 16)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("aes cipher: %w", err)
	}
	plaintext := pkcs7Pad([]byte(data), aes.BlockSize)
	mode := cipher.NewCBCEncrypter(block, iv)
	ciphertext := make([]byte, len(plaintext))
	mode.CryptBlocks(ciphertext, plaintext)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// AESDecrypt decrypts AES-256-CBC encrypted base64 data.
func AESDecrypt(baseData string, baseKey string, ivStr string) (string, error) {
	if baseKey == "" {
		baseKey = defaultAESKey
	}
	if ivStr == "" {
		ivStr = defaultAESIV
	}
	key := padKey([]byte(baseKey), 32)
	iv := padKey([]byte(ivStr), 16)

	ciphertext, err := base64.StdEncoding.DecodeString(baseData)
	if err != nil {
		// Node.js 版本使用 UTF-8 编码的密文，不是 base64
		ciphertext = []byte(baseData)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("aes cipher: %w", err)
	}
	if len(ciphertext)%aes.BlockSize != 0 {
		return "", fmt.Errorf("ciphertext is not a multiple of the block size")
	}
	mode := cipher.NewCBCDecrypter(block, iv)
	plaintext := make([]byte, len(ciphertext))
	mode.CryptBlocks(plaintext, ciphertext)
	plaintext, err = pkcs7Unpad(plaintext, aes.BlockSize)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

// RSAEncrypt encrypts data using RSA private key (PKCS#1 padding).
func RSAEncrypt(data string, privateKeyBase string) (string, error) {
	if privateKeyBase == "" {
		privateKeyBase = rsaPrivateKeyPKCS8
	}
	// 移除 \r (与 JS 版本一致)
	clearKey := removeCR(privateKeyBase)
	keyBytes, err := base64.StdEncoding.DecodeString(clearKey)
	if err != nil {
		return "", fmt.Errorf("base64 decode private key: %w", err)
	}
	privateKey, err := x509.ParsePKCS8PrivateKey(keyBytes)
	if err != nil {
		return "", fmt.Errorf("parse PKCS8 private key: %w", err)
	}
	rsaKey, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		return "", fmt.Errorf("not an RSA private key")
	}
	// JS 使用 crypto.privateEncrypt (私钥+type1 padding)，对应 Go 的 SignPKCS1v15
	// 当 hash=crypto.Hash(0) 时，Go 跳过 DigestInfo 前缀，直接对原始消息做 type1 padding 并私钥运算
	encrypted, err := rsa.SignPKCS1v15(nil, rsaKey, crypto.Hash(0), []byte(data))
	if err != nil {
		return "", fmt.Errorf("rsa sign: %w", err)
	}
	return base64.StdEncoding.EncodeToString(encrypted), nil
}

func removeCR(s string) string {
	result := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		if s[i] != '\r' {
			result = append(result, s[i])
		}
	}
	return string(result)
}

func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padtext := make([]byte, padding)
	for i := range padtext {
		padtext[i] = byte(padding)
	}
	return append(data, padtext...)
}

func pkcs7Unpad(data []byte, blockSize int) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty data")
	}
	padding := int(data[len(data)-1])
	if padding > blockSize || padding == 0 {
		return nil, fmt.Errorf("invalid padding")
	}
	for i := len(data) - padding; i < len(data); i++ {
		if data[i] != byte(padding) {
			return nil, fmt.Errorf("invalid padding")
		}
	}
	return data[:len(data)-padding], nil
}
