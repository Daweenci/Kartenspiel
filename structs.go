package main

import (
	"sync"

	"github.com/gorilla/websocket"
)

type MessageType string

const (
	RequestAuthentication MessageType = "authenticate"
	RequestLogin          MessageType = "login"
	RequestRegister       MessageType = "register"
	RequestCreateLobby    MessageType = "create_lobby"
	RequestJoinLobby      MessageType = "join_lobby"
	RequestLeaveLobby     MessageType = "leave_lobby"
	RequestStartGame      MessageType = "start_game"
	RequestCancelGame     MessageType = "cancel_game"

	ResponseWelcome             MessageType = "welcome"
	ResponseLoginSuccessful     MessageType = "login_successful"
	ResponseLoginFailed         MessageType = "login_failed"
	ResponseRegisterSuccessful  MessageType = "register_successful"
	ResponseRegisterFailed      MessageType = "register_failed"
	ResponseLobbyCreated        MessageType = "lobby_created"
	ResponseLobbyList           MessageType = "lobby_list"
	ResponseLobbyUpdated        MessageType = "lobby_updated"
	ResponseJoinLobbySuccessful MessageType = "join_lobby_successful"
	ResponseJoinLobbyFailed     MessageType = "join_lobby_failed"
	ResponseLobbyLeft           MessageType = "lobby_left"
	ResponseError               MessageType = "error"
)

type Player struct {
	ID   string
	Name string
	Conn *websocket.Conn
}

type PlayerResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
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

type Lobby struct {
	ID         string
	Name       string
	MaxPlayers int
	IsPrivate  bool
	Password   string
	Players    []*Player
	GameStart  []PlayerStarted
	Lock       sync.RWMutex
}

type LobbyResponse struct {
	ID         string           `json:"id"`
	Name       string           `json:"name"`
	MaxPlayers int              `json:"maxPlayers"`
	IsPrivate  bool             `json:"isPrivate"`
	Players    []PlayerResponse `json:"playersResponse"`
	GameStart  []PlayerStarted  `json:"gameStart"`
}

type LobbyUpdatedResponse struct {
	Type  MessageType   `json:"type"`
	Lobby LobbyResponse `json:"lobby"`
}

type LobbiesUpdateResponse struct {
	Type    MessageType     `json:"type"`
	Lobbies []LobbyResponse `json:"lobbies"`
}

type IncorrectLobbyPasswordResponse struct {
	Type    MessageType `json:"type"`
	Message string      `json:"message"`
}

type SuccessfulJoinLobbyResponse struct {
	Type  MessageType   `json:"type"`
	Lobby LobbyResponse `json:"lobby"`
}

type CreateLobbyResponse struct {
	Type  MessageType   `json:"type"`
	Lobby LobbyResponse `json:"lobby"`
}

type PlayerStarted struct {
	ID string `json:"id"`
}

type ErrorResponse struct {
	Type  MessageType `json:"type"`
	Error string      `json:"error"`
}
