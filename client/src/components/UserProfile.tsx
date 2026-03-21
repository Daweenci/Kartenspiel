import { useState } from "react";
import profileIcon from "@/assets/user-profile-icon.svg";
import type { friendRequest, Player } from "@/structs";

type Props = {
  playerName: string;
  onLogout: () => void;
  onAddFriend: (friendName: string) => void;
  pendingFriendRequests: friendRequest[];
  friendsList: Player[];
  onAcceptFriendRequest: (friendID: string, accept: boolean) => void;
};



export default function UserProfile({ playerName, onLogout, onAddFriend, pendingFriendRequests, friendsList, onAcceptFriendRequest }: Props) {
  type View = "menu" | "friends" | "settings" |  null;
  const [view, setView] = useState<View>(null);

  const toggleDropdown = () => {
  setView(prev => (prev ? null : "menu"));
};

  return (
    <div className="relative">
      <button
        onClick={toggleDropdown}
        className="px-4 py-2 bg-gray-200 rounded hover:bg-gray-300"
      >
        <span className="text-lg font-semibold">{playerName}</span>  
        <img src={profileIcon} alt="Profile Icon" className="inline w-8 h-8 ml-1" />
        {pendingFriendRequests.length > 0 && (
          <span className="absolute top-0 right-0 bg-red-500 text-white text-xs rounded-full h-5 w-5 flex items-center justify-center">
            {pendingFriendRequests.length}
          </span>
        )}
      </button>

      {view && (
        <div className="absolute right-0 mt-2 w-40 bg-white border rounded shadow">
          
          {view === "menu" && (
            <>
              <button
                onClick={() => setView("settings")}
                className="w-full text-left px-4 py-2 hover:bg-gray-100"
              >
                Settings
              </button>
              <button
                onClick={() => setView("friends")}
                className="w-full text-left px-4 py-2 hover:bg-gray-100"
              >
                Friends
              </button>
              <button
                onClick={onLogout}
                className="w-full text-left px-4 py-2 hover:bg-gray-100"
              >
                Logout
              </button>
            </>
          )}

          {view === "friends" && (
            <div className="absolute right-0 mt-2 w-fit bg-white border rounded shadow">
              <div className="p-2 flex flex-col gap-2 items-start">
                <button
                  onClick={() => setView("menu")}
                  className="text-sm border border-blue-500 text-blue-500 rounded px-3 py-1 hover:bg-blue-500 hover:text-white transition duration-200"
                >
                  ← Back
                </button>

                <div className="flex gap-2">
                  <input
                    id="friendNameInput"
                    type="text"
                    placeholder="Enter player name..."
                    className="border border-gray-300 rounded px-3 py-1 focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                  <button onClick={() => onAddFriend((document.querySelector('#friendNameInput') as HTMLInputElement)?.value || '')} className="bg-blue-500 text-white px-3 py-1 rounded hover:bg-blue-600 transition duration-200 whitespace-nowrap">
                    Add Friend
                  </button>
                </div>
                <h3 className="text-lg font-semibold mt-4">Pending Friend Requests</h3>
                {pendingFriendRequests.length === 0 ? (
                  <p className="text-gray-500">No pending requests</p>
                ) : (
                  <ul className="list-disc pl-5">
                    {pendingFriendRequests.map((req) => (
                      
                      <li key={req.friendID} className="flex items-center gap-2">
                        <span className="text-gray-500">{req.friendName}</span> 
                        <button className="bg-green-500 text-white px-2 py-1 rounded hover:bg-green-600 transition duration-200" onClick={() => onAcceptFriendRequest(req.friendID, true)}>
                          Accept
                        </button>
                        <button className="bg-red-500 text-white px-2 py-1 rounded hover:bg-red-600 transition duration-200" onClick={() => onAcceptFriendRequest(req.friendID, false)}>
                          Decline
                        </button>
                      </li>
                    ))}
                  </ul>
                )}
                <h3 className="text-lg font-semibold mt-4">Friends List</h3>
                {friendsList.length === 0 ? (
                  <p className="text-gray-500">No friends yet</p>
                ) : (
                  <ul className="list-disc pl-5">
                    {friendsList.map((friend) => (
                      <li key={friend.id} className="flex items-center gap-2">
                        <span className="text-gray-500">{friend.name}</span>
                      </li>
                    ))}
                  </ul>
                )}
              </div>
            </div>
          )}

          {view === "settings" && (
            <div className="p-2">
              <button
                onClick={() => setView("menu")}
                className="text-sm mb-2 border border-blue-500 text-blue-500 rounded px-3 py-1 hover:bg-blue-500 hover:text-white transition duration-200"
              >
                ← Back
              </button>
              <div>Settings component here</div>
            </div>
          )}

        </div>
      )}
    </div>
  );
}