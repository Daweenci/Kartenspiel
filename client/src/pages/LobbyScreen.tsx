import React, { useState, useEffect } from 'react';
import type { Lobby } from '@/structs';

type LobbyScreenProps = {
    initLobby: Lobby;
    startGame: () => void;
    cancelGame: () => void;
    leaveLobby: () => void;
};

export default function LobbyScreen({ startGame, cancelGame, leaveLobby, initLobby }: LobbyScreenProps) {

  return (
    <div className="flex items-center justify-center">
      <div className="flex-col items-center justify-center p-16 border-2 border-gray-300 rounded-4xl">
        <h1><strong>Lobby:</strong> {initLobby.name}</h1>
        <h1><strong>Players {initLobby.players.length}/{initLobby.maxPlayers}:</strong></h1>
        <ul className="list-none ">
          {initLobby.players.map((player, index) => (
            <li key={index}>{player.name}</li>
          ))}
        </ul>
        <h1><strong>Password:</strong> {initLobby.password}</h1>
        <div> 
        <button onClick={handleLeaveLobby} id="leaveLobby" className="mt-4 px-4 py-2 bg-red-500 text-white rounded hover:bg-red-600">
            Leave Lobby
        </button>
        <button onClick={handleStartGame} id="startGame" className="mt-4 px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600">
            Start Game ({initLobby.playersStartedGame}/{initLobby.maxPlayers})
        </button>
        <button onClick={handleCancelGame} id="cancelGame" className="hidden mt-4 px-4 py-2 bg-orange-500 text-white rounded hover:bg-orange-600">
            Cancel Start ({initLobby.playersStartedGame}/{initLobby.maxPlayers})
        </button>
      </div>
      </div>
    </div>
  );

  function handleLeaveLobby() {
    leaveLobby();
  }
    function handleStartGame() {
        document.getElementById('cancelGame')?.classList.remove('hidden');
        document.getElementById('startGame')?.classList.add('hidden');
        startGame();
    }
    function handleCancelGame() {
        document.getElementById('cancelGame')?.classList.add('hidden');
        document.getElementById('startGame')?.classList.remove('hidden');
        cancelGame();
    }
}