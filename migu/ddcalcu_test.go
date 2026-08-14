package migu

import (
	"testing"
)

func TestGetddCalcu(t *testing.T) {
	// 基本功能测试
	result := GetddCalcu("abcdefghij", "1234567890", "android", 3, "")
	if result == "" {
		t.Error("GetddCalcu returned empty string for valid input")
	}
	t.Logf("GetddCalcu result: %s", result)

	// 空输入
	result = GetddCalcu("", "1234567890", "android", 3, "")
	if result != "" {
		t.Error("GetddCalcu should return empty for empty puData")
	}

	// 无效 clientType
	result = GetddCalcu("abcdefghij", "1234567890", "invalid", 3, "")
	if result != "" {
		t.Error("GetddCalcu should return empty for invalid clientType")
	}
}

func TestGetddCalcuURL(t *testing.T) {
	result := GetddCalcuURL("http://example.com/video?token=abc&puData=abcdefghij", "1234567890", "android", 3, "")
	if result == "" {
		t.Error("GetddCalcuURL returned empty string for valid input")
	}
	t.Logf("GetddCalcuURL result: %s", result)

	// 空输入
	result = GetddCalcuURL("", "1234567890", "android", 3, "")
	if result != "" {
		t.Error("GetddCalcuURL should return empty for empty URL")
	}
}

func TestGetddCalcu720p(t *testing.T) {
	result := GetddCalcu720p("abcdefghij", "1234567890")
	if result == "" {
		t.Error("GetddCalcu720p returned empty string for valid input")
	}
	t.Logf("GetddCalcu720p result: %s", result)

	// 空输入
	result = GetddCalcu720p("", "1234567890")
	if result != "" {
		t.Error("GetddCalcu720p should return empty for empty puData")
	}
}

func TestGetddCalcuURL720p(t *testing.T) {
	result := GetddCalcuURL720p("http://example.com/video?token=abc&puData=abcdefghij", "1234567890")
	if result == "" {
		t.Error("GetddCalcuURL720p returned empty string for valid input")
	}
	t.Logf("GetddCalcuURL720p result: %s", result)
}

func TestGetddCalcuWithUserId(t *testing.T) {
	// 测试用户 ID 对结果的影响
	result1 := GetddCalcu("abcdefghij123456", "1234567890", "android", 3, "")
	result2 := GetddCalcu("abcdefghij123456", "1234567890", "android", 3, "12345678")
	if result1 == result2 {
		t.Error("GetddCalcu should produce different results with different user IDs")
	}
	t.Logf("Without userId: %s", result1)
	t.Logf("With userId:    %s", result2)
}
