import React, { useState, useRef } from 'react';
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'

type LoginProps = {
  onLogin: (name: string) => void;
};

export default function Login({ onLogin }: LoginProps) {
    const [name, setName] = useState('');

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim()) {
      alert('Please enter your name');
      return;
    }
    onLogin(name.trim()); 
  };

  return (
    <div className="h-screen flex items-center justify-center ">
    <div className="flex items-center justify-center p-16 border-2 border-gray-300 rounded-4xl">
    <form onSubmit={handleSubmit}>
      <h1 className="mb-6">Enter your name</h1>
      <Input className="mb-8"
        type="text"
        value={name}
        onChange={e => setName(e.target.value)}
        placeholder="Your name"
      />
      <Button type="submit" className="">Join</Button>
    </form>
    </div>
    </div>
  );
}