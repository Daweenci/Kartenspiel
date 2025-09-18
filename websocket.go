// websocket.go (updated sections)
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var (
	lobbies     = make(map[string]*Lobby)
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
	// Upgrade to WebSocket immediately (no auth check yet)
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading connection:", err)
		return
	}

	// Create temporary player for initial connection
	var player *Player
	authenticated := false

	defer func() {
		conn.Close()
		if player != nil {
			log.Println("Connection closed for player:", player.ID)

			activeConnectionsLock.Lock()
			defer activeConnectionsLock.Unlock()

			// Remove player from lobbies
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
			setPlayerOnlineStatus(player.ID, false)
			broadcastLobbies()
		}
	}()

	// Wait for authentication message
	for {
		_, msgBytes, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		var base struct {
			Type  MessageType `json:"type"`
			Token string      `json:"token,omitempty"`
		}
		if err := json.Unmarshal(msgBytes, &base); err != nil {
			conn.WriteJSON(map[string]interface{}{
				"type":  ResponseError,
				"error": "Invalid message format",
			})
			continue
		}

		// If not authenticated yet, only accept authenticate messages
		if !authenticated {
			if base.Type != "authenticate" || base.Token == "" {
				conn.WriteJSON(map[string]interface{}{
					"type":  ResponseError,
					"error": "Authentication required",
				})
				continue
			}

			// Validate JWT token
			playerID, err := parseJWT(base.Token)
			if err != nil {
				conn.WriteJSON(map[string]interface{}{
					"type":  ResponseError,
					"error": "Invalid or expired token",
				})
				continue
			}

			// Get player from database
			dbPlayer, err := getPlayerByID(playerID)
			if err != nil {
				conn.WriteJSON(map[string]interface{}{
					"type":  ResponseError,
					"error": "Player not found",
				})
				continue
			}

			// Create authenticated player
			player = &Player{
				ID:   dbPlayer.ID,
				Name: dbPlayer.Username,
				Conn: conn,
			}

			// Add to active connections
			activeConnectionsLock.Lock()
			activeConnections[player.ID] = player
			activeConnectionsLock.Unlock()

			// Set online status
			setPlayerOnlineStatus(player.ID, true)
			authenticated = true

			// Send welcome message
			conn.WriteJSON(map[string]interface{}{
				"type": ResponseWelcome,
				"player": map[string]string{
					"id":   player.ID,
					"name": player.Name,
				},
				"message": "Welcome back, " + player.Name + "!",
				"lobbies": getLobbiesList(),
			})
			continue
		}

		// Handle authenticated messages
		switch base.Type {
		case RequestJoinLobby:
			var msg JoinLobbyRequest
			if err := json.Unmarshal(msgBytes, &msg); err != nil {
				sendErrorToPlayer(player, "Invalid join_lobby message")
				continue
			}
			msg.PlayerID = player.ID
			msg.Name = player.Name
			joinLobbyHandler(msg)

		case RequestLeaveLobby:
			var msg LeaveLobbyRequest
			if err := json.Unmarshal(msgBytes, &msg); err != nil {
				sendErrorToPlayer(player, "Invalid leave_lobby message")
				continue
			}
			msg.PlayerID = player.ID
			leaveLobbyHandler(msg)

		case RequestCreateLobby:
			var msg CreateLobbyRequest
			if err := json.Unmarshal(msgBytes, &msg); err != nil {
				sendErrorToPlayer(player, "Invalid create_lobby message")
				continue
			}
			msg.PlayerID = player.ID
			msg.PlayerName = player.Name
			createLobbyHandler(msg)

		case RequestStartGame:
			var msg StartGame
			if err := json.Unmarshal(msgBytes, &msg); err != nil {
				sendErrorToPlayer(player, "Invalid start_game message")
				continue
			}
			msg.PlayerID = player.ID
			startGameHandler(msg)

		case RequestCancelGame:
			var msg CancelGame
			if err := json.Unmarshal(msgBytes, &msg); err != nil {
				sendErrorToPlayer(player, "Invalid cancel_game message")
				continue
			}
			msg.PlayerID = player.ID
			cancelGameHandler(msg)

		default:
			sendErrorToPlayer(player, "Unknown message type")
		}
	}
}

// Helper function to get lobbies list
func getLobbiesList() []BroadcastedLobby {
	lobbiesLock.Lock()
	defer lobbiesLock.Unlock()

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
	return responseLobbies
}
