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
)

type SdkAccount struct {
	Uid       string
	AccessKey string

	Platform string
	Channel  string
}

// Client 实现Princess Connect Re:Dive的API客户端
type Client struct {
	ctx        context.Context    // 内部创建的 context
	ctxCancel  context.CancelFunc // 用于取消 context
	httpClient *resty.Client
	crypto     *PCRCrypto
	sdkAccount SdkAccount
	logged     bool
	viewerId   uint
	expireTime uint
}

// ClientOption 定义客户端选项
type ClientOption func(*Client)

// WithChannelServer 渠道服Option
func WithChannelServer() ClientOption {
	return func(client *Client) {
		client.httpClient.
			SetBaseURL("https://"+config.DefaultChannelApiHost).
			SetHeader("PLATFORM-ID", "4").
			SetHeaders(config.GetBiliHeaders())
	}
}

// NewClient 创建一个新的Client。
// 默认为B服，渠道服用上面的Option。
func NewClient(sdkAccount SdkAccount, options ...ClientOption) (*Client, error) {

	// 创建一个带有取消功能的 context
	ctx, cancel := context.WithCancel(context.Background())
	// 创建并配置 HTTP 客户端
	httpClient := resty.New().
		SetBaseURL("https://" + config.DefaultBiliApiHost).
		// Debug
		SetProxy("http://127.0.0.1:7890").
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
	client := &Client{
		httpClient: httpClient,
		ctx:        ctx,
		ctxCancel:  cancel,
		viewerId:   0,
		sdkAccount: sdkAccount,
		logged:     false,
		crypto:     NewPCRCrypto(),
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

func (c *Client) preReq(ctx context.Context, request models.IRequest) (*resty.Request, error) {
	req := c.httpClient.R().
		SetContext(ctx)

	//if c.sid != "" {
	//	req.SetHeader("SID", calcSID(c.sid))
	//}

	if request.IsEncrypt() {
		encryptViewerId, err := c.crypto.EncryptViewerId(c.viewerId)
		if err != nil {
			return nil, fmt.Errorf("加密ViewerId失败: %w", err)
		}
		request.SetViewerId(encryptViewerId)

		data, err := c.crypto.EncryptData(request)
		if err != nil {
			return nil, fmt.Errorf("加密数据失败: %w", err)
		}
		req.SetBody(data)
	} else {
		request.SetViewerId(strconv.FormatUint(uint64(c.viewerId), 10))

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
func (c *Client) execReq(
	ctx context.Context,
	request models.IRequest,
	result models.IResponse,
) (*resty.Response, error) {
	// 请求预处理
	req, err := c.preReq(ctx, request)
	if err != nil {
		log.Error("准备请求失败: %v", err)
		return nil, err
	}
	//if c.requestId != "" {
	//	req.SetHeader("REQUEST-ID", c.requestId)
	//}
	url, err := request.GetUrl()
	if err != nil {
		log.Error("获取URL失败: %v", err)
		return nil, err
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
		return resp, err
	}

	log.Debug("收到响应: 状态码=%d, 内容长度=%d", resp.StatusCode(), len(resp.Body()))

	if resp.StatusCode() != http.StatusOK {
		log.Error("请求失败，状态码: %d", resp.StatusCode())
		return resp, fmt.Errorf("请求失败，状态码: %d", resp.StatusCode())
	}

	// 对加密响应进行解密处理
	if request.IsEncrypt() {
		err := c.crypto.DecryptData(string(resp.Body()), result)
		if err != nil {
			log.Error("响应解密失败: %v", err)
			return resp, err
		}
	}

	// 请求后处理
	if reqId := result.GetRequestId(); reqId != "" {
		c.httpClient.SetHeader("REQUEST-ID", reqId)
	}
	if sid := result.GetSID(); sid != "" {
		c.httpClient.SetHeader("SID", calcSID(sid))
	}

	return resp, nil
}

func (c *Client) getConfig() error {
	indexReq := models.NewSourceIniIndexReq()
	var indexResult models.BaseResponse[models.SourceIniIndexResp]
	_, err := c.execReq(c.ctx, &indexReq, &indexResult)
	if err != nil {
		return err
	}
	maintenanceReq := models.NewSourceIniGetMaintenanceStatusReq()
	var maintenanceResult models.BaseResponse[models.SourceIniGetMaintenanceStatusResp]
	_, err = c.execReq(c.ctx, &maintenanceReq, &maintenanceResult)
	if err != nil {
		return err
	}
	if maintenanceResult.Data.ManifestVer != "" {
		c.httpClient.SetHeader("MANIFEST-VER", maintenanceResult.Data.ManifestVer)
	}

	return nil
}

func (c *Client) Login() error {
	// 登录流程：
	// sdk_login: 用bsdk拿到的uid和access_key来登录，暂未测试uid是否必需。完成后拿到滚动req id，放headers
	// game_start: 用uid和viewer_id来启动游戏
	// load_index: 仿照真实流程
	// home_index: 仿照真实流程

	// 登录逻辑
	if c.logged {
		return errors.New("already logged in")
	}
	loginReq := models.NewSdkLoginReq()
	loginReq.Uid = c.sdkAccount.Uid
	loginReq.AccessKey = c.sdkAccount.AccessKey
	loginReq.Platform = c.sdkAccount.Platform
	loginReq.ChannelId = c.sdkAccount.Channel

	var loginResult models.BaseResponse[models.SdkLoginResp]
	_, err := c.execReq(c.ctx, &loginReq, &loginResult)
	if err != nil {
		return err
	}
	c.viewerId = loginResult.DataHeaders.ViewerId

	startReq := models.NewGameStartReq()
	var startResult models.BaseResponse[models.GameStartResp]
	_, err = c.execReq(c.ctx, &startReq, &startResult)
	if !startResult.Data.NowTutorial {
		return errors.New("账号还没过教程！")
	}

	loadIndexReq := models.NewLoadIndexReq()
	var loadIndexResult models.BaseResponse[models.LoadIndexResp]
	_, err = c.execReq(c.ctx, &loadIndexReq, &loadIndexResult)
	if err != nil {
		return err
	}
	c.expireTime = loadIndexResult.Data.DailyResetTime

	homeIndexReq := models.NewHomeIndexReq()
	homeIndexReq.MessageId = 1
	homeIndexReq.GoldHistory = 0
	homeIndexReq.IsFirst = 1
	homeIndexReq.TipsIdList = []int{}
	var homeIndexResult models.BaseResponse[models.HomeIndexResp]
	_, err = c.execReq(c.ctx, &homeIndexReq, &homeIndexResult)

	c.logged = true

	return nil
}
