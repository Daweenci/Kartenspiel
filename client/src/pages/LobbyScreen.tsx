import React, { useState } from 'react';
import type { yourLobby } from '@/structs';

type LobbyScreenProps = {
  initLobby: yourLobby;
  startGame: () => void;
  cancelGame: () => void;
  leaveLobby: () => void;
};

export default function LobbyScreen({
  startGame,
  cancelGame,
  leaveLobby,
  initLobby,
}: LobbyScreenProps) {
  const [gameStarting, setGameStarting] = useState(false);

  const handleLeaveLobby = () => {
    leaveLobby();
  };

  const handleToggleGameStart = () => {
    if (gameStarting) {
      cancelGame();
      setGameStarting(false);
    } else {
      startGame();
      setGameStarting(true);
    }
  };

  return (
    <div className="flex items-center justify-center min-h-screen">
      <div className="flex-col items-center justify-center p-16 border-2 border-gray-300 rounded-4xl">
        <h1>
          <strong>Lobby:</strong> {initLobby.name}
        </h1>
        <h1>
          <strong>
            Players {initLobby.players.length}/{initLobby.maxPlayers}:
          </strong>
        </h1>
        <ul className="list-none">
          {initLobby.players.map((player, index) => (
            <li key={index}>{player.name}</li>
          ))}
        </ul>
        {initLobby.password && (
          <h1>
            <strong>Password:</strong> {initLobby.password}
          </h1>
        )}
        <div>
          <button
            onClick={handleLeaveLobby}
            className="mt-4 px-4 py-2 bg-red-500 text-white rounded hover:bg-red-600"
          >
            Leave Lobby
          </button>

          <button
            onClick={handleToggleGameStart}
            className={`mt-4 px-4 py-2 text-white rounded ${
              gameStarting
                ? 'bg-orange-500 hover:bg-orange-600'
                : 'bg-blue-500 hover:bg-blue-600'
            }`}
          >
            {gameStarting ? 'Cancel Start' : 'Start Game'} ({initLobby.gameStart.length}/
            {initLobby.maxPlayers})
          </button>
        </div>
      </div>
    </div>
  );
}
