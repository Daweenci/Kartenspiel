// useGameWebSocket.ts
import { useRef, useEffect } from 'react';
import type { yourLobby, broadcastedLobby, Player, PageType } from './structs';
import { MessageTypes, Page } from './structs';
import { toast } from "sonner"

interface ExtendedWebSocket extends WebSocket {
  _joinLobbyResolve?: (value: boolean) => void;
}

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
  const ws = useRef<ExtendedWebSocket | null>(null);

  const connect = (name: string) => {
    ws.current = new WebSocket(`ws://localhost:4000/ws?name=${encodeURIComponent(name)}`) as ExtendedWebSocket;

    ws.current.onopen = () => console.log('WebSocket connected');

    ws.current.onmessage = (event) => {
      const data = JSON.parse(event.data);
      const resolve = ws.current?._joinLobbyResolve;

      switch (data.type) {
        case MessageTypes.ResponseLobbyLeft:
          onSetPage(Page.MainMenu);
          onSetLobby({} as yourLobby);
          break;
        case MessageTypes.ResponseLobbyUpdated:
          onSetLobby(data.lobby);
          break;
        case MessageTypes.ResponseLobbyList:
          onSetLobbies(data.lobbies);
          break;
        case MessageTypes.ResponseWelcome:
          onSetPlayer({ id: data.id, name });
          onSetPage(Page.MainMenu);
          onSetLobbies(data.lobbies);
          break;
        case MessageTypes.ResponseLobbyCreated:
          onSetLobby(data.lobby);
          onSetPage(Page.InLobby);
          break;
        case MessageTypes.ResponseJoinLobbySuccess:
          onSetLobby(data.lobby);
          onSetPage(Page.InLobby);
          if (resolve) {
            resolve(true);
            delete ws.current!._joinLobbyResolve;
          }
          break;
        case MessageTypes.ResponseJoinLobbyWrongPassword:
          toast("Wrong password");
          resolve?.(false);
          delete ws.current!._joinLobbyResolve;
          break;
        case MessageTypes.ResponseJoinLobbyFull:
          toast("Lobby full");
          resolve?.(false);
          delete ws.current!._joinLobbyResolve;
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
      ws.current.send(JSON.stringify(msg));
    } else {
      console.error('WebSocket is not connected');
    }
  };

  const createLobby = (lobbyName: string, maxPlayers: number, isPrivate: boolean, password: string, player: Player) => {
    sendMessage({
      type: MessageTypes.RequestCreateLobby,
      lobbyName,
      maxPlayers,
      isPrivate,
      password,
      playerID: player.id,
      playerName: player.name,
    });
  };

  const startGame = (lobbyID: string, playerID: string) => {
    sendMessage({
      type: MessageTypes.RequestStartGame,
      lobbyID,
      PlayerID: playerID,
    });
  };

  const cancelGame = (lobbyID: string, playerID: string) => {
    sendMessage({
      type: MessageTypes.RequestCancelGame,
      lobbyID,
      playerID,
    });
  };

  const leaveLobby = (lobbyID: string, playerID: string) => {
    sendMessage({
      type: MessageTypes.RequestLeaveLobby,
      lobbyID,
      playerID,
    });
  };

  const joinLobby = (lobbyID: string, password: string, playerID: string): Promise<boolean> => {
    return new Promise((resolve) => {
      if (!ws.current || ws.current.readyState !== WebSocket.OPEN) {
        console.error('WebSocket is not connected');
        resolve(false);
        return;
      }

      ws.current._joinLobbyResolve = resolve;

      sendMessage({
        type: MessageTypes.RequestJoinLobby,
        lobbyID,
        playerID,
        password,
      });
    });
  };

  return {
    connect,
    createLobby,
    startGame,
    cancelGame,
    leaveLobby,
    joinLobby,
  };
}
