import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
import { Checkbox } from '@/components/ui/checkbox';
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group';
import React, { useState, useRef, useEffect } from 'react';
import { Check } from 'lucide-react';
import type{ boradcastedLobby, Player } from '@/structs';



type CreateLobbyProps = {
  createLobby: (name: string, maxPlayers: number, isPrivate: boolean, password: string) => void;
  joinLobby: (id: string, joinPassword: string) => void;
  lobbies: boradcastedLobby[];
  currentPlayerID: string;
};

export default function mainMenu({createLobby, joinLobby, lobbies, currentPlayerID}: CreateLobbyProps) {
    const [isPrivate, setIsPrivate] = useState(false);
    const [password, setPassword] = useState('');
    const [lobbyName, setLobbyName] = useState('');
    const [maxPlayers, setMaxPlayers] = useState<"2" | "3" | "4">("2");

    const [joinPassword, setJoinPassword] = useState('');
    const [joinLobbyID, setJoinLobbyID] = useState('');

    return <div>
                <h1 className="text-4xl font-bold mb-4 flex justify-center" >Main Menu</h1>

                <div id="existingLobbies" className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-6 gap-4 px-6">
                    {lobbies?.map((lobby) => (
                        <div key={lobby.id}  onClick={() => joinLobbyAccess(lobby.id)} className="p-4 border rounded shadow bg-white hover:shadow-md transition">
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
                <div id="enterPassword" className="mb-4 hidden fixed inset-0 flex items-center justify-center bg-white/80 z-10">
                    <div className="flex flex-col p-6 border rounded shadow-lg">
                        <form onSubmit={handleJoinSubmit} className="flex flex-col gap-2">
                            <div>
                                <Label >Password:</Label>  
                                <input type="text" 
                                    onChange={(e) => setJoinPassword(e.target.value) } 
                                    placeholder="Enter password"
                                    value={joinPassword}
                                    className="border p-2" >
                                </input>
                            </div>
                            <div className="flex justify-between gap-2 py-2">
                                <Button type="button" onClick={cancelJoinPassword}>Cancel</Button>
                                <Button type="submit">Join</Button>
                            </div>
                        </form>
                    </div>
                </div>
                <div id="lobbyForm" className="mb-4 hidden fixed inset-0 flex items-center justify-center bg-white/80 z-10">
                    <div className="flex flex-col p-6 border rounded shadow-lg">
                        <h2 className="text-xl mb-2 py-2">Create Lobby</h2>
                        <form onSubmit={handleCreateSubmit} className="flex flex-col gap-2">
                        <input
                            type="text"
                            placeholder="Lobby Name"
                            className="border p-2 py-2"
                            value={lobbyName}
                            onChange={(e) => setLobbyName(e.target.value)}
                        />

                        <div className="flex justify-between gap-2 py-2">
                            <Label className="text-sm">Max Player Count:</Label>
                            <RadioGroup
                            value={maxPlayers}
                            onValueChange={(value: "2" | "3" | "4") => setMaxPlayers(value)}
                            className="flex justify-between gap-3"
                            >
                            <div className="flex items-center space-x-2">
                                <RadioGroupItem value="2" id="option-one" />
                                <Label htmlFor="option-one">2</Label>
                            </div>
                            <div className="flex items-center space-x-2">
                                <RadioGroupItem value="3" id="option-two" />
                                <Label htmlFor="option-two">3</Label>
                            </div>
                            <div className="flex items-center space-x-2">
                                <RadioGroupItem value="4" id="option-four" />
                                <Label htmlFor="option-four">4</Label>
                            </div>
                            </RadioGroup>
                        </div>
                        <div className="flex justify-between gap-2 py-2">    
                            <Label htmlFor="privateLobby">Private Lobby:</Label>
                            <Checkbox id="privateLobby" checked={isPrivate} onCheckedChange={toggleIsPrivate} className="h-4 w-4" />
                        </div>
                         <div className="flex justify-between gap-2 py-2">
                            <Label htmlFor="password">Password:</Label>
                            <input
                                type="text"
                                id="password"
                                placeholder="Enter password"
                                value={password}
                                onChange={(e) => setPassword(e.target.value)}
                                disabled={!isPrivate}
                                className="border p-2 disabled:bg-gray-100"
                                />
                        </div>
                        <div className="flex justify-between gap-2 py-2">
                            <Button type="button" onClick={cancelForm}>Cancel</Button>
                            <Button type="submit">Submit</Button>
                        </div>
                        </form>
                    </div>
                </div>
                <Button className="mb-4 text-xl p-6 fixed bottom-10 right-16 z-1" onClick={openLobbyForm}>Create Lobby</Button>
            </div>


    function cancelJoinPassword (){
        document.getElementById("enterPassword")!.classList.add('hidden');
        setJoinPassword('');
        setJoinLobbyID('');
    }

    function handleCreateSubmit(e: React.FormEvent) {
        e.preventDefault();
        if( !lobbyName.trim()) {
            alert('Please enter a lobby name');
            return;
        }
        if( isPrivate && password.trim().length < 4) {
            alert('Please enter a password with at least 4 characters');
            return;
        }
        createLobby(lobbyName.trim(), Number(maxPlayers) , isPrivate, password.trim());
        document.getElementById('lobbyForm')!.classList.add('hidden');

    }

    function toggleIsPrivate() {
        setIsPrivate(prev => {
            const newState = !prev;
            if (!newState) {
                setPassword(''); 
            }
            return newState;
        });
    }

    function openLobbyForm() {
        document.getElementById('lobbyForm')!.classList.remove('hidden');
    }

    function cancelForm() {
        document.getElementById('lobbyForm')!.classList.add('hidden');
        setLobbyName('');
        setMaxPlayers("2");
        setIsPrivate(false);
        setPassword('');
    }

    function joinLobbyAccess(id: string) {
        for (const lobby of lobbies) {
            if (lobby.id === id) {
                if (lobby.players.length >= lobby.maxPlayers) {
                    alert('This lobby is full!');
                    return;
                }

                const isAlreadyInLobby = lobby.players.some(player => player.id == currentPlayerID);
                if (isAlreadyInLobby) {
                    alert('You are already part of the lobby!');
                    return;
                }

                if (lobby.isPrivate) {
                    document.getElementById('enterPassword')!.classList.remove('hidden');
                    setJoinLobbyID(lobby.id);
                } else{
                    joinLobby(id, "")
                }
                break;
            }
        } 
    }

    function handleJoinSubmit(e: React.FormEvent) {
        e.preventDefault();
        joinLobby(joinLobbyID, joinPassword);
        cancelJoinPassword();
    }
}

