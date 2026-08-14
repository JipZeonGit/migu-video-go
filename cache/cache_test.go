package cache

import (
	"testing"
	"time"
)

func TestCacheSetGet(t *testing.T) {
	c := New()

	// 设置缓存
	c.Set("key1", "value1", nil, 1*time.Hour)

	// 获取缓存
	entry, ok := c.Get("key1")
	if !ok {
		t.Fatal("expected cache hit")
	}
	if entry.Value != "value1" {
		t.Errorf("expected 'value1', got %q", entry.Value)
	}
}

func TestCacheExpiry(t *testing.T) {
	c := New()

	// 设置短 TTL
	c.Set("key1", "value1", nil, 10*time.Millisecond)

	// 立即获取应该命中
	_, ok := c.Get("key1")
	if !ok {
		t.Fatal("expected cache hit immediately after set")
	}

	// 等待过期
	time.Sleep(20 * time.Millisecond)

	_, ok = c.Get("key1")
	if ok {
		t.Fatal("expected cache miss after expiry")
	}
}

func TestCacheMiss(t *testing.T) {
	c := New()

	_, ok := c.Get("nonexistent")
	if ok {
		t.Fatal("expected cache miss for nonexistent key")
	}
}

func TestCacheOverwrite(t *testing.T) {
	c := New()

	c.Set("key1", "value1", nil, 1*time.Hour)
	c.Set("key1", "value2", nil, 1*time.Hour)

	entry, ok := c.Get("key1")
	if !ok {
		t.Fatal("expected cache hit")
	}
	if entry.Value != "value2" {
		t.Errorf("expected 'value2', got %q", entry.Value)
	}
}
