package models

import (
	"net/http"
	"net/url"
)

// Req Resp base interface

type IRequest interface {
	// SetViewerId 设置用户ID
	SetViewerId(viewerId string)
	// IsEncrypt 是否需要加密
	IsEncrypt() bool
	// GetMethod Http Method
	GetMethod() string
	// GetUrl API Url
	GetUrl() (*url.URL, error)
}
type IResponse interface {
	// GetRequestId 请求ID
	GetRequestId() string
	// GetSID SID
	GetSID() string
	// GetResultCode 响应结果码
	GetResultCode() int
}

type RespData interface {
}

func parseModelUrl(path string) (*url.URL, error) {
	//u, err := url.Parse("https://" + config.GetInstance().GetOptVal(config.PcrApiHost).(string) + path)
	// 改由client处理api root
	u, err := url.Parse(path)

	if err != nil {
		return nil, err
	}
	return u, nil
}

// ConfigRequest 添加关联配置的请求接口
type ConfigRequest interface {
	IRequest
	//SetConfig(config *config.Config)
}

type BaseRequest struct {
	isEncrypt bool
	ViewerId  string `json:"viewer_id"`
}

func (r *BaseRequest) SetViewerId(viewerId string) {
	r.ViewerId = viewerId
}

func (r *BaseRequest) IsEncrypt() bool {
	return r.isEncrypt
}

func (r *BaseRequest) GetMethod() string {
	return http.MethodPost
}

type dataHeaders struct {
	Sid        string `json:"sid"`
	ViewerId   uint64 `json:"viewer_id"`
	RequestId  string `json:"request_id"`
	ResultCode int    `json:"result_code"`

	//StoreUrl *string `json:"store_url"`
}

type BaseResponse[T RespData] struct {
	DataHeaders dataHeaders `json:"data_headers"`
	Data        T           `json:"data"`
}

func (b BaseResponse[T]) GetResultCode() int {
	return b.DataHeaders.ResultCode
}

func (b BaseResponse[T]) GetRequestId() string {
	return b.DataHeaders.RequestId
}

func (b BaseResponse[T]) GetSID() string {
	return b.DataHeaders.Sid
}
func NewBaseRequest() BaseRequest {
	return BaseRequest{
		isEncrypt: true,
	}
}
