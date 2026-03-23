export type Player = {
	name: string;
	id:   string;
};

export type Friend = {
  id: string;
  name: string;
  isOnline: boolean;
}

export type yourLobby = {
  id: string;
  name: string;
  maxPlayers: number;
  isPrivate: boolean;
  password: string; 
  players: Player[];
  gameStart: PlayersStarted[];
};

export type broadcastedLobby = {
  id: string;
  name: string;
  maxPlayers: number;
  isPrivate: boolean;
  players: Player[];
  gameStart: PlayersStarted[];
};

export type friendRequest = {
  friendID: string;
  friendName: string;
}

export const MessageTypes = {
  //Sent from Server
  RequestAuthentication: "authenticate",
  RequestLogin: 'login',
  RequestRegister: 'register',
  RequestCreateLobby: 'create_lobby',
  RequestJoinLobby: 'join_lobby',
  RequestLeaveLobby: 'leave_lobby',
  RequestStartGame: 'start_game',
  RequestCancelGame: 'cancel_game',
  RequestAddFriend: 'add_friend',
  RequestAcceptFriendRequest: 'accept_friend_request',
  RequestGetPendingFriendRequests: 'get_pending_friend_requests',

  //Sent from Client
  ResponseLoginFailed: 'login_failed',
  ResponseLoginSuccessful: 'login_successful',
  ResponseRegisterFailed: 'register_failed',
  ResponseRegisterSuccessful: 'register_successful',
  ResponseWelcome: 'welcome',
  ResponseLobbyCreated: 'lobby_created',
  ResponseLobbyList: 'lobby_list',
  ResponseLobbyUpdated: 'lobby_updated',
  ResponseJoinLobbySuccessful: 'join_lobby_successful',
  ResponseJoinLobbyFailed: 'join_lobby_failed',
  ResponseLobbyLeft: 'lobby_left',
  ResponsePendingFriendRequests: 'pending_friend_requests',
  ResponseFriendRequestSent: 'friend_request_sent',
  ResponseFriendRequestReceived: 'friend_request_received',
  ResponseFriendRequestAccepted: 'friend_request_accepted',
  ResponseFriendOnlineStatus: 'friend_online_status',
  ResponseFriendsList: 'friends_list',
  ResponseError: 'error',
} as const;

export const Page = {
  Auth: 'auth',
  MainMenu: 'main_menu',
  InLobby: 'in_lobby',
  GameOfTwo: 'game_of_two',
  GameOfThree: 'game_of_three',
  GameOfFour: 'game_of_four',
  LobbyScreen: 'lobby_screen',
} as const;

export type PageType = typeof Page[keyof typeof Page];

type PlayersStarted = {
  playerID: string;
};
