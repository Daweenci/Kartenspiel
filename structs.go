package main

import "github.com/gorilla/websocket"

type MessageType string

const (
	RequestCreateLobby MessageType = "create_lobby"
	RequestJoinLobby   MessageType = "join_lobby"
	RequestLeaveLobby  MessageType = "leave_lobby"
	RequestStartGame   MessageType = "start_game"
	RequestCancelGame  MessageType = "cancel_game"

	ResponseWelcome                MessageType = "welcome"
	ResponseLobbyCreated           MessageType = "lobby_created"
	ResponseLobbyList              MessageType = "lobby_list"
	ResponseLobbyUpdated           MessageType = "lobby_updated"
	ResponseJoinLobbySuccess       MessageType = "join_lobby_success"
	ResponseJoinLobbyWrongPassword MessageType = "join_lobby_wrong_password"
	ResponseJoinLobbyFull          MessageType = "join_lobby_full"
	ResponseLobbyLeft              MessageType = "lobby_left"
	ResponseError                  MessageType = "error"
)

type Player struct {
	ID   string          `json:"id"`
	Name string          `json:"name"`
	Conn *websocket.Conn `json:"-"`
}

type JoinLobbyRequest struct {
	Type     MessageType `json:"type"`
	LobbyID  string      `json:"lobbyID"`
	PlayerID string      `json:"playerID"`
	Name     string      `json:"name"`
	Password string      `json:"password"`
}

type CreateLobbyRequest struct {
	Type       MessageType `json:"type"`
	LobbyName  string      `json:"lobbyName"`
	MaxPlayers int         `json:"maxPlayers"`
	IsPrivate  bool        `json:"isPrivate"`
	Password   string      `json:"password"`
	PlayerID   string      `json:"playerID"`
	PlayerName string      `json:"playerName"`
}

type LeaveLobbyRequest struct {
	Type     MessageType `json:"type"`
	LobbyID  string      `json:"lobbyID"`
	PlayerID string      `json:"playerID"`
}

type StartGame struct {
	Type     MessageType `json:"type"`
	LobbyID  string      `json:"lobbyID"`
	PlayerID string      `json:"playerID"`
}

type CancelGame struct {
	Type     MessageType `json:"type"`
	LobbyID  string      `json:"lobbyID"`
	PlayerID string      `json:"playerID"`
}

type BroadcastedLobby struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	MaxPlayers int       `json:"maxPlayers"`
	IsPrivate  bool      `json:"isPrivate"`
	Players    []*Player `json:"players"`
}

type Lobby struct {
	ID         string          `json:"id"`
	Name       string          `json:"name"`
	MaxPlayers int             `json:"maxPlayers"`
	IsPrivate  bool            `json:"isPrivate"`
	Password   string          `json:"password"`
	Players    []*Player       `json:"players"`
	GameStart  []PlayerStarted `json:"gameStart"`
}

type PlayerStarted struct {
	ID string `json:"id"`
}
