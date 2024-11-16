package main

import (
	"fmt"
	"log"
	"net/http"

	"golang.org/x/net/websocket"
)

func main() {
	ports := []int{8081, 8082, 8083}

	for _, port := range ports {
		go startServer(port)
	}

	select {}
}

func startServer(port int) {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Server-ID", fmt.Sprintf("Backend-Server-%d", port))

		response := fmt.Sprintf(`{
            "server": "Backend-%d",
            "received_headers": {
                "user-agent": "%s",
                "x-forwarded-for": "%s",
                "x-real-ip": "%s"
            },
            "path": "%s",
            "method": "%s"
        }`, port,
			r.Header.Get("User-Agent"),
			r.Header.Get("X-Forwarded-For"),
			r.Header.Get("X-Real-IP"),
			r.URL.Path,
			r.Method)

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, response)
	})

	// health route
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "OK from backend %d", port)
	})

	// websocket route
	mux.Handle("/ws", websocket.Handler(func(ws *websocket.Conn) {
		handleWebsocket(ws)
	}))

	serverAddr := fmt.Sprintf(":%d", port)
	log.Printf("start server at port %s", serverAddr)
	if err := http.ListenAndServe(serverAddr, mux); err != nil {
		log.Fatal(err)
	}
}

func handleWebsocket(ws *websocket.Conn) {
	defer ws.Close()

	// handle ws message
	for {
		// Read message from the client
		var message string
		if err := websocket.Message.Receive(ws, &message); err != nil {
			log.Printf("Error reading WebSocket message: %v", err)
			break
		}
		log.Printf("Received message on server: %s", message)

		// Echo the message back to the client
		response := fmt.Sprintf("Server received: %s", message)
		if err := websocket.Message.Send(ws, response); err != nil {
			log.Printf("Error sending WebSocket message: %v", err)
			break
		}
	}
}
