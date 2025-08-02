import React, { useState } from 'react';
import Login from './pages/Login';
import MainMenu from './pages/MainMenu';
import GameOfTwo from './pages/GameOfTwo';
import GameOfThree from './pages/GameOfThree';
import GameOfFour from './pages/GameOfFour';
import LobbyScreen from './pages/LobbyScreen';
import type { yourLobby, broadcastedLobby, PageType, Player } from './structs';
import { Page } from './structs';
import useWebSocket from './useWebSocket';
import { Toaster } from 'sonner';

export default function App() {
  const [player, setPlayer] = useState<Player>({} as Player);
  const [broadcastedLobbies, setbroadcastedLobbies] = useState<broadcastedLobby[]>([]);
  const [lobby, setLobby] = useState<yourLobby>({} as yourLobby);
  const [currentPage, setCurrentPage] = useState<PageType>(Page.Login);

  const {
    connect,
    createLobby,
    startGame,
    cancelGame,
    leaveLobby,
    joinLobby
  } = useWebSocket({
    onSetPlayer: setPlayer,
    onSetLobby: setLobby,
    onSetLobbies: setbroadcastedLobbies,
    onSetPage: setCurrentPage,
  });

  const handleLogin = (name: string) => connect(name);
  const handleCreateLobby = (name: string, max: number, priv: boolean, pass: string) =>
    createLobby(name, max, priv, pass, player);
  const handleStartGame = () => startGame(lobby.id, player.id);
  const handleCancelGame = () => cancelGame(lobby.id, player.id);
  const handleLeaveLobby = () => leaveLobby(lobby.id, player.id);
  const handleJoinLobby = (id: string, pass: string) => joinLobby(id, pass, player.id);

  return (
    <>
    <Toaster />
      {(() => {
        switch (currentPage) {
          case Page.Login:
            return <Login onLogin={handleLogin} />;
          case Page.MainMenu:
            return (
              <MainMenu
                createLobby={handleCreateLobby}
                joinLobby={handleJoinLobby}
                lobbies={broadcastedLobbies}
                currentPlayerID={player.id}
              />
            );
          case Page.InLobby:
            return (
              <LobbyScreen
                startGame={handleStartGame}
                cancelGame={handleCancelGame}
                leaveLobby={handleLeaveLobby}
                initLobby={lobby}
              />
            );
          case Page.GameOfTwo:
            return <GameOfTwo />;
          case Page.GameOfThree:
            return <GameOfThree />;
          case Page.GameOfFour:
            return <GameOfFour />;
          default:
            return null;
        }
      })()}
    </>
  );
}
