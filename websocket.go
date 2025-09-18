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
	lobbies     = make(map[string]*Lobby)
	lobbiesLock sync.Mutex

	players     = make(map[string]*Player)
	playersLock sync.Mutex

	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}

	jwtSecret = []byte("YOUR_SECRET_KEY")
)

// JWT helpers
func generateJWT(playerID string) (string, error) {
	claims := jwt.MapClaims{
		"playerID": playerID,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func parseJWT(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return "", fmt.Errorf("invalid token: %v", err)
	}

	claims := token.Claims.(jwt.MapClaims)
	playerID, ok := claims["playerID"].(string)
	if !ok {
		return "", fmt.Errorf("invalid token claims")
	}
	return playerID, nil
}

// Send error to player helper
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

	// Add player to global map
	playersLock.Lock()
	players[player.ID] = player
	playersLock.Unlock()

	// Cleanup on disconnect
	defer func() {
		conn.Close()
		log.Println("Connection closed for player:", player.ID)

		playersLock.Lock()
		defer playersLock.Unlock()

		// Remove player from lobbies
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

		delete(players, player.ID)
		broadcastLobbies()
	}()

	// Send initial lobby list
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

		// Base message to read type and token
		var base struct {
			Type  MessageType `json:"type"`
			Token string      `json:"token"`
		}
		if err := json.Unmarshal(msgBytes, &base); err != nil {
			log.Println("Invalid message format")
			sendErrorToPlayer(player, "Invalid message format")
			continue
		}

		// Require JWT for all actions except login/register
		if base.Type != RequestLogin && base.Type != RequestRegister {
			playerID, err := parseJWT(base.Token)
			if err != nil {
				sendErrorToPlayer(player, "Invalid or expired token")
				continue
			}
			// ensure player exists
			playersLock.Lock()
			authPlayer, ok := players[playerID]
			playersLock.Unlock()
			if !ok {
				sendErrorToPlayer(player, "Player not found")
				continue
			}
			player = authPlayer
		}

		switch base.Type {
		case RequestLogin:
			var msg LoginRequest
			if err := json.Unmarshal(msgBytes, &msg); err != nil {
				sendErrorToPlayer(player, "Invalid login message")
				continue
			}
			if !checkCredentials(msg.Name, msg.Password) {
				conn.WriteJSON(map[string]interface{}{
					"type":    ResponseLoginUnsuccessful,
					"message": "Invalid username or password",
				})
				continue
			}
			token, err := generateJWT(player.ID)
			if err != nil {
				sendErrorToPlayer(player, "Failed to generate token")
				continue
			}
			conn.WriteJSON(map[string]interface{}{
				"type":  ResponseLoginSuccess,
				"token": token,
			})

		case RequestRegister:
			var msg RegisterRequest
			if err := json.Unmarshal(msgBytes, &msg); err != nil {
				sendErrorToPlayer(player, "Invalid register message")
				continue
			}
			if !registerPlayer(msg.Name, msg.Password) {
				conn.WriteJSON(map[string]interface{}{
					"type":    ResponseRegisterFailed,
					"message": "Username already exists",
				})
				continue
			}
			token, err := generateJWT(player.ID)
			if err != nil {
				sendErrorToPlayer(player, "Failed to generate token")
				continue
			}
			conn.WriteJSON(map[string]interface{}{
				"type":  ResponseRegisterSuccess,
				"token": token,
			})

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

func registerPlayer(name string, password string) bool {
	panic("unimplemented")
}

func checkCredentials(name, password string) bool {
	panic("unimplemented")
}
