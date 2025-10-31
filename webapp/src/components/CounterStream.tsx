import React from 'react';
import { useCounterStream } from '../hooks';

export const CounterStream: React.FC = () => {
  const { counterOutput, handleStartCounterStream } = useCounterStream();

  return (
    <section className="bg-gray-50 rounded-lg p-6 border border-gray-200">
      <h2 className="text-xl font-semibold text-gray-900 mb-4">Server-Side Streaming RPC: StreamCounter</h2>
      <button 
        onClick={handleStartCounterStream}
        className="mb-4 px-6 py-2 bg-primary-600 text-white font-medium rounded-md hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500 transition-colors"
      >
        Start Counter Stream
      </button>
      <div className="p-4 border border-gray-300 bg-white rounded-md min-h-[100px]">
        {counterOutput.length === 0 ? (
          <p className="text-gray-400 text-sm">Click the button above to start the counter stream.</p>
        ) : (
          counterOutput.map((item, index) => (
            <p key={index} className="py-2 border-b border-gray-100 last:border-0 text-sm text-gray-700 font-mono">{item}</p>
          ))
        )}
      </div>
    </section>
  );
};
