package main

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

const defaultPath = "testproxy.ddns.net"

func main() {
	// 創建 HTTP 客戶端
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// 執行測試
	for i := 0; i < 6; i++ {
		testProxyHttp(client, "/")
		time.Sleep(1 * time.Second) // 延遲以便觀察負載均衡效果
	}

	// 測試不同路徑
	testProxyHttp(client, "/api/test")
	testProxyHttp(client, "/health")
}

// http test proxy
func testProxyHttp(client *http.Client, path string) {
	url := fmt.Sprintf("https://%s%s", defaultPath, path)
	resp, err := client.Get(url)
	if err != nil {
		fmt.Printf("請求失敗: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// 讀取響應
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("讀取響應失敗: %v\n", err)
		return
	}

	// 打印響應信息
	fmt.Printf("\n=== 請求 %s ===\n", url)
	fmt.Printf("狀態碼: %d\n", resp.StatusCode)
	fmt.Printf("Server-ID: %s\n", resp.Header.Get("Server-ID"))
	fmt.Printf("響應體: %s\n", string(body))
}

// test proxy websocket
func testProxyWebsocket() {

}
