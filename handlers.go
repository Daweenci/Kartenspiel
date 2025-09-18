package main

import (
	"log"

	"github.com/google/uuid"
)

func joinLobbyHandler(msg JoinLobbyRequest) {
	lobbiesLock.Lock()
	defer lobbiesLock.Unlock()

	lobby, ok := lobbies[msg.LobbyID]
	if !ok {
		log.Println("joinLobbyHandler: Lobby not found")
		// Lobby not found, silently ignore or send a success response if needed
		return
	}

	player, ok := players[msg.PlayerID]
	if !ok {
		log.Println("joinLobbyHandler: Player not found")
		// Player not found, silently ignore or send a success response if needed
		return
	}

	// Check if player is already in the lobby
	for _, p := range lobby.Players {
		if p.ID == msg.PlayerID {
			// Already in the lobby, silently ignore or send a success response if needed
			return
		}
	}

	// Check lobby capacity
	if len(lobby.Players) >= lobby.MaxPlayers {
		player.Conn.WriteJSON(map[string]interface{}{
			"type":    ResponseJoinLobbyUnsuccessful,
			"message": "Lobby is full",
		})
		return
	}

	// Check password
	if lobby.Password != msg.Password {
		err := player.Conn.WriteJSON(map[string]interface{}{
			"type":    ResponseJoinLobbyUnsuccessful,
			"message": "Incorrect password",
		})
		if err != nil {
			log.Println("Error sending join failure response:", err)
		}
		return
	}

	// Add player and respond
	lobby.Players = append(lobby.Players, player)
	err := player.Conn.WriteJSON(map[string]interface{}{
		"type":  ResponseJoinLobbySuccess,
		"lobby": lobby,
	})
	if err != nil {
		log.Println("Error sending Lobby join success:", err)
	}
	broadcastLobbyUpdate(lobby)
	broadcastLobbies()
}

func leaveLobbyHandler(msg LeaveLobbyRequest) {
	lobbiesLock.Lock()
	defer lobbiesLock.Unlock()

	if lobby, ok := lobbies[msg.LobbyID]; ok {
		for i := len(lobby.GameStart) - 1; i >= 0; i-- {
			if lobby.GameStart[i].ID == msg.PlayerID {
				lobby.GameStart = append(lobby.GameStart[:i], lobby.GameStart[i+1:]...)
			}
		}

		for i := len(lobby.Players) - 1; i >= 0; i-- {
			if lobby.Players[i].ID == msg.PlayerID {
				lobby.Players = append(lobby.Players[:i], lobby.Players[i+1:]...)
				break
			}
		}

		if len(lobby.Players) == 0 {
			delete(lobbies, lobby.ID)
		} else {
			broadcastLobbyUpdate(lobby)
		}
		err := players[msg.PlayerID].Conn.WriteJSON(map[string]interface{}{
			"type": ResponseLobbyLeft,
		})
		if err != nil {
			log.Println("Error sending LobbyID:", err)
			return
		}
		broadcastLobbies()
	}
}

func createLobbyHandler(msg CreateLobbyRequest) {
	lobbiesLock.Lock()
	defer lobbiesLock.Unlock()

	player := players[msg.PlayerID]
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

	err := player.Conn.WriteJSON(map[string]interface{}{
		"type":  ResponseLobbyCreated,
		"lobby": newLobby,
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
	player := players[msg.PlayerID]

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
	updatedLobby := Lobby{
		ID:         lobby.ID,
		Name:       lobby.Name,
		MaxPlayers: lobby.MaxPlayers,
		IsPrivate:  lobby.IsPrivate,
		Password:   lobby.Password,
		Players:    lobby.Players,
		GameStart:  lobby.GameStart,
	}

	playersLock.Lock()
	for _, player := range lobby.Players {
		err := player.Conn.WriteJSON(map[string]interface{}{
			"type":  ResponseLobbyUpdated,
			"lobby": updatedLobby,
		})
		if err != nil {
			log.Printf("Error sending lobby update to player %s: %v", player.ID, err)
			player.Conn.Close()
			delete(players, player.ID)
		}
	}
	playersLock.Unlock()
}

func broadcastLobbies() {
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

	playersLock.Lock()
	for id, player := range players {
		err := player.Conn.WriteJSON(map[string]interface{}{
			"type":    ResponseLobbyList,
			"lobbies": responseLobbies,
		})
		if err != nil {
			log.Println("Error broadcasting:", err)
			player.Conn.Close()
			delete(players, id)
		}
	}
	playersLock.Unlock()
}
