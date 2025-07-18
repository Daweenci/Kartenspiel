package main

import (
	"log"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

func joinLobbyHandler(msg JoinLobbyRequest) {
	lobbiesLock.Lock()
	defer lobbiesLock.Unlock()

	for i, lobby := range lobbies {
		if lobby.ID == msg.LobbyID {
			player := Player{
				ID:   msg.ID,
				Name: msg.Name,
			}
			lobbies[i].Players = append(lobbies[i].Players, player)
			broadcastLobbies()
			break
		}
	}
}

func leaveLobbyHandler(msg LeaveLobbyRequest) {
	lobbiesLock.Lock()
	defer lobbiesLock.Unlock()

	for i, lobby := range lobbies {
		if lobby.ID == msg.LobbyID {
			for j, player := range lobby.Players {
				if player.ID == msg.ID {
					lobbies[i].Players = append(lobbies[i].Players[:j], lobbies[i].Players[j+1:]...)
					break
				}
			}
			broadcastLobbies()
			break
		}
	}
}

func createLobbyHandler(msg CreateLobbyRequest, conn *websocket.Conn) {
	lobbiesLock.Lock()
	defer lobbiesLock.Unlock()

	newLobby := Lobby{
		ID:         uuid.New().String(),
		Name:       msg.LobbyName,
		MaxPlayers: msg.MaxPlayers,
		IsPrivate:  msg.IsPrivate,
		Password:   msg.Password,
		Players: []Player{
			{
				Name: msg.PlayerName,
				ID:   msg.PlayerID,
			},
		},
	}
	lobbies = append(lobbies, newLobby)
	broadcastLobbies()

	var err = conn.WriteJSON(map[string]interface{}{
		"type":  ResponseLobbyCreated,
		"lobby": newLobby,
	})
	if err != nil {
		log.Println("Error sending LobbyID:", err)
		return
	}
}

func broadcastLobbies() {
	clientsLock.Lock()
	defer clientsLock.Unlock()
	for client := range clients {
		err := client.WriteJSON(map[string]interface{}{
			"type":    ResponseLobbyList,
			"lobbies": lobbies,
		})
		if err != nil {
			client.Close()
			delete(clients, client)
			log.Println("broadcast Lobbies error:", err)
		}
		log.Println("Broadcasted lobby list to client")
	}
}
