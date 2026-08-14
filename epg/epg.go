package epg

import (
	"encoding/xml"
	"fmt"
	"os"
	"strings"
	"time"

	"migu-video-go/data"
	"migu-video-go/internal/log"
	"migu-video-go/internal/netutil"
	"migu-video-go/migu"
)

// Generator generates EPG/XMLTV data.
type Generator struct {
	workDir string
}

// NewGenerator creates a new EPG generator.
func NewGenerator(workDir string) *Generator {
	return &Generator{workDir: workDir}
}

// ChannelInfo represents a channel for EPG generation.
type ChannelInfo struct {
	Name string
	PID  string
}

// UpdatePlaybackData updates EPG data for a channel.
func (g *Generator) UpdatePlaybackData(program ChannelInfo, filePath string, timeout time.Duration, githubAnd8 time.Duration) bool {
	if cntvName, ok := data.CntvNames[program.Name]; ok {
		return g.updatePlaybackDataByCntv(program, cntvName, filePath, timeout, githubAnd8)
	}
	return g.updatePlaybackDataBySite(program, filePath, timeout, githubAnd8)
}

// updatePlaybackDataBySite fetches EPG data from video site API.
func (g *Generator) updatePlaybackDataBySite(program ChannelInfo, filePath string, timeout time.Duration, githubAnd8 time.Duration) bool {
	date := time.Now().Add(githubAnd8)
	today := date.Format("20060102")

	url := fmt.Sprintf("https://program-sc.miguvideo.com/live/v2/tv-programs-data/%s/%s", program.PID, today)
	var resp migu.PlaybackResponse
	err := netutil.FetchJSONWithTimeout(url, nil, timeout, &resp)
	if err != nil || len(resp.Body.Program) == 0 || len(resp.Body.Program[0].Content) == 0 {
		return false
	}

	playbackData := resp.Body.Program[0].Content

	// 写入频道信息
	channelXML := fmt.Sprintf("    <channel id=\"%s\">\n        <display-name lang=\"zh\">%s</display-name>\n    </channel>\n",
		xmlEscapeAttr(program.Name), xmlEscape(program.Name))
	appendToFile(filePath, channelXML)

	// 写入节目信息
	for _, item := range playbackData {
		contName := xmlEscape(item.ContName)
		startTime := time.UnixMilli(item.StartTime + int64(githubAnd8/time.Millisecond))
		endTime := time.UnixMilli(item.EndTime + int64(githubAnd8/time.Millisecond))
		programmeXML := fmt.Sprintf("    <programme channel=\"%s\" start=\"%s +0800\" stop=\"%s +0800\">\n        <title lang=\"zh\">%s</title>\n    </programme>\n",
			xmlEscapeAttr(program.Name),
			startTime.Format("20060102150405"),
			endTime.Format("20060102150405"),
			contName)
		appendToFile(filePath, programmeXML)
	}
	return true
}

// updatePlaybackDataByCntv fetches EPG data from CNTV API.
func (g *Generator) updatePlaybackDataByCntv(program ChannelInfo, cntvName string, filePath string, timeout time.Duration, githubAnd8 time.Duration) bool {
	date := time.Now().Add(githubAnd8)
	today := date.Format("20060102")

	url := fmt.Sprintf("https://api.cntv.cn/epg/epginfo3?serviceId=shiyi&d=%s&c=%s", today, cntvName)
	var resp migu.CntvEPGResponse
	err := netutil.FetchJSONWithTimeout(url, nil, timeout, &resp)
	if err != nil {
		return false
	}

	cntvData, ok := resp[cntvName]
	if !ok || len(cntvData.Program) == 0 {
		return false
	}

	// 写入频道信息
	channelXML := fmt.Sprintf("    <channel id=\"%s\">\n        <display-name lang=\"zh\">%s</display-name>\n    </channel>\n",
		xmlEscapeAttr(program.Name), xmlEscape(program.Name))
	appendToFile(filePath, channelXML)

	// 写入节目信息
	for _, item := range cntvData.Program {
		contName := xmlEscape(item.T)
		startTime := time.Unix(item.ST+int64(githubAnd8/time.Second), 0)
		endTime := time.Unix(item.ET+int64(githubAnd8/time.Second), 0)
		programmeXML := fmt.Sprintf("    <programme channel=\"%s\" start=\"%s +0800\" stop=\"%s +0800\">\n        <title lang=\"zh\">%s</title>\n    </programme>\n",
			xmlEscapeAttr(program.Name),
			startTime.Format("20060102150405"),
			endTime.Format("20060102150405"),
			contName)
		appendToFile(filePath, programmeXML)
	}
	return true
}

// xmlEscape escapes special XML characters for element content.
func xmlEscape(s string) string {
	var b strings.Builder
	if err := xml.EscapeText(&b, []byte(s)); err != nil {
		return s
	}
	return b.String()
}

// xmlEscapeAttr escapes special XML characters for attribute values.
func xmlEscapeAttr(s string) string {
	s = xmlEscape(s)
	s = strings.ReplaceAll(s, `"`, "&quot;")
	s = strings.ReplaceAll(s, `'`, "&apos;")
	return s
}

func appendToFile(filePath, content string) {
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Red("写入文件失败: " + err.Error())
		return
	}
	defer f.Close()
	f.WriteString(content)
}
