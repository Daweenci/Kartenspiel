package main

import "log"

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
