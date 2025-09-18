// websocket.go (updated sections)
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var (
	lobbies = make(map[string]*Lobby)
	lobbiesLock sync.Mutex

	// Keep active connections in memory for real-time communication
	activeConnections     = make(map[string]*Player)
	activeConnectionsLock sync.Mutex

	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}

	jwtSecret = []byte("YOUR_SECRET_KEY")
)

func sendErrorToPlayer(player *Player, errorMsg string) {
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

	activeConnectionsLock.Lock()
	activeConnections[player.ID] = player
	activeConnectionsLock.Unlock()

	defer func() {
		conn.Close()
		log.Println("Connection closed for player:", player.ID)

		activeConnectionsLock.Lock()
		defer activeConnectionsLock.Unlock()

		// Remove player from lobbies (different from inGameLobby)
		lobbiesLock.Lock()
		for _, lobby := range lobbies {
			for i, p := range lobby.Players {
				if p.ID == player.ID {
					lobby.Players = append(lobby.Players[:i], lobby.Players[i+1:]...)
					if len(lobby.Players) == 0 {
						delete(lobbies, lobby.ID)
					}
					break
				}
			}
		}
		lobbiesLock.Unlock()

		delete(activeConnections, player.ID)
		broadcastLobbies()
	}()

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

	conn.WriteJSON(map[string]interface{}{
		"type":    ResponseWelcome,
		"name":    player.Name,
		"id":      player.ID,
		"lobbies": responseLobbies,
	})

	for {
		_, msgBytes, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error for player %s: %v", player.ID, err)
			}
			break
		}

		var base struct {
			Type  MessageType `json:"type"`
			Token string      `json:"token"`
		}
		if err := json.Unmarshal(msgBytes, &base); err != nil {
			log.Println("Invalid message format")
			sendErrorToPlayer(player, "Invalid message format")
			continue
		}

		if base.Type != RequestLogin && base.Type != RequestRegister {
			playerID, err := parseJWT(base.Token)
			if err != nil {
				sendErrorToPlayer(player, "Invalid or expired token")
				continue
			}
			
			activeConnectionsLock.Lock()
			authPlayer, ok := activeConnections[playerID]
			activeConnectionsLock.Unlock()
			
			if !ok {
				dbPlayer, err := getPlayerByID(playerID)
				if err != nil {
					sendErrorToPlayer(player, "Player not found")
					continue
				}
				
				player.ID = dbPlayer.ID
				player.Name = dbPlayer.Username
				
				activeConnectionsLock.Lock()
				activeConnections[player.ID] = player
				activeConnectionsLock.Unlock()
			} else {
				player = authPlayer
			}
		}

		switch base.Type {
			case RequestJoinLobby:
				var msg JoinLobbyRequest
				if err := json.Unmarshal(msgBytes, &msg); err != nil {
					sendErrorToPlayer(player, "Invalid join_lobby message")
					continue
				}
				joinLobbyHandler(msg)

			case RequestLeaveLobby:
				var msg LeaveLobbyRequest
				if err := json.Unmarshal(msgBytes, &msg); err != nil {
					sendErrorToPlayer(player, "Invalid leave_lobby message")
					continue
				}
				leaveLobbyHandler(msg)

			case RequestCreateLobby:
				var msg CreateLobbyRequest
				if err := json.Unmarshal(msgBytes, &msg); err != nil {
					sendErrorToPlayer(player, "Invalid create_lobby message")
					continue
				}
				createLobbyHandler(msg)

			case RequestStartGame:
				var msg StartGame
				if err := json.Unmarshal(msgBytes, &msg); err != nil {
					sendErrorToPlayer(player, "Invalid start_game message")
					continue
				}
				startGameHandler(msg)

			case RequestCancelGame:
				var msg CancelGame
				if err := json.Unmarshal(msgBytes, &msg); err != nil {
					sendErrorToPlayer(player, "Invalid cancel_game message")
					continue
				}
				cancelGameHandler(msg)

			default:
				sendErrorToPlayer(player, "Unknown message type")
		}
	}
}