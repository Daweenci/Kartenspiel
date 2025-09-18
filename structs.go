package main

import "github.com/gorilla/websocket"

type MessageType string

const (
	RequestLogin       MessageType = "login"
	RequestRegister    MessageType = "register"
	RequestCreateLobby MessageType = "create_lobby"
	RequestJoinLobby   MessageType = "join_lobby"
	RequestLeaveLobby  MessageType = "leave_lobby"
	RequestStartGame   MessageType = "start_game"
	RequestCancelGame  MessageType = "cancel_game"

	ResponseWelcome               MessageType = "welcome"
	ResponseLoginSuccessful       MessageType = "login_successful"
	ResponseLoginFailed     	  MessageType = "login_unsuccessful"
	ResponseRegisterSuccessful    MessageType = "register_successful"
	ResponseRegisterFailed        MessageType = "register_unsuccessful"
	ResponseLobbyCreated          MessageType = "lobby_created"
	ResponseLobbyList             MessageType = "lobby_list"
	ResponseLobbyUpdated          MessageType = "lobby_updated"
	ResponseJoinLobbySuccessful   MessageType = "join_lobby_successful"
	ResponseJoinLobbyFailed 	  MessageType = "join_lobby_unsuccessful"
	ResponseLobbyLeft             MessageType = "lobby_left"
	ResponseError                 MessageType = "error"
)

type Player struct {
	ID   string          `json:"id"`
	Name string          `json:"name"`
	Conn *websocket.Conn `json:"-"`
}

type LoginRequest struct {
	Type     MessageType `json:"type"`
	Name     string      `json:"name"`
	Password string      `json:"password"`
}

type RegisterRequest struct {
	Type     MessageType `json:"type"`
	Name     string      `json:"name"`
	Password string      `json:"password"`
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
