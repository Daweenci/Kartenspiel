// useGameWebSocket.ts
import { useRef, useEffect } from 'react';
import type { yourLobby, broadcastedLobby, Player, PageType } from './structs';
import { MessageTypes, Page } from './structs';
import { toast } from 'sonner';

interface UseWebSocketProps {
  onSetPlayer: (player: Player) => void;
  onSetLobby: (lobby: yourLobby) => void;
  onSetLobbies: (lobbies: broadcastedLobby[]) => void;
  onSetPage: (page: PageType) => void;
}

export default function useWebSocket({
  onSetPlayer,
  onSetLobby,
  onSetLobbies,
  onSetPage,
}: UseWebSocketProps) {
  const ws = useRef<WebSocket | null>(null);

  const tokenRef = useRef<string | null>(null);

  const setAuthToken = (token: string) => {
    tokenRef.current = token;
    localStorage.setItem('gameToken', token);
  };

  const getAuthToken = () => {
    if (!tokenRef.current) tokenRef.current = localStorage.getItem('gameToken');
    return tokenRef.current;
  };

  const clearAuthToken = () => {
    tokenRef.current = null;
    localStorage.removeItem('gameToken');
  };

  // Cross-tab token sync
  useEffect(() => {
    const handleStorage = (event: StorageEvent) => {
      if (event.key === 'gameToken') tokenRef.current = event.newValue;
    };
    window.addEventListener('storage', handleStorage);
    return () => window.removeEventListener('storage', handleStorage);
  }, []);

  // Cleanup WebSocket
  useEffect(() => {
    return () => {
      ws.current?.close();
    };
  }, []);

  const connect = () => {
    const token = getAuthToken();
    if (!token) {
      console.error('No auth token found. Cannot connect WebSocket.');
      return;
    }

    ws.current = new WebSocket(process.env.REACT_APP_WS_URL || 'ws://localhost:4000/ws');

    ws.current.onopen = () => console.log('WebSocket connected');

    ws.current.onmessage = (event) => {
      const data = JSON.parse(event.data);

      switch (data.type) {
        case MessageTypes.ResponseWelcome:
          onSetPlayer(data.player);
          onSetPage(Page.MainMenu);
          toast(data.message);
          break;

        case MessageTypes.ResponseLobbyList:
          onSetLobbies(data.lobbies);
          break;

        case MessageTypes.ResponseLobbyCreated:
          onSetLobby(data.lobby);
          onSetPage(Page.InLobby);
          toast(data.message);
          break;

        case MessageTypes.ResponseJoinLobbySuccess:
          onSetLobby(data.lobby);
          onSetPage(Page.InLobby);
          toast(data.message);
          break;

        case MessageTypes.ResponseLobbyUpdated:
          onSetLobby(data.lobby);
          break;

        case MessageTypes.ResponseLobbyLeft:
          onSetLobby({} as yourLobby);
          onSetPage(Page.MainMenu);
          break;

        case MessageTypes.ResponseError:
          toast(data.error);
          break;

        default:
          console.warn('Unknown message type:', data.type);
      }
    };

    ws.current.onerror = (err) => console.error('WebSocket error:', err);
    ws.current.onclose = () => {
      console.log('WebSocket closed');
      onSetPlayer({} as Player);
    };
  };

  const sendMessage = (msg: any) => {
    if (ws.current && ws.current.readyState === WebSocket.OPEN) {
      const token = getAuthToken();
      ws.current.send(JSON.stringify({ ...msg, token }));
    } else {
      console.error('WebSocket is not connected');
    }
  };
  // Lobby / game actions
  const createLobby = (lobbyName: string, maxPlayers: number, isPrivate: boolean, password: string) =>
    sendMessage({ type: MessageTypes.RequestCreateLobby, lobbyName, maxPlayers, isPrivate, password });

  const joinLobby = (lobbyID: string, password: string) =>
    sendMessage({ type: MessageTypes.RequestJoinLobby, lobbyID, password });

  const leaveLobby = (lobbyID: string) => sendMessage({ type: MessageTypes.RequestLeaveLobby, lobbyID });

  const startGame = (lobbyID: string) => sendMessage({ type: MessageTypes.RequestStartGame, lobbyID });

  const cancelGame = (lobbyID: string) => sendMessage({ type: MessageTypes.RequestCancelGame, lobbyID });

  const logout = () => {
    clearAuthToken();
    ws.current?.close();
    onSetPlayer({} as Player);
    onSetLobby({} as yourLobby);
    onSetLobbies([]);
    onSetPage(Page.MainMenu);
  };

  return {
    connect,
    logout,
    createLobby,
    joinLobby,
    leaveLobby,
    startGame,
    cancelGame,
    getAuthToken,
    setAuthToken,
    clearAuthToken,
  };
}
