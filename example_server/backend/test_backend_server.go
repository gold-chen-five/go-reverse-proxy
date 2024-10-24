package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

func main() {
	// use terminal to decide port
	port := flag.Int("port", 8081, "server port")
	flag.Parse()

	// default route
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Server-ID", fmt.Sprintf("Backend-Server-%d", *port))

		response := fmt.Sprintf(`{
            "server": "Backend-%d",
            "received_headers": {
                "user-agent": "%s",
                "x-forwarded-for": "%s",
                "x-real-ip": "%s"
            },
            "path": "%s",
            "method": "%s"
        }`, *port,
			r.Header.Get("User-Agent"),
			r.Header.Get("X-Forwarded-For"),
			r.Header.Get("X-Real-IP"),
			r.URL.Path,
			r.Method)

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, response)
	})

	// health route
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "OK from backend %d", *port)
	})

	serverAddr := fmt.Sprintf(":%d", *port)
	log.Printf("start server at port %s", serverAddr)
	if err := http.ListenAndServe(serverAddr, nil); err != nil {
		log.Fatal(err)
	}
}
