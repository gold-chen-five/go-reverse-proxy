package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gold-chen-five/go-reverse-proxy/config"
	"github.com/gold-chen-five/go-reverse-proxy/pkg"
)

func main() {

	setting, err := config.LoadConfig("setting.yaml")
	if err != nil {
		log.Fatal("(設定檔錯誤)", err)
	}

	// Create a router to handle different routes
	mux := http.NewServeMux()

	for _, settingConfig := range setting.Servers {
		for _, route := range settingConfig.Routes {
			proxy, err := pkg.NewProxyServer(route.Proxy.Upstream)
			if err != nil {
				log.Fatal(err)
			}

			mux.HandleFunc(route.Match.Path, func(w http.ResponseWriter, r *http.Request) {
				// check the host header
				if r.Host == route.Match.Host {
					proxy.ServeHTTP(w, r)
				} else {
					http.NotFound(w, r)
				}
			})
		}
	}

	for _, serverConfig := range setting.Servers {
		go startServer(serverConfig.Listen, mux)
	}

	select {} // Block forever
}

func startServer(address string, handler http.Handler) {
	server := &http.Server{
		Addr:    address,
		Handler: handler,
	}

	fmt.Printf("Server started on %s...\n", address)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
