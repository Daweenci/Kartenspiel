package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var (
	lobbies     = []Lobby{}
	lobbiesLock sync.Mutex
	clients     = make(map[*websocket.Conn]bool)
	upgrader    = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}

	players     = make(map[string]Player)
	playersLock sync.Mutex

	clientsLock sync.Mutex
)

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	player := Player{
		ID:   uuid.New().String(),
		Name: name,
	}
	playersLock.Lock()
	players[player.ID] = player
	playersLock.Unlock()

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading connection:", err)
		return
	}
	defer conn.Close()

	clientsLock.Lock()
	clients[conn] = true
	clientsLock.Unlock()

	lobbiesLock.Lock()
	responseLobbies := make([]LobbyWithoutPassword, 0, len(lobbies))
	for _, l := range lobbies {
		responseLobbies = append(responseLobbies, LobbyWithoutPassword{
			ID:         l.ID,
			Name:       l.Name,
			MaxPlayers: l.MaxPlayers,
			IsPrivate:  l.IsPrivate,
			Players:    l.Players,
		})

	}
	lobbiesLock.Unlock()

	err = conn.WriteJSON(map[string]interface{}{
		"type":    ResponseWelcome,
		"id":      player.ID,
		"lobbies": responseLobbies,
	})
	if err != nil {
		log.Println("Error sending welcome message:", err)
		return
	}

	for {
		_, msgBytes, err := conn.ReadMessage()
		if err != nil {
			log.Println("WebSocket read error:", err)
			break
		}

		var base struct {
			Type MessageType `json:"type"`
		}
		if err := json.Unmarshal(msgBytes, &base); err != nil {
			log.Println("Invalid message format")
			continue
		}
		log.Println("Received message type:", base.Type)
		switch base.Type {
		case RequestJoinLobby:
			var msg JoinLobbyRequest
			if err := json.Unmarshal(msgBytes, &msg); err != nil {
				log.Println("Invalid join_lobby message")
				continue
			}
			joinLobbyHandler(msg)

		case RequestLeaveLobby:
			var msg LeaveLobbyRequest
			if err := json.Unmarshal(msgBytes, &msg); err != nil {
				log.Println("Invalid leave_lobby message")
				continue
			}
			leaveLobbyHandler(msg)

		case RequestCreateLobby:
			var msg CreateLobbyRequest
			if err := json.Unmarshal(msgBytes, &msg); err != nil {
				log.Println("Invalid create_lobby message")
				continue
			}
			createLobbyHandler(msg, conn)
		default:
			log.Println("Unknown message type:", base.Type)
		}
	}
}
