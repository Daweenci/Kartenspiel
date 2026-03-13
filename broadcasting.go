package main

func broadcastLobbyUpdate(lobby *Lobby) {
	lobby.Lock.RLock()
	// Copying mutable fields, leaving immutable ones out
	playersCopy := make([]*Player, len(lobby.Players))
	copy(playersCopy, lobby.Players)
	gameStartCopy := make([]PlayerStarted, len(lobby.GameStart))
	copy(gameStartCopy, lobby.GameStart)
	lobby.Lock.RUnlock()
	updatedLobby := LobbyDTO{
		ID:         lobby.ID,
		Name:       lobby.Name,
		MaxPlayers: lobby.MaxPlayers,
		IsPrivate:  lobby.IsPrivate,
		Players:    toPlayerResponses(playersCopy),
		GameStart:  gameStartCopy,
	}

	for _, player := range playersCopy {
		lobbyUpdatedResponse := LobbyUpdatedResponse{
			BaseResponse: newBaseResponse(ResponseLobbyUpdated),
			Lobby:        updatedLobby,
		}
		sendResponse(player, lobbyUpdatedResponse)
	}
}

func broadcastLobbies() {
	lobbiesLock.RLock()
	lobbiesCopy := make([]*Lobby, 0, len(lobbies))
	for _, l := range lobbies {
		lobbiesCopy = append(lobbiesCopy, l)
	}
	lobbiesLock.RUnlock()

	lobbiesResponse := make([]LobbyDTO, 0, len(lobbiesCopy))
	for _, lobby := range lobbiesCopy {
		lobby.Lock.RLock()
		playersCopy := make([]*Player, len(lobby.Players))
		copy(playersCopy, lobby.Players)
		gameStartCopy := make([]PlayerStarted, len(lobby.GameStart))
		copy(gameStartCopy, lobby.GameStart)
		lobbiesResponse = append(lobbiesResponse, LobbyDTO{
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
		lobbiesUpdateResponse := LobbiesUpdateResponse{
			BaseResponse: newBaseResponse(ResponseLobbyList),
			Lobbies:      lobbiesResponse,
		}
		sendResponse(player, lobbiesUpdateResponse)
	}
}
