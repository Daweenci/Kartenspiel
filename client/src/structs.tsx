export type Player = {
	name: string;
	id:   string;
};

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
};

export const MessageTypes = {
  //Sent from Server
  RequestLogin: 'login',
  RequestRegister: 'register',
  RequestCreateLobby: 'create_lobby',
  RequestJoinLobby: 'join_lobby',
  RequestLeaveLobby: 'leave_lobby',
  RequestStartGame: 'start_game',
  RequestCancelGame: 'cancel_game',

  //Sent from Client
  ResponseLoginFailed: 'login_unsuccessful',
  ResponseLoginSuccess: 'login_successful',
  ResponseRegisterFailed: 'register_unsuccessful',
  ResponseRegisterSuccess: 'register_successful',
  ResponseWelcome: 'welcome',
  ResponseLobbyCreated: 'lobby_created',
  ResponseLobbyList: 'lobby_list',
  ResponseLobbyUpdated: 'lobby_updated',
  ResponseJoinLobbySuccess: 'join_lobby_successful',
  ResponseJoinLobbyUnsuccessful: 'join_lobby_unsuccessful',
  ResponseLobbyLeft: 'lobby_left',
  ResponseError: 'error',
} as const;

export const Page = {
  Auth: 'auth',
  MainMenu: 'mainmenu',
  InLobby: 'inlobby',
  GameOfTwo: 'gameoftwo',
  GameOfThree: 'gameofthree',
  GameOfFour: 'gameoffour',
  LobbyScreen: 'lobbyscreen',
} as const;

export type PageType = typeof Page[keyof typeof Page];

type PlayersStarted = {
  playerID: string;
  gameStarted: boolean;
};
