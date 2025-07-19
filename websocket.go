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
	lobbies     = make(map[string]*Lobby)
	lobbiesLock sync.Mutex
	upgrader    = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}
	players     = make(map[string]*Player)
	playersLock sync.Mutex
)

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading connection:", err)
		return
	}

	name := r.URL.Query().Get("name")
	player := &Player{
		ID:   uuid.New().String(),
		Name: name,
		Conn: conn,
	}

	defer func() {
		conn.Close()
		log.Println("Connection closed for player:", player.ID)
		playersLock.Lock()
		var breakLoop = false
		for _, lobby := range lobbies {
			for index, compPlayer := range lobby.Players {
				if compPlayer.ID == player.ID {
					lobby.Players = append(lobby.Players[:index], lobby.Players[index+1:]...)
					if len(lobby.Players) == 0 {
						delete(lobbies, lobby.ID)
					}
					breakLoop = true
					break
				}
			}
			if breakLoop {
				break
			}
		}
		delete(players, player.ID)
		playersLock.Unlock()
		broadcastLobbies()
	}()

	playersLock.Lock()
	players[player.ID] = player
	playersLock.Unlock()

	lobbiesLock.Lock()
	responseLobbies := make([]BroadcastedLobby, 0, len(lobbies))
	for _, l := range lobbies {
		responseLobbies = append(responseLobbies, BroadcastedLobby{
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
		"name":    player.Name,
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
			createLobbyHandler(msg)

		case RequestStartGame:
			var msg StartGame
			if err := json.Unmarshal(msgBytes, &msg); err != nil {
				log.Println("Invalid start_game message")
				continue
			}
			startGameHandler(msg)

		case RequestCancelGame:
			var msg CancelGame
			if err := json.Unmarshal(msgBytes, &msg); err != nil {
				log.Println("Invalid cancel_game message")
				continue
			}
			cancelGameHandler(msg)

		default:
			log.Println("Unknown message type:", base.Type)
		}
	}
}
