// Package config 提供项目中使用的配置
package config

import "time"

const (
	PcrAesIV = "ha4nBYA2APUD6Uv1"
)

// 请求相关常量
const (
	// DefaultBiliApiHost 默认API Host
	DefaultBiliApiHost    = "le1-prod-all-gs-gzlj.bilibiligame.net/"
	DefaultChannelApiHost = "l1-prod-uo-gs-gzlj.bilibiligame.net/"
	// DefaultRequestTimeout 默认请求超时时间(秒)
	DefaultRequestTimeout = 5 * time.Second
)

// Headers相关常量
const (
	// DefaultAppVer 默认App版本
	DefaultAppVer = "8.1.0"
	// BiliResKey b服reskey
	biliResKey    = "ab00a0a6dd915a052a2ef7fd649083e5"
	channelResKey = "d145b29050641dac2f8b19df0afe0e59"
)

func getDefaultHeaders() map[string]string {
	return map[string]string{
		"Accept-Encoding":      "deflate, gzip",
		"User-Agent":           "UnityPlayer/2021.3.36f1c1 (UnityWebRequest/1.0, libcurl/8.5.0-DEV)",
		"X-Unity-Version":      "2021.3.36f1c1",
		"APP-VER":              GetInstance().GetOptVal(AppVer).(string),
		"BATTLE-LOGIC-VERSION": "4",
		"DEVICE":               "2",
		"DEVICE-ID":            "ln-nmsl",
		"DEVICE-NAME":          "LN_NMSL",
		"EXCEL-VER":            "1.0.0",
		"GRAPHICS-DEVICE-NAME": "LN_NMSL",
		"IP-ADDRESS":           "0.0.0.0",
		"LOCALE":               "Jpn",
		"PLATFORM-OS-VERSION":  "RedStar OS - GoPcr",
		"REGION-CODE":          "CN",
		"RES-VER":              "10002200",
		"SHORT-UDID":           "0",
	}
}

func GetBiliHeaders() map[string]string {
	defaultHeaders := getDefaultHeaders()
	defaultHeaders["RES-KEY"] = biliResKey
	return defaultHeaders
}

func GetChannelHeaders() map[string]string {
	defaultHeaders := getDefaultHeaders()
	defaultHeaders["RES-KEY"] = channelResKey
	return defaultHeaders
}
