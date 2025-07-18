import React, { useState, useRef } from 'react';
import Login from './pages/Login';
import MainMenu from './pages/MainMenu';
import GameOfTwo from './pages/GameOfTwo';
import GameOfThree from './pages/GameOfThree';
import GameOfFour from './pages/GameOfFour';
import LobbyScreen from './pages/LobbyScreen';
import type { Lobby, PageType } from './structs';
import { MessageTypes, Page } from './structs';

export default function App() {
  const [playerID, setPlayerID] = useState<string | null>(null);
  const [playerName, setPlayerName] = useState<string>('');
  const [currentLobbies, setCurrentLobbies] = useState<Lobby[]>([]);
  const [gameID, setGameID] = useState<string | null>(null);
  const [gameType, setGameType] = useState<'2' | '3' | '4' | null>(null);
  const [lobby, setLobby] = useState<Lobby>({} as Lobby);

  const [currentPage, setCurrentPage] = useState<PageType>(Page.Login);
  const ws = useRef<WebSocket | null>(null);

  const handleLogin = (name: string) => {
    ws.current = new WebSocket(`ws://localhost:4000/ws?name=${encodeURIComponent(name)}`);

    ws.current.onopen = () => {
      console.log('WebSocket connected');
    };

    ws.current.onmessage = (event) => {
      const data = JSON.parse(event.data);
      switch (data.type) {
        case MessageTypes.ResponseLobbyList:
          console.log('Received lobby list:', data.lobbies);
          setCurrentLobbies(data.lobbies);
          break;
        case MessageTypes.ResponseWelcome:
          setPlayerID(data.id);
          setPlayerName(name);
          setCurrentPage(Page.MainMenu);
          setCurrentLobbies(data.lobbies);
          break;
        case MessageTypes.ResponseLobbyCreated:
          setLobby(data.lobby);
          console.log('Lobby created:', data.lobby);
          setCurrentPage(Page.InLobby);
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
      setPlayerID(null);
      setPlayerName('');
      setCurrentPage(Page.Login); 
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
      playerID,
      playerName,
    };

    ws.current.send(JSON.stringify(lobbyData));
    console.log('Lobby creation request sent:', lobbyData);
  };

  const handleStartgame = () => {

  }

  const handleCancelGame = () => {

  };

  const handleLeaveLobby = () => {

  }

  switch (currentPage) {
    case Page.Login:
      return <Login onLogin={handleLogin}/>;
    case Page.MainMenu:
      return <MainMenu createLobby={handleCreatelobby} lobbies={currentLobbies}/>;
    case Page.InLobby:
      return <LobbyScreen startGame={handleStartgame} cancelGame={handleCancelGame} leaveLobby={handleLeaveLobby} initLobby={lobby}/>;
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