package updater

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"migu-video-go/config"
	"migu-video-go/epg"
	"migu-video-go/internal/log"
	"migu-video-go/internal/netutil"
	"migu-video-go/migu"
)

// 预编译正则（避免循环内重复编译）
var (
	reHighlights    = regexp.MustCompile(`.*集锦|训练.*`)
	reReplay        = regexp.MustCompile(`.*回放|赛.*`)
	reLiveHighlight = regexp.MustCompile(`.*集锦.*`)
)

// Updater periodically updates TV and sports data.
type Updater struct {
	cfg   *config.Config
	migu  *migu.Client
	epg   *epg.Generator
	workDir string
	ignoreCategorySet map[string]bool
}

// New creates a new Updater.
func New(cfg *config.Config, m *migu.Client, workDir string) *Updater {
	u := &Updater{
		cfg:   cfg,
		migu:  m,
		epg:   epg.NewGenerator(workDir),
		workDir: workDir,
		ignoreCategorySet: make(map[string]bool),
	}
	// 初始化忽略分类
	for _, cat := range cfg.IgnoreCategory {
		u.ignoreCategorySet[cat] = true
	}
	return u
}

// Update runs the full update cycle.
func (u *Updater) Update(hours int) error {
	// 每 720 小时刷新 token
	if hours%720 == 0 && hours > 0 {
		if u.cfg.UserID != "" && u.cfg.Token != "" {
			if u.cfg.RefreshToken {
				if migu.RefreshToken(u.cfg.UserID, u.cfg.Token, u.migu.Limiter()) {
					log.Green("token刷新成功")
				} else {
					log.Red("token刷新失败")
				}
			} else {
				log.Green("跳过token刷新")
			}
		}
	}

	// TV 更新
	if u.ignoreCategorySet["TV"] {
		log.Yellow("TV更新已屏蔽")
	} else {
		if err := u.updateTV(hours); err != nil {
			return fmt.Errorf("updateTV: %w", err)
		}
	}

	// PE 更新
	if u.ignoreCategorySet["PE"] {
		log.Yellow("PE更新已屏蔽")
	} else {
		if err := u.updatePE(hours); err != nil {
			return fmt.Errorf("updatePE: %w", err)
		}
	}

	return nil
}

// updateTV updates TV channel data.
func (u *Updater) updateTV(_hours int) error {
	start := time.Now()
	log.Yellow("开始更新TV...")

	datas, err := u.migu.DataList()
	if err != nil {
		return fmt.Errorf("fetch data list: %w", err)
	}
	log.Green("TV数据获取成功！")

	interfacePath := filepath.Join(u.workDir, "interface.txt.bak")
	interfaceTXTPath := filepath.Join(u.workDir, "interfaceTXT.txt.bak")
	playbackFile := filepath.Join(u.workDir, "playback.xml.bak")

	// 写入 M3U 头
	writeFile(interfacePath, fmt.Sprintf("#EXTM3U x-tvg-url=\"${replace}/playback.xml\" catchup=\"append\" catchup-source=\"?playbackbegin=${(b)yyyyMMddHHmmss}&playbackend=${(e)yyyyMMddHHmmss}\"\n"))
	writeFile(interfaceTXTPath, "")

	// 写入 XMLTV 头
	writeFile(playbackFile, fmt.Sprintf("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<tv generator-info-name=\"Tak\" generator-info-url=\"%s\">\n", u.cfg.Host))

	for _, cate := range datas {
		if u.ignoreCategorySet[cate.Name] {
			log.Yellow(fmt.Sprintf("TV分类###:%s 已屏蔽！", cate.Name))
			continue
		}

		// TXT 分类头
		appendFile(interfaceTXTPath, fmt.Sprintf("%s,#genre#\n", cate.Name))

		for _, ch := range cate.DataList {
			u.migu.Acquire()
			program := epg.ChannelInfo{Name: ch.Name, PID: ch.PID}
			u.epg.UpdatePlaybackData(program, playbackFile, 6*time.Second, 0)

			// 写入 M3U 条目
			appendFile(interfacePath, fmt.Sprintf("#EXTINF:-1 tvg-id=\"%s\" tvg-name=\"%s\" tvg-logo=\"%s\" group-title=\"%s\",%s\n${replace}/%s\n",
				ch.Name, ch.Name, ch.Pics.HighResolutionH, cate.Name, ch.Name, ch.PID))
			// 写入 TXT 条目
			appendFile(interfaceTXTPath, fmt.Sprintf("%s,${replace}/%s\n", ch.Name, ch.PID))
		}
		log.Green(fmt.Sprintf("分类###:%s 更新完成！", cate.Name))
	}

	appendFileSync(playbackFile, "</tv>\n")

	// 原子重命名
	if err := os.Rename(playbackFile, strings.TrimSuffix(playbackFile, ".bak")); err != nil {
		log.Red("重命名 playback.xml 失败: " + err.Error())
	}
	if err := os.Rename(interfacePath, strings.TrimSuffix(interfacePath, ".bak")); err != nil {
		log.Red("重命名 interface.txt 失败: " + err.Error())
	}
	if err := os.Rename(interfaceTXTPath, strings.TrimSuffix(interfaceTXTPath, ".bak")); err != nil {
		log.Red("重命名 interfaceTXT.txt 失败: " + err.Error())
	}

	log.Green("TV更新完成！")
	log.Yellow(fmt.Sprintf("TV更新耗时: %.1f秒", time.Since(start).Seconds()))
	return nil
}

// updatePE updates sports match data.
func (u *Updater) updatePE(_hours int) error {
	start := time.Now()
	log.Yellow("开始更新PE...")

	var resp migu.MatchListResponse
	u.migu.Acquire()
	client := &http.Client{Timeout: 6 * time.Second}
	err := netutil.FetchJSONWithClient("http://v0-sc.miguvideo.com/vms-match/v6/staticcache/basic/match-list/normal-match-list/0/all/default/1/miguvideo", nil, client, &resp)
	if err != nil {
		return fmt.Errorf("fetch match list: %w", err)
	}
	log.Green("PE数据获取成功！")

	interfacePath := filepath.Join(u.workDir, "interface.txt.bak")
	interfaceTXTPath := filepath.Join(u.workDir, "interfaceTXT.txt.bak")

	// 屏蔽所有 TV 分类
	if u.ignoreCategorySet["TV"] {
		writeFile(interfacePath, "#EXTM3U x-tvg-url=\"${replace}/playback.xml\" catchup=\"append\" catchup-source=\"?playbackbegin=${(b)yyyyMMddHHmmss}&playbackend=${(e)yyyyMMddHHmmss}\"\n")
		writeFile(interfaceTXTPath, "")
	} else {
		copyFile(filepath.Join(u.workDir, "interface.txt"), interfacePath)
		copyFile(filepath.Join(u.workDir, "interfaceTXT.txt"), interfaceTXTPath)
	}

	today := time.Now().Format("20060102")

	for i := 1; i < 4 && i < len(resp.Body.Days); i++ {
		date := resp.Body.Days[i]
		relativeDate := fmt.Sprintf("%s%s-%s", getRelativeDay(date, today), date[4:6], date[6:8])

		// 忽略分类
		ignoreKey := fmt.Sprintf("体育-%s", relativeDate[:len([]rune(relativeDate))-6])
		if u.ignoreCategorySet[ignoreKey] {
			log.Yellow(fmt.Sprintf("PE分类###: 体育-%s已屏蔽！", relativeDate))
			continue
		}

		appendFile(interfaceTXTPath, fmt.Sprintf("体育-%s,#genre#\n", relativeDate))

		matchList := resp.Body.MatchList[date]
		for _, match := range matchList {
			pkInfoTitle := match.PkInfoTitle
			if len(match.ConfrontTeams) >= 2 {
				pkInfoTitle = fmt.Sprintf("%sVS%s", match.ConfrontTeams[0].Name, match.ConfrontTeams[1].Name)
			}

			var matchDetail migu.MatchDetailResponse
			u.migu.Acquire()
			detailURL := fmt.Sprintf("https://vms-sc.miguvideo.com/vms-match/v6/staticcache/basic/basic-data/%s/miguvideo", match.MgdbID)
			err := netutil.FetchJSON(detailURL, nil, &matchDetail)
			if err != nil {
				log.Yellow(fmt.Sprintf("%s %s 更新失败 此警告不影响正常使用 可忽略", match.MgdbID, pkInfoTitle))
				continue
			}

			// 比赛已结束
			if matchDetail.Body.EndTime < time.Now().UnixMilli() {
				replayURL := fmt.Sprintf("http://app-sc.miguvideo.com/vms-match/v5/staticcache/basic/all-view-list/%s/2/miguvideo", match.MgdbID)
				var replayResp migu.ReplayListResponse
				replayList := matchDetail.Body.MultiPlayList.ReplayList
				u.migu.Acquire()
				err := netutil.FetchJSON(replayURL, nil, &replayResp)
				if err == nil && len(replayResp.Body.ReplayList) > 0 {
					replayList = replayResp.Body.ReplayList
				}
				if len(replayList) == 0 {
					log.Yellow(fmt.Sprintf("%s %s 无回放", match.MgdbID, pkInfoTitle))
					continue
				}
				for _, replay := range replayList {
					if reHighlights.MatchString(replay.Name) {
						continue
					}
					if reReplay.MatchString(replay.Name) {
						timeStr := ""
						if matchDetail.Body.Keyword != "" && len(matchDetail.Body.Keyword) > 7 {
							timeStr = matchDetail.Body.Keyword[7:]
						}
						if len(matchDetail.Body.MultiPlayList.PreList) > 0 {
							lastPre := matchDetail.Body.MultiPlayList.PreList[len(matchDetail.Body.MultiPlayList.PreList)-1]
							if lastPre.StartTimeStr != "" && len(lastPre.StartTimeStr) > 16 {
								timeStr = lastPre.StartTimeStr[11:16]
							}
						}
						competitionDesc := fmt.Sprintf("%s %s %s %s", match.CompetitionName, pkInfoTitle, replay.Name, timeStr)
						appendFileSync(interfacePath, fmt.Sprintf("#EXTINF:-1 tvg-id=\"%s\" tvg-name=\"%s\" tvg-logo=\"%s\" group-title=\"体育-%s\",%s\n${replace}/%s\n",
							pkInfoTitle, competitionDesc, match.CompetitionLogo, relativeDate, competitionDesc, replay.PID))
						appendFileSync(interfaceTXTPath, fmt.Sprintf("%s,${replace}/%s\n", competitionDesc, replay.PID))
					}
				}
				continue
			}

			// 比赛未结束
			for _, live := range matchDetail.Body.MultiPlayList.LiveList {
				if reLiveHighlight.MatchString(live.Name) || live.StartTimeStr == "" {
					continue
				}
				startTimeStr := ""
				if len(live.StartTimeStr) > 16 {
					startTimeStr = live.StartTimeStr[11:16]
				}
				competitionDesc := fmt.Sprintf("%s %s %s %s", match.CompetitionName, pkInfoTitle, live.Name, startTimeStr)
				appendFileSync(interfacePath, fmt.Sprintf("#EXTINF:-1 tvg-id=\"%s\" tvg-name=\"%s\" tvg-logo=\"%s\" group-title=\"体育-%s\",%s\n${replace}/%s\n",
					pkInfoTitle, competitionDesc, match.CompetitionLogo, relativeDate, competitionDesc, live.PID))
				appendFileSync(interfaceTXTPath, fmt.Sprintf("%s,${replace}/%s\n", competitionDesc, live.PID))
			}
		}
		log.Green(fmt.Sprintf("日期 %s 更新完成！", date))
	}

	// 原子重命名
	if err := os.Rename(interfacePath, strings.TrimSuffix(interfacePath, ".bak")); err != nil {
		log.Red("重命名 interface.txt 失败: " + err.Error())
	}
	if err := os.Rename(interfaceTXTPath, strings.TrimSuffix(interfaceTXTPath, ".bak")); err != nil {
		log.Red("重命名 interfaceTXT.txt 失败: " + err.Error())
	}

	log.Green("PE更新完成！")
	log.Yellow(fmt.Sprintf("PE更新耗时: %.1f秒", time.Since(start).Seconds()))
	return nil
}

func getRelativeDay(date, today string) string {
	if date == today {
		return "今天"
	} else if date > today {
		return "明天"
	}
	return "昨天"
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

func copyFile(src, dst string) {
	data, err := os.ReadFile(src)
	if err != nil {
		return
	}
	os.WriteFile(dst, data, 0644)
}
