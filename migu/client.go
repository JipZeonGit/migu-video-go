package migu

import (
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"migu-video-go/config"
	"migu-video-go/internal/log"
	"migu-video-go/internal/netutil"
	"migu-video-go/internal/ratelimit"
)

// Client is the video site API client.
type Client struct {
	cfg      *config.Config
	clientID string
	limiter  *ratelimit.Limiter
}

// NewClient creates a new video site API client.
func NewClient(cfg *config.Config) *Client {
	return &Client{
		cfg:      cfg,
		clientID: GetStringMD5(fmt.Sprintf("%d", time.Now().UnixMilli())),
		limiter:  ratelimit.New(5), // 每秒最多 5 个请求到视频网站 API
	}
}

// PlayResult holds the result of a play URL request.
type PlayResult struct {
	URL      string
	RateType int
	Content  *PlayURLResponse
}

// Acquire blocks until a rate limiter token is available.
func (c *Client) Acquire() {
	c.limiter.Acquire()
}

// Limiter returns the client's rate limiter for external use.
func (c *Client) Limiter() *ratelimit.Limiter {
	return c.limiter
}

// GetAndroidURL gets the playback URL for a channel with user authentication.
func (c *Client) GetAndroidURL(userID, token, pid string, rateType int) (*PlayResult, error) {
	c.limiter.Acquire()
	if rateType <= 1 {
		return &PlayResult{}, nil
	}
	timestamp := time.Now().UnixMilli()
	appVersion := "26000370"
	headers := map[string]string{
		"AppVersion":                "2600037000",
		"TerminalId":                "android",
		"X-UP-CLIENT-CHANNEL-ID":   "2600037000-99000-200300220100002",
	}
	// CCTV5 和 CCTV5+ 开启 flv 后不能回放
	if pid != "641886683" && pid != "641886773" {
		headers["appCode"] = "miguvideo_default_android"
	}
	if rateType != 2 && userID != "" && token != "" {
		headers["UserId"] = userID
		headers["UserToken"] = token
	}

	str := fmt.Sprintf("%d%s%s", timestamp, pid, appVersion)
	md5 := GetStringMD5(str)
	salt := "1230024"
	suffix := "3ce941cc3cbc40528bfd1c64f9fdf6c0migu0123"
	sign := GetStringMD5(md5 + suffix)

	enableHDRStr := ""
	if c.cfg.EnableHDR {
		enableHDRStr = "&4kvivid=true&2Kvivid=true&vivid=2"
	}
	enableH265Str := ""
	if c.cfg.EnableH265 {
		enableH265Str = "&h265N=true"
	}

	baseURL := "https://play.miguvideo.com/playurl/v1/play/playurl"
	ottStr := ""
	if rateType == 9 {
		ottStr = "&ott=true"
	}
	params := fmt.Sprintf("?sign=%s&rateType=%d&contId=%s&timestamp=%d&salt=%s&flvEnable=true&super4k=true%s%s%s",
		sign, rateType, pid, timestamp, salt, ottStr, enableH265Str, enableHDRStr)

	log.Debugf("请求链接: %s%s", baseURL, params)

	var respData PlayURLResponse
	err := netutil.FetchJSON(baseURL+params, headers, &respData)
	if err != nil {
		return nil, fmt.Errorf("fetch play URL: %w", err)
	}

	// 会员降级处理
	if respData.RID == "TIPS_NEED_MEMBER" {
		log.Yellow("该账号没有会员 正在降低画质")
		respRateType := 3
		if respData.Body.URLInfo.RateType != "" && parseInt(respData.Body.URLInfo.RateType) > 4 {
			respRateType = 4
		}
		params = fmt.Sprintf("?sign=%s&rateType=%d&contId=%s&timestamp=%d&salt=%s&flvEnable=true&super4k=true%s%s",
			sign, respRateType, pid, timestamp, salt, enableH265Str, enableHDRStr)
		log.Debugf("请求链接: %s%s", baseURL, params)
		err = netutil.FetchJSON(baseURL+params, headers, &respData)
		if err != nil {
			return nil, fmt.Errorf("fetch play URL retry: %w", err)
		}

		if respData.RID == "TIPS_NEED_MEMBER" {
			log.Yellow("账号非钻石会员 降低画质")
			params = fmt.Sprintf("?sign=%s&rateType=3&contId=%s&timestamp=%d&salt=%s&flvEnable=true&super4k=true%s%s",
				sign, pid, timestamp, salt, enableH265Str, enableHDRStr)
			log.Debugf("请求链接: %s%s", baseURL, params)
			err = netutil.FetchJSON(baseURL+params, headers, &respData)
			if err != nil {
				return nil, fmt.Errorf("fetch play URL retry2: %w", err)
			}
		}
	}

	url := respData.Body.URLInfo.URL
	if url == "" {
		return &PlayResult{Content: &respData}, nil
	}
	contID := respData.Body.Content.ContID
	if contID != "" {
		pid = contID
	}

	// 加密 URL
	resURL := GetddCalcuURL(url, pid, "android", rateType, userID)
	resultRateType := parseInt(respData.Body.URLInfo.RateType)

	return &PlayResult{
		URL:      resURL,
		RateType: resultRateType,
		Content:  &respData,
	}, nil
}

// GetAndroidURL720p gets the 720p playback URL without authentication.
func (c *Client) GetAndroidURL720p(pid string) (*PlayResult, error) {
	c.limiter.Acquire()
	timestamp := fmt.Sprintf("%d", time.Now().UnixMilli())
	appVersion := "2600034600"
	appVersionID := appVersion + "-99000-201600010010028"
	headers := map[string]string{
		"AppVersion":              appVersion,
		"TerminalId":              "android",
		"X-UP-CLIENT-CHANNEL-ID": appVersionID,
		"ClientId":                c.clientID,
	}
	log.Debugf("client_id: %s", c.clientID)

	// CCTV5 和 CCTV5+ 开启 flv 后不能回放
	if pid != "641886683" && pid != "641886773" {
		headers["appCode"] = "miguvideo_default_android"
	}

	str := timestamp + pid + appVersion[:8]
	md5 := GetStringMD5(str)

	salt := fmt.Sprintf("%06d25", rand.Intn(1000000))
	suffix := "2cac4f2c6c3346a5b34e085725ef7e33migu" + salt[:4]
	sign := GetStringMD5(md5 + suffix)

	rateType := 3
	enableHDRStr := ""
	if c.cfg.EnableHDR {
		enableHDRStr = "&4kvivid=true&2Kvivid=true&vivid=2"
	}
	enableH265Str := ""
	if c.cfg.EnableH265 {
		enableH265Str = "&h265N=true"
	}

	baseURL := "https://play.miguvideo.com/playurl/v1/play/playurl"
	params := fmt.Sprintf("?sign=%s&rateType=%d&contId=%s&timestamp=%s&salt=%s&flvEnable=true&super4k=true%s%s",
		sign, rateType, pid, timestamp, salt, enableH265Str, enableHDRStr)

	log.Debugf("请求链接: %s%s", baseURL, params)
	log.Debugf("headers: %v", headers)

	var respData PlayURLResponse
	err := netutil.FetchJSON(baseURL+params, headers, &respData)
	if err != nil {
		return nil, fmt.Errorf("fetch 720p URL: %w", err)
	}

	url := respData.Body.URLInfo.URL
	if url == "" {
		return &PlayResult{Content: &respData}, nil
	}

	resultRateType := parseInt(respData.Body.URLInfo.RateType)
	contID := respData.Body.Content.ContID
	if contID != "" {
		pid = contID
	}

	// 加密 URL
	resURL := GetddCalcuURL720p(url, pid)

	return &PlayResult{
		URL:      resURL,
		RateType: resultRateType,
		Content:  &respData,
	}, nil
}

// Get302URL follows redirects to get the final stream URL.
func (c *Client) Get302URL(rawURL string) string {
	if rawURL == "" {
		return ""
	}
	client := &http.Client{
		Timeout: 6 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	for z := 1; z <= 6; z++ {
		if z >= 2 {
			log.Yellow(fmt.Sprintf("获取失败,正在第%d次重试", z-1))
		}
		req, err := http.NewRequest("GET", rawURL, nil)
		if err != nil {
			log.Red("请求失败: " + err.Error())
			continue
		}
		resp, err := client.Do(req)
		if err != nil {
			log.Red("请求超时")
			continue
		}
		location := resp.Header.Get("Location")
		resp.Body.Close()

		if location != "" && !strings.HasPrefix(location, "http://bofang") {
			return location
		}
		if z != 6 {
			time.Sleep(150 * time.Millisecond)
		}
	}
	log.Red("获取失败,返回原链接")
	return ""
}

// PrintLoginInfo logs the authentication status.
func (c *Client) PrintLoginInfo(res *PlayResult) {
	if res.Content == nil || res.Content.Body.Auth == nil {
		log.Yellow("未登录")
		return
	}
	auth := res.Content.Body.Auth
	if auth.Logined {
		log.Green("登录认证成功")
		if auth.AuthResult == "FAIL" {
			log.Red(fmt.Sprintf("认证失败 视频内容不完整 可能缺少相关VIP: %s", auth.ResultDesc))
		}
	} else {
		log.Yellow("未登录")
	}
}

// CateList fetches the list of live TV categories.
func (c *Client) CateList() ([]Category, error) {
	c.limiter.Acquire()
	var resp CateListResponse
	err := netutil.FetchJSON("https://program-sc.miguvideo.com/live/v2/tv-data/1ff892f2b5ab4a79be6e25b69d2f5d05", nil, &resp)
	if err != nil {
		return nil, err
	}
	liveList := resp.Body.LiveList
	// 过滤 "热门" 分类（重复内容）
	filtered := make([]Category, 0, len(liveList))
	for _, item := range liveList {
		if item.Name != "热门" {
			filtered = append(filtered, item)
		}
	}
	// 央视作为首个分类
	for i, item := range filtered {
		if item.Name == "央视" && i > 0 {
			filtered[0], filtered[i] = filtered[i], filtered[0]
			break
		}
	}
	return filtered, nil
}

// DataList fetches all categories with their channel data.
func (c *Client) DataList() ([]Category, error) {
	cates, err := c.CateList()
	if err != nil {
		return nil, err
	}

	for i := range cates {
		c.limiter.Acquire()
		var resp DataListResponse
		err := netutil.FetchJSON("https://program-sc.miguvideo.com/live/v2/tv-data/"+cates[i].VomsID, nil, &resp)
		if err != nil {
			cates[i].DataList = nil
		} else {
			cates[i].DataList = resp.Body.DataList
		}
	}

	// 去重
	cates = uniqueData(cates)
	// 合并分类
	if c.cfg.MergeTVCategory {
		cates = mergeCategory(cates, c.cfg.CustomMergeCat)
	}
	return cates, nil
}

// mergeCategory merges small categories into "其他".
func mergeCategory(cates []Category, customMerge []string) []Category {
	if len(customMerge) > 0 {
		// 自定义合并模式
		mergeSet := make(map[string]bool)
		for _, name := range customMerge {
			mergeSet[name] = true
		}
		otherCategory := Category{Name: "其他", FitArea: []string{"10000"}}
		var result []Category
		for _, cate := range cates {
			if mergeSet[cate.Name] {
				otherCategory.DataList = append(otherCategory.DataList, cate.DataList...)
			} else {
				result = append(result, cate)
			}
		}
		if len(otherCategory.DataList) > 0 {
			result = append(result, otherCategory)
		}
		return result
	}

	// 自动合并模式：<=11 频道的分类合并到 "其他"
	otherCategory := Category{Name: "其他", FitArea: []string{"10000"}}
	var result []Category
	for _, cate := range cates {
		if len(cate.DataList) <= 11 {
			otherCategory.DataList = append(otherCategory.DataList, cate.DataList...)
		} else {
			result = append(result, cate)
		}
	}
	if len(otherCategory.DataList) > 0 {
		result = append(result, otherCategory)
	}
	return result
}

// uniqueData deduplicates channels across categories by channel name.
func uniqueData(cates []Category) []Category {
	seen := make(map[string]bool)
	for i := range cates {
		var unique []Channel
		for _, ch := range cates[i].DataList {
			if !seen[ch.Name] {
				seen[ch.Name] = true
				unique = append(unique, ch)
			}
		}
		cates[i].DataList = unique
	}
	return cates
}

