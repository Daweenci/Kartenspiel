// MainMenu.tsx
import { Button } from '@/components/ui/button';
import React, { useState } from 'react';
import type { broadcastedLobby } from '@/structs';
import CreateLobbyModal from './CreateLobbyModal';
import JoinPasswordModal from './JoinPasswordModal';

type MainMenuProps = {
  createLobby: (name: string, maxPlayers: number, isPrivate: boolean, password: string) => void;
  joinLobby: (id: string, joinPassword: string) => Promise<boolean> | boolean; // Should return success status
  lobbies: broadcastedLobby[];
  currentPlayerID: string;
};

export default function MainMenu({ createLobby, joinLobby, lobbies, currentPlayerID }: MainMenuProps) {
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [showJoinModal, setShowJoinModal] = useState(false);
  const [selectedLobbyId, setSelectedLobbyId] = useState('');

  const handleJoinLobby = async (id: string, password: string) => {
    const success = await joinLobby(id, password);
    if (success) {
      setShowJoinModal(false);
      setSelectedLobbyId('');
    }
    // If join failed, modal stays open so user can try again
  };

  const handleCreateLobby = (name: string, maxPlayers: number, isPrivate: boolean, password: string) => {
    createLobby(name, maxPlayers, isPrivate, password);
    setShowCreateModal(false);
  };

  const joinLobbyAccess = (id: string) => {
    for (const lobby of lobbies) {
      if (lobby.id === id) {
        if (lobby.players.length >= lobby.maxPlayers) {
          alert('This lobby is full!');
          return;
        }

        const isAlreadyInLobby = lobby.players.some(player => player.id === currentPlayerID);
        if (isAlreadyInLobby) {
          alert('You are already part of the lobby!');
          return;
        }

        if (lobby.isPrivate) {
          setSelectedLobbyId(lobby.id);
          setShowJoinModal(true);
        } else {
          joinLobby(id, "");
        }
        break;
      }
    }
  };

  return (
    <div>
      <h1 className="text-4xl font-bold mb-4 flex justify-center">Main Menu</h1>

      <div id="existingLobbies" className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-6 gap-4 px-6">
        {lobbies?.map((lobby) => (
          <div 
            key={lobby.id} 
            onClick={() => joinLobbyAccess(lobby.id)} 
            className="p-4 border rounded shadow bg-white hover:shadow-md transition cursor-pointer"
          >
            <h3 className="text-lg font-semibold mb-2">{lobby.name}</h3>
            <p>Players <strong>({lobby.players.length}/{lobby.maxPlayers})</strong>:</p>
            <ul className="list-disc pl-5 mb-2">
              {lobby.players.map((player, index) => (
                <li key={index}>{player.name}</li>
              ))}
            </ul>
            <p>{lobby.isPrivate ? '🔒 Private' : '🌐 Public'}</p>
          </div>
        ))}
      </div>

      <CreateLobbyModal
        isOpen={showCreateModal}
        onClose={() => setShowCreateModal(false)}
        onCreateLobby={handleCreateLobby}
      />

      <JoinPasswordModal
        isOpen={showJoinModal}
        onClose={() => {
          setShowJoinModal(false);
          setSelectedLobbyId('');
        }}
        onJoinLobby={(password) => handleJoinLobby(selectedLobbyId, password)}
      />

      <Button 
        className="mb-4 text-xl p-6 fixed bottom-10 right-16 z-1" 
        onClick={() => setShowCreateModal(true)}
      >
        Create Lobby
      </Button>
    </div>
  );
}