package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"golang.org/x/net/websocket"
)

const defaultPath = "localhost:8080"
const sslPath = "test2proxy.zapto.org"
const routePath = ""

var isSSL = false

func main() {

	// check is using --ssl
	for _, arg := range os.Args {
		if arg == "--ssl" {
			isSSL = true
			break
		}
	}

	// 創建 HTTP 客戶端
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// 執行測試
	for i := 0; i < 10; i++ {
		testProxyHttp(client, routePath)
		time.Sleep(1 * time.Second) // 延遲以便觀察負載均衡效果
	}

	// 測試不同路徑
	testProxyHttp(client, fmt.Sprintf("%s%s", routePath, "/api/test"))
	testProxyHttp(client, fmt.Sprintf("%s%s", routePath, "/health"))
	testProxyWebsocket(routePath)
}

// http test proxy
func testProxyHttp(client *http.Client, path string) {
	var url string
	if isSSL {
		url = fmt.Sprintf("https://%s%s", sslPath, path)
	} else {
		url = fmt.Sprintf("http://%s%s", defaultPath, path)
	}

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
func testProxyWebsocket(path string) {
	var wsPath string
	var origin string

	if isSSL {
		wsPath = fmt.Sprintf("wss://%s%s%s", sslPath, path, "/ws")
		origin = fmt.Sprintf("https://%s%s", sslPath, path)
	} else {
		wsPath = fmt.Sprintf("ws://%s%s%s", defaultPath, path, "/ws")
		origin = fmt.Sprintf("http://%s%s", defaultPath, path)
	}

	ws, err := websocket.Dial(wsPath, "", origin)

	if err != nil {
		log.Fatalf("Fail to connect websocket %v", err)
	}
	defer ws.Close()

	sendMessage := func(message string) {
		// Send message
		_, err := ws.Write([]byte(message))
		if err != nil {
			log.Printf("Failed to send message: %v\n", err)
			return
		}
		fmt.Printf("Send: %s\n", message)

		// Read response
		var response = make([]byte, 512)
		n, err := ws.Read(response)
		if err != nil {
			log.Printf("Failed to read response: %v\n", err)
			return
		}
		fmt.Printf("Received: %s\n", string(response[:n]))
	}

	// Run the test
	for i := 0; i < 5; i++ {
		sendMessage(fmt.Sprintf("Hello WebSocket %d", i+1))
		time.Sleep(1 * time.Second) // Small delay between messages
	}
}
