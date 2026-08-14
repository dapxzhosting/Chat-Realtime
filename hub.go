package main

import "log"

// Message adalah struktur pesan yang dikirim antar client
type Message struct {
	Username string `json:"username"`
	Content  string `json:"content"`
	Type     string `json:"type"` // "chat", "join", "leave"
}

// Hub menyimpan semua client yang aktif dan mengatur broadcast pesan
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan Message
	register   chan *Client
	unregister chan *Client
}

func newHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
    h.clients[client] = true
    log.Printf("client bergabung: %s (total: %d)", client.username, len(h.clients))
    joinMsg := Message{Username: client.username, Type: "join", Content: "bergabung ke chat"}
    go func() { h.broadcast <- joinMsg }()

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				log.Printf("client keluar: %s (total: %d)", client.username, len(h.clients))
			}

		case msg := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- msg:
				default:
					// buffer client penuh / macet -> anggap putus, bersihkan
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}