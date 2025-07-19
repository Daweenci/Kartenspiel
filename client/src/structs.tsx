


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

export type boradcastedLobby = {
  id: string;
  name: string;
  maxPlayers: number;
  isPrivate: boolean;
  players: Player[];
};

export const MessageTypes = {
  //Sent from Server
  RequestCreateLobby: 'create_lobby',
  RequestJoinLobby: 'join_lobby',
  RequestLeaveLobby: 'leave_lobby',
  RequestStartGame: 'start_game',
  RequestCancelGame: 'cancel_game',

  //Sent from Client
  ResponseWelcome: 'welcome',
  ResponseLobbyCreated: 'lobby_created',
  ResponseLobbyList: 'lobby_list',
  ResponseLobbyUpdated: 'lobby_updated',
  ResponseJoinLobbySuccess: 'join_lobby_success',
  ResponseJoinLobbyFailure: 'join_lobby_failure',
  ResponseLobbyLeft: 'lobby_left',
} as const;

export const Page = {
  Login: 'login',
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
