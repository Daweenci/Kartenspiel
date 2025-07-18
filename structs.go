package main

type MessageType string

const (
	//Recieved message types
	RequestCreateLobby MessageType = "create_lobby"
	RequestJoinLobby   MessageType = "join_lobby"
	RequestLeaveLobby  MessageType = "leave_lobby"

	//Sent message types
	ResponseWelcome      MessageType = "welcome"
	ResponseLobbyCreated MessageType = "lobby_created"
	ResponseLobbyList    MessageType = "lobby_list"
)

type Player struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type JoinLobbyRequest struct {
	Type    MessageType `json:"type"`
	LobbyID string      `json:"lobby_id"`
	ID      string      `json:"id"`
	Name    string      `json:"name"`
}

type CreateLobbyRequest struct {
	Type       MessageType `json:"type"`
	LobbyName  string      `json:"lobbyName"`
	MaxPlayers int         `json:"maxPlayers"`
	IsPrivate  bool        `json:"isPrivate"`
	Password   string      `json:"password,omitempty"`
	PlayerID   string      `json:"playerID"`
	PlayerName string      `json:"playerName"`
}

type LeaveLobbyRequest struct {
	Type    MessageType `json:"type"`
	LobbyID string      `json:"lobby_id"`
	ID      string      `json:"id"`
}

type Lobby struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	MaxPlayers int      `json:"maxPlayers"`
	IsPrivate  bool     `json:"isPrivate"`
	Password   string   `json:"password,omitempty"`
	Players    []Player `json:"players"`
}
