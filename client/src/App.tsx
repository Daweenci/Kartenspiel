import React, { useState, useRef } from 'react';
import Login from './pages/Login';
import MainMenu from './pages/MainMenu';
import GameOfTwo from './pages/GameOfTwo';
import GameOfThree from './pages/GameOfThree';
import GameOfFour from './pages/GameOfFour';
import LobbyScreen from './pages/LobbyScreen';
import type { yourLobby, boradcastedLobby, PageType, Player } from './structs';
import { MessageTypes, Page } from './structs';

export default function App() {
  const [player, setPlayer] = useState<Player>({} as Player);
  const [broadcastedLobbies, setbroadcastedLobbies] = useState<boradcastedLobby[]>([]);
  const [gameID, setGameID] = useState<string | null>(null);
  const [gameType, setGameType] = useState<'2' | '3' | '4' | null>(null);
  const [lobby, setLobby] = useState<yourLobby>({} as yourLobby);

  const [currentPage, setCurrentPage] = useState<PageType>(Page.Login);
  const ws = useRef<WebSocket | null>(null);

  const handleLogin = (name: string) => {
    ws.current = new WebSocket(`ws://localhost:4000/ws?name=${encodeURIComponent(name)}`);

    ws.current.onopen = () => {
      console.log('WebSocket connected');
    };

    ws.current.onmessage = async (event) => {
      const data = JSON.parse(event.data);
      switch (data.type) {
        case MessageTypes.ResponseLobbyLeft:
          console.log('Left Lobby');
          setCurrentPage(Page.MainMenu);   
          setLobby({} as yourLobby);
          break;
        case MessageTypes.ResponseLobbyUpdated:
          console.log('Lobby updated:', data.lobby);
          setLobby(data.lobby);
          break;
        case MessageTypes.ResponseLobbyList:
          console.log('Received lobby list:', data.lobbies);
          setbroadcastedLobbies(data.lobbies);
          break;
        case MessageTypes.ResponseWelcome:
          console.log('Welcome:', data.name);
          setPlayer({id: data.id, name: name,});
          setCurrentPage(Page.MainMenu);
          setbroadcastedLobbies(data.lobbies);
          break;
        case MessageTypes.ResponseLobbyCreated:
          setLobby(data.lobby);
          console.log('Lobby created:', data.lobby);
          setCurrentPage(Page.InLobby);
          break;
        case MessageTypes.ResponseJoinLobbySuccess:
          console.log('Joined Lobby:', data);
          setLobby(data.lobby)
          setCurrentPage(Page.InLobby);
          console.log('Joined Lobby:', data.lobby);
          break;
        case MessageTypes.ResponseJoinLobbyFailure:
          console.log("incorrect Password");
          break;
        default:
          console.warn('Unknown message type:', data.type);
      }
    };

    ws.current.onerror = (err) => {
      console.error('WebSocket error:', err);
    };

    ws.current.onclose = () => {
      console.log('WebSocket closed');
      setPlayer({} as Player);
    };
  };

  const handleCreatelobby = (lobbyName: string, maxPlayers: number, isPrivate: boolean, password: string) => {
    if (!ws.current || ws.current.readyState !== WebSocket.OPEN) {
      console.error('WebSocket is not connected');
      return;
    }

    const lobbyData = {
      type: MessageTypes.RequestCreateLobby,
      lobbyName,
      maxPlayers,
      isPrivate,
      password,
      playerID: player.id,
      playerName: player.name,
    };

    ws.current.send(JSON.stringify(lobbyData));
    console.log('Lobby creation request sent:', lobbyData);
  };

  const handleStartGame = () => {
    if (!ws.current || ws.current.readyState !== WebSocket.OPEN) {
      console.error('WebSocket is not connected');
      return;
    }

    const startGameData = {
      type: MessageTypes.RequestStartGame,
      lobbyID: lobby.id,
      PlayerID: player.id,
    };

    ws.current.send(JSON.stringify(startGameData));
    console.log('Game start request sent:', startGameData);
  }

  const handleCancelGame = () => {
    if (!ws.current || ws.current.readyState !== WebSocket.OPEN) {
      console.error('WebSocket is not connected');
      return;
    }

    const cancelGameData = {
      type: MessageTypes.RequestCancelGame,
      lobbyID: lobby.id,
      playerID: player.id,
    };

    ws.current.send(JSON.stringify(cancelGameData));
    console.log('Game cancel request sent:', cancelGameData);
  };

  const handleLeaveLobby = () => {
    if (!ws.current || ws.current.readyState !== WebSocket.OPEN) {
      console.error('WebSocket is not connected');
      return;
    }

    const leaveLobbyData = {
      type: MessageTypes.RequestLeaveLobby,
      lobbyID: lobby.id,
      playerID: player.id,
    };

    ws.current.send(JSON.stringify(leaveLobbyData));
    console.log('Leave lobby request sent:', leaveLobbyData);
  }

  const handleJoinLobby = (id: string, joinPassword: string) => {
    if (!ws.current || ws.current.readyState !== WebSocket.OPEN) {
      console.error('WebSocket is not connected');
      return;
    }
    
    const joinLobbyData = {
      type: MessageTypes.RequestJoinLobby,
      lobbyID: id,
      playerID: player.id,
      password: joinPassword
    }

    ws.current.send(JSON.stringify(joinLobbyData));
    console.log('Join lobby request sent:', );
  }

  switch (currentPage) {
    case Page.Login:
      return <Login onLogin={handleLogin}/>;
    case Page.MainMenu:
      return <MainMenu createLobby={handleCreatelobby} joinLobby={handleJoinLobby} lobbies={broadcastedLobbies} currentPlayerID={player.id}/>;
    case Page.InLobby:
      return <LobbyScreen startGame={handleStartGame} cancelGame={handleCancelGame} leaveLobby={handleLeaveLobby} initLobby={lobby}/>;
    case Page.GameOfTwo:
      return <GameOfTwo />;
    case Page.GameOfThree:
      return <GameOfThree />;
    case Page.GameOfFour:
      return <GameOfFour />;
    default:
      return null;
  }
}