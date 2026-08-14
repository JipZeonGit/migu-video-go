package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	UserID          string
	Token           string
	Port            int
	Host            string
	RateType        int
	RefreshToken    bool
	Debug           bool
	Pass            string
	EnableHDR       bool
	EnableH265      bool
	UpdateInterval  int
	IgnoreCategory  []string
	MergeTVCategory bool
	CustomMergeCat  []string
}

func Load() *Config {
	cfg := &Config{
		UserID:          os.Getenv("muserId"),
		Token:           os.Getenv("mtoken"),
		Port:            getEnvInt("mport", 1234),
		Host:            os.Getenv("mhost"),
		RateType:        getEnvInt("mrateType", 3),
		RefreshToken:    getEnvBool("mrefreshToken"),
		Debug:           getEnvBool("mdebug"),
		Pass:            os.Getenv("mpass"),
		EnableHDR:       !getEnvEqual("menableHDR", "false"),
		EnableH265:      !getEnvEqual("menableH265", "false"),
		UpdateInterval:  getEnvInt("mupdateInterval", 6),
		MergeTVCategory: !getEnvEqual("mmergeTVCategory", "false"),
	}
	if v := os.Getenv("mignoreCategory"); v != "" {
		cfg.IgnoreCategory = splitAndTrim(v)
	}
	if v := os.Getenv("mcustomMergeCategory"); v != "" {
		cfg.CustomMergeCat = splitAndTrim(v)
	}
	return cfg
}

func getEnvInt(key string, defaultVal int) int {
	v := os.Getenv(key)
	if v == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return defaultVal
	}
	return n
}

func getEnvBool(key string) bool {
	v := os.Getenv(key)
	return v != "" && v != "false" && v != "0"
}

func getEnvEqual(key, val string) bool {
	return os.Getenv(key) == val
}

func splitAndTrim(s string) []string {
	// 支持中英文逗号
	s = strings.ReplaceAll(s, "，", ",")
	parts := strings.Split(s, ",")
	var result []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}
