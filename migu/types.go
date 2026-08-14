package migu

// Category represents a live TV category from the video site API.
type Category struct {
	Name     string    `json:"name"`
	VomsID   string    `json:"vomsID"`
	FitArea  []string  `json:"fitArea"`
	DataList []Channel `json:"dataList"`
}

// Channel represents a TV channel within a category.
type Channel struct {
	Name  string      `json:"name"`
	PID   string      `json:"pID"`
	Pics  ChannelPics `json:"pics"`
}

// ChannelPics contains image URLs for a channel.
type ChannelPics struct {
	HighResolutionH string `json:"highResolutionH"`
}

// PlayURLResponse is the API response from play.miguvideo.com.
type PlayURLResponse struct {
	RID  string `json:"rid"`
	Body struct {
		URLInfo struct {
			URL      string `json:"url"`
			RateType string `json:"rateType"`
		} `json:"urlInfo"`
		Content struct {
			ContID string `json:"contId"`
		} `json:"content"`
		Auth *AuthInfo `json:"auth"`
	} `json:"body"`
	Message string `json:"message"`
}

// AuthInfo contains authentication status.
type AuthInfo struct {
	Logined    bool   `json:"logined"`
	AuthResult string `json:"authResult"`
	ResultDesc string `json:"resultDesc"`
}

// CateListResponse is the API response for category listing.
type CateListResponse struct {
	Body struct {
		LiveList []Category `json:"liveList"`
	} `json:"body"`
}

// DataListResponse is the API response for channel data within a category.
type DataListResponse struct {
	Body struct {
		DataList []Channel `json:"dataList"`
	} `json:"body"`
}

// PlaybackResponse is the API response for EPG data.
type PlaybackResponse struct {
	Body struct {
		Program []struct {
			Content []PlaybackItem `json:"content"`
		} `json:"program"`
	} `json:"body"`
}

// PlaybackItem represents a single programme entry.
type PlaybackItem struct {
	ContName  string `json:"contName"`
	StartTime int64  `json:"startTime"`
	EndTime   int64  `json:"endTime"`
}

// CntvEPGResponse is the API response from CNTV EPG.
type CntvEPGResponse map[string]struct {
	Program []CntvProgram `json:"program"`
}

// CntvProgram represents a single programme from CNTV.
type CntvProgram struct {
	T  string `json:"t"`
	ST int64  `json:"st"`
	ET int64  `json:"et"`
}

// MatchListResponse is the API response for sports match list.
type MatchListResponse struct {
	Body struct {
		Days      []string               `json:"days"`
		MatchList map[string][]MatchItem `json:"matchList"`
	} `json:"body"`
}

// MatchItem represents a single sports match.
type MatchItem struct {
	MgdbID         string `json:"mgdbId"`
	PkInfoTitle    string `json:"pkInfoTitle"`
	CompetitionName string `json:"competitionName"`
	CompetitionLogo string `json:"competitionLogo"`
	ConfrontTeams  []struct {
		Name string `json:"name"`
	} `json:"confrontTeams"`
}

// MatchDetailResponse is the API response for match details.
type MatchDetailResponse struct {
	Body struct {
		EndTime      int64  `json:"endTime"`
		Keyword      string `json:"keyword"`
		MultiPlayList struct {
			ReplayList []MatchReplay `json:"replayList"`
			LiveList   []MatchReplay `json:"liveList"`
			PreList    []struct {
				StartTimeStr string `json:"startTimeStr"`
			} `json:"preList"`
		} `json:"multiPlayList"`
	} `json:"body"`
}

// MatchReplay represents a replay/live entry for a match.
type MatchReplay struct {
	Name  string `json:"name"`
	PID   string `json:"pID"`
	StartTimeStr string `json:"startTimeStr"`
}

// ReplayListResponse is the API response for replay listing.
type ReplayListResponse struct {
	Body struct {
		ReplayList []MatchReplay `json:"replayList"`
	} `json:"body"`
}

// TokenRefreshResponse is the API response for token refresh.
type TokenRefreshResponse struct {
	ResultCode string `json:"resultCode"`
}
