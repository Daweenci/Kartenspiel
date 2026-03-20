import { useState } from "react";
import profileIcon from "@/assets/user-profile-icon.svg";

type Props = {
  playerName: string;
  onLogout: () => void;
};



export default function UserProfile({ playerName, onLogout }: Props) {
  const [open, setOpen] = useState(false);
  const [size, setSize] = useState(["?"]);
  const [component, setComponent] = useState();


  const toggleDropdown = () => {
    setOpen(!open);
  };

  return (
    <div className="relative">
      <button
        onClick={toggleDropdown}
        className="px-4 py-2 bg-gray-200 rounded hover:bg-gray-300"
      >
        <span className="text-lg font-semibold">{playerName}</span>  
        <img src={profileIcon} alt="Profile Icon" className="inline w-8 h-8 ml-1" />
      </button>

      {open && (
        <div className="absolute right-0 mt-2 w-40 bg-white border rounded shadow">
          <button
            onClick={() => alert("Settings coming soon!")}
            className="w-full text-left px-4 py-2 hover:bg-gray-100"
          >
            Settings
          </button>
          <button
            onClick={() => alert("Friends coming soon!")}
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
        </div>
      )}
    </div>
  );
}