package main

import (
	"errors"
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
		sendResponse(player, LobbyJoinFailedResponse{
			BaseResponse: newBaseResponse(ResponseJoinLobbyFailed),
			Message:      "Incorrect password",
		})
		lobby.Lock.Unlock()
		return
	}

	// Check lobby capacity
	if len(lobby.Players) >= lobby.MaxPlayers {
		lobbyFullResponse := LobbyJoinFailedResponse{
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

func sendFriendRequestHandler(msg AddFriendRequest) {
	activePlayersLock.RLock()
	player, playerOK := activePlayers[msg.PlayerID]
	activePlayersLock.RUnlock()
	if !playerOK {
		log.Println("sendFriendRequestHandler: Player not found")
		disconnectPlayer(msg.PlayerID)
		return
	}

	friendID, friendOK := getPlayerIdByName(msg.FriendName)
	if !friendOK {
		sendFriendRequestResult(player, false, "Player not found")
		return
	}

	if msg.PlayerID == friendID {
		sendFriendRequestResult(player, false, "You cannot add yourself")
		return
	}

	if areFriends(msg.PlayerID, friendID) {
		sendFriendRequestResult(player, false, "You are already friends")
		return
	}

	err := createFriendRequest(msg.PlayerID, friendID)
	if err != nil {
		if errors.Is(err, ErrFriendRequestExists) {
			sendFriendRequestResult(player, false, "Friend request already sent")
			return
		}

		log.Printf("sendFriendRequestHandler: %v", err)
		sendFriendRequestResult(player, false, "Error creating friend request")
		return
	}

	sendFriendRequestResult(player, true, "Friend request sent")

	activePlayersLock.RLock()
	friend, ok := activePlayers[friendID]
	activePlayersLock.RUnlock()
	if !ok {
		log.Println("sendFriendRequestHandler: Friend not online")
		return
	}

	sendResponse(friend, PendingFriendRequestsResponse{
		BaseResponse:          newBaseResponse(ResponsePendingFriendRequests),
		PendingFriendRequests: getPendingFriendRequests(friendID),
	})
}

func sendFriendRequestResult(player *Player, success bool, message string) {
	sendResponse(player, FriendRequestSentResponse{
		BaseResponse: newBaseResponse(ResponseFriendRequestSent),
		Success:      success,
		Message:      message,
	})
}

func acceptFriendRequestHandler(msg AcceptFriendRequestRequest) {
	friendID := msg.FriendID
	playerID := msg.PlayerID
	acceptRequest := msg.AcceptRequest
	err := handleFriendRequest(friendID, playerID, acceptRequest)

	if err != nil {
		activePlayersLock.RLock()
		player, ok := activePlayers[playerID]
		activePlayersLock.RUnlock()

		if ok {
			sendErrorToPlayer(player, "Error handling friend request")
		}

		log.Printf("acceptFriendRequestHandler: %v", err)
		return
	}

	activePlayersLock.RLock()
	player, playerOk := activePlayers[msg.PlayerID]
	friend, friendOk := activePlayers[friendID]
	activePlayersLock.RUnlock()
	if !playerOk {
		log.Println("acceptFriendRequestHandler: Player not found")
		disconnectPlayer(msg.PlayerID)
		return
	}
	pendingFriendRequestsPlayer := getPendingFriendRequests(playerID)
	pendingFriendRequestsResponsePlayer := PendingFriendRequestsResponse{
		BaseResponse:          newBaseResponse(ResponsePendingFriendRequests),
		PendingFriendRequests: pendingFriendRequestsPlayer,
	}
	sendResponse(player, pendingFriendRequestsResponsePlayer)
	if acceptRequest {
		friendsListResponse := FriendsListResponse{
			BaseResponse: newBaseResponse(ResponseFriendsList),
			FriendsList:  getFriendsWithOnlineStatus(playerID),
		}
		sendResponse(player, friendsListResponse)
	}

	if !friendOk {
		log.Println("acceptFriendRequestHandler: Friend not online")
		return
	}
	pendingFriendRequestsFriend := getPendingFriendRequests(friendID)
	pendingFriendRequestsResponseFriend := PendingFriendRequestsResponse{
		BaseResponse:          newBaseResponse(ResponsePendingFriendRequests),
		PendingFriendRequests: pendingFriendRequestsFriend,
	}
	sendResponse(friend, pendingFriendRequestsResponseFriend)

	if acceptRequest && friendOk {
		playerInfo, err := getPlayerByID(playerID)
		if err != nil {
			log.Printf("acceptFriendRequestHandler: Error getting player info: %v", err)
			return
		}
		playerName := playerInfo.Username
		friendRequestAcceptedResponse := FriendRequestAcceptedResponse{
			BaseResponse: newBaseResponse(ResponseFriendRequestAccepted),
			Friend: FriendDTO{
				ID:       player.ID,
				Name:     playerName,
				IsOnline: true,
			},
		}
		sendResponse(friend, friendRequestAcceptedResponse)
	}
}

func pingAllFriendsHandler(playerID string) {
	player, err := getPlayerByID(playerID)
	if err != nil {
		log.Printf("pingAllFriendsHandler: Error getting player by ID: %v", err)
		return
	}
	friendsList := getFriendsWithOnlineStatus(playerID)
	friendsCameOnline := FriendCameOnlineResponse{
		BaseResponse: newBaseResponse(ResponseFriendCameOnline),
		Friend:       PlayerDTO{ID: player.ID, Name: player.Username},
	}

	for _, friend := range friendsList {
		activePlayersLock.RLock()
		friend, ok := activePlayers[friend.ID]
		activePlayersLock.RUnlock()
		if ok {
			sendResponse(friend, friendsCameOnline)
		}
	}
}

func getFriendsWithOnlineStatus(playerID string) []FriendDTO {
	friendsList := getFriendsList(playerID)
	friendsListWithOnlineStatus := make([]FriendDTO, len(friendsList))
	for i, f := range friendsList {
		activePlayersLock.RLock()
		_, isOnline := activePlayers[f.ID]
		activePlayersLock.RUnlock()
		friendsListWithOnlineStatus[i] = FriendDTO{
			ID:       f.ID,
			Name:     f.Name,
			IsOnline: isOnline,
		}
	}
	return friendsListWithOnlineStatus
}
