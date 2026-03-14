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
		sendResponse(player, IncorrectLobbyPasswordResponse{
			BaseResponse: newBaseResponse(ResponseJoinLobbyFailed),
			Message:      "Incorrect password",
		})
		lobby.Lock.Unlock()
		return
	}

	// Check lobby capacity
	if len(lobby.Players) >= lobby.MaxPlayers {
		lobbyFullResponse := LobbyFullResponse{
			BaseResponse: newBaseResponse(ResponseJoinLobbyFailed),
			Message:      "Lobby is full",
		}
		sendResponse(player, lobbyFullResponse)
		lobby.Lock.Unlock()
		return
	}

	// Add player and respond
	lobby.Players = append(lobby.Players, player)
	lobbyResponse := LobbyDTO{
		ID:         lobby.ID,
		Name:       lobby.Name,
		MaxPlayers: lobby.MaxPlayers,
		IsPrivate:  lobby.IsPrivate,
		Players:    toPlayerResponses(lobby.Players),
		GameStart:  lobby.GameStart,
	}
	lobby.Lock.Unlock()
	successfulJoinResponse := SuccessfulJoinLobbyResponse{
		BaseResponse: newBaseResponse(ResponseJoinLobbySuccessful),
		Lobby:        lobbyResponse,
	}
	sendResponse(player, successfulJoinResponse)
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

	activePlayersLock.RLock()
	player, ok := activePlayers[msg.PlayerID]
	activePlayersLock.RUnlock()
	if !ok {
		log.Println("leaveLobbyHandler: Player not found")
		disconnectPlayer(player.ID)
		return
	}

	lobby.Lock.Lock()
	for i := len(lobby.GameStart) - 1; i >= 0; i-- {
		if lobby.GameStart[i].ID == player.ID {
			lobby.GameStart = append(lobby.GameStart[:i], lobby.GameStart[i+1:]...)
			break
		}
	}

	for i := len(lobby.Players) - 1; i >= 0; i-- {
		if lobby.Players[i].ID == player.ID {
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

	sendResponse(player, LobbyLeftResponse{
		BaseResponse: newBaseResponse(ResponseLobbyLeft),
	})
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

	newLobbyResponse := LobbyDTO{
		ID:         newLobby.ID,
		Name:       newLobby.Name,
		MaxPlayers: newLobby.MaxPlayers,
		IsPrivate:  newLobby.IsPrivate,
		Players:    toPlayerResponses(newLobby.Players),
		GameStart:  newLobby.GameStart,
	}

	createLobbyResponse := CreateLobbyResponse{
		BaseResponse: newBaseResponse(ResponseLobbyCreated),
		Lobby:        newLobbyResponse,
	}
	sendResponse(player, createLobbyResponse)
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

	activePlayersLock.RLock()
	player, ok := activePlayers[msg.PlayerID]
	activePlayersLock.RUnlock()
	if !ok {
		log.Println("cancelGameHandler: Player not found")
		disconnectPlayer(msg.PlayerID)
		return
	}

	lobby.Lock.Lock()
	for i := len(lobby.GameStart) - 1; i >= 0; i-- {
		if lobby.GameStart[i].ID == player.ID {
			lobby.GameStart = append(lobby.GameStart[:i], lobby.GameStart[i+1:]...)
			break
		}
	}
	lobby.Lock.Unlock()

	broadcastLobbyUpdate(lobby)
}
