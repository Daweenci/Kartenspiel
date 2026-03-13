package main

import (
	"encoding/json"
	"log"

	"github.com/gorilla/websocket"
)

// Helper function to get lobbies list as DTO
func getLobbiesList() []LobbyDTO {
	lobbiesLock.Lock()
	defer lobbiesLock.Unlock()

	responseLobbies := make([]LobbyDTO, 0, len(lobbies))
	for _, l := range lobbies {
		responseLobbies = append(responseLobbies, LobbyDTO{
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
func toPlayerResponses(players []*Player) []PlayerDTO {
	res := make([]PlayerDTO, len(players))

	for i, p := range players {
		res[i] = PlayerDTO{
			ID:   p.ID,
			Name: p.Name,
		}
	}

	return res
}

func disconnectPlayer(playerID string) {
	activePlayersLock.Lock()
	player, ok := activePlayers[playerID]
	if !ok {
		activePlayersLock.Unlock()
		return
	}

	delete(activePlayers, playerID)
	activePlayersLock.Unlock()

	close(player.Send)
	player.Conn.Close()
}

func (p *Player) writePump() {
	defer p.Conn.Close()

	for msg := range p.Send {
		err := p.Conn.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			log.Println("write error:", err)
			return
		}
	}
}

func sendResponse(p *Player, r Response) {
	msg, err := json.Marshal(r)

	if err != nil {
		log.Printf("MessageType:%v. Marshal error: %v", r.GetType(), err)
		return
	}

	select {
	case p.Send <- msg:
	default:
		log.Printf("Dropping message for %s (send buffer full)", p.ID)
	}
}
