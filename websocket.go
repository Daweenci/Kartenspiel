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
	lobbiesLock sync.RWMutex

	// Keep active connections in memory for real-time communication
	activePlayers     = make(map[string]*Player) // TODO: Active players, not connections
	activePlayersLock sync.RWMutex

	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}

	jwtSecret = []byte("YOUR_SECRET_KEY")
)

func sendErrorToPlayer(player *Player, errorMsg string) {
	sendErrorViaConn(player.Conn, errorMsg)
}

func sendErrorViaConn(conn *websocket.Conn, errorMsg string) {
	err := conn.WriteJSON(ErrorResponse{
		Type:  ResponseError,
		Error: errorMsg,
	})
	if err != nil {
		log.Printf("Failed to send error message to connection %s: %v", conn.RemoteAddr(), err)
	}
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	log.Printf("NEW WEBSOCKET CONNECTION from %s", r.RemoteAddr)
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
			log.Printf("Player %s removed from lobbies", player.ID)

			delete(activePlayers, player.ID)
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
			sendErrorViaConn(conn, "Invalid message format")
			continue
		}

		// If not authenticated yet, only accept authenticate messages
		if !authenticated {
			if base.Type != RequestAuthentication || base.Token == "" {
				sendErrorViaConn(conn, "Authentication required")
				continue
			}

			// Validate JWT token
			playerID, err := parseJWT(base.Token)
			if err != nil {
				sendErrorViaConn(conn, "Invalid or expired token")
				continue
			}

			// Get player from database
			dbPlayer, err := getPlayerByID(playerID)
			if err != nil {
				sendErrorViaConn(conn, "Player not found")
				continue
			}

			// Create authenticated player
			player = &Player{
				ID:   dbPlayer.ID,
				Name: dbPlayer.Username,
				Conn: conn,
			}

			// Add to active players
			activePlayersLock.Lock()
			activePlayers[player.ID] = player
			activePlayersLock.Unlock()

			// Set online status
			setPlayerOnlineStatus(player.ID, true)
			authenticated = true

			// Send welcome message
			conn.WriteJSON(WelcomeResponse{
				Type: ResponseWelcome,
				Player: PlayerResponse{
					ID:   player.ID,
					Name: player.Name,
				},
				Message: "Welcome back, " + player.Name + "!",
				Lobbies: getLobbiesList(),
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
func getLobbiesList() []LobbyResponse {
	lobbiesLock.Lock()
	defer lobbiesLock.Unlock()

	responseLobbies := make([]LobbyResponse, 0, len(lobbies))
	for _, l := range lobbies {
		responseLobbies = append(responseLobbies, LobbyResponse{
			ID:         l.ID,
			Name:       l.Name,
			MaxPlayers: l.MaxPlayers,
			IsPrivate:  l.IsPrivate,
			Players:    toPlayerResponses(l.Players),
			GameStart:  l.GameStart,
		})
	}
	return responseLobbies
}

// Helper function to convert Player to PlayerResponse
func toPlayerResponses(players []*Player) []PlayerResponse {
	res := make([]PlayerResponse, len(players))

	for i, p := range players {
		res[i] = PlayerResponse{
			ID:   p.ID,
			Name: p.Name,
		}
	}

	return res
}
