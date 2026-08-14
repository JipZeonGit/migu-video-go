package migu

import (
	"strconv"
	"strings"
	"time"
)

type ddCalcuConfig struct {
	keys              string
	words             [4]string
	thirdReplaceIndex int
	suffix            string
}

var ddCalcuList = map[string]*ddCalcuConfig{
	"h5": {
		keys:              "yzwxcdabgh",
		words:             [4]string{"", "y", "0", "w"},
		thirdReplaceIndex: 1,
		suffix:            "&sv=10000&ct=www",
	},
	"android": {
		keys:              "cdabyzwxkl",
		words:             [4]string{"v", "a", "0", "a"},
		thirdReplaceIndex: 6,
		suffix:            "&sv=10004&ct=android",
	},
}

// GetddCalcu calculates the ddCalcu signature for a given puData.
func GetddCalcu(puData, programId, clientType string, rateType int, urlUserId string) string {
	if puData == "" || programId == "" {
		return ""
	}
	if clientType != "android" && clientType != "h5" {
		return ""
	}

	cfg, ok := ddCalcuList[clientType]
	if !ok {
		return ""
	}

	// 复制 words 以避免修改全局配置
	var words [4]string
	copy(words[:], cfg.words[:])

	// 确定用户 ID
	id := urlUserId

	// 根据用户 ID 修改 words[0]
	if id != "" && len(id) > 7 {
		idx := int(id[7] - '0')
		if idx >= 0 && idx < len(cfg.keys) {
			words[0] = string(cfg.keys[idx])
		}
	}

	keys := cfg.keys
	thirdReplaceIndex := cfg.thirdReplaceIndex

	// android 平台标清
	if clientType == "android" && rateType == 2 {
		words[0] = "v"
	}
	// 短用户 ID
	if len(id) > 3 && len(id) <= 8 {
		words[0] = "e"
	}

	puDataLength := len(puData)
	var ddCalcu strings.Builder
	for i := 0; i < puDataLength/2; i++ {
		ddCalcu.WriteByte(puData[puDataLength-i-1])
		ddCalcu.WriteByte(puData[i])
		switch i {
		case 1:
			ddCalcu.WriteString(words[0])
		case 2:
			dateStr := getDateString(time.Now())
			idx := int(dateStr[0] - '0')
			if idx >= 0 && idx < len(keys) {
				ddCalcu.WriteByte(keys[idx])
			}
		case 3:
			if thirdReplaceIndex < len(programId) {
				idx := int(programId[thirdReplaceIndex] - '0')
				if idx >= 0 && idx < len(keys) {
					ddCalcu.WriteByte(keys[idx])
				}
			}
		case 4:
			ddCalcu.WriteString(words[1])
		}
	}
	return ddCalcu.String()
}

// GetddCalcuURL calculates the ddCalcu URL for a given puDataURL.
func GetddCalcuURL(puDataURL, programId, clientType string, rateType int, urlUserId string) string {
	if puDataURL == "" || programId == "" {
		return ""
	}
	if clientType != "android" && clientType != "h5" {
		return ""
	}

	parts := strings.Split(puDataURL, "&puData=")
	if len(parts) < 2 {
		return ""
	}
	puData := parts[1]
	ddCalcu := GetddCalcu(puData, programId, clientType, rateType, urlUserId)
	suffix := ddCalcuList[clientType].suffix
	return puDataURL + "&ddCalcu=" + ddCalcu + suffix
}

// GetddCalcu720p calculates the ddCalcu signature for 720p mode.
func GetddCalcu720p(puData, programId string) string {
	if puData == "" || programId == "" {
		return ""
	}
	keys := "cdabyzwxkl"

	puDataLength := len(puData)
	var ddCalcu strings.Builder
	for i := 0; i < puDataLength/2; i++ {
		ddCalcu.WriteByte(puData[puDataLength-i-1])
		ddCalcu.WriteByte(puData[i])
		switch i {
		case 1:
			ddCalcu.WriteByte('v')
		case 2:
			dateStr := getDateString(time.Now())
			idx := int(dateStr[2] - '0')
			if idx >= 0 && idx < len(keys) {
				ddCalcu.WriteByte(keys[idx])
			}
		case 3:
			if len(programId) > 6 {
				idx := int(programId[6] - '0')
				if idx >= 0 && idx < len(keys) {
					ddCalcu.WriteByte(keys[idx])
				}
			}
		case 4:
			ddCalcu.WriteByte('a')
		}
	}
	return ddCalcu.String()
}

// GetddCalcuURL720p calculates the ddCalcu URL for 720p mode.
func GetddCalcuURL720p(puDataURL, programId string) string {
	if puDataURL == "" || programId == "" {
		return ""
	}
	parts := strings.Split(puDataURL, "&puData=")
	if len(parts) < 2 {
		return ""
	}
	puData := parts[1]
	ddCalcu := GetddCalcu720p(puData, programId)
	return puDataURL + "&ddCalcu=" + ddCalcu + "&sv=10004&ct=android"
}

func getDateString(t time.Time) string {
	return t.Format("20060102")
}

func getTimeString(t time.Time) string {
	return t.Format("150405")
}

func getDateTimeString(t time.Time) string {
	return t.Format("20060102150405")
}

func getDateTimeStr(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

func getLogDateTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05.000")
}

// parseInt parses a string to int, returns 0 on error.
func parseInt(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}
