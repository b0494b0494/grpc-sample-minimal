import React from 'react';
import { useGreeting } from '../hooks';

export const Greeting: React.FC = () => {
  const { name, setName, greeting, handleSayHello } = useGreeting();

  return (
    <section className="bg-gray-50 rounded-lg p-6 border border-gray-200">
      <h2 className="text-xl font-semibold text-gray-900 mb-4">Unary RPC: SayHello</h2>
      <form onSubmit={handleSayHello} className="flex gap-3 mb-4">
        <input
          type="text"
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="Enter your name"
          className="flex-1 px-4 py-2 border border-gray-300 rounded-md shadow-sm focus:ring-primary-500 focus:border-primary-500"
        />
        <button 
          type="submit"
          className="px-6 py-2 bg-primary-600 text-white font-medium rounded-md hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500 transition-colors"
        >
          Say Hello
        </button>
      </form>
      {greeting && (
        <div className={`p-3 rounded-md ${
          greeting.startsWith('Error') || greeting.startsWith('Network Error') 
            ? 'bg-red-50 text-red-700 border border-red-200' 
            : 'bg-green-50 text-green-700 border border-green-200'
        }`}>
          <p className="font-medium">{greeting}</p>
        </div>
      )}
    </section>
  );
};
