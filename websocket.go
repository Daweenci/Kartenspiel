package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Player struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type JoinLobbyMessage struct {
	Type    string `json:"type"`
	LobbyID string `json:"lobby_id"`
	ID      string `json:"id"`
	Name    string `json:"name"`
}

type CreateLobbyMessage struct {
	Type       string `json:"type"`
	LobbyName  string `json:"lobbyName"`
	MaxPlayers int    `json:"maxPlayers"`
	IsPrivate  bool   `json:"isPrivate"`
	Password   string `json:"password,omitempty"`
	PlayerID   string `json:"playerID"`
	PlayerName string `json:"playerName"`
}

type LeaveLobbyMessage struct {
	Type    string `json:"type"`
	LobbyID string `json:"lobby_id"`
	ID      string `json:"id"`
}

type Lobby struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	MaxPlayers int      `json:"maxPlayers"`
	IsPrivate  bool     `json:"isPrivate"`
	Password   string   `json:"password,omitempty"`
	Players    []Player `json:"players"`
}

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

	err = conn.WriteJSON(map[string]interface{}{
		"type":    "welcome",
		"id":      player.ID,
		"lobbies": lobbies,
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
			Type string `json:"type"`
		}
		if err := json.Unmarshal(msgBytes, &base); err != nil {
			log.Println("Invalid message format")
			continue
		}
		log.Println("Received message type:", base.Type)
		switch base.Type {
		case "join_lobby":
			var msg JoinLobbyMessage
			if err := json.Unmarshal(msgBytes, &msg); err != nil {
				log.Println("Invalid join_lobby message")
				continue
			}
			joinLobbyHandler(msg)

		case "leave_lobby":
			var msg LeaveLobbyMessage
			if err := json.Unmarshal(msgBytes, &msg); err != nil {
				log.Println("Invalid leave_lobby message")
				continue
			}
			leaveLobbyHandler(msg)

		case "create_lobby":
			var msg CreateLobbyMessage
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

func joinLobbyHandler(msg JoinLobbyMessage) {
	lobbiesLock.Lock()
	defer lobbiesLock.Unlock()

	for i, lobby := range lobbies {
		if lobby.ID == msg.LobbyID {
			player := Player{
				ID:   msg.ID,
				Name: msg.Name,
			}
			lobbies[i].Players = append(lobbies[i].Players, player)
			broadcastLobbies()
			break
		}
	}
}

func leaveLobbyHandler(msg LeaveLobbyMessage) {
	lobbiesLock.Lock()
	defer lobbiesLock.Unlock()

	for i, lobby := range lobbies {
		if lobby.ID == msg.LobbyID {
			for j, player := range lobby.Players {
				if player.ID == msg.ID {
					lobbies[i].Players = append(lobbies[i].Players[:j], lobbies[i].Players[j+1:]...)
					break
				}
			}
			broadcastLobbies()
			break
		}
	}
}

func createLobbyHandler(msg CreateLobbyMessage, conn *websocket.Conn) {
	lobbiesLock.Lock()
	defer lobbiesLock.Unlock()

	newLobby := Lobby{
		ID:         uuid.New().String(),
		Name:       msg.LobbyName,
		MaxPlayers: msg.MaxPlayers,
		IsPrivate:  msg.IsPrivate,
		Password:   msg.Password,
		Players: []Player{
			{
				Name: msg.PlayerName,
				ID:   msg.PlayerID,
			},
		},
	}
	lobbies = append(lobbies, newLobby)
	broadcastLobbies()

	var err = conn.WriteJSON(map[string]interface{}{
		"type":  "lobby_created",
		"lobby": newLobby,
	})
	if err != nil {
		log.Println("Error sending LobbyID:", err)
		return
	}
}

func broadcastLobbies() {
	log.Println("hiiiiiiiiiiiiiiiiiiiiiiilfeeeeeeeeeeeeeeeeee:", clients)
	clientsLock.Lock()
	defer clientsLock.Unlock()
	log.Println("Broadcasting lobby list to clients:", clients)
	for client := range clients {
		err := client.WriteJSON(map[string]interface{}{
			"type":    "lobby_list",
			"lobbies": lobbies,
		})
		if err != nil {
			client.Close()
			delete(clients, client)
			log.Println("broadcast Lobbies error:", err)
		}
		log.Println("Broadcasted lobby list to client")
	}
}
