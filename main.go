package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/gold-chen-five/go-reverse-proxy/proxy"
	"golang.org/x/crypto/acme/autocert"
)

func main() {
	// FILE FLAG
	flagName := flag.String("file", "setting", "Setting file for proxy")

	// Parse the command-line arguments
	flag.Parse()

	configFileName := *flagName + ".yaml"

	loader, err := proxy.NewConfigLoader(configFileName)
	if err != nil {
		log.Fatalf("Config loader fail: %v", err)
	}

	proxyServers, err := loader.CreateProxyServers()
	if err != nil {
		log.Fatalf("Creating server fail: %v", err)
	}

	// Set up autocert manager for automatic TLS certificates
	domains := loader.Config.GetAllDomains()
	certManager := &autocert.Manager{
		Cache:      autocert.DirCache("certs"),         // Cache certificates on disk
		Prompt:     autocert.AcceptTOS,                 // Accept Let's Encrypt TOS automatically
		HostPolicy: autocert.HostWhitelist(domains...), // Replace with your domain(s)
	}

	for listen, proxyServer := range proxyServers {
		if proxyServer.Ssl {
			go startTLSServer(listen, proxyServer.HttpHandler, certManager)
		} else {
			go startServer(listen, proxyServer.HttpHandler)
		}
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

func startServer(address string, handler http.Handler) {
	server := &http.Server{
		Addr:    address,
		Handler: handler,
	}

	fmt.Printf("HTTPS Server started on %s...\n", address)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
