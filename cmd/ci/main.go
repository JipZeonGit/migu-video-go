package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"migu-video-go/config"
	"migu-video-go/epg"
	"migu-video-go/internal/log"
	"migu-video-go/migu"
)

func main() {
	cfg := config.Load()
	log.DebugMode = cfg.Debug

	workDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get working directory: %v\n", err)
		os.Exit(1)
	}

	start := time.Now()
	client := migu.NewClient(cfg)
	epgGen := epg.NewGenerator(workDir)

	// 获取数据
	datas, err := client.DataList()
	if err != nil {
		log.Red(fmt.Sprintf("数据获取失败: %v", err))
		os.Exit(1)
	}
	log.Green("数据获取成功！")

	interfacePath := filepath.Join(workDir, "interface.txt.bak")
	playbackFile := filepath.Join(workDir, "playback.xml.bak")

	// 创建写入空内容
	writeFile(interfacePath, "")
	writeFile(playbackFile, fmt.Sprintf("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<tv generator-info-name=\"Tak\" generator-info-url=\"https://github.com/JipZeonGit/migu-video-go\">\n"))

	log.Yellow("正在更新...")

	// 写入 M3U 头
	appendFile(interfacePath, "#EXTM3U x-tvg-url=\"https://gh-proxy.com/https://raw.githubusercontent.com/JipZeonGit/migu-video-go/refs/heads/main/playback.xml,https://hk.gh-proxy.org/raw.githubusercontent.com/JipZeonGit/migu-video-go/refs/heads/main/playback.xml,https://jipzeongit.github.io/migu-video-go/playback.xml\" catchup=\"append\" catchup-source=\"&playbackbegin=${(b)yyyyMMddHHmmss}&playbackend=${(e)yyyyMMddHHmmss}\"\n")

	// 分类列表
	for _, cate := range datas {
		log.Blue(fmt.Sprintf("开始更新分类###: %s", cate.Name))

		for _, ch := range cate.DataList {
			// 更新 EPG
			program := epg.ChannelInfo{Name: ch.Name, PID: ch.PID}
			epgGen.UpdatePlaybackData(program, playbackFile, 6*time.Second, 0)

			// 获取链接
			resObj, err := client.GetAndroidURL720p(ch.PID)
			if err != nil {
				log.Red(fmt.Sprintf("%s 更新失败", ch.Name))
				continue
			}

			if resObj.URL != "" {
				// 跟随重定向
				resObj.URL = followRedirects(resObj.URL)
			}

			if resObj.URL == "" {
				log.Red(fmt.Sprintf("%s 更新失败", ch.Name))
				continue
			}

			// 写入节目
			appendFile(interfacePath, fmt.Sprintf("#EXTINF:-1 tvg-id=\"%s\" tvg-name=\"%s\" tvg-logo=\"%s\" group-title=\"%s\",%s\n%s\n",
				ch.Name, ch.Name, ch.Pics.HighResolutionH, cate.Name, ch.Name, resObj.URL))
			log.Green(fmt.Sprintf("%s 更新成功！", ch.Name))
		}
	}

	appendFileSync(playbackFile, "</tv>\n")

	// 原子重命名
	os.Rename(playbackFile, strings.TrimSuffix(playbackFile, ".bak"))
	os.Rename(interfacePath, strings.TrimSuffix(interfacePath, ".bak"))

	log.Yellow(fmt.Sprintf("本次耗时: %.1f秒", time.Since(start).Seconds()))
}

func followRedirects(rawURL string) string {
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
			continue
		}
		resp, err := client.Do(req)
		if err != nil {
			time.Sleep(150 * time.Millisecond)
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

func writeFile(path, content string) {
	os.WriteFile(path, []byte(content), 0644)
}

func appendFile(path, content string) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Red("写入文件失败: " + err.Error())
		return
	}
	defer f.Close()
	f.WriteString(content)
}

func appendFileSync(path, content string) {
	appendFile(path, content)
}
