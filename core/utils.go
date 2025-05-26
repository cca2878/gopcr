package core

import (
	"math/rand"
	"time"
)

// RandEvenNum 生成随机偶数，0 <= n <= 100000
func RandEvenNum() int {
	// 使用当前时间纳秒作为种子创建新的随机源
	source := rand.NewSource(time.Now().UnixNano())
	// 偶数
	return rand.New(source).Intn(50001) * 2
}
