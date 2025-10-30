import React from 'react';
import { useGreeting } from '../hooks';

export const Greeting: React.FC = () => {
  const { name, setName, greeting, handleSayHello } = useGreeting();

  return (
    <section>
      <h2>Unary RPC: SayHello</h2>
      <form onSubmit={handleSayHello}>
        <input
          type="text"
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="Enter your name"
          style={{ width: 'calc(100% - 100px)', padding: '8px', marginRight: '10px', border: '1px solid #ccc', borderRadius: '4px' }}
        />
        <button type="submit" style={{ padding: '8px 15px', backgroundColor: '#007bff', color: 'white', border: 'none', borderRadius: '4px', cursor: 'pointer' }}>
          Say Hello
        </button>
      </form>
      {greeting && <p style={{ marginTop: '20px', fontWeight: 'bold', color: greeting.startsWith('Error') || greeting.startsWith('Network Error') ? '#dc3545' : '#28a745' }}>{greeting}</p>}
    </section>
  );
};
