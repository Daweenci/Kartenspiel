


export type Player = {
	name: string;
	id:   string;
};

export type Lobby = {
  id: string;
  name: string;
  maxPlayers: number;
  isPrivate: boolean;
  password: string; // Only set if its your own lobby or u joined it
  players: Player[];
};

export const MessageTypes = {
  //Sent from Server
  RequestCreateLobby: 'create_lobby',
  RequestJoinLobby: 'join_lobby',
  RequestLeaveLobby: 'leave_lobby',

  //Sent from Client
  ResponseWelcome: 'welcome',
  ResponseLobbyCreated: 'lobby_created',
  ResponseLobbyList: 'lobby_list',
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
