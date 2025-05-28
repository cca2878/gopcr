package models

import (
	"math/rand"
	"net/url"
	"time"
)

// SourceIniIndex
const sourceIniIndexReqPath = "source_ini/index?format=json"

type SourceIniIndexReq struct {
	BaseRequest
}

type SourceIniIndexResp struct {
	Server []string `json:"server"`
}

func NewSourceIniIndexReq() SourceIniIndexReq {
	req := SourceIniIndexReq{
		BaseRequest: NewBaseRequest(),
	}
	// 特殊不加密API
	req.isEncrypt = false
	return req
}
func (s *SourceIniIndexReq) GetUrl() (*url.URL, error) {
	return parseModelUrl(sourceIniIndexReqPath)
}

// source_ini/get_maintenance_status
const sourceIniGetMaintenanceStatusReqPath = "source_ini/get_maintenance_status?format=json"

type SourceIniGetMaintenanceStatusReq struct {
	BaseRequest
}

type SourceIniGetMaintenanceStatusResp struct {
	ManifestVer         string `json:"manifest_ver"`
	RequiredManifestVer string `json:"required_manifest_ver"`
}

func NewSourceIniGetMaintenanceStatusReq() SourceIniGetMaintenanceStatusReq {
	req := SourceIniGetMaintenanceStatusReq{
		BaseRequest: NewBaseRequest(),
	}
	// 特殊不加密API
	req.isEncrypt = false
	return req
}
func (s SourceIniGetMaintenanceStatusReq) GetUrl() (*url.URL, error) {
	return parseModelUrl(sourceIniGetMaintenanceStatusReqPath)
}

// SdkLogin
const sdkLoginReqPath = "tool/sdk_login"

type SdkLoginReq struct {
	BaseRequest

	Uid         string `json:"uid"`
	AccessKey   string `json:"access_key"`
	Platform    string `json:"platform"`
	ChannelId   string `json:"channel_id"`
	Challenge   string `json:"challenge"`
	Validate    string `json:"validate"`
	Seccode     string `json:"seccode"`
	CaptchaType string `json:"captcha_type"`
	ImageToken  string `json:"image_token"`
	CaptchaCode string `json:"captcha_code"`
}

func (s SdkLoginReq) GetUrl() (*url.URL, error) {
	return parseModelUrl(sdkLoginReqPath)
}

func NewSdkLoginReq() SdkLoginReq {
	return SdkLoginReq{
		BaseRequest: NewBaseRequest(),
	}
}

type SdkLoginResp struct {
	IsRisk *bool `json:"is_risk"`
}

// GameStart
const gameStartReqPath = "check/game_start"

type GameStartReq struct {
	BaseRequest

	AppType      int    `json:"app_type"`
	CampaignData string `json:"campaign_data"`
	CampaignUser int    `json:"campaign_user"`
}

func (g GameStartReq) GetUrl() (*url.URL, error) {
	return parseModelUrl(gameStartReqPath)
}

func NewGameStartReq() GameStartReq {
	// 使用当前时间纳秒作为种子创建新的随机源
	source := rand.NewSource(time.Now().UnixNano())

	return GameStartReq{
		BaseRequest:  NewBaseRequest(),
		AppType:      0,
		CampaignData: "",
		CampaignUser: rand.New(source).Intn(50001) * 2,
	}
}

type GameStartResp struct {
	NowTutorial  bool   `json:"now_tutorial"`
	NowName      string `json:"now_name"`
	NowTeamLevel int    `json:"now_team_level"`
}

// LoadIndex
const loadIndexReqPath = "load/index"

type LoadIndexReq struct {
	BaseRequest
	Carrier string `json:"carrier"`
}

func NewLoadIndexReq() LoadIndexReq {
	return LoadIndexReq{
		BaseRequest: NewBaseRequest(),
		Carrier:     "LN_NMSL",
	}
}

func (l LoadIndexReq) GetUrl() (*url.URL, error) {
	return parseModelUrl(loadIndexReqPath)
}

// LoadIndexResp 需要整体获取的时候就直接从resp解析
type LoadIndexResp struct {
	DailyResetTime uint `json:"daily_reset_time"`
}

// HomeIndex

const HomeIndexReqPath = "home/index"

type HomeIndexReq struct {
	BaseRequest

	MessageId   int   `json:"message_id"`
	TipsIdList  []int `json:"tips_id_list"`
	IsFirst     int   `json:"is_first"`
	GoldHistory int   `json:"gold_history"`
}

func NewHomeIndexReq() HomeIndexReq {
	return HomeIndexReq{
		BaseRequest: NewBaseRequest(),
	}
}

func (l HomeIndexReq) GetUrl() (*url.URL, error) {
	return parseModelUrl(HomeIndexReqPath)
}

type HomeIndexResp struct {
	DailyResetTime uint `json:"daily_reset_time"`
}
