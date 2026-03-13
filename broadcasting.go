package main

import "log"

func broadcastLobbyUpdate(lobby *Lobby) {
	lobby.Lock.RLock()
	// Copying mutable fields, leaving immutable ones out
	playersCopy := append([]*Player(nil), lobby.Players...)
	gameStartCopy := append([]PlayerStarted(nil), lobby.GameStart...)
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
		err := player.Conn.WriteJSON(LobbyUpdatedResponse{
			BaseResponse: newBaseResponse(ResponseLobbyUpdated),
			Lobby:        updatedLobby,
		})
		if err != nil {
			log.Printf("Error broadcasting lobby update to player %s: %v", player.ID, err)
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

	lobbiesResponse := make([]LobbyDTO, 0, len(lobbiesCopy))
	for _, lobby := range lobbiesCopy {
		lobby.Lock.RLock()
		playersCopy := append([]*Player(nil), lobby.Players...)
		gameStartCopy := append([]PlayerStarted(nil), lobby.GameStart...)
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
		err := player.Conn.WriteJSON(LobbiesUpdateResponse{
			BaseResponse: newBaseResponse(ResponseLobbyList),
			Lobbies:      lobbiesResponse,
		})
		if err != nil {
			log.Println("Error broadcasting lobbies:", err)
			disconnectPlayer(player.ID)
		}
	}
}
