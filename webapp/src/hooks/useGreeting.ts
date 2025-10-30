import React, { useState } from 'react';
import { greetService } from '../services/grpcService';
import { GreetingResponse } from '../types';

export const useGreeting = () => {
  const [name, setName] = useState<string>('');
  const [greeting, setGreeting] = useState<string>('');

  const handleSayHello = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      const data: GreetingResponse = await greetService(name);
      if (data.greeting) {
        setGreeting(data.greeting);
      } else {
        setGreeting(`Error: ${data.error || 'Unknown error'}`);
      }
    } catch (error: any) {
      setGreeting(`Network Error: ${error.message}`);
    }
  };

  return { name, setName, greeting, handleSayHello };
};
