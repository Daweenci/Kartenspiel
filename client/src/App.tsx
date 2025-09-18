import React, { useState } from 'react';
import Login from './pages/Auth';
import MainMenu from './pages/MainMenu';
import GameOfTwo from './pages/GameOfTwo';
import GameOfThree from './pages/GameOfThree';
import GameOfFour from './pages/GameOfFour';
import LobbyScreen from './pages/LobbyScreen';
import type { yourLobby, broadcastedLobby, PageType, Player } from './structs';
import { Page } from './structs';
import useWebSocket from './useWebSocket';
import { Toaster } from 'sonner';
import Auth from './pages/Auth';

export default function App() {
  const [player, setPlayer] = useState<Player>({} as Player);
  const [broadcastedLobbies, setbroadcastedLobbies] = useState<broadcastedLobby[]>([]);
  const [lobby, setLobby] = useState<yourLobby>({} as yourLobby);
  const [currentPage, setCurrentPage] = useState<PageType>(Page.Auth);

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
  const handleConnectWebSocket = () => { connect(); };
  const handleCreateLobby = (name: string, max: number, priv: boolean, pass: string) => createLobby(name, max, priv, pass);
  const handleStartGame = () => startGame(lobby.id);
  const handleCancelGame = () => cancelGame(lobby.id);
  const handleLeaveLobby = () => leaveLobby(lobby.id);
  const handleJoinLobby = (lobbyID: string, lobbyPassword: string) => joinLobby(lobbyID, lobbyPassword);

  return (
    <>
    <Toaster />
      {(() => {
        switch (currentPage) {
          case Page.Auth:
            return <Auth connectWebSocket={handleConnectWebSocket} />;
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
