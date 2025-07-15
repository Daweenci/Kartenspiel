package main

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/rs/xid"
)

type Lobby struct {
	ID      string
	Name    string
	Players []string
}

var (
	lobbies  = []Lobby{}
	clients  = make(map[*websocket.Conn]bool)
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}
	clientsLock sync.Mutex
)

func broadcastLobbies() {
	clientsLock.Lock()
	defer clientsLock.Unlock()
	for client := range clients {
		err := client.WriteJSON(map[string]interface{}{
			"type":    "lobby_list",
			"lobbies": lobbies,
		})
		if err != nil {
			client.Close()
			delete(clients, client)
		}
	}
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading connection:", err)
		return
	}
	defer conn.Close()

	for {
		var msg map[string]interface{}
		if err := conn.ReadJSON(&msg); err != nil {
			log.Println("WebSocket read error:", err)
			break
		}

		log.Printf("Received message: %+v", msg)

		switch msg["type"] {
		case "create_lobby":
			name, ok := msg["name"].(string)
			if ok && name != "" {
				newLobby := Lobby{
					ID:      xid.New().String(),
					Name:    name,
					Players: []string{},
				}
				lobbies = append(lobbies, newLobby)
				broadcastLobbies()
			}
		}
	}
}
