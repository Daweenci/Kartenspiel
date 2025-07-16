import React, { useState, useRef } from 'react';
import { Button } from '@/components/ui/button'

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
    onLogin(name.trim()); // call the parent's handler
  };

  return (
    <form onSubmit={handleSubmit}>
      <h1>Enter your name</h1>
      <input
        type="text"
        value={name}
        onChange={e => setName(e.target.value)}
        placeholder="Your name"
      />
      <Button type="submit">Join</Button>
    </form>
  );
}