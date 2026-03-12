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
	activePlayersLock.RLock()
	player, ok := activePlayers[msg.PlayerID]
	activePlayersLock.RUnlock()
	if !ok {
		log.Println("joinLobbyHandler: Player not found")
		disconnectPlayer(msg.PlayerID)
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
		err := player.Conn.WriteJSON(IncorrectLobbyPasswordResponse{
			Type:    ResponseJoinLobbyFailed,
			Message: "Incorrect password",
		})
		if err != nil {
			log.Println("Error sending join failure response:", err)
		}
		lobby.Lock.Unlock()
		return
	}

	// Check lobby capacity
	if len(lobby.Players) >= lobby.MaxPlayers {
		player.Conn.WriteJSON(LobbyFullResponse{
			Type:    ResponseJoinLobbyFailed,
			Message: "Lobby is full",
		})
		lobby.Lock.Unlock()
		return
	}

	// Add player and respond
	lobby.Players = append(lobby.Players, player)
	lobbyResponse := LobbyResponse{
		ID:         lobby.ID,
		Name:       lobby.Name,
		MaxPlayers: lobby.MaxPlayers,
		IsPrivate:  lobby.IsPrivate,
		Players:    toPlayerResponses(lobby.Players),
		GameStart:  lobby.GameStart,
	}
	lobby.Lock.Unlock()
	err := player.Conn.WriteJSON(SuccessfulJoinLobbyResponse{
		Type:  ResponseJoinLobbySuccessful,
		Lobby: lobbyResponse,
	})
	if err != nil {
		log.Println("Error sending Lobby join success:", err)
		return
	}
	broadcastLobbyUpdate(lobby)
	broadcastLobbies()
}

func leaveLobbyHandler(msg LeaveLobbyRequest) {
	lobbiesLock.RLock()
	lobby, ok := lobbies[msg.LobbyID]
	lobbiesLock.RUnlock()
	if !ok {
		log.Println("leaveLobbyHandler: Lobby not found")
		return
	}

	lobby.Lock.Lock()
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
	lobby.Lock.Unlock()

	lobbyDeleted := false
	if len(lobby.Players) == 0 {
		lobbiesLock.Lock()
		delete(lobbies, lobby.ID)
		lobbiesLock.Unlock()
		lobbyDeleted = true
	}
	if !lobbyDeleted {
		broadcastLobbyUpdate(lobby)
	}

	err := activePlayers[msg.PlayerID].Conn.WriteJSON(LobbyLeftResponse{
		Type: ResponseLobbyLeft,
	})
	if err != nil {
		log.Println("Error leaving Lobby:", err)
		return
	}
	broadcastLobbies()
}

func createLobbyHandler(msg CreateLobbyRequest) {
	activePlayersLock.RLock()
	player, ok := activePlayers[msg.PlayerID]
	activePlayersLock.RUnlock()
	if !ok {
		log.Println("createLobbyHandler: Player not found")
		disconnectPlayer(msg.PlayerID)
		return
	}
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

	lobbiesLock.Lock()
	lobbies[lobbyID] = newLobby
	lobbiesLock.Unlock()

	newLobbyResponse := LobbyResponse{
		ID:         newLobby.ID,
		Name:       newLobby.Name,
		MaxPlayers: newLobby.MaxPlayers,
		IsPrivate:  newLobby.IsPrivate,
		Players:    toPlayerResponses(newLobby.Players),
		GameStart:  newLobby.GameStart,
	}

	err := player.Conn.WriteJSON(CreateLobbyResponse{
		Type:  ResponseLobbyCreated,
		Lobby: newLobbyResponse,
	})
	if err != nil {
		log.Println("Error sending LobbyID:", err)
		return
	}
	broadcastLobbies()
}

func startGameHandler(msg StartGame) {
	lobbiesLock.RLock()
	lobby, ok := lobbies[msg.LobbyID]
	lobbiesLock.RUnlock()
	if !ok {
		log.Println("StartGameHandler: Lobby not found")
		return
	}

	activePlayersLock.RLock()
	player, ok := activePlayers[msg.PlayerID]
	activePlayersLock.RUnlock()
	if !ok {
		log.Println("StartGameHandler: Player not found")
		disconnectPlayer(msg.PlayerID)
		return
	}

	lobby.Lock.Lock()
	alreadyStarted := false
	for _, p := range lobby.GameStart {
		if p.ID == player.ID {
			alreadyStarted = true
			break
		}
	}
	if !alreadyStarted {
		lobby.GameStart = append(lobby.GameStart, PlayerStarted{ID: player.ID})
	}
	lobby.Lock.Unlock()

	broadcastLobbyUpdate(lobby)
}

func cancelGameHandler(msg CancelGame) {
	lobbiesLock.RLock()
	lobby, ok := lobbies[msg.LobbyID]
	lobbiesLock.RUnlock()
	if !ok {
		log.Println("cancelGameHandler: Lobby not found")
		return
	}

	lobby.Lock.Lock()
	for i, p := range lobby.GameStart {
		if p.ID == msg.PlayerID {
			lobby.GameStart = append(lobby.GameStart[:i], lobby.GameStart[i+1:]...)
			break
		}
	}
	lobby.Lock.Unlock()
	broadcastLobbyUpdate(lobby)
}

func broadcastLobbyUpdate(lobby *Lobby) {
	lobby.Lock.RLock()
	// Copying mutable fields, leaving immutable ones out
	playersCopy := append([]*Player(nil), lobby.Players...)
	gameStartCopy := append([]PlayerStarted(nil), lobby.GameStart...)
	lobby.Lock.RUnlock()
	updatedLobby := LobbyResponse{
		ID:         lobby.ID,
		Name:       lobby.Name,
		MaxPlayers: lobby.MaxPlayers,
		IsPrivate:  lobby.IsPrivate,
		Players:    toPlayerResponses(playersCopy),
		GameStart:  gameStartCopy,
	}

	for _, player := range playersCopy {
		err := player.Conn.WriteJSON(LobbyUpdatedResponse{
			Type:  ResponseLobbyUpdated,
			Lobby: updatedLobby,
		})
		if err != nil {
			log.Printf("Error sending lobby update to player %s: %v", player.ID, err)
			disconnectPlayer(player.ID)
		}
	}
}

func broadcastLobbies() {
	lobbiesLock.RLock()
	lobbiesCopy := make([]*Lobby, 0, len(lobbies))
	for _, l := range lobbies {
		lobbiesCopy = append(lobbiesCopy, l)
	}
	lobbiesLock.RUnlock()

	lobbiesResponse := make([]LobbyResponse, 0, len(lobbiesCopy))
	for _, lobby := range lobbiesCopy {
		lobby.Lock.RLock()
		playersCopy := append([]*Player(nil), lobby.Players...)
		gameStartCopy := append([]PlayerStarted(nil), lobby.GameStart...)
		lobbiesResponse = append(lobbiesResponse, LobbyResponse{
			ID:         lobby.ID,
			Name:       lobby.Name,
			MaxPlayers: lobby.MaxPlayers,
			IsPrivate:  lobby.IsPrivate,
			Players:    toPlayerResponses(playersCopy),
			GameStart:  gameStartCopy,
		})
		lobby.Lock.RUnlock()
	}

	activePlayersLock.RLock()
	activePlayersCopy := make([]*Player, 0, len(activePlayers))
	for _, p := range activePlayers {
		activePlayersCopy = append(activePlayersCopy, p)
	}
	activePlayersLock.RUnlock()
	for _, player := range activePlayersCopy {
		err := player.Conn.WriteJSON(LobbiesUpdateResponse{
			Type:    ResponseLobbyList,
			Lobbies: lobbiesResponse,
		})
		if err != nil {
			log.Println("Error broadcasting:", err)
			disconnectPlayer(player.ID)
		}
	}
}

func disconnectPlayer(playerID string) {
	activePlayersLock.Lock()
	defer activePlayersLock.Unlock()

	player, ok := activePlayers[playerID]
	if !ok {
		return
	}

	player.Conn.Close()
	delete(activePlayers, playerID)
}
