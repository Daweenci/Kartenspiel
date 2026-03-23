package main

import (
	"sync"

	"github.com/gorilla/websocket"
)

type MessageType string

const (
	RequestAuthentication      MessageType = "authenticate"
	RequestLogin               MessageType = "login"
	RequestRegister            MessageType = "register"
	RequestCreateLobby         MessageType = "create_lobby"
	RequestJoinLobby           MessageType = "join_lobby"
	RequestLeaveLobby          MessageType = "leave_lobby"
	RequestStartGame           MessageType = "start_game"
	RequestCancelGame          MessageType = "cancel_game"
	RequestAddFriend           MessageType = "add_friend"
	RequestAcceptFriendRequest MessageType = "accept_friend_request"

	ResponseWelcome               MessageType = "welcome"
	ResponseLoginSuccessful       MessageType = "login_successful"
	ResponseLoginFailed           MessageType = "login_failed"
	ResponseRegisterSuccessful    MessageType = "register_successful"
	ResponseRegisterFailed        MessageType = "register_failed"
	ResponseLobbyCreated          MessageType = "lobby_created"
	ResponseLobbyList             MessageType = "lobby_list"
	ResponseLobbyUpdated          MessageType = "lobby_updated"
	ResponseJoinLobbySuccessful   MessageType = "join_lobby_successful"
	ResponseJoinLobbyFailed       MessageType = "join_lobby_failed"
	ResponseLobbyLeft             MessageType = "lobby_left"
	ResponseFriendRequestSent     MessageType = "friend_request_sent"
	ResponsePendingFriendRequests MessageType = "pending_friend_requests"
	ResponseFriendRequestReceived MessageType = "friend_request_received"
	ResponseFriendRequestAccepted MessageType = "friend_request_accepted"
	ResponseFriendCameOnline      MessageType = "friend_came_online"
	ResponseFriendsList           MessageType = "friends_list"
	ResponseError                 MessageType = "error"
)

type Player struct {
	ID   string
	Name string
	Conn *websocket.Conn
	Send chan []byte
}

type PlayerDTO struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type FriendDTO struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	IsOnline bool   `json:"isOnline"`
}

type PlayerStarted struct {
	ID string `json:"playerID"`
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

type AddFriendRequest struct {
	Type       MessageType `json:"type"`
	FriendName string      `json:"friendName"`
	PlayerID   string      `json:"playerID"`
}

type GetPendingFriendRequestsRequest struct {
	Type     MessageType `json:"type"`
	PlayerID string      `json:"playerID"`
}

type AcceptFriendRequestRequest struct {
	Type          MessageType `json:"type"`
	FriendID      string      `json:"friendID"` //person that requested in the first place
	AcceptRequest bool        `json:"acceptRequest"`
	PlayerID      string      `json:"playerID"`
}

type StartGame struct { //TODO: Why StartGame not response or request?
	Type     MessageType `json:"type"`
	LobbyID  string      `json:"lobbyID"`
	PlayerID string      `json:"playerID"`
}

type CancelGame struct { //TODO: Why CancelGame not response or request?
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

type Response interface {
	GetType() MessageType
}

type BaseResponse struct {
	Type MessageType `json:"type"`
}

func newBaseResponse(t MessageType) BaseResponse {
	return BaseResponse{Type: t}
}

func (r BaseResponse) GetType() MessageType {
	return r.Type
}

type WelcomeResponse struct {
	BaseResponse
	Player                PlayerDTO   `json:"player"`
	Message               string      `json:"message"`
	Lobbies               []LobbyDTO  `json:"lobbies"`
	PendingFriendRequests []PlayerDTO `json:"pendingFriendRequests"`
	FriendsList           []PlayerDTO `json:"friendsList"`
}

type LobbyDTO struct {
	ID         string          `json:"id"`
	Name       string          `json:"name"`
	MaxPlayers int             `json:"maxPlayers"`
	IsPrivate  bool            `json:"isPrivate"`
	Players    []PlayerDTO     `json:"players"`
	GameStart  []PlayerStarted `json:"gameStart"`
}

type LobbyUpdatedResponse struct {
	BaseResponse
	Lobby LobbyDTO `json:"lobby"`
}

type LobbiesUpdateResponse struct {
	BaseResponse
	Lobbies []LobbyDTO `json:"lobbies"`
}

type LobbyJoinFailedResponse struct {
	BaseResponse
	Message string `json:"message"`
}

type SuccessfulJoinLobbyResponse struct {
	BaseResponse
	Lobby LobbyDTO `json:"lobby"`
}

type CreateLobbyResponse struct {
	BaseResponse
	Lobby LobbyDTO `json:"lobby"`
}

type LobbyLeftResponse struct {
	BaseResponse
}

type FriendRequestSentResponse struct {
	BaseResponse
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type PendingFriendRequestsResponse struct {
	BaseResponse
	PendingFriendRequests []PlayerDTO `json:"pendingFriendRequests"`
}

type FriendRequestReceivedResponse struct {
	BaseResponse
	Requester PlayerDTO `json:"requester"`
}

type FriendRequestAcceptedResponse struct {
	BaseResponse
	Friend FriendDTO `json:"friend"`
}

type FriendsListResponse struct {
	BaseResponse
	FriendsList []FriendDTO `json:"friendsList"`
}

type FriendCameOnlineResponse struct {
	BaseResponse
	Friend PlayerDTO `json:"friend"`
}

type ErrorResponse struct {
	BaseResponse
	Error string `json:"error"`
}
