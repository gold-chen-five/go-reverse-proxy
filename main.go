package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gold-chen-five/go-reverse-proxy/pkg"
)

func main() {
	// 配置上游伺服器
	upstreamURLs := []string{
		"http://localhost:8081",
		"http://localhost:8082",
		"http://localhost:8083",
	}
	proxy, err := pkg.NewProxyServer(upstreamURLs)
	if err != nil {
		log.Fatal(err)
	}

	// 創建 HTTP 伺服器
	server := &http.Server{
		Addr:    ":8080",
		Handler: proxy,
	}

	fmt.Println("反向代理伺服器啟動在 :8080...")
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
