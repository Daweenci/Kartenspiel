package main

// Helper function to get lobbies list as DTO
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

// Helper function to convert Player to DTO PlayerResponse
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
