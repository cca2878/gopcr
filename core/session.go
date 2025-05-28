package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-resty/resty/v2"
	"gopcr/config"
	"gopcr/log"
	"gopcr/models"
	"net/http"
	"strconv"
	"strings"
)

type SdkAccount struct {
	Uid       string
	AccessKey string

	Platform string
	Channel  string
}

// session 实现Princess Connect Re:Dive的API客户端
type session struct {
	ctx        context.Context    // 内部创建的 context
	ctxCancel  context.CancelFunc // 用于取消 context
	httpClient *resty.Client
	crypto     *pcrCrypto
	sdkAccount SdkAccount
	logged     bool
	viewerId   uint64
	expireTime uint
}

// SessionOption 定义客户端选项
type SessionOption func(*session)

// WithChannelServer 渠道服Option
func WithChannelServer() SessionOption {
	return func(client *session) {
		client.httpClient.
			SetBaseURL("https://"+config.DefaultChannelApiHost).
			SetHeader("PLATFORM-ID", "4").
			SetHeaders(config.GetBiliHeaders())
	}
}

// newSession 创建一个新的Client。
// 默认为B服，渠道服用上面的Option。
func newSession(sdkAccount SdkAccount, options ...SessionOption) (*session, error) {

	// 创建一个带有取消功能的 context
	ctx, cancel := context.WithCancel(context.Background())
	// 创建并配置 HTTP 客户端
	httpClient := resty.New().
		SetBaseURL("https://" + config.DefaultBiliApiHost).
		// Debug
		//SetProxy("http://127.0.0.1:8516").
		// 设置默认超时等
		SetTimeout(config.DefaultRequestTimeout).
		// 设置Headers
		SetHeaders(config.GetBiliHeaders()).
		SetHeaders(map[string]string{
			"PLATFORM":    sdkAccount.Platform,
			"PLATFORM-ID": sdkAccount.Platform,
			"CHANNEL-ID":  sdkAccount.Channel,
		})

	// 使用默认配置创建客户端
	client := &session{
		httpClient: httpClient,
		ctx:        ctx,
		ctxCancel:  cancel,
		viewerId:   0,
		sdkAccount: sdkAccount,
		logged:     false,
		crypto:     newPCRCrypto(),
	}

	// 应用选项
	for _, option := range options {
		option(client)
	}

	//
	//err := httpClient.getConfig()
	//if err != nil {
	//	return nil, fmt.Errorf("配置失败: %w", err)
	//}
	err := client.getConfig()
	if err != nil {
		return nil, err
	}
	return client, nil
}

func (s *session) preReq(request models.IRequest) (*resty.Request, error) {
	req := s.httpClient.R().
		SetContext(s.ctx)

	if request.IsEncrypt() {
		encryptViewerId, err := s.crypto.EncryptViewerId(s.viewerId)
		if err != nil {
			return nil, fmt.Errorf("加密ViewerId失败: %w", err)
		}
		request.SetViewerId(encryptViewerId)

		data, err := s.crypto.EncryptData(request)
		if err != nil {
			return nil, fmt.Errorf("加密数据失败: %w", err)
		}
		req.SetBody(data)
	} else {
		request.SetViewerId(strconv.FormatUint(uint64(s.viewerId), 10))

		jsonData, err := json.Marshal(request)
		if err != nil {
			return nil, fmt.Errorf("序列化数据失败: %w", err)
		}

		req.SetHeader("Content-Type", "application/octet-stream").
			SetBody(jsonData).ForceContentType("application/json")
	}

	return req, nil
}

// execReq 执行HTTP请求。result必须传指针
func (s *session) execReq(
	request models.IRequest,
	result models.IResponse,
) (*resty.Response, error) {
	var err error

	// 请求预处理
	req, err := s.preReq(request)
	if err != nil {
		log.Error("准备请求失败: %v", err)
		return nil, &models.ApiError{
			Operation: "execReq:preReq",
			Message:   "准备请求失败",
			Err:       err,
		}
	}
	//if s.requestId != "" {
	//	req.SetHeader("REQUEST-ID", s.requestId)
	//}
	url, err := request.GetUrl()
	if err != nil {
		log.Error("获取URL失败: %v", err)
		return nil, &models.ApiError{
			Operation: "execReq:GetUrl",
			Message:   "获取URL失败",
			Err:       err,
		}
	}

	log.Debug("发送请求: %s", url.String())

	var resp *resty.Response

	// 根据请求是否需要加密处理来选择不同的处理方式
	if request.IsEncrypt() {
		// 针对加密请求，手动处理响应解密
		resp, err = req.Execute(request.GetMethod(), url.String())
	} else {
		// 非加密请求使用resty自动解析
		resp, err = req.SetResult(result).Execute(request.GetMethod(), url.String())
	}

	if err != nil {
		log.Error("请求发送失败: %v", err)
		return resp, &models.ApiError{
			Operation: "execReq:Execute",
			Message:   "请求发送失败",
			Err:       err,
		}

	}

	log.Debug("收到响应: 状态码=%d, 内容长度=%d", resp.StatusCode(), len(resp.Body()))

	if resp.StatusCode() != http.StatusOK {
		log.Error("HTTP失败，状态码: %d", resp.StatusCode())
		return resp, &models.ApiError{
			Operation:  "execReq:HttpStatus",
			Message:    "HTTP失败",
			Err:        err,
			HTTPStatus: resp.StatusCode(),
		}
	}

	// 对加密响应进行解密处理
	if request.IsEncrypt() {
		if err = s.crypto.DecryptData(string(resp.Body()), result); err != nil {
			log.Error("响应解密失败: %v", err)
			return resp, &models.ApiError{
				Operation: "execReq:DecryptData",
				Message:   "响应解密失败",
				Err:       err,
			}
		}
	}

	// 版本号需要更新
	if result.GetResultCode() == 204 && strings.Contains(req.URL, "check/game_start") {
		var newAppVer string
		newAppVer, err = getNewAppVer()
		if err != nil {
			return resp, &models.ApiError{
				Operation: "execReq:getNewAppVer",
				Message:   "获取新版本号失败",
				Err:       err,
			}
		}
		err = config.GetInstance().SetOptVal(config.AppVer, newAppVer)
		if err != nil {
			return resp, &models.ApiError{
				Operation: "execReq:SetOptVal",
				Message:   "更新AppVer失败",
				Err:       err,
			}
		}
		s.httpClient.SetHeader("APP-VER", newAppVer)
		log.Debug("已更新AppVer: %s", newAppVer)
		return resp, &models.ApiError{
			Operation: "execReq:UpdateAppVer",
			Message:   "已更新AppVer",
			ApiCode:   result.GetResultCode(),
		}
	}

	if result.GetResultCode() != 1 {
		log.Error("API失败，结果码: %d", result.GetResultCode())
		return resp, &models.ApiError{
			Operation: "execReq:ResultCode",
			Message:   "API失败",
			ApiCode:   result.GetResultCode(),
		}
	}

	// 请求后处理
	if reqId := result.GetRequestId(); reqId != "" {
		s.httpClient.SetHeader("REQUEST-ID", reqId)
	}
	if sid := result.GetSID(); sid != "" {
		s.httpClient.SetHeader("SID", calcSID(sid))
	}

	return resp, nil
}

func (s *session) getConfig() error {
	var err error

	indexReq := models.NewSourceIniIndexReq()
	var indexResult models.BaseResponse[models.SourceIniIndexResp]

	if _, err = s.execReq(&indexReq, &indexResult); err != nil {
		return err
	}
	maintenanceReq := models.NewSourceIniGetMaintenanceStatusReq()
	var maintenanceResult models.BaseResponse[models.SourceIniGetMaintenanceStatusResp]

	if _, err = s.execReq(&maintenanceReq, &maintenanceResult); err != nil {
		return err
	}
	if maintenanceResult.Data.ManifestVer != "" {
		s.httpClient.SetHeader("MANIFEST-VER", maintenanceResult.Data.ManifestVer)
	}

	return nil
}

func (s *session) login() error {
	// 登录流程：
	// sdk_login: 用bsdk拿到的uid和access_key来登录，暂未测试uid是否必需。完成后拿到滚动req id，放headers
	// game_start: 用uid和viewer_id来启动游戏
	// load_index: 仿照真实流程
	// home_index: 仿照真实流程
	if s.logged {
		return errors.New("已经登录")
	}
	// 登录逻辑 尝试3次
	for i := range 3 {
		//err := s.innerLogin()
		log.Debug("%s 第%d次尝试登录", s.sdkAccount.Uid, i+1)
		err := func(c *session) error {
			var err error

			loginReq := models.NewSdkLoginReq()
			loginReq.Uid = c.sdkAccount.Uid
			loginReq.AccessKey = c.sdkAccount.AccessKey
			loginReq.Platform = c.sdkAccount.Platform
			loginReq.ChannelId = c.sdkAccount.Channel
			var loginResult models.BaseResponse[models.SdkLoginResp]

			if _, err = c.execReq(&loginReq, &loginResult); err != nil {
				return err
			}
			c.viewerId = loginResult.DataHeaders.ViewerId

			startReq := models.NewGameStartReq()
			var startResult models.BaseResponse[models.GameStartResp]

			if _, err = c.execReq(&startReq, &startResult); err != nil {
				return err
			}
			if !startResult.Data.NowTutorial {
				return errors.New("账号还没过教程！")
			}

			loadIndexReq := models.NewLoadIndexReq()
			var loadIndexResult models.BaseResponse[models.LoadIndexResp]

			if _, err = c.execReq(&loadIndexReq, &loadIndexResult); err != nil {
				return err
			}
			c.expireTime = loadIndexResult.Data.DailyResetTime

			homeIndexReq := models.NewHomeIndexReq()
			homeIndexReq.MessageId = 1
			homeIndexReq.GoldHistory = 0
			homeIndexReq.IsFirst = 1
			homeIndexReq.TipsIdList = []int{}
			var homeIndexResult models.BaseResponse[models.HomeIndexResp]

			if _, err = c.execReq(&homeIndexReq, &homeIndexResult); err != nil {
				return err
			}
			return nil
		}(s)
		if err != nil {
			log.Error("%s 第%d次登录失败: %v", s.sdkAccount.Uid, i+1, err)
			continue
		}
		s.logged = true
		return nil
	}

	return errors.New("登录失败")
}

// callApi 执行一般API
func (s *session) callApi(
	request models.IRequest,
	result models.IResponse,
) (*resty.Response, error) {
	var resp *resty.Response
	var err error
	if !s.logged {
		if err = s.login(); err != nil {
			return nil, err
		}
	}
	resp, err = s.execReq(request, result)
	if err != nil && err.(*models.ApiError).ApiCode == 3 {
		s.logged = false
		return nil, err
	}
	return resp, err
}

func (s *session) Close() {
	s.ctxCancel()
}
