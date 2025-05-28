package core

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// RandEvenNum 生成随机偶数，0 <= n <= 100000
//func RandEvenNum() int {
//	// 使用当前时间纳秒作为种子创建新的随机源
//	source := rand.NewSource(time.Now().UnixNano())
//	// 偶数
//	return rand.New(source).Intn(50001) * 2
//}

func getNewAppVer() (string, error) {
	const url = "https://line1-h5-pc-api.biligame.com/game/detail/content?game_base_id=102216"

	client := http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to get app version: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var result struct {
		Data struct {
			AndroidVersion string `json:"android_version"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("invalid JSON: %w", err)
	}

	if result.Data.AndroidVersion == "" {
		return "", fmt.Errorf("android_version is empty")
	}

	return result.Data.AndroidVersion, nil
}
