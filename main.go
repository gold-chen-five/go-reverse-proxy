package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"

	"github.com/gold-chen-five/go-reverse-proxy/config"
	"github.com/gold-chen-five/go-reverse-proxy/pkg"
	"golang.org/x/crypto/acme/autocert"
)

func main() {

	cfg, err := config.LoadConfig("setting.yaml")
	if err != nil {
		log.Fatal("(設定檔錯誤)", err)
	}

	// Create a router to handle different routes
	mux := http.NewServeMux()

	for _, server := range cfg.Servers {
		for _, route := range server.Routes {
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

	// Set up autocert manager for automatic TLS certificates
	domains := cfg.GetAllDomains()
	certManager := &autocert.Manager{
		Cache:      autocert.DirCache("certs"),         // Cache certificates on disk
		Prompt:     autocert.AcceptTOS,                 // Accept Let's Encrypt TOS automatically
		HostPolicy: autocert.HostWhitelist(domains...), // Replace with your domain(s)
	}

	for _, serverConfig := range cfg.Servers {
		go startTLSServer(serverConfig.Listen, mux, certManager)
	}

	// Redirect HTTP to HTTPS and handle ACME challenges
	go func() {
		http.ListenAndServe(":80", certManager.HTTPHandler(nil))
	}()

	select {} // Block forever
}

func startTLSServer(address string, handler http.Handler, certManager *autocert.Manager) {
	server := &http.Server{
		Addr:    address,
		Handler: handler,
		TLSConfig: &tls.Config{
			GetCertificate: certManager.GetCertificate,
		},
	}

	fmt.Printf("HTTPS Server started on %s...\n", address)
	if err := server.ListenAndServeTLS("", ""); err != nil {
		log.Fatal(err)
	}
}
