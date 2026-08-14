package migu

import (
	"fmt"
	"net/url"
	"time"

	"migu-video-go/internal/log"
	"migu-video-go/internal/netutil"
	"migu-video-go/internal/ratelimit"
)

// RefreshToken refreshes the user's authentication token.
func RefreshToken(userID, token string, limiter *ratelimit.Limiter) bool {
	if userID == "" || token == "" {
		return false
	}
	if limiter != nil {
		limiter.Acquire()
	}

	timestamp := time.Now().Unix()
	baseData := fmt.Sprintf(`{"userToken":"%s","autoDelay":true,"deviceId":"","userId":"%s","timestamp":"%d"}`, token, userID, timestamp)

	// AES 加密请求体
	encryData, err := AESEncrypt(baseData, "", "")
	if err != nil {
		log.Red("AES加密失败: " + err.Error())
		return false
	}
	data := `{"data":"` + encryData + `"}`

	// RSA 签名
	md5Hash := GetStringMD5(data)
	sign, err := RSAEncrypt(md5Hash, "")
	if err != nil {
		log.Red("RSA加密失败: " + err.Error())
		return false
	}
	encodedSign := url.QueryEscape(sign)

	headers := map[string]string{
		"userId":      userID,
		"userToken":   token,
		// "appsication" 是服务端实际接受的拼写（与原 JS 行为一致）
		"Content-Type": "appsication/json; charset=utf-8",
	}

	baseURL := "https://migu-app-umnb.miguvideo.com/login/token_refresh_migu_plus"
	params := fmt.Sprintf("?clientId=27fb3129-5a54-45bc-8af1-7dc8f1155501&sign=%s&signType=RSA", encodedSign)

	var respResult TokenRefreshResponse
	err = netutil.FetchPostJSON(baseURL+params, headers, data, &respResult)
	if err != nil {
		return false
	}

	if respResult.ResultCode == "REFRESH_TOKEN_SUCCESS" {
		return true
	}
	log.Debugf("Token 刷新响应: %+v", respResult)
	return false
}
