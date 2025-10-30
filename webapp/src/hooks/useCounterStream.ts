import { useState, useEffect } from 'react';
import { streamCounterService } from '../services/grpcService';

export const useCounterStream = () => {
  const [counterOutput, setCounterOutput] = useState<string[]>([]);

  const handleStartCounterStream = () => {
    setCounterOutput([]);
    const cleanup = streamCounterService(
      (data: string) => setCounterOutput((prev) => [...prev, `Count: ${data}`]),
      (event: any) => {
        console.error('EventSource failed:', event);
        setCounterOutput((prev) => [...prev, `Error in stream: ${event.data || 'Unknown error'}`]);
      },
      (data: string) => setCounterOutput((prev) => [...prev, data])
    );
    return cleanup; // Return cleanup function if needed
  };

  return { counterOutput, handleStartCounterStream };
};
