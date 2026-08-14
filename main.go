package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,

	// Untuk production sebaiknya dibatasi ke domain frontend
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func serveWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")

	if username == "" {
		username = "anonim"
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade error:", err)
		return
	}

	client := &Client{
		hub:      hub,
		conn:     conn,
		send:     make(chan Message, 256),
		username: username,
	}

	client.hub.register <- client

	go client.writePump()
	go client.readPump()
}

func main() {
	hub := newHub()
	go hub.run()

	// Serve frontend
	http.Handle("/", http.FileServer(http.Dir("./static")))

	// WebSocket endpoint
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})

	// Railway menggunakan PORT
	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	addr := ":" + port

	log.Println("Chat server berjalan di port", port)

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal("server error:", err)
	}
}