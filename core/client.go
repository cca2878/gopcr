package core

import "gopcr/models"

// Client 封装了与游戏核心玩法相关的API调用
type Client struct {
	*session
}

// NewClient 创建一个新的 GameAPI 实例
// 它需要一个已经初始化好的 Client
func NewClient(sdkAccount SdkAccount, options ...SessionOption) (*Client, error) {
	s, err := newSession(sdkAccount, options...)
	if err != nil {
		return nil, err
	}
	return &Client{session: s}, nil
}

func (c *Client) HomeIndex() (*models.BaseResponse[models.HomeIndexResp], error) {
	homeIndexReq := models.NewHomeIndexReq()
	homeIndexReq.MessageId = 1
	homeIndexReq.GoldHistory = 0
	homeIndexReq.IsFirst = 1
	homeIndexReq.TipsIdList = []int{}
	var homeIndexResult models.BaseResponse[models.HomeIndexResp]

	_, err := c.callApi(&homeIndexReq, &homeIndexResult)
	if err != nil {
		return nil, err
	}
	return &homeIndexResult, nil
}

func (c *Client) Close() {
	c.session.Close()
}
