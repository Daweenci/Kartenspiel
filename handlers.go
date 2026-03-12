package main

import (
	"log"

	"github.com/google/uuid"
)

func joinLobbyHandler(msg JoinLobbyRequest) {

	// Check if lobby exists
	lobbiesLock.RLock()
	lobby, ok := lobbies[msg.LobbyID]
	lobbiesLock.RUnlock()
	if !ok {
		log.Println("joinLobbyHandler: Lobby not found")
		// Lobby not found, silently ignore or send a response if needed
		return
	}

	// Check if player is connected
	activeConnectionsLock.RLock()
	player, ok := activeConnections[msg.PlayerID]
	activeConnectionsLock.RUnlock()
	if !ok {
		log.Println("joinLobbyHandler: Player not found")
		// Player not found, silently ignore or send a response if needed
		return
	}

	lobby.Lock.Lock()
	// Check if player is already in the lobby
	for _, p := range lobby.Players {
		if p.ID == msg.PlayerID {
			// Already in the lobby, silently ignore or send a response if needed
			lobby.Lock.Unlock()
			return
		}
	}

	// Check password
	if lobby.Password != msg.Password {
		err := player.Conn.WriteJSON(map[string]interface{}{
			"type":    ResponseJoinLobbyFailed,
			"message": "Incorrect password",
		})
		if err != nil {
			log.Println("Error sending join failure response:", err)
		}
		lobby.Lock.Unlock()
		return
	}

	// Check lobby capacity
	if len(lobby.Players) >= lobby.MaxPlayers {
		player.Conn.WriteJSON(map[string]interface{}{
			"type":    ResponseJoinLobbyFailed,
			"message": "Lobby is full",
		})
		lobby.Lock.Unlock()
		return
	}

	// Add player and respond
	lobby.Players = append(lobby.Players, player)
	lobby.Lock.Unlock()
	lobbyResponse := LobbyResponse{
		ID:         lobby.ID,
		Name:       lobby.Name,
		MaxPlayers: lobby.MaxPlayers,
		IsPrivate:  lobby.IsPrivate,
		Players:    lobby.Players,
		GameStart:  lobby.GameStart,
	}
	err := player.Conn.WriteJSON(map[string]interface{}{
		"type":  ResponseJoinLobbySuccessful,
		"lobby": lobbyResponse,
	})
	if err != nil {
		log.Println("Error sending Lobby join success:", err)
	}
	broadcastLobbyUpdate(lobby)
	broadcastLobbies()
}

func leaveLobbyHandler(msg LeaveLobbyRequest) {
	lobbiesLock.Lock()

	lobby, ok := lobbies[msg.LobbyID]

	if !ok {
		log.Println("leaveLobbyHandler: Lobby not found")
		lobbiesLock.Unlock()
		return
	}

	for i := len(lobby.GameStart) - 1; i >= 0; i-- {
		if lobby.GameStart[i].ID == msg.PlayerID {
			lobby.GameStart = append(lobby.GameStart[:i], lobby.GameStart[i+1:]...)
			break
		}
	}

	for i := len(lobby.Players) - 1; i >= 0; i-- {
		if lobby.Players[i].ID == msg.PlayerID {
			lobby.Players = append(lobby.Players[:i], lobby.Players[i+1:]...)
			break
		}
	}

	lobbyDeleted := false
	if len(lobby.Players) == 0 {
		delete(lobbies, lobby.ID)
		lobbyDeleted = true
	}
	lobbiesLock.Unlock()
	if !lobbyDeleted {
		broadcastLobbyUpdate(lobby)
	}

	err := activeConnections[msg.PlayerID].Conn.WriteJSON(map[string]interface{}{
		"type": ResponseLobbyLeft,
	})
	if err != nil {
		log.Println("Error leaving Lobby:", err)
		return
	}
	broadcastLobbies()
}

func createLobbyHandler(msg CreateLobbyRequest) {
	lobbiesLock.Lock()
	defer lobbiesLock.Unlock()

	player := activeConnections[msg.PlayerID]
	lobbyID := uuid.New().String()

	newLobby := &Lobby{
		ID:         lobbyID,
		Name:       msg.LobbyName,
		MaxPlayers: msg.MaxPlayers,
		IsPrivate:  msg.IsPrivate,
		Password:   msg.Password,
		Players:    []*Player{player},
		GameStart:  []PlayerStarted{},
	}

	lobbies[lobbyID] = newLobby

	newLobbyResponse := LobbyResponse{
		ID:         newLobby.ID,
		Name:       newLobby.Name,
		MaxPlayers: newLobby.MaxPlayers,
		IsPrivate:  newLobby.IsPrivate,
		Players:    newLobby.Players,
		GameStart:  newLobby.GameStart,
	}

	err := player.Conn.WriteJSON(map[string]interface{}{
		"type":  ResponseLobbyCreated,
		"lobby": newLobbyResponse,
	})
	if err != nil {
		log.Println("Error sending LobbyID:", err)
		return
	}
	broadcastLobbies()
}

func startGameHandler(msg StartGame) {
	lobbiesLock.Lock()
	defer lobbiesLock.Unlock()

	lobby := lobbies[msg.LobbyID]
	player := activeConnections[msg.PlayerID]

	lobby.GameStart = append(lobby.GameStart, PlayerStarted{
		ID: player.ID,
	})
	broadcastLobbyUpdate(lobby)
}

func cancelGameHandler(msg CancelGame) {
	lobbiesLock.Lock()
	defer lobbiesLock.Unlock()

	lobby, ok := lobbies[msg.LobbyID]
	if !ok {
		log.Println("cancelGameHandler: Lobby not found")
		return
	}

	for i, p := range lobby.GameStart {
		if p.ID == msg.PlayerID {
			lobby.GameStart = append(lobby.GameStart[:i], lobby.GameStart[i+1:]...)
			break
		}
	}
	broadcastLobbyUpdate(lobby)
}

func broadcastLobbyUpdate(lobby *Lobby) {
	lobby.Lock.RLock()
	players := append([]*Player(nil), lobby.Players...)
	updatedLobby := LobbyResponse{
		ID:         lobby.ID,
		Name:       lobby.Name,
		MaxPlayers: lobby.MaxPlayers,
		IsPrivate:  lobby.IsPrivate,
		Players:    players,
		GameStart:  lobby.GameStart,
	}
	lobby.Lock.RUnlock()

	activeConnectionsLock.Lock()
	for _, player := range players {
		err := player.Conn.WriteJSON(map[string]interface{}{
			"type":  ResponseLobbyUpdated,
			"lobby": updatedLobby,
		})
		if err != nil {
			log.Printf("Error sending lobby update to player %s: %v", player.ID, err)
			player.Conn.Close()
			delete(activeConnections, player.ID)
		}
	}
	activeConnectionsLock.Unlock()
}

func broadcastLobbies() {
	lobbiesLock.RLock()
	lobbiesResponse := make([]LobbyResponse, 0, len(lobbies))
	for _, lobby := range lobbies {
		lobby.Lock.RLock()
		playersCopy := append([]*Player(nil), lobby.Players...)
		lobbiesResponse = append(lobbiesResponse, LobbyResponse{
			ID:         lobby.ID,
			Name:       lobby.Name,
			MaxPlayers: lobby.MaxPlayers,
			IsPrivate:  lobby.IsPrivate,
			Players:    playersCopy,
			GameStart:  lobby.GameStart,
		})
		lobby.Lock.RUnlock()
	}
	lobbiesLock.RUnlock()

	activeConnectionsLock.Lock()
	for id, player := range activeConnections {
		err := player.Conn.WriteJSON(map[string]interface{}{
			"type":    ResponseLobbyList,
			"lobbies": lobbiesResponse,
		})
		if err != nil {
			log.Println("Error broadcasting:", err)
			player.Conn.Close()
			delete(activeConnections, id)
		}
	}
	activeConnectionsLock.Unlock()
}
