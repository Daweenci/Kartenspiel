import React, { useState, useRef } from 'react';
import Login from './pages/Login';
import MainMenu from './pages/MainMenu';
import GameOfTwo from './pages/GameOfTwo';
import GameOfThree from './pages/GameOfThree';
import GameOfFour from './pages/GameOfFour';
import Lobby from './pages/Lobby';

type Player = {
	name: string;
	id:   string;
};

type Lobby = {
  id: string;
  name: string;
  maxPlayers: number;
  isPrivate: boolean;
  players: Player[];
};

export default function App() {
  const [playerID, setPlayerID] = useState<string | null>(null);
  const [playerName, setPlayerName] = useState<string>('');
  const [currentLobbies, setCurrentLobbies] = useState<Lobby[]>([]);
  const [gameID, setGameID] = useState<string | null>(null);
  const [gameType, setGameType] = useState<'two' | 'three' | 'four' | null>(null);
  const [lobby, setLobby] = useState<string | null>(null);

  const [currentPage, setCurrentPage] = useState<'login' | 'mainmenu' | 'inlobby' | 'gameoftwo' | 'gameofthree' | 'gameoffour'>('login');
  const ws = useRef<WebSocket | null>(null);

  const handleLogin = (name: string) => {
    ws.current = new WebSocket(`ws://localhost:4000/ws?name=${encodeURIComponent(name)}`);

    ws.current.onopen = () => {
      console.log('WebSocket connected');
    };

    ws.current.onmessage = (event) => {
      const data = JSON.parse(event.data);
      switch (data.type) {
        case 'lobby_list':
          console.log('Received lobby list:', data.lobbies);
          setCurrentLobbies(data.lobbies);
          break;
        case 'welcome':
          setPlayerID(data.id);
          setPlayerName(name);
          setCurrentPage('mainmenu');
          setCurrentLobbies(data.lobbies);
          break;
        case 'lobby_created':
          setLobby(data.lobby);
          setCurrentPage('inlobby');
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
      setCurrentPage('login'); 
    };
  };

  const handleCreatelobby = (lobbyName: string, maxPlayers: number, isPrivate: boolean, password: string) => {
    if (!ws.current || ws.current.readyState !== WebSocket.OPEN) {
      console.error('WebSocket is not connected');
      return;
    }

    const lobbyData = {
      type: 'create_lobby',
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

  switch (currentPage) {
    case 'login':
      return <Login onLogin={handleLogin}/>;
    case 'mainmenu':
      return <MainMenu createLobby={handleCreatelobby} lobbies={currentLobbies}/>;
    case 'inlobby':
      return <Lobby />;
    case 'gameoftwo':
      return <GameOfTwo />;
    case 'gameofthree':
      return <GameOfThree />;
    case 'gameoffour':
      return <GameOfFour />;
    default:
      return null;
  }
}