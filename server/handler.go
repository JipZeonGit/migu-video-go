package server

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"migu-video-go/internal/log"
)

// handleHealth handles HEAD / requests (health check).
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	w.WriteHeader(http.StatusOK)
}

// handleM3U handles GET / /interface.txt /m3u /txt /playback.xml /main.m3u.
func (s *Server) handleM3U(w http.ResponseWriter, r *http.Request) {
	urlPath := r.URL.Path

	var fileName string
	contentType := "text/plain;charset=UTF-8"

	switch urlPath {
	case "/playback.xml":
		fileName = filepath.Join(s.workDir, "playback.xml")
		contentType = "text/xml;charset=UTF-8"
	case "/txt":
		fileName = filepath.Join(s.workDir, "interfaceTXT.txt")
	case "/m3u":
		fileName = filepath.Join(s.workDir, "interface.txt")
		contentType = "audio/x-mpegurl; charset=utf-8"
	case "/main.m3u":
		fileName = filepath.Join(s.workDir, "interface.txt")
		contentType = "application/octet-stream; charset=utf-8"
	default:
		fileName = filepath.Join(s.workDir, "interface.txt")
	}

	content, err := os.ReadFile(fileName)
	if err != nil {
		log.Red("文件获取失败")
		log.Debugf("error: %v", err)
		w.Header().Set("Content-Type", "text/plain;charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "获取失败")
		return
	}

	if urlPath == "/playback.xml" {
		w.Header().Set("Content-Type", contentType)
		w.WriteHeader(http.StatusOK)
		w.Write(content)
		return
	}

	// 替换占位符
	replaceHost := fmt.Sprintf("http://%s", r.Host)
	if s.cfg.Host != "" && (r.Header.Get("X-Real-Ip") != "" || r.Header.Get("X-Forwarded-For") != "" || strings.Contains(s.cfg.Host, r.Host)) {
		replaceHost = s.cfg.Host
	}
	if s.cfg.Pass != "" {
		replaceHost = replaceHost + "/" + s.cfg.Pass
	}

	urlUserID, urlToken := extractUserFromURL(r.URL.Path)
	if urlUserID != s.cfg.UserID && urlToken != s.cfg.Token {
		if urlUserID != "" {
			replaceHost = fmt.Sprintf("%s/%s/%s", replaceHost, urlUserID, urlToken)
		}
	}

	result := strings.ReplaceAll(string(content), "${replace}", replaceHost)

	w.Header().Set("Content-Type", contentType)
	if urlPath == "/m3u" {
		w.Header().Set("Content-Disposition", `inline; filename="interface.m3u"`)
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, result)
}

// handleChannel handles GET /{channelID} requests and returns 302 redirects.
func (s *Server) handleChannel(w http.ResponseWriter, r *http.Request) {
	urlPath := r.URL.Path

	result := s.channel(urlPath, r)

	if result.Code != http.StatusFound {
		log.Red(result.Desc)
		w.Header().Set("Content-Type", "application/json;charset=UTF-8")
		w.WriteHeader(result.Code)
		fmt.Fprint(w, result.Desc)
		return
	}

	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	w.Header().Set("Location", result.PlayURL)
	w.WriteHeader(http.StatusFound)
}

// channelResult holds the result of a channel request.
type channelResult struct {
	Code    int
	PID     string
	Desc    string
	PlayURL string
}

// channel processes a channel playback request.
func (s *Server) channel(urlPath string, r *http.Request) channelResult {
	result := channelResult{
		Code: http.StatusOK,
		Desc: "服务异常",
	}

	// 处理频道 ID（从 URL 末尾取，兼容 /uid/token/pid 格式）
	urlSplit := strings.Split(urlPath, "/")
	if len(urlSplit) < 2 {
		result.Desc = "地址格式错误"
		return result
	}
	pid := urlSplit[len(urlSplit)-1]
	params := ""

	// 处理回放参数
	if idx := strings.Index(pid, "?"); idx != -1 {
		log.Green("处理传入参数")
		params = pid[idx+1:]
		pid = pid[:idx]
	} else {
		log.Grey("无参数传入")
	}

	// 验证 PID 是否为数字
	if !isNumeric(pid) {
		result.Desc = "地址格式错误"
		return result
	}

	log.Yellow("频道ID " + pid)

	// 检查缓存
	if cached, ok := s.cache.Get(pid); ok {
		playURL := cached.Value
		if playURL == "" {
			result.Desc = fmt.Sprintf("%s 节目调整，暂不提供服务", pid)
			return result
		}
		if params != "" {
			playURL = appendQueryParams(playURL, params)
		}
		log.Green("使用缓存数据")
		result.Code = http.StatusFound
		result.PlayURL = playURL
		return result
	}

	// 获取 URL（通过 singleflight 合并同频道并发请求）
	urlUserID, urlToken := extractUserFromURL(r.URL.Path)
	if urlUserID == "" {
		urlUserID = s.cfg.UserID
		urlToken = s.cfg.Token
	}

	// singleflight key: pid + userID，同频道同用户的并发请求只发一次 API 调用
	sfKey := pid + "|" + urlUserID
	val, err := s.sf.Do(sfKey, func() (string, error) {
		// 限速由 client 内置的 limiter 处理
		if s.cfg.RateType >= 3 && (urlUserID == "" || urlToken == "") {
			res, err := s.migu.GetAndroidURL720p(pid)
			if err != nil {
				return "", err
			}
			// 返回 "url\nrateType" 格式
			return fmt.Sprintf("%s\n%d", res.URL, res.RateType), nil
		}
		res, err := s.migu.GetAndroidURL(urlUserID, urlToken, pid, s.cfg.RateType)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%s\n%d", res.URL, res.RateType), nil
	})
	if err != nil {
		log.Debugf("error: %v", err)
		result.Desc = "链接请求出错"
		return result
	}

	playURL := ""
	parts := strings.SplitN(val, "\n", 2)
	if len(parts) == 2 {
		playURL = parts[0]
	}

	log.Debugf("添加加密字段后链接 %s", playURL)

	// 缓存
	log.Green("添加节目缓存 " + pid)
	addTime := 3 * time.Hour
	if playURL == "" {
		addTime = 1 * time.Minute
	}
	s.cache.Set(pid, playURL, nil, addTime)

	if playURL == "" {
		result.Desc = fmt.Sprintf("%s 节目调整，暂不提供服务", pid)
		return result
	}

	// 添加回放参数
	if params != "" {
		playURL = appendQueryParams(playURL, params)
	}

	log.Green("链接获取成功")
	result.Code = http.StatusFound
	result.PlayURL = playURL
	return result
}

// extractUserFromURL extracts userId and token from URL path if present.
func extractUserFromURL(urlPath string) (userID, token string) {
	parts := strings.Split(urlPath, "/")
	if len(parts) >= 4 {
		// /{userId}/{token}/...
		return parts[1], parts[2]
	}
	return "", ""
}

// appendQueryParams appends query parameters to a URL.
func appendQueryParams(u, params string) string {
	if strings.Contains(u, "?") {
		return u + "&" + params
	}
	return u + "?" + params
}

// isNumeric checks if a string contains only digits.
func isNumeric(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func parseInt(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}
