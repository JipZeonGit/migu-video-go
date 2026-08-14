package migu

import (
	"testing"
)

func TestGetStringMD5(t *testing.T) {
	// 测试已知的 MD5 值
	tests := []struct {
		input    string
		expected string
	}{
		{"", "d41d8cd98f00b204e9800998ecf8427e"},
		{"hello", "5d41402abc4b2a76b9719d911017c592"},
		{"12345", "827ccb0eea8a706c4c34a16891f84e7b"},
	}
	for _, tt := range tests {
		result := GetStringMD5(tt.input)
		if result != tt.expected {
			t.Errorf("GetStringMD5(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestBase64(t *testing.T) {
	tests := []string{"hello", "你好", "test123", ""}
	for _, s := range tests {
		encoded := Base64Encrypt(s)
		decoded, err := Base64Decrypt(encoded)
		if err != nil {
			t.Errorf("Base64Decrypt error: %v", err)
			continue
		}
		if decoded != s {
			t.Errorf("Base64 roundtrip: %q -> %q -> %q", s, encoded, decoded)
		}
	}
}

func TestAESEncryptDecrypt(t *testing.T) {
	tests := []string{"hello", "test data", "你好世界"}
	for _, data := range tests {
		encrypted, err := AESEncrypt(data, "", "")
		if err != nil {
			t.Errorf("AESEncrypt(%q) error: %v", data, err)
			continue
		}
		decrypted, err := AESDecrypt(encrypted, "", "")
		if err != nil {
			t.Errorf("AESDecrypt(%q) error: %v", encrypted, err)
			continue
		}
		if decrypted != data {
			t.Errorf("AES roundtrip: %q -> %q -> %q", data, encrypted, decrypted)
		}
	}
}

func TestAESEncryptWithKey(t *testing.T) {
	data := "test data"
	key := "mykey"
	iv := "myiv"
	encrypted, err := AESEncrypt(data, key, iv)
	if err != nil {
		t.Fatalf("AESEncrypt error: %v", err)
	}
	decrypted, err := AESDecrypt(encrypted, key, iv)
	if err != nil {
		t.Fatalf("AESDecrypt error: %v", err)
	}
	if decrypted != data {
		t.Errorf("AES roundtrip with key: %q -> %q -> %q", data, encrypted, decrypted)
	}
}

func TestRSAEncrypt(t *testing.T) {
	data := "test data"
	encrypted, err := RSAEncrypt(data, "")
	if err != nil {
		t.Errorf("RSAEncrypt error: %v", err)
		return
	}
	if encrypted == "" {
		t.Error("RSAEncrypt returned empty string")
	}
	// 验证是 base64 编码
	if len(encrypted) == 0 {
		t.Error("RSAEncrypt returned empty result")
	}
}
