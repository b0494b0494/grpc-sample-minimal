import React from 'react';
import { useCounterStream } from '../hooks';

export const CounterStream: React.FC = () => {
  const { counterOutput, handleStartCounterStream } = useCounterStream();

  return (
    <section>
      <h2>Server-Side Streaming RPC: StreamCounter</h2>
      <button onClick={handleStartCounterStream} style={{ padding: '8px 15px', backgroundColor: '#007bff', color: 'white', border: 'none', borderRadius: '4px', cursor: 'pointer' }}>
        Start Counter Stream
      </button>
      <div style={{ marginTop: '20px', padding: '10px', border: '1px solid #eee', backgroundColor: '#f9f9f9', borderRadius: '4px', minHeight: '50px' }}>
        {counterOutput.map((item, index) => (
          <p key={index} style={{ padding: '5px 0', borderBottom: '1px dotted #ddd' }}>{item}</p>
        ))}
      </div>
    </section>
  );
};
