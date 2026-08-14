package server

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"migu-video-go/cache"
	"migu-video-go/config"
	"migu-video-go/internal/log"
	"migu-video-go/internal/singleflight"
	"migu-video-go/migu"
	"migu-video-go/updater"
)

// Server is the HTTP server for the video site proxy.
type Server struct {
	cfg     *config.Config
	migu    *migu.Client
	cache   *cache.Cache
	updater *updater.Updater
	workDir string
	sf      singleflight.Group
}

// New creates a new Server.
func New(cfg *config.Config, workDir string) *Server {
	s := &Server{
		cfg:     cfg,
		migu:    migu.NewClient(cfg),
		cache:   cache.New(),
		workDir: workDir,
	}
	s.updater = updater.New(cfg, s.migu, workDir)
	return s
}

// Start starts the HTTP server and the periodic update routine.
func (s *Server) Start() error {
	mux := http.NewServeMux()

	// 注册路由
	mux.HandleFunc("/", s.route)

	handler := loggingMiddleware(recoverMiddleware(stripTrailingSlash(mux)))

	// 启动定时更新
	go s.startUpdater()

	addr := fmt.Sprintf(":%d", s.cfg.Port)
	log.Green(fmt.Sprintf("本地地址: http://localhost:%d%s", s.cfg.Port, passSuffix(s.cfg.Pass)))
	log.Green("本程序完全免费，如果您是通过付费渠道获取，那么恭喜你成功被骗了")
	log.Green("Go语言重构分支 开源地址: https://github.com/JipZeonGit/migu-video-go 欢迎issue 感谢star")
	if s.cfg.Host != "" {
		log.Green(fmt.Sprintf("自定义地址: %s%s", s.cfg.Host, passSuffix(s.cfg.Pass)))
	}

	// 初始化数据
	s.runUpdate(0)

	return http.ListenAndServe(addr, handler)
}

// route handles all incoming requests.
func (s *Server) route(w http.ResponseWriter, r *http.Request) {
	// 身份认证
	if s.cfg.Pass != "" {
		urlSplit := strings.Split(r.URL.Path, "/")
		if len(urlSplit) < 2 || urlSplit[1] != s.cfg.Pass {
			log.Red("身份认证失败")
			w.Header().Set("Content-Type", "application/json;charset=UTF-8")
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, "身份认证失败")
			return
		}
		log.Green("身份认证成功")
		// 去掉密码前缀
		if len(urlSplit) > 3 {
			r.URL.Path = "/" + strings.Join(urlSplit[2:], "/")
		} else if len(urlSplit) == 2 {
			r.URL.Path = "/"
		} else {
			r.URL.Path = "/" + urlSplit[len(urlSplit)-1]
		}
	}

	log.Magenta("请求地址：" + r.URL.Path)

	// HEAD 请求
	if r.Method == "HEAD" {
		s.handleHealth(w, r)
		return
	}

	// 非 GET 请求
	if r.Method != "GET" {
		w.Header().Set("Content-Type", "application/json;charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"data":"请使用GET请求"}`)
		log.Red(fmt.Sprintf("使用非GET请求:%s", r.Method))
		return
	}

	// 接口列表
	interfaceList := map[string]bool{
		"/":              true,
		"/interface.txt": true,
		"/m3u":           true,
		"/txt":           true,
		"/playback.xml":  true,
		"/main.m3u":      true,
	}

	if interfaceList[r.URL.Path] {
		s.handleM3U(w, r)
		return
	}

	// 频道请求
	s.handleChannel(w, r)
}

// startUpdater runs the periodic update routine.
func (s *Server) startUpdater() {
	interval := time.Duration(s.cfg.UpdateInterval) * time.Hour
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	hours := 0
	for range ticker.C {
		hours += s.cfg.UpdateInterval
		s.runUpdate(hours)
	}
}

// runUpdate executes the data update.
func (s *Server) runUpdate(hours int) {
	log.Blue(fmt.Sprintf("准备更新文件 %s", time.Now().Format("2006-01-02 15:04:05")))
	err := s.updater.Update(hours)
	if err != nil {
		log.Red("更新失败")
		log.Debugf("error: %v", err)
	} else {
		log.Blue(fmt.Sprintf("当前已运行%d小时", hours))
	}
}

func passSuffix(pass string) string {
	if pass != "" {
		return "/" + pass
	}
	return ""
}
