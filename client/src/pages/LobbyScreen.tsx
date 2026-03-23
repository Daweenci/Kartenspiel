import React, { useState } from 'react';
import type { Friend, yourLobby } from '@/structs';
import inviteIcon from "@/assets/invite.svg";

type LobbyScreenProps = {
  initLobby: yourLobby;
  friendsList: Friend[];
  startGame: () => void;
  cancelGame: () => void;
  leaveLobby: () => void;
};

export default function LobbyScreen({
  startGame,
  cancelGame,
  leaveLobby,
  initLobby,
  friendsList,
}: LobbyScreenProps) {
  const [gameStarting, setGameStarting] = useState(false);
  const [showFriends, setShowFriends] = useState(false);


  const handleLeaveLobby = () => {
    leaveLobby();
  };

  const toggleShowFriends = () => {
    setShowFriends(prev => (prev ? false : true));
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
        <div className="inline-block relative">
          <div
            onClick={toggleShowFriends}
            className="border-2 border-gray-300 rounded p-2 cursor-pointer mb-1 w-24 hover:bg-gray-100"
          >
            <span>Invite</span>
            <img src={inviteIcon} alt="Invite" className="inline w-8 h-8 ml-1" />
          </div>

          {showFriends && (
            <div className="absolute left-0 top-full w-64 bg-white border border-gray-200 rounded-xl shadow-lg z-10 overflow-hidden">
              <FriendsList friendsList={friendsList} />
            </div>
          )}
        </div>
        <div className="mb-3"></div>
        <h1>
          <strong>Lobby:</strong> {initLobby.name}
        </h1>
        <h1>
          <strong>
            Players {initLobby.players.length ?? "undefined"}/{initLobby.maxPlayers}:
          </strong>
        </h1>
        <ul className="list-none">
          {initLobby.players?.map((player, index) => (
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

function FriendsList({ friendsList }: { friendsList: Friend[] }) {
  const onlineFriends = friendsList.filter(f => f.isOnline);

  if (onlineFriends.length === 0) {
    return (
      <div className="px-3 py-2 text-sm text-gray-500">
        No friends online
      </div>
    );
  }

  return (
    <ul className="max-h-60 overflow-y-auto divide-y divide-gray-100 m-0 p-0 list-none">
      {onlineFriends.map(friend => (
        <li
          key={friend.id}
          className="flex items-center justify-between px-3 py-2 text-sm hover:bg-gray-100 transition-colors"
        >
          <div className="flex items-center gap-2">
            {/* status dot */}
            <span className="w-2 h-2 rounded-full bg-green-500"></span>

            <span className="text-gray-800">{friend.name}</span>
          </div>
        </li>
      ))}
    </ul>
  );
}
