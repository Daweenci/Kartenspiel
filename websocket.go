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

// Helper function to send error responses
func sendErrorToPlayer(player *Player, errorMsg string) { //could include errorType for more specific json type but for now this is fine
	err := player.Conn.WriteJSON(map[string]interface{}{
		"type":  ResponseError,
		"error": errorMsg,
	})
	if err != nil {
		log.Printf("Failed to send error message to player %s: %v", player.ID, err)
	}
}

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
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error for player %s: %v", player.ID, err)
			}
			break
		}

		var base struct {
			Type MessageType `json:"type"`
		}
		if err := json.Unmarshal(msgBytes, &base); err != nil {
			log.Println("Invalid message format")
			sendErrorToPlayer(player, "Invalid message format")
			continue
		}
		log.Println("Received message type:", base.Type)

		switch base.Type {
		case RequestJoinLobby:
			var msg JoinLobbyRequest
			if err := json.Unmarshal(msgBytes, &msg); err != nil {
				log.Println("Invalid join_lobby message")
				sendErrorToPlayer(player, "Invalid join_lobby message format")
				continue
			}
			joinLobbyHandler(msg)

		case RequestLeaveLobby:
			var msg LeaveLobbyRequest
			if err := json.Unmarshal(msgBytes, &msg); err != nil {
				log.Println("Invalid leave_lobby message")
				sendErrorToPlayer(player, "Invalid leave_lobby message format")
				continue
			}
			leaveLobbyHandler(msg)

		case RequestCreateLobby:
			var msg CreateLobbyRequest
			if err := json.Unmarshal(msgBytes, &msg); err != nil {
				log.Println("Invalid create_lobby message")
				sendErrorToPlayer(player, "Invalid create_lobby message format")
				continue
			}
			createLobbyHandler(msg)

		case RequestStartGame:
			var msg StartGame
			if err := json.Unmarshal(msgBytes, &msg); err != nil {
				log.Println("Invalid start_game message")
				sendErrorToPlayer(player, "Invalid start_game message format")
				continue
			}
			startGameHandler(msg)

		case RequestCancelGame:
			var msg CancelGame
			if err := json.Unmarshal(msgBytes, &msg); err != nil {
				log.Println("Invalid cancel_game message")
				sendErrorToPlayer(player, "Invalid cancel_game message format")
				continue
			}
			cancelGameHandler(msg)

		default:
			log.Println("Unknown message type:", base.Type)
			sendErrorToPlayer(player, "Unknown message type")
		}
	}
}
