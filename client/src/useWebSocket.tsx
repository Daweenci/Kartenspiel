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

  // Close socket if tab closes
  useEffect(() => {
    const handleUnload = () => {
      if (!ws.current) return;
      ws.current.onclose = () => {};
      ws.current.close();
    };

    window.addEventListener('beforeunload', handleUnload);

    return () => {
      window.removeEventListener('beforeunload', handleUnload);
      ws.current?.close();
    };
  }, []);

  const connect = () => {

    // Prevent duplicate connections
    if (ws.current && ws.current.readyState === WebSocket.OPEN) return;

    const token = getAuthToken();
    if (!token) {
      console.error('No auth token found. Cannot connect WebSocket.');
      onSetPage(Page.Auth);
      return;
    }

    const wsUrl = import.meta.env.REACT_APP_WS_URL || 'ws://localhost:4000/ws';

    ws.current = new WebSocket(wsUrl);

    ws.current.onopen = () => {
      if (ws.current && ws.current.readyState === WebSocket.OPEN) {
        ws.current.send(JSON.stringify({
          type: MessageTypes.RequestAuthentication,
          token: token
        }));
      }
    };

    ws.current.onmessage = (event) => {
      const data = JSON.parse(event.data);
      console.log('WebSocket message received:', data);

      switch (data.type) {

        case MessageTypes.ResponseWelcome:
          onSetPlayer(data.player);
          onSetPage(Page.MainMenu);
          if (data.message) toast(data.message);
          if (data.lobbies) onSetLobbies(data.lobbies);
          break;

        case MessageTypes.ResponseLobbyList:
          onSetLobbies(data.lobbies);
          break;

        case MessageTypes.ResponseLobbyCreated:
          onSetLobby(data.lobby);
          onSetPage(Page.InLobby);
          toast('Lobby created successfully');
          break;

        case MessageTypes.ResponseJoinLobbySuccessful:
          onSetLobby(data.lobby);
          onSetPage(Page.InLobby);
          toast('Joined lobby successfully');
          break;

        case MessageTypes.ResponseLobbyUpdated:
          onSetLobby(data.lobby);
          break;

        case MessageTypes.ResponseLobbyLeft:
          onSetLobby({} as yourLobby);
          onSetPage(Page.MainMenu);
          break;

        case MessageTypes.ResponseError:
          toast(data.error || 'An error occurred');

          if (data.error?.includes('token') || data.error?.includes('auth')) {
            clearAuthToken();
            onSetPage(Page.Auth);
          }
          break;

        default:
          console.warn('Unknown message type:', data.type);
      }
    };

    ws.current.onerror = (err) => {
      console.error('WebSocket error:', err);
      toast('Connection error');
    };

    ws.current.onclose = (event) => {
      console.log('WebSocket closed', event.code, event.reason);

      onSetPlayer({} as Player);

      if (event.code === 1008 || event.code === 1011) {
        clearAuthToken();
        onSetPage(Page.Auth);
        toast('Authentication failed, please log in again');
      }
    };
  };

  const sendMessage = (msg: any) => {
    if (ws.current && ws.current.readyState === WebSocket.OPEN) {
      ws.current.send(JSON.stringify(msg));
    } else {
      console.error('WebSocket is not connected');
      toast('Connection lost, please refresh');
    }
  };

  const createLobby = (
    lobbyName: string,
    maxPlayers: number,
    isPrivate: boolean,
    password: string
  ) =>
    sendMessage({
      type: MessageTypes.RequestCreateLobby,
      lobbyName,
      maxPlayers,
      isPrivate,
      password
    });

  const joinLobby = (lobbyID: string, password: string) =>
    sendMessage({
      type: MessageTypes.RequestJoinLobby,
      lobbyID,
      password
    });

  const leaveLobby = (lobbyID: string) =>
    sendMessage({
      type: MessageTypes.RequestLeaveLobby,
      lobbyID
    });

  const startGame = (lobbyID: string) =>
    sendMessage({
      type: MessageTypes.RequestStartGame,
      lobbyID
    });

  const cancelGame = (lobbyID: string) =>
    sendMessage({
      type: MessageTypes.RequestCancelGame,
      lobbyID
    });

  const logout = () => {
    clearAuthToken();
    ws.current?.close();
    onSetPlayer({} as Player);
    onSetLobby({} as yourLobby);
    onSetLobbies([]);
    onSetPage(Page.Auth);
  };

  // AUTO CONNECT ON PAGE LOAD IF TOKEN EXISTS
  useEffect(() => {
    const token = getAuthToken();
    if (token) {
      connect();
    }
  }, []);

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