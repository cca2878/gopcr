package config

// 维护一些运行时更新的配置

import (
	"reflect"
	"sync"
)

type optionType int

const (
	AppVer     optionType = iota // 游戏版本
	PcrApiHost optionType = iota // API主机
)

type OptionConfig struct {
	mu      sync.RWMutex
	options map[optionType]any
}

type defaultOptVal struct {
	value any
	typ   reflect.Type
}

var defaultOptVals = map[optionType]defaultOptVal{
	AppVer:     {value: DefaultAppVer, typ: reflect.TypeOf(DefaultAppVer)},
	PcrApiHost: {value: DefaultBiliApiHost, typ: reflect.TypeOf(DefaultBiliApiHost)},
}

// 单例相关变量
var (
	instance *OptionConfig
	once     sync.Once
)

// GetInstance 获取配置的单例实例
func GetInstance() *OptionConfig {
	once.Do(func() {
		instance = &OptionConfig{
			options: make(map[optionType]any),
		}
	})
	return instance
}

// SetOptVal 设置单个配置项
func (h *OptionConfig) SetOptVal(optType optionType, value any) error {
	// 比较值是否相同，如果相同则直接返回
	if h.GetOptVal(optType) == value {
		return nil
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	h.options[optType] = value
	return nil
}

// GetOptVal 获取配置项的值
func (h *OptionConfig) GetOptVal(optType optionType) any {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if val, ok := h.options[optType]; ok {
		return val
	}
	return defaultOptVals[optType].value
}
