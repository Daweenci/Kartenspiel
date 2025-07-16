import React, { useState, useRef } from 'react';
import Login from './pages/Login';
import MainMenu from './pages/MainMenu';
import GameOfTwo from './pages/GameOfTwo';
import GameOfThree from './pages/GameOfThree';
import GameOfFour from './pages/GameOfFour';

export default function App() {
  const [playerID, setPlayerID] = useState<string | null>(null);
  const [playerName, setPlayerName] = useState<string>('');
  const [gameID, setGameID] = useState<string | null>(null);
  const [gameType, setGameType] = useState<'two' | 'three' | 'four' | null>(null);
  const [lobbyID, setLobbyID] = useState<string | null>(null);

  const [currentPage, setCurrentPage] = useState<'login' | 'mainmenu' | 'game' | 'gameoftwo' | 'gameofthree' | 'gameoffour'>('login');
  const ws = useRef<WebSocket | null>(null);

  const handleLogin = (name: string) => {
    ws.current = new WebSocket(`ws://localhost:4000/ws?name=${encodeURIComponent(name)}`);

    ws.current.onopen = () => {
      console.log('WebSocket connected');
    };

    ws.current.onmessage = (event) => {
      const data = JSON.parse(event.data);
      if (data.type === 'welcome') {
        setPlayerID(data.id);
        setPlayerName(name);
        setCurrentPage('mainmenu'); // nach Login ins Hauptmenü
      }
      // weitere Nachrichten verarbeiten...
    };

    ws.current.onerror = (err) => {
      console.error('WebSocket error:', err);
    };

    ws.current.onclose = () => {
      console.log('WebSocket closed');
      setPlayerID(null);
      setPlayerName('');
      setCurrentPage('login'); // zurück zum Login bei Verbindungsverlust
    };
  };

  switch (currentPage) {
    case 'login':
      return <Login onLogin={handleLogin}/>;
    case 'mainmenu':
      return <MainMenu />;
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