import { useState } from "react";

type Props = {
  playerName: string;
  onLogout: () => void;
};

export default function UserProfile({ playerName, onLogout }: Props) {
  const [open, setOpen] = useState(false);

  const toggleDropdown = () => {
    setOpen(!open);
  };

  return (
    <div className="relative">
      <button
        onClick={toggleDropdown}
        className="px-4 py-2 bg-gray-200 rounded hover:bg-gray-300"
      >
        {playerName}
      </button>

      {open && (
        <div className="absolute right-0 mt-2 w-40 bg-white border rounded shadow">
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